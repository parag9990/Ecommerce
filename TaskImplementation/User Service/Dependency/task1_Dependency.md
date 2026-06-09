# User Service - Task 1 Dependency and Setup Guide

Generated from:

```text
SERVICE_NAME=User Service
TASK_FILE_NAME=task1.md
INPUT_PATH=TaskImplementation/User Service/task1.md
OUTPUT_FILE=task1_Dependency.md
```

Is document ka goal hai ki beginner developer project clone karke dependencies install kare, environment configure kare, MySQL setup kare, migration run kare, backend service start kare, aur common setup issues debug kar sake.

Important scope note:

- `TaskImplementation/User Service/task1.md` itself is a domain documentation task. Us file ko read karne ke liye runtime dependency nahi chahiye.
- Current repo me `backend/services/user-service/` ka runnable Go service, migrations, gRPC server, and generated proto code present hai. Setup instructions below us actual runnable service ke basis par diye gaye hain.
- Business logic, implementation files, and original task file ko modify nahi kiya gaya.

---

## 1. Project Tech Stack Analysis

### Analyzed Files

| File | Why it was checked |
|---|---|
| `TaskImplementation/User Service/task1.md` | Task 1 domain boundary, profile fields, ownership split |
| `backend/services/user-service/go.mod` | Go version and direct dependencies |
| `backend/services/user-service/go.sum` | Dependency checksum lock file |
| `backend/go.work` | Local multi-module workspace |
| `backend/services/user-service/internal/config/config.go` | Environment variables and validation |
| `backend/services/user-service/cmd/server/main.go` | Service startup, MySQL connection, gRPC server |
| `backend/services/user-service/migrations/001_create_user_tables.up.sql` | MySQL schema setup |
| `backend/services/user-service/migrations/001_create_user_tables.down.sql` | MySQL rollback setup |
| `proto/ecommerce/user/v1/user.proto` | gRPC API contract |
| `proto/buf.yaml` and `proto/buf.gen.yaml` | Protobuf generation workflow |
| `backend/shared/gen/go/` | Generated Go protobuf/gRPC module |
| `.gitignore` | Confirms `.env` files are ignored and `.env.example` is allowed |

### Main Technologies

| Technology | Required? | What it is | Why project uses it |
|---|---:|---|---|
| Go 1.24+ | Yes | Go ek compiled backend language hai jo fast APIs/services banane ke liye use hoti hai. | Backend service Go me implemented hai. |
| Go Modules | Yes | Go ka dependency system. `go.mod` dependencies list karta hai, `go.sum` checksums lock karta hai. | Stable dependency versions ke liye. |
| Go Workspace | Yes for local dev | `go.work` multiple Go modules ko ek local workspace me connect karta hai. | Service local generated proto module `backend/shared/gen/go` use karta hai. |
| MySQL 8.x | Yes at runtime | MySQL ek relational database hai jisme tables, indexes, constraints hote hain. | Profiles, addresses, seller profiles, and KYC metadata structured relational data hai. |
| gRPC | Yes | gRPC high-performance service-to-service communication framework hai. | User profile APIs internal services/Gateway ko expose karne ke liye. |
| Protocol Buffers | Yes | `.proto` typed API contract define karta hai. | Request/response types generated Go code me convert hote hain. |
| Buf | Optional unless proto changes | Proto lint/generate tool. | `user.proto` change karne ke baad generated Go code update karne ke liye. |
| `database/sql` | Yes | Go standard DB abstraction. | MySQL connection pool and query execution ke liye. |
| `github.com/go-sql-driver/mysql` | Yes | Go MySQL driver. | Go app ko MySQL DSN se connect karne ke liye. |
| `log/slog` | Yes | Go standard structured logging. | JSON logs stdout par print karne ke liye. |
| `go-sqlmock` | Test only | SQL tests ke liye mock DB library. | Repository tests real MySQL ke bina run karne ke liye. |
| Docker | Optional but recommended | Containers run karne ka tool. | Beginner local MySQL setup ko easy banata hai. |
| grpcurl | Optional dev tool | Terminal se gRPC APIs test karne ka CLI. | Service verify/debug karne ke liye useful. |
| Markdown | Yes for task docs | Lightweight documentation format. | Task implementation and dependency docs Markdown me hain. |
| Mermaid | Optional docs rendering | Markdown diagrams render karta hai. | Task 1 me architecture/ER/flow diagrams explain kiye gaye hain. |

