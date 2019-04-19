package jobwaiter

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
)

type Status struct {
	Name  string `json:"context"`
	State string `json:"state"`
}

type StatusFetcher struct {
	origin            string
	owner             string
	repository        string
	pullRequestNumber int
}

func NewStatusFetcher(origin string, owner string, repository string, pullRequestNumber int) *StatusFetcher {
	return &StatusFetcher{origin: origin, owner: owner, repository: repository, pullRequestNumber: pullRequestNumber}
}

func (f *StatusFetcher) Do() ([]Status, error) {
	prDetails, err := f.pullRequestDetails()
	if err != nil {
		return nil, errors.Wrapf(err, "while getting pull request details for %d", f.pullRequestNumber)
	}

	statusURL, err := f.statusURL(prDetails)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting status URL for %d", f.pullRequestNumber)
	}

	statuses, err := f.fetchStatus(statusURL)
	if err != nil {
		return nil, errors.Wrapf(err, "while fetching status for %d", f.pullRequestNumber)
	}

	return statuses, nil
}

func (f *StatusFetcher) pullRequestDetails() (map[string]interface{}, error) {
	prDetailsURL := fmt.Sprintf("%s/repos/%s/%s/pulls/%d", f.origin, f.owner, f.repository, f.pullRequestNumber)

	resp, err := http.Get(prDetailsURL)
	if err != nil {
		return nil, errors.Wrapf(err, "while doing request to %s", prDetailsURL)
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, errors.Wrapf(err, "while decoding response from request to %s", prDetailsURL)
	}

	return result, nil
}

func (f *StatusFetcher) statusURL(prDetails map[string]interface{}) (string, error) {
	links, ok := prDetails["_links"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Incorrect type for links: %T", links)
	}

	statuses, ok := links["statuses"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Incorrect type for statuses: %T", links)
	}

	statusURL, ok := statuses["href"].(string)
	if !ok {
		return "", fmt.Errorf("Incorrect type for href: %T", links)
	}

	return statusURL, nil
}

func (f *StatusFetcher) fetchStatus(url string) ([]Status, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "while doing request to %s", url)
	}

	var result []Status
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, errors.Wrapf(err, "while decoding response from request to %s", url)
	}

	return result, nil
}
