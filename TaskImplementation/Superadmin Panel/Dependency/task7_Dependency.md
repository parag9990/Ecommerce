# Superadmin Panel Task 7 - Dependency & Setup Guide

## 1. Project Overview

Yeh handbook **Platform Settings** module ko locally install, run, test, debug, aur future full-stack environment me connect karne ke liye dependency and setup requirements explain karti hai. Business logic aur implementation snippets yahan repeat nahi kiye gaye hain.

Actual frontend implementation:

```text
frontend/superadmin-panel/src/features/settings/
frontend/superadmin-panel/tests/settings/
```

Module me commission configuration, category overrides, search synonym CRUD UI, feature flags, maintenance mode, reason-required confirmation, role checks, React Query cache invalidation, aur automated tests present hain.

### Current repository reality

| Capability | Current status | Beginner meaning |
|---|---|---|
| Task 7 frontend source | Ready | Route, components, API helpers, hooks, validators, permissions, and tests present hain |
| Frontend install/test/build | Ready | Existing pnpm workspace and lockfile se run hota hai |
| New Task 7 npm dependency | None | Koi package add karne ki zarurat nahi hai |
| New frontend environment variable | None | Existing frontend template reuse hota hai |
| Platform settings REST contract | Partial but documented | List/update routes master API contract me hain; runnable backend nahi hai |
| Search synonym REST contract | Incomplete | Frontend CRUD routes use karta hai, but master REST contract me routes registered nahi hain |
| Superadmin/Search backend runtime | Missing | Dono service folders me ignored local `.env` files hain, runnable source/modules nahi |
| MySQL schema | Bootstrap schema present | `platform_settings` and `admin_audit_logs` tables exist; versioned migrations/seeds absent hain |
| Redis settings cache | Designed but not wired | Five-minute TTL and invalidation design present hai; runnable cache integration/config incomplete hai |
| Typesense synonym store | Config/design only | Search env and local design exist; Search Service implementation absent hai |
| Docker/Compose stack | Missing | Supported project Dockerfile or Compose manifest nahi mila |

> **Important:** Task 7 tests, typecheck, build, aur frontend dev server backend ke bina run ho sakte hain. Real settings read/write aur search behavior ke liye Auth, API Gateway, Superadmin Service, Search Service, MySQL, Redis/cache wiring, and Typesense required hain.

### Task-specific API dependencies

| Method | Frontend path | Contract evidence | Expected owner | Purpose |
|---|---|---|---|---|
| `GET` | `/api/v1/admin/settings` | Present in `api/master-api.json` | Superadmin Service | Settings list |
| `PATCH` | `/api/v1/admin/settings/{key}` | Present in `api/master-api.json` | Superadmin Service | One setting update with reason |
| `GET` | `/api/v1/admin/search/synonyms` | Missing REST route | Search Service | Synonym list |
| `POST` | `/api/v1/admin/search/synonyms` | Mentioned in service design, missing master REST route | Search Service | Create synonym |
| `PATCH` | `/api/v1/admin/search/synonyms/{synonym_id}` | Missing | Search Service | Update synonym |
| `DELETE` | `/api/v1/admin/search/synonyms/{synonym_id}` | Missing | Search Service | Delete synonym |

The master contract only declares Search gRPC methods `CreateSynonym` and `ListSynonyms`. Update/delete RPCs and all corresponding Gateway REST mappings are still required.

Safe intended flow:

```text
Admin browser
  -> API Gateway (JWT, RBAC, validation, rate limit)
     |-> Superadmin Service -> MySQL superadmin_db
     |                     -> settings cache invalidation/audit
     `-> Search Service     -> Typesense synonyms
