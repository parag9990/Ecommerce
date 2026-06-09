# User Service - Task 3 Dependency and Setup Guide

Generated from:

```text
SERVICE_NAME=User Service
TASK_FILE_NAME=task3.md
INPUT_FILE_PATH=TaskImplementation/User Service/task3.md
OUTPUT_FILE_NAME=task3_Dependency.md
OUTPUT_FILE_PATH=TaskImplementation/User Service/task3_Dependency.md
```

This document explains the dependency, setup, environment, database, testing, and DevOps requirements for Task 3: repository implementation.

Important scope note:

- The original implementation task file was not modified.
- This file does not rewrite business logic or repository implementation.
- Previous dependency guides already explain the broad project setup. This file references those sections and only expands the new Task 3-specific requirements.

---

## 1. Previous Dependency File Reuse

Dependency files already present in the same service folder:

| File | Reused setup topics |
|---|---|
| `TaskImplementation/User Service/task1_Dependency.md` | Full tech stack, Go installation, Go modules, Go workspace, MySQL install, Docker MySQL, `.env`, ports, service startup, generic troubleshooting |
| `TaskImplementation/User Service/task2_Dependency.md` | MySQL schema setup, migration files, table creation order, rollback, schema verification, Task 2 database troubleshooting |

Do not repeat the base setup from those files. Follow them first:

```text
Refer:
TaskImplementation/User Service/task1_Dependency.md

Sections:
1. Project Tech Stack Analysis
2. Language-Specific Dependency System: Go
3. Database Analysis
4. Environment Variables
7. Docker and DevOps Setup
8. Project Run Instructions
9. Common Errors and Fixes
```

```text
Refer:
TaskImplementation/User Service/task2_Dependency.md

Sections:
4. Dependency Management
5. Database Setup
7. Environment Variables
8. Docker Setup
9. Local Development Setup
10. Running the Project
11. Common Errors & Fixes
```

Task 3 adds repository-layer setup details. That means the important new questions are:

- Are the Task 2 MySQL tables already created?
- Does the runtime DSN point to the same migrated database?
- Are Go dependencies downloaded from `go.mod`?
- Can repository unit tests run with `go-sqlmock`?
- Can repository code map MySQL duplicate-key and foreign-key errors correctly?

---

## 2. Analyzed Files

| File | Why it was checked |
|---|---|
| `TaskImplementation/User Service/task3.md` | Task 3 scope and repository implementation guide |
| `TaskImplementation/User Service/task1_Dependency.md` | Base setup documentation to reuse |
| `TaskImplementation/User Service/task2_Dependency.md` | MySQL schema and migration setup to reuse |
| `backend/services/user-service/go.mod` | Direct Go dependencies and Go version |
| `backend/go.work` | Local workspace modules |
| `backend/services/user-service/internal/usecase/contracts.go` | Repository interfaces |
| `backend/services/user-service/internal/repository/mysql_user_repository.go` | `users` table repository |
| `backend/services/user-service/internal/repository/mysql_address_repository.go` | `user_addresses` table repository and default-address transactions |
| `backend/services/user-service/internal/repository/mysql_seller_repository.go` | `seller_profiles` and `seller_kyc_documents` repositories |
| `backend/services/user-service/internal/repository/mysql_helpers.go` | MySQL error mapping, null helpers, row affected helpers, logger option |
| `backend/services/user-service/internal/repository/*_test.go` | Repository unit tests and `go-sqlmock` requirements |
| `backend/services/user-service/internal/config/config.go` | Runtime environment variables and DB pool config |
| `backend/services/user-service/cmd/server/main.go` | MySQL connection, repository wiring, gRPC startup |
| `backend/services/user-service/migrations/001_create_user_tables.up.sql` | Required table/index/foreign-key schema |
| `backend/services/user-service/migrations/001_create_user_tables.down.sql` | Rollback behavior |

---

## 3. Task 3 Dependency Summary

