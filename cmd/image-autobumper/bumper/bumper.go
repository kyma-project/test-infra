/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package bumper

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/kyma-project/test-infra/cmd/image-autobumper/updater"
	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/config/secret"
	"k8s.io/test-infra/prow/github"
)

const (
	forkRemoteName = "bumper-fork-remote"

	defaultHeadBranchName = "autobump"

	gitCmd = "git"
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
	// Whether to signoff the commits.
	Signoff bool `json:"signoff"`
	// The name used in the address when creating remote. This should be the same name as the fork. If fork does not exist this will be the name of the fork that is created.
	// If it is not the same as the fork, the robot will change the name of the fork to this. Format will be git@github.com:{GitLogin}/{RemoteName}.git
	RemoteName string `json:"remoteName"`
	// The name of the branch that will be used when creating the pull request. If unset, defaults to "autobump".
	HeadBranchName string `json:"headBranchName"`
	// Optional list of labels to add to the bump PR
	Labels []string `json:"labels"`
	// The GitHub host to use, defaulting to github.com
	GitHubHost string `json:"gitHubHost"`
}

// PRHandler is the interface implemented by consumer of prcreator, for
// manipulating the repo, and provides commit messages, PR title and body.
type PRHandler interface {
	// Changes returns a slice of functions, each one does some stuff, and
	// returns commit message for the changes
	Changes() []func(context.Context) (string, error)
	// PRTitleBody returns the body of the PR, this function runs after all
	// changes have been executed
	PRTitleBody() (string, string)
}

// GitAuthorOptions is specifically to read the author info for a commit
type GitAuthorOptions struct {
	GitName  string
	GitEmail string
}

// AddFlags will read the author info from the command line parameters
func (o *GitAuthorOptions) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&o.GitName, "git-name", "", "The name to use on the git commit.")
	fs.StringVar(&o.GitEmail, "git-email", "", "The email to use on the git commit.")
}

// Validate will validate the input GitAuthorOptions
func (o *GitAuthorOptions) Validate() error {
	if (o.GitEmail == "") != (o.GitName == "") {
		return fmt.Errorf("--git-name and --git-email must be specified together")
	}
	return nil
}

// GitCommand is used to pass the various components of the git command which needs to be executed
type GitCommand struct {
	baseCommand string
	args        []string
	workingDir  string
}

// Call will execute the Git command and switch the working directory if specified
func (gc GitCommand) Call(stdout, stderr io.Writer, opts ...CallOption) error {
	return Call(stdout, stderr, gc.baseCommand, gc.buildCommand(), opts...)
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
	if o.GitHubHost == "" {
		o.GitHubHost = "github.com"
	}

	return nil
}

// Run is the entrypoint which will update Prow config files based on the
// provided options.
//
// updateFunc: a function that returns commit message and error
func Run(ctx context.Context, o *Options, prh PRHandler) error {
	if err := validateOptions(o); err != nil {
		return fmt.Errorf("validating options: %w", err)
	}

	if o.SkipPullRequest {
		logrus.Debugf("--skip-pull-request is set to true, won't create a pull request.")
	}

	return processGitHub(ctx, o, prh)
}

