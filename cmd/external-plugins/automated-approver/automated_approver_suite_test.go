package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAutomatedApprover(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AutomatedApprover Suite")
}
