# 💖 Wishlist Service - Task 6: Availability Sync

![Task](https://img.shields.io/badge/Task-Wishlist%20Service%20%236-blue)
![Priority](https://img.shields.io/badge/Priority-P2-yellow)
![Dependency](https://img.shields.io/badge/Dependency-Product%20Events-brightgreen)
![Status](https://img.shields.io/badge/Status-Documented-success)
![Guide](https://img.shields.io/badge/Guide-Hinglish-orange)
![Scope](https://img.shields.io/badge/Scope-Availability%20Sync-purple)

---

## 📌 Task Summary

| Field | Detail |
|---|---|
| Service | Wishlist Service |
| Task | Task 6 - Availability sync |
| Source | `docs/01-micro-tasks.md` -> `Wishlist Service` -> Task 6 |
| Requirement | Product deleted/out-of-stock hone pe wishlist status update karo. |
| Dependency | Product events |
| Priority | P2 |
| Final Decision | **Wishlist item ko delete nahi karna; Product event consume karke item ka `availability` snapshot update karna hai** |
| Output | Structured Hinglish implementation guide for Wishlist Task 6 |

> **Simple Hinglish goal:** Agar Product Service batata hai ki product delete ho gaya ya out of stock ho gaya, to Wishlist Service apne saved wishlist items me same product ka `availability` status update karega. Buyer ki wishlist clean rahegi, but item silently remove nahi hoga.

---

## ✅ Final Output Created

```text
TaskImplementation/
└── Wishlist Service/
    ├── task1.md
    ├── task2.md
    ├── task3.md
    ├── task4.md
    ├── task5.md
    └── task6.md
```

### Why this structure?

| Folder/File | Purpose |
|---|---|
| `TaskImplementation/` | Saare task-wise implementation guides ka central folder |
| `Wishlist Service/` | Wishlist service ke task guides ko group karta hai |
| `task1.md` | Wishlist model decision guide |
| `task2.md` | Wishlist MongoDB choice guide |
| `task3.md` | `wishlists` collection design guide |
| `task4.md` | Add/remove item APIs guide |
| `task5.md` | Move-to-cart guide |
| `task6.md` | Sirf **Wishlist Service - Task 6** ka availability sync guide |

> 🟢 **Note:** Is document me Task 6 ka guide diya gaya hai. Task 7 price drop notifications aur Task 8 analytics events yahan implement nahi kiye gaye.

---

## 🧭 Documents Studied

| Document/File | Kya use kiya gaya |
|---|---|
| `docs/01-micro-tasks.md` | Task 6 ka exact scope: Product deleted/out-of-stock event ke baad wishlist status update |
| `docs/04-microservice-design.md` | Wishlist Service responsibilities: price/availability tracking and Product event consumption |
| `docs/02-system-architecture.md` | Async events ke liye Kafka/RabbitMQ, direct cross-service DB access banned |
| `docs/03-folder-structure.md` | Wishlist service me `internal/events/product_consumer.go` type structure ka reference |
| `docs/11-devops-external-services.md` | `product.events` topic/queue, event envelope, retry, DLQ, idempotent handling |
| `database/mongodb-schema-design.md` | `wishlist_db.wishlists.items.availability` field and `items.product_id` index |
| `api/master-api.json` | Wishlist public APIs me Task 6 ke liye koi naya REST endpoint required nahi |
| `TaskImplementation/Wishlist Service/task1.md` | Single private wishlist per buyer |
| `TaskImplementation/Wishlist Service/task2.md` | Wishlist Service owns MongoDB database |
| `TaskImplementation/Wishlist Service/task3.md` | `availability` enum already planned: `in_stock`, `out_of_stock`, `unknown`, `deleted` |
| `TaskImplementation/Wishlist Service/task4.md` | Product validation ke time initial availability snapshot save hota hai |
| `TaskImplementation/Wishlist Service/task5.md` | Move-to-cart Task 6 se separate hai |
| `backend/services/wishlist-service/internal/domain/wishlist.go` | Existing domain availability enum |
| `backend/services/wishlist-service/internal/repository/mongo_wishlist_repository.go` | Existing Mongo repository and `items.product_id` index |

---

## 🧱 Task Boundary

### ✅ Included in Task 6

| Included | Explanation |
|---|---|
| Product event consumer design | Wishlist Service `product.events` consume karega |
| Product deleted sync | Matching wishlist items ka `availability = deleted` hoga |
| Product out-of-stock sync | Matching wishlist items ka `availability = out_of_stock` hoga |
| MongoDB embedded item update | `wishlists.items` array me matching product status update hoga |
| Variant-aware update rule | Agar event me `variant_id` hai to same variant item update hoga; product-level event all product items update karega |
| Idempotent handling | Same event repeat hone par final status same rahega |
| Retry and DLQ guidance | DB failure pe retry; poison message pe DLQ |
| Observability | Structured logs, metrics, trace/request correlation |
| Code examples | Domain/usecase/repository/consumer snippets |
| Diagrams | Architecture, sequence, and decision flow diagrams |

### ❌ Not Included in Task 6

| Not Included | Future Task |
|---|---|
| Price drop notification | Wishlist Service Task 7 |
| Wishlist add/remove analytics events | Wishlist Service Task 8 |
| Recommendation Service event publishing | Wishlist Service Task 8 |
| Product Service event producer implementation | Product Service Task 7 / Product events task |
| Product DB direct query | Banned by microservice rule |
| Auto-remove wishlist item | Not required; Task 6 only updates status |
| Frontend wishlist UI changes | User App Frontend profile/wishlist task |
| Cart behavior changes | Cart Service / Wishlist Task 5 already separate |

> 🔴 **Important boundary:** Wishlist Service Product Service ke MongoDB ko direct read nahi karega. Product availability change async event ke through milega.

---

## 🧩 Final Design Decision

### Status update rule

| Product event | Wishlist item update |
|---|---|
| `ProductDeleted` | `availability = "deleted"` |
| `ProductOutOfStock` | `availability = "out_of_stock"` |

### Item remove nahi karna

Wishlist item ko automatically delete karna tempting lag sakta hai, but buyer ke liye saved intent important hota hai. Better UX:

```text
Product unavailable hua -> Wishlist me item dikhega -> Status "deleted" / "out_of_stock" dikhega -> Buyer manually remove kar sakta hai
```

### Public API impact

Task 6 me koi new public REST API nahi chahiye.

Existing `GET /api/v1/wishlist` response me item ka `availability` field fresh snapshot ke saath return hoga.

---

## 🏗️ Architecture

```mermaid
flowchart LR
    ProductService[Product Service] -->|publish ProductDeleted/ProductOutOfStock| MQ[(Kafka/RabbitMQ product.events)]
    MQ --> ProductConsumer[Wishlist Product Event Consumer]
    ProductConsumer --> SyncUsecase[SyncProductAvailability Usecase]
    SyncUsecase --> WishlistRepo[Wishlist Repository]
    WishlistRepo --> Mongo[(MongoDB wishlist_db.wishlists)]
    BuyerApp[Buyer App] --> Gateway[API Gateway]
    Gateway --> WishlistAPI[Wishlist GET API]
    WishlistAPI --> Mongo
```

### Hinglish Explanation

- Product Service product status/inventory change detect karta hai.
- Product Service `product.events` topic/queue me event publish karta hai.
- Wishlist Service ka Product Event Consumer event receive karta hai.
- Consumer event validate karke usecase ko call karta hai.
- Usecase repository ke through `wishlists.items.availability` update karta hai.
- Buyer jab wishlist fetch karega, updated availability dikhegi.

---

## 🔁 Event Sync Flow

```mermaid
sequenceDiagram
    participant Product as Product Service
    participant MQ as product.events
    participant Consumer as Wishlist Product Consumer
    participant Usecase as Wishlist Usecase
    participant Mongo as wishlist_db.wishlists

    Product->>MQ: Publish ProductOutOfStock(event_id, product_id)
    MQ->>Consumer: Deliver event
    Consumer->>Consumer: Validate envelope and payload
    Consumer->>Usecase: SyncProductAvailability(product_id, out_of_stock)
    Usecase->>Mongo: updateMany items.$[item].availability
    Mongo-->>Usecase: matched_count, modified_count
    Usecase-->>Consumer: Sync result
    Consumer->>MQ: Commit ack/offset after DB success
```

---

## 🪜 Step-by-Step Implementation

## Step 1: Existing Wishlist Availability Field Ko Base Banaya

Task 3 aur current backend domain me ye availability values already planned hain:

```go
type Availability string

const (
    AvailabilityInStock    Availability = "in_stock"
    AvailabilityOutOfStock Availability = "out_of_stock"
    AvailabilityUnknown    Availability = "unknown"
    AvailabilityDeleted    Availability = "deleted"
)
```

### Why this matters?

Task 6 ke liye schema migration minimal rahega kyunki `items.availability` already exists. Hume sirf event consumer aur update logic add karna hai.

---

## Step 2: Product Events Ka Minimal Contract Final Kiya

Docs me common event envelope diya gaya hai. Wishlist Service ko Product events ke liye same envelope consume karna chahiye.

### Event envelope

```json
{
  "event_id": "evt_123",
  "event_type": "ProductOutOfStock",
  "version": 1,
  "occurred_at": "2026-05-26T10:30:00Z",
  "producer": "product-service",
  "trace_id": "trace_123",
  "payload": {}
}
```

### Product deleted payload

```json
{
  "product_id": "prod_123",
  "reason": "seller_deleted"
}
```

### Product out-of-stock payload

```json
{
  "product_id": "prod_123",
  "variant_id": "var_1",
  "reason": "inventory_zero"
}
```

### Field explanation

| Field | Required | Explanation |
|---|---:|---|
| `event_id` | ✅ | Unique event id for logs, tracing, duplicate-safe behavior |
| `event_type` | ✅ | `ProductDeleted` ya `ProductOutOfStock` |
| `version` | ✅ | Schema version; unsupported version ko DLQ/send alert |
| `occurred_at` | ✅ | Event ka actual business time |
| `producer` | ✅ | Expected value: `product-service` |
| `trace_id` | Optional | Cross-service trace correlation |
| `payload.product_id` | ✅ | Wishlist item match karne ke liye |
| `payload.variant_id` | Optional | Variant-specific stock event ke liye |
| `payload.reason` | Optional | Logging/debugging ke liye |

> 🟡 **Contract note:** `api/master-api.json` me Product event schema explicitly defined nahi hai, isliye Task 6 Wishlist consumption ke liye minimal expected payload standardize karta hai.

---

## Step 3: Event Type Se Wishlist Availability Map Kiya

Mapping simple aur deterministic rakhi gayi:

```text
ProductDeleted     -> deleted
ProductOutOfStock  -> out_of_stock
```

### Go example

```go
func availabilityFromProductEvent(eventType string) (domain.Availability, bool) {
    switch eventType {
    case "ProductDeleted":
        return domain.AvailabilityDeleted, true
    case "ProductOutOfStock":
        return domain.AvailabilityOutOfStock, true
    default:
        return domain.AvailabilityUnknown, false
    }
}
```

### Why deterministic mapping?

Same event agar retry ho jaye to same item same status pe set hoga. Isse duplicate event harmless ban jata hai.

---

## Step 4: Usecase Input Define Kiya

Availability sync public buyer command nahi hai. Ye internal async event se trigger hota hai.

```go
type SyncProductAvailabilityInput struct {
    EventID      string
    EventType    string
    ProductID    string
    VariantID    string
    Availability domain.Availability
    OccurredAt   time.Time
    TraceID      string
}
```

### Validation rules

| Rule | Error behavior |
|---|---|
| `product_id` empty | Event reject, no DB update |
| Unsupported `event_type` | Ignore with debug log |
| Unsupported `version` | Send to DLQ / alert |
| Invalid `availability` | Reject as bad event |
| `occurred_at` empty | Use processing time but log warning |

---

## Step 5: Usecase Method Add Kiya

Usecase ka kaam event ko business command me convert karke repository ko call karna hai.

```go
type AvailabilitySyncResult struct {
    ProductID     string
    VariantID     string
    Availability  domain.Availability
    MatchedCount  int64
    ModifiedCount int64
}

func (s *WishlistService) SyncProductAvailability(
    ctx context.Context,
    input SyncProductAvailabilityInput,
) (AvailabilitySyncResult, error) {
    productID := normalizeID(input.ProductID)
    variantID := normalizeID(input.VariantID)
    if productID == "" {
        return AvailabilitySyncResult{}, ValidationError{
            Field:   "product_id",
            Message: "is required",
        }
    }
    if !input.Availability.IsValid() {
        return AvailabilitySyncResult{}, ValidationError{
            Field:   "availability",
            Message: "must be in_stock, out_of_stock, unknown, or deleted",
        }
    }

    occurredAt := input.OccurredAt.UTC()
    if occurredAt.IsZero() {
        occurredAt = s.clock().UTC()
    }

    result, err := s.repository.UpdateProductAvailability(
        ctx,
        productID,
        variantID,
        input.Availability,
        occurredAt,
    )
    if err != nil {
        return AvailabilitySyncResult{}, err
    }

    s.logger.Info("wishlist availability synced",
        "event_id", normalizeID(input.EventID),
        "event_type", input.EventType,
        "product_id", productID,
        "variant_id", variantID,
        "availability", input.Availability,
        "matched_count", result.MatchedCount,
        "modified_count", result.ModifiedCount,
        "trace_id", normalizeID(input.TraceID),
    )

    return AvailabilitySyncResult{
        ProductID:     productID,
        VariantID:     variantID,
        Availability:  input.Availability,
        MatchedCount:  result.MatchedCount,
        ModifiedCount: result.ModifiedCount,
    }, nil
}
```

### Why usecase layer?

Consumer sirf transport adapter hai. Business rule yahan rahega:

```text
Product event -> validated command -> wishlist availability sync
```

---

## Step 6: Repository Contract Extend Kiya

Existing repository add/remove/find karta hai. Task 6 ke liye ek targeted update method chahiye.

```go
type WishlistRepository interface {
    EnsureWishlist(ctx context.Context, userID string, now time.Time) error
    AddItemIfNotExists(ctx context.Context, userID string, item domain.WishlistItem, now time.Time) error
    RemoveItem(ctx context.Context, userID, productID string, now time.Time) error
    FindByUserID(ctx context.Context, userID string) (*domain.Wishlist, error)
    UpdateProductAvailability(
        ctx context.Context,
        productID string,
        variantID string,
        availability domain.Availability,
        updatedAt time.Time,
    ) (AvailabilityUpdateResult, error)
}
```

```go
type AvailabilityUpdateResult struct {
    MatchedCount  int64
    ModifiedCount int64
}
```

### Why separate repository method?

Availability sync ko har user ki wishlist load nahi karni chahiye. MongoDB atomic `updateMany` se matching embedded items update karna faster aur safer hai.

---

## Step 7: MongoDB Update Query Banaya

### Product-level deleted event

Product delete event usually product-level hota hai. Is case me same `product_id` wale saare wishlist items update honge.

```javascript
db.wishlists.updateMany(
  { "items.product_id": "prod_123" },
  {
    $set: {
      "items.$[item].availability": "deleted",
      "updated_at": ISODate("2026-05-26T10:30:00Z")
    }
  },
  {
    arrayFilters: [
      { "item.product_id": "prod_123" }
    ]
  }
)
```

### Variant-level out-of-stock event

Agar event me `variant_id` available hai, to sirf same variant update hogi.

```javascript
db.wishlists.updateMany(
  {
    "items.product_id": "prod_123",
    "items.variant_id": "var_1"
  },
  {
    $set: {
      "items.$[item].availability": "out_of_stock",
      "updated_at": ISODate("2026-05-26T10:30:00Z")
    }
  },
  {
    arrayFilters: [
      {
        "item.product_id": "prod_123",
        "item.variant_id": "var_1"
      }
    ]
  }
)
```

### Go repository example

```go
func (r *MongoWishlistRepository) UpdateProductAvailability(
    ctx context.Context,
    productID string,
    variantID string,
    availability domain.Availability,
    updatedAt time.Time,
) (AvailabilityUpdateResult, error) {
    productID = strings.TrimSpace(productID)
    variantID = strings.TrimSpace(variantID)
    if productID == "" {
        return AvailabilityUpdateResult{}, fmt.Errorf("%w: product_id is required", domain.ErrInvalidWishlist)
    }
    if !availability.IsValid() {
        return AvailabilityUpdateResult{}, fmt.Errorf("%w: invalid availability", domain.ErrInvalidWishlist)
    }

    filter := bson.M{"items.product_id": productID}
    itemFilter := bson.M{"item.product_id": productID}
    if variantID != "" {
        filter["items.variant_id"] = variantID
        itemFilter["item.variant_id"] = variantID
    }

    result, err := r.collection.UpdateMany(
        contextOrBackground(ctx),
        filter,
        bson.M{"$set": bson.M{
            "items.$[item].availability": availability,
            "updated_at":                 normalizeTime(updatedAt),
        }},
        options.UpdateMany().SetArrayFilters([]any{itemFilter}),
    )
    if err != nil {
        return AvailabilityUpdateResult{}, fmt.Errorf(
            "update wishlist availability for product %q variant %q: %w",
            productID,
            variantID,
            err,
        )
    }

    return AvailabilityUpdateResult{
        MatchedCount:  result.MatchedCount,
        ModifiedCount: result.ModifiedCount,
    }, nil
}
```

### Why `arrayFilters`?

Wishlist document ke andar `items` array hai. `arrayFilters` se MongoDB sirf matching array element update karta hai, baaki wishlist items untouched rehte hain.

---

## Step 8: Product Event Consumer Banaya

Consumer ka kaam:

```text
Read message -> Decode envelope -> Decode payload -> Map event type -> Call usecase -> Ack/commit
```

### Event structs

```go
type ProductEventEnvelope struct {
    EventID    string          `json:"event_id"`
    EventType  string          `json:"event_type"`
    Version    int             `json:"version"`
    OccurredAt time.Time       `json:"occurred_at"`
    Producer   string          `json:"producer"`
    TraceID    string          `json:"trace_id"`
    Payload    json.RawMessage `json:"payload"`
}

type ProductAvailabilityPayload struct {
    ProductID string `json:"product_id"`
    VariantID string `json:"variant_id"`
    Reason    string `json:"reason"`
}
```

### Handler example

```go
func (c *ProductConsumer) HandleMessage(ctx context.Context, body []byte) error {
    var envelope ProductEventEnvelope
    if err := json.Unmarshal(body, &envelope); err != nil {
        return fmt.Errorf("decode product event envelope: %w", err)
    }

    if envelope.Version != 1 {
        return fmt.Errorf("unsupported product event version %d", envelope.Version)
    }
    if envelope.Producer != "product-service" {
        return fmt.Errorf("unexpected product event producer %q", envelope.Producer)
    }

    availability, ok := availabilityFromProductEvent(envelope.EventType)
    if !ok {
        c.logger.Debug("product event ignored by wishlist",
            "event_id", envelope.EventID,
            "event_type", envelope.EventType,
        )
        return nil
    }

    var payload ProductAvailabilityPayload
    if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
        return fmt.Errorf("decode product availability payload: %w", err)
    }

    _, err := c.wishlistService.SyncProductAvailability(ctx, SyncProductAvailabilityInput{
        EventID:      envelope.EventID,
        EventType:    envelope.EventType,
        ProductID:    payload.ProductID,
        VariantID:    payload.VariantID,
        Availability: availability,
        OccurredAt:   envelope.OccurredAt,
        TraceID:      envelope.TraceID,
    })
    return err
}
```

---

## Step 9: Kafka/RabbitMQ Consumer Runtime Wire Kiya

Docs ke according Kafka recommended hai, RabbitMQ acceptable hai.

### Kafka recommended config

| Config | Example | Meaning |
|---|---|---|
| `WISHLIST_EVENTS_BACKEND` | `kafka` | Event backend |
| `WISHLIST_KAFKA_BROKERS` | `localhost:9092` | Kafka broker list |
| `WISHLIST_PRODUCT_EVENTS_TOPIC` | `product.events` | Product event topic |
| `WISHLIST_PRODUCT_EVENTS_GROUP_ID` | `wishlist-service` | Consumer group |
| `WISHLIST_PRODUCT_EVENTS_DLQ_TOPIC` | `product.events.wishlist.dlq` | Failed messages |

### RabbitMQ alternate config

| Config | Example | Meaning |
|---|---|---|
| `WISHLIST_EVENTS_BACKEND` | `rabbitmq` | Event backend |
| `WISHLIST_RABBITMQ_URL` | `amqp://guest:guest@localhost:5672/` | RabbitMQ connection |
| `WISHLIST_PRODUCT_EVENTS_QUEUE` | `wishlist.product.events` | Bound queue |
| `WISHLIST_PRODUCT_EVENTS_EXCHANGE` | `product.events` | Exchange |
| `WISHLIST_PRODUCT_EVENTS_DLQ` | `wishlist.product.events.dlq` | Dead-letter queue |

### Runtime rule

```text
DB update success -> ack/commit message
DB update failure -> retry
Permanent bad event -> DLQ
Ignored event type -> ack/commit
```

---

## Step 10: Idempotency Simple Rakhi

Task 6 me event handling naturally idempotent hai:

```text
Set prod_123 availability to out_of_stock
Set prod_123 availability to out_of_stock again
Final result: out_of_stock
```

### Why no new receipt collection in Task 6?

`docs/11-devops-external-services.md` idempotent consumer rule require karta hai. Is task me deterministic `$set` update se duplicate events safe ho jate hain. New `wishlist_events` ya receipt collection create karna analytics/audit scope me chala ja sakta hai, isliye Task 6 me avoid kiya gaya.

### Important broker rule

Product events ko `product_id` key se publish karna chahiye. Kafka me same key same partition me jati hai, jisse same product ke events ka order preserve hota hai.

---

## Step 11: Error Handling Aur Retry Design

| Scenario | Action |
|---|---|
| JSON decode fail | DLQ, because message malformed hai |
| Unsupported version | DLQ + alert |
| Unsupported event type | Ignore + ack |
| Missing product id | DLQ |
| Mongo temporary failure | Retry with backoff |
| Mongo permanent validation failure | DLQ |
| No matching wishlist item | Success; `matched_count = 0`, ack |

### Retry flow

```mermaid
flowchart TD
    Message[Product event received] --> Decode{Valid JSON and version?}
    Decode -- No --> DLQ[Send to DLQ]
    Decode -- Yes --> Type{Supported event type?}
    Type -- No --> AckIgnored[Ack ignored event]
    Type -- Yes --> Validate{Payload has product_id?}
    Validate -- No --> DLQ
    Validate -- Yes --> Update[Update wishlist availability]
    Update --> Success{DB success?}
    Success -- Yes --> Ack[Ack/commit]
    Success -- No --> Retry{Retry budget left?}
    Retry -- Yes --> Backoff[Retry with backoff]
    Backoff --> Update
    Retry -- No --> DLQ
```

---

## Step 12: Observability Add Kiya

Task 6 async hai, isliye logs and metrics important hain.

### Structured logs

```go
logger.Info("wishlist availability synced",
    "event_id", eventID,
    "event_type", eventType,
    "product_id", productID,
    "variant_id", variantID,
    "availability", availability,
    "matched_count", matchedCount,
    "modified_count", modifiedCount,
    "trace_id", traceID,
)
```

### Metrics

| Metric | Type | Labels | Meaning |
|---|---|---|---|
| `wishlist_product_events_consumed_total` | Counter | `event_type`, `result` | Kitne Product events process hue |
| `wishlist_availability_updates_total` | Counter | `availability` | Kitne wishlist item updates hue |
| `wishlist_product_event_failures_total` | Counter | `reason` | Decode/DB/version failures |
| `wishlist_product_event_processing_seconds` | Histogram | `event_type` | Consumer processing latency |
| `wishlist_product_event_lag_seconds` | Gauge/Histogram | `topic` | Event occurrence se process tak delay |

### Trace correlation

Event envelope ka `trace_id` log fields me preserve hoga. Agar OpenTelemetry add ho, consumer span name:

```text
wishlist.product_event.consume
```

---

## 🧪 Test Plan

### Unit tests

| Test case | Expected result |
|---|---|
| `ProductDeleted` event maps to `deleted` | Availability `deleted` |
| `ProductOutOfStock` event maps to `out_of_stock` | Availability `out_of_stock` |
| Unsupported event type | No DB call, no error |
| Missing `product_id` | Validation error |
| Duplicate event | Same final status, no duplicate side effect |
| Variant event with `variant_id` | Only matching variant updated |
| Product-level event without `variant_id` | All same product items updated |
| Mongo failure | Consumer returns error so retry can happen |

### Repository test example

```go
func TestUpdateProductAvailabilityUpdatesMatchingItems(t *testing.T) {
    repository := newFakeWishlistRepository(t)
    now := time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC)

    _ = repository.EnsureWishlist(context.Background(), "user_123", now)
    _ = repository.AddItemIfNotExists(context.Background(), "user_123", mustWishlistItem(t, "prod_123", "var_1", now), now)

    result, err := repository.UpdateProductAvailability(
        context.Background(),
        "prod_123",
        "var_1",
        domain.AvailabilityOutOfStock,
        now.Add(time.Minute),
    )
    if err != nil {
        t.Fatalf("UpdateProductAvailability returned error: %v", err)
    }
    if result.ModifiedCount != 1 {
        t.Fatalf("ModifiedCount = %d, want 1", result.ModifiedCount)
    }

    wishlist, _ := repository.FindByUserID(context.Background(), "user_123")
    item, _ := wishlist.FindItem("prod_123")
    if item.Availability != domain.AvailabilityOutOfStock {
        t.Fatalf("availability = %q, want out_of_stock", item.Availability)
    }
}
```

### Consumer test example

```go
func TestProductConsumerHandlesOutOfStockEvent(t *testing.T) {
    body := []byte(`{
      "event_id": "evt_123",
      "event_type": "ProductOutOfStock",
      "version": 1,
      "occurred_at": "2026-05-26T10:30:00Z",
      "producer": "product-service",
      "trace_id": "trace_123",
      "payload": {
        "product_id": "prod_123",
        "variant_id": "var_1"
      }
    }`)

    consumer := newTestProductConsumer(t)
    err := consumer.HandleMessage(context.Background(), body)
    if err != nil {
        t.Fatalf("HandleMessage returned error: %v", err)
    }

    if consumer.syncInput.ProductID != "prod_123" {
        t.Fatalf("ProductID = %q, want prod_123", consumer.syncInput.ProductID)
    }
    if consumer.syncInput.Availability != domain.AvailabilityOutOfStock {
        t.Fatalf("Availability = %q, want out_of_stock", consumer.syncInput.Availability)
    }
}
```

### Manual verification with MongoDB

1. Sample wishlist item insert karo:

```javascript
db.wishlists.insertOne({
  _id: "wish_123",
  user_id: "user_123",
  visibility: "private",
  items: [
    {
      product_id: "prod_123",
      variant_id: "var_1",
      added_at: new Date(),
      availability: "in_stock"
    }
  ],
  created_at: new Date(),
  updated_at: new Date()
})
```

2. Product out-of-stock update run karo:

```javascript
db.wishlists.updateMany(
  { "items.product_id": "prod_123", "items.variant_id": "var_1" },
  {
    $set: {
      "items.$[item].availability": "out_of_stock",
      "updated_at": new Date()
    }
  },
  {
    arrayFilters: [
      { "item.product_id": "prod_123", "item.variant_id": "var_1" }
    ]
  }
)
```

3. Result verify karo:

```javascript
db.wishlists.findOne(
  { user_id: "user_123" },
  { "items.product_id": 1, "items.variant_id": 1, "items.availability": 1 }
)
```

Expected:

```json
{
  "items": [
    {
      "product_id": "prod_123",
      "variant_id": "var_1",
      "availability": "out_of_stock"
    }
  ]
}
```

---

## 📦 External Libraries / Tools

### Used while creating this guide

| Tool | Used? | Notes |
|---|---:|---|
| External package install | ❌ | Is task ke requested output me sirf `task6.md` create hua |
| MongoDB Go Driver | Existing | Wishlist service `go.mod` me already `go.mongodb.org/mongo-driver/v2` present hai |
| Kafka/RabbitMQ client | Not installed | Future backend implementation me one backend choose hoga |

### Future backend implementation dependencies

| Library/Tool | What it is | Why used | Install |
|---|---|---|---|
| `go.mongodb.org/mongo-driver/v2` | Official MongoDB Go driver | `wishlists.items.availability` update karne ke liye | Already present |
| `github.com/segmentio/kafka-go` | Kafka Go client | `product.events` topic consume karne ke liye | `go get github.com/segmentio/kafka-go` |
| `github.com/rabbitmq/amqp091-go` | RabbitMQ Go client | RabbitMQ queue consume karne ke alternate ke liye | `go get github.com/rabbitmq/amqp091-go` |
| `mongosh` | MongoDB shell | Manual query verification ke liye | Install MongoDB tools / use Docker image |
| Docker Compose | Local infra runner | Kafka/RabbitMQ + MongoDB local run ke liye | Platform Foundation local stack |

### Kafka install/use example

```bash
cd backend/services/wishlist-service
go get github.com/segmentio/kafka-go
go mod tidy
```

Basic usage idea:

```go
reader := kafka.NewReader(kafka.ReaderConfig{
    Brokers: []string{"localhost:9092"},
    Topic:   "product.events",
    GroupID: "wishlist-service",
})
```

### RabbitMQ install/use example

```bash
cd backend/services/wishlist-service
go get github.com/rabbitmq/amqp091-go
go mod tidy
```

Basic usage idea:

```go
conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
if err != nil {
    return err
}
defer conn.Close()
```

> 🟣 **Decision tip:** Platform docs recommend Kafka for high-throughput event streaming. RabbitMQ acceptable hai agar project simpler routing/queues prefer kare.

---

## 🗂️ Recommended Implementation Folder Structure

Task 6 ke actual backend implementation me structure kuch aisa rahega:

```text
backend/
└── services/
    └── wishlist-service/
        ├── cmd/
        │   └── server/
        │       └── main.go
        ├── internal/
        │   ├── config/
        │   │   ├── config.go
        │   │   └── config_test.go
        │   ├── domain/
        │   │   └── wishlist.go
        │   ├── events/
        │   │   ├── product_consumer.go
        │   │   ├── product_consumer_test.go
        │   │   ├── product_event.go
        │   │   └── product_event_test.go
        │   ├── repository/
        │   │   ├── mongo_wishlist_repository.go
        │   │   └── mongo_wishlist_repository_test.go
        │   └── usecase/
        │       ├── wishlist_service.go
        │       └── wishlist_service_test.go
        └── go.mod
```

### File responsibilities

| File | Responsibility |
|---|---|
| `internal/events/product_event.go` | Product event envelope and payload structs |
| `internal/events/product_consumer.go` | Kafka/RabbitMQ message receive, decode, ack/retry behavior |
| `internal/usecase/wishlist_service.go` | `SyncProductAvailability` business rule |
| `internal/repository/mongo_wishlist_repository.go` | MongoDB `updateMany` with `arrayFilters` |
| `internal/config/config.go` | Event backend/topic/queue env config |
| `cmd/server/main.go` | Consumer lifecycle start/stop with HTTP server |

> 🟢 **Current requested output:** Sirf `TaskImplementation/Wishlist Service/task6.md` create kiya gaya. Backend files above implementation guide ke examples hain.

---

## 🛡️ Data Ownership Rules

| Rule | Meaning |
|---|---|
| Wishlist owns wishlist status snapshot | `items.availability` Wishlist DB me maintained snapshot hai |
| Product owns truth | Product Service actual product/inventory status ka source of truth hai |
| Events sync snapshot | Product Service event se Wishlist snapshot eventually consistent hota hai |
| No cross-service DB access | Wishlist Service Product DB ko direct query nahi karega |
| No silent deletion | Product delete event wishlist item remove nahi karega, only `deleted` mark karega |

---

## 🧠 Beginner-Friendly Mental Model

Think of wishlist availability as a cached label:

```text
Product Service = real truth
Wishlist Service = saved user's list + last known product status label
Product event = message saying label update kar do
```

Buyer ko wishlist open karte waqt fresh enough status milta hai:

```text
Before event:
prod_123 -> in_stock

After ProductOutOfStock event:
prod_123 -> out_of_stock

After ProductDeleted event:
prod_123 -> deleted
```

---

## 🔍 Edge Cases

| Edge case | Decision |
|---|---|
| Product event arrives but no wishlist has product | Success with `matched_count = 0` |
| Product-level deleted event has no `variant_id` | All wishlist items with same `product_id` update |
| Variant-level out-of-stock event has `variant_id` | Only same product + variant update |
| Wishlist item has empty `variant_id` | Product-level event can update it; variant-level event will not |
| Same event delivered twice | Same `$set`, final state same |
| Consumer crashes after DB update before ack | Message may replay; idempotent `$set` handles it |
| Product Service publishes unsupported event | Wishlist ignores and commits |
| Event payload malformed | DLQ for investigation |

---

## ✅ Acceptance Criteria

| Criteria | Status |
|---|---|
| `TaskImplementation/` folder exists | ✅ |
| `TaskImplementation/Wishlist Service/` folder exists | ✅ |
| `task6.md` created | ✅ |
| Task 6 exact scope documented | ✅ |
| Product deleted -> wishlist `deleted` rule documented | ✅ |
| Product out-of-stock -> wishlist `out_of_stock` rule documented | ✅ |
| No Task 7 price-drop implementation included | ✅ |
| No Task 8 analytics implementation included | ✅ |
| External tools/libraries explained | ✅ |
| Folder structure included | ✅ |
| Code examples included | ✅ |
| Mermaid diagrams included | ✅ |
| Beginner-friendly Hinglish explanation included | ✅ |

---

## 🚦 Final Implementation Checklist

```text
[x] Requirement identify kiya
[x] Existing wishlist availability model verify kiya
[x] Event contract define kiya
[x] Event type -> availability mapping define kiya
[x] Usecase/repository/consumer snippets add kiye
[x] MongoDB update strategy explain ki
[x] Retry/DLQ/idempotency explain ki
[x] External libraries/tools mention kiye
[x] Folder structure document ki
[x] Scope Task 6 tak limited rakha
```

---

## 🏁 Final Summary

Wishlist Service Task 6 ka final outcome:

```text
Input: ProductDeleted / ProductOutOfStock event from product.events
Processing: Wishlist Service consumer validates and maps event
DB update: Matching wishlists.items[].availability updated
Output: Buyer wishlist shows deleted/out_of_stock status
```

Is design se Wishlist Service Product Service ke saath loosely coupled rahega:

- Direct Product DB access nahi hoga.
- Wishlist item automatically remove nahi hoga.
- Duplicate events safe rahenge.
- Product unavailable hone par buyer ko clear status dikhega.

> ✅ **Task 6 complete:** Availability sync ka documentation-level implementation guide ready hai. Requested output ke according sirf `TaskImplementation/Wishlist Service/task6.md` create kiya gaya; backend service files intentionally modify nahi kiye gaye.
