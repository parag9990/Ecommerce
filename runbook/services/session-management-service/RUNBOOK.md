# Session Management Service Local Runbook

## 1. Purpose

Owns session/event ingestion, live sessions, journeys, funnels, heatmaps, cohorts, reports, privacy controls, and retention cleanup.

## 2. Location

Actual service path: `backend/services/session-service`

## 3. Tech Stack

Go `1.26.3`, HTTP, MongoDB, Redis, analytics/reporting handlers, optional retention worker.

## 4. Required Dependencies

MongoDB `session_db`, Redis DB 5, API Gateway clients, session analytics dashboard.

## 5. Environment Variables

Use `backend/services/session-service/.env.example`.

Key vars: `SESSION_HTTP_ADDR`, `SESSION_ADMIN_TOKEN`, `SESSION_MONGO_URI`, `SESSION_MONGO_DATABASE`, `SESSION_REDIS_ADDR`, `SESSION_REDIS_PASSWORD`, `SESSION_IP_HASH_SALT`, retention/report/analytics settings.

## 6. Install Dependencies

```powershell
cd backend/services/session-service
go mod download
```

## 7. Database/Migration/Seed Setup

Mongo migrations live in `backend/services/session-service/migrations`.

```powershell
docker compose up migrate-session-mongodb
```

Seed sessions/events: Not found in codebase - please confirm.

## 8. Run Command

Recommended:

```powershell
docker compose up -d --build session-service
```

Optional retention worker:

```powershell
docker compose up -d --build session-retention-worker
```

Run once:

```powershell
make session-retention-once
```

Manual:

```powershell
cd backend/services/session-service
go run ./cmd/server
```

## 9. Health Check

- `http://localhost:8086/healthz`
- Compose healthcheck uses `http://localhost:8086/healthz`

## 10. Logs

```powershell
docker compose logs -f session-service
docker compose logs -f session-retention-worker
```

Check MongoDB, Redis, privacy, report, and retention messages.

## 11. Common Issues

- Dashboard shows no data until events are ingested.
- Redis live counters unavailable.
- Mongo migrations not applied.
- GeoIP is disabled by default.

## 12. Quick Verification

```powershell
curl http://localhost:8086/healthz
```
