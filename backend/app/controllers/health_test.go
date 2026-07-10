package controllers

import (
	"testing"

	applicationhealth "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/application/health"
)

func TestNewHealthResponseMapsApplicationResult(t *testing.T) {
	t.Parallel()

	response := newHealthResponse(applicationhealth.Result{
		Status:  "ok",
		Service: "coach-connect-api",
	})

	if response.Status != "ok" {
		t.Fatalf("expected status ok, got %q", response.Status)
	}

	if response.Service != "coach-connect-api" {
		t.Fatalf("expected service coach-connect-api, got %q", response.Service)
	}
}
