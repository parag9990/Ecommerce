# Search Service - Task 2 Dependency and Setup Guide

Source implementation file:

```text
TaskImplementation/Search Service/task2.md
```

Generated dependency file:

```text
TaskImplementation/Search Service/task2_Dependency.md
```

## 0. Scope and Reuse Rule

Task 2 ka goal hai Typesense ko local Docker Compose aur Kubernetes me run-ready banana. Ye file beginner developer ko batati hai ki Typesense runtime, ports, secrets, volumes, health checks, aur cluster setup kaise samajhna hai.

Previous dependency documentation already exists:

```text
TaskImplementation/Search Service/task1_Dependency.md
```

Is file me repeated setup duplicate nahi kiya gaya. Jahan same Go, Redis, RabbitMQ, full `.env`, common run commands, ya broad troubleshooting already documented hai, wahan direct reference diya gaya hai.

## 1. Previous Dependency File Reuse Map

| Topic | Reuse Decision |
|---|---|
| Go installation, Go modules, `go mod download`, `go test` | Refer `TaskImplementation/Search Service/task1_Dependency.md`, Section `2. Go Dependency System` |
| Full Search Service local `.env` example | Refer `TaskImplementation/Search Service/task1_Dependency.md`, Section `4. Environment Variables` |
| Redis setup | Refer `TaskImplementation/Search Service/task1_Dependency.md`, Section `3.2 Redis` and `5.2 Redis` |
| RabbitMQ setup | Refer `TaskImplementation/Search Service/task1_Dependency.md`, Section `5.3 RabbitMQ` |
| Product Service and Session Service dependency explanation | Refer `TaskImplementation/Search Service/task1_Dependency.md`, Sections `5.4 Product Service` and `5.5 Session Service` |
| Generic Docker Compose startup for all infra | Refer `TaskImplementation/Search Service/task1_Dependency.md`, Section `7. Docker and DevOps Setup` |
| Full Search Service run flow | Refer `TaskImplementation/Search Service/task1_Dependency.md`, Section `8. Project Run Instructions` |
| Generic common errors and security audit | Refer `TaskImplementation/Search Service/task1_Dependency.md`, Sections `9. Common Errors and Fixes` and `10. Security and Configuration Audit` |

Only Task 2 specific Typesense runtime details are explained below.

## 2. Project Tech Stack Analysis

### Reused Technology Summary

| Technology | Required? | Status in This File |
|---|---:|---|
| Go | Yes | Already explained in `task1_Dependency.md` |
| Go modules | Yes | Already explained in `task1_Dependency.md` |
| `net/http` Search Service server | Yes | Already explained in `task1_Dependency.md` |
| Redis | Yes for full service | Already explained in `task1_Dependency.md` |
| RabbitMQ | Yes by default | Already explained in `task1_Dependency.md` |
| Product Service | Yes for real search/reindex | Already explained in `task1_Dependency.md` |
| Session Service | Optional, enabled by default | Already explained in `task1_Dependency.md` |

### Task 2 Focus Technology

| Technology | Required? | Why Used | Beginner Explanation |
|---|---:|---|---|
| Typesense | Yes | Product search index run karne ke liye | Typesense ek fast search engine hai. Simple Hinglish me, ye products ko search-friendly format me store karta hai taaki typo tolerance, filters, facets, sorting, aur autocomplete fast ho sake. |
| Docker Compose | Recommended for local | Local Typesense container run karne ke liye | Docker Compose se local machine par Typesense one command me start ho jata hai. |
| Kubernetes StatefulSet | Optional for cluster/prod-like setup | 3-node durable Typesense cluster ke liye | StatefulSet pods ko stable naam aur storage deta hai. Typesense cluster ko ye stability chahiye. |
| Kubernetes Secret | Yes in K8s | Typesense API key store karne ke liye | Secret me sensitive values rakhi jati hain. Real API key ConfigMap ya git me nahi honi chahiye. |
| Kubernetes ConfigMap | Yes in K8s | Typesense nodes file aur Search Service non-secret config ke liye | ConfigMap non-secret runtime config store karta hai, jaise service DNS, ports, collection names. |
| Persistent volume/PVC | Yes in K8s | Typesense index data durable rakhne ke liye | Pod restart ke baad bhi search index data disk par safe rahe, isliye PVC use hota hai. |

## 3. Language-Specific Dependency System

