# Project Dependency & Setup Guide

## 1. Project Overview

### Variables used by this guide

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Seller Dashboard (CMS)` |
| `TASK_FILE_NAME` | `task6.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task6_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

### What Task 6 adds

`TASK_FILE_NAME` covers Team Permissions. Simple Hinglish me: seller owner ya manager dashboard ke andar staff ko invite karega, role assign karega, role update karega, disabled/invited/active status dekhega, aur UI permission ke basis par actions show ya hide karega.

Current implementation is mainly frontend-side:

| Area | Status | Evidence |
|---|---|---|
| Team route | Implemented | `frontend/seller-dashboard/src/routes/seller-routes.tsx` has `/seller/team` |
| Team nav item | Implemented | `frontend/seller-dashboard/src/layout/nav-items.ts` has `Team` |
| Team page | Implemented | `frontend/seller-dashboard/src/features/team/pages/team-members-page.tsx` |
| Team API client | Implemented frontend client only | `frontend/seller-dashboard/src/features/team/api/seller-team-api.ts` |
| Role and permission matrix | Implemented frontend helper | `frontend/seller-dashboard/src/features/team/utils/seller-permissions.ts` |
| Invite form validation | Implemented frontend validation | `frontend/seller-dashboard/src/features/team/utils/team-validation.ts` |
| Backend team endpoints | Missing / not clearly implemented | `api/master-api.json` does not list `/api/v1/seller/team...` endpoints |

Important boundary: Is dependency guide me business logic rewrite nahi kiya gaya. Backend Auth/CMS endpoints, invite token lifecycle, and email delivery are setup requirements for real data, but `INPUT_FILE_PATH` itself documents them as outside the Task 6 implementation scope.

## 2. Tech Stack

### Reused tech stack

Do not repeat full installation explanation here. Yeh setup already explain ho chuka hai:

Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/Frontend.md`

Section:
`Installation steps`, `Required environment variables`, `Start commands`

Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/main_dependency.md`

Section:
`All detected dependencies`, `Setup order`, `Ports table`

### Task 6 technology usage

| Technology | Required? | Task 6 use | New setup? |
|---|---|---|---|
| React | Yes | Team page, invite dialog, table, filters | No, reused |
| TypeScript | Yes | Role/status/API response safety | No, reused |
| Vite | Yes | Local seller-dashboard dev server | No, reused |
| React Router | Yes | `/seller/team` route | No, already in package |
| TanStack React Query | Yes | Staff list query and invite/update/disable/resend mutations | No, already in package |
| Zustand | Yes | Active seller context used by permission hook | No, already in package |
| React Hook Form | Yes | Invite staff form state | No, already in package |
| Zod and `@hookform/resolvers` | Yes | Email and assignable-role validation | No, already in package |
| Lucide React | Yes | Team and action icons | No, already in package |
| Tailwind CSS Vite plugin | Yes | Dashboard styling | No, already in package |
| Vitest + Testing Library + jsdom | Yes for tests | Team API normalizers, permission helpers, table interactions | No, already in package |
| MSW | Optional only | `TASK_FILE_NAME` mentions mock handlers, but current package does not install or use MSW | No install needed unless future integration tests require it |
| `clsx` | Not required currently | `TASK_FILE_NAME` mentions it, but code uses local `cn()` helper instead | Do not install unless code changes |

No new package install is required for the current Task 6 code because `frontend/seller-dashboard/package.json` already contains the actual libraries used by the implementation.

## 3. Required Software

Shared required software is reused:

Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`Required Software`

Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/Frontend.md`

Section:
`Installation steps`

For Task 6 specifically, beginner developer ko these tools available chahiye:

| Software | Why needed for Task 6 | Status |
|---|---|---|
| Node.js | Frontend app run/build/test karne ke liye | Reused |
| pnpm | Frontend workspace dependency manager | Reused |
| Browser | `/seller/team` manual verification ke liye | Reused |
| MySQL client | `seller_staff` table verify/seed karne ke liye | Reused |
| Redis | Only if API Gateway rate limit enabled hai | Reused |
| Go toolchain | Backend Auth/CMS/Gateway implementation run karne ke liye, once source exists | Reused / backend source incomplete |

## 4. Dependency Management

### Frontend dependencies

Task 6 uses the existing pnpm workspace. Full pnpm explanation already exists:

Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/Frontend.md`

