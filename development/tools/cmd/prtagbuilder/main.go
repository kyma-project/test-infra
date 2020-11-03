package main

import (
	"fmt"
	"github.com/jamiealquiza/envy"
	"github.com/kyma-project/test-infra/development/tools/pkg/prtagbuilder"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/test-infra/prow/apis/prowjobs/v1"
	"k8s.io/test-infra/prow/pod-utils/downwardapi"
	"os"
)

var (
	log     = logrus.New()
	rootCmd = &cobra.Command{
		Use:   "prtagbuilder",
		Short: "short description",
		Long:  "long description",
		Run: func(cmd *cobra.Command, args []string) {
			prtagbuilder.BuildPrTag(jobSpec, fromFlags)
		},
	}
	jobSpec   *downwardapi.JobSpec
	fromFlags bool = false
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)
	jobSpec = &downwardapi.JobSpec{Refs: &v1.Refs{}}
	rootCmd.PersistentFlags().StringVarP(&jobSpec.Refs.Org, "org", "o", "", "Github organisation which owns repo.")
	rootCmd.PersistentFlags().StringVarP(&jobSpec.Refs.Repo, "repo", "r", "", "Github repository.")
	rootCmd.PersistentFlags().StringVarP(&jobSpec.Refs.BaseRef, "baseref", "b", "", "Base branch name.")
	rootCmd.MarkPersistentFlagRequired("org")
	rootCmd.MarkPersistentFlagRequired("repo")
	rootCmd.MarkPersistentFlagRequired("baseref")
	envy.ParseCobra(rootCmd, envy.CobraConfig{Prefix: "TAGBUILDER", Persistent: true, Recursive: false})
	fmt.Println(pflag.NFlag())
	if pflag.NFlag() > 0 {
		fromFlags = true
	}
	fmt.Printf("%v", jobSpec)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
