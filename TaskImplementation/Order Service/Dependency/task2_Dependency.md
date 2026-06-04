# Dependency and Setup Guide for `${SERVICE_NAME}`

## Variables used in this document

```text
SERVICE_NAME=Order Service
TASK_FILE_NAME=task2.md
INPUT_FILE_PATH=TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}
OUTPUT_FILE_NAME=task2_Dependency.md
OUTPUT_FILE_PATH=TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}
SERVICE_CODE_PATH=backend/services/order-service
PREVIOUS_DEPENDENCY_FILE=TaskImplementation/${SERVICE_NAME}/task1_Dependency.md
TASK_2_MIGRATION_UP=${SERVICE_CODE_PATH}/migrations/001_create_order_tables.up.sql
TASK_2_MIGRATION_DOWN=${SERVICE_CODE_PATH}/migrations/001_create_order_tables.down.sql
```

Use `${INPUT_FILE_PATH}` as the source task guide and `${OUTPUT_FILE_PATH}` as this generated dependency/setup guide.

## Previous Dependency Reuse

Before creating this file, previous dependency documentation inside the same `${SERVICE_NAME}` folder was checked.

| Previous file | Reused topics |
|---|---|
| `${PREVIOUS_DEPENDENCY_FILE}` | Go setup, Go modules, MySQL installation, Docker MySQL, Kafka setup, common ports, environment loading, common errors, security notes, and general run flow |

Important: this file does not repeat the full setup already explained in `${PREVIOUS_DEPENDENCY_FILE}`. It explains only what is new or task-specific for `${TASK_FILE_NAME}`.

## What Task 2 Adds

`${TASK_FILE_NAME}` is the MySQL schema task for `${SERVICE_NAME}`. It defines the durable database foundation for orders.

New task-specific output:

| Item | Status | Notes |
|---|---:|---|
| MySQL schema design | New for Task 2 | Base relational model for order data |
| `order_db` database | New for Task 2 | Owned by `${SERVICE_NAME}` |
| `orders` table | New for Task 2 | Main order header |
| `order_items` table | New for Task 2 | Purchased item snapshots |
| `order_status_history` table | New for Task 2 | Audit trail for lifecycle changes |
| `shipments` table | New for Task 2 | Seller/package fulfillment tracking |
| `order_idempotency_keys` table | New for Task 2 | Duplicate checkout prevention foundation |
| Migration files | Implemented in codebase | `${TASK_2_MIGRATION_UP}` and `${TASK_2_MIGRATION_DOWN}` |

Simple Hinglish: Task 2 ka main kaam database tables banana hai. Code business flow yaha nahi aata; yaha sirf MySQL schema, constraints, indexes, aur migration setup important hai.

## 1. Project Tech Stack Analysis

### Reused technologies

The following technologies are already explained in `${PREVIOUS_DEPENDENCY_FILE}`:

| Technology | Reuse note |
|---|---|
| Go | Same backend module; no new Go runtime dependency introduced by Task 2 |
| Go modules | Same `go.mod` and `go.sum` setup |
| Markdown | Same task documentation format |
| Mermaid | Same diagram support for task docs |
| Docker | Same local dependency container approach |
| Kafka | Not part of Task 2; already documented for later/current runtime event setup |

Refer:

```text
${PREVIOUS_DEPENDENCY_FILE}
Sections:
- Project Tech Stack Analysis
- Go Dependency System
- Docker and DevOps Setup
```

### Task-2-specific technologies

| Technology | Required? | Why used in Task 2 | Beginner explanation |
|---|---:|---|---|
| MySQL 8+ | Yes for applying schema | Stores order tables transactionally | MySQL ek relational database hai. Data rows aur tables me store hota hai, aur order creation jaise critical kaam transaction me safe rehte hain. |
| InnoDB | Yes | Enables transactions and foreign keys | InnoDB MySQL ka storage engine hai jo rollback, commit, aur FK checks support karta hai. |
| SQL migrations | Yes | Schema changes repeatably apply/rollback karne ke liye | Migration ek versioned SQL file hoti hai. `.up.sql` schema apply karta hai, `.down.sql` rollback karta hai. |
| MySQL `JSON` columns | Yes | Stores address and audit metadata snapshots | JSON column flexible structured data ke liye use hota hai, jaise delivery address snapshot. |
| MySQL `ENUM` columns | Yes | Restricts status values | ENUM allowed values ki list hoti hai. Galat status insert hone se bachata hai. |
| MySQL `CHECK` constraints | Yes | Prevents invalid amount/quantity data | CHECK rule DB level par validation karta hai, jaise amount negative nahi hona chahiye. |

