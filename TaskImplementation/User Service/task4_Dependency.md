# Project Dependency & Setup Guide

This guide is generated for:

```text
TaskImplementation/User Service/task4.md
```

Output file:

```text
TaskImplementation/User Service/task4_Dependency.md
```

Task 4 ka scope hai User Service ka internal gRPC service setup: proto contract, generated Go code, gRPC server registration, request metadata, interceptors, reflection, and `grpcurl` based local testing.

Important reuse rule:

- Go install, base Go module setup, MySQL install, Docker MySQL setup, migration setup, `.env` baseline, and repository DB troubleshooting already previous dependency files me documented hai.
- Is file me wahi content repeat nahi kiya gaya.
- Sirf Task 4 ke new/changed setup, gRPC tooling, proto workflow, ports, env notes, and debugging steps detail me explain kiye gaye hain.

---

## 1. Project Overview

Task 4 User Service ko internal gRPC API banata hai.

Simple Hinglish:

gRPC ek fast service-to-service communication system hai. Browser direct gRPC call nahi karega. API Gateway, Auth Service, Order Service, Product/CMS Service jaise internal services User Service se user/seller profile data lene ke liye gRPC call karenge.

Analyzed files:

| File | Purpose |
|---|---|
| `TaskImplementation/User Service/task4.md` | Task 4 implementation guide |
| `TaskImplementation/User Service/task1_Dependency.md` | Base Go, MySQL, Docker, env, and service run guide |
| `TaskImplementation/User Service/task2_Dependency.md` | MySQL schema and migration setup |
| `TaskImplementation/User Service/task3_Dependency.md` | Repository dependency and DB behavior guide |
| `proto/ecommerce/user/v1/user.proto` | User Service gRPC contract |
| `proto/buf.yaml` | Buf lint and breaking-change config |
| `proto/buf.gen.yaml` | Proto to Go generated-code config |
| `backend/shared/gen/go/ecommerce/user/v1/user.pb.go` | Generated protobuf message structs |
| `backend/shared/gen/go/ecommerce/user/v1/user_grpc.pb.go` | Generated gRPC client/server interfaces |
| `backend/services/user-service/internal/transport/grpc/server.go` | User gRPC method handlers |
| `backend/services/user-service/internal/transport/grpc/metadata.go` | Incoming gRPC metadata extraction |
| `backend/services/user-service/internal/transport/grpc/interceptors.go` | Unary logging and panic recovery interceptors |
| `backend/services/user-service/internal/transport/grpc/errors.go` | Domain error to gRPC status mapping |
| `backend/services/user-service/internal/transport/grpc/mapper.go` | Domain object to proto response mapping |
| `backend/services/user-service/cmd/server/main.go` | gRPC server startup and reflection registration |
| `backend/services/user-service/internal/config/config.go` | Runtime env variables |
| `backend/services/user-service/go.mod` | Go dependency versions |

Current Task 4 dependency status:

| Area | Status |
|---|---|
| Language | Go 1.24+ |
| Transport | gRPC unary APIs |
| API contract | Protocol Buffers |
| Proto tooling | Buf config present |
| Generated Go code | Present under `backend/shared/gen/go` |
| Database | Same MySQL from Task 1/Task 2 |
| Repository | Same repository layer from Task 3 |
| Redis/Kafka/RabbitMQ | Not required for Task 4 |
| Dockerfile/docker-compose | No new project Docker files added |
| Main new developer tools | `buf`, `protoc-gen-go`, `protoc-gen-go-grpc`, `grpcurl` |

---

## 2. Tech Stack

### Task 4 Technologies

| Technology | Required? | What it is | Why Task 4 uses it |
|---|---:|---|---|
| Go 1.24+ | Yes | Go ek compiled backend language hai. | User Service backend and gRPC server Go me implemented hai. |
| Go Modules | Yes | `go.mod` dependencies list karta hai, `go.sum` checksums lock karta hai. | gRPC/protobuf packages versioned form me manage hote hain. |
| Go Workspace | Yes | `go.work` local Go modules ko ek workspace me connect karta hai. | `user-service` local generated proto module `backend/shared/gen/go` use karta hai. |
| gRPC | Yes | gRPC high-performance internal API framework hai. | User Service internal services ko strongly typed methods expose karta hai. |
| Protocol Buffers | Yes | `.proto` file request/response schema define karti hai. | Service contract stable and type-safe banta hai. |
| Buf | Required when proto changes | Proto linting, breaking checks, and code generation tool. | `proto/buf.gen.yaml` ke through Go code generate hota hai. |
| `protoc-gen-go` | Required when proto changes | Protobuf Go message generator. | `user.pb.go` generate karta hai. |
| `protoc-gen-go-grpc` | Required when proto changes | Go gRPC service generator. | `user_grpc.pb.go` generate karta hai. |
| `grpcurl` | Recommended | Terminal se gRPC call test karne ka CLI. | Local server ko browser/curl ke bina verify karne ke liye. |
| gRPC reflection | Local dev recommended | Running gRPC service apne methods expose karta hai for discovery. | `grpcurl list/describe` commands ko easy banata hai. |
| MySQL | Yes for runtime | Relational database. | gRPC handler usecase/repository ke through user data read/write karta hai. |

