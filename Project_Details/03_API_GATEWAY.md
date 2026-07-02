# API Gateway

## Responsibility

The API Gateway is the edge service for REST and selected gRPC-Web browser traffic. It owns request routing, auth enforcement, CORS, request validation, rate limiting, access logging, metrics, tracing, and downstream service proxy/client wiring.

Location: `backend/services/api-gateway`.

## Runtime And Configuration

| Item | Value found |
| --- | --- |
| HTTP address | `HTTP_ADDR`, default/example `:8080` |
| gRPC-Web enabled | `GRPC_WEB_ENABLED=true` in `.env.example` |
| gRPC-Web address | `GRPC_ADDR`, example `:9090`, mapped to local host `8099` |
| Metrics address | `METRICS_ADDR`, example `:9091`, mapped to local host `19090` |
| API catalog | `API_CONTRACT_PATH`, compose uses `/app/api/master-api.json` |
| gRPC-Web policy | `backend/services/api-gateway/config/grpcweb-policies.json` |

## Routes Handled

The route catalog is `api/master-api.json`. It contains 97 REST endpoints in the current codebase.

| Service in catalog | Endpoint count |
| --- | ---: |
| `api-gateway-service` | 1 |
| `auth-service` | 8 |
| `user-service` | 8 |
| `product-service` | 6 |
| `cart-service` | 6 |
| `wishlist-service` | 4 |
| `search-service` | 6 |
| `session-service` | 14 |
| `cms-service` | 6 |
| `order-service` | 6 |
| `payment-service` | 3 |
| `notification-service` | 2 |
| `superadmin-service` | 27 |

Major REST route groups:

| Group | Routes found |
| --- | --- |
| Health | `/health/live`, `/health/ready` |
| Auth | `/api/v1/auth/login`, `/refresh`, `/logout`, `/otp/send`, `/otp/verify`, `/password/forgot`, `/password/reset`, `/signup` in catalog |
| User | `/api/v1/me`, `/api/v1/me/addresses`, `/api/v1/seller/me` |
| Seller session | `/api/v1/seller/session` |
| Product | `/api/v1/products`, `/api/v1/products/{product_id}`, `/api/v1/categories`, seller product create/update/publish |
| Search | `/api/v1/search`, `/api/v1/search/autocomplete`, admin synonym endpoints |
| Cart | `/api/v1/cart`, cart item add/update/delete, merge, coupon preview |
| Wishlist | `/api/v1/wishlist`, item add/delete, move to cart |
| Order | Checkout, buyer orders, order detail, cancel, seller orders, seller fulfillment |
| Payment | Retry, refund, provider webhook |
| CMS | Seller dashboard, coupons, campaigns |
| Session | Event ingestion, live sessions, journeys, funnels, heatmaps, cohorts, reports, privacy |
| Notification | Notification preferences get/update |
| Superadmin | Admin users, sellers, orders, payments, refunds, reconciliation, settings, audit, session analytics |

## Proxy And Service Mapping

Gateway routing is catalog-driven in `internal/transport/http/routes.go`.

| Downstream | Handling found |
| --- | --- |
| Auth | HTTP proxy when auth base URL is configured |
| User | Dedicated gRPC-backed user handlers |
| Product | HTTP proxy and special product handlers where configured |
| Search | Dedicated search handlers and gRPC-Web policy |
| Notification | Preference handlers |
| Order | Seller order handlers and downstream mapping |
| Session | Session proxy |
| Wishlist | Session-style HTTP proxy mapping |
| CMS | Session-style HTTP proxy mapping |
| Payment | HTTP proxy/internal token integration |
| Superadmin | Session-style HTTP proxy mapping |

## Middleware

Middleware found in the gateway includes:

- Request ID.
- Panic recovery.
- Access logging.
- CORS.
- Prometheus metrics.
- OpenTelemetry tracing.
- Optional request validation from API catalog schemas.
- Optional Redis-backed rate limiting.
- JWT auth and role authorization.

## Auth Handling

| Item | Implementation found |
| --- | --- |
| Token verification | JWT verifier with remote JWKS |
| JWKS URL | `JWT_JWKS_URL`, example `http://auth-service:8081/.well-known/jwks.json` |
| Issuer/audience | `JWT_ISSUER`, `JWT_AUDIENCE` |
| Route auth levels | Defined in `api/master-api.json` |
| Roles | Buyer/seller/admin/superadmin role groups from catalog |
| Webhooks | Catalog marks provider-signature auth requirement |

Catalog role groups include:

- Buyer routes: `buyer`, `seller`, `admin`, `superadmin`.
- Seller routes: `seller`, `seller_manager`, `seller_catalog_editor`, `seller_order_manager`, `superadmin`.
- Admin routes: `admin`, `operations_admin`, `finance_admin`, `catalog_admin`, `superadmin`.
- Superadmin routes: `superadmin`.

## Rate Limiting, CORS, Logging

| Capability | Found |
| --- | --- |
| CORS | Enabled/configurable with local frontend origins in `.env.example` |
| Rate limiting | Redis-backed pre-auth and post-auth configuration exists |
| Logging | Access logging middleware exists |
| Metrics | Prometheus metrics endpoint configured |
| Tracing | OpenTelemetry exporter configuration exists |

## gRPC-Web Facade

Configured policies include:

| Service/method | Auth policy |
| --- | --- |
| Search `Autocomplete` | Public |
| Recommendation `GetRecommendations` | Optional auth |
| Session `IngestEvent` | Optional auth |
| Session `GetLiveSessions` | Required admin/superadmin |
| Superadmin `ListUsers` | Operations admin/superadmin |
| Superadmin `UpdatePlatformSetting` | Superadmin |
| CMS `GetSellerAnalytics` | Seller/seller manager/superadmin |

## Environment Variables

Important variables in `backend/services/api-gateway/.env.example`:

- `HTTP_ADDR`
- `GRPC_WEB_ENABLED`
- `GRPC_ADDR`
- `METRICS_ADDR`
- `API_CONTRACT_PATH`
- `CORS_ALLOWED_ORIGINS`
- Downstream service URLs/addresses for auth, user, product, search, session, CMS, recommendation, superadmin, notification, payment, wishlist, order
- Redis rate-limit settings
- JWT issuer/audience/JWKS settings
- OpenTelemetry settings
- `PAYMENT_INTERNAL_API_TOKEN`

## Missing Or Mismatched Information

- `POST /api/v1/auth/signup` exists in the catalog and frontend auth client, but the auth service HTTP router does not register this route in the current codebase.
- Complete generated route documentation outside `api/master-api.json`: Not found in current codebase.
