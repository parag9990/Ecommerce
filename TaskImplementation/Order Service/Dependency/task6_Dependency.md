# Project Dependency & Setup Guide

## Variables Used In This Document

```text
SERVICE_NAME=Order Service
TASK_FILE_NAME=task6.md
INPUT_FILE_PATH=TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}
OUTPUT_FILE_NAME=${TASK_FILE_NAME without ".md"}_Dependency.md
OUTPUT_FILE_PATH=TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}

SERVICE_CODE_PATH=backend/services/order-service
PROTO_ROOT=proto
ORDER_PROTO_FILE=${PROTO_ROOT}/ecommerce/order/v1/order.proto
GENERATED_ORDER_PROTO_PATH=backend/shared/gen/go/ecommerce/order/v1

PREVIOUS_DEPENDENCY_FILE_TASK_1=TaskImplementation/${SERVICE_NAME}/task1_Dependency.md
PREVIOUS_DEPENDENCY_FILE_TASK_2=TaskImplementation/${SERVICE_NAME}/task2_Dependency.md
PREVIOUS_DEPENDENCY_FILE_TASK_3=TaskImplementation/${SERVICE_NAME}/task3_Dependency.md
PREVIOUS_DEPENDENCY_FILE_TASK_4=TaskImplementation/${SERVICE_NAME}/task4_Dependency.md
PREVIOUS_DEPENDENCY_FILE_TASK_5=TaskImplementation/${SERVICE_NAME}/task5_Dependency.md

TASK_6_DOMAIN=${SERVICE_CODE_PATH}/internal/domain/seller_order.go
TASK_6_LIST_USECASE=${SERVICE_CODE_PATH}/internal/usecase/list_seller_orders.go
TASK_6_UPDATE_USECASE=${SERVICE_CODE_PATH}/internal/usecase/update_seller_fulfillment.go
TASK_6_REPOSITORY=${SERVICE_CODE_PATH}/internal/repository/mysql_seller_order_repository.go
TASK_6_GRPC_HANDLER=${SERVICE_CODE_PATH}/internal/transport/grpc/order_handler.go
TASK_6_GRPC_MAPPER=${SERVICE_CODE_PATH}/internal/transport/grpc/mapper.go
TASK_6_GRPC_VALIDATION=${SERVICE_CODE_PATH}/internal/transport/grpc/validation.go
TASK_6_AUTH_CONTEXT=${SERVICE_CODE_PATH}/internal/authctx/context.go

TASK_2_BASE_MIGRATION=${SERVICE_CODE_PATH}/migrations/001_create_order_tables.up.sql
TASK_4_PAYMENT_MIGRATION=${SERVICE_CODE_PATH}/migrations/002_add_payment_coordination.up.sql
TASK_5_QUERY_INDEX_MIGRATION=${SERVICE_CODE_PATH}/migrations/003_add_grpc_query_indexes.up.sql
TASK_8_OUTBOX_MIGRATION=${SERVICE_CODE_PATH}/migrations/004_create_order_outbox_events.up.sql
```

Use `${INPUT_FILE_PATH}` as the source implementation guide. This generated setup guide is saved as `${OUTPUT_FILE_PATH}`.

## Previous Dependency Reuse

Previous dependency files inside the same `${SERVICE_NAME}` folder were checked first.

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Project tech stack, Go modules, MySQL, Kafka, Docker, ports, full `.env`, common setup errors | Same base backend runtime and infra |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_2}` | `order_db`, base tables, seller indexes, migrations | `${TASK_FILE_NAME}` reuses `orders`, `order_items`, `shipments`, and seller indexes |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_3}` | Product/Cart checkout dependencies and seller snapshot explanation | Seller view depends on checkout-time `seller_id`, product, SKU, title, price, and image snapshots |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_4}` | Payment status boundary, payment env, outbox table caveat | Seller fulfillment starts only after paid state and final delivery can create an outbox event |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_5}` | gRPC/proto setup, metadata auth, page token signing, gRPC port, no server bootstrap note | `${TASK_FILE_NAME}` extends the same gRPC boundary |

Important: Is file me Go install, MySQL install, Docker Compose basics, Kafka install, generic `.env`, and generic gRPC setup repeat nahi kiya gaya. Sirf `${TASK_FILE_NAME}` ke seller-order-view specific setup, auth metadata, DB readiness, verification, and debugging notes explain kiye gaye hain.

## 1. Project Overview

`${TASK_FILE_NAME}` ka focus seller dashboard ke liye seller-scoped order view hai.

Simple Hinglish: Ek buyer order me multiple sellers ke items ho sakte hain. Seller A ko sirf Seller A ke items, seller total, seller shipment, aur seller fulfillment status dikhna chahiye. Seller B ka item, amount, shipment, ya payment data Seller A ko expose nahi hona chahiye.

### Current Code Reality

Current repository me `${TASK_FILE_NAME}` ke seller-view implementation files present hain:

