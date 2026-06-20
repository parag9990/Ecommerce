# Environment Dependency - Superadmin Panel

## 1. What is this dependency?

Environment variables app ko runtime configuration dete hain. Frontend ke liye `VITE_*` variables build/dev time par inject hote hain. Backend/gateway ke liye `.env` files ports, database, Redis, gRPC, JWT, and timeouts configure karti hain.

## 2. Why this service uses it

Superadmin Panel ko API base URL chahiye. Gateway and Superadmin backend ko service addresses, auth config, database DSN, and operational limits chahiye.

## 3. Required or optional

Required. Frontend env example exists. Backend env files exist, but backend source/config loader clearly found nahi hua.

## 4. Where it is used in project

| Path | Use |
|------|-----|
| `frontend/superadmin-panel/.env.example` | Frontend env template |
| `frontend/superadmin-panel/src/lib/http.ts` | Reads `import.meta.env.VITE_API_BASE_URL` |
| `backend/services/superadmin-service/.env` | Backend env names |
| `backend/services/api-gateway/.env` | Gateway env names |

## 5. Installation steps

Frontend:

```bash
cd frontend/superadmin-panel
cp .env.example .env
```

Backend `.env.example` files were not found. Create backend examples with placeholder values only.

## 6. Docker setup, if possible

No compose file found. If Docker is added later, use `env_file` or config maps/secrets. Do not bake secrets into Docker images.

Suggested compose pattern:

```yaml
env_file:
  - ./backend/services/superadmin-service/.env
```

This is suggested only.

## 7. Local setup without Docker

Frontend:

```bash
cd frontend/superadmin-panel
cp .env.example .env
```

Backend suggested command based on common dotenv loading:

```bash
cd backend/services/superadmin-service
cp .env.example .env
```

Backend `.env.example` is missing, so this command is not currently runnable.

## 8. Required environment variables

### Frontend

| Variable | Required | Notes |
|----------|----------|-------|
| `VITE_API_BASE_URL` | Yes | Example points to `http://localhost:8080/api/v1` |
| `VITE_APP_NAME` | Optional | Display name |
| `VITE_ADMIN_SESSION_WARNING_MINUTES` | Optional | Warning timing |

### Superadmin backend

| Variable | Required | Notes |
|----------|----------|-------|
| `SERVICE_NAME` | Yes | Found as `superadmin-service` |
| `APP_ENV` | Yes | local/dev/prod style environment |
| `HTTP_ADDR` | Yes | Found env only, `:8088` |
| `LOG_LEVEL` | Optional | Found env only |
| `SUPERADMIN_DATABASE_DSN` | Required if DB enabled | Do not commit real credentials |
| `SUPERADMIN_REQUIRE_DATABASE` | Optional toggle | Found env only |
| `USER_SERVICE_ADMIN_BASE_URL` | Required for user workflows | Backend source not found |
| `ORDER_SERVICE_ADMIN_BASE_URL` | Required for order workflows | Backend source not found |
| `PAYMENT_SERVICE_ADMIN_BASE_URL` | Required for payment workflows | Backend source not found |
| `SUPERADMIN_PLATFORM_SETTINGS_CACHE_TTL` | Optional | Settings cache TTL |
| `SUPERADMIN_SESSION_ANALYTICS_MAX_RANGE` | Optional | Analytics query guardrail |

### API Gateway

| Variable | Required | Notes |
|----------|----------|-------|
| `HTTP_ADDR` | Yes | Found `:8080` |
| `API_BASE_PATH` | Yes | Found `/api/v1` |
| `SUPERADMIN_GRPC_ADDR` | Yes | Found `localhost:50062` |
| `REDIS_ADDR` | Required if rate limit enabled | Found `localhost:6379` |
| `JWT_ISSUER`, `JWT_AUDIENCE`, `JWT_JWKS_URL` | Yes | Admin auth verification |
| `REQUEST_VALIDATION_ENABLED` | Recommended | Input validation |

## 9. Start commands

Frontend:

```bash
cd frontend
pnpm --filter superadmin-panel dev
```

Backend/gateway commands are suggested only because source code is missing:

```bash
cd backend/services/api-gateway
go run ./cmd/api-gateway
```

## 10. Verify running commands

Frontend env verify:

```bash
cd frontend
pnpm run superadmin:typecheck
```

Gateway env verify:

```bash
curl http://localhost:8080/api/v1/admin/settings
```

Without auth, a clean 401/403 is better than connection refused.

## 11. Common errors and fixes

| Error | Reason | Fix |
|-------|--------|-----|
| `VITE_API_BASE_URL is not configured` | `.env` missing | Copy frontend `.env.example` |
| Frontend calls wrong URL | URL has duplicate `/api/v1` or missing base path | Set exact gateway base URL |
| Backend starts without DB | `SUPERADMIN_REQUIRE_DATABASE=false` may allow degraded mode | Set required flags true in production |
| JWT validation fails | JWKS/issuer/audience mismatch | Align gateway JWT env with Auth Service |

## 12. Security notes

- Never commit real DSNs, passwords, JWT private keys, provider secrets, or Redis passwords.
- Use `.env.example` with placeholders.
- Production secrets should come from secret manager/Kubernetes secrets.
- Admin service should fail fast if required production env vars are missing.

## 13. Final checklist

- [x] Frontend env example exists.
- [x] Frontend reads `VITE_API_BASE_URL`.
- [x] Gateway env file exists.
- [x] Superadmin backend env file exists.
- [ ] Backend `.env.example` found.
- [ ] Backend config loader found.

