# Cart Service Local Runbook

## 1. Purpose

Owns buyer/guest cart state, cart item updates, merge behavior, coupon previews, and cart expiry cleanup.

## 2. Location

`backend/services/cart-service`

## 3. Tech Stack

Go `1.26.3`, HTTP, MongoDB, Redis, product/CMS HTTP clients, optional expiry worker.

## 4. Required Dependencies

MongoDB `cart_db`, Redis DB 2, product service, CMS service for coupon validation.

## 5. Environment Variables

Use `backend/services/cart-service/.env.example`.

Key vars: `CART_HTTP_ADDR`, `CART_MONGO_URI`, `CART_MONGO_DATABASE`, `CART_REDIS_ADDR`, `CART_REDIS_PASSWORD`, `CART_PRODUCT_BASE_URL`, `CART_CMS_BASE_URL`.

## 6. Install Dependencies

```powershell
cd backend/services/cart-service
go mod download
```

## 7. Database/Migration/Seed Setup

Mongo migration lives in `backend/services/cart-service/migrations/mongo`.

```powershell
docker compose up migrate-mongodb
```

Seed carts: Not found in codebase - please confirm.

## 8. Run Command

Recommended:

```powershell
docker compose up -d --build cart-service
```

Optional worker:

```powershell
docker compose --profile jobs up -d --build cart-expiry-worker
```

Manual:

```powershell
cd backend/services/cart-service
go run ./cmd/server
```

## 9. Health Check

- `http://localhost:8083/healthz`
- `http://localhost:8083/readyz`

## 10. Logs

```powershell
docker compose logs -f cart-service
docker compose logs -f cart-expiry-worker
```

Check MongoDB, Redis, product, and CMS downstream calls.

## 11. Common Issues

- Coupon preview fails while CMS is unavailable.
- Product validation fails while product service is unavailable.
- Redis password or DB mismatch.
- Cart cleanup worker should not be required for basic cart testing.

## 12. Quick Verification

```powershell
curl http://localhost:8083/readyz
```
