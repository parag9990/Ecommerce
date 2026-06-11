# Project Dependency & Setup Guide

## 1. Project Overview

This guide is generated for:

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Payment Service` |
| `TASK_FILE_NAME` | `task8.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task8_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

`TASK_FILE_NAME` ka scope reconciliation job setup hai. Simple Hinglish me: provider settlement CSV report ko local `payments` ledger ke saath compare karna, result `payment_reconciliations` table me save karna, and mismatch/missing record milne par finance alert event publish karna.

Important boundary:

- Original `INPUT_FILE_PATH` modify nahi kiya gaya.
- Ye file business logic rewrite nahi karti.
- Common clone, Go install, MySQL install, Docker MySQL, provider setup, and generic troubleshooting repeat nahi kiya gaya.
- Current code me reconciliation command already exists at `backend/services/payment-service/cmd/reconciliation/main.go`.
- Current code local CSV report source use karta hai, direct provider API/SFTP download adapter nahi.

Beginner note:

Reconciliation charge/refund create karne ka flow nahi hai. Ye accounting safety check hai. Agar local payment aur provider report mismatch kare, job evidence store karta hai and alert bhejta hai. Payment status silently change nahi hota.

## 2. Tech Stack

### Reused technology documentation

Common technology setup already explained hai:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
TaskImplementation/{SERVICE_NAME}/task7_Dependency.md
```

Follow these sections instead of duplicating:

| Previous Dependency File | Section / Topic | Why reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `2. Tech Stack` | Go, MySQL, HTTP event publisher, Docker baseline already explained. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Dependency Management` | Go modules and common Go dependency commands already explained. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `5. Database Setup` | MySQL install, Docker MySQL, DSN, and migration basics already explained. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `6. Redis / Queue / External Services` | HTTP event publisher behavior and no Redis/Kafka runtime already explained. |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` | `payment_db` and `payment_reconciliations` table already explained. |
| `TaskImplementation/{SERVICE_NAME}/task7_Dependency.md` | `5. Database Setup` | Migration `005` already explained; current reconciliation repository reads retry columns from `payments`. |

### `TASK_FILE_NAME` specific technology impact

| Technology / Component | Required? | Why used in `TASK_FILE_NAME` | Beginner explanation |
|---|---:|---|---|
| Go command package | Required | Runs one reconciliation job from `cmd/reconciliation` | Ye HTTP server nahi kholta; ek batch command run karta hai and exit hota hai. |
| Go `flag` package | Required | Reads `-provider`, `-report-file`, and `-report-date` flags | Command-line options parse karne ke liye built-in Go package. |
| Go `encoding/csv` | Required | Settlement CSV parse karta hai | Standard library hai, separate install nahi chahiye. |
| Go `crypto/sha256` | Required | Deterministic `reconciliation_id` create karta hai | Same report rerun par duplicate alert/result avoid karne me help karta hai. |
| Go `log/slog` | Required | JSON structured logs print karta hai | Job summary logs machine-readable bante hain. |
| MySQL | Required | Local payment ledger read and reconciliation results write karta hai | Financial audit data relational DB me safe and queryable rehta hai. |
| `github.com/go-sql-driver/mysql` | Required, reused | Go se MySQL connect karne ke liye driver | Already `go.mod` me present hai; new dependency install nahi chahiye. |
| HTTP event publisher | Required when reconciliation enabled | Non-matched result ke liye finance alert event publish karta hai | Current repo Kafka/RabbitMQ nahi use karta; JSON HTTP POST use hota hai. |
| Settlement CSV file | Required at runtime | Provider report ka task-specific input | CSV trusted source se aana chahiye and sensitive financial file ki tarah protect hona chahiye. |

## 3. Required Software

No new base software install introduced by `TASK_FILE_NAME`.

Follow:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Sections:

- `3. Required Software`
- `4. Dependency Management`
- `5. Database Setup`
- `8. Docker Setup`

Task-specific runtime needs:

| Software / Access | Required? | Status | Notes |
|---|---:|---|---|
| Go | Required | Reused | Run `go run ./cmd/reconciliation` and tests. |
| MySQL server | Required | Reused | Reads `payments`; writes `payment_reconciliations`. |
| MySQL client | Recommended | Reused | Verify rows after job run. |
| Settlement CSV report | Required | New task-specific input | Must contain required columns listed below. |
| HTTP event receiver | Required when reconciliation enabled | Reused config, task-specific alert event | Must accept alert JSON at `PAYMENT_EVENTS_ENDPOINT` and return 2xx. |
| Docker | Optional | Reused | Useful for MySQL only unless a service container is added later. |
| Scheduler or Kubernetes CronJob | Optional for local, recommended for deploy | New operational concern | Needed for daily automated run, but no manifest is committed for this task. |

If you only run unit tests, MySQL, Docker, provider account, Redis, Kafka, RabbitMQ, and event receiver are not required.

## 4. Dependency Management

This remains a Go modules project. `TASK_FILE_NAME` does not add a new external Go dependency.

Refer:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Section:

```text
4. Dependency Management
```

Current module details:

| Item | Current value |
|---|---|
| Service module | `backend/services/payment-service/go.mod` |
| Go version declared | `1.26.3` |
| Direct external dependency | `github.com/go-sql-driver/mysql v1.10.0` |
| Indirect dependency | `filippo.io/edwards25519 v1.2.0` |

Task-specific code paths:

| Path | Purpose |
|---|---|
| `backend/services/payment-service/cmd/reconciliation/main.go` | Command entrypoint, env load, flags, DB, CSV source, alert publisher wiring. |
| `backend/services/payment-service/internal/provider/settlement/csv.go` | Settlement CSV parser and max-file-size protection. |
| `backend/services/payment-service/internal/provider/settlement/report.go` | Normalized report, row validation, report date/window rules. |
| `backend/services/payment-service/internal/usecase/reconcile_settlements.go` | Compare provider rows against local payments, persist results, publish alerts. |
| `backend/services/payment-service/internal/events/reconciliation_alert.go` | Alert envelope for finance mismatch events. |
| `backend/services/payment-service/internal/repository/mysql_payment_repository.go` | MySQL lookups and idempotent reconciliation insert. |

Install only if dependencies were not downloaded before:

```bash
cd backend/services/payment-service
go mod download
```

Do not run `go get` for `TASK_FILE_NAME`; no new module package is needed.

## 5. Database Setup

### Reused MySQL setup

MySQL install, Docker MySQL, database creation, DSN format, and generic migration command already documented hai:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task7_Dependency.md
```

