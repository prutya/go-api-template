package utils

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type ErrorResponse struct {
	Error   string              `json:"error"`
	Details []ServerErrorDetail `json:"details,omitempty"`
}

const ContentTypeText = "text/plain"
const ContentTypeJson = "application/json"

var HeaderContentType = http.CanonicalHeaderKey("Content-Type")
var HeaderContentLength = http.CanonicalHeaderKey("Content-Length")
var EmptyHeaders = map[string]string{}

func RenderError(w http.ResponseWriter, r *http.Request, err error) {
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

	RenderJson(w, r, &ErrorResponse{Error: ErrCodeInternal}, http.StatusInternalServerError, nil)
}

func RenderJson(w http.ResponseWriter, r *http.Request, object any, httpStatusCode int, headers map[string]string) {
	responseInfo, hasResponseInfo := GetRequestResponseInfo(r)

	json, marshalErr := json.Marshal(object)

	if marshalErr != nil {
		msg := []byte(ErrCodeInternal)

		w.Header().Set(HeaderContentType, ContentTypeText)
		w.Header().Set(HeaderContentLength, strconv.Itoa(len(msg)))

		w.WriteHeader(http.StatusInternalServerError)

		if hasResponseInfo {
			responseInfo.HttpStatus = http.StatusInternalServerError
		}

		w.Write(msg)

		return
	}

	w.Header().Set(HeaderContentType, ContentTypeJson)
	w.Header().Set(HeaderContentLength, strconv.Itoa(len(json)))

	for k, v := range headers {
		w.Header().Set(k, v)
	}

	w.WriteHeader(httpStatusCode)

	if hasResponseInfo {
		responseInfo.HttpStatus = httpStatusCode
	}

	w.Write(json)
}
