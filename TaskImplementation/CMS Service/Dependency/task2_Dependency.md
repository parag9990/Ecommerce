# Project Dependency & Setup Guide

## Document Variables

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `CMS Service` |
| `TASK_FILE_NAME` | `task2.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task2_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}` |

This guide explains dependency, setup, environment, database, and DevOps requirements for `INPUT_FILE_PATH`.

Important: shared setup is already documented in `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`. Follow that file first. This file focuses only on what `TASK_FILE_NAME` adds or clarifies: the MySQL decision for CMS data, schema expectations, connection rules, and MySQL-specific verification.

Simple goal: beginner developer ko clear ho ki MySQL kyun required hai, kaunsi tables/constraints expected hain, credentials kahan set karne hain, aur setup verify kaise karna hai without duplicate installation docs.

## 1. Project Overview

`TASK_FILE_NAME` chooses MySQL as the primary CMS database for structured seller dashboard data.

| Area | Decision |
|---|---|
| Primary database | MySQL 8+ |
| Storage engine | InnoDB |
| Character set | `utf8mb4` |
| Collation | `utf8mb4_unicode_ci` |
| Database name | `cms_db` |
| Source of truth | MySQL, not Redis/cache |
| Current backend path | `backend/services/cms-service` |

Hinglish explanation: MySQL ek relational database hai. Isme data tables, rows, constraints, indexes, aur transactions ke through manage hota hai. `${SERVICE_NAME}` me coupons, campaigns, seller staff, settings, redemptions, aur audit logs structured data hain, isliye MySQL best fit hai.

## 2. Tech Stack

Most tech stack setup is already explained in the previous dependency file.

| Technology | Status for `TASK_FILE_NAME` | What to do |
|---|---|---|
| Go | Reused | Refer `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`, section `2. Tech Stack Analysis` and `3. Go Dependency System`. |
| Go modules | Reused | Same module setup. No new `go get` needed for this task. |
| `net/http` | Reused | Same service runtime. |
| gRPC and Protobuf | Reused | Same internal RPC setup. |
| MySQL 8+ | Task focus | MySQL itself is already documented, but this file explains the CMS-specific database choices. |
| `github.com/go-sql-driver/mysql` | Reused dependency | Already present in `backend/services/cms-service/go.mod`. No new install step. |
| `database/sql` | Reused | Standard Go DB layer already used by repositories. |
| Docker | Reused optional tool | Use previous MySQL Docker setup if local MySQL is not installed. |

## 3. Required Software

Do not reinstall anything if you already completed `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`.

| Software | Required? | New for this task? | Reference |
|---|---:|---:|---|
| Go compatible with `go.mod` | Yes | No | `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`, section `3. Go Dependency System` |
| MySQL 8+ server | Yes | No, but central to this task | `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`, section `4. Database Analysis` |
| MySQL CLI client | Strongly recommended | No | `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`, section `4. Database Analysis` |
| Docker | Optional | No | `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`, section `8. Docker and DevOps Setup` |
| Product Service | Required for product moderation flows | No | `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`, section `6. External Services Analysis` |
| Redis, Kafka, RabbitMQ | No | No | Not used by current `${SERVICE_NAME}` startup. |

## 4. Dependency Management

No new Go package is introduced by `TASK_FILE_NAME`.

Current relevant module dependency:

| Dependency | Why it matters for this task |
|---|---|
| `github.com/go-sql-driver/mysql v1.9.3` | Go service ko MySQL se connect karne ke liye driver. |

Setup reuse:

```md
This setup is already explained in:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`3. Go Dependency System`
`9. Complete Project Run Instructions`
`15. Minimal Local Command Flow`
```

Beginner note: `go.mod` dependency list hai, `go.sum` checksum safety file hai. Is task ke liye in files me kuch add karne ki zarurat nahi hai.

## 5. Database Setup

### 5.1 Database used

