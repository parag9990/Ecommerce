# Project Dependency & Setup Guide

Input analyzed: `TaskImplementation/API Gateway Service/task1.md`

Related implementation inspected:

- `backend/services/api-gateway/go.mod`
- `backend/services/api-gateway/cmd/server/main.go`
- `backend/services/api-gateway/internal/config/config.go`
- `backend/services/api-gateway/internal/clients/*`
- `backend/services/api-gateway/internal/ratelimit/*`
- `api/master-api.json`
- `docs/02-system-architecture.md`
- `docs/05-database-design.md`
- `docs/11-devops-external-services.md`
- `docs/13-developer-guide.md`

> Simple goal: Ye guide beginner developer ko batata hai ki API Gateway Service ko local machine par run/test karne ke liye code ke alawa kya-kya chahiye: Go, environment variables, Redis, downstream gRPC services, JWT/JWKS, Docker, databases, ports, and common fixes.

---

## 1. Project Overview

API Gateway Service is project ka public entry point hai. Frontend browser REST APIs call karega, jaise:

```text
GET /api/v1/products
POST /api/v1/auth/login
POST /api/v1/orders/checkout
```

Gateway internally in requests ko correct backend microservice ke gRPC method se map karega. Example:

```text
GET /api/v1/products -> product-service -> ProductService.ListProducts
```

### Important Current Reality

`task1.md` original route-contract documentation hai, but repo me current API Gateway Go code bhi present hai. Current code:

- Go service hai.
- `api/master-api.json` se route catalog load karta hai.
- HTTP server start karta hai.
- Redis se rate limiting karta hai when enabled.
- JWT/JWKS se protected routes verify karta hai.
- 12 downstream gRPC services ko startup par dial karta hai.
- Route bridge business logic abhi complete nahi hai; route match hone par handler currently `501 ROUTE_BRIDGE_NOT_CONFIGURED` return karta hai.

### Direct vs Indirect Dependencies

| Dependency | Directly needed by API Gateway? | Why |
|---|---:|---|
| Go | Yes | Gateway Go service build/run/test karne ke liye |
| `api/master-api.json` | Yes | Route contract source of truth |
| Redis | Yes, by default | Rate limiting enabled hai by default |
| Auth JWKS endpoint | Yes for protected routes | JWT signature verify karne ke liye |
| Downstream gRPC services | Yes for startup/readiness | Gateway all service clients initialize karta hai |
| MySQL | No direct Gateway DB | Downstream services use karte hain |
| MongoDB | No direct Gateway DB | Downstream services use karte hain |
| Typesense | No direct Gateway dependency | Search Service use karega |
| Kafka/RabbitMQ | No direct Gateway dependency in current code | Async events ke liye platform-level dependency |
| Docker | Recommended | Local Redis/DB/queue setup easy banata hai |

---

## 2. Tech Stack

| Technology | What it is | Why project uses it | Required? | Beginner Hinglish explanation |
|---|---|---|---:|---|
| Go | Backend programming language | Fast, compiled API Gateway service banane ke liye | Yes | Go ek simple aur fast backend language hai. Is project me API Gateway Go me likha gaya hai. |
| Go Modules | Go dependency system | Libraries ko version ke saath manage karne ke liye | Yes | `go.mod` dependency list hoti hai, `go.sum` checksum lock file hoti hai. |
| `net/http` | Go standard HTTP server | REST routes expose karne ke liye | Yes | Isme external web framework nahi hai; Go ka built-in HTTP server use ho raha hai. |
| gRPC | High-performance RPC framework | Gateway se internal services ko strongly typed calls bhejne ke liye | Yes | gRPC service-to-service communication ke liye fast protocol hai. |
| Protobuf | Data contract format | gRPC messages/schema ke liye | Yes | Protobuf request/response ka typed format define karta hai. |
| Redis | In-memory data store | Gateway rate limiting ke counters store karne ke liye | Yes by default | Redis fast memory database hai. Gateway per IP/user/route request limit track karta hai. |
| `go-redis` | Go Redis client | Go code se Redis commands/Lua script chalane ke liye | Yes if rate limit on | Ye library Gateway ko Redis se connect karwati hai. |
| JWT | Token format | Protected routes me user identity verify karne ke liye | Yes for protected routes | JWT login ke baad milne wala access token hota hai. |
| JWKS | Public key set endpoint | JWT signature verify karne ke liye Auth Service public keys deta hai | Yes for protected routes | JWKS se Gateway ko pata chalta hai token kis public key se verify karna hai. |
| JSON route contract | API contract file | Route, auth level, target service, gRPC mapping store karne ke liye | Yes | `api/master-api.json` Gateway ka route map hai. |
| Docker | Container runtime | Redis/MySQL/Mongo/queue local setup easy karne ke liye | Recommended | Docker se dependencies isolated containers me chalti hain. |
| MySQL | Relational database | Auth/User/Order/Payment/CMS/Superadmin services ke liye | Full stack only | MySQL table-based database hai. Gateway directly use nahi karta. |
| MongoDB | Document database | Product/Cart/Wishlist/Session/Notification services ke liye | Full stack only | MongoDB flexible JSON-like documents store karta hai. |
| Typesense | Search engine | Product search/autocomplete ke liye Search Service use karega | Full stack only | Typesense fast search engine hai. |
| Kafka/RabbitMQ | Message broker | Async events, indexing, notifications ke liye | Full stack only | Queue system async background kaam manage karta hai. |

---

## 3. Required Software

### Minimum for API Gateway Unit Tests

Install these first:

| Software | Version / Notes |
|---|---|
| Git | Repo clone karne ke liye |
| Go | Actual repo uses `go 1.26.3` in `backend/services/api-gateway/go.mod` and `backend/go.work` |
| curl | Health/API test karne ke liye |

### Recommended for Local Runtime

| Software | Why |
|---|---|
| Docker Desktop / Docker Engine | Redis, MySQL, MongoDB, Typesense, Kafka/RabbitMQ ko local run karne ke liye |
| Redis CLI | Redis ping/debug ke liye |
| MySQL client | Full stack DB verify karne ke liye |
| MongoDB Shell `mongosh` | MongoDB verify karne ke liye |
| grpcurl | gRPC downstream health debug karne ke liye |

### Install Examples

#### Windows

Recommended beginner path: Docker Desktop + WSL2.

