package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitlab.com/berusmenu/api-gateway/internal/config"
)

func TestOpenAPIEndpoint(t *testing.T) {
	router := NewRouter(Dependencies{Config: config.Config{}})

	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal openapi: %v", err)
	}

	if body["openapi"] != "3.0.3" {
		t.Fatalf("openapi = %v, want 3.0.3", body["openapi"])
	}
}

func TestSwaggerUIEndpoint(t *testing.T) {
	router := NewRouter(Dependencies{Config: config.Config{}})

	req := httptest.NewRequest(http.MethodGet, "/swagger", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "SwaggerUIBundle") {
		t.Fatalf("body does not contain Swagger UI bootstrap: %s", rec.Body.String())
	}
}
