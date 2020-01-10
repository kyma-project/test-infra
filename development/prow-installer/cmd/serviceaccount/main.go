package main

import (
	"context"
	"flag"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/serviceaccount"
	"log"
	"os"
)

var (
	name            = flag.String("name", "", "Service account name. [Required]")
	projectname     = flag.String("project", "", "Project name for which create service account. [Required]")
	credentialsfile = flag.String("credentialsfile", "", "Google Application Credentials file path. [Optional]")
	prefix          = flag.String("prefix", "", "Prefix for naming resources. [Optional]")
)

func main() {
	flag.Parse()
	if *credentialsfile == "" && os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		log.Fatalf("Requires the environment variable GOOGLE_APPLICATION_CREDENTIALS to be set to a GCP service account file.")
	}
	if *name == "" {
		log.Fatalf("Missing required argument : -name")
	}
	if *projectname == "" {
		log.Fatalf("Missing required argument : -project")
	}
	ctx := context.Background()
	saOptions := serviceaccount.SAOptions{
		Name:    *name,
		Roles:   nil,
		Project: *projectname,
	}
	iam := serviceaccount.NewIAMClient(*credentialsfile, *prefix, ctx)
	iam.CreateSAAccount(saOptions)
}
