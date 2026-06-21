# Project Dependency & Setup Guide

Input implementation file:

```text
TaskImplementation/Auth Service/task7.md
```

Generated dependency file:

```text
TaskImplementation/Auth Service/task7_Dependency.md
```

This guide is for **Auth Service - Task 7: Session Link**.

Simple goal: beginner developer ko clearly samajh aaye ki login/logout/refresh-reuse session events ke liye kaunsi dependency, env variable, database migration, event publisher setup, and debugging flow chahiye.

Important reuse rule followed:

- Previous Auth Service dependency files were inspected first.
- Common clone, Go install, MySQL install, Redis install, JWT keys, OTP setup, RBAC setup, Docker examples, and full local run flow already documented hai.
- Is file me wahi content repeat nahi kiya gaya. Sirf Task 7 session-link/outbox specific setup and differences explain kiye gaye hain.

Previous dependency files inspected:

```text
TaskImplementation/Auth Service/Dependency/task1_Dependency.md
TaskImplementation/Auth Service/Dependency/task2_Dependency.md
TaskImplementation/Auth Service/Dependency/task3_Dependency.md
TaskImplementation/Auth Service/Dependency/task4_Dependency.md
TaskImplementation/Auth Service/Dependency/task5_Dependency.md
TaskImplementation/Auth Service/Dependency/task6_Dependency.md
```

Current repository note:

- `task7.md` is a design/implementation guide.
- Current repo already contains Task 7 code under `backend/services/auth-service/internal/sessionlink/`, `internal/events/`, `internal/repository/mysql_outbox_repository.go`, and migration `005_create_auth_outbox_events`.
- Current Auth Service supports `SESSION_LINK_MODE=outbox` and `SESSION_LINK_MODE=disabled`.
- `SESSION_GRPC_ADDR` exists in config, but current validation does not allow `grpc` mode yet.
- Kafka/RabbitMQ are architecture options in docs, but current Auth Service code does not import a Kafka or RabbitMQ client.

---

## 1. Project Overview

Task 7 adds a session-link integration.

Session link ka matlab: jab Auth Service login/logout/refresh-token-reuse detect kare, to Session Management Service ko safe event mil sake. Isse anonymous browser journey logged-in user se link hoti hai, active sessions update hote hain, and fraud signals generate ho sakte hain.

Current Task 7 runtime responsibility:

| Area | Current repo status | Why it matters |
|---|---|---|
| Session event contracts | Implemented in `internal/sessionlink/event.go` | Stable event envelope and payload fields define hote hain |
| Event normalization | Implemented in `internal/sessionlink/normalize.go` | Oversized/malformed metadata clean hota hai |
| Privacy hashing | Implemented in `internal/sessionlink/privacy.go` | Raw IP/device fingerprint event me nahi jata |
| Outbox linker | Implemented in `internal/sessionlink/linker.go` | Auth usecases events ko MySQL outbox me enqueue karte hain |
| Outbox table | Implemented by migration `005_create_auth_outbox_events.up.sql` | Events durable rahte hain even if publisher down ho |
| Outbox worker | Implemented in `internal/events/outbox_worker.go` | Pending/failed rows publish and retry hote hain |
| HTTP publisher | Implemented in `internal/events/publisher.go` | Event ingress endpoint ko HTTP POST karta hai |
| Login event trigger | Implemented in `internal/usecase/auth.go` | Successful login ke baad `AuthLoginSucceeded` event enqueue hota hai |
| Logout/reuse trigger | Implemented in `internal/usecase/token.go` | Logout and refresh-token reuse events enqueue hote hain |

Task 7 does **not** introduce:

- New Auth Service web framework
- New Go module dependency in current code
- Direct Kafka/RabbitMQ/NATS client
- Direct Session Service gRPC implementation
- New Auth Service HTTP port
- New Dockerfile or docker-compose file

Baseline full service setup already exists in:

```text
TaskImplementation/Auth Service/Dependency/task1_Dependency.md
Sections:
3. Required Software
4. Dependency Management
5. Database Setup
6. Redis / Queue / External Services
7. Environment Variables
8. Docker Setup
9. Local Development Setup
10. Running the Project
```

---

## 2. Tech Stack

### Task 7 specific technologies

