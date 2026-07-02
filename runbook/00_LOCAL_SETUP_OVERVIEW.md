# Local Setup Overview

## What Runs Locally

The full local platform contains infrastructure, backend services, worker jobs, API Gateway, and four frontend apps.

Recommended command:

```powershell
docker compose up -d --build
```

## Local Topology

```text
frontend apps
  -> api-gateway
    -> auth, user, product, cart, wishlist, search, session, cms,
       recommendation, order, payment, notification, superadmin
backend services
  -> MySQL, MongoDB, Redis, RabbitMQ, Kafka, Typesense, Mailpit
observability
  -> Prometheus, OpenTelemetry Collector, Jaeger
```

## Local URLs

| Component | URL |
| --- | --- |
| API Gateway | `http://localhost:8080` |
| User app | `http://localhost:3000` |
| Seller dashboard | `http://localhost:3001` |
| Session analytics dashboard | `http://localhost:3002` |
| Superadmin panel | `http://localhost:3003` |
| gRPC-Web facade | `http://localhost:8099` |
| Mailpit | `http://localhost:8025` |
| RabbitMQ management | `http://localhost:15672` |
| Typesense | `http://localhost:8108` |
| Jaeger | `http://localhost:16686` |
| Prometheus | `http://localhost:9095` |

## Main Known Gaps

- Seed/demo data: Not found in codebase - please confirm.
- Payment sandbox provider config: Not found in codebase - please confirm.
- Production secret management: Not found in codebase - please confirm.
- Auth signup route appears referenced by catalog/frontend, but auth route registration was not confirmed.
- Internal CMS/payment/event token and endpoint defaults are blank in examples.
- Manual host-run env values need changing from Docker DNS names to `localhost`.
