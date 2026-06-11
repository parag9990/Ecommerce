# Project Dependency & Setup Guide

## Variables Used In This Document

```text
SERVICE_NAME=Order Service
TASK_FILE_NAME=task7.md
INPUT_FILE_PATH=TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}
OUTPUT_FILE_NAME=${TASK_FILE_NAME without ".md"}_Dependency.md
OUTPUT_FILE_PATH=TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}
SERVICE_CODE_PATH=backend/services/order-service
TASK_2_BASE_MIGRATION=${SERVICE_CODE_PATH}/migrations/001_create_order_tables.up.sql
TASK_4_PAYMENT_MIGRATION=${SERVICE_CODE_PATH}/migrations/002_add_payment_coordination.up.sql
TASK_5_QUERY_INDEX_MIGRATION=${SERVICE_CODE_PATH}/migrations/003_add_grpc_query_indexes.up.sql
TASK_8_OUTBOX_MIGRATION=${SERVICE_CODE_PATH}/migrations/004_create_order_outbox_events.up.sql
PREVIOUS_DEPENDENCY_FILE_TASK_1=TaskImplementation/${SERVICE_NAME}/task1_Dependency.md
PREVIOUS_DEPENDENCY_FILE_TASK_2=TaskImplementation/${SERVICE_NAME}/task2_Dependency.md
PREVIOUS_DEPENDENCY_FILE_TASK_3=TaskImplementation/${SERVICE_NAME}/task3_Dependency.md
PREVIOUS_DEPENDENCY_FILE_TASK_4=TaskImplementation/${SERVICE_NAME}/task4_Dependency.md
PREVIOUS_DEPENDENCY_FILE_TASK_5=TaskImplementation/${SERVICE_NAME}/task5_Dependency.md
PREVIOUS_DEPENDENCY_FILE_TASK_6=TaskImplementation/${SERVICE_NAME}/task6_Dependency.md
```

## Previous Dependency Reuse

Is guide ka goal duplicate setup docs banana nahi hai. Common Go, MySQL, Docker, Kafka, gRPC, `.env`, and migration basics pehle dependency files me already explained hain.

Read these first:

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Project tech stack, Go modules, full `.env`, MySQL install, Kafka, Docker, common ports, generic errors | Same backend module and same local infrastructure |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_2}` | MySQL schema setup and base migration flow | `order_idempotency_keys` table yahi migration create karti hai |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_3}` | Cart/Product checkout dependencies and inventory reservation setup | `${TASK_FILE_NAME}` same cart-to-order checkout workflow protect karta hai |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_4}` | Payment coordination env/config and migration `002` | Current create-order runtime payment handoff and inventory reservation columns use karta hai |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_5}` | gRPC metadata, trusted caller token, API verification pattern, port `9094` | `${TASK_FILE_NAME}` same `CreateOrder` RPC boundary par enforce hota hai |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_6}` | Current runtime migration alignment and no-server-bootstrap note | Same current module limitations apply |

Important: Is file me Go install, MySQL install, Docker Compose basics, Kafka install, and full `.env` explanation repeat nahi kiya gaya. Sirf `${TASK_FILE_NAME}` ke idempotency-specific setup, verification, and debugging points explain kiye gaye hain.

## 1. Project Overview

`${INPUT_FILE_PATH}` checkout idempotency enforcement guide hai. Simple meaning: buyer ne same checkout request retry kiya to duplicate order create nahi hona chahiye.

Task-specific runtime rule:

```text
same authenticated user + same idempotency key + same checkout input = same order result
same authenticated user + same idempotency key + changed checkout input = conflict error
```

Current code reality:

| Area | Current Implementation |
|---|---|
| Domain status/validation | `${SERVICE_CODE_PATH}/internal/domain/idempotency.go` |
| MySQL claim repository | `${SERVICE_CODE_PATH}/internal/repository/mysql_idempotency_repository.go` |
| Checkout wrapper usecase | `${SERVICE_CODE_PATH}/internal/usecase/create_order.go` |
| Cart-to-order transaction binding | `${SERVICE_CODE_PATH}/internal/repository/mysql_order_repository.go` |
| gRPC validation/error mapping | `${SERVICE_CODE_PATH}/internal/transport/grpc/validation.go`, `error_mapping.go`, `order_handler.go` |
| Base idempotency table | `${TASK_2_BASE_MIGRATION}` |

Beginner note: Idempotency ek safety lock jaisa hai, but memory lock nahi. Yaha correctness ka source MySQL unique index hai, kyunki service restart ya multiple pods ke baad bhi duplicate order avoid karna hai.

## 2. Tech Stack

### Reused Technologies

Refer:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Sections:
Project Tech Stack Analysis
Go Dependency System
External Services Analysis
Docker and DevOps Setup
```

