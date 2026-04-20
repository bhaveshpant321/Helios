# Helios Trading Platform - System Architecture

## High-Level Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         FRONTEND (Browser)                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
│  │ Login    │  │  Trade   │  │ History  │  │ Profile  │       │
│  │ (HTML/JS)│  │(HTML/JS) │  │(HTML/JS) │  │(HTML/JS) │       │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘       │
│       │             │               │             │             │
└───────┼─────────────┼───────────────┼─────────────┼─────────────┘
        │             │               │             │
        │ HTTP/REST   │ WebSocket     │ HTTP/REST   │ HTTP/REST
        │ (JSON)      │ (Real-time)   │ (JSON)      │ (JSON)
        ▼             ▼               ▼             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    GO API SERVER (Port 8080)                    │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                     Router (Gin)                          │   │
│  │  ┌────────────┐ ┌──────────┐ ┌──────────┐               │   │
│  │  │   CORS     │ │  Logger  │ │   Auth   │  Middleware   │   │
│  │  │ Middleware │ │Middleware│ │Middleware│               │   │
│  │  └────────────┘ └──────────┘ └──────────┘               │   │
│  └───────────────────────────┬──────────────────────────────┘   │
│                              │                                   │
│  ┌───────────────────────────┴──────────────────────────────┐   │
│  │                   HTTP HANDLERS                           │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌─────────┐  │   │
│  │  │   Auth   │  │ Account  │  │  Orders  │  │ Market  │  │   │
│  │  │ Handler  │  │ Handler  │  │ Handler  │  │ Handler │  │   │
│  │  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬────┘  │   │
│  └───────┼─────────────┼─────────────┼─────────────┼────────┘   │
│          │             │             │             │             │
│          ├─────────────┴─────────────┴─────────────┘             │
│          │                                                        │
│          ▼                                                        │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │              DATABASE LAYER (db/database.go)             │   │
│  │  • Connection Pool (pgxpool)                             │   │
│  │  • Stored Procedure Calls                                │   │
│  │  • Query Builders                                        │   │
│  └──────────────────┬───────────────────────────────────────┘   │
│                     │                                            │
│  ┌──────────────────┴───────────────────────────────────────┐   │
│  │         WEBSOCKET HUB (ws/hub.go)                        │   │
│  │  ┌────────────┐  ┌──────────────┐  ┌─────────────────┐  │   │
│  │  │ Client     │  │ Broadcast    │  │ PostgreSQL      │  │   │
│  │  │ Manager    │  │ Channel      │  │ LISTEN Goroutine│  │   │
│  │  └────────────┘  └──────────────┘  └─────────────────┘  │   │
│  └──────────────────┬───────────────────────────────────────┘   │
│                     │                                            │
└─────────────────────┼────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                      POSTGRESQL DATABASE                        │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                       TABLES                              │   │
│  │  • users          • assets         • trading_pairs       │   │
│  │  • accounts       • orders         • trades              │   │
│  │  • fees                                                   │   │
│  └──────────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                  STORED PROCEDURES                        │   │
│  │  • sp_create_user            • sp_place_order            │   │
│  │  • sp_get_user_by_username   • sp_cancel_order           │   │
│  │  • sp_get_user_balances      • sp_get_order_book         │   │
│  │  • sp_get_user_order_history • sp_get_trade_history      │   │
│  └──────────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                  LISTEN/NOTIFY                            │   │
│  │  Channel: market_update                                   │   │
│  │  Triggered by: sp_place_order                             │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## Request Flow Examples

### 1. User Login Flow

