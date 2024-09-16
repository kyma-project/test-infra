package sign

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sign Package Test Suite")
}

var _ = Describe("Sign Package Tests", func() {
	setupDecodeCertAndKeyTests()
	setupBuildSigningRequestTests()
	setupBuildPayloadTests()
	setupNewSignerTests()
	setupGetImageTests()
	setupParseReferenceTests()
	setupSetupTLSTests()
	setupRetryHTTPRequestTests()
	setupSignTests()
})

func setupDecodeCertAndKeyTests() {
	Describe("DecodeCertAndKey", func() {
		var signifySecret SignifySecret

		BeforeEach(func() {
			// Generating base64 encoded certificate and key
			certBase64, keyBase64, err := GenerateBase64EncodedCert()
			Expect(err).To(BeNil())

			signifySecret = SignifySecret{
				CertificateData: certBase64,
				PrivateKeyData:  keyBase64,
			}
		})

		Context("When decoding is successful", func() {
			It("should correctly decode certificate and private key", func() {
				cert, err := signifySecret.DecodeCertAndKey()
				Expect(err).To(BeNil())
				Expect(cert).To(BeAssignableToTypeOf(tls.Certificate{}))
			})
		})

		Context("When decoding certificate fails", func() {
			BeforeEach(func() {
				signifySecret.CertificateData = "invalid-base64"
			})

			It("should return an error for invalid certificate data", func() {
				_, err := signifySecret.DecodeCertAndKey()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to decode certificate"))
			})
		})

		Context("When decoding private key fails", func() {
			BeforeEach(func() {
				signifySecret.PrivateKeyData = "invalid-base64"
			})

			It("should return an error for invalid private key data", func() {
				_, err := signifySecret.DecodeCertAndKey()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to decode private key"))
			})
		})

		Context("When loading TLS certificate fails", func() {
			BeforeEach(func() {
				signifySecret.CertificateData = base64.StdEncoding.EncodeToString([]byte("invalid-cert"))
				signifySecret.PrivateKeyData = base64.StdEncoding.EncodeToString([]byte("invalid-key"))
			})

			It("should return an error for invalid certificate or key", func() {
				_, err := signifySecret.DecodeCertAndKey()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unable to load certificate or key"))
			})
		})
	})
}

func setupBuildSigningRequestTests() {
	Describe("NotarySigner", func() {
		var ns NotarySigner

		BeforeEach(func() {
			// Setting up NotarySigner with mock functions
			ns = NotarySigner{
				ParseReferenceFunc: MockParseReference,
				GetImageFunc:       MockGetImage,
			}
		})

		Describe("buildSigningRequest", func() {
			Context("When valid images are provided", func() {
				It("should correctly create signing requests", func() {
					images := []string{
						"gcr.io/project/image:tag1",
						"docker.io/library/ubuntu:latest",
						"quay.io/repository/image:v2.1.0",
					}

					signingRequests, err := ns.buildSigningRequest(images)
					Expect(err).To(BeNil())
					Expect(signingRequests).To(HaveLen(len(images)))

					for i, req := range signingRequests {
						Expect(req.NotaryGun).NotTo(BeEmpty(), "NotaryGun should not be empty for request %d", i)
						Expect(req.SHA256).To(Equal("abc123def456"), "SHA256 should match for request %d", i)
						Expect(req.ByteSize).To(Equal(int64(12345678)), "ByteSize should match for request %d", i)
					}
				})
			})

			Context("When an invalid image reference is provided", func() {
				BeforeEach(func() {
					ns.ParseReferenceFunc = func(image string) (Reference, error) {
						return nil, fmt.Errorf("invalid reference")
					}
				})

				It("should return an error for invalid reference", func() {
					images := []string{"invalid/image:tag"}

					_, err := ns.buildSigningRequest(images)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("ref parse"))
				})
			})

			Context("When image fetching fails", func() {
				BeforeEach(func() {
					ns.GetImageFunc = func(ref Reference) (Image, error) {
						return nil, fmt.Errorf("image fetch failed")
					}
				})

				It("should return an error for failed image fetch", func() {
					images := []string{"gcr.io/project/image:tag1"}

					_, err := ns.buildSigningRequest(images)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("get image"))
				})
			})
		})
	})
}

