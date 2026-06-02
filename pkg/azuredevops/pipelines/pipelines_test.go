package pipelines_test

import (
	"context"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kyma-project/test-infra/pkg/azuredevops/pipelines"
	pipelinesMocks "github.com/kyma-project/test-infra/pkg/azuredevops/pipelines/mocks"

	adoPipelines "github.com/microsoft/azure-devops-go-api/azuredevops/v7/pipelines"
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
		mockADOClient *pipelinesMocks.MockClient
		adoConfig     pipelines.Config
		t             ginkgoT
	)

	BeforeEach(func() {
		ctx = context.Background()
		t = ginkgoT{}
		t.GinkgoTInterface = GinkgoT()
		mockADOClient = pipelinesMocks.NewMockClient(t)
		adoConfig = pipelines.Config{
			ADOOrganizationURL: "https://dev.azure.com",
			ADOProjectName:     "example-project",
			ADOPipelineID:      123,
			ADOPipelineVersion: 1,
			ADORetryStrategy: pipelines.RetryStrategy{
				Attempts: 3,
				Delay:    5 * time.Second,
			},
			ADORefreshInterval: 3 * time.Second,
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

			result, err := pipelines.GetRunResult(ctx, mockADOClient, adoConfig, ptr.To(42))

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(&adoPipelines.RunResultValues.Succeeded))
			mockADOClient.AssertCalled(t, "GetRun", ctx, runArgs)
			mockADOClient.AssertNumberOfCalls(t, "GetRun", 1)
			mockADOClient.AssertExpectations(GinkgoT())
		})

		It("should handle pipeline still in progress", func() {
			mockADOClient.On("GetRun", ctx, runArgs).Return(mockRunInProgress, nil).Once()
			mockADOClient.On("GetRun", ctx, runArgs).Return(mockRunSucceeded, nil).Once()

			result, err := pipelines.GetRunResult(ctx, mockADOClient, adoConfig, ptr.To(42))

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(&adoPipelines.RunResultValues.Succeeded))
			mockADOClient.AssertCalled(t, "GetRun", ctx, runArgs)
			mockADOClient.AssertNumberOfCalls(t, "GetRun", 2)
			mockADOClient.AssertExpectations(GinkgoT())
		})

		It("should handle ADO client error", func() {
			mockADOClient.On("GetRun", ctx, runArgs).Return(nil, fmt.Errorf("ADO client error"))

			_, err := pipelines.GetRunResult(ctx, mockADOClient, adoConfig, ptr.To(42))

			Expect(err).To(HaveOccurred())
			mockADOClient.AssertCalled(t, "GetRun", ctx, runArgs)
			mockADOClient.AssertNumberOfCalls(t, "GetRun", 3)
			mockADOClient.AssertExpectations(GinkgoT())
		})
	})

	Describe("NewRunPipelineArgs", func() {
		var (
			templateParameters map[string]string
			pipelineRunArgs    []pipelines.RunPipelineArgsOptions
		)

		BeforeEach(func() {
			templateParameters = map[string]string{
				"key1": "value1",
				"key2": "value2",
			}
		})

		Context("when NewRunPipelineArgs is successful", func() {
			It("should return the correct PipelineArgs and no error", func() {
				pipelineArgs, err := pipelines.NewRunPipelineArgs(templateParameters, adoConfig)
				Expect(err).NotTo(HaveOccurred())
				Expect(pipelineArgs.Project).To(Equal(&adoConfig.ADOProjectName))
				Expect(pipelineArgs.PipelineId).To(Equal(&adoConfig.ADOPipelineID))
				Expect(pipelineArgs.PipelineVersion).To(Equal(&adoConfig.ADOPipelineVersion))
				Expect(pipelineArgs.RunParameters.TemplateParameters).To(Equal(&templateParameters))
				Expect(pipelineArgs.RunParameters.PreviewRun).To(Equal(ptr.To(false)))
				Expect(pipelineArgs).To(BeAssignableToTypeOf(adoPipelines.RunPipelineArgs{}))
			})
			Context("when PipelinePreviewRun option is passed", func() {
				var dummyOverrideYamlPath = "./dummyOverride.yaml"
				BeforeEach(func() {
					pipelineRunArgs = []pipelines.RunPipelineArgsOptions{
						pipelines.PipelinePreviewRun(dummyOverrideYamlPath),
					}
					err := os.WriteFile(dummyOverrideYamlPath, []byte("dummyYamlContent"), 0644)
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					pipelineRunArgs = []pipelines.RunPipelineArgsOptions{}
					err := os.Remove(dummyOverrideYamlPath)
					Expect(err).NotTo(HaveOccurred())
				})

				It("should enable preview run and set YamlOverride according to file content", func() {
					pipelineArgs, err := pipelines.NewRunPipelineArgs(templateParameters, adoConfig, pipelineRunArgs...)
					Expect(err).NotTo(HaveOccurred())
					Expect(pipelineArgs.Project).To(Equal(&adoConfig.ADOProjectName))
					Expect(pipelineArgs.PipelineId).To(Equal(&adoConfig.ADOPipelineID))
					Expect(pipelineArgs.PipelineVersion).To(Equal(&adoConfig.ADOPipelineVersion))
					Expect(pipelineArgs.RunParameters.TemplateParameters).To(Equal(&templateParameters))
					Expect(pipelineArgs).To(BeAssignableToTypeOf(adoPipelines.RunPipelineArgs{}))
					Expect(pipelineArgs.RunParameters.PreviewRun).To(Equal(ptr.To(true)))
					Expect(pipelineArgs.RunParameters.YamlOverride).To(Equal(ptr.To("dummyYamlContent")))
				})
			})
		})

		Context("when NewRunPipelineArgs fails", func() {
			BeforeEach(func() {
				pipelineRunArgs = []pipelines.RunPipelineArgsOptions{
					func(args *adoPipelines.RunPipelineArgs) error {
						return fmt.Errorf("dummy error")
					},
				}
			})

			It("should return an error", func() {
				pipelineArgs, err := pipelines.NewRunPipelineArgs(templateParameters, adoConfig, pipelineRunArgs...)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed setting pipeline run args, err: dummy error"))
				Expect(pipelineArgs).To(BeEquivalentTo(adoPipelines.RunPipelineArgs{}))
			})
		})
	})

	Describe("PipelinePreviewRun", func() {
		var (
			dummyOverrideYamlPath = "./dummyOverride.yaml"
			err                   error
			pipelineArgs          *adoPipelines.RunPipelineArgs
		)

		BeforeEach(func() {
			err = os.WriteFile(dummyOverrideYamlPath, []byte("dummyYamlContent"), 0644)
			Expect(err).NotTo(HaveOccurred())
			pipelineArgs = &adoPipelines.RunPipelineArgs{
				RunParameters: &adoPipelines.RunPipelineParameters{
					PreviewRun: ptr.To(false),
				},
			}
		})

		AfterEach(func() {
			err = os.Remove(dummyOverrideYamlPath)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should prepare function that sets PreviewRun to true and reads override yaml", func() {

			pipelinePreviewRun := pipelines.PipelinePreviewRun(dummyOverrideYamlPath)

			err := pipelinePreviewRun(pipelineArgs)
			Expect(err).NotTo(HaveOccurred())
			Expect(pipelineArgs.RunParameters.PreviewRun).To(Equal(ptr.To(true)))
			Expect(pipelineArgs.RunParameters.YamlOverride).To(Equal(ptr.To("dummyYamlContent")))
		})
		Context("when the override yaml file does not exist", func() {
			It("should return an error", func() {
				nonExistentFilePath := "/path/to/non-existent/file.yaml"
				pipelinePreviewRun := pipelines.PipelinePreviewRun(nonExistentFilePath)

				err := pipelinePreviewRun(pipelineArgs)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
