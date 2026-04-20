# 🚀 Helios API Architecture

## Overview

Helios API is a RESTful HTTP server built with **Go 1.21+** and the **Gin framework**, providing endpoints for user authentication, order management, market data, and real-time WebSocket updates.

---

## 🏗️ Architecture Layers

```
┌─────────────────────────────────────────┐
│         HTTP/WebSocket Clients          │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│         Middleware Layer                │
│  • CORS                                 │
│  • JWT Authentication                   │
│  • Request Logging                      │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│         Handlers Layer                  │
│  • AuthHandler (login/register)         │
│  • OrderHandler (place/cancel)          │
│  • MarketHandler (orderbook/trades)     │
│  • AccountHandler (balances)            │
│  • WSHandler (WebSocket)                │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│         Database Layer                  │
│  • Connection Pooling (pgxpool)         │
│  • Stored Procedure Calls               │
│  • LISTEN/NOTIFY Subscription           │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│         PostgreSQL Database             │
└─────────────────────────────────────────┘
```

---

## 📁 Project Structure

```
api/
├── main.go                 # Application entry point, server setup
├── config/
│   └── config.go          # Environment configuration loading
├── db/
│   └── database.go        # Database connection and query methods
├── handlers/
│   ├── auth.go            # Registration and login endpoints
│   ├── orders.go          # Order placement and management
│   ├── market.go          # Market data endpoints
│   └── account.go         # User account endpoints
├── middleware/
│   ├── auth.go            # JWT authentication middleware
│   └── logger.go          # HTTP request logging
├── models/
│   └── models.go          # Request/response data structures
├── utils/
│   └── auth.go            # JWT generation and password hashing
└── ws/
    ├── hub.go             # WebSocket connection manager
    └── handler.go         # WebSocket endpoint handler
```

---

## 🔧 Core Components

### **1. Main Server (main.go)**

**Initialization Flow:**
```go
func main() {
    // 1. Load configuration from .env
    cfg := config.Load()
    
    // 2. Connect to PostgreSQL database
    database := db.NewDatabase(cfg)
    
    // 3. Initialize JWT utilities
    utils.InitJWTConfig(cfg)
    
    // 4. Create WebSocket hub and start LISTEN goroutine
    hub := ws.NewHub(database)
    go hub.Run()
    go hub.ListenForNotifications(ctx, database.Pool)
    
    // 5. Create Gin router with middleware
    router := gin.New()
    router.Use(gin.Recovery(), middleware.Logger(), cors.New(corsConfig))
    
    // 6. Register routes
    registerRoutes(router, handlers...)
    
    // 7. Start HTTP server
    srv.ListenAndServe()
}
```

**Server Configuration:**
- Port: `8082` (configurable via `PORT` env var)
- Mode: Development (logs all requests) or Production (minimal logs)
- Graceful Shutdown: Handles SIGINT/SIGTERM signals

---

### **2. Configuration (config/config.go)**

**Environment Variables:**
```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=helios_db
DB_MAX_CONNS=25

# Server
PORT=8082
SERVER_ENV=development

# Security
JWT_SECRET=your_secret_key_here
JWT_EXPIRY_HOURS=168  # 7 days

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5500

# Initial User Balance
INITIAL_BALANCE=10000.00
```

**Config Structure:**
```go
type Config struct {
    Database DatabaseConfig
    Server   ServerConfig
    JWT      JWTConfig
    CORS     CORSConfig
    Initial  InitialConfig
}
```

**Validation:**
- Ensures all required fields are set
- Returns error if critical config missing (DB password, JWT secret)

---

### **3. Database Layer (db/database.go)**

**Connection Pooling:**
```go
type Database struct {
    Pool *pgxpool.Pool  // Connection pool for concurrent requests
}

func NewDatabase(cfg *config.Config) (*Database, error) {
    poolConfig, _ := pgxpool.ParseConfig(connectionString)
    poolConfig.MaxConns = cfg.Database.MaxConns  // Typically 25
    
    pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
    // ...
}
```

**Key Methods:**

#### **User Management:**
```go
// Create new user with initial balance
CreateUser(ctx, username, email, passwordHash, quoteAssetID, balance) (int64, error)

// Get user by email for login
GetUserByEmail(ctx, email) (*User, error)
```

