# Database Setup

## Datastores

| Store | Local compose service | Used by |
| --- | --- | --- |
| MySQL 8.4 | `mysql` | auth, user, order, payment, CMS, superadmin |
| MongoDB 8.0 replica set | `mongodb` plus `init-mongodb-replica` | product, cart, wishlist, recommendation, session, notification |
| Redis 7.4 | `redis` | gateway, auth, cart, search, session, recommendation |
| Typesense 29 | `typesense` | search |
| RabbitMQ 4.1 | `rabbitmq` | user, product, search, notification |
| Kafka 3.9 | `kafka` | order, wishlist, recommendation |

## MySQL Databases

Created by `infra/mysql/init/00-create-databases.sql`:

- `auth_db`
- `user_db`
- `order_db`
- `payment_db`
- `cms_db`
- `superadmin_db`

## Migrations

Docker Compose includes migration jobs:

| Job | Service |
| --- | --- |
| `migrate-auth` | Auth MySQL migrations |
| `migrate-user` | User MySQL migrations |
| `migrate-order` | Order MySQL migrations |
| `migrate-payment` | Payment MySQL migrations |
| `migrate-cms` | CMS MySQL migrations |
| `migrate-superadmin` | Superadmin MySQL migrations |
| `migrate-mongodb` | Product, cart, wishlist, recommendation, session, notification Mongo migrations |
| `migrate-session-mongodb` | Session Mongo migrations |

## Recommended Setup

```powershell
docker compose up -d mysql mongodb init-mongodb-replica redis rabbitmq kafka typesense mailpit
docker compose up migrate-auth migrate-user migrate-order migrate-payment migrate-cms migrate-superadmin migrate-mongodb migrate-session-mongodb
```

The full stack command runs dependencies through `depends_on`:

```powershell
docker compose up -d --build
```

WARNING: `docker-compose.yml` currently has both `migrate-mongodb` and `migrate-session-mongodb`; `migrate-mongodb` also loops over `session-service`. The existing migration scripts appear intended to be idempotent, but avoid repeatedly running both jobs against persistent local volumes unless you are intentionally reapplying migrations.

## Seed Data

Seed/demo user, seller, product, or superadmin credentials: Not found in codebase - please confirm.
