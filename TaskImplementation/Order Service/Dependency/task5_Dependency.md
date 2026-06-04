# Project Dependency & Setup Guide

## Variables Used In This Document

```text
SERVICE_NAME=Order Service
TASK_FILE_NAME=task5.md
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

TASK_5_GRPC_TRANSPORT_PATH=${SERVICE_CODE_PATH}/internal/transport/grpc
TASK_5_AUTH_CONTEXT=${SERVICE_CODE_PATH}/internal/authctx/context.go
TASK_5_CONFIG=${SERVICE_CODE_PATH}/internal/config/config.go
TASK_5_MIGRATION_UP=${SERVICE_CODE_PATH}/migrations/003_add_grpc_query_indexes.up.sql
TASK_5_MIGRATION_DOWN=${SERVICE_CODE_PATH}/migrations/003_add_grpc_query_indexes.down.sql
```

Use `${INPUT_FILE_PATH}` as the source task guide. This file is saved as `${OUTPUT_FILE_PATH}`.

## Previous Dependency Reuse

Previous dependency files inside the same `${SERVICE_NAME}` folder were checked first.

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Project tech stack, Go modules, full `.env`, MySQL install, Kafka, Docker, ports, common setup errors | Same base backend module and runtime infrastructure |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_2}` | MySQL schema setup and base migration flow | `${TASK_FILE_NAME}` reuses the same `order_db` and adds only one query-supporting index |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_3}` | Cart/Product/Inventory dependencies and checkout runtime flow | `CreateOrder` RPC delegates to the existing checkout usecase |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_4}` | Payment coordination, payment env, and Payment Service boundary | `CreateOrder` response can include payment action data from the existing payment flow |

Important: Is file me Go install, MySQL install, Kafka install, Docker container basics, full `.env`, and generic migration troubleshooting repeat nahi kiya gaya. Sirf `${TASK_FILE_NAME}` ke gRPC/proto/runtime-specific setup points explain kiye gaye hain.

## 1. Project Overview

`${TASK_FILE_NAME}` ka focus typed internal gRPC API hai.

Task-specific RPCs:

| RPC | Purpose | Actor |
|---|---|---|
| `CreateOrder` | Cart-based checkout workflow ko gRPC se trigger karna | Buyer |
| `GetOrder` | Single order snapshot read karna | Buyer/admin/order manager |
| `ListOrders` | Buyer order history cursor pagination ke saath read karna | Buyer |
| `UpdateFulfillment` | Paid order ko packed/shipped/delivered state me move karna | Seller/admin/logistics/order manager |

Simple Hinglish: `${TASK_FILE_NAME}` me handler business logic dobara nahi likhta. Handler ka kaam request validate karna, trusted metadata read karna, proto ko usecase command me map karna, aur result ko proto response me convert karna hai.

### Current Code Reality

Current repository me `${TASK_FILE_NAME}` ke real implementation files already present hain:

| Area | Files |
|---|---|
| Proto contract | `${ORDER_PROTO_FILE}` |
| Generated Go stubs | `${GENERATED_ORDER_PROTO_PATH}/order.pb.go`, `${GENERATED_ORDER_PROTO_PATH}/order_grpc.pb.go` |
| gRPC transport | `${TASK_5_GRPC_TRANSPORT_PATH}/server.go`, `service.go`, `order_handler.go`, `validation.go`, `mapper.go`, `error_mapping.go`, `wiring.go` |
| Trusted auth metadata | `${TASK_5_AUTH_CONTEXT}` |
| Runtime config | `${TASK_5_CONFIG}` |
| Task-specific DB index | `${TASK_5_MIGRATION_UP}` |

Current proto also contains `ListSellerOrders` and `CancelOrder`. Ye forward-scope RPCs current code me present hain, but this guide ka main setup focus required Task 5 RPCs par hi hai.

## 2. Tech Stack

### Reused Technologies

These are already explained in earlier dependency files:

| Technology | Reuse Note |
|---|---|
| Go `1.26.3` | Same backend language and module setup |
| Go modules | Same `${SERVICE_CODE_PATH}/go.mod`, `go.sum`, and local `replace` for shared generated code |
| MySQL 8+ | Same `order_db`; only one Task 5 index migration is new |
| Kafka | Same outbox/event setup; not newly introduced by gRPC transport |
| Docker | Same optional local infra approach for MySQL/Kafka |
| Cart Service | Same checkout dependency from Task 3 |
| Product Service | Same inventory/product dependency from Task 3 |
| Payment Service | Same payment coordination dependency from Task 4 |

Refer:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Sections:
- Project Tech Stack Analysis
- Go Dependency System
- Database Analysis
- External Services Analysis
- Docker and DevOps Setup

${PREVIOUS_DEPENDENCY_FILE_TASK_3}
Sections:
- Tech Stack
- Required Software

${PREVIOUS_DEPENDENCY_FILE_TASK_4}
Sections:
- Tech Stack Analysis
- Required Software
```

