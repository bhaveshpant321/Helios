-- Set the language for the function
CREATE EXTENSION IF NOT EXISTS plpgsql;

-- The main function definition
CREATE OR REPLACE FUNCTION sp_place_order(
    p_user_id BIGINT,
    p_trading_pair_id INT,
    p_side order_side,
    p_type order_type,
    p_quantity DECIMAL,
    p_price DECIMAL DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    -- === Variable Declarations ===
    v_base_asset_id INT;
    v_quote_asset_id INT;
    v_available_balance DECIMAL;
    v_held_balance DECIMAL;
    v_total_cost DECIMAL;
    v_asset_to_check INT;
    v_taker_order_id BIGINT;
    v_taker_order orders%ROWTYPE;
    v_maker_order RECORD;
    v_trade_quantity DECIMAL;
    v_trade_price DECIMAL;
    v_outcome order_status := 'OPEN'::order_status;

    -- NEW: Fee variables
    v_taker_rate DECIMAL := 0.001; -- 0.1% Taker Fee
    v_maker_rate DECIMAL := 0.0005; -- 0.05% Maker Fee
    v_taker_fee DECIMAL;
    v_maker_fee DECIMAL;
    v_trade_id BIGINT;

BEGIN

    -- === 1. Initial Validation & Fund Calculation (REVISED) ===
    SELECT base_asset_id, quote_asset_id
    INTO v_base_asset_id, v_quote_asset_id
    FROM trading_pairs WHERE id = p_trading_pair_id;

    IF NOT FOUND THEN
        RETURN jsonb_build_object('status', 'ERROR', 'message', 'Invalid trading pair.');
    END IF;

    -- --- NEW: FIX 3 - REVISED COST ESTIMATION LOGIC (SLIPPAGE) ---
    IF p_side = 'BUY' THEN
        v_asset_to_check := v_quote_asset_id;

        IF p_type = 'LIMIT' THEN
            v_total_cost := (p_quantity * p_price) * (1 + v_taker_rate);
        ELSE
            -- MARKET order slippage calculation
            v_total_cost := 0;
            DECLARE
                v_quantity_to_fill DECIMAL := p_quantity;
                v_market_order RECORD;
            BEGIN
                FOR v_market_order IN
                    SELECT price, (quantity - filled_quantity) AS remaining_qty
                    FROM orders
                    WHERE trading_pair_id = p_trading_pair_id
                      AND side = 'SELL'
                      AND status IN ('OPEN', 'PARTIALLY_FILLED')
                    ORDER BY price ASC
                LOOP
                    IF v_quantity_to_fill <= v_market_order.remaining_qty THEN
                        v_total_cost := v_total_cost + (v_quantity_to_fill * v_market_order.price);
                        v_quantity_to_fill := 0;
                        EXIT;
                    ELSE
                        v_total_cost := v_total_cost + (v_market_order.remaining_qty * v_market_order.price);
                        v_quantity_to_fill := v_quantity_to_fill - v_market_order.remaining_qty;
                    END IF;
                END LOOP;

                IF v_quantity_to_fill > 0 THEN
                    RETURN jsonb_build_object('status', 'ERROR', 'message', 'Insufficient liquidity to fill market order.');
                END IF;
                
                -- Add taker fee to estimated market cost
                v_total_cost := v_total_cost * (1 + v_taker_rate);
            END;
        END IF;

    ELSE -- p_side = 'SELL'
        v_asset_to_check := v_base_asset_id;
        v_total_cost := p_quantity;
    END IF;
    -- --- END REVISED COST ESTIMATION LOGIC ---

    -- NEW: Row-level lock on account to prevent race conditions
    SELECT balance, held_balance INTO v_available_balance, v_held_balance
    FROM accounts 
    WHERE user_id = p_user_id AND asset_id = v_asset_to_check
    FOR UPDATE;

    IF NOT FOUND THEN
        -- If account doesn't exist, create it
        INSERT INTO accounts (user_id, asset_id, balance, held_balance)
        VALUES (p_user_id, v_asset_to_check, 0, 0)
        RETURNING balance, held_balance INTO v_available_balance, v_held_balance;
    END IF;

    -- Check for sufficient funds
    IF v_available_balance < v_total_cost THEN
        RETURN jsonb_build_object('status', 'ERROR', 'message', 'Insufficient funds. Required: ' || v_total_cost || ', Available: ' || v_available_balance);
    END IF;

    -- NEW: Move funds to held_balance for LIMIT orders
    IF p_type = 'LIMIT' THEN
        UPDATE accounts 
        SET balance = balance - v_total_cost,
            held_balance = held_balance + v_total_cost
        WHERE user_id = p_user_id AND asset_id = v_asset_to_check;
        
        -- Update local variables to reflect reality
        v_available_balance := v_available_balance - v_total_cost;
        v_held_balance := v_held_balance + v_total_cost;
    END IF;

    -- Insert the order to get an ID
    INSERT INTO orders (user_id, trading_pair_id, side, type, status, price, quantity, filled_quantity)
    VALUES (p_user_id, p_trading_pair_id, p_side, p_type, 'OPEN'::order_status, p_price, p_quantity, 0.0)
    RETURNING id INTO v_taker_order_id;


    -- Initialize the internal record state
    v_taker_order.id := v_taker_order_id;
    v_taker_order.quantity := p_quantity;
    v_taker_order.filled_quantity := 0.0;
    v_taker_order.status := 'OPEN'::order_status;

    -- === 2. The Matching Engine Loop ===
    FOR v_maker_order IN
        SELECT * FROM orders
        WHERE trading_pair_id = p_trading_pair_id
          AND status IN ('OPEN', 'PARTIALLY_FILLED')
          AND side != p_side
          AND ( (p_side = 'BUY' AND price <= p_price) OR (p_side = 'SELL' AND price >= p_price) OR p_type = 'MARKET' )
        ORDER BY
            CASE WHEN p_side = 'BUY' THEN price END ASC,
            CASE WHEN p_side = 'SELL' THEN price END DESC,
            created_at ASC
        FOR UPDATE SKIP LOCKED
    LOOP
        v_trade_price := v_maker_order.price;
        v_trade_quantity := LEAST(
            v_taker_order.quantity - v_taker_order.filled_quantity,
            v_maker_order.quantity - v_maker_order.filled_quantity
        );

        -- === 3. Execute The Trade ===

        -- NEW: FIX 2 - Calculate Fees
        v_taker_fee := v_trade_quantity * v_trade_price * v_taker_rate;
        v_maker_fee := v_trade_quantity * v_trade_price * v_maker_rate;

        -- A. Create the trade record
        INSERT INTO trades (maker_order_id, taker_order_id, trading_pair_id, price, quantity)
        VALUES (v_maker_order.id, v_taker_order_id, p_trading_pair_id, v_trade_price, v_trade_quantity)
        RETURNING id INTO v_trade_id; -- Get the trade ID

        -- NEW: FIX 2 - Log the fees
        INSERT INTO fees (trade_id, user_id, fee_type, amount)
        VALUES (v_trade_id, p_user_id, 'taker', v_taker_fee);

        INSERT INTO fees (trade_id, user_id, fee_type, amount)
        VALUES (v_trade_id, v_maker_order.user_id, 'maker', v_maker_fee);

        -- B. Update balances for both users (REVISED WITH HELD_BALANCE)
        IF p_side = 'BUY' THEN
            -- Taker (Buyer) pays quote, receives base
            -- If LIMIT, quote was already in held_balance
            IF p_type = 'LIMIT' THEN
                UPDATE accounts SET held_balance = held_balance - (v_trade_quantity * v_trade_price) - v_taker_fee WHERE user_id = p_user_id AND asset_id = v_quote_asset_id;
            ELSE
                UPDATE accounts SET balance = balance - (v_trade_quantity * v_trade_price) - v_taker_fee WHERE user_id = p_user_id AND asset_id = v_quote_asset_id;
            END IF;
            UPDATE accounts SET balance = balance + v_trade_quantity WHERE user_id = p_user_id AND asset_id = v_base_asset_id;
            
            -- Maker (Seller) pays base, receives quote
            -- Sellers always have their base asset in held_balance when their order is on the book
            UPDATE accounts SET held_balance = held_balance - v_trade_quantity WHERE user_id = v_maker_order.user_id AND asset_id = v_base_asset_id;
            UPDATE accounts SET balance = balance + (v_trade_quantity * v_trade_price) - v_maker_fee WHERE user_id = v_maker_order.user_id AND asset_id = v_quote_asset_id;
        ELSE -- p_side = 'SELL'
            -- Taker (Seller) pays base, receives quote
            -- If LIMIT, base was already in held_balance
            IF p_type = 'LIMIT' THEN
                UPDATE accounts SET held_balance = held_balance - v_trade_quantity WHERE user_id = p_user_id AND asset_id = v_base_asset_id;
            ELSE
                UPDATE accounts SET balance = balance - v_trade_quantity WHERE user_id = p_user_id AND asset_id = v_base_asset_id;
            END IF;
            UPDATE accounts SET balance = balance + (v_trade_quantity * v_trade_price) - v_taker_fee WHERE user_id = p_user_id AND asset_id = v_quote_asset_id;
            
            -- Maker (Buyer) pays quote, receives base
            -- Buyers always have their quote asset in held_balance when their order is on the book
            UPDATE accounts SET held_balance = held_balance - (v_trade_quantity * v_trade_price) WHERE user_id = v_maker_order.user_id AND asset_id = v_quote_asset_id;
            UPDATE accounts SET balance = balance + v_trade_quantity WHERE user_id = v_maker_order.user_id AND asset_id = v_base_asset_id;
            -- Subtract maker fee from quote received (Wait, maker buyer receives base... so fee should be from quote balance or reduced asset)
            -- For consistency with schema and previous logic, we'll subtract fee from the quote account.
            UPDATE accounts SET balance = balance - v_maker_fee WHERE user_id = v_maker_order.user_id AND asset_id = v_quote_asset_id;
        END IF;

        -- C. Update the maker order
        UPDATE orders
        SET filled_quantity = filled_quantity + v_trade_quantity,
            status = CASE WHEN (filled_quantity + v_trade_quantity) = quantity THEN 'FILLED'::order_status ELSE 'PARTIALLY_FILLED'::order_status END
        WHERE id = v_maker_order.id;

        -- D. Update the state of our incoming taker order
        v_taker_order.filled_quantity := v_taker_order.filled_quantity + v_trade_quantity;
        IF v_taker_order.filled_quantity = v_taker_order.quantity THEN
            v_taker_order.status := 'FILLED'::order_status;
            v_outcome := 'FILLED'::order_status;
            EXIT;
        ELSE
            v_outcome := 'PARTIALLY_FILLED'::order_status;
        END IF;
    END LOOP;

    -- === 4. Post-Loop Handling ===
    UPDATE orders
    SET filled_quantity = v_taker_order.filled_quantity,
        status = v_taker_order.status
    WHERE id = v_taker_order_id;

    -- === 5. Finalization ===
    PERFORM pg_notify('market_update', jsonb_build_object('pair_id', p_trading_pair_id)::text);

    RETURN jsonb_build_object(
        'status', 'SUCCESS',
        'outcome', v_outcome,
        'order_id', v_taker_order_id
    );

EXCEPTION
    WHEN others THEN
        RETURN jsonb_build_object('status', 'ERROR', 'message', SQLERRM);
END;
$$ LANGUAGE plpgsql;