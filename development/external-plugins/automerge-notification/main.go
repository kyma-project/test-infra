// Tools GitHub token is an environment variable.
//        - --dry-run=false
//        - --github-endpoint=http://ghproxy
//        - --github-endpoint=https://api.github.com
//        - --github-token-path=/etc/github/oauth
//        - --hmac-secret-file=/etc/webhook/hmac
//        - --config-path=/etc/config/config.yaml
//        - --job-config-path=/etc/job-config

package main

import (
	"encoding/json"
	"flag"
	"fmt"

	gogithub "github.com/google/go-github/v42/github"
	"github.com/kyma-project/test-infra/development/gcp/pkg/pubsub"
	toolsclient "github.com/kyma-project/test-infra/development/github/pkg/client"
	"github.com/kyma-project/test-infra/development/github/pkg/client/v2"
	"github.com/kyma-project/test-infra/development/github/pkg/git"
	"github.com/kyma-project/test-infra/development/github/pkg/repoowners"
	consolelog "github.com/kyma-project/test-infra/development/logging"
	"github.com/kyma-project/test-infra/development/prow/externalplugin"
	"golang.org/x/net/context"
	"google.golang.org/api/option"

	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pluginhelp"
)

const (
	PluginName = "automerge-notification"
)

var (
	githubClient     *client.GithubClient
	gitClientFactory *git.GitClient
	repoOwnersClient *repoowners.OwnersClient
	sapToolsClient   *toolsclient.SapToolsClient
	pubsubClient     *pubsub.Client
	automergeOptions AutomergeConfig
	pluginOptions    externalplugin.Opts
)

type AutomergeConfig struct {
	PubsubTopic string
}

type AutoMergeMessagePayload struct {
	PullRequestNumber *int     `json:"prNumber,omitempty"`
	PullRequestOrg    *string  `json:"prOrg,omitempty"`
	PullRequestRepo   *string  `json:"prRepo,omitempty"`
	PullRequestTitle  *string  `json:"prTitle,omitempty"`
	PullRequestAuthor *string  `json:"prAuthor,omitempty"`
	PullRequestURL    *string  `json:"prURL,omitempty"`
	OwnersSlackIDs    []string `json:"ownersSlackIDs,omitempty"`
}

// TODO: read required users and required labels for automerging from tide configuration
func (o *AutomergeConfig) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&o.PubsubTopic, "pubsub-topic", "automerge", "PubSub topic to publish automerge event.")
}

func helpProvider(_ []config.OrgRepo) (*pluginhelp.PluginHelp, error) {
	ph := &pluginhelp.PluginHelp{
		Description: "XXXXXXXXXXXXXX.",
	}
	return ph, nil
}

func checkIfEventSupported(pr github.PullRequestEvent) bool {
	if pr.PullRequest.Merged {
		if pr.PullRequest.User.Login == "dependabot[bot]" || pr.PullRequest.User.Login == "kyma-bot" {
			// if github.HasLabel("skip-review", pr.PullRequest.Labels) {
			//	return true
			// }
			return true
		}
	}
	return false
}

