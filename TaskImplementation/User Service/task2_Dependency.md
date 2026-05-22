# Project Dependency & Setup Guide

This guide is generated for:

```text
TaskImplementation/User Service/task2.md
```

Output file:

```text
TaskImplementation/User Service/task2_Dependency.md
```

Is document ka goal hai ki beginner developer Task 2 ke MySQL schema setup ko clearly samajh sake: database kaise ready karna hai, migration kaise run karni hai, schema kaise verify karna hai, aur common DB setup issues kaise debug karne hain.

Important reuse rule:

- Same User Service folder me already `task1_Dependency.md` present hai.
- Go install, Docker install, MySQL install, base `.env`, base service run, and common startup troubleshooting already detail me documented hai.
- Is file me repeated setup copy-paste nahi kiya gaya. Jahan setup same hai, wahan previous dependency file ka exact reference diya gaya hai.

---

## 1. Project Overview

Task 2 ka scope hai User Service ke domain ko MySQL schema me convert karna.

Simple Hinglish me:

Task 1 ne ye define kiya tha ki User Service profile data own karega. Task 2 ka kaam hai us profile data ke liye MySQL database tables, indexes, foreign keys, charset, migration files, rollback, and verification steps ko setup-ready banana.

Analyzed files:

| File | Purpose |
|---|---|
| `TaskImplementation/User Service/task2.md` | MySQL schema implementation guide |
| `TaskImplementation/User Service/task1_Dependency.md` | Existing full setup guide reused by this file |
| `backend/services/user-service/migrations/001_create_user_tables.up.sql` | Actual up migration file |
| `backend/services/user-service/migrations/001_create_user_tables.down.sql` | Actual rollback migration file |
| `backend/services/user-service/internal/config/config.go` | Runtime env variables and DB DSN loading |
| `backend/services/user-service/cmd/server/main.go` | MySQL connection and gRPC startup |
| `backend/services/user-service/go.mod` | Go dependencies; no new Task 2 Go dependency added |
| `backend/services/user-service/internal/repository/*.go` | Repository SQL depends on these tables |

Current Task 2 dependency status:

| Area | Status |
|---|---|
| Language/runtime | Same as Task 1: Go 1.24+ |
| Database | MySQL required |
| Schema files | Present in `backend/services/user-service/migrations/` |
| Migration runner | No dedicated migration tool configured in repo yet |
| Docker Compose | No project-level compose file present yet |
| Redis | Not required for Task 2 |
| Kafka/RabbitMQ | Not required for Task 2 |
| Object storage | Not required to run migration, but `storage_url` column is prepared for future KYC file storage |
| New env variables | None; same DB DSN variables from Task 1 |

---

## 2. Tech Stack

For full User Service tech stack, refer:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 2. Tech Stack
```

Task 2-specific technologies:

| Technology | Required? | What it is | Why Task 2 uses it |
|---|---:|---|---|
| MySQL 8.x | Yes | MySQL ek relational database hai jisme rows, columns, indexes, constraints hote hain. | User profile, addresses, seller profile, and KYC metadata structured relational data hai. |
| SQL migration files | Yes | Migration file ek repeatable SQL script hoti hai jo DB schema create/update karti hai. | Fresh developer machine par same tables create karne ke liye. |
| InnoDB engine | Yes | InnoDB MySQL ka transactional storage engine hai. | Foreign keys, transactions, row-level locking, and rollback support ke liye. |
| `utf8mb4` charset | Yes | Full Unicode character storage support. | Names, addresses, store names multilingual ho sakte hain. |
| MySQL foreign keys | Yes | Parent-child table relationship enforce karte hain. | Address without user, KYC without seller jaise invalid data ko block karne ke liye. |
| MySQL indexes | Yes | Query ko fast banane ke liye lookup structure. | `user_id`, `seller_id`, `status`, and default address lookups fast karne ke liye. |
| MySQL client CLI | Recommended | Terminal se SQL run/verify karne ka tool. | Migration apply, tables inspect, and troubleshooting ke liye. |
| Docker | Optional but recommended | Database ko container me run karne ka easy way. | Beginner local setup me MySQL install complexity kam karta hai. |
| golang-migrate CLI | Optional | Versioned migrations run karne ka tool. | Repo me currently wired nahi hai, but future production-style migrations ke liye useful hai. |

Beginner explanation:

- MySQL data ko tables me rakhta hai.
- Migration file table banati hai.
- Index search ko fast karta hai.
- Foreign key data ko clean rakhti hai.
- Docker optional hai, but local MySQL quickly run karne ke liye easiest path hai.

---

## 3. Required Software

Base software installation already explained here:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 3. Required Software
```

