# Project Dependency & Setup Guide

## Variables Used In This Document

```text
SERVICE_NAME=Order Service
TASK_FILE_NAME=task3.md
INPUT_FILE_PATH=TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}
OUTPUT_FILE_NAME=task3_Dependency.md
OUTPUT_FILE_PATH=TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}
SERVICE_CODE_PATH=backend/services/order-service

PREVIOUS_DEPENDENCY_FILE_TASK_1=TaskImplementation/${SERVICE_NAME}/task1_Dependency.md
PREVIOUS_DEPENDENCY_FILE_TASK_2=TaskImplementation/${SERVICE_NAME}/task2_Dependency.md

TASK_3_USECASE=${SERVICE_CODE_PATH}/internal/usecase/create_order_from_cart.go
TASK_3_CONTRACTS=${SERVICE_CODE_PATH}/internal/usecase/contracts.go
TASK_3_DOMAIN=${SERVICE_CODE_PATH}/internal/domain/checkout.go
TASK_3_TEST=${SERVICE_CODE_PATH}/internal/usecase/create_order_from_cart_test.go
```

Use `${INPUT_FILE_PATH}` as the source task guide. This file is saved as `${OUTPUT_FILE_PATH}`.

## Previous Dependency Reuse

Before writing this file, previous dependency docs inside the same `${SERVICE_NAME}` folder were checked.

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Project Tech Stack Analysis | Go, gRPC, Protobuf, MySQL, Kafka, Docker basics already explained |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Go Dependency System | Same Go module, `go.mod`, `go.sum`, and workspace setup |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Environment Variables | Existing order-service env loading and full `.env` example already documented |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | External Services Analysis | MySQL, Kafka, Docker, ports, and general runtime setup already documented |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Docker and DevOps Setup | Local MySQL/Kafka Docker setup already documented |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_2}` | Database Analysis | Base `order_db` schema and Task 2 migration flow already documented |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_2}` | Migration Order | Base migration and later migration awareness already documented |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_2}` | Common Errors and Fixes | MySQL schema/migration errors already documented |

Important: Is file me same installation steps repeat nahi kiye gaye. Sirf `${TASK_FILE_NAME}` ke new checkout-specific dependencies, runtime assumptions, verification commands, and gaps explain kiye gaye hain.

## 1. Project Overview

`${TASK_FILE_NAME}` ka focus cart-to-order flow hai:

```text
GetCart -> validate cart -> fetch fresh product snapshots -> reserve inventory -> create order
```

Simple Hinglish: Checkout ke time cart ke old price ko blindly trust nahi karna. `${SERVICE_NAME}` Cart Service se cart leta hai, Product Service se latest price/stock verify karta hai, inventory reserve karta hai, phir MySQL transaction me order snapshot save karta hai.

### What Task 3 Adds

| Area | New for Task 3? | Notes |
|---|---:|---|
| Cart Service dependency | Yes | `GetCart` call required for live checkout |
| Product Service dependency | Yes | `BatchGetProducts`, `ReserveInventory`, and `ReleaseInventory` required |
| Fresh price snapshot | Yes | Order item price Product Service se aata hai, cart se nahi |
| Inventory reservation TTL | Yes, reused env | Existing `ORDER_INVENTORY_RESERVATION_TTL` config Task 3 flow me actively used hota hai |
| Inventory release timeout | Yes, reused env | Existing `ORDER_INVENTORY_RELEASE_TIMEOUT` compensation calls me used hota hai |
| Shipping address snapshot | Yes | Order creation request me required |
| Idempotency claim binding | Yes | Direct duplicate checkout prevent karne ke liye claim required |
| MySQL transaction | Reused DB, new flow | Order, items, history, idempotency claim, and current-code outbox event same transaction me persist hote hain |

### Current Code Reality

`${INPUT_FILE_PATH}` originally documents Task 3 as a cart-to-order guide and says events are later scope. But current checked-in runtime code already builds an `OrderCreated` outbox event in `${TASK_3_USECASE}` and persists it through the repository.

Practical meaning:

- Unit tests for Task 3 do not need live MySQL, Cart Service, Product Service, or Kafka.
- Real repository/runtime tests need the correct MySQL schema.
- Current full repository runtime may need later migrations too, because current code has evolved beyond the original Task 3 boundary.

## 2. Tech Stack

### Reused Technologies

These are already explained in `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` and `${PREVIOUS_DEPENDENCY_FILE_TASK_2}`:

| Technology | Reuse Note |
|---|---|
| Go | Same backend module; no new language runtime introduced |
| Go modules | Same `go.mod`, `go.sum`, and `backend/go.work` setup |
| gRPC | Same internal service communication style |
| Protocol Buffers | Same generated shared Go contracts |
| MySQL 8+ | Same `${SERVICE_NAME}` database owner |
| Docker | Same optional local infrastructure approach |
| Kafka | Same event broker setup when outbox publishing is enabled |

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

### Task-3-Specific Technologies And Concepts

| Technology / Concept | Required? | Why Used | Beginner Explanation |
|---|---:|---|---|
| Cart Service client | Yes for live checkout | Active cart fetch karne ke liye | Cart Service user ka current cart own karta hai. `${SERVICE_NAME}` directly cart DB read nahi karega. |
| Product Service client | Yes for live checkout | Fresh product, price, seller, and stock validate karne ke liye | Product Service product ka source of truth hai. Checkout ke time latest data wahi se lena safe hai. |
| Inventory reservation | Yes for live checkout | Stock race condition avoid karne ke liye | Reservation ka matlab stock temporarily hold karna, taaki payment ke pehle koi aur same stock na le le. |
| Compensation release | Yes | DB save fail ho to reserved inventory release karna | Agar order save nahi hua, Product Service ko reservation release karna padta hai. |
| Idempotency claim | Yes | Duplicate checkout request se duplicate order avoid karna | Same `idempotency_key` repeat ho to system same outcome handle kare, double order na banaye. |
| Address snapshot | Yes | Order ke saath shipping address freeze karna | User profile later change kare to old order ka address change nahi hona chahiye. |

## 3. Required Software

For only reading `${INPUT_FILE_PATH}`:

| Software | Required? | Notes |
|---|---:|---|
| Markdown viewer/editor | Yes | Task guide read karne ke liye |
| Go | No | Sirf docs read kar rahe ho to required nahi |
| MySQL | No | Sirf docs read kar rahe ho to required nahi |

For Task 3 unit tests:

| Software | Required? | Notes |
|---|---:|---|
| Go `1.26.3` | Yes | Module declares this version |
| Internet/module cache | Usually yes | First `go test` dependencies download karega |
| MySQL | No | Task 3 usecase tests fake clients/repositories use karte hain |
| Cart Service | No | Tests fake Cart client use karte hain |
| Product Service | No | Tests fake Product client use karte hain |

For live checkout runtime:

| Software / Service | Required? | Notes |
|---|---:|---|
| MySQL 8+ | Yes | Orders and idempotency state persist karne ke liye |
| Cart Service | Yes | `GetCart` ke liye |
| Product Service | Yes | `BatchGetProducts`, `ReserveInventory`, `ReleaseInventory` ke liye |
| Kafka | If events enabled | Current config default events enabled rakhta hai |
| Docker | Optional | Local MySQL/Kafka dependencies run karne ke liye useful |

Installation for Go, MySQL, Docker, and Kafka is already covered in:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
```

