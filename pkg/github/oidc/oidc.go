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
	GithubOIDCIssuer           = map[string]string{"https://token.actions.githubusercontent.com": "https://token.actions.githubusercontent.com/.well-known/jwks"}
	TrustedOIDCIssuers         = map[string]map[string]string{"github": GithubOIDCIssuer}
)

type TokenVerifierInterface interface {
	Verify(context.Context, string) (*oidc.IDToken, error)
}

type TokenInterface interface {
	Claims(interface{}) error
}

type ClaimsInterface interface {
	Validate(e jwt.Expected) error
}

type loggerInterface interface {
	logging.StructuredLoggerInterface
	logging.WithLoggerInterface
}

type TokenProcessor struct {
	rawToken       string
	token          TokenInterface
	issuer         string
	verifierConfig oidc.Config
	logger         logging.StructuredLoggerInterface
}

type TokenProcessorOption func(*TokenProcessor) error

type GithubClaims struct {
	jwt.Claims
	JobWorkflowRef  string `json:"job_workflow_ref,omitempty"`
	JobWorkflowSHA  string `json:"job_workflow_sha,omitempty"`
	Actor           string `json:"actor,omitempty"`
	EventName       string `json:"event_name,omitempty"`
	Repository      string `json:"repository,omitempty"`
	RepositoryOwner string `json:"repository_owner,omitempty"`
	RunID           string `json:"run_id,omitempty"`
}

type VerifierConfigOptions func(config *oidc.Config) error

func NewTokenProcessor(logger loggerInterface, rawToken string, config oidc.Config, options ...TokenProcessorOption) (*TokenProcessor, error) {
	logger.Infow("Creating token processor")
	tokenProcessor := &TokenProcessor{}
	tokenProcessor.logger = logger
	tokenProcessor.rawToken = rawToken
	logger.Debugw("Added raw token to token processor", "rawToken", rawToken)
	tokenProcessor.verifierConfig = config
	logger.Debugw("Added verifier config to token processor",
		"clientID", config.ClientID,
		"SkipClientIDCheck", config.SkipClientIDCheck,
		"SkipExpiryCheck", config.SkipExpiryCheck,
		"SkipIssuerCheck", config.SkipIssuerCheck,
		"InsecureSkipSignatureCheck", config.InsecureSkipSignatureCheck,
		"SupportedSigningAlgs", config.SupportedSigningAlgs,
	)
	err := tokenProcessor.tokenIssuer(SupportedSigningAlgorithms)
	if err != nil {
		return nil, fmt.Errorf("failed to get issuer from token: %w", err)
	}
	logger.Debugw("Added issuer to token processor", "issuer", tokenProcessor.issuer)
	for _, option := range options {
		err := option(tokenProcessor)
		if err != nil {
			return nil, fmt.Errorf("failed to apply TokenProcessorOption: %w", err)
		}
	}
	logger.Infow("Created token processor", "issuer", tokenProcessor.issuer)
	return tokenProcessor, nil
}

// TODO(dekiel) verify if issuer is trusted, fail if not
func (tokenProcessor *TokenProcessor) tokenIssuer(signAlgorithm []string) error {
	tokenProcessor.logger.Infow("Getting issuer from token")
	claims := GithubClaims{}
	var signAlgs []jose.SignatureAlgorithm
	for _, alg := range signAlgorithm {
		signAlgs = append(signAlgs, jose.SignatureAlgorithm(alg))
		tokenProcessor.logger.Debugw("Added sign algorithm to token processor", "signAlgorithm", alg)
	}
	tokenProcessor.logger.Infow("Added sign algorithms to token processor")
	// TODO(dekiel) research if we can use jwt.DecodeSegment instead of jwt.ParseSigned
	//  to avoid parsing the token twice.
	parsedJWT, err := jwt.ParseSigned(tokenProcessor.rawToken, signAlgs)
	if err != nil {
		return fmt.Errorf("failed to parse oidc token: %w", err)
	}
	tokenProcessor.logger.Debugw("Parsed oidc token")
	err = parsedJWT.UnsafeClaimsWithoutVerification(claims)
	if err != nil {
		return fmt.Errorf("failed to get claims from unverified token: %w", err)
	}
	tokenProcessor.logger.Debugw("Got claims from unverified token", "claims", fmt.Sprintf("%+v", claims))
	tokenProcessor.issuer = claims.Issuer
	tokenProcessor.logger.Infow("Added issuer to TokenProcessor", "issuer", tokenProcessor.issuer)
	return nil
}

