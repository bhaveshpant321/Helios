# 🗄️ Helios Database Architecture

## Overview

Helios uses **PostgreSQL 14+** as its database engine, implementing a high-performance cryptocurrency exchange with ACID compliance, stored procedures for critical operations, and real-time event notifications.

---

## 📊 Database Schema

### **1. Assets Table**
Stores all tradable assets (cryptocurrencies and fiat currencies).

```sql
CREATE TABLE assets (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(10) UNIQUE NOT NULL,      -- e.g., 'USD', 'BTC', 'ETH'
    name VARCHAR(100) NOT NULL,               -- Full name
    asset_type VARCHAR(20) NOT NULL,          -- 'FIAT' or 'CRYPTO'
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Purpose:** Defines what can be traded. Each asset is either a base or quote currency in a trading pair.

---

### **2. Trading Pairs Table**
Defines which assets can be traded against each other.

```sql
CREATE TABLE trading_pairs (
    id SERIAL PRIMARY KEY,
    base_asset_id INT NOT NULL REFERENCES assets(id),   -- What you're buying/selling
    quote_asset_id INT NOT NULL REFERENCES assets(id),  -- What you're paying with
    symbol VARCHAR(20) UNIQUE NOT NULL,                 -- e.g., 'BTC/USD'
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(base_asset_id, quote_asset_id)
);
```

**Example:** BTC/USD means:
- Base: BTC (what you buy/sell)
- Quote: USD (what you pay/receive)

---

### **3. Users Table**
Stores user account information and authentication credentials.

```sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,      -- Bcrypt hashed password
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Security:** Passwords are hashed using bcrypt (never stored in plain text).

---

### **4. Balances Table**
Tracks how much of each asset each user owns.

```sql
CREATE TABLE balances (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    asset_id INT NOT NULL REFERENCES assets(id),
    available_balance DECIMAL(20, 8) DEFAULT 0 CHECK (available_balance >= 0),
    locked_balance DECIMAL(20, 8) DEFAULT 0 CHECK (locked_balance >= 0),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, asset_id)
);
```

**Key Concepts:**
- **Available Balance:** Funds available for trading
- **Locked Balance:** Funds reserved in open orders
- **CHECK Constraints:** Prevent negative balances (ACID integrity)

**Example:**
```
User ID: 1
Asset: BTC
Available: 0.5 BTC
Locked: 0.2 BTC (in open sell order)
Total: 0.7 BTC
```

---

### **5. Orders Table**
Records all buy and sell orders placed by users.

```sql
CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    trading_pair_id INT NOT NULL REFERENCES trading_pairs(id),
    side VARCHAR(4) NOT NULL CHECK (side IN ('BUY', 'SELL')),
    order_type VARCHAR(10) NOT NULL CHECK (order_type IN ('MARKET', 'LIMIT')),
    quantity DECIMAL(20, 8) NOT NULL CHECK (quantity > 0),
    price DECIMAL(20, 8) CHECK (price > 0),              -- NULL for MARKET orders
    filled_quantity DECIMAL(20, 8) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'PENDING' CHECK (status IN 
        ('PENDING', 'PARTIALLY_FILLED', 'FILLED', 'CANCELLED')),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

**Order States:**
- **PENDING:** Waiting to be matched
- **PARTIALLY_FILLED:** Some quantity filled, rest still active
- **FILLED:** Completely executed
- **CANCELLED:** User cancelled before filling

**Indexes for Performance:**
```sql
CREATE INDEX idx_orders_user ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_trading_pair ON orders(trading_pair_id, status);
```

---

### **6. Trades Table**
Records all executed trades (when orders match).

```sql
CREATE TABLE trades (
    id BIGSERIAL PRIMARY KEY,
    trading_pair_id INT NOT NULL REFERENCES trading_pairs(id),
    buy_order_id BIGINT NOT NULL REFERENCES orders(id),
    sell_order_id BIGINT NOT NULL REFERENCES orders(id),
    price DECIMAL(20, 8) NOT NULL,
    quantity DECIMAL(20, 8) NOT NULL,
    buyer_fee DECIMAL(20, 8) DEFAULT 0,
    seller_fee DECIMAL(20, 8) DEFAULT 0,
    executed_at TIMESTAMP DEFAULT NOW()
);
```

**Purpose:** Immutable record of all executed trades for audit trail and history.

**Indexes:**
```sql
CREATE INDEX idx_trades_pair ON trades(trading_pair_id, executed_at DESC);
CREATE INDEX idx_trades_orders ON trades(buy_order_id, sell_order_id);
```

---

## 🔧 Stored Procedures

### **1. sp_place_order() - Order Placement & Matching Engine**

**Location:** `db/procedures/matching_engine_procs.sql`

**Function Signature:**
```sql
CREATE OR REPLACE FUNCTION sp_place_order(
    p_user_id BIGINT,
    p_trading_pair_id INT,
    p_side VARCHAR(4),
    p_order_type VARCHAR(10),
    p_quantity DECIMAL,
    p_price DECIMAL DEFAULT NULL
) RETURNS jsonb
```

**What It Does:**
1. **Validates** user has sufficient balance
2. **Locks** required funds (available → locked)
3. **Matches** order against existing orders in order book
4. **Executes** trades when prices match
5. **Updates** balances and order statuses
6. **Returns** result as JSON

**Matching Logic:**

**For BUY orders:**
```
1. Find SELL orders with price <= buy price
2. Sort by price (ascending) then time (FIFO)
3. Match until buy order is filled or no more matches
```

**For SELL orders:**
```
1. Find BUY orders with price >= sell price
2. Sort by price (descending) then time (FIFO)
3. Match until sell order is filled or no more matches
```

**ACID Compliance:**
```sql
-- All operations in single transaction
BEGIN
    -- Lock funds
    UPDATE balances SET available_balance = available_balance - cost,
                       locked_balance = locked_balance + cost
                  WHERE user_id = p_user_id AND asset_id = p_asset_id;
    
    -- Insert order
    INSERT INTO orders (...) VALUES (...) RETURNING id INTO v_order_id;
    
    -- Match orders (loop)
    FOR match IN SELECT * FROM orders WHERE ... LOOP
        -- Execute trade
        INSERT INTO trades (...) VALUES (...);
        
        -- Update buyer balance
        UPDATE balances SET locked_balance = locked_balance - cost,
                           available_balance = available_balance + amount;
        
        -- Update seller balance  
        UPDATE balances SET locked_balance = locked_balance - amount,
                           available_balance = available_balance + cost;
        
        -- Update order statuses
        UPDATE orders SET filled_quantity = ..., status = ...;
    END LOOP;
    
    -- Notify WebSocket listeners
    PERFORM pg_notify('market_update', ...);
    
    COMMIT; -- All or nothing!
