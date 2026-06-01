# Frontend Stack Dependency - User App Frontend

## 1. What Is This Dependency?

Frontend stack ka matlab hai browser app build karne ke tools and libraries: Node.js, pnpm, Vite, React, React DOM, TypeScript, Tailwind CSS, React Router, and lucide-react.

Simple Hinglish: Ye stack app ko local run, build, type-check, route, style, and render karne me help karta hai.

## 2. Why This Service Uses It

`User App Frontend` buyer-facing SPA hai. Isliye:

- React UI components ke liye.
- React DOM browser me render ke liye.
- Vite fast dev server and production build ke liye.
- TypeScript type-safety ke liye.
- Tailwind CSS styling ke liye.
- React Router pages and protected routes ke liye.
- lucide-react icons ke liye.
- pnpm monorepo workspace manage karne ke liye.

## 3. Required Or Optional

| Dependency | Required? | Reason |
|------------|-----------|--------|
| Node.js | Required | Tooling runtime. |
| pnpm | Required | Workspace package manager. |
| Vite | Required | Dev server/build. |
| React + React DOM | Required | UI rendering. |
| TypeScript | Required | Build/typecheck. |
| Tailwind CSS | Required | Current app styling. |
| React Router | Required | Routes are implemented with it. |
| lucide-react | Optional but used | Icons in app shell/features. |

## 4. Where It Is Used In Project

| Area | Path |
|------|------|
| Root frontend package | `frontend/package.json` |
| User app package | `frontend/user-app/package.json` |
| Workspace config | `frontend/pnpm-workspace.yaml` |
| Vite config | `frontend/user-app/vite.config.ts` |
| App entry | `frontend/user-app/src/main.tsx` |
| Router | `frontend/user-app/src/routes/index.tsx` |
| Global styles | `frontend/user-app/src/styles/globals.css` |
| TypeScript config | `frontend/user-app/tsconfig*.json` |

## 5. Installation Steps

Exact dependency versions are already in `frontend/user-app/package.json` and locked in `frontend/pnpm-lock.yaml`.

```bash
cd frontend
corepack enable
pnpm install
```

If pnpm is not available:

```bash
corepack prepare pnpm@11.5.0 --activate
```

## 6. Docker Setup, If Possible

Actual Dockerfile was not clearly found in project files.

Suggested Docker direction based on `docs/11-devops-external-services.md`:

```dockerfile
FROM node:22-alpine AS builder
WORKDIR /repo/frontend
COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./
COPY frontend/user-app/package.json ./user-app/package.json
COPY frontend/packages/proto-client/package.json ./packages/proto-client/package.json
RUN corepack enable && pnpm install --frozen-lockfile
COPY frontend ./
RUN pnpm --filter user-app build

FROM nginx:alpine
COPY --from=builder /repo/frontend/user-app/dist /usr/share/nginx/html
```

Note: Nginx config path mentioned in docs was not clearly found in project files.

## 7. Local Setup Without Docker

```bash
cd frontend
pnpm install
pnpm --filter user-app dev
```

Browser URL is normally Vite default:

```text
http://localhost:5173
```

## 8. Required Environment Variables

Frontend stack itself does not need env vars to start, but the app needs API env values:

| Variable | Purpose |
|----------|---------|
| `VITE_API_BASE_URL` | REST API Gateway URL. |
| `VITE_GRPC_WEB_BASE_URL` | gRPC-Web bridge URL. |
| `VITE_GRPC_WEB_TIMEOUT_MS` | gRPC-Web timeout. |
| `VITE_APP_ENV` | App environment name. |
| `VITE_PAYMENT_PROVIDERS` | Checkout provider names. |

## 9. Start Commands

```bash
cd frontend
pnpm --filter user-app dev
```

Build command:

```bash
cd frontend
pnpm --filter user-app build
```

Preview command:

```bash
cd frontend
pnpm --filter user-app preview
```

## 10. Verify Running Commands

```bash
cd frontend
pnpm --filter user-app typecheck
pnpm --filter user-app lint
pnpm --filter user-app test
pnpm --filter user-app build
```

Manual browser check:

```text
Open http://localhost:5173
```

## 11. Common Errors And Fixes

| Error | Reason | Fix |
|-------|--------|-----|
| `pnpm: command not found` | pnpm not enabled | Run `corepack enable`. |
| Vite cannot resolve workspace package | dependencies not installed | Run `pnpm install` from `frontend/`. |
| API calls fail with network error | API Gateway not running | Start backend gateway or update `VITE_API_BASE_URL`. |
| gRPC-Web calls fail | bridge missing | Start Envoy/gateway bridge or update `VITE_GRPC_WEB_BASE_URL`. |
| TypeScript errors after package change | stale install | Run `pnpm install` and `pnpm --filter user-app typecheck`. |

## 12. Security Notes

- `VITE_` variables browser me expose hote hain, secrets mat rakho.
- Payment secret keys frontend me kabhi nahi aane chahiye.
- Auth token handling currently `sessionStorage` + cookie-backed requests pattern use karta hai. XSS risk reduce karne ke liye dependencies and CSP dhyan se maintain karo.
- Build artifacts `dist/` ko static server se serve karo; source maps production me policy ke hisaab se expose karo.

## 13. Final Checklist

| Item | Status |
|------|--------|
| Node/pnpm version documented | Completed |
| Vite/React/TypeScript documented | Completed |
| Tailwind/Router/icons documented | Completed |
| Local commands documented | Completed |
| Docker gap documented | Completed |
| Security notes added | Completed |
