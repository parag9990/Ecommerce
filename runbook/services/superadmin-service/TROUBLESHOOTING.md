# Superadmin Service Troubleshooting

## Common Error

- `/readyz` fails.
- Admin panel cannot load data.
- Downstream admin APIs return 401/403.

## Possible Cause

- Required downstream service is unavailable.
- MySQL migration missing.
- Admin token mismatch.

## Fix

- Start user, order, payment, and session services first.
- Run `docker compose up migrate-superadmin`.
- Match admin token env values across services.

## Verification Command

```powershell
curl http://localhost:8093/readyz
docker compose logs superadmin-service
```
