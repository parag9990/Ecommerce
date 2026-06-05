# Project Dependency & Setup Guide

## Variables Used

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Recommendation Service` |
| `TASK_FILE_NAME` | `task8.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task8_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}` |
| Backend code path | `backend/services/recommendation-service` |
| A/B migration path | `backend/services/recommendation-service/migrations/006_add_ab_test_assignment_indexes.up.js` |
| Proto contract path | `proto/ecommerce/recommendation/v1/recommendation.proto` |

Beginner note: Ye file sirf dependency, setup, environment, database, DevOps, and verification guide hai. Business logic ya `INPUT_FILE_PATH` modify nahi kiya gaya.

---

## 1. Project Overview

`TASK_FILE_NAME` ka focus hai: recommendation response me stable `strategy_id` return karna, taaki analytics conversion ko strategy-wise compare kar sake.

Simple Hinglish:

- User ko products list milti hai.
- Response ke saath `strategy_id` bhi aata hai, jaise `personalized_v1_behavior`.
- Frontend/Gateway same `strategy_id` ko impression, click, add-to-cart, checkout, purchase events ke saath Session Analytics me bhej sakta hai.
- Optional A/B assignment hook enabled ho sakta hai, jisme user/anonymous identity ko ek stable variant milta hai.

Read these previous dependency files first:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`1. Project Tech Stack Analysis`
`2. Go Dependency System`
`3. Databases`
`5. External Services`
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
`TaskImplementation/${SERVICE_NAME}/task7_Dependency.md`

Sections:
`2. Tech Stack`
`4. Dependency Management`
`7. Environment Variables`
`9. Local Development Setup`
`10. Running the Project`
```

`TASK_FILE_NAME` new setup scope:

- A/B hook environment variables.
- Optional `RECOMMENDATION_AB_EXPERIMENTS_JSON` experiment config.
- MongoDB indexes for `ab_test_assignments`.
- Verification that `strategy_id` is present in gRPC response, logs, metrics, and assignment storage.
- No new language runtime, database engine, queue engine, Docker container, SaaS provider, or port.

---

## 2. Tech Stack

Most technologies are already explained in previous dependency files. Do not duplicate full installation steps here.

| Technology | New or Reused | Required? | Beginner Explanation | Current Task Use |
|---|---|---:|---|---|
| Go | Reused | Yes | Go backend language hai. Service binary Go me run hoti hai. | A/B config parsing, assignment logic, serving response mapping. |
| Go modules | Reused | Yes | `go.mod` and `go.sum` dependencies ko version ke saath manage karte hain. | No new module dependency is added by `TASK_FILE_NAME`. |
| gRPC | Reused | Yes | gRPC fast internal API protocol hai. Simple English: typed service-to-service call. | `GetRecommendationsResponse.strategy_id` return karta hai. |
| Protocol Buffers | Reused | Yes for gRPC | Strongly typed API contract format. | `strategy_id` field already exists in proto response. |
| MongoDB | Reused, task-specific collection/index use | Optional until A/B enabled | MongoDB document database hai. JSON-like documents collections me store hote hain. | Sticky A/B assignments store karne ke liye `ab_test_assignments`. |
| Redis | Reused | Optional but recommended | Redis in-memory cache hai. Fast temporary reads ke liye use hota hai. | Cached recommendation response me `strategy_id` preserve hota hai. |
| Kafka | Reused | Optional for live events | Kafka event streaming platform hai. | Direct A/B assignment ke liye new Kafka topic nahi hai. Session events may later carry `strategy_id`. |
| Prometheus client | Reused, task-specific metrics | Recommended | Metrics expose karne ki Go library. | `recommendation_strategy_served_total` and A/B assignment counters. |
| Session Analytics contract | External integration contract | Required for conversion comparison | Session Service events user behavior track karte hain. | UI/Gateway ko `strategy_id` event properties me copy karna hota hai. |

