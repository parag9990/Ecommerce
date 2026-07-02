# Superadmin Panel Local Runbook

## 1. Purpose

Superadmin UI for users, sellers, orders, payments, refunds, reconciliation, settings, audit logs, and session analytics views.

## 2. Location

`frontend/superadmin-panel`

## 3. Tech Stack

React, TypeScript, Vite, React Router, React Query, Zustand, pnpm workspace.

## 4. Required Dependencies

API Gateway, superadmin service, user, order, payment, session, CMS, and auth flows.

## 5. Environment Variables

Use `frontend/superadmin-panel/.env.example`.

Key vars: `VITE_API_BASE_URL`, `VITE_APP_NAME`, `VITE_ADMIN_SESSION_WARNING_MINUTES`.

## 6. Install Dependencies

```powershell
cd frontend
corepack pnpm install --frozen-lockfile
```

## 7. Database/Migration/Seed Setup

No database owned by this app. Superadmin credentials: Not found in codebase - please confirm.

## 8. Run Command

Recommended:

```powershell
docker compose up -d --build superadmin-panel
```

Manual:

```powershell
cd frontend
corepack pnpm superadmin:dev -- --port 5176
```

The script does not pin a port; using a unique port avoids conflicts with user app.

## 9. Health Check

- Docker app URL: `http://localhost:3003`
- Docker image health: `http://localhost:3003/healthz`

## 10. Logs

```powershell
docker compose logs -f superadmin-panel
```

Manual Vite logs appear in the terminal.

## 11. Common Issues

- Superadmin credentials are missing.
- Superadmin service readiness fails if downstream admin services are unavailable.
- Gateway CORS does not include manual Vite origin.
- API calls fail if admin tokens are inconsistent across services.

## 12. Quick Verification

Open `http://localhost:3003` for Docker or the printed Vite URL for manual mode.
