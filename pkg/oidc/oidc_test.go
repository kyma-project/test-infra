package oidc_test

import (
	"errors"
	"fmt"
	"time"

	// "fmt"
	"os"

	// "time"

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
		err error

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

			trustedIssuers = map[string]tioidc.Issuer{
				"https://fakedings.dev-gcp.nais.io/fake": {
					Name:                   "github",
					IssuerURL:              "https://fakedings.dev-gcp.nais.io/fake",
					JWKSURL:                "https://fakedings.dev-gcp.nais.io/fake/jwks",
					ExpectedJobWorkflowRef: "kyma-project/test-infra/.github/workflows/verify-oidc-token.yml@refs/heads/main",
				},
			}

			tokenProcessor, err = tioidc.NewTokenProcessor(logger, trustedIssuers, string(rawToken), verifierConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(tokenProcessor).NotTo(BeNil())

			token = tioidc.Token{}
			mockToken = oidcmocks.MockClaimsReader{}
		})
		Describe("Issuer", func() {
			It("should return the issuer", func() {
				issuer := tokenProcessor.Issuer()
				Expect(issuer).To(Equal("https://fakedings.dev-gcp.nais.io/fake"))
			})
		})
		Describe("ValidateClaims", func() {
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

				// Run
				err = tokenProcessor.ValidateClaims(&claims, &token)

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
				expectedError := fmt.Errorf("expecations validation failed: %w", fmt.Errorf("job_workflow_ref claim expected value validation failed, expected: kyma-project/test-infra/.github/workflows/unexpected.yml@refs/heads/main, provided: kyma-project/test-infra/.github/workflows/verify-oidc-token.yml@refs/heads/main"))
				token.Token = &mockToken

				// Run
				err = tokenProcessor.ValidateClaims(&claims, &token)

				// Verify
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(expectedError))
			})
			It("should return an error when claims are not set", func() {
				mockToken.On("Claims", &claims).Return(fmt.Errorf("claims are not set"))
				token.Token = &mockToken
				Token.On("Claims", &claims).Return(fmt.Errorf("claims are not set"))

				// Run
				err = tokenProcessor.ValidateClaims(&claims, &token)

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
			token         *tioidc.Token
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
				Expect(token).To(BeAssignableToTypeOf(&tioidc.Token{}))
			})

			It("should return an error when the token is invalid", func() {
				verifier.On("Verify", mock.AnythingOfType("backgroundCtx"), string(rawToken)).Return(&oidc.IDToken{}, errors.New("invalid token"))
				token, err = tokenVerifier.Verify(ctx, string(rawToken))
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed to verify token: invalid token"))
				Expect(token).To(Equal(&tioidc.Token{}))
			})
		})
		Describe("VerifyExtendedExpiration", func() {
			var (
				expirationTimestamp time.Time
				gracePeriodMinutes  int
			)

			BeforeEach(func() {
				gracePeriodMinutes = 1
			})

			It("should return no error when the token is within the grace period", func() {
				expirationTimestamp = time.Now().Add(-30 * time.Second)
				err := tokenVerifier.VerifyExtendedExpiration(expirationTimestamp, gracePeriodMinutes)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return an error when the token is expired beyond the grace period", func() {
				expirationTimestamp = time.Now().Add(-2 * time.Minute)
				err := tokenVerifier.VerifyExtendedExpiration(expirationTimestamp, gracePeriodMinutes)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("token expired more than 1m0s ago"))
			})

			It("should return an error when the token expiration timestamp is in the future", func() {
				expirationTimestamp = time.Now().Add(1 * time.Minute)
				err := tokenVerifier.VerifyExtendedExpiration(expirationTimestamp, gracePeriodMinutes)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(fmt.Sprintf("token expiration time is in the future: %v", expirationTimestamp)))
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
