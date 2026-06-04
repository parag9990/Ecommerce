# Project Dependency & Setup Guide

## 1. Project Overview

This guide is generated for:

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Payment Service` |
| `TASK_FILE_NAME` | `task5.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task5_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

`TASK_FILE_NAME` ka scope provider webhook handler setup hai. Simple Hinglish me: payment provider jab signed webhook bhejta hai, `SERVICE_NAME` raw body preserve karta hai, signature verify karta hai, duplicate provider event ko safely handle karta hai, `payment_webhook_events` me audit row store karta hai, local payment status update karta hai, and downstream systems ko `payment.events` publish karta hai.

Important boundary:

- Original `INPUT_FILE_PATH` modify nahi kiya gaya.
- Ye file business logic rewrite nahi karti.
- Common clone, Go install, MySQL install, Docker MySQL, migration basics, and provider setup repeat nahi kiya gaya.
- Current repo me webhook code already exists under `backend/services/payment-service/internal/usecase/handle_webhook.go`.
- Current HTTP route already exists at `POST /api/v1/webhooks/payments/{provider}`.

Beginner note:

Webhook final payment truth hota hai. Frontend redirect ya checkout success screen se payment ko final `captured` mark nahi karna chahiye. Local setup me webhook tabhi useful hoga jab matching `payments` row already exist kare, usually Task 4 create-intent flow ke through.

## 2. Previous Dependency Files Reused

Before writing this file, previous dependency docs in the same `SERVICE_NAME` folder were checked:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
```

Do not duplicate these sections. Follow them directly:

| Previous Dependency File | Section / Topic | Reuse reason |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3. Required Software` | Git, Go, MySQL, Docker, curl setup already explained. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Dependency Management` | Go modules commands and common dependency failures already explained. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `5. Database Setup` | MySQL install, Docker MySQL, DSN format, and migration basics already explained. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. Environment Variables` | `.env` location, loading behavior, and broad env reference already explained. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `8. Docker Setup` | MySQL-only Docker baseline already explained. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `11. Ports & Networking` | Base HTTP, MySQL, provider, and event receiver networking notes already explained. |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` | `payment_db`, `payments`, `payment_webhook_events`, indexes, and migration commands already explained. |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `6. Redis / Queue / External Services` | Provider APIs and HTTP event receiver behavior already explained. |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `7. Environment Variables` | Provider-enabled env group already explained. |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | `9. Local Development Setup` | Create-intent setup needed before a realistic webhook test. |

This file only explains `TASK_FILE_NAME`-specific setup and verification.

## 3. Tech Stack

### Reused technology documentation

Common technology setup is already explained in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
```

Refer sections:

- `task1_Dependency.md` -> `2. Tech Stack`
- `task1_Dependency.md` -> `4. Dependency Management`
- `task2_Dependency.md` -> `5. Database Setup`
- `task3_Dependency.md` -> `6. Redis / Queue / External Services`
- `task4_Dependency.md` -> `2. Tech Stack`

### `TASK_FILE_NAME` specific technology impact

| Technology / Component | Required? | Why used in `TASK_FILE_NAME` | Beginner explanation |
|---|---:|---|---|
| Go `net/http` transport | Required | Receives `POST /api/v1/webhooks/payments/{provider}` | This is the public provider callback endpoint. No Gin/Fiber/Echo framework is used. |
| Go usecase layer | Required | `HandleWebhookUsecase` verifies provider event and coordinates persistence/publishing | Usecase is the service workflow brain. It does not read HTTP directly. |
| Provider registry | Required | Resolves route provider like `stripe_like` or `razorpay_like` | Only providers enabled by env config can process webhooks. |
| Go `crypto/hmac` and `crypto/sha256` | Required | Verifies webhook HMAC signatures | Standard library, no package install needed. |
| MySQL | Required | Stores `payments`, `payment_attempts`, and `payment_webhook_events` | Duplicate webhook protection depends on a unique MySQL key. |
| `github.com/go-sql-driver/mysql` | Required | Go MySQL driver used by repository | Already present in `go.mod`; no new `TASK_FILE_NAME` dependency added. |
| HTTP event publisher | Required when providers enabled | Publishes `PaymentCaptured`, `PaymentFailed`, refund events, etc. to `PAYMENT_EVENTS_ENDPOINT` | Current repo uses HTTP publishing, not Kafka/RabbitMQ runtime. |
| Provider webhook secret | Required per enabled provider | Signature cannot be verified without it | This is different from provider API secret key. Keep it backend-only. |

