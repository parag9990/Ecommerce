# Project Dependency & Setup Guide

Input task file:

```text
TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}
```

Generated dependency file:

```text
TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}
```

This guide explains only the dependency, setup, environment, and DevOps impact of Task 3 auth screens. It avoids repeating base frontend setup already documented in earlier dependency files.

Important beginner note:

- Task 3 ka focus login, signup, OTP verify, forgot password, and reset password screens hai.
- Frontend browser app direct database, Redis, Kafka, RabbitMQ, SMTP, or SMS provider se connect nahi karta.
- Real auth flow test karne ke liye REST API Gateway and backend Auth Service running hone chahiye.
- Static UI and client-side validation test karne ke liye only frontend dev server enough hai.

---

## 1. Project Overview

Task 3 auth feature user-facing authentication screens and client-side validation add karta hai.

Simple Hinglish:

"Auth screens ka matlab woh pages jahan buyer account create karta hai, login karta hai, OTP verify karta hai, password bhoolne par reset flow start karta hai, aur new password set karta hai."

### Task-specific files checked

| File | Purpose |
|---|---|
| `frontend/user-app/src/features/auth/pages/login-page.tsx` | Login form, safe redirect, login mutation. |
| `frontend/user-app/src/features/auth/pages/signup-page.tsx` | Buyer signup form, password strength, signup OTP challenge start. |
| `frontend/user-app/src/features/auth/pages/otp-page.tsx` | OTP verification and resend OTP flow. |
| `frontend/user-app/src/features/auth/pages/forgot-password-page.tsx` | Password reset OTP challenge start. |
| `frontend/user-app/src/features/auth/pages/reset-password-page.tsx` | OTP plus new password reset flow. |
| `frontend/user-app/src/features/auth/api/auth.api.ts` | Typed REST functions for auth endpoints. |
| `frontend/user-app/src/features/auth/schemas.ts` | Zod validation schemas for auth forms. |
| `frontend/user-app/src/features/auth/types.ts` | Auth request and response TypeScript types. |
| `frontend/user-app/src/features/auth/hooks/use-login-mutation.ts` | React Query login mutation and session cache update. |
| `frontend/user-app/src/features/auth/hooks/use-logout-mutation.ts` | Logout mutation, session cleanup, query cache cleanup. |
| `frontend/user-app/src/lib/http.ts` | REST fetch wrapper using `VITE_API_BASE_URL`, credentials, bearer token, and normalized errors. |
| `frontend/user-app/src/lib/auth-session.ts` | Browser session storage for access-token metadata and auth hint. |
| `frontend/user-app/src/routes/route-paths.ts` | Auth route constants. |
| `frontend/user-app/src/routes/index.tsx` | Auth route wiring under app shell. |
| `frontend/user-app/package.json` | Auth form dependencies and scripts. |
| `frontend/user-app/.env.example` | Existing public frontend env variables. |

### What is new for this task

| Area | New / Reused | Notes |
|---|---|---|
| Auth screens | New | Login, signup, OTP, forgot password, reset password. |
| Auth validation schemas | New | `zod` schemas for email/phone, password, OTP, confirm password, terms. |
| Form state library usage | Task-specific | `react-hook-form` powers auth form state. |
| Zod resolver | Task-specific | `@hookform/resolvers` connects schemas to forms. |
| Auth REST endpoints | Task-specific | Public auth API calls through API Gateway. |
| Session handling | Task-specific / later-task linked | Login persists access token metadata and updates auth store/query cache. |
| Database setup | Reused | No direct frontend DB. Backend Auth Service may need MySQL/Redis. |
| Docker setup | Reused | No new frontend Dockerfile or compose service detected for Task 3. |
| Environment variables | Reused | No new frontend env variable added; existing `VITE_API_BASE_URL` is required for real auth. |

---

## 2. Tech Stack

Base frontend technologies are already explained in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Sections:

```text
1. Project Tech Stack Analysis
2. Node.js Dependency System
```

Routing and app-shell dependencies are already explained in:

```text
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
```

Sections:

```text
2. Tech Stack
4. Dependency Management
```

