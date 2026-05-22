# Project Dependency & Setup Guide

This guide is for:

```text
TaskImplementation/Session Management Service/task3.md
```

Output file:

```text
TaskImplementation/Session Management Service/task3_Dependency.md
```

Goal simple hai: Task 3 me **Event Ingestion API** ke liye jo runtime/setup/devops requirements hain, unko beginner-friendly way me explain karna. Business logic yahan rewrite nahi ki gayi. Ye guide sirf dependencies, environment variables, database/cache setup, Docker notes, run steps, and debugging ke liye hai.

Important reuse rule: same service folder me pehle se full setup docs available hain:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
TaskImplementation/Session Management Service/task2_Dependency.md
```

Isliye Go install, MongoDB install, Redis install, Docker basics, complete `.env`, migrations, and common setup troubleshooting yahan duplicate nahi kiya gaya. Jahan setup same hai, exact reference diya gaya hai.

---

## 1. Project Overview

Task 3 ka focus hai:

```http
POST /api/v1/sessions/events
```

Ye endpoint frontend analytics SDK se page views, clicks, scrolls, product/cart/checkout events receive karta hai. Service event validate karti hai, sensitive properties clean karti hai, MongoDB me raw event save karti hai, existing session summary touch karti hai, Redis me active-session TTL refresh karti hai, and client ko `202 Accepted` return karti hai.

Current implemented service path:

```text
backend/services/session-service
```

Task 3 implementation currently uses:

| Area | Current status |
|---|---|
| Runtime | Go service |
| HTTP framework | Go standard `net/http` |
| Endpoint | `POST /api/v1/sessions/events` |
| Durable storage | MongoDB collection `session_events`, plus `sessions` touch |
| Cache/hot state | Redis active-session keys |
| Queue/event bus | Optional interface exists; actual publisher is `nil` right now |
| API Gateway | Expected in production for trusted headers, rate limits, CORS |
| Local Docker need | Same MongoDB + Redis setup as previous tasks |

Simple Hinglish:

Event ingestion API ek public tracking endpoint hai. Frontend SDK event bhejta hai, backend usko safe banake MongoDB me store karta hai, Redis me active session refresh karta hai, aur fast response deta hai. Kafka/RabbitMQ ka final setup abhi required nahi hai.

---

## 2. Tech Stack

### Task 3 Detected Technologies

| Technology | What it is | Why used in Task 3 | Required? | Setup reference |
|---|---|---|---:|---|
| Go | Compiled backend language | Session service backend run/test/build ke liye | Yes | `task1_Dependency.md`, section `2. Tech Stack` |
| Go `net/http` | Built-in HTTP server package | `POST /api/v1/sessions/events`, `/healthz`, `/readyz` routes ke liye | Yes | `task1_Dependency.md`, section `2. Tech Stack` |
| Go `encoding/json` | Built-in JSON package | Event request decode and JSON response encode ke liye | Yes, built-in | No external install needed |
| MongoDB | Document database | Raw events and session summary store karne ke liye | Yes | `task1_Dependency.md`, section `5.1 MongoDB Setup`; `task2_Dependency.md`, section `5.1 MongoDB Setup` |
| Redis | In-memory data store | Active session TTL and live state refresh ke liye | Yes | `task1_Dependency.md`, section `5.2 Redis Setup`; `task2_Dependency.md`, section `6.1 Redis` |
| MongoDB Go Driver | Go library | `session_events` insert and index creation ke liye | Yes | `task1_Dependency.md`, section `4. Dependency Management` |
| go-redis/v9 | Go library | Redis ping, hash/set/TTL operations ke liye | Yes | `task1_Dependency.md`, section `4. Dependency Management` |
| API Gateway | Front-door service/proxy | Trusted auth headers, rate limit, CORS, request IDs ke liye | Required in production, optional locally | Task-specific notes below |
| Kafka/RabbitMQ | Message queue/event stream | Future `session.events` async publish ke liye | Not required now | `task1_Dependency.md`, section `6. Redis / Queue / External Services` |

### Beginner Explanation

Go ek fast backend language hai. Is project me service Go me likhi gayi hai because microservices lightweight and deploy-friendly bante hain.

`net/http` Go ka built-in web server package hai. Iska matlab Gin/Fiber/Echo install karne ki zarurat nahi hai.

MongoDB ek JSON-like document database hai. Event payload flexible hota hai, isliye `properties` jaise dynamic object ke liye MongoDB useful hai.

Redis ek fast in-memory cache hai. Har event par active session ka `last_seen_at` and TTL update hota hai, so Redis fast live state ke liye perfect hai.

API Gateway production me important hai because endpoint public hai. Gateway spoofed headers strip karega, rate limit lagayega, CORS control karega, and logged-in user identity trusted headers me pass karega.

---

## 3. Required Software

Base software same hai as previous dependency docs.

| Software | Required? | Why needed | Do not duplicate; refer |
|---|---:|---|---|
| Git | Yes | Repository clone karne ke liye | `task1_Dependency.md`, section `3. Required Software` |
| Go 1.26.3 or newer | Yes | Service run/build/test ke liye | `task1_Dependency.md`, section `3. Required Software` |
| MongoDB | Yes | Event durable storage | `task1_Dependency.md`, section `5.1 MongoDB Setup` |
| mongosh | Recommended | Migrations and event verification | `task1_Dependency.md`, section `5.1 MongoDB Setup` |
| Redis | Yes | Active session cache | `task1_Dependency.md`, section `5.2 Redis Setup` |
| redis-cli | Recommended | Redis key/TTL debug | `task1_Dependency.md`, section `5.2 Redis Setup` |
| Docker | Optional but recommended | MongoDB/Redis local containers | `task1_Dependency.md`, section `8. Docker Setup` |
| curl/Postman | Recommended | Event API testing | `task1_Dependency.md`, section `9. Local Development Setup` |

Task 3 me koi new mandatory OS-level software introduce nahi hua.

---

## 4. Dependency Management

This is a Go project. Go modules, `go.mod`, `go.sum`, `go mod download`, `go mod tidy`, `go build`, and common Go module failures already explain kiye gaye hain:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
4. Dependency Management
```

