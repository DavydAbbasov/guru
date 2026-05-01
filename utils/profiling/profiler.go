package profiling

import (
	"io"

	"go.uber.org/zap"

	"guru/utils/logger"
)

type Config struct {
	CPU         bool
	Memory      bool
	Path        string
	ServiceName string
}
type Profiler struct {
	cfg    *Config
	log    logger.Logger
	cpuOut io.Closer
}

func New(cfg *Config, log logger.Logger) *Profiler {
	return &Profiler{cfg: cfg, log: log}
}

func (p *Profiler) Start() {
	if p.cfg.CPU {
		c, err := StartCPUProfiler(p.cfg.Path, p.cfg.ServiceName)
		if err != nil {
			p.log.Error("cpu profiler start failed", zap.Error(err))
			return
		}
		p.cpuOut = c
		p.log.Info("cpu profiler started",
			zap.String("path", p.cfg.Path),
			zap.String("service", p.cfg.ServiceName))
	}
}

func (p *Profiler) Stop() {
	if p.cfg.CPU && p.cpuOut != nil {
		StopCPUProfile()
		_ = p.cpuOut.Close()
		p.log.Info("cpu profiler stopped")
	}
	if p.cfg.Memory {
		if err := WriteHeapProfile(p.cfg.Path, p.cfg.ServiceName); err != nil {
			p.log.Error("heap profile write failed", zap.Error(err))
			return
		}
		p.log.Info("heap profile written",
			zap.String("path", p.cfg.Path),
			zap.String("service", p.cfg.ServiceName))
	}
}
