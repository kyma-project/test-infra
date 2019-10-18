package main

import (
	"flag"
	"fmt"
	"log"
	"prow_installer/accessmanager"
)

func main() {
	flag.Parse()

	if *config == "" {
		log.Fatalf("Missing required argument : -config")
	}

	if *credentialsfile == "" {
		log.Fatalf("Missing required argument : -credentialsfile")
	}

	var InstallerConfig installerConfig
	_ = getInstallerConfig(*config, &InstallerConfig)

	AccessManager := accessmanager.NewAccessManager(*credentialsfile, InstallerConfig.Project, *prefix)

	for _, account := range InstallerConfig.ServiceAccounts {
		_ = AccessManager.SaAccounts.CreateSAAccount(account.Name)
	}
	AccessManager.Policies.GetProjectPolicy()
	for k, v := range AccessManager.Policies.Bindings {
		fmt.Printf("key[%s] value[%s]\n", k, v.Members)
	}
}