#### **Order Operations:**
```go
// Place order (calls sp_place_order stored procedure)
PlaceOrder(ctx, userID, tradingPairID, side, orderType, quantity, price) (map[string]interface{}, error)

// Cancel order (calls sp_cancel_order stored procedure)
CancelOrder(ctx, orderID, userID) (map[string]interface{}, error)

// Get user's order history
GetOrderHistory(ctx, userID, tradingPairSymbol, limit) ([]Order, error)
```

#### **Market Data:**
```go
// Get order book for a trading pair
GetOrderBook(ctx, tradingPairID) (OrderBook, error)

// Get recent trades
GetTradeHistory(ctx, tradingPairID, limit) ([]Trade, error)

// Get all trading pairs
GetAllTradingPairs(ctx) ([]TradingPair, error)
```

#### **Account Data:**
```go
// Get user balances for all assets
GetBalances(ctx, userID) ([]Balance, error)
```

**Stored Procedure Calls:**
```go
// Example: Place Order
var result struct {
    Status  string
    Message string
    OrderID *int64
}

err := db.Pool.QueryRow(ctx, 
    "SELECT * FROM sp_place_order($1, $2, $3, $4, $5, $6)",
    userID, tradingPairID, side, orderType, quantity, price,
).Scan(&result.Status, &result.Message, &result.OrderID)

if result.Status == "ERROR" {
    return nil, errors.New(result.Message)
}
```

---

### **4. Handlers**

#### **A. AuthHandler (handlers/auth.go)**

**POST /api/v1/auth/register**
```go
func (h *AuthHandler) Register(c *gin.Context) {
    // 1. Bind and validate request JSON
    var req RegisterRequest
    c.ShouldBindJSON(&req)  // Validates: username (3-50 chars), email, password (6+ chars)
    
    // 2. Hash password with bcrypt
    passwordHash := utils.HashPassword(req.Password)
    
    // 3. Create user in database
    userID, err := h.db.CreateUser(ctx, req.Username, req.Email, passwordHash, 
                                     h.initialQuoteAssetID, h.initialBalance)
    
    // 4. Generate JWT token
    token := utils.GenerateJWT(userID, req.Username)
    
    // 5. Return token and user info
    c.JSON(http.StatusCreated, LoginResponse{Token: token, User: ...})
}
```

**Request:**
```json
{
  "username": "trader1",
  "email": "trader1@example.com",
  "password": "securepass123"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "trader1",
    "email": "trader1@example.com"
  }
}
```

---

**POST /api/v1/auth/login**
```go
func (h *AuthHandler) Login(c *gin.Context) {
    // 1. Bind request
    var req LoginRequest
    c.ShouldBindJSON(&req)
    
    // 2. Get user from database
    user := h.db.GetUserByEmail(ctx, req.Email)
    
    // 3. Verify password
    if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
        return Unauthorized
    }
    
    // 4. Generate JWT
    token := utils.GenerateJWT(user.ID, user.Username)
    
    // 5. Return token
    c.JSON(http.StatusOK, LoginResponse{Token: token, User: user})
}
```

---

#### **B. OrderHandler (handlers/orders.go)**

**POST /api/v1/orders** (Protected - Requires JWT)
```go
func (h *OrderHandler) PlaceOrder(c *gin.Context) {
    // 1. Get user ID from JWT (set by middleware)
    userID := c.GetInt64("user_id")
    
    // 2. Bind and validate request
    var req PlaceOrderRequest
    c.ShouldBindJSON(&req)  // Validates: pair, side, type, quantity > 0
    
    // 3. Resolve trading pair symbol to ID
    tradingPairID := h.db.GetTradingPairIDBySymbol(ctx, req.Pair)
    
    // 4. Call stored procedure
    result := h.db.PlaceOrder(ctx, userID, tradingPairID, 
                              req.Side, req.Type, req.Quantity, req.Price)
    
    // 5. Return result
    if result["status"] == "SUCCESS" {
        c.JSON(http.StatusOK, result)
    } else {
        c.JSON(http.StatusBadRequest, ErrorResponse{Message: result["message"]})
    }
}
```

**Request:**
```json
{
  "pair": "BTC/USD",
  "side": "BUY",
  "type": "LIMIT",
  "quantity": 0.5,
  "price": 45000.00
}
```

**Response:**
```json
{
  "status": "SUCCESS",
  "message": "Order placed successfully",
  "order_id": 12345,
  "filled_quantity": 0.5,
  "trades": [
    {
      "trade_id": 789,
      "price": 44999.00,
      "quantity": 0.5
    }
  ]
}
```

