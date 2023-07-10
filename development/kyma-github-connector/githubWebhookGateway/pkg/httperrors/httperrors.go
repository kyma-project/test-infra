package httperrors

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kyma-project/test-infra/development/kyma-github-connector/githubWebhookGateway/pkg/apperrors"
	log "github.com/sirupsen/logrus"
)

type ErrorResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

func AppErrorToResponse(appError apperrors.AppError) (int, ErrorResponse) {
	httpCode := errorCodeToHTTPStatus(appError.Code())
	errorMessage := appError.Error()
	return httpCode, ErrorResponse{httpCode, fmt.Sprintf("%s: %s", appError.Desc(), errorMessage)}
}

func errorCodeToHTTPStatus(code int) int {
	switch code {
	case apperrors.NotFoundError:
		return http.StatusNotFound
	case apperrors.AlreadyExistsError:
		return http.StatusConflict
	case apperrors.WrongInputError:
		return http.StatusBadRequest
	case apperrors.UpstreamServerCallFailedError:
		return http.StatusBadGateway
	case apperrors.AuthenticationFailedError:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

//SendErrorResponse prepares the http error response and sends it to the client
func SendErrorResponse(apperr apperrors.AppError, w http.ResponseWriter) {

	httpcode, resp := AppErrorToResponse(apperr)

	w.WriteHeader(httpcode)
	respJSON, err := json.Marshal(resp)

	if err != nil {
		marshalerr := apperrors.Internal("Failed to marshal error response: %s \nError body: %s", err, apperr.Error())
		log.Warn(marshalerr)
		return
	}
	_, err = w.Write(respJSON)
	if err != nil {
		appError := apperrors.Internal("failed send response to github: %s", err.Error())
		log.Fatal(appError.Error())
	}
}
