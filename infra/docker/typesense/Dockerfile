# syntax=docker/dockerfile:1.7

FROM golang:1.26.4-alpine AS healthcheck-build

WORKDIR /src
COPY infra/docker/typesense/http-healthcheck.go .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/http-healthcheck ./http-healthcheck.go

FROM typesense/typesense:29.0
COPY --from=healthcheck-build /out/http-healthcheck /usr/local/bin/http-healthcheck