Task 2 ne koi new Go library add nahi ki.

Refer:

```text
TaskImplementation/Search Service/task1_Dependency.md
```

Section:

```text
2. Go Dependency System
```

Important confirmation from current code:

| File | Current Detail |
|---|---|
| `backend/services/search-service/go.mod` | Go module for Search Service |
| Go version | `go 1.26.3` |
| Typesense library | `github.com/typesense/typesense-go/v2` |
| RabbitMQ library | `github.com/rabbitmq/amqp091-go` |

Task 2 setup work is infra/runtime focused. Use the same Go commands already documented in Task 1 when you need to build or run the Search Service.

## 4. Database and Storage Analysis

Search Service does not use MySQL, PostgreSQL, MongoDB, SQLite, or Cassandra directly for Task 2.

| Storage | Directly Used? | Required? | Task 2 Role |
|---|---:|---:|---|
| Typesense | Yes | Yes | Search index runtime |
| Docker named volume `typesense_data` | Local only | Yes for persistence | Persists local Typesense `/data` |
| Kubernetes PVC `data` | K8s only | Yes for persistence | Persists each Typesense pod's `/data` |
| Redis | Yes in full service | Reused | Already covered in Task 1 |
| RabbitMQ | Yes by default | Reused | Already covered in Task 1 |

### 4.1 Typesense

#### A. What it is

Typesense ek search database/search engine hai. Beginner terms me: product data ka ek search optimized copy yahan rakha jata hai. User jab search karega, API directly main product database scan nahi karegi. Search Service Typesense se fast matching product IDs/search documents nikaalega.

#### B. Why this project uses it

Task 1 me Search Service schema define hua tha. Task 2 me us schema ko run karne ke liye Typesense runtime ready hota hai.

Current repo expects these collection names:

| Collection | Purpose |
|---|---|
| `products` | Product search documents |
| `popular_queries` | Autocomplete/popular query support |

#### C. Required or optional

Required. `TYPESENSE_API_KEY` empty hua to Search Service config validation fail karegi, and Typesense admin calls unauthorized ho jayenge.

#### D. Local installation

Local install steps are already documented in:

```text
TaskImplementation/Search Service/task1_Dependency.md
```

Section:

```text
3.1 Typesense
```

Task 2 specific local setup is the actual Compose service in:

```text
infra/compose/docker-compose.local.yml
```

#### E. Docker setup

Current project Compose service:

```yaml
typesense:
  image: typesense/typesense:27.1
  container_name: ecommerce-typesense
  restart: unless-stopped
  command:
    - "--data-dir=/data"
    - "--api-key=${TYPESENSE_API_KEY:?TYPESENSE_API_KEY is required}"
    - "--enable-cors"
  ports:
    - "${TYPESENSE_PORT:-8108}:8108"
  volumes:
    - typesense_data:/data
```

Hinglish explanation:

| Config | Meaning |
|---|---|
| `typesense/typesense:27.1` | Pinned image. Sab developers same Typesense version use karenge. |
| `--data-dir=/data` | Index data container ke `/data` folder me store hota hai. |
| `--api-key=${TYPESENSE_API_KEY...}` | API calls ke liye key mandatory hai. Empty key par Compose fail fast karega. |
| `--enable-cors` | Local browser/debugging ke liye helpful. Production me restrict/disable karna better hai. |
| `${TYPESENSE_PORT:-8108}:8108` | Host machine par default port `8108` expose hota hai. |
| `typesense_data:/data` | Container delete/recreate ke baad bhi local index data bachta hai. |

#### F. Start commands

Only Typesense:

```bash
cp infra/compose/.env.local.example infra/compose/.env.local
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml up -d typesense
```

All local infra services:

```bash
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml up -d
```

#### G. Verify running

Health:

```bash
curl http://localhost:8108/health
```

Expected:

```json
{"ok":true}
```

Collection list:

```bash
curl -H "X-TYPESENSE-API-KEY: dev-typesense-key" http://localhost:8108/collections
```

Fresh setup me response empty list ho sakta hai. Collections Search Service startup/schema ensure flow ke baad create honge.

#### H. Default port

| Port | Purpose | Status |
|---:|---|---|
| `8108` | Typesense HTTP API | Required for local and K8s |
| `8107` | Typesense peering | Required only for K8s cluster |

#### I. Connection format

Search Service supports either full URL:

```env
TYPESENSE_URL=http://localhost:8108
```

