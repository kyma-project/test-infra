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
	githubClient     client.GithubClient
	gitClientFactory git.Client
	repoOwnersClient *repoowners.OwnersClient
	sapToolsClient   *toolsclient.SapToolsClient
	pubsubClient     *pubsub.Client
	automergeOptions AutomergeConfig
	pluginOptions    externalplugin.Opts
)

// AutomergeConfig holds configuration specific for plugin instance.
type AutomergeConfig struct {
	// PubSubTopic is a name of Google pubsub topic where automerge events will be published.
	PubsubTopic string
}

// AutoMergeMessagePayload is a pubsub message data field payload which will is send for automerge events.
type AutoMergeMessagePayload struct {
	PullRequestNumber *int     `json:"prNumber,omitempty"`
	PullRequestOrg    *string  `json:"prOrg,omitempty"`
	PullRequestRepo   *string  `json:"prRepo,omitempty"`
	PullRequestTitle  *string  `json:"prTitle,omitempty"`
	PullRequestAuthor *string  `json:"prAuthor,omitempty"`
	PullRequestURL    *string  `json:"prURL,omitempty"`
	OwnersSlackIDs    []string `json:"ownersSlackIDs,omitempty"`
}

// AddFlags add plugin instance specific flags to provided flag set.
// These flags are parsed along with flags defined for used clients.
func (o *AutomergeConfig) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&o.PubsubTopic, "pubsub-topic", "automerge", "PubSub topic to publish automerge event.")
}

// Plugin help description. This is published on Prow status plugin catalog.
func helpProvider(_ []config.OrgRepo) (*pluginhelp.PluginHelp, error) {
	ph := &pluginhelp.PluginHelp{
		Description: "The automerge-notification plugin send slack notification to owners about PRs merged without required review.",
	}
	return ph, nil
}

// checkIfEventSupported check conditions PR must meet to send notification.
// At the time a conditions are hard coded. In future this will be taken from Tide queries.
func checkIfEventSupported(pr github.PullRequestEvent) bool {
	if pr.PullRequest.Merged {
		if (pr.PullRequest.User.Login == "dependabot[bot]" || pr.PullRequest.User.Login == "kyma-bot") && github.HasLabel("skip-review", pr.PullRequest.Labels) {
			return true
		}
	}
	return false
}

