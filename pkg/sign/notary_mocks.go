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

// MockParseReference jest funkcją mockującą dla ParseReferenceFunc
func MockParseReference(image string) (Reference, error) {
	return image, nil // W prostym przypadku zwracamy sam string jako Reference
}

// MockGetImage jest funkcją mockującą dla GetImageFunc
func MockGetImage(ref Reference) (Image, error) {
	// Zwracamy mockowany obiekt Image z predefiniowanymi wartościami
	return &SimpleImage{
		ManifestData: Manifest{
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
		},
	}, nil
}
