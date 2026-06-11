# Project Dependency & Setup Guide

## Variables Used In This Document

```text
SERVICE_NAME=Order Service
TASK_FILE_NAME=task4.md
INPUT_FILE_PATH=TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}
OUTPUT_FILE_NAME=task4_Dependency.md
OUTPUT_FILE_PATH=TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}
SERVICE_CODE_PATH=backend/services/order-service

PREVIOUS_DEPENDENCY_FILE_TASK_1=TaskImplementation/${SERVICE_NAME}/task1_Dependency.md
PREVIOUS_DEPENDENCY_FILE_TASK_2=TaskImplementation/${SERVICE_NAME}/task2_Dependency.md
PREVIOUS_DEPENDENCY_FILE_TASK_3=TaskImplementation/${SERVICE_NAME}/task3_Dependency.md

TASK_4_MIGRATION_UP=${SERVICE_CODE_PATH}/migrations/002_add_payment_coordination.up.sql
TASK_4_MIGRATION_DOWN=${SERVICE_CODE_PATH}/migrations/002_add_payment_coordination.down.sql
TASK_4_INITIATE_USECASE=${SERVICE_CODE_PATH}/internal/usecase/initiate_order_payment.go
TASK_4_APPLY_RESULT_USECASE=${SERVICE_CODE_PATH}/internal/usecase/apply_payment_result.go
TASK_4_CONTRACTS=${SERVICE_CODE_PATH}/internal/usecase/contracts.go
TASK_4_PAYMENT_DOMAIN=${SERVICE_CODE_PATH}/internal/domain/payment.go
TASK_4_CONFIG=${SERVICE_CODE_PATH}/internal/config/config.go
TASK_4_TEST=${SERVICE_CODE_PATH}/internal/usecase/payment_coordination_test.go
```

Use `${INPUT_FILE_PATH}` as the source task guide. This file is saved as `${OUTPUT_FILE_PATH}`.

## Previous Dependency Reuse

Previous dependency files inside the same `${SERVICE_NAME}` folder were checked first.

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Project Tech Stack Analysis | Go, gRPC, Protobuf, MySQL, Kafka, Docker, ports, and base env setup already explained |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Go Dependency System | Same Go module and dependency commands are used |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Environment Variables | Full `.env` example already lists payment variables |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | External Services Analysis -> Kafka / Payment provider | Kafka and Payment Service boundary already introduced |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Docker and DevOps Setup | Local MySQL/Kafka Docker setup already explained |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_2}` | Database Analysis and Migration Order | Base `order_db`, `orders`, and `order_status_history` setup reused |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_3}` | Cart/Product/Inventory setup | Task 4 depends on Task 3 inventory reservation handoff |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_3}` | Common Errors and Fixes | Cart/Product and migration mismatch errors already explained |

Important: Is file me same Go install, MySQL install, Docker Compose, Kafka setup, and Cart/Product setup repeat nahi kiya gaya. Sirf `${TASK_FILE_NAME}` ke new payment-specific dependencies and setup notes explain kiye gaye hain.

## 1. Project Overview

`${TASK_FILE_NAME}` ka focus payment coordination hai:

```text
created order -> Payment Service intent -> pending_payment -> verified result -> paid/payment_failed
```

Simple Hinglish: Task 3 order create karke inventory temporarily reserve karta hai. Task 4 us order ko Payment Service ke through payable banata hai. Payment intent create hone ke baad order `pending_payment` hota hai. Final status browser callback se nahi, Payment Service ke verified result se update hota hai.

### What Task 4 Adds

| Area | New for Task 4? | Notes |
|---|---:|---|
| Payment Service dependency | Yes | `CreatePaymentIntent` contract required for live payment initiation |
| Payment result application | Yes | Verified `captured` or `failed` result ko order lifecycle me apply karna |
| `payment_failed` lifecycle status | Yes | Provider/payment failure ka durable order status |
| Durable inventory reservation columns | Yes | Payment result ke baad same reservation commit/release karna hota hai |
| Payment return URL validation | Yes, env already documented | `ORDER_PAYMENT_RETURN_URL` absolute HTTPS URL hona chahiye |
| Allowed payment currency whitelist | Yes, env already documented | `ORDER_PAYMENT_ALLOWED_CURRENCIES` validated ISO-style 3-letter codes |
| Payment inventory action timeout | Yes, env already documented | Commit/release inventory action ke liye timeout |
| `OrderPaid` outbox event on captured result | Current code reality | Real DB captured-flow needs outbox table migration too |

### Current Code Reality

Current `${SERVICE_CODE_PATH}` includes Task 4 usecases and DB methods:

```text
${TASK_4_INITIATE_USECASE}
${TASK_4_APPLY_RESULT_USECASE}
${TASK_4_CONTRACTS}
${TASK_4_PAYMENT_DOMAIN}
${TASK_4_MIGRATION_UP}
${TASK_4_TEST}
```

But current repo does not contain:

```text
backend/services/payment-service
proto/ecommerce/payment/v1/payment.proto
cmd/server/main.go for ${SERVICE_CODE_PATH}
public ApplyPaymentResult RPC or event consumer in ${SERVICE_CODE_PATH}
```

Practical meaning:

- Unit tests can run without MySQL, Payment Service, Product Service, or Kafka.
- Payment initiation is currently exposed through `CreateOrderResponse.payment_action`, not a separate public payment RPC.
- Applying payment result is implemented as a usecase, but a future Payment Service adapter, gRPC endpoint, or event consumer must call it.
- Real captured-result persistence needs migrations aligned with payment coordination and outbox events.

## 2. Tech Stack Analysis

### Reused Technologies

These are already explained in earlier dependency files:

| Technology | Reuse Note |
|---|---|
| Go | Same backend module; no new language runtime introduced |
| Go modules | Same `go.mod`, `go.sum`, and `backend/go.work` behavior |
| gRPC | Same internal service communication style |
| Protocol Buffers | Same generated order proto surface |
| MySQL 8+ | Same `${SERVICE_NAME}` database owner |
| Kafka | Same outbox/event broker setup when events are enabled |
| Docker | Same optional local dependency container setup |

Refer:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Sections:
- Project Tech Stack Analysis
- Go Dependency System
- External Services Analysis
- Docker and DevOps Setup

${PREVIOUS_DEPENDENCY_FILE_TASK_2}
Sections:
- Database Analysis
- Migration Order
```

