// Package pjtester implements tool for constructing prowjob specification and running it on Prow cluster to test it.
// Prowjobs are scheduled on Prow cluster by creating a prowjob resource on cluster directly.
// Prowjobs are always tested on untrusted cluster.
// Tested prowjobs are prefixed with meaningful prefix and labeled to indicate it's a pjtester scheduled prowjob test.
// Test prowjob means a prowjob which is under test.
// Test pull request is a pr used in test prowjob execution.
// A pjtester prowjob means a prowjob which runs pjtester.
// A pjtester pull request is a pr containing pjtester.yaml file and triggering pjtester prowjob.
// A pjtester pr and test pr can be the same pull request.
package pjtester

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kyma-project/test-infra/pkg/github/git"
	"github.com/kyma-project/test-infra/pkg/tools/prtagbuilder"
	"os"
	"path"
	"strconv"
	"strings"

	"sigs.k8s.io/prow/prow/pod-utils/downwardapi"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	prowapi "sigs.k8s.io/prow/prow/apis/prowjobs/v1"
	prowclient "sigs.k8s.io/prow/prow/client/clientset/versioned"
	"sigs.k8s.io/prow/prow/config"
	prowflagutil "sigs.k8s.io/prow/prow/flagutil"
	"sigs.k8s.io/prow/prow/github"
	"sigs.k8s.io/prow/prow/pjutil"
)

// Default values for kyma-project/test-infra
const (
	defaultPjPath      = "prow/jobs/"
	defaultConfigPath  = "prow/config.yaml"
	defaultMainBranch  = "main"
	defaultClonePath   = "/home/prow/go/src/github.com"
	testinfraOrg       = "kyma-project"
	testinfraRepo      = "test-infra"
	pjtesterConfigPath = "vpath/pjtester.yaml"
)

var (
	envVarsList = []string{"KUBECONFIG_PATH", "PULL_BASE_REF", "PULL_BASE_SHA", "PULL_NUMBER", "PULL_PULL_SHA", "JOB_SPEC", "REPO_OWNER", "REPO_NAME"}
	log         = logrus.New()
)

// GithubClient is used for mocking
type GithubClient interface {
	github.UserClient
	GetRef(org, repo, ref string) (string, error)
	GetPullRequest(org, repo string, number int) (*github.PullRequest, error)
}

// pjConfig holds user provided configuration of test prowjob.
type pjConfig struct {
	PjName string `yaml:"pjName" validate:"required,min=1"` // Test prowjob name.
	Report bool   `yaml:"report,omitempty"`                 // Enable reporting of prowjob status, default reporting is enabled.
}

// prConfig holds user and github provided details about pull request.
type prConfig struct {
	PrNumber    int                `yaml:"prNumber" validate:"required,number,min=1"` // User provided pr number.
	pullRequest github.PullRequest // Pull request details fetched from github.
	org         string             // PR organisation name.
	repo        string             // PR repository name.
}

// pjConfigs holds user provided configuration used by pjtester to find test prowjob definition.
type pjConfigs struct {
	PrConfig map[string]map[string]prConfig   `yaml:"prConfig,omitempty" validate:"max=1,dive,max=1"`           // Map key represent github organisation.
	ProwJobs map[string]map[string][]pjConfig `yaml:"prowJobs" validate:"required,min=1,dive,min=1,dive,min=1"` // Map key represent github organisation.
}

// testCfg holds user provided configuration for pjtester.
type testCfg struct {
	PjConfigs pjConfigs                      `yaml:"pjConfigs" validate:"required"`
	PrConfigs map[string]map[string]prConfig `yaml:"prConfigs,omitempty"` // Holds pull request details used in test prowjobs. Map key represent github organisation name.
}

