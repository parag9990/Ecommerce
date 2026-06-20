# Superadmin Panel Task 3 - Dependency & Setup Guide

## 1. Project Overview

Yeh guide `task3.md` ke **Seller Management** module ko locally validate aur eventually full stack me run karne ke liye dependency, environment, database, external-service, networking, aur troubleshooting requirements explain karti hai. Business logic aur implementation snippets yahan repeat nahi kiye gaye hain.

Actual frontend implementation yahan present hai:

```text
frontend/superadmin-panel/src/features/sellers/
frontend/superadmin-panel/tests/sellers/
```

Module me seller list/search, seller review page, KYC document metadata, approve/reject/suspend/unsuspend actions, catalog view, permission checks, aur automated tests included hain.

### Current repository reality

| Capability | Current status | Beginner meaning |
|---|---|---|
| Seller frontend source | Ready | Routes, UI, hooks, API client, permissions, and tests present hain |
| Frontend install/test/build | Ready | Existing pnpm workspace se run hota hai |
| New Task 3 npm package | None | Koi extra frontend package install nahi karna |
| New frontend environment variable | None | Existing `.env.example` reuse hota hai |
| Seller list/status API contract | Documented only | Master contract me paths hain, runnable backend nahi mila |
| KYC document API contract | Missing | Frontend path use karta hai, master contract define nahi karta |
| Seller catalog admin API contract | Missing | Frontend path use karta hai, master contract define nahi karta |
| KYC object storage | Unconfigured | Docs object storage bolte hain, provider/bucket/credentials/config absent hain |
| Docker/Compose stack | Missing | Supported Dockerfile or Compose manifest nahi mila |

> **Important:** Frontend tests, typecheck, build, aur dev server backend ke bina run ho sakte hain. Real seller/KYC/catalog data ke liye missing backend runtimes and contracts complete karne honge.

### Task-specific API dependency

| Method | Frontend path | Contract status | Purpose |
|---|---|---|---|
| `GET` | `/api/v1/admin/sellers` | Present | Seller list/search/filter and current detail lookup |
| `PATCH` | `/api/v1/admin/sellers/{seller_id}/status` | Present | Approve, reject, suspend, or unsuspend with reason |
| `GET` | `/api/v1/admin/sellers/{seller_id}/kyc-documents` | Missing | KYC metadata and secure document links |
| `GET` | `/api/v1/admin/sellers/{seller_id}/catalog` | Missing | Seller-specific catalog moderation view |

All requests must go Browser -> API Gateway -> authorized backend. Browser ko database ya internal service directly call nahi karna chahiye.

## 2. Tech Stack

Task 3 Task 1/Task 2 ka same frontend stack reuse karta hai. Full beginner definitions and installation details dobara repeat nahi kiye gaye hain.

Refer:
`TaskImplementation/Superadmin Panel/task1_Dependency.md`

Sections:
`2. Tech Stack`, `3. Required Software`, and `4. Dependency Management`

Task-specific use:

| Technology/service | Task 3 me use | Required? | New for Task 3? |
|---|---|---:|---:|
| React 19 + TypeScript | Seller pages, KYC cards, action dialogs, catalog table | Yes | No |
| React Router | `/admin/sellers` and `/admin/sellers/:sellerId` | Yes | No |
| TanStack React Query | Seller/KYC/catalog queries and status mutation cache | Yes | No |
| Zustand | Admin access token and roles | Yes | No |
| Vite + Tailwind CSS | Dev server, build, and styling | Yes | No |
| Vitest + Testing Library + jsdom | Seller API, permission, table, and action tests | Tests only | No |
| API Gateway | Authenticated browser REST boundary | Real integration | No |
| MySQL | Seller/KYC metadata and admin audit/review data | Real integration | Same server, new Task 3 tables |
| MongoDB | Product catalog in `product_db` | Catalog integration | New Task 3 database usage |
| RabbitMQ | Product events under current Product env defaults | Conditional backend runtime | New |
| Private object storage/CDN | Actual KYC file objects | KYC document viewing | Required design, not configured |

### Planning guide versus actual manifest

