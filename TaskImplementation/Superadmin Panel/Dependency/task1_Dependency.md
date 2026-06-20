# Superadmin Panel Task 1 - Dependency & Setup Guide

## 1. Project Overview

This handbook explains how to set up, run, test, and troubleshoot the admin shell described in `task1.md` and implemented in `frontend/superadmin-panel/`.

**Simple Hinglish:** Yeh project ek React frontend hai. Iska Task 1 secure admin layout, login entry, protected routes, sidebar/topbar, aur role-based menu ka foundation deta hai. Frontend browser me chalta hai; database se directly connect nahi karta.

### Scope and current status

| Area | Task 1 requirement | Current repository status |
|---|---|---|
| Frontend source | Required | Present in `frontend/superadmin-panel/` |
| Node dependencies | Required | Declared in `package.json` and locked in `frontend/pnpm-lock.yaml` |
| API Gateway | Required for real login/API data; not required for unit tests/build | URL configured, but runnable gateway source is not present |
| Auth/RBAC backend | Required for real login and real authorization | Contract documented, but complete local runtime is not present |
| Database | No direct Task 1 dependency | Frontend does not connect to a database |
| Redis/queue | No direct Task 1 dependency | Not imported or called by this frontend |
| Migrations | No direct Task 1 dependency | No frontend migration step |
| Docker | Optional | No Dockerfile or Compose file exists for this app |

> **Important:** `pnpm test`, `pnpm typecheck`, and `pnpm build` can work without MySQL, Redis, Kafka, or backend services. A real browser login needs a reachable API Gateway and Auth Service.

### Important paths

| Path | Purpose |
|---|---|
| `frontend/package.json` | Workspace-level commands |
| `frontend/pnpm-workspace.yaml` | Lists frontend workspace packages |
| `frontend/pnpm-lock.yaml` | Exact resolved dependency versions and integrity hashes |
| `frontend/superadmin-panel/package.json` | App dependencies and scripts |
| `frontend/superadmin-panel/.env.example` | Safe environment template |
| `frontend/superadmin-panel/vite.config.ts` | Vite, React, Tailwind, and Vitest configuration |
| `frontend/superadmin-panel/src/lib/http.ts` | API URL, Bearer token, timeout, and request ID handling |
| `frontend/superadmin-panel/src/stores/auth-store.ts` | Browser session storage |

## 2. Tech Stack

### Runtime and application libraries

| Technology | What it is | Project me kyon use hota hai | Required? |
|---|---|---|---|
| Node.js | JavaScript tooling runtime | Vite, TypeScript, tests, aur build commands Node par run hote hain | Yes, development/build time |
| pnpm | Fast package manager with workspace support | Monorepo ke frontend apps ko ek lockfile se install aur filter karta hai | Yes |
| React 19 | UI component library | Admin shell, pages, sidebar, topbar, forms render karta hai | Yes |
| React DOM | React browser renderer | React components ko HTML DOM me mount karta hai | Yes |
| TypeScript | JavaScript with static types | Roles, API objects, routes, aur component props me mistakes jaldi pakadta hai | Yes for build |
| Vite | Frontend dev server and bundler | Fast local server and production bundle banata hai | Yes |
| React Router DOM | Client-side router | `/login`, `/admin`, nested routes, redirects, and route guards manage karta hai | Yes |
| Zustand | Small state store | Auth user, token, hydration, and logout state manage karta hai | Yes |
| TanStack React Query | Server-state library | API data cache, retry, loading, and invalidation manage karta hai | Yes |
| Tailwind CSS | Utility-first CSS system | Admin UI styling ke liye use hota hai | Yes for current styles/build |
| Lucide React | Icon components | Sidebar/topbar icons provide karta hai | Yes |

### Development and test libraries

| Technology | Purpose | Required? |
|---|---|---|
| Vitest | Unit/component test runner | Required for `pnpm test` |
| jsdom | Browser-like DOM inside Node | Required for React tests |
| Testing Library | UI behavior test helpers | Required for component tests |
| Vite React plugin | React Fast Refresh and JSX transform | Required by Vite config |
| Tailwind Vite plugin | Tailwind processing in Vite | Required by Vite config |
| `@types/*` packages | TypeScript definitions | Required for typecheck/build |

