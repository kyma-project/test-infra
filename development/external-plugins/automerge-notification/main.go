package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/kyma-project/test-infra/development/gcp/pkg/pubsub"
	toolsclient "github.com/kyma-project/test-infra/development/github/pkg/client"
	"github.com/kyma-project/test-infra/development/github/pkg/client/v2"
	"github.com/kyma-project/test-infra/development/github/pkg/repoowners"
	consolelog "github.com/kyma-project/test-infra/development/logging"
	"github.com/kyma-project/test-infra/development/prow/externalplugin"
	toolstypes "github.com/kyma-project/test-infra/development/types"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/git/v2"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pluginhelp"
	k8sowners "k8s.io/test-infra/prow/repoowners"
)

const (
	PluginName = "automerge-notification"
)

var (
	githubClient     *client.GithubClient
	gitClientFactory git.ClientFactory
	repoOwnersClient *repoowners.OwnersClient
	sapToolsClient   *toolsclient.SapToolsClient
	pubsubClient     *pubsub.Client
)

type AllOwners map[string]struct{}

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
// func mergeOwnersConfigs(ownersConfig ...k8sowners.Config) k8sowners.Config {
//	conf := k8sowners.Config{}
//	for _, config := range ownersConfig {
//		conf.Reviewers = append(conf.Reviewers, config.Reviewers...)
//		conf.Approvers = append(conf.Approvers, config.Approvers...)
//		conf.Labels = append(conf.Labels, config.Labels...)
//		conf.RequiredReviewers = append(conf.RequiredReviewers, config.RequiredReviewers...)
//	}
//	return conf
// }

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
	for owner, _ := range a {
		if alias, ok := aliasesMap[owner]; ok {
			targets.Insert(alias.ComEnterpriseSlackChannelsnames...)
		} else if user, ok := usersMap[owner]; ok {
			targets.Insert(user.ComEnterpriseSlackUsername)
		} else {
			t := repoAliases.ExpandAlias(owner)
			if t.Len() > 0 {
				targets = targets.Union(t)
			}
		}
	}
	return targets, nil
}

// TODO: write reusing cloned repo with fetch
// func getGitRepoClient() {
//
// }

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
			pr.GUID = event.EventGUID
			owners, err := repoOwnersClient.LoadRepoOwners(pr.Repo.Owner.Login, pr.Repo.Name, "main")
			if err != nil {
				logger.Errorw("Failed load RepoOwners", "error", err.Error())
			}
			// TODO: add support for creating repo client for existing directory and fetch changes
			gitRepoClient, err := gitClientFactory.ClientFor(pr.Repo.Owner.Login, pr.Repo.Name)
			if err != nil {
				logger.Errorw("Failed create git repository client", "error", err.Error())
			}
			repoBase := gitRepoClient.Directory()
			repoAliases, err := loadOwnersAliases(repoBase, "OWNERS_ALIASES")
			if err != nil {
				XXX
			}
			changes, err := githubClient.GetPullRequestChanges(pr.Repo.Owner.Login, pr.Repo.Name, pr.Number)
			if err != nil {
				XXX
			}
			allOwners, err := getOwnersForChanges(changes, repoBase, owners)
			if err != nil {
				XXX
			}

			ctx := context.Background()
			aliasesMap, err := sapToolsClient.GetAliasesMap(ctx)
			if err != nil {
				XXX
			}
			usersMap, err := sapToolsClient.GetUsersMap(ctx)
			if err != nil {
				XXXX
			}

			targets, err := allOwners.resolveSlackNames(aliasesMap, usersMap, repoAliases)
			if err != nil {
				XXXX
			}

			pubsubClient.PublishMessage()
		}
	}
}

func main() {
	var err error
	logger := consolelog.NewLogger()
	defer logger.Sync()
	pluginOptions := externalplugin.Opts{}
	ownersOptions := repoowners.OwnersClientConfig{}
	pubsubOptions := pubsub.ClientConfig{}
	fs := pluginOptions.NewFlags()
	ownersOptions.AddFlags(fs)
	pubsubOptions.AddFlags(fs)
	pluginOptions.ParseFlags(fs)

	// TODO: get tokoen from k8s secret
	token := "token"
	// TODO: get topic from config or flags
	topic
	sapToolsClient, err = toolsclient.NewSapToolsClient(context.Background(), token)

	githubClient, err = pluginOptions.Github.NewGithubClient()
	if err != nil {
		logger.Fatalw("Failed creating GitHub client", "error", err.Error())
		panic(err)
	}
	logger.Debug("github client ready")

	gitclient, err := pluginOptions.Github.GitClient(pluginOptions.DryRun)
	if err != nil {
		logger.Fatalw("Failed creating git client", "error", err.Error())
		panic(err)
	}
	gitClientFactory = git.ClientFactoryFrom(gitclient)
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
