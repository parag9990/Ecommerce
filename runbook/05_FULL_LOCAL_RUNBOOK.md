# Full Local Runbook

## Intro / Purpose

Use this guide to run the ecommerce platform locally from a fresh checkout. It is written for a beginner-friendly Docker Compose path.

Local runtime source of truth: root [`docker-compose.yml`](../docker-compose.yml). The root [`Makefile`](../Makefile) provides shortcuts, but compose defines the actual service names, ports, dependencies, and health checks.

Screenshots are referenced as placeholders under `images/`. No real screenshots were captured during this documentation pass. Capture them manually after the stack runs; see [images/README.md](images/README.md).

Known blockers before full business-flow testing:

- Buyer/user app demo credentials are not seeded yet.
- Payment providers are disabled by default.
- Internal CMS/payment/event token and endpoint defaults are blank in examples.

## Step 1. Prerequisites

Install and verify the local tools.

```powershell
docker compose version
go version
node --version
corepack pnpm --version
buf --version
```

Expected versions from repo configuration:

| Tool | Version / note |
| --- | --- |
| Docker + Docker Compose | Required |
| Go | `1.26.3` in `backend/go.work` |
| Node.js | `>=22.13.0` in `frontend/package.json` |
| pnpm | `11.5.0` in `frontend/package.json` |
| Buf | Required for proto lint/generation |

Screenshot placeholder: `images/step-01-prerequisites-fhd.png`.

## Step 2. Folder Structure

Know the main folders before running commands.

```text
.
├── docker-compose.yml        # Local runtime source of truth
├── Makefile                  # Local shortcuts
├── LOCAL_RUNBOOK_INDEX.md    # Canonical runbook index
├── runbook/                  # Local setup, health, test, troubleshooting docs
├── backend/services/         # Go microservices
├── frontend/                 # React/Vite apps and packages
├── infra/                    # Infra config, observability, Kubernetes
├── proto/                    # Protobuf contracts
└── docs/                     # Project design docs
```

Screenshot placeholder: `images/step-02-folder-structure-fhd.png`.

## Step 3. Environment Setup

For Docker Compose local runs, do not copy every `.env.example` file. The root compose file reads service `.env.example` files directly for most application services.

For manual host runs only, copy the service env file you need and replace Docker DNS names with `localhost`.

```powershell
Copy-Item backend/services/api-gateway/.env.example backend/services/api-gateway/.env
```

Manual host value examples:

| Docker value | Host value |
| --- | --- |
| `mysql:3306` | `localhost:3306` |
| `mongodb:27017` | `localhost:27017` |
| `redis:6379` | `localhost:6379` |
| `rabbitmq:5672` | `localhost:5672` |
| `kafka:29092` | `localhost:9092` |
| `typesense:8108` | `localhost:8108` |
| `mailpit:1025` | `localhost:1025` |

Do not add production credentials to local env files.

Screenshot placeholder: `images/step-03-environment-setup-fhd.png`.

## Step 4. Service Start Order

Docker Compose already encodes the dependency order. When debugging manually, think in this order:

1. Infrastructure: MySQL, MongoDB, Redis, RabbitMQ, Kafka, Typesense, Mailpit, Jaeger, OpenTelemetry Collector, Prometheus.
2. Mongo replica-set init and migration jobs.
3. Backend services.
4. API Gateway.
5. Frontend apps.

Useful shortcuts:

```powershell
make docker-up
```

The phased shortcuts are useful for debugging:

```powershell
make infra-up
make backend-up
make frontend-up
```

Note: `make infra-up` follows the root `Makefile` and does not start `init-mongodb-replica` or `otel-collector` directly. The full compose command is the safest normal path.

Screenshot placeholder: `images/step-04-service-start-order-fhd.png`.

## Step 5. Full Local Run Steps

From the repo root, validate compose config first.

```powershell
docker compose config
```

Start the full local stack.

```powershell
docker compose up -d --build
```

Check service status.

```powershell
docker compose ps
```

If you need a phased run, use this order:

```powershell
docker compose up -d mysql mongodb init-mongodb-replica redis rabbitmq kafka typesense mailpit jaeger otel-collector prometheus
docker compose up migrate-auth migrate-user migrate-order migrate-payment migrate-cms migrate-superadmin migrate-mongodb migrate-session-mongodb
docker compose up -d --build auth-service user-service product-service cart-service wishlist-service search-service session-service cms-service recommendation-service order-service payment-service notification-service superadmin-service
docker compose up -d --build api-gateway
docker compose up -d --build user-app seller-dashboard session-analytics-dashboard superadmin-panel
```

Normal use should prefer `docker compose up -d --build` so compose manages dependency order.

Screenshot placeholder: `images/step-05-full-local-run-fhd.png`.

## Step 6. Health Checks

Start with the gateway and compose status.

```powershell
curl.exe http://localhost:8080/health/live
curl.exe http://localhost:8080/health/ready
docker compose ps
```

In Windows PowerShell, use `curl.exe` or `Invoke-WebRequest -UseBasicParsing` because bare `curl` is a PowerShell alias.

Then use:

- [06 Health Checks](06_HEALTH_CHECKS.md)
- [Service Health Checklist](checklists/SERVICE_HEALTH_CHECKLIST.md)
- [09 Ports And Endpoints](09_PORTS_AND_ENDPOINTS.md)

Important internal-only check:

```powershell
docker compose exec user-service wget -qO- http://localhost:9091/health/ready
```

Screenshot placeholder: `images/step-06-health-checks-fhd.png`.

## Step 7. User / Seller / Superadmin Flow

Use these after the stack and migration jobs are healthy. Local seller/admin demo credentials are seeded by the auth, user, CMS, and superadmin migrations. Seller product categories are seeded by the MongoDB product-service migration job.

| App | URL | What to verify |
| --- | --- | --- |
| User app | `http://localhost:3000` | Browse, search, cart, wishlist, checkout path |
| Seller dashboard | `http://localhost:3001` | Login with `seller.local@example.com` / `LocalDemo#2026!`; verify product, order, coupon/campaign, analytics screens |
| Session analytics dashboard | `http://localhost:3002` | Login with `analytics.admin.local@example.com` / `LocalDemo#2026!`; verify sessions, journeys, funnels, heatmaps, reports |
| Superadmin panel | `http://localhost:3003` | Login with `superadmin.local@example.com` / `LocalDemo#2026!`; verify users, sellers, payments, settings, audit views |
| Mailpit | `http://localhost:8025` | Local email/OTP messages |

Blockers:

- Buyer/user app credentials are still not seeded.
- Payment completion needs sandbox provider variables because providers are disabled by default.

Screenshot placeholder: `images/step-07-user-seller-superadmin-flow-fhd.png`.

## Step 8. Testing Checklist

Run automated tests from the root when the toolchain is ready.

```powershell
make test-go
make test-frontend
make test
```

If `make` is unavailable, follow [07 Testing Guide](07_TESTING_GUIDE.md).

Manual smoke checks:

- Gateway liveness and readiness pass.
- Frontend apps load.
- Mailpit opens.
- No service is crash-looping.
- Known blockers are recorded before attempting business-flow tests.

Screenshot placeholder: `images/step-08-testing-checklist-fhd.png`.

## Step 9. Troubleshooting

Use focused logs first.

```powershell
docker compose logs -f api-gateway
docker compose logs -f auth-service
docker compose logs -f <service-name>
```

Common next checks:

- Ports: [09 Ports And Endpoints](09_PORTS_AND_ENDPOINTS.md)
- Shared issues: [08 Common Troubleshooting](08_COMMON_TROUBLESHOOTING.md)
- Service-specific pages: [services](services/)

Screenshot placeholder: `images/step-09-troubleshooting-fhd.png`.

## Step 10. Stop Or Reset

Stop services without deleting local volumes.

```powershell
docker compose down
```

Reset local databases, cache, broker, search, and observability volumes.

```powershell
docker compose down -v
docker compose up -d --build
```

Warning: `docker compose down -v` deletes this compose project's local data volumes. It does not delete source code.

Screenshot placeholder: `images/step-10-stop-reset-fhd.png`.