## 4. Dependency Management

No new Go library is introduced only because of `${TASK_FILE_NAME}`.

Current module dependency files:

```text
${SERVICE_CODE_PATH}/go.mod
${SERVICE_CODE_PATH}/go.sum
backend/go.work
backend/shared/gen/go/go.mod
```

The existing dependencies already cover Task 3:

| Dependency | Why Task 3 Needs It |
|---|---|
| `google.golang.org/grpc` | Future concrete Cart/Product gRPC clients and current Order gRPC transport |
| `google.golang.org/protobuf` | Generated proto message support |
| `github.com/go-sql-driver/mysql` | MySQL persistence through repository code |
| `github.com/DATA-DOG/go-sqlmock` | Repository unit tests without real MySQL |
| `github.com/segmentio/kafka-go` | Current outbox publishing support when events are enabled |

Refer:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Section:
Go Dependency System
```

Task-specific verification command:

```bash
cd backend/services/order-service
env GOWORK=off go test ./internal/usecase -run CreateOrderFromCart
```

Why `GOWORK=off`? Current `backend/go.work` references `auth-service`, but this tree does not currently contain `backend/services/auth-service/go.mod`. Isliye module-isolated test run beginner ke liye less confusing hai.

## 5. Database Setup

### Database Used: MySQL 8+

Full MySQL explanation, local installation, Docker setup, and base schema are already covered in:

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

### Why Task 3 Uses MySQL

Task 3 order creation needs one durable transaction:

- insert `orders`
- insert `order_items`
- insert `order_status_history`
- attach existing `order_idempotency_keys` claim
- current code also inserts `order_outbox_events`

Simple Hinglish: Order create ek half-done state me nahi rehna chahiye. Ya to order header/items/history/claim/outbox saath me commit honge, ya rollback hoga.

### Required Or Optional

| Use Case | MySQL Required? |
|---|---:|
| Reading `${INPUT_FILE_PATH}` | No |
| Running Task 3 usecase unit tests | No |
| Running repository tests with `sqlmock` | No |
| Running real checkout persistence | Yes |
| Running full service runtime | Yes |

### Migration Requirement For Task 3

For original Task 3 concept, the base order tables from Task 2 are the foundation.

For current checked-in runtime code, use this practical rule:

| Scenario | Migrations Needed | Why |
|---|---|---|
| Read docs only | None | No DB touched |
| Unit test `${TASK_3_USECASE}` | None | Fakes are used |
| Task 3 original DB foundation | `001_create_order_tables.up.sql` | Base order, item, history, idempotency tables |
| Current real `CreateOrderAndAttachClaim` runtime | `001`, `002`, and `004` minimum | Code writes inventory reservation columns and outbox event rows |
| Full current module runtime | `001`, `002`, `003`, `004` in order | Keeps all current repository/query/event code aligned |

Apply full current runtime migrations from repository root:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 < backend/services/order-service/migrations/001_create_order_tables.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/002_add_payment_coordination.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/003_add_grpc_query_indexes.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/004_create_order_outbox_events.up.sql
```

