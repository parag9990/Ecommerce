# Project Dependency & Setup Guide

Input implementation file:

```text
TaskImplementation/Auth Service/task2.md
```

Output dependency file:

```text
TaskImplementation/Auth Service/task2_Dependency.md
```

This guide is for **Auth Service - Task 2: Create MySQL Schema**.

Simple goal: beginner developer ko clearly samajh aaye ki Task 2 ke schema ko local MySQL me kaise apply, verify, rollback, and debug karna hai.

Important reuse rule: Auth Service ke previous dependency guide me Go setup, full `.env`, Docker, Redis, ports, and full local run flow already documented hai. Is file me wahi content repeat nahi kiya gaya. Jahan setup same hai, wahan exact reference diya gaya hai.

---

## 1. Project Overview

Task 2 ka focus sirf **Auth Service MySQL database foundation** hai.

Task 2 introduces:

| Item | Purpose |
|---|---|
| `auth_db` | Auth Service ka owned MySQL database |
| `auth_accounts` | Account identity root table |
| `credentials` | Password hash and lockout metadata |
| `refresh_tokens` | Rotating refresh token hashes and revoke/expiry tracking |
| `otp_challenges` | OTP challenge hashes, attempts, and expiry |
| `role_assignments` | Account roles and role scope data |
| `001_create_auth_tables.up.sql` | Base schema migration |
| `001_create_auth_tables.down.sql` | Base schema rollback |

Current repository note:

- `task2.md` describes these migration files as recommended/target files.
- Current repo already contains the real files under `backend/services/auth-service/migrations/`.
- Later migrations `002` to `005` are from later Auth Service tasks. For only Task 2 schema verification, use migration `001`. For running the current full Auth Service code, apply all migrations as documented in `task1_Dependency.md`.

Refer:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Sections: 5. Database Setup, 9. Local Development Setup, 10. Running the Project
```

---

## 2. Tech Stack

Task 2 uses the same base Auth Service stack already explained in the previous dependency file.

| Technology | Required for Task 2? | Why used in Task 2 | Beginner explanation |
|---|---:|---|---|
| MySQL 8+ | Yes | Auth schema, tables, indexes, foreign keys | MySQL ek relational database hai. Data rows/columns me store hota hai, isliye auth jaise sensitive consistent data ke liye useful hai. |
| InnoDB | Yes | Foreign keys and transactions | InnoDB MySQL ka storage engine hai jo transactions and FK constraints support karta hai. |
| `utf8mb4` | Yes | Safe Unicode storage | Emails, names, metadata me full Unicode support ke liye `utf8mb4` better default hai. |
| MySQL CLI | Recommended | Migration apply/verify karne ke liye | `mysql` command se SQL files run kar sakte ho. |
| Docker | Optional | Local MySQL repeatable run karne ke liye | Docker se MySQL container me run hota hai, machine setup clean rehta hai. |
| Go modules | Not new in Task 2 | Current service code dependencies manage karta hai | Go modules dependency version lock karte hain using `go.mod` and `go.sum`. |
| `github.com/go-sql-driver/mysql` | Not new in Task 2 | Go runtime MySQL se connect karta hai | Ye Go driver app ko MySQL database se baat karne deta hai. |

Already explained:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 2. Tech Stack
Section: 4. Dependency Management
```

Task 2 does **not** introduce a new web framework, Redis setup, Kafka setup, queue broker, or third-party API.

---

## 3. Required Software

For Task 2 schema setup only:

| Software | Required? | Why |
|---|---:|---|
| MySQL Server 8+ | Yes | Schema migration run karne ke liye |
| MySQL CLI client | Yes | `.sql` files execute and verify karne ke liye |
| Docker Desktop / Docker Engine | Optional | MySQL container run karne ke liye |
| Go | Not required for migration only | Required when running Auth Service app/tests |

Do not repeat install steps here. They already exist in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 3. Required Software
Section: 5. Database Setup
Section: 8. Docker Setup
```

Quick verify commands:

```bash
mysql --version
docker --version
go version
```

Note: `go version` is only needed if you will run the service or tests after applying the schema.

---

## 4. Dependency Management

Task 2 is primarily a database schema task. It does not add a new Go package by itself.

Already documented dependency system:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 4. Dependency Management
```

