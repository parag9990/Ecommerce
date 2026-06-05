# Project Dependency & Setup Guide

## 1. Project Overview

| Variable | Value |
|---|---|
| `SERVICE_NAME` | Prompt variable from the assignment |
| `TASK_FILE_NAME` | Prompt variable from the assignment |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `{TASK_FILE_NAME without .md}_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

Current task ka focus zero-result tracking hai. Simple Hinglish me: jab buyer search kare aur Typesense se `total = 0` aaye, tab backend ek analytics event Session Service ko bhejta hai. Isse catalog team ko pata chalta hai ki users kya search kar rahe hain jo catalog me available nahi hai.

Important boundaries:

- Original `TASK_FILE_NAME` modify nahi kiya gaya.
- Business logic rewrite nahi kiya gaya.
- Base Go, Docker, Typesense, Redis, RabbitMQ, Product Service, and broad `.env` setup duplicate nahi kiya gaya.
- Previous dependency docs ko pehle follow karna hai; yahan sirf zero-result tracking ke task-specific setup, config, verification, and troubleshooting explain kiye gaye hain.

Current runnable flow in this repo snapshot:

```text
Client / API Gateway
  -> GET /api/v1/search with q plus analytics headers
  -> Typesense products search
  -> if total == 0
  -> build zero-result analytics event
  -> Redis SET NX dedupe key with TTL
  -> background worker POSTs event to Session Service
  -> search response stays non-blocking
```

Task-specific implemented files inspected:

| File | Why it matters |
|---|---|
| `backend/services/search-service/internal/usecase/search_products.go` | Calls zero-result tracker when `searchResult.Total == 0` |
| `backend/services/search-service/internal/usecase/zero_result_tracker.go` | Background queue, workers, dedupe, Session Service send behavior |
| `backend/services/search-service/internal/domain/zero_result.go` | Query normalization, skip rules, dedupe key, safe hashes |
| `backend/services/search-service/internal/repository/redis_zero_result_dedupe.go` | Redis `SET NX` dedupe implementation |
| `backend/services/search-service/internal/clients/http_session_event_client.go` | HTTP POST to Session Service event ingest endpoint |
| `backend/services/search-service/internal/requestctx/analytics_context.go` | Reads request/session analytics context |
| `backend/services/search-service/internal/transport/http/handler.go` | Accepts `X-Anonymous-ID`, `X-Session-ID`, `X-User-ID`, `X-Client-Path` headers |
| `backend/services/search-service/internal/config/config.go` | Loads zero-result and Session Service environment variables |
| `infra/compose/.env.local.example` | Contains local examples for zero-result and Session Service config |

## 2. Tech Stack

### Reused tech from previous dependency files

| Technology | Required for current task? | Reuse reference |
|---|---:|---|
| Go | Yes | `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Section `2. Go Dependency System` |
| Go modules | Yes | `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Section `2. Go Dependency System` |
| Standard Go `net/http` | Yes | `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Section `1. Project Tech Stack Analysis` |
| Typesense | Yes | `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`, Section `4.1 Typesense` |
| Docker Compose | Recommended for local infra | `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Section `7. Docker and DevOps Setup` |
| Redis | Yes | `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Section `3.2 Redis` |
| RabbitMQ | Only if `SEARCH_INDEXER_ENABLED=true` | `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`, Section `6. Redis / Queue / External Services` |
| Product Service | Needed for non-empty search hydration | `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md`, Sections `3` and `7` |

### Task-specific tech details

| Component | Required? | Beginner-friendly explanation |
|---|---:|---|
| Session Service | Yes if tracking enabled | Session Service analytics events store karta hai. Current task me `search` event with `result_count = 0` yahan send hota hai. |
| HTTP Session event client | Yes | Go `net/http` client Session Service ke `/api/v1/sessions/events` endpoint ko JSON POST karta hai. |
| Redis zero-result dedupe | Yes if tracking enabled | Same session + same query + same filters repeat hone par duplicate analytics event avoid karta hai. |
| In-memory background queue | Yes | Search response ko slow nahi karna. Event background worker send karta hai. |
| Analytics headers | Yes for tracking | `X-Anonymous-ID` and `X-Session-ID` missing honge to event skip ho jayega. |
| Structured logs | Yes | Raw query log nahi hoti; query/session hash and status logs hote hain. |

