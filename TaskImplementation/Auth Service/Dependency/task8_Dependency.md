# Project Dependency & Setup Guide

Input implementation file:

```text
TaskImplementation/Auth Service/task8.md
```

Generated dependency file:

```text
TaskImplementation/Auth Service/task8_Dependency.md
```

This guide is for **Auth Service - Task 8: Security Tests**.

Simple goal: beginner developer ko clearly samajh aaye ki token expiry, refresh token reuse, OTP replay, aur role bypass security tests run karne ke liye kaunsi dependencies, test commands, environment assumptions, aur debugging steps chahiye.

Important reuse rule followed:

- Previous Auth Service dependency files were inspected first.
- Common clone, Go install, Go modules, MySQL install, Redis install, Docker setup, JWT key setup, OTP setup, RBAC setup, session-link setup, `.env`, migrations, and full local run flow already documented hai.
- Is file me wahi setup repeat nahi kiya gaya. Sirf Task 8 security-test specific setup and differences explain kiye gaye hain.

Previous dependency files inspected:

```text
TaskImplementation/Auth Service/Dependency/task1_Dependency.md
TaskImplementation/Auth Service/Dependency/task2_Dependency.md
TaskImplementation/Auth Service/Dependency/task3_Dependency.md
TaskImplementation/Auth Service/Dependency/task4_Dependency.md
TaskImplementation/Auth Service/Dependency/task5_Dependency.md
TaskImplementation/Auth Service/Dependency/task6_Dependency.md
TaskImplementation/Auth Service/task7_Dependency.md
```

Current repository note:

- `task8.md` is a security testing guide.
- Current repo already contains Task 8 related security tests under `backend/services/auth-service/internal/...`.
- Task 8 introduces **no new runtime database**, **no new Redis/Kafka/RabbitMQ service**, **no new Docker container**, **no new public port**, and **no new `.env` variable**.
- Focused Task 8 tests mostly use Go fakes, generated test RSA keys, `httptest`, in-memory buffers, and table-driven test cases.
- Full Auth Service startup still needs the baseline setup from previous dependency docs.

---

## 1. Project Overview

Task 8 ka main goal hai Auth Service ke risky security behavior ko automated tests se lock karna.

Security test ka simple meaning:

```text
Happy path proves feature works.
Security test proves misuse fail hota hai.
```

Task 8 verifies:

| Security area | What must be proved | Current test files |
|---|---|---|
| Access token expiry and claims | Expired, forged, wrong issuer/audience, missing claims, wrong token type reject hote hain | `internal/security/token/token_test.go` |
| Refresh token reuse | Old refresh token dobara use hone par session/token family revoke hoti hai | `internal/usecase/token_security_test.go` |
| OTP replay | Verified OTP dobara use nahi ho sakta | `internal/usecase/otp_usecase_test.go` |
| Role bypass | Buyer/seller/admin wrong access nahi le sakte | `internal/authorization/rbac_security_test.go`, `internal/usecase/role_security_test.go` |
| HTTP security response | Invalid token/OTP/reuse/role errors generic response dete hain and secrets log me leak nahi hote | `internal/transport/http/security_test.go` |

Task 8 does **not** add:

- New business API
- New migration
- New table
- New external Go module
- New Dockerfile or compose file
- New queue/broker dependency
- New service port
- New production secret

Baseline full service setup already exists in:

```text
TaskImplementation/Auth Service/Dependency/task1_Dependency.md
Sections:
3. Required Software
4. Dependency Management
5. Database Setup
6. Redis / Queue / External Services
7. Environment Variables
8. Docker Setup
9. Local Development Setup
10. Running the Project
```

---

## 2. Tech Stack

### Task 8 specific technologies

