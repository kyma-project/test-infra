package pipelines_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kyma-project/test-infra/pkg/azuredevops/pipelines"
	pipelinesmocks "github.com/kyma-project/test-infra/pkg/azuredevops/pipelines/mocks"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"
	adoPipelines "github.com/microsoft/azure-devops-go-api/azuredevops/v7/pipelines"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

type ginkgoT struct {
	GinkgoTInterface
}

func (t ginkgoT) Cleanup(f func()) {
	f()
}

var _ = Describe("Pipelines", func() {
	var (
		ctx           context.Context
		mockADOClient *pipelinesmocks.MockClient
		adoConfig     pipelines.Config
		t             ginkgoT
	)

	BeforeEach(func() {
		ctx = context.Background()
		t = ginkgoT{}
		t.GinkgoTInterface = GinkgoT()
		mockADOClient = pipelinesmocks.NewMockClient(t)
		adoConfig = pipelines.Config{
			ADOOrganizationURL: "https://dev.azure.com",
			ADOProjectName:     "example-project",
			ADOPipelineID:      123,
			ADOPipelineVersion: 1,
		}
	})

	Describe("GetRunResult", func() {
		var (
			runArgs           adoPipelines.GetRunArgs
			mockRunInProgress *adoPipelines.Run
			mockRunSucceeded  *adoPipelines.Run
		)

		BeforeEach(func() {
			runArgs = adoPipelines.GetRunArgs{
				Project:    &adoConfig.ADOProjectName,
				PipelineId: &adoConfig.ADOPipelineID,
				RunId:      ptr.To(42),
			}
			mockRunInProgress = &adoPipelines.Run{State: &adoPipelines.RunStateValues.InProgress}
			mockRunSucceeded = &adoPipelines.Run{State: &adoPipelines.RunStateValues.Completed, Result: &adoPipelines.RunResultValues.Succeeded}
		})

		It("should return the pipeline run result succeeded", func() {
			mockADOClient.On("GetRun", ctx, runArgs).Return(mockRunSucceeded, nil)

			result, err := pipelines.GetRunResult(ctx, mockADOClient, adoConfig, ptr.To(42), 3*time.Second)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(&adoPipelines.RunResultValues.Succeeded))
			mockADOClient.AssertCalled(t, "GetRun", ctx, runArgs)
			mockADOClient.AssertNumberOfCalls(t, "GetRun", 1)
			mockADOClient.AssertExpectations(GinkgoT())
		})

		It("should handle pipeline still in progress", func() {
			mockADOClient.On("GetRun", ctx, runArgs).Return(mockRunInProgress, nil).Once()
			mockADOClient.On("GetRun", ctx, runArgs).Return(mockRunSucceeded, nil).Once()

			result, err := pipelines.GetRunResult(ctx, mockADOClient, adoConfig, ptr.To(42), 3*time.Second)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(&adoPipelines.RunResultValues.Succeeded))
			mockADOClient.AssertCalled(t, "GetRun", ctx, runArgs)
			mockADOClient.AssertNumberOfCalls(t, "GetRun", 2)
			mockADOClient.AssertExpectations(GinkgoT())
		})

		It("should handle ADO client error", func() {
			mockADOClient.On("GetRun", ctx, runArgs).Return(nil, fmt.Errorf("ADO client error"))

			_, err := pipelines.GetRunResult(ctx, mockADOClient, adoConfig, ptr.To(42), 3*time.Second)

			Expect(err).To(HaveOccurred())
			mockADOClient.AssertCalled(t, "GetRun", ctx, runArgs)
			mockADOClient.AssertNumberOfCalls(t, "GetRun", 1)
			mockADOClient.AssertExpectations(GinkgoT())
		})
	})

	Describe("GetRunLogs", func() {
		var (
			mockBuildClient  *pipelinesmocks.MockBuildClient
			mockHTTPClient   *pipelinesmocks.MockHTTPClient
			getBuildLogsArgs build.GetBuildLogsArgs
			mockBuildLogs    *[]build.BuildLog
		)

		BeforeEach(func() {
			mockBuildClient = pipelinesmocks.NewMockBuildClient(t)
			mockHTTPClient = pipelinesmocks.NewMockHTTPClient(t)
			getBuildLogsArgs = build.GetBuildLogsArgs{
				Project: &adoConfig.ADOProjectName,
				BuildId: ptr.To(42),
			}
			mockBuildLogs = &[]build.BuildLog{{Url: ptr.To("https://example.com/log")}}
		})

		// TODO: Need a test for HTTP response status code != 2xx
		// TODO: Need a tests for other errors returned by GetBuildLogs.
		It("should return build logs", func() {
			mockBuildClient.On("GetBuildLogs", ctx, getBuildLogsArgs).Return(mockBuildLogs, nil)
			mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("log content")),
			}, nil)

			logs, err := pipelines.GetRunLogs(ctx, mockBuildClient, mockHTTPClient, adoConfig, ptr.To(42), "somePAT")

			Expect(err).ToNot(HaveOccurred())
			Expect(logs).To(Equal("log content"))
			mockBuildClient.AssertCalled(t, "GetBuildLogs", ctx, getBuildLogsArgs)
			mockBuildClient.AssertNumberOfCalls(t, "GetBuildLogs", 1)
			mockBuildClient.AssertExpectations(GinkgoT())
			mockHTTPClient.AssertCalled(t, "Do", mock.AnythingOfType("*http.Request"))
			mockHTTPClient.AssertNumberOfCalls(t, "Do", 1)
			mockHTTPClient.AssertExpectations(GinkgoT())
		})

		It("should handle build client error", func() {
			mockBuildClient.On("GetBuildLogs", ctx, getBuildLogsArgs).Return(nil, fmt.Errorf("build client error"))

			_, err := pipelines.GetRunLogs(ctx, mockBuildClient, mockHTTPClient, adoConfig, ptr.To(42), "somePAT")

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("failed getting build logs metadata, err: build client error"))
			mockBuildClient.AssertCalled(t, "GetBuildLogs", ctx, getBuildLogsArgs)
			mockBuildClient.AssertNumberOfCalls(t, "GetBuildLogs", 1)
			mockBuildClient.AssertExpectations(GinkgoT())
			mockHTTPClient.AssertNotCalled(t, "Do", mock.AnythingOfType("*http.Request"))
			mockHTTPClient.AssertNumberOfCalls(t, "Do", 0)
			mockHTTPClient.AssertExpectations(GinkgoT())
		})

		It("should handle HTTP request error", func() {
			mockBuildClient.On("GetBuildLogs", ctx, getBuildLogsArgs).Return(mockBuildLogs, nil)
			mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).Return(nil, fmt.Errorf("HTTP request error"))

			_, err := pipelines.GetRunLogs(ctx, mockBuildClient, mockHTTPClient, adoConfig, ptr.To(42), "somePAT")

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("failed http request getting build log, err: HTTP request error"))
			mockBuildClient.AssertCalled(t, "GetBuildLogs", ctx, getBuildLogsArgs)
			mockBuildClient.AssertNumberOfCalls(t, "GetBuildLogs", 1)
			mockBuildClient.AssertExpectations(GinkgoT())
			mockHTTPClient.AssertCalled(t, "Do", mock.AnythingOfType("*http.Request"))
			mockHTTPClient.AssertNumberOfCalls(t, "Do", 1)
			mockHTTPClient.AssertExpectations(GinkgoT())
		})
	})

	Describe("Run", func() {
		var (
			templateParams  map[string]string
			runPipelineArgs adoPipelines.RunPipelineArgs
		)

		BeforeEach(func() {
			templateParams = map[string]string{"param1": "value1", "param2": "value2"}
			runPipelineArgs = adoPipelines.RunPipelineArgs{
				Project:    &adoConfig.ADOProjectName,
				PipelineId: &adoConfig.ADOPipelineID,
				RunParameters: &adoPipelines.RunPipelineParameters{
					PreviewRun:         ptr.To(false),
					TemplateParameters: &templateParams,
				},
				PipelineVersion: &adoConfig.ADOPipelineVersion,
			}
		})

		It("should run the pipeline", func() {
			mockRun := &adoPipelines.Run{Id: ptr.To(123)}
			mockADOClient.On("RunPipeline", ctx, runPipelineArgs).Return(mockRun, nil)

			run, err := pipelines.Run(ctx, mockADOClient, templateParams, adoConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(run.Id).To(Equal(ptr.To(123)))
			mockADOClient.AssertCalled(t, "RunPipeline", ctx, runPipelineArgs)
			mockADOClient.AssertNumberOfCalls(t, "RunPipeline", 1)
			mockADOClient.AssertExpectations(GinkgoT())
		})

		It("should handle ADO client error", func() {
			mockADOClient.On("RunPipeline", ctx, runPipelineArgs).Return(nil, fmt.Errorf("ADO client error"))

			_, err := pipelines.Run(ctx, mockADOClient, templateParams, adoConfig)
			Expect(err).To(HaveOccurred())
			mockADOClient.AssertCalled(t, "RunPipeline", ctx, runPipelineArgs)
			mockADOClient.AssertNumberOfCalls(t, "RunPipeline", 1)
			mockADOClient.AssertExpectations(GinkgoT())
		})

		It("should run the pipeline in preview mode", func() {
			finalYaml := "pipeline:\n  stages:\n  - stage: Build\n    jobs:\n    - job: Build\n      steps:\n      - script: echo Hello, world!\n        displayName: 'Run a one-line script'"
			runPipelineArgs.RunParameters.PreviewRun = ptr.To(true)
			mockRun := &adoPipelines.Run{Id: ptr.To(123), FinalYaml: &finalYaml}
			mockADOClient.On("RunPipeline", ctx, runPipelineArgs).Return(mockRun, nil)

			run, err := pipelines.Run(ctx, mockADOClient, templateParams, adoConfig, pipelines.PipelinePreviewRun)
			Expect(err).ToNot(HaveOccurred())
			Expect(run.Id).To(Equal(ptr.To(123)))
			Expect(run.FinalYaml).To(Equal(&finalYaml))
			mockADOClient.AssertCalled(t, "RunPipeline", ctx, runPipelineArgs)
			mockADOClient.AssertNumberOfCalls(t, "RunPipeline", 1)
			mockADOClient.AssertExpectations(GinkgoT())
		})
	})

	Describe("PipelinePreviewRun", func() {
		It("should set PreviewRun to true", func() {
			args := &adoPipelines.RunPipelineArgs{
				RunParameters: &adoPipelines.RunPipelineParameters{
					PreviewRun: ptr.To(false),
				},
			}

			pipelines.PipelinePreviewRun(args)

			Expect(args.RunParameters.PreviewRun).To(Equal(ptr.To(true)))
		})
	})
})
