package main

import (
	"github.com/kyma-project/test-infra/development/tools/pkg/pjtester"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var (
	log     = logrus.New()
	rootCmd = &cobra.Command{
		Use:   "pjtester",
		Short: "pjtester generate new prowjob spec and schedule it",
		Long:  "pjtester  generate new prowjob spec from provided path. It reuse PR refs.",
		Run: func(cmd *cobra.Command, args []string) {
			pjtester.SchedulePJ()
		},
	}
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
