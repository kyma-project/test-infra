package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kyma-project/test-infra/development/image-url-helper/pkg/component"
	"github.com/kyma-project/test-infra/development/image-url-helper/pkg/list"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"sigs.k8s.io/yaml"
)

// type componentCmdOptions struct {
// 	componentName    string // github.com/kyma-project/kyma
// 	componentVersion string
// 	appName          string
// 	outputDir        string
// 	repoContext      string
// }

// ComponentsCmd generates
func ComponentsCmd() *cobra.Command {
	options := component.ComponentOptions{}
	cmd := &cobra.Command{
		Use:     "components",
		Short:   "Check if all images use new format",
		Long:    "Find all image usages that doesn't use imageurl template",
		Example: "image-url-helper list",
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			imageComponentsMap := make(list.ImageToComponents)

			// remove trailing slash to have consistent paths
			ResourcesDirectoryClean := filepath.Clean(ResourcesDirectory)

			images := make(list.ImageMap)
			testImages := make(list.ImageMap)

			err := filepath.Walk(ResourcesDirectory, list.GetWalkFunc(ResourcesDirectoryClean, images, testImages, imageComponentsMap))
			if err != nil {
				fmt.Printf("Cannot traverse directory: %s\n", err)
				os.Exit(2)
			}
			allImages := make(list.ImageMap)
			list.MergeImageMap(allImages, images)
			list.MergeImageMap(allImages, testImages)

			componentDescriptor, err := component.GenerateComponentDescriptor(options, allImages)
			if err != nil {
				log.Fatalf("Cannot generate compoent descriptor: %s", err)
			}

			encodedComponentDescriptor, err := yaml.Marshal(componentDescriptor)
			if err != nil {
				log.Fatalf("failed to generate component descriptor: %s", err)
			}

			// try decoding the component descriptor to see if it will at least parse
			// TODO move this to tests?
			err = component.SanityCheck(encodedComponentDescriptor)
			if err != nil {
				fmt.Println("Validation check failed, generated YAML file:")
				fmt.Println(string(encodedComponentDescriptor))
				log.Fatalf("failed sanity check: %s", err)
			} else {
				fmt.Println(string(encodedComponentDescriptor))
			}

		},
	}
	addComponentCmdFlags(cmd, &options)
	return cmd
}

func addComponentCmdFlags(cmd *cobra.Command, options *component.ComponentOptions) {
	cmd.Flags().StringVarP(&options.ComponentName, "component-name", "n", "github.com/kyma-project/kyma", "name of the component")
	cmd.Flags().StringVarP(&options.ComponentVersion, "component-version", "v", "", "component version")

	cmd.Flags().StringVarP(&options.Provider, "provider", "p", "internal", "Component provider (internal/external)")

	cmd.Flags().StringVarP(&options.GitCommit, "git-commit", "c", "", "Git commit hash")
	viper.BindEnv("git-commit", "PULL_PULL_SHA")

	cmd.Flags().StringVarP(&options.GitBranch, "git-branch", "b", "", "Git branch name")
	viper.BindEnv("git-branch", "PULL_BASE_REF")

	//envy.ParseCobra(cmd, envy.CobraConfig{Persistent: true, Prefix: "IMAGE_URL_HELPER"})
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if !f.Changed && viper.IsSet(f.Name) {
			val := viper.Get(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}
