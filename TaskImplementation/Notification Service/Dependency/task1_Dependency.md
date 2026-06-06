# Dependency and Setup Guide

## Variables Used

```text
SERVICE_NAME = "Notification Service"
TASK_FILE_NAME = "task1.md"
INPUT_PATH = TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}
OUTPUT_FILE_NAME = task1_Dependency.md
OUTPUT_PATH = TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

This guide is generated from `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}`. It does not replace or modify the original task file.

## Scope First

Task 1 is mainly a channel/provider abstraction document. Simple words me: Task 1 ka kaam email, SMS, push, aur WhatsApp-like channels ka common contract define karna hai. Is task ke markdown artifact ko read karne ke liye koi database, queue, Docker container, vendor account, ya secret key mandatory nahi hai.

Current repository me `backend/services/notification-service` ka runnable Go service code bhi available hai. Us current service ko start karne ke liye MongoDB required hai. RabbitMQ, SMTP/Mailpit, SMS HTTP gateway, metrics, and provider webhooks optional feature flags ke through enable hote hain.

## Tech Stack

| Technology | Required? | What it is | Why used here |
|---|---:|---|---|
| Go 1.26.3 | Required for runnable service | Go ek compiled backend language hai, fast APIs and microservices banane ke liye use hoti hai. | Service code, domain contracts, providers, gRPC server, repository, and tests Go me hain. |
| Go modules | Required | `go.mod` and `go.sum` dependency lock system hai. | External Go libraries ko exact versions ke saath manage karta hai. |
| gRPC + Protobuf | Required for service API | gRPC high-performance RPC framework hai; Protobuf contract/schema format hai. | Service `:9090` par gRPC server expose karta hai. |
| MongoDB | Required to start current service | MongoDB document database hai jisme JSON-like documents store hote hain. | Templates, deliveries, preferences, provider events store karne ke liye. |
| RabbitMQ | Optional | RabbitMQ message broker hai, async queues ke liye. | Event consumer and retry/DLQ flow ke liye, only when event consumer enabled hai. |
| Prometheus client | Optional at runtime | Metrics collect karne ka standard ecosystem. | Metrics endpoint enable hone par delivery analytics expose karta hai. |
| SMTP/Mailpit | Optional | SMTP email sending protocol hai; Mailpit local email catcher hai. | Email provider enable karne par OTP/email test delivery ke liye. |
| HTTP SMS gateway | Optional | External SMS provider ka HTTP API. | SMS provider enable karne par `POST {"to","text"}` request bhejta hai. |
| Docker | Optional but recommended | Containers run karne ka tool. | MongoDB, RabbitMQ, Mailpit local setup easy banata hai. |

## Repository Paths

```text
TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
backend/services/notification-service
backend/services/notification-service/.env
backend/services/notification-service/go.mod
backend/services/notification-service/migrations
backend/go.work
proto/ecommerce/notification/v1/notification.proto
```

## Fresh Clone Setup

```bash
git clone <repo-url>
cd Ecommerce
```

Verify Go:

```bash
go version
```

Expected for this repo:

```text
go version go1.26.3 linux/amd64
```

If your machine has older Go, install Go 1.26.3 or use the same version through your version manager. Version mismatch se `go mod download`, `go test`, ya `go run` fail ho sakta hai.

## Go Dependency System

Go project dependencies yahan manage hote hain:

```text
backend/services/notification-service/go.mod
backend/services/notification-service/go.sum
backend/go.work
```

`go.mod` batata hai module name, Go version, and direct dependencies. `go.sum` checksum lock file hai; ye ensure karta hai ki downloaded dependency same and trusted ho. `backend/go.work` local workspace file hai jo `auth-service` and `notification-service` modules ko ek saath use karta hai.

Important commands:

```bash
cd backend/services/notification-service
go mod download
go mod tidy
go test ./...
go build ./cmd/server
go run ./cmd/server
```

Workspace se command chalani ho:

```bash
cd backend
go test ./services/notification-service/...
```

Common Go issues:

| Problem | Reason | Fix |
|---|---|---|
| `go: module requires go >= 1.26.3` | Go version old hai. | Go upgrade karo. |
| `missing go.sum entry` | Dependency checksum absent hai. | `go mod tidy` run karo. |
| Network/proxy download fail | Go proxy ya internet issue. | `go env GOPROXY` check karo; retry with stable network. |
| Import path mismatch | Wrong working directory ya module path changed. | `backend/services/notification-service` se run karo. |
| Dirty dependency files | `go mod tidy` ne expected changes kiye. | Diff review karo before commit. |

Current direct Go libraries:

| Library | Use |
|---|---|
| `google.golang.org/grpc` | gRPC server and generated handler binding. |
| `google.golang.org/protobuf` | Generated protobuf message support. |
| `go.mongodb.org/mongo-driver/v2` | MongoDB connection, queries, indexes. |
| `github.com/rabbitmq/amqp091-go` | RabbitMQ event consumer and retry queues. |
| `github.com/prometheus/client_golang` | Prometheus metrics endpoint. |

Not used by this current service:

| Technology | Status |
|---|---|
| MySQL/PostgreSQL/SQLite | Not used by this service. |
| Redis | Mentioned in project-level docs, not required here. |
| Kafka | Mentioned as an architecture option, but this implementation uses RabbitMQ. |
| Typesense | Search-service concern, not needed here. |
| Node.js/Python runtime | Not needed for this Go backend service. |

## Environment Variables

Credentials and runtime config are stored in:

```text
backend/services/notification-service/.env
```

Important: Go automatically `.env` load nahi karta. If you run from shell, export variables first.

Linux/macOS:

```bash
cd backend/services/notification-service
set -a
. ./.env
set +a
go run ./cmd/server
```

Windows PowerShell:

```powershell
cd backend/services/notification-service
Get-Content .env | Where-Object { $_ -and $_ -notmatch '^#' } | ForEach-Object {
  $name, $value = $_ -split '=', 2
  [Environment]::SetEnvironmentVariable($name, $value, 'Process')
}
go run ./cmd/server
```

Minimum env for current service startup:

```env
NOTIFICATION_MONGO_URI=mongodb://ecommerce_root:ecommerce_password@localhost:27017/notification_db?authSource=admin
NOTIFICATION_MONGO_DATABASE=notification_db
NOTIFICATION_GRPC_ADDRESS=:9090
```

Complete env variable reference:

| Variable | Required? | Example/default | Beginner note |
|---|---:|---|---|
| `NOTIFICATION_MONGO_URI` | Yes | `mongodb://ecommerce_root:ecommerce_password@localhost:27017/notification_db?authSource=admin` | MongoDB ka full connection URL. |
| `NOTIFICATION_MONGO_DATABASE` | Optional | `notification_db` | Database name. |
| `NOTIFICATION_MONGO_TEMPLATES_COLLECTION` | Optional | `notification_templates` | Message templates collection. |
| `NOTIFICATION_MONGO_DELIVERIES_COLLECTION` | Optional | `notification_deliveries` | Delivery records collection. |
| `NOTIFICATION_MONGO_PREFERENCES_COLLECTION` | Optional | `notification_preferences` | User preferences collection. |
| `NOTIFICATION_MONGO_PROVIDER_EVENTS_COLLECTION` | Optional | `provider_events` | Provider webhook/events collection. |
| `NOTIFICATION_GRPC_ADDRESS` | Optional | `:9090` | gRPC listen address. |
| `NOTIFICATION_STARTUP_TIMEOUT` | Optional | `5s` | Startup dependency timeout. |
| `NOTIFICATION_SHUTDOWN_TIMEOUT` | Optional | `10s` | Graceful shutdown timeout. |
| `NOTIFICATION_METRICS_ENABLED` | Optional | `false` | Prometheus metrics enable karta hai. |
| `NOTIFICATION_ANALYTICS_HTTP_ADDRESS` | Optional | `:8081` | Internal metrics/webhook HTTP address. |
| `NOTIFICATION_METRICS_PATH` | Optional | `/metrics` | Prometheus metrics path. |
| `NOTIFICATION_PROVIDER_WEBHOOKS_ENABLED` | Optional | `false` | Provider callbacks enable karta hai; metrics bhi enabled honi chahiye. |
| `NOTIFICATION_WEBHOOK_PATH_PREFIX` | Optional | `/internal/provider-webhooks` | Webhook URL prefix. |
| `NOTIFICATION_WEBHOOK_MAX_BODY_BYTES` | Optional | `65536` | Max webhook body size. |
| `NOTIFICATION_WEBHOOK_REPLAY_WINDOW_SECONDS` | Optional | `300` | Signature replay protection window. |
| `NOTIFICATION_EMAIL_WEBHOOK_SIGNING_SECRET` | Required if email webhook enabled | secret | Email callback signature secret. |
| `NOTIFICATION_SMS_WEBHOOK_SIGNING_SECRET` | Required if SMS webhook enabled | secret | SMS callback signature secret. |
| `NOTIFICATION_PUSH_WEBHOOK_SIGNING_SECRET` | Required if push webhook enabled | secret | Push callback signature secret. |
| `NOTIFICATION_WHATSAPP_LIKE_WEBHOOK_SIGNING_SECRET` | Required if WhatsApp-like webhook enabled | secret | WhatsApp-like callback signature secret. |
| `NOTIFICATION_EMAIL_OPEN_TRACKING_ENABLED` | Optional | `false` | Email open tracking; requires email webhook. |
| `NOTIFICATION_EVENT_CONSUMER_ENABLED` | Optional | `false` | RabbitMQ event consumer enable karta hai. |
| `NOTIFICATION_RABBITMQ_URL` | Required if consumer enabled | `amqp://ecommerce:ecommerce_password@localhost:5672/ecommerce` | RabbitMQ connection URL. |
| `NOTIFICATION_ORDER_EVENTS_QUEUE` | Optional | `notification.order.events.v1` | Order events queue. |
| `NOTIFICATION_PAYMENT_EVENTS_QUEUE` | Optional | `notification.payment.events.v1` | Payment events queue. |
| `NOTIFICATION_USER_EVENTS_QUEUE` | Optional | `notification.user.events.v1` | User events queue. |
| `NOTIFICATION_CONSUMER_PREFETCH` | Optional | `10` | Consumer ek time par kitne messages reserve karega. |
| `NOTIFICATION_RETRY_MAX_ATTEMPTS` | Optional | `4` | Max delivery attempts. |
| `NOTIFICATION_RETRY_DELAY_1` | Optional | `30s` | First retry delay. |
| `NOTIFICATION_RETRY_DELAY_2` | Optional | `2m` | Second retry delay. |
| `NOTIFICATION_RETRY_DELAY_3` | Optional | `8m` | Third retry delay. |
| `NOTIFICATION_RETRY_PREFETCH` | Optional | `10` | Retry queue prefetch. |
| `NOTIFICATION_RETRY_PUBLISH_CONFIRM_TIMEOUT` | Optional | `5s` | RabbitMQ publish confirm wait. |
| `NOTIFICATION_RETRY_ATTEMPT_LEASE` | Optional | `30s` | Processing lease duration. |
| `NOTIFICATION_DELIVERY_ENCRYPTION_KEY` | Required if consumer enabled | base64 32 bytes | Recipient encryption key. |
| `NOTIFICATION_EMAIL_ENABLED` | Optional | `false` | Email channel enable flag. |
| `NOTIFICATION_EMAIL_PROVIDER` | Required if email enabled | `smtp` | Email provider name. |
| `NOTIFICATION_EMAIL_SMTP_HOST` | Required if email enabled | `localhost` | SMTP host. |
| `NOTIFICATION_EMAIL_SMTP_PORT` | Required if email enabled | `1025` | SMTP port. |
| `NOTIFICATION_EMAIL_FROM` | Required if email enabled | `no-reply@example.com` | Sender email. |
| `NOTIFICATION_EMAIL_SMTP_TLS_MODE` | Required if email enabled | `none` | `none`, `starttls`, or `tls`. |
| `NOTIFICATION_EMAIL_SMTP_USERNAME` | Optional | empty | SMTP username. |
| `NOTIFICATION_EMAIL_SMTP_PASSWORD` | Optional | empty | SMTP password. |
| `NOTIFICATION_EMAIL_TIMEOUT` | Optional | `5s` | Email provider timeout. |
| `NOTIFICATION_SMS_ENABLED` | Optional | `false` | SMS channel enable flag. |
| `NOTIFICATION_SMS_PROVIDER` | Required if SMS enabled | `http_gateway` | SMS provider name. |
| `NOTIFICATION_SMS_HTTP_ENDPOINT` | Required if SMS enabled | blank | SMS API endpoint. |
| `NOTIFICATION_SMS_HTTP_BEARER_TOKEN` | Required outside localhost | blank | SMS API token. |
| `NOTIFICATION_SMS_TIMEOUT` | Optional | `5s` | SMS provider timeout. |
| `NOTIFICATION_PUSH_ENABLED` | Optional | `false` | Push channel config flag. |
| `NOTIFICATION_PUSH_PROVIDER` | Required if push enabled | blank | Push provider name. |
| `NOTIFICATION_WHATSAPP_LIKE_ENABLED` | Optional | `false` | WhatsApp-like channel config flag. |
| `NOTIFICATION_WHATSAPP_LIKE_PROVIDER` | Required if enabled | blank | WhatsApp-like provider name. |