Verify Task 3 runtime-required columns/tables:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW COLUMNS FROM orders LIKE 'inventory_reservation_id';"
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW COLUMNS FROM orders LIKE 'inventory_reserved_until';"
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW TABLES LIKE 'order_outbox_events';"
```

### Idempotency Claim Requirement

Direct `${TASK_3_USECASE}` expects an already acquired idempotency claim in the command. In normal flow, the wrapper usecase creates this claim before calling cart-to-order logic.

Beginner note: Agar direct repository/manual integration test bana rahe ho, pehle `order_idempotency_keys` me `processing` claim hona chahiye. Warna order attach step fail ho sakta hai.

## 6. Redis / Queue / External Services

### Cart Service

| Question | Answer |
|---|---|
| What is it? | Cart Service active user/guest cart manage karta hai |
| Why Task 3 needs it | `${SERVICE_NAME}` cart owner nahi hai; checkout ke time `GetCart` se current cart snapshot lena hota hai |
| Mandatory? | Yes for live checkout |
| Required for unit tests? | No, fake client is used |
| Direct DB access from `${SERVICE_NAME}`? | No |

Expected Cart Service method:

```text
GetCart(user_id, cart_id) -> cart_id, user_id, currency, items(product_id, variant_id, quantity)
```

Current repo note:

```text
backend/services/cart-service
```

is not present in the current service tree. `${TASK_3_CONTRACTS}` defines a `CartClient` interface, but no concrete Cart gRPC adapter or Cart Service runtime is checked in yet.

Setup implication:

- Unit tests can run now.
- Live end-to-end checkout cannot be completed until Cart Service and its adapter are implemented.
- Order Service should not connect directly to Cart MongoDB/Redis; it should call Cart Service.

### Product Service

| Question | Answer |
|---|---|
| What is it? | Product Service product, variant, seller, price, and inventory source of truth hai |
| Why Task 3 needs it | Fresh price snapshot, availability validation, and inventory reservation ke liye |
| Mandatory? | Yes for live checkout |
| Required for unit tests? | No, fake client is used |
| Direct DB access from `${SERVICE_NAME}`? | No |

Expected Product Service methods from `${TASK_3_CONTRACTS}`:

```text
BatchGetProducts(items)
ReserveInventory(user_id, cart_id, idempotency_key, ttl, items)
ReleaseInventory(reservation_id, idempotency_key, reason)
```

Data required from Product Service:

| Field | Why Required |
|---|---|
| `product_id` and `variant_id` | Cart item ko Product snapshot se match karne ke liye |
| `seller_id` | Seller order view/fulfillment ke liye |
| `sku` | Fulfillment/debug ke liye |
| `title` and `image_url` | Immutable order item snapshot ke liye |
| `currency` and `unit_amount` | Fresh checkout price ke liye |
| `published`, `variant_active`, `in_stock` | Checkout validation ke liye |

Current repo note: `${TASK_3_CONTRACTS}` defines the interface, but a concrete Product gRPC adapter is not checked in for `${SERVICE_CODE_PATH}`.

### Redis

`${SERVICE_NAME}` does not directly use Redis for Task 3.

Redis may be used indirectly by Cart Service for hot cart state. Do not add Redis credentials to `${SERVICE_NAME}` unless Order Service code actually starts using Redis.

Refer:

```text
docs/04-microservice-design.md
Section:
Cart Service
```

### Kafka / Queue

`${INPUT_FILE_PATH}` does not introduce event publishing as Task 3 scope. However current runtime code creates an outbox event row during order creation.

| Item | Required? | Notes |
|---|---:|---|
| `order_outbox_events` table | Yes for current real repository runtime | Created by migration `004_create_order_outbox_events.up.sql` |
| Kafka broker | Only if publishing worker/events enabled | Reuse Kafka setup from `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` |
| RabbitMQ/NATS | No | Not used by current `${SERVICE_NAME}` code |

For quick local runtime without Kafka publishing:

```bash
export ORDER_EVENTS_ENABLED=false
```

Warning: Even with event publishing disabled, current repository order creation still expects the outbox table if `CreateOrderAndAttachClaim` is used.

## 7. Environment Variables

No brand-new environment variable is introduced only by `${TASK_FILE_NAME}`. The required config system is already documented in:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Section:
Environment Variables
```