| Technology | What it is | Why Task 7 uses it | Required? | Beginner explanation |
|---|---|---|---:|---|
| Go | Backend programming language | Session event builders, linkers, worker, publisher Go me hain | Yes | Go ek compiled backend language hai. Is service ka runtime Go code se chalta hai. |
| Go standard `net/http` | Built-in HTTP client/server package | Outbox worker event ingress ko HTTP POST karta hai | Yes | Current code Fiber/Gin/Echo ya broker SDK use nahi karta. Plain HTTP publisher hai. |
| MySQL 8+ | Relational database | `auth_outbox_events` table me events durable save hote hain | Yes for outbox mode | MySQL tables me events store karke retry possible hota hai. |
| Outbox pattern | Reliable event delivery pattern | Login/logout ke saath event loss avoid karne ke liye | Yes for integrated mode | Pehle DB me event save hota hai, baad me worker publish karta hai. |
| HMAC SHA-256 | Hashing method | IP/device fingerprint ko pepper ke saath hash karne ke liye | Yes for outbox mode | Raw private data event me nahi jata; hash fraud correlation ke liye enough hota hai. |
| Event ingress HTTP endpoint | External service/API | Auth outbox worker events ko aage broker/Session Service tak bhejne ke liye | Optional local, required for publishing | Endpoint empty ho to events DB me pending rahenge. |
| Session Management Service | Independent service | Auth events consume karke active session/journey link update karega | Not required to start Auth locally | Auth Service sirf event produce karta hai; Session Service consume karega. |

### Already documented baseline technologies

| Baseline topic | Reuse documentation |
|---|---|
| Go `1.26.3`, Go modules, Go workspace | `TaskImplementation/Auth Service/Dependency/task1_Dependency.md`, section `4. Dependency Management` |
| MySQL 8+ installation, Docker, DSN | `TaskImplementation/Auth Service/Dependency/task1_Dependency.md`, section `5. Database Setup` |
| Redis setup for full Auth startup | `TaskImplementation/Auth Service/Dependency/task1_Dependency.md`, section `6. Redis / Queue / External Services` |
| JWT keys and refresh token config | `TaskImplementation/Auth Service/Dependency/task4_Dependency.md`, sections `7` and `8` |
| OTP/Notification setup | `TaskImplementation/Auth Service/Dependency/task5_Dependency.md` |
| RBAC role setup | `TaskImplementation/Auth Service/Dependency/task6_Dependency.md` |
| Full `.env` baseline | `TaskImplementation/Auth Service/Dependency/task1_Dependency.md`, section `7. Environment Variables` |

---

## 3. Dependency Management

This is still a Go module project.

Full beginner explanation of:

- `go.mod`
- `go.sum`
- Go modules
- `go mod download`
- `go mod tidy`
- `go build`
- `go run`
- common Go module issues

already exists in:

```text
TaskImplementation/Auth Service/Dependency/task1_Dependency.md
Section: 4. Dependency Management
```

### Task 7 current dependency impact

Current `backend/services/auth-service/go.mod` direct dependencies remain:

```text
github.com/go-sql-driver/mysql v1.9.3
golang.org/x/crypto v0.51.0
```

Task 7 current code uses Go standard library packages for:

| Package area | Why used |
|---|---|
| `crypto/hmac`, `crypto/sha256` | IP/device fingerprint privacy hashing |
| `crypto/rand`, `encoding/base64` | Event ID generation |
| `encoding/json` | Event envelope and payload JSON |
| `net/http` | HTTP event publisher |
| `database/sql` | MySQL outbox repository |
| `log/slog` | Structured enqueue/publish logs |

Do **not** run `go get` for `grpc`, `rabbitmq/amqp091-go`, `kafka-go`, or OpenTelemetry only because `task7.md` mentions them as design options. Current code does not import those packages.

### Task 7 verification commands

Run session-link focused tests:

```bash
cd backend/services/auth-service
go test ./internal/sessionlink ./internal/events ./internal/usecase
```

Run all Auth Service tests:

```bash
cd backend/services/auth-service
go test ./...
```

If Go build cache is not writable:

```bash
cd backend/services/auth-service
GOCACHE=/tmp/auth-go-cache go test ./...
```

---

## 4. Database Setup

### Database detected

Task 7 uses the existing Auth Service MySQL database:

```text
auth_db
```

New Task 7 table:

```text
auth_outbox_events
```

Migration files:

```text
backend/services/auth-service/migrations/005_create_auth_outbox_events.up.sql
backend/services/auth-service/migrations/005_create_auth_outbox_events.down.sql
```

### What MySQL is

