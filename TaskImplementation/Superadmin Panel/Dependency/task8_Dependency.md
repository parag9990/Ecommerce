# Project Dependency & Setup Guide

## 1. Project Overview

Yeh guide Task 8 ke implemented **Audit Log Viewer** ko install, run, test, aur full-stack environment se connect karne ke liye hai. Business logic yahan repeat nahi ki gayi hai.

Task implementation source:

`TaskImplementation/Superadmin Panel/task8.md`

Actual frontend implementation:

`frontend/superadmin-panel/src/features/audit/`

### Current repository reality

| Area | Required for Task 8 | Current status |
|---|---|---|
| React audit viewer | Yes | Implemented |
| Filter, table, drawer, RBAC, and export UI | Yes | Implemented |
| Node packages | Yes | Already present in the shared lockfile |
| New npm package | No | Task 8 uses existing packages and browser APIs |
| New frontend environment variable | No | Existing `VITE_API_BASE_URL` is reused |
| API Gateway and admin authentication | Real data ke liye yes | Contract/env present; runnable source absent |
| Superadmin Service | Real data ke liye yes | `.env` only; Go source/module absent |
| MySQL `superadmin_db` | Full-stack mode me yes | Bootstrap schema present; migration runner absent |
| Audit list REST endpoint | Yes | Declared in `api/master-api.json`; runtime absent |
| Audit export REST endpoint | Yes for export | Frontend calls it, but master API contract/runtime me absent |
| Redis | Indirect Gateway dependency | Reused; no Task 8-specific cache added |
| Kafka/RabbitMQ/NATS | No direct requirement | Task 8 implementation does not use a queue |
| Docker | Optional | No Dockerfile or Compose manifest exists |

> **Beginner note:** Frontend ka typecheck/build backend ke bina chal sakta hai. Real audit rows aur CSV export ke liye Gateway, Auth, Superadmin Service, aur MySQL sab reachable hone chahiye.

## 2. Tech Stack

Shared React/Vite stack ka complete beginner explanation already available hai:

Refer:
`TaskImplementation/Superadmin Panel/task1_Dependency.md`

Section:
`2. Tech Stack`

Task 8 me relevant technologies ka delta:

| Technology | Task 8 me use | Required? | New? |
|---|---|---|---|
| React 19 | Filters, table, detail drawer, export dialog | Yes | Reused |
| TypeScript | Audit payload/filter typing and validation | Yes | Reused |
| React Router DOM | `/admin/audit-logs` protected route | Yes | Reused |
| TanStack React Query | List query, 30-second stale time, export mutation | Yes | Reused |
| Zustand | Logged-in admin roles/token | Yes | Reused |
| Lucide React | Viewer and export icons | Yes | Reused |
| Browser `URL`, `Blob`, and download APIs | Query strings and CSV download | Yes | No package needed |
| Vitest + Testing Library + jsdom | Audit API/helper/component tests | Development only | Reused |
| MySQL 8 | Immutable `admin_audit_logs` persistence | Backend only | Reused |

### Planning guide versus actual dependencies

`task8.md` suggests `clsx`, Zod, `date-fns`, PapaParse, and MSW. Actual source and `package.json` do **not** use or declare them:

- Class composition uses the existing local `cn` helper.
- Validation is implemented in `validators.ts`.
- Date handling uses native `Date`.
- CSV escaping helper is implemented in `audit-csv.ts`.
- API tests stub `fetch` directly instead of MSW.

Do not run the planning-only `pnpm add` commands. Extra packages dependency tree aur attack surface unnecessarily badhayenge.

## 3. Required Software

Software installation, supported Node version, pnpm setup, Git, browser, and optional Docker steps already documented hain.

Refer:
`TaskImplementation/Superadmin Panel/task1_Dependency.md`

Sections:
`3. Required Software` and `9. Local Development Setup`

Task 8 adds no system-level software. Quick verification:

```bash
node --version
pnpm --version
git --version
```

Use Node `22.12+` on the Node 22 line and pnpm 10 as documented previously.

## 4. Dependency Management

### New dependency delta

There is no Task 8 package delta. Required packages are already declared in:

- `frontend/superadmin-panel/package.json`
- `frontend/pnpm-lock.yaml`

Install the existing locked workspace only:

```bash
cd frontend
pnpm install --frozen-lockfile
```

Do not use npm or Yarn in this pnpm workspace, and do not manually edit `pnpm-lock.yaml`.

### Commands

```bash
cd frontend
pnpm run superadmin:typecheck
pnpm run superadmin:test
pnpm run superadmin:build
```