| Technology | Status For `${TASK_FILE_NAME}` | Setup Impact |
|---|---|---|
| Go | Reused | Same module, same `go.mod`, same test flow |
| gRPC / Protobuf | Reused | Existing `CreateOrder` RPC carries `idempotency_key` |
| MySQL 8+ / InnoDB | Reused but mandatory for real runtime | `order_idempotency_keys` unique row protects checkout |
| Kafka / outbox | Reused if events enabled | Current order creation can write outbox events and publish through existing setup |
| Docker | Reused, optional | Useful for MySQL/Kafka local dependencies |

### Task-Specific Concepts

| Concept / Tool | Required? | Why Used | Beginner-Friendly Explanation |
|---|---:|---|---|
| Idempotency key | Yes | Retry ko same business operation se link karta hai | Client har new checkout ke liye ek unique key bhejta hai. Retry me wahi key reuse hoti hai. |
| MySQL unique index `(user_id, idempotency_key)` | Yes | Parallel requests me sirf ek request winner banegi | Database race condition handle karta hai; app memory par depend nahi hota. |
| `request_hash` | Yes | Same key changed cart/address/coupon ke saath reuse ho to reject | Request ka normalized fingerprint store hota hai. |
| SHA-256 | Yes, standard library | Stable 64-char hash banata hai | Extra install nahi chahiye; Go ke `crypto/sha256` se aata hai. |
| MySQL duplicate key error `1062` | Yes | Existing claim detect karne ke liye | `github.com/go-sql-driver/mysql` duplicate insert ko identify karta hai. |
| Redis lock | No | Not introduced | MySQL unique constraint enough hai for durable correctness. |

## 3. Required Software

| Use Case | Required Software | Notes |
|---|---|---|
| Sirf docs read karna | None | No runtime dependency needed |
| Focused unit/sqlmock tests | Go matching `${SERVICE_CODE_PATH}/go.mod` | Live MySQL/Kafka usually not needed |
| Real repository/manual DB verification | MySQL 8+ and MySQL client | Base idempotency table verify karne ke liye |
| Full checkout runtime | Go, MySQL, Cart/Product/Payment adapters or fakes, gRPC listener, Kafka if events enabled | Current repo me no `cmd/server/main.go`, so direct long-running server bootstrap missing hai |
| Local infra through containers | Docker | Reuse previous Docker setup |

Installation steps for Go, MySQL, Docker, and Kafka are already documented in:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
```

## 4. Dependency Management

`${TASK_FILE_NAME}` does not require a new Go module dependency. Current `go.mod` already includes the needed packages:

| Dependency | New For This Task? | Purpose |
|---|---:|---|
| `github.com/go-sql-driver/mysql` | No | MySQL access and duplicate key error detection |
| `google.golang.org/grpc` | No | Existing RPC server and status errors |
| `google.golang.org/protobuf` | No | Existing generated order API types |
| `github.com/DATA-DOG/go-sqlmock` | No | Repository tests without live MySQL |
| Go standard library `crypto/sha256`, `encoding/hex`, `encoding/json` | No install needed | Deterministic request hashing |

Dependency setup is reused from:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Section:
Go Dependency System
```

Task-specific useful commands:

```bash
cd ${SERVICE_CODE_PATH}
env GOWORK=off go mod download
env GOWORK=off go test ./internal/domain ./internal/repository ./internal/usecase ./internal/transport/grpc
```

