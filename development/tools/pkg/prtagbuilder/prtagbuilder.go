package prtagbuilder

import (
	"bufio"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/v31/github"
	"k8s.io/test-infra/prow/pod-utils/downwardapi"
)

// findPRNumber match commit message with regex to extract pull request number. By default github add pr number to the commit message.
func findPRNumber(commit *github.RepositoryCommit) (string, error) {
	re := regexp.MustCompile(`(?s)^.*\(#(?P<prNumber>\d*)?\)\s*$`)
	messageReader := strings.NewReader(commit.Commit.GetMessage())
	scanner := bufio.NewScanner(messageReader)
	scanner.Scan()
	matches := re.FindStringSubmatch(scanner.Text())
	if len(matches) != 2 {
		return "", fmt.Errorf("failed find PR number in first line of commit message: %s, found %d matched strings", scanner.Text(), len(matches))
	}
	return matches[1], nil
}

// verifyPR checks if pull request merge commit match provided commit SHA.
func verifyPR(pr *github.PullRequest, commitSHA string) error {
	if *pr.Merged {
		if *pr.MergeCommitSHA != commitSHA {
			return fmt.Errorf("commit SHA %s and matched PR merge commit SHA %s, doesn't match", commitSHA, *pr.MergeCommitSHA)
		}
	}
	return nil
}

// BuildPrTag will extract PR number from commit message, search PR, validate if correct PR was found and print pr tag.
func BuildPrTag(jobSpec *downwardapi.JobSpec, fromFlags bool, numberOnly bool) (string, error) {
	var (
		ctx    = context.Background()
		commit *github.RepositoryCommit
	)

	// create github client
	client := github.NewClient(nil)
	if fromFlags {
		// get commit for a branch
		branch, _, err := client.Repositories.GetBranch(ctx, jobSpec.Refs.Org, jobSpec.Refs.Repo, jobSpec.Refs.BaseRef)
		if err != nil {
			return "", fmt.Errorf("failed get branch %s, got error: %w", jobSpec.Refs.BaseRef, err)
		}
		commit = branch.GetCommit()
		jobSpec.Refs.BaseSHA = *commit.SHA
	} else {
		var err error
		// get git base reference from postsubmit environment variables
		jobSpec, err = downwardapi.ResolveSpecFromEnv()
		if err != nil {
			return "", fmt.Errorf("failed to read JOB_SPEC env, got error: %w", err)
		}
		// get commit details for base sha
		commit, _, err = client.Repositories.GetCommit(ctx, jobSpec.Refs.Org, jobSpec.Refs.Repo, jobSpec.Refs.BaseSHA)
		if err != nil {
			return "", fmt.Errorf("failed get commit %s, got error: %w", jobSpec.Refs.BaseSHA, err)
		}
	}
	// extract pull request number from commit message
	prNumberAsString, err := findPRNumber(commit)
	if err != nil {
		return "", fmt.Errorf("failed get PR number, got error: %w", err)
	}
	prNumber, err := strconv.Atoi(prNumberAsString)
	if err != nil {
		return "", fmt.Errorf("failed convert PR number to integer, got error: %w", err)
	}
	// get pull request details for extracted pr
	pr, _, err := client.PullRequests.Get(ctx, jobSpec.Refs.Org, jobSpec.Refs.Repo, prNumber)
	if err != nil {
		return "", fmt.Errorf("failed get Pull Request number %d, got error: %w", prNumber, err)
	}
	// check if correct pr was found
	if err := verifyPR(pr, jobSpec.Refs.BaseSHA); err != nil {
		return "", fmt.Errorf("pr verification failed, got error: %w", err)
	}
	if numberOnly {
		return fmt.Sprintf("%d", prNumber), nil
	}
	// print PR image tag
	return fmt.Sprintf("PR-%d", prNumber), nil
}
