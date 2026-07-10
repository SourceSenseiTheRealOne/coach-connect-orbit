package logging

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestNewJSONWritesStructuredServiceFields(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	logger := NewJSON(&output, "coach-connect-api", slog.LevelInfo)

	logger.Info("server started", slog.Int("port", 9000))

	var event map[string]any
	if err := json.Unmarshal(output.Bytes(), &event); err != nil {
		t.Fatalf("decode JSON log: %v", err)
	}

	assertField(t, event, "level", "INFO")
	assertField(t, event, "msg", "server started")
	assertField(t, event, "service", "coach-connect-api")
	assertField(t, event, "port", float64(9000))
}

func assertField(t *testing.T, event map[string]any, field string, want any) {
	t.Helper()

	if got := event[field]; got != want {
		t.Fatalf("expected %s %v, got %v", field, want, got)
	}
}
