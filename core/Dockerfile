# syntax=docker/dockerfile:1
# One image for the core: the api, worker and migrate binaries.
# Compose and production select the binary with "command".
FROM golang:1.27.4-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
# Cache mounts keep the module and build caches between builds.
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
# The version goes into the binaries (api -version) and into Sentry as the release.
ARG VERSION=dev
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath \
    -ldflags="-s -w -X github.com/stilnar92/modern-project-template/core/internal/platform/version.Version=${VERSION}" \
    -o /out/ ./cmd/api ./cmd/worker ./cmd/migrate

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/ /app/
# 65532 is the nonroot user of the distroless image. A numeric ID works also
# when the host cannot resolve the user name.
USER 65532:65532
EXPOSE 8080
CMD ["/app/api"]
