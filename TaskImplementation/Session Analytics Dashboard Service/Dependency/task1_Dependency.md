# Project Dependency & Setup Guide

> **Scope:** Ye handbook analytics shell Task 1 ko locally install, configure, run, test, aur troubleshoot karne ke liye hai. Business logic ya API implementation is document ka part nahi hai.
>
> **Repository reality check:** Frontend shell source available hai, lekin documented `GET /api/v1/analytics/live` backend route, runnable API Gateway, admin-auth integration, analytics collections/migrations, aur application Dockerfiles abhi repository me complete nahi hain. Isliye UI build/tests run ho sakte hain, but real metric cards ka end-to-end flow abhi complete nahi chalega.

## 1. Project Overview

Task 1 ek React dashboard shell provide karta hai:

- date range aur segment filters;
- live metric cards;
- loading, empty, error, aur retry states;
- `GET /api/v1/analytics/live` API client;
- cookie-based admin request support through `credentials: "include"`;
- unit/component/API-client tests.

### Runtime flow

```text
Browser (Vite :5174)
  -> /api/v1/analytics/live
  -> Vite development proxy
  -> intended API Gateway (:8080)
  -> intended admin authentication/authorization
  -> intended live-metrics handler in Session Service
  -> MongoDB + Redis
```

Current repository me last four integration steps complete nahi hain. Available Session Service `:8086` par HTTP privacy routes aur `/healthz` expose karta hai; live-metrics route expose nahi karta.

### Three useful setup modes

| Mode | Kya run hoga | Status |
|---|---|---|
| Frontend development | UI, local navigation, loading/error behavior | Available |
| Frontend tests/build | Unit tests, API client mocks, TypeScript, production bundle | Available after npm install |
| Full live analytics | Gateway + admin auth + live metrics + stored analytics | **Blocked by missing implementation** |

## 2. Tech Stack

### Frontend technologies

| Technology | Simple explanation | Project me use | Required? |
|---|---|---|---|
| React 18 | React component-based UI library hai. | Layout, filters, cards, states render karta hai. | Yes |
| TypeScript 5 | JavaScript ke upar type safety deta hai. | API response aur component props mismatch jaldi pakadta hai. | Yes |
| Vite 5 | Fast dev server aur production bundler hai. | Local server, `/api` proxy, build, preview. | Yes |
| Tailwind CSS 3 | Utility CSS framework hai. | Dashboard layout aur responsive styling. | Yes |
| PostCSS + Autoprefixer | CSS ko process karke browser prefixes add karte hain. | Tailwind build pipeline. | Yes |
| TanStack Query 5 | Server-state fetching/cache library hai. | Loading, retry, cache, refetch manage karta hai. | Yes |
| React Router 6 | Browser-side routing library hai. | Overview aur future analytics pages ke routes. | Yes |
| Lucide React | Lightweight icon library hai. | Filters, cards, warnings, navigation icons. | Yes |
| date-fns | Date utility library hai. | Package me installed; Task 1 ke current shell code me direct requirement limited hai. | Not used directly by Task 1, but npm installs it |
| Recharts | React charts library hai. | Later funnel/cohort routes use karte hain; Task 1 cards ko charts nahi chahiye. Current `app.tsx` all routes import karta hai, so full package build ke liye dependency required hai. | Required by current package build |
| Vitest | Vite-friendly test runner hai. | Unit aur component tests. | Development only |
| Testing Library + jsdom | Browser-like component testing tools hain. | User-visible behavior test karte hain. | Development only |
| MSW | Network mocking library hai. | API mocks ke liye installed; Task 1 client tests currently fetch stubs bhi use karte hain. | Development only |

### Backend and infrastructure dependencies

| Technology/service | Why needed | Required? | Current status |
|---|---|---|---|
| API Gateway | Browser `/api/v1/...` request ko backend tak route aur authorize karega. | Real data ke liye yes | Only `.env` exists; runnable gateway code absent |
| Admin Auth/JWKS | Analytics contract `admin` role maangta hai. | Real data ke liye yes | End-to-end flow not ready |
| Session Service | Live metrics compute/serve karega. | Real data ke liye yes | Go service exists, but live route/gRPC method absent |
| MongoDB | Historical sessions, raw events, summaries, aggregates store karega. | Backend/full flow ke liye yes | Client wiring exists; Task 1 analytics schema/migration absent |
| Redis | Active sessions aur short-lived cache/counters store karega. | Backend/full flow ke liye yes | Client wiring exists |
| Go 1.26.3 | Current Session Service compile/run karne ka runtime. | Backend work ke liye yes | Version pinned in `go.mod` and `go.work` |
| Docker + Compose | MongoDB/Redis ko isolated local containers me run karta hai. | Optional, recommended | Installed locally may vary; repo Compose file absent |
| Kafka/RabbitMQ | High-volume event ingestion ko async banane ka architecture suggestion hai. | **No, Task 1 ke liye optional** | No code/config/container exists |

