# Deployment

## Local Docker Compose

The most complete local deployment path is the root `docker-compose.yml`.

## Compose Services

### Infrastructure

| Service | Purpose |
| --- | --- |
| `mysql` | MySQL 8.4 databases |
| `mongodb` | MongoDB 8.0 replica set |
| `mongodb-init-replica` | Initializes Mongo replica set |
| `redis` | Redis cache/rate limit/session support |
| `rabbitmq` | RabbitMQ broker |
| `kafka` | Kafka broker |
| `typesense` | Search engine |
| `mailpit` | Local email sink |
| `jaeger` | Trace UI/backend |
| `otel-collector` | OpenTelemetry collector |
| `prometheus` | Metrics scraping |

### Migration Jobs

| Job | Purpose |
| --- | --- |
| `migrate-auth` | Auth MySQL migrations |
| `migrate-user` | User MySQL migrations |
| `migrate-order` | Order MySQL migrations |
| `migrate-payment` | Payment MySQL migrations |
| `migrate-cms` | CMS MySQL migrations |
| `migrate-superadmin` | Superadmin MySQL migrations |
| Mongo migration jobs | Product, cart, wishlist, recommendation, session, notification migrations |

### Backend Services

| Service | Local host ports |
| --- | --- |
| `api-gateway` | `8080`, gRPC-Web `8099`, metrics `19090` |
| `auth-service` | `8081` |
| `user-service` | gRPC `50052`, admin/health `9091` |
| `product-service` | HTTP `8082`, gRPC `9093` |
| `cart-service` | `8083` |
| `wishlist-service` | `8084` |
| `search-service` | HTTP `8085`, gRPC `50055` |
| `session-service` | `8086` |
| `cms-service` | HTTP `8087`, gRPC `50057` |
| `recommendation-service` | HTTP `8089`, gRPC `50058` |
| `order-service` | HTTP `8090`, gRPC `9094` |
| `payment-service` | `8091` |
| `notification-service` | HTTP `8092`, gRPC `50060` |
| `superadmin-service` | `8093` |

### Frontend Services

| Service | Local host port |
| --- | --- |
| `frontend-user-app` | `3000` |
| `frontend-seller-dashboard` | `3001` |
| `frontend-session-analytics-dashboard` | `3002` |
| `frontend-superadmin-panel` | `3003` |

## Infra-Only Compose

`infra/compose/docker-compose.local.yml` provides an infra-focused compose setup for MySQL, MongoDB, Typesense, Redis, RabbitMQ, optional Kafka, Jaeger, OTEL, and Prometheus.

Example environment file: `infra/compose/.env.local.example`.

## Tilt

`Tiltfile` wraps the root compose stack and labels resources into groups:

- `infra`
- `observability`
- `backend`
- `frontend`

It also models dependencies such as API Gateway depending on Redis/Auth and frontend apps depending on API Gateway.

## Kubernetes

Kubernetes manifests exist in two locations:

| Path | Scope |
| --- | --- |
| `infra/k8s/base` | Broader Kustomize base with namespaces, data, observability, jobs, and several backend services |
| `infra/k8s/overlays/dev` | Dev overlay with local image tags and replica patches |
| `deployments/k8s/local` | Partial local Kubernetes example for namespace/config/secret/api-gateway/product-service/ingress |

`deployments/k8s/local/README.md` states that Docker Compose is the easiest full-stack local option and that datastore dependencies are expected outside those local Kubernetes manifests.

## CI/CD

GitHub Actions workflow: `.github/workflows/ci.yml`.

| Job | Checks found |
| --- | --- |
| Go | Setup Go 1.26.3, module tests, vet, gofmt, golangci-lint |
| Frontend | Setup Node 22, pnpm frozen install, recursive test/lint/typecheck/build |
| Proto | Buf lint, Buf breaking against dev branch, generated client diff check |
| Containers | Hadolint, Docker buildx build, Trivy HIGH/CRITICAL scan |

CI helper docs/scripts exist under `infra/ci`.

## Build Commands

| Area | Command/source |
| --- | --- |
| Frontend build | `pnpm build` from `frontend/package.json` |
| Service images | Dockerfiles under service folders and CI matrix |
| Proto generation | `buf generate` via `buf.gen.yaml` and CI helper script |
| Compose local stack | `docker compose up` using root `docker-compose.yml` |

## Required External Services

For local compose, services are provided by the compose stack. For non-compose deployment, the application requires equivalents for:

- MySQL.
- MongoDB replica set.
- Redis.
- RabbitMQ and/or Kafka depending on enabled events.
- Typesense.
- SMTP or email provider.
- Payment provider APIs if payments are enabled.
- Observability backend if tracing/metrics are enabled.

## Missing Or Partial Deployment Information

- Production deployment flow: Not found in current codebase.
- Production secret management procedure: Not found in current codebase.
- Full Kubernetes manifests for every service in one active overlay: Not found in current codebase.
- Helm charts: Not found in current codebase.
