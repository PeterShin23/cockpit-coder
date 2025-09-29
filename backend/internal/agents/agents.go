package agents

import (
	"context"
	"fmt"
	"time"

	"github.com/PeterShin23/cockpit-coder/backend/internal/git"
)

// Agent interface for AI agents
type Agent interface {
	StartTask(ctx context.Context, instruction, repo string) (string, error)
	StreamPTY(ctx context.Context, taskID string) (<-chan []byte, error)
	GetPatches(ctx context.Context, taskID string) ([]git.FilePatch, error)
	ApplyPatches(ctx context.Context, taskID string, sel []git.PatchSelection) error
	RunCommand(ctx context.Context, cmd string) (<-chan []byte, error)
}

// Factory interface for creating agents
type Factory interface {
	For(kind string) (Agent, error)
}

// agentFactory implements the Factory interface
type agentFactory struct{}

// NewFactory creates a new agent factory
func NewFactory() Factory {
	return &agentFactory{}
}

// For creates an agent by kind
func (f *agentFactory) For(kind string) (Agent, error) {
	switch kind {
	case "claude", "cline", "roo", "kilo":
		return &MockAgent{}, nil
	default:
		return &MockAgent{}, nil // Default to mock agent
	}
}

// MockAgent implements the Agent interface for testing
type MockAgent struct{}

// StartTask starts a mock task
func (m *MockAgent) StartTask(ctx context.Context, instruction, repo string) (string, error) {
	// Simulate task start
	return "mock-task-id", nil
}

// StreamPTY streams mock PTY output
func (m *MockAgent) StreamPTY(ctx context.Context, taskID string) (<-chan []byte, error) {
	ch := make(chan []byte, 100)

	go func() {
		defer close(ch)
		
		// Send some mock colored output
		lines := []string{
			"\033[32mStarting task...\033[0m\n",
			"\033[34mAnalyzing code...\033[0m\n",
			"\033[33mGenerating patches...\033[0m\n",
			"\033[32mTask completed!\033[0m\n",
		}

		for _, line := range lines {
			select {
			case ch <- []byte(line):
			case <-ctx.Done():
				return
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	return ch, nil
}

// GetPatches returns mock patches
func (m *MockAgent) GetPatches(ctx context.Context, taskID string) ([]git.FilePatch, error) {
	// Create a mock patch
	patch := git.FilePatch{
		File:    "mock_file.txt",
		Content: "This is a mock patch content\n+ Added new line\n- Removed old line",
		Type:    "modified",
	}

	return []git.FilePatch{patch}, nil
}

// ApplyPatches applies mock patches
func (m *MockAgent) ApplyPatches(ctx context.Context, taskID string, sel []git.PatchSelection) error {
	// Simulate applying patches
	fmt.Printf("Mock applying %d patches\n", len(sel))
	return nil
}

// RunCommand runs a mock command
func (m *MockAgent) RunCommand(ctx context.Context, cmd string) (<-chan []byte, error) {
	ch := make(chan []byte, 100)

	go func() {
		defer close(ch)
		
		// Send mock command output
		output := fmt.Sprintf("Mock output for command: %s\n", cmd)
		select {
		case ch <- []byte(output):
		case <-ctx.Done():
			return
		}
	}()

	return ch, nil
}