// options holds commmon configuration for pjtester and test prowjobs.
type options struct {
	configPath    string // configPath is a location of prow config file to use to construct test prowjob.
	jobConfigPath string // jobConfigPath is a location of prowjob definition file to use to construct test prowjob.
	clonePath     string

	pjtesterPrBaseRef string // pjtesterPrBaseRef is a base branch name of pjtester pull request.
	pjtesterPrBaseSha string // pjtesterPrBaseSha is a git SHA of a base branch of github pull request under test
	pjtesterPrNumber  int    // pjtesterPrNumber is a number of github pull request under test
	pjtesterPrHeadSha string // pjtesterPrHeadSha is SHA of github pull request head under test
	pjtesterPrAuthor  string // pjtesterPrAuthor is an author of github pull request under test
	pjtesterPrOrg     string // pjtesterPrOrg is a name of organisation of pull request for pjtester prowjob.
	pjtesterPrRepo    string // pjtesterPrRepo is a name of organisation of pull request for pjtester prowjob.

	github              github.Client
	githubClient        GithubClient
	gitOptions          git.ClientConfig
	gitClient           git.Client
	prFinder            *prtagbuilder.GitHubClient
	testPullRequests    map[string]map[string]prConfig // pull requests used to run test prowjobs.
	pjConfigPullRequest prConfig                       // pull request used to load test prowjobs definition file.
}

// testProwJobOptions holds configuration specific for test prowjob.
type testProwJobOptions struct {
	orgName       string
	repoName      string
	pjConfig      pjConfig
	baseSHAGetter config.RefGetter // func providing base SHA to load inrepo config from.
	headSHAGetter config.RefGetter // func providing head SHA to load inrepo config from.
}

// checkEnvVars validate if required env variables are set.
func checkEnvVars(varsList []string) error {
	for _, evar := range varsList {
		val, present := os.LookupEnv(evar)
		if present {
			if len(val) == 0 {
				return fmt.Errorf("variable %s is empty", evar)
			}
		} else {
			return fmt.Errorf("variable %s is not set", evar)
		}
	}
	return nil
}

// newProwK8sClientset is building Prow client for provided kubeconfig.
// Kubeconfig location is taken from KUBECONFIG_PATH env variable.
func newProwK8sClientset() (*prowclient.Clientset, error) {
	k8sConfig, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG_PATH"))
	if err != nil {
		return nil, fmt.Errorf("failed create config for prow k8s clientset")
	}
	clientset, err := prowclient.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("failed create prow k8s clientset")
	}
	return clientset, nil
}

// readTestCfg read and validate data from pjtester.yaml file.
// Validation is done based on requirements defined in struct fields tags.
func readTestCfg(testCfgFile string) (testCfg, error) {
	var t testCfg
	yamlFile, err := os.ReadFile(testCfgFile)
	if err != nil {
		return testCfg{}, fmt.Errorf("failed read test config file from vpath/pjtester.yaml")
	}
	err = yaml.Unmarshal(yamlFile, &t)
	if err != nil {
		return testCfg{}, fmt.Errorf("failed unmarshal test config yaml")
	}

	validate := validator.New()
	// returns nil or ValidationErrors ( []FieldError )
	err = validate.Struct(t)
	if err != nil {
		return testCfg{}, err
	}
	return t, nil
}

// newCommonOptions builds common options and GitHub client for all tests.
// Options are build from PR env variables.
func newCommonOptions(ghOptions *prowflagutil.GitHubOptions) (options, error) {
	var o options
	var err error
	ghc, err := ghOptions.GitHubClient(false)
	if err != nil {
		return options{}, err
	}
	o.github = ghc
	o.clonePath = defaultClonePath
	o.pjtesterPrBaseRef = os.Getenv("PULL_BASE_REF")
	o.pjtesterPrBaseSha = os.Getenv("PULL_BASE_SHA")
	o.pjtesterPrOrg = os.Getenv("REPO_OWNER")
	o.pjtesterPrRepo = os.Getenv("REPO_NAME")
	o.pjtesterPrNumber, err = strconv.Atoi(os.Getenv("PULL_NUMBER"))
	if err != nil {
		return options{}, fmt.Errorf("could not get pull number from env var PULL_NUMBER")
	}
	o.pjtesterPrHeadSha = os.Getenv("PULL_PULL_SHA")
	o.pjtesterPrAuthor = gjson.Get(os.Getenv("JOB_SPEC"), "refs.pulls.0.author").String()
	o.jobConfigPath = path.Join(defaultClonePath, testinfraOrg, testinfraRepo, defaultPjPath)
	o.configPath = path.Join(defaultClonePath, testinfraOrg, testinfraRepo, defaultConfigPath)
	return o, nil
}