Section:
`Installation steps`

Only Task 6-specific note:

```bash
cd frontend
pnpm --filter seller-dashboard typecheck
pnpm --filter seller-dashboard test
pnpm --filter seller-dashboard build
```

These commands validate the Team Permissions frontend code together with the rest of seller-dashboard.

### Backend dependencies

Backend dependency setup is still incomplete in this checkout:

Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/Go_Modules.md`

Section:
`Where it is used in project`, `Common errors and fixes`

Task 6 backend reality:

| Finding | Impact |
|---|---|
| `backend/go.work` exists, but no service-level `go.mod` files were found | Backend cannot be built from this checkout as-is |
| `backend/services/cms-service/.env` exists, but CMS Go source was not clearly found | Team APIs need backend implementation before real data works |
| `backend/services/api-gateway/.env` exists, but gateway Go source was not clearly found | REST route wiring cannot be verified locally from source |

## 5. Database Setup

### Reused database: MySQL

MySQL setup, Docker setup, credentials, and migration basics are already documented.

Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/MySQL.md`

Section:
`Installation steps`, `Docker setup, if possible`, `Required environment variables`

Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/Migrations.md`

Section:
`Local setup without Docker`, `Verify running commands`

### Task 6 table requirement: `seller_staff`

Task 6 depends on the CMS `seller_staff` table.

Simple explanation: `seller_staff` table seller ke staff members ko store karti hai. Isme role, status, seller id, user id, invite source, and timestamps hote hain. Team page isi data ko staff list, status badges, and role controls ke form me show karta hai.

Evidence:

| Source | What it confirms |
|---|---|
| `database/draw.sql` | `seller_staff` table exists with `staff_id`, `seller_id`, `user_id`, `role`, `status`, `invited_by` |
| `frontend/seller-dashboard/src/features/team/types.ts` | Frontend expects matching `SellerStaffMember` fields |
| `frontend/seller-dashboard/src/features/team/api/seller-team-api.ts` | API normalizer rejects invalid role/status/member response |

Required or optional: Required for real backend Team Permissions. Without this table, `/seller/team` UI can load but real staff data cannot work.

Default port: MySQL uses `3306`, already reused.

Credentials placement:

| File | Variables |
|---|---|
| `backend/services/cms-service/.env` | `CMS_MYSQL_DSN` or `CMS_DB_HOST`, `CMS_DB_PORT`, `CMS_DB_NAME`, `CMS_DB_USER`, `CMS_DB_PASSWORD` |

Task-specific local verification:

```bash
mysql -u cms_user -p -D cms_db -e "SHOW INDEX FROM seller_staff;"
mysql -u cms_user -p -D cms_db -e "SELECT staff_id, seller_id, user_id, role, status FROM seller_staff LIMIT 5;"
```

Optional local seed for manual UI testing, only after matching seller and user ids exist:

```sql
INSERT INTO seller_staff (
  staff_id,
  seller_id,
  user_id,
  role,
  status,
  invited_by
) VALUES (
  'staff_local_001',
  'seller_local_001',
  'user_local_001',
  'seller_manager',
  'active',
  'user_owner_001'
);
```

Note: Seed data ko production migration me mat daalo. Yeh sirf local testing ke liye hai.

## 6. Redis / Queue / External Services

### Redis

Redis Task 6 ka direct dependency nahi hai. API Gateway rate limiting enabled ho to Redis required hai.

Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/Redis.md`

Section:
`Installation steps`, `Start commands`, `Common errors and fixes`

### API Gateway

API Gateway required hai because frontend `http.ts` browser requests ko `VITE_API_BASE_URL` par bhejta hai. Default base URL `http://localhost:8080` hai.

Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/API_Gateway.md`

Section:
`Required environment variables`, `Verify running commands`

Task 6-specific routes expected by frontend:

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/api/v1/seller/team?page=1&page_size=20&status=active` | Staff list with filters |
| `POST` | `/api/v1/seller/team/invites` | Invite staff |
| `PATCH` | `/api/v1/seller/team/{staff_id}/role` | Update staff role |
| `PATCH` | `/api/v1/seller/team/{staff_id}/status` | Disable staff |
| `POST` | `/api/v1/seller/team/{staff_id}/resend-invite` | Resend invite |

