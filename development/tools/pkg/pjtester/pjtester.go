package pjtester

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	"k8s.io/test-infra/prow/pod-utils/downwardapi"

	"github.com/go-playground/validator/v10"
	"github.com/go-yaml/yaml"
	ghclient "github.com/kyma-project/test-infra/development/github/pkg/client/v2"
	"github.com/kyma-project/test-infra/development/github/pkg/git"
	"github.com/kyma-project/test-infra/development/tools/pkg/prtagbuilder"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	prowapi "k8s.io/test-infra/prow/apis/prowjobs/v1"
	prowclient "k8s.io/test-infra/prow/client/clientset/versioned"
	"k8s.io/test-infra/prow/config"
	prowflagutil "k8s.io/test-infra/prow/flagutil"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pjutil"
)

// Default values for kyma-project/test-infra
const (
	defaultPjPath      = "prow/jobs/"
	defaultConfigPath  = "prow/config.yaml"
	defaultMainBranch  = "main"
	defaultClonePath   = "/home/prow/go/src/github.com"
	testinfraOrg       = "kyma-project"
	testinfraRepo      = "test-infra"
	pjtesterConfigPath = "vpath/pjtesterv2.yaml"
)

var (
	envVarsList = []string{"KUBECONFIG_PATH", "PULL_BASE_REF", "PULL_BASE_SHA", "PULL_NUMBER", "PULL_PULL_SHA", "JOB_SPEC", "REPO_OWNER", "REPO_NAME"}
	log         = logrus.New()
)

// pjCfg holds prowjob to test name and path to it's definition.
type pjConfig struct {
	PjName string `yaml:"pjName" validate:"required,min=1"`
	PjPath string `yaml:"pjPath" default:"prow/jobs/"` // path relative to repository root
	Report bool   `yaml:"report,omitempty"`
}

type pjOrg map[string][]pjConfig

// pjCfg holds number of PR to download and fetched details.
type prConfig struct {
	PrNumber    int `yaml:"prNumber" validate:"required,number,min=1"`
	pullRequest github.PullRequest
	org         string
	repo        string
}

// prOrg holds pr configs per repository.
type prOrg map[string]prConfig

type pjConfigs struct {
	PrConfig map[string]prOrg `yaml:"prConfig,omitempty"`
	ProwJobs map[string]pjOrg `yaml:"prowJobs" validate:"required,min=1"`
}

// testCfg holds prow config to test path, prowjobs to test names and paths to it's definitions.
type testCfg struct {
	PjConfigs  pjConfigs        `yaml:"pjConfigs" validate:"required,min=1"`
	ConfigPath string           `yaml:"configPath" default:"prow/config.yaml"` // path relative to repository root
	PrConfigs  map[string]prOrg `yaml:"prConfigs,omitempty"`
}

// options holds data about prowjob and pull request to test.
type options struct {
	// jobName       string
	configPath    string
	jobConfigPath string

	baseRef        string
	baseSha        string
	pullNumber     int
	pullSha        string
	pullAuthor     string
	pjtesterPrOrg  string // pjtester pr org
	pjtesterPrRepo string // pjtester pr repo

	github              ghclient.GithubClientConfig
	githubClient        *ghclient.GithubClient
	gitOptions          git.ClientConfig
	gitClient           *git.Client
	prFinder            *prtagbuilder.GitHubClient
	testPullRequests    map[string]prOrg
	pjConfigPullRequest prConfig

	baseSHAGetter config.RefGetter
	headSHAGetter config.RefGetter

	usePjtesterPR bool
}

// checkEnvVars validate if required env variables are set
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

// newProwK8sClientset is building Prow client for kubeconfig location provided as env variable
func newProwK8sClientset() *prowclient.Clientset {
	k8sConfig, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG_PATH"))
	if err != nil {
		log.WithError(err).Fatalf("Failed create config for prow k8s clientset.")
	}
	clientset, err := prowclient.NewForConfig(k8sConfig)
	if err != nil {
		log.WithError(err).Fatalf("Failed create prow k8s clientset.")
	}
	return clientset
}

