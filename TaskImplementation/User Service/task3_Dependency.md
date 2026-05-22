# Project Dependency & Setup Guide

## 1. Project Overview

This document explains the setup and dependency requirements for:

```text
TaskImplementation/User Service/task3.md
```

Output file:

```text
TaskImplementation/User Service/task3_Dependency.md
```

Task 3 ka scope hai User Service ke repository layer ko implement karna. Repository layer MySQL database aur usecase/business layer ke beech adapter ka kaam karti hai.

Simple Hinglish:

Repository ka kaam hai SQL queries chalana, MySQL rows ko domain structs me convert karna, duplicate/not-found errors ko clean domain errors me map karna, aur transactional operations safe banana. Is task me business logic rewrite nahi hoti. Sirf persistence layer setup and dependencies relevant hain.

### What Task 3 Adds

| Area | Added/Used In Task 3 |
|---|---|
| Repository interfaces | `backend/services/user-service/internal/usecase/contracts.go` |
| MySQL user repository | `backend/services/user-service/internal/repository/mysql_user_repository.go` |
| MySQL address repository | `backend/services/user-service/internal/repository/mysql_address_repository.go` |
| MySQL seller/KYC repository | `backend/services/user-service/internal/repository/mysql_seller_repository.go` |
| Repository helpers | `backend/services/user-service/internal/repository/mysql_helpers.go` |
| Repository unit tests | `backend/services/user-service/internal/repository/*_test.go` |
| Runtime DB dependency | MySQL through Go `database/sql` |
| Test dependency | `github.com/DATA-DOG/go-sqlmock` |

### What Task 3 Does Not Add

| Thing | Status |
|---|---|
| New database | No. It reuses Task 2 MySQL schema. |
| New Dockerfile | No. |
| New docker-compose file | No. |
| New Redis/Kafka/RabbitMQ setup | No. |
| New environment variables | No. |
| New migration files | No. It depends on Task 2 migration. |
| Object storage for KYC files | No. Repository stores only `storage_url` metadata. |

Important:

Before testing real repository behavior against MySQL, Task 2 schema must already exist in `user_db`.

Refer:

```text
TaskImplementation/User Service/task2_Dependency.md
Section: 5. Database Setup
Section: Task 2 Schema Files
Section: Schema-Specific Verification Queries
```

---

## 2. Tech Stack

### Task 3 Technologies

| Technology | Required? | What It Is | Why This Project Uses It |
|---|---:|---|---|
| Go 1.24 | Yes | Go ek compiled backend language hai. Fast services banane ke liye use hoti hai. | User Service backend Go me implemented hai. |
| Go Modules | Yes | Go ka dependency management system. `go.mod` dependency list karta hai and `go.sum` checksums lock karta hai. | Repository package ke external libraries versioned form me manage hote hain. |
| Go workspace | Yes | `go.work` multiple Go modules ko local development me connect karta hai. | `user-service` local generated proto module ko `replace`/workspace se use karta hai. |
| `database/sql` | Yes | Go standard library ka DB abstraction package. | MySQL connection pool, queries, transactions, and row scanning ke liye. |
| `github.com/go-sql-driver/mysql` | Yes | MySQL ke liye Go database driver. | App ko MySQL se connect karna and MySQL error codes detect karna. |
| MySQL 8.x | Yes | Relational database jisme data tables, rows, indexes, and foreign keys me store hota hai. | Users, addresses, sellers, and KYC metadata structured data hai. |
| `github.com/DATA-DOG/go-sqlmock` | Test only | Go unit test library jo fake SQL DB create karti hai. | Repository unit tests real MySQL ke bina SQL behavior verify karte hain. |
| `log/slog` | Yes | Go standard structured logging package. | Repository errors ko structured way me log karne ke liye. |

### Beginner Explanation

`database/sql` ek common interface deta hai. Ye khud MySQL se baat nahi karta. Actual MySQL se connect karne ke liye driver chahiye, aur is project me driver hai:

```text
github.com/go-sql-driver/mysql
```

Simple example:

```text
database/sql = steering wheel
go-sql-driver/mysql = engine connection to MySQL
MySQL = actual database server
```

`go-sqlmock` sirf tests ke liye hai. Isse test fast hote hain because har unit test ke liye Docker MySQL start karna zaruri nahi hota.

### Reused Tech From Previous Tasks

Base Go install, Docker install, MySQL install, gRPC server setup, and base `.env` setup already documented hai.

Refer:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 2. Tech Stack
Section: 3. Required Software
Section: 4. Dependency Management
```

---

## 3. Required Software

Task 3 ke liye required software do categories me socho:

1. Unit testing setup
2. Real MySQL runtime/manual testing setup

### Required For Unit Tests

| Software | Required? | Why |
|---|---:|---|
| Go 1.24+ | Yes | Repository tests `go test` se run honge. |
| Internet or module cache | Usually yes | First time `go mod download` dependencies fetch karega. |
| MySQL server | No | `go-sqlmock` tests fake SQL DB use karte hain. |
| Docker | No | Unit tests ke liye container required nahi. |

### Required For Real Runtime Testing

| Software | Required? | Why |
|---|---:|---|
| MySQL 8.x | Yes | Actual repository code MySQL tables query karta hai. |
| MySQL client | Recommended | Tables and sample data verify karne ke liye. |
| Docker | Optional but recommended | Local MySQL quickly run karne ke liye easiest. |
| Database migration applied | Yes | Repository queries `users`, `user_addresses`, `seller_profiles`, `seller_kyc_documents` tables expect karti hain. |

### Do Not Duplicate Setup

Install steps already explained in previous files:

| Setup Need | Use This Existing Doc |
|---|---|
| Go install | `TaskImplementation/User Service/task1_Dependency.md` -> `3. Required Software` |
| Docker install | `TaskImplementation/User Service/task1_Dependency.md` -> `Install Docker` |
| MySQL install | `TaskImplementation/User Service/task1_Dependency.md` -> `5. Database Setup` |
| MySQL schema migration | `TaskImplementation/User Service/task2_Dependency.md` -> `5. Database Setup` |
| `.env` loading | `TaskImplementation/User Service/task1_Dependency.md` -> `7. Environment Variables` |

---

## 4. Dependency Management

### Go Dependency System

This project uses Go modules.

Already explained:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 4. Dependency Management
```

Important files:

| File | Purpose |
|---|---|
| `backend/services/user-service/go.mod` | User Service module dependencies |
| `backend/services/user-service/go.sum` | Dependency checksum lock file |
| `backend/go.work` | Local workspace connecting service module and generated proto module |
| `backend/shared/gen/go/go.mod` | Generated protobuf Go module |

### Task 3 Dependency Delta

Task 3 repository implementation depends mainly on these packages:

| Dependency | Type | Required At Runtime? | Task 3 Use |
|---|---|---:|---|
| `database/sql` | Go standard library | Yes | `*sql.DB`, `QueryRowContext`, `QueryContext`, `ExecContext`, `BeginTx`, row scanning |
| `github.com/go-sql-driver/mysql v1.9.3` | External Go module | Yes | MySQL driver and typed MySQL error code checks |
| `github.com/DATA-DOG/go-sqlmock v1.5.2` | External Go module | No, test only | Unit tests for SQL queries, transactions, duplicate errors, no-row behavior |
| `log/slog` | Go standard library | Yes | Structured repository error logging |

### Commands

From repository root:

```bash
cd backend/services/user-service
go mod download
go test ./internal/repository
go test ./...
```

From `backend/` workspace root:

```bash
cd backend
go test ./services/user-service/...
```

Use `go mod tidy` only when dependency imports changed:

```bash
cd backend/services/user-service
go mod tidy
```

### Why `go-sqlmock` Matters

