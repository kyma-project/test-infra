package oidc

// oidc_unit_test.go contains tests which require access to non-exported functions and variables.

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("OIDC", func() {
	Describe("maskToken", func() {
		It("should mask the token if length is less than 15", func() {
			token := "shorttoken"
			maskedToken := maskToken(token)
			Expect(maskedToken).To(Equal("********"))
		})

		It("should mask the token if length is exactly 15", func() {
			token := "123456789012345"
			maskedToken := maskToken(token)
			Expect(maskedToken).To(Equal("12********45"))
		})

		It("should mask the token if length is greater than 15", func() {
			token := "12345678901234567890"
			maskedToken := maskToken(token)
			Expect(maskedToken).To(Equal("12********90"))
		})

		It("should mask the token if it's empty", func() {
			token := ""
			maskedToken := maskToken(token)
			Expect(maskedToken).To(Equal("********"))
		})
	})

})
