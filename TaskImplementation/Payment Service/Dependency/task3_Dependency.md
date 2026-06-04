# Project Dependency & Setup Guide

## 1. Project Overview

This guide is generated for:

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Payment Service` |
| `TASK_FILE_NAME` | `task3.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task3_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

`TASK_FILE_NAME` ka scope gateway abstraction hai. Simple Hinglish me: `SERVICE_NAME` ko directly Stripe-like ya Razorpay-like provider code se tightly couple nahi karna. Is task me common provider contract, provider registry, provider-specific adapter config, webhook signature boundary, status normalization, idempotency, and fake-provider testing pattern explain hota hai.

Important boundary:

- Original `INPUT_FILE_PATH` modify nahi kiya gaya.
- Ye file business logic rewrite nahi karti.
- Shared Go, MySQL, Docker, migration, and common setup docs duplicate nahi kiye gaye.
- Current repo me provider abstraction code already exists under `backend/services/payment-service/internal/provider`.

## 2. Tech Stack

### Reused technology documentation

Common stack already explained hai:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Refer sections:

- `2. Tech Stack`
- `3. Required Software`
- `4. Dependency Management`
- `5. Database Setup`
- `6. Redis / Queue / External Services`
- `7. Environment Variables`

### `TASK_FILE_NAME` specific technology impact

| Technology / Component | Required? | Why used in `TASK_FILE_NAME` | Beginner explanation |
|---|---:|---|---|
| Go provider package | Required | Common provider interface, registry, validation, errors, and adapters live here | Ye abstraction layer hai. Usecase ko sirf common methods dikhte hain, provider-specific API details hidden rehti hain. |
| Go `net/http` client | Required when real provider enabled | Stripe-like/Razorpay-like APIs ko outbound HTTP request bhejne ke liye | HTTP client provider gateway se baat karta hai. |
| Stripe-like adapter | Optional until enabled | Payment intent, refund, and webhook verification for Stripe-style provider | Stripe-style provider `payment_intents` and signed webhooks use karta hai. |
| Razorpay-like adapter | Optional until enabled | Order-style payment intent, refund, and webhook verification for Razorpay-style provider | Razorpay-style provider order id frontend checkout ko deta hai. |
| Provider registry | Required | Configured provider name se correct adapter resolve karta hai | Registry ek address book jaisa hai: `stripe_like` maango, Stripe-like adapter milega. |
| HMAC SHA-256 webhook verification | Required for webhooks | Fake or tampered webhook block karne ke liye | Signature verify karke confirm hota hai ki webhook provider se aaya hai. |
| HTTP event publisher | Required when providers enabled | Payment/refund/webhook outcomes ko external event receiver par POST karta hai | Queue nahi hai; current code HTTP endpoint par events bhejta hai. |
| Fake provider test helper | Test-only | Unit/integration tests ko real gateway ke bina deterministic banata hai | Tests free and fast rehte hain because real card/payment sandbox hit nahi hota. |

No new third-party Go module is added by `TASK_FILE_NAME`. Current `go.mod` still has the MySQL driver only as the direct external dependency.

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
| Go | Required | Reused | Needed to run provider package tests and backend service. |
| MySQL | Required for server | Reused | Provider flows persist payment ids, attempts, idempotency, and webhook events in `payment_db`. |
| Docker | Optional | Reused | Useful for MySQL only; no new provider container is committed. |
| Internet access to provider API | Required only when using real gateway URLs | New / task-specific | Outbound HTTPS needed for real Stripe-like/Razorpay-like calls. |
| Local provider mock server | Optional | New / task-specific | Useful when `STRIPE_LIKE_BASE_URL` or `RAZORPAY_LIKE_BASE_URL` points to loopback. |
| HTTP event receiver | Required when providers enabled | Reused but important here | `PAYMENT_EVENTS_ENDPOINT` must accept `POST` and return 2xx. |

