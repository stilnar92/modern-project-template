// Command api runs the HTTP API of the core.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stilnar92/modern-project-template/core/internal/platform/config"
	"github.com/stilnar92/modern-project-template/core/internal/platform/health"
	"github.com/stilnar92/modern-project-template/core/internal/platform/observability"
	"github.com/stilnar92/modern-project-template/core/internal/platform/openapi"
	"github.com/stilnar92/modern-project-template/core/internal/platform/problem"
	"github.com/stilnar92/modern-project-template/core/internal/platform/version"
)

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	healthcheck := flag.Bool("healthcheck", false,
		"call /healthz of the running api and exit with 0 when it answers 200 (the Docker healthcheck: the image has no shell)")
	flag.Parse()
	port := config.String("PORT", "8080")

	if *showVersion {
		fmt.Println(version.Version)
		return
	}
	if *healthcheck {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := checkHealth(ctx, "http://127.0.0.1:"+port+"/healthz")
		cancel()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	logger, flush, err := observability.Setup("api")
	if err != nil {
		slog.Error("setup", "err", err)
		os.Exit(1)
	}
	defer flush()

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           newHandler(logger, health.Server{}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("api started", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			observability.ReportError(logger, "api stopped", err)
			flush()
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		observability.ReportError(logger, "api shutdown", err)
		flush()
		os.Exit(1)
	}
	logger.Info("api stopped")
}

// newHandler serves each operation of api/openapi.yaml through the generated
// router. A missing operation is a compile error. Each error, also from the
// router, is a Problem response (backend standard 02).
func newHandler(logger *slog.Logger, server openapi.StrictServerInterface) http.Handler {
	strict := openapi.NewStrictHandlerWithOptions(server, nil, openapi.StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  problem.RequestError(logger),
		ResponseErrorHandlerFunc: problem.ServerError(logger),
	})
	mux := http.NewServeMux()
	openapi.HandlerWithOptions(strict, openapi.StdHTTPServerOptions{
		BaseRouter:       mux,
		ErrorHandlerFunc: problem.RequestError(logger),
	})
	return problem.Routes(mux)
}

// checkHealth calls the health endpoint and returns an error unless it
// answers 200.
func checkHealth(ctx context.Context, url string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }() // only the status matters
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check: status %d", resp.StatusCode)
	}
	return nil
}
