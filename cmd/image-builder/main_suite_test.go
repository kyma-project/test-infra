package main

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestImageBuilder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Image Builder Suite")
}