No new MySQL, PostgreSQL, RabbitMQ, NATS, MinIO, Elasticsearch, AWS S3, Firebase, SMTP, Stripe, Twilio, OAuth provider, Kubernetes, or Nginx setup is introduced by `TASK_FILE_NAME`.

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
| Go `1.26.3` or compatible newer | Build, test, and run backend service | Reused |
| MongoDB `7.x` | Persist A/B assignments when hook is enabled | Reused, task-specific collection/indexes |
| `mongosh` | Run and verify `006_add_ab_test_assignment_indexes.up.js` | Reused |
| Redis `7.x` | Cache response with `strategy_id` | Reused |
| Kafka | Live behavior event ingestion only | Reused, not required just for A/B assignment verification |
| `grpcurl` | Manual `GetRecommendations` verification | Reused |
| `curl` | Health and metrics checks | Reused |

Beginner decision:

- Sirf `strategy_id` response verify karna hai: Go service plus Task 7 gRPC setup enough hai.
- Sticky A/B assignment verify karna hai: MongoDB configured hona chahiye and migration `006` run honi chahiye.
- Conversion dashboard compare karna hai: Session Analytics side ko events me `strategy_id` receive/store karna hoga. Ye repo me current task ke through new service install nahi karta.

---

## 4. Dependency Management

This is still the same Go module project. Go modules, `go.mod`, `go.sum`, `go mod download`, `go mod tidy`, and common Go dependency errors are already explained in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `2. Go Dependency System`
```

`TASK_FILE_NAME` dependency status:

| Dependency | Version In `go.mod` | Used For | New Install Needed? |
|---|---:|---|---:|
| `github.com/example/ecommerce-platform/backend/proto-gen/go` | `v0.0.0` via local `replace` | `GetRecommendationsResponse.strategy_id` generated proto field | No |
| `google.golang.org/grpc` | `v1.81.1` | gRPC response serving and health checks | No |
| `google.golang.org/protobuf` | `v1.36.11` | Protobuf runtime | No |
| `go.mongodb.org/mongo-driver/v2` | `v2.6.0` | `ab_test_assignments` read/upsert | No |
| `github.com/prometheus/client_golang` | `v1.23.2` | Strategy and A/B assignment metrics | No |
| `github.com/redis/go-redis/v9` | `v9.19.0` | Cached recommendations preserving `strategy_id` | No |

No new `go get` command is needed for `TASK_FILE_NAME`.

Recommended verification:

```bash
cd backend/services/recommendation-service
go test ./internal/domain ./internal/usecase ./internal/config ./internal/repository ./internal/transport/grpc
```

Full service verification:

```bash
cd backend/services/recommendation-service
go test ./...
```

---

## 5. Database Setup

### MongoDB Role

MongoDB setup itself is reused. Current task only adds/uses A/B assignment indexes.

Simple Hinglish:

`ab_test_assignments` collection ek identity ko experiment ke andar same variant se stick rakhta hai. Agar same user baar-baar home feed open kare, to usko same experiment variant milna chahiye, warna analytics comparison dirty ho jayega.

### Required Or Optional

| Scenario | MongoDB Required? | Reason |
|---|---:|---|
| A/B disabled and normal recommendation response | No, service can still fail open depending on storage mode |
| `strategy_id` field in response | No extra DB beyond existing serving path |
| A/B enabled with sticky assignments | Yes | Assignment repository MongoDB me stored hai |
| A/B enabled but MongoDB missing and `RECOMMENDATION_AB_FAIL_OPEN=true` | Recommended but not hard-blocking | Service logs warning and returns default strategy |
| A/B enabled and `RECOMMENDATION_AB_FAIL_OPEN=false` | Yes | Assignment failure request error ban sakta hai |

### Reused MongoDB Setup

MongoDB installation, Docker run command, connection string format, credentials placement, and base collections are already explained in:

| Previous File | Section / Topic |
|---|---|
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `3. Databases` -> `MongoDB` |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` |
| `TaskImplementation/${SERVICE_NAME}/task4_Dependency.md` | `5. Database Setup` -> replica set requirement for feature jobs |
| `TaskImplementation/${SERVICE_NAME}/task5_Dependency.md` | `5. Database Setup` -> rule-based ranking data |
| `TaskImplementation/${SERVICE_NAME}/task6_Dependency.md` | `5. Database Setup` -> personalized ranking data |

