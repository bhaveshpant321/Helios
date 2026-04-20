package main

import (
	"context"
	"helios-api/config"
	"helios-api/db"
	"helios-api/handlers"
	"helios-api/middleware"
	"helios-api/utils"
	"helios-api/ws"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	log.Printf("🚀 Starting Helios API Server in %s mode...", cfg.Server.Env)

	// Initialize database
	database, err := db.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Initialize JWT utilities
	utils.InitJWTConfig(cfg)
	middleware.InitJWT(cfg)

	// Initialize WebSocket hub
	hub := ws.NewHub(database)
	go hub.Run()

	// Start PostgreSQL LISTEN goroutine
	ctx := context.Background()
	go hub.ListenForNotifications(ctx, database.Pool)

	// Set Gin mode
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	
	// CORS configuration
	corsConfig := cors.Config{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(corsConfig))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"timestamp": time.Now().Unix(),
		})
	})

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(database, cfg.Initial.QuoteAssetID, cfg.Initial.InitialBalance)
	accountHandler := handlers.NewAccountHandler(database)
	orderHandler := handlers.NewOrderHandler(database)
	marketHandler := handlers.NewMarketHandler(database)
	wsHandler := ws.NewWSHandler(hub)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Authentication routes (no auth required)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Account routes (auth required)
		account := v1.Group("/account")
		account.Use(middleware.AuthMiddleware())
		{
			account.GET("/balances", accountHandler.GetBalances)
		}

		// Order routes (auth required)
		orders := v1.Group("/orders")
		orders.Use(middleware.AuthMiddleware())
		{
			orders.POST("", orderHandler.PlaceOrder)
			orders.GET("/history", orderHandler.GetOrderHistory)
			orders.DELETE("/:id", orderHandler.CancelOrder)
		}

		// Public market data routes (no auth required)
		market := v1.Group("/market")
		{
			market.GET("/orderbook", marketHandler.GetOrderBook)
			market.GET("/trades", marketHandler.GetTradeHistory)
		}

		// Trading pairs route (no auth required)
		v1.GET("/trading-pairs", marketHandler.GetAllTradingPairs)
	}

	// WebSocket routes
	wsRoutes := router.Group("/ws/v1")
	{
		// Use *pair to capture everything including slashes (e.g., /BTC/USD)
		wsRoutes.GET("/market/*pair", wsHandler.HandleWebSocket)
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("✅ Server started on http://localhost:%s", cfg.Server.Port)
		log.Println("📡 WebSocket endpoint: ws://localhost:" + cfg.Server.Port + "/ws/v1/market/:pair")
		log.Println("📚 API Documentation: http://localhost:" + cfg.Server.Port + "/health")
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Graceful shutdown with 5 second timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}