Current Auth Service module files:

```text
backend/services/auth-service/go.mod
backend/services/auth-service/go.sum
backend/go.work
```

Beginner explanation:

- `go.mod` tells Go which module this service is and which external libraries are needed.
- `go.sum` stores checksums so dependency downloads are verified.
- `go mod download` downloads dependencies.
- `go mod tidy` should be used when dependencies are added/removed.

Task 2 specific note:

- For applying `001_create_auth_tables.up.sql`, Go dependency installation is not needed.
- For running `go test ./...` or `go run ./cmd/server`, follow the previous dependency file.

Refer:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: Dependency commands
Section: Common Go dependency issues
```

---

## 5. Database Setup

### A. Database detected

Database:

```text
MySQL 8+
```

Database name:

```text
auth_db
```

Migration files:

```text
backend/services/auth-service/migrations/001_create_auth_tables.up.sql
backend/services/auth-service/migrations/001_create_auth_tables.down.sql
```

### B. What MySQL is

MySQL ek relational database hai. Isme data table format me store hota hai: rows and columns. Auth Service me account, credential, token, OTP, and role data strongly consistent hona chahiye, isliye MySQL use hota hai.

### C. Why Task 2 uses MySQL

Task 2 ke tables me relationships important hain:

| Relationship | Meaning |
|---|---|
| `credentials.account_id -> auth_accounts.account_id` | Credential valid account se linked hona chahiye |
| `refresh_tokens.account_id -> auth_accounts.account_id` | Token valid account ke liye hi hona chahiye |
| `role_assignments.account_id -> auth_accounts.account_id` | Role valid account ke liye hi assign hona chahiye |

Hinglish: Auth data sensitive hai. Agar credential ya token kisi invalid account se link ho gaya, security issue ban sakta hai. Foreign keys DB level pe basic safety dete hain.

### D. Required or optional

MySQL is required.

For Task 2 only:

- MySQL required hai migration run karne ke liye.
- Redis required nahi hai.
- Notification Service required nahi hai.
- JWT keys required nahi hain.

For current full Auth Service startup:

- MySQL, Redis, `.env`, JWT keys, and other env variables are required.
- Full run steps already exist in `task1_Dependency.md`.

### E. Local installation

Do not duplicate install steps.

Refer:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 5. Database Setup
Subsections: D. Local installation, E. Docker setup, F. Start commands
```

### F. Docker setup

Docker setup is unchanged from Task 1 dependency documentation.

Refer:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 8. Docker Setup
```

Task 2 uses the same MySQL container settings:

| Setting | Value |
|---|---|
| Image | `mysql:8.4` or MySQL 8 compatible |
| Container name | `ecommerce-auth-mysql` |
| Host port | `3306` |
| Database | `auth_db` |
| Suggested app user | `auth_user` |
| Persistent volume | `ecommerce-auth-mysql-data` |

### G. MySQL user and permission guidance

If you use the Docker command from `task1_Dependency.md`, MySQL creates `auth_user` and `auth_db` automatically.

If you installed MySQL natively and user/database do not exist, create them once:

```sql
CREATE DATABASE IF NOT EXISTS auth_db
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

CREATE USER IF NOT EXISTS 'auth_user'@'%' IDENTIFIED BY 'auth_password';
GRANT SELECT, INSERT, UPDATE, DELETE ON auth_db.* TO 'auth_user'@'%';
FLUSH PRIVILEGES;
```

Important:

- Migration user often needs DDL permissions such as `CREATE`, `ALTER`, `DROP`, and `INDEX`.
- Runtime app user should be more limited, usually `SELECT`, `INSERT`, `UPDATE`, `DELETE`.
- Beginner local setup me root se migration run karna simple hai, but production me separate migration user better hota hai.

### H. Apply Task 2 migration

Run only base schema migration:

```bash
cd backend/services/auth-service
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/001_create_auth_tables.up.sql
```

Why root here?

- `001_create_auth_tables.up.sql` contains `CREATE DATABASE IF NOT EXISTS auth_db`.
- A limited app user may not have permission to create database/tables.

If database already exists and you have a dedicated migration user, that user needs DDL permissions.

### I. Verify Task 2 migration

Check database exists:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p -e "SHOW DATABASES LIKE 'auth_db';"
```

