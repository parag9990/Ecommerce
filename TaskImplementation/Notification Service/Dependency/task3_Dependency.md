# Project Dependency & Setup Guide

## 1. Project Overview

### Variables Used

```text
SERVICE_NAME = "Notification Service"
TASK_FILE_NAME = "task3.md"
INPUT_FILE_PATH = TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}
OUTPUT_FILE_NAME = task3_Dependency.md
OUTPUT_FILE_PATH = TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}
```

This guide is generated from `INPUT_FILE_PATH` and saved at `OUTPUT_FILE_PATH`.
Original implementation file ko modify nahi kiya gaya.

### Task Scope

`{TASK_FILE_NAME}` ka goal template engine setup samjhana hai. Simple words me: MongoDB me stored active templates ko Go `text/template` ke through variables ke saath render karna.

Task-specific implementation already present in current repo:

```text
backend/services/notification-service/internal/domain/rendered_message.go
backend/services/notification-service/internal/usecase/template_spec.go
backend/services/notification-service/internal/usecase/render_template.go
backend/services/notification-service/internal/usecase/render_template_test.go
backend/services/notification-service/migrations/002_seed_renderable_templates.up.js
backend/services/notification-service/migrations/002_seed_renderable_templates.down.js
```

Task 3 does not introduce a standalone public `RenderTemplate` API. Renderer service ke andar use hota hai, mainly OTP and event notification flows ke through.

### Read This First

Follow previous dependency guides first. Ye file duplicate setup repeat nahi karegi.

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
```

Reused topics:

| Previous file | Section / Topic | Why reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Fresh Clone Setup | Same repo clone process |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Go Dependency System | Same Go module/workspace commands |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Environment Variables | Same `.env` location and loading method |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | MongoDB Setup | Same MongoDB install, Docker, URI, and verification |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | MongoDB Migrations | Same `mongosh` migration runner |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Start The Service | Same `go run ./cmd/server` startup |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | Database Setup | Same `notification_templates` collection and active-version lookup |

## 2. Tech Stack

Only new or task-specific items are explained here.

| Technology | Required for Task 3? | Status | Beginner explanation |
|---|---:|---|---|
| Go `text/template` | Yes | New task-specific engine, standard library | `text/template` Go ka built-in template engine hai. Isse `Hi {{.name}}` jaisa text runtime values ke saath render hota hai. Alag install nahi chahiye. |
| Go `bytes` | Yes | Standard library | Rendered output memory buffer me collect karne ke liye use hota hai. Alag dependency nahi. |
| Go `strings` | Yes | Standard library | Empty variables trim karne aur email subject line-break check karne ke liye. |
| MongoDB | Yes for real stored template lookup | Reused from Task 2 | Templates MongoDB ke `notification_templates` collection me stored hain. Setup already previous docs me hai. |
| MongoDB Go Driver v2 | Yes for repository lookup | Already present | Renderer khud DB driver import nahi karta; repository active template fetch karta hai. |
| `mongosh` | Recommended | Reused tool | Migration `002` run/verify karne ke liye useful hai. |
| gRPC | Indirect | Reused runtime | Service `:9090` par gRPC expose karta hai, but Task 3 ka direct RenderTemplate RPC available nahi hai. |
| RabbitMQ / Mailpit / SMTP / SMS gateway | No for pure Task 3 | Optional later flows | Rendering unit tests ke liye required nahi. OTP send/event flows test karne par providers required ho sakte hain. |
| Redis / Kafka / MySQL / PostgreSQL | No | Not used | Is task ke liye inka setup bilkul required nahi hai. |

## 3. Required Software

Task 3 ke liye beginner ko ye chahiye:

| Software | Required? | Setup source |
|---|---:|---|
| Go 1.26.3 | Yes, unit tests/build/run ke liye | `task1_Dependency.md` section `Go Dependency System` |
| MongoDB server | Yes, actual stored template lookup/startup verify karne ke liye | `task1_Dependency.md` section `MongoDB Setup` |
| `mongosh` | Recommended | `task1_Dependency.md` section `MongoDB Setup` |
| Docker | Optional | `task1_Dependency.md` section `MongoDB With Docker` |
| Git | Yes | `task1_Dependency.md` section `Fresh Clone Setup` |

No new OS package, no new CLI, and no external template library is introduced by `{TASK_FILE_NAME}`.

## 4. Dependency Management

Refer:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: Go Dependency System
```

