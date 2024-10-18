package bumper

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/config/secret"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/robots/pr-creator/updater"
)

// Options is the options for autobumper operations.
type Options struct {
	// The target GitHub org name where the autobump PR will be created. Only required when SkipPullRequest is false.
	GitHubOrg string `json:"gitHubOrg" yaml:"gitHubOrg"`
	// The target GitHub repo name where the autobump PR will be created. Only required when SkipPullRequest is false.
	GitHubRepo string `json:"gitHubRepo" yaml:"gitHubRepo"`
	// The name of the branch in the target GitHub repo on which the autobump PR will be based.  If not specified, will be autodetected via GitHub API.
	GitHubBaseBranch string `json:"gitHubBaseBranch" yaml:"gitHubBaseBranch"`
	// The GitHub username to use. If not specified, uses values from the user associated with the access token.
	GitHubLogin string `json:"gitHubLogin" yaml:"gitHubLogin"`
	// The path to the GitHub token file. Only required when SkipPullRequest is false.
	GitHubToken string `json:"gitHubToken" yaml:"gitHubToken"`
	// The name to use on the git commit. Only required when GitEmail is specified and SkipPullRequest is false. If not specified, uses values from the user associated with the access token
	GitName string `json:"gitName" yaml:"gitName"`
	// The email to use on the git commit. Only required when GitName is specified and SkipPullRequest is false. If not specified, uses values from the user associated with the access token.
	GitEmail string `json:"gitEmail" yaml:"gitEmail"`
	// AssignTo specifies who to assign the created PR to. Takes precedence over onCallAddress and onCallGroup if set.
	AssignTo string `json:"assign_to" yaml:"assign_to"`
	// Whether to skip creating the pull request for this bump.
	SkipPullRequest bool `json:"skipPullRequest" yaml:"skipPullRequest"`
	// Whether to signoff the commits.
	Signoff bool `json:"signoff" yaml:"signoff"`
	// The name used in the address when creating remote. This should be the same name as the fork. If fork does not exist this will be the name of the fork that is created.
	// If it is not the same as the fork, the robot will change the name of the fork to this. Format will be git@github.com:{GitLogin}/{RemoteName}.git
	RemoteName string `json:"remoteName" yaml:"remoteName"`
	// The name of the branch that will be used when creating the pull request. If unset, defaults to "autobump".
	HeadBranchName string `json:"headBranchName" yaml:"headBranchName"`
	// Optional list of labels to add to the bump PR
	Labels []string `json:"labels" yaml:"labels"`
	// The GitHub host to use, defaulting to github.com
	GitHubHost string `json:"gitHubHost" yaml:"gitHubHost"`
	// The path to the git repository. If not specified, the current directory is used.
	RepoPath string `json:"repoPath" yaml:"repoPath"`
}

const (
	forkRemoteName = "bumper-fork-remote"

	defaultHeadBranchName = "autobump"
)

// PRHandler is the interface implemented by consumer of prcreator, for
// manipulating the repo, and provides commit messages, PR title and body.
type PRHandler interface {
	// Changes returns a slice of functions, each one does some stuff, and
	// returns commit message for the changes and list of files that should be added to commit
	Changes() []func(context.Context) (string, []string, error)
	// PRTitleBody returns the body of the PR, this function runs after all
	// changes have been executed
	PRTitleBody() (string, string, error)
}

func gitAdd(files []string, workTree *git.Worktree) error {
	for _, file := range files {
		// image-autobumper doesn't provide list of files to be added, so we need to add all files if -A is provided.
		if file == "-A" {
			if err := workTree.AddWithOptions(&git.AddOptions{All: true}); err != nil {
				return fmt.Errorf("add all files to the worktree: %w", err)
			}
			continue
		}
		if _, err := workTree.Add(file); err != nil {
			return fmt.Errorf("add file %s to the worktree: %w", file, err)
		}
	}

	return nil
}

func gitCommit(commitMsg string, committerName, committerEmail string, workTree *git.Worktree) error {
	commitOpts := &git.CommitOptions{
		Author: &object.Signature{
			Name:  committerName,
			Email: committerEmail,
			When:  time.Now(),
		},
	}
	if _, err := workTree.Commit(commitMsg, commitOpts); err != nil {
		return fmt.Errorf("commit changes to the remote branch: %w", err)
	}
	return nil
}

