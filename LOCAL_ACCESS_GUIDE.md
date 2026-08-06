# Local Access Guide

This guide lists the local UIs, dashboards, admin tools, ports, credentials, and quick checks for the Docker-based local stack.

Start everything:

```bash
docker compose up -d --build
```

Start one service:

```bash
docker compose up -d --build <service-name>
```

## Access Matrix

| Service | Local URL | Docker service | Login | Password / key | Credential source | Verify |
| --- | --- | --- | --- | --- | --- | --- |
| User App Frontend | http://localhost:3000 | `user-app` | `buyer.local@example.com` | `LocalDemo#2026!` | `backend/services/auth-service/migrations/007_seed_local_buyer_account.up.sql`, `backend/services/user-service/migrations/005_seed_local_buyer_user.up.sql` | http://localhost:3000/healthz |
| Seller Dashboard / CMS | http://localhost:3001 | `seller-dashboard` | `seller.local@example.com` | `LocalDemo#2026!` | `backend/services/auth-service/migrations/006_seed_local_access_accounts.up.sql`, `backend/services/user-service/migrations/004_seed_local_access_users.up.sql`, `backend/services/cms-service/migrations/008_seed_local_seller_staff.up.sql` | http://localhost:3001/healthz |
| Session Analytics Dashboard | http://localhost:3002 | `session-analytics-dashboard` | `analytics.admin.local@example.com` | `LocalDemo#2026!` | `backend/services/auth-service/migrations/006_seed_local_access_accounts.up.sql`, `backend/services/user-service/migrations/004_seed_local_access_users.up.sql`, `backend/services/superadmin-service/migrations/007_seed_local_admin_users.up.sql` | http://localhost:3002/healthz |
| Superadmin Panel | http://localhost:3003 | `superadmin-panel` | `superadmin.local@example.com` | `LocalDemo#2026!` | `backend/services/auth-service/migrations/006_seed_local_access_accounts.up.sql`, `backend/services/user-service/migrations/004_seed_local_access_users.up.sql`, `backend/services/superadmin-service/migrations/007_seed_local_admin_users.up.sql` | http://localhost:3003/healthz |
| API Gateway | http://localhost:8080 | `api-gateway` | N/A | N/A | No browser UI found; API gateway only | http://localhost:8080/health/ready |
| API Gateway gRPC-Web | http://localhost:8099 | `api-gateway` | N/A | N/A | `docker-compose.yml` | Connect with gRPC-Web client |
| API Gateway Metrics | http://localhost:19090/metrics | `api-gateway` | N/A | N/A | `docker-compose.yml` | Open metrics URL |
| RabbitMQ Management | http://localhost:15672 | `rabbitmq` | `ecommerce` | `local_rabbitmq_password` | `docker-compose.yml` | Browser login |
| Mailpit | http://localhost:8025 | `mailpit` | No login | No login | `docker-compose.yml` | Browser opens inbox |
| Jaeger | http://localhost:16686 | `jaeger` | No login | No login | `docker-compose.yml` | Browser opens tracing UI |
| Prometheus | http://localhost:9095 | `prometheus` | No login | No login | `docker-compose.yml` | Browser opens Prometheus |
| Typesense API | http://localhost:8108 | `typesense` | N/A | `local_typesense_key` | `docker-compose.yml` | http://localhost:8108/health |
| MySQL | localhost:3306 | `mysql` | `root` | `local_password` | `docker-compose.yml` | Container healthcheck |
| MongoDB | localhost:27017 | `mongodb` | `local_admin` | `local_mongo_password` | `docker-compose.yml` | Container healthcheck |
| Redis | localhost:6379 | `redis` | N/A | `local_redis_password` | `docker-compose.yml` | `redis-cli -a local_redis_password ping` |
| Kafka | localhost:9092 | `kafka` | No auth configured | No auth configured | `docker-compose.yml` | Broker reachable |

## App UIs

### User App Frontend

- URL: http://localhost:3000
- Docker service: `user-app`
- Depends on: `api-gateway`, `auth-service`, and the backend services needed by the user flow.
- Login role/type: buyer or normal user.
- Local credentials: `buyer.local@example.com` / `LocalDemo#2026!`.
- Seeded data: active auth account, active user profile, and buyer role assignment.
- If login fails: confirm `api-gateway` is ready at http://localhost:8080/health/ready and `auth-service` is ready at http://localhost:8081/readyz. If the database existed before these seed migrations were updated, rerun the migration jobs with `docker compose up --force-recreate migrate-auth migrate-user`.

### Seller Dashboard / CMS