Task 2 ke liye required tools:

| Software | Required? | Why needed |
|---|---:|---|
| Git | Yes | Repository clone karne ke liye |
| Go 1.24+ | Yes | User Service run/build/test ke liye |
| MySQL 8.x | Yes | Task 2 schema yahi create hota hai |
| MySQL client | Yes for manual setup | Migration file run and schema verify karne ke liye |
| Docker | Optional | MySQL container run karne ke liye |
| Docker Compose | Optional | Future local stack simplify karne ke liye |
| grpcurl | Optional | Service run ke baad gRPC verify karne ke liye |
| golang-migrate | Optional | Manual SQL ke bajay versioned migrations run karne ke liye |

No new software is introduced beyond Task 1. Agar aapne Task 1 dependency guide follow kar liya hai, Task 2 ke liye main extra kaam migration run and verify karna hai.

---

## 4. Dependency Management

### Go Dependencies

Task 2 ne koi new Go package add nahi kiya.

Reuse:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 4. Dependency Management
```

Important existing dependency for Task 2 runtime:

| Dependency | Why it matters for schema setup |
|---|---|
| `github.com/go-sql-driver/mysql` | Go service MySQL DSN se connect karta hai. |
| `database/sql` | Connection pool and DB ping manage karta hai. |
| `github.com/DATA-DOG/go-sqlmock` | Repository tests real MySQL ke bina SQL behavior test karte hain. |

### SQL Migration Dependency

Task 2 ka actual dependency code nahi, SQL schema files hain:

```text
backend/services/user-service/migrations/001_create_user_tables.up.sql
backend/services/user-service/migrations/001_create_user_tables.down.sql
```

Simple rule:

- `up.sql` schema create karta hai.
- `down.sql` schema rollback/drop karta hai.
- App startup currently migration automatically run nahi karta.
- Developer ko DB ready karni hogi before service APIs ko exercise karna.

### Common Go Dependency Issues

Already covered:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: Common Go Dependency Issues
```

No new Go module issue introduced by Task 2.

---

## 5. Database Setup

### Detected Database: MySQL

MySQL required hai.

