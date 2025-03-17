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
	RequiredLabels []string `yaml:"requiredLabels"`
	MergeMethod    string   `yaml:"mergeMethod"`
	MergeQueue     bool     `yaml:"mergeQueue"`
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
func (condition *AutoMergeCondition) checkRequiredLabels(logger *zap.SugaredLogger, prLabels []github.Label) bool {
	defer logger.Sync()
	if condition.RequiredLabels == nil {
		logger.Debug("No required labels defined")
		// No required labels defined
		return true
	}
	labelNames := make(map[string]interface{})
	logger.Debugf("Checking if PR has all required labels: %v", condition.RequiredLabels)
	for _, label := range prLabels {
		labelNames[label.Name] = nil
	}
	for _, requiredLabel := range condition.RequiredLabels {
		if _, ok := labelNames[requiredLabel]; !ok {
			logger.Debugf("PR is missing required label: %s", requiredLabel)
			return false
		}
	}
	logger.Debug("All required labels are present")
	return true
}

// checkPrApproveConditions checks if PR meets conditions for auto approving.
// It validates all ApproveConditions for owner/repo/PR author entity.
func (hb *HandlerBackend) checkPrAutoMergeConditions(logger *zap.SugaredLogger, conditions []AutoMergeCondition, prLabels []github.Label) bool {
	defer logger.Sync()
	for _, condition := range conditions {
		logger.Debugw("Checking condition", "condition", condition)
		labelsMatched := condition.checkRequiredLabels(logger, prLabels)
		if !labelsMatched {
			logger.Debug("Labels not matched")
			continue
		}
		return true
	}
	logger.Debug("No conditions matched")
	return false
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
func (hb *HandlerBackend) autoMergePullRequest(ctx context.Context, logger *zap.SugaredLogger, prOrg, prRepo, prUser, prHeadSha string, prNumber int, prLabels []github.Label) {
	defer logger.Sync()
	defer hb.unlockPR(logger, prOrg, prRepo, prHeadSha, prNumber)
	logger.Debugf("Checking ...")
	logger.Debugf("Checking if conditions for PR author %s exists: %t", prUser, hb.Conditions[prOrg][prRepo][prUser] != nil)
	autoMerge, autoMergeOk := hb.MergeConditions[prOrg][prRepo][prUser]
	if !autoMergeOk {
		logger.Infof("Merge conditions for PR not found, skip automatic approval, pull request %s/%s#%d, author %s", prOrg, prRepo, prNumber, prUser)
		return
	}
	logger.Debugf("Checking if PR %d meets approval conditions: %v", prNumber, conditions)

	logger.Sync() // Syncing logger to make sure all logs from calling GitHub API are written before logs from functions called in next steps.
	autoMergeMatched := hb.checkPrAutoMergeConditions(logger, autoMerge, changes, prLabels)
	if !autoMergeMatched {
		return
	}
	err = hb.enableAutoMerge(ctx, logger, prOrg, prRepo, prHeadSha, prNumber)
	if err != nil {
		logger.Errorf("failed enable auto merge for pull request %s/%s#%d sha: %s, got error: %s",
			prOrg,
			prRepo,
			prNumber,
			prHeadSha,
			err)
	}
	conditionsMatched := hb.checkPrApproveConditions(logger, conditions, changes, prLabels)
	if !conditionsMatched {
		return
	}

	// wait for statuses to finish and return if not all are success.
	err = hb.waitForPrSuccessStatuses(ctx, logger, prOrg, prRepo, prHeadSha, prNumber)
	if err != nil {
		// TODO: Non success pr statuses are not error conditions for automated approver. We should log it as info.
		// 	Need to check if other type of errors
		logger.Errorf("pull request %s/%s#%d has non success statuses, got error: %s",
			prOrg,
			prRepo,
			prNumber,
			err)

		return
	}

	// Check if context canceled to not review commit which is not a HEAD anymore.
	select {
	case <-ctx.Done():
		logger.Infof("Context canceled, skip approving pull request %s/%s#%d", prOrg, prRepo, prNumber)
		return
	default:
		err = hb.createPRReview(prHeadSha, prOrg, prRepo, prNumber, logger)
		if err != nil {
			logger.Errorf("failed approve pull request %s/%s#%d, got error: %s",
		}
	}
}
