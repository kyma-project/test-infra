package main

import (
	"flag"
	"fmt"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/serviceaccount"
	"log"
)

var (
	//name            = flag.String("name", "", ".. Service account name. [Required]")
	project         = flag.String("project", "", "GCP project name. [Required]")
	prefix          = flag.String("prefix", "", "Prefix for naming resources. [Optional]")
	credentialsfile = flag.String("credentialsfile", "", "Google Application Credentials file path. [Required]")
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var myFlags arrayFlags

func main() {
	flag.Var(&myFlags, "name", ".. Service account name. [Required]")
	flag.Parse()
	if *credentialsfile == "" {
		log.Fatalf("Argument credentialsfile is missing or empty.")
	}

	iamservice, err := serviceaccount.NewService(*credentialsfile)
	iamclient := serviceaccount.NewClient(*prefix, &iamservice)
	if err != nil {
		log.Fatalf("When creating serviceaccount got error: %w", err)
	}
	for _, value := range myFlags {
		fmt.Printf("Creating service account with values:\nname: %s\nproject: %s\nprefix: %s\n", value, *project, *prefix)
		options := serviceaccount.SAOptions{
			Name:    value,
			Roles:   nil,
			Project: *project,
		}
		iamclient.CreateSA(options)
	}

}
