# Project Dependency & Setup Guide

## 1. Project Overview

This document is for:

`TaskImplementation/Session Management Service/task5.md`

Output file:

`TaskImplementation/Session Management Service/task5_Dependency.md`

Task 5 ka main focus **Device Tracking** hai. Simple words me: Session Service incoming request/event se browser, OS, device type, client hints, language/timezone, IP hash, device fingerprint hash, and approximate geo context nikalta hai.

Important boundary:

- Business logic yahan rewrite nahi ki gayi.
- Existing implementation file `task5.md` modify nahi kiya gaya.
- Ye file sirf dependency, environment, setup, database, Docker, DevOps, and beginner onboarding guide hai.
- Previous dependency docs me jo setup already explain hai, usko duplicate nahi kiya gaya.

Current repository note:

Task 5 related runtime pieces current codebase me present hain:

| Area | Current file |
|---|---|
| Device enrichment usecase | `backend/services/session-service/internal/usecase/device_enrichment.go` |
| User-agent parser adapter | `backend/services/session-service/internal/device/user_agent_parser.go` |
| MaxMind GeoIP resolver | `backend/services/session-service/internal/device/geo_resolver.go` |
| HMAC privacy hasher | `backend/services/session-service/internal/device/privacy_hasher.go` |
| Request header extraction | `backend/services/session-service/internal/transport/http/handler.go` |
| Device indexes migration | `backend/services/session-service/migrations/003_device_tracking_indexes.up.js` |
| Config and env loading | `backend/services/session-service/internal/config/config.go` |
| Example environment file | `backend/services/session-service/.env.example` |

Task 5 setup ka biggest new dependency change:

| Area | Task 5 change |
|---|---|
| Go packages | New user-agent parser and GeoIP reader dependencies |
| Optional external data file | MaxMind/GeoLite2 `.mmdb` database if GeoIP is enabled |
| Environment | New device tracking, GeoIP, privacy, and trusted proxy variables |
| MongoDB | New indexes for device, browser/OS, country, IP hash, fingerprint hash |
| Redis | Same Redis service, but active-session snapshot now stores device context |
| Ports | No new port |
| Docker | Same MongoDB/Redis containers; optional GeoIP file mount if running service in Docker |

## 2. Tech Stack

Base tech stack already explain kiya gaya hai:

Refer:

`TaskImplementation/Session Management Service/task1_Dependency.md`

Sections:

- `2. Tech Stack`
- `3. Required Software`
- `4. Dependency Management`
- `5. Database Setup`
- `6. Redis / Queue / External Services`

Task 5 detected/reused technologies:

| Technology | What it is | Why Task 5 uses it | Required? | Detailed setup |
|---|---|---|---:|---|
| Go `1.26.3` | Backend programming language | Session Service code run karne ke liye | Yes | `task1_Dependency.md`, section `4. Dependency Management` |
| Go standard `net/http` | Built-in HTTP server/request package | Request headers like `User-Agent`, `X-Forwarded-For`, Client Hints read karne ke liye | Yes | `task1_Dependency.md`, section `2. Tech Stack` |
| Go `crypto/hmac` + `crypto/sha256` | Built-in crypto packages | IP/device/user-agent values ko privacy-safe hash banane ke liye | Yes | Built-in, no install |
| `github.com/mileusna/useragent` | Go user-agent parsing library | Raw browser UA ko browser, OS, mobile/tablet/bot fields me parse karne ke liye | Yes for current code | This file, section `4. Dependency Management` |
| `github.com/oschwald/geoip2-golang` | MaxMind GeoIP2 `.mmdb` reader | IP se approximate country/region/city resolve karne ke liye | Required package, optional runtime feature | This file, section `6. Redis / Queue / External Services` |
| MongoDB | Document database | Enriched `sessions` and `session_events` store karne ke liye | Yes | `task1_Dependency.md`, section `5.1 MongoDB Setup` |
| Redis | In-memory cache | Active session snapshot me device/browser/country quickly update karne ke liye | Yes for current startup | `task1_Dependency.md`, section `5.2 Redis Setup` |
| Docker | Container runtime | MongoDB/Redis local containers run karne ke liye | Optional but recommended | `task1_Dependency.md`, section `8. Docker Setup` |
| mongosh | MongoDB shell | `003_device_tracking_indexes.up.js` migration verify/run karne ke liye | Recommended | `task1_Dependency.md`, section `5.1 MongoDB Setup` |
| redis-cli | Redis CLI | Active session hash verify karne ke liye | Recommended | `task1_Dependency.md`, section `5.2 Redis Setup` |

Beginner Hinglish explanation:

- User-agent parser ek helper library hai. Browser ka long `User-Agent` string directly human-readable nahi hota, so library usko `Chrome`, `Android`, `mobile`, `bot` jaise fields me tod deti hai.
- GeoIP2 reader ek library hai jo MaxMind `.mmdb` file read karke approximate location batata hai. Ye GPS nahi hai. Ye sirf IP based rough country/city signal hai.
- Privacy hasher raw IP ya raw fingerprint store nahi karta. Ye same input ko same hash me convert karta hai, taaki fraud/analytics grouping possible ho but raw value leak na ho.