Channel/provider env:

| Variable | Required? | Default/Example | Notes |
|---|---:|---|---|
| `NOTIFICATION_EMAIL_ENABLED` | Optional | `false` | Enable SMTP email provider. |
| `NOTIFICATION_EMAIL_PROVIDER` | Required if email enabled | `smtp` | Provider name used by registry/logs. |
| `NOTIFICATION_EMAIL_SMTP_HOST` | Required if email enabled | `localhost` | Mailpit local host or real SMTP host. |
| `NOTIFICATION_EMAIL_SMTP_PORT` | Required if email enabled | `1025` | Mailpit SMTP port. |
| `NOTIFICATION_EMAIL_FROM` | Required if email enabled | `no-reply@example.com` | Must be valid email address. |
| `NOTIFICATION_EMAIL_SMTP_TLS_MODE` | Required if email enabled | `none`, `starttls`, `tls` | Use TLS in production. |
| `NOTIFICATION_EMAIL_SMTP_USERNAME` | Optional | empty | If set, password must also be set. |
| `NOTIFICATION_EMAIL_SMTP_PASSWORD` | Optional | empty | Keep in secret manager for production. |
| `NOTIFICATION_SMS_ENABLED` | Optional | `false` | Enable HTTP SMS provider. |
| `NOTIFICATION_SMS_PROVIDER` | Required if SMS enabled | `http_gateway` | Provider name. |
| `NOTIFICATION_SMS_HTTP_ENDPOINT` | Required if SMS enabled | `http://localhost:3001/send` | HTTPS required outside localhost. |
| `NOTIFICATION_SMS_HTTP_BEARER_TOKEN` | Required outside localhost | secret | Put in secret manager, not Git. |
| `NOTIFICATION_PUSH_ENABLED` | Optional | `false` | Config supported, current main wiring does not create push provider. |
| `NOTIFICATION_WHATSAPP_LIKE_ENABLED` | Optional | `false` | Config supported, current main wiring does not create WhatsApp-like provider. |