```

Browser ko MySQL, Redis, Typesense, gRPC ports, or backend credentials directly access nahi karne chahiye.

## 2. Tech Stack

Core frontend stack Task 1 jaisa hi hai. Full beginner definitions and installation steps repeat nahi kiye gaye hain.

Refer:
`TaskImplementation/Superadmin Panel/task1_Dependency.md`

Sections:
`2. Tech Stack`, `3. Required Software`, and `4. Dependency Management`

Task-specific mapping:

| Technology/service | Task 7 me use | Required? | New for this task? |
|---|---|---:|---:|
| React 19 + TypeScript | Settings tabs, forms, safe value shapes, permission states | Yes | No |
| React Router | Protected `/admin/settings` route | Yes | No |
| TanStack React Query | Settings/synonym queries, mutations, invalidation | Yes | No |
| Zustand | Admin token and roles | Yes | No |
| Vite + Tailwind CSS | Local server, build, and styling | Yes | No |
| Lucide React | Settings UI icons | Yes | No |
| Vitest + Testing Library + jsdom | API, validator, permission, and component tests | Tests only | No |
| API Gateway + Auth/JWKS | Authenticated REST boundary | Real integration | Reused |
| Superadmin Service | Settings persistence, validation, audit, invalidation | Real integration | Task 7 core |
| Search Service | Synonym CRUD and Typesense synchronization | Real integration | Task 7 core |
| MySQL 8 | Structured `platform_settings` and audit rows | Real integration | New Task 7 table usage |
| Redis/cache layer | Five-minute platform-setting cache/invalidation | Real integration | New Task 7 use |
| Typesense | Search index and collection-scoped synonyms | Real synonym behavior | New external service |
| RabbitMQ | Product index event consumer under current Search env | Conditional | Reused from Task 3 |

**Simple Hinglish:** MySQL durable settings and audit history store karega. Redis frequently read settings ko short time ke liye fast cache kar sakta hai. Typesense search engine hai jahan synonym changes actual search ranking/matching ko affect karenge. Frontend in services ko direct call nahi karta.

### Planning guide versus actual package manifest

`task7.md` examples `zod`, `clsx`, MSW, and several package-add commands suggest karte hain. Actual implementation and manifest ke according:

- `zod` installed/imported nahi hai; validators project-local TypeScript functions hain.
- `clsx` installed/imported nahi hai; `src/lib/classnames.ts` ka local `cn` helper use hota hai.
- MSW installed/imported nahi hai; tests Vitest fetch mocks use karte hain.
- React, Query, Router, Zustand, Lucide, Vitest, Testing Library, and `user-event` already locked hain.

> `task7.md` ke package-add commands current implementation par mat run karein. `package.json`, imports, and `pnpm-lock.yaml` source of truth hain.

## 3. Required Software

Git, Node.js 22.12+, pnpm 10, and modern browser setup unchanged hai.

Refer:
`TaskImplementation/Superadmin Panel/task1_Dependency.md`

Section:
`3. Required Software`

| Validation mode | Required software/services |
|---|---|
| Settings tests/typecheck/build | Node.js, pnpm, locked workspace packages |
| Frontend dev page/error states | Above plus modern browser |
| Real platform settings | API Gateway, Auth/JWKS, Superadmin Service, MySQL, and working cache/invalidation policy |
| Real search synonyms | Above plus Search Service and Typesense |
| Search Service exactly as current env intends | Typesense plus Redis; RabbitMQ may also be required because indexer is enabled |
| Production | Above plus TLS/mTLS, secret manager, granular RBAC, migrations, audit monitoring, backups, and private networking |

MongoDB, Kafka, NATS, SMTP, payment provider, and object storage are not direct Task 7 dependencies.

## 4. Dependency Management

### New dependency delta

| Check | Result |
|---|---|
| New runtime npm package | None |
| New development npm package | None |
| `package.json` update needed | No |
| `pnpm-lock.yaml` update needed | No |
| New frontend install command | No |
| Superadmin/Search backend Go modules | Missing; backend cannot currently install/build |

Existing pnpm concepts, exact locked-install command, scripts, cache/proxy issues, and Node version fixes are already documented in `task1_Dependency.md`, section `4. Dependency Management`. Follow that section unchanged; Task 7 has no additional install command.

Beginner rules:

- npm, yarn, and pnpm lockfiles mix mat karein.
- `node_modules/` commit mat karein.
- Sirf planning guide me package name dekhkar `pnpm add` mat chalayein.
- Missing backend folder me `go mod tidy` or `go run` backend create nahi karega.
- Future backend module setup ke liye `Dependency/Go_Modules.md` refer karein.

## 5. Database Setup

Frontend ka direct database dependency **none** hai. Database credential ko `VITE_*` variable me rakhna security leak hoga.

### MySQL `superadmin_db`

MySQL ek relational database hai. Task 7 full-stack mode me mandatory hai because platform setting values and immutable admin audit records durable storage maangte hain.

Generic MySQL installation, Docker example, start, credentials, and common errors already documented hain.

Refer:
`TaskImplementation/Superadmin Panel/Dependency/MySQL.md`

Sections:
`6. Installation steps` through `13. Security notes`

Task-specific schema:

| Table | Important columns | Task 7 use |
|---|---|---|
| `platform_settings` | `setting_key`, JSON `setting_value`, `updated_by`, `updated_at` | Commission, feature flags, maintenance values |
| `admin_audit_logs` | actor, action, resource, reason, before/after JSON, request ID | Every settings/synonym mutation audit |

`setting_key` has a unique index named `uk_platform_settings_key`, so one active row per setting key expected hai.

Task 7 frontend recognizes only:

| Setting key | Expected JSON shape |
|---|---|
| `commission.default_rate` | `{ "rate_percent": number, "applies_to": "all_sellers" }` |
| `commission.category_overrides` | `{ "overrides": [{ "category_id": string, "rate_percent": number }] }` |
| `platform.feature_flags` | `{ "flags": { "lowercase_key": boolean } }` |
| `platform.maintenance_mode` | `{ "enabled": boolean, "message": string, "starts_at": string|null, "ends_at": string|null, "allow_admin_bypass": boolean }` |

After an approved local bootstrap/migration, verify without exposing credentials in command output:

```bash
mysql -u root -p -e "USE superadmin_db; SHOW TABLES LIKE 'platform_settings'; SHOW TABLES LIKE 'admin_audit_logs';"
mysql -u root -p -e "USE superadmin_db; SHOW INDEX FROM platform_settings;"
mysql -u root -p -e "USE superadmin_db; SELECT setting_key, updated_at FROM platform_settings ORDER BY setting_key;"
```

Expected setting query me four supported keys eventually present hone chahiye. Current repository me seed `INSERT` statements nahi mile, so empty table par UI safe defaults and “No platform settings returned” state show karega.

### Migration and seed status

- Versioned Task 7 up/down migrations absent hain.
- No migration runner exists for Superadmin Service.
- `database/draw.sql` repository-wide bootstrap hai, safe production migration command nahi.
- No reviewed default-setting seed exists.
- Existing schema and public API do not provide an optimistic-lock `version`; frontend optional `version` accept karta hai, but concurrent admin update protection proven nahi hai.

Do not invent production defaults in ad-hoc SQL. Commission and maintenance defaults business/security owners approve karenge, then an idempotent versioned migration or seed job add hoga.

Migration safety already documented hai in:
`TaskImplementation/Superadmin Panel/Dependency/Migrations.md`

### Search synonym persistence

Design Typesense ko search index/synonym owner batata hai. Optional MySQL/Mongo config store ka mention hai, but current repository me synonym-specific table, collection, or migration nahi mila.

Therefore:

- Typesense ko current real synonym behavior ke liye mandatory treat karein.
- Search Service ko source-of-truth, restart recovery, and sync semantics explicitly define karne honge.
- Browser ko Typesense API key kabhi nahi milni chahiye.
- Product collection/synonym creation manually UI se bypass karke na karein; owning Search Service schema manage kare.

## 6. Redis, Queue, and External Services

### Redis and platform-setting cache

Redis ek in-memory store hai. Generic install, Docker example, `PONG` check, and security steps already documented hain:

`TaskImplementation/Superadmin Panel/Dependency/Redis.md`, sections `5` through `12`

Task 7 design delta:

| Setting | Detected value | Meaning |
|---|---|---|
| Platform settings cache TTL | `5m` | Cached value maximum five minutes old ho sakta hai |
| Update event topic | `platform.settings.updated` | Consumers should invalidate/refresh after update |
| Gateway Redis | `localhost:6379`, DB `0` | Current Gateway fail-closed rate limiter |

Important gap: Superadmin env me cache TTL/topic hain, but Superadmin-specific Redis address or broker configuration nahi mila. Runnable code bhi absent hai. Isliye cache write/invalidation/event publishing currently verify nahi ho sakta.

Expected mutation order:

```text
validate -> authorize -> DB update + audit -> invalidate/publish -> return success
```

High-risk setting ko successful report mat karein if durable update/audit fails. Cache/event partial failure ke liye retry/outbox or another approved recovery mechanism chahiye.

### Typesense: new Task 7 external service

Typesense fast search engine hai. Search synonyms isme product collection ke search terms ko equivalent terms se map karte hain. Frontend-only validation me optional hai; real synonym behavior me mandatory hai.

Repository me supported Compose file nahi hai. Existing design ke basis par local one-off Docker setup:

```bash
docker volume create ecommerce-typesense-data

