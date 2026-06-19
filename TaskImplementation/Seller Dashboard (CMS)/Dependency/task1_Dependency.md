# Project Dependency & Setup Guide

## 1. Project Overview

This guide explains the dependency, environment, database, Docker, and local setup needed to run the Seller Dashboard (CMS) Task 1 dashboard shell.

Simple Hinglish goal: is project ka frontend seller ke liye protected dashboard shell provide karta hai. User `/seller` route open karta hai, app seller session verify karti hai, active seller milta hai to sidebar, topbar, seller switcher, and main dashboard layout render hota hai.

Important boundary:

- Task 1 ka main code frontend shell hai.
- Frontend app available hai at `frontend/seller-dashboard`.
- Frontend API calls API Gateway ko hit karti hain. Default Gateway URL `http://localhost:8080` hai.
- Backend Go source, migrations folder, Dockerfiles, and docker-compose file clearly present nahi hain. Repo me backend `.env` stubs and architecture docs available hain.
- Isliye beginner developer frontend shell locally run kar sakta hai, but real authenticated seller data ke liye API Gateway, CMS Service, Auth/User services, MySQL, and Redis setup bhi chahiye.

## 2. Tech Stack

| Technology | Required? | What it is | Why used here |
|---|---:|---|---|
| React | Yes | React ek UI library hai jo components banane ke kaam aati hai. | Dashboard shell, sidebar, topbar, routes, pages banane ke liye. |
| TypeScript | Yes | TypeScript JavaScript ka typed version hai. | Seller session, API response, props, and route data safer banane ke liye. |
| Vite | Yes | Vite fast frontend dev server and build tool hai. | Local dev server `5174` par fast React app chalane ke liye. |
| pnpm | Yes | pnpm Node package manager hai. | Monorepo workspace dependencies install and filter commands run karne ke liye. |
| Tailwind CSS v4 | Yes | Utility-first CSS framework hai. | Compact operational dashboard UI style karne ke liye. |
| React Router | Yes | Frontend routing library hai. | `/seller`, protected layout, child pages, and redirects ke liye. |
| TanStack React Query | Yes | Server-state fetching/cache library hai. | Seller session and dashboard API data fetch/cache karne ke liye. |
| Zustand | Yes | Lightweight state management library hai. | Active seller and sidebar state store karne ke liye. |
| Lucide React | Yes | Icon library hai. | Sidebar/topbar icons ke liye. |
| Vitest + Testing Library | Recommended | Frontend unit/component test tools hain. | Route guard, seller switcher, API errors test karne ke liye. |
| API Gateway | Required for real data | Gateway browser REST requests ko internal services tak route karta hai. | Frontend API calls `/api/v1/...` Gateway through jaati hain. |
| Go | Required for backend services | Go backend language hai. | Intended API Gateway/CMS/Auth services Go microservices ke roop me documented hain. |
| MySQL | Required for real CMS APIs | Relational database hai jisme tables me data store hota hai. | CMS coupons, seller staff, settings, analytics/audit data ke liye. |
| Redis | Required if rate limit enabled | In-memory data store/cache hai. | API Gateway rate limiting ke liye `RATE_LIMIT_ENABLED=true` configured hai. |
| gRPC / Protobuf | Required for backend integration | Typed service-to-service communication system hai. | Gateway internal services se gRPC ke through baat karega. |
| Docker | Optional but recommended | Container runtime hai. | MySQL/Redis and future backend services repeatable local setup ke liye. |

## 3. Required Software

Install these before running the project:

| Software | Version / Note | Verify command |
|---|---|---|
| Git | Latest stable | `git --version` |
| Node.js | 22+ recommended by repo docs | `node --version` |
| Corepack | Comes with modern Node | `corepack --version` |
| pnpm | Use through Corepack | `pnpm --version` |
| Docker Desktop / Docker Engine | Recommended for DB/cache | `docker --version` |
| Docker Compose plugin | Recommended | `docker compose version` |
| MySQL | 8+ recommended | `mysql --version` |
| Redis | 7+ recommended when gateway rate limit is enabled | `redis-server --version` |
| Go | Backend docs say Go 1.24+, `backend/go.work` says `go 1.26.3` | `go version` |
| curl | API health checks | `curl --version` |