| Area | Current Files |
|---|---|
| Seller domain model and transitions | `${TASK_6_DOMAIN}` |
| Seller list usecase | `${TASK_6_LIST_USECASE}` |
| Seller fulfillment update usecase | `${TASK_6_UPDATE_USECASE}` |
| MySQL seller-scoped repository | `${TASK_6_REPOSITORY}` |
| gRPC handler branch | `${TASK_6_GRPC_HANDLER}` |
| Proto mapper | `${TASK_6_GRPC_MAPPER}` |
| Request validation | `${TASK_6_GRPC_VALIDATION}` |
| Trusted metadata auth context | `${TASK_6_AUTH_CONTEXT}` |
| Public proto contract | `${ORDER_PROTO_FILE}` |

### What `${TASK_FILE_NAME}` Adds Operationally

| Item | New For This Task? | Setup Impact |
|---|---:|---|
| `ListSellerOrders` RPC behavior | Yes | Needs trusted seller metadata and signed page token key |
| Seller branch inside `UpdateFulfillment` | Yes | Needs seller role, `x-seller-id`, and valid fulfillment sequence |
| `SellerOrderView` response projection | Yes | Gateway/API schema should use seller-safe response, not buyer full order |
| Seller-scoped MySQL queries | Yes | Reuses existing seller columns/indexes from Task 2 |
| Seller fulfillment status filters | Yes | Client must send known proto enum values only |
| Final-delivery outbox event possibility | Task-specific runtime impact | Reuses outbox table from later/current runtime migration |

No new database engine, cache, queue, container, or language runtime is introduced only by `${TASK_FILE_NAME}`.

## 2. Tech Stack

### Reused Technologies

These are already explained in previous dependency files:

| Technology | Status | Reference |
|---|---|---|
| Go `1.26.3` | Reused | `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` |
| Go modules | Reused | `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` |
| MySQL 8+ | Reused | `${PREVIOUS_DEPENDENCY_FILE_TASK_1}`, `${PREVIOUS_DEPENDENCY_FILE_TASK_2}` |
| Docker | Reused, optional local infra | `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` |
| Kafka/outbox | Reused when events enabled | `${PREVIOUS_DEPENDENCY_FILE_TASK_1}`, `${PREVIOUS_DEPENDENCY_FILE_TASK_3}`, `${PREVIOUS_DEPENDENCY_FILE_TASK_4}` |
| gRPC | Reused transport | `${PREVIOUS_DEPENDENCY_FILE_TASK_5}` |
| Protocol Buffers | Reused contract system | `${PREVIOUS_DEPENDENCY_FILE_TASK_5}` |
| Buf CLI | Reused only when proto changes | `${PREVIOUS_DEPENDENCY_FILE_TASK_5}` |

Refer:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Sections:
- Project Tech Stack Analysis
- Go Dependency System
- Database Analysis
- External Services Analysis
- Docker and DevOps Setup