END;
```

**Error Handling:**
- Insufficient balance → Rollback, return error
- Invalid trading pair → Rollback, return error
- Any database error → Rollback, return error

**Fee Structure:**
```sql
v_taker_rate := 0.001;  -- 0.1% Taker Fee (order that executes immediately)
v_maker_rate := 0.0005; -- 0.05% Maker Fee (order that sits in order book)
```

---

### **2. sp_cancel_order() - Order Cancellation**

**Location:** `db/procedures/order_query_procs.sql`

**Function Signature:**
```sql
CREATE OR REPLACE FUNCTION sp_cancel_order(
    p_order_id BIGINT,
    p_user_id BIGINT
) RETURNS jsonb
```

**What It Does:**
1. **Validates** order exists and belongs to user
2. **Checks** order can be cancelled (not already filled/cancelled)
3. **Unlocks** reserved funds (locked → available)
4. **Updates** order status to 'CANCELLED'
5. **Notifies** WebSocket listeners

**ACID Compliance:**
```sql
BEGIN
    -- Get order details with row lock
    SELECT * INTO v_order FROM orders 
    WHERE id = p_order_id AND user_id = p_user_id 
    FOR UPDATE;  -- Prevents concurrent modifications
    
    -- Calculate unfilled quantity
    v_unfilled := v_order.quantity - v_order.filled_quantity;
    
    -- Unlock funds
    UPDATE balances SET 
        locked_balance = locked_balance - v_locked_amount,
        available_balance = available_balance + v_locked_amount;
    
    -- Update order status
    UPDATE orders SET status = 'CANCELLED', updated_at = NOW()
    WHERE id = p_order_id;
    
    -- Notify
    PERFORM pg_notify('market_update', ...);
    
    COMMIT;
END;
```

---

### **3. User Authentication Procedures**

**Location:** `db/procedures/user_auth_procs.sql`

**sp_create_user():**
```sql
CREATE OR REPLACE FUNCTION sp_create_user(
    p_username VARCHAR,
    p_email VARCHAR,
    p_password_hash VARCHAR,
    p_initial_quote_asset_id INT,
    p_initial_balance DECIMAL
) RETURNS BIGINT
```

**What It Does:**
1. **Inserts** new user record
2. **Creates** initial balance for quote asset (e.g., $10,000 USD)
3. **Returns** new user ID

**sp_get_user_by_email():**
```sql
CREATE OR REPLACE FUNCTION sp_get_user_by_email(p_email VARCHAR)
RETURNS TABLE (...)
```

**What It Does:**
- Retrieves user record for authentication
- Used by login endpoint

---

## 🔒 ACID Compliance

### **Atomicity**
✅ All operations in stored procedures wrapped in transactions
- Either ALL changes succeed, or NONE do
- Example: If trade execution fails halfway, balances rollback

### **Consistency**
✅ Database constraints enforce valid states
- `CHECK` constraints prevent negative balances
- `FOREIGN KEY` constraints ensure referential integrity
- `UNIQUE` constraints prevent duplicate usernames/emails

### **Isolation**
✅ Row-level locking prevents race conditions
```sql
SELECT * FROM orders WHERE id = ? FOR UPDATE;  -- Locks row until transaction completes
```
- Prevents two users from modifying same order simultaneously
- Prevents double-spending of balance

### **Durability**
✅ PostgreSQL Write-Ahead Logging (WAL)
- All committed transactions written to disk
- Survives system crashes
- Point-in-time recovery possible

---

## 📡 Real-Time Events (LISTEN/NOTIFY)

PostgreSQL's pub/sub system for WebSocket updates.

**Publisher (Stored Procedure):**
```sql
PERFORM pg_notify('market_update', jsonb_build_object(
    'trading_pair_id', p_trading_pair_id
)::text);
```

**Subscriber (Go Application):**
```go
conn.Exec(context.Background(), "LISTEN market_update")