No new setup introduced for Kafka, RabbitMQ, NATS, MinIO, Elasticsearch, SMTP, Stripe, Twilio, Firebase, S3, or Kubernetes in Task 5.

## 3. Required Software

Base installation steps already documented hain. Repeat nahi kiya gaya.

| Software | Required for Task 5? | Purpose | Where setup is explained |
|---|---:|---|---|
| Git | Yes | Repository clone karne ke liye | `task1_Dependency.md`, section `9. Local Development Setup` |
| Go `1.26.3` or compatible newer version | Yes | Session Service build/run/test | `task1_Dependency.md`, section `3. Required Software` |
| MongoDB | Yes | Durable enriched session/event documents | `task1_Dependency.md`, section `5.1 MongoDB Setup` |
| mongosh | Recommended | Run/verify Task 5 indexes | `task1_Dependency.md`, section `5.1 MongoDB Setup` |
| Redis | Yes | Active session cache and readiness ping | `task1_Dependency.md`, section `5.2 Redis Setup` |
| redis-cli | Recommended | Verify Redis keys | `task1_Dependency.md`, section `5.2 Redis Setup` |
| Docker / Docker Compose | Optional | Local MongoDB/Redis quickly start karne ke liye | `task1_Dependency.md`, section `8. Docker Setup` |
| curl | Recommended | Health and event ingestion test karne ke liye | OS package manager if missing |
| MaxMind/GeoLite2 `.mmdb` file | Optional | GeoIP lookup enable karne ke liye | This file, section `6.1 MaxMind GeoIP` |

Task 5-specific extra requirement:

| Requirement | Why |
|---|---|
| Task 1-4 base setup completed | Device enrichment existing event ingestion and storage flow ke upar run hota hai |
| Migration `001_session_storage_indexes.up.js` applied or startup index creation successful | Base `sessions` and `session_events` collections/indexes chahiye |
| Migration `003_device_tracking_indexes.up.js` applied or startup index creation successful | Device/fraud/analytics indexes chahiye |
| Strong `SESSION_PRIVACY_HASH_PEPPER` set | Raw IP/fingerprint/user-agent hashing ke liye secret pepper chahiye |
| `SESSION_TRUSTED_PROXY_CIDRS` configured in gateway/proxy setup | Forwarded client IP headers safely trust karne ke liye |

## 4. Dependency Management

This is a Go project. Full explanation of `go.mod`, `go.sum`, Go modules, `go mod tidy`, `go mod download`, `go build`, and common Go module issues already exists.

Refer:

`TaskImplementation/Session Management Service/task1_Dependency.md`

Section:

`4. Dependency Management`

Current module:

`backend/services/session-service/go.mod`

Current direct dependencies in `go.mod`:

| Dependency | Version | Task 5 purpose |
|---|---:|---|
| `github.com/mileusna/useragent` | `v1.3.5` | Parse browser, OS, mobile/tablet/desktop/bot from `User-Agent` |
| `github.com/oschwald/geoip2-golang` | `v1.13.0` | Read MaxMind `.mmdb` database for approximate geo |
| `github.com/redis/go-redis/v9` | `v9.19.0` | Existing Redis active-session store |
| `go.mongodb.org/mongo-driver/v2` | `v2.6.0` | Existing MongoDB sessions/events storage |

Current important indirect dependency:

| Dependency | Why it appears |
|---|---|
| `github.com/oschwald/maxminddb-golang` | Used under the hood by `geoip2-golang` to read `.mmdb` files |

Quick commands:

```bash
cd backend/services/session-service
go mod download
go test ./...
go run ./cmd/server
```

If dependencies are changed later:

```bash
cd backend/services/session-service
go mod tidy
```

Beginner note:

`node_modules` jaisa folder Go me normally project ke andar nahi hota. Go dependencies module cache me download hoti hain, and exact versions `go.mod` plus `go.sum` se lock/verify hoti hain.

Task 5-specific dependency issues:

| Error | Common cause | Fix |
|---|---|---|
| `missing go.sum entry for github.com/mileusna/useragent` | Dependency checksums incomplete | `go mod tidy` run karo |
| `missing go.sum entry for github.com/oschwald/geoip2-golang` | GeoIP package checksum missing | `go mod tidy` run karo |
| `go: go.mod requires go >= 1.26.3` | Local Go old hai | Go compatible version install karo |
| `open geoip database: no such file or directory` | `SESSION_GEOIP_ENABLED=true`, but DB file path wrong hai | Correct `.mmdb` path set karo or GeoIP disable karo |

## 5. Database Setup

Detected databases for Task 5:

| Database | Required? | Purpose |
|---|---:|---|
| MongoDB | Yes | Durable enriched `sessions` and `session_events` documents |
| Redis | Yes | Active session snapshot and readiness ping |
| MaxMind `.mmdb` file | Optional data file, not a database server | IP to approximate geo lookup |
| MySQL | No for Session Service | Auth/User services may use SQL, but Task 5 does not |
| PostgreSQL | No | Not used |
| SQLite | No | Not used |

### 5.1 MongoDB

MongoDB installation, Docker run command, persistent volume, start commands, verify commands, default port, connection string, and credentials placement already explained hai.

Refer:

`TaskImplementation/Session Management Service/task1_Dependency.md`

Section:

`5.1 MongoDB Setup`

Task 5 new MongoDB setup:

| Item | Detail |
|---|---|
| New migration | `backend/services/session-service/migrations/003_device_tracking_indexes.up.js` |
| Rollback migration | `backend/services/session-service/migrations/003_device_tracking_indexes.down.js` |
| Collections affected | `sessions`, `session_events` |
| Why needed | Device analytics, browser/OS reports, geo country filters, IP/fingerprint fraud signals |

Run Task 5 migration manually:

```bash
cd backend/services/session-service
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/003_device_tracking_indexes.up.js
```

Verify indexes:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval 'db.sessions.getIndexes().map(i => i.name)'
mongosh "mongodb://localhost:27017/session_db" --eval 'db.session_events.getIndexes().map(i => i.name)'
```

Expected Task 5 index names:

```text
idx_sessions_device_type_started
idx_sessions_browser_os_started
idx_sessions_country_started
idx_sessions_ip_hash_last_seen
idx_sessions_device_fingerprint_last_seen
idx_events_device_event_time
```

Important:

Current service also calls repository `EnsureIndexes` on startup. Manual migration is still useful for controlled onboarding, CI, and production deployment where DB changes should be explicit.

### 5.2 Redis

Redis installation, Docker run command, persistent volume, start commands, verify commands, default port, and credentials placement already explained hai.

Refer:

`TaskImplementation/Session Management Service/task1_Dependency.md`

Section:

`5.2 Redis Setup`

Task 5 Redis-specific note:

Device enrichment does not introduce a new Redis server. Same Redis instance is used. Active-session hash may now include compact fields like:

```text
device_type = mobile
browser = Chrome
os = Android
country = IN
channel = user_app_web
ip_hash = <hashed-value>
```

Verify after ingesting an event:

```bash
redis-cli -n 2 keys 'session:*'
redis-cli -n 2 HGETALL 'session:active:sess_01HX9ZK6N5M2V4B7S8T9Q0R1P2'
```

If your `SESSION_REDIS_KEY_PREFIX` is not `session`, replace `session:*` with your prefix.

## 6. Redis / Queue / External Services

### 6.1 MaxMind GeoIP

MaxMind GeoIP kya hai:

MaxMind GeoIP ek IP-to-location database format provide karta hai. Is project me `geoip2-golang` library local `.mmdb` file read karke rough country, region, city, and timezone nikal sakti hai.

Simple English:

GeoIP is not exact GPS location. Ye approximate location hai based on IP address. Accuracy country level pe usually better hoti hai, city level pe vary kar sakti hai.

Why Task 5 uses it:

| Use | Example |
|---|---|
| Analytics | Mobile sessions from `IN`, `US`, etc. |
| Fraud signal | Same user suddenly different country se login/session |
| Debugging | Country/browser specific checkout issue |

Required or optional:

| Mode | Required? | Behavior |
|---|---:|---|
| `SESSION_GEOIP_ENABLED=false` | No `.mmdb` needed | `geo.source=disabled` or unknown style fallback |
| `SESSION_GEOIP_ENABLED=true` | `.mmdb` path required | Service tries to open `SESSION_GEOIP_DB_PATH` on startup |

Local setup:

1. Create a local data directory:

```bash
mkdir -p backend/services/session-service/data/geoip
```

2. Download a MaxMind GeoLite2 City or paid GeoIP2 City `.mmdb` file from your MaxMind account.

3. Place it here, for local dev example:

```text
backend/services/session-service/data/geoip/GeoLite2-City.mmdb
```

4. Configure `.env`:

```env
SESSION_GEOIP_ENABLED=true
SESSION_GEOIP_DB_PATH=./data/geoip/GeoLite2-City.mmdb
SESSION_GEOIP_TIMEOUT=20ms
```

If you do not have the GeoIP DB yet, keep it disabled:

```env
SESSION_GEOIP_ENABLED=false
SESSION_GEOIP_DB_PATH=
SESSION_GEOIP_TIMEOUT=20ms
```

Docker setup difference:

If the Go service later runs inside Docker, mount the `.mmdb` file into the container and use the container path:

```yaml
services:
  session-service:
    volumes:
      - ./backend/services/session-service/data/geoip:/data/geoip:ro
    environment:
      SESSION_GEOIP_ENABLED: "true"
      SESSION_GEOIP_DB_PATH: /data/geoip/GeoLite2-City.mmdb