func gitPush(remote, remoteBranch, baseBranch string, repo *git.Repository, auth transport.AuthMethod, dryrun bool) error {
	// Add the remote if it doesn't exist with the remote url
	_, err := repo.Remote(forkRemoteName)
	if err != nil {
		_, err := repo.CreateRemote(&config.RemoteConfig{
			Name: forkRemoteName,
			URLs: []string{remote},
		})
		if err != nil {
			return fmt.Errorf("create remote: %w", err)
		}
	}

	// Get the remote head commit
	var remoteRefHash plumbing.Hash
	remoteRef, err := repo.Reference(plumbing.NewRemoteReferenceName(forkRemoteName, remoteBranch), true)
	// Ignore the error if the remote branch does not exist
	if err != nil && err != plumbing.ErrReferenceNotFound {
		return fmt.Errorf("get remote reference: %w", err)
	} else {
		remoteRefHash = remoteRef.Hash()
	}

	// Get the local head commit
	localRef, err := repo.Reference(plumbing.NewBranchReferenceName(baseBranch), true)
	if err != nil {
		return fmt.Errorf("get local reference: %w", err)
	}

	// Check if the remote head commit is the same as the local head commit
	if remoteRefHash == localRef.Hash() {
		logrus.Info("Remote is up to date, quitting.")
		return nil
	}

	if dryrun {
		logrus.Infof("[Dryrun] Pushing to %s refs/heads/%s", remote, remoteBranch)
		return nil
	}

	// Push the changes from local main to the remote.
	// We need to use the + sign to force push the changes.
	// We need refs/heads/* on the remote branch correctly push the branch using go-git.
	// See: https://github.com/go-git/go-git/issues/712#issuecomment-1467085888
	refSpecString := fmt.Sprintf("+%s:refs/heads/%s", localRef.Name(), remoteBranch)
	logrus.Infof("Pushing changes using %s refspec", refSpecString)
	err = repo.Push(&git.PushOptions{
		Force:      true,
		RemoteName: forkRemoteName,
		RefSpecs: []config.RefSpec{
			config.RefSpec(refSpecString),
		},
		Auth: auth,
	})
	if err != nil {
		return fmt.Errorf("push changes to the remote branch: %w", err)
	}

	return nil
}

// UpdatePullRequestWithLabels updates with github client "gc" the PR of github repo org/repo
// with "title" and "body" of PR matching author and headBranch from "source" to "baseBranch" with labels
func UpdatePullRequestWithLabels(gc github.Client, org, repo, title, body, source, baseBranch,
	headBranch string, allowMods bool, labels []string, dryrun bool) error {
	logrus.Info("Creating or updating PR...")
	if dryrun {
		logrus.Info("[Dryrun] ensure PR with:")
		logrus.Info(org, repo, title, body, source, baseBranch, headBranch, allowMods, gc, labels, dryrun)
		return nil
	}
	n, err := updater.EnsurePRWithLabels(org, repo, title, body, source, baseBranch, headBranch, allowMods, gc, labels)
	if err != nil {
		return fmt.Errorf("ensure PR exists: %w", err)
	}
	logrus.Infof("PR %s/%s#%d will merge %s into %s: %s", org, repo, *n, source, baseBranch, title)
	return nil
}

func generatePRBody(body, assignment string) string {
	return body + assignment + "\n"
}

func updatePRWithLabels(gc github.Client, org, repo string, extraLineInPRBody, login, baseBranch, headBranch string, allowMods bool, summary, body string, labels []string, dryrun bool) error {
	return UpdatePullRequestWithLabels(gc, org, repo, summary, generatePRBody(body, extraLineInPRBody), login+":"+headBranch, baseBranch, headBranch, allowMods, labels, dryrun)
}

func getAssignment(assignTo string) string {
	if assignTo != "" {
		return "/cc @" + assignTo
	}
	return ""
}

func validateOptions(o *Options) error {
	if !o.SkipPullRequest {
		if o.GitHubToken == "" {
			return fmt.Errorf("gitHubToken is mandatory when skipPullRequest is false or unspecified")
		}
		if (o.GitEmail == "") != (o.GitName == "") {
			return fmt.Errorf("gitName and gitEmail must be specified together")
		}
		if o.GitHubOrg == "" || o.GitHubRepo == "" {
			return fmt.Errorf("gitHubOrg and gitHubRepo are mandatory when skipPullRequest is false or unspecified")
		}
		if o.RemoteName == "" {
			return fmt.Errorf("remoteName is mandatory when skipPullRequest is false or unspecified")
		}
	}
	if !o.SkipPullRequest {
		if o.HeadBranchName == "" {
			o.HeadBranchName = defaultHeadBranchName
		}
	}

	if o.RepoPath == "" {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get current working directory: %w", err)
		}

		o.RepoPath = wd
	}

	return nil
}