// readTestCfg read and validate data from pjtester.yaml file.
func readTestCfg(testCfgFile string) (testCfg, error) {
	var t testCfg
	yamlFile, err := ioutil.ReadFile(testCfgFile)
	if err != nil {
		log.Fatal("Failed read test config file from virtual path test-infra/vpath/pjtester.yaml")
	}
	err = yaml.Unmarshal(yamlFile, &t)
	if err != nil {
		log.Fatal("Failed unmarshal test config yaml.")
	}

	// TODO: provide correct and full validation tags.
	validate := validator.New()
	// returns nil or ValidationErrors ( []FieldError )
	err = validate.Struct(t)
	if err != nil {

		// this check is only needed when your code could produce
		// an invalid value for validation such as interface with nil
		// value most including myself do not usually have code like this.
		if _, ok := err.(*validator.InvalidValidationError); ok {
			fmt.Println(err)
			return testCfg{}, err
		}

		for _, err := range err.(validator.ValidationErrors) {

			fmt.Println(err.Namespace())
			fmt.Println(err.Field())
			fmt.Println(err.StructNamespace())
			fmt.Println(err.StructField())
			fmt.Println(err.Tag())
			fmt.Println(err.ActualTag())
			fmt.Println(err.Kind())
			fmt.Println(err.Type())
			fmt.Println(err.Value())
			fmt.Println(err.Param())
			fmt.Println()
		}

		// from here you can create your own error messages in whatever language you wish
		return testCfg{}, err
	}

	return t, nil
}

// getPjCfg is adding prowjob details to the options for triggering prowjob test.
func (o *options) setJobConfigPath(pjconfig pjConfig, org, repo string) {
	// jobConfigPath is a location of prow jobs config files to test. It was read from pjtester.yaml file or set to default.
	if pjconfig.PjPath != "" {
		o.jobConfigPath = path.Join(defaultClonePath, org, repo, pjconfig.PjPath)
	} else {
		o.jobConfigPath = path.Join(defaultClonePath, testinfraOrg, testinfraRepo, defaultPjPath)
	}
}

// configPath is a location of prow config file to test. It was read from pjtester.yaml file or set to default.
func (o *options) setProwConfigPath(testConfig testCfg) {
	// If
	if testConfig.ConfigPath != "" {
		o.configPath = path.Join(defaultClonePath, o.pjtesterPrOrg, o.pjtesterPrRepo, testConfig.ConfigPath)
	} else {
		o.configPath = path.Join(defaultClonePath, testinfraOrg, testinfraRepo, defaultConfigPath)
	}

}

// newCommonOptions is building common options for all tests.
// Options are build from PR env variables and prowjob config read from pjtester.yaml file.
// func newCommonOptions(configPath string, ghOptions prowflagutil.GitHubOptions) options {
func newCommonOptions(ghOptions prowflagutil.GitHubOptions) options {
	var o options
	var err error
	o.github = ghclient.GithubClientConfig{
		GitHubOptions: ghOptions,
		DryRun:        false,
	}
	// baseRef is a base branch name for github pull request under test.
	o.baseRef = os.Getenv("PULL_BASE_REF")
	// baseSha is a git SHA of a base branch for github pull request under test
	o.baseSha = os.Getenv("PULL_BASE_SHA")
	// pjtesterPrOrg is a name of organisation of pull request base branch
	o.pjtesterPrOrg = os.Getenv("REPO_OWNER")
	// repo is a name of repository of pull request base branch
	o.pjtesterPrRepo = os.Getenv("REPO_NAME")
	// pullNumber is a number of github pull request under test
	o.pullNumber, err = strconv.Atoi(os.Getenv("PULL_NUMBER"))
	if err != nil {
		log.WithError(err).Fatalf("could not get pull number from env var PULL_NUMBER")
	}
	// pullSha is a SHA of github pull request head under test
	o.pullSha = os.Getenv("PULL_PULL_SHA")
	// pullAuthor is an author of github pull request under test
	o.pullAuthor = gjson.Get(os.Getenv("JOB_SPEC"), "refs.pulls.0.author").String()
	return o
}

// getPullRequests will download details from GitHub for pull requests defined in pjtester test configuration.
// Downloaded pull request details are added to the options.pullRequest field.
func (o *options) getPullRequests(prconfig map[string]prOrg) (map[string]prOrg, error) {
	prs := make(map[string]prOrg)
	for org, repos := range prconfig {
		if _, ok := prs[org]; !ok {
			prs[org] = prOrg{}
		}
		for repo, prcfg := range repos {
			pr, err := o.githubClient.GetPullRequest(org, repo, prcfg.PrNumber)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch PullRequest from GitHub, error: %w", err)
			}
			prcfg.pullRequest = *pr
			prcfg.org = org
			prcfg.repo = repo
			prs[org][repo] = prcfg
		}
	}
	return prs, nil
}

