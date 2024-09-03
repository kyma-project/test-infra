package pipelines_test

import (
	"encoding/base64"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kyma-project/test-infra/pkg/azuredevops/pipelines"
)

var _ = Describe("Test OCIImageBuilderTemplateParams", func() {
	var params pipelines.OCIImageBuilderTemplateParams

	BeforeEach(func() {
		params = make(map[string]string)
	})

	It("sets the correct RepoName", func() {
		params.SetRepoName("testName")
		Expect(params["RepoName"]).To(Equal("testName"))
	})

	It("sets the correct RepoOwner", func() {
		params.SetRepoOwner("testOwner")
		Expect(params["RepoOwner"]).To(Equal("testOwner"))
	})

	It("sets the correct JobType to presubmit", func() {
		params.SetPresubmitJobType()
		Expect(params["JobType"]).To(Equal("presubmit"))
	})

	It("sets the correct JobType to postsubmit", func() {
		params.SetPostsubmitJobType()
		Expect(params["JobType"]).To(Equal("postsubmit"))
	})

	It("sets the correct JobType to workflow_dispatch", func() {
		params.SetWorkflowDispatchJobType()
		Expect(params["JobType"]).To(Equal("workflow_dispatch"))
	})

	It("sets the correct PullNumber", func() {
		params.SetPullNumber("123")
		Expect(params["PullNumber"]).To(Equal("123"))
	})

	It("sets the correct BaseSHA", func() {
		params.SetBaseSHA("abc")
		Expect(params["PullBaseSHA"]).To(Equal("abc"))
	})

	It("sets the correct PullSHA", func() {
		params.SetPullSHA("def")
		Expect(params["PullPullSHA"]).To(Equal("def"))
	})

	It("sets the correct ImageName", func() {
		params.SetImageName("my-image")
		Expect(params["Name"]).To(Equal("my-image"))
	})

	It("sets the correct DockerfilePath", func() {
		params.SetDockerfilePath("/path/to/dockerfile")
		Expect(params["Dockerfile"]).To(Equal("/path/to/dockerfile"))
	})

	It("sets the correct BuildContext", func() {
		params.SetBuildContext("/path/to/context")
		Expect(params["Context"]).To(Equal("/path/to/context"))
	})

	It("sets the correct ExportTags", func() {
		params.SetExportTags(true)
		Expect(params["ExportTags"]).To(Equal("true"))
	})

	It("sets the correct BuildArgs", func() {
		params.SetBuildArgs("arg1 arg2")
		Expect(params["BuildArgs"]).To(Equal("arg1 arg2"))
	})

	It("sets the correct ImageTags", func() {
		params.SetImageTags("tag1 tag2")
		expected := base64.StdEncoding.EncodeToString([]byte("tag1 tag2"))
		Expect(params["Tags"]).To(Equal(expected))
	})

	It("sets the correct UseKanikoConfigFromPR", func() {
		params.SetUseKanikoConfigFromPR(true)
		Expect(params["UseKanikoConfigFromPR"]).To(Equal("true"))
	})
	It("sets the correct Authorization", func() {
		params.SetAuthorization("some-token")
		Expect(params["Authorization"]).To(Equal("some-token"))
	})

	// TODO: Improve assertions with more specific matchers and values.
	It("validates the params correctly", func() {
		// Set all required parameters
		params.SetRepoName("testName")
		params.SetRepoOwner("testOwner")
		params.SetPresubmitJobType()
		params.SetBaseSHA("abc123")
		params.SetBaseRef("main")
		params.SetImageName("my-image")
		params.SetDockerfilePath("/path/to/dockerfile")
		params.SetBuildContext("/path/to/context")

		err := params.Validate()
		Expect(err).To(BeNil())
	})

	It("returns error if parameters are not set", func() {
		err := params.Validate()
		Expect(err).NotTo(BeNil())
	})

	It("returns error if JobType is not presubmit or postsubmit or workflow_dispatch", func() {
		params["JobType"] = "otherType"

		err := params.Validate()
		Expect(err).NotTo(BeNil())
	})
})