`task3.md` suggests `clsx`, MSW, and Testing Library jest-dom install commands. Actual implementation imports neither `clsx` nor MSW, and `package.json` does not declare them. Project-local `cn` helpers and Vitest fetch mocks are used. Jest-dom is also not required by the current assertions.

> `task3.md` ke package-add commands mat run karein. Actual `package.json` and lockfile source of truth hain.

## 3. Required Software

Node.js, pnpm, Git, and browser setup unchanged hai.

Refer:
`TaskImplementation/Superadmin Panel/task1_Dependency.md`

Section:
`3. Required Software`

Task 3 modes:

| Mode | Additional software |
|---|---|
| Frontend tests/build/mock development | None |
| Seller/KYC API integration | API Gateway, Auth, Superadmin, User Service, MySQL, Redis |
| Catalog integration | Product Service and MongoDB; possibly CMS depending on approved orchestration |
| Product runtime with current event defaults | RabbitMQ |
| Real KYC document opening | Approved private object-storage provider/CDN |

No runnable Gateway, Auth, Superadmin, User, Product, or CMS source/module was found in the inspected backend folders; only `.env` files are present. Isliye exact backend `go run`/binary commands honestly provide nahi kiye ja sakte.

## 4. Dependency Management

### New dependency delta

| Check | Result |
|---|---|
| New runtime npm package | None |
| New development npm package | None |
| `package.json` update needed | No |
| `pnpm-lock.yaml` update needed | No |
| New Go module | No runnable Task 3 backend module present |

Existing install process reuse karein:

Refer:
`TaskImplementation/Superadmin Panel/task1_Dependency.md`

Section:
`4. Dependency Management`

Correct clean install:

```bash
cd frontend
pnpm install --frozen-lockfile
```

Important beginner rules:

- Command repository root se nahi, `frontend/` workspace se run karein.
- npm/yarn/pnpm lockfiles mix na karein.
- `node_modules/` commit na karein.
- `go mod tidy`, `go build`, or `go run` empty backend service folders me na chalayein.
- Tests package directory/config ke through run karein; repository root se raw Vitest run karne par jsdom config miss ho sakta hai.

Backend Go module limitation already documented hai:
`TaskImplementation/Superadmin Panel/Dependency/Go_Modules.md`

## 5. Database Setup

Frontend directly kisi database ko access nahi karta. DB credentials sirf owning backend services me rahenge.

### Task 3 data ownership

```text
Browser
  -> API Gateway
     -> Superadmin Service
        -> User Service      -> MySQL user_db
        -> Product/CMS path  -> MongoDB product_db
        -> MySQL superadmin_db (review/audit data)
```

| Store | Task 3 data | Required when | Default port | Status |
|---|---|---|---:|---|
| MySQL `user_db` | `seller_profiles`, `seller_kyc_documents` metadata | Seller and KYC APIs | `3306` | Schema present |
| MySQL `superadmin_db` | `admin_audit_logs`, `admin_review_tasks` | Mutations/review workflow | `3306` | Schema present |
| MongoDB `product_db` | `products`, categories, inventory snapshots | Catalog view | `27017` | Design present, runtime absent |

### MySQL setup reuse

MySQL installation, Docker option, credentials, start, and common errors already documented hain.

Refer:
`TaskImplementation/Superadmin Panel/Dependency/MySQL.md`

Sections:
`6. Installation steps` through `12. Common errors and fixes`

Task-specific verification, only after the backend owner approves `database/draw.sql` for this environment:

```bash
mysql -u root -p -e "USE user_db; SHOW TABLES LIKE 'seller_profiles'; SHOW TABLES LIKE 'seller_kyc_documents';"
mysql -u root -p -e "USE superadmin_db; SHOW TABLES LIKE 'admin_audit_logs'; SHOW TABLES LIKE 'admin_review_tasks';"
```

Expected: all four table names return. `seller_profiles.status` and `seller_kyc_documents.status` enums must stay aligned with the frontend types.

> **Migration warning:** Versioned Task 3 up/down migrations and a migration runner are absent. `database/draw.sql` is a repository-wide bootstrap, not an approved production migration.

### MongoDB `product_db` delta

MongoDB installation/start/Docker instructions already exist, so they are not duplicated.

