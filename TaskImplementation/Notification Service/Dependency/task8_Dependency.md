# Project Dependency & Setup Guide

## 1. Project Overview

### Variables Used

```text
SERVICE_NAME = "Notification Service"
TASK_FILE_NAME = "task8.md"
INPUT_FILE_PATH = TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}
OUTPUT_FILE_NAME = {TASK_FILE_NAME without ".md"}_Dependency.md
OUTPUT_FILE_PATH = TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}
```

This file documents only the dependency, environment, database, monitoring, webhook, and operational setup required by the delivery-analytics implementation associated with `{TASK_FILE_NAME}`. The original implementation file is not modified.

Simple Hinglish summary: current task notification ke `sent`, `delivered`, final `failed`, aur policy-allowed `opened` milestones ko MongoDB me durable form me record karta hai. Prometheus operational metrics expose karta hai, aur signed provider webhooks delivery updates receive karte hain.

### Current Runnable Scope

The runnable implementation is present in:

```text
backend/services/notification-service/internal/analytics/
backend/services/notification-service/internal/transport/http/metrics.go
backend/services/notification-service/internal/transport/http/provider_webhook.go
backend/services/notification-service/internal/repository/mongo_notification_repository.go
backend/services/notification-service/internal/domain/delivery.go
backend/services/notification-service/internal/domain/delivery_event.go
backend/services/notification-service/internal/config/config.go
backend/services/notification-service/cmd/server/main.go
backend/services/notification-service/migrations/007_add_delivery_analytics.up.js
```

Important runtime facts:

- MongoDB remains mandatory and now must support transactions for delivery milestone recording.
- `provider_events` is a new durable collection created by migration `007`.
- `github.com/prometheus/client_golang` is a new direct Go dependency, but it is already locked in `go.mod` and `go.sum`.
- Prometheus metric export is optional through a feature flag. Analytics event persistence is not optional in the current send paths.
- Provider webhooks are optional, but enabling them also requires Prometheus metrics to be enabled.
- The new internal HTTP listener defaults to `:8081` and serves both metrics and configured webhook routes.
- Grafana and a Prometheus server are not required for the Go process to start, but they are required for a real monitoring dashboard.
- No Redis, Kafka, NATS, SQL database, or new RabbitMQ topology is introduced.

> Warning: `{TASK_FILE_NAME}` describes the work as documentation/reference examples only, but the current repository contains runnable source, migration `007`, and the Prometheus dependency. Treat the runnable repository as the setup authority.

### Read Previous Dependency Docs First

Shared installation and setup are intentionally not repeated:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
TaskImplementation/{SERVICE_NAME}/task6_Dependency.md
TaskImplementation/{SERVICE_NAME}/task7_Dependency.md
```

Most important reused topics:

| Previous file | Reused topic |
|---|---|
| `task1_Dependency.md` | Clone, Go install/modules, `.env` loading, base MongoDB/provider/RabbitMQ setup, startup |
| `task2_Dependency.md` | Base delivery collection, Mongo indexes, and safe delivery data |
| `task3_Dependency.md` | Template keys used as a metric dimension |
| `task4_Dependency.md` | OTP provider setup and OTP privacy boundary |
| `task5_Dependency.md` | Event-driven sends and RabbitMQ runtime |
| `task6_Dependency.md` | Final failure/DLQ semantics and migration-order warnings |
| `task7_Dependency.md` | Consent/suppression rules and open-tracking policy context |

## 2. Tech Stack

| Technology | Required? | Current-task use | Beginner-friendly explanation |
|---|---:|---|---|
| Go `1.26.3` | Yes | Analytics recorder, HTTP handlers, config, and tests | Go compiled backend language hai. Same installation and common commands `task1_Dependency.md`, section `Go Dependency System` me explained hain. |
| Go modules | Yes | Locks Prometheus and existing service dependencies | `go.mod` dependency versions batata hai aur `go.sum` downloaded modules ke checksums verify karta hai. |
| MongoDB replica set or sharded cluster | Yes for actual send/webhook milestone recording | Stores normalized provider events and atomically updates delivery milestones | MongoDB document database hai. Replica set transaction support deta hai, jisse event insert aur delivery update ek atomic operation bante hain. |
| MongoDB Go Driver `v2.6.0` | Yes | Runs the event/milestone transaction | Existing direct dependency hai; separate install nahi chahiye. |
| Prometheus Go client `v1.23.2` | Yes in the compiled service | Creates counters/histograms and `/metrics` handler | Prometheus client app ke andar metrics banata hai. Package already installed/locked hai. |
| Prometheus server | Optional locally; required for monitoring | Scrapes and stores `/metrics` time series | Prometheus metrics endpoint ko periodically read karke historical operational data store karta hai. |
| Grafana | Optional locally; recommended/required for dashboards | Visualizes delivery quality and webhook health | Grafana Prometheus queries ko charts, cards, aur alerts me show karta hai. |
| Go standard library HTTP/HMAC/SHA-256 | Yes | Internal HTTP server and signed callback verification | Built-in packages hain; extra install nahi chahiye. |
| Provider webhook capability or adapter | Conditional | Supplies delivered/failed/opened callbacks | Real provider payload ko current normalized webhook contract me convert karna pad sakta hai. |
| RabbitMQ | Conditional, reused | Event sends/retry worker can produce sent/final-failed analytics | Current task koi new queue add nahi karta. |
| Redis, Kafka, NATS, SQL database | No | No implementation/config found | In services ko current task ke liye start mat karo. |

### Analytics Semantics Relevant To Setup

| Event | Meaning in current runtime |
|---|---|
| `sent` | Provider accepted a send and returned a provider message ID |
| `delivered` | A verified provider callback reported delivery |
| `failed` | Provider terminal reject or retries exhausted; temporary retry errors are not final failures |
| `opened` | A verified, globally enabled email open callback for a non-OTP delivery |

Suppressed notifications are not provider failures. OTP open callbacks are rejected even when email open tracking is enabled.

## 3. Required Software

| Software / service | Required? | Setup source / note |
|---|---:|---|
| Git | Yes for fresh clone | Refer `task1_Dependency.md`, section `Fresh Clone Setup` |
| Go `1.26.3` | Yes | Refer `task1_Dependency.md`, section `Go Dependency System` |
| MongoDB with transaction support | Yes | New requirement explained in section 5 |
| `mongosh` | Yes for migration and verification | Runs migration `007`, initializes a local replica set, and inspects events |
| `curl` | Recommended | Verifies `/metrics` and webhook responses |
| OpenSSL | Recommended | Generates local signing secrets and HMAC signatures |
| Prometheus server | Optional for endpoint smoke; required for time-series monitoring | No committed scrape config exists |
| Grafana | Optional for local smoke; recommended for dashboarding | No committed dashboard exists |
| Docker | Optional | Convenient for a local single-node Mongo replica set and monitoring tools |
| Working provider / manual fixture | Conditional | Needed to produce real or simulated callbacks |

Prometheus and Grafana installation is new only at the monitoring layer. Existing Go, provider, RabbitMQ, and general Docker setup should be reused from previous guides.

## 4. Dependency Management

Generic Go module installation, `go mod download`, `go mod tidy`, build/run commands, proxy/cache issues, and version mismatch troubleshooting are already explained in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: Go Dependency System
```