```

Do not copy this as a full Compose file. Base Compose setup is already covered in:

`TaskImplementation/Session Management Service/task1_Dependency.md`

Section:

`8. Docker Setup`

### 6.2 API Gateway / Trusted Proxy Headers

Task 3 already documented event ingestion and trusted headers.

Refer:

`TaskImplementation/Session Management Service/task3_Dependency.md`

Sections:

- `6.2 API Gateway / Trusted Headers`
- `7. Environment Variables`

Task 5 adds extra header importance:

| Header | Who should set it | Purpose |
|---|---|---|
| `User-Agent` | Browser/app automatically | Browser/OS/device parse |
| `Accept-Language` | Browser/app automatically | Locale extraction |
| `Sec-CH-UA` | Browser if Client Hints enabled | Browser brand/version hint |
| `Sec-CH-UA-Platform` | Browser if Client Hints enabled | OS/platform hint |
| `Sec-CH-UA-Mobile` | Browser if Client Hints enabled | Mobile hint |
| `Sec-CH-UA-Model` | Browser if Client Hints enabled | Device model hint |
| `X-Forwarded-For` | Trusted gateway/proxy only | Real client IP source |
| `X-Real-IP` | Trusted gateway/proxy only | Real client IP fallback |
| `X-Trusted-Client-IP` | Internal gateway only | Explicit trusted client IP |
| `X-IP-Hash` | Gateway/Auth service if already hashed | Avoid raw IP in downstream flow |
| `X-Device-Fingerprint-Hash` | Frontend/Auth if consented and hashed | Privacy-safe device grouping |
| `X-Device-Fingerprint` | Avoid unless immediately hashed | Raw fingerprint input, do not log/persist |
| `X-Client-Channel` | Frontend/gateway | Channel like `user_app_web` |
| `X-Client-Timezone` | Frontend/gateway | Client timezone |
| `X-Device-Type` | Frontend/gateway optional hint | Device type fallback |

Trusted proxy rule:

Public clients can spoof `X-Forwarded-For`. Service trusts forwarded IP only when request remote IP matches `SESSION_TRUSTED_PROXY_CIDRS`.

### 6.3 Kafka / RabbitMQ / Queue

Task 5 implementation does not introduce Kafka, RabbitMQ, NATS, or any new queue container.

Auth session-link events are mentioned in `task5.md` as integration context, but current local setup for Task 5 does not require running a broker.

## 7. Environment Variables

Base `.env` creation, `.env.example`, and loading behavior already documented hai.

Refer:

`TaskImplementation/Session Management Service/task1_Dependency.md`

Sections:

- `7. Environment Variables`
- `Where .env should be created`
- `Very important: .env is not auto-loaded`

Task 5 variables are already present in:

`backend/services/session-service/.env.example`

Create or update:

```text
backend/services/session-service/.env
```

### 7.1 Task 5 `.env` example only

Use this block in addition to the base env from previous dependency docs:

```env
# Device tracking and privacy-safe enrichment
SESSION_DEVICE_TRACKING_ENABLED=true
SESSION_USER_AGENT_MAX_LENGTH=1024
SESSION_GEOIP_ENABLED=false
SESSION_GEOIP_DB_PATH=
SESSION_GEOIP_TIMEOUT=20ms
SESSION_PRIVACY_HASH_PEPPER=replace-this-with-a-long-random-secret
SESSION_TRUSTED_PROXY_CIDRS=
SESSION_STORE_USER_AGENT=true
```

If you prefer millisecond integer compatibility, current config also supports:

```env
SESSION_GEOIP_TIMEOUT_MS=20
```

Preferred value for current `.env.example` is:

```env
SESSION_GEOIP_TIMEOUT=20ms
```

### 7.2 Task 5 environment variable table

| Variable | Required? | Default/example | Meaning | Security note |
|---|---:|---|---|---|
| `SESSION_DEVICE_TRACKING_ENABLED` | Optional | `true` | Device enrichment on/off switch | Keep enabled in normal envs |
| `SESSION_USER_AGENT_MAX_LENGTH` | Optional | `1024` | Max UA length parsed/stored | Prevent huge header abuse |
| `SESSION_GEOIP_ENABLED` | Optional | `false` | Enable MaxMind DB lookup | If true, DB file path required |
| `SESSION_GEOIP_DB_PATH` | Required only when GeoIP enabled | `./data/geoip/GeoLite2-City.mmdb` | Local path to `.mmdb` file | Do not bake private licensed DB into public image |
| `SESSION_GEOIP_TIMEOUT` | Optional | `20ms` | Max GeoIP lookup time | Keep small, enrichment is best-effort |
| `SESSION_GEOIP_TIMEOUT_MS` | Optional compatibility | `20` | Same timeout as integer ms | Use only if old env style needed |
| `SESSION_PRIVACY_HASH_PEPPER` | Strongly required | Long random secret | HMAC pepper for privacy hashes | Never commit real value |
| `SESSION_TRUSTED_PROXY_CIDRS` | Required behind proxy | `10.0.0.0/8,172.16.0.0/12` | Networks allowed to provide forwarded IP headers | Wrong value can cause spoofed IP trust |
| `SESSION_STORE_USER_AGENT` | Optional | `true` | Store raw trimmed user-agent or `redacted` | Set `false` for stricter privacy |

### 7.3 Credentials placement

Local development:

```text
backend/services/session-service/.env
```

Production:

Use secret manager, Kubernetes Secret, CI/CD secret variables, or platform environment variables.

| Secret/config | Put where locally | Production recommendation |
|---|---|---|
| `SESSION_PRIVACY_HASH_PEPPER` | `.env` | Secret manager |
| `SESSION_IP_HASH_SALT` | Already documented in task1/task3 `.env` | Secret manager |
| `SESSION_MONGO_URI` credentials | Already documented in task1 `.env` | Secret manager / private network |
| `SESSION_REDIS_PASSWORD` | Already documented in task1 `.env` | Secret manager |
| MaxMind license/download credential | Do not put in service `.env` unless automation needs it | CI secret or vendor secret store |

Important:

`SESSION_PRIVACY_HASH_PEPPER` falls back to `SESSION_IP_HASH_SALT` if empty in current config. For clearer security, set a strong explicit `SESSION_PRIVACY_HASH_PEPPER`.

Generate a local random secret:

```bash
openssl rand -hex 32
```

## 8. Docker Setup

No new required Docker container is introduced by Task 5.

Reuse previous Docker setup:

`TaskImplementation/Session Management Service/task1_Dependency.md`

Section:

`8. Docker Setup`

Task 5 Docker-specific notes:

| Scenario | MongoDB value | Redis value | GeoIP value |
|---|---|---|---|
| Go service runs on host | `SESSION_MONGO_URI=mongodb://localhost:27017` | `SESSION_REDIS_ADDR=localhost:6379` | `SESSION_GEOIP_DB_PATH=./data/geoip/GeoLite2-City.mmdb` |
| Go service runs inside Docker Compose | `SESSION_MONGO_URI=mongodb://mongo:27017` | `SESSION_REDIS_ADDR=redis:6379` | `SESSION_GEOIP_DB_PATH=/data/geoip/GeoLite2-City.mmdb` |
| GeoIP disabled | Same as base | Same as base | `SESSION_GEOIP_ENABLED=false` |

