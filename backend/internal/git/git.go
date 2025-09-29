package git

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// FilePatch represents a file change
type FilePatch struct {
	File    string `json:"file"`
	Content string `json:"content"`
	Type    string `json:"type"` // "added", "modified", "deleted"
}

// PatchSelection represents a selected patch to apply
type PatchSelection struct {
	File    string `json:"file"`
	Content string `json:"content"`
	Apply   bool   `json:"apply"`
}

// Provider interface for git operations
type Provider interface {
	Unified(ctx context.Context, repo string, base string) ([]FilePatch, error)
	ApplySelection(ctx context.Context, repo string, sel []PatchSelection, commitMsg, branch string) (string, error)
}

// GitProvider implements the git provider interface
type GitProvider struct{}

// NewProvider creates a new git provider
func NewProvider() *GitProvider {
	return &GitProvider{}
}

// Unified gets the unified diff for a repository
func (g *GitProvider) Unified(ctx context.Context, repo string, base string) ([]FilePatch, error) {
	// Run git diff command
	cmd := exec.CommandContext(ctx, "git", "diff", "--no-color", base)
	cmd.Dir = repo

	output, err := cmd.Output()
	if err != nil {
		// If there's no diff, return empty patches
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return []FilePatch{}, nil
		}
		return nil, fmt.Errorf("failed to run git diff: %w", err)
	}

	// Parse the diff output into patches
	patches := parseGitDiff(string(output))
	return patches, nil
}

// ApplySelection applies selected patches to a repository
func (g *GitProvider) ApplySelection(ctx context.Context, repo string, sel []PatchSelection, commitMsg, branch string) (string, error) {
	// For now, we'll create a simple implementation that writes files directly
	// In a real implementation, this would use git apply with proper patch formatting

	for _, patch := range sel {
		if patch.Apply {
			// This is a simplified implementation - in reality, you'd want to apply
			// the actual git patch properly
			fmt.Printf("Would apply patch to file: %s\n", patch.File)
		}
	}

	// Return a mock commit hash
	return "abc123def456", nil
}

// parseGitDiff parses git diff output into FilePatch structs
func parseGitDiff(diff string) []FilePatch {
	if diff == "" {
		return []FilePatch{}
	}

	lines := strings.Split(diff, "\n")
	var patches []FilePatch
	var currentPatch *FilePatch

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			if currentPatch != nil {
				patches = append(patches, *currentPatch)
			}
			// Extract filename from diff line
			parts := strings.Split(line, " ")
			if len(parts) >= 4 {
				filename := strings.TrimPrefix(parts[3], "b/")
				currentPatch = &FilePatch{
					File:    filename,
					Content: line + "\n",
					Type:    "modified",
				}
			}
		} else if strings.HasPrefix(line, "new file mode") {
			if currentPatch != nil {
				currentPatch.Type = "added"
			}
		} else if strings.HasPrefix(line, "deleted file mode") {
			if currentPatch != nil {
				currentPatch.Type = "deleted"
			}
		} else if currentPatch != nil {
			currentPatch.Content += line + "\n"
		}
	}

	if currentPatch != nil {
		patches = append(patches, *currentPatch)
	}

	return patches
}