RabbitMQ/retry env:

| Variable | Required? | Example |
|---|---:|---|
| `NOTIFICATION_EVENT_CONSUMER_ENABLED` | Optional | `false` |
| `NOTIFICATION_RABBITMQ_URL` | Required if event consumer enabled | `amqp://ecommerce:ecommerce_password@localhost:5672/ecommerce` |
| `NOTIFICATION_ORDER_EVENTS_QUEUE` | Optional | `notification.order.events.v1` |
| `NOTIFICATION_PAYMENT_EVENTS_QUEUE` | Optional | `notification.payment.events.v1` |
| `NOTIFICATION_USER_EVENTS_QUEUE` | Optional | `notification.user.events.v1` |
| `NOTIFICATION_DELIVERY_ENCRYPTION_KEY` | Required if event consumer enabled | base64 encoded 32 bytes |

Generate local encryption key:

```bash
openssl rand -base64 32
```

Analytics/webhook env:

| Variable | Required? | Notes |
|---|---:|---|
| `NOTIFICATION_METRICS_ENABLED` | Optional | Enables internal HTTP listener. |
| `NOTIFICATION_ANALYTICS_HTTP_ADDRESS` | Optional | Default/example `:8081`. |
| `NOTIFICATION_METRICS_PATH` | Optional | Default `/metrics`. |
| `NOTIFICATION_PROVIDER_WEBHOOKS_ENABLED` | Optional | Requires metrics enabled. |
| `NOTIFICATION_EMAIL_WEBHOOK_SIGNING_SECRET` | Required if email webhook enabled | Secret manager value. |
| `NOTIFICATION_EMAIL_OPEN_TRACKING_ENABLED` | Optional | Only email supports open tracking. |

