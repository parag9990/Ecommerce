# Project Dependency & Setup Guide

## Variables Used

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Recommendation Service` |
| `TASK_FILE_NAME` | `task7.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task7_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}` |
| Backend code path | `backend/services/recommendation-service` |
| Proto contract path | `proto/ecommerce/recommendation/v1/recommendation.proto` |
| Generated Go proto module | `backend/proto-gen/go` |

Beginner note: Ye file sirf dependency, setup, environment, DevOps, and troubleshooting guide hai. Business logic ya original implementation file modify nahi kiya gaya.

---

## 1. Project Overview

`TASK_FILE_NAME` ka focus hai: `${SERVICE_NAME}` ka typed gRPC serving endpoint `GetRecommendations(user_id, context)`.

Simple Hinglish:

- Task 1 to Task 6 ne recommendation types, MongoDB/Redis storage, events, feature store, rule-based ranking, and personalization ready kiya.
- Task 7 ka kaam hai same recommendation result ko internal services ke liye gRPC API se expose karna.
- gRPC endpoint product details khud hydrate nahi karta. Ye ranked `product_id` references return karta hai; Gateway/Product Service product cards fetch kar sakte hain.
- Current backend code me gRPC server already wired hai and default port `9088` par start hota hai.

Read these previous dependency files first:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`1. Project Tech Stack Analysis`
`2. Go Dependency System`
`4. Environment Variables`
`6. Ports and Networking`
`7. Docker and DevOps Setup`
`8. Project Run Instructions`
`9. Common Errors and Fixes`
```

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task2_Dependency.md`

Sections:
`5. Database Setup`
`6. Redis / Queue / External Services`
`7. Environment Variables`
`8. Docker Setup`
```

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task5_Dependency.md`

Sections:
`5. Database Setup`
`7. Environment Variables`
`9. Local Development Setup`
```

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task6_Dependency.md`

Sections:
`5. Database Setup`
`7. Environment Variables`
`9. Local Development Setup`
```

`TASK_FILE_NAME` new setup scope:

- gRPC service contract: `ecommerce.recommendation.v1.RecommendationService`
- RPC method: `GetRecommendations`
- Proto source file: `proto/ecommerce/recommendation/v1/recommendation.proto`
- Generated Go stubs: `backend/proto-gen/go/ecommerce/recommendation/v1`
- gRPC transport implementation:
  - `internal/transport/grpc/server.go`
  - `internal/transport/grpc/handler.go`
  - `internal/transport/grpc/mapper.go`
  - `internal/transport/grpc/errors.go`
- gRPC health service registration.
- gRPC serving metrics and structured logs.

No new database engine, cache engine, queue engine, Docker container, third-party SaaS, or migration is introduced by `TASK_FILE_NAME`.

---

## 2. Tech Stack

Most base technologies are already explained in previous dependency files. Do not repeat their full installation steps here.

| Technology | New or Reused | Required? | Beginner Explanation | Current Task Use |
|---|---|---:|---|---|
| Go | Reused | Yes | Go backend language hai. Service binary Go me run hoti hai. | gRPC server, mapper, handler, and tests. |
| Go modules | Reused | Yes | `go.mod` and `go.sum` dependencies ko version ke saath manage karte hain. | Service module generated proto module ko local `replace` se use karta hai. |
| gRPC | Reused dependency, Task 7 primary feature | Yes | gRPC ek fast internal API protocol hai. Simple English: service-to-service typed function call. | `GetRecommendations` RPC expose karne ke liye. |
| Protocol Buffers | Reused dependency, Task 7 primary contract | Yes | Protobuf strongly typed API contract format hai. | Request/response messages and enums define karta hai. |
| Generated Go proto stubs | Reused module path, Task 7 required | Yes | `.proto` se generated Go files bante hain. Inhe manually edit nahi karna hota. | Server interface and request/response structs. |
| gRPC health service | New operational behavior for this endpoint | Recommended | Health RPC service status batata hai. | `grpc.health.v1.Health/Check` se service readiness verify karne ke liye. |
| Prometheus client | Reused, new gRPC metrics | Recommended | Metrics expose karne ki Go library. | gRPC request count, duration, in-flight requests, serve source metrics. |
| MongoDB | Reused | Required for real data-backed serving | Document database. JSON-like collections me data store hota hai. | Stored sets, personalized profiles, product features. |
| Redis | Reused | Optional but recommended | Fast in-memory cache hai. | Hot recommendation reads and TTL-based cache. |
| Kafka | Reused | Optional for live behavior refresh | Event streaming platform hai. | User behavior events se feature data fresh rakhna; direct gRPC call ke liye mandatory nahi. |
| `grpcurl` | Reused optional tool | Optional | Command-line gRPC client. | Local/manual `GetRecommendations` verification. |

