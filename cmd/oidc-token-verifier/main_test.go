package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kyma-project/test-infra/cmd/oidc-token-verifier/mocks"
	tioidc "github.com/kyma-project/test-infra/pkg/oidc"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

var _ = Describe("Output", func() {
	var (
		logger         Logger
		issuerProvider *mainmocks.MockTrustedIssuerProvider
		issuer         tioidc.Issuer
		out            output
	)

	BeforeEach(func() {
		logger = zap.NewNop().Sugar()
		issuerProvider = &mainmocks.MockTrustedIssuerProvider{}
		issuer = tioidc.Issuer{
			Name:                   "test-issuer",
			IssuerURL:              "https://test-issuer.com",
			JWKSURL:                "https://test-issuer.com/jwks",
			ExpectedJobWorkflowRef: "test-workflow",
			GithubURL:              "https://github-test.com",
			ClientID:               "test-client-id",
		}
		out = output{}
	})

	Describe("setGithubURLOutput", func() {
		Context("when the Github URL is found in the tokenProcessor trusted issuer", func() {
			It("should set the Github URL in the output struct", func() {
				issuerProvider.On("GetIssuer").Return(issuer)

				err := out.setGithubURLOutput(logger, issuerProvider)
				Expect(err).NotTo(HaveOccurred(), "Expected no error, but got: %v", err)
				Expect(out.GithubURL).To(Equal(issuer.GithubURL), "Expected Github URL to be %s, but got %s", issuer.GithubURL, out.GithubURL)
			})
		})

		Context("when the Github URL is not found in the tokenProcessor trusted issuer", func() {
			It("should return an error", func() {
				issuer.GithubURL = ""
				issuerProvider.On("GetIssuer").Return(issuer)

				err := out.setGithubURLOutput(logger, issuerProvider)
				Expect(err).To(HaveOccurred(), "Expected an error, but got none")
				Expect(err).To(MatchError(fmt.Errorf("github URL not found in the tokenProcessor trusted issuer: %s", issuerProvider.GetIssuer())), "Expected error message to be 'github URL not found in the tokenProcessor trusted issuer', but got %s", err)
				Expect(out.GithubURL).To(BeEmpty(), "Expected Github URL to be empty, but got %v", out.GithubURL)
			})
		})
	})

	Describe("setClientIDOutput", func() {
		Context("when the Client ID is found in the tokenProcessor trusted issuer", func() {
			It("should set the Client ID in the output struct", func() {
				issuerProvider.On("GetIssuer").Return(issuer)

				err := out.setClientIDOutput(logger, issuerProvider)
				Expect(err).NotTo(HaveOccurred(), "Expected no error, but got: %v", err)
				Expect(out.ClientID).To(Equal(issuer.ClientID), "Expected Client ID to be %s, but got %s", issuer.ClientID, out.ClientID)
			})
		})

		Context("when the Client ID is not found in the tokenProcessor trusted issuer", func() {
			It("should return an error", func() {
				issuer.ClientID = ""
				issuerProvider.On("GetIssuer").Return(issuer)

				err := out.setClientIDOutput(logger, issuerProvider)
				Expect(err).To(HaveOccurred(), "Expected an error, but got none")
				Expect(err).To(MatchError(fmt.Errorf("client ID not found in the tokenProcessor trusted issuer: %s", issuerProvider.GetIssuer())), "Expected error message to be 'client ID not found in the tokenProcessor trusted issuer', but got %s", err)
				Expect(out.ClientID).To(BeEmpty(), "Expected Client ID to be empty, but got %v", out.ClientID)
			})
		})
	})

	Describe("writeOutputFile", func() {
		var filePath = "./output.json"

		BeforeEach(func() {
			// Verify if the path exists and is writable
			file, err := os.Create(filePath)
			Expect(err).NotTo(HaveOccurred(), "Expected no error creating the file, but got: %v", err)
			file.Close()
		})

		AfterEach(func() {
			// Remove created artifacts
			err := os.Remove(filePath)
			Expect(err).NotTo(HaveOccurred(), "Expected no error removing the file, but got: %v", err)
		})

		Context("when the output file is successfully written", func() {
			It("should write the output values to the json file", func() {
				out.GithubURL = issuer.GithubURL
				out.ClientID = issuer.ClientID

				err := out.writeOutputFile(logger, filePath)
				Expect(err).NotTo(HaveOccurred(), "Expected no error, but got: %v", err)

				file, err := os.Open(filePath)
				Expect(err).NotTo(HaveOccurred(), "Expected no error opening the file, but got: %v", err)
				defer file.Close()

				var writtenOutput output
				err = json.NewDecoder(file).Decode(&writtenOutput)
				Expect(err).NotTo(HaveOccurred(), "Expected no error decoding the file, but got: %v", err)
				Expect(writtenOutput).To(Equal(out), "Expected written output to be %v, but got %v", out, writtenOutput)
			})
		})

		Context("when there is an error creating the output file", func() {
			It("should return an error", func() {
				filePath := "/invalid-path/output.json"

				err := out.writeOutputFile(logger, filePath)
				Expect(err).To(HaveOccurred(), "Expected an error, but got none")
			})
		})
	})
})
# (2025-03-04)