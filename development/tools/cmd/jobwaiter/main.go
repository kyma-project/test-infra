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
	Timeout time.Duration `envconfig:"default=10m"`
}

func main() {
	log.Print("Starting job waiter...")

	var cfg Config
	err := envconfig.Init(&cfg)
	exitOnError(err, "while loading configuration")

	client := jobwaiter.HTTPClient(cfg.AuthorizationToken)

	statusFetcher := jobwaiter.NewStatusFetcher(cfg.StatusFetcher, client)

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

		if len(jobwaiter.FailedStatuses(filteredStatuses)) > 0 {
			log.Fatalf("At least one job with substring %s failed. Exiting with error...", cfg.JobFilterSubstring)
		}

		pendingStatuses := jobwaiter.PendingStatuses(filteredStatuses)
		pendingStatusesLen := len(pendingStatuses)

		if  pendingStatusesLen > 0 {
			var jobNames strings.Builder
			for _, pendingStatus := range pendingStatuses {
				jobNames.WriteString(fmt.Sprintf("\t%s\n", pendingStatus.Name))
			}

			log.Printf("Waiting for jobs to finish:\n%s", jobNames.String())
			return false, nil
		}

		log.Printf("All jobs with substring %s finished.", cfg.JobFilterSubstring)

		return true, nil
	}, cfg.Timeout)
}

func exitOnError(err error, context string) {
	if err == nil {
		return
	}

	wrappedError := errors.Wrap(err, context)
	log.Fatal(wrappedError)
}
