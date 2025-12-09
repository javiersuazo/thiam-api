package logger

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

const (
	levelDebug = "debug"
	levelInfo  = "info"
	levelWarn  = "warn"
	levelError = "error"
)

type ctxKey struct{}

var requestIDKey = ctxKey{} //nolint:gochecknoglobals // used as context key

type Interface interface {
	Debug(message interface{}, args ...interface{})
	Info(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Error(message interface{}, args ...interface{})
	Fatal(message interface{}, args ...interface{})
	WithField(key string, value interface{}) Interface
	WithFields(fields map[string]interface{}) Interface
	WithRequestID(requestID string) Interface
	WithContext(ctx context.Context) Interface
}

type Logger struct {
	logger *zerolog.Logger
}

var _ Interface = (*Logger)(nil)

func New(level string) *Logger {
	var l zerolog.Level

	switch strings.ToLower(level) {
	case levelError:
		l = zerolog.ErrorLevel
	case levelWarn:
		l = zerolog.WarnLevel
	case levelInfo:
		l = zerolog.InfoLevel
	case levelDebug:
		l = zerolog.DebugLevel
	default:
		l = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(l)

	skipFrameCount := 3
	logger := zerolog.New(os.Stdout).With().Timestamp().CallerWithSkipFrameCount(zerolog.CallerSkipFrameCount + skipFrameCount).Logger()

	return &Logger{
		logger: &logger,
	}
}

func (l *Logger) Debug(message interface{}, args ...interface{}) {
	l.msg("debug", message, args...)
}

func (l *Logger) Info(message string, args ...interface{}) {
	l.log("info", message, args...)
}

func (l *Logger) Warn(message string, args ...interface{}) {
	l.log("warn", message, args...)
}

func (l *Logger) Error(message interface{}, args ...interface{}) {
	if l.logger.GetLevel() == zerolog.DebugLevel {
		l.Debug(message, args...)
	}

	l.msg("error", message, args...)
}

func (l *Logger) Fatal(message interface{}, args ...interface{}) {
	l.msg("fatal", message, args...)

	os.Exit(1)
}

func (l *Logger) WithField(key string, value interface{}) Interface {
	newLogger := l.logger.With().Interface(key, value).Logger()

	return &Logger{logger: &newLogger}
}

func (l *Logger) WithFields(fields map[string]interface{}) Interface {
	ctx := l.logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}

	newLogger := ctx.Logger()

	return &Logger{logger: &newLogger}
}

func (l *Logger) WithRequestID(requestID string) Interface {
	newLogger := l.logger.With().Str("request_id", requestID).Logger()

	return &Logger{logger: &newLogger}
}

func (l *Logger) WithContext(ctx context.Context) Interface {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok && requestID != "" {
		return l.WithRequestID(requestID)
	}

	return l
}

func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}

	return ""
}

func (l *Logger) log(level, message string, args ...interface{}) {
	var event *zerolog.Event

	switch level {
	case levelDebug:
		event = l.logger.Debug()
	case levelWarn:
		event = l.logger.Warn()
	case levelError:
		event = l.logger.Error()
	default:
		event = l.logger.Info()
	}

	if len(args) == 0 {
		event.Msg(message)
	} else {
		event.Msgf(message, args...)
	}
}

func (l *Logger) msg(level string, message interface{}, args ...interface{}) {
	switch msg := message.(type) {
	case error:
		l.log(level, msg.Error(), args...)
	case string:
		l.log(level, msg, args...)
	default:
		l.log(level, fmt.Sprintf("%s message %v has unknown type %v", level, message, msg), args...)
	}
}