## MongoDB Setup

MongoDB mandatory hai jab current service ko run karna hai. Task 1 markdown ko read karne ke liye mandatory nahi.

Default port: `27017`

Connection string format:

```env
NOTIFICATION_MONGO_URI=mongodb://username:password@localhost:27017/notification_db?authSource=admin
```

Project example:

```env
NOTIFICATION_MONGO_URI=mongodb://ecommerce_root:ecommerce_password@localhost:27017/notification_db?authSource=admin
NOTIFICATION_MONGO_DATABASE=notification_db
```

### MongoDB With Docker

```bash
docker volume create ecommerce_notification_mongo_data
docker run --name ecommerce-notification-mongo \
  -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=ecommerce_root \
  -e MONGO_INITDB_ROOT_PASSWORD=ecommerce_password \
  -v ecommerce_notification_mongo_data:/data/db \
  -d mongo:7
```

Verify:

```bash
docker ps
mongosh "mongodb://ecommerce_root:ecommerce_password@localhost:27017/admin"
```

### MongoDB Install Without Docker

Windows:

1. Install MongoDB Community Server from MongoDB official installer.
2. Install MongoDB Shell (`mongosh`).
3. Start MongoDB as Windows Service.
4. Verify: `mongosh`.

Linux Ubuntu/Debian:

```bash
sudo apt-get update
sudo apt-get install -y mongodb-mongosh
```

For server install, follow MongoDB Community Server package steps for your distro.

macOS:

```bash
brew tap mongodb/brew
brew install mongodb-community mongodb-mongosh
brew services start mongodb/brew/mongodb-community
```

## MongoDB Migrations

Migration scripts are here:

```text
backend/services/notification-service/migrations/*.up.js
```

Run all up migrations in order:

```bash
cd backend/services/notification-service
set -a
. ./.env
set +a
for file in migrations/*.up.js; do
  mongosh "$NOTIFICATION_MONGO_URI" "$file"
done
```

Windows PowerShell:

```powershell
cd backend/services/notification-service
Get-ChildItem migrations/*.up.js | Sort-Object Name | ForEach-Object {
  mongosh $env:NOTIFICATION_MONGO_URI $_.FullName
}
```

What migrations create:

| Migration | Purpose |
|---|---|
| `001_create_notification_collections.up.js` | Template and delivery collections with validators and indexes. |
| `002_seed_renderable_templates.up.js` | Default templates like OTP, order status, payment status. |
| `003_allow_optional_user_delivery_for_otp.up.js` | OTP delivery schema adjustment. |
| `004_add_event_consumer_support.up.js` | Event-related delivery fields and templates. |
| `005_add_retry_dlq_support.up.js` | Retry/DLQ fields and indexes. |
| `006_add_notification_preferences_and_suppression.up.js` | Preference collection and suppression statuses. |
| `007_add_delivery_analytics.up.js` | Provider event analytics collection and delivery analytics fields. |

