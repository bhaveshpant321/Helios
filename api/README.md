# Helios Trading Platform API

A high-performance REST API and WebSocket server for the Helios digital asset trading simulator, built with Go and PostgreSQL.

## 🚀 Features

- **RESTful API** with JWT authentication
- **Real-time WebSocket** updates for order books
- **PostgreSQL stored procedures** for business logic
- **LISTEN/NOTIFY** for efficient database-driven events
- **Concurrent request handling** with Gin framework
- **Graceful shutdown** and error handling

## 📋 Prerequisites

- **Go 1.21+**
- **PostgreSQL 14+**
- **Git** (for cloning the repository)

## 🛠️ Installation & Setup

### 1. Clone the Repository

```bash
cd Helios/api
```

### 2. Install Go Dependencies

```bash
go mod download
```

### 3. Database Setup

Make sure your PostgreSQL database is set up with:

- The schema from `../db/schema.sql`
- The stored procedures from `../db/procedures/*.sql`
- Seed data from `../db/seed_data.sql`

```bash
# From the project root
psql -U postgres -d helios -f db/schema.sql
psql -U postgres -d helios -f db/procedures/user_auth_procs.sql
psql -U postgres -d helios -f db/procedures/order_query_procs.sql
psql -U postgres -d helios -f db/procedures/matching_engine_procs.sql
psql -U postgres -d helios -f db/seed_data.sql
```

### 4. Configure Environment Variables

Copy the example environment file and update it with your settings:

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```env
PORT=8080
ENV=development

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=helios

JWT_SECRET=your_super_secret_jwt_key_change_this
JWT_EXPIRY_HOURS=24

CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5500
```

### 5. Run the Server

```bash
go run main.go
```

You should see:

```
✅ Database connection established
✅ Server started on http://localhost:8080
📡 WebSocket endpoint: ws://localhost:8080/ws/v1/market/:pair
🔔 Started listening for market updates from PostgreSQL
```

## 📡 API Endpoints

### Authentication (`/api/v1/auth`)

#### Register User

```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "john_trader",
  "email": "john@example.com",
  "password": "securepassword123"
}

Response: 201 Created
{
  "userId": 1005,
  "message": "User created successfully"
}
```

#### Login

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "john_trader",
  "password": "securepassword123"
}

Response: 200 OK
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1005,
    "username": "john_trader",
    "email": "john@example.com",
    "created_at": "2025-11-10T10:30:00Z"
  }
}
```

### Account (`/api/v1/account`) - Auth Required

#### Get Balances

```http
GET /api/v1/account/balances
Authorization: Bearer <your_jwt_token>

Response: 200 OK
[
  {
    "ticker_symbol": "USD",
    "balance": 10000.00
  },
  {
    "ticker_symbol": "BTC",
    "balance": 0.5
  }
]
```

### Orders (`/api/v1/orders`) - Auth Required

#### Place Order

```http
POST /api/v1/orders
Authorization: Bearer <your_jwt_token>
Content-Type: application/json

{
  "pair": "BTC/USD",
  "side": "BUY",
  "type": "LIMIT",
  "quantity": 0.5,
  "price": 50000.00
}

Response: 200 OK
{
  "status": "SUCCESS",
  "outcome": "FILLED",
  "order_id": 2009
}
```

#### Get Order History

```http
GET /api/v1/orders/history?pair=BTC/USD
Authorization: Bearer <your_jwt_token>

Response: 200 OK
[
  {
    "id": 2009,
    "user_id": 1005,
    "trading_pair_id": 1,
    "side": "BUY",
    "type": "LIMIT",
    "status": "FILLED",
    "price": 50000.00,
    "quantity": 0.5,
    "filled_quantity": 0.5,
    "created_at": "2025-11-10T11:00:00Z"
  }
]
```

#### Cancel Order

```http
DELETE /api/v1/orders/2009
Authorization: Bearer <your_jwt_token>

Response: 200 OK
{
  "message": "Order 2009 cancelled successfully"
}
```

### Market Data (`/api/v1/market`) - Public

#### Get Order Book

```http
GET /api/v1/market/orderbook/BTC/USD

