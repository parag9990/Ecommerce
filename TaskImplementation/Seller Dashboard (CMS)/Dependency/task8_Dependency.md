# Project Dependency & Setup Guide

## 1. Project Overview

### Variables used by this guide

| Variable | Value / derivation |
|---|---|
| `SERVICE_NAME` | Use the service folder name provided in the prompt |
| `TASK_FILE_NAME` | Use the implementation task file name provided in the prompt |
| `INPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `${TASK_FILE_NAME}` with `.md` replaced by `_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}` |

This file documents only dependency, setup, environment, DevOps, and verification needs for `INPUT_FILE_PATH`.

Simple Hinglish goal: current `SERVICE_NAME` modules ko reliable states milte hain: loading skeleton, empty state, failed state with retry/request id, permission denied state, unavailable state, and stale-data refresh warning. Is guide ka kaam business logic repeat karna nahi hai; sirf developer ko batana hai ki is polish layer ko run, configure, verify, and debug kaise karna hai.

### What changed in the current task

| Area | Status | Setup impact |
|---|---|---|
| Shared state components | New frontend code | No new package required |
| API error normalization | New frontend code | Reuses existing API Gateway envelope and request id pattern |
| HTTP timeout handling | Enhanced frontend code | Reuses existing `VITE_API_TIMEOUT_MS` |
| React Query retry behavior | Enhanced frontend config | No new service required |
| Permission gate integration | Reused and polished | Depends on existing seller roles/permissions from earlier tasks |
| Database, Redis, queues, Docker | Reused/no change | Follow previous dependency docs |
| Migrations | No new migration | Nothing new to run for this task |

Runtime flow:

```text
Browser
  -> frontend state component / module page
  -> React Query hook
  -> frontend HTTP client
  -> API Gateway through VITE_API_BASE_URL
  -> backend service
  -> success / empty / failed / permission response
  -> shared dashboard state UI