Important clarification:

`TASK_FILE_NAME` guide diagrams mention Kafka/RabbitMQ conceptually for downstream payment events, but current repo code publishes events through HTTP using `PAYMENT_EVENTS_ENDPOINT`. No Redis, Kafka, RabbitMQ, or NATS container is configured by this task.

## 4. Required Software

No new base software installation is introduced by `TASK_FILE_NAME`.

Follow:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Sections:

- `3. Required Software`
- `4. Dependency Management`
- `5. Database Setup`
- `8. Docker Setup`

`TASK_FILE_NAME` specific runtime needs:

| Software / Access | Required? | Status | Notes |
|---|---:|---|---|
| Go | Required | Reused | Needed to run service and webhook tests. |
| MySQL | Required | Reused | Webhook processing writes `payment_webhook_events` and updates `payments`. |
| MySQL client | Required for setup | Reused | Needed to verify webhook rows. |
| Docker | Optional | Reused | Useful for local MySQL only. |
| curl | Recommended | Reused | Useful to post test webhooks. |
| Provider sandbox/dashboard | Optional for real provider testing | Task-specific | Needed if you want the real provider to send signed webhooks. |
| Local HTTP event receiver | Required for full provider-mode success | Reused / task-specific | Must accept `payment.events` HTTP posts. |
| Local tunnel such as an approved HTTPS forwarding tool | Optional | Task-specific | Needed only when a real external provider must reach your local machine. |

If you only run unit tests, provider sandbox, tunnel, Docker service container, Kafka, and RabbitMQ are not required.

## 5. Dependency Management

This remains a Go modules project. `TASK_FILE_NAME` does not add a new Go module dependency.

Refer:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Section:

```text
4. Dependency Management
```

Current module details:

| Item | Current value |
|---|---|
| Service module | `backend/services/payment-service/go.mod` |
| Go version declared | `1.26.3` |
| Direct external dependency | `github.com/go-sql-driver/mysql v1.10.0` |
| Indirect dependency | `filippo.io/edwards25519 v1.2.0` |

Task-specific code paths:

| Path | Purpose |
|---|---|
| `backend/services/payment-service/internal/transport/http/handler.go` | Reads raw webhook body and calls usecase. |
| `backend/services/payment-service/internal/usecase/handle_webhook.go` | Verifies provider event, stores webhook, publishes domain event. |
| `backend/services/payment-service/internal/provider/stripe_like.go` | Stripe-like webhook signature and payload normalization. |
| `backend/services/payment-service/internal/provider/razorpay_like.go` | Razorpay-like webhook signature and payload normalization. |
| `backend/services/payment-service/internal/provider/webhook_signature.go` | Shared HMAC/timestamp helper logic. |
| `backend/services/payment-service/internal/repository/mysql_payment_repository.go` | Transactional webhook idempotency and payment status update. |
| `backend/services/payment-service/internal/domain/webhook_event.go` | Webhook event validation and domain event creation. |

Useful test commands:

```bash
cd backend/services/payment-service
go test ./internal/provider -run Webhook
go test ./internal/usecase -run HandleWebhook
go test ./internal/transport/http -run Webhook
```

## 6. Database Setup

Common MySQL setup and migration commands are already explained in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
```

Sections:

- `task1_Dependency.md` -> `5. Database Setup`
- `task2_Dependency.md` -> `5. Database Setup`

### `TASK_FILE_NAME` database dependencies

`TASK_FILE_NAME` relies on these existing tables and indexes:

| Database object | Required? | Why `TASK_FILE_NAME` needs it |
|---|---:|---|
| `payments` | Yes | Finds and updates local payment status after a verified webhook. |
| `payment_attempts` | Yes | Final webhook can update latest attempt status when mapped. |
| `payment_webhook_events` | Yes | Stores provider event id and processed flag for idempotency/audit. |
| `uk_webhook_provider_event` | Yes | Prevents duplicate processing for same `(provider, provider_event_id)`. |
| `idx_payments_provider_intent` | Yes for provider intent lookup | Added by migration `002`; helps lookup by provider intent/order id. |

Required migration files:

| Migration | Why it matters |
|---|---|
| `migrations/001_create_payment_tables.up.sql` | Creates `payment_webhook_events` and baseline payment tables. |
| `migrations/002_add_webhook_payment_lookup_index.up.sql` | Adds provider intent lookup index used by webhook matching. |

For current service run, apply all `.up.sql` migrations as already documented in `task2_Dependency.md`.

### Task-specific DB verification

After migrations, verify webhook table exists:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db -e "SHOW TABLES LIKE 'payment_webhook_events';"
```

