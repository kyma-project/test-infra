package main

import (
	"encoding/json"

	consolelog "github.com/kyma-project/test-infra/development/logging"
	"github.com/kyma-project/test-infra/development/prow/externalplugin"
	"go.uber.org/zap/zapcore"
	"k8s.io/test-infra/prow/git/v2"
	"k8s.io/test-infra/prow/github"
)

type githubClient interface {
	CreatePullRequestReviewComment(org, repo string, number int, rc github.ReviewComment) error
	CreateReview(org, repo string, number int, r github.DraftReview) error
	GetPullRequestChanges(org, repo string, number int) ([]github.PullRequestChange, error)
	AddLabel(org, repo string, number int, label string) error
	RemoveLabel(org, repo string, number int, label string) error
	GetIssueLabels(org, repo string, number int) ([]github.Label, error)
	CreateComment(org, repo string, number int, comment string) error
	IsCollaborator(org, repo, user string) (bool, error)
	AssignIssue(org, repo string, number int, logins []string) error
	BotUserChecker() (func(string) bool, error)
}

type handlerBackend struct {
	ghc githubClient
	gcf git.ClientFactory
	logLevel zapcore.Level
	conditions map[string]map[string]map[string][]ApproveCondition
}

type ApproveCondition struct {
	ChangedFiles []string
}

// checkIfEventSupported check conditions PR must meet to send notification.
// At the time a conditions are hard coded. In future this will be taken from Tide queries.
func (h *handlerBackend) checkPrApproveConditions(prEvent github.PullRequestEvent) bool {
	for condition := range h.conditions[prEvent.Repo.Owner.Login][]
	if (pr.PullRequest.User.Login == "dependabot[bot]" || pr.PullRequest.User.Login == "kyma-bot") && github.HasLabel("skip-review", pr.PullRequest.Labels) {
		return true
	}
	return false
}

// pullRequestEventHandler process pull_request event webhooks received by plugin.
func (h *handlerBackend) pullRequestEventHandler(server *externalplugin.Plugin, event externalplugin.Event) {

			// Get git client for repository.
			_, repoBase, err := gitClientFactory.GetGitRepoClient(pr.Repo.Owner.Login, pr.Repo.Name)
			if err != nil {
				logger.Errorw("Failed get repository base directory", "error", err)
			}
			// Get changes from pull request.
			changes, err := githubClient.GetPullRequestChanges(pr.Repo.Owner.Login, pr.Repo.Name, pr.Number)
			if err != nil {
				logger.Errorw("failed get pull request changes", "error", err)
			}
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

	switch prEvent.Action {
	case github.PullRequestActionReviewRequested:
		logger = logger.With("pr-number", prEvent.Number)

}
