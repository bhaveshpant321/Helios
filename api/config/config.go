package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	CORS     CORSConfig
	Initial  InitialConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	MaxConns int32
	MinConns int32
}

type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

type CORSConfig struct {
	AllowedOrigins []string
}

type InitialConfig struct {
	QuoteAssetID   int
	InitialBalance float64
}

func Load() (*Config, error) {
	// Load .env file if it exists (development)
	_ = godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Env:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "helios"),
			MaxConns: int32(getEnvAsInt("DB_MAX_CONNS", 25)),
			MinConns: int32(getEnvAsInt("DB_MIN_CONNS", 5)),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", ""),
			Expiry: time.Duration(getEnvAsInt("JWT_EXPIRY_HOURS", 24)) * time.Hour,
		},
		CORS: CORSConfig{
			AllowedOrigins: strings.Split(
				getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
				",",
			),
		},
		Initial: InitialConfig{
			QuoteAssetID:   getEnvAsInt("INITIAL_QUOTE_ASSET_ID", 1),
			InitialBalance: getEnvAsFloat("INITIAL_BALANCE", 10000.00),
		},
	}

	// Validate required fields
	if config.Database.Password == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	}

	if config.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return config, nil
}

func (c *Config) GetDatabaseURL() string {
	// Priority 1: Use DATABASE_URL if provided (Standard for Render/Railway/Neon)
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		return dbURL
	}

	// Priority 2: Build from individual components
	sslMode := getEnv("DB_SSLMODE", "disable")
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
		sslMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultValue
}