```powershell
winget install Git.Git
winget install GoLang.Go
winget install Docker.DockerDesktop
```

After install:

```powershell
git --version
go version
docker version
```

For Redis/MySQL/Mongo on Windows, Docker is easier than native install.

#### Ubuntu / Debian Linux

```bash
sudo apt update
sudo apt install -y git curl ca-certificates lsb-release
```

Install Go from official Go package/tarball if apt version is old. Verify:

```bash
go version
```

Install Docker Engine or Docker Desktop for Linux, then verify:

```bash
docker version
docker compose version
```

Useful clients:

```bash
sudo apt install -y redis-tools mysql-client
```

#### macOS

Using Homebrew:

```bash
brew install git go curl redis mysql-client
brew install --cask docker
```

Verify:

```bash
git --version
go version
docker version
```

---

## 4. Dependency Management

This is a Go project.

### Important Files

| File | Purpose |
|---|---|
| `backend/services/api-gateway/go.mod` | Module name and direct dependency versions |
| `backend/services/api-gateway/go.sum` | Downloaded dependency checksums |
| `backend/go.work` | Go workspace file; currently includes `./services/api-gateway` |

### Current Go Module

```text
module ecommerce/api-gateway
go 1.26.3
```

### Direct Dependencies

| Dependency | Version | Why used |
|---|---:|---|
| `github.com/golang-jwt/jwt/v5` | `v5.3.1` | JWT parse/verify |
| `github.com/redis/go-redis/v9` | `v9.19.0` | Redis rate limiter |
| `google.golang.org/grpc` | `v1.81.1` | gRPC clients and health checks |
| `google.golang.org/protobuf` | `v1.36.11` | Protobuf runtime |
| `google.golang.org/genproto/googleapis/rpc` | `v0.0.0-20260226221140-a57be14db171` | Google RPC status/details support |

### Common Go Commands

Run from API Gateway service directory:

```bash
cd backend/services/api-gateway
```

Download dependencies:

```bash
go mod download
```

Clean dependency list:

```bash
go mod tidy
```

Run tests:

```bash
go test ./...
```

Build binary:

```bash
go build ./cmd/server
```

Run service:

```bash
go run ./cmd/server
```

### Dependency Problems and Fixes

| Problem | Cause | Fix |
|---|---|---|
| `go: go.mod requires go >= 1.26.3` | Local Go old hai | Install Go `1.26.3` or compatible newer version |
| `missing go.sum entry` | Dependency checksum missing | Run `go mod tidy` |
| `module lookup disabled` | `GOPROXY` or network issue | Check `go env GOPROXY`; normally `https://proxy.golang.org,direct` |
| `cannot find module` | Wrong working directory | Run commands inside `backend/services/api-gateway` |
| Version mismatch | `go.work` and `go.mod` alag Go version expect kar rahe | Keep `backend/go.work` and service `go.mod` aligned |
| Corporate proxy issue | Go cannot download modules | Configure `GOPROXY`, `GONOSUMDB`, or company proxy |

---

## 5. Database Setup

### API Gateway Database Requirement

API Gateway ka direct database **none** hai.

Gateway:

- MySQL directly connect nahi karta.
- MongoDB directly connect nahi karta.
- Tables/migrations run nahi karta.
- Service-owned databases ko directly read/write nahi karta.

Ye important architecture rule hai: API Gateway sirf REST request receive karta hai, auth/rate-limit/validation karta hai, phir gRPC downstream services ko call karta hai.

### Full Platform Database Matrix

Full ecommerce stack run karne ke liye downstream services databases use karenge:

| Service | Database |
|---|---|
| Auth | MySQL + Redis |
| User | MySQL |
| Product | MongoDB |
| Cart | MongoDB + Redis |
| Wishlist | MongoDB |
| Order | MySQL |
| Payment | MySQL |
| CMS | MySQL |
| Session | MongoDB + Redis |
| Notification | MongoDB |
| Superadmin | MySQL |
| Search | Typesense |

### MySQL

#### A. What It Is

MySQL ek relational database hai jisme data tables, rows, columns ke form me store hota hai. Ye transactional data ke liye best hai, jaise users, orders, payments, audit logs.

#### B. Why This Project Uses It

MySQL strong consistency deta hai. Auth, order, payment jaise workflows me data correct rehna very important hai.

#### C. Required or Optional

| Context | Required? |
|---|---:|
| API Gateway only | No |
| Full backend stack | Yes |
| Auth/User/Order/Payment/CMS/Superadmin services | Yes |

#### D. Local Installation

Windows:

```powershell
winget install Oracle.MySQL
```

Ubuntu / Debian:

```bash
sudo apt update
sudo apt install -y mysql-server mysql-client
sudo systemctl enable mysql
sudo systemctl start mysql
```

macOS:

```bash
brew install mysql
brew services start mysql
```

#### E. Docker Setup

```bash
docker volume create ecommerce_mysql_data

docker run -d \
  --name ecommerce-mysql \
  -e MYSQL_ROOT_PASSWORD=rootpassword \
  -e MYSQL_DATABASE=auth_db \
  -p 3306:3306 \
  -v ecommerce_mysql_data:/var/lib/mysql \
  mysql:8.4
```

Docker Compose example:

```yaml
services:
  mysql:
    image: mysql:8.4
    container_name: ecommerce-mysql
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: auth_db
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost", "-prootpassword"]
      interval: 10s
      timeout: 5s
      retries: 10

volumes:
  mysql_data:
```

#### F. Start Commands

Native Linux:

```bash
sudo systemctl start mysql
```

Docker:

```bash
docker start ecommerce-mysql
```

