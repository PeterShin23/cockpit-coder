package httpserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PeterShin23/cockpit-coder/backend/internal/auth"
	"github.com/PeterShin23/cockpit-coder/backend/internal/session"
	"github.com/PeterShin23/cockpit-coder/backend/internal/pty"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Server struct {
	router        *mux.Router
	sessionManager *session.MemoryManager
	ptyManager    pty.Manager
}

func NewServer(sessionManager *session.MemoryManager, ptyManager pty.Manager) *Server {
	s := &Server{
		router:        mux.NewRouter(),
		sessionManager: sessionManager,
		ptyManager:    ptyManager,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Health check
	s.router.HandleFunc("/healthz", s.healthCheck).Methods("GET")

	// API routes
	api := s.router.PathPrefix("/api").Subrouter()
	
	// Session routes
	api.HandleFunc("/session", s.createSession).Methods("POST")
	api.HandleFunc("/session/{id}", s.getSession).Methods("GET")
	
	// Task routes
	api.HandleFunc("/tasks", s.createTask).Methods("POST")
	api.HandleFunc("/tasks/{id}", s.getTask).Methods("GET")
	api.HandleFunc("/tasks/{id}/patches", s.getTaskPatches).Methods("GET")
	api.HandleFunc("/tasks/{id}/apply", s.applyTaskPatches).Methods("POST")
	
	// Command routes
	api.HandleFunc("/cmd", s.runCommand).Methods("POST")
	
	// Git routes
	api.HandleFunc("/git/diff", s.getGitDiff).Methods("GET")
	
	// WebSocket endpoints
	s.router.HandleFunc("/ws/pty", s.handlePtyWebSocket)
	s.router.HandleFunc("/ws/events", s.handleEventsWebSocket)
	
	// CORS middleware
	s.router.Use(s.corsMiddleware)
	s.router.Use(s.loggingMiddleware)
}

func (s *Server) ListenAndServe() error {
	port := ":" + getEnv("PORT", "8080")
	return http.ListenAndServe(port, s.router)
}

// DTOs
type SessionCreateRequest struct {
	Repo  string `json:"repo"`
	Label string `json:"label,omitempty"`
	Via   string `json:"via,omitempty"`
}

type SessionCreateResponse struct {
	SessionID string            `json:"sessionId"`
	Token     string            `json:"token"`
	WS        map[string]string `json:"ws"`
	APIBase   string            `json:"apiBase"`
	ExpiresAt string            `json:"expiresAt"`
}

type TaskStartRequest struct {
	Instruction string                 `json:"instruction"`
	Branch      string                 `json:"branch,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Agent       string                 `json:"agent,omitempty"`
}

type TaskStatusResponse struct {
	TaskID    string `json:"taskId"`
	Status    string `json:"status"`
	StartedAt string `json:"startedAt"`
	EndedAt   string `json:"endedAt,omitempty"`
}

type PatchesResponse struct {
	Patches []FilePatch `json:"patches"`
}

type FilePatch struct {
	File    string `json:"file"`
	Content string `json:"content"`
	Type    string `json:"type"`
}

type ApplyRequest struct {
	Select       []interface{} `json:"select"`
	CommitMessage string        `json:"commitMessage"`
}

type CmdRequest struct {
	Cmd       string `json:"cmd"`
	Cwd       string `json:"cwd"`
	TimeoutMs int    `json:"timeoutMs,omitempty"`
}

// Handlers
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (s *Server) createSession(w http.ResponseWriter, r *http.Request) {
	var req SessionCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate repo path - for now, just check it's not empty
	if req.Repo == "" {
		http.Error(w, "Repository path required", http.StatusBadRequest)
		return
	}

	// Create session using existing session manager
	sessionID, token, err := s.sessionManager.CreateSession(req.Repo, req.Label, req.Via)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Generate WebSocket URLs
	host := r.Host
	
	// Check if we're using relay mode
	var wsURLs map[string]string
	if req.Via == "relay" {
		// Return relay WebSocket URLs
		relayHost := getEnv("RELAY_HOST", "localhost:8081")
		wsURLs = map[string]string{
			"host":   fmt.Sprintf("ws://%s/ws/host", relayHost),
			"client": fmt.Sprintf("ws://%s/ws/client", relayHost),
		}
	} else {
		// Return local WebSocket URLs
		wsURLs = map[string]string{
			"pty":    fmt.Sprintf("ws://%s/ws/pty", host),
			"events": fmt.Sprintf("ws://%s/ws/events", host),
		}
	}

	response := SessionCreateResponse{
		SessionID: sessionID,
		Token:     token,
		WS:        wsURLs,
		APIBase:   fmt.Sprintf("http://%s", host),
		ExpiresAt: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) getSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	session, err := s.sessionManager.GetSession(sessionID)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	var req TaskStartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get session from token
	sessionID, err := auth.GetSessionIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Create task
	taskID, err := s.sessionManager.CreateTask(sessionID, req.Instruction, req.Branch, req.Context, req.Agent)
	if err != nil {
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"taskId": taskID,
		"status": "planning",
		"startedAt": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) getTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	task, err := s.sessionManager.GetTask(taskID)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"taskId":    taskID,
		"status":    task.Status,
		"startedAt": task.CreatedAt.Format(time.RFC3339),
		"endedAt":   task.UpdatedAt.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) getTaskPatches(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	patches, err := s.sessionManager.GetTaskPatches(taskID)
	if err != nil {
		http.Error(w, "Failed to get patches", http.StatusInternalServerError)
		return
	}

	// Convert to the expected format
	filePatches := make([]FilePatch, len(patches))
	for i, patch := range patches {
		filePatches[i] = FilePatch{
			File:    patch.File,
			Content: patch.Patch,
			Type:    "modified",
		}
	}

	response := PatchesResponse{
		Patches: filePatches,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) applyTaskPatches(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	var req ApplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Convert interface{} to map[string]interface{}
	selections := make([]map[string]interface{}, len(req.Select))
	for i, item := range req.Select {
		if mapItem, ok := item.(map[string]interface{}); ok {
			selections[i] = mapItem
		}
	}

	if err := s.sessionManager.ApplyTaskPatches(taskID, selections, req.CommitMessage); err != nil {
		http.Error(w, "Failed to apply patches", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok": true,
		"commit": "mock-commit-hash",
		"branch": "main",
	})
}

func (s *Server) runCommand(w http.ResponseWriter, r *http.Request) {
	var req CmdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// For now, just return 202 accepted
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok": true,
		"code": 0,
	})
}

func (s *Server) getGitDiff(w http.ResponseWriter, r *http.Request) {
	// Mock response
	patches := []FilePatch{
		{
			File:    "README.md",
			Content: "mock diff content",
			Type:    "modified",
		},
	}

	response := PatchesResponse{
		Patches: patches,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// WebSocket handlers
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *Server) handlePtyWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Send mock data
	for i := 0; i < 3; i++ {
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Mock output line %d\n", i+1)))
	}

	// Send exit message
	conn.WriteJSON(map[string]interface{}{
		"type": "exit",
		"code": 0,
		"durationMs": 1000,
	})
}

func (s *Server) handleEventsWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Send mock events
	events := []map[string]interface{}{
		{"type": "status", "phase": "planning"},
		{"type": "patch", "count": 1},
		{"type": "exit", "code": 0},
	}

	for _, event := range events {
		conn.WriteJSON(event)
		time.Sleep(500 * time.Millisecond)
	}
}

// Middleware
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowedOrigins := strings.Split(getEnv("CORS_ORIGINS", "http://localhost:19006"), ",")
		
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if strings.TrimSpace(allowedOrigin) == origin {
				allowed = true
				break
			}
		}

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
