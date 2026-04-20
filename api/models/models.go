package models

import "time"

// ============================================
// Request Models
// ============================================

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type PlaceOrderRequest struct {
	Pair     string  `json:"pair" binding:"required"`      // e.g., "BTC/USD"
	Side     string  `json:"side" binding:"required"`      // "BUY" or "SELL"
	Type     string  `json:"type" binding:"required"`      // "MARKET" or "LIMIT"
	Quantity float64 `json:"quantity" binding:"required,gt=0"`
	Price    *float64 `json:"price"`                       // Required for LIMIT orders
}

// ============================================
// Response Models
// ============================================

type RegisterResponse struct {
	UserID  int64  `json:"userId"`
	Message string `json:"message"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// ============================================
// Domain Models
// ============================================

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never expose password hash
	CreatedAt    time.Time `json:"created_at"`
}

type Balance struct {
	AssetID      int     `json:"asset_id"`
	TickerSymbol string  `json:"ticker_symbol"`
	Name         string  `json:"name"`
	Balance      float64 `json:"balance,string"`
	HeldBalance  float64 `json:"held_balance,string"`
}

type Order struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	TradingPairID   int       `json:"trading_pair_id"`
	Side            string    `json:"side"`
	Type            string    `json:"type"`
	Status          string    `json:"status"`
	Price           *float64  `json:"price"`
	Quantity        float64   `json:"quantity"`
	FilledQuantity  float64   `json:"filled_quantity"`
	CreatedAt       time.Time `json:"created_at"`
}

type Trade struct {
	ID             int64     `json:"id"`
	MakerOrderID   int64     `json:"maker_order_id"`
	TakerOrderID   int64     `json:"taker_order_id"`
	TradingPairID  int       `json:"trading_pair_id"`
	Price          float64   `json:"price"`
	Quantity       float64   `json:"quantity"`
	ExecutedAt     time.Time `json:"executed_at"`
}

type OrderBookEntry struct {
	Side          string  `json:"side"`
	Price         float64 `json:"price"`
	TotalQuantity float64 `json:"total_quantity"`
}

type OrderBook struct {
	Bids []OrderBookEntry `json:"bids"`
	Asks []OrderBookEntry `json:"asks"`
}

type TradingPair struct {
	ID            int    `json:"id"`
	BaseAssetID   int    `json:"base_asset_id"`
	QuoteAssetID  int    `json:"quote_asset_id"`
	Symbol        string `json:"symbol"`
}

// ============================================
// WebSocket Models
// ============================================

type WSMessage struct {
	Type string      `json:"type"` // "orderbook", "trade", "error"
	Data interface{} `json:"data"`
}

type MarketUpdate struct {
	PairID int `json:"pair_id"`
}
