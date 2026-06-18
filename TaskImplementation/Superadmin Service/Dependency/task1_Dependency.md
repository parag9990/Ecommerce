# Project Dependency & Setup Guide

> Beginner-friendly onboarding handbook for the current `superadmin-service` implementation.

## 1. Project Overview

This service platform admins ke liye protected control plane provide karti hai. Current code admin RBAC, user/seller controls, order/payment views, refund review, session visibility authorization, platform settings, review tasks, aur audit logs handle karta hai.

### Important scope note

The source task document originally defined a **documentation-only domain task**. Repository me ab uske baad ka runnable Go implementation bhi present hai. Is guide me current runnable service ko setup karne ki requirements explain ki gayi hain.

### Service location

```text
backend/services/superadmin-service/
```

### Available setup modes

| Mode | MySQL | Downstream services | Kya work karega |
|---|---|---|---|
| Documentation viewing | Not required | Not required | Markdown and Mermaid documentation |
| Health-only local smoke test | Optional | Optional | Server and `/healthz`; protected APIs deny access and `/readyz` currently panics without a configured DB |
| Functional local setup | Required | Optional | RBAC, settings, audit logs, and local database-backed features |
| Fully integrated setup | Required | User, Order, and Payment services required | User/seller/order/payment/refund workflows |

> **Recommended for beginners:** Pehle MySQL ke saath functional local setup complete karo. Downstream services baad me connect karo.

### Runtime architecture

```text
Admin Client / API Gateway
          |
          | HTTP + trusted admin context headers
          v
superadmin-service :8088
          |
          +--> MySQL :3306
          +--> User Service admin HTTP API
          +--> Order Service admin HTTP API
          +--> Payment Service admin HTTP API
          +--> Structured logs for platform-setting events
```

## 2. Tech Stack

| Technology | Required? | Project use | Simple explanation |
|---|---|---|---|
| Go `1.26.3` | Required | Service runtime and backend code | Go ek compiled backend language hai. Isse fast aur single-binary services ban sakti hain. |
| Go standard library `net/http` | Required, included with Go | HTTP server and downstream HTTP clients | Is project me external web framework nahi hai; standard Go HTTP package use hota hai. |
| Go modules | Required, included with Go | Dependency version management | `go.mod` aur `go.sum` dependencies ko track aur verify karte hain. |
| Go workspace | Required for repository workflow | Connects repository workspace to the service module | `backend/go.work` Go ko batata hai ki workspace me kaunsa module use karna hai. |
| `database/sql` | Required when MySQL is configured | DB connection pooling and SQL execution | Ye Go ka standard database interface hai. |
| `github.com/go-sql-driver/mysql` `v1.10.0` | Required | MySQL driver | Go ke SQL calls ko MySQL server se connect karta hai. |
| MySQL 8.x | Required for functional and production use | RBAC, review tasks, platform settings, and audit logs | MySQL ek relational database hai jisme structured tables aur relationships store hote hain. |
| JSON | Required, built in | HTTP payloads, settings values, metadata, and audit snapshots | Services data ko JSON format me exchange karti hain. |
| Docker | Optional but recommended | Easy local MySQL setup and future containerization | Docker isolated containers me dependencies run karne deta hai. |
| Mermaid/Markdown | Optional runtime tool | Task and architecture documentation | Docs aur diagrams dekhne ke liye use hota hai; service run karne ke liye required nahi hai. |

### Technologies not currently used by this service

| Technology | Current status |
|---|---|
| Redis | No Redis client or Redis-backed cache exists. Platform settings cache process memory me hai. |
| Kafka / RabbitMQ / NATS | No broker client exists. Setting-update events currently sirf structured logs me publish hote hain. |
| gRPC | Design docs gRPC mention karte hain, but current implementation downstream HTTP clients use karti hai. |
| MongoDB | Current service me use nahi hota. |
| Elasticsearch / Typesense | Current service directly use nahi karti. |
| SMTP / Stripe / Twilio / Firebase / S3 | Current service me direct integration nahi hai. |
| Kubernetes | Deployment manifests currently present nahi hain. |

## 3. Required Software

| Software | Recommended version | Required? | Verify command |
|---|---:|---|---|
| Git | Recent stable | Required to clone | `git --version` |
| Go | Exactly `1.26.3` to match module files | Required | `go version` |
| MySQL Server | MySQL `8.0+` or `8.4 LTS` | Required for functional setup | `mysql --version` |
| MySQL client | Same major version as server | Required for manual migrations | `mysql --version` |
| curl | Recent stable | Recommended for health/API checks | `curl --version` |
| Docker Engine / Docker Desktop | Recent stable | Optional, recommended for MySQL | `docker version` |
| Docker Compose plugin | Compose v2 | Optional | `docker compose version` |

### Go installation notes

- **Windows:** Official Go installer use karo, then new PowerShell window open karke `go version` run karo.
- **Linux:** Official Go binary/archive recommended hai because distro package older ho sakta hai.
- **macOS:** Official installer or `brew install go` use kar sakte ho.