MySQL ek relational database hai jisme data tables ke form me store hota hai. Task 7 me MySQL event outbox ke liye use hota hai, taaki event publish fail hone par bhi event lose na ho.

### Why Task 7 uses MySQL

Login/logout flow already Auth DB use karta hai. Session event ko same Auth-owned database me pending outbox row ke form me store karna reliable hai.

Simple Hinglish:

```text
Token ban gaya, par event publisher down tha?
No problem. Event auth_outbox_events table me pending rahega.
Worker baad me retry karega.
```

### Required or optional

| Mode | MySQL outbox table required? | Notes |
|---|---:|---|
| `SESSION_LINK_MODE=disabled` | No for session-link feature, yes for full service schema consistency | Beginner local debugging me session events off ho sakte hain |
| `SESSION_LINK_MODE=outbox` | Yes | `auth_outbox_events` table missing hua to login/logout enqueue fail ho sakta hai |

Full Auth Service still requires MySQL for accounts, credentials, refresh tokens, OTP, and roles. That setup is already covered in:

```text
TaskImplementation/Auth Service/Dependency/task1_Dependency.md
Section: 5. Database Setup
```

### New migration for Task 7

If earlier migrations `001` to `004` are already applied, run only:

```bash
cd backend/services/auth-service
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/005_create_auth_outbox_events.up.sql
```

If setting up the current full service from scratch, apply all migrations in order as documented in:

```text
TaskImplementation/Auth Service/Dependency/task1_Dependency.md
Section: 5. Database Setup
Subsection: Run migrations
```

### Verify Task 7 table

```bash
mysql -h 127.0.0.1 -P 3306 -u auth_user -p auth_db
```

Inside MySQL:

```sql
SHOW TABLES LIKE 'auth_outbox_events';
SHOW COLUMNS FROM auth_outbox_events;
SHOW INDEX FROM auth_outbox_events;
```

Useful event status check:

```sql
SELECT status, COUNT(*) AS total
FROM auth_outbox_events
GROUP BY status;
```

Recent events:

```sql
SELECT event_id, event_type, routing_key, aggregate_id, status, attempts, created_at, last_error
FROM auth_outbox_events
ORDER BY created_at DESC
LIMIT 10;
```

### Outbox statuses

| Status | Meaning | Beginner explanation |
|---|---|---|
| `pending` | Waiting to be published | Event DB me safe hai, worker publish karega |
| `processing` | Worker locked the event | Worker currently publish attempt kar raha hai |
| `published` | Publish succeeded | Event endpoint ne 2xx response diya |
| `failed` | Publish failed but will retry | Backoff ke baad retry hoga |
| `dead_letter` | Max attempts reached | Manual review/fix required |

### Rollback warning

Rollback file:

```bash
cd backend/services/auth-service
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/005_create_auth_outbox_events.down.sql
```

Warning: this drops `auth_outbox_events`, so pending/published event history delete ho jayegi. Production me backup and approval ke bina rollback mat karo.

---

## 5. Environment Variables

Full `.env` setup already exists in:

```text
TaskImplementation/Auth Service/Dependency/task1_Dependency.md
Section: 7. Environment Variables
```

Task 7 specific variables are below.

### First beginner local run

Use this when you only want Auth Service login/logout to run without session event publishing:

```env
SESSION_LINK_MODE=disabled
```

In this mode:

- `SESSION_EVENT_PEPPER` is not required.
- `AUTH_EVENTS_PUBLISH_ENDPOINT` is not required.
- Login/logout will not enqueue session-link events.

### Integrated local/staging run

Use this when you want Auth Service to write outbox rows and optionally publish them:

```env
SESSION_LINK_MODE=outbox
SESSION_EVENT_PEPPER=local_session_event_pepper_change_me_32_chars
AUTH_EVENTS_TOPIC=auth.events
AUTH_EVENTS_PUBLISH_ENDPOINT=http://localhost:8085/internal/v1/events
SESSION_LINK_TIMEOUT=150ms
OUTBOX_WORKER_BATCH_SIZE=100
OUTBOX_WORKER_INTERVAL=1s
OUTBOX_MAX_ATTEMPTS=5
OUTBOX_INITIAL_BACKOFF=5s
OUTBOX_MAX_BACKOFF=10m
OUTBOX_STALE_LOCK_TIMEOUT=5m
```

### Variable details

