# Payment Service Local Runbook

## 1. Purpose

Owns payment intents, provider callbacks/webhooks, payment retry, refunds, reconciliation, and payment admin operations.

## 2. Location

`backend/services/payment-service`

## 3. Tech Stack

Go `1.26.3`, HTTP, MySQL, pluggable payment provider adapters, reconciliation command.

## 4. Required Dependencies

MySQL `payment_db`. Provider sandbox credentials are optional and disabled by default.

## 5. Environment Variables

Use `backend/services/payment-service/.env.example`.

Key vars: `PAYMENT_HTTP_ADDR`, `PAYMENT_MYSQL_DSN`, `PAYMENT_INTERNAL_API_TOKEN`, `PAYMENT_ALLOWED_PROVIDERS`, `PAYMENT_DEFAULT_PROVIDER`, `PAYMENT_EVENTS_ENDPOINT`, webhook secrets, reconciliation vars.

## 6. Install Dependencies

```powershell
cd backend/services/payment-service
go mod download
```

## 7. Database/Migration/Seed Setup

Migrations live in `backend/services/payment-service/migrations`.

```powershell
docker compose up migrate-payment
```

Payment provider seed/sandbox config: Not found in codebase - please confirm.

## 8. Run Command

Recommended:

```powershell
docker compose up -d --build payment-service
```

Optional reconciliation job:

```powershell
docker compose --profile jobs up -d --build payment-reconciliation
```

Manual:

```powershell
cd backend/services/payment-service
go run ./cmd/server
```

## 9. Health Check

- `http://localhost:8091/healthz`
- Schema health: `http://localhost:8091/internal/v1/payment-schema/health`

## 10. Logs

```powershell
docker compose logs -f payment-service
docker compose logs -f payment-reconciliation
```

Check provider, webhook, token, reconciliation, and event publisher messages.

## 11. Common Issues

- Providers are disabled by default.
- `PAYMENT_EVENTS_ENDPOINT` is empty by default, so order updates need configuration.
- Provider capture support may be unsupported in current adapters.
- Webhook signature secrets are not configured locally.

## 12. Quick Verification

```powershell
curl http://localhost:8091/healthz
curl http://localhost:8091/internal/v1/payment-schema/health
```