Task 3 ke liye current direct dependencies same hain:

| Dependency | Version | Task 3 purpose |
|---|---:|---|
| `go.mongodb.org/mongo-driver/v2` | `v2.6.0` | `session_events` insert, MongoDB indexes, collection reads |
| `github.com/redis/go-redis/v9` | `v9.19.0` | Redis active session touch and readiness ping |

No new external Go module is required for Task 3. Event ingestion uses Go standard packages like `net/http`, `encoding/json`, `mime`, `time`, `crypto/hmac`, and `crypto/sha256`.

Useful commands:

```bash
cd backend/services/session-service
go mod download
go mod tidy
go test ./...
go run ./cmd/server
```

Hinglish note:

`go.mod` dependency list hai. `go.sum` dependency checksum/security lock jaisa hai. `go mod tidy` missing dependency add karta hai and unused dependency remove karta hai.

---

## 5. Database Setup

Task 3 uses the same databases as previous tasks:

| Database | Required? | Task 3 role | Setup reference |
|---|---:|---|---|
| MongoDB | Yes | Store raw events in `session_events`; update/touch `sessions` | `task1_Dependency.md`, section `5.1 MongoDB Setup`; `task2_Dependency.md`, section `5.1 MongoDB Setup` |
| Redis | Yes | Refresh active session hash, sets, counters, TTL | `task1_Dependency.md`, section `5.2 Redis Setup`; `task2_Dependency.md`, section `6.1 Redis` |
| MySQL | No | Not used by Session Management Service Task 3 | Not applicable |
| PostgreSQL | No | Not detected | Not applicable |
| SQLite | No | Not detected | Not applicable |

### 5.1 MongoDB For Task 3

#### A. What It Is

MongoDB ek NoSQL document database hai. Ye JSON-like documents store karta hai. Event data flexible hota hai, so MongoDB is a good fit.

#### B. Why Task 3 Uses It

Task 3 me MongoDB durable source of truth hai:

| Collection | Task 3 usage |
|---|---|
| `session_events` | Har accepted event ka raw/sanitized document |
| `sessions` | `last_seen_at`, `exit_page`, status, user/session summary touch |

