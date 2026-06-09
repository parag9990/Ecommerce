# Project Dependency & Setup Guide

## 1. Project Overview

This guide is generated for these reusable variables:

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `User Service` |
| `TASK_FILE_NAME` | `task6.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task6_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

Current task ka goal validation setup samjhana hai: email, phone, address, GSTIN, seller profile details, path IDs, and safe text rules. Bad input Gateway par REST request me reject hota hai, aur service usecase layer me bhi reject hota hai. Isse DB ko garbage data se bachaya jaata hai.

Important boundary:

- Business logic rewrite nahi karna hai.
- Original `INPUT_FILE_PATH` modify nahi karna hai.
- Database schema, Docker setup, and base Go/MySQL setup repeat nahi karna hai.
- Sirf new or task-specific dependency, environment, setup, DevOps, and troubleshooting notes document karne hain.

Analyzed files:

| File | Why checked |
|---|---|
| `INPUT_FILE_PATH` | Current task scope: validation for profile, address, GSTIN, seller details |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Base Go, MySQL, Docker, `.env`, ports, gRPC tools |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | MySQL schema and migration setup |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | Repository runtime assumptions |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | gRPC transport, proto, reflection, current-code caveats |
| `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md` | API Gateway REST setup, JWT setup, Gateway validation env |
| `backend/shared/validation/` | Actual shared validation implementation |
| `backend/services/api-gateway/internal/handlers/user_validation.go` | Gateway-side REST validation |
| `backend/services/user-service/internal/validation/user_validator.go` | Service-side usecase validation |
| `backend/services/api-gateway/internal/config/config.go` | Gateway validation phone-region env loading |
| `backend/services/user-service/internal/config/config.go` | Service validation phone-region env loading |

## 2. Tech Stack

Most base technologies are already explained in previous dependency files. Follow references instead of duplicating setup.

| Technology | Required? | Status for current task | Beginner-friendly explanation |
|---|---:|---|---|
| Go `1.24` | Yes | Reused | Go backend language hai. Full Go setup already previous docs me hai. |
| Go Modules and `backend/go.work` | Yes | Reused | Multiple local modules ko connect karta hai: Gateway, current service, generated proto, shared validation. |
| API Gateway with Go `net/http` | Yes for REST validation | Reused from Task 5 | Public REST request pe validation run hoti hai before gRPC call. |
| gRPC and Protobuf | Yes | Reused | Gateway current service ko typed internal gRPC calls bhejta hai. Invalid service-side validation `InvalidArgument` banta hai. |
| MySQL | Yes for real runtime | Reused | Profile, address, seller profile data same `user_db` tables me store hota hai. |
| Shared validation module | Yes | Task-specific runtime surface | `backend/shared/validation` common rules provide karta hai: email, phone, GSTIN, postal code, HTTPS URL, safe text. |
| `github.com/nyaruka/phonenumbers` | Yes via shared validation | Reused plus service-side usage | Phone ko country-aware parse karke E.164 format me normalize karta hai, e.g. `99999 99999` to `+919999999999`. |
| Go `net/mail`, `net/url`, `regexp`, `unicode` | Yes | Built-in | Email, HTTPS URL, GSTIN/postal regex, and unsafe character checks ke liye standard library use hoti hai. |
| `google.golang.org/grpc/status` and `codes` | Yes | Reused | Validation errors ko gRPC `InvalidArgument` me map karne ke liye. |

Already explained setup:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`1. Project Tech Stack Analysis`
`2. Language-Specific Dependency System: Go`
`5. External Services Analysis`
```

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task5_Dependency.md`

Sections:
`2. Tech Stack`
`4. Dependency Management`
`7. Environment Variables`
```

Important actual-code note:

- `INPUT_FILE_PATH` mentions `github.com/go-playground/validator/v10` as a possible implementation tool.
- Current repository code does not import or require `go-playground/validator/v10`.
- Current implementation uses custom validation helpers in `backend/shared/validation`.
- Beginner rule: do not install `go-playground/validator/v10` unless code is changed to use it.

## 3. Required Software

No new software installation is required if previous dependency guides are already followed.

