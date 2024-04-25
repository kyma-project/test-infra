package oidc_test

import (
	"errors"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	tioidc "github.com/kyma-project/test-infra/pkg/oidc"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

var _ = Describe("OIDC", func() {
	var (
		// ctx            context.Context
		logger         *zap.SugaredLogger
		trustedIssuers map[string]tioidc.Issuer
		rawToken       []byte
		config         *oidc.Config
		options        []tioidc.TokenProcessorOption
		// verifier       *oidcmocks.MockTokenVerifierInterface
		tokenProcessor *tioidc.TokenProcessor
	)

	BeforeEach(func() {
		// ctx = context.Background()
		l, err := zap.NewDevelopment()
		Expect(err).NotTo(HaveOccurred())
		logger = l.Sugar()

		trustedIssuers = tioidc.TrustedOIDCIssuers
		options = []tioidc.TokenProcessorOption{}
		// verifier = new(oidcmocks.MockTokenVerifierInterface)
	})

	Describe("NewVerifierConfig", func() {
		var (
			clientID             string
			verifierConfigOption tioidc.VerifierConfigOption
		)
		BeforeEach(func() {
			clientID = "testClientID"
		})
		It("should return a new oidc.Config", func() {
			verifierConfig, err := tioidc.NewVerifierConfig(logger, clientID)
			Expect(err).NotTo(HaveOccurred())
			Expect(verifierConfig).To(BeAssignableToTypeOf(&oidc.Config{}))
			Expect(verifierConfig.ClientID).To(Equal(clientID))
		})
		When("invalid VerifierConfigOption are provided", func() {
			It("should return an error", func() {
				verifierConfigOption = func(config *oidc.Config) error {
					return errors.New("invalid VerifierConfigOption")
				}
				verifierConfig, err := tioidc.NewVerifierConfig(logger, clientID, verifierConfigOption)
				Expect(err).To(HaveOccurred())
				Expect(verifierConfig).To(BeNil())
			})
		})
		When("empty clientID is provided", func() {
			It("should return an error", func() {
				verifierConfig, err := tioidc.NewVerifierConfig(logger, "")
				Expect(err).To(HaveOccurred())
				Expect(verifierConfig).To(BeNil())
			})
		})
	})

	Describe("NewTokenProcessor", func() {
		When("issuer is trusted", func() {
			It("should return a new TokenProcessor", func() {
				var err error
				// Read the token from the file in test-fixtures directory.
				rawToken, err = os.ReadFile("test-fixtures/raw-oidc-token")
				Expect(err).NotTo(HaveOccurred())
				trustedIssuers = map[string]tioidc.Issuer{
					"https://fakedings.dev-gcp.nais.io/fake": {
						Name:      "github",
						IssuerURL: "https://fakedings.dev-gcp.nais.io/fake",
						JWKSURL:   "https://fakedings.dev-gcp.nais.io/fake/jwks",
					},
				}
				config, err = tioidc.NewVerifierConfig(logger, "testClientID")
				Expect(err).NotTo(HaveOccurred())

				tokenProcessor, err = tioidc.NewTokenProcessor(logger, trustedIssuers, string(rawToken), *config, options...)
				Expect(err).NotTo(HaveOccurred())
				Expect(tokenProcessor).To(BeAssignableToTypeOf(&tioidc.TokenProcessor{}))
			})
		})
		When("issuer is not trusted", func() {
			It("should return an error", func() {
				var err error
				// Read the token from the file in test-fixtures directory.
				rawToken, err = os.ReadFile("test-fixtures/raw-oidc-token")
				Expect(err).NotTo(HaveOccurred())
				trustedIssuers = map[string]tioidc.Issuer{
					"https://untrusted.fakedings.dev-gcp.nais.io/fake": {
						Name:      "github",
						IssuerURL: "https://fakedings.dev-gcp.nais.io/fake",
						JWKSURL:   "https://fakedings.dev-gcp.nais.io/fake/jwks",
					},
				}
				config, err = tioidc.NewVerifierConfig(logger, "testClientID")
				Expect(err).NotTo(HaveOccurred())

				tokenProcessor, err = tioidc.NewTokenProcessor(logger, trustedIssuers, string(rawToken), *config, options...)
				Expect(err).To(HaveOccurred())
				Expect(tokenProcessor).To(BeNil())
			})
		})
		When("invalid TokenProcessorOption are provided", func() {
			It("should return an error", func() {
				invalidOption := func(tp *tioidc.TokenProcessor) error {
					return errors.New("invalid TokenProcessorOoption")
				}
				tokenProcessor, err := tioidc.NewTokenProcessor(logger, trustedIssuers, "rawToken", *config, invalidOption)
				Expect(err).To(HaveOccurred())
				Expect(tokenProcessor).To(BeNil())
			})
		})
		When("invalid raw token is provided", func() {
			It("should return an error", func() {
				tokenProcessor, err := tioidc.NewTokenProcessor(logger, trustedIssuers, "invalidToken", *config, options...)
				Expect(err).To(HaveOccurred())
				Expect(tokenProcessor).To(BeNil())
			})
		})
	})

	// Describe("VerifyToken", func() {
	// 	It("should return no error when the token is valid", func() {
	// 		tokenProcessor, _ := tioidc.NewTokenProcessor(logger, trustedIssuers, rawToken, config, options...)
	// 		verifier.EXPECT().Verify(ctx, rawToken).Return(nil, nil)
	// 		err := tokenProcessor.VerifyToken(ctx, verifier)
	// 		Expect(err).NotTo(HaveOccurred())
	// 	})
	//
	// 	It("should return an error when the token is invalid", func() {
	// 		tokenProcessor, _ := tioidc.NewTokenProcessor(logger, trustedIssuers, rawToken, config, options...)
	// 		verifier.EXPECT().Verify(ctx, rawToken).Return(nil, errors.New("invalid token"))
	// 		err := tokenProcessor.VerifyToken(ctx, verifier)
	// 		Expect(err).To(HaveOccurred())
	// 	})
	// })
})