Verify collections:

```bash
mongosh "$NOTIFICATION_MONGO_URI" --eval 'db.getSiblingDB("notification_db").getCollectionNames()'
```

## RabbitMQ Setup

RabbitMQ optional hai. Enable only when you want event consumer + retry/DLQ flow.

Default ports:

| Port | Use |
|---:|---|
| `5672` | AMQP application connection |
| `15672` | Management UI |

Docker:

```bash
docker volume create ecommerce_notification_rabbitmq_data
docker run --name ecommerce-notification-rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  -e RABBITMQ_DEFAULT_USER=ecommerce \
  -e RABBITMQ_DEFAULT_PASS=ecommerce_password \
  -e RABBITMQ_DEFAULT_VHOST=ecommerce \
  -v ecommerce_notification_rabbitmq_data:/var/lib/rabbitmq \
  -d rabbitmq:3-management
```

Verify:

```bash
docker logs ecommerce-notification-rabbitmq
```

Native install notes:

Windows:

1. Install Erlang/OTP.
2. Install RabbitMQ Server for Windows.
3. Enable management plugin: `rabbitmq-plugins enable rabbitmq_management`.
4. Start RabbitMQ service from Services app or CLI.

Linux Ubuntu/Debian:

```bash
sudo apt-get update
sudo apt-get install -y rabbitmq-server
sudo systemctl enable --now rabbitmq-server
sudo rabbitmq-plugins enable rabbitmq_management
```

macOS:

```bash
brew install rabbitmq
brew services start rabbitmq
rabbitmq-plugins enable rabbitmq_management
```

Open UI:

```text
http://localhost:15672
username: ecommerce
password: ecommerce_password
```

Enable in `.env`:

```env
NOTIFICATION_EVENT_CONSUMER_ENABLED=true
NOTIFICATION_RABBITMQ_URL=amqp://ecommerce:ecommerce_password@localhost:5672/ecommerce
NOTIFICATION_EMAIL_ENABLED=true
NOTIFICATION_DELIVERY_ENCRYPTION_KEY=<base64-32-byte-key>
```

Note: current config validation requires email channel enabled when event consumer is enabled.

RabbitMQ topology created by service:

| Type | Names |
|---|---|
| Event exchanges | `order.events`, `payment.events`, `user.events` |
| Event queues | `notification.order.events.v1`, `notification.payment.events.v1`, `notification.user.events.v1` |
| Retry exchanges | `notification.delivery.attempt`, `notification.delivery.retry`, `notification.delivery.dlx` |
| Retry queues | `notification.delivery.attempt.v1`, delayed retry queues, `notification.delivery.dlq.v1` |

## Email Provider Setup With Mailpit

Mailpit optional local SMTP catcher hai. Real email send nahi hota; browser me captured email dikhta hai. Beginner setup ke liye safest option.

Ports:

| Port | Use |
|---:|---|
| `1025` | SMTP |
| `8025` | Web inbox |

Docker:

```bash
docker run --name ecommerce-notification-mailpit \
  -p 1025:1025 \
  -p 8025:8025 \
  -d axllent/mailpit
```

Open:

```text
http://localhost:8025
```

Enable email in `.env`:

```env
NOTIFICATION_EMAIL_ENABLED=true
NOTIFICATION_EMAIL_PROVIDER=smtp
NOTIFICATION_EMAIL_SMTP_HOST=localhost
NOTIFICATION_EMAIL_SMTP_PORT=1025
NOTIFICATION_EMAIL_FROM=no-reply@example.com
NOTIFICATION_EMAIL_SMTP_TLS_MODE=none
NOTIFICATION_EMAIL_SMTP_USERNAME=
NOTIFICATION_EMAIL_SMTP_PASSWORD=
```

