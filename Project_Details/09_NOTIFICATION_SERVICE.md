# Notification Service

## Responsibility

The notification service owns OTP delivery, notification templates, delivery records, notification preferences, provider event analytics, provider webhook ingestion, retry/dead-letter support, and event-driven notifications.

Location: `backend/services/notification-service`.

## Folder Structure

| Path | Responsibility |
| --- | --- |
| `api` | Notification API-related files |
| `cmd` | gRPC and analytics/internal HTTP server wiring |
| `internal/config` | Environment configuration |
| `internal/domain` | Template, delivery, preference, provider event domain models |
| `internal/usecase` | OTP, preference, delivery, analytics usecases |
| `internal/repository` | MongoDB persistence |
| `internal/provider` | Email/SMS/provider integrations |
| `internal/events` | RabbitMQ event consumption |
| `internal/retry` | Retry/dead-letter handling |
| `internal/security` | Signing/encryption/protection helpers |
| `internal/transport` | gRPC and HTTP transports |
| `migrations` | MongoDB schema migrations and seed templates |

## APIs And Routes

### gRPC Service

Defined in `proto/ecommerce/notification/v1/notification.proto`.

| Method | Purpose |
| --- | --- |
| `SendOTP` | Send OTP notification |
| `GetNotificationPreference` | Read notification preferences |
| `UpdateNotificationPreference` | Update notification preferences |

Preference gRPC methods require metadata including `x-user-id` and allowed roles such as buyer/seller/admin/superadmin.

### HTTP/Internal Server

| Method | Path | Purpose |
| --- | --- | --- |
| GET | `/healthz` | Liveness |
| GET | `/readyz` | Readiness |
| GET | `/metrics` | Prometheus metrics when enabled |
| POST | `/internal/provider-webhooks/{channel}` | Provider webhook ingestion when configured |

Provider webhook handling requires channel signing secrets and headers such as `X-Notification-Signature` and `X-Notification-Timestamp`.

### Gateway REST Exposure

API catalog exposes notification preferences:

- `GET /api/v1/notifications/preferences`
- `PATCH /api/v1/notifications/preferences`

## Models And Collections

| Collection | Purpose |
| --- | --- |
| `notification_templates` | Templates by key/channel/version |
| `notification_deliveries` | Delivery attempts and status |
| `notification_preferences` | User preferences/suppression |
| `provider_events` | Provider delivery/open/failure events |

Seeded templates include:

- `otp_verification`
- `order_status_update`
- `payment_status_update`
- `promotional_offer`
- `welcome_user`
- `seller_approved`
- `address_updated_security_notice`

## Important Business Logic

- OTP delivery through gRPC.
- Template rendering.
- Email provider support through SMTP.
- Optional HTTP SMS provider support.
- Consent/preference gate.
- RabbitMQ event consumers for order, payment, and user events.
- Retry and dead-letter handling.
- Provider webhook analytics for sent/delivered/failed/opened states.

## Event Flow

When RabbitMQ is enabled, the service consumes configured queues:

| Queue | Purpose |
| --- | --- |
| `notification.order.events.v1` | Order-driven notifications |
| `notification.payment.events.v1` | Payment-driven notifications |
| `notification.user.events.v1` | User-driven notifications |

## Dependencies

| Dependency | Usage |
| --- | --- |
| MongoDB `notification_db` | Templates, deliveries, preferences, provider events |
| RabbitMQ | Event consumption and retry queues when enabled |
| SMTP/Mailpit | Email provider in local compose |
| HTTP SMS provider | Optional, disabled by default |
| Auth service | Calls notification gRPC for OTP delivery |

## Environment Variables

Important variables in `backend/services/notification-service/.env.example`:

- `NOTIFICATION_GRPC_ADDR`
- `NOTIFICATION_HTTP_ADDR`
- MongoDB URI/database/collection settings
- RabbitMQ URL and queue settings
- Delivery encryption key
- SMTP email settings
- SMS provider settings
- Push/WhatsApp contract flags
- Metrics path and enable flag
- Provider webhook signing settings

## Health Check And Run

| Item | Found |
| --- | --- |
| Compose service | `notification-service` |
| HTTP local port | `8092` mapped to container `8081` |
| gRPC local port | `50060` mapped to container `9090` |
| Liveness | `GET /healthz` |
| Readiness | `GET /readyz` |
| Metrics | `GET /metrics` when enabled |

## Missing Or Mismatched Information

- `api/master-api.json` lists notification gRPC methods beyond the current proto surface, but the current proto only defines `SendOTP`, `GetNotificationPreference`, and `UpdateNotificationPreference`.
- Push provider implementation: Not found in current codebase.
- WhatsApp provider implementation: Not found in current codebase.