### Task-4-Specific Technologies And Concepts

| Technology / Concept | Required? | Why Used | Beginner Explanation |
|---|---:|---|---|
| Payment Service client port | Yes for live payment | Provider-specific payment logic ko `${SERVICE_NAME}` se alag rakhna | Payment Service paise ka owner hai. `${SERVICE_NAME}` sirf order status coordinate karta hai. |
| Payment intent | Yes for live checkout | Provider payment process start karne ke liye | Intent ka matlab payment start hua; payment successful hona abhi confirm nahi hai. |
| Provider-verified payment result | Yes | Browser success page trust nahi karna | Final truth provider webhook se Payment Service validate karega, phir `${SERVICE_NAME}` update hoga. |
| Inventory commit/release | Yes | Paid order me reservation commit, failed payment me release | Stock hold ko final karna ya free karna Product Service ka kaam hai. |
| Idempotency key | Yes | Duplicate payment intent/charge avoid karne ke liye | Same network retry se double payment intent nahi banana chahiye. |
| MySQL `CHECK` constraint | Yes for migration `002` | Reservation id and expiry pair enforce karna | Dono values saath honi chahiye, warna payment result finalize nahi ho paayega. |
| Outbox event row | Required for captured real DB flow | `OrderPaid` event same transition ke saath persist hota hai | Event pehle DB me safe store hota hai, Kafka worker later publish karta hai. |

## 3. Required Software

For only reading `${INPUT_FILE_PATH}`:

| Software | Required? | Notes |
|---|---:|---|
| Markdown viewer/editor | Yes | Task guide read karne ke liye |
| Go | No | Sirf docs read kar rahe ho to required nahi |
| MySQL | No | Sirf docs read kar rahe ho to required nahi |
| Payment provider account | No | `${SERVICE_NAME}` provider SDK use nahi karta |

For Task 4 unit tests:

| Software | Required? | Notes |
|---|---:|---|
| Go `1.26.3` | Yes | Module declares this version |
| Internet/module cache | Usually yes | First run may download dependencies |
| MySQL | No | Usecase tests fake repositories use karte hain |
| Payment Service | No | Tests fake `PaymentClient` use karte hain |
| Product Service | No | Tests fake inventory client use karte hain |
| Kafka | No | Unit tests do not publish to Kafka |

For live runtime/integration:

