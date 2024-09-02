package sign

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

const (
	TypeNotaryBackend = "notary"
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
	// Set URL to Notary server signing endpoint
	Endpoint string `yaml:"endpoint" json:"endpoint"`
	// Time after connection to notary server should time out
	Timeout time.Duration `yaml:"timeout" json:"timeout"`
	// RetryTimeout is time between each signing request to notary in case something fails
	// Default is 10 seconds
	RetryTimeout time.Duration `yaml:"retry-timeout" json:"retry-timeout"`
	// Paths to the certificate and private key for mTLS
	CertFile string `yaml:"cert_file" json:"cert_file"`
	KeyFile  string `yaml:"key_file" json:"key_file"`
}

// AuthSecretConfig contains auth information for notary server
type AuthSecretConfig struct {
	// Path if path to file that contains secret credentials
	Path string `yaml:"path" json:"path"`
	// Type contains credential type, based on which the service will configure signing
	Type string `yaml:"type" json:"type"`
}

// SignifySecret contains configuration of secret that is used to connect to SAP signify service
type SignifySecret struct {
	// URL to the SAP signify endpoint
	Endpoint string `yaml:"endpoint" json:"endpoint"`
	// Actual secret data that will be pushed on sign request. Assumed its in JSON format
	Payload string `yaml:"payload" json:"payload"`
}

// SigningRequest contains information about all images with tags to sign using Notary
type SigningRequest struct {
	// Global unique name, e.g. full image name with registry URL
	NotaryGun string `json:"notaryGun"`
	// SHA sum of manifest.json
	SHA256 string `json:"sha256"`
	// size of manifest.json
	ByteSize int64 `json:"byteSize"`
	// Image tag
	Version string `json:"version"`
}

// AuthFunc is a small middleware function that intercepts http.Request with specific authentication method
type AuthFunc func(r *http.Request) *http.Request

// NotarySigner is a struct that implements sign.Signer interface
// Takes care of signing requests to Notary server
type NotarySigner struct {
	authFunc     AuthFunc
	c            http.Client
	url          string
	retryTimeout time.Duration
}

// AuthToken mutates created request, so it contains bearer token of authorized user
// Serves as middleware function before sending request
func AuthToken(token string) AuthFunc {
	return func(r *http.Request) *http.Request {
		r.Header.Add("Authorization", "Token "+token)
		return r
	}
}

// SignifyAuth fetches the JWT token from provided API endpoint in configuration file
// then returns a function that populates this JWT token as Bearer in the request
func SignifyAuth(s SignifySecret) (AuthFunc, error) {
	endpoint := s.Endpoint
	payload := s.Payload
	resp, err := http.Post(endpoint, "application/json", strings.NewReader(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var t struct {
		AccessToken struct {
			Token string `json:"token"`
		} `json:"access_token"`
	}

	rBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, ErrBadResponse{status: resp.Status, message: string(rBody)}
	}
	if err := json.Unmarshal(rBody, &t); err != nil {
		return nil, err
	}

	token := t.AccessToken.Token

	return func(r *http.Request) *http.Request {
		r.Header.Add("Authorization", "Bearer "+token)
		return r
	}, nil
}

func (ns NotarySigner) buildSigningRequest(images []string) ([]SigningRequest, error) {
	var sr []SigningRequest
	for _, i := range images {
		var base, tag string
		// Split on ":"
		parts := strings.Split(i, tagDelim)
		// Verify that we aren't confusing a tag for a hostname w/ port for the purposes of weak validation.
		if len(parts) > 1 && !strings.Contains(parts[len(parts)-1], regRepoDelimiter) {
			base = strings.Join(parts[:len(parts)-1], tagDelim)
			tag = parts[len(parts)-1]
		}
		ref, err := name.ParseReference(i)
		if err != nil {
			return nil, fmt.Errorf("ref parse: %w", err)
		}
		i, err := remote.Image(ref)
		if err != nil {
			return nil, fmt.Errorf("get image: %w", err)
		}
		m, err := i.Manifest()
		if err != nil {
			return nil, fmt.Errorf("image manifest: %w", err)
		}
		sha := m.Config.Digest.Hex
		size := m.Config.Size
		sr = append(sr, SigningRequest{NotaryGun: base, Version: tag, ByteSize: size, SHA256: sha})
	}
	return sr, nil
}

func (ns NotarySigner) Sign(images []string) error {
	sImg := strings.Join(images, ", ")
	sr, err := ns.buildSigningRequest(images)
	if err != nil {
		return fmt.Errorf("build sign request: %w", err)
	}
	b, err := json.Marshal(sr)
	if err != nil {
		return fmt.Errorf("marshal signing request: %w", err)
	}
	req, err := http.NewRequest("POST", ns.url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	retries := 5
	var status string
	var respMsg []byte
	w := time.NewTicker(ns.retryTimeout)
	defer w.Stop()
	for retries > 0 {
		fmt.Printf("Trying to sign %s. %v retries remaining...\n", sImg, retries)
		// wait for ticker to run
		<-w.C
		resp, err := ns.c.Do(req)
		if err != nil {
			return fmt.Errorf("notary request: %w", err)
		}
		status = resp.Status
		respMsg, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("body read: %w", err)
		}
		switch resp.StatusCode {
		case http.StatusOK:
			// response was fine. Do not need anything else
			fmt.Printf("Successfully signed images %s!\n", sImg)
			return nil
		case http.StatusUnauthorized, http.StatusForbidden, http.StatusBadRequest, http.StatusUnsupportedMediaType:
			return ErrBadResponse{status: status, message: string(respMsg)}
		}
		retries--
	}
	fmt.Println("Reached all retries. Stopping.")
	return ErrBadResponse{status: status, message: string(respMsg)}
}

// NewSigner initializes a NotarySigner with mTLS authentication
func (nc NotaryConfig) NewSigner() (Signer, error) {
	var ns NotarySigner

	// Load the client certificate and key for mTLS
	cert, err := tls.LoadX509KeyPair(nc.CertFile, nc.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("unable to load mTLS certificate and key: %w", err)
	}

	// Create a tls.Config with the certificate
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// Create an http.Transport that uses the tls.Config
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	// Create an HTTP client that uses the transport
	ns.c = http.Client{
		Transport: transport,
		Timeout:   nc.Timeout,
	}

	ns.retryTimeout = 10 * time.Second
	if nc.RetryTimeout > 0 {
		ns.retryTimeout = nc.RetryTimeout
	}
	ns.url = nc.Endpoint
	return ns, nil
}