// getPullRequests download pull request details from GitHub.
// Returned map first level keys represent GitHub organisations names, second level represent repositories names.
// Function checks if pull request details were already downloaded and saved in options.testPullRequests.
// This check avoids downloading the same pull request details if pjtester use the same pull request as pr with prowjob
// definition and pr to use as test prowjob refs or extrarefs.
// Because of that check, downloading testPullRequests should be done before downloading pjtester pull request.
func (o *options) getPullRequests(prconfig map[string]map[string]prConfig) (map[string]map[string]prConfig, error) {
	log.Debugf("Downloading pull requests details from GitHub.")
	pullRequests := make(map[string]map[string]prConfig)
	for org, repos := range prconfig {
		if _, ok := pullRequests[org]; !ok {
			pullRequests[org] = make(map[string]prConfig)
		}
		for repo, prcfg := range repos {
			// If the same PR is provided as test prowjob pr do not download it again.
			if pr, present := o.testPullRequests[org][repo]; present {
				if pr.PrNumber == prcfg.PrNumber {
					log.Debugf("This same pull request is used as test prowjob refs. Using it: %s #%d", pr.org+"/"+pr.repo, prcfg.PrNumber)
					pullRequests[org][repo] = pr
					continue
				}
			}
			log.Debugf("Downloading pull request %s #%d details from GitHub.", org+"/"+repo, prcfg.PrNumber)
			pr, err := o.githubClient.GetPullRequest(org, repo, prcfg.PrNumber)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch PullRequest from GitHub, error: %w", err)
			}
			prcfg.pullRequest = *pr
			prcfg.org = org
			prcfg.repo = repo
			pullRequests[org][repo] = prcfg
		}
	}
	log.Debugf("Finished downloading pull requests details.")
	return pullRequests, nil
}

// getPreAndPostSubmits loads presubmits and postsubmits specifications for defined repository from static and inrepo config.
func (pjopts *testProwJobOptions) getPreAndPostSubmits(gitClient git.Client, conf *config.Config) ([]config.Presubmit, []config.Postsubmit, error) {
	if pjopts.headSHAGetter != nil {
		log.Debugf("Using headSHAGetter")
		preSubmits, err := conf.GetPresubmits(gitClient, fmt.Sprintf("%s/%s", pjopts.orgName, pjopts.repoName), pjopts.baseSHAGetter, pjopts.headSHAGetter)
		if err != nil {
			return nil, nil, fmt.Errorf("failed get presubmits, error: %s", err)
		}
		postSubmits, err := conf.GetPostsubmits(gitClient, fmt.Sprintf("%s/%s", pjopts.orgName, pjopts.repoName), pjopts.baseSHAGetter, pjopts.headSHAGetter)
		if err != nil {
			return nil, nil, fmt.Errorf("failed get postsubmits, error: %s", err)
		}
		log.Debugf("Finished loading presubmits and postsubmits.")
		return preSubmits, postSubmits, nil
	}
	log.Debugf("Not using headSHAGetter")
	preSubmits, err := conf.GetPresubmits(gitClient, fmt.Sprintf("%s/%s", pjopts.orgName, pjopts.repoName), pjopts.baseSHAGetter)
	if err != nil {
		return nil, nil, fmt.Errorf("failed get presubmits, error: %s", err)
	}
	postSubmits, err := conf.GetPostsubmits(gitClient, fmt.Sprintf("%s/%s", pjopts.orgName, pjopts.repoName), pjopts.baseSHAGetter)
	if err != nil {
		return nil, nil, fmt.Errorf("failed get postsubmits, error: %s", err)
	}
	log.Debugf("Finished loading presubmits and postsubmits.")
	return preSubmits, postSubmits, nil
}