`go-sqlmock` beginner-friendly explanation:

Real MySQL ke bina bhi repository test karna possible hota hai. Test expected SQL query, expected args, and expected result/error define karta hai. Agar repository wrong SQL chalati hai, test fail ho jata hai.

Example behavior tested by current Task 3 files:

| Test Case | What It Verifies |
|---|---|
| Find user not found | `sql.ErrNoRows` maps to `domain.ErrUserNotFound` |
| Create user duplicate | MySQL error `1062` maps to `domain.ErrDuplicateUser` |
| Batch find users | Multiple rows scan correctly |
| Set default address | Transaction begins, locks row, clears old default, sets new default, commits |
| Missing seller for KYC | MySQL foreign key error `1452` maps to `domain.ErrSellerNotFound` |

### Common Dependency Issues Specific To Task 3

| Error | Reason | Fix |
|---|---|---|
| `no required module provides package github.com/DATA-DOG/go-sqlmock` | Module dependency missing or `go.sum` stale. | Run `go mod tidy` inside `backend/services/user-service`. |
| `go: module requires go >= 1.24` | Installed Go version old hai. | Install Go 1.24+. Refer Task 1 dependency doc. |
| `missing go.sum entry` | Dependency checksum missing. | Run `go mod download` or `go mod tidy`. |
| Local generated proto import fails | Workspace/replace path issue. | Run commands from `backend/services/user-service` or `backend` where `go.work` is visible. |
| `go-sqlmock` test query mismatch | Regex expectation current SQL se match nahi kar rahi. | Test regex update karo or SQL query change verify karo. |

---

## 5. Database Setup

### Detected Database

Task 3 uses the same MySQL database from Task 2.

| Database | Required? | Purpose |
|---|---:|---|
| MySQL 8.x | Yes for runtime/manual DB tests | Repository reads/writes User Service tables. |

Full MySQL installation and schema setup already exists:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 5. Database Setup

