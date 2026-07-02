# Order Service Local Runbook

## 1. Purpose

Owns checkout, order lifecycle, buyer/seller order reads, inventory reservation coordination, payment coordination, and order events.

## 2. Location

`backend/services/order-service`

## 3. Tech Stack

Go `1.26.3`, HTTP, gRPC, MySQL, Kafka, HTTP clients to cart/product/payment.

## 4. Required Dependencies

MySQL `order_db`, cart service, product service, payment service, Kafka.

## 5. Environment Variables

Use `backend/services/order-service/.env.example`.

Key vars: `ORDER_HTTP_ADDR`, `ORDER_GRPC_ADDR`, `ORDER_MYSQL_DSN`, `ORDER_ADMIN_TOKEN`, `ORDER_CART_BASE_URL`, `ORDER_PRODUCT_BASE_URL`, `ORDER_PAYMENT_BASE_URL`, `ORDER_KAFKA_BROKERS`, `ORDER_PAYMENT_INTERNAL_TOKEN`.

## 6. Install Dependencies

```powershell
cd backend/services/order-service
go mod download
```

## 7. Database/Migration/Seed Setup

Migrations live in `backend/services/order-service/migrations`.

```powershell
docker compose up migrate-order
```

Seed orders: Not found in codebase - please confirm.

## 8. Run Command

Recommended:

```powershell
docker compose up -d --build order-service
```

Manual:

```powershell
cd backend/services/order-service
go run ./cmd/server
```

`cmd/local-server` exists, but it is not the full business service.

## 9. Health Check

- `http://localhost:8090/healthz`
- `http://localhost:8090/readyz`
- gRPC host port: `localhost:9094`

## 10. Logs

```powershell
docker compose logs -f order-service
```

Check checkout, payment event, inventory reservation, Kafka outbox, and downstream client messages.

## 11. Common Issues

- Payment service not ready blocks checkout.
- Product token mismatch blocks inventory calls.
- Kafka unavailable blocks event publishing.
- `ORDER_PAYMENT_RETURN_URL` must be HTTPS per config validation.

## 12. Quick Verification

```powershell
curl http://localhost:8090/readyz
```