// genJobSpec generate test prowjob specification.
// It will set prowjob refs and extra refs to match scenario provided in pjtester config file.
// It will prefix name and context with pjtester prefix.
func (pjopts *testProwJobOptions) genJobSpec(o options, conf *config.Config, pjCfg pjConfig) (config.JobBase, prowapi.ProwJobSpec, error) {
	var (
		preSubmits  []config.Presubmit
		postSubmits []config.Postsubmit
		err         error
	)

	preSubmits, postSubmits, err = pjopts.getPreAndPostSubmits(o.gitClient, conf)
	if err != nil {
		return config.JobBase{}, prowapi.ProwJobSpec{}, fmt.Errorf("failed load presubmits or postsubmits, error: %s", err)
	}

	log.Debugf("pjconfig pjname: %s", pjCfg.PjName)

	for _, p := range preSubmits {
		log.Debugf("presubmit.name : %s", p.Name)
		if p.Name == pjCfg.PjName {
			p.Optional = true
			// Add prefix to prowjob to test name.
			p.Name = formatPjName(o.pjtesterPrAuthor, p.Name)
			// Add prefix to prowjob to test context.
			p.Context = p.Name
			pjs := pjutil.PresubmitSpec(p, prowapi.Refs{
				Org:  pjopts.orgName,
				Repo: pjopts.repoName,
			})
			pjs, err = pjopts.setProwJobSpecRefs(pjs, o)
			if err != nil {
				return config.JobBase{}, prowapi.ProwJobSpec{}, fmt.Errorf("failed generate presubmit refs, error: %w", err)
			}
			pjs, err = pjopts.setProwJobSpecExtraRefs(pjs, o)
			if err != nil {
				return config.JobBase{}, prowapi.ProwJobSpec{}, fmt.Errorf("failed generate presubmit extrarefs, error: %w", err)
			}
			return p.JobBase, pjs, nil
		}
	}
	for _, p := range postSubmits {
		if p.Name == pjCfg.PjName {
			// Add prefix to prowjob to test name.
			p.Name = formatPjName(o.pjtesterPrAuthor, p.Name)
			// Add prefix to prowjob to test context.
			p.Context = p.Name
			pjs := pjutil.PostsubmitSpec(p, prowapi.Refs{
				Org:  pjopts.orgName,
				Repo: pjopts.repoName,
			})
			pjs, err = pjopts.setProwJobSpecRefs(pjs, o)
			if err != nil {
				return config.JobBase{}, prowapi.ProwJobSpec{}, fmt.Errorf("failed generate postsubmit refs, error: %w", err)
			}
			pjs, err = pjopts.setProwJobSpecExtraRefs(pjs, o)
			if err != nil {
				return config.JobBase{}, prowapi.ProwJobSpec{}, fmt.Errorf("failed generate postsubmit extrarefs, error: %w", err)
			}
			return p.JobBase, pjs, nil
		}
	}

	for _, p := range conf.Periodics {
		if p.Name == pjCfg.PjName {
			// Add prefix to prowjob to test name.
			p.Name = formatPjName(o.pjtesterPrAuthor, p.Name)
			var err error
			pjs := pjutil.PeriodicSpec(p)
			pjs, err = pjopts.setProwJobSpecExtraRefs(pjs, o)
			if err != nil {
				return config.JobBase{}, prowapi.ProwJobSpec{}, fmt.Errorf("failed generate periodic extrarefs, error: %w", err)
			}
			return p.JobBase, pjs, nil
		}
	}
	return config.JobBase{}, prowapi.ProwJobSpec{}, fmt.Errorf("prowjob to test not found in prowjob specification files")
}

// setRefs set prowjob refs.
func setRefs(ref prowapi.Refs, baseSHA, baseRef, pullAuthor, pullSHA string, pullNumber int) prowapi.Refs {
	ref.BaseSHA = baseSHA
	ref.BaseRef = baseRef
	ref.Pulls = []prowapi.Pull{{
		Author: pullAuthor,
		Number: pullNumber,
		SHA:    pullSHA,
	}}
	return ref
}

