# Superadmin Panel Troubleshooting

## Common Error

- Login succeeds but pages fail to load.
- Admin API calls return 401/403.
- Manual Vite server uses an occupied port.

## Possible Cause

- Superadmin service or downstream admin APIs are unhealthy.
- Superadmin credentials or role claims are missing.
- CORS does not include the manual dev server origin.

## Fix

- Verify superadmin service readiness.
- Confirm superadmin seed credentials with the team.
- Start manual dev server on a unique port and update gateway CORS if needed.

## Verification Command

```powershell
curl http://localhost:8093/readyz
curl http://localhost:8080/health/ready
```