### Task-3-Relevant Existing Variables

| Variable | New? | Required? | Task 3 Purpose |
|---|---:|---:|---|
| `ORDER_MYSQL_DSN` | No | Yes for real DB runtime | MySQL order persistence |
| `ORDER_INVENTORY_RESERVATION_TTL` | No | Optional with default | TTL sent to Product Service inventory reservation |
| `ORDER_INVENTORY_RELEASE_TIMEOUT` | No | Optional with default | Timeout for Product Service reservation release compensation |
| `ORDER_IDEMPOTENCY_TTL` | No | Optional with default | Wrapper checkout idempotency claim expiry |
| `ORDER_EVENTS_ENABLED` | No | Optional | Disable Kafka publishing locally if broker is not running |
| `ORDER_KAFKA_BROKERS` | No | Required if events enabled | Kafka broker list |
| `ORDER_GRPC_TRUSTED_CALLER_TOKEN` | No | Yes for gRPC server | Internal caller auth token |

Task-specific minimal local exports for usecase/config checks:

```bash
export ORDER_INVENTORY_RESERVATION_TTL=15m
export ORDER_INVENTORY_RELEASE_TIMEOUT=3s
```

These variables are not automatically loaded from `.env`. Existing config uses `os.Getenv`, so source the file first:

```bash
set -a
. backend/services/order-service/.env
set +a
```