> **Important:** MySQL, SMTP, S3, Stripe, NATS, Elasticsearch, MinIO, aur Kubernetes Task 1 shell ki direct dependencies nahi hain. Admin-auth ecosystem later MySQL use kar sakta hai, but current dashboard setup ko aise missing auth stack par pretend-run nahi karna chahiye.

## 3. Required Software

### Minimum versions

| Software | Recommended baseline | Verify |
|---|---:|---|
| Git | Recent stable | `git --version` |
| Node.js | 22 LTS, matching package type definitions | `node --version` |
| npm | Node 22 ke saath bundled version | `npm --version` |
| Modern browser | Current Chrome, Edge, or Firefox | Browser About page |
| Go | Exactly `1.26.3` for current backend modules | `go version` |
| Docker Desktop/Engine | Recent version with Compose v2+ | `docker --version` |
| Docker Compose | Plugin syntax supported | `docker compose version` |
| mongosh | Optional Mongo CLI | `mongosh --version` |
| redis-cli | Optional Redis CLI | `redis-cli --version` |

Frontend-only development ke liye Git + Node.js + npm + browser enough hain. Go, MongoDB, aur Redis backend work ke liye chahiye.

### Windows

1. Git aur Node.js LTS install karo:

   ```powershell
   winget install --id Git.Git -e
   winget install --id OpenJS.NodeJS.LTS -e
   ```

2. MongoDB/Redis ke easiest local setup ke liye Docker Desktop install karo:

   ```powershell
   winget install --id Docker.DockerDesktop -e
   ```

3. Docker Desktop settings me WSL 2 backend enable karo. WSL 1 supported nahi hai.
4. Naya PowerShell/terminal open karke verification commands run karo.
5. Backend code par kaam karna ho to official Go installer se `1.26.3` install karo.

Native Redis Windows service officially preferred path nahi hai; Docker Desktop ya WSL 2 use karo.

### Ubuntu/Debian Linux

```bash
sudo apt update
sudo apt install -y git ca-certificates curl
```

Node 22 ko `nvm`, `fnm`, ya official Node distribution se install karo; distro ka old `nodejs` package blindly use mat karo. Backend ke liye official Go archive/package se `1.26.3` install karo. Docker Engine install karne ke baad current user ko Docker access dena optional hai:

```bash
sudo usermod -aG docker "$USER"
```

Group change ke baad sign out/in required hota hai. Native Redis alternative:

```bash
sudo apt install -y redis-server
sudo systemctl enable --now redis-server
```

MongoDB ke liye distro/version-specific official `mongodb-org` repository instructions follow karo, ya is guide ka Docker setup use karo.

### macOS

Homebrew installed ho to:

```bash
brew install git node@22 go
brew install --cask docker
```

Native optional services:

```bash
brew tap mongodb/brew
brew install mongodb-community
brew services start mongodb-community
brew install redis
brew services start redis
```

Go install ke baad confirm karo ki active version repository ke `go.mod` se match karti hai.

## 4. Dependency Management

### Node.js/npm system

Frontend package root:

```text
frontend/session-analytics-dashboard/package.json
```

- `package.json` dependencies, dev dependencies, aur scripts declare karta hai.
- `node_modules/` downloaded packages rakhta hai; isko Git me commit nahi karna.
- Repository me frontend lockfile (`package-lock.json`, `pnpm-lock.yaml`, ya `yarn.lock`) currently missing hai.
- Lockfile na hone ki wajah se clean installs time ke saath slightly different transitive versions resolve kar sakte hain.

Install:

```bash
cd frontend/session-analytics-dashboard
npm install
```

Available scripts:

```bash
npm run dev        # Vite dev server
npm run build      # TypeScript + production bundle
npm run preview    # built bundle preview
npm run test       # one-shot test suite
npm run test:watch # watch mode
npm run typecheck  # TypeScript only
```

`npm ci` tab use karo jab generated `package-lock.json` review karke repository me commit ho chuka ho:

```bash
npm ci
```

Common npm failures:

- `ERESOLVE`: peer-dependency mismatch; Node version verify karo aur error ko inspect karo. `--force` first response nahi hona chahiye.
- `EACCES`: global npm ko `sudo` se run karne ke badle version manager use karo.
- corrupt cache: `npm cache verify`, phir clean install try karo.
- WSL 1/npm path error: WSL 2 use karo aur Node ko same environment ke andar install karo.

### Go modules

Backend module root:

```text
backend/services/session-service/go.mod
backend/services/session-service/go.sum
```

- `go.mod` module path, Go version, aur direct/indirect dependencies declare karta hai.
- `go.sum` downloaded modules ke cryptographic checksums store karta hai.
- Go modules reproducible dependency resolution provide karte hain.
- `go mod download` declared modules download karta hai.
- `go build` binary compile karta hai.
- `go run` temporary build karke server start karta hai.
- `go mod tidy` unused dependencies remove aur missing ones add karta hai; ye files modify kar sakta hai, isliye blindly onboarding step ke roop me mat run karo.

Imports deliberately change karne ke baad only:

```bash
GOWORK=off go mod tidy
git diff -- go.mod go.sum
```

Current `backend/go.work` missing `auth-service` module ko reference karta hai. Session module commands ko reliable banane ke liye workspace disable karo:

```bash
cd backend/services/session-service
GOWORK=off go mod download
GOWORK=off go build ./...
GOWORK=off go test ./...
```

Windows PowerShell equivalent:

```powershell
$env:GOWORK = "off"
go mod download
go build ./...
go test ./...
```

Common Go module issues:

| Problem | Cause | Fix |
|---|---|---|
| `go.mod requires go >= 1.26.3` | Old Go runtime | Correct Go version install/select karo |
| workspace module not found | `go.work` absent auth module point karta hai | Session commands ke liye `GOWORK=off` |
| checksum mismatch | Proxy/cache or changed dependency | Source verify karo, `go clean -modcache` only after investigation |
| proxy timeout | Network/proxy/DNS issue | `go env GOPROXY`; corporate proxy settings verify karo |
| private module auth | Git credentials unavailable | Proper Git credential/token configure karo; token source me mat likho |

## 5. Database Setup

### MongoDB

MongoDB document database hai, yani JSON-like documents collections me store hote hain. Architecture me session events, journey summaries, aur analytics aggregates ke liye use hota hai. Current Go server startup ke liye MongoDB reachable hona mandatory hai.

| Item | Value |
|---|---|
| Default port | `27017` |
| Local database | `session_db` |
| Local URI | `mongodb://localhost:27017` |
| Config variables | `SESSION_MONGO_URI`, `SESSION_MONGO_DATABASE` |
| Persistence | Docker named volume `/data/db` |

#### Docker run setup

```bash
docker volume create ecommerce-session-mongo-data
docker run -d \
  --name ecommerce-session-mongo \
  --restart unless-stopped \
  -p 127.0.0.1:27017:27017 \
  -v ecommerce-session-mongo-data:/data/db \
  mongo:8
```

Ye no-password configuration **sirf localhost development** ke liye hai. Port LAN/public interface par publish mat karo.

Verify:

```bash
docker exec ecommerce-session-mongo mongosh --quiet --eval 'db.adminCommand({ ping: 1 })'
docker logs ecommerce-session-mongo
```

Native service start commands:

```bash
# Ubuntu/Debian after official mongodb-org installation
sudo systemctl enable --now mongod
sudo systemctl status mongod

# macOS/Homebrew
brew services start mongodb-community
brew services list
```

Windows MSI install me **Install MongoD as a Service** select karo; then:

```powershell
Get-Service MongoDB
Start-Service MongoDB
```

#### Credentials and connection string

Authenticated URI format:

```env
SESSION_MONGO_URI=mongodb://USERNAME:PASSWORD@localhost:27017/session_db?authSource=admin
SESSION_MONGO_DATABASE=session_db
```

Username/password ko ignored backend `.env`, secret manager, ya container secret me rakho. Frontend `.env.local` me database credentials kabhi mat rakho. Special password characters URI-encode karne padte hain.

#### Migrations

Task 1 live analytics ke liye koi collection/index migration checked in nahi hai. Existing migration privacy controls ke later feature ki hai, overview metrics ki nahi. Agar current privacy backend run kar rahe ho, optional existing migration:

```bash
cd backend/services/session-service
mongosh mongodb://localhost:27017/session_db migrations/001_privacy_controls.mongodb.js
```

Rollback only when deliberately removing those indexes:

```bash
mongosh mongodb://localhost:27017/session_db migrations/001_privacy_controls.down.mongodb.js
```

Current server privacy repository kuch required indexes startup par bhi ensure karta hai. Production me migrations ko reviewed CI/CD step banana better hai.

### Redis

Redis in-memory data store hai. Architecture me active sessions, counters, aur short cache windows ke liye use hota hai. Current Go server startup par Redis ping mandatory hai.

| Item | Value |
|---|---|
| Default port | `6379` |
| Current logical DB | `2` |
| Local address | `localhost:6379` |
| Config variables | `SESSION_REDIS_ADDR`, `SESSION_REDIS_PASSWORD`, `SESSION_REDIS_DB` |
| Persistence | AOF + Docker named volume |

Docker run setup:

```bash
docker volume create ecommerce-session-redis-data
docker run -d \
  --name ecommerce-session-redis \
  --restart unless-stopped \
  -p 127.0.0.1:6379:6379 \
  -v ecommerce-session-redis-data:/data \
  redis:7-alpine redis-server --appendonly yes
```

Verify:

```bash
docker exec ecommerce-session-redis redis-cli ping
# Expected: PONG
```

Native service commands:

```bash
# Ubuntu/Debian
sudo systemctl enable --now redis-server
redis-cli ping

# macOS
brew services start redis
redis-cli ping
```

Production Redis me password/TLS/network policy enable karo. Password backend secret storage me rakho:

```env
SESSION_REDIS_ADDR=localhost:6379
SESSION_REDIS_PASSWORD=change-to-a-secret
SESSION_REDIS_DB=2
```