Expected version:

```text
go version go1.26.3 ...
```

> `go.mod` and `backend/go.work` both declare Go `1.26.3`. Older Go version module/workspace parsing ya build failure cause kar sakta hai.

## 4. Dependency Management

### Important files

| File | Purpose |
|---|---|
| `backend/go.work` | Repository workspace file; current service module ko include karta hai. |
| `backend/go.work.sum` | Workspace dependency checksums. |
| `backend/services/superadmin-service/go.mod` | Module name, Go version, and direct/indirect dependencies. |
| `backend/services/superadmin-service/go.sum` | Downloaded module integrity checksums. |

### Direct and indirect dependencies

```text
Direct:
github.com/go-sql-driver/mysql v1.10.0

Indirect:
filippo.io/edwards25519 v1.2.0
```

`go.sum` ko manually edit nahi karna chahiye. Go commands is file ko automatically maintain karte hain.

### Common Go commands

Run these commands from the service directory:

```bash
cd backend/services/superadmin-service

go mod download
go test ./...
go build -o /tmp/superadmin-service ./cmd/server
go run ./cmd/server
```

Dependency cleanup:

```bash
go mod tidy
```

Use `go mod tidy` tabhi run karo jab imports/dependencies change hue hon. Is command ke baad `go.mod` aur `go.sum` diff review karo.

### Dependency troubleshooting

| Problem | Likely cause | Fix |
|---|---|---|
| `go: go.mod requires go >= 1.26.3` | Installed Go old hai | Go `1.26.3` install/update karo. |
| Module download timeout | Proxy/network issue | `go env GOPROXY`; then try `GOPROXY=https://proxy.golang.org,direct go mod download`. |
| Private/proxy restriction | Corporate proxy blocks Go proxy | Approved proxy configure karo or `GOPROXY=direct` try karo. |
| Checksum mismatch | Corrupt cache or dependency tampering | Dependency source verify karo, then `go clean -modcache` and download again. |
| Workspace confusion | Command wrong directory se run hua | Service directory me run karo, or `GOWORK=off` use karke module-only behavior test karo. |
| `missing go.sum entry` | Dependencies incomplete | `go mod download` or `go mod tidy` run karo. |

## 5. Database Setup

### What database is used?

MySQL is the only database currently used by this service.

MySQL ek relational database hai. Is project me admin permissions, review tasks, platform settings, aur immutable-style audit records structured tables me store hote hain.

### Is MySQL mandatory?

- **Health-only local mode:** Optional. Server starts without a DSN and `/healthz` works, but protected admin endpoints deny access, DB-backed features remain unavailable, and current `/readyz` has a typed-nil panic bug.
- **Functional local mode:** Required.
- **Non-local environments:** Required by configuration validation.

### Database tables

| Table | Purpose |
|---|---|
| `admin_users` | Admin identity, role, status, and MFA requirement |
| `admin_permissions` | Known permission catalog |
| `admin_role_permissions` | Role-to-permission mapping |
| `admin_review_tasks` | Manual review workflow records |
| `platform_settings` | Versioned platform settings |
| `admin_audit_logs` | Sensitive admin action audit records |

> **Source of truth:** Use files inside `backend/services/superadmin-service/migrations/`. The repository-level `database/draw.sql` contains an older design snapshot and does not match all current columns and constraints.

### Default port and connection string

| Setting | Value |
|---|---|
| Default MySQL port | `3306` |
| Local database name | `superadmin_db` |
| Driver DSN format | `user:password@tcp(host:port)/database?parseTime=true&charset=utf8mb4&loc=UTC` |

Recommended local DSN:

```env
SUPERADMIN_DATABASE_DSN='superadmin:devpassword@tcp(127.0.0.1:3306)/superadmin_db?parseTime=true&charset=utf8mb4&loc=UTC'
```

`parseTime=true` important hai because repositories MySQL timestamps ko Go `time.Time` values me scan karte hain.
Single quotes Bash/Zsh `source .env` ke time parentheses and `&` ko shell syntax banne se rokti hain.

### Local installation

#### Windows

1. MySQL Installer or package manager se MySQL Server install karo.
2. Setup ke time root password securely note karo.
3. Windows Services me MySQL service start karo.
4. Verify:

```powershell
mysql --version
mysql -u root -p
```

Common service name `MySQL80` ho sakta hai:

```powershell
net start MySQL80
```

#### Ubuntu / Debian Linux

```bash
sudo apt update
sudo apt install mysql-server mysql-client
sudo systemctl enable --now mysql
sudo systemctl status mysql
sudo mysql
```

#### macOS with Homebrew

```bash
brew install mysql
brew services start mysql
mysql -u root
```

### Recommended Docker setup

Docker beginner setup host installation se simpler hota hai:

```bash
docker volume create superadmin_mysql_data

docker run -d \
  --name superadmin-mysql \
  --restart unless-stopped \
  -e MYSQL_ROOT_PASSWORD=rootpassword \
  -e MYSQL_DATABASE=superadmin_db \
  -e MYSQL_USER=superadmin \
  -e MYSQL_PASSWORD=devpassword \
  -p 3306:3306 \
  -v superadmin_mysql_data:/var/lib/mysql \
  mysql:8.4
```

Check startup:

```bash
docker ps
docker logs superadmin-mysql
docker exec -it superadmin-mysql mysql -u superadmin -pdevpassword superadmin_db
```

Stop/start later:

```bash
docker stop superadmin-mysql
docker start superadmin-mysql
```

The named volume keeps data even when the container stops. Do not remove the volume unless local database data can be deleted.

### Docker Compose example for MySQL

No repository Compose file currently exists. This is a working reference for a future local Compose file:

```yaml
services:
  superadmin-mysql:
    image: mysql:8.4
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: superadmin_db
      MYSQL_USER: superadmin
      MYSQL_PASSWORD: devpassword
    ports:
      - "3306:3306"
    volumes:
      - superadmin_mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD-SHELL", "mysqladmin ping -h 127.0.0.1 -u root -prootpassword"]
      interval: 5s
      timeout: 5s
      retries: 20

volumes:
  superadmin_mysql_data:
```

### Create database and least-privilege local user

Login as root/admin:

```bash
mysql -u root -p
```

Run:

```sql
CREATE DATABASE IF NOT EXISTS superadmin_db
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

CREATE USER IF NOT EXISTS 'superadmin'@'%' IDENTIFIED BY 'devpassword';
GRANT ALL PRIVILEGES ON superadmin_db.* TO 'superadmin'@'%';
FLUSH PRIVILEGES;
```

For production, `%` host access and broad privileges ko replace karke private host/network aur minimum required privileges use karo.

### Run migrations

There is currently no configured migration CLI or schema-version table. Apply the `.up.sql` files exactly once and in filename order:

```bash
cd backend/services/superadmin-service

for migration in migrations/*.up.sql; do
  echo "Applying $migration"
  mysql -h 127.0.0.1 -P 3306 -u superadmin -p superadmin_db < "$migration"
done
```

Migration order:

```text
001_create_superadmin_rbac.up.sql
002_seed_admin_rbac.up.sql
003_create_admin_review_tasks.up.sql
004_add_open_review_task_uniqueness.up.sql
005_create_platform_settings.up.sql
006_create_admin_audit_logs.up.sql
```

> Migration `004` is not safely repeatable. Re-running it after success will fail because the generated column/index already exists. A real migration runner should track applied versions.

### Seed a local admin

Permission migration roles/permissions seed karti hai, but `admin_users` me local admin create nahi karti. Protected APIs test karne ke liye local-only admin row add karo:

```bash
mysql -h 127.0.0.1 -P 3306 -u superadmin -p superadmin_db
```

```sql
INSERT INTO admin_users
  (admin_id, user_id, role, status, mfa_required, created_by)
VALUES
  ('admin_local', 'user_local', 'superadmin', 'active', TRUE, 'local_setup')
ON DUPLICATE KEY UPDATE
  role = VALUES(role),
  status = 'active',
  mfa_required = TRUE;
```

Production me admin bootstrap audited and authenticated workflow se hona chahiye. Manual SQL production ke liye recommended nahi hai.

### Verify database

```bash
mysqladmin -h 127.0.0.1 -P 3306 -u superadmin -p ping
mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e "SHOW TABLES;"
mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e "SELECT admin_id, role, status FROM admin_users;"
```

### Rollback warning

Down migrations destructive hain. They tables/data remove kar sakti hain. Rollback only disposable local DB ya approved backup/process ke saath run karo, and reverse order use karo:

```text
006 -> 005 -> 004 -> 003 -> 002 -> 001
```

## 6. External Services

### API Gateway

| Detail | Status |
|---|---|
| Required locally? | Optional for direct smoke testing |
| Required in production? | Yes |
| Why? | Authentication, trusted admin context, routing, rate limits, and public network protection |

Current server JWT verify nahi karta. It trusts headers such as `X-Admin-Id`, `X-User-Id`, `X-Admin-Roles`, `X-Request-Id`, and `X-MFA-Verified`. Therefore production me port `8088` ko public internet par directly expose nahi karna chahiye.

### User Service

| Detail | Value |
|---|---|
| Config variable | `USER_SERVICE_ADMIN_BASE_URL` |
| Default admin path when URL has no path | `/internal/admin` |
| Required locally? | Optional |
| Required outside local/test? | Yes |
| Purpose | User/seller listing, profiles, and status updates |

Example:

```env
USER_SERVICE_ADMIN_BASE_URL=http://127.0.0.1:8081/internal/admin
```

Health check depends on that service's implementation. At minimum verify network reachability:

```bash
curl -i http://127.0.0.1:8081/healthz
```

