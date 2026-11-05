// Package health serves the liveness endpoint used by Docker and the proxy.
package health

import (
	"context"

	"github.com/stilnar92/modern-project-template/core/internal/platform/openapi"
)

// Server implements the getHealth operation of the contract.
type Server struct{}

// GetHealth reports that the process is alive.
func (Server) GetHealth(context.Context, openapi.GetHealthRequestObject) (openapi.GetHealthResponseObject, error) {
	return openapi.GetHealth200JSONResponse{Status: openapi.Ok}, nil
}
