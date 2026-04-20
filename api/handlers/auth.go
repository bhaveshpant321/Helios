package handlers

import (
	"helios-api/db"
	"helios-api/models"
	"helios-api/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	db                 *db.Database
	initialQuoteAssetID int
	initialBalance      float64
}

func NewAuthHandler(database *db.Database, quoteAssetID int, initialBalance float64) *AuthHandler {
	return &AuthHandler{
		db:                 database,
		initialQuoteAssetID: quoteAssetID,
		initialBalance:      initialBalance,
	}
}

// Register handles user registration
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// Hash password
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "server_error",
			Message: "Failed to hash password",
		})
		return
	}

	// Create user in database
	userID, err := h.db.CreateUser(
		c.Request.Context(),
		req.Username,
		req.Email,
		passwordHash,
		h.initialQuoteAssetID,
		h.initialBalance,
	)

	if err != nil {
		// Check for duplicate username/email errors
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error:   "registration_failed",
			Message: "Username or email already exists",
		})
		return
	}

	// Option: Auto-login after registration (generate token)
	token, err := utils.GenerateJWT(userID, req.Username)
	if err != nil {
		// If token generation fails, we still created the user, 
		// so we just return success without the token or prompt login.
		c.JSON(http.StatusCreated, models.RegisterResponse{
			UserID:  userID,
			Message: "User created successfully (auto-login failed)",
		})
		return
	}

	// Return combined response (token + user info)
	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"user": models.User{
			ID:       userID,
			Username: req.Username,
			Email:    req.Email,
		},
		"message": "User created successfully",
	})
}

// Login handles user authentication
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// Get user from database
	user, err := h.db.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "authentication_failed",
			Message: "Invalid username or password",
		})
		return
	}

	// Verify password
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "authentication_failed",
			Message: "Invalid username or password",
		})
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "server_error",
			Message: "Failed to generate authentication token",
		})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		Token: token,
		User: models.User{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	})
}