or split host/port/protocol:

```env
TYPESENSE_HOST=localhost
TYPESENSE_PORT=8108
TYPESENSE_PROTOCOL=http
```

Rule:

| Run Location | Correct Host |
|---|---|
| Search Service runs on host with `go run` | `localhost` |
| Search Service runs inside Docker Compose network | `typesense` |
| Search Service runs inside Kubernetes | `typesense.data.svc.cluster.local` |

#### J. Where to place credentials

| Environment | File/Place | Key |
|---|---|---|
| Local Compose | `infra/compose/.env.local` | `TYPESENSE_API_KEY` |
| Local Go run | `backend/services/search-service/.env` or shell env | `TYPESENSE_API_KEY` |
| Typesense K8s data namespace | `infra/k8s/data/typesense/secret.example.yaml` shape, real secret from secret manager | `TYPESENSE_API_KEY` |
| Search Service K8s core namespace | `infra/k8s/core/search-service/secret.example.yaml` shape, real secret from secret manager | `TYPESENSE_API_KEY` |

Never commit real API keys.

## 5. Environment Variables

### No new Search Service variables

Task 2 does not introduce new Search Service environment variable names beyond the Typesense variables already documented in Task 1.

Refer:

```text
TaskImplementation/Search Service/task1_Dependency.md
```

Section:

```text
4. Environment Variables
```

### Task 2 specific local Typesense values

For `infra/compose/.env.local`:

```env
TYPESENSE_URL=
TYPESENSE_HOST=typesense
TYPESENSE_PORT=8108
TYPESENSE_PROTOCOL=http
TYPESENSE_API_KEY=dev-typesense-key
TYPESENSE_PRODUCTS_COLLECTION=products
TYPESENSE_POPULAR_QUERIES_COLLECTION=popular_queries
TYPESENSE_TIMEOUT_MS=300
```

For host-machine `go run`, use:

```env
TYPESENSE_HOST=localhost
TYPESENSE_PORT=8108
TYPESENSE_PROTOCOL=http
TYPESENSE_API_KEY=dev-typesense-key
```

### Variable explanation

| Variable | Required? | Purpose | Example | Security Notes |
|---|---:|---|---|---|
| `TYPESENSE_URL` | Optional | Full URL override | `http://localhost:8108` | Use HTTPS outside local if infrastructure supports it |
| `TYPESENSE_HOST` | Required if URL empty | Hostname without protocol | `localhost`, `typesense`, `typesense.data.svc.cluster.local` | Do not put secrets here |
| `TYPESENSE_PORT` | Required if URL empty | Typesense API port | `8108` | Port only, no hostname |
| `TYPESENSE_PROTOCOL` | Required if URL empty | `http` or `https` | `http` | Prefer `https` outside trusted networks |
| `TYPESENSE_API_KEY` | Yes | Auth key for Typesense API | `dev-typesense-key` | Secret. Never commit real value |
| `TYPESENSE_PRODUCTS_COLLECTION` | Optional | Product collection/alias name | `products` | Keep stable because API behavior depends on it |
| `TYPESENSE_POPULAR_QUERIES_COLLECTION` | Optional | Popular query collection name | `popular_queries` | No secret |
| `TYPESENSE_TIMEOUT_MS` | Optional | Request timeout for Typesense calls | `300` | Too low can cause flaky search |

### How the project loads env vars

The Go config uses `os.Getenv`. It does not auto-load `.env`.

For local Go run:

```bash
cd backend/services/search-service
set -a
. ./.env
set +a
go run ./cmd/server
```

Common mistakes:

| Mistake | Result | Fix |
|---|---|---|
| `TYPESENSE_API_KEY` differs between Compose and Search Service | Search Service gets `401 Unauthorized` from Typesense | Use same key in both places |
| `TYPESENSE_HOST=http://localhost` | Config validation fails | Use `TYPESENSE_HOST=localhost` and `TYPESENSE_PROTOCOL=http` |
| Using `typesense` as host from normal shell | Host machine cannot resolve Docker service DNS | Use `localhost` for host `go run` |
| Real API key added to `.env.local.example` | Secret leak risk | Keep only sample values in example files |

## 6. External Services Analysis

### 6.1 Typesense

