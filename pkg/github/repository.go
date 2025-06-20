package github

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// Repository is a client for performing local Git operations.
// It acts as an adapter for the ChangedFilesProvider port.
type Repository struct {
	repoPath string
}

// NewRepository creates a new Git repository adapter.
func NewRepository(repoPath string) (*Repository, error) {
	return &Repository{repoPath: repoPath}, nil
}

// GetChangedFiles retrieves the list of changed files, implementing the ChangedFilesProvider port.
func (r *Repository) GetChangedFiles(base string, head string) ([]string, error) {
	repo, err := git.PlainOpen(r.repoPath)
	if err != nil {
		return nil, fmt.Errorf("could not open git repository at %s: %w", r.repoPath, err)
	}

	headHash, err := repo.ResolveRevision(plumbing.Revision(head))
	if err != nil {
		return nil, fmt.Errorf("could not resolve head revision '%s': %w", head, err)
	}

	headCommit, err := repo.CommitObject(*headHash)
	if err != nil {
		return nil, fmt.Errorf("could not get head commit object '%s': %w", head, err)
	}

	baseHash, err := repo.ResolveRevision(plumbing.Revision(base))
	if err != nil {
		return nil, fmt.Errorf("could not resolve base revision '%s': %w", base, err)
	}

	baseCommit, err := repo.CommitObject(*baseHash)
	if err != nil {
		return nil, fmt.Errorf("could not get base commit object '%s': %w", base, err)
	}

	patch, err := baseCommit.Patch(headCommit)
	if err != nil {
		return nil, fmt.Errorf("could not generate patch between base and head: %w", err)
	}

	var files []string
	for _, filePatch := range patch.FilePatches() {
		from, to := filePatch.Files()
		if to != nil {
			files = append(files, to.Path())
		} else if from != nil {
			files = append(files, from.Path())
		}
	}

	return files, nil
}