Relevant sections:

- `task1_Dependency.md` -> `5. Database Setup`
- `task2_Dependency.md` -> `5. Database Setup`
- `task7_Dependency.md` -> `5. Database Setup`

### What `TASK_FILE_NAME` needs from MySQL

| DB object | Required? | Why needed |
|---|---:|---|
| `payments` | Required | Provider payment ID, captured amount, currency, status, and updated window compare karne ke liye. |
| `payment_reconciliations` | Required | Reconciliation result audit rows persist karne ke liye. |
| `idx_payments_provider_payment` | Required for performance | Provider report row ko local payment se fast match karta hai. |
| `uk_reconciliation_id` | Required for idempotency | Same report rerun par duplicate rows prevent karta hai. |
| Migration `005_add_payment_retry_lineage.up.sql` | Required with current code | Shared payment scanner retry columns read karta hai; missing migration se `Unknown column` error aa sakta hai. |

No new migration file is introduced by `TASK_FILE_NAME`, but a working DB must have all current `.up.sql` migrations applied in order.

### Verify task-specific schema

Use this only after following previous MySQL setup docs:

```bash
cd backend/services/payment-service
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db -e "SHOW TABLES LIKE 'payment_reconciliations';"
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db -e "SHOW INDEX FROM payment_reconciliations;"
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db -e "SHOW COLUMNS FROM payments LIKE 'retry_of_payment_id';"
```

Expected:

- `payment_reconciliations` table exists.
- `uk_reconciliation_id` exists.
- `payments.retry_of_payment_id` exists, because migration `005` is part of current service schema.

### Task-specific query verification after a run

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db -e "
SELECT reconciliation_id, provider, settlement_id, status, payment_id, provider_payment_id, created_at
FROM payment_reconciliations
ORDER BY id DESC
LIMIT 10;"
```

### Credentials placement

Credentials are reused:

| Credential | Where to set | Already explained in |
|---|---|---|
| MySQL DSN | `PAYMENT_MYSQL_DSN` in shell env, secret manager, or local `.env` that you manually source | `task1_Dependency.md` |
| MySQL username/password | Inside DSN, for local only commonly `root:root` | `task1_Dependency.md` |
| DB host/port/name | Inside DSN, example `127.0.0.1:3306/payment_db` | `task1_Dependency.md` |

Security note: Local `root:root` is acceptable only for beginner local setup. Staging/production me restricted DB user and secret manager use karo.

## 6. Redis / Queue / External Services

### Redis / Kafka / RabbitMQ / NATS

No new Redis, Kafka, RabbitMQ, or NATS runtime dependency is introduced by `TASK_FILE_NAME`.

Current implementation publishes reconciliation alerts through the existing HTTP event publisher. Queue/broker setup duplicate nahi kiya gaya because current code does not require it.

Refer:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Section:

```text
6. Redis / Queue / External Services
```

### Settlement CSV report source

This is the main new task-specific external input.

Required CSV columns:

| Column | Required? | Example | Rule |
|---|---:|---|---|
| `settlement_id` | Yes | `stl_20260526` | Non-empty, max 128 chars, same value for all rows in one report. |
| `provider_payment_id` | Yes | `psp_pay_701` | Non-empty, max 128 chars, unique inside report. |
| `outcome` | Yes | `captured` | Supports `captured`, `settled`, or `refunded`; `settled` normalizes to `captured`. |
| `currency` | Yes | `INR` | 3-letter uppercase ISO currency after normalization. |
| `settled_amount` | Yes | `249900` | Integer minor units, for INR this means paise. Must be greater than zero. |
| `fee_amount` | Yes | `7250` | Integer minor units. Must be zero or positive. |
| `settled_at` | Yes | `2026-05-26T16:30:00Z` | RFC3339 timestamp. |

Example CSV:

```csv
settlement_id,provider_payment_id,outcome,currency,settled_amount,fee_amount,settled_at
stl_20260526,psp_pay_701,captured,INR,249900,7250,2026-05-26T16:30:00Z
stl_20260526,psp_pay_702,settled,INR,149900,4300,2026-05-26T16:31:00Z
```

Practical notes:

- Amount float me mat rakho, jaise `2499.00`. Hamesha integer minor units use karo.
- Report file trusted provider source se aani chahiye.
- CSV file ko repo me commit mat karo if it contains real financial data.
- `PAYMENT_RECONCILIATION_MAX_REPORT_FILE_BYTES` default `67108864` bytes hai.

### HTTP event receiver for alerts

Reused setup:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
```

Task-specific behavior:

| Item | Value |
|---|---|
| Event type | `PaymentReconciliationMismatchDetected` |
| Default topic | `payment.reconciliation.alerts` |
| Header | `Authorization: Bearer <PAYMENT_EVENTS_AUTH_TOKEN>` |
| Idempotency header | `Idempotency-Key: <alert event id>` |
| Sends alert for | `mismatch`, `missing_local`, `missing_provider` |
| Does not alert for | `matched` |

Alert publisher failures are logged and counted in `alert_failure_count`. Persisted reconciliation rows are not rolled back only because alert delivery failed.

## 7. Environment Variables

`.env` location and loading behavior are already documented in previous files. Important reminder: current code reads env with `os.Getenv`; `.env` auto-load nahi hota. Local `.env` use kar rahe ho to manually source karo:

```bash
cd backend/services/payment-service
. ./.env
```

Reused full environment reference:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Section:

```text
7. Environment Variables
```

### Task-specific env overlay