| Software / Service | Required? | Notes |
|---|---:|---|
| MySQL 8+ | Yes | Orders, status history, reservation fields, and outbox rows |
| Payment Service | Yes | `CreatePaymentIntent` and verified result source |
| Product Service inventory API | Yes | Reservation commit/release |
| Cart Service | Yes for full checkout | Task 3 dependency reused by `CreateOrder` |
| Kafka | If events enabled | `OrderPaid` outbox publishing later needs broker |
| Docker | Optional | Useful for MySQL/Kafka local dependencies |

Installation for Go, MySQL, Docker, and Kafka is already covered in:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
```

## 4. Dependency Management

No new Go module dependency is introduced only by `${TASK_FILE_NAME}`.

Current module dependency files:

```text
${SERVICE_CODE_PATH}/go.mod
${SERVICE_CODE_PATH}/go.sum
backend/go.work
backend/shared/gen/go/go.mod
```

The existing dependencies already cover Task 4:

| Dependency | Why Task 4 Uses It |
|---|---|
| `google.golang.org/grpc` | Existing order gRPC transport and future concrete Payment/Product clients |
| `google.golang.org/protobuf` | Generated order proto messages like `PaymentAction` |
| `github.com/go-sql-driver/mysql` | Payment status transitions persist in MySQL |
| `github.com/DATA-DOG/go-sqlmock` | Repository tests without a real DB |
| `github.com/segmentio/kafka-go` | Outbox worker publishes `OrderPaid` when events are enabled |

Refer:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Section:
Go Dependency System
```

Task-specific verification command:

```bash
cd backend/services/order-service
env GOWORK=off go test ./internal/usecase -run Payment
```

Why `GOWORK=off`? Current `backend/go.work` can reference services not fully present in this tree. Module-isolated test run beginner ke liye less confusing hai.

## 5. Database Setup

### Database Used: MySQL 8+

Full MySQL installation, Docker setup, and base schema are already covered in:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Sections:
- Database Analysis
- Docker and DevOps Setup

${PREVIOUS_DEPENDENCY_FILE_TASK_2}
Sections:
- Database Analysis
- Task 2 Schema Requirements
- Migration Order
```

### Why Task 4 Uses MySQL

Payment coordination needs durable state because payment result may arrive later.

MySQL stores:

- `orders.payment_id`
- `orders.status`
- `orders.inventory_reservation_id`
- `orders.inventory_reserved_until`
- `order_status_history` audit entries
- `order_outbox_events` for `OrderPaid` in current captured-flow code

Simple Hinglish: Payment async hota hai. User browser close kar de tab bhi order ko pata hona chahiye ki kaunsi payment and kaunsi inventory reservation pending thi.

### Required Or Optional

| Use Case | MySQL Required? |
|---|---:|
| Reading `${INPUT_FILE_PATH}` | No |
| Running `${TASK_4_TEST}` unit tests | No |
| Running payment repository tests with `sqlmock` | No |
| Running real payment initiation persistence | Yes |
| Running real captured payment result with outbox event | Yes |
| Running full service runtime | Yes |

### Task 4 Migration Requirement

Task 4's schema delta is:

```text
${TASK_4_MIGRATION_UP}
```

It adds:

| Change | Why Needed |
|---|---|
| `payment_failed` in `orders.status` enum | Payment failure needs explicit lifecycle state |
| `payment_failed` in `order_status_history` enums | Audit trail must support the same status |
| `orders.inventory_reservation_id` | Later commit/release must target the exact reservation |
| `orders.inventory_reserved_until` | Expired reservations should not be paid |
| `idx_orders_inventory_reservation` | Operational lookup/debug by reservation id |
| `chk_orders_inventory_reservation_pair` | Reservation id and expiry must be present together |

### Migration Order

For Task 4 minimum DB setup:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 < backend/services/order-service/migrations/001_create_order_tables.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/002_add_payment_coordination.up.sql
```

For current full `${SERVICE_CODE_PATH}` runtime:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 < backend/services/order-service/migrations/001_create_order_tables.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/002_add_payment_coordination.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/003_add_grpc_query_indexes.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/004_create_order_outbox_events.up.sql
```

Why `004` may be needed for Task 4 captured result:

```text
${TASK_4_APPLY_RESULT_USECASE}
```

builds an `OrderPaid` outbox event when result is `captured`. The real repository inserts that event into `order_outbox_events`, so migration `004` must exist for captured-result integration tests.

### Verify Task 4 Schema

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW COLUMNS FROM orders LIKE 'inventory_reservation_id';"
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW COLUMNS FROM orders LIKE 'inventory_reserved_until';"
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW COLUMNS FROM orders LIKE 'payment_id';"
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW INDEX FROM orders WHERE Key_name = 'idx_orders_inventory_reservation';"
```

