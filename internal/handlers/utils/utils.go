package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"

	"prutya/go-api-template/internal/logger"
)

type ErrorResponse struct {
	Error   string              `json:"error"`
	Details []ServerErrorDetail `json:"details,omitempty"`
}

const ContentTypeText = "text/plain"
const ContentTypeJson = "application/json"

var HeaderContentType = http.CanonicalHeaderKey("Content-Type")
var HeaderContentLength = http.CanonicalHeaderKey("Content-Length")

func RenderInvalidJsonError(w http.ResponseWriter, r *http.Request) {
	RenderError(w, r, ErrInvalidJson)
}

func RenderError(w http.ResponseWriter, r *http.Request, err error) {
	logger := logger.MustFromContext(r.Context())

	// Handle `ServerError`s
	if serverError, isServerError := err.(*ServerError); isServerError {
		responseInfo, hasResponseInfo := GetRequestResponseInfo(r)

		if hasResponseInfo {
			responseInfo.ErrorCode = serverError.Code
			responseInfo.InnerError = serverError.InnerError
		}

		RenderJson(w, r, &ErrorResponse{
			Error:   serverError.Code,
			Details: serverError.Details,
		}, serverError.HttpStatusCode, nil)

		return
	}

	// Handle params validation errors
	if validationErrors, isValidationErrors := err.(validator.ValidationErrors); isValidationErrors {
		details := make([]ServerErrorDetail, len(validationErrors))
		for i, e := range validationErrors {
			details[i] = ServerErrorDetail{
				Subject:    e.Field(),
				Constraint: e.Tag(),
				Args:       e.Param(),
			}
		}

		RenderJson(w, r, &ErrorResponse{
			Error:   ErrCodeInvalidParams,
			Details: details,
		}, http.StatusUnprocessableEntity, nil)

		return
	}

	logger.WarnContext(r.Context(), "Failed to handle error", "error", err)

	RenderJson(w, r, &ErrorResponse{Error: ErrCodeInternal}, http.StatusInternalServerError, nil)
}

func RenderJson(
	w http.ResponseWriter,
	r *http.Request,
	object any,
	httpStatusCode int,
	additionalHeaders map[string]string,
) {
	json, marshalErr := json.Marshal(object)

	if marshalErr != nil {
		logger := logger.MustFromContext(r.Context())
		logger.ErrorContext(r.Context(), "Failed to render an object as JSON", "object", object)

		RenderRawJson(
			w,
			r,
			[]byte(fmt.Sprintf("{\"error\":\"%s\"}", ErrCodeInternal)),
			http.StatusInternalServerError,
			nil,
		)

		return
	}

	RenderRawJson(w, r, json, httpStatusCode, additionalHeaders)
}

func RenderRawJson(
	w http.ResponseWriter,
	r *http.Request,
	json []byte,
	httpStatusCode int,
	additionalHeaders map[string]string,
) {
	w.Header().Set(HeaderContentType, ContentTypeJson)
	w.Header().Set(HeaderContentLength, strconv.Itoa(len(json)))

	for headerName, headerValue := range additionalHeaders {
		w.Header().Set(headerName, headerValue)
	}

	w.WriteHeader(httpStatusCode)

	// Write the response status code in logs
	if responseInfo, hasResponseInfo := GetRequestResponseInfo(r); hasResponseInfo {
		responseInfo.HttpStatus = httpStatusCode
	}

	if _, err := w.Write(json); err != nil {
		logger := logger.MustFromContext(r.Context())
		logger.ErrorContext(r.Context(), "Failed to write JSON", "error", err)
		panic(err)
	}
}

func RenderNoContent(
	w http.ResponseWriter,
	r *http.Request,
	additionalHeaders map[string]string,
) {
	w.Header().Set(HeaderContentType, ContentTypeText)
	w.Header().Set(HeaderContentLength, "0")

	for headerName, headerValue := range additionalHeaders {
		w.Header().Set(headerName, headerValue)
	}

	w.WriteHeader(http.StatusNoContent)

	// Write the response status code in logs
	if responseInfo, hasResponseInfo := GetRequestResponseInfo(r); hasResponseInfo {
		responseInfo.HttpStatus = http.StatusNoContent
	}
}