### Order Service

| Detail | Value |
|---|---|
| Config variable | `ORDER_SERVICE_ADMIN_BASE_URL` |
| Default admin path when URL has no path | `/internal/admin` |
| Required locally? | Optional |
| Required outside local/test? | Yes |
| Purpose | Order lists, details, history, disputes, and manual review context |

### Payment Service

| Detail | Value |
|---|---|
| Config variable | `PAYMENT_SERVICE_ADMIN_BASE_URL` |
| Default admin path when URL has no path | `/internal/admin` |
| Required locally? | Optional |
| Required outside local/test? | Yes |
| Purpose | Payment lists, refund lookup, and refund review |

### Platform-setting event publisher

`SUPERADMIN_PLATFORM_SETTINGS_EVENT_TOPIC` event label configure karta hai. Current publisher Kafka/RabbitMQ ko connect nahi karta; it writes the event to structured application logs.

```env
SUPERADMIN_PLATFORM_SETTINGS_EVENT_TOPIC=platform.settings.updated
```

### Redis, queue, and session-service clarification

- Redis setup is **not required** for the current implementation.
- Kafka/RabbitMQ setup is **not required** for the current implementation.
- Session visibility endpoints currently authorization/config checks perform karte hain; no Session Service base URL/client is configured.
- Platform settings cache local process memory me hai, distributed cache nahi.

## 7. Environment Variables

### Where to create `.env`

Recommended local location:

```text
backend/services/superadmin-service/.env
```

The repository `.gitignore` correctly ignores `.env` files. Secrets ko Git me commit mat karo.

### Critical loading behavior

The application **does not automatically load `.env` files**. No dotenv library is used. Go process only actual operating-system environment variables reads karta hai.

Bash/Zsh:

```bash
cd backend/services/superadmin-service
set -a
source .env
set +a
go run ./cmd/server
```

PowerShell example:

```powershell
$env:APP_ENV = "local"
$env:HTTP_ADDR = ":8088"
$env:SUPERADMIN_DATABASE_DSN = "superadmin:devpassword@tcp(127.0.0.1:3306)/superadmin_db?parseTime=true&charset=utf8mb4&loc=UTC"
go run ./cmd/server
```

### Complete `.env` example

```env
# Service identity and HTTP
SERVICE_NAME=superadmin-service
APP_ENV=local
HTTP_ADDR=:8088
LOG_LEVEL=info

# MySQL
SUPERADMIN_DATABASE_DSN='superadmin:devpassword@tcp(127.0.0.1:3306)/superadmin_db?parseTime=true&charset=utf8mb4&loc=UTC'
SUPERADMIN_DB_MAX_OPEN_CONNS=25
SUPERADMIN_DB_MAX_IDLE_CONNS=25
SUPERADMIN_DB_CONN_MAX_LIFETIME=5m
SUPERADMIN_DB_PING_TIMEOUT=5s
SUPERADMIN_REQUIRE_DATABASE=true

# User Service
USER_SERVICE_ADMIN_BASE_URL=http://127.0.0.1:8081/internal/admin
USER_SERVICE_TIMEOUT=5s
SUPERADMIN_REQUIRE_USER_SERVICE=false

# Order Service
ORDER_SERVICE_ADMIN_BASE_URL=http://127.0.0.1:8082/internal/admin
ORDER_SERVICE_TIMEOUT=5s
SUPERADMIN_REQUIRE_ORDER_SERVICE=false

# Payment Service
PAYMENT_SERVICE_ADMIN_BASE_URL=http://127.0.0.1:8083/internal/admin
PAYMENT_SERVICE_TIMEOUT=5s
SUPERADMIN_REQUIRE_PAYMENT_SERVICE=false

# Settings and session visibility
SUPERADMIN_PLATFORM_SETTINGS_CACHE_TTL=5m
SUPERADMIN_PLATFORM_SETTINGS_EVENT_TOPIC=platform.settings.updated
SUPERADMIN_SESSION_ANALYTICS_MAX_RANGE=720h
SUPERADMIN_SESSION_ANALYTICS_MAX_PAGE_SIZE=100

# HTTP safety and shutdown
HTTP_READ_HEADER_TIMEOUT=5s
HTTP_SHUTDOWN_TIMEOUT=10s
```

If User/Order/Payment services are not running, remove their base URLs or keep them empty. Their related endpoints will return downstream-unavailable responses.

### Variable reference