```
┌────────┐        ┌────────┐         ┌──────────┐        ┌──────────┐
│Browser │        │  API   │         │ Database │        │  JWT     │
│        │        │ Server │         │  Layer   │        │ Utils    │
└───┬────┘        └───┬────┘         └────┬─────┘        └────┬─────┘
    │                 │                   │                   │
    │ POST /login     │                   │                   │
    ├────────────────>│                   │                   │
    │ {username,pwd}  │                   │                   │
    │                 │ GetUserByUsername │                   │
    │                 ├──────────────────>│                   │
    │                 │                   │ sp_get_user_by_  │
    │                 │                   │    username()     │
    │                 │   User Record     │                   │
    │                 │<──────────────────┤                   │
    │                 │                   │                   │
    │                 │ Verify Password   │                   │
    │                 │ (bcrypt)          │                   │
    │                 │                   │                   │
    │                 │ Generate JWT      │                   │
    │                 ├───────────────────────────────────────>│
    │                 │                   │   JWT Token       │
    │                 │<───────────────────────────────────────┤
    │ JWT Token       │                   │                   │
    │<────────────────┤                   │                   │
    │                 │                   │                   │
```

### 2. Place Order Flow with WebSocket Broadcast

```
┌────────┐  ┌────────┐  ┌──────────┐  ┌──────────┐  ┌───────────┐  ┌──────┐
│Browser │  │  API   │  │ Database │  │PostgreSQL│  │ WebSocket │  │Other │
│        │  │ Server │  │  Layer   │  │ NOTIFY   │  │    Hub    │  │Clients│
└───┬────┘  └───┬────┘  └────┬─────┘  └────┬─────┘  └─────┬─────┘  └───┬──┘
    │           │             │             │              │            │
    │POST /orders│            │             │              │            │
    ├──────────>│             │             │              │            │
    │{pair,side}│             │             │              │            │
    │           │ PlaceOrder  │             │              │            │
    │           ├────────────>│             │              │            │
    │           │             │sp_place_order()            │            │
    │           │             ├────────────>│              │            │
    │           │             │  • Validate │              │            │
    │           │             │  • Match    │              │            │
    │           │             │  • Execute  │              │            │
    │           │             │  • Update DB│              │            │
    │           │             │             │              │            │
    │           │             │  pg_notify('market_update')│            │
    │           │             │             ├─────────────>│            │
    │           │             │             │              │            │
    │           │  Result     │             │              │            │
    │           │<────────────┤             │              │            │
    │  Success  │             │             │              │            │
    │<──────────┤             │             │              │            │
    │           │             │             │ Notification │            │
    │           │             │             │ Received     │            │
    │           │             │<────────────┤              │            │
    │           │             │             │              │            │
    │           │      GetOrderBook()       │              │            │
    │           │             ├─────────────┤              │            │
    │           │       Order Book          │              │            │
    │           │             │<────────────┤              │            │
    │           │             │             │              │            │
    │           │      Broadcast to Clients │              │            │
    │           │             ├─────────────────────────────>           │
    │           │             │             │              │  Update!   │
    │           │             │             │              ├───────────>│
    │           │             │             │              │            │
```

### 3. WebSocket Real-time Update Flow

```
┌────────┐      ┌───────────┐      ┌──────────┐      ┌──────────┐
│ Client │      │ WebSocket │      │PostgreSQL│      │ Database │
│Browser │      │    Hub    │      │  LISTEN  │      │  Layer   │
└───┬────┘      └─────┬─────┘      └────┬─────┘      └────┬─────┘
    │                 │                  │                 │
    │ WS Connect      │                  │                 │
    │ /market/BTC-USD │                  │                 │
    ├────────────────>│                  │                 │
    │                 │ Register Client  │                 │
    │                 │ (Add to Hub)     │                 │
    │                 │                  │                 │
    │ Initial Order   │                  │                 │
    │ Book            │ GetOrderBook()   │                 │
    │                 ├──────────────────────────────────>│
    │<────────────────┤ Current State    │                 │
    │                 │<────────────────────────────────────┤
    │                 │                  │                 │
    │                 │ ... waiting for updates ...        │
    │                 │                  │                 │
    │                 │                  │ NOTIFY received │
    │                 │<─────────────────┤                 │
    │                 │ {pair_id: 1}     │                 │
    │                 │                  │                 │
    │                 │ GetOrderBook()   │                 │
    │                 ├──────────────────────────────────>│
    │                 │ Fresh Order Book │                 │
    │                 │<────────────────────────────────────┤
    │                 │                  │                 │
    │ Updated Order   │                  │                 │
    │ Book (JSON)     │                  │                 │
    │<────────────────┤                  │                 │
    │                 │                  │                 │
```

