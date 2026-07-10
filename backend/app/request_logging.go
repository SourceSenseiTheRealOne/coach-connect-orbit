package app

import (
	"context"
	"crypto/rand"
	"log/slog"
	"time"

	"github.com/revel/revel"
)

const (
	RequestIDKey    = "request_id"
	requestIDHeader = "X-Request-ID"
)

func newRequestID() string {
	return rand.Text()
}

func requestLogLevel(status int) slog.Level {
	switch {
	case status >= 500:
		return slog.LevelError
	case status >= 400:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}

func newRequestLoggingFilter(
	logger *slog.Logger,
	newRequestID func() string,
	now func() time.Time,
) revel.Filter {
	return func(controller *revel.Controller, filterChain []revel.Filter) {
		startedAt := now()
		requestID := newRequestID()
		if controller.Args == nil {
			controller.Args = make(map[string]interface{})
		}
		controller.Args[RequestIDKey] = requestID
		controller.Response.Out.Header().Set(requestIDHeader, requestID)

		filterChain[0](controller, filterChain[1:])

		controller.Result = requestLoggingResult{
			inner:      controller.Result,
			logger:     logger,
			requestID:  requestID,
			startedAt:  startedAt,
			finishedAt: now,
		}
	}
}

type requestLoggingResult struct {
	inner      revel.Result
	logger     *slog.Logger
	requestID  string
	startedAt  time.Time
	finishedAt func() time.Time
}

func (result requestLoggingResult) Apply(request *revel.Request, response *revel.Response) {
	if result.inner != nil {
		result.inner.Apply(request, response)
	} else if response.Status != 0 {
		response.SetStatus(response.Status)
	}

	status := response.Status
	if status == 0 {
		status = 200
	}

	method := ""
	path := ""
	requestContext := context.Background()
	if request != nil {
		method = request.Method
		if request.URL != nil {
			path = request.URL.Path
		}
		if request.In != nil {
			if ctx := request.Context(); ctx != nil {
				requestContext = ctx
			}
		}
	}

	result.logger.LogAttrs(
		requestContext,
		requestLogLevel(status),
		"request completed",
		slog.String("request_id", result.requestID),
		slog.String("method", method),
		slog.String("path", path),
		slog.Int("status", status),
		slog.Int64("duration_ms", result.finishedAt().Sub(result.startedAt).Milliseconds()),
	)
}
