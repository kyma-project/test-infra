package sign

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

const (
	TypeNotaryBackend = "notary"
	tagDelim          = ":"
	regRepoDelimiter  = "/"
)

type ErrBadResponse struct {
	status  string
	message string
}

type ErrAuthServiceNotSupported struct {
	Service string
}

func (e ErrAuthServiceNotSupported) Error() string {
	return fmt.Sprintf("'%s' auth service not supported", e.Service)
}

func (e ErrBadResponse) Error() string {
	return fmt.Sprintf("bad response from service: %s, %s", e.status, e.message)
}

type NotaryConfig struct {
	Endpoint     string            `yaml:"endpoint" json:"endpoint"`
	Secret       *AuthSecretConfig `yaml:"secret,omitempty" json:"secret,omitempty"`
	Timeout      time.Duration     `yaml:"timeout" json:"timeout"`
	RetryTimeout time.Duration     `yaml:"retry-timeout" json:"retry-timeout"`
	ReadFileFunc func(string) ([]byte, error)
}

type AuthSecretConfig struct {
	Path string `yaml:"path" json:"path"`
	Type string `yaml:"type" json:"type"`
}

type SignifySecret struct {
	CertificateData string `json:"certData"`
	PrivateKeyData  string `json:"privateKeyData"`
}

type SigningRequest struct {
	NotaryGun string `json:"notaryGun"`
	SHA256    string `json:"sha256"`
	ByteSize  int64  `json:"byteSize"`
	Version   string `json:"version"`
}

type Target struct {
	Name     string `json:"name"`
	ByteSize int64  `json:"byteSize"`
	Digest   string `json:"digest"`
}

type TrustedCollection struct {
	GUN     string   `json:"gun"`
	Targets []Target `json:"targets"`
}

type SigningPayload struct {
	TrustedCollections []TrustedCollection `json:"trustedCollections"`
}

type NotarySigner struct {
	c                   *http.Client
	url                 string
	retryTimeout        time.Duration
	signifySecret       SignifySecret
	ParseReferenceFunc  func(image string) (Reference, error)
	GetImageFunc        func(ref Reference) (Image, error)
	DecodeCertFunc      func() (tls.Certificate, error)
	BuildSigningReqFunc func([]string) ([]SigningRequest, error)
	BuildPayloadFunc    func([]SigningRequest) (SigningPayload, error)
	SetupTLSFunc        func(cert tls.Certificate) *tls.Config
	HTTPClient          *http.Client
}

type Reference interface{}
type Image interface {
	Manifest() (*Manifest, error)
}

type Manifest struct {
	Config struct {
		Digest struct {
			Hex string
		}
		Size int64
	}
}

type SimpleImage struct {
	ManifestData Manifest
}

func (si *SimpleImage) Manifest() (*Manifest, error) {
	return &si.ManifestData, nil
}

func GetImage(ref Reference) (Image, error) {
	r, ok := ref.(name.Reference)
	if !ok {
		return nil, fmt.Errorf("invalid reference type")
	}

	img, err := remote.Image(r)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image: %w", err)
	}

	manifest, err := img.Manifest()
	if err != nil {
		return nil, fmt.Errorf("failed to get manifest: %w", err)
	}

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
					Hex: manifest.Config.Digest.Hex,
				},
				Size: manifest.Config.Size,
			},
		},
	}, nil
}

type SimpleReference struct {
	Image string
	Tag   string
}

func ParseReference(image string) (Reference, error) {
	ref, err := name.ParseReference(image)
	if err != nil {
		return nil, fmt.Errorf("failed to parse image reference: %w", err)
	}

	return ref, nil
}

func (ss *SignifySecret) DecodeCertAndKey() (tls.Certificate, error) {
	certData, err := base64.StdEncoding.DecodeString(ss.CertificateData)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to decode certificate: %w", err)
	}

	keyData, err := base64.StdEncoding.DecodeString(ss.PrivateKeyData)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to decode private key: %w", err)
	}

	cert, err := tls.X509KeyPair(certData, keyData)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("unable to load certificate or key: %w", err)
	}

	return cert, nil
}

