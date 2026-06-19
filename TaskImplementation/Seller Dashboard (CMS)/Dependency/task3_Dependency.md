# Project Dependency & Setup Guide

## 1. Project Overview

### Variables used by this guide

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Seller Dashboard (CMS)` |
| `TASK_FILE_NAME` | `task3.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task3_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}` |

This guide is for the Order Manager work documented in `INPUT_FILE_PATH`.

Simple Hinglish goal: Task 3 me seller dashboard ke andar order list, order detail, fulfillment/shipment update, tracking info, aur read-only refunds view chalana hai. Ye guide business logic repeat nahi karta. Ye sirf setup, dependencies, environment, database, Docker, DevOps, and debugging requirements explain karta hai.

### What changed from Task 2

Task 1 and Task 2 already explain the common seller dashboard setup. Task 3 reuses that setup and adds the Order Service runtime boundary.

| New / Reused | Dependency | Why it matters for Task 3 |
|---|---|---|
| Reused | React, Vite, pnpm workspace | Order Manager same frontend app ke andar run hota hai. |
| Reused | React Query | Seller order list fetch, pagination cache, and fulfillment mutation invalidation ke liye. |
| Reused | React Hook Form, Zod, `@hookform/resolvers` | Shipment update form validation ke liye. These were already introduced in Task 2. |
| Reused | API Gateway on `8080` | Browser REST calls gateway ko hit karti hain. |
| Reused | Redis, if gateway rate limit enabled | Gateway rate limiting ke liye. |
| Reused | gRPC / Protobuf | Gateway internal `OrderService` methods ko call karega. |
| New for this task | Order Service boundary | Seller order list and fulfillment update APIs ka backend owner. |
| New for this task | MySQL `order_db` schema | Orders, order items, shipments, and order status history store karne ke liye. |
| Related but not directly called | Payment Service refund data | Refund creation/review seller dashboard ka scope nahi hai; seller UI only read-only refund info dikhata hai if order response contains it. |

### Runtime flow

```text
Seller browser
  -> Vite frontend app
  -> API Gateway REST API
  -> OrderService gRPC methods
  -> MySQL order_db
  -> Payment Service data only if backend enriches order response with refunds