### Task-5-Specific Technologies And Concepts

| Technology / Concept | Required? | Why Used | Beginner Explanation |
|---|---:|---|---|
| gRPC server transport | Yes | Internal typed API expose karne ke liye | gRPC ek fast service-to-service communication system hai. Request/response typed proto messages se hota hai. |
| Protocol Buffers contract | Yes | RPC schema define karne ke liye | Proto ek contract file hai jo batata hai request aur response me kaunse fields honge. |
| Generated Go stubs | Yes | Go handler interface and client types ke liye | `.pb.go` files proto se auto-generate hote hain. Inhe manually edit nahi karna. |
| Buf CLI | Required only when proto changes | Proto lint/generate workflow ke liye | Buf proto files ko check karta hai aur generated code banata hai. |
| gRPC metadata | Yes | Trusted caller, user, role, seller, request context pass karne ke liye | Metadata HTTP headers jaisa hota hai, but gRPC calls ke saath attach hota hai. |
| HMAC page tokens | Yes for `ListOrders` | Cursor token tampering prevent karne ke liye | Page token me signature hota hai; galat key se token invalid ho jayega. |
| Unary interceptors | Yes | Auth, recovery, logging common layer me run karne ke liye | Interceptor request ke handler tak pahunchne se pehle common checks karta hai. |

## 3. Required Software

For only reading `${INPUT_FILE_PATH}`:

| Software | Required? | Notes |
|---|---:|---|
| Markdown viewer/editor | Yes | Task guide and this setup guide read karne ke liye |
| Go | No | Sirf docs read kar rahe ho to required nahi |
| MySQL/Kafka/Docker | No | Sirf docs read kar rahe ho to required nahi |

For `${TASK_FILE_NAME}` transport/unit tests:

| Software | Required? | Notes |
|---|---:|---|
| Go `1.26.3` | Yes | Module declares this version |
| Internet or populated module cache | Usually yes | First run dependencies download karega |
| MySQL | No | Handler/usecase tests fakes and mocks use karte hain |
| Kafka | No | Focused gRPC tests broker publish nahi karte |
| Cart/Product/Payment services | No | Focused tests fake ports use karte hain |

For proto regeneration:

| Software | Required? | Notes |
|---|---:|---|
| Buf CLI | Yes if editing proto | `buf lint` and `buf generate` ke liye |
| `protoc-gen-go` | Yes if generating Go stubs | `buf.gen.yaml` local plugin use karta hai |
| `protoc-gen-go-grpc` | Yes if generating gRPC stubs | `order_grpc.pb.go` generate karne ke liye |

For live runtime:

| Software / Service | Required? | Notes |
|---|---:|---|
| MySQL 8+ | Yes | Read/list/update fulfillment repository data ke liye |
| Cart Service | Yes for live `CreateOrder` | Existing checkout workflow |
| Product Service | Yes for live `CreateOrder` | Price/inventory validation and reservation |
| Payment Service | Yes for payment action handoff | Existing Task 4 boundary |
| Kafka | Required only when events enabled | `ORDER_EVENTS_ENABLED` default true hai |
| Docker | Optional | Local MySQL/Kafka run karne ke liye useful |