### Missing Env Variables For Live Cart/Product Integration

Current `${SERVICE_CODE_PATH}` has no checked-in config variables like:

```text
ORDER_CART_GRPC_ADDR
ORDER_PRODUCT_GRPC_ADDR
```

This is a misconfiguration/gap for live end-to-end checkout. Jab concrete Cart/Product gRPC adapters add honge, unke target addresses, timeouts, and TLS settings ko env/config me add karna chahiye.

Suggested future config names:

```env
ORDER_CART_GRPC_ADDR=localhost:9093
ORDER_PRODUCT_GRPC_ADDR=localhost:9090
ORDER_DOWNSTREAM_GRPC_TIMEOUT=1500ms
```

These are suggestions only; they are not currently loaded by the code.

## 8. Docker Setup

`${TASK_FILE_NAME}` does not introduce a new checked-in Dockerfile, compose file, container, volume, or network for `${SERVICE_NAME}`.

Reuse:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Section:
Docker and DevOps Setup
```

### What Changes For Task 3 Live Integration

For unit tests, no Docker services are needed.

For live checkout, local compose eventually needs:

| Container / Service | Purpose | Status |
|---|---|---|
| MySQL | Order DB | Reused from previous guide |
| Kafka | Event publishing if enabled | Reused from previous guide |
| Cart Service | `GetCart` dependency | Missing from current repo |
| Product Service | Product snapshots and inventory reservation | Adapter missing from `${SERVICE_CODE_PATH}` |
| MongoDB/Redis for Cart | Cart Service-owned infra | Indirect, not owned by `${SERVICE_NAME}` |
| MongoDB for Product | Product Service-owned infra | Indirect, not owned by `${SERVICE_NAME}` |

Do not put Cart/Product database credentials inside `${SERVICE_NAME}`. Service boundary ka rule hai: Order -> Cart/Product gRPC, not direct DB access.

## 9. Local Development Setup

### Step 1: Read Previous Dependency Documentation

Follow these first:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
${PREVIOUS_DEPENDENCY_FILE_TASK_2}
```

They explain Go setup, MySQL, Docker, Kafka, base migrations, and common setup issues.

### Step 2: Read Task 3 Guide

From repository root:

```bash
sed -n '1,220p' "TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}"
```

### Step 3: Go To Project Directory

```bash
cd backend/services/order-service
```

### Step 4: Install Dependencies

No new dependency is required for Task 3. Use existing Go module commands from `${PREVIOUS_DEPENDENCY_FILE_TASK_1}`:

```bash
go mod download
env GOWORK=off go test ./internal/usecase -run CreateOrderFromCart
```

### Step 5: Setup Databases/Services

For unit tests:

```text
No MySQL, Cart Service, Product Service, Redis, or Kafka needed.
```

For real runtime:

1. Start MySQL using previous guide.
2. Apply migrations in order.
3. Start Cart Service when available.
4. Start Product Service when available.
5. Start Kafka only if events are enabled.

### Step 6: Add Environment Variables

Use existing `.env` from `${PREVIOUS_DEPENDENCY_FILE_TASK_1}`.

Task 3-specific check:

```bash
export ORDER_INVENTORY_RESERVATION_TTL=15m
export ORDER_INVENTORY_RELEASE_TIMEOUT=3s
```

### Step 7: Run Migrations If Real DB Is Used

