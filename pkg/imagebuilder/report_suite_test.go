package imagebuilder

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPipelines(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Image Builder Suite")
}