// setProwJobSpecRefs set prowjob specification refs.
func (pjopts *testProwJobOptions) setProwJobSpecRefs(pjs prowapi.ProwJobSpec, opt options) (prowapi.ProwJobSpec, error) {
	// If exist, use pull requests provided in pjtester configuration file.
	if pr, present := opt.testPullRequests[pjs.Refs.Org][pjs.Refs.Repo]; present {
		refs := setRefs(*pjs.Refs, pr.pullRequest.Base.SHA, pr.pullRequest.Base.Ref, pr.pullRequest.AuthorAssociation, pr.pullRequest.Head.SHA, pr.pullRequest.Number)
		pjs.Refs = &refs
		return pjs, nil
	}
	// Use pjtester pull request if it's in the same repo as test prowjob.
	if pjs.Refs.Org == opt.pjtesterPrOrg && pjs.Refs.Repo == opt.pjtesterPrRepo {
		// set refs with details of tested PR
		refs := setRefs(*pjs.Refs, opt.pjtesterPrBaseSha, opt.pjtesterPrBaseRef, opt.pjtesterPrAuthor, opt.pjtesterPrHeadSha, opt.pjtesterPrNumber)
		pjs.Refs = &refs
		return pjs, nil
	}
	// Use main branch head as refs
	pjs.Refs.BaseRef = defaultMainBranch
	branchPR, err := findLatestPR(pjs, opt)
	if err != nil {
		return prowapi.ProwJobSpec{}, fmt.Errorf("failed find latest PR, error: %w", err)
	}
	// For presubmits we need set pulls details in refs. Find latest PR merged to main and use it as refs.
	if pjs.Type == "presubmit" {
		pr, err := opt.githubClient.GetPullRequest(pjs.Refs.Org, pjs.Refs.Repo, branchPR)
		if err != nil {
			return prowapi.ProwJobSpec{}, fmt.Errorf("failed get pull request details from GitHub, error: %w", err)
		}
		refs := setRefs(*pjs.Refs, pr.Base.SHA, pr.Base.Ref, pr.AuthorAssociation, pr.Head.SHA, pr.Number)
		pjs.Refs = &refs
	}
	return pjs, nil
}

// setProwJobSpecExtraRefs set prowjob specification refs.
func (pjopts *testProwJobOptions) setProwJobSpecExtraRefs(pjs prowapi.ProwJobSpec, opt options) (prowapi.ProwJobSpec, error) {
	// Search and use pjtester pull request refs as extraRefs in test prowjob.
	for index, ref := range pjs.ExtraRefs {
		if ref.Org == opt.pjtesterPrOrg && ref.Repo == opt.pjtesterPrRepo {
			pjs.ExtraRefs[index] = setRefs(ref, opt.pjtesterPrBaseSha, opt.pjtesterPrBaseRef, opt.pjtesterPrAuthor, opt.pjtesterPrHeadSha, opt.pjtesterPrNumber)
		}
	}
	// Search and use pull requests refs provided in pjtester config file as extraRefs in test prowjob.
	// Override extraRefs set to pjtester pull request refs if both scenarios were matched.
	for index, ref := range pjs.ExtraRefs {
		// Add PR details to ExtraRefs if PR number was provided in pjtester.yaml
		if pr, present := opt.testPullRequests[ref.Org][ref.Repo]; present {
			pjs.ExtraRefs[index] = setRefs(ref, pr.pullRequest.Base.SHA, pr.pullRequest.Base.Ref, pr.pullRequest.AuthorAssociation, pr.pullRequest.Head.SHA, pr.pullRequest.Number)
		}
	}
	return pjs, nil
}

// findLatestPR return latest pull request number merged to branch name provided in prowapi.ProwJobSpec
func findLatestPR(pjs prowapi.ProwJobSpec, opt options) (int, error) {
	jobSpec := &downwardapi.JobSpec{Refs: pjs.Refs}
	branchPrAsString, err := prtagbuilder.BuildPrTag(jobSpec, true, true, opt.prFinder)
	if err != nil {
		return 0, fmt.Errorf("could not get pr number for %s branch head, got error: %w", pjs.Refs.BaseRef, err)
	}
	branchPR, err := strconv.Atoi(branchPrAsString)
	if err != nil {
		return 0, fmt.Errorf("failed converting pr number string to integer, got error: %w", err)
	}
	return branchPR, nil
}