Beginner note: frontend shell ke liye Node.js + pnpm enough hai. Real login/session/API data ke liye backend dependencies bhi running honi chahiye.

## 4. Dependency Management

### Frontend dependency system: pnpm workspace

This repo uses a frontend pnpm workspace:

```text
frontend/
  package.json
  pnpm-lock.yaml
  pnpm-workspace.yaml
  seller-dashboard/
    package.json
```

Important files:

| File | Purpose |
|---|---|
| `frontend/package.json` | Workspace-level scripts like `build:seller`, `test:seller`, `typecheck:seller`. |
| `frontend/pnpm-workspace.yaml` | Includes `seller-dashboard` workspace package. |
| `frontend/pnpm-lock.yaml` | Exact dependency versions lock karta hai. Isko commit karna chahiye. |
| `frontend/seller-dashboard/package.json` | Seller dashboard app dependencies and scripts. |
| `frontend/seller-dashboard/node_modules` | Installed dependency folder. Isko commit nahi karna. |

Install frontend dependencies:

```bash
cd frontend
corepack enable
pnpm install
```

Run common frontend commands:

```bash
cd frontend
pnpm --filter seller-dashboard dev
pnpm --filter seller-dashboard typecheck
pnpm --filter seller-dashboard test
pnpm --filter seller-dashboard build
pnpm --filter seller-dashboard preview
```

What these commands do:

| Command | Meaning |
|---|---|
| `pnpm install` | `package.json` + `pnpm-lock.yaml` ke according dependencies install karta hai. |
| `pnpm --filter seller-dashboard dev` | Vite dev server start karta hai on port `5174`. |
| `pnpm --filter seller-dashboard typecheck` | TypeScript compile/type errors check karta hai. |
| `pnpm --filter seller-dashboard test` | Vitest tests run karta hai. |
| `pnpm --filter seller-dashboard build` | Production build banata hai. |
| `pnpm --filter seller-dashboard preview` | Built app preview karta hai on port `4174`. |

Common Node/pnpm issues:

| Problem | Cause | Fix |
|---|---|---|
| `pnpm: command not found` | pnpm enabled/install nahi hai | Run `corepack enable`, then `corepack prepare pnpm@latest --activate`. |
| Lockfile mismatch | Package manager/version mismatch | Use pnpm from repo root under `frontend`, avoid npm/yarn. |
| `node_modules` weird errors | Corrupt install | Delete `frontend/node_modules` and `frontend/seller-dashboard/node_modules`, then `pnpm install`. |
| Vite not starting | Wrong directory | Run commands from `frontend`, not repo root. |

### Backend dependency system: Go modules

Backend side intended Go microservices use Go modules/workspace.

Important files:

| File | Purpose |
|---|---|
| `backend/go.work` | Multiple Go service modules ko one workspace me connect karta hai. |
| `backend/go.work.sum` | Workspace dependency checksums. |
| `go.mod` | Har Go service ka module file hona chahiye. |
| `go.sum` | Exact dependency checksums. |

Current repo observation: `backend/go.work` has `use ./services/auth-service`, but backend service source/module files clearly present nahi hain. Isliye backend commands tabhi work karenge jab actual service modules added hon.

Useful Go commands after backend modules exist:

```bash
cd backend
go work sync

cd services/cms-service
go mod tidy
go mod download
go build ./...
go run ./cmd/server
```

Go dependency issues:

| Problem | Cause | Fix |
|---|---|---|
| `go.mod file not found` | Service module missing hai | Correct service folder me jao or module create/add karo. |
| `module ... not in go.work` | Workspace me service included nahi | `go work use ./services/<service-name>` run karo. |
| Version mismatch | Local Go version old hai | Install required Go version. Docs say 1.24+, workspace says 1.26.3. |
| Proxy/cache error | Go module download blocked/corrupt | Try `go clean -modcache`, check network/proxy. |
| Missing generated proto code | Protobuf generation nahi hua | Add proto/buf config, then `buf generate`. |

## 5. Database Setup

### Database used: MySQL

MySQL ek relational database hai jisme data tables, rows, columns ke form me store hota hai. Is project me CMS related data like seller settings, seller staff, coupons, coupon rules, coupon redemptions, campaigns, and audit logs ke liye MySQL required hai.