| Variable | Default | Required? | Purpose / security note |
|---|---|---|---|
| `SERVICE_NAME` | `superadmin-service` | Yes | Log/service identity. Empty value fails validation. |
| `APP_ENV` | `local` | Yes in practice | `local`/`test` allow optional dependencies; other values require DB and all downstream services. |
| `HTTP_ADDR` | `:8088` | Yes | Listen address. `:8088` binds all interfaces; use `127.0.0.1:8088` for local-only exposure. |
| `LOG_LEVEL` | `info` | Optional | Primary log level variable. |
| `SUPERADMIN_LOG_LEVEL` | `info` | Optional fallback | Used only if `LOG_LEVEL` is empty. Prefer `LOG_LEVEL`. |
| `SUPERADMIN_DATABASE_DSN` | Empty | Functional setup: yes | Preferred MySQL DSN. Contains credentials; treat as secret. |
| `MYSQL_DSN` | Empty | Optional fallback | Used only if preferred DSN is empty. |
| `SUPERADMIN_DB_MAX_OPEN_CONNS` | `25` | Optional | Maximum open DB connections; must be greater than zero. |
| `SUPERADMIN_DB_MAX_IDLE_CONNS` | `25` | Optional | Idle pool size; cannot exceed max open connections. |
| `SUPERADMIN_DB_CONN_MAX_LIFETIME` | `5m` | Optional | Recycles old DB connections. |
| `SUPERADMIN_DB_PING_TIMEOUT` | `5s` | Optional | Startup DB ping deadline. |
| `SUPERADMIN_REQUIRE_DATABASE` | `false` | Recommended `true` locally | Forces DB requirement even in local/test. Non-local env already requires DB. |
| `USER_SERVICE_ADMIN_BASE_URL` | Empty | Integrated/non-local: yes | User Service admin HTTP URL. Use HTTPS or trusted private network in production. |
| `USER_SERVICE_TIMEOUT` | `5s` | Optional | User Service request timeout. |
| `SUPERADMIN_REQUIRE_USER_SERVICE` | `false` | Optional local flag | Forces URL requirement in local/test. |
| `ORDER_SERVICE_ADMIN_BASE_URL` | Empty | Integrated/non-local: yes | Order Service admin HTTP URL. |
| `ORDER_SERVICE_TIMEOUT` | `5s` | Optional | Order Service request timeout. |
| `SUPERADMIN_REQUIRE_ORDER_SERVICE` | `false` | Optional local flag | Forces URL requirement in local/test. |
| `PAYMENT_SERVICE_ADMIN_BASE_URL` | Empty | Integrated/non-local: yes | Payment Service admin HTTP URL. |
| `PAYMENT_SERVICE_TIMEOUT` | `5s` | Optional | Payment Service request timeout. |
| `SUPERADMIN_REQUIRE_PAYMENT_SERVICE` | `false` | Optional local flag | Forces URL requirement in local/test. |
| `SUPERADMIN_PLATFORM_SETTINGS_CACHE_TTL` | `5m` | Optional | In-process settings cache duration; `0s` is allowed. |
| `SUPERADMIN_PLATFORM_SETTINGS_EVENT_TOPIC` | `platform.settings.updated` | Yes | Logging event topic label; currently not a broker topic connection. |
| `SUPERADMIN_SESSION_ANALYTICS_MAX_RANGE` | `720h` | Optional | Maximum allowed analytics date range. Go duration format required. |
| `SUPERADMIN_SESSION_ANALYTICS_MAX_PAGE_SIZE` | `100` | Optional | Maximum session analytics page size; must be greater than zero. |
| `HTTP_READ_HEADER_TIMEOUT` | `5s` | Optional | Protects server from slow request headers. |
| `HTTP_SHUTDOWN_TIMEOUT` | `10s` | Optional | Graceful shutdown deadline. |

### Common environment mistakes

- `.env` create kiya but shell me load nahi kiya.
- DSN me special password characters URL/DSN-safe format me escape nahi kiye.
- `localhost` container ke andar host machine ko point nahi karta.
- Duration me plain `5` diya instead of `5s` or `5m`.
- Invalid integer/boolean values silently default par fall back ho sakti hain.
- `APP_ENV=production` set kiya but required downstream URLs missing hain.
- Secret values logs, screenshots, shell history, or Git me expose kiye.

## 8. Ports & Networking

| Service | Default/example port | Purpose | How to change |
|---|---:|---|---|
| Current backend HTTP server | `8088` | Main API, health, readiness | Change `HTTP_ADDR` |
| MySQL | `3306` | RBAC/settings/audit database | Change MySQL mapping and DSN |
| User Service | Example `8081` | User/seller admin HTTP API | Change `USER_SERVICE_ADMIN_BASE_URL` |
| Order Service | Example `8082` | Order admin HTTP API | Change `ORDER_SERVICE_ADMIN_BASE_URL` |
| Payment Service | Example `8083` | Payment/refund admin HTTP API | Change `PAYMENT_SERVICE_ADMIN_BASE_URL` |
| API Gateway | Not defined in current repo | Protected external entry point | Gateway-specific config required |

### Port conflict checks

Linux/macOS:

```bash
lsof -i :8088
lsof -i :3306
```

Alternative Linux command:

```bash
ss -ltnp | grep ':8088'
```

Windows PowerShell:

```powershell
Get-NetTCPConnection -LocalPort 8088
Get-NetTCPConnection -LocalPort 3306
```

