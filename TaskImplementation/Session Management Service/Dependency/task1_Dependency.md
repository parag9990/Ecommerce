# Project Dependency & Setup Guide

This guide is for `TaskImplementation/Session Management Service/task1.md` and the currently implemented Session Management Service under:

```text
backend/services/session-service
```

Goal simple hai: ek beginner developer repo clone karke dependencies install kare, MongoDB/Redis setup kare, `.env` configure kare, migrations run kare, backend start kare, aur common issues debug kar sake.

Important: ye document business logic rewrite nahi karta. Ye sirf dependency, setup, environment, database, Docker, and DevOps onboarding guide hai.

---

## 1. Project Overview

Session Management Service ecommerce platform ka independent Go backend service hai. Iska kaam user sessions, anonymous sessions, activity events, active sessions, and journey timeline ko handle karna hai.

Current implementation me ye cheezein present hain:

| Area | Current status |
|---|---|
| Language | Go |
| HTTP server | Go standard `net/http` |
| Config loading | Environment variables via `os.Getenv` |
| Durable database | MongoDB |
| Cache / hot state | Redis |
| Migrations | MongoDB JavaScript files for `mongosh` |
| Health endpoints | `/healthz`, `/readyz` |
| Event ingestion endpoint | `POST /api/v1/sessions/events` |
| Journey endpoint | `GET /api/v1/analytics/sessions/{session_id}/journey` |
| Dockerfile / compose | Not committed for this service yet |
| Kafka/RabbitMQ | Planned in docs, not wired in current code |

Simple Hinglish explanation:

Session service user ke browser/app activity ko track karta hai. MongoDB long-term data store karta hai, Redis fast active-session state store karta hai. Backend Go me hai, aur HTTP APIs expose karta hai.

---

## 2. Tech Stack

### Go

Go ek compiled backend programming language hai. Is project me Go use hua hai because microservices fast, simple, and deploy-friendly banane ke liye Go ka ecosystem strong hai.

| Detail | Value |
|---|---|
| Required? | Yes |
| Version in repo | `go 1.26.3` |
| Used for | Session service backend |
| Main files | `backend/services/session-service/cmd/server/main.go` |

Beginner note: Go me code run karne ke liye `go run`, test ke liye `go test`, build ke liye `go build` use hota hai.

### Go standard `net/http`

`net/http` Go ka built-in HTTP package hai. Is project me Gin/Fiber/Echo use nahi hua. Service pure Go standard library HTTP server use karta hai.

| Detail | Value |
|---|---|
| Required? | Yes, built into Go |
| Used for | HTTP routes, server, request/response handling |
| Main routes | `/healthz`, `/readyz`, `/api/v1/sessions/events` |

Hinglish: `net/http` built-in web server hai. Extra framework install karne ki zarurat nahi.

### MongoDB

MongoDB ek document database hai jisme JSON-like documents store hote hain. Is project me sessions/events flexible data hote hain, isliye MongoDB suitable hai.

| Detail | Value |
|---|---|
| Required? | Yes |
| Default port | `27017` |
| Default database | `session_db` |
| Collections | `sessions`, `session_events`, `journey_summaries` |
| Go dependency | `go.mongodb.org/mongo-driver/v2 v2.6.0` |

Hinglish: MongoDB me tables ki jagah collections hoti hain. Session events ka shape flexible ho sakta hai, isliye MongoDB helpful hai.

### Redis

Redis ek in-memory data store/cache hai. Is project me active sessions fast read/write ke liye Redis use hota hai.

| Detail | Value |
|---|---|
| Required? | Yes for current app startup |
| Default port | `6379` |
| Default DB | `2` |
| Go dependency | `github.com/redis/go-redis/v9 v9.19.0` |
| Used for | Active sessions, user/anonymous session sets, active counters |

Hinglish: Redis memory me data rakhta hai, isliye bahut fast hota hai. Live active sessions ke liye useful hai.

### MongoDB migration scripts

Migration files JavaScript me hain and `mongosh` se run hote hain.

| File | Purpose |
|---|---|
| `migrations/001_session_storage_indexes.up.js` | `sessions` and `session_events` collections, validators, indexes, TTL index |
| `migrations/001_session_storage_indexes.down.js` | Indexes/validators rollback |
| `migrations/002_journey_summaries.up.js` | `journey_summaries` collection and indexes |
| `migrations/002_journey_summaries.down.js` | Drops `journey_summaries` |

Note: App startup bhi indexes ensure karta hai, but JSON schema validators ke liye migrations run karna recommended hai.

### Docker

Docker optional hai, but beginners ke liye MongoDB and Redis run karne ka easiest way hai.

| Detail | Value |
|---|---|
| Required? | Optional but recommended |
| Current repo status | No service Dockerfile or compose file committed for session service |
| Recommended use | Run MongoDB + Redis locally |

### Kafka/RabbitMQ

Project architecture docs me Kafka/RabbitMQ planned hai for async events, but current Session Service code me message queue client configured nahi hai.

| Detail | Value |
|---|---|
| Required now? | No |
| Current code | Publisher is `nil` |
| Future topic | `session.events` |

---

## 3. Required Software

Install these before running the service.

| Software | Required? | Why needed | Verify command |
|---|---:|---|---|
| Git | Yes | Repo clone karne ke liye | `git --version` |
| Go 1.26.3 or newer | Yes | Backend run/build/test ke liye | `go version` |
| MongoDB | Yes | Session/event durable storage | `mongosh --eval "db.runCommand({ ping: 1 })"` |
| mongosh | Recommended | Migrations and DB verification | `mongosh --version` |
| Redis | Yes | Active session cache | `redis-cli ping` |
| redis-cli | Recommended | Redis verify/debug ke liye | `redis-cli --version` |
| Docker Desktop / Docker Engine | Optional | Local MongoDB/Redis containers | `docker version` |
| Docker Compose plugin | Optional | Multi-container local stack | `docker compose version` |
| curl / Postman | Recommended | API test karne ke liye | `curl --version` |

---

## 4. Dependency Management

This is a Go project, so dependencies `go.mod` and `go.sum` se manage hoti hain.

### Important files