| Software | Needed? | Setup status |
|---|---:|---|
| Go `1.24+` | Yes | Reuse `task1_Dependency.md` |
| MySQL 8.x | Yes for real runtime | Reuse `task1_Dependency.md` and `task2_Dependency.md` |
| MySQL CLI | Recommended | Reuse `task2_Dependency.md` |
| Docker | Optional | Reuse MySQL Docker setup from `task1_Dependency.md` |
| grpcurl | Optional | Reuse `task1_Dependency.md` and `task4_Dependency.md` |
| curl/Postman | Yes for REST validation checks | Reuse Task 5 style verification |
| Redis | No for current task | Not used by current validation implementation |
| Kafka/RabbitMQ | No for current task | Only later outbox worker paths need them |

## 4. Dependency Management

Go dependency basics are already documented:

```md
This setup is already explained in:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`2. Language-Specific Dependency System: Go`
```

Current task-specific module reality:

| Module file | Current task relevance |
|---|---|
| `backend/shared/validation/go.mod` | Directly requires `github.com/nyaruka/phonenumbers v1.8.0`. |
| `backend/services/api-gateway/go.mod` | Requires local `github.com/parag/ecommerce/backend/shared/validation`. |
| `backend/services/user-service/go.mod` | Requires local `github.com/parag/ecommerce/backend/shared/validation`. |
| `backend/go.work` | Includes `./shared/validation`, `./services/api-gateway`, and `./services/user-service`. |

Install/download commands for a fresh clone:

```bash
cd backend/shared/validation
go mod download
```

```bash
cd backend/services/user-service
go mod download
```

```bash
cd backend/services/api-gateway
go mod download
```

Run tests related to validation:

```bash
cd backend/shared/validation
go test ./...
```

```bash
cd backend/services/user-service
go test ./internal/validation ./internal/usecase ./internal/transport/grpc
```

```bash
cd backend/services/api-gateway
go test ./internal/handlers
```

Only run `go mod tidy` if module errors appear or dependency files were changed:

```bash
cd backend/shared/validation
go mod tidy
```

Beginner note:

- `go mod download` dependencies download karta hai.
- `go mod tidy` unused dependencies remove bhi kar sakta hai, isliye random tidy avoid karo if you are only running the project.
- `replace` entries local shared modules ko point karte hain. Inko remove kar doge to local imports break ho sakte hain.

## 5. Database Setup

Current task does not introduce a new database, table, index, migration, or credential.

Reuse MySQL setup:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`3. Database Analysis`
`3. Database Analysis -> D. Local Installation`
`3. Database Analysis -> E. Docker Setup`
`3. Database Analysis -> J. Connection String Format`
`3. Database Analysis -> K. Where To Place Credentials`
```

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`

Sections:
`5. Database Setup`
`Apply Migration`
`Schema Verification Queries`
```

How validation uses DB indirectly:

| Data area | DB table | Validation impact |
|---|---|---|
| User email/full name/phone/avatar | `users` | Invalid email, phone, unsafe names, and non-HTTPS avatar URLs reject before repository write. |
| Address book | `user_addresses` | Required address fields, postal code, phone, country, and safe text reject before write. |
| Seller profile | `seller_profiles` | Store name, display name, GSTIN, and support email validate before write. |

Current task migration status:

| Item | Status |
|---|---|
| New migration for current task | No |
| New DB port | No |
| New DB credential | No |
| New DB container | No |

Current repository caveat:

- The present codebase also contains later audit and outbox code.
- For running the current full service code, apply migrations in numeric order from `backend/services/user-service/migrations/`.
- If you are only studying or unit-testing validation packages, MySQL is not required.

## 6. Redis / Queue / External Services

Current task does not add Redis, Kafka, RabbitMQ, NATS, MinIO, Elasticsearch, SMTP, Stripe, Firebase, OAuth, Kubernetes, or any third-party credential.

| Service | Required for current task? | Notes |
|---|---:|---|
| Current service gRPC | Yes | Gateway calls current service at `localhost:50052` by default. |
| API Gateway HTTP | Yes for REST validation | REST APIs expose validation behavior to frontend clients. |
| Auth/JWT verifier | Yes for protected Gateway APIs | Setup already documented in `task5_Dependency.md`. |
| MySQL | Yes for real profile/address/seller runtime | Reused from earlier tasks. |
| Docker | Optional | Useful for reused MySQL container only. |
| Redis | No | Current validation code does not use Redis. |
| RabbitMQ/Kafka | No for current task | Current service has later event/outbox code, but brokers are needed only when outbox worker is enabled. |

Reuse external-service explanations:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`5. External Services Analysis`
`6. Ports and Networking`
`7. Docker and DevOps Setup`
```

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task4_Dependency.md`

Sections:
`6. Redis / Queue / External Services`
`10. Running the Project`
```

