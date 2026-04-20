package db

import (
	"context"
	"encoding/json"
	"fmt"
	"helios-api/config"
	"helios-api/models"
	"log"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Pool *pgxpool.Pool
}

// NewDatabase creates a new database connection pool
func NewDatabase(cfg *config.Config) (*Database, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.GetDatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("unable to parse database URL: %w", err)
	}

	poolConfig.MaxConns = cfg.Database.MaxConns
	poolConfig.MinConns = cfg.Database.MinConns

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	log.Println("✅ Database connection established")

	return &Database{Pool: pool}, nil
}

// Close closes the database connection pool
func (db *Database) Close() {
	db.Pool.Close()
}

// ============================================
// Authentication Procedures
// ============================================

// CreateUser calls sp_create_user stored procedure
func (db *Database) CreateUser(ctx context.Context, username, email, passwordHash string, initialQuoteAssetID int, initialBalance float64) (int64, error) {
	var userID int64
	err := db.Pool.QueryRow(
		ctx,
		"SELECT sp_create_user($1, $2, $3, $4, $5)",
		username,
		email,
		passwordHash,
		initialQuoteAssetID,
		initialBalance,
	).Scan(&userID)

	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	return userID, nil
}

func (db *Database) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := db.Pool.QueryRow(
		ctx,
		"SELECT * FROM sp_get_user_by_email($1)",
		email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// ============================================
// Account Procedures
// ============================================

// GetUserBalances calls sp_get_user_balances stored procedure
func (db *Database) GetUserBalances(ctx context.Context, userID int64) ([]models.Balance, error) {
	rows, err := db.Pool.Query(
		ctx,
		"SELECT * FROM sp_get_user_balances($1)",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get balances: %w", err)
	}
	defer rows.Close()

	var balances []models.Balance
	for rows.Next() {
		var balance models.Balance
		if err := rows.Scan(
			&balance.AssetID,
			&balance.TickerSymbol,
			&balance.Name,
			&balance.Balance,
			&balance.HeldBalance,
		); err != nil {
			return nil, fmt.Errorf("failed to scan balance: %w", err)
		}
		balances = append(balances, balance)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating balances: %w", err)
	}

	return balances, nil
}

// ============================================
// Order Management Procedures
// ============================================

// PlaceOrder calls sp_place_order stored procedure
func (db *Database) PlaceOrder(ctx context.Context, userID int64, tradingPairID int, side, orderType string, quantity float64, price *float64) (map[string]interface{}, error) {
	var resultJSON []byte
	err := db.Pool.QueryRow(
		ctx,
		"SELECT sp_place_order($1, $2, $3, $4, $5, $6)",
		userID,
		tradingPairID,
		side,
		orderType,
		quantity,
		price,
	).Scan(&resultJSON)

	if err != nil {
		return nil, fmt.Errorf("failed to place order: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to parse order result: %w", err)
	}

	return result, nil
}

// CancelOrder calls sp_cancel_order stored procedure
func (db *Database) CancelOrder(ctx context.Context, orderID, userID int64) error {
	_, err := db.Pool.Exec(
		ctx,
		"CALL sp_cancel_order($1, $2)",
		orderID,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	return nil
}

// GetUserOrderHistory calls sp_get_user_order_history stored procedure
func (db *Database) GetUserOrderHistory(ctx context.Context, userID int64, tradingPairID int) ([]models.Order, error) {
	rows, err := db.Pool.Query(
		ctx,
		"SELECT * FROM sp_get_user_order_history($1, $2)",
		userID,
		tradingPairID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get order history: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.TradingPairID,
			&order.Side,
			&order.Type,
			&order.Status,
			&order.Price,
			&order.Quantity,
			&order.FilledQuantity,
			&order.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// GetAllUserOrders retrieves all orders for a user across all trading pairs
func (db *Database) GetAllUserOrders(ctx context.Context, userID int64) ([]models.Order, error) {
	query := `
		SELECT 
			id, user_id, trading_pair_id, side, type, status, 
			price, quantity, filled_quantity, created_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all user orders: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.TradingPairID,
			&order.Side,
			&order.Type,
			&order.Status,
			&order.Price,
			&order.Quantity,
			&order.FilledQuantity,
			&order.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// ============================================
// Market Data Procedures
// ============================================

// GetOrderBook calls sp_get_order_book stored procedure
func (db *Database) GetOrderBook(ctx context.Context, tradingPairID int) (*models.OrderBook, error) {
	rows, err := db.Pool.Query(
		ctx,
		"SELECT * FROM sp_get_order_book($1)",
		tradingPairID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get order book: %w", err)
	}
	defer rows.Close()

	orderBook := &models.OrderBook{
		Bids: []models.OrderBookEntry{},
		Asks: []models.OrderBookEntry{},
	}

	for rows.Next() {
		var entry models.OrderBookEntry
		if err := rows.Scan(&entry.Side, &entry.Price, &entry.TotalQuantity); err != nil {
			return nil, fmt.Errorf("failed to scan order book entry: %w", err)
		}

		if entry.Side == "BUY" {
			orderBook.Bids = append(orderBook.Bids, entry)
		} else {
			orderBook.Asks = append(orderBook.Asks, entry)
		}
	}

	return orderBook, nil
}

// GetTradeHistory calls sp_get_trade_history stored procedure
func (db *Database) GetTradeHistory(ctx context.Context, tradingPairID int, limit int) ([]models.Trade, error) {
	rows, err := db.Pool.Query(
		ctx,
		"SELECT * FROM sp_get_trade_history($1, $2)",
		tradingPairID,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get trade history: %w", err)
	}
	defer rows.Close()

	var trades []models.Trade
	for rows.Next() {
		var trade models.Trade
		if err := rows.Scan(
			&trade.ID,
			&trade.MakerOrderID,
			&trade.TakerOrderID,
			&trade.TradingPairID,
			&trade.Price,
			&trade.Quantity,
			&trade.ExecutedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan trade: %w", err)
		}
		trades = append(trades, trade)
	}

	return trades, nil
}

// ============================================
// Helper Functions
// ============================================

// GetTradingPairBySymbol gets a trading pair by its symbol (e.g., "BTC/USD")
func (db *Database) GetTradingPairBySymbol(ctx context.Context, symbol string) (*models.TradingPair, error) {
	var pair models.TradingPair
	err := db.Pool.QueryRow(
		ctx,
		"SELECT id, base_asset_id, quote_asset_id, symbol FROM trading_pairs WHERE symbol = $1",
		symbol,
	).Scan(&pair.ID, &pair.BaseAssetID, &pair.QuoteAssetID, &pair.Symbol)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("trading pair not found: %s", symbol)
		}
		return nil, fmt.Errorf("failed to get trading pair: %w", err)
	}

	return &pair, nil
}

// GetTradingPairIDBySymbol is a convenience method
func (db *Database) GetTradingPairIDBySymbol(ctx context.Context, symbol string) (int, error) {
	pair, err := db.GetTradingPairBySymbol(ctx, symbol)
	if err != nil {
		return 0, err
	}
	return pair.ID, nil
}

// GetTradingPairByID gets a trading pair by its ID
func (db *Database) GetTradingPairByID(ctx context.Context, id int) (*models.TradingPair, error) {
	var pair models.TradingPair
	err := db.Pool.QueryRow(
		ctx,
		"SELECT id, base_asset_id, quote_asset_id, symbol FROM trading_pairs WHERE id = $1",
		id,
	).Scan(&pair.ID, &pair.BaseAssetID, &pair.QuoteAssetID, &pair.Symbol)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("trading pair not found with ID: %d", id)
		}
		return nil, fmt.Errorf("failed to get trading pair: %w", err)
	}

	return &pair, nil
}

// ValidateOrderType validates order side and type
func ValidateOrderType(side, orderType string) error {
	side = strings.ToUpper(side)
	orderType = strings.ToUpper(orderType)

	if side != "BUY" && side != "SELL" {
		return fmt.Errorf("invalid side: must be BUY or SELL")
	}

	if orderType != "MARKET" && orderType != "LIMIT" {
		return fmt.Errorf("invalid type: must be MARKET or LIMIT")
	}

	return nil
}

// GetAllTradingPairs retrieves all active trading pairs with asset details
func (db *Database) GetAllTradingPairs(ctx context.Context) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			tp.id,
			tp.base_asset_id,
			tp.quote_asset_id,
			tp.symbol,
			ba.name as base_name,
			ba.ticker_symbol as base_symbol,
			qa.name as quote_name,
			qa.ticker_symbol as quote_symbol
		FROM trading_pairs tp
		JOIN assets ba ON tp.base_asset_id = ba.id
		JOIN assets qa ON tp.quote_asset_id = qa.id
		ORDER BY tp.symbol
	`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query trading pairs: %w", err)
	}
	defer rows.Close()

	var pairs []map[string]interface{}
	for rows.Next() {
		var (
			id, baseID, quoteID int
			symbol, baseName, baseSymbol, quoteName, quoteSymbol string
		)

		err := rows.Scan(&id, &baseID, &quoteID, &symbol, &baseName, &baseSymbol, &quoteName, &quoteSymbol)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trading pair: %w", err)
		}

		pairs = append(pairs, map[string]interface{}{
			"id":              id,
			"base_asset_id":   baseID,
			"quote_asset_id":  quoteID,
			"symbol":          symbol,
			"base_name":       baseName,
			"base_symbol":     baseSymbol,
			"quote_name":      quoteName,
			"quote_symbol":    quoteSymbol,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating trading pairs: %w", err)
	}

	return pairs, nil
}
