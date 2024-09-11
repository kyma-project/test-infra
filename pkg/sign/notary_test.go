package sign

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDecodeCertAndKeySuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DecodeCertAndKey Suite")
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

func init() {
	setupDecodeCertAndKeyTests()
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