### What is not part of the direct stack

- Go modules are not used by this frontend app.
- npm `package-lock.json` is not used; pnpm uses `pnpm-lock.yaml`.
- MySQL, MongoDB, Redis, Kafka, RabbitMQ, NATS, Elasticsearch, MinIO, SMTP, Stripe, and cloud credentials are not direct Task 1 dependencies.
- Frontend RBAC only controls user experience. Backend authorization is still mandatory for every protected API.

## 3. Required Software

### Minimum/recommended versions

| Software | Version | Why |
|---|---|---|
| Git | Recent stable version | Clone and update repository |
| Node.js | `22.12.0` or newer Node 22 | Current Vite dependency requires Node `^20.19.0`, `^22.12.0`, or `>=24`; Node 22 is the conservative choice |
| pnpm | pnpm 10 recommended | Lockfile format is `9.0`; use one team-approved major version consistently |
| Browser | Recent Chrome, Edge, or Firefox | Run and debug the admin UI |
| Docker Desktop/Engine | Recent stable version | Optional only; no project container config exists |

Check installed tools:

```bash
git --version
node --version
pnpm --version
```

### Windows installation

PowerShell me:

```powershell
winget install --id Git.Git -e
winget install --id OpenJS.NodeJS.LTS -e
npm install --global pnpm@10
```

Terminal restart karke versions verify karein. WSL use kar rahe hain to Node/pnpm WSL ke andar bhi install karein; Windows aur WSL installations alag hote hain.

### Linux installation

`nvm` se a compatible Node 22 release install karna practical hai:

```bash
nvm install 22
nvm use 22
npm install --global pnpm@10
```

`nvm` already installed nahi hai to its official installer use karein, or distribution ke trusted Node.js package source se Node install karein. Distribution repository ka old Node version Vite ke saath fail ho sakta hai.

### macOS installation

Homebrew example:

```bash
brew install git node@22
brew link --overwrite --force node@22
npm install --global pnpm@10
```

Alternative: `nvm install 22` use kar sakte hain.

### Corepack alternative

Node installation me Corepack available ho to:

```bash
corepack enable
corepack prepare pnpm@10 --activate
pnpm --version
```

> **Current gap:** Repository `package.json` me `packageManager` field nahi hai, isliye Corepack automatically exact pnpm version select nahi kar sakta.

## 4. Dependency Management

### Files ka role

#### `package.json`

`package.json` project ka dependency manifest hai. Isme package names, accepted version ranges, and commands such as `dev`, `test`, `typecheck`, and `build` defined hain.

#### `pnpm-lock.yaml`

Lockfile exact resolved versions and integrity hashes store karta hai. Same lockfile use karne se local machine aur CI me reproducible install milta hai.

#### `node_modules/`

Installed packages yahan materialize/link hote hain. Is folder ko Git me commit nahi karna hai. Delete hone par `pnpm install` ise recreate kar sakta hai.

#### pnpm workspace

`frontend/pnpm-workspace.yaml` multiple frontend packages ko one workspace banata hai. Isliye install command `frontend/` directory se run karna recommended hai.

### Install dependencies

Repository root se:

```bash
cd frontend
pnpm install --frozen-lockfile
```

Local development me lockfile intentionally update karna ho:

```bash
cd frontend
pnpm install
```

App-only command run karne ke two supported styles:

```bash
cd frontend
pnpm --filter superadmin-panel test
pnpm --filter superadmin-panel typecheck
pnpm --filter superadmin-panel build
```

or:

```bash
cd frontend/superadmin-panel
pnpm test
pnpm typecheck
pnpm build
```

### Add/remove dependencies

Workspace root se correct package target karein:

```bash
cd frontend
pnpm --filter superadmin-panel add package-name
pnpm --filter superadmin-panel add --save-dev package-name
pnpm --filter superadmin-panel remove package-name
```

Install ke baad `package.json` and `pnpm-lock.yaml` changes review karein. npm/yarn install commands mix na karein, warna conflicting lockfiles and dependency trees aa sakte hain.

### Main scripts