### Current Task Migration

Run this only after MongoDB is running and `RECOMMENDATION_MONGO_DATABASE` points to the correct database. If you keep values in `.env`, export them first:

```bash
cd backend/services/recommendation-service
set -a
source .env
set +a
export RECOMMENDATION_MONGO_DATABASE=recommendation_db
mongosh "$RECOMMENDATION_MONGO_URI" --file migrations/006_add_ab_test_assignment_indexes.up.js
```

If your local MongoDB auth requires the admin auth database, use:

```bash
cd backend/services/recommendation-service
export RECOMMENDATION_MONGO_DATABASE=recommendation_db
mongosh "mongodb://root:secret@localhost:27017/admin" --file migrations/006_add_ab_test_assignment_indexes.up.js
```

### Collection Used By Current Task

| Collection | Purpose | Required? |
|---|---|---:|
| `ab_test_assignments` | Stores one active assignment per `experiment_id + assignment_key` | Required only when A/B testing is enabled |

Important fields:

| Field | Meaning |
|---|---|
| `experiment_id` | Experiment name, for example `reco_home_strategy_2026_05` |
| `assignment_key` | Stable bucket identity like `user:user_123` or `anon:anon_123` |
| `user_id` | Known user id, if available |
| `anonymous_id` | Guest/anonymous id, if available |
| `session_id` | Session id, if available in domain flow |
| `context` | Recommendation context, for example `home_feed` |
| `recommendation_type` | Recommendation type, for example `personalized` |
| `variant_id` | Bucket label, for example `control` or `treatment` |
| `strategy_id` | Selected recommendation strategy |
| `assigned_at` | Assignment creation/update time |
| `expires_at` | TTL expiry time |

### Current Task Indexes

| Index | Keys | Why |
|---|---|---|
| `ux_ab_assignment_experiment_identity` | `experiment_id`, `assignment_key` | Same identity ko same experiment me duplicate assignment se bachata hai |
| `idx_ab_assignment_user_experiment` | `user_id`, `experiment_id` | Known-user assignment debug karne ke liye |
| `idx_ab_assignment_anon_experiment` | `anonymous_id`, `experiment_id` | Guest assignment debug karne ke liye |
| `idx_ab_assignment_variant_recent` | `experiment_id`, `variant_id`, `assigned_at` | Variant distribution inspect karne ke liye |
| `idx_ab_assignment_ttl` | `expires_at` | Old assignments automatically delete karne ke liye |

### Verify MongoDB

Check indexes:

```bash
mongosh "$RECOMMENDATION_MONGO_URI" \
  --eval 'db.getSiblingDB("recommendation_db").ab_test_assignments.getIndexes().map(index => index.name)'
```

Expected names include:

```text
ux_ab_assignment_experiment_identity
idx_ab_assignment_user_experiment
idx_ab_assignment_anon_experiment
idx_ab_assignment_variant_recent
idx_ab_assignment_ttl
```

After a test request, check stored assignments:

```bash
mongosh "$RECOMMENDATION_MONGO_URI" \
  --eval 'db.getSiblingDB("recommendation_db").ab_test_assignments.find({}, { _id: 1, experiment_id: 1, assignment_key: 1, variant_id: 1, strategy_id: 1, expires_at: 1 }).pretty()'
```

---

## 6. Redis / Queue / External Services

### Redis

No new Redis setup is introduced by `TASK_FILE_NAME`.

Refer:

```md
`TaskImplementation/${SERVICE_NAME}/task2_Dependency.md`
Section: `6. Redis / Queue / External Services` -> `Redis`
```

Task-specific note:

- Redis cached recommendation payload preserves `strategy_id`.
- No separate A/B cache key or Redis container is required.
- If Redis is down, service can still use MongoDB/fallback paths already documented in previous files.

### Kafka

No new Kafka topic is introduced by `TASK_FILE_NAME`.

Refer:

```md
`TaskImplementation/${SERVICE_NAME}/task3_Dependency.md`
Section: `6. Redis / Queue / External Services` -> `Kafka Role`
```

Task-specific note:

- `${SERVICE_NAME}` does not need Kafka just to assign a strategy.
- For conversion analytics, frontend/gateway/session producers should include `strategy_id` in event properties when emitting impression/click/conversion events.

### Session Analytics Contract

This is the main external integration for `TASK_FILE_NAME`.

Simple Hinglish:

`${SERVICE_NAME}` response se `strategy_id` copy karke Session Service event properties me bhejna hai. Tab analytics team conversion by strategy compare kar sakti hai.

Example event property shape:

```json
{
  "recommendation_id": "reco_home_user_123_20260605",
  "strategy_id": "personalized_v1_behavior",
  "experiment_id": "reco_home_strategy_2026_05",
  "variant_id": "treatment"
}
```

`experiment_id` and `variant_id` are useful when A/B assignment is enabled. Minimum requirement for `TASK_FILE_NAME` is still `strategy_id`.

---

## 7. Environment Variables

Base `.env` variables are already documented in:

| Previous File | Section / Topic |
|---|---|
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `4. Environment Variables` |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | `7. Environment Variables` |
| `TaskImplementation/${SERVICE_NAME}/task3_Dependency.md` | `7. Environment Variables` |
| `TaskImplementation/${SERVICE_NAME}/task7_Dependency.md` | `7. Environment Variables` |

Create or update the file here:

```text
backend/services/recommendation-service/.env
```

Current backend code reads environment variables through `internal/config/config.go`.

Important: Go code uses `os.Getenv`. `.env` file automatically load nahi hota. Edit karne ke baad local shell me export karo:

```bash
cd backend/services/recommendation-service
set -a
source .env
set +a
```

### Current Task `.env` Delta

Use this only when you want to enable A/B assignment locally:

```env
RECOMMENDATION_AB_TESTING_ENABLED=true
RECOMMENDATION_AB_ASSIGNMENT_TTL_DAYS=90
RECOMMENDATION_AB_DEFAULT_SALT=reco-ab-v1-local
RECOMMENDATION_AB_FAIL_OPEN=true
RECOMMENDATION_AB_EXPERIMENTS_JSON=[{"experiment_id":"reco_home_strategy_2026_05","context":"home_feed","type":"personalized","status":"active","variants":[{"variant_id":"control","strategy_id":"trending_v1_recent_activity","weight":50},{"variant_id":"treatment","strategy_id":"personalized_v1_behavior","weight":50}]}]
```

Keep the JSON on one line in `.env` unless your local env loader supports quoted multi-line values.

### Variable Reference

| Variable | Required? | Default | Purpose | Security / Safety Note |
|---|---:|---|---|---|
| `RECOMMENDATION_AB_TESTING_ENABLED` | Optional | `false` | A/B assignment hook on/off karta hai | Keep `false` until experiment config and MongoDB are ready |
| `RECOMMENDATION_AB_ASSIGNMENT_TTL_DAYS` | Optional | `90` | Assignment kitne din valid rahega | Too short TTL users ko variants flip kara sakta hai |
| `RECOMMENDATION_AB_DEFAULT_SALT` | Optional but important | `reco-ab-v1` | Deterministic bucket hashing ke liye salt | Mid-experiment rotate mat karo, warna assignments change ho jayenge |
| `RECOMMENDATION_AB_FAIL_OPEN` | Optional | `true` | Assignment fail hone par default strategy return kare | Product flow ko block na karne ke liye `true` safer hai |
| `RECOMMENDATION_AB_EXPERIMENTS_JSON` | Required if A/B enabled with real experiment | `[]` | Experiment definitions, variants, strategy mapping | JSON valid hona chahiye and variant weights total `100` |

