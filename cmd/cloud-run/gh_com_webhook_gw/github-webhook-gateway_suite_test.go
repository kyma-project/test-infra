package main

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGithubWebhookGateway(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GithubWebhookGateway Suite")
}
