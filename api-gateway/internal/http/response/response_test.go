package response

import (
	"net/http"
	"testing"

	"google.golang.org/grpc/codes"
)

func TestGRPCCodeToHTTP(t *testing.T) {
	tests := []struct {
		name       string
		code       codes.Code
		wantStatus int
		wantCode   string
	}{
		{name: "invalid argument", code: codes.InvalidArgument, wantStatus: http.StatusBadRequest, wantCode: "invalid_argument"},
		{name: "unauthenticated", code: codes.Unauthenticated, wantStatus: http.StatusUnauthorized, wantCode: "unauthenticated"},
		{name: "permission denied", code: codes.PermissionDenied, wantStatus: http.StatusForbidden, wantCode: "permission_denied"},
		{name: "not found", code: codes.NotFound, wantStatus: http.StatusNotFound, wantCode: "not_found"},
		{name: "already exists", code: codes.AlreadyExists, wantStatus: http.StatusConflict, wantCode: "already_exists"},
		{name: "failed precondition", code: codes.FailedPrecondition, wantStatus: http.StatusPreconditionFailed, wantCode: "failed_precondition"},
		{name: "resource exhausted", code: codes.ResourceExhausted, wantStatus: http.StatusTooManyRequests, wantCode: "resource_exhausted"},
		{name: "deadline exceeded", code: codes.DeadlineExceeded, wantStatus: http.StatusGatewayTimeout, wantCode: "deadline_exceeded"},
		{name: "unavailable", code: codes.Unavailable, wantStatus: http.StatusServiceUnavailable, wantCode: "unavailable"},
		{name: "unknown", code: codes.Unknown, wantStatus: http.StatusInternalServerError, wantCode: "internal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStatus, gotCode := grpcCodeToHTTP(tt.code)
			if gotStatus != tt.wantStatus {
				t.Fatalf("status = %d, want %d", gotStatus, tt.wantStatus)
			}
			if gotCode != tt.wantCode {
				t.Fatalf("code = %q, want %q", gotCode, tt.wantCode)
			}
		})
	}
}