### Not Used By Current User Service Startup

| Technology | Status |
|---|---|
| Node.js / npm / pnpm | Current backend service start karne ke liye required nahi. |
| Python / pip / virtualenv | Current backend service ke liye required nahi. |
| Redis | Current User Service code me Redis client/env usage nahi mila. |
| Kafka / RabbitMQ / NATS | Task 8 me events planned hain, but current service startup me MQ dependency nahi hai. |
| MongoDB | User Service MySQL use karta hai, MongoDB nahi. |
| Elasticsearch / Typesense | Search service concern ho sakta hai, User Service runtime dependency nahi. |
| Kubernetes | Local development ke liye required nahi; production deployment future concern hai. |
| S3 / MinIO | Task 1 KYC storage URL concept define karta hai; actual object storage integration current service startup me nahi mila. |

---

## 2. Language-Specific Dependency System: Go

### `go.mod`

`go.mod` Go project ka dependency manifest hai.

Simple Hinglish:

`go.mod` batata hai ki project ko kaunsi libraries chahiye, kaunsa Go version chahiye, aur local module replacement kaha se load karna hai.

Main file:

```text
backend/services/user-service/go.mod
```

Current direct dependencies:

| Dependency | Required? | Purpose |
|---|---:|---|
| `github.com/go-sql-driver/mysql v1.9.3` | Yes | MySQL driver |
| `google.golang.org/grpc v1.72.2` | Yes | gRPC server |
| `google.golang.org/protobuf v1.36.6` | Yes | Protobuf runtime |
| `github.com/parag/ecommerce/backend/shared/gen/go v0.0.0` | Yes | Local generated protobuf/gRPC Go code |
| `github.com/DATA-DOG/go-sqlmock v1.5.2` | Test only | SQL repository unit tests |

Local replace:

```go
replace github.com/parag/ecommerce/backend/shared/gen/go => ../../shared/gen/go
```

Meaning:

The service does not download this generated module from internet. Ye local folder `backend/shared/gen/go` se load hota hai. Isliye repo ka folder structure intact rehna chahiye.

### `go.sum`

`go.sum` dependency checksums store karta hai.

Beginner rules:

- Manually edit mat karo.
- `go mod tidy`, `go mod download`, or `go test` automatically update kar sakte hain.
- Is file ko commit karna chahiye because checksums reproducible builds me help karte hain.

### `go.work`

Workspace file:

```text
backend/go.work
```

Current workspace modules:

```text
./services/user-service
./shared/gen/go
```

Meaning:

Local development me Go ko pata hai ki User Service aur generated proto module same repo ke andar available hain.

### Common Go Commands

Run these from service folder:

```bash
cd backend/services/user-service
go mod download
go mod tidy
go test ./...
go build ./...
go run ./cmd/server
```

Run from workspace folder:

```bash
cd backend
go test ./services/user-service/...
```

### Why Dependencies Fail

| Error | Common cause | Fix |
|---|---|---|
| `go: go.mod file not found` | Wrong directory | `cd backend/services/user-service` or `cd backend` |
| `module requires go >= 1.24` | Old Go installed | Install Go 1.24+ |
| `missing go.sum entry` | Dependency checksum missing | Run `go mod tidy` |
| `cannot find module .../shared/gen/go` | Repo structure missing or local replace path broken | Clone full repo; do not move service folder alone |
| `dial tcp ... proxy.golang.org` fails | Network/proxy issue | Retry later, set `GOPROXY`, or use cached deps |
| Version mismatch after proto change | Generated code and proto runtime mismatch | Run `buf generate`, then `go mod tidy` |
| Tests fail with DB connection | Running integration/service startup without env | Unit tests mostly mock DB; service run needs MySQL DSN |

### Useful Go Environment Debug Commands

```bash
go version
go env GOPATH
go env GOPROXY
go env GOWORK
go list -m all
```

---

## 3. Database Analysis

### Detected Database: MySQL

#### A. What It Is

MySQL ek relational database hai. Data tables ke form me store hota hai: rows, columns, primary keys, unique keys, indexes, and foreign keys.

Simple example:

- `users` table profile data store karta hai.
- `user_addresses` table user ke multiple addresses store karta hai.
- Foreign key ensure karta hai ki address kisi existing user se linked ho.

#### B. Why This Project Uses MySQL

