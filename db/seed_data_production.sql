-- ==========================================================
-- Helios: PRODUCTION Seed Data
-- ==========================================================
-- This file contains ONLY the reference data required for
-- the platform to function in production.
--
-- NO test users, NO test orders, NO test trades.
-- Real users will register through the API.
-- ==========================================================

BEGIN;

-- ----------------------------------------------------------
-- 1. ASSETS (Cryptocurrencies and Fiat Currencies)
-- ----------------------------------------------------------
-- These are the tradable assets on the platform.
-- Decimals define the precision for each asset.
-- ----------------------------------------------------------
INSERT INTO assets (ticker_symbol, name, decimals) VALUES
-- Fiat Currencies
('USD',  'United States Dollar', 2),
('EUR',  'Euro', 2),

-- Cryptocurrencies
('BTC',  'Bitcoin', 8),
('ETH',  'Ethereum', 6),
('SOL',  'Solana', 4),
('USDT', 'Tether USD', 2),
('USDC', 'USD Coin', 2),
('BNB',  'Binance Coin', 4),
('XRP',  'Ripple', 6),
('ADA',  'Cardano', 6);


-- ----------------------------------------------------------
-- 2. TRADING_PAIRS (Available Markets)
-- ----------------------------------------------------------
-- These define which assets can be traded against each other.
-- Format: (base_asset_id, quote_asset_id, symbol)
-- ----------------------------------------------------------
INSERT INTO trading_pairs (base_asset_id, quote_asset_id, symbol) VALUES
-- USD Quote Pairs (Most Common)
(3, 1, 'BTC/USD'),    -- Bitcoin / US Dollar
(4, 1, 'ETH/USD'),    -- Ethereum / US Dollar
(5, 1, 'SOL/USD'),    -- Solana / US Dollar
(8, 1, 'BNB/USD'),    -- Binance Coin / US Dollar
(9, 1, 'XRP/USD'),    -- Ripple / US Dollar
(10, 1, 'ADA/USD'),   -- Cardano / US Dollar

-- USDT Quote Pairs (Stablecoin Trading)
(3, 6, 'BTC/USDT'),
(4, 6, 'ETH/USDT'),
(5, 6, 'SOL/USDT'),
(8, 6, 'BNB/USDT'),

-- USDC Quote Pairs (Alternative Stablecoin)
(3, 7, 'BTC/USDC'),
(4, 7, 'ETH/USDC'),

-- BTC Quote Pairs (Crypto-to-Crypto)
(4, 3, 'ETH/BTC'),
(5, 3, 'SOL/BTC'),

-- EUR Quote Pairs (European Market)
(3, 2, 'BTC/EUR'),
(4, 2, 'ETH/EUR');

COMMIT;

-- ==========================================================
-- END OF PRODUCTION SEED DATA
-- ==========================================================
-- 
-- The platform is now ready to accept user registrations
-- and trading activity through the API.
--
-- Users will create accounts via POST /api/v1/auth/register
-- Orders will be placed via POST /api/v1/orders
-- All data from this point forward is real user activity.
-- ==========================================================