func pullRequestEventHandler(server *externalplugin.Plugin, event externalplugin.Event) {
	logger, atom := consolelog.NewLoggerWithLevel()
	defer logger.Sync()
	atom.SetLevel(pluginOptions.LogLevel)
	logger = logger.With(externalplugin.EventTypeField, event.EventType, github.EventGUID, event.EventGUID)
	var pr github.PullRequestEvent
	if err := json.Unmarshal(event.Payload, &pr); err != nil {
		logger.Errorw("Failed unmarshal json payload.", "error", err.Error())
	}
	logger = logger.With("pr-number", pr.Number,
		"pr-sender", pr.Sender.Login)
	switch pr.Action {
	case github.PullRequestActionClosed:
		if checkIfEventSupported(pr) {
			mergeMsgPayload := AutoMergeMessagePayload{
				PullRequestOrg:    gogithub.String(pr.Repo.Owner.Login),
				PullRequestRepo:   gogithub.String(pr.Repo.Name),
				PullRequestTitle:  gogithub.String(pr.PullRequest.Title),
				PullRequestAuthor: gogithub.String(pr.Sender.Login),
				PullRequestNumber: gogithub.Int(pr.Number),
				PullRequestURL:    gogithub.String(pr.PullRequest.HTMLURL),
			}
			pr.GUID = event.EventGUID
			owners, err := repoOwnersClient.LoadRepoOwners(pr.Repo.Owner.Login, pr.Repo.Name, "main")
			if err != nil {
				logger.Errorw("Failed load RepoOwners", "error", err.Error())
			}
			_, repoBase, err := gitClientFactory.GetGitRepoClient(pr.Repo.Owner.Login, pr.Repo.Name)
			if err != nil {
				logger.Errorw("Failed get repository base directory", "error", err.Error())
			}
			repoAliases, err := repoOwnersClient.LoadRepoAliases(repoBase, "OWNERS_ALIASES")
			if err != nil {
				logger.Errorw("failed load aliases file", "error", err.Error())
			}
			changes, err := githubClient.GetPullRequestChanges(pr.Repo.Owner.Login, pr.Repo.Name, pr.Number)
			if err != nil {
				logger.Errorw("failed get pull request changes", "error", err.Error())
			}
			allOwners, err := repoOwnersClient.GetOwnersForChanges(changes, repoBase, owners)
			if err != nil {
				logger.Errorw("filed get owners for changed files", "error", err.Error())
			}
			ctx := context.Background()
			aliasesMap, err := sapToolsClient.GetAliasesMap(ctx)
			if err != nil {
				logger.Errorw("failed get aliases map file", "error", err.Error())
			}
			usersMap, err := sapToolsClient.GetUsersMap(ctx)
			if err != nil {
				logger.Errorw("failed get users map file", "error", err.Error())
			}
			targets, err := repoOwnersClient.ResolveSlackNames(allOwners, aliasesMap, usersMap, repoAliases)
			if err != nil {
				logger.Errorw("failed resolve owners slack names", "error", err.Error())
			}
			mergeMsgPayload.OwnersSlackIDs = targets.List()
			if len(mergeMsgPayload.OwnersSlackIDs) > 0 {
				attributes := map[string]string{"automerged": "true"}
				msgID, err := pubsubClient.PublishMessageWithAttributes(ctx, mergeMsgPayload, automergeOptions.PubsubTopic, attributes)
				if err != nil {
					logger.Errorw("failed pubslihed pubsub message", "error", err.Error())
				}
				logger.Infof("PubSub message published, message ID: %s", *msgID)
			} else {
				logger.Info("No users or channels for notification found.")
			}
		}
	}
}

func main() {
	var err error

	logger, atom := consolelog.NewLoggerWithLevel()
	defer logger.Sync()

	v1GithubOptions := toolsclient.GithubClientConfig{}
	automergeOptions = AutomergeConfig{}
	pluginOptions := externalplugin.Opts{}
	ownersOptions := repoowners.OwnersClientConfig{}
	pubsubOptions := pubsub.ClientConfig{}
	gitOptions := git.GitClientConfig{}

	fs := pluginOptions.NewFlags()
	v1GithubOptions.AddFlags(fs)
	ownersOptions.AddFlags(fs)
	pubsubOptions.AddFlags(fs)
	automergeOptions.AddFlags(fs)
	gitOptions.AddFlags(fs)
	pluginOptions.ParseFlags(fs)

	atom.SetLevel(pluginOptions.LogLevel)
	// logLevel = pluginOptions.LogLevel
	if automergeOptions.PubsubTopic == "" {
		err = fmt.Errorf("pubsub topic is empty")
		logger.Fatalf("%s", err.Error())
		panic(err)
	}

	toolsToken, err := v1GithubOptions.GetToken()
	if err != nil {
		logger.Fatalw("Failed creating tools GitHub client", "error", err.Error())
		panic(err)
	}
	sapToolsClient, err = toolsclient.NewSapToolsClient(context.Background(), toolsToken)
	if err != nil {
		logger.Fatalw("Failed creating tools GitHub client", "error", err.Error())
		panic(err)
	}

	githubClient, err = pluginOptions.Github.NewGithubClient()
	if err != nil {
		logger.Fatalw("Failed creating GitHub client", "error", err.Error())
		panic(err)
	}
	logger.Debug("github client ready")

	gitClientFactory, err = gitOptions.NewGitClient(git.WithTokenPath(pluginOptions.Github.TokenPath), git.WithGithubClient(githubClient))
	if err != nil {
		logger.Fatalw("Failed creating git client", "error", err.Error())
		panic(err)
	}
	logger.Debug("git client ready")

	repoOwnersClient, err = ownersOptions.NewRepoOwnersClient(
		repoowners.WithLogger(logger),
		repoowners.WithGithubClient(githubClient),
		repoowners.WithGitClient(gitClientFactory))
	if err != nil {
		logger.Fatalw("Failed creating repoOwners client", "error", err.Error())
		panic(err)
	}
	ctx := context.Background()
	withCredentialsFile := pubsubOptions.WithGoogleOption(option.WithCredentialsFile(pubsubOptions.CredentialsFilePath))
	pubsubClient, err = pubsubOptions.NewClient(ctx, withCredentialsFile)

	logger.Debug("ownersclient ready")

	server := externalplugin.Plugin{}
	server.WithLogger(logger)
	server.Name = PluginName
	server.WithWebhookSecret(pluginOptions.WebhookSecretPath)
	server.RegisterWebhookHandler("pull_request", pullRequestEventHandler)
	externalplugin.Start(&server, helpProvider, &pluginOptions)
}