## 3. Required Software

No brand-new local software install is introduced by `TASK_FILE_NAME`.

Follow base setup first:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`2. Go Dependency System`
`3. Database and Storage Analysis`
`4. Environment Variables`
`7. Docker and DevOps Setup`
`8. Project Run Instructions`
```

For current task verification, make sure these are available:

| Software / Service | Required? | Why needed | Quick check |
|---|---:|---|---|
| Go | Yes | Run service and unit tests | `go version` |
| Typesense | Yes | Search result count yahin se aata hai | `curl http://localhost:8108/health` |
| Redis | Yes when tracking enabled | Zero-result dedupe key store karta hai | `docker exec ecommerce-redis redis-cli -a dev_redis_password ping` |
| Session Service | Yes for successful ingest | Receives analytics event | `curl http://localhost:8086/healthz` if service exposes it |
| Docker Compose | Recommended | Starts local Typesense, Redis, RabbitMQ infra | `docker compose version` |
| Product Service | Only for non-empty search tests | Hydrates product ids returned by Typesense | Check Product Service local docs |
| RabbitMQ | Only if indexer enabled | Existing product indexer consumer startup | `docker exec ecommerce-rabbitmq rabbitmq-diagnostics -q ping` |

Important local note: current Compose file starts Typesense, Redis, and RabbitMQ only. Session Service is not included in `infra/compose/docker-compose.local.yml`, so run it separately or expect `search.zero_result.ingest_failed` logs during isolated local testing.

## 4. Dependency Management

This is a Go project. Full Go module setup is already explained in:

```md
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`2. Go Dependency System`
```

Task-specific dependency audit:

| File / Dependency | Current finding | New install needed? |
|---|---|---:|
| `backend/services/search-service/go.mod` | Already has existing service dependencies | No |
| `github.com/typesense/typesense-go/v2` | Existing Typesense search dependency | No |
| `github.com/rabbitmq/amqp091-go` | Existing indexer dependency, unrelated to zero-result send path | No |
| Go standard library | `net/http`, `context`, `time`, `crypto/sha256`, `encoding/json`, `log/slog`, `sync` are used | No |
| Node/Python packages | Not used by current task | No |

No new `go get` command is required for `TASK_FILE_NAME`.

Task-specific verification commands:

```bash
cd backend/services/search-service
go test ./internal/domain ./internal/usecase ./internal/clients ./internal/repository ./internal/requestctx ./internal/transport/http ./internal/config
```

For all package tests:

```bash
cd backend/services/search-service
go test ./...
```

Beginner note: agar `go test ./...` me live Typesense integration test skip ho, ye normal ho sakta hai. Previous docs me integration-test mode already explained hai.

## 5. Database Setup

No SQL database, MongoDB collection, or SQL migration is introduced by `TASK_FILE_NAME`.

| Storage / DB | Used? | Purpose | Setup action |
|---|---:|---|---|
| Typesense `products` collection | Yes | Search result `total` determine karta hai | Reuse existing Typesense setup |
| Redis | Yes | Zero-result dedupe key stores with TTL | Reuse existing Redis setup, verify password/address |
| Session Service storage | Indirect | Analytics event ka actual persistence Session Service owns karta hai | Run/configure Session Service separately |
| MySQL/PostgreSQL/MongoDB | No direct use | Current task me service SQL write nahi karta | No setup |

Typesense setup is already explained in:

```md
`TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`

Section:
`4.1 Typesense`
```

Redis setup is already explained in:

```md
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`3.2 Redis`
```

### Task-specific Redis dedupe storage

Current task adds/uses Redis keys with this prefix:

```text
search:zero_result:v1:
```

What gets stored:

| Redis detail | Value |
|---|---|
| Operation | `SET NX` |
| Value | `1` |
| Default TTL | `30m` |
| Config variable | `SEARCH_ZERO_RESULT_DEDUPE_TTL` |
| Key input | Session id + normalized query + canonical filters |
| Privacy behavior | Key uses SHA-256 hash; raw query/session id is not present in Redis key |

Beginner note: Dedupe ka matlab hai same session me same query repeatedly refresh karne par analytics spam avoid karna. Isse catalog team ko cleaner signal milta hai.

