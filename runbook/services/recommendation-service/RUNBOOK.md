# Recommendation Service Local Runbook

## 1. Purpose

Owns recommendation retrieval, interaction ingestion, feature storage, ranking, personalization, and A/B assignment hooks.

## 2. Location

`backend/services/recommendation-service`

## 3. Tech Stack

Go `1.26.3`, HTTP, gRPC, MongoDB, Redis, Kafka.

## 4. Required Dependencies

MongoDB `recommendation_db`, Redis DB 3, Kafka recommendation event topics, API Gateway gRPC/gRPC-Web clients.

## 5. Environment Variables

Use `backend/services/recommendation-service/.env.example`.

Key vars: `RECOMMENDATION_HTTP_ADDR`, `RECOMMENDATION_GRPC_ADDR`, `RECOMMENDATION_MONGO_URI`, `RECOMMENDATION_REDIS_ADDR`, `QUEUE_PROVIDER`, `KAFKA_BROKERS`, `RECOMMENDATION_EVENTS_TOPIC`.

## 6. Install Dependencies

```powershell
cd backend/services/recommendation-service
go mod download
```

## 7. Database/Migration/Seed Setup

Mongo migrations live in `backend/services/recommendation-service/migrations`.

```powershell
docker compose up migrate-mongodb
```

Recommendation seed data: Not found in codebase - please confirm.

## 8. Run Command

Recommended:

```powershell
docker compose up -d --build recommendation-service
```

Manual:

```powershell
cd backend/services/recommendation-service
go run ./cmd/server
```

## 9. Health Check

- `http://localhost:8089/healthz`
- Compose healthcheck uses `http://localhost:8089/healthz`
- gRPC host port: `localhost:50058`

## 10. Logs

```powershell
docker compose logs -f recommendation-service
```

Check MongoDB, Redis, Kafka consumer, ranking, personalization, and DLQ messages.

## 11. Common Issues

- Kafka not healthy.
- Mongo feature collections/indexes missing.
- Redis cache unavailable.
- Cold-start recommendations may be sparse without product/interaction data.

## 12. Quick Verification

```powershell
curl http://localhost:8089/healthz
```
