package jobwaiter

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
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

// StatusFetcherConfig holds configuraiton for StatusFetcher
type StatusFetcherConfig struct {
	Origin            string `envconfig:"default=https://api.github.com,API_ORIGIN"`
	Owner             string `envconfig:"default=kyma-project,REPO_OWNER"`
	Repository        string `envconfig:"default=kyma,REPO_NAME"`
	PullRequestNumber int    `envconfig:"PULL_NUMBER"`
}

// StatusFetcher fetches all statuses for a pull request
type StatusFetcher struct {
	cfg    StatusFetcherConfig
	client *http.Client

	commitSHA string
}

// NewStatusFetcher constructs new StatusFetcher instance
func NewStatusFetcher(cfg StatusFetcherConfig, client *http.Client) *StatusFetcher {
	return &StatusFetcher{cfg: cfg, client: client}
}

// TODO: Do we really need this? What about PULL_PULL_SHA env?
// Init fetches pull request details and gathers essential pieces of information
func (f *StatusFetcher) Init() error {
	prDetails, err := f.pullRequestDetails()
	if err != nil {
		return errors.Wrapf(err, "while getting pull request details for %d", f.cfg.PullRequestNumber)
	}

	commitSHA, err := f.pullRequestHeadSHA(prDetails)
	if err != nil {
		return errors.Wrapf(err, "while getting status URL for %d", f.cfg.PullRequestNumber)
	}

	f.commitSHA = commitSHA
	return nil
}

// Do fetches statuses for a pull request
func (f *StatusFetcher) Do() ([]Status, error) {
	if f.commitSHA == "" {
		return nil, errors.New("Commit SHA not fetched")
	}

	statuses, err := f.fetchStatuses(f.commitSHA)
	if err != nil {
		return nil, errors.Wrapf(err, "while fetching status for %d", f.cfg.PullRequestNumber)
	}

	return statuses, nil
}

func (f *StatusFetcher) pullRequestDetails() (map[string]interface{}, error) {
	prDetailsURL := fmt.Sprintf("%s/repos/%s/%s/pulls/%d", f.cfg.Origin, f.cfg.Owner, f.cfg.Repository, f.cfg.PullRequestNumber)

	resp, err := f.client.Get(prDetailsURL)
	if err != nil {
		return nil, errors.Wrapf(err, "while doing request to %s", prDetailsURL)
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, errors.Wrapf(err, "while decoding response from request to %s", prDetailsURL)
	}
	defer closeResponseBody(resp)

	return result, nil
}

func (f *StatusFetcher) pullRequestHeadSHA(prDetails map[string]interface{}) (string, error) {
	head, ok := prDetails["head"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Incorrect type for head: %T", prDetails["head"])
	}

	sha, ok := head["sha"].(string)
	if !ok {
		return "", fmt.Errorf("Incorrect type for sha: %T", head["sha"])
	}

	return sha, nil
}

type statusResponse struct {
	TotalCount int      `json:"total_count"`
	Statuses   []Status `json:"statuses"`
}

func (f *StatusFetcher) fetchStatuses(commmitSHA string) ([]Status, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/commits/%s/status", f.cfg.Origin, f.cfg.Owner, f.cfg.Repository, commmitSHA)

	var statuses []Status

	pageNo := 1

	for {
		log.Printf("\tPaginating through statuses... Page number: %d\n", pageNo)
		pageURL := fmt.Sprintf("%s?page=%d&per_page=100", url, pageNo)

		resp, err := f.client.Get(pageURL)
		if err != nil {
			return nil, errors.Wrapf(err, "while doing request to %s", url)
		}

		var result statusResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			return nil, errors.Wrapf(err, "while decoding response from request to %s", url)
		}

		if len(result.Statuses) == 0 {
			return nil, fmt.Errorf("Error while paginating on pages. No statuses found on page %d", pageNo)
		}

		statuses = append(statuses, result.Statuses...)

		closeResponseBody(resp)

		log.Printf("\tFetched statuses: %d/%d\n", len(statuses), result.TotalCount)
		if len(statuses) == result.TotalCount {
			break
		}

		pageNo++
	}

	return statuses, nil
}

func closeResponseBody(resp *http.Response) {
	_, _ = io.Copy(ioutil.Discard, resp.Body)
	_ = resp.Body.Close()
}
