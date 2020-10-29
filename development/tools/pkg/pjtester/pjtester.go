package pjtester

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"k8s.io/test-infra/prow/config/secret"

	"github.com/go-yaml/yaml"
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

var (
	testCfgFile = fmt.Sprintf("%s/test-infra/vpath/pjtester.yaml", os.Getenv("KYMA_PROJECT_DIR"))
	envVarsList = []string{"KUBECONFIG_PATH", "KYMA_PROJECT_DIR", "PULL_BASE_REF", "PULL_BASE_SHA", "PULL_NUMBER", "PULL_PULL_SHA", "JOB_SPEC", "REPO_OWNER", "REPO_NAME"}
	log         = logrus.New()
)

// Default values for kyma-project/test-infra
const (
	defaultPjPath       = "test-infra/prow/jobs/"
	defaultConfigPath   = "test-infra/prow/config.yaml"
	defaultMasterBranch = "master"
)

// pjCfg holds prowjob to test name and path to it's definition.
type pjCfg struct {
	PjName string `yaml:"pjName"`
	PjPath string `yaml:"pjPath,omitempty"`
	Report bool   `yaml:"report,omitempty"`
}

// pjCfg holds number of PR to download and fetched details.
type prCfg struct {
	PrNumber    int `yaml:"prNumber"`
	pullRequest github.PullRequest
}

// prOrg holds pr configs per repository.
type prOrg map[string]prCfg

// testCfg holds prow config to test path, prowjobs to test names and paths to it's definitions.
type testCfg struct {
	PjNames    []pjCfg          `yaml:"pjNames"`
	ConfigPath string           `yaml:"configPath,omitempty"`
	PrConfigs  map[string]prOrg `yaml:"prConfigs,omitempty"`
}

// options holds data about prowjob and pull request to test.
type options struct {
	jobName       string
	configPath    string
	jobConfigPath string

	baseRef    string
	baseSha    string
	pullNumber int
	pullSha    string
	pullAuthor string
	org        string
	repo       string

	github       prowflagutil.GitHubOptions
	githubClient githubClient
	pullRequests map[string]prOrg
	prFetched    bool
}