Required or optional:

- Frontend shell open karne ke liye MySQL optional hai.
- Real authenticated dashboard and CMS APIs ke liye MySQL mandatory hai.

Evidence:

- `backend/services/cms-service/.env` has `CMS_DB_*` and `CMS_MYSQL_DSN`.
- `database/draw.sql` creates `cms_db`.
- `database/draw.sql` contains CMS tables: `seller_settings`, `seller_staff`, `coupons`, `coupon_rules`, `coupon_redemptions`, `campaigns`, `cms_audit_logs`.

Default port:

```text
3306
```

Connection string format:

```env
CMS_MYSQL_DSN=cms_user:change-me@tcp(localhost:3306)/cms_db?parseTime=true&loc=UTC
```

Alternative split config:

```env
CMS_DB_HOST=localhost
CMS_DB_PORT=3306
CMS_DB_NAME=cms_db
CMS_DB_USER=cms_user
CMS_DB_PASSWORD=change-me
CMS_DB_TIMEZONE=UTC
```

### Local MySQL installation

Windows:

```powershell
winget install Oracle.MySQL
```

Alternative Windows option: install "MySQL Installer for Windows" from Oracle, then install MySQL Server 8.x and MySQL Shell/Workbench.

Ubuntu/Debian Linux:

```bash
sudo apt update
sudo apt install mysql-server mysql-client
sudo systemctl enable --now mysql
sudo systemctl status mysql
```

macOS:

```bash
brew install mysql
brew services start mysql
mysql --version
```

### Docker MySQL setup

Use Docker when you want a clean local DB without installing MySQL directly on your OS.

```bash
docker volume create ecommerce-cms-mysql-data

docker run -d \
  --name ecommerce-cms-mysql \
  -p 3306:3306 \
  -e MYSQL_DATABASE=cms_db \
  -e MYSQL_USER=cms_user \
  -e MYSQL_PASSWORD=change-me \
  -e MYSQL_ROOT_PASSWORD=root-change-me \
  -v ecommerce-cms-mysql-data:/var/lib/mysql \
  mysql:8.4
```

Verify MySQL:

```bash
docker ps
mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db
```

### Database schema / migrations

Current repo has `database/draw.sql`, but no dedicated migration folder was clearly found for CMS service. Beginner setup ke liye current available SQL file use karo:

```bash
mysql -u root -p < database/draw.sql
```

If using Docker container:

```bash
docker exec -i ecommerce-cms-mysql mysql -uroot -proot-change-me < database/draw.sql
```

Expected future migration command after proper migration files are added:

```bash
migrate -path backend/services/cms-service/migrations -database "$CMS_MYSQL_DSN" up
```

## 6. Redis / Queue / External Services

### API Gateway

API Gateway frontend ke liye entrypoint hai. Browser REST API call karta hai, Gateway auth, validation, rate limit, and routing handle karta hai.

Required or optional:

- Required for real seller session.
- Frontend default API base URL: `http://localhost:8080`.

Important frontend behavior:

- API base: `VITE_API_BASE_URL`, default `http://localhost:8080`.
- Credentials: `include`, so cookies/session auth send hoti hai.
- Headers: `x-client-app=seller-dashboard`, `x-request-id=<generated>`.

Health check expected by dependency docs:

```bash
curl http://localhost:8080/health/live
```

Current gap: API Gateway source/Dockerfile was not clearly found, only `.env` exists.

### Redis

Redis ek in-memory key-value store hai. Is project me API Gateway rate limiting ke liye Redis configured hai.

Required or optional:

- Required if `RATE_LIMIT_ENABLED=true`.
- Optional only if local gateway config sets `RATE_LIMIT_ENABLED=false`.

Default port:

```text
6379
```

Install locally:

Ubuntu/Debian:

```bash
sudo apt update
sudo apt install redis-server
sudo systemctl enable --now redis-server
redis-cli ping
```

macOS:

```bash
brew install redis
brew services start redis
redis-cli ping
```

Windows:

Use Docker or WSL. Docker command:

```bash
docker run -d --name ecommerce-redis -p 6379:6379 redis:7-alpine
```

Docker setup:

```bash
docker volume create ecommerce-redis-data

docker run -d \
  --name ecommerce-redis \
  -p 6379:6379 \
  -v ecommerce-redis-data:/data \
  redis:7-alpine redis-server --appendonly yes
```

Verify:

```bash
redis-cli -h localhost -p 6379 ping
```

Expected output:

```text
PONG
```

### gRPC / Protobuf

gRPC service-to-service communication ke liye use hota hai. API Gateway internally Auth/User/Product/Order/CMS services ko gRPC addresses se call karega.

Required or optional:

- Required for real backend integration.
- Not required for frontend shell-only rendering.

Important gateway env examples:

```env
AUTH_GRPC_ADDR=localhost:50051
USER_GRPC_ADDR=localhost:50052
PRODUCT_GRPC_ADDR=localhost:50053
ORDER_GRPC_ADDR=localhost:50056
CMS_GRPC_ADDR=localhost:50059
```

Current mismatch to fix: CMS service `.env` says `CMS_GRPC_ADDR=:9098`, but API Gateway `.env` says `CMS_GRPC_ADDR=localhost:50059`. Dono same port/address par align karna hoga.

Verify gRPC service after it exists:

```bash
grpcurl -plaintext localhost:9098 list
```

### Kafka / RabbitMQ

Architecture docs mention Kafka or RabbitMQ for async events. Task 1 dashboard shell directly queue use nahi karta.

Required or optional:

- Optional for Task 1 frontend.
- Required later for full ecommerce event-driven flows if backend implements those features.

### SMTP, Stripe, Twilio, S3, Firebase

Task 1 me directly detected nahi hua. Inko setup karne ki need Task 1 shell ke liye nahi hai.

## 7. Environment Variables

### Where to create env files

Frontend local env file:

```text
frontend/seller-dashboard/.env.local
```

Backend env files already present as stubs:

```text
backend/services/api-gateway/.env
backend/services/cms-service/.env
```

Security note: `.gitignore` real `.env` files ignore karta hai. `.env.example` allowed hai. Real passwords/tokens commit mat karo.

### Frontend `.env.local` example

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_API_TIMEOUT_MS=15000
VITE_LOGIN_URL=/login
```

Frontend env variables:

| Variable | Required? | Example | Purpose | Security note |
|---|---:|---|---|---|
| `VITE_API_BASE_URL` | Optional | `http://localhost:8080` | API Gateway base URL. Missing ho to default same value hai. | Vite variables browser bundle me visible hote hain. Secret mat rakho. |
| `VITE_API_TIMEOUT_MS` | Optional | `15000` | API request timeout in milliseconds. | Secret nahi hai. Positive number rakho. |
| `VITE_LOGIN_URL` | Optional | `/login` | Unauthenticated user ko login page par redirect karne ke liye. | Public URL hai, secret nahi. |

### API Gateway `.env` example

```env
SERVICE_NAME=api-gateway
APP_ENV=local
HTTP_ADDR=:8080
API_BASE_PATH=/api/v1
API_CONTRACT_PATH=../../../api/master-api.json
LOG_LEVEL=debug

GRPC_TLS_ENABLED=false
GRPC_DIAL_TIMEOUT=3s

AUTH_GRPC_ADDR=localhost:50051
USER_GRPC_ADDR=localhost:50052
PRODUCT_GRPC_ADDR=localhost:50053
ORDER_GRPC_ADDR=localhost:50056
CMS_GRPC_ADDR=localhost:9098

RATE_LIMIT_ENABLED=true
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_TLS_ENABLED=false
REDIS_DIAL_TIMEOUT=3s
RATE_LIMIT_KEY_PREFIX=rl:v1
RATE_LIMIT_FAIL_OPEN=false

REQUEST_VALIDATION_ENABLED=true
REQUEST_VALIDATION_DEFAULT_MAX_BODY_BYTES=131072

JWT_ISSUER=ecommerce-auth
JWT_AUDIENCE=ecommerce-api
JWT_ALLOWED_ALGS=RS256
JWT_JWKS_URL=http://localhost:8081/.well-known/jwks.json
JWT_JWKS_CACHE_TTL=5m
JWT_JWKS_FETCH_TIMEOUT=3s
JWT_CLOCK_SKEW=30s
```