| File | Meaning |
|---|---|
| `backend/go.work` | Go workspace file. Multiple Go services ko ek workspace me connect karta hai. |
| `backend/services/session-service/go.mod` | Session service ka module name, Go version, direct dependencies. |
| `backend/services/session-service/go.sum` | Dependency checksums. Security and reproducible builds ke liye important. |

### Current Go module

```text
module github.com/example/ecommerce-platform/backend/services/session-service
go 1.26.3
```

Direct dependencies:

| Dependency | Version | Why used |
|---|---:|---|
| `go.mongodb.org/mongo-driver/v2` | `v2.6.0` | MongoDB connect, collections, indexes, CRUD |
| `github.com/redis/go-redis/v9` | `v9.19.0` | Redis connect, ping, hashes, sets, TTL |

Indirect dependencies are transitive packages required by MongoDB/Redis drivers. Usually unko manually edit nahi karna hota.

### Install/download dependencies

From service directory:

```bash
cd backend/services/session-service
go mod download
```

From backend workspace:

```bash
cd backend
go work sync
go test ./services/session-service/...
```

### Clean dependency file

Use this after adding/removing Go imports:

```bash
cd backend/services/session-service
go mod tidy
```

Hinglish: `go mod tidy` unused dependencies hata deta hai aur missing dependencies add kar deta hai.

### Run tests

```bash
cd backend/services/session-service
go test ./...
```

### Run service

```bash
cd backend/services/session-service
go run ./cmd/server
```

### Build binary

```bash
cd backend/services/session-service
go build -o bin/session-service ./cmd/server
./bin/session-service
```

On Windows PowerShell:

```powershell
cd backend/services/session-service
go build -o bin/session-service.exe ./cmd/server
.\bin\session-service.exe
```

### Common Go dependency issues

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `go: go.mod requires go >= 1.26.3` | Old Go installed | Install Go 1.26.3 or newer | Check `go version` before setup |
| `missing go.sum entry` | Dependency checksum missing | Run `go mod tidy` | Commit `go.sum` changes with dependency changes |
| `module lookup disabled by GOPROXY` | Bad Go proxy env | `go env -w GOPROXY=https://proxy.golang.org,direct` | Do not set broken proxy globally |
| `connection refused` during tests/run | MongoDB/Redis not running | Start MongoDB and Redis | Start dependencies before backend |
| Import path errors | Running from wrong folder or broken workspace | Run from `backend/services/session-service` or `backend` | Keep `backend/go.work` synced |

---

## 5. Database Setup

Detected databases for current Session Service:

| Database | Required? | Purpose |
|---|---:|---|
| MongoDB | Yes | Durable sessions, events, journey summaries |
| Redis | Yes | Active sessions and counters |
| MySQL | No | Used by Auth Service, not required for this Session Service |
| PostgreSQL | No | Not used |
| SQLite | No | Not used |

---

### 5.1 MongoDB Setup

#### A. What MongoDB is

MongoDB ek NoSQL document database hai. Data collections and documents me store hota hai. Document JSON jaisa hota hai.

Simple English: MongoDB is a flexible database that stores JSON-like documents instead of fixed SQL rows.

#### B. Why this project uses MongoDB

Session events high-volume and flexible hote hain. Example: `page_view`, `click`, `search`, `payment_result` sabke properties different ho sakte hain. MongoDB flexible event data ke liye good fit hai.

This service uses MongoDB for:

| Collection | Purpose |
|---|---|
| `sessions` | One visit/session summary |
| `session_events` | Raw events like page views, clicks, searches |
| `journey_summaries` | Calculated session journey summary |

#### C. Required or optional

MongoDB is required. App startup MongoDB ping karta hai. MongoDB down hua to service start fail ho jayegi.

#### D. Local installation

Recommended beginner approach: Docker use karo. Native install OS ke hisaab se different hota hai.

Windows:

1. Install Docker Desktop, then use Docker command below.
2. Native install chahiye to MongoDB Community Server and MongoDB Shell (`mongosh`) install karo.
3. Windows Services me MongoDB service running verify karo.
4. Verify:

```powershell
mongosh --eval "db.runCommand({ ping: 1 })"
```

Linux Ubuntu/Debian:

```bash
sudo apt update
sudo apt install -y gnupg curl
```

For native MongoDB, use official MongoDB Community repo for your OS version. Beginner ke liye Docker easier and less error-prone hai.

If MongoDB is already installed:

```bash
sudo systemctl start mongod
sudo systemctl enable mongod
mongosh --eval "db.runCommand({ ping: 1 })"
```

macOS:

```bash
brew tap mongodb/brew
brew install mongodb-community@7.0
brew services start mongodb-community@7.0
mongosh --eval "db.runCommand({ ping: 1 })"
```

If Homebrew setup nahi hai, Docker use karo.

#### E. Docker setup

Create persistent volume:

```bash
docker volume create session_mongo_data
```

Run MongoDB:

```bash
docker run -d \
  --name ecommerce-session-mongo \
  -p 27017:27017 \
  -v session_mongo_data:/data/db \
  mongo:7
```

Verify container:

```bash
docker ps
docker logs ecommerce-session-mongo
```

#### F. Start commands

Docker:

```bash
docker start ecommerce-session-mongo
```

Linux service:

```bash
sudo systemctl start mongod
```

macOS Homebrew:

```bash
brew services start mongodb-community@7.0
```

#### G. Verify running

```bash
mongosh "mongodb://localhost:27017" --eval "db.runCommand({ ping: 1 })"
```

Expected:

```json
{ "ok": 1 }
```

