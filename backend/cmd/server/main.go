package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/PeterShin23/cockpit-coder/backend/internal/auth"
	"github.com/PeterShin23/cockpit-coder/backend/internal/httpserver"
	"github.com/PeterShin23/cockpit-coder/backend/internal/pty"
	"github.com/PeterShin23/cockpit-coder/backend/internal/session"
)

func main() {
	// Parse environment variables
	port := getEnv("PORT", "8080")
	jwtSecret := getEnv("JWT_SECRET", "")
	repoAllowlist := strings.Split(getEnv("REPO_ALLOWLIST", ""), ",")
	cmdAllowlist := strings.Split(getEnv("CMD_ALLOWLIST", ""), ",")
	corsOrigins := strings.Split(getEnv("CORS_ORIGINS", "http://localhost:19006"), ",")

	// Handle JWT secret
	if jwtSecret == "" {
		jwtSecret = "dev_secret_change_me"
		log.Println("WARNING: Using development JWT secret. Set JWT_SECRET in production!")
	}
	auth.SetJWTSecret(jwtSecret)

	// Log configuration
	log.Printf("Starting server on port %s", port)
	log.Printf("CORS origins: %v", corsOrigins)
	log.Printf("Repo allowlist: %v", repoAllowlist)
	log.Printf("Command allowlist: %v", cmdAllowlist)

	// Initialize core components
	sessionManager := session.NewMemoryManager()
	ptyManager := pty.NewManager()

	// Setup HTTP server
	server := httpserver.NewServer(sessionManager, ptyManager)

	// Setup graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Start HTTP server
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	log.Println("Server started successfully")

	// Wait for shutdown signal
	<-stop
	log.Println("Shutting down server...")

	log.Println("Server stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	value := getEnv(key, "")
	if value == "" {
		return defaultValue
	}
	
	var result int
	_, err := fmt.Sscanf(value, "%d", &result)
	if err != nil {
		return defaultValue
	}
	return result
}
