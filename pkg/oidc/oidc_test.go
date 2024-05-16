package oidc_test

import (
	"errors"
	"fmt"

	// "time"

	// "fmt"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-jose/go-jose/v4/jwt"
	tioidc "github.com/kyma-project/test-infra/pkg/oidc"
	oidcmocks "github.com/kyma-project/test-infra/pkg/oidc/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

var _ = Describe("OIDC", func() {
	var (
		err            error
		ctx            context.Context
		logger         *zap.SugaredLogger
		trustedIssuers map[string]tioidc.Issuer
		rawToken       []byte
		verifierConfig tioidc.VerifierConfig
		tokenProcessor tioidc.TokenProcessor
		clientID       string
	)

	BeforeEach(func() {
		l, err := zap.NewDevelopment()
		Expect(err).NotTo(HaveOccurred())

		logger = l.Sugar()
		clientID = "testClientID"
	})

	Describe("NewVerifierConfig", func() {
		var (
			verifierConfigOption tioidc.VerifierConfigOption
		)
		It("should return a new oidc.Config", func() {
			verifierConfig, err := tioidc.NewVerifierConfig(logger, clientID)
			Expect(err).NotTo(HaveOccurred())
			Expect(verifierConfig).To(BeAssignableToTypeOf(tioidc.VerifierConfig{}))
			Expect(verifierConfig.ClientID).To(Equal(clientID))
			Expect(verifierConfig.SupportedSigningAlgs).To(Equal(tioidc.SupportedSigningAlgorithms))
		})
		When("invalid VerifierConfigOption are provided", func() {
			It("should return an error", func() {
				verifierConfigOption = func(config *tioidc.VerifierConfig) error {
					return errors.New("invalid VerifierConfigOption")
				}
				verifierConfig, err := tioidc.NewVerifierConfig(logger, clientID, verifierConfigOption)
				Expect(err).To(HaveOccurred())
				Expect(verifierConfig).To(Equal(tioidc.VerifierConfig{}))
			})
		})
		When("empty clientID is provided", func() {
			It("should return an error", func() {
				verifierConfig, err := tioidc.NewVerifierConfig(logger, "")
				Expect(err).To(HaveOccurred())
				Expect(verifierConfig).To(Equal(tioidc.VerifierConfig{}))
			})
		})
	})

	Describe("NewTokenProcessor", func() {
		var (
			invalidTokenProcessorOption tioidc.TokenProcessorOption
		)

		BeforeEach(func() {
			// Read the token from the file in test-fixtures directory.
			rawToken, err = os.ReadFile("test-fixtures/raw-oidc-token")
			Expect(err).NotTo(HaveOccurred())

			verifierConfig, err = tioidc.NewVerifierConfig(logger, clientID)
			Expect(err).NotTo(HaveOccurred())

			trustedIssuers = map[string]tioidc.Issuer{
				"https://fakedings.dev-gcp.nais.io/fake": {
					Name:      "github",
					IssuerURL: "https://fakedings.dev-gcp.nais.io/fake",
					JWKSURL:   "https://fakedings.dev-gcp.nais.io/fake/jwks",
				},
			}
			ctx = context.Background()
		})
		When("issuer is trusted", func() {
			It("should return a new TokenProcessor", func() {
				tokenProcessor, err = tioidc.NewTokenProcessor(logger, trustedIssuers, string(rawToken), verifierConfig)
				Expect(err).NotTo(HaveOccurred())
				Expect(tokenProcessor).To(BeAssignableToTypeOf(tioidc.TokenProcessor{}))
			})
		})
		When("empty verifierConfig is provided", func() {
			It("should return an error", func() {
				// Empty verifierConfig
				verifierConfig = tioidc.VerifierConfig{}
				tokenProcessor, err := tioidc.NewTokenProcessor(logger, trustedIssuers, string(rawToken), verifierConfig)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("verifierConfig clientID is empty"))
				Expect(tokenProcessor).To(Equal(tioidc.TokenProcessor{}))
			})
		})
		When("issuer is not trusted", func() {
			It("should return an error", func() {
				// Untrusted issuer
				trustedIssuers = map[string]tioidc.Issuer{
					"https://untrusted.fakedings.dev-gcp.nais.io/fake": {
						Name:      "github",
						IssuerURL: "https://untrusted.fakedings.dev-gcp.nais.io/fake",
						JWKSURL:   "https://untrusted.fakedings.dev-gcp.nais.io/fake/jwks",
					},
				}

				tokenProcessor, err = tioidc.NewTokenProcessor(logger, trustedIssuers, string(rawToken), verifierConfig)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("issuer https://fakedings.dev-gcp.nais.io/fake is not trusted"))
				Expect(tokenProcessor).To(Equal(tioidc.TokenProcessor{}))
			})
		})
		When("no trustedIssuers are provided", func() {
			It("should return an error", func() {
				tokenProcessor, err := tioidc.NewTokenProcessor(logger, nil, string(rawToken), verifierConfig)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("issuer https://fakedings.dev-gcp.nais.io/fake is not trusted"))
				Expect(tokenProcessor).To(Equal(tioidc.TokenProcessor{}))
			})
		})
		When("invalid TokenProcessorOption is provided", func() {
			It("should return an error", func() {
				invalidTokenProcessorOption = func(tp *tioidc.TokenProcessor) error {
					return errors.New("invalid TokenProcessorOption")
				}

				tokenProcessor, err := tioidc.NewTokenProcessor(logger, trustedIssuers, string(rawToken), verifierConfig, invalidTokenProcessorOption)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed to apply TokenProcessorOption: invalid TokenProcessorOption"))
				Expect(tokenProcessor).To(Equal(tioidc.TokenProcessor{}))
			})
		})
		When("invalid raw token is provided", func() {
			It("should return an error", func() {
				tokenProcessor, err := tioidc.NewTokenProcessor(logger, trustedIssuers, "invalidToken", verifierConfig)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed to get issuer from token: failed to parse oidc token: go-jose/go-jose: compact JWS format must have three parts"))
				Expect(tokenProcessor).To(Equal(tioidc.TokenProcessor{}))
			})
		})
	})

	Describe("TokenProcessor", func() {
		var (
			verifier       *oidcmocks.MockTokenVerifierInterface
			Token          oidcmocks.MockTokenInterface
			claims         tioidc.Claims
			token          tioidc.Token
			mockToken      oidcmocks.MockClaimsReader
			tokenProcessor tioidc.TokenProcessor
		)
		BeforeEach(func() {
			verifierConfig, err = tioidc.NewVerifierConfig(logger, clientID)
			Expect(err).NotTo(HaveOccurred())

			rawToken, err = os.ReadFile("test-fixtures/raw-oidc-token")
			Expect(err).NotTo(HaveOccurred())

			tokenProcessor, err = tioidc.NewTokenProcessor(logger, trustedIssuers, string(rawToken), verifierConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(tokenProcessor).NotTo(BeNil())

			ctx = context.Background()

			trustedIssuers = map[string]tioidc.Issuer{
				"https://fakedings.dev-gcp.nais.io/fake": {
					Name:                   "github",
					IssuerURL:              "https://fakedings.dev-gcp.nais.io/fake",
					JWKSURL:                "https://fakedings.dev-gcp.nais.io/fake/jwks",
					ExpectedJobWorkflowRef: "kyma-project/test-infra/.github/workflows/verify-oidc-token.yml@refs/heads/main",
				},
			}
			token = tioidc.Token{}
			mockToken = oidcmocks.MockClaimsReader{}
			verifier = &oidcmocks.MockTokenVerifierInterface{}
		})
		Describe("Issuer", func() {
			It("should return the issuer", func() {
				issuer := tokenProcessor.Issuer()
				Expect(issuer).To(Equal("https://fakedings.dev-gcp.nais.io/fake"))
			})
		})
		Describe("VerifyAndExtractClaims", func() {
			BeforeEach(func() {
				claims = tioidc.NewClaims(logger)
			})
			It("should return no error when the token is valid", func() {
				mockToken.On(
					"Claims", &claims).Run(
					func(args mock.Arguments) {
						arg := args.Get(0).(*tioidc.Claims)
						arg.Issuer = "https://fakedings.dev-gcp.nais.io/fake"
						arg.Subject = "mysub"
						arg.Audience = jwt.Audience{"myaudience"}
						arg.JobWorkflowRef = "kyma-project/test-infra/.github/workflows/verify-oidc-token.yml@refs/heads/main"
					},
				).Return(nil)
				token.Token = &mockToken
				verifier.On("Verify", mock.AnythingOfType("backgroundCtx"), string(rawToken)).Return(token, nil)

				// Run
				err = tokenProcessor.VerifyAndExtractClaims(ctx, verifier, &claims)

				// Verify
				Expect(err).NotTo(HaveOccurred())
				Expect(claims.Issuer).To(Equal("https://fakedings.dev-gcp.nais.io/fake"))
				Expect(claims.Subject).To(Equal("mysub"))
				Expect(claims.Audience).To(Equal(jwt.Audience{"myaudience"}))
			})
			It("should return an error when unexpected job workflow reference is provided", func() {
				mockToken.On(
					"Claims", &claims).Run(
					func(args mock.Arguments) {
						arg := args.Get(0).(*tioidc.Claims)
						arg.Issuer = "https://fakedings.dev-gcp.nais.io/fake"
						arg.Subject = "mysub"
						arg.Audience = jwt.Audience{"myaudience"}
						// Unexpected job workflow reference
						arg.JobWorkflowRef = "kyma-project/test-infra/.github/workflows/unexpected.yml@refs/heads/main"
					},
				).Return(nil)
				token.Token = &mockToken
				verifier.On("Verify", mock.AnythingOfType("backgroundCtx"), string(rawToken)).Return(token, nil)

				// Run
				err = tokenProcessor.VerifyAndExtractClaims(ctx, verifier, &claims)

				// Verify
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed to validate claims: job_workflow_ref claim expected value validation failed, expected: kyma-project/test-infra/.github/workflows/unexpected.yml@refs/heads/main, provided: kyma-project/test-infra/.github/workflows/verify-oidc-token.yml@refs/heads/main"))
			})
			It("should return an error when token was not verified", func() {
				verifier.On("Verify", mock.AnythingOfType("backgroundCtx"), string(rawToken)).Return(token, fmt.Errorf("token validation failed"))

				// Run
				err = tokenProcessor.VerifyAndExtractClaims(ctx, verifier, &claims)

				// Verify
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed to verify token: token validation failed"))
				Expect(claims).To(Equal(tioidc.NewClaims(logger)))
			})
			It("should return an error when claims are not set", func() {
				mockToken.On("Claims", &claims).Return(fmt.Errorf("claims are not set"))
				token.Token = &mockToken
				verifier.On("Verify", mock.AnythingOfType("backgroundCtx"), string(rawToken)).Return(token, nil)
				Token.On("Claims", &claims).Return(fmt.Errorf("claims are not set"))

				// Run
				err = tokenProcessor.VerifyAndExtractClaims(ctx, verifier, &claims)

				// Verify
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed to get claims from token: claims are not set"))
				Expect(claims).To(Equal(tioidc.NewClaims(logger)))
			})
		})
	})

	Describe("TokenVerifier", func() {
		var (
			tokenVerifier tioidc.TokenVerifier
			verifier      *oidcmocks.MockVerifier
			ctx           context.Context
			token         tioidc.Token
		)
		BeforeEach(func() {
			verifier = &oidcmocks.MockVerifier{}
			tokenVerifier = tioidc.TokenVerifier{
				Verifier: verifier,
				Logger:   logger,
			}
			ctx = context.Background()
			rawToken, err = os.ReadFile("test-fixtures/raw-oidc-token")
			Expect(err).NotTo(HaveOccurred())
		})
		Describe("Verify", func() {
			It("should return Token when the token is valid", func() {
				verifier.On("Verify", mock.AnythingOfType("backgroundCtx"), string(rawToken)).Return(&oidc.IDToken{}, nil)
				token, err = tokenVerifier.Verify(ctx, string(rawToken))
				Expect(err).NotTo(HaveOccurred())
				Expect(token).To(BeAssignableToTypeOf(tioidc.Token{}))
			})
		})
	})

	Describe("Provider", func() {
		var (
			provider       tioidc.Provider
			oidcProvider   *oidcmocks.MockVerifierProvider
			verifierConfig tioidc.VerifierConfig
		)
		BeforeEach(func() {
			oidcProvider = &oidcmocks.MockVerifierProvider{}
			provider = tioidc.Provider{
				VerifierProvider: oidcProvider,
			}
			ctx = context.Background()
			verifierConfig, err = tioidc.NewVerifierConfig(logger, clientID)
			Expect(err).NotTo(HaveOccurred())
		})
		Describe("NewVerifier", func() {
			It("should return a new TokenVerifier", func() {
				oidcProvider.On("Verifier", &verifierConfig.Config).Return(&oidc.IDTokenVerifier{})
				verifier := provider.NewVerifier(logger, verifierConfig)
				Expect(verifier).NotTo(BeNil())
				Expect(verifier).To(BeAssignableToTypeOf(tioidc.TokenVerifier{}))
			})
		})
	})
})