## 2. Language-Specific Dependency System

No new Go library is introduced specifically by `${TASK_FILE_NAME}`.

The backend module dependency system is unchanged:

```text
${SERVICE_CODE_PATH}/go.mod
${SERVICE_CODE_PATH}/go.sum
backend/go.work
backend/shared/gen/go/go.mod
```

Refer:

```text
${PREVIOUS_DEPENDENCY_FILE}
Section:
Go Dependency System
```

Task 2 is schema-focused, so the most important command is not `go get`; it is applying the SQL migration after MySQL is running.

For package verification, use the same commands from the previous dependency guide:

```bash
cd backend/services/order-service
go test ./...
go build ./...
```

## 3. Database Analysis

### Database used: MySQL 8+

Full MySQL installation and Docker setup are already documented in:

```text
${PREVIOUS_DEPENDENCY_FILE}
Sections:
- Database Analysis
- Docker and DevOps Setup
```

Only Task-2-specific schema details are documented below.

### A. What it is

MySQL relational database hai. Isme data tables me store hota hai, aur tables ke beech relation foreign keys se maintain ho sakta hai.

### B. Why `${SERVICE_NAME}` uses it

Order data transactional hota hai. Ek order create karte time header, items, status history, shipment foundation, aur idempotency row consistent rehni chahiye.

MySQL is used because:

- order and order items relational data hain
- checkout duplicate nahi hona chahiye
- status history audit trail chahiye
- money values accurate store karne hain
- indexes se buyer/seller/admin queries fast chalni chahiye

### C. Required or optional

| Use case | MySQL required? |
|---|---:|
| Reading `${INPUT_FILE_PATH}` only | No |
| Applying Task 2 migration | Yes |
| Running repository code with real DB | Yes |
| Running unit tests that use mocks only | No |

### D. Local installation

Installation is reused from `${PREVIOUS_DEPENDENCY_FILE}`.

Please follow:

```text
${PREVIOUS_DEPENDENCY_FILE}
Section:
Database Analysis -> Database used: MySQL 8+ -> Local installation
```

### E. Docker setup

Docker MySQL setup is reused from `${PREVIOUS_DEPENDENCY_FILE}`.

Please follow:

```text
${PREVIOUS_DEPENDENCY_FILE}
Section:
Database Analysis -> Database used: MySQL 8+ -> Docker setup
```

Task 2 does not need a new database container, new volume, or new Docker network.

### F. Task-specific start commands

Start MySQL using the same approach from the previous dependency guide. Then from repository root apply only the Task 2 base schema:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 < backend/services/order-service/migrations/001_create_order_tables.up.sql
```

Why no database name is passed in this command?

`${TASK_2_MIGRATION_UP}` already contains:

```sql
CREATE DATABASE IF NOT EXISTS order_db;
USE order_db;
```

So it can create/select `order_db` itself.

### G. Verify running and schema applied

After applying the migration:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW TABLES;"
```

Expected Task 2 tables:

```text
order_idempotency_keys
order_items
order_status_history
orders
shipments
```

Verify indexes:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW INDEX FROM orders;"
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW INDEX FROM order_items;"
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW INDEX FROM order_status_history;"
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW INDEX FROM shipments;"
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW INDEX FROM order_idempotency_keys;"
```

Verify foreign keys:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SELECT table_name, constraint_name, referenced_table_name FROM information_schema.key_column_usage WHERE table_schema = 'order_db' AND referenced_table_name IS NOT NULL;"
```

### H. Default port

MySQL default port is reused:

```text
3306
```

Refer:

```text
${PREVIOUS_DEPENDENCY_FILE}
Section:
Ports and Networking
```

### I. Connection string format

