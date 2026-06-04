# Project Dependency & Setup Guide

## 1. Project Overview

Ye guide `task1.md` ke basis par dependency, setup, environment, database, external service, Docker, run, and troubleshooting documentation provide karta hai.

`task1.md` ka main scope payment state machine document karna tha: payment states, allowed transitions, invalid transitions, retry/refund behavior, and provider webhook source-of-truth rules. Original task file business/design guide hai. Ye dependency guide uske around project ko local machine par run karne ke liye required non-code cheezein explain karta hai.

Important repo paths:

| Path | Purpose |
|---|---|
| `TaskImplementation/Payment Service/task1.md` | Original implementation/task documentation. Isko modify nahi karna hai. |
| `TaskImplementation/Payment Service/task1_Dependency.md` | Ye generated dependency and setup guide. |
| `backend/services/payment-service` | Actual Go backend service code. |
| `backend/services/payment-service/go.mod` | Go dependencies ka main file. |
| `backend/services/payment-service/migrations` | MySQL database migrations. |

Simple Hinglish explanation:

Payment Service ek backend service hai jo payment lifecycle handle karta hai. Isme payment ka state machine, payment intent creation, provider webhook handling, refund, retry, schema health, and reconciliation logic available hai. Payment ka final truth provider webhook/trusted provider response se aata hai, frontend success screen se nahi.

## 2. Tech Stack

| Technology | Required? | What it is | Why this project uses it |
|---|---:|---|---|
| Go `1.26.3` | Required | Go ek compiled backend language hai. Simple English me: fast APIs and services banane ke liye use hota hai. | Payment backend service Go me likha gaya hai. `go.mod` me Go version `1.26.3` defined hai. |
| Go Modules | Required | Go ka dependency management system. | External library versions lock karne ke liye `go.mod` and `go.sum` use hote hain. |
| `net/http` | Required | Go standard library ka HTTP server package. | Service HTTP endpoints expose karta hai, jaise `/healthz` and internal payment APIs. |
| `database/sql` | Required | Go standard library ka generic SQL DB package. | MySQL se connection manage karne ke liye use hota hai. |
| `github.com/go-sql-driver/mysql` | Required | Go ka MySQL driver. | Payment Service MySQL database se connect karta hai. |
| MySQL | Required | MySQL ek relational database hai jisme tables ke form me data store hota hai. | Payments, attempts, refunds, webhooks, and reconciliation records store karne ke liye. |
| SQL migrations | Required | SQL files jo database schema create/update karte hain. | `payment_db` and required tables/indexes banane ke liye. |
| HTTP payment providers | Optional until enabled | Stripe-like and Razorpay-like provider clients. | Real payment intent, webhook, refund flows ke liye external provider API call hoti hai. |
| HTTP event publisher | Required when providers/reconciliation enabled | Event endpoint par JSON POST karne wala publisher. | Payment captured/failed/refund/reconciliation events dusre services ko bhejne ke liye. |
| Docker | Optional but recommended | Containers run karne ka tool. | Beginner local setup me MySQL quickly run karne ke liye helpful hai. |
| Mermaid | Optional docs-only | Markdown diagrams ka syntax. | Original task file me state diagrams explain karne ke liye. Runtime dependency nahi hai. |
| Shields.io | Optional docs-only | Markdown badge image service. | Original docs me badges show karne ke liye. Runtime dependency nahi hai. |

Not used in current service:

| Technology | Status |
|---|---|
| Redis | Current implementation me direct Redis dependency nahi mili. |
| Kafka | Current implementation me Kafka client nahi mila. Payment events HTTP endpoint par publish hote hain. |
| RabbitMQ | Current implementation me RabbitMQ dependency nahi mili. |
| NATS | Current implementation me NATS dependency nahi mili. |
| Elasticsearch | Current payment service me direct dependency nahi mili. |
| Kubernetes | Repo me payment service ke liye Kubernetes manifests nahi mile. |

## 3. Required Software

Install these before running the backend service:

| Software | Required? | Recommended version | Verify command |
|---|---:|---|---|
| Git | Required | Latest stable | `git --version` |
| Go | Required | `1.26.3` as per `go.mod` | `go version` |
| MySQL Server | Required | MySQL 8.x recommended | `mysql --version` |
| MySQL client | Required for migrations | Same as server/client package | `mysql --version` |
| curl | Recommended | Any recent version | `curl --version` |
| Docker Desktop / Docker Engine | Optional | Latest stable | `docker --version` |
| Docker Compose plugin | Optional | Latest stable | `docker compose version` |

Beginner note:

Go service run karne ke liye Go install hona mandatory hai. MySQL bhi mandatory hai because server startup ke time DB connection open karta hai. Provider credentials optional hain only tab tak jab tak aap provider-enabled flows, webhook, refund, retry, or payment intent end-to-end test nahi kar rahe.

## 4. Dependency Management

This is a Go project.

### Important Go dependency files

| File | Meaning |
|---|---|
| `go.mod` | Project module name, Go version, and required libraries define karta hai. |
| `go.sum` | Downloaded modules ke checksums store karta hai. Ye integrity check ke liye important hai. |

Current module:

```text
github.com/example/ecommerce-platform/backend/services/payment-service
```

Current direct external dependency:

```text
github.com/go-sql-driver/mysql v1.10.0
```

Indirect dependency:

```text
filippo.io/edwards25519 v1.2.0
```

### Common Go commands

Run these from:

```bash
cd backend/services/payment-service
```

Download dependencies:

```bash
go mod download
```

Clean unused dependencies and add missing ones:

```bash
go mod tidy
```

Run tests:

```bash
go test ./...
```

Build server binary:

```bash
go build ./cmd/server
```

Run server:

```bash
go run ./cmd/server
```

Run reconciliation command:

```bash
go run ./cmd/reconciliation
```

### Why Go dependencies fail

| Problem | Cause | Fix |
|---|---|---|
| `go: go.mod requires go >= 1.26.3` | Local Go version old hai. | Correct Go version install karo. `go version` verify karo. |
| `missing go.sum entry` | Dependency downloaded/verified nahi hua. | `go mod download` then `go mod tidy`. |
| `dial tcp: lookup proxy.golang.org` | Network/proxy issue. | Internet check karo, corporate proxy set karo, or `GOPROXY` configure karo. |
| `module declares its path as ...` | Wrong module import path or local replace issue. | `go.mod` inspect karo, imports align karo. |
| `permission denied` in Go cache | Go build cache path writable nahi hai. | `go env GOCACHE` check karo and permissions fix karo. |

Useful debug commands:

```bash
go env GOPATH
go env GOPROXY
go env GOMODCACHE
go env GOCACHE
go clean -modcache
```

Warning:

`go clean -modcache` cache clear karta hai. Iske baad dependencies dobara download hongi.

## 5. Database Setup

### A. Database used

This service uses MySQL.

Simple Hinglish:

MySQL ek relational database hai jisme data tables, rows, columns ke form me store hota hai. Payment Service me payments, payment attempts, refunds, webhook events, and reconciliation records MySQL me store hote hain.

### B. Why MySQL is used

Payment data highly structured hota hai and consistency important hoti hai. MySQL transactions, unique keys, foreign keys, indexes, and relational queries support karta hai. Payment systems me duplicate payments, duplicate webhooks, refund tracking, and audit records avoid karne ke liye MySQL useful hai.

### C. Required or optional

MySQL is required for running `cmd/server`.

Reason: server startup par code `repository.OpenMySQL(cfg.Database.DSN)` call karta hai and MySQL repository initialize karta hai. Agar MySQL unavailable hai to service start fail ho jayegi.

### D. Database name and tables

Database name:

```text
payment_db
```

Main tables:

| Table | Purpose |
|---|---|
| `payments` | Main payment record, provider ids, status, amount, retry lineage. |
| `payment_attempts` | Har provider attempt ka record. |
| `refunds` | Refund requests, status, review fields, idempotency. |
| `payment_webhook_events` | Provider webhook event idempotency and processing audit. |
| `payment_reconciliations` | Settlement/reconciliation result records. |

### E. Local installation

Windows:

```powershell
winget install Oracle.MySQL
mysql --version
```

Alternative: MySQL Installer download karke "MySQL Server" and "MySQL Shell/Client" install karo.

Linux Ubuntu/Debian:

```bash
sudo apt update
sudo apt install mysql-server mysql-client
sudo systemctl enable mysql
sudo systemctl start mysql
mysql --version
```

macOS:

```bash
brew install mysql
brew services start mysql
mysql --version
```

### F. Docker setup for MySQL

Docker beginner-friendly option hai because local machine par MySQL install/configure manually nahi karna padta.

Run MySQL container:

```bash
docker volume create ecommerce-payment-mysql-data

docker run -d \
  --name ecommerce-payment-mysql \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=payment_db \
  -p 3306:3306 \
  -v ecommerce-payment-mysql-data:/var/lib/mysql \
  mysql:8.4
```

Check container:

```bash
docker ps
docker logs ecommerce-payment-mysql
```

Stop container:

```bash
docker stop ecommerce-payment-mysql
```

Start again:

```bash
docker start ecommerce-payment-mysql
```

Remove container only:

```bash
docker rm ecommerce-payment-mysql
```

Do not remove the volume unless you intentionally want to delete local database data.

### G. docker-compose example

Repo me payment service ke liye committed `docker-compose.yml` currently nahi mila. Local beginner setup ke liye example:

```yaml
services:
  payment-mysql:
    image: mysql:8.4
    container_name: ecommerce-payment-mysql
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: root
      MYSQL_DATABASE: payment_db
    ports:
      - "3306:3306"
    volumes:
      - ecommerce-payment-mysql-data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "127.0.0.1", "-uroot", "-proot"]
      interval: 10s
      timeout: 5s
      retries: 10

volumes:
  ecommerce-payment-mysql-data:
```

Commands:

```bash
docker compose up -d
docker compose ps
docker compose logs payment-mysql
docker compose down
```

### H. Verify MySQL is running

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot -e "SELECT VERSION();"
```

Verify database:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot -e "SHOW DATABASES;"
```

### I. Default port

| Database | Default port |
|---|---:|
| MySQL | 3306 |

### J. Connection string format

The app expects a Go MySQL DSN:

```env
PAYMENT_MYSQL_DSN=root:root@tcp(127.0.0.1:3306)/payment_db?parseTime=true&charset=utf8mb4,utf8&loc=UTC
```

Fallback variable also supported:

```env
MYSQL_DSN=root:root@tcp(127.0.0.1:3306)/payment_db?parseTime=true&charset=utf8mb4,utf8&loc=UTC
```

Default inside code:

```text
root:root@tcp(127.0.0.1:3306)/payment_db?parseTime=true&charset=utf8mb4,utf8&loc=UTC
```

Security note:

Default `root:root` sirf local development ke liye acceptable hai. Production/staging me strong password and non-root DB user use karo.

### K. Run migrations

Migration files are in:

```text
backend/services/payment-service/migrations
```

Run from service folder:

```bash
cd backend/services/payment-service
```

Apply all up migrations in order:

```bash
for file in migrations/*.up.sql; do
  mysql -h 127.0.0.1 -P 3306 -u root -proot < "$file"
done
```

Verify tables:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db -e "SHOW TABLES;"
```

Expected tables:

```text
payment_attempts
payment_reconciliations
payment_webhook_events
payments
refunds
```

Schema health endpoint after server starts:

```bash
curl http://localhost:8080/internal/v1/payment-schema/health
```

If migrations are missing, this endpoint may return service unavailable or show missing tables/indexes.

## 6. Redis / Queue / External Services

### Redis

Current payment service code me Redis dependency nahi hai.

Simple Hinglish:

Redis ek in-memory cache hota hai jo fast data access ke liye use hota hai. Is service me current implementation directly Redis use nahi karta, so local Payment Service run ke liye Redis install karna required nahi hai.

### Kafka / RabbitMQ / NATS

Current payment service me Kafka, RabbitMQ, or NATS client dependency nahi hai.

Important:

Original task design me event publishing ka idea mention hai, but actual current implementation HTTP event publisher use karta hai. Matlab payment events kisi HTTP endpoint par `POST` hote hain, queue broker par nahi.

### HTTP event publisher

Required when:

- `PAYMENT_ALLOWED_PROVIDERS` set hai
- or `PAYMENT_RECONCILIATION_ENABLED=true`

Environment variables:

```env
PAYMENT_EVENTS_ENDPOINT=http://127.0.0.1:9080/events
PAYMENT_EVENTS_AUTH_TOKEN=event-publisher-token-at-least-32-characters
PAYMENT_EVENTS_TIMEOUT=5s
```

What it does:

Service payment/refund/reconciliation event ko JSON envelope ke form me configured endpoint par bhejta hai. Request headers include:

```text
Authorization: Bearer <PAYMENT_EVENTS_AUTH_TOKEN>
Idempotency-Key: <event id>
Content-Type: application/json
```

Security rule:

`PAYMENT_EVENTS_ENDPOINT` local testing ke bahar HTTPS hona chahiye. Code HTTP allow karta hai only loopback hosts like `127.0.0.1` or `localhost`.

Health check:

```bash
curl -i http://127.0.0.1:9080/events
```

Actual response depends on your event receiver. For end-to-end provider/webhook testing, endpoint ko POST accept karna and 2xx return karna chahiye.

### Payment provider integrations

Supported provider names:

| Provider | Default base URL | Required when enabled |
|---|---|---|
| `stripe_like` | `https://api.stripe.com` | Public key, secret key, webhook secret |
| `razorpay_like` | `https://api.razorpay.com` | Public key, secret key, webhook secret |

Enable provider example:

```env
PAYMENT_DEFAULT_PROVIDER=stripe_like
PAYMENT_ALLOWED_PROVIDERS=stripe_like
PAYMENT_ALLOWED_CURRENCIES=INR,USD
PAYMENT_CAPTURE_MODE=automatic

STRIPE_LIKE_PUBLIC_KEY=pk_test_example
STRIPE_LIKE_SECRET_KEY=sk_test_example
STRIPE_LIKE_WEBHOOK_SECRET=whsec_test_example
```

Provider webhook endpoints:

| Provider | Webhook URL path | Required signature header |
|---|---|---|
| `stripe_like` | `/api/v1/webhooks/payments/stripe_like` | `Stripe-Signature` |
| `razorpay_like` | `/api/v1/webhooks/payments/razorpay_like` | `X-Razorpay-Signature` |

Local full URLs:

```text
http://localhost:8080/api/v1/webhooks/payments/stripe_like
http://localhost:8080/api/v1/webhooks/payments/razorpay_like
```

Provider base URL rule:

- Empty base URL means default provider URL is used.
- `http://` is allowed only for localhost/loopback test servers.
- Remote non-local provider URLs must use `https://`.

### Reconciliation job

There is a separate command:

```bash
go run ./cmd/reconciliation
```

It is disabled by default:

```env
PAYMENT_RECONCILIATION_ENABLED=false
```

When enabled, it needs:

- MySQL
- settlement CSV file
- provider name
- HTTP event publisher config

Example:

```bash
go run ./cmd/reconciliation \
  -provider stripe_like \
  -report-file ./settlements/stripe-like-2026-06-01.csv \
  -report-date 2026-06-01
```

## 7. Environment Variables

### Where to create `.env`

Recommended local file path:

```text
backend/services/payment-service/.env
```

Important:

Current code does not use a dotenv library. It reads environment variables using `os.Getenv`. Iska matlab `.env` file automatically load nahi hota. Aapko variables shell me export/source karne honge, ya run command ke saath pass karne honge.

Linux/macOS load example:

```bash
cd backend/services/payment-service
set -a
. ./.env
set +a
go run ./cmd/server
```

Single-command example:

```bash
PAYMENT_MYSQL_DSN='root:root@tcp(127.0.0.1:3306)/payment_db?parseTime=true&charset=utf8mb4,utf8&loc=UTC' go run ./cmd/server
```

Windows PowerShell example:

```powershell
$env:PAYMENT_MYSQL_DSN="root:root@tcp(127.0.0.1:3306)/payment_db?parseTime=true&charset=utf8mb4,utf8&loc=UTC"
go run ./cmd/server
```

### Minimal local `.env`

Use this when you want to run health, state machine, and schema endpoints without real provider integration:

```env
PAYMENT_HTTP_ADDR=:8080
PAYMENT_MYSQL_DSN=root:root@tcp(127.0.0.1:3306)/payment_db?parseTime=true&charset=utf8mb4,utf8&loc=UTC
PAYMENT_MYSQL_MAX_OPEN_CONNS=25
PAYMENT_MYSQL_MAX_IDLE_CONNS=10
PAYMENT_MYSQL_CONN_MAX_LIFETIME=30m
PAYMENT_MYSQL_CONN_MAX_IDLE_TIME=5m

PAYMENT_WEBHOOK_MAX_BODY_BYTES=1048576
PAYMENT_REFUND_MANUAL_REVIEW_THRESHOLD_MINOR=0
PAYMENT_RETRY_MAX_ATTEMPTS=3
PAYMENT_RETRY_COOLDOWN_SECONDS=5

PAYMENT_RECONCILIATION_ENABLED=false
```

### Provider-enabled `.env`

Use this when payment intent, webhook, refund, retry, and provider flows need to run:

```env
PAYMENT_HTTP_ADDR=:8080
PAYMENT_HTTP_READ_TIMEOUT=5s
PAYMENT_HTTP_WRITE_TIMEOUT=10s
PAYMENT_HTTP_IDLE_TIMEOUT=60s
PAYMENT_SHUTDOWN_TIMEOUT=10s

PAYMENT_MYSQL_DSN=root:root@tcp(127.0.0.1:3306)/payment_db?parseTime=true&charset=utf8mb4,utf8&loc=UTC
PAYMENT_MYSQL_MAX_OPEN_CONNS=25
PAYMENT_MYSQL_MAX_IDLE_CONNS=10
PAYMENT_MYSQL_CONN_MAX_LIFETIME=30m
PAYMENT_MYSQL_CONN_MAX_IDLE_TIME=5m

PAYMENT_INTERNAL_API_TOKEN=change-this-internal-token-at-least-32-chars

PAYMENT_WEBHOOK_MAX_BODY_BYTES=1048576
PAYMENT_WEBHOOK_TIMESTAMP_TOLERANCE=5m

PAYMENT_DEFAULT_PROVIDER=stripe_like
PAYMENT_ALLOWED_PROVIDERS=stripe_like
PAYMENT_ALLOWED_CURRENCIES=INR,USD
PAYMENT_CAPTURE_MODE=automatic
PAYMENT_PROVIDER_TIMEOUT=5s

STRIPE_LIKE_PUBLIC_KEY=pk_test_replace_me
STRIPE_LIKE_SECRET_KEY=sk_test_replace_me
STRIPE_LIKE_WEBHOOK_SECRET=whsec_replace_me
STRIPE_LIKE_BASE_URL=
STRIPE_LIKE_TIMEOUT=5s

RAZORPAY_LIKE_PUBLIC_KEY=
RAZORPAY_LIKE_SECRET_KEY=
RAZORPAY_LIKE_WEBHOOK_SECRET=
RAZORPAY_LIKE_BASE_URL=
RAZORPAY_LIKE_TIMEOUT=5s

PAYMENT_EVENTS_ENDPOINT=http://127.0.0.1:9080/events
PAYMENT_EVENTS_AUTH_TOKEN=change-this-event-token-at-least-32-chars
PAYMENT_EVENTS_TIMEOUT=5s

PAYMENT_REFUND_MANUAL_REVIEW_THRESHOLD_MINOR=50000
PAYMENT_RETRY_MAX_ATTEMPTS=3
PAYMENT_RETRY_COOLDOWN_SECONDS=5

PAYMENT_RECONCILIATION_ENABLED=false
PAYMENT_RECONCILIATION_PROVIDER=stripe_like
PAYMENT_RECONCILIATION_REPORT_FILE=
PAYMENT_RECONCILIATION_REPORT_LAG_HOURS=24
PAYMENT_RECONCILIATION_BATCH_SIZE=500
PAYMENT_RECONCILIATION_TIMEOUT=20m
PAYMENT_RECONCILIATION_MAX_REPORT_FILE_BYTES=67108864
PAYMENT_RECONCILIATION_ALERT_TOPIC=payment.reconciliation.alerts
```

### Environment variable reference

