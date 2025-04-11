package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
	"sigs.k8s.io/prow/pkg/github"
)

// String returns string representation of ApproveCondition.
func (ac *ApproveCondition) String() string {
	b, _ := json.Marshal(ac)
	return string(b)
}

// checkRequiredLabels checks if PR has all required labels.
func (ac *ApproveCondition) checkRequiredLabels(logger *zap.SugaredLogger, prLabels []github.Label) bool {
	defer logger.Sync()
	if ac.RequiredLabels == nil {
		logger.Debug("No required labels defined")
		// No required labels defined
		return true
	}
	pl := make(map[string]interface{})
	logger.Debugf("Checking if PR has all required labels: %v", ac.RequiredLabels)
	for _, l := range prLabels {
		pl[l.Name] = nil
	}
	for _, requiredLabel := range ac.RequiredLabels {
		if _, ok := pl[requiredLabel]; !ok {
			logger.Debugf("PR is missing required label: %s", requiredLabel)
			return false
		}
	}
	logger.Debug("All required labels are present")
	return true
}

// checkChangedFiles checks if PR changed only allowed files.
func (ac *ApproveCondition) checkChangedFiles(logger *zap.SugaredLogger, changes []github.PullRequestChange) bool {
	defer logger.Sync()
	logger.Debugf("Checking if PR changed only allowed files: %v", ac.ChangedFiles)
	for _, change := range changes {
		logger.Debugf("Checking file: %s", change.Filename)
		matched := slices.ContainsFunc(ac.ChangedFiles, func(allowedFile string) bool {
			filesMatcher := regexp.MustCompile(allowedFile)
			matched := filesMatcher.MatchString(change.Filename)
			logger.Debugf("File %s matched %s: %t", change.Filename, allowedFile, matched)
			return matched
		})
		if !matched {
			logger.Infof("File %s not matched", change.Filename)
			return false
		}
	}
	logger.Debug("All files matched")
	return true
}

// ApproveCondition defines Conditions for approving PR.
type ApproveCondition struct {
	RequiredLabels []string `yaml:"requiredLabels"`
	ChangedFiles   []string `yaml:"changedFiles"`
}

// ReadConfig reads config from config file.
// TODO: Rename function to reflect it's reading a file with conditions/rules only.
func (hb *HandlerBackend) ReadConfig() error {
	c := make(map[string]map[string]map[string]map[string][]ApproveCondition)
	configFile, err := os.ReadFile(hb.RulesPath)
	if err != nil {
		return err
	}
	yaml.Unmarshal(configFile, &c)
	hb.Conditions = c["conditions"]

	autoMergeConfig := make(map[string]map[string]map[string]map[string][]AutoMergeCondition)
	mergeConfigFile, err := os.ReadFile(hb.AutoMergeRulesPath)
	if err != nil {
		return err
	}
	yaml.Unmarshal(mergeConfigFile, &autoMergeConfig)
	hb.MergeConditions = autoMergeConfig["autoMergeConditions"]
	return nil
}

// checkPrApproveConditions checks if PR meets conditions for auto approving.
// It validates all ApproveConditions for owner/repo/PR author entity.
func (hb *HandlerBackend) checkPrApproveConditions(logger *zap.SugaredLogger, conditions []ApproveCondition, changes []github.PullRequestChange, prLabels []github.Label) bool {
	defer logger.Sync()
	for _, condition := range conditions {
		logger.Debugw("Checking condition", "condition", condition)
		labelsMatched := condition.checkRequiredLabels(logger, prLabels)
		if !labelsMatched {
			logger.Debug("Labels not matched")
			continue
		}
		filesMatched := condition.checkChangedFiles(logger, changes)
		if !filesMatched {
			logger.Debug("Files not matched")
			continue
		}
		return true
	}
	logger.Debug("No conditions matched")
	return false
}

// reviewPullRequest approves a pull request if it meets conditions.
// It searches conditions for owner/repo/PR author entity, validates them, waits for statuses to finish, validates their statuses, and approves PR.
func (hb *HandlerBackend) reviewPullRequest(ctx context.Context, logger *zap.SugaredLogger, prOrg, prRepo, prUser, prHeadSha string, prNumber int, prLabels []github.Label) {
	defer logger.Sync()
	defer hb.unlockPR(logger, prOrg, prRepo, prHeadSha, prNumber)
	logger.Debugf("Checking if conditions for PR author %s exists: %t", prUser, hb.Conditions[prOrg][prRepo][prUser] != nil)
	conditions, approveOk := hb.Conditions[prOrg][prRepo][prUser]
	if !approveOk {
		logger.Infof("Merge conditions for PR not found, skip automatic approval, pull request %s/%s#%d, author %s", prOrg, prRepo, prNumber, prUser)
		return
	}
	logger.Debugf("Checking if PR %d meets approval conditions: %v", prNumber, conditions)

	// Get changes from pull request.
	changes, err := hb.Ghc.GetPullRequestChanges(prOrg, prRepo, prNumber)
	if err != nil {
		logger.Errorw("failed get pull request changes", "error", err.Error())
		return
	}
	logger.Sync() // Syncing logger to make sure all logs from calling GitHub API are written before logs from functions called in next steps.
	conditionsMatched := hb.checkPrApproveConditions(logger, conditions, changes, prLabels)
	if !conditionsMatched {
		fmt.Errorf("pull request %s/%s#%d does not meet approval conditions", prOrg, prRepo, prNumber)
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
		err = hb.approvePullRequest(prHeadSha, prOrg, prRepo, prNumber, logger)
		if err != nil {
			logger.Errorf("failed approve pull request %s/%s#%d, got error: %s", prOrg, prRepo, prNumber, err)
		}
	}
}

func (hb *HandlerBackend) approvePullRequest(prHeadSha string, prOrg string, prRepo string, prNumber int, logger *zap.SugaredLogger) error {
	review := github.DraftReview{
		CommitSHA: prHeadSha,
		Body:      "",
		Action:    "APPROVE",
		Comments:  nil,
	}
	err := hb.Ghc.CreateReview(prOrg, prRepo, prNumber, review)
	if err != nil {
		return fmt.Errorf("failed create review for pull request %s/%s#%d sha: %s, got error: %s",
			prOrg,
			prRepo,
			prNumber,
			prHeadSha,
			err)
	}
	logger.Infof("Pull request %s/%s#%d was approved.", prOrg, prRepo, prNumber)
	err = hb.Ghc.AddLabel(prOrg, prRepo, prNumber, "auto-approved")
	if err != nil {
		return fmt.Errorf("failed add label to pull request %s/%s#%d, got error: %s",
			prOrg,
			prRepo,
			prNumber,
			err)
	}
	logger.Infof("Label auto-approved was added to pull request %s/%s#%d.", prOrg, prRepo, prNumber)
	return nil
}