### Beginner Explanation

`user.proto` ek contract hai. Is contract se Go code generate hota hai. Generated code me request structs, response structs, server interface, and client interface milte hain.

Flow simple hai:

```text
user.proto
  -> buf generate
  -> backend/shared/gen/go/.../user.pb.go
  -> backend/shared/gen/go/.../user_grpc.pb.go
  -> User Service gRPC handler compiles against generated interfaces
```

Layer rule:

| Layer | Allowed responsibility |
|---|---|
| Proto | Contract define karna |
| Generated code | Type-safe messages and service interface |
| gRPC handler | Request validation-lite, metadata read, usecase call, response mapping |
| Usecase | Business rules and ownership checks |
| Repository | MySQL queries |

### Reused Tech From Previous Tasks

Do not duplicate base setup. Use these existing docs:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 2. Tech Stack
Section: 3. Required Software
Section: 4. Dependency Management

TaskImplementation/User Service/task2_Dependency.md
Section: 5. Database Setup

TaskImplementation/User Service/task3_Dependency.md
Section: 4. Dependency Management
```

---

## 3. Required Software

### Required For Task 4 Development

| Software | Required? | Why |
|---|---:|---|
| Git | Yes | Repo clone/pull ke liye. |
| Go 1.24+ | Yes | Service build/test/run ke liye. |
| MySQL 8.x | Yes for runtime | gRPC methods repository/usecase ke through DB use karte hain. |
| MySQL client | Recommended | Migration and table verification ke liye. |
| Docker | Optional | MySQL container run karne ke liye. |
| Buf CLI | Required if proto is edited/regenerated | Proto lint and generated Go code create karne ke liye. |
| `protoc-gen-go` | Required if proto is edited/regenerated | Go protobuf structs generate karne ke liye. |
| `protoc-gen-go-grpc` | Required if proto is edited/regenerated | Go gRPC server/client code generate karne ke liye. |
| `grpcurl` | Recommended | Local gRPC APIs manually test karne ke liye. |

### Already Documented Base Installs

| Setup need | Refer |
|---|---|
| Install Git/Go/Docker/MySQL | `TaskImplementation/User Service/task1_Dependency.md` -> `3. Required Software` |
| Start MySQL with Docker | `TaskImplementation/User Service/task1_Dependency.md` -> `5. Database Setup` |
| Apply MySQL schema | `TaskImplementation/User Service/task2_Dependency.md` -> `5. Database Setup` |

### Task 4 Tool Installation

Task 1 mentions `buf` and `grpcurl` as optional tools. Task 4 makes them important for proto/gRPC work.

Install Go-based tools:

```bash
go install github.com/bufbuild/buf/cmd/buf@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

Verify:

```bash
buf --version
protoc-gen-go --version
protoc-gen-go-grpc --version
grpcurl -version
```

If command not found aaye:

```bash
go env GOPATH
```

Make sure this folder is in `PATH`:

```text
$(go env GOPATH)/bin
```

Beginner note:

`go install ...@latest` tool binary install karta hai. Ye app dependency nahi hoti, but developer machine par CLI available hona chahiye.

---

## 4. Dependency Management

### Go Module Files