Beginner note: `env GOWORK=off` module ko standalone mode me test karta hai. Is repo me service module ke `replace` directive ke through shared generated Go code local path se resolve hota hai.

## 5. Database Setup

### Database Used: MySQL 8+

Full MySQL installation, Docker MySQL setup, credentials, and base schema explanation are already documented in:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Sections:
Database Analysis
Docker and DevOps Setup

${PREVIOUS_DEPENDENCY_FILE_TASK_2}
Sections:
Database Analysis
Migration Order
```

### What `${TASK_FILE_NAME}` Needs From MySQL

`${TASK_FILE_NAME}` does not add a new migration file. It depends on this existing table from `${TASK_2_BASE_MIGRATION}`:

```text
order_idempotency_keys
```

Critical columns/indexes:

| DB Item | Why It Matters |
|---|---|
| `user_id VARCHAR(64)` | Key buyer-specific scope me rahe |
| `idempotency_key VARCHAR(128)` | Client retry token store hota hai |
| `request_hash CHAR(64)` | Same key changed request ke saath reuse ho to detect hota hai |
| `order_id VARCHAR(64) NULL` | Successful claim ko created order se bind karta hai |
| `status ENUM('processing','completed','failed')` | Retry decision ke liye current state |
| `expires_at TIMESTAMP` | Retention/cleanup policy ke liye timestamp |
| `UNIQUE KEY uk_order_idempotency_user_key (user_id, idempotency_key)` | Duplicate checkout prevention ka main safety guard |

### Required Or Optional

| Scenario | MySQL Required? | Migration Required? |
|---|---:|---|
| Read `${INPUT_FILE_PATH}` only | No | No |
| Focused usecase tests with fakes | No | No |
| Repository tests with sqlmock | No live DB | No live migration |
| Manual DB verification | Yes | `${TASK_2_BASE_MIGRATION}` |
| Current full checkout runtime | Yes | Usually `001`, `002`, `003`, and `004` |

Why current full runtime usually needs `001` to `004`: order creation writes inventory reservation columns from migration `002`, query paths use indexes from migration `003`, and current create-order transaction inserts an outbox event into table from migration `004`.

### Migration Order

For a real local DB aligned with the current code, apply migrations in this order:

```bash
mysql -u root -p < ${TASK_2_BASE_MIGRATION}
mysql -u root -p < ${TASK_4_PAYMENT_MIGRATION}
mysql -u root -p < ${TASK_5_QUERY_INDEX_MIGRATION}
mysql -u root -p < ${TASK_8_OUTBOX_MIGRATION}
```

If you only want to verify `${TASK_FILE_NAME}` idempotency table, migration `001` is the minimum.

### Verify Idempotency Schema

```bash
mysql -u root -p order_db -e "SHOW TABLES LIKE 'order_idempotency_keys';"
mysql -u root -p order_db -e "SHOW INDEX FROM order_idempotency_keys WHERE Key_name='uk_order_idempotency_user_key';"
mysql -u root -p order_db -e "DESCRIBE order_idempotency_keys;"
```

Expected important result:

```text
uk_order_idempotency_user_key should exist on user_id + idempotency_key
request_hash should be CHAR(64)
status should allow processing/completed/failed
```

### Credentials Placement

No new credential file is introduced by `${TASK_FILE_NAME}`.

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

Security note: DSN me DB password hota hai. `.env`, shell history, logs, screenshots, and Git commits me leak mat karo.

## 6. Redis / Queue / External Services

### Redis

No Redis setup is introduced by `${TASK_FILE_NAME}`.

Redis wider platform me session/cart caching ke liye ho sakta hai, but current `${SERVICE_NAME}` idempotency implementation Redis import/config nahi karta. Checkout duplicate prevention MySQL unique index se hoti hai.

### Kafka / Outbox

Kafka setup is reused from previous dependency files. `${TASK_FILE_NAME}` directly Kafka broker claim nahi karta, but current create-order path can insert an `OrderCreated` outbox event.

| Scenario | Kafka Needed? | Notes |
|---|---:|---|
| Focused idempotency usecase tests | No | Fakes/sqlmock enough |
| Manual MySQL table verification | No | MySQL only |
| Full runtime with events disabled | Broker no, outbox table still can be needed | Set `ORDER_EVENTS_ENABLED=false` for local no-Kafka config |
| Full runtime with events enabled | Yes | `ORDER_KAFKA_BROKERS` required |

Refer:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Section:
External Services Analysis -> Kafka

${PREVIOUS_DEPENDENCY_FILE_TASK_4}
Section:
Kafka / Queue
```

