package apperrors

import (
	"net/http"
	"testing"
)

func TestHTTPStatusMapping(t *testing.T) {
	t.Parallel()

	cases := []struct {
		code   Code
		status int
	}{
		{CodeInvalidRequest, http.StatusBadRequest},
		{CodeUnauthenticated, http.StatusUnauthorized},
		{CodeForbidden, http.StatusForbidden},
		{CodeNotFound, http.StatusNotFound},
		{CodeConflict, http.StatusConflict},
		{CodePolicyDenied, http.StatusUnprocessableEntity},
		{CodeRateLimited, http.StatusTooManyRequests},
		{CodeBackendUnavailable, http.StatusServiceUnavailable},
	}
	for _, tc := range cases {
		if got := HTTPStatus(tc.code); got != tc.status {
			t.Fatalf("code %s: got %d want %d", tc.code, got, tc.status)
		}
	}
}

func TestAsAPIError(t *testing.T) {
	t.Parallel()

	orig := New(CodePolicyDenied, "denied")
	got := AsAPIError(orig)
	if got.Code != CodePolicyDenied {
		t.Fatalf("got code %s", got.Code)
	}
}