Check enum contains `payment_failed`:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW COLUMNS FROM orders LIKE 'status';"
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW COLUMNS FROM order_status_history LIKE 'to_status';"
```

Check outbox table if testing captured result with real repository:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW TABLES LIKE 'order_outbox_events';"
```

### Rollback Note

Rollback file:

```text
${TASK_4_MIGRATION_DOWN}
```

Warning: The down migration maps `payment_failed` orders/history to `cancelled` before removing the enum value. Use it only on local/dev data where this downgrade is acceptable.

## 6. Payment Service And External Services

### Payment Service

| Question | Answer |
|---|---|
| What is it? | Payment provider abstraction service for intent, webhook, refund, and reconciliation |
| Why Task 4 needs it | `${SERVICE_NAME}` calls `CreatePaymentIntent` and later trusts verified results from Payment Service |
| Mandatory for unit tests? | No, fake `PaymentClient` is used |
| Mandatory for live checkout? | Yes |
| Direct provider SDK in `${SERVICE_NAME}`? | No |
| Direct payment DB access from `${SERVICE_NAME}`? | No |

Expected client contract from `${TASK_4_CONTRACTS}`:

```text
CreatePaymentIntent(order_id, user_id, amount, currency, idempotency_key, return_url)
  -> payment_id, status, provider, provider_intent_ref, client_action_token, expires_at
```

Allowed intent response statuses from `${TASK_4_PAYMENT_DOMAIN}`:

```text
initiated
requires_action
```

Allowed verified result values:

```text
captured
failed
```

Current repo gap:

```text
backend/services/payment-service
proto/ecommerce/payment/v1/payment.proto
```

are not checked in. So a real Payment Service adapter cannot be configured from this module yet.

Refer:

```text
docs/04-microservice-design.md
Section:
Payment Service

docs/07-payment-system.md
Sections:
- Payment Architecture
- Core Rules
- Idempotency
- Security
```

### Payment Provider

No Stripe/Razorpay-like SDK, API key, webhook secret, or card-data handling belongs to `${SERVICE_NAME}`.

Provider setup belongs to Payment Service. `${SERVICE_NAME}` only needs:

- safe `payment_id`
- safe `client_action_token`
- verified payment result callback/event from Payment Service
- amount/currency match validation

Refer:

```text
docs/11-devops-external-services.md
Section:
Payment Gateway Setup
```

### Product Service / Inventory

Task 4 reuses Task 3 inventory dependency.

Expected inventory methods from `${TASK_4_CONTRACTS}`:

```text
CommitInventory(reservation_id, idempotency_key)
ReleaseInventory(reservation_id, idempotency_key, reason)
```

When payment result is:

| Payment Result | Inventory Action |
|---|---|
| `captured` | Commit reservation with key `inventory_commit:{order_id}` |
| `failed` | Release reservation with key `inventory_release:{order_id}:payment_failed` |
| Intent creation terminal failure | Release reservation with key `inventory_release:{order_id}:payment_intent_failed` |
| Unknown intent outcome | Do not release yet; needs resolution |

Refer:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_3}
Sections:
- Redis / Queue / External Services -> Product Service
- Environment Variables -> Task-3-Relevant Existing Variables
```

### Kafka / Queue

`${TASK_FILE_NAME}` itself does not require Kafka for unit tests. Current code can create an `OrderPaid` outbox row when a captured result is applied.

| Item | Required? | Notes |
|---|---:|---|
| `order_outbox_events` table | Yes for real captured-result repository flow | Migration `004` creates it |
| Kafka broker | Only if publishing worker/events enabled | Reuse Kafka setup from `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` |
| RabbitMQ/NATS | No | Not used by current `${SERVICE_CODE_PATH}` code |

For quick local config without Kafka publishing:

```bash
export ORDER_EVENTS_ENABLED=false
```

### Redis

`${SERVICE_NAME}` does not directly use Redis for Task 4.

Redis may be used by API Gateway, Cart Service, or other services, but do not add Redis credentials to `${SERVICE_NAME}` unless code imports/configures Redis.

## 7. Environment Variables

The full environment system is already documented in:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Section:
Environment Variables
```

Task 4 actively depends on these variables:

| Variable | New? | Required? | Task 4 Purpose |
|---|---:|---:|---|
| `ORDER_MYSQL_DSN` | No | Yes for real runtime | Persist payment status, reservation fields, and history |
| `ORDER_PAYMENT_RETURN_URL` | No, but now actively required | Yes | URL passed to Payment Service intent creation |
| `ORDER_PAYMENT_ALLOWED_CURRENCIES` | No | Optional with default `INR,USD` | Currency whitelist before payment intent creation |
| `ORDER_PAYMENT_INVENTORY_ACTION_TIMEOUT` | No | Optional with default `3s` | Timeout for payment-related inventory commit/release |
| `ORDER_EVENTS_ENABLED` | No | Optional | Disable Kafka publishing locally if broker is not running |
| `ORDER_KAFKA_BROKERS` | No | Required if events enabled | Kafka broker list for outbox worker |
| `ORDER_GRPC_TRUSTED_CALLER_TOKEN` | No | Yes for gRPC server | Internal caller auth token |
| `ORDER_PAGE_TOKEN_SIGNING_KEY` | No | Yes for listing runtime | Required by current app server wiring |

Task-specific minimal exports for config/usecase integration:

```bash
export ORDER_PAYMENT_RETURN_URL=https://localhost.example/checkout/result
export ORDER_PAYMENT_ALLOWED_CURRENCIES=INR,USD
export ORDER_PAYMENT_INVENTORY_ACTION_TIMEOUT=3s
```

Validation rules from `${TASK_4_CONFIG}`:

| Variable | Validation Rule | Beginner Fix |
|---|---|---|
| `ORDER_PAYMENT_RETURN_URL` | Must be absolute HTTPS URL | Use `https://...`, not `http://...` |
| `ORDER_PAYMENT_ALLOWED_CURRENCIES` | At least one 3-letter uppercase currency code | Use `INR,USD`; avoid `rupees`, `inr`, or blank |
| `ORDER_PAYMENT_INVENTORY_ACTION_TIMEOUT` | Positive Go duration | Use `3s`, `1500ms`, or `1m`; not `3 seconds` |

Current config uses `os.Getenv`, so `.env` is not loaded automatically. Source it first:

```bash
set -a
. backend/services/order-service/.env
set +a
```

### Missing Env Variables For Live Payment Integration

Current `${SERVICE_CODE_PATH}` has no checked-in config variables like:

```text
ORDER_PAYMENT_GRPC_ADDR
ORDER_PAYMENT_RESULT_CONSUMER_GROUP
ORDER_DOWNSTREAM_GRPC_TIMEOUT
```

This is a live-integration gap. Jab concrete Payment Service adapter add hoga, target address, timeout, TLS, retry, and auth settings env/config me add karne chahiye.

Suggested future config names:

```env
ORDER_PAYMENT_GRPC_ADDR=localhost:9095
ORDER_PAYMENT_GRPC_TIMEOUT=1500ms
ORDER_PAYMENT_GRPC_TRUSTED_CALLER_TOKEN=change-this-payment-token-minimum-32-chars
ORDER_PAYMENT_RESULT_SOURCE=grpc
```

These are suggestions only; current code does not read them yet.

## 8. Docker Setup

`${TASK_FILE_NAME}` does not introduce a new checked-in Dockerfile, compose file, container, volume, or network for `${SERVICE_NAME}`.

Reuse:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Section:
Docker and DevOps Setup
```

### What Changes For Task 4 Live Integration

For unit tests, no Docker services are needed.

For live checkout/payment integration, local compose eventually needs:

| Container / Service | Purpose | Status |
|---|---|---|
| MySQL | `${SERVICE_NAME}` DB | Reused from previous guides |
| Kafka | Event publishing if enabled | Reused from previous guide |
| Cart Service | Full checkout input | Task 3 dependency; missing locally |
| Product Service | Inventory reservation commit/release | Task 3/4 dependency; adapter missing |
| Payment Service | Intent creation and verified results | Missing from current repo |
| Payment provider sandbox | Actual provider intent/webhook testing | Payment Service-owned setup |

Do not put provider secret keys inside `${SERVICE_NAME}`. Provider keys belong to Payment Service.

## 9. Local Development Setup

### Step 1: Read Previous Dependency Documentation

Follow these first:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
${PREVIOUS_DEPENDENCY_FILE_TASK_2}
${PREVIOUS_DEPENDENCY_FILE_TASK_3}
```

They explain Go setup, MySQL, Docker, Kafka, base migrations, Cart/Product inventory setup, and common setup issues.

