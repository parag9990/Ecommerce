# Dependency + Setup Documentation

## Document Variables

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `CMS Service` |
| `TASK_FILE_NAME` | `task1.md` |
| `SOURCE_TASK_PATH` | `TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task1_Dependency.md` |

This file explains the dependency, setup, environment, database, and DevOps requirements discovered from `SOURCE_TASK_PATH` and the current backend implementation under `backend/services/cms-service`.

Simple goal: beginner developer ko clear ho ki repo clone karne ke baad is service ko local machine par kaise run, configure, migrate, test, and debug karna hai.

## 1. What This Task Depends On

`TASK_FILE_NAME` defines seller permission roles and permission rules. Original task documentation-only foundation tha, but current service implementation now also contains Go code, HTTP routes, gRPC server, MySQL repositories, migrations, product moderation, coupon/campaign logic, seller analytics, seller settings, and audit logs.

| Area | Required? | Why |
|---|---:|---|
| Go runtime | Required | Service Go me likha gaya hai. |
| Go modules | Required | Dependencies `go.mod` and `go.sum` se manage hoti hain. |
| MySQL | Required | Service startup par MySQL connect karta hai and repositories initialize karta hai. |
| SQL migrations | Required | Required tables create karne ke liye migrations run karni hongi. |
| Product Service | Required for product moderation endpoints | Product status read/update/publish/unpublish ke liye internal HTTP calls hoti hain. |
| Auth RBAC / API Gateway | Required in full platform | Gateway JWT validate karke headers me user, seller, roles, staff status bhejta hai. |
| Internal auth token | Strongly recommended locally, required in production | Internal HTTP/gRPC endpoints ko protect karne ke liye. |
| Docker | Optional but recommended for beginners | MySQL ko clean local container me run karna easy hota hai. |
| Redis, Kafka, RabbitMQ, MongoDB | Not used by this service code right now | Repo docs me future/planned services ke liye mention hai, but current service startup me direct dependency nahi hai. |

## 2. Tech Stack Analysis

| Technology | What it is | Why used here | Required/Optional | Beginner explanation |
|---|---|---|---|---|
| Go | Backend programming language | Service, domain logic, repositories, HTTP/gRPC server | Required | Go ek compiled backend language hai. Fast APIs aur microservices banane ke liye use hoti hai. |
| Go modules | Go dependency system | External libraries ko version ke saath lock karta hai | Required | `go.mod` project ka dependency menu hai, aur `go.sum` checksum receipt hai. |
| `net/http` | Go standard HTTP library | `/healthz` and internal REST endpoints serve karta hai | Required | Ye Go ka built-in web server hai. Extra framework like Fiber/Gin yahan use nahi hua. |
| gRPC | High performance RPC framework | Internal service-to-service API expose karta hai | Required for gRPC interface | gRPC se services structured messages ke through baat karti hain. REST se zyada strict contract hota hai. |
| Protocol Buffers | gRPC message schema format | `proto/ecommerce/cms/v1/cms.proto` se generated Go code banta hai | Required for regeneration | Proto ek contract file hai jisme request/response ka shape define hota hai. |
| Buf | Protobuf lint/generation tool | `buf.yaml` and `buf.gen.yaml` present hain | Optional for normal run, required for proto regeneration | Buf proto files ko validate aur generated code create karne me help karta hai. |
| MySQL | Relational database | Seller staff, audit logs, coupons, campaigns, settings, analytics tables | Required | MySQL table based database hai. Structured data aur relations ke liye best fit hai. |
| `github.com/go-sql-driver/mysql` | MySQL driver for Go | Go `database/sql` ko MySQL se connect karata hai | Required | Driver bridge ki tarah kaam karta hai: Go code se MySQL queries run hoti hain. |
| `database/sql` | Go standard DB interface | Repositories SQL queries execute karte hain | Required | Ye Go ka generic DB layer hai. |
| `log/slog` | Go structured logging | JSON logs stdout par write hote hain | Required | Logs readable JSON me aate hain, debugging and monitoring easy hoti hai. |
| Product Service HTTP API | Internal product boundary | Product moderation me product details/status change ke liye call hota hai | Required for product moderation routes | Service product DB direct nahi padhta; Product Service se poochta hai. |
| Auth/API Gateway headers | Validated user context | User ID, seller ID, roles, staff status request headers se aate hain | Required for protected actions | Gateway pehle JWT validate karta hai, phir service ko trusted context deta hai. |
| Mermaid | Markdown diagram syntax in task docs | Original task file diagrams render karne ke liye | Optional | Local run ke liye install nahi chahiye. Docs preview ke liye useful hai. |

## 3. Go Dependency System

This is a Go project.

### Important files

| File | Purpose |
|---|---|
| `backend/services/cms-service/go.mod` | Module name, Go version, direct dependencies. |
| `backend/services/cms-service/go.sum` | Downloaded module checksums. Isko delete casually mat karo. |
| `backend/go.work` | Workspace file. Current workspace me `./services/cms-service` included hai. |