Already explained in:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 4. Dependency Management
```

Task 4 relevant files:

| File | Why important |
|---|---|
| `backend/services/user-service/go.mod` | Runtime gRPC/protobuf dependencies listed here. |
| `backend/services/user-service/go.sum` | Dependency checksum lock file. |
| `backend/go.work` | Connects `user-service` and generated proto module. |
| `backend/shared/gen/go/go.mod` | Generated proto Go module. |
| `proto/buf.yaml` | Buf module, lint, and breaking config. |
| `proto/buf.gen.yaml` | Generated Go output config. |

### Task 4 Dependency Delta

| Dependency | Type | Required at runtime? | Task 4 use |
|---|---|---:|---|
| `google.golang.org/grpc v1.72.2` | Go module | Yes | gRPC server, status codes, interceptors, reflection registration. |
| `google.golang.org/protobuf v1.36.6` | Go module | Yes | Generated proto messages, `FieldMask`, `Timestamp`. |
| `github.com/parag/ecommerce/backend/shared/gen/go v0.0.0` | Local module | Yes | Generated User Service protobuf/gRPC code. |
| `github.com/go-sql-driver/mysql v1.9.3` | Go module | Yes | Reused DB access from previous tasks. |
| `github.com/DATA-DOG/go-sqlmock v1.5.2` | Go module | Test only | Reused repository tests; not new in Task 4. |
| `buf` CLI | Developer tool | No | Generate and validate proto code. |
| `grpcurl` CLI | Developer tool | No | Manually test gRPC APIs. |

### Commands

Download app dependencies:

```bash
cd backend/services/user-service
go mod download
```

Clean dependency graph only when imports change:

```bash
cd backend/services/user-service
go mod tidy
```

Run all User Service tests:

```bash
cd backend/services/user-service
go test ./...
```

Run only gRPC transport tests:

```bash
cd backend/services/user-service
go test ./internal/transport/grpc
```

From workspace root:

```bash
cd backend
go test ./services/user-service/...
```

### Proto Dependency Workflow

Proto config lives under:

```text
proto/buf.yaml
proto/buf.gen.yaml
```

Run proto generation from `proto/` folder:

```bash
cd proto
buf lint
buf generate
```

Expected generated output:

```text
backend/shared/gen/go/ecommerce/user/v1/user.pb.go
backend/shared/gen/go/ecommerce/user/v1/user_grpc.pb.go
```

Important:

- Generated files manually edit mat karo.
- Proto change karo, phir `buf generate` run karo.
- Generated code commit karna required hai if repo convention generated code track karta hai. Is repo me generated files already present hain.

### Common Go/Proto Dependency Issues

| Error | Reason | Fix |
|---|---|---|
| `module ... backend/shared/gen/go not found` | `backend/shared/gen/go` missing hai or workspace path wrong hai. | Full repo clone karo and `backend/go.work` verify karo. |
| `no required module provides package .../ecommerce/user/v1` | Generated proto code missing/stale hai. | `cd proto` then `buf generate`. |
| `buf: command not found` | Buf CLI installed nahi hai or PATH me nahi hai. | Install command run karo and `$(go env GOPATH)/bin` PATH me add karo. |
| `protoc-gen-go: program not found` | Generator binary missing hai. | `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`. |
| `protoc-gen-go-grpc: program not found` | gRPC generator binary missing hai. | `go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest`. |
| Generated files changed unexpectedly | Different generator version or stale output. | Team-approved tool versions align karo, then regenerate once. |
| `go: module requires go >= 1.24` | Local Go old version hai. | Go 1.24+ install karo. Refer Task 1 dependency doc. |

---

## 5. Database Setup

### Detected Database: MySQL

Task 4 does not add a new database.

It reuses the same MySQL setup from Tasks 1 and 2 because gRPC methods call usecase and repository code that reads/writes `user_db`.

Full database setup is already documented:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 5. Database Setup

TaskImplementation/User Service/task2_Dependency.md
Section: 5. Database Setup
```

Use these exact existing sections:

| Need | Existing doc |
|---|---|
| What MySQL is | `task1_Dependency.md` -> `Detected Database: MySQL` -> `A. What It Is` |
| Why User Service uses MySQL | `task1_Dependency.md` -> `B. Why This Project Uses It` |
| Local MySQL installation | `task1_Dependency.md` -> `D. Local Installation` |
| Docker MySQL setup | `task1_Dependency.md` -> `E. Docker Setup` |
| Docker Compose example | `task1_Dependency.md` -> `F. Docker Compose Example` |
| Start/verify MySQL | `task1_Dependency.md` -> `G. Start Commands`, `H. Verify Running` |
| DSN format | `task1_Dependency.md` -> `J. Connection String Format` |
| Credentials placement | `task1_Dependency.md` -> `K. Where To Place Credentials` |
| Run migration | `task1_Dependency.md` -> `L. Run Migration` |
| Task 2 schema verification | `task2_Dependency.md` -> `Schema-Specific Verification Queries` |

### Task 4 DB Runtime Requirement

Task 4 gRPC server startup does this:

```text
Load env -> open MySQL -> ping MySQL -> create repositories -> create usecase -> start gRPC listener
```

So MySQL must be available before running the service.

Required database state:

| Requirement | Why |
|---|---|
| `user_db` database exists | DSN points to it. |
| Task 2 migration applied | gRPC methods eventually query `users` and `seller_profiles`. |
| `parseTime=true` in DSN | Timestamp columns map to Go `time.Time`. |
| App DB user has SELECT/INSERT/UPDATE permissions | `CreateUser`, `GetUser`, `UpdateUserProfile`, `GetSellerProfile` need DB access. |

### Connection String Reminder

This is not new for Task 4. It is repeated only because gRPC startup depends on DB ping.

```env
USER_SERVICE_DATABASE_DSN=ecommerce_user:ecommerce_password@tcp(127.0.0.1:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC
```

Credential placement:

```text
backend/services/user-service/.env
```

Do not commit `.env`.

---

## 6. Redis / Queue / External Services

Task 4 does not introduce Redis, Kafka, RabbitMQ, NATS, MinIO, Elasticsearch, SMTP, Stripe, Twilio, Firebase, Nginx, Kubernetes, or any new cloud service.

| Service | Required for Task 4? | Notes |
|---|---:|---|
| Redis | No | No Redis client/env var/cache code in current gRPC task. |
| Kafka/RabbitMQ/NATS | No | User events are future scope, not Task 4. |
| MinIO/S3 | No | Seller/KYC binary storage is not part of Task 4. |
| Elasticsearch/Typesense | No | Search service is separate future area. |
| SMTP/Twilio/Stripe/Firebase | No | No third-party integration introduced. |
| Docker | Optional | Useful only for MySQL local setup. |
| gRPC reflection | Yes for local debugging, optional in production | Controlled by `USER_SERVICE_GRPC_REFLECTION`. |