Important Gateway variables:

| Variable | Required? | Purpose |
|---|---:|---|
| `HTTP_ADDR` | Yes | Gateway HTTP port. Frontend expects `:8080`. |
| `API_BASE_PATH` | Yes | REST base path, expected `/api/v1`. |
| `API_CONTRACT_PATH` | Recommended | Master API contract location. |
| `AUTH_GRPC_ADDR` | Yes for auth | Auth service gRPC target. |
| `USER_GRPC_ADDR` | Yes for seller profile/session | User service gRPC target. |
| `PRODUCT_GRPC_ADDR` | Yes for product modules | Product service gRPC target. |
| `ORDER_GRPC_ADDR` | Yes for orders | Order service gRPC target. |
| `CMS_GRPC_ADDR` | Yes for CMS data | CMS service gRPC target. |
| `RATE_LIMIT_ENABLED` | Optional | Enables Redis-backed rate limit. |
| `REDIS_ADDR` | Required if rate limit enabled | Redis host/port. |
| `JWT_JWKS_URL` | Yes | JWT public keys URL for token verification. |

### CMS Service `.env` example

```env
CMS_HTTP_ADDR=:8087
CMS_GRPC_ADDR=:9098
CMS_ENV=local

CMS_INTERNAL_AUTH_HEADER=X-Internal-Token
CMS_INTERNAL_AUTH_TOKEN=change-me-local-only
CMS_MAX_BODY_BYTES=1048576

CMS_MYSQL_DSN=
CMS_DB_HOST=localhost
CMS_DB_PORT=3306
CMS_DB_NAME=cms_db
CMS_DB_USER=cms_user
CMS_DB_PASSWORD=change-me
CMS_DB_TIMEZONE=UTC
CMS_DB_MAX_OPEN_CONNS=25
CMS_DB_MAX_IDLE_CONNS=10
CMS_DB_CONN_MAX_LIFETIME_SECONDS=300

CMS_PRODUCT_SERVICE_BASE_URL=http://localhost:8080
CMS_PRODUCT_SERVICE_TIMEOUT=5s
CMS_PRODUCT_SERVICE_AUTH_HEADER=X-Internal-Token
CMS_PRODUCT_SERVICE_AUTH_TOKEN=change-me-local-only

CMS_CAMPAIGN_MAX_DURATION_DAYS=90
CMS_ANALYTICS_DEFAULT_CURRENCY=INR
CMS_ANALYTICS_DEFAULT_RANGE_DAYS=30
CMS_ANALYTICS_MAX_RANGE_DAYS=366
CMS_ANALYTICS_DEFAULT_TOP_PRODUCTS_LIMIT=5
CMS_ANALYTICS_MAX_TOP_PRODUCTS_LIMIT=20
CMS_AUDIT_DEFAULT_RANGE_DAYS=30
CMS_AUDIT_DEFAULT_PAGE_SIZE=20
CMS_AUDIT_MAX_PAGE_SIZE=100
```

Important CMS variables:

| Variable | Required? | Purpose |
|---|---:|---|
| `CMS_HTTP_ADDR` | Yes | CMS HTTP health/API port, currently `:8087`. |
| `CMS_GRPC_ADDR` | Yes | CMS gRPC port, currently `:9098`. |
| `CMS_INTERNAL_AUTH_TOKEN` | Yes | Internal service auth secret. Change from placeholder. |
| `CMS_MYSQL_DSN` | Optional | Full DB DSN override. |
| `CMS_DB_HOST` | Yes if DSN empty | MySQL host. |
| `CMS_DB_PORT` | Yes if DSN empty | MySQL port. |
| `CMS_DB_NAME` | Yes if DSN empty | Database name, expected `cms_db`. |
| `CMS_DB_USER` | Yes if DSN empty | DB username. |
| `CMS_DB_PASSWORD` | Yes if DSN empty | DB password. |
| `CMS_PRODUCT_SERVICE_BASE_URL` | Yes for product integration | Product service boundary URL. |

Common env mistakes:

- `.env.local` frontend file wrong folder me banana.
- Vite env variable without `VITE_` prefix use karna.
- Secret values `VITE_*` me daalna. Browser me expose ho jayega.
- API Gateway `CMS_GRPC_ADDR` and CMS service `CMS_GRPC_ADDR` mismatch.
- MySQL password `.env` me change karna but Docker/MySQL user me update na karna.