### No migration required

No SQL migration is needed.

Typesense collection ensure happens during service startup, and zero-result events are sent to Session Service. `{SERVICE_NAME}` khud zero-result event ko SQL table me save nahi karta.

## 6. Redis / Queue / External Services

### Session Service event ingest

Current task depends on Session Service event ingest:

| Item | Value |
|---|---|
| Method | `POST` |
| Default base URL | `http://localhost:8086` |
| Default path | `/api/v1/sessions/events` |
| Config variables | `SESSION_SERVICE_URL`, `SESSION_SERVICE_INGEST_PATH`, `SESSION_SERVICE_TIMEOUT_MS` |
| API catalog reference | `api/master-api.json`, `SessionEventInput` |
| Event type | `search` |
| Success status accepted by client | Any `2xx` response |

Example event body sent by the current HTTP client:

```json
{
  "event_type": "search",
  "anonymous_id": "anon_123",
  "session_id": "sess_123",
  "user_id": "user_123",
  "occurred_at": "2026-05-24T10:30:00Z",
  "path": "/search",
  "properties": {
    "query": "waterproof laptop bag",
    "normalized_query": "waterproof laptop bag",
    "result_count": 0,
    "zero_result": true,
    "filters": {
      "brand": "Acme"
    },
    "sort": "relevance",
    "page": 1,
    "page_size": 20,
    "source": "search-service",
    "request_id": "req_123"
  }
}
```

Privacy note: raw query is intentionally sent to analytics because catalog teams need to understand missing demand. Treat Session Service data as sensitive analytics data. Logs should use hashes, and access to analytics storage should be restricted.

### Redis dedupe service

Redis is required for successful tracking because the tracker calls dedupe before sending the Session Service event.

Verify Redis:

```bash
docker exec ecommerce-redis redis-cli -a dev_redis_password ping
```

Inspect zero-result dedupe keys after a tracked request:

```bash
docker exec ecommerce-redis redis-cli -a dev_redis_password --scan --pattern 'search:zero_result:v1:*'
```

If a key exists and you repeat the same query/session/filter within TTL, the second event is skipped as duplicate. Ye expected behavior hai.

### RabbitMQ / Kafka / queues

`TASK_FILE_NAME` does not add Kafka, RabbitMQ, NATS, or RabbitMQ queue changes.

RabbitMQ remains only for existing product indexer flow:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`

Section:
`6. Redis / Queue / External Services`
```

The zero-result tracker uses an in-memory Go channel, not RabbitMQ. Isliye restart/crash ke time pending analytics events lose ho sakte hain; see missing/misconfigured section.

## 7. Environment Variables

Full `.env` setup is already documented in:

```md
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`4. Environment Variables`
```

Current task does not introduce brand-new variable names beyond earlier dependency docs, but these values are task-critical. Verify this subset in:

```text
backend/services/search-service/.env
```

```env
# Zero-result analytics behavior
SEARCH_ZERO_RESULT_TRACKING_ENABLED=true
SEARCH_ZERO_RESULT_DEDUPE_TTL=30m
SEARCH_ZERO_RESULT_SEND_TIMEOUT_MS=150
SEARCH_ZERO_RESULT_QUEUE_SIZE=1024
SEARCH_ZERO_RESULT_WORKERS=2

# Session Service analytics ingest
SESSION_SERVICE_URL=http://localhost:8086
SESSION_SERVICE_INGEST_PATH=/api/v1/sessions/events
SESSION_SERVICE_TIMEOUT_MS=150

# Existing Redis dependency used for dedupe
SEARCH_REDIS_ADDR=localhost:6379
SEARCH_REDIS_PASSWORD=dev_redis_password
SEARCH_REDIS_DB=0
```

### Variable explanation