Reuse previous external-service explanation:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 6. Redis / Queue / External Services

TaskImplementation/User Service/task3_Dependency.md
Section: 6. Redis / Queue / External Services
```

### gRPC Reflection

Reflection ek gRPC debugging feature hai.

Simple Hinglish:

Reflection on hone par `grpcurl` running server se pooch sakta hai ki kaunse services and methods available hain. Isse local debugging easy hoti hai.

Current config:

| Env var | Default | Meaning |
|---|---|---|
| `USER_SERVICE_GRPC_REFLECTION` | `true` | Local/dev me service discovery enabled. |

Production recommendation:

```env
USER_SERVICE_GRPC_REFLECTION=false
```

Why:

Production me reflection se internal service contract discoverable ho sakta hai. Agar service private network me bhi hai, policy ke according reflection disable karna safer hota hai.

---

## 7. Environment Variables

Task 4 introduces no brand-new env variable names beyond the existing User Service config, but it makes the gRPC-related variables important.

Base env documentation:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 7. Environment Variables
```

### Task 4 Relevant Env Variables

| Env variable | Required? | Default | Task 4 purpose |
|---|---:|---|---|
| `USER_SERVICE_DATABASE_DSN` | Yes | None | MySQL DSN. Service startup fails without it. |
| `MYSQL_DSN` | Optional fallback | None | Used only if `USER_SERVICE_DATABASE_DSN` is blank. |
| `USER_SERVICE_GRPC_ADDRESS` | Optional | `:50052` | Address/port where gRPC server listens. |
| `USER_SERVICE_GRPC_REFLECTION` | Optional | `true` | Enables `grpcurl list/describe` support. |
| `USER_SERVICE_SHUTDOWN_TIMEOUT` | Optional | `10s` | Graceful gRPC shutdown wait time. |
| `USER_SERVICE_LOG_LEVEL` | Optional | `info` | JSON log level for gRPC request logs. |
| `USER_SERVICE_DB_MAX_OPEN_CONNS` | Optional | `25` | DB pool setting reused from Task 1. |
| `USER_SERVICE_DB_MAX_IDLE_CONNS` | Optional | `25` | DB pool setting reused from Task 1. |
| `USER_SERVICE_DB_CONN_MAX_LIFETIME` | Optional | `5m` | DB connection lifetime. |
| `USER_SERVICE_DB_PING_TIMEOUT` | Optional | `5s` | Startup DB ping timeout. |

### Task-Specific `.env` Delta

Use the full base `.env` from Task 1. Add or confirm only these gRPC-specific values for Task 4:

```env
USER_SERVICE_GRPC_ADDRESS=:50052
USER_SERVICE_GRPC_REFLECTION=true
USER_SERVICE_SHUTDOWN_TIMEOUT=10s
USER_SERVICE_LOG_LEVEL=debug
```

For production-like local testing:

```env
USER_SERVICE_GRPC_REFLECTION=false
USER_SERVICE_LOG_LEVEL=info
```

Important:

- `USER_SERVICE_DATABASE_DSN` is still required.
- `USER_SERVICE_GRPC_ADDRESS=:50052` means listen on all interfaces on port `50052`.
- If port `50052` is busy, use another port like `:50053` and update all `grpcurl` commands.

### gRPC Metadata Headers

Task 4 reads incoming gRPC metadata:

| Metadata key | Required? | Used for |
|---|---:|---|
| `x-user-id` | Required for user ownership checks on user-facing calls | Current user identity from API Gateway/Auth boundary. |
| `x-service-name` | Recommended | Internal caller identity, for logs/policy. |
| `x-roles` | Optional now | Future RBAC behavior. Supports comma-separated roles. |
| `x-request-id` | Recommended | Request tracing in logs. |

Beginner note:

Metadata gRPC ka header system hai. REST me jaise headers hote hain, gRPC me metadata hota hai.

Example `grpcurl` metadata:

```bash
grpcurl -plaintext \
  -H 'x-user-id: user_123' \
  -H 'x-service-name: api-gateway' \
  -H 'x-roles: buyer,seller' \
  -H 'x-request-id: req_local_001' \
  -d '{"user_id":"user_123"}' \
  localhost:50052 ecommerce.user.v1.UserService/GetUser
```

Security note:

Local testing me metadata manually bhej sakte ho. Production me API Gateway/service mesh ko trusted metadata inject karna chahiye. Public clients ko direct User Service gRPC access nahi milna chahiye.

---

## 8. Docker Setup

Task 4 adds no Dockerfile, docker-compose service, volume, network, or container image.

Docker remains useful only for MySQL, same as earlier tasks.

Reuse:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 8. Docker Setup
Section: 5. Database Setup -> E. Docker Setup