${PREVIOUS_DEPENDENCY_FILE_TASK_5}
Sections:
- Tech Stack
- Dependency Management
- API Gateway / Internal Caller Metadata
```

### Task-Specific Concepts

| Concept | Required? | Why Used | Beginner Explanation |
|---|---:|---|---|
| Seller-scoped projection | Yes | Seller ko full buyer order nahi dikhana | Projection ka matlab same DB data se limited view banana. Yaha seller ko sirf apna part dikhaya jata hai. |
| Trusted `x-seller-id` metadata | Yes for seller APIs | Seller identity request body se trust nahi karni | Gateway JWT/session validate karke internal gRPC metadata me seller identity bhejta hai. |
| Seller roles | Yes | Seller route authorization ke liye | `seller`, `seller_manager`, ya `seller_order_manager` role hona chahiye. |
| HMAC page token | Yes for pagination | Cursor token tamper-proof rakhna | Page token opaque hota hai. Client ise modify nahi karega, sirf next request me pass karega. |
| MySQL row locks for fulfillment | Yes for mutation | Concurrent updates safe rakhne ke liye | Fulfillment update transaction ke andar seller rows lock karta hai so double update/race avoid ho. |
| Seller item status aggregation | Yes | Seller list me `pending`, `packed`, `partially_shipped`, etc. dikhane ke liye | Multiple items ke statuses combine karke seller-level status derive hota hai. |

## 3. Required Software

### For Reading Documentation Only

| Software | Required? | Notes |
|---|---:|---|
| Markdown viewer/editor | Yes | `${INPUT_FILE_PATH}` and `${OUTPUT_FILE_PATH}` read karne ke liye |
| Go/MySQL/Kafka/Docker | No | Sirf docs read kar rahe ho to runtime setup nahi chahiye |

### For Focused Tests

| Software | Required? | Notes |
|---|---:|---|
| Go `1.26.3` | Yes | `${SERVICE_CODE_PATH}/go.mod` me declared hai |
| Internet or populated Go module cache | Usually yes | First test run dependencies download kar sakta hai |
| MySQL | No | Current focused tests mostly fakes/sqlmock use karte hain |
| Kafka | No | Unit tests direct broker publish nahi karte |
| Cart/Product/Payment services | No | Seller list/update tests in-process fakes/sqlmock se run ho sakte hain |

### For Live Seller Order Runtime

| Software / Service | Required? | Why |
|---|---:|---|
| MySQL 8+ | Yes | Seller order list and fulfillment update real DB se data read/write karte hain |
| API Gateway or trusted internal gRPC caller | Yes | `x-internal-token`, `x-user-id`, `x-roles`, `x-seller-id` metadata forward karne ke liye |
| Kafka | Only if events enabled | Final delivered order/event publishing outbox worker ke liye |
| Docker | Optional | Local MySQL/Kafka run karne ke liye useful |
| Cart/Product/Payment services | Not for listing existing seller orders | Needed for full checkout/order creation flow from earlier tasks |

### For Proto Regeneration

| Software | Required? | Notes |
|---|---:|---|
| Buf CLI | Yes if `${ORDER_PROTO_FILE}` changes | Proto lint/generate ke liye |
| `protoc-gen-go` | Yes if generating Go stubs | `.pb.go` generate hota hai |
| `protoc-gen-go-grpc` | Yes if generating gRPC stubs | `_grpc.pb.go` generate hota hai |

Installation details are reused from `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` and `${PREVIOUS_DEPENDENCY_FILE_TASK_5}`.

## 4. Dependency Management

No new Go module dependency needs to be installed only for `${TASK_FILE_NAME}`.

Current dependency files:

```text
${SERVICE_CODE_PATH}/go.mod
${SERVICE_CODE_PATH}/go.sum
backend/shared/gen/go/go.mod
backend/shared/gen/go/go.sum
backend/go.work
```

Existing dependencies used by this task:

| Dependency | Why `${TASK_FILE_NAME}` Uses It |
|---|---|
| `google.golang.org/grpc` | Seller RPC handlers, interceptors, metadata, status codes |
| `google.golang.org/protobuf` | Seller proto messages and timestamps |
| `github.com/example/ecommerce-platform/backend/shared/gen/go` | Generated `ecommerce.order.v1` Go stubs |
| `github.com/go-sql-driver/mysql` | Real seller repository persistence |
| `github.com/DATA-DOG/go-sqlmock` | Repository tests without live MySQL |
| `github.com/segmentio/kafka-go` | Reused outbox event publishing when events are enabled |

Beginner note: `backend/go.work` can reference modules that are not present in this tree. For focused `${SERVICE_NAME}` commands, use `GOWORK=off` from `${SERVICE_CODE_PATH}`.

Practical commands:

```bash
cd backend/services/order-service
env GOWORK=off go mod download
env GOWORK=off go test ./internal/domain ./internal/usecase ./internal/repository ./internal/transport/grpc -run 'Seller|Fulfillment|List'
```

Full module verification:

```bash
cd backend/services/order-service
env GOWORK=off go test ./...
```

Run proto commands only when `${ORDER_PROTO_FILE}` changes:

```bash
cd proto
buf lint
buf generate
```

Do not manually edit generated files in `${GENERATED_ORDER_PROTO_PATH}`.

## 5. Database Setup

### Database Used: MySQL 8+

Full MySQL installation, Docker MySQL, credentials, and base schema setup are already documented in:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Sections:
- Database Analysis
- Docker and DevOps Setup

${PREVIOUS_DEPENDENCY_FILE_TASK_2}
Sections:
- Database Analysis
- Migration Order
```

### What `${TASK_FILE_NAME}` Needs From MySQL

| Table / Column / Index | Required? | Why |
|---|---:|---|
| `orders.order_id`, `orders.status`, `orders.currency`, `orders.created_at` | Yes | Parent order summary for seller projection |
| `order_items.seller_id` | Yes | Main seller ownership filter |
| `order_items.fulfillment_status` | Yes | Seller status filter and aggregation |
| `order_items.total_amount` | Yes | Seller-only subtotal |
| `shipments.seller_id` | Yes | Seller-owned shipment/tracking view |
| `idx_order_items_seller_created` | Yes, already in Task 2 | Faster seller order listing |
| `idx_shipments_seller_status` | Yes, already in Task 2 | Faster seller shipment filtering |
| `order_outbox_events` | Needed only for final delivered event path | Current code can insert event when parent becomes delivered |

### New Migration?

No new migration file is introduced only by `${TASK_FILE_NAME}`.

Practical current-runtime migration rule:

| Scenario | Migrations Needed | Why |
|---|---|---|
| Read docs only | None | No DB touched |
| Focused unit/sqlmock tests | None | No live DB |
| Seller list on real DB | `001` minimum, with current runtime usually `001`, `002`, `003` | Base tables plus current module schema alignment |
| Seller fulfillment update ending before final delivered event | `001`, `002`, `003` recommended | Keeps order status enum and current repository expectations aligned |
| Seller fulfillment update that makes parent order delivered | `001`, `002`, `003`, `004` | Delivered event can be inserted into `order_outbox_events` |
| Full current module runtime | `001`, `002`, `003`, `004` in order | Aligns checkout, payment, query, and event code |

Apply full current runtime migrations from repository root:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 < backend/services/order-service/migrations/001_create_order_tables.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/002_add_payment_coordination.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/003_add_grpc_query_indexes.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/004_create_order_outbox_events.up.sql
```

These commands are already explained in earlier docs. They are repeated here only as the task-specific migration checklist for current live runtime.

### Verify Seller-Required Schema

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW COLUMNS FROM order_items LIKE 'seller_id';"
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW COLUMNS FROM order_items LIKE 'fulfillment_status';"
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW COLUMNS FROM shipments LIKE 'seller_id';"
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW INDEX FROM order_items WHERE Key_name = 'idx_order_items_seller_created';"
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW INDEX FROM shipments WHERE Key_name = 'idx_shipments_seller_status';"
```

Verify outbox table only if testing final-delivery event behavior:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW TABLES LIKE 'order_outbox_events';"
```

### Credentials Placement

No new database credential is introduced by `${TASK_FILE_NAME}`.

Use existing variable:

```text
ORDER_MYSQL_DSN
```

Connection string format:

```text
order_user:order_password@tcp(localhost:3306)/order_db?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci
```

Where to place it:

```text
${SERVICE_CODE_PATH}/.env
```

Current config reads `os.Getenv`, so `.env` is not loaded automatically. Source it before a real run:

```bash
set -a
. backend/services/order-service/.env
set +a
```

Security note: DB password secret hai. `.env` Git me commit mat karo.

## 6. Redis / Queue / External Services

### Redis

No Redis usage is introduced by `${TASK_FILE_NAME}`.

Redis may exist in the wider platform for gateway/session/cart use, but `${SERVICE_NAME}` code does not configure Redis for seller order view.

### Kafka / Outbox

Kafka setup is reused from previous dependency files.

| Scenario | Kafka Needed? |
|---|---:|
| Read `${INPUT_FILE_PATH}` | No |
| Focused Task 6 tests | No |
| Seller list real DB read | No |
| Seller fulfillment update to packed/shipped | No direct broker needed |
| Outbox worker publishing `OrderDelivered` | Yes, if events enabled |

Important runtime note: Current seller fulfillment code can build an `OrderDelivered` outbox event when final seller delivery aggregates parent status to delivered. The table setup is reused from `${PREVIOUS_DEPENDENCY_FILE_TASK_3}` and `${PREVIOUS_DEPENDENCY_FILE_TASK_4}`. Kafka broker setup is reused from `${PREVIOUS_DEPENDENCY_FILE_TASK_1}`.

For quick local runtime without broker publishing:

```bash
export ORDER_EVENTS_ENABLED=false
```

Warning: Turning events off avoids Kafka config validation, but it does not remove the need for the `order_outbox_events` table if code path inserts an outbox row.

### API Gateway / Auth Boundary

This is mandatory for real seller APIs.

`${SERVICE_NAME}` must not trust `seller_id` from request JSON/body. The trusted Gateway or internal caller must validate JWT/session first, then forward gRPC metadata:

| Metadata Key | Required? | Example | Purpose |
|---|---:|---|---|
| `x-internal-token` | Yes | same value as `ORDER_GRPC_TRUSTED_CALLER_TOKEN` | Proves caller is trusted |
| `x-user-id` | Yes | `user_1` | Authenticated user/actor |
| `x-roles` | Yes for protected seller calls | `seller_order_manager` | Authorization |
| `x-seller-id` | Yes for seller routes | `seller_1` | Seller ownership filter |
| `x-request-id` | Optional | `req_local_123` | Logs/tracing; generated if missing |
| `x-session-id` | Optional | `session_123` | Audit/session context |

Allowed seller roles in current code:

```text
seller
seller_manager
seller_order_manager
```

### Product/User/Payment Dependencies

| Service | Required For `${TASK_FILE_NAME}` Runtime? | Why Mentioned |
|---|---:|---|
| User/Auth Service | Yes at Gateway boundary | Seller identity and roles must come from trusted auth claims |
| Product Service | Not called by seller list directly | Product/seller/price snapshots are created during checkout from earlier task flow |
| Payment Service | Not called by seller list directly | Parent order must be paid before seller fulfillment is allowed |

Refer:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_3}
Sections:
- Cart Service
- Product Service

${PREVIOUS_DEPENDENCY_FILE_TASK_4}
Sections:
- Payment Service dependency

${PREVIOUS_DEPENDENCY_FILE_TASK_5}
Section:
- API Gateway / Internal Caller Metadata
```