### Task 3 specific technologies

| Technology | Required? | Where Used | Beginner-Friendly Explanation |
|---|---:|---|---|
| `react-hook-form` | Required | Auth pages | Form state manage karta hai: input value, touched state, validation error, submit loading. Isse forms fast and clean rehte hain. |
| `zod` | Required | `src/features/auth/schemas.ts` | Zod validation library hai. Ye check karta hai ki email, phone, OTP, and password valid format me hain. |
| `@hookform/resolvers` | Required | Auth pages | Ye bridge hai jo Zod schema ko React Hook Form ke validation system se connect karta hai. |
| React Query | Required by current auth hooks | `use-login-mutation.ts`, `use-logout-mutation.ts` | Server mutation state manage karta hai: loading, success, error, cache update. Full explanation previous docs me hai. |
| Zustand | Required by current auth store | `src/stores/auth-store.ts` | Browser app ka lightweight state store hai. Auth user info persist karne ke liye use hota hai. Full explanation previous docs me hai. |
| Fetch API | Required | `src/lib/http.ts` | Browser built-in HTTP client hai. Auth requests API Gateway ko bhejne ke liye use hota hai. |
| REST API Gateway | Required for real auth | `VITE_API_BASE_URL` | Frontend auth request pehle gateway ko jaati hai, gateway backend Auth Service ko route karta hai. |
| Backend Auth Service | Required for real auth | Behind API Gateway | Signup, login, OTP, password reset, logout ka real source of truth backend hai. |
| Vitest + Testing Library | Required for existing tests | `login-page.test.tsx`, `schemas.test.ts` | Form validation and login page behavior test karne ke liye use hota hai. |

### Auth API contract used by Task 3

| Screen / Action | Method | Endpoint | Auth | Backend Service |
|---|---|---|---|---|
| Signup | `POST` | `/api/v1/auth/signup` | Public | `auth-service` |
| Login | `POST` | `/api/v1/auth/login` | Public | `auth-service` |
| Logout | `POST` | `/api/v1/auth/logout` | Buyer auth | `auth-service` |
| Send OTP | `POST` | `/api/v1/auth/otp/send` | Public | `auth-service` |
| Verify OTP | `POST` | `/api/v1/auth/otp/verify` | Public | `auth-service` |
| Forgot password | `POST` | `/api/v1/auth/password/forgot` | Public | `auth-service` |
| Reset password | `POST` | `/api/v1/auth/password/reset` | Public | `auth-service` |

Note:

- `/api/v1/auth/refresh` exists in the API contract, but Task 3 pages do not directly implement refresh-token flow.
- Frontend validation UX ke liye hai. Backend validation and authorization final authority rahega.

---

## 3. Required Software

No new system software was introduced by Task 3.

Follow the base setup from:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Sections:

```text
2. Node.js Dependency System
8. Complete Project Run Instructions
```

Minimum tools:

| Software | Required? | Why |
|---|---:|---|
| Git | Required | Repository clone/pull ke liye. |
| Node.js `>=22.13.0` | Required | Vite, TypeScript, ESLint, Vitest run karne ke liye. |
| pnpm `>=11.5.0` | Required | Frontend workspace dependency install ke liye. |
| Browser | Required | Auth screens manually test karne ke liye. |
| Docker | Optional for Task 3 UI | Backend infra locally run karne ke liye useful, but frontend UI ke liye mandatory nahi. |
| API Gateway + Auth Service | Required for real auth | Actual login/signup/OTP/password-reset calls ke liye. |

---

## 4. Dependency Management

This is a Node.js + pnpm workspace frontend.

Do not duplicate base pnpm setup here. Refer:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Sections:

```text
2. Node.js Dependency System
Quick Command Summary
```

### Package files involved

| File | Purpose |
|---|---|
| `frontend/package.json` | Frontend workspace scripts and Node/pnpm engine requirement. |
| `frontend/pnpm-workspace.yaml` | Workspace package discovery. |
| `frontend/pnpm-lock.yaml` | Exact dependency lockfile. |
| `frontend/user-app/package.json` | Current package dependencies and scripts. |

