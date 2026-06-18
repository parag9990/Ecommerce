# Project Dependency & Setup Guide

> Incremental dependency handbook for the MySQL database-choice task. This file intentionally references the previous dependency guide instead of duplicating common setup.

## Document Variables

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Superadmin Service` |
| `TASK_FILE_NAME` | `task2.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task2_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

## 1. Project Overview

`INPUT_FILE_PATH` selects **MySQL 8.x** as the current service's primary database because admin identities, role-permission mappings, review workflows, platform settings, and audit trails are structured and consistency-sensitive.

### Scope of this dependency guide

- MySQL-specific requirements introduced by the database decision explain karta hai.
- Existing installation, Docker, environment, migration, and run instructions ko previous dependency guide se reuse karta hai.
- Current runnable implementation aur documentation-only task ke differences clearly identify karta hai.
- Business logic, API behavior, and original task document ko modify nahi karta.

### Important current-state clarification

`INPUT_FILE_PATH` says that migrations and backend implementation are future work. Repository ke current state me later tasks ka runnable Go service, MySQL repositories, and six SQL migrations already present hain.

For an actual local run:

- Conceptual SQL snippets in `INPUT_FILE_PATH` ko execute mat karo.
- Repository migrations ko source of truth treat karo.
- Full setup ke liye previous dependency guide follow karo.

Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`5. Database Setup`, `10. Local Development Setup`, and `11. Running the Project`

## 2. Tech Stack

### New task-specific technology decision

| Technology | Required? | Why it is used | Beginner-friendly explanation |
|---|---|---|---|
| MySQL `8.x` | Required for functional use | Structured admin data, constraints, transactions, indexes, JSON settings, and audit records | MySQL ek relational database hai. Data tables me organized rehta hai aur important admin changes ko reliable way me store kar sakta hai. |
| InnoDB | Required, built into MySQL | Transactions, foreign keys, row-level locking, and crash recovery | InnoDB MySQL ka storage engine hai jo data consistency aur transaction support deta hai. |
| `utf8mb4` with `utf8mb4_unicode_ci` | Required by current migrations | Full Unicode text storage and consistent comparisons | Is charset se normal text ke saath international characters bhi safely store hote hain. |

### Already documented technologies

Go, Go modules, `database/sql`, `github.com/go-sql-driver/mysql`, Docker, HTTP runtime, and dependency commands already fully explained hain.

Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`2. Tech Stack`, `3. Required Software`, and `4. Dependency Management`

### No new external service introduced

This task does **not** introduce Redis, Kafka, RabbitMQ, NATS, MongoDB, Elasticsearch, SMTP, S3, or another queue/cache. MySQL is the only task-specific infrastructure decision.

## 3. Required Software

No new software is required beyond the previous dependency guide.

| Software | Requirement for this task | Status |
|---|---|---|
| MySQL Server | Use MySQL `8.0+`; MySQL `8.4 LTS` is a suitable local target | Reused |
| MySQL CLI/client | Needed for manual migration and verification commands | Reused |
| Go `1.26.3` | Needed by the current runnable service | Reused |
| Docker | Optional, recommended for local MySQL | Reused |

For installation and version verification, refer to:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`3. Required Software`

> **Compatibility note:** MariaDB ko automatically drop-in replacement assume mat karo. JSON, generated-column, collation, and migration behavior ko separately test kiye bina use karna risky hai.

## 4. Dependency Management

### Current implementation dependency

The current Go module already contains:

```text
github.com/go-sql-driver/mysql v1.10.0
```

No new package install is needed for this task.

`INPUT_FILE_PATH` recommends `golang-migrate/migrate` as a future tool, but it is **not currently configured or used** by the service. Isliye usko install karna current run flow ka required step nahi hai.

For `go mod download`, `go mod tidy`, build commands, and Go module troubleshooting, refer to:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`4. Dependency Management`

## 5. Database Setup

### Database decision

| Item | Current requirement |
|---|---|
| Database product | MySQL `8.x` |
| Database name | `superadmin_db` |
| Storage engine | InnoDB |
| Character set | `utf8mb4` |
| Collation | `utf8mb4_unicode_ci` |
| Default port | `3306` |
| Functional setup status | Mandatory |

