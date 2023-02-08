package apm

import (
	"go.uber.org/zap"
)

// Logging provides structured logging methods.
// All methods are safe for concurrent use.
// It is the main interface implemented by zap.Logger
//
// Deprecated: use *zap.Logger directly instead.
type Logging interface {
	Debug(string, ...zap.Field)
	Error(string, ...zap.Field)
	Fatal(string, ...zap.Field)
	Info(string, ...zap.Field)
	Panic(string, ...zap.Field)
	Warn(string, ...zap.Field)
}