```

Important boundary:

- Frontend directly MySQL, Redis, Payment Service, ya Order Service gRPC ko call nahi karta.
- Frontend sirf API Gateway ko REST calls bhejta hai.
- Backend must validate seller ownership. Frontend active seller context ko trust mat karo.
- Current repo me `backend/services/order-service` folder clearly present nahi hai, so real backend run tabhi possible hoga jab Order Service source add hoga.

## 2. Tech Stack

### Reused stack

Frontend, pnpm workspace, API Gateway, Redis, MySQL installation basics, Go Modules, gRPC, and Protobuf setup already documented hain.

Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Also refer:

```text
TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md
TaskImplementation/${SERVICE_NAME}/Dependency/API_Gateway.md
TaskImplementation/${SERVICE_NAME}/Dependency/Go_Modules.md
TaskImplementation/${SERVICE_NAME}/Dependency/gRPC.md
TaskImplementation/${SERVICE_NAME}/Dependency/Protobuf.md
TaskImplementation/${SERVICE_NAME}/Dependency/MySQL.md
TaskImplementation/${SERVICE_NAME}/Dependency/Redis.md
TaskImplementation/${SERVICE_NAME}/Dependency/Environment.md
```

### Task 3 specific stack

| Technology | Required? | Beginner-friendly explanation | Why Task 3 uses it |
|---|---:|---|---|
| Order Service | Yes for real data | Order Service order lifecycle ka backend owner hai. Ye order create, list, status, fulfillment, cancellation, and order history manage karta hai. | `GET /api/v1/seller/orders` and `PATCH /api/v1/seller/orders/{order_id}/fulfillment` isi service se aate hain. |
| MySQL `order_db` | Yes for real Order Service | MySQL relational database hai. Tables ke form me transactional data store hota hai. | Orders financial and auditable data hain, isliye relational DB required hai. |
| API Gateway `ORDER_GRPC_ADDR` | Yes | Gateway ko pata hona chahiye ki Order Service gRPC kis host/port par listen kar raha hai. | Seller order REST endpoints ko `OrderService` methods tak forward karne ke liye. |
| React Hook Form | Yes, already installed | Form state manage karta hai without unnecessary heavy re-render. | Shipment update form me status, carrier, tracking number manage karne ke liye. |
| Zod | Yes, already installed | Validation schema library hai. Simple words me, form data valid hai ya nahi check karta hai. | Carrier/tracking length and fulfillment status validate karne ke liye. |
| Seller permission helper | Yes | Frontend role permissions check karta hai. | `orders:view` and `orders:update_fulfillment` ke basis par list/detail/update access control. |

No new frontend package installation is required right now because Task 3 packages already exist in `frontend/seller-dashboard/package.json`.

## 3. Required Software

Common software is already documented in previous dependency files:

| Software | Follow this existing doc |
|---|---|
| Git | `task1_Dependency.md`, section `Required Software` |
| Node.js, Corepack, pnpm | `Dependency/Frontend.md` |
| Go | `Dependency/Go_Modules.md` |
| Docker / Docker Compose | `task1_Dependency.md`, sections `Required Software` and `Docker Setup` |
| MySQL | `Dependency/MySQL.md` |
| Redis | `Dependency/Redis.md` |
| curl | `task1_Dependency.md`, section `Required Software` |
| grpcurl, optional | `Dependency/gRPC.md`, verification notes |

Task 3 specific practical need:

| Software | Required? | Verify command | Notes |
|---|---:|---|---|
| MySQL client/server | Yes for real Order Service | `mysql --version` | Installation is reused, but `order_db` schema is task-specific. |
| Order Service binary/source | Yes for real APIs | Not available yet | Current repo does not clearly contain `backend/services/order-service`. |

Beginner note: Agar sirf frontend UI open karna hai, Node + pnpm enough hai. Agar real orders dekhne/update karne hain, API Gateway, Order Service, Auth/User seller session, MySQL `order_db`, and Redis if rate limiting is enabled chahiye.

## 4. Dependency Management

### Frontend dependencies

Frontend dependency setup is already explained in:

```text
TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md
```

Task 3 dependency status:

| Package | Used by | Current status |
|---|---|---|
| `@tanstack/react-query` | `useSellerOrders`, `useFulfillmentUpdate` | Already installed. |
| `react-hook-form` | `ShipmentUpdateForm` | Already installed. |
| `zod` | Shipment form validation schema | Already installed. |
| `@hookform/resolvers` | Zod resolver for React Hook Form | Already installed. |
| `react-router-dom` | `/seller/orders` and `/seller/orders/:orderId` routes | Already installed. |
| `lucide-react` | Back/save/filter icons | Already installed. |
| `zustand` | Active seller store | Already installed. |

Do not run `pnpm add` for Task 3 unless one of these packages is missing.

Useful checks after normal install:

```bash
cd frontend
pnpm --filter seller-dashboard typecheck
pnpm --filter seller-dashboard test
```

### Backend dependencies

Backend dependency setup for Go modules is already explained in:

```text
TaskImplementation/${SERVICE_NAME}/Dependency/Go_Modules.md
```

Current repo observation:

```text
backend/services/order-service
```

was not found. Also no Order Service `go.mod`, `cmd/server`, migrations folder, or `.env` file was found.

When Order Service source is added, expected Go setup will be:

```bash
cd backend
go work use ./services/order-service
go work sync

