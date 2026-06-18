# Environment Dependency - Seller Dashboard (CMS)

## 1. What is this dependency?

Environment variables app configuration provide karte hain. Is project me frontend Vite vars, API Gateway vars, and CMS Service vars required hain.

## 2. Why this service uses it

Seller Dashboard ko different local/prod URLs, timeouts, DB credentials, Redis settings, JWT settings, and gRPC targets configure karne hote hain without code changes.

## 3. Required or optional

Required.

Some frontend env vars optional defaults ke saath kaam karte hain, but backend DB/JWT/gRPC vars required hain.

## 4. Where it is used in project

| Path | Use |
|---|---|
| `frontend/seller-dashboard/src/vite-env.d.ts` | Frontend env type declarations |
| `frontend/seller-dashboard/src/lib/http.ts` | `VITE_API_BASE_URL`, `VITE_API_TIMEOUT_MS` |
| `frontend/seller-dashboard/src/pages/login-redirect-page.tsx` | `VITE_LOGIN_URL` |
| `backend/services/api-gateway/.env` | Gateway runtime config |
| `backend/services/cms-service/.env` | CMS runtime config |

`.env.example` files were not clearly found.

## 5. Installation steps

No package installation needed for env files.

Suggested setup:

1. Create sanitized `.env.example` files.
2. Keep real `.env` local only.
3. Load env through service config loader.
4. Validate required env at startup.

## 6. Docker setup, if possible

Not clearly found in project files.

When Docker Compose is added, pass env through:

```yaml
env_file:
  - ./backend/services/cms-service/.env
```

For production, use secret manager/Kubernetes Secrets instead of committed env files.

## 7. Local setup without Docker

Frontend optional local env:

```text
VITE_API_BASE_URL=http://localhost:8080
VITE_API_TIMEOUT_MS=15000
VITE_LOGIN_URL=/login
```

CMS/Gateway env files already exist, but secret values should be reviewed and sanitized before sharing.

## 8. Required environment variables

### Frontend

| Variable | Required | Purpose |
|---|---|---|
| `VITE_API_BASE_URL` | Optional | Gateway base URL |
| `VITE_API_TIMEOUT_MS` | Optional | Request timeout |
| `VITE_LOGIN_URL` | Optional | Login redirect URL |

### CMS Service

| Variable | Required | Purpose |
|---|---|---|
| `CMS_HTTP_ADDR` | Yes | HTTP listen address |
| `CMS_GRPC_ADDR` | Yes | gRPC listen address |
| `CMS_ENV` | Yes | Runtime env |
| `CMS_MYSQL_DSN` | Optional | Full MySQL DSN |
| `CMS_DB_HOST` | Yes | MySQL host |
| `CMS_DB_PORT` | Yes | MySQL port |
| `CMS_DB_NAME` | Yes | MySQL database |
| `CMS_DB_USER` | Yes | MySQL username |
| `CMS_DB_PASSWORD` | Yes | MySQL password |
| `CMS_PRODUCT_SERVICE_BASE_URL` | Yes | Internal product boundary |
| `CMS_INTERNAL_AUTH_TOKEN` | Yes | Internal service token |

### API Gateway

| Variable | Required | Purpose |
|---|---|---|
| `HTTP_ADDR` | Yes | Gateway HTTP listen |
| `API_BASE_PATH` | Yes | REST base path |
| `API_CONTRACT_PATH` | Yes | API contract path |
| `JWT_JWKS_URL` | Yes | JWKS source |
| `PRODUCT_GRPC_ADDR` | Yes | Product service target |
| `ORDER_GRPC_ADDR` | Yes | Order service target |
| `CMS_GRPC_ADDR` | Yes | CMS service target |
| `USER_GRPC_ADDR` | Yes | User service target |
| `REDIS_ADDR` | If rate limit enabled | Redis target |

## 9. Start commands

Environment files do not start as services.

Backend suggested command after implementation:

```bash
cd backend/services/cms-service
go run ./cmd/server
```

Frontend:

```bash
cd frontend
pnpm --filter seller-dashboard dev
```

## 10. Verify running commands

Frontend env check is indirect:

```bash
cd frontend
pnpm --filter seller-dashboard typecheck
```

Backend env validation command was not clearly found. Suggested future behavior: service should fail fast if required env vars are missing.

## 11. Common errors and fixes

| Error | Reason | Fix |
|---|---|---|
| Frontend hits wrong API | `VITE_API_BASE_URL` missing/wrong | Set correct Gateway URL |
| Gateway cannot dial CMS | `CMS_GRPC_ADDR` mismatch | Align Gateway and CMS env |
| DB auth fails | Wrong `CMS_DB_*` | Fix DB credentials and permissions |
| Secret leaked | Real `.env` committed/shared | Rotate secret and add sanitized `.env.example` |
| Config silently defaults | Missing validation | Add startup config validation |

## 12. Security notes

- Do not commit production secrets.
- Use `.env.example` for names only, not real values.
- Vite `VITE_*` variables are public in browser bundle.
- Rotate any placeholder or exposed secrets before real deployment.
- Separate local, staging, and production env files.

## 13. Final checklist

- [x] Frontend env vars identified.
- [x] CMS env vars identified.
- [x] Gateway env vars identified.
- [x] CMS/Gateway port mismatch documented.
- [ ] `.env.example` not found.
- [ ] Backend config loader not clearly found.
- [ ] Backend config validation not clearly found.