| Item | Detail |
|---|---|
| What | Search engine |
| Why used | Product search index, facets, filters, sorting, typo tolerance |
| Mandatory | Yes |
| Local image | `typesense/typesense:27.1` |
| Local container | `ecommerce-typesense` |
| Health check | `curl http://localhost:8108/health` |
| Auth header | `X-TYPESENSE-API-KEY` |
| Data path | `/data` |

### 6.2 Docker

Docker explanation and installation is already covered in Task 1. Task 2 only needs Docker for the local Typesense container.

Task-specific commands:

```bash
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml ps typesense
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml logs -f typesense
```

### 6.3 Kubernetes

Task 2 introduces/uses these Typesense Kubernetes manifests:

| File | Purpose |
|---|---|
| `infra/k8s/data/typesense/namespace.yaml` | Creates `data` namespace |
| `infra/k8s/data/typesense/secret.example.yaml` | Shows required secret key shape |
| `infra/k8s/data/typesense/configmap.yaml` | Provides Typesense cluster `nodes` file |
| `infra/k8s/data/typesense/service-headless.yaml` | Stable pod DNS for cluster peering |
| `infra/k8s/data/typesense/service.yaml` | Client ClusterIP service for Search Service |
| `infra/k8s/data/typesense/statefulset.yaml` | 3-node Typesense cluster with PVCs and probes |
| `infra/k8s/data/typesense/pdb.yaml` | Keeps at least 2 pods available during voluntary disruption |

K8s is optional for local development, but required for production-like cluster deployment.

### 6.4 Redis and RabbitMQ

No new Redis or RabbitMQ setup is introduced by Task 2.

Refer:

```text
TaskImplementation/Search Service/task1_Dependency.md
```

Sections:

```text
5.2 Redis
5.3 RabbitMQ
```

## 7. Ports and Networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Search Service HTTP API | `8085` | Main Search API | Reused from Task 1 |
| Typesense API | `8108` | Search engine HTTP API | Required for Task 2 |
| Typesense peering | `8107` | K8s node-to-node cluster peering | Task 2 K8s specific |
| Redis | `6379` | Cache, locks, dedupe | Reused from Task 1 |
| RabbitMQ AMQP | `5672` | Product event consumer | Reused from Task 1 |
| RabbitMQ Management UI | `15672` | RabbitMQ browser/admin API | Reused from Task 1 |
| Product Service | `8082` | Product hydration/export | Reused from Task 1 |
| Session Service | `8086` | Zero-result analytics ingest | Reused from Task 1 |

### Local networking

| Caller | Typesense Host | Why |
|---|---|---|
| Browser/curl on host | `localhost:8108` | Compose publishes host port |
| Search Service via host `go run` | `localhost:8108` | Host shell cannot resolve Docker service name |
| Search Service inside Compose network | `typesense:8108` | Docker DNS resolves service name |

### Kubernetes networking

| DNS | Purpose |
|---|---|
| `typesense.data.svc.cluster.local:8108` | Search Service client traffic |
| `typesense-0.typesense-headless.data.svc.cluster.local` | Stable pod DNS for node 0 |
| `typesense-1.typesense-headless.data.svc.cluster.local` | Stable pod DNS for node 1 |
| `typesense-2.typesense-headless.data.svc.cluster.local` | Stable pod DNS for node 2 |

If `8108` conflicts locally:

```env
TYPESENSE_PORT=8118
```

Then host health check changes to:

```bash
curl http://localhost:8118/health
```

Inside Docker/K8s, container port remains `8108`.

## 8. Docker and DevOps Setup

### 8.1 Local Docker Compose

Existing local infra file:

```text
infra/compose/docker-compose.local.yml
```

Task 2 Typesense settings:

| Setting | Value |
|---|---|
| Image | `typesense/typesense:27.1` |
| Container name | `ecommerce-typesense` |
| Restart policy | `unless-stopped` |
| Data volume | `typesense_data:/data` |
| Network | `ecommerce-local` |
| Healthcheck | `/health` |

Validate Compose config:

```bash
cp infra/compose/.env.local.example infra/compose/.env.local
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml config
```

Start Typesense:

```bash
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml up -d typesense
```

Check status:

```bash
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml ps typesense
```

Check logs:

```bash
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml logs -f typesense
```

Stop without deleting data:

```bash
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml stop typesense
```

Stop and remove container/network without deleting volume:

```bash
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml down
```

Do not run `docker compose down -v` unless you intentionally want to delete local Typesense index data.