Only add or change these values when running reconciliation:

```env
PAYMENT_RECONCILIATION_ENABLED=true
PAYMENT_RECONCILIATION_PROVIDER=stripe_like
PAYMENT_RECONCILIATION_REPORT_FILE=./settlements/stripe-like-2026-05-26.csv
PAYMENT_RECONCILIATION_REPORT_LAG_HOURS=24
PAYMENT_RECONCILIATION_BATCH_SIZE=500
PAYMENT_RECONCILIATION_TIMEOUT=20m
PAYMENT_RECONCILIATION_MAX_REPORT_FILE_BYTES=67108864
PAYMENT_RECONCILIATION_ALERT_TOPIC=payment.reconciliation.alerts

PAYMENT_EVENTS_ENDPOINT=http://127.0.0.1:9080/events
PAYMENT_EVENTS_AUTH_TOKEN=event-publisher-token-at-least-32-characters
PAYMENT_EVENTS_TIMEOUT=5s
```

### Variable reference for new or task-specific values

| Variable | Required? | Default | Purpose | Security / validation notes |
|---|---:|---|---|---|
| `PAYMENT_RECONCILIATION_ENABLED` | Required to run job | `false` | Enables command work. If false, command logs disabled and exits. | Keep false in environments where report source is not ready. |
| `PAYMENT_RECONCILIATION_PROVIDER` | Required when enabled unless `-provider` flag is passed | Empty | Normalized provider name, like `stripe_like`. | Max 64 chars; keep same provider value used in `payments.provider`. |
| `PAYMENT_RECONCILIATION_REPORT_FILE` | Required when enabled unless `-report-file` flag is passed | Empty | Path to settlement CSV. | Sensitive financial file; do not commit real reports. |
| `PAYMENT_RECONCILIATION_REPORT_LAG_HOURS` | Optional | `24` | If `-report-date` is omitted, command picks date by subtracting this lag from now. | Cannot be negative. |
| `PAYMENT_RECONCILIATION_BATCH_SIZE` | Optional | `500` | Local payment candidate query page size. | Must be `1` to `10000`. |
| `PAYMENT_RECONCILIATION_TIMEOUT` | Optional | `20m` | Max time for one command run. | Must be greater than zero. |
| `PAYMENT_RECONCILIATION_MAX_REPORT_FILE_BYTES` | Optional | `67108864` | Maximum CSV size read by command. | Must be greater than zero; protects against accidental huge files. |
| `PAYMENT_RECONCILIATION_ALERT_TOPIC` | Required when enabled | `payment.reconciliation.alerts` | Topic field sent to event receiver. | Not secret, but keep stable for finance routing. |
| `PAYMENT_EVENTS_ENDPOINT` | Required when enabled | Empty | HTTP receiver for reconciliation alerts. | Must be HTTPS outside localhost/loopback. |
| `PAYMENT_EVENTS_AUTH_TOKEN` | Required when enabled | Empty | Bearer token used for alert POST. | Must be at least 32 chars. Secret. |
| `PAYMENT_EVENTS_TIMEOUT` | Optional | `5s` | Alert HTTP client timeout. | Must be greater than zero. |

Existing variables still needed:

| Existing variable | Why reconciliation needs it | Where already explained |
|---|---|---|
| `PAYMENT_MYSQL_DSN` | Reads `payments`; writes `payment_reconciliations`. | `task1_Dependency.md`, `task2_Dependency.md` |
| `PAYMENT_MYSQL_MAX_OPEN_CONNS` and related pool vars | DB connection pool for command. | `task1_Dependency.md` |

Important: Provider API keys like `STRIPE_LIKE_SECRET_KEY` are not required for current CSV reconciliation command unless your environment also enables provider APIs elsewhere. This task reads a local CSV, not provider API.

## 8. Docker Setup

No new Dockerfile, docker-compose service, Redis, Kafka, RabbitMQ, provider mock container, or reconciliation-specific container is introduced by `TASK_FILE_NAME`.

Reuse:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Section:

```text
8. Docker Setup
```