#### C. Required Or Optional

MongoDB required hai. Event insert fail hua to API `503 INGEST_STORAGE_UNAVAILABLE` return kar sakti hai because durable write unavailable hai.

#### D. Local Installation

Do not repeat. Follow:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
5.1 MongoDB Setup
```

#### E. Docker Setup

Same as previous tasks. Follow:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Sections:
- 5.1 MongoDB Setup
- 8. Docker Setup
- 9. Local Development Setup, Step 4
```

#### F. Start Commands

Quick reminder only:

```bash
docker start ecommerce-session-mongo
```

#### G. Verify Running

```bash
mongosh "mongodb://localhost:27017" --eval "db.runCommand({ ping: 1 })"
```

Expected:

```text
ok: 1
```

#### H. Default Port

| Setting | Value |
|---|---|
| MongoDB port | `27017` |
| Local URI | `mongodb://localhost:27017` |
| Default DB | `session_db` |

#### I. Connection String Format

Already documented in:

```text
TaskImplementation/Session Management Service/task2_Dependency.md
Section:
5.1 MongoDB Setup -> I. Connection String Format
```

Task 3 uses same values:

```env
SESSION_MONGO_URI=mongodb://localhost:27017
SESSION_MONGO_DATABASE=session_db
```

#### J. Where To Place Credentials

Local:

```text
backend/services/session-service/.env
```

Production:

Use secret manager, Kubernetes Secret, CI/CD secret variables, or managed platform environment variables. Real MongoDB password Git me commit mat karo.

### 5.2 MongoDB Indexes Needed By Task 3

Task 3 depends on indexes already explained in:

```text
TaskImplementation/Session Management Service/task2_Dependency.md
Section:
5.3 Required MongoDB Indexes
```

Task 3-specific important indexes:

| Index | Collection | Why it matters for Event Ingestion |
|---|---|---|
| `idx_session_timeline` | `session_events` | Session journey/timeline reads fast bante hain |
| `idx_event_type_time` | `session_events` | Event type wise analytics queries ke liye |
| `idx_session_events_request_id` | `session_events` | Request tracing/debugging ke liye |
| `ttl_raw_events_90_days` or configured TTL name | `session_events` | Old raw events cleanup |
| `uniq_session_id` | `sessions` | Same session duplicate summary avoid hota hai |

Verify:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval "db.session_events.getIndexes()"
mongosh "mongodb://localhost:27017/session_db" --eval "db.sessions.getIndexes()"
```

### 5.3 Migrations

Migration process same hai. Follow:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
9. Local Development Setup, Step 8: Run MongoDB migrations

TaskImplementation/Session Management Service/task2_Dependency.md
Section:
5.4 Migrations
```

Current service also calls `EnsureIndexes` on startup, but migrations are still recommended because migration files include collection validators.

---

## 6. Redis / Queue / External Services

### 6.1 Redis

Redis setup already documented:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
5.2 Redis Setup

