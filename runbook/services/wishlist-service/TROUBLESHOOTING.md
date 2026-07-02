# Wishlist Service Troubleshooting

## Common Error

- `/readyz` fails.
- Price-drop or recommendation events are not published.
- Move-to-cart fails.

## Possible Cause

- Kafka is not healthy.
- Cart or product service is unavailable.
- MongoDB migration did not run.

## Fix

- Start Kafka, product, cart, and MongoDB migrations.
- Verify `WISHLIST_KAFKA_BROKERS` and downstream base URLs.
- Use gateway-authenticated requests for role-protected operations.

## Verification Command

```powershell
curl http://localhost:8084/readyz
docker compose logs wishlist-service
```
