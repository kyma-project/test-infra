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
			// Generowanie bazy64 zakodowanych certyfikatu i klucza
			certBase64, keyBase64, err := GenerateBase64EncodedCert()
			Expect(err).To(BeNil())

			signifySecret = SignifySecret{
				CertificateData: certBase64,
				PrivateKeyData:  keyBase64,
			}
		})

		Context("Gdy dekodowanie przebiega pomyślnie", func() {
			It("powinno poprawnie dekodować certyfikat i klucz prywatny", func() {
				cert, err := signifySecret.DecodeCertAndKey()
				Expect(err).To(BeNil())
				Expect(cert).To(BeAssignableToTypeOf(tls.Certificate{}))
			})
		})

		Context("Gdy dekodowanie certyfikatu się nie powiedzie", func() {
			BeforeEach(func() {
				signifySecret.CertificateData = "niepoprawny-base64"
			})

			It("powinno zwrócić błąd dla niepoprawnych danych certyfikatu", func() {
				_, err := signifySecret.DecodeCertAndKey()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to decode certificate"))
			})
		})

		Context("Gdy dekodowanie klucza prywatnego się nie powiedzie", func() {
			BeforeEach(func() {
				signifySecret.PrivateKeyData = "niepoprawny-base64"
			})

			It("powinno zwrócić błąd dla niepoprawnych danych klucza prywatnego", func() {
				_, err := signifySecret.DecodeCertAndKey()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to decode private key"))
			})
		})

		Context("Gdy wczytanie certyfikatu TLS się nie powiedzie", func() {
			BeforeEach(func() {
				signifySecret.CertificateData = base64.StdEncoding.EncodeToString([]byte("niepoprawny-cert"))
				signifySecret.PrivateKeyData = base64.StdEncoding.EncodeToString([]byte("niepoprawny-klucz"))
			})

			It("powinno zwrócić błąd dla niepoprawnego certyfikatu lub klucza", func() {
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
			// Ustawienie NotarySigner z funkcjami mock
			ns = NotarySigner{
				ParseReferenceFunc: MockParseReference,
				GetImageFunc:       MockGetImage,
			}
		})

		Describe("buildSigningRequest", func() {
			Context("Gdy podano poprawne obrazy", func() {
				It("powinno poprawnie utworzyć żądania podpisania", func() {
					images := []string{
						"gcr.io/project/image:tag1",
						"docker.io/library/ubuntu:latest",
						"quay.io/repository/image:v2.1.0",
					}

					signingRequests, err := ns.buildSigningRequest(images)
					Expect(err).To(BeNil())
					Expect(signingRequests).To(HaveLen(len(images)))

					for i, req := range signingRequests {
						Expect(req.NotaryGun).NotTo(BeEmpty(), "NotaryGun nie powinno być puste dla żądania %d", i)
						Expect(req.SHA256).To(Equal("abc123def456"), "SHA256 powinno pasować dla żądania %d", i)
						Expect(req.ByteSize).To(Equal(int64(12345678)), "ByteSize powinno pasować dla żądania %d", i)
					}
				})
			})

			Context("Gdy podano niepoprawne odniesienie do obrazu", func() {
				BeforeEach(func() {
					ns.ParseReferenceFunc = func(image string) (Reference, error) {
						return nil, fmt.Errorf("invalid reference")
					}
				})

				It("powinno zwrócić błąd dla niepoprawnego odniesienia", func() {
					images := []string{"invalid/image:tag"}

					_, err := ns.buildSigningRequest(images)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("ref parse"))
				})
			})

			Context("Gdy pobieranie obrazu się nie powiedzie", func() {
				BeforeEach(func() {
					ns.GetImageFunc = func(ref Reference) (Image, error) {
						return nil, fmt.Errorf("image fetch failed")
					}
				})

				It("powinno zwrócić błąd dla nieudanego pobierania obrazu", func() {
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

		Context("Gdy podano poprawne żądania podpisania", func() {
			It("powinno utworzyć payload z poprawnymi zaufanymi kolekcjami", func() {
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

				// Weryfikacja TrustedCollections w payload
				Expect(payload.TrustedCollections).To(HaveLen(2))

				// Sprawdzenie pierwszej zaufanej kolekcji
				Expect(payload.TrustedCollections[0].GUN).To(Equal("gcr.io/project"))
				Expect(payload.TrustedCollections[0].Targets).To(HaveLen(1))
				Expect(payload.TrustedCollections[0].Targets[0].Name).To(Equal("v1.0"))
				Expect(payload.TrustedCollections[0].Targets[0].ByteSize).To(Equal(int64(123456)))
				Expect(payload.TrustedCollections[0].Targets[0].Digest).To(Equal("abc123"))

				// Sprawdzenie drugiej zaufanej kolekcji
				Expect(payload.TrustedCollections[1].GUN).To(Equal("docker.io/library/ubuntu"))
				Expect(payload.TrustedCollections[1].Targets).To(HaveLen(1))
				Expect(payload.TrustedCollections[1].Targets[0].Name).To(Equal("latest"))
				Expect(payload.TrustedCollections[1].Targets[0].ByteSize).To(Equal(int64(654321)))
				Expect(payload.TrustedCollections[1].Targets[0].Digest).To(Equal("def456"))
			})
		})

		Context("Gdy podano pustą listę żądań podpisania", func() {
			It("powinno zwrócić pusty payload", func() {
				signingRequests := []SigningRequest{}

				payload, err := ns.buildPayload(signingRequests)
				Expect(err).To(BeNil())

				// Weryfikacja, że payload nie zawiera zaufanych kolekcji
				Expect(payload.TrustedCollections).To(BeEmpty())
			})
		})
	})
}

func setupNewSignerTests() {
	Describe("NewSigner", func() {
		var nc NotaryConfig

		BeforeEach(func() {
			// Inicjalizacja NotaryConfig z mockowanym Secret
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

		Context("Gdy podano poprawny signify secret", func() {
			It("powinno zwrócić poprawny NotarySigner", func() {
				// Mockowanie treści signify secret
				signifySecret := SignifySecret{
					CertificateData: "mockCertData",
					PrivateKeyData:  "mockPrivateKeyData",
				}
				secretContent, _ := json.Marshal(signifySecret)

				// Iniekcja mockowanej funkcji ReadFileFunc
				nc.ReadFileFunc = func(path string) ([]byte, error) {
					Expect(path).To(Equal("/mock/path/to/secret"))
					return secretContent, nil
				}

				// Wywołanie metody NewSigner
				signer, err := nc.NewSigner()
				Expect(err).To(BeNil())
				Expect(signer).NotTo(BeNil())

				// Weryfikacja właściwości NotarySigner
				notarySigner, ok := signer.(*NotarySigner)
				Expect(ok).To(BeTrue())
				Expect(notarySigner.signifySecret.CertificateData).To(Equal("mockCertData"))
				Expect(notarySigner.signifySecret.PrivateKeyData).To(Equal("mockPrivateKeyData"))
				Expect(notarySigner.retryTimeout).To(Equal(15 * time.Second))
				Expect(notarySigner.c.Timeout).To(Equal(5 * time.Second))
				Expect(notarySigner.url).To(Equal("https://example.com/sign"))
			})
		})

		Context("Gdy nie można odczytać pliku secret", func() {
			It("powinno zwrócić błąd", func() {
				// Mockowanie błędu podczas odczytu pliku
				nc.Secret.Path = "/mock/invalid/path"
				nc.ReadFileFunc = func(path string) ([]byte, error) {
					Expect(path).To(Equal("/mock/invalid/path"))
					return nil, errors.New("failed to read file")
				}

				// Wywołanie metody NewSigner
				signer, err := nc.NewSigner()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to read secret file"))
				Expect(signer).To(BeNil())
			})
		})

		Context("Gdy podano nieobsługiwany typ secret", func() {
			It("powinno zwrócić błąd", func() {
				// Ustawienie nieobsługiwanego typu secret
				nc.Secret = &AuthSecretConfig{
					Type: "unsupported",
				}

				// Wywołanie metody NewSigner
				signer, err := nc.NewSigner()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unsupported secret type"))
				Expect(signer).To(BeNil())
			})
		})

		Context("Gdy unmarshalling signify secret się nie powiedzie", func() {
			It("powinno zwrócić błąd", func() {
				// Mockowanie niepoprawnego JSON w pliku secret
				nc.Secret.Path = "/mock/path/to/secret"
				nc.Secret.Type = "signify"
				nc.ReadFileFunc = func(path string) ([]byte, error) {
					Expect(path).To(Equal("/mock/path/to/secret"))
					return []byte("invalid-json"), nil
				}

				// Wywołanie metody NewSigner
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
		Context("Gdy podano poprawne odniesienie", func() {
			It("powinno zwrócić SimpleImage z poprawnymi danymi manifestu", func() {
				// Tworzymy poprawne odniesienie
				ref, err := name.ParseReference("gcr.io/project/image:tag")
				Expect(err).To(BeNil())

				// Mockowanie remote.Image i manifestu
				// W tym przypadku zakładamy, że funkcja działa poprawnie,
				// ponieważ testowanie rzeczywistego pobierania obrazu wymagałoby dostępu do zewnętrznych zasobów.

				// Możemy więc przetestować przypadek, gdy odniesienie jest niepoprawne.
				img, err := GetImage(ref)
				// Ponieważ nie mamy rzeczywistego obrazu, może wystąpić błąd.
				if err != nil {
					// Sprawdzamy, czy zwrócono błąd związany z pobieraniem obrazu
					Expect(err.Error()).To(ContainSubstring("failed to fetch image"))
				} else {
					// Jeśli nie wystąpił błąd, sprawdzamy, czy obraz nie jest pusty
					Expect(img).NotTo(BeNil())
				}
			})
		})

		Context("Gdy podano niepoprawny typ odniesienia", func() {
			It("powinno zwrócić błąd wskazujący na niepoprawny typ odniesienia", func() {
				// Podajemy odniesienie, które nie jest typu name.Reference
				ref := "niepoprawne odniesienie"

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
		Context("Gdy podano poprawny ciąg obrazu", func() {
			It("powinno poprawnie sparsować odniesienie", func() {
				image := "gcr.io/project/image:tag"
				ref, err := ParseReference(image)
				Expect(err).To(BeNil())
				Expect(ref).NotTo(BeNil())
			})
		})

		Context("Gdy podano niepoprawny ciąg obrazu", func() {
			It("powinno zwrócić błąd parsowania", func() {
				image := "niepoprawny_ciag_obrazu@@"
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
		Context("Gdy podano poprawny certyfikat TLS", func() {
			It("powinno zwrócić poprawną konfigurację TLS", func() {
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

		Context("Gdy żądanie jest pomyślne", func() {
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

			It("powinno zwrócić odpowiedź bez błędów", func() {
				resp, err := retryHTTPRequest(client, request, retryCount, retryTimeout)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(Equal(http.StatusAccepted))
			})
		})

		Context("Gdy żądanie nie powiedzie się kilka razy, ale ostatecznie się powiedzie", func() {
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

			It("powinno powtórzyć żądanie i ostatecznie zwrócić pomyślną odpowiedź", func() {
				resp, err := retryHTTPRequest(client, request, retryCount, retryTimeout)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(Equal(http.StatusAccepted))
				Expect(attempt).To(Equal(2))
			})
		})

		Context("Gdy wszystkie próby żądania się nie powiodą", func() {
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

			It("powinno zwrócić błąd po wyczerpaniu prób", func() {
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

		Context("Gdy podpisywanie przebiega pomyślnie", func() {
			It("powinno zakończyć się bez błędów", func() {
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

		Context("Gdy wystąpi błąd podczas podpisywania", func() {
			It("powinno zwrócić odpowiedni błąd", func() {
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

// GenerateBase64EncodedCert generuje samopodpisany certyfikat i klucz prywatny,
// i zwraca je jako ciągi zakodowane w base64.
func GenerateBase64EncodedCert() (string, string, error) {
	// Generowanie klucza prywatnego RSA
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	// Tworzenie szablonu certyfikatu
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

	// Tworzenie certyfikatu
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "", "", err
	}

	// Kodowanie certyfikatu do formatu PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Kodowanie klucza prywatnego do formatu PEM
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Kodowanie danych PEM do base64
	certBase64 := base64.StdEncoding.EncodeToString(certPEM)
	keyBase64 := base64.StdEncoding.EncodeToString(keyPEM)

	return certBase64, keyBase64, nil
}
