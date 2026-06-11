# Project Dependency & Setup Guide

## 1. Project Overview

This guide is generated for:

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Payment Service` |
| `TASK_FILE_NAME` | `task6.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task6_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

`TASK_FILE_NAME` ka scope refund flow hai. Simple Hinglish me: captured payment ke against full ya partial refund request create karna, duplicate/over-refund avoid karna, finance review policy follow karna, provider refund submit karna, aur verified success/failure ke baad `payments` and `refunds` records update karna.

Important boundary:

- Original `INPUT_FILE_PATH` modify nahi kiya gaya.
- Ye file business logic rewrite nahi karti.
- Common clone, Go install, MySQL install, Docker MySQL, provider setup, webhook setup, and migration basics duplicate nahi kiye gaye.
- Current repo me refund flow code already exists under `backend/services/payment-service/internal/usecase/refund_payment.go`.
- Current HTTP routes are:
  - `POST /api/v1/payments/{payment_id}/refund`
  - `GET /api/v1/refunds/{refund_id}`
  - `POST /internal/v1/refunds/{refund_id}/review`

Previous dependency files checked in the same `SERVICE_NAME` folder:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
```

Use this file as an incremental setup guide for `TASK_FILE_NAME`. Pehle previous setup docs follow karo, phir yahan refund-specific checks apply karo.

## 2. Tech Stack

### Reused technology documentation

Common stack already explained hai:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
```

Refer sections:

- `task1_Dependency.md` -> `2. Tech Stack`
- `task1_Dependency.md` -> `3. Required Software`
- `task1_Dependency.md` -> `4. Dependency Management`
- `task2_Dependency.md` -> `5. Database Setup`
- `task3_Dependency.md` -> `6. Redis / Queue / External Services`
- `task5_Dependency.md` -> `7. Redis / Queue / External Services`

### `TASK_FILE_NAME` specific technology impact

| Technology / Component | Required? | Why used in `TASK_FILE_NAME` | Beginner explanation |
|---|---:|---|---|
| Go usecase layer | Required | `RefundPaymentUsecase` refund validation, reservation, review, provider call, and event publish coordinate karta hai | Usecase workflow ka brain hota hai. HTTP handler sirf request ko usecase input me convert karta hai. |
| Go `net/http` transport | Required | Refund, get refund, and review routes expose karta hai | Current service Gin/Fiber/Echo use nahi karta; Go standard HTTP server use hota hai. |
| MySQL transactions | Required | Payment row lock, open refund reservation, and final amount update atomic rakhne ke liye | Transaction ka matlab related DB changes ek saath safe commit/rollback hote hain. Payment system me ye mandatory hai. |
| `refunds` table | Required | Har refund request ka immutable/auditable record store karta hai | Refund amount/reason/idempotency key insert ke baad edit nahi hone chahiye. |
| Provider registry | Required for provider submission | Payment ke `provider` field se `stripe_like` or `razorpay_like` adapter resolve hota hai | Registry provider adapter ki address book jaisa hai. |
| Stripe-like/Razorpay-like HTTP providers | Required when refund is approved or auto-approved | Actual external refund API call karne ke liye | SDK install nahi hai; current adapters direct HTTP requests use karte hain. |
| HTTP event publisher | Required when providers enabled | Refund lifecycle events `PAYMENT_EVENTS_ENDPOINT` par bhejta hai | Current repo Kafka/RabbitMQ runtime use nahi karta; HTTP event receiver reuse hota hai. |
| Provider refund webhooks | Required for async final outcome | Provider later `refund.succeeded` / `refund.failed` jaisa event bhej sakta hai | Webhook final truth provide karta hai, especially processing refunds ke liye. |
| Manual review policy | Required business control | High-value refunds ko provider call se pehle review me hold karta hai | Default threshold `0` hai, jo fail-safe behavior hai: every refund needs review. |

No new third-party Go module is introduced by `TASK_FILE_NAME`.

## 3. Required Software