Refer:
`TaskImplementation/Superadmin Panel/Dependency/MongoDB.md`

Sections:
`5. Installation steps` through `11. Common errors and fixes`

Task 3 specifically uses `product_db`, not the Task 2 `session_db`. Detected backend variables:

```dotenv
PRODUCT_MONGO_URI=mongodb://localhost:27017
PRODUCT_MONGO_DATABASE=product_db
PRODUCT_MONGO_AUTO_CREATE_COLLECTIONS=false
```

Because auto-create is disabled, the Product Service owner must provision collections/indexes from `database/mongodb-schema-design.md`. For a local approved instance:

```bash
mongosh product_db --eval 'db.products.createIndex({seller_id:1,status:1,updated_at:-1}); db.products.createIndex({category_id:1,status:1,updated_at:-1}); db.products.createIndex({"variants.sku":1},{unique:true}); db.products.createIndex({title:"text",description:"text",brand:"text"})'
mongosh product_db --eval 'db.products.getIndexes()'
```

Do not run schema/index commands against production without backup, change review, and duplicate-SKU checks.

## 6. Redis, Queue, and External Services

### API Gateway and authentication

Gateway/JWT/JWKS/CORS/Redis setup unchanged hai. Reuse:

- `TaskImplementation/Superadmin Panel/Dependency/API_Gateway.md`
- `TaskImplementation/Superadmin Panel/Dependency/Redis.md`
- `TaskImplementation/Superadmin Panel/task2_Dependency.md`, sections `6` and `7`

Task 3 needs Gateway to route seller endpoints to Superadmin on its configured `SUPERADMIN_GRPC_ADDR`. Server-side RBAC must distinguish seller view, KYC view, lifecycle mutation, and catalog access.

### RabbitMQ: new conditional dependency

RabbitMQ ek message broker hai—services events ko asynchronously publish/consume kar sakti hain. Browser RabbitMQ ko kabhi direct connect nahi karta.

Current Product Service env sets:

```dotenv
PRODUCT_EVENTS_ENABLED=true
PRODUCT_EVENTS_TOPIC=product.events
PRODUCT_EVENT_BROKER=rabbitmq
PRODUCT_OUTBOX_WORKER_ENABLED=true
RABBITMQ_URL=amqp://<user>:<password>@localhost:5672/
```

Therefore RabbitMQ **frontend-only validation ke liye optional**, but current default Product runtime ke event/outbox behavior ke liye likely required hai. Runnable Product code absent hone ki wajah se startup fail-open/fail-closed behavior verify nahi ho saka.

Local Docker setup, only if the backend team confirms RabbitMQ is required:

```bash
docker volume create ecommerce-rabbitmq-data
docker run --name ecommerce-rabbitmq \
  -e RABBITMQ_DEFAULT_USER=ecommerce_dev \
  -e RABBITMQ_DEFAULT_PASS='<strong-local-only-password>' \
  -p 127.0.0.1:5672:5672 \
  -p 127.0.0.1:15672:15672 \
  -v ecommerce-rabbitmq-data:/var/lib/rabbitmq \
  -d rabbitmq:3-management
```

Verify:

```bash
docker exec ecommerce-rabbitmq rabbitmq-diagnostics -q ping
curl -I http://127.0.0.1:15672
```

Use the same local user/password in `RABBITMQ_URL`; URL-encode special characters. Port `5672` is AMQP, while `15672` is the optional local management UI. Production me both ports public expose mat karein.

Kafka, NATS, and RabbitMQ alternatives ek saath setup karne ki zarurat nahi hai. Detected Product config specifically `rabbitmq` select karta hai.

### KYC object storage: required but missing

MySQL ka `storage_url` actual file nahi hai; woh object ka reference hai. Architecture docs KYC files ko object storage/CDN me rakhne ko kehte hain, but repository me S3/MinIO/provider, bucket, region, endpoint, credential, upload policy, encryption key, or signed-URL TTL config nahi mila.

Full KYC integration se pehle backend owner must define:

- Private bucket/container and region/endpoint.
- Backend-only credentials or workload identity.
- Encryption at rest and key ownership.
- Malware/content-type/size validation on upload.
- Short-lived signed `https://` download URLs or an authenticated proxy.
- Audit logging and document retention/deletion policy.
- Allowed origins/CSP without making the bucket public.

