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
	"io/ioutil"
	"os"
	"path/filepath"

	gogithub "github.com/google/go-github/v42/github"
	"github.com/kyma-project/test-infra/development/gcp/pkg/pubsub"
	toolsclient "github.com/kyma-project/test-infra/development/github/pkg/client"
	"github.com/kyma-project/test-infra/development/github/pkg/client/v2"
	"github.com/kyma-project/test-infra/development/github/pkg/git"
	"github.com/kyma-project/test-infra/development/github/pkg/repoowners"
	consolelog "github.com/kyma-project/test-infra/development/logging"
	"github.com/kyma-project/test-infra/development/prow/externalplugin"
	toolstypes "github.com/kyma-project/test-infra/development/types"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/test-infra/prow/config"
	prowgit "k8s.io/test-infra/prow/git/v2"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pluginhelp"
	k8sowners "k8s.io/test-infra/prow/repoowners"
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
	pubsubTopic      string
	clondeRepos      ClonedRepos
)

// type RequiredUsers []string

// type RequiredLabels []string

// ClonedRepos is a map with information about already cloned repositories.
// A map keys represent hold org/repo and values are a path to a repository root.
type ClonedRepos map[string]string

type InstanceConfig struct {
	ToolsGithubTokenEnv string
	PubsubTopic         string
}

type AllOwners map[string]struct{}

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
func (o *InstanceConfig) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&o.ToolsGithubTokenEnv, "tools-github-token-env", "TOOLS_GITHUB_TOKEN", "Environment variable name with github token for tools repository.")
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
		if pr.Sender.Login == "dependabot[bot]" || pr.Sender.Login == "kyma-bot" {
			// if github.HasLabel("skip-review", pr.PullRequest.Labels) {
			//	return true
			// }
			return true
		}
	}
	return false
}

// TODO: move it to lib
func parseOwnersFiles(ownersFilePath, repoBase string, owners k8sowners.RepoOwner) (k8sowners.Config, error) {
	var (
		fullconfig   k8sowners.FullConfig
		simpleconfig k8sowners.SimpleConfig
		err          error
	)
	ownersFile := filepath.Join(repoBase, ownersFilePath, "OWNERS")
	simpleconfig, err = owners.ParseSimpleConfig(ownersFile)
	if err != nil {
		return k8sowners.Config{}, fmt.Errorf("failed parsing %s owners file to simpleconfig, error: %w", ownersFilePath, err)
	}
	if simpleconfig.Empty() {
		fullconfig, err = owners.ParseFullConfig(ownersFile)
		if err != nil {
			return k8sowners.Config{}, fmt.Errorf("failed parsing %s owners file to fullconfig, error: %w", ownersFilePath, err)
		}
		return fullconfig.Filters[".*"], nil
	}
	return simpleconfig.Config, nil
}

// TODO: move it to lib
func (a AllOwners) addOwners(approvers []string) {
	for _, approver := range approvers {
		a[approver] = struct{}{}
	}
}

// TODO: move it to lib
func getOwnersForChanges(changes []github.PullRequestChange, repoBase string, owners k8sowners.RepoOwner) (AllOwners, error) {
	var (
		conf k8sowners.Config
		err  error
	)
	allOwners := AllOwners{}
	for _, change := range changes {
		approversFile := owners.FindApproverOwnersForFile(change.Filename)
		reviewersFile := owners.FindReviewersOwnersForFile(change.Filename)
		conf, err = parseOwnersFiles(approversFile, repoBase, owners)
		if err != nil {
			return nil, err
		}
		allOwners.addOwners(conf.Approvers)
		if approversFile != reviewersFile {
			conf, err = parseOwnersFiles(reviewersFile, repoBase, owners)
			if err != nil {
				return nil, err
			}
		}
		allOwners.addOwners(conf.Reviewers)
	}
	return allOwners, nil
}

// TODO: move it to lib
func (a AllOwners) resolveSlackNames(aliases []toolstypes.Alias, users []toolstypes.User, repoAliases k8sowners.RepoAliases) (sets.String, error) {
	aliasesMap := make(map[string]toolstypes.Alias)
	usersMap := make(map[string]toolstypes.User)
	targets := sets.NewString()
	for _, alias := range aliases {
		aliasesMap[alias.ComGithubAliasname] = alias
	}
	for _, user := range users {
		usersMap[user.ComGithubUsername] = user
	}
	for owner := range a {
		if alias, ok := aliasesMap[owner]; ok {
			if !alias.SkipAutomergeNotifications {
				targets.Insert(alias.ComEnterpriseSlackChannelsnames...)
			}
		} else if user, ok := usersMap[owner]; ok {
			if !user.SkipAutomergeNotifications {
				targets.Insert(user.ComEnterpriseSlackUsername)
			}
		} else {
			userOwners := repoAliases.ExpandAlias(owner)
			if userOwners.Len() > 0 {
				for userOwner, _ := range userOwners {
					if user, ok := usersMap[userOwner]; ok {
						if !user.SkipAutomergeNotifications {
							targets.Insert(user.ComEnterpriseSlackUsername)
						}
					}
				}
			}
		}
	}
	return targets, nil
}