type githubClient interface {
	GetPullRequest(org, repo string, number int) (*github.PullRequest, error)
	GetRef(org, repo, ref string) (string, error)
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
// It will set default path for prowjobs and config files if not provided in a file.
func readTestCfg(testCfgFile string) testCfg {
	var t testCfg
	yamlFile, err := ioutil.ReadFile(testCfgFile)
	if err != nil {
		log.Fatal("Failed read test config file from virtual path KYMA_PROJECT_DIR/test-infra/vpath/pjtester.yaml")
	}
	err = yaml.Unmarshal(yamlFile, &t)
	if err != nil {
		log.Fatal("Failed unmarshal test config yaml.")
	}
	if len(t.PjNames) == 0 {
		log.WithError(err).Fatalf("PjNames were not provided.")
	}
	for index, pj := range t.PjNames {
		if pj.PjName == "" {
			log.WithError(err).Fatalf("jobName to test was not provided.")
		}
		if pj.PjPath == "" {
			t.PjNames[index].PjPath = defaultPjPath
		}
	}
	if t.ConfigPath == "" {
		t.ConfigPath = defaultConfigPath
	}
	if len(t.PrConfigs) > 0 {
		for _, repo := range t.PrConfigs {
			if len(repo) > 0 {
				for _, pr := range repo {
					if pr.PrNumber == 0 {
						log.WithError(err).Fatalf("Pull request number for repo was not provided.")
					}
				}
			} else {
				log.WithError(err).Fatalf("Pull request number for repo was not provided.")
			}
		}
	}
	return t
}

// getPjCfg is adding prowjob details to the options for triggering prowjob test.
func getPjCfg(pjCfg pjCfg, o options) options {
	// jobName is a name of prowjob to test. It was read from pjtester.yaml file.
	o.jobName = pjCfg.PjName
	// jobConfigPath is a location of prow jobs config files to test. It was read from pjtester.yaml file or set to default.
	o.jobConfigPath = fmt.Sprintf("%s/%s", os.Getenv("KYMA_PROJECT_DIR"), pjCfg.PjPath)
	return o
}

// gatherOptions is building common options for all tests.
// Options are build from PR env variables and prowjob config read from pjtester.yaml file.
func gatherOptions(configPath string, ghOptions prowflagutil.GitHubOptions) options {
	var o options
	var err error
	o.github = ghOptions
	// configPath is a location of prow config file to test. It was read from pjtester.yaml file or set to default.
	o.configPath = fmt.Sprintf("%s/%s", os.Getenv("KYMA_PROJECT_DIR"), configPath)
	// baseRef is a base branch name for github pull request under test.
	o.baseRef = os.Getenv("PULL_BASE_REF")
	// baseSha is a git SHA of a base branch for github pull request under test
	o.baseSha = os.Getenv("PULL_BASE_SHA")
	// org is a name of organisation of pull request base branch
	o.org = os.Getenv("REPO_OWNER")
	// repo is a name of repository of pull request base branch
	o.repo = os.Getenv("REPO_NAME")
	// pullNumber is a number of github pull request under test
	o.pullNumber, err = strconv.Atoi(os.Getenv("PULL_NUMBER"))
	if err != nil {
		logrus.WithError(err).Fatalf("could not get pull number from env var PULL_NUMBER")
	}
	// pullSha is a SHA of github pull request head under test
	o.pullSha = os.Getenv("PULL_PULL_SHA")
	// pullAuthor is an author of github pull request under test
	o.pullAuthor = gjson.Get(os.Getenv("JOB_SPEC"), "refs.pulls.0.author").String()
	o.prFetched = false
	return o
}

// withGithubClientOptions will add default flags and values for github client.
func (o options) withGithubClientOptions() options {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	o.github.AddFlagsWithoutDefaultGitHubTokenPath(fs)
	_ = fs.Parse(os.Args[1:])
	if err := o.github.Validate(false); err != nil {
		logrus.WithError(err).Fatalf("github options validation failed")
	}
	return o
}

// getPullRequests will download pull requests details from github, for numbers provided in pjtester.yaml.
func (o *options) getPullRequests(t testCfg) {
	o.pullRequests = t.PrConfigs
	for org, repos := range t.PrConfigs {
		for repo, prcfg := range repos {
			pr, err := o.githubClient.GetPullRequest(org, repo, prcfg.PrNumber)
			if err != nil {
				logrus.WithError(err).Fatalf("failed to fetch PullRequest from GitHub: %v", err)
			}
			prcfg.pullRequest = *pr
			o.pullRequests[org][repo] = prcfg
		}
	}
	o.prFetched = true
}

// genJobSpec will generate job specifications for prowjob to test
// It will set test-infra pull request to test in Refs and move other repos refs to ExtraRefs.
// Refs will point to test-infra pull request to test head SHA.
func (o *options) genJobSpec(conf *config.Config, name string) (config.JobBase, prowapi.ProwJobSpec) {
	for fullRepoName, ps := range conf.PresubmitsStatic {
		org, repo, err := splitRepoName(fullRepoName)
		if err != nil {
			logrus.WithError(err).Warnf("Invalid repo name %s.", fullRepoName)
			continue
		}
		for _, p := range ps {
			if p.Name == o.jobName {
				pjs := pjutil.PresubmitSpec(p, prowapi.Refs{
					Org:  org,
					Repo: repo,
				})
				pjs = submitRefs(pjs, *o)
				return p.JobBase, pjs
			}
		}
	}
	for fullRepoName, ps := range conf.PostsubmitsStatic {
		org, repo, err := splitRepoName(fullRepoName)
		if err != nil {
			logrus.WithError(err).Warnf("invalid repo name %s", fullRepoName)
			continue
		}
		for _, p := range ps {
			if p.Name == o.jobName {
				pjs := pjutil.PostsubmitSpec(p, prowapi.Refs{
					Org:  org,
					Repo: repo,
				})
				pjs = submitRefs(pjs, *o)
				return p.JobBase, pjs
			}
		}
	}
	for _, p := range conf.Periodics {
		if p.Name == o.jobName {
			pjs := pjutil.PeriodicSpec(p)
			pjs = periodicRefs(pjs, *o)
			return p.JobBase, pjs
		}
	}
	return config.JobBase{}, prowapi.ProwJobSpec{}
}

// splitRepoName will extract org and repo names.
func splitRepoName(repo string) (string, string, error) {
	s := strings.SplitN(repo, "/", 2)
	if len(s) != 2 {
		return "", "", fmt.Errorf("repo %s cannot be split into org/repo", repo)
	}
	return s[0], s[1], nil
}

// setPrHeadSHA set pull request head SHA for provided refs.
func setPrHeadSHA(refs *prowapi.Refs, o options) {
	refs.BaseSHA = o.baseSha
	refs.BaseRef = o.baseRef
	refs.Pulls = []prowapi.Pull{{
		Author: o.pullAuthor,
		Number: o.pullNumber,
		SHA:    o.pullSha,
	}}
}

// matchRefPR will add pull request details to ExtraRefs.
func (o *options) matchRefPR(ref *prowapi.Refs) bool {
	if pr, present := o.pullRequests[ref.Org][ref.Repo]; present {
		ref.Pulls = []prowapi.Pull{{
			Author: pr.pullRequest.User.Login,
			Number: pr.PrNumber,
			SHA:    pr.pullRequest.Head.SHA,
		}}
		ref.BaseSHA = pr.pullRequest.Base.SHA
		ref.BaseRef = pr.pullRequest.Base.Ref
		return true
	}
	return false
}

// submitRefs build prowjob refs and extra refs according to tested PR refs and PRs refs from pjtester.yaml
// It ensure, refs for presubmit and postsubmit contain valid PR head SHA.
// If pjtester.yaml doesn't specify PR from other repo, test-infra refs from tested PR are used as prowjob refs.
// It adds PR refs to ExtraRefs for PR provided in pjtester.yaml.
func submitRefs(pjs prowapi.ProwJobSpec, opt options) prowapi.ProwJobSpec {
	// If prowjob specification refs point to test infra repo, add test-infra PR refs because we are going to test code from this PR.
	if pjs.Refs.Org == opt.org && pjs.Refs.Repo == opt.repo {
		setPrHeadSHA(pjs.Refs, opt)
		//Add PR details to ExtraRefs if PR number was provided in pjtester.yaml
		for index, ref := range pjs.ExtraRefs {
			matched := opt.matchRefPR(&ref)
			if matched {
				pjs.ExtraRefs[index] = ref
			}
		}
		return pjs
	}
	// If prowjob specification refs point to another repo.
	if pjs.Refs.Org != opt.org || pjs.Refs.Repo != opt.repo {
		//Check if PR number for prowjob specification refs was provided in pjtester.yaml.
		matched := opt.matchRefPR(pjs.Refs)
		if !matched {
			// If PR number not provided set BaseRef to master
			pjs.Refs.BaseRef = defaultMasterBranch
		}
		// Set PR refs for prowjob ExtraRefs if PR number provided in pjtester.yaml.
		for index, ref := range pjs.ExtraRefs {
			// If ExtraRefs ref points to test-infra, use refs from tested PR.
			if ref.Org == opt.org && ref.Repo == opt.repo {
				setPrHeadSHA(&ref, opt)
				// If for prowjob specification refs was provided PR number in pjtester.yaml, keep test-infra refs in ExtraRefs. Otherwise swap with current prowjob refs.
				if matched {
					pjs.ExtraRefs[index] = ref
				} else {
					refs := pjs.Refs
					pjs.Refs = &ref
					pjs.ExtraRefs[index] = *refs
				}
			} else {
				matched = opt.matchRefPR(&ref)
				if matched {
					pjs.ExtraRefs[index] = ref
				}
			}
		}
	}
	return pjs
}

// periodicRefs set pull request head SHA for test-infra extra refs.
// Periodics are not bound to any repo so there is no prowjob refs.
func periodicRefs(pjs prowapi.ProwJobSpec, opt options) prowapi.ProwJobSpec {
	for index, ref := range pjs.ExtraRefs {
		if ref.Org == opt.org && ref.Repo == opt.repo {
			setPrHeadSHA(&ref, opt)
			pjs.ExtraRefs[index] = ref
		} else {
			matched := opt.matchRefPR(&ref)
			if matched {
				pjs.ExtraRefs[index] = ref
			}
		}
	}
	return pjs
}

// formatPjName builds and formats testing prowjobname to match gcp cluster labels restrictions.
func formatPjName(pullAuthor, pjName string) string {
	fullName := fmt.Sprintf("%s_test_of_prowjob_%s", pullAuthor, pjName)
	formated := strings.ToLower(fullName)
	// Cut prowjob name to not exceed 63 kb.
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

// newTestPJ is building a prowjob definition for test
func newTestPJ(pjCfg pjCfg, opt options) prowapi.ProwJob {
	o := getPjCfg(pjCfg, opt)
	conf, err := config.Load(o.configPath, o.jobConfigPath)
	if err != nil {
		logrus.WithError(err).Fatal("Error loading prow config")
	}
	job, pjs := o.genJobSpec(conf, o.jobName)
	if job.Name == "" {
		logrus.Fatalf("Job %s not found.", o.jobName)
	}
	// Building prowjob based on generated job specifications.
	pj := pjutil.NewProwJob(pjs, job.Labels, job.Annotations)
	// Add prefix to prowjob to test name.
	pj.Spec.Job = formatPjName(opt.pullAuthor, pj.Spec.Job)
	// Make sure prowjob to test will run on untrusted-workload cluster.
	pj.Spec.Cluster = "untrusted-workload"
	if pjCfg.Report {
		pj.Spec.Report = true
	} else {
		pj.Spec.Report = false
	}
	return pj
}

// SchedulePJ will generate prowjob for testing and schedule it on prow for execution.
func SchedulePJ(ghOptions prowflagutil.GitHubOptions) {
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)
	var err error
	if err := checkEnvVars(envVarsList); err != nil {
		logrus.WithError(err).Fatalf("Required environment variable not set.")
	}
	testCfg := readTestCfg(testCfgFile)
	o := gatherOptions(testCfg.ConfigPath, ghOptions)
	prowClient := newProwK8sClientset()
	pjsClient := prowClient.ProwV1()
	var secretAgent *secret.Agent
	if o.github.TokenPath != "" {
		secretAgent = &secret.Agent{}
		if err := secretAgent.Start([]string{o.github.TokenPath}); err != nil {
			logrus.WithError(err).Fatal("Failed to start secret agent")
		}
	}
	o.githubClient, err = o.github.GitHubClient(secretAgent, false)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to get GitHub client")
	}
	var testPrCfg *map[string]prOrg
	if testPrCfg = &testCfg.PrConfigs; testPrCfg != nil && !o.prFetched {
		o.getPullRequests(testCfg)
	}
	for _, pjCfg := range testCfg.PjNames {
		pj := newTestPJ(pjCfg, o)
		result, err := pjsClient.ProwJobs(metav1.NamespaceDefault).Create(&pj)
		if err != nil {
			log.WithError(err).Fatalf("Failed schedule test of prowjob")
		}
		fmt.Printf("##########\nProwjob %s is %s\n##########\n", pj.Spec.Job, result.Status.State)
	}

}
