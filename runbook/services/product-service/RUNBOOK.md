# Product Service Local Runbook

## 1. Purpose

Owns product catalog, categories, seller product workflow, inventory reservations, product reads, and product events.

## 2. Location

`backend/services/product-service`

## 3. Tech Stack

Go `1.26.3`, HTTP, gRPC, MongoDB, RabbitMQ/Kafka event publishing.

## 4. Required Dependencies

MongoDB `product_db`, RabbitMQ locally, Kafka optional, API consumers such as cart/search/order/CMS.

## 5. Environment Variables

Use `backend/services/product-service/.env.example`.

Key vars: `PRODUCT_HTTP_ADDR`, `PRODUCT_GRPC_ADDR`, `PRODUCT_MONGO_URI`, `PRODUCT_MONGO_DATABASE`, `PRODUCT_EVENTS_ENABLED`, `PRODUCT_EVENT_BROKER`, `RABBITMQ_URL`, `PRODUCT_INTERNAL_SERVICE_TOKEN`.

## 6. Install Dependencies

```powershell
cd backend/services/product-service
go mod download
```

## 7. Database/Migration/Seed Setup

Mongo migrations live in `backend/services/product-service/migrations/mongo`.

```powershell
docker compose up migrate-mongodb
```

Product seed data: Not found in codebase - please confirm.

## 8. Run Command

Recommended:

```powershell
docker compose up -d --build product-service
```

Manual:

```powershell
cd backend/services/product-service
go run ./cmd/server
```

`cmd/local-server` exists, but it is a runtime shell and not the full business service.

## 9. Health Check

- `http://localhost:8082/healthz`
- `http://localhost:8082/readyz`
- gRPC host port from compose: `localhost:9093`

## 10. Logs

```powershell
docker compose logs -f product-service
```

Check MongoDB, outbox worker, RabbitMQ, inventory, and internal token messages.

## 11. Common Issues

- Mongo replica set not initialized.
- RabbitMQ unavailable while product events are enabled.
- Search index remains empty until product events or reindex flow succeeds.
- Internal calls fail when service tokens mismatch.

## 12. Quick Verification

```powershell
curl http://localhost:8082/readyz
```
