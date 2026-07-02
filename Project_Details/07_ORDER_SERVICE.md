# Order Service

## Responsibility

The order service owns checkout, order records, order item records, order status history, seller fulfillment updates, order idempotency, payment-event application, and order event outbox publishing.

Location: `backend/services/order-service`.

## Folder Structure

| Path | Responsibility |
| --- | --- |
| `cmd` | HTTP/gRPC/admin server wiring |
| `internal/config` | Environment configuration |
| `internal/domain` | Order aggregate, statuses, event domain models |
| `internal/usecase` | Checkout, query, cancellation, fulfillment, payment result logic |
| `internal/repository` | MySQL persistence |
| `internal/clients` | Cart, product, payment clients |
| `internal/transport` | HTTP/gRPC handlers |
| `internal/events` | Outbox/event publishing |
| `internal/authctx` | Actor/auth context handling |
| `migrations` | MySQL schema migrations |

## APIs And Routes

### gRPC Service

Defined in `proto/ecommerce/order/v1/order.proto`.

| Method | Purpose |
| --- | --- |
| `CreateOrder` | Checkout/create order |
| `GetOrder` | Get order detail |
| `ListOrders` | List buyer orders |
| `ListSellerOrders` | List seller orders |
| `CancelOrder` | Cancel order |
| `UpdateFulfillment` | Seller fulfillment update |

The gRPC handler enforces actor roles. Examples found include buyer roles for create/list, seller roles for seller order operations, and admin/logistics style roles for cancellations or fulfillment.

### HTTP Routes

Registered in `cmd/server/main.go` and `cmd/server/admin.go`.

| Method | Path | Purpose |
| --- | --- | --- |
| GET | `/healthz` | Liveness |
| GET | `/readyz` | Readiness |
| POST | `/internal/v1/payment-events` | Apply payment event to order |
| GET | `/internal/admin/orders` | Admin list orders |
| GET | `/internal/admin/orders/{order_id}` | Admin order detail |
| GET | `/internal/admin/orders/{order_id}/status-history` | Admin status history |

`/internal/v1/payment-events` is protected by bearer `ORDER_PAYMENT_EVENTS_TOKEN`. Admin routes require admin token and `X-Admin-Id`.

### Gateway REST Exposure

API catalog exposes:

- `POST /api/v1/orders/checkout`
- `GET /api/v1/orders`
- `GET /api/v1/orders/{order_id}`
- `POST /api/v1/orders/{order_id}/cancel`
- `GET /api/v1/seller/orders`
- `PATCH /api/v1/seller/orders/{order_id}/fulfillment`

## Models And Tables

| Table | Purpose |
| --- | --- |
| `orders` | Order header/state |
| `order_items` | Product line items |
| `order_status_history` | Status transition history |
| `shipments` | Shipment metadata |
| `order_idempotency_keys` | Checkout idempotency |
| `order_outbox_events` | Order event outbox |

## Important Business Logic

- Checkout orchestration across cart, product inventory, and payment intent creation.
- Inventory reservation tracking with reservation ID and expiry.
- Idempotent order creation.
- Buyer order listing/detail.
- Seller order listing and fulfillment updates.
- Order cancellation.
- Payment captured/failed event application.
- Outbox event creation for order lifecycle events.

## Dependencies

| Dependency | Usage |
| --- | --- |
| MySQL `order_db` | Primary persistence |
| Cart service | Reads cart during checkout |
| Product service | Validates products and reserves/releases/commits inventory |
| Payment service | Creates payment intents and receives payment event results |
| Kafka | Optional order event outbox publishing |

## Environment Variables

Important variables in `backend/services/order-service/.env.example`:

- `ORDER_HTTP_ADDR`
- `ORDER_GRPC_ADDR`
- `ORDER_MYSQL_DSN`
- `ORDER_ADMIN_TOKEN`
- Inventory/idempotency TTLs
- `ORDER_TRUSTED_CALLER_TOKEN`
- Page token signing key
- Kafka/outbox settings
- `ORDER_PAYMENT_EVENTS_TOKEN`
- Payment return URL and currency settings
- `ORDER_CART_BASE_URL`
- `ORDER_PRODUCT_BASE_URL`
- `ORDER_PRODUCT_SERVICE_TOKEN`
- `ORDER_PAYMENT_BASE_URL`
- `ORDER_PAYMENT_INTERNAL_TOKEN`

## Health Check And Run

| Item | Found |
| --- | --- |
| Compose service | `order-service` |
| HTTP local port | `8090` |
| gRPC local port | `9094` |
| Liveness | `GET /healthz` |
| Readiness | `GET /readyz` |

## Missing Information

- Direct public HTTP order routes in order service: Not found in current codebase. Public REST exposure is through API Gateway/catalog.
- Dedicated order-service runbook: Not found in current codebase.