docker run -d \
  --name ecommerce-typesense \
  -p 127.0.0.1:8108:8108 \
  -v ecommerce-typesense-data:/data \
  typesense/typesense:latest \
  --data-dir /data \
  --api-key local-dev-only-change-me
```

This command local development example hai. CI/staging/production me reviewed immutable image tag pin karein; floating `latest` use na karein. Throwaway key ko production me reuse mat karein.

Verify health:

```bash
curl -s http://localhost:8108/health
```

Expected healthy response:

```json
{"ok":true}
```

Authenticated diagnostic, same local key use karke:

```bash
export TYPESENSE_API_KEY='<local-development-key>'
curl -s \
  -H "X-TYPESENSE-API-KEY: $TYPESENSE_API_KEY" \
  http://localhost:8108/collections
```

Stop/start:

```bash
docker stop ecommerce-typesense
docker start ecommerce-typesense
docker logs ecommerce-typesense
```

Security rules:

- Port `8108` ko local/private network par bind karein; public internet par expose na karein.
- Search-only and admin keys ko least privilege se separate karein where supported.
- Admin API key backend secret manager me rahe; `VITE_*`, Git, screenshots, or logs me nahi.
- Browser Typesense ko direct call nahi karta, so broad CORS enable karna unnecessary hai.
- Production data volume encryption, backup/restore, replicas, and key rotation test karein.

### Search Service and internal communication

Current intent:

| Item | Detected config |
|---|---|
| Search HTTP listener | `:8085` |
| Gateway Search gRPC target | `localhost:50058` |
| Typesense target | `localhost:8108`, HTTP |
| Admin auth | Enabled |
| Admin mutation limit | 30 per minute over a one-minute window |

Search Service source, `go.mod`, gRPC listener variable/server, health endpoints, and synonym handlers absent hain. HTTP `8085` healthy hone se Gateway gRPC `50058` automatically healthy nahi hoga.

Generic internal gRPC/protobuf gaps already documented hain in `Dependency/gRPC.md` and `Dependency/Protobuf.md`.

### RabbitMQ

Search env currently `SEARCH_INDEXER_ENABLED=true`, `QUEUE_PROVIDER=rabbitmq`, and a Product event queue define karta hai. Product index event consumer ke liye RabbitMQ conditional runtime dependency ho sakta hai; direct synonym CRUD ke liye broker logically required nahi hona chahiye.

Because Search source is absent, startup fail-open/fail-closed behavior verify nahi ho sakta. RabbitMQ setup repeat nahi kiya gaya:

Refer:
`TaskImplementation/Superadmin Panel/task3_Dependency.md`

Section:
`6. Redis, Queue, and External Services` -> `RabbitMQ: new conditional dependency`

Do not add Kafka/NATS or assume `platform.settings.updated` RabbitMQ par publish hota hai until one owner/provider contract exists.

## 7. Environment Variables

### Frontend delta

Task 7 introduces **no new frontend variable**.

Existing creation/loading/security rules:

Refer:
`TaskImplementation/Superadmin Panel/task1_Dependency.md`

Section:
`7. Environment Variables`

The existing three-variable template is already documented in the referenced section and is not duplicated here. Only `VITE_API_BASE_URL` is relevant to Task 7. MySQL DSN, Redis password, Typesense API key, JWT secret, and internal addresses browser env me mat add karein.

### Existing API base-path blocker

This shared defect was first documented in:

`TaskImplementation/Superadmin Panel/task2_Dependency.md`

Section:
`7. Environment Variables` -> `Critical base-path mismatch`

Base already `/api/v1` par end hota hai, while Task 7 paths bhi `/api/v1/...` se begin hote hain. Current real URLs become:

```text
http://localhost:8080/api/v1/api/v1/admin/settings
http://localhost:8080/api/v1/api/v1/admin/search/synonyms
```

Sirf base URL ko host origin par change karna feature calls fix karega but current login path composition break karega.

> **Required alignment:** One convention choose karein. Recommended: base `http://localhost:8080/api/v1`, platform path `/admin/settings`, synonym path `/admin/search/synonyms`, and login path `/auth/login`. Change ke baad Vite restart and production-like URL tests add karein.