TaskImplementation/Session Management Service/task2_Dependency.md
Section:
6.1 Redis
```

Task 3 Redis behavior:

| Redis data | Purpose |
|---|---|
| `session:active:{session_id}` | Active session snapshot hash |
| `session:last_seen:{session_id}` | Last seen timestamp |
| `session:anon:{anonymous_id}` | Anonymous visitor active session set |
| `session:user:{user_id}` | Logged-in user active session set |
| `session:counter:active:{yyyyMMddHH}` | Active counter for live metrics |

Important: Redis touch failure does not reject event after MongoDB write succeeds. Service logs warning and still returns accepted. MongoDB durable write is the important dependency.

Verify after sending event:

```bash
redis-cli -n 2 HGETALL session:active:sess_local_001
redis-cli -n 2 TTL session:active:sess_local_001
```

### 6.2 API Gateway / Trusted Headers

Task 3 endpoint is public-style ingestion. Local testing direct service par ho sakti hai, but production me API Gateway strongly required hai.

API Gateway responsibilities:

| Responsibility | Why important |
|---|---|
| Rate limiting | Broken SDK/bots se event spam avoid karna |
| CORS allowlist | Sirf trusted frontend origins allow karna |
| Strip spoofed identity headers | Public client fake `X-User-ID` na bhej sake |
| Inject trusted user ID | Authenticated request me real user identity pass karna |
| Add request/trace ID | Debugging and logs correlate karna |
| Forward safe client metadata | IP hash, channel, device type headers pass karna |

Headers consumed by current service:

| Header | Purpose |
|---|---|
| `X-User-ID`, `X-Authenticated-User-ID`, `X-Auth-User-ID` | Trusted logged-in user ID |
| `X-Request-ID`, `X-Trace-ID`, `Traceparent` | Request ID/trace propagation |
| `X-IP-Hash` | Precomputed privacy-safe IP hash |
| `X-IP-Version` | `ipv4`, `ipv6`, or `unknown` |
| `X-Forwarded-For`, `X-Real-IP` | Client IP source if gateway does not send `X-IP-Hash` |
| `X-Client-Channel` | Client channel, for example `user_app_web` |
| `X-Device-Type` | Device type, for example `desktop` or `mobile` |

Warning:

Body ka `user_id` current handler trust nahi karta. Logged-in user ID trusted gateway/auth header se aana chahiye. Ye intentional security behavior hai.

### 6.3 Kafka / RabbitMQ / `session.events`

Task 3 implementation guide me optional `session.events` publish mention hai, but current actual service me publisher `nil` pass hota hai.

| Service | Required now? | Current status |
|---|---:|---|
| Kafka | No | No Kafka client configured |
| RabbitMQ | No | No RabbitMQ client configured |
| NATS | No | Not detected |
| MinIO/S3 | No | Not detected |
| SMTP/Twilio/Stripe/Firebase | No | Not detected |

Future env idea from task guide:

```env
SESSION_EVENTS_TOPIC=session.events
```

Do not add this to required local `.env` yet unless publisher implementation is added later.

---

## 7. Environment Variables

Full `.env` already documented here:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
7. Environment Variables
```

Storage variables already documented here:

```text
TaskImplementation/Session Management Service/task2_Dependency.md
Section:
7. Environment Variables
```

Task 3 only needs these event-ingestion-specific variables highlighted.

### 7.1 Task 3 `.env` Example

Add/verify these in:

```text
backend/services/session-service/.env
```

```env
# Event ingestion API limits
SESSION_INGEST_MAX_BODY_BYTES=65536
SESSION_ALLOWED_CLOCK_SKEW=5m
SESSION_MAX_EVENT_AGE=24h

# Privacy-safe IP hashing
SESSION_IP_HASH_SALT=replace-this-with-a-long-random-secret

# Event properties validation
SESSION_MAX_EVENT_TYPE_LENGTH=64
SESSION_MAX_EVENT_PROPERTY_KEY_LENGTH=128
SESSION_MAX_EVENT_PROPERTY_STRING_LENGTH=2048
SESSION_MAX_EVENT_PROPERTIES=50
SESSION_MAX_EVENT_PROPERTY_DEPTH=6
```

Compatibility variables also supported:

```env
SESSION_ALLOWED_CLOCK_SKEW_SECONDS=300
SESSION_MAX_EVENT_AGE_HOURS=24
```

Recommended: use duration-style variables (`5m`, `24h`) because they are easier to read.

### 7.2 Task 3 Environment Variable Table