The app-level DSN format is already documented in `${PREVIOUS_DEPENDENCY_FILE}`.

For Task 2 schema verification, this DSN should point to `order_db`:

```env
ORDER_MYSQL_DSN=order_user:order_password@tcp(localhost:3306)/order_db?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci
```

### J. Where to place credentials

For local runtime, credentials are reused from:

```text
${PREVIOUS_DEPENDENCY_FILE}
Sections:
- Database Analysis -> Where to place credentials
- Environment Variables
```

Task 2 itself does not add a new credentials file. For manual migration commands, credentials are typed into the MySQL CLI or provided through your local MySQL client configuration.

## 4. Task 2 Schema Requirements

### Migration files

| File | Purpose |
|---|---|
| `${TASK_2_MIGRATION_UP}` | Creates `order_db` and Task 2 base tables |
| `${TASK_2_MIGRATION_DOWN}` | Drops Task 2 tables in foreign-key-safe order |

### Tables created by Task 2

| Table | Purpose | Required? |
|---|---|---:|
| `orders` | Main order header, buyer, cart reference, status, totals, address snapshot | Yes |
| `order_items` | Item-level purchased product snapshots and seller references | Yes |
| `order_status_history` | Append-only audit history for status transitions | Yes |
| `shipments` | Shipment/tracking foundation for split fulfillment | Yes |
| `order_idempotency_keys` | Prevents duplicate checkout requests by user/key | Yes |

### Actual migration note

`${INPUT_FILE_PATH}` explains the intended base schema. The checked-in `${TASK_2_MIGRATION_UP}` implements the same base schema and also includes these audit-friendly columns in `order_status_history`:

| Column | Why it exists |
|---|---|
| `actor_type` | Tells who/what changed the status: system, buyer, seller, admin, payment service, or logistics |
| `metadata` | Stores optional JSON context for debugging/audit |

Beginner note: Agar task markdown aur actual SQL compare kar rahe ho, ye difference expected implementation detail hai. Setup ke liye bas actual migration file ko source of truth mano.

### MySQL features required by the schema

| Feature | Used where | Setup impact |
|---|---|---|
| `CREATE DATABASE IF NOT EXISTS` | `order_db` bootstrap | Migration can create DB automatically |
| `utf8mb4` charset | Database/table text | Supports full Unicode text safely |
| `TIMESTAMP` columns | Audit fields | Server timezone should be consistent; migration sets UTC session timezone |
| `JSON` columns | `address_snapshot`, `metadata`, later outbox payloads | Use MySQL 8+ |
| `ENUM` columns | Status fields | App and DB status values must stay aligned |
| `CHECK` constraints | Amounts, quantity, inventory pair later | MySQL 8+ enforces checks |
| Foreign keys | Child tables reference `orders(order_id)` | Tables must use InnoDB |
| Indexes | Buyer/seller/status/idempotency queries | No extra setup, but verify after migration |

## 5. Environment Variables

No new environment variables are introduced by `${TASK_FILE_NAME}`.

Reused runtime variable:

| Variable | Required for app runtime? | Required for manual Task 2 migration? | Notes |
|---|---:|---:|---|
| `ORDER_MYSQL_DSN` | Yes | No | Used by Go code to connect to MySQL; CLI migration can use `mysql -u ...` directly |

Refer:

```text
${PREVIOUS_DEPENDENCY_FILE}
Section:
Environment Variables
```

Task-specific `.env` reminder:

```env
ORDER_MYSQL_DSN=order_user:order_password@tcp(localhost:3306)/order_db?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci
```

Security note: DB password secret hota hai. `.env` commit mat karo.

Common Task 2 mistakes:

| Mistake | Result | Fix |
|---|---|---|
| `.env` created but migration command uses root CLI login | Confusion about which credentials are used | App DSN and MySQL CLI credentials are separate things |
| `ORDER_MYSQL_DSN` points to wrong DB name | App starts but cannot find tables | DSN database must be `order_db` |
| DSN omits `parseTime=true` | Go timestamp scanning may fail later | Keep `parseTime=true` |
| Docker host uses `mysql` but command runs on host | Connection failure | From host use `localhost`; inside compose use `mysql` |

