package handlers

import (
	"helios-api/db"
	"helios-api/models"
	"net/http"
	"strconv"
	"log"
	
	"github.com/gin-gonic/gin"
)

type MarketHandler struct {
	db *db.Database
}

func NewMarketHandler(database *db.Database) *MarketHandler {
	return &MarketHandler{db: database}
}

// GetOrderBook retrieves the order book for a trading pair
// GET /api/v1/market/trades?pair=BTC/USD&limit=100
func (h *MarketHandler) GetOrderBook(c *gin.Context) {
	pair := c.Query("pair")
	if pair == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Trading pair is required",
		})
		return
	}

	// Get trading pair ID
	tradingPairID, err := h.db.GetTradingPairIDBySymbol(c.Request.Context(), pair)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_pair",
			Message: "Invalid trading pair: " + pair,
		})
		return
	}

	orderBook, err := h.db.GetOrderBook(c.Request.Context(), tradingPairID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "server_error",
			Message: "Failed to retrieve order book",
		})
		return
	}

	c.JSON(http.StatusOK, orderBook)
}

// GetTradeHistory retrieves recent trades for a trading pair
// GET /api/v1/market/trades/:pair?limit=100
func (h *MarketHandler) GetTradeHistory(c *gin.Context) {
	pair := c.Query("pair")
	log.Printf("DEBUG GetTradeHistory: pair='%s', full URL='%s'", pair, c.Request.URL.String())
	if pair == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Trading pair is required (use ?pair=BTC/USD)",
		})
		return
	}

	// Parse limit parameter
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 100
	}

	// Get trading pair ID
	tradingPairID, err := h.db.GetTradingPairIDBySymbol(c.Request.Context(), pair)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_pair",
			Message: "Invalid trading pair: " + pair,
		})
		return
	}

	trades, err := h.db.GetTradeHistory(c.Request.Context(), tradingPairID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "server_error",
			Message: "Failed to retrieve trade history",
		})
		return
	}

	c.JSON(http.StatusOK, trades)
}

// GetAllTradingPairs retrieves all active trading pairs
// GET /api/v1/trading-pairs
func (h *MarketHandler) GetAllTradingPairs(c *gin.Context) {
	tradingPairs, err := h.db.GetAllTradingPairs(c.Request.Context())
	if err != nil {
        // -------------------------------------------------------------------
        // ADD THIS LINE to print the actual database error to your console
        // -------------------------------------------------------------------
        log.Printf("❌ DB Error in GetAllTradingPairs: %v", err) 
        // -------------------------------------------------------------------
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "server_error",
			Message: "Failed to retrieve trading pairs",
		})
		return
	}

	c.JSON(http.StatusOK, tradingPairs)
}
