# Seller Dashboard CMS Local Runbook

## 1. Purpose

Seller-facing dashboard for product management, order management, coupons/campaigns, and seller analytics.

## 2. Location

Actual app path: `frontend/seller-dashboard`

## 3. Tech Stack

React, TypeScript, Vite, React Router, React Query, Zustand, pnpm workspace.

## 4. Required Dependencies

API Gateway, auth, product, CMS, order, payment, and seller/user APIs through the gateway.

## 5. Environment Variables

Use `frontend/seller-dashboard/.env.example`.

Key vars: `VITE_API_BASE_URL`, `VITE_API_TIMEOUT_MS`, `VITE_LOGIN_URL`.

## 6. Install Dependencies

```powershell
cd frontend
corepack pnpm install --frozen-lockfile
```

## 7. Database/Migration/Seed Setup

No database owned by this app. Seller credentials and seed seller data: Not found in codebase - please confirm.

## 8. Run Command

Recommended:

```powershell
docker compose up -d --build seller-dashboard
```

Manual:

```powershell
cd frontend
corepack pnpm dev:seller
```

Manual Vite port is configured as `5174`.

## 9. Health Check

- Docker app URL: `http://localhost:3001`
- Docker image health: `http://localhost:3001/healthz`

## 10. Logs

```powershell
docker compose logs -f seller-dashboard
```

Manual Vite logs appear in the terminal.

## 11. Common Issues

- Login URL points to `http://localhost:3000/login`; adjust for manual user-app port.
- Seller credentials are missing.
- Gateway CORS does not include manual Vite origin.
- CMS/product APIs must be healthy for dashboard data.

## 12. Quick Verification

Open `http://localhost:3001` for Docker or `http://localhost:5174` for manual Vite.
