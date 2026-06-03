# Project Dependency & Setup Guide

Generated for:

```text
SERVICE_NAME=User Service
TASK_FILE_NAME=task2.md
INPUT_FILE_PATH=TaskImplementation/User Service/task2.md
OUTPUT_FILE_PATH=TaskImplementation/User Service/task2_Dependency.md
```

Is guide ka goal hai Task 2 ke MySQL schema setup ko beginner-friendly way me explain karna: DB kaise ready karni hai, migration kaise apply/rollback karni hai, schema kaise verify karna hai, aur Task 2 se related setup issues kaise debug karne hain.

Important reuse note:

- Same folder me previous dependency guide already present hai: `TaskImplementation/User Service/task1_Dependency.md`.
- Go install, Go modules, MySQL local install, Docker MySQL setup, base `.env`, gRPC run commands, and generic troubleshooting already wahan detail me documented hain.
- Is file me sirf Task 2 ke new/incremental setup details explain kiye gaye hain. Same setup repeat nahi kiya gaya.

---

## 1. Project Overview

Task 2 ka scope hai User Service domain ko MySQL schema me convert karna.

Simple Hinglish me:

Task 1 ne define kiya tha ki User Service profile, address, seller profile, aur KYC metadata own karega. Task 2 us domain ke liye actual MySQL database structure ready karta hai: tables, indexes, foreign keys, up migration, down migration, and verification queries.

Analyzed files:

| File | Why checked |
|---|---|
| `TaskImplementation/User Service/task2.md` | Task 2 ka implementation/schema guide |
| `TaskImplementation/User Service/task1_Dependency.md` | Previous dependency documentation reused to avoid duplicate setup |
| `backend/services/user-service/migrations/001_create_user_tables.up.sql` | Actual schema creation migration |
| `backend/services/user-service/migrations/001_create_user_tables.down.sql` | Actual rollback migration |
| `backend/services/user-service/internal/config/config.go` | DB DSN and runtime environment variables |
| `backend/services/user-service/cmd/server/main.go` | MySQL ping and gRPC startup behavior |
| `backend/services/user-service/go.mod` | Confirms no new Go package was added for Task 2 |
| `backend/go.work` | Local Go workspace linking service and generated proto module |

Task 2 dependency summary:

| Area | Status |
|---|---|
| New Go dependency | None |
| New environment variable | None |
| New database requirement | MySQL schema tables are required |
| New migration files | `001_create_user_tables.up.sql`, `001_create_user_tables.down.sql` |
| New Docker service | None |
| Redis/Kafka/RabbitMQ | Not required for Task 2 |
| Object storage | Not required to run Task 2; `storage_url` column is future metadata only |

---

## 2. Tech Stack

Full tech stack already explained in:

```text
Refer:
TaskImplementation/User Service/task1_Dependency.md

Section:
1. Project Tech Stack Analysis
```

Task 2-specific technology focus:

| Technology | Required? | Beginner explanation | Why Task 2 uses it |
|---|---:|---|---|
| MySQL 8.x | Yes | MySQL ek relational database hai jisme data tables ke form me store hota hai. | User, address, seller, and KYC metadata relational data hai. |
| SQL migration files | Yes | Migration ek repeatable SQL script hoti hai jo DB schema create ya rollback karti hai. | Fresh setup me same tables reliably create karne ke liye. |
| InnoDB | Yes | InnoDB MySQL ka engine hai jo transactions and foreign keys support karta hai. | Parent-child relationships enforce karne ke liye. |
| `utf8mb4` | Yes | Full Unicode charset, names/address text safely store karne ke liye. | Multilingual user/store/address data avoid corruption. |
| Foreign keys | Yes | FK ensure karta hai ki child row ka parent row exist kare. | Address without user and KYC without seller block hota hai. |
| Indexes | Yes | Index query lookup ko fast banata hai. | `user_id`, `seller_id`, `status`, default address lookups fast karne ke liye. |
| MySQL CLI | Recommended | Terminal se SQL run/verify karne ka tool. | Migration apply and schema verification ke liye. |

No Redis, Kafka, RabbitMQ, MongoDB, Elasticsearch, Stripe, Firebase, SMTP, or Kubernetes dependency is introduced by this task.

---

## 3. Required Software

Base software setup already explained in:

```text
Refer:
TaskImplementation/User Service/task1_Dependency.md

Sections:
2. Language-Specific Dependency System: Go
3. Database Analysis
7. Docker and DevOps Setup
```

Task 2 does not introduce new mandatory software beyond Task 1.

