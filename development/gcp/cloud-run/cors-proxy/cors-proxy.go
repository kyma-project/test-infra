// simple CORS proxy that adds necessary headers
package main

import (
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/kyma-project/test-infra/development/gcp/pkg/cloudfunctions"
	crhttp "github.com/kyma-project/test-infra/development/gcp/pkg/http"
)

var (
	componentName   string
	applicationName string
	listenPort      string
)

func main() {
	componentName = os.Getenv("COMPONENT_NAME")     // github-webhook-gateway
	applicationName = os.Getenv("APPLICATION_NAME") // github-webhook-gateway
	listenPort = os.Getenv("LISTEN_PORT")

	mainLogger := cloudfunctions.NewLogger()
	mainLogger.WithComponent(componentName) // search-github-issue
	mainLogger.WithLabel("io.kyma.app", applicationName)
	mainLogger.WithLabel("io.kyma.component", componentName)

	http.HandleFunc("/", CORSProxy)
	// Determine listenPort for HTTP service.
	if listenPort == "" {
		listenPort = "8080"
		mainLogger.LogInfo("Defaulting to listenPort %s", listenPort)
	}
	// Start HTTP server.
	mainLogger.LogInfo("Listening on listenPort %s", listenPort)
	if err := http.ListenAndServe(":"+listenPort, nil); err != nil {
		mainLogger.LogCritical("failed listen on listenPort %s, error: %s", listenPort, err)
	}
}

func CORSProxy(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	logger := cloudfunctions.NewLogger()
	logger.WithComponent(componentName)
	logger.WithLabel("io.kyma.app", applicationName)
	logger.WithLabel("io.kyma.component", componentName)

	escapedURL := r.URL.Query().Get("url")
	if escapedURL == "" {
		crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "got empty url param %s", r.URL)
		return
	}

	logger.LogInfo("Got request for %s", escapedURL)

	requestedURL, err := url.QueryUnescape(escapedURL)
	if err != nil {
		crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed unescaping url, error: %s", err)
		return
	}

	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
	w.Header().Add("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

	if r.Method != "OPTIONS" {
		// don't forward OPTIONS
		upstreamRequest, err := http.NewRequest(r.Method, requestedURL, r.Body)
		if err != nil {
			crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "couldn't create upstream request, error: %s", err)
			return
		}
		upstreamRequest.Header.Add("Content-Type", "application/json")

		client := http.Client{}
		resp, err := client.Do(upstreamRequest)
		if err != nil {
			crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "couldn't execute upstream request, error: %s", err)
			return
		}

		_, err = io.Copy(w, resp.Body)
		if err != nil {
			crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "couldn't copy response, error: %s", err)
		}
	}
}