| Variable | Required? | Purpose | Example | Security / mistake note |
|---|---:|---|---|---|
| `SEARCH_ZERO_RESULT_TRACKING_ENABLED` | Optional | Enables/disables zero-result event tracking | `true` | Set `false` only for isolated local testing |
| `SEARCH_ZERO_RESULT_DEDUPE_TTL` | Optional | Duplicate suppression window | `30m` | Too high may hide useful repeated demand |
| `SEARCH_ZERO_RESULT_SEND_TIMEOUT_MS` | Optional | Timeout for Redis dedupe + Session send worker | `150` | Too low can cause `session_ingest_error` |
| `SEARCH_ZERO_RESULT_QUEUE_SIZE` | Optional | In-memory event queue capacity | `1024` | Too small can cause `queue_full`; too high uses memory |
| `SEARCH_ZERO_RESULT_WORKERS` | Optional | Background workers sending events | `2` | Increase carefully after measuring Session Service capacity |
| `SESSION_SERVICE_URL` | Required if tracking enabled | Session Service base URL | `http://localhost:8086` | Internal URL only; must be absolute `http` or `https` |
| `SESSION_SERVICE_INGEST_PATH` | Required if tracking enabled | Event ingest path | `/api/v1/sessions/events` | Wrong path creates ingest failure |
| `SESSION_SERVICE_TIMEOUT_MS` | Optional | HTTP client timeout for Session Service | `150` | Keep bounded so analytics cannot slow service shutdown |
| `SEARCH_REDIS_ADDR` | Yes for tracking | Redis host and port for dedupe | `localhost:6379` | Use `redis:6379` only inside Docker network |
| `SEARCH_REDIS_PASSWORD` | Required if Redis auth enabled | Redis password | `dev_redis_password` | Secret. Must match Compose `REDIS_PASSWORD` |
| `SEARCH_REDIS_DB` | Optional | Redis logical DB | `0` | Avoid sharing prod DB with unrelated workloads |

### How env is loaded

The Go service reads environment variables from the process through `config.Load()`. It does not automatically parse `.env` files.

For local shell runs:

```bash
cd backend/services/search-service
set -a
. ./.env
set +a
go run ./cmd/server
```

Host vs Docker examples:

| Run location | `SESSION_SERVICE_URL` | `SEARCH_REDIS_ADDR` |
|---|---|---|
| Host `go run` | `http://localhost:8086` | `localhost:6379` |
| Inside Docker Compose network | `http://session-service:8086` | `redis:6379` |
| Kubernetes | Cluster service DNS | Redis service DNS |

Common mistake: `infra/compose/.env.local.example` uses Docker service DNS values like `typesense` and `session-service`. If Go runs on the host, use `localhost` values in `backend/services/search-service/.env`.

## 8. Docker Setup

No new Docker container, volume, network, or image is introduced by `TASK_FILE_NAME`.

Existing local infra setup is already covered in:

```md
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`7. Docker and DevOps Setup`
`8. Project Run Instructions`
```

Current Compose services reused:

| Container | Needed? | Why |
|---|---:|---|
| `ecommerce-typesense` | Yes | Search result count comes from Typesense |
| `ecommerce-redis` | Yes for tracking | Stores zero-result dedupe keys |
| `ecommerce-rabbitmq` | Only if indexer enabled | Existing product event indexer dependency |

Start existing local infra:

```bash
cd infra/compose
cp .env.local.example .env.local
docker compose --env-file .env.local -f docker-compose.local.yml up -d typesense redis rabbitmq
```

Session Service note: `docker-compose.local.yml` does not currently define a Session Service container. For successful ingest testing, run Session Service separately, add it to Compose later, or point `SESSION_SERVICE_URL` to a reachable local/staging instance.

### Ports and networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Backend HTTP API | `8085` | Serves `/api/v1/search` | Reused |
| Typesense API | `8108` | Search engine API | Reused |
| Redis | `6379` | Zero-result dedupe keys | Reused, task-critical |
| Session Service | `8086` | Analytics event ingest | Task-specific dependency, not Compose-provided |
| Product Service | `8082` | Product hydration for non-empty hits | Reused |
| RabbitMQ AMQP | `5672` | Existing product indexer queue | Reused when indexer enabled |
| RabbitMQ Management | `15672` | Queue debugging UI | Reused when indexer enabled |

No Kubernetes or Docker health check changes are introduced by current task.

## 9. Local Development Setup

### Step 1: Read previous dependency documentation first

Read these before current guide:

```md
1. `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`
2. `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`
3. `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` if product indexer is enabled
4. `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` for Search API behavior and analytics headers
5. `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md` if autocomplete/popular queries are also being tested
6. `TaskImplementation/{SERVICE_NAME}/task6_Dependency.md` if admin synonym changes affect zero-result testing
```