Check database:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval "db.getCollectionNames()"
```

#### H. Default port

MongoDB default port is:

```text
27017
```

#### I. Connection string format

Local without auth:

```env
SESSION_MONGO_URI=mongodb://localhost:27017
SESSION_MONGO_DATABASE=session_db
```

With username/password:

```env
SESSION_MONGO_URI=mongodb://session_user:strong_password@localhost:27017/session_db?authSource=admin
SESSION_MONGO_DATABASE=session_db
```

Docker Compose internal networking:

```env
SESSION_MONGO_URI=mongodb://mongo:27017
SESSION_MONGO_DATABASE=session_db
```

#### J. Where to place credentials

Local development:

```text
backend/services/session-service/.env
```

Production:

Use secret manager, Kubernetes Secret, CI/CD secret variables, or platform environment variables. Do not hardcode credentials in Go files.

Variables:

| Variable | Put what here |
|---|---|
| `SESSION_MONGO_URI` | Host, port, username, password, auth options |
| `SESSION_MONGO_DATABASE` | Database name like `session_db` |
| `SESSION_MONGO_CONNECT_TIMEOUT` | Connect timeout like `5s` |

---

### 5.2 Redis Setup

#### A. What Redis is

Redis ek in-memory cache/data store hai. Ye RAM me data rakhta hai, so reads/writes very fast hote hain.

Simple English: Redis is used for fast temporary data, not as the main durable database.

#### B. Why this project uses Redis

Current code Redis me active sessions maintain karta hai:

| Redis key pattern | Purpose |
|---|---|
| `session:active:{session_id}` | Active session hash |
| `session:last_seen:{session_id}` | Last seen timestamp |
| `session:anon:{anonymous_id}` | Anonymous ID to session IDs |
| `session:user:{user_id}` | User ID to session IDs |
| `session:counter:active:{yyyyMMddHH}` | Hourly active session counter |

#### C. Required or optional

Redis is required for current startup. App Redis ping karta hai. Redis unavailable hua to service start fail ho jayegi.

#### D. Local installation

Windows:

Redis ka official native Windows server generally recommended nahi hota. Best options:

1. Docker Desktop use karo.
2. Ya WSL2 Ubuntu me Redis install karo.

Docker recommended:

```powershell
docker run -d --name ecommerce-session-redis -p 6379:6379 redis:7-alpine
docker exec -it ecommerce-session-redis redis-cli ping
```

Linux Ubuntu/Debian:

```bash
sudo apt update
sudo apt install -y redis-server
sudo systemctl start redis-server
sudo systemctl enable redis-server
redis-cli ping
```

macOS:

```bash
brew install redis
brew services start redis
redis-cli ping
```

#### E. Docker setup

Create persistent volume:

```bash
docker volume create session_redis_data
```

Run Redis:

```bash
docker run -d \
  --name ecommerce-session-redis \
  -p 6379:6379 \
  -v session_redis_data:/data \
  redis:7-alpine \
  redis-server --appendonly yes
```

With password:

```bash
docker run -d \
  --name ecommerce-session-redis \
  -p 6379:6379 \
  -v session_redis_data:/data \
  redis:7-alpine \
  redis-server --appendonly yes --requirepass dev-redis-password
```

If password is enabled:

```env
SESSION_REDIS_PASSWORD=dev-redis-password
```

#### F. Start commands

Docker:

```bash
docker start ecommerce-session-redis
```

Linux service:

```bash
sudo systemctl start redis-server
```

macOS Homebrew:

```bash
brew services start redis
```

#### G. Verify running

No password:

```bash
redis-cli ping
```

Expected:

```text
PONG
```

With password:

```bash
redis-cli -a dev-redis-password ping
```

Check DB 2:

```bash
redis-cli -n 2 keys 'session:*'
```

#### H. Default port

Redis default port is:

```text
6379
```

#### I. Connection string format

This project does not use a Redis URL. It uses address/password/db separately:

```env
SESSION_REDIS_ADDR=localhost:6379
SESSION_REDIS_PASSWORD=
SESSION_REDIS_DB=2
```

Docker Compose internal networking:

```env
SESSION_REDIS_ADDR=redis:6379
SESSION_REDIS_PASSWORD=
SESSION_REDIS_DB=2
```

#### J. Where to place credentials

Local:

```text
backend/services/session-service/.env
```

Production:

Use Kubernetes Secret, secret manager, or platform env vars.

Variables:

| Variable | Put what here |
|---|---|
| `SESSION_REDIS_ADDR` | Host and port |
| `SESSION_REDIS_PASSWORD` | Redis password, blank only for local dev |
| `SESSION_REDIS_DB` | Numeric Redis database, default `2` |
| `SESSION_REDIS_KEY_PREFIX` | Key prefix, default `session` |

---

## 6. Redis / Queue / External Services

### Redis

Redis setup details are above because it is a required datastore for this service.

Health check:

```bash
redis-cli ping
curl http://localhost:8086/readyz
```

Common Redis issues:

| Issue | Cause | Fix |
|---|---|---|
| `dial tcp localhost:6379: connect: connection refused` | Redis not running | Start Redis container/service |
| `NOAUTH Authentication required` | Redis password enabled but `.env` blank | Set `SESSION_REDIS_PASSWORD` |
| `WRONGPASS invalid username-password pair` | Wrong password | Update `.env` and restart service |
| Keys not visible | Wrong Redis DB selected | Use `redis-cli -n 2 keys 'session:*'` |

### Kafka / RabbitMQ

Kafka/RabbitMQ current code me mandatory nahi hai.

Project docs future me `session.events` publish karne ki baat karte hain, but `internal/app/dependencies.go` currently publisher as `nil` pass karta hai. Isliye local setup me Kafka/RabbitMQ start karne ki zarurat nahi.

Future production note:

| Service | Future purpose | Current required? |
|---|---|---:|
| Kafka | High-throughput event streaming | No |
| RabbitMQ | Simpler queues/routing | No |
| NATS | Not detected | No |
| MinIO/S3 | Not detected | No |
| SMTP/Twilio/Stripe/Firebase | Not detected | No |
| Nginx/API Gateway | Recommended in production | Not required locally |

### API Gateway / Auth headers

Current service reads identity and role from headers:

| Header | Purpose |
|---|---|
| `X-User-ID`, `X-Authenticated-User-ID`, `X-Auth-User-ID` | Trusted user ID from gateway/auth layer |
| `X-User-Roles`, `X-User-Role`, `X-Authenticated-Roles`, `X-Auth-Roles` | Admin role check for journey endpoint |
| `X-Client-Channel` | Client channel, example `user_app_web` |
| `X-Device-Type` | Device type, example `desktop` |
| `X-IP-Hash` | Pre-computed IP hash if gateway provides it |
| `X-IP-Version` | `ipv4`, `ipv6`, `unknown` |
| `X-Forwarded-For`, `X-Real-IP` | Client IP forwarding |
| `X-Request-ID`, `X-Trace-ID`, `Traceparent` | Request tracing |

Security note: Production me API Gateway ko spoofed headers strip/validate karne chahiye. Direct public traffic ko trusted identity headers set karne nahi dena chahiye.

---

## 7. Environment Variables

### Where `.env` should be created

Create this file:

```text
backend/services/session-service/.env
```

There is already an example file:

```text
backend/services/session-service/.env.example
```

Copy it:

```bash
cd backend/services/session-service
cp .env.example .env
```

Windows PowerShell:

```powershell
cd backend/services/session-service
Copy-Item .env.example .env
```

### Very important: `.env` is not auto-loaded

Current Go code uses `os.Getenv`. It does not import `godotenv`. That means simply creating `.env` is not enough. You must load/export variables before running.

Linux/macOS:

```bash
cd backend/services/session-service
set -a
source .env
set +a
go run ./cmd/server
```

Windows PowerShell:

```powershell
cd backend/services/session-service
Get-Content .env | ForEach-Object {
  if ($_ -match '^\s*#' -or $_ -match '^\s*$') { return }
  $name, $value = $_ -split '=', 2
  [Environment]::SetEnvironmentVariable($name.Trim(), $value.Trim(), 'Process')
}
go run ./cmd/server
```

### Complete `.env` example

Use this for local development:

```env
# HTTP server
SESSION_HTTP_ADDR=:8086
SESSION_HTTP_READ_TIMEOUT=5s
SESSION_HTTP_WRITE_TIMEOUT=10s
SESSION_HTTP_IDLE_TIMEOUT=60s
SESSION_SHUTDOWN_TIMEOUT=10s