| Area | Task 3 status |
|---|---|
| New Go runtime dependency | No new package beyond existing `go.mod`; Task 3 actively uses `database/sql` and `github.com/go-sql-driver/mysql` |
| New Go test dependency | No new package beyond existing `go.mod`; repository tests use `github.com/DATA-DOG/go-sqlmock` |
| New database requirement | No new database beyond Task 2, but Task 2 schema must be applied before runtime repository calls work |
| New migration file | None |
| New environment variable | None |
| New Docker service | None |
| Redis | Not used |
| Kafka / RabbitMQ / NATS | Not used |
| Object storage | Not required for Task 3; KYC `storage_url` is only stored as text metadata |
| gRPC | Existing runtime surface, but Task 3 itself is repository layer |

Simple beginner explanation:

Task 3 does not ask you to install a new database or queue. It asks the service to talk to the existing MySQL schema through repository code. If Go dependencies are downloaded and the Task 2 migration is applied to MySQL, Task 3 has the main setup it needs.

---

## 4. Project Tech Stack Analysis

Full stack explanation is already documented here:

```text
Refer:
TaskImplementation/User Service/task1_Dependency.md

Section:
1. Project Tech Stack Analysis
```

Task 3-specific technologies:

| Technology | Required? | What it is | Why Task 3 uses it |
|---|---:|---|---|
| Go 1.24 compatible toolchain | Yes | Backend programming language and compiler/runtime toolchain | Repository code is written in Go |
| Go Modules | Yes | Go dependency manager using `go.mod` and `go.sum` | Downloads MySQL driver, gRPC packages, protobuf packages, and test helpers |
| Go Workspace | Yes for local repo workflow | Connects local modules under `backend/` | Allows `user-service` to use generated proto module from `backend/shared/gen/go` |
| `database/sql` | Yes | Go standard SQL abstraction | Provides `sql.DB`, `QueryRowContext`, `ExecContext`, transactions, and `sql.ErrNoRows` |
| `github.com/go-sql-driver/mysql` | Yes | MySQL driver for Go | Makes `sql.Open("mysql", dsn)` work and exposes MySQL error numbers |
| MySQL 8.x with InnoDB | Yes at runtime | Relational database with transactions and foreign keys | Repositories persist users, addresses, sellers, and KYC metadata |
| `github.com/DATA-DOG/go-sqlmock` | Test only | SQL mocking library | Repository unit tests verify SQL calls without a real MySQL server |
| `log/slog` | Yes | Go structured logging package | Repository constructors accept a logger option and log repository errors |

Not introduced by Task 3:

| Technology | Status |
|---|---|
| Redis | Not required |
| Kafka | Not required |
| RabbitMQ | Not required |
| NATS | Not required |
| MongoDB | Not required |
| Elasticsearch / Typesense | Not required |
| S3 / MinIO | Not required to run Task 3 |
| Kubernetes | Not required for local development |

---

## 5. Go Dependency Management

Go module setup is already explained here:

```text
Refer:
TaskImplementation/User Service/task1_Dependency.md

Section:
2. Language-Specific Dependency System: Go
```

Task 3 uses dependencies already declared in:

```text
backend/services/user-service/go.mod
```

Important direct dependencies for Task 3:

| Dependency | Required for | Notes |
|---|---|---|
| `github.com/go-sql-driver/mysql v1.9.3` | Runtime and repository error mapping | Used by `cmd/server/main.go` and `mysql_helpers.go` |
| `github.com/DATA-DOG/go-sqlmock v1.5.2` | Unit tests | Used by `internal/repository/*_test.go` |
| `github.com/parag/ecommerce/backend/shared/gen/go v0.0.0` | Service build | Local generated proto module via `replace` |
| `google.golang.org/grpc v1.72.2` | Service build/startup | gRPC server is wired in `cmd/server/main.go` |
| `google.golang.org/protobuf v1.36.6` | Service build/startup | Generated proto runtime |

Do not manually edit `go.mod` or `go.sum` for Task 3 setup. For a fresh clone, run the existing dependency install command:

```bash
cd backend/services/user-service
go mod download
```

If Go reports missing or stale module checksums:

```bash
cd backend/services/user-service
go mod tidy
```

If the local generated proto module is not found, use the full repo layout and run from the workspace:

```bash
cd backend
go test ./services/user-service/...
```

---

## 6. Repository Layer Files

Task 3 repository code lives in:

```text
backend/services/user-service/internal/repository/
```