| Database | Required? | Default port | Status |
|---|---:|---:|---|
| MySQL 8+ | Yes | 3306 | Reused setup, task-specific schema decision |

MySQL installation, Docker run command, DB user creation, and migration commands are already covered in:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`4. Database Analysis`
`8. Docker and DevOps Setup`
`9. Complete Project Run Instructions`
```

### 5.2 Why MySQL is required for this task

| CMS data | Why relational DB fits |
|---|---|
| `seller_settings` | One seller should have one settings row. Unique constraint makes this clean. |
| `seller_staff` | Seller, user, role, and status relation needs uniqueness and indexed lookup. |
| `coupons` | Coupon code uniqueness, status, validity window, and seller filtering are structured. |
| `coupon_rules` | One coupon can have many rules. Parent-child relation is clear in MySQL. |
| `coupon_redemptions` | Duplicate redemption must be blocked with unique keys and transactions. |
| `campaigns` | Time-window queries and budget/status filters need indexes. |
| `cms_audit_logs` | Audit history needs ordered, queryable, immutable-style records. |

Simple explanation: Coupon ya campaign update half-save nahi hona chahiye. MySQL transactions help karte hain ki either complete write ho, ya rollback ho.

### 5.3 Current CMS migrations to run

The previous dependency file already gives exact migration commands. For this task, understand the table purpose:

| Migration | Task relevance |
|---|---|
| `001_create_cms_access_tables.up.sql` | Creates `seller_staff` and `cms_audit_logs`, needed for permission persistence and audit trail. |
| `003_create_coupon_engine_tables.up.sql` | Creates `coupons`, `coupon_rules`, and `coupon_redemptions`, central to MySQL coupon design. |
| `004_create_offer_campaigns.up.sql` | Creates `campaigns` and adds campaign linkage to `coupon_redemptions`. |
| `006_create_seller_settings.up.sql` | Creates `seller_settings`, required for seller configuration. |
| `007_harden_cms_audit_logs.up.sql` | Adds request/actor metadata and indexes to audit logs. |

Also run the other CMS migrations listed in `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` because the current service initializes all repositories during startup.

Warning: do not run `database/draw.sql` and then blindly run all service migrations on the same database. `database/draw.sql` is a full reference DDL, while `backend/services/cms-service/migrations/` is the service migration source. Mixing both can cause duplicate table/column errors.

### 5.4 Task-specific MySQL verification

After following the previous setup and running migrations, use these checks:

```bash
cd backend/services/cms-service
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SHOW TABLES;" cms_db
```

Expected task-related tables:

```text
seller_settings
seller_staff
coupons
coupon_rules
coupon_redemptions
campaigns
cms_audit_logs
```

Check charset/collation:

```bash
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SELECT @@character_set_database, @@collation_database;" cms_db
```

Expected:

```text
utf8mb4    utf8mb4_unicode_ci
```

Check important unique/index rules:

```bash
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SHOW INDEX FROM coupons; SHOW INDEX FROM coupon_redemptions; SHOW INDEX FROM seller_staff;" cms_db
```

Look for:

| Table | Important key |
|---|---|
| `coupons` | `uk_coupons_code` |
| `coupon_redemptions` | `uk_coupon_redemptions_coupon_order` |
| `seller_staff` | `uk_seller_staff_seller_user` |

### 5.5 Connection string behavior

The service builds MySQL connection using `backend/services/cms-service/internal/database/mysql.go`.

Task-specific behavior:

| Behavior | Why it matters |
|---|---|
| `CMS_MYSQL_DSN` wins if set | Direct DSN overrides component vars like host/user/password. |
| `parseTime=true` | Go can scan MySQL timestamps into `time.Time`. |
| `utf8mb4` + `utf8mb4_unicode_ci` | Supports broad Unicode text in seller policies, metadata, and audit snapshots. |
| `CMS_DB_TIMEZONE=UTC` | Service sets UTC-oriented MySQL time behavior. |
| Startup pings MySQL | Service fails fast if DB is unavailable. |
| Pool settings are applied | `CMS_DB_MAX_OPEN_CONNS`, `CMS_DB_MAX_IDLE_CONNS`, and lifetime control DB load. |

## 6. Redis / Queue / External Services

No new Redis, Kafka, RabbitMQ, NATS, MinIO, Elasticsearch, SMTP, Stripe, Twilio, or OAuth setup is introduced by `TASK_FILE_NAME`.

| Service | Required for this task? | Notes |
|---|---:|---|
| Redis | No | Future coupon validation cache may use Redis, but MySQL remains source of truth. |
| Kafka/RabbitMQ | No | No queue dependency in current CMS startup for this task. |
| Product Service | Reused | Required for product moderation routes, already documented in `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`. |
| API Gateway/Auth RBAC | Reused | Internal auth headers and seller context already documented in `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`. |

Cache note: future Redis cache should only store short-lived coupon validation results. Do not make Redis the source of truth for coupons, campaigns, settings, or audit logs.

## 7. Environment Variables

No new environment variable is introduced by `TASK_FILE_NAME`.

Use the complete `.env` from:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Section:
`5. Environment Variables`
```

