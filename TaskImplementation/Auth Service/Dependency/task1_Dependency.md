# Project Dependency & Setup Guide

This document explains the non-code setup needed to run the Auth Service locally.
It is based on:

- `TaskImplementation/Auth Service/task1.md`
- `backend/go.work`
- `backend/services/auth-service/go.mod`
- `backend/services/auth-service/cmd/server/main.go`
- `backend/services/auth-service/internal/config/config.go`
- `backend/services/auth-service/migrations/*.sql`

Simple goal: ek beginner developer bhi repo clone karke dependencies install kare, MySQL/Redis start kare, environment variables set kare, migrations run kare, aur Auth Service start/debug kar sake.

Important note: `task1.md` originally says code/migrations were not included in Task 1, but current repository already contains Auth Service code and MySQL migrations. This guide reflects the current repository state.

---

## 1. Project Overview

Auth Service platform ka identity and access control service hai. Iska kaam hai:

- password credentials store/verify karna
- login ke baad access token and refresh token issue karna
- OTP challenge create/verify karna
- refresh token rotate karna
- logout par refresh tokens revoke karna
- roles/RBAC assignments manage karna
- session-related auth events outbox me store karna

Auth Service user profile ka owner nahi hai. Name, address, avatar, seller profile jaise fields User Service ya dusre domain services me rahenge.

Local run ke liye minimum services:

| Dependency | Required? | Why |
|---|---:|---|
| Go | Yes | Auth Service Go me implemented hai |
| MySQL 8+ | Yes | Auth accounts, credentials, refresh tokens, OTP, roles, outbox tables |
| Redis | Yes | OTP send/verify rate limiting |
| RSA JWT keys | Yes | RS256 access token sign/verify |
| Notification endpoint | Config required, runtime OTP required | OTP send API HTTP POST karta hai |
| Session link/outbox | Optional for beginner if disabled | Login/logout events Session Service ko bhejne ke liye |
| Docker | Optional but recommended | MySQL/Redis easily run karne ke liye |

---

## 2. Tech Stack

| Technology | What it is | Why this project uses it | Required? | Beginner explanation |
|---|---|---|---:|---|
| Go `1.26.3` | Backend programming language | Fast, simple, compiled microservices banane ke liye | Yes | Go ek compiled language hai. Is service ka server, usecases, repositories sab Go me hain. |
| Go modules | Dependency management system | External packages version-lock karne ke liye | Yes | `go.mod` dependencies ki list hai, `go.sum` checksum lock hai. |
| Go workspace | Multi-module workspace | `backend/go.work` service module ko include karta hai | Yes | Workspace se backend ke multiple Go modules ek saath develop ho sakte hain. |
| Standard `net/http` | Go ka built-in HTTP server | REST endpoints expose karne ke liye | Yes | Is repo me Fiber/Gin/Echo nahi hai. Plain Go HTTP server use ho raha hai. |
| `database/sql` | Go standard DB abstraction | MySQL queries and transactions chalane ke liye | Yes | Ye generic DB layer hai, MySQL driver ke saath kaam karta hai. |
| `github.com/go-sql-driver/mysql` | MySQL Go driver | Go service ko MySQL se connect karne ke liye | Yes | Driver bina Go app MySQL se baat nahi kar paayega. |
| MySQL 8+ | Relational database | Auth data strongly consistent tables me store hota hai | Yes | MySQL ek relational DB hai jisme data tables me store hota hai. |
| Redis | In-memory key-value store | OTP cooldown, send limits, verify IP limits ke liye | Yes | Redis fast memory store hai. Is project me short-lived rate limit counters ke liye use hota hai. |
| `golang.org/x/crypto` | Go crypto helpers | Argon2id, bcrypt password hashing ke liye | Yes | Password plain text store nahi hota. Strong hash generate hota hai. |
| Argon2id | Password hashing algorithm | New passwords securely hash karne ke liye | Yes | Argon2id slow hash hai jo brute force attacks ko hard banata hai. |
| bcrypt | Password hashing fallback | Legacy password hashes verify/rehash karne ke liye | Optional fallback | Agar purane bcrypt hashes hon to service verify karke Argon2id me migrate kar sakti hai. |
| JWT RS256 | Signed access token format | APIs ko stateless auth claims dene ke liye | Yes | JWT signed token hota hai. RS256 me private key se sign, public key se verify hota hai. |
| JWKS | Public key JSON endpoint | Gateway/other services JWT verify kar sakein | Yes | `/.well-known/jwks.json` public keys expose karta hai. |
| OTP HMAC SHA-256 | OTP hashing method | DB me raw OTP store na ho | Yes for OTP | OTP code ko pepper ke saath hash karke store kiya jata hai. |
| Outbox pattern | Durable event publishing pattern | Login/logout events reliably publish karne ke liye | Optional local, recommended prod | Event pehle MySQL table me store hota hai, worker baad me publish karta hai. |
| HTTP Notification Client | Outbound HTTP call | Notification Service ko OTP send request dene ke liye | Config required | Auth OTP generate karta hai, delivery Notification Service karega. |
| Docker | Container runtime | MySQL/Redis local setup simple karne ke liye | Optional | Docker se dependencies isolated containers me chalti hain. |
| OpenSSL | Crypto CLI | Local RSA private/public key generate karne ke liye | Recommended | JWT signing keys banane ke liye useful tool hai. |

### Technologies not currently used by this Auth Service code

| Technology | Status |
|---|---|
| Gin/Fiber/Echo | Not used. Service uses standard `net/http`. |
| `github.com/redis/go-redis/v9` | Not used. Redis client custom TCP RESP implementation hai. |
| Kafka/RabbitMQ Go client | Not used directly. Current publisher is HTTP-based outbox publisher. |
| gRPC server | Not implemented in this service yet. `SESSION_GRPC_ADDR` config exists but current `SESSION_LINK_MODE` supports only `outbox` or `disabled`. |
| Dockerfile/docker-compose | Not present in repo yet. Samples are provided below for local setup. |

---

## 3. Required Software

Install these before running the service.

| Software | Recommended version | Required? | Verify command |
|---|---|---:|---|
| Git | Any modern version | Yes | `git --version` |
| Go | Match `go.mod`: `1.26.3` | Yes | `go version` |
| MySQL Server | `8.x` | Yes | `mysql --version` |
| Redis Server | `7.x` or newer | Yes | `redis-cli --version` |
| OpenSSL | Any modern version | Yes for local key generation | `openssl version` |
| curl | Any modern version | Recommended | `curl --version` |
| Docker Desktop / Docker Engine | Modern Docker with Compose v2 | Optional but recommended | `docker --version` and `docker compose version` |