No new base software installation is introduced beyond previous guides.

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
| Go | Required | Reused | Needed to run service and refund tests. |
| MySQL | Required | Reused | Refund flow reads/updates `payments`, `refunds`, and webhook audit rows. |
| MySQL client | Required for setup | Reused | Needed to apply and verify refund-related migrations. |
| Docker | Optional | Reused | Useful for local MySQL only. |
| curl | Recommended | Reused | Useful to verify refund HTTP routes. |
| Provider sandbox or local mock | Required for approved/auto-approved refund submission | Reused / task-specific | Without this, manual-review request can be created, but provider submission may fail. |
| HTTP event receiver | Required when providers enabled | Reused | Refund lifecycle events are posted to `PAYMENT_EVENTS_ENDPOINT`. |
| Provider webhook delivery setup | Required for async refund finalization | Reused from Task 5 | Needed when provider returns `processing` and final result arrives later. |

If you only run unit tests, real provider account, tunnel, Redis, Kafka, RabbitMQ, and Kubernetes are not required.

## 4. Dependency Management

This remains a Go modules project.

Common Go module explanation already exists in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Section:

```text
4. Dependency Management
```

Current relevant dependency files:

| File | Purpose |
|---|---|
| `backend/services/payment-service/go.mod` | Go module and dependency versions. |
| `backend/services/payment-service/go.sum` | Go module checksum lock file. |

Current direct external dependency remains:

```text
github.com/go-sql-driver/mysql v1.10.0
```

Task-specific code paths:

| Path | Purpose |
|---|---|
| `backend/services/payment-service/internal/usecase/refund_payment.go` | Refund request, review, provider submission, and lifecycle event orchestration. |
| `backend/services/payment-service/internal/usecase/refund_payment_test.go` | Refund policy, auto-approval, idempotency, and maker-checker tests. |
| `backend/services/payment-service/internal/domain/refund.go` | Refund entity, statuses, validation, and reservation rules. |
| `backend/services/payment-service/internal/domain/refund_webhook.go` | Verified refund webhook outcome handling. |
| `backend/services/payment-service/internal/repository/mysql_payment_repository.go` | Refund reservation, review, processing, completion, and webhook DB transactions. |
| `backend/services/payment-service/internal/transport/http/handler.go` | Refund HTTP routes, auth, and error mapping. |
| `backend/services/payment-service/internal/provider/types.go` | Provider `Refund(...)` request/response contract. |
| `backend/services/payment-service/internal/provider/stripe_like.go` | Stripe-like refund HTTP adapter. |
| `backend/services/payment-service/internal/provider/razorpay_like.go` | Razorpay-like refund HTTP adapter. |

Useful task-specific tests:

```bash
cd backend/services/payment-service
go test ./internal/usecase -run Refund
go test ./internal/domain -run Refund
go test ./internal/transport/http -run Refund
go test ./internal/provider -run Refund
```

Full current service test command remains:

```bash
go test ./...
```

## 5. Database Setup

