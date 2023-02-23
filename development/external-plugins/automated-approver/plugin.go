package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/cenkalti/backoff/v4"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"

	consolelog "github.com/kyma-project/test-infra/development/logging"
	"github.com/kyma-project/test-infra/development/prow/externalplugin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/test-infra/prow/github"
)

type githubClient interface {
	CreatePullRequestReviewComment(org, repo string, number int, rc github.ReviewComment) error
	CreateReview(org, repo string, number int, r github.DraftReview) error
	GetPullRequestChanges(org, repo string, number int) ([]github.PullRequestChange, error)
	GetCombinedStatus(org, repo, ref string) (*github.CombinedStatus, error)
	AddLabel(org, repo string, number int, label string) error
	CreateComment(org, repo string, number int, comment string) error
}

type handlerBackend struct {
	ghc                    githubClient
	logLevel               zapcore.Level
	waitForStatusesTimeout int                                                 // in seconds
	configPath             string                                              // Path to yaml config file
	conditions             map[string]map[string]map[string][]ApproveCondition `yaml:"conditions"`
}

type ApproveCondition struct {
	RequiredLabels []string `yaml:"requiredLabels"`
	ChangedFiles   []string `yaml:"changedFiles"`
}

func (ac *ApproveCondition) checkRequiredLabels(prLabels []github.Label) bool {
	if ac.RequiredLabels == nil {
		// No required labels defined
		return true
	}
	pl := make(map[string]interface{})
	for _, l := range prLabels {
		pl[l.Name] = nil
	}
	for _, requiredLabel := range ac.RequiredLabels {
		if _, ok := pl[requiredLabel]; !ok {
			return false
		}
	}
	return true
}

func (ac *ApproveCondition) checkChangedFiles(changes []github.PullRequestChange) bool {
	for _, change := range changes {
		change := change
		matched := slices.ContainsFunc(ac.ChangedFiles, func(allowedFile string) bool {
			filesMatcher := regexp.MustCompile(allowedFile)
			matched := filesMatcher.MatchString(change.Filename)
			return matched
		})
		if !matched {
			return false
		}
	}
	return true
}

// readConfig reads config from config file.
func (h *handlerBackend) readConfig() error {
	h.conditions = make(map[string]map[string]map[string][]ApproveCondition)
	configFile, err := os.ReadFile(h.configPath)
	if err == nil {
		return yaml.Unmarshal(configFile, h)
	}
	return err
}

// checkPrStatuses checks if all statuses are in success state.
// Tide required status check is not taken into account. It will be always pending until PR is ready to merge.
// Timeout limits time waiting for statuses became success.
func (h *handlerBackend) checkPrStatuses(prStatuses []github.Status) error {
	backOff := backoff.NewExponentialBackOff()
	backOff.MaxElapsedTime = time.Duration(h.waitForStatusesTimeout) * time.Second
	backOff.MaxInterval = 15 * time.Minute
	backOff.InitialInterval = 5 * time.Minute
	err := backoff.Retry(func() error {
		for _, prStatus := range prStatuses {
			if prStatus.State == "failure" {
				return backoff.Permanent(fmt.Errorf("pull request status check %s failed", prStatus.Context))
			} else if prStatus.State == "pending" && prStatus.Context != "tide" {
				return fmt.Errorf("pull request status check %s is pending", prStatus.Context)
			}
		}
		return nil
	}, backOff)
	return err
}

// checkIfEventSupported check conditions PR must meet to send notification.
// At the time a conditions are hard coded. In future this will be taken from Tide queries.
func (h *handlerBackend) checkPrApproveConditions(conditions []ApproveCondition, changes []github.PullRequestChange, prLabels []github.Label) bool {
	for _, condition := range conditions {
		labelsMatched := condition.checkRequiredLabels(prLabels)
		if !labelsMatched {
			continue
		}
		filesMatched := condition.checkChangedFiles(changes)
		if !filesMatched {
			continue
		}
		return true
	}
	return false
}

func (h *handlerBackend) handleReviewRequestedAction(logger *zap.SugaredLogger, prEvent github.PullRequestEvent) {
	if conditions, ok := h.conditions[prEvent.Repo.Owner.Name][prEvent.Repo.Name][prEvent.Sender.Login]; ok {
		// Get changes from pull request.
		changes, err := h.ghc.GetPullRequestChanges(prEvent.Repo.Owner.Name, prEvent.Repo.Name, prEvent.Number)
		if err != nil {
			logger.Errorw("failed get pull request changes", "error", err.Error())
		}
		conditionsMatched := h.checkPrApproveConditions(conditions, changes, prEvent.PullRequest.Labels)
		if !conditionsMatched {
			return
		}
		prStatuses, err := h.ghc.GetCombinedStatus(prEvent.Repo.Owner.Name, prEvent.Repo.Name, prEvent.PullRequest.Head.SHA)
		if err != nil {
			logger.Errorw("failed get pull request contexts combined status", "error", err.Error())
		}
		// Don't check if pr checks status is success as that means all context are success, even tide context.
		// That means a pr was already approved and is ready for merge, because tide context transition to success
		// when pr is ready for merge.
		switch prState := prStatuses.State; prState {
		case "failure":
			logger.Infof("Pull request %d is in failure state, skip approving.", prEvent.Number)
			return
		case "pending":
			err = h.checkPrStatuses(prStatuses.Statuses)
			logger.Sync()
			if err != nil {
				logger.Errorf("pull request %s/%s#%d has non success statuses, got error: %s",
					prEvent.Repo.Owner.Name,
					prEvent.Repo.Name,
					prEvent.Number,
					err)
				return
			}
			review := github.DraftReview{
				CommitSHA: prEvent.PullRequest.Head.SHA,
				Body:      "",
				Action:    "APPROVE",
				Comments:  nil,
			}
			err := h.ghc.CreateReview(prEvent.Repo.Owner.Name, prEvent.Repo.Name, prEvent.Number, review)
			if err != nil {
				logger.Errorf("failed create review for pull request %s/%s#%d sha: %s, got error: %s",
					prEvent.Repo.Owner.Name,
					prEvent.Repo.Name,
					prEvent.Number,
					prEvent.PullRequest.Head.SHA,
					err)
				return
			}
			err = h.ghc.AddLabel(prEvent.Repo.Owner.Name, prEvent.Repo.Name, prEvent.Number, "auto-approved")
			if err != nil {
				logger.Errorf("failed add label to pull request %s/%s#%d, got error: %s",
					prEvent.Repo.Owner.Name,
					prEvent.Repo.Name,
					prEvent.Number,
					err)
			}
		}
	}
	logger.Infof("Pull request %s/%s#%d doesn't meet conditions to be auto approved.",
		prEvent.Repo.Owner.Name,
		prEvent.Repo.Name,
		prEvent.Number)
}

func (h *handlerBackend) pullRequestEventHandler(_ *externalplugin.Plugin, payload externalplugin.Event) {

	logger, atom := consolelog.NewLoggerWithLevel()
	defer logger.Sync()
	atom.SetLevel(h.logLevel)
	logger = logger.With(externalplugin.EventTypeField, payload.EventType, github.EventGUID, payload.EventGUID)

	logger.Debug("Got pull_request payload")
	var prEvent github.PullRequestEvent
	if err := json.Unmarshal(payload.Payload, &prEvent); err != nil {
		logger.Errorw("Failed unmarshal json payload.", "error", err)
		return
	}

	logger.Sync()
	if prEvent.Action == github.PullRequestActionReviewRequested {
		logger = logger.With("pr-number", prEvent.Number)
		h.handleReviewRequestedAction(logger, prEvent)
	}
}
