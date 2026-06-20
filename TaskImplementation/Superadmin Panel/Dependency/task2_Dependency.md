# Superadmin Panel Task 2 - Dependency & Setup Guide

## 1. Project Overview

Yeh guide `task2.md` ke **User Management** module ke liye dependency, environment, backend-service, database, run, aur troubleshooting requirements explain karti hai. Business logic ya implementation steps yahan repeat nahi kiye gaye hain.

Task 2 ka frontend implementation `frontend/superadmin-panel/src/features/users/` me present hai. Isme user list/search, profile, block/unblock, recent sessions, aur session journey UI included hain.

### Current repository reality

| Capability | Current status | Beginner meaning |
|---|---|---|
| Frontend source | Ready | UI code, routes, hooks, tests, and package manifest present hain |
| Frontend install/test/build | Ready | Existing pnpm workspace se run ho sakta hai |
| New Task 2 npm dependency | None | Koi extra package install nahi karna |
| New frontend environment variable | None | Existing `.env.example` reuse hota hai |
| Real login and user APIs | Blocked locally | API Gateway, Auth, Superadmin, User, and Session services ka runnable source present nahi hai |
| Task 2 database migration | None in frontend | Frontend migration nahi chalata; backend migrations incomplete hain |
| Docker/Compose runtime | Missing | Repo me is app/service ke liye supported Dockerfile or Compose file nahi mila |

> **Important:** Frontend tests, typecheck, build, aur dev server backend ke bina run ho sakte hain. Real users, status update, and session data tabhi work karenge jab backend runtime supplied ho.

### Task-specific API dependency

| Method | Contract path | Owner | UI use |
|---|---|---|---|
| `GET` | `/api/v1/admin/users` | Superadmin Service, with User Service data | Search/list and profile lookup |
| `PATCH` | `/api/v1/admin/users/{user_id}/status` | Superadmin Service and User Service | Block/unblock with audit reason |
| `GET` | `/api/v1/analytics/sessions` | Session Service | User's recent sessions |
| `GET` | `/api/v1/analytics/sessions/{session_id}/journey` | Session Service | Ordered session events |

All four endpoints require an authenticated admin request according to `api/master-api.json`.

## 2. Tech Stack

Task 2 existing Task 1 frontend stack ko reuse karta hai. Full technology definitions and installation explanation dobara repeat nahi ki gayi hain.

Refer:
`TaskImplementation/Superadmin Panel/task1_Dependency.md`

Sections:
`2. Tech Stack`, `3. Required Software`, and `4. Dependency Management`

Task-specific use:

| Technology | Task 2 me use | Required? | New? |
|---|---|---:|---:|
| React 19 + TypeScript | User pages, filters, profile, status dialog, session drawer | Yes | No |
| React Router | `/admin/users` and `/admin/users/:userId` | Yes | No |
| TanStack React Query | List/detail/session queries, mutation, cache invalidation | Yes | No |
| Zustand | Logged-in admin token and roles | Yes | No |
| Vite + Tailwind CSS | Dev server, build, and UI styles | Yes | No |
| Vitest + Testing Library | User API, permission, and status-dialog tests | Tests only | No |
| Lucide React | User/session/action icons | Yes | No |

### Planning guide versus actual package manifest

`task2.md` suggests installing `clsx`, MSW, and other packages. Actual implementation does **not** import `clsx` or MSW, and they are not declared in `frontend/superadmin-panel/package.json`. Tests use Vitest fetch mocks, while class merging uses project-local helpers.

> Do not run the package-add commands from `task2.md` for the current implementation. `pnpm install --frozen-lockfile` is enough.

## 3. Required Software

No new developer software is introduced by Task 2.

Follow the existing Git, Node.js 22.12+, pnpm 10, and browser setup here:

Refer:
`TaskImplementation/Superadmin Panel/task1_Dependency.md`

Section:
`3. Required Software`

For real end-to-end behavior, these additional runtime components are logically required, but their runnable code is currently missing:

| Component | Required for | Repository evidence/status |
|---|---|---|
| API Gateway | Every real Task 2 API | `.env` and API contract only |
| Auth Service/JWKS | Login and Bearer-token validation | Env/design only; runnable service not found |
| Superadmin Service | User list and status workflow | `.env` only |
| User Service | User records and account status | `.env` only |
| Session Service | Session list and journey | `.env` only |
| MySQL 8 | User data and Superadmin audit/admin data | Schema and env values exist; service repositories/migrations missing |
| MongoDB | Session and journey documents | Schema design and Session env exist; runnable service missing |
| Redis | Gateway rate limit and Session short-lived data | Env exists; runnable Gateway/Session code missing |

Kafka, RabbitMQ, NATS, Elasticsearch, MinIO, SMTP, Stripe, and Twilio are not part of the direct Task 2 request path.

## 4. Dependency Management

### New dependency delta

| Check | Result |
|---|---|
| New runtime npm package | None |
| New development npm package | None |
| `package.json` change needed | No |
| `pnpm-lock.yaml` change needed | No |
| New Go module | No Task 2 backend module is present |

Install the already locked workspace dependencies by following:

Refer:
`TaskImplementation/Superadmin Panel/task1_Dependency.md`

Section:
`4. Dependency Management`

The correct clean-install command remains:

```bash
cd frontend
pnpm install --frozen-lockfile
```

Do not mix npm, yarn, and pnpm lockfiles. Do not run `go mod tidy`, `go build`, or `go run` inside `backend/services/superadmin-service/`: that folder currently has no `go.mod` or Go source.

Backend Go workspace gaps are already explained in:
`TaskImplementation/Superadmin Panel/Dependency/Go_Modules.md`

## 5. Database Setup

### Direct frontend database requirement

None. Browser code must never connect directly to MySQL, MongoDB, or Redis and must never receive their credentials.

### Real Task 2 backend data flow

```text
Browser
  -> API Gateway
     -> Superadmin Service -> User Service -> MySQL user_db
                           -> MySQL superadmin_db audit data
     -> Session Service    -> MongoDB session_db
                           -> Redis DB 2
```

| Data store | Why Task 2 needs it | Required when? | Default port | Setup source |
|---|---|---|---:|---|
| MySQL `user_db` | User profile/status data | Real user APIs | `3306` | `database/draw.sql` plus backend owner setup |
| MySQL `superadmin_db` | Admin identity/permissions and audit records | Real status mutation/audit | `3306` | `Dependency/MySQL.md` |
| MongoDB `session_db` | Sessions and journey events | Session panel/journey | `27017` | `Dependency/MongoDB.md` |
| Redis DB `0` | Gateway rate limits in current env | Gateway with rate limiting | `6379` | `Dependency/Redis.md` |
| Redis DB `2` | Session cache/counters in current env | Session runtime | `6379` | Session Service env; same Redis installation can use separate logical DB |

Installation and generic Docker commands are already documented and are not repeated:

- MySQL: `TaskImplementation/Superadmin Panel/Dependency/MySQL.md`, sections `6` to `12`.
- MongoDB: `TaskImplementation/Superadmin Panel/Dependency/MongoDB.md`, sections `5` to `11`.
- Redis: `TaskImplementation/Superadmin Panel/Dependency/Redis.md`, sections `5` to `11`.
- Migration limitations: `TaskImplementation/Superadmin Panel/Dependency/Migrations.md`.

### Task-specific schema verification

Only run this after the backend owner has approved `database/draw.sql` for the target environment:

```bash
mysql -u root -p -e "USE user_db; SHOW TABLES LIKE 'users';"
mysql -u root -p -e "USE superadmin_db; SHOW TABLES LIKE 'admin_audit_logs';"
```

Expected: one `users` table and one `admin_audit_logs` table.

> **Migration warning:** There are no versioned Task 2 up/down migrations or migration runner. `database/draw.sql` is a repository-wide bootstrap script, not a safe substitute for reviewed production migrations.

## 6. Redis, Queue, and External Services

### API Gateway

Mandatory for real Task 2 API calls. It is expected on port `8080`, validates JWT/RBAC, applies request rules/rate limits, and routes requests to downstream services.

Generic Gateway setup is already documented in:
`TaskImplementation/Superadmin Panel/Dependency/API_Gateway.md`

