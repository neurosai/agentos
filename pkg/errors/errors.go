package apperrors

import (
	"errors"
	"fmt"
	"net/http"
)

// Code is a stable machine-readable error code from the AgentOS API spec.
type Code string

const (
	CodeInvalidRequest      Code = "invalid_request"
	CodeUnauthenticated     Code = "unauthenticated"
	CodeForbidden           Code = "forbidden"
	CodeNotFound            Code = "not_found"
	CodeConflict            Code = "conflict"
	CodePolicyDenied        Code = "policy_denied"
	CodeRateLimited         Code = "rate_limited"
	CodeBackendUnavailable  Code = "backend_unavailable"
)

// Error is the canonical API error envelope.
type Error struct {
	Code       Code
	Message    string
	RequestID  string
	Retryable  bool
	Details    map[string]any
	StatusCode int
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// HTTPStatus returns the HTTP status for this error code.
func HTTPStatus(code Code) int {
	switch code {
	case CodeInvalidRequest:
		return http.StatusBadRequest
	case CodeUnauthenticated:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	case CodePolicyDenied:
		return http.StatusUnprocessableEntity
	case CodeRateLimited:
		return http.StatusTooManyRequests
	case CodeBackendUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// New creates an API error with the default HTTP status for code.
func New(code Code, message string) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		StatusCode: HTTPStatus(code),
	}
}

// WithRequestID attaches a request identifier.
func (e *Error) WithRequestID(id string) *Error {
	e.RequestID = id
	return e
}

// WithDetails attaches structured details.
func (e *Error) WithDetails(details map[string]any) *Error {
	e.Details = details
	return e
}

// WithRetryable marks whether the client may retry.
func (e *Error) WithRetryable(retryable bool) *Error {
	e.Retryable = retryable
	return e
}

// AsAPIError unwraps or maps err to *Error.
func AsAPIError(err error) *Error {
	var apiErr *Error
	if errors.As(err, &apiErr) {
		return apiErr
	}
	return New(CodeBackendUnavailable, err.Error()).WithRetryable(true)
}
