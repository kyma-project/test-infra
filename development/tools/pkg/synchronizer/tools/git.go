package tools

import (
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// FindCommitHash returns git hash for specific directory
func FindCommitHash(dir string, componentPath string) (string, error) {
	format := "--pretty=format:\"%H\""
	cmd := exec.Command("git", "log", "-1", format, componentPath)
	cmd.Dir = dir

	out, err := cmd.CombinedOutput()
	result := string(out)
	if err != nil {
		return "", errors.Wrap(err, "while run 'git log -1' with single hash parameter")
	}
	if result == "" {
		return "", errors.Errorf("Result for git log command in %q directory is empty\n", componentPath)
	}

	return strings.Trim(result, "\""), nil
}

// FetchCommitsHistory returns git hash commits and commit dates for specific directory limited by days sorted by commit dates
func FetchCommitsHistory(dir string, componentPath string, daysLimit int) (map[int64]string, error) {
	format := "--pretty=format:\"%H-%ct\""
	cmd := exec.Command("git", "log", format, componentPath)
	cmd.Dir = dir

	out, err := cmd.CombinedOutput()
	result := string(out)
	if err != nil {
		return nil, errors.Wrap(err, "while run git command")
	}
	if result == "" {
		return nil, errors.Errorf("Result for git log command in with format %q in %q directory is empty\n", format, componentPath)
	}

	history, err := generateHistoryFromCommandResult(result)
	if err != nil {
		return nil, errors.Wrap(err, "while transform git result command to result map")
	}

	return limitGitHistoryTimePeriod(history, daysLimit), nil
}

func generateHistoryFromCommandResult(result string) (map[int64]string, error) {
	history := map[int64]string{}

	result = strings.Trim(result, "\n")
	gitLogs := strings.Split(result, "\n")

	historyParts := [][]string{}
	for _, gitLog := range gitLogs {
		gitLog = strings.Trim(gitLog, "\"")
		gitLogElement := strings.Split(gitLog, "-")
		if len(gitLogElement) != 2 {
			return nil, errors.New("each log should contain git hash and commit date separate by dash")
		}
		if len(gitLogElement[1]) != 10 {
			return nil, errors.New("each log should contain commit date in unix time")
		}

		historyParts = append(historyParts, gitLogElement)
	}

	for _, values := range historyParts {
		unixDate, err := strconv.ParseInt(values[1], 10, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "while parsing %s string to int64", values[1])
		}
		history[unixDate] = values[0]
	}

	return history, nil
}

func limitGitHistoryTimePeriod(history map[int64]string, outOfDateDays int) map[int64]string {
	output := make(map[int64]string)

	timeLimit := time.Now().AddDate(0, 0, -(outOfDateDays)).Unix()
	for k, v := range history {
		if k < timeLimit {
			continue
		}
		output[k] = v
	}

	return output
}

// FindFileDifference returns list of files which were changed before two commits
func FindFileDifference(dir, componentPath, hashFrom, hashTo string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", hashFrom, hashTo, componentPath)
	cmd.Dir = dir

	out, err := cmd.CombinedOutput()
	result := string(out)
	if err != nil {
		return []string{}, errors.Wrap(err, "while run git command")
	}

	result = strings.Trim(result, "\n")
	return strings.Split(result, "\n"), nil
}
