# Project Dependency & Setup Guide

## 1. Project Overview

### Variables Used

```text
SERVICE_NAME = "Notification Service"
TASK_FILE_NAME = "task4.md"
INPUT_FILE_PATH = TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}
OUTPUT_FILE_NAME = {TASK_FILE_NAME without ".md"}_Dependency.md
OUTPUT_FILE_PATH = TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}
```

This guide is generated from `INPUT_FILE_PATH` and saved at `OUTPUT_FILE_PATH`.
Original implementation file ko modify nahi kiya gaya.

### Current Task Scope

`{TASK_FILE_NAME}` ka goal Auth Service ke OTP request ke liye internal `SendOTP` flow support karna hai. Simple Hinglish me: Auth Service OTP generate/hash/verify karega, aur `{SERVICE_NAME}` sirf short-lived plaintext OTP ko email ya SMS provider tak deliver karega.

Important boundary:

| Area | Owner |
|---|---|
| OTP generate, hash, expire, verify, resend limit | Auth Service |
| OTP email/SMS render and dispatch | `{SERVICE_NAME}` |
| OTP plaintext persistence | Not allowed |
| Public `/auth/otp/send` API | Auth/Gateway side |
| Internal `SendOTP` gRPC method | `{SERVICE_NAME}` |

Task-specific implementation files already present in current repo:

```text
proto/ecommerce/notification/v1/notification.proto
backend/services/notification-service/api/notification/v1/notification.pb.go
backend/services/notification-service/api/notification/v1/notification_grpc.pb.go
backend/services/notification-service/internal/domain/otp_delivery.go
backend/services/notification-service/internal/usecase/send_otp.go
backend/services/notification-service/internal/transport/grpc/notification_handler.go
backend/services/notification-service/internal/provider/email_provider.go
backend/services/notification-service/internal/provider/sms_provider.go
backend/services/notification-service/internal/provider/smtp_client.go
backend/services/notification-service/internal/provider/http_sms_client.go
backend/services/notification-service/migrations/003_allow_optional_user_delivery_for_otp.up.js
```

### Read Previous Dependency Docs First

Do not repeat common setup. Follow these first:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
```

Reused topics:

| Previous file | Section / Topic | Why reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Fresh Clone Setup | Same repository clone flow |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Go Dependency System | Same Go version, module, and common commands |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Environment Variables | Same `.env` file and shell loading process |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | MongoDB Setup | Same MongoDB install, Docker, URI, and health checks |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Email Provider Setup With Mailpit | Same local SMTP testing setup |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | SMS Provider Setup | Same HTTP SMS gateway contract |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Start The Service | Same gRPC startup command |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | Database Setup | Same `notification_deliveries` collection |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | Task 3 Database Impact | Same `otp_verification` templates |

## 2. Tech Stack

Only new or current-task-specific items are explained here.

| Technology | Required for current task? | Status | Beginner explanation |
|---|---:|---|---|
| gRPC + Protobuf | Yes | Existing dependency, current-task RPC | gRPC internal services ke beech fast typed call ke liye use hota hai. `SendOTP` Auth Service se receive hota hai. |
| Go 1.26.3 | Yes | Reused | Service Go me written hai. Setup already previous guide me covered hai. |
| Go standard library `crypto/rand` | Yes | No install needed | Secure random delivery id generate karne ke liye use hota hai. |
| Go standard library `regexp` and `strings` | Yes | No install needed | OTP digits, purpose, target, and safe text validation ke liye. |
| MongoDB | Yes for runtime delivery trace | Reused with one schema change | OTP-free delivery metadata store hota hai. OTP/body/raw target store nahi hote. |
| SMTP email provider | Required for real email OTP smoke | Reused provider setup | Email OTP send karne ke liye SMTP adapter use hota hai. Local me Mailpit safest hai. |
| HTTP SMS gateway | Required for real SMS OTP smoke | Reused provider setup | SMS OTP send karne ke liye HTTP endpoint `POST {"to","text"}` accept karta hai. |
| Analytics recorder | Runtime dependency in current code | Reused current service infrastructure | Provider accepted/failed event ko sanitized lifecycle event me record karta hai. |
| RabbitMQ | No for direct OTP send | Not required | Event consumer/retry tasks ke liye hai, direct OTP RPC ke liye nahi. |
| Redis/Kafka/MySQL/PostgreSQL | No | Not used by this service task | OTP state Auth Service side ho sakta hai; yahan setup nahi chahiye. |

## 3. Required Software

Current task ke liye fresh machine par ye software enough hai:

| Software | Required? | Setup source |
|---|---:|---|
| Git | Yes | `task1_Dependency.md` section `Fresh Clone Setup` |
| Go 1.26.3 | Yes | `task1_Dependency.md` section `Go Dependency System` |
| MongoDB server | Yes for service startup and delivery trace | `task1_Dependency.md` section `MongoDB Setup` |
| `mongosh` | Recommended | Migrations and DB verification |
| Mailpit | Optional | Required only for local email OTP smoke |
| Local/fake SMS HTTP endpoint | Optional | Required only for local SMS OTP smoke |
| Docker | Optional | Useful for MongoDB/Mailpit/local infra |
| `grpcurl` | Optional | Useful for manual `SendOTP` smoke test |
| `protoc` / Go protobuf plugins | Optional | Needed only if proto source changes and generated Go files must be regenerated |

No new operating-system package is introduced only by `{TASK_FILE_NAME}`. Mailpit/SMS setup was already documented earlier; this file only tells when to enable them for OTP testing.

## 4. Dependency Management

Refer:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: Go Dependency System
```

