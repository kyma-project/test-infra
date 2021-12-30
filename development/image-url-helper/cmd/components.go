package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/kyma-project/test-infra/development/image-url-helper/pkg/common"
	"github.com/kyma-project/test-infra/development/image-url-helper/pkg/component"
	"github.com/kyma-project/test-infra/development/image-url-helper/pkg/list"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"sigs.k8s.io/yaml"
)

// ComponentsCmd generates component descripto file with all images used in Kyma
func ComponentsCmd() *cobra.Command {
	options := component.ComponentOptions{}
	cmd := &cobra.Command{
		Use:     "components",
		Short:   "Generates component descriptor file for Kyma",
		Long:    "Generates component descriptor file for Kyma from values.yaml files",
		Example: "image-url-helper components --component-version 0.1.0 --git-commit 123456 --git-branch main",
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			// remove trailing slash to have consistent paths
			ResourcesDirectoryClean := filepath.Clean(ResourcesDirectory)

			images := make(common.ComponentImageMap)
			testImages := make(common.ComponentImageMap)

			err := filepath.Walk(ResourcesDirectory, list.GetWalkFunc(ResourcesDirectoryClean, images, testImages))
			if err != nil {
				fmt.Printf("Cannot traverse directory: %s\n", err)
				os.Exit(2)
			}
			allImages := make(common.ComponentImageMap)
			common.MergeImageMap(allImages, images)
			common.MergeImageMap(allImages, testImages)

			componentDescriptor, err := component.GenerateComponentDescriptor(options, allImages)
			if err != nil {
				log.Fatalf("Cannot generate compoent descriptor: %s", err)
			}

			encodedComponentDescriptor, err := yaml.Marshal(componentDescriptor)
			if err != nil {
				log.Fatalf("failed to generate component descriptor: %s", err)
			}

			// try decoding the component descriptor to see if it will at least parse
			err = component.SanityCheck(encodedComponentDescriptor)
			if err != nil {
				fmt.Println("Validation check failed, generated YAML file:")
				fmt.Println(string(encodedComponentDescriptor))
				log.Fatalf("failed sanity check: %s", err)
			}

			if options.OutputDir != "" {
				outputDirClean := path.Clean(options.OutputDir)
				err = os.MkdirAll(outputDirClean, os.ModePerm)
				if err != nil {
					log.Fatalf("failed to create output directory: %s", err)
				}

				ioutil.WriteFile(outputDirClean+"/component-descriptor.yaml", encodedComponentDescriptor, 0666)
			}

			if options.RepoContext != "" {
				err = component.PushDescriptor(encodedComponentDescriptor, options.RepoContext)
				if err != nil {
					log.Fatalf("failed to push component descriptor: %s", err)
				}
			}

		},
	}
	addComponentCmdFlags(cmd, &options)
	return cmd
}

func addComponentCmdFlags(cmd *cobra.Command, options *component.ComponentOptions) {
	cmd.Flags().StringVarP(&options.ComponentName, "component-name", "n", "github.com/kyma-project/kyma", "name of the component")
	cmd.Flags().StringVarP(&options.ComponentVersion, "component-version", "v", "", "component version")
	cmd.MarkFlagRequired("component-version")

	cmd.Flags().StringVarP(&options.Provider, "provider", "p", "internal", "Component provider (internal/external)")

	cmd.Flags().StringVarP(&options.GitCommit, "git-commit", "c", "", "Git commit hash")
	viper.BindEnv("git-commit", "PULL_PULL_SHA")
	cmd.MarkFlagRequired("git-commit")

	cmd.Flags().StringVarP(&options.GitBranch, "git-branch", "b", "", "Git branch name")
	viper.BindEnv("git-branch", "PULL_BASE_REF")
	cmd.MarkFlagRequired("git-branch")

	cmd.Flags().BoolVarP(&options.SkipHashConversion, "skip-hash-conversion", "s", false, "Keeps the image tags unchanged, without conversion to hashes")

	cmd.Flags().StringVarP(&options.OutputDir, "output-dir", "o", "", "Name of the output directory")
	cmd.Flags().StringVarP(&options.RepoContext, "repo-context", "C", "", "Name of the Docker repository to push component descriptor to")

	// use values form enviroment when a flag was not provided
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if !f.Changed && viper.IsSet(f.Name) {
			val := viper.Get(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}
