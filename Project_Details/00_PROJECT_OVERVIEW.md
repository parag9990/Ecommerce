# Project Overview

## Project Name

Scalable Ecommerce Platform.

Source: `api/master-api.json` declares `project: scalable-ecommerce-platform`.

## Purpose

This repository implements a multi-service ecommerce platform with buyer, seller, admin, analytics, and superadmin surfaces. The codebase includes service APIs for authentication, users, products, cart, wishlist, search, orders, payments, notifications, sessions, CMS, recommendations, and superadmin operations.

## Tech Stack

| Area | Technology found |
| --- | --- |
| Backend | Go workspace with service-specific Go modules |
| Internal APIs | gRPC and generated protobuf clients |
| Edge/API | REST API Gateway, gRPC-Web facade |
| Frontend | React, TypeScript, Vite, React Router, React Query, Zustand |
| Databases | MySQL, MongoDB |
| Cache/session/rate limit | Redis |
| Messaging/events | RabbitMQ, Kafka, outbox tables/collections |
| Search | Typesense |
| Observability | Prometheus, OpenTelemetry Collector, Jaeger |
| Local orchestration | Docker Compose, Tilt |
| Kubernetes | Kustomize manifests under `infra/k8s` and `deployments/k8s/local` |
| CI | GitHub Actions, Buf, golangci-lint, Vitest, Docker build, Trivy |

## Main Modules And Services

| Module | Location | Notes |
| --- | --- | --- |
| API Gateway | `backend/services/api-gateway` | REST gateway, gRPC-Web facade, auth/rate-limit/CORS/metrics/tracing middleware |
| Auth Service | `backend/services/auth-service` | Login, refresh, logout, OTP, password reset, JWT/JWKS, role assignment internals |
| User Service | `backend/services/user-service` | User profiles, addresses, seller profiles, seller KYC/status admin APIs |
| Product Service | `backend/services/product-service` | Product catalog, categories, seller product workflow, inventory reservations |
| Cart Service | `backend/services/cart-service` | Cart APIs exist in repo and API catalog; not requested as a dedicated document |
| Wishlist Service | `backend/services/wishlist-service` | Wishlist APIs exist in repo and API catalog; not requested as a dedicated document |
| Search Service | `backend/services/search-service` | Search/autocomplete/synonyms APIs exist in repo and API catalog |
| Session Service | `backend/services/session-service` | Session analytics and event ingestion APIs exist in repo and API catalog |
| CMS Service | `backend/services/cms-service` | Seller dashboard coupons/campaigns/analytics APIs exist in repo and API catalog |
| Order Service | `backend/services/order-service` | Checkout, orders, seller fulfillment, payment-event handling |
| Payment Service | `backend/services/payment-service` | Payment intents, provider webhooks, retry, refunds, reconciliation |
| Notification Service | `backend/services/notification-service` | OTP delivery, notification preferences, event-driven delivery processing |
| Recommendation Service | `backend/services/recommendation-service` | Recommendation APIs exist in repo and gRPC-Web policy |
| Superadmin Service | `backend/services/superadmin-service` | Admin users/sellers/orders/payments/settings/audit/session analytics |
| Shared backend packages | `backend/shared`, `backend/gen`, `backend/proto-gen` | Generated clients and common platform code |
| User frontend | `frontend/user-app` | Buyer shopping/account experience |
| Seller dashboard | `frontend/seller-dashboard` | Seller products, orders, offers, analytics |
| Session analytics dashboard | `frontend/session-analytics-dashboard` | Admin session analytics UI |
| Superadmin panel | `frontend/superadmin-panel` | Superadmin operations UI |
| Shared frontend package | `frontend/packages/proto-client` | Generated protobuf/connect TypeScript clients |

## High-Level Business Flow

1. Buyers authenticate through auth endpoints, then browse/search products through the API Gateway.
2. Buyers manage cart and wishlist, then create orders through the order checkout endpoint.
3. Order checkout reads cart data, reserves inventory through product service internals, creates an order, and creates a payment intent through the payment service.
4. Payment provider webhook/events update payment state and are forwarded to the order service to apply captured/failed outcomes.
5. Order status changes can commit or release inventory and publish order events.
6. Notification service consumes order/payment/user events and sends templated notifications through configured providers.
7. Sellers manage catalog items, orders, coupons/campaigns, and analytics through seller dashboard APIs.
8. Admin/superadmin surfaces expose operational views for users, sellers, orders, payments, refunds, settings, audit, and analytics.

## Repository Structure Summary

| Path | Responsibility |
| --- | --- |
| `.github/workflows` | CI workflow for Go, frontend, proto, and container checks |
| `api/master-api.json` | Main API catalog and route/auth metadata |
| `backend` | Go workspace, services, shared packages, generated protobuf Go clients |
| `database` | Shared database design notes and SQL drawing file |
| `deployments/k8s/local` | Local Kubernetes sample manifests for a partial stack |
| `docs` | Existing project documentation and runbooks |
| `frontend` | React/Vite workspaces and generated TypeScript proto client package |
| `infra` | Compose, Docker, Envoy, Kubernetes, MySQL init, observability, CI helper scripts |
| `proto` | Protobuf service definitions |
| `docker-compose.yml` | Full local platform composition |
| `Tiltfile` | Tilt orchestration over the compose stack |

## Notable Gaps In Overview

- Dedicated docs for cart, wishlist, search, session, CMS, recommendation, and superadmin services were not requested, even though those services exist.
- `docs/runbooks/LOCAL_PORTS.md` and `docs/runbooks/LOCAL_PLATFORM_RUNBOOK.md` were not found in the current codebase.
