package main

import (
	"context"
	"encoding/json"
	"fmt"

	githubql "github.com/shurcooL/githubv4"
	"go.uber.org/zap"
	"sigs.k8s.io/prow/pkg/github"
)

// GraphQL mutation for enabling auto-merge
const enableAutoMergeMutation = `
mutation EnableAutoMerge($pullRequestId: ID!, $mergeMethod: PullRequestMergeMethod!) {
  enablePullRequestAutoMerge(input: { pullRequestId: $pullRequestId, mergeMethod: $mergeMethod }) {
    pullRequest {
      id
    }
  }
}
`

type AutoMergeCondition struct {
	ExcludeLabels []string `yaml:"excludeLabels"`
	MergeMethod   string   `yaml:"mergeMethod"`
	MergeQueue    bool     `yaml:"mergeQueue"`
}

// String returns string representation of ApproveCondition.
func (condition *AutoMergeCondition) String() string {
	bytes, _ := json.Marshal(condition)
	return string(bytes)
}

type autoMergeResponse struct {
	EnablePullRequestAutoMerge struct {
		PullRequest struct {
			ID string `json:"id"`
		} `json:"pullRequest"`
	} `json:"enablePullRequestAutoMerge"`
}

// checkRequiredLabels checks if PR has all required labels.
func (condition *AutoMergeCondition) checkPrHasExcludeLabels(logger *zap.SugaredLogger, prLabels []github.Label) bool {
	defer logger.Sync()
	if condition.ExcludeLabels == nil {
		logger.Debug("No exclude labels defined")
		// No exclude labels found
		return false
	}
	logger.Debugf("Exclude labels defined: %v", condition.ExcludeLabels)
	logger.Debug("Getting pull request labels names")
	prLabelsNames := make(map[string]interface{})
	for _, label := range prLabels {
		prLabelsNames[label.Name] = nil
	}
	logger.Debug("Checking if PR has exclude labels")
	for _, excludeLabel := range condition.ExcludeLabels {
		if _, ok := prLabelsNames[excludeLabel]; ok {
			logger.Infof("PR has exclude label: %s", excludeLabel)
			return true
		}
	}
	logger.Debug("No exclude labels found")
	return false
}

// checkPrApproveConditions checks if PR meets conditions for auto approving.
// It validates all ApproveConditions for owner/repo/PR author entity.
func (hb *HandlerBackend) checkPrAutoMergeConditionsMatch(logger *zap.SugaredLogger, conditions []AutoMergeCondition, prLabels []github.Label) bool {
	defer logger.Sync()
	for _, condition := range conditions {
		logger.Debugw("Checking condition", "condition", condition)
		prExcluded := condition.checkPrHasExcludeLabels(logger, prLabels)
		if prExcluded {
			logger.Debug("PR does not meet auto merge conditions")
			return false
		}
	}
	logger.Debug("PR meets auto merge conditions")
	return true
}

// TODO (dekiel): Tests missing for enableAutoMerge
// enableAutoMerge enables auto merge for a pull request.
// It sends a GraphQL mutation to GitHub to enable auto merge.
func (hb *HandlerBackend) enableAutoMerge(ctx context.Context, logger *zap.SugaredLogger, prOrg, prRepo, prHeadSha string, prNumber int) error {
	defer logger.Sync()

	// vars := map[string]interface{}{
	// 	"pullRequestId": githubql.ID(fmt.Sprintf("%s", prNumber)),
	// "pullRequestId": strconv.Itoa(prNumber),
	// "mergeMethod":   "MERGE",
	// }

	mergeMethod := githubql.PullRequestMergeMethodMerge
	input := githubql.EnablePullRequestAutoMergeInput{
		PullRequestID: githubql.ID(fmt.Sprintf("%d", prNumber)),
		MergeMethod:   &mergeMethod,
	}

	// response := autoMergeResponse{}

	err := hb.Ghc.MutateWithGitHubAppsSupport(ctx, enableAutoMergeMutation, input, nil, prOrg)
	if err != nil {
		logger.Errorf("failed enable auto merge for pull request %s/%s#%d sha: %s, got error: %s",
			prOrg,
			prRepo,
			prNumber,
			prHeadSha,
			err)
		return err
	}

	logger.Infow(fmt.Sprintf("Auto merge enabled for pull request %s/%s#%d.", prOrg, prRepo, prNumber))
	return nil
}

// reviewPullRequest approves a pull request if it meets conditions.
// It searches conditions for owner/repo/PR author entity, validates them, waits for statuses to finish, validates their statuses, and approves PR.
func (hb *HandlerBackend) setPullRequestAutoMerge(ctx context.Context, logger *zap.SugaredLogger, prOrg, prRepo, prUser, prHeadSha string, prNumber int, prLabels []github.Label) {
	defer logger.Sync()
	defer hb.unlockPR(logger, prOrg, prRepo, prHeadSha, prNumber)
	logger.Debugf("Checking if auto merge conditions for PR %s/%s/%d exists", prOrg, prRepo, prNumber)
	autoMerge, autoMergeOk := hb.MergeConditions[prOrg][prRepo]
	if !autoMergeOk {
		logger.Infof("Merge conditions for PR %s/%s#%d not found", prOrg, prRepo, prNumber)
		return
	}

	logger.Debugf("Checking if PR %d meets approval conditions: %v", prNumber, hb.MergeConditions)
	logger.Sync() // Syncing logger to make sure all logs from calling GitHub API are written before logs from functions called in next steps.
	autoMergeMatched := hb.checkPrAutoMergeConditionsMatch(logger, autoMerge, prLabels)
	if !autoMergeMatched {
		return
	}

	// Check if context canceled to not review commit which is not a HEAD anymore.
	select {
	case <-ctx.Done():
		logger.Infof("Context canceled, skip approving pull request %s/%s#%d", prOrg, prRepo, prNumber)
		return
	default:
		err := hb.enableAutoMerge(ctx, logger, prOrg, prRepo, prHeadSha, prNumber)
		if err != nil {
			logger.Errorf("failed enable auto merge for pull request %s/%s#%d sha: %s, got error: %s",
				prOrg,
				prRepo,
				prNumber,
				prHeadSha,
				err)
		}
	}
}
