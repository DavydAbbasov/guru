package http

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"guru/backend/notification/internal/config"
	"guru/utils/logger"
	"guru/utils/metrics"
)

type Server struct {
	app        *fiber.App
	addr       string
	log        logger.Logger
	shutdowner fx.Shutdowner
}

func NewServer(
	cfg *config.ServerConfig,
	m *metrics.Metrics,
	log logger.Logger,
	shutdowner fx.Shutdowner,
) *Server {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	app.Get("/health/live", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	app.Get("/health/ready", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ready"})
	})
	app.Get("/metrics", metrics.PrometheusHandler(m))

	return &Server{
		app:        app,
		addr:       fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		log:        log,
		shutdowner: shutdowner,
	}
}

func RunServer(lc fx.Lifecycle, s *Server) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				s.log.Info("starting admin HTTP server", zap.String("addr", s.addr))
				if err := s.app.Listen(s.addr); err != nil && !errors.Is(err, net.ErrClosed) {
					s.log.Error("admin HTTP server error", zap.Error(err))
					if shErr := s.shutdowner.Shutdown(); shErr != nil {
						s.log.Error("failed to request fx shutdown", zap.Error(shErr))
					}
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			s.log.Info("shutting down admin HTTP server")
			return s.app.ShutdownWithContext(ctx)
		},
	})
}
