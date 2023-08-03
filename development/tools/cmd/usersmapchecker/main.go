package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"

	log "github.com/sirupsen/logrus"

	gcplogging "github.com/kyma-project/test-infra/development/gcp/pkg/logging"
	"github.com/kyma-project/test-infra/development/github/pkg/client"
	"github.com/kyma-project/test-infra/development/prow"
)

// Example fields in gcp logging.
// logName: "projects/sap-kyma-prow/logs/stdout"
//
//	resource: {
//	  labels: {
//	    cluster_name: "trusted-workload-kyma-prow"
//	    container_name: "test"
//	    location: "europe-west3"
//	    namespace_name: "default"
//	    pod_name: "cbb59657-fa91-11eb-baea-4e9acc7ce5e6"
//	    project_id: "sap-kyma-prow"
//	  }
//	  type: "k8s_container"
//
//	labels: {
//	  compute.googleapis.com/resource_name: "gke-trusted-workload-k-high-cpu-16-32-c8294afe-skrq"
//	  k8s-pod/created-by-prow: "true"
//	  k8s-pod/event-GUID: "cb549a8a-fa91-11eb-80a9-35f1ac609512"
//	  k8s-pod/preset-build-main: "true"
//	  k8s-pod/preset-cluster-use-ssd: "true"
//	  k8s-pod/preset-cluster-version: "true"
//	  k8s-pod/preset-debug-commando-oom: "true"
//	  k8s-pod/preset-dind-enabled: "true"
//	  k8s-pod/preset-docker-push-repository-gke-integration: "true"
//	  k8s-pod/preset-gc-compute-envs: "true"
//	  k8s-pod/preset-gc-project-env: "true"
//	  k8s-pod/preset-gke-upgrade-post-job: "true"
//	  k8s-pod/preset-kyma-artifacts-bucket: "true"
//	  k8s-pod/preset-kyma-guard-bot-github-token: "true"
//	  k8s-pod/preset-log-collector-slack-token: "true"
//	  k8s-pod/preset-sa-gke-kyma-integration: "true"
//	  k8s-pod/preset-sa-test-gcr-push: "true"
//	  k8s-pod/prow_k8s_io/build-id: "1425409012446269440"
//	  k8s-pod/prow_k8s_io/context: "post-main-kyma-gke-upgrade"
//	  k8s-pod/prow_k8s_io/id: "cbb59657-fa91-11eb-baea-4e9acc7ce5e6"
//	  k8s-pod/prow_k8s_io/job: "post-main-kyma-gke-upgrade"
//	  k8s-pod/prow_k8s_io/plank-version: "v20210714-62f15287bd"
//	  k8s-pod/prow_k8s_io/pubsub_project: "sap-kyma-prow"
//	  k8s-pod/prow_k8s_io/pubsub_runID: "post-main-kyma-gke-upgrade"
//	  k8s-pod/prow_k8s_io/pubsub_topic: "prowjobs"
//	  k8s-pod/prow_k8s_io/refs_base_ref: "main"
//	  k8s-pod/prow_k8s_io/refs_org: "kyma-project"
//	  k8s-pod/prow_k8s_io/refs_repo: "kyma"
//	  k8s-pod/prow_k8s_io/type: "postsubmit"
//	}

// Logger is an interface for logging messages.
type Logger interface {
	LogInfo(message string)
	LogError(message string)
	Flush()
}

// DualLogger is a type that implements the Logger interface.
// It writes log messages to two destinations - a GCP logger and a Logrus logger.
type DualLogger struct {
	gcpLogger    *gcplogging.Logger
	logrusLogger *log.Entry
}

// NewDualLogger creates a new DualLogger.
func NewDualLogger(gcpLogger *gcplogging.Logger, context string) Logger {
	logrusLogger := log.WithField("context", context)
	return &DualLogger{
		gcpLogger:    gcpLogger,
		logrusLogger: logrusLogger,
	}
}

// LogInfo logs an information message using both the GCP and Logrus loggers.
func (dl *DualLogger) LogInfo(message string) {
	dl.gcpLogger.LogInfo(message)
	dl.logrusLogger.Info(message)
}