### 8.2 Kubernetes Typesense Setup

Safe apply order:

```bash
kubectl apply -f infra/k8s/data/typesense/namespace.yaml
kubectl create secret generic typesense-secret \
  --namespace data \
  --from-literal=TYPESENSE_API_KEY='replace-with-strong-random-key'
kubectl apply -f infra/k8s/data/typesense/configmap.yaml
kubectl apply -f infra/k8s/data/typesense/service-headless.yaml
kubectl apply -f infra/k8s/data/typesense/service.yaml
kubectl apply -f infra/k8s/data/typesense/statefulset.yaml
kubectl apply -f infra/k8s/data/typesense/pdb.yaml
```

`infra/k8s/data/typesense/secret.example.yaml` documents the required Secret shape only. Do not apply placeholder values to a real shared environment. If you already created the secret and need to update it, use your team's secret manager flow or replace the secret intentionally.

For throwaway local clusters only, you can inspect the example:

```bash
sed -n '1,120p' infra/k8s/data/typesense/secret.example.yaml
```

StatefulSet details:

| Config | Current Value | Why |
|---|---|---|
| Replicas | `3` | HA/quorum-friendly cluster |
| Image | `typesense/typesense:27.1` | Pinned version |
| API port | `8108` | Client traffic |
| Peering port | `8107` | Cluster peer traffic |
| PVC size | `20Gi` per pod | Persistent index data |
| Startup probe | `/health` | Allows slower boot |
| Readiness probe | `/health` | Avoids routing to unready pods |
| Liveness probe | `/health` | Restarts stuck process |
| PDB | `minAvailable: 2` | Keeps quorum safer during voluntary disruption |

Rollout and health:

```bash
kubectl rollout status statefulset/typesense -n data
kubectl get pods -n data -l app.kubernetes.io/name=typesense -o wide
kubectl get pvc -n data -l app.kubernetes.io/name=typesense
kubectl get svc -n data typesense typesense-headless
```

Port-forward and check:

```bash
kubectl port-forward svc/typesense 8108:8108 -n data
curl http://localhost:8108/health
```

Debug cluster state:

```bash
curl "http://localhost:8108/debug" \
  -H "X-TYPESENSE-API-KEY: replace-with-strong-random-key"
```

Expected healthy cluster behavior: one leader and follower nodes. If all pods show leader-like behavior or peers cannot connect, check the nodes ConfigMap, headless service DNS, and peering port `8107`.

## 9. Project Run Instructions

For full beginner onboarding, refer:

```text
TaskImplementation/Search Service/task1_Dependency.md
```

Section:

```text
8. Project Run Instructions
```

Task 2 focused flow:

### Step 1: Create local infra env file

```bash
cp infra/compose/.env.local.example infra/compose/.env.local
```

For local dev, confirm:

```env
TYPESENSE_API_KEY=dev-typesense-key
TYPESENSE_PORT=8108
```

### Step 2: Start Typesense

```bash
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml up -d typesense
```

### Step 3: Verify Typesense

```bash
curl http://localhost:8108/health
curl -H "X-TYPESENSE-API-KEY: dev-typesense-key" http://localhost:8108/collections
```

### Step 4: Configure Search Service to use it

In `backend/services/search-service/.env` for host `go run`:

```env
TYPESENSE_HOST=localhost
TYPESENSE_PORT=8108
TYPESENSE_PROTOCOL=http
TYPESENSE_API_KEY=dev-typesense-key
```

Full `.env` still comes from Task 1 dependency guide.

### Step 5: Start Search Service

```bash
cd backend/services/search-service
set -a
. ./.env
set +a
go run ./cmd/server
```

### Step 6: Verify Search Service can use Typesense

```bash
curl http://localhost:8085/healthz
curl http://localhost:8085/internal/v1/search/schema/products
```

Check Typesense collections:

```bash
curl -H "X-TYPESENSE-API-KEY: dev-typesense-key" http://localhost:8108/collections/products
curl -H "X-TYPESENSE-API-KEY: dev-typesense-key" http://localhost:8108/collections/popular_queries
```

## 10. Migrations and Data Setup

Task 2 has no SQL migrations.

| Data Setup | How It Works |
|---|---|
| Typesense runtime | Docker Compose or K8s starts the server |
| Typesense schema | Search Service code/schema ensure flow creates required collections |
| Product documents | Not Task 2. Product event ingestion comes in Task 3, API search in Task 4, reindex job in Task 8 |
| Local persistence | Docker volume `typesense_data` |
| K8s persistence | StatefulSet PVC mounted at `/data` |

