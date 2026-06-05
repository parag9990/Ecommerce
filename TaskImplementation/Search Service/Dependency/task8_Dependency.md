# Project Dependency & Setup Guide

## 1. Project Overview

| Variable | Value |
|---|---|
| `SERVICE_NAME` | Prompt variable from the assignment |
| `TASK_FILE_NAME` | Prompt variable from the assignment |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `{TASK_FILE_NAME without .md}_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

Current task ka focus **full catalog reindex job** hai. Simple Hinglish me: Product Service se complete searchable catalog page-by-page read hota hai, har product ko Typesense document me map kiya jata hai, new/indexed collection validate hoti hai, and safe case me `products` alias new collection par point kar diya jata hai.

Important boundaries:

- Original `TASK_FILE_NAME` modify nahi kiya gaya.
- Business logic rewrite nahi kiya gaya.
- Base Go, Docker, Typesense, Redis, RabbitMQ, Product Service, and broad `.env` setup duplicate nahi kiya gaya.
- Previous dependency docs ko pehle follow karna hai; yahan sirf `TASK_FILE_NAME` ke reindex-specific setup, verification, K8s job, and troubleshooting explain kiye gaye hain.

Current runnable flow in this repo snapshot:

```text
CLI or Admin API
  -> validate reindex request
  -> acquire Redis lock search:reindex:products:lock
  -> read Product Service search-export pages
  -> map ProductIndexPayload to ProductDocument
  -> import documents into Typesense target collection
  -> validate count and smoke search
  -> swap Typesense alias if mode=alias and dry_run=false
  -> release Redis lock and optionally cleanup old collections
```

Task-specific files inspected:

| File | Why it matters |
|---|---|
| `backend/services/search-service/cmd/reindex/main.go` | Local/K8s CLI entrypoint for full reindex |
| `backend/services/search-service/cmd/reindex-alias/main.go` | Manual alias rollback or alias move command |
| `backend/services/search-service/internal/config/config.go` | Loads `PRODUCT_SERVICE_SEARCH_EXPORT_*`, `SEARCH_REINDEX_*`, and import timeout env vars |
| `backend/services/search-service/internal/usecase/reindex_catalog.go` | Main workflow: lock, page export, import, validation, alias swap, cleanup |
| `backend/services/search-service/internal/usecase/reindex_starter.go` | Async in-process admin trigger support |
| `backend/services/search-service/internal/clients/product_export_client.go` | Calls Product Service search export endpoint |
| `backend/services/search-service/internal/repository/typesense_reindex_repository.go` | Creates collections, imports products, counts docs, smoke searches, swaps aliases |
| `backend/services/search-service/internal/repository/redis_reindex_lock.go` | Redis `SET NX` lock with refresh and safe release |
| `backend/services/search-service/internal/transport/http/handler.go` | `POST /api/v1/admin/search/reindex` route |
| `infra/k8s/jobs/search-reindex/` | Kubernetes batch job ConfigMap, Secret example, Job manifest, README |

## 2. Tech Stack

### Reused tech from previous dependency files

| Technology | Required for current task? | Reuse reference |
|---|---:|---|
| Go | Yes | `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Section `2. Go Dependency System` |
| Go modules | Yes | `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Section `2. Go Dependency System` |
| Typesense | Yes | `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`, Section `4.1 Typesense` |
| Redis | Yes | `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Section `3.2 Redis` |
| Docker Compose | Recommended locally | `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Section `7. Docker and DevOps Setup` |
| Kubernetes | Required for K8s job flow | `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Section `7. Docker and DevOps Setup` |
| Product Service | Yes | `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md`, Section `6. Redis / Queue / External Services` |
| RabbitMQ | Not required for the K8s reindex job | `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`, Section `6. Redis / Queue / External Services` |

### Task-specific tech details

