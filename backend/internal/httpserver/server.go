package httpserver

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PeterShin23/cockpit-coder/backend/internal/auth"
	"github.com/PeterShin23/cockpit-coder/backend/internal/session"
	"github.com/PeterShin23/cockpit-coder/backend/internal/pty"
	"github.com/gorilla/mux"
)

type Server struct {
	router        *mux.Router
	sessionManager *session.Manager
	ptyManager    *pty.Manager
}

func NewServer(sessionManager *session.Manager, ptyManager *pty.Manager) *Server {
	s := &Server{
		router:        mux.NewRouter(),
		sessionManager: sessionManager,
		ptyManager:    ptyManager,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
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
	
	// WebSocket endpoints
	s.router.HandleFunc("/ws/pty", s.handlePtyWebSocket)
	s.router.HandleFunc("/ws/events", s.handleEventsWebSocket)
	
	// CORS middleware
	s.router.Use(s.corsMiddleware)
}

func (s *Server) ListenAndServe() error {
	port := ":" + getEnv("PORT", "8080")
	return http.ListenAndServe(port, s.router)
}

func (s *Server) createSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Repo  string `json:"repo"`
		Label string `json:"label"`
		Via   string `json:"via"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate repo path
	if !s.isValidRepoPath(req.Repo) {
		http.Error(w, "Invalid repository path", http.StatusBadRequest)
		return
	}

	// Create session
	sessionID, token, err := s.sessionManager.CreateSession(req.Repo, req.Label, req.Via)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Generate WebSocket URLs
	wsURL := getEnv("WS_URL", "ws://localhost:8080")
	apiBase := getEnv("API_BASE", "http://localhost:8080")

	response := map[string]interface{}{
		"sessionId": sessionID,
		"token":     token,
		"ws": map[string]string{
			"pty":  wsURL + "/ws/pty",
			"events": wsURL + "/ws/events",
		},
		"apiBase":   apiBase,
		"expiresAt": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
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
	var req struct {
		Instruction string                 `json:"instruction"`
		Branch      string                 `json:"branch"`
		Context     map[string]interface{} `json:"context"`
		Agent       string                 `json:"agent"`
	}

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
		"id":         taskID,
		"status":     "pending",
		"createdAt":  time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (s *Server) getTaskPatches(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	patches, err := s.sessionManager.GetTaskPatches(taskID)
	if err != nil {
		http.Error(w, "Failed to get patches", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"patches": patches,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) applyTaskPatches(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	var req struct {
		Select       []map[string]interface{} `json:"select"`
		CommitMessage string                   `json:"commitMessage"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.sessionManager.ApplyTaskPatches(taskID, req.Select, req.CommitMessage); err != nil {
		http.Error(w, "Failed to apply patches", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Patches applied successfully"})
}

func (s *Server) runCommand(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Cmd       string `json:"cmd"`
		Cwd       string `json:"cwd"`
		TimeoutMs int    `json:"timeoutMs"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate command
	if !s.isAllowedCommand(req.Cmd) {
		http.Error(w, "Command not allowed", http.StatusForbidden)
		return
	}

	// Get session from token
	sessionID, err := auth.GetSessionIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get session info for repo path
	session, err := s.sessionManager.GetSession(sessionID)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Run command
	output, err := s.ptyManager.RunCommand(session.Repo, req.Cmd, req.Cwd, time.Duration(req.TimeoutMs)*time.Millisecond)
	if err != nil {
		http.Error(w, "Command execution failed", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"output": string(output),
		"status": "completed",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handlePtyWebSocket(w http.ResponseWriter, r *http.Request) {
	// WebSocket upgrade handled by pty hub
}

func (s *Server) handleEventsWebSocket(w http.ResponseWriter, r *http.Request) {
	// WebSocket upgrade handled by events hub
}

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

func (s *Server) isValidRepoPath(path string) bool {
	// Implement path validation and allowlist checking
	repoAllowlist := strings.Split(getEnv("REPO_ALLOWLIST", ""), ",")
	for _, allowedPath := range repoAllowlist {
		if strings.TrimSpace(allowedPath) == path {
			return true
		}
	}
	return false
}

func (s *Server) isAllowedCommand(cmd string) bool {
	// Implement command allowlist checking
	cmdAllowlist := strings.Split(getEnv("CMD_ALLOWLIST", ""), ",")
	for _, allowedCmd := range cmdAllowlist {
		if strings.TrimSpace(allowedCmd) == cmd {
			return true
		}
	}
	return false
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