Task 2 specifically needs both `SUPERADMIN_GRPC_ADDR=localhost:50062` and `SESSION_GRPC_ADDR=localhost:50060` from the current Gateway env. Neither target has a runnable server implementation in the inspected repository.

### Authentication and authorization

A real request needs:

- A valid access token from Auth Service.
- Correct issuer, audience, algorithm, and JWKS configuration in Gateway.
- View permission for list/profile/session APIs.
- Stronger mutation permission for block/unblock.
- Service-side authorization; frontend menu/button hiding is only UX behavior.

### Redis

Current Gateway env has rate limiting enabled and `RATE_LIMIT_FAIL_OPEN=false`. Iska meaning: Redis unavailable hua to Gateway requests/startup fail ho sakte hain, depending on future implementation. Session Service separately uses Redis DB `2`.

### Message queues

No Kafka/RabbitMQ/NATS installation is required by the current Task 2 frontend. Architecture docs suggest a queue for future session event processing, but the user-management UI does not connect to one and no runnable integration was found.

## 7. Environment Variables

### New or changed Task 2 variables

None. Do not add database, Redis, MongoDB, JWT signing, or service credentials to the frontend `.env`.

Existing frontend env creation and security rules are documented in:

Refer:
`TaskImplementation/Superadmin Panel/task1_Dependency.md`

Section:
`7. Environment Variables`

### Critical base-path mismatch

Current files disagree:

- `.env.example` sets `VITE_API_BASE_URL=http://localhost:8080/api/v1`.
- Login calls relative path `/auth/login`, producing the correct `/api/v1/auth/login` URL.
- Task 2's `users-api.ts` calls paths already beginning with `/api/v1/...`.

Therefore the current example produces URLs like:

```text
http://localhost:8080/api/v1/api/v1/admin/users
http://localhost:8080/api/v1/api/v1/analytics/sessions
```

Setting the base URL to `http://localhost:8080` fixes Task 2 URLs but breaks the current login URL by removing `/api/v1`. There is no single `.env` value that makes both call styles correct.

> **Required code/config alignment before real integration:** Pick one convention. Recommended: keep `VITE_API_BASE_URL=http://localhost:8080/api/v1` and change feature paths to `/admin/users` and `/analytics/sessions`. This guide only documents the issue; it does not modify business/application code.

After any `.env` change, restart Vite. All `VITE_*` values are browser-visible, so secrets must remain backend-only.

### Backend credentials placement

Backend variables are currently in service `.env` files, but sanitized `.env.example` files and config loaders are missing. Required task-related groups are:

| Owner | Variable group | Notes |
|---|---|---|
| Gateway | `JWT_*`, `REDIS_*`, `SUPERADMIN_GRPC_ADDR`, `SESSION_GRPC_ADDR` | Required for auth, rate limit, and routing |
| Superadmin Service | `SUPERADMIN_DATABASE_DSN`, `USER_SERVICE_ADMIN_BASE_URL`, required/timeouts | User workflow and audit storage |
| User Service | `USER_SERVICE_DATABASE_DSN` or documented fallback | User DB credential; backend-only |
| Session Service | `SESSION_MONGO_*`, `SESSION_REDIS_*`, privacy salts/peppers | Session storage and privacy; backend-only |

Never commit real DSNs, Redis passwords, JWT keys, hash salts, or privacy peppers. Placeholder values such as `replace-this-with-a-long-random-secret` are insecure outside throwaway local development.

## 8. Docker Setup, Ports, and Networking

### Docker status

Task 2 introduces no new container, volume, network, health check, or restart policy. No supported Dockerfile or Compose file exists for this frontend/backend chain.

Reuse the current Docker limitation and optional frontend-only container guidance from:
`TaskImplementation/Superadmin Panel/task1_Dependency.md`, section `8. Docker Setup`.

Do not run an invented `docker compose up` command: no Compose manifest was found.

### Ports table

