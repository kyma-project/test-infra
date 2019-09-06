package buildjob

import (
	"fmt"
	. "github.com/kyma-project/test-infra/development/tools/jobs/tester"
)

type Option func(suite *Suite)

func Component(name, image string) Option {
	return func(suite *Suite) {
		suite.path = fmt.Sprintf("components/%s", name)
		suite.image = image
		suite.expectedRunIfChanged = fmt.Sprintf("^%s/", suite.path)
		suite.fileExpectedToTriggerJob = fmt.Sprintf("%s/fix", suite.path)
	}
}

func Test(name, image string) Option {
	return func(suite *Suite) {
		suite.path = fmt.Sprintf("tests/%s", name)
		suite.image = image
		suite.expectedRunIfChanged = fmt.Sprintf("^%s/", suite.path)
		suite.fileExpectedToTriggerJob = fmt.Sprintf("%s/fix", suite.path)
	}
}

func Tool(name, image string) Option {
	return func(suite *Suite) {
		suite.path = fmt.Sprintf("tools/%s", name)
		suite.image = image
		suite.expectedRunIfChanged = fmt.Sprintf("^%s/", suite.path)
		suite.fileExpectedToTriggerJob = fmt.Sprintf("%s/fix", suite.path)
	}
}

func KymaRepo() Option {
	return func(suite *Suite) {
		suite.repository = "github.com/kyma-project/kyma"
	}
}

func JobFileSuffix(suffix string) Option {
	return func(suite *Suite) {
		suite.jobsFileSuffix = "-" + suffix
	}
}

func Until(rel *SupportedRelease) Option {
	return func(suite *Suite) {
		suite.releases = GetKymaReleasesUntil(rel)
		suite.doNotExpectMasterJobs = true
	}
}

func AllReleases() Option {
	return func(suite *Suite) {
		suite.releases = GetAllKymaReleases()
	}
}

func Since(rel *SupportedRelease) Option {
	return func(suite *Suite) {
		suite.releases = GetKymaReleasesSince(rel)
	}
}

func RunIfChanged(regexp, fileToCheck string) Option {
	return func(suite *Suite) {
		suite.expectedRunIfChanged = regexp
		suite.fileExpectedToTriggerJob = fileToCheck
	}
}
