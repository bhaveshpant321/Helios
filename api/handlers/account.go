package handlers

import (
	"helios-api/db"
	"helios-api/middleware"
	"helios-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	db *db.Database
}

func NewAccountHandler(database *db.Database) *AccountHandler {
	return &AccountHandler{db: database}
}

// GetBalances retrieves user's account balances
// GET /api/v1/account/balances
func (h *AccountHandler) GetBalances(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	balances, err := h.db.GetUserBalances(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "server_error",
			Message: "Failed to retrieve balances",
		})
		return
	}

	c.JSON(http.StatusOK, balances)
}
