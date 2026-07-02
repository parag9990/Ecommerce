# Auth Service Troubleshooting

## Common Error

- Service fails at startup.
- OTP send fails.
- Gateway rejects issued tokens.

## Possible Cause

- `AUTH_MYSQL_DSN` or Redis settings are wrong.
- JWT key path is missing for manual runs.
- Notification gRPC service is down.
- JWT issuer/audience mismatch with gateway.

## Fix

- Start MySQL, Redis, notification service, and `migrate-auth`.
- Use Docker for local key generation behavior, or provide local key files for host runs.
- Match `JWT_ISSUER` and `JWT_AUDIENCE` with gateway env.

## Verification Command

```powershell
curl http://localhost:8081/readyz
curl http://localhost:8081/.well-known/jwks.json
```