Base MySQL explanation, local install, Docker install, Docker Compose, start commands, connection string, credentials placement, and app permissions are already documented:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 5. Database Setup
```

Use these exact existing sections:

| Need | Refer previous file section |
|---|---|
| What MySQL is | `Detected Database: MySQL` -> `A. What It Is` |
| Why User Service uses MySQL | `B. Why This Project Uses It` |
| Required or optional | `C. Required Or Optional` |
| Local installation | `D. Local Installation` |
| Docker MySQL container | `E. Docker Setup` |
| Docker Compose example | `F. Docker Compose Example` |
| Start MySQL | `G. Start Commands` |
| Verify MySQL running | `H. Verify Running` |
| Default port | `I. Default Port` |
| DSN format | `J. Connection String Format` |
| Credential placement | `K. Where To Place Credentials` |
| Run migration | `L. Run Migration` |
| App user permissions | `M. App User Permissions` |

### Task 2 Schema Files

Actual migration files currently present:

```text
backend/services/user-service/migrations/001_create_user_tables.up.sql
backend/services/user-service/migrations/001_create_user_tables.down.sql
```

The up migration does these setup actions:

| Step | What it does | Why it matters |
|---:|---|---|
| 1 | `SET NAMES utf8mb4` | MySQL connection ko full Unicode mode me rakhta hai. |
| 2 | `SET time_zone = '+00:00'` | Session timestamps UTC me consistent rehte hain. |
| 3 | `CREATE DATABASE IF NOT EXISTS user_db` | User Service DB create karta hai if missing. |
| 4 | `USE user_db` | Table creation correct database me hoti hai. |
| 5 | Create `users` | Base profile table. |
| 6 | Create `user_addresses` | User address book table. |
| 7 | Create `seller_profiles` | Seller business profile table. |
| 8 | Create `seller_kyc_documents` | Seller KYC metadata table. |

### Tables Created By Task 2

| Table | Required? | Purpose | Parent dependency |
|---|---:|---|---|
| `users` | Yes | Buyer/seller/admin base profile | None |
| `user_addresses` | Yes | Shipping/billing address book | `users(user_id)` |
| `seller_profiles` | Yes | Seller store and approval profile | `users(user_id)` |
| `seller_kyc_documents` | Yes | KYC document metadata | `seller_profiles(seller_id)` |

### Schema Creation Order

Creation order important hai because foreign keys parent table pe depend karti hain:

| Order | Table | Reason |
|---:|---|---|
| 1 | `users` | Parent table |
| 2 | `user_addresses` | References `users(user_id)` |
| 3 | `seller_profiles` | References `users(user_id)` |
| 4 | `seller_kyc_documents` | References `seller_profiles(seller_id)` |

Rollback order reverse hai:

| Order | Table |
|---:|---|
| 1 | `seller_kyc_documents` |
| 2 | `seller_profiles` |
| 3 | `user_addresses` |
| 4 | `users` |

Reason:

Child tables pehle drop hote hain, parent table baad me. Warna MySQL foreign key error de sakta hai.

### Database Name

Task 2 expects:

```text
user_db
```

Connection string me bhi same DB name hona chahiye:

```env
USER_SERVICE_DATABASE_DSN=ecommerce_user:ecommerce_password@tcp(127.0.0.1:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC
```

If you follow Task 2's optional isolated Docker example using host port `3307`, use:

```env
USER_SERVICE_DATABASE_DSN=ecommerce_user:ecommerce_password@tcp(127.0.0.1:3307)/user_db?parseTime=true&charset=utf8mb4&loc=UTC
```

Important:

- MySQL container internal port is still `3306`.
- Host port can be `3306` or `3307`.
- DSN must use the host port you exposed.

### Schema-Specific Verification Queries

After migration, connect to MySQL and run:

```sql
SHOW DATABASES LIKE 'user_db';
USE user_db;
SHOW TABLES;
```

Expected tables:

```text
seller_kyc_documents
seller_profiles
user_addresses
users
```

Check charset and collation:

```sql
SELECT
  table_name,
  table_collation
FROM information_schema.tables
WHERE table_schema = 'user_db';
```

Expected collation:

```text
utf8mb4_unicode_ci
```

Check foreign keys:

```sql
SELECT
  table_name,
  constraint_name,
  referenced_table_name
FROM information_schema.key_column_usage
WHERE table_schema = 'user_db'
  AND referenced_table_name IS NOT NULL;
```

Expected:

| Table | References |
|---|---|
| `user_addresses` | `users` |
| `seller_profiles` | `users` |
| `seller_kyc_documents` | `seller_profiles` |

Check indexes:

```sql
SHOW INDEX FROM users;
SHOW INDEX FROM user_addresses;
SHOW INDEX FROM seller_profiles;
SHOW INDEX FROM seller_kyc_documents;
```

### Credentials Placement

Same as Task 1:

```text
backend/services/user-service/.env
```

Refer:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: K. Where To Place Credentials
```

Task 2 does not add new credentials. It only requires that the existing MySQL DSN points to the database where this schema was created.

---

## 6. Redis / Queue / External Services

Task 2 does not require Redis, Kafka, RabbitMQ, NATS, Elasticsearch, SMTP, Stripe, Twilio, Firebase, or Kubernetes.

Reuse previous external-service explanation:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 6. Redis / Queue / External Services
```

Task 2-specific note:

| Service | Required now? | Why mentioned |
|---|---:|---|
| Redis | No | No Redis connection in current User Service startup. |
| Kafka/RabbitMQ | No | User events are future scope, not Task 2 schema setup. |
| Object storage / S3 / MinIO | No for migration | `seller_kyc_documents.storage_url` stores future document location, but no object storage service is needed just to create tables. |
| Docker | Optional | Useful only for running MySQL locally. |

Important KYC storage note:

`storage_url` should be treated as sensitive metadata. In production, prefer private object keys or signed URL flow. Public permanent URLs for KYC documents are unsafe.

---

## 7. Environment Variables

Task 2 introduces no new env variables.

Use the existing env docs:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 7. Environment Variables
```

Existing variables still relevant:

| Env variable | Required? | Task 2 relevance |
|---|---:|---|
| `USER_SERVICE_DATABASE_DSN` | Yes | Must point to `user_db` where Task 2 schema exists. |
| `MYSQL_DSN` | Optional fallback | Used only if `USER_SERVICE_DATABASE_DSN` is blank. |
| `USER_SERVICE_GRPC_ADDRESS` | Optional | Service port; default `:50052`. |
| `USER_SERVICE_DB_MAX_OPEN_CONNS` | Optional | DB connection pool tuning. |
| `USER_SERVICE_DB_MAX_IDLE_CONNS` | Optional | DB connection pool tuning. |
| `USER_SERVICE_DB_CONN_MAX_LIFETIME` | Optional | DB connection reuse lifetime. |
| `USER_SERVICE_DB_PING_TIMEOUT` | Optional | Startup DB ping timeout. |
| `USER_SERVICE_LOG_LEVEL` | Optional | Log verbosity. |

### Task-Specific `.env` Variant

Only use this if you run MySQL on host port `3307` as shown in Task 2's isolated Docker example:

```env
USER_SERVICE_DATABASE_DSN=ecommerce_user:ecommerce_password@tcp(127.0.0.1:3307)/user_db?parseTime=true&charset=utf8mb4&loc=UTC
```

If you use Task 1's standard Docker setup on host port `3306`, keep:

```env
USER_SERVICE_DATABASE_DSN=ecommerce_user:ecommerce_password@tcp(127.0.0.1:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC
```

Beginner warning:

`parseTime=true` remove mat karo. Without this, Go code timestamp columns ko `time.Time` me scan karte time error de sakta hai.

---

## 8. Docker Setup

Full Docker installation and MySQL container setup already exists:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 8. Docker Setup
Section: 5. Database Setup -> E. Docker Setup
```

Task 2 adds only one practical choice:

| Choice | Host port | When to use |
|---|---:|---|
| Standard Task 1 setup | `3306` | Agar machine par MySQL already nahi chal raha. |
| Task 2 isolated setup | `3307` | Agar `3306` already occupied hai by local MySQL or another container. |

If using `3307`, Docker port mapping should expose:

```text
3307:3306
```

Meaning:

- Left side `3307` is your laptop/host port.
- Right side `3306` is MySQL container port.
- DSN must use `127.0.0.1:3307`.

No new Dockerfile, Docker Compose service, volume, or network has been added by Task 2.

---

## 9. Local Development Setup

For complete clone-to-run onboarding, refer:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 9. Local Development Setup
```

Task 2 incremental setup flow:

### Step 1: Confirm Migration Files Exist

From repo root:

```bash
ls backend/services/user-service/migrations
```

Expected:

```text
001_create_user_tables.down.sql
001_create_user_tables.up.sql
```

### Step 2: Start MySQL

Use the MySQL setup from Task 1:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: Step 4: Start MySQL
```

### Step 3: Apply Task 2 Up Migration

Use the migration command already documented in Task 1:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: L. Run Migration
```

Make sure you apply this file:

```text
backend/services/user-service/migrations/001_create_user_tables.up.sql
```

### Step 4: Verify Tables, Indexes, And FKs

Run the schema-specific verification queries from this file:

```text
Section: 5. Database Setup -> Schema-Specific Verification Queries
```

### Step 5: Configure `.env`

Use existing `.env` guidance:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: Step 6: Create .env
```

Only adjust host port if you used `3307`.

### Step 6: Run User Service

Use existing run commands:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 10. Running The Project
```

Important:

The Go service pings MySQL on startup, but it does not automatically create tables. A DB can be reachable while tables are still missing. Migration is still required before APIs/repository flows work correctly.

---

## 10. Running The Project

Task 2 does not introduce a new executable service. Running flow is same as Task 1:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 10. Running The Project
```

Task 2-specific run order:

| Order | Action | Why |
|---:|---|---|
| 1 | Start MySQL | Database must accept connections. |
| 2 | Apply `001_create_user_tables.up.sql` | Required tables must exist. |
| 3 | Set `USER_SERVICE_DATABASE_DSN` | Service must know DB credentials and database name. |
| 4 | Run User Service | gRPC service starts and connects to MySQL. |
| 5 | Verify gRPC/reflection/tests | Confirms service process is healthy. |

### Ports And Networking

| Service | Port | Purpose | Notes |
|---|---:|---|---|
| User Service gRPC | `50052` | Internal gRPC API | Default from `USER_SERVICE_GRPC_ADDRESS=:50052`. |
| MySQL container internal | `3306` | MySQL server inside container | Always `3306` inside MySQL container. |
| MySQL host standard | `3306` | Local app connects to MySQL | Used by Task 1 standard setup. |
| MySQL host alternate | `3307` | Avoid local port conflict | Used by Task 2 example if `3306` is busy. |

Port rule:

If Docker says port already allocated, either stop the other MySQL service or expose this container on `3307` and update the DSN.

---

## 11. Common Errors & Fixes

Base errors already documented:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 11. Common Errors & Fixes
```

Task 2-specific errors:

### `ERROR 1452: Cannot add or update a child row`

Cause:

Foreign key fail ho rahi hai. Example: address insert kar rahe ho but `users` table me matching `user_id` nahi hai.

Fix:

```sql
SELECT user_id FROM users WHERE user_id = 'user_123';
```

Parent row pehle insert karo, child row baad me.

### `ERROR 1062: Duplicate entry`

Cause:

Unique key violate ho rahi hai. Example same `user_id`, `auth_account_id`, `email`, `seller_id`, or `document_id` dobara insert karna.

Fix:

Check existing row:

```sql
SELECT user_id, auth_account_id, email FROM users WHERE email = 'buyer@example.com';
```

Use a new unique identifier or update existing row instead of insert.

### `Cannot drop table 'users' referenced by a foreign key constraint`

Cause:

Manual rollback wrong order me run hua. Parent table `users` drop karne se pehle child tables exist kar rahi hain.

Fix:

Use the provided down migration:

```text
backend/services/user-service/migrations/001_create_user_tables.down.sql
```

It drops child tables first.

### `ERROR 1049: Unknown database 'user_db'` while rollback

Cause:

Down migration me `USE user_db` hai, but DB create hi nahi hui ya delete ho chuki hai.

Fix:

For local dev, either run up migration first or connect to MySQL and confirm:

```sql
SHOW DATABASES LIKE 'user_db';
```

### Service starts but API fails with `Table ... doesn't exist`

Cause:

MySQL reachable hai, but Task 2 migration run nahi hui.

Fix:

Run:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: L. Run Migration
```

Then verify tables using this file's schema verification queries.

### Timestamp scan error in Go

Cause:

DSN me `parseTime=true` missing hai.

Fix:

Use:

```env
USER_SERVICE_DATABASE_DSN=ecommerce_user:ecommerce_password@tcp(127.0.0.1:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC
```

### Migration user cannot create database

Cause:

`001_create_user_tables.up.sql` has `CREATE DATABASE IF NOT EXISTS user_db`. Production/staging DB user ke paas `CREATE` permission nahi ho sakti.

Fix:

- Local dev me root/admin user se migration run kar sakte ho.
- Production me DBA/infra pipeline DB create kare.
- App runtime user ko limited table permissions do.

---

## 12. Security & Best Practices

Base security guidance already exists:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 12. Security & Best Practices
```

Task 2-specific best practices:

| Area | Recommendation |
|---|---|
| DB credentials | `.env` commit mat karo. Production me secret manager/Kubernetes Secret use karo. |
| Migration user | Migration/admin user and runtime app user separate rakho. |
| Runtime DB permissions | App user ko only needed permissions do: `SELECT`, `INSERT`, `UPDATE`, maybe controlled `DELETE` if required. |
| KYC metadata | `storage_url`, `gst_number`, `rejection_reason` logs me plain text dump mat karo. |
| Address PII | Address lines and phone numbers sensitive hain; logs and analytics me mask karo. |
| Public MySQL port | Production me MySQL internet par expose mat karo. Private network/VPC use karo. |
| Rollback | Down migration drops data. Local/test ke alawa carefully use karo. |
| Charset | `utf8mb4` keep karo so names and addresses corrupt na hon. |
| Timezone | UTC timestamps maintain karo for multi-region consistency. |

Beginner rule:

Database password code me hardcode nahi karna. Hamesha env variable or secret manager se load karo.

---

## 13. Missing or Misconfigured Things

These are setup/devops gaps found while inspecting Task 2 and current User Service files:

| Gap | Impact | Suggested fix |
|---|---|---|
| No project-level `docker-compose.yml` | Beginners ko MySQL start karne ke liye manual Docker command follow karna padega. | Future me root/local compose file add karo with MySQL healthcheck. |
| No `.env.example` | Fresh developer ko required env vars guess karne pad sakte hain. | `backend/services/user-service/.env.example` add karo with safe placeholder values. |
| No migration runner wired | App startup DB ping karta hai but schema auto-create nahi karta. | Add documented migration command or CI/CD migration job. |
| No schema version table | Plain SQL files applied manually hain; applied migration tracking missing hai. | `golang-migrate`, `goose`, or similar tool standardize karo. |
| Up migration includes `CREATE DATABASE` | Production migration user may not have CREATE permission. | Infra should create DB; table migrations can run with scoped migration user. |
| Down migration drops all user tables | Accidental run se local/staging data loss ho sakta hai. | Protect production rollback and require manual approval. |
| Object storage not configured | `storage_url` column exists but actual KYC upload/storage service not configured in Task 2. | Future task should add MinIO/S3 env variables and private file access policy. |
| One default address not DB-enforced | Direct SQL can set multiple `is_default=true` rows for same user. | Keep repository transaction, and consider stricter DB strategy if direct writes become possible. |
| No DB healthcheck endpoint in service | Orchestrators cannot check DB-backed readiness directly. | Add gRPC health/reflection/readiness pattern in future devops task. |
| Task 2 doc says migration file is future, but repo currently has files | Docs may look slightly stale to beginner. | Keep dependency guide pointing to actual current files. |

Hardcoded credential audit:

- No password is hardcoded in Go config.
- Migration files do not contain DB usernames/passwords.
- Example credentials appear only in documentation; use local-only placeholders and never commit real secrets.

---

## 14. References to Previous Dependency Files

Previous dependency files found in same folder:

```text
TaskImplementation/User Service/task1_Dependency.md
```

No other `task*_Dependency.md` files were present at generation time.

Reuse map:

| Reused topic | Previous file reference |
|---|---|
| Project overview and service startup context | `task1_Dependency.md` -> `1. Project Overview` |
| Full tech stack | `task1_Dependency.md` -> `2. Tech Stack` |
| Go/Docker/MySQL installation | `task1_Dependency.md` -> `3. Required Software` |
| Go modules and dependency commands | `task1_Dependency.md` -> `4. Dependency Management` |
| MySQL install and Docker setup | `task1_Dependency.md` -> `5. Database Setup` |
| Base migration command | `task1_Dependency.md` -> `L. Run Migration` |
| Env variable list and `.env` loading | `task1_Dependency.md` -> `7. Environment Variables` |
| Docker status and useful Docker commands | `task1_Dependency.md` -> `8. Docker Setup` |
| Clone-to-run onboarding | `task1_Dependency.md` -> `9. Local Development Setup` |
| Run service and grpcurl verification | `task1_Dependency.md` -> `10. Running The Project` |
| Common startup/debug errors | `task1_Dependency.md` -> `11. Common Errors & Fixes` |
| General security and best practices | `task1_Dependency.md` -> `12. Security & Best Practices` |

---

## 15. Final Checklist

Task 2 setup checklist:

| Check | Done |
|---|---|
| Repository cloned using Task 1 guide |  |
| Go 1.24+ installed |  |
| MySQL 8.x running locally or via Docker |  |
| Correct MySQL host port selected: `3306` or `3307` |  |
| `USER_SERVICE_DATABASE_DSN` points to `user_db` |  |
| DSN includes `parseTime=true` |  |
| Up migration file exists |  |
| Down migration file exists |  |
| `001_create_user_tables.up.sql` applied |  |
| `SHOW TABLES` shows all four Task 2 tables |  |
| Foreign keys verified in `information_schema` |  |
| Indexes verified with `SHOW INDEX` |  |
| Runtime app user has appropriate DB permissions |  |
| `.env` is local-only and not committed |  |
| User Service starts after DB setup |  |

Quick beginner reminder:

First DB run karo, then migration run karo, then `.env` set karo, then Go service start karo. Agar service MySQL se connect ho rahi hai but table errors aa rahe hain, migration step missing hai.
