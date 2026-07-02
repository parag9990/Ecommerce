# Local Runbook Index

Start here when running the full ecommerce platform locally. This file and [runbook/README.md](runbook/README.md) are the canonical local runbook entry points.

Source of truth for local runtime: root [docker-compose.yml](docker-compose.yml). The root [Makefile](Makefile) provides shortcuts, but the compose file is the final authority for service names, ports, dependencies, and health checks.

## Recommended Path

Use Docker Compose for the complete platform.

Step-by-step guide: [05 Full Local Runbook](runbook/05_FULL_LOCAL_RUNBOOK.md).

```powershell
docker compose up -d --build
```

Then verify the gateway and apps:

```powershell
curl http://localhost:8080/health/live
curl http://localhost:8080/health/ready
```

## Main Guides

| Guide | Purpose |
| --- | --- |
| [Runbook README](runbook/README.md) | Service inventory and evidence summary |
| [00 Local Setup Overview](runbook/00_LOCAL_SETUP_OVERVIEW.md) | Topology, local assumptions, known blockers |
| [01 Prerequisites](runbook/01_PREREQUISITES.md) | Required tools and versions |
| [02 Environment Variables](runbook/02_ENVIRONMENT_VARIABLES.md) | Env sources, copied files, high-risk secrets |
| [03 Database Setup](runbook/03_DATABASE_SETUP.md) | MySQL, MongoDB, migrations, seed gaps |
| [04 Service Start Order](runbook/04_SERVICE_START_ORDER.md) | Startup order and dependency reasons |
| [05 Full Local Runbook](runbook/05_FULL_LOCAL_RUNBOOK.md) | Beginner-friendly fresh clone to running platform guide |
| [06 Health Checks](runbook/06_HEALTH_CHECKS.md) | Service health URLs and commands |
| [07 Testing Guide](runbook/07_TESTING_GUIDE.md) | Backend, frontend, and functional test flow |
| [08 Common Troubleshooting](runbook/08_COMMON_TROUBLESHOOTING.md) | Shared failure modes and fixes |
| [09 Ports And Endpoints](runbook/09_PORTS_AND_ENDPOINTS.md) | Ports, URLs, health URLs, dependencies |
| [10 Local Reset And Cleanup](runbook/10_LOCAL_RESET_AND_CLEANUP.md) | Safe stop, reset, and cleanup commands |
| [Screenshot Placeholders](runbook/images/README.md) | Manual HD/FHD screenshot capture list |

## Checklists

| Checklist | Use |
| --- | --- |
| [Pre-run](runbook/checklists/PRE_RUN_CHECKLIST.md) | Before starting the stack |
| [Post-run](runbook/checklists/POST_RUN_CHECKLIST.md) | After services are started |
| [Service health](runbook/checklists/SERVICE_HEALTH_CHECKLIST.md) | Health endpoint verification |
| [Functional test](runbook/checklists/FUNCTIONAL_TEST_CHECKLIST.md) | Buyer, seller, and superadmin smoke flow |

## Service Runbooks

Service pages live under [runbook/services](runbook/services/). Each service has a `RUNBOOK.md` and `TROUBLESHOOTING.md`.