Current task does not require a new `go get`. The module already contains the needed libraries:

```text
backend/services/notification-service/go.mod
```

Relevant existing dependencies:

| Dependency | Version in current repo | Why relevant |
|---|---|---|
| `google.golang.org/grpc` | `v1.81.1` | `SendOTP` gRPC server/client types |
| `google.golang.org/protobuf` | `v1.36.11` | Generated protobuf message support |
| `go.mongodb.org/mongo-driver/v2` | `v2.6.0` | Delivery record insert and lookup |
| `github.com/prometheus/client_golang` | `v1.23.2` | Optional metrics observer in current full service |

Standard library imports used by current-task implementation:

```text
context
crypto/rand
encoding/hex
errors
fmt
log/slog
regexp
strconv
strings
time
```

Focused verification commands:

```bash
cd backend/services/notification-service
GOWORK=off go test ./internal/domain -run TestSendOTP
GOWORK=off go test ./internal/usecase -run TestSendOTP
GOWORK=off go test ./internal/transport/grpc -run TestNotificationHandlerSendOTP
GOWORK=off go test ./internal/config
```

Why `GOWORK=off`? Current `backend/go.work` also references `auth-service`. Agar checkout me woh module present nahi hai, workspace-level Go commands fail ho sakte hain. Service folder se `GOWORK=off` focused module test ko clean rakhta hai.

## 5. Database Setup

### Reused MongoDB Setup

