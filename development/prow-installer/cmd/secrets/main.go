package main

import (
	"context"
	"flag"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-project/test-infra/development/prow-installer/pkg/installer"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/secrets"
)

var (
	config          = flag.String("config", "", "Config file path [Required]")
	credentialsfile = flag.String("credentialsfile", "", "Google Application Credentials file path [Required]")
	prefix          = flag.String("prefix", "", "Prefix for naming resources [Optional]")
)

func main() {
	flag.Parse()

	if *config == "" {
		log.Fatalf("Missing required argument : -config")
	}

	if *prefix == "" {
		log.Fatalf("Missing required argument : -prefix")
	}

	var InstallerConfig installer.InstallerConfig
	InstallerConfig.ReadConfig(*config)

	ctx := context.Background()
	client, err := secrets.New(ctx, secrets.Option{Prefix: *prefix})
	if err != nil {
		log.Errorf("Could not create SecretClient: %v", err)
		os.Exit(1)
	}
	client.ReadSecret("daniel-test", "test", "test")
}
