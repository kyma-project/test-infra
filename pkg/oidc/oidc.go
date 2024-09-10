// Package oidc provides the OIDC token verification and extraction of claims.
// It uses the OIDC discovery to get the public keys for token verification.
// It supports the GitHub OIDC issuer and GitHub specific claims.
// Its main purpose is to work as part of the Azure DevOps pipeline to verify the calls from GitHub workflows.
package oidc

import (
	"errors"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/kyma-project/test-infra/pkg/logging"
	"golang.org/x/net/context"
)

var (
	// SupportedSigningAlgorithms is a list of supported oidc token signing algorithms.
	SupportedSigningAlgorithms = []string{"RS256"}
	// GithubOIDCIssuer is the known GitHub OIDC issuer.
	GithubOIDCIssuer = Issuer{
		Name:                   "github",
		IssuerURL:              "https://token.actions.githubusercontent.com",
		JWKSURL:                "https://token.actions.githubusercontent.com/.well-known/jwks",
		ExpectedJobWorkflowRef: "kyma-project/test-infra/.github/workflows/image-builder.yml@refs/heads/main",
		GithubURL:              "https://github.com",
	}
	GithubToolsSAPOIDCIssuer = Issuer{
		Name:                   "github-tools-sap",
		IssuerURL:              "https://github.tools.sap/_services/token",
		JWKSURL:                "https://github.tools.sap/_services/token/.well-known/jwks",
		ExpectedJobWorkflowRef: "kyma/oci-image-builder/.github/workflows/image-builder.yml@refs/heads/main",
		GithubURL:              "https://github.tools.sap",
	}
	TrustedOIDCIssuers = map[string]Issuer{GithubOIDCIssuer.IssuerURL: GithubOIDCIssuer, GithubToolsSAPOIDCIssuer.IssuerURL: GithubToolsSAPOIDCIssuer}
)

// TODO(dekiel) interfaces need to be clenup up to remove redundancy.

type TokenVerifierInterface interface {
	Verify(context.Context, string) (Token, error)
}

type Verifier interface {
	Verify(context.Context, string) (*oidc.IDToken, error)
}

type TokenInterface interface {
	Claims(interface{}) error
}

type ClaimsReader interface {
	Claims(interface{}) error
}

type ProviderInterface interface {
	NewVerifier(VerifierConfig) TokenVerifier
}

type VerifierProvider interface {
	Verifier(*oidc.Config) *oidc.IDTokenVerifier
}

type ClaimsInterface interface {
	// Validate(jwt.Expected) error
	ValidateExpectations(Issuer) error
}

type LoggerInterface interface {
	logging.StructuredLoggerInterface
	logging.WithLoggerInterface
}

// Issuer is the OIDC issuer.
// Name is the human-readable name of the issuer.
// IssuerURL is the OIDC discovery issuer endpoint.
// JWKSURL is the OIDC issuer public keys endpoint.
type Issuer struct {
	Name                   string `json:"name" yaml:"name"`
	IssuerURL              string `json:"issuer_url" yaml:"issuer_url"`
	JWKSURL                string `json:"jwks_url" yaml:"jwks_url"`
	ExpectedJobWorkflowRef string `json:"expected_job_workflow_ref" yaml:"expected_job_workflow_ref"`
	GithubURL              string `json:"github_url" yaml:"github_url"`
}

func (i Issuer) GetGithubURL() string {
	return i.GithubURL
}

// VerifierConfig is the configuration for a verifier.
// It abstracts the oidc.Config.
type VerifierConfig struct {
	oidc.Config
}

// TokenProcessor is responsible for processing the token.
type TokenProcessor struct {
	rawToken       string
	trustedIssuers map[string]Issuer
	issuer         Issuer
	verifierConfig VerifierConfig
	logger         LoggerInterface
}

func (tokenProcessor *TokenProcessor) GetIssuer() Issuer {
	return tokenProcessor.issuer
}

// TokenProcessorOption is a function that modifies the TokenProcessor.
type TokenProcessorOption func(*TokenProcessor) error

// Claims represent the OIDC token claims.
// It extends the jwt.Claims which provides the standard claims.
// It adds the GitHub specific claims.
type Claims struct {
	jwt.Claims
	GithubClaims
	LoggerInterface
}

