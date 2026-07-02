# Notification Service Troubleshooting

## Common Error

- OTP delivery fails.
- `/readyz` fails.
- Events remain in retry/DLQ.

## Possible Cause

- MongoDB, RabbitMQ, or Mailpit is unavailable.
- Email/SMS provider env is missing.
- Encryption key is invalid.

## Fix

- Start MongoDB, RabbitMQ, and Mailpit.
- Use local SMTP values from `.env.example`.
- Keep SMS/push disabled unless provider credentials are configured.

## Verification Command

```powershell
curl http://localhost:8092/readyz
docker compose logs notification-service
```
