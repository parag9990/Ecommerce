# Payment Service

## Responsibility

The payment service owns payment intents, payment state machine/schema endpoints, provider webhook handling, payment retry, refunds, refund review, payment administration, reconciliation records, and payment event publishing.

Location: `backend/services/payment-service`.

## Folder Structure

| Path | Responsibility |
| --- | --- |
| `cmd` | HTTP server and reconciliation command |
| `internal/config` | Environment configuration |
| `internal/domain` | Payment, refund, webhook, state-machine domain models |
| `internal/usecase` | Intent, webhook, refund, retry, reconciliation business logic |
| `internal/repository` | MySQL persistence |
| `internal/provider` | Payment provider integrations |
| `internal/transport` | HTTP handlers |
| `internal/events` | HTTP payment event publisher |
| `migrations` | MySQL schema migrations |

## APIs And Routes

Registered in `internal/transport/http/handler.go`.

| Method | Path | Purpose |
| --- | --- | --- |
| GET | `/healthz` | Health |
| GET | `/internal/v1/payment-state-machine` | State machine description |
| POST | `/internal/v1/payment-state-machine/validate` | Validate state transition |
| POST | `/internal/v1/payment-state-machine/provider-event/normalize` | Normalize provider event |
| GET | `/internal/v1/payment-schema` | Payment schema |
| GET | `/internal/v1/payment-schema/health` | Schema health |
| POST | `/internal/v1/payment-intents` | Create payment intent |
| POST | `/api/v1/webhooks/payments/{provider}` | Provider webhook |
| POST | `/api/v1/payments/{payment_id}/retry` | Retry payment |
| POST | `/api/v1/payments/{payment_id}/refund` | Request refund |
| GET | `/api/v1/refunds/{refund_id}` | Get refund |
| POST | `/internal/v1/refunds/{refund_id}/review` | Review refund |
| GET | `/internal/admin/payments` | Admin list payments |
| GET | `/internal/admin/payments/{payment_id}` | Admin payment detail |
| GET | `/internal/admin/payments/{payment_id}/refunds` | Admin payment refunds |
| GET | `/internal/admin/refunds` | Admin list refunds |
| GET | `/internal/admin/refunds/{refund_id}` | Admin refund detail |
| POST | `/internal/admin/refunds/{refund_id}/review` | Admin refund review |
| GET | `/internal/admin/payment-reconciliations` | Admin list reconciliations |
| GET | `/internal/admin/payment-reconciliations/{reconciliation_id}` | Admin reconciliation detail |

`/internal/v1/payment-intents` requires bearer `PAYMENT_INTERNAL_API_TOKEN`.

Buyer/admin APIs require bearer auth plus actor headers such as `X-Actor-ID` and role checks.

## Models And Tables

| Table | Purpose |
| --- | --- |
| `payments` | Payment record and provider references |
| `payment_attempts` | Attempt history |
| `refunds` | Refund records and review status |
| `payment_webhook_events` | Webhook idempotency/audit |
| `payment_reconciliations` | Settlement reconciliation records |

## Payment Providers

Provider interface supports:

- `CreateIntent`
- `Capture`
- `Refund`
- `VerifyWebhook`

Implementations found:

| Provider | Status |
| --- | --- |
| `stripe_like` | Create intent, refund, webhook verification implemented; capture returns unsupported |
| `razorpay_like` | Create intent, refund, webhook verification implemented; capture returns unsupported |

Providers are disabled by default in `.env.example` because `PAYMENT_ALLOWED_PROVIDERS` and `PAYMENT_DEFAULT_PROVIDER` are empty.

## Webhook Flow

1. Provider posts to `/api/v1/webhooks/payments/{provider}`.
2. Payment service verifies provider signature through configured provider.
3. Webhook event is recorded for idempotency/audit.
4. Payment/refund state is updated.
5. Payment event publisher sends configured event to `PAYMENT_EVENTS_ENDPOINT` when enabled.
6. Order service can receive payment events at `/internal/v1/payment-events`.

## Refund Flow

- Buyer/admin refund request route exists.
- Manual review threshold is configurable.
- Admin/internal refund review routes exist.
- Provider refund calls are implemented for provider integrations.

## Reconciliation

`cmd/reconciliation/main.go` runs reconciliation only when `PAYMENT_RECONCILIATION_ENABLED=true`. It reads settlement CSV input, writes reconciliation records, and can publish alerts/events.

## Dependencies

| Dependency | Usage |
| --- | --- |
| MySQL `payment_db` | Primary persistence |
| Provider APIs | Stripe-like/Razorpay-like integrations when configured |
| Order service | Receives payment events when event endpoint is configured |

## Environment Variables

Important variables in `backend/services/payment-service/.env.example`:

- `PAYMENT_HTTP_ADDR`
- `PAYMENT_MYSQL_DSN`
- `PAYMENT_ALLOWED_PROVIDERS`
- `PAYMENT_DEFAULT_PROVIDER`
- `PAYMENT_ALLOWED_CURRENCIES`
- `PAYMENT_CAPTURE_MODE`
- `PAYMENT_INTERNAL_API_TOKEN`
- Webhook size/timestamp tolerance settings
- `PAYMENT_EVENTS_ENDPOINT`
- `PAYMENT_EVENTS_AUTH_TOKEN`
- Refund/retry/reconciliation settings
- Commented provider settings for Stripe-like and Razorpay-like integrations

## Health Check And Run

| Item | Found |
| --- | --- |
| Compose service | `payment-service` |
| HTTP local port | `8091` mapped to container `8080` |
| Health | `GET /healthz` |

## Missing Or Limited Information

- Real provider credentials: Not found in current codebase.
- Capture support for included providers: Not found in current codebase; provider code returns unsupported.
- Providers enabled by default: Not found in current codebase.
- Payment events endpoint default: Not found in current codebase; `.env.example` leaves it empty.
