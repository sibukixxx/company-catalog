package houjin

import (
	"errors"
	"fmt"
	"testing"
)

func TestAPIErrorErrorMessageContainsCodeAndMessage(t *testing.T) {
	e := &APIError{
		Code:    ErrorCodeNotFound,
		Message: "no result",
	}

	msg := e.Error()
	if msg == "" {
		t.Fatal("Error() returned empty string")
	}
	for _, want := range []string{"not_found", "no result"} {
		if !contains(msg, want) {
			t.Errorf("Error() = %q, missing %q", msg, want)
		}
	}
}

func TestAPIErrorUnwrapReturnsCause(t *testing.T) {
	cause := errors.New("network down")
	e := &APIError{
		Code:  ErrorCodeTimeout,
		Cause: cause,
	}

	if got := errors.Unwrap(e); !errors.Is(got, cause) {
		t.Errorf("Unwrap() = %v, want %v", got, cause)
	}
}

func TestAPIErrorWorksWithErrorsAs(t *testing.T) {
	original := &APIError{Code: ErrorCodeRateLimited, Message: "slow down"}
	wrapped := fmt.Errorf("call failed: %w", original)

	var got *APIError
	if !errors.As(wrapped, &got) {
		t.Fatal("errors.As should match *APIError")
	}
	if got.Code != ErrorCodeRateLimited {
		t.Errorf("got code %q, want %q", got.Code, ErrorCodeRateLimited)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
