package jobwaiter

import (
	"fmt"
	"net/http"
)

func HTTPClient(token string) *http.Client {
	return &http.Client{
		Transport: newGhRoundTripper(token),
	}
}

type ghRoundTripper struct {
	rt    http.RoundTripper
	token string
}

func newGhRoundTripper(token string) *ghRoundTripper {
	return &ghRoundTripper{
		token: token,
		rt:    http.DefaultTransport,
	}
}

func (t *ghRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Accept", "application/vnd.github.v3+json")

	if t.token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", t.token))
	}

	return t.rt.RoundTrip(req)
}