| Variable | Required? | Example | Purpose | Security notes |
|---|---:|---|---|---|
| `PAYMENT_HTTP_ADDR` | Optional | `:8080` | HTTP server bind address. | Avoid exposing publicly in local dev. |
| `HTTP_ADDR` | Optional fallback | `:8080` | Fallback if `PAYMENT_HTTP_ADDR` empty. | Prefer service-specific variable. |
| `PAYMENT_HTTP_READ_TIMEOUT` | Optional | `5s` | Request read timeout. | Keep positive. |
| `PAYMENT_HTTP_WRITE_TIMEOUT` | Optional | `10s` | Response write timeout. | Keep positive. |
| `PAYMENT_HTTP_IDLE_TIMEOUT` | Optional | `60s` | Keep-alive idle timeout. | Keep positive. |
| `PAYMENT_SHUTDOWN_TIMEOUT` | Optional | `10s` | Graceful shutdown timeout. | Keep positive. |
| `PAYMENT_MYSQL_DSN` | Required | `root:root@tcp(127.0.0.1:3306)/payment_db?...` | MySQL connection string. | Contains password; never commit real value. |
| `MYSQL_DSN` | Optional fallback | Same as above | Fallback if `PAYMENT_MYSQL_DSN` empty. | Prefer service-specific variable. |
| `PAYMENT_MYSQL_MAX_OPEN_CONNS` | Optional | `25` | DB connection pool max open connections. | Too high can overload DB. |
| `PAYMENT_MYSQL_MAX_IDLE_CONNS` | Optional | `10` | DB idle connection count. | Must not exceed max open conns. |
| `PAYMENT_MYSQL_CONN_MAX_LIFETIME` | Optional | `30m` | Max lifetime of one DB connection. | Useful for DB/network rotation. |
| `PAYMENT_MYSQL_CONN_MAX_IDLE_TIME` | Optional | `5m` | Max idle time for one DB connection. | Keep positive. |
| `PAYMENT_INTERNAL_API_TOKEN` | Required when providers enabled | 32+ chars | Bearer token for internal/buyer/admin protected endpoints. | Secret. Use strong random value. |
| `PAYMENT_WEBHOOK_MAX_BODY_BYTES` | Optional | `1048576` | Max webhook body size. | Prevents huge request abuse. |
| `PAYMENT_WEBHOOK_TIMESTAMP_TOLERANCE` | Optional | `5m` | Allowed webhook timestamp skew. | Too large reduces replay protection. |
| `PAYMENT_DEFAULT_PROVIDER` | Required when providers enabled | `stripe_like` | Default provider for payment intent flow. | Must be in allowed providers. |
| `PAYMENT_ALLOWED_PROVIDERS` | Optional | `stripe_like,razorpay_like` | Enables provider integrations. | If set, provider secrets become required. |
| `PAYMENT_ALLOWED_CURRENCIES` | Optional | `INR,USD` | Allowed payment currencies. | Use ISO 3-letter codes. |
| `PAYMENT_CAPTURE_MODE` | Optional | `automatic` | Provider capture mode. Values: `automatic`, `manual`. | Must match provider support. |
| `PAYMENT_PROVIDER_TIMEOUT` | Optional | `5s` | Shared provider HTTP timeout fallback. | Keep finite. |
| `STRIPE_LIKE_PUBLIC_KEY` | Required if `stripe_like` enabled | `pk_test_...` | Client/public provider key. | Public-ish but still avoid random exposure. |
| `STRIPE_LIKE_SECRET_KEY` | Required if `stripe_like` enabled | `sk_test_...` | Provider API secret. | Secret. Never commit. |
| `STRIPE_LIKE_WEBHOOK_SECRET` | Required if `stripe_like` enabled | `whsec_...` | Webhook signature verification secret. | Secret. Never commit. |
| `STRIPE_LIKE_BASE_URL` | Optional | empty or `http://127.0.0.1:9999` | Override provider API base URL. | Remote URL must use HTTPS. |
| `STRIPE_LIKE_TIMEOUT` | Optional | `5s` | Stripe-like provider timeout. | Keep positive. |
| `RAZORPAY_LIKE_PUBLIC_KEY` | Required if `razorpay_like` enabled | `rzp_test_...` | Provider public/key id. | Avoid committing real value. |
| `RAZORPAY_LIKE_SECRET_KEY` | Required if `razorpay_like` enabled | `secret...` | Provider API secret. | Secret. Never commit. |
| `RAZORPAY_LIKE_WEBHOOK_SECRET` | Required if `razorpay_like` enabled | `webhook_secret` | Webhook signature verification secret. | Secret. Never commit. |
| `RAZORPAY_LIKE_BASE_URL` | Optional | empty or `http://127.0.0.1:9999` | Override provider API base URL. | Remote URL must use HTTPS. |
| `RAZORPAY_LIKE_TIMEOUT` | Optional | `5s` | Razorpay-like provider timeout. | Keep positive. |
| `PAYMENT_EVENTS_ENDPOINT` | Required when event publishing enabled | `http://127.0.0.1:9080/events` | HTTP event receiver endpoint. | HTTPS required outside local loopback. |
| `PAYMENT_EVENTS_AUTH_TOKEN` | Required when event publishing enabled | 32+ chars | Bearer token sent to event receiver. | Secret. Never commit. |
| `PAYMENT_EVENTS_TIMEOUT` | Optional | `5s` | Event publish timeout. | Keep positive. |
| `PAYMENT_REFUND_MANUAL_REVIEW_THRESHOLD_MINOR` | Optional | `50000` | Refund amount threshold in minor units. | Example: 50000 paise = INR 500. |
| `PAYMENT_RETRY_MAX_ATTEMPTS` | Optional | `3` | Max payment retry attempts. | Must be at least 2. |
| `PAYMENT_RETRY_COOLDOWN_SECONDS` | Optional | `5` | Retry cooldown in seconds. | Cannot be negative. |
| `PAYMENT_RECONCILIATION_ENABLED` | Optional | `false` | Enables reconciliation command behavior. | If true, event publisher config required. |
| `PAYMENT_RECONCILIATION_PROVIDER` | Required when reconciliation enabled | `stripe_like` | Provider used for settlement reconciliation. | Keep normalized provider name. |
| `PAYMENT_RECONCILIATION_REPORT_FILE` | Required at runtime when reconciliation enabled | `./settlements/file.csv` | Settlement CSV path. | File may contain financial data; protect it. |
| `PAYMENT_RECONCILIATION_REPORT_LAG_HOURS` | Optional | `24` | Default report date lag. | Cannot be negative. |
| `PAYMENT_RECONCILIATION_BATCH_SIZE` | Optional | `500` | Reconciliation batch size. | Must be 1 to 10000. |
| `PAYMENT_RECONCILIATION_TIMEOUT` | Optional | `20m` | Reconciliation job timeout. | Keep positive. |
| `PAYMENT_RECONCILIATION_MAX_REPORT_FILE_BYTES` | Optional | `67108864` | Max CSV file size. | Prevents accidental huge files. |
| `PAYMENT_RECONCILIATION_ALERT_TOPIC` | Required when reconciliation enabled | `payment.reconciliation.alerts` | Topic name in event payload. | No secret, but keep consistent. |