// TODO: move to lib
func getGitRepoClient(org, repo string) (prowgit.RepoClient, string, error) {
	if path, ok := clondeRepos[fmt.Sprintf("%s/%s", org, repo)]; ok {
		gitRepoClient, err := gitClientFactory.ClientFromDir(org, repo, path)
		if err != nil {
			return nil, "", fmt.Errorf("failed create git repository client from directory, org: %s, repo: %s, directory: %s, error: %w", org, repo, path, err)
		}
		err = gitRepoClient.Fetch()
		if err != nil {
			return nil, "", fmt.Errorf("failed fetch repostiory, org: %s, repo: %s, error: %w", org, repo, err)
		}
		return gitRepoClient, path, nil
	} else {
		gitRepoClient, err := gitClientFactory.ClientFor(org, repo)
		if err != nil {
			return nil, "", fmt.Errorf("failed create git repository client, org: %s, repo: %s, error: %w", org, repo, err)
		}
		clondeRepos[fmt.Sprintf("%s/%s", org, repo)] = gitRepoClient.Directory()
		return gitRepoClient, clondeRepos[fmt.Sprintf("%s/%s", org, repo)], nil
	}
}

// TODO: move it to lib
func loadOwnersAliases(basedir, filename string) (k8sowners.RepoAliases, error) {
	path := filepath.Join(basedir, filename)
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return k8sowners.ParseAliasesConfig(b)
}

func pullRequestEventHandler(server *externalplugin.Plugin, event externalplugin.Event) {
	logger := consolelog.NewLogger()
	defer logger.Sync()
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
			_, repoBase, err := getGitRepoClient(pr.Repo.Owner.Login, pr.Repo.Name)
			if err != nil {
				logger.Errorw("Failed get repository base directory", "error", err.Error())
			}
			repoAliases, err := loadOwnersAliases(repoBase, "OWNERS_ALIASES")
			if err != nil {
				logger.Errorw("failed load aliases file", "error", err.Error())
			}
			changes, err := githubClient.GetPullRequestChanges(pr.Repo.Owner.Login, pr.Repo.Name, pr.Number)
			if err != nil {
				logger.Errorw("failed get pull request changes", "error", err.Error())
			}
			allOwners, err := getOwnersForChanges(changes, repoBase, owners)
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
			targets, err := allOwners.resolveSlackNames(aliasesMap, usersMap, repoAliases)
			if err != nil {
				logger.Errorw("failed resolve owners slack names", "error", err.Error())
			}
			mergeMsgPayload.OwnersSlackIDs = targets.List()
			if len(mergeMsgPayload.OwnersSlackIDs) > 0 {
				attributes := map[string]string{"automerged": "true"}
				msgID, err := pubsubClient.PublishMessageWithAttributes(ctx, mergeMsgPayload, pubsubTopic, attributes)
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
	clondeRepos = ClonedRepos{}
	pluginInstanceOptions := InstanceConfig{}
	pluginOptions := externalplugin.Opts{}
	ownersOptions := repoowners.OwnersClientConfig{}
	pubsubOptions := pubsub.ClientConfig{}
	gitOptions := git.GitClientConfig{}
	fs := pluginOptions.NewFlags()
	ownersOptions.AddFlags(fs)
	pubsubOptions.AddFlags(fs)
	pluginInstanceOptions.AddFlags(fs)
	gitOptions.AddFlags(fs)
	pluginOptions.ParseFlags(fs)
	atom.SetLevel(pluginOptions.LogLevel)
	if pluginInstanceOptions.PubsubTopic == "" {
		err = fmt.Errorf("pubsub topic is empty")
		logger.Fatalf("%s", err.Error())
		panic(err)
	}

	pubsubTopic = pluginInstanceOptions.PubsubTopic

	toolsToken := os.Getenv(pluginInstanceOptions.ToolsGithubTokenEnv)
	if toolsToken == "" {
		err = fmt.Errorf("tools GitHub token is empty")
		logger.Fatalf("%s", err.Error())
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

	// gitclient, err := pluginOptions.Github.GitClient(pluginOptions.DryRun)
	// if err != nil {
	//	logger.Fatalw("Failed creating git client", "error", err.Error())
	//	panic(err)
	// }
	// gitClientFactory = git.ClientFactoryFrom(gitclient)
	gitClientFactory, err = gitOptions.NewGitClient(git.WithTokenPath(pluginOptions.Github.TokenPath), git.WithGithubClient(githubClient))
	if err != nil {
		logger.Fatalw("Failed creating git client", "error", err.Error())
		panic(err)
	}
	logger.Debug("git client ready")

	// gitClient, err = git.NewGitClient(git.WithGithubClient(githubClient))
	// if err != nil {
	//	logger.Fatalw("Failed creating git client", "error", err.Error())
	//	panic(err)
	// }

	repoOwnersClient, err = ownersOptions.NewRepoOwnersClient(
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
