// Package health serves the system checks: liveness (/healthz) for Docker
// and readiness (/readyz) for the proxy and an orchestrator.
package health

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/stilnar92/modern-project-template/core/internal/platform/openapi"
)

// pingTimeout keeps /readyz fast when the database hangs.
const pingTimeout = 2 * time.Second

// Check calls a liveness URL and returns nil when it answers 200. The
// -healthcheck flag of the api and the worker uses it: the Docker image has
// no shell and no curl.
func Check(ctx context.Context, url string) error {
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

// Server implements the system operations of the contract.
type Server struct {
	DB     *pgxpool.Pool
	Logger *slog.Logger
}

// GetHealth reports that the process is alive. It does not check the
// database: a database failure must not make Docker restart the api.
func (Server) GetHealth(context.Context, openapi.GetHealthRequestObject) (openapi.GetHealthResponseObject, error) {
	return openapi.GetHealth200JSONResponse{Status: openapi.Ok}, nil
}

// GetReady reports that the process can serve requests: the database
// answers. The reason of a failure goes only to the log.
func (s Server) GetReady(ctx context.Context, _ openapi.GetReadyRequestObject) (openapi.GetReadyResponseObject, error) {
	ctx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()
	if err := s.DB.Ping(ctx); err != nil {
		s.Logger.WarnContext(ctx, "not ready: the database does not answer", "err", err)
		detail := "The service is not ready."
		return openapi.GetReady503ApplicationProblemPlusJSONResponse{
			ProblemApplicationProblemPlusJSONResponse: openapi.ProblemApplicationProblemPlusJSONResponse{
				Type:   openapi.AboutBlank,
				Title:  http.StatusText(http.StatusServiceUnavailable),
				Status: http.StatusServiceUnavailable,
				Detail: &detail,
			},
		}, nil
	}
	return openapi.GetReady200JSONResponse{Status: openapi.Ok}, nil
}
