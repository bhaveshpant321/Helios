-- ==========================================================
--  Helios: Real-Time Digital Asset Trading Platform
--  PostgreSQL Database Schema Definition
-- ==========================================================

DROP TYPE IF EXISTS order_side;
DROP TYPE IF EXISTS order_type;
DROP TYPE IF EXISTS order_status;

-- ==========================================================
-- ENUM TYPES
-- ==========================================================
CREATE TYPE order_side AS ENUM ('BUY', 'SELL');
CREATE TYPE order_type AS ENUM ('MARKET', 'LIMIT');
CREATE TYPE order_status AS ENUM ('OPEN', 'FILLED', 'PARTIALLY_FILLED', 'CANCELLED');

-- ==========================================================
-- USERS TABLE
-- ==========================================================
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ==========================================================
-- ASSETS TABLE
-- ==========================================================
CREATE TABLE assets (
    id SERIAL PRIMARY KEY,
    ticker_symbol VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    decimals INT DEFAULT 8 CHECK (decimals BETWEEN 0 AND 18)
);

-- ==========================================================
-- TRADING_PAIRS TABLE
-- Prevents trading an asset against itself
-- ==========================================================
CREATE TABLE trading_pairs (
    id SERIAL PRIMARY KEY,
    base_asset_id INT NOT NULL REFERENCES assets(id) ON DELETE RESTRICT,
    quote_asset_id INT NOT NULL REFERENCES assets(id) ON DELETE RESTRICT,
    symbol VARCHAR(20) UNIQUE NOT NULL,
    CHECK (base_asset_id <> quote_asset_id)
);

-- ==========================================================
-- ACCOUNTS TABLE
-- ==========================================================
CREATE TABLE accounts (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    asset_id INT NOT NULL REFERENCES assets(id) ON DELETE RESTRICT,
    balance DECIMAL(38,18) DEFAULT 0 CHECK (balance >= 0),
    held_balance DECIMAL(38,18) DEFAULT 0 CHECK (held_balance >=0),
    PRIMARY KEY (user_id, asset_id)
);

CREATE INDEX idx_accounts_user_id ON accounts (user_id);

-- ==========================================================
-- ORDERS TABLE
-- Enforces that LIMIT orders must have a price, MARKET orders must not
-- ==========================================================
CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    trading_pair_id INT NOT NULL REFERENCES trading_pairs(id) ON DELETE RESTRICT,
    side order_side NOT NULL,
    type order_type NOT NULL,
    status order_status DEFAULT 'OPEN',
    price DECIMAL(38,18),
    quantity DECIMAL(38,18) NOT NULL CHECK (quantity > 0),
    filled_quantity DECIMAL(38,18) DEFAULT 0 CHECK (filled_quantity >= 0),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    CHECK (
        (type = 'MARKET' AND price IS NULL) OR
        (type = 'LIMIT' AND price IS NOT NULL AND price > 0)
    )
);

CREATE INDEX idx_orders_trading_pair_side_status_price
    ON orders (trading_pair_id, side, status, price);

-- ==========================================================
-- TRADES TABLE
-- Prevents an order from being matched with itself
-- ==========================================================
CREATE TABLE trades (
    id BIGSERIAL PRIMARY KEY,
    maker_order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
    taker_order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
    trading_pair_id INT NOT NULL REFERENCES trading_pairs(id) ON DELETE RESTRICT,
    price DECIMAL(38,18) NOT NULL CHECK (price > 0),
    quantity DECIMAL(38,18) NOT NULL CHECK (quantity > 0),
    executed_at TIMESTAMPTZ DEFAULT NOW(),
    CHECK (maker_order_id <> taker_order_id)
);

CREATE INDEX idx_trades_trading_pair_id ON trades (trading_pair_id);
CREATE INDEX idx_trades_executed_at ON trades (executed_at);

-- ==========================================================
-- FEES TABLE
-- Supports maker/taker/exchange fees per trade (extensible design)
-- ==========================================================
CREATE TABLE fees (
    id BIGSERIAL PRIMARY KEY,
    trade_id BIGINT NOT NULL REFERENCES trades(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    fee_type VARCHAR(20) NOT NULL,  -- e.g. 'maker', 'taker', 'exchange'
    amount DECIMAL(38,18) NOT NULL CHECK (amount >= 0)
);

CREATE INDEX idx_fees_trade_id ON fees (trade_id);
CREATE INDEX idx_fees_user_id ON fees (user_id);

-- ==========================================================
-- OPTIONAL PARTITIONING NOTE
-- Future enhancement: partition orders by status to isolate OPEN orders
-- Example:
--   CREATE TABLE orders_open PARTITION OF orders FOR VALUES IN ('OPEN');
--   CREATE TABLE orders_closed PARTITION OF orders FOR VALUES IN ('FILLED', 'CANCELLED', 'PARTIALLY_FILLED');
-- ==========================================================

-- ==========================================================
-- END OF SCHEMA
-- ==========================================================