### RabbitMQ / NATS / MinIO / Elasticsearch

No RabbitMQ, NATS, MinIO, or Elasticsearch setup is introduced by `${TASK_FILE_NAME}`.

## 7. Environment Variables

The full `.env` example is already documented in:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Section:
Environment Variables
```

`${TASK_FILE_NAME}` introduces no brand-new environment variable. It actively depends on existing runtime variables:

| Variable | New? | Required? | Task-Specific Purpose |
|---|---:|---:|---|
| `ORDER_MYSQL_DSN` | No | Yes for real DB runtime | Seller list/update repository |
| `ORDER_GRPC_ADDR` | No | Optional, default `:9094` | Same gRPC listener used by seller RPCs |
| `ORDER_GRPC_TRUSTED_CALLER_TOKEN` | No | Yes | Validates `x-internal-token` metadata |
| `ORDER_PAGE_TOKEN_SIGNING_KEY` | No | Yes | Signs/verifies seller list pagination tokens |
| `ORDER_PAYMENT_RETURN_URL` | No | Yes for current full config validation | Existing payment config validation still runs |
| `ORDER_PAYMENT_ALLOWED_CURRENCIES` | No | Optional default `INR,USD` | Existing payment config |
| `ORDER_EVENTS_ENABLED` | No | Optional default true | Disable locally if Kafka is not running |
| `ORDER_KAFKA_BROKERS` | No | Required if events enabled | Existing outbox publisher broker list |

Task-specific minimal local exports using existing variables:

```bash
export ORDER_GRPC_ADDR=:9094
export ORDER_GRPC_TRUSTED_CALLER_TOKEN=local-trusted-caller-token-minimum-32-chars
export ORDER_PAGE_TOKEN_SIGNING_KEY=local-page-token-signing-key-minimum-32
export ORDER_MYSQL_DSN='order_user:order_password@tcp(localhost:3306)/order_db?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci'
export ORDER_PAYMENT_RETURN_URL=https://localhost.example.com/payment/return
export ORDER_EVENTS_ENABLED=false
```

If events are enabled, also set:

```bash
export ORDER_KAFKA_BROKERS=localhost:9092
```

Security notes:

- `ORDER_GRPC_TRUSTED_CALLER_TOKEN` secret hai. Browser/frontend ko kabhi expose mat karo.
- `ORDER_PAGE_TOKEN_SIGNING_KEY` stable secret rakho. Change karne se old page tokens invalid ho jayenge.
- `ORDER_MYSQL_DSN` me password hota hai. Logs, screenshots, commits me leak mat karo.
- `.env` file `${SERVICE_CODE_PATH}/.env` me local rakho, but Git me commit mat karo.

Common mistakes:

| Mistake | Result | Fix |
|---|---|---|
| `.env` create kiya but source nahi kiya | Config still missing | `set -a; . backend/services/order-service/.env; set +a` |
| Token/signing key 32 chars se short | Config validation fails | Long random string use karo |
| `ORDER_EVENTS_ENABLED=true` but Kafka missing | Startup/config validation fails | Set brokers or use `ORDER_EVENTS_ENABLED=false` locally |
| Seller API metadata me `x-seller-id` missing | Seller route unauthenticated/forbidden | Gateway se trusted seller ID forward karo |

## 8. Docker Setup

No new Dockerfile, Docker Compose service, volume, network, health check, or restart policy is introduced only by `${TASK_FILE_NAME}`.

Reuse previous Docker setup:

| Container / Service | Status | Reference |
|---|---|---|
| MySQL | Reused | `${PREVIOUS_DEPENDENCY_FILE_TASK_1}`, `${PREVIOUS_DEPENDENCY_FILE_TASK_2}` |
| Kafka | Reused only if events enabled | `${PREVIOUS_DEPENDENCY_FILE_TASK_1}`, `${PREVIOUS_DEPENDENCY_FILE_TASK_3}` |
| `${SERVICE_NAME}` gRPC server | Reused code helper; no checked-in bootstrap | `${PREVIOUS_DEPENDENCY_FILE_TASK_5}` |

### Ports & Networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| MySQL | `3306` | `order_db` persistence | Reused |
| Kafka broker | `9092` | Event publishing when enabled | Reused |
| `${SERVICE_NAME}` gRPC API | `9094` | Internal `OrderService` RPCs, including seller RPCs | Reused |
| API Gateway HTTP | `8080` | Public REST route owner for `/api/v1/seller/orders` | Reused platform convention |

Port conflict fix for local gRPC:

```bash
export ORDER_GRPC_ADDR=:19094
```

Docker networking reminder: Host machine DSN usually uses `localhost:3306`. Container-to-container DSN usually uses service DNS like `mysql:3306`. This is already explained in previous Docker docs.

## 9. Local Development Setup

### Step 1: Read Previous Dependency Documentation First

Follow these first:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
${PREVIOUS_DEPENDENCY_FILE_TASK_2}
${PREVIOUS_DEPENDENCY_FILE_TASK_3}
${PREVIOUS_DEPENDENCY_FILE_TASK_4}
${PREVIOUS_DEPENDENCY_FILE_TASK_5}
```