## 6. Redis / Queue / External Services

### API Gateway and admin authentication

Real dashboard response ke liye intended Gateway port `8080` hai. Frontend dev proxy isi target ko use karta hai. API contract live endpoint ko `admin` protected mark karta hai, aur frontend browser cookies include karta hai.

Current limitations:

- Gateway folder me only ignored `.env` hai; executable source/module absent hai.
- Session Service current implementation HTTP `:8086` use karta hai, while Gateway config intended gRPC target `:50060` declare karti hai.
- Session Service me gRPC server/method aur `GET /api/v1/analytics/live` HTTP route dono absent hain.
- Admin login/session-cookie flow dashboard ke saath wired nahi hai.

Therefore `VITE_API_PROXY_TARGET=http://localhost:8086` set karna Task 1 ko fix nahi karega; current backend us path par `404` return karega.

### Kafka or RabbitMQ

Architecture notes high-volume downstream event processing ke liye Kafka/RabbitMQ suggest karte hain. Task 1 code, package manifests, aur local config me broker integration nahi hai. Inko install/start karna current task ke liye unnecessary hai. Broker ko mandatory tabhi banao jab producer, consumer, topic/queue config, retry/DLQ policy, aur health check implementation repository me aaye.

### Docker

Docker optional tool hai, external runtime dependency nahi. Beginner ke liye MongoDB aur Redis containers recommended hain because install/uninstall predictable rehta hai. Application images currently available nahi hain.

## 7. Environment Variables

### Frontend `.env.local`

Create:

```text
frontend/session-analytics-dashboard/.env.local
```

Content:

```env
# Empty means same-origin /api requests; Vite proxy handles them in development.
VITE_API_BASE_URL=

# Intended Gateway target. It is not runnable in the current repository state.
VITE_API_PROXY_TARGET=http://localhost:8080

# Positive integer in milliseconds.
VITE_REQUEST_TIMEOUT_MS=10000
```

| Variable | Purpose | Required? | Security note |
|---|---|---|---|
| `VITE_API_BASE_URL` | Browser fetch base URL. Empty value same-origin path use karta hai. | No; empty recommended locally | Every `VITE_` value browser bundle me visible hoti hai; secret mat rakho |
| `VITE_API_PROXY_TARGET` | Vite dev server `/api` proxy destination. | No; defaults to `http://localhost:8080` | Server-side dev config hai, but still secret storage nahi |
| `VITE_REQUEST_TIMEOUT_MS` | Request abort timeout. | No; default `10000` | Positive integer use karo |

Vite `.env`, `.env.local`, `.env.development`, etc. automatically load karta hai. Change ke baad dev server restart karo. `VITE_API_BASE_URL` direct cross-origin URL set karoge to backend CORS + credential policy properly configured honi chahiye.

### Current Session Service environment

The Go code `.env` automatically load **nahi** karta; it calls `os.Getenv`. Minimal local file:

```text
backend/services/session-service/.env
```

```env
SESSION_HTTP_ADDR=:8086
SESSION_HTTP_READ_TIMEOUT=5s
SESSION_HTTP_WRITE_TIMEOUT=10s
SESSION_HTTP_IDLE_TIMEOUT=60s
SESSION_SHUTDOWN_TIMEOUT=10s

SESSION_MONGO_URI=mongodb://localhost:27017
SESSION_MONGO_DATABASE=session_db
SESSION_MONGO_CONNECT_TIMEOUT=5s

SESSION_REDIS_ADDR=localhost:6379
SESSION_REDIS_PASSWORD=
SESSION_REDIS_DB=2
SESSION_REDIS_DIAL_TIMEOUT=2s
SESSION_REDIS_READ_TIMEOUT=2s
SESSION_REDIS_WRITE_TIMEOUT=2s
SESSION_REDIS_KEY_PREFIX=session

# Required by current config validation. Generate a unique local value.
SESSION_PRIVACY_HASH_PEPPER=replace-with-a-long-random-value
SESSION_PRIVACY_DELETION_LIST_LIMIT=50
```

Generate a pepper:

```bash
openssl rand -hex 32
```

