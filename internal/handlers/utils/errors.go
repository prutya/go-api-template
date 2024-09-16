package utils

import (
	"net/http"
)

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

var ErrInvalidJson = NewServerError(ErrCodeInvalidJson, http.StatusBadRequest)

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

func RenderInvalidJsonError(w http.ResponseWriter, r *http.Request) {
	RenderError(w, r, ErrInvalidJson)
}
