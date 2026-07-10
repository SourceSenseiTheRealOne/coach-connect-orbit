package app

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/revel/revel"

	observabilitylogging "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/observability/logging"
)

func TestRequestLoggingFilterLogsCompletedRequest(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	logger := observabilitylogging.NewJSON(&output, "coach-connect-api", slog.LevelInfo)
	start := time.Date(2026, time.July, 10, 12, 0, 0, 0, time.UTC)
	times := []time.Time{start, start.Add(25 * time.Millisecond)}
	now := func() time.Time {
		current := times[0]
		times = times[1:]
		return current
	}

	controller := revel.NewControllerEmpty()
	controller.Request.Method = "GET"
	controller.Request.URL = &url.URL{Path: "/api/v1/health", RawQuery: "token=must-not-be-logged"}

	filter := newRequestLoggingFilter(logger, func() string { return "request-test-123" }, now)
	filter(controller, []revel.Filter{func(c *revel.Controller, _ []revel.Filter) {
		c.Result = statusApplyingResult{status: 503}
	}})

	if got := controller.Args[RequestIDKey]; got != "request-test-123" {
		t.Fatalf("expected request ID in controller args, got %v", got)
	}

	if output.Len() != 0 {
		t.Fatalf("request must not be logged before Result.Apply, got %q", output.String())
	}

	controller.Result.Apply(controller.Request, controller.Response)

	var event map[string]any
	if err := json.Unmarshal(output.Bytes(), &event); err != nil {
		t.Fatalf("decode JSON log: %v", err)
	}

	assertLogField(t, event, "level", "ERROR")
	assertLogField(t, event, "msg", "request completed")
	assertLogField(t, event, "service", "coach-connect-api")
	assertLogField(t, event, "request_id", "request-test-123")
	assertLogField(t, event, "method", "GET")
	assertLogField(t, event, "path", "/api/v1/health")
	assertLogField(t, event, "status", float64(503))
	assertLogField(t, event, "duration_ms", float64(25))

	for _, forbidden := range []string{"query", "headers", "body", "token"} {
		if _, exists := event[forbidden]; exists {
			t.Fatalf("sensitive field %q must not be logged", forbidden)
		}
	}
}

func TestRequestLoggingFilterReturnsRequestIDHeader(t *testing.T) {
	t.Parallel()

	controller := revel.NewControllerEmpty()
	controller.Request.URL = &url.URL{Path: "/api/v1/health"}
	header := &recordingServerHeader{values: make(map[string][]string)}
	controller.Response.Out.Header().Server = header

	filter := newRequestLoggingFilter(
		observabilitylogging.NewJSON(&bytes.Buffer{}, "coach-connect-api", slog.LevelInfo),
		func() string { return "request-response-123" },
		time.Now,
	)
	filter(controller, []revel.Filter{revel.NilFilter})

	values := header.Get(requestIDHeader)
	if len(values) != 1 || values[0] != "request-response-123" {
		t.Fatalf("expected response request ID header, got %v", values)
	}
}

func TestRequestLoggingResultPreservesStatusOnlyResponse(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	start := time.Date(2026, time.July, 10, 12, 0, 0, 0, time.UTC)
	times := []time.Time{start, start.Add(10 * time.Millisecond)}
	now := func() time.Time {
		current := times[0]
		times = times[1:]
		return current
	}

	controller := revel.NewControllerEmpty()
	controller.Request.URL = &url.URL{Path: "/status-only"}
	header := &recordingServerHeader{values: make(map[string][]string)}
	controller.Response.Out.Header().Server = header

	filter := newRequestLoggingFilter(
		observabilitylogging.NewJSON(&output, "coach-connect-api", slog.LevelInfo),
		func() string { return "request-status-only" },
		now,
	)
	filter(controller, []revel.Filter{func(c *revel.Controller, _ []revel.Filter) {
		c.Response.Status = 204
	}})
	controller.Result.Apply(controller.Request, controller.Response)

	if header.status != 204 {
		t.Fatalf("expected status-only response to apply HTTP 204, got %d", header.status)
	}

	var event map[string]any
	if err := json.Unmarshal(output.Bytes(), &event); err != nil {
		t.Fatalf("decode JSON log: %v", err)
	}
	assertLogField(t, event, "status", float64(204))
}

