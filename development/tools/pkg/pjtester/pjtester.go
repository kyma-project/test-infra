package pjtester

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/go-yaml/yaml"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	prowapi "k8s.io/test-infra/prow/apis/prowjobs/v1"
	prowclient "k8s.io/test-infra/prow/client/clientset/versioned"
	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/pjutil"
)

var (
	testCfgFile = fmt.Sprintf("%s/test-infra/vpath/pjtester.yaml", os.Getenv("KYMA_PROJECT_DIR"))
	envVarsList = []string{"KUBECONFIG_PATH", "KYMA_PROJECT_DIR", "PULL_BASE_REF", "PULL_BASE_SHA", "PULL_NUMBER", "PULL_PULL_SHA", "JOB_SPEC"}
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
}

// testCfg holds prow config to test path, prowjobs to test names and paths to it's definitions.
type testCfg struct {
	PjNames    []pjCfg `yaml:"pjNames"`
	ConfigPath string  `yaml:"configPath,omitempty"`
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
			return fmt.Errorf("Variable %s is not set", evar)
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

// readTestCfg read and validate data to test from pjtester.yaml file.
// It will set default location for prowjobs and config files location if not provided in a file.
func readTestCfg(testCfgFile string) testCfg {
	var t testCfg
	yamlFile, err := ioutil.ReadFile(testCfgFile)
	if err != nil {
		log.Fatal("Failed read test config file from virtual path KYMA_PROJECT_DIR/test-infra/vpath/pjtester.yaml")
	}
	err = yaml.Unmarshal(yamlFile, t)
	if err != nil {
		log.Fatal("Failed unmarshal test config yaml.")
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
	return t
}

// gatherOptions is building options of prowjob test.
// Options are build from PR env variables and prowjob config read from pjtester.yaml file.
func gatherOptions(pjCfg pjCfg, configPath string) options {
	var o options
	var err error
	// jobName is a name of prowjob to test. It was read from pjtester.yaml file.
	o.jobName = pjCfg.PjName
	// configPath is a location of prow config file to test. It was read from pjtester.yaml file or set to default.
	o.configPath = fmt.Sprintf("%s/%s", os.Getenv("KYMA_PROJECT_DIR"), configPath)
	// jobConfigPath is a location of prow jobs config files to test. It was read from pjtester.yaml file or set to default.
	o.jobConfigPath = fmt.Sprintf("%s/%s", os.Getenv("KYMA_PROJECT_DIR"), pjCfg.PjPath)
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
		logrus.WithError(err).Fatalf("Could not get pull number from env var PULL_NUMBER.")
	}
	// pullSha is a SHA of github pull request head under test
	o.pullSha = os.Getenv("PULL_PULL_SHA")
	// pullAuthor is an author of github pull request under test
	o.pullAuthor = gjson.Get(os.Getenv("JOB_SPEC"), "refs.pulls.0.author").String()
	return o
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
			logrus.WithError(err).Warnf("Invalid repo name %s.", fullRepoName)
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

// submitRefs set test-infra refs as prowjob refs.
// If refs points to other repo than test-infra it, will be moved to prowjob extra refs.
// Pull request head SHA is added to test-infra refs.
func submitRefs(pjs prowapi.ProwJobSpec, opt options) prowapi.ProwJobSpec {
	// If prowjob refs point to test infra repo, add PR to test head SHA to refs because we are going to test code from this PR.
	if pjs.Refs.Org == opt.org && pjs.Refs.Repo == opt.repo {
		setPrHeadSHA(pjs.Refs, opt)
		return pjs
	}
	// If prowjob refs point to another repo, move refs to extra refs and set refs to test-infra PR to test head SHA, because we are going to test code from this PR.
	// Because pjtester test changes in test-infra repository, it must be present in extra refs.
	if pjs.Refs.Org != opt.org || pjs.Refs.Repo != opt.repo {
		refs := pjs.Refs
		refs.BaseRef = defaultMasterBranch
		for index, ref := range pjs.ExtraRefs {
			if ref.Org == opt.org && ref.Repo == opt.repo {
				setPrHeadSHA(&ref, opt)
				pjs.Refs = &ref
				pjs.ExtraRefs[index] = *refs
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
		}
	}
	return pjs
}

// newTestPJ is building a prowjob definition for test
func newTestPJ(pjCfg pjCfg, configPath string) prowapi.ProwJob {
	o := gatherOptions(pjCfg, configPath)
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
	pj.Spec.Job = fmt.Sprintf("test_of_prowjob_%s", pj.Spec.Job)
	// Make sure prowjob to test will run on untrusted-workload cluster.
	pj.Spec.Cluster = "untrusted-workload"
	return pj
}

// SchedulePJ will generate prowjob for testing and schedule it on prow for execution.
func SchedulePJ() {
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)
	if err := checkEnvVars(envVarsList); err != nil {
		logrus.WithError(err).Fatalf("Required environment variable not set.")
	}
	prowClient := newProwK8sClientset()
	pjsClient := prowClient.ProwV1()
	testCfg := readTestCfg(testCfgFile)
	for _, pjCfg := range testCfg.PjNames {
		pj := newTestPJ(pjCfg, testCfg.ConfigPath)
		result, err := pjsClient.ProwJobs(metav1.NamespaceDefault).Create(&pj)
		if err != nil {
			log.WithError(err).Fatalf("Failed schedule test of prowjob")
		}
		fmt.Printf("Prowjob %s is %s", pj.Spec.Job, result.Status.State)
	}

}
