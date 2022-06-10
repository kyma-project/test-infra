package pjtester

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	// "log"
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
	defaultPjPath     = "prow/jobs/"
	defaultConfigPath = "prow/config.yaml"
	defaultMainBranch = "main"
	defaultClonePath  = "/home/prow/go/src/github.com"
)

var (
	// TODO: use test pjtesterv2.yaml file, change this to use a production pjtester.yaml
	testCfgFile = path.Join(defaultClonePath, "kyma-project", "test-infra/vpath/pjtesterv2.yaml")
	envVarsList = []string{"KUBECONFIG_PATH", "PULL_BASE_REF", "PULL_BASE_SHA", "PULL_NUMBER", "PULL_PULL_SHA", "JOB_SPEC", "REPO_OWNER", "REPO_NAME"}
	log         = logrus.New()
)

// pjCfg holds prowjob to test name and path to it's definition.
type pjConfig struct {
	PjName string `yaml:"pjName" validate:"required,min=1"`
	PjPath string `yaml:"pjPath" default:"test-infra/prow/jobs/"`
	Report bool   `yaml:"report,omitempty"`
}

// type someType struct {
//	PjCfgs []pjCfg `yaml:"pjCfgs" validate:"required,min=1"`
//	//PrConfig prCfg `yaml:"prConfig,omitempty"`
// }

type pjOrg map[string][]pjConfig

// pjCfg holds number of PR to download and fetched details.
type prConfig struct {
	PrNumber    int `yaml:"prNumber" validate:"required,number,min=1"`
	pullRequest github.PullRequest
}

// prOrg holds pr configs per repository.
type prOrg map[string]prConfig

// testCfg holds prow config to test path, prowjobs to test names and paths to it's definitions.
type testCfg struct {
	PjConfigs  map[string]pjOrg `yaml:"pjConfigs" validate:"required,min=1"`
	ConfigPath string           `yaml:"configPath" default:"test-infra/prow/config.yaml"`
	PrConfigs  map[string]prOrg `yaml:"prConfigs,omitempty"`
}

// options holds data about prowjob and pull request to test.
type options struct {
	// jobName       string
	configPath    string
	jobConfigPath string

	baseRef    string
	baseSha    string
	pullNumber int
	pullSha    string
	pullAuthor string
	org        string
	repo       string

	github       ghclient.GithubClientConfig
	githubClient *ghclient.GithubClient
	gitOptions   git.GitClientConfig
	gitClient    *git.GitClient
	prFinder     *prtagbuilder.GitHubClient
	pullRequests map[string]map[string]prConfig
}

// type githubClient interface {
//	GetPullRequest(org, repo string, number int) (*github.PullRequest, error)
//	GetRef(org, repo, ref string) (string, error)
// }

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
	// jobName is a name of prowjob to test. It was read from pjtester.yaml file.
	//	o.jobName = pjCfg.PjName
	// jobConfigPath is a location of prow jobs config files to test. It was read from pjtester.yaml file or set to default.
	if pjconfig.PjPath != "" {
		o.jobConfigPath = path.Join(defaultClonePath, org, repo, pjconfig.PjPath)
	} else if org == o.org && repo == o.repo && repo != "test-infra" {
		if prowDirInfo, err := os.Stat(path.Join(defaultClonePath, org, repo, ".prow")); !os.IsNotExist(err) && prowDirInfo.IsDir() {
			o.jobConfigPath = path.Join(defaultClonePath, org, repo, ".prow")
		} else {
			o.jobConfigPath = path.Join(defaultClonePath, org, repo, ".prow.yaml")
		}
	} else {
		o.jobConfigPath = path.Join(defaultClonePath, "kyma-project", "test-infra", defaultPjPath)
	}
}

// configPath is a location of prow config file to test. It was read from pjtester.yaml file or set to default.
func (o *options) setConfigPath(testConfig testCfg) {
	// If
	if testConfig.ConfigPath != "" {
		o.configPath = path.Join(defaultClonePath, o.org, o.repo, testConfig.ConfigPath)
	} else {
		o.configPath = path.Join(defaultClonePath, "kyma-project", "test-infra", defaultConfigPath)
	}

}