Task 3 does not need any new `go get`.

Reason: renderer uses Go standard library imports:

```text
bytes
context
errors
fmt
reflect
strings
text/template
```

Existing module remains:

```text
backend/services/notification-service/go.mod
```

Task-specific dependency audit:

| Dependency | New in Task 3? | Why |
|---|---:|---|
| `text/template` | Yes as feature, no as install | Standard library renderer |
| `go.mongodb.org/mongo-driver/v2` | No | Already used by Task 2 repository |
| `google.golang.org/grpc` | No | Existing service transport |
| `github.com/rabbitmq/amqp091-go` | No | Later event/retry runtime, not pure rendering |
| `github.com/prometheus/client_golang` | No | Metrics runtime, not pure rendering |

Task-specific test command:

```bash
cd backend/services/notification-service
go test ./internal/usecase -run TestTemplateRenderer
```

If your checkout has a broken `backend/go.work` reference, run the service module directly:

```bash
cd backend/services/notification-service
GOWORK=off go test ./internal/usecase -run TestTemplateRenderer
```

Full service tests are still reused from previous setup:

```bash
cd backend/services/notification-service
go test ./...
```

## 5. Database Setup

### Reused MongoDB Setup

MongoDB installation, Docker command, connection string, and generic verification are already explained in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: MongoDB Setup
```

Collection design is already explained in:

```text
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
Section: Database Setup
```

### Task 3 Database Impact

Task 3 reuses the existing database and collection:

| Item | Value | Status |
|---|---|---|
| Database | `notification_db` | Reused |
| Collection | `notification_templates` | Reused |
| Main index | `{ template_key: 1, channel: 1, version: -1 }` | Reused from migration `001` |
| New task-specific migration | `002_seed_renderable_templates.up.js` | New for Task 3 |

### New Seed Migration

Task 3 adds renderable default template documents through:

```text
backend/services/notification-service/migrations/002_seed_renderable_templates.up.js
```

It seeds these template families:

| Template key | Channels seeded | Required variables |
|---|---|---|
| `otp_verification` | `email`, `sms` | `otp`, `expires_in_minutes` |
| `order_status_update` | `email`, `sms`, `push` | `name`, `order_id`, `status` |
| `payment_status_update` | `email`, `sms` | `name`, `order_id`, `payment_status`, `amount` |
| `promotional_offer` | `email`, `push` | `name`, `offer_title`, `coupon_code`, `valid_until` |

Beginner note: template body me Go syntax use hota hai:

```text
{{.name}}
{{.otp}}
{{.order_id}}
```

`{{name}}` wrong hai for current implementation. Dot missing hoga to parse/render issue aa sakta hai.

### Run Task 3 Migration

Use the migration runner already documented in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: MongoDB Migrations
```

If only Task 3 seed migration verify karni hai, run after migration `001`:

```bash
cd backend/services/notification-service
set -a
. ./.env
set +a
mongosh "$NOTIFICATION_MONGO_URI" migrations/002_seed_renderable_templates.up.js
```

Recommended for a fresh DB:

```bash
cd backend/services/notification-service
set -a
. ./.env
set +a
for file in migrations/*.up.js; do
  mongosh "$NOTIFICATION_MONGO_URI" "$file"
done
```

### Verify Seeded Templates

```bash
mongosh "$NOTIFICATION_MONGO_URI" --quiet --eval '
const d = db.getSiblingDB(process.env.NOTIFICATION_MONGO_DATABASE || "notification_db");
printjson(d.notification_templates.find(
  { template_key: { $in: ["otp_verification", "order_status_update", "payment_status_update", "promotional_offer"] } },
  { _id: 1, template_key: 1, channel: 1, status: 1, version: 1 }
).sort({ template_key: 1, channel: 1 }).toArray());
'
```

Expected high-level result: 9 active template documents from migration `002`.

### Rollback Task 3 Seed Data

Only if you intentionally want to remove seeded renderable templates:

```bash
cd backend/services/notification-service
set -a
. ./.env
set +a
mongosh "$NOTIFICATION_MONGO_URI" migrations/002_seed_renderable_templates.down.js
```

Warning: rollback se OTP/order/payment/promotional rendering DB lookup fail karega unless equivalent active templates already exist.

## 6. Redis / Queue / External Services