Task-specific DB variables to verify:

| Variable | Required? | Why it matters for this task |
|---|---:|---|
| `CMS_MYSQL_DSN` | Optional | If set, this should point to `cms_db` with `parseTime=true`, UTF-8, and UTC location. |
| `CMS_DB_HOST` | Required if DSN empty | MySQL host. Use `localhost` on host machine, container service name inside Docker network. |
| `CMS_DB_PORT` | Required if DSN empty | MySQL port, default `3306`. |
| `CMS_DB_NAME` | Required if DSN empty | Should be `cms_db` for CMS migrations. |
| `CMS_DB_USER` | Required if DSN empty | Dedicated CMS DB user. Do not use root in app runtime. |
| `CMS_DB_PASSWORD` | Required in production if DSN empty | Keep in `.env` locally and secret manager in production. |
| `CMS_DB_TIMEZONE` | Required | Use `UTC` to keep coupon/campaign windows predictable. |
| `CMS_DB_MAX_OPEN_CONNS` | Optional | Controls max DB concurrency. |
| `CMS_DB_MAX_IDLE_CONNS` | Optional | Must not exceed max open conns. |
| `CMS_DB_CONN_MAX_LIFETIME_SECONDS` | Optional | Recycles DB connections. |

Security note: `.env` should be created at `backend/services/cms-service/.env` for local development and must not be committed.

## 8. Docker Setup

No new Docker container, volume, network, Dockerfile, or compose service is introduced by `TASK_FILE_NAME`.

Reuse:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`4. Database Analysis` -> MySQL Docker setup
`8. Docker and DevOps Setup`
```

Task-specific Docker reminder:

| Scenario | Correct DB host |
|---|---|
| Go service running on host, MySQL in Docker | `CMS_DB_HOST=localhost` |
| Go service running in Docker Compose with MySQL service | `CMS_DB_HOST=<mysql-compose-service-name>` |

If you later add a service Dockerfile, include a DB healthcheck/wait strategy. Current server pings MySQL on startup and exits if DB is not reachable.

## 9. Local Development Setup

Follow this incremental flow:

1. Read previous dependency documentation first:

```text
TaskImplementation/${SERVICE_NAME}/task1_Dependency.md
```

2. Go to service directory:

```bash
cd backend/services/cms-service
```

3. Start MySQL using the previous dependency file.

4. Run all CMS migrations in numeric order using the previous dependency file.

5. Verify task-specific MySQL tables and indexes using section `5.4` in this file.

6. Create or update `.env` using the previous dependency file, then re-check the DB variables listed in section `7`.

7. Start the backend service using the previous dependency file.

8. Verify the CMS service health:

```bash
curl http://localhost:8087/healthz
```

9. Verify `TASK_FILE_NAME` behavior at setup level:

```bash
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SHOW TABLES LIKE 'coupons'; SHOW TABLES LIKE 'campaigns'; SHOW TABLES LIKE 'seller_settings';" cms_db
```

## 10. Running the Project

Run commands are reused from:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`9. Complete Project Run Instructions`
`15. Minimal Local Command Flow`
```