No frontend `.env` object-storage secret add karein. Current frontend sirf backend response ka `storage_url` open karta hai.

### CMS and Product services

Catalog UI ka frontend endpoint Superadmin namespace me hai, but master contract and Superadmin env do not define its Product/CMS downstream integration. CMS env has Product Service settings, yet Task 3 frontend directly CMS ko call nahi karta.

Backend team ko one approved ownership path choose karna hoga:

```text
Gateway -> Superadmin -> Product Service
```

or

```text
Gateway -> Superadmin -> CMS -> Product Service
```

Cross-service DB reads mat implement karein. Selected path contract, auth, timeout, health check, and environment variables me explicitly add hona chahiye.

## 7. Environment Variables

### Frontend delta

Task 3 introduces **no new or changed frontend environment variable**.

Existing `.env` creation and Vite security rules:

Refer:
`TaskImplementation/Superadmin Panel/task1_Dependency.md`

Section:
`7. Environment Variables`

Do not add MySQL, MongoDB, RabbitMQ, KYC storage, JWT signing, or internal service secrets to `frontend/superadmin-panel/.env`. Every `VITE_*` value browser bundle me visible hota hai.

### Existing API base-path blocker

Task 2 already documented the same unresolved mismatch:

Refer:
`TaskImplementation/Superadmin Panel/task2_Dependency.md`

Section:
`7. Environment Variables` -> `Critical base-path mismatch`

`VITE_API_BASE_URL` already ends with `/api/v1`, while Task 3 API constants also begin with `/api/v1`. Current example therefore produces:

```text
http://localhost:8080/api/v1/api/v1/admin/sellers
```

There is no safe Task 3-only `.env` workaround because changing the base also affects login. Recommended permanent convention: keep base `http://localhost:8080/api/v1` and make feature paths start with `/admin/...`; then cover login and all modules with URL integration tests.

### Backend-only Task 3 variables

| Owner | Detected variables | Purpose/status |
|---|---|---|
| Gateway | `SUPERADMIN_GRPC_ADDR`, `PRODUCT_GRPC_ADDR`, `CMS_GRPC_ADDR`, `JWT_*`, `REDIS_*` | Routing, auth, rate limits; source absent |
| Superadmin | `SUPERADMIN_DATABASE_DSN`, `USER_SERVICE_ADMIN_BASE_URL`, required/timeouts | Audit/User path exists in env; Product/CMS/KYC storage path missing |
| User | `USER_SERVICE_DATABASE_DSN`, DB pool variables | Seller/KYC metadata DB |
| Product | `PRODUCT_MONGO_*`, `PRODUCT_CMS_*`, `PRODUCT_EVENTS_*`, `RABBITMQ_URL` | Product catalog/runtime configuration |
| CMS | `CMS_PRODUCT_SERVICE_*`, `CMS_INTERNAL_AUTH_*`, `CMS_MYSQL_*` | Conditional orchestration; not wired into Task 3 contract |
| KYC storage | None found | Required configuration must be implemented before real document viewing |

Do not invent variable names in deployment manifests before the owning config loader exists. Add sanitized `.env.example` files, validate required values at startup, and inject real secrets through a secret manager.

## 8. Docker Setup, Ports, and Networking

Task 3 adds no repository Dockerfile, Compose service, network, health check, or restart policy. Existing frontend Docker limitations are documented in `task1_Dependency.md`, section `8. Docker Setup`.

The RabbitMQ one-off container in section 6 is the only new concrete Docker setup in this guide. MongoDB and MySQL Docker setup is referenced from their existing dependency files.

### Ports table