### Step 2: Go To Project Directory

From repository root:

```bash
cd backend/services/order-service
```

This is `${SERVICE_CODE_PATH}` for `${SERVICE_NAME}`.

### Step 3: Install Only New Dependencies If Any

No `go get` is required for `${TASK_FILE_NAME}`.

Use:

```bash
env GOWORK=off go mod download
```

### Step 4: Setup Only New Databases/Services If Any

No new database/service is added.

For real seller runtime, make sure previous MySQL migrations are applied in order. For event-enabled final delivery, verify migration `004` too.

### Step 5: Add Only New Or Changed Environment Variables

No new env variable is added. Confirm existing gRPC/auth variables:

```bash
export ORDER_GRPC_TRUSTED_CALLER_TOKEN=local-trusted-caller-token-minimum-32-chars
export ORDER_PAGE_TOKEN_SIGNING_KEY=local-page-token-signing-key-minimum-32
```

### Step 6: Run Migrations If Needed

Docs and unit tests do not need migrations.

Real DB runtime should use:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 < backend/services/order-service/migrations/001_create_order_tables.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/002_add_payment_coordination.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/003_add_grpc_query_indexes.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/004_create_order_outbox_events.up.sql
```

### Step 7: Start Backend Service

Current `${SERVICE_CODE_PATH}` has gRPC server helpers and application wiring, but no checked-in long-running `cmd/server/main.go`.

Present helper files:

```text
${SERVICE_CODE_PATH}/internal/transport/grpc/service.go
${SERVICE_CODE_PATH}/internal/transport/grpc/wiring.go
```

Practical meaning:

| Action | Status |
|---|---|
| Run focused tests | Works now |
| Create in-memory gRPC server in tests | Works now |
| Start with `go run ./cmd/server` | Not available until bootstrap is added |

When server bootstrap exists, it must wire:

```text
ORDER_GRPC_ADDR
ORDER_GRPC_TRUSTED_CALLER_TOKEN
ORDER_PAGE_TOKEN_SIGNING_KEY
ORDER_MYSQL_DSN
Order repository
ID generator
Optional outbox worker
Gateway/internal metadata caller
```

### Step 8: Verify APIs Or Functionality Related To `${TASK_FILE_NAME}`

Focused test verification:

```bash
cd backend/services/order-service
env GOWORK=off go test ./internal/domain ./internal/usecase ./internal/repository ./internal/transport/grpc -run 'Seller|Fulfillment|List'
```

Full module test verification:

```bash
cd backend/services/order-service
env GOWORK=off go test ./...
```

Optional live `grpcurl` seller list check after bootstrap exists:

```bash
grpcurl -plaintext \
  -H "x-internal-token: local-trusted-caller-token-minimum-32-chars" \
  -H "x-user-id: user_1" \
  -H "x-roles: seller_order_manager" \
  -H "x-seller-id: seller_1" \
  -d '{"pageSize":20}' \
  localhost:9094 ecommerce.order.v1.OrderService/ListSellerOrders
```

Optional live seller fulfillment shipped check:

```bash
grpcurl -plaintext \
  -H "x-internal-token: local-trusted-caller-token-minimum-32-chars" \
  -H "x-user-id: user_1" \
  -H "x-roles: seller_order_manager" \
  -H "x-seller-id: seller_1" \
  -d '{"orderId":"ord_1","targetStatus":"ORDER_STATUS_SHIPPED","carrier":"BlueDart","trackingNumber":"BD123456789"}' \
  localhost:9094 ecommerce.order.v1.OrderService/UpdateFulfillment