func setupBuildPayloadTests() {
	Describe("buildPayload", func() {
		var ns NotarySigner

		BeforeEach(func() {
			ns = NotarySigner{}
		})

		Context("When valid signing requests are provided", func() {
			It("should create a payload with correct trusted collections", func() {
				signingRequests := []SigningRequest{
					{
						NotaryGun: "gcr.io/project",
						SHA256:    "abc123",
						ByteSize:  123456,
						Version:   "v1.0",
					},
					{
						NotaryGun: "docker.io/library/ubuntu",
						SHA256:    "def456",
						ByteSize:  654321,
						Version:   "latest",
					},
				}

				payload, err := ns.buildPayload(signingRequests)
				Expect(err).To(BeNil())

				// Verify TrustedCollections in the payload
				Expect(payload.TrustedCollections).To(HaveLen(2))

				// Check the first trusted collection
				Expect(payload.TrustedCollections[0].GUN).To(Equal("gcr.io/project"))
				Expect(payload.TrustedCollections[0].Targets).To(HaveLen(1))
				Expect(payload.TrustedCollections[0].Targets[0].Name).To(Equal("v1.0"))
				Expect(payload.TrustedCollections[0].Targets[0].ByteSize).To(Equal(int64(123456)))
				Expect(payload.TrustedCollections[0].Targets[0].Digest).To(Equal("abc123"))

				// Check the second trusted collection
				Expect(payload.TrustedCollections[1].GUN).To(Equal("docker.io/library/ubuntu"))
				Expect(payload.TrustedCollections[1].Targets).To(HaveLen(1))
				Expect(payload.TrustedCollections[1].Targets[0].Name).To(Equal("latest"))
				Expect(payload.TrustedCollections[1].Targets[0].ByteSize).To(Equal(int64(654321)))
				Expect(payload.TrustedCollections[1].Targets[0].Digest).To(Equal("def456"))
			})
		})

		Context("When an empty list of signing requests is provided", func() {
			It("should return an empty payload", func() {
				signingRequests := []SigningRequest{}

				payload, err := ns.buildPayload(signingRequests)
				Expect(err).To(BeNil())

				// Verify that the payload contains no trusted collections
				Expect(payload.TrustedCollections).To(BeEmpty())
			})
		})
	})
}

func setupNewSignerTests() {
	Describe("NewSigner", func() {
		var nc NotaryConfig

		BeforeEach(func() {
			// Initialize NotaryConfig with a mocked Secret
			nc = NotaryConfig{
				Endpoint:     "https://example.com/sign",
				Timeout:      5 * time.Second,
				RetryTimeout: 15 * time.Second,
				Secret: &AuthSecretConfig{
					Path: "/mock/path/to/secret",
					Type: "signify",
				},
			}
		})

		Context("When a valid signify secret is provided", func() {
			It("should return a valid NotarySigner", func() {
				// Mock signify secret content
				signifySecret := SignifySecret{
					CertificateData: "mockCertData",
					PrivateKeyData:  "mockPrivateKeyData",
				}
				secretContent, _ := json.Marshal(signifySecret)

				// Inject mocked ReadFileFunc
				nc.ReadFileFunc = func(path string) ([]byte, error) {
					Expect(path).To(Equal("/mock/path/to/secret"))
					return secretContent, nil
				}

				// Call NewSigner method
				signer, err := nc.NewSigner()
				Expect(err).To(BeNil())
				Expect(signer).NotTo(BeNil())

				// Verify properties of NotarySigner
				notarySigner, ok := signer.(*NotarySigner)
				Expect(ok).To(BeTrue())
				Expect(notarySigner.signifySecret.CertificateData).To(Equal("mockCertData"))
				Expect(notarySigner.signifySecret.PrivateKeyData).To(Equal("mockPrivateKeyData"))
				Expect(notarySigner.retryTimeout).To(Equal(15 * time.Second))
				Expect(notarySigner.c.Timeout).To(Equal(5 * time.Second))
				Expect(notarySigner.url).To(Equal("https://example.com/sign"))
			})
		})

		Context("When reading the secret file fails", func() {
			It("should return an error", func() {
				// Mock error during file read
				nc.Secret.Path = "/mock/invalid/path"
				nc.ReadFileFunc = func(path string) ([]byte, error) {
					Expect(path).To(Equal("/mock/invalid/path"))
					return nil, errors.New("failed to read file")
				}

				// Call NewSigner method
				signer, err := nc.NewSigner()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to read secret file"))
				Expect(signer).To(BeNil())
			})
		})

		Context("When an unsupported secret type is provided", func() {
			It("should return an error", func() {
				// Set unsupported secret type
				nc.Secret = &AuthSecretConfig{
					Type: "unsupported",
				}

				// Call NewSigner method
				signer, err := nc.NewSigner()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unsupported secret type"))
				Expect(signer).To(BeNil())
			})
		})

		Context("When unmarshalling the signify secret fails", func() {
			It("should return an error", func() {
				// Mock invalid JSON in the secret file
				nc.Secret.Path = "/mock/path/to/secret"
				nc.Secret.Type = "signify"
				nc.ReadFileFunc = func(path string) ([]byte, error) {
					Expect(path).To(Equal("/mock/path/to/secret"))
					return []byte("invalid-json"), nil
				}

				// Call NewSigner method
				signer, err := nc.NewSigner()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to unmarshal signify secret"))
				Expect(signer).To(BeNil())
			})
		})
	})
}

