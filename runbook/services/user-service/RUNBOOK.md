# User Service Local Runbook

## 1. Purpose

Owns user profiles, addresses, seller profile/KYC data, seller status admin APIs, and user events.

## 2. Location

`backend/services/user-service`

## 3. Tech Stack

Go `1.26.3`, gRPC, admin HTTP, MySQL, RabbitMQ/Kafka event publishing.

## 4. Required Dependencies

MySQL `user_db`, RabbitMQ for local event outbox, optional Kafka config.

## 5. Environment Variables

Use `backend/services/user-service/.env.example`.

Key vars: `USER_SERVICE_GRPC_ADDRESS`, `USER_SERVICE_HTTP_ADDRESS`, `USER_SERVICE_DATABASE_DSN`, `USER_SERVICE_ADMIN_TOKEN`, `USER_EVENTS_PROVIDER`, `RABBITMQ_URL`, `KAFKA_BROKERS`.

## 6. Install Dependencies

```powershell
cd backend/services/user-service
go mod download
```

## 7. Database/Migration/Seed Setup

Migrations live in `backend/services/user-service/migrations`.

```powershell
docker compose up migrate-user
```

Seed users: Not found in codebase - please confirm.

## 8. Run Command

Recommended:

```powershell
docker compose up -d --build user-service
```

Manual:

```powershell
cd backend/services/user-service
go run ./cmd/server
```

## 9. Health Check

- gRPC host port: `localhost:50052`
- Admin/health URL inside compose: `http://localhost:9091/health/ready`
- WARNING: `docker-compose.yml` does not publish admin HTTP port `9091` to the host.

## 10. Logs

```powershell
docker compose logs -f user-service
```

Check DB, admin token, outbox worker, and RabbitMQ messages.

## 11. Common Issues

- Admin token too short or missing.
- MySQL migration not applied.
- RabbitMQ URL/vhost mismatch.
- Gateway gRPC connection fails if port or address differs.

## 12. Quick Verification

```powershell
docker compose exec user-service wget -qO- http://localhost:9091/health/live
docker compose exec user-service wget -qO- http://localhost:9091/health/ready
```

For a manual host run, use `curl http://localhost:9091/health/ready`.
