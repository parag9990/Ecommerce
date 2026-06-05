# Project Dependency & Setup Guide

## 1. Project Overview

| Variable | Value |
|---|---|
| `SERVICE_NAME` | Prompt variable from the assignment |
| `TASK_FILE_NAME` | Prompt variable from the assignment |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `{TASK_FILE_NAME without .md}_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

Current task ka focus admin synonym management hai. Simple Hinglish me: catalog admin ya superadmin search terms ke alternate words manage karta hai, jaise `mobile` ke liye `phone`, `smartphone`, `cellphone`. Runtime me ye synonyms Typesense ke `products` collection par save hote hain, taaki buyer search relevance better ho.

Important boundaries:

- Original `TASK_FILE_NAME` modify nahi kiya gaya.
- Business logic rewrite nahi kiya gaya.
- Base Go, Docker, Redis, RabbitMQ, Typesense, and full `.env` setup duplicate nahi kiya gaya.
- Previous dependency docs ko pehle follow karna hai; yahan sirf synonym-management-specific setup aur checks explain kiye gaye hain.

Current runnable flow in this repo snapshot:

```text
Admin client / Gateway
  -> GET or POST /api/v1/admin/search/synonyms
  -> header-based admin auth in local Search HTTP handler
  -> fixed-window admin mutation rate limiter for POST
  -> synonym validation and deterministic id creation
  -> Typesense products collection synonym upsert/list
  -> JSON response plus structured audit log
```

Note: `api/master-api.json` defines `SearchService.CreateSynonym` and `SearchService.ListSynonyms`, but this repo snapshot does not contain generated proto/gRPC server files for this service. The currently runnable path is the Go HTTP handler.

## 2. Tech Stack

### Reused tech from previous dependency files

| Technology | Required for this task? | Reuse reference |
|---|---:|---|
| Go | Yes | `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Section `2. Go Dependency System` |
| Go modules | Yes | `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Section `2. Go Dependency System` |
| Standard Go `net/http` | Yes | `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Section `1. Project Tech Stack Analysis` |
| Typesense | Yes | `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`, Section `4.1 Typesense` |
| Docker Compose | Recommended for local infra | `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Section `7. Docker and DevOps Setup` |
| Redis | Required by full service startup | `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Section `3.2 Redis` |
| RabbitMQ | Required only when product indexer is enabled | `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`, Section `6. Redis / Queue / External Services` |

### Task-specific tech details

| Component | Required? | Beginner-friendly explanation |
|---|---:|---|
| Admin synonym HTTP endpoint | Yes | Admin-only route hai jisse synonyms create/update and list hote hain. |
| Typesense collection synonyms API | Yes | Typesense ke collection-level synonym feature me rules save hote hain. Ye normal SQL table nahi hai. |
| `products` collection / alias | Yes | Product search isi collection par hota hai, isliye synonyms bhi isi collection ke saath attach hote hain. |
| Header-based admin auth | Yes in current local code | Local implementation `X-User-ID` and `X-Roles` headers check karta hai. Production me gateway/JWT se trusted headers forward hone chahiye. |
| Fixed-window admin rate limiter | Yes for writes | Same admin ko short window me too many mutation requests bhejne se rokta hai. |
| Structured audit logs | Yes | Admin mutation ka trace log hota hai: actor, role, action, resource id, root, count. |

## 3. Required Software

No brand-new runtime software install is introduced by `TASK_FILE_NAME`.

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
| Typesense | Yes | Stores and applies collection synonyms | `curl http://localhost:8108/health` |
| Redis | Yes for full service boot | Used by existing cache/dedupe/lock wiring | `docker exec ecommerce-redis redis-cli -a dev_redis_password ping` |
| Docker Compose | Recommended | Starts local Typesense, Redis, RabbitMQ | `docker compose version` |
| RabbitMQ | Only if `SEARCH_INDEXER_ENABLED=true` | Existing product indexer consumer starts with service | `docker exec ecommerce-rabbitmq rabbitmq-diagnostics -q ping` |
| API Gateway / Auth | Production required | Real JWT/RBAC should protect admin route before traffic reaches this service | See security notes below |

## 4. Dependency Management

This is a Go project. `go.mod`, `go.sum`, `go mod download`, `go mod tidy`, and common Go module issues are already explained in:

```md
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`2. Go Dependency System`
```

Task-specific dependency audit:

| File / Dependency | Current finding | New install needed? |
|---|---|---:|
| `backend/services/search-service/go.mod` | Already has `github.com/typesense/typesense-go/v2 v2.0.0` | No |
| `github.com/rabbitmq/amqp091-go` | Already exists for indexer startup, not synonym logic directly | No |
| Go standard library | `net/http`, `context`, `time`, `log/slog`, `regexp`, `strings` are used | No |
| Node/Python packages | Not used by current task | No |