// formatPjName add pjtester test prefix and formats testing prowjobname to match gcp cluster labels restrictions.
// Prowjobname is formated to lowercase and trimed to 63 bytes.
func formatPjName(pullAuthor, pjName string) string {
	fullName := fmt.Sprintf("%s_test_of_prowjob_%s", pullAuthor, pjName)
	formated := strings.ToLower(fullName)
	// Cut prowjob name to not exceed 63 bytes.
	if len(formated) > 63 {
		runes := bytes.Runes([]byte(formated))
		for i := len(runes); i > 2; i-- {
			if len(string(runes[:i])) <= 63 {
				return string(runes[:i])
			}
		}
	}
	return formated
}

// checkoutTestInfraPjConfigPR checkout options.pjConfigPullRequest in local test-infra repo.
// Test-infra repo is always cloned as extraRefs in pjtester prowjob.
func (o *options) checkoutTestInfraPjConfigPR() error {
	log.Debugf("Checkout kyma-project/test-infra pull request #%d", o.pjConfigPullRequest.PrNumber)
	repoclient, _, err := o.gitClient.GetGitRepoClientFromDir(testinfraOrg, testinfraRepo, path.Join(o.clonePath, testinfraOrg, testinfraRepo))
	if err != nil {
		return fmt.Errorf("failed get git client for repository from local directory %s, error: %w", path.Join(o.clonePath, testinfraOrg, testinfraRepo), err)
	}
	err = repoclient.CheckoutPullRequest(o.pjConfigPullRequest.PrNumber)
	if err != nil {
		return fmt.Errorf("failed checkout pull request %s, error: %w", fmt.Sprintf("%s/%s#%d", testinfraOrg, testinfraRepo, o.pjConfigPullRequest.PrNumber), err)
	}
	log.Debugf("Successful checkout kyma-project/test-infra pull request #%d", o.pjConfigPullRequest.PrNumber)
	return nil
}

// setRefsGetters set options.baseSHAGetter and options.headSHAGetter or checkout test-infra PR with prowjob definition.
// This is used to get correct version of prowjob definition for user provided pjtester config.
func (pjopts *testProwJobOptions) setRefsGetters(opts options) error {
	log.Debugf("Setting base and head refs getters to download inrepo config.")
	// Use pull request with test prowjob definition if it was provided in pjtester config.
	if opts.pjConfigPullRequest.pullRequest.Number != 0 {
		log.Debugf("Pull request with prowjob definition provided")
		if opts.pjConfigPullRequest.org == testinfraOrg && opts.pjConfigPullRequest.repo == testinfraRepo {
			log.Debugf("Pull request with prowjob definition is from test-infra, skip set head getters, checkout pr in local repo.")
			// PR with test prowjob definition is from test-infra repo. Test-infra holds static prowjobs defnitions.
			// Checkout pr locally to load static test prowjob definition from jobConfigPath.
			err := opts.checkoutTestInfraPjConfigPR()
			if err != nil {
				return fmt.Errorf("failed checkout test-infra repo to commit with pjtester config, error: %w", err)
			}
			baseSHA, err := pjopts.getMainBranchSHA(opts)
			if err != nil {
				return fmt.Errorf("failed to get baseSHA: %w", err)
			}
			log.Debugf("Set baseSHAGetter to return: %s", baseSHA)
			pjopts.baseSHAGetter = func() (string, error) {
				return baseSHA, nil
			}
			log.Debugf("Do not use headSHAGetter, set it to null.")
			pjopts.headSHAGetter = nil
			return nil
		}
		log.Debugf("Pull request with prowjob definition not from test-infra")
		log.Debugf("Set baseSHAGetter to return: %s", opts.pjConfigPullRequest.pullRequest.Base.SHA)
		// PR with test prowjob definition is not from test-infra repo.
		// Prow must load prowjob definition as inrepo config.
		// Set base and head getters to point to the PR base SHA and head SHA.
		pjopts.baseSHAGetter = func() (string, error) {
			return opts.pjConfigPullRequest.pullRequest.Base.SHA, nil
		}
		log.Debugf("Set headSHAGetter to return: %s", opts.pjConfigPullRequest.pullRequest.Head.SHA)
		pjopts.headSHAGetter = func() (string, error) {
			return opts.pjConfigPullRequest.pullRequest.Head.SHA, nil
		}
		return nil
	}
	log.Debugf("Pull request with prowjob definition not provided")
	// Pull request with test prowjob was not provided in pjtester config file.
	if pjopts.orgName == opts.pjtesterPrOrg && pjopts.repoName == opts.pjtesterPrRepo {
		log.Debugf("Test prowjob is defined for the same repo where pjtester pull request exist, prowjob org: %s, prowjob repo %s", pjopts.orgName, pjopts.repoName)
		log.Debugf("Using pjtester pull request as source of prowjob definition, set it to return: %s", opts.pjtesterPrBaseSha)
		pjopts.baseSHAGetter = func() (string, error) {
			return opts.pjtesterPrBaseSha, nil
		}
		log.Debugf("Use headSHAGetter, set it to return: %s", opts.pjtesterPrHeadSha)
		pjopts.headSHAGetter = func() (string, error) {
			return opts.pjtesterPrHeadSha, nil
		}
		return nil
	}
	log.Debugf("Test prowjob is defined for other repo than pjtester pull request exist, prowjob org: %s, prowjob repo %s", pjopts.orgName, pjopts.repoName)
	log.Debugf("Set baseSHAGetter to get test prowjob definition from main branch head.")
	baseSHA, err := pjopts.getMainBranchSHA(opts)
	if err != nil {
		return fmt.Errorf("failed to get baseSHA: %w", err)
	}
	log.Debugf("Set baseSHAGetter to return: %s", baseSHA)
	pjopts.baseSHAGetter = func() (string, error) {
		return baseSHA, nil
	}
	log.Debugf("Do not use headSHAGetter, set it to null.")
	pjopts.headSHAGetter = nil
	return nil
}