### Task-specific Superadmin backend delta

Local ignored file location:
`backend/services/superadmin-service/.env`

Only Task 7-specific values:

```dotenv
SUPERADMIN_PLATFORM_SETTINGS_CACHE_TTL=5m
SUPERADMIN_PLATFORM_SETTINGS_EVENT_TOPIC=platform.settings.updated
```

Real full-stack settings mode also needs existing database variables from `Dependency/Environment.md`. Current local `SUPERADMIN_REQUIRE_DATABASE=false` is unsafe for production settings mutations; production/integration policy should fail closed:

```dotenv
SUPERADMIN_REQUIRE_DATABASE=true
```

| Variable | Required? | Purpose | Security note |
|---|---:|---|---|
| `SUPERADMIN_PLATFORM_SETTINGS_CACHE_TTL` | Required if cache enabled | Staleness bound | Parse/validate positive duration at startup |
| `SUPERADMIN_PLATFORM_SETTINGS_EVENT_TOPIC` | Required if event invalidation selected | Invalidation topic name | Topic is not secret; broker credentials are |
| `SUPERADMIN_REQUIRE_DATABASE` | Yes for real mutations | Prevent degraded write path | Set true outside explicitly mocked frontend mode |

### Task-specific Search backend delta

Local ignored file location:
`backend/services/search-service/.env`

Sanitized task-specific template:

```dotenv
SEARCH_HTTP_ADDR=:8085

TYPESENSE_URL=
TYPESENSE_HOST=localhost
TYPESENSE_PORT=8108
TYPESENSE_PROTOCOL=http
TYPESENSE_API_KEY=replace-with-a-local-secret
TYPESENSE_PRODUCTS_COLLECTION=products

SEARCH_ADMIN_AUTH_ENABLED=true
SEARCH_ADMIN_MUTATION_RATE_LIMIT=30
SEARCH_ADMIN_MUTATION_RATE_WINDOW=1m
```

| Variable | Required? | Purpose | Security note |
|---|---:|---|---|
| `TYPESENSE_HOST`, `PORT`, `PROTOCOL` | Yes when URL is blank | Backend endpoint | Use private DNS and HTTPS in deployed environments |
| `TYPESENSE_URL` | Optional alternative | Combined endpoint | Loader precedence must be documented |
| `TYPESENSE_API_KEY` | Yes | Search Service authentication | Backend secret only; rotate regularly |
| `TYPESENSE_PRODUCTS_COLLECTION` | Yes | Collection where product synonyms apply | Environment-specific name may prevent accidental cross-env writes |
| `SEARCH_ADMIN_AUTH_ENABLED` | Yes | Protect admin search operations | Keep true outside isolated tests |
| `SEARCH_ADMIN_MUTATION_RATE_*` | Yes | Bound admin changes | Must align with Gateway and documented 30/min policy |

Current Gateway separately needs `SEARCH_GRPC_ADDR=localhost:50058`, but Search Service has no matching detected gRPC listener configuration. Add one only with the future config loader/server implementation.

Backend `.env.example` files are missing. Real keys/DSNs belong in ignored local env or deployment secret manager; commit only sanitized placeholders.

## 8. Docker Setup, Ports, and Networking

Task 7 adds no supported application Dockerfile, Compose network, restart policy, migration job, or health check. Reuse:

- `task1_Dependency.md`, section `8. Docker Setup`, for frontend limitation.
- `Dependency/MySQL.md` for existing MySQL example.
- `Dependency/Redis.md` for existing Redis example.
- Section 6 of this guide for the new Typesense one-off container.
- `task3_Dependency.md` for conditional RabbitMQ setup.

