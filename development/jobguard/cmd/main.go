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

	Prow struct {
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
	expSuccessJobs, err := getExpSuccessJobs(cfg.Prow.ConfigFile, cfg.Prow.JobsDirectory, repoName, cfg.JobNamePattern)
	exitOnError(err, "while calculating number of jobs")
	log.Printf("Expected jobs: %v", expSuccessJobs)

	client := jobguard.HTTPClient(cfg.AuthorizationToken)
	statusFetcher := jobguard.NewStatusFetcher(cfg.Status, client)

	err = waitForDependentJobs(statusFetcher, cfg, expSuccessJobs)
	exitOnError(err, "while waiting for success statuses")
}

func getExpSuccessJobs(prowConfigFile, jobsDirectory, repoName, pattern string) ([]string, error) {
	c, err := prowCfg.Load(prowConfigFile, jobsDirectory)
	if err != nil {
		return nil, err
	}

	var jobs []jobguard.Status
	// copied from prow/plugins/trigger/pull-request.go buildAll() method
	for _, job := range c.JobConfig.PresubmitsStatic[repoName] {
		if job.SkipReport { // if skipped then it will not be reported on GitHub
			continue
		}

		if job.AlwaysRun || job.RunIfChanged != "" {
			jobs = append(jobs, jobguard.Status{Name: job.Name})
		}
	}

	byNames, err := jobguard.NameRegexpPredicate(pattern)
	if err != nil {
		return nil, err
	}
	requiredJobs := jobguard.Filter(jobs, byNames)

	var requiredJobsNames []string
	for _, j := range requiredJobs {
		requiredJobsNames = append(requiredJobsNames, j.Name)
	}

	return requiredJobsNames, nil
}

func waitForDependentJobs(statusFetcher *jobguard.GithubStatusFetcher, cfg config, expSuccessJobs []string) error {
	return jobguard.WaitAtMost(func() (bool, error) {
		ghJobs, err := statusFetcher.Do()
		if err != nil {
			return false, err
		}

		notYetReported, pending, failed := classifyJobs(ghJobs, expSuccessJobs)

		switch {
		case len(failed) > 0:
			log.Fatalf("[ERROR] At least one job with name matching pattern '%s' failed: [%s]", cfg.JobNamePattern, printJobNames(failed))
		case len(notYetReported) > 0:
			log.Printf("Waiting for jobs [%s] to report their status", printJobNames(notYetReported))
			return false, nil
		case len(pending) > 0:
			log.Printf("Waiting for jobs to finish: [%s]", printJobNames(pending))
			return false, nil
		}

		log.Printf("[SUCCESS] All jobs with name matching pattern '%s' finished.", cfg.JobNamePattern)

		return true, nil
	}, cfg.RetryInterval, cfg.Timeout)
}

func classifyJobs(gotJobs jobguard.IndexedStatuses, expSuccessJobs []string) (notYetReported []string, pending []string, failed []string) {
	for _, name := range expSuccessJobs {
		status, found := gotJobs[name]
		switch {
		case !found:
			notYetReported = append(notYetReported, name)
		case jobguard.IsFailedStatus(status):
			failed = append(failed, name)
		case jobguard.IsPendingStatus(status):
			pending = append(pending, name)
		}
	}

	return notYetReported, pending, failed
}

func printJobNames(in []string) string {
	return strings.Join(in, ",")
}

func exitOnError(err error, context string) {
	if err == nil {
		return
	}
	wrappedError := errors.Wrap(err, context)
	log.Fatal(wrappedError)
}
