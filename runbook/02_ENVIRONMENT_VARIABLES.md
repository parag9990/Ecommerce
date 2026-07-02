# Environment Variables

## Source Of Truth

Each service has a local `.env.example`. Docker Compose uses those files directly for most app services.

For manual host execution, copy the relevant `.env.example` to `.env` and replace Docker service names with `localhost`.

## Backend Env Files

| Service | Env file | Important groups |
| --- | --- | --- |
| API Gateway | `backend/services/api-gateway/.env.example` | HTTP, gRPC-Web, downstream URLs, Redis, JWT, tracing |
| Auth | `backend/services/auth-service/.env.example` | MySQL, Redis, JWT keys, peppers, notification gRPC |
| User | `backend/services/user-service/.env.example` | MySQL, admin token, RabbitMQ/Kafka outbox |
| Product | `backend/services/product-service/.env.example` | MongoDB, events, RabbitMQ/Kafka, internal token |
| Cart | `backend/services/cart-service/.env.example` | MongoDB, Redis, product/CMS URLs |
| Wishlist | `backend/services/wishlist-service/.env.example` | MongoDB, Kafka, product/cart URLs |
| Search | `backend/services/search-service/.env.example` | Typesense, product URL, Redis, RabbitMQ |
| Session | `backend/services/session-service/.env.example` | MongoDB, Redis, admin token, analytics limits |
| CMS | `backend/services/cms-service/.env.example` | MySQL, product URL, internal auth token |
| Recommendation | `backend/services/recommendation-service/.env.example` | MongoDB, Redis, Kafka |
| Order | `backend/services/order-service/.env.example` | MySQL, Kafka, cart/product/payment URLs |
| Payment | `backend/services/payment-service/.env.example` | MySQL, provider config, webhooks, reconciliation |
| Notification | `backend/services/notification-service/.env.example` | MongoDB, RabbitMQ, SMTP/Mailpit, encryption key |
| Superadmin | `backend/services/superadmin-service/.env.example` | MySQL, admin downstream URLs and tokens |

## Frontend Env Files

| App | Env file | Important variables |
| --- | --- | --- |
| User app | `frontend/user-app/.env.example` | `VITE_API_BASE_URL`, `VITE_GRPC_WEB_BASE_URL` |
| Seller dashboard | `frontend/seller-dashboard/.env.example` | `VITE_API_BASE_URL`, `VITE_LOGIN_URL` |
| Session analytics dashboard | `frontend/session-analytics-dashboard/.env.example` | `VITE_API_BASE_URL`, `VITE_API_PROXY_TARGET` |
| Superadmin panel | `frontend/superadmin-panel/.env.example` | `VITE_API_BASE_URL`, `VITE_APP_NAME` |

## Manual Host Execution Changes

When running Go services on the host instead of Docker, change examples like:

| Docker value | Host value |
| --- | --- |
| `mysql:3306` | `localhost:3306` |
| `mongodb:27017` | `localhost:27017` |
| `redis:6379` | `localhost:6379` |
| `rabbitmq:5672` | `localhost:5672` |
| `kafka:29092` | `localhost:9092` |
| `typesense:8108` | `localhost:8108` |
| `mailpit:1025` | `localhost:1025` |

## Warnings

- `PAYMENT_ALLOWED_PROVIDERS` and `PAYMENT_DEFAULT_PROVIDER` are empty by default.
- `PAYMENT_EVENTS_ENDPOINT` is empty by default, so payment events may not update orders until configured.
- `AUTH_EVENTS_PUBLISH_ENDPOINT` is empty by default, so auth/session event publishing needs explicit configuration.
- `CMS_INTERNAL_AUTH_TOKEN` is blank in the example.
- Auth JWT key paths point to container paths. Host execution needs local keys or alternate env.
- Seed/demo credentials: Not found in codebase - please confirm.