Change local backend port:

```env
HTTP_ADDR=127.0.0.1:18088
```

### Docker networking notes

- Host se Docker MySQL connect karte time `127.0.0.1:3306` use karo.
- Another Compose container se connect karte time service name use karo, for example `superadmin-mysql:3306`.
- Container ke andar `localhost` same container ko refer karta hai.
- MySQL port public internet par expose mat karo.
- Production internal service traffic private network, firewall rules, TLS/mTLS, and service authentication use kare.

## 9. Docker Setup

### Current repository status

| Item | Present? | Impact |
|---|---|---|
| Service `Dockerfile` | No | Backend image cannot currently be built from a repository Dockerfile. |
| Local `docker-compose.yml` | No | Full local stack command currently unavailable. |
| Kubernetes manifests | No | Kubernetes deployment is not ready. |
| MySQL Docker setup | Documentation command available above | Local DB can still run in Docker. |

### Useful Docker commands

```bash
docker ps
docker logs -f superadmin-mysql
docker stop superadmin-mysql
docker start superadmin-mysql
docker compose up -d
docker compose down
docker compose logs -f
```

`docker compose` commands tabhi work karenge jab Compose file create ho. Current repository me referenced `infra/compose/docker-compose.local.yml` present nahi hai.

### Local setup versus Docker

| Approach | Benefit | Tradeoff |
|---|---|---|
| Go locally + MySQL in Docker | Beginner-friendly, fast rebuilds, easy DB cleanup | Host Go still required |
| Everything installed locally | No Docker dependency | OS-specific setup and cleanup harder |
| Full Compose stack | Reproducible team setup | Currently missing and needs implementation |

Recommended current approach: Go locally run karo and MySQL Docker container use karo.

## 10. Local Development Setup

### Step 1: Clone repository

```bash
git clone https://github.com/parag9990/Ecommerce.git
cd Ecommerce
```

### Step 2: Verify runtime

```bash
git --version
go version
```

Expected Go version is `1.26.3`.

### Step 3: Go to service directory

```bash
cd backend/services/superadmin-service
```

### Step 4: Download dependencies

```bash
go mod download
```

### Step 5: Start MySQL

Use either local MySQL service or the Docker command from the database section.

Verify:

```bash
mysqladmin -h 127.0.0.1 -P 3306 -u superadmin -p ping
```

### Step 6: Run migrations

```bash
for migration in migrations/*.up.sql; do
  echo "Applying $migration"
  mysql -h 127.0.0.1 -P 3306 -u superadmin -p superadmin_db < "$migration"
done
```

### Step 7: Seed a local admin

Add `admin_local` using the SQL shown in the database section.

### Step 8: Create and load `.env`

Create the ignored local `.env`, then:

```bash
set -a
source .env
set +a
```

### Step 9: Run tests and build

```bash
go test ./...
go build -o /tmp/superadmin-service ./cmd/server
```

### Step 10: Start service

```bash
go run ./cmd/server
```

Expected log includes the listen address:

```text
superadmin service listening
```

### Step 11: Verify health and readiness

New terminal:

```bash
curl -i http://127.0.0.1:8088/healthz
curl -i http://127.0.0.1:8088/readyz
```

Expected JSON:

```json
{"status":"ok"}
```

With a configured and reachable database, readiness should return:

```json
{"status":"ready"}
```

`/healthz` means process alive hai. `/readyz` configured DB ko ping karta hai.

> **Known bug:** If no DB DSN is configured, current code passes a typed nil `*sql.DB` into the readiness interface. Calling `/readyz` then panics and the client receives an empty response. Configure MySQL for readiness checks, and fix the handler construction before relying on no-DB readiness.

### Step 12: Verify protected RBAC API

```bash
curl -i \
  -H 'X-Admin-Id: admin_local' \
  -H 'X-User-Id: user_local' \
  -H 'X-Admin-Roles: superadmin' \
  -H 'X-Request-Id: req_local_001' \
  -H 'X-MFA-Verified: true' \
  http://127.0.0.1:8088/api/v1/admin/rbac/permissions
```

The database `admin_users` row, not only the `X-Admin-Roles` header, determines permissions.

### Step 13: Connect downstream services

Start User, Order, and Payment services, update their base URLs, reload environment variables, and restart this service.

### Step 14: Graceful shutdown

Press `Ctrl+C`. Server allows up to `HTTP_SHUTDOWN_TIMEOUT` for graceful shutdown.

## 11. Running the Project

### Minimal health-only run

```bash
cd backend/services/superadmin-service
APP_ENV=local HTTP_ADDR=127.0.0.1:8088 go run ./cmd/server
```

This is useful only for process/health verification. Protected APIs will deny access because no admins exist in the static empty permission repository.
Use `/healthz` only in this mode; `/readyz` currently triggers the known typed-nil DB panic.

### Functional local run

```bash
cd backend/services/superadmin-service
set -a
source .env
set +a
go run ./cmd/server
```