Installation for Go, MySQL, Docker, and Kafka is already covered in:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
```

## 4. Dependency Management

No new Go module dependency needs to be installed only for `${TASK_FILE_NAME}`. Current `${SERVICE_CODE_PATH}/go.mod` already contains:

| Dependency | Why Task 5 Uses It |
|---|---|
| `github.com/example/ecommerce-platform/backend/shared/gen/go` | Generated order proto module |
| `google.golang.org/grpc` | gRPC server, client, metadata, status codes, interceptors |
| `google.golang.org/protobuf` | Generated proto messages and timestamps |
| `github.com/go-sql-driver/mysql` | Real repository reads/list/update persistence |
| `github.com/DATA-DOG/go-sqlmock` | Repository tests without real MySQL |
| `github.com/segmentio/kafka-go` | Reused event/outbox publishing when enabled |

### Important Dependency Files

```text
${SERVICE_CODE_PATH}/go.mod
${SERVICE_CODE_PATH}/go.sum
backend/shared/gen/go/go.mod
backend/shared/gen/go/go.sum
backend/go.work
```

Beginner note: `backend/go.work` currently references `./services/auth-service`, but that module is not present in this tree. For focused `${SERVICE_NAME}` commands, use `GOWORK=off` from `${SERVICE_CODE_PATH}`.

### Practical Commands

From repository root:

```bash
cd backend/services/order-service
env GOWORK=off go mod download
env GOWORK=off go test ./internal/transport/grpc ./internal/usecase -run 'Order|Fulfillment|List'
```

Why this command? Ye Task 5 ke gRPC handler, auth metadata, mapping, status code, and list/fulfillment usecase tests ko run karta hai without live MySQL/Kafka.

### Proto Commands

Run only when `${ORDER_PROTO_FILE}` changes:

```bash
cd proto
buf lint
buf generate
```

Expected generated output:

```text
${GENERATED_ORDER_PROTO_PATH}/order.pb.go
${GENERATED_ORDER_PROTO_PATH}/order_grpc.pb.go
```

Rule: Generated `.pb.go` files manually edit mat karo. Proto file edit karo, phir `buf generate` rerun karo.

## 5. Database Setup

### Database Used: MySQL 8+

Full MySQL explanation, local installation, Docker setup, and base schema are already documented in:

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

### What Task 5 Adds

`${TASK_5_MIGRATION_UP}` adds this index:

```text
idx_orders_user_created_order
```

Simple Hinglish: `ListOrders` buyer ke orders newest-first cursor pagination ke saath read karta hai. Is query ko fast banane ke liye `user_id`, `created_at`, aur `order_id` par index chahiye.

### Required Or Optional

| Use Case | MySQL Required? | Migration Required? |
|---|---:|---:|
| Reading `${INPUT_FILE_PATH}` | No | No |
| Running focused gRPC handler tests | No | No |
| Regenerating proto stubs | No | No |
| Running real `GetOrder`/`ListOrders` repository flow | Yes | `001`, `002`, and `003` |
| Running full current module with create/payment/events | Yes | `001`, `002`, `003`, and `004` |

Why `002` is also needed for current Task 5 runtime? Current repository read mapper selects payment/reservation columns added by Task 4, so real DB schema should be aligned with current code, not only the original Task 5 guide.

### Migration Order

For Task 5 real read/list runtime:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 < backend/services/order-service/migrations/001_create_order_tables.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/002_add_payment_coordination.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/003_add_grpc_query_indexes.up.sql
```

For full current runtime including outbox events:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/004_create_order_outbox_events.up.sql
```

### Verify Task 5 Index

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW INDEX FROM orders WHERE Key_name = 'idx_orders_user_created_order';"
```