| Technology | What it is | Why Task 8 uses it | Required? | Beginner explanation |
|---|---|---|---:|---|
| Go | Backend programming language | Security tests and Auth Service code Go me hain | Yes | Go ek compiled backend language hai. Tests bhi same module me Go se run hote hain. |
| Go `testing` package | Built-in test framework | Unit/security test functions run karne ke liye | Yes | `go test` files ending with `_test.go` ko compile and run karta hai. |
| Table-driven tests | Go testing style | Many invalid token/role cases ek clean table me test karne ke liye | Yes | Ek table me cases define hote hain, loop me sab cases run hote hain. |
| `httptest` | Go standard HTTP test helper | Middleware and error responses bina real server start kiye test karne ke liye | Yes | Fake request/response recorder se HTTP behavior check hota hai. |
| Generated test RSA keys | In-memory RSA key pair | JWT sign/verify tests ke liye real local PEM file ki zarurat avoid karne ke liye | Yes for token tests | Test apni temporary key generate karta hai, production key use nahi hoti. |
| Fake repositories | In-memory test doubles | Refresh/OTP/role flows ko DB ke bina test karne ke liye | Yes | Fake repo se MySQL start kiye bina security logic test hota hai. |
| `sync.WaitGroup` concurrency test | Go concurrency helper | Same refresh token parallel use karne par only one success prove karne ke liye | Yes | Race condition catch karne me help karta hai. |
| `log/slog` with buffer | Structured log capture | Secret redaction prove karne ke liye | Yes | Test logs ko memory buffer me capture karke check karta hai ki token/OTP leak nahi hua. |

### Already documented baseline technologies

| Baseline topic | Reuse documentation |
|---|---|
| Go `1.26.3`, Go modules, Go workspace | `TaskImplementation/Auth Service/Dependency/task1_Dependency.md`, section `4. Dependency Management` |
| MySQL 8+ setup | `TaskImplementation/Auth Service/Dependency/task1_Dependency.md`, section `5. Database Setup` |
| Redis setup | `TaskImplementation/Auth Service/Dependency/task1_Dependency.md`, section `6. Redis / Queue / External Services` |
| JWT RS256, JWKS, refresh token pepper | `TaskImplementation/Auth Service/Dependency/task4_Dependency.md` |
| OTP HMAC, Redis rate limit, Notification endpoint | `TaskImplementation/Auth Service/Dependency/task5_Dependency.md` |
| RBAC roles and role assignment setup | `TaskImplementation/Auth Service/Dependency/task6_Dependency.md` |
| Session link and outbox setup | `TaskImplementation/Auth Service/task7_Dependency.md` |

---

## 3. Required Software

### For Task 8 focused security tests only

| Software | Required? | Purpose | Verify command |
|---|---:|---|---|
| Go | Yes | Compile and run Auth Service tests | `go version` |
| Git | Yes if cloning fresh | Clone repo and inspect changes | `git --version` |
| Internet access or existing Go module cache | Usually yes | Download modules if cache is empty | `go mod download` |

Focused Task 8 tests do **not** need these services:

| Service/tool | Needed for focused Task 8 tests? | Why |
|---|---:|---|
| MySQL server | No | Token/OTP/role security tests use fakes for these scenarios |
| Redis server | No | OTP rate limit Redis behavior is not required for Task 8 focused commands |
| Docker | No | No dependency container is needed for focused tests |
| JWT PEM files | No | Token tests generate RSA keys in memory |
| Notification Service | No | OTP replay tests do not need real OTP delivery |
| Session Service | No | Session-link tests use fakes/outbox helpers |
| Kafka/RabbitMQ/NATS | No | Current Auth Service has no direct broker client |

### For full Auth Service startup

Full service startup is different from focused tests. `cmd/server/main.go` validates and pings multiple dependencies.

| Software/service | Required for full server? | Setup reference |
|---|---:|---|
| Go | Yes | `task1_Dependency.md`, section `3` |
| MySQL 8+ | Yes | `task1_Dependency.md`, section `5` |
| Redis | Yes | `task1_Dependency.md`, section `6` |
| JWT RSA key files | Yes | `task4_Dependency.md`, sections `7`, `8`, `9` |
| Notification endpoint or mock | Config required, real OTP send needs it | `task5_Dependency.md`, sections `6`, `7`, `9` |
| Session event outbox config | Required unless disabled | `task7_Dependency.md`, sections `6`, `7`, `9` |
| Docker | Optional | `task1_Dependency.md`, section `8` |

---

## 4. Dependency Management

This is still a Go module project.