# Session model defaults
SESSION_SCHEMA_VERSION=1
SESSION_INACTIVITY_TIMEOUT=30m
SESSION_DEFAULT_STATUS=active
SESSION_DEFAULT_CHANNEL=unknown
SESSION_DEFAULT_DEVICE_TYPE=unknown
SESSION_DEFAULT_IP_VERSION=unknown
SESSION_DEFAULT_RISK_LEVEL=unknown

# MongoDB
SESSION_MONGO_URI=mongodb://localhost:27017
SESSION_MONGO_DATABASE=session_db
SESSION_MONGO_CONNECT_TIMEOUT=5s

# Redis
SESSION_REDIS_ADDR=localhost:6379
SESSION_REDIS_PASSWORD=
SESSION_REDIS_DB=2
SESSION_REDIS_DIAL_TIMEOUT=2s
SESSION_REDIS_READ_TIMEOUT=2s
SESSION_REDIS_WRITE_TIMEOUT=2s
SESSION_REDIS_KEY_PREFIX=session

# TTL and retention
SESSION_ACTIVE_TTL=35m
SESSION_ACTIVE_COUNTER_TTL=48h
SESSION_RAW_EVENT_TTL_DAYS=90

# API limits and clock rules
SESSION_INGEST_MAX_BODY_BYTES=65536
SESSION_ALLOWED_CLOCK_SKEW=5m
SESSION_MAX_EVENT_AGE=24h
SESSION_IP_HASH_SALT=replace-this-with-a-long-random-secret

# List and journey limits
SESSION_DEFAULT_LIST_LIMIT=100
SESSION_MAX_LIST_LIMIT=500
SESSION_JOURNEY_DEFAULT_LIMIT=1000
SESSION_JOURNEY_MAX_LIMIT=1000
SESSION_JOURNEY_SUMMARY_TOP_PATHS_LIMIT=10
SESSION_JOURNEY_SUMMARY_UPSERT_ENABLED=true

