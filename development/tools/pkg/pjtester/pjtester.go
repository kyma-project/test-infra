package pjtester

import (
	"errors"
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
	"k8s.io/test-infra/prow/config/secret"
	prowflagutil "k8s.io/test-infra/prow/flagutil"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pjutil"
)

var (
	testCfgFile = fmt.Sprintf("%s/test-infra/vpath/pjtester.yaml", os.Getenv("KYMA_PROJECT_DIR"))
	//prowCfgPath = fmt.Sprintf("%s/test-infra/prow/config.yaml", os.Getenv("KYMA_PROJECT_DIR"))
	envVarsList = []string{"KUBECONFIG_PATH", "KYMA_PROJECT_DIR", "PULL_BASE_REF", "PULL_BASE_SHA", "PULL_NUMBER", "PULL_PULL_SHA", "JOB_SPEC"}
	log         = logrus.New()
)

const (
	defaultPjPath       = "test-infra/prow/jobs/"
	defaultConfigPath   = "test-infra/prow/config.yaml"
	defaultMasterBranch = "master"
)

type testCfg struct {
	PjName     string `yaml:"pjName"`
	PjPath     string `yaml:"pjPath,omitempty"`
	ConfigPath string `yaml:"configPath,omitempty"`
}

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
	pullRequest  *github.PullRequest
}

type githubClient interface {
	GetPullRequest(org, repo string, number int) (*github.PullRequest, error)
	GetRef(org, repo, ref string) (string, error)
}

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

func (o *options) Validate() error {
	if o.jobName == "" {
		return errors.New("jobName to test is not set")
	}

	if o.jobConfigPath == "" {
		return errors.New("jobPath to job to test is not set")
	}

	if err := o.github.Validate(false); err != nil {
		return err
	}

	return nil
}

func readTestCfg() *testCfg {
	var t *testCfg
	yamlFile, err := ioutil.ReadFile(testCfgFile)
	if err != nil {
		log.Fatal("Failed read test config file from virtual path KYMA_PROJECT_DIR/test-infra/vpath/pjtester.yaml")
	}
	err = yaml.Unmarshal(yamlFile, &t)
	if err != nil {
		log.Fatal("Failed unmarshal test config yaml.")
	}
	if t.PjPath == "" {
		t.PjPath = defaultPjPath
	}
	if t.ConfigPath == "" {
		t.ConfigPath = defaultConfigPath
	}
	return t
}

func gatherOptions(testCfg *testCfg) options {
	var o options
	var err error
	//fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	o.jobName = testCfg.PjName
	o.configPath = fmt.Sprintf("%s/%s", os.Getenv("KYMA_PROJECT_DIR"), testCfg.ConfigPath)
	o.jobConfigPath = fmt.Sprintf("%s/%s", os.Getenv("KYMA_PROJECT_DIR"), testCfg.PjPath)
	o.baseRef = os.Getenv("PULL_BASE_REF") // Git base ref under test
	o.baseSha = os.Getenv("PULL_BASE_SHA") // Git base SHA under test
	o.org = os.Getenv("REPO_OWNER")
	o.repo = os.Getenv("REPO_NAME")
	o.pullNumber, err = strconv.Atoi(os.Getenv("PULL_NUMBER"))
	if err != nil {
		logrus.WithError(err).Fatalf("Could not get pull number from env var PULL_NUMBER.")
	} // Git pull number under test
	o.pullSha = os.Getenv("PULL_PULL_SHA")                                          // Git pull SHA under test
	o.pullAuthor = gjson.Get(os.Getenv("JOB_SPEC"), "refs.pulls.0.author").String() // Git pull author under test")
	o.github = *prowflagutil.NewGitHubOptions()
	// TODO: remove after tests if not needed
	//defaultGitHubTokenPath := ""
	//if wantDefaultGitHubTokenPath {
	//	defaultGitHubTokenPath = "/etc/github/oauth"
	//}
	//fs.StringVar(&o.TokenPath, "github-token-path", defaultGitHubTokenPath, "Path to the file containing the GitHub OAuth secret.")
	//fs.StringVar(&o.deprecatedTokenFile, "github-token-file", "", "DEPRECATED: use -github-token-path instead.  -github-token-file may be removed anytime after 2019-01-01.")
	return o
}

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
					//BaseRef: o.baseRef,
					//BaseSHA: o.baseSha,
					//Pulls: []prowapi.Pull{{
					//	Author: o.pullAuthor,
					//	Number: o.pullNumber,
					//	SHA:    o.pullSha,
					//}},
				})
				return p.JobBase, presubmitPJRefs(pjs, *o)
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
				return p.JobBase, pjutil.PostsubmitSpec(p, prowapi.Refs{
					Org:  org,
					Repo: repo,
					//BaseRef: o.baseRef,
					//BaseSHA: o.baseSha,
				})
			}
		}
	}
	for _, p := range conf.Periodics {
		if p.Name == o.jobName {
			return p.JobBase, pjutil.PeriodicSpec(p)
		}
	}
	return config.JobBase{}, prowapi.ProwJobSpec{}
}

