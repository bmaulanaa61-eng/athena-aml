package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthRouteReturnsServiceStatus(t *testing.T) {
	router := SetupRoutes(nil)
	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %s", rec.Code, rec.Body.String())
	}

	var payload map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if payload["service"] != "athena-api" || payload["status"] != "OK" || payload["database"] != "DOWN" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected request id header to be set")
	}
}
