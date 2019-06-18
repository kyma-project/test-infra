package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kyma-project/test-infra/development/tools/pkg/yamlmerge"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
)

var (
	path   = flag.String("path", "", "Path to yaml folder [Required]")
	target = flag.String("target", "", "Path to target file [Required]")
	dryRun = flag.Bool("dryRun", true, "Dry Run enabled, nothing is overriden")
)

func main() {
	flag.Parse()

	common.ShoutFirst("Running with arguments: path: \"%s\", dryRun: %t", *path, *dryRun)
	// ctx := context.Background()

	if *path == "" {
		fmt.Fprintln(os.Stderr, "missing -path flag")
		flag.Usage()
		os.Exit(2)
	}

	if *target == "" {
		fmt.Fprintln(os.Stderr, "missing -target flag")
		flag.Usage()
		os.Exit(2)
	}

	yamlmerge.MergeFiles(*path, ".yaml", *target, !*dryRun)

	common.Shout("Finished")
}
