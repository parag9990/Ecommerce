.PHONY: setup env infra-up infra-down backend-up frontend-up analytics-up analytics-build analytics-test docker-up docker-down docker-reset docker-logs session-retention-once test-go test-frontend test proto-lint proto-generate k8s-check k8s-up k8s-down tilt-up tilt-down

INFRA = mysql mongodb redis rabbitmq kafka typesense mailpit jaeger prometheus
BACKEND = auth-service user-service product-service cart-service wishlist-service search-service session-service session-retention-worker cms-service recommendation-service order-service payment-service notification-service superadmin-service api-gateway
FRONTEND = user-app seller-dashboard session-analytics-dashboard superadmin-panel

setup:
	docker compose config

env:
	@find backend/services frontend -name '.env.example' -print

infra-up:
	docker compose up -d $(INFRA)

infra-down:
	docker compose stop $(INFRA)

backend-up:
	docker compose up -d --build $(BACKEND)

frontend-up:
	docker compose up -d --build $(FRONTEND)

analytics-up:
	docker compose up -d --build session-analytics-dashboard

analytics-build:
	cd frontend && corepack pnpm --filter @ecommerce/session-analytics-dashboard build

analytics-test:
	cd frontend && corepack pnpm --filter @ecommerce/session-analytics-dashboard test

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

docker-reset:
	docker compose down -v

docker-logs:
	docker compose logs -f

session-retention-once:
	docker compose run --rm -e SESSION_RETENTION_WORKER_RUN_ONCE=true session-retention-worker

test-go:
	@set -e; for service in backend/services/* backend/shared/* backend/proto-gen/go; do if [ -f "$$service/go.mod" ]; then echo "==> $$service"; (cd "$$service" && go test ./...); fi; done

test-frontend:
	cd frontend && corepack pnpm install --frozen-lockfile && corepack pnpm --workspace-concurrency=1 -r --if-present test

test: test-go test-frontend

proto-lint:
	buf lint

proto-generate:
	buf generate

k8s-check:
	kubectl apply --dry-run=client -k deployments/k8s/local

k8s-up:
	kubectl apply -k deployments/k8s/local

k8s-down:
	kubectl delete namespace ecommerce-local --ignore-not-found

tilt-up:
	tilt up

tilt-down:
	tilt down