From repository root:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 < backend/services/order-service/migrations/001_create_order_tables.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/002_add_payment_coordination.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/003_add_grpc_query_indexes.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/004_create_order_outbox_events.up.sql
```

### Step 8: Verify Task 3 Functionality

Unit-level verification:

```bash
cd backend/services/order-service
env GOWORK=off go test ./internal/domain -run 'Cart|CreatedOrder|Checkout'
env GOWORK=off go test ./internal/usecase -run CreateOrderFromCart
```

Repository-level verification:

```bash
env GOWORK=off go test ./internal/repository -run CreateOrderAndAttachClaim
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
env GOWORK=off go test ./internal/usecase -run CreateOrderFromCart
env GOWORK=off go test ./...
env GOWORK=off go build ./...
```

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
| Cart Service gRPC | TBD | `GetCart` dependency | New external dependency; not configured in current code |
| Product Service gRPC | 9090 planned in docs | Product snapshot/inventory dependency | New external dependency; adapter missing in `${SERVICE_CODE_PATH}` |
| Redis | 6379 | Cart hot state indirectly | Not direct for `${SERVICE_NAME}` |
| MongoDB | 27017 | Cart/Product owned storage indirectly | Not direct for `${SERVICE_NAME}` |

Port conflict checks are already explained in:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Section:
Ports and Networking
```

Task 3 networking rule:

```text
Order Service must call Cart/Product APIs. It must not connect to Cart/Product databases.
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
```