Audit-only tests:

```bash
cd frontend/superadmin-panel
pnpm exec vitest run tests/audit
```

Complete pnpm concepts and package-failure fixes are reused from:

`TaskImplementation/Superadmin Panel/task1_Dependency.md` -> `4. Dependency Management`

## 5. Database Setup

### Direct frontend requirement

Browser MySQL se direct connect nahi karta. Frontend-only typecheck, tests, build, aur mocked UI work ke liye database setup skip kar sakte hain.

### Full-stack MySQL requirement

MySQL ek relational database hai. Task 8 me Superadmin Service ko filterable, ordered, immutable audit records read karne ke liye `superadmin_db.admin_audit_logs` chahiye. Full-stack viewer ke liye yeh mandatory hai.

MySQL install, optional Docker command, credentials, startup, aur generic verification repeat nahi kiye gaye hain.

Refer:
`TaskImplementation/Superadmin Panel/Dependency/MySQL.md`

Sections:
`6. Installation steps` through `12. Common errors and fixes`

### Task 8 schema verification

The bootstrap definition is in `database/draw.sql`. It provides:

| Schema item | Purpose | Status |
|---|---|---|
| `audit_id` unique key | Stable public audit identifier | Present |
| `(actor_admin_id, created_at)` index | Actor/time filtering | Present |
| `(resource_type, resource_id)` index | Resource investigation | Present |
| `(action, created_at)` index | Action/time filtering | Present |
| `request_id` | Cross-service correlation | Present, not indexed |
| `ip_hash` | Privacy-safe request correlation | Present |
| `before_json`, `after_json` | Change summaries | Present |
| append-only DB enforcement | Prevent update/delete | Not proven |
| retention/archive policy | Control table growth | Not implemented |

After following the reused MySQL setup, verify only the Task 8 table:

```bash
mysql -u root -p -e "USE superadmin_db; SHOW CREATE TABLE admin_audit_logs\G"
mysql -u root -p -e "USE superadmin_db; SHOW INDEX FROM admin_audit_logs;"
mysql -u root -p -e "USE superadmin_db; SELECT COUNT(*) AS audit_rows FROM admin_audit_logs;"
```

### Schema/API mismatches to resolve

- Frontend expects `actor_role`; the table has no `actor_role` column. Backend must derive it reliably or add a reviewed immutable snapshot field.
- Table columns are `before_json`/`after_json`; frontend contract expects `before_summary`/`after_summary`. Repository mapping must be explicit.
- Table has numeric `id` and public `audit_id`; API must state which value it returns as frontend `id`.
- UI filters by `request_id`, but no request-ID index exists. Large datasets may require an index after query-plan measurement.
- The bootstrap SQL is not a versioned migration. Production setup needs reviewed up/down migrations and backup/rollback procedures.

Migration principles and existing gaps are already explained in:

`TaskImplementation/Superadmin Panel/Dependency/Migrations.md`

## 6. Redis / Queue / External Services

### Required request path

```text
Admin browser
  -> API Gateway :8080 (JWT, RBAC, validation, rate limiting)
  -> Superadmin Service gRPC :50062
  -> MySQL :3306 / superadmin_db.admin_audit_logs
```

### API Gateway and Auth

Gateway real audit data ke liye mandatory hai. It must:

- validate the Bearer JWT;
- allow only `superadmin` and `readonly_admin` to list logs;
- allow only `superadmin` to export;
- pass trusted actor/request context to the service;
- return stable JSON/error envelopes and CSV responses;
- apply stricter rate/size limits to export.

Shared Gateway setup is already documented in:

`TaskImplementation/Superadmin Panel/Dependency/API_Gateway.md`

### Redis

Gateway `.env` points to Redis on `localhost:6379` for shared rate limiting. Task 8 adds no Redis key, cache, database number, or persistence rule.

Refer:
`TaskImplementation/Superadmin Panel/Dependency/Redis.md`

### Queues and other services

Kafka, RabbitMQ, NATS, MongoDB, Typesense, Elasticsearch/OpenSearch, MinIO/S3, SMTP, and payment providers are not called by the implemented Task 8 frontend. Do not start or install them merely for this viewer.

Long-running/very large exports may later use an async worker and object storage, but that architecture is not implemented and must not be treated as a current dependency.

## 7. Environment Variables

### Frontend delta

Task 8 introduces no new frontend variable. It reuses:

| Variable | Required? | Purpose | Secret? |
|---|---|---|---|
| `VITE_API_BASE_URL` | Yes for real APIs | Gateway URL prefix | No |

### Critical base-path mismatch

Current `.env.example` contains:

```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

Audit code passes `/api/v1/admin/audit-logs`, so the shared HTTP client builds:

```text
http://localhost:8080/api/v1/api/v1/admin/audit-logs
```

Changing only the environment value to `http://localhost:8080` fixes audit URLs, but current login code then calls `/auth/login` while the master contract declares `/api/v1/auth/login`. Therefore, no single environment value currently fixes every feature.

Recommended project-wide convention:

```env
# frontend/superadmin-panel/.env
VITE_API_BASE_URL=http://localhost:8080
```

Then every API path, including login, must explicitly begin with `/api/v1`. This code alignment is still missing and should be completed before real full-stack onboarding.

### Backend delta

Existing Superadmin `.env` has database settings but no implemented config loader/source. For an audit viewer deployment, database availability should not silently degrade:

```env
# backend/services/superadmin-service/.env
SUPERADMIN_REQUIRE_DATABASE=true
```

The checked local value is `false`; change it only when the backend implements and validates the setting. Real DSNs/passwords belong in ignored local env files or a secret manager, never in frontend `VITE_*` values.

Audit export row limits, maximum date range, retention, archive destination, and keyed IP hashing need backend configuration, but no accepted variable names or loader exist. Do not invent env names and assume they work; add them with backend implementation and a sanitized `.env.example`.

Environment loading and security rules are reused from:

`TaskImplementation/Superadmin Panel/task1_Dependency.md` -> `7. Environment Variables`

## 8. Docker Setup

Task 8 adds no container, volume, network, health check, or restart policy. The repository still has no supported Dockerfile/Compose setup for this frontend/Gateway/Superadmin chain.

Refer:
`TaskImplementation/Superadmin Panel/task1_Dependency.md`

Section:
`8. Docker Setup`

MySQL and Redis one-off Docker examples already exist in their shared dependency files. Do not run `docker compose up` from this repository: no Compose manifest was found.

Future production containers should add:

- non-root frontend and backend images;
- SPA fallback for direct `/admin/audit-logs` navigation;
- MySQL/Redis persistent volumes owned by their services;
- private backend network and TLS/mTLS policy;
- liveness plus dependency-aware readiness checks;
- pinned image versions, resource limits, restart policy, and secret injection;
- pre-deploy versioned migration job.

### Ports and networking

| Service | Port/address | Purpose | Status |
|---|---:|---|---|
| Vite dev server | `5173` | Audit viewer UI | Reused |
| Vite preview | `4173` commonly | Local bundle preview | Reused/optional |
| API Gateway HTTP | `8080` | Public admin REST entry | Reused; env only |
| Superadmin HTTP | `8088` | Planned service HTTP listener | Reused; env only |
| Superadmin gRPC | `50062` | Gateway-to-service calls | Reused; server absent |
| MySQL | `3306` | `superadmin_db` | Reused |
| Redis | `6379` | Gateway rate limiting | Reused/indirect |
| New Task 8 port | None | No new process introduced | New delta: none |

If `5173` is busy:

```bash
cd frontend
pnpm --filter superadmin-panel dev -- --port 5174
```

Gateway CORS must then allow the exact new origin.

## 9. Local Development Setup

### Step 1: Read shared setup first

Start with:

`TaskImplementation/Superadmin Panel/task1_Dependency.md`

It owns clone, Node/pnpm installation, base `.env`, optional Docker, and generic troubleshooting.

### Step 2: Go to the workspace

```bash
cd Ecommerce/frontend
```

### Step 3: Install locked dependencies

```bash
pnpm install --frozen-lockfile
```

No Task 8-specific package installation is needed.

### Step 4: Configure frontend environment

```bash
cd superadmin-panel
cp .env.example .env
```

For mocked tests/build, the existing template is enough because no real request is made. For full-stack use, resolve the base-path convention described in Section 7 before starting.

### Step 5: Choose validation mode

Frontend-only mode:

- MySQL, Redis, Gateway, and services are not required.
- Run audit tests, typecheck, and build.

Full-stack mode additionally requires:

1. MySQL with `superadmin_db.admin_audit_logs`.
2. Redis if Gateway rate limiting requires it.
3. Auth Service and a valid admin account.
4. Runnable Superadmin gRPC service on the configured target.
5. Runnable API Gateway with aligned REST routes.
6. List and export contracts implemented end to end.