| Service | Port/address | Purpose | Task 3 status |
|---|---:|---|---|
| Vite frontend | `5173` | Admin UI | Reused; runnable |
| Vite preview | `4173` commonly | Built bundle preview | Reused; optional |
| API Gateway HTTP | `8080` | Browser REST entry | Reused; runtime missing |
| Auth/JWKS HTTP | `8081` in current Gateway config | Token/JWKS | Reused; runtime missing |
| Superadmin HTTP | `8088` | Env-only service HTTP | Reused; runtime missing |
| Superadmin gRPC | `50062` | Gateway seller workflow target | Reused; required |
| User gRPC | `50052` | Seller/KYC metadata owner | Task 3 required; runtime missing |
| Product gRPC | `50053` | Gateway-configured Product target | Task 3 catalog dependency; runtime missing |
| Product HTTP | `8082` in design/other service URLs | Product internal/public runtime | Task 3 conditional; no listener config found |
| CMS HTTP | `8087` | CMS internal HTTP | Conditional |
| CMS gRPC | `9098` | CMS internal gRPC | Conditional; differs from Gateway `50059` |
| MySQL | `3306` | Seller/KYC/audit schemas | Reused |
| MongoDB | `27017` | `product_db` | New Task 3 usage |
| Redis | `6379` | Gateway rate limiting | Reused |
| RabbitMQ AMQP | `5672` | Product events/outbox | New; conditional |
| RabbitMQ management | `15672` | Local-only diagnostics UI | New; optional |
| KYC object storage | TBD | Private document objects | Missing design |

> **Networking warning:** Gateway has `CMS_GRPC_ADDR=localhost:50059`, but CMS env declares `CMS_GRPC_ADDR=:9098`. Yeh mismatch resolve kiye bina that route connect nahi karega.

Docker/container network me `localhost` means the current container itself. Future Compose/Kubernetes config must use service DNS names, private networks, health/readiness probes, persistent DB/broker volumes, and secrets—not host-only addresses.

## 9. Local Development Setup

### Step 1: Read reused setup first

Follow `TaskImplementation/Superadmin Panel/task1_Dependency.md` for clone, Node/pnpm, frontend `.env`, and standard startup. Read Task 2 section 7 for the unresolved API-base issue.

### Step 2: Go to the workspace

```bash
cd Ecommerce/frontend
```

### Step 3: Install locked dependencies

```bash
pnpm install --frozen-lockfile
```

No Task 3 package-add command is needed.

### Step 4: Configure frontend environment

Create the existing local file as described in Task 1. Add no Task 3 variable. For real integration, fix the shared path convention in code/config before testing.

### Step 5: Choose validation mode

Frontend-only mode:

- Skip MySQL, MongoDB, Redis, RabbitMQ, and migrations.
- Use seller automated tests as the reliable validation boundary.

Full-stack mode:

- Obtain runnable Gateway/Auth/Superadmin/User/Product and optional CMS code.
- Apply reviewed MySQL migrations and Product MongoDB indexes.
- Configure private KYC object storage.
- Start Redis for current Gateway rate limits.
- Start RabbitMQ only if Product events/outbox stay enabled.
- Add and implement missing KYC/catalog contracts before calling those screens ready.

### Step 6: Run Task 3 tests

```bash
cd frontend/superadmin-panel
pnpm exec vitest run tests/sellers
```

### Step 7: Run quality checks

```bash
cd frontend
pnpm run superadmin:typecheck
pnpm run superadmin:test
pnpm run superadmin:build
```

## 10. Running the Project

Start the UI:

```bash
cd frontend
pnpm run superadmin:dev
```

Open:

```text
http://localhost:5173/admin/sellers
```

Without a stored valid admin session, route `/login` par redirect karega. Backend absent ho to UI real data fetch nahi karega; this is not an npm install failure.

### Verified frontend baseline

On 2026-06-20, current implementation produced:

| Check | Result |
|---|---|
| Seller tests | `4` files, `14` tests passed |
| TypeScript project build | Passed |
| Vite production build | Passed |
| Build note | Main JS chunk about `560 kB`; optimization warning only |

### Full-stack read-only verification

Use a dedicated non-production admin. Prefer a secure API client; avoid leaving tokens in shared shell history.

```bash
export ADMIN_TOKEN='<temporary-admin-access-token>'
export TEST_SELLER_ID='<approved-test-seller-id>'

curl -i \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "X-Request-ID: task3-seller-list-check" \
  'http://localhost:8080/api/v1/admin/sellers?status=pending_review&page=1&page_size=20'

curl -i \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "X-Request-ID: task3-kyc-check" \
  "http://localhost:8080/api/v1/admin/sellers/$TEST_SELLER_ID/kyc-documents"

curl -i \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "X-Request-ID: task3-catalog-check" \
  "http://localhost:8080/api/v1/admin/sellers/$TEST_SELLER_ID/catalog"
```