## 7. Environment Variables

Base `.env` setup is already documented:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`4. Environment Variables`
```

Gateway `.env` setup is already documented:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task5_Dependency.md`

Section:
`7. Environment Variables`
```

Current task-specific new variable:

```bash
# backend/services/user-service/.env
USER_SERVICE_VALIDATION_PHONE_REGION=IN
```

Gateway phone-region setup is already explained in:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task5_Dependency.md`

Section:
`7. Environment Variables`

Variable:
`API_GATEWAY_VALIDATION_PHONE_REGION`
```

For current task, just make sure Gateway and current service use the same region value.

Variable details:

| Variable | Required? | Example | Where | Purpose | Security note |
|---|---:|---|---|---|---|
| `USER_SERVICE_VALIDATION_PHONE_REGION` | No, default is `IN` | `IN` | `backend/services/user-service/.env` | Service-side phone parsing default region. | Not secret. Keep it non-empty and intentional. |

How `.env` is loaded:

- Current code reads environment variables using `os.Getenv`.
- It does not automatically load `.env` files.
- Source the correct file before running each process.

For current service:

```bash
cd backend/services/user-service
set -a
. ./.env
set +a
go run ./cmd/server
```

For API Gateway:

```bash
cd backend/services/api-gateway
set -a
. ./.env
set +a
go run ./cmd/server
```

Common `.env` mistakes for current task:

| Mistake | Result | Fix |
|---|---|---|
| Service and Gateway use different phone regions | Phone may pass one layer and fail another | Set both validation region variables to same value, usually `IN`. |
| Variable set to blank in deployment | Config load can fail or fallback may not behave as expected | Use explicit non-empty region code. |
| Same terminal sources both `.env` files | Gateway target/listen variables can overwrite each other | Use separate terminals. |
| `.env` edited but process not restarted | Old region still used | Restart service after env changes. |

## 8. Docker Setup

No new Dockerfile, docker-compose service, container, volume, network, restart policy, or health check is introduced by current task.

Reuse Docker setup:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`7. Docker and DevOps Setup`
```

If containerizing current task later, add only this new validation-specific current service value:

```yaml
environment:
  USER_SERVICE_VALIDATION_PHONE_REGION: IN
```

For Gateway container env, reuse `API_GATEWAY_VALIDATION_PHONE_REGION` from `task5_Dependency.md` and keep it aligned with the value above.

Container networking reminder:

- Inside Docker, Gateway should call current service using service DNS like `user-service:50052`, not `localhost:50052`.
- MySQL should remain private to current service.
- Gateway HTTP port can be exposed to host or ingress.
- Current task does not require exposing gRPC publicly.

## 9. Local Development Setup

### Step 1: Read previous dependency documentation

Follow these first:

| Order | File | Why |
|---:|---|---|
| 1 | `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Base Go, MySQL, Docker, `.env`, ports, gRPC tools |
| 2 | `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | MySQL schema and migration setup |
| 3 | `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | Repository assumptions and DB verification |
| 4 | `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | gRPC startup and grpcurl verification |
| 5 | `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md` | API Gateway REST, JWT, and Gateway env setup |

### Step 2: Go to project directory

```bash
cd /home/parag/Ecommerce
```

### Step 3: Install only current task dependencies

No new manual `go get` is needed for current task. Download existing module dependencies:

```bash
cd backend/shared/validation
go mod download
```

```bash
cd backend/services/user-service
go mod download
```

```bash
cd backend/services/api-gateway
go mod download
```

### Step 4: Setup only reused databases/services

No new database or external service setup for current task.

Do this from previous docs:

- Start MySQL if running real APIs.
- Apply existing migrations.
- Start current service gRPC on `:50052`.
- Start API Gateway on `:8080`.

### Step 5: Add only new or changed environment variables

Add explicit current service validation region:

```bash
# backend/services/user-service/.env
USER_SERVICE_VALIDATION_PHONE_REGION=IN
```

Gateway already has `API_GATEWAY_VALIDATION_PHONE_REGION` documented in `task5_Dependency.md`. Verify it uses the same value, usually `IN`.

Also ensure Gateway has the JWT config already described in `task5_Dependency.md`. Current Gateway config requires an HS256 secret through `API_GATEWAY_JWT_HS256_SECRET` or `JWT_HS256_SECRET`.

### Step 6: Run migrations if needed

Current task adds no migration.

For a fresh local runtime, apply existing migrations from:

```text
backend/services/user-service/migrations/
```

Follow earlier instructions:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`

Section:
`5. Database Setup -> Apply Migration`
```

### Step 7: Start backend services

Terminal 1: start current service.

```bash
cd backend/services/user-service
set -a
. ./.env
set +a
go run ./cmd/server
```

Expected log should include:

```text
user_service_grpc_listening
```

Terminal 2: start API Gateway.

```bash
cd backend/services/api-gateway
set -a
. ./.env
set +a
go run ./cmd/server
```

Expected log should include:

```text
api_gateway_http_listening
```

### Step 8: Verify APIs related to `TASK_FILE_NAME`

Use a valid buyer/seller JWT token from the Auth setup. Replace `$TOKEN` with a real token.

Invalid phone should fail at Gateway:

```bash
curl -i -X PATCH http://localhost:8080/api/v1/me \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"phone":"abc"}'
```

Expected:

```text
HTTP/1.1 400 Bad Request
VALIDATION_ERROR
```

Invalid Indian PIN code should fail:

```bash
curl -i -X POST http://localhost:8080/api/v1/me/addresses \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Aarav Sharma","line1":"221B MG Road","city":"Bengaluru","state":"Karnataka","postal_code":"012345","country":"IN"}'
```

Invalid GSTIN should fail for seller token:

```bash
curl -i -X PATCH http://localhost:8080/api/v1/sellers/me \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"gst_number":"99AAAAA0000A1Z5"}'
```

## 10. Running the Project

Ports and networking:

| Service | Port | Purpose | Status |
|---|---:|---|---|
| MySQL | `3306` | Stores profile/address/seller data | Reused |
| Current service gRPC | `50052` | Internal Gateway to service calls | Reused |
| API Gateway HTTP | `8080` | REST validation surface for frontend | Reused |
| Redis | `6379` | Not used by current validation code | Not required |
| RabbitMQ | `5672` | Later outbox worker only | Not required for current task |
| Kafka | `9092` | Later outbox worker only | Not required for current task |

How to change ports:

| Need | Variable |
|---|---|
| Change current service gRPC listen address | `USER_SERVICE_GRPC_ADDRESS` in `backend/services/user-service/.env` |
| Change Gateway upstream target | `USER_SERVICE_GRPC_ADDRESS` or `USER_GRPC_ADDR` in `backend/services/api-gateway/.env` |
| Change Gateway HTTP listen address | `API_GATEWAY_HTTP_ADDRESS` or `HTTP_ADDR` |
| Change MySQL host/port | `USER_SERVICE_DATABASE_DSN` |

Important beginner note:

`USER_SERVICE_GRPC_ADDRESS` has two meanings depending on process:

- In current service `.env`, `:50052` means "listen on this port".
- In Gateway `.env`, `localhost:50052` means "dial current service here".

Do not source both `.env` files in the same terminal.

## 11. Common Errors & Fixes

Generic Go/MySQL/Docker errors are already documented in previous dependency files. New validation-specific issues:

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `VALIDATION_ERROR` for `phone` | Phone cannot be parsed for configured region | Use E.164 like `+919999999999`, or set phone region to `IN` for Indian local numbers | Keep Gateway and service region env same |
| Phone passes Gateway but fails service | `API_GATEWAY_VALIDATION_PHONE_REGION` and `USER_SERVICE_VALIDATION_PHONE_REGION` differ | Align both values | Put both in deployment config together |
| `postal_code must be valid for country` | Postal code does not match country rule | For `IN`, use six digits without leading zero, e.g. `560001` | Normalize country to `IN` and validate form before submit |
| `gst_number must be a valid GSTIN` | Bad GSTIN length, state code, or pattern | Use 15-character GSTIN with valid state code | Remember current validation is format-level, not government API verification |
| `avatar_url must be a valid HTTPS URL` | URL is missing host or uses `http`, `javascript`, or `data` | Use `https://example.com/avatar.png` | Store only trusted HTTPS URLs |
| `API_GATEWAY_JWT_HS256_SECRET is required` | Gateway env has no HS256 secret | Add `API_GATEWAY_JWT_HS256_SECRET` or `JWT_HS256_SECRET` | Follow `task5_Dependency.md` Gateway env setup |
| `.env` values ignored | Code does not auto-load `.env` | Source `.env` before `go run` | Use separate terminal startup scripts |
| `dial user service at localhost:50052` | Current service is not running or Gateway target is wrong | Start current service and verify port `50052` | Start services in order: DB, current service, Gateway |
| Module error for shared validation | Running outside workspace or local `replace` paths broken | Run from correct module and keep `backend/go.work` intact | Do not delete local shared module replace entries |

