package prowjob_parser

import (
	"github.com/jamiealquiza/envy"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	var log = logrus.New()
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)

	var rootCmd = &cobra.Command{
		Use:   "image-syncer",
		Short: "image-syncer copies images between docker registries",
		Long:  `image-syncer copies docker images. It compares checksum between source and target and protects target images against overriding`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := copyImages(cfg); err != nil {
				log.Fatal(err)
			}
		},
	}

	rootCmd.PersistentFlags().StringVarP(&cfg.ImagesFile, "images-file", "i", "", "yaml file containing list of images")
	rootCmd.PersistentFlags().StringVarP(&cfg.TargetKeyFile, "target-key-file", "t", "", "JSON key file used for authorization to target repo")
	rootCmd.PersistentFlags().BoolVarP(&cfg.DryRun, "dry-run", "d", true, "dry run mode")

	rootCmd.MarkPersistentFlagRequired("images-file")
	rootCmd.MarkPersistentFlagRequired("target-key-file")
	envy.ParseCobra(rootCmd, envy.CobraConfig{Prefix: "SYNCER", Persistent: true, Recursive: false})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
