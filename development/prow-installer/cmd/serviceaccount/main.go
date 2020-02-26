package main

import (
	//"encoding/json"
	"flag"
	"fmt"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/serviceaccount"
	"log"
	"os"
)

var (
	project         = flag.String("project", "", "GCP project name. [Required]")
	prefix          = flag.String("prefix", "", "Prefix for naming resources. [Optional]")
	credentialsfile = flag.String("credentialsfile", "", "Google Application Credentials file path. [Required]")
	remove          = flag.Bool("remove", false, "When set, installer will remove resources defined in config. Default false. [Optional]")
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
	if err != nil {
		log.Fatalf("When creating serviceaccount got error: %w", err)
	}
	iamclient := serviceaccount.NewClient(*prefix, iamservice)

	if *remove {
		for _, value := range myFlags {
			_, err := iamclient.Delete(value)
			if err != nil {
				log.Printf("%v", err)
			}
		}
		os.Exit(0)
	}

	for _, value := range myFlags {
		log.Printf("Creating service account with values:\nname: %s\nproject: %s\nprefix: %s\n", value, *project, *prefix)
		sa, err := iamclient.CreateSA(value, *project)
		if err != nil {
			log.Printf("Failed create serviceaccount %s, got error: %w", value, err)
		}
		key, err := iamclient.CreateSAKey(sa.Email)
		if err != nil {
			log.Printf("Failed create key for serviceaccount %s, got error: %w", value, err)
		}
		fmt.Println(key)
		log.Printf("Got key for serviceaccount: %s", value)
	}

}
