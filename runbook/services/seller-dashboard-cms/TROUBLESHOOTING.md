# Seller Dashboard CMS Troubleshooting

## Common Error

- Dashboard loads but data is empty.
- Login redirect fails.
- API calls return 401/403.

## Possible Cause

- Seller credentials or roles are missing.
- `VITE_LOGIN_URL` points at the wrong user app.
- CMS/product/order services are unhealthy.

## Fix

- Confirm seller seed credentials with the team.
- Set `VITE_LOGIN_URL` to the actual user app login URL.
- Verify API Gateway and CMS/product/order readiness.

## Verification Command

```powershell
curl http://localhost:8080/health/ready
curl http://localhost:8087/healthz
```
