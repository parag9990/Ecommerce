# Frontend

## Frontend Apps

| App | Location | Purpose |
| --- | --- | --- |
| User app | `frontend/user-app` | Buyer shopping, auth, cart, wishlist, checkout, orders, account |
| Seller dashboard | `frontend/seller-dashboard` | Seller catalog, orders, offers, analytics, team/audit views |
| Session analytics dashboard | `frontend/session-analytics-dashboard` | Admin analytics for sessions, journeys, funnels, heatmaps, reports, privacy |
| Superadmin panel | `frontend/superadmin-panel` | Superadmin operations for users, sellers, orders, payments, sessions, settings, audit |
| Proto client package | `frontend/packages/proto-client` | Generated TypeScript protobuf/connect clients |

Workspace root: `frontend/package.json`.

## Shared Tooling

| Item | Found |
| --- | --- |
| Package manager | `pnpm@11.5.0` |
| Node version | `>=22.13.0` |
| Framework | React with Vite |
| Language | TypeScript |
| Testing | Vitest and Testing Library |
| Common libraries | React Router, React Query, Zustand, lucide-react |

Root scripts include recursive build, user app lint/typecheck/build/preview, seller dashboard dev/build/test/typecheck, analytics dashboard dev/build/test/typecheck, and superadmin dev/test/typecheck/build.

## User App

| Area | Details found |
| --- | --- |
| Framework | React 19, Vite 8, TypeScript |
| Routing | React Router 7 |
| Server state | React Query |
| Client state | Zustand |
| Forms | react-hook-form and zod |
| gRPC-Web | Connect clients from proto package |

Routes found in `frontend/user-app/src/routes/index.tsx`:

- Public browse/search/product/category/deals routes.
- Auth routes for login, signup, OTP, forgot password, reset password.
- Protected wishlist, cart, checkout, payment result.
- Protected account routes for profile, addresses, orders, order detail, notifications.

API integration:

- REST through `src/lib/http.ts`.
- Base URL from `VITE_API_BASE_URL`, default local API Gateway.
- Credentials included in requests.
- Optional bearer token from auth session helper.
- `X-Request-Source: user-app` header.
- gRPC-Web through `src/lib/grpc-client.ts` to `VITE_GRPC_WEB_BASE_URL`.

Env variables in `.env.example`:

- `VITE_API_BASE_URL`
- `VITE_GRPC_WEB_BASE_URL`
- `VITE_API_TIMEOUT_MS`
- `VITE_APP_ENV`
- `VITE_PAYMENT_PROVIDERS`

## Seller Dashboard

| Area | Details found |
| --- | --- |
| Framework | React 19, Vite 8, TypeScript |
| Routing | React Router 7 |
| Server state | React Query |
| Client state | Zustand |

Routes include:

- `/login`
- Protected `/seller`
- Products list/new/edit
- Orders list/detail
- Offers coupons/campaigns
- Analytics
- Team
- Audit
- Permission denied view

API integration:

- REST through `src/lib/http.ts`.
- Base URL from `VITE_API_BASE_URL`.
- Credentials included.
- Adds `x-client-app` and `x-request-id`.
- Seller session loaded from `GET /api/v1/seller/session`.

Env variables in `.env.example`:

- `VITE_API_BASE_URL`
- `VITE_API_TIMEOUT_MS`
- `VITE_SELLER_LOGIN_URL`

## Session Analytics Dashboard

| Area | Details found |
| --- | --- |
| Framework | React 19, Vite 5, TypeScript |
| Routing | React Router 6 |
| Server state | React Query |
| Charts | Recharts |

Routes include:

- Login.
- Protected analytics layout.
- Overview, live sessions, journeys, funnels, heatmaps, cohorts, reports, privacy.

API integration:

- REST analytics API client in `src/api/session-api.ts`.
- Session stored in `sessionStorage` key `session-analytics.admin-session`.
- Adds Authorization bearer token and `X-Request-ID`.
- Supports CSV/blob report downloads.

Env variables in `.env.example`:

- `VITE_API_BASE_URL`
- `VITE_PROXY_API_TARGET`
- `VITE_API_TIMEOUT_MS`

## Superadmin Panel

| Area | Details found |
| --- | --- |
| Framework | React 19, Vite 8, TypeScript |
| Routing | React Router 7 |
| Server state | React Query |
| Client state | Zustand |

Routes include:

- `/login`
- Protected `/admin`
- Users and user detail.
- Sellers and seller detail.
- Orders and order detail.
- Payments/refunds.
- Sessions.
- Search placeholder module.
- Settings.
- Audit logs.

API integration:

- REST through API Gateway.
- Session stored in `sessionStorage` key `superadmin.session`.
- Adds bearer token and `X-Request-ID`.

Env variables in `.env.example`:

- `VITE_API_BASE_URL`
- `VITE_APP_NAME`
- `VITE_ADMIN_SESSION_WARNING_MINUTES`

## Run Commands

From `frontend/package.json`:

| Command | Purpose |
| --- | --- |
| `pnpm dev` or `pnpm dev:user` | Run user app |
| `pnpm build` | Recursive build |
| `pnpm build:user` | Build user app |
| `pnpm lint:user` | Lint user app |
| `pnpm typecheck:user` | Typecheck user app |
| `pnpm dev:seller` | Run seller dashboard |
| `pnpm build:seller` | Build seller dashboard |
| `pnpm test:seller` | Test seller dashboard |
| `pnpm dev:analytics` | Run session analytics dashboard |
| `pnpm build:analytics` | Build session analytics dashboard |
| `pnpm dev:superadmin` | Run superadmin panel |
| `pnpm build:superadmin` | Build superadmin panel |

## Missing Or Mismatched Information

- User app calls `/api/v1/auth/signup`, but the auth service route was not found.
- Seller dashboard product listing uses `/api/v1/products` with seller filtering in client code, while product service also implements `/api/v1/seller/products`.
- Superadmin search page is a placeholder module in the current router.