Complete beginner explanation of:

- `go.mod`
- `go.sum`
- Go modules
- `go mod download`
- `go mod tidy`
- `go build`
- `go run`
- common Go module failures

already exists in:

```text
TaskImplementation/Auth Service/Dependency/task1_Dependency.md
Section: 4. Dependency Management
```

### Current `go.mod` status

Current file:

```text
backend/services/auth-service/go.mod
```

Direct dependencies currently present:

```text
github.com/go-sql-driver/mysql v1.9.3
golang.org/x/crypto v0.51.0
```

Task 8 does **not** require a new external package. Do not run `go get` for Testcontainers, miniredis, gRPC, Kafka, RabbitMQ, or a JWT library only for this task. Current tests use Go standard library plus the existing module code.

### Task 8 useful test commands

Download dependencies if needed:

```bash
cd backend/services/auth-service
go mod download
```

Run token expiry and JWT claim tests:

```bash
cd backend/services/auth-service
go test ./internal/security/token
```

Run refresh reuse, OTP replay, and role usecase security tests:

```bash
cd backend/services/auth-service
go test ./internal/usecase -run 'RefreshToken|Logout|VerifyOTP|AssignRole|GetUserRoles|RequireFreshRole'
```

Run RBAC helper and HTTP middleware security tests:

```bash
cd backend/services/auth-service
go test ./internal/authorization ./internal/transport/http -run 'Bypass|Permission|AuthRequired|RequireRoles|UsecaseError'
```

Run all Auth Service tests:

```bash
cd backend/services/auth-service
go test ./...
```

Run race check for concurrency-sensitive security tests:

```bash
cd backend/services/auth-service
CGO_ENABLED=1 go test -race ./internal/usecase ./internal/authorization ./internal/transport/http
```

Note: `-race` requires cgo and a C compiler such as `gcc`. If your local machine does not have that toolchain, run normal `go test` locally and keep the race pass in CI or on a machine with compiler support.

If Go build cache path creates permission issue:

```bash
cd backend/services/auth-service
GOCACHE=/tmp/auth-go-cache go test ./...
```

---

## 5. Database Setup

### Database detected

Task 8 introduces no new database.

Focused Task 8 tests:

```text
MySQL required: No
Redis required: No
Migrations required: No
```

### Why no new DB setup is needed

Security tests use fake repositories and in-memory fixtures for the risky scenarios:

| Scenario | Real DB needed for focused test? | Reason |
|---|---:|---|
| JWT expiry | No | JWT is signed and verified in memory |
| Refresh token reuse | No | Fake refresh token repository stores rows in memory |
| OTP replay | No | Fake OTP repository simulates challenge state |
| Role bypass | No | Auth context and fake role repo provide roles |
| HTTP error redaction | No | `httptest` and fake usecases are enough |

### Full service database requirement

Full Auth Service still uses MySQL tables from previous tasks:

| Table area | Why used | Setup reference |
|---|---|---|
| `auth_accounts`, `credentials` | Account and password state | `task1_Dependency.md`, `task2_Dependency.md`, `task3_Dependency.md` |
| `refresh_tokens` | Refresh token rotation/reuse detection | `task4_Dependency.md` |
| `otp_challenges` | OTP expiry, attempts, replay protection | `task5_Dependency.md` |
| `role_assignments` | RBAC source of truth | `task6_Dependency.md` |
| `auth_outbox_events` | Session-link event delivery | `task7_Dependency.md` |

If you are only running the Task 8 focused tests, do not start MySQL and do not apply migrations just for this task.

If you are running the full Auth Service or doing manual API verification, follow:

```text
TaskImplementation/Auth Service/Dependency/task1_Dependency.md
Section: 5. Database Setup

TaskImplementation/Auth Service/Dependency/task4_Dependency.md
Task-specific token migrations and refresh token setup

TaskImplementation/Auth Service/Dependency/task5_Dependency.md
OTP table setup

TaskImplementation/Auth Service/Dependency/task6_Dependency.md
RBAC table hardening

TaskImplementation/Auth Service/task7_Dependency.md
Outbox migration setup
```

---

## 6. Redis / Queue / External Services

