# Project Dependency & Setup Guide

Input task file:

```text
TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}
```

Generated dependency file:

```text
TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}
```

This guide explains the dependency and setup impact of the App Shell task. It intentionally reuses the already-written setup documentation from `task1_Dependency.md` instead of repeating the same Node.js, pnpm, Vite, Tailwind, database, Docker, and environment setup.

Important beginner note:

- Pure App Shell ka kaam browser layout banana hai: header, navigation, search box, account menu, cart badge, mobile nav, and route outlet.
- Browser frontend direct database, Redis, Kafka, RabbitMQ, or migrations se connect nahi karta.
- Current codebase me later tasks ke integrations bhi present hain, jaise cart query, auth logout, autocomplete, React Query, and Zustand. Ye Task 2 ke core scope ka part nahi the, but local testing me visible ho sakte hain.

---

## 1. Project Overview

Task 2 App Shell frontend app ka reusable outer frame banata hai.

Simple Hinglish:

"App shell matlab woh common layout jo har page ke around dikhega. Jaise ecommerce app me top header, logo, search bar, account menu, cart icon, desktop nav, mobile nav, aur page content area."

### Task-specific files checked

| File | Purpose |
|---|---|
| `frontend/user-app/src/app-shell/app-shell.tsx` | Common shell wrapper with header, mobile nav, and `Outlet`. |
| `frontend/user-app/src/app-shell/header.tsx` | Sticky header composition. |
| `frontend/user-app/src/app-shell/nav-links.tsx` | Reusable desktop/mobile navigation links. |
| `frontend/user-app/src/app-shell/search-box.tsx` | Search input and route navigation. Current code also calls autocomplete after 2+ characters. |
| `frontend/user-app/src/app-shell/account-menu.tsx` | Account dropdown. Current code uses auth/UI stores and logout mutation. |
| `frontend/user-app/src/app-shell/cart-badge.tsx` | Cart link and item count badge. Current code can query cart when auth session exists. |
| `frontend/user-app/src/app-shell/mobile-nav.tsx` | Mobile browse menu. Current code stores open/close state in Zustand UI store. |
| `frontend/user-app/src/routes/index.tsx` | Browser route tree under `AppShell`. |
| `frontend/user-app/src/routes/route-paths.ts` | Central route constants. |
| `frontend/user-app/src/app.tsx` | `RouterProvider` mount. |
| `frontend/user-app/src/main.tsx` | React root with `AppProviders`. |

### What is new for this task

| Area | New / Reused | Notes |
|---|---|---|
| App shell components | New | Header, nav, search, account menu, cart badge, mobile nav. |
| React Router usage | Task-specific | Required for `RouterProvider`, `Outlet`, `Link`, `NavLink`, `useNavigate`, and `useSearchParams`. |
| lucide icons | Task-specific | Required for header icons like search, account, cart, menu, close, logout. |
| API Gateway | Reused / later-task effect | Pure Task 2 does not need API calls, but current search/cart/account code can call APIs. |
| Database setup | Reused / not direct | Frontend has no direct DB dependency. |
| Docker setup | Reused / unchanged | No new Dockerfile or compose service detected for this task. |
| Environment variables | Reused / unchanged | No new Task 2 env variables detected. |

---

## 2. Tech Stack

Most base technologies are already explained in `task1_Dependency.md`.

Refer:

```text
task1_Dependency.md
```

Sections:

```text
1. Project Tech Stack Analysis
2. Node.js Dependency System
```

### Task 2 specific technologies

| Technology | Required? | Why used in App Shell | Setup status |
|---|---:|---|---|
| `react-router-dom` | Required | Links, active nav state, browser routing, route outlet, and search URL navigation ke liye. | Present in `frontend/user-app/package.json`. |
| `lucide-react` | Required for current UI | Search, user, cart, menu, close, logout icons ke liye. | Present in `frontend/user-app/package.json`. |
| React Router `RouterProvider` | Required | App ko browser route tree se connect karta hai. | Used in `src/app.tsx`. |
| React Router `Outlet` | Required | Shell ke andar current page render karta hai. | Used in `src/app-shell/app-shell.tsx`. |
| React Router `Link` / `NavLink` | Required | Header links and active nav styling ke liye. | Used in app-shell components. |
| React Router `useNavigate` / `useSearchParams` | Required for current search box | Search query ko `/search?q=...` URL me convert karta hai. | Used in `search-box.tsx`. |
| Zustand | Current-code dependency | Mobile nav and account menu open/close state current code me `useUiStore` se aati hai. | Already explained in `task1_Dependency.md`. |
| TanStack React Query | Current-code dependency | Autocomplete and cart badge current code me query hooks use karte hain. | Already explained in `task1_Dependency.md`. |