### Step 2: Read Task 4 Guide

From repository root:

```bash
sed -n '1,220p' "TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}"
```

### Step 3: Go To Project Directory

```bash
cd backend/services/order-service
```

### Step 4: Install Dependencies

No new package is required for Task 4. Use existing Go module commands:

```bash
go mod download
env GOWORK=off go test ./internal/usecase -run Payment
```

### Step 5: Setup Databases/Services

For unit tests:

```text
No MySQL, Payment Service, Product Service, Cart Service, Redis, or Kafka needed.
```

For real runtime:

1. Start MySQL using previous guide.
2. Apply migrations in order.
3. Start Cart Service when available.
4. Start Product Service when available.
5. Start Payment Service when available.
6. Start Kafka only if events are enabled.

### Step 6: Add Environment Variables

Use existing `.env` from `${PREVIOUS_DEPENDENCY_FILE_TASK_1}`.

Task 4-specific check:

```bash
export ORDER_PAYMENT_RETURN_URL=https://localhost.example/checkout/result
export ORDER_PAYMENT_ALLOWED_CURRENCIES=INR,USD
export ORDER_PAYMENT_INVENTORY_ACTION_TIMEOUT=3s
```

### Step 7: Run Migrations If Real DB Is Used

From repository root:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 < backend/services/order-service/migrations/001_create_order_tables.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/002_add_payment_coordination.up.sql
```

For full current runtime:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/003_add_grpc_query_indexes.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/004_create_order_outbox_events.up.sql
```

### Step 8: Verify Task 4 Functionality

Unit-level verification:

```bash
cd backend/services/order-service
env GOWORK=off go test ./internal/domain -run Payment
env GOWORK=off go test ./internal/usecase -run Payment
```

Repository-level verification:

```bash
env GOWORK=off go test ./internal/repository -run 'Payment|CreateOrderAndAttachClaim'
```

Full module verification:

```bash
env GOWORK=off go test ./...
```

## 10. Running The Project

### Current State

`${SERVICE_CODE_PATH}` currently has no checked-in:

```text
cmd/server/main.go
```

So there is no direct production-like start command for `${SERVICE_NAME}` yet.

Refer:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Section:
Project Run Instructions -> Step 10: Start backend service
```

### What Can Be Run Today

```bash
cd backend/services/order-service
env GOWORK=off go test ./internal/usecase -run Payment
env GOWORK=off go test ./...
env GOWORK=off go build ./...
```

### Current Client-Facing Flow

Current order proto has:

```text
CreateOrderResponse.payment_action
PaymentAction.payment_id
PaymentAction.client_action_token
PaymentAction.expires_at
```

So the current flow is:

```text
CreateOrder -> creates/replays order -> initiates payment -> returns payment_action
```

There is no separate `InitiateOrderPayment` RPC in current `proto/ecommerce/order/v1/order.proto`.

### Future Live Runtime Flow

After server entrypoint and concrete downstream adapters are added:

```bash
cd backend/services/order-service
set -a
. .env
set +a
go run ./cmd/server
```

Expected future checks:

```bash
grpcurl -plaintext localhost:9094 list
grpcurl -plaintext localhost:9094 ecommerce.order.v1.OrderService/CreateOrder
```

Note: `grpcurl` verification will require proper internal caller metadata because current gRPC server validates a trusted caller token.

## 11. Ports And Networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| MySQL | 3306 | `${SERVICE_NAME}` persistence | Reused |
| `${SERVICE_NAME}` gRPC API | 9094 | Internal Order RPC API | Reused; server entrypoint missing |
| Kafka | 9092 | Event broker if events enabled | Reused |
| Cart Service gRPC | TBD | Full checkout cart dependency | Task 3 dependency; not configured |
| Product Service gRPC | 9090 planned in docs | Product snapshot/inventory dependency | Adapter missing in `${SERVICE_CODE_PATH}` |
| Payment Service gRPC | TBD | `CreatePaymentIntent` dependency | Missing in current repo |
| Payment provider webhook HTTP | Provider-specific | Provider -> Payment Service | Not owned by `${SERVICE_NAME}` |
| Redis | 6379 | Not direct for `${SERVICE_NAME}` | No direct config |

Port conflict checks are already explained in:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Section:
Ports and Networking
```

Task 4 networking rules:

```text
${SERVICE_NAME} -> Payment Service API, not payment_db.
${SERVICE_NAME} -> Product Service inventory API, not product DB.
Payment provider webhook -> Payment Service, not ${SERVICE_NAME}.
```

