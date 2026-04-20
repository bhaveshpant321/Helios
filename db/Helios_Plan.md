
# Helios Trading Platform: Stored Procedure Contracts

This document defines the official interfaces for all PostgreSQL stored procedures. All backend development must adhere to these signatures.

---

## Part 1: User & Account Management  
**Owner:** Muneef

Procedures for user registration, authentication, and balance retrieval.

### `sp_create_user`
**Purpose:** Create a new user and their initial quote currency account (e.g., USD).

- **Signature:**
  ```sql
  FUNCTION sp_create_user(
    p_username VARCHAR,
    p_email VARCHAR,
    p_password_hash VARCHAR,
    p_initial_quote_asset_id INT,
    p_initial_balance DECIMAL
  ) RETURNS BIGINT
  ```
- **Description:**  
  Inserts a new record into the `users` table and creates an associated entry in the `accounts` table with a starting balance for the primary quote currency.
- **Parameters:**
  - `p_username`: User's chosen username
  - `p_email`: User's email address
  - `p_password_hash`: Pre-hashed password
  - `p_initial_quote_asset_id`: Asset ID for default currency (e.g., USD)
  - `p_initial_balance`: Starting paper trading balance
- **Returns:**  
  The ID of the newly created user (`BIGINT`)

---

### `sp_get_user_by_username`
**Purpose:** Retrieve user details for login verification.

- **Signature:**
  ```sql
  FUNCTION sp_get_user_by_username(
    p_username VARCHAR
  ) RETURNS SETOF users
  ```
- **Description:**  
  Fetches the full user record based on username for password verification.
- **Returns:**  
  A full row from the `users` table

---

### `sp_get_user_balances`
**Purpose:** Fetch a user's balances for all assets.

- **Signature:**
  ```sql
  FUNCTION sp_get_user_balances(
    p_user_id BIGINT
  ) RETURNS TABLE(ticker_symbol VARCHAR, balance DECIMAL)
  ```
- **Description:**  
  Joins `accounts` and `assets` tables to return all assets a user holds and their balances.
- **Returns:**  
  Table of `ticker_symbol` and `balance` for the user

---

## Part 2: Order Management & Data Retrieval  
**Owner:** Sandhya

Procedures for canceling orders and querying data for the UI.

### `sp_cancel_order`
**Purpose:** Cancel an active, open order.

- **Signature:**
  ```sql
  PROCEDURE sp_cancel_order(
    p_order_id BIGINT,
    p_user_id BIGINT
  )
  ```
- **Description:**  
  Sets an order's status to `CANCELLED`. Verifies that `p_user_id` matches the order's owner and that the order is `OPEN` or `PARTIALLY_FILLED`.
- **Returns:**  
  Nothing. Raises exception on failure (e.g., "Order not found or permission denied").

---

### `sp_get_user_order_history`
**Purpose:** Retrieve a user's complete order history for a trading pair.

- **Signature:**
  ```sql
  FUNCTION sp_get_user_order_history(
    p_user_id BIGINT,
    p_trading_pair_id INT
  ) RETURNS SETOF orders
  ```
- **Description:**  
  Fetches all orders for a user and trading pair.
- **Returns:**  
  Set of rows from the `orders` table

---

### `sp_get_order_book`
**Purpose:** Retrieve the aggregated, public order book for a trading pair.

- **Signature:**
  ```sql
  FUNCTION sp_get_order_book(
    p_trading_pair_id INT
  ) RETURNS TABLE(side order_side, price DECIMAL, total_quantity DECIMAL)
  ```
- **Description:**  
  Queries all `OPEN` or `PARTIALLY_FILLED` orders, groups by price, and sums remaining quantity.
- **Returns:**  
  Table of `side`, `price`, and aggregated `total_quantity`

---

### `sp_get_trade_history`
**Purpose:** Retrieve recent public trades for a trading pair.

- **Signature:**
  ```sql
  FUNCTION sp_get_trade_history(
    p_trading_pair_id INT,
    p_limit INT DEFAULT 100
  ) RETURNS SETOF trades
  ```
- **Description:**  
  Fetches the last N trades for a pair, ordered by execution time.
- **Returns:**  
  Set of recent rows from the `trades` table

---

## Part 3: Core Matching Engine  
**Owner:** Bhavesh

The heart of the exchange.

### `sp_place_order`
**Purpose:** Submit a new order and trigger matching logic.

- **Signature:**
  ```sql
  FUNCTION sp_place_order(
    p_user_id BIGINT,
    p_trading_pair_id INT,
    p_side order_side,
    p_type order_type,
    p_quantity DECIMAL,
    p_price DECIMAL DEFAULT NULL
  ) RETURNS JSONB
  ```
- **Description:**  
  Main transactional procedure: validates and holds funds, matches orders, creates trades, updates balances/fees, and inserts/updates order status.
- **Returns:**  
  JSONB object detailing the outcome.

  **Example (Success):**
  ```json
  {
    "status": "SUCCESS",
    "outcome": "PARTIALLY_FILLED",
    "order_id": 105,
    "trades_executed": [
      {"trade_id": 210, "price": "50000.00", "quantity": "0.5"},
      {"trade_id": 211, "price": "50000.10", "quantity": "0.2"}
    ]
  }
  ```

  **Example (Failure):**
  ```json
  {
    "status": "ERROR",
    "message": "Insufficient funds."
  }
  ```

---
