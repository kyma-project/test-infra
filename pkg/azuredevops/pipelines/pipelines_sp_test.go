package pipelines_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"

	"github.com/kyma-project/test-infra/pkg/azuredevops/pipelines"
	pipelinesMocks "github.com/kyma-project/test-infra/pkg/azuredevops/pipelines/mocks"
)

type mockTokenProvider struct {
	token string
	err   error
}

func (m *mockTokenProvider) GetToken(_ context.Context) (string, error) {
	return m.token, m.err
}

var _ = Describe("Pipelines SP auth", func() {
	var (
		ctx       context.Context
		adoConfig pipelines.Config
		t         ginkgoT
	)

	BeforeEach(func() {
		ctx = context.Background()
		t = ginkgoT{}
		t.GinkgoTInterface = GinkgoT()
		adoConfig = pipelines.Config{
			ADOOrganizationURL: "https://dev.azure.com/org",
			ADOProjectName:     "example-project",
			ADOPipelineID:      123,
			ADORetryStrategy: pipelines.RetryStrategy{
				Attempts: 3,
				Delay:    0,
			},
		}
	})

	Describe("NewClientWithSP", func() {
		It("should return a client when token provider succeeds", func() {
			provider := &mockTokenProvider{token: "valid-bearer-token"}

			client, err := pipelines.NewClientWithSP(ctx, adoConfig.ADOOrganizationURL, provider)

			Expect(err).ToNot(HaveOccurred())
			Expect(client).ToNot(BeNil())
		})

		It("should return error when token provider fails", func() {
			provider := &mockTokenProvider{err: errors.New("token acquisition failed")}

			_, err := pipelines.NewClientWithSP(ctx, adoConfig.ADOOrganizationURL, provider)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed getting service principal token"))
			Expect(err.Error()).To(ContainSubstring("token acquisition failed"))
		})
	})

	Describe("NewBuildClientWithSP", func() {
		It("should return error when token provider fails", func() {
			provider := &mockTokenProvider{err: errors.New("token acquisition failed")}

			_, err := pipelines.NewBuildClientWithSP(ctx, adoConfig.ADOOrganizationURL, provider)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed getting service principal token"))
			Expect(err.Error()).To(ContainSubstring("token acquisition failed"))
		})
	})

	Describe("GetRunLogsWithBearerToken", func() {
		var (
			mockBuildClient *pipelinesMocks.MockBuildClient
			mockHTTPClient  *pipelinesMocks.MockHTTPClient
			mockBuildLogs   *[]build.BuildLog
		)

		BeforeEach(func() {
			mockBuildClient = pipelinesMocks.NewMockBuildClient(t)
			mockHTTPClient = pipelinesMocks.NewMockHTTPClient(t)
			mockBuildLogs = &[]build.BuildLog{{Url: ptr.To("https://example.com/log")}}
		})

		It("should return build logs with Bearer auth header", func() {
			mockBuildClient.On("GetBuildLogs", ctx, build.GetBuildLogsArgs{
				Project: &adoConfig.ADOProjectName,
				BuildId: ptr.To(42),
			}).Return(mockBuildLogs, nil)
			mockHTTPClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
				return req.Header.Get("Authorization") == "Bearer test-token"
			})).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("log content")),
			}, nil)

			logs, err := pipelines.GetRunLogsWithBearerToken(ctx, mockBuildClient, mockHTTPClient, adoConfig, ptr.To(42), "test-token")

			Expect(err).ToNot(HaveOccurred())
			Expect(logs).To(Equal("log content"))
			mockBuildClient.AssertExpectations(GinkgoT())
			mockHTTPClient.AssertExpectations(GinkgoT())
		})

		It("should return error when build client fails", func() {
			mockBuildClient.On("GetBuildLogs", ctx, build.GetBuildLogsArgs{
				Project: &adoConfig.ADOProjectName,
				BuildId: ptr.To(42),
			}).Return(nil, fmt.Errorf("build client error"))

			_, err := pipelines.GetRunLogsWithBearerToken(ctx, mockBuildClient, mockHTTPClient, adoConfig, ptr.To(42), "test-token")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed getting build logs metadata"))
			mockHTTPClient.AssertNotCalled(t, "Do", mock.Anything)
			mockBuildClient.AssertExpectations(GinkgoT())
		})

		It("should return error when HTTP request fails", func() {
			mockBuildClient.On("GetBuildLogs", ctx, build.GetBuildLogsArgs{
				Project: &adoConfig.ADOProjectName,
				BuildId: ptr.To(42),
			}).Return(mockBuildLogs, nil)
			mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).
				Return(nil, fmt.Errorf("HTTP request error"))

			_, err := pipelines.GetRunLogsWithBearerToken(ctx, mockBuildClient, mockHTTPClient, adoConfig, ptr.To(42), "test-token")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed http request getting build log"))
			mockBuildClient.AssertExpectations(GinkgoT())
			mockHTTPClient.AssertExpectations(GinkgoT())
		})
	})
})