Task-specific runtime expectations:

| Check | Expected result |
|---|---|
| Service starts | Logs include `cms.mysql.connected`. |
| DB unavailable | Service logs `cms.mysql.connect_failed` and exits. |
| Missing tables | Repository calls fail later with `Table ... doesn't exist`. Run migrations. |
| Wrong timezone | Config/DSN building can fail with timezone loading error. Use `UTC`. |

Ports and networking:

| Service | Port | Purpose | Status |
|---|---:|---|---|
| CMS HTTP API | 8087 | Health and internal/seller HTTP routes | Reused |
| CMS gRPC API | 9098 | Internal CMS RPC calls | Reused |
| MySQL | 3306 | CMS database | Reused |
| Product Service | 8080 | Product moderation internal calls | Reused |

No new port is introduced by `TASK_FILE_NAME`.

## 11. Common Errors & Fixes

Generic errors like Docker daemon down, MySQL connection refused, access denied, missing `.env`, port already in use, and Go module failures are already documented in `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`, section `11. Common Errors and Fixes`.

Only `TASK_FILE_NAME`-specific setup issues are listed here:

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `ERROR 1215` or foreign key failure while running migration `004` | `coupon_redemptions` or `campaigns` dependency order is wrong. | Run migrations in numeric order from `001` onward. | Never run later migrations first. |
| Duplicate table/column errors after using `database/draw.sql` | Reference DDL and service migrations both applied to same DB. | Recreate local DB or inspect schema and apply only missing migrations. | Use service migrations as source for local service setup. |
| `CHECK constraint` errors on old DB engine | MySQL version is too old or MariaDB behaves differently. | Use MySQL 8+. | Match the task decision: MySQL 8+, not older MySQL/MariaDB unless tested. |
| Coupon insert fails with duplicate code | `coupons.code` is unique by design. | Use a different coupon code. | Treat coupon code as globally unique unless implementation changes scope. |
| Duplicate redemption insert fails | `coupon_id + order_id` unique key prevents retry double-write. | Return idempotent success or handle duplicate as already recorded. | Keep redemption writes idempotent. |
| Time-window queries behave oddly | Non-UTC session/app timezone or bad `CMS_DB_TIMEZONE`. | Use `CMS_DB_TIMEZONE=UTC`. | Store and compare coupon/campaign times in UTC. |
| `load mysql timezone` error | Invalid timezone value in `CMS_DB_TIMEZONE`. | Set `CMS_DB_TIMEZONE=UTC`. | Do not invent timezone names. |
| JSON insert/update fails | Invalid JSON for `rule_value`, `metadata`, or `settings_json`. | Validate JSON before writing. | Keep flexible fields valid JSON and frequently filtered values in normal columns. |

## 12. Security & Best Practices