Current repo Docker artifact status:

| Artifact | Current status | Impact |
|---|---|---|
| Official session-service Dockerfile | Not committed in current service folder | Full service image build is not standardized yet |
| Official local docker-compose file | Not committed in current service folder | Use previous MongoDB/Redis Docker commands or compose example |
| MongoDB/Redis containers | Documented previously | Enough for Task 5 local dev |
| GeoIP file mount | Not in repo | Needed only if containerized service runs with GeoIP enabled |

Persistent data reminder:

- MongoDB volume matters because enriched `sessions` and `session_events` must survive restart.
- Redis can be persisted locally for debugging, but MongoDB is durable source of truth.
- GeoIP `.mmdb` file should be mounted read-only in containers.

## 9. Local Development Setup

This section is incremental for Task 5. Full clone/install flow is already documented:

`TaskImplementation/Session Management Service/task1_Dependency.md`

Section:

`9. Local Development Setup`

Task 5 onboarding flow:

1. Clone repo and enter project.
2. Install/verify Go.
3. Start MongoDB.
4. Start Redis.
5. Create/load base `.env` from previous docs.
6. Add Task 5 env variables from section `7. Environment Variables`.
7. Keep `SESSION_GEOIP_ENABLED=false` unless you have a valid `.mmdb` file.
8. Run base migrations first if not already done.
9. Run Task 5 migration `003_device_tracking_indexes.up.js`.
10. Run tests.
11. Start backend.
12. Send a test event with a realistic User-Agent and headers.
13. Verify MongoDB and Redis contain device context.

Commands:

```bash
cd backend/services/session-service
go mod download
```

Load env in bash:

```bash
set -a
. ./.env
set +a
```

Run migrations in order:

```bash
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/001_session_storage_indexes.up.js
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/002_journey_summaries.up.js
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/003_device_tracking_indexes.up.js
```

Run tests:

```bash
go test ./...
```

Start service:

```bash
go run ./cmd/server
```

Verify health:

```bash
curl -i http://localhost:8086/healthz
curl -i http://localhost:8086/readyz
```

## 10. Running the Project

Daily Task 5 run flow:

```bash
cd backend/services/session-service
set -a
. ./.env
set +a
go run ./cmd/server
```

### 10.1 Send a device-enriched test event