TaskImplementation/User Service/task2_Dependency.md
Section: 8. Docker Setup
```

### Task 4 Docker Impact

| Docker item | New in Task 4? | Notes |
|---|---:|---|
| User Service Dockerfile | No | Service still runs with `go run` locally. |
| gRPC container port | No Docker mapping yet | If Dockerfile is added later, expose `50052`. |
| MySQL container | Reused | Same setup from Task 1/Task 2. |
| MySQL volume | Reused | Preserve DB data and schema. |
| Docker network | Not added | Needed only when app and DB both run in containers. |
| Healthcheck | Not added for User Service | Recommended future improvement. |

### Future Docker Notes

If a future Dockerfile is added for User Service, it should:

- Build Go binary from `backend/services/user-service`.
- Include generated proto module from `backend/shared/gen/go`.
- Expose gRPC port `50052`.
- Read all secrets from env, not baked image files.
- Use non-root runtime user.
- Have a health check strategy, such as gRPC health checking.

No such Dockerfile exists in current Task 4 implementation.

---

## 9. Local Development Setup

This is the Task 4 clone-to-test flow with previous docs reused for common setup.

### Step 1: Clone Repository

Follow:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 9. Local Development Setup -> Step 1: Clone Repository
```

### Step 2: Install Base Tools

Follow:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 3. Required Software
```

You need:

```text
Go 1.24+
MySQL 8.x or Docker
MySQL client
```

### Step 3: Install Task 4 gRPC/Proto Tools

```bash
go install github.com/bufbuild/buf/cmd/buf@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

Verify:

```bash
buf --version
grpcurl -version
```

### Step 4: Download Go Dependencies

```bash
cd backend/services/user-service
go mod download
```

### Step 5: Start MySQL And Apply Migration

Do not repeat DB setup. Follow:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: Step 4: Start MySQL

TaskImplementation/User Service/task2_Dependency.md
Section: Step 3: Apply Task 2 Up Migration
Section: Step 4: Verify Tables, Indexes, And FKs
```

### Step 6: Configure `.env`

Use base `.env` from:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 7. Environment Variables
```

Confirm these Task 4 values:

```env
USER_SERVICE_GRPC_ADDRESS=:50052
USER_SERVICE_GRPC_REFLECTION=true
USER_SERVICE_SHUTDOWN_TIMEOUT=10s
```

### Step 7: Validate Proto Files

Run from `proto/`:

```bash
cd proto
buf lint
```

If proto changed, regenerate:

```bash
cd proto
buf generate
```

Then run Go tests:

```bash
cd ../backend/services/user-service
go test ./...
```

### Step 8: Run gRPC Transport Tests

```bash
cd backend/services/user-service
go test ./internal/transport/grpc
```

These tests verify:

| Test area | Meaning |
|---|---|
| Required request fields | Missing IDs return `InvalidArgument`. |
| Error mapping | Domain errors become correct gRPC status codes. |
| FieldMask mapping | `update_mask.paths` pass to usecase. |
| Metadata extraction | `x-user-id`, `x-service-name`, roles reach usecase caller input. |

### Step 9: Run User Service

Load env using the method from Task 1, then:

```bash
cd backend/services/user-service
go run ./cmd/server
```

Expected behavior:

```text
Service loads config
Service pings MySQL
Service starts gRPC listener on :50052
Service logs JSON messages to stdout
```

### Step 10: Verify gRPC Reflection

In another terminal:

```bash
grpcurl -plaintext localhost:50052 list
```

Expected service:

```text
ecommerce.user.v1.UserService
```

Describe:

```bash
grpcurl -plaintext localhost:50052 describe ecommerce.user.v1.UserService
```

If reflection is disabled, `list` and `describe` may fail. You can still call methods if you provide proto descriptors or enable reflection locally.

---

## 10. Running The Project

### Run Order

| Order | Action | Why |
|---:|---|---|
| 1 | Start MySQL | Service startup pings DB. |
| 2 | Apply Task 2 migration | gRPC calls need tables. |
| 3 | Set `USER_SERVICE_DATABASE_DSN` | Service needs DB credentials. |
| 4 | Set/confirm `USER_SERVICE_GRPC_ADDRESS` | Service needs listen port. |
| 5 | Run `go run ./cmd/server` | Starts gRPC server. |
| 6 | Test with `grpcurl` | Confirms gRPC endpoint is reachable. |

### Quick Commands

Common setup commands are in previous docs. Task 4-specific commands:

```bash
cd proto
buf lint
buf generate
```

```bash
cd backend/services/user-service
go test ./internal/transport/grpc
go test ./...
go run ./cmd/server
```

```bash
grpcurl -plaintext localhost:50052 list
grpcurl -plaintext localhost:50052 describe ecommerce.user.v1.UserService
```

### Manual gRPC Calls

Create user:

```bash
grpcurl -plaintext \
  -H 'x-service-name: auth-service' \
  -H 'x-request-id: req_create_user_001' \
  -d '{"auth_account_id":"auth_123","email":"buyer@example.com","phone":"+919999999999","full_name":"Aarav Sharma"}' \
  localhost:50052 ecommerce.user.v1.UserService/CreateUser
```

Get user:

```bash
grpcurl -plaintext \
  -H 'x-user-id: user_123' \
  -H 'x-service-name: api-gateway' \
  -H 'x-request-id: req_get_user_001' \
  -d '{"user_id":"user_123"}' \
  localhost:50052 ecommerce.user.v1.UserService/GetUser
```

Update user profile:

```bash
grpcurl -plaintext \
  -H 'x-user-id: user_123' \
  -H 'x-service-name: api-gateway' \
  -H 'x-request-id: req_update_user_001' \
  -d '{"user_id":"user_123","full_name":"Aarav S.","update_mask":{"paths":["full_name"]}}' \
  localhost:50052 ecommerce.user.v1.UserService/UpdateUserProfile
```

Get seller profile:

```bash
grpcurl -plaintext \
  -H 'x-service-name: product-service' \
  -H 'x-request-id: req_seller_001' \
  -d '{"seller_id":"seller_123"}' \
  localhost:50052 ecommerce.user.v1.UserService/GetSellerProfile
```

Beginner note:

These calls need matching DB rows. If user/seller does not exist, response `NotFound` aayega. That is expected behavior, not setup failure.

### Ports & Networking

| Service | Port | Purpose | New in Task 4? |
|---|---:|---|---:|
| User Service gRPC | `50052` | Internal gRPC API | Yes, Task 4 focuses on this API |
| MySQL host port | `3306` | Local DB access | No, reused |
| MySQL alternate host port | `3307` | Avoid local conflict | No, reused from Task 2 docs |
| Redis | `6379` | Not used | No |
| Kafka | `9092` | Not used | No |

Port rule:

- If `:50052` is busy, set `USER_SERVICE_GRPC_ADDRESS=:50053`.
- Update `grpcurl` host to `localhost:50053`.
- MySQL port and gRPC port are different. Do not put MySQL port in `grpcurl`.

---

## 11. Common Errors & Fixes

Base errors already documented:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 11. Common Errors & Fixes

TaskImplementation/User Service/task2_Dependency.md
Section: 11. Common Errors & Fixes

TaskImplementation/User Service/task3_Dependency.md
Section: 11. Common Errors & Fixes
```

Task 4-specific errors:

### `grpcurl: command not found`

Cause:

`grpcurl` installed nahi hai or Go bin folder `PATH` me nahi hai.

Fix:

```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
go env GOPATH
```

Add this to shell PATH:

```text
$(go env GOPATH)/bin
```

### `Failed to list services: server does not support the reflection API`

Cause:

`USER_SERVICE_GRPC_REFLECTION=false` hai or reflection register nahi hui.

Fix for local:

```env
USER_SERVICE_GRPC_REFLECTION=true
```

Restart service, then:

```bash
grpcurl -plaintext localhost:50052 list
```

Production note:

Production me reflection intentionally disabled ho sakti hai. Ye error production setup me normal ho sakta hai.

### `connection refused` on `localhost:50052`

Cause:

Service running nahi hai, wrong port use ho raha hai, or startup DB ping fail hua.

Fix:

1. Service terminal logs check karo.
2. `USER_SERVICE_DATABASE_DSN` valid hai ya nahi check karo.
3. MySQL running hai ya nahi verify karo.
4. `USER_SERVICE_GRPC_ADDRESS` ka port confirm karo.

### `listen tcp :50052: bind: address already in use`

Cause:

Port `50052` already occupied hai.

Fix:

```env
USER_SERVICE_GRPC_ADDRESS=:50053
```

Then:

```bash
grpcurl -plaintext localhost:50053 list
```

### `rpc error: code = InvalidArgument desc = user_id is required`

Cause:

Request payload me `user_id` missing/blank hai.

Fix:

```bash
grpcurl -plaintext \
  -d '{"user_id":"user_123"}' \
  localhost:50052 ecommerce.user.v1.UserService/GetUser
```

### `rpc error: code = InvalidArgument desc = invalid request` during update

Cause:

`update_mask.paths` missing hai or unsupported field diya gaya hai.

Allowed update mask paths:

```text
full_name
phone
avatar_url
```

Fix:

```bash
grpcurl -plaintext \
  -H 'x-user-id: user_123' \
  -d '{"user_id":"user_123","full_name":"Aarav S.","update_mask":{"paths":["full_name"]}}' \
  localhost:50052 ecommerce.user.v1.UserService/UpdateUserProfile
```

### `rpc error: code = PermissionDenied desc = permission denied`

Cause:

`x-user-id` metadata request ke `user_id` se match nahi karta.

Current usecase rule:

```text
If x-user-id is blank, service-level calls pass.
If x-user-id is present, it must match requested user_id.
```

Fix for user-facing call:

```bash
grpcurl -plaintext \
  -H 'x-user-id: user_123' \
  -d '{"user_id":"user_123"}' \
  localhost:50052 ecommerce.user.v1.UserService/GetUser