cd services/order-service
go mod tidy
go test ./...
go run ./cmd/server
```

Run these only after the Order Service module exists.

## 5. Database Setup

### Reused database technology: MySQL

MySQL installation, Docker container basics, credentials placement, and generic troubleshooting are already explained in:

```text
TaskImplementation/${SERVICE_NAME}/Dependency/MySQL.md
```

Task 3 does not need a second MySQL server. It needs a new MySQL database/schema on the same local MySQL setup.

### Task 3 database: `order_db`

#### What it is

`order_db` Order Service ka MySQL database hai. Simple Hinglish me: ye database seller ke orders, order items, shipment details, aur status history ko relational tables me store karta hai.

#### Why Task 3 uses it

Order Manager screen ke liye backend ko ye data chahiye:

- seller ke orders list karna
- specific seller ke order items filter karna
- fulfillment/shipment status update karna
- status history maintain karna
- shipment tracking number and carrier save karna

#### Required or optional

| Scenario | Required? |
|---|---:|
| Frontend UI with mocked data | No |
| Real `GET /api/v1/seller/orders` | Yes |
| Real fulfillment update | Yes |
| Refund panel populated from backend response | Depends on Payment Service integration |

#### Evidence in project

| Path | Evidence |
|---|---|
| `docs/04-microservice-design.md` | Order Service database choice is MySQL. |
| `database/draw.sql` | Creates `order_db` and order tables. |
| `api/master-api.json` | Seller order list and fulfillment routes map to `order-service`. |

#### Tables needed by Task 3

| Table | Why it matters |
|---|---|
| `orders` | Order header, status, totals, payment id, address snapshot. |
| `order_items` | Seller-specific items, product snapshot, quantity, fulfillment status. |
| `order_status_history` | Status timeline for audit/readability. |
| `shipments` | Carrier, tracking number, shipment status, shipped/delivered timestamps. |
| `order_idempotency_keys` | Checkout safety; not directly used by UI but part of Order Service schema. |

Related refund data lives in Payment Service tables, not Order Service tables. Seller UI should only display refund data if the Order Service/API response includes it.

#### Default port

```text
3306
```

#### Local schema bootstrap

No dedicated Order Service migration folder was found. Current available schema source is:

```text
database/draw.sql
```

For local bootstrap, if you are okay creating all project databases from this file:

```bash
mysql -u root -p < database/draw.sql
```

Verify `order_db`:

```bash
mysql -u root -p -e "SHOW DATABASES LIKE 'order_db';"
mysql -u root -p -e "SHOW TABLES FROM order_db;"
```

Expected Task 3 tables include:

```text
orders
order_items
order_status_history
shipments
order_idempotency_keys
```

#### Recommended Order Service connection string

Current repo does not include `backend/services/order-service/.env`. When that service is added, use a local DSN like:

```env
ORDER_MYSQL_DSN=order_user:change-me@tcp(localhost:3306)/order_db?parseTime=true&loc=UTC
```

Alternative split config pattern:

```env
ORDER_DB_HOST=localhost
ORDER_DB_PORT=3306
ORDER_DB_NAME=order_db
ORDER_DB_USER=order_user
ORDER_DB_PASSWORD=change-me
ORDER_DB_TIMEZONE=UTC
```

Credentials placement:

```text
backend/services/order-service/.env
```

Do not put MySQL credentials in frontend `VITE_*` variables.

#### Create a local DB user, if needed

If you do not want Order Service to use root credentials:

```sql
CREATE USER IF NOT EXISTS 'order_user'@'%' IDENTIFIED BY 'change-me';
GRANT SELECT, INSERT, UPDATE, DELETE ON order_db.* TO 'order_user'@'%';
FLUSH PRIVILEGES;
```

Production note: use stronger passwords, narrower hosts, TLS where possible, and migration-specific users for schema changes.

## 6. Redis / Queue / External Services

### Redis

Redis setup is reused from:

```text
TaskImplementation/${SERVICE_NAME}/Dependency/Redis.md
```

Task 3 does not add a direct Redis dependency in the frontend. Redis still matters if API Gateway rate limiting remains enabled:

```env
RATE_LIMIT_ENABLED=true
REDIS_ADDR=localhost:6379
```

### API Gateway and Order Service

API Gateway setup is reused from:

```text
TaskImplementation/${SERVICE_NAME}/Dependency/API_Gateway.md
```

Task 3 specifically needs this gateway variable:

```env
ORDER_GRPC_ADDR=localhost:50056
```

This value already exists in:

```text
backend/services/api-gateway/.env
```

Order Service, when implemented, should listen on `:50056` or the gateway value should be updated to match the real service address.

### Payment Service and refunds

Payment Service is related but not directly called by Task 3 frontend.

What Task 3 does:

- read-only refund panel
- display refunds if `Order` response includes `refunds`

What Task 3 does not do:

- create refunds
- approve/reject refunds
- call payment provider
- handle payment webhooks

So no new Stripe, Razorpay, SMTP, Kafka, RabbitMQ, or webhook setup is introduced by Task 3.

## 7. Environment Variables

Common frontend, gateway, MySQL, Redis, and CMS env behavior is already explained in:

```text
TaskImplementation/${SERVICE_NAME}/Dependency/Environment.md
```

### Frontend env

No new frontend env variable is introduced by Task 3.

Reused optional frontend env:

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_API_TIMEOUT_MS=15000
VITE_LOGIN_URL=/login
```

Where to place it:

```text
frontend/seller-dashboard/.env.local
```

Beginner note: Vite `VITE_*` values browser bundle me visible hote hain, so secrets yahan kabhi mat rakho.