### Production-like configuration expectations

For any `APP_ENV` other than empty, `local`, or `test`, startup validation requires:

- `SUPERADMIN_DATABASE_DSN`
- `USER_SERVICE_ADMIN_BASE_URL`
- `ORDER_SERVICE_ADMIN_BASE_URL`
- `PAYMENT_SERVICE_ADMIN_BASE_URL`

Production must additionally provide API Gateway protection, secret management, private networking, TLS/mTLS, observability, backups, and automated migrations.

## 12. Common Errors & Fixes

| Error / symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `bind: address already in use` | Port `8088` busy hai | Process stop karo or `HTTP_ADDR` change karo | Team port convention document karo |
| `connect: connection refused` for MySQL | MySQL stopped, wrong host/port, or container not ready | MySQL start karo; `docker logs`; DSN verify karo | Healthcheck and startup wait use karo |
| MySQL `Access denied for user` | Wrong username/password/host grants | Credentials and `GRANT` verify karo | Least-privilege managed credentials use karo |
| `Unknown database 'superadmin_db'` | Database create nahi hua | Database create command run karo | Bootstrap script/Compose init add karo |
| `Table ... doesn't exist` | Migrations missing or wrong DB selected | All `.up.sql` files correct order me run karo | Migration runner add karo |
| Duplicate column/index during migration `004` | Migration already applied | Do not re-run it; inspect schema first | Versioned migration runner use karo |
| Audit insert foreign-key failure | Actor `admin_id` not in `admin_users` | Valid active admin seed/provision karo | Admin bootstrap workflow create karo |
| `sql: Scan error` around timestamps | DSN missing `parseTime=true` | Add `parseTime=true` and restart | Standard DSN template use karo |
| Protected endpoint returns `FORBIDDEN` | Admin missing/disabled or role lacks DB permission | Check `admin_users` and seeded role permissions | Provision admins through audited workflow |
| Admin context/request-context error | Required headers missing | Add admin ID, user ID, roles, and request ID headers | API Gateway must inject validated context |
| `MFA_REQUIRED` | High-risk endpoint requires MFA context | For local test add `X-MFA-Verified: true`; production must verify real MFA | Never trust client-provided MFA header publicly |
| Downstream unavailable response | Base URL empty, service down, wrong path, or timeout | Start service and verify URL/path/network | Add service health monitoring and retries where safe |
| `.env` values ignored | App does not auto-load dotenv | Export/source `.env` before `go run` | Add explicit config loader or standard run script |
| Invalid duration falls back to default | Value such as `5` instead of `5s` | Use Go duration syntax | Validate config in CI/deploy |
| Invalid integer/boolean silently uses default | Parser ignores malformed value | Correct value and restart | Change config parser to fail fast |
| Docker daemon not running | Docker Desktop/Engine stopped | Start Docker and retry | Enable startup and monitor Docker status |
| Docker MySQL cannot be reached from another container | DSN uses `localhost` | Use Compose service DNS name | Document container network names |
| `go mod download` fails | Network, proxy, cache, or version issue | Verify Go version/GOPROXY; retry; inspect cache | Pin versions and use CI dependency cache |
| Permission denied running commands/files | OS permissions or protected directory | Use correct user/permissions; avoid running project as root | Keep workspace user-owned |
| `/readyz` returns empty reply and server logs a nil-pointer panic | No-DB startup passes typed nil `*sql.DB` as a non-nil readiness interface | Configure DB now; code fix should pass a truly nil checker or guard typed nil | Add no-DB readiness test |
| `/readyz` is incomplete in integrated mode | Current readiness checks only MySQL, not downstream services | Add dependency-aware readiness policy | Define environment-specific readiness requirements |

## 13. Security & Best Practices

### Security and configuration audit

| Severity | Finding | Recommended fix |
|---|---|---|
| Critical for public deployment | Service trusts admin identity, roles, request ID, and MFA status from HTTP headers; it does not verify JWTs itself | Keep service on private network, allow only trusted gateway traffic, authenticate gateway-to-service calls, and verify signed identity context |
| High | No API Gateway implementation/config is present in this repository | Add gateway route, admin JWT validation, rate limiting, and strict header sanitization |
| High | Downstream HTTP clients have no built-in service credentials or mTLS | Add service authentication and TLS/mTLS for internal calls |
| High | No automated migration runner/version table exists | Add a migration tool and CI/CD migration stage |
| High | No production admin bootstrap/provisioning workflow exists | Add secure, audited admin provisioning with MFA enrollment |
| Medium | No service Dockerfile, Compose stack, or Kubernetes manifests exist | Add reproducible container and deployment files |
| Medium | No committed `.env.example` exists | Add a secret-free `.env.example` containing all supported keys |
| Medium | No metrics, tracing, or external alerting implementation is present | Add request metrics, DB/downstream latency, error-rate alerts, and trace propagation |
| Medium | Settings events only log; no durable broker publisher exists | Add transactional outbox plus Kafka/RabbitMQ publisher if durable events are required |
| High | `/readyz` panics when no DB is configured because of a typed-nil interface | Pass a truly nil readiness checker or add a safe nil guard; add regression test |
| Medium | Readiness checks only MySQL and does not report downstream dependency state | Make readiness policy explicit and add appropriate dependency checks |
| Medium | Invalid integer/boolean config silently falls back to defaults | Fail startup with clear validation errors |
| Medium | Design docs mention gRPC, but current clients use HTTP | Update architecture docs or implement intended gRPC contracts |
| Medium | Audit writes and downstream mutations cannot share one DB transaction | Add idempotency, reconciliation, and durable workflow/outbox patterns |
| Low | Repository-level `database/draw.sql` differs from current migrations | Mark it as conceptual/legacy or regenerate it from current schema |

