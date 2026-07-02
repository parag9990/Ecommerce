# Product Service

## Responsibility

The product service owns catalog products, categories, brands, seller product workflow, inventory reservations, product read APIs, product internal APIs, and product event outbox records.

Location: `backend/services/product-service`.

## Folder Structure

| Path | Responsibility |
| --- | --- |
| `cmd` | HTTP/gRPC server wiring and worker startup |
| `internal/app` | Application composition |
| `internal/config` | Environment configuration |
| `internal/domain` | Product, category, inventory, event domain models |
| `internal/usecase` | Catalog, seller workflow, inventory usecases |
| `internal/repository` | MongoDB persistence |
| `internal/transport` | HTTP/gRPC handlers |
| `internal/events` | Event outbox publishing |
| `internal/mapper` | Domain/transport mapping |
| `migrations` | MongoDB schema migrations |

## APIs And Routes

### HTTP Routes

Registered in `cmd/server/main.go`.

| Method | Path | Purpose |
| --- | --- | --- |
| GET | `/healthz` | Liveness |
| GET | `/readyz` | Readiness |
| GET | `/api/v1/products` | List public products |
| GET | `/api/v1/products/{product_id}` | Get public product detail |
| GET | `/api/v1/categories` | List categories |
| GET | `/api/v1/seller/products` | List seller products |
| GET | `/api/v1/seller/products/{product_id}` | Get seller product |
| POST | `/api/v1/seller/products` | Create seller product |
| PATCH | `/api/v1/seller/products/{product_id}` | Update seller product |
| POST | `/api/v1/seller/products/{product_id}/publish` | Submit/publish seller product |
| GET | `/internal/v1/products/{product_id}` | Internal product detail |
| POST | `/internal/v1/products/batch` | Internal batch get |
| GET | `/internal/v1/products/search-export` | Internal search export |
| POST | `/internal/v1/products/{product_id}/publish` | Internal publish |
| POST | `/internal/v1/products/{product_id}/unpublish` | Internal unpublish |
| PATCH | `/internal/v1/products/{product_id}/status` | Internal status update |
| POST | `/internal/v1/inventory/reservations` | Reserve inventory |
| POST | `/internal/v1/inventory/reservations/{reservation_id}/release` | Release reservation |
| POST | `/internal/v1/inventory/reservations/{reservation_id}/commit` | Commit reservation |

Internal `/internal/v1` routes are protected by `X-Service-Token`.

### gRPC Service

Defined in `proto/ecommerce/product/v1/product.proto`.

| Method | Purpose |
| --- | --- |
| `ListProducts` | List products |
| `GetProduct` | Get product |
| `ListCategories` | List categories |
| `CreateProduct` | Create seller product |
| `UpdateProduct` | Update product |
| `PublishProduct` | Publish product |
| `UnpublishProduct` | Unpublish product |
| `BatchGetProducts` | Batch get products |
| `ReserveInventory` | Reserve inventory |
| `ReleaseInventory` | Release reservation |
| `CommitInventory` | Commit reservation |

Internal gRPC methods require metadata `x-service-token`.

## Models And Collections

| Collection | Purpose |
| --- | --- |
| `products` | Product catalog documents |
| `categories` | Category hierarchy |
| `brands` | Brand metadata |
| `inventory_snapshots` | Inventory state snapshots |
| `price_books` | Price books |
| `inventory_reservations` | Inventory reserve/release/commit workflow |
| `product_event_outbox` | Product event publishing |

## Important Business Logic

- Public catalog reads.
- Seller product create/update/publish workflow.
- Product status transitions including publish/unpublish/status update.
- Inventory reservation with TTL and idempotency support.
- Search export endpoint for downstream indexing.
- Product event outbox for created/updated/published/unpublished/inventory-changed events.

## Dependencies

| Dependency | Usage |
| --- | --- |
| MongoDB `product_db` | Primary persistence |
| RabbitMQ or Kafka | Optional product event publishing |
| API Gateway | REST exposure |
| Order service | Calls internal product/inventory APIs |
| Search service | Uses search export path by configuration/design |

## Environment Variables

Important variables in `backend/services/product-service/.env.example`:

- `PRODUCT_HTTP_ADDR`
- `PRODUCT_GRPC_ADDR`
- `PRODUCT_MONGO_URI`
- `PRODUCT_MONGO_DATABASE`
- Currency and product validation limits
- CMS moderation settings
- Inventory reservation TTL/worker settings
- Product event outbox broker settings
- `PRODUCT_INTERNAL_SERVICE_TOKEN`

## Health Check And Run

| Item | Found |
| --- | --- |
| Compose service | `product-service` |
| HTTP local port | `8082` |
| gRPC local port | `9093` mapped to container `9092` |
| Liveness | `GET /healthz` |
| Readiness | `GET /readyz` |

## Missing Or Mismatched Information

- `api/master-api.json` lists seller product create/update/publish but does not list every seller product HTTP route implemented by product service.
- Dedicated product-service runbook: Not found in current codebase.
