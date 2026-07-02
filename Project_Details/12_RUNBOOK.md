# Runbook

## Local Setup Prerequisites

Found from repo configuration:

| Tool | Version/notes |
| --- | --- |
| Docker and Docker Compose | Required for root compose stack |
| Go | `1.26.3` in `backend/go.work` and CI |
| Node.js | `>=22.13.0` in `frontend/package.json` |
| pnpm | `11.5.0` in `frontend/package.json` |
| Buf | Used for proto lint/generation |

## Environment Setup

Use service `.env.example` files as the source of truth. Important examples:

- `backend/services/api-gateway/.env.example`
- `backend/services/auth-service/.env.example`
- `backend/services/user-service/.env.example`
- `backend/services/product-service/.env.example`
- `backend/services/order-service/.env.example`
- `backend/services/payment-service/.env.example`
- `backend/services/notification-service/.env.example`
- `frontend/*/.env.example`
- `infra/compose/.env.local.example`

Root `docker-compose.yml` already wires many local environment values directly for compose services.

## Database Setup

Root compose includes:

1. MySQL boot with initial database creation from `infra/mysql/init/00-create-databases.sql`.
2. MongoDB replica-set initialization through `mongodb-init-replica`.
3. MySQL migration jobs for auth, user, order, payment, CMS, and superadmin.
4. Mongo migration jobs for product, cart, wishlist, recommendation, session, and notification.

Manual database setup commands outside compose were not found as a dedicated runbook.

## Suggested Local Startup Order

Based on `docker-compose.yml` dependencies:

1. Start infrastructure: MySQL, MongoDB, Redis, RabbitMQ, Kafka, Typesense, Mailpit, observability.
2. Wait for Mongo replica set and MySQL health checks.
3. Run migration jobs.
4. Start auth, user, product, cart, wishlist, search, session, CMS, recommendation, order, payment, notification, and superadmin services.
5. Start API Gateway.
6. Start frontend apps.

With compose, use the root file to let dependencies coordinate startup:

```powershell
docker compose up --build
```

## Common Local URLs

| Component | URL |
| --- | --- |
| API Gateway | `http://localhost:8080` |
| gRPC-Web facade | `http://localhost:8099` |
| User app | `http://localhost:3000` |
| Seller dashboard | `http://localhost:3001` |
| Session analytics dashboard | `http://localhost:3002` |
| Superadmin panel | `http://localhost:3003` |
| Mailpit | `http://localhost:8025` |
| RabbitMQ management | `http://localhost:15672` |
| Jaeger | `http://localhost:16686` |
| Prometheus | `http://localhost:9090` |

## Health Checks

| Service | Health endpoints found |
| --- | --- |
| API Gateway | `/health/live`, `/health/ready` |
| Auth | `/healthz`, `/readyz` |
| User | `/health/live`, `/health/ready`, `/metrics` |
| Product | `/healthz`, `/readyz` |
| Order | `/healthz`, `/readyz` |
| Payment | `/healthz` |
| Notification | `/healthz`, `/readyz`, `/metrics` |

Other service health endpoints likely exist in code, but they were not part of the requested service-specific analysis.

## Smoke Test Checklist

1. Confirm API Gateway health returns ready.
2. Confirm auth JWKS is available at `http://localhost:8081/.well-known/jwks.json`.
3. Confirm frontend user app loads on port `3000`.
4. Confirm Mailpit is reachable on port `8025` for local OTP/email delivery.
5. Confirm product read endpoints respond through API Gateway.
6. Confirm login flow works for seeded/existing credentials if seed data is available.
7. Confirm cart/checkout only after product, cart, order, and payment dependencies are configured.
8. Confirm notification preferences through authenticated API after auth is working.

Seed users or demo credentials: Not found in current codebase.

## Test Commands

Commands found in repo configuration:

```powershell
pnpm --dir frontend build
pnpm --dir frontend test
```

Go CI runs tests per backend module. A single root backend test command was not found; use Go workspace/module commands or CI helper scripts.

## Common Errors And Fixes

| Symptom | Likely cause from code/config | Fix |
| --- | --- | --- |
| Gateway auth fails | `JWT_JWKS_URL`, issuer, or audience mismatch | Check API Gateway and auth env values |
| Signup route returns missing/not found | Auth service does not register public signup route | Implement/register signup or remove caller dependency |
| Payments unavailable | Providers disabled by default | Set allowed/default provider and provider credentials |
| Payment events do not update orders | `PAYMENT_EVENTS_ENDPOINT` empty or token mismatch | Configure endpoint to order service payment-event route and matching token |
| Auth outbox does not publish | `AUTH_EVENTS_PUBLISH_ENDPOINT` empty | Configure endpoint or change session-link mode |
| Frontend API calls fail CORS | Origin not in gateway `CORS_ALLOWED_ORIGINS` | Add local frontend origin to gateway env |
| Mongo migrations fail | Mongo replica set not initialized/ready | Wait for `mongodb-init-replica` and rerun migration job |

## Missing Runbook Information

- `docs/runbooks/LOCAL_PORTS.md`: Not found in current codebase.
- `docs/runbooks/LOCAL_PLATFORM_RUNBOOK.md`: Not found in current codebase.
- Seed/demo data instructions: Not found in current codebase.
- Production incident runbooks: Not found in current codebase.
