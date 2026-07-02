# CMS Service Local Runbook

## 1. Purpose

Owns seller dashboard APIs for seller access, coupons, campaigns, product moderation, seller analytics, settings, and audit logs.

## 2. Location

`backend/services/cms-service`

## 3. Tech Stack

Go `1.26.3`, HTTP, gRPC, MySQL, product HTTP client.

## 4. Required Dependencies

MySQL `cms_db`, product service, API Gateway and cart/order callers for internal CMS APIs.

## 5. Environment Variables

Use `backend/services/cms-service/.env.example`.

Key vars: `CMS_HTTP_ADDR`, `CMS_GRPC_ADDR`, `CMS_MYSQL_DSN`, `CMS_INTERNAL_AUTH_HEADER`, `CMS_INTERNAL_AUTH_TOKEN`, `CMS_PRODUCT_SERVICE_BASE_URL`, `CMS_PRODUCT_SERVICE_AUTH_TOKEN`.

## 6. Install Dependencies

```powershell
cd backend/services/cms-service
go mod download
```

## 7. Database/Migration/Seed Setup

Migrations live in `backend/services/cms-service/migrations`.

```powershell
docker compose up migrate-cms
```

Seller/CMS seed data: Not found in codebase - please confirm.

## 8. Run Command

Recommended:

```powershell
docker compose up -d --build cms-service
```

Manual:

```powershell
cd backend/services/cms-service
go run ./cmd/server
```

## 9. Health Check

- `http://localhost:8087/healthz`
- gRPC host port: `localhost:50057`

## 10. Logs

```powershell
docker compose logs -f cms-service
```

Check MySQL, product client, internal token, and seller analytics messages.

## 11. Common Issues

- `CMS_INTERNAL_AUTH_TOKEN` is blank in local example.
- Product service token mismatch.
- MySQL migrations not applied.
- Seller staff/status source needs gateway context for real flows.

## 12. Quick Verification

```powershell
curl http://localhost:8087/healthz
```
