# syntax=docker/dockerfile:1.7
FROM node:24-alpine AS build
WORKDIR /workspace
RUN corepack enable && corepack prepare pnpm@11.5.0 --activate
COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml frontend/tsconfig.base.json ./
COPY frontend/session-analytics-dashboard ./session-analytics-dashboard
RUN --mount=type=cache,target=/root/.local/share/pnpm/store pnpm install --no-frozen-lockfile
ARG VITE_API_BASE_URL=
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
RUN pnpm --filter @ecommerce/session-analytics-dashboard build
FROM nginxinc/nginx-unprivileged:1.27-alpine
COPY frontend/session-analytics-dashboard/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /workspace/session-analytics-dashboard/dist /usr/share/nginx/html
EXPOSE 8080
