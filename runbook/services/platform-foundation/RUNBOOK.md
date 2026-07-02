# Platform Foundation Local Runbook

## 1. Purpose

Provides local infrastructure, shared backend modules, protobuf contracts, generated clients, observability config, CI scripts, Docker Compose, and Tilt orchestration.

## 2. Location

- `docker-compose.yml`
- `Makefile`
- `Tiltfile`
- `infra/`
- `backend/shared/`
- `backend/proto-gen/`
- `proto/`
- `api/master-api.json`
- `frontend/packages/proto-client`

## 3. Tech Stack

Docker Compose, MySQL, MongoDB, Redis, RabbitMQ, Kafka, Typesense, Mailpit, Prometheus, OpenTelemetry Collector, Jaeger, Go shared modules, Buf/protobuf, pnpm workspace.

## 4. Required Dependencies

Docker, Docker Compose, Go, Node, pnpm, Buf for proto workflows.

## 5. Environment Variables

Root compose supports host port overrides such as `MYSQL_PORT`, `REDIS_PORT`, `RABBITMQ_PORT`, `RABBITMQ_MANAGEMENT_PORT`, and `TYPESENSE_HOST_PORT`.

Service env files live beside each service. Root `.env.example`: Not found in codebase - please confirm.

## 6. Install Dependencies

Validate compose:

```powershell
docker compose config
```

Frontend workspace:

```powershell
cd frontend
corepack pnpm install --frozen-lockfile
```

Backend workspace:

```powershell
cd backend
go work sync
```

Do not run `go test ./...` from `backend/`; this workspace root has `go.work` but no root `go.mod`.

## 7. Database/Migration/Seed Setup

MySQL database creation is in `infra/mysql/init/00-create-databases.sql`.

Compose migration jobs handle service migrations. Seed data: Not found in codebase - please confirm.

## 8. Run Command

All services:

```powershell
docker compose up -d --build
```

Make shortcuts:

```powershell
make infra-up
make backend-up
make frontend-up
```

`make infra-up` follows the root `Makefile` and does not directly start `init-mongodb-replica` or `otel-collector`; use the full compose command for a complete normal local run.

Tilt:

```powershell
make tilt-up
```

## 9. Health Check

```powershell
docker compose ps
curl http://localhost:8080/health/live
curl http://localhost:8108/health
```

## 10. Logs

```powershell
docker compose logs -f
```

Check infra services first, then app services.

## 11. Common Issues

- Docker daemon not running.
- Compose health dependency timeout.
- Port conflicts with local MySQL, Redis, Kafka, or frontend dev servers.
- Proto/generated clients drift if Buf generation is not run after proto edits.

## 12. Quick Verification

```powershell
docker compose config
docker compose ps
```