Task 1 ne profile domain finalize kiya. Current migration us domain ko MySQL tables me map karta hai.

Detected database:

```text
user_db
```

Detected tables:

| Table | Purpose |
|---|---|
| `users` | Base buyer/seller/admin profile |
| `user_addresses` | Shipping/billing address book |
| `seller_profiles` | Seller business profile and review status |
| `seller_kyc_documents` | KYC metadata, not actual binary files |

#### C. Required Or Optional

MySQL required hai when running the backend service.

Why:

- `cmd/server/main.go` startup me DB open karta hai.
- Service startup `PingContext` se MySQL ping karta hai.
- Agar DB unreachable hai, service start fail karegi.

Task 1 document read karne ke liye MySQL required nahi hai.

#### D. Local Installation

Windows:

```powershell
winget install Oracle.MySQL
mysql --version
```

Alternative:

- MySQL Installer for Windows install karo.
- MySQL Server and MySQL Shell/Client select karo.
- Root password yaad rakho.

macOS:

```bash
brew install mysql
brew services start mysql
mysql --version
```

Ubuntu/Debian Linux:

```bash
sudo apt update
sudo apt install -y mysql-server mysql-client
sudo systemctl enable --now mysql
mysql --version
```

Verify service:

```bash
mysqladmin ping -h 127.0.0.1 -P 3306 -u root -p
```

#### E. Docker Setup

Docker local setup beginners ke liye recommended hai because MySQL install/config manual steps kam ho jate hain.

```bash
docker run --name ecommerce-user-mysql \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=user_db \
  -e MYSQL_USER=ecommerce_user \
  -e MYSQL_PASSWORD=change_me_local_password \
  -p 3306:3306 \
  -v ecommerce_user_mysql_data:/var/lib/mysql \
  -d mysql:8
```

If port `3306` busy hai:

```bash
docker run --name ecommerce-user-mysql \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=user_db \
  -e MYSQL_USER=ecommerce_user \
  -e MYSQL_PASSWORD=change_me_local_password \
  -p 3307:3306 \
  -v ecommerce_user_mysql_data:/var/lib/mysql \
  -d mysql:8
```

Then DSN me `3307` use karo.

#### F. Docker Compose Example

No checked-in `docker-compose.yml` was found. Ye beginner local example hai:

```yaml
services:
  user-mysql:
    image: mysql:8
    container_name: ecommerce-user-mysql
    environment:
      MYSQL_ROOT_PASSWORD: root
      MYSQL_DATABASE: user_db
      MYSQL_USER: ecommerce_user
      MYSQL_PASSWORD: change_me_local_password
    ports:
      - "3306:3306"
    volumes:
      - ecommerce_user_mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "127.0.0.1", "-uroot", "-proot"]
      interval: 10s
      timeout: 5s
      retries: 10

volumes:
  ecommerce_user_mysql_data:
```

Start:

```bash
docker compose up -d user-mysql
```

Stop:

```bash
docker compose down
```

Stop and delete local DB data:

```bash
docker compose down -v
```

Warning: `down -v` data delete karta hai. Sirf disposable local DB ke liye use karo.

#### G. Start Commands

Native Linux MySQL:

```bash
sudo systemctl start mysql
```

macOS Homebrew:

```bash
brew services start mysql
```

Docker:

```bash
docker start ecommerce-user-mysql
```

#### H. Verify Running

```bash
docker ps --filter name=ecommerce-user-mysql
mysqladmin ping -h 127.0.0.1 -P 3306 -u root -p
mysql -h 127.0.0.1 -P 3306 -u root -p -e "SHOW DATABASES;"
```

Verify tables after migration:

```bash
mysql -h 127.0.0.1 -P 3306 -u ecommerce_user -p user_db -e "SHOW TABLES;"
```

Expected:

```text
seller_kyc_documents
seller_profiles
user_addresses
users
```

#### I. Default Port

| Service | Default port |
|---|---:|
| MySQL | `3306` |
| Alternate local MySQL mapping | `3307` |

#### J. Connection String Format

Go MySQL driver DSN format:

```env
USER_SERVICE_DATABASE_DSN=ecommerce_user:change_me_local_password@tcp(127.0.0.1:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC
```

Important parts:

| Part | Meaning |
|---|---|
| `ecommerce_user` | DB username |
| `change_me_local_password` | DB password |
| `127.0.0.1` | DB host |
| `3306` | DB port |
| `user_db` | Database name |
| `parseTime=true` | Go timestamps ko correctly scan karne ke liye important |
| `charset=utf8mb4` | Unicode support |
| `loc=UTC` | UTC time handling |

#### K. Where To Place Credentials

Local development:

```text
backend/services/user-service/.env
```

Security rules:

- `.env` me real password ho sakta hai, commit mat karo.
- `.gitignore` already `.env` and nested `.env` files ignore karta hai.
- Sanitized `.env.example` add karna recommended hai.
- Production me secrets manager, deployment env vars, or CI/CD secret store use karo.

#### L. Run Migration

Migration file:

```text
backend/services/user-service/migrations/001_create_user_tables.up.sql
```

Apply:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p < backend/services/user-service/migrations/001_create_user_tables.up.sql
```

Rollback file:

```text
backend/services/user-service/migrations/001_create_user_tables.down.sql
```

Rollback only in disposable local/dev DB:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p < backend/services/user-service/migrations/001_create_user_tables.down.sql
```

Current limitation:

- Plain SQL migration files exist.
- No versioned migration runner such as `golang-migrate` or `goose` is wired in current service startup.
- Developer ko migration manually run karni hogi before running real service flows.

---

## 4. Environment Variables

### Complete `.env` Example

Create this file for local development:

```text
backend/services/user-service/.env
```

Example:

```env
# Required: MySQL DSN for User Service runtime.
USER_SERVICE_DATABASE_DSN=ecommerce_user:change_me_local_password@tcp(127.0.0.1:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC

# Optional fallback if USER_SERVICE_DATABASE_DSN is blank.
MYSQL_DSN=ecommerce_user:change_me_local_password@tcp(127.0.0.1:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC

# gRPC server config.
USER_SERVICE_GRPC_ADDRESS=:50052
USER_SERVICE_GRPC_REFLECTION=true
USER_SERVICE_SHUTDOWN_TIMEOUT=10s

# Database pool config.
USER_SERVICE_DB_MAX_OPEN_CONNS=25
USER_SERVICE_DB_MAX_IDLE_CONNS=25
USER_SERVICE_DB_CONN_MAX_LIFETIME=5m
USER_SERVICE_DB_PING_TIMEOUT=5s

# Logging.
USER_SERVICE_LOG_LEVEL=debug
```

### How Project Loads Environment Variables

Code uses `os.Getenv` in:

```text
backend/services/user-service/internal/config/config.go
```

Important:

The code does not automatically parse `.env`. Shell me variables export karne padenge.

Use:

```bash
cd backend/services/user-service
set -a
. ./.env
set +a
go run ./cmd/server
```

### Environment Variable Table

| Variable | Required? | Example | Purpose | Security notes |
|---|---:|---|---|---|
| `USER_SERVICE_DATABASE_DSN` | Yes | `ecommerce_user:...@tcp(127.0.0.1:3306)/user_db?...` | Main MySQL connection string | Contains password; do not log/commit |
| `MYSQL_DSN` | Conditional fallback | Same DSN format | Used only if main DSN blank | Contains password; prefer main var for clarity |
| `USER_SERVICE_GRPC_ADDRESS` | Optional | `:50052` | gRPC listen address/port | Bind internally in production |
| `USER_SERVICE_GRPC_REFLECTION` | Optional | `true` local, `false` prod | Enables grpcurl reflection | Disable or restrict in public/prod networks |
| `USER_SERVICE_SHUTDOWN_TIMEOUT` | Optional | `10s` | Graceful shutdown wait time | Must be positive |
| `USER_SERVICE_DB_MAX_OPEN_CONNS` | Optional | `25` | Max open DB connections | Too high can overload DB |
| `USER_SERVICE_DB_MAX_IDLE_CONNS` | Optional | `25` | Max idle DB connections | Cannot be negative |
| `USER_SERVICE_DB_CONN_MAX_LIFETIME` | Optional | `5m` | DB connection recycle duration | Helps avoid stale long-lived connections |
| `USER_SERVICE_DB_PING_TIMEOUT` | Optional | `5s` | Startup DB ping timeout | Must be positive |
| `USER_SERVICE_LOG_LEVEL` | Optional | `debug` local, `info` prod | Log verbosity | Avoid debug logs with PII in production |

### Common `.env` Mistakes

