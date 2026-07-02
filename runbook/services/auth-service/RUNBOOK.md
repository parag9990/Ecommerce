# Auth Service Local Runbook

## 1. Purpose

Handles login, refresh/logout, OTP, password flows, JWT issuing, JWKS, and role/auth internals.

## 2. Location

`backend/services/auth-service`

## 3. Tech Stack

Go `1.26.3`, HTTP, MySQL, Redis, JWT RSA keys, gRPC client to notification service.

## 4. Required Dependencies

MySQL `auth_db`, Redis DB 1, notification service gRPC, RSA signing keys.

## 5. Environment Variables

Use `backend/services/auth-service/.env.example`.

Key vars: `AUTH_HTTP_ADDR`, `AUTH_MYSQL_DSN`, `AUTH_REDIS_ADDR`, `AUTH_REDIS_PASSWORD`, `JWT_*`, `REFRESH_TOKEN_PEPPER`, `OTP_*_PEPPER`, `NOTIFICATION_GRPC_ADDR`, `AUTH_EVENTS_PUBLISH_ENDPOINT`.

## 6. Install Dependencies

```powershell
cd backend/services/auth-service
go mod download
```

## 7. Database/Migration/Seed Setup

Migrations live in `backend/services/auth-service/migrations`.

```powershell
docker compose up migrate-auth
```

Seed users: Not found in codebase - please confirm.

## 8. Run Command

Recommended:

```powershell
docker compose up -d --build auth-service
```

Manual:

```powershell
cd backend/services/auth-service
go run ./cmd/server
```

Host manual runs need local JWT key paths and localhost dependency addresses.

## 9. Health Check

- `http://localhost:8081/healthz`
- `http://localhost:8081/readyz`
- `http://localhost:8081/.well-known/jwks.json`

## 10. Logs

```powershell
docker compose logs -f auth-service
```

Check MySQL, Redis, JWT key, notification gRPC, and OTP messages.

## 11. Common Issues

- JWT key files missing during host runs.
- Redis or MySQL not ready.
- Notification service unavailable for OTP.
- Signup is referenced by catalog/frontend, but auth router registration needs confirmation.

## 12. Quick Verification

```powershell
curl http://localhost:8081/readyz
curl http://localhost:8081/.well-known/jwks.json
```
