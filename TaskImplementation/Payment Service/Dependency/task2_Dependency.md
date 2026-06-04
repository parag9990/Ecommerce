# Project Dependency & Setup Guide

## 1. Project Overview

This guide is generated for:

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Payment Service` |
| `TASK_FILE_NAME` | `task2.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task2_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

`TASK_FILE_NAME` ka scope hai MySQL schema create karna. Is dependency guide ka focus sirf setup, database migration, verification, environment, DevOps, and beginner onboarding par hai. Business logic, API implementation, provider SDK integration, refund workflow, retry workflow, and reconciliation worker ko yahan rewrite nahi kiya gaya.

Important reuse note:

Most common setup already documented hai in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Is file me only `TASK_FILE_NAME` ke new or task-specific parts detail me explain kiye gaye hain: `payment_db` schema, baseline migration, table/index/FK verification, schema health endpoint, and migration safety.

## 2. Tech Stack

### Reused technology documentation

Refer:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Sections:

- `2. Tech Stack`
- `3. Required Software`
- `4. Dependency Management`
- `5. Database Setup`

Same setup already explained hai for Go, Go Modules, `net/http`, `database/sql`, `github.com/go-sql-driver/mysql`, MySQL, Docker, SQL migrations, and common Go commands. Isliye yahan duplicate installation guide repeat nahi kiya gaya.

### `TASK_FILE_NAME` specific technology impact

| Technology | Required? | `TASK_FILE_NAME` use | Beginner explanation |
|---|---:|---|---|
| MySQL 8.x | Required | `payment_db` and payment tables create karne ke liye | MySQL ek relational database hai. Payment data rows/tables me store hota hai, jahan unique keys duplicate payment/webhook prevent karte hain. |
| InnoDB | Required | Foreign keys and transactional safety ke liye | InnoDB MySQL ka storage engine hai jo transaction, row locking, and foreign key support deta hai. Payment systems ke liye yeh important hai. |
| SQL migrations | Required | Schema ko repeatable tarike se apply/update karne ke liye | Migration file ek versioned SQL script hota hai. Isse har developer same DB structure bana sakta hai. |
| MySQL JSON columns | Required | Sanitized provider payload and reconciliation details store karne ke liye | JSON column flexible data store karta hai, but secrets/card data store nahi karna. |
| Schema health endpoint | Required for verification | DB tables/indexes/FKs ready hain ya nahi check karne ke liye | Endpoint app se DB schema verify karta hai, beginner ko quick signal milta hai ki migration properly lagi ya nahi. |

## 3. Required Software

No new software is introduced beyond previous setup.

Please follow:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Sections:

- `3. Required Software`
- `5. Database Setup`
- `8. Docker Setup`

`TASK_FILE_NAME` ke liye minimum required tools:

| Software | Why needed for `TASK_FILE_NAME` | Status |
|---|---|---|
| Go | Service and tests run karne ke liye | Reused |
| MySQL Server | `payment_db` schema host karne ke liye | Reused |
| MySQL client | Migration files execute and schema verify karne ke liye | Reused |
| Docker | Local MySQL quickly run karne ke liye optional | Reused |
| curl | Schema health endpoint verify karne ke liye | Reused |

## 4. Dependency Management

This remains a Go modules project. `TASK_FILE_NAME` schema work does not introduce any new Go package.

Refer:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Section:

```text
4. Dependency Management
```

Current relevant files:

| File | Purpose |
|---|---|
| `backend/services/payment-service/go.mod` | Go module and direct dependency list |
| `backend/services/payment-service/go.sum` | Module checksum lock file |
| `backend/services/payment-service/migrations/001_create_payment_tables.up.sql` | `TASK_FILE_NAME` baseline schema migration |
| `backend/services/payment-service/migrations/001_create_payment_tables.down.sql` | `TASK_FILE_NAME` rollback migration |

`TASK_FILE_NAME` dependency status:

| Dependency | New in `TASK_FILE_NAME`? | Notes |
|---|---:|---|
| `github.com/go-sql-driver/mysql` | No | Already needed for MySQL access. |
| External provider SDK | No | `TASK_FILE_NAME` is schema only. |
| Redis/Kafka/RabbitMQ client | No | Not used by `TASK_FILE_NAME`. |
| Migration CLI package | No | Repo uses raw `.sql` files; no committed migration runner dependency found. |

## 5. Database Setup

### What is new for `TASK_FILE_NAME`

`TASK_FILE_NAME` introduces the core MySQL schema foundation for `payment_db`.

Database:

```text
payment_db
```

Baseline migration:

```text
backend/services/payment-service/migrations/001_create_payment_tables.up.sql
```

Rollback migration:

```text
backend/services/payment-service/migrations/001_create_payment_tables.down.sql
```

### Tables created by `TASK_FILE_NAME` baseline migration

| Table | Purpose | Mandatory? |
|---|---|---:|
| `payments` | Main payment aggregate: order, user, provider, amount, status, idempotency | Yes |
| `payment_attempts` | Attempt-level audit trail for provider calls/retries | Yes |
| `refunds` | Refund request and outcome records | Yes |
| `payment_webhook_events` | Provider webhook idempotency and processing audit | Yes |
| `payment_reconciliations` | Settlement/reconciliation result records | Yes |

Simple Hinglish:

`SERVICE_NAME` me paisa-related data safe and auditable hona chahiye. Isliye schema me unique keys, indexes, foreign keys, timestamps, and status fields use kiye gaye hain. Duplicate payment intent, duplicate refund, and duplicate webhook ko database level par block karna important hai.

### Baseline schema responsibilities

| Area | Database support |
|---|---|
| Duplicate payment prevention | `uk_payments_idempotency` on `(provider, idempotency_key)` |
| Stable payment lookup | `uk_payments_payment_id` |
| Webhook idempotency | `uk_webhook_provider_event` on `(provider, provider_event_id)` |
| Refund idempotency | `uk_refunds_idempotency` on `(payment_id, idempotency_key)` |
| Attempt audit | `payment_attempts` table with FK to `payments(payment_id)` |
| Refund audit | `refunds` table with actor and review fields |
| Finance reconciliation | `payment_reconciliations` table |

### Current repo migration ordering

`TASK_FILE_NAME` baseline is `001`, but current service folder also contains later migrations from later tasks. For running the current service end-to-end, apply all `.up.sql` files in order.

| Migration | Task relationship | What it changes |
|---|---|---|
| `001_create_payment_tables.up.sql` | `TASK_FILE_NAME` baseline | Creates `payment_db` and five core tables |
| `002_add_webhook_payment_lookup_index.up.sql` | Later task support | Adds `idx_payments_provider_intent` |
| `003_add_refund_review_reason.up.sql` | Later task support | Adds `refunds.review_reason` |
| `004_add_refund_provider_lookup_index.up.sql` | Later task support | Adds `idx_refunds_provider_refund` |
| `005_add_payment_retry_lineage.up.sql` | Later task support | Adds retry lineage columns/indexes/FK on `payments` |

Important:

- If you are studying only `TASK_FILE_NAME`, understand `001_create_payment_tables.up.sql`.
- If you are running current local service, apply all current `.up.sql` migrations because code schema verification expects later indexes/columns too.

### Apply migrations

Common MySQL/Docker setup is already explained in `task1_Dependency.md`. After MySQL is running, use the existing repo migration files.

From service folder:

```bash
cd backend/services/payment-service
```

Apply all current up migrations:

```bash
for file in migrations/*.up.sql; do
  mysql -h 127.0.0.1 -P 3306 -u root -proot < "$file"
done
```