If no row returns, Task 5 index migration is not applied.

### Credentials Placement

No new database credentials are introduced by `${TASK_FILE_NAME}`.

Use the existing variable:

```text
ORDER_MYSQL_DSN
```

Refer:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Section:
Environment Variables
```

Security note: DB password secret hai. `.env` commit mat karo.

## 6. Redis / Queue / External Services

### Redis

No Redis usage is introduced by `${TASK_FILE_NAME}`.

### RabbitMQ / NATS / MinIO / Elasticsearch

No RabbitMQ, NATS, MinIO, or Elasticsearch setup is introduced by `${TASK_FILE_NAME}`.

### Kafka

Kafka is reused from earlier runtime/event setup. `${TASK_FILE_NAME}` ka gRPC handler directly Kafka publish nahi karta.

| Scenario | Kafka Needed? |
|---|---:|
| Focused Task 5 handler tests | No |
| Proto lint/generate | No |
| Local service startup with `ORDER_EVENTS_ENABLED=false` | No |
| Full current runtime with events enabled | Yes |

Refer:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Sections:
- External Services Analysis
- Docker and DevOps Setup
```

### API Gateway / Internal Caller Metadata

This is the most important external boundary for `${TASK_FILE_NAME}`.

Current gRPC auth interceptor requires trusted metadata:

| Metadata Key | Required? | Example | Purpose |
|---|---:|---|---|
| `x-internal-token` | Yes | same value as `ORDER_GRPC_TRUSTED_CALLER_TOKEN` | Proves caller is trusted Gateway/internal service |
| `x-user-id` | Yes | `user_1` | Authenticated actor identity |
| `x-roles` | Required for protected actions | `buyer`, `admin`, `seller_order_manager` | Authorization decisions |
| `x-seller-id` | Required for seller routes | `seller_1` | Seller-scoped list/fulfillment |
| `x-request-id` | Optional | `req_123` | Logs/tracing; generated if missing |
| `x-session-id` | Optional | `session_123` | Checkout audit/security context |

Beginner note: Client payload se `user_id` trust nahi hota. Gateway JWT/session validate karke metadata forward karega. `${SERVICE_NAME}` metadata ko token ke saath trust karta hai.

## 7. Environment Variables