| Component | Required? | Beginner-friendly explanation |
|---|---:|---|
| Reindex CLI | Yes | `go run ./cmd/reindex` ek standalone command hai. Ye HTTP server ke long request ke andar heavy kaam karne ke bajay batch flow run karta hai. |
| Reindex alias CLI | Recommended | `go run ./cmd/reindex-alias` alias rollback ke liye hai. Agar new index issue kare, alias old collection par wapas move kar sakte ho. |
| Product search export endpoint | Yes | Product Service source of truth hai. Reindex direct Product DB read nahi karta; Product Service se `/internal/v1/products/search-export` pages read karta hai. |
| Typesense bulk import | Yes | Documents ko Typesense me `import?action=upsert` style bulk path se write kiya jata hai. Ye one-by-one writes se faster hota hai. |
| Typesense alias swap | Yes for `alias` mode | Stable `products` alias buyer search ko point karta hai. New collection ready hone ke baad alias swap hota hai. |
| Redis lock heartbeat | Yes | Reindex heavy job hai. Redis lock duplicate full reindex avoid karta hai, aur heartbeat long-running job ke lock ko alive rakhta hai. |
| Kubernetes Job | Required for K8s ops flow | `batch/v1 Job` one-off pod run karta hai using `/search-reindex` command. |
| Admin trigger | Optional | `POST /api/v1/admin/search/reindex` async in-process reindex start karta hai. Local/admin use ke liye useful, production me K8s Job safer hai. |

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
| Go | Yes | Run CLI, server, and tests | `go version` |
| Typesense | Yes | Target index/alias lives here | `curl http://localhost:8108/health` |
| Redis | Yes | Reindex lock lives here | `docker exec ecommerce-redis redis-cli -a dev_redis_password ping` |
| Product Service | Yes | Provides paginated catalog export | Check `/internal/v1/products/search-export?limit=1` |
| Docker Compose | Recommended locally | Starts Typesense/Redis quickly | `docker compose version` |
| Kubernetes + kubectl | Required for Job manifest | Run `infra/k8s/jobs/search-reindex/job.yaml` | `kubectl version --client` |
| RabbitMQ | Only if normal server starts with indexer enabled | Existing product event consumer dependency | Disable with `SEARCH_INDEXER_ENABLED=false` for reindex-only process |

## 4. Dependency Management

This is a Go module project. Full Go module setup, common commands, proxy/cache issues, and version mismatch troubleshooting are already documented in:

```md
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`2. Go Dependency System`
```

Task-specific dependency audit:

| File / Dependency | Current finding | New install needed? |
|---|---|---:|
| `backend/services/search-service/go.mod` | Already includes `github.com/typesense/typesense-go/v2` | No |
| `github.com/rabbitmq/amqp091-go` | Already used by product indexer; not needed when reindex job disables indexer | No |
| Go standard library | Reindex uses `flag`, `context`, `log/slog`, `net/http`, `time` | No |
| Redis client | Existing custom client under repository package | No |
| Product export client | Local project code, not external package | No |

Practical command only if dependencies were not downloaded yet:

```bash
cd backend/services/search-service
go mod download
```

Common Go dependency errors are not repeated here. Follow the referenced Task 1 dependency guide.

## 5. Database Setup

### Typesense

Typesense setup, installation, Docker, port, API key, and Kubernetes StatefulSet details are already explained in:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`

Section:
`4.1 Typesense`
```

Task-specific Typesense behavior:

| Item | Value / Behavior |
|---|---|
| Stable alias | `TYPESENSE_PRODUCTS_COLLECTION`, default `products` |
| Versioned physical collection | `SEARCH_REINDEX_COLLECTION_PREFIX` plus UTC timestamp, example `products_20260605_120000` |
| Write path | Bulk import/upsert into target collection |
| Validation | Document count plus smoke search before alias swap |
| Rollback | Move `products` alias back to previous collection with `cmd/reindex-alias` |
| Cleanup | Old versioned collections older than `SEARCH_REINDEX_OLD_COLLECTION_RETENTION` can be deleted after successful swap |

Hinglish note: Typesense search engine hai. Current task me new collection pehle background me build hoti hai, fir alias swap hota hai. Isse buyer search old working index par chalta rehta hai jab tak new index validated na ho.

### Redis

