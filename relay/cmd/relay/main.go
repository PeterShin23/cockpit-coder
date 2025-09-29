package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"relay/internal/relay"
)

func main() {
	flag.Parse()

	// Load environment variables
	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = "8081"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid PORT: %v", err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	sessionTTLStr := os.Getenv("SESSION_TTL_SECONDS")
	if sessionTTLStr == "" {
		sessionTTLStr = "86400"
	}
	sessionTTL, err := strconv.ParseInt(sessionTTLStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid SESSION_TTL_SECONDS: %v", err)
	}
	sessionTTLDur := time.Duration(sessionTTL) * time.Second

	idleTimeoutStr := os.Getenv("IDLE_TIMEOUT_SECONDS")
	if idleTimeoutStr == "" {
		idleTimeoutStr = "1800"
	}
	idleTimeout, err := strconv.ParseInt(idleTimeoutStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid IDLE_TIMEOUT_SECONDS: %v", err)
	}
	idleTimeoutDur := time.Duration(idleTimeout) * time.Second

	ringBufferStr := os.Getenv("RING_BUFFER_BYTES")
	if ringBufferStr == "" {
		ringBufferStr = "131072"
	}
	ringBufferBytes, err := strconv.ParseInt(ringBufferStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid RING_BUFFER_BYTES: %v", err)
	}

	rateLimitStr := os.Getenv("RATE_LIMIT_BPS")
	if rateLimitStr == "" {
		rateLimitStr = "65536"
	}
	rateLimitBPS, err := strconv.ParseInt(rateLimitStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid RATE_LIMIT_BPS: %v", err)
	}

	corsOrigins := os.Getenv("CORS_ORIGINS")
	redisURL := os.Getenv("REDIS_URL")
	relayMintStr := os.Getenv("RELAY_MINT")
	relayMint := relayMintStr == "true"
	adminToken := os.Getenv("ADMIN_TOKEN")

	// Parse CORS origins to slice
	var corsOriginsSlice []string
	if corsOrigins != "" {
		corsOriginsSlice = strings.Split(corsOrigins, ",")
		for i := range corsOriginsSlice {
			corsOriginsSlice[i] = strings.TrimSpace(corsOriginsSlice[i])
		}
	}

	// Log config (no secrets)
	log.Printf("Starting relay server on port %d", port)
	log.Printf("Session TTL: %v", sessionTTLDur)
	log.Printf("Idle timeout: %v", idleTimeoutDur)
	log.Printf("Ring buffer size: %d bytes", ringBufferBytes)
	log.Printf("Rate limit: %d bytes/sec", rateLimitBPS)
	if len(corsOriginsSlice) > 0 {
		log.Printf("CORS origins: %v", corsOriginsSlice)
	}
	if redisURL != "" {
		log.Printf("Using Redis: %s (hidden)", "redacted")
	} else {
		log.Println("No Redis configured")
	}
	if relayMint {
		log.Println("Session minting enabled for testing")
	} else {
		log.Println("Session minting disabled; use backend")
	}
	if adminToken != "" {
		log.Println("Metrics endpoint protected")
	}

	// Create hub
	hub := relay.NewHub(sessionTTLDur, idleTimeoutDur, int(ringBufferBytes), int(rateLimitBPS), redisURL)

	// Create server
	srv := relay.NewServer(hub, corsOriginsSlice, relayMint, adminToken, jwtSecret)

	// HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      srv,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Listening on :%d", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")

	// Shutdown context
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}

	// Close hub
	hub.CloseAll()

	log.Println("Server stopped")
}