| Command from `frontend/` | Result |
|---|---|
| `pnpm run superadmin:dev` | Vite dev server starts |
| `pnpm run superadmin:test` | Test suite once run hoti hai |
| `pnpm run superadmin:typecheck` | TypeScript errors check hote hain |
| `pnpm run superadmin:build` | Typecheck plus production bundle in `dist/` |

### Dependency failure checklist

1. `node --version` check karein; unsupported Node common cause hai.
2. Command `frontend/` or app directory se run karein.
3. `pnpm-lock.yaml` present hona chahiye.
4. Proxy/VPN/corporate certificate npm registry access block to nahi kar raha, verify karein.
5. Corrupt install suspected ho to `node_modules` remove karke `pnpm install --frozen-lockfile` rerun karein.
6. pnpm store inspect/prune commands use kar sakte hain:

```bash
pnpm store path
pnpm store prune
pnpm install --frozen-lockfile
```

> `pnpm store prune` shared unused cache clean karta hai; normal setup me routinely required nahi hai.

## 5. Database Setup

### Direct database requirement: none

Task 1 frontend MySQL, PostgreSQL, MongoDB, SQLite, or Redis se directly connect nahi karta. Browser application ko database credentials dena insecure hota hai. Correct flow:

```text
Browser -> API Gateway -> backend service -> database
```

Therefore:

- No database server is needed for tests, typecheck, build, or shell rendering.
- No database URL belongs in `frontend/superadmin-panel/.env`.
- No migration command exists for this frontend.
- Never add a MySQL password, Mongo URI, or Redis password to a `VITE_*` variable; Vite variables browser bundle me public ho jate hain.

### Indirect full-stack databases

Other repository documents describe backend databases, but they are outside Task 1 setup. Their backend source/migrations are not complete in the currently inspected service folders.

| Backend dependency | Possible use | Task 1 install? |
|---|---|---|
| MySQL | Admin users/settings/audit data in backend design | No |
| MongoDB | Session analytics data through session service | No |
| Redis | Gateway rate limiting and possible live metrics/cache | No |

Do not run `database/draw.sql` just to start this frontend. Full-stack owners should provide versioned migrations and runnable services before database bootstrap is treated as supported.

## 6. Redis, Queue, and External Services

### API Gateway

API Gateway ek backend entry point hai. Frontend ka `VITE_API_BASE_URL` default `http://localhost:8080/api/v1` hai, and login request becomes:

```text
POST http://localhost:8080/api/v1/auth/login
```

| Question | Answer |
|---|---|
| Required for tests/build? | No |
| Required for real login and API pages? | Yes |
| Expected local address | `http://localhost:8080/api/v1` |
| Current runnable source present? | No; only `backend/services/api-gateway/.env` was found |
| Health endpoint proven by this app? | No |

Basic reachability check when a gateway implementation is supplied:

```bash
curl -i http://localhost:8080/api/v1/auth/login
```

`GET` may return `404` or `405`; that still proves a server answered. Use the backend team's documented health endpoint for a real readiness check. Do not invent `/health` as a guaranteed current route.

### Auth Service and RBAC

Real login must return an access token and a user with at least one accepted admin role. The frontend accepts roles such as `superadmin`, `operations_admin`, `finance_admin`, `catalog_admin`, and `readonly_admin` according to its RBAC helpers.

Auth is mandatory for real use. Frontend route/menu checks are not security enforcement; Gateway and backend services must validate JWT signature, expiry, audience, issuer, and permissions.

### Browser session storage

Zustand stores the active session under `sessionStorage` key `superadmin.session`. It is browser-tab scoped and is not an external service.

Useful debug commands in browser DevTools Console:

```js
sessionStorage.getItem("superadmin.session")
sessionStorage.removeItem("superadmin.session")
```

Do not paste production tokens into tickets, chat, screenshots, or documentation.

### Redis and message queues

- Redis is not called by frontend code.
- Kafka, RabbitMQ, and NATS are not used by Task 1 code.
- No local installation, container, credentials, or health check is required for these services to run the frontend.
- A future complete Gateway may require Redis, but that setup belongs to the Gateway service documentation.

## 7. Environment Variables

### Create the local file

From repository root:

```bash
cd frontend/superadmin-panel
cp .env.example .env
```

Windows PowerShell:

```powershell
Copy-Item .env.example .env
```

Safe local example:

```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_APP_NAME=Superadmin Panel
VITE_ADMIN_SESSION_WARNING_MINUTES=2
```

The root `.gitignore` excludes `.env` and allows `.env.example`, so real local values should remain uncommitted.

### Variable reference

| Variable | Required? | Example | Purpose | Current usage/security note |
|---|---|---|---|---|
| `VITE_API_BASE_URL` | Yes for API calls | `http://localhost:8080/api/v1` | Prefix for login and admin API requests | Used by `src/lib/http.ts`; not a secret |
| `VITE_APP_NAME` | Optional | `Superadmin Panel` | Intended display name | Present in template but currently not read by source |
| `VITE_ADMIN_SESSION_WARNING_MINUTES` | Optional | `2` | Intended warning lead time | Present in template but currently not read by source |

### How Vite loads variables

- Vite reads `.env`, `.env.local`, and mode-specific files when the dev server/build starts.
- Only variables prefixed with `VITE_` are exposed to client code through `import.meta.env`.
- Restart `pnpm dev` after changing `.env`.
- Production values are embedded at build time, not fetched dynamically at container start.
- Any `VITE_*` value is visible to users in downloaded JavaScript. Never put passwords, private keys, JWT signing secrets, database DSNs, or provider secret keys there.

### Environment mistakes

| Symptom | Likely cause | Fix |
|---|---|---|
| `VITE_API_BASE_URL is not configured` | `.env` missing, wrong name, or server not restarted | Copy template, check spelling, restart Vite |
| Request URL contains duplicated `/api/v1` | Base URL and API path both include it | Keep base exactly at gateway API root |
| Browser calls wrong server after `.env` change | Old build/dev process | Stop and restart server; rebuild production assets |
| Secret visible in browser bundle | Secret was put under `VITE_*` | Rotate secret and move it to backend-only configuration |

## 8. Docker Setup

### Current repository status

There is no app Dockerfile and no Compose service for this frontend. Docker is therefore optional and `docker compose up` is not currently a valid project setup command.

**Beginner recommendation:** Local Node + pnpm setup use karein because it matches repository scripts and gives simplest debugging.

### Optional one-off development container

Docker already installed ho and local Node avoid karna ho, repository root se this temporary container can run the app:

```bash
docker run --rm -it \
  -p 5173:5173 \
  -v "$PWD:/workspace" \
  -v superadmin_node_modules:/workspace/frontend/node_modules \
  -w /workspace/frontend \
  node:22-bookworm \
  sh -lc "corepack enable && corepack prepare pnpm@10 --activate && pnpm install --frozen-lockfile && pnpm run superadmin:dev"
```

This named volume only caches installed packages; it is not application data. On Windows PowerShell, replace `$PWD` with `${PWD}` if required by Docker Desktop shell behavior.

Container check:

```bash
docker ps
docker logs <container-id>
```

Stop it with `Ctrl+C` because `--rm` removes the stopped container.

### Networking note

`localhost` inside a container means that container itself. If the frontend dev container must call a Gateway running on the host, use a browser-visible URL. Since API calls execute in the host browser, `http://localhost:8080/api/v1` normally remains correct for this Vite setup. For server-side containers, Docker service names or `host.docker.internal` may be needed.

### Future production container requirements

Before production containerization, the project should add and review:

- Multi-stage Dockerfile: Node build stage plus non-root static web server stage.
- SPA fallback to `index.html` for `/admin/*` routes.
- Runtime/environment strategy because Vite values are build-time values.
- Health check for the static server.
- Read-only filesystem, minimal image, and security headers.
- Explicit port, restart policy, resource limits, and image version.

No persistent database volume is needed for this frontend container.

## 9. Local Development Setup

### Step 1: Clone repository

```bash
git clone <repository-url>
cd Ecommerce
```

If repository already cloned hai:

```bash
git status
git pull --ff-only
```

Do not discard existing local changes before pulling; review `git status` first.

### Step 2: Verify tools

```bash
node --version
pnpm --version
```

Use Node 22.12+ in the Node 22 line.

### Step 3: Install workspace dependencies

```bash
cd frontend
pnpm install --frozen-lockfile
```