Existing variables reused by current task:

| Variable | Why Current Task Needs It | Already Documented In |
|---|---|---|
| `RECOMMENDATION_MONGO_URI` | Assignment repository MongoDB se connect karta hai | `task2_Dependency.md`, section `7. Environment Variables` |
| `RECOMMENDATION_MONGO_DATABASE` | `ab_test_assignments` collection same DB me hoti hai | `task2_Dependency.md`, section `7. Environment Variables` |
| `RECOMMENDATION_MONGO_ENSURE_INDEXES` | App startup par indexes ensure kar sakta hai | `task2_Dependency.md`, section `7. Environment Variables` |
| `RECOMMENDATION_MAX_IDENTIFIER_LENGTH` | Experiment identity fields length validate karne ke liye reused | `task1_Dependency.md`, section `4. Environment Variables` |
| `RECOMMENDATION_GRPC_ADDR` | `strategy_id` gRPC response verify karne ke liye | `task7_Dependency.md`, section `7. Environment Variables` |

### Experiment JSON Rules

| Field | Required? | Example | Notes |
|---|---:|---|---|
| `experiment_id` | Yes | `reco_home_strategy_2026_05` | No spaces/control characters |
| `context` | Yes | `home_feed` | Must match supported context |
| `type` | Yes | `personalized` | Must match supported recommendation type |
| `status` | Yes for active experiment | `active` | Other valid values: `draft`, `paused`, `stopped` |
| `salt` | Optional | `reco-home-2026-05` | If empty, `RECOMMENDATION_AB_DEFAULT_SALT` is used |
| `variants` | Yes | See `.env` delta | At least one variant, weights total `100` |
| `starts_at` | Optional | `2026-06-01T00:00:00Z` | Must be UTC/parseable when used |
| `ends_at` | Optional | `2026-07-01T00:00:00Z` | Must be after `starts_at` |

Supported `context` values:

```text
home_feed
product_detail
category_listing
cart
checkout
search_results
seller_store
```

Supported `type` values:

```text
similar_products
trending
personalized
frequently_bought_together
```

Allowed `strategy_id` format:

```text
<type>_v<version>_<main_logic>
```

Current strategy examples:

```text
similar_v1_attributes
trending_v1_recent_activity
personalized_v1_behavior
fbt_v1_order_cooccurrence
trending_v1_fallback_global
trending_v1_fallback_category
```

---

## 8. Docker Setup

No new Docker container, volume, network, exposed port, or compose service is introduced by `TASK_FILE_NAME`.

Reuse Docker instructions from:

| Previous File | Section / Topic |
|---|---|
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `7. Docker and DevOps Setup` |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | `8. Docker Setup` |
| `TaskImplementation/${SERVICE_NAME}/task3_Dependency.md` | `8. Docker Setup` for Kafka when events are enabled |
| `TaskImplementation/${SERVICE_NAME}/task7_Dependency.md` | `8. Docker Setup` for gRPC port note |

Task-specific Docker notes:

- If backend runs on host machine, MongoDB URI usually uses `localhost:27017`.
- If backend runs inside Docker Compose, MongoDB URI usually uses service DNS like `mongo:27017`.
- A/B assignment requires only the same MongoDB container already used by previous tasks.
- Run migration `006` against the same `RECOMMENDATION_MONGO_DATABASE` used by the backend.

---

## 9. Local Development Setup

### Step 1: Read Previous Dependency Documentation

