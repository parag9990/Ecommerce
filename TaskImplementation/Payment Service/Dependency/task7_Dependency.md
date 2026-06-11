# Project Dependency & Setup Guide

## 1. Project Overview

This guide is generated for:

| Variable | Value |
|---|---|
| `SERVICE_NAME` | Use the prompt variable `SERVICE_NAME` |
| `TASK_FILE_NAME` | Use the prompt variable `TASK_FILE_NAME` |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `{TASK_FILE_NAME without .md}_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

`TASK_FILE_NAME` ka scope failed payment retry handling hai. Simple Hinglish me: buyer ka pehla payment fail ho gaya ho to same order ke liye safe retry allow karna, but duplicate charge, double-click charge, aur late webhook confusion avoid karna.

Current repo inspection ke according retry runtime implementation available hai:

| Runtime path | Setup relevance |
|---|---|
| `backend/services/payment-service/internal/domain/payment_retry.go` | Retry policy, retry request validation, and new retry payment lineage model. |
| `backend/services/payment-service/internal/usecase/retry_payment_intent.go` | Retry orchestration, provider intent creation, replay handling, and provider failure handling. |
| `backend/services/payment-service/internal/repository/mysql_payment_repository.go` | MySQL transaction, row locks, active-attempt checks, retry reservation, and replay lookup. |
| `backend/services/payment-service/internal/transport/http/handler.go` | Public retry API route and auth/idempotency validation. |
| `backend/services/payment-service/internal/config/config.go` | Retry environment variables and validation. |
| `backend/services/payment-service/migrations/005_add_payment_retry_lineage.up.sql` | Task-specific MySQL schema migration. |
| `api/master-api.json` | Public route contract and `RetryPaymentRequest` with required `idempotency_key`. |

Important boundary:

- Original `INPUT_FILE_PATH` modify nahi kiya gaya.
- Ye file business logic rewrite nahi karti.
- Common clone, Go install, MySQL install, Docker MySQL, provider setup, webhook setup, and generic troubleshooting duplicate nahi kiya gaya.
- Only new or changed setup for `TASK_FILE_NAME` detail me explain kiya gaya hai.

Previous dependency files checked in the same `SERVICE_NAME` folder:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
TaskImplementation/{SERVICE_NAME}/task6_Dependency.md
```

Use this file as an incremental onboarding guide. Pehle previous setup docs follow karo, phir yahan retry-specific migration, env, headers, and verification steps apply karo.

## 2. Tech Stack

### Reused technology documentation

Common stack already explained hai:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Sections:

- `2. Tech Stack`
- `3. Required Software`
- `4. Dependency Management`
- `5. Database Setup`
- `6. Redis / Queue / External Services`
- `7. Environment Variables`
- `8. Docker Setup`

Provider and payment intent setup already explained hai:

```text
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
```

### `TASK_FILE_NAME` specific technology impact

| Technology / service | Required? | Why this task uses it | Beginner explanation |
|---|---:|---|---|
| Go `net/http` | Required | `POST /api/v1/payments/{payment_id}/retry` route expose hota hai. | Go ka built-in HTTP server package hai. |
| Go Modules | Required | Existing module dependencies manage karta hai. | `go.mod` dependency list hai, `go.sum` checksum lock hai. |
| MySQL | Required | Retry lineage, idempotency, and concurrent retry lock ke liye DB constraints chahiye. | MySQL relational database hai jisme payment rows tables me store hote hain. |
| MySQL transactions and row locks | Required | Same order ke parallel retry attempts ko block karta hai. | Transaction ka matlab DB work all-or-nothing commit hota hai. |
| `github.com/go-sql-driver/mysql` | Required, reused | Go service ko MySQL se connect karata hai. | Existing dependency hai; new `go get` required nahi hai. |
| Provider registry | Required for live retry API | Retry new provider intent create karta hai. | Stripe-like/Razorpay-like adapter gateway calls isolate karte hain. |
| HTTP event publisher | Required when providers are enabled | Provider mode startup validation event endpoint require karta hai. | Events dusre services ko HTTP POST ke through bheje jate hain. |
| Redis / Kafka / RabbitMQ / NATS | Not required by current code | `TASK_FILE_NAME` code broker runtime use nahi karta. | Install mat karo unless platform separately require kare. |
| Docker | Optional, reused | Local MySQL quickly run karne ke liye useful. | New retry-specific container nahi add hua. |

