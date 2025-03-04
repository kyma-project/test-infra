package main

import (
	"github.com/google/go-containerregistry/pkg/authn"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("newAuthenticator", func() {
	var cfg Config

	BeforeEach(func() {
		cfg = Config{}
	})

	Context("when TargetKeyFile is provided", func() {
		It("should create a basic authenticator", func() {
			cfg.TargetKeyFile = "./test-fixtures/keyfile.json"
			auth, err := cfg.newAuthenticator()
			Expect(err).NotTo(HaveOccurred())
			Expect(auth).To(Equal(&authn.Basic{
				Username: "_json_key",
				Password: "test_content",
			}))
		})

		It("should return an error if the key file cannot be read", func() {
			cfg.TargetKeyFile = "./test-fixtures/invalid_keyfile.json"
			auth, err := cfg.newAuthenticator()
			Expect(err).To(HaveOccurred())
			Expect(auth).To(BeNil())
		})
	})

	Context("when AccessToken is provided", func() {
		It("should create a bearer authenticator", func() {
			cfg.AccessToken = "valid-access-token"
			auth, err := cfg.newAuthenticator()
			Expect(err).NotTo(HaveOccurred())
			Expect(auth).To(Equal(&authn.Bearer{
				Token: "valid-access-token",
			}))
		})
	})

	Context("when neither TargetKeyFile nor AccessToken is provided", func() {
		It("should return an error", func() {
			cfg.TargetKeyFile = ""
			cfg.AccessToken = ""
			auth, err := cfg.newAuthenticator()
			Expect(err).To(HaveOccurred())
			Expect(auth).To(BeNil())
		})
	})
})
# (2025-03-04)