```bash
curl -i \
  -X POST "http://localhost:8086/api/v1/sessions/events" \
  -H "Content-Type: application/json" \
  -H "User-Agent: Mozilla/5.0 (Linux; Android 14; Pixel 7) AppleWebKit/537.36 Chrome/125.0 Mobile Safari/537.36" \
  -H "Accept-Language: en-US,en;q=0.9" \
  -H "X-Client-Channel: user_app_web" \
  -H "X-Client-Timezone: Asia/Kolkata" \
  -H "X-Device-Fingerprint-Hash: dev_hashed_fingerprint_example" \
  --data '{
    "event_type": "product_view",
    "anonymous_id": "anon_device_test",
    "session_id": "sess_device_test",
    "path": "/products/prod_123",
    "occurred_at": "2026-05-22T10:00:00Z",
    "properties": {
      "product_id": "prod_123",
      "viewport_width": 390,
      "viewport_height": 844,
      "screen_width": 1080,
      "screen_height": 2400,
      "timezone": "Asia/Kolkata"
    }
  }'
```

Expected response:

```text
HTTP/1.1 202 Accepted
```

### 10.2 Verify MongoDB enriched fields

```bash
mongosh "mongodb://localhost:27017/session_db" --eval 'db.session_events.findOne({session_id:"sess_device_test"}, {device:1, client:1, geo:1, ip_hash:1, user_agent_hash:1, device_fingerprint_hash:1})'
mongosh "mongodb://localhost:27017/session_db" --eval 'db.sessions.findOne({session_id:"sess_device_test"}, {device:1, client:1, geo:1, ip_hash:1, user_agent_hash:1, device_fingerprint_hash:1})'
```

Expected idea:

```text
device.type should be mobile
device.browser should be Chrome or parsed equivalent
device.os should be Android
client.locale should be en-US
geo.source should be disabled, unknown, header, or geoip depending on config
raw IP should not be stored
```

### 10.3 Verify Redis active snapshot

```bash
redis-cli -n 2 HGETALL 'session:active:sess_device_test'
redis-cli -n 2 TTL 'session:active:sess_device_test'
```

### 10.4 Ports & Networking

| Service | Port | Purpose | New in Task 5? |
|---|---:|---|---:|
| Session Service HTTP | `8086` | API, health, readiness | No |
| MongoDB | `27017` | Durable sessions/events storage | No |
| Redis | `6379` | Active session cache | No |
| MaxMind `.mmdb` | No port | Local file read, not a server | Yes as optional file |
| Kafka/RabbitMQ | N/A | Not required for Task 5 local setup | No |

Networking notes:

- Host-run service uses `localhost` for MongoDB/Redis.
- Container-run service should use Compose service names like `mongo` and `redis`.
- GeoIP DB path must be valid from the process point of view. Host path and container path are different.
- `X-Forwarded-For` is trusted only when remote address belongs to `SESSION_TRUSTED_PROXY_CIDRS`.

## 11. Common Errors & Fixes

General Go/MongoDB/Redis/event ingestion errors already documented:

Refer:

`TaskImplementation/Session Management Service/task1_Dependency.md`

Section:

`12. Common Errors & Fixes`

Refer also:

`TaskImplementation/Session Management Service/task3_Dependency.md`

Section:

`11. Common Errors & Fixes`

Task 5-specific troubleshooting:

### 1. `SESSION_GEOIP_DB_PATH is required when SESSION_GEOIP_ENABLED=true`

Cause:

GeoIP enabled hai but path blank hai.

Fix:

```env
SESSION_GEOIP_ENABLED=false
SESSION_GEOIP_DB_PATH=
```

Or provide a real file:

```env
SESSION_GEOIP_ENABLED=true
SESSION_GEOIP_DB_PATH=./data/geoip/GeoLite2-City.mmdb
```

### 2. `open geoip database: no such file or directory`

Cause:

Path wrong hai, file download nahi hui, ya Docker container ke andar path exist nahi karta.

Fix:

```bash
ls -la backend/services/session-service/data/geoip
```

If running inside Docker, mount the folder and use container path:

```env
SESSION_GEOIP_DB_PATH=/data/geoip/GeoLite2-City.mmdb
```

### 3. `SESSION_GEOIP_TIMEOUT must be greater than zero`

Cause:

Timeout `0`, negative, ya invalid value set hai.

Fix:

```env
SESSION_GEOIP_TIMEOUT=20ms
```

### 4. `SESSION_TRUSTED_PROXY_CIDRS contains invalid CIDR`

Cause:

CIDR format wrong hai.

Fix:

```env
SESSION_TRUSTED_PROXY_CIDRS=10.0.0.0/8,172.16.0.0/12,192.168.0.0/16
```

Bad examples:

```text
10.0.0.1
localhost
*
```

### 5. Device always shows `unknown`

Possible causes:

- `User-Agent` header missing.
- `SESSION_DEVICE_TRACKING_ENABLED=false`.
- Parser cannot identify unusual UA.
- Frontend/gateway strips headers.

Fix:

```bash
curl -i http://localhost:8086/healthz \
  -H "User-Agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/125.0 Safari/537.36"
```