| Variable | Required? | Default | Purpose | Beginner note |
|---|---:|---:|---|---|
| `SESSION_INGEST_MAX_BODY_BYTES` | Optional | `65536` | Max JSON body size per event | 64 KiB se huge payload abuse stop hota hai |
| `SESSION_ALLOWED_CLOCK_SKEW` | Optional | `5m` | Client timestamp future tolerance | Client clock thoda ahead ho sakta hai |
| `SESSION_ALLOWED_CLOCK_SKEW_SECONDS` | Optional | `300` | Older compatibility format | Use only if duration variable nahi use kar rahe |
| `SESSION_MAX_EVENT_AGE` | Optional | `24h` | Old event accept window | Bahut purane retry events reject honge |
| `SESSION_MAX_EVENT_AGE_HOURS` | Optional | `24` | Older compatibility format | Use only if duration variable nahi use kar rahe |
| `SESSION_IP_HASH_SALT` | Strongly required | empty | HMAC salt for IP hash | Secret jaisa treat karo |
| `SESSION_MAX_EVENT_TYPE_LENGTH` | Optional | `64` | Max `event_type` length | Bad/huge event names prevent karta hai |
| `SESSION_MAX_EVENT_PROPERTY_KEY_LENGTH` | Optional | `128` | Max property key length | Payload shape controlled rahe |
| `SESSION_MAX_EVENT_PROPERTY_STRING_LENGTH` | Optional | `2048` | Max string value length | Huge text fields avoid hote hain |
| `SESSION_MAX_EVENT_PROPERTIES` | Optional | `50` | Max properties count | Event payload compact rakhta hai |
| `SESSION_MAX_EVENT_PROPERTY_DEPTH` | Optional | `6` | Max nested object/array depth | Deep nested JSON abuse avoid hota hai |

### 7.3 `.env` Loading Reminder

Current Go code uses `os.Getenv`. `.env` automatic load nahi hota.

Refer:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
7. Environment Variables -> Very important: `.env` is not auto-loaded
```

Linux/macOS quick reminder:

```bash
cd backend/services/session-service
set -a
source .env
set +a
go run ./cmd/server
```

### 7.4 Credentials Placement

| Config/secret | Local place | Production place |
|---|---|---|
| `SESSION_IP_HASH_SALT` | `backend/services/session-service/.env` | Secret manager / Kubernetes Secret |
| MongoDB credentials | `backend/services/session-service/.env` | Secret manager / Kubernetes Secret |
| Redis password | `backend/services/session-service/.env` | Secret manager / Kubernetes Secret |
| API Gateway trusted header config | Gateway config, not service `.env` | Gateway/IaC config |

Warning:

`SESSION_IP_HASH_SALT` real secret hai. Agar ye leak ho gaya to IP hashes less safe ho sakte hain. Git me commit mat karo.

---

## 8. Docker Setup

Docker setup same as previous files. Do not duplicate full Docker commands here.

Refer:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Sections:
- 8. Docker Setup
- 9. Local Development Setup, Step 4

TaskImplementation/Session Management Service/task2_Dependency.md
Section:
8. Docker Setup
```

Task 3-specific Docker notes:

| Area | Local Go run on host | Go service inside Docker Compose |
|---|---|---|
| MongoDB URI | `mongodb://localhost:27017` | `mongodb://mongo:27017` |
| Redis address | `localhost:6379` | `redis:6379` |
| Event endpoint URL | `http://localhost:8086/api/v1/sessions/events` | exposed service port, usually `8086` |
| API Gateway URL | optional locally | gateway service name/port if Compose includes gateway |

Current repo status:

| Artifact | Status |
|---|---|
| Session service Dockerfile | Not committed |
| Session local docker-compose file | Not committed |
| MongoDB/Redis Docker commands | Documented in previous dependency guides |

Beginner warning:

Container ke andar `localhost` ka matlab same container hota hai. Agar Go service bhi Docker me hai, Mongo/Redis service names use karo.

---

## 9. Local Development Setup

Full onboarding flow already exists:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
9. Local Development Setup
```

Task 3 incremental flow:

1. Follow `task1_Dependency.md` sections `3`, `4`, `5`, `7`, `8`, and `9`.
2. Follow `task2_Dependency.md` section `5.4 Migrations`.
3. Ensure Task 3 ingestion env variables from section `7.1` above are present.
4. Start MongoDB and Redis.
5. Load `.env`.
6. Run tests.
7. Start service.
8. Send an event to `POST /api/v1/sessions/events`.
9. Verify MongoDB and Redis.

Commands:

```bash
docker start ecommerce-session-mongo ecommerce-session-redis

cd backend/services/session-service
go mod download

set -a
source .env
set +a

