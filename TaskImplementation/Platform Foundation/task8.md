# ☸️ Platform Foundation - Task 8: Kubernetes Base Manifests

![Task](https://img.shields.io/badge/Task-Platform%20Foundation%20%238-blue)
![Priority](https://img.shields.io/badge/Priority-P1-orange)
![Dependency](https://img.shields.io/badge/Dependency-Docker%20Setup-yellow)
![Status](https://img.shields.io/badge/Status-Documented-success)
![Guide](https://img.shields.io/badge/Guide-Hinglish-orange)
![Scope](https://img.shields.io/badge/Scope-Kubernetes%20Base%20Manifests-purple)

## 📌 Task Summary

| Field | Detail |
|---|---|
| Task Name | Kubernetes base manifests |
| Source | `docs/01-micro-tasks.md` → `Platform Foundation` → Task 8 |
| Priority | `P1` platform deployment foundation |
| Dependency | Platform Foundation Task 4: Docker Compose local stack and Docker setup |
| Main Goal | Namespace, ConfigMap, Secret, Deployment, Service, aur Ingress templates banana |
| Output Type | Structured implementation guide |
| Not Included | Helm charts, production deploy pipeline, cloud infra provisioning, HPA/PDB/ServiceMonitor actual manifests, service mesh, canary rollout |

> **Simple Hinglish goal:** Har microservice ke liye ek common Kubernetes manifest pattern define karna hai, taaki future me services ko fast and consistent way me deploy kiya ja sake. Is task ka kaam base template banana hai, full production platform deploy karna nahi.

---

## ✅ Final Output Created

```text
TaskImplementation/
└── Platform Foundation/
    ├── task1.md
    ├── task2.md
    ├── task3.md
    ├── task4.md
    ├── task5.md
    ├── task6.md
    ├── task7.md
    └── task8.md
```

### Why this structure?

- `TaskImplementation/` project ke task-wise implementation guides ka central folder hai.
- `Platform Foundation/` folder already present tha, isliye usko keep kiya gaya.
- `task8.md` sirf **Platform Foundation - Task 8** ka guide hai.
- Actual `infra/k8s/` manifest files create nahi kiye gaye, kyunki user ne output me specifically required folder structure aur `task8.md` content manga hai.

---

## 🧭 Implementation Approach

Is guide ko banate time project ke existing docs and previous platform foundation tasks ko base maana gaya:

| Document | Kya use kiya gaya |
|---|---|
| `docs/01-micro-tasks.md` | Task 8 ka exact scope: namespace, config map, secret, deployment, service, ingress templates |
| `docs/02-system-architecture.md` | Kubernetes cluster namespaces: `edge`, `core`, `data`, `observability`, `jobs` |
| `docs/03-folder-structure.md` | `infra/k8s/` folder convention and config load order |
| `docs/06-auth-security.md` | Secrets commit na karne ka rule and Kubernetes Secret usage |
| `docs/11-devops-external-services.md` | Deployment shape, service discovery, Docker image expectations |
| `docs/13-developer-guide.md` | `kubectl`, Docker, health endpoints, config validation, production readiness |
| `TaskImplementation/Platform Foundation/task4.md` | Docker Compose services and dependency names |
| `TaskImplementation/Platform Foundation/task5.md` | API Gateway ports, route boundary, service discovery target examples |
| `TaskImplementation/Platform Foundation/task6.md` | Health, metrics, tracing endpoint expectations |
| `TaskImplementation/Platform Foundation/task7.md` | CI boundary: Docker image build exists, deploy pipeline not included |

---

## 🧱 Task Boundary

### Included in Task 8

- Kubernetes namespace template
- Kubernetes ConfigMap template
- Kubernetes Secret template with safe placeholder values
- Kubernetes Deployment template
- Kubernetes Service template
- Kubernetes Ingress template
- Base folder structure for future `infra/k8s/`
- Kustomize-friendly organization
- API Gateway public ingress example
- Internal service deployment example
- Health probe, resource, label, and security baseline
- Service discovery naming rule
- Local validation and apply commands
- Mermaid architecture and request flow diagrams
- External tools/libraries explanation with install/use commands

### Not Included in Task 8

- Helm chart implementation
- Argo CD, Flux, or GitOps setup
- Production cloud cluster provisioning
- Docker image registry push job
- Full CD pipeline
- Real production secrets
- Managed database provisioning
- Stateful MySQL, MongoDB, Kafka, Redis, or Typesense production manifests
- HPA, PDB, and ServiceMonitor actual files
- Canary deployment or blue/green rollout
- Service mesh, mTLS, NetworkPolicy, or External Secrets Operator implementation

> 🟢 **Rule:** Task 8 base Kubernetes shape define karega. Production hardening later DevOps tasks me add hogi.

---

## 🗂️ Target Kubernetes Folder Structure

Task 8 ke liye recommended future implementation structure:

```text
infra/
└── k8s/
    ├── README.md
    ├── base/
    │   ├── namespaces/
    │   │   ├── edge.yaml
    │   │   ├── core.yaml
    │   │   ├── data.yaml
    │   │   ├── observability.yaml
    │   │   └── jobs.yaml
    │   ├── services/
    │   │   ├── api-gateway/
    │   │   │   ├── configmap.yaml
    │   │   │   ├── secret.example.yaml
    │   │   │   ├── deployment.yaml
    │   │   │   ├── service.yaml
    │   │   │   ├── ingress.yaml
    │   │   │   └── kustomization.yaml
    │   │   └── product-service/
    │   │       ├── configmap.yaml
    │   │       ├── secret.example.yaml
    │   │       ├── deployment.yaml
    │   │       ├── service.yaml
    │   │       └── kustomization.yaml
    │   └── kustomization.yaml
    └── overlays/
        ├── dev/
        │   ├── kustomization.yaml
        │   └── patches/
        │       ├── api-gateway-replicas.yaml
        │       └── product-service-image.yaml
        ├── staging/
        │   └── kustomization.yaml
        └── prod/
            └── kustomization.yaml
```

### Folder Responsibility

| Path | Responsibility |
|---|---|
| `infra/k8s/base/namespaces/` | Cluster namespaces define karta hai |
| `infra/k8s/base/services/<service>/configmap.yaml` | Non-secret runtime config |
| `infra/k8s/base/services/<service>/secret.example.yaml` | Secret key names and local placeholder shape |
| `infra/k8s/base/services/<service>/deployment.yaml` | Pods, image, probes, env, resources |
| `infra/k8s/base/services/<service>/service.yaml` | ClusterIP service discovery |
| `infra/k8s/base/services/<service>/ingress.yaml` | Public HTTP entry point, only edge-facing services ke liye |
| `infra/k8s/base/kustomization.yaml` | Base resources ko bundle karta hai |
| `infra/k8s/overlays/dev/` | Dev-specific image tags, replicas, hostnames |
| `infra/k8s/overlays/staging/` | Staging-specific config |
| `infra/k8s/overlays/prod/` | Production-specific config and stricter patches |

> 🟡 **Note:** Ye target structure hai. Current requested output ke hisaab se actual YAML files create nahi kiye gaye, sirf implementation guide create hua hai.

---

## 🏗️ Kubernetes Architecture

```mermaid
flowchart TB
    Internet[Internet Traffic] --> LB[Cloud Load Balancer]
    LB --> Ingress[Kubernetes Ingress Controller]

    subgraph Edge[edge namespace]
        Ingress --> GatewaySvc[api-gateway Service]
        GatewaySvc --> GatewayPods[api-gateway Deployment Pods]
    end

    subgraph Core[core namespace]
        AuthSvc[auth-service Service] --> AuthPods[auth-service Pods]
        ProductSvc[product-service Service] --> ProductPods[product-service Pods]
        CartSvc[cart-service Service] --> CartPods[cart-service Pods]
        OrderSvc[order-service Service] --> OrderPods[order-service Pods]
    end

    subgraph Data[data namespace]
        Redis[Redis]
        Typesense[Typesense]
        MQ[Kafka or RabbitMQ]
    end

    subgraph Observability[observability namespace]
        Prometheus[Prometheus]
        Jaeger[Jaeger or Tempo]
        Grafana[Grafana]
    end

    GatewayPods --> AuthSvc
    GatewayPods --> ProductSvc
    GatewayPods --> CartSvc
    GatewayPods --> OrderSvc
    GatewayPods --> Redis
    ProductPods --> Typesense
    OrderPods --> MQ
```

### Hinglish Explanation

- Public traffic pehle Load Balancer se Ingress Controller tak aata hai.
- Ingress sirf `api-gateway` ko expose karta hai.
- Business services public nahi honi chahiye. Wo `ClusterIP` services ke through internal DNS se call hoti hain.
- `edge` namespace public-facing layer ke liye hai.
- `core` namespace business microservices ke liye hai.
- `data` namespace dev/staging self-managed infra ke liye hai. Production me managed DB recommended hai.
- `observability` namespace monitoring tools ke liye hai.

---

## 🔁 Manifest Relationship Flow

```mermaid
flowchart LR
    NS[Namespace] --> CM[ConfigMap]
    NS --> SEC[Secret]
    CM --> DEP[Deployment]
    SEC --> DEP
    DEP --> POD[Pods]
    POD --> SVC[Service]
    SVC --> ING[Ingress]

    ING -->|public only| GW[API Gateway]
    SVC -->|internal DNS| Core[Core Services]
```

### Hinglish Explanation

- Namespace pehle banta hai, kyunki baaki resources uske andar deploy honge.
- ConfigMap and Secret app config provide karte hain.
- Deployment pods create karta hai and config consume karta hai.
- Service pods ko stable DNS name deta hai.
- Ingress public HTTP route banata hai. Ye mostly `api-gateway` ke liye use hoga.

---

## 🧰 External Tools / Libraries Used

Task 8 me code library se zyada infrastructure tools use hote hain.

| Tool | What it is | Why used | Install | Use |
|---|---|---|---|---|
| Kubernetes | Container orchestration platform | Deployments, Services, ConfigMaps, Secrets, Ingress manage karne ke liye | Docker Desktop, cloud cluster, Kind, ya Minikube | `kubectl apply -f ...` |
| `kubectl` | Kubernetes CLI | Manifests apply, inspect, logs, rollout status check karne ke liye | `brew install kubectl` or package manager | `kubectl get pods -n core` |
| Kustomize | YAML composition tool | Base and environment overlays manage karne ke liye | `kubectl` me built-in hota hai | `kubectl apply -k infra/k8s/overlays/dev` |
| Kind | Local Kubernetes cluster in Docker | Local validation ke liye lightweight cluster | `go install sigs.k8s.io/kind@latest` or package manager | `kind create cluster` |
| Minikube | Local Kubernetes cluster | Local testing alternative | package manager | `minikube start` |
| NGINX Ingress Controller | Ingress implementation | Public HTTP routing ke liye | provider manifest or Helm | `kubectl get pods -n ingress-nginx` |
| Docker | Container image builder/runtime | Service images build and local cluster image loading ke liye | Docker Desktop or Docker Engine | `docker build ...` |
| kubeconform | Manifest schema validator | CI/local YAML validation ke liye optional tool | package manager or binary | `kubeconform -strict infra/k8s/**/*.yaml` |

> 🟡 **Important:** Helm production packaging ke liye useful hai, but Task 8 me raw YAML plus Kustomize pattern enough hai. Helm chart actual implement nahi kiya gaya.

---

## ⚙️ Install and Use Commands

### 1. `kubectl` check

```bash
kubectl version --client
kubectl config current-context
```

**Explanation:**  
Ye confirm karta hai ki local machine Kubernetes cluster se connect kar sakti hai.

### 2. Local cluster with Kind

```bash
kind create cluster --name ecommerce-dev
kubectl cluster-info
```

**Explanation:**  
Kind Docker ke andar local Kubernetes cluster run karta hai. Base manifests locally validate karne ke liye ye fast option hai.

### 3. Apply base manifests

```bash
kubectl apply -k infra/k8s/base
```

**Explanation:**  
Kustomize base ke saare resources ek saath apply karta hai.

### 4. Apply dev overlay

```bash
kubectl apply -k infra/k8s/overlays/dev
```

**Explanation:**  
Dev overlay base manifests ko reuse karta hai and dev-specific image tag, replica count, aur hostnames set karta hai.

### 5. Validate rollout

```bash
kubectl rollout status deployment/api-gateway -n edge
kubectl get pods -n edge
kubectl get svc -n edge
```

**Explanation:**  
Deployment pods healthy hain ya nahi, ye commands quickly check karti hain.

---

## 🧩 Naming and Label Standards

Har Kubernetes resource me consistent naming and labels hone chahiye.

### Naming rule

| Resource | Format | Example |
|---|---|---|
| Namespace | short domain name | `edge`, `core`, `data` |
| Deployment | service name | `api-gateway`, `product-service` |
| Service | service name | `product-service` |
| ConfigMap | `<service>-config` | `api-gateway-config` |
| Secret | `<service>-secret` | `product-service-secret` |
| Ingress | `<service>-ingress` | `api-gateway-ingress` |

### Common labels

```yaml
labels:
  app.kubernetes.io/name: product-service
  app.kubernetes.io/part-of: ecommerce-platform
  app.kubernetes.io/component: backend-service
  app.kubernetes.io/managed-by: kustomize
```

**Explanation:**  
Labels monitoring, selection, debugging, ownership, and future automation me help karte hain.

---

## 🪪 Step 1: Namespaces Banao

Namespaces cluster ke andar logical isolation provide karte hain.

### `edge` namespace

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: edge
  labels:
    app.kubernetes.io/part-of: ecommerce-platform
    ecommerce.io/tier: edge
```

### `core` namespace

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: core
  labels:
    app.kubernetes.io/part-of: ecommerce-platform
    ecommerce.io/tier: core
```

### Recommended namespaces

| Namespace | Purpose |
|---|---|
| `edge` | Ingress, API Gateway, future Envoy |
| `core` | Business microservices |
| `data` | Local/dev Redis, Typesense, Kafka/RabbitMQ, DB adapters |
| `observability` | Prometheus, Grafana, Loki, Jaeger/Tempo |
| `jobs` | Batch jobs, reindex jobs, cleanup jobs |

**How this part was built:**  
`docs/02-system-architecture.md` me recommended namespace split already defined hai. Task 8 me usko Kubernetes `Namespace` resources ke form me template kiya gaya.

---

## 🟦 Step 2: ConfigMap Template Banao

ConfigMap non-secret runtime config ke liye use hota hai.

### API Gateway ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: api-gateway-config
  namespace: edge
  labels:
    app.kubernetes.io/name: api-gateway
    app.kubernetes.io/part-of: ecommerce-platform
    app.kubernetes.io/component: api-gateway
data:
  SERVICE_NAME: "api-gateway"
  ENVIRONMENT: "dev"
  HTTP_PORT: "8080"
  GRPC_PORT: "9090"
  LOG_LEVEL: "info"
  REDIS_ADDR: "redis.data.svc.cluster.local:6379"
  AUTH_GRPC_ADDR: "auth-service.core.svc.cluster.local:9090"
  USER_GRPC_ADDR: "user-service.core.svc.cluster.local:9090"
  PRODUCT_GRPC_ADDR: "product-service.core.svc.cluster.local:9090"
  CART_GRPC_ADDR: "cart-service.core.svc.cluster.local:9090"
  ORDER_GRPC_ADDR: "order-service.core.svc.cluster.local:9090"
  PAYMENT_GRPC_ADDR: "payment-service.core.svc.cluster.local:9090"
  SEARCH_GRPC_ADDR: "search-service.core.svc.cluster.local:9090"
  CMS_GRPC_ADDR: "cms-service.core.svc.cluster.local:9090"
  SESSION_GRPC_ADDR: "session-service.core.svc.cluster.local:9090"
  SUPERADMIN_GRPC_ADDR: "superadmin-service.core.svc.cluster.local:9090"
  TRACE_EXPORTER_OTLP_ENDPOINT: "otel-collector.observability.svc.cluster.local:4317"
```

### Product Service ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: product-service-config
  namespace: core
  labels:
    app.kubernetes.io/name: product-service
    app.kubernetes.io/part-of: ecommerce-platform
    app.kubernetes.io/component: backend-service
data:
  SERVICE_NAME: "product-service"
  ENVIRONMENT: "dev"
  HTTP_PORT: "8080"
  GRPC_PORT: "9090"
  LOG_LEVEL: "info"
  MONGO_DATABASE: "product_db"
  REDIS_ADDR: "redis.data.svc.cluster.local:6379"
  TYPESENSE_HOST: "typesense.data.svc.cluster.local"
  TYPESENSE_PORT: "8108"
  KAFKA_BROKERS: "kafka.data.svc.cluster.local:9092"
  TRACE_EXPORTER_OTLP_ENDPOINT: "otel-collector.observability.svc.cluster.local:4317"
```

**How this part was built:**  
`docs/03-folder-structure.md` ke config load order me Kubernetes ConfigMap third source hai. Isliye service ke normal config values ConfigMap me rakhe gaye. Passwords, API keys, JWT keys yahan nahi rakhe gaye.

---

## 🔐 Step 3: Secret Template Banao

Secret sensitive values ke liye use hota hai. Real secret values git me commit nahi karne.

### Safe Secret Example

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: product-service-secret
  namespace: core
  labels:
    app.kubernetes.io/name: product-service
    app.kubernetes.io/part-of: ecommerce-platform
    app.kubernetes.io/component: backend-service
type: Opaque
stringData:
  MONGO_URI: "mongodb://product_user:change-me@mongo-product.data.svc.cluster.local:27017/product_db"
  TYPESENSE_API_KEY: "change-me"
```

### API Gateway Secret Example

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: api-gateway-secret
  namespace: edge
  labels:
    app.kubernetes.io/name: api-gateway
    app.kubernetes.io/part-of: ecommerce-platform
    app.kubernetes.io/component: api-gateway
type: Opaque
stringData:
  JWT_PUBLIC_KEY: "replace-with-public-key"
  REDIS_PASSWORD: "change-me"
```

### Safer command for real environments

```bash
kubectl create secret generic product-service-secret \
  --namespace core \
  --from-literal=MONGO_URI='mongodb://real-user:real-pass@host:27017/product_db' \
  --from-literal=TYPESENSE_API_KEY='real-typesense-key'
```

**How this part was built:**  
`docs/06-auth-security.md` clearly bolta hai: "Never commit secrets." Isliye `secret.example.yaml` me sirf placeholder shape define hogi. Real secret `kubectl create secret`, External Secrets Operator, Sealed Secrets, ya cloud secret manager se inject hoga.

> 🔴 **Important:** `stringData` examples learning ke liye readable hain. Production me plain text secret YAML git me commit nahi karna.

---

## 🚀 Step 4: Deployment Template Banao

Deployment pods manage karta hai. Isme image, env, ports, probes, resources, and rolling update strategy define hoti hai.

### API Gateway Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
  namespace: edge
  labels:
    app.kubernetes.io/name: api-gateway
    app.kubernetes.io/part-of: ecommerce-platform
    app.kubernetes.io/component: api-gateway
spec:
  replicas: 2
  revisionHistoryLimit: 5
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app.kubernetes.io/name: api-gateway
  template:
    metadata:
      labels:
        app.kubernetes.io/name: api-gateway
        app.kubernetes.io/part-of: ecommerce-platform
        app.kubernetes.io/component: api-gateway
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/path: "/metrics"
        prometheus.io/port: "8080"
    spec:
      terminationGracePeriodSeconds: 30
      containers:
        - name: api-gateway
          image: registry.example.com/ecommerce/api-gateway:dev
          imagePullPolicy: IfNotPresent
          ports:
            - name: http
              containerPort: 8080
              protocol: TCP
            - name: grpc
              containerPort: 9090
              protocol: TCP
          envFrom:
            - configMapRef:
                name: api-gateway-config
            - secretRef:
                name: api-gateway-secret
          readinessProbe:
            httpGet:
              path: /health/ready
              port: http
            initialDelaySeconds: 5
            periodSeconds: 10
            timeoutSeconds: 2
            failureThreshold: 3
          livenessProbe:
            httpGet:
              path: /health/live
              port: http
            initialDelaySeconds: 15
            periodSeconds: 20
            timeoutSeconds: 2
            failureThreshold: 3
          resources:
            requests:
              cpu: "100m"
              memory: "128Mi"
            limits:
              cpu: "500m"
              memory: "512Mi"
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            runAsNonRoot: true
            runAsUser: 10001
            capabilities:
              drop:
                - ALL
```

### Product Service Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: product-service
  namespace: core
  labels:
    app.kubernetes.io/name: product-service
    app.kubernetes.io/part-of: ecommerce-platform
    app.kubernetes.io/component: backend-service
spec:
  replicas: 2
  revisionHistoryLimit: 5
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app.kubernetes.io/name: product-service
  template:
    metadata:
      labels:
        app.kubernetes.io/name: product-service
        app.kubernetes.io/part-of: ecommerce-platform
        app.kubernetes.io/component: backend-service
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/path: "/metrics"
        prometheus.io/port: "8080"
    spec:
      terminationGracePeriodSeconds: 30
      containers:
        - name: product-service
          image: registry.example.com/ecommerce/product-service:dev
          imagePullPolicy: IfNotPresent
          ports:
            - name: http
              containerPort: 8080
              protocol: TCP
            - name: grpc
              containerPort: 9090
              protocol: TCP
          envFrom:
            - configMapRef:
                name: product-service-config
            - secretRef:
                name: product-service-secret
          readinessProbe:
            httpGet:
              path: /health/ready
              port: http
            initialDelaySeconds: 5
            periodSeconds: 10
            timeoutSeconds: 2
            failureThreshold: 3
          livenessProbe:
            httpGet:
              path: /health/live
              port: http
            initialDelaySeconds: 15
            periodSeconds: 20
            timeoutSeconds: 2
            failureThreshold: 3
          resources:
            requests:
              cpu: "100m"
              memory: "128Mi"
            limits:
              cpu: "500m"
              memory: "512Mi"
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            runAsNonRoot: true
            runAsUser: 10001
            capabilities:
              drop:
                - ALL
```

**How this part was built:**  
`docs/11-devops-external-services.md` me Deployment shape defined hai. `docs/13-developer-guide.md` production readiness checklist me health endpoints, graceful shutdown, config validation, structured logs, metrics, traces mention hain. Isliye Deployment me probes, resources, rolling update, and security context include kiye gaye.

---

## 🩺 Step 5: Health Probes Standard Karo

Task 6 observability baseline ke according services health endpoints expose karengi.

| Endpoint | Probe | Purpose |
|---|---|---|
| `/health/live` | Liveness | Process alive hai ya nahi |
| `/health/ready` | Readiness | Service traffic receive karne ke liye ready hai ya nahi |
| `/metrics` | Prometheus scrape | Metrics collect karne ke liye |

### Probe example

```yaml
readinessProbe:
  httpGet:
    path: /health/ready
    port: http
  initialDelaySeconds: 5
  periodSeconds: 10
  timeoutSeconds: 2
  failureThreshold: 3

livenessProbe:
  httpGet:
    path: /health/live
    port: http
  initialDelaySeconds: 15
  periodSeconds: 20
  timeoutSeconds: 2
  failureThreshold: 3
```

**Explanation:**  
Readiness fail hone par pod Service endpoints se remove ho jayega. Liveness fail hone par Kubernetes pod restart karega.

---

## 🔌 Step 6: Service Template Banao

Service pods ke dynamic IPs ke upar stable DNS name provide karta hai.

### API Gateway Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: api-gateway
  namespace: edge
  labels:
    app.kubernetes.io/name: api-gateway
    app.kubernetes.io/part-of: ecommerce-platform
    app.kubernetes.io/component: api-gateway
spec:
  type: ClusterIP
  selector:
    app.kubernetes.io/name: api-gateway
  ports:
    - name: http
      port: 80
      targetPort: http
      protocol: TCP
    - name: grpc
      port: 9090
      targetPort: grpc
      protocol: TCP
```

### Product Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: product-service
  namespace: core
  labels:
    app.kubernetes.io/name: product-service
    app.kubernetes.io/part-of: ecommerce-platform
    app.kubernetes.io/component: backend-service
spec:
  type: ClusterIP
  selector:
    app.kubernetes.io/name: product-service
  ports:
    - name: http
      port: 8080
      targetPort: http
      protocol: TCP
    - name: grpc
      port: 9090
      targetPort: grpc
      protocol: TCP
```

**How this part was built:**  
Project architecture me internal services gRPC use karte hain. Isliye `grpc` port `9090` expose kiya gaya. Health and metrics HTTP pe honge, isliye `http` port bhi expose kiya gaya.

### Kubernetes DNS pattern

```text
<service-name>.<namespace>.svc.cluster.local:<port>
```

Examples:

```text
auth-service.core.svc.cluster.local:9090
product-service.core.svc.cluster.local:9090
api-gateway.edge.svc.cluster.local:80
```

---

## 🌐 Step 7: Ingress Template Banao

Ingress public HTTP routing ke liye use hota hai. Base scope me API Gateway ko expose karna enough hai.

### API Gateway Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: api-gateway-ingress
  namespace: edge
  labels:
    app.kubernetes.io/name: api-gateway
    app.kubernetes.io/part-of: ecommerce-platform
    app.kubernetes.io/component: api-gateway
  annotations:
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/proxy-body-size: "10m"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "60"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "60"
spec:
  ingressClassName: nginx
  rules:
    - host: api.ecommerce.local
      http:
        paths:
          - path: /api
            pathType: Prefix
            backend:
              service:
                name: api-gateway
                port:
                  name: http
          - path: /health
            pathType: Prefix
            backend:
              service:
                name: api-gateway
                port:
                  name: http
```

### TLS-ready Ingress shape

```yaml
spec:
  tls:
    - hosts:
        - api.ecommerce.example.com
      secretName: api-gateway-tls
  rules:
    - host: api.ecommerce.example.com
      http:
        paths:
          - path: /api
            pathType: Prefix
            backend:
              service:
                name: api-gateway
                port:
                  name: http
```

**How this part was built:**  
`docs/02-system-architecture.md` ke traffic flow me Internet → Load Balancer → Ingress → API Gateway path defined hai. Isliye Ingress only gateway ko expose karta hai. Core services public nahi hain.

> 🟡 **Note:** TLS certificate management, cert-manager, WAF, and production DNS setup Task 8 ke actual scope me nahi hain. Ingress TLS shape future readiness ke liye dikhaya gaya hai.

---

## 🧬 Step 8: Kustomization Banao

Kustomize se base resources and environment overlays cleanly manage hote hain.

### Service-level kustomization

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - configmap.yaml
  - secret.example.yaml
  - deployment.yaml
  - service.yaml
  - ingress.yaml
```

### Base kustomization

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - namespaces/edge.yaml
  - namespaces/core.yaml
  - namespaces/data.yaml
  - namespaces/observability.yaml
  - namespaces/jobs.yaml
  - services/api-gateway
  - services/product-service
```

### Dev overlay

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../base

images:
  - name: registry.example.com/ecommerce/api-gateway
    newTag: dev-local
  - name: registry.example.com/ecommerce/product-service
    newTag: dev-local

patches:
  - path: patches/api-gateway-replicas.yaml
  - path: patches/product-service-replicas.yaml
```

### Replica patch example

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
  namespace: edge
spec:
  replicas: 1
```

**How this part was built:**  
Kustomize same base YAML ko dev, staging, prod me reuse karne deta hai. Task 8 base template banata hai. Environment-specific values overlays me patch honge.

---

## 🛣️ Step 9: Request Flow Verify Karo

```mermaid
sequenceDiagram
    participant Browser as Browser / Frontend
    participant Ingress as K8s Ingress
    participant GatewaySvc as api-gateway Service
    participant GatewayPod as api-gateway Pod
    participant ProductSvc as product-service Service
    participant ProductPod as product-service Pod

    Browser->>Ingress: GET /api/v1/products
    Ingress->>GatewaySvc: Route /api prefix
    GatewaySvc->>GatewayPod: HTTP 8080
    GatewayPod->>ProductSvc: gRPC product-service.core.svc.cluster.local:9090
    ProductSvc->>ProductPod: gRPC 9090
    ProductPod-->>GatewayPod: Product response
    GatewayPod-->>Browser: REST JSON
```

**Explanation:**  
Frontend ko sirf public gateway URL pata hota hai. Gateway internal Kubernetes DNS ke through Product Service ko call karta hai.

---

## 📦 Step 10: Docker Image Strategy Connect Karo

Task 8 Docker setup pe dependent hai. Kubernetes Deployment me image tabhi work karegi jab Docker image build ho chuki ho.

### Image naming pattern

```text
registry.example.com/ecommerce/<service-name>:<git-sha-or-version>
```

Examples:

```text
registry.example.com/ecommerce/api-gateway:dev
registry.example.com/ecommerce/product-service:dev
registry.example.com/ecommerce/order-service:2026.05.19-abc123
```

### Local Kind image load

```bash
docker build -f backend/services/api-gateway/deploy/Dockerfile \
  -t registry.example.com/ecommerce/api-gateway:dev .

kind load docker-image registry.example.com/ecommerce/api-gateway:dev \
  --name ecommerce-dev
```

**Explanation:**  
Local Kind cluster external registry se image pull nahi karta unless registry configured ho. `kind load docker-image` local image cluster nodes me load kar deta hai.

---

## 🔒 Step 11: Security Baseline Add Karo

Base Deployment me minimal pod security standards use hone chahiye.

### Recommended container security context

```yaml
securityContext:
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true
  runAsNonRoot: true
  runAsUser: 10001
  capabilities:
    drop:
      - ALL
```

### Security rules

| Rule | Why |
|---|---|
| Real secrets git me commit mat karo | Credential leak avoid hota hai |
| `runAsNonRoot` use karo | Container compromise impact kam hota hai |
| `readOnlyRootFilesystem` use karo | Runtime tampering reduce hoti hai |
| CPU/memory limits set karo | Noisy pod cluster ko exhaust nahi karega |
| Public ingress sirf gateway ke liye rakho | Internal services direct expose nahi hongi |
| Image tag `latest` avoid karo | Rollback and audit easier hota hai |

---

## 📊 Step 12: Observability Hooks Add Karo

Task 6 ke baseline ke saath compatible annotations:

```yaml
annotations:
  prometheus.io/scrape: "true"
  prometheus.io/path: "/metrics"
  prometheus.io/port: "8080"
```

Environment variables:

```yaml
data:
  LOG_LEVEL: "info"
  TRACE_EXPORTER_OTLP_ENDPOINT: "otel-collector.observability.svc.cluster.local:4317"
```

**Explanation:**  
Pods metrics expose karenge, Prometheus scrape kar sakega, and traces OTLP collector tak ja sakenge. Actual `ServiceMonitor` object Task 8 me implement nahi kiya gaya.

---

## 🧪 Step 13: Local Validation Commands

### Dry-run apply

```bash
kubectl apply -k infra/k8s/base --dry-run=client
```

### Server-side dry-run

```bash
kubectl apply -k infra/k8s/base --dry-run=server
```

### Inspect resources

```bash
kubectl get ns
kubectl get deploy -n edge
kubectl get deploy -n core
kubectl get svc -n core
kubectl describe ingress api-gateway-ingress -n edge
```

### Pod logs

```bash
kubectl logs deployment/api-gateway -n edge
```

### Rollout status

```bash
kubectl rollout status deployment/api-gateway -n edge
kubectl rollout status deployment/product-service -n core
```

### Port forward gateway

```bash
kubectl port-forward svc/api-gateway -n edge 8080:80
curl http://localhost:8080/health/live
```

**Explanation:**  
Port-forward local browser or curl ko cluster ke Service se connect karta hai without public Ingress setup.

---

## 🧰 Step 14: New Service Add Karne Ka Template

Jab koi new service deploy-ready ho, same pattern follow karo.

### Checklist

| Step | Action |
|---|---|
| 1 | `infra/k8s/base/services/<service-name>/` folder banao |
| 2 | `configmap.yaml` add karo |
| 3 | `secret.example.yaml` add karo |
| 4 | `deployment.yaml` add karo |
| 5 | `service.yaml` add karo |
| 6 | Public service hai to `ingress.yaml` add karo |
| 7 | Service-level `kustomization.yaml` update karo |
| 8 | Base `kustomization.yaml` me service folder add karo |
| 9 | Dev overlay me image tag and replica patch add karo |
| 10 | `kubectl apply -k infra/k8s/overlays/dev --dry-run=server` run karo |

### Minimal internal service folder

```text
infra/k8s/base/services/order-service/
├── configmap.yaml
├── secret.example.yaml
├── deployment.yaml
├── service.yaml
└── kustomization.yaml
```

### Public edge service folder

```text
infra/k8s/base/services/api-gateway/
├── configmap.yaml
├── secret.example.yaml
├── deployment.yaml
├── service.yaml
├── ingress.yaml
└── kustomization.yaml
```

---

## 🧭 Environment Overlay Strategy

| Environment | Replica default | Image tag | Ingress host | Secret source |
|---|---:|---|---|---|
| `dev` | 1 | `dev-local` or branch SHA | `api.ecommerce.local` | local Kubernetes Secret |
| `staging` | 2 | release candidate SHA | `api.staging.example.com` | external secret or sealed secret |
| `prod` | 3+ | semantic version or release SHA | `api.example.com` | cloud secret manager |

### Hinglish Explanation

Base same rahega. Dev, staging, prod differences overlays me rahenge. Isse copy-paste YAML drift kam hota hai.

---

## 🧱 Example Complete API Gateway Set

Ye full base shape ek service ke liye dikhata hai.

```yaml
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: api-gateway-config
  namespace: edge
data:
  SERVICE_NAME: "api-gateway"
  ENVIRONMENT: "dev"
  HTTP_PORT: "8080"
  GRPC_PORT: "9090"
  LOG_LEVEL: "info"
  AUTH_GRPC_ADDR: "auth-service.core.svc.cluster.local:9090"
  PRODUCT_GRPC_ADDR: "product-service.core.svc.cluster.local:9090"
---
apiVersion: v1
kind: Secret
metadata:
  name: api-gateway-secret
  namespace: edge
type: Opaque
stringData:
  JWT_PUBLIC_KEY: "replace-with-public-key"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
  namespace: edge
spec:
  replicas: 2
  selector:
    matchLabels:
      app.kubernetes.io/name: api-gateway
  template:
    metadata:
      labels:
        app.kubernetes.io/name: api-gateway
    spec:
      containers:
        - name: api-gateway
          image: registry.example.com/ecommerce/api-gateway:dev
          ports:
            - name: http
              containerPort: 8080
            - name: grpc
              containerPort: 9090
          envFrom:
            - configMapRef:
                name: api-gateway-config
            - secretRef:
                name: api-gateway-secret
          readinessProbe:
            httpGet:
              path: /health/ready
              port: http
          livenessProbe:
            httpGet:
              path: /health/live
              port: http
---
apiVersion: v1
kind: Service
metadata:
  name: api-gateway
  namespace: edge
spec:
  selector:
    app.kubernetes.io/name: api-gateway
  ports:
    - name: http
      port: 80
      targetPort: http
    - name: grpc
      port: 9090
      targetPort: grpc
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: api-gateway-ingress
  namespace: edge
spec:
  ingressClassName: nginx
  rules:
    - host: api.ecommerce.local
      http:
        paths:
          - path: /api
            pathType: Prefix
            backend:
              service:
                name: api-gateway
                port:
                  name: http
```

> 🟡 **Note:** Real repo implementation me inhe separate files me rakhna better hai. Ek file example sirf learning ke liye hai.

---

## 🧩 Config Flow Diagram

```mermaid
flowchart TB
    Defaults[Code Defaults] --> Loader[Service Config Loader]
    EnvFile[Local .env] --> Loader
    CM[Kubernetes ConfigMap] --> PodEnv[Pod Environment]
    Sec[Kubernetes Secret] --> PodEnv
    PodEnv --> Loader
    Loader --> App[Running Service]
```

### Hinglish Explanation

Local development me `.env` useful hai. Kubernetes me ConfigMap and Secret values pod environment me inject hote hain. Service startup pe config validate honi chahiye.

---

## 🔍 Troubleshooting Guide

| Problem | Check | Fix |
|---|---|---|
| Pod `ImagePullBackOff` | `kubectl describe pod` | Image tag correct karo or registry secret add karo |
| Pod `CrashLoopBackOff` | `kubectl logs` | Missing config/secret or startup panic fix karo |
| Readiness fail | `/health/ready` logs | DB/Redis/gRPC dependency connection check karo |
| Service no endpoints | `kubectl get endpoints` | Deployment labels and Service selector match karo |
| Ingress 404 | `kubectl describe ingress` | Host/path and ingress class check karo |
| Secret not found | `kubectl get secret -n <ns>` | Secret create karo ya name match karo |
| Config not updating | `kubectl rollout restart deployment/<name>` | EnvFrom changes pod restart ke baad apply hoti hain |

---

## ✅ Verification Checklist for This Task

| Verification | Status |
|---|---|
| `TaskImplementation/` folder exists | ✅ Done |
| `TaskImplementation/Platform Foundation/` folder kept | ✅ Done |
| `task8.md` created | ✅ Done |
| Guide Hinglish me likha gaya | ✅ Done |
| Step-by-step implementation included | ✅ Done |
| Clean folder structure included | ✅ Done |
| Namespace template included | ✅ Done |
| ConfigMap template included | ✅ Done |
| Secret template included | ✅ Done |
| Deployment template included | ✅ Done |
| Service template included | ✅ Done |
| Ingress template included | ✅ Done |
| Mermaid architecture diagram included | ✅ Done |
| Mermaid request flow diagram included | ✅ Done |
| External tools explained | ✅ Done |
| Install/use commands included | ✅ Done |
| Code examples included | ✅ Done |
| Scope limited to Platform Foundation Task 8 | ✅ Done |
| No Helm chart created | ✅ Done |
| No production deployment pipeline created | ✅ Done |
| No real secrets committed | ✅ Done |

---

## 🚫 Out of Scope for Task 8

Ye items intentionally implement nahi kiye gaye:

- Actual `infra/k8s/` YAML files create karna
- Helm chart setup
- Argo CD or Flux GitOps setup
- Docker image push pipeline
- Dev/staging/prod CD workflow
- Cloud Load Balancer provisioning
- DNS and TLS certificate automation
- External Secrets Operator implementation
- Sealed Secrets implementation
- HPA, PDB, ServiceMonitor actual CRDs
- Stateful production DB manifests
- Kafka/RabbitMQ production cluster setup
- Service mesh and mTLS setup
- Canary rollout, blue/green rollout, or rollback automation

> 🔴 **Reason:** User ne specifically required folder structure aur `task8.md` content generate karne ko bola. Platform Foundation Task 8 ka documentation scope Kubernetes base manifest templates tak limited rakha gaya.

---

## ✅ Final Task 8 Standard

```text
Kubernetes base manifests = Namespace + ConfigMap + Secret + Deployment + Service + Ingress + Kustomize-ready folder structure
```

**Final Hinglish conclusion:**  
Platform Foundation Task 8 ke liye Kubernetes base manifest strategy define ho gayi. Ab future services ke paas ek consistent deploy template hoga: namespace isolation, safe config handling, secret placeholder pattern, rolling Deployment, internal ClusterIP Service, and public API Gateway Ingress. Production deployment automation and advanced Kubernetes hardening intentionally later tasks ke liye reserved hai.