Check Task 2 tables:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p -e "SHOW TABLES FROM auth_db;"
```

Expected after Task 2 only:

```text
auth_accounts
credentials
otp_challenges
refresh_tokens
role_assignments
```

Check table columns:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p -e "DESCRIBE auth_db.auth_accounts;"
mysql -h 127.0.0.1 -P 3306 -u root -p -e "DESCRIBE auth_db.credentials;"
mysql -h 127.0.0.1 -P 3306 -u root -p -e "DESCRIBE auth_db.refresh_tokens;"
mysql -h 127.0.0.1 -P 3306 -u root -p -e "DESCRIBE auth_db.otp_challenges;"
mysql -h 127.0.0.1 -P 3306 -u root -p -e "DESCRIBE auth_db.role_assignments;"
```

Check indexes:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p -e "SHOW INDEX FROM auth_db.refresh_tokens;"
mysql -h 127.0.0.1 -P 3306 -u root -p -e "SHOW INDEX FROM auth_db.role_assignments;"
```

Check foreign keys:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p -e "
SELECT
  TABLE_NAME,
  COLUMN_NAME,
  REFERENCED_TABLE_NAME,
  REFERENCED_COLUMN_NAME
FROM information_schema.KEY_COLUMN_USAGE
WHERE TABLE_SCHEMA = 'auth_db'
  AND REFERENCED_TABLE_NAME IS NOT NULL;
"
```

Expected Task 2 foreign keys:

| Table | Column | References |
|---|---|---|
| `credentials` | `account_id` | `auth_accounts(account_id)` |
| `refresh_tokens` | `account_id` | `auth_accounts(account_id)` |
| `role_assignments` | `account_id` | `auth_accounts(account_id)` |

Note: Task 2 base migration does not add a foreign key on `otp_challenges.account_id`. That hardening is added later by migration `003_add_otp_account_fk.up.sql`.

### J. Rollback Task 2 migration

Rollback file:

```text
backend/services/auth-service/migrations/001_create_auth_tables.down.sql
```

Run:

```bash
cd backend/services/auth-service
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/001_create_auth_tables.down.sql
```

Warning:

- Rollback drops Task 2 tables.
- Dropping tables deletes data.
- Local rollback for practice is okay.
- Production rollback needs backup, approval, and tested recovery plan.

### K. Connection string format

No new env variable is introduced by Task 2.

The current service uses:

```env
AUTH_MYSQL_DSN='auth_user:auth_password@tcp(127.0.0.1:3306)/auth_db?parseTime=true&charset=utf8mb4&loc=UTC'
```

This is already documented in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 5. Database Setup
Subsection: I. Connection string format
Section: 7. Environment Variables
```

Important:

- `parseTime=true` is required by Go when scanning MySQL `TIMESTAMP` into `time.Time`.
- `charset=utf8mb4` aligns with Task 2 schema charset.
- `loc=UTC` helps keep timestamp behavior predictable.

---

## 6. Redis / Queue / External Services

Task 2 itself does not introduce Redis, Kafka, RabbitMQ, NATS, MinIO, Elasticsearch, SMTP, Stripe, Twilio, OAuth, or Kubernetes.

| Service | Task 2 status | Notes |
|---|---|---|
| Redis | Not introduced by Task 2 | Current full Auth Service uses Redis for OTP rate limits. |
| Notification Service | Not introduced by Task 2 | Used by later OTP flow implementation. |
| Kafka/RabbitMQ | Not direct dependency | Current code has HTTP outbox publisher behavior, not direct Kafka client. |
| Docker | Optional for MySQL | Same setup as previous dependency file. |

Refer for existing full-service setup:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 6. Redis / Queue / External Services
Section: 8. Docker Setup
```

Beginner note:

Task 2 migration ko apply karne ke liye Redis ya queue start karne ki zaroorat nahi hai. Lekin `go run ./cmd/server` karoge to current code Redis ping bhi karta hai, so full server run ke liye Redis setup follow karo.

---

## 7. Environment Variables