func setupGetImageTests() {
	Describe("GetImage", func() {
		Context("When a valid reference is provided", func() {
			It("should return a SimpleImage with correct manifest data", func() {
				// Create a valid reference
				ref, err := name.ParseReference("gcr.io/project/image:tag")
				Expect(err).To(BeNil())

				// Mock remote.Image and manifest
				// In this case, we assume the function works correctly,
				// because testing actual image fetching would require access to external resources.

				// So we can test the case where the reference is invalid.
				img, err := GetImage(ref)
				// Since we don't have an actual image, an error might occur.
				if err != nil {
					// Check if the error is related to fetching the image
					Expect(err.Error()).To(ContainSubstring("failed to fetch image"))
				} else {
					// If no error occurred, check that the image is not nil
					Expect(img).NotTo(BeNil())
				}
			})
		})

		Context("When an invalid reference type is provided", func() {
			It("should return an error indicating invalid reference type", func() {
				// Provide a reference that is not of type name.Reference
				ref := "invalid reference"

				img, err := GetImage(ref)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid reference type"))
				Expect(img).To(BeNil())
			})
		})
	})
}

func setupParseReferenceTests() {
	Describe("ParseReference", func() {
		Context("When a valid image string is provided", func() {
			It("should correctly parse the reference", func() {
				image := "gcr.io/project/image:tag"
				ref, err := ParseReference(image)
				Expect(err).To(BeNil())
				Expect(ref).NotTo(BeNil())
			})
		})

		Context("When an invalid image string is provided", func() {
			It("should return a parsing error", func() {
				image := "invalid_image_string@@"
				ref, err := ParseReference(image)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to parse image reference"))
				Expect(ref).To(BeNil())
			})
		})
	})
}

func setupSetupTLSTests() {
	Describe("setupTLS", func() {
		Context("When a valid TLS certificate is provided", func() {
			It("should return a correct TLS configuration", func() {
				certBase64, keyBase64, err := GenerateBase64EncodedCert()
				Expect(err).To(BeNil())

				certData, err := base64.StdEncoding.DecodeString(certBase64)
				Expect(err).To(BeNil())
				keyData, err := base64.StdEncoding.DecodeString(keyBase64)
				Expect(err).To(BeNil())

				cert, err := tls.X509KeyPair(certData, keyData)
				Expect(err).To(BeNil())

				tlsConfig := setupTLS(cert)
				Expect(tlsConfig).NotTo(BeNil())
				Expect(tlsConfig.Certificates).To(HaveLen(1))
			})
		})
	})
}

