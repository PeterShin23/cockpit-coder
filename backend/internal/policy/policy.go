package policy

import (
	"path/filepath"
	"strings"
	"time"
)

// Policy holds the security and configuration policies
type Policy struct {
	RepoAllowlist        []string
	CmdAllowlist         []string
	MaxCmdDuration       time.Duration
	DenyNetworkByDefault bool
}

// NewPolicy creates a new policy instance
func NewPolicy(repoAllowlist, cmdAllowlist []string, maxCmdDuration time.Duration, denyNetworkByDefault bool) *Policy {
	// Clean up allowlists
	cleanRepoAllowlist := make([]string, 0, len(repoAllowlist))
	for _, repo := range repoAllowlist {
		trimmed := strings.TrimSpace(repo)
		if trimmed != "" {
			cleanRepoAllowlist = append(cleanRepoAllowlist, trimmed)
		}
	}

	cleanCmdAllowlist := make([]string, 0, len(cmdAllowlist))
	for _, cmd := range cmdAllowlist {
		trimmed := strings.TrimSpace(cmd)
		if trimmed != "" {
			cleanCmdAllowlist = append(cleanCmdAllowlist, trimmed)
		}
	}

	return &Policy{
		RepoAllowlist:        cleanRepoAllowlist,
		CmdAllowlist:         cleanCmdAllowlist,
		MaxCmdDuration:       maxCmdDuration,
		DenyNetworkByDefault: denyNetworkByDefault,
	}
}

// IsRepoAllowed checks if a repository path is allowed
func (p *Policy) IsRepoAllowed(path string) bool {
	cleanPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	for _, allowedPath := range p.RepoAllowlist {
		allowedAbsPath, err := filepath.Abs(allowedPath)
		if err != nil {
			continue
		}

		// Check if the path is within the allowed path
		rel, err := filepath.Rel(allowedAbsPath, cleanPath)
		if err != nil {
			continue
		}

		// If rel doesn't start with "..", then cleanPath is within allowedPath
		if !strings.HasPrefix(rel, "..") {
			return true
		}
	}

	return false
}

// IsCmdAllowed checks if a command is allowed
func (p *Policy) IsCmdAllowed(cmd string) bool {
	trimmedCmd := strings.TrimSpace(cmd)
	if trimmedCmd == "" {
		return false
	}

	for _, allowedCmd := range p.CmdAllowlist {
		if strings.TrimSpace(allowedCmd) == trimmedCmd {
			return true
		}
	}

	return false
}