Not introduced by `TASK_FILE_NAME`:

| Technology | Status |
|---|---|
| MySQL / PostgreSQL / SQLite | Not used |
| RabbitMQ / NATS | Not used |
| Elasticsearch / vector DB / ML model server | Not used |
| AWS S3 / Firebase / SMTP / Stripe / Twilio / OAuth | Not used |
| Kubernetes / Nginx | Not present as task-specific setup |

---

## 3. Required Software

Base software installation is already documented in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Sections:
`2. Go Dependency System`
`3. Databases`
`5. External Services`
`7. Docker and DevOps Setup`
```

For `TASK_FILE_NAME`, confirm these are available:

| Software | Needed For | New or Reused |
|---|---|---|
| Go `1.26.3` or compatible newer | Build/test/run service | Reused |
| Generated proto module under `backend/proto-gen/go` | Compile gRPC transport | Reused, required |
| `protoc-gen-go` and `protoc-gen-go-grpc` | Regenerate stubs only when `.proto` changes | Reused optional tooling |
| Buf | Proto lint/generation workflow | Reused optional tooling |
| `grpcurl` | Manual gRPC calls and health check | Reused optional tooling |
| MongoDB and Redis | Real recommendation output | Reused |
| Kafka | Live event-driven freshness only | Reused |
| `curl` | HTTP health/metrics checks | Reused |

Beginner decision:

- Sirf gRPC server start/health verify karna hai: Go plus existing generated proto files enough hain.
- Real recommendations chahiye: Task 2, Task 5, and Task 6 setup follow karo.
- Live behavior se recommendations fresh karni hain: Task 3 and Task 4 setup bhi follow karo.

---

## 4. Dependency Management

This is still the same Go module project. Go modules, `go.mod`, `go.sum`, `go mod download`, and common Go module errors are already explained in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `2. Go Dependency System`
```

`TASK_FILE_NAME` dependency status:

| Dependency | Version In `go.mod` | Used For | New Install Needed? |
|---|---:|---|---:|
| `github.com/example/ecommerce-platform/backend/proto-gen/go` | `v0.0.0` via local `replace` | Generated `RecommendationService` Go interface and protobuf messages | No |
| `google.golang.org/grpc` | `v1.81.1` | gRPC server, status codes, health server, tests | No |
| `google.golang.org/protobuf` | `v1.36.11` | Protobuf runtime and timestamp mapping | No |
| `github.com/prometheus/client_golang` | `v1.23.2` | gRPC and serving metrics | No |
| `go.mongodb.org/mongo-driver/v2` | `v2.6.0` | Real stored/personalized recommendation reads | No |
| `github.com/redis/go-redis/v9` | `v9.19.0` | Cache reads/writes | No |

No new `go get` command is needed for `TASK_FILE_NAME`.

Recommended verification:

```bash
cd backend/services/recommendation-service
go test ./internal/transport/grpc ./internal/usecase
```

Full service verification:

```bash
cd backend/services/recommendation-service
go test ./...
```

If `.proto` changes in future, regenerate stubs from repo root:

```bash
cd proto
buf lint
buf generate
```

Important: generated `.pb.go` files should not be manually edited. Contract source is:

```text
proto/ecommerce/recommendation/v1/recommendation.proto
```

---

## 5. Database Setup

No new database setup or migration is introduced by `TASK_FILE_NAME`.

MongoDB installation, Docker setup, connection string format, credentials placement, and migrations are already explained in:

| Previous File | Section / Topic |
|---|---|
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `3. Databases` -> `MongoDB` |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` |
| `TaskImplementation/${SERVICE_NAME}/task4_Dependency.md` | `5. Database Setup` -> feature-store and replica set requirement |
| `TaskImplementation/${SERVICE_NAME}/task5_Dependency.md` | `5. Database Setup` -> rule-based ranking indexes |
| `TaskImplementation/${SERVICE_NAME}/task6_Dependency.md` | `5. Database Setup` -> personalized ranking index and input data |

Task 7 reads whatever Task 5/6 serving use cases can produce:

| Data Source | Why It Matters For gRPC |
|---|---|
| `recommendation_sets` | Stored/generated top lists returned on cache miss or fallback. |
| `product_features` | Rule-based and personalized candidate source. |
| `user_feature_profiles` | Personalized profile source. |
| `user_product_counters` | Personalized direct affinity source. |
| Redis cache | Fast response path before durable data fallback. |

Current migration status for Task 7:

| Migration | Needed For Task 7? | Notes |
|---|---:|---|
| `001_create_recommendation_storage.up.js` | Yes for stored set reads | Reused from Task 2 |
| `002_add_event_ingestion_indexes.up.js` | Only for live event ingestion | Reused from Task 3 |
| `003_create_feature_store_schema.up.js` | Yes for feature-based ranking/personalization | Reused from Task 4 |
| `004_add_rule_based_ranking_indexes.up.js` | Yes for Task 5 generated rankings | Reused from Task 5 |
| `005_add_personalized_ranking_indexes.up.js` | Yes for Task 6 personalization | Reused from Task 6 |
| `006_add_ab_test_assignment_indexes.up.js` | Optional; Task 8 A/B hooks | Not Task 7-specific |

Credentials placement remains:

```text
backend/services/recommendation-service/.env
```

Do not commit real MongoDB passwords. Local dummy values are okay only for local development.

---

## 6. Redis / Queue / External Services

No new Redis, Kafka, RabbitMQ, or third-party external service is introduced by `TASK_FILE_NAME`.

Reuse previous setup:

| Service | Previous File | Section / Topic |
|---|---|---|
| Redis | `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `5. External Services` -> `Redis` |
| Redis cache details | `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | `6. Redis / Queue / External Services` |
| Kafka | `TaskImplementation/${SERVICE_NAME}/task3_Dependency.md` | `6. Redis / Queue / External Services` |
| Docker infra examples | `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `7. Docker and DevOps Setup` |

Task 7-specific gRPC external surface:

| Surface | Purpose | Mandatory? | Health Check |
|---|---|---:|---|
| `ecommerce.recommendation.v1.RecommendationService/GetRecommendations` | Internal recommendation serving RPC | Yes for Task 7 verification | Use `grpcurl` with proto file |
| `grpc.health.v1.Health/Check` | gRPC service health | Recommended | `grpcurl -plaintext localhost:9088 grpc.health.v1.Health/Check` |
| `/metrics` over HTTP | gRPC metrics scrape | Recommended | `curl http://localhost:8088/metrics` |

Metadata supported by the gRPC server:

| Metadata Key | Purpose |
|---|---|
| `x-request-id` / `request-id` / `traceparent` | Request tracing in logs. |
| `x-session-id` / `session-id` | Session identity passed into recommendation resolution. |

Simple Hinglish: gRPC request me IDs body me jaate hain, but trace/session metadata headers ki tarah gRPC metadata me ja sakti hai. Logs me debugging easy hoti hai.

---

## 7. Environment Variables

No new environment variable is introduced by `TASK_FILE_NAME`. The gRPC variables were already documented earlier.

Refer:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `4. Environment Variables`
```

Task 7 relevant existing variables:

| Variable | Default | Required? | Task 7 Purpose | Security / Config Note |
|---|---:|---:|---|---|
| `RECOMMENDATION_GRPC_ADDR` | `:9088` | Optional | gRPC listen address. | Keep private/internal in production. |
| `RECOMMENDATION_GRPC_MAX_RECV_BYTES` | `65536` | Optional | Max request message size. | Protects service from oversized payloads. |
| `RECOMMENDATION_GRPC_MAX_SEND_BYTES` | `262144` | Optional | Max response message size. | Increase carefully if large result lists are required. |
| `RECOMMENDATION_GRPC_DEFAULT_DEADLINE` | `500ms` | Optional | Server adds this deadline when client does not send one. | Avoid infinite waits; clients should still set deadlines. |
| `RECOMMENDATION_DEFAULT_LIMIT` | `12` | Optional | Default number of returned items. | Used when request limit is `0`. |
| `RECOMMENDATION_MAX_LIMIT` | `100` | Optional | Upper bound for returned items. | Prevents expensive responses. |
| `RECOMMENDATION_MAX_IDENTIFIER_LENGTH` | `128` | Optional | Guards user/product/category/seller/cache identifiers. | Prevents abusive IDs and cache key issues. |

Task 7 `.env` incremental snippet:

```env
# Reused gRPC serving config from earlier dependency docs.
# Add only if you want to override defaults.
RECOMMENDATION_GRPC_ADDR=:9088
RECOMMENDATION_GRPC_MAX_RECV_BYTES=65536
RECOMMENDATION_GRPC_MAX_SEND_BYTES=262144
RECOMMENDATION_GRPC_DEFAULT_DEADLINE=500ms
```

Important `.env` loading behavior is unchanged:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `4. Environment Variables` -> `Important .env loading behavior`
```

