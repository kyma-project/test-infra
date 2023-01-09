package http

import (
	"fmt"
	"net/http"

	"github.com/kyma-project/test-infra/development/gcp/pkg/cloudfunctions"
)

func WriteHttpErrorResponse(w http.ResponseWriter, statusCode int, logger *cloudfunctions.LogEntry, format string, args ...interface{}) {
	errorMessage := fmt.Sprintf(format, args...)
	logger.LogError(errorMessage)
	http.Error(w, errorMessage, statusCode)
}
