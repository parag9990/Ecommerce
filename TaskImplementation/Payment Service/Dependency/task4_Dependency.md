# Project Dependency & Setup Guide

## 1. Project Overview

This guide is generated for:

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Payment Service` |
| `TASK_FILE_NAME` | `task4.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task4_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

`TASK_FILE_NAME` ka scope create payment intent flow hai. Simple Hinglish me: Order Service checkout ke time amount, currency, customer details, provider, and idempotency key ke saath `SERVICE_NAME` ko call karega. `SERVICE_NAME` local DB me payment/attempt record create karega, configured payment provider ko intent/order create karne ke liye call karega, aur frontend ko safe client payload return karega.

Important boundary:

- Original `INPUT_FILE_PATH` modify nahi kiya gaya.
- Ye file business logic rewrite nahi karti.
- Common Go, MySQL, Docker, migration, and provider setup repeat nahi kiya gaya.
- Current repo me create-intent code already exists under `backend/services/payment-service/internal/usecase/create_payment_intent.go`.
- HTTP route already exists at `POST /internal/v1/payment-intents`.

Beginner note:

Create payment intent se payment final paid nahi hota. Frontend provider checkout complete kar sakta hai, but final payment success webhook se confirm hoga. Webhook flow previous and later docs me cover hota hai; yahan focus sirf create-intent setup and verification par hai.

## 2. Tech Stack

### Reused technology documentation

