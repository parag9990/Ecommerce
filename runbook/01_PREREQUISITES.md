# Prerequisites

## Required Tools

| Tool | Version found | Why |
| --- | --- | --- |
| Docker Desktop or Docker Engine | Docker Compose v2 compatible | Full local stack |
| Go | `1.26.3` in `backend/go.work`; most backend Dockerfiles use `GO_VERSION=1.26.4` | Backend services and tests |
| Node.js | `>=22.13.0` in `frontend/package.json` | Frontend workspace |
| pnpm | `11.5.0` in `frontend/package.json` | Frontend package manager |
| Corepack | Bundled with Node | Activates pnpm |
| Buf | Version not pinned in repo | Proto lint/generate |
| curl | Any recent version | Health checks |

## Optional Tools

| Tool | Why |
| --- | --- |
| Make | Runs root `Makefile` shortcuts |
| Tilt | `Tiltfile` exists for compose orchestration |
| grpcurl | Manual gRPC health checks |
| mongosh | Manual MongoDB inspection |
| mysql client | Manual MySQL inspection |
| redis-cli | Manual Redis inspection |

## Quick Checks

```powershell
docker compose version
go version
node --version
corepack --version
pnpm --version
```

## Version Notes

- Backend service `go.mod` files use `go 1.26.3`.
- Most backend Dockerfiles build with `golang:1.26.4-alpine`; `cart-service` uses `1.26.3`.
- `backend/shared/gen/go` and `backend/shared/validation` use `go 1.25.0`.
- Frontend package engines require Node `>=22.13.0`; frontend Dockerfiles use `node:24-alpine`.

If `pnpm` is missing:

```powershell
corepack enable
corepack prepare pnpm@11.5.0 --activate
```
