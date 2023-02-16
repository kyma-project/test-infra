package main

import (
	"encoding/json"
	"regexp"

	"golang.org/x/exp/slices"

	consolelog "github.com/kyma-project/test-infra/development/logging"
	"github.com/kyma-project/test-infra/development/prow/externalplugin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/test-infra/prow/git/v2"
	"k8s.io/test-infra/prow/github"
)

type githubClient interface {
	CreatePullRequestReviewComment(org, repo string, number int, rc github.ReviewComment) error
	CreateReview(org, repo string, number int, r github.DraftReview) error
	GetPullRequestChanges(org, repo string, number int) ([]github.PullRequestChange, error)
	GetCombinedStatus(org, repo, ref string) (*github.CombinedStatus, error)
	AddLabel(org, repo string, number int, label string) error
	CreateComment(org, repo string, number int, comment string) error
	IsCollaborator(org, repo, user string) (bool, error)
	BotUserChecker() (func(string) bool, error)
}

type handlerBackend struct {
	ghc githubClient
	gcf git.ClientFactory
	logLevel zapcore.Level
	conditions map[string]map[string]map[string][]ApproveCondition
}

type ApproveCondition struct {
	RequiredLabels []string
	ChangedFiles []string
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


func (h *handlerBackend) handleReviewRequestedAction (logger *zap.SugaredLogger, prEvent github.PullRequestEvent) {
	if conditions, ok := h.conditions[prEvent.Repo.Owner.Name][prEvent.Repo.Name][prEvent.Sender.Login]; ok{
		// Get changes from pull request.
		changes, err := h.ghc.GetPullRequestChanges(prEvent.Repo.Owner.Name, prEvent.Repo.Name, prEvent.Number)
		if err != nil{
			logger.Errorw("failed get pull request changes", "error", err)
		}
		conditionsMatched := h.checkPrApproveConditions(conditions, changes, prEvent.PullRequest.Labels)
		if !conditionsMatched {
			return
		}
		prStatuses, err := h.ghc.GetCombinedStatus(prEvent.Repo.Owner.Name, prEvent.Repo.Name, prEvent.PullRequest.Head.SHA)
		if err != nil {
			Log error
		}
		// Don't check if pr checks status is success as that means all context are success, even tide context.
		// That means a pr was already approved and is ready for merge, because tide context transition to success
		// when pr is ready for merge.
		switch prState := prStatuses.State; prState {
		case "failure":
			logger.Infof("Pull request %d is in failure state, skip approving.", prEvent.Number)
			return
		case "pending":
			for _, prStatus := range prStatuses.Statuses {
				if prStatus.State != "failure"{
					Do a comment on pr
					return
				}
			}
			Approve pr.
		}
	}
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

	if prEvent.Action == github.PullRequestActionReviewRequested {
		logger = logger.With("pr-number", prEvent.Number)
		h.handleReviewRequestedAction(logger, prEvent)
	}
}