MongoDB installation, Docker command, connection string, and generic health check are already explained in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: MongoDB Setup
```

Collection design is already explained in:

```text
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
Section: Database Setup
```

### Current Task Database Impact

Current task reuses:

| Item | Value | Status |
|---|---|---|
| Database | `notification_db` | Reused |
| Template collection | `notification_templates` | Reused from Task 2/3 |
| Delivery collection | `notification_deliveries` | Reused with one validator change |
| OTP template key | `otp_verification` | Reused from Task 3 seed |
| New migration | `003_allow_optional_user_delivery_for_otp.up.js` | New for current task |

### Why Migration `003` Is Needed

Signup OTPs can happen before user account creation. Isliye OTP delivery record me `user_id` optional hona chahiye only when:

```text
template_key = otp_verification
```

For all other delivery types, `user_id` still required hai.

Migration file:

```text
backend/services/notification-service/migrations/003_allow_optional_user_delivery_for_otp.up.js
```

Rollback file:

```text
backend/services/notification-service/migrations/003_allow_optional_user_delivery_for_otp.down.js
```

### Safe OTP Delivery Record

Allowed safe metadata:

| Field | Stored? | Notes |
|---|---:|---|
| `delivery_id` / `_id` | Yes | Internal trace id |
| `channel` | Yes | `email` or `sms` |
| `template_key` | Yes | `otp_verification` |
| `status` | Yes | `accepted`, `rejected`, or `failed` |
| `provider` | Yes | Example `smtp` or `http_gateway` |
| `provider_message_id` | Yes if provider returns it | Safe provider reference |
| `payload.challenge_id` | Yes | Auth challenge correlation |
| `payload.purpose` | Yes | Example `login`, `signup` |
| `payload.correlation_id` | Optional | Request trace |
| `otp` | No | Plain OTP secret |
| rendered email/SMS body | No | Contains OTP |
| raw email/phone target | No | PII |

### Run Required Migrations

Fresh DB minimum order for current task:

```bash
cd backend/services/notification-service
set -a
. ./.env
set +a
mongosh "$NOTIFICATION_MONGO_URI" migrations/001_create_notification_collections.up.js
mongosh "$NOTIFICATION_MONGO_URI" migrations/002_seed_renderable_templates.up.js
mongosh "$NOTIFICATION_MONGO_URI" migrations/003_allow_optional_user_delivery_for_otp.up.js
```

For the current full service in this repo, run all `.up.js` migrations in order because later code also includes event, retry, preference, and analytics support:

```bash
cd backend/services/notification-service
set -a
. ./.env
set +a
for file in migrations/*.up.js; do
  mongosh "$NOTIFICATION_MONGO_URI" "$file"
done
```

### Verify OTP Template And Validator

Verify the email/SMS OTP templates from migration `002`:

```bash
mongosh "$NOTIFICATION_MONGO_URI" --quiet --eval '
const d = db.getSiblingDB(process.env.NOTIFICATION_MONGO_DATABASE || "notification_db");
printjson(d.notification_templates.find(
  { template_key: "otp_verification", status: "active" },
  { _id: 1, channel: 1, version: 1, subject: 1 }
).sort({ channel: 1 }).toArray());
'
```

Expected high-level result: one active email template and one active SMS template.

Verify migration `003` validator shape:

```bash
mongosh "$NOTIFICATION_MONGO_URI" --quiet --eval '
const d = db.getSiblingDB(process.env.NOTIFICATION_MONGO_DATABASE || "notification_db");
const name = process.env.NOTIFICATION_MONGO_DELIVERIES_COLLECTION || "notification_deliveries";
const info = d.getCollectionInfos({ name })[0];
printjson(info.options.validator.$jsonSchema.oneOf);
'
```

Expected high-level result: one branch allows `template_key: "otp_verification"` without `user_id`, and the other branch requires `user_id` for non-OTP deliveries.

## 6. Redis / Queue / External Services

No new Redis, Kafka, RabbitMQ, NATS, MinIO, Elasticsearch, Firebase, Stripe, Twilio, or OAuth setup is introduced by `{TASK_FILE_NAME}`.

Current-task service matrix:

| Service | Required? | Status | What to do |
|---|---:|---|---|
| MongoDB | Yes for runtime | Reused | Follow previous Mongo setup and run migration `003` |
| SMTP/Mailpit | Only for email OTP smoke | Reused | Follow `task1_Dependency.md` email provider section |
| SMS HTTP gateway | Only for SMS OTP smoke | Reused | Follow `task1_Dependency.md` SMS provider section |
| Auth Service | Integration dependency | External to this setup guide | Public OTP request and OTP hash storage belong there |
| RabbitMQ | No for direct OTP | Later task | Keep disabled unless testing event/retry flow |
| Provider webhooks/analytics HTTP | No for direct OTP smoke | Later/current full-service optional | Keep disabled unless testing provider callbacks |

Hinglish note: direct OTP RPC ko RabbitMQ ki zarurat nahi. Email/SMS provider zarur chahiye agar real message send karna hai. Unit tests fake providers use karte hain, so unke liye Mailpit/SMS gateway bhi nahi chahiye.

## 7. Environment Variables

Complete `.env` loading process is already documented in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: Environment Variables
```

Current task introduces no brand-new env variable names. It uses existing email/SMS/Mongo/gRPC variables.

Create or update:

```text
backend/services/notification-service/.env
```

### Required Reused Variables

| Variable | Required? | Purpose |
|---|---:|---|
| `NOTIFICATION_MONGO_URI` | Yes | Mongo connection for template lookup and delivery trace |
| `NOTIFICATION_MONGO_DATABASE` | Optional | Defaults to `notification_db` |
| `NOTIFICATION_MONGO_TEMPLATES_COLLECTION` | Optional | Defaults to `notification_templates` |
| `NOTIFICATION_MONGO_DELIVERIES_COLLECTION` | Optional | Defaults to `notification_deliveries` |
| `NOTIFICATION_MONGO_PROVIDER_EVENTS_COLLECTION` | Required by current full service config | Defaults to `provider_events` |
| `NOTIFICATION_GRPC_ADDRESS` | Optional | Defaults to `:9090` |

### Email OTP Smoke Values

Use only when testing email OTP delivery:

```env
NOTIFICATION_EMAIL_ENABLED=true
NOTIFICATION_EMAIL_PROVIDER=smtp
NOTIFICATION_EMAIL_SMTP_HOST=localhost
NOTIFICATION_EMAIL_SMTP_PORT=1025
NOTIFICATION_EMAIL_FROM=no-reply@example.com
NOTIFICATION_EMAIL_SMTP_TLS_MODE=none
NOTIFICATION_EMAIL_SMTP_USERNAME=
NOTIFICATION_EMAIL_SMTP_PASSWORD=
NOTIFICATION_EMAIL_TIMEOUT=5s
```

Security note: production SMTP credentials require `tls` or `starttls`. Do not use real username/password in committed files.

### SMS OTP Smoke Values

Use only when testing SMS OTP delivery:

```env
NOTIFICATION_SMS_ENABLED=true
NOTIFICATION_SMS_PROVIDER=http_gateway
NOTIFICATION_SMS_HTTP_ENDPOINT=http://localhost:3001/send
NOTIFICATION_SMS_HTTP_BEARER_TOKEN=
NOTIFICATION_SMS_TIMEOUT=5s
```

Production rule: non-local SMS endpoint must be HTTPS and must have `NOTIFICATION_SMS_HTTP_BEARER_TOKEN`.

### Common Env Mistakes

| Mistake | Result | Fix |
|---|---|---|
| Email/SMS enabled but provider name blank | Config validation fails | Set `NOTIFICATION_EMAIL_PROVIDER` or `NOTIFICATION_SMS_PROVIDER` |
| Email enabled but SMTP host/port/from missing | Service fails before gRPC starts | Fill SMTP values or disable email |
| SMTP username set without password | Config validation fails | Set both or leave both blank |
| SMTP credentials with TLS mode `none` | Config validation fails | Use `tls` or `starttls` |
| SMS enabled with blank endpoint | Config validation fails | Set `NOTIFICATION_SMS_HTTP_ENDPOINT` |
| SMS production endpoint uses HTTP | Config validation fails | Use HTTPS outside localhost |
| `.env` not loaded | Defaults may disable providers | Load `.env` before `go run` |

## 8. Docker Setup

No new service Dockerfile, Docker Compose service, volume, or network is introduced by `{TASK_FILE_NAME}`.

Reuse:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: MongoDB With Docker
Section: Email Provider Setup With Mailpit
Section: SMS Provider Setup
Section: Docker Compose Example
```

Docker impact:

| Docker item | Status |
|---|---|
| MongoDB container | Reused |
| Mailpit container | Reused, only for email OTP smoke |
| SMS fake gateway container | Not provided by repo |
| RabbitMQ container | Not required for direct OTP |
| New volume | None |
| New network | None |
| New health check | None |

### Ports & Networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Backend gRPC | `9090` | Internal `SendOTP` RPC | Reused |
| MongoDB | `27017` | Templates and delivery trace | Reused |
| Mailpit SMTP | `1025` | Local email OTP capture | Reused optional |
| Mailpit UI | `8025` | View captured email OTP | Reused optional |
| SMS fake gateway | `3001` example | Local SMS HTTP smoke | Optional/new only if you create one |
| Analytics HTTP | `8081` | Metrics/webhooks | Reused optional, not needed for direct OTP |
| RabbitMQ AMQP | `5672` | Event/retry flow | Not required |

Port conflict guidance for current task:

| Symptom | Likely cause | Fix |
|---|---|---|
| gRPC cannot bind `:9090` | Another service instance running | Stop old process or set `NOTIFICATION_GRPC_ADDRESS=:9091` |
| Email smoke sends nothing | Mailpit not running or wrong SMTP port | Check Mailpit container and `NOTIFICATION_EMAIL_SMTP_PORT=1025` |
| SMS smoke fails config validation | Endpoint blank or non-local HTTP | Set loopback HTTP for dev, HTTPS + token for production |
| Mongo validator rejects OTP delivery without user | Migration `003` not applied | Run migrations in order |

## 9. Local Development Setup

### Step 1: Read Previous Dependency Documentation

Start with:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
```

### Step 2: Go To Project Directory

Documentation folder:

```bash
cd "TaskImplementation/{SERVICE_NAME}"
```

Runnable backend service folder:

```bash
cd backend/services/notification-service
```

### Step 3: Install Only New Dependencies

No new Go dependency install is needed.

Fresh clone reminder:

```bash
cd backend/services/notification-service
GOWORK=off go mod download
```

### Step 4: Setup Only New Databases/Services

No new database service. Start MongoDB from previous docs.

Optional provider services:

| Need | Setup |
|---|---|
| Email OTP smoke | Start Mailpit from `task1_Dependency.md` |
| SMS OTP smoke | Start a local HTTP test endpoint or use provider sandbox credentials |

### Step 5: Add Only New Or Changed Environment Variables

No new variable names. For OTP smoke, enable only the channel you are testing:

```env
NOTIFICATION_EMAIL_ENABLED=true
NOTIFICATION_SMS_ENABLED=false
```

or:

```env
NOTIFICATION_EMAIL_ENABLED=false
NOTIFICATION_SMS_ENABLED=true
```

Beginner note: dono channels ek saath enable karna optional hai. Agar SMS endpoint ready nahi hai, SMS disabled rakho.

### Step 6: Run Migrations

Minimum current-task migrations:

```bash
cd backend/services/notification-service
set -a
. ./.env
set +a
mongosh "$NOTIFICATION_MONGO_URI" migrations/001_create_notification_collections.up.js
mongosh "$NOTIFICATION_MONGO_URI" migrations/002_seed_renderable_templates.up.js
mongosh "$NOTIFICATION_MONGO_URI" migrations/003_allow_optional_user_delivery_for_otp.up.js
```

Recommended for current full repo:

```bash
for file in migrations/*.up.js; do
  mongosh "$NOTIFICATION_MONGO_URI" "$file"
done
```

### Step 7: Start Backend Service

Startup process is reused from:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: Start The Service
```

Short command:

```bash
cd backend/services/notification-service
set -a
. ./.env
set +a
GOWORK=off go run ./cmd/server
```

Expected log:

```text
notification.grpc.started
```

### Step 8: Verify Current Task Functionality

Unit-level verification:

```bash
cd backend/services/notification-service
GOWORK=off go test ./internal/domain -run TestSendOTP
GOWORK=off go test ./internal/usecase -run TestSendOTP
GOWORK=off go test ./internal/transport/grpc -run TestNotificationHandlerSendOTP
```

Optional manual gRPC email smoke with `grpcurl`, after MongoDB, migration `003`, service startup, and Mailpit are ready:

```bash
grpcurl -plaintext \
  -import-path proto \
  -proto proto/ecommerce/notification/v1/notification.proto \
  -d '{
    "challenge_id": "challenge_local_001",
    "channel": "OTP_CHANNEL_EMAIL",
    "target": "person@example.com",
    "otp": "482991",
    "expires_in_minutes": 5,
    "purpose": "login",
    "correlation_id": "local_request_001"
  }' \
  localhost:9090 ecommerce.notification.v1.NotificationService/SendOTP
```

Expected high-level result:

```json
{
  "delivery_id": "delivery_otp_...",
  "status": "accepted"
}
```

Then verify Mongo has no OTP/body/raw target:

```bash
mongosh "$NOTIFICATION_MONGO_URI" --quiet --eval '
const d = db.getSiblingDB(process.env.NOTIFICATION_MONGO_DATABASE || "notification_db");
printjson(d.notification_deliveries.find(
  { template_key: "otp_verification" },
  { _id: 1, user_id: 1, channel: 1, status: 1, provider: 1, provider_message_id: 1, payload: 1 }
).sort({ created_at: -1 }).limit(3).toArray());
'
```

## 10. Running the Project

For normal run, follow previous startup docs and only add OTP-specific provider settings when needed.

Runtime flow:

| Step | What happens |
|---|---|
| 1 | Service loads `.env` |
| 2 | Service connects to MongoDB |
| 3 | Template renderer reads `otp_verification` template |
| 4 | Provider registry resolves email or SMS |
| 5 | Provider sends OTP message |
| 6 | Mongo stores safe delivery metadata |
| 7 | Handler returns `delivery_id` and `status` |

Direct OTP prerequisites:

| Prerequisite | Email OTP | SMS OTP |
|---|---:|---:|
| Mongo running | Yes | Yes |
| Migration `001` | Yes | Yes |
| Migration `002` | Yes | Yes |
| Migration `003` | Yes | Yes |
| Email env enabled | Yes | No |
| SMTP/Mailpit running | Yes | No |
| SMS env enabled | No | Yes |
| SMS endpoint running | No | Yes |

Auth Service integration note: real user-facing OTP flow should go through Auth Service. Direct gRPC smoke is only for local backend verification.

## 11. Common Errors & Fixes

Generic Go/Mongo/Docker/env errors are already covered in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: Common Setup Issues
```

Current-task-specific issues:

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `invalid OTP delivery request` | Missing challenge id, unsupported purpose, invalid OTP, missing target, or wrong channel | Send `email`/`sms`, 6+ digit OTP, valid purpose like `login` | Validate Auth request before calling Notification |
| `unsupported OTP channel` | Push/WhatsApp-like channel used for OTP | Use email or SMS only | Keep Auth channel mapping explicit |
| `notification template not found` | Migration `002` missing or template inactive | Run seed migration and verify `otp_verification` templates | Run migrations in order on fresh DB |
| `email subject and text body are required` | Email template missing subject/body | Fix active email OTP template | Do not publish incomplete email templates |
| `SMS recipient must be in E.164 format` | Phone target not like `+14155550100` | Normalize phone number before calling `SendOTP` | Store phone numbers in E.164 format |
| `notification channel disabled` | Email/SMS flag is false | Enable channel in `.env` | Enable only providers you are testing |
| `notification provider not configured` | Enabled provider name does not match registered provider setup | Set provider name and required provider env | Keep `.env` examples close to actual provider wiring |
| `OTP delivery is not configured` | Template/provider config failed | Check migrations, provider env, channel flags | Run config tests after env changes |
| `OTP delivery provider unavailable` | SMTP/SMS endpoint down, timeout, or raw provider failure | Start Mailpit/SMS endpoint and retry | Add local health checks for smoke setup |
| Delivery insert rejected by Mongo | Migration `003` missing or collection validator old | Apply migration `003` | Run migrations after DB reset |
| Error text includes no raw provider detail | By design | Check provider logs separately, safely | Never surface OTP or raw target in user errors |
| `go: cannot load module ../auth-service listed in go.work file` | Workspace references missing auth module | Use `GOWORK=off` from service folder | Keep workspace modules checked out together |

## 12. Security & Best Practices

### Current Task Security Rules

| Rule | Beginner-friendly reason |
|---|---|
| Do not store plaintext OTP | OTP leak ho gaya to account takeover risk hota hai |
| Do not store rendered body | Rendered body me OTP hota hai |
| Do not store raw email/phone in delivery payload | PII minimize karna zaruri hai |
| Store only challenge/correlation ids | Traceability milti hai without secret exposure |
| Keep OTP verification in Auth Service | `{SERVICE_NAME}` ko auth state owner nahi banana |
| Only email/SMS allowed for OTP | Requirement narrow hai; extra channels accidentally security policy bypass kar sakte hain |
| Use provider timeouts | Auth request indefinitely hang nahi hona chahiye |
| Use HTTPS + bearer token for production SMS | SMS provider request secret and PII protect hote hain |
| Use TLS for production SMTP credentials | SMTP username/password plaintext network me nahi jaane chahiye |
| Keep `accepted` separate from `delivered` | Provider accept ka matlab user received/read nahi hota |

### Current Implementation Audit

| Audit item | Current observation | Recommendation |
|---|---|---|
| OTP request validation | `SendOTPRequest.Validate()` checks challenge, channel, target, digits, expiry, and purpose | Keep Auth and Notification validation both active |
| Recipient validation | Email uses Go mail parser; SMS requires E.164 | Normalize recipient before internal call |
| Sensitive persistence | Tests assert OTP/body/raw target absent from delivery record | Keep this test pattern for future changes |
| gRPC error sanitization | Handler maps errors to generic messages | Do not include raw provider error in client response |
| Provider secrets | `.env` has placeholders/empty secret fields | Use secret manager in production |
| OTP retry | Domain forbids durable retry fields on OTP deliveries | Let Auth own resend/cooldown policy |
| Analytics/open tracking | OTP opens are blocked in analytics recorder | Do not enable open tracking semantics for OTP |
| Internal transport security | Plain local gRPC is okay for dev only | Use mTLS/service mesh or private network in production |

### Best Practices For Future Changes

| Practice | Why |
|---|---|
| Add tests for every new OTP purpose | Purpose allowlist and Auth mapping stay in sync |
| Keep OTP templates short and direct | User clarity and SMS cost improve |
| Never log `req.OTP`, target, or rendered body | Logs are widely copied in debugging systems |
| Make provider errors category-based | User-facing and Auth-facing errors stay safe |
| Keep migration `003` in all environments | Signup OTP without user id depends on it |
| Keep local Mailpit separate from production SMTP | Accidental real email sends avoid hote hain |

## 13. Missing or Misconfigured Things

| Gap / Risk | Impact | Suggested fix |
|---|---|---|
| No committed service-specific Docker Compose file | Beginners rely on referenced commands/examples | Add official local compose later if infra folder is finalized |
| `backend/go.work` references missing `auth-service` module in this checkout | Workspace-level Go commands can fail | Use `GOWORK=off` for focused module work or add the missing module |
| `NOTIFICATION_EMAIL_ENABLED=false` by default | Direct email OTP smoke returns configuration error | Enable email only when Mailpit/SMTP is ready |
| `NOTIFICATION_SMS_HTTP_ENDPOINT` blank by default | SMS channel cannot start when enabled | Set local fake gateway or production HTTPS provider endpoint |
| Auth Service implementation not present in this service folder | Full public OTP flow cannot be tested end-to-end here | Use direct gRPC smoke or add Auth Service integration environment |
| Generated proto code can become stale if proto changes | Handler/client types mismatch | Regenerate generated Go files whenever `proto/.../notification.proto` changes |
| Current full service includes later analytics/preference code | Running only migrations `001-003` may not be enough for all current features | For current full service, run all migrations in order |
| No local fake SMS server committed | SMS smoke needs external/manual test endpoint | Add a small dev-only SMS stub later if team wants one-command onboarding |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Fresh Clone Setup | Same repo clone and root setup |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Go Dependency System | Same Go module and test/build commands |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Environment Variables | Same `.env` path and loading method |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | MongoDB Setup | Same MongoDB install, Docker, URI, and verification |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | MongoDB Migrations | Same `mongosh` execution pattern |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Email Provider Setup With Mailpit | Same local SMTP setup for email OTP smoke |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | SMS Provider Setup | Same HTTP SMS gateway contract |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Docker Compose Example | Same optional local infra pattern |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Start The Service | Same `go run ./cmd/server` startup |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Common Setup Issues | Same generic Go/Mongo/Docker/env troubleshooting |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | Database Setup | Same `notification_deliveries` collection |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | Security & Best Practices | Same sensitive payload rules |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | Database Setup | Same `otp_verification` template seed |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | Common Errors & Fixes | Same template missing/rendering errors |

## 15. Final Checklist

- [ ] Previous dependency documentation checked
- [ ] No duplicate clone/Go/Mongo/Docker setup copied unnecessarily
- [ ] Confirmed no new `go get` is required
- [ ] MongoDB is running
- [ ] `.env` is created and loaded from `backend/services/notification-service/.env`
- [ ] Migration `001_create_notification_collections.up.js` applied
- [ ] Migration `002_seed_renderable_templates.up.js` applied
- [ ] Migration `003_allow_optional_user_delivery_for_otp.up.js` applied
- [ ] `otp_verification` email and SMS templates verified
- [ ] Email provider enabled only if Mailpit/SMTP is ready
- [ ] SMS provider enabled only if HTTP gateway endpoint is ready
- [ ] Focused OTP domain/usecase/gRPC tests pass
- [ ] Backend service starts and logs `notification.grpc.started`
- [ ] Direct gRPC smoke returns `accepted` when provider is available
- [ ] Mongo delivery record has no OTP, rendered body, or raw target
- [ ] Logs checked for absence of OTP and raw recipient
- [ ] Auth Service remains owner of OTP generation/hash/verify/rate limits
- [ ] Production provider secrets are kept outside Git
- [ ] No duplicate setup documentation added
