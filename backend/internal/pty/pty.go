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
)

// Proc represents a running process
type Proc interface {
	Write([]byte) (int, error)
	Resize(cols, rows int) error
	Close() error
	Stream() <-chan []byte
	Done() <-chan State
}

// State represents the final state of a process
type State struct {
	ExitCode int
	Err      error
	Duration time.Duration
}

// Manager interface for PTY operations
type Manager interface {
	Start(ctx context.Context, cmd string, args []string, cwd string, env []string) (Proc, error)
}

// managerImpl implements the Manager interface
type managerImpl struct {
	sessions map[string]*ptySession
	mu       sync.RWMutex
}

// ptySession implements the Proc interface
type ptySession struct {
	ID       string
	Cmd      *exec.Cmd
	PtyFile  *os.File
	Size     *pty.Winsize
	StreamCh chan []byte
	DoneCh   chan State
	Cancel   context.CancelFunc
	start    time.Time
}

// NewManager creates a new PTY manager
func NewManager() Manager {
	return &managerImpl{
		sessions: make(map[string]*ptySession),
	}
}

// Start creates and starts a new PTY process
func (m *managerImpl) Start(ctx context.Context, cmd string, args []string, cwd string, env []string) (Proc, error) {
	// Create command
	fullCmd := exec.CommandContext(ctx, cmd, args...)
	fullCmd.Dir = cwd
	if env != nil {
		fullCmd.Env = append(os.Environ(), env...)
	}

	// Create PTY
	ptyFile, err := pty.Start(fullCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create PTY: %v", err)
	}

	// Create session
	session := &ptySession{
		ID:       generateID(),
		Cmd:      fullCmd,
		PtyFile:  ptyFile,
		Size:     &pty.Winsize{Cols: 80, Rows: 24},
		StreamCh: make(chan []byte, 1000),
		DoneCh:   make(chan State, 1),
		start:    time.Now(),
	}

	// Start reading from PTY
	go session.readFromPty()

	// Store session
	m.mu.Lock()
	m.sessions[session.ID] = session
	m.mu.Unlock()

	return session, nil
}

// readFromPty reads data from PTY and streams it
func (s *ptySession) readFromPty() {
	defer close(s.StreamCh)
	defer close(s.DoneCh)
	defer s.PtyFile.Close()

	reader := bufio.NewReader(s.PtyFile)
	for {
		// Read with timeout
		s.PtyFile.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if os.IsTimeout(err) {
				continue
			}
			if err != io.EOF {
				select {
				case s.StreamCh <- []byte(fmt.Sprintf("\nError reading PTY: %v\n", err)):
				default:
				}
			}
			// Process finished
			s.handleProcessExit()
			return
		}
		select {
		case s.StreamCh <- line:
		default:
			// Channel full, drop data
		}
	}
}

// handleProcessExit handles process completion
func (s *ptySession) handleProcessExit() {
	duration := time.Since(s.start)
	
	// Get exit code
	exitCode := 0
	if err := s.Cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	state := State{
		ExitCode: exitCode,
		Err:      nil,
		Duration: duration,
	}

	select {
	case s.DoneCh <- state:
	default:
	}
}

// Write writes data to the PTY
func (s *ptySession) Write(data []byte) (int, error) {
	return s.PtyFile.Write(data)
}

// Resize resizes the PTY
func (s *ptySession) Resize(cols, rows int) error {
	s.Size = &pty.Winsize{Cols: uint16(cols), Rows: uint16(rows)}
	return pty.Setsize(s.PtyFile, s.Size)
}

// Close closes the PTY session
func (s *ptySession) Close() error {
	// Cancel the context to stop the command
	if s.Cmd != nil && s.Cmd.Process != nil {
		s.Cmd.Process.Kill()
	}
	
	// Close the PTY file
	if s.PtyFile != nil {
		s.PtyFile.Close()
	}

	return nil
}

// Stream returns the output stream channel
func (s *ptySession) Stream() <-chan []byte {
	return s.StreamCh
}

// Done returns the completion channel
func (s *ptySession) Done() <-chan State {
	return s.DoneCh
}

// generateID creates a random ID
func generateID() string {
	return fmt.Sprintf("pty_%d_%d", time.Now().UnixNano(), os.Getpid())
}
