# Search Service Troubleshooting

## Common Error

- `/readyz` fails.
- Search returns no products.
- Reindex command fails.

## Possible Cause

- Typesense is down or API key mismatch.
- Product service is unavailable or has no data.
- RabbitMQ product event stream is not configured.

## Fix

- Start `typesense`, `product-service`, `redis`, and `rabbitmq`.
- Verify `TYPESENSE_API_KEY` matches compose command.
- Run a reindex after products exist.

## Verification Command

```powershell
curl http://localhost:8085/readyz
curl http://localhost:8108/health
```