### Gateway env

Task 3 required gateway variable:

```env
ORDER_GRPC_ADDR=localhost:50056
```

Where it exists:

```text
backend/services/api-gateway/.env
```

| Variable | Required? | Purpose | Example | Security notes |
|---|---:|---|---|---|
| `ORDER_GRPC_ADDR` | Yes for Task 3 real APIs | Gateway ko Order Service gRPC target batata hai. | `localhost:50056` | Not secret, but wrong value causes order APIs to fail. |

### Recommended future Order Service env

Current repo does not contain `backend/services/order-service/.env`. When Order Service is added, create:

```text
backend/services/order-service/.env
```

Suggested local example:

```env
SERVICE_NAME=order-service
APP_ENV=local
LOG_LEVEL=debug

# gRPC server
ORDER_GRPC_ADDR=:50056
ORDER_GRPC_MAX_RECV_MSG_BYTES=1048576
ORDER_GRPC_MAX_SEND_MSG_BYTES=1048576

# MySQL
ORDER_MYSQL_DSN=order_user:change-me@tcp(localhost:3306)/order_db?parseTime=true&loc=UTC
ORDER_DB_HOST=localhost
ORDER_DB_PORT=3306
ORDER_DB_NAME=order_db
ORDER_DB_USER=order_user
ORDER_DB_PASSWORD=change-me
ORDER_DB_TIMEZONE=UTC

# Order behavior
ORDER_DEFAULT_PAGE_SIZE=20
ORDER_MAX_PAGE_SIZE=100
ORDER_FULFILLMENT_ALLOWED_STATUSES=packed,shipped,delivered
```

New or task-specific variable explanation:

| Variable | Required? | Purpose | Example | Security notes |
|---|---:|---|---|---|
| `ORDER_GRPC_ADDR` | Yes | Order Service gRPC listen address. | `:50056` | Must align with gateway `ORDER_GRPC_ADDR=localhost:50056`. |
| `ORDER_MYSQL_DSN` | Yes, if DSN style used | Full MySQL connection string. | `order_user:change-me@tcp(localhost:3306)/order_db?parseTime=true&loc=UTC` | Contains password; do not commit real values. |
| `ORDER_DB_HOST` | Yes, if split config used | MySQL host. | `localhost` | Use Docker service name inside compose networks. |
| `ORDER_DB_PORT` | Yes, if split config used | MySQL port. | `3306` | Not secret. |
| `ORDER_DB_NAME` | Yes, if split config used | Database name. | `order_db` | Keep separate per environment. |
| `ORDER_DB_USER` | Yes, if split config used | DB username. | `order_user` | Use least privilege. |
| `ORDER_DB_PASSWORD` | Yes, if split config used | DB password. | `change-me` | Secret; never expose to frontend or commit production value. |
| `ORDER_DEFAULT_PAGE_SIZE` | Recommended | Default order list page size. | `20` | Aligns with frontend default `page_size=20`. |
| `ORDER_MAX_PAGE_SIZE` | Recommended | Prevents very large list queries. | `100` | Abuse guardrail. |

## 8. Docker Setup

Generic Docker setup is already covered in previous dependency files.

Task 3 does not require a new container type if MySQL is already running. It requires the `order_db` schema inside MySQL.

### Task 3 container/service status

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Frontend Vite | `5174` | Frontend dev server | Reused |
| Vite preview | `4174` | Built frontend preview | Reused |
| API Gateway | `8080` | REST edge API | Reused |
| Redis | `6379` | Gateway rate limit, if enabled | Reused |
| MySQL | `3306` | `order_db` plus other service DBs | Reused server, new schema for this task |
| Order Service gRPC | `50056` | Seller order APIs through gateway | New requirement, service source missing |
| Payment Service gRPC | `50057` | Refund/payment backend, only if backend enriches refunds | Related, not directly called by frontend |

### Reusing existing MySQL Docker

If the MySQL container from previous docs is already running, do not start another MySQL container on `3306`. Just apply/verify `order_db`.

```bash
docker ps
mysql -h 127.0.0.1 -P 3306 -u root -p -e "SHOW DATABASES LIKE 'order_db';"
```

### Suggested future compose addition

No project `docker-compose.yml` was found. If a compose file is added later, Order Service should be connected to MySQL and gateway like this:

