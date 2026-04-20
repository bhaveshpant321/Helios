package handlers

import (
	"helios-api/db"
	"helios-api/middleware"
	"helios-api/models"
	"net/http"
	"strconv"
	"strings"
	"log"
	"fmt"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	db *db.Database
}

func NewOrderHandler(database *db.Database) *OrderHandler {
	return &OrderHandler{db: database}
}

// PlaceOrder creates a new order
// POST /api/v1/orders
func (h *OrderHandler) PlaceOrder(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	var req models.PlaceOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("DEBUG PlaceOrder: Validation error: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}
	priceStr := "nil"
	if req.Price != nil {
		priceStr = fmt.Sprintf("%f", *req.Price)
	}
	log.Printf("DEBUG PlaceOrder: pair='%s', side='%s', type='%s', quantity=%f, price=%s",
		req.Pair, req.Side, req.Type, req.Quantity, priceStr)

	// Validate order type
	side := strings.ToUpper(req.Side)
	orderType := strings.ToUpper(req.Type)

	if err := db.ValidateOrderType(side, orderType); err != nil {
		log.Printf("DEBUG PlaceOrder: ValidateOrderType failed: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// LIMIT orders must have a price
	if orderType == "LIMIT" && req.Price == nil {
		log.Printf("DEBUG PlaceOrder: LIMIT order missing price")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Price is required for LIMIT orders",
		})
		return
	}

	// MARKET orders should not have a price
	if orderType == "MARKET" && req.Price != nil {
		req.Price = nil
	}

	// Get trading pair ID from symbol
	tradingPairID, err := h.db.GetTradingPairIDBySymbol(c.Request.Context(), req.Pair)
	if err != nil {
		log.Printf("DEBUG PlaceOrder: GetTradingPairIDBySymbol failed: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_pair",
			Message: "Invalid trading pair: " + req.Pair,
		})
		return
	}

	// Place the order
	log.Printf("DEBUG PlaceOrder: Calling db.PlaceOrder with tradingPairID=%d, side=%s, type=%s, quantity=%f, price=%v", 
		tradingPairID, side, orderType, req.Quantity, req.Price)


	result, err := h.db.PlaceOrder(
		c.Request.Context(),
		userID,
		tradingPairID,
		side,
		orderType,
		req.Quantity,
		req.Price,
	)

	if err != nil {
		log.Printf("DEBUG PlaceOrder: db.PlaceOrder returned error: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "order_failed",
			Message: err.Error(),
		})
		return
	}

	log.Printf("DEBUG PlaceOrder: db.PlaceOrder result: %+v", result)

	// Check if the stored procedure returned an error
	if status, ok := result["status"].(string); ok && status == "ERROR" {
		message := "Unknown error"
		if msg, ok := result["message"].(string); ok {
			message = msg
		}
		log.Printf("DEBUG PlaceOrder: Stored procedure returned ERROR: %s", message)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "order_failed",
			Message: message,
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetOrderHistory retrieves user's order history
// GET /api/v1/orders/history?pair=BTC/USD (pair is optional)
func (h *OrderHandler) GetOrderHistory(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	pair := c.Query("pair")

	// If pair is provided, filter by trading pair
	if pair != "" {
		// Get trading pair ID
		tradingPairID, err := h.db.GetTradingPairIDBySymbol(c.Request.Context(), pair)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_pair",
				Message: "Invalid trading pair: " + pair,
			})
			return
		}

		orders, err := h.db.GetUserOrderHistory(c.Request.Context(), userID, tradingPairID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "server_error",
				Message: "Failed to retrieve order history",
			})
			return
		}

		c.JSON(http.StatusOK, orders)
		return
	}

	// Get all orders for user (across all pairs)
	orders, err := h.db.GetAllUserOrders(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "server_error",
			Message: "Failed to retrieve order history",
		})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// CancelOrder cancels an existing order
// DELETE /api/v1/orders/:id
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid order ID",
		})
		return
	}

	err = h.db.CancelOrder(c.Request.Context(), orderID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "cancel_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order " + orderIDStr + " cancelled successfully",
	})
}