# syntax=docker/dockerfile:1.7
ARG GO_VERSION=1.26.3
FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src
COPY backend/shared ./backend/shared
COPY backend/proto-gen ./backend/proto-gen
COPY backend/services/cart-service ./backend/services/cart-service
WORKDIR /src/backend/services/cart-service
RUN --mount=type=cache,target=/go/pkg/mod go mod download
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/service ./cmd/server && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/cart-expiry-worker ./cmd/cart-expiry-worker
FROM alpine:3.21
RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=build /out/service /app/service
COPY --from=build /out/cart-expiry-worker /app/cart-expiry-worker
USER app
EXPOSE 8083
ENTRYPOINT ["/app/service"]