Current gap: These team routes are not clearly listed in `api/master-api.json`. Backend/Gateway route registration must be added or confirmed before real API verification can pass.

### CMS Service

CMS Service required hai for real seller staff data.

Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/CMS_Backend.md`

Section:
`Why this service uses it`, `Required environment variables`, `Common errors and fixes`

Task 6-specific CMS responsibilities:

- Read staff list for active seller.
- Create invited staff row.
- Update assignable staff role.
- Disable staff member.
- Return response shape expected by frontend.
- Enforce seller ownership and permission server-side.

### Auth / Session boundary

Task 6 needs seller session data because `useSellerPermissions()` reads active seller roles from the seller session.

Frontend expects `/api/v1/seller/session` to return an active seller with at least one of:

- `roles`
- `staff_role`
- `role`

Task-specific allowed role values:

```text
seller
seller_manager
seller_catalog_editor
seller_order_manager
```

If role data missing hai, current permission hook falls back to `seller`. That is acceptable for local UI resilience, but backend must never trust frontend fallback. Real authorization API Gateway + Auth/CMS side enforce karega.

### Notification / Email service

Invite email delivery is not implemented in current Task 6 frontend code, but real invite flow normally needs Notification Service or SMTP.

Status: Future backend requirement.

Setup note: If invite email is implemented later, add sanitized env variables for provider credentials, sender address, invite link base URL, token expiry, and retry policy. Do not put those secrets into frontend `VITE_*` variables.

### Kafka / RabbitMQ / MinIO / Elasticsearch

No new Kafka, RabbitMQ, MinIO, Elasticsearch, Stripe, Twilio, or Firebase requirement was detected for Task 6.

## 7. Environment Variables

### Reused env docs

Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/Environment.md`

Section:
`Required environment variables`, `Common errors and fixes`, `Security notes`

### Task 6 new or changed variables

No new frontend environment variable is required for Task 6.

Task 6 reuses:

```text
VITE_API_BASE_URL=http://localhost:8080
VITE_API_TIMEOUT_MS=15000
```

Where to place frontend env:

```text
frontend/seller-dashboard/.env.local
```

Task 6 backend env is also reused:

| File | Important reused variables |
|---|---|
| `backend/services/api-gateway/.env` | `HTTP_ADDR`, `API_BASE_PATH`, `CMS_GRPC_ADDR`, `USER_GRPC_ADDR`, `JWT_JWKS_URL`, `REDIS_ADDR` |
| `backend/services/cms-service/.env` | `CMS_GRPC_ADDR`, `CMS_MYSQL_DSN`, `CMS_DB_*`, `CMS_INTERNAL_AUTH_TOKEN` |
| `backend/services/auth-service/.env` | JWT/session/RBAC-related settings |

Important current mismatch:

| File | Value |
|---|---|
| `backend/services/cms-service/.env` | `CMS_GRPC_ADDR=:9098` |
| `backend/services/api-gateway/.env` | `CMS_GRPC_ADDR=localhost:50059` |

Fix one side before testing Team APIs:

```text
# Option A: keep CMS service on 9098 and update gateway target
CMS_GRPC_ADDR=localhost:9098

# Option B: keep gateway target 50059 and make CMS listen on 50059
CMS_GRPC_ADDR=:50059
```

Security notes:

- `VITE_*` values browser bundle me visible hote hain, so secrets kabhi mat daalo.
- DB password and internal auth token sirf backend `.env` or secret manager me rakho.
- Real `.env` files share karne se pehle sanitize karo.

## 8. Docker Setup

