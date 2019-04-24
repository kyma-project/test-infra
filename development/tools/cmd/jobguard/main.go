package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kyma-project/test-infra/development/tools/pkg/jobguard"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type config struct {
	StatusFetcher jobguard.StatusFetcherConfig

	AuthorizationToken string `envconfig:"optional,GITHUB_TOKEN"`

	JobNamePattern string `envconfig:"default=components"`

	InitialSleepTime time.Duration `envconfig:"default=1m"`
	TickTime         time.Duration `envconfig:"default=15s"` // TODO rename
	Timeout          time.Duration `envconfig:"default=15m"`
}

func main() {
	log.Print("Starting Job Guard...")

	var cfg config
	err := envconfig.Init(&cfg)
	exitOnError(err, "while loading configuration")

	client := jobguard.HTTPClient(cfg.AuthorizationToken)
	statusFetcher := jobguard.NewStatusFetcher(cfg.StatusFetcher, client)

	log.Printf("Sleeping for %v...", cfg.InitialSleepTime)
	time.Sleep(cfg.InitialSleepTime)

	err = waitForDependentJobs(statusFetcher, cfg)
	exitOnError(err, "while waiting for success statuses")
}

func waitForDependentJobs(statusFetcher *jobguard.StatusFetcher, cfg config) error {
	byNames, err := jobguard.NameRegexpPredicate(cfg.JobNamePattern)
	if err != nil {
		return err
	}
	return jobguard.WaitAtMost(func() (bool, error) {
		statuses, err := statusFetcher.Do()
		if err != nil {
			return false, err
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
	}, cfg.TickTime, cfg.Timeout)
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