func TestNewRequestIDIsCryptographicallyRandomHeaderSafeText(t *testing.T) {
	t.Parallel()

	first := newRequestID()
	second := newRequestID()

	if first == second {
		t.Fatal("expected independently generated request IDs")
	}
	if len(first) < 26 {
		t.Fatalf("expected at least 128 bits of base32 randomness, got %d characters", len(first))
	}
	if !regexp.MustCompile(`^[A-Z2-7]+$`).MatchString(first) {
		t.Fatalf("request ID is not RFC 4648 base32 text: %q", first)
	}
}

func TestRequestLogLevelUsesHTTPStatusClass(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status int
		want   slog.Level
	}{
		{name: "success", status: 200, want: slog.LevelInfo},
		{name: "redirect", status: 307, want: slog.LevelInfo},
		{name: "client error", status: 404, want: slog.LevelWarn},
		{name: "server error", status: 500, want: slog.LevelError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := requestLogLevel(test.status); got != test.want {
				t.Fatalf("expected level %s for status %d, got %s", test.want, test.status, got)
			}
		})
	}
}

func TestRequestLoggingFilterLogsRecoveredPanicAfterErrorResultApply(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	controller := revel.NewControllerEmpty()
	controller.Request.Method = "GET"
	controller.Request.URL = &url.URL{Path: "/panic"}
	filter := newRequestLoggingFilter(
		observabilitylogging.NewJSON(&output, "coach-connect-api", slog.LevelInfo),
		func() string { return "request-panic" },
		time.Now,
	)

	filter(controller, []revel.Filter{
		recoverPanicToResultFilter,
		func(_ *revel.Controller, _ []revel.Filter) { panic("expected test panic") },
	})

	if controller.Result == nil {
		t.Fatal("expected recovered panic to produce a result")
	}
	if output.Len() != 0 {
		t.Fatalf("panic response must not be logged before Result.Apply, got %q", output.String())
	}

	controller.Result.Apply(controller.Request, controller.Response)

	var event map[string]any
	if err := json.Unmarshal(output.Bytes(), &event); err != nil {
		t.Fatalf("decode JSON log: %v", err)
	}
	assertLogField(t, event, "level", "ERROR")
	assertLogField(t, event, "status", float64(500))
}

func recoverPanicToResultFilter(controller *revel.Controller, filterChain []revel.Filter) {
	defer func() {
		if recover() != nil {
			controller.Result = statusApplyingResult{status: 500}
		}
	}()
	filterChain[0](controller, filterChain[1:])
}

func assertLogField(t *testing.T, event map[string]any, field string, want any) {
	t.Helper()

	if got := event[field]; got != want {
		t.Fatalf("expected %s %v, got %v", field, want, got)
	}
}

type recordingServerHeader struct {
	values map[string][]string
	status int
}

func (h *recordingServerHeader) SetCookie(string) {}

func (h *recordingServerHeader) GetCookie(string) (revel.ServerCookie, error) {
	return nil, nil
}

func (h *recordingServerHeader) Set(key, value string) {
	h.values[key] = []string{value}
}

func (h *recordingServerHeader) Add(key, value string) {
	h.values[key] = append(h.values[key], value)
}

func (h *recordingServerHeader) Del(key string) {
	delete(h.values, key)
}

func (h *recordingServerHeader) Get(key string) []string {
	return h.values[key]
}

func (h *recordingServerHeader) GetKeys() []string {
	keys := make([]string, 0, len(h.values))
	for key := range h.values {
		keys = append(keys, key)
	}
	return keys
}

func (h *recordingServerHeader) SetStatus(status int) {
	h.status = status
}

type statusApplyingResult struct {
	status int
}

func (result statusApplyingResult) Apply(_ *revel.Request, response *revel.Response) {
	response.Status = result.status
}