Follow these first:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
`TaskImplementation/${SERVICE_NAME}/task2_Dependency.md`
`TaskImplementation/${SERVICE_NAME}/task7_Dependency.md`
```

For live event freshness, also follow:

```md
`TaskImplementation/${SERVICE_NAME}/task3_Dependency.md`
`TaskImplementation/${SERVICE_NAME}/task4_Dependency.md`
`TaskImplementation/${SERVICE_NAME}/task5_Dependency.md`
`TaskImplementation/${SERVICE_NAME}/task6_Dependency.md`
```

### Step 2: Go To Project Directory

```bash
cd backend/services/recommendation-service
```

### Step 3: Install Dependencies

No new dependencies are needed. Reuse Go module setup:

```bash
go mod download
```

### Step 4: Start Required Services

For A/B assignment verification:

| Service | Required? | Reference |
|---|---:|---|
| MongoDB | Yes | `task2_Dependency.md`, section `5. Database Setup` |
| Redis | Optional | `task2_Dependency.md`, section `6. Redis / Queue / External Services` |
| Kafka | No for assignment, yes for live events | `task3_Dependency.md`, section `6. Redis / Queue / External Services` |

### Step 5: Add Current Task Environment Variables

Update:

```text
backend/services/recommendation-service/.env
```

Minimum local A/B delta:

```env
RECOMMENDATION_AB_TESTING_ENABLED=true
RECOMMENDATION_AB_ASSIGNMENT_TTL_DAYS=90
RECOMMENDATION_AB_DEFAULT_SALT=reco-ab-v1-local
RECOMMENDATION_AB_FAIL_OPEN=true
RECOMMENDATION_AB_EXPERIMENTS_JSON=[{"experiment_id":"reco_home_strategy_2026_05","context":"home_feed","type":"personalized","status":"active","variants":[{"variant_id":"control","strategy_id":"trending_v1_recent_activity","weight":50},{"variant_id":"treatment","strategy_id":"personalized_v1_behavior","weight":50}]}]
```

Then export `.env` into the shell before migration or `go run`:

```bash
set -a
source .env
set +a
```

### Step 6: Run Migration

```bash
export RECOMMENDATION_MONGO_DATABASE=recommendation_db
mongosh "$RECOMMENDATION_MONGO_URI" --file migrations/006_add_ab_test_assignment_indexes.up.js
```

### Step 7: Start Backend Service

Use the same command already documented previously:

```bash
go run ./cmd/server
```

Expected startup logs:

```text
recommendation.http.started
recommendation.grpc.started
```

If A/B is disabled:

```text
recommendation.ab_testing.disabled
```

If A/B is enabled but no experiment JSON is present:

```text
recommendation.ab_testing.no_experiments
```

### Step 8: Verify Functionality Related To `TASK_FILE_NAME`

Health:

```bash
curl http://localhost:8088/healthz
```

gRPC request:

```bash
grpcurl -plaintext \
  -import-path ../../../proto \
  -proto ecommerce/recommendation/v1/recommendation.proto \
  -H 'x-request-id: local-ab-hook-001' \
  -d '{"context":"RECOMMENDATION_CONTEXT_HOME_FEED","anonymous_id":"anon_ab_demo","limit":12,"type":"RECOMMENDATION_TYPE_PERSONALIZED"}' \
  localhost:9088 \
  ecommerce.recommendation.v1.RecommendationService/GetRecommendations
```

Expected response should include `strategyId` in `grpcurl` JSON output:

```json
{
  "recommendationId": "...",
  "type": "RECOMMENDATION_TYPE_PERSONALIZED",
  "strategyId": "personalized_v1_behavior",
  "items": []
}
```

Note: Protobuf JSON uses camelCase `strategyId`. REST/API JSON may use snake_case `strategy_id`.

Verify assignment document:

```bash
mongosh "$RECOMMENDATION_MONGO_URI" \
  --eval 'db.getSiblingDB("recommendation_db").ab_test_assignments.find({ experiment_id: "reco_home_strategy_2026_05" }, { assignment_key: 1, variant_id: 1, strategy_id: 1, expires_at: 1 }).pretty()'
