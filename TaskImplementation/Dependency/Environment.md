# User Service - Environment Configuration Dependency

## 1. What is this dependency?

Environment variables runtime configuration ka source hain. User Service config file OS environment se values read karta hai.

Important: Code `.env` file automatically load nahi karta. `.env` ko shell me export karna padega.

## 2. Why User Service uses it

Service ko DB DSN, gRPC address, reflection, graceful shutdown, DB pool, and log level runtime pe configurable chahiye.

## 3. Required or Optional

| Item | Required? | Notes |
|---|---:|---|
| `USER_SERVICE_DATABASE_DSN` | Yes | Service startup fail karega agar missing hai |
| `MYSQL_DSN` | Conditional | Fallback if main DSN missing |
| gRPC env vars | Optional | Defaults exist |
| DB pool env vars | Optional | Defaults exist |
| `.env` file | Optional local convenience | Exists locally but no `.env.example` found |

## 4. Where it is used in project

| Path | Purpose |
|---|---|
| `backend/services/user-service/internal/config/config.go` | Env read and validation |
| `backend/services/user-service/.env` | Local env values found in workspace |
| `backend/services/user-service/cmd/server/main.go` | Uses loaded config for DB/gRPC/logger |
| `backend/services/api-gateway/.env` | Contains `USER_GRPC_ADDR=localhost:50052`, but gateway source not clearly found |

## 5. Installation Steps

No package install required.

Verify shell can see variables:

```bash
env | grep USER_SERVICE_
```

If using `.env`, export it:

```bash
cd backend/services/user-service
set -a
. ./.env
set +a
```

## 6. Docker Setup, If Possible

No Dockerfile/compose was found. If service is containerized later, pass env vars using compose `environment:` or explicit `-e` flags.

Suggested Docker run style:

```bash
docker run \
  -e USER_SERVICE_DATABASE_DSN='ecommerce_user:Ecom8880User@tcp(user-mysql:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC' \
  -e USER_SERVICE_GRPC_ADDRESS=':50052' \
  user-service:local
```

This is suggested only because no User Service image exists.

## 7. Local Setup Without Docker

Use shell export:

```bash
export USER_SERVICE_DATABASE_DSN='ecommerce_user:Ecom8880User@tcp(127.0.0.1:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC'
export USER_SERVICE_GRPC_ADDRESS=':50052'
export USER_SERVICE_GRPC_REFLECTION='true'
export USER_SERVICE_LOG_LEVEL='debug'
```

Or source local `.env`:

```bash
cd backend/services/user-service
set -a
. ./.env
set +a
```

## 8. Required Environment Variables

| Variable | Required | Default | Validation in code |
|---|---:|---|---|
| `USER_SERVICE_DATABASE_DSN` | Yes | None | Required unless `MYSQL_DSN` exists |
| `MYSQL_DSN` | Conditional | None | Used only as fallback |
| `USER_SERVICE_GRPC_ADDRESS` | No | `:50052` | No strict format validation |
| `USER_SERVICE_GRPC_REFLECTION` | No | `true` | Parses bool, fallback if invalid |
| `USER_SERVICE_SHUTDOWN_TIMEOUT` | No | `10s` | Must be positive |
| `USER_SERVICE_DB_MAX_OPEN_CONNS` | No | `25` | Must be greater than zero |
| `USER_SERVICE_DB_MAX_IDLE_CONNS` | No | `25` | Cannot be negative |
| `USER_SERVICE_DB_CONN_MAX_LIFETIME` | No | `5m` | Invalid value falls back |
| `USER_SERVICE_DB_PING_TIMEOUT` | No | `5s` | Must be positive |
| `USER_SERVICE_LOG_LEVEL` | No | `info` | Unknown values become info |

## 9. Start Commands

Suggested command based on project structure:

```bash
cd backend/services/user-service
set -a
. ./.env
set +a
go run ./cmd/server
```

## 10. Verify Running Commands

Check env exported:

```bash
printf '%s\n' "$USER_SERVICE_DATABASE_DSN"
printf '%s\n' "$USER_SERVICE_GRPC_ADDRESS"
```

Check service startup logs:

```text
user_service_starting
user_service_grpc_listening
```

Check gRPC port:

```bash
ss -ltnp | grep 50052
```

## 11. Common Errors and Fixes

| Error | Cause | Fix |
|---|---|---|
| `USER_SERVICE_DATABASE_DSN is required` | `.env` sourced but not exported, or missing | Use `set -a` before sourcing |
| DB connects to wrong port | DSN points to `3306` while Docker mapped `3307` | Update DSN port |
| Bool/duration ignored | Invalid env value | Use values like `true`, `false`, `10s`, `5m` |
| Reflection unexpectedly enabled | Default is `true` | Set `USER_SERVICE_GRPC_REFLECTION=false` in production |
| Debug logs too noisy | `USER_SERVICE_LOG_LEVEL=debug` | Use `info` or stricter level |

## 12. Security Notes

- `.env` contains DB credentials. Do not commit real secrets.
- No `.env.example` was found. Add sanitized sample later.
- Prefer secrets manager or platform env injection in production.
- Avoid logging DSN because it contains password.
- Bind gRPC to private/internal network in production.

## 13. Final Checklist

| Check | Done |
|---|---|
| Required DSN exported | [ ] |
| DSN points to migrated DB | [ ] |
| `.env` not committed with real secrets | [ ] |
| Reflection disabled or internal-only in production | [ ] |
| DB pool values are sane | [ ] |
| Log level appropriate for environment | [ ] |
| Sanitized `.env.example` added later | [ ] |
