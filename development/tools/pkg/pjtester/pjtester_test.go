package pjtester

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	gogithub "github.com/google/go-github/v31/github"
	"github.com/kyma-project/test-infra/development/tools/pkg/prtagbuilder/mocks"
	"github.com/stretchr/testify/assert"
	prowapi "k8s.io/test-infra/prow/apis/prowjobs/v1"
	prowflagutil "k8s.io/test-infra/prow/flagutil"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/github/fakegithub"
)

var (
	testInfraPrAuthor       string
	testInfraBaseRef        string
	testInfraBaseSHA        string
	testInfraPrNumber       int
	testInfraPrHeadSHA      string
	testInfraPrOrg          string
	testInfraPrRepo         string
	kymaPrAuthor            string
	kymaPrHeadSHA           string
	kymaPrNumber            int
	kymaPrOrg               string
	kymaPrRepo              string
	kymaBaseRef             string
	kymaBaseSHA             string
	fakeRepoPrAuthor        string
	fakeRepoPrHeadSHA       string
	fakeRepoPrNumber        int
	fakeRepoPrOrg           string
	fakeRepoPrRepo          string
	fakeRepoBaseRef         string
	fakeRepoBaseSHA         string
	fakeRepoMasterName      string
	fakeRepoProtectedBranch bool
	fakeRepoMerged          bool
	fakeRepoCommitMessage   string
	fakeRepoRefs            prowapi.Refs
	kymaRefs                prowapi.Refs
	testInfraRefs           prowapi.Refs
	ghOptions               *prowflagutil.GitHubOptions
)

func TestMain(m *testing.M) {
	pjPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	// set data for testing
	testInfraPrAuthor = "testInfraAuthor"
	testInfraBaseRef = "master"
	testInfraBaseSHA = "testInfraMasterSHA"
	testInfraPrNumber = 12345
	testInfraPrHeadSHA = "testInfraPrHeadSHA"
	testInfraPrOrg = "kyma-project"
	testInfraPrRepo = "test-infra"
	kymaPrAuthor = "kymaAuthor"
	kymaPrHeadSHA = "kymaPrHeadSHA"
	kymaPrNumber = 1212
	kymaPrOrg = "kyma-project"
	kymaPrRepo = "kyma"
	kymaBaseRef = "master"
	kymaBaseSHA = "kymaBaseSHA"
	fakeRepoPrAuthor = "fakeRepoAuthor"
	fakeRepoPrHeadSHA = "fakeRepoSHA"
	fakeRepoPrNumber = 1515
	fakeRepoPrOrg = "kyma-project"
	fakeRepoPrRepo = "fake-repo"
	fakeRepoBaseRef = "master"
	fakeRepoBaseSHA = "fakeRepoSHA"
	fakeRepoMasterName = "master"
	fakeRepoProtectedBranch = false
	fakeRepoMerged = true
	fakeRepoCommitMessage = fmt.Sprintf("Fake Repo commit message (#%s)", strconv.Itoa(fakeRepoPrNumber))
	repoDir := filepath.Clean(fmt.Sprintf("%s/../../../../..", pjPath))
	kymaRefs = prowapi.Refs{
		Org:     kymaPrOrg,
		Repo:    kymaPrRepo,
		BaseRef: kymaBaseRef,
		BaseSHA: kymaBaseSHA,
		Pulls: []prowapi.Pull{{
			Number: kymaPrNumber,
			Author: kymaPrAuthor,
			SHA:    kymaPrHeadSHA,
		}},
		PathAlias: fmt.Sprintf("github.com/%s/%s", kymaPrOrg, kymaPrRepo),
	}
	testInfraRefs = prowapi.Refs{
		Org:     testInfraPrOrg,
		Repo:    testInfraPrRepo,
		BaseRef: testInfraBaseRef,
		BaseSHA: testInfraBaseSHA,
		Pulls: []prowapi.Pull{{
			Number: testInfraPrNumber,
			Author: testInfraPrAuthor,
			SHA:    testInfraPrHeadSHA,
		}},
		PathAlias: fmt.Sprintf("github.com/%s/%s", testInfraPrOrg, testInfraPrRepo),
	}
	fakeRepoRefs = prowapi.Refs{
		Org:     fakeRepoPrOrg,
		Repo:    fakeRepoPrRepo,
		BaseRef: fakeRepoBaseRef,
		BaseSHA: fakeRepoBaseSHA,
		Pulls: []prowapi.Pull{{
			Number: fakeRepoPrNumber,
			Author: fakeRepoPrAuthor,
			SHA:    fakeRepoPrHeadSHA,
		}},
		PathAlias: fmt.Sprintf("github.com/%s/%s", fakeRepoPrOrg, fakeRepoPrRepo),
	}

	// set env variables for pjtester
	os.Setenv("KYMA_PROJECT_DIR", repoDir)
	os.Setenv("PULL_BASE_REF", testInfraBaseRef)
	os.Setenv("PULL_BASE_SHA", testInfraBaseSHA)
	os.Setenv("PULL_NUMBER", strconv.Itoa(testInfraPrNumber))
	os.Setenv("PULL_PULL_SHA", testInfraPrHeadSHA)
	os.Setenv("REPO_OWNER", testInfraPrOrg)
	os.Setenv("REPO_NAME", testInfraPrRepo)
	os.Setenv("JOB_SPEC", fmt.Sprintf("{\"type\":\"presubmit\",\"job\":\"job-name\",\"buildid\":\"0\",\"prowjobid\":\"uuid\",\"refs\":{\"org\":\"org-name\",\"repo\":\"repo-name\",\"base_ref\":\"base-ref\",\"base_sha\":\"base-sha\",\"pulls\":[{\"number\":1,\"author\":\"%s\",\"sha\":\"pull-sha\"}]}}", testInfraPrAuthor))
	testCfgFile = fmt.Sprintf("%s/test-infra/development/tools/pkg/pjtester/test_artifacts/pjtester.yaml", repoDir)
	ghOptions = prowflagutil.NewGitHubOptions()
	os.Exit(m.Run())
}

