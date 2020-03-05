package main

import (
	"flag"
	"fmt"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/roles"
	"google.golang.org/api/cloudresourcemanager/v1"

	log "github.com/sirupsen/logrus"
)

var (
	saname          = flag.String("saname", "", ".. Service account name. [Required]")
	project         = flag.String("project", "", "GCP project name. [Required]")
	credentialsfile = flag.String("credentialsfile", "", "Google Application Credentials file path. [Required]")
	expression      = flag.String("expression", "", "GCP policy binding condition expression string. [Optional]")
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var roleFlags arrayFlags

func main() {
	flag.Var(&roleFlags, "roles", ".. GCP role name which assign to serviceaccount. [Required]")
	flag.Parse()
	if *credentialsfile == "" {
		log.Fatalf("Argument credentialsfile is missing or empty.")
	}
	crmservice, err := roles.NewService(*credentialsfile)
	if err != nil {
		log.Fatalf("When assinging role for serviceaccount got error: %s", err.Error())
	}
	crmclient, err := roles.New(crmservice)
	if err != nil {
		log.Fatalf("Failed creating cloudresourcemanager client.")
	}
	var condition *cloudresourcemanager.Expr
	if *expression == "" {
		condition = nil
	} else {
		condition = &cloudresourcemanager.Expr{Expression: *expression}
	}
	_, err = crmclient.AddSAtoRole(*saname, roleFlags, *project, condition)
	if err != nil {
		log.Fatalf("When assigning role for serviceaccount got error: %s", err.Error())
	} else {
		log.Printf("Roles added to serviceaccount: %s\nroles: %v", *saname, roleFlags)
	}
}
