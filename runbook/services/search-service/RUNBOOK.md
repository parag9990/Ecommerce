# Search Service Local Runbook

## 1. Purpose

Owns product search, autocomplete, synonym/admin search APIs, product indexing, zero-result tracking, and reindex jobs.

## 2. Location

`backend/services/search-service`

## 3. Tech Stack

Go `1.26.3`, HTTP, gRPC, Typesense, Redis, RabbitMQ, product/session HTTP clients.

## 4. Required Dependencies

Typesense, product service, Redis DB 4, RabbitMQ, optional session service for tracking.

## 5. Environment Variables

Use `backend/services/search-service/.env.example`.

Key vars: `SEARCH_HTTP_ADDR`, `SEARCH_GRPC_ADDR`, `TYPESENSE_HOST`, `TYPESENSE_PORT`, `TYPESENSE_API_KEY`, `PRODUCT_SERVICE_URL`, `RABBITMQ_URL`, `SEARCH_REDIS_ADDR`.

## 6. Install Dependencies

```powershell
cd backend/services/search-service
go mod download
```

## 7. Database/Migration/Seed Setup

No SQL/Mongo migration found. Typesense collections are managed by service/reindex code.

Product seed/index data: Not found in codebase - please confirm.

## 8. Run Command

Recommended:

```powershell
docker compose up -d --build search-service
```

Manual:

```powershell
cd backend/services/search-service
go run ./cmd/server
```

Optional reindex commands:

```powershell
cd backend/services/search-service
go run ./cmd/reindex
go run ./cmd/reindex-alias
```

## 9. Health Check

- `http://localhost:8085/healthz`
- `http://localhost:8085/readyz`
- gRPC host port: `localhost:50055`

## 10. Logs

```powershell
docker compose logs -f search-service
```

Check Typesense, product export, RabbitMQ indexer, Redis cache, and reindex logs.

## 11. Common Issues

- Typesense health fails.
- Search index is empty because products are not seeded or reindexed.
- Product service export path fails.
- RabbitMQ product event consumer is not receiving events.

## 12. Quick Verification

```powershell
curl http://localhost:8085/readyz
curl http://localhost:8108/health
```