Production SMTP:

```env
NOTIFICATION_EMAIL_SMTP_HOST=smtp.example.com
NOTIFICATION_EMAIL_SMTP_PORT=587
NOTIFICATION_EMAIL_SMTP_TLS_MODE=starttls
NOTIFICATION_EMAIL_SMTP_USERNAME=<from-secret-manager>
NOTIFICATION_EMAIL_SMTP_PASSWORD=<from-secret-manager>
```

## SMS Provider Setup

SMS optional hai. Current SMS client sends:

```json
{"to":"+919876543210","text":"Your code is 123456"}
```

Expected response:

```json
{"message_id":"provider-message-id"}
```

Local endpoint can use HTTP on localhost. Non-local endpoint must be HTTPS and should use bearer token.

`.env` example:

```env
NOTIFICATION_SMS_ENABLED=true
NOTIFICATION_SMS_PROVIDER=http_gateway
NOTIFICATION_SMS_HTTP_ENDPOINT=http://localhost:3001/send
NOTIFICATION_SMS_HTTP_BEARER_TOKEN=
```

## Docker Compose Example

No service-specific compose file was found in the repo. You can create a local scratch compose file if needed:

```yaml
services:
  mongo:
    image: mongo:7
    ports:
      - "27017:27017"
    environment:
      MONGO_INITDB_ROOT_USERNAME: ecommerce_root
      MONGO_INITDB_ROOT_PASSWORD: ecommerce_password
    volumes:
      - notification_mongo_data:/data/db

  rabbitmq:
    image: rabbitmq:3-management
    ports:
      - "5672:5672"
      - "15672:15672"
    environment:
      RABBITMQ_DEFAULT_USER: ecommerce
      RABBITMQ_DEFAULT_PASS: ecommerce_password
      RABBITMQ_DEFAULT_VHOST: ecommerce
    volumes:
      - notification_rabbitmq_data:/var/lib/rabbitmq

  mailpit:
    image: axllent/mailpit
    ports:
      - "1025:1025"
      - "8025:8025"

volumes:
  notification_mongo_data:
  notification_rabbitmq_data:
```

Run:

```bash
docker compose up -d
```

## Start The Service

Minimal local run:

```bash
cd backend/services/notification-service
set -a
. ./.env
set +a
go run ./cmd/server
```

Expected successful log includes:

```text
notification.grpc.started
```

Default ports:

| Port | Required? | Use |
|---:|---:|---|
| `9090` | Yes | gRPC server |
| `8081` | Only if metrics/webhooks enabled | Internal analytics HTTP |
| `27017` | Yes when running current service | MongoDB |
| `5672` | Only if event consumer enabled | RabbitMQ |
| `1025` | Only if email enabled locally | Mailpit SMTP |
| `8025` | Only for local email viewing | Mailpit UI |

## Test Commands

Unit tests:

```bash
cd backend/services/notification-service
go test ./...
```

Build:

```bash
cd backend/services/notification-service
go build ./cmd/server
```

From backend workspace:

```bash
cd backend
go test ./services/notification-service/...
```

## API And Proto Notes

Proto source:

```text
proto/ecommerce/notification/v1/notification.proto
```

Generated Go files are already present:

```text
backend/services/notification-service/api/notification/v1/notification.pb.go
backend/services/notification-service/api/notification/v1/notification_grpc.pb.go
```

Current RPCs:

| RPC | Purpose |
|---|---|
| `SendOTP` | Send OTP using email or SMS channel. |
| `GetNotificationPreference` | Read logged-in user's notification preferences. |
| `UpdateNotificationPreference` | Update email/SMS/push/marketing preferences. |

## Credentials Placement

Never put real secrets in task markdown or committed code.