### New Direct Go Dependency

The current module already contains:

```text
github.com/prometheus/client_golang v1.23.2
```

It provides:

- Prometheus counters and histograms.
- An isolated metric registry for this service.
- The `promhttp` handler used by the configured metrics route.

Fresh-clone dependency command:

```bash
cd backend/services/notification-service
GOWORK=off go mod download
GOWORK=off go mod verify
```

Do not run `go get` only for the current task. Required versions and indirect dependencies are already locked.

Task-focused verification:

```bash
cd backend/services/notification-service
GOWORK=off go test ./internal/analytics ./internal/transport/http ./internal/config ./internal/repository ./internal/domain ./internal/retry
GOWORK=off go test ./internal/usecase -run 'Test.*(OTP|Event|DeliveryAttempt)'
GOWORK=off go build -o /tmp/notification-service-server ./cmd/server
```

Current repository verification:

- Task-focused tests pass.
- The server build passes.
- `GOWORK=off go mod verify` passes.
- `GOWORK=off go test ./...` currently fails at the pre-existing preference test `TestManagePreferencePatchPreservesExplicitFalseAndOmittedFields` because its injected update time is earlier than stored `created_at`.

## 5. Database Setup

### Reused MongoDB Setup

MongoDB basics, credentials, `mongosh`, base collections, generic migration execution, and common errors are already documented in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: MongoDB Setup
Section: MongoDB Migrations

TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
Section: Database Setup
```

### New Mandatory Requirement: MongoDB Transactions

The analytics repository inserts a provider event and updates the matching delivery milestone inside `WithTransaction`. MongoDB transactions require:

- A replica set, including a single-node replica set for local development; or
- A sharded production cluster that supports transactions.

An earlier standalone local MongoDB container may allow service startup and migrations, but actual send/webhook analytics will fail with a transaction-related error.

#### Beginner-Friendly Local Replica Set With Docker

This local-only example binds MongoDB to the host loopback interface and intentionally does not configure production authentication:

```bash
docker volume create ecommerce_notification_mongo_rs_data

docker run --name ecommerce-notification-mongo-rs \
  -p 127.0.0.1:27017:27017 \
  -v ecommerce_notification_mongo_rs_data:/data/db \
  -d mongo:7 \
  --replSet rs0 --bind_ip_all
