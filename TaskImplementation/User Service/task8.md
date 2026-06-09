# 👤 User Service - Task 8: Add User Events

![Task](https://img.shields.io/badge/Task-User%20Service%20%238-blue)
![Priority](https://img.shields.io/badge/Priority-P2-lightgrey)
![Dependency](https://img.shields.io/badge/Dependency-Message%20Queue-yellow)
![Status](https://img.shields.io/badge/Status-Documented-success)
![Guide](https://img.shields.io/badge/Guide-Hinglish-orange)
![Scope](https://img.shields.io/badge/Scope-User%20Events-purple)
![Pattern](https://img.shields.io/badge/Pattern-Outbox%20%2B%20MQ-teal)

## 📌 Task Summary

| Field | Detail |
|---|---|
| Task Name | Add user events |
| Source | `docs/01-micro-tasks.md` → `User Service` → Task 8 |
| Priority | `P2` async workflows, analytics, recommendation, notification |
| Dependency | Message queue from Platform Foundation |
| Main Goal | `UserCreated`, `SellerApproved`, aur `AddressUpdated` events publish karna |
| Core Boundary | User Service apne DB write ke baad reliable event publish karega. Consumers apna kaam async karenge. |
| Output Type | Structured implementation guide |
| Not Included | Notification Service implementation, Recommendation logic, Analytics dashboards, Kafka/RabbitMQ infra setup, new public REST APIs |

> **Simple Hinglish goal:** Jab User Service me important profile changes hote hain, baaki services ko direct DB read karne ki zarurat nahi honi chahiye. User Service event publish karega, jaise `UserCreated`, `SellerApproved`, `AddressUpdated`. Notification, Analytics, Recommendation jaise consumers ye events consume karke apna kaam async karenge.

---

## ✅ Final Output Created

```text
TaskImplementation/
├── Auth Service/
├── Platform Foundation/
└── User Service/
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
- `TaskImplementation/User Service/` already present tha, isliye usko preserve kiya gaya.
- `task8.md` sirf **User Service - Task 8** ka implementation guide hai.
- Existing `task1.md` to `task7.md`, Auth Service guides, Platform Foundation guides, docs, database files, aur backend source files untouched rakhe gaye.
- Actual backend source code create nahi kiya gaya, kyunki current request ka required output folder structure aur complete `task8.md` content hai.

---

## 🧭 Implementation Approach

Is guide ko banate time project ke official docs aur existing User Service task guides ko source of truth maana gaya:

| Document/File | Kya use kiya gaya |
|---|---|
| `docs/01-micro-tasks.md` | User Service Task 8 ka exact scope: `UserCreated`, `SellerApproved`, `AddressUpdated` events |
| `docs/02-system-architecture.md` | Kafka/RabbitMQ async communication and idempotent consumer rule |
| `docs/03-folder-structure.md` | `internal/events` folder convention |
| `docs/04-microservice-design.md` | User Service responsibilities: profile, address, seller profile, KYC metadata, profile events |
| `docs/05-database-design.md` | User DB tables and service-owned database boundary |
| `docs/11-devops-external-services.md` | `user.events` topic/queue, event envelope, retry, DLQ, schema versioning |
| `docs/12-logging-monitoring-scalability.md` | Queue metrics, tracing, logs, and PII redaction rules |
| `TaskImplementation/Platform Foundation/task3.md` | Shared `events.Envelope` and `events.Publisher` pattern |
| `TaskImplementation/Platform Foundation/task4.md` | RabbitMQ default and Kafka alternate local stack decision |
| `TaskImplementation/User Service/task1.md` | User/Auth ownership boundary |
| `TaskImplementation/User Service/task2.md` | User schema and table names |
| `TaskImplementation/User Service/task3.md` | Repository layer and transaction pattern |
| `TaskImplementation/User Service/task4.md` | gRPC methods and internal service boundary |
| `TaskImplementation/User Service/task5.md` | REST Gateway routes for profile/address/seller flows |
| `TaskImplementation/User Service/task6.md` | Validation before DB write |
| `TaskImplementation/User Service/task7.md` | Audit actor fields used inside event metadata |
| `database/draw.sql` | Existing `users`, `user_addresses`, `seller_profiles`, `seller_kyc_documents` DDL |
| `api/master-api.json` | User profile, address, seller profile, and admin status request shapes |

---

## 🧱 Task Boundary

### ✅ Included in Task 8

- User Service event publishing design
- `UserCreated`, `SellerApproved`, `AddressUpdated` event contracts
- Common event envelope usage
- Reliable outbox pattern design
- `user_outbox_events` migration example
- Event payload examples with minimal PII
- Event recorder interface and usecase wiring
- Repository transaction examples
- RabbitMQ publisher example
- Kafka alternate publisher example
- Retry, DLQ, idempotency, and duplicate handling rules
- Metrics, logs, tracing, and operational checks
- Mermaid architecture and sequence diagrams
- Beginner-friendly step-by-step Hinglish guide
- External libraries/tools install and usage explanation

### 🚫 Not Included in Task 8

- Notification Service consumer implementation
- Recommendation Service personalization logic
- Analytics warehouse/dashboard implementation
- Kafka/RabbitMQ Docker Compose files or Kubernetes manifests
- New public REST APIs
- Auth credential/token changes
- Seller approval UI
- Full Superadmin audit-log implementation
- Cross-service database joins
- Events beyond `UserCreated`, `SellerApproved`, and `AddressUpdated`

> 🟢 **Rule:** Task 8 ka focus sirf User Service ke important writes ko async events me convert karna hai. Consumers future services me apna kaam karenge. User Service kisi consumer ke DB ko direct write nahi karega.

---

## 🗂️ Target Implementation Folder Structure

Future me actual User Service event implementation ka clean target structure ye hoga:

```text
backend/
└── services/
    └── user-service/
        ├── cmd/
        │   └── server/
        │       └── main.go
        ├── internal/
        │   ├── domain/
        │   │   ├── user.go
        │   │   ├── address.go
        │   │   ├── seller_profile.go
        │   │   └── user_events.go
        │   ├── events/
        │   │   ├── recorder.go
        │   │   ├── envelope.go
        │   │   ├── outbox_recorder.go
        │   │   ├── outbox_worker.go
        │   │   ├── rabbitmq_publisher.go
        │   │   └── kafka_publisher.go
        │   ├── repository/
        │   │   ├── mysql_user_repository.go
        │   │   ├── mysql_address_repository.go
        │   │   ├── mysql_seller_repository.go
        │   │   └── mysql_outbox_repository.go
        │   ├── usecase/
        │   │   ├── create_user.go
        │   │   ├── manage_address.go
        │   │   ├── seller_status.go
        │   │   └── event_helpers.go
        │   ├── config/
        │   │   └── config.go
        │   └── transport/
        │       └── grpc/
        │           └── server.go
        └── migrations/
            ├── 008_create_user_outbox_events.up.sql
            └── 008_create_user_outbox_events.down.sql
```

### Folder responsibility

| Path | Responsibility |
|---|---|
| `internal/domain/user_events.go` | Typed payload structs for `UserCreated`, `SellerApproved`, `AddressUpdated` |
| `internal/events/recorder.go` | Usecase-facing event interface |
| `internal/events/envelope.go` | Event envelope builder and version constants |
| `internal/events/outbox_recorder.go` | Transaction ke andar outbox event row create karna |
| `internal/events/outbox_worker.go` | Pending outbox events ko MQ pe publish karna |
| `internal/events/rabbitmq_publisher.go` | RabbitMQ implementation for local/default mode |
| `internal/events/kafka_publisher.go` | Kafka alternate implementation |
| `internal/repository/mysql_outbox_repository.go` | Outbox insert, lock pending, mark published/failed queries |
| `internal/usecase/event_helpers.go` | Event payload build helpers |
| `migrations/008_create_user_outbox_events.*.sql` | Outbox table create/drop |

> 🟡 **Important:** Ye target source structure hai. Is current task output me sirf `TaskImplementation/User Service/task8.md` create kiya gaya.

---

## 🧠 Event Concept

Event ka matlab hai: "User Service me kuch important ho chuka hai."

Event command nahi hota. Event consumer ko ye nahi bolta ki kya karna hai. Event sirf fact batata hai.

```text
Command: Approve this seller.
Event: SellerApproved happened.
```

### Why events?

| Problem | Event-based solution |
|---|---|
| Notification ko welcome email bhejna hai | `UserCreated` consume karo |
| Analytics ko signup count chahiye | `UserCreated` consume karo |
| Recommendation ko seller active signal chahiye | `SellerApproved` consume karo |
| Address changes se shipping/profile projections update karni hain | `AddressUpdated` consume karo |
| User DB ko dusri services direct read kar rahi hain | Event projection use karo |

### Event golden rules

| Rule | Reason |
|---|---|
| Producer owns fact | `UserCreated` ka source User Service hai |
| Consumer idempotent hona chahiye | MQ at-least-once duplicate delivery kar sakta hai |
| Event immutable hota hai | Publish ke baad event edit nahi hota |
| Payload versioned hota hai | Future schema changes safe hote hain |
| Sensitive data avoid karo | Events many consumers tak ja sakte hain |
| DB write and event enqueue reliable honi chahiye | Outbox pattern event loss avoid karta hai |

---

## 🏗️ High-Level Architecture

```mermaid
flowchart LR
    Auth[Auth Service] -->|CreateUser gRPC| UserSvc[User Service]
    Gateway[API Gateway] -->|Address/Profile gRPC| UserSvc
    Superadmin[Superadmin Service] -->|Approve Seller| UserSvc

    UserSvc --> Usecase[User Usecase Layer]
    Usecase --> Repo[User Repository]
    Repo --> DB[(user_db MySQL)]
    Repo --> Outbox[(user_outbox_events)]

    Worker[User Outbox Worker] --> Outbox
    Worker --> MQ[user.events Topic/Queue]

    MQ --> Notification[Notification Service]
    MQ --> Analytics[Analytics / Session Pipeline]
    MQ --> Recommendation[Recommendation Service]

    classDef source fill:#e8f4ff,stroke:#1976d2,color:#111;
    classDef service fill:#d4edda,stroke:#28a745,color:#111;
    classDef db fill:#fff3cd,stroke:#d39e00,color:#111;
    classDef mq fill:#f3e5f5,stroke:#8e24aa,color:#111;

    class Auth,Gateway,Superadmin source;
    class UserSvc,Usecase,Repo,Worker,Notification,Analytics,Recommendation service;
    class DB,Outbox db;
    class MQ mq;
```

### Hinglish explanation

User Service pehle apna main DB write karega. Same transaction me outbox row insert karega. Background worker pending outbox rows read karke `user.events` topic/queue pe publish karega. Agar MQ down hai, DB me pending row safe rahegi and worker retry karega.

---

## 🔁 Outbox Pattern Flow

```mermaid
sequenceDiagram
    autonumber
    participant Caller as Auth/Gateway/Superadmin
    participant User as User Service Usecase
    participant DB as user_db Transaction
    participant Outbox as user_outbox_events
    participant Worker as Outbox Worker
    participant MQ as user.events
    participant Consumer as Notification/Analytics/Recommendation

    Caller->>User: Create/update request
    User->>DB: BEGIN
    User->>DB: Insert/update user-owned table
    User->>Outbox: Insert pending event row
    User->>DB: COMMIT
    User-->>Caller: Success response
    Worker->>Outbox: Lock pending events
    Worker->>MQ: Publish event envelope
    MQ-->>Consumer: Deliver event
    Consumer->>Consumer: Process idempotently by event_id
    Worker->>Outbox: Mark published
```

### Why outbox?

| Failure | Without outbox | With outbox |
|---|---|---|
| DB write success, MQ publish fail | Event lost ho sakta hai | Pending outbox row retry hogi |
| App crash after DB commit | Consumer ko pata nahi chalega | Worker restart ke baad publish karega |
| MQ temporary down | Request fail ya event lost | Business write safe, event delayed |
| Duplicate publish | Consumer confuse ho sakta hai | `event_id` idempotency se ignore |

---

## 📣 Event Topic/Queue Naming

| Item | Name |
|---|---|
| Main topic/queue | `user.events` |
| Dead-letter queue | `user.events.dlq` |
| Producer service | `user-service` |
| Consumer group examples | `notification-service.user-events`, `analytics.user-events`, `recommendation.user-events` |
| Partition/routing key | `aggregate_id`, usually `user_id`, `seller_id`, or `address_id` |

> 🟠 **Note:** Platform Foundation me local default RabbitMQ hai, Kafka alternate documented hai. Is guide me topic name `user.events` generic rakha gaya hai, taaki RabbitMQ exchange ya Kafka topic dono me use ho sake.

---

## 📦 Common Event Envelope

Platform shared events package ke style ko follow karo:

```go
package events

import "time"

type Envelope[T any] struct {
    EventID     string    `json:"event_id"`
    EventType   string    `json:"event_type"`
    Version     int       `json:"version"`
    Source      string    `json:"source"`
    RequestID   string    `json:"request_id,omitempty"`
    TraceID     string    `json:"trace_id,omitempty"`
    AggregateID string    `json:"aggregate_id"`
    OccurredAt  time.Time `json:"occurred_at"`
    Payload     T         `json:"payload"`
}
```

### Example envelope

```json
{
  "event_id": "evt_01HYUSERCREATED001",
  "event_type": "UserCreated",
  "version": 1,
  "source": "user-service",
  "request_id": "req_123",
  "trace_id": "trace_456",
  "aggregate_id": "usr_123",
  "occurred_at": "2026-05-21T10:00:00Z",
  "payload": {
    "user_id": "usr_123",
    "auth_account_id": "auth_123",
    "status": "active",
    "created_by": "service:auth-service"
  }
}
```

### Envelope fields explanation

| Field | Meaning |
|---|---|
| `event_id` | Unique ID for dedupe and debugging |
| `event_type` | Event name, for example `UserCreated` |
| `version` | Payload schema version |
| `source` | Producer service name |
| `request_id` | Original request correlation id |
| `trace_id` | Distributed trace id |
| `aggregate_id` | Main entity id, like `user_id`, `seller_id`, or `address_id` |
| `occurred_at` | Business event time in UTC |
| `payload` | Event-specific data |

---

## 🧾 Event Contracts

### 1. `UserCreated`

`UserCreated` tab publish hoga jab Auth signup ke baad User Service me `users` row successfully create ho jaye.

| Field | Value |
|---|---|
| Event type | `UserCreated` |
| Topic/queue | `user.events` |
| Aggregate type | `user` |
| Aggregate id | `user_id` |
| Trigger | `CreateUser` usecase success |
| Main consumers | Notification, Analytics, Recommendation |
| Version | `1` |

#### Payload

```go
type UserCreatedPayload struct {
    UserID        string    `json:"user_id"`
    AuthAccountID string    `json:"auth_account_id"`
    Status        string    `json:"status"`
    EmailHash     string    `json:"email_hash,omitempty"`
    PhoneHash     string    `json:"phone_hash,omitempty"`
    CreatedBy     string    `json:"created_by"`
    CreatedAt     time.Time `json:"created_at"`
}
```

#### Example JSON

```json
{
  "event_id": "evt_user_created_123",
  "event_type": "UserCreated",
  "version": 1,
  "source": "user-service",
  "request_id": "req_signup_123",
  "trace_id": "trace_signup_123",
  "aggregate_id": "usr_123",
  "occurred_at": "2026-05-21T10:00:00Z",
  "payload": {
    "user_id": "usr_123",
    "auth_account_id": "auth_123",
    "status": "active",
    "email_hash": "sha256:...",
    "created_by": "service:auth-service",
    "created_at": "2026-05-21T10:00:00Z"
  }
}
```

#### Why no raw password/token?

Password, OTP, JWT, refresh token, auth headers, raw device fingerprint, aur secrets event me kabhi nahi jayenge. `UserCreated` profile event hai, auth secret event nahi.

---

### 2. `SellerApproved`

`SellerApproved` tab publish hoga jab seller profile ka status `pending_review` ya `draft` se `active` ho jaye after admin approval.

| Field | Value |
|---|---|
| Event type | `SellerApproved` |
| Topic/queue | `user.events` |
| Aggregate type | `seller` |
| Aggregate id | `seller_id` |
| Trigger | `UpdateSellerStatus` / seller approval usecase success |
| Main consumers | Notification, Product/CMS projections, Analytics, Recommendation |
| Version | `1` |

#### Payload

```go
type SellerApprovedPayload struct {
    SellerID         string    `json:"seller_id"`
    UserID           string    `json:"user_id"`
    StoreName        string    `json:"store_name,omitempty"`
    PreviousStatus   string    `json:"previous_status"`
    CurrentStatus    string    `json:"current_status"`
    ApprovedBy       string    `json:"approved_by"`
    ApprovedAt       time.Time `json:"approved_at"`
    StatusReason     string    `json:"status_reason,omitempty"`
}
```

#### Example JSON

```json
{
  "event_id": "evt_seller_approved_123",
  "event_type": "SellerApproved",
  "version": 1,
  "source": "user-service",
  "request_id": "req_admin_123",
  "trace_id": "trace_admin_123",
  "aggregate_id": "sel_123",
  "occurred_at": "2026-05-21T11:00:00Z",
  "payload": {
    "seller_id": "sel_123",
    "user_id": "usr_123",
    "store_name": "Parag Store",
    "previous_status": "pending_review",
    "current_status": "active",
    "approved_by": "admin_456",
    "approved_at": "2026-05-21T11:00:00Z",
    "status_reason": "KYC verified"
  }
}
```

#### Scope note

Task 8 me sirf `SellerApproved` required hai. `SellerRejected`, `SellerSuspended`, `SellerKYCSubmitted` useful future events ho sakte hain, but is task me implement nahi karne.

---

### 3. `AddressUpdated`

`AddressUpdated` tab publish hoga jab user address state change hoti hai: address create, update, delete, ya default-address switch.

| Field | Value |
|---|---|
| Event type | `AddressUpdated` |
| Topic/queue | `user.events` |
| Aggregate type | `address` |
| Aggregate id | `address_id` |
| Trigger | `CreateAddress`, `UpdateAddress`, `DeleteAddress`, default change usecase success |
| Main consumers | Analytics, Notification, Order/shipping read projections if needed |
| Version | `1` |

#### Payload

```go
type AddressUpdatedPayload struct {
    AddressID     string    `json:"address_id"`
    UserID        string    `json:"user_id"`
    ChangeType    string    `json:"change_type"` // created, updated, deleted, default_changed
    City          string    `json:"city,omitempty"`
    State         string    `json:"state,omitempty"`
    Country       string    `json:"country,omitempty"`
    IsDefault     bool      `json:"is_default"`
    UpdatedBy     string    `json:"updated_by"`
    UpdatedAt     time.Time `json:"updated_at"`
}
```

#### Example JSON

```json
{
  "event_id": "evt_address_updated_123",
  "event_type": "AddressUpdated",
  "version": 1,
  "source": "user-service",
  "request_id": "req_address_123",
  "trace_id": "trace_address_123",
  "aggregate_id": "addr_123",
  "occurred_at": "2026-05-21T12:00:00Z",
  "payload": {
    "address_id": "addr_123",
    "user_id": "usr_123",
    "change_type": "updated",
    "city": "Delhi",
    "state": "Delhi",
    "country": "IN",
    "is_default": true,
    "updated_by": "usr_123",
    "updated_at": "2026-05-21T12:00:00Z"
  }
}
```

#### PII rule for address events

Full address lines, exact house number, raw phone, and delivery instructions event me mat bhejo. Event consumers ko exact shipping address chahiye to woh authorized gRPC call se User Service se fetch karein.

---

## 🔐 Payload Privacy Rules

| Data | Event me allowed? | Reason |
|---|---|---|
| `user_id` | ✅ Yes | Stable internal reference |
| `seller_id` | ✅ Yes | Stable seller reference |
| `address_id` | ✅ Yes | Stable address reference |
| `status` | ✅ Yes | Workflow signal |
| `email_hash` | ✅ Yes | Analytics dedupe without raw email |
| raw `email` | ⚠️ Avoid by default | PII, broad consumers tak leak risk |
| raw `phone` | ⚠️ Avoid by default | PII |
| full address line | ❌ No | Sensitive location data |
| password / OTP / JWT / refresh token | ❌ Never | Security secret |
| `created_by` / `updated_by` | ✅ Yes | Audit trace |
| `request_id` / `trace_id` | ✅ Yes | Debugging and tracing |

> 🟠 **Notification note:** Agar Notification Service ko actual email/phone chahiye, better pattern hai authorized lookup via User Service ya separate `notification.commands` flow. Broad `user.events` topic ko raw PII bus nahi banana chahiye.

---

## 🛠️ External Libraries / Tools

Task 8 me actual implementation karte time ye tools useful honge:

| Tool/Library | What it is | Why used | Install | Use |
|---|---|---|---|---|
| `github.com/google/uuid` | Go UUID generator | `event_id` generate karne ke liye | `go get github.com/google/uuid` | `uuid.NewString()` |
| `github.com/rabbitmq/amqp091-go` | RabbitMQ Go client | Local/default MQ publisher | `go get github.com/rabbitmq/amqp091-go` | `amqp.Dial`, `Channel.PublishWithContext` |
| `github.com/segmentio/kafka-go` | Kafka Go client | Kafka alternate event publisher | `go get github.com/segmentio/kafka-go` | `kafka.Writer.WriteMessages` |
| `github.com/go-sql-driver/mysql` | MySQL driver | Outbox repository DB access | `go get github.com/go-sql-driver/mysql` | `sql.Open("mysql", dsn)` |
| Docker Compose | Local infra runner | MySQL + RabbitMQ/Kafka run karne ke liye | Docker Desktop / Docker Engine | `docker compose up mysql rabbitmq` |
| RabbitMQ Management UI | Browser UI | Queue, exchanges, DLQ inspect karne ke liye | Comes with `rabbitmq:management` image | `http://localhost:15672` |
| `jq` | JSON CLI viewer | Event payload debug pretty print | `sudo apt install jq` or package manager | `cat event.json | jq` |

### Install commands

```bash
cd backend/services/user-service
go get github.com/google/uuid
go get github.com/rabbitmq/amqp091-go
go get github.com/segmentio/kafka-go
go get github.com/go-sql-driver/mysql
```

### Local RabbitMQ start example

```bash
docker compose up -d mysql rabbitmq
```

### RabbitMQ UI

```text
URL:      http://localhost:15672
Username: ecommerce
Password: ecommerce_password
```

> 🟡 **Note:** Platform Foundation Task 4 already local stack design karta hai. User Service Task 8 me infra files create nahi karni; sirf service event publishing design implement hoga.

---

## ⚙️ Configuration

```env
USER_EVENTS_ENABLED=true
USER_EVENTS_MODE=outbox
USER_EVENTS_TOPIC=user.events
USER_EVENTS_DLQ=user.events.dlq
USER_EVENTS_PROVIDER=rabbitmq

RABBITMQ_URL=amqp://ecommerce:ecommerce_password@rabbitmq:5672/
KAFKA_BROKERS=kafka:9092

OUTBOX_WORKER_ENABLED=true
OUTBOX_WORKER_BATCH_SIZE=100
OUTBOX_WORKER_POLL_INTERVAL=2s
OUTBOX_MAX_ATTEMPTS=10
```

### Config explanation

| Config | Meaning |
|---|---|
| `USER_EVENTS_ENABLED` | Local debugging me events disable karne ka switch |
| `USER_EVENTS_MODE` | `outbox` recommended, `disabled` local fallback |
| `USER_EVENTS_TOPIC` | Main topic/queue name |
| `USER_EVENTS_PROVIDER` | `rabbitmq` default ya `kafka` alternate |
| `OUTBOX_WORKER_BATCH_SIZE` | Ek poll me kitne pending events process honge |
| `OUTBOX_MAX_ATTEMPTS` | Iske baad event DLQ/manual review me jayega |

---

## 🧩 Step 1: Event Trigger Points Final Karo

Sabse pehle decide karo ki event exactly kab publish hoga.

| Event | Trigger point | Publish timing |
|---|---|---|
| `UserCreated` | `users` row insert success | Same DB transaction me outbox insert |
| `SellerApproved` | `seller_profiles.status` changes to `active` | Same DB transaction me outbox insert |
| `AddressUpdated` | Address create/update/delete/default change | Same DB transaction me outbox insert |

### Important rule

Event tabhi enqueue hoga jab domain write successful ho. Agar validation fail, permission fail, DB error, ya transaction rollback ho gaya, event nahi publish hoga.

```text
Validation fail -> no DB write -> no event
DB rollback -> no event
DB commit -> outbox row exists -> event eventually publish
```

---

## 🧩 Step 2: Outbox Migration Add Karo

User Service MySQL DB me ek outbox table chahiye.

### Up migration

```sql
CREATE TABLE user_outbox_events (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  event_id VARCHAR(64) NOT NULL,
  event_type VARCHAR(80) NOT NULL,
  event_version INT NOT NULL,
  topic VARCHAR(120) NOT NULL,
  aggregate_type VARCHAR(80) NOT NULL,
  aggregate_id VARCHAR(64) NOT NULL,
  payload JSON NOT NULL,
  request_id VARCHAR(128) NULL,
  trace_id VARCHAR(128) NULL,
  status ENUM('pending', 'published', 'failed', 'dead') NOT NULL DEFAULT 'pending',
  attempts INT NOT NULL DEFAULT 0,
  next_attempt_at TIMESTAMP NULL,
  locked_until TIMESTAMP NULL,
  last_error VARCHAR(1024) NULL,
  occurred_at TIMESTAMP NOT NULL,
  published_at TIMESTAMP NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_user_outbox_event_id (event_id),
  KEY idx_user_outbox_pending (status, next_attempt_at, created_at),
  KEY idx_user_outbox_lock (locked_until),
  KEY idx_user_outbox_aggregate (aggregate_type, aggregate_id),
  KEY idx_user_outbox_type_time (event_type, occurred_at)
) ENGINE=InnoDB;
```

### Down migration

```sql
DROP TABLE IF EXISTS user_outbox_events;
```

### Column explanation

| Column | Purpose |
|---|---|
| `event_id` | Dedupe key |
| `event_type` | `UserCreated`, `SellerApproved`, `AddressUpdated` |
| `event_version` | Payload schema version |
| `topic` | Usually `user.events` |
| `aggregate_type` | `user`, `seller`, `address` |
| `aggregate_id` | Main entity id |
| `payload` | JSON event envelope or payload |
| `status` | Publish lifecycle |
| `attempts` | Retry count |
| `next_attempt_at` | Backoff scheduling |
| `locked_until` | Multi-worker duplicate lock avoid karna |
| `last_error` | Debugging |
| `occurred_at` | Business event time |
| `published_at` | MQ publish success time |

---

## 🧩 Step 3: Domain Event Structs Banao

`internal/domain/user_events.go` me typed payloads define karo.

```go
package domain

import "time"

const (
    EventUserCreated    = "UserCreated"
    EventSellerApproved = "SellerApproved"
    EventAddressUpdated = "AddressUpdated"
)

type UserCreatedPayload struct {
    UserID        string    `json:"user_id"`
    AuthAccountID string    `json:"auth_account_id"`
    Status        string    `json:"status"`
    EmailHash     string    `json:"email_hash,omitempty"`
    PhoneHash     string    `json:"phone_hash,omitempty"`
    CreatedBy     string    `json:"created_by"`
    CreatedAt     time.Time `json:"created_at"`
}

type SellerApprovedPayload struct {
    SellerID       string    `json:"seller_id"`
    UserID         string    `json:"user_id"`
    StoreName      string    `json:"store_name,omitempty"`
    PreviousStatus string    `json:"previous_status"`
    CurrentStatus  string    `json:"current_status"`
    ApprovedBy     string    `json:"approved_by"`
    ApprovedAt     time.Time `json:"approved_at"`
    StatusReason   string    `json:"status_reason,omitempty"`
}

type AddressUpdatedPayload struct {
    AddressID string    `json:"address_id"`
    UserID    string    `json:"user_id"`
    ChangeType string   `json:"change_type"`
    City      string    `json:"city,omitempty"`
    State     string    `json:"state,omitempty"`
    Country   string    `json:"country,omitempty"`
    IsDefault bool      `json:"is_default"`
    UpdatedBy string    `json:"updated_by"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### Why typed structs?

Typed structs se field typo compile time pe pakad me aa jata hai. Random `map[string]any` me galti runtime pe dikhegi, jo event contracts ke liye risky hai.

---

## 🧩 Step 4: Event Recorder Interface Banao

Usecase layer ko MQ details nahi pata honi chahiye. Usecase sirf event recorder ko call karega.

```go
package events

import (
    "context"
    "database/sql"

    "ecommerce/backend/services/user-service/internal/domain"
)

type Recorder interface {
    RecordUserCreated(ctx context.Context, tx *sql.Tx, payload domain.UserCreatedPayload) error
    RecordSellerApproved(ctx context.Context, tx *sql.Tx, payload domain.SellerApprovedPayload) error
    RecordAddressUpdated(ctx context.Context, tx *sql.Tx, payload domain.AddressUpdatedPayload) error
}
```

### Hinglish explanation

Usecase ko bas itna pata hai: "mujhe event record karna hai". Event direct RabbitMQ me jayega, Kafka me jayega, ya outbox table me save hoga - ye implementation detail `internal/events` package handle karega.

---

## 🧩 Step 5: Envelope Builder Banao

```go
package events

import (
    "time"

    "github.com/google/uuid"
)

type Envelope[T any] struct {
    EventID     string    `json:"event_id"`
    EventType   string    `json:"event_type"`
    Version     int       `json:"version"`
    Source      string    `json:"source"`
    RequestID   string    `json:"request_id,omitempty"`
    TraceID     string    `json:"trace_id,omitempty"`
    AggregateID string    `json:"aggregate_id"`
    OccurredAt  time.Time `json:"occurred_at"`
    Payload     T         `json:"payload"`
}

func NewEnvelope[T any](
    eventType string,
    aggregateID string,
    requestID string,
    traceID string,
    occurredAt time.Time,
    payload T,
) Envelope[T] {
    return Envelope[T]{
        EventID:     "evt_" + uuid.NewString(),
        EventType:   eventType,
        Version:     1,
        Source:      "user-service",
        RequestID:   requestID,
        TraceID:     traceID,
        AggregateID: aggregateID,
        OccurredAt:  occurredAt.UTC(),
        Payload:     payload,
    }
}
```

### Why `occurred_at` UTC?

Distributed systems me timezone confusion common bug hai. Events always UTC me store/publish karo.

---

## 🧩 Step 6: Outbox Recorder Implement Karo

```go
package events

import (
    "context"
    "database/sql"
    "encoding/json"
    "time"

    "ecommerce/backend/services/user-service/internal/domain"
)

type OutboxRepository interface {
    Insert(ctx context.Context, tx *sql.Tx, row OutboxEvent) error
}

type OutboxRecorder struct {
    repo  OutboxRepository
    topic string
    clock func() time.Time
}

func NewOutboxRecorder(repo OutboxRepository, topic string) *OutboxRecorder {
    return &OutboxRecorder{
        repo:  repo,
        topic: topic,
        clock: func() time.Time { return time.Now().UTC() },
    }
}

func (r *OutboxRecorder) RecordUserCreated(ctx context.Context, tx *sql.Tx, payload domain.UserCreatedPayload) error {
    env := NewEnvelope(domain.EventUserCreated, payload.UserID, requestIDFromContext(ctx), traceIDFromContext(ctx), r.clock(), payload)
    return r.insertEnvelope(ctx, tx, "user", payload.UserID, env)
}

func (r *OutboxRecorder) RecordSellerApproved(ctx context.Context, tx *sql.Tx, payload domain.SellerApprovedPayload) error {
    env := NewEnvelope(domain.EventSellerApproved, payload.SellerID, requestIDFromContext(ctx), traceIDFromContext(ctx), r.clock(), payload)
    return r.insertEnvelope(ctx, tx, "seller", payload.SellerID, env)
}

func (r *OutboxRecorder) RecordAddressUpdated(ctx context.Context, tx *sql.Tx, payload domain.AddressUpdatedPayload) error {
    env := NewEnvelope(domain.EventAddressUpdated, payload.AddressID, requestIDFromContext(ctx), traceIDFromContext(ctx), r.clock(), payload)
    return r.insertEnvelope(ctx, tx, "address", payload.AddressID, env)
}

func (r *OutboxRecorder) insertEnvelope(ctx context.Context, tx *sql.Tx, aggregateType, aggregateID string, env any) error {
    raw, err := json.Marshal(env)
    if err != nil {
        return err
    }

    base := envelopeBase(env)

    return r.repo.Insert(ctx, tx, OutboxEvent{
        EventID:       base.EventID,
        EventType:     base.EventType,
        EventVersion:  base.Version,
        Topic:         r.topic,
        AggregateType: aggregateType,
        AggregateID:   aggregateID,
        Payload:       raw,
        RequestID:     base.RequestID,
        TraceID:       base.TraceID,
        OccurredAt:    base.OccurredAt,
    })
}
```

### Implementation note

`requestIDFromContext`, `traceIDFromContext`, and `envelopeBase` helper examples simplified hain. Actual code me shared middleware/tracing package se values read karna better rahega.

---

## 🧩 Step 7: Outbox Repository Queries

```go
type OutboxEvent struct {
    EventID       string
    EventType     string
    EventVersion  int
    Topic         string
    AggregateType string
    AggregateID   string
    Payload       []byte
    RequestID     string
    TraceID       string
    OccurredAt    time.Time
}
```

### Insert query

```go
func (r *MySQLOutboxRepository) Insert(ctx context.Context, tx *sql.Tx, evt OutboxEvent) error {
    const query = `
        INSERT INTO user_outbox_events (
            event_id,
            event_type,
            event_version,
            topic,
            aggregate_type,
            aggregate_id,
            payload,
            request_id,
            trace_id,
            occurred_at,
            status,
            next_attempt_at
        ) VALUES (?, ?, ?, ?, ?, ?, CAST(? AS JSON), ?, ?, ?, 'pending', UTC_TIMESTAMP())
    `

    _, err := tx.ExecContext(
        ctx,
        query,
        evt.EventID,
        evt.EventType,
        evt.EventVersion,
        evt.Topic,
        evt.AggregateType,
        evt.AggregateID,
        string(evt.Payload),
        nullString(evt.RequestID),
        nullString(evt.TraceID),
        evt.OccurredAt,
    )
    return err
}
```

### Lock pending query

```sql
SELECT id, event_id, topic, payload, attempts
FROM user_outbox_events
WHERE status IN ('pending', 'failed')
  AND (next_attempt_at IS NULL OR next_attempt_at <= UTC_TIMESTAMP())
  AND (locked_until IS NULL OR locked_until < UTC_TIMESTAMP())
ORDER BY created_at ASC
LIMIT ?
FOR UPDATE SKIP LOCKED;
```

### Why `SKIP LOCKED`?

Agar multiple worker replicas chal rahe hain, same event do workers pick na karein. `FOR UPDATE SKIP LOCKED` ek worker ko row lock dega, baaki workers locked row skip karenge.

---

## 🧩 Step 8: `CreateUser` Usecase Me `UserCreated` Wire Karo

```go
func (uc *CreateUserUsecase) Execute(ctx context.Context, input CreateUserInput) (*domain.User, error) {
    if err := uc.validator.ValidateCreateUser(input); err != nil {
        return nil, err
    }

    tx, err := uc.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    defer rollbackUnlessCommitted(ctx, tx)

    user, err := uc.users.Create(ctx, tx, input)
    if err != nil {
        return nil, err
    }

    payload := domain.UserCreatedPayload{
        UserID:        user.UserID,
        AuthAccountID: user.AuthAccountID,
        Status:        user.Status,
        EmailHash:     hashForEvent(user.Email),
        PhoneHash:     hashForEvent(user.Phone),
        CreatedBy:     actorFromContext(ctx),
        CreatedAt:     user.CreatedAt,
    }

    if err := uc.events.RecordUserCreated(ctx, tx, payload); err != nil {
        return nil, err
    }

    if err := tx.Commit(); err != nil {
        return nil, err
    }

    return user, nil
}
```

### Build explanation

- Validation pehle hoti hai, taaki bad data DB/outbox tak na jaye.
- User row and outbox row same transaction me save hote hain.
- Raw email/phone event me nahi jaate; hash useful analytics dedupe ke liye enough hota hai.
- Commit ke baad response success return hota hai.
- MQ publish response path me nahi hota; worker async publish karta hai.

---

## 🧩 Step 9: Address Flows Me `AddressUpdated` Wire Karo

Address create/update/delete/default-change sab address state change hain. Task wording `AddressUpdated` bolta hai, isliye single event type use kar sakte hain with `change_type`.

```go
func (uc *AddressUsecase) UpdateAddress(ctx context.Context, input UpdateAddressInput) (*domain.Address, error) {
    if err := uc.validator.ValidateAddress(input.Address); err != nil {
        return nil, err
    }

    tx, err := uc.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    defer rollbackUnlessCommitted(ctx, tx)

    address, err := uc.addresses.Update(ctx, tx, input.UserID, input.AddressID, input.Address)
    if err != nil {
        return nil, err
    }

    payload := domain.AddressUpdatedPayload{
        AddressID:  address.AddressID,
        UserID:     address.UserID,
        ChangeType: "updated",
        City:       address.City,
        State:      address.State,
        Country:    address.Country,
        IsDefault:  address.IsDefault,
        UpdatedBy:  actorFromContext(ctx),
        UpdatedAt:  address.UpdatedAt,
    }

    if err := uc.events.RecordAddressUpdated(ctx, tx, payload); err != nil {
        return nil, err
    }

    if err := tx.Commit(); err != nil {
        return nil, err
    }

    return address, nil
}
```

### Change type values

| Change type | Kab use hoga |
|---|---|
| `created` | New address add hua |
| `updated` | Existing address fields change hue |
| `deleted` | Address soft delete hua |
| `default_changed` | Default address switch hua |

### Default address special case

Agar ek update me old default unset and new default set hota hai, ek hi meaningful event publish karo for the new current state. Consumer ko low-level DB row changes nahi, user-facing final state chahiye.

---

## 🧩 Step 10: Seller Approval Me `SellerApproved` Wire Karo

```go
func (uc *SellerStatusUsecase) ApproveSeller(ctx context.Context, input ApproveSellerInput) (*domain.SellerProfile, error) {
    actor := actorFromContext(ctx)
    if !actorIsAdmin(actor) {
        return nil, domain.ErrForbidden
    }

    tx, err := uc.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    defer rollbackUnlessCommitted(ctx, tx)

    before, err := uc.sellers.GetForUpdate(ctx, tx, input.SellerID)
    if err != nil {
        return nil, err
    }

    seller, err := uc.sellers.UpdateStatus(ctx, tx, input.SellerID, "active", actor, input.Reason)
    if err != nil {
        return nil, err
    }

    payload := domain.SellerApprovedPayload{
        SellerID:       seller.SellerID,
        UserID:         seller.UserID,
        StoreName:      seller.StoreName,
        PreviousStatus: before.Status,
        CurrentStatus:  seller.Status,
        ApprovedBy:     actor,
        ApprovedAt:     seller.ApprovedAt,
        StatusReason:   input.Reason,
    }

    if err := uc.events.RecordSellerApproved(ctx, tx, payload); err != nil {
        return nil, err
    }

    if err := tx.Commit(); err != nil {
        return nil, err
    }

    return seller, nil
}
```

### Important safeguards

| Safeguard | Why |
|---|---|
| Admin actor check | Seller self-approve na kar sake |
| `GetForUpdate` | Concurrent approval/suspension conflict avoid |
| Previous/current status in payload | Consumer ko transition context mile |
| Same transaction outbox | Status active hua to event guaranteed pending hoga |

---

## 🧩 Step 11: Outbox Worker Banao

Worker background me pending events publish karega.

```go
type OutboxWorker struct {
    repo      OutboxRepository
    publisher Publisher
    batchSize int
    maxAttempts int
}

func (w *OutboxWorker) RunOnce(ctx context.Context) error {
    rows, err := w.repo.LockPending(ctx, w.batchSize)
    if err != nil {
        return err
    }

    for _, row := range rows {
        err := w.publisher.Publish(ctx, row.Topic, row.Payload)
        if err != nil {
            nextAttempt := nextBackoff(row.Attempts)
            _ = w.repo.MarkFailed(ctx, row.EventID, nextAttempt, err.Error())
            continue
        }

        _ = w.repo.MarkPublished(ctx, row.EventID)
    }

    return nil
}
```

### Retry policy

| Attempt | Delay idea |
|---:|---|
| 1 | 5 seconds |
| 2 | 30 seconds |
| 3 | 2 minutes |
| 4 | 10 minutes |
| 5 | 30 minutes |
| 6+ | Exponential backoff with cap |
| After max attempts | Mark `dead` and send to DLQ/manual review |

### Worker loop

```go
func (w *OutboxWorker) Run(ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            _ = w.RunOnce(ctx)
        }
    }
}
```

---

## 🧩 Step 12: RabbitMQ Publisher Example

RabbitMQ local/default queue ke liye simple and beginner-friendly option hai.

```go
type RabbitMQPublisher struct {
    channel *amqp.Channel
}

func (p *RabbitMQPublisher) Publish(ctx context.Context, topic string, payload []byte) error {
    return p.channel.PublishWithContext(
        ctx,
        topic, // exchange
        "",    // routing key, fanout exchange example
        false,
        false,
        amqp.Publishing{
            ContentType:  "application/json",
            DeliveryMode: amqp.Persistent,
            Body:         payload,
        },
    )
}
```

### RabbitMQ setup idea

| Item | Setting |
|---|---|
| Exchange | `user.events` |
| Exchange type | `fanout` for broad events or `topic` for routing |
| Queue | `notification.user-events`, `analytics.user-events`, `recommendation.user-events` |
| DLQ | `user.events.dlq` |
| Message persistence | Enabled |

### RabbitMQ why used?

- Local development me simple hai.
- Management UI se messages inspect kar sakte ho.
- Notification-style async jobs ke liye good MVP fit hai.
- Durable queue + persistent messages basic reliability dete hain.

---

## 🧩 Step 13: Kafka Publisher Alternate

Kafka high-volume stream and replay ke liye better fit hai. Agar team analytics/recommendation event replay heavily use kare, Kafka choose kar sakti hai.

```go
type KafkaPublisher struct {
    writer *kafka.Writer
}

func (p *KafkaPublisher) Publish(ctx context.Context, topic string, payload []byte) error {
    return p.writer.WriteMessages(ctx, kafka.Message{
        Topic: topic,
        Key:   []byte(extractAggregateID(payload)),
        Value: payload,
        Headers: []kafka.Header{
            {Key: "content-type", Value: []byte("application/json")},
        },
    })
}
```

### Kafka key choice

| Event | Good Kafka key |
|---|---|
| `UserCreated` | `user_id` |
| `SellerApproved` | `seller_id` |
| `AddressUpdated` | `user_id` or `address_id` |

Key ka benefit ye hai ki same aggregate ke events same partition me jaane ke chances badhte hain, ordering relatively better hoti hai.

---

## 🧩 Step 14: Consumer Idempotency Contract Document Karo

User Service producer hai, but consumer rules clearly document karna zaruri hai.

```text
if event_id already processed:
    ACK and do nothing
else:
    process event
    mark event_id processed
    ACK
```

### Consumer-side table example

```sql
CREATE TABLE processed_events (
  event_id VARCHAR(64) NOT NULL,
  consumer_name VARCHAR(120) NOT NULL,
  processed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (event_id, consumer_name)
);
```

### Consumer behavior

| Situation | Expected behavior |
|---|---|
| Duplicate event | Ignore and ACK |
| Unknown newer version | Send to DLQ or ignore safely with alert |
| Missing optional field | Use default/fallback |
| Invalid JSON | Send to DLQ |
| Consumer DB down | NACK/retry |

> 🟢 **Rule:** Producer at-least-once publish karega. Consumer exactly-once assume nahi karega.

---

## 🧩 Step 15: Observability Add Karo

Events invisible async work hote hain, isliye logs, metrics, and traces important hain.

### Logs

```json
{
  "level": "info",
  "service": "user-service",
  "message": "user event enqueued",
  "event_type": "UserCreated",
  "event_id": "evt_user_created_123",
  "aggregate_id": "usr_123",
  "request_id": "req_123",
  "trace_id": "trace_456"
}
```

### Do not log

- raw email
- raw phone
- full address
- JWT
- refresh token
- OTP
- secrets

### Metrics

| Metric | Type | Labels | Meaning |
|---|---|---|---|
| `user_outbox_events_created_total` | Counter | `event_type` | Outbox rows created |
| `user_event_publish_total` | Counter | `event_type`, `result` | Publish success/failure count |
| `user_outbox_pending` | Gauge | none | Pending rows |
| `user_outbox_oldest_age_seconds` | Gauge | none | Oldest pending event age |
| `user_event_publish_duration_ms` | Histogram | `event_type` | MQ publish latency |

### Trace spans

| Span | Meaning |
|---|---|
| `user.usecase.create_user` | Main user creation usecase |
| `user.outbox.insert` | Outbox row insert |
| `user.outbox.lock_pending` | Worker pending fetch |
| `user.events.publish` | MQ publish call |

---

## 🧩 Step 16: Failure Handling

| Failure | Behavior |
|---|---|
| Validation fail | No DB write, no event |
| User DB insert fail | No event |
| Outbox insert fail | Transaction fail, caller gets error |
| MQ down | Business write already committed; event remains pending |
| Worker crash | Pending/locked rows retry after `locked_until` expires |
| Duplicate publish | Consumer dedupes by `event_id` |
| Poison message | Move to DLQ/manual review after max attempts |
| Slow consumer | Queue lag alert |

### Why outbox insert failure fails request?

For these domain events, event is part of business side effect. Agar user create ho gaya but event outbox me nahi gaya, downstream consumers miss karenge. Isliye transaction me outbox insert fail ho to full transaction rollback karna safer hai.

---

## 🧩 Step 17: Testing Plan

### Unit tests

| Test | Expected |
|---|---|
| `NewEnvelope` creates unique `event_id` | IDs unique and prefix `evt_` |
| `UserCreated` payload does not include raw password/token | Sensitive fields absent |
| `AddressUpdated` payload excludes full address lines | Only city/state/country allowed |
| `SellerApproved` only emitted on transition to `active` | No event for rejected/suspended in Task 8 |
| Backoff increases attempts | Next attempt moves forward |

### Repository tests

| Test | Expected |
|---|---|
| Insert outbox row | Row status `pending` |
| Duplicate `event_id` | Unique constraint error |
| Lock pending rows | Rows locked once |
| Mark published | Status `published`, `published_at` set |
| Mark failed | Attempts increment, next attempt set |

### Usecase transaction tests

| Scenario | Expected |
|---|---|
| Create user success | User row + `UserCreated` outbox row |
| Create user DB rollback | No outbox row |
| Address update success | Address row updated + `AddressUpdated` outbox row |
| Seller approve success | Seller active + `SellerApproved` outbox row |
| Seller already active | No duplicate approval event unless business explicitly allows replay |

### Worker tests

| Scenario | Expected |
|---|---|
| Publish success | Outbox row marked `published` |
| Publish failure | Row marked `failed`, attempts increment |
| Max attempts crossed | Row marked `dead` or DLQ path used |
| Worker restart | Pending rows still publish |

---

## 🧩 Step 18: Manual Verification Queries

### Pending events

```sql
SELECT event_id, event_type, aggregate_id, status, attempts, created_at
FROM user_outbox_events
WHERE status IN ('pending', 'failed')
ORDER BY created_at ASC;
```

### Published count by event type

```sql
SELECT event_type, status, COUNT(*) AS total
FROM user_outbox_events
GROUP BY event_type, status;
```

### Oldest pending event age

```sql
SELECT TIMESTAMPDIFF(SECOND, MIN(created_at), UTC_TIMESTAMP()) AS oldest_pending_age_seconds
FROM user_outbox_events
WHERE status IN ('pending', 'failed');
```

### Inspect one event payload

```sql
SELECT JSON_PRETTY(payload)
FROM user_outbox_events
WHERE event_id = 'evt_user_created_123';
```

---

## 🧩 Step 19: Event Flow Examples

### `UserCreated` flow

```mermaid
sequenceDiagram
    autonumber
    participant Auth as Auth Service
    participant User as User Service
    participant DB as user_db
    participant Outbox as user_outbox_events
    participant MQ as user.events
    participant Notif as Notification Service

    Auth->>User: CreateUser(auth_account_id, email, full_name)
    User->>DB: Insert users row
    User->>Outbox: Insert UserCreated pending event
    User-->>Auth: CreateUser success
    Outbox-->>MQ: Worker publishes UserCreated
    MQ-->>Notif: Consumer receives UserCreated
    Notif->>Notif: Send welcome/onboarding notification later
```

### `SellerApproved` flow

```mermaid
sequenceDiagram
    autonumber
    participant Admin as Superadmin Service
    participant User as User Service
    participant DB as seller_profiles
    participant Outbox as user_outbox_events
    participant MQ as user.events
    participant Rec as Recommendation/Analytics

    Admin->>User: ApproveSeller(seller_id, reason)
    User->>DB: Update status active, approved_by, approved_at
    User->>Outbox: Insert SellerApproved event
    User-->>Admin: Success
    Outbox-->>MQ: Worker publishes SellerApproved
    MQ-->>Rec: Consumer updates seller signal/projection
```

### `AddressUpdated` flow

```mermaid
sequenceDiagram
    autonumber
    participant Client as Frontend
    participant Gateway as API Gateway
    participant User as User Service
    participant DB as user_addresses
    participant Outbox as user_outbox_events
    participant MQ as user.events
    participant Analytics as Analytics Consumer

    Client->>Gateway: PATCH /api/v1/me/addresses/{address_id}
    Gateway->>User: UpdateAddress gRPC
    User->>DB: Update address row
    User->>Outbox: Insert AddressUpdated event
    User-->>Gateway: Updated address
    Gateway-->>Client: 200 OK
    Outbox-->>MQ: Worker publishes AddressUpdated
    MQ-->>Analytics: Consumer records profile/address change signal
```

---

## 🧩 Step 20: Event Schema Evolution

Event schemas future me change honge. Breaking changes avoid karne ke liye:

| Change | Safe? | Rule |
|---|---|---|
| Add optional field | ✅ Safe | Existing consumers ignore kar sakte hain |
| Remove field | ❌ Breaking | New version banao |
| Rename field | ❌ Breaking | New version banao |
| Change type string to number | ❌ Breaking | New version banao |
| Add new event type | ✅ Safe | Consumers opt-in karenge |

### Versioning pattern

```go
const UserCreatedVersion = 1
```

Agar breaking change aaye:

```text
UserCreated version 1 remains supported.
UserCreated version 2 publish/consume migration plan ke saath introduce hota hai.
```

---

## 🧩 Step 21: Security and Access Notes

| Risk | Protection |
|---|---|
| Raw PII broad event topic me leak | Minimal payload + hashes |
| Unauthorized producer event inject kare | Broker credentials per service |
| Consumer duplicate side effects | `event_id` idempotency |
| Event tampering | TLS/auth in broker, private network, optional signing future |
| Sensitive logs | Redaction rules |
| Cross-service DB coupling | Consumers gRPC/projections use karein, User DB direct nahi |

### Broker access rule

Only User Service ko `user.events` publish permission mile. Consumers ko read permission mile, publish nahi.

---

## 🧩 Step 22: REST/gRPC API Impact

Task 8 me new public API add nahi hoti.

| Existing flow | Event side effect |
|---|---|
| `CreateUser` gRPC | `UserCreated` outbox event |
| `POST /api/v1/me/addresses` via Gateway | `AddressUpdated` with `change_type=created` |
| `PATCH /api/v1/me/addresses/{address_id}` via Gateway | `AddressUpdated` with `change_type=updated` |
| `DELETE /api/v1/me/addresses/{address_id}` via Gateway | `AddressUpdated` with `change_type=deleted` |
| Superadmin seller approval | `SellerApproved` |

### Response shape

Public REST response me event internals expose nahi karne:

```json
{
  "address_id": "addr_123",
  "city": "Delhi",
  "state": "Delhi",
  "country": "IN",
  "is_default": true
}
```

Event ID internal debugging ke liye logs/traces/outbox me rahega. User-facing API ko MQ implementation details nahi pata honi chahiye.

---

## 🧩 Step 23: Clean Rollout Plan

1. Outbox table migration deploy karo.
2. Event structs and envelope builder add karo.
3. Outbox repository add karo.
4. `UserCreated` enqueue wire karo.
5. `AddressUpdated` enqueue wire karo.
6. `SellerApproved` enqueue wire karo.
7. Worker disabled mode me deploy karo and outbox rows verify karo.
8. RabbitMQ/Kafka publisher enable karo.
9. Consumers ko idempotent handling ke saath connect karo.
10. Metrics/alerts add karo for pending outbox and publish failures.

### Rollout safety

| Phase | Safety check |
|---|---|
| Migration only | Existing service behavior unchanged |
| Enqueue only | Outbox rows created, MQ not required |
| Worker enabled | Events publish start |
| Consumers enabled | Downstream effects begin |

---

## 📊 Event State Machine

```mermaid
stateDiagram-v2
    [*] --> pending
    pending --> published: publish success
    pending --> failed: publish error
    failed --> pending: next_attempt_at reached
    failed --> dead: max attempts exceeded
    dead --> pending: manual replay
    published --> [*]
```

### Status explanation

| Status | Meaning |
|---|---|
| `pending` | Event ready for publish |
| `failed` | Last publish attempt failed, retry scheduled |
| `published` | Broker accepted event |
| `dead` | Max attempts crossed, manual review needed |

---

## 🧪 Sample Test Case Code

```go
func TestCreateUserEnqueuesUserCreated(t *testing.T) {
    ctx := withActorAndTrace(context.Background(), "service:auth-service", "req_1", "trace_1")

    tx := testDB.BeginTx(t, ctx)
    recorder := events.NewOutboxRecorder(outboxRepo, "user.events")

    user := domain.User{
        UserID:        "usr_123",
        AuthAccountID: "auth_123",
        Email:         "test@example.com",
        Status:        "active",
        CreatedAt:     time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC),
    }

    err := recorder.RecordUserCreated(ctx, tx, domain.UserCreatedPayload{
        UserID:        user.UserID,
        AuthAccountID: user.AuthAccountID,
        Status:        user.Status,
        EmailHash:     hashForEvent(user.Email),
        CreatedBy:     "service:auth-service",
        CreatedAt:     user.CreatedAt,
    })
    require.NoError(t, err)
    require.NoError(t, tx.Commit())

    row := outboxRepo.MustFindByAggregate(t, "user", "usr_123")
    require.Equal(t, "UserCreated", row.EventType)
    require.Equal(t, "pending", row.Status)
    require.NotContains(t, string(row.Payload), "test@example.com")
}
```

### Test explanation

- Event created after user creation path.
- Outbox status `pending` hai.
- Raw email payload me nahi hai.
- Event type correct hai.

---

## 🧯 Common Mistakes Avoid Karo

| Mistake | Problem | Better approach |
|---|---|---|
| DB commit ke baad direct MQ publish only | Crash hua to event lost | Outbox use karo |
| Event me full address/phone/email daalna | PII leak risk | Minimal payload/hash |
| Consumer non-idempotent banana | Duplicate side effects | `event_id` processed table |
| Event publish fail pe silently ignore | Downstream data missing | Outbox retry and alert |
| Gateway se event publish karna | Domain owner bypass | User Service publish kare |
| Dusri service ko user_db read karne dena | Service boundary break | gRPC or event projection |
| Version field skip karna | Future schema migration hard | `version` always include |
| `updated_at` local timezone me bhejna | Time bugs | UTC only |

---

## ✅ Completion Checklist

User Service Task 8 complete tab maana jayega jab:

- [x] `TaskImplementation/User Service/task8.md` created
- [x] Scope only User Service Task 8 tak limited hai
- [x] `UserCreated` event contract documented
- [x] `SellerApproved` event contract documented
- [x] `AddressUpdated` event contract documented
- [x] Outbox pattern documented
- [x] `user_outbox_events` migration example included
- [x] RabbitMQ publisher example included
- [x] Kafka alternate example included
- [x] External libraries/tools explained with install commands
- [x] Folder structure documented
- [x] Mermaid architecture and flow diagrams included
- [x] PII/security rules documented
- [x] Testing and verification plan included
- [x] No backend implementation beyond requested documentation was created

---

## 🏁 Final Takeaway

User Service Task 8 ka core outcome ye hai ki User Service apne important domain facts ko reliable events me convert kare:

- `UserCreated` se signup/profile creation downstream systems ko pata chalega.
- `SellerApproved` se seller activation notification, analytics, recommendation, and projections update ho sakenge.
- `AddressUpdated` se address-related projections and analytics eventually consistent rahenge.

Is design me User Service apni DB ownership maintain karta hai, consumers direct `user_db` read nahi karte, aur outbox pattern event loss avoid karta hai. Simple words me: **profile write pehle durable, event publish eventually reliable.**

