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
	"strconv"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/stilnar92/modern-project-template/core/internal/platform/config"
	"github.com/stilnar92/modern-project-template/core/internal/platform/health"
	"github.com/stilnar92/modern-project-template/core/internal/platform/httpmw"
	"github.com/stilnar92/modern-project-template/core/internal/platform/observability"
	"github.com/stilnar92/modern-project-template/core/internal/platform/openapi"
	"github.com/stilnar92/modern-project-template/core/internal/platform/problem"
	"github.com/stilnar92/modern-project-template/core/internal/platform/version"
)

// apiConfig is the environment of the api (.env.example).
type apiConfig struct {
	config.Base
	Port        int    `env:"PORT" envDefault:"8080"`
	DatabaseURL string `env:"DATABASE_URL,required,notEmpty"`
}

// Server timeouts. A slow client cannot hold a connection for ever.
// WriteTimeout limits each response: a future streaming endpoint (SSE)
// must extend its own deadline with http.ResponseController.
const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 120 * time.Second
	shutdownTimeout   = 10 * time.Second
)

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	healthcheck := flag.Bool("healthcheck", false,
		"call /healthz of the running api and exit with 0 when it answers 200 (the Docker healthcheck: the image has no shell)")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.Version)
		return
	}
	if *healthcheck {
		os.Exit(runHealthcheck())
	}
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// runHealthcheck needs only PORT, so it does not load the whole config.
func runHealthcheck() int {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := health.Check(ctx, "http://127.0.0.1:"+port+"/healthz"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func run() error {
	cfg, err := config.Load[apiConfig]()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	tel, err := observability.Setup(ctx, "api", cfg.Base)
	if err != nil {
		return err
	}
	defer tel.Shutdown(context.WithoutCancel(ctx))
	logger := tel.Logger

	// The pool connects on first use. If the database is down, the api still
	// starts, and /readyz reports it.
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("database pool: %w", err)
	}
	defer pool.Close()

	srv := &http.Server{
		Addr:              ":" + strconv.Itoa(cfg.Port),
		Handler:           newHandler(logger, health.Server{DB: pool, Logger: logger}),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
	metrics := tel.MetricsServer(cfg.MetricsPort, nil)

	errc := make(chan error, 2)
	for _, s := range []*http.Server{srv, metrics} {
		go func() {
			logger.InfoContext(ctx, "listening", "addr", s.Addr)
			if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errc <- fmt.Errorf("listen %s: %w", s.Addr, err)
			}
		}()
	}

	var runErr error
	select {
	case <-ctx.Done():
	case runErr = <-errc:
		observability.ReportError(ctx, logger, "api stopped", runErr)
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()
	if err := errors.Join(srv.Shutdown(shutdownCtx), metrics.Shutdown(shutdownCtx)); err != nil {
		observability.ReportError(shutdownCtx, logger, "api shutdown", err)
		return errors.Join(runErr, err)
	}
	logger.InfoContext(shutdownCtx, "api stopped")
	return runErr
}

// newHandler serves each operation of api/openapi.yaml through the generated
// router. A missing operation is a compile error. Each error, also from the
// router, is a Problem response (backend standard 02). The middleware order
// is in the httpmw package comment.
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
	h := crossOrigin().Handler(problem.Routes(mux))
	h = httpmw.Recover(logger)(h)
	h = httpmw.AccessLog(logger)(h)
	h = httpmw.RequestID(h)
	return otelhttp.NewHandler(h, "api")
}

// crossOrigin blocks a state-changing request (POST, PUT, PATCH, DELETE)
// that a browser sends from another site (CSRF, backend standard 05,
// ADR 0012). It reads Sec-Fetch-Site, or compares Origin with Host. The
// session cookie is SameSite=Lax as a second layer. Webhooks come from
// servers of other services, so they skip the check: their signature
// protects them.
func crossOrigin() *http.CrossOriginProtection {
	cop := http.NewCrossOriginProtection()
	cop.AddInsecureBypassPattern("/v1/webhooks/")
	cop.SetDenyHandler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		problem.Write(w, http.StatusForbidden, "The request came from another site.")
	}))
	return cop
}