### Windows install options

Beginner-friendly recommended path:

1. Install Git for Windows.
2. Install Go from the official Go installer and ensure `go version` shows `go1.26.3` or compatible.
3. Install Docker Desktop.
4. Run MySQL and Redis through Docker instead of native Windows services.
5. Use PowerShell or Git Bash for commands.

Useful commands:

```powershell
winget install Git.Git
winget install GoLang.Go
winget install Docker.DockerDesktop
```

For MySQL and Redis on Windows, Docker setup is usually easiest:

```powershell
docker --version
docker compose version
```

### Linux install options

Ubuntu/Debian example:

```bash
sudo apt update
sudo apt install -y git curl openssl mysql-client redis-tools
```

Install Go separately if apt repository does not provide the version required by `go.mod`.

Docker Engine optional:

```bash
sudo apt install -y docker.io docker-compose-plugin
sudo systemctl enable --now docker
docker --version
docker compose version
```

### macOS install options

Homebrew example:

```bash
brew install git go mysql redis openssl curl
brew install --cask docker
```

Verify:

```bash
go version
mysql --version
redis-cli --version
openssl version
```

---

## 4. Dependency Management

This is a Go project.

### Important files

| File | Purpose |
|---|---|
| `backend/go.work` | Go workspace file. It includes `./services/auth-service`. |
| `backend/go.work.sum` | Workspace dependency checksum file. |
| `backend/services/auth-service/go.mod` | Auth Service module name, Go version, direct dependencies. |
| `backend/services/auth-service/go.sum` | Dependency integrity checksums. Do not manually edit normally. |

### Current direct Go dependencies

| Dependency | Version | Why used |
|---|---:|---|
| `github.com/go-sql-driver/mysql` | `v1.9.3` | MySQL driver for `database/sql` |
| `golang.org/x/crypto` | `v0.32.0` | Argon2id and bcrypt hashing |

Indirect dependencies may appear in `go.sum` because direct libraries need them.

### Dependency commands

Run from Auth Service module:

```bash
cd backend/services/auth-service
go mod download
go test ./...
go build ./cmd/server
go run ./cmd/server
```

Run from backend workspace root:

```bash
cd backend
go work sync
go test ./services/auth-service/...
```

### What each Go command does

| Command | Meaning |
|---|---|
| `go mod download` | Dependencies module cache me download karta hai. |
| `go mod tidy` | Unused dependencies remove karta hai and missing required dependencies add karta hai. Use only when dependencies change. |
| `go build ./cmd/server` | Auth server binary compile karta hai. |
| `go run ./cmd/server` | Compile + run in one command. Local development ke liye useful. |
| `go test ./...` | Saare packages ke tests run karta hai. |
| `go work sync` | Workspace module versions sync karta hai. |

### Common Go dependency issues

| Issue | Cause | Fix |
|---|---|---|
| `go: go.mod requires go >= 1.26.3` | Installed Go old hai | Go upgrade karo to `1.26.3` or compatible. |
| `checksum mismatch` | Module cache corrupted ya dependency changed | `go clean -modcache`, then `go mod download`. |
| `dial tcp ... proxy.golang.org` | Network/proxy issue | `go env -w GOPROXY=https://proxy.golang.org,direct` or company proxy configure karo. |
| `read-only file system` in Go build cache | Home cache writable nahi hai | `GOCACHE=/tmp/auth-go-cache go test ./...`. |
| `package not in std` | Wrong directory se command run hua | `cd backend/services/auth-service` or use workspace root command. |

Verified command in this workspace:

```bash
cd backend/services/auth-service
env GOCACHE=/tmp/auth-go-cache go test ./...
```

---

## 5. Database Setup

### Database detected: MySQL 8+

#### A. What it is

MySQL ek relational database hai jisme data tables, rows, columns ke form me store hota hai. Auth Service ke liye relational DB important hai kyunki credentials, refresh tokens, OTP attempts, and role assignments me consistency chahiye.

#### B. Why this project uses it

Auth Service MySQL me ye tables use karta hai:

| Table | Purpose |
|---|---|
| `auth_accounts` | Account identity, email/phone, status, user/seller/tenant IDs |
| `credentials` | Password hash, algorithm, failed attempts, lockout metadata |
| `refresh_tokens` | Rotating refresh token hashes and revocation state |
| `otp_challenges` | OTP hash, attempts, expiry, verification status |
| `role_assignments` | Buyer/seller/admin roles and scopes |
| `auth_outbox_events` | Session-link events for async publishing |

#### C. Required or optional

Required. Service startup calls `db.PingContext`. Agar MySQL unavailable hua to server start nahi hoga.

#### D. Local installation

Windows:

- Beginner recommended: Docker Desktop + MySQL container.
- Native option: install MySQL Installer for Windows, then enable MySQL Server and MySQL Shell/CLI.

Linux:

```bash
sudo apt update
sudo apt install -y mysql-server mysql-client
sudo systemctl enable --now mysql
sudo systemctl status mysql
```

macOS:

```bash
brew install mysql
brew services start mysql
mysql.server status
```

#### E. Docker setup

Persistent volume:

```bash
docker volume create ecommerce-auth-mysql-data
```

Run MySQL:

```bash
docker run --name ecommerce-auth-mysql \
  -e MYSQL_ROOT_PASSWORD=root_password \
  -e MYSQL_DATABASE=auth_db \
  -e MYSQL_USER=auth_user \
  -e MYSQL_PASSWORD=auth_password \
  -p 3306:3306 \
  -v ecommerce-auth-mysql-data:/var/lib/mysql \
  -d mysql:8.4
```

Check logs:

```bash
docker logs ecommerce-auth-mysql
```

#### F. Start commands

Native Linux:

```bash
sudo systemctl start mysql
```

Native macOS:

```bash
brew services start mysql
```

Docker:

```bash
docker start ecommerce-auth-mysql
```