### Task 3 dependency status

Current `frontend/user-app/package.json` already contains:

```json
{
  "dependencies": {
    "@hookform/resolvers": "^5.4.0",
    "react-hook-form": "^7.76.1",
    "zod": "^4.4.3"
  }
}
```

These are required for auth forms.

### Install command only if dependencies are missing

Run from `frontend` workspace root:

```bash
pnpm --filter user-app add react-hook-form zod @hookform/resolvers
```

Then verify:

```bash
pnpm --filter user-app typecheck
pnpm --filter user-app test
pnpm --filter user-app build
```

### Task-specific dependency issues

| Error | Cause | Fix |
|---|---|---|
| `Cannot find module 'react-hook-form'` | Dependency missing or install incomplete. | Run `cd frontend && pnpm install`. If package absent, run the add command above. |
| `Cannot find module 'zod'` | Validation package missing. | Run `cd frontend && pnpm --filter user-app add zod`. |
| `Cannot find module '@hookform/resolvers/zod'` | Resolver package missing. | Run `cd frontend && pnpm --filter user-app add @hookform/resolvers`. |
| Zod schema tests fail after dependency change | Zod version or schema behavior changed. | Re-run `pnpm install`, check lockfile, and run `pnpm --filter user-app test`. |

---

## 5. Database Setup

No direct database setup is required for Task 3 frontend.

### Direct frontend database status

| Database | Directly used by auth screens? | Required to render forms? | Required for real backend auth? | Status |
|---|---:|---:|---:|---|
| MySQL | No | No | Yes, if Auth Service stores users/tokens there | Reused backend dependency |
| Redis | No | No | Yes, if backend uses it for OTP/rate-limit/session checks | Reused backend dependency |
| MongoDB | No | No | Not for core auth in current docs | Reused platform dependency only |
| PostgreSQL | No | No | Not detected | Not used |
| SQLite | No | No | Not detected | Not used |

Simple Hinglish:

"Frontend form sirf API call bhejta hai. User table, password hash, OTP hash, refresh token hash, and rate-limit counters backend side par store hote hain. Isliye DB username/password frontend `.env` me kabhi nahi dalna."

### Backend database notes for auth

Auth/security docs mention:

- Passwords backend me Argon2id or bcrypt se hash hone chahiye.
- Refresh token hash backend storage me rahega.
- OTP challenge/token plain text store nahi hoga; only hash store hoga.
- OTP expiry around 5 minutes and max attempts backend enforce karega.

For MySQL, Redis, Docker commands, ports, connection strings, and credential placement, reuse:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Sections:

```text
3. Database Analysis
5. External Services Analysis
7. Docker and DevOps Setup
```

Credential placement:

- Backend credentials: backend service `.env`, Docker Compose env, or secret manager.
- Frontend public config: `frontend/user-app/.env.local`.
- Never put `MYSQL_DSN`, DB password, JWT secret, Redis password, SMTP password, or SMS provider secret in frontend env.

---

## 6. Redis / Queue / External Services

Task 3 does not add direct browser access to Redis, Kafka, RabbitMQ, NATS, MinIO, Elasticsearch, Firebase, SMTP, Twilio, Stripe, or Kubernetes.

### Services relevant to real auth testing

| Service | Required for static auth UI? | Required for real auth flow? | Purpose | Status |
|---|---:|---:|---|---|
| Vite dev server | Yes | Yes | Runs frontend app in browser. | Reused |
| REST API Gateway | No | Yes | Receives auth REST calls from browser. | Reused, Task 3 depends on it |
| Backend Auth Service | No | Yes | Handles signup, login, OTP, password reset, logout. | Backend dependency |
| Notification service / email or SMS provider | No | Yes for real OTP delivery | Sends OTP to email/phone. | Backend dependency, not frontend |
| Redis | No | Backend-dependent | Rate limits, OTP retry/cooldown, sessions if configured. | Reused |
| MySQL | No | Backend-dependent | User, password hash, refresh-token/OTP hash storage if configured. | Reused |
| gRPC-Web bridge | No | No for Task 3 REST auth screens | Later typed gRPC-Web features. | Reused, not Task 3-specific |
| Kafka/RabbitMQ | No | Optional backend events | OTP verified/user events may be async later. | Reused platform dependency |

