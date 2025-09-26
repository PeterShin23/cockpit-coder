package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/PeterShin23/cockpit-coder/backend/internal/httpserver"
	"github.com/PeterShin23/cockpit-coder/backend/internal/session"
	"github.com/PeterShin23/cockpit-coder/backend/internal/pty"
)

func main() {
	// Initialize session manager
	sessionManager := session.NewMemoryManager()

	// Initialize PTY manager
	ptyManager := pty.NewManager()

	// Setup HTTP server
	server := httpserver.NewServer(sessionManager, ptyManager)

	// Start WebSocket hub for PTY connections
	ptyHub := pty.NewHub(sessionManager, ptyManager)
	go ptyHub.Run()

	// Start WebSocket hub for events
	eventsHub := pty.NewEventsHub(sessionManager)
	go eventsHub.Run()

	// Setup graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Start HTTP server
	port := getEnv("PORT", "8080")
	log.Printf("Starting server on port %s", port)
	
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-stop
	log.Println("Shutting down server...")

	// Cleanup
	ptyHub.Shutdown()
	eventsHub.Shutdown()
	
	log.Println("Server stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
