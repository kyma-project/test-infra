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
	pullNumber int
)

func TestMain(m *testing.M) {
	pjPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	os.Setenv("KYMA_PROJECT_DIR", filepath.Clean(fmt.Sprintf("%s/../../../../..", pjPath)))
	os.Setenv("PULL_BASE_REF", "master")
	os.Setenv("PULL_BASE_SHA", "masterbranchsha")
	os.Setenv("PULL_NUMBER", "12345")
	os.Setenv("PULL_PULL_SHA", "prheadsha")
	os.Setenv("REPO_OWNER", "kyma-project")
	os.Setenv("REPO_NAME", "test-infra")
	os.Setenv("JOB_SPEC", "{\"type\":\"presubmit\",\"job\":\"job-name\",\"buildid\":\"0\",\"prowjobid\":\"uuid\",\"refs\":{\"org\":\"org-name\",\"repo\":\"repo-name\",\"base_ref\":\"base-ref\",\"base_sha\":\"base-sha\",\"pulls\":[{\"number\":1,\"author\":\"testAuthor\",\"sha\":\"pull-sha\"}]}}")
	pullNumber, _ = strconv.Atoi(os.Getenv("PULL_NUMBER"))
	testCfgFile = fmt.Sprintf("%s/test-infra/development/tools/pkg/pjtester/test_artifacts/pjtester.yaml", os.Getenv("KYMA_PROJECT_DIR"))
	os.Exit(m.Run())

}

func TestReadTestCfg(t *testing.T) {
	testCfg := readTestCfg(testCfgFile)
	assert.Containsf(t, testCfg.PjNames, pjCfg{
		PjName: "presubmit-test-job",
		PjPath: "test-infra/development/tools/pkg/pjtester/test_artifacts/",
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
}

func TestNewTestPJ(t *testing.T) {
	testCfg := readTestCfg(testCfgFile)
	for _, pjCfg := range testCfg.PjNames {
		fmt.Printf("Testing with values\n\tPjName: %s\n\tPjPath: %s\n", pjCfg.PjName, pjCfg.PjPath)
		pj := newTestPJ(pjCfg, testCfg.ConfigPath)
		assert.Equalf(t, "untrusted-workload", pj.Spec.Cluster, "Prowjob cluster name is not : trusted-workload")
		assert.Regexpf(t, "^test_of_prowjob_.*", pj.Spec.Job, "Prowjob name doesn't start with : test_of_prowjob_")
		if pj.Spec.Type == "periodic" {
			assert.Containsf(t, pj.Spec.ExtraRefs, prowapi.Refs{
				Org:     os.Getenv("REPO_OWNER"),
				Repo:    os.Getenv("REPO_NAME"),
				BaseRef: os.Getenv("PULL_BASE_REF"),
				BaseSHA: os.Getenv("PULL_BASE_SHA"),
				Pulls: []prowapi.Pull{{
					Number: pullNumber,
					Author: "testAuthor",
					SHA:    os.Getenv("PULL_PULL_SHA"),
				}},
				PathAlias: fmt.Sprintf("github.com/%s/%s", os.Getenv("REPO_OWNER"), os.Getenv("REPO_NAME")),
			}, "Refs BaseRef is not : master")
		} else {
			assert.Equalf(t, prowapi.Refs{
				Org:     os.Getenv("REPO_OWNER"),
				Repo:    os.Getenv("REPO_NAME"),
				BaseRef: os.Getenv("PULL_BASE_REF"),
				BaseSHA: os.Getenv("PULL_BASE_SHA"),
				Pulls: []prowapi.Pull{{
					Number: pullNumber,
					Author: "testAuthor",
					SHA:    os.Getenv("PULL_PULL_SHA"),
				}},
				PathAlias: fmt.Sprintf("github.com/%s/%s", os.Getenv("REPO_OWNER"), os.Getenv("REPO_NAME")),
			}, *pj.Spec.Refs, "Refs BaseRef is not : master")
		}
	}
}