For event ingestion, include a real browser UA in the event request.

### 6. Geo always shows `disabled`

Cause:

`SESSION_GEOIP_ENABLED=false`, or device tracking disabled.

Fix:

Enable only when `.mmdb` file is ready:

```env
SESSION_DEVICE_TRACKING_ENABLED=true
SESSION_GEOIP_ENABLED=true
SESSION_GEOIP_DB_PATH=./data/geoip/GeoLite2-City.mmdb
```

### 7. Geo always shows `unknown`

Common causes:

- Local/private IP like `127.0.0.1`, `10.x.x.x`, `192.168.x.x`.
- GeoIP file has no city match for that IP.
- Geo lookup timed out.
- Forwarded IP not trusted because `SESSION_TRUSTED_PROXY_CIDRS` is blank/wrong.

Fix:

For local dev, `unknown` is normal because requests come from localhost/private IP.

Behind gateway, configure:

```env
SESSION_TRUSTED_PROXY_CIDRS=<gateway-private-cidr>
```

### 8. Raw IP appears in `ip_hash`

Cause:

Gateway sent bad `X-IP-Hash` value or raw IP through hash header.

Current code ignores `X-IP-Hash` if it parses as an IP, then hashes client IP. Still, gateway should be fixed.

Fix:

- Gateway should send raw IP only in trusted IP headers.
- Gateway should send hashed value only in `X-IP-Hash`.
- Never send raw IP as `X-IP-Hash`.

### 9. User-agent stored as `redacted`

Cause:

`SESSION_STORE_USER_AGENT=false`.

Fix:

This may be intentional for privacy. Parser still uses UA for enrichment, but stored raw UA becomes `redacted`.

### 10. Device tracking indexes missing

Cause:

Migration not run and service startup did not create indexes due to Mongo issue.

Fix:

```bash
cd backend/services/session-service
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/003_device_tracking_indexes.up.js
```

## 12. Security & Best Practices

### 12.1 Security audit for Task 5

| Finding | Risk | Recommendation |
|---|---|---|
| `.env` can contain real secrets | Secret leak if committed | Keep `.env` untracked; use `.env.example` only for placeholders |
| `SESSION_PRIVACY_HASH_PEPPER` has example placeholder | Weak hash privacy if copied to prod | Generate strong random per environment |
| GeoIP DB file may be licensed/vendor controlled | License or data leak risk | Store privately; mount read-only; do not publish in public image |
| `SESSION_STORE_USER_AGENT=true` stores raw UA | UA can be privacy-sensitive | Consider `false` for stricter privacy environments |
| `X-Forwarded-For` can be spoofed | Wrong IP/geo/fraud signal | Configure `SESSION_TRUSTED_PROXY_CIDRS` and only trust gateway networks |
| Raw `X-Device-Fingerprint` header exists | High-risk identifier | Prefer `X-Device-Fingerprint-Hash`; never log raw fingerprint |
| GeoIP enabled requires local file | Startup failure if missing | Keep disabled by default locally unless file exists |
| No committed service Dockerfile/Compose | Deployment not standardized | Add official Docker artifacts before production rollout |
| MaxMind DB update process not automated | Stale geo data | Add scheduled download/refresh process outside app code |

### 12.2 Task 5 best practices

Avoid repeating base best practices from previous docs. For base setup, refer:

`TaskImplementation/Session Management Service/task1_Dependency.md`

Section:

`13. Security & Best Practices`

Task 5-specific practices:

| Practice | Why |
|---|---|
| Keep enrichment best-effort | Parser/geo failure should not reject valid events |
| Store hashes, not raw IP/fingerprint | Reduces privacy and breach impact |
| Keep GeoIP approximate | Country/region/city is enough for analytics; avoid exact lat/long |
| Treat `ip_hash` as sensitive-ish | It can still correlate behavior |
| Mask hashes in logs/admin UI | Debugging does not need full identifiers |
| Keep `SESSION_GEOIP_TIMEOUT` small | Device enrichment should not slow event ingestion |
| Use trusted proxy CIDRs | Prevent spoofed forwarded IP |
| Use fixed device enums | `desktop`, `mobile`, `tablet`, `bot`, `unknown` makes analytics cleaner |
| Keep parser behind adapter | Future library replacement easy rahega |

## 13. Missing or Misconfigured Things