Verify idempotency indexes:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db -e "SHOW INDEX FROM payment_webhook_events;"
```

Must-have index names:

```text
uk_webhook_provider_event
uk_webhook_event_id
idx_webhook_processed_received
```

Verify provider intent lookup index:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db -e "SHOW INDEX FROM payments WHERE Key_name = 'idx_payments_provider_intent';"
```

Beginner note:

If a signed webhook returns `success: true` with `processing_status: ignored`, DB setup may be fine. It often means the webhook was valid but could not match a local payment row. Create a payment intent first using Task 4 flow, then send a webhook whose metadata references that exact `payment_id`.

## 7. Redis / Queue / External Services

### Redis / Kafka / RabbitMQ / NATS

No Redis, Kafka, RabbitMQ, or NATS dependency is introduced by `TASK_FILE_NAME`.

Reuse:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
```

Sections:

- `task1_Dependency.md` -> `6. Redis / Queue / External Services`
- `task3_Dependency.md` -> `6. Redis / Queue / External Services`

### Payment provider webhook delivery

Webhook endpoint:

```text
POST /api/v1/webhooks/payments/{provider}
```

Provider keys supported by current code:

| Provider key | Signature header | Required secret env | Notes |
|---|---|---|---|
| `stripe_like` | `Stripe-Signature` | `STRIPE_LIKE_WEBHOOK_SECRET` | Header contains timestamp and one or more `v1` signatures. |
| `razorpay_like` | `X-Razorpay-Signature` | `RAZORPAY_LIKE_WEBHOOK_SECRET` | Header is HMAC hex over raw request body. |

Real provider sandbox setup:

| Need | Why |
|---|---|
| Public HTTPS callback URL | Provider cannot call `localhost` directly. |
| Exact route path | Provider dashboard must point to `/api/v1/webhooks/payments/{provider}`. |
| Webhook signing secret | Must match env secret loaded by backend. |
| Raw body forwarding | Gateway/proxy must not parse and reserialize body before `SERVICE_NAME`. |

Local testing options:

| Option | Use when | Notes |
|---|---|---|
| Unit tests | Fastest developer check | No provider account, event receiver, or tunnel needed. |
| curl with locally generated HMAC | You want to test HTTP route locally | Requires matching provider env and existing payment row. |
| Provider sandbox dashboard | You want realistic provider delivery | Requires public HTTPS URL/tunnel and sandbox keys. |

### HTTP event receiver

Webhook processing publishes domain events to `PAYMENT_EVENTS_ENDPOINT`.

Reuse:

```text
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
```

Section:

```text
6. Redis / Queue / External Services -> HTTP event receiver
```

Task-specific event behavior:

| Event topic | Example event types | Receiver requirement |
|---|---|---|
| `payment.events` | `PaymentCaptured`, `PaymentFailed`, `PaymentAuthorized`, `PaymentRequiresAction` | Accept `POST`, bearer token, idempotency key, and return 2xx. |

If event receiver is down after webhook DB processing, current usecase logs `payment.webhook.event_publish_failed` but still returns webhook success. This avoids provider retry storms after the payment status has already been stored.

## 8. Environment Variables

Common env loading and provider-enabled env reference already exist in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
```

Sections:

- `task1_Dependency.md` -> `7. Environment Variables`
- `task3_Dependency.md` -> `7. Environment Variables`
- `task4_Dependency.md` -> `7. Environment Variables`

### `TASK_FILE_NAME` required env group

No brand-new env variable is introduced by `TASK_FILE_NAME`. For webhook processing, these existing variables must be correct:

| Variable | Required? | Why `TASK_FILE_NAME` needs it | Beginner note |
|---|---:|---|---|
| `PAYMENT_ALLOWED_PROVIDERS` | Yes | Enables provider registry and webhook usecase | Example: `stripe_like` or `razorpay_like`. |
| `PAYMENT_DEFAULT_PROVIDER` | Yes when providers enabled | Provider config validation requires it | Must be included in allowed providers. |
| `PAYMENT_INTERNAL_API_TOKEN` | Yes when providers enabled | Server startup validation requires it | Webhook route itself does not use user/internal bearer auth. |
| `PAYMENT_WEBHOOK_MAX_BODY_BYTES` | Optional | Limits raw webhook body size | Default is `1048576`. Too small causes `INVALID_WEBHOOK_BODY`. |
| `PAYMENT_WEBHOOK_TIMESTAMP_TOLERANCE` | Optional | Stripe-like replay protection window | Default is `5m`. Server clock must be accurate. |
| `STRIPE_LIKE_PUBLIC_KEY` | Yes if `stripe_like` enabled | Provider config validation | Needed even though webhook primarily uses webhook secret. |
| `STRIPE_LIKE_SECRET_KEY` | Yes if `stripe_like` enabled | Provider config validation and provider API calls | Backend-only secret. |
| `STRIPE_LIKE_WEBHOOK_SECRET` | Yes if `stripe_like` enabled | HMAC signature verification | Must match provider dashboard/tunnel/test signature secret. |
| `STRIPE_LIKE_BASE_URL` | Optional | Provider API base URL override | Use `http://127.0.0.1...` only for local mock. |
| `RAZORPAY_LIKE_PUBLIC_KEY` | Yes if `razorpay_like` enabled | Provider config validation | Provider public key. |
| `RAZORPAY_LIKE_SECRET_KEY` | Yes if `razorpay_like` enabled | Provider config validation and provider API calls | Backend-only secret. |
| `RAZORPAY_LIKE_WEBHOOK_SECRET` | Yes if `razorpay_like` enabled | HMAC signature verification | Must match webhook signing secret. |
| `RAZORPAY_LIKE_BASE_URL` | Optional | Provider API base URL override | Use HTTPS except loopback/local testing. |
| `PAYMENT_EVENTS_ENDPOINT` | Yes when providers enabled | HTTP publisher initialization | Use HTTPS outside localhost/loopback. |
| `PAYMENT_EVENTS_AUTH_TOKEN` | Yes when providers enabled | Bearer token sent to event receiver | Minimum 32 chars. |
| `PAYMENT_EVENTS_TIMEOUT` | Optional | HTTP publish timeout | Default is `5s`. |

### Minimal webhook-focused `.env` example

This example enables one provider. Reuse the full provider `.env` from previous docs if you need all flows.

```env
PAYMENT_HTTP_ADDR=:8080
PAYMENT_MYSQL_DSN=root:root@tcp(127.0.0.1:3306)/payment_db?parseTime=true&charset=utf8mb4,utf8&loc=UTC

PAYMENT_INTERNAL_API_TOKEN=change-this-internal-token-at-least-32-chars

PAYMENT_DEFAULT_PROVIDER=stripe_like
PAYMENT_ALLOWED_PROVIDERS=stripe_like
PAYMENT_ALLOWED_CURRENCIES=INR,USD

PAYMENT_WEBHOOK_MAX_BODY_BYTES=1048576
PAYMENT_WEBHOOK_TIMESTAMP_TOLERANCE=5m

STRIPE_LIKE_PUBLIC_KEY=pk_test_replace_me
STRIPE_LIKE_SECRET_KEY=sk_test_replace_me
STRIPE_LIKE_WEBHOOK_SECRET=whsec_test_replace_me
STRIPE_LIKE_BASE_URL=
STRIPE_LIKE_TIMEOUT=5s

PAYMENT_EVENTS_ENDPOINT=http://127.0.0.1:9080/events
PAYMENT_EVENTS_AUTH_TOKEN=change-this-event-token-at-least-32-chars
PAYMENT_EVENTS_TIMEOUT=5s
```

`.env` automatic load nahi hota. Load it manually:

```bash
cd backend/services/payment-service
set -a
. ./.env
set +a
```

Security notes:

- Never commit real webhook secrets.
- Never log signature header values.
- Keep provider webhook secret separate from provider API secret key.
- Use HTTPS for remote provider and event receiver URLs.

## 9. Docker Setup

No new Dockerfile, docker-compose service, Redis, Kafka, RabbitMQ, or provider mock container is introduced by `TASK_FILE_NAME`.

