package oidc

// oidc_unit_test.go contains tests which require access to non-exported functions and variables.

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

var _ = Describe("OIDC", func() {
	var (
		logger *zap.SugaredLogger
	)
	BeforeEach(func() {
		l, err := zap.NewDevelopment()
		Expect(err).NotTo(HaveOccurred())

		logger = l.Sugar()
	})

	Describe("maskToken", func() {
		It("should mask the token if length is less than 15", func() {
			token := "shorttoken"
			maskedToken := maskToken(token)
			Expect(maskedToken).To(Equal("********"))
		})

		It("should mask the token if length is exactly 15", func() {
			token := "123456789012345"
			maskedToken := maskToken(token)
			Expect(maskedToken).To(Equal("12********45"))
		})

		It("should mask the token if length is greater than 15", func() {
			token := "12345678901234567890"
			maskedToken := maskToken(token)
			Expect(maskedToken).To(Equal("12********90"))
		})

		It("should mask the token if it's empty", func() {
			token := ""
			maskedToken := maskToken(token)
			Expect(maskedToken).To(Equal("********"))
		})
	})

	Describe("NewVerifierConfig", func() {
		var (
			tokenProcessor TokenProcessor
			err            error
			trustedIssuers map[string]Issuer
			rawToken       []byte
		)

		BeforeEach(func() {
			rawToken, err = os.ReadFile("test-fixtures/raw-oidc-token")
			Expect(err).NotTo(HaveOccurred())

			trustedIssuers = map[string]Issuer{
				"https://fakedings.dev-gcp.nais.io/fake": {
					Name:                   "github",
					IssuerURL:              "https://fakedings.dev-gcp.nais.io/fake",
					JWKSURL:                "https://fakedings.dev-gcp.nais.io/fake/jwks",
					ExpectedJobWorkflowRef: "kyma-project/test-infra/.github/workflows/verify-oidc-token.yml@refs/heads/main",
					ClientID:               "testClientID",
				},
			}

			tokenProcessor, err = NewTokenProcessor(logger, trustedIssuers, string(rawToken))
			Expect(err).NotTo(HaveOccurred())
			Expect(tokenProcessor).NotTo(BeNil())
		})

		When("empty clientID is provided", func() {
			It("should return an error", func() {
				tokenProcessor.issuer.ClientID = ""
				verifierConfig, err := tokenProcessor.NewVerifierConfig()
				Expect(err).To(HaveOccurred())
				Expect(verifierConfig).To(Equal(VerifierConfig{}))
			})
		})
	})
})