## 4. Dependency Management

This remains a Go modules project.

Refer:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Section:

```text
4. Dependency Management
```

Task-specific code paths:

| Path | Purpose |
|---|---|
| `backend/services/payment-service/internal/provider/types.go` | Common provider interface and request/response structs |
| `backend/services/payment-service/internal/provider/registry.go` | Provider registry/factory |
| `backend/services/payment-service/internal/provider/config.go` | Provider config validation |
| `backend/services/payment-service/internal/provider/stripe_like.go` | Stripe-like adapter |
| `backend/services/payment-service/internal/provider/razorpay_like.go` | Razorpay-like adapter |
| `backend/services/payment-service/internal/provider/webhook_signature.go` | HMAC webhook signature validation |
| `backend/services/payment-service/internal/provider/status_mapper.go` | Provider status normalization |
| `backend/services/payment-service/internal/provider/providertest/fake_provider.go` | Test fake provider |

Useful task-specific tests:

```bash
cd backend/services/payment-service
go test ./internal/provider ./internal/config
go test ./internal/usecase -run CreatePaymentIntent
```

No Stripe or Razorpay SDK install command is required for current code. The adapters use HTTP requests directly.

## 5. Database Setup

No new database engine or new migration is introduced by `TASK_FILE_NAME`.

Reuse:

```text
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
```

Sections:

- `5. Database Setup`
- `9. Local Development Setup`
- `11. Common Errors & Fixes`

Why DB still matters for this task:

| DB Area | Why provider abstraction needs it |
|---|---|
| `payments.provider` | Stores selected provider name, like `stripe_like` or `razorpay_like`. |
| `payments.provider_intent_id` | Stores provider-side intent/order id returned by adapter. |
| `payments.provider_payment_id` | Stores provider-side payment/charge id when available. |
| `payments.idempotency_key` | Prevents duplicate intent creation for same provider operation. |
| `payment_attempts.raw_provider_response` | Stores sanitized provider response for audit/debug. |
| `payment_webhook_events.provider_event_id` | Prevents duplicate webhook processing. |

For full current service run, apply all `.up.sql` migrations as already documented in `task2_Dependency.md`.

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

### Payment provider APIs

`TASK_FILE_NAME` introduces provider abstraction configuration for these provider names:

| Provider key | Default base URL | Required when enabled? | Notes |
|---|---|---:|---|
| `stripe_like` | `https://api.stripe.com` | Optional until configured | Uses form-encoded API calls and `Stripe-Signature` webhook header. |
| `razorpay_like` | `https://api.razorpay.com` | Optional until configured | Uses JSON/form provider calls and `X-Razorpay-Signature` webhook header. |

Important:

- If `PAYMENT_ALLOWED_PROVIDERS` is empty, providers are disabled by default.
- If providers are enabled, provider keys, internal API token, and event publisher config become mandatory.
- Remote provider base URLs must use HTTPS.
- `http://` provider base URL is accepted only for loopback/local testing.

### HTTP event receiver

Provider-enabled flows need an event receiver because `cmd/server` initializes `events.NewHTTPPublisher` when providers are configured.

Required behavior:

| Requirement | Detail |
|---|---|
| URL | `PAYMENT_EVENTS_ENDPOINT`, for example `http://127.0.0.1:9080/events` |
| Method | Must accept `POST` |
| Auth | Receives `Authorization: Bearer <PAYMENT_EVENTS_AUTH_TOKEN>` |
| Idempotency | Receives `Idempotency-Key` header |
| Success | Must return any 2xx status |
| Security | HTTPS required outside localhost/loopback |

Simple Hinglish:

Event receiver ek chhota HTTP service hai jo payment domain events receive karta hai. Agar provider enabled hai but event receiver config missing hai, app startup fail karega.

## 7. Environment Variables

Common env loading and full variable reference already exist in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Sections:

- `7. Environment Variables`
- `12. Common Errors & Fixes`

Current code reads env vars with `os.Getenv`. `.env` file automatic load nahi hota. Agar local `.env` use kar rahe ho, shell me load karo:

```bash
cd backend/services/payment-service
set -a
. ./.env
set +a
```

### Task-specific provider activation snippet

Use this only when you want `TASK_FILE_NAME` provider abstraction to be active:

```env
PAYMENT_DEFAULT_PROVIDER=stripe_like
PAYMENT_ALLOWED_PROVIDERS=stripe_like
PAYMENT_ALLOWED_CURRENCIES=INR,USD
PAYMENT_CAPTURE_MODE=automatic
PAYMENT_PROVIDER_TIMEOUT=5s
PAYMENT_WEBHOOK_TIMESTAMP_TOLERANCE=5m

PAYMENT_INTERNAL_API_TOKEN=change-this-internal-token-at-least-32-chars
PAYMENT_EVENTS_ENDPOINT=http://127.0.0.1:9080/events
PAYMENT_EVENTS_AUTH_TOKEN=event-publisher-token-at-least-32-characters
PAYMENT_EVENTS_TIMEOUT=5s

STRIPE_LIKE_PUBLIC_KEY=pk_test_replace_me
STRIPE_LIKE_SECRET_KEY=sk_test_replace_me
STRIPE_LIKE_WEBHOOK_SECRET=whsec_replace_me
STRIPE_LIKE_BASE_URL=
STRIPE_LIKE_TIMEOUT=5s
```

If using Razorpay-like instead:

```env
PAYMENT_DEFAULT_PROVIDER=razorpay_like
PAYMENT_ALLOWED_PROVIDERS=razorpay_like

RAZORPAY_LIKE_PUBLIC_KEY=rzp_test_replace_me
RAZORPAY_LIKE_SECRET_KEY=replace_me
RAZORPAY_LIKE_WEBHOOK_SECRET=replace_me
RAZORPAY_LIKE_BASE_URL=
RAZORPAY_LIKE_TIMEOUT=5s
```

### New / task-specific env behavior

| Variable | Required? | Purpose | Security note |
|---|---:|---|---|
| `PAYMENT_DEFAULT_PROVIDER` | Required when providers enabled | Default adapter selected when request has no provider | Must be in `PAYMENT_ALLOWED_PROVIDERS`. |
| `PAYMENT_ALLOWED_PROVIDERS` | Required to enable providers | Comma-separated provider allowlist | Do not allow unknown provider names. |
| `PAYMENT_ALLOWED_CURRENCIES` | Optional | Currency whitelist, default `INR,USD` | Keep business-approved currencies only. |
| `PAYMENT_CAPTURE_MODE` | Optional | `automatic` or `manual` | Current adapters validate value; capture method itself may still be provider-dependent. |
| `PAYMENT_INTERNAL_API_TOKEN` | Required when providers enabled | Bearer token for internal payment APIs | Minimum 32 chars; never commit real token. |
| `PAYMENT_EVENTS_ENDPOINT` | Required when providers enabled | Event receiver URL | HTTPS outside local loopback. |
| `PAYMENT_EVENTS_AUTH_TOKEN` | Required when providers enabled | Bearer token sent to event receiver | Minimum 32 chars; never commit. |
| `STRIPE_LIKE_*` | Required if `stripe_like` enabled | Public key, secret key, webhook secret, optional base URL/timeout | Secret key and webhook secret are backend-only. |
| `RAZORPAY_LIKE_*` | Required if `razorpay_like` enabled | Public key/key id, secret key, webhook secret, optional base URL/timeout | Secret key and webhook secret are backend-only. |

Common mistake:

If `PAYMENT_ALLOWED_PROVIDERS=stripe_like` is set but `PAYMENT_EVENTS_ENDPOINT` or `PAYMENT_INTERNAL_API_TOKEN` is missing, app startup fails before server listens.

