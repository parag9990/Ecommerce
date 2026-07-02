# Product Service Troubleshooting

## Common Error

- `/readyz` fails.
- Product writes work but search does not update.
- Order inventory reservation fails.

## Possible Cause

- Mongo migration did not run.
- RabbitMQ event publishing is unavailable.
- `PRODUCT_INTERNAL_SERVICE_TOKEN` differs from caller config.

## Fix

- Run `docker compose up migrate-mongodb`.
- Start RabbitMQ before product service.
- Match product token with order/search/CMS env values.

## Verification Command

```powershell
curl http://localhost:8082/readyz
docker compose logs product-service
```
