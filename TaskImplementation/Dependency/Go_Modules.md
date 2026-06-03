# User Service - Go Modules Dependency

## 1. What is this dependency?

Go ek compiled backend language hai. Go Modules dependency manager hai jo `go.mod` me packages list karta hai and `go.sum` me checksum lock karta hai.

Go Workspace (`go.work`) multiple local Go modules ko ek saath connect karta hai.

## 2. Why User Service uses it

User Service Go me implemented hai. Is service ko MySQL driver, gRPC runtime, protobuf runtime, generated proto module, and test helper packages chahiye.

## 3. Required or Optional

| Item | Required? | Notes |
|---|---:|---|
| Go 1.24 compatible toolchain | Yes | `go.mod` and `go.work` both `go 1.24` declare karte hain |
| Go Modules | Yes | Backend build/test/run ke liye |
| Go Workspace | Yes for local repo workflow | `user-service` and `shared/gen/go` modules connect karta hai |
| go-sqlmock | Test only | Repository unit tests ke liye |

## 4. Where it is used in project

| Path | Purpose |
|---|---|
| `backend/services/user-service/go.mod` | User Service module and dependencies |
| `backend/services/user-service/go.sum` | Dependency checksums |
| `backend/shared/gen/go/go.mod` | Generated protobuf Go module |
| `backend/shared/gen/go/go.sum` | Generated module checksums |
| `backend/go.work` | Local workspace registration |
| `backend/go.work.sum` | Workspace checksums |

Current direct dependencies:

| Module | Why used |
|---|---|
| `github.com/go-sql-driver/mysql` | MySQL database driver |
| `github.com/parag/ecommerce/backend/shared/gen/go` | Generated User Service protobuf/gRPC code |
| `google.golang.org/grpc` | gRPC server/runtime |
| `google.golang.org/protobuf` | Protobuf message/timestamp/field mask runtime |
| `github.com/DATA-DOG/go-sqlmock` | Repository unit tests |

## 5. Installation Steps

Install Go 1.24 compatible version.

Verify:

```bash
go version
```

Download dependencies:

```bash
cd backend/services/user-service
go mod download
```

If imports change, suggested command based on project structure:

```bash
cd backend/services/user-service
go mod tidy
```

## 6. Docker Setup, If Possible

No User Service Dockerfile was found. Go module download inside Docker is not configured in project files.

Suggested future Docker pattern:

```Dockerfile
FROM golang:1.24 AS build
WORKDIR /app/backend
COPY backend/go.work backend/go.work.sum ./
COPY backend/shared/gen/go ./shared/gen/go
COPY backend/services/user-service ./services/user-service
WORKDIR /app/backend/services/user-service
RUN go test ./...
RUN go build -o /out/user-service ./cmd/server
```

This is a suggested example only, because no checked-in Dockerfile currently exists.

## 7. Local Setup Without Docker

Use the workspace from `backend/`:

```bash
cd backend
go work sync
```

Run tests:

```bash
cd backend/services/user-service
go test ./...
```

Start service after MySQL/env setup:

```bash
cd backend/services/user-service
go run ./cmd/server
```

The start command is suggested based on project structure.

## 8. Required Environment Variables

Go Modules itself does not require service env vars.

Runtime needs these important variables:

| Variable | Needed for Go module download? | Needed for service runtime? |
|---|---:|---:|
| `USER_SERVICE_DATABASE_DSN` | No | Yes |
| `MYSQL_DSN` | No | Conditional fallback |
| `USER_SERVICE_GRPC_ADDRESS` | No | Optional |

## 9. Start Commands

Backend start requires env + MySQL first:

```bash
cd backend/services/user-service
set -a
. ./.env
set +a
go run ./cmd/server
```

## 10. Verify Running Commands

Check modules/tests:

```bash
cd backend/services/user-service
go test ./...
```

Check workspace includes service:

```bash
cd backend
go work edit -json
```

Expected workspace modules include:

```text
./services/user-service
./shared/gen/go
```

## 11. Common Errors and Fixes

| Error | Cause | Fix |
|---|---|---|
| `go: go.mod requires go >= 1.24` | Old Go version | Install Go 1.24 compatible version |
| `module ... shared/gen/go ... not found` | Workspace/replace path issue | Run from repo layout or keep `replace ../../shared/gen/go` valid |
| `missing go.sum entry` | Dependency checksum missing | Run `go mod tidy` |
| `package ... userv1 not found` | Generated proto module missing | Verify `backend/shared/gen/go/ecommerce/user/v1` exists |
| Tests fail after proto change | Generated code stale | Run `buf generate` from `proto/` |

## 12. Security Notes

- `go.sum` ko commit karna important hai. Ye dependency tampering detect karne me help karta hai.
- Random internet code ko `go get` karne se pehle module source verify karo.
- Private modules use hon to `GOPRIVATE` configure karo.
- Production builds me pinned versions use karo, only `@latest` blindly mat use karo.

## 13. Final Checklist

| Check | Done |
|---|---|
| Go 1.24 compatible version installed | [ ] |
| `backend/services/user-service/go.mod` present | [ ] |
| `backend/services/user-service/go.sum` present | [ ] |
| `backend/go.work` includes User Service | [ ] |
| `backend/go.work` includes generated proto module | [ ] |
| `go mod download` completed | [ ] |
| `go test ./...` passes | [ ] |