Last two calls currently require new approved API contracts/implementations. Lifecycle mutation QA only disposable seller par run karein and verify `admin_audit_logs` in the same acceptance flow.

### Expected frontend permissions

| Role | Seller view | KYC document view/action | Lifecycle mutation | Catalog view |
|---|---:|---:|---:|---:|
| `superadmin` | Yes | Yes | Yes | Yes |
| `operations_admin` | Yes | Yes | Yes | Yes |
| `catalog_admin` | Yes | No | No | Yes |
| `readonly_admin` | Yes | Masked | No | Yes |
| `finance_admin` | No | No | No | No |

Backend must enforce the same or a stricter approved matrix. Hidden frontend buttons are not authorization.

## 11. Common Errors & Fixes

Generic pnpm, Node, Docker, port, login, and CORS errors already covered hain:

Refer:
`TaskImplementation/Superadmin Panel/task1_Dependency.md`

Section:
`11. Common Errors & Fixes`

Task 3-specific issues:

| Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| URL contains `/api/v1/api/v1/` | Base and seller path both contain prefix | Apply one shared path convention; section 7 dekhein | Add config-level URL tests |
| Seller list/status returns `502` | Gateway cannot reach Superadmin `50062` | Supply/start runtime and check target/health | Readiness checks and service DNS |
| KYC or catalog returns `404` | Frontend endpoints absent from master contract/backend | Approve contract and implement routes | Contract-first CI validation |
| Wrong seller opens from detail route | Detail helper searches list and falls back to first result | Add exact `GET .../{seller_id}` endpoint; remove fallback | Exact-ID contract test |
| Seller cards show missing counts/status | Contract `SellerProfile` lacks frontend projection fields | Extend response schema and backend mapping | Keep generated/shared contract types |
| KYC link returns `403` later | Signed URL expired | Refetch metadata for a fresh short-lived URL | Never cache signed URL beyond TTL |
| KYC link exposes a public file | Bucket/CDN is public or URL is permanent | Make object private and issue authenticated signed URL | Bucket policy tests/security review |
| Catalog loads but product image breaks | Bad/blocked external `image_url` or CSP | Validate approved HTTPS host and placeholder fallback | Media proxy/allowlist and CSP tests |
| Catalog returns `502` | Superadmin-to-Product/CMS path absent or Mongo unavailable | Implement selected path; check `product_db` and indexes | Downstream readiness/timeout metrics |
| Product service cannot publish events | RabbitMQ down or `RABBITMQ_URL` wrong | Start/verify broker or use backend-approved disabled-event mode | Broker readiness and secret validation |
| RabbitMQ auth fails | URL credentials differ or password not URL-encoded | Correct secret and encode URI characters | Avoid hardcoded default credentials |
| Status mutation gets `403` | API role policy differs from frontend roles | Align granular server permissions | One authoritative permission matrix |
| Status changes without audit row | Mutation and audit are not atomic/implemented | Treat operation as failed; investigate logs/DB | Transactional mandatory audit write |
| Seller tests say `document/window is not defined` | Vitest launched from wrong cwd/config | Run from `frontend/superadmin-panel` or workspace script | Keep documented package command in CI |
| Build warns chunk over 500 kB | Admin app currently ships one large bundle | Add route-level lazy loading/code splitting | Track bundle size budget; warning is not failure |

## 12. Security & Best Practices

### Task-specific security rules