# Validation limits
SESSION_MAX_ID_LENGTH=128
SESSION_MAX_HASH_LENGTH=256
SESSION_MAX_PAGE_LENGTH=2048
SESSION_MAX_REFERRER_LENGTH=2048
SESSION_MAX_USER_AGENT_LENGTH=1024
SESSION_MAX_METADATA_LENGTH=128
SESSION_MAX_UTM_VALUE_LENGTH=256
SESSION_MAX_GEO_VALUE_LENGTH=128
SESSION_MAX_RISK_REASONS=20
SESSION_MAX_RISK_REASON_TEXT_LENGTH=256
SESSION_MAX_EVENT_TYPE_LENGTH=64
SESSION_MAX_EVENT_PROPERTY_KEY_LENGTH=128
SESSION_MAX_EVENT_PROPERTY_STRING_LENGTH=2048
SESSION_MAX_EVENT_PROPERTIES=50
SESSION_MAX_EVENT_PROPERTY_DEPTH=6
```

You can also use older compatibility variables:

| Preferred | Also supported |
|---|---|
| `SESSION_INACTIVITY_TIMEOUT=30m` | `SESSION_INACTIVITY_TIMEOUT_SECONDS=1800` |
| `SESSION_ACTIVE_TTL=35m` | `SESSION_ACTIVE_TTL_SECONDS=2100` |
| `SESSION_ACTIVE_COUNTER_TTL=48h` | `SESSION_ACTIVE_COUNTER_TTL_SECONDS=172800` |
| `SESSION_ALLOWED_CLOCK_SKEW=5m` | `SESSION_ALLOWED_CLOCK_SKEW_SECONDS=300` |
| `SESSION_MAX_EVENT_AGE=24h` | `SESSION_MAX_EVENT_AGE_HOURS=24` |

### Environment variable table

| Variable | Required? | Example | Purpose | Security note |
|---|---:|---|---|---|
| `SESSION_HTTP_ADDR` | Yes | `:8086` | Backend listen address | Avoid exposing directly in prod without gateway |
| `SESSION_HTTP_READ_TIMEOUT` | Optional | `5s` | Request read timeout | Helps slow-client protection |
| `SESSION_HTTP_WRITE_TIMEOUT` | Optional | `10s` | Response write timeout | Helps avoid hung requests |
| `SESSION_HTTP_IDLE_TIMEOUT` | Optional | `60s` | Keep-alive idle timeout | Tune behind load balancer |
| `SESSION_SHUTDOWN_TIMEOUT` | Optional | `10s` | Graceful shutdown wait | Increase for slower shutdown |
| `SESSION_SCHEMA_VERSION` | Yes | `1` | Current session schema version | Must remain `1` for current code |
| `SESSION_INACTIVITY_TIMEOUT` | Optional | `30m` | Session inactivity rule | Keep product/team aligned |
| `SESSION_DEFAULT_STATUS` | Optional | `active` | Default session status | Must be valid enum |
| `SESSION_DEFAULT_CHANNEL` | Optional | `unknown` | Default traffic channel | Must be valid enum |
| `SESSION_DEFAULT_DEVICE_TYPE` | Optional | `unknown` | Default device type | Must be valid enum |
| `SESSION_DEFAULT_IP_VERSION` | Optional | `unknown` | Default IP version | Must be `ipv4`, `ipv6`, or `unknown` |
| `SESSION_DEFAULT_RISK_LEVEL` | Optional | `unknown` | Default risk level | Must be valid enum |
| `SESSION_MONGO_URI` | Yes | `mongodb://localhost:27017` | MongoDB connection URI | Use auth/TLS in prod |
| `SESSION_MONGO_DATABASE` | Yes | `session_db` | MongoDB database name | Separate dev/stage/prod DBs |
| `SESSION_MONGO_CONNECT_TIMEOUT` | Optional | `5s` | Mongo connect/ping timeout | Keep small for fast fail |
| `SESSION_REDIS_ADDR` | Yes | `localhost:6379` | Redis host/port | Use private network in prod |
| `SESSION_REDIS_PASSWORD` | Optional local, required prod | `dev-redis-password` | Redis auth password | Never commit real password |
| `SESSION_REDIS_DB` | Optional | `2` | Redis logical DB number | Keep service-specific DB/prefix |
| `SESSION_REDIS_DIAL_TIMEOUT` | Optional | `2s` | Redis connect timeout | Small timeout improves fail-fast |
| `SESSION_REDIS_READ_TIMEOUT` | Optional | `2s` | Redis read timeout | Tune for network latency |
| `SESSION_REDIS_WRITE_TIMEOUT` | Optional | `2s` | Redis write timeout | Tune for network latency |
| `SESSION_REDIS_KEY_PREFIX` | Optional | `session` | Redis key prefix | Avoid whitespace |
| `SESSION_ACTIVE_TTL` | Optional | `35m` | Active session Redis TTL | Should exceed inactivity window slightly |
| `SESSION_ACTIVE_COUNTER_TTL` | Optional | `48h` | Active counter key TTL | Cost/memory control |
| `SESSION_RAW_EVENT_TTL_DAYS` | Optional | `90` | Mongo raw event TTL | Retention/compliance setting |
| `SESSION_INGEST_MAX_BODY_BYTES` | Optional | `65536` | Max event request body size | Abuse protection |
| `SESSION_ALLOWED_CLOCK_SKEW` | Optional | `5m` | Future timestamp tolerance | Prevent bad client clocks |
| `SESSION_MAX_EVENT_AGE` | Optional | `24h` | Old event accept window | Prevent stale analytics |
| `SESSION_IP_HASH_SALT` | Strongly required | Long random string | HMAC salt for IP hashing | Treat as secret; rotate carefully |
| `SESSION_DEFAULT_LIST_LIMIT` | Optional | `100` | Default list limit | Avoid huge reads |
| `SESSION_MAX_LIST_LIMIT` | Optional | `500` | Max list limit | Protect DB |
| `SESSION_JOURNEY_DEFAULT_LIMIT` | Optional | `1000` | Default journey events limit | Protect DB/API |
| `SESSION_JOURNEY_MAX_LIMIT` | Optional | `1000` | Max journey limit | Protect DB/API |
| `SESSION_JOURNEY_SUMMARY_TOP_PATHS_LIMIT` | Optional | `10` | Top paths in summary | Controls summary size |
| `SESSION_JOURNEY_SUMMARY_UPSERT_ENABLED` | Optional | `true` | Save journey summary on journey fetch | Disable only for debugging |

### Common `.env` mistakes

| Mistake | Result | Fix |
|---|---|---|
| `.env` created but not sourced | Defaults used, credentials ignored | Run `set -a; source .env; set +a` |
| Wrong folder | Service cannot find expected file because code does not read it anyway | Keep `.env` in service folder and source it |
| Spaces around values | Some shells include spaces in value | Use `KEY=value`, no extra spaces |
| Quoted duration incorrectly | Duration parse falls back to default | Use `5s`, `30m`, `24h` |
| Empty `SESSION_IP_HASH_SALT` | IP hash becomes `unavailable` | Set a long random secret |
| Committing `.env` | Secret leak | `.gitignore` already ignores `.env`; never force-add it |

---

## 8. Docker Setup

### Current repo status

Session Service currently does not have a committed Dockerfile or docker-compose file. Docker is still useful for MongoDB and Redis.

### Recommended beginner Docker Compose

Create this as a local-only file if needed:

```text
docker-compose.session.local.yml
```

Example:

```yaml
services:
  mongo:
    image: mongo:7
    container_name: ecommerce-session-mongo
    ports:
      - "27017:27017"
    volumes:
      - session_mongo_data:/data/db
    healthcheck:
      test: ["CMD", "mongosh", "--quiet", "--eval", "db.runCommand({ ping: 1 }).ok"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: ecommerce-session-redis
    command: ["redis-server", "--appendonly", "yes"]
    ports:
      - "6379:6379"
    volumes:
      - session_redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  session_mongo_data:
  session_redis_data:
```

Start:

```bash
docker compose -f docker-compose.session.local.yml up -d
```

Check:

```bash
docker compose -f docker-compose.session.local.yml ps
docker compose -f docker-compose.session.local.yml logs mongo
docker compose -f docker-compose.session.local.yml logs redis
```

Stop:

```bash
docker compose -f docker-compose.session.local.yml down
```

Stop and delete volumes:

```bash
docker compose -f docker-compose.session.local.yml down -v
```

Warning: `down -v` deletes MongoDB/Redis local data.

### Recommended service Dockerfile

Not currently committed. If containerizing the service later, use a shape like this:

```dockerfile
FROM golang:1.26.3-alpine AS builder
WORKDIR /src
COPY backend/services/session-service/go.mod backend/services/session-service/go.sum ./backend/services/session-service/
WORKDIR /src/backend/services/session-service
RUN go mod download
COPY backend/services/session-service ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/session-service ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /out/session-service /session-service
EXPOSE 8086
ENTRYPOINT ["/session-service"]
```