---

**DELETE /api/v1/orders/:id** (Protected)
```go
func (h *OrderHandler) CancelOrder(c *gin.Context) {
    userID := c.GetInt64("user_id")
    orderID := c.Param("id")  // From URL path
    
    result := h.db.CancelOrder(ctx, orderID, userID)
    
    c.JSON(http.StatusOK, result)
}
```

---

**GET /api/v1/orders/history** (Protected)
```go
func (h *OrderHandler) GetOrderHistory(c *gin.Context) {
    userID := c.GetInt64("user_id")
    pair := c.Query("pair")       // Optional: filter by pair
    limit := c.Query("limit")     // Optional: limit results
    
    orders := h.db.GetOrderHistory(ctx, userID, pair, limit)
    
    c.JSON(http.StatusOK, DataResponse{Data: orders})
}
```

**Response:**
```json
{
  "data": [
    {
      "id": 12345,
      "trading_pair_symbol": "BTC/USD",
      "side": "BUY",
      "order_type": "LIMIT",
      "quantity": 0.5,
      "price": 45000.00,
      "filled_quantity": 0.5,
      "status": "FILLED",
      "created_at": "2025-11-13T00:15:30Z"
    }
  ]
}
```

---

#### **C. MarketHandler (handlers/market.go)**

**GET /api/v1/market/orderbook** (Public)
```go
func (h *MarketHandler) GetOrderBook(c *gin.Context) {
    pair := c.Query("pair")  // e.g., "BTC/USD"
    
    tradingPairID := h.db.GetTradingPairIDBySymbol(ctx, pair)
    orderBook := h.db.GetOrderBook(ctx, tradingPairID)
    
    c.JSON(http.StatusOK, DataResponse{Data: orderBook})
}
```

**Response:**
```json
{
  "data": {
    "bids": [
      {"price": 44995.00, "total_quantity": 1.5, "side": "BUY"},
      {"price": 44990.00, "total_quantity": 2.3, "side": "BUY"}
    ],
    "asks": [
      {"price": 45005.00, "total_quantity": 0.8, "side": "SELL"},
      {"price": 45010.00, "total_quantity": 1.2, "side": "SELL"}
    ]
  }
}
```

---

**GET /api/v1/market/trades** (Public)
```go
func (h *MarketHandler) GetTradeHistory(c *gin.Context) {
    pair := c.Query("pair")
    limit := c.Query("limit")  // Default: 50
    
    trades := h.db.GetTradeHistory(ctx, tradingPairID, limit)
    
    c.JSON(http.StatusOK, DataResponse{Data: trades})
}
```

---

**GET /api/v1/trading-pairs** (Public)
```go
func (h *MarketHandler) GetAllTradingPairs(c *gin.Context) {
    pairs := h.db.GetAllTradingPairs(ctx)
    c.JSON(http.StatusOK, DataResponse{Data: pairs})
}
```

**Response:**
```json
{
  "data": [
    {"id": 1, "symbol": "BTC/USD", "base_asset_id": 2, "quote_asset_id": 1},
    {"id": 2, "symbol": "ETH/USD", "base_asset_id": 3, "quote_asset_id": 1}
  ]
}
```

---

#### **D. AccountHandler (handlers/account.go)**

**GET /api/v1/account/balances** (Protected)
```go
func (h *AccountHandler) GetBalances(c *gin.Context) {
    userID := c.GetInt64("user_id")
    
    balances := h.db.GetBalances(ctx, userID)
    
    c.JSON(http.StatusOK, DataResponse{Data: balances})
}
```

**Response:**
```json
{
  "data": [
    {
      "asset_symbol": "USD",
      "available_balance": 5000.00,
      "locked_balance": 1000.00
    },
    {
      "asset_symbol": "BTC",
      "available_balance": 0.5,
      "locked_balance": 0.2
    }
  ]
}
```

---

### **5. Middleware**

#### **A. Auth Middleware (middleware/auth.go)**