```

Important boundary:

- Current task is frontend reliability polish.
- It does not add a new backend service.
- It does not add a new database table.
- It does not add Redis/Kafka/RabbitMQ.
- It does not require Docker changes.
- Real 401/403/500 behavior still depends on API Gateway and backend services returning proper status codes and `request_id`.

## 2. Tech Stack

Most technology explanation is already documented. Follow the reused docs first:

Refer:
`TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md`

Section:
`What is this dependency`, `Installation steps`, `Required environment variables`, `Start commands`

Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Section:
`Tech Stack`, `Dependency Management`, `Environment Variables`

### Current task technology usage

| Technology | Required? | Simple explanation | Why used here | New setup? |
|---|---:|---|---|---|
| React | Yes | React UI components banane ke liye library hai. | Shared state components and module pages render karne ke liye. | No, reused |
| TypeScript | Yes | JavaScript me type safety add karta hai. | `AppError`, state props, and query data safer rakhne ke liye. | No, reused |
| Vite | Yes | Fast frontend dev server/build tool hai. | Local dev server `5174` and production build ke liye. | No, reused |
| pnpm workspace | Yes | Monorepo package install/run manage karta hai. | `seller-dashboard` app scripts run karne ke liye. | No, reused |
| React Router | Yes | Browser routes manage karta hai. | Permission denied links and protected seller routes ke liye. | No, reused |
| TanStack React Query | Yes | Server-state fetch/cache library hai. | Loading, retry, stale data, and refetch states handle karne ke liye. | No, reused |
| Zustand | Yes | Lightweight frontend store hai. | Active seller/permission context read karne ke liye. | No, reused |
| Tailwind CSS | Yes | Utility classes se UI style hota hai. | Skeletons, state shells, responsive layout polish ke liye. | No, reused |
| Lucide React | Yes | Icon library hai. | Empty, failed, denied, unavailable states me icons ke liye. | No, reused |
| Vitest + Testing Library | Recommended | Frontend unit/component tests ke tools hain. | `AsyncStateBoundary`, API error mapper, and permission UI verify karne ke liye. | No, reused |
| API Gateway | Required for real data | Browser REST calls ko backend services tak route karta hai. | 401/403/500, `request_id`, and envelope response real app me yahin se aate hain. | No, reused |
| MySQL | Required by earlier real modules | Relational DB hai, tables me data store karta hai. | Products/orders/offers/team/audit real data ke liye earlier modules need karte hain. | No change |
| Redis | Conditional | In-memory cache/rate-limit store hai. | Gateway rate limit enabled ho to required. | No change |

### Package status

No new runtime package is required for the current task. Actual `frontend/seller-dashboard/package.json` already contains the needed dependencies:

| Package | Current use |
|---|---|
| `@tanstack/react-query` | Query states, retries, stale data behavior |
| `react-router-dom` | State action links and protected routes |
| `zustand` | Seller store and permission context |
| `lucide-react` | UI state icons |
| `tailwindcss` / `@tailwindcss/vite` | State shell and skeleton styling |
| `vitest`, `@testing-library/react`, `@testing-library/user-event`, `jsdom` | Component tests |

Practical note: current code uses local helper `frontend/seller-dashboard/src/lib/cn.ts`; do not install `clsx` only because sample snippets in `INPUT_FILE_PATH` mention class composition.

## 3. Required Software

Base software setup is reused. Do not reinstall everything if previous setup already works.

| Software | Required for current task? | Follow this existing doc | Current task note |
|---|---:|---|---|
| Git | Yes | `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | Clone/pull repository. |
| Node.js | Yes | `TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md` | Frontend run/test/build. |
| Corepack + pnpm | Yes | `TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md` | Workspace package commands. |
| Browser | Yes | `TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md` | Manual state verification. |
| API Gateway | Yes for real API states | `TaskImplementation/${SERVICE_NAME}/Dependency/API_Gateway.md` | Real failed/permission responses. |
| Go toolchain | Backend only | `TaskImplementation/${SERVICE_NAME}/Dependency/Go_Modules.md` | Needed only when running backend services. |
| MySQL | Real earlier modules only | `TaskImplementation/${SERVICE_NAME}/Dependency/MySQL.md` | No new table for current task. |
| Redis | Conditional | `TaskImplementation/${SERVICE_NAME}/Dependency/Redis.md` | Only if Gateway rate limiting is enabled. |
| Docker | Optional | `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | No new container for current task. |

Beginner note: frontend-only verification ke liye Node.js + pnpm enough hai. Real permission/network/error responses dekhne ke liye API Gateway and backend services also running hone chahiye.

## 4. Dependency Management

### Frontend dependency system

This setup is already explained in:
`TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md`

Section:
`Installation steps`

Install only when `node_modules` is missing or lockfile changed:

```bash
cd frontend
pnpm install
```

Task-specific verification commands:

```bash
cd frontend
pnpm --filter seller-dashboard typecheck
pnpm --filter seller-dashboard test
pnpm --filter seller-dashboard build
```

Same commands also exist at workspace level:

```bash
cd frontend
pnpm typecheck:seller
pnpm test:seller
pnpm build:seller
```

### Backend dependency system

Current task does not add Go modules or backend dependencies.

Refer:
`TaskImplementation/${SERVICE_NAME}/Dependency/Go_Modules.md`

Section:
`Where it is used in project`, `Common errors and fixes`

Current task backend expectation:

| Expectation | Why it matters |
|---|---|
| Backend returns proper HTTP status codes | Frontend maps `401/403` to permission state and `5xx/network` to failed state. |
| Backend envelope includes `request_id` | Failed state can show safe debugging id. |
| Backend avoids stack traces in `error.message` | Frontend displays safe messages, but backend should also avoid leaking internals. |

## 5. Database Setup

No new database setup is introduced by the current task.

Reused database docs:

Refer:
`TaskImplementation/${SERVICE_NAME}/Dependency/MySQL.md`

Section:
`Installation steps`, `Docker setup, if possible`, `Required environment variables`

Refer:
`TaskImplementation/${SERVICE_NAME}/Dependency/Migrations.md`

Section:
`Local setup without Docker`, `Verify running commands`

### Current task database impact

| Database/service | Status | Explanation |
|---|---|---|
| MySQL `cms_db` | Reused | Earlier modules use CMS tables. Current task only changes how frontend shows empty/failed/permission states. |
| Order DB | Reused from earlier docs | Orders page state UI depends on existing order APIs if real backend is used. |
| MongoDB/RabbitMQ from product docs | Not newly introduced | If product module needs them locally, follow earlier task docs. Current task does not add them. |
| New migrations | None | Do not run or create a task-specific migration for current state polish. |

Credentials placement remains unchanged:

| File | Purpose |
|---|---|
| `backend/services/cms-service/.env` | CMS MySQL and CMS service settings reused from earlier tasks. |
| `backend/services/api-gateway/.env` | Gateway routing, JWT, Redis/rate-limit, and service target settings. |
| `frontend/seller-dashboard/.env.local` | Frontend API base URL and timeout settings. |

## 6. Redis / Queue / External Services

No new external service is added.

| Service | Required now? | Status | Follow this doc |
|---|---:|---|---|
| API Gateway | Yes for real API behavior | Reused | `TaskImplementation/${SERVICE_NAME}/Dependency/API_Gateway.md` |
| Redis | Conditional | Reused if Gateway rate limit enabled | `TaskImplementation/${SERVICE_NAME}/Dependency/Redis.md` |
| gRPC/Protobuf | Backend integration only | Reused | `TaskImplementation/${SERVICE_NAME}/Dependency/gRPC.md`, `TaskImplementation/${SERVICE_NAME}/Dependency/Protobuf.md` |
| Kafka | No | Not introduced | No current setup needed |
| RabbitMQ | No new setup | Only earlier product docs mention it | `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` if product flow requires it |
| MinIO/S3/SMTP/Stripe/Twilio/Firebase | No | Not used by current task | No current setup needed |

Current task external-service behavior:

- Network failure from API Gateway should render `FailedState`.
- `401` or `403` should render `PermissionDeniedState`.
- `429` should show a retry-later style safe message.
- Existing data should remain visible when a background refetch fails.

## 7. Environment Variables

No new environment variable is required.

This setup is already explained in:
`TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md`

Section:
`Required environment variables`

Reused frontend variables:

| Variable | Required? | Used by current task? | Purpose |
|---|---:|---:|---|
| `VITE_API_BASE_URL` | Optional | Yes | API Gateway base URL. Default is `http://localhost:8080`. |
| `VITE_API_TIMEOUT_MS` | Optional | Yes | Request timeout. Timeout maps to `REQUEST_TIMEOUT` failed state. |
| `VITE_LOGIN_URL` | Optional | Indirect | Login redirect flow if seller session is unavailable. |

