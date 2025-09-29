package cmdexec

import (
	"context"
	"strings"
	"time"

	"github.com/PeterShin23/cockpit-coder/backend/internal/pty"
)

// Runner interface for command execution
type Runner interface {
	Run(ctx context.Context, cmd string, cwd string, timeout time.Duration) (pty.Proc, error)
	Allowed(cmd string) bool
}

// CmdRunner implements the command runner interface
type CmdRunner struct {
	policy PolicyWrapper
	pty    pty.Manager
}

// PolicyWrapper wraps the policy interface
type PolicyWrapper struct {
	IsCmdAllowed func(cmd string) bool
}

// NewRunner creates a new command runner
func NewRunner(policy PolicyWrapper, ptyManager pty.Manager) *CmdRunner {
	return &CmdRunner{
		policy: policy,
		pty:    ptyManager,
	}
}

// Run executes a command under PTY
func (r *CmdRunner) Run(ctx context.Context, cmd string, cwd string, timeout time.Duration) (pty.Proc, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Split command into executable and arguments
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil, nil
	}

	executable := parts[0]
	var args []string
	if len(parts) > 1 {
		args = parts[1:]
	}

	// Start the command under PTY
	proc, err := r.pty.Start(ctx, executable, args, cwd, nil)
	if err != nil {
		return nil, err
	}

	return proc, nil
}

// Allowed checks if a command is allowed by policy
func (r *CmdRunner) Allowed(cmd string) bool {
	return r.policy.IsCmdAllowed(cmd)
}