Redis installation, Docker setup, password, port, verification, and common auth issues are already explained in:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`3.2 Redis`
```

Task-specific Redis behavior:

| Item | Value / Behavior |
|---|---|
| Lock key | `search:reindex:products:lock` |
| Acquire behavior | `SET NX` style single-owner lock |
| Refresh behavior | Lock heartbeat extends TTL while job is alive |
| Release behavior | Lua check releases only if lock value matches current job id |
| Required? | Yes for real reindex, because duplicate jobs can overload Product Service and Typesense |

### SQL migrations

No MySQL/PostgreSQL/MongoDB migration is introduced by `TASK_FILE_NAME`.

```md
This task does not create relational tables.
No `migrate up` step is required for the reindex job itself.
```

## 6. Redis / Queue / External Services

### Product Service export

This task adds a hard runtime dependency on Product Service search export.

| Topic | Detail |
|---|---|
| Base URL env | `PRODUCT_SERVICE_URL` |
| Export path env | `PRODUCT_SERVICE_SEARCH_EXPORT_PATH` |
| Default path | `/internal/v1/products/search-export` |
| Timeout env | `PRODUCT_SERVICE_SEARCH_EXPORT_TIMEOUT_MS` |
| Query params | `limit`, optional `cursor` |
| Expected response fields | `items` or `products`, `next_cursor`, `has_more`, optional `total` |
| Required? | Yes |

Expected minimal response shape:

```json
{
  "items": [],
  "next_cursor": "",
  "has_more": false,
  "total": 0
}
```

Important rules:

- `has_more=true` with empty `items` is treated as invalid, because it can cause an infinite loop.
- `has_more=true` must include a non-empty, advancing `next_cursor`.
- Each exported product must include enough fields to build `ProductDocument`.
- Product Service should expose only published/searchable products or include status/deleted fields so this service can skip invalid products.

### RabbitMQ / queues

RabbitMQ is not required for the standalone Kubernetes reindex job because `infra/k8s/jobs/search-reindex/configmap.yaml` sets:

```env
SEARCH_INDEXER_ENABLED=false
```

For the normal HTTP server, RabbitMQ may still be required if the product indexer is enabled. Reuse:

```md
`TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`

Section:
`6. Redis / Queue / External Services`
```

### Kubernetes

Task-specific manifests:

| Path | Purpose |
|---|---|
| `infra/k8s/jobs/namespace.yaml` | Creates/declares `jobs` namespace |
| `infra/k8s/jobs/search-reindex/configmap.yaml` | Non-secret reindex job config |
| `infra/k8s/jobs/search-reindex/secret.example.yaml` | Required secret keys example |
| `infra/k8s/jobs/search-reindex/job.yaml` | One-off `batch/v1 Job` running `/search-reindex` |
| `infra/k8s/jobs/search-reindex/README.md` | Short apply instructions |

## 7. Environment Variables

The full local `.env` and detailed variable explanations are already documented in:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`4. Environment Variables`
`Complete local .env example`
`Environment variable explanation`
```

Task 8 does not require repeating the full `.env`. Verify this task-specific subset only:

```env
# Product Service full-catalog export
PRODUCT_SERVICE_SEARCH_EXPORT_PATH=/internal/v1/products/search-export
PRODUCT_SERVICE_SEARCH_EXPORT_TIMEOUT_MS=2000

# Reindex behavior
SEARCH_REINDEX_MODE=alias
SEARCH_REINDEX_BATCH_SIZE=500
SEARCH_REINDEX_MAX_BATCH_SIZE=5000
SEARCH_REINDEX_LOCK_TTL=3h
SEARCH_REINDEX_JOB_TIMEOUT=2h
SEARCH_REINDEX_COLLECTION_PREFIX=products
SEARCH_REINDEX_OLD_COLLECTION_RETENTION=72h
TYPESENSE_REINDEX_IMPORT_TIMEOUT_MS=5000
```

### Variable explanation