Task-specific verification commands:

```bash
cd backend/services/search-service
go test ./internal/domain ./internal/usecase ./internal/repository ./internal/transport/http
```

Live Typesense integration tests are skipped unless explicitly enabled:

```bash
cd backend/services/search-service
TYPESENSE_INTEGRATION=1 go test ./internal/repository
```

Use the integration command only after Typesense is running and `TYPESENSE_API_KEY` matches local config.

## 5. Database Setup

No SQL database, MongoDB collection, or SQL migration is introduced by `TASK_FILE_NAME`.

| Storage / DB | Used? | Purpose | Setup action |
|---|---:|---|---|
| Typesense `products` collection | Yes | Stores product search documents and collection-level synonyms | Reuse existing Typesense setup |
| Redis | Indirect | Service startup and existing cache/dedupe/lock features | Reuse existing Redis setup |
| MySQL/PostgreSQL/MongoDB | No | Current task does not write relational/admin audit tables | No setup |

Typesense setup is already explained in:

```md
`TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`

Section:
`4.1 Typesense`
```

### Task-specific Typesense synonym storage

Typesense synonyms are not stored as documents inside `products`. They are collection-level rules under:

```text
/collections/products/synonyms/{synonym_id}
```

Current implementation uses one-way synonym shape:

```json
{
  "root": "mobile",
  "synonyms": ["phone", "smartphone", "cellphone"]
}
```

Beginner note: `root` main search term hai, aur `synonyms` alternate terms hain. Search engine in terms ko relevance ke liye use karta hai. Ye Product Service ka source-of-truth data nahi hai.

### No migration required

Search schema setup is Typesense collection ensure, not SQL migration:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`8. Project Run Instructions`
```

## 6. Redis / Queue / External Services

### Typesense synonym API

Current task depends on Typesense collection synonyms:

| Operation | Current route / API | Purpose |
|---|---|---|
| Create/update synonym | `PUT /collections/products/synonyms/{id}` via Go client | Save one root synonym rule |
| List synonyms | `GET /collections/products/synonyms` via Go client | Admin UI list view |

Task-specific cURL check directly against Typesense:

```bash
curl -X PUT "http://localhost:8108/collections/products/synonyms/syn_mobile" \
  -H "X-TYPESENSE-API-KEY: dev-typesense-key" \
  -H "Content-Type: application/json" \
  -d '{"root":"mobile","synonyms":["phone","smartphone","cellphone"]}'
```

List directly:

```bash
curl "http://localhost:8108/collections/products/synonyms" \
  -H "X-TYPESENSE-API-KEY: dev-typesense-key"
```

Use direct Typesense cURL only for local debugging. Normal application flow should go through the admin API so validation, auth, rate limiting, and audit logs run.

### Redis and RabbitMQ

Current task does not add new Redis keys or queues.

```md
Redis setup:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`
Section `3.2 Redis`

RabbitMQ setup:
`TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`
Section `6. Redis / Queue / External Services`
```

For synonym-only local testing, this service can avoid RabbitMQ consumer startup by setting:

```env
SEARCH_INDEXER_ENABLED=false
```

Redis is still initialized by current server wiring, so keep Redis running or configured correctly.

## 7. Environment Variables

Full `.env` setup is already documented in:

```md
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`4. Environment Variables`
```

For current task, verify this task-specific subset in:

```text
backend/services/search-service/.env
```

```env
# Typesense target for synonym rules
TYPESENSE_PRODUCTS_COLLECTION=products
TYPESENSE_API_KEY=dev-typesense-key
TYPESENSE_HOST=localhost
TYPESENSE_PORT=8108
TYPESENSE_PROTOCOL=http

# Admin synonym endpoint behavior
SEARCH_ADMIN_TIMEOUT_MS=500
SEARCH_ADMIN_AUTH_ENABLED=true
SEARCH_ADMIN_MUTATION_RATE_LIMIT=30
SEARCH_ADMIN_MUTATION_RATE_WINDOW=1m
```

### Variable explanation