Task-specific Docker notes:

| Component | Status | Notes |
|---|---|---|
| MySQL container | Reused | Same `payment_db` container setup; all migrations must be applied. |
| Reconciliation command container | Not committed | Run with `go run ./cmd/reconciliation` unless a Dockerfile is added later. |
| Settlement report volume | New if containerized later | Mount report file read-only, for example `/reports/settlement.csv:ro`. |
| Event receiver container | Optional / external | Needed only if local event receiver runs in Docker. |
| Scheduler container/CronJob | Not committed | Deployment can add Kubernetes CronJob or cron later. |

### Ports and networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Backend HTTP API | `8080` by default | Main HTTP service, not required by reconciliation command | Reused |
| MySQL | `3306` | Payment DB access | Reused |
| HTTP event receiver | `9080` example | Receives reconciliation alert events | Reused config, task-specific event |
| Reconciliation command | None | Batch command opens no listening port | New task-specific note |

Docker networking rules:

- If command runs on host and MySQL runs in Docker with `-p 3306:3306`, DSN can use `127.0.0.1:3306`.
- If command runs in a future Docker container, DSN should use Docker service name, not `127.0.0.1`.
- If event receiver runs in Docker and command runs on host, `http://127.0.0.1:9080/events` works only if port is published.
- If both command and receiver run inside Docker network later, use receiver service name in `PAYMENT_EVENTS_ENDPOINT`.

## 9. Local Development Setup

### Step 1: Read previous setup docs first

Follow these before task-specific steps:

| File | Why |
|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Clone, Go, MySQL, Docker, env loading, base run commands. |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `payment_db`, schema, migration verification. |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | Provider/event publisher concept. |
| `TaskImplementation/{SERVICE_NAME}/task7_Dependency.md` | Migration `005` and retry columns required by current shared repository scanner. |

### Step 2: Go to service directory

```bash
cd backend/services/payment-service
```

### Step 3: Install dependencies only if not already done

```bash
go mod download
```

### Step 4: Start MySQL and apply migrations

Use previous MySQL/Docker setup. For current service code, apply all current migrations:

```bash
for file in migrations/*.up.sql; do
  mysql -h 127.0.0.1 -P 3306 -u root -proot < "$file"
done
```

If your DB already has migrations `001` to `004`, at minimum ensure migration `005_add_payment_retry_lineage.up.sql` is also applied.

### Step 5: Prepare settlement report file

Recommended local folder:

```text
backend/services/payment-service/settlements/
```

Example report path:

```text
backend/services/payment-service/settlements/stripe-like-2026-05-26.csv
```

Minimum CSV shape:

```csv
settlement_id,provider_payment_id,outcome,currency,settled_amount,fee_amount,settled_at
stl_20260526,psp_pay_701,captured,INR,249900,7250,2026-05-26T16:30:00Z
```

For a meaningful match, `provider_payment_id` in CSV should exist in local `payments.provider_payment_id`. If not found, result becomes `missing_local`.

### Step 6: Configure env

Use the task-specific overlay from section `7. Environment Variables`. Also make sure the DB DSN is present:

```env
PAYMENT_MYSQL_DSN=root:root@tcp(127.0.0.1:3306)/payment_db?parseTime=true&charset=utf8mb4,utf8&loc=UTC
```

If using a local `.env`:

```bash
. ./.env
```

### Step 7: Start or point to event receiver

Reconciliation enabled mode validates `PAYMENT_EVENTS_ENDPOINT` and `PAYMENT_EVENTS_AUTH_TOKEN`.

Receiver expectations:

| Requirement | Value |
|---|---|
| Method | `POST` |
| Content type | `application/json` |
| Auth header | `Authorization: Bearer <PAYMENT_EVENTS_AUTH_TOKEN>` |
| Success response | Any `2xx` |

If all report rows are matched, no alert POST is sent. If any mismatch/missing result exists and receiver is down, command still stores DB rows but logs alert failures.

### Step 8: Run reconciliation command

With explicit flags:

```bash
go run ./cmd/reconciliation \
  -provider stripe_like \
  -report-file ./settlements/stripe-like-2026-05-26.csv \
  -report-date 2026-05-26
```

With env defaults:

```bash
go run ./cmd/reconciliation
```

Expected success log includes:

```text
payment.reconciliation.completed
matched_count
mismatch_count
missing_local_count
missing_provider_count
new_result_count
alerted_count
alert_failure_count
```

### Step 9: Verify DB results

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db -e "
SELECT provider, settlement_id, status, COUNT(*) AS total
FROM payment_reconciliations
GROUP BY provider, settlement_id, status
ORDER BY settlement_id DESC, status;"
```

## 10. Running the Project

### Run task-specific tests

```bash
cd backend/services/payment-service
go test ./cmd/reconciliation ./internal/provider/settlement ./internal/usecase
```

### Run all service tests

```bash
cd backend/services/payment-service
go test ./...
```

### Run command disabled mode

Useful smoke test when env is not ready:

```bash
PAYMENT_RECONCILIATION_ENABLED=false go run ./cmd/reconciliation
```

Expected log:

```text
payment.reconciliation.disabled
```

### Run command enabled mode

```bash
cd backend/services/payment-service
. ./.env
go run ./cmd/reconciliation \
  -provider stripe_like \
  -report-file ./settlements/stripe-like-2026-05-26.csv \
  -report-date 2026-05-26
```

### Backfill old report date

Use explicit `-report-date`:

```bash
go run ./cmd/reconciliation \
  -provider stripe_like \
  -report-file ./settlements/stripe-like-2026-05-20.csv \
  -report-date 2026-05-20
```

Same provider, settlement ID, and report evidence rerun should not create duplicate alerts because `reconciliation_id` is deterministic and unique.

## 11. Common Errors & Fixes

Generic Go/MySQL/Docker/provider errors are already documented in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
TaskImplementation/{SERVICE_NAME}/task7_Dependency.md
```

Task-specific errors:

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `payment.reconciliation.disabled` | `PAYMENT_RECONCILIATION_ENABLED=false` or missing. | Set `PAYMENT_RECONCILIATION_ENABLED=true`. | Keep separate local env overlay for reconciliation. |
| `provider is required` | Provider flag/env empty. | Set `PAYMENT_RECONCILIATION_PROVIDER` or pass `-provider`. | Use same provider value stored in `payments.provider`. |
| `report-file is required` | Report path flag/env empty. | Set `PAYMENT_RECONCILIATION_REPORT_FILE` or pass `-report-file`. | Keep report path in secure run config. |
| `report-date must be YYYY-MM-DD` | Wrong date format like `05/26/2026`. | Use `2026-05-26`. | Use ISO date in scripts and runbooks. |
| `open settlement report: no such file or directory` | CSV path wrong or command run from different directory. | Use absolute path or run from `backend/services/payment-service`. | Store report path relative to service dir or use absolute path in scheduled jobs. |
| `permission denied` opening report | OS user running command cannot read CSV. | Fix file permissions or mount read-only with correct user. | Use controlled report directory and least privilege. |
| `settlement report is missing required column` | CSV header incomplete or misspelled. | Add required columns exactly as documented. | Validate CSV before scheduling. |
| `settled_at must be RFC3339` | Timestamp is not like `2026-05-26T16:30:00Z`. | Export provider report with RFC3339 timestamps. | Normalize report in adapter before running. |
| `settled_amount must be an integer minor-unit amount` | Amount has decimal or text value. | Use minor units, e.g. `249900` paise. | Never use float money in reconciliation reports. |
| `settlement report contains duplicate provider_payment_id` | Same provider payment appears twice in one report. | Fix report or investigate provider export. | Add upstream report validation/checksum. |
| `settlement report contains multiple settlement_id values` | One CSV mixes multiple provider settlement batches. | Split CSV by settlement ID. | One command run should process one settlement report identity. |
| `settlement report exceeds ... bytes` | CSV larger than configured max. | Increase `PAYMENT_RECONCILIATION_MAX_REPORT_FILE_BYTES` only after review, or split report. | Keep report size budget documented. |
| `Unknown column 'retry_of_payment_id'` | Migration `005` missing while current code selects retry columns. | Apply `migrations/005_add_payment_retry_lineage.up.sql`. | Run all migrations before command. |
| `PAYMENT_EVENTS_ENDPOINT is required when payment event publishing is enabled` | Reconciliation enabled but event endpoint missing. | Set `PAYMENT_EVENTS_ENDPOINT`. | Keep env validation in deployment. |
| `PAYMENT_EVENTS_AUTH_TOKEN must be at least 32 characters` | Token too short. | Use strong 32+ char secret. | Generate token through secret manager. |
| `PAYMENT_EVENTS_ENDPOINT must use HTTPS outside local testing` | Remote endpoint uses plain HTTP. | Use HTTPS for non-loopback endpoint. | Keep local-only HTTP restricted to `localhost` or `127.0.0.1`. |
| `publish payment event: endpoint returned status ...` | Event receiver returned non-2xx. | Check receiver logs/auth/topic support. | Monitor `alert_failure_count`. |
| `reconciliation conflict` | Same deterministic result ID already exists but new evidence differs. | Investigate revised/corrupted report. Use a new settlement/report identity for corrected reports. | Do not overwrite old financial evidence. |
| No `matched` rows, many `missing_local` | CSV provider IDs do not exist in local DB. | Seed/run real payment flow first or use matching provider IDs. | Compare report provider with local `payments.provider`. |
| Many `missing_provider` rows | Local captured payments in report window absent from CSV. | Check report date/window and provider settlement lag. | Use `PAYMENT_RECONCILIATION_REPORT_LAG_HOURS` and provider-ready schedule. |