#### G. Verify running

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p
```

Inside MySQL:

```sql
SELECT VERSION();
SHOW DATABASES;
```

#### H. Default port

MySQL default port is `3306`.

#### I. Connection string format

The service reads one env variable:

```env
AUTH_MYSQL_DSN='auth_user:auth_password@tcp(127.0.0.1:3306)/auth_db?parseTime=true&charset=utf8mb4&loc=UTC'
```

Important: `parseTime=true` is required because repository code scans MySQL `TIMESTAMP` columns into Go `time.Time`.
If you load `.env` using `source .env`, keep this DSN quoted because `&` has special meaning in shells.

#### J. Where to place credentials

Local development:

- Put DB username/password in `backend/services/auth-service/.env`.
- The app does not load `.env` automatically, so export/source it before running.
- Do not commit `.env`.

Production:

- Store DB password in secret manager or Kubernetes Secret.
- Inject `AUTH_MYSQL_DSN` as an environment variable.
- Do not put production DB credentials in `docker-compose.yml` or Git.

### Run migrations

Migration files are in:

```text
backend/services/auth-service/migrations/
```

Run them in order:

```bash
cd backend/services/auth-service
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/001_create_auth_tables.up.sql
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/002_add_token_claim_metadata.up.sql
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/003_add_otp_account_fk.up.sql
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/004_harden_role_assignments.up.sql
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/005_create_auth_outbox_events.up.sql
```

Verify tables:

```bash
mysql -h 127.0.0.1 -P 3306 -u auth_user -p -e "SHOW TABLES FROM auth_db;"
```

Expected tables:

```text
auth_accounts
credentials
refresh_tokens
otp_challenges
role_assignments
auth_outbox_events
```

### Rollback migrations

Rollback files exist, but rollback should be done carefully because it can delete data.

Order for rollback is reverse order:

```bash
cd backend/services/auth-service
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/005_create_auth_outbox_events.down.sql
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/004_harden_role_assignments.down.sql
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/003_add_otp_account_fk.down.sql
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/002_add_token_claim_metadata.down.sql
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/001_create_auth_tables.down.sql
```

Warning: local rollback is fine for practice. Production rollback needs backup and approval.

---

## 6. Redis / Queue / External Services

### Redis

#### What it is

Redis ek fast in-memory key-value store hai. Is project me OTP rate limiting ke liye use hota hai.

#### Why used

OTP endpoints abuse-prone hote hain. Redis me short-lived counters store hote hain:

- resend cooldown
- send limit per window
- daily send limit
- verify attempts per IP/source

#### Mandatory or optional

Required for current server startup. `main.go` Redis client create karke `PING` karta hai. Redis unavailable hua to server start nahi hoga.

#### Local installation

Linux:

```bash
sudo apt install -y redis-server redis-tools
sudo systemctl enable --now redis-server
```

macOS:

```bash
brew install redis
brew services start redis
```

Windows:

- Recommended: run Redis using Docker Desktop or WSL.
- Native Windows Redis packages are not the best default path for beginners.

#### Docker setup

Without password for local only:

```bash
docker volume create ecommerce-auth-redis-data
docker run --name ecommerce-auth-redis \
  -p 6379:6379 \
  -v ecommerce-auth-redis-data:/data \
  -d redis:7-alpine redis-server --appendonly yes
```

With password:

```bash
docker run --name ecommerce-auth-redis \
  -p 6379:6379 \
  -v ecommerce-auth-redis-data:/data \
  -d redis:7-alpine redis-server --appendonly yes --requirepass dev_redis_password