// getMainBranchSHA get baseSHA of heads/main for test prowjob repository.
func (pjopts *testProwJobOptions) getMainBranchSHA(opts options) (string, error) {
	log.Debugf("Downloading SHA of %s heads/main", pjopts.orgName+"/"+pjopts.repoName)
	var err error
	baseSHA, err := opts.githubClient.GetRef(pjopts.orgName, pjopts.repoName, "heads/main")
	if err != nil {
		return "", fmt.Errorf("failed to get baseSHA: %w", err)
	}
	return baseSHA, nil
}

// newTestPJ is building a ProwJob k8s resource specification for test prowjobs provided in pjtester configuration.
func (pjopts *testProwJobOptions) newTestPJ(pjCfg pjConfig, opt options) (prowapi.ProwJob, error) {
	log.Debugf("Preparing test prowjob %s for repo %s", pjCfg.PjName, pjopts.orgName+"/"+pjopts.repoName)
	err := pjopts.setRefsGetters(opt)
	if err != nil {
		return prowapi.ProwJob{}, fmt.Errorf("failed set RefsGetters, error: %w", err)
	}

	log.Debugf("Loading Prow config from %s and jobs config from %s.", opt.configPath, opt.jobConfigPath)
	// Loading Prow config and Prow Jobs config from files. If files were changed in pull request, new values will be used for test.
	conf, err := config.Load(opt.configPath, opt.jobConfigPath, nil, "")
	if err != nil {
		return prowapi.ProwJob{}, fmt.Errorf("error loading prow config: %w", err)
	}

	log.Debug("Generating prowjob specification to test.")
	job, pjSpecification, err := pjopts.genJobSpec(opt, conf, pjCfg)
	if err != nil {
		return prowapi.ProwJob{}, fmt.Errorf("failed generating prowjob specification to test: %w", err)
	}

	log.Debug("Adding pjtester labels to ProwJob.")
	// Add pjtester labels to ProwJob.
	job.Labels["created-by-pjtester"] = "true"
	job.Labels["prow.k8s.io/is-optional"] = "true"

	// Building ProwJob k8s resource based on generated job specifications.
	log.Debug("Generating ProwJob k8s resource.")
	pj := pjutil.NewProwJob(pjSpecification, job.Labels, job.Annotations)
	// Make sure prowjob to test will run on untrusted-workload cluster.
	pj.Spec.Cluster = "untrusted-workload"
	// Enable all reporting, otherwise send slack messages to null channel.
	if pjCfg.Report {
		pj.Spec.Report = true
	} else {
		pj.Spec.ReporterConfig = &prowapi.ReporterConfig{Slack: &prowapi.SlackReporterConfig{Channel: "kyma-prow-dev-null"}}
	}
	return pj, nil
}