Common MySQL install, Docker MySQL, DSN format, and broad migration commands are already documented in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
```

Sections:

- `task1_Dependency.md` -> `5. Database Setup`
- `task2_Dependency.md` -> `5. Database Setup`

### `TASK_FILE_NAME` database dependencies

Refund flow uses the existing `payment_db` MySQL database.

| Database object | Required? | Why `TASK_FILE_NAME` needs it |
|---|---:|---|
| `payments` | Yes | Original payment, provider, captured amount, and refunded amount track karta hai. |
| `refunds` | Yes | Refund request, provider refund id, status, idempotency, review audit store karta hai. |
| `payment_webhook_events` | Yes for async outcome | Provider refund webhook dedupe and audit ke liye. |
| `uk_refunds_idempotency` | Yes | Same `(payment_id, idempotency_key)` duplicate refund prevent karta hai. |
| `idx_refunds_payment_status` | Yes | Open reservation amount calculate karne me help karta hai. |
| `review_reason` column | Yes in current code | Review approve/reject reason audit ke liye required hai. |
| `idx_refunds_provider_refund` | Yes for webhook lookup | Provider refund id se refund row find karne ke liye. |

### Required migrations

For current local service run, apply all current `.up.sql` migrations in order. `TASK_FILE_NAME` specifically depends on these refund-related pieces:

| Migration | Refund relevance |
|---|---|
| `001_create_payment_tables.up.sql` | Creates baseline `refunds` table, `payments.refunded_amount`, refund statuses, and idempotency key. |
| `003_add_refund_review_reason.up.sql` | Adds `refunds.review_reason`; current repository query expects this column. |
| `004_add_refund_provider_lookup_index.up.sql` | Adds `idx_refunds_provider_refund` for provider refund webhook lookup. |

Apply all current migrations:

```bash
cd backend/services/payment-service
for file in migrations/*.up.sql; do
  mysql -h 127.0.0.1 -P 3306 -u root -proot < "$file"
done
```

Beginner note:

MySQL install and Docker container setup same hai as previous docs. Yahan sirf refund-specific schema verification add ho rahi hai.

### Verify refund schema

Verify `review_reason` exists:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db \
  -e "SHOW COLUMNS FROM refunds LIKE 'review_reason';"
```

Verify provider refund lookup index:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db \
  -e "SHOW INDEX FROM refunds WHERE Key_name = 'idx_refunds_provider_refund';"
```

Verify refund idempotency index:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db \
  -e "SHOW INDEX FROM refunds WHERE Key_name = 'uk_refunds_idempotency';"
```

Verify service-level schema health:

```bash
curl -s http://127.0.0.1:8080/internal/v1/payment-schema/health
```

### Credentials placement

Database credentials are not new for `TASK_FILE_NAME`.

Refer:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Section:

```text
7. Environment Variables
```

Current code reads MySQL DSN from:

| Variable | Required? | Example |
|---|---:|---|
| `PAYMENT_MYSQL_DSN` | Required | `root:root@tcp(127.0.0.1:3306)/payment_db?parseTime=true&charset=utf8mb4,utf8&loc=UTC` |
| `MYSQL_DSN` | Optional fallback | Same format |

## 6. Redis / Queue / External Services

### Redis / Kafka / RabbitMQ / NATS

No Redis, Kafka, RabbitMQ, or NATS dependency is introduced by `TASK_FILE_NAME`.

Reuse:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
```

Sections:

- `task1_Dependency.md` -> `6. Redis / Queue / External Services`
- `task3_Dependency.md` -> `6. Redis / Queue / External Services`
- `task5_Dependency.md` -> `7. Redis / Queue / External Services`

Current implementation publishes refund domain events through HTTP, not Kafka/RabbitMQ runtime.

### Payment provider refund APIs

Provider setup and credential basics are already explained in:

```text
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
```

Sections:

- `6. Redis / Queue / External Services`
- `7. Environment Variables`

`TASK_FILE_NAME` adds the refund operation on those same providers:

| Provider key | Refund API used by current adapter | Required credentials | Notes |
|---|---|---|---|
| `stripe_like` | `POST /v1/refunds` | `STRIPE_LIKE_SECRET_KEY`, `STRIPE_LIKE_WEBHOOK_SECRET` | Uses `payment_intent`/provider payment reference, amount, reason, metadata, and idempotency key. |
| `razorpay_like` | `POST /v1/payments/{provider_payment_id}/refund` | `RAZORPAY_LIKE_SECRET_KEY`, `RAZORPAY_LIKE_WEBHOOK_SECRET` | Uses provider payment id, amount, notes, and idempotency key header. |

Important task-specific provider notes:

| Area | Detail |
|---|---|
| Captured payment required | Local `payments.status` must be `captured` or `partially_refunded`. |
| Provider payment id required | `payments.provider_payment_id` must be present, because provider refund APIs need original charge/payment reference. |
| Amount format | Minor units only. Example: INR 250.00 = `25000`. |
| Idempotency | Same local refund idempotency key is passed to provider where supported. |
| Webhook finalization | Processing refunds become final through verified provider refund webhook or synchronous provider response. |
| Secrets | Provider secret key and webhook secret must stay backend-only. |

### HTTP event receiver

Refund flow reuses the same event receiver documented in:

```text
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
```

Sections:

- `task3_Dependency.md` -> `6. Redis / Queue / External Services -> HTTP event receiver`
- `task5_Dependency.md` -> `7. Redis / Queue / External Services -> HTTP event receiver`

Refund events published to topic `payment.events`:

| Event type | When emitted |
|---|---|
| `RefundRequested` | Refund row safely reserved. |
| `RefundApproved` | Manual or automatic approval happens. |
| `RefundRejected` | Reviewer rejects the request. |
| `RefundProcessing` | Provider accepts refund request and returns provider refund id. |
| `RefundSucceeded` | Provider success is verified and DB totals are updated. |
| `RefundFailed` | Provider final failure is verified or non-retryable provider call fails. |

## 7. Environment Variables

Common `.env` location, loading behavior, and broad env reference already exist in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
```

Sections:

- `task1_Dependency.md` -> `7. Environment Variables`
- `task3_Dependency.md` -> `7. Environment Variables`
- `task5_Dependency.md` -> `8. Environment Variables`

Current code reads env vars with `os.Getenv`. `.env` file automatic load nahi hota. Agar local `.env` use kar rahe ho:

```bash
cd backend/services/payment-service
set -a
. ./.env
set +a
```

### `TASK_FILE_NAME` new or task-specific variable

`PAYMENT_REFUND_MANUAL_REVIEW_THRESHOLD_MINOR` is the key task-specific configuration for refund policy. It is already listed in `task1_Dependency.md`, but its refund behavior matters most in `TASK_FILE_NAME`.

| Variable | Required? | Example | Purpose | Security notes |
|---|---:|---|---|---|
| `PAYMENT_REFUND_MANUAL_REVIEW_THRESHOLD_MINOR` | Optional | `0` | Refund amount at or above this minor-unit value needs manual review. `0` means every refund needs review. | Not a secret, but wrong value can bypass finance review unexpectedly. |

Minimal task-specific `.env` overlay:

```env
# Fail-safe local/manual-review mode: every refund is created as requested first.
PAYMENT_REFUND_MANUAL_REVIEW_THRESHOLD_MINOR=0
```

Auto-approval example for small refunds:

```env
# Example only: refunds below INR 500.00 auto-submit to provider.
PAYMENT_REFUND_MANUAL_REVIEW_THRESHOLD_MINOR=50000
```

Warning:

If threshold is `50000`, a refund amount of `25000` is below threshold, so current usecase auto-approves and calls provider immediately. Use this only when provider sandbox/mock, event receiver, and secrets are correctly configured.

### Existing variables required when provider refund submission should run

Do not duplicate full provider env docs here. Follow:

```text
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
```

At minimum, provider-enabled refund flow needs the existing variables below to be valid:

| Variable group | Why needed |
|---|---|
| `PAYMENT_ALLOWED_PROVIDERS`, `PAYMENT_DEFAULT_PROVIDER` | Refund usecase is only configured when provider registry has enabled providers. |
| `PAYMENT_INTERNAL_API_TOKEN` | Refund HTTP routes require `Authorization: Bearer <token>`. |
| Provider public/secret/webhook vars | Provider config validation and refund/webhook verification need them. |
| `PAYMENT_EVENTS_ENDPOINT`, `PAYMENT_EVENTS_AUTH_TOKEN` | Refund lifecycle events are posted to event receiver. |
| `PAYMENT_MYSQL_DSN` | Refund records and payment totals are stored in MySQL. |

### Required refund request headers

These are not env variables, but they are mandatory for local curl/API testing:

| Header | Required? | Example | Why |
|---|---:|---|---|
| `Authorization` | Yes | `Bearer <PAYMENT_INTERNAL_API_TOKEN>` | Internal/admin route protection. |
| `X-Actor-ID` | Yes | `finance_2` | Stored as `requested_by` or `reviewed_by`. |
| `X-Actor-Role` | Yes | `finance_admin` | Must be `admin`, `finance_admin`, or `superadmin`. |
| `X-Request-ID` | Optional | `req_refund_001` | Included in provider metadata/log correlation. |

## 8. Docker Setup

No new Dockerfile, docker-compose service, Redis, Kafka, RabbitMQ, or provider mock container is introduced by `TASK_FILE_NAME`.

Reuse:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
```

Sections:

- `task1_Dependency.md` -> `8. Docker Setup`
- `task5_Dependency.md` -> `9. Docker Setup`

Task-specific Docker notes:

| Component | Status | Notes |
|---|---|---|
| MySQL container | Reused | Same `payment_db` container/volume setup. |
| Payment service container | Not committed | Run service with `go run ./cmd/server` unless a Dockerfile is added later. |
| Provider mock container | Not committed | If you use a mock provider, run it separately and point provider base URL to it. |
| Event receiver container | Optional / external | Needed only if your event receiver runs in Docker. |

### Ports and networking

Detailed networking explanation already exists in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Section:

```text
11. Ports & Networking
```

`TASK_FILE_NAME` port table:

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Backend API | `8080` | Refund routes and health endpoints | Reused |
| MySQL | `3306` | `payment_db` | Reused |
| HTTP event receiver | `9080` example | Receives `payment.events` posts | Reused / external |
| Provider API HTTPS | `443` | Outbound Stripe-like/Razorpay-like refund calls | Reused / external |

No new inbound port is introduced by refund flow.

## 9. Local Development Setup

### Step 1: Read previous setup docs

Follow these first:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
```

Why:

| Previous file | Need before `TASK_FILE_NAME` |
|---|---|
| `task1_Dependency.md` | Clone, Go, MySQL, Docker, env, base run commands. |
| `task2_Dependency.md` | `payment_db` schema and migration verification. |
| `task3_Dependency.md` | Provider registry, provider env, and event receiver setup. |
| `task4_Dependency.md` | Create a payment intent before realistic refund testing. |
| `task5_Dependency.md` | Capture/payment webhook setup so payment becomes refundable. |

### Step 2: Go to service directory

```bash
cd backend/services/payment-service
```

### Step 3: Install dependencies only if not already done

No new dependency install is needed for `TASK_FILE_NAME`.

```bash
go mod download
```

### Step 4: Start MySQL and apply migrations

Use the previous MySQL/Docker setup. Then apply all current migrations:

```bash
for file in migrations/*.up.sql; do
  mysql -h 127.0.0.1 -P 3306 -u root -proot < "$file"
done
```

Verify refund-specific schema:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db \
  -e "SHOW COLUMNS FROM refunds LIKE 'review_reason';"

mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db \
  -e "SHOW INDEX FROM refunds WHERE Key_name = 'idx_refunds_provider_refund';"
```

### Step 5: Configure env

Load the provider-enabled env from previous docs, then add the refund policy overlay:

```env
PAYMENT_REFUND_MANUAL_REVIEW_THRESHOLD_MINOR=0
```

Beginner-friendly recommendation:

Keep `0` during first local test. Isse refund request `requested` status par rukegi and provider call turant nahi hoga.

### Step 6: Start event receiver if providers are enabled

Refund usecase publishes events when providers are enabled. Event receiver setup is reused from:

```text
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
```

Required behavior:

| Requirement | Detail |
|---|---|
| Method | `POST` |
| URL | Value of `PAYMENT_EVENTS_ENDPOINT` |
| Auth | `Authorization: Bearer <PAYMENT_EVENTS_AUTH_TOKEN>` |
| Success | Any 2xx response |

### Step 7: Start backend service

```bash
go run ./cmd/server
```

Verify health:

```bash
curl -s http://127.0.0.1:8080/healthz
```

Verify refund routes are available only after providers are enabled:

```bash
curl -s -X POST http://127.0.0.1:8080/api/v1/payments/pay_missing/refund
```

Without auth this should return a forbidden/authorization error, which confirms route exists.

### Step 8: Prepare a refundable payment

Refund requires a real local `payments` row with:

| Field | Required value |
|---|---|
| `status` | `captured` or `partially_refunded` |
| `captured_amount` | Greater than `0` |
| `refunded_amount` | Less than `captured_amount` |
| `provider_payment_id` | Non-empty |
| `provider` | Enabled provider key, like `stripe_like` |

Preferred flow:

1. Follow Task 4 create-intent setup.
2. Follow Task 5 webhook setup to move payment to `captured`.
3. Use that `payment_id` for refund testing.

Dev-only manual seed option:

Use this only for local manual-review route testing. It will not make a real provider refund valid unless your provider mock understands `ch_refund_local_1`.

```sql
USE payment_db;

INSERT INTO payments (
  payment_id, order_id, user_id, provider, provider_payment_id, provider_intent_id,
  status, currency, amount, captured_amount, refunded_amount, idempotency_key,
  created_at, updated_at
) VALUES (
  'pay_refund_local_1', 'ord_refund_local_1', 'usr_refund_local_1',
  'stripe_like', 'ch_refund_local_1', 'pi_refund_local_1',
  'captured', 'INR', 100000, 100000, 0, 'seed:pay_refund_local_1',
  UTC_TIMESTAMP(), UTC_TIMESTAMP()
);
```

## 10. Running the Project

### Run refund unit tests

```bash
cd backend/services/payment-service
go test ./internal/usecase -run Refund
go test ./internal/domain -run Refund
go test ./internal/transport/http -run Refund
```

### Manual-review refund request

This is the safest first manual API test because threshold `0` creates `requested` status and does not call provider immediately.

```bash
curl -s -X POST http://127.0.0.1:8080/api/v1/payments/pay_refund_local_1/refund \
  -H "Authorization: Bearer ${PAYMENT_INTERNAL_API_TOKEN}" \
  -H "X-Actor-ID: finance_requester_1" \
  -H "X-Actor-Role: finance_admin" \
  -H "X-Request-ID: req_refund_local_1" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": { "amount": 25000, "currency": "INR" },
    "reason": "One item cancelled before dispatch",
    "idempotency_key": "refund:pay_refund_local_1:item_cancelled_001"
  }'
```

Expected result:

| Field | Expected |
|---|---|
| HTTP status | `201 Created` for new request |
| `status` | `requested` when threshold is `0` |
| `requested_by` | Value from `X-Actor-ID` |

### Get refund

Replace `rfnd_...` with response `refund_id`.

```bash
curl -s http://127.0.0.1:8080/api/v1/refunds/rfnd_replace_me \
  -H "Authorization: Bearer ${PAYMENT_INTERNAL_API_TOKEN}" \
  -H "X-Actor-ID: finance_viewer_1" \
  -H "X-Actor-Role: finance_admin"
```

### Review refund

Approval triggers provider refund submission. Use this only when provider sandbox/mock and event receiver are ready.

```bash
curl -s -X POST http://127.0.0.1:8080/internal/v1/refunds/rfnd_replace_me/review \
  -H "Authorization: Bearer ${PAYMENT_INTERNAL_API_TOKEN}" \
  -H "X-Actor-ID: finance_approver_2" \
  -H "X-Actor-Role: finance_admin" \
  -H "Content-Type: application/json" \
  -d '{
    "decision": "approved",
    "reason": "Approved after order cancellation validation"
  }'
```

Maker-checker note:

For manual-review refunds, `ReviewedBy` cannot be same as `RequestedBy`. Agar same actor approve karega, API `MAKER_CHECKER_REQUIRED` return karegi.

### Verify DB after refund request

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db \
  -e "SELECT refund_id, payment_id, provider_refund_id, status, amount, currency, requested_by, reviewed_by, review_reason FROM refunds ORDER BY id DESC LIMIT 5;"
```

After successful provider-confirmed refund:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db \
  -e "SELECT payment_id, status, captured_amount, refunded_amount FROM payments WHERE payment_id='pay_refund_local_1';"
```

## 11. Common Errors & Fixes

Generic Go/MySQL/Docker/provider/webhook errors are already documented in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
```

Only `TASK_FILE_NAME` specific errors are listed here.

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `REFUND_UNAVAILABLE` | Providers are disabled, so refund usecase was not configured. | Set `PAYMENT_ALLOWED_PROVIDERS`, provider credentials, `PAYMENT_INTERNAL_API_TOKEN`, and event publisher env. | Use provider-enabled env from `task3_Dependency.md` before refund API testing. |
| `FORBIDDEN` | Missing bearer token, missing `X-Actor-ID`, or role is not `admin`/`finance_admin`/`superadmin`. | Send correct auth headers. | Keep a local curl template with all required headers. |
| `PAYMENT_NOT_REFUNDABLE` | Payment is not `captured`/`partially_refunded`, or `provider_payment_id` is empty. | Complete Task 4 + Task 5 flow, or seed a valid dev-only captured row. | Verify payment row before refund testing. |
| `REFUND_AMOUNT_NOT_AVAILABLE` | Amount exceeds captured minus refunded minus open reserved refunds, or currency mismatches. | Use remaining amount and same currency as payment. | Check existing open refunds before retrying. |
| `IDEMPOTENCY_CONFLICT` | Same refund key reused with different amount/reason/requester. | Use same payload for retries or create a new key for genuinely new refund. | Generate stable idempotency key per refund business reason. |
| `MAKER_CHECKER_REQUIRED` | Same actor requested and approved a manual-review refund. | Use a different finance approver. | Separate maker and checker accounts in local test data. |
| `INVALID_REFUND_TRANSITION` | Trying to review non-requested refund or complete invalid status. | Check current refund status with `GET /api/v1/refunds/{refund_id}`. | Follow lifecycle: `requested` -> `approved/rejected` -> `processing` -> `succeeded/failed`. |
| `Unknown column 'review_reason' in 'field list'` | Migration `003_add_refund_review_reason.up.sql` not applied. | Apply all `.up.sql` migrations. | Always run migrations in order on fresh DB. |
| Refund webhook ignored | Provider refund id/refund id/payment id does not match local refund row. | Ensure provider metadata contains local `refund_id` or provider refund id is stored. | Use current adapters and do not strip metadata in provider mock. |
| Provider refund call fails | Provider base URL, secret, provider payment id, or sandbox/mock setup invalid. | Check provider env and use a valid captured provider payment id. | Start with manual-review request before approval/provider call. |
| Event receiver logs missing | `PAYMENT_EVENTS_ENDPOINT` wrong or receiver not running. | Start event receiver and verify token. | Keep local event sink running in provider mode. |

## 12. Security & Best Practices

### Security audit observations

| Observation | Risk | Recommended action |
|---|---|---|
| Default manual review threshold is `0`. | Safe but can surprise developers because provider is not called until review. | Keep `0` for first local test; change intentionally for auto-submit tests. |
| Refund routes use shared internal bearer token plus actor headers. | Weak token or missing gateway controls can expose financial mutation. | Use strong 32+ char token and enforce gateway-level RBAC in deployed environments. |
| Provider-enabled startup requires event publisher config. | Missing event receiver blocks provider/refund usecase setup. | Configure `PAYMENT_EVENTS_ENDPOINT` and token before provider mode. |
| Current repo has no committed service Dockerfile/compose. | Local setup can differ across developers. | Add compose profile later with MySQL healthcheck, app container, event sink, and provider mock. |
| No durable outbox is present for refund lifecycle events. | Event publish failure after DB commit can require manual recovery. | Add outbox/retry publisher before production-grade financial eventing. |
| Dev docs contain example SQL seed rows. | Bad if copied into production. | Treat seed rows as local-only and never use fake provider ids in real provider environments. |

### Best practices for `TASK_FILE_NAME`

- Keep refund amounts in minor units only; float money calculation mat use karo.
- Use one stable idempotency key per business refund reason.
- Never edit `refunds.amount`, `refunds.currency`, `refunds.reason`, or `refunds.idempotency_key` after insert.
- Do not delete refund records; use status transitions for audit.
- Keep `PAYMENT_REFUND_MANUAL_REVIEW_THRESHOLD_MINOR=0` until finance approval flow is tested.
- Use different actors for requester and approver in manual-review testing.
- Do not log provider secret key, webhook secret, authorization header, card details, OTP, or raw unredacted provider payload.
- Verify provider refund webhooks with raw body signature validation from Task 5.
- Check `payments.refunded_amount` only after verified success, not after initial request.

## 13. Missing or Misconfigured Things

| Item | Current status | Impact | Suggested fix |
|---|---|---|---|
| Payment service Dockerfile/compose | Not committed | Full local stack is manual. | Add service Dockerfile and compose with MySQL, event receiver, and provider mock. |
| Migration runner | Raw SQL files only | Beginners must run migrations manually. | Add a small migration command/tool or documented make target. |
| Event receiver implementation | External/not included here | Provider-enabled refund flow needs a 2xx event sink. | Add local event-sink utility or compose service. |
| Provider mock service | Not committed | Approved refund manual testing needs real sandbox or custom mock. | Add provider mock for `stripe_like` and `razorpay_like` refund endpoints. |
| Durable outbox | Not present in current code | Event publish failures can lose downstream notification until repaired. | Add transactional outbox for payment/refund events. |
| Admin review UI/queue | Not part of current service | Manual review is API-only. | Build finance/admin UI separately. |
| gRPC server implementation | Task docs mention gRPC contract, current code exposes HTTP routes | Integrators expecting gRPC may be confused. | Add gRPC transport or keep API docs explicit that current local run uses HTTP. |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `2. Tech Stack` | Go, Go modules, MySQL driver, Docker baseline already explained. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3. Required Software` | Git, Go, MySQL, Docker, curl setup already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Dependency Management` | Common Go commands and module errors already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `5. Database Setup` | MySQL install, Docker MySQL, DSN format, and base DB setup already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. Environment Variables` | `.env` loading and broad env reference already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `8. Docker Setup` | MySQL-only Docker baseline already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `11. Ports & Networking` | Base HTTP, DB, event receiver, and provider networking already documented. |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` | `payment_db`, core tables, migration commands, and schema verification already documented. |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `6. Redis / Queue / External Services` | Provider APIs and HTTP event receiver setup already documented. |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `7. Environment Variables` | Provider-enabled env group already documented. |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | `9. Local Development Setup` | Create-intent flow needed before realistic captured payment testing. |
| `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md` | `7. Redis / Queue / External Services` | Webhook delivery, signatures, and event receiver behavior already documented. |
| `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md` | `11. Running the Project` | Webhook-style verification pattern reused for refund outcome events. |
| `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md` | `13. Common Errors & Fixes` | Provider webhook/signature/event receiver troubleshooting already documented. |

## 15. Final Checklist

- [ ] Previous dependency documentation checked before this file.
- [ ] Original `INPUT_FILE_PATH` not modified.
- [ ] No duplicate Go/MySQL/Docker/provider setup copied from previous files.
- [ ] Go dependencies verified; no new Go module required for `TASK_FILE_NAME`.
- [ ] MySQL is running and `PAYMENT_MYSQL_DSN` points to `payment_db`.
- [ ] All current `.up.sql` migrations applied.
- [ ] `refunds.review_reason` column exists.
- [ ] `idx_refunds_provider_refund` index exists.
- [ ] Provider-enabled env configured when refund submission/review should call provider.
- [ ] `PAYMENT_REFUND_MANUAL_REVIEW_THRESHOLD_MINOR` intentionally set.
- [ ] `PAYMENT_INTERNAL_API_TOKEN` is strong and loaded.
- [ ] `PAYMENT_EVENTS_ENDPOINT` receiver is running in provider mode.
- [ ] Refund request sends `Authorization`, `X-Actor-ID`, and `X-Actor-Role`.
- [ ] Test payment is `captured` or `partially_refunded` and has `provider_payment_id`.
- [ ] Idempotency key is stable and task-specific.
- [ ] Manual-review maker and checker are different actors.
- [ ] Refund API verified with local curl or unit tests.
- [ ] Logs checked for provider/event errors.
- [ ] No real secrets committed in `.env`, docs, SQL, or logs.
