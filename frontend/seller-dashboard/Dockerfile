# syntax=docker/dockerfile:1.7
FROM node:24-alpine AS build
WORKDIR /workspace
RUN corepack enable && corepack prepare pnpm@11.5.0 --activate
COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml frontend/tsconfig.base.json ./
COPY frontend/seller-dashboard ./seller-dashboard
RUN --mount=type=cache,target=/root/.local/share/pnpm/store pnpm install --no-frozen-lockfile
ARG VITE_API_BASE_URL=http://localhost:8080
ARG VITE_LOGIN_URL=http://localhost:3000/login
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL VITE_LOGIN_URL=$VITE_LOGIN_URL
RUN pnpm --filter seller-dashboard build
FROM nginx:1.27-alpine
COPY frontend/docker/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /workspace/seller-dashboard/dist /usr/share/nginx/html
EXPOSE 8080