## Component Responsibilities

### Frontend (HTML/JavaScript)
- **Responsibility**: User interface, form handling, API calls
- **Technology**: HTML5, CSS3, JavaScript (Vanilla)
- **Key Files**: `login.html`, `trade.html`, `history.html`, `profile.html`

### API Server (Go)
- **Responsibility**: HTTP routing, authentication, business logic orchestration
- **Technology**: Go 1.21+, Gin framework
- **Key Files**: `main.go`, `handlers/*.go`, `middleware/*.go`

### Database Layer (Go)
- **Responsibility**: PostgreSQL connection management, stored procedure calls
- **Technology**: pgx/v5, pgxpool
- **Key File**: `db/database.go`

### WebSocket Hub (Go)
- **Responsibility**: Real-time client management, message broadcasting
- **Technology**: Gorilla WebSocket
- **Key Files**: `ws/hub.go`, `ws/handler.go`

### PostgreSQL Database
- **Responsibility**: Data persistence, matching engine logic, NOTIFY events
- **Technology**: PostgreSQL 14+, PL/pgSQL
- **Key Files**: `../db/schema.sql`, `../db/procedures/*.sql`

## Data Flow Patterns

### 1. Authentication Pattern
```
Request → CORS → Auth Middleware → Handler → Database → Response
```

### 2. Public Endpoint Pattern
```
Request → CORS → Logger → Handler → Database → Response
```

### 3. WebSocket Pattern
```
Connection → Upgrade → Hub Registration → Listen Loop → Broadcast
```

### 4. Database Notification Pattern
```
LISTEN Goroutine → Wait for NOTIFY → Parse Payload → Fetch Data → Broadcast
```

## Security Layers

```
┌─────────────────────────────────────────────────────────┐
│  Layer 1: CORS (Origin Validation)                      │
├─────────────────────────────────────────────────────────┤
│  Layer 2: JWT Authentication (Token Validation)         │
├─────────────────────────────────────────────────────────┤
│  Layer 3: Database-level (User ID Verification)         │
├─────────────────────────────────────────────────────────┤
│  Layer 4: Stored Procedures (Permission Checks)         │
└─────────────────────────────────────────────────────────┘
```

## Scalability Considerations

### Current Design (Single Instance)
```
Frontend → API Server → PostgreSQL
              ↓
         WebSocket Hub
```

### Future Scaling (Multiple Instances)
```
                  ┌─── API Server 1 ───┐
Frontend ── LB ───┤                     ├─── PostgreSQL
                  ├─── API Server 2 ───┤
                  └─── API Server 3 ───┘
                         ↓
                    Redis PubSub
                  (for WebSocket sync)
```

## Key Design Decisions

1. **Go Language**: Performance, concurrency, simplicity
2. **Gin Framework**: Fast routing, middleware support
3. **JWT Authentication**: Stateless, scalable
4. **pgx Driver**: Best performance for PostgreSQL
5. **WebSocket Hub**: In-memory, single-instance optimization
6. **LISTEN/NOTIFY**: Database-driven events, no polling
7. **Stored Procedures**: Business logic in database

---

This architecture provides:
- ✅ High performance (Go concurrency)
- ✅ Real-time updates (WebSocket + NOTIFY)
- ✅ Security (JWT + middleware)
- ✅ Maintainability (clean separation)
- ✅ Scalability (connection pooling, stateless design)
