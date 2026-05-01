package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	logger *zap.Logger
}

func NewZapLogger(level zapcore.Level) (Logger, error) {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(level)

	instance, err := config.Build(
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, err
	}

	// callers using zap.S()/zap.L() rely on this global being replaced
	zap.ReplaceGlobals(instance)

	return &ZapLogger{
		logger: instance,
	}, nil
}

func (z *ZapLogger) Info(message string, fields ...zap.Field) {
	z.logger.Info(message, fields...)
}

func (z *ZapLogger) Debug(message string, fields ...zap.Field) {
	z.logger.Debug(message, fields...)
}

func (z *ZapLogger) Warn(message string, fields ...zap.Field) {
	z.logger.Warn(message, fields...)
}

func (z *ZapLogger) Error(message string, fields ...zap.Field) {
	z.logger.Error(message, fields...)
}

func (z *ZapLogger) Fatal(message string, fields ...zap.Field) {
	z.logger.Fatal(message, fields...)
}

func (z *ZapLogger) With(fields ...zap.Field) Logger {
	return &ZapLogger{logger: z.logger.With(fields...)}
}

func (z *ZapLogger) Sync() error {
	return z.logger.Sync()
}