| Variable | Required? | Purpose | Example | Security / mistake note |
|---|---:|---|---|---|
| `TYPESENSE_PRODUCTS_COLLECTION` | Yes | Collection or alias where product search and synonyms live | `products` | Must match collection ensured on startup |
| `TYPESENSE_API_KEY` | Yes | Allows service to call Typesense admin APIs | `dev-typesense-key` | Secret. Never expose in browser |
| `SEARCH_ADMIN_TIMEOUT_MS` | Optional | Timeout for synonym create/list repository calls | `500` | Too low can cause flaky admin saves |
| `SEARCH_ADMIN_AUTH_ENABLED` | Optional | Enables local header-based admin auth | `true` | Set `false` only for local private testing |
| `SEARCH_ADMIN_MUTATION_RATE_LIMIT` | Optional | Max admin write calls per window | `30` | Applies to POST mutation route |
| `SEARCH_ADMIN_MUTATION_RATE_WINDOW` | Optional | Rate limiter window | `1m` | Keep bounded for admin mutation protection |

### How env is loaded

The Go service reads environment variables through `config.Load()`. It does not automatically parse `.env` files. For local shell runs:

```bash
cd backend/services/search-service
set -a
. ./.env
set +a
go run ./cmd/server
```

Common mistake: using Docker DNS names like `typesense` while running Go from the host. Host runs usually need `TYPESENSE_HOST=localhost`.

## 8. Docker Setup

No new Docker container, volume, or network is introduced by `TASK_FILE_NAME`.

Existing local infra is already covered in:

```md
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`7. Docker and DevOps Setup`
`8. Project Run Instructions`
```

Current task uses the existing Typesense container:

| Container | Needed? | Why |
|---|---:|---|
| `ecommerce-typesense` | Yes | Stores collection-level synonyms |
| `ecommerce-redis` | Yes for full server startup | Existing service wiring initializes Redis |
| `ecommerce-rabbitmq` | Only if indexer enabled | Existing product consumer dependency |

Start existing local infra:

```bash
cd infra/compose
cp .env.local.example .env.local
docker compose --env-file .env.local -f docker-compose.local.yml up -d typesense redis rabbitmq
```

### Ports and networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Backend HTTP API | `8085` | Serves `/api/v1/admin/search/synonyms` | Reused |
| Typesense API | `8108` | Collection synonym API | Reused |
| Redis | `6379` | Existing service cache/dedupe/lock dependency | Reused |
| RabbitMQ AMQP | `5672` | Existing product indexer queue when enabled | Reused |
| RabbitMQ Management | `15672` | Local queue debugging UI | Reused |

No new port is required for synonym management.

## 9. Local Development Setup

### Step 1: Read previous dependency documentation first

Read these before current guide:

```md
1. `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`
2. `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`
3. `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`
4. `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md`
5. `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md`
```

### Step 2: Go to project directory

```bash
cd backend/services/search-service
```

### Step 3: Install only new dependencies if any

No new dependency install is required for current task.

If dependencies are missing after clone:

```bash
go mod download
go mod tidy
```

### Step 4: Setup only new databases/services if any

No new database/service setup is needed. Reuse existing Typesense.

Verify Typesense:

```bash
curl "http://localhost:8108/health"
```

Expected response includes:

```json
{
  "ok": true
}
```

### Step 5: Add only new or changed environment variables

No brand-new variable names are introduced beyond what previous dependency docs already listed. For current task, confirm `SEARCH_ADMIN_*` and `TYPESENSE_PRODUCTS_COLLECTION` values are present in `backend/services/search-service/.env`.

### Step 6: Run migrations if needed

No SQL migration is needed.

The service ensures the `products` collection schema at startup. Synonym rules are created through admin API calls.

### Step 7: Start backend service

For full local mode:

```bash
cd backend/services/search-service
set -a
. ./.env
set +a
go run ./cmd/server
```

For synonym-only local API testing without RabbitMQ consumer:

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

Create/update synonym:

```bash
curl -X POST "http://localhost:8085/api/v1/admin/search/synonyms" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: local-admin" \
  -H "X-Roles: superadmin" \
  -d '{"root":"mobile","synonyms":["phone","smartphone","cellphone"]}'
```

Expected response:

```json
{
  "synonym_id": "syn_mobile",
  "root": "mobile",
  "synonyms": ["phone", "smartphone", "cellphone"]
}
```

List synonyms:

```bash
curl "http://localhost:8085/api/v1/admin/search/synonyms?page=1&page_size=50" \
  -H "X-User-ID: local-admin" \
  -H "X-Roles: catalog_admin"
```

Expected response shape:

```json
{
  "synonyms": [
    {
      "synonym_id": "syn_mobile",
      "root": "mobile",
      "synonyms": ["phone", "smartphone", "cellphone"]
    }
  ]
}
```

## 10. Running the Project

### API contract