### Beginner explanation

`react-router-dom` ek routing library hai. Isse React app me page change hota hai bina full browser reload ke. Task 2 me ye important hai kyunki App Shell parent layout hai aur pages `Outlet` ke andar render hote hain.

`lucide-react` ek icon library hai. Isse search, cart, user, menu jaise common icons clean and consistent milte hain. Custom SVG paste karne ki zarurat nahi padti.

Zustand and React Query pure Task 2 ke liye mandatory nahi the, but current implementation me later task changes ke through imported hain. Inka detailed setup repeat nahi kiya gaya hai.

---

## 3. Required Software

No new system software was introduced by this task.

Use the same required software from:

```text
task1_Dependency.md
```

Section:

```text
2. Node.js Dependency System
```

Minimum tools:

| Software | Required? | Notes |
|---|---:|---|
| Node.js `>=22.13.0` | Required | Vite, TypeScript, ESLint, and tests chalane ke liye. |
| pnpm `>=11.5.0` | Required | Frontend workspace dependencies install karne ke liye. |
| Git | Required | Repository clone/pull ke liye. |
| Docker | Optional for this task | App Shell UI run karne ke liye required nahi. Backend stack ke liye useful. |

Do not repeat full installation steps here. Follow the previous dependency guide first.

---

## 4. Dependency Management

This is a pnpm workspace based Node.js frontend.

Refer:

```text
task1_Dependency.md
```

Sections:

```text
2. Node.js Dependency System
Quick Command Summary
```

### Package files involved

| File | Why it matters |
|---|---|
| `frontend/package.json` | Workspace scripts and engine versions. |
| `frontend/pnpm-workspace.yaml` | Defines workspace packages. |
| `frontend/pnpm-lock.yaml` | Exact dependency versions lock karta hai. |
| `frontend/user-app/package.json` | User app dependencies and scripts. |

### Task 2 package check

Current `frontend/user-app/package.json` already contains:

```json
{
  "dependencies": {
    "lucide-react": "^1.17.0",
    "react-router-dom": "^7.16.0"
  }
}
```

If a fresh branch is missing these packages, install from the `frontend` workspace root:

```bash
cd frontend
pnpm --filter user-app add react-router-dom lucide-react
```

Then verify:

```bash
pnpm --filter user-app typecheck
pnpm --filter user-app lint
pnpm --filter user-app build
```

### Why dependency install can fail

| Error | Cause | Fix |
|---|---|---|
| `Cannot find module 'react-router-dom'` | Package missing or install incomplete. | Run `pnpm install` from `frontend`. If still missing, add the package command above. |
| `Cannot find module 'lucide-react'` | Icon package missing. | Run `pnpm --filter user-app add lucide-react`. |
| Workspace package not found | Command ran from wrong folder. | Run commands from `frontend`, not from `frontend/user-app` unless you know the workspace impact. |
| Lockfile conflict | `package.json` and lockfile out of sync. | Run `pnpm install` from `frontend` and commit lockfile changes if dependencies changed. |

---

## 5. Database Setup

No new database is required for Task 2.

### Direct frontend database status

| Database | Directly used by App Shell? | Required to run App Shell UI? | Status |
|---|---:|---:|---|
| MySQL | No | No | Reused backend-side dependency only. |
| MongoDB | No | No | Reused backend-side dependency only. |
| Redis | No direct browser use | No | Backend-side cache/session support only. |
| PostgreSQL | No | No | Not detected. |
| SQLite | No | No | Not detected. |
| Kafka/RabbitMQ | No direct browser use | No | Backend/event infrastructure only. |

Simple Hinglish:

"Frontend browser app database username/password directly use nahi karta. Database credentials hamesha backend services ke env me rahenge. User App sirf API Gateway ya gRPC-Web bridge ko call karega."

For database installation, Docker commands, ports, connection strings, and credential placement, reuse:

```text
task1_Dependency.md
```

Sections:

```text
3. Database Analysis
5. External Services Analysis
7. Docker and DevOps Setup
```

