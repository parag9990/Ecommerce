# Project Dependency & Setup Guide

Input analyzed:

`TaskImplementation/API Gateway Service/task7.md`

Previous dependency documentation reviewed first:

- `TaskImplementation/API Gateway Service/task1_Dependency.md`
- `TaskImplementation/API Gateway Service/task2_Dependency.md`
- `TaskImplementation/API Gateway Service/task3_Dependency.md`
- `TaskImplementation/API Gateway Service/task4_Dependency.md`
- `TaskImplementation/API Gateway Service/task5_Dependency.md`
- `TaskImplementation/API Gateway Service/task6_Dependency.md`

Related implementation inspected:

- `backend/services/api-gateway/go.mod`
- `backend/services/api-gateway/cmd/server/main.go`
- `backend/services/api-gateway/internal/config/config.go`
- `backend/services/api-gateway/internal/logger/logger.go`
- `backend/services/api-gateway/internal/observability/*`
- `backend/services/api-gateway/internal/server/metrics_server.go`
- `backend/services/api-gateway/internal/clients/*`
- `backend/services/api-gateway/internal/transport/http/*`

Simple goal: Ye guide API Gateway Task 7, yani observability setup, ke liye required dependencies, environment variables, Prometheus metrics, OpenTelemetry traces, request id, structured logs, Docker/devops setup, and troubleshooting explain karta hai. Business logic yahan repeat nahi ki gayi.

Important reuse rule:

Base Go install, Redis setup, JWT/JWKS setup, downstream gRPC addresses, request validation setup, error mapping setup, and full database setup pehle dependency docs me already covered hain. Is file me unko duplicate nahi kiya gaya. Jahan setup same hai, wahan exact previous file/section reference diya gaya hai.

---

## 1. Project Overview

Task 7 API Gateway ko observable banata hai.

Request flow simple words me:

```text
Client request
  -> API Gateway
  -> request id middleware
  -> JSON access logs
  -> Prometheus metrics
  -> OpenTelemetry HTTP span
  -> downstream gRPC client span/metadata
  -> response with request_id
```

Current implementation reality:

| Area | Current Status |
|---|---|
| Runtime language | Go |
| HTTP stack | Go standard `net/http` |
| Logger | Go standard `log/slog` JSON handler |
| Request id | `X-Request-Id` header, generated if missing/unsafe |
| Metrics library | `github.com/prometheus/client_golang` |
| Metrics endpoint | Separate HTTP server, default `GET /metrics` on `:9090` |
| Tracing library | OpenTelemetry Go SDK |
| Trace exporter | OTLP over gRPC using `otlptracegrpc` |
| Trace propagation | W3C `traceparent`/`tracestate` plus gRPC metadata |
| User privacy | Raw user id is hashed before log field `user_id_hash` |
| Direct DB | None |
| New Redis usage | None, Redis is inherited from Task 4 rate limiting |
| New queue/Kafka/RabbitMQ | None |
| New app port | No new public API port; metrics uses separate default port `9090` |

What Task 7 adds on top of previous tasks:

| New / Task-specific item | Purpose |
|---|---|
| `internal/observability/config.go` | Observability config defaults and validation |
| `internal/observability/request_id.go` | Request id generate/preserve/store in context |
| `internal/observability/logging.go` | Structured access logs |
| `internal/observability/metrics.go` | Prometheus collectors and `/metrics` handler |
| `internal/observability/tracing.go` | OpenTelemetry provider and HTTP tracing middleware |
| `internal/observability/grpc_client_interceptor.go` | Outbound gRPC trace and metrics interception |
| `internal/observability/redaction.go` | Sensitive field redaction and user id hashing |
| `internal/server/metrics_server.go` | Separate metrics HTTP server |

Already documented elsewhere and not repeated in full:

| Topic | Refer |
|---|---|
| Base clone/install/run flow | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `9. Local Development Setup` |
| Go module basics | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `4. Dependency Management` |
| Full `.env` for Gateway baseline | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `7. Environment Variables` |
| Downstream gRPC addresses | `TaskImplementation/API Gateway Service/task2_Dependency.md` -> `7. Environment Variables` |
| JWT/JWKS config | `TaskImplementation/API Gateway Service/task3_Dependency.md` -> `7. Environment Variables` |
| Redis rate limiting | `TaskImplementation/API Gateway Service/task4_Dependency.md` -> `5. Database Setup` |
| Request validation env | `TaskImplementation/API Gateway Service/task5_Dependency.md` -> `7. Environment Variables` |
| Error envelope behavior | `TaskImplementation/API Gateway Service/task6_Dependency.md` -> `1. Project Overview` |

Beginner note:

```text
Observability ka matlab hai: service ke andar kya ho raha hai wo logs, metrics, and traces se samajhna.
Logs exact request/event batate hain.
Metrics aggregate health batate hain.
Traces ek request ka full journey dikhate hain.
```

---

## 2. Tech Stack

### Task 7 Specific Technologies

