# User Service

## Responsibility

The user service owns buyer user profiles, addresses, seller profiles, seller KYC document metadata, seller/user status administration, and user-domain outbox events.

Location: `backend/services/user-service`.

## Folder Structure

| Path | Responsibility |
| --- | --- |
| `cmd` | gRPC server and admin HTTP server wiring |
| `internal/config` | Environment configuration |
| `internal/domain` | User, address, seller, KYC domain models |
| `internal/usecase` | Profile/address/seller/status business logic |
| `internal/repository` | MySQL persistence |
| `internal/transport` | gRPC and admin HTTP transports |
| `internal/events` | Outbox publishing |
| `internal/audit` | Audit metadata support |
| `migrations` | MySQL schema migrations |

## APIs And Routes

### gRPC Service

Defined in `proto/ecommerce/user/v1/user.proto`.

| Method | Purpose |
| --- | --- |
| `CreateUser` | Create user profile |
| `GetUser` | Fetch user profile |
| `UpdateUserProfile` | Update profile fields |
| `ListUserAddresses` | List addresses |
| `CreateAddress` | Add address |
| `UpdateAddress` | Update address |
| `DeleteAddress` | Delete address |
| `GetSellerProfile` | Fetch seller profile |
| `UpdateSellerProfile` | Update seller profile |
| `UpdateUserStatus` | Admin/user status update |
| `UpdateSellerStatus` | Admin/seller status update |
| `ReviewKYCDocument` | Review seller KYC document |

### Admin HTTP Routes

Registered in `cmd/server/admin.go`. Admin routes require bearer `USER_SERVICE_ADMIN_TOKEN`.

| Method | Path | Purpose |
| --- | --- | --- |
| GET | `/health/live` | Liveness |
| GET | `/health/ready` | Readiness |
| GET | `/metrics` | Metrics |
| GET | `/internal/admin/users` | List users |
| GET | `/internal/admin/users/{user_id}` | Get user |
| PATCH | `/internal/admin/users/{user_id}/status` | Update user status |
| GET | `/internal/admin/sellers` | List sellers |
| GET | `/internal/admin/sellers/{seller_id}` | Get seller |
| PATCH | `/internal/admin/sellers/{seller_id}/status` | Update seller status |

### Gateway REST Exposure

The API Gateway exposes buyer/seller REST routes and maps them to user service logic:

- `GET /api/v1/me`
- `PATCH /api/v1/me`
- `GET /api/v1/me/addresses`
- `POST /api/v1/me/addresses`
- `PATCH /api/v1/me/addresses/{address_id}`
- `DELETE /api/v1/me/addresses/{address_id}`
- `GET /api/v1/seller/me`
- `PATCH /api/v1/seller/me`

## Models And Entities

| Entity/table | Purpose |
| --- | --- |
| `users` | Buyer/user profile and status |
| `user_addresses` | Addresses owned by a user |
| `seller_profiles` | Seller identity/profile/status |
| `seller_kyc_documents` | Seller KYC metadata |
| `user_outbox_events` | User-domain events |

## Important Business Logic

- User profile creation and update.
- Address CRUD.
- Seller profile read/update.
- User and seller status changes through admin paths.
- KYC document review.
- Audit/status/deleted metadata support.
- Outbox event publication through RabbitMQ or Kafka when enabled.

## Dependencies

| Dependency | Usage |
| --- | --- |
| MySQL `user_db` | Primary persistence |
| RabbitMQ or Kafka | Optional outbox event publishing |
| API Gateway | REST exposure for frontend user/seller routes |

## Environment Variables

Important variables in `backend/services/user-service/.env.example`:

- `USER_SERVICE_GRPC_ADDRESS`
- `USER_SERVICE_HTTP_ADDRESS`
- `USER_SERVICE_ADMIN_TOKEN`
- `USER_SERVICE_MYSQL_DSN`
- `USER_SERVICE_PHONE_DEFAULT_REGION`
- Outbox/event provider settings
- RabbitMQ URL/exchange/routing-key settings
- Kafka broker/topic settings

## Health Check And Run

| Item | Found |
| --- | --- |
| Compose service | `user-service` |
| gRPC local port | `50052` |
| Admin/health local port | `9091` |
| Liveness | `GET /health/live` |
| Readiness | `GET /health/ready` |
| Metrics | `GET /metrics` |

## Missing Information

- Public direct HTTP API from user service: Not found in current codebase.
- Dedicated user-service runbook: Not found in current codebase.
