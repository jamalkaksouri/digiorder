// cmd/main.go - PRODUCTION-READY VERSION
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/jamalkaksouri/DigiOrder/internal/server"
)

func main() {
	// Setup logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting DigiOrder v3.0...")

	// Validate environment
	if err := validateEnvironment(); err != nil {
		log.Fatal("Environment validation failed:", err)
	}

	// Database connection with retry
	database, err := connectWithRetry(5, 2*time.Second)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	log.Println("Database connection established")

	// Create and configure server
	srv := server.New(database)

	// Start server in goroutine
	go func() {
		port := getEnv("SERVER_PORT", "5582")
		addr := ":" + port

		log.Printf("Server starting on %s", addr)
		if err := srv.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	sig := <-quit
	log.Printf("Received signal: %v. Shutting down server...", sig)

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited gracefully")
}

func validateEnvironment() error {
	required := []string{
		"DB_HOST",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
		"JWT_SECRET",
	}

	for _, env := range required {
		if os.Getenv(env) == "" {
			return fmt.Errorf("required environment variable %s is not set", env)
		}
	}

	// Validate JWT secret length
	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters, got %d", len(jwtSecret))
	}

	return nil
}

func connectWithRetry(maxRetries int, retryDelay time.Duration) (*sql.DB, error) {
	var database *sql.DB
	var err error

	for i := 0; i < maxRetries; i++ {
		database, err = db.Connect()
		if err == nil {
			return database, nil
		}

		log.Printf("Database connection attempt %d/%d failed: %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			log.Printf("Retrying in %v...", retryDelay)
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
		}
	}

	return nil, fmt.Errorf("failed to connect after %d attempts: %w", maxRetries, err)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}