Task 8 introduces no new Redis, queue, broker, or third-party integration.

### Focused test requirement

| External service | Needed for Task 8 focused tests? | Explanation |
|---|---:|---|
| Redis | No | OTP replay tests use fakes and do not require real rate-limit keys |
| Kafka | No | Current Auth Service has no Kafka client |
| RabbitMQ | No | Current Auth Service has no RabbitMQ client |
| NATS | No | Current Auth Service has no NATS client |
| Notification Service | No | Replay tests do not send real OTP |
| Session Management Service | No | Token reuse tests use a fake session linker |
| MinIO/S3/Stripe/Twilio/Firebase | No | Not part of Auth Service Task 8 |

### Full service external dependencies

Full server startup still initializes:

- Redis client for OTP rate limit repository
- gRPC Notification client for OTP delivery
- Session-link outbox mode unless `SESSION_LINK_MODE=disabled`

These are already documented in:

```text
TaskImplementation/Auth Service/Dependency/task1_Dependency.md
Section: 6. Redis / Queue / External Services

TaskImplementation/Auth Service/Dependency/task5_Dependency.md
Section: 6. Redis / Queue / External Services

TaskImplementation/Auth Service/task7_Dependency.md
Section: 6. Redis / Queue / External Services
```

---

## 7. Environment Variables

Task 8 introduces **no new or changed `.env` variables**.

### For focused security tests

No `.env` file is required for these commands:

```bash
cd backend/services/auth-service
go test ./internal/security/token
go test ./internal/usecase -run 'RefreshToken|Logout|VerifyOTP|AssignRole|GetUserRoles|RequireFreshRole'
go test ./internal/authorization ./internal/transport/http -run 'Bypass|Permission|AuthRequired|RequireRoles|UsecaseError'
```

Why:

- Token tests generate RSA keys in memory.
- Refresh token tests use `test-refresh-pepper` as a test-only constant.
- OTP tests use fake notification/rate-limit repositories.
- HTTP tests use fake usecases and `httptest`.
- Role tests use in-memory context/role fixtures.

### Optional shell-only variables for local test comfort

These are not project `.env` values. Use them only in your terminal if needed.

| Variable | Required? | Example | Purpose | Security note |
|---|---:|---|---|---|
| `GOCACHE` | Optional | `/tmp/auth-go-cache` | Use writable Go build cache path | Not a secret |
| `GOFLAGS` | Optional | `-count=1` | Force fresh test run when debugging | Not a secret |

Example:

```bash
cd backend/services/auth-service
GOCACHE=/tmp/auth-go-cache GOFLAGS=-count=1 go test ./internal/security/token
```

### Full service `.env`

For full server startup, reuse previous docs. Do not duplicate the full `.env` here.

```text
TaskImplementation/Auth Service/Dependency/task1_Dependency.md
Section: 7. Environment Variables

TaskImplementation/Auth Service/Dependency/task4_Dependency.md
JWT and refresh token variables

TaskImplementation/Auth Service/Dependency/task5_Dependency.md
OTP and Notification variables

TaskImplementation/Auth Service/Dependency/task6_Dependency.md
RBAC role reason variable

TaskImplementation/Auth Service/task7_Dependency.md
Session link and outbox variables
```

Important beginner note:

```text
Task 8 focused tests should not depend on production secrets.
If a test needs JWT keys from .env, it is probably not isolated enough.
```

---

## 8. Docker Setup

Task 8 introduces **no new Docker setup**.

### Current Docker impact

| Docker item | Task 8 focused tests need it? | Status |
|---|---:|---|
| App Dockerfile | No | No new Dockerfile added by Task 8 |
| Docker Compose | No | Existing dependency examples are reused |
| MySQL container | No for focused tests | Required only for full service/manual API flow |
| Redis container | No for focused tests | Required only for full service/manual OTP flow |
| Kafka/RabbitMQ container | No | Not used by current Auth Service |
| New volume/network | No | No Task 8 change |

For baseline Docker dependency setup, refer:

```text
TaskImplementation/Auth Service/Dependency/task1_Dependency.md
Section: 8. Docker Setup
```

For session-link/outbox runtime details, refer:

```text
TaskImplementation/Auth Service/task7_Dependency.md
Section: 8. Docker Setup
```

---

## 9. Local Development Setup

### Step 1: Read previous dependency documentation

Read these first so base setup is clear:

```text
TaskImplementation/Auth Service/Dependency/task1_Dependency.md
TaskImplementation/Auth Service/Dependency/task4_Dependency.md
TaskImplementation/Auth Service/Dependency/task5_Dependency.md
TaskImplementation/Auth Service/Dependency/task6_Dependency.md
TaskImplementation/Auth Service/task7_Dependency.md
```

### Step 2: Go to project directory

For Task 8 tests, work inside the Auth Service module:

```bash
cd backend/services/auth-service
```

### Step 3: Install only new dependencies if any

Task 8 has no new external dependency.

If modules are not downloaded:

```bash
go mod download
```

Do not run `go get` unless a new import is actually added to code.

### Step 4: Setup only new databases/services if any

No new database or external service is required for focused Task 8 tests.

For full manual API verification, follow previous MySQL/Redis/JWT/OTP/session-link setup docs.

### Step 5: Add only new or changed environment variables

No new Task 8 `.env` variable exists.

For focused tests, keep the environment minimal:

```bash
go test ./internal/security/token
```

### Step 6: Run migrations if needed

No new Task 8 migration exists.

Only run migrations if you are starting the full service or manually testing real HTTP APIs.

### Step 7: Start backend service

Focused security tests do not require starting the backend server.

For full server startup, reuse:

```text
TaskImplementation/Auth Service/Dependency/task1_Dependency.md
Section: 10. Running the Project
```

### Step 8: Verify functionality related to `task8.md`

Recommended focused verification order:

```bash
cd backend/services/auth-service

go test ./internal/security/token
go test ./internal/usecase -run 'RefreshToken|Logout|VerifyOTP|AssignRole|GetUserRoles|RequireFreshRole'
go test ./internal/authorization ./internal/transport/http -run 'Bypass|Permission|AuthRequired|RequireRoles|UsecaseError'
CGO_ENABLED=1 go test -race ./internal/usecase ./internal/authorization ./internal/transport/http
go test ./...
```

Expected result:

```text
All selected packages should pass.
Expired/invalid JWTs should reject.
Refresh token reuse should reject and revoke family.
OTP replay should reject.
Wrong role should receive forbidden/permission denied behavior.
Logs/responses should not leak raw token, OTP, or Authorization header.
```

---

## 10. Running the Project

Task 8 is test-focused, so the normal action is:

```bash
cd backend/services/auth-service
go test ./...
```

Starting the HTTP server is optional for Task 8.

### If you still want to start the full server

Follow previous dependency docs for:

- MySQL startup and migrations
- Redis startup
- JWT RSA key files
- OTP peppers and Notification endpoint
- Session link mode and event pepper

Then run the service using the command already documented in Task 1 dependency guide.

### Task 8 related routes for manual checks

Manual checks are optional because automated tests are the source of truth for Task 8.

| Route | Method | Security behavior to verify |
|---|---|---|
| `/api/v1/auth/refresh` | `POST` | Reused/expired refresh token returns generic invalid refresh response |
| `/api/v1/auth/logout` | `POST` | Revoked token cannot refresh later |
| `/api/v1/auth/otp/verify` | `POST` | Already used OTP returns generic invalid OTP response |
| `/internal/v1/auth/tokens/verify` | `POST` | Expired/forged/wrong-audience access token rejects |
| `/internal/v1/auth/roles` | `GET` | Missing/invalid bearer token rejects |
| `/internal/v1/auth/roles/assign` | `POST` | Wrong role cannot mutate roles |
| `/.well-known/jwks.json` | `GET` | Public key endpoint available for JWT verification |
| `/healthz` | `GET` | Basic service health check |

### Ports and networking

Task 8 introduces no new port.

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Auth Service HTTP | `8081` by default through `AUTH_HTTP_ADDR=:8081` | Auth APIs, JWKS, health | Reused |
| MySQL | `3306` | Full service auth database | Reused, not needed for focused tests |
| Redis | `6379` | Full service OTP rate limits | Reused, not needed for focused tests |
| Notification Service | `8084` example URL | Full OTP delivery | Reused, not needed for focused tests |
| Session Service / event ingress | Config dependent | Session-link event consumer | Reused from Task 7 |

