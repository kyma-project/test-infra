package pjtester_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPjtester(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pjtester Suite")
}