### Current direct dependencies

| Dependency | Why used |
|---|---|
| `github.com/go-sql-driver/mysql` | MySQL database connection ke liye. |
| `google.golang.org/grpc` | gRPC server and generated RPC code ke liye. |
| `google.golang.org/protobuf` | Protocol Buffer message types ke liye. |

### Beginner commands

From repo root:

```bash
cd backend/services/cms-service
go mod download
go mod tidy
go test ./...
go build ./cmd/server
```

Run the service from service folder:

```bash
cd backend/services/cms-service
set -a
source .env
set +a
go run ./cmd/server
```

Important note: code uses `os.Getenv`. It does not automatically load `.env`. Isliye `source .env` required hai on Linux/macOS shell.

### Common Go dependency issues

| Error | Cause | Fix |
|---|---|---|
| `go: errors parsing go.mod` | Installed Go version old hai. | `go version` check karo and `go.mod` ke version ke compatible Go install karo. |
| `missing go.sum entry` | Dependency checksum absent hai. | `go mod tidy` run karo. |
| `module lookup failed` | Network/proxy issue. | Internet check karo, `go env GOPROXY` check karo, then `go mod download`. |
| `package ... is not in std` | Wrong folder se command run hua. | `backend/services/cms-service` me jaakar run karo. |
| Generated proto import missing | Generated files stale/missing hain. | Buf/protoc setup karke proto regenerate karo. |

## 4. Database Analysis

### Database used

| Database | Required? | Default port | Why used |
|---|---:|---:|---|
| MySQL | Required | 3306 | Seller permissions, staff records, audit logs, product moderation reviews, coupons, campaigns, seller analytics, and seller settings structured tables me store hote hain. |

MySQL ek relational database hai. Simple words me: data rows and columns wali tables me store hota hai. Is service me permissions/audit/coupon/campaign jaise structured data ke liye MySQL mandatory hai.

### Tables expected by migrations

Run migrations in numeric order.

| Migration | Creates/changes |
|---|---|
| `001_create_cms_access_tables.up.sql` | `seller_staff`, `cms_audit_logs` |
| `002_create_product_moderation_reviews.up.sql` | `product_moderation_reviews` |
| `003_create_coupon_engine_tables.up.sql` | `coupons`, `coupon_rules`, `coupon_redemptions` |
| `004_create_offer_campaigns.up.sql` | `campaigns`, adds campaign fields/indexes to `coupon_redemptions` |
| `005_create_seller_analytics_tables.up.sql` | `seller_analytics_daily`, `seller_product_analytics_daily`, `seller_conversion_daily`, `seller_analytics_event_dedupe` |
| `006_create_seller_settings.up.sql` | `seller_settings` |
| `007_harden_cms_audit_logs.up.sql` | extra audit columns and indexes |

### Local MySQL installation

Windows:

```powershell
winget install Oracle.MySQL
mysql --version
```

Alternative: install MySQL Installer from the official MySQL site, then start "MySQL80" from Windows Services.

Linux Ubuntu/Debian:

```bash
sudo apt update
sudo apt install mysql-server mysql-client
sudo systemctl enable mysql
sudo systemctl start mysql
mysql --version
```

macOS:

```bash
brew install mysql
brew services start mysql
mysql --version
```

### MySQL Docker setup

Docker beginner-friendly option:

```bash
docker run --name cms-mysql \
  -e MYSQL_ROOT_PASSWORD=root_password \
  -e MYSQL_DATABASE=cms_db \
  -e MYSQL_USER=cms_user \
  -e MYSQL_PASSWORD=cms_password \
  -p 3306:3306 \
  -v cms_mysql_data:/var/lib/mysql \
  -d mysql:8.4
```

Verify:

```bash
docker ps
mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db
```

### Docker Compose example for MySQL only

No compose file exists in the repo for this service right now. If you want a local infra compose file, use this pattern:

```yaml
services:
  cms-mysql:
    image: mysql:8.4
    container_name: cms-mysql
    environment:
      MYSQL_ROOT_PASSWORD: root_password
      MYSQL_DATABASE: cms_db
      MYSQL_USER: cms_user
      MYSQL_PASSWORD: cms_password
      TZ: UTC
    ports:
      - "3306:3306"
    volumes:
      - cms_mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "127.0.0.1", "-u", "cms_user", "-pcms_password"]
      interval: 10s
      timeout: 5s
      retries: 10

volumes:
  cms_mysql_data:
```

Start/stop:

```bash
docker compose up -d
docker compose ps
docker compose logs cms-mysql
docker compose down
```

### Create database and user manually

If using local installed MySQL:

```sql
CREATE DATABASE IF NOT EXISTS cms_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'cms_user'@'%' IDENTIFIED BY 'cms_password';
GRANT ALL PRIVILEGES ON cms_db.* TO 'cms_user'@'%';
FLUSH PRIVILEGES;
```

Run with:

```bash
mysql -u root -p
```

### Run migrations

From service folder:

```bash
cd backend/services/cms-service
mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db < migrations/001_create_cms_access_tables.up.sql
mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db < migrations/002_create_product_moderation_reviews.up.sql
mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db < migrations/003_create_coupon_engine_tables.up.sql
mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db < migrations/004_create_offer_campaigns.up.sql
mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db < migrations/005_create_seller_analytics_tables.up.sql
mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db < migrations/006_create_seller_settings.up.sql
mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db < migrations/007_harden_cms_audit_logs.up.sql
```

Verify tables:

```bash
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SHOW TABLES;" cms_db
```

Expected result me tables like `seller_staff`, `cms_audit_logs`, `coupons`, `campaigns`, and `seller_settings` dikhne chahiye.

### Connection string formats

Component variables:

```env
CMS_DB_HOST=localhost
CMS_DB_PORT=3306
CMS_DB_NAME=cms_db
CMS_DB_USER=cms_user
CMS_DB_PASSWORD=cms_password
```

Direct DSN alternative:

```env
CMS_MYSQL_DSN=cms_user:cms_password@tcp(localhost:3306)/cms_db?parseTime=true&loc=UTC&charset=utf8mb4
```

If `CMS_MYSQL_DSN` set hai, service component variables ko ignore karke direct DSN use karega.

## 5. Environment Variables

Create local env file here:

```text
backend/services/cms-service/.env
```

Security note: `.env` file me secrets hote hain. `.gitignore` already `.env` files ignore karta hai. Real passwords, tokens, API keys commit mat karo.

### Complete `.env` example

```env
CMS_HTTP_ADDR=:8087
CMS_HTTP_READ_TIMEOUT=5s
CMS_HTTP_WRITE_TIMEOUT=10s
CMS_HTTP_IDLE_TIMEOUT=60s
CMS_SHUTDOWN_TIMEOUT=10s
CMS_MAX_BODY_BYTES=1048576

CMS_GRPC_ADDR=:9098
CMS_GRPC_MAX_RECV_MSG_BYTES=1048576
CMS_GRPC_MAX_SEND_MSG_BYTES=1048576
CMS_GRPC_ALLOWED_INTERNAL_CALLERS=cart-service,order-service,api-gateway

CMS_ENV=local
CMS_INTERNAL_AUTH_HEADER=X-Internal-Token
CMS_INTERNAL_AUTH_TOKEN=local-cms-internal-token

CMS_STAFF_STATUS_SOURCE=gateway

CMS_MYSQL_DSN=
CMS_DB_HOST=localhost
CMS_DB_PORT=3306
CMS_DB_NAME=cms_db
CMS_DB_USER=cms_user
CMS_DB_PASSWORD=cms_password
CMS_DB_MAX_OPEN_CONNS=25
CMS_DB_MAX_IDLE_CONNS=10
CMS_DB_CONN_MAX_LIFETIME_SECONDS=300
CMS_DB_TIMEZONE=UTC

CMS_PRODUCT_SERVICE_BASE_URL=http://localhost:8080
CMS_PRODUCT_SERVICE_TIMEOUT=5s
CMS_PRODUCT_SERVICE_AUTH_HEADER=X-Internal-Token
CMS_PRODUCT_SERVICE_AUTH_TOKEN=local-product-internal-token

CMS_CAMPAIGN_MAX_DURATION_DAYS=90

CMS_ANALYTICS_DEFAULT_CURRENCY=INR
CMS_ANALYTICS_DEFAULT_RANGE_DAYS=30
CMS_ANALYTICS_MAX_RANGE_DAYS=366
CMS_ANALYTICS_DEFAULT_TOP_PRODUCTS_LIMIT=5
CMS_ANALYTICS_MAX_TOP_PRODUCTS_LIMIT=20

CMS_AUDIT_DEFAULT_RANGE_DAYS=30
CMS_AUDIT_DEFAULT_PAGE_SIZE=20
CMS_AUDIT_MAX_PAGE_SIZE=100
```

### Environment variable reference