TaskImplementation/User Service/task2_Dependency.md
Section: 5. Database Setup
```

Do not repeat those installation commands here. Follow previous docs for:

| Need | Existing Section |
|---|---|
| What MySQL is | `task1_Dependency.md` -> `Detected Database: MySQL` |
| Local MySQL install | `task1_Dependency.md` -> `D. Local Installation` |
| Docker MySQL setup | `task1_Dependency.md` -> `E. Docker Setup` |
| Docker Compose example | `task1_Dependency.md` -> `F. Docker Compose Example` |
| Start/verify MySQL | `task1_Dependency.md` -> `G. Start Commands`, `H. Verify Running` |
| DSN format | `task1_Dependency.md` -> `J. Connection String Format` |
| Credentials placement | `task1_Dependency.md` -> `K. Where To Place Credentials` |
| Task 2 migration | `task2_Dependency.md` -> `Task 2 Schema Files` |
| Schema verification queries | `task2_Dependency.md` -> `Schema-Specific Verification Queries` |

### Task 3 Repository Tables

Repository code expects these tables:

| Table | Repository File | Used For |
|---|---|---|
| `users` | `mysql_user_repository.go` | Create user, find by user ID, find by auth account ID, batch find, profile update |
| `user_addresses` | `mysql_address_repository.go` | List, create, update, soft delete, set default address |
| `seller_profiles` | `mysql_seller_repository.go` | Create seller profile, lookup by user ID or seller ID, update seller profile |
| `seller_kyc_documents` | `mysql_seller_repository.go` | Add and list KYC document metadata |

### Task 3 DB Requirements

| Requirement | Why It Matters |
|---|---|
| `user_db` must exist | DSN points to this database. |
| Task 2 migration must be applied | Repository SQL references Task 2 tables. |
| `parseTime=true` must be present in DSN | Repository scans `TIMESTAMP` columns into Go `time.Time`. |
| MySQL user must have SELECT/INSERT/UPDATE permissions | Repository performs reads and writes. |
| Foreign keys must remain enabled | Repository expects FK errors to protect invalid child records. |

### Connection String Reminder

This is not a new Task 3 variable. It is repeated only as a reminder because repository timestamp scanning depends on `parseTime=true`.

```env
USER_SERVICE_DATABASE_DSN=ecommerce_user:ecommerce_password@tcp(127.0.0.1:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC
```

Credentials placement is already documented:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 7. Environment Variables
Section: K. Where To Place Credentials
```

### Task 3 Manual DB Sanity Checks

After following Task 2 migration setup, verify these table names exist:

```sql
USE user_db;
SHOW TABLES;
```

Expected tables:

```text
users
user_addresses
seller_profiles
seller_kyc_documents
```

Check important constraints:

```sql
SHOW INDEX FROM users;
SHOW INDEX FROM user_addresses;
SHOW INDEX FROM seller_profiles;
SHOW INDEX FROM seller_kyc_documents;
```

Why this matters:

Repository duplicate error mapping depends on unique indexes. Foreign key error mapping depends on FK constraints.

---

## 6. Redis / Queue / External Services

Task 3 does not introduce Redis, Kafka, RabbitMQ, NATS, Elasticsearch, MinIO, S3, SMTP, Stripe, Twilio, Firebase, Nginx, Kubernetes, or any new external service.

| Service | Required For Task 3? | Notes |
|---|---:|---|
| Redis | No | No Redis client, env var, or cache code in repository task. |
| Kafka | No | Event publisher is future scope, not repository setup. |
| RabbitMQ/NATS | No | No queue code in current User Service repository layer. |
| MinIO/S3 | No runtime integration yet | KYC repository stores `storage_url` only. File upload/storage service is not configured in Task 3. |
| Docker | Optional | Useful for MySQL only, same as previous tasks. |

Reuse previous explanation:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 6. Redis / Queue / External Services

TaskImplementation/User Service/task2_Dependency.md
Section: 6. Redis / Queue / External Services
```

### KYC Storage Important Note

`seller_kyc_documents.storage_url` stores a reference to a file, not the file itself.

Beginner explanation:

Repository sirf ye save karti hai ki KYC document kahan stored hai. Actual PDF/image upload, S3 bucket, MinIO container, CDN, encryption, signed URLs, and antivirus scanning Task 3 me implemented nahi hain.

So do not add fake S3/MinIO env variables for Task 3.

---

## 7. Environment Variables

Task 3 introduces no new environment variables.

Use the existing env documentation:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 7. Environment Variables

TaskImplementation/User Service/task2_Dependency.md
Section: 7. Environment Variables
```

### Existing Variables Still Relevant

| Env Variable | Required? | Task 3 Relevance |
|---|---:|---|
| `USER_SERVICE_DATABASE_DSN` | Yes | Repository runtime needs MySQL connection. |
| `MYSQL_DSN` | Optional fallback | Used only if `USER_SERVICE_DATABASE_DSN` is blank. |
| `USER_SERVICE_DB_MAX_OPEN_CONNS` | Optional | Affects repository DB pool concurrency. |
| `USER_SERVICE_DB_MAX_IDLE_CONNS` | Optional | Affects idle DB connection reuse. |
| `USER_SERVICE_DB_CONN_MAX_LIFETIME` | Optional | Avoids stale long-lived DB connections. |
| `USER_SERVICE_DB_PING_TIMEOUT` | Optional | Startup DB ping timeout. |
| `USER_SERVICE_GRPC_ADDRESS` | Optional | Service startup port, not repository-specific. |
| `USER_SERVICE_GRPC_REFLECTION` | Optional | Debugging gRPC reflection, not repository-specific. |
| `USER_SERVICE_LOG_LEVEL` | Optional | Repository error logs follow service logger. |

### Task-Specific `.env` Example

No new Task 3 variables are required.

If your `.env` already follows Task 1/Task 2, keep it unchanged. The most important value remains:

```env
USER_SERVICE_DATABASE_DSN=ecommerce_user:ecommerce_password@tcp(127.0.0.1:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC
```

Warning:

Do not commit `.env` files with real passwords. DSN contains username and password.

---

## 8. Docker Setup

Task 3 adds no Dockerfile, docker-compose service, container image, volume, or network.

Docker remains useful only for running MySQL locally.

Reuse:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 8. Docker Setup
Section: 5. Database Setup -> E. Docker Setup

TaskImplementation/User Service/task2_Dependency.md
Section: 8. Docker Setup
```

### Task 3 Docker Impact

| Docker Item | New In Task 3? | Notes |
|---|---:|---|
| User Service Dockerfile | No | Service still runs directly with `go run` unless future Docker task adds image. |
| MySQL container | Reused | Same DB setup from Task 1/Task 2. |
| MySQL volume | Reused | Keep persistent volume to avoid losing schema/data. |
| Docker network | Reused/optional | Only needed if app itself runs in Docker later. |
| Healthcheck | Not added | Recommended future improvement for compose setup. |

### Beginner Rule

If you only run repository unit tests, Docker is not needed.

If you run the actual service or manual DB verification, MySQL must be running, either native or Docker.

---

## 9. Local Development Setup

This section gives the complete onboarding flow, but repeated install/migration details are referenced instead of duplicated.

### Step 1: Clone Repository

Follow:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 9. Local Development Setup -> Step 1: Clone Repository
```

### Step 2: Install Go, Docker, MySQL Tools

Follow:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 3. Required Software
```

For Task 3 unit tests, Go is enough. For runtime repository testing, MySQL is also required.

### Step 3: Download Dependencies

```bash
cd backend/services/user-service
go mod download
```

If imports were changed:

```bash
go mod tidy
```

### Step 4: Start MySQL

Follow one of the existing MySQL setup paths:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 5. Database Setup

TaskImplementation/User Service/task2_Dependency.md
Section: Step 2: Start MySQL
```

### Step 5: Apply Task 2 Migration

Task 3 repository SQL requires Task 2 tables.

Follow:

```text
TaskImplementation/User Service/task2_Dependency.md
Section: Step 3: Apply Task 2 Up Migration
Section: Step 4: Verify Tables, Indexes, And FKs
```

### Step 6: Configure Environment

No new env variables.

Follow:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 7. Environment Variables
```

Most important:

```env
USER_SERVICE_DATABASE_DSN=ecommerce_user:ecommerce_password@tcp(127.0.0.1:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC
```

### Step 7: Run Repository Unit Tests

These tests do not need real MySQL:

```bash
cd backend/services/user-service
go test ./internal/repository
```

Run all service tests:

```bash
go test ./...
```

### Step 8: Run User Service

Task 3 repository code is used when service usecases call persistence methods. Service startup instructions are already documented.

Refer:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 10. Running The Project
```

Command reminder:

```bash
cd backend/services/user-service
go run ./cmd/server
```

### Step 9: Optional Manual Repository Data Test

If you want to manually validate DB constraints, insert parent rows before child rows.

Example order:

```text
1. users
2. user_addresses
3. seller_profiles
4. seller_kyc_documents
```

Why:

`user_addresses.user_id` references `users.user_id`.
`seller_profiles.user_id` references `users.user_id`.
`seller_kyc_documents.seller_id` references `seller_profiles.seller_id`.

---

## 10. Running The Project

### Run Modes

| Mode | Requires MySQL? | Requires Docker? | Command |
|---|---:|---:|---|
| Repository unit tests | No | No | `go test ./internal/repository` |
| All User Service tests | No for current sqlmock/usecase tests | No | `go test ./...` |
| Service startup | Yes | Optional | `go run ./cmd/server` |
| Manual DB checks | Yes | Optional | MySQL client commands from Task 2 docs |

### Ports And Networking

| Service | Port | Required For Task 3? | Purpose |
|---|---:|---:|---|
| User Service gRPC | `50052` by default | Only when running service | gRPC API server from existing service startup |
| MySQL container internal | `3306` | Yes for Docker MySQL | MySQL server inside container |
| MySQL host standard | `3306` | Yes if using standard local DSN | Host app connects to MySQL |
| MySQL host alternate | `3307` | Optional | Useful when local `3306` is already busy |
| Redis | `6379` | No | Not used in Task 3 |
| Kafka | `9092` | No | Not used in Task 3 |

Port conflict handling is already documented:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: Ports & Networking
Section: Port Conflicts

TaskImplementation/User Service/task2_Dependency.md
Section: Ports And Networking
```

### Important Startup Rule

Repository code does not create tables automatically.

Correct order:

```text
1. Start MySQL
2. Run Task 2 migration
3. Set USER_SERVICE_DATABASE_DSN
4. Run tests or service
```

---

## 11. Common Errors & Fixes

Only Task 3 specific or repository-specific errors are explained here. Base Go, Docker, MySQL, DSN, and migration errors are already covered in Task 1/Task 2 dependency docs.

### `go-sqlmock` Query Expectation Failed

Example:

```text
could not match actual sql: "SELECT ..." with expected regexp
```

Reason:

Test expected SQL regex does not match actual SQL string. Repository queries are multiline, so tests often use regex like `(?s)SELECT.*FROM users`.

Fix:

- Check actual query in repository file.
- Update test regex carefully.
- Keep regex flexible for whitespace but strict for table names, WHERE clauses, and args.

### Duplicate Error Not Mapping To Domain Error

Example symptom:

```text
expected ErrDuplicateUser, got insert user: ...
```

Reason:

Repository maps duplicate errors only when error is typed as:

```go
*mysql.MySQLError{Number: 1062}
```

If a test returns plain `errors.New("duplicate")`, `isDuplicateKey` will not detect it.

Fix in tests:

```go
WillReturnError(&mysql.MySQLError{Number: 1062, Message: "Duplicate entry"})
```

### Foreign Key Error While Creating Address, Seller, Or KYC

Example:

```text
Error 1452: Cannot add or update a child row
```

Reason:

Child row references a parent row that does not exist.

| Operation | Required Parent |
|---|---|
| Create address | Matching `users.user_id` |
| Create seller profile | Matching `users.user_id` |
| Add KYC document | Matching `seller_profiles.seller_id` |

Fix:

- Insert/create parent entity first.
- Verify Task 2 migration was applied.
- Verify IDs match exactly. `user_123` and `User_123` are not the same.

### `Table 'user_db.users' doesn't exist`

Reason:

MySQL is running but Task 2 migration was not applied.

Fix:

Refer:

```text
TaskImplementation/User Service/task2_Dependency.md
Section: Step 3: Apply Task 2 Up Migration
```

### Timestamp Scan Error

Example:

```text
unsupported Scan, storing driver.Value type []uint8 into type *time.Time
```

Reason:

MySQL DSN likely missing `parseTime=true`.

Fix:

```env
USER_SERVICE_DATABASE_DSN=ecommerce_user:ecommerce_password@tcp(127.0.0.1:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC
```

This is also documented in:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: Timestamp Scan Error
```

### Default Address Transaction Lock Wait

Example:

```text
Error 1205: Lock wait timeout exceeded
```

Reason:

`SetDefaultAddress` uses a transaction and `FOR UPDATE`. If another transaction holds the same address/user rows too long, MySQL can timeout.

Fix:

- Keep transactions short.
- Do not manually hold locks in MySQL shell while testing.
- Retry at usecase/API level if this becomes a real production case.
- Check slow queries and open transactions.

### `RowsAffected() == 0` Maps To Not Found

Symptom:

Update/delete returns not found even though ID looks correct.

Reason:

Repository update/delete queries include filters like:

```sql
AND deleted_at IS NULL
AND status <> 'deleted'
```

So soft-deleted or deleted-status records are treated as not found.

Fix:

- Verify record is not soft deleted.
- Check `deleted_at` and `status`.
- Use exact `user_id`, `address_id`, or `seller_id`.

---

## 12. Security & Best Practices

### Secrets

Task 3 does not add new secrets.

Still important:

- `USER_SERVICE_DATABASE_DSN` contains DB password.
- Do not commit real `.env`.
- Use secret manager in production.
- Use separate DB users for migration/admin and runtime app where possible.

Refer:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 12. Security & Best Practices
```

### Repository Logging

Current repository logging records operation name and error type. It does not log SQL args like email, phone, address, or KYC URL.

Good practice:

- Do not log full DSN.
- Do not log KYC document URLs if they can expose private storage paths.
- Do not log phone/address fields in production error logs.

### MySQL Error Mapping

Task 3 maps infrastructure errors to domain errors:

| MySQL/SQL Condition | Domain Error |
|---|---|
| `sql.ErrNoRows` for user | `domain.ErrUserNotFound` |
| Duplicate key `1062` for user | `domain.ErrDuplicateUser` |
| Duplicate key `1062` for address | `domain.ErrDuplicateAddress` |
| Duplicate key `1062` for seller | `domain.ErrDuplicateSeller` |
| Foreign key `1452` while creating address | `domain.ErrUserNotFound` |
| Foreign key `1452` while adding KYC | `domain.ErrSellerNotFound` |

Why this is good:

Usecase and transport layers do not need to understand MySQL-specific error codes.

### Transaction Best Practices

`SetDefaultAddress` and default-address creation use transactions.

Beginner explanation:

Transaction ka matlab hai multiple DB changes ek group me run hote hain. Agar beech me error aaye, changes rollback ho jate hain.

Best practices:

- Keep transaction code short.
- Do not call slow external APIs inside transaction.
- Use context timeout at API/usecase level.
- Always rollback on error.
- Commit only after all DB operations succeed.

### KYC Metadata Safety

Task 3 stores only KYC metadata:

```text
document_type
storage_url
status
reviewed_by
reviewed_at
rejection_reason
```

Best practices for future storage implementation:

- Store files in private bucket/container.
- Use signed URLs, not public URLs.
- Encrypt sensitive documents.
- Keep audit trail for reviewer actions.
- Do not expose raw `storage_url` to unauthorized users.

---

## 13. Missing or Misconfigured Things

These are not blockers for Task 3 unit tests, but they matter for full onboarding and future production readiness.

| Finding | Impact | Suggested Fix |
|---|---|---|
| No project-level Docker Compose stack | Beginners must run MySQL manually using previous docs. | Add local compose with MySQL healthcheck in a future DevOps task. |
| No User Service Dockerfile | Cannot containerize service directly yet. | Add production multi-stage Dockerfile later. |
| Migrations are not auto-run by app | Service can connect to DB but fail on missing tables. | Add Makefile/migration runner or documented migration command. |
| No migration version tracking tool configured | Manual SQL application can become inconsistent. | Standardize `golang-migrate`, `goose`, or similar. |
| Address repository exists but current `main.go` does not wire it into `usecase.NewService` | Address repository tests can run, but service surface may not expose address flows yet. | Wire address usecases when address APIs are implemented. |
| KYC `storage_url` has no storage backend config | Metadata can be stored, but uploads/downloads are not handled. | Add S3/MinIO setup only when file upload task is implemented. |
| Single default address is enforced by transaction, not a DB unique constraint | App bugs or manual SQL could create multiple defaults. | Consider DB-level constraint strategy or cleanup job if this becomes critical. |
| No repository integration tests against real MySQL | sqlmock verifies SQL calls but not actual MySQL behavior. | Add integration tests with Docker/Testcontainers or compose profile. |
| gRPC reflection default comes from previous setup | Useful locally, risky if publicly exposed. | Set `USER_SERVICE_GRPC_REFLECTION=false` in production. |

### Hardcoded Credential Audit

No hardcoded DB username/password was found in Task 3 repository code.

Credentials are expected through:

```text
USER_SERVICE_DATABASE_DSN
MYSQL_DSN fallback
```

### Config Audit

No new config file was added by Task 3.

Repository constructors receive:

```text
*sql.DB
optional logger
```

So repository package itself does not load environment variables. Config loading remains in:

```text
backend/services/user-service/internal/config/config.go
```

---

## 14. References to Previous Dependency Files

This file intentionally avoids duplicate setup content.

| Topic | Refer |
|---|---|
| Base project overview and User Service startup | `TaskImplementation/User Service/task1_Dependency.md` |
| Go installation | `TaskImplementation/User Service/task1_Dependency.md` -> `3. Required Software` |
| Go modules, `go.mod`, `go.sum`, `go.work` | `TaskImplementation/User Service/task1_Dependency.md` -> `4. Dependency Management` |
| MySQL installation | `TaskImplementation/User Service/task1_Dependency.md` -> `5. Database Setup` |
| Docker MySQL setup | `TaskImplementation/User Service/task1_Dependency.md` -> `5. Database Setup -> E. Docker Setup` |
| Docker Compose example | `TaskImplementation/User Service/task1_Dependency.md` -> `5. Database Setup -> F. Docker Compose Example` |
| Base `.env` variables | `TaskImplementation/User Service/task1_Dependency.md` -> `7. Environment Variables` |
| Running User Service | `TaskImplementation/User Service/task1_Dependency.md` -> `10. Running The Project` |
| Base common errors | `TaskImplementation/User Service/task1_Dependency.md` -> `11. Common Errors & Fixes` |
| Task 2 schema migration | `TaskImplementation/User Service/task2_Dependency.md` -> `5. Database Setup` |
| Schema verification | `TaskImplementation/User Service/task2_Dependency.md` -> `Schema-Specific Verification Queries` |
| Task 2 DB-specific troubleshooting | `TaskImplementation/User Service/task2_Dependency.md` -> `11. Common Errors & Fixes` |

---

## 15. Final Checklist

Use this checklist for Task 3 repository setup.

| Check | Done |
|---|---|
| Read `task1_Dependency.md` for base Go/Docker/MySQL/env setup |  |
| Read `task2_Dependency.md` for schema migration setup |  |
| Go 1.24+ installed |  |
| Dependencies downloaded with `go mod download` |  |
| `backend/services/user-service/go.mod` includes MySQL driver |  |
| `backend/services/user-service/go.mod` includes `go-sqlmock` for tests |  |
| Repository unit tests run with `go test ./internal/repository` |  |
| Full service tests run with `go test ./...` |  |
| MySQL running if testing service/runtime behavior |  |
| Task 2 migration applied before real DB testing |  |
| `USER_SERVICE_DATABASE_DSN` points to `user_db` |  |
| DSN includes `parseTime=true` |  |
| `users`, `user_addresses`, `seller_profiles`, `seller_kyc_documents` tables exist |  |
| Duplicate key and foreign key behavior understood |  |
| No Redis/Kafka/S3 env variables added for Task 3 |  |
| `.env` with real password is not committed |  |

### Quick Task 3 Commands

```bash
cd backend/services/user-service
go mod download
go test ./internal/repository
go test ./...
```

### Final Beginner Rule

Unit tests ke liye Go dependencies enough hain. Real service run karne ke liye MySQL running hona chahiye, Task 2 migration applied honi chahiye, and `USER_SERVICE_DATABASE_DSN` correct hona chahiye. Agar repository errors aa rahe hain, pehle table existence, foreign keys, unique indexes, and DSN `parseTime=true` verify karo.
