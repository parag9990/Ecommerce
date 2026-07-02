# Architecture

## Architecture Type

The repository implements a service-oriented ecommerce platform with:

- A REST API Gateway at the edge.
- gRPC between several backend services.
- gRPC-Web facade for selected browser-to-service APIs.
- Service-owned databases with no direct cross-service database foreign keys found outside a service boundary.
- Event-driven integration through outbox workers, RabbitMQ, Kafka, and HTTP event publishing.
- Separate frontend applications for buyer, seller, analytics admin, and superadmin workflows.

## Service Communication

| Communication | Found usage |
| --- | --- |
| Browser to API Gateway | REST endpoints under `/api/v1` |
| Browser to gRPC-Web facade | Configured in API Gateway with `config/grpcweb-policies.json` |
| API Gateway to services | HTTP proxy and gRPC clients, depending on route/service |
| Auth to Notification | gRPC client for OTP delivery |
| Order to Cart | HTTP client |
| Order to Product | HTTP internal APIs using product service token |
| Order to Payment | HTTP internal APIs using payment internal token |
| Payment to Order | HTTP event publisher to payment-event endpoint when configured |
| Services to queues | RabbitMQ/Kafka via service event/outbox workers |
| Observability | OTLP, Prometheus scrape endpoints, Jaeger export |

## Dependency Graph

```text
Frontend apps
  -> API Gateway
    -> Auth Service
    -> User Service
    -> Product Service
    -> Cart Service
    -> Wishlist Service
    -> Search Service
    -> Session Service
    -> CMS Service
    -> Order Service
    -> Payment Service
    -> Notification Service
    -> Superadmin Service

Auth Service
  -> MySQL auth_db
  -> Redis
  -> Notification Service gRPC
  -> Auth outbox/event endpoint when configured

User Service
  -> MySQL user_db
  -> RabbitMQ or Kafka outbox publisher when enabled

Product Service
  -> MongoDB product_db
  -> RabbitMQ or Kafka outbox publisher when enabled

Order Service
  -> MySQL order_db
  -> Cart Service
  -> Product Service
  -> Payment Service
  -> Kafka outbox publisher when enabled

Payment Service
  -> MySQL payment_db
  -> Stripe-like or Razorpay-like provider when configured
  -> Order Service payment event endpoint when configured

Notification Service
  -> MongoDB notification_db
  -> RabbitMQ event queues when enabled
  -> SMTP email provider when enabled
  -> HTTP SMS provider when enabled
```

## Request Flow

### Public REST Request

1. Frontend sends a REST request to API Gateway.
2. Gateway applies request ID, recovery, CORS, access logging, metrics, tracing, optional validation, rate limiting, and auth/role checks from the API catalog.
3. Gateway dispatches to a special handler, HTTP proxy, or gRPC-backed client.
4. Downstream service performs domain logic and persistence.
5. Gateway returns the service response to the frontend.

### Checkout Flow

1. Buyer creates checkout through order API.
2. Order service reads cart data through cart client.
3. Order service validates and reserves inventory through product service internal APIs.
4. Order service creates order rows and creates a payment intent through payment service.
5. Payment service handles provider webhook or retry/refund flows when providers are configured.
6. Payment event handling updates order status and inventory reservation state.
7. Order and notification events are published through configured event paths.

### OTP Flow

1. Auth service receives OTP send/verify request.
2. Auth service stores hashed OTP challenge and rate-limit state.
3. Auth service sends OTP through notification service gRPC.
4. Notification service renders a template and sends through configured provider.

## Folder And Module Responsibilities

| Folder pattern | Responsibility |
| --- | --- |
| `cmd` | Service entrypoints and server wiring |
| `internal/config` | Environment parsing and defaults |
| `internal/domain` | Domain types, validation concepts, state machines |
| `internal/usecase` | Application business logic |
| `internal/repository` | Persistence implementations |
| `internal/transport` | HTTP/gRPC handlers and middleware |
| `internal/events` | Outbox, event publishing, event consumption |
| `migrations` | SQL or MongoDB migrations owned by a service |
| `proto/ecommerce/*` | Protobuf service contracts |
| `frontend/*/src/routes` | Frontend route definitions |
| `frontend/*/src/lib` | API clients, HTTP clients, generated client wiring |
| `infra/*` | Local/CI/deployment/observability infrastructure |

## Important Design Decisions Found

- API behavior is catalog-driven through `api/master-api.json`.
- Gateway authorization maps routes to auth levels and role requirements from the catalog.
- JWT verification at the gateway uses remote JWKS from the auth service.
- Backend services own their own migrations and persistence schemas.
- Product, notification, cart, wishlist, recommendation, and session use MongoDB-backed migrations.
- Auth, user, order, payment, CMS, and superadmin use MySQL schemas in the local compose stack.
- Payment provider integrations are pluggable and disabled by default in local env examples.
- Several services use outbox tables/collections for event publication.
- Docker Compose is the most complete local runtime path in the current codebase.

## Missing Or Partial Architecture Information

- Full sequence diagrams: Not found in current codebase.
- Production deployment topology: Not found in current codebase.
- Complete Kubernetes coverage for every service: Not found in current codebase.
- Dedicated architectural decision records: Not found in current codebase.
