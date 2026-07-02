# Local Reset And Cleanup

## Stop Services

```powershell
docker compose down
```

## Stop Infrastructure Only

```powershell
make infra-down
```

Note: `make infra-down` stops the infrastructure list from the root `Makefile`. It does not include `otel-collector`, and `init-mongodb-replica` is normally an exited setup job.

If `make` is unavailable:

```powershell
docker compose stop mysql mongodb redis rabbitmq kafka typesense mailpit jaeger otel-collector prometheus
```

## Remove Local Compose Volumes

WARNING: this deletes local database/cache/broker/search data for this compose project. It does not delete source code.

```powershell
docker compose down -v
```

## Rebuild From Scratch

```powershell
docker compose down -v
docker compose up -d --build
```

## Reinstall Frontend Dependencies

```powershell
cd frontend
corepack pnpm install --frozen-lockfile
```

If dependency folders are corrupted, delete only `frontend/node_modules` and app-local `node_modules` directories from inside this repository, then rerun install.

## Refresh Go Modules

```powershell
cd backend
go work sync
```

Run backend tests through the repo loop:

```powershell
make test-go
```

If `make` is unavailable, run tests module by module:

```powershell
Get-ChildItem backend/services,backend/shared,backend/proto-gen -Recurse -Filter go.mod -File | ForEach-Object {
  Push-Location $_.DirectoryName
  go test ./...
  Pop-Location
}
```

Run `go mod tidy` only inside a service module when you intentionally changed imports.

## Safe Cleanup Rules

- Do not delete source folders.
- Do not run broad Docker prune commands unless you intentionally want to remove unrelated local Docker assets.
- Prefer `docker compose down -v` over manual container/volume deletion because it is scoped to this project.
