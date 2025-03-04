// simple CORS proxy that adds necessary headers
package main

import (
	"fmt"
	"github.com/kyma-project/test-infra/pkg/gcp/cloudfunctions"
	crhttp "github.com/kyma-project/test-infra/pkg/gcp/http"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
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

	allowedDomain, err := domainInSAP(requestedURL)
	if err != nil {
		crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "couldn't check if domain is allowed%s", err)
		return
	}
	if !allowedDomain {
		crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "blocked domain requested: %s", requestedURL)
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

// domainInSAP checks if the host ends with sap.com and block all other requests
func domainInSAP(requestedURL string) (bool, error) {
	targetURL, err := url.Parse(requestedURL)
	if err != nil {
		return false, fmt.Errorf("couldn't parse URL, error: %s", err)
	}
	targetHost := targetURL.Host
	if splitHost, _, err := net.SplitHostPort(targetHost); err == nil {
		// there's a port in address, remove it
		targetHost = splitHost
	}

	if !strings.HasSuffix(targetHost, ".sap.com") {
		return false, fmt.Errorf("blocked domain requested: %s", targetHost)
	}
	return true, nil
}
# (2025-03-04)