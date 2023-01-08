package main

import (
	// "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http/httputil"

	// "io"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/go-playground/validator/v10"
	"github.com/google/go-github/v48/github"
	"github.com/kyma-project/test-infra/development/gcp/pkg/cloudfunctions"
	crhttp "github.com/kyma-project/test-infra/development/gcp/pkg/http"
	"github.com/kyma-project/test-infra/development/gcp/pkg/pubsub"
	"github.com/kyma-project/test-infra/development/types"
	"github.com/spf13/viper"
	"github.com/zricethezav/gitleaks/v8/config"
	"github.com/zricethezav/gitleaks/v8/detect"
	"github.com/zricethezav/gitleaks/v8/report"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

/*
 {"project": "sap-kyma-prow", "topic": "test-topic", "runid": "runid", "status": "finished", "url": "test-url", "gcs_path": "https://gcsweb.build.kyma-project.io/gcs/kyma-prow-logs/pr-logs/pull/kyma-incubator_reconciler/1255/pre-main-kyma-incubator-component-reconciler/1600242115000930304/", "job_type": "presubmit", "job_name": "poc-scan-logs"}
*/
var (
	componentName   string
	applicationName string
	gcsPrefix       string
	projectID       string
	vc              config.ViperConfig
	findings        []report.Finding
	err             error
	cfg             config.Config
	storageClient   *storage.Client
)

type message struct {
	pubsub.ProwMessage
	types.SecretsLeakScannerMessage
	types.GCPStorageMetadata
}

func main() {
	componentName = os.Getenv("COMPONENT_NAME")     // logs-scanner
	applicationName = os.Getenv("APPLICATION_NAME") // scan-logs-for-secrets-leaks
	projectID = os.Getenv("PROJECT_ID")
	gcsPrefix = os.Getenv("GCS_PREFIX") // gcsweb.build.kyma-project.io/gcs/
	port := os.Getenv("LISTEN_PORT")

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

	// Setup configuration for gitleaks
	viper.SetConfigType("toml")
	if err := viper.ReadConfig(strings.NewReader(config.DefaultConfig)); err != nil {
		mainLogger.LogError("failed read default viper config for leaks scanner, error: %s", err)
	}

	// Load config
	if err = viper.Unmarshal(&vc); err != nil {
		mainLogger.LogCritical("Failed unmarshal viper config, got error: %s", err)
	}
	cfg, err = vc.Translate()
	if err != nil {
		mainLogger.LogCritical("Failed translate to leaks scanner config, got error: %s", err)
	}

	http.HandleFunc("/", scanLogsForSecrets)
	// Determine port for HTTP service.
	if port == "" {
		port = "8080"
		mainLogger.LogInfo("Defaulting to port %s", port)
	}
	// Start HTTP server.
	mainLogger.LogInfo("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		mainLogger.LogCritical("failed listen on port %d, error: %s", port, err)
	}
}

func scanLogsForSecrets(w http.ResponseWriter, r *http.Request) {
	var (
		bucketName  string
		objectName  string
		trace       string
		traceHeader string
		msg         message
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
		logger.LogError("failed dump http request, error:", err)
	}
	logger.LogDebug("request:\n%v", string(requestDump))

	detector := detect.NewDetector(cfg)
	detector.Redact = true

	event, err := cloudevents.NewEventFromHTTPRequest(r)
	if err != nil {
		crhttp.WriteHttpErrorResponse(w, http.StatusBadRequest, logger, "failed to parse CloudEvent from request: %s", err.Error())
		return
	}

	logger.LogInfo("got message, id: %s, type: %s", event.ID(), event.Type())

	var inmsg pubsub.Message

	err = event.DataAs(&inmsg)
	if err != nil {
		crhttp.WriteHttpErrorResponse(w, http.StatusInternalServerError, logger, "failed marshal event, error: %s", err.Error())
		return
	}

	err = json.Unmarshal(inmsg.Message.Data, &msg)
	if err != nil {
		crhttp.WriteHttpErrorResponse(w, http.StatusBadRequest, logger, "failed unmarshall pubsub message data, error: %s", err.Error())
		return
	}

	validate := validator.New()
	err = validate.Struct(msg)
	if err != nil {
		crhttp.WriteHttpErrorResponse(w, http.StatusBadRequest, logger, "missing values in config: %s", err)
		return
	}

	// Creates the new bucket.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, bucketAfter, found := strings.Cut(*msg.GcsPath, gcsPrefix)
	if !found {
		crhttp.WriteHttpErrorResponse(w, http.StatusBadRequest, logger, "failed get logs bucket name, [%s] prefix not found in gcs url", gcsPrefix)
		return
	}
	bucketName, objectName, found = strings.Cut(bucketAfter, "/")
	if !found {
		crhttp.WriteHttpErrorResponse(w, http.StatusBadRequest, logger, "failed get logs bucket name, could not find value expected separator: [/]")
		return
	}
	objectName = strings.TrimPrefix(objectName, "/")
	bucket := storageClient.Bucket(bucketName)
	battrs, err := bucket.Attrs(ctx)
	if err != nil {
		crhttp.WriteHttpErrorResponse(w, http.StatusInternalServerError, logger, "failed read google cloud storage bucket attributes, error: %s", err)
		return
	}
	logger.LogInfo("bucket: %s", battrs.Name)
	it := bucket.Objects(ctx, &storage.Query{
		Prefix: objectName,
	})
	var allFindings []report.Finding
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// TODO: Need better error messages for this scenario
			crhttp.WriteHttpErrorResponse(w, http.StatusInternalServerError, logger, "Bucket(%s).Objects: %s", bucketName, err.Error())
			return
		}
		// Wrapping into anonymous function to let defer work in expected way.
		func() {
			logger.LogInfo("starting scanning file: %s", attrs.Name)
			handle := bucket.Object(attrs.Name)
			objectReader, err := handle.NewReader(ctx)
			if err != nil {
				logger.LogError("failed create gcs object reader, error: %s", err.Error())
			}
			defer objectReader.Close()

			findings, err = detector.DetectReader(objectReader, 10)
			if err != nil {
				// log fatal to exit, no need to continue since a leaksReport
				// will not be generated when scanning from a pipe...for now
				logger.LogError("failed scan a file: %s, got error: %s", attrs.Name, err.Error())
			}
		}()

		if len(findings) != 0 {
			allFindings = append(allFindings, findings...)
			logger.LogInfo("finished scanning file, leaks found!!!!!!")
		} else {
			logger.LogInfo("finished scanning file, no leaks found")
		}
	}
	responseEvent := cloudevents.NewEvent()
	responseEvent.SetDataContentEncoding(cloudevents.TextPlain)
	responseEvent.SetSource(applicationName + "/" + componentName)
	responseEvent.SetID(applicationName + "/" + componentName + "/" + trace)

	msg.BucketName = github.String(bucketName)
	msg.Directory = github.String(objectName)
	// TODO: generating reports should be a separate function.
	if len(allFindings) != 0 {
		msg.LeaksFound = github.Bool(true)
		responseEvent.SetType("prowjob.logs.leaks.found")
		msg.LeaksReport = allFindings
	} else {
		msg.LeaksFound = github.Bool(false)
		responseEvent.SetType("prowjob.logs.leaks.notfound")
	}
	err = responseEvent.SetData(cloudevents.ApplicationJSON, msg)
	if err != nil {
		crhttp.WriteHttpErrorResponse(w, http.StatusInternalServerError, logger, "failed set event data, error: %s", err.Error())
		return
	}
	// body, err := json.Marshal(responseEvent)
	// if err != nil {
	// 	crhttp.WriteHttpErrorResponse(w, http.StatusInternalServerError, logger, "failed marshal event, error: %s", err.Error())
	// 	return
	// }
	headers := w.Header()
	headers.Set("Content-Type", cloudevents.ApplicationJSON)
	headers.Set("X-Cloud-Trace-Context", traceHeader)
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(responseEvent); err != nil {
		crhttp.WriteHttpErrorResponse(w, http.StatusInternalServerError, logger, "failed write response body, error: %s", err.Error())
		return
	}
	// w.Write(body)
}