| Variable | Purpose and example | Required? | Security/validation note |
|---|---|---|---|
| `SESSION_HTTP_ADDR` | Listen address, e.g. `:8086` | No; defaults to `:8086` | Public bind behind trusted proxy/firewall only |
| `SESSION_HTTP_READ_TIMEOUT` | Request read limit, e.g. `5s` | No | Valid Go duration use karo |
| `SESSION_HTTP_WRITE_TIMEOUT` | Response write limit, e.g. `10s` | No | Long value resource exhaustion hide kar sakta hai |
| `SESSION_HTTP_IDLE_TIMEOUT` | Keep-alive idle limit, e.g. `60s` | No | Valid Go duration use karo |
| `SESSION_SHUTDOWN_TIMEOUT` | Graceful stop budget, e.g. `10s` | No | Deployment termination grace se align karo |
| `SESSION_MONGO_URI` | Mongo server/credentials, e.g. `mongodb://localhost:27017` | Default exists; reachable DB mandatory | Secret-bearing URI log/commit mat karo |
| `SESSION_MONGO_DATABASE` | Database name, e.g. `session_db` | No; default exists | Dev/prod database names separate rakho |
| `SESSION_MONGO_CONNECT_TIMEOUT` | Startup connection budget, e.g. `5s` | No | Invalid value silently default par fall back hoti hai |
| `SESSION_REDIS_ADDR` | Redis host and port, e.g. `localhost:6379` | Default exists; reachable Redis mandatory | Container me service hostname use karo |
| `SESSION_REDIS_PASSWORD` | Redis authentication secret | No for unprotected local Redis | Production me secret manager/env injection use karo |
| `SESSION_REDIS_DB` | Logical database number, e.g. `2` | No | Same instance consumers ke DB/key prefix isolate karo |
| `SESSION_REDIS_DIAL_TIMEOUT` | Connect budget, e.g. `2s` | No | Valid Go duration use karo |
| `SESSION_REDIS_READ_TIMEOUT` | Command read budget, e.g. `2s` | No | Too-low setting transient failures cause kar sakti hai |
| `SESSION_REDIS_WRITE_TIMEOUT` | Command write budget, e.g. `2s` | No | Too-low setting transient failures cause kar sakti hai |
| `SESSION_REDIS_KEY_PREFIX` | Key namespace, e.g. `session` | No | Shared Redis me collision-free value use karo |
| `SESSION_PRIVACY_HASH_PEPPER` | Identifier hashing secret | **Yes** | Empty startup fail karta hai; unique strong secret use karo |
| `SESSION_PRIVACY_DELETION_LIST_LIMIT` | Deletion list page size, e.g. `50` | No | Must be `1..200` |

Bash/zsh me load and run:

```bash
cd backend/services/session-service
set -a
source .env
set +a
GOWORK=off go run ./cmd/server
```

PowerShell me variables process scope me set karo before `go run`:

```powershell
$env:SESSION_HTTP_ADDR = ":8086"
$env:SESSION_MONGO_URI = "mongodb://localhost:27017"
$env:SESSION_MONGO_DATABASE = "session_db"
$env:SESSION_REDIS_ADDR = "localhost:6379"
$env:SESSION_REDIS_DB = "2"
$env:SESSION_PRIVACY_HASH_PEPPER = "replace-with-a-long-random-value"
$env:GOWORK = "off"
go run ./cmd/server
```

> Existing backend `.env` contains many future configuration names that current `config.go` does not read. Environment variable sirf file me likhne se feature active nahi hota; loader and implementation dono required hain.

### Common environment mistakes

- `.env.local` wrong directory me banana.
- Vite dev server restart na karna.
- Browser secret ko `VITE_` variable me rakhna.
- Backend `.env` create karke assume karna ki Go auto-load karega.
- Docker container ke andar `localhost` use karna; wahan service name, e.g. `mongo:27017`, required hota hai.
- Password me `#`, `@`, `:` jaise characters ko URI-encode na karna.
- Real credentials commit karna. `.gitignore` env files ignore karta hai; `.env.example` me placeholders only hone chahiye.

## 8. Docker Setup

Repository me Dockerfile ya Compose file currently nahi hai. Neeche **documentation example** hai, checked-in file nahi. Isse sirf MongoDB + Redis run honge, frontend/backend images nahi.

Create a local, uncommitted `compose.session-analytics.local.yml` at repository root if needed:

```yaml
services:
  mongo:
    image: mongo:8
    restart: unless-stopped
    ports:
      - "127.0.0.1:27017:27017"
    volumes:
      - session_mongo_data:/data/db
    healthcheck:
      test: ["CMD", "mongosh", "--quiet", "--eval", "db.adminCommand({ ping: 1 })"]
      interval: 10s
      timeout: 5s
      retries: 10

  redis:
    image: redis:7-alpine
    restart: unless-stopped
    command: ["redis-server", "--appendonly", "yes"]
    ports:
      - "127.0.0.1:6379:6379"
    volumes:
      - session_redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 10

volumes:
  session_mongo_data:
  session_redis_data:
```

Commands:

```bash
docker compose -f compose.session-analytics.local.yml up -d
docker compose -f compose.session-analytics.local.yml ps
docker ps
docker compose -f compose.session-analytics.local.yml logs -f
docker compose -f compose.session-analytics.local.yml down
```

Data delete karna explicitly intended ho tabhi `down -v` use karo. Named volumes container recreation ke baad data preserve karte hain.

`Dockerfile` application image banane ki recipe hoti hai; `compose` multiple containers, volumes, health checks, aur networking ko ek saath define karta hai. Upar ke services `restart: unless-stopped` use karte hain, so Docker restart ke baad containers recover kar sakte hain. Compose automatically ek private default network banata hai jahan services apne names (`mongo`, `redis`) se resolve hoti hain.

### Docker networking mental model

