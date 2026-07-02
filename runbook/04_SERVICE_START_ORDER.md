# Service Start Order

Docker Compose already encodes most startup dependencies. Use this order when starting manually or debugging.

## Recommended Order

1. Platform foundation and infrastructure.
2. Databases: `mysql`, `mongodb`, `init-mongodb-replica`.
3. Cache, brokers, search, mail, observability: `redis`, `rabbitmq`, `kafka`, `typesense`, `mailpit`, `jaeger`, `otel-collector`, `prometheus`.
4. Migration jobs: `migrate-auth`, `migrate-user`, `migrate-order`, `migrate-payment`, `migrate-cms`, `migrate-superadmin`, `migrate-mongodb`, `migrate-session-mongodb`.
5. Low-level backend services: `notification-service`, `user-service`, `product-service`, `cms-service`, `payment-service`, `session-service`, `recommendation-service`.
6. Dependent backend services: `cart-service`, `wishlist-service`, `search-service`, `order-service`, `superadmin-service`.
7. API Gateway.
8. Frontend apps and admin panels.

## Why This Order

| Step | Why |
| --- | --- |
| Infrastructure first | Services fail fast if MySQL, MongoDB, Redis, RabbitMQ, Kafka, or Typesense are unavailable |
| Migrations before services | Service tables, collections, and indexes must exist before handlers start serving traffic |
| Auth/session early | Gateway JWT validation and session analytics depend on them |
| Notification before auth | `auth-service` has a compose health dependency on `notification-service` |
| Core services before order | `order-service` has compose dependencies on cart, product, payment, and Kafka |
| Search after product | `search-service` has a compose health dependency on `product-service` plus Typesense/Redis/RabbitMQ |
| Superadmin after admin dependencies | `superadmin-service` waits for user, order, payment, and session services |
| Gateway after backend | Gateway readiness checks downstream services |
| Frontends last | Apps need the gateway URL and CORS to be available |

## Compose Shortcuts

```powershell
make infra-up
make backend-up
make frontend-up
```

Note: `make infra-up` follows the root `Makefile` and does not directly start `init-mongodb-replica` or `otel-collector`. Use the full compose command for the normal complete local run, or start those services separately when debugging phased startup.

Or run everything:

```powershell
make docker-up
```

If `make` is unavailable:

```powershell
docker compose up -d --build
```
