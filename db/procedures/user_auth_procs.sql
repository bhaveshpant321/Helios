-- =========================================
-- PART 1: USER & ACCOUNT MANAGEMENT
-- Author: Muneef Khan
-- Project: Helios - Real-Time Digital Asset Trading Platform
-- =========================================
-- NOTE: Tables are defined in schema.sql
-- This file contains ONLY stored procedures/functions
-- =========================================

-- 1️⃣ FUNCTION: Create a new user and initial account
CREATE OR REPLACE FUNCTION sp_create_user(
    p_username VARCHAR,
    p_email VARCHAR,
    p_password_hash VARCHAR,
    p_initial_quote_asset_id INT,
    p_initial_balance DECIMAL
)
RETURNS BIGINT AS $$
DECLARE
    v_user_id BIGINT;
BEGIN
    -- Insert new user
    INSERT INTO users (username, email, password_hash)
    VALUES (p_username, p_email, p_password_hash)
    RETURNING id INTO v_user_id;

    -- Create initial account with starting balance (includes held_balance column)
    INSERT INTO accounts (user_id, asset_id, balance, held_balance)
    VALUES (v_user_id, p_initial_quote_asset_id, p_initial_balance, 0);

    -- Return user ID
    RETURN v_user_id;
END;
$$ LANGUAGE plpgsql;


-- 2️⃣ FUNCTION: Get user by email (for login verification)
CREATE OR REPLACE FUNCTION sp_get_user_by_email(
    p_email VARCHAR
)
RETURNS SETOF users AS $$
BEGIN
    RETURN QUERY
    SELECT * FROM users WHERE email = p_email;
END;
$$ LANGUAGE plpgsql;


-- 3️⃣ FUNCTION: Get all user balances
CREATE OR REPLACE FUNCTION sp_get_user_balances(
    p_user_id BIGINT
)
RETURNS TABLE(
    asset_id INT,
    ticker_symbol VARCHAR,
    name VARCHAR,
    balance DECIMAL,
    held_balance DECIMAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        a.id AS asset_id,
        a.ticker_symbol,
        a.name,
        acc.balance,
        acc.held_balance
    FROM accounts acc
    JOIN assets a ON acc.asset_id = a.id
    WHERE acc.user_id = p_user_id
    ORDER BY a.ticker_symbol;
END;
$$ LANGUAGE plpgsql;

-- =========================================
-- END OF USER & ACCOUNT PROCEDURES
-- =========================================