- Host par Go process: `localhost:27017`, `localhost:6379`.
- Compose network ke andar app container: `mongo:27017`, `redis:6379`.
- `127.0.0.1:HOST:CONTAINER` binding database ko only local machine par expose karti hai.
- Port conflict ho to host side change karo, e.g. `127.0.0.1:27018:27017`, and backend URI bhi `27018` karo.

### Application Docker gaps

Production-ready containerization ke liye abhi ye artifacts missing hain:

- multi-stage frontend Dockerfile and static web server config;
- Session Service Dockerfile;
- API Gateway executable/image;
- admin auth integration;
- secrets handling;
- readiness checks that verify dependencies;
- complete Compose/Kubernetes service definitions and non-root users.

## 9. Local Development Setup

### Step 1: Clone

```bash
git clone <repository-url> Ecommerce
cd Ecommerce
```

### Step 2: Verify tools

```bash
git --version
node --version
npm --version
```

For backend work also:

```bash
go version
docker --version
docker compose version
```

### Step 3: Install frontend dependencies

```bash
cd frontend/session-analytics-dashboard
npm install
```

### Step 4: Configure frontend

```bash
cp .env.example .env.local
```

Windows PowerShell:

```powershell
Copy-Item .env.example .env.local
```

Keep the local proxy target at intended Gateway `http://localhost:8080`. UI-only mode me API error state expected hai.

### Step 5: Validate frontend

```bash
npm run typecheck
npm run test
npm run build
```

### Step 6: Start frontend

```bash
npm run dev
```

Open `http://localhost:5174`. Vite `strictPort: false` use karta hai, so occupied port par next free port terminal output me show hoga.

### Step 7: Optional current backend prerequisites

Repository root se MongoDB and Redis start karo using Docker run or Compose example. Then:

```bash
cd backend/services/session-service
GOWORK=off go mod download
set -a
source .env
set +a
GOWORK=off go run ./cmd/server
```

Verify process health:

```bash
curl -i http://localhost:8086/healthz
```

Expected body:

```json
{"status":"ok"}
```

This health endpoint live analytics readiness prove nahi karta. Current server Mongo/Redis unavailable hone par startup fail karta hai, but `/healthz` itself dependency details report nahi karta.

### Step 8: Optional existing migration

```bash
cd backend/services/session-service
mongosh mongodb://localhost:27017/session_db migrations/001_privacy_controls.mongodb.js
```

Task 1 overview ke liye is migration ko live analytics migration assume mat karo.

### Step 9: API verification and expected blocker

Intended request:

```bash
curl -i \
  'http://localhost:8080/api/v1/analytics/live?from=2026-06-12&to=2026-06-19&device_type=all&channel=all&source=all&user_type=all'
```

Current checkout me connection failure expected hai because Gateway implementation absent hai. Even current Session Service port target karne par live route missing hai:

```bash
curl -i \
  'http://localhost:8086/api/v1/analytics/live?from=2026-06-12&to=2026-06-19&device_type=all&channel=all&source=all&user_type=all'
# Expected current behavior: 404
```

Admin-protected final endpoint ko valid admin session cookie/JWT policy ke through call karna hoga. Credentials source code ya curl history me hardcode mat karo.

## 10. Running the Project

### Frontend daily workflow

Terminal 1:

```bash
cd frontend/session-analytics-dashboard
npm run dev
```

Terminal 2:

```bash
cd frontend/session-analytics-dashboard
npm run test:watch
```

Before commit:

```bash
npm run typecheck
npm run test
npm run build
```

### Production-build preview

```bash
npm run build
npm run preview
```

Open `http://localhost:4174`. Preview server production hosting replacement nahi hai; deploy me static web server/CDN, TLS, cache headers, SPA fallback, and API routing configure karna hoga.

### Ports and networking

| Service | Default port | Purpose | Change method |
|---|---:|---|---|
| Vite dev server | `5174` | Frontend development | `npm run dev -- --port 5175` |
| Vite preview | `4174` | Local bundle preview | `npm run preview -- --port 4175` |
| Intended API Gateway | `8080` | Browser API entry point | Gateway `HTTP_ADDR` + `VITE_API_PROXY_TARGET` |
| Current Session HTTP | `8086` | Current privacy API and health | `SESSION_HTTP_ADDR=:8087` |
| Intended Session gRPC | `50060` | Gateway downstream contract | Missing implementation/configuration |
| Auth/JWKS intent | `8081` | Admin token/JWKS | Missing runnable integration |
| MongoDB | `27017` | Session documents/aggregates | Docker host mapping + Mongo URI |
| Redis | `6379` | Active state/cache | Docker host mapping + Redis address |

Port check examples:

```bash
# Linux
ss -ltnp | grep -E ':(5174|8080|8086|27017|6379)\b'

# macOS
lsof -nP -iTCP -sTCP:LISTEN
```

Windows PowerShell:

```powershell
Get-NetTCPConnection -State Listen | Where-Object LocalPort -in 5174,8080,8086,27017,6379
```