// LogError logs an error message using both the GCP and Logrus loggers.
func (dl *DualLogger) LogError(message string) {
	dl.gcpLogger.LogError(message)
	dl.logrusLogger.Error(message)
}

// Flush flushes the GCP logger to ensure all log messages are written.
func (dl *DualLogger) Flush() {
	dl.gcpLogger.Flush()
}

func main() {
	// exitCode holds exit code to report at the end of main execution, it's safe to set it from multiple goroutines.
	var exitCode atomic.Value
	// Set exit code for exec. This will be call last when exiting from main function.
	defer func() {
		os.Exit(exitCode.Load().(int))
	}()
	ctx := context.Background()
	var wg sync.WaitGroup
	// Serviceaccount credentials to access google cloud logging API.
	saProwjobGcpLoggingClientKeyPath := os.Getenv("SA_PROWJOB_GCP_LOGGING_CLIENT_KEY_PATH")
	// Create kyma implementation Google cloud logging client with defaults for logging from prowjobs.
	logClient, err := gcplogging.NewProwjobClient(ctx, saProwjobGcpLoggingClientKeyPath, gcplogging.ProwLogsProjectID)
	if err != nil {
		log.Errorf("creating gcp logging client failed, got error: %v", err)
	}
	gcpLogger := logClient.NewProwjobLogger().WithGeneratedTrace()
	// Flush all buffered messages when exiting from main function.
	logger := NewDualLogger(gcpLogger, "checking if user exists in users map")
	defer logger.Flush()
	// Github access token, provided by preset-bot-github-sap-token
	accessToken := os.Getenv("BOT_GITHUB_SAP_TOKEN")
	githubComAccessToken := os.Getenv("BOT_GITHUB_TOKEN")
	saptoolsClient, err := client.NewSapToolsClient(ctx, accessToken)
	if err != nil {
		logger.LogError(fmt.Sprintf("failed creating sap tools github client, got error: %v", err))
	}

	githubComClient, err := client.NewClient(ctx, githubComAccessToken)
	if err != nil {
		logger.LogError(fmt.Sprintf("failed creating github.com client, got error: %v", err))
	}
	usersMap, err := saptoolsClient.GetUsersMap(ctx)
	if err != nil {
		logger.LogError(fmt.Sprintf("error when getting users map: got error %v", err))
	}
	authors, err := prow.GetPrAuthorForPresubmit()
	if err != nil {
		if notPresubmit := prow.IsNotPresubmitError(err); *notPresubmit {
			logger.LogInfo(err.Error())
		} else {
			logger.LogError(fmt.Sprintf("error when getting pr author for presubmit: got error %v", err))
		}
	}

	org, err := prow.GetOrgForPresubmit()
	if err != nil {
		if notPresubmit := prow.IsNotPresubmitError(err); *notPresubmit {
			logger.LogInfo(err.Error())
		} else {
			logger.LogError(fmt.Sprintf("error when getting org for presubmit: got error %v", err))
		}
	}
	wg.Add(len(authors))
	logger.LogInfo(fmt.Sprintf("found %d authors in job spec env variable", len(authors)))
	for _, author := range authors {
		member, _, err := githubComClient.Organizations.IsMember(ctx, org, author)
		if err != nil {
			logger.LogInfo(fmt.Sprintf("failed check if user %s is an github organisation member", author))
		}
		if member {
			go func(wg *sync.WaitGroup, author string, exitCode *atomic.Value) {
				defer wg.Done()
				for _, user := range usersMap {
					if user.ComGithubUsername == author {
						logger.LogInfo(fmt.Sprintf("user %s is present in users map", author))
						return
					}
				}
				logger.LogError(fmt.Sprintf("user %s is not present in users map, please add user to users-map.yaml file.", author))
				exitCode.Store(1)
			}(&wg, author, &exitCode)
		} else {
			wg.Done()
		}
	}
	wg.Wait()
	// If exitcode is nil, that means no errors were reported.
	if exitCode.Load() == nil {
		logger.LogInfo("all authors present in users map or are not members of pull request github organisation")
		logger.Flush()

		// Report successful prowjob execution.
		exitCode.Store(0)
	}
}