**Purpose:** Validates JWT tokens and extracts user information.

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Extract token from Authorization header
        authHeader := c.GetHeader("Authorization")
        if !strings.HasPrefix(authHeader, "Bearer ") {
            c.AbortWithStatusJSON(401, ErrorResponse{Message: "Missing token"})
            return
        }
        
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        
        // 2. Parse and validate JWT
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return []byte(jwtSecret), nil
        })
        
        if err != nil || !token.Valid {
            c.AbortWithStatusJSON(401, ErrorResponse{Message: "Invalid token"})
            return
        }
        
        // 3. Extract claims
        claims := token.Claims.(jwt.MapClaims)
        userID := int64(claims["user_id"].(float64))
        username := claims["username"].(string)
        
        // 4. Set in context for handlers
        c.Set("user_id", userID)
        c.Set("username", username)
        
        c.Next()  // Continue to handler
    }
}
```

**Usage:**
```go
// Protected routes require authentication
orders := v1.Group("/orders")
orders.Use(middleware.AuthMiddleware())  // Apply to all routes in group
{
    orders.POST("", orderHandler.PlaceOrder)
    orders.DELETE("/:id", orderHandler.CancelOrder)
}
```

---

#### **B. Logger Middleware (middleware/logger.go)**

**Purpose:** Logs all HTTP requests with method, path, status, and latency.

```go
func Logger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        method := c.Request.Method
        
        c.Next()  // Process request
        
        latency := time.Since(start)
        status := c.Writer.Status()
        clientIP := c.ClientIP()
        
        log.Printf("[%s] %s %s | Status: %d | Latency: %s | IP: %s",
            method, path, c.Request.Proto, status, latency, clientIP)
    }
}
```

**Output:**
```
[GET] /api/v1/market/orderbook HTTP/1.1 | Status: 200 | Latency: 15ms | IP: 192.168.1.100
[POST] /api/v1/orders HTTP/1.1 | Status: 200 | Latency: 45ms | IP: 192.168.1.100
```

---

### **6. WebSocket System (ws/)**

#### **Hub (ws/hub.go)**

**Purpose:** Manages WebSocket connections and broadcasts market updates.

```go
type Hub struct {
    // Map of trading pair -> map of clients
    clients map[string]map[*Client]bool
    
    // Channels for client management
    register   chan *Client
    unregister chan *Client
    broadcast  chan Message
    
    db *db.Database
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            // Add client to trading pair group
            h.clients[client.tradingPair][client] = true
            
        case client := <-h.unregister:
            // Remove client and close connection
            delete(h.clients[client.tradingPair], client)
            close(client.send)
            
        case message := <-h.broadcast:
            // Send to all clients watching this trading pair
            for client := range h.clients[message.TradingPair] {
                select {
                case client.send <- message.Data:
                default:
                    // Client buffer full, disconnect
                    close(client.send)
                    delete(h.clients[message.TradingPair], client)
                }
            }
        }
    }
}
```

**PostgreSQL LISTEN:**
```go
func (h *Hub) ListenForNotifications(ctx context.Context, pool *pgxpool.Pool) {
    conn, _ := pool.Acquire(ctx)
    defer conn.Release()
    
    // Subscribe to PostgreSQL notifications
    conn.Exec(ctx, "LISTEN market_update")
    
    for {
        notification, err := conn.Conn().WaitForNotification(ctx)
        
        // Parse notification payload
        var update MarketUpdate
        json.Unmarshal([]byte(notification.Payload), &update)
        
        // Fetch updated order book
        orderBook := h.db.GetOrderBook(ctx, update.PairID)
        
        // Broadcast to WebSocket clients
        h.BroadcastOrderBook(ctx, update.PairID, tradingPairSymbol)
    }
}
```

---

#### **Handler (ws/handler.go)**

**GET /ws/v1/market/*pair**
```go
func (h *WSHandler) HandleWebSocket(c *gin.Context) {
    pair := c.Param("pair")
    pair = strings.TrimPrefix(pair, "/")  // Remove leading /
    
    // Upgrade HTTP connection to WebSocket
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Printf("Failed to upgrade: %v", err)
        return
    }
    
    // Create client and register
    client := &Client{
        hub:         h.hub,
        conn:        conn,
        send:        make(chan []byte, 256),
        tradingPair: pair,
    }
    
    h.hub.register <- client
    
    // Send initial order book
    tradingPairID := h.hub.db.GetTradingPairIDBySymbol(ctx, pair)
    h.hub.BroadcastOrderBook(ctx, tradingPairID, pair)
    
    // Start read/write pumps
    go client.writePump()
    go client.readPump()
}
```

**WebSocket Message Format:**
```json
{
  "type": "orderbook",
  "data": {
    "bids": [...],
    "asks": [...]
  }
}
```

---

## 🔐 Security Features

### **1. Password Hashing (utils/auth.go)**
```go
// Uses bcrypt with cost factor 14
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