## 8. Docker Setup

Current repo observation: no service-specific Dockerfile or docker-compose file clearly found for this task. Docker examples below recommended local setup examples hain.

### Recommended beginner `docker-compose.local.yml`

Create this in future under `infra/compose/docker-compose.local.yml`:

```yaml
services:
  mysql:
    image: mysql:8.4
    container_name: ecommerce-cms-mysql
    environment:
      MYSQL_DATABASE: cms_db
      MYSQL_USER: cms_user
      MYSQL_PASSWORD: change-me
      MYSQL_ROOT_PASSWORD: root-change-me
    ports:
      - "3306:3306"
    volumes:
      - cms_mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 10

  redis:
    image: redis:7-alpine
    container_name: ecommerce-redis
    command: ["redis-server", "--appendonly", "yes"]
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 10

volumes:
  cms_mysql_data:
  redis_data:
```

Run:

```bash
docker compose -f infra/compose/docker-compose.local.yml up -d
docker compose -f infra/compose/docker-compose.local.yml ps
docker compose -f infra/compose/docker-compose.local.yml logs -f
docker compose -f infra/compose/docker-compose.local.yml down
```

When Docker should be used:

- Beginner setup me MySQL/Redis ke liye Docker easiest hai.
- Frontend local dev ke liye direct `pnpm dev` better hai.
- Backend services ke Dockerfiles add hone ke baad full stack compose use karna better hoga.

Docker networking notes:

- Host machine se MySQL: `localhost:3306`.
- Container-to-container MySQL: `mysql:3306`.
- Host machine se Redis: `localhost:6379`.
- Container-to-container Redis: `redis:6379`.
- Frontend browser always host URL hit karta hai, so `VITE_API_BASE_URL=http://localhost:8080` local browser ke liye correct hai.

## 9. Local Development Setup

### Step 1: Clone repository

```bash
git clone <repo-url>
cd Ecommerce
```

### Step 2: Install frontend dependencies

```bash
cd frontend
corepack enable
pnpm install
```

### Step 3: Create frontend env

Create `frontend/seller-dashboard/.env.local`:

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_API_TIMEOUT_MS=15000
VITE_LOGIN_URL=/login
```

### Step 4: Start only the frontend shell

```bash
cd frontend
pnpm --filter seller-dashboard dev
```

Open:

```text
http://localhost:5174
```

Expected behavior without backend:

- App loads.
- `/seller` guard tries `GET http://localhost:8080/api/v1/seller/session`.
- If Gateway not running, app may show auth/network state depending route handling.
- For full success state, backend session endpoint must return authenticated active seller.

### Step 5: Start database/cache for real backend

```bash
docker run -d --name ecommerce-cms-mysql \
  -p 3306:3306 \
  -e MYSQL_DATABASE=cms_db \
  -e MYSQL_USER=cms_user \
  -e MYSQL_PASSWORD=change-me \
  -e MYSQL_ROOT_PASSWORD=root-change-me \
  mysql:8.4

docker run -d --name ecommerce-redis -p 6379:6379 redis:7-alpine
```

### Step 6: Load CMS DB schema

```bash
docker exec -i ecommerce-cms-mysql mysql -uroot -proot-change-me < database/draw.sql
```

### Step 7: Configure backend env

Review and align:

```text
backend/services/api-gateway/.env
backend/services/cms-service/.env
```

Minimum alignment:

```env
# Gateway should point to actual CMS gRPC address
CMS_GRPC_ADDR=localhost:9098

# CMS should point to local MySQL
CMS_DB_HOST=localhost
CMS_DB_PORT=3306
CMS_DB_NAME=cms_db
CMS_DB_USER=cms_user
CMS_DB_PASSWORD=change-me
```

### Step 8: Start backend services

Current repo does not clearly contain backend service entrypoints. After they are added, expected commands will look like:

```bash
cd backend/services/cms-service
go run ./cmd/server

cd backend/services/api-gateway
go run ./cmd/server
```

Also required for real seller session:

- Auth service or JWT/JWKS endpoint.
- User/seller profile service.
- API route that serves `GET /api/v1/seller/session`.

### Step 9: Verify APIs

```bash
curl http://localhost:8080/health/live
curl http://localhost:8087/health/live
curl -i http://localhost:8080/api/v1/seller/session
```

For authenticated session, browser/curl must include valid auth cookies or token depending backend implementation.

## 10. Running the Project

### Frontend dev server

```bash
cd frontend
pnpm --filter seller-dashboard dev
```

URL:

```text
http://localhost:5174
```

### Frontend production build

```bash
cd frontend
pnpm --filter seller-dashboard build
pnpm --filter seller-dashboard preview
```

Preview URL:

```text
http://localhost:4174
```

### Quality checks

```bash
cd frontend
pnpm --filter seller-dashboard typecheck
pnpm --filter seller-dashboard test
```

### Ports and networking

| Service | Port | Required? | Purpose |
|---|---:|---:|---|
| Seller Dashboard Vite dev | 5174 | Yes for frontend dev | React dev server. |
| Seller Dashboard preview | 4174 | Optional | Preview production build. |
| API Gateway HTTP | 8080 | Required for real data | Frontend REST API base. |
| CMS Service HTTP | 8087 | Required for CMS health/direct checks | CMS health/internal HTTP. |
| CMS Service gRPC | 9098 | Required for Gateway-to-CMS | Internal gRPC API. |
| MySQL | 3306 | Required for CMS backend | CMS relational database. |
| Redis | 6379 | Required if rate limit enabled | Gateway rate limiting/cache. |
| Auth JWKS | 8081 | Required for JWT validation if configured | Public key endpoint. |

Port conflict fixes:

```bash
# Linux/macOS
lsof -i :5174
lsof -i :8080

# Docker
docker ps
docker stop <container-name>
```

Change Vite port in:

```text
frontend/seller-dashboard/vite.config.ts
```

Change API Gateway port in:

```text
backend/services/api-gateway/.env
```

Change frontend API target in:

```text
frontend/seller-dashboard/.env.local
```

## 11. Common Errors & Fixes

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `pnpm: command not found` | Corepack/pnpm not enabled | `corepack enable` then reopen terminal | Use Node 22+ and Corepack. |
| `ERR_PNPM_NO_MATCHING_VERSION` | Bad package/version resolution | Run from `frontend`, use committed lockfile | Do not mix npm/yarn/pnpm. |
| Vite port already in use | Another app uses `5174` | Stop old process or change Vite port | Check `lsof -i :5174` before running. |
| Frontend network error | API Gateway not running | Start Gateway or set `VITE_API_BASE_URL` | Keep `.env.local` documented. |
| Redirect/login loop | No valid seller session | Login with seller account or mock backend session | Ensure Auth/User services are ready. |
| `401 Unauthorized` | Missing/invalid cookie/JWT | Login again, verify JWKS and cookies | Do not manually edit auth cookies. |
| `403 Permission denied` | Seller missing, pending, or suspended | Activate seller profile in backend/admin flow | Seed active seller test account. |
| `GET /api/v1/seller/session` 404 | Endpoint not implemented/contract missing | Add Gateway route/session API or align frontend to existing endpoint | Keep frontend and API contract synced. |
| MySQL connection refused | MySQL not running/wrong port | Start MySQL, check `localhost:3306` | Use Docker healthcheck. |
| MySQL access denied | Wrong user/password | Match `CMS_DB_USER` and `CMS_DB_PASSWORD` with MySQL user | Store local credentials in `.env`, not memory. |
| Migration/schema failed | DB not created or SQL run in wrong order | Run `database/draw.sql` as root/admin | Add proper up/down migrations. |
| Redis connection refused | Redis not running | Start Redis or disable local rate limit | Use compose with healthcheck. |
| `NOAUTH Authentication required` | Redis has password but env empty | Set `REDIS_PASSWORD` | Keep Redis env consistent. |
| Gateway cannot dial CMS | gRPC address mismatch | Align `CMS_GRPC_ADDR` in Gateway and CMS service | Maintain a ports table. |
| Docker daemon not running | Docker Desktop/Engine stopped | Start Docker | Verify `docker ps` first. |
| Permission denied | OS/docker permissions | Linux: add user to docker group or use sudo | Avoid running random chmod commands. |
| `go.mod file not found` | Backend service module missing | Add service `go.mod` and source or use correct folder | Keep `go.work` updated. |

