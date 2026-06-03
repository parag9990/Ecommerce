# Project Dependency & Setup Guide

## 1. Project Overview

This guide is for the implementation task referenced by these variables:

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `User Service` |
| `TASK_FILE_NAME` | `task4.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task4_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

Task 4 ka goal internal gRPC transport layer setup samajhna hai. Browser direct is service ko call nahi karega. API Gateway, Auth Service, Order/Product/CMS type internal services gRPC ke through user profile data lenge.

Task 4 ke core methods:

| gRPC Method | Purpose | Runtime dependency |
|---|---|---|
| `CreateUser` | Auth signup ke baad profile create karta hai | MySQL `users` table, caller/audit metadata |
| `GetUser` | User profile read karta hai | MySQL `users` table |
| `UpdateUserProfile` | Profile patch update karta hai | MySQL `users` table, `FieldMask`, caller/audit metadata |
| `GetSellerProfile` | Seller profile read karta hai | MySQL `seller_profiles` table |

Previous dependency files already cover most base setup. Is file ka focus duplicate setup repeat karna nahi hai. Yahan mainly Task 4 specific gRPC/proto/runtime notes diye gaye hain.

Analyzed files:

| File | Why checked |
|---|---|
| `INPUT_FILE_PATH` | Task 4 scope and intended methods |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Base Go, MySQL, Docker, `.env`, gRPC, grpcurl setup |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | MySQL schema and migration setup |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | Repository setup and DB access requirements |
| `proto/ecommerce/user/v1/user.proto` | Actual gRPC contract |
| `proto/buf.yaml`, `proto/buf.gen.yaml` | Proto lint/generation config |
| `backend/shared/gen/go/ecommerce/user/v1/` | Generated Go proto/gRPC code |
| `backend/services/user-service/internal/transport/grpc/` | gRPC server, mapper, metadata, interceptors, error mapping |
| `backend/services/user-service/internal/config/config.go` | Runtime environment variables |
| `backend/services/user-service/cmd/server/main.go` | Service startup, DB wiring, gRPC listener |

Important current-repo note:

- Task 4 itself is gRPC transport setup.
- Current repository has later User Service tasks also implemented, including validation, audit fields, and outbox events.
- Because of that, running the present service code may require later migrations/settings even though those are not new Task 4 dependencies. This guide calls those spots out clearly.

## 2. Tech Stack

| Technology | Required? | New for Task 4? | Beginner explanation |
|---|---:|---:|---|
| Go `1.24` | Yes | Reused | Go backend language hai. Service, repository, gRPC handler sab Go me likhe gaye hain. |
| Go Modules | Yes | Reused | `go.mod` dependencies list karta hai, `go.sum` checksums lock karta hai. Setup already explained in `task1_Dependency.md`. |
| `backend/go.work` | Local dev | Reused | Workspace multiple Go modules ko connect karta hai: user-service, api-gateway, shared generated code, shared validation. |
| MySQL 8.x | Yes for real runtime | Reused | Profile and seller data relational tables me store hota hai. Full setup `task1_Dependency.md` and schema `task2_Dependency.md` me hai. |
| gRPC (`google.golang.org/grpc`) | Yes | Active Task 4 surface | gRPC high-performance internal service communication framework hai. Is task me User Service methods expose hote hain. |
| Protocol Buffers | Yes | Active Task 4 surface | Proto file ek contract hai. Request/response shape yahi define karta hai. |
| Buf CLI | Needed when regenerating proto | Reused | `buf generate` proto se Go files banata hai. Install/setup already `task1_Dependency.md` me explained hai. |
| `google.protobuf.FieldMask` | Yes for update calls | Task 4 specific | Partial update me kaunse fields change karne hain, ye batata hai. Empty string accidental overwrite se bachata hai. |
| `google.protobuf.Timestamp` | Yes | Task 4 specific mapping | Go `time.Time` ko proto timestamp me convert karta hai. |
| gRPC metadata | Yes for caller context | Task 4 specific runtime behavior | Headers jaise `x-user-id`, `x-service-name`, `x-actor-id` service ko caller/audit context dete hain. |
| gRPC reflection | Optional local dev | Reused | `grpcurl list/describe` ko easy banata hai. Production me disable/restrict karna chahiye. |
| `slog` structured logging | Yes | Reused/current runtime | gRPC interceptors request method, code, duration, request id log karte hain. |

Not required for Task 4:

| Service/tool | Task 4 status | Note |
|---|---|---|
| Redis | Not required | Current Task 4 gRPC path me Redis client/env nahi hai. |
| Kafka | Not required for Task 4 | Current repo has later event code, but Task 4 does not need Kafka. |
| RabbitMQ | Not required for Task 4 | Only needed if later outbox worker is enabled. |
| MongoDB | Not used | User profile data MySQL me hai. |
| Kubernetes | Not required locally | Production readiness topic hai, local Task 4 setup nahi. |

## 3. Required Software

Do not reinstall base tools if you already followed previous dependency guides.

| Software | Required? | Follow this setup |
|---|---:|---|
| Go `1.24+` | Yes | `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, section `2. Language-Specific Dependency System: Go` |
| MySQL 8.x | Yes for running service | `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, section `3. Database Analysis` |
| MySQL CLI | Recommended | `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`, section `5. Database Setup` |
| Docker | Optional | Reuse Docker MySQL setup from `task1_Dependency.md` |
| Buf CLI | Required only when changing `.proto` | Reuse `task1_Dependency.md`, section `5. External Services Analysis -> Protobuf and Buf` |
| grpcurl | Optional but useful | Reuse `task1_Dependency.md`, section `5. External Services Analysis -> gRPC` |

Task 4 does not require a new Docker container, Redis, Kafka, RabbitMQ, object storage, or external SaaS credentials.

## 4. Dependency Management

### Reused Go Dependency Setup

Go module basics, `go mod download`, `go mod tidy`, `go.sum`, workspace notes, and common module errors are already documented in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: 2. Language-Specific Dependency System: Go
```

