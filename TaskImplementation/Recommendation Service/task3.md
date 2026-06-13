# 🧠 Recommendation Service - Task 3: Event Ingestion

![Task](https://img.shields.io/badge/Task-Recommendation%20Service%20%233-blue)
![Priority](https://img.shields.io/badge/Priority-P1-orange)
![Dependency](https://img.shields.io/badge/Dependency-Kafka%20%2F%20RabbitMQ-yellow)
![Status](https://img.shields.io/badge/Status-Documented-success)
![Guide](https://img.shields.io/badge/Guide-Hinglish-ff69b4)
![Scope](https://img.shields.io/badge/Scope-Event%20Ingestion-purple)
![Storage](https://img.shields.io/badge/Writes-MongoDB%20user__interactions-47A248)

## 📌 Task Summary

| Field | Detail |
|---|---|
| Task Name | Event ingestion |
| Source | `docs/01-micro-tasks.md` → `Recommendation Service` → Task 3 |
| Priority | `P1` core event pipeline |
| Dependency | Kafka/RabbitMQ message queue |
| Previous Tasks | Task 1: recommendation types, Task 2: MongoDB plus Redis choice |
| Main Goal | Product views, add-to-cart, wishlist, aur purchase events consume karke `recommendation_db.user_interactions` me reliable interaction log banana |
| Output Type | Structured implementation guide |
| Not Included | Feature store schema, ranking engine, personalized scoring, gRPC endpoint, A/B testing |

> **Simple Hinglish goal:** Is task ka kaam hai Recommendation Service ko events sunna sikhana. Jab user product dekhe, cart me add kare, wishlist kare, ya purchase kare, to wo signal queue se consume hoga, validate hoga, normalized interaction banega, aur MongoDB me idempotently store hoga.

---

## ✅ Final Output Created

```text
TaskImplementation/
└── Recommendation Service/
    ├── task1.md
    ├── task2.md
    └── task3.md
```

### Why this structure?

- `TaskImplementation/` project ke task-wise guides ka central folder hai.
- `Recommendation Service/` already exist karta tha, isliye usko preserve kiya gaya.
- `task1.md` recommendation types define karta hai.
- `task2.md` MongoDB plus Redis storage decision define karta hai.
- `task3.md` sirf **Recommendation Service - Task 3** ka event ingestion guide hai.
- `backend/services/recommendation-service/` me actual service code create nahi kiya gaya, kyunki requested output documentation artifact hai.

---

## 🧭 Documents Studied

| Document | Kya use kiya gaya |
|---|---|
| `docs/01-micro-tasks.md` | Task 3 ka exact scope: product views, add-to-cart, wishlist, purchase events consume karna |
| `docs/04-microservice-design.md` | Recommendation Service responsibilities: user behavior ingestion, scoring, trending, personalized ranking |
| `docs/11-devops-external-services.md` | Kafka/RabbitMQ choice, topics, event envelope, retry/DLQ rules |
| `docs/12-logging-monitoring-scalability.md` | Consumer lag, async events, graceful degradation, recommendation cache TTL |
| `docs/03-folder-structure.md` | Future `internal/events/interaction_consumer.go` folder placement |
| `docs/05-database-design.md` | Batch consumers and async events for recommendation features |
| `database/mongodb-schema-design.md` | `recommendation_db.user_interactions` collection and indexes |
| `api/master-api.json` | Future recommendation API/event contracts |
| `TaskImplementation/Recommendation Service/task1.md` | Recommendation types and contexts |
| `TaskImplementation/Recommendation Service/task2.md` | MongoDB/Redis decision and cache invalidation idea |

---

## 🧱 Task Boundary

### ✅ Included in Task 3

- Queue se recommendation-relevant events consume karna
- Supported events define karna:
  - `ProductViewed`
  - `CartItemAdded`
  - `WishlistItemAdded`
  - `WishlistItemRemoved`
  - `OrderPaid` / `PurchaseCompleted`
- Event envelope validate karna
- Payload ko normalized `user_interactions` document me map karna
- Idempotent MongoDB writes define karna
- Duplicate event safe handling define karna
- Retry, backoff, aur dead-letter queue strategy define karna
- Consumer group/topic naming define karna
- Redis cache invalidation ko best-effort side effect ke roop me define karna
- Logs, metrics, tracing fields define karna
- Beginner-friendly Go code examples dena
- Architecture and flow diagrams add karna

### ❌ Not Included in Task 3

- Product/Cart/Wishlist/Order services me event producer implement karna
- Detailed feature store schema banana, kyunki wo Task 4 hai
- Trending/category/seller ranking engine banana, kyunki wo Task 5 hai
- Personalized ranking algorithm banana, kyunki wo Task 6 hai
- `GetRecommendations(user_id, context)` gRPC endpoint banana, kyunki wo Task 7 hai
- A/B testing hooks implement karna, kyunki wo Task 8 hai
- ML embeddings/vector search implement karna
- Frontend recommendation UI banana

> 🟢 **Rule:** Task 3 ka output raw interaction pipeline hai. Ye sirf signals collect karega. Ranking/scoring ka final decision later tasks me hoga.

---

## 🧩 Event Ingestion Overview

### What happens in Task 3?

```text
User action happens
→ Producer service publishes event
→ Message queue stores event
→ Recommendation Service consumer reads event
→ Event is validated
→ Payload is normalized
→ Interaction document is inserted into MongoDB
→ Related recommendation cache keys are marked stale or deleted
→ Metrics/logs are emitted
```

### Supported source events

| Source Service | Source Event | Normalized Event Type | Why Important |
|---|---|---|---|
| Session Service | `ProductViewed` | `product_view` | User interest ka light signal |
| Cart Service | `CartItemAdded` | `add_to_cart` | Strong purchase intent signal |
| Wishlist Service | `WishlistItemAdded` | `wishlist_add` | Preference signal |
| Wishlist Service | `WishlistItemRemoved` | `wishlist_remove` | Preference removal signal |
| Order Service | `OrderPaid` / `PurchaseCompleted` | `purchase` | Strongest conversion signal |

### Event weight idea

| Normalized Event | Weight | Meaning |
|---|---:|---|
| `product_view` | `1` | Light interest |
| `wishlist_add` | `3` | Medium interest |
| `wishlist_remove` | `-2` | Interest reduced |
| `add_to_cart` | `4` | Strong intent |
| `purchase` | `8` | Strong conversion |

> 🟡 **Important:** Ye weights Task 3 me final ranking ke liye use nahi ho rahe. Ye sirf raw interaction ke saath store honge, taaki Task 4/5/6 me feature store aur ranking use kar sake.

---

## 🏗️ High-Level Architecture

```mermaid
flowchart LR
    UI[User App / Web] --> Session[Session Service]
    UI --> Cart[Cart Service]
    UI --> Wishlist[Wishlist Service]
    UI --> Order[Order Service]

    Session -->|ProductViewed| Queue[(Kafka / RabbitMQ)]
    Cart -->|CartItemAdded| Queue
    Wishlist -->|WishlistItemAdded / Removed| Queue
    Order -->|OrderPaid / PurchaseCompleted| Queue

    Queue --> Consumer[Recommendation Interaction Consumer]
    Consumer --> Validator[Validate Envelope + Payload]
    Validator --> Mapper[Normalize Interaction]
    Mapper --> Mongo[(MongoDB recommendation_db.user_interactions)]
    Mapper --> Redis[(Redis cache invalidation - best effort)]
    Consumer --> Metrics[Logs / Metrics / Traces]
    Consumer --> DLQ[(Dead Letter Queue)]
```

### Hinglish explanation

- User action kisi source service me hota hai.
- Source service event publish karta hai.
- Recommendation Service direct source DB read nahi karegi.
- Consumer queue se event read karega.
- Valid event MongoDB me store hoga.
- Invalid/non-retryable event DLQ me jayega.
- Duplicate event safely ignore hoga.

---

## 🎯 Topic / Queue Strategy

### Recommended approach: dedicated fan-in topic

Recommendation Service ke liye best beginner-friendly approach hai ki producers recommendation-relevant events ek dedicated topic me publish karein:

```text
recommendation.events
```

### Supporting retry/DLQ topics

```text
recommendation.events.retry
recommendation.events.dlq
```

### Consumer group

```text
recommendation-service-v1
```

### Why dedicated topic?

| Benefit | Explanation |
|---|---|
| Simpler consumer | Consumer ko sirf recommendation-relevant events milte hain |
| Less filtering | Source topics ke unrelated events skip nahi karne padte |
| Easy replay | Recommendation data rebuild karna simpler hota hai |
| Independent scaling | Recommendation consumer apni speed se scale ho sakta hai |

### Alternative approach

Agar dedicated topic available nahi hai, Recommendation Service ye topics directly consume kar sakti hai:

```text
session.events
order.events
wishlist.events
cart.events
```

Project docs me common topics list me `session.events`, `order.events`, aur `recommendation.events` already mention hain. Isliye Task 3 ke liye preferred contract ye hai:

```text
Source service → recommendation.events → Recommendation Service
```

---

## 📦 Event Envelope Contract

Docs ke event envelope ko base banaya gaya:

```json
{
  "event_id": "evt_123",
  "event_type": "ProductViewed",
  "version": 1,
  "occurred_at": "2026-05-18T14:00:00Z",
  "producer": "session-service",
  "trace_id": "trace_123",
  "payload": {}
}
```

### Envelope fields

| Field | Required | Why |
|---|---:|---|
| `event_id` | Yes | Idempotency and debugging |
| `event_type` | Yes | Mapping decide karne ke liye |
| `version` | Yes | Schema evolution safe rakhne ke liye |
| `occurred_at` | Yes | User action ka actual time |
| `producer` | Yes | Source service identify karne ke liye |
| `trace_id` | Recommended | Distributed tracing/debugging |
| `payload` | Yes | Business data |

### Go example: generic envelope

```go
package events

import "time"

type Envelope[T any] struct {
    EventID    string    `json:"event_id"`
    EventType  string    `json:"event_type"`
    Version    int       `json:"version"`
    OccurredAt time.Time `json:"occurred_at"`
    Producer   string    `json:"producer"`
    TraceID    string    `json:"trace_id"`
    Payload    T         `json:"payload"`
}
```

---

## 🧾 Event Payload Examples

### 1. Product viewed

```json
{
  "event_id": "evt_view_001",
  "event_type": "ProductViewed",
  "version": 1,
  "occurred_at": "2026-05-27T10:30:00Z",
  "producer": "session-service",
  "trace_id": "trace_abc",
  "payload": {
    "user_id": "user_123",
    "anonymous_id": "anon_456",
    "session_id": "sess_789",
    "product_id": "prod_123",
    "variant_id": "var_1",
    "category_id": "cat_shoes",
    "seller_id": "seller_456",
    "page": "product_detail"
  }
}
```

### 2. Cart item added

```json
{
  "event_id": "evt_cart_001",
  "event_type": "CartItemAdded",
  "version": 1,
  "occurred_at": "2026-05-27T10:35:00Z",
  "producer": "cart-service",
  "trace_id": "trace_cart",
  "payload": {
    "user_id": "user_123",
    "anonymous_id": "anon_456",
    "cart_id": "cart_123",
    "product_id": "prod_123",
    "variant_id": "var_1",
    "category_id": "cat_shoes",
    "seller_id": "seller_456",
    "quantity": 2
  }
}
```

### 3. Wishlist item added

```json
{
  "event_id": "evt_wish_001",
  "event_type": "WishlistItemAdded",
  "version": 1,
  "occurred_at": "2026-05-27T10:40:00Z",
  "producer": "wishlist-service",
  "trace_id": "trace_wish",
  "payload": {
    "user_id": "user_123",
    "anonymous_id": "anon_456",
    "wishlist_id": "wish_123",
    "product_id": "prod_123",
    "variant_id": "var_1",
    "category_id": "cat_shoes",
    "seller_id": "seller_456"
  }
}
```

### 4. Purchase event

Purchase event me ek order ke andar multiple products ho sakte hain. Consumer ek event se multiple `purchase` interactions generate karega.

```json
{
  "event_id": "evt_order_001",
  "event_type": "OrderPaid",
  "version": 1,
  "occurred_at": "2026-05-27T11:00:00Z",
  "producer": "order-service",
  "trace_id": "trace_order",
  "payload": {
    "user_id": "user_123",
    "anonymous_id": "anon_456",
    "order_id": "order_123",
    "items": [
      {
        "product_id": "prod_123",
        "variant_id": "var_1",
        "category_id": "cat_shoes",
        "seller_id": "seller_456",
        "quantity": 2,
        "unit_price_amount": 299900,
        "currency": "INR"
      }
    ]
  }
}
```

---

## 🗃️ MongoDB Target Collection

Task 2 me MongoDB choose ho chuka hai. Task 3 me consumer `recommendation_db.user_interactions` collection me write karega.

### Recommended document shape

```json
{
  "_id": "interaction_evt_view_001_prod_123",
  "dedupe_key": "evt_view_001#product_view#prod_123#var_1",
  "event_id": "evt_view_001",
  "event_type": "ProductViewed",
  "normalized_event_type": "product_view",
  "version": 1,
  "producer": "session-service",
  "trace_id": "trace_abc",
  "user_id": "user_123",
  "anonymous_id": "anon_456",
  "session_id": "sess_789",
  "product_id": "prod_123",
  "variant_id": "var_1",
  "category_id": "cat_shoes",
  "seller_id": "seller_456",
  "weight": 1,
  "quantity": 1,
  "metadata": {
    "page": "product_detail"
  },
  "occurred_at": "2026-05-27T10:30:00Z",
  "received_at": "2026-05-27T10:30:01Z"
}
```

### Why `dedupe_key`?

At-least-once messaging me same event duplicate aa sakta hai. Purchase event ek order me multiple products contain kar sakta hai, isliye sirf `event_id` unique rakhna enough nahi hai.

Recommended unique key:

```text
{event_id}#{normalized_event_type}#{product_id}#{variant_id}
```

Example:

```text
evt_order_001#purchase#prod_123#var_1
```

### Recommended indexes

```javascript
use recommendation_db

db.user_interactions.createIndex(
  { dedupe_key: 1 },
  { unique: true }
)

db.user_interactions.createIndex(
  { user_id: 1, occurred_at: -1 }
)

db.user_interactions.createIndex(
  { anonymous_id: 1, occurred_at: -1 }
)

db.user_interactions.createIndex(
  { product_id: 1, normalized_event_type: 1, occurred_at: -1 }
)

db.user_interactions.createIndex(
  { category_id: 1, normalized_event_type: 1, occurred_at: -1 }
)

db.user_interactions.createIndex(
  { occurred_at: 1 },
  { expireAfterSeconds: 15552000 }
)
```

### Index explanation

| Index | Why |
|---|---|
| `dedupe_key` unique | Same event duplicate insert na ho |
| `user_id + occurred_at` | User behavior history read karne ke liye |
| `anonymous_id + occurred_at` | Guest personalization ke liye |
| `product_id + event_type + occurred_at` | Similar/trending calculations ke liye |
| `category_id + event_type + occurred_at` | Category popular/trending ke liye |
| `occurred_at` TTL | Old raw events auto-cleanup |

---

## 🪜 Step-by-Step Implementation Guide

### Step 1: Task scope confirm kiya

`docs/01-micro-tasks.md` me Task 3 ka requirement hai:

```text
Product views, add-to-cart, wishlist, purchase events consume karo.
Dependency: Kafka/RabbitMQ.
```

Iska matlab:

- Recommendation Service ab queue consumer banega.
- Wo direct Product/Session/Order/Wishlist database read nahi karega.
- Raw event data MongoDB me apni collection me store karega.
- Future tasks isi raw interaction data se features/ranking generate karenge.

---

### Step 2: Queue topic choose kiya

Recommended topic:

```text
recommendation.events
```

Retry and DLQ:

```text
recommendation.events.retry
recommendation.events.dlq
```

Consumer group:

```text
recommendation-service-v1
```

### Why Kafka preferred?

Project docs ke according Kafka high-throughput event streaming ke liye recommended hai. Recommendation events high volume ho sakte hain, especially product views. Isliye Kafka ka consumer group model future scaling ke liye better fit hai.

### RabbitMQ acceptable fallback

RabbitMQ simpler routing/command queues ke liye acceptable hai. Agar local stack me RabbitMQ already use ho raha hai, same ingestion logic RabbitMQ queue consumer se bhi implement ho sakta hai.

---

### Step 3: Supported event names freeze kiye

Consumer ko sirf supported events process karne chahiye.

```go
package domain

type SourceEventType string

const (
    EventProductViewed     SourceEventType = "ProductViewed"
    EventCartItemAdded     SourceEventType = "CartItemAdded"
    EventWishlistItemAdded SourceEventType = "WishlistItemAdded"
    EventWishlistItemRemoved SourceEventType = "WishlistItemRemoved"
    EventOrderPaid         SourceEventType = "OrderPaid"
    EventPurchaseCompleted SourceEventType = "PurchaseCompleted"
)
```

Normalized event types:

```go
package domain

type InteractionType string

const (
    InteractionProductView    InteractionType = "product_view"
    InteractionAddToCart      InteractionType = "add_to_cart"
    InteractionWishlistAdd    InteractionType = "wishlist_add"
    InteractionWishlistRemove InteractionType = "wishlist_remove"
    InteractionPurchase       InteractionType = "purchase"
)
```

---

### Step 4: Common payload model define kiya

Most interaction events single product ke around hote hain.

```go
package events

type ProductInteractionPayload struct {
    UserID      string         `json:"user_id"`
    AnonymousID string         `json:"anonymous_id"`
    SessionID   string         `json:"session_id"`
    ProductID   string         `json:"product_id"`
    VariantID   string         `json:"variant_id"`
    CategoryID  string         `json:"category_id"`
    SellerID    string         `json:"seller_id"`
    Quantity    int            `json:"quantity"`
    Metadata    map[string]any `json:"metadata"`
}
```

Purchase event multiple items carry kar sakta hai.

```go
package events

type PurchasePayload struct {
    UserID      string         `json:"user_id"`
    AnonymousID string         `json:"anonymous_id"`
    OrderID     string         `json:"order_id"`
    Items       []PurchaseItem `json:"items"`
}

type PurchaseItem struct {
    ProductID       string `json:"product_id"`
    VariantID       string `json:"variant_id"`
    CategoryID      string `json:"category_id"`
    SellerID        string `json:"seller_id"`
    Quantity        int    `json:"quantity"`
    UnitPriceAmount int64  `json:"unit_price_amount"`
    Currency        string `json:"currency"`
}
```

---

### Step 5: Validation rules add kiye

Consumer DB me bad data insert nahi karega.

| Validation | Rule |
|---|---|
| Envelope | `event_id`, `event_type`, `version`, `occurred_at`, `producer` required |
| Version | MVP me `version = 1` support karo |
| Event type | Unknown event ko skip ya DLQ karo |
| Identity | `user_id` ya `anonymous_id` me se at least one required |
| Product event | `product_id` required |
| Purchase event | `items` non-empty required |
| Time | Invalid `occurred_at` reject karo |
| PII | Email, phone, name, raw IP store mat karo |

### Validation helper example

```go
package events

import "errors"

func ValidateEnvelope(eventID string, eventType string, version int, producer string) error {
    if eventID == "" {
        return errors.New("event_id is required")
    }
    if eventType == "" {
        return errors.New("event_type is required")
    }
    if version != 1 {
        return errors.New("unsupported event version")
    }
    if producer == "" {
        return errors.New("producer is required")
    }
    return nil
}
```

---

### Step 6: Event ko normalized interaction me map kiya

Source services alag naming use kar sakti hain. Recommendation Service ke andar ek consistent model use hoga.

```go
package domain

import "time"

type Interaction struct {
    ID                  string         `bson:"_id"`
    DedupeKey           string         `bson:"dedupe_key"`
    EventID             string         `bson:"event_id"`
    SourceEventType     string         `bson:"event_type"`
    NormalizedEventType string         `bson:"normalized_event_type"`
    Version             int            `bson:"version"`
    Producer            string         `bson:"producer"`
    TraceID             string         `bson:"trace_id"`
    UserID              string         `bson:"user_id,omitempty"`
    AnonymousID         string         `bson:"anonymous_id,omitempty"`
    SessionID           string         `bson:"session_id,omitempty"`
    ProductID           string         `bson:"product_id"`
    VariantID           string         `bson:"variant_id,omitempty"`
    CategoryID          string         `bson:"category_id,omitempty"`
    SellerID            string         `bson:"seller_id,omitempty"`
    Weight              int            `bson:"weight"`
    Quantity            int            `bson:"quantity"`
    Metadata            map[string]any `bson:"metadata,omitempty"`
    OccurredAt          time.Time      `bson:"occurred_at"`
    ReceivedAt          time.Time      `bson:"received_at"`
}
```

### Mapping table

| Source Event | Normalized Type | Weight | Mapping Note |
|---|---|---:|---|
| `ProductViewed` | `product_view` | `1` | Single interaction |
| `CartItemAdded` | `add_to_cart` | `4` | Quantity preserve karo |
| `WishlistItemAdded` | `wishlist_add` | `3` | Single interaction |
| `WishlistItemRemoved` | `wishlist_remove` | `-2` | Negative preference signal |
| `OrderPaid` | `purchase` | `8` | Har order item ke liye one interaction |
| `PurchaseCompleted` | `purchase` | `8` | Same as `OrderPaid` |

---

### Step 7: Idempotent MongoDB write implement kiya

Message queues usually at-least-once delivery deti hain. Same event duplicate aa sakta hai.

Recommended rule:

```text
Mongo insert succeeds → event processed
Mongo duplicate key → event already processed, treat as success
Mongo temporary error → retry
Validation error → DLQ
```

### Repository example

```go
package repository

import (
    "context"

    "go.mongodb.org/mongo-driver/mongo"
)

type InteractionRepository struct {
    collection *mongo.Collection
}

func (r *InteractionRepository) Insert(ctx context.Context, doc any) error {
    _, err := r.collection.InsertOne(ctx, doc)
    if mongo.IsDuplicateKeyError(err) {
        return nil
    }
    return err
}
```

### Why duplicate is success?

Consumer retry ke time same event dobara aa sakta hai. Agar document already inserted hai, pipeline ka final state correct hai. Isliye duplicate ko failure treat karna galat hoga.

---

### Step 8: Purchase event ko multiple interactions me expand kiya

Purchase event order-level hota hai, but recommendation ko product-level signals chahiye.

```mermaid
flowchart TD
    A[OrderPaid event] --> B{Items count}
    B --> C[Item 1 -> purchase interaction]
    B --> D[Item 2 -> purchase interaction]
    B --> E[Item N -> purchase interaction]
    C --> F[(user_interactions)]
    D --> F
    E --> F
```

### Example rule

```text
OrderPaid event_id = evt_order_001
Item product_id = prod_123
Interaction dedupe_key = evt_order_001#purchase#prod_123#var_1
```

This avoids duplicate writes and still allows one order event to generate many product-level interactions.

---

### Step 9: Redis cache invalidation best-effort rakha

Task 2 me Redis recommendation cache define hua tha. Task 3 me new interaction aane ke baad related cached recommendations stale ho sakti hain.

Task 3 me consumer ranking recompute nahi karega. Sirf best-effort invalidation/dirty marking karega.

### Suggested invalidation rules

| Event | Cache Action |
|---|---|
| Any user event with `user_id` | Delete `reco:v1:personalized:user:{user_id}` |
| Any guest event with `anonymous_id` | Delete `reco:v1:personalized:anon:{anonymous_id}` |
| Product/category event | Mark `reco:dirty:category:{category_id}` |
| Seller event | Mark `reco:dirty:seller:{seller_id}` |
| Purchase event | Mark product/category/seller dirty |

### Redis example commands

```text
DEL reco:v1:personalized:user:user_123
DEL reco:v1:personalized:anon:anon_456
SETEX reco:dirty:category:cat_shoes 900 "1"
SETEX reco:dirty:seller:seller_456 900 "1"
```

> 🟡 **Important:** Redis failure se event processing fail nahi honi chahiye. MongoDB write primary outcome hai. Cache TTL 15 min hai, so stale cache temporary safe hai.

---

### Step 10: Consumer handler flow define kiya

```mermaid
flowchart TD
    A[Receive message] --> B[Parse JSON envelope]
    B --> C{Envelope valid?}
    C -->|No| D[Send to DLQ]
    C -->|Yes| E{Supported event?}
    E -->|No| F[Skip or DLQ with reason]
    E -->|Yes| G[Map to interactions]
    G --> H{Interactions valid?}
    H -->|No| D
    H -->|Yes| I[Insert into MongoDB]
    I --> J{Duplicate key?}
    J -->|Yes| K[Treat as success]
    J -->|No| L{Insert ok?}
    L -->|No transient| M[Retry with backoff]
    L -->|No permanent| D
    L -->|Yes| N[Best-effort Redis invalidation]
    K --> O[Commit/Ack message]
    N --> O
```

### Handler pseudo-code

```go
func (c *InteractionConsumer) Handle(ctx context.Context, raw []byte) error {
    envelope, err := c.decoder.Decode(raw)
    if err != nil {
        return c.dlq.Publish(ctx, raw, "invalid_json")
    }

    if err := envelope.Validate(); err != nil {
        return c.dlq.Publish(ctx, raw, err.Error())
    }

    interactions, err := c.mapper.Map(envelope)
    if err != nil {
        return c.dlq.Publish(ctx, raw, err.Error())
    }

    for _, interaction := range interactions {
        if err := c.repo.Insert(ctx, interaction); err != nil {
            return err
        }
    }

    c.cacheInvalidator.InvalidateBestEffort(ctx, interactions)
    return nil
}
```

---

### Step 11: Kafka commit rule define kiya

Kafka me message commit tab karna hai jab MongoDB write complete ho jaye.

```text
Read message
→ Process message
→ MongoDB insert success or duplicate
→ Best-effort cache invalidation
→ Commit offset
```

### Kafka consumer config

```text
KAFKA_BROKERS=kafka:9092
RECOMMENDATION_EVENTS_TOPIC=recommendation.events
RECOMMENDATION_EVENTS_GROUP=recommendation-service-v1
RECOMMENDATION_EVENTS_DLQ_TOPIC=recommendation.events.dlq
```

### kafka-go example

```go
reader := kafka.NewReader(kafka.ReaderConfig{
    Brokers: []string{"kafka:9092"},
    Topic:   "recommendation.events",
    GroupID: "recommendation-service-v1",
})

for {
    msg, err := reader.FetchMessage(ctx)
    if err != nil {
        return err
    }

    if err := consumer.Handle(ctx, msg.Value); err != nil {
        // Do not commit. Retry policy will decide next step.
        continue
    }

    if err := reader.CommitMessages(ctx, msg); err != nil {
        return err
    }
}
```

---

### Step 12: RabbitMQ ack rule define kiya

RabbitMQ use karne par message `Ack` tab hoga jab DB write success ho.

```text
Receive delivery
→ Process
→ MongoDB insert success or duplicate
→ Ack
```

Failure:

```text
Transient error → Nack/retry
Permanent validation error → Publish DLQ → Ack original
```

### RabbitMQ queue naming

```text
exchange: ecommerce.events
routing key: recommendation.events
queue: recommendation.events.q
retry queue: recommendation.events.retry.q
dead-letter queue: recommendation.events.dlq
```

---

### Step 13: Retry and DLQ strategy add kiya

### Error categories

| Error Type | Example | Action |
|---|---|---|
| Invalid JSON | Body parse nahi hua | DLQ |
| Validation error | Missing `product_id` | DLQ |
| Unsupported version | `version = 99` | DLQ |
| Unknown event | `CouponViewed` | Skip or DLQ, config-based |
| Mongo temporary error | Network timeout | Retry |
| Mongo duplicate key | Same `dedupe_key` | Success |
| Redis error | Cache delete failed | Log only, success |

### Retry policy

```text
Attempt 1: immediate
Attempt 2: after 5 seconds
Attempt 3: after 30 seconds
Attempt 4: after 2 minutes
After max attempts: DLQ
```

### DLQ payload shape

```json
{
  "failed_at": "2026-05-27T10:45:00Z",
  "reason": "missing product_id",
  "consumer": "recommendation-service",
  "original_topic": "recommendation.events",
  "original_event": {
    "event_id": "evt_bad_001",
    "event_type": "ProductViewed"
  }
}
```

---

### Step 14: Observability add kiya

Consumer silent nahi hona chahiye. Har important path logs and metrics me visible hona chahiye.

### Logs

Recommended structured log fields:

```json
{
  "service": "recommendation-service",
  "component": "interaction_consumer",
  "event_id": "evt_view_001",
  "event_type": "ProductViewed",
  "normalized_event_type": "product_view",
  "producer": "session-service",
  "trace_id": "trace_abc",
  "result": "stored"
}
```

### Metrics

| Metric | Type | Labels |
|---|---|---|
| `recommendation_events_consumed_total` | Counter | `event_type`, `producer`, `result` |
| `recommendation_event_processing_duration_ms` | Histogram | `event_type` |
| `recommendation_event_dlq_total` | Counter | `reason`, `event_type` |
| `recommendation_event_duplicate_total` | Counter | `event_type` |
| `recommendation_mongo_insert_errors_total` | Counter | `error_type` |
| `recommendation_cache_invalidation_errors_total` | Counter | `cache_action` |
| `queue_consumer_lag` | Gauge | `topic`, `consumer_group` |

### Tracing

Trace propagation:

```text
Source request trace_id
→ Event envelope trace_id
→ Consumer processing span
→ MongoDB insert span
→ Redis invalidation span
```

---

### Step 15: Privacy and data safety define kiya

Recommendation events useful hain, but PII-heavy data store nahi karna chahiye.

### Store allowed

- `user_id`
- `anonymous_id`
- `session_id`
- `product_id`
- `variant_id`
- `category_id`
- `seller_id`
- `quantity`
- Non-sensitive metadata

### Avoid storing

- Email
- Phone number
- Full name
- Full address
- Raw IP address
- Payment card details
- Auth tokens

> 🟣 **Reason:** Recommendation Service ko behavior signal chahiye, user ka sensitive profile nahi.

---

## 📁 Clean Folder Structure

### Actual documentation output

```text
TaskImplementation/
└── Recommendation Service/
    ├── task1.md
    ├── task2.md
    └── task3.md
```

### Future backend implementation structure

Ye Task 3 ke real code implementation ke liye recommended structure hai. Is guide generation me ye files create nahi kiye gaye.

```text
backend/
└── services/
    └── recommendation-service/
        ├── cmd/
        │   └── server/
        │       └── main.go
        ├── internal/
        │   ├── config/
        │   │   └── config.go
        │   ├── domain/
        │   │   ├── interaction.go
        │   │   └── interaction_type.go
        │   ├── events/
        │   │   ├── envelope.go
        │   │   ├── interaction_consumer.go
        │   │   ├── interaction_mapper.go
        │   │   ├── interaction_validator.go
        │   │   └── dlq_publisher.go
        │   ├── repository/
        │   │   ├── mongo_interaction_repository.go
        │   │   └── redis_invalidation_repository.go
        │   └── observability/
        │       └── metrics.go
        └── deploy/
            └── README.md
```

### File responsibility

| File | Responsibility |
|---|---|
| `cmd/server/main.go` | Config load, Mongo/Redis/queue client init, consumer start |
| `internal/config/config.go` | Env vars parse karna |
| `internal/domain/interaction.go` | MongoDB interaction model |
| `internal/domain/interaction_type.go` | Supported normalized event types |
| `internal/events/envelope.go` | Event envelope model |
| `internal/events/interaction_consumer.go` | Queue message read/handle loop |
| `internal/events/interaction_mapper.go` | Source event to interaction mapping |
| `internal/events/interaction_validator.go` | Envelope and payload validation |
| `internal/events/dlq_publisher.go` | Bad events DLQ me publish karna |
| `internal/repository/mongo_interaction_repository.go` | Idempotent Mongo insert |
| `internal/repository/redis_invalidation_repository.go` | Best-effort cache invalidation |
| `internal/observability/metrics.go` | Consumer metrics |

---

## 🔄 Sequence Diagram

```mermaid
sequenceDiagram
    participant UI as User App
    participant Source as Source Service
    participant MQ as Kafka/RabbitMQ
    participant Reco as Recommendation Consumer
    participant Mongo as MongoDB
    participant Redis as Redis
    participant DLQ as DLQ

    UI->>Source: User action
    Source->>MQ: Publish recommendation event
    Reco->>MQ: Fetch message
    Reco->>Reco: Validate envelope
    alt Invalid event
        Reco->>DLQ: Publish failed event
        Reco->>MQ: Commit/Ack original
    else Valid event
        Reco->>Reco: Map to interaction docs
        Reco->>Mongo: Insert with dedupe_key
        alt Duplicate key
            Mongo-->>Reco: Already exists
            Reco->>MQ: Commit/Ack as success
        else Insert success
            Mongo-->>Reco: Stored
            Reco->>Redis: Best-effort invalidate dirty keys
            Reco->>MQ: Commit/Ack
        else Temporary DB error
            Mongo-->>Reco: Error
            Reco-->>MQ: Retry later
        end
    end
```

---

## ⚙️ Environment Variables

Future Recommendation Service `.env` me ye values useful hongi:

```env
SERVICE_NAME=recommendation-service
LOG_LEVEL=info

RECOMMENDATION_MONGO_URI=mongodb://mongo:27017
RECOMMENDATION_MONGO_DATABASE=recommendation_db
RECOMMENDATION_INTERACTIONS_COLLECTION=user_interactions

REDIS_ADDR=redis:6379
REDIS_PASSWORD=
REDIS_DB=0

QUEUE_PROVIDER=kafka
KAFKA_BROKERS=kafka:9092
RECOMMENDATION_EVENTS_TOPIC=recommendation.events
RECOMMENDATION_EVENTS_RETRY_TOPIC=recommendation.events.retry
RECOMMENDATION_EVENTS_DLQ_TOPIC=recommendation.events.dlq
RECOMMENDATION_EVENTS_GROUP=recommendation-service-v1

RABBITMQ_URL=amqp://ecommerce:ecommerce_password@rabbitmq:5672/ecommerce
RABBITMQ_EXCHANGE=ecommerce.events
RABBITMQ_RECOMMENDATION_QUEUE=recommendation.events.q
RABBITMQ_RECOMMENDATION_DLQ=recommendation.events.dlq

EVENT_MAX_RETRY_ATTEMPTS=4
EVENT_RETRY_BACKOFF_SECONDS=5,30,120
EVENT_SUPPORTED_VERSION=1
```

---

## 📦 External Libraries / Tools

> 🟢 **Task 3 status:** Is documentation task me koi dependency install/run nahi ki gayi. Neeche libraries/tools future code implementation ke liye documented hain.

### 1. Kafka

| Field | Detail |
|---|---|
| What | Distributed event streaming platform |
| Why used | Product views high volume hote hain; Kafka consumer groups and replay recommendation pipelines ke liye strong fit hain |
| Install local | Docker/Compose ke through Kafka service run karo |
| Use | `recommendation.events` consume, retry/DLQ topics maintain |

Example local Docker command:

```bash
docker run --name ecommerce-kafka -p 9092:9092 bitnami/kafka:3.7
```

> Real project me Platform Foundation local compose stack use karna better hoga.

### 2. RabbitMQ

| Field | Detail |
|---|---|
| What | Message broker with exchanges and queues |
| Why used | Simpler routing/queue setup ke liye acceptable fallback |
| Install local | RabbitMQ management Docker image |
| Use | `recommendation.events.q`, retry queue, DLQ |

Example:

```bash
docker run --name ecommerce-rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3.13-management-alpine
```

### 3. kafka-go

| Field | Detail |
|---|---|
| Package | `github.com/segmentio/kafka-go` |
| Why used | Go service me Kafka consumer/producer implement karne ke liye |
| Install | `go get github.com/segmentio/kafka-go` |
| Future file | `internal/events/interaction_consumer.go` |

Basic use:

```go
reader := kafka.NewReader(kafka.ReaderConfig{
    Brokers: []string{"kafka:9092"},
    Topic:   "recommendation.events",
    GroupID: "recommendation-service-v1",
})
```

### 4. amqp091-go

| Field | Detail |
|---|---|
| Package | `github.com/rabbitmq/amqp091-go` |
| Why used | Go service se RabbitMQ consume/publish karne ke liye |
| Install | `go get github.com/rabbitmq/amqp091-go` |
| Future file | `internal/events/interaction_consumer.go` if RabbitMQ selected |

### 5. MongoDB Go Driver

| Field | Detail |
|---|---|
| Package | `go.mongodb.org/mongo-driver/mongo` |
| Why used | `user_interactions` me idempotent insert karne ke liye |
| Install | `go get go.mongodb.org/mongo-driver/mongo` |
| Future file | `internal/repository/mongo_interaction_repository.go` |

### 6. go-redis

| Field | Detail |
|---|---|
| Package | `github.com/redis/go-redis/v9` |
| Why used | Personalized/trending cache keys invalidate ya dirty mark karne ke liye |
| Install | `go get github.com/redis/go-redis/v9` |
| Future file | `internal/repository/redis_invalidation_repository.go` |

### 7. Prometheus Client

| Field | Detail |
|---|---|
| Package | `github.com/prometheus/client_golang/prometheus` |
| Why used | Consumer metrics expose karne ke liye |
| Install | `go get github.com/prometheus/client_golang/prometheus` |
| Future file | `internal/observability/metrics.go` |

---

## 🧪 Testing Strategy

### Unit tests

| Test | Expected Result |
|---|---|
| `ProductViewed` maps to `product_view` | One interaction with weight `1` |
| `CartItemAdded` maps to `add_to_cart` | Quantity preserved, weight `4` |
| `WishlistItemAdded` maps to `wishlist_add` | One interaction with weight `3` |
| `WishlistItemRemoved` maps to `wishlist_remove` | One interaction with weight `-2` |
| `OrderPaid` with 3 items | Three `purchase` interactions |
| Missing `product_id` | Validation error |
| Duplicate `dedupe_key` | Repository returns success |
| Unsupported version | DLQ path |

### Integration tests

| Flow | Verification |
|---|---|
| Publish valid event | MongoDB document inserted |
| Publish same event twice | Only one document exists |
| Publish invalid JSON | DLQ receives failed event |
| Mongo temporary failure | Message retried |
| Redis down | Mongo write succeeds, cache error logged |

### Manual Mongo verification

```javascript
use recommendation_db

db.user_interactions.find({
  user_id: "user_123"
}).sort({
  occurred_at: -1
}).limit(5)
```

### Manual duplicate verification

```javascript
db.user_interactions.countDocuments({
  dedupe_key: "evt_view_001#product_view#prod_123#var_1"
})
```

Expected:

```text
1
```

---

## 🧯 Failure Handling Matrix

| Failure | Impact | Handling |
|---|---|---|
| Queue temporarily unavailable | Consumer cannot read new events | Consumer reconnects, events remain in broker |
| MongoDB temporarily unavailable | Events cannot be persisted | Retry with backoff, do not commit/ack too early |
| Redis unavailable | Cache invalidation skipped | Log metric, continue after Mongo success |
| Invalid payload | Cannot safely map event | Send to DLQ |
| Duplicate delivery | Same event received again | Unique `dedupe_key` makes it safe |
| Unknown event type | Consumer does not know mapping | Skip or DLQ based on config |
| High consumer lag | Recommendations become stale | Scale consumer replicas/partitions |

---

## 🔐 Security Checklist

| Area | Rule |
|---|---|
| PII | Email, phone, raw IP, address store mat karo |
| Auth data | Tokens/passwords kabhi event payload me nahi hone chahiye |
| Payload size | Maximum message size enforce karo |
| Schema version | Unsupported version DLQ karo |
| Broker security | Production me auth/TLS enable karo |
| MongoDB access | Recommendation Service ko sirf apne DB ka access do |
| Logs | Sensitive payload full dump mat karo |

---

## 📊 Data Flow by Event Type

### Product view flow

```mermaid
flowchart LR
    A[User opens product page] --> B[Session Service]
    B --> C[ProductViewed event]
    C --> D[(recommendation.events)]
    D --> E[Recommendation Consumer]
    E --> F[product_view interaction]
    F --> G[(MongoDB user_interactions)]
```

### Add-to-cart flow

```mermaid
flowchart LR
    A[User adds product to cart] --> B[Cart Service]
    B --> C[CartItemAdded event]
    C --> D[(recommendation.events)]
    D --> E[Recommendation Consumer]
    E --> F[add_to_cart interaction]
    F --> G[(MongoDB user_interactions)]
```

### Purchase flow

```mermaid
flowchart LR
    A[Payment succeeds] --> B[Order Service]
    B --> C[OrderPaid event with items]
    C --> D[(recommendation.events)]
    D --> E[Recommendation Consumer]
    E --> F[One purchase interaction per item]
    F --> G[(MongoDB user_interactions)]
```

---

## 🧠 Beginner-Friendly Mental Model

```text
Event = user action ka message
Consumer = message sunne wala worker
Mapper = source event ko internal shape me convert karne wala code
Interaction = recommendation-friendly product signal
Idempotency = duplicate message se duplicate data na banana
DLQ = bad events ka safe parking area
```

Simple example:

```text
User ne shoes dekhe
→ ProductViewed event aaya
→ Consumer ne event validate kiya
→ product_view interaction bana
→ MongoDB me save hua
→ Future ranking bol sakti hai: user shoes me interested hai
```

---

## ✅ Validation Checklist

| Check | Status |
|---|---|
| `TaskImplementation/Recommendation Service/` folder preserved | ✅ Done |
| `task3.md` created | ✅ Done |
| Task 3 scope event ingestion tak limited | ✅ Done |
| Product view event covered | ✅ Done |
| Add-to-cart event covered | ✅ Done |
| Wishlist add/remove events covered | ✅ Done |
| Purchase event covered | ✅ Done |
| Event envelope documented | ✅ Done |
| MongoDB `user_interactions` write shape documented | ✅ Done |
| Idempotency strategy documented | ✅ Done |
| Retry and DLQ strategy documented | ✅ Done |
| Kafka/RabbitMQ tools explained | ✅ Done |
| External Go libraries explained | ✅ Done |
| Folder structure included | ✅ Done |
| Mermaid diagrams included | ✅ Done |
| No ranking/feature-store/API implementation added | ✅ Done |

---

## 🧾 Acceptance Criteria

Task 3 complete tab maana jayega jab future code implementation me:

- Consumer `recommendation.events` topic/queue se events read kare.
- Supported event types validate ho.
- Invalid events DLQ me jaye.
- Valid events `user_interactions` me insert ho.
- Duplicate events duplicate documents na banaye.
- Purchase event multiple product interactions me expand ho.
- MongoDB write ke baad hi message commit/ack ho.
- Redis invalidation best-effort ho.
- Logs, metrics, and trace fields available ho.
- Recommendation API/ranking logic is task me implement na ho.

---

## 🚀 Quick Recap

Recommendation Service Task 3 ka final result ek reliable event ingestion blueprint hai:

- Queue input: `recommendation.events`
- Consumer group: `recommendation-service-v1`
- Durable store: `recommendation_db.user_interactions`
- Idempotency key: `dedupe_key`
- Retry path: retry topic/queue
- Failure parking: DLQ
- Cache side effect: best-effort invalidation

Iske baad next task naturally **Recommendation Service - Task 4: Feature store schema** hoga, jahan ye raw interaction data counters/features me convert hoga.