If you need a fully fresh local index:

```bash
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml down
```

Then delete the `typesense_data` volume only if you intentionally want to remove all local search data. Avoid volume deletion in shared/staging/prod environments.

## 11. Common Task 2 Errors and Fixes

| Error | Likely Cause | Fix |
|---|---|---|
| Compose says `TYPESENSE_API_KEY is required` | `infra/compose/.env.local` missing or key empty | Copy `.env.local.example` and set `TYPESENSE_API_KEY` |
| `curl http://localhost:8108/health` fails | Typesense not running or host port changed | Run `docker compose ... ps typesense`; check `TYPESENSE_PORT` |
| `401 Unauthorized` from `/collections` | Wrong API key header | Use same key as `infra/compose/.env.local` |
| Search Service config says `TYPESENSE_HOST must not include a protocol` | Host set to `http://localhost` | Use `TYPESENSE_HOST=localhost`, `TYPESENSE_PROTOCOL=http`, or set `TYPESENSE_URL` |
| Search Service cannot resolve `typesense` | Running Go on host, not inside Docker network | Use `TYPESENSE_HOST=localhost` |
| Data disappears after local cleanup | Volume deleted with `down -v` | Avoid `-v` unless resetting local index intentionally |
| K8s pod stays Pending | PVC/storage class problem | Check `kubectl describe pod` and `kubectl get pvc -n data` |
| K8s cluster cannot form | Nodes file, headless service, or peering DNS issue | Check `typesense-nodes` ConfigMap and `typesense-headless` endpoints |
| K8s Search Service cannot connect | Wrong service DNS or namespace | Use `typesense.data.svc.cluster.local:8108` |
| Multiple local Typesense versions behave differently | Unpinned or stale container image | Use `typesense/typesense:27.1` consistently |

## 12. Security and Operations Notes

| Area | Rule |
|---|---|
| API key | Use strong per-environment key. Never expose it to frontend code. |
| Secret storage | Real values go to `.env`, `.env.local`, Kubernetes Secret, or secret manager, not example files. |
| CORS | `--enable-cors` is okay for local debugging. Production should restrict direct browser access. |
| Network exposure | Keep Typesense private/internal. Public users should call API Gateway/Search Service, not Typesense directly. |
| Backups | Local uses Docker volume. K8s/prod should use storage snapshots or backup plan for `/data`. |
| Upgrades | Snapshot first, upgrade staging, run health/debug/search smoke tests, then promote. |
| Volume deletion | Treat `docker compose down -v` and PVC deletion as data destructive. |
| Observability | Monitor `/health`, `/debug`, pod restarts, PVC usage, search latency, and indexing failures. |

## 13. Final Checklist

- [ ] Previous dependency file reviewed: `TaskImplementation/Search Service/task1_Dependency.md`
- [ ] `infra/compose/.env.local` created from `.env.local.example`
- [ ] `TYPESENSE_API_KEY` set in local Compose env
- [ ] Typesense container starts with `typesense/typesense:27.1`
- [ ] `curl http://localhost:8108/health` returns `{"ok":true}`
- [ ] `/collections` works with `X-TYPESENSE-API-KEY`
- [ ] Host `go run` uses `TYPESENSE_HOST=localhost`
- [ ] Docker-internal service uses `TYPESENSE_HOST=typesense`
- [ ] K8s Search Service config uses `typesense.data.svc.cluster.local`
- [ ] K8s `typesense-secret` created with real key, not example value
- [ ] K8s StatefulSet pods are Running
- [ ] K8s PVCs are Bound
- [ ] Headless and client services exist
- [ ] PDB exists with `minAvailable: 2`
- [ ] No real API key committed
- [ ] No accidental local/prod volume deletion

## Quick Beginner Path

For only Task 2 Typesense setup:

```bash
cp infra/compose/.env.local.example infra/compose/.env.local
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml up -d typesense
curl http://localhost:8108/health
curl -H "X-TYPESENSE-API-KEY: dev-typesense-key" http://localhost:8108/collections
```

Then use the full Search Service run instructions from:

```text
TaskImplementation/Search Service/task1_Dependency.md
```

Section:

```text
8. Project Run Instructions
```