| Variable | Required? | Example | Purpose | Security note |
|---|---:|---|---|---|
| `SESSION_LINK_MODE` | Optional | `disabled` or `outbox` | Enables/disables session link behavior | Use `disabled` for simple local run; use `outbox` for integrated env |
| `SESSION_EVENT_PEPPER` | Required when mode is `outbox` | long random string | HMAC hashes IP/device fingerprint values | Secret value; never commit |
| `AUTH_EVENTS_TOPIC` | Required when mode is `outbox` | `auth.events` | Logical topic name sent in `X-Event-Topic` header | Must match event-ingress/broker mapping |
| `AUTH_EVENTS_PUBLISH_ENDPOINT` | Optional | `http://localhost:8085/internal/v1/events` | HTTP endpoint where outbox worker posts event JSON | Empty means worker disabled and rows remain pending |
| `SESSION_LINK_TIMEOUT` | Optional | `150ms` | HTTP publish timeout | Keep short so publisher does not hang |
| `SESSION_LINK_TIMEOUT_MS` | Optional alias | `150` | Millisecond-style timeout fallback | Prefer `SESSION_LINK_TIMEOUT` for clarity |
| `OUTBOX_WORKER_BATCH_SIZE` | Optional | `100` | Events locked per worker run | Increase only after measuring DB/publisher capacity |
| `OUTBOX_WORKER_INTERVAL` | Optional | `1s` | Poll interval | Lower interval means more DB polling |
| `OUTBOX_WORKER_INTERVAL_MS` | Optional alias | `1000` | Millisecond-style interval fallback | Prefer `OUTBOX_WORKER_INTERVAL` for clarity |
| `OUTBOX_MAX_ATTEMPTS` | Optional | `5` | Failed publish attempts before dead-letter | Monitor dead letters |
| `OUTBOX_INITIAL_BACKOFF` | Optional | `5s` | First retry delay | Must be positive |
| `OUTBOX_MAX_BACKOFF` | Optional | `10m` | Maximum retry delay | Must be greater than or equal to initial backoff |
| `OUTBOX_STALE_LOCK_TIMEOUT` | Optional | `5m` | Reclaims stuck `processing` events | Helps if worker crashes |
| `SESSION_GRPC_ADDR` | Optional/reserved | `session-service:9090` | Future direct Session Service gRPC address | Current code does not use it |

### Where `.env` should be created

Local Auth Service env file:

```text
backend/services/auth-service/.env
```

Important: current Go app does **not** auto-load `.env`. Before running the server:

```bash
cd backend/services/auth-service
set -a
source .env
set +a
go run ./cmd/server
```

Common mistake:

```text
Developer creates .env but does not source it.
App starts with missing env values and fails validation.
```

---

## 6. External Services Analysis

### MySQL

MySQL setup is reused.

Refer:

```text
TaskImplementation/Auth Service/Dependency/task1_Dependency.md
Section: 5. Database Setup
```

Task 7 only adds migration `005_create_auth_outbox_events.up.sql`.

### Redis

Auth Service full startup still pings Redis because OTP rate limiting uses Redis. Task 7 itself does not add a new Redis keyspace in Auth Service.

Refer:

```text
TaskImplementation/Auth Service/Dependency/task1_Dependency.md
Section: 6. Redis / Queue / External Services
Subsection: Redis
```

Session Management Service may use Redis for active sessions, but that is Session Service setup, not an Auth Service runtime dependency in this repo.

### Event ingress HTTP endpoint

This is the main Task 7 external publishing dependency.

What it is:

```text
An internal HTTP endpoint that accepts Auth outbox events and forwards/stores them for Session Service.
```

Current Auth Service publisher behavior:

| Request part | Value |
|---|---|
| Method | `POST` |
| URL | `AUTH_EVENTS_PUBLISH_ENDPOINT` |
| Body | Event envelope JSON from `auth_outbox_events.payload` |
| Header | `Content-Type: application/json` |
| Header | `X-Event-Topic: auth.events` |
| Header | `X-Event-ID: evt_...` |
| Header | `X-Event-Type: AuthLoginSucceeded` |
| Header | `X-Event-Routing-Key: auth.login_succeeded` |
| Success condition | Any HTTP `2xx` response |

Mandatory or optional:

- Optional for first local run.
- Required if you want outbox rows to become `published`.
- If empty with `SESSION_LINK_MODE=outbox`, the server starts and logs `auth.outbox.publisher_disabled`.

Health check:

```bash
curl -i http://localhost:8085/healthz
```