Task 2 adds no new environment variables.

Already documented:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 7. Environment Variables
```

Task 2 only depends on MySQL connection information when running the app:

```env
AUTH_MYSQL_DSN='auth_user:auth_password@tcp(127.0.0.1:3306)/auth_db?parseTime=true&charset=utf8mb4&loc=UTC'
```

Do not duplicate the full `.env` here because it already exists in the previous dependency guide.

Where credentials go:

| Credential | Local placement | Production placement |
|---|---|---|
| MySQL username | `backend/services/auth-service/.env` inside `AUTH_MYSQL_DSN` | Secret manager / Kubernetes Secret |
| MySQL password | `backend/services/auth-service/.env` inside `AUTH_MYSQL_DSN` | Secret manager / Kubernetes Secret |
| DB host/port | `AUTH_MYSQL_DSN` | Environment variable from deployment config |
| DB name | `AUTH_MYSQL_DSN` and migration SQL | Environment variable / migration config |

Security note:

- Current workspace has `backend/services/auth-service/.env`.
- Treat it as local-only.
- Do not commit `.env`.
- If credentials have been shared accidentally, rotate them.

---

## 8. Docker Setup

No new Dockerfile or docker-compose file is introduced by Task 2.

Existing repo status already documented:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 8. Docker Setup
```

Task 2 Docker requirement:

| Docker item | Needed? | Explanation |
|---|---:|---|
| MySQL container | Optional but recommended | Easiest local DB setup |
| Redis container | Not for Task 2 migration | Required for full current server startup |
| App container | Available | `backend/services/auth-service/Dockerfile` and root Compose service are present |
| Persistent MySQL volume | Recommended | Keeps schema/data after container restart |
| Docker network | Optional | Useful if app also runs inside Compose later |

If MySQL runs in Docker and Go app runs on host:

```env
AUTH_MYSQL_DSN='auth_user:auth_password@tcp(127.0.0.1:3306)/auth_db?parseTime=true&charset=utf8mb4&loc=UTC'
```

If app later runs inside Docker Compose with MySQL service named `mysql`:

```env
AUTH_MYSQL_DSN='auth_user:auth_password@tcp(mysql:3306)/auth_db?parseTime=true&charset=utf8mb4&loc=UTC'
```

Do not add a new compose example here. Reuse the previous dependency guide.

### Ports & networking

Task 2 introduces no new application port. It only needs MySQL access for schema migration.

| Service | Port | Purpose | Task 2 status |
|---|---:|---|---|
| MySQL | `3306` | Apply and verify `auth_db` schema | Reused from previous setup |
| Auth Service HTTP | `8081` | Full service API runtime | Not needed for migration-only Task 2 |
| Redis | `6379` | OTP rate limit cache in current full service | Not needed for migration-only Task 2 |
| Notification Service | `8084` | OTP delivery integration in current full service | Not needed for migration-only Task 2 |

Networking rules:

- If MySQL runs in Docker and commands run on host, use `127.0.0.1:3306`.
- If the app later runs inside Docker Compose, use the MySQL service name, for example `mysql:3306`.
- If port `3306` is already busy, map MySQL to another host port such as `3307` and update `AUTH_MYSQL_DSN`.

Full networking troubleshooting is already documented in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 11. Ports & Networking
```

---

## 9. Local Development Setup

This section shows the Task 2 schema-only onboarding flow. For the full Auth Service run flow, use `task1_Dependency.md`.

### Step 1: Clone repository

Already documented:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 9. Local Development Setup
Subsection: Step 1: Clone repository
```

### Step 2: Start MySQL

Use existing MySQL setup from:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 5. Database Setup
Section: 8. Docker Setup
```

Verify MySQL:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p -e "SELECT VERSION();"
```

### Step 3: Apply Task 2 base migration

```bash
cd backend/services/auth-service
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/001_create_auth_tables.up.sql
```

