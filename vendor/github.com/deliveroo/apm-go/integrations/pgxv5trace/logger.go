package pgxv5trace

import (
	"context"

	"github.com/jackc/pgx/v5/tracelog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	apmgo "github.com/deliveroo/apm-go"
)

type Logger struct {
	apm apmgo.Service
}

// NewLogger returns an apm-go integrated pgx logger.
func NewLogger(apm apmgo.Service) *Logger {
	return &Logger{apm: apm}
}

// Log implements the tracelog.Logger interface.
func (l *Logger) Log(_ context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	fields := make([]zapcore.Field, len(data))
	i := 0
	for k, v := range data {
		fields[i] = zap.Reflect(k, v)
		i++
	}

	logging := l.apm.Logger()
	switch level {
	case tracelog.LogLevelTrace:
		logging.Debug(msg, append(fields, zap.Stringer("PGX_LOG_LEVEL", level))...)
	case tracelog.LogLevelDebug:
		logging.Debug(msg, fields...)
	case tracelog.LogLevelInfo:
		logging.Info(msg, fields...)
	case tracelog.LogLevelWarn:
		logging.Warn(msg, fields...)
	case tracelog.LogLevelError:
		logging.Error(msg, fields...)
	default:
		logging.Error(msg, append(fields, zap.Stringer("PGX_LOG_LEVEL", level))...)
	}
}
