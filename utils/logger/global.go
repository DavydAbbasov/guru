package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Init(level zapcore.Level) error {
	_, err := NewZapLogger(level)
	return err
}

// InitFromString falls back to InfoLevel on parse error rather than returning it.
func InitFromString(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		lvl = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}
	return Init(lvl.Level())
}

func Info(message string, fields ...zap.Field) {
	zap.L().Info(message, fields...)
}

func Debug(message string, fields ...zap.Field) {
	zap.L().Debug(message, fields...)
}

func Warn(message string, fields ...zap.Field) {
	zap.L().Warn(message, fields...)
}

func Error(message string, fields ...zap.Field) {
	zap.L().Error(message, fields...)
}

func Fatal(message string, fields ...zap.Field) {
	zap.L().Fatal(message, fields...)
}

func Sync() error {
	return zap.L().Sync()
}