### Cart, Product, Inventory, Payment

`${TASK_FILE_NAME}` protects the checkout workflow that was already documented in earlier tasks.

| External Boundary | Setup Status | Why Relevant |
|---|---|---|
| Cart Service | Reused from `${PREVIOUS_DEPENDENCY_FILE_TASK_3}` | Checkout reads active cart |
| Product Service | Reused from `${PREVIOUS_DEPENDENCY_FILE_TASK_3}` | Product snapshot and inventory reservation happen before order save |
| Inventory reservation | Reused from `${PREVIOUS_DEPENDENCY_FILE_TASK_3}` and `${PREVIOUS_DEPENDENCY_FILE_TASK_4}` | `${TASK_FILE_NAME}` sends stable child idempotency keys for reservation/release |
| Payment Service | Reused from `${PREVIOUS_DEPENDENCY_FILE_TASK_4}` | Existing order result can re-initiate or return payment action safely |

Current code still uses injected ports for these dependencies. No checked-in env variables like `ORDER_CART_GRPC_ADDR` or `ORDER_PRODUCT_GRPC_ADDR` are loaded yet.

### RabbitMQ / NATS / MinIO / Elasticsearch

No RabbitMQ, NATS, MinIO, Elasticsearch, SMTP, Stripe, Twilio, Firebase, or OAuth provider setup is introduced by `${TASK_FILE_NAME}`.

## 7. Environment Variables

The full `.env` example is already documented in:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Section:
Environment Variables
```

`${TASK_FILE_NAME}` introduces no brand-new environment variable. It depends heavily on one existing checkout variable:

| Variable | New? | Required? | Task-Specific Purpose | Example |
|---|---:|---:|---|---|
| `ORDER_IDEMPOTENCY_TTL` | No | Optional with default | New claim ka `expires_at` calculate karta hai | `24h` |
| `ORDER_MYSQL_DSN` | No | Yes for real DB runtime | Idempotency row claim/read/update | `order_user:order_password@tcp(localhost:3306)/order_db?parseTime=true` |
| `ORDER_INVENTORY_RESERVATION_TTL` | No | Optional with default | Downstream inventory reservation TTL | `15m` |
| `ORDER_INVENTORY_RELEASE_TIMEOUT` | No | Optional with default | Failed order persistence ke baad reservation release timeout | `3s` |
| `ORDER_PAYMENT_RETURN_URL` | No | Yes for current config validation | Payment handoff config still validates during full app config load | `https://localhost.example/checkout/result` |
| `ORDER_GRPC_ADDR` | No | Optional, default `:9094` | Existing gRPC listener address | `:9094` |
| `ORDER_GRPC_TRUSTED_CALLER_TOKEN` | No | Yes | Trusted gateway/internal caller authentication | `32+ chars` |
| `ORDER_PAGE_TOKEN_SIGNING_KEY` | No | Yes for current full config | Pagination signing used by list APIs | `32+ chars` |
| `ORDER_EVENTS_ENABLED` | No | Optional default true | Local no-Kafka setup ke liye `false` kar sakte ho | `false` |
| `ORDER_KAFKA_BROKERS` | No | Required if events enabled | Existing outbox publisher broker list | `localhost:9092` |

Task-specific minimal local exports using existing variables:

```bash
export ORDER_IDEMPOTENCY_TTL=24h
export ORDER_MYSQL_DSN='order_user:order_password@tcp(localhost:3306)/order_db?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci'
export ORDER_PAYMENT_RETURN_URL=https://localhost.example/checkout/result
export ORDER_GRPC_TRUSTED_CALLER_TOKEN=local-trusted-caller-token-minimum-32-chars
export ORDER_PAGE_TOKEN_SIGNING_KEY=local-page-token-signing-key-minimum-32
export ORDER_EVENTS_ENABLED=false
```

If events are enabled:

```bash
export ORDER_KAFKA_BROKERS=localhost:9092
```

Where to place `.env`:

```text
${SERVICE_CODE_PATH}/.env
```

Common beginner mistake: code uses `os.Getenv`, so `.env` is not auto-loaded. Source it before running commands:

```bash
set -a
. ${SERVICE_CODE_PATH}/.env
set +a
```

## 8. Docker Setup

`${TASK_FILE_NAME}` introduces no new Dockerfile, Docker Compose service, container, volume, network, health check, restart policy, or exposed port.

Reuse previous Docker setup:

| Container / Service | Status | Reference |
|---|---|---|
| MySQL | Reused | `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` and `${PREVIOUS_DEPENDENCY_FILE_TASK_2}` |
| Kafka | Reused only if events enabled | `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` |
| Redis | Not used by this task | No setup required |

Docker reminder: Host machine commands usually use `localhost:3306`; container-to-container DSNs usually use service DNS like `mysql:3306`. This networking behavior is already explained in earlier Docker docs.

### Ports & Networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| MySQL | `3306` | `order_db` and `order_idempotency_keys` persistence | Reused |
| gRPC API | `9094` | Existing order RPC listener if server bootstrap is added | Reused |
| Kafka broker | `9092` | Outbox/event publishing when enabled | Reused |
| Kafka controller | `9093` | Local KRaft controller in earlier Docker example | Reused |
| Redis | `6379` | Not used directly by `${SERVICE_NAME}` `${TASK_FILE_NAME}` | Not used |

No new firewall rule is needed only for `${TASK_FILE_NAME}`.

## 9. Local Development Setup

### Step 1: Read Previous Dependency Documentation First