| File | Runtime dependency detail |
|---|---|
| `mysql_user_repository.go` | Requires `users` table, unique keys, timestamp scanning, duplicate key mapping |
| `mysql_address_repository.go` | Requires `user_addresses` table, foreign key to `users`, soft delete column, transaction support |
| `mysql_seller_repository.go` | Requires `seller_profiles` and `seller_kyc_documents`, foreign keys, seller/KYC unique keys |
| `mysql_helpers.go` | Requires Go MySQL driver error types for duplicate-key and foreign-key detection |

Repository interfaces are defined in:

```text
backend/services/user-service/internal/usecase/contracts.go
```

These interfaces keep usecase/business logic independent from MySQL details.

---

## 7. Database Setup Required For Task 3

Full MySQL install, Docker MySQL, DSN format, and migration instructions already exist:

```text
Refer:
TaskImplementation/User Service/task1_Dependency.md

Sections:
3. Database Analysis
7. Docker and DevOps Setup
8. Project Run Instructions
```

```text
Refer:
TaskImplementation/User Service/task2_Dependency.md

Sections:
5. Database Setup
8. Docker Setup
9. Local Development Setup
10. Running the Project
```

Task 3-specific rule:

The repository layer will not create tables automatically. The Task 2 up migration must already be applied before real runtime calls are made.

Required database:

```text
user_db
```

Required tables:

| Table | Used by repository |
|---|---|
| `users` | `MySQLUserRepository` |
| `user_addresses` | `MySQLAddressRepository` |
| `seller_profiles` | `MySQLSellerRepository` |
| `seller_kyc_documents` | `MySQLSellerRepository` |

Required Task 3 database features:

| Feature | Why it matters |
|---|---|
| InnoDB engine | Needed for transactions and foreign keys |
| Foreign keys | Address creation maps missing user to `ErrUserNotFound`; KYC creation maps missing seller to `ErrSellerNotFound` |
| Unique keys | Duplicate user/seller/address/KYC rows are detected and mapped to domain duplicate errors |
| `TIMESTAMP` columns | Repository scans timestamps into Go `time.Time` |
| `deleted_at` on `user_addresses` | Address delete is a soft delete, not a physical delete |
| `is_default` on `user_addresses` | Default-address updates depend on this boolean column |

Minimal verification after migration:

```bash
mysql -h 127.0.0.1 -P 3306 -u ecommerce_user -p user_db -e "SHOW TABLES;"
```

Expected tables:

```text
seller_kyc_documents
seller_profiles
user_addresses
users
```

Detailed schema verification is already documented here:

```text
Refer:
TaskImplementation/User Service/task2_Dependency.md

Section:
5. Database Setup -> Schema Verification Queries
```

---

## 8. Environment Variables

Task 3 introduces no new environment variables.

Use the same environment setup already documented here:

```text
Refer:
TaskImplementation/User Service/task1_Dependency.md

Section:
4. Environment Variables
```

```text
Refer:
TaskImplementation/User Service/task2_Dependency.md

Section:
7. Environment Variables
```

Task 3-specific environment reminders:

| Variable | Required? | Task 3 note |
|---|---:|---|
| `USER_SERVICE_DATABASE_DSN` | Yes at runtime | Must point to migrated `user_db` |
| `MYSQL_DSN` | Optional fallback | Used only if `USER_SERVICE_DATABASE_DSN` is empty |
| `USER_SERVICE_DB_MAX_OPEN_CONNS` | Optional | Affects repository query concurrency |
| `USER_SERVICE_DB_MAX_IDLE_CONNS` | Optional | Affects idle DB connection pool size |
| `USER_SERVICE_DB_CONN_MAX_LIFETIME` | Optional | Affects how long pooled DB connections live |
| `USER_SERVICE_DB_PING_TIMEOUT` | Optional | Startup DB ping timeout |
| `USER_SERVICE_LOG_LEVEL` | Optional | Repository errors are logged through the configured logger |

Important DSN requirement:

```text
parseTime=true
```

Why:

Repository code scans MySQL `TIMESTAMP` values into Go `time.Time`. Without `parseTime=true`, timestamp scanning can fail.

Example DSN format:

```env
USER_SERVICE_DATABASE_DSN=ecommerce_user:Ecom8880User@tcp(127.0.0.1:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC
```

Do not commit real credentials in `.env`.

---

## 9. Testing Setup