Container env should use service names:

```env
SESSION_HTTP_ADDR=:8086
SESSION_MONGO_URI=mongodb://mongo:27017
SESSION_MONGO_DATABASE=session_db
SESSION_REDIS_ADDR=redis:6379
SESSION_REDIS_DB=2
SESSION_IP_HASH_SALT=replace-this-with-a-long-random-secret
```

### Local vs Docker setup

| Approach | Best for | Pros | Cons |
|---|---|---|---|
| Native MongoDB/Redis | Experienced local dev | Faster direct tools | OS-specific setup issues |
| Docker MongoDB/Redis + Go on host | Beginners | Easy reset, simple debugging | Docker required |
| Everything in Docker | Production-like local | Same networking as deploy | Needs Dockerfile/compose |

Recommended for beginners: Run MongoDB and Redis in Docker, run Go service on host using `go run`.

---

## 9. Local Development Setup

### Step 1: Clone repository

```bash
git clone <repo-url>
cd Ecommerce
```

### Step 2: Verify Go

```bash
go version
```

Expected Go version should be `1.26.3` or newer.

### Step 3: Download Go dependencies

```bash
cd backend/services/session-service
go mod download
```

### Step 4: Start MongoDB and Redis

Docker quick start:

```bash
docker volume create session_mongo_data
docker volume create session_redis_data

docker run -d \
  --name ecommerce-session-mongo \
  -p 27017:27017 \
  -v session_mongo_data:/data/db \
  mongo:7

docker run -d \
  --name ecommerce-session-redis \
  -p 6379:6379 \
  -v session_redis_data:/data \
  redis:7-alpine \
  redis-server --appendonly yes
```

If containers already exist:

```bash
docker start ecommerce-session-mongo ecommerce-session-redis
```

### Step 5: Verify MongoDB and Redis

```bash
mongosh "mongodb://localhost:27017" --eval "db.runCommand({ ping: 1 })"
redis-cli ping
```

Expected:

```text
PONG
```

### Step 6: Create `.env`

```bash
cd backend/services/session-service
cp .env.example .env
```

Edit:

```env
SESSION_MONGO_URI=mongodb://localhost:27017
SESSION_MONGO_DATABASE=session_db
SESSION_REDIS_ADDR=localhost:6379
SESSION_REDIS_DB=2
SESSION_IP_HASH_SALT=replace-this-with-a-long-random-secret
```

Generate a local salt:

```bash
openssl rand -hex 32
```

If `openssl` unavailable, manually use a long random string for local dev.

### Step 7: Load environment variables

Linux/macOS:

```bash
cd backend/services/session-service
set -a
source .env
set +a
```

Windows PowerShell:

```powershell
cd backend/services/session-service
Get-Content .env | ForEach-Object {
  if ($_ -match '^\s*#' -or $_ -match '^\s*$') { return }
  $name, $value = $_ -split '=', 2
  [Environment]::SetEnvironmentVariable($name.Trim(), $value.Trim(), 'Process')
}
```

### Step 8: Run MongoDB migrations

From service folder:

```bash
cd backend/services/session-service
export SESSION_MONGO_DATABASE=session_db
mongosh "$SESSION_MONGO_URI" migrations/001_session_storage_indexes.up.js
mongosh "$SESSION_MONGO_URI" migrations/002_journey_summaries.up.js
```

If MongoDB is running in Docker and `mongosh` is not installed on host:

```bash
cd backend/services/session-service
docker exec -i -e SESSION_MONGO_DATABASE=session_db ecommerce-session-mongo mongosh < migrations/001_session_storage_indexes.up.js
docker exec -i -e SESSION_MONGO_DATABASE=session_db ecommerce-session-mongo mongosh < migrations/002_journey_summaries.up.js
```

Verify collections:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval "db.getCollectionNames()"
```

Expected collections:

```text
sessions
session_events
journey_summaries
```

### Step 9: Run tests

```bash
cd backend/services/session-service
go test ./...
```

### Step 10: Start backend

```bash
cd backend/services/session-service
go run ./cmd/server
```

Expected log:

```json
{"msg":"session.http.started","addr":":8086"}
```

### Step 11: Verify health

In another terminal:

```bash
curl http://localhost:8086/healthz
curl http://localhost:8086/readyz
```

Expected:

```json
{"status":"ok"}
```

Readiness should include Mongo and Redis:

```json
{"checks":{"mongo":"ok","redis":"ok"},"status":"ok"}
```

### Step 12: Send test event

Linux/macOS:

```bash
NOW=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

curl -i -X POST "http://localhost:8086/api/v1/sessions/events" \
  -H "Content-Type: application/json" \
  -H "X-Client-Channel: user_app_web" \
  -H "X-Device-Type: desktop" \
  -H "X-Request-ID: req_local_001" \
  -d "{
    \"event_type\": \"page_view\",
    \"anonymous_id\": \"anon_local_001\",
    \"session_id\": \"sess_local_001\",
    \"occurred_at\": \"$NOW\",
    \"path\": \"/\",
    \"properties\": {
      \"title\": \"Home\"
    }
  }"
```

Expected status:

```text
HTTP/1.1 202 Accepted
```

Response:

```json
{
  "accepted": true,
  "request_id": "req_local_001",
  "event_id": "evt_..."
}
```

Verify MongoDB:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval "db.session_events.findOne({ session_id: 'sess_local_001' })"
mongosh "mongodb://localhost:27017/session_db" --eval "db.sessions.findOne({ session_id: 'sess_local_001' })"
```

Verify Redis:

```bash
redis-cli -n 2 hgetall session:active:sess_local_001
redis-cli -n 2 ttl session:active:sess_local_001
```

---

## 10. Running the Project

### Daily run flow

Terminal 1: dependencies

```bash
docker start ecommerce-session-mongo ecommerce-session-redis
```

Terminal 2: backend

```bash
cd backend/services/session-service
set -a
source .env
set +a
go run ./cmd/server
```

Terminal 3: health check

```bash
curl http://localhost:8086/healthz
curl http://localhost:8086/readyz
```

### Run from backend workspace

```bash
cd backend
go test ./services/session-service/...
go run ./services/session-service/cmd/server
```

Note: env variables still need to be exported in the shell.

### Build and run binary