Do not create a new full `.env` block for this task because the same frontend variables are already documented. If local API state testing fails, confirm only these existing values:

```bash
cd frontend/seller-dashboard
test -f .env.local && sed -n '1,80p' .env.local
```

Security notes:

- Never put secrets in `VITE_*` values. Vite exposes them to the browser bundle.
- `x-request-id` is safe to display for support/debugging.
- Do not display tokens, database passwords, cookies, or backend stack traces in UI state messages.

## 8. Docker Setup

No Dockerfile, compose service, port, volume, or network is added by the current task.

Reused Docker setup:

Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Section:
`Docker Setup`

Refer:
`TaskImplementation/${SERVICE_NAME}/Dependency/MySQL.md`

Section:
`Docker setup, if possible`

Refer:
`TaskImplementation/${SERVICE_NAME}/Dependency/Redis.md`

Section:
`Docker setup, if possible`

Current task Docker impact:

| Docker item | Change? | Note |
|---|---:|---|
| Frontend container | No | Existing docs also note no dedicated frontend Dockerfile was clearly found. |
| MySQL container | No | Reuse previous setup only if backend APIs need real data. |
| Redis container | No | Reuse only if Gateway rate limiting is enabled. |
| Docker network | No | No new service-to-service dependency. |
| Volumes | No | No new persistent storage. |
| Health checks | No new container health check | Verify frontend through Vite and tests. |

## 9. Local Development Setup

### Step 1: Read previous dependency documentation first

Read these first to avoid duplicate setup:

1. `TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md`
2. `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
3. `TaskImplementation/${SERVICE_NAME}/task6_Dependency.md`
4. `TaskImplementation/${SERVICE_NAME}/task7_Dependency.md`

### Step 2: Go to project directory

From repository root:

```bash
pwd
```

Expected project root should contain:

```text
frontend/
TaskImplementation/
docs/
```

### Step 3: Install only existing frontend dependencies

```bash
cd frontend
pnpm install
```

Do not add a new package for the current task unless a future code change actually imports it.

### Step 4: Configure only reused frontend env

If frontend needs API Gateway:

```bash
cd frontend/seller-dashboard
test -f .env.local || touch .env.local
```

Then keep values aligned with earlier docs:

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_API_TIMEOUT_MS=15000
```

These values are reused, not newly introduced.

