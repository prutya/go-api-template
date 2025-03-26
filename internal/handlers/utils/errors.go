package utils

import "net/http"

const ErrCodeNotFound = "not_found"
const ErrCodeMethodNotAllowed = "method_not_allowed"
const ErrCodeInternal = "internal_error"
const ErrCodeInvalidJson = "invalid_json"
const ErrCodeInvalidParams = "invalid_params"
const ErrCodeTimeout = "timeout"
const ErrCodeUnauthorized = "unauthorized"

var ErrNotFound = NewServerError(ErrCodeNotFound, http.StatusNotFound)
var ErrMethodNotAllowed = NewServerError(ErrCodeMethodNotAllowed, http.StatusMethodNotAllowed)
var ErrInvalidJson = NewServerError(ErrCodeInvalidJson, http.StatusBadRequest)
var ErrTimeout = NewServerError(ErrCodeTimeout, http.StatusGatewayTimeout)
var ErrUnauthorized = NewServerError(ErrCodeUnauthorized, http.StatusUnauthorized)

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