### Ports and networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Vite User App dev server | `5173` | Browser frontend for auth screens. | Reused |
| Vite preview server | `4173` | Preview production build locally. | Reused |
| REST API Gateway | `8080` | Auth REST API target from `VITE_API_BASE_URL`. | Reused, required for real auth |
| gRPC-Web bridge | `8082` | Later gRPC-Web target. | Reused, not required for Task 3 |
| Auth Service | Backend-specific | Internal service behind gateway. | No direct browser port |
| MySQL | `3306` | Backend auth persistence if used. | Reused backend setup |
| Redis | `6379` | Backend OTP/session/rate-limit support if used. | Reused backend setup |

### API Gateway health check

If real auth calls are failing, first verify gateway:

```bash
curl http://localhost:8080/health/live
curl http://localhost:8080/health/ready
```

If these endpoints are not available in the local backend implementation, use the backend team's documented health endpoint.

---

## 7. Environment Variables

No new frontend environment variable was introduced by Task 3.

Reuse existing env documentation from:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Section:

```text
4. Environment Variables
```

### Existing variable needed for real auth

| Variable | New for Task 3? | Required for static UI? | Required for real auth? | Example | Purpose |
|---|---:|---:|---:|---|---|
| `VITE_API_BASE_URL` | No | No | Yes | `http://localhost:8080` | API Gateway base URL for auth REST endpoints. |

### Where to create env file

Use the existing frontend env location:

```text
frontend/user-app/.env.local
```

Create it from the example only if real API behavior is needed:

```bash
cd frontend
cp user-app/.env.example user-app/.env.local
```

For Task 3, check this value:

```env
VITE_API_BASE_URL=http://localhost:8080
```

### Important env rules

- Vite exposes only `VITE_` prefixed variables to browser code.
- All `VITE_` values are public after build.
- Do not add backend secrets to frontend env.
- Restart Vite after changing `.env.local`.
- Auth requests use `credentials: 'include'`, so backend CORS must allow the active frontend origin and credentials.

---

## 8. Docker Setup

No new Docker container, Dockerfile, volume, network, health check, or port mapping was introduced by Task 3.

### Detected Task 3 Docker status

| Item | Status |
|---|---|
| New frontend Dockerfile | Not detected. |
| New compose service for auth screens | Not required. |
| New Docker volume | None. |
| New Docker network | None. |
| New port mapping | None. |
| New health check | None. |

For backend database/cache/service containers, reuse:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Sections:

```text
3. Database Analysis
5. External Services Analysis
7. Docker and DevOps Setup
```

Optional central reference:

```text
TaskImplementation/Dependency/Docker.md
```

Beginner note:

- Auth UI render karne ke liye Docker needed nahi hai.
- Real signup/login/OTP flow ke liye backend stack running hona chahiye.
- Agar backend Docker setup available ho, API Gateway, Auth Service, MySQL, Redis, and notification dependencies ko start karo.

---

## 9. Local Development Setup

Follow this order for Task 3.

### Step 1: Read previous dependency documentation first