### Step 5: Setup databases/services only if real APIs are needed

For component tests and frontend build:

```text
No MySQL, Redis, or backend service is required.
```

For real browser verification against backend:

1. Start MySQL/Redis/API Gateway/CMS/Auth services as explained in previous dependency docs.
2. Confirm `VITE_API_BASE_URL` points to the running Gateway.
3. Confirm backend returns correct `401`, `403`, `429`, `500`, and `request_id` values for error testing.

### Step 6: Run migrations if needed

No new migration is needed for current task.

If earlier modules are not bootstrapped yet, follow:

`TaskImplementation/${SERVICE_NAME}/Dependency/Migrations.md`

### Step 7: Start frontend

```bash
cd frontend
pnpm --filter seller-dashboard dev
```

Open:

```text
http://localhost:5174
```

### Step 8: Verify functionality related to `TASK_FILE_NAME`

Run automated checks:

```bash
cd frontend
pnpm --filter seller-dashboard typecheck
pnpm --filter seller-dashboard test
pnpm --filter seller-dashboard build
```

Manual checks:

| Scenario | Expected result |
|---|---|
| Initial query pending | Stable skeleton, no layout jump |
| API returns empty list | Helpful empty state, not failed state |
| API returns `500` | Failed state with retry and request id when available |
| API returns `401/403` | Permission denied state, not generic failed state |
| Background refetch fails with old data present | Old data remains visible plus refresh warning |
| Mobile viewport | State text and actions wrap without clipping |

## 10. Running the Project

### Ports and networking

Detailed networking setup is already documented in:
`TaskImplementation/${SERVICE_NAME}/task7_Dependency.md`

Section:
`Ports and networking`

Current task port table:

| Service | Port | Purpose | Status |
|---|---:|---|---|
| `SERVICE_NAME` Vite dev server | `5174` | Browser UI for state verification | Reused |
| `SERVICE_NAME` Vite preview | `4174` | Preview production build | Reused |
| API Gateway | `8080` | Real API responses and request ids | Reused |
| MySQL | `3306` | Earlier backend module data | Reused/no change |
| Redis | `6379` | Gateway rate limiting when enabled | Reused/no change |
| CMS gRPC | `9098` or `50059` depending env alignment | Backend service target | Reused/no change |

Port conflict guidance:

- If `5174` is busy, change Vite port in `frontend/seller-dashboard/vite.config.ts` or stop the existing process.
- If `8080` is busy, update Gateway `HTTP_ADDR` and frontend `VITE_API_BASE_URL` together.
- If CMS gRPC address mismatch appears, follow the alignment notes in `task6_Dependency.md` and `task7_Dependency.md`.

### Expected API response shape

Current task expects the frontend HTTP client to normalize the Gateway response envelope:

```json
{
  "data": {},
  "request_id": "req_example",
  "error": null
}
```

Error response example:

```json
{
  "data": null,
  "request_id": "req_failed_example",
  "error": {
    "code": "FORBIDDEN",
    "message": "Missing permission"
  }
}
```

Beginner note: frontend safe copy does not depend on raw backend message for common statuses. `401`, `403`, `404`, `429`, timeout, and `5xx` are mapped to friendly messages in `frontend/seller-dashboard/src/lib/api-error.ts`.

## 11. Common Errors & Fixes

Generic install, Docker, DB, Redis, and Gateway errors are already documented in previous dependency files. Current task-specific troubleshooting:

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| Empty data shows failed state | Module treats empty array as error | Use `AsyncStateBoundary` with correct `isEmpty` function | Keep empty checks explicit per module |
| `403` shows generic retry error | Error mapper not using `isPermissionError` | Route `401/403` to `PermissionDeniedState` | Test permission errors with component tests |
| Failed state has no request id | Backend did not return `request_id`, or HTTP client did not attach one | Confirm Gateway envelope and frontend `x-request-id` header | Keep request id in both request and response logs |
| Background refetch clears table | Page renders skeleton whenever `isFetching` is true | Use skeleton only for first load; keep stale data during refetch | Distinguish `isPending` from background `isFetching` |
| Timeout always happens | `VITE_API_TIMEOUT_MS` too low or Gateway too slow | Increase timeout locally or fix backend latency | Keep timeout value realistic, for example `15000` |
| Tests fail with router/link errors | State actions use `Link` but test lacks router provider | Wrap component tests in `MemoryRouter` | Use shared test render helper where possible |
| Permission text says `undefined` | Permission label missing for a permission key | Add label in permission utility before using gate | Keep permission types and labels in sync |
| Mobile UI clips request id | Long id not wrapped | Ensure `break-all` or equivalent is applied | Test failed states on narrow viewport |

