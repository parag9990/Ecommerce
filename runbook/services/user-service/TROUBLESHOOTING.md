# User Service Troubleshooting

## Common Error

- `/health/ready` fails.
- Gateway cannot reach user gRPC.
- Outbox worker logs publish failures.

## Possible Cause

- MySQL is not migrated.
- `USER_SERVICE_DATABASE_DSN` uses Docker DNS during host run.
- RabbitMQ is down or vhost credentials mismatch.

## Fix

- Run `docker compose up migrate-user`.
- Use `mysql:3306` in Docker and `localhost:3306` on host.
- Confirm `RABBITMQ_URL`.

## Verification Command

```powershell
docker compose exec user-service wget -qO- http://localhost:9091/health/ready
docker compose logs user-service
```
