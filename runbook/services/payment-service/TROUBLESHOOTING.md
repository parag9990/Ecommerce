# Payment Service Troubleshooting

## Common Error

- Payment intent provider unavailable.
- Webhook rejected.
- Payment does not update order.

## Possible Cause

- `PAYMENT_ALLOWED_PROVIDERS` and `PAYMENT_DEFAULT_PROVIDER` are empty.
- Provider sandbox credentials are missing.
- `PAYMENT_EVENTS_ENDPOINT` is empty or token mismatch.
- Webhook signing secret mismatch.

## Fix

- Configure one sandbox provider and matching provider secrets.
- Configure `PAYMENT_EVENTS_ENDPOINT` to the order payment event endpoint when testing payment-to-order flow.
- Match `PAYMENT_INTERNAL_API_TOKEN` with gateway/order/superadmin env.

## Verification Command

```powershell
curl http://localhost:8091/healthz
docker compose logs payment-service
```
