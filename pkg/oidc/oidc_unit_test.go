package oidc

// oidc_unit_test.go contains tests which require access to non-exported functions and variables.

import (
	"fmt"
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

	Describe("TokenProcessor", func() {
		var (
			tokenProcessor TokenProcessor
			trustedIssuers map[string]Issuer
			rawToken       []byte
			mockIssuer     Issuer
			err            error
		)

		BeforeEach(func() {
			mockIssuer = Issuer{
				Name:                   "github",
				IssuerURL:              "https://fakedings.dev-gcp.nais.io/fake",
				JWKSURL:                "https://fakedings.dev-gcp.nais.io/fake/jwks",
				ExpectedJobWorkflowRef: "kyma-project/test-infra/.github/workflows/verify-oidc-token.yml@refs/heads/main",
				ClientID:               "testClientID",
			}

			trustedIssuers = map[string]Issuer{
				mockIssuer.IssuerURL: mockIssuer,
			}

			rawToken, err = os.ReadFile("test-fixtures/raw-oidc-token")
			Expect(err).NotTo(HaveOccurred())

			tokenProcessor = TokenProcessor{
				logger:         logger,
				trustedIssuers: trustedIssuers,
				rawToken:       string(rawToken),
			}
		})

		// This NewVerifierConfig scenario is tested here because it requires an access to the TokenProcessor issuer struct field.
		Describe("NewVerifierConfig", func() {
			When("empty clientID is provided", func() {
				It("should return an error", func() {
					tokenProcessor.issuer = mockIssuer
					tokenProcessor.issuer.ClientID = ""

					verifierConfig, err := tokenProcessor.NewVerifierConfig()

					Expect(err).To(HaveOccurred(), "Expected an error to occur when clientID is empty, but no error occurred")
					Expect(verifierConfig).To(Equal(VerifierConfig{}), "Expected verifierConfig to be an empty VerifierConfig struct, but got: %v", verifierConfig)
				})
			})
		})

		Describe("setIssuer", func() {
			It("should set the issuer successfully", func() {
				err := tokenProcessor.setIssuer()
				Expect(err).NotTo(HaveOccurred(), "Expected no error, but got: %v", err)
				Expect(tokenProcessor.issuer).To(Equal(mockIssuer), "Expected issuer to be set to mockIssuer, but got: %v", tokenProcessor.issuer)
			})

			It("should return an error if issuer is not trusted", func() {
				tokenProcessor.trustedIssuers = map[string]Issuer{
					"https://untrusted.issuer": mockIssuer,
				}
				err := tokenProcessor.setIssuer()
				Expect(err).To(HaveOccurred(), "Expected an error, but got none")
				Expect(err).To(MatchError(fmt.Sprintf("issuer %s is not trusted", mockIssuer.IssuerURL), "Expected error message to match"))
			})

			It("should return an error if issuer is  not valid", func() {
				tokenProcessor.trustedIssuers[mockIssuer.IssuerURL] = Issuer{
					Name:      "mock",
					IssuerURL: "https://mock.issuer",
					JWKSURL:   "https://mock.issuer/jwks",
					ClientID:  "",
				}

				err := tokenProcessor.setIssuer()
				Expect(err).To(HaveOccurred(), "Expected an error, but got none")
				Expect(err).To(MatchError("trusted issuer clientID is empty"), "Expected error message to match")
			})
		})
	})

	Describe("Issuer", func() {
		var (
			issuer Issuer
		)

		Describe("validateIssuer", func() {
			When("the issuer name is empty", func() {
				It("should return an error", func() {
					issuer = Issuer{
						IssuerURL: "https://valid.url",
						JWKSURL:   "https://valid.url/jwks",
						ClientID:  "valid-client-id",
					}
					err := issuer.validateIssuer(logger)
					Expect(err).To(HaveOccurred(), "Expected an error when issuer name is empty, but got none")
					Expect(err).To(MatchError("issuer name is empty"), "Expected error message to match")
				})
			})

			When("the issuer URL is empty", func() {
				It("should return an error", func() {
					issuer = Issuer{
						Name:     "valid-name",
						JWKSURL:  "https://valid.url/jwks",
						ClientID: "valid-client-id",
					}
					err := issuer.validateIssuer(logger)
					Expect(err).To(HaveOccurred(), "Expected an error when issuer URL is empty, but got none")
					Expect(err).To(MatchError("issuer URL is empty"), "Expected error message to match")
				})
			})

			When("the issuer URL is not valid", func() {
				It("should return an error", func() {
					issuer = Issuer{
						Name:      "valid-name",
						IssuerURL: "invalid-url",
						JWKSURL:   "https://valid.url/jwks",
						ClientID:  "valid-client-id",
					}
					err := issuer.validateIssuer(logger)
					Expect(err).To(HaveOccurred(), "Expected an error when issuer URL is not valid, but got none")
				})
			})

			When("the issuer URL is not using https", func() {
				It("should return an error", func() {
					issuer = Issuer{
						Name:      "valid-name",
						IssuerURL: "http://valid.url",
						JWKSURL:   "https://valid.url/jwks",
						ClientID:  "valid-client-id",
					}
					err := issuer.validateIssuer(logger)
					Expect(err).To(HaveOccurred(), "Expected an error when issuer URL is not using https, but got none")
					Expect(err).To(MatchError("issuer URL is not using https"), "Expected error message to match")
				})
			})

			When("the issuer JWKS URL is empty", func() {
				It("should return an error", func() {
					issuer = Issuer{
						Name:      "valid-name",
						IssuerURL: "https://valid.url",
						ClientID:  "valid-client-id",
					}
					err := issuer.validateIssuer(logger)
					Expect(err).To(HaveOccurred(), "Expected an error when issuer JWKS URL is empty, but got none")
					Expect(err).To(MatchError("issuer JWKS URL is empty"), "Expected error message to match")
				})
			})

			When("the issuer JWKS URL is not valid", func() {
				It("should return an error", func() {
					issuer = Issuer{
						Name:      "valid-name",
						IssuerURL: "https://valid.url",
						JWKSURL:   "invalid-url",
						ClientID:  "valid-client-id",
					}
					err := issuer.validateIssuer(logger)
					Expect(err).To(HaveOccurred(), "Expected an error when issuer JWKS URL is not valid, but got none")
				})
			})

			When("the issuer JWKS URL is not using https", func() {
				It("should return an error", func() {
					issuer = Issuer{
						Name:      "valid-name",
						IssuerURL: "https://valid.url",
						JWKSURL:   "http://valid.url/jwks",
						ClientID:  "valid-client-id",
					}
					err := issuer.validateIssuer(logger)
					Expect(err).To(HaveOccurred(), "Expected an error when issuer JWKS URL is not using https, but got none")
					Expect(err).To(MatchError("issuer JWKS URL is not using https"), "Expected error message to match")
				})
			})

			When("the issuer clientID is empty", func() {
				It("should return an error", func() {
					issuer = Issuer{
						Name:      "valid-name",
						IssuerURL: "https://valid.url",
						JWKSURL:   "https://valid.url/jwks",
					}
					err := issuer.validateIssuer(logger)
					Expect(err).To(HaveOccurred(), "Expected an error when issuer clientID is empty, but got none")
					Expect(err).To(MatchError("issuer clientID is empty"), "Expected error message to match")
				})
			})

			When("all issuer fields are valid", func() {
				It("should not return an error", func() {
					issuer = Issuer{
						Name:      "valid-name",
						IssuerURL: "https://valid.url",
						JWKSURL:   "https://valid.url/jwks",
						ClientID:  "valid-client-id",
					}
					err := issuer.validateIssuer(logger)
					Expect(err).NotTo(HaveOccurred(), "Expected no error when all issuer fields are valid, but got: %v", err)
				})
			})
		})
	})
})