Task 4 does not require beginners to manually edit `go.mod`.

Important direct dependencies already present in `backend/services/user-service/go.mod`:

| Dependency | Why Task 4 needs it |
|---|---|
| `google.golang.org/grpc` | gRPC server, status codes, interceptors, reflection support |
| `google.golang.org/protobuf` | Generated protobuf messages, `FieldMask`, timestamps |
| `github.com/parag/ecommerce/backend/shared/gen/go` | Generated `UserServiceServer` interface and request/response types |
| `github.com/go-sql-driver/mysql` | Runtime service still reaches MySQL through repository/usecase layers |
| `github.com/DATA-DOG/go-sqlmock` | Repository tests, not required for gRPC handler tests |

### Proto Generation Workflow

Full Buf install explanation already exists in `task1_Dependency.md`. Task 4 specific workflow:

```bash
cd proto
buf lint
buf generate
```

Expected generated files:

```text
backend/shared/gen/go/ecommerce/user/v1/user.pb.go
backend/shared/gen/go/ecommerce/user/v1/user_grpc.pb.go
```

Beginner note:

- `.proto` file edit karo, generated Go file manually edit mat karo.
- Agar generated code old hai, gRPC server compile error de sakta hai.
- `buf.gen.yaml` remote plugins use karta hai. Fresh machine par first generation ke time network access required ho sakta hai.

## 5. Database Setup

### Reused Database Setup

MySQL installation, Docker MySQL command, DSN format, credentials placement, migration apply/rollback basics already documented hain:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: 3. Database Analysis

TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
Section: 5. Database Setup
```

Do not duplicate those steps here. Follow those first.

### What Task 4 Uses In MySQL

Task 4 gRPC handler direct SQL nahi chalata. Flow ye hai:

```text
gRPC request -> transport/grpc handler -> usecase -> repository -> MySQL
```

| Method | Main table needed | Why |
|---|---|---|
| `CreateUser` | `users` | New profile row create hota hai. |
| `GetUser` | `users` | Profile read hota hai. |
| `UpdateUserProfile` | `users` | Profile fields patch update hote hain. |
| `GetSellerProfile` | `seller_profiles` | Seller profile lookup hota hai. |

For the original Task 4 scope, the Task 2 schema is the required database setup.

### Current Repository Caveat

Current codebase has later tasks merged after Task 4. Because of this:

| Current code behavior | Setup impact |
|---|---|
| Repository queries use audit columns like `created_by`, `updated_by`, `deleted_at`, `status_changed_by` | Apply migration `002_add_user_audit_fields.up.sql` when running current code |
| Event recorder is wired and `USER_EVENTS_ENABLED` defaults to `true` | Apply migration `003_create_user_outbox_events.up.sql` or set `USER_EVENTS_ENABLED=false` for Task 4-only smoke testing |
| Outbox worker is disabled by default | RabbitMQ/Kafka is not needed unless `OUTBOX_WORKER_ENABLED=true` |

If you are only testing Task 4 gRPC handlers with unit tests, MySQL is not required. If you are starting the real current service with `go run ./cmd/server`, apply migrations in numeric order.

Minimal current-repo migration order:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p < backend/services/user-service/migrations/001_create_user_tables.up.sql
mysql -h 127.0.0.1 -P 3306 -u root -p < backend/services/user-service/migrations/002_add_user_audit_fields.up.sql
mysql -h 127.0.0.1 -P 3306 -u root -p < backend/services/user-service/migrations/003_create_user_outbox_events.up.sql
```

Note: Migration `002` and `003` are later-task runtime requirements in the current repo. They are mentioned here only to avoid local setup confusion.

## 6. Redis / Queue / External Services

Task 4 gRPC setup needs no Redis, Kafka, RabbitMQ, NATS, MinIO, Elasticsearch, SMTP, Stripe, Firebase, or OAuth provider.

| External service | Required for Task 4? | Explanation |
|---|---:|---|
| MySQL | Yes for real runtime | Existing profile/seller tables are read/written through repositories. |
| Docker | Optional | Helpful only to run MySQL locally. |
| grpcurl | Optional dev tool | Useful to manually verify gRPC methods. |
| Buf | Optional unless proto changes | Needed for proto lint/generation. |
| RabbitMQ/Kafka | No | Later event/outbox worker only. Keep worker disabled for Task 4. |

Current repo event safety for Task 4-only local runs:

```env
USER_EVENTS_ENABLED=false
OUTBOX_WORKER_ENABLED=false
```

Use this only when you are intentionally avoiding later Task 8 event setup. Full event setup should be documented with Task 8, not duplicated here.

## 7. Environment Variables

### Reused `.env`

Base `.env` file location, loading process, DSN, gRPC address, reflection, DB pool, shutdown timeout, and logging variables are already documented:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: 4. Environment Variables
```

Task 4 does not introduce a new required `.env` variable.

Important reused variables for Task 4:

| Variable | Required? | Task 4 usage |
|---|---:|---|
| `USER_SERVICE_DATABASE_DSN` | Yes for real service | MySQL connection used by repositories behind gRPC methods |
| `USER_SERVICE_GRPC_ADDRESS` | Optional | gRPC listen address, default `:50052` |
| `USER_SERVICE_GRPC_REFLECTION` | Optional | Enables `grpcurl list/describe` locally |
| `USER_SERVICE_SHUTDOWN_TIMEOUT` | Optional | Graceful shutdown wait time |
| `USER_SERVICE_LOG_LEVEL` | Optional | Controls structured log verbosity |

### Task 4 Runtime Metadata

These are not `.env` variables. They are gRPC metadata headers sent by internal callers.

| Metadata key | Required? | Purpose |
|---|---:|---|
| `x-request-id` | Recommended | Trace one request across services/logs |
| `x-user-id` | Required for user-owned update flows | Current user identity for ownership checks |
| `x-service-name` | Recommended for service callers | Identifies Auth/API Gateway/internal service |
| `x-actor-id` | Required for writes if no user/service fallback is available | Audit actor id |
| `x-actor-type` | Required with `x-actor-id` when type cannot be inferred | Valid values: `user`, `admin`, `service`, `system` |
| `x-roles` | Optional/currently useful for future RBAC | Comma-separated roles like `buyer,seller` |

Beginner note:

Writes like `CreateUser` and `UpdateUserProfile` may fail with `audit actor is required` if caller metadata is missing. This is setup/caller-context issue, not a MySQL issue.

## 8. Docker Setup

No new Dockerfile, `docker-compose.yml`, container, volume, network, or health check is introduced by Task 4.

Use previous Docker setup:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: 7. Docker and DevOps Setup
```