## 6. External Services Analysis

### MySQL

Task 2 requires MySQL when applying or testing the schema against a real database.

Detailed installation and Docker setup are reused from:

```text
${PREVIOUS_DEPENDENCY_FILE}
Sections:
- Database Analysis
- External Services Analysis -> MySQL
```

### Kafka

Kafka is not introduced by `${TASK_FILE_NAME}`.

It is documented for later/current runtime event publishing in:

```text
${PREVIOUS_DEPENDENCY_FILE}
Section:
External Services Analysis -> Kafka
```

Task 2 creates no Kafka topic and requires no Kafka broker.

### Redis

Redis is not used by Task 2.

### RabbitMQ/NATS

RabbitMQ and NATS are not used by Task 2.

### Payment provider

No payment provider SDK or credential is introduced by Task 2. `payment_id` is only a nullable reference field in the base `orders` table.

## 7. Ports and Networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| MySQL | 3306 | Apply and verify Task 2 schema | Reused |
| gRPC API | 9094 | Runtime API port documented previously | Reused, not needed for Task 2 migration |
| Kafka | 9092 | Event broker for later/current runtime events | Reused, not needed for Task 2 |
| Redis | 6379 | Not used by this task | Not used |

Task 2 introduces no new port.

For port conflicts, reuse:

```text
${PREVIOUS_DEPENDENCY_FILE}
Section:
Ports and Networking
```

Quick MySQL port check:

```bash
ss -ltnp | grep ':3306'
```

## 8. Docker and DevOps Setup

Task 2 does not introduce:

- a Dockerfile
- a new compose file
- a new container
- a new Docker network
- a new Docker volume
- a new health check

Use the same MySQL Docker setup from:

```text
${PREVIOUS_DEPENDENCY_FILE}
Section:
Docker and DevOps Setup
```

After MySQL is running, apply `${TASK_2_MIGRATION_UP}`.

Suggested local flow:

```bash
docker start order-mysql
mysql -u root -p -h 127.0.0.1 -P 3306 < backend/services/order-service/migrations/001_create_order_tables.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW TABLES;"
```

If the MySQL container does not exist yet, create it using the Docker command already documented in `${PREVIOUS_DEPENDENCY_FILE}`.

## 9. Project Run Instructions

This is the beginner-friendly onboarding flow for Task 2 without repeating the full previous guide.

### Step 1: Clone repository

Use the clone instructions from:

```text
${PREVIOUS_DEPENDENCY_FILE}
Section:
Project Run Instructions -> Step 1
```

### Step 2: Read the Task 2 guide

```bash
cd Ecommerce
sed -n '1,220p' "TaskImplementation/Order Service/task2.md"
```

### Step 3: Start MySQL

Use installed MySQL or Docker MySQL from `${PREVIOUS_DEPENDENCY_FILE}`.

Verify MySQL accepts connections:

```bash
mysqladmin ping -h 127.0.0.1 -P 3306 -u root -p
```

### Step 4: Apply only the Task 2 base migration

From repository root:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 < backend/services/order-service/migrations/001_create_order_tables.up.sql
```

### Step 5: Verify tables

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW TABLES;"
```

Expected:

```text
orders
order_items
order_status_history
shipments
order_idempotency_keys
```

### Step 6: Verify important columns

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "DESCRIBE orders;"
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "DESCRIBE order_items;"
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "DESCRIBE order_status_history;"
```

### Step 7: Optional Go verification

Task 2 itself is schema documentation, but the current module can still be checked:

```bash
cd backend/services/order-service
go test ./...
```

### Step 8: Rollback if needed

Only rollback on a local/dev database where data loss is acceptable:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/001_create_order_tables.down.sql
```

Warning: down migration drops Task 2 tables and their data.

## 10. Migration Order

For Task 2 only:

```text
001_create_order_tables.up.sql
```

For the current full `${SERVICE_CODE_PATH}` runtime, later migrations also exist:

| Migration | Related task | Purpose | Task 2 dependency file action |
|---|---|---|---|
| `001_create_order_tables.up.sql` | Task 2 | Base schema | Explained here |
| `002_add_payment_coordination.up.sql` | Task 4 | Adds payment failure/inventory reservation fields | Reuse previous runtime guide or future task dependency docs |
| `003_add_grpc_query_indexes.up.sql` | Task 5 | Adds buyer listing index | Not part of Task 2 |
| `004_create_order_outbox_events.up.sql` | Task 8 | Adds outbox events table | Not part of Task 2 |

Beginner rule: Agar sirf Task 2 verify karna hai, migration `001` enough hai. Agar current backend runtime ke all tests/integration setup ko real DB ke saath chalana hai, migrations `001` to `004` order me apply karni pad sakti hain.

## 11. Common Errors and Fixes

Common Go, Docker, Kafka, and general env errors are already listed in:

```text
${PREVIOUS_DEPENDENCY_FILE}
Section:
Common Errors and Fixes
```

Task-2-specific errors:

| Error | Cause | Fix |
|---|---|---|
| `Unknown database 'order_db'` | You ran a command against `order_db` before applying `001` | Run `${TASK_2_MIGRATION_UP}` first |
| `Table 'orders' already exists` | Schema was already applied manually or partially | Use `SHOW TABLES;`; migration uses `IF NOT EXISTS`, but manual changes may conflict |
| Foreign key error while dropping tables | Parent table dropped before child tables | Use `${TASK_2_MIGRATION_DOWN}` because it drops child tables first |
| `CHECK constraint ... is violated` | Negative amount or zero quantity inserted | Use positive quantity and non-negative amount values |
| Invalid `ENUM` value | Status value not in schema enum | Use only documented lifecycle status values |
| JSON insert error | Invalid JSON for `address_snapshot` or `metadata` | Validate JSON syntax before insert |
| `Access denied for user` | Wrong MySQL username/password | Check Docker env vars and CLI login |
| `connection refused 127.0.0.1:3306` | MySQL is not running or port changed | Start MySQL or use correct port |
| App cannot find tables | `ORDER_MYSQL_DSN` points to different database | Set DSN database name to `order_db` |

## 12. Security and Configuration Notes

Reused security guidance:

```text
${PREVIOUS_DEPENDENCY_FILE}
Sections:
- Security and Configuration Audit
- Best Practices
```

Task-2-specific security notes:

- Use a dedicated MySQL app user for runtime, not root.
- Keep root password local only.
- Do not commit `.env`.
- Do not store real customer addresses in sample docs.
- Treat `address_snapshot` as personal data.
- Restrict direct DB access in production; services should access order data through service APIs.
- Back up database before running down migrations in shared environments.

## 13. Final Checklist

- [ ] `${INPUT_FILE_PATH}` reviewed
- [ ] `${PREVIOUS_DEPENDENCY_FILE}` checked for reused setup
- [ ] MySQL 8+ installed or Docker MySQL running
- [ ] MySQL port `3306` available or alternate port understood
- [ ] `${TASK_2_MIGRATION_UP}` exists
- [ ] `${TASK_2_MIGRATION_DOWN}` exists
- [ ] `001_create_order_tables.up.sql` applied
- [ ] `order_db` created
- [ ] `orders` table exists
- [ ] `order_items` table exists
- [ ] `order_status_history` table exists
- [ ] `shipments` table exists
- [ ] `order_idempotency_keys` table exists
- [ ] Indexes verified with `SHOW INDEX`
- [ ] Foreign keys verified with `information_schema`
- [ ] Local `ORDER_MYSQL_DSN` points to `order_db` if running app code
- [ ] No new Kafka/Redis/RabbitMQ setup assumed for Task 2
- [ ] Rollback command understood before using it

## Quick Beginner Summary

For `${TASK_FILE_NAME}`, the new dependency is not a new Go package. The important dependency is MySQL 8+ plus the base schema migration.

Follow `${PREVIOUS_DEPENDENCY_FILE}` for installing MySQL/Docker. Then run:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 < backend/services/order-service/migrations/001_create_order_tables.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW TABLES;"
```

Simple Hinglish rule: pehle MySQL chalao, phir Task 2 migration apply karo, phir tables verify karo. Kafka, Redis, RabbitMQ, aur new env vars Task 2 ke liye required nahi hain.
