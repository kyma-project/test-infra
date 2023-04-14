package bumper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/config/secret"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/robots/pr-creator/updater"
)

// Options is the options for autobumper operations.
type Options struct {
	// The target GitHub org name where the autobump PR will be created. Only required when SkipPullRequest is false.
	GitHubOrg string `json:"gitHubOrg"`
	// The target GitHub repo name where the autobump PR will be created. Only required when SkipPullRequest is false.
	GitHubRepo string `json:"gitHubRepo"`
	// The name of the branch in the target GitHub repo on which the autobump PR will be based.  If not specified, will be autodetected via GitHub API.
	GitHubBaseBranch string `json:"gitHubBaseBranch"`
	// The GitHub username to use. If not specified, uses values from the user associated with the access token.
	GitHubLogin string `json:"gitHubLogin"`
	// The path to the GitHub token file. Only required when SkipPullRequest is false.
	GitHubToken string `json:"gitHubToken"`
	// The name to use on the git commit. Only required when GitEmail is specified and SkipPullRequest is false. If not specified, uses values from the user associated with the access token
	GitName string `json:"gitName"`
	// The email to use on the git commit. Only required when GitName is specified and SkipPullRequest is false. If not specified, uses values from the user associated with the access token.
	GitEmail string `json:"gitEmail"`
	// AssignTo specifies who to assign the created PR to. Takes precedence over onCallAddress and onCallGroup if set.
	AssignTo string `json:"assign_to"`
	// Whether to skip creating the pull request for this bump.
	SkipPullRequest bool `json:"skipPullRequest"`
	// The name used in the address when creating remote. This should be the same name as the fork. If fork does not exist this will be the name of the fork that is created.
	// If it is not the same as the fork, the robot will change the name of the fork to this. Format will be git@github.com:{GitLogin}/{RemoteName}.git
	RemoteName string `json:"remoteName"`
	// The name of the branch that will be used when creating the pull request. If unset, defaults to "autobump".
	HeadBranchName string `json:"headBranchName"`
	// Optional list of labels to add to the bump PR
	Labels []string `json:"labels"`
}

// GitCommand is used to pass the various components of the git command which needs to be executed
type GitCommand struct {
	baseCommand string
	args        []string
	workingDir  string
}

// Call will execute the Git command and switch the working directory if specified
func (gc GitCommand) Call(stdout, stderr io.Writer) error {
	return Call(stdout, stderr, gc.baseCommand, gc.buildCommand()...)
}

func (gc GitCommand) buildCommand() []string {
	args := []string{}
	if gc.workingDir != "" {
		args = append(args, "-C", gc.workingDir)
	}
	args = append(args, gc.args...)
	return args
}

func (gc GitCommand) getCommand() string {
	return fmt.Sprintf("%s %s", gc.baseCommand, strings.Join(gc.buildCommand(), " "))
}

const (
	forkRemoteName = "bumper-fork-remote"

	defaultHeadBranchName = "index-md-autobump"

	gitCmd = "git"
)

// PRHandler is the interface implemented by consumer of prcreator, for
// manipulating the repo, and provides commit messages, PR title and body.
type PRHandler interface {
	// Changes returns a slice of functions, each one does some stuff, and
	// returns commit message for the changes
	Changes() []func(context.Context) (string, error)
	// PRTitleBody returns the body of the PR, this function runs after all
	// changes have been executed
	PRTitleBody() (string, string, error)
}

type HideSecretsWriter struct {
	Delegate io.Writer
	Censor   func(content []byte) []byte
}

func (w HideSecretsWriter) Write(content []byte) (int, error) {
	_, err := w.Delegate.Write(w.Censor(content))
	if err != nil {
		return 0, err
	}
	return len(content), nil
}

func Call(stdout, stderr io.Writer, cmd string, args ...string) error {
	(&logrus.Logger{
		Out:       stderr,
		Formatter: logrus.StandardLogger().Formatter,
		Hooks:     logrus.StandardLogger().Hooks,
		Level:     logrus.StandardLogger().Level,
	}).WithField("cmd", cmd).
		// The default formatting uses a space as separator, which is hard to read if an arg contains a space
		WithField("args", fmt.Sprintf("['%s']", strings.Join(args, "', '"))).
		Info("running command")

	c := exec.Command(cmd, args...)
	c.Stdout = stdout
	c.Stderr = stderr
	return c.Run()
}

func getTreeRef(stderr io.Writer, refname string) (string, error) {
	revParseStdout := &bytes.Buffer{}
	if err := Call(revParseStdout, stderr, gitCmd, "rev-parse", refname+":"); err != nil {
		return "", fmt.Errorf("parse ref: %w", err)
	}
	fields := strings.Fields(revParseStdout.String())
	if n := len(fields); n < 1 {
		return "", errors.New("got no otput when trying to rev-parse")
	}
	return fields[0], nil
}

func gitStatus(stdout io.Writer, stderr io.Writer) (string, error) {
	tmpRead, tmpWrite, err := os.Pipe()
	if err != nil {
		return "", err
	}

	if err := Call(tmpWrite, stderr, gitCmd, "status"); err != nil {
		return "", fmt.Errorf("git status: %w", err)
	}
	tmpWrite.Close()
	output, err := io.ReadAll(tmpRead)
	if err != nil {
		return "", err
	}
	stdout.Write(output)
	return string(output), nil
}