```

Initialize once:

```bash
docker exec ecommerce-notification-mongo-rs mongosh --quiet --eval '
rs.initiate({
  _id: "rs0",
  members: [{ _id: 0, host: "localhost:27017" }]
})
'
```

Verify:

```bash
docker exec ecommerce-notification-mongo-rs mongosh --quiet --eval 'rs.status().members.map(m => ({name:m.name,stateStr:m.stateStr}))'
```

Expected state is `PRIMARY`.

Local host-run connection string:

```env
NOTIFICATION_MONGO_URI=mongodb://localhost:27017/notification_db?replicaSet=rs0
```

> Warning: this no-auth example is only for an isolated developer machine and is bound to `127.0.0.1`. Production must use authenticated, TLS-protected, access-controlled MongoDB. Do not expose this local container publicly.

If the service and MongoDB run in containers, use a resolvable Docker service name as the replica-set member host and in the URI. A replica-set member advertised as `localhost` is only appropriate when the Go service runs on the host.

### New Database Change: Migration `007`

Task-specific files:

```text
backend/services/notification-service/migrations/007_add_delivery_analytics.up.js
backend/services/notification-service/migrations/007_add_delivery_analytics.down.js
```

Migration `007` performs these changes:

| Change | Purpose |
|---|---|
| Adds `campaign_id` | Supports durable campaign-range queries when a safe campaign ID exists |
| Adds `sent_at`, `delivered_at`, `failed_at`, `opened_at` | Stores first observed delivery milestones |
| Adds `failure_code` | Stores sanitized terminal failure reason |
| Creates `provider_events` with strict validation | Stores normalized, recipient-free provider lifecycle events |
| Creates unique index `provider_event_dedup` | Prevents the same provider event from being applied twice |
| Creates index `delivery_event_timeline` | Supports per-delivery event investigation |
| Creates index `template_channel_quality` | Supports durable quality reporting by template/channel/type/time |
| Creates sparse index `campaign_delivery_range` | Supports campaign delivery-range queries |
| Replaces the delivery validator | Allows analytics fields while retaining earlier retry/preference rules |

### Migration Order Warning

> Warning: migration `007` uses MongoDB `collMod` and replaces the current delivery validator. It must run after migrations `001` through `006`. Re-running an older validator migration afterward can remove analytics validation support.

Use the correct flow:

| Database state | Correct action |
|---|---|
| Fresh database | Run all current `.up.js` files once in filename order |
| Database has `001` through `006` only | Run migration `007` once |
| Database already has `007` | Do not replay older migrations; verify schema and indexes |

The repository has no automated migration-history tracker. Team ko applied migration versions carefully record karne chahiye.

### Run Migration `007`

For a database known to be at migration `006`:

```bash
cd backend/services/notification-service
set -a
. ./.env
set +a
mongosh "$NOTIFICATION_MONGO_URI" migrations/007_add_delivery_analytics.up.js
```

For a fresh current-repository database:

```bash
cd backend/services/notification-service
set -a
. ./.env
set +a
for file in migrations/*.up.js; do
  mongosh "$NOTIFICATION_MONGO_URI" "$file"
done
```

Run the fresh-database loop only once. It is not a safe migration replay mechanism for an existing environment.

### Verify Migration `007`

Verify collection and indexes:

```bash
mongosh "$NOTIFICATION_MONGO_URI" --quiet --eval '
const d = db.getSiblingDB(process.env.NOTIFICATION_MONGO_DATABASE || "notification_db");
const events = process.env.NOTIFICATION_MONGO_PROVIDER_EVENTS_COLLECTION || "provider_events";
const deliveries = process.env.NOTIFICATION_MONGO_DELIVERIES_COLLECTION || "notification_deliveries";
printjson(d.getCollectionInfos({ name: events }));
printjson(d.getCollection(events).getIndexes());
printjson(d.getCollection(deliveries).getIndexes().filter(i => i.name === "campaign_delivery_range"));
'
```

Expected named indexes:

```text
provider_event_dedup
delivery_event_timeline
template_channel_quality
campaign_delivery_range
```

Verify recent milestones without printing recipient/content data:

```bash
mongosh "$NOTIFICATION_MONGO_URI" --quiet --eval '
const d = db.getSiblingDB(process.env.NOTIFICATION_MONGO_DATABASE || "notification_db");
const deliveries = process.env.NOTIFICATION_MONGO_DELIVERIES_COLLECTION || "notification_deliveries";
printjson(d.getCollection(deliveries).find(
  { $or: [
    { sent_at: { $exists: true } },
    { delivered_at: { $exists: true } },
    { failed_at: { $exists: true } },
    { opened_at: { $exists: true } }
  ] },
  {
    _id: 1, channel: 1, template_key: 1, provider: 1,
    sent_at: 1, delivered_at: 1, failed_at: 1, opened_at: 1, failure_code: 1
  }
).sort({ updated_at: -1 }).limit(10).toArray());
'
```

### Rollback Caution

`007_add_delivery_analytics.down.js`:

- Removes analytics milestone and campaign fields from all deliveries.
- Drops the `campaign_delivery_range` index.
- Drops the entire configured provider-event collection.
- Restores the older preference-aware validator.

Rollback is destructive. Backup data, stop writes, and obtain explicit approval before using it.

## 6. Redis / Queue / External Services

### External Service Summary

| Service | Required? | Current-task change |
|---|---:|---|
| MongoDB replica set / transaction-capable cluster | Yes | Transaction support is newly mandatory for analytics writes |
| Prometheus server | Optional for local smoke; required for monitoring | New scraper/storage integration |
| Grafana | Optional for local smoke; recommended for dashboards | New dashboard consumer |
| Provider webhook sender/adapter | Required only for delivered/failed/opened callbacks | New signed callback integration |
| RabbitMQ | Conditional, reused | No new exchanges or queues |
| Mailpit/SMTP and SMS provider | Conditional, reused | Can produce `sent`; basic local providers do not automatically emit callbacks |
| Redis/Kafka/NATS | No | No current-task adapter or configuration |

### Prometheus Metrics

When enabled, the service exposes:

| Metric | Type | Labels | Purpose |
|---|---|---|---|
| `notification_delivery_events_total` | Counter | `event`, `channel`, `provider`, `template_key` | Counts deduplicated milestone changes |
| `notification_provider_webhooks_total` | Counter | `provider`, `outcome` | Counts accepted, duplicate, invalid, unmatched, and failed callbacks |
| `notification_provider_webhook_processing_seconds` | Histogram | `provider` | Measures callback processing time |
| `notification_delivery_event_lag_seconds` | Histogram | `event`, `channel`, `provider` | Measures provider-event ingestion lag |

Prometheus counters are in memory. Process restart ke baad counter values reset ho sakte hain; MongoDB milestone/provider-event data durable source hai.

Minimal scrape concept:

```yaml
scrape_configs:
  - job_name: notification-service
    metrics_path: /metrics
    static_configs:
      - targets: ["notification-service-host:8081"]
```

Replace the target with an address resolvable from the Prometheus process. If Prometheus runs in Docker and the Go service runs on the host, configure approved host networking instead of assuming `localhost` means the host.

Grafana should use Prometheus as its datasource. No dashboard JSON or alert rules are committed, so panels and thresholds must be created/reviewed by the monitoring team.

Task-specific starter PromQL:

```promql
sum(rate(notification_delivery_events_total{event="sent"}[5m]))
```

```promql
sum(rate(notification_delivery_events_total{event="delivered"}[5m]))
/
clamp_min(sum(rate(notification_delivery_events_total{event="sent"}[5m])), 0.000001)
```

```promql
sum by (provider, outcome) (
  rate(notification_provider_webhooks_total[5m])
)
```

Monitoring health checks, after the platform tools are configured:

```bash
curl -fsS http://127.0.0.1:8081/metrics >/dev/null
curl -fsS http://127.0.0.1:9091/-/ready
curl -fsS http://127.0.0.1:3000/api/health
```

The examples assume Prometheus was deliberately mapped to local port `9091` to avoid the existing gRPC `9090` port. Adjust addresses for the approved deployment. Prometheus/Grafana package installation and container lifecycle remain platform-owned because the repository has no committed portable monitoring stack.

### Provider Webhook Contract

The current HTTP handler expects a normalized JSON body:

```json
{
  "provider_event_id": "evt_123",
  "provider_message_id": "msg_123",
  "type": "delivered",
  "occurred_at": "2026-06-06T12:00:00Z"
}
```

For a terminal failure, include a safe failure code:

```json
{
  "provider_event_id": "evt_124",
  "provider_message_id": "msg_123",
  "type": "failed",
  "occurred_at": "2026-06-06T12:05:00Z",
  "failure_code": "hard_bounce"
}
```

Required headers:

```text
X-Notification-Timestamp: <unix-seconds>
X-Notification-Signature: sha256=<hex-hmac>
```

Signature input:

```text
<timestamp>.<raw-request-body>
```

Real email/SMS vendors often use different payloads and signature schemes. An adapter or gateway must normalize and re-sign them before calling this handler unless the provider already matches this exact contract.

## 7. Environment Variables

### Reused Environment Setup

`.env` location, loading behavior, credentials placement, and all existing Mongo/provider/RabbitMQ variables are already documented in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: Environment Variables
Section: Credentials Placement
```

The Go service does not automatically load `.env`. Export it before migrations or startup:

```bash
cd backend/services/notification-service
set -a
. ./.env
set +a
```

### New Or Changed Variables

| Variable | Required? | Default / example | Purpose | Security / coupling note |
|---|---:|---|---|---|
| `NOTIFICATION_MONGO_PROVIDER_EVENTS_COLLECTION` | Optional | `provider_events` | Normalized provider-event collection name | Must be distinct and must match migration `007` |
| `NOTIFICATION_METRICS_ENABLED` | Optional | `false` | Enables Prometheus observer and metrics route | Also required before webhooks can be enabled |
| `NOTIFICATION_ANALYTICS_HTTP_ADDRESS` | Required when metrics/webhooks enabled | `:8081` | Internal HTTP listener address | Default binds all interfaces; restrict in deployment |
| `NOTIFICATION_METRICS_PATH` | Optional | `/metrics` | Prometheus scrape route | Must be an absolute path and differ from webhook prefix |
| `NOTIFICATION_PROVIDER_WEBHOOKS_ENABLED` | Optional | `false` | Enables signed provider callback routes | Requires metrics enabled and at least one valid channel secret |
| `NOTIFICATION_WEBHOOK_PATH_PREFIX` | Optional | `/internal/provider-webhooks` | Callback route prefix | Final route is `<prefix>/<channel>` |
| `NOTIFICATION_WEBHOOK_MAX_BODY_BYTES` | Optional | `65536` | Rejects oversized callback bodies | Must be positive |
| `NOTIFICATION_WEBHOOK_REPLAY_WINDOW_SECONDS` | Optional | `300` | Acceptable timestamp skew/replay window | Must be positive; keep clocks synchronized |
| `NOTIFICATION_EMAIL_WEBHOOK_SIGNING_SECRET` | Conditional | Secret manager value | Enables email callback endpoint | Email channel must also be enabled |
| `NOTIFICATION_SMS_WEBHOOK_SIGNING_SECRET` | Conditional | Secret manager value | Enables SMS callback endpoint | SMS channel must also be enabled |
| `NOTIFICATION_PUSH_WEBHOOK_SIGNING_SECRET` | Conditional | Secret manager value | Enables push callback endpoint | Push channel/provider must be runnable first |
| `NOTIFICATION_WHATSAPP_LIKE_WEBHOOK_SIGNING_SECRET` | Conditional | Secret manager value | Enables WhatsApp-like callback endpoint | Channel/provider must be runnable first |
| `NOTIFICATION_EMAIL_OPEN_TRACKING_ENABLED` | Optional | `false` | Allows non-OTP email open callbacks | Requires webhooks and email signing secret; privacy approval needed |
| `NOTIFICATION_SMS_OPEN_TRACKING_ENABLED` | Must remain false | `false` | Parsed by config but unsupported | Startup rejects `true` |
| `NOTIFICATION_PUSH_OPEN_TRACKING_ENABLED` | Must remain false | `false` | Parsed by config but unsupported | Startup rejects `true` |
| `NOTIFICATION_WHATSAPP_LIKE_OPEN_TRACKING_ENABLED` | Must remain false | `false` | Parsed by config but unsupported | Startup rejects `true` |

Existing channel flags and provider names become coupled to webhook configuration:

- A signing secret for a disabled channel causes startup failure.
- Webhooks require at least one signing secret for an enabled channel.
- Email open tracking requires the email channel, webhooks, metrics, and email signing secret.
- Current main wiring creates concrete outbound providers only for email and SMS.

### Incremental `.env` Example: Metrics Only

```env
# Existing transaction-capable MongoDB URI.
NOTIFICATION_MONGO_URI=mongodb://localhost:27017/notification_db?replicaSet=rs0

# Current-task analytics storage and internal metrics listener.
NOTIFICATION_MONGO_PROVIDER_EVENTS_COLLECTION=provider_events
NOTIFICATION_METRICS_ENABLED=true
NOTIFICATION_ANALYTICS_HTTP_ADDRESS=127.0.0.1:8081
NOTIFICATION_METRICS_PATH=/metrics

# Keep callbacks disabled for a metrics-only local smoke.
NOTIFICATION_PROVIDER_WEBHOOKS_ENABLED=false
NOTIFICATION_EMAIL_OPEN_TRACKING_ENABLED=false
```

### Incremental `.env` Example: Local Email Webhook Fixture

This example assumes the reused email/Mailpit variables are already configured:

```env
NOTIFICATION_METRICS_ENABLED=true
NOTIFICATION_ANALYTICS_HTTP_ADDRESS=127.0.0.1:8081
NOTIFICATION_PROVIDER_WEBHOOKS_ENABLED=true
NOTIFICATION_WEBHOOK_PATH_PREFIX=/internal/provider-webhooks
NOTIFICATION_WEBHOOK_MAX_BODY_BYTES=65536
NOTIFICATION_WEBHOOK_REPLAY_WINDOW_SECONDS=300
NOTIFICATION_EMAIL_WEBHOOK_SIGNING_SECRET=<generate-and-keep-local>
NOTIFICATION_EMAIL_OPEN_TRACKING_ENABLED=false
```

Generate a local-only secret:

```bash
openssl rand -hex 32
```

Never commit real provider signing secrets. Production values should come from a secret manager, Kubernetes Secret, or approved runtime injection.

### Common Environment Mistakes

- Enabling webhooks while metrics are disabled.
- Adding a webhook secret but leaving its channel disabled.
- Enabling email open tracking without webhooks or an email secret.
- Setting any non-email open-tracking flag to `true`.
- Using the same path for metrics and webhook prefix.
- Using a standalone Mongo URI without replica-set/transaction support.
- Changing the provider-event collection name without running migration `007` with the same environment.
- Binding `:8081` publicly without network controls.

## 8. Docker Setup

No committed service-specific `Dockerfile`, Docker Compose file, Prometheus config, Grafana dashboard, or alert rules were found.

Reuse existing container instructions from:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: RabbitMQ Setup
Section: Email Provider Setup With Mailpit
Section: Docker Compose Example
```

The earlier standalone MongoDB example is not enough for current analytics writes. Use the transaction-capable local replica-set example in section 5 or an approved managed MongoDB cluster.

### Current-Task Docker Impact

| Docker item | Status |
|---|---|
| MongoDB container | Changed: must support transactions |
| MongoDB persistent volume | Still required; new replica-set-aware local volume recommended |
| Notification service image | Missing; no Dockerfile committed |
| Prometheus container/config | Missing; optional local tool, required platform integration |
| Grafana container/dashboard | Missing; optional local tool, required dashboard integration |
| New RabbitMQ container/topology | None |
| New application Docker network | None committed |
| New restart policy / health check | None committed |

### Ports & Networking

| Service / endpoint | Port | Purpose | Status |
|---|---:|---|---|
| MongoDB | `27017` | Delivery/provider-event persistence and transactions | Reused, transaction capability newly required |
| gRPC service | `9090` | Existing internal service API | Reused |
| Analytics HTTP | `8081` | Metrics and signed provider webhooks | New |
| RabbitMQ AMQP | `5672` | Existing event/retry flow | Reused, conditional |
| RabbitMQ UI | `15672` | Existing broker inspection | Reused, optional |
| Mailpit SMTP/UI | `1025` / `8025` | Existing local email provider | Reused, optional |
| Prometheus UI/server | commonly `9090` | Monitoring scraper/storage | New external tool; conflicts with gRPC on the same host if default port is used |
| Grafana UI | commonly `3000` | Dashboard UI | New external tool |

Port guidance:

- For local metrics-only smoke, bind `NOTIFICATION_ANALYTICS_HTTP_ADDRESS=127.0.0.1:8081`.
- For container/Kubernetes scraping, bind on a reachable internal interface and restrict ingress.
- Do not expose `/metrics` to public internet traffic.
- Provider webhooks may need public HTTPS ingress, but `/metrics` shares the same listener. Route only the webhook prefix publicly through an approved reverse proxy/API gateway.
- If local Prometheus uses its default `9090`, change its host mapping or change the service gRPC address to avoid a same-host conflict.
- Keep system clocks synchronized because callback timestamps outside the replay window are rejected.

## 9. Local Development Setup

### Step 1: Read Previous Dependency Documentation

Start with the previous guides listed in section 1. They contain clone, Go, `.env`, provider, RabbitMQ, and generic Docker instructions.

### Step 2: Go To The Runnable Service

```bash
cd backend/services/notification-service
```

### Step 3: Install Existing Dependencies

```bash
GOWORK=off go mod download
GOWORK=off go mod verify
```

No current-task `go get` is needed.

### Step 4: Start A Transaction-Capable MongoDB

Use:

- The local single-node replica-set example from section 5; or
- An approved authenticated managed replica set/sharded cluster.

Verify `PRIMARY` or managed-cluster health before sending notifications.

### Step 5: Add And Load Current-Task Environment

Add only the new/changed values from section 7 to:

```text
backend/services/notification-service/.env
```

Then load:

```bash
set -a
. ./.env
set +a
```

### Step 6: Run Migration `007`

For an incremental database at migration `006`:

```bash
mongosh "$NOTIFICATION_MONGO_URI" migrations/007_add_delivery_analytics.up.js
```

For a fresh database, run all `.up.js` migrations once in order.

### Step 7: Run Tests And Build

```bash
GOWORK=off go test ./internal/analytics ./internal/transport/http ./internal/config ./internal/repository ./internal/domain ./internal/retry
GOWORK=off go test ./internal/usecase -run 'Test.*(OTP|Event|DeliveryAttempt)'
GOWORK=off go build -o /tmp/notification-service-server ./cmd/server
```

### Step 8: Start The Service

```bash
GOWORK=off go run ./cmd/server
```

Expected logs when metrics or webhooks are enabled:

```text
notification.grpc.started
notification.analytics_http.started
```

RabbitMQ logs appear only when the reused event-consumer flag is enabled.

## 10. Running The Project

### Verify Metrics Endpoint

```bash
curl -fsS http://127.0.0.1:8081/metrics | grep '^notification_'
```

Before events occur, some labeled metric series may not appear yet. After a send or callback, expect current-task metric names from section 6.

### Verify Durable `sent` Or Final `failed`

Use an existing OTP or event-delivery smoke flow from previous dependency guides. After a provider-accepted send:

```bash
mongosh "$NOTIFICATION_MONGO_URI" --quiet --eval '
const d = db.getSiblingDB(process.env.NOTIFICATION_MONGO_DATABASE || "notification_db");
const events = process.env.NOTIFICATION_MONGO_PROVIDER_EVENTS_COLLECTION || "provider_events";
printjson(d.getCollection(events).find(
  { type: { $in: ["sent", "failed"] } },
  { _id: 0, provider: 1, delivery_id: 1, type: 1, channel: 1, template_key: 1, occurred_at: 1, failure_code: 1 }
).sort({ received_at: -1 }).limit(10).toArray());
'
```

No recipient, rendered message, raw webhook, or signing secret should appear in provider-event documents.

### Send A Signed Local Webhook Fixture

Prerequisites:

- Webhooks and metrics enabled.
- Email channel and email webhook signing secret configured.
- An existing email delivery has provider `smtp` and a known `provider_message_id`.
- Replace `msg_123` with that known ID.

```bash
SECRET="$NOTIFICATION_EMAIL_WEBHOOK_SIGNING_SECRET"
TIMESTAMP="$(date +%s)"
BODY='{"provider_event_id":"local_evt_1","provider_message_id":"msg_123","type":"delivered","occurred_at":"2026-06-06T12:00:00Z"}'
SIGNATURE="$(printf '%s.%s' "$TIMESTAMP" "$BODY" | openssl dgst -sha256 -hmac "$SECRET" -hex | awk '{print $2}')"

curl -i \
  -X POST http://127.0.0.1:8081/internal/provider-webhooks/email \
  -H "Content-Type: application/json" \
  -H "X-Notification-Timestamp: $TIMESTAMP" \
  -H "X-Notification-Signature: sha256=$SIGNATURE" \
  --data "$BODY"
```

Expected first accepted response:

```text
HTTP/1.1 204 No Content
```

Post the same body again with a fresh valid timestamp/signature. It should also return `204`, but the delivery milestone counter must not increment twice.

Useful webhook response meanings:

| HTTP status | Meaning |
|---:|---|
| `204` | Accepted or known duplicate |
| `202` | Signature valid, but provider message did not match a delivery |
| `400` | Invalid/unsupported payload |
| `401` | Invalid or stale signature |
| `404` | Channel endpoint is not configured |
| `413` | Body exceeds configured limit |
| `422` | Open tracking is disabled |
| `503` | Database/analytics processing failed |

### Verify Webhook Result

```bash
curl -fsS http://127.0.0.1:8081/metrics | grep 'notification_provider_webhooks_total'
```

```bash
mongosh "$NOTIFICATION_MONGO_URI" --quiet --eval '
const d = db.getSiblingDB(process.env.NOTIFICATION_MONGO_DATABASE || "notification_db");
const events = process.env.NOTIFICATION_MONGO_PROVIDER_EVENTS_COLLECTION || "provider_events";
printjson(d.getCollection(events).find(
  { provider_event_id: "local_evt_1" },
  { _id: 0, provider: 1, provider_event_id: 1, delivery_id: 1, type: 1, channel: 1, occurred_at: 1, received_at: 1 }
).toArray());
'
```

### Prometheus And Grafana Operational Flow

1. Prometheus scrapes the private metrics route on port `8081`.
2. Grafana uses Prometheus as a datasource.
3. Create panels for sent/delivered/final-failed/opened event rates and webhook outcomes.
4. Keep `opened` clearly labeled as an observed signal, not guaranteed human reading.
5. Compare durable MongoDB milestone counts with Prometheus trends after restarts/incidents.

No committed PromQL dashboard or alert threshold exists, so production queries and alerts need monitoring-team review.

## 11. Common Errors & Fixes

Generic Go, `.env`, MongoDB, Docker, RabbitMQ, and provider setup errors are already covered in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: Common Setup Issues

TaskImplementation/{SERVICE_NAME}/task6_Dependency.md
Section: Common Errors & Fixes
```

Current-task issues:

| Error / symptom | Cause | Fix | Prevention |
|---|---|---|---|
| Transaction numbers are only allowed on a replica set member or mongos | MongoDB is standalone | Use a replica set/sharded cluster and update URI | Make transaction support an environment readiness check |
| `NOTIFICATION_PROVIDER_WEBHOOKS_ENABLED requires NOTIFICATION_METRICS_ENABLED` | Webhooks enabled without metrics | Enable metrics or disable webhooks | Validate env in CI/deployment |
| Webhooks require a signing secret for an enabled channel | No valid endpoint can be registered | Enable a runnable channel and set its secret | Configure channel and callback values together |
| Signing secret configured for disabled channel | Secret exists while channel flag is false | Remove secret or enable/configure channel | Avoid stale secrets in environment |
| Open tracking requires notification webhooks | Open tracking enabled alone | Enable required webhook chain or keep open tracking off | Default open tracking to false |
| Open tracking is not enabled for notification channel | SMS/push/WhatsApp-like open flag set true | Keep non-email open flags false | Validate supported capabilities |
| Metrics/webhook paths must be absolute or different | Invalid path or same route used twice | Use `/metrics` and `/internal/provider-webhooks` | Keep reviewed defaults |
| Analytics HTTP `address already in use` | Port `8081` occupied | Stop conflicting process or change address | Reserve/document ports |
| Local Prometheus cannot bind `9090` | Service gRPC already uses same host port | Map Prometheus to a different host port | Include port plan in local monitoring setup |
| Webhook returns `401` | Bad signature, body changed after signing, stale timestamp, or clock skew | Sign exact raw body and sync clocks | Monitor invalid-signature outcomes |
| Webhook returns `202` | No delivery matches provider, message ID, and channel | Verify outbound provider/message correlation | Integration-test provider mapping |
| Webhook returns `503` | Mongo transaction/write unavailable | Check replica-set health, URI, migration, and logs | Add readiness and transaction smoke checks |
| Callback returns `204` but no event counter increments again | Duplicate provider event or milestone already set | Expected idempotent behavior; inspect DB/metrics outcome | Use unique provider event IDs |
| Migration `007` `collMod` fails | Earlier migrations/collection missing or wrong DB/name | Apply prior migrations and load matching env | Track migration history |
| Analytics fields later stop validating | Older validator migration replayed after `007` | Restore latest approved validator migration | Never replay old migrations on live DB |
| `/metrics` reachable publicly | Listener/ingress too broad | Restrict network and proxy only webhook path publicly | Separate public/private routing controls |
| Full test suite fails in preference test | Existing fixture clock is earlier than stored create time | Fix the test fixture before declaring CI green | Keep injected test clocks monotonic |

## 12. Security & Best Practices

### Current-Task Security Rules

| Rule | Why |
|---|---|
| Store only normalized provider events | Raw callbacks may contain recipient data, content, or vendor secrets |
| Verify HMAC and replay window before DB mutation | Prevents forged/stale callback manipulation |
| Keep signing secrets in secret manager | `.env`, source, docs, and logs are not production secret stores |
| Restrict `/metrics` to internal monitoring | Metrics reveal operational behavior and labels |
| Route only webhook prefix through public HTTPS ingress | Metrics and callbacks currently share one listener |
| Keep MongoDB authenticated/TLS-protected in production | Provider events and delivery history are sensitive operational data |
| Keep open tracking off by default | Open tracking needs privacy/product/legal approval |
| Never treat `opened` as guaranteed human reading | Privacy proxies and image blocking make it approximate |
| Keep OTP opens rejected | Security notification tracking has a stricter privacy boundary |
| Keep metric labels bounded | User ID, delivery ID, message ID, arbitrary campaign ID, and raw failure text would create cardinality/privacy risk |
| Monitor invalid/unmatched callbacks | Spikes can indicate secret mismatch, provider issue, attack, or correlation bug |
| Back up before rollback | Down migration deletes analytics data |

### Task-Specific Best Practices

- Use an authenticated multi-node/managed replica set in production, not the local single-node example.
- Add a readiness check that proves MongoDB transaction support, not only basic connectivity.
- Treat `template_key` as a controlled vocabulary because it is a Prometheus label.
- Add retention policy review for `provider_events`; the current migration creates no TTL index.
- Build reconciliation between durable MongoDB milestones and in-memory Prometheus counters.
- Add provider-specific adapters that validate the vendor's native signature before normalizing callbacks.
- Keep provider event IDs stable and unique per provider.
- Alert on invalid signatures, unmatched events, webhook failures, callback lag, and sustained final-failure rate.
- Document how provider signing secrets rotate without callback downtime.
- Test out-of-order callbacks and duplicate provider retries.

## 13. Missing or Misconfigured Things

| Audit item | Current observation | Recommendation |
|---|---|---|
| Input-document status | `{TASK_FILE_NAME}` says source/migration/dependency were not implemented, but they now exist | Update the implementation document later so onboarding does not follow stale scope statements |
| Official transaction-capable local stack | Previous local Mongo example is standalone; no committed replica-set Compose exists | Add an approved one-command replica-set development stack |
| Migration history | Ordered scripts exist, but no applied-migration tracker/runner exists | Add one-time migration history before production |
| Monitoring deployment | No Prometheus scrape config, Grafana dashboard, or alert rules are committed | Add reviewed monitoring-as-code |
| Health/readiness | No dedicated endpoint reports Mongo transaction capability or analytics HTTP health | Add internal readiness/liveness endpoints |
| Provider event retention | `provider_events` has no TTL/archival policy | Define legal/support retention and add approved cleanup |
| Reconciliation | No scheduled Mongo-to-Prometheus reconciliation job exists | Add a periodic mismatch check |
| Vendor callback compatibility | Handler accepts a custom normalized payload/HMAC contract, not vendor-native contracts | Implement provider-specific verification/normalization adapters |
| Listener exposure | Metrics and potentially public webhooks share one HTTP listener | Use strict reverse-proxy routing or separate listeners in a reviewed change |
| Transport security | Analytics HTTP server has no built-in TLS/client auth | Keep behind private network/HTTPS ingress; do not expose directly |
| Open-consent enforcement | Runtime has a global email open flag and OTP rejection, but no per-delivery/user consent marker check in callback recording | Add approved per-delivery consent evidence before production open tracking |
| Durable metric consistency | Mongo transaction can commit before an in-memory metric increment is lost during crash | Use reconciliation or an approved durable metrics-outbox/exporter design |
| Send failure ambiguity | A provider may accept a send before analytics transaction fails, causing the request/worker path to report an error | Add reconciliation/idempotent resend protection and alert on analytics-write failures |
| Metrics cardinality | `template_key` is a metric label; uncontrolled dynamic keys can grow series count | Enforce controlled template keys and monitor cardinality |
| Full service container | No service Dockerfile was found | Add a production image before container deployment |
| Live integration tests | Unit tests exist, but no automated replica-set/webhook/Prometheus end-to-end test was found | Add container-backed integration tests |
| Full test suite | One unrelated preference fixture test currently fails | Fix it before considering service CI green |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Fresh Clone Setup | Same repository clone process |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Go Dependency System | Same Go install, download, build, and generic module troubleshooting |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Environment Variables / Credentials Placement | Same `.env` path, shell loading, and secret-handling rules |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | MongoDB Setup / MongoDB Migrations | Same base installation, URI concepts, and `mongosh` process; transaction change documented here |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | RabbitMQ / Mailpit / SMS Setup | Same optional event and outbound provider infrastructure |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Start The Service / Common Setup Issues | Same base run command and generic errors |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | Database Setup / Security & Best Practices | Same base delivery collection and safe-data rules |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | Database Setup | Same controlled templates used by analytics labels |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | Provider Setup / Security & Best Practices | Same OTP send flow and OTP secrecy boundary |
| `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md` | RabbitMQ / Running The Project | Same event-driven send flow |
| `TaskImplementation/{SERVICE_NAME}/task6_Dependency.md` | Retry/DLQ / Migration Order Warning | Same final-failure semantics and validator replay risk |
| `TaskImplementation/{SERVICE_NAME}/task7_Dependency.md` | Consent / Security & Best Practices | Same suppression and privacy context |

## 15. Final Checklist

- [ ] Previous dependency documentation reviewed
- [ ] Go `1.26.3` installed
- [ ] Existing modules downloaded and verified
- [ ] Prometheus Go client present through locked `go.mod`
- [ ] MongoDB uses a transaction-capable replica set or sharded cluster
- [ ] MongoDB URI points to the correct transaction-capable deployment
- [ ] Existing migrations `001` through `006` applied
- [ ] Migration `007` applied exactly once after earlier migrations
- [ ] `provider_events` validator and indexes verified
- [ ] Provider-event collection name matches config and migration environment
- [ ] Old validator migrations not replayed over migration `007`
- [ ] New analytics environment variables reviewed and loaded
- [ ] Metrics-only or webhook-enabled mode chosen intentionally
- [ ] Analytics HTTP listener binds only to an approved interface
- [ ] Port `8081` reachable from approved monitoring/webhook routes
- [ ] `/metrics` blocked from public traffic
- [ ] Signing secrets stored outside source control
- [ ] At least one enabled channel has a valid secret when webhooks are enabled
- [ ] Non-email open-tracking flags remain false
- [ ] Email open tracking remains off until privacy/product approval
- [ ] Task-focused tests pass
- [ ] Known full-suite preference test failure reviewed
- [ ] Service build passes
- [ ] Service starts with analytics HTTP log when enabled
- [ ] `/metrics` scrape verified
- [ ] One send records a durable `sent` or final `failed` event
- [ ] Signed callback fixture accepted
- [ ] Duplicate callback does not recount milestone
- [ ] Invalid/stale signature is rejected
- [ ] Mongo milestones/provider events contain no recipient/content/secrets
- [ ] Prometheus scrape and Grafana dashboard/alerts configured for production
- [ ] Provider-native callback adapter and HTTPS ingress reviewed
- [ ] Retention, reconciliation, and readiness gaps tracked
- [ ] No duplicate shared setup documentation added
