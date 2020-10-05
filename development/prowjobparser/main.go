package main

import (
	"fmt"
	"os"

	"github.com/jamiealquiza/envy"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/test-infra/prow/config"
)

var (
	log     = logrus.New()
	rootCmd = &cobra.Command{
		Use:   "prowjobparser",
		Short: "prowjobparser print filtered list of prowjob names",
		Long:  "prowjobparser parse all prowjobs definitions from provided path and prints filtered results.",
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

// RunConfig holds cli params passed for execution.
type RunConfig struct {
	configPath    string
	jobPath       string
	includePreset []string
	excludePreset []string
	cluster       string
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

func matchCluster(jobBase config.JobBase) {
	if jobBase.Cluster == runCfg.cluster {
		fmt.Printf("%s\n", jobBase.Name)
	}
}

func findPresubmits(presubmits []config.Presubmit, filter func(base config.JobBase)) {
	for _, job := range presubmits {
		filter(job.JobBase)
	}
}

func findPostsubmits(postsubmits []config.Postsubmit, filter func(base config.JobBase)) {
	for _, job := range postsubmits {
		filter(job.JobBase)
	}
}

func findPeriodics(periodics []config.Periodic, filter func(base config.JobBase)) {
	for _, job := range periodics {
		filter(job.JobBase)
	}
}

func findProwjobs() error {
	var err error
	if prowCfg, err = parseProwjobs(); err != nil {
		return err
	}
	if len(runCfg.includePreset) > 0 || len(runCfg.excludePreset) > 0 {
		findPresubmits(prowCfg.JobConfig.AllStaticPresubmits([]string{}), matchLabels)
		findPostsubmits(prowCfg.JobConfig.AllStaticPostsubmits([]string{}), matchLabels)
		findPeriodics(prowCfg.JobConfig.AllPeriodics(), matchLabels)
	}
	if runCfg.cluster != "" {
		findPresubmits(prowCfg.JobConfig.AllStaticPresubmits([]string{}), matchCluster)
		findPostsubmits(prowCfg.JobConfig.AllStaticPostsubmits([]string{}), matchCluster)
		findPeriodics(prowCfg.JobConfig.AllPeriodics(), matchCluster)
	}

	return nil
}

func parseProwjobs() (*config.Config, error) {
	log.Info(fmt.Sprintf("Path to config file: %s", runCfg.configPath))
	log.Info(fmt.Sprintf("Path to jobs directory: %s", runCfg.jobPath))
	log.Info(fmt.Sprintf("Included presets: %v", runCfg.includePreset))
	log.Info(fmt.Sprintf("Excluded presets: %v", runCfg.excludePreset))
	log.Info(fmt.Sprintf("Cluster: %s", runCfg.cluster))
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
	rootCmd.PersistentFlags().StringArrayVarP(&runCfg.includePreset, "includepreset", "i", []string{}, "Prowjobs must contain this preset to be included in output.")
	rootCmd.PersistentFlags().StringArrayVarP(&runCfg.excludePreset, "excludepreset", "e", []string{}, "Prowjobs with this preset added will be excluded from output.")
	rootCmd.PersistentFlags().StringVarP(&runCfg.cluster, "cluster", "C", "", "Print prowjobs with cluster set to the flag value.")

	rootCmd.MarkPersistentFlagRequired("configpath")
	rootCmd.MarkPersistentFlagRequired("jobpath")
	envy.ParseCobra(rootCmd, envy.CobraConfig{Prefix: "PROWJOBPARSER", Persistent: true, Recursive: false})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
