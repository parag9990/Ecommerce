# Project Dependency & Setup Guide

This guide is generated for:

```text
TaskImplementation/User Service/task1.md
```

Output file:

```text
TaskImplementation/User Service/task1_Dependency.md
```

Is document ka goal hai ki ek beginner developer bhi User Service ko local machine par setup, configure, migrate, run, and debug kar sake. Ye business logic rewrite nahi karta. Ye sirf dependencies, environment, database, Docker, DevOps, and onboarding setup explain karta hai.

---

## 1. Project Overview

User Service e-commerce platform ka profile service hai.

Simple Hinglish me:

User Service ka kaam login/password manage karna nahi hai. Ye service buyer, seller, admin base profile, addresses, seller profile, aur KYC metadata manage karti hai. Authentication Auth Service karega, lekin profile data User Service ke MySQL database me rahega.

Analyzed implementation files:

| File | Purpose |
|---|---|
| `TaskImplementation/User Service/task1.md` | Domain boundary and task implementation guide |
| `backend/services/user-service/go.mod` | Go dependencies |
| `backend/go.work` | Go workspace setup |
| `backend/services/user-service/internal/config/config.go` | Runtime environment variables |
| `backend/services/user-service/cmd/server/main.go` | gRPC server startup and MySQL connection |
| `backend/services/user-service/migrations/001_create_user_tables.up.sql` | MySQL schema migration |
| `proto/ecommerce/user/v1/user.proto` | gRPC contract |
| `backend/shared/gen/go/...` | Generated Go protobuf/gRPC code |

Current implementation status:

| Area | Status |
|---|---|
| Language | Go |
| Runtime service | gRPC server |
| Database | MySQL |
| REST API | Not directly in User Service; expected via API Gateway later |
| Redis | Not required by current User Service code |
| Kafka/RabbitMQ | Not required by current Task 1/current service startup |
| Docker files | Not present yet in repo |
| Migrations | Plain SQL files exist |
| `.env.example` | Not present yet |

---

## 2. Tech Stack

### Main Technologies

| Technology | Required? | What it is | Why this project uses it |
|---|---:|---|---|
| Go 1.24+ | Yes | Go ek compiled backend language hai jo fast, simple, and production-friendly services banane ke liye use hoti hai. | User Service backend Go me implemented hai. |
| Go Modules | Yes | Go ka dependency management system. `go.mod` dependencies list karta hai and `go.sum` checksum lock karta hai. | Libraries ko version ke saath manage karne ke liye. |
| Go Workspace | Yes | `go.work` multiple Go modules ko ek workspace me connect karta hai. | `user-service` local generated protobuf module ko use karta hai. |
| gRPC | Yes | gRPC ek high-performance internal service communication framework hai. | User Service internal APIs expose karta hai for Auth/Gateway/other services. |
| Protocol Buffers | Yes | Proto files strongly typed request/response contract define karte hain. | User Service ke gRPC messages generated code se type-safe bante hain. |
| MySQL | Yes | MySQL ek relational database hai jisme data tables ke form me store hota hai. | Users, addresses, seller profiles, and KYC metadata structured data hai. |
| `database/sql` | Yes | Go standard DB abstraction package. | MySQL queries and connection pooling ke liye. |
| `github.com/go-sql-driver/mysql` | Yes | Go MySQL driver. | Go app ko MySQL se connect karne ke liye. |
| `log/slog` | Yes | Go standard structured logging package. | JSON logs stdout par print karne ke liye. |
| `go-sqlmock` | Test only | SQL repository tests ke liye mock database library. | Unit tests me real MySQL ke bina SQL behavior test karne ke liye. |
| Docker | Optional but recommended | Containers run karne ka tool. | Beginner setup me MySQL ko quickly run karne ke liye best option. |
| grpcurl | Optional dev tool | gRPC APIs ko terminal se test karne ka CLI. | Browser/curl gRPC directly nahi bolta, grpcurl debugging easy banata hai. |
| buf CLI | Optional unless proto changes | Protobuf generation/linting tool. | Proto update ke baad generated Go code banane ke liye. |

### Not Required For Current User Service Startup

| Technology | Current status |
|---|---|
| Node.js / npm / pnpm | User Service backend run karne ke liye required nahi. Frontend/future tools ke liye docs me mentioned hai. |
| Python / pip / venv | Current User Service ke liye required nahi. |
| Redis | Current User Service code me Redis connection nahi hai. Future cache/rate-limit/session use case ho sakta hai. |
| Kafka / RabbitMQ | Current Task 1/current startup me event publisher nahi wired. Future Task 8 me user events ke liye planned hai. |
| MongoDB | User Service use nahi karta. Other services jaise Product/Cart/Wishlist use kar sakte hain. |
| Typesense | User Service use nahi karta. Search Service ke liye planned hai. |
| Kubernetes | Local development ke liye required nahi. Production deployment ke liye planned. |

---

## 3. Required Software

### Minimum Required For This User Service

| Software | Version | Required? | Verify command |
|---|---|---:|---|
| Git | Any modern version | Yes | `git --version` |
| Go | 1.24 or higher | Yes | `go version` |
| MySQL | 8.x recommended | Yes | `mysql --version` |
| Docker | Latest stable | Optional but recommended | `docker --version` |
| Docker Compose | Compose v2 | Optional but recommended | `docker compose version` |
| grpcurl | Latest | Optional | `grpcurl -version` |
| buf | Latest | Optional unless proto regeneration needed | `buf --version` |

### Install Go

Go is mandatory.

Windows:

```powershell
winget install GoLang.Go
go version
```

Alternative: Download installer from the official Go website and restart terminal.

macOS:

```bash
brew install go
go version
```

