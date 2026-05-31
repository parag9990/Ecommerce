# Superadmin Service - Environment Dependency

## 1. What Is This Dependency?

Environment variables service ko runtime configuration deti hain. Example: HTTP port, DB DSN, downstream service URLs, timeout, cache TTL.

## 2. Why This Service Uses It

Superadmin Service different environments me run ho sakti hai: local, test, staging, production. Hardcoding config unsafe hota hai, isliye code `os.Getenv` se env values read karta hai.

## 3. Required Or Optional

Required for real run. Local mode me kuch values optional hain, but DB and downstream URLs ke bina many admin features unavailable ho jayenge.

## 4. Where It Is Used In Project

| Path | Usage |
|------|-------|
| `backend/services/superadmin-service/internal/config/config.go` | Loads and validates env vars |
| `backend/services/superadmin-service/.env` | Local values found |
| `backend/services/superadmin-service/cmd/server/main.go` | Uses config to wire DB, clients, server |

## 5. Installation Steps

No package install required. You need to create/export env vars.

Current `.env` file exists:

```text
backend/services/superadmin-service/.env
```

Code does not automatically parse `.env`, so source it before running.

## 6. Docker Setup, If Possible

No compose file is clearly found.

If Docker Compose is added later, pass env using:

```yaml
env_file:
  - .env
```

Current status: Not clearly found in project files.

## 7. Local Setup Without Docker

```bash
cd backend/services/superadmin-service
set -a
source .env
set +a
go run ./cmd/server
```

Or export individual values:

```bash
export HTTP_ADDR=:8088
export SUPERADMIN_DATABASE_DSN='superadmin:change_me@tcp(127.0.0.1:3306)/superadmin_db?parseTime=true&charset=utf8mb4&loc=UTC'
```

## 8. Required Environment Variables

| Variable | Default / Example | Required? | Purpose |
|----------|-------------------|-----------|---------|
| `SERVICE_NAME` | `superadmin-service` | Yes | Logger/service name |
| `APP_ENV` | `local` | Yes | Environment mode |
| `HTTP_ADDR` | `:8088` | Yes | HTTP bind address |
| `LOG_LEVEL` | `info` / `debug` | Optional | Logging level |
| `SUPERADMIN_DATABASE_DSN` | MySQL DSN | Required for DB-backed run | MySQL connection |
| `MYSQL_DSN` | MySQL DSN | Optional fallback | Used if Superadmin DSN empty |
| `SUPERADMIN_REQUIRE_DATABASE` | `false` | Optional local, recommended true prod | Strict DB requirement |
| `USER_SERVICE_ADMIN_BASE_URL` | `http://127.0.0.1:8081/internal/admin` | Required for user/seller APIs | Downstream HTTP |
| `USER_SERVICE_TIMEOUT` | `5s` | Optional | Downstream timeout |
| `SUPERADMIN_REQUIRE_USER_SERVICE` | `false` | Optional local, recommended true prod | Strict downstream requirement |
| `ORDER_SERVICE_ADMIN_BASE_URL` | `http://127.0.0.1:8084/internal/admin` | Required for order APIs | Downstream HTTP |
| `ORDER_SERVICE_TIMEOUT` | `5s` | Optional | Downstream timeout |
| `SUPERADMIN_REQUIRE_ORDER_SERVICE` | `false` | Optional local, recommended true prod | Strict downstream requirement |
| `PAYMENT_SERVICE_ADMIN_BASE_URL` | `http://127.0.0.1:8085/internal/admin` | Required for payment APIs | Downstream HTTP |
| `PAYMENT_SERVICE_TIMEOUT` | `5s` | Optional | Downstream timeout |
| `SUPERADMIN_REQUIRE_PAYMENT_SERVICE` | `false` | Optional local, recommended true prod | Strict downstream requirement |
| `SUPERADMIN_PLATFORM_SETTINGS_CACHE_TTL` | `5m` | Optional | In-memory cache TTL |
| `SUPERADMIN_PLATFORM_SETTINGS_EVENT_TOPIC` | `platform.settings.updated` | Optional | Logged event topic |
| `SUPERADMIN_SESSION_ANALYTICS_MAX_RANGE` | `720h` | Optional | Max analytics date range |
| `SUPERADMIN_SESSION_ANALYTICS_MAX_PAGE_SIZE` | `100` | Optional | Max page size |
| `HTTP_READ_HEADER_TIMEOUT` | `5s` | Optional | HTTP hardening |
| `HTTP_SHUTDOWN_TIMEOUT` | `10s` | Optional | Graceful shutdown |

## 9. Start Commands

Suggested command based on project structure:

```bash
cd backend/services/superadmin-service
set -a
source .env
set +a
go run ./cmd/server
```

## 10. Verify Running Commands

Check exported env:

```bash
printenv SERVICE_NAME
printenv HTTP_ADDR
printenv SUPERADMIN_DATABASE_DSN
```

Check service:

```bash
curl http://127.0.0.1:8088/healthz
curl http://127.0.0.1:8088/readyz
```

## 11. Common Errors And Fixes

| Error | Reason | Fix |
|-------|--------|-----|
| Service uses default config unexpectedly | `.env` was not sourced | Run `set -a; source .env; set +a` |
| `SUPERADMIN_DATABASE_DSN is required` | Non-local env or strict flag | Export DSN |
| Downstream unavailable | Base URL empty or service down | Set URL and start downstream service |
| Invalid duration warning | Env like `5sec` instead of `5s` | Use Go duration format: `5s`, `5m`, `720h` |
| `/readyz` not ready | DB ping failed | Check DSN and MySQL status |

## 12. Security Notes

- `.env` contains secrets; use `.env.example` with placeholders for docs.
- Do not commit production DSNs/passwords.
- Use separate credentials per environment.
- Set `SUPERADMIN_REQUIRE_DATABASE=true` outside local/test.
- Set downstream require flags true in staging/prod so misconfiguration fails fast.

## 13. Final Checklist

- [x] Runtime env loader exists.
- [x] Local `.env` exists.
- [x] Config validation exists.
- [ ] `.env.example` missing.
- [ ] `.env` is not auto-loaded by code.
- [ ] Secret handling needs hardening for production.