| Variable | Required? | Purpose | Example | Security note |
|---|---:|---|---|---|
| `PRODUCT_SERVICE_SEARCH_EXPORT_PATH` | Yes | Product Service se paginated searchable catalog read karne ka endpoint | `/internal/v1/products/search-export` | Internal endpoint only |
| `PRODUCT_SERVICE_SEARCH_EXPORT_TIMEOUT_MS` | Optional | One Product Service page fetch ka timeout | `2000` | No secret |
| `SEARCH_REINDEX_MODE` | Optional | `alias` safest rebuild, `in_place` quick repair | `alias` | Keep `alias` in production |
| `SEARCH_REINDEX_BATCH_SIZE` | Optional | Per page products count | `500` | Too high value Product Service ko stress kar sakta hai |
| `SEARCH_REINDEX_MAX_BATCH_SIZE` | Optional | API/CLI max allowed batch size | `5000` | Keep bounded |
| `SEARCH_REINDEX_LOCK_TTL` | Optional | Redis lock expiry | `3h` | Must be longer than expected heartbeat gaps |
| `SEARCH_REINDEX_JOB_TIMEOUT` | Optional | Complete job deadline | `2h` | Tune for catalog size |
| `SEARCH_REINDEX_COLLECTION_PREFIX` | Optional | Versioned collection prefix | `products` | No secret |
| `SEARCH_REINDEX_OLD_COLLECTION_RETENTION` | Optional | Old collection cleanup age | `72h` | Set `0` to disable cleanup |
| `TYPESENSE_REINDEX_IMPORT_TIMEOUT_MS` | Optional | One Typesense import batch timeout | `5000` | No secret |

### K8s job-specific env differences

The Kubernetes Job intentionally disables unrelated server background behavior:

| Variable | K8s job value | Why |
|---|---|---|
| `SEARCH_INDEXER_ENABLED` | `false` | Reindex Job should not connect to RabbitMQ product consumer |
| `SEARCH_ZERO_RESULT_TRACKING_ENABLED` | `false` | Reindex Job does not serve buyer search analytics |

### Where credentials go

| Runtime | File / Place | Credentials |
|---|---|---|
| Local Compose infra | `infra/compose/.env.local` | `TYPESENSE_API_KEY`, `REDIS_PASSWORD`, RabbitMQ local creds |
| Local Go run | Shell env or local uncommitted service `.env` | `TYPESENSE_API_KEY`, `SEARCH_REDIS_PASSWORD` |
| K8s reindex job | `infra/k8s/jobs/search-reindex/secret.example.yaml` copied to real Secret | `TYPESENSE_API_KEY`, `SEARCH_REDIS_PASSWORD` |
| Production | Secret manager or sealed/encrypted Kubernetes Secret | Real API keys and passwords |

Important: current Go config reads `os.Getenv`. It does not automatically parse `.env` by itself. Use your shell, process manager, IDE run config, or a dotenv wrapper to load local env values before running Go commands.

## 8. Docker Setup