| Mistake | Symptom | Fix |
|---|---|---|
| `.env` created but not sourced | `USER_SERVICE_DATABASE_DSN is required` | Use `set -a; . ./.env; set +a` |
| Wrong DB port | `connection refused` | Match Docker/native port in DSN |
| Missing `parseTime=true` | Timestamp scan errors | Add `parseTime=true` |
| Password contains special characters | DSN parse error | URL-escape password or use simple local password |
| Quotes included incorrectly | Auth failure or bad DSN | Prefer no quotes in `.env`, or source with shell-compatible syntax |
| Using root DB user for app | Works locally but unsafe | Create app user with limited grants |

---

## 5. External Services Analysis

### MySQL

| Item | Detail |
|---|---|
| What it is | Relational DB for structured profile data |
| Why used | User, address, seller, KYC metadata persistence |
| Mandatory? | Yes for running backend |
| Credentials | `USER_SERVICE_DATABASE_DSN` or `MYSQL_DSN` |
| Health check | `mysqladmin ping -h 127.0.0.1 -P 3306 -u root -p` |
| Docker | `mysql:8` container recommended locally |
| Common issue | Wrong DSN user/password/port |

### gRPC

| Item | Detail |
|---|---|
| What it is | Internal API communication framework |
| Why used | Auth/API Gateway/internal services call User Service |
| Mandatory? | Yes for current backend server |
| Port | Default `50052` |
| Health check | `grpcurl -plaintext localhost:50052 list` when reflection enabled |
| Credentials | No TLS/auth config found for current local server |
| Common issue | Reflection disabled or service not running |

Install grpcurl:

```bash
# macOS
brew install grpcurl

# Go install alternative
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

Verify:

```bash
grpcurl -plaintext localhost:50052 list
grpcurl -plaintext localhost:50052 describe ecommerce.user.v1.UserService
```

### Protobuf and Buf

| Item | Detail |
|---|---|
| What it is | API contract and code generation tooling |
| Why used | `user.proto` generates typed Go gRPC code |
| Mandatory? | Runtime uses generated code; Buf only needed when proto changes |
| Config files | `proto/buf.yaml`, `proto/buf.gen.yaml` |
| Generated output | `backend/shared/gen/go/` |
| Common issue | Proto changed but generated code stale |

Generate after proto changes:

```bash
cd proto
buf generate
```

### Docker

| Item | Detail |
|---|---|
| What it is | Container runtime |
| Why used | Optional local MySQL setup |
| Mandatory? | No, native MySQL works |
| Dockerfile | Not found for User Service |
| docker-compose | Not found in repo |
| Health check | `docker ps`, `docker logs ecommerce-user-mysql` |
| Common issue | Port already allocated or stale volume password |

### API Gateway, Auth Service, Superadmin Service

Task 1 defines service boundaries:

| Service | Current relation | Required for local User Service startup? |
|---|---|---:|
| Auth Service | Owns password, OTP, JWT, roles; calls `CreateUser` conceptually | No |
| API Gateway | Browser REST should go through Gateway, then gRPC to User Service | No |
| Superadmin Service | Seller/admin workflows and audit ownership | No |

For isolated local backend startup, only User Service + MySQL are required. For complete platform flow, Gateway/Auth/Superadmin will be needed later.

### Object Storage / S3 / MinIO

Task 1 says actual KYC files should not live in MySQL. DB stores only `storage_url`.

Current status:

- No S3/MinIO client or env variables found in User Service startup.
- Object storage is a future integration, not a required Task 1 runtime dependency.

### Redis, Kafka, RabbitMQ, NATS

Current status:

| Service | Required now? | Why |
|---|---:|---|
| Redis | No | No Redis client/env usage in current User Service |
| Kafka | No | Events are planned in later docs, not wired here |
| RabbitMQ | No | No queue publisher/consumer in current service |
| NATS | No | No NATS usage found |

Do not add fake env variables for these services until implementation actually uses them.

---

## 6. Ports and Networking

| Service | Port | Purpose | Required? |
|---|---:|---|---:|
| User Service gRPC | `50052` | Internal gRPC API | Yes for backend |
| MySQL | `3306` | Database connection | Yes |
| MySQL alternate | `3307` | Local fallback if `3306` busy | Optional |
| API Gateway HTTP | `8080` | Browser REST entrypoint, future/full platform | Not for isolated service |
| Redis | `6379` | Not used by current User Service | No |
| Kafka | `9092` | Future events only | No |
| RabbitMQ | `5672` | Not used by current User Service | No |

### Port Conflicts

Check open ports:

```bash
ss -ltnp | grep 50052
ss -ltnp | grep 3306
```

If gRPC port busy:

```bash
export USER_SERVICE_GRPC_ADDRESS=':50053'
```

Then verify with:

```bash
grpcurl -plaintext localhost:50053 list
```

If MySQL port busy:

```bash
docker run --name ecommerce-user-mysql \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=user_db \
  -e MYSQL_USER=ecommerce_user \
  -e MYSQL_PASSWORD=change_me_local_password \
  -p 3307:3306 \
  -v ecommerce_user_mysql_data:/var/lib/mysql \
  -d mysql:8