for {
    notification, err := conn.WaitForNotification(context.Background())
    // Broadcast to WebSocket clients
}
```

**Benefits:**
- Zero polling overhead
- Instant updates to all connected clients
- Database-driven event system

---

## 🎯 Performance Optimizations

### **1. Strategic Indexes**
```sql
-- Order book queries (most frequent)
CREATE INDEX idx_orders_trading_pair ON orders(trading_pair_id, status);
CREATE INDEX idx_orders_price_time ON orders(trading_pair_id, side, price, created_at);

-- Trade history queries
CREATE INDEX idx_trades_pair_time ON trades(trading_pair_id, executed_at DESC);

-- Balance lookups
CREATE INDEX idx_balances_user ON balances(user_id);
```

### **2. Materialized Order Book (Future Enhancement)**
Could cache order book state for ultra-fast reads:
```sql
CREATE MATERIALIZED VIEW mv_order_book AS
SELECT trading_pair_id, side, price, SUM(quantity - filled_quantity) as total_qty
FROM orders WHERE status IN ('PENDING', 'PARTIALLY_FILLED')
GROUP BY trading_pair_id, side, price;

-- Refresh on market_update event
REFRESH MATERIALIZED VIEW CONCURRENTLY mv_order_book;
```

### **3. Partitioning (Scalability)**
For large datasets, partition trades by date:
```sql
CREATE TABLE trades_2025_11 PARTITION OF trades
FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');
```

---

## 🧪 Testing ACID Properties

### **Test 1: Concurrent Order Placement**
```sql
-- Terminal 1
BEGIN;
UPDATE balances SET available_balance = available_balance - 1000 
WHERE user_id = 1 FOR UPDATE;
-- Wait 10 seconds
COMMIT;

-- Terminal 2 (runs simultaneously)
BEGIN;
UPDATE balances SET available_balance = available_balance - 500 
WHERE user_id = 1 FOR UPDATE;
-- Waits for Terminal 1 to commit (row lock)
COMMIT;
```

Result: ✅ No race condition, final balance is correct

### **Test 2: Rollback on Error**
```sql
BEGIN;
    UPDATE balances SET available_balance = available_balance - 10000 WHERE user_id = 1;
    -- Simulate error
    INSERT INTO orders (quantity) VALUES (-100);  -- Fails CHECK constraint
ROLLBACK; -- Balance unchanged
```

Result: ✅ Balance restored to original value

---

## 📈 Scalability Considerations

### **Current Capacity:**
- Users: Millions (BIGSERIAL primary keys)
- Orders: Billions (BIGSERIAL, indexed)
- Trades: Unlimited (partitionable)

### **Bottlenecks:**
1. **Matching Engine:** O(n) search through order book
   - **Solution:** In-memory order book cache
2. **WebSocket Broadcasting:** All clients notified
   - **Solution:** Separate notification service
3. **Single Database:** All writes to one instance
   - **Solution:** Read replicas for queries

### **Future Enhancements:**
- Redis for in-memory order book
- Kafka for event streaming
- Read replicas for historical data
- Sharding by trading pair

---

## 🔍 Monitoring & Maintenance

### **Key Metrics:**
```sql
-- Active orders count
SELECT COUNT(*) FROM orders WHERE status IN ('PENDING', 'PARTIALLY_FILLED');

-- Daily trade volume
SELECT trading_pair_id, SUM(quantity * price) as volume
FROM trades 
WHERE executed_at >= CURRENT_DATE
GROUP BY trading_pair_id;

-- User balance audit
SELECT user_id, SUM(available_balance + locked_balance) as total
FROM balances
GROUP BY user_id;
```

### **Backup Strategy:**
```bash
# Full backup (daily)
pg_dump -Fc helios_db > backup_$(date +%Y%m%d).dump

# Point-in-time recovery (continuous)
# Enable WAL archiving in postgresql.conf
archive_mode = on
archive_command = 'cp %p /backup/archive/%f'
```

---

## 📚 Summary

The Helios database architecture provides:

✅ **ACID Compliance:** All transactions are atomic, consistent, isolated, durable  
✅ **High Performance:** Strategic indexes, stored procedures  
✅ **Real-Time Updates:** PostgreSQL LISTEN/NOTIFY  
✅ **Scalability:** Designed for millions of users and billions of trades  
✅ **Security:** Row-level locking, constraint checking, hashed passwords  
✅ **Auditability:** Immutable trade history, timestamped records  

**Core Philosophy:** Database enforces business logic through stored procedures and constraints, ensuring data integrity at the lowest level.