### Step 2: Go to project directory

```bash
cd backend/services/search-service
```

### Step 3: Install only new dependencies if any

No new dependency install is required for current task.

For a fresh clone:

```bash
go mod download
```

### Step 4: Setup only new databases/services if any

No new database is added. Reuse existing Typesense and Redis.

Start infra:

```bash
cd ../../..
cd infra/compose
docker compose --env-file .env.local -f docker-compose.local.yml up -d typesense redis rabbitmq
```

If Session Service is available separately, start it before verification. If it is not available, the search request still returns normally, but event ingest logs will show failure.

### Step 5: Add only new or changed environment variables

Verify Section `7. Environment Variables`.

For full zero-result tracking:

```env
SEARCH_ZERO_RESULT_TRACKING_ENABLED=true
SESSION_SERVICE_URL=http://localhost:8086
SESSION_SERVICE_INGEST_PATH=/api/v1/sessions/events
```

For isolated Search API testing without Session Service:

```env
SEARCH_ZERO_RESULT_TRACKING_ENABLED=false
```

Beginner warning: disabling tracking is useful for local boot, but then current task behavior cannot be fully verified.

### Step 6: Run migrations if needed

No SQL migration is needed.

The service ensures Typesense collections on startup. Redis dedupe keys are created dynamically when zero-result searches happen.

### Step 7: Start backend service

Full mode:

```bash
cd backend/services/search-service
set -a
. ./.env
set +a
go run ./cmd/server
```

API-only mode without RabbitMQ consumer:

```bash
cd backend/services/search-service
set -a
. ./.env
set +a
SEARCH_INDEXER_ENABLED=false go run ./cmd/server
```

### Step 8: Verify API related to `TASK_FILE_NAME`

Health:

```bash
curl "http://localhost:8085/healthz"
```

Trigger a zero-result search with required analytics headers:

```bash
curl -i "http://localhost:8085/api/v1/search?q=waterproof+laptop+bag&page=1&page_size=20" \
  -H "X-Request-ID: req_zero_1" \
  -H "X-Anonymous-ID: anon_123" \
  -H "X-Session-ID: sess_123" \
  -H "X-User-ID: user_123" \
  -H "X-Client-Path: /search"
```

Expected Search API response shape:

```json
{
  "products": [],
  "facets": {},
  "total": 0
}
```

Exact `facets` shape can include empty facet arrays depending on schema policy, but `total` should be `0`.

Check logs:

| Log | Meaning |
|---|---|
| `search.zero_result.tracked` | Event was deduped and sent successfully |
| `search.zero_result.ingest_failed` | Session Service was unreachable or returned non-2xx |
| `search.zero_result.duplicate_skipped` | Same session/query/filter already tracked inside TTL |
| `search.zero_result.skipped` with `missing_session_context` | Required analytics headers were missing |
| `search.zero_result.skipped` with `browse_query` | Query was empty or `*`, so it is not a real zero-result search |

Check Redis dedupe key:

```bash
docker exec ecommerce-redis redis-cli -a dev_redis_password --scan --pattern 'search:zero_result:v1:*'
```

## 10. Running the Project

### Task-specific API behavior

| Item | Value |
|---|---|
| Client route | `GET /api/v1/search` |
| Trigger condition | Typesense search result `total == 0` |
| Skipped query | Empty/browse query normalized to `*` |
| Required headers for event | `X-Anonymous-ID`, `X-Session-ID` |
| Optional headers | `X-Request-ID`, `X-Correlation-ID`, `X-User-ID`, `X-Client-Path` |
| Event destination | `POST {SESSION_SERVICE_URL}{SESSION_SERVICE_INGEST_PATH}` |
| Event type | `search` |
| Result count property | `result_count: 0` |
| Dedupe store | Redis |
| Dedupe key prefix | `search:zero_result:v1:` |
| Default dedupe TTL | `30m` |
| Delivery style | Background, non-blocking best effort |

### Tracking rules