Task 3 has repository unit tests in:

```text
backend/services/user-service/internal/repository/
```

The tests use `github.com/DATA-DOG/go-sqlmock`, so they do not require a real MySQL server.

Run only repository tests:

```bash
cd backend/services/user-service
go test ./internal/repository
```

Run all User Service tests:

```bash
cd backend/services/user-service
go test ./...
```

Run from workspace:

```bash
cd backend
go test ./services/user-service/...
```

What repository tests currently verify:

| Test area | Setup requirement |
|---|---|
| User not found mapping | `sql.ErrNoRows` becomes `domain.ErrUserNotFound` |
| Duplicate user mapping | MySQL error `1062` becomes `domain.ErrDuplicateUser` |
| Batch user lookup | SQL placeholders and nullable fields scan correctly |
| Set default address | Transaction begins, locks row, clears old defaults, sets new default, commits |
| Missing default address | Transaction rolls back and returns `domain.ErrAddressNotFound` |
| Soft delete address | Zero affected rows becomes `domain.ErrAddressNotFound` |
| Duplicate seller | MySQL error `1062` becomes `domain.ErrDuplicateSeller` |
| Missing seller for KYC | MySQL error `1452` becomes `domain.ErrSellerNotFound` |

If `go test ./internal/repository` fails with `unmet sql expectations`, it usually means the SQL query shape or query arguments changed but the test mock was not updated.

---

## 10. Running The Service After Task 3

Full run instructions are already documented here:

```text
Refer:
TaskImplementation/User Service/task1_Dependency.md

Section:
8. Project Run Instructions
```

Task 3-specific startup sequence:

1. Download Go dependencies.
2. Start MySQL.
3. Apply Task 2 migration.
4. Export `USER_SERVICE_DATABASE_DSN`.
5. Start the backend service.

Commands are intentionally not repeated in full here because the earlier dependency docs already provide them. The important Task 3 check is that the DSN points to the migrated database, not an empty or different MySQL instance.

At runtime, `cmd/server/main.go` wires:

```text
sql.DB -> MySQLUserRepository
sql.DB -> MySQLSellerRepository
repositories -> usecase service
usecase service -> gRPC server
```

Note:

`MySQLAddressRepository` is implemented for Task 3, but current `cmd/server/main.go` wires user and seller repositories into the current usecase service. Address repository support may be wired by later API/usecase work.

---

## 11. Docker And External Services

Task 3 does not add a Dockerfile or docker-compose file.

Docker/MySQL setup is already explained here:

```text
Refer:
TaskImplementation/User Service/task1_Dependency.md

Section:
7. Docker and DevOps Setup
```

```text
Refer:
TaskImplementation/User Service/task2_Dependency.md

Section:
8. Docker Setup
```

Task 3 external service status:

| External service | Required for Task 3? | Reason |
|---|---:|---|
| MySQL | Yes for real runtime calls | Repository persists and queries MySQL tables |
| Docker | Optional | Helpful for local MySQL only |
| Redis | No | No Redis client or env var found |
| Kafka | No | Event publishing is outside Task 3 |
| RabbitMQ | No | No queue dependency found |
| Object storage | No | KYC documents store only `storage_url` metadata |

---

## 12. Repository-Specific Common Errors And Fixes

Base troubleshooting is already documented here:

```text
Refer:
TaskImplementation/User Service/task1_Dependency.md

Section:
9. Common Errors and Fixes
```

```text
Refer:
TaskImplementation/User Service/task2_Dependency.md

Section:
11. Common Errors & Fixes
```

Task 3-specific issues:

| Error / Symptom | Likely cause | Fix |
|---|---|---|
| `Table 'user_db.users' doesn't exist` | Task 2 migration was not applied to this DB | Apply `001_create_user_tables.up.sql` and confirm DSN database |
| `Unknown column ...` | Migration/schema differs from repository queries | Reapply correct migration in dev or compare schema with Task 2 migration |
| Timestamp scan error | DSN missing `parseTime=true` | Add `parseTime=true` to `USER_SERVICE_DATABASE_DSN` |
| Duplicate user/seller/address/KYC error | Unique key conflict, MySQL error `1062` | Use unique IDs/email or handle duplicate domain error in caller |
| Missing parent row on address insert | `user_addresses.user_id` references missing `users.user_id` | Create the user first, then create address |
| Missing parent row on KYC insert | `seller_kyc_documents.seller_id` references missing seller | Create seller profile first, then add KYC metadata |
| `ErrAddressNotFound` on default address | Address is missing, belongs to another user, or is soft deleted | Verify `user_id`, `address_id`, and `deleted_at IS NULL` |
| Default address update blocks or times out | Competing transaction is locking same user's address rows | Retry after transaction completes; keep transactions short |
| `NewMySQL...Repository returned error: db is required` | Constructor received nil `*sql.DB` | Ensure DB is opened before repository construction |
| `go test` sqlmock expectation failure | SQL or arguments changed after tests were written | Update expected query regex, arguments, or result rows |
| Service starts but repository calls fail | Service can ping DB, but tables/schema are missing | Startup only pings MySQL; migration still must be applied manually |

---

## 13. Security And DevOps Notes

Task 3 repository security and operations requirements:

| Area | Guidance |
|---|---|
| DB credentials | Keep in `.env` or platform secrets, not committed source |
| Runtime DB grants | Runtime user usually needs `SELECT`, `INSERT`, `UPDATE`, `DELETE` on `user_db.*` |
| Migration DB grants | Migration/admin user needs schema permissions like `CREATE`, `ALTER`, `DROP`, indexes, and foreign keys |
| Root DB user | Do not use root credentials in app runtime DSN |
| Logs | Repository logging records operation and error type, not SQL parameters or DSN |
| KYC storage URL | Treat `storage_url` as sensitive metadata if it contains private object paths |
| Transactions | Default-address changes depend on transaction support; use InnoDB |
| Production migration | Backup before running down migration because it drops tables |

---

## 14. Beginner Setup Path For Task 3

Use this as the clean onboarding order:

| Step | What to do | Detailed reference |
|---:|---|---|
| 1 | Clone repository | `task1_Dependency.md`, section `8. Project Run Instructions -> Step 1: Clone Repository` |
| 2 | Install Go 1.24 compatible toolchain | `task1_Dependency.md`, section `2. Language-Specific Dependency System: Go` |
| 3 | Download Go dependencies | `task1_Dependency.md`, section `8. Project Run Instructions -> Step 3: Install/Download Go Dependencies` |
| 4 | Start MySQL locally or with Docker | `task1_Dependency.md`, section `3. Database Analysis`; `task2_Dependency.md`, section `8. Docker Setup` |
| 5 | Apply Task 2 migration | `task2_Dependency.md`, section `5. Database Setup -> Apply Migration` |
| 6 | Export `.env` variables | `task1_Dependency.md`, section `4. Environment Variables` |
| 7 | Run repository tests | This file, section `9. Testing Setup` |
| 8 | Start service | `task1_Dependency.md`, section `8. Project Run Instructions -> Step 7: Start Backend Service` |
| 9 | Debug setup issues | This file, section `12. Repository-Specific Common Errors And Fixes` |

---

## 15. Final Checklist

| Check | Done |
|---|---|
| Read `task1_Dependency.md` for base setup | [ ] |
| Read `task2_Dependency.md` for schema/migration setup | [ ] |
| Go 1.24 compatible toolchain installed | [ ] |
| `go mod download` completed in `backend/services/user-service` | [ ] |
| MySQL is running | [ ] |
| Task 2 up migration applied | [ ] |
| `user_db` contains all four required tables | [ ] |
| Runtime DSN points to the same migrated `user_db` | [ ] |
| DSN includes `parseTime=true` | [ ] |
| Repository tests pass with `go test ./internal/repository` | [ ] |
| Service starts after env export and DB ping succeeds | [ ] |
| No Redis/Kafka/RabbitMQ setup attempted for Task 3 | [ ] |

---

## 16. Quick Reference

Task 3 depends on:

```text
Go 1.24 compatible toolchain
Go modules from backend/services/user-service/go.mod
Go workspace in backend/go.work
MySQL 8.x compatible server
Task 2 schema migration
USER_SERVICE_DATABASE_DSN with parseTime=true
```

Task 3 does not depend on:

```text
Redis
Kafka
RabbitMQ
MongoDB
Object storage
Kubernetes
New Docker files
New environment variables
```