| Item | Value |
|---|---|
| Create/update method | `POST` |
| List method | `GET` |
| Path | `/api/v1/admin/search/synonyms` |
| Auth in current local code | Header based |
| Allowed roles | `catalog_admin`, `superadmin` |
| Create request | `{"root":"mobile","synonyms":["phone"]}` |
| Create response | `{"synonym_id":"syn_mobile","root":"mobile","synonyms":["phone"]}` |
| List query params | `page`, `page_size` |
| List defaults | `page=1`, `page_size=50` |
| Max list page size | `100` |
| API catalog | `api/master-api.json` |
| Current runnable transport | Go HTTP handler |

### Validation rules

| Rule | Current value | Why |
|---|---:|---|
| Root min length | `2` | Single-character roots are noisy |
| Root max length | `50` | Prevent accidental/abusive long terms |
| Synonym min count | `1` | Empty synonym rule has no value |
| Synonym max count | `20` | Large synonym sets can hurt relevance |
| Synonym term max length | `50` | Keep search terms concise |
| Duplicate terms | Blocked | Clean Typesense data |
| Root inside synonyms | Blocked | Avoid redundant mapping |
| Normalization | lowercase + trim + collapse spaces | Prevent duplicate variants |
| Allowed characters | `a-z`, `0-9`, space, `+`, `&`, `.`, `-` | Keep admin inputs safe and predictable |

### Deterministic id behavior

Same root creates same `synonym_id`, so repeated submit updates existing rule:

| Root | Generated id |
|---|---|
| `mobile` | `syn_mobile` |
| `smart phone` | `syn_smart_phone` |
| `t-shirt` | `syn_t_shirt` |

Beginner note: Agar admin same root dobara submit karega, duplicate record banane ke bajay same id par upsert hoga.

### Useful manual checks

Check invalid auth:

```bash
curl -i "http://localhost:8085/api/v1/admin/search/synonyms"
```

Expected: `401` with `UNAUTHENTICATED`.

Check forbidden role:

```bash
curl -i "http://localhost:8085/api/v1/admin/search/synonyms" \
  -H "X-User-ID: buyer-1" \
  -H "X-Roles: buyer"
```

Expected: `403` with `PERMISSION_DENIED`.

Check validation:

```bash
curl -i -X POST "http://localhost:8085/api/v1/admin/search/synonyms" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: local-admin" \
  -H "X-Roles: superadmin" \
  -d '{"root":"m","synonyms":["mobile"]}'
```

Expected: `400` with `INVALID_SYNONYM`.

## 11. Common Errors & Fixes

Generic Docker, Go module, Redis, RabbitMQ, and base Typesense errors are already documented in:

```md
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`9. Common Errors and Fixes`
```

Task-specific errors:

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `401 UNAUTHENTICATED` | Missing admin actor header | Send `X-User-ID` or trusted gateway actor header | Keep admin client/gateway forwarding actor identity |
| `403 PERMISSION_DENIED` | Role is not `catalog_admin` or `superadmin` | Send `X-Roles: catalog_admin` or `X-Roles: superadmin` locally | Keep RBAC mapping aligned with Auth/Gateway |
| `400 INVALID_SYNONYM` | Empty/short root, duplicate terms, root repeated in synonyms, unsupported chars, or too many terms | Fix request body according to validation rules | Add same validation hints in admin UI |
| `400 INVALID_JSON` | Bad JSON, unknown fields, or multiple JSON objects | Send one valid object with `root` and `synonyms` only | Use typed admin client/request schema |
| `400 INVALID_PAGINATION` | `page` or `page_size` is negative/non-integer/repeated | Use positive integers and one value per param | Keep query builder allowlisted |
| `429 RATE_LIMITED` | Same admin exceeded mutation limit | Wait for `SEARCH_ADMIN_MUTATION_RATE_WINDOW` or increase limit carefully | Avoid autosave loops in admin UI |
| `503 SEARCH_BACKEND_UNAVAILABLE` | Typesense down, wrong API key, wrong collection, or timeout | Check Typesense health, key, `TYPESENSE_PRODUCTS_COLLECTION`, logs | Start Typesense before service and keep secrets aligned |
| List response is empty | No synonyms have been created in current Typesense data volume | Create a synonym through admin API or direct local Typesense cURL | Seed local demo synonyms when needed |
| Synonym saved but search result unchanged | Product documents do not contain expected terms, collection alias mismatch, or Typesense behavior is one-way | Verify `TYPESENSE_PRODUCTS_COLLECTION`, indexed products, and test both root/alternate queries | Run search checks after every admin change |
| Local run cannot resolve `typesense` | Docker service DNS name used from host process | Use `TYPESENSE_HOST=localhost` for host `go run` | Keep separate host `.env` and Compose env |

## 12. Security & Best Practices

### Task-specific security notes

