package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kyma-project/test-infra/development/owners_changer/pkg"
	"github.com/spf13/cobra"

	"github.com/hairyhenderson/go-codeowners"
)

var (
	rootCmd = &cobra.Command{
		Use:   "image-url-helper",
		Short: "Image URL helper CLI",
		Long:  "Command-line tool to perform image listing and checks.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			repoDirectory := args[0]
			repoDirectoryClean := filepath.Clean(repoDirectory)

			ownersAliases, err := pkg.GetOwnersAliases(repoDirectoryClean)
			if err != nil {
				fmt.Printf("Cannot parse OWNERS_ALIASES: %s\n", err)
				os.Exit(2)
			}

			outputCodeOwners := make([]codeowners.Codeowner, 0)

			err = filepath.Walk(repoDirectoryClean, pkg.GetWalkFunc(repoDirectoryClean, ownersAliases, &outputCodeOwners))
			if err != nil {
				fmt.Printf("Cannot traverse directory: %s\n", err)
				os.Exit(2)
			}

			for _, codeowner := range outputCodeOwners {
				ownerString := strings.ReplaceAll(codeowner.String(), "\t", " ")

				fmt.Println(ownerString)
			}
		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}
