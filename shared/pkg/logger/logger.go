package logger

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

// Logger wraps zerolog with OpenTelemetry trace context
type Logger struct {
	zerolog.Logger
}

// New creates a new logger with pretty console output for development
func New(serviceName string) *Logger {
	log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).
		With().
		Timestamp().
		Str("service", serviceName).
		Logger()

	return &Logger{Logger: log}
}

// NewProduction creates a JSON logger for production
func NewProduction(serviceName string) *Logger {
	log := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", serviceName).
		Logger()

	return &Logger{Logger: log}
}

// InfoWithTrace logs with trace context from OpenTelemetry
func (l *Logger) InfoWithTrace(ctx context.Context) *zerolog.Event {
	span := trace.SpanFromContext(ctx)
	spanCtx := span.SpanContext()

	if spanCtx.IsValid() {
		return l.Info().
			Str("trace_id", spanCtx.TraceID().String()).
			Str("span_id", spanCtx.SpanID().String())
	}

	return l.Info()
}

// ErrorWithTrace logs errors with trace context
func (l *Logger) ErrorWithTrace(ctx context.Context) *zerolog.Event {
	span := trace.SpanFromContext(ctx)
	spanCtx := span.SpanContext()

	if spanCtx.IsValid() {
		return l.Error().
			Str("trace_id", spanCtx.TraceID().String()).
			Str("span_id", spanCtx.SpanID().String())
	}

	return l.Error()
}

// WarnWithTrace logs warnings with trace context
func (l *Logger) WarnWithTrace(ctx context.Context) *zerolog.Event {
	span := trace.SpanFromContext(ctx)
	spanCtx := span.SpanContext()

	if spanCtx.IsValid() {
		return l.Warn().
			Str("trace_id", spanCtx.TraceID().String()).
			Str("span_id", spanCtx.SpanID().String())
	}

	return l.Warn()
}
