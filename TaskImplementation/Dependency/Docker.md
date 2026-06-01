# Docker Dependency - User App Frontend

## 1. What Is This Dependency?

Docker app ko container image me package karne ke liye use hota hai. Frontend case me typical flow hota hai: Node image me build, phir Nginx/static server se `dist/` serve.

Simple Hinglish: Docker se app same environment me build/run hota hai, local machine differences kam hote hain.

## 2. Why This Service Uses It

Project docs Docker deployment direction mention karte hain, but actual Dockerfile/compose file was not clearly found. For User App Frontend, Docker useful hoga:

- Production static build serve karne ke liye.
- CI build consistency ke liye.
- Local full-stack compose setup ke liye.

## 3. Required Or Optional

Current repo state me optional/planned. Actual frontend can run locally without Docker using pnpm.

Docker is missing as implementation artifact.

## 4. Where It Is Used In Project

| Path | Finding |
|------|---------|
| `docs/11-devops-external-services.md` | Frontend Dockerfile example exists in docs. |
| Actual `Dockerfile` | Not clearly found in project files. |
| Actual `docker-compose*.yml` | Not clearly found in project files. |
| `infra/docker/nginx.conf` | Referenced in docs, not clearly found. |

## 5. Installation Steps

Install Docker Desktop or Docker Engine.

Verify:

```bash
docker --version
docker compose version
```

## 6. Docker Setup, If Possible

Suggested Dockerfile based on docs:

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

Suggested command based on project structure:

```bash
docker build -f infra/docker/user-app.Dockerfile -t user-app-frontend .
```

Actual file is missing, so this command is only a suggested future command.

## 7. Local Setup Without Docker

Recommended current setup:

```bash
cd frontend
pnpm install
pnpm --filter user-app dev
```

## 8. Required Environment Variables

Frontend Docker build needs Vite env values at build time:

| Variable | Example |
|----------|---------|
| `VITE_API_BASE_URL` | `http://localhost:8080` |
| `VITE_GRPC_WEB_BASE_URL` | `http://localhost:8082` |
| `VITE_GRPC_WEB_TIMEOUT_MS` | `5000` |
| `VITE_APP_ENV` | `production` |
| `VITE_PAYMENT_PROVIDERS` | `stripe,razorpay` |

## 9. Start Commands

Current local start:

```bash
cd frontend
pnpm --filter user-app dev
```

Suggested Docker start after Dockerfile exists:

```bash
docker run --rm -p 8081:80 user-app-frontend
```

Suggested command based on project structure.

## 10. Verify Running Commands

Without Docker:

```bash
cd frontend
pnpm --filter user-app build
```

With Docker after Dockerfile exists:

```bash
curl http://localhost:8081
```

## 11. Common Errors And Fixes

| Error | Reason | Fix |
|-------|--------|-----|
| Dockerfile missing | Not implemented yet | Add frontend Dockerfile. |
| Compose service missing | Not implemented yet | Add `user-app` service to compose. |
| Env values wrong in built image | Vite env baked at build time | Pass correct build args/env during build. |
| SPA routes 404 on refresh | Nginx not configured for history fallback | Add Nginx fallback to `index.html`. |
| API calls fail from container | Browser still uses configured URL | Set public reachable API URL, not container-internal-only URL. |

## 12. Security Notes

- Do not bake secrets into frontend image. `VITE_` values are public.
- Use minimal runtime image like Nginx Alpine or distroless static server.
- Configure security headers in Nginx/CDN.
- Keep dependency install reproducible with `pnpm-lock.yaml`.
- Scan final image in CI.

## 13. Final Checklist

| Item | Status |
|------|--------|
| Docker docs studied | Completed |
| Actual frontend Dockerfile found | Missing |
| Actual compose service found | Missing |
| Suggested Docker setup documented | Completed |
| Security notes added | Completed |