## 8. Docker Setup

No new Dockerfile, docker-compose file, provider container, Redis, Kafka, or RabbitMQ setup is introduced by `TASK_FILE_NAME`.

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
| MySQL container | Reused | Same `payment_db` container from previous setup. |
| Provider mock container | Optional / not committed | Useful if you build a local Stripe-like/Razorpay-like test server. |
| Event receiver container | Optional / not committed | Needed only if your event receiver runs in Docker. |
| Docker network | No new committed network | If app and receiver run in different containers, use service names instead of `127.0.0.1`. |

### Ports & Networking

Detailed generic networking notes are reused from:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Section:

```text
11. Ports & Networking
```

`TASK_FILE_NAME` specific port view:

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Backend API | 8080 | Main HTTP server: health, schema, payment intent, webhook routes | Reused |
| MySQL | 3306 | Stores payments, attempts, refunds, webhooks, reconciliation records | Reused |
| HTTP event receiver | 9080 example | Receives payment/refund/webhook domain events over HTTP POST | Reused / required when providers enabled |
| Provider API HTTPS | 443 | Outbound calls to `https://api.stripe.com` or `https://api.razorpay.com` | New task-specific external dependency |
| Local provider mock | 9999 example | Optional loopback mock for `STRIPE_LIKE_BASE_URL` or `RAZORPAY_LIKE_BASE_URL` | New optional dev helper |

Port conflict notes:

- If `:8080` is busy, change `PAYMENT_HTTP_ADDR`.
- If MySQL `3306` is busy, change Docker port mapping and update `PAYMENT_MYSQL_DSN`.
- If event receiver runs in Docker, `127.0.0.1` from inside another container points to that container itself. Use Docker service name or host gateway as appropriate.
- Remote provider APIs should be reachable over outbound HTTPS; corporate firewall/proxy can cause `provider_unavailable`.

## 9. Local Development Setup

### Step 1: Read previous dependency documentation

Follow these first:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
```

Use them for Go install, MySQL install, Docker MySQL, migrations, and base service run.

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

Use the reused MySQL setup from `task1_Dependency.md`, then apply migrations from `task2_Dependency.md`.

### Step 5: Decide provider mode

| Mode | Env state | What works |
|---|---|---|
| Minimal mode | Leave `PAYMENT_ALLOWED_PROVIDERS` empty | Health, schema, state-machine endpoints. Provider-backed intent/webhook/refund flows unavailable. |
| Provider mode | Set provider env vars, internal token, event endpoint | Create intent, webhook, retry, and refund usecases can initialize. |

### Step 6: Load provider env only if needed

Provider mode needs the snippet in section `7. Environment Variables`.

### Step 7: Start backend service

```bash
go run ./cmd/server
```

### Step 8: Verify provider abstraction tests

```bash
go test ./internal/provider ./internal/config
```

Optional create-intent tests:

```bash
go test ./internal/usecase -run CreatePaymentIntent
```

## 10. Running the Project

### Minimal run

Use this when you only need base service without real gateway calls:

```bash
cd backend/services/payment-service
go mod download
go run ./cmd/server
```

Verify:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/internal/v1/payment-state-machine
curl http://localhost:8080/internal/v1/payment-schema/health
```

### Provider-enabled run

Before provider mode:

- MySQL must be running.
- Migrations must be applied.
- Provider env vars must be loaded.
- Event receiver must be reachable.
- Provider real API or local mock must be reachable.

Then run:

```bash
go run ./cmd/server
```

Create intent endpoint:

```text
POST http://localhost:8080/internal/v1/payment-intents
Authorization: Bearer <PAYMENT_INTERNAL_API_TOKEN>
Content-Type: application/json
```

Beginner note:

`Authorization` token yahan buyer token nahi hai. Ye internal service token hai, jo Order Service/API Gateway jaise trusted caller use karega.