```

Security note:

Blank `x-user-id` being treated as service-level access is a trust-boundary assumption. Production should enforce service identity at gateway/service mesh/interceptor level.

### `rpc error: code = NotFound desc = user not found`

Cause:

Setup can be fine, but DB me requested user row nahi hai.

Fix:

- Confirm Task 2 migration applied.
- Create user via `CreateUser` first.
- Check `users` table in MySQL.

### `buf: command not found`

Cause:

Buf CLI missing or PATH issue.

Fix:

```bash
go install github.com/bufbuild/buf/cmd/buf@latest
```

Confirm Go bin path in `PATH`.

### `Failure: plugin protoc-gen-go: could not find protoc plugin`

Cause:

Go protobuf generator missing.

Fix:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Generated Code Stale After Proto Change

Symptoms:

- Go compile errors in `internal/transport/grpc`.
- Method exists in proto but not generated interface.
- Request field not found in Go struct.

Fix:

```bash
cd proto
buf generate
cd ../backend/services/user-service
go test ./...
```

### `go test ./internal/transport/grpc` Fails With Import Error

Cause:

Generated proto module missing or workspace not resolved.

Fix:

```bash
cd backend
go test ./services/user-service/internal/transport/grpc
```

Also verify:

```text
backend/go.work
backend/shared/gen/go/go.mod
backend/shared/gen/go/ecommerce/user/v1/user.pb.go
```

---

## 12. Security & Best Practices

### Secrets

Reuse:

```text
TaskImplementation/User Service/task1_Dependency.md
Section: 12. Security & Best Practices -> Secrets
```

Task 4-specific:

- Do not put DB passwords in proto files, generated code, or task docs with real values.
- Keep `.env` local and uncommitted.
- gRPC metadata can contain user IDs and request IDs. Avoid logging full tokens or secrets.

### gRPC Access Boundary

User Service gRPC should be internal-only.

Best practices:

- Public internet se `50052` expose mat karo.
- API Gateway/Auth Service should validate JWT before calling User Service.
- Service-to-service auth should be added in production, usually mTLS, service mesh identity, or signed internal token.
- Reflection local/dev me useful hai, production me disable karna safer hai unless there is a clear ops requirement.

### Metadata Trust

Current code reads:

```text
x-user-id
x-service-name
x-roles
x-request-id
```

Best practices:

- Metadata trusted source se aaye, such as API Gateway.
- Client-provided metadata directly trust mat karo.
- `x-request-id` log-friendly rakho, secret data mat rakho.
- `x-roles` future RBAC ke liye validate/sign hona chahiye.

### Error Handling

Current code maps domain errors to gRPC status codes:

| Domain/error type | gRPC code |
|---|---|
| `context.Canceled` | `Canceled` |
| `context.DeadlineExceeded` | `DeadlineExceeded` |
| `ErrUserNotFound`, `ErrSellerNotFound` | `NotFound` |
| `ErrDuplicateUser`, `ErrDuplicateSeller` | `AlreadyExists` |
| `ErrForbidden` | `PermissionDenied` |
| `ErrDeletedResource` | `FailedPrecondition` |
| validation/invalid argument | `InvalidArgument` |
| unknown error | `Internal` |

Best practice:

Internal SQL details caller ko leak nahi karne. Detailed error logs server side me rakho.

### Deadlines And Timeouts

Clients should use deadlines:

| Method | Suggested local timeout |
|---|---:|
| `GetUser` | `300ms-700ms` |
| `CreateUser` | `1s-1.5s` |
| `UpdateUserProfile` | `700ms-1s` |
| `GetSellerProfile` | `300ms-700ms` |

Why:

Deadline nahi hoga to stuck network/DB issue resources hold kar sakta hai.

### Proto Best Practices

- Existing field numbers delete/reuse mat karo.
- Field rename carefully karo because JSON name/client behavior impact ho sakta hai.
- Breaking changes ke liye future `v2` package use karo.
- Generated files manually edit mat karo.
- `FieldMask` use karo for partial updates.

### Logging

Current interceptors log:

```text
method
code
duration
request_id
```

Best practices:

- Email/phone/KYC document values logs me avoid/mask karo.
- Panic recovery enabled hai, but root cause fix karna still required hai.
- `USER_SERVICE_LOG_LEVEL=debug` local me ok hai; production me `info` or stricter use karo.

---

## 13. Missing or Misconfigured Things

This section is an audit of setup/devops gaps found while inspecting Task 4 and current code.

| Area | Current status | Risk | Suggested fix |
|---|---|---|---|
| `.env.example` | Not present | New developers may not know required vars. | Add sanitized `backend/services/user-service/.env.example`. |
| User Service Dockerfile | Not present | Containerized local/prod run not documented as actual file. | Add Dockerfile in future DevOps task. |
| Project docker-compose | Not present | Beginners must run MySQL manually or create temporary compose. | Add local compose with MySQL and optional service container. |
| gRPC health check | Not implemented | Load balancer/orchestrator cannot easily check service health. | Add standard gRPC health service. |
| Service-to-service auth interceptor | Not implemented | Metadata can be spoofed if network boundary is weak. | Add mTLS/internal auth interceptor before production. |
| Reflection default | `true` | Useful locally, risky if exposed in prod. | Set false in production env. |
| Migration runner | Plain SQL only | Service startup does not auto-migrate. | Add migration tool or documented migration command in deployment pipeline. |
| Buf breaking check baseline | Config present, baseline not documented | Breaking changes may slip if CI does not run `buf breaking`. | Add CI step for `buf lint` and `buf breaking`. |
| Tool version pinning | Uses latest install commands | Different dev machines can generate different output. | Pin tool versions in docs or tool manifest. |
| Public port exposure policy | Not enforced by code | `:50052` could be exposed accidentally. | Restrict network, firewall, compose/K8s service type. |

### Hardcoded Credentials Audit

No hardcoded DB credentials were found in Task 4 gRPC implementation files.

Credential source:

```text
USER_SERVICE_DATABASE_DSN
MYSQL_DSN fallback
```

Risk:

The DSN contains username/password, so logs and docs should not print real production DSNs.

### Config Audit

Current config validates:

| Config | Validation |
|---|---|
| `USER_SERVICE_DATABASE_DSN` | Required |
| `USER_SERVICE_DB_MAX_OPEN_CONNS` | Must be greater than zero |
| `USER_SERVICE_DB_MAX_IDLE_CONNS` | Cannot be negative |
| `USER_SERVICE_SHUTDOWN_TIMEOUT` | Must be positive |
| `USER_SERVICE_DB_PING_TIMEOUT` | Must be positive |

Potential improvement:

Invalid int/bool/duration env values currently fall back silently. For production, fail-fast validation is usually better because typoed env vars should be visible.

---

## 14. References to Previous Dependency Files

Use these files instead of duplicating setup:

| Topic | Reference |
|---|---|
| Base project overview | `TaskImplementation/User Service/task1_Dependency.md` -> `1. Project Overview` |
| Go install | `TaskImplementation/User Service/task1_Dependency.md` -> `3. Required Software` |
| Go modules, `go.mod`, `go.sum`, `go.work` | `TaskImplementation/User Service/task1_Dependency.md` -> `4. Dependency Management` |
| MySQL local install | `TaskImplementation/User Service/task1_Dependency.md` -> `5. Database Setup` |
| Docker MySQL setup | `TaskImplementation/User Service/task1_Dependency.md` -> `5. Database Setup -> E. Docker Setup` |
| Docker Compose example | `TaskImplementation/User Service/task1_Dependency.md` -> `5. Database Setup -> F. Docker Compose Example` |
| Base `.env` | `TaskImplementation/User Service/task1_Dependency.md` -> `7. Environment Variables` |
| Base run commands | `TaskImplementation/User Service/task1_Dependency.md` -> `10. Running The Project` |
| Base common errors | `TaskImplementation/User Service/task1_Dependency.md` -> `11. Common Errors & Fixes` |
| MySQL schema | `TaskImplementation/User Service/task2_Dependency.md` -> `5. Database Setup` |
| Migration verification | `TaskImplementation/User Service/task2_Dependency.md` -> `Schema-Specific Verification Queries` |
| Repository dependencies | `TaskImplementation/User Service/task3_Dependency.md` -> `4. Dependency Management` |
| Repository DB requirements | `TaskImplementation/User Service/task3_Dependency.md` -> `5. Database Setup` |
| Repository troubleshooting | `TaskImplementation/User Service/task3_Dependency.md` -> `11. Common Errors & Fixes` |

---

## 15. Final Checklist

Task 4 setup checklist:

- [ ] Read `TaskImplementation/User Service/task1_Dependency.md` for base Go/MySQL/Docker/env setup.
- [ ] Read `TaskImplementation/User Service/task2_Dependency.md` for migration setup.
- [ ] Read `TaskImplementation/User Service/task3_Dependency.md` for repository DB requirements.
- [ ] Install Go 1.24+.
- [ ] Install Task 4 tools: `buf`, `protoc-gen-go`, `protoc-gen-go-grpc`, `grpcurl`.
- [ ] Confirm `proto/ecommerce/user/v1/user.proto` exists.
- [ ] Confirm `proto/buf.yaml` and `proto/buf.gen.yaml` exist.
- [ ] Run `cd proto` and `buf lint`.
- [ ] If proto changed, run `buf generate`.
- [ ] Confirm generated files exist under `backend/shared/gen/go/ecommerce/user/v1/`.
- [ ] Start MySQL.
- [ ] Apply Task 2 migration.
- [ ] Create/load `.env` with `USER_SERVICE_DATABASE_DSN`.
- [ ] Confirm `USER_SERVICE_GRPC_ADDRESS=:50052`.
- [ ] Keep `USER_SERVICE_GRPC_REFLECTION=true` for local `grpcurl` testing.
- [ ] Run `cd backend/services/user-service` and `go mod download`.
- [ ] Run `go test ./internal/transport/grpc`.
- [ ] Run `go test ./...`.
- [ ] Run `go run ./cmd/server`.
- [ ] Verify `grpcurl -plaintext localhost:50052 list`.
- [ ] Test `CreateUser`, `GetUser`, `UpdateUserProfile`, and `GetSellerProfile` with metadata.
- [ ] For production, disable reflection unless explicitly needed.
- [ ] Do not expose port `50052` publicly.
- [ ] Do not commit `.env` or real DSN credentials.

Final beginner rule:

Task 4 ka setup tab successful maana jayega jab MySQL running ho, migration applied ho, service `:50052` par start ho, `grpcurl list` User Service show kare, and `go test ./internal/transport/grpc` pass ho.
