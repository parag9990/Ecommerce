# Local Platform Runbook

Docker Compose is the supported first path. It starts infrastructure, migrations, 14 backend containers, and the 3 buildable frontend apps. Session Analytics remains available under the `incomplete` profile for repair work.

## Prerequisites

- Docker Desktop with Compose v2
- 12 GB free RAM recommended
- Go 1.26+ and Node 24+ only for host development
- `corepack`/pnpm 11.5 for frontend host development

## One-time check

```bash
docker compose config
docker compose pull
```

Committed `.env.example` files contain safe local-only values and are read directly by Compose. Do not reuse them outside local development.

## Option 1: full Docker Compose

```bash
docker compose up -d --build
docker compose ps
docker compose logs -f api-gateway
```

Open `http://localhost:3000` and check gateway liveness:

```bash
curl http://localhost:8080/health/live
```

Gateway readiness is expected to be `503` until the Pending gRPC/REST bridges are implemented. Product, Order, and Session run dependency/health shells, not business APIs. Session Analytics is excluded by default because its production TypeScript build fails.

## Option 2: apps on host, infrastructure in Docker

```bash
docker compose up -d mysql mongodb redis rabbitmq kafka typesense mailpit jaeger prometheus
docker compose run --rm migrate-auth
docker compose run --rm migrate-user
docker compose run --rm migrate-order
docker compose run --rm migrate-payment
docker compose run --rm migrate-cms
docker compose run --rm migrate-superadmin
docker compose run --rm migrate-mongodb
```

Copy a service `.env.example`, replace Docker hosts (`mysql`, `mongodb`, `redis`) with `localhost`, then run from that service folder:

```bash
go run ./cmd/server
```

Frontend:

```bash
cd frontend
corepack pnpm install
corepack pnpm --filter user-app dev --host 0.0.0.0
```

## Option 3: Kubernetes and Tilt

The repository includes a local namespace/config foundation, but full app Deployments are intentionally blocked by the incomplete Product/Order transport wiring and gateway protocol mismatch. Use Compose today; see `KUBERNETES_TILT_LOCAL_RUNBOOK.md` for the future path.

## Stop or reset

```bash
docker compose down
docker compose down -v  # deletes local DB/queue/search data
```

## More help

- `DOCKER_LOCAL_RUNBOOK.md`
- `ENVIRONMENT_SETUP_GUIDE.md`
- `LOCAL_VERIFICATION_CHECKLIST.md`
- `LOCAL_SECURITY_AND_CONFIG_AUDIT.md`
