package session

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/PeterShin23/cockpit-coder/backend/internal/auth"
)

// Session represents a user session
type Session struct {
	ID        string    `json:"id"`
	Repo      string    `json:"repo"`
	Label     string    `json:"label"`
	Via       string    `json:"via"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	Token     string    `json:"token"`
}

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

// Manager handles session and task management
type Manager struct {
	sessions map[string]*Session
	tasks    map[string]*Task
	mu       sync.RWMutex
}

// NewMemoryManager creates a new in-memory session manager
func NewMemoryManager() *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
		tasks:    make(map[string]*Task),
	}
}

// CreateSession creates a new session
func (m *Manager) CreateSession(repo, label, via string) (string, string, error) {
	sessionID := generateID()
	token, err := auth.GenerateToken(sessionID)
	if err != nil {
		return "", "", err
	}

	session := &Session{
		ID:        sessionID,
		Repo:      repo,
		Label:     label,
		Via:       via,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Token:     token,
	}

	m.mu.Lock()
	m.sessions[sessionID] = session
	m.mu.Unlock()

	return sessionID, token, nil
}

// GetSession retrieves a session by ID
func (m *Manager) GetSession(sessionID string) (*Session, error) {
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

// CreateTask creates a new task for a session
func (m *Manager) CreateTask(sessionID, instruction, branch string, context map[string]interface{}, agent string) (string, error) {
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

// GetTask retrieves a task by ID
func (m *Manager) GetTask(taskID string) (*Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return nil, errors.New("task not found")
	}

	return task, nil
}

// GetTaskPatches retrieves patches for a task
func (m *Manager) GetTaskPatches(taskID string) ([]Patch, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return nil, errors.New("task not found")
	}

	return task.Patches, nil
}

// ApplyTaskPatches applies selected patches to a task
func (m *Manager) ApplyTaskPatches(taskID string, selections []map[string]interface{}, commitMessage string) error {
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

// UpdateTaskStatus updates the status of a task
func (m *Manager) UpdateTaskStatus(taskID, status string) error {
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

// AddPatchesToTask adds patches to a task
func (m *Manager) AddPatchesToTask(taskID string, patches []Patch) error {
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

// generateID creates a random ID
func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