## 12. Security & Best Practices

- Never commit real `.env` files.
- Create sanitized `.env.example` files for frontend, API Gateway, and CMS service.
- Never put private tokens, DB passwords, JWT private keys, or internal auth tokens in `VITE_*` variables. Vite env browser me public hota hai.
- Replace placeholder secrets like `change-me` before shared/staging/prod usage.
- Use strong JWT keys and JWKS endpoint for Gateway validation.
- Use separate local, staging, and production env files.
- Enable Redis password/TLS in production.
- Use MySQL least-privilege user. Root user se app run mat karo.
- Keep Docker volumes for persistent DB data.
- Add health checks for MySQL, Redis, API Gateway, and CMS service.
- Log request IDs, but do not log JWTs, cookies, passwords, or internal tokens.
- Keep API contract and frontend endpoint paths synced.
- Run `pnpm --filter seller-dashboard typecheck` and tests before pushing changes.
- Add DB migrations with up/down files instead of only one large SQL dump.

## 13. Missing or Misconfigured Things

| Area | Current finding | Impact | Suggested fix |
|---|---|---|---|
| API contract | Frontend calls `/api/v1/seller/session`, but `api/master-api.json` does not clearly list that path. | Protected shell cannot authenticate against documented contract. | Add route to API contract/Gateway or change frontend to documented seller profile/session API. |
| Backend source | Only `backend/go.work`, `go.work.sum`, and service `.env` files were clearly found. | Gateway/CMS cannot be started from source in current repo shape. | Add Go service modules, entrypoints, handlers, config loaders, and tests. |
| Go workspace | `backend/go.work` references `./services/auth-service`, but module/source not clearly present. | Go workspace commands may fail. | Add auth service module or update workspace to real modules. |
| CMS gRPC port | CMS `.env` has `:9098`; Gateway `.env` had `localhost:50059`. | Gateway cannot reach CMS if both use different ports. | Align to one value, for example `localhost:9098`. |
| Docker Compose | No compose file clearly found. | Beginner full-stack startup is manual. | Add `infra/compose/docker-compose.local.yml`. |
| Dockerfiles | Frontend and service Dockerfiles not clearly found. | App cannot be containerized consistently yet. | Add service Dockerfiles and frontend Nginx/static Dockerfile. |
| Migrations | No CMS migration folder clearly found. | DB setup/rollback not production-safe. | Add `backend/services/cms-service/migrations/*.up.sql` and `*.down.sql`. |
| `.env.example` | Not clearly found. | Beginners may copy unsafe real env or miss required vars. | Add sanitized examples for every service/app. |
| Secrets | Existing `.env` files contain placeholders like `change-me`. | Safe for local stubs, unsafe if reused in shared env. | Rotate and use secret manager for real envs. |
| Health checks | Health endpoints expected in docs, implementation not clearly found. | Hard to debug readiness. | Add `/health/live` and `/health/ready` in services. |

## 14. Final Checklist

- [ ] Git installed.
- [ ] Node.js 22+ installed.
- [ ] Corepack/pnpm enabled.
- [ ] Frontend dependencies installed with `cd frontend && pnpm install`.
- [ ] `frontend/seller-dashboard/.env.local` created.
- [ ] Seller Dashboard dev server starts on `http://localhost:5174`.
- [ ] Typecheck passes with `pnpm --filter seller-dashboard typecheck`.
- [ ] Tests pass with `pnpm --filter seller-dashboard test`.
- [ ] MySQL running if real CMS backend is needed.
- [ ] `cms_db` schema loaded from `database/draw.sql` or future migrations.
- [ ] Redis running if Gateway rate limiting is enabled.
- [ ] API Gateway configured on `http://localhost:8080`.
- [ ] Gateway `CMS_GRPC_ADDR` aligned with CMS service gRPC address.
- [ ] Auth/JWKS service available for JWT validation.
- [ ] Seller session endpoint available and returns active seller for test account.
- [ ] Real secrets are not committed.
- [ ] Common errors section reviewed before debugging.