---

## 6. Redis / Queue / External Services

Pure Task 2 App Shell ke liye no new Redis, Kafka, RabbitMQ, NATS, MinIO, Elasticsearch, SMTP, Stripe, Twilio, Firebase, or Kubernetes setup was introduced.

### Current-code runtime services to be aware of

Because the live codebase already has later-task integrations, these services can affect local testing:

| Service | Required for pure shell layout? | When it becomes visible | Env / URL |
|---|---:|---|---|
| Vite dev server | Yes | Always, for local frontend. | Usually `http://localhost:5173`. |
| REST API Gateway | No for static shell, yes for live suggestions/cart/logout | Search autocomplete after 2+ chars, cart badge with auth session, logout click. | `VITE_API_BASE_URL`, default `http://localhost:8080`. |
| gRPC-Web bridge | No for Task 2 shell | Used by generated clients and later gRPC features. | `VITE_GRPC_WEB_BASE_URL`, default `http://localhost:8082`. |
| Backend auth/cart/search services | No for pure shell | Needed for real account/cart/search behavior. | Behind API Gateway. |

### Task-specific endpoint behavior

| UI action | Current behavior | Backend needed? |
|---|---|---:|
| Open home/header/nav | Shell renders in browser. | No. |
| Click nav links | React Router changes route. | No for client route change. |
| Type search query with 2+ characters | Current code may call `/api/v1/search/autocomplete`. | Yes for suggestions. |
| Submit search | Browser navigates to `/search?q=<query>`. Current search page may call product search APIs. | Yes for real results. |
| Open account menu | Uses local/store state. | No. |
| Click logout | Calls `/api/v1/auth/logout`. | Yes. |
| Show authenticated cart count | Calls `/api/v1/cart` if user/session exists. | Yes. |

### Ports and networking for this task

Task 2 ne koi new port introduce nahi kiya. Neeche table current App Shell testing ke perspective se hai:

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Vite User App dev server | `5173` | Browser me User App Frontend run karna | Reused default |
| REST API Gateway | `8080` | Autocomplete, cart count, logout, and later REST APIs | Reused / only needed for live API behavior |
| gRPC-Web bridge | `8082` | Later typed gRPC-Web browser calls | Reused / not needed for pure shell |
| Backend auth/search/cart services | Backend-specific | API Gateway ke behind actual business APIs | Reused backend setup |
| MySQL / MongoDB / Redis / queues | Backend-specific | Backend persistence, cache, events | Reused, no direct frontend access |

Beginner notes:

- Agar sirf header/nav/search layout dekhna hai, `5173` enough hai.
- Agar search suggestions, cart count, logout, or later pages real data call kar rahe hain, `8080` API Gateway reachable hona chahiye.
- Browser se cookies bhejne ke liye current HTTP client `credentials: 'include'` use karta hai. Backend CORS config me frontend origin, usually `http://localhost:5173`, allow hona chahiye.
- If Vite says `Port 5173 is already in use`, terminal me printed alternate URL open karo.
- Production static hosting me nested routes like `/account/orders` refresh karne par SPA fallback `index.html` par point hona chahiye.

Refer:

```text
task1_Dependency.md
```

Sections:

```text
5. External Services Analysis
6. Ports and Networking
```

---

## 7. Environment Variables

No new Task 2-specific environment variable was detected.

Use the existing frontend env setup from:

```text
task1_Dependency.md
```

Sections:

```text
4. Environment Variables
```

### Existing variables relevant to current App Shell testing

