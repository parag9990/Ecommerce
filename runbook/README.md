# Local Runbook

This folder documents how to run the full project locally from the repository state found during analysis. Use this README together with the top-level [Local Runbook Index](../LOCAL_RUNBOOK_INDEX.md) as the canonical local documentation entry point.

## Start Here

1. Read [05 Full Local Runbook](05_FULL_LOCAL_RUNBOOK.md) for the beginner-friendly step-by-step flow.
2. Keep [09 Ports And Endpoints](09_PORTS_AND_ENDPOINTS.md) open while checking URLs.
3. Use [06 Health Checks](06_HEALTH_CHECKS.md) and [checklists/SERVICE_HEALTH_CHECKLIST.md](checklists/SERVICE_HEALTH_CHECKLIST.md) after the stack starts.
4. Use [08 Common Troubleshooting](08_COMMON_TROUBLESHOOTING.md) when a service is unhealthy.

Screenshot placeholders for the full guide are listed in [images/README.md](images/README.md).

## Runtime Source Of Truth

Root [docker-compose.yml](../docker-compose.yml) is the local runtime source of truth. The root [Makefile](../Makefile) contains useful shortcuts, but compose defines the service names, ports, dependencies, and health checks used by the runbook.

## Evidence Used

- `docker-compose.yml`
- `Makefile`
- `backend/go.work`
- Backend service `go.mod`, `Dockerfile`, `cmd`, `migrations`, `.env.example`, and config files
- Frontend `package.json`, `pnpm-workspace.yaml`, Vite configs, Dockerfiles, and `.env.example`
- `api/master-api.json`
- Existing docs under `Project_Details/` and `docs/`

## Service Inventory

| Service | Location | Runtime |
| --- | --- | --- |
| API Gateway | `backend/services/api-gateway` | Go HTTP plus gRPC-Web |
| Auth Service | `backend/services/auth-service` | Go HTTP |
| User Service | `backend/services/user-service` | Go gRPC plus admin HTTP |
| Product Service | `backend/services/product-service` | Go HTTP plus gRPC |
| Cart Service | `backend/services/cart-service` | Go HTTP plus optional worker |
| Wishlist Service | `backend/services/wishlist-service` | Go HTTP |
| Search Service | `backend/services/search-service` | Go HTTP plus gRPC plus reindex commands |
| Session Management Service | `backend/services/session-service` | Go HTTP plus retention worker |
| CMS Service | `backend/services/cms-service` | Go HTTP plus gRPC |
| Recommendation Service | `backend/services/recommendation-service` | Go HTTP plus gRPC |
| Order Service | `backend/services/order-service` | Go HTTP plus gRPC |
| Payment Service | `backend/services/payment-service` | Go HTTP plus reconciliation command |
| Notification Service | `backend/services/notification-service` | Go gRPC plus internal HTTP |
| Superadmin Service | `backend/services/superadmin-service` | Go HTTP |
| User App Frontend | `frontend/user-app` | React/Vite |
| Seller Dashboard CMS | `frontend/seller-dashboard` | React/Vite |
| Session Analytics Dashboard | `frontend/session-analytics-dashboard` | React/Vite |
| Superadmin Panel | `frontend/superadmin-panel` | React/Vite |
| Platform Foundation | `docker-compose.yml`, `infra/`, `backend/shared`, `proto/` | Infra, shared code, proto |

## Recommended Local Mode

Use Docker Compose. It is the only complete local runtime found in the codebase.

```powershell
docker compose up -d --build
```

Manual local execution is possible service by service, but it requires host-local env changes because `.env.example` files use Docker service DNS names like `mysql`, `mongodb`, `redis`, `rabbitmq`, and `kafka`.

## Important Warnings

- Do not put production credentials in `.env.example`.
- Payment providers are disabled by default. Real payment flow testing needs sandbox provider variables.
- Seed/demo users and seeding commands were not found.
- Auth signup is referenced by the catalog/frontend, but auth route registration was not confirmed.
- Internal CMS/payment/event token and endpoint defaults are blank in examples.
- Older local runbook names under `docs/runbooks/` were replaced by this `runbook/` folder. The remaining `docs/runbooks/merge-feature-services-to-dev.md` is a git workflow guide.