- KYC URLs must be short-lived, HTTPS, role-authorized, auditable, and backed by a private bucket.
- Backend should return KYC metadata only after permission check. Masking in React alone is insufficient.
- Allowlist document/image schemes and hosts. Frontend currently opens returned `storage_url` and loads returned `image_url`.
- Never place storage keys, RabbitMQ credentials, DSNs, internal auth tokens, or JWT secrets in `VITE_*` variables.
- KYC/GST/support email are sensitive. Return minimum fields per role and avoid logging full responses.
- `catalog_admin` must not receive KYC URLs; `finance_admin` must not inherit broad generic admin access.
- Approve/reject/suspend/unsuspend must validate allowed state transitions, structured reason code, bounded free-text reason, actor, and request ID server-side.
- Status mutation and immutable audit write should succeed/fail atomically.
- Use a disposable seller and synthetic documents/products for QA. Never test destructive actions on a real merchant.
- Do not put tokens or signed URLs in screenshots, tickets, analytics, browser logs, or shell history.
- Keep MySQL, MongoDB, Redis, RabbitMQ management, and object-storage administration private.

### Configuration audit

| Severity | Finding | Required action |
|---|---|---|
| Critical | Tracked backend `.env` files contain credential-like DSNs/passwords/tokens | Rotate any real/reused values; replace tracked files with sanitized examples and secret injection |
| Critical | KYC and catalog endpoints used by frontend are absent from API contract | Contract, authorize, implement, and integration-test before release |
| High | No object-storage provider/config/security policy exists for KYC | Select provider and implement private signed-URL/proxy flow |
| High | API base-path duplication breaks real Task 3 URLs | Normalize base/path convention across all modules |
| High | Backend `auth: admin` is broader/less precise than frontend Task 3 matrix | Add endpoint/action-specific server permissions |
| High | Superadmin env has no Product/CMS/KYC downstream config | Implement one explicit orchestration path with auth/timeouts/health |
| High | Superadmin required-dependency flags currently allow DB/User dependency to be optional | Fail closed for production seller mutations and audit requirements |
| High | No versioned Task 3 migrations or runnable service source | Add migrations, modules, start commands, and deployment artifacts |
| Medium | Seller detail can fall back to the first search result | Add exact detail endpoint and reject mismatched IDs |
| Medium | Contract response omits KYC/catalog counts and several UI fields | Version and document the complete admin projection |
| Medium | RabbitMQ credentials/config are in a tracked service `.env` | Use strong environment-specific secret and private broker connectivity |
| Medium | CMS gRPC configured ports disagree (`50059` vs `9098`) | Choose one endpoint and update both caller/server config |
| Medium | Product/CMS ownership for catalog moderation is unclear | Record one service owner and prohibit cross-DB reads |
| Low | Current build emits a large-chunk warning | Lazy-load admin modules and establish bundle budget |

### Task-specific operational best practices

- Add health endpoints for Gateway, Superadmin, User, Product/CMS, MySQL, MongoDB, Redis, RabbitMQ, and storage dependency checks.
- Distinguish liveness from readiness; broker or DB outage should not create misleading healthy status.
- Carry `X-Request-ID` through every service and audit row.
- Emit metrics for seller review latency, mutation failures, KYC signed-URL failures, downstream timeouts, and outbox backlog.
- Use least-privilege DB/bucket/broker identities per service and environment.
- Add API contract tests that use the real frontend base URL convention.
- Add route-level lazy loading for seller and other admin modules.

## 13. Missing or Misconfigured Things

Task 3 frontend setup ready hai, but complete production onboarding is blocked by these repository gaps:

1. Runnable Gateway/Auth/Superadmin/User/Product/CMS backends are absent.
2. KYC documents and catalog admin endpoints are not in `api/master-api.json`.
3. Dedicated seller detail and individual KYC/catalog moderation contracts are absent.
4. Seller response contract does not describe all fields consumed by the UI.
5. Superadmin has no Product/CMS/object-storage downstream configuration.
6. Private KYC object storage and signed-URL implementation are absent.
7. API base path is duplicated by current frontend config/path composition.
8. No versioned migrations or supported Compose/deployment stack exists.
9. Service credential-like values are tracked in `.env` files.
10. Gateway/CMS gRPC port configuration disagrees.
11. Product runtime event requirements cannot be verified without source.
12. Frontend catalog is currently a read-only table; no product moderation mutation exists.
13. Frontend status mutation treats seller approval as a whole-seller status change; per-document KYC decisions are not implemented.

These are application/platform prerequisites, not reasons to add random npm packages.

## 14. References to Previous Dependency Files