Do not run an invented `docker compose up`: no Compose manifest exists.

### Ports table

| Service | Port/address | Purpose | Task 7 status |
|---|---:|---|---|
| Vite frontend | `5173` default | Admin UI | Reused; runnable |
| Vite preview | `4173` commonly | Built bundle preview | Reused; optional |
| API Gateway HTTP | `8080` | Browser REST entry | Reused; runtime missing |
| Auth/JWKS HTTP | `8081` in Gateway config | Token/JWKS dependency | Reused; runtime missing |
| Superadmin HTTP | `8088` | Detected settings service listener | Task 7 env only |
| Superadmin gRPC | `50062` | Gateway settings target | Task 7 required; server missing |
| Search HTTP | `8085` | Detected Search listener | Task 7 env only |
| Search gRPC | `50058` | Gateway synonym target | Task 7 required; server/listener config missing |
| Typesense | `8108` | Search collection and synonyms | New; local container possible |
| MySQL | `3306` | `superadmin_db` | Reused; Task 7 required full stack |
| Redis | `6379` | Gateway rate limiting/settings cache design | Reused; cache wiring incomplete |
| RabbitMQ AMQP | `5672` | Product-index events | Conditional; reused |
| RabbitMQ management | `15672` | Local diagnostics | Conditional; do not expose publicly |

Networking findings:

- Search HTTP `8085` conflicts with the planned Payment internal HTTP target already documented in Task 5.
- Docker/Kubernetes me `localhost` current container/pod hota hai. Cross-service calls ke liye private service DNS use karein.
- Browser only Gateway ko call kare; Gateway CORS exact frontend origin and required headers allow kare.
- MySQL, Redis, Typesense, RabbitMQ, and internal gRPC/HTTP ports public internet par expose mat karein.
- Changing a port requires every caller, health probe, firewall rule, and deployment manifest update.

Useful listener check:

```bash
ss -ltnp | grep -E ':5173|:8080|:8085|:8088|:50058|:50062|:8108|:3306|:6379'
```

## 9. Local Development Setup

### Step 1: Read reused setup first

Follow:

- `task1_Dependency.md` for clone, Node/pnpm, locked install, frontend `.env`, and normal startup.
- `task2_Dependency.md`, section 7, for the unresolved API URL convention.
- `Dependency/MySQL.md` and `Dependency/Redis.md` only for real integration infrastructure.
- Section 6 here for the new Typesense local setup.

### Step 2: Go to the frontend workspace

```bash
cd Ecommerce/frontend
```

### Step 3: Install locked dependencies

Run the exact workspace locked-install command from `task1_Dependency.md`, section `4. Dependency Management`. No Task 7 package-add command is needed.

### Step 4: Configure frontend environment

Create the frontend `.env` exactly as documented in `task1_Dependency.md`, section `7. Environment Variables`. No Task 7 frontend variable add karein. Real API integration se pehle section 7 ka `/api/v1` duplication resolve karein.

### Step 5: Choose validation mode

Frontend-only mode:

- MySQL, Redis, Typesense, RabbitMQ, Gateway, Auth, and backend migrations skip kar sakte hain.
- Fetch-mocked settings tests reliable local validation boundary hain.

Future full-stack mode:

1. Obtain runnable Auth, API Gateway, Superadmin Service, and Search Service source/manifests.
2. Approve and implement missing synonym REST/gRPC contracts.
3. Add sanitized backend env templates and matching Search/Superadmin gRPC listeners.
4. Start MySQL and apply reviewed versioned Superadmin migrations/seeds.
5. Start Redis/cache/event dependencies according to the selected implementation.
6. Start Typesense, verify `/health`, and let Search Service own collection schema.
7. Start RabbitMQ only if the enabled Search product indexer requires it.
8. Start Superadmin/Search services and verify dependency readiness.
9. Start Auth/JWKS, then Gateway, then verify both gRPC targets.
10. Provision least-privilege non-production admins and synthetic settings/search fixtures.

Exact backend build/start commands cannot honestly be supplied until source, modules, listener config, and entry points exist.

### Step 6: Run Task 7 tests

```bash
cd frontend/superadmin-panel
pnpm exec vitest run tests/settings
```

### Step 7: Run frontend quality checks

Use the unchanged typecheck, full-test, and build commands from `task1_Dependency.md`, section `10. Running the Project`. Task 7 adds no new quality script.

## 10. Running the Project

Start the frontend with the unchanged command in `task1_Dependency.md`, section `10. Running the Project`; it is intentionally not duplicated here.

Open:

```text
http://localhost:5173/admin/settings
```

Current menu/route allows only `superadmin`. Without a valid stored session, route `/login` par redirect karega. Backend unavailable ho to settings/synonym requests fail hona npm install problem nahi hai.

### Verified frontend baseline

Repository verification on 2026-06-20:

| Check | Result |
|---|---|
| Task 7 settings tests | `6` files, `15` tests passed |
| TypeScript project check | Passed |
| Vite production build | Passed |
| Build output | Main JavaScript chunk `560.24 kB`; optimization warning only |
| New package required | None |

