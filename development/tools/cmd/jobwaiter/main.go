package main

import (
	"github.com/kyma-project/test-infra/development/tools/pkg/jobwaiter"
	"log"
	"time"
)


type Config struct {
	Origin string
	Owner string
	Repository string
	PullRequestNumber string

	Timeout time.Duration
}


func main() {
	timeout := 10 * time.Minute
	statusFetcher := jobwaiter.NewStatusFetcher("https://api.github.com", "kyma-project", "kyma", 3792)

	err := jobwaiter.WaitAtMost(func() (bool, error) {

		statuses, err := statusFetcher.Do()
		if err != nil {
			return false, err
		}

		log.Printf("Statuses %+v\n", statuses)

		// TODO: filter statuses by name

		return true, nil

	}, timeout)

	if err != nil {
		log.Fatal(err.Error())
	}

}