`TASK_FILE_NAME`-only baseline apply command:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot < migrations/001_create_payment_tables.up.sql
```

Beginner note:

Full current service run ke liye all migrations apply karna better hai. Sirf `001` apply karne par baseline `TASK_FILE_NAME` schema ban jayega, but current `/internal/v1/payment-schema/health` later migration fields missing dikha sakta hai.

### Verify database and tables

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot -e "SHOW DATABASES LIKE 'payment_db';"
```

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db -e "SHOW TABLES;"
```

Expected `TASK_FILE_NAME` baseline tables:

```text
payment_attempts
payment_reconciliations
payment_webhook_events
payments
refunds
```

### Verify important indexes

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db -e "SHOW INDEX FROM payments;"
```

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db -e "SHOW INDEX FROM payment_webhook_events;"
```

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db -e "SHOW INDEX FROM refunds;"
```

Must-have `TASK_FILE_NAME` indexes:

| Table | Index / Unique key | Why important |
|---|---|---|
| `payments` | `uk_payments_payment_id` | Stable payment ID duplicate nahi hoga |
| `payments` | `uk_payments_idempotency` | Same provider + idempotency key duplicate payment block karega |
| `payments` | `idx_payments_order` | Order based lookup fast hoga |
| `payments` | `idx_payments_provider_payment` | Provider webhook/reconciliation lookup fast hoga |
| `payments` | `idx_payments_status_created` | Admin/job status filtering fast hoga |
| `payment_attempts` | `uk_payment_attempts_attempt_id` | Duplicate attempt ID block hoga |
| `refunds` | `uk_refunds_refund_id` | Duplicate refund ID block hoga |
| `refunds` | `uk_refunds_idempotency` | Duplicate refund request block hoga |
| `payment_webhook_events` | `uk_webhook_provider_event` | Duplicate provider webhook no-op ho sakega |
| `payment_reconciliations` | `uk_reconciliation_id` | Duplicate reconciliation rows avoid honge |