| Variable | Required? | Purpose | Example | Security/config note |
|---|---:|---|---|---|
| `CMS_HTTP_ADDR` | Required by config default | HTTP bind address | `:8087` | Change if port conflict. |
| `CMS_HTTP_READ_TIMEOUT` | Optional | Max read duration | `5s` | Keep positive. |
| `CMS_HTTP_WRITE_TIMEOUT` | Optional | Max response write duration | `10s` | Keep positive. |
| `CMS_HTTP_IDLE_TIMEOUT` | Optional | Idle connection timeout | `60s` | Keep positive. |
| `CMS_SHUTDOWN_TIMEOUT` | Optional | Graceful shutdown timeout | `10s` | Keep positive. |
| `CMS_MAX_BODY_BYTES` | Optional | HTTP request body limit | `1048576` | Prevents huge payload abuse. |
| `CMS_GRPC_ADDR` | Required by config default | gRPC bind address | `:9098` | Change if port conflict. |
| `CMS_GRPC_MAX_RECV_MSG_BYTES` | Optional | gRPC request size limit | `1048576` | Keep sane to avoid memory pressure. |
| `CMS_GRPC_MAX_SEND_MSG_BYTES` | Optional | gRPC response size limit | `1048576` | Keep sane. |
| `CMS_GRPC_ALLOWED_INTERNAL_CALLERS` | Optional | Allowed internal caller names | `cart-service,order-service,api-gateway` | Used by gRPC auth metadata flow. |
| `CMS_ENV` | Optional locally, important in production | Runtime environment | `local`, `production` | Production enables stricter validation. |
| `CMS_INTERNAL_AUTH_HEADER` | Required | Header name for internal auth | `X-Internal-Token` | Must match Gateway/internal clients. |
| `CMS_INTERNAL_AUTH_TOKEN` | Required in production, strongly recommended locally | Internal endpoint shared token | `local-cms-internal-token` | Never use `change-me` in real env. |
| `CMS_STAFF_STATUS_SOURCE` | Required by config default | Role source mode | `gateway` or `mysql` | `mysql` is safer when staff status table is populated. |
| `CMS_MYSQL_DSN` | Optional | Direct MySQL DSN | `cms_user:...@tcp(...)` | If set, component DB vars are ignored. |
| `CMS_DB_HOST` | Required if DSN empty | MySQL host | `localhost` | Use service/container name inside Docker network. |
| `CMS_DB_PORT` | Required if DSN empty | MySQL port | `3306` | Must be 1-65535. |
| `CMS_DB_NAME` | Required if DSN empty | Database name | `cms_db` | Must exist before service starts. |
| `CMS_DB_USER` | Required if DSN empty | Database username | `cms_user` | Grant permissions only to this DB. |
| `CMS_DB_PASSWORD` | Required in production if DSN empty | Database password | `cms_password` | Never commit real value. |
| `CMS_DB_MAX_OPEN_CONNS` | Optional | DB connection pool max open | `25` | Must be greater than zero. |
| `CMS_DB_MAX_IDLE_CONNS` | Optional | DB connection pool idle count | `10` | Cannot exceed open conns. |
| `CMS_DB_CONN_MAX_LIFETIME_SECONDS` | Optional | Connection recycle time | `300` | Helps avoid stale DB connections. |
| `CMS_DB_TIMEZONE` | Optional | MySQL time location | `UTC` | Invalid timezone breaks DSN build. |
| `CMS_PRODUCT_SERVICE_BASE_URL` | Required for current startup | Product Service base URL | `http://localhost:8080` | Must be valid `http` or `https`. |
| `CMS_PRODUCT_SERVICE_TIMEOUT` | Optional | Product Service call timeout | `5s` | Keep positive. |
| `CMS_PRODUCT_SERVICE_AUTH_HEADER` | Required when product client used | Product internal auth header | `X-Internal-Token` | Must match Product Service. |
| `CMS_PRODUCT_SERVICE_AUTH_TOKEN` | Strongly recommended | Product internal auth token | `local-product-internal-token` | Current validation does not require it, but real internal calls should use it. |
| `CMS_CAMPAIGN_MAX_DURATION_DAYS` | Optional | Campaign max length guardrail | `90` | Negative value invalid. |
| `CMS_ANALYTICS_DEFAULT_CURRENCY` | Optional | Default analytics currency | `INR` | Must be 3 uppercase letters. |
| `CMS_ANALYTICS_DEFAULT_RANGE_DAYS` | Optional | Default analytics window | `30` | Must be positive. |
| `CMS_ANALYTICS_MAX_RANGE_DAYS` | Optional | Max analytics window | `366` | Must be 1-366. |
| `CMS_ANALYTICS_DEFAULT_TOP_PRODUCTS_LIMIT` | Optional | Default top product count | `5` | Must be positive. |
| `CMS_ANALYTICS_MAX_TOP_PRODUCTS_LIMIT` | Optional | Max top product count | `20` | Must be 1-20. |
| `CMS_AUDIT_DEFAULT_RANGE_DAYS` | Optional | Default audit log range | `30` | Must be positive. |
| `CMS_AUDIT_DEFAULT_PAGE_SIZE` | Optional | Default audit page size | `20` | Must be positive. |
| `CMS_AUDIT_MAX_PAGE_SIZE` | Optional | Max audit page size | `100` | Must be 1-100. |

### How to load `.env`

Linux/macOS:

```bash
cd backend/services/cms-service
set -a
source .env
set +a
go run ./cmd/server
```

Windows PowerShell:

```powershell
cd backend/services/cms-service
Get-Content .env | ForEach-Object {
  if ($_ -match "^\s*([^#][^=]+)=(.*)$") {
    [Environment]::SetEnvironmentVariable($matches[1].Trim(), $matches[2].Trim(), "Process")
  }
}
go run ./cmd/server
```

Common `.env` mistakes:

| Mistake | Result | Fix |
|---|---|---|
| File created in repo root only | Service does not auto-load it | Source correct file before `go run`. |
| `CMS_INTERNAL_AUTH_TOKEN=change-me` | Weak security | Use a long random token. |
| `CMS_DB_HOST=localhost` inside service container | Container points to itself | Use compose service name like `cms-mysql`. |
| Quotes copied into values | Header/token mismatch | Prefer raw values unless shell requires quotes. |
| `CMS_ANALYTICS_DEFAULT_CURRENCY=inr` | Config validation fails | Use `INR`. |

## 6. External Services Analysis

### Product Service

What it is: Product Service product catalog ka owner hai. This service product data direct DB se nahi padhta.

Why used: Product moderation flows call Product Service for:

- `GET /internal/v1/products/{product_id}`
- `PATCH /internal/v1/products/{product_id}/status`
- `POST /internal/v1/products/{product_id}/publish`
- `POST /internal/v1/products/{product_id}/unpublish`

Required or optional:

- Required for product moderation endpoints.
- Service startup requires a valid base URL because product client initialization needs it.
- Health endpoint and some non-product routes can still work if Product Service is not reachable, but product moderation calls fail.

Config:

```env
CMS_PRODUCT_SERVICE_BASE_URL=http://localhost:8080
CMS_PRODUCT_SERVICE_TIMEOUT=5s
CMS_PRODUCT_SERVICE_AUTH_HEADER=X-Internal-Token
CMS_PRODUCT_SERVICE_AUTH_TOKEN=local-product-internal-token
```

Health check idea:

```bash
curl http://localhost:8080/healthz
```

Common issues:

| Issue | Cause | Fix |
|---|---|---|
| `product service unavailable` | Product Service not running or URL wrong | Start Product Service or update `CMS_PRODUCT_SERVICE_BASE_URL`. |
| `401` from Product Service | Internal token mismatch | Match product token/header with Product Service env. |
| Timeout | Product Service slow/down | Check logs and `CMS_PRODUCT_SERVICE_TIMEOUT`. |

### Auth RBAC / API Gateway

What it is: Auth/API Gateway JWT validate karta hai and service ko trusted headers deta hai.

Why used: This service expects identity in headers:

| Header | Purpose |
|---|---|
| `X-User-ID` | Actor user id |
| `X-Session-ID` | Session id |
| `X-Roles` | Comma separated roles |
| `X-Seller-ID` | Seller workspace id |
| `X-Tenant-ID` | Tenant id |
| `X-Staff-Status` | `active`, `invited`, `disabled`, `revoked` |
| `X-Request-ID` | Request tracing/audit |
| `X-Trace-ID` | Alternative trace id |
| `X-Internal-Token` | Internal auth shared token |

Required or optional:

- Required in production architecture.
- For local manual curl, you can send these headers yourself.

Security note: Service does not validate JWT directly in current code. Gateway/Auth validation must happen before requests reach it.

### gRPC clients

What it is: Internal RPC clients like cart/order/gateway can call this service.

Current gRPC methods:

| Method | Purpose |
|---|---|
| `ValidateCoupon` | Coupon validation for cart/order checkout. |
| `GetSellerSettings` | Read seller settings. |
| `GetCampaign` | Read one campaign. |
| `ListCampaigns` | List seller campaigns. |
| `ListAuditLogs` | Read audit logs. |

Setup:

- Generated files already exist in `backend/services/cms-service/internal/gen`.
- If proto changes, regenerate using Buf/protobuf toolchain.

### Services not required right now

| Service | Current status |
|---|---|
| Redis | Not directly used by current service code. |
| Kafka/RabbitMQ | Not directly used by current service code. |
| MongoDB | Not directly used by current service code. |
| MinIO/S3 | Not used. |
| Elasticsearch/Typesense | Not used. |
| SMTP/Twilio/Stripe/Firebase | Not used. |
| Kubernetes | Not present for this service. |

## 7. Ports and Networking

| Service | Port | Purpose | Config |
|---|---:|---|---|
| Service HTTP API | 8087 | Health and internal/seller HTTP endpoints | `CMS_HTTP_ADDR=:8087` |
| Service gRPC API | 9098 | Internal RPC calls | `CMS_GRPC_ADDR=:9098` |
| MySQL | 3306 | Database | `CMS_DB_PORT=3306` |
| Product Service | 8080 | Product moderation internal calls | `CMS_PRODUCT_SERVICE_BASE_URL=http://localhost:8080` |

Port conflict checks:

```bash
lsof -i :8087
lsof -i :9098
lsof -i :3306
```

If conflict happens:

```env
CMS_HTTP_ADDR=:18087
CMS_GRPC_ADDR=:19098
CMS_DB_PORT=3307
```

Docker networking note: host machine se MySQL `localhost:3306` hota hai. Service container se MySQL `cms-mysql:3306` hoga if same compose network me hai.

Firewall note: local firewall/VPN sometimes ports block kar sakta hai. First test with `curl` and `mysql` CLI before blaming code.

## 8. Docker and DevOps Setup

Current repo state for this service:

| Item | Present? | Notes |
|---|---:|---|
| Service Dockerfile | No | Not found under service folder. |
| docker-compose.yml | No | Repo does not currently provide compose stack. |
| MySQL migrations | Yes | SQL files under `migrations/`. |
| Health endpoint | Yes | `GET /healthz`. |
| Kubernetes manifests | No | Not present for current service. |

### Recommended beginner approach

Use Docker for MySQL, run Go service directly from host. Ye setup simplest hai because:

- MySQL clean container me run hota hai.
- Go debugger/local logs easy milte hain.
- Dockerfile missing hone se service container build karne ki zarurat nahi.

Commands:

```bash
docker run --name cms-mysql \
  -e MYSQL_ROOT_PASSWORD=root_password \
  -e MYSQL_DATABASE=cms_db \
  -e MYSQL_USER=cms_user \
  -e MYSQL_PASSWORD=cms_password \
  -p 3306:3306 \
  -v cms_mysql_data:/var/lib/mysql \
  -d mysql:8.4

docker ps
docker logs cms-mysql
```

Useful Docker commands:

```bash
docker ps
docker logs cms-mysql
docker stop cms-mysql
docker start cms-mysql
docker rm cms-mysql
docker volume ls
```

Warning: `docker rm` container remove karta hai. Data volume stays unless you remove `cms_mysql_data`.

## 9. Complete Project Run Instructions

### Step 1: Clone repository

```bash
git clone <repo-url>
cd Ecommerce
```

### Step 2: Check required tools

```bash
go version
mysql --version
docker --version
curl --version
```

Docker optional hai if local MySQL already installed.

### Step 3: Start MySQL

Docker option:

```bash
docker start cms-mysql
```

If container does not exist yet, create it:

```bash
docker run --name cms-mysql \
  -e MYSQL_ROOT_PASSWORD=root_password \
  -e MYSQL_DATABASE=cms_db \
  -e MYSQL_USER=cms_user \
  -e MYSQL_PASSWORD=cms_password \
  -p 3306:3306 \
  -v cms_mysql_data:/var/lib/mysql \
  -d mysql:8.4
```

### Step 4: Verify database connection

```bash
mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db
```

Password example: `cms_password`.

### Step 5: Run migrations

```bash
cd backend/services/cms-service
mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db < migrations/001_create_cms_access_tables.up.sql
mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db < migrations/002_create_product_moderation_reviews.up.sql
mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db < migrations/003_create_coupon_engine_tables.up.sql
mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db < migrations/004_create_offer_campaigns.up.sql
mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db < migrations/005_create_seller_analytics_tables.up.sql
mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db < migrations/006_create_seller_settings.up.sql
mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db < migrations/007_harden_cms_audit_logs.up.sql
```

### Step 6: Create `.env`

Create or update:

```text
backend/services/cms-service/.env
```

Use the complete `.env` example from section 5. Make sure DB credentials match your MySQL setup.

### Step 7: Install/download Go dependencies

```bash
cd backend/services/cms-service
go mod download
go mod tidy
```

### Step 8: Run tests

```bash
go test ./...
```

### Step 9: Start service

```bash
cd backend/services/cms-service
set -a
source .env
set +a
go run ./cmd/server
```

Expected logs:

```text
cms.mysql.connected
cms.http.started
cms.grpc.started
```

### Step 10: Verify health

In another terminal:

```bash
curl http://localhost:8087/healthz
```

Expected:

```json
{"success":true}
```

### Step 11: Verify internal authorization endpoint

```bash
curl -X POST http://localhost:8087/internal/v1/cms/authorize \
  -H "Content-Type: application/json" \
  -H "X-Internal-Token: local-cms-internal-token" \
  -H "X-User-ID: user_1" \
  -H "X-Seller-ID: seller_1" \
  -H "X-Roles: seller" \
  -H "X-Staff-Status: active" \
  -H "X-Request-ID: req_local_1" \
  -d '{
    "resource_seller_id": "seller_1",
    "required_permission": "cms:settings:update",
    "resource_type": "settings",
    "resource_id": "seller_1"
  }'
```

Expected result should show `allowed` as true for seller role and same seller scope.

## 10. HTTP and gRPC Endpoint Setup Notes

### HTTP routes