| Technology | What it is | Why this project uses it | Required? | Beginner Hinglish explanation |
|---|---|---|---:|---|
| Go | Backend programming language | API Gateway service Go me implemented hai | Yes | Go ek fast compiled backend language hai. Gateway isi me run/test hota hai. |
| Go Modules | Dependency management system | `go.mod` and `go.sum` dependency versions manage karte hain | Yes | Go modules npm/pip jaisa dependency manager hai, Go projects ke liye. |
| `net/http` | Go standard HTTP server package | Middleware, health routes, metrics route, response recording ke liye | Yes | External web framework nahi hai; Go ka built-in HTTP package use ho raha hai. |
| `log/slog` | Go standard structured logging package | JSON logs stdout par write karne ke liye | Yes | `slog` structured key-value logs banata hai, jo Docker/Loki easily parse kar sakte hain. |
| Request ID | Per-request unique id | Logs, response, support debugging correlate karne ke liye | Yes | Har request ko ek tracking number milta hai. Support ticket me ye id useful hota hai. |
| Prometheus Go client | Metrics library | Counters, histograms, gauges expose karne ke liye | Yes if metrics enabled | Prometheus metrics se RPS, latency, error rate jaise numbers milte hain. |
| Prometheus | Metrics scraper/time-series DB | Gateway `/metrics` endpoint scrape karne ke liye | Optional for local, recommended in dev/prod | Prometheus service metrics collect karta hai and time ke saath store karta hai. |
| OpenTelemetry | Observability standard | Distributed tracing and context propagation ke liye | Yes if tracing enabled | OpenTelemetry ek common standard hai jisse request ka journey trace hota hai. |
| OTLP gRPC exporter | Trace exporter | Gateway se collector/Jaeger/Tempo ko traces send karne ke liye | Required when tracing enabled | OTLP trace data bhejne ka protocol hai. |
| OpenTelemetry Collector | Telemetry pipeline | Traces receive karke Jaeger/Tempo ya backend ko forward karne ke liye | Recommended if tracing enabled | Collector middle layer hai jo telemetry receive, process, and export karta hai. |
| Jaeger or Tempo | Trace backend | Distributed traces inspect karne ke liye | Optional local, recommended dev/prod | Trace UI me dikhta hai ki request Gateway se kaunsi downstream service tak gayi. |
| Grafana | Dashboard UI | Metrics/logs/traces visualize karne ke liye | Optional local, recommended dev/prod | Grafana dashboard me latency, errors, traffic charts dikhte hain. |
| Loki | Log aggregation backend | JSON stdout logs centrally query karne ke liye | Optional | Loki logs store/query karta hai. Local me direct terminal logs bhi enough ho sakte hain. |
| gRPC metadata | Key/value metadata with gRPC calls | `x-request-id` and `traceparent` downstream services tak bhejne ke liye | Yes for full request tracing | HTTP headers jaisa concept, but internal gRPC calls ke saath. |

### Important Difference From `task7.md` Prose

`task7.md` examples mention `go.uber.org/zap` and `otelgrpc` as possible libraries. Current repo code does not use them.

| Library mentioned in task guide | Current repo status | What to do |
|---|---|---|
| `go.uber.org/zap` | Not present in `go.mod` | Install mat karo unless code intentionally migrates from `log/slog` to zap |
| `go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc` | Not present in `go.mod` | Current repo has custom gRPC interceptors; add only if implementation changes |

Beginner rule:

```text
Sirf documentation me package ka naam dekh kar `go get` mat run karo.
Dependency tab add karo jab code us package ko import karta ho.
```

### Reused Stack

General Gateway, gRPC, JWT/JWKS, Redis, request validation, and gRPC error mapping stack already explained hai:

- `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `2. Tech Stack`
- `TaskImplementation/API Gateway Service/task2_Dependency.md` -> `2. Tech Stack`
- `TaskImplementation/API Gateway Service/task3_Dependency.md` -> `2. Tech Stack`
- `TaskImplementation/API Gateway Service/task4_Dependency.md` -> `2. Tech Stack`
- `TaskImplementation/API Gateway Service/task5_Dependency.md` -> `2. Tech Stack`
- `TaskImplementation/API Gateway Service/task6_Dependency.md` -> `2. Tech Stack`

---

## 3. Required Software

### Minimum for Task 7 Code/Test Work

| Software | Required? | Why |
|---|---:|---|
| Git | Yes | Repository clone/update ke liye |
| Go `1.26.3` compatible toolchain | Yes | `backend/services/api-gateway/go.mod` declares `go 1.26.3` |
| curl | Recommended | Health and metrics endpoints verify karne ke liye |
| jq | Optional | JSON health responses pretty-print karne ke liye |

Base installation steps already documented:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`3. Required Software`

### Required for Full Observability Local Stack

| Software / Service | Required? | Why |
|---|---:|---|
| Docker / Docker Compose | Recommended | Prometheus, Jaeger, OTel Collector, Grafana, Loki local run karne ke liye |
| Prometheus | Optional but useful | Gateway metrics scrape karne ke liye |
| OpenTelemetry Collector | Required if `TRACE_ENABLED=true` and collector endpoint is used | Gateway traces receive karne ke liye |
| Jaeger or Tempo | Optional but useful | Trace UI me spans dekhne ke liye |
| Grafana | Optional | Dashboards ke liye |
| Loki | Optional | Logs centralize/query karne ke liye |

Beginner setup modes:

| Mode | What you need | Best for |
|---|---|---|
| Unit-test only | Go | Fast Task 7 verification |
| Metrics-only local | Go + running Gateway + curl | `/metrics` check karna |
| Full tracing local | Go + Gateway + OTel Collector + Jaeger/Tempo | Trace propagation verify karna |
| Full platform runtime | Everything above + Redis + JWKS + 12 downstream gRPC services | Real Gateway local run |

Important:

```text
Full `go run ./cmd/server` all 12 downstream gRPC services dial karta hai.
Agar aap sirf observability code verify kar rahe ho, pehle unit tests run karo.
Full Gateway runtime ke liye Task 1-6 dependencies bhi ready honi chahiye.
```

---

## 4. Dependency Management

This is a Go project.

Base Go module explanation already covered hai:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`4. Dependency Management`

### Current Go Module

```text
module ecommerce/api-gateway
go 1.26.3
```

### Task 7 Relevant Go Dependencies

Current `backend/services/api-gateway/go.mod` already includes Task 7 dependencies:

| Dependency | Version in `go.mod` | Task 7 Use |
|---|---:|---|
| `github.com/prometheus/client_golang` | `v1.23.2` | Prometheus collectors and `promhttp` metrics handler |
| `go.opentelemetry.io/otel` | `v1.43.0` | Tracer API, propagation, attributes |
| `go.opentelemetry.io/otel/sdk` | `v1.43.0` | Tracer provider, sampler, batch exporter |
| `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc` | `v1.43.0` | Sends traces to OTLP gRPC endpoint |
| `go.opentelemetry.io/otel/trace` | `v1.43.0` | Span context and trace ids |
| `google.golang.org/grpc` | `v1.81.1` | gRPC client interceptors and metadata propagation |

Standard library packages used by Task 7:

| Package | Why used |
|---|---|
| `log/slog` | JSON structured logs |
| `crypto/rand` | Secure request id generation |
| `crypto/sha256` | User id hashing |
| `encoding/hex` | Request id/hash formatting |
| `net/http` | Middleware, response recorder, metrics server |
| `context` | Request id and trace context flow |
| `time` | Latency measurement and trace shutdown timeout |

### Do You Need To Install a New Go Package?

No new third-party package is required right now.

Use this only to download existing dependencies:

```bash
cd backend/services/api-gateway
go mod download
```

If someone edits imports or dependency versions:

```bash
cd backend/services/api-gateway
go mod tidy
```

Task 7 focused tests:

```bash
cd backend/services/api-gateway
go test ./internal/observability ./internal/transport/http ./internal/clients
```

All Gateway tests:

```bash
cd backend/services/api-gateway
go test ./...
```

Build:

```bash
cd backend/services/api-gateway
go build ./cmd/server
```

Run:

```bash
cd backend/services/api-gateway
go run ./cmd/server
```

### Common Go Dependency Issues

General Go dependency problems already documented:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`4. Dependency Management` -> `Dependency Problems and Fixes`

Task 7 specific notes:

| Error | Likely Cause | Fix |
|---|---|---|
| `cannot find package github.com/prometheus/client_golang/...` | Dependencies not downloaded or `go.sum` missing | Run `go mod download`, then `go mod tidy` |
| `cannot find package go.opentelemetry.io/otel/...` | Dependency cache empty or module file changed | Run `go mod download` |
| `go: go.mod requires go >= 1.26.3` | Local Go version old hai | Install compatible Go version |
| Developer added zap import and build fails | Current repo does not depend on zap | Remove copied zap import or intentionally add/migrate logger |
| Developer added `otelgrpc` import and build fails | Current repo uses custom interceptors | Remove copied import or intentionally add dependency with design change |

---

## 5. Database Setup

### API Gateway Task 7 Direct Database Requirement

Task 7 does not introduce a SQL or document database.

| Store | Directly used by Gateway Task 7? | Purpose |
|---|---:|---|
| MySQL | No | Owned by downstream services |
| MongoDB | No | Owned by downstream services |
| Redis | No new usage | Inherited from Task 4 rate limiting |
| Typesense | No | Owned by Search Service |
| Kafka/RabbitMQ | No | Platform async/event layer, not Task 7 direct runtime |

### Observability Storage Is Different From App Database

Prometheus, Loki, Jaeger, and Tempo store observability data, not ecommerce business data.

| Observability backend | Stores | Required for Gateway to start? |
|---|---|---:|
| Prometheus | Metrics time series | No |
| Loki | Logs | No |
| Jaeger/Tempo | Traces | No, but needed to view traces |
| OpenTelemetry Collector | Telemetry pipeline data temporarily | Required only when tracing endpoint is used |

### Reuse Previous Database Documentation

Full platform database setup already explained in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Sections:

- `5. Database Setup`
- `6. Redis / Queue / External Services`

Redis rate-limiting details already explained in:

`TaskImplementation/API Gateway Service/task4_Dependency.md`

Sections:

- `5. Database Setup`
- `6. Redis / Queue / External Services`

### Migrations

Task 7 has no database migration.

```text
No SQL migration.
No MongoDB migration.
No Redis schema migration.
No Kafka topic migration.
```

Prometheus/Loki/Jaeger Docker volumes may be created for local observability data, but those are infrastructure volumes, not application migrations.

---

## 6. Redis / Queue / External Services

### Direct New External Services for Task 7

| Service | What it does | Required? | Used by current code? |
|---|---|---:|---:|
| Prometheus | Scrapes `/metrics` | Optional local, recommended dev/prod | Gateway exposes metrics for it |
| OpenTelemetry Collector | Receives OTLP traces | Required if tracing enabled and endpoint points to collector | Yes when `TRACE_ENABLED=true` |
| Jaeger or Tempo | Stores/displays traces | Optional local, recommended dev/prod | Indirect through collector or direct OTLP endpoint |
| Grafana | Dashboard UI | Optional | Not directly called by Gateway |
| Loki | Log aggregation | Optional | Not directly called; reads stdout logs via log shipper/runtime |

### Prometheus

#### A. What It Is

Prometheus ek metrics scraper and time-series database hai. Ye `/metrics` endpoint ko scrape karta hai and values ko time ke saath store karta hai.

#### B. Why This Project Uses It

API Gateway ye metrics expose karta hai:

| Metric | Purpose |
|---|---|
| `http_requests_total` | Total HTTP requests by route/status/error code |
| `http_request_duration_seconds` | HTTP latency histogram |
| `grpc_client_requests_total` | Outbound gRPC request count |
| `grpc_client_duration_seconds` | Outbound gRPC latency histogram |
| `gateway_rate_limited_total` | 429/rate limit signals |
| `gateway_auth_failures_total` | Auth/RBAC failure signals |
| `gateway_validation_failures_total` | Request validation failure signals |
| `gateway_downstream_errors_total` | Downstream gRPC errors |
| `gateway_inflight_requests` | Current in-flight requests |

#### C. Required or Optional

Optional for local coding. Recommended for realistic dev/staging/prod.

Gateway can start and expose `/metrics` without Prometheus. Prometheus is needed when you want historical charts and alerts.

#### D. Local Installation

Use Docker for local setup. Base Docker install is already covered in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`3. Required Software`

#### E. Docker Setup

If Gateway runs on host machine and Prometheus runs in Docker, use a scrape config like this:

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: api-gateway
    metrics_path: /metrics
    static_configs:
      - targets:
          - host.docker.internal:19090
```

For Linux Docker, add this host mapping when running Prometheus:

```bash
docker run --rm \
  --add-host=host.docker.internal:host-gateway \
  -p 9090:9090 \
  -v "$PWD/infra/monitoring/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro" \
  prom/prometheus:latest
```

Note:

```text
`prom/prometheus:latest` local learning ke liye okay hai.
Shared/dev/prod infra me tested version pin karo.
```

#### F. Start Commands

Gateway metrics server default `:9090` par start hota hai. Prometheus UI bhi default `:9090` use karta hai, so local conflict avoid karne ke liye Gateway metrics port ko `19090` set karna easy hai:

```bash
export METRICS_ADDR=:19090
```

Start Prometheus after config file exists:

```bash
docker run --rm \
  --add-host=host.docker.internal:host-gateway \
  -p 9090:9090 \
  -v "$PWD/infra/monitoring/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro" \
  prom/prometheus:latest
```

#### G. Verify Running

Gateway metrics endpoint:

```bash
curl http://localhost:19090/metrics
```

Prometheus UI:

```text
http://localhost:9090
```

Prometheus query examples:

```promql
http_requests_total{service="api-gateway"}
```

```promql
histogram_quantile(
  0.95,
  sum(rate(http_request_duration_seconds_bucket{service="api-gateway"}[5m])) by (le, route)
)
```

#### H. Default Port

| Item | Port |
|---|---:|
| Gateway metrics endpoint | `9090` by default via `METRICS_PORT=9090` |
| Recommended local Gateway metrics port when Prometheus UI also runs | `19090` via `METRICS_ADDR=:19090` |
| Prometheus UI | `9090` |

#### I. Connection String / URL Format

Gateway does not connect to Prometheus. Prometheus scrapes Gateway:

```text
http://<gateway-host>:<metrics-port>/metrics
```

Examples:

```text
http://localhost:19090/metrics
http://api-gateway:9090/metrics
```

#### J. Where To Place Credentials

No Prometheus credentials are used by current Gateway code.

If production Prometheus requires auth/TLS, configure it in infrastructure config, not in Gateway app code.

### OpenTelemetry Collector

#### A. What It Is

OpenTelemetry Collector telemetry pipeline hai. Gateway traces collector ko bhejta hai; collector un traces ko Jaeger, Tempo, or another backend ko forward karta hai.

#### B. Why This Project Uses It

Collector se services vendor-neutral rehti hain. Aaj Jaeger use kar sakte ho, kal Tempo/Datadog/New Relic without changing Gateway code.

#### C. Required or Optional

Required only when:

```env
TRACE_ENABLED=true
```

and `TRACE_EXPORTER_OTLP_ENDPOINT` collector ko point karta hai.

For simple local work without tracing UI:

```env
TRACE_ENABLED=false
```

#### D. Local Installation

Docker recommended. Native install beginner ke liye zaruri nahi.

#### E. Docker Setup

Example `otel-collector-config.yml`:

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

exporters:
  otlp/jaeger:
    endpoint: jaeger:4317
    tls:
      insecure: true

service:
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [otlp/jaeger]
```

Docker run:

```bash
docker run --rm \
  -p 4317:4317 \
  -p 4318:4318 \
  -v "$PWD/infra/monitoring/otel-collector/otel-collector-config.yml:/etc/otelcol/config.yml:ro" \
  otel/opentelemetry-collector-contrib:latest \
  --config=/etc/otelcol/config.yml
```

#### F. Start Commands

If Gateway runs on host:

```bash
export TRACE_ENABLED=true
export TRACE_EXPORTER_OTLP_ENDPOINT=localhost:4317
export TRACE_EXPORTER_OTLP_INSECURE=true
```

If Gateway runs inside Docker Compose with collector service named `otel-collector`:

```bash
export TRACE_EXPORTER_OTLP_ENDPOINT=otel-collector:4317
```

#### G. Verify Running

Check collector port:

```bash
curl -I http://localhost:4318
```

Note: OTLP gRPC on `4317` is not a normal browser endpoint. Use traces generated by Gateway and then check Jaeger/Tempo.

#### H. Default Port

| Protocol | Port |
|---|---:|
| OTLP gRPC | `4317` |
| OTLP HTTP | `4318` |

Current Gateway code uses OTLP gRPC.

#### I. Connection String / URL Format

Current code uses `otlptracegrpc.WithEndpoint`, so use host and port without `http://`:

```env
TRACE_EXPORTER_OTLP_ENDPOINT=localhost:4317
```

Do not use this format for current code:

```env
TRACE_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
```

#### J. Where To Place Credentials

Local insecure mode:

```env
TRACE_EXPORTER_OTLP_INSECURE=true
```

Production:

```env
TRACE_EXPORTER_OTLP_INSECURE=false
```

If production collector requires client certs or auth headers, current Gateway code does not yet expose env variables for those. Add explicit config support before enabling that deployment mode.

### Jaeger or Tempo

#### A. What It Is

Jaeger/Tempo trace backend hai. Ye trace spans store karta hai and UI me request journey dikhata hai.

#### B. Why This Project Uses It

Gateway se downstream gRPC services tak request kahan slow/fail hui, ye trace UI me quickly dikhta hai.

#### C. Required or Optional

Optional for local. Recommended for dev/staging/prod debugging.

#### D. Local Installation

Docker recommended.

#### E. Docker Setup

Jaeger all-in-one local command:

```bash
docker run --rm \
  --name jaeger \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest
```

If you send Gateway traces directly to Jaeger:

```env
TRACE_ENABLED=true
TRACE_EXPORTER_OTLP_ENDPOINT=localhost:4317
TRACE_EXPORTER_OTLP_INSECURE=true
```

#### F. Start Commands

```bash
docker run --rm \
  --name jaeger \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest
```

#### G. Verify Running

Open:

```text
http://localhost:16686
```

Search service:

```text
api-gateway
```

#### H. Default Port

| Item | Port |
|---|---:|
| Jaeger UI | `16686` |
| OTLP gRPC | `4317` |
| OTLP HTTP | `4318` |

#### I. Connection String / URL Format

Gateway OTLP gRPC:

```env
TRACE_EXPORTER_OTLP_ENDPOINT=localhost:4317
```

#### J. Where To Place Credentials

Local Jaeger all-in-one has no credentials. Do not expose it publicly.

### Grafana and Loki

#### A. What They Are

Grafana dashboard UI hai. Loki log storage/query backend hai.

#### B. Why This Project Uses Them

Gateway logs stdout par JSON format me aate hain. Docker/Kubernetes log shipper un logs ko Loki me bhej sakta hai. Grafana me logs, metrics, and traces ek jagah correlate kar sakte ho.

#### C. Required or Optional

Optional for local. Recommended for shared environments.

#### D. Local Installation

Docker recommended.

#### E. Docker Setup

Basic local run:

```bash
docker run --rm --name loki -p 3100:3100 grafana/loki:latest
docker run --rm --name grafana -p 3000:3000 grafana/grafana:latest
```

#### F. Start Commands

Same as Docker setup above.

#### G. Verify Running

Grafana:

```text
http://localhost:3000
```

Loki ready check:

```bash
curl http://localhost:3100/ready
```

#### H. Default Ports

| Service | Port |
|---|---:|
| Grafana UI | `3000` |
| Loki HTTP | `3100` |

#### I. Connection String / URL Format

Grafana data source examples:

```text
Prometheus: http://prometheus:9090
Loki: http://loki:3100
Jaeger: http://jaeger:16686
```

#### J. Where To Place Credentials

Grafana local default is commonly `admin/admin`, but change it for shared environments.

Current Gateway app does not store Grafana/Loki credentials.

### Ports & Networking

| Service | Port | Purpose | New or Reused |
|---|---:|---|---|
| API Gateway HTTP | `8080` via `HTTP_ADDR=:8080` | Public REST and health endpoints | Reused |
| API Gateway metrics | `9090` via `METRICS_PORT=9090` | Prometheus scrape endpoint | Task 7 |
| Recommended local Gateway metrics alternate | `19090` via `METRICS_ADDR=:19090` | Avoid conflict with Prometheus UI | Task 7 |
| Prometheus UI | `9090` | Query metrics and targets | Task 7 external |
| OpenTelemetry Collector gRPC | `4317` | Receive OTLP gRPC traces | Task 7 external |
| OpenTelemetry Collector HTTP | `4318` | Receive OTLP HTTP telemetry | Task 7 external |
| Jaeger UI | `16686` | View traces | Task 7 external |
| Grafana UI | `3000` | Dashboards | Task 7 external |
| Loki HTTP | `3100` | Logs API | Task 7 external |
| Redis | `6379` | Rate-limit counters | Reused Task 4 |
| Downstream gRPC services | `50051` to `50062` local examples | Gateway startup/readiness | Reused Task 2 |

---

## 7. Environment Variables

Full Gateway `.env` is already documented:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`7. Environment Variables`

Task 7 adds or heavily depends on these observability-specific variables.

### Where To Create `.env`

Recommended local file:

```text
backend/services/api-gateway/.env
```

Important:

```text
Current Go code uses `os.Getenv`.
It does not auto-load `.env`.
You must export variables before `go run`.
```

Export command:

```bash
cd backend/services/api-gateway
set -a
source .env
set +a
```

### Task 7 `.env` Additions

Use this subset only for Task 7 observability variables. Keep the full Gateway variables from previous dependency docs.

```env
# Observability identity
SERVICE_NAME=api-gateway
APP_ENV=local
LOG_LEVEL=info

# Request id
REQUEST_ID_HEADER=X-Request-Id

# Metrics
METRICS_ENABLED=true
METRICS_ADDR=:19090
# Alternative if METRICS_ADDR is not set:
# METRICS_PORT=9090

# Tracing
TRACE_ENABLED=true
TRACE_EXPORTER_OTLP_ENDPOINT=localhost:4317
TRACE_EXPORTER_OTLP_INSECURE=true
TRACE_SAMPLE_RATIO=1.0
TRACE_SHUTDOWN_TIMEOUT=5s

# Privacy
OBSERVABILITY_HASH_SALT=replace-with-local-random-salt
```

For simplest local development without collector:

```env
TRACE_ENABLED=false
METRICS_ENABLED=true
METRICS_ADDR=:19090
```

### Variable Explanation

| Variable | Required? | Default in code | Purpose | Notes |
|---|---:|---|---|---|
| `SERVICE_NAME` | Optional | `api-gateway` | Logs, metrics labels, trace resource name | Keep stable across environments |
| `APP_ENV` | Optional | `local` | Environment tag | Examples: `local`, `dev`, `staging`, `prod` |
| `LOG_LEVEL` | Optional | `info` | `slog` level | Use `debug` locally only when needed |
| `REQUEST_ID_HEADER` | Optional | `X-Request-Id` | Header to read/write request id | Must not contain whitespace |
| `METRICS_ENABLED` | Optional | `true` | Enables metrics registry/server | Set false only for debug |
| `METRICS_ADDR` | Optional | Derived from `METRICS_PORT` | Full metrics bind address | Takes priority over `METRICS_PORT` |
| `METRICS_PORT` | Optional | `9090` | Metrics port if `METRICS_ADDR` blank | Code converts `9090` to `:9090` |
| `TRACE_ENABLED` | Optional | `true` | Enables OpenTelemetry tracing | Set false if no collector/backend locally |
| `TRACE_EXPORTER_OTLP_ENDPOINT` | Required when tracing enabled | `otel-collector:4317` | OTLP gRPC target | Use `localhost:4317` on host, `otel-collector:4317` in Compose |
| `TRACE_EXPORTER_OTLP_INSECURE` | Optional | `true` | Disable TLS for OTLP local | Use false for TLS-enabled production collector |
| `TRACE_SAMPLE_RATIO` | Optional | `1.0` | Trace sampling ratio | Must be between `0` and `1` |
| `TRACE_SHUTDOWN_TIMEOUT` | Optional | `5s` | Max time to flush traces during shutdown | Must be positive Go duration |
| `OBSERVABILITY_HASH_SALT` | Strongly recommended | blank | Salt for `user_id_hash` | Set a non-empty secret per environment |

### Credential Placement

| Credential / Secret | Where to place locally | Production recommendation |
|---|---|---|
| `OBSERVABILITY_HASH_SALT` | `backend/services/api-gateway/.env` | Secret manager or Kubernetes Secret |
| OTLP TLS/auth values | Not fully supported by current Gateway config | Add explicit config before production collector auth |
| Grafana admin password | Grafana container env or secret | Secret manager, rotate default password |
| Redis/JWT/downstream secrets | Previous task `.env` sections | See Task 3 and Task 4 dependency docs |

Do not commit `.env` with real secrets.

### Validation Rules

Current config validation checks:

| Variable | Rule |
|---|---|
| `REQUEST_ID_HEADER` | Required after defaults, no whitespace |
| `METRICS_ADDR` | Required if metrics enabled |
| `TRACE_EXPORTER_OTLP_ENDPOINT` | Required if tracing enabled |
| `TRACE_SAMPLE_RATIO` | Must be `0 <= value <= 1` |
| `TRACE_SHUTDOWN_TIMEOUT` | Must be positive duration |

---

## 8. Docker Setup

### Current Repo Status

At the time of inspection, no API Gateway Dockerfile or local Docker Compose file was found in the repo.

Already documented:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`8. Docker Setup`

Task 7 does not add an API Gateway container. It adds observability service expectations.

### Local Observability Compose Example

This is a documentation example. Create the referenced config files only if you want a local observability stack.

```yaml
services:
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    extra_hosts:
      - "host.docker.internal:host-gateway"
    volumes:
      - ./infra/monitoring/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus-data:/prometheus
    restart: unless-stopped

  jaeger:
    image: jaegertracing/all-in-one:latest
    environment:
      COLLECTOR_OTLP_ENABLED: "true"
    ports:
      - "16686:16686"
      - "4317:4317"
      - "4318:4318"
    restart: unless-stopped

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    volumes:
      - grafana-data:/var/lib/grafana
    restart: unless-stopped

  loki:
    image: grafana/loki:latest
    ports:
      - "3100:3100"
    volumes:
      - loki-data:/loki
    restart: unless-stopped

volumes:
  prometheus-data:
  grafana-data:
  loki-data:
```

### Docker Networking Rules

| Where Gateway runs | Metrics target for Prometheus | Trace endpoint for Gateway |
|---|---|---|
| Gateway on host, Prometheus in Docker | `host.docker.internal:19090` | `localhost:4317` |
| Gateway inside same Compose network | `api-gateway:9090` | `jaeger:4317` or `otel-collector:4317` |
| Gateway in Kubernetes | Kubernetes service DNS | Collector service DNS |

### Volumes

| Volume | Purpose |
|---|---|
| `prometheus-data` | Prometheus metrics history |
| `grafana-data` | Grafana dashboards/data sources |
| `loki-data` | Local log data |

Beginner note:

```text
Docker volume delete karne se local metrics/logs/dashboards delete ho sakte hain.
Production me volume retention policy carefully define karo.
```

### Restart Policies

For local development:

```yaml
restart: unless-stopped
```

For production, use Kubernetes Deployments/StatefulSets or managed observability services. Task 7 does not include Kubernetes manifests.

---

## 9. Local Development Setup

### Reused Base Onboarding

Follow these first:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Sections:

- `9. Local Development Setup`
- `10. Running the Project`

For full runtime, also follow:

- `TaskImplementation/API Gateway Service/task2_Dependency.md` -> downstream gRPC services
- `TaskImplementation/API Gateway Service/task3_Dependency.md` -> JWKS/Auth setup
- `TaskImplementation/API Gateway Service/task4_Dependency.md` -> Redis setup
- `TaskImplementation/API Gateway Service/task5_Dependency.md` -> validation env
- `TaskImplementation/API Gateway Service/task6_Dependency.md` -> error response checks

### Task 7 Incremental Setup

Step 1: Go to Gateway service.

```bash
cd backend/services/api-gateway
```

Step 2: Download dependencies.

```bash
go mod download
```

Step 3: Create/export full Gateway env from previous docs.

```bash
set -a
source .env
set +a
```

Step 4: Choose observability mode.

Simple local mode:

```bash
export TRACE_ENABLED=false
export METRICS_ENABLED=true
export METRICS_ADDR=:19090
```

Full trace mode with Jaeger:

```bash
export TRACE_ENABLED=true
export TRACE_EXPORTER_OTLP_ENDPOINT=localhost:4317
export TRACE_EXPORTER_OTLP_INSECURE=true
export TRACE_SAMPLE_RATIO=1.0
export METRICS_ADDR=:19090
```

Step 5: Run focused tests.

```bash
go test ./internal/observability ./internal/transport/http ./internal/clients
```

Step 6: Start optional observability backend.

```bash
docker run --rm \
  --name jaeger \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest
```

Step 7: Start full Gateway only after previous task dependencies are ready.

```bash
go run ./cmd/server
```

Step 8: Verify request id, logs, metrics.

```bash
curl -i -H "X-Request-Id: req_local_123" http://localhost:8080/health/live
curl http://localhost:19090/metrics
```

Expected:

| Check | Expected |
|---|---|
| Response header | `X-Request-Id: req_local_123` |
| Response body | JSON envelope includes `request_id` |
| Terminal logs | JSON log with `request_id`, route, status, latency |
| Metrics endpoint | Prometheus text with `http_requests_total` |
| Jaeger UI | Service `api-gateway` appears after traced requests |

### Current Runtime Limitation

Current handler for route-contract APIs still returns:

```text
501 ROUTE_BRIDGE_NOT_CONFIGURED
```

That means:

- Request id, logs, HTTP metrics, and HTTP traces can still be verified.
- Real downstream gRPC child spans appear only when an implemented handler actually invokes downstream RPCs.
- `/health/ready` does call downstream gRPC health checks, but full readiness requires all configured downstream services to be running and serving health responses.

---

## 10. Running the Project

### Minimal Task 7 Verification

No Redis, no database, no JWKS, no downstream services required for this:

```bash
cd backend/services/api-gateway
go test ./internal/observability
```

Broader middleware/client verification:

```bash
cd backend/services/api-gateway
go test ./internal/observability ./internal/transport/http ./internal/clients
```

### Full Gateway Runtime

Full runtime still needs previous task dependencies:

| Dependency | Why |
|---|---|
| Redis | Rate limiting if `RATE_LIMIT_ENABLED=true` |
| 12 downstream gRPC services | Startup client registry and readiness |
| Auth JWKS endpoint | Protected routes |
| API contract file | Route catalog and validation schemas |

Start full Gateway:

```bash
cd backend/services/api-gateway
set -a
source .env
set +a
go run ./cmd/server
```

Verify health:

```bash
curl http://localhost:8080/health/live
curl http://localhost:8080/health/ready
```

Verify metrics:

```bash
curl http://localhost:19090/metrics
```

Send sample request with request id:

```bash
curl -i \
  -H "X-Request-Id: req_observe_123" \
  http://localhost:8080/health/live
```

### Migrations

Task 7 has no migrations.

Application database migrations are downstream-service responsibilities and already covered in previous dependency guides.

---

## 11. Common Errors & Fixes

### Task 7 Specific Troubleshooting

| Error / Symptom | Likely Cause | Fix |
|---|---|---|
| `listen tcp :9090: bind: address already in use` | Gateway metrics and Prometheus UI both want port `9090` | Set `METRICS_ADDR=:19090` or move Prometheus UI to another host port |
| `/metrics` returns connection refused | Metrics server not running or wrong port | Check `METRICS_ENABLED=true` and `METRICS_ADDR`/`METRICS_PORT` |
| `/metrics` returns 404 | Metrics object nil or metrics disabled | Set `METRICS_ENABLED=true` and restart Gateway |
| `METRICS_ADDR is required when metrics are enabled` | Metrics enabled but address resolved blank | Set `METRICS_ADDR=:19090` or `METRICS_PORT=9090` |
| `TRACE_SAMPLE_RATIO must be between 0 and 1` | Invalid env value | Use `1.0`, `0.5`, `0.1`, or `0` |
| `TRACE_SHUTDOWN_TIMEOUT must be positive` | Bad duration | Use `5s` or another positive Go duration |
| Traces not visible in Jaeger | Wrong OTLP endpoint, collector down, or no traced request exported yet | Use `TRACE_EXPORTER_OTLP_ENDPOINT=localhost:4317`, start Jaeger, send request |
| `otel-collector: no such host` or export failures | Compose DNS used while Gateway runs on host | Use `localhost:4317` locally, or run Gateway in same Compose network |
| Logs missing `request_id` | Request did not pass through Gateway middleware or custom handler bypassed middleware | Use router created by `NewRouterWithOptions`; do not bypass middleware |
| Logs missing `trace_id` | Tracing disabled, trace span invalid, or no trace backend configured | Set `TRACE_ENABLED=true`; verify OTel setup |
| Prometheus has too many series | Raw ids used as metric labels | Keep route labels templated, e.g. `/api/v1/products/{product_id}` |
| `request_id` header not preserved | Incoming id invalid by regex | Use id length 8-128 chars with letters/numbers/underscore/colon/dash |
| User ids visible in logs | Developer logged raw auth claims outside observability helpers | Use `HashUserID`; never log raw user id/token |
| No gRPC child spans for normal API route | Route bridge currently returns 501 and may not call downstream RPC yet | Verify HTTP spans now; gRPC spans appear once real route bridge invokes clients |

### Reused Troubleshooting

| Topic | Refer |
|---|---|
| Go module failures | `task1_Dependency.md` -> `4. Dependency Management` |
| Port conflicts and Gateway startup | `task1_Dependency.md` -> `11. Common Errors & Fixes` |
| gRPC dial failures | `task2_Dependency.md` -> `11. Common Errors & Fixes` |
| JWT/JWKS failures | `task3_Dependency.md` -> `11. Common Errors & Fixes` |
| Redis failures | `task4_Dependency.md` -> `11. Common Errors & Fixes` |
| Validation failures | `task5_Dependency.md` -> `11. Common Errors & Fixes` |
| Error envelope surprises | `task6_Dependency.md` -> `11. Common Errors & Fixes` |

---

## 12. Security & Best Practices

### Task 7 Security Notes

| Area | Best Practice |
|---|---|
| Request id | Preserve only safe ids; generate new id for unsafe input |
| Logs | Log structured fields, not raw request/response body |
| Tokens | Never log JWT, refresh token, API key, cookie, OTP, password, card, CVV |
| User identity | Use `user_id_hash`, not raw `user_id` |
| Hash salt | Set `OBSERVABILITY_HASH_SALT` per environment |
| Metrics labels | Never use raw ids, emails, phone numbers, tokens, or full paths as labels |
| Trace attributes | Keep attributes useful but non-sensitive |
| Metrics endpoint | Bind privately; do not expose public internet without auth/network policy |
| OTLP endpoint | Use TLS/auth in production when collector supports it |
| Grafana | Change default admin password |
| Loki/log retention | Define retention; logs can contain operationally sensitive data |

### Beginner Best Practices

- Local me `TRACE_SAMPLE_RATIO=1.0` okay hai because debugging easy hoti hai.
- Production high traffic me sampling lower rakho, for example `0.1`, unless critical flow needs full tracing.
- `LOG_LEVEL=debug` local debugging ke liye use karo; prod me usually `info` or `warn`.
- Prometheus labels low-cardinality rakho. Route template good hai, raw product/order/user id bad hai.
- Metrics port and Prometheus UI port dono `9090` na rakho on same host.
- `.env` me real secrets commit mat karo.
- Every support-facing error response ka `request_id` preserve karo so logs/traces easily search ho sake.

### Already Documented Best Practices

General security and config best practices already covered:

- `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `12. Security & Best Practices`
- `TaskImplementation/API Gateway Service/task3_Dependency.md` -> `12. Security & Best Practices`
- `TaskImplementation/API Gateway Service/task4_Dependency.md` -> `12. Security & Best Practices`
- `TaskImplementation/API Gateway Service/task6_Dependency.md` -> `12. Security & Best Practices`

---

## 13. Missing or Misconfigured Things

### Audit Findings

| Finding | Impact | Suggested Fix |
|---|---|---|
| No API Gateway Dockerfile found | Gateway cannot be containerized from repo standard yet | Add Dockerfile when deployment task begins |
| No local `docker-compose.yml` found | Observability stack cannot be started by one repo command yet | Add `infra/compose` stack or document external stack |
| No Prometheus config file found | Prometheus cannot scrape Gateway without a config | Add `infra/monitoring/prometheus/prometheus.yml` |
| No OTel Collector config found | Collector pipeline not reproducible yet | Add `infra/monitoring/otel-collector/config.yml` |
| No Grafana dashboard JSON found | Dashboards must be built manually | Add dashboard provisioning later |
| No alert rules found | High latency/error alerts not automated | Add Prometheus alert rules later |
| `TRACE_ENABLED=true` default points to `otel-collector:4317` | Host-local beginners may not have this DNS/service | Set `TRACE_ENABLED=false` for simple local, or use `localhost:4317` |
| `METRICS_PORT=9090` default can conflict with Prometheus UI | Local startup confusion | Use `METRICS_ADDR=:19090` when Prometheus UI uses `9090` |
| `OBSERVABILITY_HASH_SALT` default blank | User hash is less protected/cross-env correlation less controlled | Set non-empty environment-specific salt |
| `task7.md` mentions zap/otelgrpc but current code uses `slog` and custom interceptors | Developers may install unused deps | Keep this dependency guide aligned with actual `go.mod` |
| Metrics endpoint has no app auth | Public exposure can leak operational details | Bind to private network or protect at ingress/network policy |
| gRPC child spans depend on real downstream calls | Current route bridge may only show HTTP spans for routes returning 501 | Add downstream bridge instrumentation as route bridge matures |

### Hardcoded Credentials Check

No hardcoded database credentials were found in Task 7 observability code.

Sensitive config still needs environment handling:

| Item | Current handling | Recommendation |
|---|---|---|
| Redis password | Inherited env `REDIS_PASSWORD` | Keep in `.env` locally, secret manager in prod |
| JWT/JWKS | Inherited env `JWT_JWKS_URL` etc. | HTTPS in non-local environments |
| Observability hash salt | Env `OBSERVABILITY_HASH_SALT`, default blank | Set non-empty secret |
| Grafana admin password | Not used by Gateway | Configure in Grafana deployment secrets |
| OTLP TLS/auth | Only insecure/TLS toggle currently exposed | Add auth/cert config if production collector requires it |

---

## 14. References to Previous Dependency Files

| Reused Topic | Previous File | Section |
|---|---|---|
| Base project overview and full Gateway setup | `TaskImplementation/API Gateway Service/task1_Dependency.md` | `1. Project Overview` |
| Git/Go/Docker installation | `TaskImplementation/API Gateway Service/task1_Dependency.md` | `3. Required Software` |
| Go modules, `go mod download`, `go mod tidy`, `go test`, `go build`, `go run` | `TaskImplementation/API Gateway Service/task1_Dependency.md` | `4. Dependency Management` |
| MySQL, MongoDB, Redis, Typesense, Kafka/RabbitMQ setup | `TaskImplementation/API Gateway Service/task1_Dependency.md` | `5. Database Setup`, `6. Redis / Queue / External Services` |
| Full baseline `.env` | `TaskImplementation/API Gateway Service/task1_Dependency.md` | `7. Environment Variables` |
| Docker setup baseline and missing Dockerfile note | `TaskImplementation/API Gateway Service/task1_Dependency.md` | `8. Docker Setup` |
| Downstream gRPC service addresses and health checks | `TaskImplementation/API Gateway Service/task2_Dependency.md` | `6. Redis / Queue / External Services`, `7. Environment Variables` |
| Optional observability variables previously introduced | `TaskImplementation/API Gateway Service/task2_Dependency.md` | `6. Redis / Queue / External Services` -> `Optional Observability Services` |
| JWT/JWKS setup | `TaskImplementation/API Gateway Service/task3_Dependency.md` | `6. Redis / Queue / External Services`, `7. Environment Variables` |
| Redis rate-limit setup | `TaskImplementation/API Gateway Service/task4_Dependency.md` | `5. Database Setup`, `7. Environment Variables` |
| Request validation env and API contract dependency | `TaskImplementation/API Gateway Service/task5_Dependency.md` | `6. Redis / Queue / External Services`, `7. Environment Variables` |
| Error mapping and error-code observability impact | `TaskImplementation/API Gateway Service/task6_Dependency.md` | `6. Redis / Queue / External Services`, `7. Environment Variables` |
| Platform-wide observability baseline | `TaskImplementation/Platform Foundation/task6.md` | `Observability Architecture`, `Prometheus Metrics Standard`, `OpenTelemetry Trace Standard` |

---

## 15. Final Checklist

Use this checklist for Task 7 observability setup:

```text
[ ] Previous Task 1-6 dependency docs reviewed
[ ] Go compatible with `go 1.26.3`
[ ] `go mod download` completed
[ ] `go test ./internal/observability` passes
[ ] Full Gateway `.env` from previous docs is present
[ ] `.env` exported before running Gateway
[ ] `SERVICE_NAME=api-gateway`
[ ] `APP_ENV` set correctly
[ ] `LOG_LEVEL` intentionally chosen
[ ] `REQUEST_ID_HEADER=X-Request-Id` or another safe header name
[ ] `METRICS_ENABLED=true` for realistic local/dev testing
[ ] Metrics port does not conflict with Prometheus UI
[ ] `METRICS_ADDR=:19090` used locally if Prometheus UI uses `9090`
[ ] `curl http://localhost:<metrics-port>/metrics` works
[ ] `TRACE_ENABLED=false` for simple local mode without collector
[ ] If tracing enabled, `TRACE_EXPORTER_OTLP_ENDPOINT` uses `host:port` format
[ ] If Gateway runs on host, trace endpoint is usually `localhost:4317`
[ ] If Gateway runs in Compose, trace endpoint can be `otel-collector:4317` or `jaeger:4317`
[ ] `TRACE_SAMPLE_RATIO` is between `0` and `1`
[ ] `TRACE_SHUTDOWN_TIMEOUT` is positive
[ ] `OBSERVABILITY_HASH_SALT` is non-empty outside throwaway local testing
[ ] Prometheus scrape config uses route `/metrics`
[ ] Prometheus labels use route templates, not raw ids
[ ] Logs do not include JWT/password/OTP/card/cookie/signature values
[ ] Jaeger/Grafana/Loki are not exposed publicly without protection
[ ] Full runtime dependencies from Task 1-6 are running before `go run ./cmd/server`
```

Final beginner note:

Task 7 ka core setup teen cheezon par depend karta hai: JSON logs stdout par, Prometheus metrics `/metrics` par, and OpenTelemetry traces OTLP endpoint par. Agar local run confusing ho, pehle `TRACE_ENABLED=false` karke metrics/logs verify karo. Uske baad Jaeger ya OTel Collector start karke tracing enable karo.
