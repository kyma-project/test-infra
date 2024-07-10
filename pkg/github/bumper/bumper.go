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
	// The name used in the address when creating remote. This should be the same name as the fork. If fork does not exist this will be the name of the fork that is created.
	// If it is not the same as the fork, the robot will change the name of the fork to this. Format will be git@github.com:{GitLogin}/{RemoteName}.git
	RemoteName string `json:"remoteName" yaml:"remoteName"`
	// The name of the branch that will be used when creating the pull request. If unset, defaults to "autobump".
	HeadBranchName string `json:"headBranchName" yaml:"headBranchName"`
	// Optional list of labels to add to the bump PR
	Labels []string `json:"labels" yaml:"labels"`
	// The URL where upstream image references are located. Only required if Target Version is "upstream" or "upstreamStaging". Use "https://raw.githubusercontent.com/{ORG}/{REPO}"
	// Images will be bumped based off images located at the address using this URL and the refConfigFile or stagingRefConfigFile for each Prefix.
	UpstreamURLBase string `yaml:"upstreamURLBase"`
	// The config paths to be included in this bump, in which only .yaml files will be considered. By default, all files are included.
	IncludedConfigPaths []string `yaml:"includedConfigPaths"`
	// The config paths to be excluded in this bump, in which only .yaml files will be considered.
	ExcludedConfigPaths []string `yaml:"excludedConfigPaths"`
	// The extra non-yaml file to be considered in this bump.
	ExtraFiles []string `yaml:"extraFiles"`
	// The target version to bump images version to, which can be one of latest, upstream, upstream-staging and vYYYYMMDD-deadbeef.
	TargetVersion string `yaml:"targetVersion"`
	// List of prefixes that the autobumped is looking for, and other information needed to bump them. Must have at least 1 prefix.
	Prefixes []Prefix `yaml:"prefixes"`
	// The oncall address where we can get the JSON file that stores the current oncall information.
	OncallAddress string `json:"onCallAddress"`
	// The oncall group that is responsible for reviewing the change, i.e. "test-infra".
	OncallGroup string `json:"onCallGroup"`
	// Whether skip if no oncall is discovered
	SkipIfNoOncall bool `yaml:"skipIfNoOncall"`
	// SkipOncallAssignment skips assigning to oncall.
	// The OncallAddress and OncallGroup are required for auto-bumper to figure out whether there are active oncall,
	// which is used to avoid bumping when there is no active oncall.
	SkipOncallAssignment bool `yaml:"skipOncallAssignment"`
	// SelfAssign is used to comment `/assign` and `/cc` so that blunderbuss wouldn't assign
	// bump PR to someone else.
	SelfAssign bool `yaml:"selfAssign"`
	// ImageRegistryAuth determines a way the autobumper with authenticate when talking to image registry.
	// Allowed values:
	// * "" (empty) -- uses no auth token
	// * "google" -- uses Google's "Application Default Credentials" as defined on https://pkg.go.dev/golang.org/x/oauth2/google#hdr-Credentials.
	ImageRegistryAuth string `yaml:"imageRegistryAuth"`
	// AdditionalPRBody allows for generic, additional content in the body of the PR
	AdditionalPRBody string `yaml:"additionalPRBody"`
	// GitHubHost is the host of the GitHub instance. If not set, it defaults to "github.com".
	GitHubHost string `yaml:"GitHubHost"`
}

// prefix is the information needed for each prefix being bumped.
type Prefix struct {
	// Name of the tool being bumped
	Name string `yaml:"name"`
	// The image prefix that the autobumper should look for
	Prefix string `yaml:"prefix"`
	// File that is looked at to determine current upstream image when bumping to upstream. Required only if targetVersion is "upstream"
	RefConfigFile string `yaml:"refConfigFile"`
	// File that is looked at to determine current upstream staging image when bumping to upstream staging. Required only if targetVersion is "upstream-staging"
	StagingRefConfigFile string `yaml:"stagingRefConfigFile"`
	// The repo where the image source resides for the images with this prefix. Used to create the links to see comparisons between images in the PR summary.
	Repo string `yaml:"repo"`
	// Whether or not the format of the PR summary for this prefix should be summarised.
	Summarise bool `yaml:"summarise"`
	// Whether the prefix tags should be consistent after the bump
	ConsistentImages bool `yaml:"consistentImages"`
	// A list of images whose tags are not required to be consistent after the bump. Requires `consistentImages: true`.
	ConsistentImageExceptions []string `yaml:"consistentImageExceptions"`
}

const (
	forkRemoteName = "bumper-fork-remote"

	defaultHeadBranchName = "autobump"

	gitCmd = "git"
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

func gitAdd(files []string, stdout, stderr io.Writer) error {
	if err := Call(stdout, stderr, gitCmd, "add", strings.Join(files, " ")); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	return nil
}

func gitCommit(name, email, message string, stdout, stderr io.Writer) error {
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
	stdout := CensoredWriter{Delegate: os.Stdout, Censor: secret.Censor}
	stderr := CensoredWriter{Delegate: os.Stderr, Censor: secret.Censor}
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

	for i, changeFunc := range prh.Changes() {
		commitMsg, filesToBeAdded, err := changeFunc(context.Background())
		if err != nil {
			return fmt.Errorf("failed to process function %d: %s", i, err)
		}

		changed, err := HasChanges()
		if err != nil {
			return fmt.Errorf("checking changes: %w", err)
		}

		if !changed {
			logrus.WithField("function", i).Info("Nothing changed, skip commit ...")
			continue
		}

		resp, err := gitStatus(stdout, stderr)
		if err != nil {
			return fmt.Errorf("git status: %w", err)
		}
		if strings.Contains(resp, "nothing to commit, working tree clean") {
			fmt.Println("No changes, quitting.")
			return nil
		}

		if err := gitAdd(filesToBeAdded, stdout, stderr); err != nil {
			return fmt.Errorf("add changes to commit %w", err)
		}

		if err := gitCommit(o.GitName, o.GitEmail, commitMsg, stdout, stderr); err != nil {
			return fmt.Errorf("commit changes to the remote branch: %w", err)
		}
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
