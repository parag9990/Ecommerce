# 🛍️ Product Service - Task 7: Publish Search Events

![Task](https://img.shields.io/badge/Task-Product%20Service%20%237-blue)
![Priority](https://img.shields.io/badge/Priority-P1-orange)
![Dependency](https://img.shields.io/badge/Dependency-Message%20Queue-brightgreen)
![Status](https://img.shields.io/badge/Status-Documented-success)
![Guide](https://img.shields.io/badge/Guide-Hinglish-orange)
![Scope](https://img.shields.io/badge/Scope-Search%20Events-purple)

## 📌 Task Summary

| Field | Detail |
|---|---|
| Service | Product Service |
| Task No. | 7 |
| Task Name | Publish search events |
| Source | `docs/01-micro-tasks.md` -> `Product Service` -> Task 7 |
| Goal | Product changes ke baad Search Service ko async indexing event publish karna |
| Dependency | Message queue |
| Priority | P1 |
| Event Topic / Queue | `product.events` |
| Main Consumers | Search Service, Recommendation Service future me |
| Main Event Types | `ProductCreated`, `ProductUpdated`, `ProductPublished`, `ProductUnpublished`, `ProductInventoryChanged` |
| Output Type | Documentation-only implementation guide |
| Not Included | Search Service consumer implementation, Typesense schema creation, full reindex job, media metadata task |

> 🟢 **Simple Hinglish goal:** Is task me Product Service ko aisa design kiya gaya hai ki jab product create, update, publish, unpublish, ya stock availability change ho, tab Product Service ek reliable event publish kare. Search Service ye event consume karke Typesense index ko async update karega. Product Service source of truth rahega, Search Service sirf indexed/searchable copy rakhega.

---

## ✅ Final Output Created

```text
TaskImplementation/
└── Product Service/
    ├── task1.md
    ├── task2.md
    ├── task3.md
    ├── task4.md
    ├── task5.md
    ├── task6.md
    └── task7.md
```

### Why this structure?

| Path | Purpose |
|---|---|
| `TaskImplementation/` | Saare task-wise implementation guides ka central folder |
| `TaskImplementation/Product Service/` | Product Service ke implementation guides ko group karne ke liye |
| `task7.md` | Product search event publishing ka complete implementation guide |

> 🔵 **Important:** Is task me sirf `task7.md` create kiya gaya hai. Actual Go service code, proto file, message broker config, MongoDB index execution, Search Service consumer, Typesense indexer, ya tests repo me create nahi kiye gaye. Code examples implementation samjhane ke liye hain.

---

## 🧭 Docs Studied Before Implementation

| Document | Kya use kiya gaya |
|---|---|
| `docs/01-micro-tasks.md` | Product Service Task 7 ka exact scope identify kiya |
| `docs/02-system-architecture.md` | `ProductUpdated` -> Search async event flow align kiya |
| `docs/03-folder-structure.md` | Product Service ke `internal/events` folder pattern ko follow kiya |
| `docs/04-microservice-design.md` | Product Service responsibility: product index events emit karna |
| `docs/05-database-design.md` | Typesense searchable fields aur async event principle use kiya |
| `docs/11-devops-external-services.md` | Message queue topics, envelope, retry, DLQ rules align kiye |
| `docs/12-logging-monitoring-scalability.md` | Queue lag, async events, observability expectations align kiye |
| `TaskImplementation/Product Service/task1.md` | Product catalog model and Search Service boundary reuse ki |
| `TaskImplementation/Product Service/task3.md` | MongoDB `products` collection and future event publisher boundary li |
| `TaskImplementation/Product Service/task4.md` | Create/update/publish/unpublish trigger points reuse kiye |
| `TaskImplementation/Product Service/task5.md` | Search Service boundary confirm ki: full-text search Product Service ka scope nahi |
| `TaskImplementation/Product Service/task6.md` | Inventory availability change ko search event trigger me include kiya |
| `TaskImplementation/Platform Foundation/task3.md` | Shared event envelope and publisher interface reuse kiya |
| `TaskImplementation/Platform Foundation/task4.md` | RabbitMQ default and Kafka alternate message queue decision reuse kiya |
| `api/master-api.json` | ProductService and SearchService contracts verify kiye |

---

## 🧱 Task Boundary

### ✅ Task 7 me kya design/document kiya gaya

- Product Service se search-related product events publish karna
- `product.events` topic/queue ka usage
- Event envelope standard
- Product event payload schema
- Product write workflows me event trigger points
- MongoDB-backed transactional outbox pattern
- Outbox relay worker design
- RabbitMQ default publisher adapter
- Kafka alternate strategy
- Idempotent event publishing rules
- Retry, DLQ, and backoff behavior
- Event versioning and compatibility rules
- Observability: logs, metrics, traces
- Testing strategy for event publishing
- Mermaid architecture, sequence, and state/flow diagrams
- Beginner-friendly Go code examples

### ❌ Task 7 me kya implement nahi kiya gaya

| Item | Reason |
|---|---|
| Search Service consumer | Search Service Task 3 ka scope hai |
| Typesense collection/schema creation | Search Service Task 1/2 ka scope hai |
| Full catalog reindex job | Search Service Task 8 ka scope hai |
| Public search API | Search Service Task 4 ka scope hai |
| Product read APIs | Product Service Task 5 ka scope tha |
| Inventory reserve/release/commit logic | Product Service Task 6 ka scope tha |
| Media metadata/CDN fields | Product Service Task 8 ka scope hai |
| Recommendation event processing | Recommendation Service ka scope hai |

> 🔴 **Golden rule:** Task 7 ka kaam Product Service ke andar event publish mechanism banana hai. Search index update ka actual consumer/indexer yahan implement nahi karna.

---

## 🏁 Final Event Decision

Product Service canonical product data own karega. Search Service Typesense me searchable copy rakhega. Dono ko loosely coupled rakhne ke liye Product Service synchronous gRPC call nahi karega; event publish karega.

| Decision | Value |
|---|---|
| Communication style | Async event |
| Topic / queue | `product.events` |
| Default local broker | RabbitMQ |
| Production-capable alternate | Kafka |
| Reliability pattern | MongoDB transactional outbox + relay worker |
| Event format | Shared event envelope JSON |
| Consumer expectation | Idempotent handling by `event_id` |

### Why async event?

| Synchronous Search update | Async event update |
|---|---|
| Product write slow ho sakti hai | Product write fast rahegi |
| Search down ho to product publish fail ho sakta hai | Search down ho to event retry hoga |
| Services tightly coupled ho jayengi | Product and Search loosely coupled rahenge |
| Scale karna harder | Queue consumers independently scale honge |

---

## 🧩 High-Level Architecture

```mermaid
flowchart TB
    Seller[Seller Dashboard] --> Gateway[API Gateway]
    Gateway --> Product[Product Service]

    Product --> Usecase[Create / Update / Publish / Unpublish Usecases]
    Usecase --> Mongo[(MongoDB product_db)]
    Mongo --> Products[(products)]
    Mongo --> Outbox[(product_event_outbox)]

    Product --> Relay[Outbox Relay Worker]
    Relay --> MQ[product.events<br/>RabbitMQ or Kafka]
    MQ --> Search[Search Service Consumer]
    Search --> Typesense[(Typesense products index)]

    MQ -. future consumer .-> Recommendation[Recommendation Service]
```

**Explanation:**  
Seller product change Gateway ke through Product Service tak aata hai. Product Service pehle MongoDB me product change save karta hai. Same reliable workflow me ek outbox event bhi save hota hai. Relay worker outbox se unpublished events read karta hai aur `product.events` queue/topic pe publish karta hai. Search Service baad me event consume karke Typesense index update karega.

---

## 🔄 Event Flow Diagram

```mermaid
sequenceDiagram
    participant Seller as Seller Dashboard
    participant GW as API Gateway
    participant Product as Product Service
    participant Mongo as MongoDB
    participant Relay as Outbox Relay
    participant MQ as product.events
    participant Search as Search Service
    participant TS as Typesense

    Seller->>GW: PATCH /api/v1/seller/products/{id}
    GW->>Product: gRPC UpdateProduct
    Product->>Mongo: Update products document
    Product->>Mongo: Insert outbox event
    Product-->>GW: Product response
    Relay->>Mongo: Poll unpublished outbox events
    Relay->>MQ: Publish ProductUpdated
    MQ-->>Relay: Ack
    Relay->>Mongo: Mark event as published
    Search->>MQ: Consume ProductUpdated
    Search->>TS: Upsert or delete search document
```

**How this part was built:**  
Product write response ko queue availability se block nahi kiya gaya. Outbox event DB me safe rahega. Agar RabbitMQ/Kafka temporarily down ho, relay retry karega.

---

## 📨 Event Types and Triggers

| Event Type | Trigger | Search Action | Publish Condition |
|---|---|---|---|
| `ProductCreated` | Seller product draft create karta hai | Usually `DELETE` / ignore | Always publish for audit/projections, Search indexes only if public |
| `ProductUpdated` | Product fields, category, brand, attributes, images, price change | `UPSERT` if published, else `DELETE` | Product update success ke baad |
| `ProductPublished` | Product public visible hota hai | `UPSERT` | Publish transition success ke baad |
| `ProductUnpublished` | Product public se hide hota hai | `DELETE` | Unpublish transition success ke baad |
| `ProductInventoryChanged` | Stock availability changes after Task 6 operations | `UPSERT` if published, update `in_stock` | Availability value changed ho |

### Search action rule

```text
if product.status == "published" and product.is_active == true:
    search_action = "UPSERT"
else:
    search_action = "DELETE"
```

> 🟡 **Note:** Draft product ke liye event publish ho sakta hai, but Search Service usko public index me add nahi karega. Payload me `search_action` clearly diya jayega.

---

## 📦 Product Search Event Payload

Event payload Search Service ko enough data dega taaki wo index update kar sake. Lekin source of truth Product Service hi rahega.

```json
{
  "product_id": "prod_123",
  "seller_id": "seller_456",
  "title": "Nike Running Shoes",
  "description": "Lightweight running shoes",
  "brand": "Nike",
  "category_id": "cat_shoes",
  "category_path": ["Fashion", "Footwear", "Running Shoes"],
  "status": "published",
  "search_action": "UPSERT",
  "price": {
    "amount": 2499,
    "currency": "INR"
  },
  "rating": 4.5,
  "popularity_score": 87,
  "in_stock": true,
  "image_url": "https://cdn.example.com/products/prod_123/main.jpg",
  "attributes": {
    "color": "black",
    "size": "9"
  },
  "updated_at": "2026-05-23T10:30:00Z"
}
```

### Field mapping to Typesense

| Event Payload Field | Typesense Field | Purpose |
|---|---|---|
| `product_id` | `id` | Stable search document id |
| `title` | `title` | Searchable title |
| `description` | `description` | Searchable description |
| `brand` | `brand` | Brand facet |
| `category_id`, `category_path` | `category_ids` | Category filters |
| `seller_id` | `seller_id` | Seller/store filter |
| `price.amount` | `price` | Price filter and sort |
| `rating` | `rating` | Rating filter and sort |
| `popularity_score` | `popularity_score` | Ranking signal |
| `in_stock` | `in_stock` | Stock filter |
| `updated_at` | `created_at` / update metadata | Freshness and debugging |

> 🔵 **Important:** Search Service may still call `ProductService.BatchGetProducts` for canonical hydration/verification. Event payload indexing ko fast banata hai, canonical truth replace nahi karta.

---

## ✉️ Standard Event Envelope

Platform Foundation me shared envelope already define kiya gaya hai. Product Service Task 7 usi pattern ko use karega.

```json
{
  "event_id": "evt_7b7f4ef2",
  "event_type": "ProductUpdated",
  "version": 1,
  "source": "product-service",
  "request_id": "req_abc",
  "trace_id": "trace_xyz",
  "occurred_at": "2026-05-23T10:30:00Z",
  "payload": {
    "product_id": "prod_123",
    "search_action": "UPSERT"
  }
}
```

### Envelope fields

| Field | Required | Why needed |
|---|---|---|
| `event_id` | Yes | Consumer idempotency ke liye |
| `event_type` | Yes | Consumer ko route/action decide karne ke liye |
| `version` | Yes | Future schema evolution ke liye |
| `source` | Yes | Debugging and ownership |
| `request_id` | Yes | Logs correlation |
| `trace_id` | Yes | Distributed tracing |
| `occurred_at` | Yes | Ordering/debugging |
| `payload` | Yes | Actual product change data |

---

## 🗃️ Supporting Outbox Collection

Task 7 reliable publish ke liye Product Service DB me ek supporting collection use karega:

```text
product_event_outbox
```

### Example document

```json
{
  "_id": "evt_7b7f4ef2",
  "topic": "product.events",
  "event_type": "ProductUpdated",
  "version": 1,
  "source": "product-service",
  "request_id": "req_abc",
  "trace_id": "trace_xyz",
  "payload": {
    "product_id": "prod_123",
    "search_action": "UPSERT"
  },
  "status": "pending",
  "attempts": 0,
  "next_attempt_at": "2026-05-23T10:30:00Z",
  "occurred_at": "2026-05-23T10:30:00Z",
  "published_at": null,
  "last_error": null
}
```

### Indexes

```javascript
db.product_event_outbox.createIndex({ status: 1, next_attempt_at: 1, occurred_at: 1 })
db.product_event_outbox.createIndex({ event_type: 1, occurred_at: -1 })
db.product_event_outbox.createIndex({ published_at: 1 }, { expireAfterSeconds: 2592000 })
```

### Why outbox?

| Problem | Outbox solution |
|---|---|
| Product DB update success, queue publish fail | Event DB me pending rahega and retry hoga |
| Duplicate publish attempts | Same `event_id` se consumer idempotent rahega |
| Broker temporarily down | Relay backoff ke saath retry karega |
| Debugging hard | Outbox me status, attempts, last_error visible honge |

> 🟡 **MongoDB note:** Transactional outbox best tab hota hai jab MongoDB replica set transactions available hon. Local simple setup me transaction unavailable ho to product write ke immediately baad outbox insert retry-safe code se karna hoga, aur monitoring lagani hogi.

---

## 🧪 Outbox State Machine

```mermaid
stateDiagram-v2
    [*] --> pending: Product write creates event
    pending --> publishing: Relay picks event
    publishing --> published: Broker ack received
    publishing --> pending: Publish failed, attempts < max
    pending --> dead_lettered: attempts >= max
    published --> [*]
    dead_lettered --> pending: Manual replay after fix
```

### Status meaning

| Status | Meaning |
|---|---|
| `pending` | Event publish ke liye ready hai |
| `publishing` | Relay currently publish try kar raha hai |
| `published` | Broker ack mil gaya |
| `dead_lettered` | Max retry fail ho chuka hai, manual/admin replay chahiye |

---

## 🪜 Step-by-Step Implementation

## Step 1: Task requirement identify kiya

`docs/01-micro-tasks.md` me Product Service Task 7 ye define hai:

```text
Publish search events:
Product changes ko Search Service me index karne ke liye events publish karo.
Search data async sync hoga.
```

**How this part was built:**  
Task source se clear hua ki Product Service ko product write ke baad events publish karne hain. Isliye guide ka focus event producer, event shape, outbox reliability, retry, and observability pe rakha gaya. Search Service indexing code intentionally out of scope rakha gaya.

---

## Step 2: Event trigger points decide kiye

Product Service ke existing/future write workflows me event trigger add honge:

| Usecase | Trigger point | Event |
|---|---|---|
| `CreateProduct` | Product insert success ke baad | `ProductCreated` |
| `UpdateProduct` | Product update success ke baad | `ProductUpdated` |
| `PublishProduct` | Status `published` hone ke baad | `ProductPublished` |
| `UnpublishProduct` | Status `unpublished` hone ke baad | `ProductUnpublished` |
| `CommitInventory` / stock update | Availability changed hone ke baad | `ProductInventoryChanged` |

**How this part was built:**  
Task 4 already create/update/publish/unpublish event names mention karta hai. Task 6 inventory operations se `in_stock` search facet affect hota hai, isliye availability change event include kiya gaya.

---

## Step 3: Future folder structure align ki

Actual Product Service implementation ke liye recommended structure:

```text
backend/
└── services/
    └── product-service/
        ├── cmd/
        │   └── server/
        │       └── main.go
        └── internal/
            ├── config/
            │   └── config.go
            ├── domain/
            │   ├── product.go
            │   ├── product_event.go
            │   └── search_document.go
            ├── usecase/
            │   ├── create_product.go
            │   ├── update_product.go
            │   ├── publish_product.go
            │   └── unpublish_product.go
            ├── repository/
            │   ├── product_repository.go
            │   ├── mongo_product_repository.go
            │   ├── outbox_repository.go
            │   └── mongo_outbox_repository.go
            ├── events/
            │   ├── envelope.go
            │   ├── product_events.go
            │   ├── publisher.go
            │   ├── rabbitmq_publisher.go
            │   └── outbox_relay.go
            ├── mapper/
            │   └── search_event_mapper.go
            └── transport/
                └── grpc/
                    ├── server.go
                    └── product_handler.go
```

### Folder responsibility

| Folder | Responsibility |
|---|---|
| `domain` | Product event types and search event payload models |
| `usecase` | Product write workflows me outbox event create karna |
| `repository` | Product and outbox MongoDB persistence |
| `events` | Broker publisher and relay worker |
| `mapper` | Product document ko search event payload me convert karna |
| `transport/grpc` | Existing gRPC handlers, event publishing directly yahan nahi |

**How this part was built:**  
`docs/03-folder-structure.md` Product Service me `internal/events` pattern already define karta hai. Task 7 me wahi folder actual event publisher, adapter, and relay worker ke liye use hota hai.

---

## Step 4: Product event domain model define kiya

```go
package domain

import "time"

type ProductEventType string

const (
	ProductEventCreated          ProductEventType = "ProductCreated"
	ProductEventUpdated          ProductEventType = "ProductUpdated"
	ProductEventPublished        ProductEventType = "ProductPublished"
	ProductEventUnpublished      ProductEventType = "ProductUnpublished"
	ProductEventInventoryChanged ProductEventType = "ProductInventoryChanged"
)

type SearchAction string

const (
	SearchActionUpsert SearchAction = "UPSERT"
	SearchActionDelete SearchAction = "DELETE"
)

type Money struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

type ProductSearchEventPayload struct {
	ProductID       string         `json:"product_id"`
	SellerID        string         `json:"seller_id"`
	Title           string         `json:"title"`
	Description     string         `json:"description"`
	Brand           string         `json:"brand"`
	CategoryID      string         `json:"category_id"`
	CategoryPath    []string       `json:"category_path"`
	Status          string         `json:"status"`
	SearchAction    SearchAction   `json:"search_action"`
	Price           Money          `json:"price"`
	Rating          float64        `json:"rating"`
	PopularityScore int32          `json:"popularity_score"`
	InStock         bool           `json:"in_stock"`
	ImageURL        string         `json:"image_url"`
	Attributes      map[string]any `json:"attributes"`
	UpdatedAt       time.Time      `json:"updated_at"`
}
```

**Explanation:**  
Payload me wo fields rakhe gaye jo Typesense search schema me useful hain: title, brand, category, price, rating, popularity, stock, and updated time. `SearchAction` se consumer ko clear instruction milti hai ki document upsert karna hai ya delete.

---

## Step 5: Shared event envelope use kiya

```go
package events

import "time"

type Envelope[T any] struct {
	EventID    string    `json:"event_id"`
	EventType  string    `json:"event_type"`
	Version    int       `json:"version"`
	Source     string    `json:"source"`
	RequestID  string    `json:"request_id"`
	TraceID    string    `json:"trace_id"`
	OccurredAt time.Time `json:"occurred_at"`
	Payload    T         `json:"payload"`
}
```

**How this part was built:**  
Platform Foundation Task 3 me shared event envelope already define hai. Product Service Task 7 me custom format invent nahi kiya gaya, same envelope reuse kiya gaya.

---

## Step 6: Product se search payload mapper banaya

```go
package mapper

import "time"

func ToProductSearchPayload(product Product) ProductSearchEventPayload {
	action := SearchActionDelete
	if product.Status == "published" && product.IsActive {
		action = SearchActionUpsert
	}

	return ProductSearchEventPayload{
		ProductID:       product.ID,
		SellerID:        product.SellerID,
		Title:           product.Title,
		Description:     product.Description,
		Brand:           product.Brand,
		CategoryID:      product.CategoryID,
		CategoryPath:    product.CategoryPath,
		Status:          product.Status,
		SearchAction:    action,
		Price:           product.PrimaryPrice,
		Rating:          product.Rating,
		PopularityScore: product.PopularityScore,
		InStock:         product.AvailableQuantity() > 0,
		ImageURL:        product.PrimaryImageURL(),
		Attributes:      product.SearchableAttributes(),
		UpdatedAt:       time.Now().UTC(),
	}
}
```

**Explanation:**  
Mapper product document ko search-friendly event payload me convert karta hai. Ye mapping separate rakhne se usecase clean rehta hai aur future me Typesense schema change ho to mapping ek jagah update hogi.

---

## Step 7: Outbox repository define kiya

```go
package repository

import (
	"context"
	"time"
)

type OutboxEvent struct {
	ID            string
	Topic         string
	EventType     string
	Version       int
	Source        string
	RequestID     string
	TraceID       string
	Payload       any
	Status        string
	Attempts      int
	NextAttemptAt time.Time
	OccurredAt    time.Time
	PublishedAt   *time.Time
	LastError      string
}

type OutboxRepository interface {
	Insert(ctx context.Context, event OutboxEvent) error
	ListPending(ctx context.Context, limit int, now time.Time) ([]OutboxEvent, error)
	MarkPublishing(ctx context.Context, eventID string) error
	MarkPublished(ctx context.Context, eventID string, publishedAt time.Time) error
	MarkFailed(ctx context.Context, eventID string, nextAttemptAt time.Time, lastError string) error
	MarkDeadLettered(ctx context.Context, eventID string, lastError string) error
}
```

**Explanation:**  
Outbox repository event publish state manage karta hai. Product usecase sirf event insert karega. Relay worker pending events uthakar broker me publish karega.

---

## Step 8: Product write ke saath outbox event save kiya

```go
func (uc *UpdateProductUsecase) Execute(ctx context.Context, cmd UpdateProductCommand) (*Product, error) {
	product, err := uc.products.Update(ctx, cmd.ProductID, cmd.Patch)
	if err != nil {
		return nil, err
	}

	payload := mapper.ToProductSearchPayload(*product)

	event := events.Envelope[ProductSearchEventPayload]{
		EventID:    uc.ids.NewEventID(),
		EventType:  string(ProductEventUpdated),
		Version:    1,
		Source:     "product-service",
		RequestID:  RequestIDFromContext(ctx),
		TraceID:    TraceIDFromContext(ctx),
		OccurredAt: uc.clock.Now().UTC(),
		Payload:    payload,
	}

	if err := uc.outbox.Insert(ctx, ToOutboxEvent("product.events", event)); err != nil {
		return nil, err
	}

	return product, nil
}
```

**Explanation:**  
Product update ke baad event outbox me store hota hai. Production implementation me product update and outbox insert ko MongoDB transaction me wrap karna best hai, taaki dono together commit hon.

---

## Step 9: RabbitMQ publisher adapter banaya

RabbitMQ local baseline ke liye default broker hai.

```go
package events

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisher struct {
	channel *amqp.Channel
}

func (p *RabbitMQPublisher) Publish(ctx context.Context, topic string, event Envelope[any]) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.channel.PublishWithContext(
		ctx,
		"",
		topic,
		true,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    event.EventID,
			Type:         event.EventType,
			Body:         body,
		},
	)
}
```

**Explanation:**  
`DeliveryMode: amqp.Persistent` event ko broker restart ke against safer banata hai, agar durable queue configured hai. `MessageId` me `event_id` dene se consumer dedupe easy hota hai.

---

## Step 10: Outbox relay worker banaya

```go
func (r *OutboxRelay) RunOnce(ctx context.Context) error {
	events, err := r.outbox.ListPending(ctx, 100, r.clock.Now().UTC())
	if err != nil {
		return err
	}

	for _, item := range events {
		if err := r.outbox.MarkPublishing(ctx, item.ID); err != nil {
			continue
		}

		envelope := item.ToEnvelope()
		err := r.publisher.Publish(ctx, item.Topic, envelope)
		if err == nil {
			_ = r.outbox.MarkPublished(ctx, item.ID, r.clock.Now().UTC())
			continue
		}

		if item.Attempts+1 >= r.maxAttempts {
			_ = r.outbox.MarkDeadLettered(ctx, item.ID, err.Error())
			continue
		}

		nextAttempt := r.backoff.Next(item.Attempts + 1)
		_ = r.outbox.MarkFailed(ctx, item.ID, nextAttempt, err.Error())
	}

	return nil
}
```

**Explanation:**  
Relay worker pending events batch me uthata hai. Publish success pe `published` mark karta hai. Failure pe retry schedule karta hai. Repeated failure pe event `dead_lettered` status me chala jata hai.

---

## Step 11: Retry and DLQ strategy define ki

| Attempt | Delay example |
|---:|---|
| 1 | 5 seconds |
| 2 | 30 seconds |
| 3 | 2 minutes |
| 4 | 10 minutes |
| 5 | Dead-letter |

### Failure handling rules

| Failure | Action |
|---|---|
| Broker connection down | Retry with backoff |
| Publish timeout | Retry with same `event_id` |
| JSON marshal error | Dead-letter, developer fix needed |
| Invalid payload | Dead-letter, schema bug fix needed |
| Duplicate event at consumer | Consumer ignores using `event_id` |

**How this part was built:**  
`docs/11-devops-external-services.md` explicitly says consumers should be idempotent, use DLQ, retry with backoff, and schema versioning. Task 7 producer follows same reliability rules.

---

## Step 12: Idempotency rules add kiye

### Producer side

| Rule | Reason |
|---|---|
| Har event ka globally unique `event_id` | Duplicate publish detect karne ke liye |
| Same outbox event retry me same `event_id` reuse | Consumer duplicate ko ignore kar sake |
| Product write retries me duplicate event avoid karna | Idempotency key supported commands me useful |
| Outbox status atomic update | Multiple relay workers same event publish na karein |

### Consumer side expectation

Search Service apni side processed event ids store kare:

```text
processed_product_events(event_id, processed_at)
```

> 🔵 **Scope note:** Consumer table/collection implement karna Search Service task ka kaam hai. Product Service sirf stable `event_id` provide karega.

---

## Step 13: Ordering strategy define ki

Product updates out of order arrive ho sakte hain. Isliye payload me `updated_at` and envelope me `occurred_at` required hai.

```text
Search consumer rule:
if incoming.updated_at < existing_index.updated_at:
    ignore stale event
else:
    apply event
```

### Why ordering important?

Example:

```text
10:00 ProductUpdated title="Old Title"
10:01 ProductUpdated title="New Title"
```

Agar queue retry ke wajah se 10:00 event baad me consume hua, Search Service stale event ignore karega.

---

## Step 14: Config values define kiye

```env
PRODUCT_EVENTS_ENABLED=true
PRODUCT_EVENTS_TOPIC=product.events
PRODUCT_EVENT_BROKER=rabbitmq
RABBITMQ_URL=amqp://ecommerce:ecommerce@localhost:5672/
PRODUCT_OUTBOX_POLL_INTERVAL_MS=1000
PRODUCT_OUTBOX_BATCH_SIZE=100
PRODUCT_OUTBOX_MAX_ATTEMPTS=5
PRODUCT_OUTBOX_WORKER_ENABLED=true
```

### Config explanation

| Config | Purpose |
|---|---|
| `PRODUCT_EVENTS_ENABLED` | Emergency off switch |
| `PRODUCT_EVENTS_TOPIC` | Topic/queue name |
| `PRODUCT_EVENT_BROKER` | `rabbitmq` or `kafka` |
| `RABBITMQ_URL` | RabbitMQ connection URL |
| `PRODUCT_OUTBOX_POLL_INTERVAL_MS` | Relay polling frequency |
| `PRODUCT_OUTBOX_BATCH_SIZE` | Ek cycle me kitne events publish karne hain |
| `PRODUCT_OUTBOX_MAX_ATTEMPTS` | Dead-letter se pehle retry count |
| `PRODUCT_OUTBOX_WORKER_ENABLED` | Worker ko same service me run karna hai ya separate process |

---

## Step 15: Search Service ke saath contract boundary define ki

Product Service ye guarantee karega:

- Event JSON envelope valid hoga.
- `event_id` stable and unique hoga.
- `event_type` supported list me hoga.
- `version` backward compatible hoga.
- `payload.product_id` always present hoga.
- `payload.search_action` always `UPSERT` ya `DELETE` hoga.
- Published product ke liye searchable fields populated honge.

Search Service ye guarantee karega:

- `event_id` idempotently process karega.
- `UPSERT` pe Typesense document create/update karega.
- `DELETE` pe Typesense document remove karega.
- Stale `updated_at` events ignore karega.
- Failed events retry/DLQ me handle karega.

> 🟡 **Boundary:** Product Service Typesense client import nahi karega. Typesense access sirf Search Service ka responsibility hai.

---

## Step 16: Observability add ki

### Logs

```json
{
  "level": "info",
  "service": "product-service",
  "message": "product event queued",
  "event_id": "evt_7b7f4ef2",
  "event_type": "ProductUpdated",
  "product_id": "prod_123",
  "topic": "product.events",
  "request_id": "req_abc",
  "trace_id": "trace_xyz"
}
```

### Metrics

| Metric | Type | Purpose |
|---|---|---|
| `product_events_outbox_pending_total` | gauge | Pending outbox backlog |
| `product_events_published_total` | counter | Successfully published events |
| `product_events_publish_failed_total` | counter | Publish failures |
| `product_events_dead_lettered_total` | counter | Events needing manual intervention |
| `product_events_publish_duration_ms` | histogram | Broker publish latency |
| `product_events_oldest_pending_age_seconds` | gauge | Queue lag detection |

### Tracing

```text
ProductService.UpdateProduct span
├── Mongo.UpdateProduct span
├── Mongo.InsertOutboxEvent span
└── OutboxRelay.PublishProductEvent span
    └── RabbitMQ.Publish span
```

**Explanation:**  
Event publish pipeline async hai, isliye metrics and logs very important hain. Agar Search results stale lag rahe hain, pending outbox count and oldest pending age se root cause quickly mil sakta hai.

---

## Step 17: Tests define kiye

### Unit tests

| Test | Expected result |
|---|---|
| Published product maps to `UPSERT` | Search action `UPSERT` |
| Draft product maps to `DELETE` | Search action `DELETE` |
| Update product creates outbox event | Event type `ProductUpdated` |
| Publish product creates outbox event | Event type `ProductPublished` |
| Unpublish product creates delete action | Search action `DELETE` |
| Outbox retry increments attempts | `attempts + 1` and next attempt set |
| Max attempts dead-letters event | Status `dead_lettered` |

### Integration tests

| Test | Expected result |
|---|---|
| Product update + outbox insert transaction | Both commit together |
| RabbitMQ publish success | Outbox status `published` |
| RabbitMQ unavailable | Outbox remains retryable |
| Duplicate relay workers | One worker owns event publish attempt |
| Event JSON schema | Required fields present |

### Contract tests

| Contract | Check |
|---|---|
| Envelope version | `version=1` supported |
| Event type enum | Search consumer supported type |
| Payload required fields | `product_id`, `search_action`, `updated_at` |
| Delete action payload | Enough info to delete by `product_id` |

---

## 📚 External Libraries / Tools Used

> 🔵 **Note:** Current task documentation file me koi package install nahi kiya gaya. Neeche actual implementation ke time required/recommended tools listed hain.

### 1. RabbitMQ

| Field | Detail |
|---|---|
| What | Message broker |
| Why used | Local MVP me simple queue/routing, async Product -> Search events ke liye |
| Topic/Queue | `product.events` |

Install/use with Docker Compose:

```yaml
rabbitmq:
  image: rabbitmq:3-management
  ports:
    - "5672:5672"
    - "15672:15672"
  environment:
    RABBITMQ_DEFAULT_USER: ecommerce
    RABBITMQ_DEFAULT_PASS: ecommerce
```

Go client install:

```bash
go get github.com/rabbitmq/amqp091-go
```

Basic use:

```go
conn, err := amqp.Dial("amqp://ecommerce:ecommerce@localhost:5672/")
if err != nil {
	return err
}
defer conn.Close()
```

### 2. Kafka optional alternate

| Field | Detail |
|---|---|
| What | Event streaming platform |
| Why used | High-throughput product/session/recommendation streams ke liye future production option |
| Topic | `product.events` |

Go client install:

```bash
go get github.com/segmentio/kafka-go
```

Basic producer idea:

```go
writer := kafka.NewWriter(kafka.WriterConfig{
	Brokers: []string{"localhost:9092"},
	Topic:   "product.events",
})
```

### 3. MongoDB Go Driver

| Field | Detail |
|---|---|
| What | Official MongoDB driver for Go |
| Why used | `products` and `product_event_outbox` collections ke liye |
| Usage | Product update, outbox insert, transaction, pending event query |

Install:

```bash
go get go.mongodb.org/mongo-driver/mongo
```

Transaction idea:

```go
session, err := client.StartSession()
if err != nil {
	return err
}
defer session.EndSession(ctx)
```

### 4. Google UUID

| Field | Detail |
|---|---|
| What | UUID generator |
| Why used | `event_id` generate karne ke liye |
| Usage | `evt_` prefixed unique id |

Install:

```bash
go get github.com/google/uuid
```

Example:

```go
eventID := "evt_" + uuid.NewString()
```

### 5. Mermaid

| Field | Detail |
|---|---|
| What | Markdown-friendly diagram syntax |
| Why used | Architecture, sequence, and state diagrams readable banane ke liye |
| Install | GitHub/GitLab/Markdown viewers me usually built-in render hota hai |

Usage:

````markdown
```mermaid
flowchart LR
    Product --> MQ
    MQ --> Search
```
````

---

## 🧾 Event Examples

### Product published event

```json
{
  "event_id": "evt_prod_published_001",
  "event_type": "ProductPublished",
  "version": 1,
  "source": "product-service",
  "request_id": "req_001",
  "trace_id": "trace_001",
  "occurred_at": "2026-05-23T10:30:00Z",
  "payload": {
    "product_id": "prod_123",
    "seller_id": "seller_456",
    "title": "Nike Running Shoes",
    "description": "Lightweight running shoes",
    "brand": "Nike",
    "category_id": "cat_running_shoes",
    "category_path": ["Fashion", "Footwear", "Running Shoes"],
    "status": "published",
    "search_action": "UPSERT",
    "price": {
      "amount": 2499,
      "currency": "INR"
    },
    "rating": 4.5,
    "popularity_score": 87,
    "in_stock": true,
    "image_url": "https://cdn.example.com/products/prod_123/main.jpg",
    "attributes": {
      "color": "black",
      "size": "9"
    },
    "updated_at": "2026-05-23T10:30:00Z"
  }
}
```

### Product unpublished event

```json
{
  "event_id": "evt_prod_unpublished_001",
  "event_type": "ProductUnpublished",
  "version": 1,
  "source": "product-service",
  "request_id": "req_002",
  "trace_id": "trace_002",
  "occurred_at": "2026-05-23T11:00:00Z",
  "payload": {
    "product_id": "prod_123",
    "seller_id": "seller_456",
    "title": "Nike Running Shoes",
    "status": "unpublished",
    "search_action": "DELETE",
    "updated_at": "2026-05-23T11:00:00Z"
  }
}
```

---

## 🧯 Failure Scenarios

| Scenario | Expected behavior |
|---|---|
| Product update fails | No outbox event created |
| Product update succeeds, broker down | Outbox event remains pending |
| Relay publishes but mark published fails | Event may republish, same `event_id` dedupe handles it |
| Search Service down | Broker retains event or consumer retries |
| Event schema bug | Event dead-lettered, alert triggered |
| Old event arrives late | Search Service ignores using `updated_at` |

---

## 🔐 Security and Data Rules

| Rule | Reason |
|---|---|
| Events must not contain seller private data | Search consumers do not need PII |
| Events must not contain internal cost price | Public search only needs sale/listing price |
| Broker credentials from env/secret | Secrets repo me hardcode nahi honge |
| Internal topic only | Browser/frontend direct event publish nahi karega |
| Payload schema reviewed | Accidental sensitive field leak avoid hota hai |

### Safe payload

```text
product_id, seller_id, title, description, brand, category, price, stock boolean, public image URL
```

### Unsafe payload

```text
seller bank details, cost price, moderation notes, private audit comments, internal fraud score
```

---

## 🚀 Beginner-Friendly Build Order

1. Event payload struct banao.
2. Product -> search payload mapper banao.
3. Outbox collection model banao.
4. Product write usecases me outbox insert add karo.
5. RabbitMQ publisher adapter banao.
6. Outbox relay worker banao.
7. Retry/backoff and dead-letter status add karo.
8. Logs/metrics/traces add karo.
9. Unit tests likho.
10. RabbitMQ integration test run karo.
11. Search Service team ko event contract share karo.

---

## ✅ Acceptance Checklist

| Check | Status |
|---|---|
| Scope sirf Product Service Task 7 tak limited hai | ✅ |
| `TaskImplementation/Product Service/task7.md` created | ✅ |
| Hinglish step-by-step implementation included | ✅ |
| Product event trigger points documented | ✅ |
| Event envelope and payload examples added | ✅ |
| Outbox reliability pattern documented | ✅ |
| RabbitMQ default and Kafka alternate explained | ✅ |
| External libraries/tools install/use documented | ✅ |
| Clean folder structure included | ✅ |
| Mermaid diagrams included | ✅ |
| Search Service implementation out of scope rakha | ✅ |
| Task 8 media metadata untouched | ✅ |

---

## 🚫 Out of Scope for Task 7

```text
Task 7 = Product Service publishes search events

Not Task 7:
- Search Service consumer/indexer code
- Typesense schema migration
- Full catalog reindex job
- Search REST API
- Recommendation ranking
- CDN/media metadata
- Product CRUD/read/inventory implementation changes
```

> 🔴 **Reason:** User ne specifically Product Service Task 7 ka implementation guide generate karne ko bola hai. Isliye yahan sirf event publishing design documented hai.

---

## ✅ Final Task 7 Standard

Product Service Task 7 complete hai as a **documentation-only implementation guide**. Is design me Product Service product write ke baad reliable outbox event create karega, relay worker event ko `product.events` queue/topic pe publish karega, aur Search Service async consume karke Typesense index update karega. RabbitMQ local default hai, Kafka production/high-throughput alternate hai, and event envelope/idempotency/retry/DLQ rules pipeline ko production-ready direction dete hain.