go test ./...
go run ./cmd/server
```

Health check:

```bash
curl http://localhost:8086/healthz
curl http://localhost:8086/readyz
```

Expected readiness shape:

```json
{
  "status": "ok",
  "checks": {
    "mongo": "ok",
    "redis": "ok"
  }
}
```

---

## 10. Running the Project

### 10.1 Daily Run Flow

Use previous daily flow:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
10. Running the Project
```

Task 3 API test:

```bash
NOW_UTC=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

curl -i -X POST http://localhost:8086/api/v1/sessions/events \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: req_local_task3_001" \
  -H "X-Client-Channel: user_app_web" \
  -H "X-Device-Type: desktop" \
  -d '{
    "event_type": "page_view",
    "anonymous_id": "anon_local_001",
    "session_id": "sess_local_001",
    "occurred_at": "'"$NOW_UTC"'",
    "path": "/",
    "properties": {
      "title": "Home"
    }
  }'
```

Expected response:

```http
HTTP/1.1 202 Accepted
Content-Type: application/json
```

Current implementation response includes `event_id`:

```json
{
  "accepted": true,
  "request_id": "req_local_task3_001",
  "event_id": "evt_..."
}
```

### 10.2 Verify MongoDB Writes

```bash
mongosh "mongodb://localhost:27017/session_db" --eval "db.session_events.find({session_id:'sess_local_001'}).sort({occurred_at:1}).toArray()"
mongosh "mongodb://localhost:27017/session_db" --eval "db.sessions.findOne({session_id:'sess_local_001'})"
```

### 10.3 Verify Redis Touch

```bash
redis-cli -n 2 HGETALL session:active:sess_local_001
redis-cli -n 2 TTL session:active:sess_local_001
```

TTL should be positive. Default active session TTL is `35m`.

### 10.4 Ports & Networking

| Service | Port | Purpose | Task 3 status |
|---|---:|---|---|
| Session Backend API | `8086` | Event ingestion, health, readiness | Reused |
| MongoDB | `27017` | Raw event and session storage | Reused |
| Redis | `6379` | Active session TTL/cache | Reused |
| API Gateway | Usually platform-specific, often `8080` | Production public entry point | Expected in prod, optional local |
| Kafka/RabbitMQ | N/A | Future `session.events` stream | Not required now |

Port conflict troubleshooting:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
11. Ports & Networking
```

---

## 11. Common Errors & Fixes

Common setup errors already documented:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
12. Common Errors & Fixes

TaskImplementation/Session Management Service/task2_Dependency.md
Section:
11. Common Errors & Fixes
```

Task 3-specific additions:

### 1. `415 UNSUPPORTED_MEDIA_TYPE`

Cause:

`Content-Type` header missing or not `application/json`.

Fix:

```bash
curl -X POST http://localhost:8086/api/v1/sessions/events \
  -H "Content-Type: application/json" \
  -d '{"event_type":"page_view"}'
```

### 2. `413 PAYLOAD_TOO_LARGE`

Cause:

Request body exceeds `SESSION_INGEST_MAX_BODY_BYTES`.

Fix:

```env
SESSION_INGEST_MAX_BODY_BYTES=65536
```

Also reduce event `properties`. Do not send screenshots, full HTML, full form text, or huge nested objects.

### 3. `400 INVALID_REQUEST` With Unknown Field

Cause:

Handler uses strict JSON decoding. Unknown top-level fields reject request.

Allowed top-level fields:

```text
event_type
anonymous_id
session_id
user_id
occurred_at
path
properties
```

Fix:

Move custom event details inside `properties`.

### 4. `occurred_at must be RFC3339`

Cause:

Timestamp format wrong.

Good:

```json
"occurred_at": "2026-05-22T10:15:30Z"
```

Bad:

```json
"occurred_at": "22-05-2026 10:15"
```

### 5. `occurred_at is too far in the future`

Cause:

Client clock ahead more than `SESSION_ALLOWED_CLOCK_SKEW`.

Fix:

Sync client/server clock. Local testing me current UTC time use karo.

### 6. `occurred_at is too old`

Cause:

Event older than `SESSION_MAX_EVENT_AGE`, default `24h`.

Fix:

For local tests, fresh timestamp use karo. Offline/batch import ke liye future separate design needed hoga.