// GithubClaims provide the GitHub specific claims.
type GithubClaims struct {
	JobWorkflowRef  string `json:"job_workflow_ref,omitempty" yaml:"job_workflow_ref,omitempty"`
	JobWorkflowSHA  string `json:"job_workflow_sha,omitempty" yaml:"job_workflow_sha,omitempty"`
	Actor           string `json:"actor,omitempty" yaml:"actor,omitempty"`
	EventName       string `json:"event_name,omitempty" yaml:"event_name,omitempty"`
	Repository      string `json:"repository,omitempty" yaml:"repository,omitempty"`
	RepositoryOwner string `json:"repository_owner,omitempty" yaml:"repository_owner,omitempty"`
	RunID           string `json:"run_id,omitempty" yaml:"run_id,omitempty"`
}

// VerifierConfigOption is a function that modifies the VerifierConfig.
type VerifierConfigOption func(*VerifierConfig) error

// Provider is the OIDC provider.
// It abstracts the provider implementation.
// VerifierProvider provides the OIDC token verifier.
type Provider struct {
	VerifierProvider VerifierProvider
	logger           LoggerInterface
}

// Token is the OIDC token.
// It abstracts the IDToken implementation.
type Token struct {
	Token ClaimsReader
}

// TokenVerifier is the OIDC token verifier.
// It abstracts the Verifier implementation.
type TokenVerifier struct {
	Verifier Verifier
	Logger   LoggerInterface
}

// NewVerifierConfig creates a new VerifierConfig.
// It verifies the clientID is not empty.
func NewVerifierConfig(logger LoggerInterface, clientID string, options ...VerifierConfigOption) (VerifierConfig, error) {
	if clientID == "" {
		return VerifierConfig{}, fmt.Errorf("clientID is empty")
	}

	verifierConfig := VerifierConfig{}
	verifierConfig.ClientID = clientID
	verifierConfig.SkipClientIDCheck = false
	verifierConfig.SkipExpiryCheck = false
	verifierConfig.SkipIssuerCheck = false
	verifierConfig.InsecureSkipSignatureCheck = false
	verifierConfig.SupportedSigningAlgs = SupportedSigningAlgorithms

	logger.Debugw("Created Verifier config with default values",
		"clientID", clientID,
		"SkipClientIDCheck", verifierConfig.SkipClientIDCheck,
		"SkipExpiryCheck", verifierConfig.SkipExpiryCheck,
		"SkipIssuerCheck", verifierConfig.SkipIssuerCheck,
		"InsecureSkipSignatureCheck", verifierConfig.InsecureSkipSignatureCheck,
		"SupportedSigningAlgs", verifierConfig.SupportedSigningAlgs,
	)
	logger.Debugw("Applying VerifierConfigOptions")
	for _, option := range options {
		err := option(&verifierConfig)
		if err != nil {
			return VerifierConfig{}, fmt.Errorf("failed to apply VerifierConfigOption: %w", err)
		}
	}
	logger.Debugw("Applied all VerifierConfigOptions",
		"clientID", clientID,
		"SkipClientIDCheck", verifierConfig.SkipClientIDCheck,
		"SkipExpiryCheck", verifierConfig.SkipExpiryCheck,
		"SkipIssuerCheck", verifierConfig.SkipIssuerCheck,
		"InsecureSkipSignatureCheck", verifierConfig.InsecureSkipSignatureCheck,
		"SupportedSigningAlgs", verifierConfig.SupportedSigningAlgs,
	)
	return verifierConfig, nil
}

// Verify verifies the raw OIDC token.
// It returns a Token struct which contains the verified token if successful.
func (tokenVerifier *TokenVerifier) Verify(ctx context.Context, rawToken string) (Token, error) {
	logger := tokenVerifier.Logger
	logger.Debugw("Verifying token")
	logger.Debugw("Got raw token value", "rawToken", rawToken)
	idToken, err := tokenVerifier.Verifier.Verify(ctx, rawToken)
	if err != nil {
		token := Token{}
		return token, fmt.Errorf("failed to verify token: %w", err)
	}
	logger.Debugw("Token verified successfully")
	token := Token{
		Token: idToken,
	}
	return token, nil
}