Reuse:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
```

Sections:

- `task1_Dependency.md` -> `8. Docker Setup`
- `task3_Dependency.md` -> `8. Docker Setup`

Task-specific Docker notes:

| Container / Service | Required? | Notes |
|---|---:|---|
| MySQL container | Required if no local MySQL | Same `payment_db` setup from previous docs. |
| `SERVICE_NAME` container | Not committed | Run with `go run ./cmd/server` unless you add a Dockerfile later. |
| Event receiver container | Optional / external | Needed only if your event receiver runs in Docker. |
| Provider mock container | Optional / not committed | Useful for deterministic local webhook tests but not present in repo. |
| Kafka/RabbitMQ container | Not required | Current code uses HTTP event publisher. |

Networking reminders:

| Situation | Correct setup |
|---|---|
| Service runs on host and MySQL runs in Docker | DSN can use `127.0.0.1:3306` if Docker published port 3306. |
| Service runs in Docker and MySQL runs in Docker | DSN should use Docker service name, not `127.0.0.1`. |
| Event receiver runs in Docker | `PAYMENT_EVENTS_ENDPOINT` should point to reachable host/service name. |
| Real provider calls local service | Public HTTPS tunnel must forward to local `:8080`. |

## 10. Local Development Setup

### Step 1: Read previous setup docs

Start with these docs instead of repeating full setup here:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
```

Use them for clone, Go dependencies, MySQL, Docker, migrations, provider env, and create-intent.

### Step 2: Go to service directory

```bash
cd backend/services/payment-service
```

### Step 3: Install Go dependencies

Reuse `task1_Dependency.md` -> `4. Dependency Management`.

Current quick command:

```bash
go mod download
```

### Step 4: Start MySQL and apply migrations

Reuse:

```text
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
```

Section:

```text
5. Database Setup -> Apply migrations
```

For `TASK_FILE_NAME`, make sure migration `001` and `002` are applied.

### Step 5: Create or reuse a payment row

Realistic webhook processing needs a matching local payment.

Recommended beginner path:

```text
Follow TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
Section: 9. Local Development Setup
```

Create a payment intent first, then use that returned `payment_id`, `order_id`, and provider intent/payment ids in webhook payloads.

Without a matching payment row:

| Case | Expected behavior |
|---|---|
| Signature invalid | Request fails before DB write. |
| Signature valid but payment not found | Webhook event can be stored and marked `ignored`. |
| Signature valid and payment found | Payment status can update and event can publish. |

### Step 6: Configure provider webhook env

Use the env group in this file section `8. Environment Variables`.

At minimum for `stripe_like` webhook testing:

```text
PAYMENT_ALLOWED_PROVIDERS=stripe_like
PAYMENT_DEFAULT_PROVIDER=stripe_like
STRIPE_LIKE_PUBLIC_KEY=...
STRIPE_LIKE_SECRET_KEY=...
STRIPE_LIKE_WEBHOOK_SECRET=...
PAYMENT_EVENTS_ENDPOINT=...
PAYMENT_EVENTS_AUTH_TOKEN=...
```

### Step 7: Start an event receiver

Use the existing receiver from the platform if available. It must accept:

```text
POST <PAYMENT_EVENTS_ENDPOINT>
Authorization: Bearer <PAYMENT_EVENTS_AUTH_TOKEN>
Idempotency-Key: <webhook_event_id>
Content-Type: application/json
```

It should return any `2xx` status.

If you skip the event receiver, webhook DB processing can still complete, but you should expect an event publish failure log.

### Step 8: Start backend service

Reuse:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Section:

```text
10. Running the Project
```

Current command:

```bash
cd backend/services/payment-service
go run ./cmd/server
```

Expected startup signs:

| Log / behavior | Meaning |
|---|---|
| `payment.provider.registry_ready` with provider names | Provider env loaded successfully. |
| `payment.http.started` | HTTP server listening, default `:8080`. |
| `/healthz` returns `ok` | Base HTTP server is alive. |

### Step 9: Run webhook tests

Fastest verification:

```bash
cd backend/services/payment-service
go test ./internal/provider -run Webhook
go test ./internal/usecase -run HandleWebhook
go test ./internal/transport/http -run Webhook
```

What these tests prove:

| Test area | Proves |
|---|---|
| Provider webhook tests | Signature verification and payload normalization work. |
| Usecase webhook tests | Duplicate events do not publish twice and invalid signatures do not persist. |
| HTTP webhook tests | Raw body is preserved and oversized bodies are rejected. |

## 11. Running the Project

### Minimal test-only run

Use this when you do not want provider sandbox or local tunnel:

```bash
cd backend/services/payment-service
go test ./internal/provider -run Webhook
go test ./internal/usecase -run HandleWebhook
go test ./internal/transport/http -run Webhook
```

