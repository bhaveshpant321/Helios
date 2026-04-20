# Helios API Testing Guide

This guide helps you test the Helios API endpoints.

## Prerequisites

1. PostgreSQL database is running with schema and stored procedures
2. API server is running on `http://localhost:8080`
3. Install a REST client (Postman, Insomnia, or use cURL)

## Test Flow

### 1. Health Check

First, verify the server is running:

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "timestamp": 1699632000
}
```

### 2. Register a New User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "trader_alice",
    "email": "alice@trader.com",
    "password": "securepass123"
  }'
```

Expected response:
```json
{
  "userId": 1005,
  "message": "User created successfully"
}
```

### 3. Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "trader_alice",
    "password": "securepass123"
  }'
```

Expected response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1005,
    "username": "trader_alice",
    "email": "alice@trader.com",
    "created_at": "2025-11-10T12:00:00Z"
  }
}
```

**Save the token** for subsequent requests!

### 4. Get Account Balances

Replace `YOUR_TOKEN` with the JWT from the login response:

```bash
curl -X GET http://localhost:8080/api/v1/account/balances \
  -H "Authorization: Bearer YOUR_TOKEN"
```

Expected response:
```json
[
  {
    "ticker_symbol": "USD",
    "balance": 10000.00
  }
]
```

### 5. View Order Book (Public)

```bash
curl -X GET http://localhost:8080/api/v1/market/orderbook/BTC/USD
```

Expected response:
```json
{
  "bids": [
    {
      "side": "BUY",
      "price": 40000.00,
      "total_quantity": 3.00000000
    }
  ],
  "asks": [
    {
      "side": "SELL",
      "price": 40100.00,
      "total_quantity": 2.00000000
    }
  ]
}
```

### 6. Place a LIMIT Order

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "pair": "BTC/USD",
    "side": "BUY",
    "type": "LIMIT",
    "quantity": 0.5,
    "price": 39000.00
  }'
```

Expected response:
```json
{
  "status": "SUCCESS",
  "outcome": "OPEN",
  "order_id": 2010
}
```

### 7. Place a MARKET Order

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "pair": "BTC/USD",
    "side": "BUY",
    "type": "MARKET",
    "quantity": 0.1
  }'
```

Expected response:
```json
{
  "status": "SUCCESS",
  "outcome": "FILLED",
  "order_id": 2011
}
```

### 8. Get Order History

```bash
curl -X GET "http://localhost:8080/api/v1/orders/history?pair=BTC/USD" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

Expected response:
```json
[
  {
    "id": 2010,
    "user_id": 1005,
    "trading_pair_id": 1,
    "side": "BUY",
    "type": "LIMIT",
    "status": "OPEN",
    "price": 39000.00,
    "quantity": 0.5,
    "filled_quantity": 0.0,
    "created_at": "2025-11-10T12:05:00Z"
  }
]
```

### 9. Cancel an Order

```bash
curl -X DELETE http://localhost:8080/api/v1/orders/2010 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

Expected response:
```json
{
  "message": "Order 2010 cancelled successfully"
}
```

### 10. Get Trade History (Public)

```bash
curl -X GET "http://localhost:8080/api/v1/market/trades/BTC/USD?limit=10"
```

Expected response:
```json
[
  {
    "id": 3001,
    "maker_order_id": 2003,
    "taker_order_id": 2007,
    "trading_pair_id": 1,
    "price": 39995.00,
    "quantity": 0.25000000,
    "executed_at": "2025-11-10T10:55:00Z"
  }
]
```

## WebSocket Testing

### Using JavaScript (Browser Console)

```javascript
// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8080/ws/v1/market/BTC/USD');

ws.onopen = () => {
  console.log('✅ Connected to BTC/USD market feed');
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('📡 Market update:', message);
};

ws.onerror = (error) => {
  console.error('❌ WebSocket error:', error);
};

ws.onclose = () => {
  console.log('🔌 Disconnected');
};
```

### Using wscat (Command Line)

Install wscat:
```bash
npm install -g wscat
```

Connect:
```bash
wscat -c ws://localhost:8080/ws/v1/market/BTC/USD
```

You'll receive real-time order book updates whenever orders are placed or cancelled.

## Common Error Scenarios

### 401 Unauthorized
```json
{
  "error": "unauthorized",
  "message": "Authorization header required"
}
```
**Solution:** Include valid JWT token in Authorization header

### 400 Bad Request
```json
{
  "error": "validation_error",
  "message": "Price is required for LIMIT orders"
}
```
**Solution:** Check request body matches expected format

### 409 Conflict
```json
{
  "error": "registration_failed",
  "message": "Username or email already exists"
}
```
**Solution:** Use a different username or email

### Insufficient Funds
```json
{
  "status": "ERROR",
  "message": "Insufficient funds. Estimated Cost: 5000.00, Available: 1000.00"
}
```
**Solution:** Reduce order quantity or deposit more funds

## Testing with Postman

1. Import the collection from `postman_collection.json`
2. Set environment variable `base_url` to `http://localhost:8080`
3. After login, save the JWT token to environment variable `auth_token`
4. Run the collection tests in order

## Automated Testing Script

Create a file `test_api.sh`:

```bash
#!/bin/bash

BASE_URL="http://localhost:8080"

echo "🧪 Testing Helios API"
echo "===================="
echo ""

# Health check
echo "1. Health Check..."
curl -s $BASE_URL/health | jq
echo ""

# Register
echo "2. Register User..."
REGISTER_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test_'$RANDOM'","email":"test'$RANDOM'@test.com","password":"test123"}')
echo $REGISTER_RESPONSE | jq
echo ""

# Login
echo "3. Login..."
TOKEN_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"sandhya_test","password":"sandhya_hash"}')
TOKEN=$(echo $TOKEN_RESPONSE | jq -r '.token')
echo "Token obtained: ${TOKEN:0:20}..."
echo ""

# Get Balances
echo "4. Get Balances..."
curl -s -X GET $BASE_URL/api/v1/account/balances \
  -H "Authorization: Bearer $TOKEN" | jq
echo ""

# Get Order Book
echo "5. Get Order Book..."
curl -s -X GET $BASE_URL/api/v1/market/orderbook/BTC/USD | jq
echo ""

echo "✅ All tests completed!"
```

Run with: `chmod +x test_api.sh && ./test_api.sh`

## Performance Testing

Use Apache Bench to test API performance:

```bash
# Test login endpoint
ab -n 1000 -c 10 -p login_data.json -T application/json \
  http://localhost:8080/api/v1/auth/login

# Test order book (public endpoint)
ab -n 10000 -c 50 http://localhost:8080/api/v1/market/orderbook/BTC/USD
```

## Next Steps

- Test with your frontend application
- Monitor WebSocket connections in real-time
- Test concurrent order placement
- Verify database state after operations
- Test error handling and edge cases
