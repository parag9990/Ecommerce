# Session Analytics Dashboard Troubleshooting

## Common Error

- Empty charts or failed API requests.
- Manual dev server port conflict.
- Admin endpoints return 401/403.

## Possible Cause

- Session service has no ingested data.
- Seller dashboard is already using port `5174`.
- Missing admin/superadmin auth.

## Fix

- Generate traffic in the user app or ingest session events.
- Run manual dashboard with `-- --port 5175`.
- Verify gateway/session readiness and admin credentials.

## Verification Command

```powershell
curl http://localhost:8086/healthz
curl http://localhost:8080/health/ready
```