The Go code uses `os.Getenv`; it does not auto-load `.env`. Run:

```bash
cd backend/services/recommendation-service
set -a
source .env
set +a
go run ./cmd/server
```

---

## 8. Docker Setup

No new Docker container, volume, network, Dockerfile, or docker-compose service is introduced by `TASK_FILE_NAME`.

Reuse previous Docker documentation:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `7. Docker and DevOps Setup`
```

Current repo status remains:

| Item | Status |
|---|---|
| Service Dockerfile | Not detected |
| Root docker-compose for this service | Not detected |
| Dependency containers | MongoDB, Redis, Kafka examples already documented |
| New Task 7 container | Not needed |
| New Task 7 volume/network | Not needed |

If you containerize the service later, expose the gRPC port internally:

```yaml
ports:
  - "8088:8088"
  - "9088:9088"
```

Production note: gRPC `9088` should normally be reachable only by Gateway/internal services, not public internet.

---

## 9. Local Development Setup

### Step 1: Read previous dependency documentation

Follow these first:

| Order | File | Why |
|---:|---|---|
| 1 | `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | Go, env loading, ports, gRPC basics, Docker examples. |
| 2 | `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | MongoDB/Redis storage and cache setup. |
| 3 | `TaskImplementation/${SERVICE_NAME}/task5_Dependency.md` | Rule-based generated recommendation data. |
| 4 | `TaskImplementation/${SERVICE_NAME}/task6_Dependency.md` | Personalized recommendation data and fallback behavior. |

### Step 2: Go to project directory

```bash
cd backend/services/recommendation-service
```

### Step 3: Install only new dependencies if any

No new Go dependency is needed for `TASK_FILE_NAME`.

Verify current dependencies:

```bash
go mod download
go test ./internal/transport/grpc
```

### Step 4: Setup only new databases/services if any

No new database/service setup is needed for `TASK_FILE_NAME`.

For real data-backed output, ensure previous MongoDB/Redis setup is already done:

```md
`TaskImplementation/${SERVICE_NAME}/task2_Dependency.md`
Section: `5. Database Setup`
```

### Step 5: Add only new or changed environment variables

No new variables. Use defaults or the small gRPC override snippet from section `7. Environment Variables`.

### Step 6: Run migrations if needed

No Task 7 migration exists.

If your local database is fresh, run previous migrations in order as documented in:

```md
`TaskImplementation/${SERVICE_NAME}/task6_Dependency.md`
Section: `5. Database Setup` -> `Fresh local database flow`
```

### Step 7: Start backend service

```bash
cd backend/services/recommendation-service
set -a
source .env
set +a
go run ./cmd/server
```

Expected startup logs should include:

```text
recommendation.http.started
recommendation.grpc.started
```

### Step 8: Verify functionality related to `TASK_FILE_NAME`

HTTP health:

```bash
curl http://localhost:8088/healthz
```

gRPC health:

```bash
grpcurl -plaintext localhost:9088 grpc.health.v1.Health/Check
```

Task 7 home-feed RPC:

```bash
grpcurl -plaintext \
  -import-path ../../../proto \
  -proto ecommerce/recommendation/v1/recommendation.proto \
  -H 'x-request-id: local-task7-001' \
  -H 'x-session-id: session_demo' \
  -d '{"context":"RECOMMENDATION_CONTEXT_HOME_FEED","anonymous_id":"anon_demo","limit":12}' \
  localhost:9088 \
  ecommerce.recommendation.v1.RecommendationService/GetRecommendations
```

Task 7 product-detail RPC:

```bash
grpcurl -plaintext \
  -import-path ../../../proto \
  -proto ecommerce/recommendation/v1/recommendation.proto \
  -d '{"context":"RECOMMENDATION_CONTEXT_PRODUCT_DETAIL","product_id":"prod_demo","category_id":"cat_demo","limit":8}' \
  localhost:9088 \
  ecommerce.recommendation.v1.RecommendationService/GetRecommendations