Items 3-6 cannot currently be started from the inspected repository because their runnable source/module is absent.

### Step 6: Run Task 8 checks

```bash
cd frontend/superadmin-panel
pnpm exec vitest run tests/audit
pnpm typecheck
pnpm build
```

### Step 7: Start the frontend

```bash
cd frontend
pnpm run superadmin:dev
```

Open:

```text
http://localhost:5173/admin/audit-logs
```

Without a hydrated admin session, route should redirect to `/login`. `superadmin` and `readonly_admin` can view; only `superadmin` can export.

## 10. Running the Project

### Verified frontend baseline

During this documentation audit:

- TypeScript build-mode check passed.
- Vite production build passed.
- Build output reported an existing JavaScript chunk around `560 kB` and the standard `500 kB` warning.
- Audit test files were discovered, but Vitest workers timed out in the restricted WSL runner; no test assertion failure was produced. Run the canonical command in a normal local/CI environment.

### Full-stack API verification

Once missing backends/contracts are supplied, set a short-lived test token without printing it:

```bash
export ADMIN_ACCESS_TOKEN='<short-lived-admin-token>'
```

List logs:

```bash
curl -i \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN" \
  'http://localhost:8080/api/v1/admin/audit-logs?page=1&page_size=25&resource_type=refund'
```

Export filtered logs:

```bash
curl -i \
  -X POST \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"filters":{"page":1,"page_size":25,"resource_type":"refund"},"reason":"Quarterly security review"}' \
  'http://localhost:8080/api/v1/admin/audit-logs/export'
```

Expected behavior:

- no token -> `401`;
- unauthorized admin role -> `403`;
- `superadmin`/`readonly_admin` list request -> paginated JSON;
- `readonly_admin` export -> `403`;
- `superadmin` valid export -> CSV/blob response;
- successful export -> separate immutable `audit_logs.exported` audit entry;
- response/request logs -> same request ID, without token/raw IP/secrets.

### Required API contract delta

| Capability | Frontend sends/expects | Master contract status |
|---|---|---|
| List route | `GET /api/v1/admin/audit-logs` | Present |
| Pagination | `page`, `page_size`; optional `total` | Only generic/weakly typed |
| Filters | actor, action, resource, request ID, from/to | Contract declares only actor/resource fields |
| List row | actor role, summaries, request ID, IP hash, reason, time | Response item is untyped object |
| Export | `POST /api/v1/admin/audit-logs/export` with filters/reason | Missing |
| Export response | CSV/blob | Missing |
| Export RBAC/audit | superadmin only; audit the export | Missing/unproven |

## 11. Common Errors & Fixes

Generic pnpm, Node, port, CORS, JWT, Docker, and permission issues are already covered in `task1_Dependency.md` -> `11. Common Errors & Fixes`.

Only Task 8-specific issues are added here:

| Error/symptom | Cause | Fix | Prevention |
|---|---|---|---|
| Request path contains `/api/v1/api/v1` | Base URL and feature path both contain prefix | Adopt one project-wide convention; see Section 7 | Add URL composition contract tests using `.env.example` value |
| Audit list returns `404`/`502` | Gateway route or Superadmin runtime absent | Implement/register route, gRPC client/server, and service | Readiness checks plus end-to-end smoke test |
| Export returns `404` or `405` | Export endpoint is absent from master contract/runtime | Add approved POST contract and implementation | Contract-test frontend against master API |
| Export returns `403` | Role is not `superadmin` | Use properly provisioned account; do not weaken RBAC | API-level role matrix tests |
| Empty table after successful request | Default range is last 24 hours or filters are too narrow | Reset/expand date range and inspect query params | Show active range and return stable totals |
| `From date must be before to date` | Invalid local date range | Correct dates | Keep frontend and backend range validation |
| Table shows `unknown_*` values | Backend row names/fields do not match expected API model | Fix contract/repository mapping | Typed schema and response contract tests |
| Next-page behavior is wrong | Backend omits `total`, `page`, or `page_size` | Return consistent pagination metadata | Make response fields required in contract |
| Export times out after 15 seconds | Shared HTTP client aborts at 15 seconds or export is too large | Reduce range; implement capped/async export | Server row/range limits and export job design |
| Spreadsheet executes a cell as formula | Backend CSV did not neutralize `=`, `+`, `-`, or `@` prefixes | Escape on the backend before producing CSV | CSV injection tests at export endpoint |
| Sensitive value appears in summary | Backend returned unsanitized JSON; frontend mask is only defensive | Remove/rotate leaked secret and sanitize at write/read boundary | Allowlist audit fields; never rely only on key-name regex |
| Vitest worker startup timeout | Restricted WSL/container cannot start worker pool | Run in normal shell/CI; check process/socket restrictions | Keep a supported CI runner and pinned Node/pnpm |