## 12. Common Errors & Fixes

Generic Go, MySQL, Docker, Kafka, and env errors are already covered in:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Section:
Common Errors and Fixes

${PREVIOUS_DEPENDENCY_FILE_TASK_2}
Section:
Common Errors and Fixes

${PREVIOUS_DEPENDENCY_FILE_TASK_3}
Section:
Common Errors & Fixes
```

Task-4-specific errors:

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `ORDER_PAYMENT_RETURN_URL must be an absolute HTTPS URL` | Missing URL or `http://` URL | Set `https://...` value | Keep local `.env` explicit |
| `ORDER_PAYMENT_ALLOWED_CURRENCIES includes invalid currency` | Bad currency like `inr`, `IN`, or `rupees` | Use uppercase 3-letter codes like `INR,USD` | Validate env in startup tests |
| `ORDER_PAYMENT_INVENTORY_ACTION_TIMEOUT must be greater than zero` | Missing/invalid negative duration | Use `3s` or `1500ms` | Use Go duration format |
| `unsupported payment currency` | Order currency not in allowed list | Add valid currency or fix order/Product data | Align Product and Payment currency config |
| `inventory reservation expired` | `inventory_reserved_until` is past | Recreate checkout/reservation | Keep reservation TTL realistic |
| `invalid payment intent response` | Payment Service returned empty payment id, expired intent, bad status, or missing action token | Fix Payment Service adapter/contract mapping | Add contract tests |
| `payment intent outcome needs resolution` | Timeout/transport error after intent request outcome unknown | Retry/query Payment Service with same idempotency key | Adapter should wrap unknown outcomes correctly |
| `payment intent creation failed` | Payment Service/provider terminally rejected intent | Order becomes `payment_failed`; inventory release should run | Surface retry flow through Payment Service later |
| `payment does not match order` | Result `payment_id` does not equal order's saved payment id | Reject result and investigate | Payment Service must send bound order/payment pair |
| `payment amount or currency does not match order` | Result amount/currency differs from order snapshot | Reject result; do not finalize inventory | Payment Service must use server order amount |
| `invalid payment order transition` | Result arrived for wrong current status | Check order status and duplicate/late events | Idempotent result handling and monitoring |
| `inventory finalization failed` | Product Service commit/release failed after DB status update | Retry/reconcile reservation action | Add durable inventory action outbox later |
| `Unknown column 'inventory_reservation_id'` | Migration `002` not applied | Apply `${TASK_4_MIGRATION_UP}` | Apply migrations in order |
| `Table 'order_outbox_events' doesn't exist` | Captured result wrote `OrderPaid` event but migration `004` missing | Apply `004_create_order_outbox_events.up.sql` | Use full runtime migration order |
| Live checkout cannot reach Payment Service | No concrete Payment adapter/address config exists | Implement adapter and env config | Add integration tests and health checks |
| Payment result never updates order | `ApplyPaymentResultUsecase` has no transport/event consumer wired | Add gRPC/event consumer from Payment Service | Define payment result contract |

## 13. Security & Best Practices

Reused security guidance:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Sections:
- Security and Configuration Audit
- Best Practices

${PREVIOUS_DEPENDENCY_FILE_TASK_2}
Section:
Security and Configuration Notes