## 3. Required Software

Common software install steps duplicate nahi kiye gaye. Refer:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Sections:

- `3. Required Software`
- `4. Dependency Management`
- `5. Database Setup`
- `8. Docker Setup`

For `TASK_FILE_NAME`, ensure these are available:

| Software | Required? | Status | Why needed for retry setup |
|---|---:|---|---|
| Go `1.26.3` compatible toolchain | Required | Reused | Build and test service. |
| MySQL server | Required | Reused | Stores retry lineage columns and constraints. |
| MySQL client | Required for setup | Reused | Apply and verify migration `005`. |
| curl or API client | Required for manual API verification | Reused | Send retry request and health checks. |
| Docker | Optional | Reused | Only for local MySQL if you do not install MySQL directly. |
| Provider sandbox/mock | Required for live API success | Reused | Retry endpoint creates a new provider intent. |
| HTTP event receiver | Required when providers enabled | Reused | Config validation needs `PAYMENT_EVENTS_ENDPOINT` and token in provider mode. |

If you only run unit tests, real provider account, Docker, Redis, Kafka, RabbitMQ, and event receiver are not needed.

## 4. Dependency Management

### Existing Go module

Service module:

```text
backend/services/payment-service/go.mod
```

Current module dependency summary:

| Dependency | Status | Why relevant |
|---|---|---|
| `github.com/go-sql-driver/mysql v1.10.0` | Existing direct dependency | MySQL transactions, row locks, and retry schema reads/writes. |
| `filippo.io/edwards25519 v1.2.0` | Existing indirect dependency | Pulled by MySQL driver chain; no direct setup action. |

No new external Go module is introduced by `TASK_FILE_NAME`.

### Commands

Fresh checkout dependency install is already documented in previous files. If dependencies are not downloaded yet:

```bash
cd backend/services/payment-service
go mod download
```

Task-specific test command:

```bash
cd backend/services/payment-service
go test ./internal/domain ./internal/usecase ./internal/transport/http ./internal/config
```

Full service test command:

```bash
cd backend/services/payment-service
go test ./...
```

Beginner note:

- `go mod tidy` only run karo jab dependency list actually change hui ho.
- `TASK_FILE_NAME` ke liye `go get` required nahi hai, because MySQL driver already present hai.
- Agar `go test` module download error de, previous dependency guide ke Go proxy/cache troubleshooting section follow karo.

## 5. Database Setup

### Reused MySQL setup