```yaml
services:
  order-service:
    build:
      context: ./backend/services/order-service
    env_file:
      - ./backend/services/order-service/.env
    depends_on:
      mysql:
        condition: service_healthy
    expose:
      - "50056"

  api-gateway:
    environment:
      ORDER_GRPC_ADDR: order-service:50056
```

Docker networking note:

- Host-run service uses `ORDER_MYSQL_DSN=...localhost:3306...`.
- Compose-run service should use MySQL service name, for example `mysql:3306`.
- Gateway inside Compose should use `ORDER_GRPC_ADDR=order-service:50056`, not `localhost:50056`.

## 9. Local Development Setup

### Step 1: Read previous dependency documentation first

Follow these first because Task 3 reuses them:

```text
TaskImplementation/${SERVICE_NAME}/task1_Dependency.md
TaskImplementation/${SERVICE_NAME}/task2_Dependency.md
TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md
TaskImplementation/${SERVICE_NAME}/Dependency/API_Gateway.md
TaskImplementation/${SERVICE_NAME}/Dependency/MySQL.md
TaskImplementation/${SERVICE_NAME}/Dependency/Redis.md
TaskImplementation/${SERVICE_NAME}/Dependency/gRPC.md
TaskImplementation/${SERVICE_NAME}/Dependency/Environment.md
```

### Step 2: Go to project directory

```bash
cd <repo-root>
```

Documentation for this task lives in:

```text
TaskImplementation/${SERVICE_NAME}
```

### Step 3: Install only required frontend dependencies

No new packages are needed for Task 3. Use the normal frontend install:

```bash
cd frontend
corepack enable
pnpm install
```

### Step 4: Configure frontend API target

Create or update:

```text
frontend/seller-dashboard/.env.local
```

Only if defaults are not enough:

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_API_TIMEOUT_MS=15000
```

### Step 5: Setup MySQL `order_db`

If MySQL setup from previous docs is already done, only apply/verify the schema:

```bash
cd <repo-root>
mysql -u root -p < database/draw.sql
mysql -u root -p -e "SHOW TABLES FROM order_db;"
```

If `database/draw.sql` was already applied earlier, this step is safe because it uses `CREATE DATABASE IF NOT EXISTS` and `CREATE TABLE IF NOT EXISTS`.

### Step 6: Confirm gateway Order Service target

Check:

```text
backend/services/api-gateway/.env
```

Required value:

```env
ORDER_GRPC_ADDR=localhost:50056
```

### Step 7: Add Order Service env when service source exists

Current blocker: `backend/services/order-service` is missing. When added, create:

```text
backend/services/order-service/.env
```

Use the recommended env example from section 7 and align `ORDER_GRPC_ADDR`.

### Step 8: Run migrations

Dedicated Order Service migrations were not found.

Use current local bootstrap:

```bash
mysql -u root -p < database/draw.sql
```

Future expected command after migrations are added:

```bash
migrate -path backend/services/order-service/migrations -database "$ORDER_MYSQL_DSN" up
```

### Step 9: Start backend services

Current repo does not include runnable Order Service source. Expected future flow:

```bash
cd backend/services/order-service
go run ./cmd/server
```

Then start API Gateway after its source exists:

```bash
cd backend/services/api-gateway
go run ./cmd/server
```

Also ensure Auth/User seller session backend is available, because seller routes require authenticated seller context.

### Step 10: Start frontend

```bash
cd frontend
pnpm --filter seller-dashboard dev
```

Open:

```text
http://localhost:5174/seller/orders
```

## 10. Running the Project

### Verify frontend build and tests

```bash
cd frontend
pnpm --filter seller-dashboard typecheck
pnpm --filter seller-dashboard test
pnpm --filter seller-dashboard build
```

### Verify database

```bash
mysql -u root -p -e "SELECT COUNT(*) AS orders_count FROM order_db.orders;"
mysql -u root -p -e "SELECT COUNT(*) AS order_items_count FROM order_db.order_items;"
mysql -u root -p -e "SELECT COUNT(*) AS shipments_count FROM order_db.shipments;"
```

### Verify gateway/order APIs

These calls require valid seller authentication cookies/headers. Without auth, `401` or `403` is expected.

```bash
curl -i http://localhost:8080/api/v1/seller/orders
```

Fulfillment update example:

```bash
curl -i -X PATCH http://localhost:8080/api/v1/seller/orders/order_123/fulfillment \
  -H "content-type: application/json" \
  -d '{"order_id":"order_123","status":"packed","carrier":"Delhivery","tracking_number":"TRACK123"}'
