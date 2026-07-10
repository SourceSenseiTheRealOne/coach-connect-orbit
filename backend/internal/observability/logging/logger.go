package logging

import (
	"io"
	"log/slog"
)

func NewJSON(output io.Writer, service string, level slog.Leveler) *slog.Logger {
	handler := slog.NewJSONHandler(output, &slog.HandlerOptions{Level: level})
	return slog.New(handler).With(slog.String("service", service))
}