| Software | Required? | Task 2 reason |
|---|---:|---|
| Git | Yes | Repo clone karne ke liye |
| Go 1.24+ | Yes for service run/test | Service code Go me hai |
| MySQL 8.x | Yes | Task 2 schema yahi create hota hai |
| MySQL client CLI | Yes for manual migration | `mysql < migration.sql` and verification queries ke liye |
| Docker | Optional | MySQL container run karne ke liye |
| grpcurl | Optional | Service start ke baad gRPC reflection verify karne ke liye |
| `golang-migrate` / `goose` | Optional future improvement | Repo me abhi wired nahi hai; future migration tracking ke liye useful |

---

## 4. Dependency Management

### Go Dependencies

Task 2 ne `go.mod` me koi new package add nahi kiya.

Reuse Go module setup from:

```text
Refer:
TaskImplementation/User Service/task1_Dependency.md

Section:
2. Language-Specific Dependency System: Go
```

Existing dependencies relevant to Task 2:

| Dependency | Why relevant |
|---|---|
| `github.com/go-sql-driver/mysql` | Go service MySQL DSN se connect karta hai. |
| `database/sql` | Standard connection pool and ping flow use hota hai. |
| `github.com/DATA-DOG/go-sqlmock` | Repository SQL tests real MySQL ke bina run karne me help karta hai. |

### SQL Migration Dependency

Task 2 ka main non-code dependency SQL migration files hain:

```text
backend/services/user-service/migrations/001_create_user_tables.up.sql
backend/services/user-service/migrations/001_create_user_tables.down.sql
```

Beginner rule:

- `up.sql` DB/tables create karta hai.
- `down.sql` tables drop karta hai.
- Service startup DB ko ping karta hai, lekin migration automatically run nahi karta.
- Isliye service flows test karne se pehle migration manually apply karni hogi.

---

## 5. Database Setup

### Reused MySQL Setup

MySQL explanation, local installation, Docker command, Docker Compose example, DSN format, credentials placement, and base migration command already documented hain:

```text
Refer:
TaskImplementation/User Service/task1_Dependency.md

Sections:
3. Database Analysis
3. Database Analysis -> D. Local Installation
3. Database Analysis -> E. Docker Setup
3. Database Analysis -> F. Docker Compose Example
3. Database Analysis -> J. Connection String Format
3. Database Analysis -> K. Where To Place Credentials
3. Database Analysis -> L. Run Migration
```

### Task 2 Database Objects

Task 2 creates/uses database:

```text
user_db
```

Tables created by the up migration:

| Table | Required? | Purpose | Parent dependency |
|---|---:|---|---|
| `users` | Yes | Base user profile | None |
| `user_addresses` | Yes | Shipping/billing address book | `users(user_id)` |
| `seller_profiles` | Yes | Seller store/profile review state | `users(user_id)` |
| `seller_kyc_documents` | Yes | Seller KYC metadata, not binary files | `seller_profiles(seller_id)` |

Creation order matters:

| Order | Table | Why |
|---:|---|---|
| 1 | `users` | Parent table |
| 2 | `user_addresses` | References `users(user_id)` |
| 3 | `seller_profiles` | References `users(user_id)` |
| 4 | `seller_kyc_documents` | References `seller_profiles(seller_id)` |

Rollback order is reverse:

| Order | Table |
|---:|---|
| 1 | `seller_kyc_documents` |
| 2 | `seller_profiles` |
| 3 | `user_addresses` |
| 4 | `users` |

Reason: Foreign key child tables pehle drop karne padte hain. Parent pehle drop karoge to MySQL FK error de sakta hai.

### Apply Migration

Use the command from previous dependency guide:

```text
Refer:
TaskImplementation/User Service/task1_Dependency.md

Section:
3. Database Analysis -> L. Run Migration
```

Task-specific file to apply:

```text
backend/services/user-service/migrations/001_create_user_tables.up.sql
```

### Rollback Migration

Rollback file:

```text
backend/services/user-service/migrations/001_create_user_tables.down.sql
```

Warning:

Down migration drops all four Task 2 tables. Local/test DB me use karo. Production/staging me accidental rollback data loss kar sakta hai.

### Schema Verification Queries

Migration ke baad MySQL me ye checks run karo:

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

Check charset/collation:

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

Expected relationship:

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

No new credentials are introduced by Task 2.

Use existing local file:

```text
backend/services/user-service/.env
```

Main DSN must point to `user_db`:

```env
USER_SERVICE_DATABASE_DSN=ecommerce_user:change_me_local_password@tcp(127.0.0.1:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC
```

Security note: DSN contains password. `.env` commit mat karo.

---

## 6. Redis / Queue / External Services

Reusable external-service analysis:

```text
Refer:
TaskImplementation/User Service/task1_Dependency.md

Section:
5. External Services Analysis
```

Task 2-specific status:

| Service | Required for Task 2? | Reason |
|---|---:|---|
| MySQL | Yes | Schema migration yahi run hoti hai. |
| Docker | Optional | MySQL container ke liye useful. |
| Redis | No | Current User Service code me Redis client/env nahi mila. |
| Kafka/RabbitMQ/NATS | No | Events later task/future scope hain. |
| S3/MinIO/Object storage | No for migration | `storage_url` column exists, but upload/storage integration wired nahi hai. |
| Kubernetes | No | Local Task 2 setup ke liye required nahi. |

KYC storage note:

`seller_kyc_documents.storage_url` ko sensitive metadata treat karo. Future me actual KYC files private bucket/object storage me hone chahiye, public permanent URLs me nahi.

---

## 7. Environment Variables

Complete env setup already documented:

```text
Refer:
TaskImplementation/User Service/task1_Dependency.md

Section:
4. Environment Variables
```

Task 2 introduces no new env variables.

Existing variables still important:

| Variable | Required? | Task 2 relevance |
|---|---:|---|
| `USER_SERVICE_DATABASE_DSN` | Yes | Must point to MySQL `user_db` where Task 2 tables exist. |
| `MYSQL_DSN` | Conditional fallback | Used only when `USER_SERVICE_DATABASE_DSN` is blank. |
| `USER_SERVICE_GRPC_ADDRESS` | Optional | Reused service port; default `:50052`. |
| `USER_SERVICE_GRPC_REFLECTION` | Optional | Useful for local grpcurl verification. |
| `USER_SERVICE_DB_MAX_OPEN_CONNS` | Optional | DB pool tuning. |
| `USER_SERVICE_DB_MAX_IDLE_CONNS` | Optional | DB pool tuning. |
| `USER_SERVICE_DB_CONN_MAX_LIFETIME` | Optional | Connection lifetime tuning. |
| `USER_SERVICE_DB_PING_TIMEOUT` | Optional | Startup DB ping timeout. |
| `USER_SERVICE_LOG_LEVEL` | Optional | Local debugging/log verbosity. |

Common Task 2 `.env` mistake:

DB can be reachable but schema missing. `USER_SERVICE_DATABASE_DSN` correct hone ke baad bhi agar tables nahi hain, service/API flows fail karenge. Migration run karna separate step hai.

---

## 8. Docker Setup

Docker install, MySQL container command, volume, health check, and useful Docker commands already covered:

```text
Refer:
TaskImplementation/User Service/task1_Dependency.md

Sections:
3. Database Analysis -> E. Docker Setup
7. Docker and DevOps Setup
```

Task 2 Docker changes:

| Docker item | Status |
|---|---|
| New User Service Dockerfile | Not added by Task 2 |
| New docker-compose file | Not added by Task 2 |
| New container | None, reuse MySQL container from Task 1 |
| New volume | None |
| New network | None |
| New health check | None |

If host port `3306` is busy, use the alternate MySQL port pattern already explained in Task 1 and update DSN to the mapped host port, for example `127.0.0.1:3307`.

---

## 9. Local Development Setup

Complete clone-to-run flow is reused from:

```text
Refer:
TaskImplementation/User Service/task1_Dependency.md

Section:
8. Project Run Instructions
```

Task 2 incremental onboarding flow:

1. Read previous dependency documentation first:

```text
TaskImplementation/User Service/task1_Dependency.md
```

2. Go to repository root and confirm migration files:

```bash
ls backend/services/user-service/migrations
```

Expected:

```text
001_create_user_tables.down.sql
001_create_user_tables.up.sql
```

3. Start MySQL using Task 1 setup.

4. Apply Task 2 up migration:

```text
backend/services/user-service/migrations/001_create_user_tables.up.sql
```

5. Verify tables, indexes, and foreign keys using section `5. Database Setup -> Schema Verification Queries` in this file.

6. Create/source `.env` using Task 1 guide.

7. Start backend service using Task 1 run commands.

8. Verify Task 2 functionality indirectly:

- Service starts without MySQL ping failure.
- `SHOW TABLES` shows all four schema tables.
- Repository/service tests can run against mocked SQL.
- gRPC methods that use `users`/`seller_profiles` have required tables available.

