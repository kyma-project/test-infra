package jobguard

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/pkg/errors"
)

// StatusState is state of a status (job)
type StatusState string

// StatusStateSuccess represents success state of a status
const StatusStateSuccess StatusState = "success"

// StatusStatePending represents pending state of a status
const StatusStatePending StatusState = "pending"

// StatusStateError represents error state of a status
const StatusStateError StatusState = "error"

// StatusStateFailure represents failure state of a status
const StatusStateFailure StatusState = "failure"

// Status stores essential data for a status
type Status struct {
	Name  string `json:"context"`
	State string `json:"state"`
}

// StatusConfig holds configuraiton for GithubStatusFetcher
type StatusConfig struct {
	Origin     string `envconfig:"default=https://api.github.com,API_ORIGIN"`
	Owner      string `envconfig:"default=kyma-project,REPO_OWNER"`
	Repository string `envconfig:"default=kyma,REPO_NAME"`
	CommitSHA  string `envconfig:"COMMIT_SHA"`
}

// GithubStatusFetcher fetches all statuses for the given commit
type GithubStatusFetcher struct {
	cfg    StatusConfig
	client *http.Client

}

// NewStatusFetcher constructs new GithubStatusFetcher instance
func NewStatusFetcher(cfg StatusConfig, client *http.Client) *GithubStatusFetcher {
	return &GithubStatusFetcher{cfg: cfg, client: client}
}


type statusResponse struct {
	TotalCount int      `json:"total_count"`
	Statuses   []Status `json:"statuses"`
}

func (f *GithubStatusFetcher) Do() ([]Status, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/commits/%s/status", f.cfg.Origin, f.cfg.Owner, f.cfg.Repository, f.cfg.CommitSHA)

	var statuses []Status

	pageNo := 1

	for {
		pageURL := fmt.Sprintf("%s?page=%d&per_page=100", url, pageNo)

		resp, err := f.client.Get(pageURL)
		if err != nil {
			return nil, errors.Wrapf(err, "while doing request to %s", url)
		}

		if resp.StatusCode != http.StatusOK {
			f.closeResponseBody(resp)
			return nil, fmt.Errorf("returned unexpected status code, expected: [%d], got: [%d]", http.StatusOK, resp.StatusCode)
		}

		var result statusResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		f.closeResponseBody(resp)
		if err != nil {
			return nil, errors.Wrapf(err, "while decoding response from request to %s", url)
		}

		if len(result.Statuses) == 0 {
			return nil, fmt.Errorf("error while paginating on pages. No statuses found on page %d", pageNo)
		}

		statuses = append(statuses, result.Statuses...)

		log.Printf("\tFetched statuses: %d/%d\n", len(statuses), result.TotalCount)
		if len(statuses) == result.TotalCount {
			break
		}
		pageNo++
	}

	return statuses, nil
}

func (f *GithubStatusFetcher) closeResponseBody(resp *http.Response) {
	_, _ = io.Copy(ioutil.Discard, resp.Body)
	err := resp.Body.Close()
	if err != nil {
		log.Println("\tGot error on closing response body:", err)
	}
}