### Full-stack read-only verification

Only after missing runtimes, URL convention, and contracts are fixed, use a temporary non-production token:

```bash
export ADMIN_TOKEN='<temporary-superadmin-access-token>'

curl -i \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "X-Request-ID: task7-settings-list-check" \
  'http://localhost:8080/api/v1/admin/settings'
```

After the missing synonym list route is approved and implemented:

```bash
curl -i \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "X-Request-ID: task7-synonym-list-check" \
  'http://localhost:8080/api/v1/admin/search/synonyms'
```

Do not paste a production token into shared shell history. Read-only checks first karein. Commission, flags, maintenance, and synonyms mutation only synthetic/staging fixtures, approval, rollback plan, and audit observation ke saath test karein.

Successful end-to-end verification should prove:

- Final URL contains `/api/v1` exactly once.
- JWT and granular backend permission are enforced.
- Four supported setting keys have correct JSON shapes.
- Every mutation rejects missing/short reason server-side.
- `admin_audit_logs` receives actor, reason, before/after, request ID, and timestamp.
- Settings cache is invalidated immediately after update.
- Search synonym mutation becomes effective in Typesense/search results.
- Secrets never appear in response, UI, browser storage, or logs.

## 11. Common Errors & Fixes

Generic Node, pnpm, Vite, login, Docker, CORS, MySQL, and Redis issues are already documented in previous guides. Task 7-specific troubleshooting:

| Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| URL contains `/api/v1/api/v1/` | Base and Task 7 paths both versioned | Normalize shared URL convention; section 7 dekhein | Test real `.env.example` URL composition |
| Platform settings return `502` | Gateway cannot reach Superadmin gRPC `50062` | Supply/start server and verify DNS/TLS/readiness | Dependency-aware Gateway readiness |
| Synonym request returns `404` | REST CRUD routes absent from master contract/Gateway | Approve mappings and implement required RPCs | Contract-first CI tests |
| Search HTTP `8085` healthy but Gateway fails | Gateway expects missing gRPC `50058` | Implement and configure Search gRPC listener | Document both listeners and probes |
| Page says no platform settings | No seed rows/migration or wrong DB | Apply approved seed/migration and verify `superadmin_db` | Idempotent versioned seed plus startup readiness |
| One expected card shows defaults | API omitted key or returned unsupported key | Inspect sanitized response/key spelling | Shared enum/schema and contract fixtures |
| Settings update returns `403` | Token lacks `superadmin` or backend policy differs | Use approved role; align one permission matrix | End-to-end RBAC tests |
| Readonly/catalog helper test passes but route is denied | Menu/route itself permits only `superadmin` | Decide intended UX; align route, menu, and helper | One authoritative role matrix |
| Reason dialog never submits | Trimmed reason under 10 characters or value invalid | Enter meaningful reason and correct validation errors | Same server/client validation rules |
| Maintenance time is rejected/wrong | Invalid timestamp, end before start, or timezone confusion | Use valid ISO/UTC values and show timezone clearly | Contract tests around DST/time boundaries |
| Typesense connection refused | Container/service down or Docker hostname wrong | Check `/health`, port, and service DNS | Readiness probe and supported local stack |
| Typesense returns `401` | API key mismatch/missing | Align backend secret and rotate if exposed | Secret-manager injection and startup validation |
| Synonym saves but search is unchanged | Search Service did not sync Typesense/collection | Inspect Search logs and collection config; reconcile | Treat Typesense sync as part of mutation success |
| Setting saves but another process reads old value | Cache invalidation/event path missing/failed | Invalidate synchronously or recover through durable retry | Cache-age metrics and invalidation tests |
| Two admins overwrite each other | No proven version/ETag optimistic locking | Add version/ETag and return `409` on stale write | Concurrency integration tests |
| Search Service startup asks for RabbitMQ | Indexer enabled under current env | Start approved broker or disable indexer only in an approved local profile | Explicit optional-dependency startup modes |
| Build warns chunk over 500 kB | App ships a large main bundle | Add route-level lazy loading later | Bundle budget; warning is not current failure |

## 12. Security & Best Practices

### Task-specific security rules

- Backend must authorize every read/mutation; React route/button guards are UX only.
- `platform:settings:write` should remain `superadmin` only.
- Search synonym permission should be explicitly limited to approved `superadmin`/`catalog_admin` roles.
- Every mutation needs actor, reason, request ID, before/after value, timestamp, and outcome in immutable audit storage.
- Never store API keys, DSNs, JWT secrets, passwords, or private endpoints inside generic platform settings JSON.
- Only allowlisted setting keys and strict typed shapes accept karein; arbitrary JSON merge unsafe hai.
- Commission changes financial impact rakhte hain. Range validation, staged rollout, effective time, and optional maker-checker use karein.
- Maintenance mode ke liye clear UTC window, admin bypass policy, automatic expiry behavior, and tested rollback path chahiye.
- Feature flags ko secret store or permanent authorization mechanism na banayein. Flag ownership, default, expiry, and rollback record karein.
- Typesense admin key backend-only/private network me rakhein. Browser direct Typesense admin API access nahi kare.
- Cache invalidation mutation contract ka part ho; stale maintenance/commission settings operational incident create kar sakte hain.
- Admin mutations par 30/minute or stricter approved limit Gateway and service both enforce karein.
- Production admin accounts need MFA, short sessions, TLS, least privilege, and monitored audit alerts.