```

Expected behavior:

- `GET /api/v1/seller/orders` returns seller-owned orders.
- `PATCH /api/v1/seller/orders/{order_id}/fulfillment` updates only allowed fulfillment statuses.
- Backend rejects unauthorized seller, wrong seller ownership, invalid status transition, and invalid payload.

### Ports and networking quick table

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Frontend dev server | `5174` | Browser UI | Reused |
| API Gateway | `8080` | REST API | Reused |
| MySQL | `3306` | `order_db` | Reused server, task-specific DB |
| Redis | `6379` | Gateway rate limit | Reused if enabled |
| Order Service gRPC | `50056` | Order backend target | Task 3 required, implementation missing |

Port conflict tips:

- If `5174` is busy, change `frontend/seller-dashboard/vite.config.ts`.
- If `8080` is busy, update gateway `HTTP_ADDR` and frontend `VITE_API_BASE_URL`.
- If `3306` is busy, update MySQL container mapping and `ORDER_MYSQL_DSN`.
- If `50056` is busy, update both Order Service listen address and gateway `ORDER_GRPC_ADDR`.

## 11. Common Errors & Fixes

Generic Node, pnpm, Redis, MySQL, Docker, and gateway errors are already documented in previous dependency files. Task 3 specific issues are below.

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `/seller/orders` shows permission denied | Active seller missing or current role lacks `orders:view`. | Login as seller, verify seller session, verify role permissions. | Seed/assign seller roles correctly before testing. |
| Fulfillment form not visible | Role lacks `orders:update_fulfillment` or order status is terminal. | Use `seller`, `seller_manager`, or `seller_order_manager`; test with `paid`, `packed`, or `shipped` order. | Keep frontend permission matrix aligned with backend RBAC. |
| Direct `/seller/orders/{id}` refresh shows unavailable state | Current API contract does not include seller-specific order detail fetch. UI relies on list/cache/location state. | Open order from `/seller/orders` list. Add backend `GET /api/v1/seller/orders/{id}` later if direct refresh is required. | Do not invent frontend API calls that are not in `api/master-api.json`. |
| `GET /api/v1/seller/orders` returns 404 | Gateway route not implemented or contract not loaded. | Check `api/master-api.json` and gateway route registration. | Add route tests for seller order endpoints. |
| `502`, `grpc unavailable`, or connection refused | Order Service not running or `ORDER_GRPC_ADDR` mismatch. | Start Order Service on `:50056` or update gateway env. | Keep service listen address and gateway target in one reviewed env example. |
| `Table 'order_db.orders' doesn't exist` | `database/draw.sql` not applied or wrong DB selected. | Run `mysql -u root -p < database/draw.sql`; verify tables. | Add dedicated migrations for Order Service. |
| Empty orders list but seed data exists | `order_items.seller_id` does not match active seller. | Check seller id in session and `order_items`. | Seed realistic seller-specific order items. |
| Fulfillment update rejected | Invalid status transition or backend status machine denies it. | Move through allowed statuses: `packed`, `shipped`, `delivered`. | Keep frontend `nextFulfillmentStatuses` in sync with backend rules. |
| Refund panel empty | Order API response does not include `refunds`. | Backend must enrich response from Payment Service if read-only refund display is required. | Keep refund creation/review in Payment/Admin boundary. |
| Tracking/carrier validation error | Input exceeds form validation length. | Use carrier/tracking values up to 80 characters. | Backend should enforce same or stricter validation. |
| Browser network error | Gateway not running or `VITE_API_BASE_URL` wrong. | Start gateway or update `frontend/seller-dashboard/.env.local`. | Keep `.env.local` documented and avoid random base URLs. |

## 12. Security & Best Practices

Task 3 specific best practices:

- Backend must enforce seller ownership for every order. Frontend seller id is UX context, not security proof.
- Backend must enforce allowed fulfillment transitions. Frontend validation is helpful but not authoritative.
- Do not add seller refund mutation endpoints in this task. Refund creation/review belongs to Payment/Admin flow.
- Do not expose MySQL credentials through frontend `VITE_*` variables.
- Avoid logging customer PII, shipping address, phone, email, payment id, or refund reason in frontend/backend logs.
- Use request ids from `http.ts` and backend logs to debug failed order operations safely.
- Keep order list pagination capped. Large unbounded order reads can slow MySQL.
- Index seller order queries. Existing schema has `idx_order_items_seller_created`; Order Service queries should use it.
- Use DB transactions for fulfillment updates so `shipments`, `order_items`, `orders`, and `order_status_history` stay consistent.
- Treat tracking number and carrier as user-entered text. Validate length and avoid rendering unsafe HTML.
- In production, internal gRPC should use TLS/mTLS or private service mesh networking.

