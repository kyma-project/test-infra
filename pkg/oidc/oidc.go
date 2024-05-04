package oidc

import (
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/kyma-project/test-infra/pkg/logging"
	"golang.org/x/net/context"
)

var (
	SupportedSigningAlgorithms = []string{"RS256"}
	GithubOIDCIssuer           = Issuer{
		Name:      "github",
		IssuerURL: "https://token.actions.githubusercontent.com",
		JWKSURL:   "https://token.actions.githubusercontent.com/.well-known/jwks",
	}
	TrustedOIDCIssuers = map[string]Issuer{GithubOIDCIssuer.IssuerURL: GithubOIDCIssuer}
)

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

type NewVerifier interface {
	Verifier(*oidc.Config) *oidc.IDTokenVerifier
}

type ClaimsInterface interface {
	Validate(e jwt.Expected) error
}

type LoggerInterface interface {
	logging.StructuredLoggerInterface
	logging.WithLoggerInterface
}

type Issuer struct {
	Name      string `json:"name" yaml:"name"`
	IssuerURL string `json:"issuer_url" yaml:"issuer_url"`
	JWKSURL   string `json:"jwks_url" yaml:"jwks_url"`
}

type VerifierConfig struct {
	oidc.Config
}

type TokenProcessor struct {
	rawToken       string
	trustedIssuers map[string]Issuer
	issuer         string
	verifierConfig VerifierConfig
	logger         LoggerInterface
}

type TokenProcessorOption func(*TokenProcessor) error

type Claims struct {
	jwt.Claims
	GithubClaims
}

type GithubClaims struct {
	JobWorkflowRef  string `json:"job_workflow_ref,omitempty" yaml:"job_workflow_ref,omitempty"`
	JobWorkflowSHA  string `json:"job_workflow_sha,omitempty" yaml:"job_workflow_sha,omitempty"`
	Actor           string `json:"actor,omitempty" yaml:"actor,omitempty"`
	EventName       string `json:"event_name,omitempty" yaml:"event_name,omitempty"`
	Repository      string `json:"repository,omitempty" yaml:"repository,omitempty"`
	RepositoryOwner string `json:"repository_owner,omitempty" yaml:"repository_owner,omitempty"`
	RunID           string `json:"run_id,omitempty" yaml:"run_id,omitempty"`
}

type VerifierConfigOption func(*VerifierConfig) error

type Provider struct {
	Provider NewVerifier
	logger   LoggerInterface
}

type Token struct {
	Token ClaimsReader
}

type TokenVerifier struct {
	Verifier Verifier
	Logger   LoggerInterface
}

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

func (tokenVerifier *TokenVerifier) Verify(ctx context.Context, rawToken string) (Token, error) {
	logger := tokenVerifier.Logger
	logger.Infow("Verifying token")
	logger.Debugw("Got raw token value", "rawToken", rawToken)
	idToken, err := tokenVerifier.Verifier.Verify(ctx, rawToken)
	if err != nil {
		token := Token{}
		return token, fmt.Errorf("failed to verify token: %w", err)
	}
	logger.Infow("Token verified successfully")
	token := Token{
		Token: idToken,
	}
	return token, nil
}

func (token *Token) Claims(claims interface{}) error {
	return token.Token.Claims(claims)
}

func NewProviderFromDiscovery(ctx context.Context, logger LoggerInterface, issuer string) (Provider, error) {
	logger.Debugw("Creating Provider using discovery",
		"issuer", issuer,
	)
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return Provider{}, fmt.Errorf("failed to create Provider using oidc discovery: %w", err)
	}
	newProvider := Provider{
		Provider: provider,
		logger:   logger,
	}
	logger.Debugw("Created Provider using oidc discovery")
	return newProvider, nil
}

func (provider *Provider) NewVerifier(verifierConfig VerifierConfig) TokenVerifier {
	logger := provider.logger
	verifier := provider.Provider.Verifier(&verifierConfig.Config)
	return TokenVerifier{
		Verifier: verifier,
		Logger:   logger,
	}
}