```

These live commands need server bootstrap, valid DB data, and `grpcurl`.

## 10. Running The Project

### Documentation-Only Flow

```bash
cd /home/parag/Ecommerce
less TaskImplementation/"${SERVICE_NAME}"/"${TASK_FILE_NAME}"
less TaskImplementation/"${SERVICE_NAME}"/"${OUTPUT_FILE_NAME}"
```

### Task-Specific Test Flow

```bash
cd /home/parag/Ecommerce/backend/services/order-service
env GOWORK=off go test ./internal/domain ./internal/usecase ./internal/repository ./internal/transport/grpc -run 'Seller|Fulfillment|List'
```

### Proto Validation Flow

Run only if proto was touched:

```bash
cd /home/parag/Ecommerce/proto
buf lint
```

### Full Runtime Flow

Full runtime currently needs a service bootstrap file that wires:

| Dependency | Why Needed |
|---|---|
| MySQL connection | Seller repository reads/updates |
| gRPC server | Internal `OrderService` listener |
| Trusted gateway/internal caller | Seller metadata and auth |
| ID generator | Shipment/history/event/request IDs |
| Page token signing key | `ListSellerOrders` pagination |
| Optional outbox worker | Event publishing |

Because no `cmd/server/main.go` exists in current `${SERVICE_CODE_PATH}`, a beginner should not expect `go run ./cmd/server` to work yet.

## 11. Common Errors & Fixes

Generic Go/MySQL/Docker/Kafka errors are already documented in previous dependency files. This section lists task-specific or newly relevant errors.

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `seller order role required` | Seller route called without allowed seller role | Send `x-roles: seller`, `seller_manager`, or `seller_order_manager` | Gateway role mapping maintain karo |
| `authentication required` on seller route | Missing/wrong `x-internal-token`, missing `x-user-id`, or missing `x-seller-id` | Send trusted metadata from Gateway | Public client se direct service call mat allow karo |
| Seller gets empty order list | `x-seller-id` does not match `order_items.seller_id`, or no paid/active items exist | Verify seller ID and test data | Seller ID auth claims and checkout snapshots align rakho |
| `order not found` during seller fulfillment | Order exists but current seller has no items in it | Use the authenticated seller who owns at least one item | Always filter by seller ownership |
| `invalid request` for `ListSellerOrders` | `page_size` < 1, > 100, bad page token, or unknown filter enum | Use page size 1-100 and server-returned token only | Treat page token as opaque |
| `invalid request` for fulfillment target | Target status is not packed, shipped, or delivered | Use `ORDER_STATUS_PACKED`, `ORDER_STATUS_SHIPPED`, or `ORDER_STATUS_DELIVERED` | Keep seller UI options aligned with proto |
| `tracking required` or `invalid request` on shipped update | `SHIPPED` target without carrier/tracking number | Send both carrier and tracking number | UI should require both before shipping |
| `operation cannot be applied in the current state` | Parent order is not paid/packed/shipped, or item transition skipped a step | Move pending -> packed -> shipped -> delivered only after payment | Show current status before action |
| `ORDER_PAGE_TOKEN_SIGNING_KEY must be at least 32 characters` | Signing key too short | Use 32+ chars and keep stable | Generate secret once per environment |
| `ORDER_GRPC_TRUSTED_CALLER_TOKEN must be at least 32 characters` | Trusted caller token too short | Use 32+ chars | Store in env/secret manager |
| `Table 'order_outbox_events' doesn't exist` during delivery | Final delivery code tried to insert delivered event but migration `004` missing | Apply `${TASK_8_OUTBOX_MIGRATION}` | Use full current runtime migrations |
| `ORDER_KAFKA_BROKERS is required when order events are enabled` | Events default true but Kafka broker is not configured | Set brokers or `ORDER_EVENTS_ENABLED=false` locally | Local `.env` me explicit events setting rakho |
| Port `9094` already in use | Another process is listening | Set `ORDER_GRPC_ADDR=:19094` | Reserve local ports per service |
| `go test` confused by workspace | `backend/go.work` references modules not present locally | Run from `${SERVICE_CODE_PATH}` with `env GOWORK=off` | Use module-isolated commands for this service |
| More than one seller shipment already exists | Current repository supports one seller shipment row per seller/order path | Clean test data or implement multi-parcel support later | Avoid manually inserting duplicate seller shipments |

## 12. Security & Best Practices

### Task-Specific Security Rules

| Area | Best Practice |
|---|---|
| Seller identity | `seller_id` request body/query se trust mat karo. Use trusted metadata `x-seller-id`. |
| Response shape | Seller APIs should return `SellerOrderView`, not full buyer `Order`. |
| SQL ownership | Every seller read/update me `seller_id = ?` predicate mandatory rakho. |
| Fulfillment update | `order_id AND seller_id` dono ke basis par rows update karo. |
| Secrets | `ORDER_GRPC_TRUSTED_CALLER_TOKEN`, `ORDER_PAGE_TOKEN_SIGNING_KEY`, and DB password logs/commits me mat daalo. |
| gRPC exposure | Production me `${SERVICE_NAME}` gRPC port internal network only rakho. Public internet par direct expose mat karo. |
| TLS/mTLS | Trusted token ke saath production TLS/mTLS add karo. Current code token interceptor use karta hai, TLS bootstrap nahi. |
| Page token | Client ko token opaque treat karna chahiye. Decode/modify karne ki zarurat nahi. |
| Logs | `seller_id` and `actor_user_id` operational logs access-controlled rakho; secrets, address, payment tokens log mat karo. |

### Beginner-Friendly Best Practices

- Seller API test karte time metadata checklist rakho: internal token, user ID, roles, seller ID, request ID.
- Seller list response me parent total amount expose mat karo; seller-only total expose karo.
- UI me seller fulfillment buttons current status ke hisab se enable karo.
- `SHIPPED` action ke liye carrier/tracking mandatory rakho.
- Page size default 20 and max 100 ke andar rakho.
- Local Kafka nahi chahiye to `ORDER_EVENTS_ENABLED=false` explicitly set karo.
- DB seed data me multi-seller order banate waqt har `order_items` row ka correct `seller_id` set karo.
- Generated `.pb.go` files manually edit mat karo. Proto edit karo, phir `buf generate`.