### Common `.env` mistakes

| Mistake | Result | Fix |
|---|---|---|
| `.env` created but not loaded | App still uses defaults/empty values. | `set -a; . ./.env; set +a` before running. |
| Token shorter than 32 chars | Config validation fails. | Use 32+ character random tokens. |
| `PAYMENT_ALLOWED_PROVIDERS` set but provider keys empty | App startup fails. | Fill provider keys or unset allowed providers. |
| Provider enabled but `PAYMENT_EVENTS_ENDPOINT` empty | App startup fails. | Set event endpoint and auth token. |
| MySQL password in DSN wrong | `access denied` or connection failure. | Match DSN with DB credentials. |
| `localhost` inside Docker points to wrong place | Container cannot reach host DB/service. | Use Docker network service name or host gateway. |

## 8. Docker Setup

### What exists now

Repo me currently payment service ke liye `Dockerfile` or committed `docker-compose.yml` nahi mila. Docker is still useful for MySQL local setup.

### Docker concepts

| Concept | Simple explanation |
|---|---|
| Image | Pre-built package, jaise `mysql:8.4`. |
| Container | Running instance of an image. |
| Volume | Persistent storage. Container delete hone ke baad bhi DB data bach sakta hai. |
| Network | Containers ek dusre se communicate karne ke liye virtual network. |
| Port mapping | Host machine ka port container ke port se connect karta hai. |

### Recommended beginner approach

Use Docker only for MySQL. Run Go service directly on host machine:

```bash
docker start ecommerce-payment-mysql

cd backend/services/payment-service
set -a
. ./.env
set +a
go run ./cmd/server
```

Why:

Ye setup beginner ke liye easiest hai. Database isolated container me hota hai, but code edit/run local machine par simple rehta hai.

### Useful Docker commands

```bash
docker ps
docker logs ecommerce-payment-mysql
docker exec -it ecommerce-payment-mysql mysql -u root -proot payment_db
docker stop ecommerce-payment-mysql
docker start ecommerce-payment-mysql
```

### Missing Docker improvements

Professional DevOps setup ke liye future me add karna useful hoga:

- `backend/services/payment-service/Dockerfile`
- root or service-level `docker-compose.yml`
- MySQL healthcheck
- service healthcheck hitting `/healthz`
- env file example
- separate dev/staging/prod config
- migration runner service
- non-root DB user
- restart policy

## 9. Local Development Setup

### Step 1: Clone repository

```bash
git clone <repo-url>
cd Ecommerce
```

### Step 2: Go to service directory

```bash
cd backend/services/payment-service
```

### Step 3: Verify Go version

```bash
go version
```

Expected: Go version compatible with `go.mod` value `1.26.3`.

### Step 4: Install Go dependencies

```bash
go mod download
go mod tidy
```

### Step 5: Start MySQL

Using Docker:

```bash
docker start ecommerce-payment-mysql
```

If container does not exist:

```bash
docker volume create ecommerce-payment-mysql-data

docker run -d \
  --name ecommerce-payment-mysql \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=payment_db \
  -p 3306:3306 \
  -v ecommerce-payment-mysql-data:/var/lib/mysql \
  mysql:8.4
```

### Step 6: Run migrations

```bash
for file in migrations/*.up.sql; do
  mysql -h 127.0.0.1 -P 3306 -u root -proot < "$file"
done
```

### Step 7: Create and load environment variables

Create local env file at:

```text
backend/services/payment-service/.env
```

Use the minimal local `.env` from section 7.

Load it:

```bash
set -a
. ./.env
set +a
```

### Step 8: Run tests

```bash
go test ./...
```

### Step 9: Start backend service

```bash
go run ./cmd/server
```

Expected log:

```text
payment.http.started addr=:8080
```

### Step 10: Verify APIs

Health:

```bash
curl http://localhost:8080/healthz
```

Payment state machine:

```bash
curl http://localhost:8080/internal/v1/payment-state-machine
```

Payment schema:

```bash
curl http://localhost:8080/internal/v1/payment-schema
```

Payment schema health:

```bash
curl http://localhost:8080/internal/v1/payment-schema/health
```

Validate transition:

```bash
curl -X POST http://localhost:8080/internal/v1/payment-state-machine/validate \
  -H 'Content-Type: application/json' \
  -d '{"from":"initiated","to":"authorized","event":"provider_authorized"}'
```

Provider event normalize:

```bash
curl -X POST http://localhost:8080/internal/v1/payment-state-machine/provider-event/normalize \
  -H 'Content-Type: application/json' \
  -d '{"provider_event":"payment_intent.succeeded"}'
```

Protected endpoint example:

```bash
curl -X POST http://localhost:8080/internal/v1/payment-intents \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer change-this-internal-token-at-least-32-chars' \
  -H 'X-Request-ID: req_local_001' \
  -d '{
    "order_id": "order_123",
    "user_id": "user_123",
    "amount": 10000,
    "currency": "INR",
    "idempotency_key": "idem_order_123_001",
    "provider": "stripe_like",
    "customer": {
      "name": "Test User",
      "email": "test@example.com",
      "phone": "+919999999999"
    },
    "metadata": {
      "source": "local"
    }
  }'
```

Note:

Protected provider endpoints require provider setup, internal token, event publisher config, and a reachable provider API or local mock provider base URL.

## 10. Running the Project

### Minimal local run

Use this for state machine and schema docs endpoints:

```bash
cd backend/services/payment-service

go mod download

docker start ecommerce-payment-mysql

for file in migrations/*.up.sql; do
  mysql -h 127.0.0.1 -P 3306 -u root -proot < "$file"
done

export PAYMENT_HTTP_ADDR=:8080
export PAYMENT_MYSQL_DSN='root:root@tcp(127.0.0.1:3306)/payment_db?parseTime=true&charset=utf8mb4,utf8&loc=UTC'

go run ./cmd/server
```

Verify:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/internal/v1/payment-state-machine
curl http://localhost:8080/internal/v1/payment-schema/health
```

### Provider-enabled run

Use provider-enabled `.env`, then:

```bash
cd backend/services/payment-service
set -a
. ./.env
set +a
go run ./cmd/server
```

Important checks:

- `PAYMENT_ALLOWED_PROVIDERS` should contain enabled provider names.
- `PAYMENT_DEFAULT_PROVIDER` should be one of allowed providers.
- Provider public key, secret key, and webhook secret must be set.
- `PAYMENT_INTERNAL_API_TOKEN` must be 32+ characters.
- `PAYMENT_EVENTS_ENDPOINT` and `PAYMENT_EVENTS_AUTH_TOKEN` must be set.

### Reconciliation run

```bash
cd backend/services/payment-service
set -a
. ./.env
set +a

export PAYMENT_RECONCILIATION_ENABLED=true

go run ./cmd/reconciliation \
  -provider stripe_like \
  -report-file ./settlements/report.csv \
  -report-date 2026-06-01
```

## 11. Ports & Networking

| Service | Port | Purpose |
|---|---:|---|
| Payment backend API | 8080 | Main HTTP service, health, internal APIs, webhooks. |
| MySQL | 3306 | Payment database. |
| HTTP event receiver | 9080 example | Receives payment/reconciliation events if configured. |
| Stripe-like provider API | 443 | External provider API when default base URL is used. |
| Razorpay-like provider API | 443 | External provider API when default base URL is used. |

### Port conflicts

If port `8080` already used:

```bash
export PAYMENT_HTTP_ADDR=:8081
go run ./cmd/server
```

Then verify:

```bash
curl http://localhost:8081/healthz
```

If MySQL port `3306` already used:

Run Docker MySQL on another host port:

```bash
docker run -d \
  --name ecommerce-payment-mysql-3307 \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=payment_db \
  -p 3307:3306 \
  mysql:8.4