Port conflict guidance is already covered in previous dependency docs. Task 8 does not change ports.

---

## 11. Common Errors & Fixes

General Go, MySQL, Redis, Docker, JWT key, OTP, RBAC, and session-link setup errors are already documented in previous dependency files. This section lists Task 8 specific or test-focused issues only.

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `go: go.mod requires go >= 1.26.3` | Local Go version is older than module requirement | Install compatible Go version or use the repo-approved toolchain | Check `go version` before running tests |
| `pattern ./...: directory prefix . does not contain main module` | Command run from wrong folder | Run `cd backend/services/auth-service` first | Always run module tests from service module root |
| Token expiry test fails unexpectedly | Clock skew or test clock changed | Check `JWT_CLOCK_SKEW` logic and fixed clock in token tests | Keep expiry tests deterministic, do not rely on real waiting |
| Forged JWT test passes incorrectly | Verifier is accepting bad signature/header | Inspect `internal/security/token/verifier.go` | Always assert `ErrInvalidAccessToken` for forged/unknown `kid` cases |
| Refresh concurrent reuse test is flaky | Fake repo locking or transaction-like behavior changed | Run `go test -race ./internal/usecase`; inspect fake repository mutex behavior | Keep refresh rotation atomic in tests and real repo |
| OTP replay test accepts second verify | `verified_at` or replay guard is not checked | Inspect `VerifyOTP` usecase and fake OTP repo state transitions | Test first success and second failure in the same scenario |
| Role bypass test allows wrong role | Role hierarchy or permission mapping changed | Inspect `internal/authorization/rbac.go` and `permission.go` | Keep role matrix table tests updated with every new role |
| HTTP security test leaks `Authorization` or raw token in logs | Logging code added sensitive headers/inputs | Remove secret fields from logs; log trace/request ID only | Add redaction assertion for every new auth failure log |
| `go test -race` is slow | Race detector adds overhead | Run it on focused packages first | Use `go test` for quick loop, `go test -race` before merge |
| `go: -race requires cgo` or `gcc not found` | Race detector needs cgo and a C compiler | Install a C compiler such as `gcc` and run with `CGO_ENABLED=1`, or run normal `go test` on machines without compiler tools | Keep CI runner/toolchain configured with cgo and compiler support |

---

## 12. Security & Best Practices

### Task 8 security best practices

| Practice | Why it matters |
|---|---|
| Keep tests deterministic | Security tests should not fail randomly because of real time, external services, or network |
| Generate JWT test keys in memory | Production PEM files should not be needed in unit tests |
| Store only token/OTP hashes in fake assertions | Tests should prove plain secrets are not persisted |
| Assert generic API errors | Responses should not reveal "reused", "already used", raw role internals, or secret values |
| Assert log redaction | Logs must not contain raw JWT, refresh token, OTP, password, Authorization header, IP, or device fingerprint |
| Keep concurrency tests | Refresh reuse and OTP replay are race-prone; one parallel request should not create two successes |
| Use table-driven role matrices | Buyer/seller/admin/superadmin boundaries are easier to audit in table form |
| Run `go test -race` before merging auth changes | Race detector helps catch unsafe fake or real repo behavior |

### What not to do

- Do not add real production secrets to test files.
- Do not make tests depend on `.env` unless it is an integration test with explicit reason.
- Do not log raw `Authorization`, refresh token, OTP, password, peppers, or private keys.
- Do not install new modules only because a design doc mentions them.
- Do not weaken error assertions to "any error" for high-risk security tests.

---

## 13. Missing or Misconfigured Things

This audit is based on the current repository state and Task 8 scope.