Read:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
```

Why:

- Node.js, pnpm, Vite, Tailwind setup already documented hai.
- Routing and app-shell setup already documented hai.
- Database, Docker, generic env, and common errors already documented hain.

### Step 2: Go to frontend workspace

```bash
cd frontend
```

### Step 3: Install dependencies

If fresh clone:

```bash
pnpm install
```

If only Task 3 form packages are missing:

```bash
pnpm --filter user-app add react-hook-form zod @hookform/resolvers
```

### Step 4: Configure env only for real auth API calls

For static UI and form validation:

```text
.env.local is optional because defaults exist.
```

For real API calls:

```bash
cp user-app/.env.example user-app/.env.local
```

Then confirm:

```env
VITE_API_BASE_URL=http://localhost:8080
```

### Step 5: Start backend services only for real auth

For only viewing screens:

```text
Backend not required.
```

For signup/login/OTP/password reset:

```text
Start API Gateway, Auth Service, Auth Service database, Redis/rate-limit dependency if enabled, and notification provider/dev OTP adapter if required by backend.
```

Use previous dependency docs and backend service docs for exact backend commands.

### Step 6: Run migrations if needed

Frontend Task 3 has no migration.

```text
No frontend migration required.
```

Backend Auth Service may need migrations for users, passwords, refresh tokens, OTP challenges, sessions, or audit tables. Run those according to backend migration docs before real auth testing.

### Step 7: Start frontend

```bash
pnpm --filter user-app dev
```

Or:

```bash
pnpm dev:user
```

Usually open:

```text
http://localhost:5173
```

### Step 8: Verify Task 3 routes

| Route | Expected |
|---|---|
| `/login` | Login form loads with email/phone and password. |
| `/signup` | Signup form loads with name, email, phone, password, confirm password, terms. |
| `/auth/otp` | OTP form loads; without challenge it shows missing challenge info. |
| `/forgot-password` | Forgot password form loads. |
| `/reset-password` | Reset form loads; without challenge it shows missing challenge info. |

---

## 10. Running the Project

### Quality checks

Run from `frontend`:

```bash
pnpm --filter user-app typecheck
pnpm --filter user-app lint
pnpm --filter user-app test
pnpm --filter user-app build
```

### Task 3 manual verification

| Check | Expected Result |
|---|---|
| Empty login submit | Inline validation errors show. |
| Invalid email/phone login | Schema error says valid email or phone required. |
| Signup weak password | Password requirement messages show. |
| Signup password mismatch | Confirm password error shows. |
| Signup without terms | Terms validation error shows. |
| OTP non-numeric or short code | 6 digit OTP error shows. |
| Forgot password invalid identifier | Email/phone validation error shows. |
| Reset password weak/mismatch | Password and confirm errors show. |
| Server returns error | Safe form-level error banner shows. |
| API Gateway down | Network error shows; app should not crash. |
| Successful login | Access token metadata persists, auth store updates, redirect uses safe path. |
| Successful OTP verify | Redirects to login with verified notice. |
| Successful password reset | Redirects to login with password reset notice. |

### Auth endpoints to verify through browser/network tab

| Action | Endpoint |
|---|---|
| Signup | `POST /api/v1/auth/signup` |
| Signup OTP send | `POST /api/v1/auth/otp/send` |
| Login | `POST /api/v1/auth/login` |
| OTP verify | `POST /api/v1/auth/otp/verify` |
| Forgot password | `POST /api/v1/auth/password/forgot` |
| Reset password | `POST /api/v1/auth/password/reset` |
| Logout from account menu | `POST /api/v1/auth/logout` |

---

## 11. Common Errors & Fixes

Generic Node, pnpm, Docker, database, CORS, and env troubleshooting is already documented in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Section:

```text
9. Common Errors and Fixes
```

### Task 3 specific errors

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `Cannot find module 'react-hook-form'` | Dependency missing or install incomplete. | Run `cd frontend && pnpm install`. If absent, add the package. | Install from workspace root and commit lockfile. |
| `Cannot find module '@hookform/resolvers/zod'` | Resolver package missing. | Run `pnpm --filter user-app add @hookform/resolvers`. | Keep auth dependencies together. |
| Auth form submits but no API request appears | Client validation failed before submit. | Check inline errors and browser console. | Keep schemas aligned with form fields. |
| `NETWORK_ERROR` during login/signup | API Gateway stopped or wrong `VITE_API_BASE_URL`. | Start gateway and check `curl http://localhost:8080/health/live`. | Verify env before real auth testing. |
| Browser CORS error | Gateway does not allow Vite origin or credentials. | Allow `http://localhost:5173` and `credentials` in backend CORS config. | Keep local/staging/prod origin allowlist updated. |
| `404` on auth endpoint | Gateway route missing or base URL wrong. | Compare endpoint with `api/master-api.json` and auth.api.ts. | Keep frontend API contract synced with gateway routes. |
| `401` on logout | Access token expired, auth cookie missing, or backend session invalid. | Login again or clear browser storage and retry. | Implement refresh flow in later auth/session task. |
| OTP challenge missing | `/auth/otp` or `/reset-password` opened without `challenge_id`. | Start from signup or forgot-password flow, or resend OTP when target exists. | Navigate with `challenge_id` from backend response. |
| OTP expired | Backend expiry passed, commonly around 5 minutes. | Click resend OTP and use new code. | Show expiry/cooldown UX and avoid stale tabs. |
| OTP resend fails | Missing target, backend cooldown, or rate limit. | Ensure URL has `target`, wait for cooldown, check backend logs. | Backend should return clear `Retry-After` or error message. |
| Password reset fails after resend | Form still has old challenge ID or backend rejected OTP. | Use latest OTP after resend; page updates hidden challenge ID. | Always update challenge state after resend. |
| Login succeeds but protected pages still redirect | Auth store not hydrated, token expired, or current-user API failed. | Refresh once, check session storage, verify `/api/v1/me` if used. | Keep auth store/query setup consistent. |
| `.env.local` changed but auth still calls old URL | Vite server was not restarted. | Stop and restart `pnpm --filter user-app dev`. | Restart Vite after env edits. |

