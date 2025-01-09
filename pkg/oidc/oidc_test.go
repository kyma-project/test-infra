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
	)

	BeforeEach(func() {
		l, err := zap.NewDevelopment()
		Expect(err).NotTo(HaveOccurred())

		logger = l.Sugar()

		trustedIssuers = map[string]tioidc.Issuer{
			"https://fakedings.dev-gcp.nais.io/fake": {
				Name:                   "github",
				IssuerURL:              "https://fakedings.dev-gcp.nais.io/fake",
				JWKSURL:                "https://fakedings.dev-gcp.nais.io/fake/jwks",
				ExpectedJobWorkflowRef: "kyma-project/test-infra/.github/workflows/verify-oidc-token.yml@refs/heads/main",
				ClientID:               "testClientID",
			},
		}
	})

	Describe("SkipExpiryCheck", func() {
		It("should set SkipExpiryCheck to true", func() {
			config := &tioidc.VerifierConfig{}
			option := tioidc.SkipExpiryCheck()
			err := option(config)
			Expect(err).NotTo(HaveOccurred())
			Expect(config.SkipExpiryCheck).To(BeTrue())
		})
	})

	Describe("NewVerifierConfig", func() {
		var (
			verifierConfigOption tioidc.VerifierConfigOption
		)

		BeforeEach(func() {
			rawToken, err = os.ReadFile("test-fixtures/raw-oidc-token")
			Expect(err).NotTo(HaveOccurred())

			tokenProcessor, err = tioidc.NewTokenProcessor(logger, trustedIssuers, string(rawToken))
			Expect(err).NotTo(HaveOccurred())
			Expect(tokenProcessor).NotTo(BeNil())
		})

		It("should create a new default VerifierConfig", func() {
			verifierConfig, err := tokenProcessor.NewVerifierConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(verifierConfig).To(BeAssignableToTypeOf(tioidc.VerifierConfig{}))
			Expect(verifierConfig.ClientID).To(Equal(trustedIssuers["https://fakedings.dev-gcp.nais.io/fake"].ClientID))
			Expect(verifierConfig.SupportedSigningAlgs).To(Equal(tioidc.SupportedSigningAlgorithms))
			Expect(verifierConfig.SkipExpiryCheck).To(BeFalse())
			Expect(verifierConfig.SkipClientIDCheck).To(BeFalse())
			Expect(verifierConfig.SkipIssuerCheck).To(BeFalse())
			Expect(verifierConfig.InsecureSkipSignatureCheck).To(BeFalse())
		})
		When("invalid VerifierConfigOption are provided", func() {
			It("should return an error", func() {
				verifierConfigOption = func(config *tioidc.VerifierConfig) error {
					return errors.New("invalid VerifierConfigOption")
				}
				verifierConfig, err := tokenProcessor.NewVerifierConfig(verifierConfigOption)
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
		})

		When("issuer is trusted", func() {
			It("should return a new TokenProcessor", func() {
				tokenProcessor, err = tioidc.NewTokenProcessor(logger, trustedIssuers, string(rawToken))
				Expect(err).NotTo(HaveOccurred())
				Expect(tokenProcessor).To(BeAssignableToTypeOf(tioidc.TokenProcessor{}))
			})
		})
		When("issuer with empty clientID is provided", func() {
			It("should return an error", func() {
				// Empty issuer clientID
				issuer := trustedIssuers["https://fakedings.dev-gcp.nais.io/fake"]
				issuer.ClientID = ""
				trustedIssuers["https://fakedings.dev-gcp.nais.io/fake"] = issuer

				tokenProcessor, err := tioidc.NewTokenProcessor(logger, trustedIssuers, string(rawToken))
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("trusted issuer clientID is empty"))
				Expect(tokenProcessor).To(Equal(tioidc.TokenProcessor{}))
			})
		})
		When("issuer is not trusted", func() {
			It("should return an error", func() {
				// Untrusted issuer
				trustedIssuers = map[string]tioidc.Issuer{
					"https://other-trusted.fakedings.dev-gcp.nais.io/fake": {
						Name:      "github",
						IssuerURL: "https://other-trusted.fakedings.dev-gcp.nais.io/fake",
						JWKSURL:   "https://other-trusted.fakedings.dev-gcp.nais.io/fake/jwks",
						ClientID:  "testClientID",
					},
				}

				tokenProcessor, err = tioidc.NewTokenProcessor(logger, trustedIssuers, string(rawToken))
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("issuer https://fakedings.dev-gcp.nais.io/fake is not trusted"))
				Expect(tokenProcessor).To(Equal(tioidc.TokenProcessor{}))
			})
		})
		When("no trustedIssuers are provided", func() {
			It("should return an error", func() {
				tokenProcessor, err := tioidc.NewTokenProcessor(logger, nil, string(rawToken))
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

				tokenProcessor, err := tioidc.NewTokenProcessor(logger, trustedIssuers, string(rawToken), invalidTokenProcessorOption)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed to apply TokenProcessorOption: invalid TokenProcessorOption"))
				Expect(tokenProcessor).To(Equal(tioidc.TokenProcessor{}))
			})
		})
		When("invalid raw token is provided", func() {
			It("should return an error", func() {
				tokenProcessor, err := tioidc.NewTokenProcessor(logger, trustedIssuers, "invalidToken")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed to get issuer from token: failed to parse oidc token: go-jose/go-jose: compact JWS format must have three parts"))
				Expect(tokenProcessor).To(Equal(tioidc.TokenProcessor{}))
			})
		})
	})

	Describe("TokenProcessor", func() {
		var (
			Token     oidcmocks.MockTokenInterface
			claims    tioidc.Claims
			token     tioidc.Token
			mockToken oidcmocks.MockClaimsReader
		)
		BeforeEach(func() {
			rawToken, err = os.ReadFile("test-fixtures/raw-oidc-token")
			Expect(err).NotTo(HaveOccurred())

			tokenProcessor, err = tioidc.NewTokenProcessor(logger, trustedIssuers, string(rawToken))
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

		Describe("WithExtendedExpiration option", func() {

			It("should set expiration time", func() {
				expirationTime := 1
				option := tioidc.WithExtendedExpiration(expirationTime)
				err := option(&tokenVerifier)
				Expect(err).NotTo(HaveOccurred())
				Expect(tokenVerifier.ExpirationTimeMinutes).To(Equal(expirationTime))
			})
		})

		Describe("Verify", func() {
			It("should return Token when the token is valid with standard expiration time", func() {
				// prepare
				issuedAt := time.Now().Add(-4 * time.Minute)
				idToken := &oidc.IDToken{IssuedAt: issuedAt}
				tokenVerifier.Config.SkipExpiryCheck = false
				tokenVerifier.ExpirationTimeMinutes = 10
				verifier.On("Verify", mock.AnythingOfType("backgroundCtx"), string(rawToken)).Return(idToken, nil)

				// execute
				token, err = tokenVerifier.Verify(ctx, string(rawToken))

				// verify
				Expect(err).NotTo(HaveOccurred())
				Expect(token).To(Equal(&tioidc.Token{Token: idToken, IssuedAt: issuedAt}))
			})

			It("should return Token when the token is valid with extended expiration time", func() {
				// prepare
				issuedAt := time.Now().Add(-9 * time.Minute)
				idToken := &oidc.IDToken{IssuedAt: issuedAt}
				verifier.On("Verify", mock.AnythingOfType("backgroundCtx"), string(rawToken)).Return(idToken, nil)
				tokenVerifier.Config.SkipExpiryCheck = true
				tokenVerifier.ExpirationTimeMinutes = 10

				// execute
				token, err = tokenVerifier.Verify(ctx, string(rawToken))

				// verify
				Expect(err).NotTo(HaveOccurred())
				Expect(token).To(Equal(&tioidc.Token{Token: idToken, IssuedAt: issuedAt}))
			})

			It("should return an error when the token is invalid with standard expiration time", func() {
				// prepare
				verifier.On("Verify", mock.AnythingOfType("backgroundCtx"), string(rawToken)).Return(nil, errors.New("invalid token"))
				tokenVerifier.Config.SkipExpiryCheck = false
				tokenVerifier.ExpirationTimeMinutes = 10

				// execute
				token, err = tokenVerifier.Verify(ctx, string(rawToken))

				// verify
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed to verify token: invalid token"))
				Expect(token).To(BeNil())
			})

			It("should return an error when the token is invalid with extended expiration time", func() {
				// prepare
				verifier.On("Verify", mock.AnythingOfType("backgroundCtx"), string(rawToken)).Return(nil, errors.New("invalid token"))
				tokenVerifier.Config.SkipExpiryCheck = true
				tokenVerifier.ExpirationTimeMinutes = 10

				// execute
				token, err = tokenVerifier.Verify(ctx, string(rawToken))

				// verify
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed to verify token: invalid token"))
				Expect(token).To(BeNil())
			})

			It("should return an error when token expired with standard expiration check", func() {
				// prepare
				// issuedAt := time.Now().Add(-6 * time.Minute)
				// idToken := &oidc.IDToken{IssuedAt: issuedAt}
				verifier.On("Verify", mock.AnythingOfType("backgroundCtx"), string(rawToken)).Return(nil, errors.New("token expired"))
				tokenVerifier.Config.SkipExpiryCheck = false
				tokenVerifier.ExpirationTimeMinutes = 10

				// execute
				token, err = tokenVerifier.Verify(ctx, string(rawToken))

				// verify
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed to verify token: token expired"))
				Expect(token).To(BeNil())
			})

			It("should return an error when token expired with extended expiration check", func() {
				// prepare
				issuedAt := time.Now().Add(-11 * time.Minute)
				idToken := &oidc.IDToken{IssuedAt: issuedAt}
				verifier.On("Verify", mock.AnythingOfType("backgroundCtx"), string(rawToken)).Return(idToken, nil)
				tokenVerifier.Config.SkipExpiryCheck = true
				tokenVerifier.ExpirationTimeMinutes = 10

				// execute
				token, err = tokenVerifier.Verify(ctx, string(rawToken))

				// verify
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(fmt.Errorf("failed to verify token: %w",
					fmt.Errorf("token expired, tokenIssuedAt: %v, expired at %v", issuedAt, issuedAt.Add(time.Minute*time.Duration(tokenVerifier.ExpirationTimeMinutes))),
				)))
				Expect(token).To(BeNil())
			})
		})
	})

	Describe("Token", func() {

		Describe("IsTokenExpired", func() {
			var (
				now                   time.Time
				expirationTimeMinutes int
			)

			BeforeEach(func() {
				now = time.Now()
				expirationTimeMinutes = 1
			})

			It("should return no error when the token is within the expiration time", func() {
				token := tioidc.Token{
					IssuedAt: now,
				}
				err := token.IsTokenExpired(logger, expirationTimeMinutes)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return an error when the token is expired", func() {
				token := tioidc.Token{
					IssuedAt: now.Add(-2 * time.Minute), // 2 minutes ago
				}
				extendedExpiration := token.IssuedAt.Add(time.Minute * time.Duration(expirationTimeMinutes))
				err := token.IsTokenExpired(logger, expirationTimeMinutes)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(fmt.Errorf("token expired, tokenIssuedAt: %v, expired at %v", token.IssuedAt, extendedExpiration)))
			})

			It("should return an error when the token issued in the future", func() {
				token := tioidc.Token{
					IssuedAt: now.Add(2 * time.Minute), // 2 minutes ago
				}
				err := token.IsTokenExpired(logger, expirationTimeMinutes)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(MatchRegexp("token issued in the future, tokenIssuedAt: .+, now: .+")))
			})
		})
	})

	Describe("Provider", func() {
		var (
			provider     tioidc.Provider
			oidcProvider *oidcmocks.MockVerifierProvider
		)
		BeforeEach(func() {
			oidcProvider = &oidcmocks.MockVerifierProvider{}
			provider = tioidc.Provider{
				VerifierProvider: oidcProvider,
			}

			rawToken, err = os.ReadFile("test-fixtures/raw-oidc-token")
			Expect(err).NotTo(HaveOccurred())

			tokenProcessor, err = tioidc.NewTokenProcessor(logger, trustedIssuers, string(rawToken))
			Expect(err).NotTo(HaveOccurred())
			Expect(tokenProcessor).NotTo(BeNil())

			verifierConfig, err = tokenProcessor.NewVerifierConfig()
			Expect(err).NotTo(HaveOccurred())
		})
		Describe("NewVerifier", func() {
			It("should return a new TokenVerifier", func() {
				oidcProvider.On("Verifier", &verifierConfig.Config).Return(&oidc.IDTokenVerifier{})
				verifier, err := provider.NewVerifier(logger, verifierConfig)
				Expect(err).NotTo(HaveOccurred())
				Expect(verifier).NotTo(BeNil())
				Expect(verifier).To(BeAssignableToTypeOf(tioidc.TokenVerifier{}))
			})
		})
	})
})