Ubuntu/Debian Linux:

```bash
sudo apt update
sudo apt install -y golang-go
go version
```

Note: Ubuntu package manager kabhi-kabhi old Go version install karta hai. Agar `go version` 1.24 se lower aaye, official Go tarball ya version manager use karo.

### Install Docker

Docker optional hai, but beginners ke liye MySQL setup Docker se easiest hota hai.

Windows/macOS:

```text
Install Docker Desktop.
Start Docker Desktop.
Run: docker --version
Run: docker compose version
```

Ubuntu/Debian Linux:

```bash
sudo apt update
sudo apt install -y docker.io docker-compose-plugin
sudo systemctl enable --now docker
docker --version
docker compose version
```

If Docker command permission denied aaye:

```bash
sudo usermod -aG docker "$USER"
```

Then logout/login again.

---

## 4. Dependency Management

### This Is A Go Project

User Service dependencies yahan defined hain:

```text
backend/services/user-service/go.mod
backend/services/user-service/go.sum
backend/go.work
backend/go.work.sum
backend/shared/gen/go/go.mod
backend/shared/gen/go/go.sum
```

### `go.mod`

`go.mod` project ka dependency manifest hai.

Simple Hinglish:

`go.mod` bataata hai ki service ko kaunsi Go libraries chahiye aur kis version me chahiye.

Current direct dependencies:

| Dependency | Required? | Purpose |
|---|---:|---|
| `github.com/go-sql-driver/mysql v1.9.3` | Yes | MySQL database driver |
| `google.golang.org/grpc v1.72.2` | Yes | gRPC server |
| `google.golang.org/protobuf v1.36.6` | Yes | Generated proto message support |
| `github.com/parag/ecommerce/backend/shared/gen/go v0.0.0` | Yes | Local generated protobuf/gRPC Go code |
| `github.com/DATA-DOG/go-sqlmock v1.5.2` | Test only | SQL unit testing |

Important local replace:

```go
replace github.com/parag/ecommerce/backend/shared/gen/go => ../../shared/gen/go
```

Meaning:

User Service remote package download nahi karega. Ye local folder `backend/shared/gen/go` use karega. Isliye repo ka full structure intact hona chahiye.

### `go.sum`

`go.sum` dependency checksums store karta hai.

Beginner rule:

- Manually edit mat karo.
- `go mod tidy` ya `go mod download` automatically update karega.
- Git me commit karna chahiye.

### `go.work`

Workspace file:

```text
backend/go.work
```

Current content:

```go
go 1.24

use (
    ./services/user-service
    ./shared/gen/go
)
```

Meaning:

Backend folder ke andar Go ko pata hai ki `user-service` and generated proto module same workspace ka part hain.

Recommended command location:

```bash
cd backend/services/user-service
```

or:

```bash
cd backend
```

### Install/Download Go Dependencies

From service folder:

```bash
cd backend/services/user-service
go mod download
```

Clean dependency graph:

```bash
go mod tidy
```

Run tests:

```bash
go test ./...
```

Build binary:

```bash
go build -o ./bin/user-service ./cmd/server
```

Run service:

```bash
go run ./cmd/server
```

### Common Go Dependency Issues

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `go.mod requires go >= 1.24` | Installed Go version old hai. | Install Go 1.24+. | `go version` setup ke start me check karo. |
| `module ... backend/shared/gen/go not found` | Repo partial clone hai ya `backend/shared/gen/go` missing hai. | Full repo clone karo, folder check karo. | `git status` and `ls backend/shared/gen/go` verify karo. |
| `missing go.sum entry` | Dependency checksum missing. | `go mod tidy` run karo. | `go.sum` commit karo. |
| `no required module provides package` | Import exists but module missing hai. | `go mod tidy` ya correct import path. | Local package paths carefully use karo. |
| `go mod download` network fail | Internet/proxy issue. | Network fix, proxy configure, or dependency cache use karo. | Corporate proxy docs maintain karo. |
| `GONOSUMDB`/checksum issue | Private module ya checksum DB issue. | Private module settings configure karo. | Team Go env setup document karo. |

Useful debug commands:

```bash
go env
go env GOPROXY
go env GOMOD
go list -m all
```

---

## 5. Database Setup

### Detected Database: MySQL

MySQL required hai.

#### A. What It Is

MySQL ek relational database hai. Isme data tables, rows, columns, indexes, and foreign keys ke form me store hota hai.

Simple Hinglish:

MySQL ek structured data store hai. User profile, address, seller profile jaise fixed fields ke liye MySQL perfect fit hai.

#### B. Why This Project Uses It

User Service data relational hai:

- One user has many addresses.
- One user may have one seller profile.
- One seller profile has many KYC documents.
- Unique constraints chahiye for `user_id`, `auth_account_id`, `email`, `seller_id`.
- Foreign key constraints data consistency maintain karte hain.

#### C. Required Or Optional

Required.

Without MySQL, User Service startup fail karega because startup par DB ping hota hai.

Expected startup failure if DB/env missing:

```text
USER_SERVICE_DATABASE_DSN is required
```

or:

```text
ping mysql: ...
```

#### D. Local Installation

Recommended beginner approach: Docker use karo.

Native install steps:

Windows:

```powershell
winget install Oracle.MySQL
mysql --version
```

Alternative: MySQL Installer use karo, root password set karo, and MySQL Server start karo.

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

Verify MySQL server:

```bash
mysqladmin ping -h 127.0.0.1 -P 3306 -u root -p
```

#### E. Docker Setup

Create persistent Docker volume:

```bash
docker volume create ecommerce_user_mysql_data
```

Run MySQL container:

```bash
docker run -d \
  --name ecommerce-user-mysql \
  -e MYSQL_ROOT_PASSWORD=root_password \
  -e MYSQL_DATABASE=user_db \
  -e MYSQL_USER=ecommerce_user \
  -e MYSQL_PASSWORD=ecommerce_password \
  -p 3306:3306 \
  -v ecommerce_user_mysql_data:/var/lib/mysql \
  mysql:8.4
```

Check container:

```bash
docker ps
docker logs ecommerce-user-mysql
```

Verify MySQL:

```bash
docker exec -it ecommerce-user-mysql mysqladmin ping -uroot -proot_password
```

#### F. Docker Compose Example

Current repo me `docker-compose.yml` present nahi hai. Agar beginner local setup banana ho, root folder me temporary local compose file create kar sakte ho, but commit karne se pehle team convention follow karo.

Example:

```yaml
services:
  user-mysql:
    image: mysql:8.4
    container_name: ecommerce-user-mysql
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: root_password
      MYSQL_DATABASE: user_db
      MYSQL_USER: ecommerce_user
      MYSQL_PASSWORD: ecommerce_password
      TZ: UTC
    ports:
      - "3306:3306"
    volumes:
      - user_mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "127.0.0.1", "-uroot", "-proot_password"]
      interval: 10s
      timeout: 5s
      retries: 10

volumes:
  user_mysql_data:
```

Run:

```bash
docker compose up -d
docker compose ps
docker compose logs user-mysql
```

Stop:

```bash
docker compose down
```

Stop and delete DB data:

```bash
docker compose down -v
```

Warning: `down -v` database volume delete karta hai. Local data lost ho jayega.

#### G. Start Commands

Native Linux:

```bash
sudo systemctl start mysql
sudo systemctl status mysql
```

macOS Homebrew:

```bash
brew services start mysql
brew services list
```

Docker:

```bash
docker start ecommerce-user-mysql
docker ps
```

#### H. Verify Running