| Area | Best practice |
|---|---|
| Admin route exposure | Never expose header-based admin auth directly to public internet. Put it behind API Gateway/JWT/RBAC. |
| Allowed roles | Only `catalog_admin` and `superadmin` should manage platform-wide synonyms. |
| Typesense API key | Keep server-side only. Browser or frontend bundle me key expose mat karo. |
| Local auth bypass | `SEARCH_ADMIN_AUTH_ENABLED=false` sirf private local testing ke liye use karo. |
| Audit logging | Every write should log actor id, role, request id, action, resource id, root, and synonym count. |
| Input hygiene | Admin input trusted mat maan lo. Current validation blocks bad terms before Typesense write. |
| Rate limiting | Keep mutation rate limit enabled so broken UI/autosave loops cannot spam Typesense. |
| Search relevance review | Large synonym changes ko catalog/search owner review kare. Galat synonyms buyer relevance hurt kar sakte hain. |

### Existing security topics reused

```md
Admin header trust and local auth risks:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`
Section `10. Security and Configuration Audit`

Admin RBAC/rate limit principle:
`docs/06-auth-security.md`
Sections around admin permissions and admin mutations
```

## 13. Missing or Misconfigured Things

| Finding | Impact | Suggested fix |
|---|---|---|
| No generated proto/gRPC implementation is present in this repo snapshot | `api/master-api.json` has `CreateSynonym` and `ListSynonyms`, but current runnable implementation is HTTP | Add proto files, generated clients, and gRPC server/gateway wiring when internal RPC layer is implemented |
| Header-based admin auth is local/service-level only | Direct exposure could allow spoofed admin headers | Keep endpoint behind API Gateway/Auth middleware and strip incoming client-controlled admin headers |
| Audit is structured log only | No durable admin audit table write from this service | Forward audit metadata to Superadmin/Admin Audit service when that service is wired |
| No delete synonym endpoint | Admin can create/update/list only | Add explicit delete API later if product/search owners require it |
| No bulk import/export | Large synonym catalogs need manual repeated calls | Add reviewed bulk workflow later with validation report and audit trail |
| Typesense direct cURL can bypass app validation | Local debugging may create inconsistent synonym rules | Prefer admin API for normal writes; reserve direct cURL for cleanup/debug |
| `/healthz` is shallow | It does not prove Typesense synonym API is reachable | Add readiness check or admin diagnostics endpoint later |
| Compose stack starts infra only | Beginners may expect backend API to start with Compose | Run Go service separately or add backend service image/Dockerfile later |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `2. Go Dependency System` | Same Go module setup and commands already documented |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Environment Variables` | Full `.env` explanation already includes admin and shared service variables |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. Docker and DevOps Setup` | Same Compose stack, containers, volumes, and network reused |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `8. Project Run Instructions` | Same clone/install/start baseline reused |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `9. Common Errors and Fixes` | Generic setup, port, Redis, RabbitMQ, and Typesense errors already covered |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `10. Security and Configuration Audit` | Admin header trust and local auth bypass risk already documented |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `4.1 Typesense` | Same Typesense local install, Docker, health check, and credentials reused |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `6. Redis / Queue / External Services` | Same RabbitMQ/indexer setup reused when indexer is enabled |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | Search API behavior and `products` collection dependency | Synonyms affect existing buyer search over the same collection |
| `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md` | Existing task-specific style and no-proto snapshot note | Same repo snapshot constraint applies to admin synonym contract |

## 15. Final Checklist

- [ ] Previous dependency documentation checked before following this file
- [ ] No duplicate base Go/Docker/Typesense/Redis/RabbitMQ setup copied
- [ ] Typesense container is running and healthy
- [ ] `TYPESENSE_API_KEY` matches local Typesense container
- [ ] `TYPESENSE_PRODUCTS_COLLECTION` points to the active `products` collection or alias
- [ ] Redis is running for current service startup
- [ ] RabbitMQ is running or `SEARCH_INDEXER_ENABLED=false` is set for API-only local testing
- [ ] `SEARCH_ADMIN_AUTH_ENABLED=true` is used except for private local testing
- [ ] Admin request includes `X-User-ID`
- [ ] Admin request includes `X-Roles: superadmin` or `X-Roles: catalog_admin`
- [ ] Create/update synonym API returns `synonym_id`
- [ ] List synonym API returns saved synonym
- [ ] Invalid synonym input returns `400 INVALID_SYNONYM`
- [ ] Unauthorized/forbidden checks return `401` or `403`
- [ ] Logs contain `search.admin.audit` for synonym upsert
- [ ] Direct Typesense cURL is used only for local debugging
- [ ] No real secrets are committed to example env files