// gatherOptions is building common options for all tests.
// Options are build from PR env variables and prowjob config read from pjtester.yaml file.
// func gatherOptions(configPath string, ghOptions prowflagutil.GitHubOptions) options {
func gatherOptions(ghOptions prowflagutil.GitHubOptions) options {
	var o options
	var err error
	o.github = ghclient.GithubClientConfig{
		GitHubOptions: ghOptions,
		DryRun:        false,
	}
	// configPath is a location of prow config file to test. It was read from pjtester.yaml file or set to default.
	// o.configPath = fmt.Sprintf("%s/%s", os.Getenv("KYMA_PROJECT_DIR"), configPath)
	// o.configPath = path.Join(defaultClonePath, o.org, o.repo, configPath)
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

// getPullRequests will download details from GitHub for pull requests defined in pjtester test configuration.
// Downloaded pull request details are added to the options.pullRequest field.
func (o *options) getPullRequests(testcfg testCfg) error {
	if o.pullRequests == nil {
		o.pullRequests = make(map[string]map[string]prConfig)
	}
	for org, repos := range testcfg.PrConfigs {
		if _, ok := o.pullRequests[org]; !ok {
			o.pullRequests[org] = prOrg{}
		}
		for repo, prcfg := range repos {
			pr, err := o.githubClient.GetPullRequest(org, repo, prcfg.PrNumber)
			if err != nil {
				return fmt.Errorf("failed to fetch PullRequest from GitHub, error: %w", err)
			}
			prcfg.pullRequest = *pr
			o.pullRequests[org][repo] = prcfg
		}
	}
	return nil
}

// genJobSpec will generate job specifications for prowjob to test
// For presubmits it will find and download PR details for prowjob Refs, if the PR number for that repo was not provided in pjtester.yaml
// All test-infra refs will be set to pull request head SHA for which pjtester is triggered for.
// func (o *options) genJobSpec(pjCfg pjConfig, name, org, repo string) (config.JobBase, prowapi.ProwJobSpec, error) {
func (o *options) genJobSpec(pjCfg pjConfig, org, repo string) (config.JobBase, prowapi.ProwJobSpec, error) {
	var (
		preSubmits  []config.Presubmit
		postSubmits []config.Postsubmit
		err         error
	)
	baseSHAGetter := func() (string, error) {
		var err error
		baseSHA, err := o.githubClient.GetRef(o.org, o.repo, "heads/"+o.baseRef)
		if err != nil {
			return "", fmt.Errorf("failed to get baseSHA: %w", err)
		}
		return baseSHA, nil
	}

	headSHAGetter := func() (string, error) {
		return o.pullSha, nil
	}

	// jobConfigPath is a location of prow jobs config files to test. It was read from pjtester.yaml file or set to default.
	// opt.jobConfigPath = fmt.Sprintf("%s/%s", os.Getenv("KYMA_PROJECT_DIR"), pjCfg.PjPath)
	o.setJobConfigPath(pjCfg, org, repo)
	// Loading Prow config and Prow Jobs config from files. If files were changed in pull request, new values will be used for test.
	conf, err := config.Load(o.configPath, o.jobConfigPath, nil, "")
	if err != nil {
		return config.JobBase{}, prowapi.ProwJobSpec{}, fmt.Errorf("error loading prow config: %w", err)
	}

	// for fullRepoName, ps := range conf.PresubmitsStatic {
	//	org, repo, err := splitRepoName(fullRepoName)
	//	if err != nil {
	//		logrus.WithError(err).Warnf("Invalid repo name %s.", fullRepoName)
	//		continue
	//	}

	if o.org == org && o.repo == repo {
		// read pjspec from local files
		// pjspec must be in test-infra or .prow, both are already cloned
		// path to jobs were set to point to local fiels when read prowjob specs.
		preSubmits = conf.GetPresubmitsStatic(fmt.Sprintf("%s/%s", org, repo))
	} else {
		// test-infra from local extra-refs, inrepo from remote
		// test-infra already cloned
		preSubmits, err = conf.GetPresubmits(o.gitClient.ClientFactory, fmt.Sprintf("%s/%s", org, repo), baseSHAGetter, headSHAGetter)
	}
	for _, p := range preSubmits {
		// if p.Name == o.jobName {
		if p.Name == pjCfg.PjName {
			pjs := pjutil.PresubmitSpec(p, prowapi.Refs{
				Org:  org,
				Repo: repo,
			})
			pjs, err = presubmitRefs(pjs, *o)
			if err != nil {
				logrus.WithError(err).Fatalf("failed generate presubmit refs or extrarefs")
			}
			return p.JobBase, pjs, nil
		}
	}
	// }
	// for fullRepoName, ps := range conf.PostsubmitsStatic {
	//	org, repo, err := splitRepoName(fullRepoName)
	//	if err != nil {
	//		logrus.WithError(err).Warnf("invalid repo name %s", fullRepoName)
	//		continue
	//	}
	if o.org == org && o.repo == repo {
		// read pjspec from local files
		// pjspec must be in test-infra or .prow, both are already cloned
		postSubmits = conf.GetPostsubmitsStatic(fmt.Sprintf("%s/%s", org, repo))
	} else {
		// test-infra from local extra-refs, inrepo from remote
		// test-infra already cloned
		postSubmits, err = conf.GetPostsubmits(o.gitClient.ClientFactory, fmt.Sprintf("%s/%s", org, repo), baseSHAGetter, headSHAGetter)
	}
	for _, p := range postSubmits {
		// if p.Name == o.jobName {
		if p.Name == pjCfg.PjName {
			pjs := pjutil.PostsubmitSpec(p, prowapi.Refs{
				Org:  org,
				Repo: repo,
			})
			pjs, err = postsubmitRefs(pjs, *o)
			if err != nil {
				logrus.WithError(err).Fatalf("failed generate postsubmit refs and extrarefs")
			}
			return p.JobBase, pjs, nil
		}
	}
	// }

	// jobConfigPath is a location of prow jobs config files to test. It was read from pjtester.yaml file or set to default.
	// opt.jobConfigPath = fmt.Sprintf("%s/%s", os.Getenv("KYMA_PROJECT_DIR"), pjCfg.PjPath)
	// o.setJobConfigPath(pjCfg, org, repo)
	// Loading Prow config and Prow Jobs config from files. If files were changed in pull request, new values will be used for test.
	conf, err = config.Load(o.configPath, path.Join(defaultClonePath, "kyma-project", "test-infra", defaultPjPath), nil, "")
	if err != nil {
		return config.JobBase{}, prowapi.ProwJobSpec{}, fmt.Errorf("error loading prow config: %w", err)
	}

	for _, p := range conf.Periodics {
		// if p.Name == o.jobName {
		if p.Name == pjCfg.PjName {
			var err error
			pjs := pjutil.PeriodicSpec(p)
			pjs, err = periodicRefs(pjs, *o)
			if err != nil {
				logrus.WithError(err).Fatalf("failed generate periodic extrarefs")
			}
			return p.JobBase, pjs, nil
		}
	}
	return config.JobBase{}, prowapi.ProwJobSpec{}, fmt.Errorf("prowjob to test not found not found in prowjob specification files")
}

// splitRepoName will extract org and repo names.
// func splitRepoName(repo string) (string, string, error) {
//	s := strings.SplitN(repo, "/", 2)
//	if len(s) != 2 {
//		return "", "", fmt.Errorf("repo %s cannot be split into org/repo", repo)
//	}
//	return s[0], s[1], nil
// }

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

// presubmitRefs build prowjob refs and extrarefs according.
// It ensure, refs for test-infra is set to details of pull request from which pjtester was triggered.
// It ensures refs contains pull requests details for presubmit jobs.
// It ensures details of pull request numbers provided in pjtester.yaml are set for respecting refs or extra refs.
// TODO: You can't run pj against PR which is in other repo than pj is defined for.
func presubmitRefs(pjs prowapi.ProwJobSpec, opt options) (prowapi.ProwJobSpec, error) {
	// If prowjob specification refs point to test infra repo, add test-infra PR refs because we are going to test code from this PR.
	// TODO: check if PR number provided in pjtester.yaml prConfigs is for the same repository. If yes it should be used. You can have a prowjob def here byt want to test it against abother PR.
	if pjs.Refs.Org == opt.org && pjs.Refs.Repo == opt.repo {
		// set refs with details of tested PR
		setPrHeadSHA(pjs.Refs, opt)
		// Add PR details to ExtraRefs if PR number was provided in pjtester.yaml
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
		// Check if PR number for prowjob specification refs was provided in pjtester.yaml.
		// TODO: create dummy PR and run test against it, remove PR after test. This way test doesn't interfere with existing PR.
		if !opt.matchRefPR(pjs.Refs) {
			// If PR number not provided set BaseRef to main
			// TODO: create dummy PR and run test against it, remove PR after test. This way test doesn't interfere with existing PR.
			pjs.Refs.BaseRef = defaultMainBranch
			// get latest PR number for BaseRef branch and use it to set extra refs
			jobSpec := &downwardapi.JobSpec{Refs: pjs.Refs}
			branchPrAsString, err := prtagbuilder.BuildPrTag(jobSpec, true, true, opt.prFinder)
			if err != nil {
				fmt.Printf("level=info msg=failed get pr number for main branch head, using master\n")
				jobSpec.Refs.BaseRef = "master"
				branchPrAsString, err = prtagbuilder.BuildPrTag(jobSpec, true, true, opt.prFinder)
				if err != nil {
					return pjs, fmt.Errorf("could not get pr number for branch head, got error: %w", err)
				}
			}
			branchPR, err := strconv.Atoi(branchPrAsString)
			if err != nil {
				return pjs, fmt.Errorf("failed converting pr number string to integer, got error: %w", err)
			}
			// TODO: get pr manually and add it to the options.PullRequests field.
			err = opt.getPullRequests(testCfg{PrConfigs: map[string]prOrg{pjs.Refs.Org: {pjs.Refs.Repo: prConfig{PrNumber: branchPR}}}})
			if err != nil {
				return prowapi.ProwJobSpec{}, fmt.Errorf("failed get pull request deatils from GitHub, error: %w", err)
			}
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
		// Add PR details to ExtraRefs if PR number was provided in pjtester.yaml
		for index, ref := range pjs.ExtraRefs {
			if opt.matchRefPR(&ref) {
				pjs.ExtraRefs[index] = ref
			}
		}
		return pjs, nil
	}
	// If prowjob specification refs point to another repo.
	if pjs.Refs.Org != opt.org || pjs.Refs.Repo != opt.repo {
		// Check if PR number for prowjob specification refs was provided in pjtester.yaml.
		matched := opt.matchRefPR(pjs.Refs)
		if !matched {
			// If PR number not provided set BaseRef to main
			pjs.Refs.BaseRef = defaultMainBranch
			fakeJobSpec := &downwardapi.JobSpec{Refs: pjs.Refs}
			_, err := prtagbuilder.BuildPrTag(fakeJobSpec, true, true, opt.prFinder)
			if err != nil {
				fmt.Printf("level=info msg=failed get pr number for main branch head, using master\n")
				pjs.Refs.BaseRef = "master"
			}
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

// TODO: make it a method of pjConfig
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

// TODO: make it a method of pjConfig
// newTestPJ is building a prowjob definition to test prowjobs provided in pjtester test configuration.
func newTestPJ(pjCfg pjConfig, opt options, org, repo string) (prowapi.ProwJob, error) {
	// jobName is a name of prowjob to test. It was read from pjtester.yaml file.
	// opt.jobName = pjCfg.PjName
	// Building relative path from kyma-project directory.
	// jobConfigPath is a location of prow jobs config files to test. It was read from pjtester.yaml file or set to default.
	// opt.jobConfigPath = fmt.Sprintf("%s/%s", os.Getenv("KYMA_PROJECT_DIR"), pjCfg.PjPath)
	// opt.setJobConfigPath(pjCfg, org, repo)
	// Loading Prow config and Prow Jobs config from files. If files were changed in pull request, new values will be used for test.
	// conf, err := config.Load(opt.configPath, opt.jobConfigPath, nil, "")
	// if err != nil {
	//	return prowapi.ProwJob{}, fmt.Errorf("error loading prow config: %w", err)
	// }
	job, pjSpecification, err := opt.genJobSpec(pjCfg, org, repo)
	if err != nil {
		return prowapi.ProwJob{}, fmt.Errorf("failed generating prowjob specification to test: %w", err)
	}
	// if job.Name == "" {
	//	return prowapi.ProwJob{}, fmt.Errorf("job %s not found", opt.jobName)
	// }
	// Building prowjob based on generated job specifications.
	pj := pjutil.NewProwJob(pjSpecification, job.Labels, job.Annotations)
	// Add prefix to prowjob to test name.
	pj.Spec.Job = formatPjName(opt.pullAuthor, pj.Spec.Job)
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
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)
	var err error
	if err := checkEnvVars(envVarsList); err != nil {
		logrus.WithError(err).Fatalf("Required environment variable not set.")
	}
	testCfg, err := readTestCfg(testCfgFile)
	if err != nil {
		log.Fatal("Pjtester config validation failed.")
	}
	o := gatherOptions(ghOptions)
	o.setConfigPath(testCfg)
	prowClient := newProwK8sClientset()
	pjsClient := prowClient.ProwV1()
	ghc, err := o.github.NewGithubClient()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to get GitHub client")
	}
	o.githubClient = ghc
	o.gitOptions = git.GitClientConfig{}
	o.gitClient, err = o.gitOptions.NewGitClient(git.WithGithubClient(o.githubClient))
	if err != nil {
		logrus.WithError(err).Fatal("Failed to get git client")
	}
	// TODO: migrate to use test-infra/development/github/pkg/client
	o.prFinder = prtagbuilder.NewGitHubClient(nil)
	if &testCfg.PrConfigs != nil {
		err := o.getPullRequests(testCfg)
		if err != nil {
			log.WithError(err).Fatalf("Failed get pull request deatils from GitHub.")
		}
	}
	// Go over prowjob names to test and create prowjob definitions for each.
	for orgName, pjOrg := range testCfg.PjConfigs {
		for repoName, pjconfigs := range pjOrg {
			for _, pjconfig := range pjconfigs {
				pj, err := newTestPJ(pjconfig, o, orgName, repoName)
				if err != nil {
					log.WithError(err).Fatalf("Failed schedule test of prowjob")
				}
				result, err := pjsClient.ProwJobs(metav1.NamespaceDefault).Create(context.Background(), &pj, metav1.CreateOptions{})
				if err != nil {
					log.WithError(err).Fatalf("Failed schedule test of prowjob")
				}
				fmt.Printf("##########\nProwjob %s is %s\n##########\n", pj.Spec.Job, result.Status.State)
			}
		}
	}

}