Connect with root:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p
```

Connect with app user:

```bash
mysql -h 127.0.0.1 -P 3306 -u ecommerce_user -p user_db
```

Show tables:

```sql
SHOW DATABASES;
USE user_db;
SHOW TABLES;
```

#### I. Default Port

```text
3306
```

#### J. Connection String Format

Go MySQL driver DSN format:

```env
USER_SERVICE_DATABASE_DSN=ecommerce_user:ecommerce_password@tcp(127.0.0.1:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC
```

Important:

- `parseTime=true` required hai because code timestamps ko Go `time.Time` me scan karta hai.
- `charset=utf8mb4` emoji/multilingual text safe banata hai.
- `loc=UTC` timestamps consistent rakhta hai.

Alternative fallback env var:

```env
MYSQL_DSN=ecommerce_user:ecommerce_password@tcp(127.0.0.1:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC
```

Code pehle `USER_SERVICE_DATABASE_DSN` read karta hai. Agar blank ho to `MYSQL_DSN` use karta hai.

#### K. Where To Place Credentials

Recommended local file:

```text
backend/services/user-service/.env
```

Do not commit `.env`.

Credentials placement:

| Credential | Env variable | Example |
|---|---|---|
| DB username | Inside `USER_SERVICE_DATABASE_DSN` | `ecommerce_user` |
| DB password | Inside `USER_SERVICE_DATABASE_DSN` | `ecommerce_password` |
| DB host | Inside `USER_SERVICE_DATABASE_DSN` | `127.0.0.1` |
| DB port | Inside `USER_SERVICE_DATABASE_DSN` | `3306` |
| DB name | Inside `USER_SERVICE_DATABASE_DSN` | `user_db` |

Production:

- Use Kubernetes Secret, cloud secret manager, or CI/CD secret store.
- Root DB password app ko mat do.
- App user ko limited permissions do.

#### L. Run Migration

Migration file:

```text
backend/services/user-service/migrations/001_create_user_tables.up.sql
```

It creates:

- `user_db`
- `users`
- `user_addresses`
- `seller_profiles`
- `seller_kyc_documents`

Run with native/local MySQL:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p < backend/services/user-service/migrations/001_create_user_tables.up.sql
```

Run inside Docker container:

```bash
docker exec -i ecommerce-user-mysql mysql -uroot -proot_password < backend/services/user-service/migrations/001_create_user_tables.up.sql
```

Verify:

```bash
mysql -h 127.0.0.1 -P 3306 -u ecommerce_user -p user_db -e "SHOW TABLES;"
```

Expected tables:

```text
seller_kyc_documents
seller_profiles
user_addresses
users
```

Rollback migration:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p < backend/services/user-service/migrations/001_create_user_tables.down.sql
```

Warning: Down migration drops tables. Local debugging me okay, production me backup ke bina kabhi run mat karo.

#### M. App User Permissions

If you created MySQL manually, app user create karo:

```sql
CREATE DATABASE IF NOT EXISTS user_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE USER IF NOT EXISTS 'ecommerce_user'@'%' IDENTIFIED BY 'ecommerce_password';

GRANT SELECT, INSERT, UPDATE, DELETE
ON user_db.*
TO 'ecommerce_user'@'%';

FLUSH PRIVILEGES;
```

For migrations, root/admin user usually required hota hai because migration creates database/tables.

---

## 6. Redis / Queue / External Services

### Current User Service Requirement

Current code path me only MySQL mandatory hai.

| Service | Required now? | Notes |
|---|---:|---|
| Redis | No | Not connected in current User Service code. |
| Kafka | No | Future user events may use Kafka. |
| RabbitMQ | No | Future user events may use RabbitMQ. |
| MinIO/S3/Object Storage | No for startup | Task 1 mentions KYC file storage conceptually, but current code stores only `storage_url` metadata. |
| API Gateway | No for direct gRPC run | Public REST API expected later via gateway. |
| Nginx | No | Not needed for current local gRPC service. |
| Kubernetes | No | Production deployment only. |

### Redis

What it is:

Redis ek in-memory cache/data store hai. Fast access ke liye use hota hai, jaise sessions, OTP counters, rate limiting, cache.

Why this project may use it:

Overall platform docs Redis ko Auth, Cart, Session, Recommendation cache ke liye mention karte hain. User Service current code me Redis use nahi karta.

Required or optional:

Optional/future for User Service.

Docker setup if needed later:

```bash
docker volume create ecommerce_redis_data

docker run -d \
  --name ecommerce-redis \
  -p 6379:6379 \
  -v ecommerce_redis_data:/data \
  redis:7-alpine \
  redis-server --appendonly yes --requirepass dev_redis_password
```

Verify:

```bash
docker exec -it ecommerce-redis redis-cli -a dev_redis_password ping
```

Expected:

```text
PONG
```

Default port:

```text
6379
```

Credential placement if future code adds Redis:

```env
REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_PASSWORD=dev_redis_password
```

### RabbitMQ

What it is:

RabbitMQ ek message broker hai. Services async messages bhej sakti hain, jaise `UserCreated` event ya notification command.

Why this project may use it:

Future Task 8 style event publishing ke liye `user.events` queue/exchange use ho sakta hai.

Required or optional:

Optional/future for current User Service.

Docker setup:

```bash
docker run -d \
  --name ecommerce-rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  -e RABBITMQ_DEFAULT_USER=ecommerce \
  -e RABBITMQ_DEFAULT_PASS=ecommerce_password \
  rabbitmq:3-management
```

Health check:

```bash
docker exec ecommerce-rabbitmq rabbitmq-diagnostics ping
```

Management UI:

```text
http://localhost:15672
username: ecommerce
password: ecommerce_password
```

Default ports:

| Port | Purpose |
|---:|---|
| 5672 | AMQP app connection |
| 15672 | Browser management UI |

Future credential placement:

```env
USER_EVENTS_PROVIDER=rabbitmq
RABBITMQ_URL=amqp://ecommerce:ecommerce_password@127.0.0.1:5672/
```

### Kafka

What it is:

Kafka ek distributed event streaming platform hai. High-throughput events ke liye useful hota hai.

Why this project may use it:

Large-scale e-commerce me user, order, product, payment events stream karne ke liye Kafka use ho sakta hai.

Required or optional:

Optional/future for current User Service.

Future env example:

```env
USER_EVENTS_PROVIDER=kafka
KAFKA_BROKERS=127.0.0.1:9092
USER_EVENTS_TOPIC=user.events
```

Health check:

```bash
# Depends on chosen Kafka image/tooling.
# Current repo does not include Kafka compose yet.
```

Beginner recommendation:

Current User Service run karne ke liye Kafka install mat karo. Jab event publisher implementation add ho, tab team-provided Docker Compose use karo.

### Object Storage / MinIO / S3

What it is:

Object storage files store karta hai, jaise images, invoices, KYC PDFs. S3 cloud version hai; MinIO local S3-compatible server hai.

Why this project may use it:

Task 1 domain guide me KYC actual files database me store nahi karne ka rule hai. DB me sirf `storage_url` metadata rahega.

Required or optional:

Optional/future. Current service startup ke liye required nahi.

Local MinIO Docker setup if needed later:

```bash
docker volume create ecommerce_minio_data

docker run -d \
  --name ecommerce-minio \
  -p 9000:9000 \
  -p 9001:9001 \
  -e MINIO_ROOT_USER=ecommerce_minio \
  -e MINIO_ROOT_PASSWORD=ecommerce_minio_password \
  -v ecommerce_minio_data:/data \
  quay.io/minio/minio server /data --console-address ":9001"
```

Console:

```text
http://localhost:9001
```

Future credential placement:

```env
OBJECT_STORAGE_ENDPOINT=http://127.0.0.1:9000
OBJECT_STORAGE_BUCKET=user-kyc
OBJECT_STORAGE_ACCESS_KEY=ecommerce_minio
OBJECT_STORAGE_SECRET_KEY=ecommerce_minio_password
```

---

## 7. Environment Variables

### How Project Loads Env Variables

Current code uses Go standard `os.Getenv`.

Important beginner note:

The binary does not automatically read `.env` file. `.env` ek convention hai. Aapko variables shell me export karne honge, Docker Compose `env_file` use karna hoga, ya direnv/tooling use karni hogi.

Config loader file:

```text
backend/services/user-service/internal/config/config.go
```

### Where To Create `.env`

Recommended local path:

```text
backend/services/user-service/.env
```

Do not commit it. Root `.gitignore` already `.env` files ignore karta hai.

### Complete `.env` Example

```env
# Required: MySQL DSN for User Service
USER_SERVICE_DATABASE_DSN=ecommerce_user:ecommerce_password@tcp(127.0.0.1:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC

# Optional fallback if USER_SERVICE_DATABASE_DSN is not set
MYSQL_DSN=ecommerce_user:ecommerce_password@tcp(127.0.0.1:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC

# gRPC server config
USER_SERVICE_GRPC_ADDRESS=:50052
USER_SERVICE_GRPC_REFLECTION=true
USER_SERVICE_SHUTDOWN_TIMEOUT=10s

# Database pool config
USER_SERVICE_DB_MAX_OPEN_CONNS=25
USER_SERVICE_DB_MAX_IDLE_CONNS=25
USER_SERVICE_DB_CONN_MAX_LIFETIME=5m
USER_SERVICE_DB_PING_TIMEOUT=5s

# Logging
USER_SERVICE_LOG_LEVEL=debug
```

### Env Variable Details

| Variable | Required? | Default | Purpose | Example | Security note |
|---|---:|---|---|---|---|
| `USER_SERVICE_DATABASE_DSN` | Yes | None | Main MySQL connection string. | `ecommerce_user:...@tcp(127.0.0.1:3306)/user_db?...` | Contains password. Never commit. |
| `MYSQL_DSN` | Conditional | None | Fallback DSN if main DSN is blank. | Same as above | Avoid setting both differently. |
| `USER_SERVICE_GRPC_ADDRESS` | No | `:50052` | gRPC listen address. | `:50052` | In production bind carefully. |
| `USER_SERVICE_GRPC_REFLECTION` | No | `true` | Allows grpcurl/proto discovery. | `true` local, `false` prod | Disable in production unless intentionally exposed internally. |
| `USER_SERVICE_SHUTDOWN_TIMEOUT` | No | `10s` | Graceful shutdown wait time. | `10s` | Too low can drop requests. |
| `USER_SERVICE_DB_MAX_OPEN_CONNS` | No | `25` | Max open DB connections. | `25` | Too high can overload MySQL. |
| `USER_SERVICE_DB_MAX_IDLE_CONNS` | No | `25` | Max idle DB connections. | `10` | Keep <= max open conns. |
| `USER_SERVICE_DB_CONN_MAX_LIFETIME` | No | `5m` | Recycle DB connections after duration. | `5m` | Useful with load balancers/proxies. |
| `USER_SERVICE_DB_PING_TIMEOUT` | No | `5s` | Startup DB ping timeout. | `5s` | Too low can fail on slow local Docker startup. |
| `USER_SERVICE_LOG_LEVEL` | No | `info` | Log level. | `debug`, `info`, `warn`, `error` | Avoid debug logs with PII in production. |

### Load `.env` On Linux/macOS

```bash
cd backend/services/user-service
set -a
source .env
set +a
go run ./cmd/server
```

### Load Env On Windows PowerShell

```powershell
cd backend/services/user-service
$env:USER_SERVICE_DATABASE_DSN="ecommerce_user:ecommerce_password@tcp(127.0.0.1:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC"
$env:USER_SERVICE_GRPC_ADDRESS=":50052"
$env:USER_SERVICE_GRPC_REFLECTION="true"
$env:USER_SERVICE_LOG_LEVEL="debug"
go run .\cmd\server
```

### Common Env Mistakes

| Mistake | Symptom | Fix |
|---|---|---|
| `.env` created but not sourced | App says DSN required. | Run `set -a; source .env; set +a`. |
| Missing `parseTime=true` | Timestamp scan errors. | Add `?parseTime=true` to DSN. |
| Wrong DB host in Docker | Connection refused. | If app runs on host use `127.0.0.1`; if app runs in compose use service name like `user-mysql`. |
| Password contains special chars | DSN parse error. | URL-escape special characters or use simpler local password. |
| Port already used | MySQL or gRPC cannot bind. | Change host port or service address. |
| Both DSNs set differently | Confusing DB connection. | Prefer only `USER_SERVICE_DATABASE_DSN`. |

---

## 8. Docker Setup

### Current Docker Status

Current repo does not include a User Service Dockerfile or docker-compose file.

So Docker is currently mainly useful for:

- Running MySQL locally.
- Running optional future services like Redis/RabbitMQ/MinIO.

### Recommended Beginner Approach

Use Docker for MySQL, run Go service directly on host.

Why:

- Easy DB setup.
- Go service logs directly visible in terminal.
- Debugging simpler.
- No need to build service image until Dockerfile exists.

### Useful Docker Commands

Start MySQL container:

```bash
docker start ecommerce-user-mysql
```

Stop MySQL container:

```bash
docker stop ecommerce-user-mysql
```

Show containers:

```bash
docker ps
docker ps -a
```

Show logs:

```bash
docker logs ecommerce-user-mysql
```

Open MySQL shell:

```bash
docker exec -it ecommerce-user-mysql mysql -uroot -proot_password
```

Remove stopped container:

```bash
docker rm ecommerce-user-mysql
```

Warning: Remove container data only if volume is preserved. If volume deleted, DB data lost.

### Future Dockerfile Shape

Not currently present, but production-ready Go service Dockerfile should roughly:

- Use `golang:1.24-alpine` or official Go builder image.
- Build static binary.
- Copy binary into distroless/alpine runtime.
- Expose gRPC port `50052`.
- Run as non-root user if possible.
- Read config from env variables.

Example concept:

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY backend/go.work backend/go.work.sum ./backend/
COPY backend/shared ./backend/shared
COPY backend/services/user-service ./backend/services/user-service
WORKDIR /app/backend/services/user-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/user-service ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /out/user-service /user-service
EXPOSE 50052
ENTRYPOINT ["/user-service"]
```

Do not add this unless task specifically asks for Docker implementation.

### Volumes

MySQL volume:

```text
ecommerce_user_mysql_data:/var/lib/mysql
```

Why important:

Volume ke bina container delete hone par database data delete ho sakta hai.

### Networks

If app also runs in Docker Compose, DSN host should be service name:

```env
USER_SERVICE_DATABASE_DSN=ecommerce_user:ecommerce_password@tcp(user-mysql:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC
```

If app runs on your host machine and MySQL runs in Docker:

```env
USER_SERVICE_DATABASE_DSN=ecommerce_user:ecommerce_password@tcp(127.0.0.1:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC
```

---

## 9. Local Development Setup

### Step 1: Clone Repository

```bash
git clone <repository-url>
cd Ecommerce
```

Verify important folders:

```bash
ls backend/services/user-service
ls backend/shared/gen/go
ls proto/ecommerce/user/v1
```

### Step 2: Install Required Tools

```bash
git --version
go version
docker --version
docker compose version
```

Need Go 1.24+.

### Step 3: Download Go Dependencies

```bash
cd backend/services/user-service
go mod download
go mod tidy
```

### Step 4: Start MySQL

Docker recommended:

```bash
docker volume create ecommerce_user_mysql_data

docker run -d \
  --name ecommerce-user-mysql \
  -e MYSQL_ROOT_PASSWORD=root_password \
  -e MYSQL_DATABASE=user_db \
  -e MYSQL_USER=ecommerce_user \
  -e MYSQL_PASSWORD=ecommerce_password \
  -p 3306:3306 \
  -v ecommerce_user_mysql_data:/var/lib/mysql \
  mysql:8.4
```

Wait until ready:

```bash
docker logs ecommerce-user-mysql
docker exec -it ecommerce-user-mysql mysqladmin ping -uroot -proot_password
```

### Step 5: Run Database Migration

From repo root:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p < backend/services/user-service/migrations/001_create_user_tables.up.sql
```

If using Docker and no local `mysql` CLI installed:

```bash
docker exec -i ecommerce-user-mysql mysql -uroot -proot_password < backend/services/user-service/migrations/001_create_user_tables.up.sql
```

Verify:

```bash
docker exec -it ecommerce-user-mysql mysql -uecommerce_user -pecommerce_password user_db -e "SHOW TABLES;"
```

### Step 6: Create `.env`

Create:

```text
backend/services/user-service/.env
```

Content:

```env
USER_SERVICE_DATABASE_DSN=ecommerce_user:ecommerce_password@tcp(127.0.0.1:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC
USER_SERVICE_GRPC_ADDRESS=:50052
USER_SERVICE_GRPC_REFLECTION=true
USER_SERVICE_SHUTDOWN_TIMEOUT=10s
USER_SERVICE_DB_MAX_OPEN_CONNS=25
USER_SERVICE_DB_MAX_IDLE_CONNS=25
USER_SERVICE_DB_CONN_MAX_LIFETIME=5m
USER_SERVICE_DB_PING_TIMEOUT=5s
USER_SERVICE_LOG_LEVEL=debug
```

### Step 7: Load Env And Run Service

Linux/macOS:

```bash
cd backend/services/user-service
set -a
source .env
set +a
go run ./cmd/server
```

Expected log:

```json
{"msg":"user_service_grpc_listening","address":":50052"}
```

### Step 8: Verify gRPC Service

Install grpcurl if not installed:

```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

Make sure Go bin is in PATH:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

List services:

```bash
grpcurl -plaintext localhost:50052 list
```

Describe User Service:

```bash
grpcurl -plaintext localhost:50052 describe ecommerce.user.v1.UserService
```

Create test user:

```bash
grpcurl -plaintext \
  -d '{"auth_account_id":"auth_dev_001","email":"dev.user@example.com","phone":"+919999999999","full_name":"Dev User"}' \
  localhost:50052 ecommerce.user.v1.UserService/CreateUser
```

Get user:

```bash
grpcurl -plaintext \
  -d '{"user_id":"<returned_user_id>"}' \
  localhost:50052 ecommerce.user.v1.UserService/GetUser
```

Note:

Current service is internal gRPC. Public REST routes are expected through API Gateway in later tasks, not directly in this service.

### Step 9: Run Tests

```bash
cd backend/services/user-service
go test ./...
```

### Step 10: Build Binary

```bash
cd backend/services/user-service
go build -o ./bin/user-service ./cmd/server
```

Run built binary:

```bash
set -a
source .env
set +a
./bin/user-service
```

---

## 10. Running The Project

### Quick Start Commands

From repo root:

```bash
# 1. Start MySQL
docker start ecommerce-user-mysql

# 2. If first time only, run migration
docker exec -i ecommerce-user-mysql mysql -uroot -proot_password < backend/services/user-service/migrations/001_create_user_tables.up.sql

# 3. Run User Service
cd backend/services/user-service
set -a
source .env
set +a
go run ./cmd/server
```

### Service Startup Flow

When `go run ./cmd/server` starts:

1. Loads env variables from process environment.
2. Validates `USER_SERVICE_DATABASE_DSN`.
3. Opens MySQL connection.
4. Pings MySQL with timeout.
5. Creates repositories.
6. Creates usecase service.
7. Starts gRPC server on `USER_SERVICE_GRPC_ADDRESS`.
8. Enables gRPC reflection if configured.
9. Waits for SIGINT/SIGTERM.
10. Gracefully shuts down.

### Ports & Networking

| Service | Port | Required now? | Purpose |
|---|---:|---:|---|
| User Service gRPC | 50052 | Yes | Internal gRPC API |
| MySQL | 3306 | Yes | User database |
| API Gateway HTTP | 8080 | No for current service | Future public REST entrypoint |
| Redis | 6379 | No | Future cache/session/rate-limit |
| RabbitMQ AMQP | 5672 | No | Future async events |
| RabbitMQ UI | 15672 | No | Future queue debugging |
| Kafka | 9092 | No | Future event streaming |
| MinIO API | 9000 | No | Future KYC/object storage |
| MinIO Console | 9001 | No | Future object storage admin UI |

### Port Conflicts

Check port usage:

Linux/macOS:

```bash
lsof -i :50052
lsof -i :3306
```

Linux alternative:

```bash
ss -ltnp | grep 50052
ss -ltnp | grep 3306
```

Windows PowerShell:

```powershell
netstat -ano | findstr :50052
netstat -ano | findstr :3306
```

Change gRPC port:

```env
USER_SERVICE_GRPC_ADDRESS=:50053
```

Change MySQL Docker host port:

```bash
docker run ... -p 3307:3306 ...
```

Then DSN:

```env
USER_SERVICE_DATABASE_DSN=ecommerce_user:ecommerce_password@tcp(127.0.0.1:3307)/user_db?parseTime=true&charset=utf8mb4&loc=UTC
```

Firewall notes:

- Local development me `127.0.0.1` safest hai.
- Production me DB port publicly expose mat karo.
- gRPC service ko internal network/service mesh ke andar rakho.

---

## 11. Common Errors & Fixes

### `USER_SERVICE_DATABASE_DSN is required`

Cause:

Env variable set nahi hai.

Fix:

```bash
cd backend/services/user-service
set -a
source .env
set +a
go run ./cmd/server
```

Prevention:

`.env.example` add karo and onboarding docs follow karo.

### `ping mysql: dial tcp 127.0.0.1:3306: connect: connection refused`

Cause:

MySQL running nahi hai, wrong host/port, or Docker container stopped.

Fix:

```bash
docker ps -a
docker start ecommerce-user-mysql
docker logs ecommerce-user-mysql
```

Prevention:

Service run karne se pehle MySQL health check karo.

### `Access denied for user`

Cause:

Wrong username/password or user permissions missing.

Fix:

```bash
docker exec -it ecommerce-user-mysql mysql -uroot -proot_password
```

Then:

```sql
CREATE USER IF NOT EXISTS 'ecommerce_user'@'%' IDENTIFIED BY 'ecommerce_password';
GRANT SELECT, INSERT, UPDATE, DELETE ON user_db.* TO 'ecommerce_user'@'%';
FLUSH PRIVILEGES;
```

Prevention:

DB credentials `.env` and Docker env same rakho.

### `Unknown database 'user_db'`

Cause:

Migration nahi chali or DB create nahi hua.

Fix:

```bash
docker exec -i ecommerce-user-mysql mysql -uroot -proot_password < backend/services/user-service/migrations/001_create_user_tables.up.sql
```

Prevention:

First setup checklist me migration step skip mat karo.

### `Error 1146: Table ... doesn't exist`

Cause:

DB exists but tables missing.

Fix:

Run up migration again:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p < backend/services/user-service/migrations/001_create_user_tables.up.sql
```

Prevention:

After migration `SHOW TABLES;` verify karo.

### Timestamp Scan Error

Cause:

DSN me `parseTime=true` missing hai.

Fix:

```env
USER_SERVICE_DATABASE_DSN=ecommerce_user:ecommerce_password@tcp(127.0.0.1:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC
```

Prevention:

Always copy DSN from this guide or `.env.example`.

### `listen tcp :50052: bind: address already in use`

Cause:

Another process same gRPC port use kar raha hai.

Fix:

```bash
lsof -i :50052
```

Stop old process or change:

```env
USER_SERVICE_GRPC_ADDRESS=:50053
```

Prevention:

One service instance per local port.

### Docker Daemon Not Running

Cause:

Docker Desktop/Engine started nahi hai.

Fix:

Windows/macOS: Docker Desktop start karo.

Linux:

```bash
sudo systemctl start docker
sudo systemctl status docker
```

Prevention:

Setup ke start me `docker ps` run karo.

### Docker Container Name Already Exists

Cause:

`ecommerce-user-mysql` container pehle se created hai.

Fix:

```bash
docker start ecommerce-user-mysql
```

If broken and data not needed:

```bash
docker rm ecommerce-user-mysql
```

Prevention:

Use `docker ps -a` before creating new container.

### `go: command not found`

Cause:

Go installed nahi hai or PATH me nahi hai.

Fix:

Install Go and restart terminal.

Prevention:

`go version` setup checklist ka first step banao.

### `go mod download` Failed

Cause:

Network/proxy issue or Go proxy blocked.

Fix:

```bash
go env GOPROXY
go env -w GOPROXY=https://proxy.golang.org,direct
go mod download
```

Corporate network me proxy settings configure karo.

Prevention:

Dependencies commit karo through `go.mod` and `go.sum`; CI me dependency download test karo.

### `grpcurl: command not found`

Cause:

grpcurl installed nahi ya Go bin PATH me missing.

Fix:

```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
export PATH="$PATH:$(go env GOPATH)/bin"
```

Prevention:

Dev setup checklist me grpcurl optional tool mention karo.

### `grpcurl localhost:50052 list` Fails

Cause:

Service not running, wrong port, or reflection disabled.

Fix:

```bash
docker ps
go run ./cmd/server
```

If reflection disabled:

```env
USER_SERVICE_GRPC_REFLECTION=true
```

Prevention:

Reflection local dev me true rakho, production me policy ke according disable karo.

### Permission Denied On Docker

Cause:

Linux user docker group me nahi hai.

Fix:

```bash
sudo usermod -aG docker "$USER"
```

Logout/login.

Prevention:

Docker setup ke baad `docker run hello-world` test karo.

---

## 12. Security & Best Practices

### Secrets

Best practices:

- `.env` commit mat karo.
- DB password strong rakho.
- Local dev passwords simple ho sakte hain, production me strong generated secrets use karo.
- Root DB user app ko mat do.
- Production secrets Kubernetes Secret, cloud secret manager, or CI secret store me rakho.

### MySQL

Best practices:

- App user ko limited permissions do.
- Production me backups enable karo.
- Migration se pehle backup/snapshot lo.
- `utf8mb4` charset use karo.
- DB publicly expose mat karo.
- Connection pool DB capacity ke according tune karo.

### gRPC

Best practices:

- Production me gRPC reflection disabled rakho unless internal debugging allowed hai.
- Internal service traffic mTLS/service mesh behind rakho.
- API Gateway/Auth Service verified metadata pass kare.
- Request deadlines/timeouts set karo.
- Health service add karo for readiness/liveness.

### Logging

Current code JSON logs stdout par print karta hai.

Avoid logging:

- Passwords
- Tokens
- OTPs
- Full phone numbers
- Full KYC URLs
- Full addresses
- DB DSN with password

### Docker

Best practices:

- Use volumes for DB data.
- Do not put production passwords in compose files.
- Use healthchecks.
- Pin image versions for stable builds.
- Run app containers as non-root when Dockerfile is added.

### Development

Best practices:

- `go test ./...` before commit.
- `go mod tidy` after dependency changes.
- Keep `go.mod` and `go.sum` committed.
- Keep migrations up/down paired.
- Do not directly edit generated protobuf files unless generation process is unavailable.
- Keep User Service DB private to User Service. Other services should use gRPC/events.

---

## 13. Missing Or Misconfigured Things

This is a professional audit of setup/devops gaps found while analyzing the current implementation.

| Finding | Impact | Suggested fix |
|---|---|---|
| No `backend/services/user-service/.env.example` | Beginners do not know required env vars. | Add `.env.example` with safe placeholder DSN. |
| No Dockerfile for User Service | Cannot containerize service directly yet. | Add service Dockerfile when deployment task starts. |
| No repo Docker Compose stack | Beginners must create MySQL manually. | Add `infra/compose/docker-compose.local.yml` with MySQL and future services. |
| No integrated migration runner | Migrations must be run manually through MySQL CLI. | Add documented `make migrate-up` or golang-migrate support. |
| Migration hardcodes `user_db` | Local okay, multi-env less flexible. | Keep for local or introduce environment-specific migration workflow. |
| No gRPC health service | Orchestrators cannot check readiness cleanly. | Add `grpc.health.v1.Health` service. |
| gRPC reflection defaults to true | Good for local, risky if exposed broadly. | Set `USER_SERVICE_GRPC_REFLECTION=false` in production. |
| No metrics/tracing config | Production observability incomplete. | Add Prometheus/OpenTelemetry in future infra task. |
| No strict auth enforcement in service startup docs | Internal trust boundary can be misunderstood. | Gateway/Auth must pass verified metadata; add auth interceptor/mTLS before production. |
| Current `requireSelfOrService` allows empty caller user id | Internal calls without identity may pass self-check. | Review security policy; require service identity or authenticated user for protected methods. |
| No Redis/Kafka env in current config | Future event/cache docs may confuse beginners. | Clearly mark Redis/Kafka as optional/future until code is wired. |
| No object storage config for KYC files | KYC upload flow not runnable yet. | Add MinIO/S3 config only when file upload feature is implemented. |
| No API Gateway in current User Service run path | REST examples from task docs are not directly runnable. | Document direct gRPC now, REST via Gateway later. |

### Hardcoded Credentials Audit

Current code does not hardcode DB username/password. Good.

Credentials are expected through env DSN:

```text
USER_SERVICE_DATABASE_DSN
```

But documentation/examples must never use production-like secrets. Local examples in this file are development-only.

### Insecure Defaults Audit

| Default | Local okay? | Production recommendation |
|---|---:|---|
| `USER_SERVICE_GRPC_REFLECTION=true` | Yes | Set false or internal-only |
| Bind `:50052` | Yes | Bind within private network/container |
| MySQL root password in Docker command | Local only | Use secret manager |
| Plaintext gRPC local testing | Local only | Use mTLS/service mesh in prod |

---

## 14. Final Checklist

Use this checklist for fresh setup.

- [ ] Clone repository.
- [ ] Install Git.
- [ ] Install Go 1.24+.
- [ ] Install Docker or native MySQL.
- [ ] Verify `go version`.
- [ ] Verify `docker --version` if using Docker.
- [ ] Start MySQL on port `3306`.
- [ ] Create/preserve MySQL Docker volume.
- [ ] Run `001_create_user_tables.up.sql`.
- [ ] Verify `user_db` exists.
- [ ] Verify tables: `users`, `user_addresses`, `seller_profiles`, `seller_kyc_documents`.
- [ ] Create `backend/services/user-service/.env`.
- [ ] Add `USER_SERVICE_DATABASE_DSN` with `parseTime=true`.
- [ ] Load `.env` into shell.
- [ ] Run `go mod download`.
- [ ] Run `go test ./...`.
- [ ] Run `go run ./cmd/server`.
- [ ] Verify gRPC listens on `localhost:50052`.
- [ ] Install `grpcurl` if manual API testing needed.
- [ ] Run `grpcurl -plaintext localhost:50052 list`.
- [ ] Never commit `.env`.
- [ ] Keep Redis/Kafka/MinIO optional until future tasks wire them in.

---

## Quick Reference

### Most Important Commands

```bash
# Start DB
docker start ecommerce-user-mysql

# Run migration
docker exec -i ecommerce-user-mysql mysql -uroot -proot_password < backend/services/user-service/migrations/001_create_user_tables.up.sql

# Run service
cd backend/services/user-service
set -a
source .env
set +a
go run ./cmd/server

# Test service
grpcurl -plaintext localhost:50052 list

# Run tests
go test ./...
```

### Most Important Env Variable

```env
USER_SERVICE_DATABASE_DSN=ecommerce_user:ecommerce_password@tcp(127.0.0.1:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC
```

### Beginner Rule

Pehle MySQL chalao, phir migration chalao, phir `.env` load karo, phir Go service run karo. Agar service start nahi ho rahi, 80% cases me issue DB, DSN, port, ya `.env` loading ka hota hai.
