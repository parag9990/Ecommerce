# Frontend Dependency - Superadmin Panel

## 1. What is this dependency?

Frontend stack Superadmin Panel ka actual runnable UI layer hai. Ye React + TypeScript + Vite app hai jo browser me run hota hai.

## 2. Why this service uses it

Superadmin Panel admins ko secure routes, role-based menu, data tables, review actions, settings forms, and audit log views deta hai. UI backend ko REST APIs ke through call karta hai.

## 3. Required or optional

Required. Actual source code `frontend/superadmin-panel/` me found hua.

## 4. Where it is used in project

| Path | Use |
|------|-----|
| `frontend/superadmin-panel/package.json` | App scripts and dependencies |
| `frontend/superadmin-panel/src/app/router.tsx` | Admin routes |
| `frontend/superadmin-panel/src/lib/http.ts` | API client |
| `frontend/superadmin-panel/src/stores/auth-store.ts` | Admin session state |
| `frontend/superadmin-panel/src/config/admin-menu.ts` | Role-based menu |
| `frontend/superadmin-panel/tests/` | Vitest test suites |
| `frontend/package.json` | Workspace scripts |
| `frontend/pnpm-workspace.yaml` | Monorepo workspace registration |

## 5. Detected packages

| Package | Type | Why used |
|---------|------|----------|
| `react`, `react-dom` | Runtime | UI components render karne ke liye |
| `react-router-dom` | Runtime | `/admin/...` routes and protected routing |
| `@tanstack/react-query` | Runtime | API data fetching, caching, retry |
| `zustand` | Runtime | Admin auth/session state |
| `lucide-react` | Runtime | Sidebar/menu/icons |
| `vite` | Dev/build | Local dev server and build |
| `typescript` | Dev/build | Type checking |
| `tailwindcss`, `@tailwindcss/vite` | Dev/style | Utility CSS styling |
| `vitest`, `jsdom` | Test | Unit/component tests |
| `@testing-library/react`, `@testing-library/user-event` | Test | React component tests |

Note: Task guides mention `clsx` and `msw`, but these are not in actual `frontend/superadmin-panel/package.json`. Current code uses local `src/lib/classnames.ts` and tests use fetch mocks/Vitest helpers.

## 6. Installation steps

Actual command:

```bash
cd frontend
pnpm install
```

If pnpm is not available:

```bash
corepack enable
corepack prepare pnpm@latest --activate
```

## 7. Docker setup, if possible

No frontend Dockerfile found for Superadmin Panel.

Suggested command based on common Vite production setup:

```bash
pnpm --filter superadmin-panel build
```

After build, `dist/` can be served by Nginx or any static file server, but no repo Docker setup was clearly found.

## 8. Local setup without Docker

```bash
cd frontend
cp superadmin-panel/.env.example superadmin-panel/.env
pnpm install
pnpm --filter superadmin-panel dev
```

## 9. Required environment variables

| Variable | Required | Purpose |
|----------|----------|---------|
| `VITE_API_BASE_URL` | Yes | API Gateway base URL |
| `VITE_APP_NAME` | Optional | Display/application name |
| `VITE_ADMIN_SESSION_WARNING_MINUTES` | Optional | Session warning timing |

## 10. Start commands

```bash
cd frontend
pnpm run superadmin:dev
```

Direct filter command:

```bash
cd frontend
pnpm --filter superadmin-panel dev
```

## 11. Verify running commands

```bash
cd frontend
pnpm run superadmin:typecheck
pnpm run superadmin:test
pnpm run superadmin:build
```

Browser verify:

```text
http://localhost:5173/admin
```

Vite can choose another port if `5173` is busy.

## 12. Common errors and fixes

| Error | Reason | Fix |
|-------|--------|-----|
| `VITE_API_BASE_URL is not configured` | `.env` missing or variable absent | Copy `.env.example` to `.env` and set `VITE_API_BASE_URL` |
| Login works but admin APIs fail | API Gateway/backend not running | Start API Gateway and required backend services |
| 401/403 responses | Token missing or role not admin | Login with admin role: `superadmin`, `operations_admin`, `finance_admin`, `catalog_admin`, or `readonly_admin` |
| Package install fails | pnpm/Corepack not enabled | Run Corepack commands above |
| Tests fail due fetch/env | Test env not stubbing `VITE_API_BASE_URL` | Follow existing test setup patterns |

## 13. Security notes

- Frontend route guards are UX protection only. Real RBAC backend/API Gateway me enforce hona chahiye.
- `Authorization: Bearer <token>` header `src/lib/http.ts` se attach hota hai.
- Admin session local browser storage sensitive hota hai. XSS prevention important hai.
- Do not commit real `.env` secrets.
- Export/audit/payment flows me PII masking and reason capture mandatory honi chahiye.

## 14. Final checklist

- [x] Frontend package found.
- [x] Workspace registration found.
- [x] Env example found.
- [x] API client found.
- [x] Route and RBAC files found.
- [ ] Backend runtime available for real API calls.