Current repo status:

| Docker item | Status |
|---|---|
| User service Dockerfile | Not found |
| Project docker-compose | Not found |
| MySQL container | Optional local setup from previous docs |
| gRPC app container | Not available yet |
| Queue containers | Not required for Task 4 |

Recommended beginner path remains:

1. Run MySQL locally or through Docker.
2. Run Go service directly with `go run`.
3. Use `grpcurl` from host machine against `localhost:50052`.

## 9. Local Development Setup

### Step 1: Read Previous Dependency Documentation

Follow these first:

| Step | File | Why |
|---:|---|---|
| 1 | `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Go, MySQL, Docker, `.env`, ports, grpcurl |
| 2 | `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | MySQL schema and migration |
| 3 | `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | Repository layer DB expectations |

### Step 2: Go To Project Directory

For service commands:

```bash
cd backend/services/user-service
```

For workspace-wide commands:

```bash
cd backend
```

### Step 3: Install Only Needed Dependencies

No new Task 4 dependency install is needed if previous docs were followed.

Fresh clone command:

```bash
cd backend/services/user-service
go mod download
```

### Step 4: Generate Proto Only If Needed

If `proto/ecommerce/user/v1/user.proto` changed:

```bash
cd proto
buf lint
buf generate
```

If generated files are already present and up to date, skip this.

### Step 5: Run Task 4 gRPC Handler Tests

These tests use fake usecases and do not need MySQL:

```bash
cd backend/services/user-service
go test ./internal/transport/grpc
```

This is the fastest Task 4 setup verification.

### Step 6: Setup Database For Real Runtime

For original Task 4 runtime, use Task 2 migration setup.

For current repo runtime, apply migrations in numeric order as noted in section 5.

### Step 7: Add Environment Variables

Reuse `.env` from `task1_Dependency.md`.

For Task 4-only local smoke testing on current repo, optional event isolation:

```env
USER_EVENTS_ENABLED=false
OUTBOX_WORKER_ENABLED=false
```

### Step 8: Start Backend Service

```bash
cd backend/services/user-service
set -a
. ./.env
set +a
go run ./cmd/server
```

Expected log signal:

```text
user_service_grpc_listening
```

## 10. Running the Project

### Verify Reflection

Reflection must be enabled locally:

```env
USER_SERVICE_GRPC_REFLECTION=true
```

Then run:

```bash
grpcurl -plaintext localhost:50052 list
grpcurl -plaintext localhost:50052 describe ecommerce.user.v1.UserService
```

### Verify Task 4 Methods

Create user with service caller metadata:

```bash
grpcurl -plaintext \
  -H 'x-request-id: local-task4-create-user' \
  -H 'x-service-name: auth-service' \
  -H 'x-actor-id: auth-service' \
  -H 'x-actor-type: service' \
  -d '{"auth_account_id":"auth_123","email":"buyer@example.com","phone":"+919999999999","full_name":"Aarav Sharma"}' \
  localhost:50052 ecommerce.user.v1.UserService/CreateUser
```

Get user:

```bash
grpcurl -plaintext \
  -H 'x-request-id: local-task4-get-user' \
  -H 'x-service-name: api-gateway' \
  -d '{"user_id":"user_123"}' \
  localhost:50052 ecommerce.user.v1.UserService/GetUser
