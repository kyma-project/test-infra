package main

import (
	"net/http"

	"github.com/kyma-project/test-infra/pkg/gcp/cloudfunctions"
)

func main() {}

func CLAAssistant(logger *cloudfunctions.LogEntry) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