// SchedulePJ will generate ProwJob for testing and create it on prow k8s cluster for execution.
func SchedulePJ(ghOptions *prowflagutil.GitHubOptions) {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(os.Stdout)
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.DebugLevel)
	var err error
	if err := checkEnvVars(envVarsList); err != nil {
		log.WithError(err).Fatalf("Required environment variable not set.")
	}

	// build common options for tests
	o, err := newCommonOptions(ghOptions)
	if err != nil {
		log.WithError(err).Fatal("Failed create common options")
	}
	testCfgFile := path.Join(o.clonePath, o.pjtesterPrOrg, o.pjtesterPrRepo, pjtesterConfigPath)
	// read pjtester.yaml file
	testCfg, err := readTestCfg(testCfgFile)
	if err != nil {
		log.WithError(err).Fatal("Failed read pjtester.yaml file")
	}

	// build Prow client
	prowClientSet, err := newProwK8sClientset()
	if err != nil {
		log.WithError(err).Fatal("Failed create Prow Clientset")
	}
	prowClient := prowClientSet.ProwV1()

	// build GitHub client
	o.githubClient = o.github

	// build git client
	o.gitOptions = git.ClientConfig{}
	o.gitClient, err = o.gitOptions.NewClient(git.WithGithubClient(o.githubClient))
	if err != nil {
		log.WithError(err).Fatal("Failed create git client")
	}

	// build prtagbuilder GitHub client
	o.prFinder = prtagbuilder.NewGitHubClient(nil)

	// Download pull requests details from GitHub. PR to download are provided in pjtester.yaml file.
	if testCfg.PrConfigs != nil {
		log.Debugf("Getting details of pull requests used in test prowjobs.")
		pullRequests, err := o.getPullRequests(testCfg.PrConfigs)
		if err != nil {
			log.WithError(err).Fatal("Failed get pull request details from GitHub.")
		}
		o.testPullRequests = pullRequests
	}
	if testCfg.PjConfigs.PrConfig != nil {
		log.Debugf("Getting details of pull requests with test prowjobs specification.")

		pullRequests, err := o.getPullRequests(testCfg.PjConfigs.PrConfig)
		if err != nil {
			log.WithError(err).Fatal("Failed get pull request details from GitHub.")
		}
		// There is only one element in each map level. This is enforced by testCfg struct values validation.
		for _, prorg := range pullRequests {
			for _, prconfig := range prorg {
				o.pjConfigPullRequest = prconfig
			}
		}
	}

	// Go over prowjob names to test and create ProwJob k8s resources for each.
	for orgName, pjOrg := range testCfg.PjConfigs.ProwJobs {
		for repoName, pjconfigs := range pjOrg {
			for _, pjconfig := range pjconfigs {
				testPjOpts := testProwJobOptions{
					repoName: repoName,
					orgName:  orgName,
					pjConfig: pjconfig,
				}
				// generate ProwJob k8s resource specification to test.
				pj, err := testPjOpts.newTestPJ(pjconfig, o)
				if err != nil {
					log.WithError(err).Fatalf("Failed schedule test of prowjob")
				}
				// create ProwJob on Prow k8s cluster.
				result, err := prowClient.ProwJobs(metav1.NamespaceDefault).Create(context.Background(), &pj, metav1.CreateOptions{})
				if err != nil {
					log.WithError(err).Fatalf("Failed schedule test of prowjob")
				}
				fmt.Printf("##########\nProwjob %s is %s\n##########\n", pj.Spec.Job, result.Status.State)
			}
		}
	}

}