// genJobSpec will generate job specifications for prowjob to test
// For presubmits it will find and download PR details for prowjob Refs, if the PR number for that repo was not provided in pjtester.yaml
// All test-infra refs will be set to pull request head SHA for which pjtester is triggered for.
func (o *options) genJobSpec(pjCfg pjConfig, org, repo string) (config.JobBase, prowapi.ProwJobSpec, error) {
	var (
		preSubmits  []config.Presubmit
		postSubmits []config.Postsubmit
		err         error
	)

	log.Debugf("pjtesterPR: %v", o.testPullRequests)
	if _, present := o.testPullRequests[o.pjtesterPrOrg][o.pjtesterPrRepo]; present {
		o.usePjtesterPR = false
		log.Debugf("not using pjtester PR: %v", o.usePjtesterPR)
	} else {
		o.usePjtesterPR = true
		log.Debugf("using pjtester PR: %v", o.usePjtesterPR)
	}

	// Loading Prow config and Prow Jobs config from files. If files were changed in pull request, new values will be used for test.
	conf, err := config.Load(o.configPath, o.jobConfigPath, nil, "")
	if err != nil {
		return config.JobBase{}, prowapi.ProwJobSpec{}, fmt.Errorf("error loading prow config: %w", err)
	}

	if o.headSHAGetter != nil {
		preSubmits, err = conf.GetPresubmits(o.gitClient.ClientFactory, fmt.Sprintf("%s/%s", org, repo), o.baseSHAGetter, o.headSHAGetter)
		if err != nil {
			log.WithError(err).Fatalf("failed get presubmits")
		}
		log.Debugf("Use head getter: %v", o.headSHAGetter)
	} else {
		log.Debugf("fetching presubmits")
		preSubmits, err = conf.GetPresubmits(o.gitClient.ClientFactory, fmt.Sprintf("%s/%s", org, repo), o.baseSHAGetter)
		if err != nil {
			log.WithError(err).Fatalf("failed get presubmits")
		}
		log.Debugf("Not use head getter")
		log.Debugf("presubmits count: %d", len(preSubmits))
	}
	log.Debugf("pjconfig pjname: %s", pjCfg.PjName)
	for _, p := range preSubmits {
		log.Debugf("presubmit.name : %s", p.Name)
		if p.Name == pjCfg.PjName {
			p.Optional = true
			pjs := pjutil.PresubmitSpec(p, prowapi.Refs{
				Org:  org,
				Repo: repo,
			})
			pjs, err = presubmitRefs(pjs, *o)
			if err != nil {
				log.WithError(err).Fatalf("failed generate presubmit refs or extrarefs")
			}
			return p.JobBase, pjs, nil
		}
	}
	if o.headSHAGetter != nil {
		postSubmits, err = conf.GetPostsubmits(o.gitClient.ClientFactory, fmt.Sprintf("%s/%s", org, repo), o.baseSHAGetter, o.headSHAGetter)
		if err != nil {
			log.WithError(err).Fatalf("failed get postsubmits")
		}
	} else {
		postSubmits, err = conf.GetPostsubmits(o.gitClient.ClientFactory, fmt.Sprintf("%s/%s", org, repo), o.baseSHAGetter)
		if err != nil {
			log.WithError(err).Fatalf("failed get postsubmits")
		}
	}
	for _, p := range postSubmits {
		if p.Name == pjCfg.PjName {
			pjs := pjutil.PostsubmitSpec(p, prowapi.Refs{
				Org:  org,
				Repo: repo,
			})
			pjs, err = postsubmitRefs(pjs, *o)
			if err != nil {
				log.WithError(err).Fatalf("failed generate postsubmit refs and extrarefs")
			}
			return p.JobBase, pjs, nil
		}
	}

	for _, p := range conf.Periodics {
		if p.Name == pjCfg.PjName {
			var err error
			pjs := pjutil.PeriodicSpec(p)
			pjs, err = periodicRefs(pjs, *o)
			if err != nil {
				log.WithError(err).Fatalf("failed generate periodic extrarefs")
			}
			return p.JobBase, pjs, nil
		}
	}
	return config.JobBase{}, prowapi.ProwJobSpec{}, fmt.Errorf("prowjob to test not found in prowjob specification files")
}

// setPrHeadSHA set pull request head details for provided refs.
func setRefs(ref *prowapi.Refs, baseSHA, baseRef, pullAuthor, pullSHA string, pullNumber int) {
	ref.BaseSHA = baseSHA
	ref.BaseRef = baseRef
	ref.Pulls = []prowapi.Pull{{
		Author: pullAuthor,
		Number: pullNumber,
		SHA:    pullSHA,
	}}
}

