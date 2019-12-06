package main

import "flag"

imrpot (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/gcpaccessmanager"
)

var (
	name 			= flag.String("name", "", "Service account name. [Required]")
	roles			= flag.String("roles", "", "Role name which assign to sa. Multiple flag")
	credentialsfile = flag.String("credentialsfile", "", "Google Application Credentials file path. [Required]")
	prefix          = flag.String("prefix", "", "Prefix for naming resources. [Optional]")
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var myFlags arrayFlags

func main() {
	flag.Var(&myFlags, "list1", "Some description for this param.")
	flag.Parse()
}