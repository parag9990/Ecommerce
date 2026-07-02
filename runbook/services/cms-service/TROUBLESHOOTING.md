# CMS Service Troubleshooting

## Common Error

- `/healthz` fails.
- Coupon validation fails from cart.
- Product moderation calls fail.

## Possible Cause

- MySQL migration did not run.
- Product service is unavailable.
- Internal token or product token mismatch.

## Fix

- Run `docker compose up migrate-cms`.
- Start product service before CMS-dependent flows.
- Confirm `CMS_PRODUCT_SERVICE_AUTH_TOKEN` matches `PRODUCT_INTERNAL_SERVICE_TOKEN`.

## Verification Command

```powershell
curl http://localhost:8087/healthz
docker compose logs cms-service
```