// Run is the entrypoint which will update files based on the provided options and PRHandler.
func Run(_ context.Context, o *Options, prh PRHandler) error {
	if err := validateOptions(o); err != nil {
		return fmt.Errorf("validating options: %w", err)
	}

	if o.SkipPullRequest {
		logrus.Debugf("--skip-pull-request is set to true, won't create a pull request.")
	}

	return processGitHub(o, prh)
}

func processGitHub(o *Options, prh PRHandler) error {
	if err := secret.Add(o.GitHubToken); err != nil {
		return fmt.Errorf("start secrets agent: %w", err)
	}

	githubHost := "github.com"
	gitHubGraphQLEndpoint := github.DefaultGraphQLEndpoint
	gitHubAPIEndpoint := github.DefaultAPIEndpoint
	if o.GitHubHost != "" {
		// Override the default endpoints if a custom GitHub host is provided.
		// Use default endpoint for GitHub Enterprise Server.
		// See: https://docs.github.com/en/enterprise-server@3.12/rest/quickstart?apiVersion=2022-11-28#using-curl-in-the-command-line
		// See: https://docs.github.com/en/enterprise-server@3.12/graphql/overview/about-the-graphql-api
		githubHost = o.GitHubHost
		gitHubAPIEndpoint = fmt.Sprintf("https://%s/api/v3", o.GitHubHost)
		gitHubGraphQLEndpoint = fmt.Sprintf("https://%s/api/graphql", o.GitHubHost)
	}

	gc, err := github.NewClient(secret.GetTokenGenerator(o.GitHubToken), secret.Censor, gitHubGraphQLEndpoint, gitHubAPIEndpoint)
	if err != nil {
		return err
	}
	if o.GitHubLogin == "" || o.GitName == "" || o.GitEmail == "" {
		user, err := gc.BotUser()
		if err != nil {
			return fmt.Errorf("get the user data for the provided GH token: %w", err)
		}
		if o.GitHubLogin == "" {
			o.GitHubLogin = user.Login
		}
		if o.GitName == "" {
			o.GitName = user.Name
		}
		if o.GitEmail == "" {
			o.GitEmail = user.Email
		}
	}

	if o.GitHubBaseBranch == "" {
		repo, err := gc.GetRepo(o.GitHubOrg, o.GitHubRepo)
		if err != nil {
			return fmt.Errorf("detect default remote branch for %s/%s: %w", o.GitHubOrg, o.GitHubRepo, err)
		}
		o.GitHubBaseBranch = repo.DefaultBranch
	}

	gitRepo, err := git.PlainOpen(o.RepoPath)
	if err != nil {
		return fmt.Errorf("open git repo: %w", err)
	}
	workTree, err := gitRepo.Worktree()
	if err != nil {
		return fmt.Errorf("get worktree: %w", err)
	}

	for i, changeFunc := range prh.Changes() {
		commitMsg, filesToBeAdded, err := changeFunc(context.Background())
		if err != nil {
			return fmt.Errorf("failed to process function %d: %s", i, err)
		}

		logrus.Info("Checking status of the worktree")
		status, err := workTree.Status()
		if err != nil {
			return fmt.Errorf("get worktree status: %w", err)
		}
		if status.IsClean() {
			logrus.Info("No changes, quitting.")
			return nil
		}

		logrus.Info("Adding changes to the stage")
		if err := gitAdd(filesToBeAdded, workTree); err != nil {
			return fmt.Errorf("add files to the worktree: %w", err)
		}

		logrus.Infof("Committing changes with message: %s", commitMsg)
		if err := gitCommit(commitMsg, o.GitName, o.GitEmail, workTree); err != nil {
			return fmt.Errorf("commit changes to the worktree: %w", err)
		}
	}

	remote := fmt.Sprintf("https://%s/%s/%s.git", githubHost, o.GitHubLogin, o.RemoteName)
	authMethod := &http.BasicAuth{
		Username: o.GitHubLogin,
		Password: string(secret.GetTokenGenerator(o.GitHubToken)()),
	}
	if err := gitPush(remote, o.HeadBranchName, o.GitHubBaseBranch, gitRepo, authMethod, o.SkipPullRequest); err != nil {
		return fmt.Errorf("push changes to the remote branch: %w", err)
	}

	summary, body, err := prh.PRTitleBody()
	if err != nil {
		return fmt.Errorf("creating PR summary and body: %w", err)
	}
	if err := updatePRWithLabels(gc, o.GitHubOrg, o.GitHubRepo, getAssignment(o.AssignTo), o.GitHubLogin, o.GitHubBaseBranch, o.HeadBranchName, updater.PreventMods, summary, body, o.Labels, o.SkipPullRequest); err != nil {
		return fmt.Errorf("to create the PR: %w", err)
	}
	return nil
}
