# syntax=docker/dockerfile:1.7
ARG GO_VERSION=1.26.4
FROM golang:${GO_VERSION}-alpine AS build
RUN apk add --no-cache openssl
WORKDIR /src
COPY backend/shared ./backend/shared
COPY backend/proto-gen ./backend/proto-gen
COPY backend/services/auth-service ./backend/services/auth-service
WORKDIR /src/backend/services/auth-service
RUN --mount=type=cache,target=/go/pkg/mod go mod download
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/service ./cmd/server
RUN mkdir -p /out/keys && openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out /out/keys/private.pem && openssl rsa -pubout -in /out/keys/private.pem -out /out/keys/public.pem && chmod 0444 /out/keys/*.pem

FROM alpine:3.21
RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=build /out/service /app/service
COPY --from=build /out/keys /app/keys
USER app
EXPOSE 8081
ENTRYPOINT ["/app/service"]
