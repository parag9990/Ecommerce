# syntax=docker/dockerfile:1.7
ARG GO_VERSION=1.26.4
FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src
COPY backend/shared ./backend/shared
COPY backend/proto-gen ./backend/proto-gen
COPY backend/services/api-gateway ./backend/services/api-gateway
WORKDIR /src/backend/services/api-gateway
RUN --mount=type=cache,target=/go/pkg/mod go mod download
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/service ./cmd/server

FROM alpine:3.21
RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=build /out/service /app/service
COPY api/master-api.json /app/config/master-api.json
COPY backend/services/api-gateway/config/grpcweb-policies.json /app/config/grpcweb-policies.json
USER app
EXPOSE 8080 9090 9091
ENTRYPOINT ["/app/service"]
