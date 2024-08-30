package http

import (
	"fmt"
	"net/http"

	"github.com/kyma-project/test-infra/pkg/gcp/cloudfunctions"
)

// WriteHttpErrorResponse format error message, log it with error severity using passed logger
// It writes http error response with provided status code and formatted error message to http.ResponseWrite function argument.
func WriteHTTPErrorResponse(w http.ResponseWriter, statusCode int, logger *cloudfunctions.LogEntry, format string, args ...interface{}) {
	logger.LogError(format, args...)
	http.Error(w, fmt.Sprintf(format, args...), statusCode)
}