- URL: http://localhost:3001
- Docker service: `seller-dashboard`
- Depends on: `api-gateway`, `auth-service`, `cms-service`, `product-service`, `order-service`, `payment-service`, and an active seller account.
- Login role/type: seller.
- Local credentials: `seller.local@example.com` / `LocalDemo#2026!`.
- Seeded data: active auth account, active user, active seller profile `seller_local_demo`, active CMS staff row, and active product categories for product creation.
- If login fails: confirm `api-gateway` is ready at http://localhost:8080/health/ready and `auth-service` is ready at http://localhost:8081/readyz. If the database existed before these seed migrations were added, rerun the migration jobs with `docker compose up --force-recreate migrate-auth migrate-user migrate-cms migrate-mongodb`.

### Session Analytics Dashboard

- URL: http://localhost:3002
- Docker service: `session-analytics-dashboard`
- Depends on: `api-gateway`, `auth-service`, and `session-service`.
- Login role/type: admin, operations admin, finance admin, catalog admin, or superadmin.
- Local credentials: `analytics.admin.local@example.com` / `LocalDemo#2026!`.
- Seeded role: `operations_admin`.
- If login fails: verify http://localhost:8080/health/ready and http://localhost:8086/healthz. If the database existed before these seed migrations were added, rerun the migration jobs with `docker compose up --force-recreate migrate-auth migrate-user migrate-superadmin`.

### Superadmin Panel

- URL: http://localhost:3003
- Docker service: `superadmin-panel`
- Depends on: `api-gateway`, `auth-service`, `superadmin-service`, `user-service`, `order-service`, `payment-service`, and `session-service`.
- Login role/type: superadmin or admin role.
- Local credentials: `superadmin.local@example.com` / `LocalDemo#2026!`.
- Seeded role: `superadmin`.
- If login fails: verify http://localhost:8080/health/ready and http://localhost:8093/readyz. If the database existed before these seed migrations were added, rerun the migration jobs with `docker compose up --force-recreate migrate-auth migrate-user migrate-superadmin`.

## API And Service Checks

| Service | URL |
| --- | --- |
| API Gateway live check | http://localhost:8080/health/live |
| API Gateway ready check | http://localhost:8080/health/ready |
| Auth Service | http://localhost:8081/readyz |
| Product Service | http://localhost:8082/readyz |
| Cart Service | http://localhost:8083/readyz |
| Wishlist Service | http://localhost:8084/readyz |
| Search Service | http://localhost:8085/readyz |
| Session Service | http://localhost:8086/healthz |
| CMS Service | http://localhost:8087/healthz |
| Recommendation Service | http://localhost:8089/healthz |
| Order Service | http://localhost:8090/readyz |
| Payment Service | http://localhost:8091/healthz |
| Notification Service | http://localhost:8092/readyz |
| Superadmin Service | http://localhost:8093/readyz |

No Swagger, Redoc, GraphQL Playground, database UI, Redis UI, Kafka UI, or search dashboard is configured in the root Docker Compose file. The API contract file is `api/master-api.json`.

## Notification And Infra Dashboards

- Notification service API: http://localhost:8092
- Email capture UI: http://localhost:8025
- RabbitMQ queues and exchanges: http://localhost:15672
- Traces: http://localhost:16686
- Metrics: http://localhost:9095 and http://localhost:19090/metrics
- Typesense health/API: http://localhost:8108/health

## Seeded Local Credentials

These human app credentials are local/demo only:

| Area | Email | Password | Role/context |
| --- | --- | --- | --- |
| User App Frontend | `buyer.local@example.com` | `LocalDemo#2026!` | `buyer` |
| Seller Dashboard / CMS | `seller.local@example.com` | `LocalDemo#2026!` | `seller`, `seller_local_demo` |
| Session Analytics Dashboard | `analytics.admin.local@example.com` | `LocalDemo#2026!` | `operations_admin` |
| Superadmin Panel | `superadmin.local@example.com` | `LocalDemo#2026!` | `superadmin` |

Do not put production credentials in these files. Use obvious local-only emails and passwords, and keep any seed marked as local/demo data.

## Quick Troubleshooting

1. Check the main gateway first: http://localhost:8080/health/ready.
2. Check the target frontend health URL from the Access Matrix.
3. Check backend logs with `docker compose logs <service-name>`.
4. If a seeded app login fails or seller product categories are missing on an older database, rerun the relevant migration jobs with `docker compose up --force-recreate migrate-auth migrate-user migrate-cms migrate-superadmin migrate-mongodb`.
5. If RabbitMQ, MongoDB, MySQL, Redis, or Typesense login fails, compare the password with `docker-compose.yml`.
