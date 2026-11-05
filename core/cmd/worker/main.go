// Command worker runs background jobs of the core. River queues are added
// when the first job appears; until then the process only waits for a signal.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/stilnar92/modern-project-template/core/internal/platform/observability"
)

func main() {
	logger, flush, err := observability.Setup("worker")
	if err != nil {
		slog.Error("setup", "err", err)
		os.Exit(1)
	}
	defer flush()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("worker started")
	<-ctx.Done()
	logger.Info("worker stopped")
}