Base Docker Compose setup is already documented in:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`7. Docker and DevOps Setup`
```

Task-specific Docker findings:

| Item | Status |
|---|---|
| New local Compose container | None |
| Required local containers | Reuse `typesense` and `redis`; RabbitMQ only if normal indexer enabled |
| New volume | None |
| New Docker network | None |
| New health check | K8s Job uses existing service health indirectly; no new Compose health check |
| Service-specific Dockerfile | Not present in repo snapshot |

Local infra needed for reindex-only dry run:

```bash
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml up -d typesense redis
```

If you start the normal server and keep `SEARCH_INDEXER_ENABLED=true`, also start RabbitMQ:

```bash
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml up -d rabbitmq
```

Kubernetes reindex job files:

```bash
kubectl apply -f infra/k8s/jobs/namespace.yaml
kubectl apply -f infra/k8s/jobs/search-reindex/configmap.yaml
kubectl apply -f infra/k8s/jobs/search-reindex/job.yaml
```

Before real K8s run, create `search-reindex-secret` from real secret manager values. Do not apply the example secret with `replace-me` values to production.

## 9. Local Development Setup

### Step 1: Read previous dependency documentation first

Follow these before running Task 8:

```md
1. `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`
2. `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`
3. `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md`
```

Task 1 already contains base reindex env and command references, so this guide focuses on verification and operational differences.

### Step 2: Go to project directory

For implementation guide review:

```bash
cd "TaskImplementation/{SERVICE_NAME}"
```

For running code:

```bash
cd backend/services/search-service
```

### Step 3: Install only new dependencies if any

No new Go package is required for `TASK_FILE_NAME`.

If this is a fresh clone, use the standard Go module command:

```bash
go mod download
```

### Step 4: Setup only new databases/services if any

No new database or queue is introduced. Reuse Typesense and Redis from previous docs.

Task-specific service requirement:

```text
Product Service must expose:
GET /internal/v1/products/search-export?limit=500&cursor=<cursor>
```

### Step 5: Add only new or changed environment variables

If your local env was created from Task 1, reindex variables are already listed there. Just verify the subset from Section `7. Environment Variables`.

For K8s job, verify:

```text
infra/k8s/jobs/search-reindex/configmap.yaml
infra/k8s/jobs/search-reindex/secret.example.yaml
```

### Step 6: Run migrations if needed

No database migration is required for this task.

### Step 7: Start backend service

For CLI-only local reindex, you do not need to start the HTTP server. You do need Typesense, Redis, and Product Service.

For admin API trigger, start the normal server:

```bash
go run ./cmd/server
```

### Step 8: Verify API or functionality related to `TASK_FILE_NAME`

Dry run first:

```bash
go run ./cmd/reindex --dry-run --reason=local-check --actor-id=local-dev
```

Run real alias reindex only after Product Service export returns valid pages:

```bash
go run ./cmd/reindex --mode=alias --batch-size=500 --reason=local-seed --actor-id=local-dev
```

Manual alias rollback/move:

```bash
go run ./cmd/reindex-alias --alias=products --target=products_20260605_120000 --reason=rollback --actor-id=local-dev
```

Admin trigger smoke request:

```bash
curl -i -X POST http://localhost:8085/api/v1/admin/search/reindex \
  -H 'Content-Type: application/json' \
  -H 'X-User-ID: local-admin' \
  -H 'X-Roles: superadmin' \
  -d '{"mode":"alias","batch_size":500,"reason":"local-admin-dry-run","dry_run":true}'