## 12. Security & Best Practices

Task-specific best practices:

- Gateway validation and service usecase validation dono rakho. Gateway fast feedback deta hai, service layer internal gRPC callers ke against safety deta hai.
- Store phone numbers in canonical E.164 format. Same phone ko multiple formats me store karna future duplicate/search issues create karega.
- Keep `API_GATEWAY_VALIDATION_PHONE_REGION` and `USER_SERVICE_VALIDATION_PHONE_REGION` aligned.
- Do not log raw JWT tokens, phone numbers, email addresses, or full addresses in debug logs.
- Do not add unused validation libraries. Unused dependencies supply-chain risk badhate hain.
- GSTIN validation current code me format-level hai. Real GST registration active hai ya nahi, uske liye future external GST verification integration chahiye.
- DB constraints last safety net hain. User-friendly validation error application layer se aana chahiye.
- Keep field-level validation messages stable enough for frontend error rendering.
- Reject unsafe text characters like control characters and `<` / `>` before storing profile and address display fields.

## 13. Missing or Misconfigured Things

Observed setup/config audit:

| Item | Status | Impact | Suggested fix |
|---|---|---|---|
| `go-playground/validator/v10` in `INPUT_FILE_PATH` | Mentioned in guide, not used in current code | Beginner may install unnecessary dependency | Do not install unless implementation is changed to use struct-tag validation. |
| `USER_SERVICE_VALIDATION_PHONE_REGION` missing from local `.env` | Default `IN` exists | Service still runs, but config is implicit | Add it explicitly for clarity. |
| `API_GATEWAY_VALIDATION_PHONE_REGION` missing from local Gateway `.env` | Default `IN` exists | Gateway still runs if other required env exists | Add it explicitly and keep aligned with service. |
| Gateway `.env` has JWKS-style values while current config expects HS256 secret | Gateway can fail on startup if no HS256 secret exists | Protected REST APIs cannot run | Add `API_GATEWAY_JWT_HS256_SECRET` or update Gateway auth config to use JWKS intentionally. |
| `.env` auto-loader not present | `go run` does not read `.env` automatically | Env changes appear ignored | Source `.env` manually or add a documented local startup script. |
| Docker Compose not checked in | No one-command local stack | Beginners must start MySQL/services manually | Add compose later if team wants repeatable onboarding. |
| Rich gRPC validation details not attached | Gateway gets generic service-side `validation failed` for internal invalids | Field-level errors are clearer at Gateway validation layer than service gRPC layer | Future enhancement: attach `google.rpc.BadRequest` details. |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Go setup, MySQL setup, Docker setup, base `.env`, ports, gRPC tools | Current task does not change base runtime setup. |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | MySQL schema, migration apply/rollback, schema verification | Current task uses same tables and adds no migration. |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | Repository DB assumptions and DB runtime checks | Validation passes clean input into existing repositories. |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | gRPC transport, proto generation, reflection, current repo caveat | Service-side validation errors map through gRPC. |
| `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md` | API Gateway REST setup, JWT config, Gateway env, curl verification, Gateway validation phone region | Current task validates the REST endpoints introduced in Task 5. |

## 15. Final Checklist

- [ ] Previous dependency documentation checked before using this guide
- [ ] No duplicate MySQL/Docker/Go setup copied into this file
- [ ] `backend/shared/validation` dependencies downloaded
- [ ] `backend/services/user-service` dependencies downloaded
- [ ] `backend/services/api-gateway` dependencies downloaded
- [ ] `USER_SERVICE_VALIDATION_PHONE_REGION` added or default intentionally accepted
- [ ] `API_GATEWAY_VALIDATION_PHONE_REGION` verified from `task5_Dependency.md`
- [ ] Gateway JWT secret configured as per `task5_Dependency.md`
- [ ] MySQL setup reused and migrations applied if running real APIs
- [ ] Current service started on gRPC port `50052`
- [ ] API Gateway started on HTTP port `8080`
- [ ] Invalid phone request returns `VALIDATION_ERROR`
- [ ] Invalid postal code request returns `VALIDATION_ERROR`
- [ ] Invalid GSTIN request returns `VALIDATION_ERROR`
- [ ] Logs checked for validation and startup errors
- [ ] No unused validation dependency installed