Actual path depends on your event-ingress implementation. The Auth Service only needs `AUTH_EVENTS_PUBLISH_ENDPOINT` to be a valid absolute URL.

### Kafka/RabbitMQ/NATS

Current Auth Service code does **not** directly use Kafka, RabbitMQ, or NATS.

Architecture docs mention MQ because the wider platform may route `auth.events` through a broker. In current code, broker integration should sit behind the event-ingress HTTP endpoint.

Do not install direct broker Go clients unless code is changed to import them.

### Session Management Service

What it is:

```text
Independent service that consumes Auth session events and updates active sessions, journey linking, analytics, and fraud signals.
```

Auth Service dependency status:

| Item | Status |
|---|---|
| Direct runtime call from Auth to Session Service | Not currently implemented |
| `SESSION_GRPC_ADDR` | Config exists but mode `grpc` is not valid in current code |
| Event consumption | Expected through event-ingress/broker/Session Service consumer |
| MongoDB for Session Service | Not required to start Auth Service |
| Redis for Session Service active sessions | Separate from Auth Service Redis usage |

Refer for Session Service concepts:

```text
docs/08-session-management-system.md
```

---

## 7. Ports & Networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Auth Service HTTP | `8081` default | Login, refresh, logout, JWKS, internal auth APIs | Reused |
| MySQL | `3306` | Auth database and outbox table | Reused |
| Redis | `6379` | OTP rate limit store for full Auth startup | Reused |
| Notification Service | `8084` example | OTP delivery endpoint | Reused |
| Event ingress HTTP | no fixed port | Receives outbox event POSTs | New/optional |
| Session Service gRPC | `9090` example | Reserved future direct session link | Reserved, not wired |
| MongoDB | `27017` default | Session Service event storage | External to Auth |
| Kafka/RabbitMQ | varies | Platform event broker behind ingress | External to Auth |

Port conflict references:

```text
TaskImplementation/Auth Service/Dependency/task1_Dependency.md
Section: 11. Ports & Networking
```

Task 7 specific networking notes:

- `AUTH_EVENTS_PUBLISH_ENDPOINT` must be reachable from the Auth Service process.
- If Auth runs on host and event ingress runs in Docker with `-p 8085:8085`, use `http://127.0.0.1:8085/...`.
- If Auth runs inside Docker Compose, use the Compose service name, for example `http://event-ingress:8085/internal/v1/events`.
- If endpoint returns non-2xx, outbox row becomes `failed` and later `dead_letter` after max attempts.

---

## 8. Docker & DevOps Setup

The Auth Service Dockerfile and root Compose service are present. Compose applies migration `005`, waits for Redis and Notification Service readiness, and uses `/readyz` for Auth health.

Existing local dependency Docker setup is reused:

```text
TaskImplementation/Auth Service/Dependency/task1_Dependency.md
Section: 8. Docker Setup
```

### What changes for Task 7

| DevOps item | Change |
|---|---|
| MySQL container | No new container, but migration `005` must be applied |
| Redis container | No Task 7 change for Auth Service |
| Auth Service container | Dockerfile and root Compose service present |
| Event ingress | Optional new external service if publishing is enabled |
| Broker | Not direct Auth dependency; can be behind event ingress |
| Volumes | Existing MySQL volume stores `auth_outbox_events` |
| Health checks | Add monitoring for pending/failed/dead-letter outbox rows |

### Production deployment assumptions

Use secret manager/Kubernetes Secrets for:

```text
SESSION_EVENT_PEPPER
AUTH_MYSQL_DSN
REFRESH_TOKEN_PEPPER
OTP_HASH_PEPPER
OTP_RATE_LIMIT_PEPPER
JWT private key material/path
```

Recommended operational alerts:

| Alert | Why |
|---|---|
| `dead_letter` count > 0 | Event publishing permanently failing |
| `pending` rows older than expected | Worker disabled or event ingress down |
| `failed` rows increasing | Event ingress returning non-2xx or network issue |
| publish latency high | Event ingress/broker slow |
| `auth.session_link.enqueue_failed` logs | DB/table/config issue affecting event creation |

---

## 9. Project Run Instructions

### Option A: Beginner local run without session events

Follow the base run flow:

```text
TaskImplementation/Auth Service/Dependency/task1_Dependency.md
Sections: 9. Local Development Setup, 10. Running the Project
```

Task 7 local override:

```env
SESSION_LINK_MODE=disabled
```

Then:

```bash
cd backend/services/auth-service
set -a
source .env
set +a
go run ./cmd/server
```

Expected result:

- Auth Service starts.
- Login can return `session_id`.
- No session-link events are written.

### Option B: Integrated run with outbox rows but no publisher

Use this when you want to verify DB event creation but do not have event ingress ready:

```env
SESSION_LINK_MODE=outbox
SESSION_EVENT_PEPPER=local_session_event_pepper_change_me_32_chars
AUTH_EVENTS_TOPIC=auth.events
AUTH_EVENTS_PUBLISH_ENDPOINT=
```

Expected result:

- Auth Service starts.
- It logs `auth.outbox.publisher_disabled`.
- Login/logout/reuse events insert into `auth_outbox_events`.
- Rows remain `pending`.

### Option C: Integrated run with event publishing

Use this when event ingress is available:

```env
SESSION_LINK_MODE=outbox
SESSION_EVENT_PEPPER=local_session_event_pepper_change_me_32_chars
AUTH_EVENTS_TOPIC=auth.events
AUTH_EVENTS_PUBLISH_ENDPOINT=http://localhost:8085/internal/v1/events
SESSION_LINK_TIMEOUT=150ms
```

Expected result:

- Auth Service starts.
- Outbox worker starts.
- Pending rows are posted to event ingress.
- Successful rows become `published`.

### Manual smoke test

Login request example:

```bash
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "X-Anonymous-ID: anon_456" \
  -H "X-Device-Fingerprint: local-device-1" \
  -H "X-Client-Channel: web" \
  -d '{
    "identifier": "buyer@example.com",
    "password": "correct-password",
    "device": {
      "anonymous_id": "anon_456",
      "channel": "web",
      "locale": "en-US"
    }
  }'
```

Expected response shape:

```json
{
  "user": {
    "user_id": "user_123",
    "roles": ["buyer"]
  },
  "tokens": {
    "access_token": "...",
    "refresh_token": "...",
    "expires_in": 900
  },
  "session_id": "sess_..."
}
```

Verify outbox:

```sql
SELECT event_id, event_type, routing_key, aggregate_id, status
FROM auth_outbox_events
ORDER BY created_at DESC
LIMIT 5;
```

Expected event:

```text
AuthLoginSucceeded
```

Possible status:

| Status | Meaning |
|---|---|
| `pending` | Event saved, publisher endpoint empty or worker has not run yet |
| `published` | Event ingress returned 2xx |
| `failed` | Publish failed and will retry |
| `dead_letter` | Publish failed too many times |

---

## 10. Event Contract Summary

Current event types:

| Event type | Routing key | Trigger |
|---|---|---|
| `AuthLoginSucceeded` | `auth.login_succeeded` | Successful public login |
| `AuthSignupSucceeded` | `auth.signup_succeeded` | Supported by event builder/linker, but no public signup route is currently wired |
| `AuthLogoutSucceeded` | `auth.logout_succeeded` | Successful logout |
| `RefreshTokenReuseDetected` | `auth.refresh_token_reuse_detected` | Revoked refresh token reused |

Envelope fields:

| Field | Meaning |
|---|---|
| `event_id` | Unique id for idempotency |
| `event_type` | Type like `AuthLoginSucceeded` |
| `version` | Current event schema version, `1` |
| `occurred_at` | UTC event time |
| `producer` | `auth-service` |
| `trace_id` | Request trace/request id if present |
| `aggregate_type` | Usually `session`, sometimes `account` for all-device logout |
| `aggregate_id` | Session ID or account ID |
| `payload` | Event-specific safe metadata |

Sensitive data must not appear in event payload:

- Password
- Access token
- Refresh token
- OTP
- Raw IP address
- Raw device fingerprint
- Authorization header
- Cookies

---

## 11. Common Errors & Fixes