${PREVIOUS_DEPENDENCY_FILE_TASK_3}
Section:
Security & Best Practices
```

Task-4-specific best practices:

- Frontend amount trust mat karo. Payment request amount `orders.total_amount` se hi jana chahiye.
- Browser success callback ko final status mat mano. Payment Service ka provider-verified result hi source of truth hai.
- Provider keys, webhook secrets, card data, CVV, and raw provider payloads `${SERVICE_NAME}` me store mat karo.
- Payment idempotency key deterministic rakho: `payment_intent:{order_id}:1`.
- Unknown intent outcome par order fail/release mat karo; same idempotency key se resolve/retry karo.
- Payment result me `payment_id`, `order_id`, amount, currency, and provider event id validate karo.
- `provider_event_id` audit metadata me store karo, but secrets/card data logs me mat daalo.
- Inventory commit/release idempotency keys deterministic rakho.
- Payment DB direct access mat karo. Payment data Payment Service API se hi aayega.
- Production me Payment Service webhook signature verification, secret manager, TLS/mTLS, and reconciliation mandatory honge.
- `payment_failed` rollback data-losing semantic change kar sakta hai; shared DB me down migration carefully use karo.

## 14. Missing Or Misconfigured Things

| Item | Status | Impact | Suggested Fix |
|---|---|---|---|
| `cmd/server/main.go` | Missing | Service ko real process ki tarah start nahi kar sakte | Add bootstrap wiring config, DB, repositories, usecases, gRPC server |
| Payment Service implementation | Missing from current repo | Live `CreatePaymentIntent` unavailable | Implement `backend/services/payment-service` or connect to external running service |
| Payment proto | Missing | Typed concrete payment gRPC client cannot be generated yet | Add `proto/ecommerce/payment/v1/payment.proto` |
| Payment concrete adapter | Missing in `${SERVICE_CODE_PATH}` | `PaymentClient` cannot call real service | Add gRPC adapter implementing `PaymentClient` |
| Payment result transport | Missing | Verified results cannot reach `${SERVICE_NAME}` automatically | Add gRPC endpoint, Kafka consumer, or internal handler wiring |
| Payment address env vars | Missing | No config for Payment Service target | Add validated payment gRPC address, timeout, TLS/auth vars |
| Product inventory concrete adapter | Missing | Live commit/release unavailable | Add Product Service adapter implementing `InventoryClient` |
| `.env.example` | Missing for `${SERVICE_CODE_PATH}` | Beginners may create incomplete env | Add sanitized example file |
| Docker Compose for full checkout/payment | Missing | Hard to run MySQL/Kafka/Cart/Product/Payment together | Add compose once services exist |
| Durable inventory finalization retry | Not implemented | DB status can update but inventory commit/release may fail | Add outbox/retry worker for inventory finalization |
| Payment retry flow | Not implemented in `${SERVICE_NAME}` | `payment_failed` retry path needs future UX/API rules | Implement with Payment Service retry rules |

## 15. References To Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Use It For |
|---|---|---|
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Project Tech Stack Analysis | Go/gRPC/MySQL/Kafka/Docker basics |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Go Dependency System | `go mod`, `go test`, module/workspace notes |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Environment Variables | Full `.env` sample and env loading behavior |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | External Services Analysis -> Kafka | Kafka setup if events are enabled |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | External Services Analysis -> Payment provider | Provider setup boundary |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Docker and DevOps Setup | MySQL/Kafka Docker setup |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Ports and Networking | Common port conflicts |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Common Errors and Fixes | Generic setup troubleshooting |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_2}` | Database Analysis | Base schema and MySQL rules |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_2}` | Migration Order | Applying `001` before later migrations |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_3}` | Redis / Queue / External Services -> Product Service | Inventory reservation dependency |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_3}` | Environment Variables | Inventory TTL/release timeout setup |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_3}` | Missing or Misconfigured Things | Cart/Product adapter and live checkout gaps |

## Final Checklist

- [ ] `${INPUT_FILE_PATH}` reviewed
- [ ] Previous dependency files checked before writing duplicate setup
- [ ] No duplicate Go/MySQL/Docker/Kafka installation guide added
- [ ] Task 4 payment-specific dependencies documented
- [ ] Payment Service boundary documented
- [ ] Payment provider secrets kept outside `${SERVICE_NAME}`
- [ ] Product inventory commit/release dependency documented
- [ ] Migration `002` purpose documented
- [ ] Full runtime migration order documented
- [ ] `payment_failed` status setup documented
- [ ] `inventory_reservation_id` and `inventory_reserved_until` setup documented
- [ ] Payment env validation rules documented
- [ ] Missing Payment Service/proto/adapter called out
- [ ] Missing payment result transport called out
- [ ] Unit test command documented
- [ ] Common Task 4 setup/runtime errors documented

## Quick Beginner Summary

`${TASK_FILE_NAME}` ke liye naya Go package install nahi karna. Main new setup is:

```text
Payment Service contract + MySQL migration 002 + payment env vars + Product inventory finalization
```

Fast local check:

```bash
cd backend/services/order-service
env GOWORK=off go test ./internal/usecase -run Payment
```

Real DB check:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 < backend/services/order-service/migrations/001_create_order_tables.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/002_add_payment_coordination.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW COLUMNS FROM orders LIKE 'inventory_reservation_id';"
```

Live checkout/payment abhi fully runnable nahi hai because Payment Service, payment proto, concrete Payment adapter, Product inventory adapter, and server entrypoint are not checked in yet. Unit tests and package-level verification are the safe current path.