## 13. Missing or Misconfigured Things

Current repo gaps found during Task 3 dependency analysis:

| Gap | Impact | Suggested fix |
|---|---|---|
| `backend/services/order-service` folder not found | Real seller order APIs cannot run from this repo yet. | Add Order Service source, `go.mod`, config loader, gRPC server, handlers, and tests. |
| `backend/services/order-service/.env` not found | No concrete Order Service runtime config exists. | Add sanitized `.env.example` and local `.env` with `ORDER_GRPC_ADDR` and DB config. |
| Dedicated Order Service migrations not found | Local setup depends on broad `database/draw.sql`. | Add `backend/services/order-service/migrations`. |
| No project `docker-compose.yml` found | Beginners must start MySQL/Redis/services manually. | Add compose file for local MySQL, Redis, gateway, and service dependencies. |
| Seller-specific order detail endpoint absent | Direct refresh of `/seller/orders/{orderId}` cannot fetch detail from backend. | Either keep current cache-based UX or add `GET /api/v1/seller/orders/{order_id}` to contract/backend. |
| Refund data source not defined in seller order response | Refund panel may remain empty. | Decide whether Order Service enriches refunds from Payment Service or exposes a safe read-only seller refund summary. |
| `.env.example` files not clearly found | New developers may copy real `.env` values by mistake. | Add sanitized examples for gateway, frontend, and future Order Service. |
| Health checks not clearly found for Order Service | DevOps cannot reliably verify readiness. | Add `/health/live`, `/health/ready`, or gRPC health service. |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | Required software, base frontend, gateway, Redis, MySQL basics | Task 3 runs inside same frontend and infra pattern. |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | React Hook Form, Zod, frontend form dependency pattern | Shipment update form uses the same form/validation stack. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md` | pnpm install, Vite dev/build/test commands | No new frontend package setup is needed. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/API_Gateway.md` | Gateway purpose, base port, downstream gRPC targets | Seller order REST APIs go through gateway. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/Go_Modules.md` | Go module workflow | Future Order Service should follow same backend module setup. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/gRPC.md` | gRPC explanation and troubleshooting | Gateway calls `OrderService` through gRPC. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/Protobuf.md` | Protobuf contract generation pattern | Order service methods should be contract-driven. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/MySQL.md` | MySQL install/Docker/troubleshooting | Task 3 uses MySQL too, but with `order_db`. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/Redis.md` | Redis install and rate limit setup | Redis remains gateway-level, not order-specific. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/Environment.md` | Existing frontend/gateway env patterns | Task 3 only adds/depends on order-specific env alignment. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/Migrations.md` | Migration command pattern | Future Order Service migrations should follow this style. |

## 15. Final Checklist

- [ ] Previous dependency documentation checked before starting Task 3 setup.
- [ ] Frontend dependencies installed from `frontend/pnpm-lock.yaml`.
- [ ] No duplicate frontend packages added for Task 3.
- [ ] `frontend/seller-dashboard/.env.local` points to the correct API Gateway, if default is not enough.
- [ ] MySQL server running on expected host/port.
- [ ] `order_db` exists.
- [ ] `orders`, `order_items`, `order_status_history`, and `shipments` tables exist.
- [ ] Gateway `ORDER_GRPC_ADDR` matches future Order Service listen address.
- [ ] Redis running if `RATE_LIMIT_ENABLED=true`.
- [ ] Order Service source/config added before attempting real API verification.
- [ ] Seller auth/session backend available.
- [ ] Seller role includes `orders:view` for list/detail.
- [ ] Seller role includes `orders:update_fulfillment` for shipment updates.
- [ ] `GET /api/v1/seller/orders` verified with authenticated seller context.
- [ ] `PATCH /api/v1/seller/orders/{order_id}/fulfillment` verified with allowed status transition.
- [ ] Direct detail refresh limitation understood.
- [ ] Refund panel treated as read-only.
- [ ] Logs checked without exposing secrets or customer PII.
- [ ] No duplicate setup documentation added.
