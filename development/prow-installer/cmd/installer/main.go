package main

import (
	"flag"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-project/test-infra/development/prow-installer/pkg/gcpaccessmanager"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/installer"
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

	if *credentialsfile == "" {
		log.Fatalf("Missing required argument : -credentialsfile")
	}

	var InstallerConfig installer.Config
	InstallerConfig.ReadConfig(*config)

	AccessManager := gcpaccessmanager.NewAccessManager(*credentialsfile)

	for _, account := range InstallerConfig.ServiceAccounts {
		_ = AccessManager.iam.createSAAccount(account.Name, InstallerConfig.Project)
	}
	AccessManager.projects.GetProjectPolicy(InstallerConfig.Project)
	log.Printf("%+v", AccessManager.projects.projects[InstallerConfig.Project].policy)
	//AccessManager.projects.AssignRoles(InstallerConfig.Project, InstallerConfig.ServiceAccounts)
}
