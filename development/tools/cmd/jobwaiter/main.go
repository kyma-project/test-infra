package main

import (
	"fmt"
	"github.com/kyma-project/test-infra/development/tools/pkg/jobwaiter"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"log"
	"strings"
	"time"
)

type Config struct {
	StatusFetcher jobwaiter.StatusFetcherConfig

	AuthorizationToken string `envconfig:"optional,BOT_GITHUB_TOKEN"`

	JobFilterSubstring string `envconfig:"default=components"`

	InitialSleepTime time.Duration `envconfig:"default=1m"`
	TickTime         time.Duration `envconfig:"default=15s"`
	Timeout          time.Duration `envconfig:"default=15m"`
}

func main() {
	log.Print("Starting Job Waiter...")

	var cfg Config
	err := envconfig.Init(&cfg)
	exitOnError(err, "while loading configuration")

	client := jobwaiter.HTTPClient(cfg.AuthorizationToken)
	statusFetcher := jobwaiter.NewStatusFetcher(cfg.StatusFetcher, client)

	log.Printf("Sleeping for %.f minute(s)...", cfg.InitialSleepTime.Minutes())
	time.Sleep(cfg.InitialSleepTime)

	log.Print("Initializing...")
	err = statusFetcher.Init()
	exitOnError(err, "while initialization")

	err = waitForDependentJobs(statusFetcher, cfg)
	exitOnError(err, "while waiting for success statuses")
}

func waitForDependentJobs(statusFetcher *jobwaiter.StatusFetcher, cfg Config) error {
	return jobwaiter.WaitAtMost(func() (bool, error) {
		statuses, err := statusFetcher.Do()
		if err != nil {
			return false, err
		}

		filteredStatuses := jobwaiter.FilterStatusByName(statuses, cfg.JobFilterSubstring)

		failedStatuses := jobwaiter.FailedStatuses(filteredStatuses)

		if len(failedStatuses) > 0 {
			log.Fatalf("[ERROR] At least one job with substring '%s' failed:\n%s", cfg.JobFilterSubstring, printJobNames(failedStatuses))
		}

		pendingStatuses := jobwaiter.PendingStatuses(filteredStatuses)
		pendingStatusesLen := len(pendingStatuses)

		if pendingStatusesLen > 0 {
			log.Printf("Waiting for jobs to finish:\n%s", printJobNames(pendingStatuses))
			return false, nil
		}

		log.Printf("[SUCCESS] All jobs with substring '%s' finished.", cfg.JobFilterSubstring)

		return true, nil
	}, cfg.TickTime, cfg.Timeout)
}

func printJobNames(in []jobwaiter.Status) string {
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