| Previous dependency file | Section/topic reused | Why reused |
|---|---|---|
| `task1_Dependency.md` | Tech stack, required software, pnpm, env creation, Docker limitation, run commands, generic errors | Same frontend application/toolchain |
| `task2_Dependency.md` | API base-path mismatch, Gateway/auth model, backend credential rules | Same unresolved shared HTTP configuration |
| `Dependency/Frontend.md` | Actual package inventory and package-guide mismatch | Task 3 adds no frontend dependency |
| `Dependency/Environment.md` | Frontend/backend env ownership | Same Vite and secret-placement model |
| `Dependency/API_Gateway.md` | Gateway auth/routing/Redis setup | Same browser REST entry point |
| `Dependency/MySQL.md` | MySQL installation, Docker, credentials, start, verification | Same MySQL server; Task 3 adds table-specific checks only |
| `Dependency/MongoDB.md` | MongoDB installation, Docker, start, health checks | Same MongoDB runtime; Task 3 specifically adds `product_db` context |
| `Dependency/Redis.md` | Redis install, variables, and health check | Same Gateway rate-limit dependency; no Task 3 Redis delta |
| `Dependency/Migrations.md` | Bootstrap-versus-migration warning | No Task 3 migration runner exists |
| `Dependency/gRPC.md` | Gateway-to-service communication and missing implementation | Same internal architecture |
| `Dependency/Protobuf.md` | Contract generation/versioning gap | New endpoints need approved versioned contracts |
| `Dependency/Go_Modules.md` | Missing backend modules/start commands | Backend runtime remains unavailable |
| `Dependency/main_dependency.md` | Overall topology and missing-runtime audit | Same service foundation |

## 15. Final Checklist

### Frontend-only readiness

- [ ] Previous dependency guides read
- [ ] Node.js 22.12+ and pnpm 10 available
- [ ] Workspace installed with `pnpm install --frozen-lockfile`
- [ ] No unnecessary `clsx`, MSW, or duplicate package added
- [ ] Existing frontend `.env` created without secrets
- [ ] Seller tests pass from the correct package context
- [ ] Full Superadmin typecheck passes
- [ ] Full Superadmin test suite passes
- [ ] Production build passes
- [ ] Large-bundle warning recorded as optimization work
- [ ] `/admin/sellers` and `/admin/sellers/:sellerId` routes open with an approved local session

### Full-stack readiness

- [ ] Shared `/api/v1` base-path convention fixed and integration-tested
- [ ] Runnable Gateway/Auth/Superadmin/User/Product runtimes supplied
- [ ] CMS inclusion/exclusion and catalog ownership explicitly decided
- [ ] Sanitized backend env templates supplied and tracked secrets rotated
- [ ] MySQL seller/KYC/audit tables applied through reviewed migrations
- [ ] MongoDB `product_db` collections/indexes applied and verified
- [ ] Redis reachable for Gateway rate limiting
- [ ] RabbitMQ requirement confirmed; broker healthy if events/outbox enabled
- [ ] KYC private object storage, encryption, retention, and signed URLs configured
- [ ] Missing KYC/catalog/detail contracts approved and implemented
- [ ] Complete seller admin response schema documented
- [ ] Gateway-to-CMS port mismatch resolved if CMS is used
- [ ] JWT/JWKS, exact CORS origin, and request headers verified
- [ ] Server RBAC matches approved seller/KYC/catalog matrix
- [ ] Superadmin fails closed when required DB/User/audit dependencies are unavailable
- [ ] Dedicated non-production admin and disposable seller provisioned
- [ ] Seller list/search/detail verified
- [ ] KYC masking and secure-link expiry verified per role
- [ ] Catalog data and image-host policy verified
- [ ] Approve/reject/suspend/unsuspend tested only on disposable data
- [ ] Mandatory audit row verified for every mutation
- [ ] Logs/metrics checked by request ID without leaking PII, tokens, or signed URLs
- [ ] No duplicate setup documentation added

Frontend Task 3 tooling is verified. Complete Seller Management readiness tabhi claim karein when missing backend contracts/runtimes, secure KYC storage, database migrations, role enforcement, and audit behavior are implemented and tested end to end.