```

Update DSN:

```env
PAYMENT_MYSQL_DSN=root:root@tcp(127.0.0.1:3307)/payment_db?parseTime=true&charset=utf8mb4,utf8&loc=UTC
```

### Firewall/network issues

| Problem | Fix |
|---|---|
| Local curl cannot connect | Check service running and correct port. |
| Docker MySQL not reachable | Check `docker ps`, port mapping, and DSN host/port. |
| Provider API timeout | Check internet, provider base URL, firewall/proxy. |
| Event publish fails | Check event receiver is running and returns 2xx. |
| Webhook not reaching local machine | Use a tunnel tool or provider local CLI if available. |

## 12. Common Errors & Fixes

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `PAYMENT_MYSQL_DSN cannot be empty` | DSN variable empty and fallback empty. | Set `PAYMENT_MYSQL_DSN`. | Keep `.env.example` updated. |
| `payment.mysql.open_failed` | DSN invalid or driver cannot open connection. | Check DSN format. | Use tested DSN template. |
| `connect: connection refused` | MySQL not running or wrong port. | Start MySQL, verify `3306`. | Add Docker healthcheck. |
| `Access denied for user` | Wrong DB username/password. | Match DSN with MySQL credentials. | Avoid changing password silently. |
| `Unknown database 'payment_db'` | DB not created. | Run migration `001_create_payment_tables.up.sql`. | Run all migrations before server. |
| Schema health not ready | Tables/indexes/FKs missing. | Run all migrations in order. | Use migration runner. |
| `bind: address already in use` | Port `8080` busy. | Change `PAYMENT_HTTP_ADDR=:8081`. | Document local port ownership. |
| `go: go.mod requires go >= ...` | Old Go installed. | Install correct Go version. | Check `go version` first. |
| `go mod download` fails | Network/proxy issue. | Check internet/proxy/GOPROXY. | Keep dependency cache in CI. |
| `PAYMENT_INTERNAL_API_TOKEN is required when payment providers are enabled` | Providers enabled but token missing. | Set 32+ char token. | Use provider-enabled `.env` template. |
| `PAYMENT_EVENTS_ENDPOINT is required when payment event publishing is enabled` | Providers or reconciliation enabled without event endpoint. | Set endpoint. | Start event receiver for end-to-end tests. |
| `PAYMENT_EVENTS_AUTH_TOKEN must be at least 32 characters` | Token too short. | Generate longer token. | Use random secret manager value. |
| `provider "stripe_like" public key is required` | Provider enabled but keys missing. | Fill provider env vars. | Do not set allowed provider until keys are ready. |
| `base URL must use HTTPS outside local testing` | Remote provider URL uses HTTP. | Use HTTPS or loopback HTTP. | Keep secure defaults. |
| `INVALID_WEBHOOK_SIGNATURE` | Wrong webhook secret/header/body. | Match provider webhook secret and signature header. | Do not modify raw webhook body before verification. |
| `WEBHOOK_UNAVAILABLE` | Provider registry/usecase not configured. | Enable provider env vars. | Test config on startup. |
| `UNAUTHORIZED` | Missing/incorrect Bearer token. | Send `Authorization: Bearer <PAYMENT_INTERNAL_API_TOKEN>`. | Use shared internal auth config. |
| `FORBIDDEN` refund endpoint | Missing admin role headers. | Add `X-Actor-ID` and `X-Actor-Role: finance_admin`. | Document caller contracts. |
| `IDEMPOTENCY_CONFLICT` | Same idempotency key used for different request. | Use new key for different payload. | Generate deterministic keys per operation. |
| Docker daemon not running | Docker Desktop/Engine stopped. | Start Docker. | Enable Docker on boot if needed. |
| Permission denied running MySQL/Docker | User lacks permission. | Use correct user/group/admin terminal. | Configure Docker permissions once. |
| Migration failed duplicate column/index | Migration already applied or partially applied. | Inspect schema, run only pending migrations. | Use migration tool with schema history. |

## 13. Security & Best Practices

### Security audit observations

| Observation | Risk | Suggested fix |
|---|---|---|
| Default DSN uses `root:root`. | Weak credential if reused outside local dev. | Use non-root DB user and strong password in env/secret manager. |
| No committed `.env.example` found for this service. | Beginners may miss required variables. | Add `.env.example` with safe placeholders. |
| Code does not auto-load `.env`. | Developers may think `.env` is active but app uses defaults. | Document `source .env` flow or add dotenv only if project wants it. |
| No Dockerfile/docker-compose for payment service found. | Local setup inconsistent across developers. | Add service Dockerfile and compose profile. |
| No migration runner tool found. | Manual SQL order can be error-prone. | Use a migration tool or Make target. |
| Provider/event tokens validate length but examples can still be weak. | Weak local patterns can leak into prod. | Generate strong random tokens. |
| Event publisher is HTTP-based, not queue-backed. | Event receiver downtime can break payment event handling. | Add retry/outbox pattern for production reliability. |
| Provider webhooks depend on raw body and signatures. | Body mutation breaks verification. | Keep raw body unchanged before verification. |
| Payment provider remote base URLs must be HTTPS. | Insecure provider traffic leaks sensitive data. | Keep HTTPS-only outside loopback. |
| Docker healthchecks missing in repo. | Service startup order issues. | Add MySQL and app healthchecks. |

### Best practices for beginners

- Never commit `.env` with real secrets.
- Keep `.env.example` with dummy values only.
- Use strong `PAYMENT_INTERNAL_API_TOKEN` and `PAYMENT_EVENTS_AUTH_TOKEN`.
- Use MySQL Docker volume so local data does not disappear after container restart.
- Run migrations before starting the service.
- Run `go test ./...` before pushing changes.
- Use idempotency keys for payment intent, retry, refund, and event publishing.
- Keep webhook secrets separate per environment.
- Use HTTPS for provider and event endpoints outside local testing.
- Keep provider keys in a secret manager for staging/production.
- Do not store card number, CVV, or raw payment credentials in app database/logs.
- Redact provider secrets from logs.
- Monitor `/healthz` and `/internal/v1/payment-schema/health`.
- Use separate dev/staging/prod databases.
- Back up payment database regularly.
- Keep dependencies updated and review `go.sum` changes.

## 14. Missing or Misconfigured Things

Based on current inspection, these setup items are missing or should be improved:

| Item | Current status | Why it matters |
|---|---|---|
| Service `.env` file | Not present under `backend/services/payment-service`. | New developers need a template. |
| `.env.example` | Not present for this service. | Safe onboarding without leaking secrets. |
| Dockerfile | Not found for payment service. | Needed for containerized deployment. |
| docker-compose | Not found for payment service. | Needed for repeatable local DB/service setup. |
| Migration runner | Not found. | Manual SQL execution can miss order/history. |
| Non-root DB user docs | Not present. | Production should avoid root DB account. |
| Event receiver setup | Not included in repo. | Provider-enabled flows need `PAYMENT_EVENTS_ENDPOINT`. |
| Queue broker docs | Not applicable to current code. | Avoid telling beginners to install Kafka/RabbitMQ when service does not use them. |
| Redis docs | Not applicable to current code. | Avoid unnecessary setup. |
| Kubernetes manifests | Not found. | Needed only for cluster deployment. |

## 15. Final Checklist

Use this checklist before saying local setup is done:

- [ ] Git installed.
- [ ] Correct Go version installed.
- [ ] MySQL server or Docker MySQL running.
- [ ] MySQL reachable on expected host/port.
- [ ] `payment_db` database exists.
- [ ] All `.up.sql` migrations applied in order.
- [ ] Go dependencies downloaded with `go mod download`.
- [ ] `.env` created locally or variables exported in shell.
- [ ] `.env` not committed to git.
- [ ] `PAYMENT_MYSQL_DSN` points to correct database.
- [ ] If providers disabled, `PAYMENT_ALLOWED_PROVIDERS` is empty/unset.
- [ ] If providers enabled, provider keys and webhook secrets are configured.
- [ ] If providers enabled, `PAYMENT_INTERNAL_API_TOKEN` is 32+ chars.
- [ ] If providers enabled, `PAYMENT_EVENTS_ENDPOINT` and token are configured.
- [ ] `go test ./...` passes.
- [ ] `go run ./cmd/server` starts successfully.
- [ ] `curl http://localhost:8080/healthz` returns OK.
- [ ] `curl http://localhost:8080/internal/v1/payment-state-machine` works.
- [ ] `curl http://localhost:8080/internal/v1/payment-schema/health` shows schema ready.
- [ ] Common errors section reviewed.
- [ ] Secrets are not printed in logs or committed.

Final beginner note:

Pehle minimal setup run karo: MySQL + migrations + `go run ./cmd/server`. Jab health and state-machine endpoints work karne lagen, tab provider-enabled setup par jao. Payment provider, webhook, refund, retry, and reconciliation flows me credentials, event endpoint, signatures, and idempotency keys correctly configured hona mandatory hai.