### **2. JWT Tokens**
```go
func GenerateJWT(userID int64, username string) (string, error) {
    claims := jwt.MapClaims{
        "user_id":  userID,
        "username": username,
        "exp":      time.Now().Add(7 * 24 * time.Hour).Unix(),  // 7 days
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(jwtSecret))
}
```

**Token Structure:**
```
Header:  {"alg": "HS256", "typ": "JWT"}
Payload: {"user_id": 1, "username": "trader1", "exp": 1731542400}
Signature: HMACSHA256(base64(header) + "." + base64(payload), secret)
```

### **3. CORS Configuration**
```go
corsConfig := cors.Config{
    AllowOrigins:     []string{"http://localhost:3000"},  // Whitelist
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    AllowCredentials: true,  // Allow cookies/auth headers
}
```

### **4. Input Validation**
```go
// Gin binding tags automatically validate
type PlaceOrderRequest struct {
    Pair     string  `json:"pair" binding:"required"`
    Side     string  `json:"side" binding:"required"`
    Quantity float64 `json:"quantity" binding:"required,gt=0"`  // Must be > 0
}
```

---

## 📊 Performance Optimizations

### **1. Database Connection Pooling**
- Pool Size: 25 concurrent connections
- Reuses connections across requests
- Prevents connection overhead

### **2. Context-Based Cancellation**
```go
func (h *Handler) PlaceOrder(c *gin.Context) {
    ctx := c.Request.Context()  // Cancelled if client disconnects
    result := h.db.PlaceOrder(ctx, ...)  // Query cancelled automatically
}
```

### **3. Goroutines for Concurrency**
- WebSocket hub runs in separate goroutine
- PostgreSQL LISTEN runs in separate goroutine
- Each WebSocket client has read/write goroutines

### **4. Efficient JSON Parsing**
- Uses `json` package for fast encoding/decoding
- Streaming JSON for large responses

---

## 🧪 Error Handling

### **Consistent Error Responses:**
```go
type ErrorResponse struct {
    Error   string `json:"error"`    // Error code
    Message string `json:"message"`  // Human-readable message
}
```

**Examples:**
```json
// Validation Error
{"error": "validation_error", "message": "quantity must be greater than 0"}

// Authentication Error
{"error": "authentication_failed", "message": "Invalid username or password"}

// Server Error
{"error": "server_error", "message": "Failed to connect to database"}
```

### **HTTP Status Codes:**
- `200 OK`: Successful request
- `201 Created`: Resource created (e.g., user registration)
- `400 Bad Request`: Invalid input
- `401 Unauthorized`: Missing/invalid JWT
- `403 Forbidden`: Route not matched
- `404 Not Found`: Resource not found
- `409 Conflict`: Duplicate username/email
- `500 Internal Server Error`: Server-side error

---

## 🚀 Deployment

### **Build:**
```bash
cd api
go build -o helios-api main.go
```

### **Run:**
```bash
export PORT=8082
export DB_PASSWORD=your_password
export JWT_SECRET=your_secret
./helios-api
```

### **Docker (Future):**
```dockerfile
FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go build -o helios-api main.go
CMD ["./helios-api"]
```

---

## 📚 Summary

The Helios API provides:

✅ **RESTful Design:** Standard HTTP methods and JSON  
✅ **Real-Time Updates:** WebSocket with PostgreSQL NOTIFY  
✅ **Secure Authentication:** JWT tokens, bcrypt password hashing  
✅ **High Performance:** Connection pooling, goroutines, efficient queries  
✅ **Type Safety:** Go's strong typing prevents runtime errors  
✅ **Scalability:** Stateless design, horizontal scaling ready  
✅ **Maintainability:** Clean architecture, separation of concerns  

**Key Technology Stack:**
- **Go 1.21+**: High performance, compiled language
- **Gin Framework**: Fast HTTP router with middleware support
- **pgx/v5**: PostgreSQL driver with connection pooling
- **Gorilla WebSocket**: RFC 6455 compliant WebSocket implementation
- **JWT-go**: JSON Web Token generation and validation
- **bcrypt**: Industry-standard password hashing
