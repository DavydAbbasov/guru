package profiling

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/pprof"
	"time"
)

func StartCPUProfiler(path, serviceName string) (io.Closer, error) {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return nil, fmt.Errorf("create profile dir: %w", err)
	}
	fullPath := filepath.Join(path, fmt.Sprintf("%s-cpu-%s.prof",
		serviceName, time.Now().UTC().Format("20060102-150405")))
	f, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("create cpu profile: %w", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("start cpu profile: %w", err)
	}
	return f, nil
}

func StopCPUProfile() {
	pprof.StopCPUProfile()
}