| Scenario | Event sent? | Reason |
|---|---:|---|
| `q=waterproof laptop bag`, `total=0`, headers present | Yes | Real zero-result search |
| Same session + same query + same filters repeated inside TTL | No | Duplicate skipped |
| `q=*` or empty query | No | Browse/listing query, not meaningful analytics |
| Missing `X-Anonymous-ID` | No | Session analytics context incomplete |
| Missing `X-Session-ID` | No | Session analytics context incomplete |
| Typesense fails before result count | No | Search error, not a zero-result result |
| Non-empty result | No | Current task tracks only zero-result queries |
| Autocomplete endpoint returns no suggestions | No | Current implementation tracks Search API zero results only |

### Useful manual checks

Check that missing headers skip tracking:

```bash
curl -i "http://localhost:8085/api/v1/search?q=missing+catalog+item"
```

Expected: Search response can still be `200`, but logs should show skip reason `missing_session_context`.

Check duplicate behavior:

```bash
curl "http://localhost:8085/api/v1/search?q=rare+query" \
  -H "X-Anonymous-ID: anon_dup" \
  -H "X-Session-ID: sess_dup"

curl "http://localhost:8085/api/v1/search?q=rare+query" \
  -H "X-Anonymous-ID: anon_dup" \
  -H "X-Session-ID: sess_dup"
```

Expected: first request tracks or attempts ingest; second request logs duplicate skip while Redis TTL is active.

## 11. Common Errors & Fixes

Generic Docker, Go module, Typesense, Redis, and RabbitMQ troubleshooting is already documented in:

```md
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`9. Common Errors and Fixes`
```

Task-specific errors:

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| No zero-result event appears | Missing `X-Anonymous-ID` or `X-Session-ID` | Send both headers from frontend/gateway | Add gateway contract tests for analytics headers |
| Log shows `missing_session_context` | Session metadata missing or blank | Confirm headers reach service after gateway/proxy | Do not let proxies strip custom headers |
| Log shows `browse_query` | Query is empty or normalized to `*` | Use real search text for analytics test | Keep browse/listing pages separate from search analytics |
| Service fails with `SESSION_SERVICE_URL must be a valid absolute URL` | Invalid URL while tracking enabled | Use `http://localhost:8086` or disable tracking locally | Keep `.env` examples valid |
| Log shows `search.zero_result.ingest_failed` | Session Service down, wrong URL/path, timeout, or non-2xx response | Start Session Service, fix URL/path, increase timeout carefully | Add Session Service health/readiness checks |
| Log shows `search.zero_result.dedupe_failed` | Redis down, wrong password, wrong host, timeout | Verify Redis with `redis-cli`, fix `SEARCH_REDIS_*` | Keep Redis env aligned with Compose |
| Second identical query does not send event | Redis dedupe TTL still active | Wait for TTL or use different session/query/filter | Understand dedupe before debugging |
| Log shows `queue_full` | Session Service slow/down and event queue filled | Fix Session Service, increase workers/queue only after measuring | Monitor queue-full logs |
| Zero-result tracking disabled | `SEARCH_ZERO_RESULT_TRACKING_ENABLED=false` | Set it to `true` for task verification | Use disabled mode only for isolated local boot |
| Search returns non-empty results, no event | Event only triggers on `total == 0` | Use query that has no indexed matches | Seed/index data intentionally for tests |
| Redis key exists but Session Service did not receive event | Dedupe key is written before event send, then ingest failed | Delete key or wait TTL before retrying after fixing Session Service | Consider retry-after-send design improvement later |

## 12. Security & Best Practices

Task-specific best practices:

- Treat search query analytics as sensitive data. Query text can contain personal information typed by users.
- Keep Session Service endpoint internal-only. Browser clients should not call internal analytics ingest directly unless product/session architecture explicitly allows it.
- Keep `TYPESENSE_API_KEY`, `SEARCH_REDIS_PASSWORD`, and real Session Service credentials out of git.
- Do not log raw search query in backend logs. Current code logs query hash and query length; keep that pattern.
- Gateway should generate or validate `anonymous_id`, `session_id`, and `user_id`. `{SERVICE_NAME}` should not trust client-spoofed identity in production without gateway validation.
- Keep zero-result delivery non-blocking. Analytics failure should not break buyer search UX.
- Tune `SEARCH_ZERO_RESULT_DEDUPE_TTL` based on analytics needs. Too short creates spam; too long can hide repeated demand.
- Alert on `search.zero_result.ingest_failed`, `search.zero_result.dedupe_failed`, and `search.zero_result.queue_full`.
- Use zero-result insights together with synonyms and catalog data. Not every zero-result query needs a synonym; some need new products/categories.