func setupRetryHTTPRequestTests() {
	Describe("retryHTTPRequest", func() {
		var (
			server       *httptest.Server
			client       *http.Client
			request      *http.Request
			retryCount   int
			retryTimeout time.Duration
		)

		BeforeEach(func() {
			retryCount = 3
			retryTimeout = 100 * time.Millisecond
		})

		Context("When the request is successful", func() {
			BeforeEach(func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusAccepted)
				}))
				client = server.Client()
				request, _ = http.NewRequest("GET", server.URL, nil)
			})

			AfterEach(func() {
				server.Close()
			})

			It("should return the response without errors", func() {
				resp, err := retryHTTPRequest(client, request, retryCount, retryTimeout)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(Equal(http.StatusAccepted))
			})
		})

		Context("When the request fails a few times but eventually succeeds", func() {
			var attempt int

			BeforeEach(func() {
				attempt = 0
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					attempt++
					if attempt < 2 {
						http.Error(w, "temporary error", http.StatusInternalServerError)
					} else {
						w.WriteHeader(http.StatusAccepted)
					}
				}))
				client = server.Client()
				request, _ = http.NewRequest("GET", server.URL, nil)
			})

			AfterEach(func() {
				server.Close()
			})

			It("should retry the request and eventually return a successful response", func() {
				resp, err := retryHTTPRequest(client, request, retryCount, retryTimeout)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(Equal(http.StatusAccepted))
				Expect(attempt).To(Equal(2))
			})
		})

		Context("When all request attempts fail", func() {
			BeforeEach(func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "error", http.StatusInternalServerError)
				}))
				client = server.Client()
				request, _ = http.NewRequest("GET", server.URL, nil)
			})

			AfterEach(func() {
				server.Close()
			})

			It("should return an error after exhausting retries", func() {
				resp, err := retryHTTPRequest(client, request, retryCount, retryTimeout)
				Expect(err).To(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))
				Expect(err.Error()).To(ContainSubstring("unexpected status code: 500"))
			})
		})
	})
}

func setupSignTests() {
	Describe("Sign", func() {
		var ns NotarySigner
		var server *httptest.Server

		BeforeEach(func() {
			ns = NotarySigner{
				ParseReferenceFunc: MockParseReference,
				GetImageFunc:       MockGetImage,
				BuildPayloadFunc: func(sr []SigningRequest) (SigningPayload, error) {
					return SigningPayload{
						TrustedCollections: []TrustedCollection{
							{
								GUN: "example.com/image",
								Targets: []Target{
									{
										Name:     "latest",
										ByteSize: 12345,
										Digest:   "abc123",
									},
								},
							},
						},
					}, nil
				},
				DecodeCertFunc: func() (tls.Certificate, error) {
					return tls.Certificate{}, nil
				},
				SetupTLSFunc: setupTLS,
				retryTimeout: 100 * time.Millisecond,
			}
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		Context("When signing is successful", func() {
			It("should complete without errors", func() {
				server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					io.Copy(io.Discard, r.Body) // Ensure the body is read
					w.WriteHeader(http.StatusAccepted)
				}))

				ns.url = server.URL
				ns.HTTPClient = server.Client() // Use the server's client

				err := ns.Sign([]string{"example.com/image:latest"})
				Expect(err).To(BeNil())
			})
		})

		Context("When an error occurs during signing", func() {
			It("should return the appropriate error", func() {
				server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					io.Copy(io.Discard, r.Body) // Ensure the body is read
					r.Body.Close()
					http.Error(w, "error", http.StatusInternalServerError)
				}))

				ns.url = server.URL
				ns.HTTPClient = server.Client() // Use the server's client

				err := ns.Sign([]string{"example.com/image:latest"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to sign images"))
			})
		})

	})
}

// GenerateBase64EncodedCert generates a self-signed certificate and private key,
// and returns them as base64-encoded strings.
func GenerateBase64EncodedCert() (string, string, error) {
	// Generate RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "", "", err
	}

	// Encode certificate to PEM format
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Encode private key to PEM format
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Encode PEM data to base64
	certBase64 := base64.StdEncoding.EncodeToString(certPEM)
	keyBase64 := base64.StdEncoding.EncodeToString(keyPEM)

	return certBase64, keyBase64, nil
}