Read:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
${PREVIOUS_DEPENDENCY_FILE_TASK_2}
${PREVIOUS_DEPENDENCY_FILE_TASK_3}
${PREVIOUS_DEPENDENCY_FILE_TASK_4}
${PREVIOUS_DEPENDENCY_FILE_TASK_5}
${PREVIOUS_DEPENDENCY_FILE_TASK_6}
```

They explain base Go setup, MySQL, migrations, Docker, Kafka, checkout dependencies, payment coordination, gRPC metadata, and current runtime limitations.

### Step 2: Go To Project Directory

```bash
cd ${SERVICE_CODE_PATH}
```

### Step 3: Install Only New Dependencies If Needed

No new dependency is required only for `${TASK_FILE_NAME}`.

If dependencies are not downloaded:

```bash
env GOWORK=off go mod download
```

### Step 4: Setup Only New Databases/Services If Needed

No new database server is introduced.

For real DB verification, make sure MySQL is running and apply migrations:

```bash
mysql -u root -p < ${TASK_2_BASE_MIGRATION}
mysql -u root -p < ${TASK_4_PAYMENT_MIGRATION}
mysql -u root -p < ${TASK_5_QUERY_INDEX_MIGRATION}
mysql -u root -p < ${TASK_8_OUTBOX_MIGRATION}
```

### Step 5: Add Only New Or Changed Environment Variables

No new env variable. For `${TASK_FILE_NAME}`, verify `ORDER_IDEMPOTENCY_TTL` is valid if you override it:

```bash
export ORDER_IDEMPOTENCY_TTL=24h
```

For full config load, also set the existing required variables from section 7.

### Step 6: Run Migrations If Needed

For idempotency table only:

```bash
mysql -u root -p < ${TASK_2_BASE_MIGRATION}
```

For current full checkout runtime:

```bash
mysql -u root -p < ${TASK_2_BASE_MIGRATION}
mysql -u root -p < ${TASK_4_PAYMENT_MIGRATION}
mysql -u root -p < ${TASK_5_QUERY_INDEX_MIGRATION}
mysql -u root -p < ${TASK_8_OUTBOX_MIGRATION}
```

### Step 7: Start Backend Service

Current repo has gRPC server package code, but no checked-in:

```text
${SERVICE_CODE_PATH}/cmd/server/main.go
```

So a beginner should not expect this to work today:

```bash
go run ./cmd/server
```

After bootstrap exists, it should wire:

| Dependency | Why Needed |
|---|---|
| Config loader | Reads `ORDER_MYSQL_DSN`, `ORDER_IDEMPOTENCY_TTL`, gRPC and event settings |
| MySQL connection | Claim/read/update `order_idempotency_keys` and save order |
| `MySQLIdempotencyRepository` | Atomic idempotency claim implementation |
| `MySQLOrderRepository` | Save order and attach claim in one transaction |
| Cart/Product/Payment adapters or local fakes | Full checkout workflow dependencies |
| gRPC listener | Exposes existing `CreateOrder` RPC |
| Outbox worker/Kafka config | Needed only when event publishing is enabled |

### Step 8: Verify APIs Or Functionality Related To `${TASK_FILE_NAME}`

Focused test command:

```bash
cd ${SERVICE_CODE_PATH}
env GOWORK=off go test ./internal/domain ./internal/repository ./internal/usecase ./internal/transport/grpc
```

Manual DB verification:

```bash
mysql -u root -p order_db -e "SHOW INDEX FROM order_idempotency_keys WHERE Key_name='uk_order_idempotency_user_key';"
```

Live gRPC verification after server bootstrap and downstream dependencies exist:

```bash
grpcurl -plaintext \
  -H 'x-internal-token: local-trusted-caller-token-minimum-32-chars' \
  -H 'x-user-id: usr_1001' \
  -H 'x-roles: buyer' \
  -H 'x-request-id: req_demo_1' \
  -d '{
    "cartId": "cart_1001",
    "idempotencyKey": "checkout_demo_001",
    "shippingAddress": {
      "recipientName": "Demo User",
      "phone": "9999999999",
      "line1": "Test Street",
      "city": "Bengaluru",
      "state": "KA",
      "postalCode": "560001",
      "countryCode": "IN"
    }
  }' \
  localhost:9094 ecommerce.order.v1.OrderService/CreateOrder
