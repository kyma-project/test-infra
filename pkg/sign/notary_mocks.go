package sign

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/google/go-containerregistry/pkg/name"
)

// MockImageRepository implements ImageRepositoryInterface
type MockImageRepository struct {
	MockParseReference  func(image string) (name.Reference, error)
	MockGetImage        func(ref name.Reference) (ImageInterface, error)
	MockIsManifestList  func(ref name.Reference) (bool, error)
	MockGetManifestList func(ref name.Reference) (ManifestListInterface, error)
}

func (mir *MockImageRepository) ParseReference(image string) (name.Reference, error) {
	if mir.MockParseReference != nil {
		return mir.MockParseReference(image)
	}
	return nil, fmt.Errorf("MockParseReference not implemented")
}

func (mir *MockImageRepository) GetImage(ref name.Reference) (ImageInterface, error) {
	if mir.MockGetImage != nil {
		return mir.MockGetImage(ref)
	}
	return nil, fmt.Errorf("MockGetImage not implemented")
}

func (mir *MockImageRepository) IsManifestList(ref name.Reference) (bool, error) {
	if mir.MockIsManifestList != nil {
		return mir.MockIsManifestList(ref)
	}
	return false, fmt.Errorf("MockIsManifestList not implemented")
}

func (mir *MockImageRepository) GetManifestList(ref name.Reference) (ManifestListInterface, error) {
	if mir.MockGetManifestList != nil {
		return mir.MockGetManifestList(ref)
	}
	return nil, fmt.Errorf("MockGetManifestList not implemented")
}

type MockManifestList struct {
	MockGetDigest func() (string, error)
	MockGetSize   func() (int64, error)
}

func (mml *MockManifestList) GetDigest() (string, error) {
	if mml.MockGetDigest != nil {
		return mml.MockGetDigest()
	}
	return "", fmt.Errorf("MockGetDigest not implemented")
}

func (mml *MockManifestList) GetSize() (int64, error) {
	if mml.MockGetSize != nil {
		return mml.MockGetSize()
	}
	return 0, fmt.Errorf("MockGetSize not implemented")
}

// MockPayloadBuilder implements PayloadBuilderInterface
type MockPayloadBuilder struct {
	MockBuildPayload func(images []string) (SigningPayload, error)
}

func (m *MockPayloadBuilder) BuildPayload(images []string) (SigningPayload, error) {
	if m.MockBuildPayload != nil {
		return m.MockBuildPayload(images)
	}
	return SigningPayload{}, fmt.Errorf("MockBuildPayload not implemented")
}

// MockTLSProvider implements TLSProviderInterface
type MockTLSProvider struct {
	MockGetTLSConfig func() (*tls.Config, error)
}

func (mtp *MockTLSProvider) GetTLSConfig() (*tls.Config, error) {
	if mtp.MockGetTLSConfig != nil {
		return mtp.MockGetTLSConfig()
	}
	return nil, fmt.Errorf("MockGetTLSConfig not implemented")
}

// MockHTTPClient implements HTTPClientInterface
type MockHTTPClient struct {
	MockDo           func(req *http.Request) (*http.Response, error)
	MockSetTLSConfig func(tlsConfig *tls.Config) error
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.MockDo != nil {
		return m.MockDo(req)
	}
	return nil, fmt.Errorf("MockDo not implemented")
}

func (m *MockHTTPClient) SetTLSConfig(tlsConfig *tls.Config) error {
	if m.MockSetTLSConfig != nil {
		return m.MockSetTLSConfig(tlsConfig)
	}
	return fmt.Errorf("MockSetTLSConfig not implemented")
}

// MockImage implements ImageInterface
type MockImage struct {
	MockGetDigest func() (string, error)
	MockGetSize   func() (int64, error)
}

func (mi *MockImage) GetDigest() (string, error) {
	if mi.MockGetDigest != nil {
		return mi.MockGetDigest()
	}
	return "", fmt.Errorf("MockGetDigest not implemented")
}

func (mi *MockImage) GetSize() (int64, error) {
	if mi.MockGetSize != nil {
		return mi.MockGetSize()
	}
	return 0, fmt.Errorf("MockGetDigest not implemented")
}

// MockManifest implements ManifestInterface
type MockManifest struct {
	MockGetConfigSize   func() int64
	MockGetConfigDigest func() string
}

func (mm *MockManifest) GetConfigSize() int64 {
	if mm.MockGetConfigSize != nil {
		return mm.MockGetConfigSize()
	}
	return 0
}

func (mm *MockManifest) GetConfigDigest() string {
	if mm.MockGetConfigDigest != nil {
		return mm.MockGetConfigDigest()
	}
	return ""
}