### Verify foreign keys

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot payment_db -e "SELECT table_name, constraint_name FROM information_schema.table_constraints WHERE table_schema='payment_db' AND constraint_type='FOREIGN KEY';"
```

Expected `TASK_FILE_NAME` baseline foreign keys:

| Table | Foreign key | Purpose |
|---|---|---|
| `payment_attempts` | `fk_payment_attempts_payment` | Attempt must belong to valid payment |
| `refunds` | `fk_refunds_payment` | Refund must belong to valid payment |

If all migrations are applied, current repo also expects:

| Table | Foreign key | Purpose |
|---|---|---|
| `payments` | `fk_payments_retry_of` | Retry payment can reference original payment |

## 6. Redis / Queue / External Services

No new external service is introduced by `TASK_FILE_NAME`.

Reuse:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Sections:

- `6. Redis / Queue / External Services`
- `8. Docker Setup`
- `11. Ports & Networking`

`TASK_FILE_NAME` status:

| Service | Required for `TASK_FILE_NAME`? | Notes |
|---|---:|---|
| Redis | No | Current schema task does not use Redis. |
| Kafka | No | Current service publishes events over HTTP in later flows, not Kafka. |
| RabbitMQ | No | No RabbitMQ client found. |
| NATS | No | No NATS client found. |
| HTTP event receiver | No for `TASK_FILE_NAME` | Required only when provider/reconciliation event publishing is enabled. |
| Payment provider API | No for `TASK_FILE_NAME` | Provider integration comes in later tasks. |

## 7. Environment Variables

No new environment variables are introduced by `TASK_FILE_NAME` beyond existing MySQL connection settings.

Refer:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Section:

```text
7. Environment Variables
```

`TASK_FILE_NAME` relevant existing variables:

| Variable | Required? | Why `TASK_FILE_NAME` cares |
|---|---:|---|
| `PAYMENT_MYSQL_DSN` | Required for service run | Points service to `payment_db` |
| `MYSQL_DSN` | Optional fallback | Used if `PAYMENT_MYSQL_DSN` is empty |
| `PAYMENT_MYSQL_MAX_OPEN_CONNS` | Optional | DB connection pool tuning |
| `PAYMENT_MYSQL_MAX_IDLE_CONNS` | Optional | DB idle connection tuning |
| `PAYMENT_MYSQL_CONN_MAX_LIFETIME` | Optional | DB connection lifecycle |
| `PAYMENT_MYSQL_CONN_MAX_IDLE_TIME` | Optional | DB idle lifecycle |

Minimal `TASK_FILE_NAME` `.env` snippet:

```env
PAYMENT_HTTP_ADDR=:8080
PAYMENT_MYSQL_DSN=root:root@tcp(127.0.0.1:3306)/payment_db?parseTime=true&charset=utf8mb4,utf8&loc=UTC
PAYMENT_MYSQL_MAX_OPEN_CONNS=25
PAYMENT_MYSQL_MAX_IDLE_CONNS=10
PAYMENT_MYSQL_CONN_MAX_LIFETIME=30m
PAYMENT_MYSQL_CONN_MAX_IDLE_TIME=5m
```

Where to place local env file:

```text
backend/services/payment-service/.env
```

Important:

Current code reads environment variables through `os.Getenv`. `.env` automatic load nahi hota. Shell me load karna padega:

```bash
set -a
. ./.env
set +a
```

Common mistake:

`.env` file bana diya but source nahi kiya, to app default DSN use karega. Agar default DB credentials local MySQL se match nahi karte, connection fail hoga.

## 8. Docker Setup

No new Docker container is introduced by `TASK_FILE_NAME`.

Reuse:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Sections:

- `5. Database Setup`
- `8. Docker Setup`
- `11. Ports & Networking`

`TASK_FILE_NAME` Docker-specific need:

| Container | Port | Purpose | Status |
|---|---:|---|---|
| MySQL | 3306 | Hosts `payment_db` schema | Reused |

If MySQL Docker container is already created from previous guide:

```bash
docker start ecommerce-payment-mysql
```

Then apply migrations from:

```text
backend/services/payment-service/migrations
```

No committed `Dockerfile` or service-level `docker-compose.yml` was found for current payment service setup. This is already called out in previous dependency documentation and remains unchanged for `TASK_FILE_NAME`.

## 9. Local Development Setup

### Step 1: Read previous dependency documentation

Start here:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Follow these sections first:

- `3. Required Software`
- `4. Dependency Management`
- `5. Database Setup`
- `7. Environment Variables`
- `8. Docker Setup`
- `9. Local Development Setup`

### Step 2: Go to service directory

```bash
cd backend/services/payment-service
```

### Step 3: Install dependencies

No new dependency added by `TASK_FILE_NAME`. Use the existing Go setup:

```bash
go mod download
```

### Step 4: Start MySQL

If using the reused Docker setup:

```bash
docker start ecommerce-payment-mysql
```

### Step 5: Apply `TASK_FILE_NAME` schema

`TASK_FILE_NAME`-only baseline:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -proot < migrations/001_create_payment_tables.up.sql
```

Recommended for current repo:

```bash
for file in migrations/*.up.sql; do
  mysql -h 127.0.0.1 -P 3306 -u root -proot < "$file"
done
```

### Step 6: Load environment variables

```bash
set -a
. ./.env
set +a
```

### Step 7: Start backend service

```bash
go run ./cmd/server
```

### Step 8: Verify `TASK_FILE_NAME` schema endpoints

```bash
curl http://localhost:8080/internal/v1/payment-schema
```

```bash
curl http://localhost:8080/internal/v1/payment-schema/health
```

Expected health behavior:

| Response | Meaning |
|---|---|
| HTTP `200` and `"ready": true` | DB schema expected by current code is ready |
| HTTP `503` and missing tables/indexes/FKs | One or more migrations are missing |
| Connection error | Server not running or wrong port |

### Step 9: Run schema-related tests

```bash
go test ./internal/domain ./internal/usecase ./internal/repository
```

Full current service tests:

```bash
go test ./...
```

## 10. Running the Project

`TASK_FILE_NAME` does not change the normal run command. It changes the required DB readiness before server verification.

Minimal `TASK_FILE_NAME` run flow:

```bash
cd backend/services/payment-service
go mod download
docker start ecommerce-payment-mysql
for file in migrations/*.up.sql; do
  mysql -h 127.0.0.1 -P 3306 -u root -proot < "$file"
done
export PAYMENT_MYSQL_DSN='root:root@tcp(127.0.0.1:3306)/payment_db?parseTime=true&charset=utf8mb4,utf8&loc=UTC'
go run ./cmd/server
```

Verify:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/internal/v1/payment-schema
curl http://localhost:8080/internal/v1/payment-schema/health
```

Important beginner note:

`/healthz` only tells you HTTP server is alive. `TASK_FILE_NAME` verification ke liye `/internal/v1/payment-schema/health` use karo because wo MySQL schema readiness check karta hai.

## 11. Common Errors & Fixes

Generic setup errors already documented in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Section:

```text
12. Common Errors & Fixes
```

`TASK_FILE_NAME` specific errors:

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `Unknown database 'payment_db'` | `001_create_payment_tables.up.sql` run nahi hua or DSN wrong DB name point kar raha hai | Run baseline migration or fix DSN database name | Start from migration step before server |
| Schema health shows missing table | One or more tables from migration missing hain | Apply `001_create_payment_tables.up.sql`; for current repo apply all `.up.sql` files | Use one repeatable migration command |
| Schema health shows missing index | Later migration missing ho sakti hai, especially `002`, `004`, or `005` | Apply all `.up.sql` migrations in order | Use migration runner or documented order |
| Schema health shows missing FK | FK creation fail hua, usually parent table missing or wrong engine | Ensure all tables use InnoDB and migration order correct | Do not edit migration order manually |
| Duplicate key error on payment insert | Same `(provider, idempotency_key)` already exists | Reuse existing payment response or generate correct operation-specific idempotency key | Idempotency key deterministic rakho per operation |
| Duplicate key error on webhook insert | Same provider webhook event already stored | Treat as successful no-op if already processed | Keep `(provider, provider_event_id)` unique |
| Cannot add foreign key constraint | Referenced column/index mismatch, table missing, or engine issue | Use committed migration as-is and ensure `payments(payment_id)` exists first | Avoid manual partial schema edits |
| `Invalid JSON text` | `raw_provider_response`, `payload`, or `details` me invalid JSON insert kiya | Valid JSON use karo or NULL where allowed | App layer me JSON marshal use karo |
| Migration says table already exists | Migration already applied or partially applied | Inspect `SHOW TABLES;` and avoid re-running destructive changes | Use migration tracking tool in future |
| Current schema health fails after only `001` | Current code expects later migrations too | Apply `002` to `005` as well | For full service, always apply all current `.up.sql` files |

## 12. Security & Best Practices

### `TASK_FILE_NAME` security rules

| Rule | Why it matters |
|---|---|
| Do not store card number, CVV, OTP, or provider secrets | Payment compliance and user safety |
| Store amount in minor units | Floating point money bugs avoid hote hain |
| Keep `idempotency_key` unique per operation | Duplicate payment/refund prevent hota hai |
| Keep provider webhook event IDs unique | Duplicate webhook status update avoid hota hai |
| Sanitize JSON payloads | Provider payload me sensitive data aa sakta hai |
| Use non-root DB user outside local dev | `root:root` only local convenience hai |
| Back up `payment_db` | Payment records financial audit ke liye critical hain |
| Apply schema changes through migrations only | Manual DB edits team environments me drift create karte hain |

### DB user best practice

Previous guide documents local `root:root` setup. Production/staging me dedicated user use karo:

```sql
CREATE USER 'payment_app'@'%' IDENTIFIED BY 'replace-with-strong-password';
GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, ALTER, INDEX, REFERENCES
ON payment_db.* TO 'payment_app'@'%';
FLUSH PRIVILEGES;
```

Then update:

```env
PAYMENT_MYSQL_DSN=payment_app:replace-with-strong-password@tcp(127.0.0.1:3306)/payment_db?parseTime=true&charset=utf8mb4,utf8&loc=UTC
```

Security note:

Migration user and runtime app user can be separate in production. Migration user may need `CREATE`, `ALTER`, `INDEX`, and `REFERENCES`; runtime user usually should have narrower permissions.

## 13. Missing or Misconfigured Things

`TASK_FILE_NAME` setup audit:

| Item | Current status | Risk | Suggested fix |
|---|---|---|---|
| Migration tracking table/tool | Not found | Manual SQL may be applied twice or out of order | Add migration runner like `golang-migrate`, Goose, or a Make target |
| Service `.env.example` | Not found | Beginners may not know DSN/env names | Add safe template with placeholders |
| Service Dockerfile/compose | Not found | Local setup differs per developer | Add payment-service Dockerfile and compose profile |
| Non-root local DB user | Not documented as default | People may copy `root:root` beyond local dev | Add dev/staging/prod credential guidance |
| Schema health expects current schema | Current code includes later migrations | Running only `TASK_FILE_NAME` baseline can show missing later fields | Document all-migrations command for current service |
| Raw JSON payload storage | Present in schema | Sensitive provider data may be stored if not sanitized | Enforce redaction before DB insert |
| Manual migration rollback | Down files exist but no runner | Rollback can be risky manually | Use migration tool with version history |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `2. Tech Stack` | Same Go/MySQL/backend stack already explained |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3. Required Software` | Same Git, Go, MySQL, Docker, curl requirements |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Dependency Management` | Same Go modules and dependency commands |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `5. Database Setup` | MySQL install, Docker MySQL, DSN, and basic migration flow already documented |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `6. Redis / Queue / External Services` | Same external-service status: Redis/Kafka/RabbitMQ not required |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. Environment Variables` | Existing env loading and complete variable reference already documented |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `8. Docker Setup` | Same MySQL-only Docker setup reused |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `11. Ports & Networking` | Same `8080`, `3306`, and optional event receiver port notes |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `12. Common Errors & Fixes` | Generic Go/MySQL/Docker/provider troubleshooting already documented |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `13. Security & Best Practices` | Generic secret, provider, Docker, and local setup practices already documented |

