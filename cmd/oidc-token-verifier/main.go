package main

import (
	"fmt"
	"os"

	"github.com/kyma-project/test-infra/pkg/logging"
	tioidc "github.com/kyma-project/test-infra/pkg/oidc"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

// Cobra root command for the OIDC claim extractor
// Path: cmd/oidc-token-verifier/main.go

type Logger interface {
	logging.StructuredLoggerInterface
	logging.WithLoggerInterface
}

type options struct {
	token                   string
	debug                   bool
	oidcTokenExpirationTime int // OIDC token expiration time in minutes
}

var (
	rootCmd   *cobra.Command
	verifyCmd *cobra.Command
	opts      = options{}
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "token",
		Short: "OIDC token verifier and claim extractor",
		Long: `oidc is a CLI tool to verify OIDC tokens and extract claims from them. It can use cached public keys to verify tokens.
	It uses OIDC discovery to get the public keys and verify the token whenever the public keys are not cached or expired.`,
	}
	rootCmd.PersistentFlags().StringVarP(&opts.token, "token", "t", "", "OIDC token to verify")
	rootCmd.PersistentFlags().BoolVarP(&opts.debug, "debug", "d", false, "Enable debug mode")
	rootCmd.PersistentFlags().IntVarP(&opts.oidcTokenExpirationTime, "oidc-token-expiration-time", "e", 10, "OIDC token expiration time in minutes")
	return rootCmd
}

func NewVerifyCmd() *cobra.Command {
	verifyCmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify token and expected claims values",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := opts.verifyToken(); err != nil {
				return err
			}
			return nil
		},
	}
	return verifyCmd

}

func init() {
	rootCmd = NewRootCmd()
	verifyCmd = NewVerifyCmd()
	rootCmd.AddCommand(verifyCmd)
}

// isTokenProvided checks if the token flag is set.
// If not, check if AUTHORIZATION environment variable is set.
// If neither is set, return an error.
func isTokenProvided(logger Logger, opts *options) error {
	if opts.token == "" {
		logger.Infow("Token flag not provided, checking for AUTHORIZATION environment variable")
		opts.token = os.Getenv("AUTHORIZATION")
		if opts.token == "" {
			return fmt.Errorf("token not provided, set the --token flag or the AUTHORIZATION environment variable with the OIDC token")
		}
		logger.Infow("Token found in AUTHORIZATION environment variable, using the token")
	} else {
		logger.Infow("Token flag provided, using the token from the flag")
	}
	logger.Debugw("Token value", "token", opts.token)
	return nil
}

// extractClaims verifies the OIDC token.
// The OIDC token is read from the file specified by the --token flag or the AUTHORIZATION environment variable.
// It returns an error if the token validation failed.
// It verifies the token signature and expiration time, verifies if the token is issued by a trusted issuer,
// and the claims have expected values.
// It uses OIDC discovery to get the identity provider public keys.
func (opts *options) verifyToken() error {
	var (
		zapLogger *zap.Logger
		err       error
		token     *tioidc.Token
	)
	if opts.debug {
		zapLogger, err = zap.NewDevelopment()
	} else {
		zapLogger, err = zap.NewProduction()
	}
	if err != nil {
		return err
	}
	logger := zapLogger.Sugar()

	err = isTokenProvided(logger, opts)
	if err != nil {
		return err
	}

	// Create a new token processor
	// It reads issuer from the token and verifies if the issuer is trusted.
	// The tokenProcessor is a main object that is used to verify the token and extract the claim values.
	// TODO(dekiel): add support for providing trusted issuers instead of using the value from the package.
	tokenProcessor, err := tioidc.NewTokenProcessor(logger, tioidc.TrustedOIDCIssuers, opts.token)
	if err != nil {
		return err
	}

	logger.Infow("Token processor created for trusted issuer", "issuer", tokenProcessor.Issuer())

	// TODO (dekiel): implement writing output data to the file. This will give us separated clear output for a data and logs.
	fmt.Printf("GITHUB_URL=%s\n", tokenProcessor.GetIssuer().GetGithubURL())

	// Create a new verifier config that will be used to verify the token.
	// The standard expiration check is skipped.
	// We use custom expiration time check to allow longer token expiration time than the value in the token.
	// The extended expiration time is needed due to Azure DevOps delays in starting the pipeline.
	// The delay was causing the token to expire before the pipeline started.
	verifierConfig, err := tokenProcessor.NewVerifierConfig(tioidc.SkipExpiryCheck())
	if err != nil {
		return err
	}

	logger.Infow("Verifier config created")

	ctx := context.Background()
	// Create a new provider using OIDC discovery to get the public keys.
	// It uses the issuer from the token to get the OIDC discovery endpoint.
	provider, err := tioidc.NewProviderFromDiscovery(ctx, logger, tokenProcessor.Issuer())
	if err != nil {
		return err
	}
	logger.Infow("Provider created using OIDC discovery", "issuer", tokenProcessor.Issuer())

	// Create a new verifier using the provider and the verifier config.
	// The verifier is used to verify the token signature, expiration time and execute standard OIDC validation.
	// TODO (dekiel): Consider using verifier config as the only way to parametrize the verification process.
	//   The WithExtendedExpiration could be moved to the verifier config.
	//   The WithExtendedExpiration could disable the standard expiration check.
	//   This would allow to have a single place to configure the verification process.
	verifier, err := provider.NewVerifier(logger, verifierConfig, tioidc.WithExtendedExpiration(opts.oidcTokenExpirationTime))
	if err != nil {
		return err
	}

	logger.Infow("New verifier created")

	// Verify the token
	token, err = verifier.Verify(ctx, opts.token)
	if err != nil {
		return err
	}

	logger.Infow("Token verified successfully")

	// Create claims
	claims := tioidc.NewClaims(logger)

	logger.Infow("Verifying token claims")

	// Pass the token to ValidateClaims
	err = tokenProcessor.ValidateClaims(&claims, token)

	if err != nil {
		return err
	}
	logger.Infow("Token claims expectations verified successfully")
	logger.Infow("All token checks passed successfully")

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
