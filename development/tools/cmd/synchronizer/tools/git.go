package tools

import (
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

// FindCommitHash returns git hash for specific directory
func FindCommitHash(dir string, componentPath string) (string, error) {
	result, err := findCommitParameter(dir, componentPath, "H")
	if err != nil {
		return "", errors.Wrap(err, "during find commit hash")
	}

	return strings.Trim(result, "\""), nil
}

// FindCommitDate returns git commit date for specific directory
func FindCommitDate(dir string, componentPath string) (string, error) {
	result, err := findCommitParameter(dir, componentPath, "ct")
	if err != nil {
		return "", errors.Wrap(err, "during find commit date")
	}

	return strings.Trim(result, "\""), nil
}

// FindFileDifference returns list of files which were changed before two commits
func FindFileDifference(dir, componentPath, hashFrom, hashTo string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", hashFrom, hashTo, componentPath)
	cmd.Dir = dir

	out, err := cmd.CombinedOutput()
	result := string(out)
	if err != nil {
		return []string{}, errors.Wrap(err, "during run git command")
	}

	result = strings.Trim(result, "\n")
	return strings.Split(result, "\n"), nil
}

func findCommitParameter(dir string, componentPath string, format string) (string, error) {
	format = "--pretty=format:\"%" + format + "\""

	// run git command: 'git log -1 --pretty=format"%?" path/to/component' where '?' is format of data to get
	cmd := exec.Command("git", "log", "-1", format, componentPath)
	cmd.Dir = dir

	out, err := cmd.CombinedOutput()
	result := string(out)
	if err != nil {
		return "", errors.Wrap(err, "during run git command")
	}
	if result == "" {
		return "", errors.Errorf("Result for git log command in %q directory is empty\n", format, componentPath)
	}

	return result, nil
}