func gitCommit(name, email, message string, stdout, stderr io.Writer) error {
	if err := Call(stdout, stderr, gitCmd, "add", "docs/index.md"); err != nil {
		return fmt.Errorf("git add: %w", err)
	}
	commitArgs := []string{"commit", "-m", message}
	if name != "" && email != "" {
		commitArgs = append(commitArgs, "--author", fmt.Sprintf("%s <%s>", name, email))
	}
	if err := Call(stdout, stderr, gitCmd, commitArgs...); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	return nil
}

func gitPush(remote, remoteBranch string, stdout, stderr io.Writer, dryrun bool) error {
	if err := Call(stdout, stderr, gitCmd, "remote", "add", forkRemoteName, remote); err != nil {
		return fmt.Errorf("add remote: %w", err)
	}
	fetchStderr := &bytes.Buffer{}
	var remoteTreeRef string
	if err := Call(stdout, fetchStderr, gitCmd, "fetch", forkRemoteName, remoteBranch); err != nil {
		logrus.Info("fetchStderr is : ", fetchStderr.String())
		if !strings.Contains(strings.ToLower(fetchStderr.String()), fmt.Sprintf("couldn't find remote ref %s", remoteBranch)) {
			return fmt.Errorf("fetch from fork: %w", err)
		}
	} else {
		var err error
		remoteTreeRef, err = getTreeRef(stderr, fmt.Sprintf("refs/remotes/%s/%s", forkRemoteName, remoteBranch))
		if err != nil {
			return fmt.Errorf("get remote tree ref: %w", err)
		}
	}
	localTreeRef, err := getTreeRef(stderr, "HEAD")
	if err != nil {
		return fmt.Errorf("get local tree ref: %w", err)
	}

	if dryrun {
		logrus.Info("[Dryrun] Skip git push with: ")
		logrus.Info(forkRemoteName, remoteBranch, stdout, stderr, "")
		return nil
	}
	// Avoid doing metadata-only pushes that re-trigger tests and remove lgtm
	if localTreeRef != remoteTreeRef {
		if err := GitPush(forkRemoteName, remoteBranch, stdout, stderr, ""); err != nil {
			return err
		}
	} else {
		logrus.Info("Not pushing as up-to-date remote branch already exists")
	}
	return nil
}

// GitPush push the changes to the given remote and branch.
func GitPush(remote, remoteBranch string, stdout, stderr io.Writer, workingDir string) error {
	logrus.Info("Pushing to remote...")
	gc := GitCommand{
		baseCommand: gitCmd,
		args:        []string{"push", "-f", remote, fmt.Sprintf("HEAD:%s", remoteBranch)},
		workingDir:  workingDir,
	}
	if err := gc.Call(stdout, stderr); err != nil {
		return fmt.Errorf("%s: %w", gc.getCommand(), err)
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

	return nil
}

// Run is the entrypoint which will update index.md file based on the provided options.
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
	stdout := HideSecretsWriter{Delegate: os.Stdout, Censor: secret.Censor}
	stderr := HideSecretsWriter{Delegate: os.Stderr, Censor: secret.Censor}
	if err := secret.Add(o.GitHubToken); err != nil {
		return fmt.Errorf("start secrets agent: %w", err)
	}

	gc, err := github.NewClient(secret.GetTokenGenerator(o.GitHubToken), secret.Censor, github.DefaultGraphQLEndpoint, github.DefaultAPIEndpoint)
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

	resp, err := gitStatus(stdout, stderr)
	if err != nil {
		return fmt.Errorf("git status: %w", err)
	}
	if strings.Contains(resp, "nothing to commit, working tree clean") {
		fmt.Println("No changes, quitting.")
		return nil
	}

	if err := gitCommit(o.GitName, o.GitEmail, "Bumping index.md", stdout, stderr); err != nil {
		return fmt.Errorf("commit changes to the remote branch: %w", err)
	}

	remote := fmt.Sprintf("https://%s:%s@github.com/%s/%s.git", o.GitHubLogin, string(secret.GetTokenGenerator(o.GitHubToken)()), o.GitHubLogin, o.RemoteName)
	if err := gitPush(remote, o.HeadBranchName, stdout, stderr, o.SkipPullRequest); err != nil {
		return fmt.Errorf("push changes to the remote branch: %w", err)
	}

	summary, body, err := prh.PRTitleBody()
	if err != nil {
		return fmt.Errorf("creating PR summary and body: %w", err)
	}
	if o.GitHubBaseBranch == "" {
		repo, err := gc.GetRepo(o.GitHubOrg, o.GitHubRepo)
		if err != nil {
			return fmt.Errorf("detect default remote branch for %s/%s: %w", o.GitHubOrg, o.GitHubRepo, err)
		}
		o.GitHubBaseBranch = repo.DefaultBranch
	}
	if err := updatePRWithLabels(gc, o.GitHubOrg, o.GitHubRepo, getAssignment(o.AssignTo), o.GitHubLogin, o.GitHubBaseBranch, o.HeadBranchName, updater.PreventMods, summary, body, o.Labels, o.SkipPullRequest); err != nil {
		return fmt.Errorf("to create the PR: %w", err)
	}
	return nil
}
