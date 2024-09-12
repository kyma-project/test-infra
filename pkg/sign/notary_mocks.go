package sign

import (
	"net/http"
)

// MockRoundTripper allows us to mock HTTP client behavior for tests.
type MockRoundTripper struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripFunc(req)
}

type MockReference struct{}
type MockImage struct{}

func (m *MockImage) Manifest() (*Manifest, error) {
	return &Manifest{
		Config: struct {
			Digest struct {
				Hex string
			}
			Size int64
		}{
			Digest: struct {
				Hex string
			}{
				Hex: "abc123def456",
			},
			Size: 12345678,
		},
	}, nil
}

func mockParseReference(image string) (Reference, error) {
	return &MockReference{}, nil
}

func mockGetImage(ref Reference) (Image, error) {
	return &MockImage{}, nil
}