No new Redis, Kafka, RabbitMQ, NATS, MinIO, Elasticsearch, Firebase, Stripe, Twilio, SMTP, or OAuth setup is introduced by `{TASK_FILE_NAME}`.

Task-specific service matrix:

| Service | Required for renderer unit tests? | Required for DB-backed runtime? | Notes |
|---|---:|---:|---|
| MongoDB | No | Yes | Real service startup and active template lookup require it |
| RabbitMQ | No | No for Task 3 | Needed only when event consumer is enabled |
| Mailpit/SMTP | No | No for Task 3 | Needed only when actually sending email OTP |
| SMS HTTP gateway | No | No for Task 3 | Needed only when actually sending SMS OTP |
| Redis | No | No | Not used |
| Kafka | No | No | Not used by current implementation |

Hinglish note: Agar sirf template renderer test karna hai, `go test ./internal/usecase -run TestTemplateRenderer` enough hai. MongoDB tab chahiye jab stored templates DB se load karke service run/verify karni ho.

## 7. Environment Variables

No new environment variable is introduced by Task 3.

Reuse:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: Environment Variables
```

Task 3 relevant existing variables:

| Variable | Required? | Why relevant |
|---|---:|---|
| `NOTIFICATION_MONGO_URI` | Yes for DB-backed run | Migration and repository use this URI |
| `NOTIFICATION_MONGO_DATABASE` | Optional, defaults to `notification_db` | Migration `002` reads this |
| `NOTIFICATION_MONGO_TEMPLATES_COLLECTION` | Optional, defaults to `notification_templates` | Migration `002` inserts templates here |
| `NOTIFICATION_GRPC_ADDRESS` | Optional, defaults to `:9090` | Service startup port, reused |

Minimal `.env` for Task 3 DB verification:

```env
NOTIFICATION_MONGO_URI=mongodb://ecommerce_root:ecommerce_password@localhost:27017/notification_db?authSource=admin
NOTIFICATION_MONGO_DATABASE=notification_db
NOTIFICATION_MONGO_TEMPLATES_COLLECTION=notification_templates
```

Create or update this file:

```text
backend/services/notification-service/.env
```

Common mistakes:

| Mistake | Result | Fix |
|---|---|---|
| `.env` not loaded before `mongosh` | Migration uses default DB/collection unexpectedly | Follow `.env` loading command from previous docs |
| Custom template collection in `.env`, but query checks default collection | Seed appears missing | Query the same value as `NOTIFICATION_MONGO_TEMPLATES_COLLECTION` |
| Running `002` before `001` on a fresh DB | Validator/index baseline may be missing | Run migrations in order |

## 8. Docker Setup

No new Dockerfile, Docker Compose service, container, volume, or network is introduced by `{TASK_FILE_NAME}`.

Docker setup is reused from:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: MongoDB With Docker
Section: Docker Compose Example
```

Task 3 Docker impact:

| Docker item | Status |
|---|---|
| Service Dockerfile | No task-specific change |
| MongoDB container | Reused |
| RabbitMQ container | Not required for Task 3 |
| Mailpit container | Not required for Task 3 |
| New volume | None |
| New network | None |
| New health check | None |

## 9. Local Development Setup

### Step 1: Read Previous Dependency Documentation