### Why MySQL is mandatory

Current repositories use MySQL for:

- Admin identity and role-permission checks
- Review-task persistence and uniqueness rules
- Versioned platform settings
- Admin audit-log persistence

Without a configured database, process health can work in local mode, but protected and database-backed functionality cannot work correctly.

### Installation, Docker, credentials, and migrations

These are already documented completely. Do not repeat or invent a second setup flow.

Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:

- `5. Database Setup`
- `Run migrations`
- `Seed a local admin`
- `Verify database`
- `Rollback warning`

### Source-of-truth rule

Use the SQL files in:

```text
backend/services/superadmin-service/migrations/
```

Do **not** use the conceptual DDL examples in `INPUT_FILE_PATH` as executable migrations. Those examples differ from the current schema, including platform-setting columns, audit-log columns, indexes, constraints, and permission keys.

The repository-level `database/draw.sql` is also an older design snapshot. Current service migrations are authoritative for the runnable implementation.

### New MySQL compatibility checks

After MySQL starts, verify the selected database assumptions:

```bash
mysql -h 127.0.0.1 -P 3306 -u superadmin -p -e \
  "SELECT @@version AS version, @@version_comment AS distribution, @@default_storage_engine AS engine, @@character_set_server AS charset, @@collation_server AS collation_name;"
```

Expected direction:

- Version is MySQL `8.x`
- Default engine is `InnoDB`
- Server supports `utf8mb4`

After migrations, verify table engines and collations:

```bash
mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e \
  "SELECT TABLE_NAME, ENGINE, TABLE_COLLATION FROM information_schema.TABLES WHERE TABLE_SCHEMA='superadmin_db' ORDER BY TABLE_NAME;"
```

Verify the generated-column uniqueness migration:

```bash
mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e \
  "SHOW CREATE TABLE admin_review_tasks\G"
```

Look for `open_resource_key` and `uk_admin_review_tasks_open_resource`.

### Database ownership

The current service owns `superadmin_db`. Cross-service IDs may be stored as references, but direct reads, writes, or foreign keys into another service's database are not allowed.

## 6. Redis / Queue / External Services

No new Redis, queue, broker, or third-party integration is required by this task.

| Service | Task-specific status |
|---|---|
| Redis | Not used |
| Kafka / RabbitMQ / NATS | Not used |
| User / Order / Payment services | Unchanged; unrelated to choosing MySQL |
| API Gateway | Unchanged; required for secure production routing |

For existing external-service behavior, refer to:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`6. External Services`

### Ports and networking delta

No port is newly introduced or changed by this task.

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Current backend HTTP server | `8088` | API, health, and readiness | Reused |
| MySQL | `3306` | Current task's selected primary database | Reused from previous setup guide |
| User Service | Local `.env` example `8081` | Existing downstream admin API | Unchanged |
| Order Service | Local `.env` example `8084` | Existing downstream admin API | Unchanged |
| Payment Service | Local `.env` example `8085` | Existing downstream admin API | Unchanged |

Downstream ports have no code defaults; their configured base URLs must match the services actually running. The previous setup guide uses illustrative `8082`/`8083` examples for Order/Payment, while the current local ignored `.env` uses `8084`/`8085`.

For port-conflict checks, firewall guidance, and Docker networking, refer to:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`8. Ports & Networking`

## 7. Environment Variables

### New or changed variables

There are **no new or changed runtime environment variables** introduced by this task beyond those already documented.

### Important mismatch between task design and runnable code

`INPUT_FILE_PATH` shows future split variables such as:

```text
SUPERADMIN_DB_HOST
SUPERADMIN_DB_PORT
SUPERADMIN_DB_NAME
SUPERADMIN_DB_USER
SUPERADMIN_DB_PASSWORD
SUPERADMIN_DB_CONN_MAX_LIFETIME_SECONDS
```

The current application does **not** read those variables.

The runnable service expects:

- `SUPERADMIN_DATABASE_DSN`, or fallback `MYSQL_DSN`
- `SUPERADMIN_DB_MAX_OPEN_CONNS`
- `SUPERADMIN_DB_MAX_IDLE_CONNS`
- `SUPERADMIN_DB_CONN_MAX_LIFETIME` using Go duration format such as `5m`
- `SUPERADMIN_DB_PING_TIMEOUT`
- `SUPERADMIN_REQUIRE_DATABASE`

Do not create another full `.env` example here. Use the complete, implementation-verified example in:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`7. Environment Variables`

> **Security note:** A local ignored `.env` currently contains a non-placeholder-looking MySQL credential. It is not tracked by Git, but it must still be treated as a secret and rotated if it was ever shared or exposed.

## 8. Docker Setup

No new container, volume, network, health check, or restart policy is introduced by this task.

Use the existing MySQL Docker setup from:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`Recommended Docker setup`, `Docker Compose example for MySQL`, and `9. Docker Setup`

Task-specific Docker checks:

```bash
docker ps --filter name=superadmin-mysql
docker logs superadmin-mysql
docker exec -it superadmin-mysql mysql -u superadmin -p superadmin_db
```

> Inside another container, `127.0.0.1:3306` points to that same container. Use the MySQL container/service DNS name instead.

## 9. Local Development Setup

### Step 1: Read the reused setup first

Follow:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`5. Database Setup` and `10. Local Development Setup`

### Step 2: Enter the current service directory

```bash
cd backend/services/superadmin-service
```

### Step 3: Install dependencies

No new dependency is introduced. Run the previously documented Go dependency command.

### Step 4: Start and verify MySQL

Use the previous guide's local or Docker setup, then run the compatibility checks from section 5 of this file.

### Step 5: Configure environment

Load the existing DSN-based environment configuration. Do not use the unsupported split DB variables shown in `INPUT_FILE_PATH`.

### Step 6: Apply migrations

Apply repository `.up.sql` migrations exactly once and in filename order, using the previous guide's command.

### Step 7: Start and verify the service

Run the previous guide's test, build, start, `/healthz`, and `/readyz` checks.

### Step 8: Verify this task's database decision

Confirm:

- MySQL version is `8.x`
- Tables use InnoDB
- Tables use the expected collation
- Required generated column/index exists
- Service readiness succeeds with the configured MySQL database

## 10. Running the Project

No new start command is introduced by this task.

Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`10. Local Development Setup`, `11. Running the Project`, and `Quick Command Reference`

Task-specific success criteria:

```text
MySQL is reachable.
All current migrations are applied.
The DSN is loaded into the process environment.
/readyz returns HTTP 200 with {"status":"ready"}.
```

## 11. Common Errors & Fixes

Generic MySQL, migration, Docker, environment, and Go module errors are already documented in:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`12. Common Errors & Fixes`

Only new task-specific issues are listed below.

| Error / symptom | Cause | Fix | Prevention |
|---|---|---|---|
| Split `SUPERADMIN_DB_HOST/...` variables are set but DB remains unconfigured | Those variables are conceptual and current code does not read them | Configure `SUPERADMIN_DATABASE_DSN` and reload the shell environment | Use the implementation-verified `.env` reference |
| SQL copied from `INPUT_FILE_PATH` fails or does not match repositories | Task SQL is conceptual and differs from current migrations | Recreate disposable local DB and apply repository migrations | Treat migration files as schema source of truth |
| Migration tool command is unavailable | `golang-migrate` is recommended in task docs but not configured | Use the documented manual ordered migration flow | Add a version-aware runner as a separate implementation task |
| MariaDB behaves differently | MariaDB was treated as an unverified MySQL replacement | Use MySQL `8.x` or test every migration/query against MariaDB | Pin and document the supported database product |
| App starts locally without DB but protected features fail | Local config permits an empty DSN | Set the DSN and preferably `SUPERADMIN_REQUIRE_DATABASE=true` | Require DB in functional local profiles |

## 12. Security & Best Practices

### Task-specific security audit

