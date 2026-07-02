# User App Frontend Local Runbook

## 1. Purpose

Buyer-facing app for authentication, product browsing, search, cart, wishlist, checkout, payments, and account flows.

## 2. Location

`frontend/user-app`

## 3. Tech Stack

React, TypeScript, Vite, React Router, React Query, Zustand, Connect/gRPC-Web client package, pnpm workspace.

## 4. Required Dependencies

API Gateway, gRPC-Web facade, auth/product/cart/wishlist/order/payment/search/session services through the gateway.

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

## 12. Quick Verification

Open `http://localhost:3000` for Docker or the Vite URL printed by the manual dev server.
