# 🔎 Search Service - Task 3: Product Indexer

![Task](https://img.shields.io/badge/Task-Search%20Service%20%233-blue)
![Priority](https://img.shields.io/badge/Priority-P0-red)
![Dependency](https://img.shields.io/badge/Dependency-Product%20Events-orange)
![Status](https://img.shields.io/badge/Status-Documented-success)
![Engine](https://img.shields.io/badge/Search-Typesense%2027.1-2ea44f)
![Queue](https://img.shields.io/badge/Queue-RabbitMQ%20or%20Kafka-8A2BE2)
![Guide](https://img.shields.io/badge/Guide-Hinglish-yellow)

## 📌 Task Summary

| Field | Detail |
|---|---|
| Service | Search Service |
| Task | Task 3 - Product indexer |
| Source | `docs/01-micro-tasks.md` -> `Search Service (Typesense)` -> Task 3 |
| Priority | `P0` |
| Dependency | Product events |
| Main Goal | Product events consume karke Typesense `products` collection me document upsert/delete karna |
| Output Type | Documentation-only implementation guide |

> **Simple Hinglish goal:** Product Service jab product publish/update/delete/inventory/price event bhejega, Search Service ka indexer us event ko consume karega aur Typesense search index ko update karega. Isse frontend search fast rahegi aur Product Service source of truth bhi bana rahega.

---

## ✅ Final Output Created

```text
TaskImplementation/
└── Search Service/
    ├── task1.md
    ├── task2.md
    └── task3.md
```

### Why this structure?

| Path | Purpose |
|---|---|
| `TaskImplementation/` | Saare task-wise implementation guides ka central folder |
| `TaskImplementation/Search Service/` | Search Service ke implementation guides ka group |
| `task3.md` | Sirf Search Service - Task 3 ka detailed guide |

> 🟢 **Boundary:** Is request me sirf `task3.md` guide create ki gayi hai. Actual backend Go files, queue bindings, Docker files, ya Kubernetes manifests add nahi kiye gaye. Guide me implementation snippets diye gaye hain taaki Task 3 ko safely build kiya ja sake.

---

## 🧭 Requirement Sources Studied

| Document | Kya use kiya gaya |
|---|---|
| `docs/01-micro-tasks.md` | Task 3 ka exact scope: `Product indexer` |
| `docs/04-microservice-design.md` | Search Service responsibilities: product index updates, Typesense, full reindex later |
| `docs/05-database-design.md` | Typesense `products` collection fields |
| `docs/02-system-architecture.md` | Product Service -> MQ -> Search Service architecture |
| `docs/03-folder-structure.md` | Expected `search-service` folder layout |
| `docs/11-devops-external-services.md` | Queue topics, event envelope, consumer rules |
| `TaskImplementation/Search Service/task1.md` | Search schema and searchable document contract |
| `TaskImplementation/Search Service/task2.md` | Typesense runtime and Search Service connection config |
| `TaskImplementation/Platform Foundation/task3.md` | Shared event envelope, idempotency, logging, health patterns |
| `TaskImplementation/Platform Foundation/task4.md` | Local RabbitMQ/Kafka queue names and event stack |

---

## 🧱 Scope of Task 3

### ✅ Included

- Product event consumer design
- Event payload contract for search indexing
- Product event -> Typesense document mapper
- Upsert and delete logic for `products` collection
- Idempotent event handling
- Retry, dead-letter queue, and ack/nack strategy
- Typesense repository interface
- Local config variables for queue and Typesense
- Logs, metrics, trace propagation points
- Unit/integration test plan
- Mermaid diagrams and Go code examples

### 🚫 Not Included

| Not Included | Reason |
|---|---|
| Product Service event publisher implementation | Product Service Task 7 ka scope |
| Typesense runtime setup | Already covered in Search Service Task 2 |
| Search API endpoint | Search Service Task 4 ka scope |
| Autocomplete API | Search Service Task 5 ka scope |
| Synonym CRUD | Search Service Task 6 ka scope |
| Zero-result analytics | Search Service Task 7 ka scope |
| Full catalog reindex command/job | Search Service Task 8 ka scope |
| Frontend search UI | Frontend Product Browsing/Search tasks ka scope |

---

## 🧩 High-Level Architecture

```mermaid
flowchart LR
    PRODUCT[Product Service<br/>Source of Truth]
    MQ[(Kafka/RabbitMQ<br/>product.events)]
    CONSUMER[Search Service<br/>Product Consumer]
    INDEXER[Product Indexer<br/>Map + Validate + Route]
    TYPESENSE[(Typesense<br/>products collection)]

    PRODUCT -->|ProductPublished/ProductUpdated/ProductDeleted| MQ
    MQ --> CONSUMER
    CONSUMER --> INDEXER
    INDEXER -->|upsert/delete| TYPESENSE
```

### Hinglish explanation

- **Product Service** canonical product data own karta hai.
- Product change hone par Product Service `product.events` queue/topic me event publish karega.
- **Search Service consumer** event receive karega.
- **Product Indexer** decide karega: document upsert karna hai ya delete.
- **Typesense** me sirf searchable projection store hogi, product ka source of truth nahi.

---

## 🧠 Key Decisions

| Decision | Value | Why |
|---|---|---|
| Queue/topic | `product.events` | Platform docs me Product -> Search event route defined hai |
| Consumer group/queue | `search-service.product-indexer` | Search Service ka isolated consumer, Recommendation se independent |
| Search collection | `products` | Task 1 schema contract |
| Write actions | `upsert`, `delete` | Published/searchable products index honge; hidden/deleted products remove honge |
| Event style | Full searchable snapshot preferred | Consumer ko Product Service DB/API pe depend nahi rehna padega |
| Delivery model | At-least-once | RabbitMQ/Kafka me duplicates possible hain, isliye idempotency zaruri |
| Failure policy | Retry + DLQ | Temporary Typesense/queue failures recover ho sakte hain |
| Canonical data owner | Product Service | Typesense projection stale ho sakti hai, source of truth Product Service hi rahega |

---

## 🗂️ Recommended Implementation Folder Structure

> Ye future code structure Task 3 implement karne ke liye recommended hai. Is request me sirf `TaskImplementation/Search Service/task3.md` create hua hai.

```text
ecommerce-platform/
├── TaskImplementation/
│   └── Search Service/
│       ├── task1.md
│       ├── task2.md
│       └── task3.md
└── backend/
    └── services/
        └── search-service/
            ├── cmd/
            │   └── server/
            │       └── main.go
            ├── internal/
            │   ├── config/
            │   │   └── config.go
            │   ├── domain/
            │   │   └── product_document.go
            │   ├── events/
            │   │   ├── product_consumer.go
            │   │   └── product_event.go
            │   ├── indexer/
            │   │   ├── product_indexer.go
            │   │   └── mapper.go
            │   ├── repository/
            │   │   └── typesense_repository.go
            │   └── schema/
            │       └── typesense_schema.go
            └── deploy/
```

### Folder responsibility

| Folder/File | Responsibility |
|---|---|
| `cmd/server/main.go` | Config load, logger, queue consumer, Typesense client start karega |
| `internal/config/config.go` | Typesense and queue env config read karega |
| `internal/domain/product_document.go` | Typesense me store hone wala product document model |
| `internal/events/product_event.go` | Product event payload structs |
| `internal/events/product_consumer.go` | RabbitMQ/Kafka messages consume karega |
| `internal/indexer/product_indexer.go` | Event type ke basis pe upsert/delete workflow run karega |
| `internal/indexer/mapper.go` | Product event payload ko Typesense document me convert karega |
| `internal/repository/typesense_repository.go` | Typesense upsert/delete API calls wrap karega |
| `internal/schema/typesense_schema.go` | Task 1 schema ko code form me ensure karega |

---

## 🪜 Step-by-Step Implementation Guide

## Step 1: Task boundary clear karo

Task 3 ka kaam **event-driven indexing** hai.

Iska matlab:

- Product event consume karo.
- Event validate karo.
- Decide karo product searchable hai ya nahi.
- Searchable hai to Typesense me upsert karo.
- Searchable nahi hai to Typesense se delete karo.
- Duplicate events se duplicate side effects avoid karo.
- Failure pe retry/DLQ use karo.

Task 3 me user-facing search endpoint nahi banega. Search API Task 4 me banegi.

---

## Step 2: Product event flow define karo

```mermaid
sequenceDiagram
    participant Product as Product Service
    participant MQ as product.events
    participant Consumer as Search Consumer
    participant Indexer as Product Indexer
    participant TS as Typesense

    Product->>MQ: Publish ProductUpdated event
    MQ->>Consumer: Deliver event
    Consumer->>Consumer: Decode envelope + validate
    Consumer->>Indexer: Handle event
    Indexer->>Indexer: Map payload to search document
    Indexer->>TS: Upsert document
    TS-->>Indexer: OK
    Indexer-->>Consumer: Success
    Consumer-->>MQ: Ack message
```

### How this part was built

- Queue route `product.events` docs se liya gaya.
- `Search Consumer` transport layer hai, jo RabbitMQ/Kafka details handle karega.
- `Product Indexer` business workflow hai, jo event type ke basis pe action choose karega.
- Typesense write repository ke through hoga, taaki tests me fake repository use ho sake.

---

## Step 3: Event envelope standard follow karo

Platform Foundation me shared envelope pattern already define hai. Task 3 usi ko use karega.

```json
{
  "event_id": "evt_01HYSEARCH123",
  "event_type": "ProductUpdated",
  "version": 1,
  "producer": "product-service",
  "request_id": "req_01HYREQ123",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "correlation_id": "prod_123",
  "occurred_at": "2026-05-23T10:30:00Z",
  "payload": {}
}
```

### Required envelope fields

| Field | Required | Why |
|---|---:|---|
| `event_id` | ✅ | Idempotency and debugging |
| `event_type` | ✅ | Upsert/delete route decide karne ke liye |
| `version` | ✅ | Future payload changes handle karne ke liye |
| `producer` | ✅ | Source service identify karne ke liye |
| `request_id` | ✅ | Logs correlate karne ke liye |
| `trace_id` | ✅ | Distributed tracing ke liye |
| `correlation_id` | ✅ | Usually `product_id`, product journey trace karne ke liye |
| `occurred_at` | ✅ | Event timing and lag measure karne ke liye |
| `payload` | ✅ | Search document banane ke liye product data |

> 🟡 **Rule:** Payload me secrets, seller private data, supplier cost, internal moderation notes, ya user PII nahi aani chahiye.

---

## Step 4: Product indexing event types define karo

Task 3 me consumer ye event types handle karega:

| Event Type | Action | Reason |
|---|---|---|
| `ProductPublished` | Upsert | Product public search me aana chahiye |
| `ProductUpdated` | Upsert or Delete | Status ke basis pe searchable/non-searchable decide hoga |
| `ProductPriceChanged` | Upsert | Price filter/sort fresh rehna chahiye |
| `ProductInventoryChanged` | Upsert | `in_stock` filter fresh rehna chahiye |
| `ProductUnpublished` | Delete | Draft/unpublished product search me nahi aana chahiye |
| `ProductDeleted` | Delete | Deleted product search se remove hona chahiye |
| `ProductBlocked` | Delete | Blocked/rejected product public search me nahi dikhna chahiye |

### Hinglish explanation

Har product event ka final goal simple hai:

- Agar product **publicly searchable** hai -> Typesense `products` collection me upsert.
- Agar product **publicly searchable nahi** hai -> Typesense se delete.

---

## Step 5: Product payload contract banao

Best approach ye hai ki Product Service event me complete searchable snapshot bheje. Isse Search Service ko har event pe Product Service API call nahi karni padegi.

```json
{
  "product_id": "prod_123",
  "title": "Nike Running Shoes",
  "description": "Lightweight running shoes for daily training",
  "brand": "Nike",
  "category_ids": ["cat_shoes", "cat_running"],
  "seller_id": "seller_456",
  "price": 2499.0,
  "rating": 4.5,
  "popularity_score": 982,
  "in_stock": true,
  "status": "published",
  "is_deleted": false,
  "updated_at": "2026-05-23T10:29:45Z"
}
```

### Payload field mapping

| Payload Field | Typesense Field | Required | Notes |
|---|---|---:|---|
| `product_id` | `id` | ✅ | Typesense document id |
| `title` | `title` | ✅ | Searchable |
| `description` | `description` | ❌ | Searchable but optional |
| `brand` | `brand` | ❌ | Searchable + facet |
| `category_ids` | `category_ids` | ✅ | Category filters |
| `seller_id` | `seller_id` | ✅ | Seller filter |
| `price` | `price` | ✅ | Price filter/sort |
| `rating` | `rating` | ❌ | Rating sort/filter |
| `popularity_score` | `popularity_score` | ✅ | Default ranking |
| `in_stock` | `in_stock` | ✅ | Stock filter |
| `updated_at` | Internal validation | ✅ | Stale event handling/logging |

> 🔵 **Note:** Typesense schema Task 1 ke fields ko follow karega. `updated_at` ko MVP me log/stale-check metadata ke roop me use kiya ja sakta hai; searchable schema me add karna zaruri nahi.

---

## Step 6: Searchable product rule banao

Search Service ko har event pe decide karna hai ki product index me rehna chahiye ya nahi.

```go
func IsSearchable(p ProductIndexPayload) bool {
    if p.ProductID == "" {
        return false
    }

    if p.IsDeleted {
        return false
    }

    if p.Status != "published" {
        return false
    }

    if p.Title == "" || p.SellerID == "" || len(p.CategoryIDs) == 0 {
        return false
    }

    return true
}
```

### Why ye rule important hai?

| Condition | Reason |
|---|---|
| `is_deleted = true` | Deleted product search me kabhi nahi dikhna chahiye |
| `status != published` | Draft, rejected, blocked, unpublished products hidden rahenge |
| Empty `title` | Search result quality break hogi |
| Empty `seller_id` | Seller filter and ownership trace missing ho jayega |
| Empty `category_ids` | Category browse/filter weak ho jayega |

> 🟢 **Stock rule:** `in_stock=false` product ko delete karna zaruri nahi. Isse document index me reh sakta hai aur Task 4 Search API default filter `in_stock:true` apply kar sakti hai. Agar business rule ho ki out-of-stock products search me kabhi na dikhe, tab indexer delete strategy use kar sakta hai.

---

## Step 7: Domain models define karo

### Product event payload

```go
package events

import "time"

type ProductIndexPayload struct {
    ProductID       string    `json:"product_id"`
    Title           string    `json:"title"`
    Description     string    `json:"description"`
    Brand           string    `json:"brand"`
    CategoryIDs     []string  `json:"category_ids"`
    SellerID        string    `json:"seller_id"`
    Price           float64   `json:"price"`
    Rating          float64   `json:"rating"`
    PopularityScore int32     `json:"popularity_score"`
    InStock         bool      `json:"in_stock"`
    Status          string    `json:"status"`
    IsDeleted       bool      `json:"is_deleted"`
    UpdatedAt       time.Time `json:"updated_at"`
}
```

### Typesense document model

```go
package domain

type ProductDocument struct {
    ID              string   `json:"id"`
    Title           string   `json:"title"`
    Description     string   `json:"description,omitempty"`
    Brand           string   `json:"brand,omitempty"`
    CategoryIDs     []string `json:"category_ids"`
    SellerID        string   `json:"seller_id"`
    Price           float64  `json:"price"`
    Rating          float64  `json:"rating,omitempty"`
    PopularityScore int32    `json:"popularity_score"`
    InStock         bool     `json:"in_stock"`
    CreatedAt       int64    `json:"created_at"`
}
```

### How this part was built

- `ProductIndexPayload` event input model hai.
- `ProductDocument` Typesense output model hai.
- Field names Task 1 schema ke exactly aligned rakhe gaye.
- `json` tags Typesense document keys ke saath match karte hain.

---

## Step 8: Mapper implement karo

Mapper ka kaam payload ko clean Typesense document me convert karna hai.

```go
package indexer

import (
    "strings"
    "time"

    "ecommerce/backend/services/search-service/internal/domain"
    searchevents "ecommerce/backend/services/search-service/internal/events"
)

func MapProductToDocument(payload searchevents.ProductIndexPayload) domain.ProductDocument {
    return domain.ProductDocument{
        ID:              strings.TrimSpace(payload.ProductID),
        Title:           strings.TrimSpace(payload.Title),
        Description:     strings.TrimSpace(payload.Description),
        Brand:           strings.TrimSpace(payload.Brand),
        CategoryIDs:     cleanStringList(payload.CategoryIDs),
        SellerID:        strings.TrimSpace(payload.SellerID),
        Price:           payload.Price,
        Rating:          payload.Rating,
        PopularityScore: payload.PopularityScore,
        InStock:         payload.InStock,
        CreatedAt:       payload.UpdatedAt.Unix(),
    }
}

func cleanStringList(values []string) []string {
    cleaned := make([]string, 0, len(values))
    seen := map[string]struct{}{}

    for _, value := range values {
        item := strings.TrimSpace(value)
        if item == "" {
            continue
        }
        if _, exists := seen[item]; exists {
            continue
        }
        seen[item] = struct{}{}
        cleaned = append(cleaned, item)
    }

    return cleaned
}

func unixNow() int64 {
    return time.Now().UTC().Unix()
}
```

> 🟡 **Implementation note:** Task 1 schema me `created_at` sort field hai. Agar product payload me original created time available ho, `CreatedAt` ko `payload.CreatedAt.Unix()` se set karna better hai. Agar event me sirf `updated_at` available ho, Product Service event contract me `created_at` add karna recommended hai.

---

## Step 9: Typesense repository interface banao

Indexer ko Typesense client details directly nahi pata hone chahiye. Repository interface se testing easy hoti hai.

```go
package repository

import (
    "context"

    "ecommerce/backend/services/search-service/internal/domain"
)

type ProductIndexRepository interface {
    UpsertProduct(ctx context.Context, doc domain.ProductDocument) error
    DeleteProduct(ctx context.Context, productID string) error
}
```

### Why interface?

| Benefit | Explanation |
|---|---|
| Testability | Unit tests me fake repository inject kar sakte hain |
| Loose coupling | Indexer Typesense client package se independent rahega |
| Future flexibility | OpenSearch/Meilisearch migration easier ho sakti hai |

---

## Step 10: Typesense upsert/delete implementation banao

Typesense Go client document upsert ke liye use hoga.

```go
package repository

import (
    "context"

    "github.com/typesense/typesense-go/v2/typesense"
    "github.com/typesense/typesense-go/v2/typesense/api"

    "ecommerce/backend/services/search-service/internal/domain"
)

const productsCollection = "products"

type TypesenseProductRepository struct {
    client *typesense.Client
}

func NewTypesenseProductRepository(client *typesense.Client) *TypesenseProductRepository {
    return &TypesenseProductRepository{client: client}
}

func (r *TypesenseProductRepository) UpsertProduct(ctx context.Context, doc domain.ProductDocument) error {
    action := "upsert"
    params := &api.DocumentIndexParameters{Action: &action}

    _, err := r.client.Collection(productsCollection).
        Documents().
        Upsert(ctx, doc, params)

    return err
}

func (r *TypesenseProductRepository) DeleteProduct(ctx context.Context, productID string) error {
    _, err := r.client.Collection(productsCollection).
        Document(productID).
        Delete(ctx)

    if isTypesenseNotFound(err) {
        return nil
    }

    return err
}
```

### How this part was built

- `products` collection Task 1 se liya gaya.
- Upsert action safe hai: document exist karta hai to update, nahi karta to create.
- Delete idempotent banaya gaya: agar document already missing hai, success treat karo.
- `isTypesenseNotFound` helper client error type ke basis pe implement hoga.

Example helper:

```go
func isTypesenseNotFound(err error) bool {
    if err == nil {
        return false
    }

    // Exact implementation Typesense Go client error type ke according adjust hogi.
    return strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "Not Found")
}
```

> 🔴 **Production note:** String matching MVP ke liye okay ho sakta hai, but production me Typesense client ke typed error/status code ko check karna better hai.

---

## Step 11: Product indexer workflow banao

```mermaid
flowchart TD
    START([Message received])
    DECODE[Decode envelope]
    DUP{Already processed?}
    VALID{Payload valid?}
    DELETE_EVENT{Delete-type event?}
    SEARCHABLE{Product searchable?}
    UPSERT[Typesense upsert]
    DELETE[Typesense delete]
    MARK[Mark event processed]
    ACK[Ack message]
    RETRY[Retry/Nack]
    DLQ[Dead-letter queue]

    START --> DECODE
    DECODE --> DUP
    DUP -- yes --> ACK
    DUP -- no --> VALID
    VALID -- no --> DLQ
    VALID -- yes --> DELETE_EVENT
    DELETE_EVENT -- yes --> DELETE
    DELETE_EVENT -- no --> SEARCHABLE
    SEARCHABLE -- yes --> UPSERT
    SEARCHABLE -- no --> DELETE
    UPSERT --> MARK
    DELETE --> MARK
    MARK --> ACK
    UPSERT -. transient failure .-> RETRY
    DELETE -. transient failure .-> RETRY
    RETRY -. max retries exceeded .-> DLQ
```

Indexer code example:

```go
package indexer

import (
    "context"
    "errors"

    sharedEvents "ecommerce/backend/shared/events"
    "ecommerce/backend/services/search-service/internal/events"
    "ecommerce/backend/services/search-service/internal/repository"
)

type ProductIndexer struct {
    repo repository.ProductIndexRepository
}

func NewProductIndexer(repo repository.ProductIndexRepository) *ProductIndexer {
    return &ProductIndexer{repo: repo}
}

func (i *ProductIndexer) Handle(ctx context.Context, envelope sharedEvents.Envelope[events.ProductIndexPayload]) error {
    payload := envelope.Payload

    if payload.ProductID == "" {
        return ErrInvalidProductEvent
    }

    if isDeleteEvent(envelope.EventType) {
        return i.repo.DeleteProduct(ctx, payload.ProductID)
    }

    if !IsSearchable(payload) {
        return i.repo.DeleteProduct(ctx, payload.ProductID)
    }

    doc := MapProductToDocument(payload)
    return i.repo.UpsertProduct(ctx, doc)
}

func isDeleteEvent(eventType string) bool {
    switch eventType {
    case "ProductDeleted", "ProductUnpublished", "ProductBlocked":
        return true
    default:
        return false
    }
}

var ErrInvalidProductEvent = errors.New("invalid product event")
```

### Hinglish explanation

- Delete-type event direct Typesense delete karega.
- Non-delete event pe `IsSearchable` check hoga.
- Agar product publish nahi hai, index se remove hoga.
- Agar product valid and published hai, document upsert hoga.

---

## Step 12: Idempotency add karo

Queue delivery at-least-once hoti hai. Same message duplicate aa sakta hai. Isliye `event_id` processed store me save karna zaruri hai.

```go
type ProcessedEventStore interface {
    WasProcessed(ctx context.Context, eventID string) (bool, error)
    MarkProcessed(ctx context.Context, eventID string) error
}
```

Consumer wrapper example:

```go
func (c *Consumer) handleMessage(ctx context.Context, raw []byte) error {
    envelope, err := decodeProductEnvelope(raw)
    if err != nil {
        return PermanentError{Err: err}
    }

    processed, err := c.eventStore.WasProcessed(ctx, envelope.EventID)
    if err != nil {
        return err
    }
    if processed {
        return nil
    }

    if err := c.indexer.Handle(ctx, envelope); err != nil {
        return err
    }

    return c.eventStore.MarkProcessed(ctx, envelope.EventID)
}
```

### Storage options

| Option | Use Case | Notes |
|---|---|---|
| Redis set/key with TTL | MVP and fast duplicate protection | `processed:event_id` key with 7-30 day TTL |
| MySQL table | Strong audit trail | Search Service currently DB-owning nahi hai |
| Kafka offset only | Kafka-specific simple path | Duplicate event id still useful |

> 🟢 **Recommendation:** MVP ke liye Redis-backed `ProcessedEventStore` use karo, kyunki platform me Redis already planned hai.

---

## Step 13: RabbitMQ consumer design karo

Local Platform Foundation me RabbitMQ default beginner-friendly option hai.

### Queue naming

```text
Exchange: product.events
Queue: search-service.product-indexer
Routing keys:
  product.published
  product.updated
  product.price_changed
  product.inventory_changed
  product.unpublished
  product.deleted
  product.blocked
DLQ:
  search-service.product-indexer.dlq
```

### Consumer config

```dotenv
QUEUE_PROVIDER=rabbitmq
RABBITMQ_URL=amqp://ecommerce:ecommerce_password@rabbitmq:5672/ecommerce
PRODUCT_EVENTS_EXCHANGE=product.events
PRODUCT_INDEXER_QUEUE=search-service.product-indexer
PRODUCT_INDEXER_DLX=product.events.dlx
PRODUCT_INDEXER_DLQ=search-service.product-indexer.dlq
PRODUCT_INDEXER_PREFETCH=20
PRODUCT_INDEXER_MAX_RETRIES=5
```

### RabbitMQ consume example

```go
msgs, err := channel.Consume(
    "search-service.product-indexer",
    "search-service-1",
    false, // autoAck false: success ke baad manual ack
    false,
    false,
    false,
    nil,
)
if err != nil {
    return err
}

for msg := range msgs {
    ctx := context.Background()

    err := consumer.handleMessage(ctx, msg.Body)
    switch {
    case err == nil:
        _ = msg.Ack(false)
    case IsPermanentError(err):
        publishToDLQ(msg.Body, err)
        _ = msg.Ack(false)
    default:
        _ = msg.Nack(false, true)
    }
}
```

### Ack/Nack rules

| Situation | Action | Why |
|---|---|---|
| Successful upsert/delete | `Ack` | Message processed |
| Duplicate `event_id` | `Ack` | Already processed |
| Invalid JSON/schema | DLQ + `Ack` | Requeue se same poison message loop hoga |
| Typesense temporary error | `Nack` with retry | Service recover ho sakti hai |
| Max retries crossed | DLQ + `Ack` | Infinite retry avoid |

---

## Step 14: Kafka alternate design rakho

Docs me Kafka/RabbitMQ dono allowed hain. Agar team Kafka choose kare:

```dotenv
QUEUE_PROVIDER=kafka
KAFKA_BROKERS=kafka:9092
PRODUCT_EVENTS_TOPIC=product.events
PRODUCT_INDEXER_GROUP=search-service-product-indexer
```

Kafka consumer idea:

```go
reader := kafka.NewReader(kafka.ReaderConfig{
    Brokers: []string{"kafka:9092"},
    Topic:   "product.events",
    GroupID: "search-service-product-indexer",
})

for {
    msg, err := reader.FetchMessage(ctx)
    if err != nil {
        return err
    }

    if err := consumer.handleMessage(ctx, msg.Value); err != nil {
        // Do not commit offset on transient failure.
        continue
    }

    if err := reader.CommitMessages(ctx, msg); err != nil {
        return err
    }
}
```

> 🔵 **Choice:** Current guide RabbitMQ ko default maan kar chalti hai because local stack guides me RabbitMQ simple MVP default hai. Kafka high-throughput/replay-heavy production event streaming ke liye strong option rahega.

---

## Step 15: Retry and DLQ strategy banao

```mermaid
flowchart LR
    MSG[Product event]
    CONSUMER[Consumer]
    OK[Typesense success]
    RETRY[Retry with backoff]
    DLQ[Dead-letter queue]
    ALERT[Alert/log for manual review]

    MSG --> CONSUMER
    CONSUMER -->|success| OK
    CONSUMER -->|temporary failure| RETRY
    RETRY -->|retry available| CONSUMER
    RETRY -->|max retries exceeded| DLQ
    DLQ --> ALERT
```

### Error classification

| Error Type | Example | Handling |
|---|---|---|
| Permanent | Invalid JSON, missing `product_id`, unsupported event version | DLQ |
| Transient | Typesense timeout, network reset, queue connection blip | Retry |
| Idempotent success | Delete missing document | Treat as success |
| Unknown | Unexpected panic/error | Retry first, DLQ after max retries |

### Backoff suggestion

```text
Attempt 1: 1 second
Attempt 2: 5 seconds
Attempt 3: 15 seconds
Attempt 4: 60 seconds
Attempt 5: 300 seconds
After max: DLQ
```

> 🔴 **Important:** DLQ messages ko ignore mat karo. Monitoring alert lagao, warna index stale ho sakta hai.

---

## Step 16: Typesense client config banao

Task 2 me Typesense env already define hua tha. Task 3 same config use karega.

```dotenv
TYPESENSE_HOST=typesense
TYPESENSE_PORT=8108
TYPESENSE_PROTOCOL=http
TYPESENSE_API_KEY=dev-typesense-key
TYPESENSE_PRODUCTS_COLLECTION=products
```

Go config example:

```go
type Config struct {
    TypesenseHost               string
    TypesensePort               string
    TypesenseProtocol           string
    TypesenseAPIKey             string
    TypesenseProductsCollection string

    QueueProvider       string
    RabbitMQURL         string
    ProductEventsSource string
    ProductIndexerQueue string
    ProductIndexerDLQ   string
    ProductIndexerMaxRetries int
}
```

Typesense client example:

```go
client := typesense.NewClient(
    typesense.WithServer(cfg.TypesenseProtocol + "://" + cfg.TypesenseHost + ":" + cfg.TypesensePort),
    typesense.WithAPIKey(cfg.TypesenseAPIKey),
)
```

---

## Step 17: Ensure collection exists on startup

Task 1 ne schema define kiya tha. Task 3 indexer ko startup pe ensure karna chahiye ki `products` collection available hai.

```go
func EnsureProductsCollection(ctx context.Context, client *typesense.Client) error {
    schema := &api.CollectionSchema{
        Name: "products",
        Fields: []api.Field{
            {Name: "id", Type: "string"},
            {Name: "title", Type: "string"},
            {Name: "description", Type: "string", Optional: boolPtr(true)},
            {Name: "brand", Type: "string", Facet: boolPtr(true), Optional: boolPtr(true)},
            {Name: "category_ids", Type: "string[]", Facet: boolPtr(true)},
            {Name: "seller_id", Type: "string", Facet: boolPtr(true)},
            {Name: "price", Type: "float", Facet: boolPtr(true), Sort: boolPtr(true)},
            {Name: "rating", Type: "float", Facet: boolPtr(true), Sort: boolPtr(true), Optional: boolPtr(true)},
            {Name: "popularity_score", Type: "int32", Facet: boolPtr(true), Sort: boolPtr(true)},
            {Name: "in_stock", Type: "bool", Facet: boolPtr(true)},
            {Name: "created_at", Type: "int64", Sort: boolPtr(true)},
        },
        DefaultSortingField: stringPtr("popularity_score"),
    }

    _, err := client.Collections().Create(ctx, schema)
    if isAlreadyExists(err) {
        return nil
    }
    return err
}
```

### Why startup check?

| Reason | Explanation |
|---|---|
| Fresh local setup | Empty Typesense me consumer fail nahi hoga |
| Kubernetes restart | Service self-check kar sakega |
| Developer UX | Manual collection creation step kam hoga |

> 🟡 **Boundary:** Schema migration/versioning ka advanced process later improve ho sakta hai. Full catalog reindex Task 8 me covered hoga.

---

## Step 18: Main wiring define karo

`main.go` ka high-level responsibility:

```go
func main() {
    cfg := config.Load()
    log := logger.New(cfg.ServiceName)

    tsClient := typesense.NewClient(
        typesense.WithServer(cfg.TypesenseURL()),
        typesense.WithAPIKey(cfg.TypesenseAPIKey),
    )

    if err := schema.EnsureProductsCollection(context.Background(), tsClient); err != nil {
        log.Fatal("failed to ensure products collection", err)
    }

    repo := repository.NewTypesenseProductRepository(tsClient)
    productIndexer := indexer.NewProductIndexer(repo)

    consumer := events.NewProductConsumer(events.ProductConsumerConfig{
        RabbitMQURL: cfg.RabbitMQURL,
        QueueName:  cfg.ProductIndexerQueue,
        MaxRetries: cfg.ProductIndexerMaxRetries,
    }, productIndexer)

    if err := consumer.Run(context.Background()); err != nil {
        log.Fatal("product indexer stopped", err)
    }
}
```

### How this part was built

- Startup order deliberate hai: config -> logger -> Typesense -> schema -> repo -> indexer -> consumer.
- Consumer tabhi start hota hai jab Typesense ready and schema available ho.
- Fatal startup failure better hai because half-running consumer silently messages fail karega.

---

## Step 19: Observability add karo

Task 3 me indexing silent nahi honi chahiye. Har event ka traceable log and metric hona chahiye.

### Logs

```json
{
  "level": "info",
  "service": "search-service",
  "component": "product-indexer",
  "event_id": "evt_01HYSEARCH123",
  "event_type": "ProductUpdated",
  "product_id": "prod_123",
  "action": "upsert",
  "result": "success",
  "duration_ms": 18,
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736"
}
```

### Metrics

| Metric | Type | Labels | Why |
|---|---|---|---|
| `search_indexer_events_total` | Counter | `event_type`, `result` | Event processing volume |
| `search_indexer_duration_ms` | Histogram | `action` | Indexing latency |
| `search_indexer_failures_total` | Counter | `error_type` | Alerting |
| `search_indexer_dlq_total` | Counter | `reason` | Poison message visibility |
| `search_indexer_queue_lag` | Gauge | `queue` | Backlog monitoring |
| `typesense_writes_total` | Counter | `action`, `result` | Typesense write health |

### Trace propagation

| Source Field | Use |
|---|---|
| `trace_id` | Create/continue span for async processing |
| `request_id` | Log correlation |
| `event_id` | Span attribute and idempotency key |
| `correlation_id` | Product-level journey trace |

---

## Step 20: Security and data safety rules add karo

| Rule | Explanation |
|---|---|
| Typesense API key server-side only | Browser/app me expose nahi karna |
| Product payload sanitized | Private seller/admin fields index me nahi aane chahiye |
| Event payload allowlist | Sirf defined fields map karo, raw product object dump mat karo |
| DLQ access restricted | DLQ me business data ho sakta hai |
| Logs PII-free | Product indexer logs me user tokens, emails, phone numbers nahi |
| Queue credentials secret me | `RABBITMQ_URL`, `TYPESENSE_API_KEY` git me real value ke saath commit nahi |

---

## Step 21: Local smoke test plan banao

### Start dependencies

```bash
docker compose -f infra/compose/docker-compose.local.yml --env-file infra/compose/.env.local up -d typesense rabbitmq
```

### Publish test event

RabbitMQ UI ya local publisher se ye event bhejo:

```json
{
  "event_id": "evt_test_001",
  "event_type": "ProductPublished",
  "version": 1,
  "producer": "product-service",
  "request_id": "req_test_001",
  "trace_id": "trace_test_001",
  "correlation_id": "prod_test_001",
  "occurred_at": "2026-05-23T10:30:00Z",
  "payload": {
    "product_id": "prod_test_001",
    "title": "Test Running Shoes",
    "description": "Comfortable shoes for testing search indexing",
    "brand": "TestBrand",
    "category_ids": ["cat_shoes"],
    "seller_id": "seller_test",
    "price": 1299,
    "rating": 4.2,
    "popularity_score": 10,
    "in_stock": true,
    "status": "published",
    "is_deleted": false,
    "updated_at": "2026-05-23T10:29:45Z"
  }
}
```

### Verify document in Typesense

```bash
curl "http://localhost:8108/collections/products/documents/prod_test_001" \
  -H "X-TYPESENSE-API-KEY: dev-typesense-key"
```

### Delete test event

```json
{
  "event_id": "evt_test_002",
  "event_type": "ProductDeleted",
  "version": 1,
  "producer": "product-service",
  "request_id": "req_test_002",
  "trace_id": "trace_test_002",
  "correlation_id": "prod_test_001",
  "occurred_at": "2026-05-23T10:35:00Z",
  "payload": {
    "product_id": "prod_test_001",
    "status": "deleted",
    "is_deleted": true,
    "updated_at": "2026-05-23T10:34:45Z"
  }
}
```

Expected result:

```text
GET /collections/products/documents/prod_test_001 -> 404 Not Found
```

---

## Step 22: Test cases likho

### Unit tests

| Test | Expected |
|---|---|
| `ProductPublished` with valid payload | `UpsertProduct` called |
| `ProductUpdated` with `status=published` | `UpsertProduct` called |
| `ProductUpdated` with `status=draft` | `DeleteProduct` called |
| `ProductDeleted` | `DeleteProduct` called |
| Missing `product_id` | Permanent validation error |
| Duplicate `category_ids` | Mapper returns unique clean category ids |
| Delete missing Typesense document | Success |
| Duplicate `event_id` | Indexer skipped, message acked |

### Integration tests

| Test | Expected |
|---|---|
| Consumer receives valid event from RabbitMQ | Typesense document created |
| Consumer receives delete event | Typesense document removed |
| Typesense temporarily down | Message retried |
| Invalid JSON event | Message moved to DLQ |
| Same event delivered twice | One index write, second acked |

Example unit test skeleton:

```go
func TestProductIndexerUpsertsPublishedProduct(t *testing.T) {
    repo := &FakeProductIndexRepository{}
    indexer := NewProductIndexer(repo)

    envelope := events.Envelope[ProductIndexPayload]{
        EventID:   "evt_1",
        EventType: "ProductPublished",
        Payload: ProductIndexPayload{
            ProductID:   "prod_1",
            Title:       "Running Shoes",
            CategoryIDs: []string{"cat_shoes"},
            SellerID:    "seller_1",
            Status:      "published",
            Price:       999,
            InStock:     true,
        },
    }

    err := indexer.Handle(context.Background(), envelope)

    require.NoError(t, err)
    require.Equal(t, "prod_1", repo.Upserted[0].ID)
}
```

---

## Step 23: Deployment configuration plan banao

### Search Service env

```yaml
env:
  - name: TYPESENSE_HOST
    value: "typesense.data.svc.cluster.local"
  - name: TYPESENSE_PORT
    value: "8108"
  - name: TYPESENSE_PROTOCOL
    value: "http"
  - name: PRODUCT_EVENTS_EXCHANGE
    value: "product.events"
  - name: PRODUCT_INDEXER_QUEUE
    value: "search-service.product-indexer"
  - name: PRODUCT_INDEXER_PREFETCH
    value: "20"
  - name: PRODUCT_INDEXER_MAX_RETRIES
    value: "5"
  - name: TYPESENSE_API_KEY
    valueFrom:
      secretKeyRef:
        name: search-service-secrets
        key: TYPESENSE_API_KEY
  - name: RABBITMQ_URL
    valueFrom:
      secretKeyRef:
        name: search-service-secrets
        key: RABBITMQ_URL
```

### Readiness check

Search Service should be ready only when:

- Typesense `/health` pass ho.
- `products` collection available ho.
- Queue connection established ho.

---

## Step 24: Rollout strategy define karo

```mermaid
flowchart TD
    A[Deploy Search Service indexer disabled]
    B[Verify config and Typesense connection]
    C[Enable consumer for product.events]
    D[Publish test product event]
    E[Verify Typesense document]
    F[Monitor logs, metrics, DLQ]
    G[Scale replicas gradually]

    A --> B --> C --> D --> E --> F --> G
```

### Safe rollout checklist

| Check | Expected |
|---|---|
| Typesense health | Healthy |
| Queue connection | Connected |
| `products` collection | Exists |
| Test upsert event | Document visible |
| Test delete event | Document removed |
| DLQ count | Zero for valid events |
| Queue lag | Stable/decreasing |
| Error rate | Near zero |

---

## 📦 External Tools / Libraries Used

## 1. Typesense Go Client

| Item | Detail |
|---|---|
| What | Official Go client for Typesense APIs |
| Why used | `products` collection me document upsert/delete karne ke liye |
| Install | `go get github.com/typesense/typesense-go/v2/typesense` |
| Used in | `internal/repository/typesense_repository.go` |

Example:

```bash
go get github.com/typesense/typesense-go/v2/typesense
```

```go
client := typesense.NewClient(
    typesense.WithServer("http://localhost:8108"),
    typesense.WithAPIKey("dev-typesense-key"),
)
```

---

## 2. RabbitMQ AMQP 0-9-1 Go Client

| Item | Detail |
|---|---|
| What | Go client for RabbitMQ AMQP protocol |
| Why used | `product.events` queue consume karne ke liye |
| Install | `go get github.com/rabbitmq/amqp091-go` |
| Used in | `internal/events/product_consumer.go` |

Example:

```bash
go get github.com/rabbitmq/amqp091-go
```

```go
conn, err := amqp.Dial(cfg.RabbitMQURL)
if err != nil {
    return err
}
defer conn.Close()
```

---

## 3. Kafka Go Client - Optional Alternative

| Item | Detail |
|---|---|
| What | Go Kafka client |
| Why used | Agar team RabbitMQ ki jagah Kafka choose kare |
| Install | `go get github.com/segmentio/kafka-go` |
| Used in | Kafka-specific `product_consumer.go` implementation |

Example:

```bash
go get github.com/segmentio/kafka-go
```

```go
reader := kafka.NewReader(kafka.ReaderConfig{
    Brokers: []string{"kafka:9092"},
    Topic:   "product.events",
    GroupID: "search-service-product-indexer",
})
```

---

## 4. Shared Platform Packages

| Package | What | Why used |
|---|---|---|
| `backend/shared/events` | Event envelope/interface | Consistent event shape |
| `backend/shared/logger` | Structured logging | Traceable indexing logs |
| `backend/shared/config` | Env loading | Predictable config |
| `backend/shared/idempotency` | Duplicate protection interface | Safe retries |
| `backend/shared/health` | Health checks | Kubernetes readiness/liveness |

Install needed?

```text
No external install. Ye monorepo ke shared packages honge.
```

---

## 5. Mermaid

| Item | Detail |
|---|---|
| What | Markdown-friendly diagram syntax |
| Why used | Architecture, event flow, retry flow visually explain karne ke liye |
| Install needed? | Usually no, GitHub/GitLab/Markdown tools render kar dete hain |
| Used here? | Yes |

---

## 🧪 Verification Checklist

| Check | Status |
|---|---|
| `TaskImplementation/` folder exists | ✅ Done |
| `TaskImplementation/Search Service/` folder exists | ✅ Done |
| `task3.md` created | ✅ Done |
| Task source clearly mentioned | ✅ Done |
| Hinglish step-by-step guide added | ✅ Done |
| Product event consumer design documented | ✅ Done |
| Typesense upsert/delete flow documented | ✅ Done |
| External libraries/tools explained | ✅ Done |
| Folder structure included | ✅ Done |
| Code examples included | ✅ Done |
| Mermaid diagrams included | ✅ Done |
| Task 4+ implementation avoided | ✅ Done |

---

## 🧾 Final Notes

Search Service - Task 3 ka output ek event-driven product indexer blueprint hai:

- Product events `product.events` se consume honge.
- Published/searchable products Typesense `products` collection me upsert honge.
- Deleted/unpublished/blocked products Typesense se remove honge.
- Duplicate events idempotency se safe rahenge.
- Failures retry and DLQ flow se controlled rahenge.
- Search API, autocomplete, synonyms, zero-result analytics, aur full reindex intentionally skip kiye gaye because wo later Search Service tasks hain.

> 🟢 **Next task:** Search Service - Task 4 me `GET /api/v1/search` / `SearchProducts` flow implement hoga, jahan Typesense query, filters, facets, sorting, pagination, aur Product Service canonical verification add hogi.