## 12. Security & Best Practices

### Security audit observations

| Observation | Risk | Recommended fix |
|---|---|---|
| Settlement CSV path is local file based. | Manual file handling can leak financial data. | Store reports in controlled encrypted location, mount read-only, and delete according to retention policy. |
| No provider API/SFTP report downloader exists in current code. | Human file transfer can be error-prone. | Later add authenticated read-only provider report adapter with checksum/signature validation. |
| Alert endpoint is required but receiver implementation is external. | Mismatch rows can persist while finance alert delivery fails. | Monitor logs for `alert_failure_count` and add alert delivery dashboard/retry if needed. |
| No committed CronJob/cron config found. | Daily reconciliation can be forgotten. | Add Kubernetes CronJob or scheduler once deployment stack is ready. |
| Local default MySQL DSN uses `root:root`. | Unsafe outside local dev. | Use restricted DB credentials from secret manager in non-local environments. |
| `.env` is not auto-loaded. | Beginner may think variables are set when they are not. | Source `.env` explicitly or use process manager/secret injection. |
| No service Dockerfile/compose found. | Local setup can vary between developers. | Reuse previous docs now; add standardized compose profile later. |

### Best practices for `TASK_FILE_NAME`

- Run reconciliation after provider settlement report is complete, not immediately after payment capture.
- Keep one CSV to one settlement ID/report identity.
- Use integer minor units for money, never float.
- Keep CSV files out of Git, logs, tickets, and screenshots if they contain real finance data.
- Treat `missing_local`, amount mismatch, and currency mismatch as critical finance signals.
- Do not automatically mutate payment/refund state from reconciliation result.
- Rerun same report safely, but treat changed evidence for same settlement ID as investigation-worthy.
- Keep `PAYMENT_EVENTS_AUTH_TOKEN` strong and rotate it like a real secret.
- Use HTTPS event endpoints outside loopback local testing.
- Review `payment_reconciliations.details` before adding new evidence fields; no card, bank, UPI, token, or unnecessary PII.

## 13. Missing or Misconfigured Things