func splitRepoName(repo string) (string, string, error) {
	s := strings.SplitN(repo, "/", 2)
	if len(s) != 2 {
		return "", "", fmt.Errorf("repo %s cannot be split into org/repo", repo)
	}
	return s[0], s[1], nil
}

func presubmitPJRefs(pjs prowapi.ProwJobSpec, opt options) prowapi.ProwJobSpec {
	// If prowjob refs point to test infra repo, add refs details of this PR to prowjob refs because we are going to test code from this PR.
	if pjs.Refs.Org == opt.org && pjs.Refs.Repo == opt.repo {
		pjs.Refs.BaseSHA = opt.baseSha
		pjs.Refs.BaseRef = opt.baseRef
		pjs.Refs.Pulls = []prowapi.Pull{{
			Author: opt.pullAuthor,
			Number: opt.pullNumber,
			SHA:    opt.pullSha,
		}}
	}
	// If prowjob refs point to another repo, move refs to extra refs and set refs to details from this PR, because we are going to test code from this PR.
	//extraRefs := pjs.ExtraRefs
	refs := pjs.Refs
	if refs.Org != opt.org || refs.Repo != opt.repo {
		pjs.Refs.BaseRef = defaultMasterBranch
		for index, ref := range pjs.ExtraRefs {
			if ref.Org == opt.org && ref.Repo == opt.repo {
				pjs.ExtraRefs[index].BaseRef = opt.baseRef
				pjs.ExtraRefs[index].BaseSHA = opt.baseSha
				pjs.ExtraRefs[index].Pulls = []prowapi.Pull{{
					Author: opt.pullAuthor,
					Number: opt.pullNumber,
					SHA:    opt.pullSha,
				}}
			}
		}
	}
	return pjs
}

func newTestPJ() prowapi.ProwJob {
	testCfg := readTestCfg()
	o := gatherOptions(testCfg)
	if err := o.Validate(); err != nil {
		logrus.WithError(err).Fatalf("Missing required data.")
	}

	conf, err := config.Load(o.configPath, o.jobConfigPath)
	if err != nil {
		logrus.WithError(err).Fatal("Error loading prow config")
	}

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
	job, pjs := o.genJobSpec(conf, o.jobName)
	if job.Name == "" {
		logrus.Fatalf("Job %s not found.", o.jobName)
	}
	//o.org = pjs.Refs.Org
	//o.repo = pjs.Refs.Repo
	pj := pjutil.NewProwJob(pjs, job.Labels, job.Annotations)
	pj.Spec.Job = fmt.Sprintf("testing_of_prowjob_%s", pj.Spec.Job)
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
	pj := newTestPJ()
	result, err := pjsClient.ProwJobs(metav1.NamespaceDefault).Create(&pj)
	if err != nil {
		log.WithError(err).Fatalf("Failed schedule test of prowjob")
	}
	fmt.Printf("Test of prowjob %s is %s", pj.Spec.Job, result.Status.State)

}