## 13. Missing Or Misconfigured Things

| Item | Current Status | Impact | Suggested Fix |
|---|---|---|---|
| API schema for seller endpoints | `api/master-api.json` still maps seller list/update to generic `OrderListResponse` / `Order` names | Gateway could accidentally expose full buyer order payload | Add seller-safe schemas like `SellerOrderListResponse` and `SellerOrderView` |
| Service bootstrap | No `cmd/server/main.go` in `${SERVICE_CODE_PATH}` | Long-running service cannot be started directly | Add bootstrap for config, DB, repository, gRPC listener, graceful shutdown |
| `.env.example` | Not present in `${SERVICE_CODE_PATH}` | Beginners may miss required vars | Add sanitized `.env.example` with fake values |
| Health check service | No standard gRPC health registration found | Readiness/liveness checks harder | Register gRPC health service in bootstrap |
| TLS/mTLS config | Current gRPC server validates trusted token only | Production transport security incomplete | Add TLS/mTLS config for deployed environments |
| Gateway implementation | REST `/api/v1/seller/orders` mapping is in API inventory, but concrete gateway code is not present here | Browser cannot call seller flow end-to-end from this service alone | Implement Gateway route and metadata forwarding |
| Multi-parcel seller fulfillment | Repository rejects more than one shipment row for a seller/order in current path | Sellers with split packages need later enhancement | Design explicit multi-shipment API before enabling multi-parcel |
| Config validation coupling | Current config validates payment vars even for seller-only runtime | Local seller-only startup can fail if payment env missing | Keep `.env` complete or split config validation by enabled runtime mode |
| Outbox migration ordering | Delivered event path expects `order_outbox_events` table | Final delivery can fail without migration `004` | Apply full current runtime migrations |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Project Tech Stack Analysis | Go, gRPC, MySQL, Kafka, Docker basics already explained |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Go Dependency System | Same module/dependency workflow reused |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Environment Variables | Full `.env` explanation already exists |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Docker and DevOps Setup | MySQL/Kafka container basics reused |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_2}` | Database Analysis | `order_db`, `order_items.seller_id`, `shipments.seller_id`, and seller indexes reused |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_2}` | Migration Order | Base migration `001` creates seller-visible tables |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_3}` | Product Service setup | Seller/product snapshots are produced during checkout |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_3}` | Outbox migration note | Current runtime can require `order_outbox_events` |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_4}` | Payment coordination setup | Seller fulfillment is blocked until paid/fulfillable parent status |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_4}` | Kafka/outbox troubleshooting | Outbox table and event-publishing caveats reused |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_5}` | gRPC metadata and auth | Same trusted metadata system powers seller APIs |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_5}` | Ports & Networking | Existing `9094`, `3306`, `9092` port guidance reused |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_5}` | Missing bootstrap note | Same no-`cmd/server` runtime limitation applies |

## 15. Final Checklist

- [ ] Previous dependency documentation checked first
- [ ] `${INPUT_FILE_PATH}` reviewed
- [ ] Current seller domain/usecase/repository/gRPC files identified
- [ ] No original implementation file overwritten
- [ ] No duplicate Go/MySQL/Docker/Kafka installation docs added
- [ ] No new Go module dependency required for `${TASK_FILE_NAME}`
- [ ] `ORDER_MYSQL_DSN` configured for real DB runtime
- [ ] `ORDER_GRPC_TRUSTED_CALLER_TOKEN` is 32+ chars
- [ ] `ORDER_PAGE_TOKEN_SIGNING_KEY` is 32+ chars
- [ ] `ORDER_PAYMENT_RETURN_URL` provided for current config validation
- [ ] Kafka configured or `ORDER_EVENTS_ENABLED=false` set locally
- [ ] Gateway/internal caller sends `x-internal-token`
- [ ] Gateway/internal caller sends `x-user-id`
- [ ] Gateway/internal caller sends allowed seller role
- [ ] Gateway/internal caller sends trusted `x-seller-id`
- [ ] MySQL migrations `001`, `002`, `003` applied for current real runtime
- [ ] Migration `004` applied if testing final delivered event path
- [ ] Seller indexes from Task 2 verified
- [ ] Seller list response uses `SellerOrderView`
- [ ] Seller API never exposes other sellers' items or totals
- [ ] Fulfillment sequence tested: packed -> shipped -> delivered
- [ ] Shipped status requires carrier and tracking number
- [ ] Focused tests run with `env GOWORK=off`
- [ ] `buf lint` run if proto was touched
- [ ] Missing server bootstrap understood before trying `go run ./cmd/server`
- [ ] Secrets kept out of Git and logs
- [ ] No duplicate setup documentation added
