package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberPprof "github.com/gofiber/fiber/v2/middleware/pprof"
	fiberSwagger "github.com/swaggo/fiber-swagger"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"guru/backend/products/internal/config"
	"guru/backend/products/internal/transport/http/handlers"
	"guru/utils/logger"
	"guru/utils/metrics"
)

const (
	fallbackReadTimeout  = 10 * time.Second
	fallbackWriteTimeout = 10 * time.Second
	fallbackIdleTimeout  = 30 * time.Second
	fallbackBodyLimit    = 4 * 1024 * 1024
)

type Server struct {
	app        *fiber.App
	addr       string
	log        logger.Logger
	shutdowner fx.Shutdowner
}

type Params struct {
	fx.In

	Config     *config.ServerConfig
	Handlers   []handlers.Handler `group:"handlers"`
	Metrics    *metrics.Metrics
	Log        logger.Logger
	Shutdowner fx.Shutdowner
}

func NewServer(p Params) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:  durationOr(p.Config.ReadTimeout, fallbackReadTimeout),
		WriteTimeout: durationOr(p.Config.WriteTimeout, fallbackWriteTimeout),
		IdleTimeout:  durationOr(p.Config.IdleTimeout, fallbackIdleTimeout),
		BodyLimit:    intOr(p.Config.BodyLimit, fallbackBodyLimit),
	})

	app.Use(requestIDMiddleware())
	app.Use(recoverMiddleware(p.Log))
	app.Use(cors.New())
	app.Use(loggingMiddleware(p.Log))
	app.Use(p.Metrics.FiberMiddleware())

	app.Get("/health/live", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	app.Get("/health/ready", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ready"})
	})

	app.Get("/metrics", metrics.PrometheusHandler(p.Metrics))
	app.Get("/docs/swagger/*", fiberSwagger.WrapHandler)
	app.Use(fiberPprof.New()) // mounts /debug/pprof/* — keep behind a private LB in prod

	api := app.Group("/api/v1")
	for _, h := range p.Handlers {
		if h == nil {
			continue
		}
		h.Register(api)
	}

	return &Server{
		app:        app,
		addr:       fmt.Sprintf("%s:%d", p.Config.Host, p.Config.Port),
		log:        p.Log,
		shutdowner: p.Shutdowner,
	}
}

func durationOr(v, fallback time.Duration) time.Duration {
	if v > 0 {
		return v
	}
	return fallback
}

func intOr(v, fallback int) int {
	if v > 0 {
		return v
	}
	return fallback
}

func RunServer(lc fx.Lifecycle, s *Server) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				s.log.Info("starting HTTP server",
					zap.String("addr", s.addr))
				if err := s.app.Listen(s.addr); err != nil && !errors.Is(err, net.ErrClosed) {
					s.log.Error("HTTP server error",
						zap.Error(err))
					if shErr := s.shutdowner.Shutdown(); shErr != nil {
						s.log.Error("failed to request fx shutdown",
							zap.Error(shErr))
					}
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			s.log.Info("shutting down HTTP server")
			return s.app.ShutdownWithContext(ctx)
		},
	})
}