#### G. Verify Running

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p
```

Inside MySQL:

```sql
SHOW DATABASES;
```

#### H. Default Port

```text
3306
```

#### I. Connection String Format

Go MySQL DSN format:

```env
AUTH_MYSQL_DSN=auth_user:strong_password@tcp(127.0.0.1:3306)/auth_db?parseTime=true&charset=utf8mb4&loc=UTC
USER_SERVICE_DATABASE_DSN=user_user:strong_password@tcp(127.0.0.1:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC
```

#### J. Where To Place Credentials

MySQL credentials Gateway `.env` me nahi aate. Ye downstream service `.env` files me aate hain, for example:

```text
backend/services/auth-service/.env
backend/services/user-service/.env
```

Do not commit `.env`. Root `.gitignore` already ignores `.env` files.

#### Migrations / Schema

Current repo has a SQL design file:

```text
database/draw.sql
```

Beginner local setup:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p < database/draw.sql
```

Warning: `draw.sql` creates multiple service databases. Production should use proper migration tooling with up/down migrations, not one big manual SQL file.

### MongoDB

#### A. What It Is

MongoDB ek document database hai. Isme JSON-like documents store hote hain. Product catalog, cart, session events jaise flexible data ke liye useful hai.

#### B. Why This Project Uses It

Products ke attributes dynamic ho sakte hain: size, color, warranty, variants. Session events high-volume and flexible hote hain. MongoDB is type ke data ke liye suitable hai.

#### C. Required or Optional

| Context | Required? |
|---|---:|
| API Gateway only | No |
| Full backend stack | Yes |
| Product/Cart/Wishlist/Session/Notification services | Yes |

#### D. Local Installation

Windows:

- Beginner recommended: use Docker.
- Native option: MongoDB Community Server installer from MongoDB official site.

Ubuntu / Debian:

- Beginner recommended: use Docker because official apt setup has distro-specific repo steps.

macOS:

```bash
brew tap mongodb/brew
brew install mongodb-community
brew services start mongodb-community
```

#### E. Docker Setup

```bash
docker volume create ecommerce_mongo_data

docker run -d \
  --name ecommerce-mongo \
  -p 27017:27017 \
  -v ecommerce_mongo_data:/data/db \
  mongo:7
```

Docker Compose example:

```yaml
services:
  mongodb:
    image: mongo:7
    container_name: ecommerce-mongo
    ports:
      - "27017:27017"
    volumes:
      - mongo_data:/data/db
    healthcheck:
      test: ["CMD", "mongosh", "--eval", "db.adminCommand('ping')"]
      interval: 10s
      timeout: 5s
      retries: 10

volumes:
  mongo_data:
```

#### F. Start Commands

Native macOS:

```bash
brew services start mongodb-community
```

Docker:

```bash
docker start ecommerce-mongo
```

#### G. Verify Running

```bash
mongosh "mongodb://localhost:27017"
```

Inside shell:

```javascript
db.adminCommand({ ping: 1 })
```

#### H. Default Port

```text
27017
```

#### I. Connection String Format

```env
SESSION_MONGO_URI=mongodb://localhost:27017
SESSION_MONGO_DATABASE=session_db
PRODUCT_MONGO_URI=mongodb://localhost:27017
PRODUCT_MONGO_DATABASE=product_db
```

#### J. Where To Place Credentials

Mongo credentials Gateway `.env` me nahi aate. They belong to downstream service `.env` files, such as:

```text
backend/services/session-service/.env
backend/services/product-service/.env
```

Mongo schema design reference:

```text
database/mongodb-schema-design.md
```

### Redis

#### A. What It Is

Redis ek in-memory data store hai. Ye very fast hota hai and temporary counters, cache, sessions, OTP rate limits, and distributed locks ke liye use hota hai.

#### B. Why This Project Uses It

API Gateway directly Redis use karta hai for rate limiting. Example: ek IP 1 minute me kitni requests kar sakta hai.

#### C. Required or Optional

| Context | Required? |
|---|---:|
| API Gateway with `RATE_LIMIT_ENABLED=true` | Yes |
| API Gateway with `RATE_LIMIT_ENABLED=false` | No, but not recommended for real dev/prod |
| Full backend stack | Yes |

#### D. Local Installation

Windows:

- Recommended: Docker or WSL2.

Ubuntu / Debian:

```bash
sudo apt update
sudo apt install -y redis-server redis-tools
sudo systemctl enable redis-server
sudo systemctl start redis-server
```

macOS:

```bash
brew install redis
brew services start redis
```

#### E. Docker Setup

Without password, simple local dev:

```bash
docker volume create ecommerce_redis_data

docker run -d \
  --name ecommerce-redis \
  -p 6379:6379 \
  -v ecommerce_redis_data:/data \
  redis:7-alpine redis-server --appendonly yes
```

With password, closer to production:

```bash
docker run -d \
  --name ecommerce-redis \
  -p 6379:6379 \
  -v ecommerce_redis_data:/data \
  redis:7-alpine redis-server --appendonly yes --requirepass localredispass
```

Docker Compose example:

```yaml
services:
  redis:
    image: redis:7-alpine
    container_name: ecommerce-redis
    command: ["redis-server", "--appendonly", "yes", "--requirepass", "localredispass"]
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "localredispass", "PING"]
      interval: 10s
      timeout: 5s
      retries: 10

volumes:
  redis_data:
```

#### F. Start Commands

Native Linux:

```bash
sudo systemctl start redis-server
```

Docker:

```bash
docker start ecommerce-redis
```

#### G. Verify Running

No password:

```bash
redis-cli -h localhost -p 6379 PING
```

With password:

```bash
redis-cli -h localhost -p 6379 -a localredispass PING
```

Expected:

```text
PONG
```

#### H. Default Port

```text
6379
```

#### I. Connection String / Env Format

API Gateway uses separate env vars:

```env
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=localredispass
REDIS_DB=0
REDIS_TLS_ENABLED=false
```

#### J. Where To Place Credentials

For API Gateway:

```text
backend/services/api-gateway/.env
```

But `.env` is not committed. Create it locally and export it before running.

---

## 6. Redis / Queue / External Services

### Direct API Gateway External Services

| Service | Directly used? | Why |
|---|---:|---|
| Redis | Yes | Rate limiting |
| Auth Service JWKS endpoint | Yes | JWT verification |
| Downstream gRPC services | Yes | Startup clients and readiness health |

### Downstream gRPC Services

The Gateway config requires all these addresses:

| Env Var | Service | Example local value | Example Docker/K8s value |
|---|---|---|---|
| `AUTH_GRPC_ADDR` | Auth Service | `localhost:50051` | `auth-service:9090` |
| `USER_GRPC_ADDR` | User Service | `localhost:50052` | `user-service:9090` |
| `PRODUCT_GRPC_ADDR` | Product Service | `localhost:50053` | `product-service:9090` |
| `CART_GRPC_ADDR` | Cart Service | `localhost:50054` | `cart-service:9090` |
| `WISHLIST_GRPC_ADDR` | Wishlist Service | `localhost:50055` | `wishlist-service:9090` |
| `ORDER_GRPC_ADDR` | Order Service | `localhost:50056` | `order-service:9090` |
| `PAYMENT_GRPC_ADDR` | Payment Service | `localhost:50057` | `payment-service:9090` |
| `SEARCH_GRPC_ADDR` | Search Service | `localhost:50058` | `search-service:9090` |
| `CMS_GRPC_ADDR` | CMS Service | `localhost:50059` | `cms-service:9090` |
| `SESSION_GRPC_ADDR` | Session Service | `localhost:50060` | `session-service:9090` |
| `NOTIFICATION_GRPC_ADDR` | Notification Service | `localhost:50061` | `notification-service:9090` |
| `SUPERADMIN_GRPC_ADDR` | Superadmin Service | `localhost:50062` | `superadmin-service:9090` |

Important: Current Gateway startup dials every service and waits until connection is ready. Agar service running nahi hai, Gateway start fail karega with dial timeout.

Health check expectation: Downstream services should implement standard gRPC health service. Gateway readiness calls service names like:

```text
ecommerce.auth.v1.AuthService
ecommerce.product.v1.ProductService
ecommerce.session.v1.SessionService
```

### Auth JWKS Endpoint

#### What It Is

JWKS endpoint public keys expose karta hai. Gateway JWT access token ko verify karne ke liye ye keys fetch/cache karta hai.

#### Required?

Yes, because `api/master-api.json` contains protected routes like `/api/v1/me`, `/api/v1/cart`, `/api/v1/admin/users`.

#### Env Placement

```env
JWT_JWKS_URL=http://localhost:8081/.well-known/jwks.json
JWT_ISSUER=ecommerce-auth
JWT_AUDIENCE=ecommerce-api
JWT_ALLOWED_ALGS=RS256
```

#### Health Check

```bash
curl -i http://localhost:8081/.well-known/jwks.json
```

Expected: JSON with `keys`.

### Typesense

#### What It Is

Typesense ek search engine hai. Search Service product search/autocomplete ke liye use karega.

#### Required?

Not directly for API Gateway, but required when Search Service is fully implemented.

#### Docker Setup

```bash
docker volume create ecommerce_typesense_data

docker run -d \
  --name ecommerce-typesense \
  -p 8108:8108 \
  -v ecommerce_typesense_data:/data \
  typesense/typesense:latest \
  --data-dir /data \
  --api-key dev-typesense-key \
  --enable-cors
```

Verify:

```bash
curl http://localhost:8108/health
```

### Kafka or RabbitMQ

#### What It Is

Kafka/RabbitMQ message broker hai. Async events ke liye use hota hai, jaise `OrderCreated`, `PaymentCaptured`, `ProductUpdated`.

#### Required?

Not directly for current API Gateway code. Required for full event-driven platform.

#### Kafka-Compatible Local Setup Using Redpanda

```bash
docker run -d \
  --name ecommerce-redpanda \
  -p 9092:9092 \
  redpandadata/redpanda:latest \
  redpanda start \
  --overprovisioned \
  --smp 1 \
  --memory 512M \
  --reserve-memory 0M \
  --node-id 0 \
  --check=false \
  --kafka-addr PLAINTEXT://0.0.0.0:9092 \
  --advertise-kafka-addr PLAINTEXT://localhost:9092
```

#### RabbitMQ Alternative

```bash
docker run -d \
  --name ecommerce-rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:3-management
```

RabbitMQ management UI:

```text
http://localhost:15672
```

Default dev login is usually `guest` / `guest` for local Docker only. Do not use this in production.

### Payment Providers / SMTP / S3

Task 1 route contract includes payment webhook route:

```text
POST /api/v1/webhooks/payments/{provider}
```

Current API Gateway only checks that webhook signature header exists. Actual Stripe/Razorpay-like provider secret verification belongs to Payment Service / later tasks.

| Service | Current Gateway requirement |
|---|---|
| Stripe/Razorpay | Only webhook signature header name config |
| SMTP | Not direct Gateway dependency |
| AWS S3/MinIO | Not direct Gateway dependency |
| Firebase/Twilio | Not direct Gateway dependency |

---

## 7. Environment Variables

### Where To Create `.env`

Recommended local file:

```text
backend/services/api-gateway/.env
```

Important: Current Go code uses `os.Getenv`. It does **not** automatically load `.env`. You must export variables before running.

Linux/macOS:

```bash
cd backend/services/api-gateway
set -a
source .env
set +a
go run ./cmd/server
```

Windows PowerShell example:

```powershell
cd backend/services/api-gateway
Get-Content .env | ForEach-Object {
  if ($_ -match "^\s*#" -or $_ -match "^\s*$") { return }
  $name, $value = $_.Split("=", 2)
  [Environment]::SetEnvironmentVariable($name, $value, "Process")
}
go run ./cmd/server
```

### Complete `.env` Example

```env
# Basic app config
SERVICE_NAME=api-gateway
APP_ENV=local
HTTP_ADDR=:8080
API_BASE_PATH=/api/v1
API_CONTRACT_PATH=../../../api/master-api.json
LOG_LEVEL=debug
HTTP_READ_HEADER_TIMEOUT=5s
HTTP_SHUTDOWN_TIMEOUT=10s

# gRPC client config
GRPC_TLS_ENABLED=false
GRPC_DIAL_TIMEOUT=3s

# Downstream gRPC targets
AUTH_GRPC_ADDR=localhost:50051
USER_GRPC_ADDR=localhost:50052
PRODUCT_GRPC_ADDR=localhost:50053
CART_GRPC_ADDR=localhost:50054
WISHLIST_GRPC_ADDR=localhost:50055
ORDER_GRPC_ADDR=localhost:50056
PAYMENT_GRPC_ADDR=localhost:50057
SEARCH_GRPC_ADDR=localhost:50058
CMS_GRPC_ADDR=localhost:50059
SESSION_GRPC_ADDR=localhost:50060
NOTIFICATION_GRPC_ADDR=localhost:50061
SUPERADMIN_GRPC_ADDR=localhost:50062

# Redis for rate limiting
RATE_LIMIT_ENABLED=true
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_TLS_ENABLED=false
REDIS_DIAL_TIMEOUT=3s
RATE_LIMIT_KEY_PREFIX=rl:v1
RATE_LIMIT_FAIL_OPEN=false
RATE_LIMIT_DEFAULT_IP_LIMIT=600
RATE_LIMIT_DEFAULT_IP_WINDOW=1m
RATE_LIMIT_TRUSTED_PROXY_CIDRS=
RATE_LIMIT_TARGET_BODY_LIMIT_BYTES=65536

# Request validation
REQUEST_VALIDATION_ENABLED=true
REQUEST_VALIDATION_DEFAULT_MAX_BODY_BYTES=131072
REQUEST_VALIDATION_MAX_HEADER_BYTES=32768
REQUEST_VALIDATION_MAX_QUERY_BYTES=8192

# JWT / Auth Service JWKS
JWT_ISSUER=ecommerce-auth
JWT_AUDIENCE=ecommerce-api
JWT_ALLOWED_ALGS=RS256
JWT_JWKS_URL=http://localhost:8081/.well-known/jwks.json
JWT_JWKS_CACHE_TTL=5m
JWT_JWKS_FETCH_TIMEOUT=3s
JWT_CLOCK_SKEW=30s

# Webhooks
WEBHOOK_SIGNATURE_HEADER=X-Provider-Signature
```

### Variable Explanation

| Variable | Required? | Example | Purpose | Security notes |
|---|---:|---|---|---|
| `SERVICE_NAME` | Optional | `api-gateway` | Logs/health response service name | Safe |
| `APP_ENV` | Optional | `local` | Environment label | Safe |
| `HTTP_ADDR` | Optional | `:8080` | HTTP bind address | Avoid exposing publicly in dev |
| `API_BASE_PATH` | Optional | `/api/v1` | REST route prefix | Must start with `/` |
| `API_CONTRACT_PATH` | Required if auto-discovery fails | `../../../api/master-api.json` | Route contract JSON path | Safe |
| `LOG_LEVEL` | Optional | `debug` | Logging verbosity | Avoid debug in prod if it logs too much |
| `HTTP_READ_HEADER_TIMEOUT` | Optional | `5s` | Slowloris protection | Keep positive |
| `HTTP_SHUTDOWN_TIMEOUT` | Optional | `10s` | Graceful shutdown time | Keep positive |
| `GRPC_TLS_ENABLED` | Optional | `false` | TLS for downstream gRPC | Prod should use TLS/mTLS |
| `GRPC_DIAL_TIMEOUT` | Optional | `3s` | Startup dial timeout | Keep realistic |
| `AUTH_GRPC_ADDR` | Yes | `localhost:50051` | Auth Service gRPC target | No spaces |
| `USER_GRPC_ADDR` | Yes | `localhost:50052` | User Service gRPC target | No spaces |
| `PRODUCT_GRPC_ADDR` | Yes | `localhost:50053` | Product Service gRPC target | No spaces |
| `CART_GRPC_ADDR` | Yes | `localhost:50054` | Cart Service gRPC target | No spaces |
| `WISHLIST_GRPC_ADDR` | Yes | `localhost:50055` | Wishlist Service gRPC target | No spaces |
| `ORDER_GRPC_ADDR` | Yes | `localhost:50056` | Order Service gRPC target | No spaces |
| `PAYMENT_GRPC_ADDR` | Yes | `localhost:50057` | Payment Service gRPC target | No spaces |
| `SEARCH_GRPC_ADDR` | Yes | `localhost:50058` | Search Service gRPC target | No spaces |
| `CMS_GRPC_ADDR` | Yes | `localhost:50059` | CMS Service gRPC target | No spaces |
| `SESSION_GRPC_ADDR` | Yes | `localhost:50060` | Session Service gRPC target | No spaces |
| `NOTIFICATION_GRPC_ADDR` | Yes | `localhost:50061` | Notification Service gRPC target | No spaces |
| `SUPERADMIN_GRPC_ADDR` | Yes | `localhost:50062` | Superadmin Service gRPC target | No spaces |
| `RATE_LIMIT_ENABLED` | Optional | `true` | Enable Redis rate limiter | Keep true for realistic testing |
| `REDIS_ADDR` | Required when rate limit on | `localhost:6379` | Redis host/port | No spaces |
| `REDIS_PASSWORD` | Optional | blank or secret | Redis auth password | Secret, do not commit |
| `REDIS_DB` | Optional | `0` | Redis logical DB | Must be non-negative |
| `REDIS_TLS_ENABLED` | Optional | `false` | Redis TLS | Prod managed Redis may require true |
| `REDIS_DIAL_TIMEOUT` | Optional | `3s` | Redis connect timeout | Keep positive |
| `RATE_LIMIT_KEY_PREFIX` | Required when rate limit on | `rl:v1` | Redis key namespace | No spaces |
| `RATE_LIMIT_FAIL_OPEN` | Optional | `false` | If Redis fails, allow or block traffic | Prod usually fail closed for sensitive routes |
| `RATE_LIMIT_DEFAULT_IP_LIMIT` | Optional | `600` | Default IP request limit | Positive integer |
| `RATE_LIMIT_DEFAULT_IP_WINDOW` | Optional | `1m` | Rate limit window | Positive duration |
| `RATE_LIMIT_TRUSTED_PROXY_CIDRS` | Optional | `10.0.0.0/8` | Trusted proxy IP ranges | Wrong value can spoof client IP |
| `RATE_LIMIT_TARGET_BODY_LIMIT_BYTES` | Optional | `65536` | Body bytes inspected for target limits | Positive integer |
| `REQUEST_VALIDATION_ENABLED` | Optional | `true` | Request validation middleware | Keep true |
| `REQUEST_VALIDATION_DEFAULT_MAX_BODY_BYTES` | Optional | `131072` | Default request body limit | Positive integer |
| `REQUEST_VALIDATION_MAX_HEADER_BYTES` | Optional | `32768` | Header limit | Positive integer |
| `REQUEST_VALIDATION_MAX_QUERY_BYTES` | Optional | `8192` | Query string limit | Positive integer |
| `JWT_ISSUER` | Required | `ecommerce-auth` | Expected token issuer | Must match Auth Service |
| `JWT_AUDIENCE` | Required | `ecommerce-api` | Expected token audience | Must match Auth Service |
| `JWT_ALLOWED_ALGS` | Required | `RS256` | Allowed JWT signing algorithms | Do not allow `none`; code supports RSA algs |
| `JWT_JWKS_URL` | Required for protected routes | `http://localhost:8081/.well-known/jwks.json` | Public key fetch URL | Use HTTPS in prod |
| `JWT_JWKS_CACHE_TTL` | Optional | `5m` | JWKS cache duration | Keep positive |
| `JWT_JWKS_FETCH_TIMEOUT` | Optional | `3s` | JWKS HTTP fetch timeout | Keep positive |
| `JWT_CLOCK_SKEW` | Optional | `30s` | Small clock difference tolerance | Code rejects more than `5m` |
| `WEBHOOK_SIGNATURE_HEADER` | Optional | `X-Provider-Signature` | Header required for webhook routes | Actual secret verification should be in Payment Service |

### Common `.env` Mistakes

| Mistake | Result | Fix |
|---|---|---|
| `.env` created but not exported | Go cannot see variables | Use `set -a; source .env; set +a` |
| `API_CONTRACT_PATH` wrong | Startup fails: contract not readable | Use relative path from `backend/services/api-gateway` or absolute path |
| Missing gRPC address | Startup validation fails | Add all `*_GRPC_ADDR` variables |
| Redis password mismatch | Startup fails: `redis ping failed` | Match `REDIS_PASSWORD` with Redis server |
| `JWT_JWKS_URL` blank | Router fails when protected routes exist | Start Auth Service JWKS endpoint and set URL |
| Quoted values with extra quotes | Some env values include quote chars | Use simple `KEY=value` format |

### Ports & Networking

| Service | Default Port | Purpose | Directly needed by API Gateway? | How to change |
|---|---:|---|---:|---|
| API Gateway HTTP | `8080` | Public REST API and health endpoints | Yes | `HTTP_ADDR=:8081` |
| Auth Service HTTP/JWKS | `8081` | JWKS public key endpoint | Yes for protected routes | Auth Service env, then update `JWT_JWKS_URL` |
| Auth Service gRPC | `50051` local or `9090` Docker/K8s | Auth route target and health | Yes | `AUTH_GRPC_ADDR` |
| User Service gRPC | `50052` local or `9090` Docker/K8s | User/profile target | Yes | `USER_GRPC_ADDR` |
| Product Service gRPC | `50053` local or `9090` Docker/K8s | Product/catalog target | Yes | `PRODUCT_GRPC_ADDR` |
| Cart Service gRPC | `50054` local or `9090` Docker/K8s | Cart target | Yes | `CART_GRPC_ADDR` |
| Wishlist Service gRPC | `50055` local or `9090` Docker/K8s | Wishlist target | Yes | `WISHLIST_GRPC_ADDR` |
| Order Service gRPC | `50056` local or `9090` Docker/K8s | Checkout/order target | Yes | `ORDER_GRPC_ADDR` |
| Payment Service gRPC | `50057` local or `9090` Docker/K8s | Payment/refund/webhook target | Yes | `PAYMENT_GRPC_ADDR` |
| Search Service gRPC | `50058` local or `9090` Docker/K8s | Search/autocomplete target | Yes | `SEARCH_GRPC_ADDR` |
| CMS Service gRPC | `50059` local or `9090` Docker/K8s | Seller CMS target | Yes | `CMS_GRPC_ADDR` |
| Session Service gRPC | `50060` local or `9090` Docker/K8s | Session/analytics target | Yes | `SESSION_GRPC_ADDR` |
| Notification Service gRPC | `50061` local or `9090` Docker/K8s | Notification preferences target | Yes | `NOTIFICATION_GRPC_ADDR` |
| Superadmin Service gRPC | `50062` local or `9090` Docker/K8s | Admin/superadmin target | Yes | `SUPERADMIN_GRPC_ADDR` |
| Redis | `6379` | Rate limit counters | Yes when rate limit enabled | `REDIS_ADDR=host:port` |
| MySQL | `3306` | Downstream relational DB | No direct | Downstream service DSN |
| MongoDB | `27017` | Downstream document DB | No direct | Downstream service Mongo URI |
| Typesense | `8108` | Search index | No direct | Search Service config |
| Kafka / Redpanda | `9092` | Event streaming | No direct | Producer/consumer config |
| RabbitMQ AMQP | `5672` | Queue alternative | No direct | Producer/consumer config |
| RabbitMQ UI | `15672` | Local queue dashboard | No direct | Docker port mapping |

Networking beginner notes:

- `localhost` ka matlab same machine. Agar Gateway host machine par run ho raha hai and Redis Docker me hai with `-p 6379:6379`, then `REDIS_ADDR=localhost:6379` works.
- Docker Compose ke andar service-to-service call karte waqt `localhost` use mat karo. Compose network me service name use karo, jaise `redis:6379` or `auth-service:9090`.
- Kubernetes me service DNS use hota hai, jaise `auth-service.core.svc.cluster.local:9090`.
- Port conflict aaye to app ka env port change karo or old process stop karo.
- Firewall/VPN/corporate proxy local ports block kar sakte hain. Pehle `curl`, `redis-cli`, `mysql`, or `grpcurl` se direct connectivity verify karo.

Port conflict debug:

```bash
lsof -i :8080
lsof -i :6379
```

Windows:

```powershell
netstat -ano | findstr :8080
netstat -ano | findstr :6379
```

---

## 8. Docker Setup

### Current Repo Status

No committed Dockerfile or docker-compose file was found for the current API Gateway service or `infra` folder.

So this guide provides recommended Docker examples, but these files are not currently present:

```text
backend/services/api-gateway/deploy/Dockerfile
infra/compose/docker-compose.local.yml
```

### Recommended Beginner Approach

Use Docker for dependencies first:

- Redis
- MySQL
- MongoDB
- Typesense
- Kafka-compatible broker or RabbitMQ

Run API Gateway locally with `go run` until the project has a maintained Dockerfile.

### Local Dependency Compose Example

```yaml
services:
  redis:
    image: redis:7-alpine
    command: ["redis-server", "--appendonly", "yes"]
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  mysql:
    image: mysql:8.4
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql

  mongodb:
    image: mongo:7
    ports:
      - "27017:27017"
    volumes:
      - mongo_data:/data/db

  typesense:
    image: typesense/typesense:latest
    command: ["--data-dir", "/data", "--api-key", "dev-typesense-key", "--enable-cors"]
    ports:
      - "8108:8108"
    volumes:
      - typesense_data:/data

  rabbitmq:
    image: rabbitmq:3-management
    ports:
      - "5672:5672"
      - "15672:15672"

volumes:
  redis_data:
  mysql_data:
  mongo_data:
  typesense_data:
```

### Common Docker Commands

Start stack:

```bash
docker compose up -d
```

Stop stack:

```bash
docker compose down
```

Stop and delete volumes:

```bash
docker compose down -v
```

View logs:

```bash
docker compose logs -f redis
docker compose logs -f mysql
```

List containers:

```bash
docker ps
```

### Docker Volumes

Volumes persist data even if container restarts.

| Volume | Purpose |
|---|---|
| `redis_data` | Redis append-only data |
| `mysql_data` | MySQL databases |
| `mongo_data` | MongoDB documents |
| `typesense_data` | Search index data |

Beginner note: Agar local data corrupt ho jaye, `docker compose down -v` clean reset karta hai, but all local DB data delete ho jayega.

### API Gateway Dockerfile Gap

Expected future Dockerfile shape:

```dockerfile
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY backend/go.work backend/go.work.sum ./backend/
COPY backend/services/api-gateway ./backend/services/api-gateway
COPY api ./api
WORKDIR /app/backend/services/api-gateway
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api-gateway ./cmd/server

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=builder /out/api-gateway /api-gateway
COPY api/master-api.json /app/api/master-api.json
EXPOSE 8080
ENTRYPOINT ["/api-gateway"]
```

This is documentation guidance only. Do not assume it already exists in repo.

---

## 9. Local Development Setup

### Step 1: Clone Repository

```bash
git clone <repo-url>
cd Ecommerce
```

### Step 2: Check Go

```bash
go version
```

Expected compatible version:

```text
go1.26.3
```

### Step 3: Install Go Dependencies

```bash
cd backend/services/api-gateway
go mod download
```

### Step 4: Run Tests

```bash
go test ./...
```

This verifies code-level dependencies without needing Redis/downstream services.

### Step 5: Start Redis

Simple local Redis:

```bash
docker run -d --name ecommerce-redis -p 6379:6379 redis:7-alpine
```

Verify:

```bash
redis-cli -h localhost -p 6379 PING
```

### Step 6: Start Downstream gRPC Services

Gateway needs all downstream gRPC services reachable before it can start normally.

Required:

```text
auth-service
user-service
product-service
cart-service
wishlist-service
order-service
payment-service
search-service
cms-service
session-service
notification-service
superadmin-service
```

If those services are not implemented/running yet, `go run ./cmd/server` will fail during gRPC client initialization. For dependency-only verification, run tests instead.

### Step 7: Start Auth JWKS Endpoint

Protected routes require:

```env
JWT_JWKS_URL=http://localhost:8081/.well-known/jwks.json
```

Verify:

```bash
curl -i http://localhost:8081/.well-known/jwks.json
```

### Step 8: Create and Export API Gateway Env

```bash
cd backend/services/api-gateway
set -a
source .env
set +a
```

### Step 9: Run API Gateway

```bash
go run ./cmd/server
```

### Step 10: Verify Health

```bash
curl -i http://localhost:8080/health/live
curl -i http://localhost:8080/health/ready
```

Expected live response:

```json
{
  "data": {
    "status": "ok",
    "service": "api-gateway"
  },
  "request_id": "...",
  "error": null
}
```

Readiness requires route catalog and downstream services to be healthy.

### Step 11: Verify Route Contract

Public route:

```bash
curl -i http://localhost:8080/api/v1/products
```

Current expected behavior after route match:

```text
501 ROUTE_BRIDGE_NOT_CONFIGURED
```

This is expected because Task 1 defines the route contract; downstream REST-to-gRPC business bridge is later implementation.

Protected route without token:

```bash
curl -i http://localhost:8080/api/v1/me
```

Expected:

```text
401 Unauthorized
```

---

## 10. Running the Project

### Quick Test-Only Flow

Use this when downstream services are not ready:

```bash
cd backend/services/api-gateway
go mod download
go test ./...
```

### Full Local Runtime Flow

1. Clone repo.
2. Install Go.
3. Start Redis.
4. Start all downstream gRPC services.
5. Start Auth Service JWKS endpoint.
6. Export API Gateway `.env`.
7. Run Gateway.
8. Check `/health/live`.
9. Check `/health/ready`.
10. Call public route.

Commands:

```bash
cd backend/services/api-gateway
go mod download

set -a
source .env
set +a

go run ./cmd/server
```

In another terminal:

```bash
curl -i http://localhost:8080/health/live
curl -i http://localhost:8080/health/ready
curl -i http://localhost:8080/api/v1/products
```

### Migration Commands

API Gateway itself has no DB migrations.

For full MySQL schema setup:

```bash
cd /home/parag/Ecommerce
mysql -h 127.0.0.1 -P 3306 -u root -p < database/draw.sql
```

MongoDB currently has schema design documentation, not executable migrations:

```text
database/mongodb-schema-design.md
```

---

## 11. Common Errors & Fixes

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `API_CONTRACT_PATH is not readable` | Wrong path or running from wrong directory | Set `API_CONTRACT_PATH=/home/parag/Ecommerce/api/master-api.json` | Use absolute path in local `.env` |
| `AUTH_GRPC_ADDR is required` | Missing env var | Add all `*_GRPC_ADDR` vars | Keep `.env.example` updated |
| `dial auth service: context deadline exceeded` | Downstream service not running or wrong port | Start service or correct `AUTH_GRPC_ADDR` | Use Docker Compose/service discovery |
| `redis ping failed` | Redis down, wrong password, wrong DB/host | Start Redis and verify with `redis-cli PING` | Add Redis healthcheck |
| `JWT_JWKS_URL is required when protected routes are configured` | Master API has protected routes but JWKS URL blank | Start Auth Service JWKS endpoint and set URL | Include JWKS in `.env.example` |
| `listen tcp :8080: bind: address already in use` | Port 8080 busy | Change `HTTP_ADDR=:8088` or stop old process | Document port ownership |
| `/health/ready` returns 503 | Downstream gRPC health not serving | Check each service and gRPC health implementation | Add readiness probes |
| Protected route returns 401 | Missing/invalid `Authorization: Bearer` token | Login and pass valid token | Use API client collection |
| Webhook route rejects request | Missing signature header | Send configured `X-Provider-Signature` header | Provider-specific docs |
| `go mod download` fails | Network/proxy/cache problem | Check internet, `GOPROXY`, company proxy | Vendor/cache dependencies in CI if needed |
| `go version` too old | Installed Go does not match module | Install required Go version | Use asdf/gvm/toolchain docs |
| Docker command fails | Docker daemon not running | Start Docker Desktop/daemon | Add onboarding check |
| MySQL `Access denied` | Wrong user/password | Reset credentials or update DSN | Store credentials consistently |
| Mongo `connection refused` | MongoDB not running | Start MongoDB container/service | Use compose healthcheck |
| Redis `NOAUTH Authentication required` | Redis requires password but env blank | Set `REDIS_PASSWORD` | Keep Redis config and `.env` aligned |
| Kafka/Rabbit connection errors | Broker not ready or wrong port | Wait for broker health, verify port | Use compose dependencies/healthchecks |

Port debug:

Linux/macOS:

```bash
lsof -i :8080
```

Windows:

```powershell
netstat -ano | findstr :8080
```

---

## 12. Security & Best Practices

### Security Notes

- Never commit `.env`, secrets, private keys, database dumps, or `.pem` files.
- Root `.gitignore` already ignores `.env`, `.env.*`, `*.pem`, `*.key`, and `secrets/*`.
- Use strong Redis password in shared environments.
- Use HTTPS for `JWT_JWKS_URL` outside local development.
- Use `GRPC_TLS_ENABLED=true` or mTLS in staging/production.
- Keep `JWT_ALLOWED_ALGS` strict. Current code supports RSA algorithms like `RS256`, `RS384`, `RS512`.
- Do not log full JWT tokens, Redis passwords, DB DSNs, provider API keys, or webhook secrets.
- Keep `RATE_LIMIT_ENABLED=true` for real environments.
- Use separate dev/staging/prod configs.
- Rotate JWT keys and Redis/DB passwords.
- Use Docker volumes for local DB persistence.
- Use health checks for Redis, MySQL, MongoDB, and downstream gRPC services.
- Back up MySQL and MongoDB before destructive operations.

### Beginner Best Practices

| Practice | Why |
|---|---|
| Keep one `.env` per service | Config confusion kam hota hai |
| Use Docker for local dependencies | OS-specific install problems kam hote hain |
| Run `go test ./...` before `go run` | Dependency/code sanity check ho jata hai |
| Use absolute `API_CONTRACT_PATH` if confused | Relative path mistakes avoid hoti hain |
| Start dependencies first, app second | Startup errors easy debug hote hain |
| Check `/health/live` then `/health/ready` | Live means process alive; ready means dependencies ready |
| Keep ports documented | Port conflicts jaldi solve hote hain |
| Use strong local secrets too | Bad habits prod me leak nahi hote |

---

## 13. Missing or Misconfigured Things

This section is a professional setup audit based on current repo state.

| Finding | Impact | Suggested Fix |
|---|---|---|
| No API Gateway `.env.example` found | Beginners do not know required vars | Add `backend/services/api-gateway/.env.example` with safe placeholders |
| No committed Docker Compose file found | Local dependency startup manual ho jata hai | Add `infra/compose/docker-compose.local.yml` |
| No API Gateway Dockerfile found | Containerized run/deploy incomplete | Add service Dockerfile under `backend/services/api-gateway/deploy/` |
| Gateway code does not auto-load `.env` | `go run` fails if user forgets export | Document export command or add dev-only env loader |
| Startup requires every downstream gRPC service | Partial local development difficult | Add mock mode or allow disabled downstream clients for route-contract-only dev |
| `JWT_JWKS_URL` is config-optional but runtime-required for protected routes | Confusing startup error | Mark it required in docs and maybe config validation when contract has protected routes |
| Route bridge returns `501` | Public routes are defined but not functionally proxied yet | Later task should implement REST-to-gRPC bridge handlers |
| MySQL has `database/draw.sql` but no migration runner | Repeatable DB setup is weak | Add per-service migrations and migration command |
| MongoDB has design docs but no migration/index automation | Index creation can be missed | Add index bootstrap scripts per Mongo-backed service |
| Existing local `.env` files are present under other services | Local secrets can accidentally leak if gitignore changes | Keep `.gitignore`, rotate any shared sample secrets, add `.env.example` only |
| No healthcheck compose config in repo | Readiness debugging harder | Add healthchecks for Redis/MySQL/Mongo/Typesense/brokers |
| No documented gRPC port map for all services | Address config guesswork | Standardize local gRPC ports or use Docker DNS names |

---

## 14. Final Checklist

Use this before asking "why is Gateway not running?"

```text
[ ] Git installed
[ ] Go installed and compatible with go.mod
[ ] Repository cloned
[ ] Ran: cd backend/services/api-gateway
[ ] Ran: go mod download
[ ] Ran: go test ./...
[ ] Redis running if RATE_LIMIT_ENABLED=true
[ ] REDIS_ADDR and REDIS_PASSWORD correct
[ ] api/master-api.json exists
[ ] API_CONTRACT_PATH points to readable file
[ ] API Gateway .env created locally
[ ] .env variables exported into shell
[ ] All 12 *_GRPC_ADDR variables configured
[ ] Downstream gRPC services running
[ ] Downstream services implement gRPC health checks
[ ] Auth JWKS endpoint running
[ ] JWT_JWKS_URL configured
[ ] HTTP_ADDR port is free
[ ] Gateway starts with go run ./cmd/server
[ ] /health/live returns ok
[ ] /health/ready returns ready
[ ] Public route like /api/v1/products reaches Gateway
[ ] Protected route without token returns 401
[ ] .env and secrets are not staged in git
```

Final beginner note: API Gateway ke liye sabse common confusion ye hai ki Gateway direct DB use nahi karta, but phir bhi startup par Redis, Auth JWKS, and downstream gRPC services chahiye. Agar aap sirf code dependencies verify karna chahte ho, `go test ./...` enough hai. Agar actual runtime verify karna hai, to dependencies first, Gateway second.