MySQL install, Docker MySQL, DSN format, database credentials, and base migration flow already documented hai:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
```

Sections:

- `task1_Dependency.md` -> `5. Database Setup`
- `task2_Dependency.md` -> `5. Database Setup`
- `task2_Dependency.md` -> `9. Local Development Setup`
- `task2_Dependency.md` -> `11. Common Errors & Fixes`

### New retry migration

`TASK_FILE_NAME` requires this migration:

```text
backend/services/payment-service/migrations/005_add_payment_retry_lineage.up.sql
```

It adds retry lineage to the existing `payments` table:

| DB change | Required? | Purpose |
|---|---:|---|
| `retry_of_payment_id` column | Required | Child retry payment kis failed payment se created hai. |
| `root_payment_id` column | Required | Same retry chain ka first payment identify karta hai. |
| `attempt_no` column | Required | Attempt number track karta hai, like 1, 2, 3. |
| `retry_request_key` column | Required | Same client retry request ko replay-safe banata hai. |
| `fk_payments_retry_of` | Required | Retry parent payment valid ho. |
| `uk_payments_retry_request` | Required | Same parent plus same retry key duplicate child row na banaye. |
| `uk_payments_root_attempt` | Required | Same retry chain me duplicate attempt number na ho. |
| `idx_payments_order_status` | Required | Active/captured payment checks fast ho. |
| `idx_payments_root_status` | Required | Retry chain status checks fast ho. |

Why mandatory:

Retry flow sirf application code se safe nahi hota. MySQL constraints and transaction locks are part of duplicate-charge safety. Agar migration `005` missing hai, endpoint compile ho sakta hai but runtime DB operations fail karenge.

### Apply migration

If previous migrations are already applied, apply only the new retry migration:

```bash
cd backend/services/payment-service
mysql -h 127.0.0.1 -P 3306 -u root -proot < migrations/005_add_payment_retry_lineage.up.sql
```

Fresh database setup ke liye previous docs ka full migration flow follow karo. Practical command pattern:

```bash
cd backend/services/payment-service
for file in migrations/*.up.sql; do
  mysql -h 127.0.0.1 -P 3306 -u root -proot < "$file"
done
```

Security note:

- `-proot` sirf local example hai.
- Production me password command history me mat rakho.
- Use `PAYMENT_MYSQL_DSN` through environment/secret manager.

### Verify retry schema

Check retry columns:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db \
  -e "SHOW COLUMNS FROM payments WHERE Field IN ('retry_of_payment_id','root_payment_id','attempt_no','retry_request_key');"
```

Check retry indexes:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db \
  -e "SHOW INDEX FROM payments WHERE Key_name IN ('uk_payments_retry_request','uk_payments_root_attempt','idx_payments_order_status','idx_payments_root_status');"
```

Check retry foreign key:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db \
  -e "SELECT CONSTRAINT_NAME FROM information_schema.TABLE_CONSTRAINTS WHERE TABLE_SCHEMA='payment_db' AND TABLE_NAME='payments' AND CONSTRAINT_NAME='fk_payments_retry_of';"
```

Check service schema health after backend starts:

```bash
curl -s http://127.0.0.1:8080/internal/v1/payment-schema/health
```

Expected concept:

```json
{
  "database_name": "payment_db",
  "ready": true
}
```

### Rollback warning

Down migration:

```text
backend/services/payment-service/migrations/005_add_payment_retry_lineage.down.sql
```

It drops retry columns and indexes. Production rollback me isse retry audit data lose ho sakta hai. Beginner rule: local testing ke alawa down migration run karne se pehle DBA/senior approval lo.

## 6. Redis / Queue / External Services

### Redis / Kafka / RabbitMQ / NATS

No Redis, Kafka, RabbitMQ, or NATS dependency is introduced by `TASK_FILE_NAME`.

Refer:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Section:

```text
6. Redis / Queue / External Services
```

Task guide me Kafka/RabbitMQ platform concept mention ho sakta hai, but current repo code payment events HTTP publisher se bhejta hai. Isliye local retry setup ke liye broker install mat karo.

### Provider API

Retry API live run karne ke liye provider mode enabled hona chahiye. Provider setup already documented hai:

```text
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
```

Sections:

- `task3_Dependency.md` -> `6. Redis / Queue / External Services`
- `task3_Dependency.md` -> `7. Environment Variables`
- `task4_Dependency.md` -> `6. Redis / Queue / External Services`
- `task4_Dependency.md` -> `10. Running the Project`
- `task5_Dependency.md` -> `7. Redis / Queue / External Services`

Task-specific behavior:

| Service | Required for retry API? | Why |
|---|---:|---|
| Stripe-like provider | Required if selected | Retry creates a new intent through provider adapter. |
| Razorpay-like provider | Required if selected | Retry creates a new provider order/intent. |
| Provider webhook delivery | Required for final truth | Retry response is not final captured proof; webhook confirms final state. |
| HTTP event receiver | Required when providers enabled | Config validation requires event publishing config. |

If no provider is enabled, backend can still expose health/schema routes, but retry usecase is not configured and retry endpoint returns:

```text
PAYMENT_RETRY_UNAVAILABLE
```

## 7. Environment Variables

Common `.env` location, loading behavior, MySQL DSN, provider variables, webhook secrets, and event publisher variables are already documented:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
TaskImplementation/{SERVICE_NAME}/task6_Dependency.md
```

Sections:

- `task1_Dependency.md` -> `7. Environment Variables`
- `task3_Dependency.md` -> `7. Environment Variables`
- `task4_Dependency.md` -> `7. Environment Variables`
- `task5_Dependency.md` -> `8. Environment Variables`
- `task6_Dependency.md` -> `7. Environment Variables`

Current code reads env vars using `os.Getenv`. `.env` automatic load nahi hota. Local `.env` use kar rahe ho to manually source karo:

```bash
cd backend/services/payment-service
. ./.env
```

### New or task-specific variables

| Variable | Required? | Default | Purpose | Validation / security notes |
|---|---:|---:|---|---|
| `PAYMENT_RETRY_MAX_ATTEMPTS` | Optional but important | `3` | One order/payment chain me max attempts control karta hai. | Must be at least `2`. Keep sane value to avoid abuse/cost. |
| `PAYMENT_RETRY_COOLDOWN_SECONDS` | Optional | `5` | Very fast repeated retry actions throttle karta hai. | Must not be negative. Idempotency ka replacement nahi hai. |

Minimal task-specific `.env` overlay:

```bash
PAYMENT_RETRY_MAX_ATTEMPTS=3
PAYMENT_RETRY_COOLDOWN_SECONDS=5
```

### Existing variables needed for live retry API

These are not new, so full explanation repeat nahi kiya gaya:

| Existing variable | Why retry needs it | Where already explained |
|---|---|---|
| `PAYMENT_MYSQL_DSN` | Retry lineage rows MySQL me store hote hain. | `task1_Dependency.md`, `task2_Dependency.md` |
| `PAYMENT_INTERNAL_API_TOKEN` | Retry endpoint internal bearer auth require karta hai. | `task4_Dependency.md` |
| `PAYMENT_DEFAULT_PROVIDER` | Retry provider intent create karta hai. | `task3_Dependency.md`, `task4_Dependency.md` |
| `PAYMENT_ALLOWED_PROVIDERS` | Retry usecase only enabled provider use karega. | `task3_Dependency.md` |
| `PAYMENT_ALLOWED_CURRENCIES` | Parent payment currency provider request me validate hoti hai. | `task3_Dependency.md`, `task4_Dependency.md` |
| `STRIPE_LIKE_*` / `RAZORPAY_LIKE_*` | Provider API and webhook secrets. | `task3_Dependency.md`, `task5_Dependency.md` |
| `PAYMENT_EVENTS_ENDPOINT` | Provider mode startup validates event publisher config. | `task3_Dependency.md`, `task5_Dependency.md` |
| `PAYMENT_EVENTS_AUTH_TOKEN` | Event publisher bearer token. | `task3_Dependency.md`, `task5_Dependency.md` |

Important validation notes:

- When providers are enabled, `PAYMENT_INTERNAL_API_TOKEN` must be at least 32 characters.
- `PAYMENT_EVENTS_AUTH_TOKEN` must be at least 32 characters.
- `PAYMENT_EVENTS_ENDPOINT` must be HTTPS outside local loopback testing.
- Provider base URLs can use HTTP only for localhost/loopback testing.

### Required retry request headers

Manual API calls need these headers:

| Header | Required? | Example | Purpose |
|---|---:|---|---|
| `Authorization` | Required | `Bearer <PAYMENT_INTERNAL_API_TOKEN>` | Internal API auth. |
| `X-Actor-ID` | Required | `usr_123` | Buyer identity used for ownership check. |
| `X-Actor-Role` | Required | `buyer` | Retry endpoint accepts buyer role only. |
| `Idempotency-Key` | Required unless body has same key | `retry_pay_1001_client_abc` | Same request replay ko same retry payment return karne ke liye. |
| `X-Request-ID` | Optional | `request_123` | Logs and provider metadata correlation. |

Body can also include `idempotency_key`. If header and body both exist, they must match.

## 8. Docker Setup

No new Dockerfile, docker-compose service, Redis, Kafka, RabbitMQ, provider mock container, or retry-specific container is introduced by `TASK_FILE_NAME`.

Reuse:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Section:

```text
8. Docker Setup
```

Task-specific Docker notes:

| Component | Status | Notes |
|---|---|---|
| MySQL container | Reused | Same `payment_db` setup; just apply migration `005`. |
| Backend service container | Not committed in current repo | Run with `go run ./cmd/server` unless a Dockerfile is added later. |
| Provider mock/sandbox | External/reused | Use previous provider setup docs. |
| Event receiver container | Optional/external | Needed only if your event receiver runs in Docker. |

### Ports and networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Backend API | `8080` | Retry endpoint and health/schema APIs | Reused |
| MySQL | `3306` | `payment_db` storage | Reused |
| Provider API HTTPS | `443` | Outbound provider intent creation | Reused / external |
| Provider mock | Provider-specific | Local provider simulation if configured | Reused / optional |
| Event receiver | Usually `9080` or app-specific | Receives payment domain events | Reused / external |
| Redis/Kafka/RabbitMQ | N/A | Not used by current retry code | Not required |

Port conflict notes:

- Change backend port with `PAYMENT_HTTP_ADDR`, for example `PAYMENT_HTTP_ADDR=:8081`.
- If MySQL runs in Docker and service runs on host, DSN can use `127.0.0.1:3306` if port is published.
- If both service and MySQL run in Docker later, DSN should use Docker service name, not `127.0.0.1`.
- No new firewall rule is introduced by retry handling.

## 9. Local Development Setup

### Step 1: Read previous setup docs

Follow these first:

| File | Topic |
|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Clone, Go, MySQL, Docker, env, base run commands. |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `payment_db`, core tables, migration verification. |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | Provider registry and provider env. |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | Create-intent setup before realistic payment retry. |
| `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md` | Webhook setup because final retry outcome comes through webhook. |
| `TaskImplementation/{SERVICE_NAME}/task6_Dependency.md` | Shared provider/event publisher startup behavior. |

### Step 2: Go to service directory

```bash
cd backend/services/payment-service
```

### Step 3: Install dependencies only if needed

```bash
go mod download
```

No new `TASK_FILE_NAME` dependency install is required.

### Step 4: Start MySQL and apply migrations

Use previous MySQL/Docker setup. Then apply all current migrations, or at least migration `005` if earlier migrations are already applied.

Task-specific migration:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot < migrations/005_add_payment_retry_lineage.up.sql
```

Verify:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db \
  -e "SHOW COLUMNS FROM payments WHERE Field IN ('retry_of_payment_id','root_payment_id','attempt_no','retry_request_key');"
```

### Step 5: Configure env

For test-only:

```bash
PAYMENT_RETRY_MAX_ATTEMPTS=3
PAYMENT_RETRY_COOLDOWN_SECONDS=5
```

For live retry API, also configure existing provider and event publisher env from previous docs:

```bash
PAYMENT_INTERNAL_API_TOKEN=internal-payment-token-at-least-32-characters
PAYMENT_EVENTS_ENDPOINT=http://127.0.0.1:9080/events
PAYMENT_EVENTS_AUTH_TOKEN=event-publisher-token-at-least-32-chars
```

Provider-specific variables like `PAYMENT_DEFAULT_PROVIDER`, `PAYMENT_ALLOWED_PROVIDERS`, `STRIPE_LIKE_PUBLIC_KEY`, `STRIPE_LIKE_SECRET_KEY`, and webhook secrets are reused from earlier dependency guides.

### Step 6: Start event receiver if providers are enabled

Provider mode validates event publisher config. Use the event receiver setup already documented in:

```text
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
```

If no event receiver is running, startup may fail or payment event publish may fail depending on the flow.

### Step 7: Start backend service

```bash
go run ./cmd/server
```

Verify:

```bash
curl -s http://127.0.0.1:8080/healthz
curl -s http://127.0.0.1:8080/internal/v1/payment-schema/health
```

### Step 8: Prepare a failed payment

Best realistic path:

1. Use Task 4 create-intent setup to create an initial payment.
2. Use Task 5 webhook setup or provider sandbox to mark the initial payment `failed`.
3. Retry that failed payment with `TASK_FILE_NAME` endpoint.

Local-only demo seed, only for development DB:

```sql
USE payment_db;

INSERT INTO payments (
  payment_id, order_id, user_id, provider,
  status, currency, amount, captured_amount, refunded_amount,
  idempotency_key
) VALUES (
  'pay_retry_demo_1', 'ord_retry_demo_1', 'usr_retry_demo', 'stripe_like',
  'failed', 'INR', 249900, 0, 0,
  'payment_intent:ord_retry_demo_1:1'
)
ON DUPLICATE KEY UPDATE status = 'failed';

INSERT INTO payment_attempts (
  attempt_id, payment_id, status, failure_code, failure_message
) VALUES (
  'pat_retry_demo_1', 'pay_retry_demo_1', 'failed', 'local_demo', 'Local failed payment for retry testing'
)
ON DUPLICATE KEY UPDATE status = 'failed';
```

Beginner warning:

- Demo insert sirf local DB ke liye hai.
- Production me payment status provider/webhook truth se aana chahiye.
- Provider in seeded row must match enabled provider config.

## 10. Running the Project

### Run retry tests

```bash
cd backend/services/payment-service
go test ./internal/domain ./internal/usecase ./internal/transport/http ./internal/config
```

### Run full service tests

```bash
cd backend/services/payment-service
go test ./...
```

### Manual retry API verification

Start backend first:

```bash
cd backend/services/payment-service
go run ./cmd/server
```

Send retry request:

```bash
curl -i -X POST "http://127.0.0.1:8080/api/v1/payments/pay_retry_demo_1/retry" \
  -H "Authorization: Bearer ${PAYMENT_INTERNAL_API_TOKEN}" \
  -H "X-Actor-ID: usr_retry_demo" \
  -H "X-Actor-Role: buyer" \
  -H "X-Request-ID: request_retry_demo_001" \
  -H "Idempotency-Key: retry_pay_retry_demo_1_client_001" \
  -H "Content-Type: application/json" \
  -d '{"payment_id":"pay_retry_demo_1","idempotency_key":"retry_pay_retry_demo_1_client_001"}'
```

Expected first successful response:

| Field | Expected concept |
|---|---|
| HTTP status | `201 Created` |
| `payment_id` | New retry payment id |
| `retry.retry_of_payment_id` | Original failed payment id |
| `retry.root_payment_id` | First/root payment id |
| `retry.attempt_no` | Usually `2` for first retry |
| `retry.replayed` | `false` |

Repeat the same curl with same idempotency key:

| Field | Expected concept |
|---|---|
| HTTP status | `200 OK` |
| `payment_id` | Same retry payment id as first call |
| `retry.replayed` | `true` |
| New DB row | No extra retry payment row |

Verify DB lineage:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db \
  -e "SELECT payment_id, retry_of_payment_id, root_payment_id, attempt_no, retry_request_key, status FROM payments WHERE order_id = 'ord_retry_demo_1' ORDER BY attempt_no;"
```

Expected concept:

```text
pay_retry_demo_1 | NULL             | NULL             | 1 | NULL                                 | failed
pay_...          | pay_retry_demo_1 | pay_retry_demo_1 | 2 | retry_pay_retry_demo_1_client_001   | requires_action
```

Final payment success/failure verification is still webhook-driven. Use Task 5 webhook setup for final outcome testing.

## 11. Common Errors & Fixes

Generic Go/MySQL/Docker/provider/webhook errors are already documented in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
```

Only retry-specific errors are listed here.

| Error / symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `Unknown column 'retry_of_payment_id'` | Migration `005_add_payment_retry_lineage.up.sql` not applied. | Apply migration `005`; verify columns. | Run migrations before service start/API testing. |
| Schema health shows missing `uk_payments_retry_request` | Retry indexes missing. | Apply migration `005` or repair schema. | Include migration `005` in deployment migration order. |
| `PAYMENT_RETRY_MAX_ATTEMPTS must be at least 2` | Env set to `0` or `1`. | Set `PAYMENT_RETRY_MAX_ATTEMPTS=3` or valid value. | Keep retry policy values in `.env.example`/secret config. |
| `PAYMENT_RETRY_COOLDOWN_SECONDS cannot be negative` | Env value is negative. | Set `0` or positive seconds. | Validate env before deploy. |
| `PAYMENT_RETRY_UNAVAILABLE` | Providers are not enabled, so retry usecase is not wired. | Configure provider env and event publisher env, then restart. | For environments exposing retry, keep provider mode enabled. |
| `UNAUTHORIZED` | Missing bearer token, missing `X-Actor-ID`, or role is not `buyer`. | Send `Authorization`, `X-Actor-ID`, and `X-Actor-Role: buyer`. | Keep a retry curl/API client template. |
| `VALIDATION_ERROR` with idempotency mismatch | Header `Idempotency-Key` and body `idempotency_key` differ. | Use one key in both places or only one location. | Generate one request key per intentional retry click. |
| `VALIDATION_ERROR` with missing idempotency | No retry key sent. | Send `Idempotency-Key` header or `idempotency_key` body. | UI/API gateway must require retry idempotency key. |
| `PAYMENT_NOT_FOUND` for existing payment | Buyer user id does not own the payment, or payment id is wrong. | Use correct `X-Actor-ID` and failed payment id. | Gateway must map authenticated buyer identity correctly. |
| `PAYMENT_RETRY_NOT_ALLOWED` | Parent payment is not `failed`, or already captured/refunded state blocks retry. | Retry only failed payments. | UI should show retry button only for eligible failed payments. |
| `PAYMENT_RETRY_IN_PROGRESS` | Same order already has active `initiated`, `requires_action`, or `authorized` attempt. | Continue existing active attempt or wait for final outcome. | Do not create multiple active checkout sessions for same order. |
| `PAYMENT_ALREADY_CAPTURED` | Same order already has captured payment. | Do not retry; move to order paid/refund flow. | Webhook should update order/payment state quickly. |
| `PAYMENT_RETRY_LIMIT_REACHED` | Attempt count reached `PAYMENT_RETRY_MAX_ATTEMPTS`. | Increase policy only if product approves, or block retry. | Keep max attempts aligned with risk/fraud policy. |
| `PAYMENT_RETRY_COOLDOWN` | Retry requested before cooldown window ends. | Wait and retry later. | UI should debounce and show cooldown message. |
| `PAYMENT_CLIENT_PAYLOAD_UNAVAILABLE` | Replayed retry cannot recover browser payload from provider. | Check provider mock/sandbox retrieve support and provider connectivity. | Use provider adapters that support replay-safe payload retrieval. |
| Provider error during retry | Provider key/base URL invalid, provider down, or request rejected. | Check provider env and previous provider setup docs. | Test provider create-intent before retry flow. |

## 12. Security & Best Practices

### Security audit observations

| Observation | Risk | Recommendation |
|---|---|---|
| Retry endpoint uses internal bearer token plus buyer headers. | If gateway passes spoofed `X-Actor-ID`, buyer can target wrong payment. | Only trusted gateway/internal caller should set actor headers. |
| `PAYMENT_INTERNAL_API_TOKEN` is required in provider mode. | Weak token can expose payment operations. | Use 32+ character secret from secret manager. |
| Retry idempotency key is mandatory. | Missing key can create duplicate retry attempts. | UI/backend caller must generate one key per intentional retry action. |
| Migration `005` constraints are required. | Without DB constraints, duplicate rows can bypass app-level assumptions. | Treat migration `005` as part of production deploy, not optional docs. |
| Demo SQL can force failed status locally. | Production misuse can corrupt financial truth. | Use demo SQL only in local dev; production status comes from provider/webhook. |
| Provider metadata includes retry ids. | Sensitive metadata could leak if wrong fields are added. | Keep metadata to safe IDs only; never add card PAN/CVV/secrets. |
| MySQL default DSN uses `root:root`. | Unsafe outside local development. | Override `PAYMENT_MYSQL_DSN` in staging/prod. |
| No committed service Dockerfile/compose found. | Dev setup can differ across machines. | Reuse previous docs now; add compose profile later if team standardizes it. |

### Best practices for `TASK_FILE_NAME`

- Apply migration `005` before deploying retry-enabled code.
- Keep `PAYMENT_RETRY_MAX_ATTEMPTS` low unless product/risk team approves higher retries.
- Keep cooldown greater than zero in user-facing environments to reduce accidental spam.
- Use same retry idempotency key for network replay of the same click; use a new key only for a new intentional retry after final failure.
- Do not retry `captured`, `authorized`, `requires_action`, or refund-related states.
- Never mark retry payment successful from frontend redirect alone; webhook remains final truth.
- Log retry fields: `payment_id`, `retry_of_payment_id`, `root_payment_id`, `attempt_no`, `order_id`, `provider`, and replay flag.
- Monitor spikes in retry limit errors, cooldown errors, and provider unavailable errors.
- Keep event receiver reachable when providers are enabled, because provider-mode startup validates event publisher config.

## 13. Missing or Misconfigured Things

| Item | Current status | Setup impact | Suggested fix |
|---|---|---|---|
| Service `.env.example` | Not found in service folder | Beginners may miss retry variables and required auth headers. | Add safe `.env.example` with placeholders for retry/provider/event vars. |
| Service Dockerfile/compose | Not found in current repo scan | Full local stack remains manual. | Add compose profile later with MySQL, service, provider mock, and event sink. |
| Migration runner | Raw SQL files only | Beginners can forget migration `005` or run out of order. | Add Make target or migration tool wrapper. |
| Retry API gateway mapping | Not present in this service code | `X-Actor-ID`/role must be trustworthy. | Gateway should set actor headers from authenticated buyer context. |
| Provider mock for replay | Not committed as standalone container | Manual retry replay may fail if provider retrieve is not supported. | Provide local provider mock or document sandbox retrieve behavior. |
| Broker integration | Not used by current code | Kafka/RabbitMQ mentions can confuse setup. | Keep docs clear that current runtime uses HTTP event publisher. |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `2. Tech Stack` | Go, Go modules, MySQL driver, MySQL, Docker baseline already explained. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3. Required Software` | Git, Go, MySQL, Docker, curl install already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `5. Database Setup` | MySQL local/Docker install, DSN format, and base migration flow already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `6. Redis / Queue / External Services` | Redis/Kafka/RabbitMQ absence and HTTP event publisher baseline already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. Environment Variables` | `.env` location, manual loading, and broad variable reference already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `8. Docker Setup` | MySQL-only Docker baseline already documented. |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` | `payment_db`, base `payments` and `payment_attempts` tables, and migration verification reused. |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `11. Common Errors & Fixes` | Generic schema/migration troubleshooting already documented. |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `6. Redis / Queue / External Services` | Provider API and event receiver setup already documented. |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `7. Environment Variables` | Provider-enabled env group already documented. |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | `9. Local Development Setup` | Create-intent flow needed before realistic retry testing. |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | `10. Running the Project` | Provider-enabled service run and payment row creation pattern reused. |
| `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md` | `7. Redis / Queue / External Services` | Webhook delivery, signatures, and event receiver behavior already documented. |
| `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md` | `11. Running the Project` | Webhook verification pattern reused for final retry outcome. |
| `TaskImplementation/{SERVICE_NAME}/task6_Dependency.md` | `7. Environment Variables` | Shared provider/event publisher startup behavior and auth expectations reused. |
| `TaskImplementation/{SERVICE_NAME}/task6_Dependency.md` | `11. Common Errors & Fixes` | Existing auth/provider/event troubleshooting reused where unchanged. |

## 15. Final Checklist

- [ ] Previous dependency docs checked before using this guide.
- [ ] No duplicate Go/MySQL/Docker/provider setup copied from previous files.
- [ ] Go dependencies downloaded if this is a fresh checkout.
- [ ] MySQL is running and `PAYMENT_MYSQL_DSN` points to `payment_db`.
- [ ] All previous migrations applied in order.
- [ ] Migration `005_add_payment_retry_lineage.up.sql` applied.
- [ ] Retry columns exist on `payments`.
- [ ] Retry unique indexes and FK exist.
- [ ] `PAYMENT_RETRY_MAX_ATTEMPTS` is at least `2`.
- [ ] `PAYMENT_RETRY_COOLDOWN_SECONDS` is `0` or positive.
- [ ] Provider env configured if live retry API should work.
- [ ] `PAYMENT_INTERNAL_API_TOKEN` configured with a strong value.
- [ ] `PAYMENT_EVENTS_ENDPOINT` and `PAYMENT_EVENTS_AUTH_TOKEN` configured when providers are enabled.
- [ ] Backend starts with `go run ./cmd/server`.
- [ ] `/healthz` returns `ok`.
- [ ] `/internal/v1/payment-schema/health` reports ready.
- [ ] Retry request sends `Authorization`, `X-Actor-ID`, `X-Actor-Role: buyer`, and idempotency key.
- [ ] Same retry request replay returns same retry payment, not a new row.
- [ ] Final payment outcome verified through webhook flow.
- [ ] Logs checked for `payment.retry.created` and provider errors.
- [ ] No real provider secrets or `.env` committed.

Final beginner note:

`TASK_FILE_NAME` setup ka core formula simple hai: previous setup plus migration `005`, valid retry env, enabled provider mode, trusted buyer headers, and one idempotency key per intentional retry. Agar koi retry issue aaye, pehle schema verify karo, phir env/auth headers, phir provider/event receiver connectivity.