// Claims gets the claims from the token and unmarshal them into the provided claims struct.
func (token *Token) Claims(claims interface{}) error {
	return token.Token.Claims(claims)
}

// NewClaims creates a new Claims struct with the provided logger.
func NewClaims(logger LoggerInterface) Claims {
	return Claims{
		LoggerInterface: logger,
	}
}

// ValidateExpectations validates the claims against the trusted issuer expected values.
// It checks audience, issuer, and job_workflow_ref claims.
func (claims *Claims) ValidateExpectations(issuer Issuer) error {
	logger := claims.LoggerInterface
	logger.Debugw("Validating job_workflow_ref claim against expected value", "job_workflow_ref", claims.JobWorkflowRef, "expected", issuer.ExpectedJobWorkflowRef)
	if claims.JobWorkflowRef != issuer.ExpectedJobWorkflowRef {
		return fmt.Errorf("job_workflow_ref claim expected value validation failed, expected: %s, provided: %s", claims.JobWorkflowRef, issuer.ExpectedJobWorkflowRef)
	}
	logger.Debugw("Claims validated successfully")
	return nil
}

// NewProviderFromDiscovery creates a new Provider for issuer using OIDC discovery.
// It returns a Provider struct containing the new Provider if successful.
func NewProviderFromDiscovery(ctx context.Context, logger LoggerInterface, issuer string) (Provider, error) {
	logger.Debugw("Creating Provider using discovery",
		"issuer", issuer,
	)
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return Provider{}, fmt.Errorf("failed to create Provider using oidc discovery: %w", err)
	}
	newProvider := Provider{
		VerifierProvider: provider,
		logger:           logger,
	}
	logger.Debugw("Created Provider using oidc discovery")
	return newProvider, nil
}

// NewVerifier creates a new TokenVerifier for provider.
// It returns a TokenVerifier struct containing the new Verifier if successful.
func (provider *Provider) NewVerifier(logger LoggerInterface, verifierConfig VerifierConfig) TokenVerifier {
	logger.Debugw("Creating new verifier with config", "config", fmt.Sprintf("%+v", verifierConfig))
	verifier := provider.VerifierProvider.Verifier(&verifierConfig.Config)
	tokenVerifier := TokenVerifier{
		Verifier: verifier,
		Logger:   logger,
	}
	logger.Debugw("Created new verifier")
	return tokenVerifier
}

// NewTokenProcessor creates a new TokenProcessor for trusted issuers.
// It reads the token, gets the issuer from the token, and checks if the issuer is trusted.
// It verifies the VerifierConfig has a clientID.
func NewTokenProcessor(
	logger LoggerInterface,
	trustedIssuers map[string]Issuer,
	rawToken string,
	config VerifierConfig,
	options ...TokenProcessorOption,
) (TokenProcessor, error) {
	logger.Debugw("Creating token processor")
	tokenProcessor := TokenProcessor{}

	tokenProcessor.logger = logger

	tokenProcessor.rawToken = rawToken
	logger.Debugw("Added raw token to token processor", "rawToken", rawToken)

	tokenProcessor.verifierConfig = config
	logger.Debugw("Added Verifier config to token processor",
		"clientID", config.ClientID,
		"SkipClientIDCheck", config.SkipClientIDCheck,
		"SkipExpiryCheck", config.SkipExpiryCheck,
		"SkipIssuerCheck", config.SkipIssuerCheck,
		"InsecureSkipSignatureCheck", config.InsecureSkipSignatureCheck,
		"SupportedSigningAlgs", config.SupportedSigningAlgs,
	)
	if tokenProcessor.verifierConfig.ClientID == "" {
		return TokenProcessor{}, errors.New("verifierConfig clientID is empty")
	}

	tokenProcessor.trustedIssuers = trustedIssuers
	issuer, err := tokenProcessor.tokenIssuer(SupportedSigningAlgorithms)
	if err != nil {
		return TokenProcessor{}, fmt.Errorf("failed to get issuer from token: %w", err)
	}
	logger.Debugw("Got issuer from token", "issuer", issuer)
	trustedIssuer, err := tokenProcessor.isTrustedIssuer(issuer, tokenProcessor.trustedIssuers)
	if err != nil {
		return TokenProcessor{}, err
	}
	tokenProcessor.issuer = trustedIssuer
	logger.Debugw("Added trusted issuer to TokenProcessor", "issuer", tokenProcessor.issuer)

	logger.Debugw("Applying TokenProcessorOptions")
	for _, option := range options {
		err := option(&tokenProcessor)
		if err != nil {
			return TokenProcessor{}, fmt.Errorf("failed to apply TokenProcessorOption: %w", err)
		}
	}
	logger.Debugw("Applied all TokenProcessorOptions")

	logger.Debugw("Created token processor", "issuer", tokenProcessor.issuer)
	return tokenProcessor, nil
}

