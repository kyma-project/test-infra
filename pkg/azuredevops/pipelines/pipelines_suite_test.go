package pipelines_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPipelines(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pipelines Suite")
}
# (2025-03-04)