Webhook endpoints:

| Provider | URL | Required signature header |
|---|---|---|
| `stripe_like` | `http://localhost:8080/api/v1/webhooks/payments/stripe_like` | `Stripe-Signature` |
| `razorpay_like` | `http://localhost:8080/api/v1/webhooks/payments/razorpay_like` | `X-Razorpay-Signature` |

## 11. Common Errors & Fixes

Generic Go/MySQL/Docker errors are already documented in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
```

`TASK_FILE_NAME` specific errors:

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `default provider is required when providers are enabled` | `PAYMENT_ALLOWED_PROVIDERS` set hai but `PAYMENT_DEFAULT_PROVIDER` empty hai | Default provider set karo | Provider env as a complete group load karo. |
| `default provider "x" is not allowed` | Default provider allowlist me nahi hai | `PAYMENT_ALLOWED_PROVIDERS` me same provider add karo | Provider names lowercase: `stripe_like`, `razorpay_like`. |
| `config missing for provider` | Allowlist me provider hai but required provider config missing hai | Corresponding `*_PUBLIC_KEY`, `*_SECRET_KEY`, `*_WEBHOOK_SECRET` set karo | Single provider se start karo. |
| `PAYMENT_INTERNAL_API_TOKEN is required when payment providers are enabled` | Provider mode enabled but internal token missing | 32+ char token set karo | Local `.env` me placeholder nahi, actual dev token rakho. |
| `PAYMENT_EVENTS_ENDPOINT is required when payment event publishing is enabled` | Provider mode event receiver ke bina start ho raha hai | Event receiver URL and auth token set karo | Provider mode start se pehle event receiver run/verify karo. |
| `PAYMENT_EVENTS_ENDPOINT must use HTTPS outside local testing` | Remote HTTP endpoint diya gaya | HTTPS URL use karo | HTTP only `localhost`/`127.0.0.1` ke liye. |
| `invalid payment webhook signature` | Wrong webhook secret, wrong header, body mutated, or old timestamp | Correct secret/header use karo and raw body preserve karo | Provider dashboard secret and env secret match rakho. |
| `provider_unknown_status` | Provider returned unsupported status | Mapper update karo or provider mock response fix karo | Unknown status ko success mat treat karo. |
| `provider_unavailable` | Provider API unreachable, timeout, bad base URL, DNS/firewall issue | Base URL, timeout, internet, and mock server check karo | Local mock URL loopback rakho; remote HTTPS use karo. |
| `PAYMENT_PROVIDER_UNAVAILABLE` | Provider registry empty and create intent called | Provider env vars enable karo | Minimal mode me provider APIs test mat karo. |

## 12. Security & Best Practices

Task-specific security rules:

| Rule | Why it matters |
|---|---|
| Secret keys backend-only rakho | Frontend ya Git me leak hua to real payments/refunds risk me aate hain. |
| Webhook secret always verify karo | Fake success webhook se fraud ho sakta hai. |
| Card number, CVV, OTP store/log mat karo | Hosted provider checkout/card vaulting use hona chahiye. |
| Amount server-side verify karo | Client-sent amount trust karne se undercharge/abuse ho sakta hai. |
| Idempotency key required rakho | Duplicate provider calls duplicate charges/refunds create kar sakte hain. |
| Provider base URL HTTPS rakho | Secrets and payment payload network me leak nahi hone chahiye. |
| Event receiver token strong rakho | Payment events spoof hone ka risk kam hota hai. |
| Sanitized provider payload only store karo | `raw_provider_response` debug-friendly ho but secrets-free ho. |

Best practices for juniors:

- Pehle minimal mode run karo, phir provider mode enable karo.
- Ek time par ek provider configure karo.
- Local testing ke liye provider test keys use karo, live keys nahi.
- `PAYMENT_ALLOWED_CURRENCIES` business-approved list tak limited rakho.
- Webhook tests me raw request body change mat karo; signature raw bytes par verify hota hai.
- Logs me `secret_key`, `webhook_secret`, `Authorization`, card data, OTP, token words avoid karo.

## 13. Missing or Misconfigured Things

Current setup audit:

| Item | Current status | Risk | Suggested fix |
|---|---|---|---|
| `.env.example` | Not found in service folder | Beginners may miss grouped provider requirements | Add safe `.env.example` with placeholders. |
| Event receiver service | Not committed | Provider-enabled startup can fail without endpoint | Add local dev event sink or compose profile. |
| Provider mock service | Not committed | Real provider sandbox needed for manual E2E | Add mock server for deterministic local tests. |
| Docker compose for full provider mode | Not committed | DB, app, event receiver, mock provider setup manual rahega | Add compose file after service containerization. |
| Secret manager integration | Not present | Production env may rely on raw env vars | Use Vault/AWS Secrets Manager/Kubernetes Secrets in deployment. |
| Migration runner | Not found | Manual migration ordering mistakes possible | Add migration tool/Make target as already noted in previous docs. |
| Real observability dashboards | Not present | Provider failures hard to track in prod | Add metrics for provider latency, error code, retryable failures. |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `2. Tech Stack` | Same Go backend stack already explained. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3. Required Software` | Same Git, Go, MySQL, Docker, curl setup. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Dependency Management` | Same Go modules commands and dependency behavior. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `5. Database Setup` | Same MySQL local/Docker setup and DSN format. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `6. Redis / Queue / External Services` | Redis/Kafka/RabbitMQ absence and HTTP event publisher already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. Environment Variables` | Full provider-enabled env reference already exists. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `8. Docker Setup` | Same MySQL-only Docker baseline. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `11. Ports & Networking` | Same backend, DB, event receiver, and provider API networking notes. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `12. Common Errors & Fixes` | Generic setup/provider startup issues already covered. |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` | Same `payment_db` schema and migration verification reused. |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `11. Common Errors & Fixes` | Schema-specific errors reused. |

## 15. Final Checklist

- [ ] Previous dependency documentation checked.
- [ ] `INPUT_FILE_PATH` reviewed.
- [ ] `OUTPUT_FILE_PATH` created without modifying original task file.
- [ ] No duplicate Go/MySQL/Docker installation documentation added.
- [ ] Provider abstraction files under `internal/provider` identified.
- [ ] No new third-party Go SDK dependency assumed.
- [ ] MySQL setup reused from previous guides.
- [ ] Provider mode env vars added only when needed.
- [ ] `PAYMENT_INTERNAL_API_TOKEN` is at least 32 characters.
- [ ] `PAYMENT_EVENTS_ENDPOINT` is reachable and returns 2xx on POST.
- [ ] `PAYMENT_EVENTS_AUTH_TOKEN` is at least 32 characters.
- [ ] Only intended providers are listed in `PAYMENT_ALLOWED_PROVIDERS`.
- [ ] Provider public/secret/webhook keys are set for enabled provider.
- [ ] Remote provider base URLs use HTTPS.
- [ ] Webhook signature headers are configured per provider.
- [ ] No card data, CVV, OTP, secret key, webhook secret, or auth token is logged/stored.
- [ ] `go test ./internal/provider ./internal/config` passes.
- [ ] `go test ./internal/usecase -run CreatePaymentIntent` passes if testing create-intent abstraction.
- [ ] Backend starts in minimal mode.
- [ ] Backend starts in provider mode when event receiver and provider config are available.
- [ ] Logs checked for provider registry and event publisher initialization errors.

Final beginner note:

`TASK_FILE_NAME` ka core setup point ye hai: provider abstraction code tabhi active and useful hota hai jab provider allowlist, provider secrets, internal API token, event receiver, MySQL schema, and webhook secrets saath me correctly configured hon. Pehle base service green karo, phir provider mode enable karo.
