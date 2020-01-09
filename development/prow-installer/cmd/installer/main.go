package main

import (
	"flag"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-project/test-infra/development/prow-installer/pkg/accessmanager"
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

	var InstallerConfig installer.InstallerConfig
	InstallerConfig.ReadConfig(*config)

	AccessManager := accessmanager.NewAccessManager(*credentialsfile)

	for _, account := range InstallerConfig.ServiceAccounts {
		_ = AccessManager.IAM.CreateSAAccount(account.Name, InstallerConfig.Project)
	}
	AccessManager.Projects.GetProjectPolicy(InstallerConfig.Project)
	log.Printf("%+v", AccessManager.Projects.Projects[InstallerConfig.Project].Policy)
	//AccessManager.Projects.AssignRoles(InstallerConfig.Project, InstallerConfig.ServiceAccounts)
}