| Error/symptom | Cause | Fix |
|---|---|---|
| `SESSION_EVENT_PEPPER cannot be empty` | `SESSION_LINK_MODE=outbox` but pepper missing | Set `SESSION_EVENT_PEPPER` or use `SESSION_LINK_MODE=disabled` |
| `SESSION_LINK_MODE must be outbox or disabled` | Tried `grpc` from design doc | Use `outbox` or `disabled`; gRPC mode is not implemented in current code |
| Login succeeds but no outbox row appears | `SESSION_LINK_MODE=disabled` | Switch to `SESSION_LINK_MODE=outbox` |
| Login logs `auth.session_link.enqueue_failed` | Missing migration/table or DB insert error | Apply migration `005` and verify `auth_outbox_events` |
| Rows stay `pending` forever | `AUTH_EVENTS_PUBLISH_ENDPOINT` empty or worker not running | Set endpoint and restart service |
| Rows become `failed` | Endpoint unreachable or non-2xx response | Check endpoint URL, network, logs, and status code |
| Rows become `dead_letter` | Publish failed more than `OUTBOX_MAX_ATTEMPTS` | Fix endpoint, inspect `last_error`, manually requeue if safe |
| `AUTH_EVENTS_PUBLISH_ENDPOINT must be an absolute URL` | Endpoint missing scheme/host | Use `http://localhost:8085/internal/v1/events` |
| Event ingress receives event but Session Service does nothing | Consumer not connected or topic mismatch | Verify `X-Event-Topic=auth.events` and consumer subscription |
| Raw IP appears in payload | Upstream sent raw hash field incorrectly or code changed | Use `SESSION_EVENT_PEPPER`; pass `X-IP-Hash` only if already hashed |
| Duplicate session events processed | Consumer not idempotent | Session Service should dedupe by `event_id` |
| Beginner expects Kafka container to be required | Current Auth code publishes HTTP only | Use event ingress as broker adapter; do not add direct broker dependency unless code changes |

Requeue dead-letter manually only after fixing the root cause:

```sql
UPDATE auth_outbox_events
SET status = 'pending',
    attempts = 0,
    next_attempt_at = NULL,
    locked_at = NULL,
    locked_by = NULL,
    last_error = NULL
WHERE event_id = 'evt_replace_me'
  AND status = 'dead_letter';
```

---

## 12. Security & Best Practices

Task 7 security checklist:

- Keep `SESSION_EVENT_PEPPER` secret and long/random.
- Do not reuse `SESSION_EVENT_PEPPER` as JWT, OTP, or refresh-token pepper.
- Use `SESSION_LINK_MODE=disabled` only for simple local debugging.
- Use `SESSION_LINK_MODE=outbox` for integrated/staging/production.
- Monitor `pending`, `failed`, and `dead_letter` outbox rows.
- Never log access tokens, refresh tokens, OTPs, raw IPs, or raw device fingerprints.
- Keep event payloads minimal.
- Make Session Service consumer idempotent by `event_id`.
- Keep `AUTH_EVENTS_PUBLISH_ENDPOINT` on an internal network, not public internet.
- Add service auth, mTLS, or internal gateway protection for event ingress in production.

Operational best practices:

| Practice | Why |
|---|---|
| Short publish timeout | Auth process should not hang on slow event ingress |
| Exponential backoff | Prevents hammering a broken dependency |
| Dead-letter review | Failed events may contain important fraud/logout signals |
| Consumer idempotency | Retries can publish the same logical event more than once |
| Schema versioning | `version=1` allows future event changes safely |

---

## 13. Final Setup Checklist

For simple local run:

- [ ] Base Auth setup from `TaskImplementation/Auth Service/Dependency/task1_Dependency.md` completed.
- [ ] MySQL and Redis running.
- [ ] Auth migrations applied.
- [ ] JWT/OTP/refresh-token env variables sourced.
- [ ] `SESSION_LINK_MODE=disabled` set.
- [ ] `go run ./cmd/server` starts.

For Task 7 integrated outbox run:

- [ ] Migration `005_create_auth_outbox_events.up.sql` applied.
- [ ] `SESSION_LINK_MODE=outbox` set.
- [ ] `SESSION_EVENT_PEPPER` set to a secret value.
- [ ] `AUTH_EVENTS_TOPIC=auth.events` set.
- [ ] `AUTH_EVENTS_PUBLISH_ENDPOINT` set if publishing should happen.
- [ ] Login creates `AuthLoginSucceeded` outbox row.
- [ ] Logout creates `AuthLogoutSucceeded` outbox row.
- [ ] Refresh-token reuse creates `RefreshTokenReuseDetected` outbox row.
- [ ] Event ingress returns 2xx and rows become `published`.
- [ ] Pending/failed/dead-letter rows are monitored.

Task 7 essence: Auth Service remains owner of authentication, while Session Service receives safe session events through a durable outbox path. Login should stay reliable, events should be retryable, and sensitive auth secrets must never leave Auth Service.