### Step 4: Create frontend environment file

```bash
cd superadmin-panel
cp .env.example .env
```

Confirm `VITE_API_BASE_URL` matches the environment you intend to use.

### Step 5: Database and migration step

Skip it for Task 1. Frontend ke liye database/migration nahi hai.

### Step 6: Check code before starting

```bash
cd ..
pnpm run superadmin:typecheck
pnpm run superadmin:test
pnpm run superadmin:build
```

### Step 7: Start frontend

```bash
pnpm run superadmin:dev
```

Open:

```text
http://localhost:5173/login
http://localhost:5173/admin
```

Without a stored admin session, `/admin` should redirect to `/login`.

### Step 8: Full-stack verification

This step is only possible after runnable Gateway/Auth services and a valid admin account are supplied.

1. Start backend dependencies using their own service documentation.
2. Ensure Gateway is reachable at the configured base URL.
3. Ensure Gateway CORS permits the exact frontend origin, normally `http://localhost:5173`.
4. Login with an admin-role account.
5. In Browser DevTools Network tab, verify `POST /auth/login` and later API requests.
6. Confirm requests contain `Authorization: Bearer ...` after login.

> **Current limitation:** The inspected Gateway and Superadmin backend folders contain `.env` files but no runnable service code. Full-stack startup commands cannot honestly be provided from the current repository state.

## 10. Running the Project

### Daily development commands

Terminal 1:

```bash
cd frontend
pnpm run superadmin:dev
```

Terminal 2:

```bash
cd frontend
pnpm run superadmin:test
```

Quality gate before commit:

```bash
cd frontend
pnpm run superadmin:typecheck
pnpm run superadmin:test
pnpm run superadmin:build
```

### Production build preview

Build:

```bash
cd frontend
pnpm run superadmin:build
```

Preview generated bundle:

```bash
pnpm --filter superadmin-panel exec vite preview --host 0.0.0.0
```

Vite preview commonly uses port `4173`; terminal output is the source of truth. Preview is for validation, not a production web server.

### Ports and networking

| Component | Default/expected port | Purpose | Required now? |
|---|---:|---|---|
| Vite development server | `5173` | Local frontend UI | Yes for browser development |
| Vite preview | `4173` commonly | Local production-bundle preview | Optional |
| API Gateway | `8080` | Login and REST API base | Only for real API use |
| Superadmin backend HTTP | `8088` in backend env | Backend service design | Not directly called by frontend |
| Superadmin gRPC | `50062` in gateway env | Gateway-to-service design | Not directly called by frontend |
| Redis | `6379` in gateway env | Gateway rate limiting design | Not a direct frontend dependency |

Change Vite port temporarily:

```bash
pnpm --filter superadmin-panel dev -- --port 5174
```

Find port owner:

```bash
# Linux/macOS
lsof -i :5173

# Windows PowerShell / Command Prompt
netstat -ano | findstr :5173
```

Because the dev script binds `0.0.0.0`, devices on the same network may reach it if firewall allows. Do not expose a development server or test admin token to an untrusted network.

## 11. Common Errors & Fixes