---

## 10. Running the Project

Task 2 does not add a new executable. Running flow is the same backend service flow from Task 1.

Correct order:

| Order | Action | Why |
|---:|---|---|
| 1 | Start MySQL | DB must accept connections. |
| 2 | Apply `001_create_user_tables.up.sql` | Tables must exist before DB-backed flows. |
| 3 | Export `.env` variables | Service needs DSN and runtime config. |
| 4 | Start User Service | gRPC server connects to MySQL. |
| 5 | Verify with MySQL queries/grpcurl/tests | Confirms schema and service are usable. |

Ports and networking:

| Service | Port | Purpose | Status |
|---|---:|---|---|
| User Service gRPC | `50052` | Internal gRPC API | Reused from Task 1 |
| MySQL | `3306` | Database connection | Reused from Task 1 |
| MySQL alternate | `3307` | Local conflict workaround | Optional reused pattern |
| Redis | `6379` | Not used by current User Service | Not required |
| Kafka | `9092` | Future eventing only | Not required |
| RabbitMQ | `5672` | Not used by current User Service | Not required |

Important:

The Go service checks DB connectivity with `PingContext`, but it does not verify every table on startup. MySQL ping success does not guarantee migration was applied.

---

## 11. Common Errors & Fixes

Generic setup errors are already covered:

```text
Refer:
TaskImplementation/User Service/task1_Dependency.md

Section:
9. Common Errors and Fixes
```

Task 2-specific issues:

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `Table 'user_db.users' doesn't exist` | MySQL running hai but migration apply nahi hui. | Run `001_create_user_tables.up.sql`. | Always run migration before service API testing. |
| `ERROR 1452: Cannot add or update a child row` | Child row insert ho raha hai without parent, e.g. address before user. | Parent `users`/`seller_profiles` row pehle insert karo. | Respect migration relationship order and app workflows. |
| `ERROR 1062: Duplicate entry` | Unique key duplicate, e.g. same `email`, `user_id`, `seller_id`. | Existing row check karo, then update or use new ID. | Treat unique IDs as immutable. |
| `Cannot drop table 'users' referenced by a foreign key constraint` | Manual rollback wrong order me hua. | Provided down migration use karo. | Child tables first, parent tables last. |
| `ERROR 1049: Unknown database 'user_db'` | DB create nahi hui ya wrong DSN/database name. | Up migration run karo or DSN me DB name fix karo. | DSN and migration database name same rakho. |
| Timestamp scan error | DSN me `parseTime=true` missing. | DSN me `parseTime=true&charset=utf8mb4&loc=UTC` add karo. | Copy Task 1 DSN format exactly. |
| Migration user cannot create database | Up migration has `CREATE DATABASE`, but user lacks permission. | Local me admin/root se migration run karo; prod me infra pipeline DB create kare. | Runtime app user and migration/admin user separate rakho. |

Quick debug SQL:

```sql
SHOW DATABASES LIKE 'user_db';
USE user_db;
SHOW TABLES;
SHOW CREATE TABLE users;
```

---

## 12. Security & Best Practices

General security guidance reused from:

```text
Refer:
TaskImplementation/User Service/task1_Dependency.md

Sections:
10. Security and Configuration Audit
11. Best Practices
```

Task 2-specific best practices:

| Area | Recommendation |
|---|---|
| DB credentials | `.env` local-only rakho; real passwords commit mat karo. |
| Runtime permissions | App user ko DDL permissions mat do; runtime ke liye scoped CRUD grants better hain. |
| Migration permissions | Migration/admin user separate rakho because schema changes powerful hote hain. |
| KYC metadata | `storage_url`, `gst_number`, `rejection_reason` logs me plain text dump mat karo. |
| Address PII | Address lines, phone, postal code sensitive hain; logs and analytics me mask karo. |
| Rollback safety | Down migration drops data; production me approval/protection ke bina run mat karo. |
| Charset | `utf8mb4_unicode_ci` keep karo so multilingual text safe rahe. |
| Timezone | UTC use karo to multi-region/debugging easier rahe. |
| Cross-service boundaries | Auth DB par foreign key mat banao; `auth_account_id` string reference enough hai. |

Beginner note:

Database schema security ka matlab sirf password secure karna nahi hota. PII fields, permissions, rollback scripts, and logs bhi security ka part hain.

---

## 13. Missing or Misconfigured Things

Setup/DevOps gaps found while analyzing Task 2:

| Gap | Current impact | Suggested fix |
|---|---|---|
| No checked-in `.env.example` | New developer required env vars guess kar sakta hai. | Add sanitized `.env.example` with fake DSN. |
| No project-level `docker-compose.yml` | MySQL startup manual Docker command par depend karta hai. | Add local compose with MySQL healthcheck. |
| No migration runner wired | Manual SQL apply karna padta hai; applied migration tracking missing hai. | Standardize `golang-migrate`, `goose`, or CI/CD migration job. |
| No schema version table | Team ko pata nahi chalega migration already applied hai ya nahi. | Migration tool se version table maintain karo. |
| Up migration creates database | Production migration user may not have `CREATE DATABASE` permission. | Infra/DBA pipeline DB create kare; table migration scoped user se run ho. |
| Down migration drops all Task 2 tables | Accidental rollback data loss kar sakta hai. | Production rollback approvals and backups mandatory rakho. |
| Object storage not configured | `storage_url` column exists, but file storage integration absent. | Future KYC task me private S3/MinIO setup add karo. |
| One default address not fully DB-enforced | Direct SQL multiple default addresses set kar sakta hai. | Repository transaction/usecase validation maintain karo; stricter DB constraint consider karo. |
| No DB-backed readiness health endpoint | Orchestrators ko schema readiness directly nahi pata chalega. | Future me gRPC health/readiness add karo. |

Hardcoded credential audit:

- Go config me DB username/password hardcoded nahi mila.
- Migration SQL files me credentials nahi hain.
- Documentation examples placeholder password use karte hain.
- Real secrets `.env`, CI secrets, or secret manager se aane chahiye.

---

## 14. References to Previous Dependency Files

Previous dependency files found in the same folder:

```text
TaskImplementation/User Service/task1_Dependency.md
```

Reuse map:

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `task1_Dependency.md` | `1. Project Tech Stack Analysis` | Same Go/MySQL/gRPC stack already explained. |
| `task1_Dependency.md` | `2. Language-Specific Dependency System: Go` | Same `go.mod`, `go.sum`, `go.work`, and Go commands. |
| `task1_Dependency.md` | `3. Database Analysis` | MySQL install, Docker setup, DSN, credentials, and base migration flow already documented. |
| `task1_Dependency.md` | `4. Environment Variables` | Task 2 adds no new env variables. |
| `task1_Dependency.md` | `5. External Services Analysis` | Redis/Kafka/Object storage status unchanged. |
| `task1_Dependency.md` | `6. Ports and Networking` | Ports are reused: gRPC `50052`, MySQL `3306`/optional `3307`. |
| `task1_Dependency.md` | `7. Docker and DevOps Setup` | Docker/MySQL local setup unchanged. |
| `task1_Dependency.md` | `8. Project Run Instructions` | Clone, dependency install, `.env`, service start, and grpcurl flow unchanged. |
| `task1_Dependency.md` | `9. Common Errors and Fixes` | Generic setup errors already covered. |
| `task1_Dependency.md` | `10. Security and Configuration Audit` | Base secrets/reflection/healthcheck guidance unchanged. |
| `task1_Dependency.md` | `11. Best Practices` | General Go/MySQL/Docker practices already covered. |

---

## 15. Final Checklist

Task 2 setup checklist:

- [ ] Previous dependency documentation checked: `task1_Dependency.md`
- [ ] Repository cloned and full folder structure present
- [ ] Go 1.24+ available if service will be run/tested
- [ ] MySQL 8.x running locally or via Docker
- [ ] Correct MySQL host port selected: `3306` or `3307`
- [ ] `USER_SERVICE_DATABASE_DSN` points to `user_db`
- [ ] DSN includes `parseTime=true`
- [ ] Up migration file exists
- [ ] Down migration file exists
- [ ] `001_create_user_tables.up.sql` applied
- [ ] `SHOW TABLES` shows `users`, `user_addresses`, `seller_profiles`, `seller_kyc_documents`
- [ ] Foreign keys verified in `information_schema`
- [ ] Indexes verified with `SHOW INDEX`
- [ ] `.env` remains local-only and is not committed
- [ ] Runtime app user permissions reviewed
- [ ] Down migration understood as destructive
- [ ] User Service starts after DB setup
- [ ] Logs checked for DB connection/schema errors
- [ ] No duplicate setup documentation added

Quick reminder:

First MySQL start karo, then Task 2 migration apply karo, then `.env` export karo, then Go service run karo. Agar MySQL connect ho raha hai but table errors aa rahe hain, almost always migration step missing hai.
