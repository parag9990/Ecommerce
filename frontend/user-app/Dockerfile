# syntax=docker/dockerfile:1.7
FROM node:24-alpine AS build
WORKDIR /workspace
RUN corepack enable && corepack prepare pnpm@11.5.0 --activate
COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml frontend/tsconfig.base.json ./
COPY frontend/packages ./packages
COPY frontend/user-app ./user-app
RUN --mount=type=cache,target=/root/.local/share/pnpm/store pnpm install --no-frozen-lockfile
ARG VITE_API_BASE_URL=http://localhost:8080
ARG VITE_GRPC_WEB_BASE_URL=http://localhost:8099
ARG VITE_APP_ENV=local
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL VITE_GRPC_WEB_BASE_URL=$VITE_GRPC_WEB_BASE_URL VITE_APP_ENV=$VITE_APP_ENV
RUN pnpm --filter user-app build
FROM nginx:1.27-alpine
COPY frontend/docker/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /workspace/user-app/dist /usr/share/nginx/html
EXPOSE 8080
