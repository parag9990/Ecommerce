# API Gateway Local Runbook

## 1. Purpose

Edge service for REST traffic, gRPC-Web, auth enforcement, CORS, rate limiting, routing, metrics, and tracing.

## 2. Location

`backend/services/api-gateway`

## 3. Tech Stack

Go `1.26.3`, HTTP, gRPC-Web, Redis-backed rate limiting, JWT/JWKS validation, OpenTelemetry, Prometheus.

## 4. Required Dependencies

Redis, auth service/JWKS, downstream backend services, OTEL Collector when tracing is enabled.

## 5. Environment Variables

Use `backend/services/api-gateway/.env.example`.

Key vars: `HTTP_ADDR`, `GRPC_WEB_ENABLED`, `GRPC_ADDR`, `METRICS_ADDR`, `CORS_ALLOWED_ORIGINS`, downstream `*_ADDR`/`*_URL`, `REDIS_ADDR`, `JWT_JWKS_URL`, `TRACE_EXPORTER_OTLP_ENDPOINT`, `PAYMENT_INTERNAL_API_TOKEN`.

## 6. Install Dependencies

```powershell
cd backend/services/api-gateway
go mod download
```

## 7. Database/Migration/Seed Setup

No service-owned database or migration found. Redis is required for rate limiting when enabled.

## 8. Run Command

Recommended:

```powershell
docker compose up -d --build api-gateway
```

Manual:

```powershell
cd backend/services/api-gateway
go run ./cmd/server
```

For manual host runs, replace Docker service DNS names in env values with `localhost`.

## 9. Health Check

- `http://localhost:8080/health/live`
- `http://localhost:8080/health/ready`
- gRPC-Web host port: `http://localhost:8099`
- Metrics host port: `http://localhost:19090`

## 10. Logs

Docker logs:

```powershell
docker compose logs -f api-gateway
```

Check route, JWT, CORS, Redis, tracing, and downstream health messages.

## 11. Common Issues

- Readiness fails while downstream services are still starting.
- CORS rejects a manual frontend port not listed in `CORS_ALLOWED_ORIGINS`.
- JWT validation fails when auth JWKS, issuer, or audience do not match.
- `CMS_INTERNAL_AUTH_TOKEN` is blank in local examples.

## 12. Quick Verification

```powershell
curl http://localhost:8080/health/live
curl http://localhost:8080/health/ready
```
