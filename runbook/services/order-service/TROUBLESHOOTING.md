# Order Service Troubleshooting

## Common Error

- Checkout fails.
- Payment result does not update order.
- `/readyz` fails.

## Possible Cause

- Cart, product, or payment service is unavailable.
- MySQL migration missing.
- Payment/internal tokens mismatch.
- Kafka is down.

## Fix

- Start `cart-service`, `product-service`, `payment-service`, and Kafka first.
- Run `docker compose up migrate-order`.
- Match `ORDER_PRODUCT_SERVICE_TOKEN`, `ORDER_PAYMENT_INTERNAL_TOKEN`, and payment service env.

## Verification Command

```powershell
curl http://localhost:8090/readyz
docker compose logs order-service
```
