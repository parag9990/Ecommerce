# Environment Setup Guide

Compose reads committed `.env.example` files. Values are safe dummy local values, not production defaults.

| Component | Template | Important values |
|---|---|---|
| API Gateway | `backend/services/api-gateway/.env.example` | downstream addresses, JWKS, Redis, tracing |
| Auth | `backend/services/auth-service/.env.example` | MySQL, Redis, JWT key paths, token/OTP peppers |
| User, CMS, Order, Payment, Superadmin | each service `.env.example` | MySQL DSN and internal tokens |
| Product, Cart, Wishlist, Recommendation, Session, Notification | each service `.env.example` | Mongo URI, Redis/queue settings |
| Search | `backend/services/search-service/.env.example` | Typesense, RabbitMQ, Redis |
| Four frontend apps | `frontend/<app>/.env.example` | public API/gRPC-Web base URLs only |

## Docker host versus host machine

| Dependency | Inside Docker | App running on host |
|---|---|---|
| MySQL | `mysql:3306` | `localhost:3306` |
| MongoDB | `mongodb:27017` | `localhost:27017` |
| Redis | `redis:6379` | `localhost:6379` |
| RabbitMQ | `rabbitmq:5672` | `localhost:5672` |
| Kafka | `kafka:29092` | `localhost:9092` |
| Typesense | `typesense:8108` | `localhost:8108` |

For host execution, copy the template to `.env.local`, change hostnames, and load it with your shell/tool. `.env`, `.env.local`, and `.env.dev.local` are ignored by Git.

## Values you must change outside local dev

- Every password, pepper, encryption key, internal token, API key, and JWT key
- Payment provider keys and webhook secrets
- Public frontend URLs and CORS origins
- SMTP/provider endpoints

## Common mistakes

- Using `localhost` inside a container
- Using Kafka host port `9092` from containers instead of `kafka:29092`
- Putting secrets in Vite variables; every `VITE_*` value is public in browser assets
- Enabling a payment provider without a sandbox endpoint and 32+ character internal token
- Reusing local dummy secrets in shared/staging environments
