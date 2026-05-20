# 🐳 Platform Foundation - Task 4: Docker Compose Local Stack

![Task](https://img.shields.io/badge/Task-Platform%20Foundation%20%234-blue)
![Priority](https://img.shields.io/badge/Priority-P0-red)
![Dependency](https://img.shields.io/badge/Dependency-Repo%20Standards-yellow)
![Status](https://img.shields.io/badge/Status-Documented-success)
![Guide](https://img.shields.io/badge/Guide-Hinglish-orange)
![Scope](https://img.shields.io/badge/Scope-Docker%20Compose%20Local%20Stack-purple)

## 📌 Task Summary

| Field | Detail |
|---|---|
| Task Name | Docker Compose local stack |
| Source | `docs/01-micro-tasks.md` → `Platform Foundation` → Task 4 |
| Priority | `P0` foundation/blocker |
| Dependency | Platform Foundation Task 1: Define repo standards |
| Main Goal | MySQL, MongoDB, Redis, Typesense, Kafka/RabbitMQ, Jaeger, aur Prometheus ko local Docker Compose se run karna |
| Output Type | Structured implementation guide |
| Not Included | API Gateway, business microservices, Kubernetes manifests, CI pipeline, production secrets, cloud managed services |

> **Simple Hinglish goal:** Is task ka purpose local development ke liye saari base dependencies ek command se start karna hai. Developer ko MySQL, MongoDB, Redis, search engine, message queue, tracing, aur metrics manually install nahi karne padenge. Sab kuch Docker Compose network ke andar predictable names aur ports ke saath run hoga.

---

## ✅ Final Output Created

```text
TaskImplementation/
└── Platform Foundation/
    ├── task1.md
    ├── task2.md
    ├── task3.md
    └── task4.md
```

### Why this structure?

- `TaskImplementation/` task-wise implementation guides ke liye central folder hai.
- `Platform Foundation/` folder already present tha, isliye usko keep kiya gaya.
- `task4.md` sirf **Platform Foundation - Task 4** ka guide hai.
- Actual `docker-compose.local.yml` file create nahi ki gayi, kyunki requested output sirf folder structure aur `task4.md` content generate karna tha.

---

## 🧭 Implementation Approach

Is guide ko banate time project ke existing documentation ko base banaya gaya:

| Document | Kya use kiya gaya |
|---|---|
| `docs/01-micro-tasks.md` | Task 4 ka exact scope: local stack with MySQL, MongoDB, Redis, Typesense, Kafka/RabbitMQ, Jaeger, Prometheus |
| `docs/02-system-architecture.md` | Services, DB ownership, Redis, message queue, observability flow samjha |
| `docs/03-folder-structure.md` | `infra/compose/`, `infra/docker/`, `infra/monitoring/` style align kiya |
| `docs/04-microservice-design.md` | Kaun service MySQL/Mongo/Redis/MQ use karegi ye mapping samjhi |
| `docs/11-devops-external-services.md` | Local Docker Compose stack aur external services expectations align kiye |
| `docs/12-logging-monitoring-scalability.md` | Prometheus metrics aur Jaeger/OpenTelemetry tracing requirements align kiye |
| `docs/13-developer-guide.md` | Local setup command flow: compose up, migrations, proto generate, services start |

---

## 🧱 Task Boundary

### Included in Task 4

- Local Docker Compose stack ka target folder structure
- Environment variables ka beginner-friendly setup
- MySQL, MongoDB, Redis, Typesense, RabbitMQ, Jaeger, Prometheus service design
- Kafka alternate option ka explanation, kyunki requirement me `Kafka/RabbitMQ` diya gaya hai
- Volumes, network, health checks, ports, and startup commands
- Service-to-dependency connection examples
- Mermaid architecture and flow diagrams
- Troubleshooting checklist

### Not Included in Task 4

- Backend microservice code
- API Gateway implementation
- Actual database migrations
- Product/Search/Auth/Order service implementation
- Kubernetes manifests
- CI/CD pipeline
- Production-grade secret management
- Grafana/Mailpit setup, because Task 4 ke exact scope me required nahi hai

> 🟢 **Rule:** Ye stack local development ke liye hai. Production me managed MySQL, managed MongoDB, managed Redis, managed Kafka/RabbitMQ, managed observability, aur secret manager use karna better rahega.

---

## 🗂️ Target Local Stack Folder Structure

Task 4 ke liye recommended implementation structure:

```text
infra/
└── compose/
    ├── docker-compose.local.yml
    ├── .env.local.example
    └── prometheus/
        └── prometheus.yml
```

### Folder Responsibility

| Path | Responsibility |
|---|---|
| `infra/compose/docker-compose.local.yml` | Local dependency containers define karega |
| `infra/compose/.env.local.example` | Safe dev-only sample environment variables |
| `infra/compose/prometheus/prometheus.yml` | Prometheus scrape configuration |

> 🟡 **Note:** Is guide me ye files ka recommended content diya gaya hai. Actual files create karna next hands-on implementation step hoga.

---

## 🧩 Local Stack Architecture

```mermaid
flowchart TB
    subgraph Compose[Docker Compose Network: ecommerce-local]
        MYSQL[(MySQL)]
        MONGO[(MongoDB)]
        REDIS[(Redis)]
        TYPESENSE[(Typesense)]
        MQ[RabbitMQ<br/>or Kafka]
        JAEGER[Jaeger]
        PROM[Prometheus]
    end

    subgraph FutureServices[Future local backend services]
        GW[API Gateway]
        AUTH[Auth Service]
        USER[User Service]
        PRODUCT[Product Service]
        CART[Cart Service]
        ORDER[Order Service]
        SEARCH[Search Service]
        SESSION[Session Service]
    end

    AUTH --> MYSQL
    USER --> MYSQL
    ORDER --> MYSQL
    PRODUCT --> MONGO
    CART --> MONGO
    SESSION --> MONGO
    GW --> REDIS
    CART --> REDIS
    SEARCH --> TYPESENSE
    PRODUCT --> MQ
    ORDER --> MQ
    SEARCH --> MQ
    GW --> JAEGER
    AUTH --> JAEGER
    PRODUCT --> JAEGER
    PROM --> GW
    PROM --> AUTH
    PROM --> PRODUCT
```

**Hinglish explanation:**  
Docker Compose ek private network banata hai. Future services host port se nahi, service name se connect karengi, jaise `mysql:3306`, `mongo:27017`, `redis:6379`, `typesense:8108`, `rabbitmq:5672`. Isse local aur container-based service config predictable rahega.

---

## 🔌 External Tools and Libraries Used

| Tool/Image | What it is | Why used | Install/Use |
|---|---|---|---|
| Docker | Container runtime | Local dependencies isolated containers me run karne ke liye | Docker Desktop ya Docker Engine install karo |
| Docker Compose | Multi-container orchestration | Ek command se full local stack start/stop karne ke liye | `docker compose version` se verify karo |
| MySQL | Relational database | Auth, User, Order, Payment, CMS, Superadmin structured data ke liye | Compose service `mysql` |
| MongoDB | Document database | Product, Cart, Wishlist, Session, Notification flexible data ke liye | Compose service `mongo` |
| Redis | In-memory store/cache | Rate limits, active sessions, cart cache, OTP counters ke liye | Compose service `redis` |
| Typesense | Search engine | Product search, facets, typo-tolerant queries ke liye | Compose service `typesense` |
| RabbitMQ | Message broker | Local event-driven workflows ke liye simple queue setup | Compose service `rabbitmq` |
| Kafka | Event streaming platform | High-throughput event streaming ke liye alternate option | Optional alternate profile |
| Jaeger | Distributed tracing UI/backend | Request trace dekhne ke liye | Compose service `jaeger` |
| Prometheus | Metrics database/scraper | Service metrics scrape/store karne ke liye | Compose service `prometheus` |

### Installation check

```bash
docker --version
docker compose version
```

### Basic usage commands

```bash
# local stack start
docker compose -f infra/compose/docker-compose.local.yml --env-file infra/compose/.env.local up -d

# running containers check
docker compose -f infra/compose/docker-compose.local.yml ps

# logs check
docker compose -f infra/compose/docker-compose.local.yml logs -f

# local stack stop
docker compose -f infra/compose/docker-compose.local.yml down

# local stack stop + data delete
docker compose -f infra/compose/docker-compose.local.yml down -v
```

> 🔴 **Warning:** `down -v` volumes delete karta hai. Isse MySQL, MongoDB, Redis, Typesense, RabbitMQ, Prometheus ka local data remove ho jayega.

---

## 🪜 Step-by-Step Implementation

## Step 1: Compose directory create karo

Recommended path:

```text
infra/compose/
```

Command:

```bash
mkdir -p infra/compose/prometheus
```

**Explanation:**  
Infrastructure related local files ko `infra/compose/` me rakhna repo standard ke saath aligned hai. Future me `infra/k8s/`, `infra/helm/`, aur `infra/monitoring/` alag grow kar sakte hain.

---

## Step 2: Local environment file define karo

Recommended file:

```text
infra/compose/.env.local.example
```

Example:

```dotenv
COMPOSE_PROJECT_NAME=ecommerce

MYSQL_ROOT_PASSWORD=dev_root_password
MYSQL_DATABASE=ecommerce_dev
MYSQL_USER=ecommerce_user
MYSQL_PASSWORD=ecommerce_password

MONGO_INITDB_ROOT_USERNAME=ecommerce_root
MONGO_INITDB_ROOT_PASSWORD=ecommerce_password

REDIS_PASSWORD=dev_redis_password

TYPESENSE_API_KEY=dev-typesense-key

RABBITMQ_DEFAULT_USER=ecommerce
RABBITMQ_DEFAULT_PASS=ecommerce_password
RABBITMQ_DEFAULT_VHOST=ecommerce
```

Local developer actual env file banayega:

```bash
cp infra/compose/.env.local.example infra/compose/.env.local
```

**Explanation:**  
`.env.local.example` commit-safe sample file hoti hai. Actual `.env.local` local machine-specific hoti hai aur normally git ignore me rahegi.

> 🟡 **Security note:** Ye credentials sirf local development ke liye hain. Production secrets kabhi repo me commit nahi karne.

---

## Step 3: Docker Compose network aur volumes define karo

Compose file ke bottom me common network and volumes define hote hain:

```yaml
networks:
  ecommerce-local:
    name: ecommerce-local
    driver: bridge

volumes:
  mysql_data:
  mongo_data:
  redis_data:
  typesense_data:
  rabbitmq_data:
  prometheus_data:
```

**Explanation:**  
Network services ko ek dusre ke naam se connect karne deta hai. Volumes container restart ke baad bhi local data safe rakhte hain.

---

## Step 4: MySQL service add karo

MySQL structured transactional data ke liye use hoga.

```yaml
mysql:
  image: mysql:8.4
  container_name: ecommerce-mysql
  restart: unless-stopped
  environment:
    MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
    MYSQL_DATABASE: ${MYSQL_DATABASE}
    MYSQL_USER: ${MYSQL_USER}
    MYSQL_PASSWORD: ${MYSQL_PASSWORD}
  ports:
    - "3306:3306"
  volumes:
    - mysql_data:/var/lib/mysql
  healthcheck:
    test: ["CMD-SHELL", "mysqladmin ping -h 127.0.0.1 -uroot -p$${MYSQL_ROOT_PASSWORD}"]
    interval: 10s
    timeout: 5s
    retries: 10
  networks:
    - ecommerce-local
```

### MySQL used by

| Service | Data |
|---|---|
| Auth Service | accounts, credentials, refresh tokens, roles |
| User Service | users, addresses, seller profiles |
| Order Service | orders, order items, fulfillment |
| Payment Service | payments, refunds, webhook events |
| CMS Service | coupons, offers, seller settings |
| Superadmin Service | admin actions, permissions, settings |

### Connection example

```dotenv
MYSQL_DSN=ecommerce_user:ecommerce_password@tcp(mysql:3306)/ecommerce_dev?parseTime=true
```

**Hinglish explanation:**  
Container ke andar se host `mysql` hoga, kyunki Compose service name DNS name ban jata hai. Host machine se connect karna ho to `localhost:3306` use kar sakte ho.

---

## Step 5: MongoDB service add karo

MongoDB flexible document data ke liye use hoga.

```yaml
mongo:
  image: mongo:7
  container_name: ecommerce-mongo
  restart: unless-stopped
  environment:
    MONGO_INITDB_ROOT_USERNAME: ${MONGO_INITDB_ROOT_USERNAME}
    MONGO_INITDB_ROOT_PASSWORD: ${MONGO_INITDB_ROOT_PASSWORD}
  ports:
    - "27017:27017"
  volumes:
    - mongo_data:/data/db
  healthcheck:
    test: ["CMD-SHELL", "mongosh --quiet --eval 'db.adminCommand({ ping: 1 }).ok' || exit 1"]
    interval: 10s
    timeout: 5s
    retries: 10
  networks:
    - ecommerce-local
```

### MongoDB used by

| Service | Data |
|---|---|
| Product Service | products, categories, brands, inventory snapshots |
| Cart Service | carts, cart audit events |
| Wishlist Service | wishlists, wishlist events |
| Session Management Service | session events, journey data |
| Recommendation Service | recommendation features and lists |
| Notification Service | templates and delivery logs |

### Connection example

```dotenv
MONGO_URI=mongodb://ecommerce_root:ecommerce_password@mongo:27017/ecommerce_dev?authSource=admin
```

**Hinglish explanation:**  
Product attributes dynamic ho sakte hain, jaise fashion aur electronics ka schema different hota hai. Isliye MongoDB flexible document model provide karta hai.

---

## Step 6: Redis service add karo

Redis fast cache, counters, locks, and active session state ke liye use hoga.

```yaml
redis:
  image: redis:7.2-alpine
  container_name: ecommerce-redis
  restart: unless-stopped
  command: ["redis-server", "--appendonly", "yes", "--requirepass", "${REDIS_PASSWORD}"]
  ports:
    - "6379:6379"
  volumes:
    - redis_data:/data
  healthcheck:
    test: ["CMD-SHELL", "redis-cli -a $${REDIS_PASSWORD} ping | grep PONG"]
    interval: 10s
    timeout: 5s
    retries: 10
  environment:
    REDIS_PASSWORD: ${REDIS_PASSWORD}
  networks:
    - ecommerce-local
```

### Redis used by

| Use Case | Example |
|---|---|
| Rate limiting | API Gateway per IP/user limits |
| Active sessions | Session Management active session lookup |
| Cart cache | Cart badge count and hot cart state |
| OTP retry counters | Auth OTP resend/verify limits |
| Recommendation cache | Trending/personalized list cache |

### Connection example

```dotenv
REDIS_ADDR=redis:6379
REDIS_PASSWORD=dev_redis_password
```

**Hinglish explanation:**  
Redis ko primary database ki tarah use nahi karna. Ye fast temporary/hot data ke liye hai. Durable source MySQL ya MongoDB rahega.

---

## Step 7: Typesense service add karo

Typesense product search ke liye local search engine hoga.

```yaml
typesense:
  image: typesense/typesense:27.1
  container_name: ecommerce-typesense
  restart: unless-stopped
  command:
    - "--data-dir=/data"
    - "--api-key=${TYPESENSE_API_KEY}"
    - "--enable-cors"
  ports:
    - "8108:8108"
  volumes:
    - typesense_data:/data
  healthcheck:
    test: ["CMD-SHELL", "wget -q -O - http://localhost:8108/health | grep true"]
    interval: 10s
    timeout: 5s
    retries: 10
  networks:
    - ecommerce-local
```

### Typesense used by

| Service | Purpose |
|---|---|
| Search Service | Product search queries |
| Product indexer | Product events consume karke search document upsert/delete |
| CMS/Superadmin later | Synonyms, merchandising, search config |

### Connection example

```dotenv
TYPESENSE_HOST=typesense
TYPESENSE_PORT=8108
TYPESENSE_PROTOCOL=http
TYPESENSE_API_KEY=dev-typesense-key
```

**Hinglish explanation:**  
Product Service canonical product data own karega. Search Service Typesense me indexed searchable copy rakhega. Product update event ke baad Search Service index update karega.

---

## Step 8: Message queue choose karo

Task requirement me `Kafka/RabbitMQ` diya gaya hai. Local beginner-friendly stack ke liye **RabbitMQ default** choose kiya gaya hai.

### Why RabbitMQ default?

| Reason | Explanation |
|---|---|
| Simple local setup | Single container me broker + management UI mil jata hai |
| Easy debugging | Browser UI se queues, exchanges, bindings dekh sakte hain |
| Good for MVP | Notification commands, indexing jobs, async events start karne ke liye enough |
| Kafka compatible future | Later high-throughput event streaming ke liye Kafka add/replace kar sakte hain |

### RabbitMQ service add karo

```yaml
rabbitmq:
  image: rabbitmq:3.13-management-alpine
  container_name: ecommerce-rabbitmq
  restart: unless-stopped
  environment:
    RABBITMQ_DEFAULT_USER: ${RABBITMQ_DEFAULT_USER}
    RABBITMQ_DEFAULT_PASS: ${RABBITMQ_DEFAULT_PASS}
    RABBITMQ_DEFAULT_VHOST: ${RABBITMQ_DEFAULT_VHOST}
  ports:
    - "5672:5672"
    - "15672:15672"
  volumes:
    - rabbitmq_data:/var/lib/rabbitmq
  healthcheck:
    test: ["CMD-SHELL", "rabbitmq-diagnostics -q ping"]
    interval: 10s
    timeout: 5s
    retries: 10
  networks:
    - ecommerce-local
```

### RabbitMQ URLs

```dotenv
RABBITMQ_URL=amqp://ecommerce:ecommerce_password@rabbitmq:5672/ecommerce
```

| UI | URL |
|---|---|
| RabbitMQ Management UI | `http://localhost:15672` |
| Username | `ecommerce` |
| Password | `ecommerce_password` |

### Event queue examples

| Exchange/Queue | Producer | Consumer |
|---|---|---|
| `product.events` | Product Service | Search, Recommendation |
| `order.events` | Order Service | Notification, Analytics |
| `payment.events` | Payment Service | Order, Notification |
| `session.events` | Session Service | Recommendation, Analytics |
| `notification.commands` | Auth/Order/Payment | Notification Service |

---

## Step 9: Kafka alternate option document karo

Kafka high-throughput event streaming ke liye better fit hai. Agar team Kafka choose kare, RabbitMQ ke place par Kafka profile use kiya ja sakta hai.

### Kafka kab choose kare?

| Situation | Kafka fit |
|---|---|
| High event volume | Session events, product views, recommendation pipeline |
| Replay required | Past events replay karke projections rebuild karne hain |
| Multiple consumer groups | Search, analytics, recommendation alag-alag consume karenge |
| Long-term event stream | Event history retain karni hai |

### Local Kafka profile example

```yaml
kafka:
  image: bitnami/kafka:3.7
  container_name: ecommerce-kafka
  restart: unless-stopped
  profiles:
    - kafka
  environment:
    KAFKA_ENABLE_KRAFT: "yes"
    KAFKA_CFG_NODE_ID: "1"
    KAFKA_CFG_PROCESS_ROLES: "broker,controller"
    KAFKA_CFG_CONTROLLER_LISTENER_NAMES: "CONTROLLER"
    KAFKA_CFG_LISTENERS: "PLAINTEXT://:9092,CONTROLLER://:9093"
    KAFKA_CFG_ADVERTISED_LISTENERS: "PLAINTEXT://kafka:9092"
    KAFKA_CFG_CONTROLLER_QUORUM_VOTERS: "1@kafka:9093"
    ALLOW_PLAINTEXT_LISTENER: "yes"
  ports:
    - "9092:9092"
  networks:
    - ecommerce-local
```

Start command:

```bash
docker compose -f infra/compose/docker-compose.local.yml --profile kafka up -d kafka
```

Connection example:

```dotenv
KAFKA_BROKERS=kafka:9092
```

> 🟡 **Decision:** Task 4 local baseline me RabbitMQ default rakha gaya hai. Kafka alternate documented hai, kyunki requirement me `Kafka/RabbitMQ` choice hai.

---

## Step 10: Jaeger tracing service add karo

Jaeger distributed traces view karne ke liye use hoga.

```yaml
jaeger:
  image: jaegertracing/all-in-one:1.56
  container_name: ecommerce-jaeger
  restart: unless-stopped
  environment:
    COLLECTOR_OTLP_ENABLED: "true"
  ports:
    - "16686:16686"
    - "4317:4317"
    - "4318:4318"
  networks:
    - ecommerce-local
```

### Jaeger URLs and ports

| Purpose | Value |
|---|---|
| UI | `http://localhost:16686` |
| OTLP gRPC | `jaeger:4317` |
| OTLP HTTP | `http://jaeger:4318` |

### Service tracing env example

```dotenv
OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger:4318
OTEL_SERVICE_NAME=product-service
OTEL_TRACES_EXPORTER=otlp
```

**Hinglish explanation:**  
Request Gateway se Product Service, Order Service, Payment Service tak jaata hai. Jaeger trace dikhata hai ki request me kaun sa step slow ya failed hua.

---

## Step 11: Prometheus service add karo

Prometheus metrics scrape aur store karega.

```yaml
prometheus:
  image: prom/prometheus:v2.52.0
  container_name: ecommerce-prometheus
  restart: unless-stopped
  command:
    - "--config.file=/etc/prometheus/prometheus.yml"
    - "--storage.tsdb.path=/prometheus"
  ports:
    - "9090:9090"
  volumes:
    - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
    - prometheus_data:/prometheus
  networks:
    - ecommerce-local
```

### Prometheus config

Recommended file:

```text
infra/compose/prometheus/prometheus.yml
```

Example:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: prometheus
    static_configs:
      - targets:
          - prometheus:9090

  # Future backend services jab implement honge, tab ye targets enable honge.
  # - job_name: ecommerce-services
  #   metrics_path: /metrics
  #   static_configs:
  #     - targets:
  #         - api-gateway:8080
  #         - auth-service:8080
  #         - product-service:8080
```

### Prometheus URL

```text
http://localhost:9090
```

**Hinglish explanation:**  
Task 4 me Prometheus container ready hoga. Actual service metrics tab visible honge jab backend services `/metrics` expose karenge. Prometheus ka self-scrape immediately verify kar sakte hain.

---

## Step 12: Complete Compose file ka recommended shape

Recommended file:

```text
infra/compose/docker-compose.local.yml
```

```yaml
services:
  mysql:
    image: mysql:8.4
    container_name: ecommerce-mysql
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
      MYSQL_DATABASE: ${MYSQL_DATABASE}
      MYSQL_USER: ${MYSQL_USER}
      MYSQL_PASSWORD: ${MYSQL_PASSWORD}
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD-SHELL", "mysqladmin ping -h 127.0.0.1 -uroot -p$${MYSQL_ROOT_PASSWORD}"]
      interval: 10s
      timeout: 5s
      retries: 10
    networks:
      - ecommerce-local

  mongo:
    image: mongo:7
    container_name: ecommerce-mongo
    restart: unless-stopped
    environment:
      MONGO_INITDB_ROOT_USERNAME: ${MONGO_INITDB_ROOT_USERNAME}
      MONGO_INITDB_ROOT_PASSWORD: ${MONGO_INITDB_ROOT_PASSWORD}
    ports:
      - "27017:27017"
    volumes:
      - mongo_data:/data/db
    healthcheck:
      test: ["CMD-SHELL", "mongosh --quiet --eval 'db.adminCommand({ ping: 1 }).ok' || exit 1"]
      interval: 10s
      timeout: 5s
      retries: 10
    networks:
      - ecommerce-local

  redis:
    image: redis:7.2-alpine
    container_name: ecommerce-redis
    restart: unless-stopped
    command: ["redis-server", "--appendonly", "yes", "--requirepass", "${REDIS_PASSWORD}"]
    environment:
      REDIS_PASSWORD: ${REDIS_PASSWORD}
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD-SHELL", "redis-cli -a $${REDIS_PASSWORD} ping | grep PONG"]
      interval: 10s
      timeout: 5s
      retries: 10
    networks:
      - ecommerce-local

  typesense:
    image: typesense/typesense:27.1
    container_name: ecommerce-typesense
    restart: unless-stopped
    command:
      - "--data-dir=/data"
      - "--api-key=${TYPESENSE_API_KEY}"
      - "--enable-cors"
    ports:
      - "8108:8108"
    volumes:
      - typesense_data:/data
    healthcheck:
      test: ["CMD-SHELL", "wget -q -O - http://localhost:8108/health | grep true"]
      interval: 10s
      timeout: 5s
      retries: 10
    networks:
      - ecommerce-local

  rabbitmq:
    image: rabbitmq:3.13-management-alpine
    container_name: ecommerce-rabbitmq
    restart: unless-stopped
    environment:
      RABBITMQ_DEFAULT_USER: ${RABBITMQ_DEFAULT_USER}
      RABBITMQ_DEFAULT_PASS: ${RABBITMQ_DEFAULT_PASS}
      RABBITMQ_DEFAULT_VHOST: ${RABBITMQ_DEFAULT_VHOST}
    ports:
      - "5672:5672"
      - "15672:15672"
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq
    healthcheck:
      test: ["CMD-SHELL", "rabbitmq-diagnostics -q ping"]
      interval: 10s
      timeout: 5s
      retries: 10
    networks:
      - ecommerce-local

  jaeger:
    image: jaegertracing/all-in-one:1.56
    container_name: ecommerce-jaeger
    restart: unless-stopped
    environment:
      COLLECTOR_OTLP_ENABLED: "true"
    ports:
      - "16686:16686"
      - "4317:4317"
      - "4318:4318"
    networks:
      - ecommerce-local

  prometheus:
    image: prom/prometheus:v2.52.0
    container_name: ecommerce-prometheus
    restart: unless-stopped
    command:
      - "--config.file=/etc/prometheus/prometheus.yml"
      - "--storage.tsdb.path=/prometheus"
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus_data:/prometheus
    networks:
      - ecommerce-local

networks:
  ecommerce-local:
    name: ecommerce-local
    driver: bridge

volumes:
  mysql_data:
  mongo_data:
  redis_data:
  typesense_data:
  rabbitmq_data:
  prometheus_data:
```

---

## Step 13: Startup flow define karo

```mermaid
sequenceDiagram
    participant Dev as Developer
    participant Compose as Docker Compose
    participant DB as Databases
    participant MQ as Message Queue
    participant Obs as Observability

    Dev->>Compose: docker compose up -d
    Compose->>DB: Start MySQL, MongoDB, Redis, Typesense
    Compose->>MQ: Start RabbitMQ
    Compose->>Obs: Start Jaeger and Prometheus
    Compose->>Compose: Create network and volumes
    Dev->>Compose: docker compose ps
    Compose-->>Dev: Show running/healthy containers
```

Start command:

```bash
docker compose -f infra/compose/docker-compose.local.yml --env-file infra/compose/.env.local up -d
```

Check status:

```bash
docker compose -f infra/compose/docker-compose.local.yml ps
```

Check logs:

```bash
docker compose -f infra/compose/docker-compose.local.yml logs -f mysql
docker compose -f infra/compose/docker-compose.local.yml logs -f mongo
docker compose -f infra/compose/docker-compose.local.yml logs -f rabbitmq
```

---

## Step 14: Local URLs and ports document karo

| Component | Container Name | Internal Host | Host URL/Port |
|---|---|---|---|
| MySQL | `ecommerce-mysql` | `mysql:3306` | `localhost:3306` |
| MongoDB | `ecommerce-mongo` | `mongo:27017` | `localhost:27017` |
| Redis | `ecommerce-redis` | `redis:6379` | `localhost:6379` |
| Typesense | `ecommerce-typesense` | `typesense:8108` | `http://localhost:8108` |
| RabbitMQ | `ecommerce-rabbitmq` | `rabbitmq:5672` | `localhost:5672` |
| RabbitMQ UI | `ecommerce-rabbitmq` | `rabbitmq:15672` | `http://localhost:15672` |
| Jaeger UI | `ecommerce-jaeger` | `jaeger:16686` | `http://localhost:16686` |
| Jaeger OTLP HTTP | `ecommerce-jaeger` | `jaeger:4318` | `http://localhost:4318` |
| Prometheus | `ecommerce-prometheus` | `prometheus:9090` | `http://localhost:9090` |

> 🟡 **Port conflict rule:** Agar local machine par MySQL ya Redis already installed hai, to Compose file me left-side host port change karo. Example: `"3307:3306"` means host pe `localhost:3307`, container network pe still `mysql:3306`.

---

## Step 15: Future backend services ke env examples define karo

Jab backend services implement honge, unke local `.env` me dependencies aise point karenge:

```dotenv
APP_ENV=local

MYSQL_DSN=ecommerce_user:ecommerce_password@tcp(mysql:3306)/ecommerce_dev?parseTime=true
MONGO_URI=mongodb://ecommerce_root:ecommerce_password@mongo:27017/ecommerce_dev?authSource=admin

REDIS_ADDR=redis:6379
REDIS_PASSWORD=dev_redis_password

TYPESENSE_HOST=typesense
TYPESENSE_PORT=8108
TYPESENSE_PROTOCOL=http
TYPESENSE_API_KEY=dev-typesense-key

RABBITMQ_URL=amqp://ecommerce:ecommerce_password@rabbitmq:5672/ecommerce

OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger:4318
PROMETHEUS_METRICS_PATH=/metrics
```

**Hinglish explanation:**  
Ye env values container-to-container communication ke liye hain. Agar service directly host machine pe run ho rahi ho, to hosts `localhost` use karne padenge. Agar service bhi Docker Compose me run ho rahi ho, to hosts `mysql`, `mongo`, `redis`, etc use honge.

---

## Step 16: Dependency ownership map banao

```mermaid
flowchart LR
    AUTH[Auth Service] --> MYSQL[(MySQL)]
    USER[User Service] --> MYSQL
    ORDER[Order Service] --> MYSQL
    PAYMENT[Payment Service] --> MYSQL
    CMS[CMS Service] --> MYSQL
    ADMIN[Superadmin Service] --> MYSQL

    PRODUCT[Product Service] --> MONGO[(MongoDB)]
    CART[Cart Service] --> MONGO
    WISHLIST[Wishlist Service] --> MONGO
    SESSION[Session Service] --> MONGO
    RECO[Recommendation Service] --> MONGO
    NOTIF[Notification Service] --> MONGO

    GATEWAY[API Gateway] --> REDIS[(Redis)]
    CART --> REDIS
    AUTH --> REDIS
    SESSION --> REDIS

    SEARCH[Search Service] --> TYPESENSE[(Typesense)]

    PRODUCT --> MQ[RabbitMQ/Kafka]
    ORDER --> MQ
    PAYMENT --> MQ
    SESSION --> MQ
    MQ --> SEARCH
    MQ --> RECO
    MQ --> NOTIF

    GATEWAY --> JAEGER[Jaeger]
    AUTH --> JAEGER
    PRODUCT --> JAEGER
    PROM[Prometheus] --> GATEWAY
    PROM --> AUTH
    PROM --> PRODUCT
```

### Important rule

| Rule | Explanation |
|---|---|
| No cross-service DB access | Ek service dusri service ke DB ko directly read/write nahi karegi |
| Data chahiye to API use karo | Sync data ke liye gRPC call |
| Async update ke liye events | Product update → Search index update |
| Redis durable source nahi | Redis cache/session/hot state ke liye hai |
| Typesense canonical DB nahi | Search index copy hai, source of truth Product Service rahega |

---

## Step 17: Health verification commands

### MySQL

```bash
docker exec ecommerce-mysql mysqladmin ping -h 127.0.0.1 -uroot -pdev_root_password
```

Expected:

```text
mysqld is alive
```

### MongoDB

```bash
docker exec ecommerce-mongo mongosh --quiet --eval "db.adminCommand({ ping: 1 })"
```

Expected:

```json
{ "ok": 1 }
```

### Redis

```bash
docker exec ecommerce-redis redis-cli -a dev_redis_password ping
```

Expected:

```text
PONG
```

### Typesense

```bash
curl http://localhost:8108/health
```

Expected:

```json
{ "ok": true }
```

### RabbitMQ

```bash
docker exec ecommerce-rabbitmq rabbitmq-diagnostics -q ping
```

Expected:

```text
Ping succeeded
```

### Jaeger

```bash
curl http://localhost:16686
```

Expected:

```text
Jaeger UI HTML response
```

### Prometheus

```bash
curl http://localhost:9090/-/ready
```

Expected:

```text
Prometheus Server is Ready.
```

---

## Step 18: Beginner-friendly local workflow

```mermaid
flowchart TD
    A[Clone repo] --> B[Copy .env.local.example]
    B --> C[Run docker compose up -d]
    C --> D[Check docker compose ps]
    D --> E[Run migrations later]
    E --> F[Start backend services later]
    F --> G[Use Jaeger and Prometheus for debugging]
```

### Daily commands

```bash
# Start dependencies
docker compose -f infra/compose/docker-compose.local.yml --env-file infra/compose/.env.local up -d

# Stop dependencies without deleting data
docker compose -f infra/compose/docker-compose.local.yml down

# Restart one dependency
docker compose -f infra/compose/docker-compose.local.yml restart redis

# View one dependency log
docker compose -f infra/compose/docker-compose.local.yml logs -f typesense
```

### Reset local data

```bash
docker compose -f infra/compose/docker-compose.local.yml down -v
docker compose -f infra/compose/docker-compose.local.yml --env-file infra/compose/.env.local up -d
```

**Hinglish explanation:**  
Normal daily work me `down` enough hai. Jab local DB corrupt ho jaye ya fresh start chahiye tabhi `down -v` use karo.

---

## 🧪 Smoke Test Checklist

| Check | Command/URL | Expected |
|---|---|---|
| Compose config valid | `docker compose -f infra/compose/docker-compose.local.yml config` | No YAML error |
| Containers running | `docker compose -f infra/compose/docker-compose.local.yml ps` | All required services up |
| MySQL healthy | `docker exec ecommerce-mysql mysqladmin ping ...` | `mysqld is alive` |
| Mongo healthy | `docker exec ecommerce-mongo mongosh ...` | `{ ok: 1 }` |
| Redis healthy | `docker exec ecommerce-redis redis-cli ... ping` | `PONG` |
| Typesense healthy | `curl http://localhost:8108/health` | `{ "ok": true }` |
| RabbitMQ UI | `http://localhost:15672` | Login page |
| Jaeger UI | `http://localhost:16686` | Jaeger dashboard |
| Prometheus UI | `http://localhost:9090` | Prometheus dashboard |

---

## 🛠️ Troubleshooting

| Problem | Reason | Fix |
|---|---|---|
| Port already in use | Local MySQL/Redis/Mongo already running | Compose host port change karo, e.g. `"3307:3306"` |
| Container keeps restarting | Env missing ya invalid command | `docker compose logs -f <service>` run karo |
| MySQL auth fail | Wrong username/password | `.env.local` values and DSN match karo |
| Mongo auth fail | `authSource=admin` missing | Mongo URI me `?authSource=admin` add karo |
| Redis `NOAUTH` error | Password required | Client config me `REDIS_PASSWORD` set karo |
| Typesense unauthorized | Wrong API key | `TYPESENSE_API_KEY` same rakho |
| RabbitMQ login fail | Wrong default user/pass | Volume reset karo if old credentials stored hain |
| Prometheus no service targets | Backend services not implemented yet | Task 4 me self-scrape enough hai |
| Jaeger empty | No service traces sent yet | Backend OpenTelemetry setup ke baad traces aayenge |

### RabbitMQ credentials reset note

RabbitMQ first boot pe credentials volume me save kar deta hai. Agar `.env.local` change karne ke baad login fail ho:

```bash
docker compose -f infra/compose/docker-compose.local.yml down -v
docker compose -f infra/compose/docker-compose.local.yml --env-file infra/compose/.env.local up -d rabbitmq
```

---

## 🔐 Local Security Rules

| Rule | Why |
|---|---|
| `.env.local` commit mat karo | Local secrets repo me leak ho sakte hain |
| Dev credentials production me use mat karo | Predictable passwords unsafe hote hain |
| Public network pe ports expose mat karo | Local DBs internet accessible nahi hone chahiye |
| Logs me secrets print mat karo | Password/API key leak risk |
| Volumes cleanup carefully karo | `down -v` local data permanently delete karta hai |

---

## 📦 Clean Implementation Summary

Task 4 local stack ka final conceptual output:

```text
infra/
└── compose/
    ├── .env.local.example
    ├── docker-compose.local.yml
    └── prometheus/
        └── prometheus.yml

Docker Compose services:
├── mysql       → relational data
├── mongo       → document data
├── redis       → cache/session/rate limit
├── typesense   → product search
├── rabbitmq    → local event queue
├── jaeger      → distributed tracing
└── prometheus  → metrics scraping/storage
```

---

## ✅ Definition of Done

| Requirement | Status |
|---|---|
| `TaskImplementation/` folder exists | ✅ Done |
| `TaskImplementation/Platform Foundation/` folder kept | ✅ Done |
| `task4.md` created | ✅ Done |
| Step-by-step Hinglish implementation guide | ✅ Done |
| External tools/libraries explained | ✅ Done |
| Folder structure included | ✅ Done |
| Code examples included | ✅ Done |
| Mermaid diagrams included | ✅ Done |
| Scope limited to Platform Foundation Task 4 | ✅ Done |

---

## 🧠 Final Notes

- Ye task foundation dependency layer ko local machine par reproducible banata hai.
- Future backend services is stack ko use karke DB, cache, queue, tracing, aur metrics integration test kar sakenge.
- RabbitMQ local default hai for simplicity. Kafka alternate documented hai for future high-throughput streaming use cases.
- Prometheus and Jaeger containers ready rahenge, lekin real metrics/traces backend services implement hone ke baad visible honge.
