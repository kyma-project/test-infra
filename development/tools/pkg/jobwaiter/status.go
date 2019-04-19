package jobwaiter

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
)

type StatusState string

const StatusStateSuccess StatusState = "success"
const StatusStatePending StatusState = "pending"
const StatusStateError StatusState = "error"
const StatusStateFailure StatusState = "failure"

type Status struct {
	Name  string `json:"context"`
	State string `json:"state"`
}

type StatusFetcherConfig struct {
	Origin            string `envconfig:"default=https://api.github.com"`
	Owner             string `envconfig:"default=kyma-project,REPO_OWNER"`
	Repository        string `envconfig:"default=kyma,REPO_NAME"`
	PullRequestNumber int    `envconfig:"PULL_NUMBER"`
}

type StatusFetcher struct {
	cfg    StatusFetcherConfig
	client *http.Client

	commitSHA string
}

func NewStatusFetcher(cfg StatusFetcherConfig, client *http.Client) *StatusFetcher {
	return &StatusFetcher{cfg: cfg, client: client}
}

// TODO: Do we really need this? What about PULL_PULL_SHA env?
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
	Statuses []Status `json:"statuses"`
}

func (f *StatusFetcher) fetchStatuses(commmitSHA string) ([]Status, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/commits/%s/status", f.cfg.Origin, f.cfg.Owner, f.cfg.Repository, commmitSHA)

	// TODO: Paginate!!!

	resp, err := f.client.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "while doing request to %s", url)
	}

	var result statusResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, errors.Wrapf(err, "while decoding response from request to %s", url)
	}
	defer closeResponseBody(resp)

	return result.Statuses, nil
}

func closeResponseBody(resp *http.Response) {
	_, _ = io.Copy(ioutil.Discard, resp.Body)
	_ = resp.Body.Close()
}
