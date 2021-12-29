package testuntrusted

import (
	"context"
	"encoding/json"

	"github.com/kyma-project/test-infra/development/prow"
	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pluginhelp"
)

const (
	PluginName = "test-untrusted"
)

type GithubClient interface {
	CreateComment(org, repo string, number int, comment string) error
	CreateCommentWithContext(ctx context.Context, org, repo string, number int, comment string) error
}

type testUntrusted struct {
	prow.Server
}

func DemuxEvent(eventType, eventGUID string, payload []byte) error {
	l := logrus.WithFields(
		logrus.Fields{
			prow.EventTypeField: eventType,
			github.EventGUID:    eventGUID,
		},
	)
	var srcRepo string
	switch eventType {
	case "pull_request":
		var pr github.PullRequestEvent
		if err := json.Unmarshal(payload, &pr); err != nil {
			return err
		}
		pr.GUID = eventGUID
		srcRepo = pr.Repo.FullName
	case "push":
		var pe github.PushEvent
		if err := json.Unmarshal(payload, &pe); err != nil {
			return err
		}
		pe.GUID = eventGUID
		srcRepo = pe.Repo.FullName
	default:
		l.Debug("Ignoring unhandled event type. (Might still be handled by external plugins.)")
	}
	return nil
}

func HelpProvider(_ []config.OrgRepo) (*pluginhelp.PluginHelp, error) {
	ph := &pluginhelp.PluginHelp{
		Description: "needs-tws checks if the Pull Request has modified Markdown files and blocks the merge until it is reviewed and approved by one of the Technical Writers.",
	}
	return ph, nil
}

func main() {
	cliOptions := prow.Opts{}
	fs := cliOptions.GatherDefaultOptions()
	cliOptions.Parse(fs)

	server := testUntrusted{}
	server.Name = PluginName
	server.WithTokenGenerator(cliOptions.WebhookSecretPath)
	server.WithGithubClient(cliOptions.Github, cliOptions.DryRun)
	server.DemuxEvent = DemuxEvent
	prow.Start(&server, HelpProvider, &cliOptions)
}
