package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamiealquiza/envy"
	image_url_helper "github.com/kyma-project/test-infra/development/image-url-helper/pkg/image-url-helper"
	"github.com/spf13/cobra"
)

type checkCmdOptions struct {
	skipComments bool
}

func CheckCmd() *cobra.Command {
	options := checkCmdOptions{}
	cmd := &cobra.Command{
		Use:     "check",
		Short:   "aaaa",
		Long:    "aaa",
		Example: "image-url-helper check",
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			foundIncompatible := false

			skipComentsInfo := ""
			if options.skipComments {
				skipComentsInfo = ", excluding commented out lines"
			}

			// for all files in resources
			fmt.Printf("Looking for incompatible images in \"%s\"%s:\n\n", ResourcesDirectory, skipComentsInfo)

			err := filepath.Walk(ResourcesDirectory, getWalkFunc(&foundIncompatible, options.skipComments))
			if err != nil {
				fmt.Printf("Cannot traverse directory: %s", err)
				os.Exit(2)
			}

			if foundIncompatible {
				fmt.Printf("\nFound incompatible image lines\n")
				os.Exit(3)
			} else {
				fmt.Println("All images seems to be in the new format")
			}
		},
	}
	addCheckCmdFlags(cmd, &options)
	return cmd
}

func addCheckCmdFlags(cmd *cobra.Command, options *checkCmdOptions) {
	cmd.Flags().BoolVarP(&options.skipComments, "skipComments", "s", false, "Skip commented out lines")
	envy.ParseCobra(cmd, envy.CobraConfig{Persistent: true, Prefix: "IMAGE_URL_HELPER"})
}

func getWalkFunc(foundIncompatible *bool, skipComments bool) filepath.WalkFunc {
	return func(path string, info fs.FileInfo, err error) error {
		//pass the error further, this shouldn't ever happen
		if err != nil {
			return err
		}

		// skip directory entries, we just want files
		if info.IsDir() {
			return nil
		}

		// we only want to check .yaml files
		if !strings.Contains(info.Name(), ".yaml") {
			return nil
		}

		// check if this file contains any image: lines that aren't using new templates
		incompatible, err := image_url_helper.FileHasIncorrectImage(path, skipComments)
		if err != nil {
			return nil
		}

		if incompatible {
			*foundIncompatible = true
		}

		return nil
	}
}