### Local HTTP run with generated signature

Prerequisites:

| Requirement | Why |
|---|---|
| MySQL running | Webhook stores and updates DB rows. |
| Migrations applied | `payment_webhook_events` table must exist. |
| Provider env loaded | Registry and webhook secret must be configured. |
| Matching payment row exists | Webhook needs payment lookup. |
| Event receiver configured | Full downstream event success needs receiver. |

Stripe-like local signed request example:

```bash
BODY='{"id":"evt_manual_capture_001","type":"payment_intent.succeeded","created":1710000000,"data":{"object":{"id":"pi_replace_from_payment","latest_charge":"ch_manual_001","amount":1000,"amount_received":1000,"currency":"inr","metadata":{"payment_id":"pay_replace_from_db","order_id":"ord_replace_from_db"}}}}'
TS="$(date +%s)"
SIG="$(printf '%s.%s' "$TS" "$BODY" | openssl dgst -sha256 -hmac "$STRIPE_LIKE_WEBHOOK_SECRET" -hex | awk '{print $2}')"

curl -X POST "http://127.0.0.1:8080/api/v1/webhooks/payments/stripe_like" \
  -H "Content-Type: application/json" \
  -H "Stripe-Signature: t=$TS,v1=$SIG" \
  --data "$BODY"
```

Expected response shape:

```json
{
  "success": true,
  "webhook_event_id": "whe_...",
  "provider_event_id": "evt_manual_capture_001",
  "processing_status": "processed",
  "payment_id": "pay_...",
  "payment_status": "captured"
}
```

Razorpay-like local signed request example:

```bash
BODY='{"event":"payment.captured","created_at":1710000000,"payload":{"payment":{"entity":{"id":"pay_gateway_manual_001","order_id":"order_gateway_001","amount":1000,"currency":"INR","notes":{"payment_id":"pay_replace_from_db","order_id":"ord_replace_from_db"}}}}}'
SIG="$(printf '%s' "$BODY" | openssl dgst -sha256 -hmac "$RAZORPAY_LIKE_WEBHOOK_SECRET" -hex | awk '{print $2}')"

curl -X POST "http://127.0.0.1:8080/api/v1/webhooks/payments/razorpay_like" \
  -H "Content-Type: application/json" \
  -H "X-Razorpay-Signature: $SIG" \
  --data "$BODY"
```

Beginner caution:

The `BODY` must not change after signature generation. Even whitespace changes can invalidate the signature because verification uses exact raw bytes.

### Verify DB after webhook

Check webhook event:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db -e "SELECT provider, provider_event_id, event_type, processed, processed_at FROM payment_webhook_events WHERE provider_event_id = 'evt_manual_capture_001';"
```

Check payment status:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db -e "SELECT payment_id, status, provider_payment_id, captured_amount FROM payments WHERE payment_id = 'pay_replace_from_db';"
```

Send same webhook again to verify duplicate behavior:

```bash
curl -X POST "http://127.0.0.1:8080/api/v1/webhooks/payments/stripe_like" \
  -H "Content-Type: application/json" \
  -H "Stripe-Signature: t=$TS,v1=$SIG" \
  --data "$BODY"
```

Expected duplicate behavior:

| Item | Expected |
|---|---|
| HTTP status | `200 OK` |
| `processing_status` | `duplicate` or safe no-op depending existing row state |
| Payment status | Not updated a second time |
| Domain event publish | Not published again for duplicate processed event |

## 12. Ports & Networking

Common ports and networking are already explained in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Section:

```text
11. Ports & Networking
```

Task-specific ports and URLs:

| Component | Default / Example | Direction | Notes |
|---|---|---|---|
| `SERVICE_NAME` HTTP server | `http://127.0.0.1:8080` | Inbound | Webhook route is hosted here. |
| Webhook endpoint | `/api/v1/webhooks/payments/{provider}` | Inbound | Provider calls this route. |
| MySQL | `127.0.0.1:3306` | Outbound from service | DSN controlled by `PAYMENT_MYSQL_DSN`. |
| Event receiver | `http://127.0.0.1:9080/events` | Outbound from service | Controlled by `PAYMENT_EVENTS_ENDPOINT`. |
| Provider API | HTTPS 443 | Outbound from service | Used by provider flows; webhook verification itself is local HMAC. |
| Public webhook URL | HTTPS tunnel or deployed URL | Inbound from provider | Needed for real provider sandbox webhook delivery. |