| Item | Status | Risk | Suggested fix |
|---|---|---|---|
| Dedicated security test command in CI | Not visible in current repo | Security tests may be skipped if CI only runs narrow packages | Add CI step for `go test ./...` and focused `CGO_ENABLED=1 go test -race ./internal/usecase ./internal/authorization ./internal/transport/http` |
| Real MySQL transaction integration test for refresh reuse | Focused tests use fakes | Fake can pass while real `SELECT ... FOR UPDATE` behavior regresses | Add DB integration test later when test DB harness exists |
| Real MySQL transaction integration test for OTP replay | Focused tests use fakes | DB locking bug may not be caught by unit fakes | Add integration test around `otp_challenges` row lock |
| Docker Compose in repo | Still not present for current service dependencies | Beginner setup depends on manual commands/docs | Keep referencing Task 1 docs now; add compose later for repeatable local dependencies |
| Broker mode for session link | Current config supports `outbox` and `disabled`, not direct Kafka/RabbitMQ | Design docs may confuse beginners | Follow `task7_Dependency.md`; do not install broker clients for Task 8 |
| Test tags for slow/security tests | Not separated | `go test ./...` may grow slow over time | Add tags only if suite becomes heavy; current focused tests can run normally |

No new hardcoded production credential was introduced by Task 8 tests. Test constants like `test-refresh-pepper` are test-only and should not be copied into `.env`.

---

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/Auth Service/Dependency/task1_Dependency.md` | Required Software, Go modules, MySQL, Redis, Docker, full local run | Baseline Auth Service setup already documented |
| `TaskImplementation/Auth Service/Dependency/task2_Dependency.md` | Base auth tables and account schema | Task 8 uses existing account/auth schema concepts only |
| `TaskImplementation/Auth Service/Dependency/task3_Dependency.md` | Password dependency and credential setup | Login/security tests depend on existing credential behavior but add no new password setup |
| `TaskImplementation/Auth Service/Dependency/task4_Dependency.md` | JWT RS256, refresh token rotation, JWT env variables | Task 8 tests token expiry and refresh reuse on existing token setup |
| `TaskImplementation/Auth Service/Dependency/task5_Dependency.md` | OTP DB, Redis rate limit, Notification endpoint, OTP env variables | Task 8 tests OTP replay on existing OTP setup |
| `TaskImplementation/Auth Service/Dependency/task6_Dependency.md` | RBAC roles, permissions, role assignment setup | Task 8 tests role bypass on existing RBAC setup |
| `TaskImplementation/Auth Service/task7_Dependency.md` | Session link, outbox, reuse/logout event setup | Task 8 checks refresh reuse/logout event behavior using existing session-link abstraction |

---

## 15. Final Checklist

### Task 8 focused test checklist

- [ ] Previous dependency documentation checked.
- [ ] Confirmed Task 8 adds no new runtime dependency.
- [ ] Confirmed no new `.env` variable is needed for focused tests.
- [ ] Confirmed no new migration is needed for Task 8.
- [ ] Go version is compatible with `backend/services/auth-service/go.mod`.
- [ ] Dependencies downloaded with `go mod download` if needed.
- [ ] JWT security tests pass:

```bash
cd backend/services/auth-service
go test ./internal/security/token
```

- [ ] Refresh reuse, OTP replay, and role usecase tests pass:

```bash
cd backend/services/auth-service
go test ./internal/usecase -run 'RefreshToken|Logout|VerifyOTP|AssignRole|GetUserRoles|RequireFreshRole'
```

- [ ] RBAC and HTTP security tests pass:

```bash
cd backend/services/auth-service
go test ./internal/authorization ./internal/transport/http -run 'Bypass|Permission|AuthRequired|RequireRoles|UsecaseError'
```

- [ ] Race detector run completed for concurrency-sensitive packages, if local Go toolchain has cgo and `gcc`/C compiler support:

```bash
cd backend/services/auth-service
CGO_ENABLED=1 go test -race ./internal/usecase ./internal/authorization ./internal/transport/http
```

- [ ] Full Auth Service test suite passes:

```bash
cd backend/services/auth-service
go test ./...
```

- [ ] Logs checked for no raw JWT, refresh token, OTP, Authorization header, IP, or device fingerprint leakage.
- [ ] No duplicate setup documentation added.
- [ ] New documentation saved at `TaskImplementation/Auth Service/task8_Dependency.md`.
