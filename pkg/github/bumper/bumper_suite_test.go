package bumper

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBumper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bumper Suite")
}
