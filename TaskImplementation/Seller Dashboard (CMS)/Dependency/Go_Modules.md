# Go Modules Dependency - Seller Dashboard (CMS)

## 1. What is this dependency?

Go Modules Go backend ke dependencies manage karte hain. `go.mod` service-level module define karta hai, `go.sum` dependency checksums store karta hai, and `go.work` multiple modules ko local workspace me connect karta hai.

## 2. Why this service uses it

Seller Dashboard (CMS) backend side Go microservices par depend karta hai:

- API Gateway
- CMS Service
- Product Service
- Order Service
- User/Auth boundary

In services ko build/run/test karne ke liye proper Go modules required hain.

## 3. Required or optional

Required for backend services.

Frontend ke liye Go Modules required nahi hain.

## 4. Where it is used in project

| Path | Use |
|---|---|
| `backend/go.work` | Workspace file |
| `backend/go.work.sum` | Workspace checksum file |
| `backend/services/cms-service/` | Expected CMS service module location |
| `backend/services/api-gateway/` | Expected Gateway module location |

Current issue:

- `backend/go.work` currently only includes `./services/auth-service`.
- `backend/services/auth-service` was not clearly found.
- `backend/services/cms-service/go.mod` was not found.
- `backend/services/api-gateway/go.mod` was not found.

## 5. Installation steps

Install Go.

```bash
go version
```

Suggested command based on project structure after service modules exist:

```bash
cd backend/services/cms-service
go mod tidy
```

## 6. Docker setup, if possible

Not clearly found in project files.

When Dockerfiles are added, they should copy the correct service `go.mod` and shared modules before build.

## 7. Local setup without Docker

Expected flow after modules exist:

```bash
cd backend
go work use ./services/api-gateway
go work use ./services/cms-service
go work sync
```

Then:

```bash
cd backend/services/cms-service
go test ./...
go run ./cmd/server
```

These are suggested commands based on project structure.

## 8. Required environment variables

Go Modules itself does not require app env vars.

Useful local env:

| Variable | Purpose |
|---|---|
| `GOWORK` | Controls workspace usage |
| `GOPRIVATE` | Private module domains if any |
| `GONOSUMDB` | Private module checksum behavior if any |

Service runtime env vars are documented in `Environment.md`.

## 9. Start commands

Suggested command based on project structure:

```bash
cd backend/services/cms-service
go run ./cmd/server
```

Current status: Not runnable until Go source/module files exist.

## 10. Verify running commands

Suggested command based on project structure:

```bash
cd backend/services/cms-service
go test ./...
```

Workspace verify:

```bash
cd backend
go work edit -json
```

## 11. Common errors and fixes

| Error | Reason | Fix |
|---|---|---|
| `go: cannot find main module` | `go.mod` missing | Create service module |
| `directory ../auth-service does not exist` | `go.work` references missing module | Update `go.work` with real module paths |
| Missing package import | Generated proto/shared package missing | Generate proto or add shared module |
| `go.sum` changed | Dependency updated | Review and commit checksum changes |

## 12. Security notes

- Do not use unreviewed public modules for auth/crypto casually.
- Pin versions through `go.mod` and review `go.sum` changes.
- Private modules should use `GOPRIVATE`.
- Keep generated code reproducible in CI.

## 13. Final checklist

- [x] `backend/go.work` found.
- [x] `backend/go.work.sum` found.
- [ ] CMS service `go.mod` not found.
- [ ] API Gateway `go.mod` not found.
- [ ] CMS service not registered in `go.work`.
- [ ] API Gateway not registered in `go.work`.
- [ ] `go.work` references a module path not clearly present.