func processGitHub(ctx context.Context, o *Options, prh PRHandler) error {
	stdout := HideSecretsWriter{Delegate: os.Stdout, Censor: secret.Censor}
	stderr := HideSecretsWriter{Delegate: os.Stderr, Censor: secret.Censor}
	if err := secret.Add(o.GitHubToken); err != nil {
		return fmt.Errorf("start secrets agent: %w", err)
	}

	gitHubHost := "https://api.github.com"
	if o.GitHubHost != "" {
		gitHubHost = fmt.Sprintf("https://%s/api/v3", o.GitHubHost)
	}

	gc, err := github.NewClient(secret.GetTokenGenerator(o.GitHubToken), secret.Censor, gitHubHost, gitHubHost)
	if err != nil {
		return fmt.Errorf("failed to construct GitHub client: %v", err)
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

	// Make change, commit and push
	var anyChange bool
	for i, changeFunc := range prh.Changes() {
		msg, err := changeFunc(ctx)
		if err != nil {
			return fmt.Errorf("process function %d: %w", i, err)
		}

		changed, err := HasChanges()
		if err != nil {
			return fmt.Errorf("checking changes: %w", err)
		}

		if !changed {
			logrus.WithField("function", i).Info("Nothing changed, skip commit ...")
			continue
		}

		anyChange = true
		if err := gitCommit(o.GitName, o.GitEmail, msg, stdout, stderr, o.Signoff); err != nil {
			return fmt.Errorf("git commit: %w", err)
		}
	}
	if !anyChange {
		logrus.Info("Nothing changed from all functions, skip PR ...")
		return nil
	}

	if err := MinimalGitPush(fmt.Sprintf("https://%s:%s@%s/%s/%s.git", o.GitHubLogin, string(secret.GetTokenGenerator(o.GitHubToken)()), o.GitHubHost, o.GitHubLogin, o.RemoteName), o.HeadBranchName, stdout, stderr, o.SkipPullRequest); err != nil {
		return fmt.Errorf("push changes to the remote branch: %w", err)
	}

	summary, body := prh.PRTitleBody()
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

type callOptions struct {
	ctx context.Context
	dir string
}

type CallOption func(*callOptions)

func Call(stdout, stderr io.Writer, cmd string, args []string, opts ...CallOption) error {
	var options callOptions
	for _, opt := range opts {
		opt(&options)
	}
	logger := (&logrus.Logger{
		Out:       stderr,
		Formatter: logrus.StandardLogger().Formatter,
		Hooks:     logrus.StandardLogger().Hooks,
		Level:     logrus.StandardLogger().Level,
	}).WithField("cmd", cmd).
		// The default formatting uses a space as separator, which is hard to read if an arg contains a space
		WithField("args", fmt.Sprintf("['%s']", strings.Join(args, "', '")))

	if options.dir != "" {
		logger = logger.WithField("dir", options.dir)
	}
	logger.Info("running command")

	var c *exec.Cmd
	if options.ctx != nil {
		c = exec.CommandContext(options.ctx, cmd, args...)
	} else {
		c = exec.Command(cmd, args...)
	}
	c.Stdout = stdout
	c.Stderr = stderr
	if options.dir != "" {
		c.Dir = options.dir
	}
	return c.Run()
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

func updatePRWithLabels(gc github.Client, org, repo string, extraLineInPRBody, login, baseBranch, headBranch string, allowMods bool, summary, body string, labels []string, dryrun bool) error {
	return UpdatePullRequestWithLabels(gc, org, repo, summary, generatePRBody(body, extraLineInPRBody), login+":"+headBranch, baseBranch, headBranch, allowMods, labels, dryrun)
}

// UpdatePullRequestWithLabels updates with GitHub client "gc" the PR of GitHub repo org/repo
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

// HasChanges checks if the current git repo contains any changes
func HasChanges() (bool, error) {
	// Configure Git to recognize the /workspace directory as safe
	additionalArgs := []string{"config", "--global", "user.email", "dl_666c0cf3e82c7d0136da22ea@global.corp.sap"}
	logrus.WithField("cmd", gitCmd).WithField("args", additionalArgs).Info("running command ...")
	additionalOutput, configErr := exec.Command(gitCmd, additionalArgs...).CombinedOutput()
	if configErr != nil {
		logrus.WithField("cmd", gitCmd).Debugf("output is '%s'", string(additionalOutput))
		return false, fmt.Errorf("running command %s %s: %w", gitCmd, additionalArgs, configErr)
	}

	additionalArgs2 := []string{"config", "--global", "user.name", "autobumper-github-tools-sap-serviceuser"}
	logrus.WithField("cmd", gitCmd).WithField("args", additionalArgs2).Info("running command ...")
	additionalOutput2, configErr := exec.Command(gitCmd, additionalArgs2...).CombinedOutput()
	if configErr != nil {
		logrus.WithField("cmd", gitCmd).Debugf("output is '%s'", string(additionalOutput2))
		return false, fmt.Errorf("running command %s %s: %w", gitCmd, additionalArgs2, configErr)
	}

	// Configure Git to recognize the /workspace directory as safe
	configArgs := []string{"config", "--global", "--add", "safe.directory", "'*'"}
	logrus.WithField("cmd", gitCmd).WithField("args", configArgs).Info("running command ...")
	configOutput, configErr := exec.Command(gitCmd, configArgs...).CombinedOutput()
	if configErr != nil {
		logrus.WithField("cmd", gitCmd).Debugf("output is '%s'", string(configOutput))
		return false, fmt.Errorf("running command %s %s: %w", gitCmd, configArgs, configErr)
	}

	// Check for changes using git status
	statusArgs := []string{"status", "--porcelain"}
	logrus.WithField("cmd", gitCmd).WithField("args", statusArgs).Info("running command ...")
	combinedOutput, err := exec.Command(gitCmd, statusArgs...).CombinedOutput()
	if err != nil {
		logrus.WithField("cmd", gitCmd).Debugf("output is '%s'", string(combinedOutput))
		return false, fmt.Errorf("running command %s %s: %w", gitCmd, statusArgs, err)
	}
	hasChanges := len(strings.TrimSuffix(string(combinedOutput), "\n")) > 0

	// If there are changes, get the diff
	if hasChanges {
		diffArgs := []string{"diff"}
		logrus.WithField("cmd", gitCmd).WithField("args", diffArgs).Info("running command ...")
		diffOutput, diffErr := exec.Command(gitCmd, diffArgs...).CombinedOutput()
		if diffErr != nil {
			logrus.WithField("cmd", gitCmd).Debugf("output is '%s'", string(diffOutput))
			return true, fmt.Errorf("running command %s %s: %w", gitCmd, diffArgs, diffErr)
		}
		logrus.WithField("cmd", gitCmd).Debugf("diff output is '%s'", string(diffOutput))
	}

	return hasChanges, nil
}

func gitCommit(name, email, message string, stdout, stderr io.Writer, signoff bool) error {
	if err := Call(stdout, stderr, gitCmd, []string{"add", "-A"}); err != nil {
		return fmt.Errorf("git add: %w", err)
	}
	commitArgs := []string{"commit", "-m", message}
	if name != "" && email != "" {
		commitArgs = append(commitArgs, "--author", fmt.Sprintf("%s <%s>", name, email))
	}
	if signoff {
		commitArgs = append(commitArgs, "--signoff")
	}
	if err := Call(stdout, stderr, gitCmd, commitArgs); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	return nil
}

// MinimalGitPush pushes the content of the local repository to the remote, checking to make
// sure that there are real changes that need updating by diffing the tree refs, ensuring that
// no metadata-only pushes occur, as those re-trigger tests, remove LGTM, and cause churn without
// changing the content being proposed in the PR.
func MinimalGitPush(remote, remoteBranch string, stdout, stderr io.Writer, dryrun bool, opts ...CallOption) error {
	if err := Call(stdout, stderr, gitCmd, []string{"remote", "add", forkRemoteName, remote}, opts...); err != nil {
		return fmt.Errorf("add remote: %w", err)
	}
	fetchStderr := &bytes.Buffer{}
	var remoteTreeRef string
	if err := Call(stdout, fetchStderr, gitCmd, []string{"fetch", forkRemoteName, remoteBranch}, opts...); err != nil {
		logrus.Info("fetchStderr is : ", fetchStderr.String())
		if !strings.Contains(strings.ToLower(fetchStderr.String()), fmt.Sprintf("couldn't find remote ref %s", remoteBranch)) {
			return fmt.Errorf("fetch from fork: %w", err)
		}
	} else {
		var err error
		remoteTreeRef, err = getTreeRef(stderr, fmt.Sprintf("refs/remotes/%s/%s", forkRemoteName, remoteBranch), opts...)
		if err != nil {
			return fmt.Errorf("get remote tree ref: %w", err)
		}
	}
	localTreeRef, err := getTreeRef(stderr, "HEAD", opts...)
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
		if err := GitPush(forkRemoteName, remoteBranch, stdout, stderr, "", opts...); err != nil {
			return err
		}
	} else {
		logrus.Info("Not pushing as up-to-date remote branch already exists")
	}
	return nil
}

// GitPush push the changes to the given remote and branch.
func GitPush(remote, remoteBranch string, stdout, stderr io.Writer, workingDir string, opts ...CallOption) error {
	logrus.Info("Pushing to remote...")
	gc := GitCommand{
		baseCommand: gitCmd,
		args:        []string{"push", "-f", remote, fmt.Sprintf("HEAD:%s", remoteBranch)},
		workingDir:  workingDir,
	}
	if err := gc.Call(stdout, stderr, opts...); err != nil {
		return fmt.Errorf("%s: %w", gc.getCommand(), err)
	}
	return nil
}
func generatePRBody(body, assignment string) string {
	return body + assignment + "\n"
}

func getAssignment(assignTo string) string {
	if assignTo != "" {
		return "/cc @" + assignTo
	}
	return ""
}

func getTreeRef(stderr io.Writer, refname string, opts ...CallOption) (string, error) {
	revParseStdout := &bytes.Buffer{}
	if err := Call(revParseStdout, stderr, gitCmd, []string{"rev-parse", refname + ":"}, opts...); err != nil {
		return "", fmt.Errorf("parse ref: %w", err)
	}
	fields := strings.Fields(revParseStdout.String())
	if n := len(fields); n < 1 {
		return "", errors.New("got no output when trying to rev-parse")
	}
	return fields[0], nil
}