The full `.env` example is already documented in:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Section:
Environment Variables
```

`${TASK_FILE_NAME}` actively depends on these existing variables:

| Variable | New? | Required? | Task 5 Purpose |
|---|---:|---:|---|
| `ORDER_GRPC_ADDR` | No | Optional, default `:9094` | gRPC server listen address |
| `ORDER_GRPC_TRUSTED_CALLER_TOKEN` | No | Yes | Validates `x-internal-token` metadata |
| `ORDER_PAGE_TOKEN_SIGNING_KEY` | No | Yes | Signs/verifies `ListOrders` cursor tokens |
| `ORDER_MYSQL_DSN` | No | Yes for real runtime | Repository reads/list/update |
| `ORDER_PAYMENT_RETURN_URL` | No | Yes for current full config validation | Existing payment coordination config |
| `ORDER_EVENTS_ENABLED` | No | Optional | Set false locally if Kafka is not needed |
| `ORDER_KAFKA_BROKERS` | No | Required if events enabled | Existing event/outbox Kafka broker list |

Task-specific minimal gRPC exports:

```bash
export ORDER_GRPC_ADDR=:9094
export ORDER_GRPC_TRUSTED_CALLER_TOKEN=local-trusted-caller-token-minimum-32-chars
export ORDER_PAGE_TOKEN_SIGNING_KEY=local-page-token-signing-key-minimum-32
```

Validation rules from `${TASK_5_CONFIG}`:

| Variable | Rule | Beginner Fix |
|---|---|---|
| `ORDER_GRPC_ADDR` | Non-empty listen address | Use `:9094` locally, or `127.0.0.1:9094` |
| `ORDER_GRPC_TRUSTED_CALLER_TOKEN` | Minimum 32 characters | Generate a long random secret |
| `ORDER_PAGE_TOKEN_SIGNING_KEY` | Minimum 32 characters/bytes | Use a stable long random secret |

Current config uses `os.Getenv`, so `.env` is not loaded automatically. Source it before running a real service:

```bash
set -a
. backend/services/order-service/.env
set +a
```

## 8. Docker Setup

No new Dockerfile, Docker Compose service, volume, or network is introduced by `${TASK_FILE_NAME}`.

Reuse earlier setup for:

| Container / Service | Status | Reference |
|---|---|---|
| MySQL | Reused | `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` |
| Kafka | Reused if events enabled | `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` |

### Ports & Networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| MySQL | `3306` | `order_db` persistence | Reused |
| Kafka broker | `9092` | Outbox/event publishing when enabled | Reused |
| `${SERVICE_NAME}` gRPC API | `9094` | Internal `OrderService` RPCs from `ORDER_GRPC_ADDR=:9094` | Reused but now actively used by Task 5 |
| Cart Service gRPC | Project-specific | Checkout dependency for live `CreateOrder` | Reused from Task 3 |
| Product Service gRPC | Project-specific | Product/inventory dependency for live `CreateOrder` | Reused from Task 3 |
| Payment Service gRPC | Project-specific | Payment handoff dependency for live `CreateOrder` | Reused from Task 4 |

Port conflict fix:

```bash
export ORDER_GRPC_ADDR=:19094
```

Docker networking note: Container-to-container DSNs should use service names like `mysql:3306`, but host-machine commands usually use `localhost:3306`. This is already explained in earlier Docker setup; do not mix them blindly.

## 9. Local Development Setup

### Step 1: Read Previous Dependency Documentation First

Follow these before starting `${TASK_FILE_NAME}` setup:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
${PREVIOUS_DEPENDENCY_FILE_TASK_2}
${PREVIOUS_DEPENDENCY_FILE_TASK_3}
${PREVIOUS_DEPENDENCY_FILE_TASK_4}
```

### Step 2: Go To Project Directory

From repository root:

```bash
cd backend/services/order-service
```

This is `${SERVICE_CODE_PATH}` for `${SERVICE_NAME}`.

### Step 3: Install Only New Dependencies If Needed

For normal tests:

```bash
env GOWORK=off go mod download
```

No `go get` is required for `${TASK_FILE_NAME}` because gRPC/protobuf dependencies are already present.

For proto updates only, ensure these tools exist:

```bash
buf --version
protoc-gen-go --version
protoc-gen-go-grpc --version
```

### Step 4: Setup Only New Databases/Services If Needed

No new database or service is added.

If running real repository flows, apply Task 5 index migration after previous migrations:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/003_add_grpc_query_indexes.up.sql
```

### Step 5: Add Only New Or Changed Environment Variables

No brand-new env variable was introduced only by `${TASK_FILE_NAME}`. For Task 5 gRPC startup, make sure existing gRPC variables are configured:

```bash
export ORDER_GRPC_ADDR=:9094
export ORDER_GRPC_TRUSTED_CALLER_TOKEN=local-trusted-caller-token-minimum-32-chars
export ORDER_PAGE_TOKEN_SIGNING_KEY=local-page-token-signing-key-minimum-32
```

### Step 6: Run Migrations If Needed

Docs or unit tests do not need migrations.

Real read/list/update runtime should use:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 < backend/services/order-service/migrations/001_create_order_tables.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/002_add_payment_coordination.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/003_add_grpc_query_indexes.up.sql
```

### Step 7: Start Backend Service

Current `${SERVICE_CODE_PATH}` has gRPC server helpers but no checked-in `cmd/server/main.go`.

Present helper:

```text
${TASK_5_GRPC_TRANSPORT_PATH}/service.go
```

Practical meaning:

| Action | Status |
|---|---|
| Run focused tests | Works now |
| Create in-memory gRPC server in tests | Works now |
| Start long-running service with `go run ./cmd/server` | Not available until server bootstrap is added |

When bootstrap exists, expected startup shape will use:

```text
ORDER_GRPC_ADDR
ORDER_GRPC_TRUSTED_CALLER_TOKEN
ORDER_PAGE_TOKEN_SIGNING_KEY
ORDER_MYSQL_DSN
```

### Step 8: Verify APIs Or Functionality Related To `${TASK_FILE_NAME}`

Focused verification:

```bash
cd backend/services/order-service
env GOWORK=off go test ./internal/transport/grpc ./internal/usecase -run 'Order|Fulfillment|List'
```

Proto verification:

```bash
cd proto
buf lint
```

Optional live gRPC check after a real server bootstrap exists:

```bash
grpcurl -plaintext \
  -H "x-internal-token: local-trusted-caller-token-minimum-32-chars" \
  -H "x-user-id: user_1" \
  -H "x-roles: buyer" \
  -d '{"orderId":"ord_1"}' \
  localhost:9094 ecommerce.order.v1.OrderService/GetOrder
```

This command needs live server, valid DB data, and `grpcurl` installed.

## 10. Running The Project

### Documentation-Only Flow

```bash
cd /home/parag/Ecommerce
less TaskImplementation/"${SERVICE_NAME}"/"${TASK_FILE_NAME}"
less TaskImplementation/"${SERVICE_NAME}"/"${OUTPUT_FILE_NAME}"
```

### Task 5 Test Flow

```bash
cd /home/parag/Ecommerce/backend/services/order-service
env GOWORK=off go test ./internal/transport/grpc ./internal/usecase -run 'Order|Fulfillment|List'
```

### Proto Validation Flow

```bash
cd /home/parag/Ecommerce/proto
buf lint
```

### Full Runtime Flow

Full runtime currently needs a service bootstrap file that wires:

| Dependency | Why Needed |
|---|---|
| MySQL connection | Repository implementation |
| Checkout usecase and Cart/Product adapters | Live `CreateOrder` |
| Payment initiation adapter | Payment action in `CreateOrderResponse` |
| ID generator | Request IDs, history IDs, order IDs |
| gRPC server | Internal `OrderService` listener |
| Optional outbox worker | Kafka event publishing |

Because no `cmd/server/main.go` exists in current `${SERVICE_CODE_PATH}`, a fresher developer should not expect `go run ./cmd/server` to work yet.