| Route | Auth | Purpose |
|---|---|---|
| `GET /healthz` | No internal token required | Basic health check. |
| `POST /internal/v1/cms/authorize` | Internal token | Permission decision API. |
| `GET /internal/v1/cms/permissions/catalog` | Internal token | Permission catalog. |
| `GET /api/v1/seller/dashboard/summary` | Internal token | Seller analytics summary. |
| `GET /api/v1/seller/audit-logs` | Internal token | Seller audit logs. |
| `POST /internal/v1/cms/seller/products/{product_id}/submit-review` | Internal token | Submit product review. |
| `POST /internal/v1/cms/seller/products/{product_id}/publish` | Internal token | Publish approved product. |
| `POST /internal/v1/cms/seller/products/{product_id}/unpublish` | Internal token | Seller product unpublish. |
| `GET /internal/v1/cms/admin/catalog/reviews` | Internal token | Admin moderation list. |
| `POST /internal/v1/cms/admin/catalog/reviews/{review_id}/decision` | Internal token | Admin approve/reject. |
| `POST /internal/v1/cms/admin/catalog/products/{product_id}/unpublish` | Internal token | Admin force unpublish. |
| `GET /internal/v1/cms/seller/coupons` | Internal token | List coupons. |
| `POST /internal/v1/cms/seller/coupons` | Internal token | Create coupon. |
| `PATCH /internal/v1/cms/seller/coupons/{coupon_id}` | Internal token | Update coupon. |
| `POST /internal/v1/cms/seller/coupons/{coupon_id}/disable` | Internal token | Disable coupon. |
| `POST /internal/v1/cms/coupons/validate` | Internal token | Validate coupon. |
| `POST /internal/v1/cms/coupons/redemptions` | Internal token | Record redemption. |
| `GET /internal/v1/cms/seller/campaigns` | Internal token | List campaigns. |
| `POST /internal/v1/cms/seller/campaigns` | Internal token | Create campaign. |
| `PATCH /internal/v1/cms/seller/campaigns/{campaign_id}` | Internal token | Update campaign. |
| `POST /internal/v1/cms/seller/campaigns/{campaign_id}/disable` | Internal token | Disable campaign. |
| `POST /internal/v1/cms/campaigns/validate` | Internal token | Validate campaign. |

### gRPC setup

Proto source:

```text
proto/ecommerce/cms/v1/cms.proto
```

Generated Go output:

```text
backend/services/cms-service/internal/gen/ecommerce/cms/v1
```

If proto changes, install Buf and Go protobuf plugins, then run from repo root:

```bash
buf generate
```

If plugin missing:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Make sure `$GOPATH/bin` is in `PATH`.

## 11. Common Errors and Fixes

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `cms.config.load_failed` | Invalid/missing env value | Check exact error in log and update `.env` | Keep `.env.example` with valid defaults. |
| `CMS_INTERNAL_AUTH_TOKEN cannot be empty in production` | `CMS_ENV=production` but token empty | Set strong token | Never run production without internal auth. |
| `ping mysql: connection refused` | MySQL down/wrong host/port | Start MySQL, verify `CMS_DB_HOST` and `CMS_DB_PORT` | Run `mysql` CLI before service. |
| `Access denied for user` | Wrong MySQL user/password | Update `CMS_DB_USER`/`CMS_DB_PASSWORD` or grant privileges | Use one known local credential set. |
| `Unknown database 'cms_db'` | DB not created | Create DB or set correct `CMS_DB_NAME` | Create DB before migrations. |
| `Table ... doesn't exist` | Migrations not run | Run all migrations in order | Add migration runner later. |
| Migration fails on duplicate column | Migration run twice, especially alter migration | Use migration tool tracking versions or inspect schema before rerun | Do not manually rerun alter migrations blindly. |
| `Error 1215 cannot add foreign key` | Earlier migration missing/table engine mismatch | Run migrations in order, use InnoDB | Keep DB charset/engine consistent. |
| `address already in use` | Port 8087/9098 busy | Change `CMS_HTTP_ADDR` or `CMS_GRPC_ADDR` | Check ports before starting. |
| `INTERNAL_AUTH_REQUIRED` | Missing/wrong internal auth header | Send `X-Internal-Token` matching env | Keep tokens consistent across services. |
| `unauthenticated` | Missing `X-User-ID` or seller context | Send required identity headers | Gateway should inject headers. |
| `cross_seller_access` | Actor seller and resource seller mismatch | Use same seller id or deny request | Never bypass same-seller rule. |
| `inactive_staff` | `X-Staff-Status` not `active`, or MySQL staff inactive | Set active status for local test | Use correct staff lifecycle. |
| `unknown_permission` | Permission string typo | Use catalog endpoint or constants | Do not free-type permissions. |
| `permission_denied` | Role lacks permission | Use correct role or update policy intentionally | Review role matrix. |
| Product moderation fails | Product Service not running/token mismatch | Start Product Service or fix URL/token | Health check dependencies. |
| `go mod download failed` | Network/proxy issue | Check internet, `GOPROXY`, retry | Cache deps in CI. |
| `go version` mismatch | Installed Go incompatible with `go.mod` | Install matching/newer Go | Pin Go toolchain in docs/CI. |
| Docker daemon not running | Docker Desktop/service stopped | Start Docker | Verify `docker ps`. |
| Permission denied on scripts/ports | OS permission issue | Use non-privileged ports, correct file perms | Avoid ports below 1024. |

## 12. Security and Configuration Audit