func TestReadTestCfg(t *testing.T) {
	testCfg := readTestCfg(testCfgFile)
	assert.Containsf(t, testCfg.PjNames, pjCfg{
		PjName: "test-infra-presubmit-test-job",
		PjPath: "test-infra/development/tools/pkg/pjtester/test_artifacts/",
		Report: true,
	}, "pjCfg for test-infra-presubmit-test-job has wrong values.")
	assert.Containsf(t, testCfg.PjNames, pjCfg{
		PjName: "fake-repo-presubmit-test-job",
		PjPath: "test-infra/development/tools/pkg/pjtester/test_artifacts/",
		Report: false,
	}, "pjCfg for fake-repo-presubmit-test-job has wrong values.")
	assert.Containsf(t, testCfg.PjNames, pjCfg{
		PjName: "orphaned-disks-cleaner",
		PjPath: "test-infra/prow/jobs/",
	}, "pjCfg for orphaned-disks-cleaner has wrong values")
	assert.Containsf(t, testCfg.PjNames, pjCfg{
		PjName: "post-master-kyma-gke-integration",
		PjPath: "test-infra/prow/jobs/",
	}, "pjCfg for post-master-kyma-gke-integration has wrong values.")
	assert.Equalf(t, "test-infra/prow/config.yaml", testCfg.ConfigPath, "pjtester has wrong path to prow config.yaml file.")
	assert.Equalf(t, 1212, testCfg.PrConfigs["kyma-project"]["kyma"].PrNumber, "PR number for kyma read from pjtester.yaml file is wrong.")
}