```

Update user profile with `FieldMask`:

```bash
grpcurl -plaintext \
  -H 'x-request-id: local-task4-update-user' \
  -H 'x-user-id: user_123' \
  -H 'x-actor-id: user_123' \
  -H 'x-actor-type: user' \
  -d '{"user_id":"user_123","full_name":"Aarav S.","update_mask":{"paths":["full_name"]}}' \
  localhost:50052 ecommerce.user.v1.UserService/UpdateUserProfile
```

Get seller profile:

```bash
grpcurl -plaintext \
  -H 'x-request-id: local-task4-get-seller' \
  -H 'x-service-name: product-service' \
  -d '{"seller_id":"seller_123"}' \
  localhost:50052 ecommerce.user.v1.UserService/GetSellerProfile
```

Important:

- IDs in examples must exist in your local database.
- `CreateUser` creates a new generated `user_id`; use the returned id for later calls.
- `GetSellerProfile` needs a seeded seller row unless your local flow already created one.

## 11. Common Errors & Fixes

Generic Go/MySQL/Docker/gRPC errors already exist in `task1_Dependency.md` and `task2_Dependency.md`. Task 4 specific additions:

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `audit actor is required` | Write call missing caller/audit metadata | Add `x-actor-id` + `x-actor-type`, or at least `x-user-id`/`x-service-name` | Keep grpcurl examples with metadata headers |
| `permission denied` | `x-user-id` does not match target `user_id` | Use same user id for self-service update or call as trusted service | Gateway should inject correct user id |
| `validation failed` | Current repo validation rejects email/phone/name format | Send valid email, E.164-ish phone, non-empty names | Reuse realistic test data |
| `Table 'user_db.users' doesn't exist` | Task 2 migration not applied | Apply `001_create_user_tables.up.sql` | Run migration before real runtime |
| `Unknown column 'created_by'` or `Unknown column 'deleted_at'` | Current repo audit migration missing | Apply `002_add_user_audit_fields.up.sql` | Apply migrations in numeric order on current repo |
| `Table 'user_db.user_outbox_events' doesn't exist` | Current repo event recorder enabled but outbox migration missing | Apply `003_create_user_outbox_events.up.sql` or set `USER_EVENTS_ENABLED=false` for Task 4-only smoke | Decide whether you are testing Task 4 only or full current service |
| `server does not support reflection` | `USER_SERVICE_GRPC_REFLECTION=false` | Set it to `true` locally and restart service | Keep reflection local-only |
| `unknown service ecommerce.user.v1.UserService` | Generated proto/service registration mismatch or wrong server | Run `buf generate`, rebuild, confirm correct port | Do not edit generated files manually |
| `InvalidArgument` for `update_mask` | Empty/invalid mask in profile update | Send `{"paths":["full_name"]}` or the field you want to patch | Always include update mask for patch calls |
| `DeadlineExceeded` | Caller deadline too short or DB slow | Increase local timeout and check MySQL | Use realistic client timeouts and DB indexes |

## 12. Security & Best Practices

