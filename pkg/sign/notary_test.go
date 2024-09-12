package sign

import (
	"bytes"
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
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sign Package Test Suite")
}

func setupDecodeCertAndKeyTests() {
	Describe("DecodeCertAndKey", func() {
		var signifySecret SignifySecret

		BeforeEach(func() {
			// Use the GenerateBase64EncodedCert function to generate base64-encoded cert and key
			certBase64, keyBase64, err := GenerateBase64EncodedCert()
			Expect(err).To(BeNil())

			signifySecret = SignifySecret{
				CertificateData: certBase64,
				PrivateKeyData:  keyBase64,
			}
		})

		Context("When decoding is successful", func() {
			It("should decode certificate and private key successfully", func() {
				cert, err := signifySecret.DecodeCertAndKey()
				Expect(err).To(BeNil())
				Expect(cert).To(BeAssignableToTypeOf(tls.Certificate{}))
			})
		})

		Context("When certificate decoding fails", func() {
			BeforeEach(func() {
				signifySecret.CertificateData = "invalid-base64"
			})

			It("should return an error for invalid certificate data", func() {
				_, err := signifySecret.DecodeCertAndKey()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to decode certificate"))
			})
		})

		Context("When private key decoding fails", func() {
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
			// Set up NotarySigner with mock functions
			ns = NotarySigner{
				ParseReferenceFunc: mockParseReference,
				GetImageFunc:       mockGetImage,
			}
		})

		Describe("buildSigningRequest", func() {
			Context("When valid images are provided", func() {
				It("should create signing requests successfully", func() {
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

			Context("When fetching the image fails", func() {
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
			It("should create a payload with the correct trusted collections", func() {
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

				// Verify the TrustedCollections in the payload
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

		Context("When an empty signing request list is provided", func() {
			It("should return an empty payload", func() {
				signingRequests := []SigningRequest{}

				payload, err := ns.buildPayload(signingRequests)
				Expect(err).To(BeNil())

				// Verify that the payload has no trusted collections
				Expect(payload.TrustedCollections).To(BeEmpty())
			})
		})
	})
}

func setupSignTests() {
	Describe("Sign", func() {
		var ns NotarySigner
		var mockTransport *MockRoundTripper
		var fakeCert tls.Certificate

		BeforeEach(func() {
			// Create mock HTTP transport
			mockTransport = &MockRoundTripper{}

			// Mock certificate and key
			fakeCert = tls.Certificate{}

			// Initialize NotarySigner with mocks
			ns = NotarySigner{
				signifySecret: SignifySecret{},
				url:           "https://mock.notarysigner.com/sign",
				retryTimeout:  time.Second, // For fast retry testing
				BuildSigningReqFunc: func(images []string) ([]SigningRequest, error) {
					return []SigningRequest{
						{
							NotaryGun: "gcr.io/project",
							SHA256:    "abc123",
							ByteSize:  123456,
							Version:   "v1.0",
						},
					}, nil
				},
				BuildPayloadFunc: func(sr []SigningRequest) (SigningPayload, error) {
					return SigningPayload{
						TrustedCollections: []TrustedCollection{
							{
								GUN: "gcr.io/project",
								Targets: []Target{
									{
										Name:     "v1.0",
										ByteSize: 123456,
										Digest:   "abc123",
									},
								},
							},
						},
					}, nil
				},
				DecodeCertFunc: func() (tls.Certificate, error) {
					return fakeCert, nil
				},
				c: &http.Client{
					Transport: mockTransport,
					Timeout:   time.Second,
				},
			}
		})

		Context("When signing succeeds on the first attempt", func() {
			It("should sign images successfully", func() {
				// Mock successful HTTP response
				mockTransport.RoundTripFunc = func(req *http.Request) (*http.Response, error) {
					Expect(req.URL.String()).To(Equal("https://mock.notarysigner.com/sign"))

					return &http.Response{
						StatusCode: http.StatusAccepted,
						Body:       io.NopCloser(bytes.NewReader([]byte(`Success`))),
					}, nil
				}

				// Call the Sign function
				err := ns.Sign([]string{"gcr.io/project/image:v1.0"})
				Expect(err).To(BeNil(), "Expected signing to succeed on the first attempt")
			})
		})

		Context("When signing fails after retries", func() {
			It("should return an error after retrying", func() {
				// Mock HTTP request/response to simulate a failure
				mockTransport.RoundTripFunc = func(req *http.Request) (*http.Response, error) {
					Expect(req.URL.String()).To(Equal("https://mock.notarysigner.com/sign"))

					return &http.Response{
						StatusCode: http.StatusBadRequest,
						Body:       io.NopCloser(bytes.NewReader([]byte(`Bad Request`))),
					}, nil
				}

				// Call the Sign function
				err := ns.Sign([]string{"gcr.io/project/image:v1.0"})
				Expect(err).To(HaveOccurred(), "Expected signing to fail after retries")
				Expect(err.Error()).To(ContainSubstring("failed to sign images"))
			})
		})
	})
}

func setupNewSignerTests() {
	Describe("NewSigner", func() {
		var nc NotaryConfig

		BeforeEach(func() {
			// Initialize NotaryConfig with mocked Secret
			nc = NotaryConfig{
				Timeout:      5 * time.Second,
				RetryTimeout: 15 * time.Second,
				Secret: &AuthSecretConfig{ // Properly initialize Secret
					Path: "/mock/path/to/secret",
					Type: "signify",
				},
			}
		})

		Context("When valid signify secret is provided", func() {
			It("should return a valid NotarySigner", func() {
				// Mock signify secret content
				signifySecret := SignifySecret{
					CertificateData: "mockCertData",
					PrivateKeyData:  "mockPrivateKeyData",
				}
				secretContent, _ := json.Marshal(signifySecret)

				// Inject mock ReadFileFunc to simulate file reading without a real file
				nc.ReadFileFunc = func(path string) ([]byte, error) {
					// Ensure the correct path is used
					Expect(path).To(Equal("/mock/path/to/secret"))
					// Return the mock signify secret content
					return secretContent, nil
				}

				// Call the NewSigner method
				signer, err := nc.NewSigner()
				Expect(err).To(BeNil())
				Expect(signer).NotTo(BeNil())

				// Verify NotarySigner properties
				notarySigner, ok := signer.(*NotarySigner)
				Expect(ok).To(BeTrue())
				Expect(notarySigner.signifySecret.CertificateData).To(Equal("mockCertData"))
				Expect(notarySigner.signifySecret.PrivateKeyData).To(Equal("mockPrivateKeyData"))
				Expect(notarySigner.retryTimeout).To(Equal(15 * time.Second))
				Expect(notarySigner.c.Timeout).To(Equal(5 * time.Second))
			})
		})

		Context("When secret file cannot be read", func() {
			It("should return an error", func() {
				// Mock an error when trying to read the secret file
				nc.Secret.Path = "/mock/invalid/path"
				nc.ReadFileFunc = func(path string) ([]byte, error) {
					Expect(path).To(Equal("/mock/invalid/path"))
					return nil, errors.New("failed to read file")
				}

				// Call the NewSigner method
				signer, err := nc.NewSigner()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to read secret file"))
				Expect(signer).To(BeNil())
			})
		})

		Context("When unsupported secret type is provided", func() {
			It("should return an error", func() {
				// Set an unsupported secret type, but don't provide a file path
				nc.Secret = &AuthSecretConfig{
					Type: "unsupported",
				}

				// Call the NewSigner method
				signer, err := nc.NewSigner()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unsupported secret type"))
				Expect(signer).To(BeNil())
			})
		})

		Context("When signify secret unmarshalling fails", func() {
			It("should return an error", func() {
				// Mock signify secret file with invalid JSON
				nc.Secret.Path = "/mock/path/to/secret"
				nc.Secret.Type = "signify"
				nc.ReadFileFunc = func(path string) ([]byte, error) {
					Expect(path).To(Equal("/mock/path/to/secret"))
					return []byte("invalid-json"), nil
				}

				// Call the NewSigner method
				signer, err := nc.NewSigner()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to unmarshal signify secret"))
				Expect(signer).To(BeNil())
			})
		})
	})
}

func init() {
	setupDecodeCertAndKeyTests()
	setupBuildSigningRequestTests()
	setupBuildPayloadTests()
	setupSignTests()
	setupNewSignerTests()
}

// GenerateBase64EncodedCert generates a self-signed certificate and private key,
// and returns them as base64 encoded strings.
func GenerateBase64EncodedCert() (string, string, error) {
	// Generate a private RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	// Create a self-signed certificate template
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

	// Create the certificate using the template and the private key
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "", "", err
	}

	// Encode the certificate to PEM format
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Encode the private key to PEM format
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Encode PEM data to base64
	certBase64 := base64.StdEncoding.EncodeToString(certPEM)
	keyBase64 := base64.StdEncoding.EncodeToString(keyPEM)

	return certBase64, keyBase64, nil
}