### 7. `anonymous_id must start with anon_`

Cause:

Current validation expects anonymous IDs prefixed with `anon_`.

Fix:

```json
"anonymous_id": "anon_local_001"
```

### 8. `session_id must start with sess_`

Cause:

Current validation expects session IDs prefixed with `sess_`.

Fix:

```json
"session_id": "sess_local_001"
```

### 9. Sensitive Property Rejected Or Missing

Cause:

Sensitive fields like `password`, `otp`, `token`, `cookie`, `card_number`, `cvv`, `keystroke`, `input_value`, etc. are blocked/sanitized.

Fix:

Do not send secrets or form text in analytics events. Send safe metadata only, like button ID, product ID, page path, or scroll depth.

### 10. Redis Touch Fails But API Still Returns `202`

Cause:

MongoDB durable write succeeded but Redis active session update failed.

Meaning:

This is graceful degradation. Event is stored, but live active-session view may be stale.

Fix:

Check Redis:

```bash
redis-cli ping
curl http://localhost:8086/readyz
```

### 11. `503 INGEST_STORAGE_UNAVAILABLE`

Cause:

MongoDB insert or session touch failed.

Fix:

```bash
mongosh "mongodb://localhost:27017" --eval "db.runCommand({ ping: 1 })"
curl http://localhost:8086/readyz
```

If readiness says Mongo error, start MongoDB and verify `.env`.

---

## 12. Security & Best Practices

### 12.1 Security Audit For Task 3

| Finding | Risk | Recommendation |
|---|---|---|
| Public event ingestion endpoint | Bot/spam/high write volume risk | Put API Gateway rate limits in front |
| Body `user_id` should not be trusted | Analytics identity spoofing | Use trusted gateway/auth headers only |
| `SESSION_IP_HASH_SALT` may be empty locally | IP hash becomes `unavailable` | Set a long random secret in every environment |
| Raw IP can come via forwarded headers | Privacy risk if stored directly | Current code hashes IP when salt exists; prefer gateway-provided `X-IP-Hash` |
| Event properties are flexible | Secret/PII leakage risk | Keep sanitizer and validation strict |
| Large payloads can abuse CPU/storage | Cost and performance risk | Keep `SESSION_INGEST_MAX_BODY_BYTES=65536` unless justified |
| No Kafka/RabbitMQ publisher wired | Downstream analytics not async yet | Add publisher/outbox later if needed |
| No committed Dockerfile/compose | Onboarding/deploy inconsistency | Add service Docker artifacts in future platform task |
| `.env` is not auto-loaded | Developers may run with defaults accidentally | Source `.env` before `go run` |
| No explicit local API Gateway config | Header trust differs between local/prod | Document gateway header stripping/injection in deployment config |

### 12.2 Event Privacy Rules

Do not send or store:

```text
password
otp
card_number
credit_card
cvv
pin
token
authorization
cookie
raw_ip
private_message
secret
keystroke
field_value
input_value
```

Safe examples:

```text
product_id
variant_id
category_id
button_id
element_id
path
title
referrer
scroll depth
viewport width/height
```

### 12.3 Beginner Best Practices

- Start MongoDB and Redis before running the Go service.
- Always send `Content-Type: application/json`.
- Use fresh `occurred_at` timestamps while testing.
- Keep event payload small and safe.
- Put custom analytics details inside `properties`.
- Never commit real `.env` values.
- Use API Gateway for production rate limiting and trusted identity headers.
- Use `SCAN`, not `KEYS`, on production Redis.
- Treat MongoDB as source of truth; Redis is fast live state only.

Previously documented best practices:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
13. Security & Best Practices

