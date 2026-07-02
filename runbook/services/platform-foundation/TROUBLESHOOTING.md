# Platform Foundation Troubleshooting

## Common Error

- `docker compose up` fails.
- Infrastructure service is unhealthy.
- Proto or frontend workspace commands fail.

## Possible Cause

- Docker is not running.
- Local port conflict.
- Compose env override is invalid.
- Node/pnpm/Go version mismatch.
- Running `go test ./...` from `backend/`, which is not a module root.

## Fix

- Start Docker.
- Free ports listed in [Ports And Endpoints](../../09_PORTS_AND_ENDPOINTS.md).
- Run `docker compose config`.
- Use Node `>=22.13.0`, pnpm `11.5.0`, and Go `1.26.3`.
- Use `make test-go` or run `go test ./...` inside each module directory.

## Verification Command

```powershell
docker compose config
docker compose ps
```
