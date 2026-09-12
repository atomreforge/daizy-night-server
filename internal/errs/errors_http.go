package errs

import (
	"net/http"
)

type BizError struct {
	http int `json:"-"`
	//BizEventCode       int    `json:"code"`
	Message string `json:"message"`
}

// standard Error method for Golang
func (e *BizError) Error() string { return e.Message }

// returns the errtype's http stat code
func (e *BizError) StatusCode() int { return e.http }

// returns a user-friendly message for the errtype, which can be used in the response body
// [!] notice that there may be a risk to expose sensitive info here
func (e *BizError) Respond() string { return e.Message }

// http errs below.

func BuildBadRequest(msg string) *BizError {
	return &BizError{http: http.StatusBadRequest, Message: msg}
}

func BuildUnauthorized(msg string) *BizError {
	return &BizError{http: http.StatusUnauthorized, Message: msg}
}
func BuildNotFound(msg string) *BizError {
	return &BizError{http: http.StatusNotFound, Message: msg}
}
func BuildInternal(msg string) *BizError {
	return &BizError{http: http.StatusInternalServerError, Message: msg}
}