Common stack already explained hai:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
```

Refer sections:

- `task1_Dependency.md` -> `2. Tech Stack`
- `task1_Dependency.md` -> `3. Required Software`
- `task1_Dependency.md` -> `4. Dependency Management`
- `task2_Dependency.md` -> `5. Database Setup`
- `task3_Dependency.md` -> `2. Tech Stack`
- `task3_Dependency.md` -> `6. Redis / Queue / External Services`

### `TASK_FILE_NAME` specific technology impact

| Technology / Component | Required? | Why used in `TASK_FILE_NAME` | Beginner explanation |
|---|---:|---|---|
| Go usecase layer | Required | `CreatePaymentIntentUsecase` request validate karta hai, DB records create karta hai, provider call karta hai | Usecase business workflow ka coordinator hota hai. Controller se request aati hai, usecase actual decision flow run karta hai. |
| Go HTTP transport | Required | `/internal/v1/payment-intents` endpoint expose karta hai | HTTP route Order Service/API Gateway jaise trusted caller ke liye entry point hai. |
| MySQL repository | Required | Payment and attempt rows idempotently persist karne ke liye | DB duplicate payment prevent karta hai using provider + idempotency key. |
| Provider registry | Required for successful intent | `stripe_like` or `razorpay_like` adapter resolve karta hai | Registry ek provider address book jaisa hai. Config me provider enable hoga to adapter milega. |
| Stripe-like / Razorpay-like HTTP providers | Required when enabled | Real provider intent/order create karne ke liye outbound HTTP call hoti hai | Provider external payment gateway hai. Local test me mock base URL use kar sakte ho. |
| Internal bearer token auth | Required | Create-intent endpoint ko public access se protect karta hai | `Authorization: Bearer <PAYMENT_INTERNAL_API_TOKEN>` ke bina endpoint reject karega. |
| HTTP event publisher | Required when providers are enabled | Server provider mode me event publisher initialize karta hai | Create-intent setup me provider enable karte hi event endpoint config bhi mandatory ho jata hai. |

No new Go module, SDK, database engine, cache, queue, or container is introduced by `TASK_FILE_NAME`.

## 3. Required Software

No new local software install is introduced beyond previous guides.

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
| Go | Required | Reused | Needed to run service and `CreatePaymentIntent` tests. |
| MySQL | Required | Reused | Intent creation writes `payments` and `payment_attempts`. |
| MySQL client | Required for setup | Reused | Needed to apply migrations and inspect rows. |
| Docker | Optional | Reused | Useful for local MySQL only. |
| curl | Recommended | Reused | Useful to verify `/internal/v1/payment-intents`. |
| Provider sandbox or local mock | Required for manual success path | Reused / task-specific | Create intent calls outbound provider API. |
| HTTP event receiver | Required when providers enabled | Reused / task-specific | Config validation requires event endpoint and auth token in provider mode. |

## 4. Dependency Management

This remains a Go modules project. `TASK_FILE_NAME` does not add a new package to `go.mod`.

Refer:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Section:

```text
4. Dependency Management
```

Current direct external dependency remains:

```text
github.com/go-sql-driver/mysql v1.10.0
```

Task-specific code paths:

| Path | Purpose |
|---|---|
| `backend/services/payment-service/internal/usecase/create_payment_intent.go` | Main create-intent workflow, validation, idempotency, provider call, safe response mapping |
| `backend/services/payment-service/internal/usecase/create_payment_intent_test.go` | Usecase tests for provider result persistence, idempotent replay, validation, provider failure |
| `backend/services/payment-service/internal/transport/http/handler.go` | HTTP handler and auth/error mapping for `/internal/v1/payment-intents` |
| `backend/services/payment-service/internal/transport/http/dto.go` | JSON request/response DTOs |
| `backend/services/payment-service/internal/provider/types.go` | Provider `CreateIntentRequest` and `CreateIntentResponse` contracts |
| `backend/services/payment-service/internal/provider/stripe_like.go` | Stripe-style create intent adapter |
| `backend/services/payment-service/internal/provider/razorpay_like.go` | Razorpay-style order create adapter |
| `backend/services/payment-service/internal/repository/mysql_payment_repository.go` | DB persistence and idempotency lookup |

Useful task-specific tests:

```bash
cd backend/services/payment-service
go test ./internal/usecase -run CreatePaymentIntent
go test ./internal/transport/http -run CreatePaymentIntent
go test ./internal/provider ./internal/config
```

Full current service test command remains:

```bash
go test ./...
```

## 5. Database Setup

No new database engine or new migration file is introduced by `TASK_FILE_NAME`.

Reuse:

```text
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
```

Sections:

- `5. Database Setup`
- `9. Local Development Setup`
- `11. Common Errors & Fixes`

### Why MySQL matters for create intent

Create intent writes and updates these already-created tables:

| Table | Used by `TASK_FILE_NAME`? | Purpose |
|---|---:|---|
| `payments` | Yes | Stores internal payment id, order id, user id, provider, amount, status, provider intent id, idempotency key |
| `payment_attempts` | Yes | Stores provider attempt id, attempt status, failure code/message, sanitized provider response |
| `refunds` | No direct create-intent use | Reused by later refund flow |
| `payment_webhook_events` | No direct create-intent use | Reused by webhook source-of-truth flow |
| `payment_reconciliations` | No direct create-intent use | Reused by reconciliation flow |

### Idempotency database behavior

`TASK_FILE_NAME` depends on this existing unique key:

```text
uk_payments_idempotency on (provider, idempotency_key)
```

Simple Hinglish:

Same provider aur same idempotency key se duplicate request aaye to system duplicate charge create nahi karega. Agar request data same hai, existing intent response replay ho sakta hai. Agar amount/order/user/currency different hai, API conflict return karegi.

### Required migrations

For current service run, apply all existing `.up.sql` migrations as already documented:

```bash
cd backend/services/payment-service
for file in migrations/*.up.sql; do
  mysql -h 127.0.0.1 -P 3306 -u root -proot < "$file"
done
```

Task-specific verification:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db -e "SHOW INDEX FROM payments WHERE Key_name='uk_payments_idempotency';"
```

After a successful create-intent request, inspect created rows:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db -e "SELECT payment_id, order_id, provider, provider_intent_id, status, amount, currency, idempotency_key FROM payments ORDER BY id DESC LIMIT 5;"
```

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db -e "SELECT attempt_id, payment_id, provider_attempt_id, status, failure_code FROM payment_attempts ORDER BY id DESC LIMIT 5;"
```

Security note:

`payment_attempts.raw_provider_response` should contain sanitized provider audit JSON only. It must not contain `client_secret`, card number, CVV, OTP, API token, webhook secret, or provider secret key.

## 6. Redis / Queue / External Services

### Redis / Kafka / RabbitMQ / NATS

No Redis, Kafka, RabbitMQ, or NATS dependency is introduced by `TASK_FILE_NAME`.

Reuse:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Section:

```text
6. Redis / Queue / External Services
```

### Payment provider API

Create intent requires a provider if you want successful `201 Created` responses.

| Provider key | Create-intent endpoint used by adapter | Default base URL | Required credentials |
|---|---|---|---|
| `stripe_like` | `POST /v1/payment_intents` | `https://api.stripe.com` | `STRIPE_LIKE_PUBLIC_KEY`, `STRIPE_LIKE_SECRET_KEY`, `STRIPE_LIKE_WEBHOOK_SECRET` |
| `razorpay_like` | `POST /v1/orders` | `https://api.razorpay.com` | `RAZORPAY_LIKE_PUBLIC_KEY`, `RAZORPAY_LIKE_SECRET_KEY`, `RAZORPAY_LIKE_WEBHOOK_SECRET` |

Provider setup is already explained in:

```text
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
```

Sections:

- `6. Redis / Queue / External Services`
- `7. Environment Variables`
- `9. Local Development Setup`
- `10. Running the Project`

Task-specific provider notes:

| Area | Detail |
|---|---|
| Amount | Must be minor unit, jaise INR 100.00 = `10000` paise. |
| Currency | Must be allowed by `PAYMENT_ALLOWED_CURRENCIES`, default `INR,USD`. |
| Idempotency | Sent to provider as `Idempotency-Key` for Stripe-like and `X-Razorpay-Idempotency-Key` for Razorpay-like. |
| Customer phone | Must be E.164 format, example `+919876543210`. |
| Metadata | Keys become lowercase and must not contain sensitive words like `token`, `secret`, `card_number`, `cvv`, `otp`. |
| Stripe-like replay | Existing Stripe-like intent may retrieve client secret again using provider API. |
| Razorpay-like payload | Response returns provider order id in `client_payload.provider_order_id`. |

### HTTP event receiver

No new event receiver implementation is added by `TASK_FILE_NAME`, but provider mode cannot start without event publisher config.

Reuse:

```text
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
```

Section:

```text
6. Redis / Queue / External Services
```

Required behavior remains:

| Requirement | Detail |
|---|---|
| URL variable | `PAYMENT_EVENTS_ENDPOINT` |
| Auth variable | `PAYMENT_EVENTS_AUTH_TOKEN` |
| Method | Must accept `POST` |
| Success | Must return 2xx |
| Security | HTTPS outside localhost/loopback |

## 7. Environment Variables

No brand-new env variable is introduced by `TASK_FILE_NAME`.

Use the existing provider-enabled env documentation from:

```text
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
```

Section:

```text
7. Environment Variables
```

### `TASK_FILE_NAME` required env group

For create-intent manual testing, these existing variables must be set as a group:

| Variable | Required? | Why `TASK_FILE_NAME` needs it | Security note |
|---|---:|---|---|
| `PAYMENT_INTERNAL_API_TOKEN` | Yes | `/internal/v1/payment-intents` rejects calls without bearer token | Minimum 32 chars; never commit real value. |
| `PAYMENT_DEFAULT_PROVIDER` | Yes | Used when request body omits `provider` | Must exist in `PAYMENT_ALLOWED_PROVIDERS`. |
| `PAYMENT_ALLOWED_PROVIDERS` | Yes | Enables provider registry | Keep allowlist small and intentional. |
| `PAYMENT_ALLOWED_CURRENCIES` | Optional | Validates request currency | Use business-approved ISO codes only. |
| `PAYMENT_CAPTURE_MODE` | Optional | Passed to provider create-intent request | Must be `automatic` or `manual`. |
| `PAYMENT_PROVIDER_TIMEOUT` | Optional | Shared outbound provider timeout | Keep finite to avoid stuck checkout. |
| `PAYMENT_EVENTS_ENDPOINT` | Yes when providers enabled | Required by server startup in provider mode | Use HTTPS outside local loopback. |
| `PAYMENT_EVENTS_AUTH_TOKEN` | Yes when providers enabled | Bearer token for event receiver | Minimum 32 chars; never commit. |
| `STRIPE_LIKE_*` | Yes if `stripe_like` enabled | Provider public/secret/webhook config | Secret key and webhook secret are backend-only. |
| `RAZORPAY_LIKE_*` | Yes if `razorpay_like` enabled | Provider public/secret/webhook config | Secret key and webhook secret are backend-only. |

### Important auth behavior

`PAYMENT_INTERNAL_API_TOKEN` is not a buyer token. Ye trusted internal caller token hai.

Create-intent HTTP request must include:

```text
Authorization: Bearer <PAYMENT_INTERNAL_API_TOKEN>
```

If token is missing in env or request header:

| Case | Result |
|---|---|
| Env token empty | Endpoint always returns `401 UNAUTHORIZED`. |
| Header missing | Endpoint returns `401 UNAUTHORIZED`. |
| Header token wrong | Endpoint returns `401 UNAUTHORIZED`. |
| Provider env missing but auth correct | Endpoint can return `503 PAYMENT_PROVIDER_UNAVAILABLE`. |

### Where `.env` should live

Reuse existing location:

```text
backend/services/payment-service/.env
```

Current code reads env vars via `os.Getenv`; `.env` automatic load nahi hota.

```bash
cd backend/services/payment-service
set -a
. ./.env
set +a
```

## 8. Docker Setup

No new Dockerfile, docker-compose service, Redis, Kafka, RabbitMQ, or provider container is introduced by `TASK_FILE_NAME`.

Reuse:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Sections:

- `5. Database Setup`
- `8. Docker Setup`
- `11. Ports & Networking`

Task-specific Docker notes:

| Item | Status | Detail |
|---|---|---|
| MySQL container | Reused | Same `ecommerce-payment-mysql` container can be used. |
| Payment service container | Not committed | Run Go service directly on host as previous docs explain. |
| Provider mock container | Optional / not committed | Helpful if `STRIPE_LIKE_BASE_URL` or `RAZORPAY_LIKE_BASE_URL` points to a local mock. |
| Event receiver container | Optional / not committed | Needed only if your event receiver runs in Docker. |
| Docker network | No new committed network | If app/provider mock/event receiver are containerized, use service names instead of `127.0.0.1`. |

### Ports & networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Backend API | 8080 | `/healthz`, schema APIs, create-intent API, webhook APIs | Reused |
| MySQL | 3306 | `payment_db` | Reused |
| HTTP event receiver | 9080 example | Receives payment/refund/webhook events | Reused / required in provider mode |
| Stripe-like provider API | 443 | Outbound HTTPS create intent | Reused provider dependency |
| Razorpay-like provider API | 443 | Outbound HTTPS create order | Reused provider dependency |
| Local provider mock | 9999 example | Optional create-intent mock | Optional task-specific helper |

Port conflict notes are already covered in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Section:

```text
11. Ports & Networking
```

## 9. Local Development Setup

### Step 1: Read previous dependency documentation

Follow these first:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
```

Use them for:

- Go install and module commands
- MySQL install/Docker setup
- Migration commands
- Provider env setup
- Event receiver setup
- Common errors

### Step 2: Go to service directory

```bash
cd backend/services/payment-service
```

### Step 3: Install dependencies

No new module was added by `TASK_FILE_NAME`.

```bash
go mod download
```

### Step 4: Start DB and apply migrations

Use the reused MySQL setup from `task1_Dependency.md`, then migration setup from `task2_Dependency.md`.

Quick current-service command:

```bash
docker start ecommerce-payment-mysql
for file in migrations/*.up.sql; do
  mysql -h 127.0.0.1 -P 3306 -u root -proot < "$file"
done
```

### Step 5: Configure provider mode

Create intent success path needs provider mode.

Use:

```text
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
```

Section:

```text
7. Environment Variables
```

Minimum provider-mode decisions:

| Decision | Recommended local value |
|---|---|
| Provider | Start with one provider, usually `stripe_like` or a local mock |
| Currency | `INR` or `USD`, matching `PAYMENT_ALLOWED_CURRENCIES` |
| Event endpoint | Local event sink, for example `http://127.0.0.1:9080/events` |
| Token | 32+ char dev-only internal token |
| Provider base URL | Empty for real sandbox, loopback URL for local mock |

### Step 6: Load environment variables

```bash
set -a
. ./.env
set +a
```

### Step 7: Start backend service

```bash
go run ./cmd/server
```

Expected useful startup logs:

```text
payment.provider.registry_ready
payment.http.started
```

If `payment.provider.registry_ready` shows empty providers, create-intent calls will not reach a provider and can return `PAYMENT_PROVIDER_UNAVAILABLE`.

### Step 8: Verify task-specific tests

```bash
go test ./internal/usecase -run CreatePaymentIntent
go test ./internal/transport/http -run CreatePaymentIntent
```

## 10. Running the Project

### Minimal service run

Minimal mode is useful for health/schema/state endpoints, but it is not enough for a successful create-intent provider call.

```bash
cd backend/services/payment-service
go mod download
docker start ecommerce-payment-mysql
for file in migrations/*.up.sql; do
  mysql -h 127.0.0.1 -P 3306 -u root -proot < "$file"
done
go run ./cmd/server
```

Verify base service:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/internal/v1/payment-schema/health
```

### Provider-enabled create-intent run

Before running this flow:

- MySQL is running.
- All migrations are applied.
- `.env` is loaded.
- `PAYMENT_INTERNAL_API_TOKEN` is 32+ chars.
- `PAYMENT_ALLOWED_PROVIDERS` and `PAYMENT_DEFAULT_PROVIDER` are set.
- Enabled provider credentials are set.
- `PAYMENT_EVENTS_ENDPOINT` is reachable and returns 2xx for POST.
- Provider sandbox or local mock is reachable.

Start service:

```bash
cd backend/services/payment-service
set -a
. ./.env
set +a
go run ./cmd/server
```

Task-specific create-intent request:

```bash
curl -i -X POST http://localhost:8080/internal/v1/payment-intents \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer change-this-internal-token-at-least-32-chars' \
  -H 'X-Request-ID: req_create_intent_001' \
  -d '{
    "order_id": "ord_123",
    "user_id": "usr_123",
    "amount": 1000,
    "currency": "INR",
    "idempotency_key": "payment_intent:ord_123:attempt_1",
    "provider": "stripe_like",
    "customer": {
      "name": "Test Buyer",
      "email": "buyer@example.com",
      "phone": "+919876543210"
    },
    "metadata": {
      "checkout_reference": "checkout_123"
    }
  }'
```

Expected success shape:

```json
{
  "payment_id": "pay_...",
  "order_id": "ord_123",
  "provider": "stripe_like",
  "provider_intent_id": "pi_...",
  "status": "requires_action",
  "amount": 1000,
  "currency": "INR",
  "client_payload": {
    "public_key": "pk_test_...",
    "client_secret": "..."
  }
}
```

Important:

- Status can be `initiated`, `requires_action`, `authorized`, `captured`, or `failed`, depending on provider response.
- `client_payload.client_secret` is safe for frontend SDK, but it should not be stored in DB logs or audit payloads.
- Same idempotency key with same data can return `200 OK` replay.
- Same idempotency key with different amount/order/user/currency returns conflict.

### Verify DB rows after create intent

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db -e "SELECT payment_id, order_id, provider, provider_intent_id, status, amount, currency, idempotency_key FROM payments WHERE order_id='ord_123';"
```

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db -e "SELECT attempt_id, payment_id, provider_attempt_id, status, failure_code FROM payment_attempts ORDER BY id DESC LIMIT 5;"
```

## 11. Common Errors & Fixes

Generic Go/MySQL/Docker/provider setup errors are already documented in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
```

`TASK_FILE_NAME` specific errors:

| Error / API code | Cause | Fix | Prevention |
|---|---|---|---|
| `UNAUTHORIZED` | Missing/wrong `Authorization: Bearer ...` header, or env token empty | Set `PAYMENT_INTERNAL_API_TOKEN` and send same token in request | Treat create-intent as internal API only. |
| `PAYMENT_PROVIDER_UNAVAILABLE` | Provider registry empty, usually `PAYMENT_ALLOWED_PROVIDERS` not set | Enable provider env vars or use a local mock provider config | Verify startup log `payment.provider.registry_ready`. |
| `UNSUPPORTED_PROVIDER` | Request provider not in configured registry | Use provider from `PAYMENT_ALLOWED_PROVIDERS` | Keep request provider and env provider names same. |
| `VALIDATION_ERROR: order_id is required` | Request body missing required field | Send `order_id` | Validate request in Order Service before calling. |
| `VALIDATION_ERROR: amount must be greater than zero` | Amount is zero/negative or not minor unit | Send positive integer minor amount | Never send decimal money values. |
| `VALIDATION_ERROR: currency ... is not allowed` | Currency not in `PAYMENT_ALLOWED_CURRENCIES` | Add approved ISO code or fix request currency | Keep allowed currency list aligned with business rules. |
| `VALIDATION_ERROR: customer.phone must be in E.164 format` | Phone sent as local number like `99999` | Use format like `+919876543210` | Normalize phone before request. |
| `VALIDATION_ERROR: metadata contains sensitive key` | Metadata key has words like `token`, `secret`, `card_number`, `cvv`, `otp` | Remove sensitive metadata | Never pass card/auth/provider secrets in metadata. |
| `IDEMPOTENCY_CONFLICT` | Same provider + idempotency key used with different order/user/amount/currency | Use deterministic key per unique checkout attempt | Key format like `payment_intent:<order_id>:attempt_<n>` helps. |
| `PAYMENT_INTENT_PROCESSING` | Same idempotency key exists but provider intent id not persisted yet | Retry after short delay or inspect logs | Avoid parallel duplicate calls from Order Service. |
| `PAYMENT_INTENT_FAILED` | Previous provider create call failed for same idempotency key | Start a new checkout attempt with new idempotency key | Do not reuse failed operation keys. |
| `PAYMENT_CLIENT_PAYLOAD_UNAVAILABLE` | Existing Stripe-like intent replay could not retrieve client payload | Check provider credentials/base URL and provider intent id | Keep provider retrieval API reachable for replay. |
| Provider error mapped to `provider_authentication` | Secret key invalid or sandbox/live mismatch | Fix provider secret/key pair | Store provider credentials as environment secrets. |
| Provider error mapped to `provider_rate_limited` | Too many provider calls | Back off and retry later | Use idempotency and avoid client-side duplicate submits. |
| Provider error mapped to `provider_unavailable` | Provider base URL/DNS/network/timeout issue | Check internet, firewall, mock server, timeout | Use loopback mock for local deterministic testing. |

## 12. Security & Best Practices

Task-specific rules:

| Rule | Why it matters |
|---|---|
| Keep create-intent endpoint internal | Public callers could create unwanted provider intents. |
| Use strong `PAYMENT_INTERNAL_API_TOKEN` | Weak token makes internal payment APIs easy to abuse. |
| Amount must come from server-side order calculation | Client amount can be manipulated. |
| Store amount in minor units | Floating point money bugs avoid hote hain. |
| Require idempotency key | Duplicate browser clicks/retries duplicate provider calls create kar sakte hain. |
| Never log provider secret, webhook secret, auth token, card number, CVV, OTP | Payment systems me secret leakage high-risk hota hai. |
| Keep raw provider audit payload sanitized | Debug data useful rahe, secrets DB me na jayein. |
| Use HTTPS for remote provider/event URLs | Secrets and payment payload network me leak nahi hone chahiye. |
| Use one provider first in local setup | Debugging simple hoti hai and config mistakes kam hoti hain. |

Best practices for junior developers:

- Pehle `go test ./internal/usecase -run CreatePaymentIntent` green karo.
- Then base service `/healthz` and `/internal/v1/payment-schema/health` green karo.
- Then provider env load karo.
- Then create-intent curl test run karo.
- Same curl request dobara run karke idempotent replay behavior check karo.
- Different amount with same idempotency key try karke conflict behavior samjho.
- Local testing me provider test keys or mock server use karo, live keys nahi.

## 13. Missing or Misconfigured Things

`TASK_FILE_NAME` setup audit:

| Item | Current status | Risk | Suggested fix |
|---|---|---|---|
| Service `.env.example` | Not found in service folder | Beginners may miss token/provider/event variable grouping | Add safe `.env.example` with placeholders. |
| Local event receiver | Not committed | Provider mode startup can fail without `PAYMENT_EVENTS_ENDPOINT` | Add tiny dev event sink or compose profile. |
| Local provider mock server | Not committed | Manual create-intent E2E may require real sandbox access | Add mock provider service for deterministic local tests. |
| Payment service Dockerfile/compose | Not committed | Full service setup remains manual | Add service Dockerfile and compose with MySQL, app, event sink, provider mock. |
| Migration runner | Not found | Manual SQL loop can be applied out of order or repeated | Add Make target or migration tool. |
| Internal auth distribution | Env-token only | Token rotation and service-to-service auth are manual | Use secret manager and gateway/service mesh auth in deployment. |
| Public API boundary | Create intent route is internal but exposed on same HTTP server | Misconfigured gateway could expose internal endpoint | Restrict route at gateway/network layer. |
| Provider sandbox docs | Not committed per provider | New developers may not know dashboard/key setup | Add provider-specific sandbox onboarding doc when real provider is chosen. |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `2. Tech Stack` | Same Go backend stack already explained. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3. Required Software` | Same Git, Go, MySQL, Docker, curl setup. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Dependency Management` | Same Go modules commands and common dependency failures. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `5. Database Setup` | Same MySQL install, Docker MySQL, DSN format, and migration location. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `6. Redis / Queue / External Services` | Redis/Kafka/RabbitMQ absence and HTTP event publisher baseline already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. Environment Variables` | Env loading behavior and broad variable reference already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `8. Docker Setup` | Same MySQL-only Docker baseline reused. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `11. Ports & Networking` | Same backend, DB, event receiver, provider API networking notes. |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` | Same `payment_db`, `payments`, `payment_attempts`, indexes, and migrations reused. |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `11. Common Errors & Fixes` | Schema-specific migration and index issues reused. |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `2. Tech Stack` | Provider abstraction components already explained. |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `6. Redis / Queue / External Services` | Provider APIs and event receiver setup already explained. |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `7. Environment Variables` | Provider-enabled env group already explained. |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `10. Running the Project` | Provider-enabled run pattern already documented. |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `11. Common Errors & Fixes` | Provider config, webhook signature, and outbound provider errors reused. |

## 15. Final Checklist

- [ ] Previous dependency documentation checked.
- [ ] `INPUT_FILE_PATH` reviewed.
- [ ] `OUTPUT_FILE_PATH` created without modifying original task file.
- [ ] No duplicate Go/MySQL/Docker installation documentation added.
- [ ] No new Go module incorrectly assumed.
- [ ] MySQL setup reused from previous dependency docs.
- [ ] All current `.up.sql` migrations applied.
- [ ] `uk_payments_idempotency` exists on `payments`.
- [ ] Provider mode env vars loaded only when create-intent provider flow is needed.
- [ ] `PAYMENT_INTERNAL_API_TOKEN` set to 32+ chars.
- [ ] Create-intent curl request sends `Authorization: Bearer <token>`.
- [ ] `PAYMENT_ALLOWED_PROVIDERS` contains requested provider.
- [ ] `PAYMENT_DEFAULT_PROVIDER` is part of allowed providers.
- [ ] Enabled provider public/secret/webhook keys are configured.
- [ ] `PAYMENT_EVENTS_ENDPOINT` is reachable and returns 2xx on POST.
- [ ] Remote provider and event URLs use HTTPS.
- [ ] Request amount is positive minor-unit integer.
- [ ] Request currency is in `PAYMENT_ALLOWED_CURRENCIES`.
- [ ] Customer phone uses E.164 format if provided.
- [ ] Idempotency key is deterministic and operation-specific.
- [ ] Metadata contains no sensitive keys or values.
- [ ] `go test ./internal/usecase -run CreatePaymentIntent` passes.
- [ ] `go test ./internal/transport/http -run CreatePaymentIntent` passes.
- [ ] Backend starts and logs `payment.http.started`.
- [ ] `/healthz` and `/internal/v1/payment-schema/health` verified.
- [ ] `/internal/v1/payment-intents` returns expected response in provider mode.
- [ ] DB rows checked in `payments` and `payment_attempts`.
- [ ] Logs checked for provider/auth/event publisher errors.

Final beginner note:

`TASK_FILE_NAME` ka setup success mainly five cheezon par depend karta hai: MySQL ready, migrations applied, provider configured, internal token sent, and event receiver reachable. Inme se ek bhi missing hua to create intent fail karega, even if code compile ho raha ho.