Firewall rule usually sirf localhost development ke liye needed nahi. Databases ko public firewall me open mat karo. Remote/container deployments me allowed origins, cookie `Secure`/`SameSite`, TLS, and network policies coordinate karo.

## 11. Common Errors & Fixes

| Error/symptom | Likely cause | Fix and prevention |
|---|---|---|
| `npm: command not found` | Node/npm absent or PATH stale | Node 22 install karo, terminal reopen, versions verify karo |
| npm says WSL 1 unsupported | Windows Node and WSL 1 mix | WSL 2 enable karo; Node same environment me install karo |
| `Cannot find module ...` | `node_modules` absent/incomplete | App directory me `npm install`; correct working directory verify karo |
| `npm ci` lockfile error | Lockfile absent | First reviewed `npm install` se lock generate/commit karo; then CI uses `npm ci` |
| `ERESOLVE` | Dependency/peer mismatch | Node version and package error inspect; arbitrary `--force` avoid karo |
| Vite port already in use | Another process uses `5174` | Terminal-selected fallback URL use karo or `--port` set karo |
| Dashboard shows API error | Gateway/live endpoint absent or down | Browser Network tab and proxy target check; current repo me full endpoint missing hai |
| Browser request timeout | Backend slow/unreachable | target/logs check; only then `VITE_REQUEST_TIMEOUT_MS` adjust karo |
| CORS/credential error | Direct cross-origin base URL + missing cookie policy | Local same-origin proxy use karo or backend CORS credentials/origin correctly configure karo |
| `401 Unauthorized` | Admin session missing/expired | Admin login/auth wiring complete karo; cookie/JWT issuer/audience verify karo |
| `403 Forbidden` | Valid user lacks admin role | RBAC assignment and backend authorization verify karo |
| Mongo `connection refused` | Mongo not running/wrong port | container/service start, logs, `mongosh` ping, URI verify |
| Mongo authentication failed | Wrong user/password/authSource | credentials and URI encoding verify; secret rotate if exposed |
| Redis `connection refused` | Redis down/wrong address | `redis-cli ping`, Docker logs, address/port verify |
| Redis `NOAUTH` | Server needs password | matching `SESSION_REDIS_PASSWORD` securely set karo |
| Session config says pepper required | Secret absent because `.env` auto-loaded nahi | env export/source karo; unique long value use karo |
| `address already in use` backend | `8086` occupied | process stop or `SESSION_HTTP_ADDR=:8087` |
| Docker daemon not running | Desktop/Engine stopped or permissions | Docker start; Linux group/socket access verify |
| Docker container exits | Invalid config, port, or volume permissions | `docker ps -a` and `docker logs <name>` inspect karo |
| Docker app cannot reach `localhost` DB | `localhost` points to same container | Compose service hostname `mongo`/`redis` use karo |
| Migration command fails | `mongosh` absent, DB unreachable, permissions | CLI install/use `docker exec`; URI/role verify; backup first |
| Go workspace module error | `go.work` missing auth module reference | Session work ke liye `GOWORK=off` |
| Go version mismatch | Runtime older than `1.26.3` | Correct toolchain install/select karo |
| `go mod download` proxy failure | DNS/proxy/firewall/cache | `go env GOPROXY`, network, corporate CA/proxy verify |
| `permission denied` | File/port/Docker socket ownership | Ownership and least-required permissions fix; broad `chmod 777` avoid karo |

### Debug order

1. Browser console and Network tab me exact URL/status dekho.
2. `VITE_API_PROXY_TARGET` and terminal-selected Vite port verify karo.
3. Gateway process existence and logs check karo.
4. Auth cookie/role and response request ID inspect karo.
5. Session endpoint implementation/routes verify karo.
6. Mongo/Redis health and service logs check karo.

## 12. Security & Best Practices

- `.env` aur `.env.local` commit mat karo; only sanitized `.env.example` commit karo.
- Real-looking local passwords/peppers bhi reusable credentials nahi hone chahiye. Existing ignored auth `.env` secrets ko sample documentation me duplicate mat karo; exposed/reused hon to rotate karo.
- `VITE_` variables public hain. DB passwords, JWT private keys, Redis secrets, peppers wahan kabhi mat rakho.
- Production MongoDB/Redis ko public ports par expose mat karo; private network, authentication, TLS, and firewall use karo.
- Admin analytics endpoint server-side role check kare. UI navigation hide karna authorization nahi hai.
- Cookie auth ke saath CSRF protection, explicit trusted origins, `Secure`, `HttpOnly`, and appropriate `SameSite` policy define karo.
- Raw IP, passwords, OTP, card data, private form text, and unbounded user-agent data analytics me capture mat karo.
- Logs me cookies, authorization headers, connection strings, or identifiers leak mat karo. Request IDs useful hain.
- Mongo TTL/index migrations review and backup policy define karo. Named volume backup ke bina production durability nahi milti.
- Dependency lockfile commit karo and automated vulnerability/dependency updates enable karo.
- Dev and production configuration separate rakho; insecure local defaults production me fail-closed hon.
- Readiness endpoint dependencies and migrations check kare; liveness endpoint lightweight rakho.
- Base image versions pin karo, containers non-root run karo, and images scan karo when Dockerfiles are added.