The current `.env.example` contains:

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_GRPC_WEB_BASE_URL=http://localhost:8082
VITE_GRPC_WEB_TIMEOUT_MS=5000
VITE_APP_ENV=local
VITE_PAYMENT_PROVIDERS=stripe,razorpay
```

### What App Shell actually needs

| Variable | New for Task 2? | Required for pure shell? | Required for current live integrations? | Purpose |
|---|---:|---:|---:|---|
| `VITE_API_BASE_URL` | No | No | Yes | REST API Gateway base URL for autocomplete, cart, auth logout, and later pages. |
| `VITE_GRPC_WEB_BASE_URL` | No | No | Only for gRPC-Web features | gRPC-Web bridge base URL. |
| `VITE_GRPC_WEB_TIMEOUT_MS` | No | No | Only for gRPC-Web features | Client request timeout. |
| `VITE_APP_ENV` | No | No | Recommended | Frontend environment label. |
| `VITE_PAYMENT_PROVIDERS` | No | No | Checkout later | Payment provider list for later checkout UI. |

### Where to create `.env`

Create the file here if you need real API behavior:

```text
frontend/user-app/.env
```

Beginner warning:

- Vite only exposes env variables starting with `VITE_` to browser code.
- Do not put database password, JWT secret, SMTP password, payment secret key, or private API key in frontend `.env`.
- Frontend env values are public after build. Treat them as configuration, not secrets.
- `.env` change karne ke baad Vite dev server restart karo, warna old values browser bundle me reh sakti hain.

---

## 8. Docker Setup

No new Docker container, Dockerfile, volume, network, or health check was introduced by Task 2.

### Detected Docker status

| Item | Status |
|---|---|
| Frontend Dockerfile for user app | Not detected for this task. |
| Root docker-compose file | Not detected in inspected files. |
| New Task 2 container | None. |
| New Task 2 volume | None. |
| New Task 2 Docker network | None. |
| New Task 2 port mapping | None. |

For backend databases and external services, reuse:

```text
task1_Dependency.md
```

Sections:

```text
7. Docker and DevOps Setup
```

Optional central reference:

```text
TaskImplementation/Dependency/Docker.md
```

Use Docker only when you need real backend stack behavior. For pure App Shell UI, Vite is enough.

---

## 9. Local Development Setup

Follow this order so beginner developers do not get stuck.

### Step 1: Clone repository and read previous setup first

Start with:

```text
task1_Dependency.md
```

Why:

- Node.js version setup already documented hai.
- pnpm workspace setup already documented hai.
- Vite/Tailwind/TypeScript setup already documented hai.
- env, Docker, database, and common errors already documented hain.

Clone command, Git basics, and full first-time onboarding steps previous dependency guide me already covered hain. Yahan duplicate nahi kiya gaya.

### Step 2: Go to frontend workspace

```bash
cd frontend
```

### Step 3: Install dependencies

If dependencies are not installed:

```bash
pnpm install
```

If only Task 2 packages are missing:

```bash
pnpm --filter user-app add react-router-dom lucide-react
```

### Step 4: Create env only if real APIs are needed

For static shell checks, `.env` is optional because defaults exist in `src/lib/env.ts`.

For API-backed testing:

```bash
cp user-app/.env.example user-app/.env
```

Then adjust:

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_GRPC_WEB_BASE_URL=http://localhost:8082
```

### Step 5: Start backend only if testing real search/cart/auth behavior

For pure shell layout:

```text
Backend not required.
```

For autocomplete/cart/logout:

```text
Start API Gateway and related backend services according to previous dependency docs.
```

### Step 6: Run migrations

Frontend App Shell ke liye koi database migration nahi hai.

```text
No frontend migration required.
```

Backend APIs test karne hain to backend service migrations previous dependency docs ke according run karo.

### Step 7: Start frontend

```bash
pnpm --filter user-app dev
```

Or from `frontend` root shortcut:

```bash
pnpm dev:user
```

### Step 8: Open browser

Usually:

```text
http://localhost:5173
```

If port `5173` is busy, Vite may print another port. Use the URL shown in terminal.

---

## 10. Running the Project

### Quality verification

Run from `frontend`:

```bash
pnpm --filter user-app typecheck
pnpm --filter user-app lint
pnpm --filter user-app build
```

Optional tests:

```bash
pnpm --filter user-app test
```

### Task 2 manual verification

| Check | Expected result |
|---|---|
| App loads | No blank page, header visible. |
| Header desktop | Logo, nav, search, account, cart aligned. |
| Header mobile | Search visible and Browse menu can open/close. |
| Navigation links | Clicking links changes route without full reload. |
| Active link | Current nav item gets active styling. |
| Search submit | Query navigates to `/search?q=<encoded-query>`. |
| Empty search | Empty/space-only submit should not navigate. |
| Account menu | Opens and closes without layout shift. |
| Cart badge | Shows count when available; stays stable at zero. |
| Keyboard use | Tab focus visible on links/buttons/search. |

### Current-code extra verification

Because current shell has later-task API hooks:

| Check | Expected result |
|---|---|
| Search suggestions with backend down | UI should not crash. Suggestions may not appear. |
| Cart badge without login | Should remain zero or hidden. |
| Cart badge with auth session | Calls API Gateway and shows item count if API works. |
| Logout click | Calls API Gateway logout endpoint and clears local session. |

---

## 11. Common Errors & Fixes

Generic Node, pnpm, Docker, database, and env errors are already documented in:

```text
task1_Dependency.md
```

Section:

```text
9. Common Errors and Fixes
```

### Task 2 specific errors

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `Cannot find module 'react-router-dom'` | Dependency missing or install incomplete. | Run `cd frontend && pnpm install`. If package is absent, run `pnpm --filter user-app add react-router-dom`. | Keep `package.json` and lockfile committed together. |
| `Cannot find module 'lucide-react'` | Icon dependency missing. | Run `cd frontend && pnpm --filter user-app add lucide-react`. | Do not manually delete lockfile/node_modules without reinstalling. |
| `useNavigate() may be used only in the context of a Router` | SearchBox rendered outside `RouterProvider`. | Render app through `src/app.tsx` and `RouterProvider`, not isolated without router test wrapper. | For tests, wrap components in a memory router. |
| `useSearchParams() may be used only in the context of a Router` | Current SearchBox needs router context. | Use `RouterProvider` or router-aware test wrapper. | Keep app-shell components under route tree. |
| `No QueryClient set, use QueryClientProvider to set one` | Current SearchBox/CartBadge/AccountMenu can use React Query hooks from later tasks. | Render app through `src/main.tsx` with `AppProviders`, or wrap tests in `QueryClientProvider`. | Use existing `AppProviders` for app boot and tests. |
| Search suggestions not showing | API Gateway not running, wrong `VITE_API_BASE_URL`, CORS issue, or query shorter than 2 chars. | Start API Gateway, check `.env`, type 2+ characters, inspect browser network tab. | Keep `.env` aligned with backend port. |
| Browser blocks API request with CORS error | Backend did not allow Vite origin or credentials. | Allow `http://localhost:5173` and credentials on API Gateway/backend CORS config. | Keep frontend origin list synced across local/staging/prod. |
| Cart count stays zero | No logged-in user/session, API Gateway down, or cart endpoint failed. | Login first and verify `/api/v1/cart` response. | Do not treat header badge as source of truth. |
| Logout fails | `/api/v1/auth/logout` unavailable or auth cookie/token invalid. | Start backend auth service/API Gateway and clear stale session if needed. | Handle failed logout gracefully in UI. |
| 404 after browser refresh on nested route | Static host not configured for SPA fallback. | Configure server to serve `index.html` for app routes. | Add SPA fallback in production Nginx/static hosting config. |
| `.env` changed but app still calls old URL | Vite dev server was not restarted. | Stop and restart `pnpm --filter user-app dev`. | Restart Vite after env changes. |

---

## 12. Security & Best Practices

### Security audit for Task 2

| Area | Finding | Recommendation |
|---|---|---|
| Frontend secrets | No Task 2 secret required. | Never put DB passwords, JWT secrets, SMTP credentials, or payment secret keys in Vite env. |
| Search query | Query is encoded with `URLSearchParams`. | Keep encoding URL params. Never inject search text as raw HTML. |
| Account menu | Current code shows `user?.name` in a truncated label. | Avoid rendering sensitive PII in header. Prefer first name/display name only. |
| Cart count | Current badge can use server cart query. | Treat count as display-only. Checkout/payment must use server-side source of truth. |
| Auth logout | Current logout calls backend and clears local session on settle. | Backend should invalidate cookies/tokens securely. |
| API base URL | Public frontend config. | Use HTTPS in staging/production and configure CORS/cookies safely. |
| Cookie-based requests | HTTP/gRPC clients include credentials. | Backend cookies should use secure `SameSite`, domain, and HTTPS settings per environment. |
| Accessibility | App shell includes labels, focus classes, and ARIA state. | Keep keyboard and screen-reader checks in reviews. |

### Task-specific best practices

- Keep route paths centralized in `route-paths.ts` to avoid duplicate string paths.
- Keep App Shell business-light. Header should not own product/cart/auth business rules deeply.
- Use `Link` and `NavLink` instead of raw `<a>` for internal routes.
- Use `URLSearchParams` for query strings.
- Use accessible labels for icon-only actions.
- Keep mobile nav open/close state simple and predictable.
- Avoid adding heavy dropdown/menu libraries unless account menu behavior becomes complex.
- Keep frontend env public-safe. Anything secret belongs in backend config.