| Service | Port/address | Purpose | Task 2 status |
|---|---:|---|---|
| Vite frontend | `5173` | Local admin UI | Reused; runnable |
| Vite preview | `4173` commonly | Built bundle preview | Reused; optional |
| API Gateway HTTP | `8080` | Browser REST entry point | Reused; env only |
| Auth/JWKS HTTP | `8081` in Gateway JWKS URL | JWT public keys/auth design | Reused; runtime missing |
| Superadmin HTTP | `8088` | Env-only backend HTTP address | Reused; not directly called by UI |
| Superadmin gRPC | `50062` | Gateway user-management target | Task 2 required; server missing |
| User gRPC | `50052` | User-service target in Gateway env | Task 2 indirect; server missing |
| Session HTTP | `8086` | Session env-only HTTP server | Task 2 indirect; server missing |
| Session gRPC | `50060` | Gateway analytics target | Task 2 required; server missing |
| MySQL | `3306` | User and admin/audit databases | Reused |
| MongoDB | `27017` | Session database | Newly required by Task 2 session panel |
| Redis | `6379` | Gateway rate limit and Session cache | Reused/indirect |

### Networking checks

- Browser must reach the Gateway URL configured at build/dev time.
- Gateway must reach Auth JWKS, Superadmin gRPC, Session gRPC, and Redis.
- Superadmin Service must reach User Service and its database according to the eventual implementation.
- Session Service must reach MongoDB and Redis.
- Docker me `localhost` container ko khud refer karta hai; service names or an approved host bridge would be needed in a future Compose network.
- Gateway CORS must allow the exact frontend origin, normally `http://localhost:5173`, plus `Authorization`, `Content-Type`, and `X-Request-ID` headers.

## 9. Local Development Setup

### Step 1: Read reused setup first

Follow `TaskImplementation/Superadmin Panel/task1_Dependency.md` for clone, Node/pnpm, install, `.env`, and base frontend startup.

### Step 2: Go to the project directory

```bash
cd Ecommerce/frontend
```

### Step 3: Install dependencies

No Task 2 package is new:

```bash
pnpm install --frozen-lockfile
```

### Step 4: Configure environment

Create `.env` from the existing template as described in the previous guide. Before end-to-end testing, resolve the base-path mismatch documented in section 7; an env-only change cannot currently fix both login and Task 2 calls.

### Step 5: Database/service setup

- Frontend-only development: skip database, Redis, MongoDB, and migrations.
- Real integration: follow the owning dependency guides, then obtain runnable Gateway/Auth/Superadmin/User/Session service code from the backend team.
- Do not guess backend start commands from folder names; the required source/module files are absent.

### Step 6: Run Task 2 tests

```bash
cd frontend/superadmin-panel
pnpm exec vitest run tests/users
```

### Step 7: Run all quality checks

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
http://localhost:5173/admin/users
```

Without a valid stored admin session, the route redirects to `/login`. Without backend services, real login/API data will not work; use the automated tests to verify the frontend module independently.

### Full-stack verification when backend is supplied

Use a dedicated, non-production admin account and disposable test user.

```bash
export ADMIN_TOKEN='<temporary-admin-access-token>'
curl -i \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "X-Request-ID: task2-setup-check" \
  'http://localhost:8080/api/v1/admin/users?page=1&limit=25'
```

For a known test user:

```bash
export TEST_USER_ID='<disposable-test-user-id>'
curl -i \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "X-Request-ID: task2-session-check" \
  "http://localhost:8080/api/v1/analytics/sessions?user_id=$TEST_USER_ID&page=1&limit=20"
