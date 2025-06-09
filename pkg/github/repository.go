package github

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/kyma-project/test-infra/pkg/pathsfilter" // Import the package with the ports
)

// Repository is a client for performing local Git operations.
// It acts as an adapter for the ChangedFilesProvider port.
type Repository struct {
	workingDir string
}

// NewRepository creates a new Git repository adapter.
func NewRepository(workingDir string) (*Repository, error) {
	if _, err := exec.LookPath("git"); err != nil {
		return nil, fmt.Errorf("command 'git' not found in PATH")
	}

	return &Repository{workingDir: workingDir}, nil
}

// GetChangedFiles retrieves the list of changed files, implementing the ChangedFilesProvider port.
func (r *Repository) GetChangedFiles(base, head string) ([]pathsfilter.ChangedFile, error) {
	cmd := exec.Command("git", "-C", r.workingDir, "diff", "--name-status", fmt.Sprintf("%s...%s", base, head))
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("failed to run 'git diff': %s, stderr: %s", exitErr, string(exitErr.Stderr))
		}

		return nil, err
	}

	var files []pathsfilter.ChangedFile
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			files = append(files, pathsfilter.ChangedFile{
				Status: parts[0],
				Path:   strings.Join(parts[1:], " "),
			})
		}
	}

	return files, nil
}