```

Verification behavior:

| Test | Expected Result |
|---|---|
| Same request + same `idempotencyKey` retry | Existing order result returned |
| Same key + changed address/cart | gRPC `AlreadyExists` conflict |
| Two parallel same-key requests | One wins; other can receive `Aborted` while processing or replay after completion |
| Missing key | gRPC `InvalidArgument` |
| Key longer than 128 chars | gRPC `InvalidArgument` |

## 10. Running The Project

### Documentation-Only Flow

No runtime command needed. Read `${INPUT_FILE_PATH}` and this file.

### Task-Specific Test Flow

```bash
cd ${SERVICE_CODE_PATH}
env GOWORK=off go test ./internal/domain ./internal/repository ./internal/usecase ./internal/transport/grpc
```

### Real DB Verification Flow

```bash
mysql -u root -p < ${TASK_2_BASE_MIGRATION}
mysql -u root -p order_db -e "DESCRIBE order_idempotency_keys;"
```

### Full Runtime Flow

Full runtime needs the missing server bootstrap plus downstream adapters. Once those exist:

1. Start MySQL.
2. Apply migrations `001` to `004`.
3. Start Kafka or set `ORDER_EVENTS_ENABLED=false`.
4. Source `.env`.
5. Start `${SERVICE_NAME}` gRPC server on `ORDER_GRPC_ADDR`.
6. Call `CreateOrder` with a buyer role and `idempotency_key`.
7. Retry same request with same key and confirm no duplicate order row is created.

## 11. Common Errors & Fixes

Generic Go, MySQL, Docker, Kafka, `.env`, and gRPC errors are already documented in previous dependency files. This section only lists `${TASK_FILE_NAME}`-specific or newly relevant errors.

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `Table 'order_idempotency_keys' doesn't exist` | Base migration `001` not applied | Apply `${TASK_2_BASE_MIGRATION}` | Always run migrations before real DB testing |
| Duplicate orders created on retry | Unique index missing or code bypasses idempotency repository | Verify `uk_order_idempotency_user_key`; route checkout through `CreateOrderUsecase` | Do not create orders directly from retryable transport paths |
| gRPC `AlreadyExists` on checkout | Same key reused with changed cart/address/coupon fingerprint | Generate a fresh idempotency key for changed checkout input | Frontend should bind key to one checkout attempt |
| gRPC `Aborted` with message about processing | Same key is currently being processed by another request | Retry same request/key after short delay | Client retry policy should use backoff |
| gRPC `FailedPrecondition` for previous failure | Key is marked failed before order creation | Create a new checkout attempt with a new key after fixing cart/input | Do not reuse failed keys forever |
| `ORDER_IDEMPOTENCY_TTL must be greater than zero` | Env override is `0`, negative, or invalid duration | Use `24h`, `12h`, etc. | Keep duration strings in Go format |
| `idempotency key is invalid` | Key is blank after trim or longer than 128 chars | Send non-empty key within 128 chars | Use UUID/ULID-like key from client/gateway |
| `Unknown column inventory_reservation_id` | Current runtime used without migration `002` | Apply `${TASK_4_PAYMENT_MIGRATION}` | Apply current migrations in order |
| `Table 'order_outbox_events' doesn't exist` | Current order creation inserts outbox event but migration `004` missing | Apply `${TASK_8_OUTBOX_MIGRATION}` | Use full current runtime migration set |
| Retry returns payment action again | Order exists and payment is still payable | This can be expected; Payment Service must be idempotent by order/payment key | Keep Payment Service idempotency from `${PREVIOUS_DEPENDENCY_FILE_TASK_4}` |

## 12. Security & Best Practices

### Task-Specific Security Rules

| Rule | Why |
|---|---|
| Never trust `user_id` from request body | Idempotency scope must use authenticated actor from gateway metadata |
| Do not log raw `idempotency_key` | Key can reveal client retry behavior; current code logs hashed reference |
| Do not log full `request_hash` in normal info logs | It can become sensitive operational data |
| Use parameterized SQL only | Prevents SQL injection and keeps MySQL driver escaping correct |
| Keep `ORDER_MYSQL_DSN` secret | DB password exposure gives direct data access |
| Keep `ORDER_GRPC_TRUSTED_CALLER_TOKEN` secret | Browser/frontend should never know internal gateway token |
| Do not delete `processing` rows blindly | Crash may have committed an order before status completion |
| Do not replace MySQL uniqueness with only Redis/memory lock | Durable retries across restarts need database-backed state |

### Beginner-Friendly Best Practices

- Client should create one fresh idempotency key for one new checkout attempt.
- Network timeout or 5xx retry should reuse the same key and same request payload.
- Changed cart/address/coupon should use a new key.
- Keep `ORDER_IDEMPOTENCY_TTL` long enough for payment page retries and mobile network retries.
- Monitor conflict/in-progress counts, but do not put `user_id` or raw key in metric labels.
- Test two parallel requests with the same key before production release.
- Apply migrations in order, especially when current code has advanced beyond the original task boundary.

## 13. Missing Or Misconfigured Things

| Area | Current Gap | Setup Impact | Suggested Fix |
|---|---|---|---|
| Server bootstrap | No `${SERVICE_CODE_PATH}/cmd/server/main.go` found | Cannot start a long-running service with `go run ./cmd/server` | Add bootstrap wiring config, DB, repositories, adapters, gRPC listen, and graceful shutdown |
| Idempotency cleanup | `expires_at` exists, but no cleanup worker/command found | Old keys may accumulate | Add safe cleanup job that avoids active/recent `processing` rows |
| Downstream service config | No loaded `ORDER_CART_GRPC_ADDR`, `ORDER_PRODUCT_GRPC_ADDR`, or payment client address variables | Full live checkout adapters cannot be configured from env yet | Add adapter config when concrete clients are implemented |
| Docker Compose | No committed compose file for this service module | Beginners must copy commands from docs | Add repo-level compose for MySQL/Kafka if team wants one-command local infra |
| Replay response marker | Current proto response has no explicit `idempotency_replayed` flag | Client receives same order but may not know response was replayed | Add optional response metadata/field only if product needs it |
| Processing recovery | No reconciliation command for stuck `processing` claims found | Operational recovery is manual | Add admin/reconciliation workflow before production traffic |
| Kafka security config | Current config exposes broker list, not SASL/TLS env vars | Production Kafka auth not represented | Add TLS/SASL config when deploying to secured Kafka |