Start with:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
```

They already cover clone, Go install, MongoDB install, `.env`, Docker, generic migrations, and generic errors.

### Step 2: Go To Project Directory

```bash
cd backend/services/notification-service
```

### Step 3: Install Only New Dependencies

No new install command.

If dependencies are missing from a fresh machine, reuse:

```bash
go mod download
```

### Step 4: Setup Only New Databases/Services

No new database service. Start MongoDB using previous docs.

### Step 5: Add Only New Or Changed Environment Variables

No new variables. Ensure existing Mongo variables are correct in:

```text
backend/services/notification-service/.env
```

### Step 6: Run Migrations

For Task 3, migration `002` must be applied after migration `001`.

Fresh DB recommended command:

```bash
cd backend/services/notification-service
set -a
. ./.env
set +a
for file in migrations/*.up.js; do
  mongosh "$NOTIFICATION_MONGO_URI" "$file"
done
```

### Step 7: Start Backend Service

Service startup is reused from:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: Start The Service
```

Minimal command:

```bash
cd backend/services/notification-service
set -a
. ./.env
set +a
go run ./cmd/server
```

Expected log:

```text
notification.grpc.started
```

### Step 8: Verify Task 3 Functionality

Task 3 has no standalone RenderTemplate RPC. Verify in two practical ways.

Unit-level verification:

```bash
cd backend/services/notification-service
go test ./internal/usecase -run TestTemplateRenderer
```

If Go complains about another module from `backend/go.work`, use:

```bash
cd backend/services/notification-service
GOWORK=off go test ./internal/usecase -run TestTemplateRenderer
```

Database seed verification:

```bash
mongosh "$NOTIFICATION_MONGO_URI" --quiet --eval '
const d = db.getSiblingDB(process.env.NOTIFICATION_MONGO_DATABASE || "notification_db");
print(d.notification_templates.countDocuments({
  template_key: { $in: ["otp_verification", "order_status_update", "payment_status_update", "promotional_offer"] },
  status: "active"
}));
'
```

Expected count after only migration `002`: `9`.

## 10. Running the Project

Running the full project is same as previous docs:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: Start The Service
```

Task 3-specific runtime checklist:

| Check | Expected |
|---|---|
| MongoDB reachable | Service does not fail with `ping notification mongo` |
| Migration `001` applied | `notification_templates` exists with validator/index |
| Migration `002` applied | 9 active renderable task templates exist |
| `NOTIFICATION_MONGO_TEMPLATES_COLLECTION` correct | Repository reads the same collection that migration seeded |
| Provider flags disabled for pure startup | Email/SMS/RabbitMQ not accidentally required |

Ports:

| Service | Port | Purpose | Status |
|---|---:|---|---|
| MongoDB | `27017` | Template storage lookup | Reused |
| Backend gRPC | `9090` | Existing service API | Reused |
| Analytics HTTP | `8081` | Metrics/webhooks when enabled | Reused optional |
| RabbitMQ AMQP | `5672` | Event/retry later tasks | Not Task 3 |
| RabbitMQ UI | `15672` | Broker admin UI | Not Task 3 |
| Mailpit SMTP | `1025` | Local email sending tests | Not Task 3 |
| Mailpit UI | `8025` | View local emails | Not Task 3 |

No new port conflict is introduced by `{TASK_FILE_NAME}`.

## 11. Common Errors & Fixes

Generic Go/Mongo/Docker/env errors are already covered in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: Common Setup Issues
```

Task 3-specific errors:

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `notification template not found` | Migration `002` not applied, wrong collection, or template inactive | Run/verify `002_seed_renderable_templates.up.js` | Always run migrations in order after fresh DB setup |
| `invalid notification render request: required variable ... is missing` | Caller did not pass required variable like `otp` or `order_id` | Pass every required variable for that template key | Check `template_spec.go` before adding a new caller |
| `unsupported notification template key` | Template key not in renderer allowlist | Add key to `rendered_message.go`, `template_spec.go`, migration seed, and tests | Keep code allowlist and DB seed in sync |
| `channel "sms" is not allowed for template "promotional_offer"` | Channel not allowlisted for that template family | Use allowed channel or update spec intentionally | Review channel matrix before publishing templates |
| `notification template rendering failed: parse body` | Stored template syntax invalid, commonly `{{.name` or unsupported format | Fix Mongo template body/subject syntax | Use Go `{{.variable}}` syntax and test before activation |
| `notification template rendering failed: render body` | Stored template references a variable not in approved context | Update template or allowlist | Do not publish templates referencing unknown variables |
| `email subject contains a line break` | Rendered subject has `\r` or `\n`, possible header injection risk | Clean input/status value before rendering | Never allow user-controlled line breaks in email subject variables |
| Count is not `9` after Task 3 seed | Only later migrations were checked, or custom collection is used | Query correct DB/collection and run migration `002` | Keep `.env` loaded in the same shell |
| `welcome_user` renders in tests but not after only Task 3 migration | That key is seeded by later event migration `004`, not Task 3 migration `002` | Run all migrations if testing later event templates | Do not mix Task 3-only setup with later-task template keys |
| `go: cannot load module ../auth-service listed in go.work file` | Workspace references a module whose `go.mod` is missing in this checkout | From service folder run `GOWORK=off go test ./internal/usecase -run TestTemplateRenderer`, or restore the missing module | Keep `backend/go.work` aligned with checked-out services |

## 12. Security & Best Practices

### Task 3 Security Rules

| Rule | Why it matters |
|---|---|
| Do not log rendered OTP body | OTP sensitive hota hai; logs me leak ho sakta hai |
| Keep template editing admin-controlled | End users ko template source edit karne dena unsafe hai |
| Use `missingkey=error` | Missing values silently `<no value>` banne se broken notification ja sakti hai |
| Filter extra variables | Caller accidentally `password` or `access_token` pass kare to renderer context me expose nahi hona chahiye |
| Reject email subject line breaks | Header injection style issues avoid hote hain |
| Use versioned templates | Wording change audit/rollback easy hota hai |
| Use `html/template` for future HTML body | HTML variables automatically escape honge |

### Current Implementation Notes

| Check | Current status |
|---|---|
| Extra variables filtered | Yes, `approvedVariables` copies only required variables |
| Missing variables rejected | Yes, blank/whitespace values rejected |
| Unknown channels rejected | Yes, channel validation exists |
| Unknown template keys rejected | Yes, allowlist exists |
| Active template selected | Yes, repository filters `status: active` and sorts version descending |
| Email subject line breaks rejected | Yes, renderer checks rendered subject |
| Rendered body logging | No direct renderer log found |

### Best Practices For Future Template Changes

| Practice | Beginner-friendly reason |
|---|---|
| Add tests before adding a new template key | Code allowlist, DB seed, and channel rules sync me rahenge |
| Publish new template wording as new `version` | Old version rollback ke liye available rahega |
| Keep variable names snake_case | Mongo template, Go map, and event payload readable rahenge |
| Keep SMS short | SMS length limit/cost issue avoid hoga |
| Keep OTP template minimal | Sensitive data and phishing risk kam hota hai |
| Verify seed in Mongo before service test | Template missing error ka root cause jaldi milta hai |

## 13. Missing or Misconfigured Things

Task-specific audit:

| Item | Status | Recommendation |
|---|---|---|
| Direct `RenderTemplate` API | Not exposed | Fine for Task 3; verify through unit tests or caller flows |
| Official Docker Compose file in repo | Not found | Previous docs provide example; add official compose later if team wants one command setup |
| `backend/go.work` dependency on auth service | Current checkout may not include `backend/services/auth-service/go.mod` | Use `GOWORK=off` for this service module until workspace modules are complete |
| Task 3 env variables | None | No action needed |
| External template engine | Not used | Good; standard library is enough |
| HTML body support | Not part of current task | Use `html/template` if HTML body becomes first-class |
| Template seed vs code allowlist | Task 3 keys are seeded in migration `002`; later keys in migration `004` | Run all migrations when testing current full service |
| Production secret handling | Reused concern | Follow `task1_Dependency.md` section `Credentials Placement` |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Fresh Clone Setup | Same repository onboarding |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Go Dependency System | Same Go version, `go.mod`, `go.sum`, and test/build commands |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Environment Variables | Same `.env` path and shell loading process |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | MongoDB Setup | Same database installation, Docker setup, URI, and health checks |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | MongoDB Migrations | Same `mongosh` execution pattern |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Docker Compose Example | Same optional local infra pattern |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Start The Service | Same `go run ./cmd/server` flow |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Common Setup Issues | Same generic Go/Mongo/Docker/env troubleshooting |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | Database Setup | Same `notification_templates` collection and versioned active lookup |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | Security & Best Practices | Same Mongo template storage and versioning guidance |

## 15. Final Checklist

- [ ] Previous dependency documentation checked
- [ ] No duplicate clone/Go/Mongo/Docker setup copied
- [ ] Confirmed no new third-party Go dependency is required
- [ ] Confirmed `text/template` is standard library
- [ ] MongoDB is running if DB-backed verification is needed
- [ ] `.env` is created/loaded from `backend/services/notification-service/.env`
- [ ] Migration `001_create_notification_collections.up.js` applied first
- [ ] Migration `002_seed_renderable_templates.up.js` applied for Task 3 templates
- [ ] Seeded template count verified as `9` active Task 3 templates
- [ ] Renderer unit tests pass
- [ ] Backend service starts if full runtime verification is needed
- [ ] No RabbitMQ/Mailpit/SMS gateway enabled unless testing later sending flows
- [ ] Rendered OTP/body content is not logged
- [ ] Template variables use Go dot syntax like `{{.otp}}`
- [ ] No duplicate setup documentation added