```

Update DSN:

```env
USER_SERVICE_DATABASE_DSN=ecommerce_user:change_me_local_password@tcp(127.0.0.1:3307)/user_db?parseTime=true&charset=utf8mb4&loc=UTC
```

### Docker Network Notes

From host to Docker MySQL:

```text
127.0.0.1:3306
```

From another container to MySQL in same compose network:

```text
user-mysql:3306
```

Common beginner confusion:

- Host machine uses published port: `127.0.0.1:3306`.
- Container-to-container uses service name: `user-mysql:3306`.
- Do not use `localhost` inside one container to reach another container.

---

## 7. Docker and DevOps Setup

### Current Repo Status

| Item | Status |
|---|---|
| User Service Dockerfile | Not found |
| Project `docker-compose.yml` | Not found |
| MySQL Docker usage | Recommended local option |
| Docker volumes | Needed for persistent MySQL data |
| Docker networks | Needed if service is containerized later |
| Restart policies | Future compose/deployment concern |
| Health checks | MySQL example provided; app health endpoint not found |

### Recommended Beginner Approach

Use Docker only for MySQL, and run the Go backend directly:

```bash
docker run --name ecommerce-user-mysql \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=user_db \
  -e MYSQL_USER=ecommerce_user \
  -e MYSQL_PASSWORD=change_me_local_password \
  -p 3306:3306 \
  -v ecommerce_user_mysql_data:/var/lib/mysql \
  -d mysql:8
```

Then run service directly:

```bash
cd backend/services/user-service
set -a
. ./.env
set +a
go run ./cmd/server
```

### Useful Docker Commands

```bash
docker ps
docker logs ecommerce-user-mysql
docker stop ecommerce-user-mysql
docker start ecommerce-user-mysql
docker rm ecommerce-user-mysql
docker volume ls
```

Docker Compose commands if you create a local compose file:

```bash
docker compose up -d
docker compose down
docker compose logs
docker compose ps
```

### Local vs Docker Setup

| Approach | Pros | Cons |
|---|---|---|
| Native MySQL | Fast, simple if already installed | OS-specific install/config issues |
| Docker MySQL | Clean, repeatable, easy reset | Needs Docker daemon and port mapping understanding |
| Full Docker app stack | Production-like | Not available yet because Dockerfile/compose missing |

For beginners, Docker MySQL + direct `go run` is best currently.

---

## 8. Project Run Instructions

### Step 1: Clone Repository

```bash
git clone <repository-url>
cd Ecommerce
```

### Step 2: Go To Service Directory

```bash
cd backend/services/user-service
```

### Step 3: Install/Download Go Dependencies

```bash
go mod download
go mod tidy
```

### Step 4: Setup MySQL

Docker option:

```bash
docker run --name ecommerce-user-mysql \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=user_db \
  -e MYSQL_USER=ecommerce_user \
  -e MYSQL_PASSWORD=change_me_local_password \
  -p 3306:3306 \
  -v ecommerce_user_mysql_data:/var/lib/mysql \
  -d mysql:8
```

Wait until MySQL is ready:

```bash
docker logs ecommerce-user-mysql
mysqladmin ping -h 127.0.0.1 -P 3306 -u root -p
```

### Step 5: Run Migration

From repository root:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p < backend/services/user-service/migrations/001_create_user_tables.up.sql
```

Verify:

```bash
mysql -h 127.0.0.1 -P 3306 -u ecommerce_user -p user_db -e "SHOW TABLES;"
```

### Step 6: Create `.env`

Create:

```text
backend/services/user-service/.env
```

Use:

```env
USER_SERVICE_DATABASE_DSN=ecommerce_user:change_me_local_password@tcp(127.0.0.1:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC
USER_SERVICE_GRPC_ADDRESS=:50052
USER_SERVICE_GRPC_REFLECTION=true
USER_SERVICE_SHUTDOWN_TIMEOUT=10s
USER_SERVICE_DB_MAX_OPEN_CONNS=25
USER_SERVICE_DB_MAX_IDLE_CONNS=25
USER_SERVICE_DB_CONN_MAX_LIFETIME=5m
USER_SERVICE_DB_PING_TIMEOUT=5s
USER_SERVICE_LOG_LEVEL=debug
```

### Step 7: Start Backend Service

```bash
cd backend/services/user-service
set -a
. ./.env
set +a
go run ./cmd/server
```

Expected log messages:

```text
user_service_starting
user_service_grpc_listening
```

### Step 8: Verify gRPC APIs

In another terminal:

```bash
grpcurl -plaintext localhost:50052 list
grpcurl -plaintext localhost:50052 describe ecommerce.user.v1.UserService
```

Current proto methods:

| Method | Purpose |
|---|---|
| `CreateUser` | Create base profile after Auth signup |
| `GetUser` | Fetch profile |
| `UpdateUserProfile` | Update profile fields |
| `GetSellerProfile` | Fetch seller profile |

### Step 9: Run Tests

```bash
cd backend/services/user-service
go test ./...
```

---

## 9. Common Errors and Fixes

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `USER_SERVICE_DATABASE_DSN is required` | `.env` not exported or variable missing | Run `set -a; . ./.env; set +a` | Add startup notes to README |
| `ping mysql: connect: connection refused` | MySQL stopped or wrong host/port | Start MySQL; verify `3306` or `3307` | Keep DSN aligned with Docker port |
| `Unknown database 'user_db'` | Migration/DB creation not run | Run up migration | Run migration before service |
| `Access denied for user` | Wrong username/password/grants | Recreate user grants or fix DSN | Use one source of truth for DB credentials |
| `port is already allocated` | Docker MySQL port busy | Map `3307:3306` and update DSN | Check `ss -ltnp` before run |
| `listen tcp :50052: bind: address already in use` | gRPC port busy | Set `USER_SERVICE_GRPC_ADDRESS=:50053` | Reserve local ports in docs |
| `grpcurl: connection refused` | Service not running or wrong port | Start service and use correct port | Watch startup logs |
| `server does not support reflection` | Reflection disabled | Set `USER_SERVICE_GRPC_REFLECTION=true` locally | Keep reflection local-only |
| `go: module requires go >= 1.24` | Old Go version | Install Go 1.24+ | Verify `go version` first |
| `missing go.sum entry` | Dependencies not downloaded/tidied | Run `go mod tidy` | Commit updated `go.sum` |
| `cannot find shared/gen/go` | Running service folder without full repo | Clone full repo | Do not copy module alone |
| `Docker daemon not running` | Docker Desktop/service stopped | Start Docker | Verify `docker ps` before setup |
| Migration foreign key error on rollback | Dropping parent before child | Use provided down migration | Do rollback only with versioned scripts |
| Timestamp scan error | DSN missing `parseTime=true` | Add `parseTime=true` | Keep DSN example exact |
| Invalid bool/duration ignored | Env value parse failed | Use `true`, `false`, `10s`, `5m` | Copy validated examples |
| Permission denied on Docker | Linux user not in docker group | Add user to docker group and re-login | Configure Docker after install |

---

## 10. Security and Configuration Audit

