# Docker Local Runbook

## First run

```bash
docker compose config
docker compose up -d --build
docker compose ps
```

First build downloads Go, Node, and image dependencies and can take several minutes.

## Daily commands

```bash
docker compose logs -f
docker compose logs -f api-gateway
docker compose up -d --build search-service
docker compose restart auth-service
docker compose down
```

Infra only:

```bash
docker compose up -d mysql mongodb redis rabbitmq kafka typesense mailpit jaeger prometheus
```

Backend after infra:

```bash
docker compose up -d auth-service user-service product-service cart-service wishlist-service search-service session-service cms-service recommendation-service order-service payment-service notification-service superadmin-service api-gateway
```

Frontends:

```bash
docker compose up -d user-app seller-dashboard superadmin-panel
```

Session Analytics is kept in the `incomplete` profile because its TypeScript build currently fails:

```bash
docker compose --profile incomplete build session-analytics-dashboard
```

## Verify

```bash
docker compose ps
curl http://localhost:8080/health/live
curl http://localhost:8081/healthz
curl http://localhost:8083/readyz
curl http://localhost:8085/healthz
```

## Troubleshooting

- Port busy: change only the left side of a Compose port mapping.
- Migration failed: inspect `docker compose logs migrate-auth` (replace service name).
- App restarts: inspect its logs; usually an env validation or dependency issue.
- Stale schema/data: `docker compose down -v`, then start again. This deletes all local data.
- Gateway readiness `503`: currently expected; incompatible downstream gRPC contracts are audited Pending.
- Analytics image fails: its existing React typings/strict TypeScript errors must be reconciled first.
- Low memory: start infra and only the services you are working on.