```

Expected HTTP status:

```text
202 Accepted
```

## 10. Running the Project

### Reindex CLI flags

| Flag | Default source | Purpose |
|---|---|---|
| `--mode` | `SEARCH_REINDEX_MODE` | `alias` or `in_place` |
| `--batch-size` | `SEARCH_REINDEX_BATCH_SIZE` | Product export page size |
| `--reason` | `manual` | Audit/log reason |
| `--actor-id` | `cli` | Actor value in logs |
| `--dry-run` | `false` | Read/map/validate without writing or alias swap |
| `--target` | empty | Advanced target collection override |

### Ports and networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| `SERVICE_NAME` HTTP API | `8085` | Admin trigger and search APIs | Reused |
| Typesense API | `8108` | Collection create, import, alias, search validation | Reused |
| Redis | `6379` | Reindex lock | Reused |
| Product Service | `8082` | Search export source | Reused dependency |
| RabbitMQ AMQP | `5672` | Normal product indexer only | Not needed for K8s reindex job |
| RabbitMQ Management | `15672` | Queue UI | Not needed for K8s reindex job |

Host vs Docker network rule:

- Host `go run`: use `localhost` addresses.
- Kubernetes Job: use cluster DNS like `typesense.data.svc.cluster.local` and `product-service.core.svc.cluster.local`.
- Docker Compose service names like `typesense` work only inside the Compose network.

### Kubernetes run flow

1. Apply/create real secret first, not the example secret.
2. Apply job ConfigMap.
3. Apply Job manifest.
4. Watch pod logs.

Useful commands:

```bash
kubectl get jobs -n jobs
kubectl get pods -n jobs -l app.kubernetes.io/name=search-reindex
kubectl logs -n jobs job/search-reindex-products
```

If you need to rerun the same named Job, delete the completed Job first:

```bash
kubectl delete job search-reindex-products -n jobs
kubectl apply -f infra/k8s/jobs/search-reindex/job.yaml
```

## 11. Common Errors & Fixes

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `PRODUCT_SERVICE_SEARCH_EXPORT_PATH cannot be empty` | Env set to blank | Set `/internal/v1/products/search-export` | Keep Task 8 env subset in local run config |
| `product catalog export unavailable` | Product Service down, wrong URL/path, non-2xx response, or bad JSON | Check `PRODUCT_SERVICE_URL`, path, logs, and curl export endpoint | Add Product Service health check before reindex |
| `empty page with has_more=true` | Product export contract bug | Fix Product Service response pagination | Contract test export endpoint |
| `next_cursor is required when has_more=true` | Product export returned incomplete page metadata | Return a valid cursor | Use cursor-based pagination tests |
| `product export cursor did not advance` | Same cursor repeated | Fix Product Service cursor generation | Guard Product Service pagination |
| `search reindex already running` | Redis lock exists or admin starter already active | Wait for active job, inspect logs, then retry after lock expiry | Use one deployment path at a time |
| `SEARCH_REINDEX_BATCH_SIZE must be less than or equal to SEARCH_REINDEX_MAX_BATCH_SIZE` | Batch config invalid | Lower batch size or raise max carefully | Keep default `500` unless tested |
| `target collection would be empty` | Product export returned no searchable products or dry-run has no valid docs | Verify Product Service data/status fields | Seed products before real alias swap |
| `indexed count below required minimum` | Partial import/export or many invalid products | Inspect skipped product logs and Product Service total | Run dry-run and compare counts first |
| `smoke search returned no products` | Typesense import/index not searchable | Check schema fields, `query_by=title`, and import errors | Keep schema aligned with mapper |
| `swap alias ... failed` | Typesense unavailable or API key lacks permission | Check Typesense health and key | Validate backend access before swap |
| `--target is required` | Alias rollback command missing target collection | Pass `--target=products_YYYYMMDD_HHMMSS` | Copy previous collection from reindex logs |
| K8s `ImagePullBackOff` | Placeholder image `registry.example.com/search-service:latest` not replaced or unavailable | Publish real image and update Job manifest | Use CI-built immutable tag |
| K8s `CreateContainerConfigError` | Secret or ConfigMap missing | Create `search-reindex-secret` and apply ConfigMap | Apply manifests in documented order |
| Admin API `401` | Missing admin identity header | Add `X-User-ID` | Use local admin headers in smoke tests |
| Admin API `403` | Role is not `superadmin` or `catalog_admin` | Add `X-Roles: superadmin` or configure gateway mapping | Keep RBAC mapping documented |
| Admin API `409` | Async starter already active | Wait for current job to finish | Prefer K8s Job for production full rebuilds |

Generic Go/Docker/Typesense/Redis/RabbitMQ troubleshooting is already covered in previous dependency files, especially `task1_Dependency.md` and `task2_Dependency.md`.

## 12. Security & Best Practices

Task-specific security notes:

- Keep `SEARCH_REINDEX_MODE=alias` in production. Alias mode protects buyer search from half-built indexes.
- Always run `--dry-run` before a real reindex when Product Service export changed.
- Never log `TYPESENSE_API_KEY`, Redis password, request auth headers, or secret file contents.
- Keep Product Service export endpoint internal-only. Ye public internet par expose nahi hona chahiye.
- Add service-to-service authentication for Product export before production if gateway/network policy does not already enforce it.
- Use `SEARCH_REINDEX_BATCH_SIZE=500` as safe starting value. Very large values can overload Product Service and Typesense.
- Keep old collections for rollback. Do not set retention too low until rollback flow is tested.
- Use real Kubernetes Secret from secret manager, not `secret.example.yaml`.
- Use immutable image tags for the K8s Job instead of `latest` in production.
- Watch `search.reindex.progress`, `search.reindex.completed`, `search.reindex.failed`, and `search.reindex.alias_swapped` logs.
- Prefer K8s Job for production reindex. Admin API currently starts an in-process goroutine, so process restart can interrupt it.

Existing broad security topics reused:

```md
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`
Section: `10. Security and Configuration Audit`

`TaskImplementation/{SERVICE_NAME}/task6_Dependency.md`
Section: `12. Security & Best Practices`
```

## 13. Missing or Misconfigured Things