| Error/symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `pnpm: command not found` | pnpm not installed/activated | Install pnpm 10 or enable Corepack, then reopen shell | Verify versions before install |
| `Unsupported engine` | Node version too old/incompatible | Switch to Node 22.12+ | Add/use a team version manager config |
| `ERR_PNPM_OUTDATED_LOCKFILE` | Manifest changed without lockfile update | On intended dependency change run `pnpm install`; otherwise restore correct manifest/lockfile pair | Commit both files together |
| Package download fails | DNS, proxy, registry, VPN, or TLS certificate issue | Check `pnpm config get registry`, proxy settings, and network access | Document corporate registry/CA setup |
| Module cannot be found | Incomplete install or wrong working directory | Run `pnpm install --frozen-lockfile` from `frontend/` | Use workspace commands consistently |
| Port already in use | Another process owns `5173` | Stop it or start Vite with `--port 5174` | Reserve/document local ports |
| `VITE_API_BASE_URL is not configured` | `.env` absent/misspelled | Copy `.env.example` to `.env` and restart Vite | Keep template current |
| `Failed to fetch` / connection refused | Gateway not running or wrong URL | Check Gateway address with `curl`; fix `.env` | Run backend reachability check first |
| CORS error | Gateway does not allow frontend origin | Backend must allow exact scheme/host/port and required headers | Maintain environment-specific origin allowlist |
| `401 Unauthorized` | Missing, expired, invalid, or wrongly scoped JWT | Clear session and login again; inspect backend auth logs/request ID | Short sessions plus correct refresh/logout design |
| `403 Forbidden` | Account lacks required admin role/permission | Use correctly provisioned account; fix backend RBAC assignment | Test role matrix at API layer |
| Redirect loop to `/login` | Invalid session shape, no admin role, or storage cleared | Clear `superadmin.session`, login again, inspect login response | Contract-test auth response |
| Blank page after static deployment | Web server lacks SPA history fallback | Route unknown paths to `index.html` | Add deployment test for `/admin/*` direct loads |
| Docker daemon not running | Docker Desktop/Engine stopped | Start daemon and retry `docker ps` | Verify Docker before optional container flow |
| Docker host/API unreachable | Incorrect container/host networking assumption | Use browser-visible host URL; inspect published ports | Document network topology |
| Permission denied writing pnpm store | Store owned by another OS user/root | Use a user-owned pnpm store; repair ownership through OS admin process | Never run normal `pnpm install` with `sudo` |
| Tests fail but browser works | jsdom/setup or stale dependency mismatch | Run locked install, then targeted Vitest with verbose output | Keep tests in quality gate |
| Migration failed | Not a Task 1 frontend operation | Stop and use the owning backend service migration guide | Never apply ad hoc DB scripts from frontend setup |
| Redis/Kafka failure | Not a direct frontend dependency | Diagnose in the owning backend/gateway service | Keep service boundaries clear |

### Useful browser debugging

1. Open DevTools -> Network.
2. Select failed request and inspect URL, status, response, and `X-Request-ID`.
3. Check Console for CORS or configuration errors.
4. Check Application -> Session Storage -> `superadmin.session`.
5. Do not share token contents while reporting the issue; share request ID and sanitized response instead.

## 12. Security & Best Practices

### Security rules

- Never commit `.env`; commit only sanitized `.env.example`.
- Never place secrets in `VITE_*` variables.
- Use HTTPS for deployed API and frontend endpoints.
- Backend must enforce RBAC; hidden menu items are not authorization.
- Use short-lived admin access tokens, MFA, revocation, and audit logs.
- Avoid logging passwords, tokens, raw PII, or entire auth responses.
- Restrict CORS to known frontend origins; do not use wildcard origin with credentials.
- Treat `sessionStorage` token access as an XSS risk. Apply CSP, dependency hygiene, output escaping, and security headers.
- Run typecheck, tests, and production build before merge.
- Review lockfile changes and dependency advisories before upgrades.
- Serve the production bundle with SPA fallback and headers such as CSP, HSTS, `X-Content-Type-Options`, and an appropriate `Referrer-Policy`.

### Configuration audit

| Finding | Risk | Recommended fix |
|---|---|---|
| pnpm version is not pinned in `package.json` | Developer/CI version drift | Add a reviewed `packageManager` value such as the agreed pnpm 10 patch version |
| Node version is not pinned by `.nvmrc`/`.node-version` or `engines` | Vite engine mismatch | Add Node 22 version policy and enforce it in CI |
| `VITE_APP_NAME` is declared but unused | Misleading configuration | Wire it into UI or remove it from template |
| `VITE_ADMIN_SESSION_WARNING_MINUTES` is declared but unused | Developers may assume expiry warning exists | Implement warning/expiry behavior or remove variable |
| Access token is stored in JavaScript-readable `sessionStorage` | XSS could steal admin token | Prefer reviewed secure HttpOnly cookie/BFF design, or strengthen CSP and minimize token lifetime |
| Stored `expiresAt` is not enforced by the route guard | Expired UI session may remain until API rejection | Validate expiry during hydration and clear/refresh session |
| Logout clears only local browser state | Server refresh/session may remain valid | Call backend logout/revocation endpoint before local cleanup |
| No proven frontend/gateway CORS setup | Browser API calls may fail or be over-permissive | Add explicit origin and header configuration in Gateway |
| No frontend Dockerfile/Compose/deployment manifest | Container setup is not reproducible | Add reviewed deployment artifacts only when deployment target is known |
| No app health check | Orchestrator cannot prove static server readiness | Add static server health endpoint/probe in deployment layer |
| Gateway/Superadmin folders are env-only | Full-stack onboarding is blocked | Add runnable source, module manifests, migrations, service README, and health checks |
| Backend `.env` files contain credential-like values | Secret leakage and insecure defaults | Replace with sanitized `.env.example`, rotate real credentials, and inject secrets outside Git |
| Dev server binds `0.0.0.0` | UI can be exposed on LAN | Bind to `127.0.0.1` unless LAN/container access is intentionally required |