Docker setup for this service is already audited as missing/incomplete.

Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/main_dependency.md`

Section:
`Docker setup`

Refer:
`TaskImplementation/{SERVICE_NAME}/task5_Dependency.md`

Section:
`Docker Setup`

Task 6 does not introduce a new container, volume, network, or port. It reuses the same expected stack:

| Container / service | Task 6 need | Status |
|---|---|---|
| Seller Dashboard frontend | Run UI | Reused |
| API Gateway | Expose seller team REST endpoints | Reused, routes missing/need confirmation |
| CMS Service | Own `seller_staff` data | Reused, source missing/need implementation |
| Auth/User/Session service | Provide active seller and roles | Reused |
| MySQL | Store CMS `seller_staff` rows | Reused |
| Redis | Gateway rate limiting if enabled | Reused |
| Notification service | Needed only for real email invite delivery | Future / optional for current UI |

No Docker command is repeated here. Follow the existing Docker guidance first, then verify Task 6 routes after backend is available.

## 9. Local Development Setup

### Step 1: Read previous dependency documentation first

Follow these in order:

1. `TaskImplementation/{SERVICE_NAME}/Dependency/main_dependency.md`
2. `TaskImplementation/{SERVICE_NAME}/Dependency/Frontend.md`
3. `TaskImplementation/{SERVICE_NAME}/Dependency/Environment.md`
4. `TaskImplementation/{SERVICE_NAME}/Dependency/MySQL.md`
5. `TaskImplementation/{SERVICE_NAME}/Dependency/API_Gateway.md`
6. `TaskImplementation/{SERVICE_NAME}/Dependency/CMS_Backend.md`

### Step 2: Go to project directory

Documentation for this task lives here:

```bash
cd "TaskImplementation/{SERVICE_NAME}"
```

Frontend app runs from:

```bash
cd frontend
```

### Step 3: Install only new dependencies if any

No new dependencies are required for Task 6. If dependencies are missing locally, use the reused frontend install:

```bash
cd frontend
pnpm install
```

### Step 4: Setup databases/services

Reused services:

- MySQL with CMS schema from `database/draw.sql`
- Redis if gateway rate limit is enabled
- API Gateway
- CMS Service
- Auth/User/Session boundary

Task 6-specific DB check:

```bash
mysql -u cms_user -p -D cms_db -e "SHOW TABLES LIKE 'seller_staff';"
```

### Step 5: Add only new or changed environment variables

No new env variable is required. Confirm reused values:

```text
frontend/seller-dashboard/.env.local
VITE_API_BASE_URL=http://localhost:8080
VITE_API_TIMEOUT_MS=15000
```

Then align Gateway and CMS `CMS_GRPC_ADDR` as explained in Section 7.

### Step 6: Run migrations if needed

Dedicated migrations are not found. Local fallback remains:

```bash
mysql -u root -p < database/draw.sql
```

Production-ready fix: create proper CMS up/down migrations for `seller_staff` and other CMS tables.

### Step 7: Start backend services

Current backend source is incomplete in this checkout. Once backend code exists, expected startup order:

1. MySQL
2. Redis if enabled
3. Auth/User/Session service
4. CMS Service
5. API Gateway
6. Seller Dashboard frontend

Task 6 backend smoke checks after implementation:

```bash
curl http://localhost:8080/health/live
curl "http://localhost:8080/api/v1/seller/team?page=1&page_size=20"
```

### Step 8: Start Seller Dashboard frontend

```bash
cd frontend
pnpm --filter seller-dashboard dev
```

Open:

```text
http://localhost:5174/seller/team
```

### Step 9: Verify Task 6 functionality

Manual checks:

- Login/session returns an active seller.
- Active seller role is `seller` or `seller_manager` for team controls.
- `/seller/team` renders Team permissions page.
- Staff status cards show invited/active/disabled counts.
- Staff table loads from `/api/v1/seller/team`.
- Status filter sends `status` query only when not `all`.
- Invite dialog validates email and role.
- Invite form does not allow owner role.
- Role update only works for assignable non-owner roles.
- Disabled staff cannot be edited.
- Resend invite appears only for invited staff.

Automated checks:

```bash
cd frontend
pnpm --filter seller-dashboard typecheck
pnpm --filter seller-dashboard test
```

## 10. Running the Project

### Ports and networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Seller Dashboard Vite dev server | `5174` | Browser UI | Reused |
| Seller Dashboard preview | `4174` | Built frontend preview | Reused |
| API Gateway HTTP | `8080` | Frontend REST entrypoint | Reused |
| CMS Service HTTP | `8087` | CMS health/HTTP if implemented | Reused |
| CMS Service gRPC | `9098` | CMS service listen address | Reused, must align |
| Gateway CMS target | `50059` currently | Gateway to CMS target | Mismatch risk |
| MySQL | `3306` | CMS database | Reused |
| Redis | `6379` | Gateway rate limiting | Reused |

Networking notes:

- Browser calls `VITE_API_BASE_URL`, default `http://localhost:8080`.
- `credentials: include` is used by the frontend HTTP client, so cookies and CORS must be configured correctly.
- If services run inside Docker Compose later, use service DNS names like `cms-service:9098`; container-to-container calls should not use host `localhost`.
- If `5174` is busy, change `frontend/seller-dashboard/vite.config.ts` or run Vite on another port.
- If `8080` is busy, change Gateway `HTTP_ADDR` and frontend `VITE_API_BASE_URL` together.