### Configuration audit

| Severity | Finding | Impact | Required action |
|---|---|---|---|
| Critical | Superadmin, Search, Gateway, and Auth runnable source absent | Full-stack Task 7 cannot start | Add modules, entry points, listeners, health checks, and READMEs |
| Critical | Frontend duplicates `/api/v1` with current env template | Real Task 7 calls use wrong URLs | Normalize one path convention and test it |
| Critical | Synonym REST CRUD contract is missing/incomplete | UI list/update/delete cannot integrate | Add REST mappings, RPCs, schemas, auth, and contract tests |
| High | No setting migration/default seed | Empty UI or unsafe ad-hoc defaults | Add reviewed idempotent versioned migration/seed |
| High | Cache TTL/event topic exist without proven Redis/broker wiring | Stale platform behavior | Implement one explicit cache/invalidation topology |
| High | Search gRPC target `50058` has no detected server listener | Gateway synonym path unavailable | Add/configure listener and readiness |
| High | Route/menu permits only superadmin while helpers include readonly/catalog | Intended role capabilities are unreachable | Align route placement and permission policy |
| High | Generic contract auth for setting reads/search RPCs may be broader than UI policy | Sensitive config/search admin exposure | Use endpoint-specific backend permissions |
| High | No optimistic lock field in DB/public contract | Lost updates between admins | Add version/ETag/updated-at precondition |
| High | No Typesense reconciliation/rollback behavior | UI success may not affect search | Define atomicity, retry, and drift repair |
| Medium | Search `8085` conflicts with planned Payment internal address | Multi-service local startup collision | Maintain central unique port registry/service DNS |
| Medium | Search indexer is enabled with RabbitMQ but startup behavior is unknown | Unclear local dependency chain | Add explicit feature profile and readiness policy |
| Medium | Backend sanitized `.env.example` files absent | Onboarding/secret validation unreliable | Commit placeholders and validate at startup |
| Medium | No Docker/Compose/Kubernetes manifests or probes | Runtime not reproducible | Add only after runtimes/contracts are implemented |
| Low | Planning guide suggests unused packages | Lockfile churn/confusion | Follow actual imports and manifest |
| Low | Main frontend chunk exceeds warning threshold | Slower first load possible | Lazy-load admin routes and remeasure |

Local backend `.env` files are ignored by Git, which is safer than tracking credentials. Still, values may be reused or leaked locally; rotate any real key exposed elsewhere and add sanitized `.env.example` templates.

### Task-specific operational best practices

- Emit metrics for settings read/update failures, cache age, invalidation failures, audit failures, Typesense sync latency, synonym drift, and rejected admin mutations.
- Alert when maintenance mode remains enabled after its end window.
- Keep a known-good settings snapshot and tested rollback procedure.
- Correlate Browser -> Gateway -> service -> audit/cache/Search with `X-Request-ID`.
- Back up and restore `platform_settings`/audit data according to compliance policy; test restoration without overwriting newer settings.
- Add synthetic search fixtures proving synonym behavior before production rollout.
- Use canary/staging for commission, flag, and maintenance changes; avoid direct unobserved production edits.
- Separate liveness from readiness. A process may be alive but not ready if DB/audit/Typesense/cache is unavailable.

## 13. Missing or Misconfigured Things

Complete production onboarding is currently blocked by:

1. Runnable Auth, Gateway, Superadmin Service, and Search Service source/modules are absent.
2. Search synonym Gateway REST routes are absent from `api/master-api.json`.
3. Search gRPC only declares create/list; frontend also requires update/delete.
4. Search gRPC listener/address is not configured on the server side.
5. Platform setting versioned migration, rollback, and four-key seed are absent.
6. Settings API/schema lacks proven optimistic concurrency control.
7. Shared frontend base/path combination duplicates `/api/v1`.
8. Settings route/menu is `superadmin`-only although permission helpers describe readonly/catalog capabilities.
9. Backend granular read/synonym RBAC is not aligned with the frontend role matrix.
10. Superadmin cache TTL/event topic has no proven Redis/broker client wiring.
11. Typesense synonym persistence, mutation atomicity, reconciliation, and rollback are not implemented/proven.
12. Search HTTP port conflicts with an existing planned Payment internal address.
13. Search's enabled RabbitMQ indexer dependency cannot be verified without source.
14. Sanitized backend `.env.example`, supported Docker/Compose manifests, and exact health/readiness commands are absent.
15. Task 7 planning dependency commands do not match the actual package manifest/imports.
16. Existing tests mock an origin-only base and therefore do not catch the real `.env.example` URL defect.

These are backend/platform prerequisites—not reasons to install random frontend packages.

## 14. References to Previous Dependency Files

All earlier dependency files in this service folder were reviewed before creating this incremental guide.