## 12. Security & Best Practices

### Current task security rules

| Rule | Why it matters |
|---|---|
| Backend remains source of truth for authorization | Frontend permission checks are only UX; API must still enforce RBAC. |
| Show request id, not stack trace | Support can debug safely without exposing internals. |
| Do not show secrets in failed states | Browser UI is visible to users. |
| Do not store secrets in `VITE_*` env vars | Vite exposes them client-side. |
| Keep stale data during soft refresh failure | User should not lose usable data because refetch failed. |
| Do not fake analytics/orders/products | Empty/unavailable state is safer than misleading business data. |

### Best practices specific to current task

- Use `PermissionDeniedState` for access issues, not `FailedState`.
- Use module-specific empty copy. Example: orders should not say "Create order".
- Keep retry buttons keyboard-focusable.
- Keep icons decorative with `aria-hidden="true"` unless they provide unique meaning.
- Use `role="alert"` only for actual failed states; use polite status for non-critical states.
- Keep state components small and shared so modules do not invent different UX patterns.

## 13. Missing or Misconfigured Things

| Finding | Impact | Suggested fix |
|---|---|---|
| Frontend `.env.example` was not clearly found in previous docs | New developers may miss `VITE_API_BASE_URL` and timeout values | Add a non-secret `.env.example` later, reusing existing frontend env docs |
| Dedicated frontend Dockerfile was not clearly found | Containerized frontend local/prod run is not standardized | Add Dockerfile only when deployment target is finalized |
| Real error-state testing depends on backend envelopes | UI cannot show real request ids if Gateway/service does not return them | Keep API Gateway response envelope consistent |
| Permission labels/types must stay aligned | Missing labels can create confusing permission denied copy | Add tests for permission label coverage |
| Some task-guide sample libraries are not actual dependencies | Unnecessary installs can bloat the app | Follow actual `package.json`; current code uses local `cn` helper |

These are not blockers for frontend build/test, but they are useful cleanup items for smoother onboarding.

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md` | Installation steps, env vars, start commands | Same React/Vite/pnpm app is used. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | Tech stack, frontend env, Docker basics, MySQL/Redis overview | Base dashboard setup already documented. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/API_Gateway.md` | Gateway dependency and health checks | Real failed/permission states depend on Gateway responses. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/MySQL.md` | MySQL setup | No database change; earlier CMS data setup is reused. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/Redis.md` | Redis setup | Only needed if Gateway rate limiting is enabled. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/Go_Modules.md` | Backend dependency management | Current task adds no Go dependency. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/Migrations.md` | Migration setup | No new migration; earlier setup reused. |
| `TaskImplementation/${SERVICE_NAME}/task6_Dependency.md` | Team permissions, env alignment, ports | Current permission denied UX builds on earlier role/permission setup. |
| `TaskImplementation/${SERVICE_NAME}/task7_Dependency.md` | API response shape, request ids, Gateway/CMS alignment, ports | Current failed state and request id behavior reuse these conventions. |

## 15. Final Checklist

- [ ] Previous dependency documentation checked before using this guide.
- [ ] No duplicate install/database/Docker setup copied into current task docs.
- [ ] `frontend/seller-dashboard/package.json` checked; no new package needed.
- [ ] Existing frontend dependencies installed with `pnpm install` if needed.
- [ ] `VITE_API_BASE_URL` points to the intended API Gateway when real API testing is needed.
- [ ] `VITE_API_TIMEOUT_MS` is present or default timeout is acceptable.
- [ ] No new database, Redis, queue, Docker, or migration added for current task.
- [ ] Frontend typecheck passes.
- [ ] Frontend tests pass, especially `AsyncStateBoundary`.
- [ ] Frontend build passes.
- [ ] Empty, loading, failed, permission denied, unavailable, and soft refresh states manually verified.
- [ ] Request id appears in failed state when backend/Gateway provides one.
- [ ] Permission denied states do not leak restricted data.
- [ ] UI state copy is beginner-friendly, safe, and not blame-heavy.
- [ ] No secrets added to `VITE_*` variables or visible error messages.