## 12. Security & Best Practices

### Task-specific security rules

- Backend RBAC is mandatory. Hidden buttons and route guards are UX controls, not authorization.
- Audit rows must be append-only. Application DB credentials should not have routine `UPDATE`/`DELETE` rights on historical rows.
- Export must require `superadmin`, a server-validated reason, a bounded filter/range, and a new immutable export audit entry.
- Never include password, OTP, token, cookie, authorization header, private key, full card data, CVV, raw secret, or raw IP in summaries/CSV/logs.
- IP correlation should use a documented keyed one-way hash and secret rotation strategy. Plain SHA-256 of predictable IPs is reversible by guessing.
- Frontend summary masking is defense in depth only. Backend must sanitize before persistence and again before response/export.
- CSV must be escaped by the backend because the implemented page downloads the server blob directly; the local safe CSV helper is not used in that path.
- Enforce maximum page size (`100` in UI), maximum date range, maximum export rows/bytes, timeout, and rate limit on the server.
- Protect list and export endpoints with TLS; use mTLS/private networking for internal gRPC in production.
- Use `X-Request-ID` to correlate browser, Gateway, service, DB query, and export audit entry without logging the token.
- Never allow frontend `VITE_*` variables to contain DB DSNs, signing secrets, hash keys, or service credentials.

### Operational best practices

- Default last-24-hours query is good for load, but backend must enforce its own safe defaults.
- Debounce free-text filters or add an explicit Apply action before production traffic; current input changes can trigger frequent requests.
- Measure query plans before adding indexes. Consider `(request_id)` and time-leading indexes only from real filter patterns.
- Define retention, legal hold, archive, backup, restore, and tamper-evidence procedures before compliance claims.
- Use a read replica/search store only after correctness and freshness rules are documented; MySQL remains the source of truth.
- Record metrics for list latency/errors, scanned rows, export attempts/denials/size/duration, DB failures, and audit-write failures.
- Alert on unusual export volume, repeated denied export, audit gaps, and append-only permission violations.
- Test restore and archive retrieval; a backup that was never restored is not a proven recovery plan.
- Code-split admin routes to address the existing production bundle-size warning.

## 13. Missing or Misconfigured Things

Full Task 8 production onboarding is currently blocked by:

1. Runnable API Gateway and Superadmin Service source/modules are absent.
2. Superadmin service is not registered in `backend/go.work`; only Auth is registered.
3. Superadmin `.proto`, generated code, gRPC listener/config, handler, use case, and repository are absent.
4. `POST /api/v1/admin/audit-logs/export` is absent from `api/master-api.json`.
5. List contract omits action, request-ID, and date filters used by the frontend.
6. List response items are untyped and do not guarantee frontend-required fields/pagination metadata.
7. Frontend base URL/path composition duplicates `/api/v1` for audit while login uses the opposite convention.
8. Versioned `admin_audit_logs` up/down migrations and a migration runner are absent.
9. Append-only DB enforcement, retention, archive, legal-hold, and restore procedures are absent.
10. Table/API naming differs for audit ID and before/after summaries; `actor_role` persistence/derivation is undefined.
11. Backend export CSV escaping, reason validation, role enforcement, row/range cap, and export-audit transaction are unproven.
12. `SUPERADMIN_REQUIRE_DATABASE=false` permits an unsafe local default for a DB-dependent audit service.
13. Sanitized backend `.env.example`, supported Dockerfiles/Compose, and health/readiness checks are absent.
14. No single documented full-stack startup command can work until missing runtimes are supplied.
15. Existing API tests use an origin-only stub and do not exercise the checked `.env.example` base-path defect.
16. Audit tests could not start workers in the restricted verification runner; a normal local/CI run remains required.

These are runtime, contract, and operational gaps. Installing extra frontend libraries will not fix them.

## 14. References to Previous Dependency Files

All earlier dependency guides in this service folder were explored before creating this incremental guide.

