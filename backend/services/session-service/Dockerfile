# syntax=docker/dockerfile:1.7
ARG GO_VERSION=1.26.4
FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src
COPY backend/shared ./backend/shared
COPY backend/proto-gen ./backend/proto-gen
COPY backend/services/session-service ./backend/services/session-service
WORKDIR /src/backend/services/session-service
RUN --mount=type=cache,target=/go/pkg/mod go mod download
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/service ./cmd/server && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/retention-worker ./cmd/retention-worker
FROM alpine:3.21
RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=build /out/service /app/service
COPY --from=build /out/retention-worker /app/retention-worker
USER app
EXPOSE 8086
ENTRYPOINT ["/app/service"]
