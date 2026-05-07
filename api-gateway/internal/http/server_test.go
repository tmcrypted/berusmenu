package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tmcrypted/berusmenu/api-gateway/internal/config"
)

func TestHealthEndpoints(t *testing.T) {
	router := NewRouter(Dependencies{Config: config.Config{}})

	tests := []string{
		"/health/live",
		"/health/ready",
	}

	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}
		})
	}
}

func TestSwaggerAndOpenAPIEndpointsAreRemoved(t *testing.T) {
	router := NewRouter(Dependencies{Config: config.Config{}})

	tests := []string{
		"/openapi.json",
		"/swagger",
	}

	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
			}
		})
	}
}