## 15. Final Checklist

- [ ] Previous dependency documentation checked.
- [ ] `INPUT_FILE_PATH` reviewed.
- [ ] `OUTPUT_FILE_PATH` created without modifying original task file.
- [ ] No duplicate Go/MySQL/Docker installation documentation added.
- [ ] MySQL is running locally or through Docker.
- [ ] `payment_db` exists.
- [ ] `TASK_FILE_NAME` baseline migration `001_create_payment_tables.up.sql` applied.
- [ ] For current service run, all `.up.sql` migrations applied in order.
- [ ] `payments` table exists.
- [ ] `payment_attempts` table exists.
- [ ] `refunds` table exists.
- [ ] `payment_webhook_events` table exists.
- [ ] `payment_reconciliations` table exists.
- [ ] Important unique keys and indexes verified.
- [ ] Foreign keys verified.
- [ ] `PAYMENT_MYSQL_DSN` points to `payment_db`.
- [ ] `.env` loaded into shell if using local env file.
- [ ] `go test ./...` passes.
- [ ] `go run ./cmd/server` starts successfully.
- [ ] `/healthz` returns OK.
- [ ] `/internal/v1/payment-schema` returns schema definition.
- [ ] `/internal/v1/payment-schema/health` returns `"ready": true`.
- [ ] No card data, CVV, OTP, provider secrets, or raw sensitive payloads stored in DB.
- [ ] Local `root:root` credentials are not used outside local development.
- [ ] Logs checked for DB/schema errors.

Final beginner note:

`TASK_FILE_NAME` ka sabse important setup point hai: MySQL schema must exist before you trust `SERVICE_NAME` flows. Pehle `payment_db` and tables verify karo, phir service run karo, phir `/internal/v1/payment-schema/health` se final confirmation lo.