Response: 200 OK
{
  "bids": [
    {
      "side": "BUY",
      "price": 40000.00,
      "total_quantity": 3.0
    }
  ],
  "asks": [
    {
      "side": "SELL",
      "price": 40100.00,
      "total_quantity": 2.0
    }
  ]
}
```

#### Get Trade History

```http
GET /api/v1/market/trades/BTC/USD?limit=100

Response: 200 OK
[
  {
    "id": 3001,
    "maker_order_id": 2003,
    "taker_order_id": 2007,
    "trading_pair_id": 1,
    "price": 39995.00,
    "quantity": 0.25,
    "executed_at": "2025-11-10T10:55:00Z"
  }
]
```

### WebSocket (`/ws/v1/market/:pair`)

Connect to real-time market updates:

```javascript
const ws = new WebSocket('ws://localhost:8080/ws/v1/market/BTC/USD');

ws.onopen = () => {
  console.log('Connected to market feed');
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  
  if (message.type === 'orderbook') {
    console.log('Order book update:', message.data);
    // message.data contains { bids: [...], asks: [...] }
  }
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('Disconnected from market feed');
};
```

## 🏗️ Project Structure

```

api/
├── main.go              # Application entry point
├── go.mod               # Go module dependencies
├── .env.example         # Example environment configuration
│
├── config/
│   └── config.go        # Configuration management
│
├── db/
│   └── database.go      # PostgreSQL connection & stored procedure calls
│
├── handlers/
│   ├── auth.go          # Authentication endpoints
│   ├── account.go       # Account management endpoints
│   ├── orders.go        # Order management endpoints
│   └── market.go        # Public market data endpoints
│
├── middleware/
│   ├── auth.go          # JWT authentication middleware
│   └── logger.go        # Request logging middleware
│
├── models/
│   └── models.go        # Data models and DTOs
│
├── utils/
│   └── auth.go          # Authentication utilities (JWT, bcrypt)
│
└── ws/
    ├── hub.go           # WebSocket hub and LISTEN/NOTIFY
    └── handler.go       # WebSocket HTTP handler
```

## 🔐 Authentication Flow

1. **Register**: User signs up with username, email, and password
2. **Password Hashing**: Password is hashed with bcrypt before storage
3. **Login**: User provides credentials
4. **JWT Token**: Server generates a JWT token valid for 24 hours
5. **Protected Routes**: Client includes token in `Authorization: Bearer <token>` header

## 🔔 Real-time Updates Flow

1. User places an order via `POST /api/v1/orders`
2. Stored procedure `sp_place_order` executes the matching engine
3. On successful match, procedure calls `pg_notify('market_update', '{"pair_id": 1}')`
4. Go WebSocket hub receives notification via LISTEN
5. Hub fetches updated order book from database
6. Hub broadcasts new order book to all connected WebSocket clients for that pair

## 🧪 Testing with cURL

### Register a new user
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@test.com","password":"password123"}'
```

### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123"}'
```

### Get balances (save token from login)
```bash
curl -X GET http://localhost:8080/api/v1/account/balances \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Place an order
```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"pair":"BTC/USD","side":"BUY","type":"LIMIT","quantity":0.1,"price":40000}'
```

## 🚀 Production Deployment

### Build for Production

```bash
go build -o helios-api main.go
```

### Environment Variables

Set these in your production environment:

- `ENV=production` - Enables production mode
- `JWT_SECRET` - Use a strong, random secret key
- `DB_PASSWORD` - Your production database password
- `CORS_ALLOWED_ORIGINS` - Your frontend domain(s)

### Running the Binary

```bash
./helios-api
```

## 🐛 Troubleshooting

### Database Connection Issues
- Verify PostgreSQL is running: `pg_isready`
- Check connection string in `.env`
- Ensure database and tables exist

### JWT Authentication Failures
- Check token expiry
- Verify JWT_SECRET matches in all environments
- Ensure `Authorization: Bearer <token>` header format

### WebSocket Connection Issues
- Check CORS settings
- Verify trading pair symbol format (e.g., "BTC/USD")
- Check browser console for connection errors

## 📝 License

Part of the Helios Trading Platform project.

## 👥 Contributors

- **Muneef Khan** - User & Account Management
- **Sandhya** - Order Management & Data Retrieval
- **Bhavesh** - Matching Engine
- **API Integration** - Complete Go backend

---

**Happy Trading! 🚀💰**
