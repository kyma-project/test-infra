package cmd

import "github.com/spf13/cobra"

var (
	ResourcesDirectory string
	rootCmd            = &cobra.Command{
		Use:   "image-url-helper",
		Short: "Image URL helper CLI",
		Long:  "Command-line tool to perform image listing and checks.",
	}
)

func Execute() error {
	rootCmd.AddCommand(CheckCmd())
	rootCmd.AddCommand(ListCmd())

	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&ResourcesDirectory, "resources-directory", "r", "", "Path to resources directory")
	rootCmd.MarkPersistentFlagRequired("resources-directory")
}