func (tokenProcessor *TokenProcessor) Issuer() string {
	for name, issuer := range TrustedOIDCIssuers {
		if tokenProcessor.issuer, exist := issuer[tokenProcessor.issuer]
		exist{
			tokenProcessor.logger.Debugw("Issuer is trusted", "name", name, "issuer", tokenProcessor.issuer),
			return name
		}
	}
	return tokenProcessor.issuer
}

func (tokenProcessor *TokenProcessor) VerifyToken(ctx context.Context, verifier TokenVerifierInterface) error {
	tokenProcessor.logger.Infow("Verifying oidc token")
	token, err := verifier.Verify(ctx, tokenProcessor.rawToken)
	if err != nil {
		return fmt.Errorf("failed to verify oidc token: %w", err)
	}
	tokenProcessor.token = token
	tokenProcessor.logger.Infow("OIDC token verified successfully")
	return nil
}

func (tokenProcessor *TokenProcessor) Claims(claims ClaimsInterface) error {
	err := tokenProcessor.token.Claims(claims)
	if err != nil {
		return fmt.Errorf("failed to get claims from token: %w", err)
	}
	return nil
}

func NewVerifierConfig(logger loggerInterface, clientID string, options ...VerifierConfigOptions) (*oidc.Config, error) {
	verifierConfig := &oidc.Config{
		ClientID:                   clientID,
		SkipClientIDCheck:          false,
		SkipExpiryCheck:            false,
		SkipIssuerCheck:            false,
		InsecureSkipSignatureCheck: false,
		SupportedSigningAlgs:       SupportedSigningAlgorithms,
	}
	logger.Debugw("Created verifier config with default values",
		"clientID", clientID,
		"SkipClientIDCheck", verifierConfig.SkipClientIDCheck,
		"SkipExpiryCheck", verifierConfig.SkipExpiryCheck,
		"SkipIssuerCheck", verifierConfig.SkipIssuerCheck,
		"InsecureSkipSignatureCheck", verifierConfig.InsecureSkipSignatureCheck,
		"SupportedSigningAlgs", verifierConfig.SupportedSigningAlgs,
	)
	logger.Debugw("Applying VerifierConfigOptions")
	for _, option := range options {
		err := option(verifierConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to apply VerifierConfigOption: %w", err)
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

// newVerifierFromDiscovery verifies the OIDC token using the public keys fetched from OIDC discovery.
// It supports issuers defined in OIDCIssuers variable.
// It writes the public keys to the file specified by the --public-key-path flag.
// It sets the environment variable specified by --new-public-keys-var-name to true.
func (tokenProcessor *TokenProcessor) NewVerifierFromDiscovery(ctx context.Context) (*oidc.IDTokenVerifier, error) {
	if tokenProcessor.issuer == "" {
		return nil, fmt.Errorf("TokenProcessor issuer is empty")
	}
	tokenProcessor.logger.Debugw("Creating provider using oidc discovery",
		"issuer", tokenProcessor.issuer,
	)
	provider, err := oidc.NewProvider(ctx, tokenProcessor.issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider using oidc discovery: %w", err)
	}
	tokenProcessor.logger.Debugw("Created provider using oidc discovery")
	return provider.Verifier(&tokenProcessor.verifierConfig), nil
}