## 11. Common Errors & Fixes

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `/seller/team` redirects to login | `/api/v1/seller/session` returns 401 or session backend unavailable | Login again, start session/auth backend, verify cookies | Keep session backend running before frontend testing |
| Team page shows permission denied | Active seller missing, inactive, or role lacks `team:view` | Use role `seller` or `seller_manager`; verify seller session response | Seed valid `seller_staff` and backend RBAC data |
| Staff list returns 404 | `/api/v1/seller/team` route not registered in API Gateway/master API | Add/confirm Gateway route and CMS handler | Keep frontend-used routes in API contract |
| Staff list network error | API Gateway not running or `VITE_API_BASE_URL` wrong | Start gateway or update `.env.local` | Keep frontend env aligned with gateway port |
| Gateway returns CMS unavailable | CMS service down or `CMS_GRPC_ADDR` mismatch | Align CMS and Gateway gRPC addresses | Maintain one source of truth for service ports |
| `seller_staff` table missing | CMS schema not applied | Apply schema or migration | Use versioned migrations for CMS schema |
| Invite fails with validation error | Email invalid or role is not assignable | Enter valid email and choose manager/catalog/order role | Keep frontend and backend role enums aligned |
| `Owner role cannot be assigned...` | Frontend blocks assigning `seller` owner role | Choose `seller_manager`, `seller_catalog_editor`, or `seller_order_manager` | Never allow owner invite from dashboard UI |
| `Invite response did not include a valid staff member` | Backend response missing `staff_id`, `email`, valid `role`, or valid `status` | Return normalized staff member from API | Add contract tests for team endpoints |
| Status filter does nothing | Backend ignores `status` query param | Implement status filter in CMS query | Test `invited`, `active`, `disabled`, and `all` cases |
| Cookies not sent | CORS/same-site credentials not configured | Configure Gateway CORS and cookie settings | Remember frontend uses `credentials: include` |

## 12. Security & Best Practices

Task 6-specific security checklist:

- Frontend permission checks are UX only. Backend must enforce `team:view`, `team:invite`, `team:update_role`, and `team:disable`.
- Do not allow seller staff to manage another seller's team. Every query/mutation must be scoped by authenticated seller id.
- `seller` owner role should not be assignable from invite or role update APIs.
- Disable action should never delete staff rows silently. Keep status history/audit trail.
- Invite tokens should be random, short-lived, hashed server-side, and single-use.
- Resend invite should be rate-limited to avoid email abuse.
- Role changes and disable actions should produce audit events for Task 7.
- Do not expose DB credentials, internal auth tokens, invite token secrets, or SMTP credentials in frontend env.
- Return stable error codes without leaking sensitive internal details.
- Keep frontend role matrix aligned with backend RBAC policy, but treat backend as source of truth.

## 13. Missing or Misconfigured Things