> **High-priority audit note:** A credential-like MySQL DSN is present in `backend/services/superadmin-service/.env`. Treat it as compromised if it has ever been real: rotate the password and keep only a placeholder example in version-controlled documentation.

### Operational best practices

- Use separate local, test, staging, and production API origins.
- Backends should expose live/readiness checks and structured logs with request IDs.
- Pin production images and dependencies; avoid floating `latest` tags.
- Keep admin accounts separate from normal customer accounts.
- Apply least privilege to every admin role and default unknown roles to deny.
- Keep a tested incident path for forced admin session revocation.
- No frontend database backup exists; backup policies belong to backend database owners.

## 13. Missing or Misconfigured Things

### Required before full-stack onboarding is complete

- Runnable API Gateway code and start command.
- Runnable Auth/Superadmin service code and service-specific dependency manifests.
- Defined CORS configuration for frontend origins and `Authorization`, `Content-Type`, and `X-Request-ID` headers.
- Backend health/readiness endpoints with exact URLs.
- Valid local admin seed/provisioning process without shared hardcoded credentials.
- Versioned backend migrations and rollback instructions.
- Sanitized backend `.env.example` files instead of credential-bearing `.env` files.
- Pinned Node and pnpm versions.
- Deployment artifacts if Docker/Kubernetes deployment is intended.

### Not missing for Task 1

- Frontend package manifest and lockfile are present.
- Vite, TypeScript, Tailwind, and Vitest configuration are present.
- Frontend `.env.example` is present.
- App scripts for dev, test, typecheck, and build are present.
- Direct database, Redis, queue, and migration configuration is correctly unnecessary for frontend-only Task 1.

## 14. Final Checklist

### Frontend-only setup

- [ ] Git installed and repository cloned
- [ ] Node 22.12+ installed
- [ ] pnpm 10 installed/activated
- [ ] `node --version` and `pnpm --version` verified
- [ ] `pnpm install --frozen-lockfile` completed from `frontend/`
- [ ] `frontend/superadmin-panel/.env` created from `.env.example`
- [ ] No secret added to any `VITE_*` variable
- [ ] Typecheck passes
- [ ] Tests pass
- [ ] Production build passes
- [ ] Vite dev server starts
- [ ] `/admin` redirects unauthenticated user to `/login`
- [ ] Browser console checked for errors

### Full-stack setup

- [ ] Runnable API Gateway supplied and started
- [ ] Auth Service supplied and started
- [ ] Required downstream admin services supplied and started
- [ ] Backend databases/migrations completed using owning service guides
- [ ] Gateway URL matches `VITE_API_BASE_URL`
- [ ] Gateway CORS allows the frontend origin and required headers
- [ ] A least-privilege admin test account is provisioned
- [ ] Real login returns a valid admin session
- [ ] Protected APIs return expected `200`, `401`, and `403` behavior
- [ ] Logout/revocation behavior verified
- [ ] Logs and request IDs checked
- [ ] No credential-bearing `.env` file is committed

### Quick success commands

```bash
cd frontend
pnpm install --frozen-lockfile
cp superadmin-panel/.env.example superadmin-panel/.env
pnpm run superadmin:typecheck
pnpm run superadmin:test
pnpm run superadmin:build
pnpm run superadmin:dev
```

When these frontend commands pass, Task 1 ka local code/tooling setup ready hai. Real login tabhi ready maana jayega jab backend runtime, CORS, admin account, and authorization checks separately verified hon.