| Finding | Impact | Recommended fix |
|---|---|---|
| `infra/k8s/jobs/search-reindex/job.yaml` uses placeholder image `registry.example.com/search-service:latest` | K8s Job will fail unless image exists | Replace with real registry/image tag from CI |
| `secret.example.yaml` contains `replace-me` values | Job cannot authenticate to Typesense/Redis | Create real `search-reindex-secret` from secret manager |
| No service-specific Dockerfile found in repo snapshot | Building `/search-reindex` image is not fully documented in code | Add Dockerfile or central build pipeline that emits `cmd/server`, `cmd/reindex`, and `cmd/reindex-alias` binaries |
| Admin reindex trigger runs in-process async | Server restart kills active admin-triggered reindex | For production, make admin route create a K8s Job or enqueue durable work |
| Product export client sends no service auth token | Internal endpoint relies on network trust/header injection elsewhere | Add service-to-service auth env/config before production |
| `infra/compose/.env.local.example` does not include Task 8 reindex subset | Local developers may miss export/reindex env when using only compose env | Keep local service env aligned with Task 1 full `.env` section |
| K8s Job applies ConfigMap and Job, README says create secret separately | Easy for beginner to forget secret | Add secret creation command in team runbook using real secret manager |
| Product Service implementation is outside this task | Reindex fails if export endpoint not implemented | Product Service team must provide/export contract and test data |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `2. Go Dependency System` | Go module setup, build/run commands, and common dependency failures already documented |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3.2 Redis` | Redis install, Docker setup, password, port, and verification already documented |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Environment Variables` | Full local `SERVICE_NAME` `.env`, including reindex variables, already documented |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. Docker and DevOps Setup` | Compose startup, K8s base paths, logs, volumes, and Docker cautions already documented |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `8. Project Run Instructions` | Fresh clone onboarding and base service start flow already documented |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `9. Common Errors and Fixes` | Generic Go/Docker/Typesense/Redis/RabbitMQ/Product errors already documented |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `4.1 Typesense` | Typesense installation, ports, credentials, local Compose, and K8s setup already documented |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `8. Docker and DevOps Setup` | Typesense-specific Docker/K8s operations already documented |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `6. Redis / Queue / External Services` | RabbitMQ product event topology already documented; not repeated because K8s reindex job disables it |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | `6. Redis / Queue / External Services` | Product Service dependency pattern already documented; Task 8 only adds search export path specifics |
| `TaskImplementation/{SERVICE_NAME}/task6_Dependency.md` | `12. Security & Best Practices` | Admin/RBAC mutation security pattern already documented |
| `TaskImplementation/{SERVICE_NAME}/task7_Dependency.md` | `7. Environment Variables` | `SEARCH_ZERO_RESULT_TRACKING_ENABLED=false` K8s job difference builds on previous zero-result config |

## 15. Final Checklist

- [ ] Previous dependency documentation checked first.
- [ ] No duplicate Go/Typesense/Redis/RabbitMQ installation docs added.
- [ ] Typesense is running and `TYPESENSE_API_KEY` is valid.
- [ ] Redis is running and `SEARCH_REDIS_PASSWORD` matches local/K8s secret.
- [ ] Product Service search export endpoint is reachable.
- [ ] `PRODUCT_SERVICE_SEARCH_EXPORT_PATH` and timeout are configured.
- [ ] `SEARCH_REINDEX_MODE=alias` for production-like runs.
- [ ] `SEARCH_REINDEX_BATCH_SIZE` is positive and not above max.
- [ ] `SEARCH_REINDEX_LOCK_TTL` and `SEARCH_REINDEX_JOB_TIMEOUT` fit expected catalog size.
- [ ] Dry-run reindex completed before real alias swap.
- [ ] Real reindex logs show `search.reindex.completed`.
- [ ] Typesense alias points to expected collection after swap.
- [ ] Rollback target collection is known before deleting old collections.
- [ ] K8s `search-reindex-secret` uses real secret values, not examples.
- [ ] K8s Job image points to a real built image.
- [ ] Admin API trigger tested with proper `X-User-ID` and role headers if used.
- [ ] No real secrets committed to examples or task documentation.