// tokenIssuer gets the issuer from the token.
// It doesn't verify the token, just parses its claims.
// It's used to create a new TokenProcessor.
// TODO: signAlgorithm should be a TokenProcessor field.
// TODO: should we abstract usage of jwt and go-jose libraries?
func (tokenProcessor *TokenProcessor) tokenIssuer(signAlgorithm []string) (string, error) {
	logger := tokenProcessor.logger
	logger.Debugw("Getting issuer from token")
	claims := NewClaims(logger)
	var signAlgs []jose.SignatureAlgorithm
	for _, alg := range signAlgorithm {
		signAlgs = append(signAlgs, jose.SignatureAlgorithm(alg))
		logger.Debugw("Added sign algorithm to token processor", "signAlgorithm", alg)
	}
	logger.Debugw("Added sign algorithms to token processor")
	// TODO(dekiel) research if we can use jwt.DecodeSegment instead of jwt.ParseSigned
	//  to avoid parsing the token twice.
	parsedJWT, err := jwt.ParseSigned(tokenProcessor.rawToken, signAlgs)
	if err != nil {
		return "", fmt.Errorf("failed to parse oidc token: %w", err)
	}
	logger.Debugw("Parsed oidc token", "parsedJWT", fmt.Sprintf("%+v", parsedJWT))
	err = parsedJWT.UnsafeClaimsWithoutVerification(&claims)
	if err != nil {
		return "", fmt.Errorf("failed to get claims from unverified token: %w", err)
	}
	logger.Debugw("Got claims from unverified token", "claims", fmt.Sprintf("%+v", claims))
	return claims.Issuer, nil
}

// isTrustedIssuer checks if the issuer is trusted.
// It's used to create a new TokenProcessor.
func (tokenProcessor *TokenProcessor) isTrustedIssuer(issuer string, trustedIssuers map[string]Issuer) (Issuer, error) {
	logger := tokenProcessor.logger
	logger.Debugw("Checking if issuer is trusted", "issuer", issuer)
	logger.Debugw("Trusted issuers", "trustedIssuers", trustedIssuers)
	if trustedIssuer, exists := trustedIssuers[issuer]; exists {
		logger.Debugw("Issuer is trusted", "issuer", issuer)
		return trustedIssuer, nil
	}
	logger.Debugw("Issuer is not trusted", "issuer", issuer)
	return Issuer{}, fmt.Errorf("issuer %s is not trusted", issuer)
}

// TODO(dekiel): This should return Issuer struct or has a different name.
// Issuer returns the issuer of the token.
func (tokenProcessor *TokenProcessor) Issuer() string {
	return tokenProcessor.issuer.IssuerURL
}

// VerifyAndExtractClaims verify and parse the token to get the token claims.
// It uses the provided verifier to verify the token signature and expiration time.
// It verifies if the token claims have expected values.
// It unmarshal the claims into the provided claims struct.
func (tokenProcessor *TokenProcessor) VerifyAndExtractClaims(ctx context.Context, verifier TokenVerifierInterface, claims ClaimsInterface) error {
	logger := tokenProcessor.logger
	token, err := verifier.Verify(ctx, tokenProcessor.rawToken)
	if err != nil {
		return fmt.Errorf("failed to verify token: %w", err)
	}
	logger.Debugw("Getting claims from token")
	err = token.Claims(claims)
	if err != nil {
		return fmt.Errorf("failed to get claims from token: %w", err)
	}
	logger.Debugw("Got claims from token", "claims", fmt.Sprintf("%+v", claims))
	err = claims.ValidateExpectations(tokenProcessor.issuer)
	if err != nil {
		return fmt.Errorf("failed to validate claims: %w", err)
	}
	return nil
}
