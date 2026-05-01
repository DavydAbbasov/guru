package logger

import "go.uber.org/zap"

type Logger interface {
	Info(message string, fields ...zap.Field)
	Debug(message string, fields ...zap.Field)
	Warn(message string, fields ...zap.Field)
	Error(message string, fields ...zap.Field)
	Fatal(message string, fields ...zap.Field)
	With(fields ...zap.Field) Logger
	Sync() error
}
