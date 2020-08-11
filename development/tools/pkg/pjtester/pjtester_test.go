package pjtester

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	prowapi "k8s.io/test-infra/prow/apis/prowjobs/v1"
)

var (
	prNumber      int
	prAuthor      string
	otherPrNumber int
	otherPrAuthor string
	otherPrSHA    string
	otherPrOrg    string
	otherPrRepo   string
)

func TestMain(m *testing.M) {
	pjPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	prAuthor = "testAuthor"
	otherPrAuthor = "otherPrAuthor"
	otherPrSHA = "otherPrHeadSHA"
	otherPrNumber = 1212
	otherPrOrg = "kyma-project"
	otherPrRepo = "kyma"
	os.Setenv("KYMA_PROJECT_DIR", filepath.Clean(fmt.Sprintf("%s/../../../../..", pjPath)))
	os.Setenv("PULL_BASE_REF", "master")
	os.Setenv("PULL_BASE_SHA", "masterBranchSHA")
	os.Setenv("PULL_NUMBER", "12345")
	os.Setenv("PULL_PULL_SHA", "prHeadSHA")
	os.Setenv("REPO_OWNER", "kyma-project")
	os.Setenv("REPO_NAME", "test-infra")
	os.Setenv("JOB_SPEC", fmt.Sprintf("{\"type\":\"presubmit\",\"job\":\"job-name\",\"buildid\":\"0\",\"prowjobid\":\"uuid\",\"refs\":{\"org\":\"org-name\",\"repo\":\"repo-name\",\"base_ref\":\"base-ref\",\"base_sha\":\"base-sha\",\"pulls\":[{\"number\":1,\"author\":\"%s\",\"sha\":\"pull-sha\"}]}}", prAuthor))
	prNumber, _ = strconv.Atoi(os.Getenv("PULL_NUMBER"))
	testCfgFile = fmt.Sprintf("%s/test-infra/development/tools/pkg/pjtester/test_artifacts/pjtester.yaml", os.Getenv("KYMA_PROJECT_DIR"))

	os.Exit(m.Run())
}

func TestReadTestCfg(t *testing.T) {
	testCfg := readTestCfg(testCfgFile)
	assert.Containsf(t, testCfg.PjNames, pjCfg{
		PjName: "presubmit-test-job",
		PjPath: "test-infra/development/tools/pkg/pjtester/test_artifacts/",
		Report: true,
	}, "Unexpected value of first prowjob to test name.")
	assert.Containsf(t, testCfg.PjNames, pjCfg{
		PjName: "orphaned-disks-cleaner",
		PjPath: "test-infra/prow/jobs/",
	}, "Unexpected value of first prowjob to test name.")
	assert.Containsf(t, testCfg.PjNames, pjCfg{
		PjName: "post-master-kyma-gke-integration",
		PjPath: "test-infra/prow/jobs/",
	}, "Unexpected value of first prowjob to test name.")
	assert.Equalf(t, "test-infra/prow/config.yaml", testCfg.ConfigPath, "Unexpected value of prow config to test.")
	assert.Equalf(t, 1212, testCfg.PrConfigs["kyma-project"]["kyma"].PrNumber, "Unexpected value of other pr number.")
}

func TestNewTestPJ(t *testing.T) {
	testCfg := readTestCfg(testCfgFile)
	o := gatherOptions(testCfg.ConfigPath)
	fakeGitHubClient := &fakegithub.FakeClient{}
	fakeGitHubClient.PullRequests = map[int]*github.PullRequest{otherPrNumber: {
		User: github.User{Login: otherPrAuthor},
		Head: github.PullRequestBranch{SHA: otherPrSHA},
	}}
	o.githubClient = fakeGitHubClient
	var testPrCfg *map[string]prOrg
	if testPrCfg = &testCfg.PrConfigs; testPrCfg != nil {
		o.getPullRequests(testCfg)
	}
	for _, pjCfg := range testCfg.PjNames {
		fmt.Printf("Testing with values\n\tPjName: %s\n\tPjPath: %s\n", pjCfg.PjName, pjCfg.PjPath)
		pj := newTestPJ(pjCfg, o)
		assert.Equalf(t, "untrusted-workload", pj.Spec.Cluster, "Prowjob cluster name is not : trusted-workload")
		assert.Regexpf(t, "^test_of_prowjob_.*", pj.Spec.Job, "Prowjob name doesn't start with : test_of_prowjob_")
		if pj.Spec.Type == "periodic" {
			assert.Containsf(t, pj.Spec.ExtraRefs, prowapi.Refs{
				Org:     os.Getenv("REPO_OWNER"),
				Repo:    os.Getenv("REPO_NAME"),
				BaseRef: os.Getenv("PULL_BASE_REF"),
				BaseSHA: os.Getenv("PULL_BASE_SHA"),
				Pulls: []prowapi.Pull{{
					Number: prNumber,
					Author: prAuthor,
					SHA:    os.Getenv("PULL_PULL_SHA"),
				}},
				PathAlias: fmt.Sprintf("github.com/%s/%s", os.Getenv("REPO_OWNER"), os.Getenv("REPO_NAME")),
			}, "Refs BaseRef is not : master")
			if pj.Spec.Report == true {
				assert.Equalf(t, pj.Spec.Report, true, "Unexpected Spec.Report value")
			} else {
				assert.Equalf(t, pj.Spec.Report, false, "Unexpected Spec.Report value")
			}
		} else {
			assert.Equalf(t, prowapi.Refs{
				Org:     os.Getenv("REPO_OWNER"),
				Repo:    os.Getenv("REPO_NAME"),
				BaseRef: os.Getenv("PULL_BASE_REF"),
				BaseSHA: os.Getenv("PULL_BASE_SHA"),
				Pulls: []prowapi.Pull{{
					Number: prNumber,
					Author: prAuthor,
					SHA:    os.Getenv("PULL_PULL_SHA"),
				}},
				PathAlias: fmt.Sprintf("github.com/%s/%s", os.Getenv("REPO_OWNER"), os.Getenv("REPO_NAME")),
			}, *pj.Spec.Refs, "Refs BaseRef is not : master")
			assert.Lenf(t, pj.Spec.ExtraRefs, 1, "ExtraRefs slice doesn't contain one element.")
			assert.Equalf(t, pj.Spec.ExtraRefs[0].Org, otherPrOrg, "Unexpected value of ExtraRefs.Org")
			assert.Equalf(t, pj.Spec.ExtraRefs[0].Repo, otherPrRepo, "Unexpected value of ExtraRefs.Repo")
			assert.Equalf(t, pj.Spec.ExtraRefs[0].BaseRef, "master", "Unexpected value of ExtraRefs.BaseRef")
			assert.Equalf(t, pj.Spec.ExtraRefs[0].Pulls[0].Number, otherPrNumber, "Unexpected value of Pulls.Number")
			assert.Equalf(t, pj.Spec.ExtraRefs[0].Pulls[0].Author, otherPrAuthor, "Unexpected value of Pulls.Author")
			assert.Equalf(t, pj.Spec.ExtraRefs[0].Pulls[0].SHA, otherPrSHA, "Unexpected value of Pulls.SHA")
			if pj.Spec.Report == true {
				assert.Equalf(t, pj.Spec.Report, true, "Unexpected Spec.Report value")
			} else {
				assert.Equalf(t, pj.Spec.Report, false, "Unexpected Spec.Report value")
			}
		}
	}
}