| Finding | Risk | Recommendation |
|---|---|---|
| Local `.env` contains placeholder tokens like `change-me` | Weak auth if copied to shared/prod env | Replace with long random secrets. Add/update `.env.example` separately. |
| Internal auth is bypassed when `CMS_INTERNAL_AUTH_TOKEN` is empty | Local endpoints become open | Always set token locally too, not just production. |
| Service trusts Gateway headers for user/role/seller context in `gateway` mode | Direct access can spoof headers if network exposed | Keep service private, use API Gateway, set internal auth, prefer mTLS/network policies in production. |
| `CMS_STAFF_STATUS_SOURCE=gateway` trusts `X-Staff-Status` | Revoked staff may pass if gateway context stale/wrong | Use `CMS_STAFF_STATUS_SOURCE=mysql` where seller staff table is reliable. |
| Product Service auth token is not required by validation | Product calls may be unauthenticated if token empty | Require token in production and align Product Service checks. |
| No Dockerfile or compose stack for this service | Onboarding and deployment less repeatable | Add service Dockerfile and local compose with MySQL healthcheck. |
| No migration runner/version table | Manual migration reruns can fail | Use a migration tool like `golang-migrate` or a controlled migration runner. |
| MySQL password optional in local | Easy local setup but unsafe if reused | Keep local-only values separate from production. |
| Health endpoint only checks process, not DB/Product dependencies | Health may pass while dependencies fail | Add readiness endpoint for DB and product dependency. |
| Ports bind to all interfaces when address is `:8087`/`:9098` | Service may be reachable beyond localhost | For local-only use `127.0.0.1:8087` and `127.0.0.1:9098`. |
| No JWT verification inside service | Security depends on Gateway boundary | Keep CMS private; do not expose directly to public internet. |

Professional fix priority:

1. Set real internal tokens and keep `.env` local-only.
2. Add `.env.example` with placeholders, not secrets.
3. Add migration runner or documented migration tracking.
4. Add Dockerfile and docker-compose local infra.
5. Add readiness check for DB/Product Service.
6. Prefer `mysql` staff status source for stronger authorization.

## 13. Best Practices

- Never commit `.env` with real secrets.
- Use `.env.example` for safe placeholders.
- Use strong random values for `CMS_INTERNAL_AUTH_TOKEN` and `CMS_PRODUCT_SERVICE_AUTH_TOKEN`.
- Keep MySQL data in Docker volume, not inside container filesystem only.
- Run migrations in order and track which ones already ran.
- Keep DB backups before destructive/down migrations.
- Use separate local, staging, and production configs.
- Keep service behind API Gateway/internal network.
- Do not trust client-supplied role headers directly from public traffic.
- Use `CMS_STAFF_STATUS_SOURCE=mysql` when staff table is populated and authoritative.
- Keep dependencies updated with `go mod tidy`, but review version changes.
- Run `go test ./...` before merging.
- Log request IDs and use them during debugging.
- Keep Product Service URL and internal token aligned across services.
- Document every new env variable when code adds one.

## 14. Quick Beginner Checklist

- [ ] Repository cloned.
- [ ] Go installed and compatible with `go.mod`.
- [ ] MySQL installed or Docker MySQL container running.
- [ ] `cms_db` database created.
- [ ] `cms_user` has privileges on `cms_db`.
- [ ] All CMS migrations run in order.
- [ ] `backend/services/cms-service/.env` created.
- [ ] `.env` has real local values, not `change-me`.
- [ ] Environment variables loaded before `go run`.
- [ ] Go dependencies downloaded.
- [ ] `go test ./...` passes.
- [ ] Service starts without config or MySQL errors.
- [ ] `GET /healthz` returns success.
- [ ] Internal auth token works with curl.
- [ ] Product Service URL configured for moderation routes.
- [ ] Logs checked for `cms.http.started` and `cms.grpc.started`.
- [ ] Common errors section reviewed.

## 15. Minimal Local Command Flow

Use this when you want the shortest working path:

```bash
cd Ecommerce

docker run --name cms-mysql \
  -e MYSQL_ROOT_PASSWORD=root_password \
  -e MYSQL_DATABASE=cms_db \
  -e MYSQL_USER=cms_user \
  -e MYSQL_PASSWORD=cms_password \
  -p 3306:3306 \
  -v cms_mysql_data:/var/lib/mysql \
  -d mysql:8.4

cd backend/services/cms-service

mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db < migrations/001_create_cms_access_tables.up.sql
mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db < migrations/002_create_product_moderation_reviews.up.sql
mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db < migrations/003_create_coupon_engine_tables.up.sql
mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db < migrations/004_create_offer_campaigns.up.sql
mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db < migrations/005_create_seller_analytics_tables.up.sql
mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db < migrations/006_create_seller_settings.up.sql
mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db < migrations/007_harden_cms_audit_logs.up.sql

go mod download
go test ./...

set -a
source .env
set +a
go run ./cmd/server
```

Then verify:

```bash
curl http://localhost:8087/healthz
```

Final simple summary: is service ko run karne ke liye sabse important cheezein hain Go, MySQL, migrations, `.env`, internal auth token, and Product Service URL. Redis/Kafka/RabbitMQ/MongoDB current service startup ke liye required nahi hain.