### Step 4: Verify schema

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p -e "SHOW TABLES FROM auth_db;"
```

Expected Task 2 tables:

```text
auth_accounts
credentials
otp_challenges
refresh_tokens
role_assignments
```

### Step 5: Decide what you want to run next

| Goal | Next step |
|---|---|
| Only verify Task 2 schema | Stop after table/index verification |
| Run current full Auth Service | Apply migrations `002` to `005`, then follow `task1_Dependency.md` |
| Test later features like JWT/OTP/RBAC/outbox | Use full setup from previous dependency guide |

Important:

Current code has progressed beyond Task 2. If you only apply `001_create_auth_tables.up.sql` and then start the latest server, later features may fail because expected later columns/tables are missing.

---

## 10. Running the Project

Task 2 does not add a new process to run. It adds database schema.

To run only Task 2 validation:

```bash
cd backend/services/auth-service
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/001_create_auth_tables.up.sql
mysql -h 127.0.0.1 -P 3306 -u root -p -e "SHOW TABLES FROM auth_db;"
```

To run the current Auth Service server:

Refer:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 9. Local Development Setup
Section: 10. Running the Project
```

Reason:

- Full server startup needs MySQL + Redis + JWT keys + `.env`.
- Full server expects all current migrations, not only Task 2 base schema.

---

## 11. Common Errors & Fixes

Only Task 2 specific errors are listed here. General Go, Redis, Docker, and full service errors already exist in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 12. Common Errors & Fixes
```

| Error | Likely cause | Fix |
|---|---|---|
| `ERROR 1049 (42000): Unknown database 'auth_db'` | Down migration ran, or DB was not created | Re-run `001_create_auth_tables.up.sql` using a user with `CREATE DATABASE` permission |
| `ERROR 1044` or `Access denied for user` | Migration user lacks DDL permissions | Run migration as root locally, or grant migration user `CREATE`, `ALTER`, `DROP`, `INDEX` on `auth_db` |
| `ERROR 1050: Table already exists` | Migration partially ran before | Inspect existing tables with `SHOW TABLES FROM auth_db;`; for local only, rollback or recreate DB |
| `ERROR 1215` / `Cannot add foreign key constraint` | Parent table missing, wrong table order, or charset/collation mismatch | Ensure `auth_accounts` exists first and all tables use `utf8mb4` |
| `ERROR 1071: Specified key was too long` | Old MySQL version or incompatible charset/index limit | Use MySQL 8+ as required |
| No tables shown after migration | Ran command against wrong MySQL instance/port | Check `mysql -h 127.0.0.1 -P 3306` and Docker port mapping |
| `parseTime` related app error later | DSN missing `parseTime=true` | Use documented `AUTH_MYSQL_DSN` format |
| Full server fails after only Task 2 migration | Current code expects later migrations | Apply migrations `002` to `005` as documented in previous dependency guide |
| Duplicate active global role allowed | MySQL unique indexes treat `NULL` values as distinct | Apply later migration `004_harden_role_assignments.up.sql`; Task 2 doc already warns about this |

Beginner debugging commands:

```bash
docker ps
docker logs ecommerce-auth-mysql
mysql -h 127.0.0.1 -P 3306 -u root -p -e "SHOW DATABASES;"
mysql -h 127.0.0.1 -P 3306 -u root -p -e "SHOW TABLES FROM auth_db;"
```

---

## 12. Security & Best Practices

Do not repeat the full security guide. Refer:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 13. Security & Best Practices
```

Task 2 specific security practices:

- Never store plain passwords in `credentials.password_hash`.
- Never store plain refresh tokens in `refresh_tokens.token_hash`.
- Never store plain OTP codes in `otp_challenges.otp_hash`.
- Use app-level hashing/HMAC before inserting token or OTP hashes.
- Keep `auth_db` owned by Auth Service only.
- Other services should not directly read/write Auth DB tables.
- Use MySQL least privilege: app user should not need `DROP` in production.
- Take backup before running destructive migrations.
- Keep rollback scripts protected because they drop tables.
- Use UTC timestamps consistently.

Schema-specific audit:

| Finding | Risk | Recommendation |
|---|---|---|
| `001_create_auth_tables.down.sql` drops all Task 2 tables | Data loss if run accidentally | Restrict production rollback permissions and require backup |
| `otp_challenges.account_id` has no FK in Task 2 base schema | Optional account link can point to missing account | Apply later migration `003_add_otp_account_fk.up.sql` in full system |
| `role_assignments.uk_role_active` includes nullable columns | MySQL allows multiple `NULL` values in unique indexes | Apply later migration `004_harden_role_assignments.up.sql` |
| Compose `migrate-auth` job | Applies ordered migrations before Auth starts | Keep migration image/version and ordering validated |
| Repository `.gitignore` and sanitized `.env.example` are present | Local secrets stay outside source control | Never replace placeholders with production values |
| Dockerfile and root Compose are present | Local setup is repeatable | Keep image build and Compose health checks in CI |

---

## 13. Missing or Misconfigured Things

Task 2 specific missing/improvement items:

| Item | Why it matters | Suggested fix |
|---|---|---|
| Migration runner | Compose `migrate-auth` runs `golang-migrate` | Keep migration execution in CI and startup dependency checks |
| Migration tracking table | `golang-migrate` records applied versions | Monitor dirty migration state and test rollback procedures |
| Dedicated migration user | Root migration is okay locally but not ideal in prod | Create separate migration DB user with DDL permissions |
| `.env.example` | Sanitized template is present | Keep placeholder values only |
| `.gitignore` | Env, secrets, PEM, and binaries are ignored | Keep secret scanning enabled |
| Readiness check | `/readyz` verifies MySQL and Redis connectivity | Keep migration completion as a Compose/Kubernetes rollout prerequisite |

Already documented broader missing items:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 14. Missing or Misconfigured Things
```

---

## 14. References to Previous Dependency Files

Previous dependency files found in the same service folder:

| File | Reused for |
|---|---|
| `TaskImplementation/Auth Service/task1_Dependency.md` | Go setup, full dependency installation, MySQL install, Redis setup, Docker commands, `.env`, full local run, ports, common errors, security checklist |

Exact sections reused:

| Topic | Refer |
|---|---|
| Required software install | `task1_Dependency.md` section `3. Required Software` |
| Go modules and commands | `task1_Dependency.md` section `4. Dependency Management` |
| MySQL installation and Docker | `task1_Dependency.md` section `5. Database Setup` |
| Redis and external services | `task1_Dependency.md` section `6. Redis / Queue / External Services` |
| Full `.env` | `task1_Dependency.md` section `7. Environment Variables` |
| Docker Compose example | `task1_Dependency.md` section `8. Docker Setup` |
| Full local app run | `task1_Dependency.md` sections `9. Local Development Setup` and `10. Running the Project` |
| Ports and networking | `task1_Dependency.md` section `11. Ports & Networking` |
| Common runtime errors | `task1_Dependency.md` section `12. Common Errors & Fixes` |
| Security baseline | `task1_Dependency.md` section `13. Security & Best Practices` |

No `task2_Dependency.md`, `task3_Dependency.md`, etc. existed before this file was generated.

---

## 15. Final Checklist

Task 2 schema-only checklist:

- [ ] MySQL 8+ is installed or running in Docker.
- [ ] MySQL CLI works with `mysql --version`.
- [ ] You can connect to local MySQL on expected host/port.
- [ ] `backend/services/auth-service/migrations/001_create_auth_tables.up.sql` exists.
- [ ] `001_create_auth_tables.up.sql` was applied successfully.
- [ ] `auth_db` database exists.
- [ ] `auth_accounts` table exists.
- [ ] `credentials` table exists.
- [ ] `refresh_tokens` table exists.
- [ ] `otp_challenges` table exists.
- [ ] `role_assignments` table exists.
- [ ] Foreign keys exist for `credentials`, `refresh_tokens`, and `role_assignments`.
- [ ] `AUTH_MYSQL_DSN` points to `auth_db` if you plan to run the app.
- [ ] `AUTH_MYSQL_DSN` includes `parseTime=true&charset=utf8mb4&loc=UTC`.
- [ ] You understand that current full server needs migrations `001` to `005`.
- [ ] You did not commit `.env` or private secrets.

Quick Task 2 verification:

```bash
cd backend/services/auth-service
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/001_create_auth_tables.up.sql
mysql -h 127.0.0.1 -P 3306 -u root -p -e "SHOW TABLES FROM auth_db;"
```

If the expected five Task 2 tables are visible, Task 2 database setup is ready.
