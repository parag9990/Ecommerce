# API Gateway Troubleshooting

## Common Error

- `/health/ready` returns non-2xx.
- Browser requests fail with CORS or 401.
- Gateway logs show downstream dial failures.

## Possible Cause

- Backend service is unhealthy.
- Redis is down or password is wrong.
- JWT config differs from auth service.
- Frontend origin is not in `CORS_ALLOWED_ORIGINS`.

## Fix

- Start backend services before the gateway.
- Verify `REDIS_ADDR`, `REDIS_PASSWORD`, `JWT_JWKS_URL`, `JWT_ISSUER`, and `JWT_AUDIENCE`.
- Add manual frontend ports to `CORS_ALLOWED_ORIGINS`.
- Inspect `docker compose logs -f api-gateway`.

## Verification Command

```powershell
curl http://localhost:8080/health/ready
curl http://localhost:8081/.well-known/jwks.json
```
