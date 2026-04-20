-- ==========================================================
-- Helios: Part 2 Stored Procedures (Owner: Sandhya)
-- Order Management and Data Retrieval
-- ==========================================================

-- ----------------------------------------------------------
-- 1. sp_cancel_order (FIXED VERSION)
-- Purpose: Cancel an active order and release held funds
-- ----------------------------------------------------------
CREATE OR REPLACE PROCEDURE sp_cancel_order(
    p_order_id BIGINT,
    p_user_id BIGINT
)
LANGUAGE plpgsql
AS $$
DECLARE
    current_status order_status;
    order_owner_id BIGINT;
    order_side order_side;
    order_type order_type;
    remaining_qty DECIMAL(38,18);
    order_price DECIMAL(38,18);
    order_trading_pair_id INT;
    base_asset_id INT;
    quote_asset_id INT;
BEGIN
    -- 1. Check if the order exists and get its details (FIXED: JOIN with trading_pairs)
    SELECT 
        o.status, 
        o.user_id, 
        o.side, 
        o.type,
        o.quantity - o.filled_quantity,
        o.price,
        o.trading_pair_id,
        tp.base_asset_id,
        tp.quote_asset_id
    INTO 
        current_status, 
        order_owner_id, 
        order_side, 
        order_type,
        remaining_qty,
        order_price,
        order_trading_pair_id,
        base_asset_id,
        quote_asset_id
    FROM orders o
    JOIN trading_pairs tp ON o.trading_pair_id = tp.id
    WHERE o.id = p_order_id;

    -- If no order found, raise an exception
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Order ID % not found.', p_order_id;
    END IF;

    -- 2. Verify ownership
    IF order_owner_id <> p_user_id THEN
        RAISE EXCEPTION 'Permission denied. User % is not the owner of Order %.', p_user_id, p_order_id;
    END IF;

    -- 3. Verify status (must be cancellable: OPEN or PARTIALLY_FILLED)
    IF current_status NOT IN ('OPEN', 'PARTIALLY_FILLED') THEN
        RAISE EXCEPTION 'Order % cannot be cancelled. Current status: %.', p_order_id, current_status;
    END IF;

    -- 4. Perform the cancellation
    UPDATE orders
    SET status = 'CANCELLED'
    WHERE id = p_order_id;

    -- ======================================================
    -- 5. FUND RELEASE LOGIC (FIXED)
    -- Only LIMIT orders have held funds (MARKET orders execute immediately)
    -- ======================================================
    IF order_type = 'LIMIT' AND remaining_qty > 0 THEN
        IF order_side = 'BUY' THEN
            -- BUY order: quote currency was held (price * remaining quantity * (1 + fee_rate))
            -- FIXED: Include the 0.1% taker fee in the release logic
            UPDATE accounts
            SET held_balance = held_balance - (remaining_qty * order_price * 1.001),
                balance = balance + (remaining_qty * order_price * 1.001)
            WHERE user_id = p_user_id AND asset_id = quote_asset_id;

        ELSIF order_side = 'SELL' THEN
            -- SELL order: base currency was held (remaining quantity)
            -- FIXED: Changed currency_id to asset_id and variable names
            UPDATE accounts
            SET held_balance = held_balance - remaining_qty,
                balance = balance + remaining_qty
            WHERE user_id = p_user_id AND asset_id = base_asset_id;
        END IF;
    END IF;

    -- Optional: Log the cancellation (uncomment if you add an audit table)
    -- INSERT INTO order_audit(order_id, user_id, action, timestamp)
    -- VALUES (p_order_id, p_user_id, 'CANCELLED', NOW());

END;
$$;


-- ----------------------------------------------------------
-- 2. sp_get_user_order_history
-- Purpose: Retrieve a user's complete order history for a trading pair.
-- ----------------------------------------------------------
CREATE OR REPLACE FUNCTION sp_get_user_order_history(
    p_user_id BIGINT,
    p_trading_pair_id INT
)
RETURNS SETOF orders  -- Returns rows from the existing orders table schema
LANGUAGE sql
AS $$
    SELECT *
    FROM orders
    WHERE user_id = p_user_id
    AND trading_pair_id = p_trading_pair_id
    ORDER BY created_at DESC;
$$;


-- ----------------------------------------------------------
-- 3. sp_get_order_book
-- Purpose: Retrieve the aggregated, public order book (open bids/asks).
-- ----------------------------------------------------------
CREATE OR REPLACE FUNCTION sp_get_order_book(
    p_trading_pair_id INT
)
RETURNS TABLE(
    side order_side,
    price DECIMAL,
    total_quantity DECIMAL
)
LANGUAGE sql
AS $$
    SELECT
        o.side,
        o.price,
        SUM(o.quantity - o.filled_quantity) AS total_quantity -- Sum of remaining quantity
    FROM orders o
    WHERE o.trading_pair_id = p_trading_pair_id
    -- Only orders that are OPEN or PARTIALLY_FILLED are on the order book
    AND o.status IN ('OPEN', 'PARTIALLY_FILLED')
    -- Only LIMIT orders have a price and are shown on the book
    AND o.type = 'LIMIT'
    GROUP BY 1, 2
    ORDER BY 
        CASE WHEN o.side = 'BUY' THEN 1 ELSE 2 END,       -- Group BUYs first, then SELLs
        CASE WHEN o.side = 'BUY' THEN o.price END DESC,   -- Bids (BUYs) ordered high-to-low
        CASE WHEN o.side = 'SELL' THEN o.price END ASC;   -- Asks (SELLs) ordered low-to-high
$$;


-- ----------------------------------------------------------
-- 4. sp_get_trade_history
-- Purpose: Retrieve recent public trades for a trading pair.
-- ----------------------------------------------------------
CREATE OR REPLACE FUNCTION sp_get_trade_history(
    p_trading_pair_id INT,
    p_limit INT DEFAULT 100
)
RETURNS SETOF trades -- Returns rows from the existing trades table schema
LANGUAGE sql
AS $$
    SELECT *
    FROM trades
    WHERE trading_pair_id = p_trading_pair_id
    ORDER BY executed_at DESC
    LIMIT p_limit;
$$;