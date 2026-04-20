# Seed Data Analysis

## Quick Answer

**Is the seed data complete?** ⚠️ **Mostly, but has issues**

**Do we need it in production?** ❌ **NO - This is TEST DATA ONLY**

---

## Detailed Analysis

### 🔴 CRITICAL ISSUES

#### 1. **Fake Password Hashes**
```sql
(1001, 'sandhya_test', 'sandhya@test.com', 'sandhya_hash'),
```

**Problem:** These aren't real bcrypt hashes. The API will reject them during login.

**Real bcrypt hash looks like:**
```
$2a$10$N9qo8uLOickgx2ZMRZoMyeIjbJhV8kdpqYJ2N9qPbsxD4Z3VCp4xC
```

**Impact:** 
- ❌ Cannot test login functionality
- ❌ sp_get_user_by_username works, but password verification will fail
- ✅ Registration still works (creates new users with proper hashes)

---

#### 2. **Incorrect Trade Reference** (Already documented in DATABASE_FIXES.md)
```sql
(3004, 2, 2003, 2004, 2500.00, 10.000000, NOW() - INTERVAL '1 minute');
```

**Problem:**
- Trade references trading_pair_id = 2 (ETH/USD)
- But maker_order_id = 2003 and taker_order_id = 2004 are BTC/USD orders
- Orders 2003 and 2004 don't match trading pair 2

**Impact:**
- ⚠️ Data integrity violation
- ⚠️ Reports will show incorrect trade data
- Won't crash but gives wrong results

---

#### 3. **Held Balance Not Reflected**
```sql
INSERT INTO accounts (user_id, asset_id, balance, held_balance) VALUES
(1001, 1, 50000.00, 0),  -- All held_balance = 0
```

**Problem:**
- Several orders are marked as 'OPEN' (2001, 2002, 2004, 2005)
- These orders should have held funds
- But all `held_balance` values are 0

**What should happen:**
- Order 2001: BUY 1 BTC @ $40,000 → Should hold $40,000 USD
- Order 2004: SELL 1.5 BTC @ $40,100 → Should hold 1.5 BTC

**Impact:**
- ⚠️ Data inconsistency
- Users appear to have more available funds than they should
- When cancelling these "existing" orders, funds won't be released (nothing to release)

---

### ✅ WHAT'S CORRECT

1. **Assets (Section 1)** - ✅ Perfect
   - Proper ticker symbols
   - Correct decimal places
   - All major assets covered

2. **Trading Pairs (Section 2)** - ✅ Perfect
   - Logical pairs (BTC/USD, ETH/USD, SOL/BTC)
   - Correct base/quote relationships
   - Proper symbols

3. **Order Structure** - ✅ Good
   - Mix of OPEN, FILLED, CANCELLED, PARTIALLY_FILLED statuses
   - Good for testing order history
   - Realistic quantities and prices

4. **Fees Table** - ✅ Complete
   - Proper maker/taker fee distinction
   - Links to correct trades and users

5. **Sequence Management** - ✅ Perfect
   - All sequences properly set with `setval()`
   - Prevents ID conflicts with new inserts

---

## Should You Keep Seed Data in Production?

### ❌ **NO - Remove for Production**

**Why:**

1. **Security Risk**
   - Contains test users with known credentials
   - Fake passwords are a security vulnerability
   - Test accounts could be exploited

2. **Data Pollution**
   - Test orders clutter the order book
   - Fake trades skew statistics
   - Test users aren't real customers

3. **Not Required**
   - Only **Assets** and **Trading Pairs** are needed
   - User accounts should come from real registrations
   - Orders/trades come from real activity

---

## Recommended Production Seed Data

Create a new file: `db/seed_data_production.sql`

```sql
-- ==========================================================
-- Helios: PRODUCTION Seed Data
-- Only reference data required for platform operation
-- ==========================================================

BEGIN;

-- ----------------------------------------------------------
-- 1. ASSETS (Required for platform to function)
-- ----------------------------------------------------------
INSERT INTO assets (ticker_symbol, name, decimals) VALUES
('USD', 'United States Dollar', 2),
('BTC', 'Bitcoin', 8),
('ETH', 'Ethereum', 6),
('SOL', 'Solana', 4),
('USDT', 'Tether USD', 2),
('USDC', 'USD Coin', 2);

-- ----------------------------------------------------------
-- 2. TRADING_PAIRS (Define available markets)
-- ----------------------------------------------------------
INSERT INTO trading_pairs (base_asset_id, quote_asset_id, symbol) VALUES
-- USD pairs
(2, 1, 'BTC/USD'),
(3, 1, 'ETH/USD'),
(4, 1, 'SOL/USD'),

-- USDT pairs
(2, 5, 'BTC/USDT'),
(3, 5, 'ETH/USDT'),

-- USDC pairs
(2, 6, 'BTC/USDC'),
(3, 6, 'ETH/USDC');

COMMIT;
```

