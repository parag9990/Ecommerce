# Known Limitations

This file lists gaps and risk areas found from the current repository. No `TODO`, `FIXME`, or `HACK` markers were found under the main source/documentation directories checked.

## API And Contract Mismatches

| Area | Limitation |
| --- | --- |
| Auth signup | `POST /api/v1/auth/signup` exists in `api/master-api.json` and frontend user app, but the auth service HTTP router does not register the route. |
| Auth signup domain file | IDE context referenced `backend/services/auth-service/internal/domain/signup.go`, but that file was not found in the current filesystem. |
| Notification proto/catalog | API catalog references notification capabilities beyond the current `notification.proto`, which only defines `SendOTP`, `GetNotificationPreference`, and `UpdateNotificationPreference`. |
| Seller product routes | Product service implements seller product list/detail routes, while the API catalog does not list every implemented seller product route. |
| Seller dashboard product listing | Seller dashboard client uses `/api/v1/products` with seller filtering while product service also provides `/api/v1/seller/products`. |

## Payments

| Area | Limitation |
| --- | --- |
| Providers disabled by default | `PAYMENT_ALLOWED_PROVIDERS` and `PAYMENT_DEFAULT_PROVIDER` are empty in `.env.example`. |
| Capture support | Stripe-like and Razorpay-like provider code returns unsupported for capture. |
| Payment event publishing | `PAYMENT_EVENTS_ENDPOINT` is empty by default, so order payment-event application requires configuration. |
| Real credentials | Real provider credentials are not present, as expected for source control. |

## Auth And Sessions

| Area | Limitation |
| --- | --- |
| Auth event publishing | `AUTH_EVENTS_PUBLISH_ENDPOINT` is empty in `.env.example`; outbox publishing depends on configuration. |
| Public password reset route | Catalog lists password reset, but public handler registration was not confirmed in the visible auth router. |
| Seed users | Seed/demo credentials were not found. |

## Deployment

| Area | Limitation |
| --- | --- |
| Full production deployment | Not found in current codebase. |
| Helm charts | Not found in current codebase. |
| Kubernetes completeness | Local Kubernetes manifests are partial; Docker Compose is the complete local stack. |
| Kustomize base coverage | `infra/k8s/base` includes many services, but a single complete active overlay for every service was not found. |
| Secret management | Production secret management procedure was not found. |

## Documentation

| Area | Limitation |
| --- | --- |
| Local ports runbook | `docs/runbooks/LOCAL_PORTS.md` was not found. |
| Local platform runbook | `docs/runbooks/LOCAL_PLATFORM_RUNBOOK.md` was not found. |
| Incident runbooks | Not found in current codebase. |
| Service-specific docs for all services | Dedicated docs for cart, wishlist, search, session, CMS, recommendation, and superadmin were not requested here. |

## Frontend

| Area | Limitation |
| --- | --- |
| User signup flow | Frontend has signup route/API call, but backend signup route was not found. |
| Superadmin search | Superadmin panel search route is currently a placeholder module. |
| Auth token consistency | Apps use different session storage patterns; no single frontend auth architecture document was found. |

## Database And Operations

| Area | Limitation |
| --- | --- |
| ERD | Not found in current codebase. |
| Backup/restore | Not found in current codebase. |
| Retention policy | Not found in current codebase. |
| MySQL seeders | Dedicated MySQL seeder framework was not found. |

## Testing And Quality

| Area | Limitation |
| --- | --- |
| End-to-end tests | Not found in current codebase. |
| Contract drift checks for REST catalog vs service routers | Not found in current codebase. |
| Runtime smoke test script | Not found in current codebase. |

## Hardcoded Or Local-Only Values

- Local compose includes local ports, local service DNS names, and local credentials for development infrastructure.
- Payment provider values are placeholders/commented examples.
- Frontend `.env.example` files default to local API Gateway URLs.

## Suggested Risk Priority

1. Resolve the auth signup contract/backend mismatch.
2. Add contract tests that compare `api/master-api.json`, gateway route registration, service routers, and frontend clients.
3. Decide whether Docker Compose or Kubernetes is the canonical local platform and update runbooks accordingly.
4. Configure and test one payment provider end to end in a sandbox.
5. Add seed/demo data and a repeatable smoke test script.
