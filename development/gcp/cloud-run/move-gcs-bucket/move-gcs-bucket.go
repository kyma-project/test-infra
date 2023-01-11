package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/go-github/v48/github"
	"github.com/kyma-project/test-infra/development/gcp/pkg/cloudfunctions"
	crhttp "github.com/kyma-project/test-infra/development/gcp/pkg/http"
	"github.com/kyma-project/test-infra/development/gcp/pkg/pubsub"
	gcptypes "github.com/kyma-project/test-infra/development/gcp/pkg/types"
	githubtypes "github.com/kyma-project/test-infra/development/github/pkg/types"
	"github.com/kyma-project/test-infra/development/types"
	"google.golang.org/api/iterator"
)

var (
	dstBucketName   string
	componentName   string
	applicationName string
	projectID       string
	err             error
	storageClient   *storage.Client
)

// TODO: Rename to msg.
type message struct {
	pubsub.ProwMessage
	types.SecretsLeakScannerMessage
	githubtypes.SearchIssuesResult
	gcptypes.GCPBucketMetadata
	gcptypes.GCPProjectMetadata
	githubtypes.IssueMetadata
}

func main() {
	componentName = os.Getenv("COMPONENT_NAME")
	applicationName = os.Getenv("APPLICATION_NAME")
	projectID = os.Getenv("PROJECT_ID")
	port := os.Getenv("LISTEN_PORT")
	dstBucketName = os.Getenv("DST_BUCKET_NAME")

	mainLogger := cloudfunctions.NewLogger()
	mainLogger.WithComponent(componentName)
	mainLogger.WithLabel("io.kyma.app", applicationName)
	mainLogger.WithLabel("io.kyma.component", componentName)

	ctx := context.Background()
	// Creates a storageClient.
	storageClient, err = storage.NewClient(ctx)
	if err != nil {
		mainLogger.LogCritical("failed to create storageClient: %s", err)
	}

	defer storageClient.Close()

	http.HandleFunc("/", moveGCPBucket)
	// Determine port for HTTP service.
	if port == "" {
		port = "8080"
		mainLogger.LogInfo("Defaulting to port %s", port)
	}
	// Start HTTP server.
	mainLogger.LogInfo("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		mainLogger.LogCritical("failed listen on port %s, error: %s", port, err.Error())
	}
}

func moveGCPBucket(w http.ResponseWriter, r *http.Request) {
	var (
		msg         message
		trace       string
		traceHeader string
	)

	traceHeader = r.Header.Get("X-Cloud-Trace-Context")

	if projectID != "" {
		traceParts := strings.Split(traceHeader, "/")
		if len(traceParts) > 0 && len(traceParts[0]) > 0 {
			trace = fmt.Sprintf("projects/%s/traces/%s", projectID, traceParts[0])
		}
	}

	logger := cloudfunctions.NewLogger()
	logger.WithComponent(componentName)
	logger.WithLabel("io.kyma.app", applicationName)
	logger.WithLabel("io.kyma.component", componentName)
	logger.WithTrace(trace)

	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		logger.LogError("failed dump http request, error: %s", err)
	}
	logger.LogDebug("request:\n%v", string(requestDump))

	event, err := cloudevents.NewEventFromHTTPRequest(r)
	if err != nil {
		crhttp.WriteHTTPErrorResponse(w, http.StatusBadRequest, logger, "failed to parse CloudEvent from request: %s", err.Error())
		return
	}

	logger.LogInfo("got message, id: %s type: %s", event.ID(), event.Type())

	if err = event.DataAs(&msg); err != nil {
		crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed marshal event, error: %s", err.Error())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	objectName := strings.TrimSuffix(*msg.GCPBucketDirectory, "/")
	bucket := storageClient.Bucket(*msg.GCPBucketName)
	battrs, err := bucket.Attrs(ctx)
	if err != nil {
		crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed read google cloud storage bucket attributes, error: %s", err)
	}
	logger.LogInfo("bucket: %s", battrs.Name)
	it := bucket.Objects(ctx, &storage.Query{
		Prefix: objectName,
	})
	logger.LogInfo("starting moving blobs")
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed getting next bucket object %s, error: %s", attrs.Name, err.Error())
		}
		logger.LogDebug("starting copying file: %s", attrs.Name)
		src := storageClient.Bucket(*msg.GCPBucketName).Object(attrs.Name)
		dst := storageClient.Bucket(dstBucketName).Object(attrs.Name)
		logger.LogDebug("src object name: %s", *msg.GCPBucketName+"/"+src.ObjectName())
		logger.LogDebug("dst object name: %s", dstBucketName+"/"+dst.ObjectName())
		if _, err = dst.CopierFrom(src).Run(ctx); err != nil {
			crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed copy object %s to bucket %s, error: %s", *msg.GCPBucketDirectory, dstBucketName, err.Error())
			return
		}
		logger.LogDebug("Removing source object %s", *msg.GCPBucketName+"/"+src.ObjectName())
		if err := src.Delete(ctx); err != nil {
			crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed remove object %s from bucket %s, error: %s", *msg.GCPBucketDirectory, *msg.GCPBucketName, err.Error())
			return
		}
		logger.LogInfo("Blob %s moved to %s", *msg.GCPBucketName+"/"+*msg.GCPBucketDirectory, dstBucketName+"/"+*msg.GCPBucketDirectory)
	}
	responseEvent := cloudevents.NewEvent()
	responseEvent.SetSource(applicationName + "/" + componentName)
	responseEvent.SetID(applicationName + "/" + componentName + "/" + trace)
	responseEvent.SetType("gcp.storage.bucket.moved")
	responseEvent.SetDataContentEncoding(cloudevents.TextPlain)
	msg.GCPBucketName = github.String(dstBucketName)
	if err = responseEvent.SetData(cloudevents.ApplicationJSON, msg); err != nil {
		crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed set event data, error: %s", err.Error())
		return
	}
	headers := w.Header()
	headers.Set("Content-Type", cloudevents.ApplicationJSON)
	headers.Set("X-Cloud-Trace-Context", traceHeader)
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(responseEvent); err != nil {
		crhttp.WriteHTTPErrorResponse(w, http.StatusInternalServerError, logger, "failed write response body, error: %s", err.Error())
		return
	}
}
