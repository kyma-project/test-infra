package pjtester

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"k8s.io/test-infra/prow/pod-utils/downwardapi"

	"k8s.io/test-infra/prow/config/secret"

	"github.com/go-yaml/yaml"
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
	prFinder     *prtagbuilder.GitHubClient
	pullRequests map[string]prOrg
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

// getPullRequests will download pull requests details from github.
func (o *options) getPullRequests(t testCfg) {
	if o.pullRequests == nil {
		o.pullRequests = make(map[string]prOrg)
	}
	for org, repos := range t.PrConfigs {
		if _, ok := o.pullRequests[org]; !ok {
			o.pullRequests[org] = prOrg{}
		}
		for repo, prcfg := range repos {
			pr, err := o.githubClient.GetPullRequest(org, repo, prcfg.PrNumber)
			if err != nil {
				logrus.WithError(err).Fatalf("failed to fetch PullRequest from GitHub")
			}
			prcfg.pullRequest = *pr
			o.pullRequests[org][repo] = prcfg
		}
	}
}

// genJobSpec will generate job specifications for prowjob to test
// For presubmits it will find and download PR details for prowjob Refs, if the PR number for that repo was not provided in pjtester.yaml
// All test-infra refs will be set to pull request head SHA for which pjtester is triggered for.
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
				pjs, err = presubmitRefs(pjs, *o)
				if err != nil {
					logrus.WithError(err).Fatalf("failed generate presubmit refs or extrarefs")
				}
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
				pjs, err = postsubmitRefs(pjs, *o)
				if err != nil {
					logrus.WithError(err).Fatalf("failed generate postsubmit refs and extrarefs")
				}
				return p.JobBase, pjs
			}
		}
	}
	for _, p := range conf.Periodics {
		if p.Name == o.jobName {
			var err error
			pjs := pjutil.PeriodicSpec(p)
			pjs, err = periodicRefs(pjs, *o)
			if err != nil {
				logrus.WithError(err).Fatalf("failed generate periodic extrarefs")
			}
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

// setPrHeadSHA set pull request head details for provided refs.
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

// submitRefs build prowjob refs and extrarefs according.
// It ensure, refs for test-infra is set to details of pull request fro which pjtester was triggered.
// It ensures refs contains pull requests details for presubmit jobs.
// It ensures details of pull request numbers provided in pjtester.yaml are set for respecting refs or extra refs.
func presubmitRefs(pjs prowapi.ProwJobSpec, opt options) (prowapi.ProwJobSpec, error) {
	// If prowjob specification refs point to test infra repo, add test-infra PR refs because we are going to test code from this PR.
	if pjs.Refs.Org == opt.org && pjs.Refs.Repo == opt.repo {
		// set refs with details of tested PR
		setPrHeadSHA(pjs.Refs, opt)
		//Add PR details to ExtraRefs if PR number was provided in pjtester.yaml
		for index, ref := range pjs.ExtraRefs {
			matched := opt.matchRefPR(&ref)
			if matched {
				pjs.ExtraRefs[index] = ref
			}
		}
		return pjs, nil
	}
	// If prowjob specification refs point to another repo.
	if pjs.Refs.Org != opt.org || pjs.Refs.Repo != opt.repo {
		//Check if PR number for prowjob specification refs was provided in pjtester.yaml.
		if !opt.matchRefPR(pjs.Refs) {
			// If PR number not provided set BaseRef to master
			pjs.Refs.BaseRef = defaultMasterBranch
			// get latest PR number for BaseRef branch and use it to set extra refs
			jobSpec := &downwardapi.JobSpec{Refs: pjs.Refs}
			branchPrAsString, err := prtagbuilder.BuildPrTag(jobSpec, true, true, opt.prFinder)
			if err != nil {
				return pjs, fmt.Errorf("could not get pr number for branch head, got error: %w", err)
			}
			branchPR, err := strconv.Atoi(branchPrAsString)
			if err != nil {
				return pjs, fmt.Errorf("failed converting pr number string to integer, got error: %w", err)
			}
			opt.getPullRequests(testCfg{PrConfigs: map[string]prOrg{pjs.Refs.Org: {pjs.Refs.Repo: prCfg{PrNumber: branchPR}}}})
			opt.matchRefPR(pjs.Refs)
		}
		// Set PR refs for prowjob ExtraRefs if PR number provided in pjtester.yaml.
		for index, ref := range pjs.ExtraRefs {
			// If ExtraRefs ref points to test-infra, use refs from tested PR.
			if ref.Org == opt.org && ref.Repo == opt.repo {
				setPrHeadSHA(&ref, opt)
				// If for prowjob specification refs was provided PR number in pjtester.yaml, keep test-infra refs in ExtraRefs. Otherwise swap with current prowjob refs.
				pjs.ExtraRefs[index] = ref
			} else {
				matchedExtraRef := opt.matchRefPR(&ref)
				if matchedExtraRef {
					pjs.ExtraRefs[index] = ref
				}
			}
		}
	}
	return pjs, nil
}

func postsubmitRefs(pjs prowapi.ProwJobSpec, opt options) (prowapi.ProwJobSpec, error) {
	// If prowjob specification refs point to test infra repo, add test-infra PR refs because we are going to test code from this PR.
	if pjs.Refs.Org == opt.org && pjs.Refs.Repo == opt.repo {
		setPrHeadSHA(pjs.Refs, opt)
		//Add PR details to ExtraRefs if PR number was provided in pjtester.yaml
		for index, ref := range pjs.ExtraRefs {
			if opt.matchRefPR(&ref) {
				pjs.ExtraRefs[index] = ref
			}
		}
		return pjs, nil
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
				pjs.ExtraRefs[index] = ref
			} else {
				matchedExtraRef := opt.matchRefPR(&ref)
				if matchedExtraRef {
					pjs.ExtraRefs[index] = ref
				}
			}
		}
	}
	return pjs, nil
}

// periodicRefs set pull request head SHA for test-infra extra refs.
// Periodics are not bound to any repo so there is no prowjob refs.
func periodicRefs(pjs prowapi.ProwJobSpec, opt options) (prowapi.ProwJobSpec, error) {
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
		pj.Spec.ReporterConfig = &prowapi.ReporterConfig{Slack: &prowapi.SlackReporterConfig{Channel: "kyma-prow-dev-null"}}
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
	o.prFinder = prtagbuilder.NewGitHubClient(nil)
	var testPrCfg *map[string]prOrg
	//if testPrCfg = &testCfg.PrConfigs; testPrCfg != nil && !o.prFetched {
	if testPrCfg = &testCfg.PrConfigs; testPrCfg != nil {
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