**That's it!** No test users, no fake orders, no test trades.

---

## Testing vs Production Strategy

### For Development/Testing:
```bash
# Use full seed data with test users
psql -d helios -f db/schema.sql
psql -d helios -f db/procedures/user_auth_procs.sql
psql -d helios -f db/procedures/order_query_procs.sql
psql -d helios -f db/procedures/matching_engine_procs.sql
psql -d helios -f db/seed_data.sql  # ← Full test data
```

### For Production:
```bash
# Use minimal production seed data
psql -d helios_prod -f db/schema.sql
psql -d helios_prod -f db/procedures/user_auth_procs.sql
psql -d helios_prod -f db/procedures/order_query_procs.sql
psql -d helios_prod -f db/procedures/matching_engine_procs.sql
psql -d helios_prod -f db/seed_data_production.sql  # ← Only reference data
```

---

## Fixing Test Seed Data (Optional)

If you want realistic test data for development:

### Fix 1: Generate Real Password Hashes

```bash
# In PowerShell or terminal with Go installed
cd api
go run -c 'package main
import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    hash, _ := bcrypt.GenerateFromPassword([]byte("test123"), 10)
    fmt.Println(string(hash))
}'
```

Or use this online: https://bcrypt-generator.com/
- Password: `test123`
- Rounds: 10

Then update seed_data.sql:
```sql
INSERT INTO users (id, username, email, password_hash) VALUES
(1001, 'sandhya_test', 'sandhya@test.com', '$2a$10$actual_hash_here'),
```

### Fix 2: Fix the ETH/USD Trade

Either remove it or change:
```sql
-- Option A: Make it a real ETH/USD trade (needs ETH orders first)
-- Option B: Change to BTC/USD pair
(3004, 1, 2003, 2004, 40000.00, 0.5, NOW() - INTERVAL '1 minute');
--     ^ Changed 2 to 1 (BTC/USD)
```

### Fix 3: Calculate Held Balances

For open orders, calculate what should be held:
```sql
-- Order 2001: BUY 1 BTC @ $40,000 → Hold $40,000
-- Order 2002: BUY 2 BTC @ $40,000 → Hold $80,000
-- Order 2003: BUY 5 BTC @ $39,990, filled 2 → Hold $119,970 (for remaining 3)
-- Order 2004: SELL 1.5 BTC → Hold 1.5 BTC
-- Order 2005: SELL 0.5 BTC → Hold 0.5 BTC
-- Order 2006: SELL 3 BTC, filled 1 → Hold 2 BTC (for remaining 2)

INSERT INTO accounts (user_id, asset_id, balance, held_balance) VALUES
-- Sandhya (1001): Has order 2001 (hold $40,000)
(1001, 1, 10000.00, 40000.00),   -- $50k total, $40k held
(1001, 2, 5.00000000, 0),

-- Muneef (1002): Has order 2004 (hold 1.5 BTC)
(1002, 1, 10000.00, 0),
(1002, 2, 1.50000000, 1.50000000),  -- 3 BTC total, 1.5 held

-- Bhavesh (1003): Has orders 2005 + 2006 (hold 2.5 BTC total)
(1003, 1, 8000.00, 0),
(1003, 2, 2.50000000, 2.50000000),  -- 5 BTC total, 2.5 held

-- Market Maker (1004): Has orders 2002 + 2003 (hold $199,970)
(1004, 1, 300030.00, 199970.00),   -- $500k total, ~$200k held
(1004, 2, 100.00000000, 0),
(1004, 3, 2000.000000, 0),
(1004, 4, 5000.0000, 0);
```

---

## Summary

| Aspect | Status | Action |
|--------|--------|--------|
| **Assets & Trading Pairs** | ✅ Perfect | Keep in production |
| **Test Users** | ❌ Fake passwords | Fix or remove for prod |
| **Test Orders** | ⚠️ Inconsistent held balances | Fix for better testing |
| **Test Trades** | ⚠️ One incorrect reference | Fix trade 3004 |
| **Fees** | ✅ Good | Remove for production |
| **Overall Completeness** | 80% | Functional but needs fixes |

### Final Recommendation:

1. **For Now (Development):**
   - ✅ Current seed data works for basic testing
   - ⚠️ Just know you can't test login (passwords are fake)
   - ✅ You can still test registration, orders, trades

2. **For Better Testing:**
   - Fix password hashes
   - Fix held_balance calculations
   - Fix trade 3004

3. **For Production:**
   - Create `seed_data_production.sql` with ONLY assets and trading pairs
   - Remove all test users, orders, trades, fees
   - Let real users register and trade

---

**Bottom Line:** The seed data is good enough to test your API, but it's purely for development. Production should start with a clean slate (just assets and trading pairs).