## 11. Common Errors & Fixes

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `authentication required` from gRPC | Missing/wrong `x-internal-token` or missing `x-user-id` | Send metadata matching `ORDER_GRPC_TRUSTED_CALLER_TOKEN` and actor fields | Gateway should always inject trusted metadata |
| `buyer role required` | `CreateOrder`/`ListOrders` called without `buyer` role | Send `x-roles: buyer` for buyer calls | Keep role mapping consistent in Gateway |
| `seller order role required` | Seller route called without seller role | Use `seller`, `seller_manager`, or `seller_order_manager` | Do not reuse buyer metadata for seller APIs |
| `Unauthenticated` on seller route | `x-seller-id` missing | Add `x-seller-id` for seller-scoped calls | Gateway should derive seller ID from trusted auth/session |
| `ORDER_GRPC_TRUSTED_CALLER_TOKEN must be at least 32 characters` | Token too short | Use a long random value | Store in secret manager/env, not code |
| `ORDER_PAGE_TOKEN_SIGNING_KEY must be at least 32 characters` | Page token signing key too short | Use 32+ chars and keep stable | Rotate carefully because old tokens become invalid |
| `invalid request` for list API | `page_size` less than 1, greater than 100, or bad token/filter | Use `page_size` 1-100 and pass server-returned token only | Treat page token as opaque |
| `invalid request` for create API | Empty cart, missing address, missing/oversized idempotency key | Send valid `cart_id`, shipping address, and key up to current domain limit | Validate on Gateway before gRPC call |
| `tracking required` / `invalid request` for fulfillment | `SHIPPED` without carrier/tracking or unsupported target status | Send carrier/tracking for shipped; use packed/shipped/delivered only | Keep UI status options aligned with proto/domain |
| `operation cannot be applied in the current state` | Domain rejected lifecycle transition | Check current order status before update | Do not bypass domain state machine |
| Port `9094` already in use | Another process is listening | Change `ORDER_GRPC_ADDR`, e.g. `:19094` | Reserve service ports in local docs |
| `buf: command not found` | Buf CLI not installed or not in `PATH` | Install Buf and reopen shell | Add tool install step to onboarding |
| `protoc-gen-go: program not found` during `buf generate` | Go proto plugin missing | Install `protoc-gen-go` and `protoc-gen-go-grpc` | Keep developer machine tools documented |
| Go workspace mentions missing `auth-service` | `backend/go.work` references a module not present in this tree | Use `env GOWORK=off` from `${SERVICE_CODE_PATH}` | Keep workspace file aligned or run module-isolated commands |
| MySQL `Unknown column inventory_reservation_id` | Task 4 migration missing for current code | Apply `002_add_payment_coordination.up.sql` | Apply migrations in order |
| Slow buyer order listing | Task 5 index missing | Apply `${TASK_5_MIGRATION_UP}` | Verify `idx_orders_user_created_order` exists |
| `ORDER_KAFKA_BROKERS is required when order events are enabled` | Events default true but Kafka not configured | Set brokers or `ORDER_EVENTS_ENABLED=false` locally | Use local `.env` with explicit events setting |