| Severity | Finding | Recommended fix |
|---|---|---|
| High | Task documentation's broad database grant can allow more access than the runtime needs | Use separate migration and runtime users; runtime user should not receive DDL permissions |
| High | Broad table-wide `UPDATE` permission conflicts with the append-only audit-log goal | Use table-specific grants and prevent audit-log update/delete for the runtime identity |
| High | Production DB TLS configuration is described as a goal but not enforced by current config validation | Require verified TLS in production DSNs and deployment policy |
| Medium | No automated migration runner/version table exists | Add a migration tool and CI/CD migration stage |
| Medium | Conceptual task SQL and current migrations differ | Mark conceptual schemas clearly and keep runnable schema documentation generated from migrations |
| Medium | No real-MySQL integration test currently verifies migrations and repositories together | Add disposable MySQL integration tests in CI |
| Medium | Local ignored `.env` contains an actual-looking DB password | Keep it untracked, restrict access, and rotate on suspected exposure |

### MySQL-specific best practices

- Application ko MySQL `root` user se run mat karo.
- Migration user aur runtime user separate rakho.
- Production MySQL ko public internet par expose mat karo.
- Database connections ke liye encrypted private networking/TLS use karo.
- Audit timestamps ke liye application and database time handling ko UTC me standardize karo.
- Automated backups ke saath restore drill bhi test karo.
- MySQL version and container image ko pin karo.
- Schema changes sirf versioned migrations se apply karo.
- MySQL `8.x` se MariaDB ya another engine par switch karne se pehle compatibility tests run karo.

For broader security findings and best practices, refer to:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`13. Security & Best Practices` and `14. Missing or Misconfigured Things`

## 13. Missing or Misconfigured Things

Only MySQL-decision-specific gaps are listed here:

- [ ] No automated migration runner or schema-version table
- [ ] No separate documented migration-user versus runtime-user grant policy
- [ ] No enforced production MySQL TLS policy
- [ ] No real-MySQL migration/repository integration test
- [ ] Conceptual SQL in `INPUT_FILE_PATH` differs from current migrations
- [ ] Conceptual split DB environment variables differ from current DSN-based configuration
- [ ] Repository-level `database/draw.sql` differs from current migrations
- [ ] No tested audit-log append-only database permission policy
- [ ] Previous guide's illustrative Order/Payment ports differ from the current local ignored `.env`

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `task1_Dependency.md` | `2. Tech Stack` | Same Go runtime, driver, framework, and tools |
| `task1_Dependency.md` | `3. Required Software` | Same Go, MySQL, client, Docker, and curl requirements |
| `task1_Dependency.md` | `4. Dependency Management` | No new Go package introduced |
| `task1_Dependency.md` | `5. Database Setup` | MySQL installation, Docker, DSN, migrations, seed, and rollback already complete |
| `task1_Dependency.md` | `6. External Services` | No external-service change |
| `task1_Dependency.md` | `7. Environment Variables` | Current DSN and pool variables already documented |
| `task1_Dependency.md` | `8. Ports & Networking` | Existing ports and network behavior unchanged |
| `task1_Dependency.md` | `9. Docker Setup` | Same MySQL container and repository Docker limitations |
| `task1_Dependency.md` | `10. Local Development Setup` | Same onboarding flow |
| `task1_Dependency.md` | `11. Running the Project` | Same run and health-check commands |
| `task1_Dependency.md` | `12. Common Errors & Fixes` | Generic setup troubleshooting already covered |
| `task1_Dependency.md` | `13. Security & Best Practices` | Shared service security findings already covered |

Full reference path:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

## 15. Final Checklist

- [ ] Previous dependency documentation read first
- [ ] MySQL `8.x` selected and verified
- [ ] InnoDB, charset, and collation checks completed
- [ ] Repository migrations treated as source of truth
- [ ] Conceptual SQL from `INPUT_FILE_PATH` not used as migrations
- [ ] Current DSN-based variables used instead of unsupported split DB variables
- [ ] MySQL installed locally or running in Docker
- [ ] Dedicated non-root local DB user configured
- [ ] All current migrations applied exactly once
- [ ] Generated-column uniqueness migration verified
- [ ] Service tests and build pass
- [ ] `/healthz` and DB-backed `/readyz` checked
- [ ] Production migration/runtime user separation planned
- [ ] Production DB TLS, backups, and private networking planned
- [ ] Local credentials kept out of Git and rotated if exposed
- [ ] No duplicate setup documentation added