| Previous dependency file | Section/topic reused | Why reused |
|---|---|---|
| `task1_Dependency.md` | Tech stack, software, pnpm, clone, frontend env/Docker/startup, generic errors | Same React application and toolchain |
| `task2_Dependency.md` | API base-path mismatch, Gateway/Auth/Redis model, credential placement | Same unresolved shared HTTP/runtime configuration |
| `task3_Dependency.md` | RabbitMQ setup and external-service decision rule | Current Search env conditionally enables RabbitMQ indexer |
| `task4_Dependency.md` | Service networking, gRPC/readiness, contract-first practices | Same missing backend/runtime boundary |
| `task5_Dependency.md` | Search/Payment `8085` collision and high-risk mutation controls | Same local port registry and admin safety concern |
| `task6_Dependency.md` | Incremental setup pattern, Redis/private networking, frontend verification | Same application and no-speculation rule |
| `Dependency/Frontend.md` | Actual package inventory and commands | Task 7 adds no npm package |
| `Dependency/Environment.md` | Frontend/backend variable ownership and settings env names | Same Vite/secret placement; Task 7 adds exact deltas only |
| `Dependency/API_Gateway.md` | Settings routes, JWT/RBAC, Redis, and Gateway responsibility | Same public REST entry point |
| `Dependency/MySQL.md` | Install, Docker example, credentials, start, verification | Same server; this guide adds Task 7 table/key checks |
| `Dependency/Redis.md` | Install, Docker example, start, `PONG`, and generic errors | Same Redis; this guide adds settings cache/invalidation gap |
| `Dependency/Migrations.md` | Bootstrap-versus-versioned migration warning | Task 7 lacks migration and seed runner |
| `Dependency/gRPC.md` | Internal communication and missing implementation | Gateway expects Superadmin `50062` and Search `50058` |
| `Dependency/Protobuf.md` | Generation/versioning gap | Settings/synonym RPC implementations are absent/incomplete |
| `Dependency/Go_Modules.md` | Missing backend modules/workspace setup | Required backends cannot currently build |
| `Dependency/main_dependency.md` | Overall topology, setup order, ports, and missing-runtime audit | Same service foundation |
| `Dependency/MongoDB.md` | Reviewed; no setup reused | MongoDB is not on the Task 7 request path |

## 15. Final Checklist

### Frontend-only readiness

- [ ] Previous dependency documentation checked
- [ ] Node.js 22.12+ and pnpm 10 available
- [ ] Workspace cloned and installed with `pnpm install --frozen-lockfile`
- [ ] No unnecessary Zod, clsx, MSW, broker, or duplicate package added
- [ ] Existing frontend `.env` created without secrets
- [ ] Shared API base-path defect understood before real integration
- [ ] Task 7 settings tests pass
- [ ] Frontend typecheck passes
- [ ] Full frontend test suite passes
- [ ] Production build passes; chunk warning tracked separately
- [ ] Vite starts and `/admin/settings` route opens for an approved superadmin session
- [ ] Invalid commission, flag key, maintenance window, synonym, and short reason are rejected in UI tests
- [ ] Backend absence is not mistaken for dependency-install failure

### Full-stack readiness

- [ ] `/api/v1` convention fixed for login, settings, and synonyms
- [ ] Runnable Auth, Gateway, Superadmin Service, and Search Service supplied
- [ ] Search REST CRUD and gRPC create/list/update/delete contracts approved and implemented
- [ ] Superadmin `50062` and Search `50058` gRPC listeners/readiness verified
- [ ] Sanitized backend `.env.example` files tracked; real secrets injected securely
- [ ] MySQL `superadmin_db` reachable through least-privilege credentials/TLS as required
- [ ] Versioned settings/audit migrations and rollback procedure tested
- [ ] Four supported settings seeded through an approved idempotent process
- [ ] Setting value schemas and server validation match frontend expectations
- [ ] Optimistic locking/version conflict strategy implemented and tested
- [ ] Redis/cache topology and five-minute TTL policy explicitly implemented
- [ ] Update invalidates cache/event consumers immediately and recoverably
- [ ] Typesense private endpoint, persistent volume, pinned image, API key, backup, and health configured
- [ ] Search Service owns collection/synonym schema and drift reconciliation
- [ ] RabbitMQ requirement confirmed; broker healthy only if enabled indexer needs it
- [ ] Search/Payment `8085` collision resolved for combined environments
- [ ] Gateway and service RBAC match approved superadmin/catalog/read-only policy
- [ ] MFA, short admin session, TLS/mTLS, CORS, request validation, and mutation rate limits verified
- [ ] Every mutation writes immutable actor/reason/before/after/request-ID audit data
- [ ] Commission change verified only with synthetic/staging financial fixtures
- [ ] Feature flag enable/disable and rollback verified
- [ ] Maintenance schedule, UTC display, bypass, expiry alert, and rollback verified
- [ ] Synonym create/list/update/delete changes actual test search behavior
- [ ] Cache and Typesense failure paths do not report false success
- [ ] No database DSN, Redis password, Typesense key, token, or internal secret reaches browser/logs
- [ ] Metrics, alerts, backup/restore, rollback, and incident runbooks tested
- [ ] No duplicate setup documentation added

Frontend Task 7 setup is independently verifiable. Complete Platform Settings readiness tabhi claim karein when missing runtimes/contracts, migrations/seeds, cache invalidation, Typesense synchronization, granular RBAC, audit behavior, and concurrency safety end to end pass ho.