---

## 12. Security & Best Practices

### Security audit for Task 3

| Area | Finding | Recommendation |
|---|---|---|
| Frontend secrets | No DB/JWT/payment secret detected in `.env.example`. | Keep all secrets backend-only. |
| Public env | `VITE_API_BASE_URL` is public browser config. | Use it only for URLs, never credentials. |
| Token handling | Current code stores access token metadata in `sessionStorage`. | Keep access tokens short-lived, clear on logout, and harden app against XSS. Prefer HttpOnly refresh cookies backend-side. |
| Cookies/credentials | HTTP client uses `credentials: 'include'`. | Gateway CORS must allow trusted origins only and credentials safely. |
| Password/OTP logging | Auth forms should not log passwords, OTPs, or tokens. | Avoid `console.log(formData)` and sanitize error reporting. |
| OTP storage | Frontend only handles challenge ID and OTP input. | Backend must hash OTP and enforce expiry/attempt limits. |
| Redirect handling | Login uses safe redirect path. | Keep rejecting external or protocol-relative redirects. |
| Password policy | Frontend enforces minimum length, uppercase, lowercase, and number. | Backend policy must match or be stricter. |
| Query params | `challenge_id` and `target` appear in URL. | Never put OTP, password, access token, or refresh token in query params. |
| Buyer role | Signup sends buyer role for this app. | Backend must still enforce allowed roles and ignore unauthorized role escalation. |

### Task-specific best practices

- Keep auth validation schemas in one file so pages do not duplicate rules.
- Keep frontend validation user-friendly, but never trust it as security control.
- Keep API endpoint strings centralized in `auth.api.ts`.
- Keep auth errors readable and generic enough to avoid account enumeration.
- Use `autoComplete="current-password"`, `autoComplete="new-password"`, and `autoComplete="one-time-code"` correctly.
- Disable submit buttons while requests are pending.
- Clear auth session on logout even if backend logout returns an error, but log/monitor backend failure separately.
- Use HTTPS in staging/production for auth routes.
- Configure backend rate limits for login and OTP routes.
- Add E2E tests for complete signup/login/reset flows once backend test environment exists.

---

## 13. Missing or Misconfigured Things

These are not blockers for generating this dependency file, but beginners should know them.

| Item | Status | Impact | Suggested Fix |
|---|---|---|---|
| Frontend Dockerfile | Not detected | Frontend container packaging is not one-command yet. | Add Dockerfile/Nginx config when deployment packaging starts. |
| Local full-stack compose | Not detected in inspected root/frontend files | Beginners cannot start gateway/auth/db/redis with one command from frontend docs. | Add or reference a local backend compose file. |
| Auth backend startup command | Not visible in this task file | Real auth testing may be confusing. | Add per-service backend run and migration docs. |
| Auth migrations | Not documented in frontend task | Backend user/token/OTP tables may be missing locally. | Backend Auth Service docs should list migration commands. |
| OTP delivery provider/dev mode | Not documented in frontend task | Signup/forgot password may create challenge but user may not receive OTP. | Backend should document email/SMS provider env or local fake OTP strategy. |
| Refresh-token flow | API contract exists, Task 3 does not implement direct refresh flow | Long sessions may expire and protected pages may redirect. | Implement in session/state management task. |
| Production CORS/cookie matrix | Not shown here | Auth cookies/tokens may fail across domains. | Document allowed origins, cookie domain, SameSite, Secure, and HTTPS rules. |
| Full auth E2E tests | Not detected here | Browser + backend auth regressions may need manual testing. | Add Playwright or equivalent E2E tests after backend test stack is stable. |

