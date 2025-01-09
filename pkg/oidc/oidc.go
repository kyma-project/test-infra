// Package oidc provides the OIDC token verification and extraction of claims.
// It uses the OIDC discovery to get the public keys for token verification.
// It supports the GitHub OIDC issuer and GitHub specific claims.
// Its main purpose is to work as part of the Azure DevOps pipeline to verify the calls from GitHub workflows.
package oidc

import (
	"errors"
	"fmt"
	"time"

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
		ClientID: "image-builder",
	}
	GithubToolsSAPOIDCIssuer = Issuer{
		Name:                   "github-tools-sap",
		IssuerURL:              "https://github.tools.sap/_services/token",
		JWKSURL:                "https://github.tools.sap/_services/token/.well-known/jwks",
		ExpectedJobWorkflowRef: "kyma/oci-image-builder/.github/workflows/image-builder.yml@refs/heads/main",
		GithubURL:              "https://github.tools.sap",
		ClientID: "image-builder",
	}
	TrustedOIDCIssuers = map[string]Issuer{
		GithubOIDCIssuer.IssuerURL:         GithubOIDCIssuer,
		GithubToolsSAPOIDCIssuer.IssuerURL: GithubToolsSAPOIDCIssuer,
	}
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
	validateExpectations(Issuer) error
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
	// The clientID is used to verify the audience claim in the token.
	ClientID string `json:"client_id" yaml:"client_id"`
}

func (i Issuer) GetGithubURL() string {
	return i.GithubURL
}

// VerifierConfig is the configuration for a verifier.
// It abstracts the oidc.Config.
type VerifierConfig struct {
	oidc.Config
}

// String returns the string representation of the VerifierConfig.
// It's used for logging purposes.
func (config *VerifierConfig) String() string {
	return fmt.Sprintf("ClientID: %s, SkipClientIDCheck: %t, SkipExpiryCheck: %t, SkipIssuerCheck: %t, InsecureSkipSignatureCheck: %t, SupportedSigningAlgs: %v, Now: %T",
		config.ClientID, config.SkipClientIDCheck, config.SkipExpiryCheck, config.SkipIssuerCheck, config.InsecureSkipSignatureCheck, config.SupportedSigningAlgs, config.Now)
}

// TokenProcessor is responsible for processing the token.
type TokenProcessor struct {
	rawToken       string
	trustedIssuers map[string]Issuer
	issuer         Issuer
	logger LoggerInterface
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

// SkipExpiryCheck set verifier config to skip the token expiration check.
// It's used to allow longer token expiration time.
// The Verifier must run its own expiration check for extended expiration time.
func SkipExpiryCheck() VerifierConfigOption {
	return func(config *VerifierConfig) error {
		config.SkipExpiryCheck = true
		return nil
	}
}

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
	Token    ClaimsReader
	IssuedAt time.Time
}

// TokenVerifier is the OIDC token verifier.
// It abstracts the Verifier implementation.
type TokenVerifier struct {
	Config                VerifierConfig
	ExpirationTimeMinutes int
	Verifier              Verifier
	Logger                LoggerInterface
}

// TokenVerifierOption is a function that modifies the TokenVerifier.
type TokenVerifierOption func(verifier *TokenVerifier) error

// WithExtendedExpiration sets the custom expiration time used by the TokenVerifier.
func WithExtendedExpiration(expirationTimeMinutes int) TokenVerifierOption {
	return func(verifier *TokenVerifier) error {
		verifier.ExpirationTimeMinutes = expirationTimeMinutes
		return nil
	}
}

// maskToken masks the token value. It's used for debug logging purposes.
func maskToken(token string) string {
	if len(token) < 15 {
		return "********"
	}
	return token[:2] + "********" + token[len(token)-2:]
}