| Area | Finding | Risk | Suggested fix |
|---|---|---|---|
| `.env` file | Local `.env` exists and is ignored by Git | Secrets can leak if copied manually | Keep ignored; add sanitized `.env.example` |
| `.env.example` | Not found | Beginners may not know required variables | Add example with fake values |
| DB password in DSN | DSN contains password | Logs/screenshots can expose secrets | Never log full DSN; mask password |
| MySQL root usage | Migration examples use root | Unsafe for runtime | Use root/admin only for migration; app user for service |
| Runtime DB grants | App user should not need DDL in production | Excess permissions increase blast radius | Grant SELECT/INSERT/UPDATE/DELETE only |
| gRPC reflection | Defaults to true | Useful locally, risky if public | Set `USER_SERVICE_GRPC_REFLECTION=false` in prod |
| gRPC transport security | No TLS config found | Internal plaintext only | Use private network or add TLS/mTLS later |
| Health check | No gRPC health service found | Harder orchestration/debugging | Add standard `grpc_health_v1` health service |
| Dockerfile | Not found | No containerized app runtime | Add Dockerfile in DevOps task |
| docker-compose | Not found | Beginners need manual DB steps | Add local compose with MySQL healthcheck |
| Migration runner | Plain SQL only | No version tracking | Add `golang-migrate` or `goose` later |
| Object storage | KYC URL concept exists, storage integration absent | Future uploads may be insecure if rushed | Use private bucket, signed URLs, no public KYC files |
| PII logging | User domain includes email, phone, address, KYC URL | Privacy risk | Mask PII and never log tokens/passwords |
| Auth boundaries | Password/OTP/JWT intentionally not in User Service | Good separation | Keep Auth Service as source of truth |
| Redis/Kafka env | Not present and not needed now | Fake env docs confuse setup | Add only when implementation uses them |

Hardcoded credential note:

- The repository has local `.env` values in the working tree, but `.gitignore` excludes them.
- Dependency docs should use `change_me_local_password`, not real local passwords.
- Production secrets should come from a secret manager or deployment environment.

---

## 11. Best Practices

- Never commit `.env`.
- Create and commit `.env.example` with fake values.
- Use strong passwords, even in shared dev environments.
- Do not use MySQL root user for app runtime.
- Keep migration/admin user separate from runtime app user.
- Always include `parseTime=true` in Go MySQL DSN.
- Use Docker volumes for local MySQL persistence.
- Use `docker compose down -v` only when you intentionally want to delete local DB data.
- Disable gRPC reflection in production or expose it only inside private networks.
- Do not log passwords, OTP, JWT, refresh tokens, full DSN, or KYC document URLs.
- Keep User Service profile ownership separate from Auth Service credential ownership.
- Run `go mod tidy` after dependency changes.
- Run `go test ./...` before pushing backend changes.
- Use versioned migration tooling before adding many migrations.
- Document every required service when Redis/Kafka/object storage is actually implemented.

---

## 12. Final Checklist

| Check | Done |
|---|---|
| Required language runtime Go 1.24+ installed | [ ] |
| Git installed and repo cloned | [ ] |
| Full repo structure available, including `backend/shared/gen/go` | [ ] |
| Go dependencies downloaded with `go mod download` | [ ] |
| MySQL installed or Docker MySQL running | [ ] |
| MySQL port confirmed (`3306` or alternate `3307`) | [ ] |
| `user_db` created | [ ] |
| Migration `001_create_user_tables.up.sql` applied | [ ] |
| Tables verified with `SHOW TABLES` | [ ] |
| `backend/services/user-service/.env` created locally | [ ] |
| `USER_SERVICE_DATABASE_DSN` points to correct DB host/port/user/password | [ ] |
| `.env` sourced/exported before running service | [ ] |
| Backend started with `go run ./cmd/server` | [ ] |
| Startup logs show `user_service_grpc_listening` | [ ] |
| gRPC verified with `grpcurl` | [ ] |
| `go test ./...` passes | [ ] |
| Common errors section reviewed | [ ] |
| `.env.example` planned or added with fake values | [ ] |
| Production notes reviewed: no root DB user, reflection disabled, secrets not logged | [ ] |

---

## Quick Start Summary

```bash
# 1. Start MySQL with Docker
docker run --name ecommerce-user-mysql \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=user_db \
  -e MYSQL_USER=ecommerce_user \
  -e MYSQL_PASSWORD=change_me_local_password \
  -p 3306:3306 \
  -v ecommerce_user_mysql_data:/var/lib/mysql \
  -d mysql:8

# 2. Apply migration from repo root
mysql -h 127.0.0.1 -P 3306 -u root -p < backend/services/user-service/migrations/001_create_user_tables.up.sql

# 3. Start service
cd backend/services/user-service
set -a
. ./.env
set +a
go run ./cmd/server

# 4. Verify
grpcurl -plaintext localhost:50052 list
```

Final note:

Task 1 ka core output domain boundary hai: User Service profile data own karta hai; password, OTP, JWT, roles Auth Service me rahenge; admin permissions/audit Superadmin boundary me rahenge; actual KYC binary files object storage me rahenge. Runtime setup ke liye current service ko Go, MySQL, gRPC/proto generated code, and environment variables chahiye.
