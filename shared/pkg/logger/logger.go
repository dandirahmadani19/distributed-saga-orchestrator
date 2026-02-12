package logger

import (
	"context"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

type Logger struct {
	zerolog.Logger
}

func New(serviceName string) *Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	log := zerolog.New(
		zerolog.ConsoleWriter{Out: os.Stdout},
	).With().
		Timestamp().
		Str("service", serviceName).
		Logger()

	return &Logger{Logger: log}
}

func NewProduction(serviceName string) *Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	log := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", serviceName).
		Logger()

	return &Logger{Logger: log}
}

func SetLogLevel(level string) {
	switch strings.ToLower(level) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

func (l *Logger) withTrace(ctx context.Context, event *zerolog.Event) *zerolog.Event {
	if ctx == nil {
		return event
	}

	span := trace.SpanFromContext(ctx)
	spanCtx := span.SpanContext()

	if spanCtx.IsValid() {
		event = event.
			Str("trace_id", spanCtx.TraceID().String()).
			Str("span_id", spanCtx.SpanID().String())
	}

	return event
}

func (l *Logger) DebugWithTrace(ctx context.Context) *zerolog.Event {
	return l.withTrace(ctx, l.Debug())
}

func (l *Logger) InfoWithTrace(ctx context.Context) *zerolog.Event {
	return l.withTrace(ctx, l.Info())
}

func (l *Logger) WarnWithTrace(ctx context.Context) *zerolog.Event {
	return l.withTrace(ctx, l.Warn())
}

func (l *Logger) ErrorWithTrace(ctx context.Context) *zerolog.Event {
	return l.withTrace(ctx, l.Error())
}

func (l *Logger) FatalWithTrace(ctx context.Context) *zerolog.Event {
	return l.withTrace(ctx, l.Fatal())
}

func (l *Logger) PanicWithTrace(ctx context.Context) *zerolog.Event {
	return l.withTrace(ctx, l.Panic())
}

func (l *Logger) ErrorWithStack(ctx context.Context, err error) *zerolog.Event {
	return l.withTrace(ctx, l.Error().Err(err))
}

func (l *Logger) WithField(key string, value interface{}) *Logger {
	child := l.Logger.With().
		Interface(key, value).
		Logger()

	return &Logger{Logger: child}
}

func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	ctx := l.Logger.With()

	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}

	child := ctx.Logger()
	return &Logger{Logger: child}
}
