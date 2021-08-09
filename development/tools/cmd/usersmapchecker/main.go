package usersmapchecker

import (
	"cloud.google.com/go/logging"
	"context"
	"github.com/kyma-project/test-infra/development/github/pkg/client"
	"google.golang.org/api/option"
	"os"
)

// TODO: move to common package
const (
	errorReportingType = "type.googleapis.com/google.devtools.clouderrorreporting.v1beta1.ReportedErrorEvent"
)

// TODO: move to common package
type loggingPayload struct {
	Message   string `json:"message"`
	Operation string `json:"operation,omitempty"`
	Type      string `json:"@type,omitempty"`
}

func main() {
	ctx := context.Background()
	saProwjobGcpLoggingClientKeyPath := os.Getenv("SA_PROWJOB_GCP_LOGGING_CLIENT_KEY_PATH")
	logging.NewClient(ctx, "testLogName")
	// provided by preset-bot-github-sap-token
	accessToken := os.Getenv("BOT_GITHUB_SAP_TOKEN")
	saptoolsClient, err := client.NewSapToolsClient(ctx, accessToken)
	if err != nil {

	}
}