| Previous dependency file | Section/topic reused | Why reused |
|---|---|---|
| `task1_Dependency.md` | Tech stack, software, pnpm, clone, env, Docker, startup, generic errors | Same frontend application/toolchain |
| `task2_Dependency.md` | API base-path mismatch, Gateway/Auth/Redis model, credentials | Same unresolved shared request path |
| `task3_Dependency.md` | External-service decision rule and secret handling | Avoid speculative brokers/storage |
| `task4_Dependency.md` | Service networking, gRPC, readiness, contract-first checks | Same missing backend boundary |
| `task5_Dependency.md` | High-risk admin action and export-style controls | Same sensitive operational domain |
| `task6_Dependency.md` | Privacy, request correlation, frontend verification pattern | Same admin oversight foundation |
| `task7_Dependency.md` | MySQL audit writes, incremental setup, current backend gaps | Task 8 reads records written by prior admin mutations |
| `Dependency/Frontend.md` | Actual package inventory and commands | No new Task 8 package |
| `Dependency/Environment.md` | Frontend/backend variable ownership | Same Vite and secret-placement model |
| `Dependency/API_Gateway.md` | Audit list route, JWT/RBAC, Redis, errors | Same public REST entry point |
| `Dependency/MySQL.md` | Install, Docker option, credentials, start, verify | Same `superadmin_db`; only audit-specific checks added |
| `Dependency/Redis.md` | Install, Docker option, `PONG`, common issues | Same indirect Gateway Redis |
| `Dependency/Migrations.md` | Bootstrap versus versioned migration warning | Audit migration runner remains absent |
| `Dependency/gRPC.md` | `ListAuditLogs`, `50062`, missing runtime | Same intended internal protocol |
| `Dependency/Protobuf.md` | Contract generation/versioning gap | Audit/export RPC schema is absent/incomplete |
| `Dependency/Go_Modules.md` | Missing Superadmin module/workspace setup | Backend cannot currently build |
| `Dependency/main_dependency.md` | Overall topology, ports, setup order, missing-runtime audit | Same service foundation |
| `Dependency/MongoDB.md` | Reviewed; no setup reused | MongoDB is not on the Task 8 request path |

## 15. Final Checklist

### Frontend-only readiness

- [ ] Previous dependency documentation checked
- [ ] Node 22.12+ and pnpm 10 available
- [ ] Workspace installed with `pnpm install --frozen-lockfile`
- [ ] No planning-only `clsx`, Zod, `date-fns`, PapaParse, or MSW package added
- [ ] Existing frontend `.env` created without secrets
- [ ] API base-path mismatch understood before real integration
- [ ] Audit tests pass in a supported local/CI runner
- [ ] TypeScript check passes
- [ ] Production build passes
- [ ] Bundle-size warning recorded for route code splitting
- [ ] `/admin/audit-logs` is visible only to approved viewer roles
- [ ] Readonly export remains disabled
- [ ] Invalid range and short export reason are rejected
- [ ] Backend absence is not mistaken for npm dependency failure

### Full-stack readiness

- [ ] One `/api/v1` URL convention applied to login and every feature
- [ ] Runnable Auth, Gateway, and Superadmin Service supplied
- [ ] Superadmin Go module registered and buildable
- [ ] Approved protobuf/gRPC list and export contracts generated and registered
- [ ] `GET /api/v1/admin/audit-logs` supports all implemented filters and pagination
- [ ] `POST /api/v1/admin/audit-logs/export` is contracted and implemented
- [ ] MySQL credentials use secret storage and least privilege
- [ ] `admin_audit_logs` versioned migration and rollback tested
- [ ] Audit ID, actor role, before/after mapping, and pagination schema aligned
- [ ] Append-only enforcement and tamper monitoring verified
- [ ] Gateway and service independently enforce view/export RBAC
- [ ] Export reason, filter, range, row/byte, timeout, and rate limits enforced server-side
- [ ] Backend-generated CSV blocks spreadsheet formula injection
- [ ] Export creates its own immutable audit row
- [ ] Raw IP, credentials, tokens, and sensitive summary fields never persist or export
- [ ] Request IDs correlate browser, Gateway, service, query, and export logs
- [ ] Redis is healthy only if Gateway rate limiting requires it
- [ ] CORS permits the exact trusted frontend origin
- [ ] TLS/mTLS, health/readiness, structured logs, metrics, and alerts verified
- [ ] Retention, legal hold, archive, backup, restore, and incident runbooks tested
- [ ] List and export smoke tests pass for `401`, `403`, success, empty, error, and large-range cases
- [ ] No duplicate setup documentation added

Task 8 frontend setup is independently buildable. Complete Audit Log Viewer readiness tabhi claim karein when the missing backend runtime, exact contracts, DB migration/immutability, secure export, RBAC, retention, and end-to-end checks pass.
