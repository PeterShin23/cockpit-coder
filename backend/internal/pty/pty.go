package pty

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/PeterShin23/cockpit-coder/backend/internal/session"
)

// Manager handles PTY operations
type Manager struct {
	sessions map[string]*PtySession
	mu       sync.RWMutex
}

// PtySession represents a PTY session
type PtySession struct {
	ID        string
	Cmd       *exec.Cmd
	Pty       *os.File
	Size      *pty.Winsize
	Output    []byte
	Done      chan struct{}
	Cancel    context.CancelFunc
}

// NewManager creates a new PTY manager
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*PtySession),
	}
}

// RunCommand runs a command in the specified working directory
func (m *Manager) RunCommand(repo, cmd, cwd string, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create command
	fullCmd := exec.CommandContext(ctx, "sh", "-c", cmd)
	fullCmd.Dir = cwd

	// Start the command
	output, err := fullCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("command failed: %v, output: %s", err, string(output))
	}

	return output, nil
}

// CreatePtySession creates a new PTY session
func (m *Manager) CreatePtySession(repo, cmd string) (string, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create command
	fullCmd := exec.CommandContext(ctx, "sh", "-c", cmd)
	fullCmd.Dir = repo

	// Create PTY
	ptyFile, err := pty.Start(fullCmd)
	if err != nil {
		cancel()
		return "", fmt.Errorf("failed to create PTY: %v", err)
	}

	sessionID := generateID()
	session := &PtySession{
		ID:     sessionID,
		Cmd:    fullCmd,
		Pty:    ptyFile,
		Size:   &pty.Winsize{Cols: 80, Rows: 24},
		Output: make([]byte, 0),
		Done:   make(chan struct{}),
		Cancel: cancel,
	}

	// Start reading from PTY
	go m.readFromPty(session)

	m.mu.Lock()
	m.sessions[sessionID] = session
	m.mu.Unlock()

	return sessionID, nil
}

// readFromPty reads data from PTY and stores it
func (m *Manager) readFromPty(session *PtySession) {
	defer close(session.Done)
	defer session.Pty.Close()

	reader := bufio.NewReader(session.Pty)
	for {
		select {
		case <-session.Done:
			return
		default:
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					session.Output = append(session.Output, []byte(fmt.Sprintf("\nError reading PTY: %v\n", err))...)
				}
				return
			}
			session.Output = append(session.Output, line...)
		}
	}
}

// WriteToPty writes data to a PTY session
func (m *Manager) WriteToPty(sessionID string, data []byte) error {
	m.mu.RLock()
	session, exists := m.sessions[sessionID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("session not found")
	}

	_, err := session.Pty.Write(data)
	return err
}

// ResizePty resizes a PTY session
func (m *Manager) ResizePty(sessionID string, cols, rows uint16) error {
	m.mu.RLock()
	session, exists := m.sessions[sessionID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("session not found")
	}

	session.Size = &pty.Winsize{Cols: cols, Rows: rows}
	return pty.Setsize(session.Pty, session.Size)
}

// GetPtyOutput retrieves the output of a PTY session
func (m *Manager) GetPtyOutput(sessionID string) ([]byte, error) {
	m.mu.RLock()
	session, exists := m.sessions[sessionID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	return session.Output, nil
}

// ClosePtySession closes a PTY session
func (m *Manager) ClosePtySession(sessionID string) error {
	m.mu.Lock()
	session, exists := m.sessions[sessionID]
	m.mu.Unlock()

	if !exists {
		return fmt.Errorf("session not found")
	}

	// Cancel the context to stop the command
	session.Cancel()

	// Close the PTY
	if session.Pty != nil {
		session.Pty.Close()
	}

	// Remove from sessions
	m.mu.Lock()
	delete(m.sessions, sessionID)
	m.mu.Unlock()

	return nil
}

// ListPtySessions returns all active PTY sessions
func (m *Manager) ListPtySessions() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]string, 0, len(m.sessions))
	for id := range m.sessions {
		sessions = append(sessions, id)
	}
	return sessions
}

// generateID creates a random ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