| Credential | Where to put locally | Production placement |
|---|---|---|
| Mongo username/password | `.env` for local only | Secret manager/Kubernetes Secret |
| RabbitMQ username/password | `.env` for local only | Secret manager/Kubernetes Secret |
| SMTP username/password | `.env` for local only | Secret manager/Kubernetes Secret |
| SMS bearer token | `.env` for local only | Secret manager/Kubernetes Secret |
| Webhook signing secrets | `.env` for local only | Secret manager/Kubernetes Secret |
| Delivery encryption key | `.env` for local only | Secret manager/Kubernetes Secret |

## Common Setup Issues

| Error/Symptom | Likely cause | Fix |
|---|---|---|
| `NOTIFICATION_MONGO_URI is required` | Env file not loaded. | Export `.env` before `go run`. |
| `ping notification mongo` | MongoDB not running or wrong credentials. | Start MongoDB and verify URI. |
| `server selection error` | MongoDB port unreachable. | Check `docker ps`, port `27017`, and firewall. |
| `provider name is required` | Channel enabled but provider name empty. | Set `NOTIFICATION_EMAIL_PROVIDER` or disable channel. |
| `SMTP credentials require TLS` | Username/password set with TLS mode `none`. | Use `starttls`/`tls` or remove local credentials. |
| `notification event consumer requires the email channel to be enabled` | RabbitMQ consumer enabled without email channel. | Enable email provider or disable consumer. |
| `NOTIFICATION_DELIVERY_ENCRYPTION_KEY must be base64 encoding of 32 bytes` | Missing/invalid retry encryption key. | Generate with `openssl rand -base64 32`. |
| `address already in use :9090` | Another process is using gRPC port. | Stop old process or change `NOTIFICATION_GRPC_ADDRESS`. |
| Mailpit inbox empty | Email channel disabled or SMTP config wrong. | Set email enabled and host `localhost`, port `1025`. |
| SMS provider unavailable | Endpoint not running or response shape wrong. | Ensure endpoint returns `{"message_id":"..."}`. |

## Beginner Runbook

1. Clone repo.
2. Install Go 1.26.3.
3. Start MongoDB.
4. Export `backend/services/notification-service/.env`.
5. Run Mongo migrations.
6. Run `go test ./...`.
7. Start service with `go run ./cmd/server`.
8. Optional: start Mailpit and enable email.
9. Optional: start RabbitMQ, enable event consumer, add encryption key.
10. Optional: enable metrics and open `http://localhost:8081/metrics`.

## Task 1 Dependency Summary

For the original Task 1 markdown artifact:

| Dependency | Mandatory? |
|---|---:|
| Markdown viewer | Yes |
| Mermaid-capable viewer | Optional |
| Internet for Shields.io badge images | Optional |
| Go runtime | No, only for reference snippets/current service |
| MongoDB | No for Task 1; yes for current runnable service |
| RabbitMQ | No for Task 1; optional for event/retry runtime |
| Mailpit/SMTP | No for Task 1; optional for email runtime |
| SMS gateway | No for Task 1; optional for SMS runtime |

## Deployment Assumptions

Production deployment should not use committed `.env` secrets. Config values should come from a secret manager, Kubernetes Secret, or platform-level environment injection.

Expected production assumptions:

| Area | Assumption |
|---|---|
| Runtime | Linux container running the Go binary built from `./cmd/server`. |
| Network | gRPC `:9090` is internal service-to-service traffic. |
| MongoDB | Managed MongoDB or secured MongoDB cluster with auth, backups, and TLS. |
| RabbitMQ | Only needed when event consumer is enabled; use durable broker/cluster. |
| Metrics/webhooks | `:8081` should stay internal, not public internet. |
| Secrets | SMTP password, SMS token, webhook secrets, RabbitMQ password, Mongo password, and encryption key must be injected securely. |
| Logs | Do not log OTP, API keys, full email addresses, full phone numbers, or device tokens. |
| Migrations | Run reviewed migrations before deploying code that depends on new collections/indexes. |
