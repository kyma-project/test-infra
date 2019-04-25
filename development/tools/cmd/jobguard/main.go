package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kyma-project/test-infra/development/tools/pkg/jobguard"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	prowCfg "k8s.io/test-infra/prow/config"
)

type config struct {
	InitialSleepTime time.Duration `envconfig:"default=1m"`
	RetryInterval    time.Duration `envconfig:"default=15s"`
	Timeout          time.Duration `envconfig:"default=15m"`

	Status             jobguard.StatusConfig
	AuthorizationToken string `envconfig:"optional,GITHUB_TOKEN"`
	JobNamePattern     string `envconfig:"default=components"`

	Prow             struct {
		ConfigFile    string
		JobsDirectory string
	}
}

func main() {
	log.Print("Starting Job Guard...")

	var cfg config
	err := envconfig.Init(&cfg)

	exitOnError(err, "while loading configuration")

	log.Printf("Sleeping for %v...", cfg.InitialSleepTime)
	time.Sleep(cfg.InitialSleepTime)

	repoName := fmt.Sprintf("%s/%s", cfg.Status.Owner, cfg.Status.Repository)
	jobsNumber, err := getNumberOfJobs(cfg.Prow.ConfigFile, cfg.Prow.JobsDirectory, repoName)
	exitOnError(err, "while calculating number of jobs")
	log.Printf("Expected number of jobs: %d", jobsNumber)

	client := jobguard.HTTPClient(cfg.AuthorizationToken)
	statusFetcher := jobguard.NewStatusFetcher(cfg.Status, client)

	err = waitForDependentJobs(statusFetcher, cfg, jobsNumber)
	exitOnError(err, "while waiting for success statuses")
}

func getNumberOfJobs(prowConfigFile, jobsDirectory, repoName string) (int, error) {
	c, err := prowCfg.Load(prowConfigFile, jobsDirectory)
	if err != nil {
		return 0, err
	}

	matchingJobs := 0
	for _, job := range c.Presubmits[repoName] {
		if job.AlwaysRun || job.RunIfChanged != "" {
			matchingJobs++
		}
	}
	return matchingJobs, nil
}

func waitForDependentJobs(statusFetcher *jobguard.GithubStatusFetcher, cfg config, totalNumberOfJobs int) error {
	byNames, err := jobguard.NameRegexpPredicate(cfg.JobNamePattern)
	if err != nil {
		return err
	}
	return jobguard.WaitAtMost(func() (bool, error) {
		statuses, err := statusFetcher.Do()
		if err != nil {
			return false, err
		}

		if len(statuses) != totalNumberOfJobs {
			log.Printf("Got %d statuses, expected %d", len(statuses), totalNumberOfJobs)
			return false, nil
		}
		filteredStatuses := jobguard.Filter(statuses, byNames)
		log.Printf("Got %d statuses, %d of them match name regexp\n", len(statuses), len(filteredStatuses))

		failedStatuses := jobguard.Filter(filteredStatuses, jobguard.FailedStatusPredicate)

		if len(failedStatuses) > 0 {
			log.Fatalf("[ERROR] At least one job with name matching pattern '%s' failed:\n%s", cfg.JobNamePattern, printJobNames(failedStatuses))
		}

		pendingStatuses := jobguard.Filter(filteredStatuses, jobguard.PendingStatusPredicate)
		pendingStatusesLen := len(pendingStatuses)

		if pendingStatusesLen > 0 {
			log.Printf("Waiting for jobs to finish:\n%s", printJobNames(pendingStatuses))
			return false, nil
		}

		log.Printf("[SUCCESS] All jobs with name matching pattern '%s' finished.", cfg.JobNamePattern)

		return true, nil
	}, cfg.RetryInterval, cfg.Timeout)
}

func printJobNames(in []jobguard.Status) string {
	var jobNames strings.Builder
	for _, status := range in {
		jobNames.WriteString(fmt.Sprintf("\t%s\n", status.Name))
	}

	return jobNames.String()
}

func exitOnError(err error, context string) {
	if err == nil {
		return
	}

	wrappedError := errors.Wrap(err, context)
	log.Fatal(wrappedError)
}