```

Verify gRPC metrics over HTTP:

```bash
curl http://localhost:8088/metrics
```

Look for metrics like:

```text
recommendation_grpc_requests_total
recommendation_grpc_duration_seconds
recommendation_grpc_inflight_requests
recommendation_serve_source_total
recommendation_strategy_served_total
```

---

## 10. Running the Project

Minimal Task 7 run:

```bash
cd backend/services/recommendation-service
set -a
source .env
set +a
go run ./cmd/server
```

Task-specific verification flow:

1. Confirm previous setup docs were checked.
2. Start MongoDB/Redis only if real recommendation data is needed.
3. Load `.env`.
4. Start service.
5. Check HTTP health.
6. Check gRPC health.
7. Call `GetRecommendations`.
8. Check `/metrics` for gRPC counters.
9. Check logs for `recommendation.grpc.request`.

Expected successful gRPC behavior:

| Request Field | Meaning |
|---|---|
| `context` | Required placement, for example `HOME_FEED` or `PRODUCT_DETAIL`. |
| `user_id` | Optional logged-in identity. |
| `anonymous_id` | Optional guest identity. |
| `product_id` | Required for product-detail style recommendations. |
| `category_id` | Used for category scoped/fallback recommendations. |
| `seller_id` | Used for seller-store scoped recommendations. |
| `cart_product_ids` | Used for cart/checkout contexts. |
| `limit` | `0` means default; negative value is invalid. |
| `type` | Optional explicit type; unspecified lets service resolve it. |

Expected response fields:

| Response Field | Meaning |
|---|---|
| `recommendation_id` | Stable generated/stored recommendation result id. |
| `type` | Resolved recommendation type. |
| `strategy_id` | Strategy used, useful for logs/analytics/A-B hooks. |
| `items` | Ranked product references with score/reason. |
| `generated_at` | When result was generated. |
| `cache_ttl_seconds` | Remaining/effective cache TTL. |

---

## 11. Common Errors & Fixes

Generic Go, MongoDB, Redis, Kafka, Docker, and `.env` issues are already documented in earlier files. This section lists only Task 7-specific gRPC issues.

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `Failed to dial target host "localhost:9088"` | gRPC server not running or wrong port. | Start service and confirm `RECOMMENDATION_GRPC_ADDR`. | Keep port `9088` free or document overrides. |
| `address already in use` for `:9088` | Another process uses gRPC port. | Stop old process or set `RECOMMENDATION_GRPC_ADDR=:9090`. | Check ports before running multiple local services. |
| `server does not support the reflection API` | Reflection is not registered in current code. | Use `grpcurl` with `-import-path` and `-proto`. | Document proto-based `grpcurl` commands for local testing. |
| `InvalidArgument: context is required` | Request used `RECOMMENDATION_CONTEXT_UNSPECIFIED` or omitted context. | Send a supported context enum. | Treat `context` as mandatory in every client. |
| `InvalidArgument: limit cannot be negative` | Request sent negative `limit`. | Use `0` for default or a positive number. | Validate client input before RPC call. |
| `InvalidArgument: unknown type enum` | Client generated code/proto is older or mismatched. | Regenerate client from current `recommendation.proto`. | Keep clients pinned to same proto version. |
| `DeadlineExceeded: recommendation deadline exceeded` | Client or server default deadline expired. | Increase client deadline for local debugging or improve data/cache readiness. | Gateway should set realistic deadlines and avoid slow request paths. |
| `Canceled: recommendation request canceled` | Caller canceled context, browser request ended, or gateway timed out. | Retry only if caller still needs data. | Propagate cancellations intentionally. |
| `Unavailable: recommendations temporarily unavailable` | Storage/ranking path failed. | Verify MongoDB, Redis, Task 5/6 data, and logs. | Keep fallback/generated sets warm. |
| Empty `items` array | No generated/cached candidates or missing feature data. | Run Task 5/6 setup, seed product features, or inspect `recommendation_sets`. | Keep ranking jobs and feature builder healthy. |
| Response exceeds send limit | Too many items or large payload versus `RECOMMENDATION_GRPC_MAX_SEND_BYTES`. | Lower `limit` or increase send bytes carefully. | Keep `RECOMMENDATION_MAX_LIMIT` bounded. |
| Request exceeds receive limit | Huge `cart_product_ids` or identifiers exceed configured size. | Reduce request size or increase receive bytes carefully. | Validate client payload size. |
| `go: module proto-gen/go not found` | Service folder moved or local `replace` path broken. | Run from intact repo layout. | Clone/use whole repository, not only service folder. |

---

## 12. Security & Best Practices

Task 7-specific best practices:

- Keep gRPC port `9088` inside private network. Public browser traffic should go through Gateway/API layer.
- Set client deadlines. Server default is `500ms`, but Gateway/Product Service should still pass explicit deadlines.
- Pass request tracing metadata like `x-request-id` or `traceparent`.
- Keep `RECOMMENDATION_MAX_LIMIT` bounded; do not allow clients to request unbounded lists.
- Do not log full user behavior payloads or sensitive identity data. Use request IDs and low-cardinality labels.
- Regenerate clients from `recommendation.proto`; never hand-write proto-compatible JSON structures in production clients.
- Keep product hydration outside `${SERVICE_NAME}` so service ownership boundaries stay clean.
- Monitor `recommendation_grpc_requests_total`, `recommendation_grpc_duration_seconds`, and `recommendation_strategy_served_total`.
- Treat `strategy_id` as operational/analytics metadata. Task 8 owns A/B assignment setup in detail.

Security audit notes:

| Finding | Status | Suggested Fix |
|---|---|---|
| gRPC has no visible auth/inter-service mTLS in repo | Missing for production hardening | Add Gateway/service mesh auth or mTLS before production exposure. |
| No service Dockerfile detected | Packaging gap reused from earlier tasks | Add Dockerfile with HTTP and gRPC health checks. |
| No Kubernetes manifests detected | Deployment assumptions not codified | Add readiness/liveness probes for `/healthz` and gRPC health. |
| gRPC reflection not registered | Acceptable for production, inconvenient for local debugging | Keep disabled in prod; optionally enable only in local/dev. |
| `.env` is not auto-loaded | Existing behavior | Use shell export step or a supervised run script. |
| No Task 7 migration | Correct | Reuse existing schema migrations from Task 2-6. |

---

## 13. Missing or Misconfigured Things

| Item | Impact | Action |
|---|---|---|
| Dedicated service Dockerfile absent | Beginner cannot containerize backend service directly from repo. | Add Dockerfile when deployment task begins. |
| gRPC reflection absent | `grpcurl` without proto file will fail. | Use `-proto` command shown above or enable reflection only in dev. |
| No public REST gateway route implemented in this service | Browser cannot call gRPC endpoint directly. | Gateway should call gRPC and hydrate products. |
| Product hydration not done by recommendation service | Response contains product IDs, not product documents. | Gateway/Product Service should batch fetch products. |
| Empty result possible without Task 5/6 data | gRPC works but returns no useful items. | Seed feature data and generated recommendation sets. |
| No separate gRPC readiness probe config file | Deployment readiness is manual. | Add gRPC health probe in deployment manifests. |

---

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | Project tech stack, Go modules, `.env` loading, gRPC env vars, ports, Docker examples, run commands | Same base backend setup already documented. |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | MongoDB/Redis setup, cache variables, storage migrations | Task 7 serves data from same storage/cache stack. |
| `TaskImplementation/${SERVICE_NAME}/task3_Dependency.md` | Kafka setup and event troubleshooting | Live behavior freshness uses same ingestion pipeline. |
| `TaskImplementation/${SERVICE_NAME}/task4_Dependency.md` | Feature-store collections, replica set requirement | Personalized/ranking data source reused by gRPC endpoint. |
| `TaskImplementation/${SERVICE_NAME}/task5_Dependency.md` | Rule-based ranking setup and generated set verification | Trending/fallback results served by Task 7. |
| `TaskImplementation/${SERVICE_NAME}/task6_Dependency.md` | Personalized ranking setup and personalized gRPC verification pattern | Personalized result path served by Task 7. |

---

## 15. Final Checklist

- [ ] Previous dependency documentation checked.
- [ ] No duplicate MongoDB/Redis/Kafka/Docker setup added.
- [ ] No new Go dependency required for `TASK_FILE_NAME`.
- [ ] `recommendation.proto` exists at `proto/ecommerce/recommendation/v1/recommendation.proto`.
- [ ] Generated Go proto files exist under `backend/proto-gen/go`.
- [ ] `go test ./internal/transport/grpc` passes.
- [ ] `RECOMMENDATION_GRPC_ADDR` default/override confirmed.
- [ ] Backend service starts and logs `recommendation.grpc.started`.
- [ ] gRPC health check verified with `grpcurl`.
- [ ] `GetRecommendations` verified for at least one context.
- [ ] `/metrics` checked for gRPC metrics.
- [ ] MongoDB/Redis data prerequisites checked if response is empty.
- [ ] gRPC port is kept internal/private for production.
- [ ] No original implementation task file modified.
