package houjin

import "fmt"

// ErrorCode classifies acquisition errors; source-package errors are ordinary
// errors and need not carry a code.
type ErrorCode string

const (
	ErrorCodeInvalidInput ErrorCode = "invalid_input"
	ErrorCodeUnauthorized ErrorCode = "unauthorized"
	ErrorCodeNotFound     ErrorCode = "not_found"
	ErrorCodeRateLimited  ErrorCode = "rate_limited"
	ErrorCodeAPIError     ErrorCode = "api_error"
	ErrorCodeParseError   ErrorCode = "parse_error"
	ErrorCodeTimeout      ErrorCode = "timeout"
)

// APIError carries a safe top-level message and optional HTTP status. Error()
// omits Cause; the unwrapped transport cause may include a secret request URL.
// Use errors.As for classification and do not log raw causes.
type APIError struct {
	Code       ErrorCode
	StatusCode int
	Message    string
	Cause      error
}

// Error formats the public message without appending the transport cause.
func (e *APIError) Error() string {
	if e.StatusCode != 0 {
		return fmt.Sprintf("houjin: %s (status=%d): %s", e.Code, e.StatusCode, e.Message)
	}
	return fmt.Sprintf("houjin: %s: %s", e.Code, e.Message)
}

// Unwrap exposes Cause for errors.Is/As; it may contain sensitive transport data.
func (e *APIError) Unwrap() error {
	return e.Cause
}