Generic MySQL/Docker/Kafka errors are already covered in:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
${PREVIOUS_DEPENDENCY_FILE_TASK_2}
${PREVIOUS_DEPENDENCY_FILE_TASK_3}
${PREVIOUS_DEPENDENCY_FILE_TASK_4}
```

## 12. Security & Best Practices

### Task-5-Specific Security

| Area | Best Practice |
|---|---|
| Trusted caller token | `ORDER_GRPC_TRUSTED_CALLER_TOKEN` 32+ random chars rakho. Browser/public client ko kabhi expose mat karo. |
| Actor identity | `user_id`, roles, and seller identity request body se mat lo. Trusted Gateway metadata se lo. |
| gRPC network | Production me gRPC port internal network/VPC/Kubernetes service tak limited rakho. Public internet par expose mat karo. |
| TLS/mTLS | Token auth ke saath production me TLS/mTLS add karo. Current code token/interceptor use karta hai, TLS config nahi. |
| Page token key | `ORDER_PAGE_TOKEN_SIGNING_KEY` secret and stable rakho. Rotation planned window me karo. |
| Logs | `request_id`, method, actor ID log karo; phone, address, payment action token, and raw secrets log mat karo. |
| Generated code | `.pb.go` files manually edit mat karo. Proto contract update + `buf generate` use karo. |
| Error messages | SQL/internal errors client ko leak mat karo. Current `toStatusError` safe generic messages map karta hai. |
| Deadlines | Gateway se finite timeout set karo; handler downstream `ctx` pass kare. |

### Beginner-Friendly Best Practices

- gRPC request test karte time metadata checklist rakho: token, user ID, roles, request ID.
- `ListOrders` page token ko decode/modify karne ki koshish mat karo; token opaque hai.
- `page_size` ko 20 default and 100 max ke aas-paas rakho.
- Fulfillment UI me sirf valid target statuses show karo: packed, shipped, delivered.
- `SHIPPED` ke liye carrier and tracking number mandatory rakho.
- Proto breaking change se pehle client teams ko inform karo.
- `buf lint` local run karo before commit.
- Local Kafka nahi chahiye to explicitly `ORDER_EVENTS_ENABLED=false` set karo.

## 13. Missing Or Misconfigured Things

| Item | Current Status | Impact | Suggested Fix |
|---|---|---|---|
| Service bootstrap | No `cmd/server/main.go` in `${SERVICE_CODE_PATH}` | Long-running service cannot be started directly | Add bootstrap wiring config, DB, adapters, gRPC listener, shutdown |
| `.env.example` | Not present in `${SERVICE_CODE_PATH}` | Beginners may miss required vars | Add sanitized `.env.example` with fake values |
| Docker Compose | No committed compose file found in repo root | Local infra setup depends on docs/manual commands | Add compose for MySQL/Kafka if team wants one-command setup |
| TLS/mTLS config | Current gRPC server uses trusted token interceptor only | Production transport security incomplete | Add TLS/mTLS config for non-local deployments |
| Health check service | No gRPC health service registration found | Readiness checks harder | Register standard gRPC health service in bootstrap |
| Cart/Product/Payment adapter env | No concrete target address config found in current `${SERVICE_CODE_PATH}` | Live `CreateOrder` wiring unclear | Add env names for downstream gRPC targets, timeouts, and auth tokens |
| Workspace file | `backend/go.work` references missing `auth-service` module | Workspace-level `go test` can confuse beginners | Add missing module or update workspace; meanwhile use `GOWORK=off` |
| Current proto includes future RPCs | `ListSellerOrders` and `CancelOrder` are present beyond core Task 5 methods | Scope can confuse readers | Keep task docs clear about current-code extras versus Task 5 core |

## 14. References To Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Project Tech Stack Analysis | Go, gRPC, Protobuf, MySQL, Kafka, Docker basics already explained |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Go Dependency System | Same `go.mod`, `go.sum`, shared generated module, and test commands style |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Environment Variables | Full `.env` example already includes gRPC, DB, payment, event variables |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | External Services Analysis | MySQL/Kafka/Docker service basics already documented |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Ports & Networking | Existing `9094`, `3306`, `9092` port explanations reused |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_2}` | Database Analysis | `order_db`, base tables, and migration usage reused |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_2}` | Migration Order | Task 5 index must be applied after earlier schema migrations |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_3}` | Cart/Product/Inventory setup | `CreateOrder` RPC delegates to Task 3 checkout flow |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_3}` | Common Errors and Fixes | Cart/Product and migration mismatch issues reused |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_4}` | Payment coordination setup | `CreateOrderResponse.payment_action` depends on existing payment handoff |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_4}` | Environment Variables | Payment return URL and payment timeout validation reused |

## 15. Final Checklist

- [ ] Previous dependency documentation checked first
- [ ] `${INPUT_FILE_PATH}` reviewed
- [ ] Existing gRPC/proto implementation files identified
- [ ] No duplicate Go/MySQL/Docker/Kafka installation docs added
- [ ] No original implementation file overwritten
- [ ] No new Go dependency required for `${TASK_FILE_NAME}`
- [ ] Buf workflow documented for proto changes
- [ ] `ORDER_GRPC_ADDR` confirmed for local runtime
- [ ] `ORDER_GRPC_TRUSTED_CALLER_TOKEN` is 32+ chars
- [ ] `ORDER_PAGE_TOKEN_SIGNING_KEY` is 32+ chars
- [ ] Gateway/internal caller sends `x-internal-token`
- [ ] Gateway/internal caller sends `x-user-id`
- [ ] Gateway/internal caller sends correct `x-roles`
- [ ] Seller routes send `x-seller-id` when required
- [ ] MySQL migrations `001`, `002`, and `003` applied for real Task 5 DB runtime
- [ ] Task 5 index `idx_orders_user_created_order` verified
- [ ] Kafka configured or `ORDER_EVENTS_ENABLED=false` set locally
- [ ] Focused tests run with `env GOWORK=off`
- [ ] `buf lint` run if proto was touched
- [ ] Missing server bootstrap understood before trying `go run ./cmd/server`
- [ ] Secrets kept out of Git and logs
- [ ] No duplicate setup documentation added
