package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/PeterShin23/cockpit-coder/backend/internal/auth"
)

// Task represents a coding task
type Task struct {
	ID          string                 `json:"id"`
	SessionID   string                 `json:"sessionId"`
	Instruction string                 `json:"instruction"`
	Branch      string                 `json:"branch"`
	Context     map[string]interface{} `json:"context"`
	Agent       string                 `json:"agent"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	Patches     []Patch                `json:"patches"`
}

// Patch represents a code patch
type Patch struct {
	File  string `json:"file"`
	Patch string `json:"patch"`
}

// MemoryManager handles session and task management
type MemoryManager struct {
	sessions map[string]*Session
	tasks    map[string]*Task
	mu       sync.RWMutex
}

// NewMemoryManager creates a new in-memory session manager
func NewMemoryManager() *MemoryManager {
	manager := &MemoryManager{
		sessions: make(map[string]*Session),
		tasks:    make(map[string]*Task),
	}

	// Start cleanup goroutine
	go manager.cleanupExpiredSessions()

	return manager
}

// Create creates a new session
func (m *MemoryManager) Create(ctx context.Context, repo string, ttl time.Duration) (Session, error) {
	sessionID := generateID()

	session := &Session{
		ID:        sessionID,
		Repo:      repo,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	}

	m.mu.Lock()
	m.sessions[sessionID] = session
	m.mu.Unlock()

	return *session, nil
}

// Get retrieves a session by ID
func (m *MemoryManager) Get(ctx context.Context, id string) (Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[id]
	if !exists {
		return Session{}, false
	}

	if time.Now().After(session.ExpiresAt) {
		return Session{}, false
	}

	return *session, true
}

// Touch updates the session's expiration time
func (m *MemoryManager) Touch(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[id]
	if !exists {
		return errors.New("session not found")
	}

	session.ExpiresAt = time.Now().Add(24 * time.Hour)
	return nil
}

// End removes a session
func (m *MemoryManager) End(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, id)
	return nil
}

// cleanupExpiredSessions removes expired sessions periodically
func (m *MemoryManager) cleanupExpiredSessions() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for id, session := range m.sessions {
			if now.After(session.ExpiresAt) {
				delete(m.sessions, id)
			}
		}
		m.mu.Unlock()
	}
}

// generateID creates a random ID
func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Legacy methods for compatibility with existing code
func (m *MemoryManager) CreateSession(repo, label, via string) (string, string, error) {
	sessionID := generateID()
	
	var token string
	var err error
	
	if via == "relay" {
		// Generate relay-compatible token
		// Use "t_demo" as default tenant ID for demo purposes
		token, err = auth.GenerateRelayToken(sessionID, "t_demo", 86400) // 24 hours
	} else {
		// Generate standard local token
		token, err = auth.GenerateToken(sessionID)
	}
	
	if err != nil {
		return "", "", err
	}

	session := &Session{
		ID:        sessionID,
		Repo:      repo,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	m.mu.Lock()
	m.sessions[sessionID] = session
	m.mu.Unlock()

	return sessionID, token, nil
}

func (m *MemoryManager) GetSession(sessionID string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, errors.New("session not found")
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, errors.New("session expired")
	}

	return session, nil
}

func (m *MemoryManager) CreateTask(sessionID, instruction, branch string, context map[string]interface{}, agent string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, exists := m.sessions[sessionID]
	if !exists {
		return "", errors.New("session not found")
	}

	taskID := generateID()
	task := &Task{
		ID:          taskID,
		SessionID:   sessionID,
		Instruction: instruction,
		Branch:      branch,
		Context:     context,
		Agent:       agent,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	m.tasks[taskID] = task
	return taskID, nil
}

func (m *MemoryManager) GetTask(taskID string) (*Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return nil, errors.New("task not found")
	}

	return task, nil
}

func (m *MemoryManager) GetTaskPatches(taskID string) ([]Patch, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return nil, errors.New("task not found")
	}

	return task.Patches, nil
}

func (m *MemoryManager) ApplyTaskPatches(taskID string, selections []map[string]interface{}, commitMessage string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return errors.New("task not found")
	}

	// In a real implementation, this would apply the patches to the actual code
	// For now, we'll just update the task status
	task.Status = "completed"
	task.UpdatedAt = time.Now()

	return nil
}

func (m *MemoryManager) UpdateTaskStatus(taskID, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return errors.New("task not found")
	}

	task.Status = status
	task.UpdatedAt = time.Now()
	return nil
}

func (m *MemoryManager) AddPatchesToTask(taskID string, patches []Patch) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return errors.New("task not found")
	}

	task.Patches = append(task.Patches, patches...)
	task.UpdatedAt = time.Now()
	return nil
}