```

Do not paste tokens into shell history on shared machines. Do not test block/unblock against a real customer. Mutation verification must use an approved disposable account and must confirm an audit record was written.

### Expected UI permissions

| Role | View users/profile/sessions | Block/unblock |
|---|---:|---:|
| `superadmin` | Yes | Yes |
| `operations_admin` | Yes | Yes |
| `readonly_admin` | Yes in frontend | No |
| `finance_admin` | No | No |
| `catalog_admin` | No | No |

Backend must enforce the same matrix; frontend checks are not a security boundary.

## 11. Common Errors & Fixes

Generic pnpm, Node, port, CORS, Docker, and login errors are already covered in `task1_Dependency.md`, section `11. Common Errors & Fixes`.

Only Task 2-specific errors are listed below.

| Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| Request contains `/api/v1/api/v1/` | Base URL and Task 2 feature path both include prefix | Align path convention in code; see section 7 | Add integration URL tests using the real `.env.example` value |
| Login works but Users returns `404` | Same duplicated-prefix defect | Do not keep changing ports; inspect final Network-tab URL | Centralize API path construction |
| Tests pass but real API URL is wrong | Tests stub base URL as origin-only and do not test `.env.example` | Add a config-level test for login and Task 2 paths together | Test production-like config |
| Users page returns `502`/`Unavailable` | Gateway cannot reach Superadmin gRPC | Supply/start service and check `localhost:50062` | Add Gateway and service readiness checks |
| Session panel fails while profile loads | Session Service/Mongo/Redis or Gateway route is unavailable | Check Session target `50060`, Mongo `27017`, Redis `6379` | Separate health/readiness per dependency |
| User detail says not found although ID exists | Current detail lookup searches list endpoint using `q=user_id` | Backend search must support exact ID, or add a dedicated detail contract | Contract-test ID lookup semantics |
| `403` for `readonly_admin` | Frontend allows it, but `master-api.json` `admin` role list omits it | Align API contract and backend read-only permission policy | Keep one reviewed role matrix |
| Finance/catalog admin can mutate through direct API | Contract's broad `admin` class may be less strict than frontend | Enforce operation-specific permission server-side | Never rely on hidden buttons |
| Block/unblock button disabled | Audit reason is fewer than 10 trimmed characters | Enter a clear reason of at least 10 characters | Display and enforce matching backend validation |
| Status changes but audit record missing | Backend audit repository/transaction is absent or failed | Treat mutation as unsuccessful and investigate service logs | Make audit write mandatory and observable |
| Sessions are empty | No events linked to `user_id`, retention expired, or ingestion absent | Check approved test data and Session Service ingestion/retention | Seed non-sensitive integration fixtures |

## 12. Security & Best Practices

- Backend must authorize every view and mutation independently of the React UI.
- Apply least privilege: read-only admins must never receive mutation access.
- Require and persist actor ID, target user ID, old/new status, reason, timestamp, request ID, and outcome for each status mutation.
- Do not log Bearer tokens, passwords, raw DSNs, Redis passwords, session privacy salts, or full API payloads containing PII.
- Phone is masked by the current UI, but email is displayed in full. Backend and product policy should decide whether email must also be masked for read-only/operations roles.
- Session responses should exclude raw IP, unapproved device fingerprints, secrets, passwords, OTPs, card data, and private form text.
- Use TLS externally and mTLS or an equivalent trusted internal identity mechanism for production service traffic.
- Keep MongoDB, MySQL, and Redis private; never publish their ports to the internet.
- Use separate Redis logical DB/key prefixes and credentials where practical; logical DB numbers are not a security boundary.
- Add server-side reason length/content validation and an upper limit. Client-side minimum length can be bypassed.
- Preserve request IDs across Gateway and services so user-status incidents can be traced without exposing sensitive data.
- Use a disposable user for destructive QA and restore its state after testing.

## 13. Missing or Misconfigured Things

| Severity | Finding | Impact | Recommended fix |
|---|---|---|---|
| Critical | Frontend base/path convention duplicates `/api/v1` for Task 2 | Real users and sessions requests hit wrong URLs | Normalize all feature paths against one API base convention |
| Critical | Gateway, Auth, Superadmin, User, and Session runnable source is absent | Full-stack Task 2 cannot start from this repository | Add source, manifests, start commands, health checks, and service READMEs |
| High | Backend `.env` files exist instead of sanitized `.env.example` files | Credentials/default secrets may leak or drift | Keep placeholders in Git; inject real secrets securely and rotate exposed values |
| High | API contract admin-role set and frontend role matrix disagree | `readonly_admin` may be denied; broader roles may be over-authorized | Define endpoint-specific permissions in one authoritative contract |
| High | No proven backend granular authorization/audit implementation | Direct API calls may bypass UI rules or mutations may be unaudited | Enforce permissions and mandatory audit writes server-side |
| High | No versioned up/down migrations | Environment drift and unsafe deployment | Add migrations for `user_db`/`superadmin_db` changes and a documented runner |
| High | No gRPC proto/generated/server/client implementation | Gateway downstream calls cannot work | Add reviewed protobuf contract and generated code |
| Medium | User detail is implemented through list search | Profile lookup depends on undocumented search behavior | Add exact-ID endpoint or guarantee/test exact ID search |
| Medium | Session env contains placeholder privacy secrets | Predictable hashing/pseudonymization outside local dev | Require strong, secret-managed, environment-specific values |
| Medium | No Dockerfiles/Compose/health checks | Local and deployment runtime is not reproducible | Add deployment artifacts after service topology is implemented |
| Medium | Existing tests do not catch the real base-path combination | Unit tests can pass while integration fails | Add URL/config and Gateway contract integration tests |
| Low | `task2.md` lists unused `clsx` and MSW installs | Beginners may add unnecessary packages/lockfile churn | Treat actual imports and `package.json` as source of truth |

## 14. References to Previous Dependency Files

| Previous dependency file | Section/topic reused | Why reused |
|---|---|---|
| `task1_Dependency.md` | Tech stack, required software, pnpm, env, Docker, run commands, generic errors | Same frontend application and toolchain |
| `Dependency/Frontend.md` | Actual package inventory and package-guide mismatch | Task 2 adds no package |
| `Dependency/Environment.md` | Frontend/Gateway/backend env model | Same env loading and credential rules |
| `Dependency/API_Gateway.md` | Gateway responsibility, auth, Redis, and common API errors | Same public REST entry point |
| `Dependency/MySQL.md` | MySQL install, Docker option, credentials, verification | Same MySQL runtime; Task 2 only adds user/audit usage context |
| `Dependency/MongoDB.md` | MongoDB install and health verification | Same Session Service data store |
| `Dependency/Redis.md` | Redis install, Docker option, variables, and health check | Same Gateway/Session Redis instances |
| `Dependency/Migrations.md` | Migration gap and bootstrap limitations | No new Task 2 migration exists |
| `Dependency/gRPC.md` | Internal service communication and missing implementation | Same Gateway-to-service architecture |
| `Dependency/Go_Modules.md` | Missing backend module/workspace setup | Backend source remains unavailable |
| `Dependency/main_dependency.md` | Overall topology, ports, and missing-runtime audit | Same service-level foundation |

## 15. Final Checklist

### Frontend-only readiness

- [ ] Previous dependency documentation read
- [ ] Node.js 22.12+ and pnpm 10 available
- [ ] Workspace dependencies installed with frozen lockfile
- [ ] No unnecessary `clsx` or MSW package added
- [ ] Existing frontend `.env` created without secrets
- [ ] Task 2 user tests pass
- [ ] Full frontend typecheck passes
- [ ] Full frontend test suite passes
- [ ] Production build passes
- [ ] Vite starts and `/admin/users` route is registered
- [ ] Backend absence is not mistaken for a frontend install failure

### Full-stack readiness

- [ ] `/api/v1` base-path convention fixed and tested for login plus Task 2 APIs
- [ ] Runnable Gateway/Auth/Superadmin/User/Session services supplied
- [ ] Sanitized backend env templates supplied
- [ ] Strong local-only test credentials/secrets configured securely
- [ ] MySQL `user_db` and `superadmin_db` schema/migrations applied
- [ ] MongoDB `session_db` and required indexes/retention configured
- [ ] Redis reachable for Gateway and Session Service
- [ ] Gateway JWT/JWKS validation works
- [ ] Gateway CORS allows the exact frontend origin and required headers
- [ ] Superadmin and Session gRPC targets are healthy
- [ ] Backend role matrix matches the frontend, including `readonly_admin`
- [ ] Dedicated admin and disposable user test accounts provisioned
- [ ] User list/search/profile verified
- [ ] Session list/journey verified with privacy-safe data
- [ ] Block/unblock verified only on disposable data
- [ ] Mandatory audit record verified after mutation
- [ ] Logs checked using request IDs without exposing tokens/PII
- [ ] No duplicate setup documentation or unneeded dependency added

When the frontend checks pass, Task 2 UI/tooling setup ready hai. Full Task 2 readiness tabhi claim karein jab the missing backend runtime, URL alignment, service health, databases, authorization, and audit behavior are all verified.