func NewTokenProcessor(
	logger LoggerInterface,
	trustedIssuers map[string]Issuer,
	rawToken string,
	config VerifierConfig,
	options ...TokenProcessorOption,
) (TokenProcessor, error) {
	logger.Infow("Creating token processor")
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
	tokenProcessor.trustedIssuers = trustedIssuers
	issuer, err := tokenProcessor.tokenIssuer(SupportedSigningAlgorithms)
	if err != nil {
		return TokenProcessor{}, fmt.Errorf("failed to get issuer from token: %w", err)
	}
	logger.Debugw("Got issuer from token", "issuer", issuer)
	if !tokenProcessor.isTrustedIssuer(issuer, tokenProcessor.trustedIssuers) {
		return TokenProcessor{}, fmt.Errorf("%s issuer is not trusted", issuer)
	}
	tokenProcessor.issuer = issuer
	logger.Infow("Added trusted issuer to TokenProcessor", "issuer", tokenProcessor.issuer)
	// tokenProcessor.Provider = Provider
	// logger.Debugw("Added Provider to token processor")
	for _, option := range options {
		err := option(&tokenProcessor)
		if err != nil {
			return TokenProcessor{}, fmt.Errorf("failed to apply TokenProcessorOption: %w", err)
		}
	}
	logger.Infow("Created token processor", "issuer", tokenProcessor.issuer)
	return tokenProcessor, nil
}

func (tokenProcessor *TokenProcessor) tokenIssuer(signAlgorithm []string) (string, error) {
	logger := tokenProcessor.logger
	logger.Infow("Getting issuer from token")
	claims := jwt.Claims{}
	var signAlgs []jose.SignatureAlgorithm
	for _, alg := range signAlgorithm {
		signAlgs = append(signAlgs, jose.SignatureAlgorithm(alg))
		logger.Debugw("Added sign algorithm to token processor", "signAlgorithm", alg)
	}
	logger.Infow("Added sign algorithms to token processor")
	// TODO(dekiel) research if we can use jwt.DecodeSegment instead of jwt.ParseSigned
	//  to avoid parsing the token twice.
	parsedJWT, err := jwt.ParseSigned(tokenProcessor.rawToken, signAlgs)
	if err != nil {
		return "", fmt.Errorf("failed to parse oidc token: %w", err)
	}
	logger.Debugw("Parsed oidc token")
	err = parsedJWT.UnsafeClaimsWithoutVerification(&claims)
	if err != nil {
		return "", fmt.Errorf("failed to get claims from unverified token: %w", err)
	}
	logger.Debugw("Got claims from unverified token", "claims", fmt.Sprintf("%+v", claims))
	return claims.Issuer, nil
}

// isTrustedIssuer checks if the issuer is trusted.
func (tokenProcessor *TokenProcessor) isTrustedIssuer(issuer string, trustedIssuers map[string]Issuer) bool {
	logger := tokenProcessor.logger
	logger.Debugw("Checking if issuer is trusted", "issuer", issuer)
	logger.Debugw("Trusted issuers", "trustedIssuers", trustedIssuers)
	if _, exists := trustedIssuers[issuer]; exists {
		logger.Debugw("Issuer is trusted", "issuer", issuer)
		return true
	}
	logger.Debugw("Issuer is not trusted", "issuer", issuer)
	return false
}

// Issuer returns the issuer of the token.
func (tokenProcessor *TokenProcessor) Issuer() string {
	return tokenProcessor.issuer
}

func (tokenProcessor *TokenProcessor) Claims(ctx context.Context, verifier TokenVerifierInterface, claims ClaimsInterface) error {
	logger := tokenProcessor.logger
	idToken, err := verifier.Verify(ctx, tokenProcessor.rawToken)
	if err != nil {
		return fmt.Errorf("failed to verify token: %w", err)
	}
	logger.Debugw("Getting claims from token")
	err = idToken.Claims(claims)
	if err != nil {
		return fmt.Errorf("failed to get claims from token: %w", err)
	}
	logger.Debugw("Got claims from token", "claims", fmt.Sprintf("%+v", claims))
	return nil
}