| Item | Current status | Impact | Suggested fix |
|---|---|---|---|
| Official session-service Dockerfile | Missing in service folder | Beginners cannot build service container consistently | Add Dockerfile before deployment |
| Official docker-compose for full Session Service | Missing in service folder | Local app + Mongo + Redis orchestration manual hai | Add compose with Mongo, Redis, service, optional GeoIP mount |
| GeoIP DB provisioning automation | Missing | Manual `.mmdb` download/update | Add secure CI/job or ops runbook |
| GeoIP DB path default | Blank, disabled by default | Safe local default, but GeoIP not active | Enable only in envs with file mounted |
| `SESSION_PRIVACY_HASH_PEPPER` placeholder in `.env.example` | Example only | Risk if copied to production | Generate real secret in each env |
| `SESSION_TRUSTED_PROXY_CIDRS` blank by default | Forwarded IP not trusted locally | Geo/IP from real client may not work behind gateway | Set gateway CIDRs in deployed env |
| `.env` auto-load | Not implemented | Beginners may create `.env` but app reads defaults | Source env before running or add documented runner |
| MaxMind license/download docs | Not in repo | New developers may not know where `.mmdb` comes from | Add internal vendor access instructions |

## 14. References to Previous Dependency Files

Do not duplicate these sections. Follow the existing docs:

| Topic | Previous file | Section |
|---|---|---|
| Base project overview and full setup | `TaskImplementation/Session Management Service/task1_Dependency.md` | `1. Project Overview` |
| Go, net/http, MongoDB, Redis beginner explanations | `task1_Dependency.md` | `2. Tech Stack` |
| Required software install | `task1_Dependency.md` | `3. Required Software` |
| Go modules, `go.mod`, `go.sum`, `go mod tidy`, `go mod download` | `task1_Dependency.md` | `4. Dependency Management` |
| MongoDB install, Docker run, verify, credentials | `task1_Dependency.md` | `5.1 MongoDB Setup` |
| Redis install, Docker run, verify, credentials | `task1_Dependency.md` | `5.2 Redis Setup` |
| Full `.env` creation/loading | `task1_Dependency.md` | `7. Environment Variables` |
| Docker Compose and local containers | `task1_Dependency.md` | `8. Docker Setup` |
| Full local development setup | `task1_Dependency.md` | `9. Local Development Setup` |
| Running service basics | `task1_Dependency.md` | `10. Running the Project` |
| Ports and networking basics | `task1_Dependency.md` | `11. Ports & Networking` |
| General troubleshooting | `task1_Dependency.md` | `12. Common Errors & Fixes` |
| Storage architecture and MongoDB/Redis collection/key design | `task2_Dependency.md` | `5. Database Setup`, `6. Redis / Queue / External Services` |
| Event ingestion setup and trusted headers | `task3_Dependency.md` | `6.2 API Gateway / Trusted Headers`, `9. Local Development Setup` |
| Event ingestion errors | `task3_Dependency.md` | `11. Common Errors & Fixes` |
| Journey migration and admin route context | `task4_Dependency.md` | `5. Database Setup`, `9. Local Development Setup` |

Task 5 new material in this file:

- User-agent parser dependency and setup notes
- MaxMind GeoIP `.mmdb` file setup
- Device tracking env variables
- Trusted proxy CIDR impact for device geo/IP context
- MongoDB device indexes migration
- Device-specific Redis verification
- Device tracking troubleshooting
- Device privacy and security audit

## 15. Final Checklist

Use this checklist before saying Task 5 dependency/setup is ready:

| Check | Done |
|---|---|
| Go installed and `go version` is compatible with `go.mod` |  |
| `go mod download` succeeds in `backend/services/session-service` |  |
| `github.com/mileusna/useragent` exists in `go.mod` |  |
| `github.com/oschwald/geoip2-golang` exists in `go.mod` |  |
| MongoDB is running on `27017` or configured URI |  |
| Redis is running on `6379` or configured address |  |
| `.env` is created or env vars are exported |  |
| Base MongoDB variables are set |  |
| Base Redis variables are set |  |
| `SESSION_DEVICE_TRACKING_ENABLED=true` or intentionally disabled |  |
| `SESSION_USER_AGENT_MAX_LENGTH > 0` |  |
| `SESSION_PRIVACY_HASH_PEPPER` is a strong non-placeholder secret |  |
| `SESSION_STORE_USER_AGENT` policy is intentionally chosen |  |
| `SESSION_GEOIP_ENABLED=false` if no `.mmdb` file exists |  |
| If GeoIP enabled, `SESSION_GEOIP_DB_PATH` points to a real `.mmdb` file |  |
| `SESSION_GEOIP_TIMEOUT` is greater than zero, example `20ms` |  |
| `SESSION_TRUSTED_PROXY_CIDRS` is set correctly behind gateway/proxy |  |
| Migration `001_session_storage_indexes.up.js` applied or startup index creation works |  |
| Migration `003_device_tracking_indexes.up.js` applied or startup index creation works |  |
| Device indexes are visible in MongoDB |  |
| `go test ./...` passes |  |
| `go run ./cmd/server` starts successfully |  |
| `/healthz` returns success |  |
| `/readyz` confirms MongoDB and Redis are reachable |  |
| Test event returns `202 Accepted` |  |
| MongoDB event/session contains `device`, `client`, `geo`, `ip_hash`, and `user_agent_hash` fields |  |
| Redis active session hash contains compact device fields |  |
| No raw IP or raw device fingerprint is stored/logged |  |