func setRefsFromCurrentPR(pjs prowapi.ProwJobSpec, opt options) (prowapi.ProwJobSpec, bool) {
	pjsRefsSet := false
	if pjs.Refs.Org == opt.pjtesterPrOrg && pjs.Refs.Repo == opt.pjtesterPrRepo {
		// set refs with details of tested PR
		setRefs(pjs.Refs, opt.baseSha, opt.baseRef, opt.pullAuthor, opt.pullSha, opt.pullNumber)
		pjsRefsSet = true
		// TODO: extra setting extrarefs to separate function
	} else {
		for index, ref := range pjs.ExtraRefs {
			if ref.Org == opt.pjtesterPrOrg && ref.Repo == opt.pjtesterPrRepo {
				setRefs(&ref, opt.baseSha, opt.baseRef, opt.pullAuthor, opt.pullSha, opt.pullNumber)
				pjs.ExtraRefs[index] = ref
			}
		}
	}
	return pjs, pjsRefsSet
}

func setRefsFromConfigs(pjs prowapi.ProwJobSpec, opt options) (prowapi.ProwJobSpec, bool) {
	pjsRefsSet := false
	if pr, present := opt.testPullRequests[pjs.Refs.Org][pjs.Refs.Repo]; present {
		setRefs(pjs.Refs, pr.pullRequest.Base.SHA, pr.pullRequest.Base.Ref, pr.pullRequest.AuthorAssociation, pr.pullRequest.Head.SHA, pr.pullRequest.Number)
		pjsRefsSet = true
	}
	// TODO: extra setting extrarefs to separate function
	for index, ref := range pjs.ExtraRefs {
		// Add PR details to ExtraRefs if PR number was provided in pjtester.yaml
		if pr, present := opt.testPullRequests[ref.Org][ref.Repo]; present {
			setRefs(&ref, pr.pullRequest.Base.SHA, pr.pullRequest.Base.Ref, pr.pullRequest.AuthorAssociation, pr.pullRequest.Head.SHA, pr.pullRequest.Number)
			pjs.ExtraRefs[index] = ref
		}
	}
	return pjs, pjsRefsSet
}

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

// presubmitRefs build prowjob refs and extrarefs according.
// It ensure, refs for test-infra is set to details of pull request from which pjtester was triggered.
// It ensures refs contains pull requests details for presubmit jobs.
// It ensures details of pull request numbers provided in pjtester.yaml are set for respecting refs or extra refs.
// TODO: You can't run pj against PR which is in other repo than pj is defined for.
func presubmitRefs(pjs prowapi.ProwJobSpec, opt options) (prowapi.ProwJobSpec, error) {
	pjsRefsSet := false
	// If prowjob specification refs point to repo with pjtester,
	// and if pjstester config doesn't have prConfig for the same repo as prowjob specification refs,
	// use pjtester PR refs as prowjob specification refs because we are going to test code from this PR.
	// TODO: check if PR number provided in pjtester.yaml prConfigs is for the same repository. If yes it should be used. You can have a prowjob def here byt want to test it against abother PR.
	if opt.usePjtesterPR {
		pjs, pjsRefsSet = setRefsFromCurrentPR(pjs, opt)
	}
	pjs, pjsRefsSet = setRefsFromConfigs(pjs, opt)
	if !pjsRefsSet {
		pjs.Refs.BaseRef = defaultMainBranch
		branchPR, err := findLatestPR(pjs, opt)
		if err != nil {
			pjs.Refs.BaseRef = "master"
			branchPR, err = findLatestPR(pjs, opt)
			if err != nil {
				return prowapi.ProwJobSpec{}, fmt.Errorf("unknown default branch name, got error: %w", err)
			}
		}
		pr, err := opt.githubClient.GetPullRequest(pjs.Refs.Org, pjs.Refs.Repo, branchPR)
		if err != nil {
			return prowapi.ProwJobSpec{}, fmt.Errorf("failed get pull request deatils from GitHub, error: %w", err)
		}
		setRefs(pjs.Refs, pr.Base.SHA, pr.Base.Ref, pr.AuthorAssociation, pr.Head.SHA, pr.Number)
	}
	return pjs, nil
}