func TestNewTestPJ(t *testing.T) {
	testCfg := readTestCfg(testCfgFile)
	o := gatherOptions(testCfg.ConfigPath, *ghOptions)
	fakeGitHubClient := &fakegithub.FakeClient{}
	fakeGitHubClient.PullRequests = map[int]*github.PullRequest{
		kymaPrNumber: {
			User: github.User{
				Login: kymaPrAuthor,
			},
			Head: github.PullRequestBranch{
				SHA: kymaPrHeadSHA,
			},
			Number: kymaPrNumber,
			Base: github.PullRequestBranch{
				SHA: kymaBaseSHA,
				Ref: kymaBaseRef,
			},
		},
		fakeRepoPrNumber: {
			User: github.User{
				Login: fakeRepoPrAuthor,
			},
			Head: github.PullRequestBranch{
				SHA: fakeRepoPrHeadSHA,
			},
			Number: fakeRepoPrNumber,
			Base: github.PullRequestBranch{
				SHA: fakeRepoBaseSHA,
				Ref: fakeRepoBaseRef,
			},
		},
	}
	o.githubClient = fakeGitHubClient
	o.prFinder = mocks.NewFakeGhClient(nil)
	ctx := context.Background()
	o.prFinder.Repositories.(*mocks.GithubRepoService).On("GetBranch", ctx, fakeRepoPrOrg, fakeRepoPrRepo, fakeRepoBaseRef).Return(&gogithub.Branch{
		Name: &fakeRepoMasterName,
		Commit: &gogithub.RepositoryCommit{
			Commit: &gogithub.Commit{
				SHA:     &fakeRepoBaseSHA,
				Message: &fakeRepoCommitMessage,
			},
			SHA: &fakeRepoBaseSHA,
		},
		Protected: &fakeRepoProtectedBranch,
	}, nil, nil)
	//o.prFinder.Repositories.(*mocks.GithubRepoService).On("GetCommit", ctx, fakeRepoPrOrg, fakeRepoPrRepo, fakeRepoBaseSHA)
	o.prFinder.PullRequests.(*mocks.GithubPullRequestsService).On("Get", ctx, fakeRepoPrOrg, fakeRepoPrRepo, fakeRepoPrNumber).Return(&gogithub.PullRequest{
		Merged:         &fakeRepoMerged,
		MergeCommitSHA: &fakeRepoPrHeadSHA,
	}, nil, nil)
	//o.prFinder.PullRequests.(*mocks.GithubPullRequestsService).MethodCalled("GetBranch", ctx, fakeRepoPrOrg, fakeRepoPrRepo, fakeRepoBaseRef)
	//o.prFinder.Repositories.(*mocks.GithubRepoService).MethodCalled("Get", ctx, fakeRepoPrOrg, fakeRepoPrRepo, fakeRepoPrNumber)
	defer o.prFinder.Repositories.(*mocks.GithubRepoService).AssertExpectations(t)
	defer o.prFinder.PullRequests.(*mocks.GithubPullRequestsService).AssertExpectations(t)
	var testPrCfg *map[string]prOrg
	if testPrCfg = &testCfg.PrConfigs; testPrCfg != nil {
		o.getPullRequests(testCfg)
	}
	for _, pjCfg := range testCfg.PjNames {
		pj := newTestPJ(pjCfg, o)
		assert.Equalf(t, "untrusted-workload", pj.Spec.Cluster, "Prowjob cluster name is not : untrusted-workload")
		assert.Regexpf(t, "^testinfraauthor_test_of_prowjob_.*", pj.Spec.Job, "Prowjob name doesn't start with : <author github user>_test_of_prowjob_")
		assert.LessOrEqualf(t, len(pj.Spec.Job), 63, "Size of prowjob name is greater than 63 bytes.")
		if pj.Spec.Report == true {
			assert.Equalf(t, pj.Spec.Report, true, "Unexpected Spec.Report value")
		} else {
			assert.Equalf(t, pj.Spec.Report, false, "Unexpected Spec.Report value")
		}
		if strings.Contains(pj.Spec.Job, "orphaned-disks-cleaner") {
			assert.Containsf(t, pj.Spec.ExtraRefs, testInfraRefs, "ExtraRefs for test-infra is not present")
			o.prFinder.Repositories.(*mocks.GithubRepoService).AssertNotCalled(t, "GetCommit", ctx, fakeRepoPrOrg, fakeRepoPrRepo, fakeRepoBaseSHA)
		} else if strings.Contains(pj.Spec.Job, "post-master-kyma-gke-integration") {
			assert.Equalf(t, kymaRefs, *pj.Spec.Refs, "Postsubmit Refs for kyma has wrong values")
			assert.Lenf(t, pj.Spec.ExtraRefs, 1, "ExtraRefs slice doesn't contain one element.")
			assert.Equalf(t, testInfraRefs, pj.Spec.ExtraRefs[0], "ExtraRefs for test-infra is not present")
			o.prFinder.Repositories.(*mocks.GithubRepoService).AssertNotCalled(t, "GetCommit", ctx, fakeRepoPrOrg, fakeRepoPrRepo, fakeRepoBaseSHA)
		} else if strings.Contains(pj.Spec.Job, "test-infra-presubmit-test-job") {
			assert.Equalf(t, testInfraRefs, *pj.Spec.Refs, "Presubmit Refs for test-infra has wrong values")
			assert.Lenf(t, pj.Spec.ExtraRefs, 1, "ExtraRefs slice doesn't contain one element.")
			assert.Equalf(t, kymaRefs.Repo, pj.Spec.ExtraRefs[0].Repo, "Pre ExtraRefs Repo is not as expected")
			assert.Equalf(t, kymaRefs.Org, pj.Spec.ExtraRefs[0].Org, "Pre ExtraRefs Org is not as expected")
			assert.Equalf(t, kymaRefs.BaseSHA, pj.Spec.ExtraRefs[0].BaseSHA, "Pre ExtraRefs Org is not as expected")
			assert.Equalf(t, kymaRefs.BaseRef, pj.Spec.ExtraRefs[0].BaseRef, "Pre ExtraRefs Org is not as expected")
			assert.Equalf(t, kymaRefs.Pulls[0].SHA, pj.Spec.ExtraRefs[0].Pulls[0].SHA, "Pre ExtraRefs Org is not as expected")
			assert.Equalf(t, kymaRefs.Pulls[0].Author, pj.Spec.ExtraRefs[0].Pulls[0].Author, "Pre ExtraRefs Org is not as expected")
			assert.Equalf(t, kymaRefs.Pulls[0].Number, pj.Spec.ExtraRefs[0].Pulls[0].Number, "Pre ExtraRefs Org is not as expected")
			o.prFinder.Repositories.(*mocks.GithubRepoService).AssertNotCalled(t, "GetCommit", ctx, fakeRepoPrOrg, fakeRepoPrRepo, fakeRepoBaseSHA)
		}
		if strings.Contains(pj.Spec.Job, "fake-repo-presubmit-test-job") {
			assert.Lenf(t, pj.Spec.ExtraRefs, 2, "ExtraRefs slice doesn't contain two elements.")
			assert.Equalf(t, testInfraRefs.Repo, pj.Spec.ExtraRefs[1].Repo, "Pre ExtraRefs Repo is not as expected")
			assert.Equalf(t, testInfraRefs.Org, pj.Spec.ExtraRefs[1].Org, "Pre ExtraRefs Org is not as expected")
			assert.Equalf(t, testInfraRefs.BaseSHA, pj.Spec.ExtraRefs[1].BaseSHA, "Pre ExtraRefs Org is not as expected")
			assert.Equalf(t, testInfraRefs.BaseRef, pj.Spec.ExtraRefs[1].BaseRef, "Pre ExtraRefs Org is not as expected")
			assert.Equalf(t, testInfraRefs.Pulls[0].SHA, pj.Spec.ExtraRefs[1].Pulls[0].SHA, "Pre ExtraRefs Org is not as expected")
			assert.Equalf(t, testInfraRefs.Pulls[0].Author, pj.Spec.ExtraRefs[1].Pulls[0].Author, "Pre ExtraRefs Org is not as expected")
			assert.Equalf(t, testInfraRefs.Pulls[0].Number, pj.Spec.ExtraRefs[1].Pulls[0].Number, "Pre ExtraRefs Org is not as expected")
			assert.Equalf(t, kymaRefs.Repo, pj.Spec.ExtraRefs[0].Repo, "Pre ExtraRefs Repo is not as expected")
			assert.Equalf(t, kymaRefs.Org, pj.Spec.ExtraRefs[0].Org, "Pre ExtraRefs Org is not as expected")
			assert.Equalf(t, kymaRefs.BaseSHA, pj.Spec.ExtraRefs[0].BaseSHA, "Pre ExtraRefs Org is not as expected")
			assert.Equalf(t, kymaRefs.BaseRef, pj.Spec.ExtraRefs[0].BaseRef, "Pre ExtraRefs Org is not as expected")
			assert.Equalf(t, kymaRefs.Pulls[0].SHA, pj.Spec.ExtraRefs[0].Pulls[0].SHA, "Pre ExtraRefs Org is not as expected")
			assert.Equalf(t, kymaRefs.Pulls[0].Author, pj.Spec.ExtraRefs[0].Pulls[0].Author, "Pre ExtraRefs Org is not as expected")
			assert.Equalf(t, kymaRefs.Pulls[0].Number, pj.Spec.ExtraRefs[0].Pulls[0].Number, "Pre ExtraRefs Org is not as expected")
		}
	}
}
