package http

import (
	"runtime/debug"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"guru/utils/logger"
)

const requestIDHeader = "X-Request-ID"

const requestIDLocalKey = "request_id"

func requestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		rid := c.Get(requestIDHeader)
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Locals(requestIDLocalKey, rid)
		c.Set(requestIDHeader, rid)
		c.SetUserContext(logger.ContextWithRequestID(c.UserContext(), rid))
		return c.Next()
	}
}

// loggingMiddleware logs after the handler chain so status and duration are final.
func loggingMiddleware(log logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		status := c.Response().StatusCode()
		entry := logger.FromContext(c.UserContext(), log)
		fields := []zap.Field{
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", status),
			zap.Duration("duration", time.Since(start)),
		}
		switch {
		case status >= 500:
			entry.Error("http request", fields...)
		case status >= 400:
			entry.Warn("http request", fields...)
		default:
			entry.Info("http request", fields...)
		}
		return err
	}
}

func recoverMiddleware(log logger.Logger) fiber.Handler {
	return recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e any) {
			log.Error("panic recovered",
				zap.Any("err", e),
				zap.ByteString("stack", debug.Stack()),
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
			)
		},
	})
}