| Item | Current state | Impact | Suggested improvement |
|---|---|---|---|
| Task-specific migration | Not needed | Existing table is in migration `001`; shared scanner still needs migration `005`. | Keep migration order clearly documented in deploy runbook. |
| Provider report downloader | Not implemented | CSV must be delivered manually or by external process. | Add provider-specific read-only report adapter later. |
| Scheduler/CronJob manifest | Not committed | Job is manual locally and deployment scheduler is undefined. | Add Kubernetes CronJob or approved scheduler config. |
| Event receiver service | Not in this service folder | Alert publishing depends on external endpoint. | Provide local event sink or shared event receiver compose profile. |
| Metrics backend config | Not committed here | Command logs summary, but Prometheus/alert rules are external. | Add metrics/alert integration in platform monitoring task. |
| Sample settlement CSV | Not committed | Beginners need to construct file manually. | Add sanitized sample fixture with fake IDs only. |
| `.env.example` for task overlay | Not found in service scan | Beginners can miss required values. | Add safe `.env.example` without secrets. |
| Service Dockerfile/compose | Not found in service scan | Containerized run instructions remain conceptual. | Add service Dockerfile and compose with MySQL, event receiver, report mount. |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `2. Tech Stack` | Go, MySQL, HTTP event publisher, Docker baseline already explained. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3. Required Software` | Git, Go, MySQL, Docker, curl setup already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Dependency Management` | Go modules, `go mod download`, `go mod tidy`, build/run commands already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `5. Database Setup` | MySQL install, Docker setup, DSN, migration flow already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `6. Redis / Queue / External Services` | No Redis/Kafka runtime and HTTP event publisher behavior already explained. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. Environment Variables` | Broad env reference already includes reconciliation variables; this file only adds task-specific overlay and behavior. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `8. Docker Setup` | MySQL-only Docker baseline already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `11. Ports & Networking` | Base backend, MySQL, and event receiver ports already documented. |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` | `payment_db`, `payments`, `payment_reconciliations`, indexes, and schema verification reused. |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `6. Redis / Queue / External Services` | Provider and HTTP event receiver setup context reused. |
| `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md` | `8. Environment Variables` | Event publisher and HTTPS/loopback rules reused. |
| `TaskImplementation/{SERVICE_NAME}/task7_Dependency.md` | `5. Database Setup` | Migration `005` retry columns required by current shared repository scanner reused. |

## 15. Final Checklist

- [ ] Previous dependency documentation checked before following this file.
- [ ] No duplicate Go/MySQL/Docker/provider setup copied into this task guide.
- [ ] Go dependencies downloaded with `go mod download` if needed.
- [ ] MySQL is running and `PAYMENT_MYSQL_DSN` points to `payment_db`.
- [ ] All current `.up.sql` migrations applied, including migration `005`.
- [ ] `payment_reconciliations` table and `uk_reconciliation_id` verified.
- [ ] Settlement CSV has all required columns.
- [ ] Amounts are integer minor units.
- [ ] `settled_at` values are RFC3339 timestamps.
- [ ] `PAYMENT_RECONCILIATION_ENABLED=true` only in environments ready to run the job.
- [ ] Provider is set via `PAYMENT_RECONCILIATION_PROVIDER` or `-provider`.
- [ ] Report path is set via `PAYMENT_RECONCILIATION_REPORT_FILE` or `-report-file`.
- [ ] `PAYMENT_EVENTS_ENDPOINT` and 32+ char `PAYMENT_EVENTS_AUTH_TOKEN` configured.
- [ ] Event receiver returns `2xx` for alert POSTs.
- [ ] Command run with explicit `-report-date` for manual/backfill runs.
- [ ] Job logs checked for matched, mismatch, missing, created, alerted, and alert failure counts.
- [ ] DB results verified in `payment_reconciliations`.
- [ ] Real settlement CSV files are not committed to Git.
- [ ] No automatic payment/refund mutation is done from reconciliation results.

`TASK_FILE_NAME` setup ka short formula: previous setup plus all migrations, secure settlement CSV, reconciliation env overlay, event receiver config, and one deterministic run per provider/report date. Agar issue aaye, pehle CSV format and report date verify karo, phir DB migrations, phir event receiver/auth logs.
