# Superadmin Service Local Runbook

## 1. Purpose

Owns superadmin RBAC, user/seller/order/payment/session admin views, platform settings, reviews, and audit operations.

## 2. Location

`backend/services/superadmin-service`

## 3. Tech Stack

Go `1.26.3`, HTTP, MySQL, HTTP clients to user/order/payment/session services.

## 4. Required Dependencies

MySQL `superadmin_db`, user service, order service, payment service, session service.

## 5. Environment Variables

Use `backend/services/superadmin-service/.env.example`.

Key vars: `HTTP_ADDR`, `SUPERADMIN_DATABASE_DSN`, `SUPERADMIN_REQUIRE_*`, `USER_SERVICE_ADMIN_BASE_URL`, `ORDER_SERVICE_ADMIN_BASE_URL`, `PAYMENT_SERVICE_ADMIN_BASE_URL`, `SESSION_SERVICE_ADMIN_BASE_URL`, admin tokens.

## 6. Install Dependencies

```powershell
cd backend/services/superadmin-service
go mod download
```

## 7. Database/Migration/Seed Setup

Migrations live in `backend/services/superadmin-service/migrations`.

```powershell
docker compose up migrate-superadmin
```

Superadmin credentials: Not found in codebase - please confirm.

## 8. Run Command

Recommended:

```powershell
docker compose up -d --build superadmin-service
```

Manual:

```powershell
cd backend/services/superadmin-service
go run ./cmd/server
```

## 9. Health Check

- `http://localhost:8093/healthz`
- `http://localhost:8093/readyz`

## 10. Logs

```powershell
docker compose logs -f superadmin-service
```

Check MySQL, downstream admin tokens, user/order/payment/session readiness.

## 11. Common Issues

- Readiness fails when any required downstream service is unavailable.
- Admin token mismatch blocks downstream calls.
- Superadmin seed credentials are missing.

## 12. Quick Verification

```powershell
curl http://localhost:8093/readyz
```
