package git

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// ChangedFile represents a single file that has been changed.
type ChangedFile struct {
	Path   string
	Status string
}

// Repository is a client for performing local Git operations.
type Repository struct {
	workingDir string
}

// NewRepository creates a new local Git repository client.
func NewRepository(workingDir string) (*Repository, error) {
	if _, err := exec.LookPath("git"); err != nil {
		return nil, fmt.Errorf("command 'git' not found in PATH")
	}

	return &Repository{workingDir: workingDir}, nil
}

// GetChangedFiles retrieves the list of changed files between two git refs.
func (r *Repository) GetChangedFiles(base, head string) ([]ChangedFile, error) {
	cmd := exec.Command("git", "-C", r.workingDir, "diff", "--name-status", fmt.Sprintf("%s...%s", base, head))
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("failed to run 'git diff': %s, stderr: %s", exitErr, string(exitErr.Stderr))
		}

		return nil, err
	}

	var files []ChangedFile
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			files = append(files, ChangedFile{
				Status: parts[0],
				Path:   strings.Join(parts[1:], " "),
			})
		}
	}

	return files, nil
}