Task-3-specific errors:

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `cart not found` | Cart Service returned nil/wrong cart id | Verify `cart_id` and user ownership | Always pass authenticated user id from gateway |
| `cart is empty` | Cart has no items | Add items before checkout | Frontend should block checkout for empty cart |
| `cart does not belong to buyer` | Cart user id mismatch | Use logged-in user's cart only | Never trust user-provided `user_id` |
| `duplicate cart item` | Same product/variant appears twice | Merge quantities in Cart Service | Cart Service should enforce unique variant rows |
| `product is unavailable` | Product missing/unpublished/invalid snapshot | Refresh product data or block checkout | Product Service should return authoritative availability |
| `variant unavailable` | Variant inactive | User must select active variant | Product Service should validate variant state |
| `inventory unavailable` | Stock cannot be reserved | Show out-of-stock message | Reserve inventory atomically in Product Service |
| `product currency does not match cart currency` | Cart currency and product snapshot differ | Rebuild cart totals in correct currency | Keep cart currency consistent |
| `invalid product price` | Product unit amount negative | Fix product pricing data | Product Service should validate prices before publish |
| `invalid reservation` | Product Service returned empty/expired reservation | Fix Product Service reservation response | Contract tests for reservation response |
| `order_persistence_failed` release path | MySQL write failed after inventory reserve | Check DB availability; reservation release should run | Keep release timeout configured |
| `Unknown column 'inventory_reservation_id'` | Migration `002` not applied | Apply `002_add_payment_coordination.up.sql` | Apply migrations in order |
| `Table 'order_outbox_events' doesn't exist` | Migration `004` not applied | Apply `004_create_order_outbox_events.up.sql` | Use full current runtime migration order |
| `ORDER_KAFKA_BROKERS is required` | Events default enabled but brokers empty | Set brokers or `ORDER_EVENTS_ENABLED=false` locally | Keep local `.env` explicit |
| Live checkout cannot reach Cart/Product | No concrete adapter/address config exists | Implement adapters and config variables | Add integration tests and health checks |
| `cannot load module ../auth-service listed in go.work file` | `backend/go.work` references a service without `go.mod` | Run module tests with `GOWORK=off` | Keep workspace entries aligned with actual modules |

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
```

Task-3-specific best practices:

- Cart price trust mat karo. Fresh `unit_amount` Product Service se lo.
- Product, variant, seller, currency, and stock validate karo before order create.
- Inventory reservation idempotency key deterministic child key se banao.
- DB save fail ho to reservation release karo.
- DB commit outcome unknown ho to reservation automatically release mat karo; reconciliation chahiye.
- Shipping address PII hai. Logs me full address print mat karo.
- `idempotency_key` raw logs me avoid karo; reference/hash use karo.
- Cart/Product gRPC calls me deadlines lagao. Checkout path hang nahi hona chahiye.
- Internal gRPC auth token 32+ chars rakho; production me secret manager/mTLS use karo.
- Direct cross-service DB access mat karo.
- Current-code live DB me `001` to `004` migrations order me apply karo.

## 14. Missing or Misconfigured Things

| Item | Status | Impact | Suggested Fix |
|---|---|---|---|
| `cmd/server/main.go` | Missing | Service ko real process ki tarah start nahi kar sakte | Add bootstrap wiring config, DB, repositories, usecases, gRPC server |
| Cart Service implementation | Missing from current service tree | Live `GetCart` dependency unavailable | Implement `backend/services/cart-service` or provide external running service |
| Product concrete adapter | Missing in `${SERVICE_CODE_PATH}` | Live product snapshot/reservation calls unavailable | Add gRPC adapter implementing `ProductClient` |
| Cart concrete adapter | Missing in `${SERVICE_CODE_PATH}` | Live cart fetch unavailable | Add gRPC adapter implementing `CartClient` |
| Downstream address env vars | Missing | No config for Cart/Product targets | Add validated env vars for Cart/Product gRPC addresses and timeouts |
| `.env.example` | Missing for `${SERVICE_CODE_PATH}` | Beginners may create incomplete env | Add sanitized example file |
| Docker Compose for full checkout | Missing | Hard to run MySQL/Kafka/Cart/Product together | Add compose once services exist |
| Health checks | Missing | Hard to know if dependencies are ready | Add gRPC health checks and DB/Kafka readiness |
| Current code needs later migrations | Present mismatch with Task 3 boundary | Real DB runtime fails if only migration `001` is applied | Document/apply full current runtime migrations |

## 15. References To Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Project Tech Stack Analysis | Same Go/gRPC/MySQL/Kafka/Docker stack |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Go Dependency System | Same module and dependency commands |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Environment Variables | Full existing `.env` documentation already present |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | External Services Analysis -> Kafka | Kafka setup reused when events are enabled |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Docker and DevOps Setup | MySQL/Kafka Docker setup reused |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Ports and Networking | MySQL, Kafka, and Order gRPC ports reused |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Common Errors and Fixes | Generic setup troubleshooting reused |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_2}` | Database Analysis | Base MySQL schema reused |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_2}` | Task 2 Schema Requirements | Order tables and idempotency table reused |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_2}` | Migration Order | Existing migration order reused and extended for current runtime note |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_2}` | Security and Configuration Notes | DB security guidance reused |

## Final Checklist

- [ ] `${INPUT_FILE_PATH}` reviewed
- [ ] Previous dependency docs checked before adding new documentation
- [ ] No duplicate Go/MySQL/Docker/Kafka installation guide added
- [ ] `${TASK_3_USECASE}` checked
- [ ] `${TASK_3_CONTRACTS}` checked
- [ ] `${TASK_3_DOMAIN}` checked
- [ ] Task 3 unit test command identified
- [ ] Cart Service dependency documented
- [ ] Product Service dependency documented
- [ ] Inventory reservation TTL/release timeout documented as reused env
- [ ] MySQL migration requirements documented
- [ ] Current-code outbox/migration mismatch documented
- [ ] Missing Cart/Product adapters documented
- [ ] Missing server entrypoint documented
- [ ] Common Task 3 setup/runtime errors documented
- [ ] Security and best practices documented
- [ ] No original task file overwritten

## Quick Beginner Summary

`${TASK_FILE_NAME}` ke liye naya local install mostly nahi chahiye. Go/MySQL/Docker/Kafka basics previous dependency files me already hain.

Fastest verification:

```bash
cd backend/services/order-service
env GOWORK=off go test ./internal/usecase -run CreateOrderFromCart
```

Live checkout ke liye extra cheezein chahiye: Cart Service, Product Service, concrete gRPC adapters, downstream address config, MySQL migrations, and a server entrypoint. Current repo me unit tests ready hain, but full end-to-end runtime abhi bootstrap/adapters ke bina complete nahi hai.