TaskImplementation/Session Management Service/task2_Dependency.md
Section:
12. Security & Best Practices
```

---

## 13. Missing or Misconfigured Things

Task 3 dependency/devops gaps found:

| Item | Status | Impact | Suggested fix |
|---|---|---|---|
| Service Dockerfile | Missing | Cannot build official container image yet | Add Dockerfile for `backend/services/session-service` |
| Local docker-compose file | Missing | Beginners must manually run MongoDB/Redis | Add `docker-compose.session.local.yml` or platform compose |
| Kafka/RabbitMQ publisher | Not wired | `session.events` topic is future-only | Add publisher plus retry/outbox if required |
| `SESSION_EVENTS_TOPIC` | Not used by current code | Adding it now has no effect | Introduce only when publisher is implemented |
| API Gateway rate-limit config | Not in service repo | Public endpoint can be abused in prod | Add gateway/IaC config for per-IP/session limits |
| CORS config | Not handled in service | Browser direct calls may fail or be unsafe | Prefer gateway-level CORS allowlist |
| `.env` auto-loading | Not implemented | Local `.env` ignored unless sourced | Use shell export or add safe local runner |
| Migration tracking collection | Missing | Hard to know which migrations ran | Add migration tool/tracking later |
| `SESSION_IP_HASH_SALT` example is weak | Example only | Bad if copied to prod | Generate strong random secret per environment |

---

## 14. References to Previous Dependency Files

Use these instead of duplicating setup:

| Topic | Refer |
|---|---|
| Go installation and required software | `TaskImplementation/Session Management Service/task1_Dependency.md`, section `3. Required Software` |
| Go modules, `go.mod`, `go.sum`, `go mod tidy` | `task1_Dependency.md`, section `4. Dependency Management` |
| MongoDB install and Docker run | `task1_Dependency.md`, section `5.1 MongoDB Setup` |
| Redis install and Docker run | `task1_Dependency.md`, section `5.2 Redis Setup` |
| Complete `.env` example | `task1_Dependency.md`, section `7. Environment Variables` |
| Docker Compose example | `task1_Dependency.md`, section `8. Docker Setup` |
| Full local onboarding flow | `task1_Dependency.md`, section `9. Local Development Setup` |
| Ports and networking | `task1_Dependency.md`, section `11. Ports & Networking` |
| General troubleshooting | `task1_Dependency.md`, section `12. Common Errors & Fixes` |
| MongoDB collections/indexes/TTL | `task2_Dependency.md`, sections `5.2`, `5.3`, `5.4` |
| Redis key design and TTL rules | `task2_Dependency.md`, sections `6.2`, `6.3` |
| Storage-specific security notes | `task2_Dependency.md`, section `12. Security & Best Practices` |

---

## 15. Final Checklist

Before saying Task 3 local setup is ready, verify:

| Check | Done |
|---|---|
| Repo cloned and `backend/services/session-service` exists |  |
| Go version is `1.26.3` or newer |  |
| `go mod download` succeeds |  |
| MongoDB is running on `27017` |  |
| Redis is running on `6379` |  |
| `.env` exists at `backend/services/session-service/.env` |  |
| `.env` has MongoDB and Redis values from previous docs |  |
| `.env` has Task 3 ingestion variables |  |
| `SESSION_IP_HASH_SALT` is set to a long random secret locally |  |
| `.env` is sourced/exported before `go run` |  |
| MongoDB migrations/indexes are applied or startup created indexes |  |
| `go test ./...` passes |  |
| `go run ./cmd/server` starts on `:8086` |  |
| `GET /healthz` returns `200` |  |
| `GET /readyz` returns Mongo and Redis `ok` |  |
| `POST /api/v1/sessions/events` returns `202 Accepted` |  |
| MongoDB `session_events` contains the test event |  |
| MongoDB `sessions` contains/touched the test session |  |
| Redis active session key exists with positive TTL |  |
| API Gateway rate limit/CORS/trusted header config planned for production |  |

Quick final commands:

```bash
cd backend/services/session-service
go test ./...
go run ./cmd/server
```

In another terminal:

```bash
NOW_UTC=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

curl http://localhost:8086/readyz
curl -i -X POST http://localhost:8086/api/v1/sessions/events \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: req_local_task3_check" \
  -d '{
    "event_type": "page_view",
    "anonymous_id": "anon_local_001",
    "session_id": "sess_local_001",
    "occurred_at": "'"$NOW_UTC"'",
    "path": "/",
    "properties": {
      "title": "Home"
    }
  }'
```

If `202 Accepted` comes and MongoDB/Redis verification passes, Task 3 dependency/setup side is ready.
