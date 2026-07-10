package health

import (
	"context"
	"testing"
)

func TestServiceCheckReturnsLiveStatus(t *testing.T) {
	t.Parallel()

	service := NewService()

	result := service.Check(context.Background())

	if result.Status != "ok" {
		t.Fatalf("expected status ok, got %q", result.Status)
	}

	if result.Service != "coach-connect-api" {
		t.Fatalf("expected service coach-connect-api, got %q", result.Service)
	}
}