| Service | Runbook | Troubleshooting |
| --- | --- | --- |
| API Gateway | [RUNBOOK](runbook/services/api-gateway/RUNBOOK.md) | [TROUBLESHOOTING](runbook/services/api-gateway/TROUBLESHOOTING.md) |
| Auth Service | [RUNBOOK](runbook/services/auth-service/RUNBOOK.md) | [TROUBLESHOOTING](runbook/services/auth-service/TROUBLESHOOTING.md) |
| User Service | [RUNBOOK](runbook/services/user-service/RUNBOOK.md) | [TROUBLESHOOTING](runbook/services/user-service/TROUBLESHOOTING.md) |
| Product Service | [RUNBOOK](runbook/services/product-service/RUNBOOK.md) | [TROUBLESHOOTING](runbook/services/product-service/TROUBLESHOOTING.md) |
| Cart Service | [RUNBOOK](runbook/services/cart-service/RUNBOOK.md) | [TROUBLESHOOTING](runbook/services/cart-service/TROUBLESHOOTING.md) |
| Wishlist Service | [RUNBOOK](runbook/services/wishlist-service/RUNBOOK.md) | [TROUBLESHOOTING](runbook/services/wishlist-service/TROUBLESHOOTING.md) |
| Search Service | [RUNBOOK](runbook/services/search-service/RUNBOOK.md) | [TROUBLESHOOTING](runbook/services/search-service/TROUBLESHOOTING.md) |
| Session Management Service | [RUNBOOK](runbook/services/session-management-service/RUNBOOK.md) | [TROUBLESHOOTING](runbook/services/session-management-service/TROUBLESHOOTING.md) |
| CMS Service | [RUNBOOK](runbook/services/cms-service/RUNBOOK.md) | [TROUBLESHOOTING](runbook/services/cms-service/TROUBLESHOOTING.md) |
| Recommendation Service | [RUNBOOK](runbook/services/recommendation-service/RUNBOOK.md) | [TROUBLESHOOTING](runbook/services/recommendation-service/TROUBLESHOOTING.md) |
| Order Service | [RUNBOOK](runbook/services/order-service/RUNBOOK.md) | [TROUBLESHOOTING](runbook/services/order-service/TROUBLESHOOTING.md) |
| Payment Service | [RUNBOOK](runbook/services/payment-service/RUNBOOK.md) | [TROUBLESHOOTING](runbook/services/payment-service/TROUBLESHOOTING.md) |
| Notification Service | [RUNBOOK](runbook/services/notification-service/RUNBOOK.md) | [TROUBLESHOOTING](runbook/services/notification-service/TROUBLESHOOTING.md) |
| Superadmin Service | [RUNBOOK](runbook/services/superadmin-service/RUNBOOK.md) | [TROUBLESHOOTING](runbook/services/superadmin-service/TROUBLESHOOTING.md) |
| User App Frontend | [RUNBOOK](runbook/services/user-app-frontend/RUNBOOK.md) | [TROUBLESHOOTING](runbook/services/user-app-frontend/TROUBLESHOOTING.md) |
| Seller Dashboard CMS | [RUNBOOK](runbook/services/seller-dashboard-cms/RUNBOOK.md) | [TROUBLESHOOTING](runbook/services/seller-dashboard-cms/TROUBLESHOOTING.md) |
| Session Analytics Dashboard | [RUNBOOK](runbook/services/session-analytics-dashboard/RUNBOOK.md) | [TROUBLESHOOTING](runbook/services/session-analytics-dashboard/TROUBLESHOOTING.md) |
| Superadmin Panel | [RUNBOOK](runbook/services/superadmin-panel/RUNBOOK.md) | [TROUBLESHOOTING](runbook/services/superadmin-panel/TROUBLESHOOTING.md) |
| Platform Foundation | [RUNBOOK](runbook/services/platform-foundation/RUNBOOK.md) | [TROUBLESHOOTING](runbook/services/platform-foundation/TROUBLESHOOTING.md) |

Known blockers to confirm before full business-flow testing:

- Buyer seed/demo user: Not found in codebase - please confirm.
- Seller, session analytics admin, and superadmin local credentials are listed in [LOCAL_ACCESS_GUIDE.md](LOCAL_ACCESS_GUIDE.md).
- Payment providers are disabled in local defaults.
- Internal CMS/payment/event token and endpoint defaults are blank in examples.

## Documentation Structure

- Local run docs live in [runbook/](runbook/).
- Service-specific local docs live in [runbook/services/](runbook/services/).
- General project design docs live in [docs/](docs/).
- `docs/runbooks/merge-feature-services-to-dev.md` is a git workflow runbook, not the canonical local setup guide.