No hardcoded DB password or payment secret was found in the `${TASK_FILE_NAME}` idempotency files reviewed. Keep checking before deployment because secrets sometimes appear in bootstrap/config later.

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Project Tech Stack Analysis | Go, gRPC, MySQL, Kafka, Docker basics already explained |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Go Dependency System | Same `go.mod`, `go.sum`, module download, and common Go issues |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Environment Variables | Full `.env` example already includes `ORDER_IDEMPOTENCY_TTL` and runtime secrets |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Docker and DevOps Setup | MySQL/Kafka container setup and port conflicts already documented |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_2}` | Database Analysis | Base `order_db` schema and `order_idempotency_keys` table reused |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_2}` | Migration Order | Migration `001` creates the idempotency table |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_3}` | Cart/Product checkout setup | `${TASK_FILE_NAME}` protects the same cart-to-order workflow |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_3}` | Inventory reservation dependency notes | `${TASK_FILE_NAME}` uses stable child keys for downstream reservation/release |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_4}` | Payment coordination setup | `${TASK_FILE_NAME}` replay path can re-enter payment initiation safely |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_4}` | Migration `002` notes | Current runtime needs inventory reservation columns |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_5}` | gRPC metadata and auth | Same trusted caller token and buyer metadata required for `CreateOrder` |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_5}` | Ports & Networking | Existing gRPC `9094`, MySQL `3306`, Kafka `9092` guidance reused |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_6}` | Current runtime migration alignment | Same `001` to `004` current-code readiness applies |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_6}` | Missing bootstrap note | Same no-`cmd/server` limitation applies |

## 15. Final Checklist

- [ ] Previous dependency documentation checked first
- [ ] `${INPUT_FILE_PATH}` reviewed
- [ ] No original implementation file overwritten
- [ ] No duplicate Go/MySQL/Docker/Kafka installation docs added
- [ ] No new Go module dependency required only for `${TASK_FILE_NAME}`
- [ ] `order_idempotency_keys` table exists
- [ ] `uk_order_idempotency_user_key` unique index exists on `user_id, idempotency_key`
- [ ] `request_hash` column is `CHAR(64)`
- [ ] `ORDER_IDEMPOTENCY_TTL` is positive if overridden
- [ ] `ORDER_MYSQL_DSN` points to `order_db`
- [ ] `ORDER_GRPC_TRUSTED_CALLER_TOKEN` is 32+ chars
- [ ] `ORDER_PAYMENT_RETURN_URL` is valid HTTPS for current config validation
- [ ] Kafka configured or `ORDER_EVENTS_ENABLED=false` set locally
- [ ] Migration `001` applied for `${TASK_FILE_NAME}` table verification
- [ ] Migrations `001`, `002`, `003`, and `004` applied for current full runtime
- [ ] Gateway/internal caller sends trusted `x-internal-token`
- [ ] Gateway/internal caller sends buyer `x-user-id` and `x-roles: buyer`
- [ ] Checkout request sends non-empty `idempotency_key` within 128 chars
- [ ] Same request/key retry verified
- [ ] Same key changed request conflict verified
- [ ] Parallel same-key request behavior verified
- [ ] Raw idempotency keys not logged
- [ ] Stuck `processing` recovery plan understood
- [ ] Missing server bootstrap understood before trying `go run ./cmd/server`
- [ ] Secrets kept out of Git and logs
- [ ] No duplicate setup documentation added