func postsubmitRefs(pjs prowapi.ProwJobSpec, opt options) (prowapi.ProwJobSpec, error) {
	pjsRefsSet := false
	// If prowjob specification refs point to repo with pjtester,
	// and if pjstester config doesn't have prConfig for the same repo as prowjob specification refs,
	// use pjtester PR refs as prowjob specification refs because we are going to test code from this PR.
	// TODO: check if PR number provided in pjtester.yaml prConfigs is for the same repository. If yes it should be used. You can have a prowjob def here byt want to test it against abother PR.
	if opt.usePjtesterPR {
		pjs, pjsRefsSet = setRefsFromCurrentPR(pjs, opt)
	}
	pjs, pjsRefsSet = setRefsFromConfigs(pjs, opt)
	if !pjsRefsSet {
		pjs.Refs.BaseRef = defaultMainBranch
		_, err := findLatestPR(pjs, opt)
		if err != nil {
			pjs.Refs.BaseRef = "master"
			_, err = findLatestPR(pjs, opt)
			if err != nil {
				return prowapi.ProwJobSpec{}, fmt.Errorf("unknown default branch name, got error: %w", err)
			}
		}

	}
	return pjs, nil
}

// periodicRefs set pull request head SHA for test-infra extra refs.
// Periodics are not bound to any repo so there is no prowjob refs.
func periodicRefs(pjs prowapi.ProwJobSpec, opt options) (prowapi.ProwJobSpec, error) {
	for index, ref := range pjs.ExtraRefs {
		if opt.usePjtesterPR {
			if ref.Org == opt.pjtesterPrOrg && ref.Repo == opt.pjtesterPrRepo {
				setRefs(&ref, opt.baseSha, opt.baseRef, opt.pullAuthor, opt.pullSha, opt.pullNumber)
				pjs.ExtraRefs[index] = ref
			}
		} else {
			if pr, present := opt.testPullRequests[ref.Org][ref.Repo]; present {
				setRefs(&ref, pr.pullRequest.Base.SHA, pr.pullRequest.Base.Ref, pr.pullRequest.AuthorAssociation, pr.pullRequest.Head.SHA, pr.pullRequest.Number)
				pjs.ExtraRefs[index] = ref
			}
		}
	}
	return pjs, nil
}

// formatPjName builds and formats testing prowjobname to match gcp cluster labels restrictions.
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

func (o *options) checkoutTestInfra() error {
	repoclient, _, err := o.gitClient.GetGitRepoClientFromDir(testinfraOrg, testinfraRepo, path.Join(defaultClonePath, testinfraOrg, testinfraRepo))
	if err != nil {
		return fmt.Errorf("failed get git client for repository from local directory %s, error: %w", path.Join(defaultClonePath, testinfraOrg, testinfraRepo), err)
	}
	err = repoclient.CheckoutPullRequest(o.pjConfigPullRequest.PrNumber)
	if err != nil {
		return fmt.Errorf("failed checkout pull request %s, error: %w", fmt.Sprintf("%s/%s#%d", testinfraOrg, testinfraRepo, o.pjConfigPullRequest.PrNumber), err)
	}
	return nil
}

func (o *options) setRefsGetters(currentPrOrg, currentPrRepo string) error {
	// Test prowjob from pull request in test-infra. Checkout this PR in test-infra repo from extraRefs.
	if o.pjConfigPullRequest.pullRequest.Number != 0 {
		if o.pjConfigPullRequest.org == testinfraOrg && o.pjConfigPullRequest.repo == testinfraRepo {
			err := o.checkoutTestInfra()
			if err != nil {
				return fmt.Errorf("")
			}
		} else {
			o.baseSHAGetter = func() (string, error) {
				return o.pjConfigPullRequest.pullRequest.Base.SHA, nil
			}

			o.headSHAGetter = func() (string, error) {
				return o.pjConfigPullRequest.pullRequest.Head.SHA, nil
			}
			return nil
		}
	}
	o.baseSHAGetter = func() (string, error) {
		var err error
		baseSHA, err := o.githubClient.GetRef(currentPrOrg, currentPrRepo, "heads/main")
		if err != nil {
			return "", fmt.Errorf("failed to get baseSHA: %w", err)
		}
		return baseSHA, nil
	}
	o.headSHAGetter = nil
	return nil
}