## 13. Missing or Misconfigured Things

### Audit findings

| Severity | Finding | Professional fix |
|---|---|---|
| Blocker | Contracted `GET /api/v1/analytics/live` is not implemented by current Session Service | Handler/use case/repository plus contract tests implement karo |
| Blocker | Runnable API Gateway source is absent | Gateway executable, route mapping, auth middleware, and health checks add karo |
| Blocker | Admin auth/session integration is absent; auth service module referenced by workspace is missing | Runnable auth module restore/implement karo and admin login/RBAC flow document karo |
| High | Session protocol mismatch: intended Gateway uses gRPC `:50060`, current service exposes HTTP `:8086` only | One explicit transport contract implement karo; config/docs align karo |
| High | Live analytics storage schema, seed/event ingestion, indexes, and migrations absent | Versioned Mongo migrations plus realistic dev seed/ingestion path add karo |
| High | No Dockerfiles or checked-in Compose stack | Reproducible app and local-infra definitions add/test karo |
| Medium | Frontend dependency lockfile absent | One package manager choose karo; lockfile generate, review, commit karo |
| Medium | `backend/go.work` nonexistent auth module reference se broken hai | Missing module restore or workspace entry remove in correct implementation task |
| Medium | Backend has ignored `.env` but no sanitized `.env.example` | Loaded variables only containing safe example add karo |
| Medium | Current backend `.env` lists many future variables that code never reads | Implement loader/validation or remove stale names to prevent false confidence |
| Medium | `/healthz` shallow process check hai | Separate `/livez` and dependency-aware `/readyz` add karo |
| Medium | Vite `strictPort: false` silently next port choose karta hai | Docs/automation actual URL consume kare, or CI me strict port use karo |
| Medium | No explicit frontend runtime validation for malformed API URL | Startup/config validation and clear user-facing diagnostics add karo |
| Security | Ignored auth config contains real-looking static credentials and placeholder peppers | Never promote them to examples; rotate any reused values; use generated local secrets/secret manager |
| Security | Cookie requests require unimplemented CORS/CSRF/session policy | Gateway-level secure cookie and CSRF design implement/test karo |
| Operational | No monitoring, backup, migration runner, or analytics worker implementation | Metrics/logging, backup restore drills, worker health, and runbooks add karo |

### Definition of “full setup complete”

Live dashboard ready tab maana jayega jab:

1. Admin can authenticate and receive the approved browser credential.
2. Gateway `:8080` health/readiness passes and proxies the route.
3. Session Service serves the exact API contract or intended gRPC method.
4. MongoDB contains indexed analytics data and Redis contains live state.
5. Browser request returns schema-compatible metrics.
6. Integration tests cover `200`, `401`, `403`, empty data, timeout, and upstream failure.

## 14. Final Checklist

### Frontend shell

- [ ] Git installed
- [ ] Node.js 22 and npm available in the same environment
- [ ] Repository cloned
- [ ] `frontend/session-analytics-dashboard` selected as working directory
- [ ] `npm install` completed
- [ ] `.env.local` created from `.env.example`
- [ ] No secrets placed in `VITE_` variables
- [ ] `npm run typecheck` passes
- [ ] `npm run test` passes
- [ ] `npm run build` passes
- [ ] Dev server URL from terminal opens
- [ ] API error state is understood as expected in current checkout

### Optional current backend

- [ ] Go `1.26.3` installed
- [ ] `GOWORK=off` used for Session Service commands
- [ ] Docker/Compose or native MongoDB/Redis installed
- [ ] MongoDB ping succeeds
- [ ] Redis returns `PONG`
- [ ] Backend `.env` uses local/generated secrets
- [ ] Backend env exported into process
- [ ] `go test ./...` passes
- [ ] Session Service starts on `:8086`
- [ ] `/healthz` returns `200`
- [ ] Existing privacy migration applied only if that feature is needed

### Full analytics readiness

- [ ] API Gateway implementation exists and starts on intended port
- [ ] Admin authentication and RBAC work
- [ ] Live metrics endpoint/method is implemented
- [ ] Gateway and Session transport protocols match
- [ ] Analytics Mongo schema/index migrations are available
- [ ] Event ingestion or seed data path exists
- [ ] Redis live-session data path exists
- [ ] API returns the frontend’s expected schema
- [ ] CORS, cookies, CSRF, TLS, and secrets are production-safe
- [ ] Dependency-aware readiness checks pass
- [ ] Logs and request IDs reviewed
- [ ] Backup, monitoring, and deployment runbooks exist

---

**Beginner mental model:** Frontend ko independently build aur test kiya ja sakta hai. MongoDB/Redis start karna current Go process ko run karne me help karta hai, but databases alone missing live API ko create nahi karte. Real cards tab populate honge jab Gateway, admin auth, live-metrics backend, and analytics data pipeline sab aligned aur running hon.