Gateway/proxy requirements:

| Requirement | Why |
|---|---|
| Preserve raw request body | Signature verification uses exact bytes. |
| Forward signature headers | Provider adapter needs `Stripe-Signature` or `X-Razorpay-Signature`. |
| Do not require buyer JWT on webhook route | Provider sends no user session. |
| Use HTTPS in non-local environments | Webhook payload and signatures should not travel over plain HTTP. |
| Keep server clock accurate | Stripe-like timestamp tolerance rejects stale/future timestamps. |

## 13. Common Errors & Fixes

Generic Go/MySQL/Docker/provider setup errors are already documented in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
```

`TASK_FILE_NAME` specific errors:

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `WEBHOOK_UNAVAILABLE` | Server started without webhook usecase, usually providers disabled | Set `PAYMENT_ALLOWED_PROVIDERS`, provider secrets, internal token, and event publisher env | Check startup log has provider names. |
| `UNSUPPORTED_PROVIDER` | Route provider not enabled or misspelled | Use `/api/v1/webhooks/payments/stripe_like` or enabled provider key | Keep provider dashboard URL aligned with `PAYMENT_ALLOWED_PROVIDERS`. |
| `INVALID_WEBHOOK_SIGNATURE` | Wrong webhook secret, changed body bytes, stale Stripe-like timestamp, or wrong header | Regenerate signature with exact raw body and correct secret | Never parse/format body before signing or forwarding. |
| `INVALID_WEBHOOK_BODY` | Body unreadable or exceeds `PAYMENT_WEBHOOK_MAX_BODY_BYTES` | Increase limit if legitimate, or send smaller payload | Keep limit reasonable, default 1 MiB. |
| `INVALID_WEBHOOK` | Signed payload is malformed or unsupported event type | Send supported provider event shape | Use provider tests as examples. |
| Valid webhook returns `ignored` | Payment not found, ambiguous lookup, amount/currency mismatch, or invalid transition | Ensure local payment exists and payload metadata references it | Create intent first and copy exact `payment_id`. |
| Event receiver logs missing | `PAYMENT_EVENTS_ENDPOINT` points to wrong service or receiver not running | Start receiver and verify bearer token | Keep local event sink available in provider mode. |
| Provider keeps retrying | Endpoint returned non-2xx before storing event | Check signature, body size, DB connectivity, provider route | Return 2xx for verified duplicates and stored ignored events. |
| Duplicate event publishes twice | Unique webhook index missing or not applied | Apply migration `001`; verify `uk_webhook_provider_event` | Always apply migrations before provider testing. |
| Stripe-like request fails after a few minutes | `Stripe-Signature` timestamp outside tolerance | Regenerate timestamp/signature or fix system clock | Use NTP and default `5m` tolerance. |
| `payment.webhook.event_publish_failed` log | DB processing succeeded but HTTP event receiver failed | Fix receiver URL/token/availability | Event receiver should return 2xx quickly. |

Supported provider webhook event examples:

| Provider | Supported event names |
|---|---|
| `stripe_like` | `payment_intent.requires_action`, `payment_intent.authorized`, `payment_intent.amount_capturable_updated`, `payment_intent.succeeded`, `payment_intent.payment_failed`, `payment_intent.canceled`, `refund.succeeded`, `refund.failed`, final `refund.updated` |
| `razorpay_like` | `payment.authorized`, `payment.captured`, `order.paid`, `payment.failed`, `refund.processed`, `refund.failed` |

## 14. Security & Best Practices

Task-specific security rules:

| Rule | Why it matters |
|---|---|
| Verify signature before persistence | Fake webhook could otherwise mark payment as paid. |
| Preserve exact raw body | HMAC verification breaks if body is reserialized. |
| Do not require user JWT for provider webhook | Provider does not have buyer session; signature is the auth mechanism. |
| Keep provider allowlist small | Unknown provider paths should fail. |
| Never log webhook secret or signature value | Signature data can be abused for replay/debug leakage. |
| Use timestamp tolerance where provider supports it | Reduces replay attack window. |
| Store provider event id with unique key | Duplicate retries should not double-update payment/order. |
| Mark verified duplicates as successful | Avoid provider retry storms. |
| Validate amount and currency before final status update | Prevent mismatched/tampered financial state. |
| Use HTTPS outside localhost | Protect webhook payload and event publication. |

Best practices for junior developers:

- First run webhook unit tests.
- Then verify MySQL schema health.
- Then create a payment intent using Task 4 setup.
- Then send a signed webhook with exact matching `payment_id`.
- Then send the same webhook again and confirm duplicate behavior.
- Then check event receiver got exactly one `payment.events` message.
- Never use live provider keys while learning locally.

## 15. Missing or Misconfigured Things

`TASK_FILE_NAME` setup audit:

| Item | Current status | Risk | Suggested fix |
|---|---|---|---|
| Service `.env.example` | Not found in service folder | Beginners may miss webhook and event receiver env grouping | Add safe `.env.example` with placeholders. |
| Local event receiver | Not committed | Full webhook flow cannot prove downstream publish locally | Add tiny dev event sink or compose profile. |
| Local provider mock server | Not committed | Manual signed webhook testing uses curl/provider sandbox manually | Add mock provider that signs payloads. |
| Payment service Dockerfile/compose | Not committed | Full stack setup remains manual | Add service Dockerfile and compose with MySQL, event sink, provider mock. |
| Migration runner | Not found | Manual SQL loops can be applied out of order | Add Make target or migration tool. |
| Public tunnel setup doc | Not committed | Real provider sandbox cannot call local `localhost` | Add provider sandbox and tunnel onboarding doc. |
| Broker integration | Not present in current code | Diagrams may confuse readers expecting Kafka/RabbitMQ | Keep docs clear that current implementation uses HTTP publisher. |
| Provider SDK | Not installed | Signature handling is custom standard-library code | Add official SDK only if production provider selection requires it. |

## 16. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `2. Tech Stack` | Same Go backend stack already explained. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3. Required Software` | Same Git, Go, MySQL, Docker, curl setup. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Dependency Management` | Same Go modules dependency commands. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `5. Database Setup` | Same MySQL local/Docker setup and DSN format. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `6. Redis / Queue / External Services` | Redis/Kafka/RabbitMQ absence and HTTP event publisher baseline already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. Environment Variables` | Env loading and full variable reference already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `8. Docker Setup` | Same MySQL-only Docker baseline reused. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `11. Ports & Networking` | Same backend, DB, event receiver, and provider networking notes. |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` | Same `payment_db`, `payments`, `payment_webhook_events`, and migration verification reused. |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `11. Common Errors & Fixes` | Schema-specific migration/index issues reused. |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `2. Tech Stack` | Provider abstraction components already explained. |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `6. Redis / Queue / External Services` | Provider APIs and HTTP event receiver setup already explained. |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `7. Environment Variables` | Provider-enabled env group already explained. |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `10. Running the Project` | Provider-enabled run pattern already documented. |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `11. Common Errors & Fixes` | Provider config, webhook signature, and outbound provider errors reused. |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | `9. Local Development Setup` | Create-intent setup is required before realistic webhook status update testing. |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | `10. Running the Project` | Provider-enabled service run and DB row verification are reused. |

## 17. Final Checklist

Setup checklist:

- [ ] Previous dependency docs were read before this file.
- [ ] No duplicate Go/MySQL/Docker install documentation added.
- [ ] `backend/services/payment-service/go.mod` dependencies understood.
- [ ] MySQL is running.
- [ ] All current `.up.sql` migrations are applied.
- [ ] `payment_webhook_events` table exists.
- [ ] `uk_webhook_provider_event` unique index exists.
- [ ] Provider env is loaded.
- [ ] Webhook secret matches provider/dashboard/local signature generator.
- [ ] `PAYMENT_WEBHOOK_MAX_BODY_BYTES` is positive.
- [ ] `PAYMENT_EVENTS_ENDPOINT` and `PAYMENT_EVENTS_AUTH_TOKEN` are configured when providers are enabled.
- [ ] Event receiver returns 2xx for `payment.events`.
- [ ] Matching local payment row exists before manual webhook.
- [ ] `go test ./internal/provider -run Webhook` passes.
- [ ] `go test ./internal/usecase -run HandleWebhook` passes.
- [ ] `go test ./internal/transport/http -run Webhook` passes.
- [ ] Duplicate webhook returns safe success and does not republish event.

Mental model:

`TASK_FILE_NAME` setup succeeds when five things line up: provider is enabled, raw body and signature header reach the service unchanged, webhook secret matches, MySQL schema is ready, and the webhook payload can match an existing payment. If any one is wrong, the safest debugging path is tests first, then env, then DB indexes, then signed curl/provider dashboard delivery.
