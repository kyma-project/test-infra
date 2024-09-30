package sign

import (
	"crypto/tls"
	"fmt"
	"net/http"
)

// MockImageRepository implements ImageRepositoryInterface
type MockImageRepository struct {
	MockParseReference func(image string) (ReferenceInterface, error)
	MockGetImage       func(ref ReferenceInterface) (ImageInterface, error)
}

func (mir *MockImageRepository) ParseReference(image string) (ReferenceInterface, error) {
	if mir.MockParseReference != nil {
		return mir.MockParseReference(image)
	}
	return nil, fmt.Errorf("MockParseReference not implemented")
}

func (mir *MockImageRepository) GetImage(ref ReferenceInterface) (ImageInterface, error) {
	if mir.MockGetImage != nil {
		return mir.MockGetImage(ref)
	}
	return nil, fmt.Errorf("MockGetImage not implemented")
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

// MockReference implements ReferenceInterface
type MockReference struct {
	MockName              func() string
	MockString            func() string
	MockGetRepositoryName func() string
	MockGetTag            func() (string, error)
}

func (mr *MockReference) Name() string {
	if mr.MockName != nil {
		return mr.MockName()
	}
	return ""
}

func (mr *MockReference) String() string {
	if mr.MockString != nil {
		return mr.MockString()
	}
	return ""
}

func (mr *MockReference) GetRepositoryName() string {
	if mr.MockGetRepositoryName != nil {
		return mr.MockGetRepositoryName()
	}
	return ""
}

func (mr *MockReference) GetTag() (string, error) {
	if mr.MockGetTag != nil {
		return mr.MockGetTag()
	}
	return "", fmt.Errorf("MockGetTag not implemented")
}

// MockImage implements ImageInterface
type MockImage struct {
	MockManifest func() (ManifestInterface, error)
}

func (mi *MockImage) Manifest() (ManifestInterface, error) {
	if mi.MockManifest != nil {
		return mi.MockManifest()
	}
	return nil, fmt.Errorf("MockManifest not implemented")
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