// Verify verifies the raw OIDC token.
// It returns a Token struct which contains the verified token if successful.
// Verify allow checking extended expiration time for the token.
// It runs the token expiration check if the upstream verifier is configured to skip it.
func (tokenVerifier *TokenVerifier) Verify(ctx context.Context, rawToken string) (*Token, error) {
	logger := tokenVerifier.Logger
	logger.Debugw("verifying token", "rawToken", maskToken(rawToken))
	idToken, err := tokenVerifier.Verifier.Verify(ctx, rawToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}
	logger.Debugw("upstream verifier checks finished")
	token := Token{
		Token:    idToken,
		IssuedAt: idToken.IssuedAt,
	}
	logger.Debugw("checking if upstream verifier is configured to skip token expiration check", "SkipExpiryCheck", tokenVerifier.Config.SkipExpiryCheck)
	if tokenVerifier.Config.SkipExpiryCheck {
		logger.Debugw("upstream verifier configured to skip token expiration check, running our own check", "expirationTimeMinutes", tokenVerifier.ExpirationTimeMinutes, "tokenIssuedAt", token.IssuedAt)
		err = token.IsTokenExpired(logger, tokenVerifier.ExpirationTimeMinutes)
		logger.Debugw("finished token expiration check")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}
	logger.Debugw("token verified successfully")
	return &token, nil
}

// IsTokenExpired checks the OIDC token expiration timestamp against the provided expiration time.
// It allows accepting tokens after the token original expiration time elapsed.
// The other aspects of the token must be verified separately with the expiration check disabled.
func (token *Token) IsTokenExpired(logger LoggerInterface, expirationTimeMinutes int) error {
	logger.Debugw("verifying token expiration time", "tokenIssuedAt", token.IssuedAt, "expirationTimeMinutes", expirationTimeMinutes)
	now := time.Now()
	if token.IssuedAt.After(now) {
		return fmt.Errorf("token issued in the future, tokenIssuedAt: %v, now: %v", token.IssuedAt, now)
	}
	logger.Debugw("token issued in the past")
	expirationTime := token.IssuedAt.Add(time.Minute * time.Duration(expirationTimeMinutes))
	logger.Debugw("computed expiration time", "expirationTime", expirationTime)
	if expirationTime.After(now) || expirationTime.Equal(now) {
		logger.Debugw("token not expired")
		return nil
	}
	return fmt.Errorf("token expired, tokenIssuedAt: %v, expired at %v", token.IssuedAt, expirationTime)
}

// Claims gets the claims from the token and unmarshal them into the provided claims struct.
// TODO: Should we have a tests for this method?
//
//	We can test for returned error.
//	We can test the claims struct is populated with the expected values.
func (token *Token) Claims(claims interface{}) error {
	return token.Token.Claims(claims)
}

// NewClaims creates a new Claims struct with the provided logger.
func NewClaims(logger LoggerInterface) Claims {
	return Claims{
		LoggerInterface: logger,
	}
}

