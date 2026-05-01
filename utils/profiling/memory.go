package profiling

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"
)

func WriteHeapProfile(path, serviceName string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("create profile dir: %w", err)
	}
	fullPath := filepath.Join(path, fmt.Sprintf("%s-heap-%s.prof",
		serviceName, time.Now().UTC().Format("20060102-150405")))
	f, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("create heap profile: %w", err)
	}
	defer f.Close()

	runtime.GC() // get up-to-date stats
	if err := pprof.WriteHeapProfile(f); err != nil {
		return fmt.Errorf("write heap profile: %w", err)
	}
	return nil
}