```

Verify metrics:

```bash
curl http://localhost:8088/metrics
```

Look for:

```text
recommendation_strategy_served_total
recommendation_ab_assignment_total
recommendation_ab_assignment_error_total
```

---

## 10. Running the Project

### Minimal Run Without A/B

Use this when you only want stable `strategy_id` in responses:

```env
RECOMMENDATION_AB_TESTING_ENABLED=false
RECOMMENDATION_AB_EXPERIMENTS_JSON=[]
```

Then:

```bash
cd backend/services/recommendation-service
go run ./cmd/server
```

This mode should still return a `strategy_id` from existing serving logic.

### A/B Assignment Run

Use this when you want a request identity to get a sticky strategy variant:

1. MongoDB running.
2. Migration `006_add_ab_test_assignment_indexes.up.js` applied.
3. `RECOMMENDATION_AB_TESTING_ENABLED=true`.
4. `RECOMMENDATION_AB_EXPERIMENTS_JSON` contains at least one `active` experiment.
5. Request contains `user_id` or `anonymous_id`.

Example experiment behavior:

| Variant | Strategy | Weight |
|---|---|---:|
| `control` | `trending_v1_recent_activity` | 50 |
| `treatment` | `personalized_v1_behavior` | 50 |

Same `experiment_id + assignment_key + salt` should choose the same bucket each time.

---

## 11. Common Errors & Fixes

Generic setup errors are already documented in previous dependency files. Current task-specific issues:

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `RECOMMENDATION_AB_EXPERIMENTS_JSON must be valid experiment JSON` | JSON malformed hai | JSON lint karo and one-line `.env` value use karo | Commit example config separately from secrets |
| `variant weights must total 100` | Variant weights ka sum 100 nahi hai | Weights adjust karo, for example `50 + 50` | Experiment launch checklist me weight validation rakho |
| `experiment_id cannot contain whitespace` | ID me space, tab, newline hai | Use `reco_home_strategy_2026_05` style | IDs ko lowercase snake-case rakho |
| `strategy_id must match <type>_v<version>_<main_logic>` | Strategy naming invalid hai | Existing valid strategy ID use karo | User/product IDs kabhi strategy ID me mat dalo |
| `strategy type ... is incompatible with experiment type` | Experiment `type` and variant `strategy_id` mismatch | `personalized` experiment me personalized or trending fallback strategy use karo | Type/strategy table review karo before enabling |
| No documents in `ab_test_assignments` | A/B disabled, no active experiment, no identity, or request context/type mismatch | Env vars and request payload check karo | Verification request me `anonymous_id` or `user_id` include karo |
| Log says `recommendation.ab_testing.repository_not_configured` | MongoDB not configured/unreachable | Set `RECOMMENDATION_MONGO_URI`, start MongoDB, rerun service | A/B enable karne se pehle storage status verify karo |
| Request still succeeds with default strategy after assignment failure | `RECOMMENDATION_AB_FAIL_OPEN=true` | Expected fail-open behavior hai | Production me product flow availability ke liye document this behavior |
| Request fails when MongoDB is down | `RECOMMENDATION_AB_FAIL_OPEN=false` | MongoDB restore karo or set fail-open true | Fail-closed mode sirf strict experiments ke liye use karo |
| Assignments change unexpectedly | Salt changed, experiment ID changed, TTL expired, or identity changed | Original salt/experiment id restore karo | Active experiment ke beech salt rotate mat karo |
| `idx_ab_assignment_ttl` missing | Migration `006` not run or wrong DB selected | Set `RECOMMENDATION_MONGO_DATABASE=recommendation_db` and rerun migration | Fresh DB setup checklist me migration `006` include karo |

---

## 12. Security & Best Practices

Task-specific practices:

| Practice | Why |
|---|---|
| Keep `strategy_id` non-sensitive | Public/API response and metrics me expose hota hai |
| Do not include user IDs in `strategy_id`, `experiment_id`, or `variant_id` | Metrics labels low-cardinality and non-PII rehne chahiye |
| Keep experiment IDs stable | Analytics grouping clean rahega |
| Do not rotate salt during active experiment | Same users different variants me flip ho sakte hain |
| Prefer `RECOMMENDATION_AB_FAIL_OPEN=true` for commerce flows | Recommendation failure product/cart/checkout journey block nahi karegi |
| Use MongoDB TTL on assignments | Old assignments automatic cleanup ho jate hain |
| Keep variant weights explicit and reviewed | Accidental 90/10 split conversion analysis ko bias kar sakta hai |
| Store production experiment JSON through managed config/secret process | Manual `.env` edits risky hote hain |
| Keep metrics labels low-cardinality | Prometheus cost and performance stable rahega |

---

## 13. Missing Or Misconfigured Things

Current repository audit for `TASK_FILE_NAME`:

| Item | Status | Recommendation |
|---|---|---|
| A/B dashboard | Not implemented | Future Superadmin/Analytics task me add karo |
| Statistical significance engine | Not implemented | Analytics/reporting layer me add karo |
| Dedicated tracking RPC | Not implemented in this task | Session Analytics event contract me `strategy_id` pass karo |
| Frontend event copy | Not implemented in backend repo | UI/Gateway team ko response `strategy_id` event properties me copy karna hoga |
| Migration runner | Manual `mongosh` migrations | Production me repeatable migration runner/process add karo |
| Production salt management | Sample `.env` value exists | Environment-specific salt managed config me rakho |
| A/B config validation before deploy | Runtime config validation exists | CI/pre-deploy config validation helpful hoga |
| Docker Compose file in repo root | Not found during inspection | Previous docs use direct Docker commands; add compose later if team wants one-command setup |

---

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | Project tech stack, Go modules, MongoDB basics, external services, ports, Docker, common errors | Base service setup already documented |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | MongoDB, Redis, storage env vars, Docker setup | Current task reuses same MongoDB and Redis setup |
| `TaskImplementation/${SERVICE_NAME}/task3_Dependency.md` | Kafka event setup and troubleshooting | Conversion/event flow can use existing event infrastructure, no new topic required |
| `TaskImplementation/${SERVICE_NAME}/task4_Dependency.md` | Feature store MongoDB replica set notes | Full feature jobs may still need replica set setup |
| `TaskImplementation/${SERVICE_NAME}/task5_Dependency.md` | Rule-based ranking data and `trending_v1_recent_activity` strategy | A/B control variant can reuse this strategy |
| `TaskImplementation/${SERVICE_NAME}/task6_Dependency.md` | Personalized ranking data and `personalized_v1_behavior` strategy | A/B treatment variant can reuse this strategy |
| `TaskImplementation/${SERVICE_NAME}/task7_Dependency.md` | gRPC setup, ports, `grpcurl`, metrics verification | `TASK_FILE_NAME` verifies `strategy_id` through the same gRPC endpoint |

---

## 15. Final Checklist

- [ ] Previous dependency documentation checked before using this file.
- [ ] No duplicate base installation steps added for Go, MongoDB, Redis, Kafka, or Docker.
- [ ] No new Go dependencies required for `TASK_FILE_NAME`.
- [ ] MongoDB is running if A/B assignment is enabled.
- [ ] `RECOMMENDATION_MONGO_URI` and `RECOMMENDATION_MONGO_DATABASE` point to the correct database.
- [ ] Migration `006_add_ab_test_assignment_indexes.up.js` applied.
- [ ] `RECOMMENDATION_AB_TESTING_ENABLED` set intentionally.
- [ ] `RECOMMENDATION_AB_EXPERIMENTS_JSON` is valid JSON.
- [ ] Variant weights total `100`.
- [ ] Experiment `context`, `type`, and `strategy_id` values are compatible.
- [ ] Backend service starts successfully.
- [ ] gRPC response contains `strategyId`.
- [ ] Logs include `strategy_id`, and when assigned, `experiment_id` plus `variant_id`.
- [ ] `ab_test_assignments` documents appear after matching test requests.
- [ ] Metrics include strategy served and A/B assignment counters.
- [ ] Frontend/Gateway/Session event contract copies `strategy_id` for conversion analytics.