---

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `task1_Dependency.md` | `1. Project Tech Stack Analysis` | Base React, TypeScript, Vite, Tailwind, React Query, Zustand, and testing stack already explained. |
| `task1_Dependency.md` | `2. Node.js Dependency System` | Node.js, pnpm, workspace install, lockfile, and scripts already documented. |
| `task1_Dependency.md` | `3. Database Analysis` | Frontend has no direct DB; backend MySQL/MongoDB/Redis setup already covered. |
| `task1_Dependency.md` | `4. Environment Variables` | `.env.local`, Vite env loading, and existing `VITE_*` variables already documented. |
| `task1_Dependency.md` | `5. External Services Analysis` | API Gateway, gRPC-Web bridge, Redis, queues, and payment provider concepts already covered. |
| `task1_Dependency.md` | `6. Ports and Networking` | Existing ports, port conflicts, CORS, and networking notes already documented. |
| `task1_Dependency.md` | `7. Docker and DevOps Setup` | Docker recommendations and backend infrastructure setup already covered. |
| `task1_Dependency.md` | `9. Common Errors and Fixes` | Generic setup troubleshooting already documented. |
| `task1_Dependency.md` | `10. Security and Configuration Audit` | General frontend secret handling and config risks already covered. |
| `task2_Dependency.md` | `2. Tech Stack` | React Router and app-shell routing dependencies already explained. |
| `task2_Dependency.md` | `6. Redis / Queue / External Services` | App shell API side effects and reused gateway behavior already documented. |
| `task2_Dependency.md` | `11. Common Errors & Fixes` | Router/provider errors relevant to auth pages are already covered. |

---

## 15. Final Checklist

- [ ] Previous dependency documentation checked: `task1_Dependency.md` and `task2_Dependency.md`.
- [ ] No duplicate Node.js, pnpm, Vite, Tailwind, Docker, or database setup copied.
- [ ] Task 3 auth dependencies confirmed: `react-hook-form`, `zod`, `@hookform/resolvers`.
- [ ] `frontend/user-app/package.json` and `frontend/pnpm-lock.yaml` are in sync.
- [ ] No new frontend env variable added.
- [ ] `VITE_API_BASE_URL` points to the API Gateway for real auth testing.
- [ ] No database credentials added to frontend env.
- [ ] No Redis/Kafka/RabbitMQ credentials added to frontend env.
- [ ] No new Docker container or port introduced for Task 3.
- [ ] API Gateway running on `http://localhost:8080` when testing real auth.
- [ ] Backend Auth Service running behind the gateway for real auth.
- [ ] Backend Auth Service migrations completed if required.
- [ ] Backend Redis/rate-limit/OTP dependencies running if enabled.
- [ ] Email/SMS/dev OTP strategy available for OTP testing.
- [ ] Frontend dev server opens auth routes on `http://localhost:5173`.
- [ ] Login, signup, OTP, forgot password, and reset password screens render.
- [ ] Client-side validation errors display correctly.
- [ ] Server/network errors show safe form-level messages.
- [ ] Passwords, OTPs, and tokens are not logged.
- [ ] CORS allows the active frontend origin with credentials.
- [ ] `pnpm --filter user-app typecheck` passes.
- [ ] `pnpm --filter user-app lint` passes.
- [ ] `pnpm --filter user-app test` passes.
- [ ] `pnpm --filter user-app build` passes.
- [ ] No duplicate setup documentation added.