| Area | Finding | Impact | Suggested fix |
|---|---|---|---|
| API contract | Team endpoints are not clearly listed in `api/master-api.json` | Frontend may call routes that Gateway does not expose | Add `/api/v1/seller/team...` contract entries |
| Backend source | CMS/Gateway Go source and service `go.mod` files not clearly found | Real Team APIs cannot be run from this checkout | Add service modules, handlers, usecases, repositories |
| gRPC port alignment | CMS listens on `:9098`, Gateway points to `localhost:50059` | Gateway cannot reach CMS | Align `CMS_GRPC_ADDR` values |
| Migrations | Dedicated CMS up/down migrations not found | Local schema bootstrap is manual, rollback unsafe | Add versioned migrations |
| `.env.example` | Sanitized env examples not found | Beginners may copy real `.env` values by mistake | Add `.env.example` files with placeholder values |
| Backend authorization | Not clearly implemented for team actions | Frontend-only guard can be bypassed | Enforce permissions server-side |
| Invite email delivery | Notification/SMTP setup not documented for Task 6 runtime | Real invite may create row but not deliver email | Add Notification Service integration when backend invite flow is built |
| API response contract | Frontend expects normalized staff member fields | Bad response shape creates UI errors | Add backend contract tests for list/invite/update/disable/resend |
| Audit linkage | Task 6 changes should appear in Task 7 audit timeline | Team actions may be invisible later | Emit audit events for invite, role update, disable, resend |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Required Software, Dependency Management, MySQL, Redis, API Gateway, Environment Variables | Base seller dashboard setup already documented |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | Frontend dependency reuse, API Gateway, role-based frontend behavior | Same pnpm/frontend/API Gateway pattern |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | Ports, frontend env, gateway verification, permission errors | Same frontend and gateway runtime flow |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | CMS Service, MySQL CMS tables, CMS/Gateway gRPC alignment | Task 6 is also CMS-backed |
| `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md` | Analytics-era CMS env, Docker notes, ports, missing backend audit | Same backend/CMS/Gateway setup state |
| `TaskImplementation/{SERVICE_NAME}/Dependency/main_dependency.md` | Overall setup order, detected dependencies, ports table, missing implementation audit | Central dependency summary |
| `TaskImplementation/{SERVICE_NAME}/Dependency/Frontend.md` | pnpm install, Vite ports, frontend env, frontend commands | Task 6 adds frontend module only |
| `TaskImplementation/{SERVICE_NAME}/Dependency/Environment.md` | Env files, required vars, security notes | Task 6 has no new env vars |
| `TaskImplementation/{SERVICE_NAME}/Dependency/MySQL.md` | MySQL install/Docker/credentials | `seller_staff` reuses CMS MySQL |
| `TaskImplementation/{SERVICE_NAME}/Dependency/Migrations.md` | Schema bootstrap and migration gaps | `seller_staff` needs migrations |
| `TaskImplementation/{SERVICE_NAME}/Dependency/API_Gateway.md` | Gateway env, health checks, common errors | Team APIs must go through Gateway |
| `TaskImplementation/{SERVICE_NAME}/Dependency/CMS_Backend.md` | CMS backend purpose, env, missing source status | Team permissions are CMS-owned |
| `TaskImplementation/{SERVICE_NAME}/Dependency/Redis.md` | Redis setup | Gateway rate limiting may require Redis |
| `TaskImplementation/{SERVICE_NAME}/Dependency/gRPC.md` | gRPC env and port mismatch troubleshooting | Gateway-to-CMS calls require gRPC alignment |

## 15. Final Checklist

- [x] Previous dependency documentation checked.
- [x] `INPUT_FILE_PATH` analyzed.
- [x] Existing frontend Task 6 files inspected.
- [x] Previous setup duplication avoided.
- [x] No original implementation file modified.
- [x] No business logic rewritten.
- [x] New package requirement checked.
- [x] New environment variable requirement checked.
- [x] MySQL `seller_staff` requirement documented.
- [x] Redis/Gateway/CMS/Auth dependencies marked as reused.
- [x] Task-specific ports and networking notes added.
- [x] Missing team backend endpoints documented.
- [x] CMS/Gateway gRPC mismatch documented.
- [x] Security and permission enforcement notes added.
- [x] References to previous dependency files included.
- [ ] Backend `/api/v1/seller/team...` routes implemented and verified.
- [ ] CMS team repository/usecase/handler implemented.
- [ ] Auth/session role contract confirmed.
- [ ] Notification invite delivery implemented if real emails are required.
- [ ] CMS migrations added with up/down files.
- [ ] Sanitized `.env.example` files added.
- [ ] `pnpm --filter seller-dashboard typecheck` run successfully.
- [ ] `pnpm --filter seller-dashboard test` run successfully.
- [ ] Manual `/seller/team` browser verification completed.
