package cmd

import "github.com/spf13/cobra"

var (
	// ResourcesDirectory stores path to the Kyma resources directory
	ResourcesDirectory string
	rootCmd            = &cobra.Command{
		Use:   "image-url-helper",
		Short: "Image URL helper CLI",
		Long:  "Command-line tool to perform image listing and checks.",
	}
)

// Execute is a main Cobra function
func Execute() error {
	rootCmd.AddCommand(CheckCmd())
	rootCmd.AddCommand(ListCmd())
	rootCmd.AddCommand(PromoteCmd())
	rootCmd.AddCommand(MissingCmd())

	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&ResourcesDirectory, "resources-directory", "r", "", "Path to resources directory")
	rootCmd.MarkPersistentFlagRequired("resources-directory")
}
# (2025-03-04)