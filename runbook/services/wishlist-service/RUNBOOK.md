# Wishlist Service Local Runbook

## 1. Purpose

Owns wishlist APIs, wishlist-to-cart actions, product event reactions, price-drop candidates, and recommendation events.

## 2. Location

`backend/services/wishlist-service`

## 3. Tech Stack

Go `1.26.3`, HTTP, MongoDB, Kafka, product/cart HTTP clients.

## 4. Required Dependencies

MongoDB `wishlist_db`, Kafka, product service, cart service, notification/recommendation event topics.

## 5. Environment Variables

Use `backend/services/wishlist-service/.env.example`.

Key vars: `WISHLIST_HTTP_ADDR`, `WISHLIST_MONGO_URI`, `WISHLIST_MONGO_DATABASE`, `WISHLIST_PRODUCT_SERVICE_BASE_URL`, `WISHLIST_CART_SERVICE_BASE_URL`, `WISHLIST_KAFKA_BROKERS`, `WISHLIST_EVENTS_BACKEND`.

## 6. Install Dependencies

```powershell
cd backend/services/wishlist-service
go mod download
```

## 7. Database/Migration/Seed Setup

Mongo migrations live in `backend/services/wishlist-service/migrations/mongo`.

```powershell
docker compose up migrate-mongodb
```

Seed wishlists: Not found in codebase - please confirm.

## 8. Run Command

Recommended:

```powershell
docker compose up -d --build wishlist-service
```

Manual:

```powershell
cd backend/services/wishlist-service
go run ./cmd/server
```

## 9. Health Check

- `http://localhost:8084/healthz`
- `http://localhost:8084/readyz`

## 10. Logs

```powershell
docker compose logs -f wishlist-service
```

Check MongoDB, Kafka consumer/producer, product, and cart messages.

## 11. Common Issues

- Kafka not ready.
- Mongo migration missing.
- Product or cart service unavailable.
- Auth headers missing when testing direct service APIs.

## 12. Quick Verification

```powershell
curl http://localhost:8084/readyz
```