// validateExpectations validates the claims against the trusted issuer expected values.
// It checks audience, issuer, and job_workflow_ref claims.
func (claims *Claims) validateExpectations(issuer Issuer) error {
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
func (provider *Provider) NewVerifier(logger LoggerInterface, verifierConfig VerifierConfig, options ...TokenVerifierOption) (TokenVerifier, error) {
	logger.Debugw("Creating new verifier with config", "config", fmt.Sprintf("%+v", verifierConfig))
	verifier := provider.VerifierProvider.Verifier(&verifierConfig.Config)
	tokenVerifier := TokenVerifier{
		Config:   verifierConfig,
		Verifier: verifier,
		Logger:   logger,
	}
	if len(options) > 0 {
		logger.Debugw("Applying TokenVerifierOptions")
		for _, option := range options {
			err := option(&tokenVerifier)
			if err != nil {
				return TokenVerifier{}, fmt.Errorf("failed to apply TokenVerifierOption: %w", err)
			}
		}
		logger.Debugw("Applied all TokenVerifierOptions")
	}
	logger.Debugw("Created new verifier")
	return tokenVerifier, nil
}

// NewTokenProcessor creates a new TokenProcessor for a trusted issuer.
// It reads the issuer from the raw token and checks if the issuer is trusted.
// It verifies the trusted issuer has a non-empty clientID.
func NewTokenProcessor(
	logger LoggerInterface,
	trustedIssuers map[string]Issuer,
	rawToken string,
	options ...TokenProcessorOption,
) (TokenProcessor, error) {
	logger.Debugw("Creating token processor")
	tokenProcessor := TokenProcessor{}

	tokenProcessor.logger = logger

	tokenProcessor.rawToken = rawToken
	logger.Debugw("Added raw token to token processor", "rawToken", maskToken(rawToken))

	tokenProcessor.trustedIssuers = trustedIssuers
	logger.Debugw("Added trusted issuers to token processor", "trustedIssuers", trustedIssuers)

	issuer, err := tokenProcessor.tokenIssuer(SupportedSigningAlgorithms)
	if err != nil {
		return TokenProcessor{}, fmt.Errorf("failed to get issuer from token: %w", err)
	}

	logger.Debugw("Got issuer from token", "issuer", issuer)

	trustedIssuer, err := tokenProcessor.isTrustedIssuer(issuer, tokenProcessor.trustedIssuers)
	if err != nil {
		return TokenProcessor{}, err
	}

	logger.Debugw("Matched issuer with trusted issuer", "trustedIssuer", trustedIssuer)

	if trustedIssuer.ClientID == "" {
		return TokenProcessor{}, errors.New("trusted issuer clientID is empty")
	}

	logger.Debugw("Verified trusted issuer clientID", "clientID", trustedIssuer.ClientID)

	tokenProcessor.issuer = trustedIssuer

	logger.Debugw("Added trusted issuer to TokenProcessor", "issuer", tokenProcessor.issuer)

	if len(options) > 0 {
		logger.Debugw("Applying TokenProcessorOptions")
		for _, option := range options {
			err := option(&tokenProcessor)
			if err != nil {
				return TokenProcessor{}, fmt.Errorf("failed to apply TokenProcessorOption: %w", err)
			}
		}
		logger.Debugw("Applied all TokenProcessorOptions")
	}

	logger.Debugw("Created token processor", "issuer", tokenProcessor.issuer)

	return tokenProcessor, nil
}

// NewVerifierConfig creates a new VerifierConfig for trusted issuer.
// It verifies if the clientID in the tokenProcessor is not empty.
func (tokenProcessor *TokenProcessor) NewVerifierConfig(options ...VerifierConfigOption) (VerifierConfig, error) {
	logger := tokenProcessor.logger

	if tokenProcessor.issuer.ClientID == "" {
		return VerifierConfig{}, fmt.Errorf("clientID is empty")
	}

	logger.Debugw("TokenProcessor clientID is not empty", "clientID", tokenProcessor.issuer.ClientID)

	verifierConfig := VerifierConfig{}
	verifierConfig.ClientID = tokenProcessor.issuer.ClientID
	verifierConfig.SkipClientIDCheck = false
	verifierConfig.SkipExpiryCheck = false
	verifierConfig.SkipIssuerCheck = false
	verifierConfig.InsecureSkipSignatureCheck = false
	verifierConfig.SupportedSigningAlgs = SupportedSigningAlgorithms

	logger.Debugw("Created Verifier config with default values",
		"clientID", verifierConfig.ClientID,
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
		"clientID", verifierConfig.ClientID,
		"SkipClientIDCheck", verifierConfig.SkipClientIDCheck,
		"SkipExpiryCheck", verifierConfig.SkipExpiryCheck,
		"SkipIssuerCheck", verifierConfig.SkipIssuerCheck,
		"InsecureSkipSignatureCheck", verifierConfig.InsecureSkipSignatureCheck,
		"SupportedSigningAlgs", verifierConfig.SupportedSigningAlgs,
	)

	return verifierConfig, nil
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

// Issuer returns the issuer of the token.
// TODO(dekiel): This should return Issuer struct or has a different name.
func (tokenProcessor *TokenProcessor) Issuer() string {
	return tokenProcessor.issuer.IssuerURL
}

// ValidateClaims verify and parse the token to get the token claims.
// It uses the provided verifier to verify the token signature and expiration time.
// It verifies if the token claims have expected values.
// It unmarshal the claims into the provided claims struct.
func (tokenProcessor *TokenProcessor) ValidateClaims(claims ClaimsInterface, token ClaimsReader) error {
	logger := tokenProcessor.logger

	// Ensure that the token is initialized
	if token == nil {
		return fmt.Errorf("token cannot be nil")
	}

	logger.Debugw("Getting claims from token")
	err := token.Claims(claims)
	if err != nil {
		return fmt.Errorf("failed to get claims from token: %w", err)
	}
	logger.Debugw("Got claims from token", "claims", fmt.Sprintf("%+v", claims))

	err = claims.validateExpectations(tokenProcessor.issuer)
	if err != nil {
		return fmt.Errorf("expecations validation failed: %w", err)
	}
	return nil
}