---

## 13. Missing or Misconfigured Things

These are not blockers for the Task 2 documentation file, but beginners should know them.

| Item | Status | Impact | Suggested fix |
|---|---|---|---|
| Task 2 original guide says no API calls | Current code has later autocomplete/cart/logout hooks | Beginner may see network errors while testing shell. | Treat those as later-task integrations and start backend only when testing real data. |
| Frontend Dockerfile | Not detected | Beginner cannot containerize user app directly from repo file. | Add a frontend Dockerfile when deployment packaging starts. |
| Root compose file | Not detected | Full local stack startup is not one command yet. | Add local compose file for API Gateway plus backend services. |
| Explicit Vite port config | Not set in `vite.config.ts` | Vite defaults to `5173`, but may choose another port if busy. | Add `server.port` only if team needs fixed port. |
| CORS origin list | Not visible in frontend config | API calls can fail even when backend is running. | Backend/API Gateway must allow the active Vite origin and credentials. |
| App shell unit tests | No dedicated app-shell test files detected in inspected files | Router/search/header regressions may rely on manual checks. | Add focused tests with router and query wrappers. |
| Production SPA fallback | Not shown in repo config | Refreshing nested routes can 404 on static hosting. | Configure Nginx/CDN/static host fallback to `index.html`. |
| Backend health checks | Not part of frontend task | Search/cart/auth issues can be confusing. | Backend/API Gateway should expose health endpoints and logs. |

---

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `task1_Dependency.md` | `1. Project Tech Stack Analysis` | Base React, TypeScript, Vite, Tailwind, React Query, Zustand, gRPC-Web stack already explained. |
| `task1_Dependency.md` | `2. Node.js Dependency System` | Node.js, pnpm, workspace install, lockfile, scripts, and common package issues already documented. |
| `task1_Dependency.md` | `3. Database Analysis` | Frontend has no direct DB connection; backend database setup already documented. |
| `task1_Dependency.md` | `4. Environment Variables` | `.env` location, Vite env loading, and existing frontend env variables already documented. |
| `task1_Dependency.md` | `5. External Services Analysis` | API Gateway, gRPC-Web bridge, payment providers, search engine, queues, and Docker-related services already explained. |
| `task1_Dependency.md` | `6. Ports and Networking` | Existing ports and networking concepts already documented. |
| `task1_Dependency.md` | `7. Docker and DevOps Setup` | Docker recommendations and backend service containers already covered. |
| `task1_Dependency.md` | `8. Complete Project Run Instructions` | Clone, install, env, backend startup, and frontend startup flow already covered. |
| `task1_Dependency.md` | `9. Common Errors and Fixes` | Generic Node, pnpm, Docker, database, and env troubleshooting already covered. |
| `task1_Dependency.md` | `10. Security and Configuration Audit` | General frontend secret handling and config risks already covered. |
| `task1_Dependency.md` | `11. Best Practices` | General dependency, env, API, Docker, and security best practices already covered. |

---

## 15. Final Checklist

- [ ] Previous dependency documentation checked: `task1_Dependency.md`.
- [ ] No duplicate Node.js/pnpm/Vite/Tailwind setup copied.
- [ ] Task 2 specific packages checked: `react-router-dom`, `lucide-react`.
- [ ] No new direct database dependency introduced.
- [ ] No new Redis/Kafka/RabbitMQ dependency introduced.
- [ ] No new Docker container, volume, network, or port introduced.
- [ ] Ports checked: Vite `5173`, API Gateway `8080`, gRPC-Web `8082`.
- [ ] No frontend database migration required.
- [ ] Existing frontend env variables understood.
- [ ] `.env` created only if real API behavior is being tested.
- [ ] Frontend dependencies installed from `frontend` workspace root.
- [ ] `pnpm --filter user-app typecheck` passes.
- [ ] `pnpm --filter user-app lint` passes.
- [ ] `pnpm --filter user-app build` passes.
- [ ] App Shell manually checked on desktop and mobile.
- [ ] Search route navigation verified.
- [ ] Account menu and mobile menu keyboard/focus behavior checked.
- [ ] Current-code API effects understood: autocomplete, cart count, logout.
- [ ] No frontend secrets added.
- [ ] SPA fallback considered for production hosting.
