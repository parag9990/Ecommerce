# Local Security and Config Audit

## Fixed in this pass

- Added Git ignore coverage for local env/secret files and Docker build context exclusions.
- Centralized safe local-only credentials in templates and Compose; no production secret is included.
- MongoDB, Redis, RabbitMQ, MySQL, and Typesense require local credentials/API keys.
- Auth Docker builds a throwaway RSA keypair instead of committing a private key.
- Added unique host ports and health checks for locally runnable services.
- Added an explicit gateway degraded-start flag; production default remains fail-fast.

## Pending Decision / Required Work

| Finding | Risk | State |
|---|---|---|
| Gateway contract routes return `501 ROUTE_BRIDGE_NOT_CONFIGURED` | Frontends cannot complete end-to-end flows through gateway | Pending implementation |
| Gateway expects gRPC for HTTP-only services | readiness stays red and RPCs fail | Pending architecture decision |
| Product business handlers have no network server | catalog/cart/search workflows are incomplete | Pending implementation |
| Order gRPC layer has no executable/concrete adapters | checkout/order flows are incomplete | Pending implementation |
| Session has conflicting runtime implementations and does not compile | all Session business APIs unavailable | Pending reconciliation |
| Session Analytics production TypeScript build fails | analytics UI image is excluded from default Compose startup | Pending dependency/type reconciliation |
| Product uses RabbitMQ while Wishlist product consumer uses Kafka | availability/price-drop events do not arrive | Pending architecture decision |
| Superadmin adapters point at services without compatible admin HTTP APIs | control actions are partial | Pending implementation |
| No shared demo user/product seeding | first-run data is empty | Pending developer-experience work |
| Full Kubernetes Deployments/PVCs absent | K8s cannot run full platform | Blocked by app readiness |

## Local exposure

Database and broker ports are intentionally published for debugging. Do not use this Compose file on a public host. Vite `VITE_*` variables are public and must never hold secrets.