### Beginner-friendly best practices

- Never commit `.env`, DSNs, passwords, API keys, certificates, or tokens.
- Commit a safe `.env.example` with placeholder values.
- Use `127.0.0.1:8088` for local-only direct runs.
- Keep MySQL and service ports private in production.
- Use a dedicated DB user; root credentials application ko mat do.
- Use strong unique passwords and a secret manager in staging/production.
- Always include `parseTime=true` in MySQL DSN.
- Take database backups and test restore process.
- Run migrations through a version-aware migration tool.
- Run `go test ./...` before commit/deploy.
- Review `go.mod` and `go.sum` changes.
- Use real MFA verification for refund/settings actions.
- Do not accept admin context headers directly from untrusted clients.
- Log request IDs, but never log secrets or sensitive raw PII.
- Add health checks, metrics, alerts, and downstream timeout dashboards.
- Use separate local, test, staging, and production configurations.
- Use Docker named volumes for local persistent data.
- Pin container image versions instead of relying on `latest`.

## 14. Missing or Misconfigured Things

These items are not blockers for a basic local Go run, but they are required for a smooth team onboarding or production deployment:

- [ ] No repository service `Dockerfile`
- [ ] No working local Docker Compose file
- [ ] No `.env.example`
- [ ] No migration runner or migration history table
- [ ] No automatic database/user bootstrap
- [ ] No secure admin provisioning workflow
- [ ] No API Gateway route/security configuration
- [ ] No JWT verification inside the service
- [ ] No service-to-service authentication or mTLS
- [ ] No Redis or distributed cache, despite broader architecture docs mentioning Redis
- [ ] No durable Kafka/RabbitMQ event publisher
- [ ] No Session Service client
- [ ] No metrics/tracing implementation
- [ ] No Kubernetes manifests
- [ ] No integration tests against real MySQL/downstream services
- [ ] No regression test for `/readyz` without a configured DB
- [ ] Current implementation/docs mismatch: HTTP clients versus planned gRPC
- [ ] Current schema migrations versus older `database/draw.sql` mismatch

## 15. Final Checklist

### Beginner local setup

- [ ] Repository cloned
- [ ] Git installed and working
- [ ] Go `1.26.3` installed
- [ ] Service directory opened
- [ ] `go mod download` completed
- [ ] MySQL installed or Docker MySQL running
- [ ] `superadmin_db` created
- [ ] Dedicated local DB user created
- [ ] All six up migrations applied exactly once
- [ ] Local admin row seeded
- [ ] `.env` created and kept out of Git
- [ ] `.env` exported into the shell
- [ ] `go test ./...` passes
- [ ] `go build -o /tmp/superadmin-service ./cmd/server` passes
- [ ] Service starts on expected port
- [ ] `/healthz` returns `200`
- [ ] `/readyz` returns `200` when MySQL is configured
- [ ] RBAC permission endpoint works for seeded admin
- [ ] Logs reviewed for warnings/errors

### Fully integrated setup

- [ ] User Service running and URL configured
- [ ] Order Service running and URL configured
- [ ] Payment Service running and URL configured
- [ ] Downstream admin paths reachable
- [ ] API Gateway protects admin routes
- [ ] Untrusted clients cannot inject admin headers
- [ ] MFA context is verified, not blindly trusted
- [ ] MySQL backups configured
- [ ] Migrations automated and version tracked
- [ ] Secrets stored outside Git
- [ ] Metrics, tracing, logs, and alerts configured
- [ ] Common errors and security audit reviewed

## Quick Command Reference

```bash
# Enter service
cd backend/services/superadmin-service

# Dependencies and verification
go mod download
go test ./...
go build -o /tmp/superadmin-service ./cmd/server

# Load environment
set -a
source .env
set +a

# Run
go run ./cmd/server

# Health checks
curl -i http://127.0.0.1:8088/healthz
curl -i http://127.0.0.1:8088/readyz

# MySQL checks
mysqladmin -h 127.0.0.1 -P 3306 -u superadmin -p ping
mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e "SHOW TABLES;"
```