Reuse previous security notes for `.env`, DB credentials, root user, Docker, reflection, and plaintext gRPC:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: 10. Security and Configuration Audit
```

Task 4 specific best practices:

- Keep this gRPC service internal-only. Public browser clients should go through API Gateway.
- Do not trust metadata from public traffic. Gateway/service mesh should authenticate caller identity before forwarding metadata.
- Disable or restrict gRPC reflection outside local/dev networks.
- Do not log raw email, phone, address, KYC URLs, tokens, or DB DSNs.
- Keep domain errors mapped to safe gRPC status codes. Raw SQL/internal errors should not leak to callers.
- Add client deadlines for every gRPC call. Handler should pass `ctx` through to usecase/repository.
- Do not import proto packages inside domain/usecase layers. Proto belongs at transport boundary.
- Do not return gRPC `status.Error` from repository or domain. Map errors in `internal/transport/grpc/errors.go`.
- Regenerate proto code with `buf generate`; never hand-edit generated `*.pb.go` files.
- Use TLS/mTLS or private network/service mesh for production gRPC traffic.

## 13. Missing or Misconfigured Things

| Finding | Impact | Suggested fix |
|---|---|---|
| No committed `.env.example` found | Beginners can miss required env names | Add sanitized `.env.example` with fake DSN and local gRPC settings |
| No User Service Dockerfile found | App cannot be containerized directly yet | Add Dockerfile in DevOps task |
| No project `docker-compose.yml` found | Beginners manually start MySQL/migrations | Add local compose with MySQL healthcheck and optional service container |
| No standard gRPC health service found | Orchestrators cannot check gRPC readiness cleanly | Add `grpc_health_v1` health service |
| gRPC metadata is trusted in app layer | Spoofed headers are risky if service exposed publicly | Enforce gateway-only/private network/mTLS and auth interceptors |
| Reflection defaults useful locally but risky publicly | Method discovery exposure | Set `USER_SERVICE_GRPC_REFLECTION=false` in production |
| Current repo has later event defaults | Task 4 local setup can fail due outbox table | For Task 4-only smoke use `USER_EVENTS_ENABLED=false`, or apply migration `003` |
| `docs/04-microservice-design.md` mentions `BatchGetUsers`, but proto currently does not include it | Potential future contract mismatch | Add only when a future task explicitly needs it |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `1. Project Tech Stack Analysis` | Base Go/MySQL/gRPC stack already explained |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `2. Language-Specific Dependency System: Go` | Go modules, `go.mod`, `go.sum`, `go.work`, commands already documented |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3. Database Analysis` | MySQL install, Docker setup, DSN, credentials, default port already documented |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Environment Variables` | Base `.env`, gRPC address, reflection, shutdown, DB pool, logging already documented |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `5. External Services Analysis -> gRPC / Protobuf and Buf` | grpcurl, gRPC reflection, Buf generation already explained |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `6. Ports and Networking` | User Service gRPC `50052`, MySQL `3306`, alternate MySQL port already documented |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. Docker and DevOps Setup` | Docker MySQL and missing app Dockerfile already covered |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `9. Common Errors and Fixes` | Generic `.env`, MySQL, Docker, reflection, Go module errors already covered |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` | Table creation, migration, rollback, verification queries already covered |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `11. Common Errors & Fixes` | Missing tables, FK errors, migration permission issues already covered |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `7. Database Setup Required For Task 3` | Repository layer depends on migrated MySQL schema |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `12. Repository-Specific Common Errors And Fixes` | Timestamp scan, duplicate key, foreign key, schema mismatch issues already explained |

## 15. Final Checklist

- [ ] Previous dependency documentation checked first.
- [ ] `INPUT_FILE_PATH` reviewed for Task 4 scope.
- [ ] No duplicate Go/MySQL/Docker setup copied from previous files.
- [ ] Go `1.24+` available.
- [ ] `go mod download` completed if this is a fresh clone.
- [ ] `proto/ecommerce/user/v1/user.proto` checked.
- [ ] `buf lint` and `buf generate` run only if proto changed.
- [ ] Generated Go proto files present under `backend/shared/gen/go/ecommerce/user/v1/`.
- [ ] Task 4 handler tests pass with `go test ./internal/transport/grpc`.
- [ ] MySQL running if starting real service.
- [ ] Required migrations applied for the runtime version being tested.
- [ ] `.env` sourced before `go run ./cmd/server`.
- [ ] `USER_SERVICE_GRPC_REFLECTION=true` for local grpcurl testing.
- [ ] Event/outbox settings intentionally chosen for current repo local run.
- [ ] gRPC service lists with `grpcurl -plaintext localhost:50052 list`.
- [ ] `CreateUser`, `GetUser`, `UpdateUserProfile`, and `GetSellerProfile` verified with correct metadata.
- [ ] Reflection disabled/restricted for non-local environments.
- [ ] No secrets committed in `.env`.
- [ ] Logs checked for `user_service_grpc_listening`.
- [ ] No original implementation file modified.
