# Superadmin Service - Go Modules Dependency

## 1. What Is This Dependency?

Go modules Go project ka dependency management system hai. `go.mod` batata hai ki module ka naam kya hai, Go version kya hai, and kaunsi external libraries use hoti hain.

## 2. Why This Service Uses It

Superadmin Service Go me implemented hai. Build, test, dependency download, and workspace registration sab Go modules/workspace ke through manage hota hai.

## 3. Required Or Optional

Required. Without Go toolchain and modules, service build/start nahi hoga.

## 4. Where It Is Used In Project

| Path | Usage |
|------|-------|
| `backend/services/superadmin-service/go.mod` | Module declaration and dependencies |
| `backend/services/superadmin-service/go.sum` | Dependency checksum lock |
| `backend/go.work` | Workspace registration |
| `backend/services/superadmin-service/cmd/server/main.go` | Main executable package |

## 5. Installation Steps

Install Go version compatible with:

```text
go 1.26.3
```

Check local Go:

```bash
go version
```

Download modules:

```bash
cd backend/services/superadmin-service
go mod download
```

## 6. Docker Setup, If Possible

Service Dockerfile is not clearly found in project files.

Suggested command based on project structure, if Dockerfile is added later:

```bash
docker build -t ecommerce-superadmin-service backend/services/superadmin-service
```

Current status: Missing Dockerfile.

## 7. Local Setup Without Docker

```bash
cd backend/services/superadmin-service
go mod tidy
go test ./...
go run ./cmd/server
```

Note: `go mod tidy` can update `go.mod`/`go.sum`; run it only when dependency changes are intended.

## 8. Required Environment Variables

Go modules do not require service env vars by themselves, but running the service needs:

| Variable | Needed For |
|----------|------------|
| `SUPERADMIN_DATABASE_DSN` | DB-backed repositories |
| `HTTP_ADDR` | HTTP server address |
| `USER_SERVICE_ADMIN_BASE_URL` | User/seller admin flows |
| `ORDER_SERVICE_ADMIN_BASE_URL` | Order admin flows |
| `PAYMENT_SERVICE_ADMIN_BASE_URL` | Payment/refund admin flows |

## 9. Start Commands

Exact command is not clearly found in project files.

Suggested command based on project structure:

```bash
cd backend/services/superadmin-service
go run ./cmd/server
```

With env:

```bash
cd backend/services/superadmin-service
set -a
source .env
set +a
go run ./cmd/server
```

## 10. Verify Running Commands

```bash
cd backend/services/superadmin-service
go test ./...
go list ./...
```

Check service:

```bash
curl http://127.0.0.1:8088/healthz
```

## 11. Common Errors And Fixes

| Error | Reason | Fix |
|-------|--------|-----|
| `go: go.mod file not found` | Command run from wrong directory | Run inside `backend/services/superadmin-service` |
| `module requires Go 1.26.3` | Older Go installed | Install compatible Go version |
| Missing checksum in `go.sum` | Dependency changed | Run `go mod tidy` |
| MySQL driver import error | Module download incomplete | Run `go mod download` |
| Config validation error | Env vars not exported | Source `.env` or export variables |

## 12. Security Notes

- Do not add secrets to `go.mod`, code, or tests.
- Keep dependency versions reviewed.
- Run `go mod verify` in CI if possible.
- Avoid adding unused gRPC/protobuf libraries until real gRPC code exists.

## 13. Final Checklist

- [x] `go.mod` exists.
- [x] `go.sum` exists.
- [x] MySQL driver is declared.
- [x] `backend/go.work` includes this service.
- [x] Entrypoint exists at `cmd/server/main.go`.
- [ ] No Makefile or documented exact start command found.