Previous security guidance is reused from:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`12. Security and Configuration Audit`
`13. Best Practices`
```

Task-specific practices:

| Practice | Why |
|---|---|
| Use dedicated `cms_user`, not root | Limits damage if app credentials leak. |
| Grant only required privileges on `cms_db` | `${SERVICE_NAME}` should not access other service databases. |
| Keep `cms_db` owned by `${SERVICE_NAME}` only | Other services should use CMS HTTP/gRPC APIs, not direct DB queries. |
| Use transactions for mutation + audit writes | Coupon, campaign, staff, and settings changes should not save half-state. |
| Store money as minor units in integer columns | Avoid floating-point rounding bugs. |
| Store and compare time in UTC | Coupon/campaign validity stays predictable. |
| Keep frequently filtered fields as columns | `status`, `seller_id`, `starts_at`, and `ends_at` should be indexed columns, not hidden in JSON. |
| Treat audit logs as append-style history | Do not casually update/delete audit entries in normal workflows. |
| Use parameterized queries | Avoid SQL injection. |

Production note: current DSN builder does not force TLS by default. For production, configure MySQL TLS through a production DSN or driver TLS setup and store credentials in a secret manager.

## 13. Missing or Misconfigured Things

| Finding | Impact | Suggested fix |
|---|---|---|
| No migration tracking tool is wired into the service | Manual migration runs can be skipped or repeated. | Add a migration runner such as `golang-migrate` or a deployment migration job. |
| No service Dockerfile/compose stack for `${SERVICE_NAME}` | Full local infra is less repeatable. | Add `${SERVICE_NAME}` Dockerfile and compose with MySQL healthcheck. |
| `database/draw.sql` and service migrations both exist | Beginners may run both and create duplicate schema issues. | Document service migrations as local source of truth; keep `draw.sql` as architecture reference. |
| MySQL TLS is not enforced by config defaults | Production DB traffic may be unencrypted if infrastructure does not enforce it. | Add production DSN/TLS guidance and validation. |
| Redis cache mentioned as future scaling path, not implemented | Developers may think Redis is required now. | Keep Redis out of local setup until a task adds it. |
| `CMS_STAFF_STATUS_SOURCE` defaults to gateway | Staff revocation depends on trusted gateway headers unless switched. | Use `CMS_STAFF_STATUS_SOURCE=mysql` when seller staff table is populated and authoritative. |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `2. Tech Stack Analysis` | Same Go, HTTP, gRPC, MySQL driver, logging, Product Service, and Gateway context. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `3. Go Dependency System` | Same `go.mod`, `go.sum`, module commands, and common Go issues. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `4. Database Analysis` | MySQL install, Docker setup, user creation, migration commands, and connection variables already explained. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `5. Environment Variables` | Complete `.env` example already exists; this task adds no new env vars. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `6. External Services Analysis` | Product Service and API Gateway/Auth context are unchanged. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `7. Ports and Networking` | No new ports introduced by `TASK_FILE_NAME`. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `8. Docker and DevOps Setup` | Docker/MySQL container setup is unchanged. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `9. Complete Project Run Instructions` | Clone, install, migrate, run, and health verification flow is reused. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `11. Common Errors and Fixes` | Generic setup troubleshooting is unchanged. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `12. Security and Configuration Audit` and `13. Best Practices` | Shared service security guidance is reused; this file adds MySQL-specific notes only. |

## 15. Final Checklist

- [ ] Previous dependency documentation checked first.
- [ ] No duplicate MySQL install or Docker setup copied into this file.
- [ ] MySQL 8+ selected as `${SERVICE_NAME}` primary database.
- [ ] `cms_db` exists with `utf8mb4` and `utf8mb4_unicode_ci`.
- [ ] Dedicated `cms_user` configured for `cms_db`.
- [ ] `.env` exists at `backend/services/cms-service/.env`.
- [ ] DB variables match local MySQL setup or `CMS_MYSQL_DSN` is correct.
- [ ] All CMS migrations run in numeric order.
- [ ] Task-specific tables verified: `seller_settings`, `seller_staff`, `coupons`, `coupon_rules`, `coupon_redemptions`, `campaigns`, `cms_audit_logs`.
- [ ] Important unique keys verified for coupon code, staff mapping, and redemption idempotency.
- [ ] Service starts and logs `cms.mysql.connected`.
- [ ] `/healthz` returns successfully.
- [ ] No Redis/Kafka/RabbitMQ setup added because current task does not require them.
- [ ] Security notes reviewed for root user avoidance, UTC time, TLS, transactions, and audit logging.

Final simple summary: `TASK_FILE_NAME` does not add a new runtime package, port, queue, cache, or Docker service. It confirms MySQL 8+ as the required CMS source of truth and clarifies the schema, migrations, constraints, connection variables, and verification steps needed for a clean local setup.
