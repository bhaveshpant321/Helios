-- ==========================================================
-- Helios: Database Seed Data (FULLY UPDATED for Fund Management)
-- This script inserts the initial, non-transactional, and test data
-- required for the platform to function and be testable.
-- MUST be run after schema.sql (which includes the held_balance column).
-- ==========================================================

BEGIN;

-- ----------------------------------------------------------
-- 1. ASSETS (Base Currencies)
-- ----------------------------------------------------------
INSERT INTO assets (ticker_symbol, name, decimals) VALUES
('USD', 'United States Dollar (Quote)', 2),   -- Primary Quote Currency
('BTC', 'Bitcoin (Base)', 8),
('ETH', 'Ethereum (Base)', 6),
('SOL', 'Solana (Base)', 4);

-- ----------------------------------------------------------
-- 2. TRADING_PAIRS (Tradable Instruments)
-- ----------------------------------------------------------
-- Format: (base_asset_id, quote_asset_id, symbol)
-- ----------------------------------------------------------
INSERT INTO trading_pairs (base_asset_id, quote_asset_id, symbol) VALUES
(2, 1, 'BTC/USD'),   -- ID: 1
(3, 1, 'ETH/USD'),   -- ID: 2
(4, 2, 'SOL/BTC');   -- ID: 3


-- ----------------------------------------------------------
-- 3. TEST USERS
-- ----------------------------------------------------------
INSERT INTO users (id, username, email, password_hash) VALUES
(1001, 'sandhya_test', 'sandhya@test.com', 'sandhya_hash'),
(1002, 'muneef_test',  'muneef@test.com',  'muneef_hash'),
(1003, 'bhavesh_test', 'bhavesh@test.com', 'bhavesh_hash'),
(1004, 'market_maker', 'mm@test.com',      'mm_hash');

SELECT setval('users_id_seq', 1004, true);


-- ----------------------------------------------------------
-- 4. INITIAL ACCOUNTS & BALANCES
-- ----------------------------------------------------------
-- Updated to include held_balance for fund management
-- Each user should have at least one row per asset they trade.
-- ----------------------------------------------------------
INSERT INTO accounts (user_id, asset_id, balance, held_balance) VALUES
-- Sandhya (1001): Trades BTC/USD
(1001, 1, 50000.00, 0),         -- 50,000 USD
(1001, 2, 5.00000000, 0),       -- 5 BTC

-- Muneef (1002): Trades BTC/USD
(1002, 1, 10000.00, 0),
(1002, 2, 3.00000000, 0),

-- Bhavesh (1003): Trades BTC/USD
(1003, 1, 8000.00, 0),
(1003, 2, 5.00000000, 0),

-- Market Maker (1004): Deep liquidity provider for all pairs
(1004, 1, 500000.00, 0),        -- USD
(1004, 2, 100.00000000, 0),     -- BTC
(1004, 3, 2000.000000, 0),      -- ETH
(1004, 4, 5000.0000, 0);        -- SOL


-- ----------------------------------------------------------
-- 5. TEST ORDERS
-- ----------------------------------------------------------
-- Relies on Trading Pair IDs (1=BTC/USD, 2=ETH/USD, 3=SOL/BTC)
-- ----------------------------------------------------------
INSERT INTO orders (
    id, user_id, trading_pair_id, side, type, price,
    quantity, filled_quantity, status, created_at
) VALUES
-- BTC/USD BUY Side (Bids)
(2001, 1001, 1, 'BUY',  'LIMIT', 40000.00, 1.00000000, 0.00000000, 'OPEN',             NOW() - INTERVAL '1 hour'),
(2002, 1004, 1, 'BUY',  'LIMIT', 40000.00, 2.00000000, 0.00000000, 'OPEN',             NOW() - INTERVAL '45 minutes'),
(2003, 1004, 1, 'BUY',  'LIMIT', 39990.00, 5.00000000, 2.00000000, 'PARTIALLY_FILLED', NOW() - INTERVAL '30 minutes'),

-- BTC/USD SELL Side (Asks)
(2004, 1002, 1, 'SELL', 'LIMIT', 40100.00, 1.50000000, 0.00000000, 'OPEN',             NOW() - INTERVAL '20 minutes'),
(2005, 1003, 1, 'SELL', 'LIMIT', 40100.00, 0.50000000, 0.00000000, 'OPEN',             NOW() - INTERVAL '15 minutes'),
(2006, 1003, 1, 'SELL', 'LIMIT', 40200.00, 3.00000000, 1.00000000, 'PARTIALLY_FILLED', NOW() - INTERVAL '10 minutes'),

-- Closed Orders (for history testing)
(2007, 1001, 1, 'BUY',  'MARKET', NULL, 0.50000000, 0.50000000, 'FILLED',              NOW() - INTERVAL '5 minutes'),
(2008, 1002, 1, 'SELL', 'LIMIT',  40050.00, 1.00000000, 0.00000000, 'CANCELLED',       NOW() - INTERVAL '2 minutes');

SELECT setval('orders_id_seq', 2008, true);


-- ----------------------------------------------------------
-- 6. TEST TRADES
-- ----------------------------------------------------------
-- Trades reference existing orders from section 5
-- ----------------------------------------------------------
INSERT INTO trades (
    id, trading_pair_id, maker_order_id, taker_order_id,
    price, quantity, executed_at
) VALUES
(3001, 1, 2003, 2007, 39995.00, 0.25000000, NOW() - INTERVAL '4 minutes'),
(3002, 1, 2003, 2007, 39995.00, 0.25000000, NOW() - INTERVAL '3 minutes'),
(3003, 1, 2006, 2007, 40090.00, 0.50000000, NOW() - INTERVAL '2 minutes'),

-- Example ETH/USD trade
(3004, 2, 2003, 2004, 2500.00, 10.000000, NOW() - INTERVAL '1 minute');

SELECT setval('trades_id_seq', 3004, true);


-- ----------------------------------------------------------
-- 7. FEES
-- ----------------------------------------------------------
-- Links to Trade IDs above
-- ----------------------------------------------------------
INSERT INTO fees (trade_id, user_id, fee_type, amount) VALUES
(3001, 1001, 'taker', 0.0010),
(3001, 1004, 'maker', 0.0005),
(3002, 1001, 'taker', 0.0010),
(3003, 1003, 'maker', 0.0004),
(3004, 1004, 'maker', 0.0005);

-- ----------------------------------------------------------
-- Finalize
-- ----------------------------------------------------------
COMMIT;

-- ==========================================================
-- END OF SEED FILE
-- ==========================================================