// pullRequestEventHandler process pull_request event webhooks received by plugin.
func pullRequestEventHandler(_ *externalplugin.Plugin, event externalplugin.Event) {
	logger, atom := consolelog.NewLoggerWithLevel()
	defer logger.Sync()
	atom.SetLevel(pluginOptions.LogLevel)
	logger = logger.With(externalplugin.EventTypeField, event.EventType, github.EventGUID, event.EventGUID)

	var pr github.PullRequestEvent
	if err := json.Unmarshal(event.Payload, &pr); err != nil {
		logger.Errorw("Failed unmarshal json payload.", "error", err)
	}
	logger = logger.With("pr-number", pr.Number,
		"pr-sender", pr.Sender.Login)

	switch pr.Action {
	case github.PullRequestActionClosed:
		if checkIfEventSupported(pr) {
			// Set pubsub message payload values.
			mergeMsgPayload := AutoMergeMessagePayload{
				PullRequestOrg:    gogithub.String(pr.Repo.Owner.Login),
				PullRequestRepo:   gogithub.String(pr.Repo.Name),
				PullRequestTitle:  gogithub.String(pr.PullRequest.Title),
				PullRequestAuthor: gogithub.String(pr.Sender.Login),
				PullRequestNumber: gogithub.Int(pr.Number),
				PullRequestURL:    gogithub.String(pr.PullRequest.HTMLURL),
			}
			pr.GUID = event.EventGUID
			// Load repository owners.
			owners, err := repoOwnersClient.LoadRepoOwners(pr.Repo.Owner.Login, pr.Repo.Name, "main")
			if err != nil {
				logger.Errorw("Failed load RepoOwners", "error", err)
			}
			// Get git client for repository.
			_, repoBase, err := gitClientFactory.GetGitRepoClient(pr.Repo.Owner.Login, pr.Repo.Name)
			if err != nil {
				logger.Errorw("Failed get repository base directory", "error", err)
			}
			// Load repository owner aliases.
			repoAliases, err := repoOwnersClient.LoadRepoAliases(repoBase, "OWNERS_ALIASES")
			if err != nil {
				logger.Errorw("failed load aliases file", "error", err)
			}
			// Get changes from pull request.
			changes, err := githubClient.GetPullRequestChanges(pr.Repo.Owner.Login, pr.Repo.Name, pr.Number)
			if err != nil {
				logger.Errorw("failed get pull request changes", "error", err)
			}
			// Get owners for changes from pull request.
			allOwners, err := repoOwnersClient.GetOwnersForChanges(changes, repoBase, owners)
			if err != nil {
				logger.Errorw("filed get owners for changed files", "error", err)
			}
			ctx := context.Background()
			// Load aliases map file.
			aliasesMap, err := sapToolsClient.GetAliasesMap(ctx)
			if err != nil {
				logger.Errorw("failed get aliases map file", "error", err)
			}
			// Load users map file.
			usersMap, err := sapToolsClient.GetUsersMap(ctx)
			if err != nil {
				logger.Errorw("failed get users map file", "error", err)
			}
			// Get slack names to send notifications too.
			targets, err := repoOwnersClient.ResolveSlackNames(allOwners, aliasesMap, usersMap, repoAliases)
			if err != nil {
				logger.Errorw("failed resolve owners slack names", "error", err)
			}
			// Add slack names to notify to pubsub message payload.
			mergeMsgPayload.OwnersSlackIDs = targets.List()
			if len(mergeMsgPayload.OwnersSlackIDs) > 0 {
				// Set pubsub message attributes. This is used to filter messages in pubsub subscriptions.
				attributes := map[string]string{"automerged": "true"}
				// Publish pubsub message
				msgID, err := pubsubClient.PublishMessageWithAttributes(ctx, mergeMsgPayload, automergeOptions.PubsubTopic, attributes)
				if err != nil {
					logger.Errorw("failed pubslihed pubsub message", "error", err)
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

	// Initialize configuration options for clients.
	v1GithubOptions := toolsclient.GithubClientConfig{}
	automergeOptions = AutomergeConfig{}
	pluginOptions := externalplugin.Opts{}
	ownersOptions := repoowners.OwnersClientConfig{}
	pubsubOptions := pubsub.ClientConfig{}
	gitOptions := git.ClientConfig{}

	// Add client and plugin cli flags.
	fs := pluginOptions.NewFlags()
	v1GithubOptions.AddFlags(fs)
	ownersOptions.AddFlags(fs)
	pubsubOptions.AddFlags(fs)
	automergeOptions.AddFlags(fs)
	gitOptions.AddFlags(fs)
	pluginOptions.ParseFlags(fs)

	atom.SetLevel(pluginOptions.LogLevel)

	if automergeOptions.PubsubTopic == "" {
		err = fmt.Errorf("pubsub topic is empty")
		logger.Fatalf("%s", err.Error())
		panic(err)
	}

	// Get token for tools github.
	toolsToken, err := v1GithubOptions.GetToken()
	if err != nil {
		logger.Fatalw("Failed creating tools GitHub client", "error", err)
		panic(err)
	}

	// Create tools github client.
	sapToolsClient, err = toolsclient.NewSapToolsClient(context.Background(), toolsToken)
	if err != nil {
		logger.Fatalw("Failed creating tools GitHub client", "error", err)
		panic(err)
	}

	// Create github.com client.
	githubClient, err = pluginOptions.Github.NewGithubClient()
	if err != nil {
		logger.Fatalw("Failed creating GitHub client", "error", err)
		panic(err)
	}
	logger.Debug("github client ready")

	// Create git factory for github.com.
	gitClientFactory, err = gitOptions.NewClient(git.WithTokenPath(pluginOptions.Github.TokenPath), git.WithGithubClient(githubClient))
	if err != nil {
		logger.Fatalw("Failed creating git client", "error", err)
		panic(err)
	}
	logger.Debug("git client ready")

	// Create repository owners client.
	repoOwnersClient, err = ownersOptions.NewRepoOwnersClient(
		repoowners.WithLogger(logger),
		repoowners.WithGithubClient(&githubClient),
		repoowners.WithGitClient(gitClientFactory))
	if err != nil {
		logger.Fatalw("Failed creating repoOwners client", "error", err)
		panic(err)
	}

	// Create Google pubsub client.
	ctx := context.Background()
	withCredentialsFile := pubsubOptions.WithGoogleOption(option.WithCredentialsFile(pubsubOptions.CredentialsFilePath))
	pubsubClient, err = pubsubOptions.NewClient(ctx, withCredentialsFile)
	if err != nil {
		logger.Fatalf("An error occurred during pubsub client configuration: %v", err)
	}

	logger.Debug("ownersclient ready")

	// Create and start plugin instance.
	server := externalplugin.Plugin{}
	server.WithLogger(logger)
	server.Name = PluginName
	server.WithWebhookSecret(pluginOptions.WebhookSecretPath)
	server.RegisterWebhookHandler("pull_request", pullRequestEventHandler)
	externalplugin.Start(&server, helpProvider, &pluginOptions)
}
