package main

import (
	"fmt"
	"os"

	"github.com/kyma-project/test-infra/pkg/tools/prtagbuilder"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	v1 "k8s.io/test-infra/prow/apis/prowjobs/v1"
	"k8s.io/test-infra/prow/pod-utils/downwardapi"
)

var (
	log     = logrus.New()
	rootCmd = &cobra.Command{
		Use:   "prtagbuilder [-o, --org string], [-r, --repo string], [-b --baseref string], [-O --numberonly]",
		Short: "prtagbuilder will find pull request number for commit or branch head.",
		//nolint:revive
		RunE: func(cmd *cobra.Command, args []string) error {
			numFlags := cmd.Flags().NFlag()
			if numFlags == 1 && !cmd.Flags().Changed("numberonly") || numFlags > 1 {
				err := checkFlags(cmd.Flags())
				if err != nil {
					return fmt.Errorf("required flag is empty, got error: %w", err)
				}
				if numFlags > 2 {
					fromFlags = true
				}
			}
			ghClient := prtagbuilder.NewGitHubClient(nil)
			prNumber, err := prtagbuilder.BuildPrTag(jobSpec, fromFlags, numberOnly, ghClient)
			if err != nil {
				return fmt.Errorf("failed build prtag, got error: %w", err)
			}
			fmt.Print(prNumber)
			return nil
		},
	}
	jobSpec    *downwardapi.JobSpec
	fromFlags  = false
	numberOnly bool
)

// checkFlags checks if flags required to be set together are not empty
func checkFlags(cmdFlags *pflag.FlagSet) error {
	// flags names required to be set together
	flagNames := []string{"org", "repo", "baseref"}
	for index, val := range flagNames {
		flagValue, _ := cmdFlags.GetString(val)
		if flagValue == "" {
			return fmt.Errorf("flag %s is empty", flagNames[index])
		}
	}
	return nil
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)
	jobSpec = &downwardapi.JobSpec{Refs: &v1.Refs{}}
	rootCmd.PersistentFlags().StringVarP(&jobSpec.Refs.Org, "org", "o", "", "Github organisation which owns repo.")
	rootCmd.PersistentFlags().StringVarP(&jobSpec.Refs.Repo, "repo", "r", "", "Github repository.")
	rootCmd.PersistentFlags().StringVarP(&jobSpec.Refs.BaseRef, "baseref", "b", "", "Base branch name.")
	rootCmd.PersistentFlags().BoolVarP(&numberOnly, "numberonly", "O", false, "Return only PR number.")
	if err := rootCmd.Execute(); err != nil {
		logrus.WithError(err).Fatalf("prtagbuilder execution failed")
	}
}