Existing broader security/config notes are already documented in:

```md
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`10. Security and Configuration Audit`
```

## 13. Missing or Misconfigured Things

| Finding | Impact | Suggested fix |
|---|---|---|
| Session Service is not present in local Compose | Beginners may see ingest failure even when Search API works | Add Session Service to Compose when available, or document separate startup |
| Current event sink is HTTP, not generated gRPC | `api/master-api.json` mentions `SessionService.IngestEvent`, but runnable code uses HTTP JSON | Add proto/gRPC client/server wiring when internal RPC layer is implemented |
| No durable retry queue for analytics events | In-memory queue loses pending events on crash/restart | Add durable async queue or retry policy if analytics completeness becomes critical |
| Dedupe key is written before Session ingest succeeds | A failed first send suppresses retries until TTL expires | Consider marking dedupe after successful ingest or adding retry state |
| Metrics recorder is currently no-op unless wired | Operations may rely on logs only | Wire Prometheus/OpenTelemetry metrics for tracked/skipped/failed/duplicate outcomes |
| `/healthz` is shallow | It does not prove Typesense, Redis, Product Service, or Session Service are reachable | Add readiness endpoint with dependency checks |
| Session identity comes from headers | Direct exposure could allow spoofed analytics identity | Put service behind API Gateway and strip untrusted client headers |
| Raw query is sent to analytics payload | Sensitive user input may be stored in Session Service | Apply retention, access control, PII review, and deletion policy |
| Compose env uses Docker DNS names | Host `go run` can fail if same values are sourced directly | Use host-specific `.env` with `localhost` values |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `2. Go Dependency System` | Same Go module setup and commands already documented |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3.2 Redis` | Redis installation, password, and local verification are unchanged |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Environment Variables` | Full `.env` example already includes zero-result and Session Service variables |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `5.5 Session Service` | Base Session Service purpose and disable flag already documented |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. Docker and DevOps Setup` | Same Compose stack, containers, volumes, and network are reused |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `9. Common Errors and Fixes` | Generic setup and infra errors already covered |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `10. Security and Configuration Audit` | Broad secrets, health checks, and deployment gaps already documented |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `4.1 Typesense` | Same Typesense local install, Docker, health check, and credentials reused |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `6. Redis / Queue / External Services` | RabbitMQ/indexer setup unchanged and only needed if indexer is enabled |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | Search API, analytics headers, Product Service hydration | Zero-result tracking is triggered from the existing Search API path |
| `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md` | Popular query/autocomplete boundary | Current task does not track autocomplete no-suggestion events |
| `TaskImplementation/{SERVICE_NAME}/task6_Dependency.md` | Synonym management boundary | Zero-result insights can inform synonyms, but synonym setup is unchanged |

## 15. Final Checklist

- [ ] Previous dependency documentation checked before using this file
- [ ] No duplicate Go/Docker/Typesense/Redis/RabbitMQ installation docs copied
- [ ] Typesense is running on port `8108`
- [ ] Redis is running on port `6379`
- [ ] `SEARCH_REDIS_PASSWORD` matches Redis container password
- [ ] `SEARCH_ZERO_RESULT_TRACKING_ENABLED` is set intentionally
- [ ] `SESSION_SERVICE_URL` is a valid internal absolute URL when tracking is enabled
- [ ] `SESSION_SERVICE_INGEST_PATH` is `/api/v1/sessions/events` unless Session Service contract changes
- [ ] Session Service is running or ingest failure is expected during isolated local testing
- [ ] `X-Anonymous-ID` and `X-Session-ID` are sent with search requests
- [ ] Zero-result search returns `total = 0`
- [ ] Redis dedupe key appears with prefix `search:zero_result:v1:`
- [ ] Logs checked for tracked, duplicate, skipped, and failed outcomes
- [ ] No SQL migration expected
- [ ] No new Docker container expected from current task
- [ ] PII/security review done for raw query analytics payload
- [ ] No duplicate setup documentation added
