# User App Frontend Local Runbook

## 1. Purpose

Buyer-facing app for authentication, product browsing, search, cart, wishlist, checkout, payments, and account flows.

## 2. Location

`frontend/user-app`

## 3. Tech Stack

React, TypeScript, Vite, React Router, React Query, Zustand, Connect/gRPC-Web client package, pnpm workspace.

## 4. Required Dependencies

API Gateway, gRPC-Web facade, auth/product/cart/wishlist/order/payment/search/session services through the gateway.

Product browse and category pages read published products from Product Service through the gateway, so products created in Seller Dashboard/CMS become visible after they are saved as `published`. Search, popular products, autocomplete, and filtered result pages also depend on Search Service and its Typesense index.

Checkout depends on Cart Service, User Service addresses, Order Service, Product Service inventory, and Payment Service. The frontend sends the active `cart_id`, selected `address_id`, payment provider, and idempotency key to `POST /api/v1/orders/checkout` through API Gateway.

## 5. Environment Variables

Use `frontend/user-app/.env.example`.

Key vars: `VITE_API_BASE_URL`, `VITE_GRPC_WEB_BASE_URL`, `VITE_GRPC_WEB_TIMEOUT_MS`, `VITE_APP_ENV`, `VITE_PAYMENT_PROVIDERS`.

## 6. Install Dependencies

```powershell
cd frontend
corepack pnpm install --frozen-lockfile
```

## 7. Database/Migration/Seed Setup

No database owned by this app. Test users/products: Not found in codebase - please confirm.

## 8. Run Command

Recommended:

```powershell
docker compose up -d --build user-app
```

Manual:

```powershell
cd frontend
corepack pnpm dev:user
```

Manual Vite default port is `5173` unless overridden.

## 9. Health Check

- Docker app URL: `http://localhost:3000`
- Docker image health: `http://localhost:3000/healthz`

## 10. Logs

Docker:

```powershell
docker compose logs -f user-app
```

Manual Vite logs appear in the terminal running `pnpm dev:user`.

## 11. Common Issues

- `VITE_API_BASE_URL` points at the wrong gateway URL.
- Gateway CORS does not include the manual Vite origin.
- Signup flow may fail until auth signup route mismatch is resolved.
- Payment UI can show providers while backend providers are disabled.
- Popular products or filtered search fail if Search Service/Typesense is unhealthy or its product index has not caught up.
- Checkout fails if the gateway cannot reach Order Service over gRPC or User Service cannot return the selected delivery address.

## 12. Quick Verification

Open `http://localhost:3000` for Docker or the Vite URL printed by the manual dev server, then verify:

- Home page product grid loads published products.
- Search and category pages show product cards with images/prices.
- Deals page loads popular products.
- Wishlist items show hydrated product title, price, image, and move-to-cart works.
- Checkout review step can create an order/payment intent from a non-empty cart and saved address.