```bash
cd backend/services/session-service
go build -o bin/session-service ./cmd/server
set -a
source .env
set +a
./bin/session-service
```

### Stop service

Press:

```text
Ctrl+C
```

The service handles `SIGINT`/`SIGTERM` and tries graceful shutdown.

---

## 11. Ports & Networking

| Service | Port | Purpose | Change using |
|---|---:|---|---|
| Session Backend API | `8086` | HTTP APIs and health endpoints | `SESSION_HTTP_ADDR=:8087` |
| MongoDB | `27017` | Session/event database | `SESSION_MONGO_URI` |
| Redis | `6379` | Active session cache | `SESSION_REDIS_ADDR` |
| API Gateway | Usually `8080` in platform docs | Future gateway in front of services | Gateway config |
| Kafka/RabbitMQ | N/A now | Future async events | Not configured now |

### Port conflict

Check port usage:

Linux/macOS:

```bash
lsof -i :8086
lsof -i :27017
lsof -i :6379
```

Alternative:

```bash
ss -ltnp | grep ':8086'
```

Windows PowerShell:

```powershell
netstat -ano | findstr :8086
```

Change backend port:

```env
SESSION_HTTP_ADDR=:8087
```

Then:

```bash
go run ./cmd/server
curl http://localhost:8087/healthz
```

### Docker networking note

If Go service runs on your host machine:

```env
SESSION_MONGO_URI=mongodb://localhost:27017
SESSION_REDIS_ADDR=localhost:6379
```

If Go service runs inside Docker Compose:

```env
SESSION_MONGO_URI=mongodb://mongo:27017
SESSION_REDIS_ADDR=redis:6379
```

Hinglish: Container ke andar `localhost` ka matlab same container hota hai, host machine nahi. Compose service names use karo.

---

## 12. Common Errors & Fixes

### 1. `connect: connection refused` for MongoDB

Cause:

MongoDB running nahi hai ya wrong host/port set hai.

Fix:

```bash
docker ps
docker start ecommerce-session-mongo
mongosh "mongodb://localhost:27017" --eval "db.runCommand({ ping: 1 })"
```

Prevention:

Backend start karne se pehle MongoDB verify karo.

### 2. `ping mongo: server selection error`

Cause:

Wrong `SESSION_MONGO_URI`, Docker networking issue, or MongoDB still starting.

Fix:

```bash
echo "$SESSION_MONGO_URI"
docker logs ecommerce-session-mongo
```

If service runs on host, use:

```env
SESSION_MONGO_URI=mongodb://localhost:27017
```

If service runs in compose, use:

```env
SESSION_MONGO_URI=mongodb://mongo:27017
```

Prevention:

Use `/readyz` and Docker health checks.

### 3. Redis connection refused

Cause:

Redis not running or wrong `SESSION_REDIS_ADDR`.

Fix:

```bash
docker start ecommerce-session-redis
redis-cli ping
```

Prevention:

Keep Redis in compose with healthcheck.

### 4. `.env` values not applied

Cause:

Code does not auto-load `.env`.

Fix:

```bash
cd backend/services/session-service
set -a
source .env
set +a
go run ./cmd/server
```

Prevention:

Add this to your local run notes or use a task runner later.

### 5. `SESSION_SCHEMA_VERSION must be 1`

Cause:

Wrong schema version set in env.

Fix:

```env
SESSION_SCHEMA_VERSION=1
```

Prevention:

Do not change schema version unless code and migrations change together.

### 6. `Content-Type must be application/json`

Cause:

Event ingestion request missing JSON content type.

Fix:

```bash
curl -H "Content-Type: application/json" ...
```

Prevention:

Frontend SDK should always send JSON content type.

### 7. `occurred_at is too old`

Cause:

Event timestamp older than `SESSION_MAX_EVENT_AGE`, default `24h`.

Fix:

Use current UTC timestamp:

```bash
date -u +"%Y-%m-%dT%H:%M:%SZ"
```

Prevention:

Client should send fresh RFC3339 UTC timestamps.

### 8. `occurred_at is too far in the future`

Cause:

Client clock ahead beyond `SESSION_ALLOWED_CLOCK_SKEW`, default `5m`.

Fix:

Sync local system clock or increase skew for local debugging:

```env
SESSION_ALLOWED_CLOCK_SKEW=10m
```

Prevention:

Use NTP/time sync on clients and servers.

### 9. `PAYLOAD_TOO_LARGE`

Cause:

Request body exceeds `SESSION_INGEST_MAX_BODY_BYTES`, default `65536`.

Fix:

Reduce event properties payload.

Prevention:

Do not send huge DOM snapshots, screenshots, or form values.

### 10. `Permission denied` with Docker

Cause:

Current Linux user not in `docker` group.

Fix:

```bash
sudo usermod -aG docker "$USER"
newgrp docker
docker ps
```

Prevention:

Configure Docker after install.

### 11. `Docker daemon not running`

Cause:

Docker Desktop/Engine stopped.

Fix:

Linux:

```bash
sudo systemctl start docker
```

Windows/macOS:

Start Docker Desktop.

Prevention:

Enable Docker startup if you use it daily.

### 12. `mongosh: command not found`

Cause:

MongoDB shell not installed on host.

Fix:

Use Docker exec migration command:

```bash
docker exec -i -e SESSION_MONGO_DATABASE=session_db ecommerce-session-mongo mongosh < migrations/001_session_storage_indexes.up.js
```

Prevention:

Install `mongosh` or use MongoDB Docker container shell.

### 13. Journey endpoint returns `401` or `403`

Cause:

Journey API requires admin role headers.

Fix for local testing:

```bash
curl -H "X-User-ID: local-admin" \
  -H "X-User-Roles: admin" \
  http://localhost:8086/api/v1/analytics/sessions/sess_local_001/journey
```

Prevention:

In production, API Gateway/Auth Service should inject verified role headers.

### 14. `go mod download` failed

Cause:

Network/proxy issue or private module config issue.

Fix:

```bash
go env GOPROXY
go env -w GOPROXY=https://proxy.golang.org,direct
go clean -modcache
go mod download
```

Prevention:

Keep Go proxy settings clean and avoid committing local env assumptions.

### 15. Mongo TTL index conflict after changing retention days

Cause:

`SESSION_RAW_EVENT_TTL_DAYS` changed after old TTL index already exists.

Fix:

Drop old TTL index and restart/run migration carefully:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval "db.session_events.getIndexes()"
mongosh "mongodb://localhost:27017/session_db" --eval "db.session_events.dropIndex('ttl_raw_events_90_days')"
```

Prevention:

Plan retention changes as migrations, not random local env changes.

---

## 13. Security & Best Practices

### Security audit from current implementation

| Finding | Risk | Recommended fix |
|---|---|---|
| `.env` is not auto-loaded | Beginners may think config is applied when it is not | Document source command or add a dev runner later |
| `SESSION_IP_HASH_SALT` can be empty | IP hash becomes `unavailable`, reducing fraud/debug value | Make it mandatory in deployment config |
| Redis password blank by default | Fine local, unsafe production | Enable Redis auth/TLS in prod |
| MongoDB local URI has no auth | Fine local, unsafe production | Use Mongo users, password, TLS, private network |
| Trusted identity/role headers accepted directly | Header spoofing risk if service exposed publicly | Put behind API Gateway and strip client-provided auth headers |
| No rate limiting inside service | Public ingest endpoint can be abused | Add API Gateway Redis rate limiting |
| No service Dockerfile/compose committed | Onboarding/deploy inconsistency | Add official Dockerfile and local compose |
| No migration runner | Manual migration mistakes possible | Add Makefile/task script or migration tool |
| Kafka/RabbitMQ publisher disabled | Downstream async analytics not active | Add publisher/outbox when needed |
| Limited observability | Logs exist, metrics/tracing not wired | Add Prometheus metrics and tracing later |

### Best practices for beginners

- Never commit `.env`.
- Keep `.env.example` updated when new env vars are added.
- Use strong random `SESSION_IP_HASH_SALT`.
- Do not store raw IP, passwords, OTP, card numbers, cookies, tokens, or private input values in event properties.
- Use Docker volumes for MongoDB/Redis so data survives container restart.
- Use separate DB names for local, test, staging, and production.
- Use Redis key prefix per service/environment.
- Run `go test ./...` before pushing.
- Run migrations before testing APIs on a fresh database.
- Use `/readyz` to verify MongoDB and Redis, not only `/healthz`.
- In production, expose service through API Gateway, not directly to the internet.
- Use TLS/auth for MongoDB and Redis in production.
- Back up MongoDB if data matters.
- Keep raw event TTL aligned with privacy and analytics requirements.

---

## 14. Missing or Misconfigured Things

These are not blockers for local development, but should be handled before production.

| Missing / misconfigured | Current impact | Suggested action |
|---|---|---|
| No committed Dockerfile for session service | Cannot build official container image | Add Dockerfile under service/deploy path |
| No committed docker-compose for session local stack | Beginners must create/run containers manually | Add local compose with MongoDB, Redis, healthchecks |
| No Makefile/task runner | Commands are manual and easy to forget | Add `make test`, `make run-session`, `make migrate-session` |
| No automatic `.env` loader | Local run can silently use defaults | Use documented source command or add dev-only loader/runner |
| No migration version tracking | JS migrations can be re-run manually without history | Add migration tool or migration collection |
| `SESSION_IP_HASH_SALT` not startup-required | IP hashing can degrade silently | Require non-empty salt in non-local env |
| Auth headers trusted by service | Unsafe if public | Enforce gateway-only access and header sanitization |
| No message queue configured | Future `session.events` not published | Add Kafka/RabbitMQ config when downstream consumers are implemented |
| No metrics endpoint | Harder production monitoring | Add Prometheus metrics endpoint later |
| No Kubernetes manifests | Not deploy-ready to cluster | Add Deployment, Service, ConfigMap, Secret, probes |

---

## 15. Final Checklist

Use this checklist for a fresh machine.

- [ ] Install Git.
- [ ] Install Go 1.26.3 or newer.
- [ ] Clone repository.
- [ ] Run `go version`.
- [ ] Run `go mod download` in `backend/services/session-service`.
- [ ] Install Docker Desktop/Engine or install MongoDB and Redis natively.
- [ ] Start MongoDB on port `27017`.
- [ ] Start Redis on port `6379`.
- [ ] Verify MongoDB with `mongosh`.
- [ ] Verify Redis with `redis-cli ping`.
- [ ] Create `backend/services/session-service/.env` from `.env.example`.
- [ ] Set `SESSION_MONGO_URI`.
- [ ] Set `SESSION_MONGO_DATABASE`.
- [ ] Set `SESSION_REDIS_ADDR`.
- [ ] Set `SESSION_REDIS_DB`.
- [ ] Set a strong `SESSION_IP_HASH_SALT`.
- [ ] Source/load `.env` before running Go.
- [ ] Run MongoDB migrations.
- [ ] Run `go test ./...`.
- [ ] Start backend with `go run ./cmd/server`.
- [ ] Check `curl http://localhost:8086/healthz`.
- [ ] Check `curl http://localhost:8086/readyz`.
- [ ] Send test event to `POST /api/v1/sessions/events`.
- [ ] Verify event in MongoDB.
- [ ] Verify active session key in Redis.
- [ ] Never commit `.env`.

---

## Quick Command Summary

```bash
git clone <repo-url>
cd Ecommerce

docker volume create session_mongo_data
docker volume create session_redis_data

docker run -d --name ecommerce-session-mongo -p 27017:27017 -v session_mongo_data:/data/db mongo:7
docker run -d --name ecommerce-session-redis -p 6379:6379 -v session_redis_data:/data redis:7-alpine redis-server --appendonly yes

cd backend/services/session-service
cp .env.example .env

set -a
source .env
set +a

go mod download
mongosh "$SESSION_MONGO_URI" migrations/001_session_storage_indexes.up.js
mongosh "$SESSION_MONGO_URI" migrations/002_journey_summaries.up.js
go test ./...
go run ./cmd/server
```

In another terminal:

```bash
curl http://localhost:8086/healthz
curl http://localhost:8086/readyz
```

Final Hinglish summary: Pehle Go, MongoDB, Redis ready karo. Fir `.env` banao and source karo. Migrations run karo. Backend start karo. `/readyz` green hai to service MongoDB and Redis dono se properly connected hai.
