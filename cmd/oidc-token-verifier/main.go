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
	token                string
	clientID             string
	outputPath           string
	publicKeyPath        string
	newPublicKeysVarName string
	trustedWorkflows     []string
	debug                bool
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
	rootCmd.PersistentFlags().StringVarP(&opts.newPublicKeysVarName, "new-keys-var", "n", "OIDC_NEW_PUBLIC_KEYS", "Name of the environment variable to set when new public keys are fetched")
	// This flag should be enabled once we add support for it in the code.
	// rootCmd.PersistentFlags().StringSliceVarP(&opts.trustedWorkflows, "trusted-workflows", "w", []string{}, "List of trusted workflows")
	// err := rootCmd.MarkPersistentFlagRequired("trusted-workflows")
	// if err != nil {
	// 	panic(err)
	// }
	rootCmd.PersistentFlags().StringVarP(&opts.clientID, "client-id", "c", "image-builder", "OIDC token client ID, this is used to verify the audience claim in the token. The value should be the same as the audience claim value in the token.")
	rootCmd.PersistentFlags().StringVarP(&opts.publicKeyPath, "public-key-path", "p", "", "Path to the cached public keys directory")
	rootCmd.PersistentFlags().BoolVarP(&opts.debug, "debug", "d", false, "Enable debug mode")
	return rootCmd
}

func NewVerifyCmd() *cobra.Command {
	verifyCmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify token and expected claims values",
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Printf("GITHUB_URL=%s\n", "https://github.com")
			//if err := opts.extractClaims(); err != nil {
			//	return err
			//}
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
func (opts *options) extractClaims() error {
	var (
		zapLogger *zap.Logger
		err       error
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

	// Print used options values.
	logger.Infow("Using the following trusted workflows", "trusted-workflows", opts.trustedWorkflows)
	logger.Infow("Using the following client ID", "client-id", opts.clientID)
	logger.Infow("Using the following public key path", "public-key-path", opts.publicKeyPath)
	logger.Infow("Using the following new public keys environment variable", "new-keys-var", opts.newPublicKeysVarName)
	logger.Infow("Using the following claims output path", "claims-output-path", opts.outputPath)

	// Create a new verifier config that will be used to verify the token.
	// The clientID is used to verify the audience claim in the token.
	verifyConfig, err := tioidc.NewVerifierConfig(logger, opts.clientID)
	if err != nil {
		return err
	}
	logger.Infow("Verifier config created", "config", verifyConfig)

	// Create a new token processor
	// It reads issuer from the token and verifies if the issuer is trusted.
	// The tokenProcessor is a main object that is used to verify the token and extract the claim values.
	// TODO(dekiel): add support for providing trusted issuers instead of using the value from the package.
	tokenProcessor, err := tioidc.NewTokenProcessor(logger, tioidc.TrustedOIDCIssuers, opts.token, verifyConfig)
	if err != nil {
		return err
	}
	logger.Infow("Token processor created for trusted issuer", "issuer", tokenProcessor.Issuer())
	fmt.Printf("GITHUB_URL=%s\n", tokenProcessor.GetIssuer().GetGithubURL())

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
	verifier := provider.NewVerifier(logger, verifyConfig)
	logger.Infow("New verifier created")

	// claims will store the extracted claim values from the token.
	claims := tioidc.NewClaims(logger)
	// Verifies the token and check if the claims have expected values.
	// Verifies custom claim values too.
	// Extract the claim values from the token into the claims struct.
	// It provides a final result if the token is valid and the claims have expected values.
	err = tokenProcessor.VerifyAndExtractClaims(ctx, &verifier, &claims)
	if err != nil {
		return err
	}
	logger.Infow("Token verified successfully")

	return nil
}

// If the public keys are not cached or expired, it uses OIDC discovery to get the public keys.
// New public keys are written to the file specified by the --public-key-path flag.
// If new public keys are fetched, it sets ado environment variable to true.

// loadPublicKeysFromLocal loads the public keys from the file specified by the --public-key-path flag.
// example implementation https://gist.github.com/nilsmagnus/199d56ce849b83bdd7df165b25cb2f56
// func (opts *options) loadPublicKeysFromLocal() error {
//
// }
//

// savePublicKeysFromRemote fetches the public keys from the OIDC discovery endpoint.
// It writes the public keys to the file specified by the --public-key-path flag.
// It sets the environment variable specified by --new-public-keys-var-name to true to indicate that new public keys are fetched.
// func (opts *options) savePublicKeysFromRemote(issuer string) error {
//
// }

// setAdoEnvVar sets the Azure DevOps pipeline environment variable to true.
// Environment variable name is specified by --new-public-keys-var-name flag.
// func (opts *options) setAdoEnvVar() error {
//
// }

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