// newTestPJ is building a prowjob definition to test prowjobs provided in pjtester test configuration.
func newTestPJ(pjCfg pjConfig, opt options, org, repo string) (prowapi.ProwJob, error) {
	err := opt.setRefsGetters(org, repo)
	if err != nil {
		return prowapi.ProwJob{}, fmt.Errorf("failed set RefsGetters, error: %w", err)
	}
	opt.setJobConfigPath(pjCfg, org, repo)
	log.Debugf("job path: %s", opt.jobConfigPath)
	_, pjSpecification, err := opt.genJobSpec(pjCfg, org, repo)
	if err != nil {
		return prowapi.ProwJob{}, fmt.Errorf("failed generating prowjob specification to test: %w", err)
	}
	// Building prowjob based on generated job specifications.
	pj := pjutil.NewProwJob(pjSpecification, map[string]string{"created-by-pjtester": "true", "prow.k8s.io/is-optional": "true"}, map[string]string{})
	// Add prefix to prowjob to test name.
	pj.Spec.Job = formatPjName(opt.pullAuthor, pj.Spec.Job)
	// Add prefix to prowjob to test context.
	pj.Spec.Context = pj.Spec.Job
	// Make sure prowjob to test will run on untrusted-workload cluster.
	pj.Spec.Cluster = "untrusted-workload"
	if pjCfg.Report {
		pj.Spec.Report = true
	} else {
		pj.Spec.ReporterConfig = &prowapi.ReporterConfig{Slack: &prowapi.SlackReporterConfig{Channel: "kyma-prow-dev-null"}}
	}
	return pj, nil
}

// SchedulePJ will generate prowjob for testing and schedule it on prow for execution.
func SchedulePJ(ghOptions prowflagutil.GitHubOptions) {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(os.Stdout)
	// TODO: use our logging clients
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.DebugLevel)
	var err error
	if err := checkEnvVars(envVarsList); err != nil {
		log.WithError(err).Fatalf("Required environment variable not set.")
	}
	o := newCommonOptions(ghOptions)
	// TODO: use test pjtesterv2.yaml file, change this to use a production pjtester.yaml
	testCfgFile := path.Join(defaultClonePath, o.pjtesterPrOrg, o.pjtesterPrRepo, "vpath/pjtesterv2.yaml")
	// read pjtester.yaml file
	testCfg, err := readTestCfg(testCfgFile)
	if err != nil {
		log.Fatal("Pjtester config validation failed.")
	}
	// configPath is a location of prow config file to test. It was read from pjtester.yaml file or set to default.
	o.setProwConfigPath(testCfg)
	prowClientSet := newProwK8sClientset()
	prowClient := prowClientSet.ProwV1()
	ghc, err := o.github.NewGithubClient()
	if err != nil {
		log.WithError(err).Fatal("Failed to get GitHub client")
	}
	o.githubClient = ghc
	o.gitOptions = git.ClientConfig{}
	o.gitClient, err = o.gitOptions.NewClient(git.WithGithubClient(o.githubClient))
	if err != nil {
		log.WithError(err).Fatal("Failed to get git client")
	}
	// TODO: migrate to use test-infra/development/github/pkg/client
	o.prFinder = prtagbuilder.NewGitHubClient(nil)

	log.Debugf("prconfigs: %v", &testCfg.PrConfigs)
	if &testCfg.PrConfigs != nil {
		log.Debugf("getting details of pull requests for tested prowjobs")
		pullRequests, err := o.getPullRequests(testCfg.PrConfigs)
		if err != nil {
			log.WithError(err).Fatalf("Failed get pull request deatils from GitHub.")
		}
		o.testPullRequests = pullRequests
	}
	if &testCfg.PjConfigs.PrConfig != nil {
		pullRequests, err := o.getPullRequests(testCfg.PjConfigs.PrConfig)
		if err != nil {
			log.WithError(err).Fatalf("Failed get pull request deatils from GitHub.")
		}
		for _, prorg := range pullRequests {
			for _, prconfig := range prorg {
				o.pjConfigPullRequest = prconfig
			}
		}
	}

	// Go over prowjob names to test and create prowjob definitions for each.
	for orgName, pjOrg := range testCfg.PjConfigs.ProwJobs {
		for repoName, pjconfigs := range pjOrg {
			for _, pjconfig := range pjconfigs {
				// generate prowjob specification to test.
				pj, err := newTestPJ(pjconfig, o, orgName, repoName)
				if err != nil {
					log.WithError(err).Fatalf("Failed schedule test of prowjob")
				}
				result, err := prowClient.ProwJobs(metav1.NamespaceDefault).Create(context.Background(), &pj, metav1.CreateOptions{})
				if err != nil {
					log.WithError(err).Fatalf("Failed schedule test of prowjob")
				}
				fmt.Printf("##########\nProwjob %s is %s\n##########\n", pj.Spec.Job, result.Status.State)
			}
		}
	}

}
