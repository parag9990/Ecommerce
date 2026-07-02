# Notification Service Local Runbook

## 1. Purpose

Owns OTP delivery, notification templates, preferences, event-driven delivery, retries/DLQ, analytics, and provider callbacks.

## 2. Location

`backend/services/notification-service`

## 3. Tech Stack

Go `1.26.3`, gRPC, internal HTTP, MongoDB, RabbitMQ, SMTP/Mailpit, optional SMS/push providers.

## 4. Required Dependencies

MongoDB `notification_db`, RabbitMQ, Mailpit for local SMTP, auth service as a caller.

## 5. Environment Variables

Use `backend/services/notification-service/.env.example`.

Key vars: `NOTIFICATION_GRPC_ADDRESS`, `NOTIFICATION_ANALYTICS_HTTP_ADDRESS`, `NOTIFICATION_MONGO_URI`, `NOTIFICATION_RABBITMQ_URL`, `NOTIFICATION_EMAIL_*`, `NOTIFICATION_DELIVERY_ENCRYPTION_KEY`.

## 6. Install Dependencies

```powershell
cd backend/services/notification-service
go mod download
```

## 7. Database/Migration/Seed Setup

Mongo migrations live in `backend/services/notification-service/migrations`.

```powershell
docker compose up migrate-mongodb
```

Templates are seeded by notification Mongo migrations.

## 8. Run Command

Recommended:

```powershell
docker compose up -d --build notification-service
```

Manual:

```powershell
cd backend/services/notification-service
go run ./cmd/server
```

## 9. Health Check

- `http://localhost:8092/healthz`
- `http://localhost:8092/readyz`
- gRPC host port: `localhost:50060`
- Mailpit: `http://localhost:8025`

## 10. Logs

```powershell
docker compose logs -f notification-service
```

Check MongoDB, RabbitMQ consumers, SMTP, template render, retry, and webhook messages.

## 11. Common Issues

- Mailpit not running blocks local email delivery.
- Encryption key must decode to expected key size.
- RabbitMQ event queues unavailable.
- SMS/push/WhatsApp providers are disabled by default.

## 12. Quick Verification

```powershell
curl http://localhost:8092/readyz
```
