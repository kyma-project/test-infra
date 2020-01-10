package main

import (
	"flag"
)

var (
	name            = flag.String("name", "", "Service account name. [Required]")
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
	flag.Var(&myFlags, "role", "Role name which assign to sa. Multiple flag instances allowed.")
	flag.Parse()
}