func (ns NotarySigner) buildSigningRequest(images []string) ([]SigningRequest, error) {
	var signingRequests []SigningRequest

	for _, i := range images {
		var base, tag string

		parts := strings.Split(i, tagDelim)
		if len(parts) > 1 && !strings.Contains(parts[len(parts)-1], regRepoDelimiter) {
			base = strings.Join(parts[:len(parts)-1], tagDelim)
			tag = parts[len(parts)-1]
		}

		ref, err := ns.ParseReferenceFunc(i)
		if err != nil {
			return nil, fmt.Errorf("ref parse: %w", err)
		}
		img, err := ns.GetImageFunc(ref)
		if err != nil {
			return nil, fmt.Errorf("get image: %w", err)
		}
		manifest, err := img.Manifest()
		if err != nil {
			return nil, fmt.Errorf("failed getting image manifest: %w", err)
		}

		signingRequests = append(signingRequests, SigningRequest{
			NotaryGun: base,
			SHA256:    manifest.Config.Digest.Hex,
			ByteSize:  manifest.Config.Size,
			Version:   tag,
		})
	}

	return signingRequests, nil
}

func (ns NotarySigner) buildPayload(sr []SigningRequest) (SigningPayload, error) {
	var trustedCollections []TrustedCollection

	for _, req := range sr {
		target := Target{
			Name:     req.Version,
			ByteSize: req.ByteSize,
			Digest:   req.SHA256,
		}

		trustedCollection := TrustedCollection{
			GUN:     req.NotaryGun,
			Targets: []Target{target},
		}

		trustedCollections = append(trustedCollections, trustedCollection)
	}

	payload := SigningPayload{
		TrustedCollections: trustedCollections,
	}

	return payload, nil
}

func (ns NotarySigner) Sign(images []string) error {
	sImg := strings.Join(images, ", ")

	signingRequests, err := ns.buildSigningRequest(images)
	if err != nil {
		return fmt.Errorf("build signing request: %w", err)
	}

	payload, err := ns.BuildPayloadFunc(signingRequests)
	if err != nil {
		return fmt.Errorf("build payload: %w", err)
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal signing request: %w", err)
	}

	var client *http.Client
	if ns.HTTPClient != nil {
		client = ns.HTTPClient
	} else {
		cert, err := ns.DecodeCertFunc()
		if err != nil {
			return fmt.Errorf("failed to load certificate and key: %w", err)
		}

		tlsConfig := ns.SetupTLSFunc(cert)
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
			Timeout: ns.retryTimeout,
		}
	}

	req, err := http.NewRequest("POST", ns.url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := retryHTTPRequest(client, req, 5, ns.retryTimeout)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		respMsg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to sign images: %w", ErrBadResponse{status: resp.Status, message: string(respMsg)})
	}

	fmt.Printf("Successfully signed images %s!\n", sImg)
	return nil
}

func retryHTTPRequest(client *http.Client, req *http.Request, retries int, retryInterval time.Duration) (*http.Response, error) {
	var lastResp *http.Response
	var lastErr error

	// Read and store the request body
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		req.Body.Close()
	}

	for retries > 0 {
		// Reset the request body for each retry
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
		} else if resp.StatusCode == http.StatusAccepted {
			return resp, nil
		} else {
			lastErr = fmt.Errorf("failed to sign images, unexpected status code: %d", resp.StatusCode)
		}

		lastResp = resp
		retries--
		if retries == 0 {
			break
		}
		time.Sleep(retryInterval)
	}
	return lastResp, fmt.Errorf("request failed after retries: %w", lastErr)
}

func setupTLS(cert tls.Certificate) *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
}

func (nc NotaryConfig) NewSigner() (Signer, error) {
	ns := NotarySigner{
		retryTimeout:       nc.RetryTimeout,
		url:                nc.Endpoint,
		ParseReferenceFunc: ParseReference,
		GetImageFunc:       GetImage,
		SetupTLSFunc:       setupTLS,
	}

	ns.url = "https://signing-manage.repositories.cloud.sap/trusted-collections/publish"
	ns.BuildPayloadFunc = ns.buildPayload
	ns.DecodeCertFunc = ns.signifySecret.DecodeCertAndKey

	// HTTP client configuration
	ns.c = &http.Client{
		Timeout: nc.Timeout,
	}

	if nc.Secret != nil {
		switch nc.Secret.Type {
		case "signify":
			readFileFunc := nc.ReadFileFunc
			if readFileFunc == nil {
				readFileFunc = os.ReadFile
			}

			f, err := readFileFunc(nc.Secret.Path)
			if err != nil {
				return nil, fmt.Errorf("failed to read secret file: %w", err)
			}

			var s SignifySecret
			err = json.Unmarshal(f, &s)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal signify secret: %w", err)
			}

			ns.signifySecret = s
		default:
			return nil, fmt.Errorf("unsupported secret type: %s", nc.Secret.Type)
		}
	}

	return &ns, nil
}
