# Cart Service Troubleshooting

## Common Error

- `/readyz` fails.
- Cart item validation fails.
- Coupon preview fails.

## Possible Cause

- MongoDB or Redis is down.
- Product or CMS base URL is wrong.
- Running on host with Docker DNS env values.

## Fix

- Start `mongodb`, `redis`, `product-service`, and `cms-service`.
- Use `localhost` hostnames for manual runs.
- Check `CART_PRODUCT_BASE_URL` and `CART_CMS_BASE_URL`.

## Verification Command

```powershell
curl http://localhost:8083/readyz
docker compose logs cart-service
```
