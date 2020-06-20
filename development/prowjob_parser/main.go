package main

import (
	"fmt"
	"github.com/jamiealquiza/envy"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/test-infra/prow/config"
	"os"
)

var (
	log     = logrus.New()
	rootCmd = &cobra.Command{
		Use:   "prowjob-parser",
		Short: "prowjob-parser parse all prowjobs definitions from provided path and print complete definition",
		Long:  "prowjob-parser parse all prowjobs definitions from provided path and print complete definition. It support multiple filters to narrow down results.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := findProwjobs(); err != nil {
				log.Fatal(err)
			}
		},
	}
	runCfg    = &RunConfig{}
	prowCfg   = &config.Config{}
	includeok bool
	excludeok bool
)

type RunConfig struct {
	configPath    string
	jobPath       string
	includePreset []string
	excludePreset []string
}

func matchLabels(jobBase config.JobBase) {
	if len(runCfg.includePreset) > 0 {
		includeok = false
		for _, preset := range runCfg.includePreset {
			if _, includeok = jobBase.Labels[preset]; includeok == false {
				break
			}
		}
	} else {
		includeok = true
	}
	if len(runCfg.excludePreset) > 0 {
		excludeok = true
		for _, preset := range runCfg.excludePreset {
			if _, excludeok = jobBase.Labels[preset]; excludeok == true {
				break
			}
		}
	} else {
		excludeok = false
	}
	if includeok && !excludeok {
		fmt.Printf("%s\n", jobBase.Name)
	}
}

func findPresubmits(presubmits []config.Presubmit) {
	for _, job := range presubmits {
		matchLabels(job.JobBase)
	}
}

func findPostsubmits(postsubmits []config.Postsubmit) {
	for _, job := range postsubmits {
		matchLabels(job.JobBase)
	}
}

func findPeridics(periodics []config.Periodic) {
	for _, job := range periodics {
		matchLabels(job.JobBase)
	}
}

func findProwjobs() error {
	var err error
	if prowCfg, err = parseProwjobs(); err != nil {
		return err
	}
	findPresubmits(prowCfg.JobConfig.AllStaticPresubmits([]string{}))
	findPostsubmits(prowCfg.JobConfig.AllStaticPostsubmits([]string{}))
	findPeridics(prowCfg.JobConfig.AllPeriodics())

	return nil
}

func parseProwjobs() (*config.Config, error) {
	log.Info(fmt.Sprintf("Path to config file: %s", runCfg.configPath))
	log.Info(fmt.Sprintf("Path to jobs directory: %s", runCfg.jobPath))
	log.Info(fmt.Sprintf("Included presets: %v", runCfg.includePreset))
	log.Info(fmt.Sprintf("Excluded presets: %v", runCfg.excludePreset))
	var err error
	if prowCfg, err = config.Load(runCfg.configPath, runCfg.jobPath); err != nil {
		return nil, err
	}
	return prowCfg, nil
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)

	rootCmd.PersistentFlags().StringVarP(&runCfg.configPath, "configpath", "c", "", "Path to yaml file containing prow configuration.")
	rootCmd.PersistentFlags().StringVarP(&runCfg.jobPath, "jobpath", "j", "", "Path to directory containing yaml files with prowjobs definitions.")
	rootCmd.PersistentFlags().StringArrayVarP(&runCfg.includePreset, "includepreset", "i", []string{}, "Prowjobs with this preset added will be included in output.")
	rootCmd.PersistentFlags().StringArrayVarP(&runCfg.excludePreset, "excludepreset", "e", []string{}, "Prowjobs with this preset added will be excluded from output.")

	rootCmd.MarkPersistentFlagRequired("configpath")
	rootCmd.MarkPersistentFlagRequired("jobpath")
	envy.ParseCobra(rootCmd, envy.CobraConfig{Prefix: "PROWJOBPARSER", Persistent: true, Recursive: false})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