```

If using password, set:

```env
AUTH_REDIS_PASSWORD=dev_redis_password
```

#### Verify Redis

No password:

```bash
redis-cli -h 127.0.0.1 -p 6379 PING
```

With password:

```bash
redis-cli -h 127.0.0.1 -p 6379 -a dev_redis_password PING
```

Expected response:

```text
PONG
```

#### Redis default port

`6379`

#### Redis env format

```env
AUTH_REDIS_ADDR=127.0.0.1:6379
AUTH_REDIS_PASSWORD=
AUTH_REDIS_DB=0
```

### Notification Service

#### What it is

Notification Service email/SMS bhejne ke liye responsible service hai. Auth Service OTP generate karta hai, but delivery Notification Service ko delegate karta hai.

#### Why used

Auth Service ko email/SMS provider details own nahi karne chahiye. OTP send ke time Auth Service HTTP POST karta hai:

```text
POST http://localhost:8084/internal/v1/notifications/otp
```

#### Mandatory or optional

- Config mandatory hai because `NOTIFICATION_OTP_ENDPOINT` absolute URL hona chahiye.
- Server startup endpoint ko ping nahi karta.
- OTP send API tabhi successfully chalegi jab endpoint `2xx` response dega.
- Login/password/token APIs OTP endpoint ke bina bhi run ho sakte hain.

#### Credentials placement

Current Auth Service client notification auth token nahi bhejta. Production me mTLS, service token, or internal gateway auth add karna recommended hai.

```env
NOTIFICATION_OTP_ENDPOINT=http://localhost:8084/internal/v1/notifications/otp
NOTIFICATION_TIMEOUT=3s
```

#### Health check

Notification service available ho to:

```bash
curl -i http://localhost:8084/healthz
```

Actual health endpoint Notification Service implementation par depend karega.

### Session Link / Outbox

#### What it is

Session link Auth Service login/logout/refresh-reuse events ko Session Service tak pahunchane ka mechanism hai.

#### Current implementation

Current code supports:

| Mode | Behavior |
|---|---|
| `disabled` | Local debugging ke liye session event linking off |
| `outbox` | Events MySQL `auth_outbox_events` table me insert hote hain |

If `SESSION_LINK_MODE=outbox` and `AUTH_EVENTS_PUBLISH_ENDPOINT` is set, background worker HTTP POST karke events publish karega.

If `SESSION_LINK_MODE=outbox` and `AUTH_EVENTS_PUBLISH_ENDPOINT` empty hai, service start hogi but worker disabled warning log karega. Events table me pending rows accumulate ho sakti hain.

#### Beginner recommendation

First local run ke liye:

```env
SESSION_LINK_MODE=disabled
```

Integrated local run ke liye:

```env
SESSION_LINK_MODE=outbox
SESSION_EVENT_PEPPER=local_session_event_pepper_change_me_32_chars
AUTH_EVENTS_TOPIC=auth.events
AUTH_EVENTS_PUBLISH_ENDPOINT=http://localhost:8085/internal/v1/events
```

#### Kafka/RabbitMQ status

Task documentation architecture Kafka/RabbitMQ mention karta hai, but current Auth Service code direct Kafka/RabbitMQ client use nahi karta. Current publishing path HTTP publisher hai. `AUTH_EVENTS_TOPIC=auth.events` HTTP header `X-Event-Topic` me send hota hai.

Use Kafka/RabbitMQ only if your platform event-ingress service consumes the HTTP publish request and forwards it to a broker.

### OpenSSL / JWT Keys

Auth Service RS256 JWT sign karta hai. Iske liye RSA private key required hai.

Generate local keys:

```bash
cd backend/services/auth-service
mkdir -p secrets
openssl genrsa -out secrets/jwt-private.pem 2048
openssl rsa -in secrets/jwt-private.pem -pubout -out secrets/jwt-public.pem
chmod 600 secrets/jwt-private.pem
chmod 644 secrets/jwt-public.pem
```

Set env:

```env
JWT_KEY_ID=auth-local-1
JWT_PRIVATE_KEY_PEM_PATH=secrets/jwt-private.pem
JWT_PUBLIC_KEY_PEM_PATH=secrets/jwt-public.pem
JWT_SIGNING_ALG=RS256
```

Warning: `secrets/*.pem` files should never be committed.

---

## 7. Environment Variables

The service loads config using `os.Getenv`. It does not automatically read `.env`.

Recommended local file location:

```text
backend/services/auth-service/.env
```

Before running:

```bash
cd backend/services/auth-service
set -a
source .env
set +a
go run ./cmd/server
```

### Complete local `.env` example

Use this as a local template and change secrets.

```env
# HTTP server
AUTH_HTTP_ADDR=:8081
AUTH_HTTP_READ_TIMEOUT=5s
AUTH_HTTP_WRITE_TIMEOUT=10s
AUTH_HTTP_IDLE_TIMEOUT=60s
AUTH_SHUTDOWN_TIMEOUT=10s

# MySQL
AUTH_MYSQL_DSN='auth_user:auth_password@tcp(127.0.0.1:3306)/auth_db?parseTime=true&charset=utf8mb4&loc=UTC'

# Redis
AUTH_REDIS_ADDR=127.0.0.1:6379
AUTH_REDIS_PASSWORD=
AUTH_REDIS_DB=0
AUTH_REDIS_DIAL_TIMEOUT=2s
AUTH_REDIS_READ_TIMEOUT=2s
AUTH_REDIS_WRITE_TIMEOUT=2s

# Password policy and hashing
AUTH_PASSWORD_MIN_LENGTH=8
AUTH_PASSWORD_MAX_LENGTH=128
AUTH_PASSWORD_ARGON2_MEMORY_KIB=65536
AUTH_PASSWORD_ARGON2_ITERATIONS=3
AUTH_PASSWORD_ARGON2_PARALLELISM=2
AUTH_PASSWORD_SALT_LENGTH=16
AUTH_PASSWORD_KEY_LENGTH=32
AUTH_PASSWORD_BCRYPT_COST=12
AUTH_PASSWORD_MAX_FAILED_ATTEMPTS=5
AUTH_PASSWORD_LOCKOUT_DURATION=15m

# JWT access tokens and refresh tokens
JWT_ISSUER=ecommerce-auth
JWT_AUDIENCE=ecommerce-api
JWT_ACCESS_TTL=15m
REFRESH_TOKEN_TTL=720h
JWT_CLOCK_SKEW=30s
JWT_SIGNING_ALG=RS256
JWT_KEY_ID=auth-local-1
JWT_PRIVATE_KEY_PEM_PATH=secrets/jwt-private.pem
JWT_PUBLIC_KEY_PEM_PATH=secrets/jwt-public.pem
REFRESH_TOKEN_PEPPER=local_refresh_token_pepper_change_me_32_chars

# OTP
OTP_HASH_PEPPER=local_otp_hash_pepper_change_me_32_chars
OTP_RATE_LIMIT_PEPPER=local_otp_rate_pepper_change_me_32_chars
OTP_LENGTH=6
OTP_TTL=5m
OTP_MAX_ATTEMPTS=5
OTP_RESEND_COOLDOWN=1m
OTP_SEND_LIMIT_WINDOW=15m
OTP_SEND_LIMIT_PER_WINDOW=3
OTP_DAILY_LIMIT=10
OTP_VERIFY_IP_WINDOW=5m
OTP_VERIFY_IP_LIMIT=20

# RBAC
AUTH_ROLE_REASON_MAX_LENGTH=512

# Notification integration
NOTIFICATION_OTP_ENDPOINT=http://localhost:8084/internal/v1/notifications/otp
NOTIFICATION_TIMEOUT=3s

# Session link
# Beginner local mode:
SESSION_LINK_MODE=disabled

# Integrated mode only:
AUTH_EVENTS_TOPIC=auth.events
AUTH_EVENTS_PUBLISH_ENDPOINT=
SESSION_GRPC_ADDR=session-service:9090
SESSION_LINK_TIMEOUT=150ms
SESSION_EVENT_PEPPER=
OUTBOX_WORKER_BATCH_SIZE=100
OUTBOX_WORKER_INTERVAL=1s
OUTBOX_MAX_ATTEMPTS=5
OUTBOX_INITIAL_BACKOFF=5s
OUTBOX_MAX_BACKOFF=10m
OUTBOX_STALE_LOCK_TIMEOUT=5m
```

### Environment variable explanation

| Variable | Required? | Purpose | Example | Security notes |
|---|---:|---|---|---|
| `AUTH_HTTP_ADDR` | Optional | HTTP bind address | `:8081` | Use `127.0.0.1:8081` for local-only bind. |
| `AUTH_HTTP_READ_TIMEOUT` | Optional | Request read timeout | `5s` | Prevents slow client abuse. |
| `AUTH_HTTP_WRITE_TIMEOUT` | Optional | Response write timeout | `10s` | Prevents stuck responses. |
| `AUTH_HTTP_IDLE_TIMEOUT` | Optional | Keep-alive idle timeout | `60s` | Tune for gateway/proxy. |
| `AUTH_SHUTDOWN_TIMEOUT` | Optional | Graceful shutdown window | `10s` | Needed for clean deploy shutdown. |
| `AUTH_MYSQL_DSN` | Yes | MySQL connection string | `auth_user:auth_password@tcp(127.0.0.1:3306)/auth_db?parseTime=true&charset=utf8mb4&loc=UTC` | Contains password. Never commit. |
| `AUTH_REDIS_ADDR` | Optional | Redis host:port | `127.0.0.1:6379` | Use internal network name in Docker/K8s. |
| `AUTH_REDIS_PASSWORD` | Optional local | Redis password | `dev_redis_password` | Required if Redis has `requirepass`. |
| `AUTH_REDIS_DB` | Optional | Redis logical DB index | `0` | Keep services separated if sharing Redis. |
| `AUTH_REDIS_DIAL_TIMEOUT` | Optional | Redis connect timeout | `2s` | Keep small for fast startup failure. |
| `AUTH_REDIS_READ_TIMEOUT` | Optional | Redis read timeout | `2s` | Tune for infra latency. |
| `AUTH_REDIS_WRITE_TIMEOUT` | Optional | Redis write timeout | `2s` | Tune for infra latency. |
| `AUTH_PASSWORD_MIN_LENGTH` | Optional | Minimum password length | `8` | Higher is safer. |
| `AUTH_PASSWORD_MAX_LENGTH` | Optional | Maximum password length | `128` | Prevents huge request abuse. |
| `AUTH_PASSWORD_ARGON2_MEMORY_KIB` | Optional | Argon2id memory cost | `65536` | Production can tune based on CPU/RAM. |
| `AUTH_PASSWORD_ARGON2_ITERATIONS` | Optional | Argon2id iterations | `3` | Higher is slower/stronger. |
| `AUTH_PASSWORD_ARGON2_PARALLELISM` | Optional | Argon2id parallelism | `2` | Tune by CPU. |
| `AUTH_PASSWORD_SALT_LENGTH` | Optional | Salt length bytes | `16` | Keep non-zero. |
| `AUTH_PASSWORD_KEY_LENGTH` | Optional | Derived key bytes | `32` | Keep non-zero. |
| `AUTH_PASSWORD_BCRYPT_COST` | Optional | bcrypt fallback cost | `12` | Legacy support. |
| `AUTH_PASSWORD_MAX_FAILED_ATTEMPTS` | Optional | DB lockout threshold | `5` | Brute-force protection. |
| `AUTH_PASSWORD_LOCKOUT_DURATION` | Optional | Account lock duration | `15m` | Prevents rapid guessing. |
| `JWT_ISSUER` | Optional | JWT issuer claim | `ecommerce-auth` | Must match verifiers. |
| `JWT_AUDIENCE` | Optional | JWT audience claim | `ecommerce-api` | Must match gateway/services. |
| `JWT_ACCESS_TTL` | Optional | Access token validity | `15m` | Short TTL recommended. |
| `REFRESH_TOKEN_TTL` | Optional | Refresh token validity | `720h` | 720h = 30 days. |
| `JWT_CLOCK_SKEW` | Optional | Allowed clock skew | `30s` | Keep small. |
| `JWT_SIGNING_ALG` | Optional but constrained | Must be `RS256` | `RS256` | Code rejects other values. |
| `JWT_KEY_ID` | Yes | JWT `kid` header | `auth-local-1` | Needed for key rotation. |
| `JWT_PRIVATE_KEY_PEM_PATH` | Yes | RSA private key path | `secrets/jwt-private.pem` | Never commit private key. |
| `JWT_PUBLIC_KEY_PEM_PATH` | Optional recommended | RSA public key path | `secrets/jwt-public.pem` | Public key can be shared. |
| `REFRESH_TOKEN_PEPPER` | Yes | HMAC pepper for refresh token hashes | long random string | Treat as secret. Rotating invalidates old hashes. |
| `OTP_HASH_PEPPER` | Yes | HMAC pepper for OTP hashes | long random string | Treat as secret. |
| `OTP_RATE_LIMIT_PEPPER` | Optional if OTP hash pepper set | HMAC pepper for Redis lookup keys | long random string | Better to separate from OTP hash pepper. |
| `OTP_LENGTH` | Optional | OTP digits length | `6` | Code allows 6 to 10. |
| `OTP_TTL` | Optional | OTP expiry | `5m` | Short TTL is safer. |
| `OTP_MAX_ATTEMPTS` | Optional | OTP verify attempts per challenge | `5` | Prevents brute force. |
| `OTP_RESEND_COOLDOWN` | Optional | Minimum resend gap | `1m` | Uses Redis cooldown key. |
| `OTP_SEND_LIMIT_WINDOW` | Optional | Rate limit window | `15m` | Uses Redis. |
| `OTP_SEND_LIMIT_PER_WINDOW` | Optional | Sends allowed per window | `3` | Uses Redis. |
| `OTP_DAILY_LIMIT` | Optional | Daily sends per target | `10` | Uses Redis UTC day key. |
| `OTP_VERIFY_IP_WINDOW` | Optional | Verify attempt window per source | `5m` | Uses Redis. |
| `OTP_VERIFY_IP_LIMIT` | Optional | Verify attempts per source/window | `20` | Uses Redis. |
| `AUTH_ROLE_REASON_MAX_LENGTH` | Optional | Max audit reason length | `512` | Role changes need reason. |
| `NOTIFICATION_OTP_ENDPOINT` | Optional default, URL required | OTP delivery endpoint | `http://localhost:8084/internal/v1/notifications/otp` | Production should authenticate internal calls. |
| `NOTIFICATION_TIMEOUT` | Optional | Notification HTTP timeout | `3s` | Keep small. |
| `SESSION_LINK_MODE` | Optional | `disabled` or `outbox` | `disabled` local | Use `outbox` in integrated env. |
| `AUTH_EVENTS_TOPIC` | Required for outbox | Event topic name/header | `auth.events` | Must match event ingress/broker mapping. |
| `AUTH_EVENTS_PUBLISH_ENDPOINT` | Optional | HTTP event publisher endpoint | `http://localhost:8085/internal/v1/events` | Empty means outbox worker disabled. |
| `SESSION_GRPC_ADDR` | Optional/reserved | Session service gRPC address | `session-service:9090` | Currently not used by main. |
| `SESSION_LINK_TIMEOUT` | Optional | Publish/link timeout | `150ms` | Also supports `SESSION_LINK_TIMEOUT_MS`. |
| `SESSION_EVENT_PEPPER` | Required for outbox | Hashes IP/device values in session events | long random string | Required when mode is `outbox`. |
| `OUTBOX_WORKER_BATCH_SIZE` | Optional | Events per worker batch | `100` | Tune by event volume. |
| `OUTBOX_WORKER_INTERVAL` | Optional | Worker polling interval | `1s` | Also supports `OUTBOX_WORKER_INTERVAL_MS`. |
| `OUTBOX_MAX_ATTEMPTS` | Optional | Retry attempts before dead letter | `5` | Monitor dead letters. |
| `OUTBOX_INITIAL_BACKOFF` | Optional | First retry delay | `5s` | Exponential backoff starts here. |
| `OUTBOX_MAX_BACKOFF` | Optional | Max retry delay | `10m` | Must be >= initial backoff. |
| `OUTBOX_STALE_LOCK_TIMEOUT` | Optional | Recover stuck processing rows | `5m` | Helps worker crash recovery. |

### Common `.env` mistakes

| Mistake | What happens | Fix |
|---|---|---|
| Creating `.env` but not sourcing it | App says env variables are empty | Run `set -a; source .env; set +a`. |
| Missing `parseTime=true` in DSN | MySQL time scan errors | Add `?parseTime=true&charset=utf8mb4&loc=UTC`. |
| Relative key path from wrong directory | `read jwt private key` error | Run from `backend/services/auth-service` or use absolute key path. |
| `SESSION_LINK_MODE=outbox` but no `SESSION_EVENT_PEPPER` | Startup validation fails | Set pepper or use `SESSION_LINK_MODE=disabled` locally. |
| Redis password mismatch | Redis auth failure | Set `AUTH_REDIS_PASSWORD` exactly as Redis requires. |

---

## 8. Docker Setup

### Current repo status

No Dockerfile and no docker-compose file were found in the repository. That means:

- You can still run MySQL/Redis in Docker.
- Run the Go app locally with `go run`.
- App containerization needs a future Dockerfile.

### Beginner recommended local approach

Use Docker for dependencies only:

```bash
docker volume create ecommerce-auth-mysql-data
docker volume create ecommerce-auth-redis-data
```

Start MySQL:

```bash
docker run --name ecommerce-auth-mysql \
  -e MYSQL_ROOT_PASSWORD=root_password \
  -e MYSQL_DATABASE=auth_db \
  -e MYSQL_USER=auth_user \
  -e MYSQL_PASSWORD=auth_password \
  -p 3306:3306 \
  -v ecommerce-auth-mysql-data:/var/lib/mysql \
  -d mysql:8.4
```

Start Redis:

```bash
docker run --name ecommerce-auth-redis \
  -p 6379:6379 \
  -v ecommerce-auth-redis-data:/data \
  -d redis:7-alpine redis-server --appendonly yes
```

Useful Docker commands:

```bash
docker ps
docker logs ecommerce-auth-mysql
docker logs ecommerce-auth-redis
docker stop ecommerce-auth-mysql ecommerce-auth-redis
docker start ecommerce-auth-mysql ecommerce-auth-redis
```

### Docker Compose example for dependencies

Since compose file is missing, this is a reference example for local dependencies:

```yaml
services:
  mysql:
    image: mysql:8.4
    container_name: ecommerce-auth-mysql
    environment:
      MYSQL_ROOT_PASSWORD: root_password
      MYSQL_DATABASE: auth_db
      MYSQL_USER: auth_user
      MYSQL_PASSWORD: auth_password
    ports:
      - "3306:3306"
    volumes:
      - auth_mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "127.0.0.1", "-uroot", "-proot_password"]
      interval: 10s
      timeout: 5s
      retries: 10
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    container_name: ecommerce-auth-redis
    command: ["redis-server", "--appendonly", "yes"]
    ports:
      - "6379:6379"
    volumes:
      - auth_redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "PING"]
      interval: 10s
      timeout: 5s
      retries: 10
    restart: unless-stopped

volumes:
  auth_mysql_data:
  auth_redis_data:
```

Commands if a compose file exists:

```bash
docker compose up -d
docker compose ps
docker compose logs mysql
docker compose logs redis
docker compose down
```

### Docker concepts

| Concept | Explanation |
|---|---|
| Container | Running isolated process, jaise MySQL server. |
| Image | Template from which container starts. |
| Volume | Persistent storage. Container delete hone par data safe rehta hai. |
| Network | Containers ek dusre ko service name se access kar sakte hain. |
| Restart policy | Crash/reboot ke baad container auto-start behavior. |

---

## 9. Local Development Setup

Follow this from a fresh clone.

### Step 1: Clone repository

```bash
git clone <your-repo-url>
cd Ecommerce
```

### Step 2: Verify Go workspace

```bash
cd backend
go work sync
```

### Step 3: Install Go dependencies

```bash
cd services/auth-service
go mod download
```

### Step 4: Start MySQL and Redis

Using Docker:

```bash
docker start ecommerce-auth-mysql ecommerce-auth-redis
```

If containers do not exist, create them using the Docker commands in section 8.

### Step 5: Run MySQL migrations

```bash
cd backend/services/auth-service
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/001_create_auth_tables.up.sql
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/002_add_token_claim_metadata.up.sql
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/003_add_otp_account_fk.up.sql
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/004_harden_role_assignments.up.sql
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/005_create_auth_outbox_events.up.sql
```

### Step 6: Generate JWT keys

```bash
cd backend/services/auth-service
mkdir -p secrets
openssl genrsa -out secrets/jwt-private.pem 2048
openssl rsa -in secrets/jwt-private.pem -pubout -out secrets/jwt-public.pem
chmod 600 secrets/jwt-private.pem
chmod 644 secrets/jwt-public.pem
```

### Step 7: Create local `.env`

Create:

```text
backend/services/auth-service/.env
```

Use the example from section 7.

For first run, keep:

```env
SESSION_LINK_MODE=disabled
```

### Step 8: Load environment variables

```bash
cd backend/services/auth-service
set -a
source .env
set +a
```

### Step 9: Run tests

```bash
env GOCACHE=/tmp/auth-go-cache go test ./...
```

### Step 10: Start Auth Service

```bash
go run ./cmd/server
```

Expected log contains:

```text
auth.http.started
```

### Step 11: Verify health

In another terminal:

```bash
curl -i http://localhost:8081/healthz
```

Expected:

```json
{"status":"ok"}
```

### Step 12: Optional local bootstrap user

Current service has internal credential APIs, but no public signup API route in the current HTTP router. For local login testing, create an account row, create credential, and add a role.

Insert account and role:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p auth_db
```

```sql
INSERT INTO auth_accounts (
  account_id,
  user_id,
  email,
  email_verified,
  status
) VALUES (
  'acct_dev_001',
  'user_dev_001',
  'dev@example.com',
  TRUE,
  'active'
) ON DUPLICATE KEY UPDATE
  user_id = VALUES(user_id),
  email_verified = TRUE,
  status = 'active';

INSERT INTO role_assignments (
  account_id,
  role,
  reason
) VALUES (
  'acct_dev_001',
  'buyer',
  'local bootstrap'
) ON DUPLICATE KEY UPDATE
  reason = VALUES(reason),
  revoked_at = NULL;
```

Create password credential through internal endpoint:

```bash
curl -i -X POST http://localhost:8081/internal/v1/auth/credentials \
  -H "Content-Type: application/json" \
  -d '{"account_id":"acct_dev_001","password":"DevPassword123!"}'
```

Login:

```bash
curl -i -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"identifier":"dev@example.com","password":"DevPassword123!"}'
```

Note: This bootstrap is local-only. Production signup should go through proper User Service/Auth orchestration.

---

## 10. Running the Project

### Normal local run

Terminal 1:

```bash
docker start ecommerce-auth-mysql ecommerce-auth-redis
```

Terminal 2:

```bash
cd backend/services/auth-service
set -a
source .env
set +a
go run ./cmd/server
```

Terminal 3:

```bash
curl http://localhost:8081/healthz
```

### Build binary

```bash
cd backend/services/auth-service
go build -o auth-service ./cmd/server
```

Run binary:

```bash
set -a
source .env
set +a
./auth-service
```

### Key HTTP endpoints

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/healthz` | Health check |
| `GET` | `/.well-known/jwks.json` | Public JWT keys |
| `POST` | `/api/v1/auth/login` | Login with identifier/password |
| `POST` | `/api/v1/auth/refresh` | Rotate refresh token |
| `POST` | `/api/v1/auth/logout` | Revoke refresh tokens |
| `POST` | `/api/v1/auth/otp/send` | Create/send OTP challenge |
| `POST` | `/api/v1/auth/otp/verify` | Verify OTP |
| `POST` | `/api/v1/auth/password/forgot` | Start password reset OTP |
| `POST` | `/internal/v1/auth/credentials` | Create password credential |
| `POST` | `/internal/v1/auth/password/verify` | Verify password internally |
| `POST` | `/internal/v1/auth/password/reset` | Reset password internally |
| `POST` | `/internal/v1/auth/tokens/issue` | Issue token pair internally |
| `POST` | `/internal/v1/auth/tokens/verify` | Verify access token internally |
| `GET` | `/internal/v1/auth/roles` | Get roles, requires bearer token |
| `POST` | `/internal/v1/auth/roles/assign` | Assign role, requires privileged bearer token |
| `POST` | `/internal/v1/auth/roles/revoke` | Revoke role, requires privileged bearer token |

---

## 11. Ports & Networking

| Service | Default port | Purpose | Config |
|---|---:|---|---|
| Auth Service HTTP | `8081` | Main Auth API | `AUTH_HTTP_ADDR=:8081` |
| MySQL | `3306` | Auth database | `AUTH_MYSQL_DSN` |
| Redis | `6379` | OTP rate limits | `AUTH_REDIS_ADDR` |
| Notification Service | `8084` example | OTP delivery | `NOTIFICATION_OTP_ENDPOINT` |
| Session Service gRPC | `9090` example | Reserved config/currently not wired | `SESSION_GRPC_ADDR` |
| Event publish HTTP endpoint | No fixed port | Outbox worker publish target | `AUTH_EVENTS_PUBLISH_ENDPOINT` |

### Port conflict fixes

| Problem | Fix |
|---|---|
| Auth port `8081` already used | Change `AUTH_HTTP_ADDR=:8082`. |
| MySQL port `3306` already used | Docker map `-p 3307:3306` and update DSN to `127.0.0.1:3307`. |
| Redis port `6379` already used | Docker map `-p 6380:6379` and update `AUTH_REDIS_ADDR=127.0.0.1:6380`. |
| Docker container cannot connect to host | Use Docker network service names or `host.docker.internal` depending on OS. |
| Firewall blocks port | Allow local/private network access or bind to localhost only. |

### Local vs Docker networking

If Go app runs on host and DB runs in Docker:

```env
AUTH_MYSQL_DSN='auth_user:auth_password@tcp(127.0.0.1:3306)/auth_db?parseTime=true&charset=utf8mb4&loc=UTC'
AUTH_REDIS_ADDR=127.0.0.1:6379
```

If Go app also runs inside Docker Compose:

```env
AUTH_MYSQL_DSN='auth_user:auth_password@tcp(mysql:3306)/auth_db?parseTime=true&charset=utf8mb4&loc=UTC'
AUTH_REDIS_ADDR=redis:6379
```

---

## 12. Common Errors & Fixes

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `AUTH_MYSQL_DSN cannot be empty` | Env not set or `.env` not sourced | `set -a; source .env; set +a` | Remember app does not auto-load `.env`. |
| `auth.mysql.ping_failed` | MySQL down, wrong port, wrong user/password | Start MySQL, verify DSN, run `mysql -h 127.0.0.1 -P 3306 -u auth_user -p` | Use Docker healthcheck and correct DSN. |
| `unknown database auth_db` | Migrations/database not created | Run `001_create_auth_tables.up.sql` | Run migrations before app. |
| MySQL scan time error | Missing `parseTime=true` | Add `parseTime=true` to `AUTH_MYSQL_DSN` | Keep DSN template. |
| `Access denied for user` | Wrong MySQL credentials | Reset user password/grants or update DSN | Store one local `.env` source of truth. |
| `auth.redis.ping_failed` | Redis down or wrong password | Start Redis, set `AUTH_REDIS_PASSWORD` if required | Verify with `redis-cli PING`. |
| `NOAUTH Authentication required` | Redis requires password but env empty | Set `AUTH_REDIS_PASSWORD` | Match Docker Redis command and `.env`. |
| `JWT_KEY_ID cannot be empty` | Missing key id env | Set `JWT_KEY_ID=auth-local-1` | Include JWT vars in `.env`. |
| `read jwt private key` | File missing or wrong path | Generate keys or use absolute path | Run app from service directory or use absolute paths. |
| `JWT_SIGNING_ALG must be RS256` | Unsupported value | Set `JWT_SIGNING_ALG=RS256` | Do not change unless code supports more algorithms. |
| `REFRESH_TOKEN_PEPPER cannot be empty` | Missing secret | Set long random pepper | Keep secrets in env/secret manager. |
| `OTP_HASH_PEPPER cannot be empty` | Missing OTP secret | Set long random pepper | Keep OTP vars in `.env`. |
| `SESSION_EVENT_PEPPER cannot be empty` | `SESSION_LINK_MODE=outbox` but pepper missing | Set pepper or use `SESSION_LINK_MODE=disabled` locally | Use disabled for first local run. |
| OTP send returns `OTP_DELIVERY_UNAVAILABLE` | Notification endpoint not running or returns non-2xx | Start Notification Service or mock endpoint | Health check external services. |
| `OTP_RATE_LIMIT_UNAVAILABLE` | Redis issue during OTP flow | Fix Redis connection | Keep Redis stable. |
| `port already in use` | Another process on same port | Change port or stop process | Use `lsof -i :8081` or `netstat`. |
| Docker daemon not running | Docker Desktop/Engine stopped | Start Docker | Enable Docker startup. |
| `go test` fails with read-only Go cache | Build cache under read-only home | `GOCACHE=/tmp/auth-go-cache go test ./...` | Set writable `GOCACHE` in restricted environments. |
| `go mod download` fails | Network/proxy problem | Check internet, set `GOPROXY` | Configure company proxy once. |
| Login returns `INVALID_CREDENTIALS` | Account/credential missing or wrong password | Seed account, create credential, verify email/phone/status | Use local bootstrap steps. |
| Login returns role/subject error | Account lacks `user_id` or active role | Add `user_id` and role assignment | Seed complete auth account data. |
| Re-running migrations gives duplicate column/table errors | Migration already applied | Do not rerun same migration on same DB | Use migration tool or fresh DB for local reset. |

---

## 13. Security & Best Practices

### Security checklist

- Never commit `.env`.
- Never commit `secrets/jwt-private.pem`.
- Use strong random values for:
  - `REFRESH_TOKEN_PEPPER`
  - `OTP_HASH_PEPPER`
  - `OTP_RATE_LIMIT_PEPPER`
  - `SESSION_EVENT_PEPPER`
- Use `SESSION_LINK_MODE=disabled` only for local debugging.
- Use `SESSION_LINK_MODE=outbox` in integrated/staging/prod.
- Use MySQL app user with limited privileges, not root.
- Use Redis password/TLS in production.
- Use short access token TTL, for example `15m`.
- Rotate JWT keys with new `JWT_KEY_ID`.
- Keep refresh token pepper in secret manager.
- Keep DB backups before migrations.
- Do not log passwords, OTPs, refresh tokens, or private keys.
- Bind local service to `127.0.0.1:8081` if you do not want LAN access.

### Configuration audit from current implementation

| Finding | Risk | Recommendation |
|---|---|---|
| No `.env.example` found | Beginners may miss required variables | Add a sanitized `.env.example` based on section 7. |
| No `.gitignore` found | `.env` and PEM files can be accidentally committed | Add `.env`, `*.pem`, `secrets/`, build outputs to `.gitignore`. |
| No Dockerfile found | App cannot be containerized directly | Add service Dockerfile before deployment. |
| No docker-compose found | Local dependency setup manual | Add local compose for MySQL/Redis and optional notification mock. |
| No migration runner found | Manual migration order can be skipped/misordered | Add `make migrate-up` or use a standard migration tool. |
| `.env` not auto-loaded | User can create `.env` but app still fails | Document `source .env` or add dev-only loader/Makefile. |
| `/healthz` only returns process health | It does not verify MySQL/Redis | Add readiness endpoint checking DB/Redis for Kubernetes. |
| Notification client sends no auth header | Internal endpoint can be abused if exposed | Use internal network, mTLS, or service auth token. |
| `AUTH_EVENTS_PUBLISH_ENDPOINT` empty with outbox mode | Pending events can accumulate | Set endpoint or monitor pending/dead-letter rows. |
| `SESSION_GRPC_ADDR` exists but not used | Config confusion | Update docs or implement gRPC mode. |
| Task docs mention Redis Go client | Current code uses custom Redis TCP client | Keep docs aligned with implementation. |
| Task docs mention Kafka/RabbitMQ | Current code has HTTP publisher only | Clarify broker is behind event-ingress, not direct Auth dependency. |
| Login endpoint has DB lockout but no Redis/IP login rate limiter | More brute-force exposure | Add gateway/auth middleware rate limit for login. |

### Production notes

Production deployment should use:

- Kubernetes Secret or cloud secret manager for all secrets.
- Managed MySQL with backups and monitoring.
- Managed Redis with password/TLS.
- Internal-only Notification endpoint.
- Outbox publisher endpoint with retries, DLQ, and alerts.
- Structured logs shipped to logging platform.
- Metrics for login failures, OTP failures, refresh reuse, outbox lag.
- Readiness and liveness probes.

---

## 14. Missing or Misconfigured Things

These are not blockers for reading code, but they matter for smooth onboarding.

| Missing/misconfigured item | Impact |
|---|---|
| `.env.example` | New developers do not know required env names. |
| `.gitignore` | High chance of committing secrets or generated binaries. |
| Dockerfile | Cannot build Auth Service container from repo yet. |
| `docker-compose.yml` | Developers must manually run MySQL/Redis. |
| Migration command or Makefile | Migrations are manual and error-prone. |
| Seed script | Local login testing requires manual SQL/API bootstrap. |
| Notification mock | OTP flow cannot be tested unless Notification Service exists. |
| Readiness endpoint | `/healthz` does not prove DB/Redis connectivity. |
| Broker integration | Kafka/RabbitMQ are architectural docs only, not direct code dependency. |
| gRPC mode | `SESSION_GRPC_ADDR` is configured but current validation allows only `outbox` and `disabled`. |

Recommended future files:

```text
.gitignore
backend/services/auth-service/.env.example
backend/services/auth-service/Dockerfile
backend/services/auth-service/docker-compose.local.yml
backend/services/auth-service/Makefile
backend/services/auth-service/scripts/migrate-up.sh
backend/services/auth-service/scripts/seed-local-user.sql
```

---

## 15. Final Checklist

Use this checklist before saying "Auth Service local setup done".

- [ ] Git installed.
- [ ] Go installed and `go version` is compatible with `go.mod`.
- [ ] Docker installed if using containerized MySQL/Redis.
- [ ] MySQL running on expected port.
- [ ] Redis running on expected port.
- [ ] MySQL migrations `001` to `005` applied.
- [ ] `auth_db` has all required tables.
- [ ] JWT RSA private/public keys generated.
- [ ] Local `.env` created in `backend/services/auth-service`.
- [ ] `.env` sourced before running app.
- [ ] `AUTH_MYSQL_DSN` includes `parseTime=true`.
- [ ] `REFRESH_TOKEN_PEPPER`, `OTP_HASH_PEPPER`, and other secrets set.
- [ ] `SESSION_LINK_MODE=disabled` for first beginner local run, or outbox configured properly.
- [ ] `go mod download` completed.
- [ ] `go test ./...` passes, using `GOCACHE=/tmp/auth-go-cache` if needed.
- [ ] Auth Service starts successfully.
- [ ] `curl http://localhost:8081/healthz` returns `{"status":"ok"}`.
- [ ] Optional local user seeded if login testing is needed.
- [ ] `.env` and `secrets/*.pem` are not committed.

---

## Quick Start Summary

Fast path for a beginner:

```bash
git clone <your-repo-url>
cd Ecommerce/backend/services/auth-service

docker start ecommerce-auth-mysql ecommerce-auth-redis

mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/001_create_auth_tables.up.sql
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/002_add_token_claim_metadata.up.sql
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/003_add_otp_account_fk.up.sql
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/004_harden_role_assignments.up.sql
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/005_create_auth_outbox_events.up.sql

mkdir -p secrets
openssl genrsa -out secrets/jwt-private.pem 2048
openssl rsa -in secrets/jwt-private.pem -pubout -out secrets/jwt-public.pem

set -a
source .env
set +a

env GOCACHE=/tmp/auth-go-cache go test ./...
go run ./cmd/server
```

Then verify:

```bash
curl http://localhost:8081/healthz
```
