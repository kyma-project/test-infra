package main

import (
	tioidc "github.com/kyma-project/test-infra/pkg/github/oidc"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

// Cobra root command for the OIDC claim extractor
// Path: cmd/oidc/main.go

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
	// TrustedOIDCIssuers = map[string]string{"https://token.actions.githubusercontent.com": "https://token.actions.githubusercontent.com/.well-known/jwks"}
	rootCmd = &cobra.Command{
		Use:   "oidc",
		Short: "OIDC token verifier and claim extractor",
		Long: `oidc is a CLI tool to verify OIDC tokens and extract claims from them. It can use cached public keys to verify tokens.
	It use OIDC discovery to get the public keys and verify the token whenever the public keys are not cached or expired.`,
	}
	claimsCmd = &cobra.Command{
		Use:   "claims",
		Short: "Work with OIDC claims",
	}
	extractCmd = &cobra.Command{
		Use:   "extract",
		Short: "Verify token and extract claims from an OIDC token",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.extractClaims(); err != nil {
				return err
			}
			return nil
		},
	}
	opts = options{}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&opts.token, "token", "t", "", "OIDC token")
	err := rootCmd.MarkPersistentFlagRequired("token")
	if err != nil {
		panic(err)
	}
	rootCmd.PersistentFlags().StringVarP(&opts.newPublicKeysVarName, "new-public-keys-var-name", "n", "OIDC_NEW_PUBLIC_KEYS", "Name of the environment variable to set when new public keys are fetched")
	err = rootCmd.MarkPersistentFlagRequired("new-public-keys-var-name")
	if err != nil {
		panic(err)
	}
	rootCmd.PersistentFlags().StringSliceVarP(&opts.trustedWorkflows, "trusted-workflows", "w", []string{}, "List of trusted workflows")
	err = rootCmd.MarkPersistentFlagRequired("trusted-workflows")
	if err != nil {
		panic(err)
	}
	rootCmd.PersistentFlags().StringVarP(&opts.clientID, "client-id", "c", "", "OIDC token client ID")
	err = rootCmd.MarkPersistentFlagRequired("client-id")
	if err != nil {
		panic(err)
	}
	rootCmd.PersistentFlags().StringVarP(&opts.publicKeyPath, "public-key-path", "p", "", "Path to the public keys directory")
	rootCmd.PersistentFlags().BoolVarP(&opts.debug, "debug", "d", false, "Enable debug mode")
	extractCmd.PersistentFlags().StringVarP(&opts.outputPath, "claims-output-path", "o", "", "Path to write the extracted claims")
	err = extractCmd.MarkPersistentFlagRequired("claims-output-path")
	if err != nil {
		panic(err)
	}
	rootCmd.AddCommand(claimsCmd)
	claimsCmd.AddCommand(extractCmd)
}

// extractClaims verifies the OIDC token and extracts the claims from it.
// The OIDC token is read from the file specified by the --token flag.
// It returns an error if the token is invalid or the claims cannot be extracted.
// It uses cached public keys to verify the token.
// If the public keys are not cached or expired, it uses OIDC discovery to get the public keys.
// New public keys are written to the file specified by the --public-key-path flag.
// If new public keys are fetched, it sets ado environment variable to true.
// Extracted claims are written to the file specified by the --claims-output-path flag.
func (opts *options) extractClaims() error {
	// https://stackoverflow.com/questions/25609734/testing-stdout-with-go-and-ginkgo
	var (
		logger *zap.Logger
		err    error
	)
	if opts.debug {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		return err
	}
	sugaredLogger := logger.Sugar()

	verifyConfig, err := tioidc.NewVerifierConfig(sugaredLogger, opts.clientID)
	if err != nil {
		return err
	}

	tokenProcessor, err := tioidc.NewTokenProcessor(sugaredLogger, opts.token, *verifyConfig)
	if err != nil {
		return err
	}

	ctx := context.Background()
	verifier, err := tokenProcessor.NewVerifierFromDiscovery(ctx)
	if err != nil {
		return err
	}

	err = tokenProcessor.VerifyToken(ctx, verifier)
	if err != nil {
		return err
	}

	if tokenProcessor
		claims := tioidc.GithubClaims{}
	err = tokenProcessor.Claims(&claims)
	if err != nil {
		return err
	}
	return nil
}

// verifyToken verifies the OIDC token.
// It uses public keys loaded from the file specified by the --public-key-path flag.
// If the public keys are not available or expired, it fetches the public keys from the OIDC discovery endpoint.
// It writes the public keys to the file specified by the --public-key-path flag.
// It sets the environment variable specified by the --new-public-keys-var-name flag to true to indicate that new public keys are fetched.
// It returns an error if the token is invalid or the public keys are not available.
// func (opts *options) verifyToken() (*oidc.IDToken, error) {
// }
//
// func (opts *options) newVerifierFromStaticKeys() (*oidc.IDTokenVerifier, error) {
//
// }

// loadPublicKeysFromLocal loads the public keys from the file specified by the --public-key-path flag.
// example implementation https://gist.github.com/nilsmagnus/199d56ce849b83bdd7df165b25cb2f56
// func (opts *options) loadPublicKeysFromLocal() error {
//
// }
//
// func (opts *options) loadPublicKeysFromRemote(issuer string) error {
//
// }

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
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
