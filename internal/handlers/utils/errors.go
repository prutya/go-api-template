package utils

import "net/http"

const ErrCodeNotFound = "not_found"
const ErrCodeMethodNotAllowed = "method_not_allowed"
const ErrCodeInternal = "internal_error"
const ErrCodeInvalidJson = "invalid_json"
const ErrCodeInvalidQuery = "invalid_query"
const ErrCodeUnauthorized = "unauthorized"
const ErrCodeConflict = "conflict"
const ErrCodeUnprocessableContent = "unprocessable_content"
const ErrCodeInvalidParams = "invalid_params"
const ErrCodeTimeout = "timeout"
const ErrCodeInvalidPayload = "invalid_payload"
const ErrCodeInvalidCaptcha = "invalid_captcha"
const ErrCodeTooManyRequests = "too_many_requests"

var ErrNotFound = NewServerError(ErrCodeNotFound, http.StatusNotFound)
var ErrMethodNotAllowed = NewServerError(ErrCodeMethodNotAllowed, http.StatusMethodNotAllowed)
var ErrInvalidJson = NewServerError(ErrCodeInvalidJson, http.StatusBadRequest)
var ErrInvalidQuery = NewServerError(ErrCodeInvalidQuery, http.StatusBadRequest)
var ErrUnauthorized = NewServerError(ErrCodeUnauthorized, http.StatusUnauthorized)
var ErrConflict = NewServerError(ErrCodeConflict, http.StatusConflict)
var ErrUnprocessableContent = NewServerError(ErrCodeUnprocessableContent, http.StatusUnprocessableEntity)
var ErrInvalidPayload = NewServerError(ErrCodeInvalidPayload, http.StatusUnprocessableEntity)
var ErrInvalidCaptcha = NewServerError(ErrCodeInvalidCaptcha, http.StatusUnprocessableEntity)
var ErrTimeout = NewServerError(ErrCodeTimeout, http.StatusGatewayTimeout)
var ErrTooManyRequests = NewServerError(ErrCodeTooManyRequests, http.StatusTooManyRequests)

type ServerError struct {
	HttpStatusCode int
	Code           string
	Details        []ServerErrorDetail
	InnerError     error
}

type ServerErrorDetail struct {
	Subject    string `json:"subject"`
	Constraint string `json:"constraint"`
	Args       string `json:"args"`
}

func NewServerError(code string, httpStatusCode int) *ServerError {
	return NewServerErrorWithInner(code, httpStatusCode, nil)
}

func NewServerErrorWithInner(code string, httpStatusCode int, innerError error) *ServerError {
	return &ServerError{
		HttpStatusCode: httpStatusCode,
		Code:           code,
		InnerError:     innerError,
	}
}

// Debug only
func (e *ServerError) Error() string {
	return e.Code
}
