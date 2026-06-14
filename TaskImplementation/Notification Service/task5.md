# 🔔 Notification Service - Task 5: Event Consumers

![Task](https://img.shields.io/badge/Task-Notification%20Service%20%235-blue)
![Priority](https://img.shields.io/badge/Priority-P1-orange)
![Dependency](https://img.shields.io/badge/Dependency-Message%20Queue-blueviolet)
![Broker](https://img.shields.io/badge/Local%20Default-RabbitMQ-FF6600)
![Pattern](https://img.shields.io/badge/Pattern-Async%20Consumers-informational)
![Status](https://img.shields.io/badge/Status-Documented-success)
![Guide](https://img.shields.io/badge/Language-Hinglish-yellow)

## 📌 Task Summary

| Field | Detail |
|---|---|
| Task Name | Event consumers |
| Requirement Source | `docs/01-micro-tasks.md` -> `Notification Service` -> Task 5 |
| Original Goal | Order/payment/user events consume karke notifications trigger karo |
| Dependency | Message queue |
| Priority | `P1` |
| Local Broker Decision | RabbitMQ, because Platform Foundation local stack me ye default documented hai |
| Consumed Event Groups | `order.events`, `payment.events`, `user.events` |
| Reused Notification Building Blocks | Channels/providers (Task 1), Mongo delivery storage (Task 2), templates (Task 3) |
| Deliverable | Async event-consumer implementation blueprint aur step-by-step guide |
| Implementation Type | Documentation + reference code/design only |

> **Simple Hinglish goal:** Order, Payment aur User Service apna business event queue par publish karenge. Notification Service un events ko asynchronously receive karega, event type ko correct notification template se map karega, duplicate event ko safely ignore karega, aur existing notification send pipeline ko trigger karega. Producer service ko notification provider ka wait nahi karna padega.

---

## ✅ Output Created

Request ke according repository me sirf ye documentation artifact add kiya gaya:

```text
TaskImplementation/
└── Notification Service/
    ├── task1.md                         # Existing: channels/provider abstraction
    ├── task2.md                         # Existing: MongoDB templates/delivery storage
    ├── task3.md                         # Existing: template rendering
    ├── task4.md                         # Existing: OTP delivery flow
    └── task5.md                         # Current: async event consumer guide
```

> 🟢 **Important:** Neeche ka Go code, RabbitMQ topology, event schemas, Mongo index aur test plan ek implementable reference blueprint hai. Is output me `backend/services/notification-service/`, queue configuration, dependencies, migrations, producers, providers ya running integration physically create/modify nahi kiye gaye.

---

## 📚 Project Requirements Se Kya Samjha

| Source | Task 5 ke liye decision |
|---|---|
| `docs/01-micro-tasks.md` | Exact Notification task: order/payment/user events consume karke notifications trigger karna; priority `P1` |
| `docs/01-micro-tasks.md` | Order Service `OrderCreated`, `OrderPaid`, `OrderCancelled`, `OrderDelivered` emit karega |
| `docs/01-micro-tasks.md` | User Service `UserCreated`, `SellerApproved`, `AddressUpdated` publish karega aur Notification consumer hoga |
| `docs/02-system-architecture.md` | Async processing ke liye Kafka/RabbitMQ; at-least-once consumers ko idempotent hona chahiye |
| `docs/03-folder-structure.md` | Notification event consumer ka intended location `internal/events/consumer.go` hai |
| `docs/04-microservice-design.md` | Notifications event consume karke async send hongi; order status aur payment success/failure supported responsibilities hain |
| `docs/11-devops-external-services.md` | Topics/queues me `user.events`, `order.events`, `payment.events`; standard event envelope fields defined hain |
| `TaskImplementation/Platform Foundation/task4.md` | Local beginner-friendly stack me RabbitMQ default hai; Kafka future/high-throughput alternate hai |
| `TaskImplementation/Notification Service/task1.md` | Consumer trigger ko provider-neutral channel contract reuse karna hai |
| `TaskImplementation/Notification Service/task2.md` | `notification_deliveries` message attempt/status store karega |
| `TaskImplementation/Notification Service/task3.md` | Template renderer final channel-specific content banayega |
| `TaskImplementation/Notification Service/task4.md` | OTP Auth-driven internal operation hai; Task 5 event routing OTP flow ko replace nahi karta |

### Payment Event Naming Clarification

Project docs payment **success/failure notification** require karti hain, lekin payment event type names ko order events jaisa exact enumerate nahi karti. Is guide me versioned consumer contract ke liye `PaymentSucceeded` aur `PaymentFailed` recommended names use kiye gaye hain. Producer implementation ke time Payment Service aur Notification Service ko same contract freeze karna hoga.

---

## 🚧 Scope Boundary

Task 5 ka kaam **queue events ko receive, validate, deduplicate, map aur notification trigger karna** hai. Ye complete retry platform, user preference engine ya analytics pipeline nahi hai.

### Included In Task 5

| Included | Kyon |
|---|---|
| RabbitMQ consumer topology for order/payment/user events | Message queue dependency ko concrete local implementation milta hai |
| Common versioned event envelope decode/validate karna | Producers aur consumer ka stable contract banta hai |
| Supported event type -> template/channel mapping | Event se kaunsi notification trigger hogi clear hota hai |
| Existing render/send pipeline invoke karna | Consumer provider aur template logic duplicate nahi karta |
| Event-level idempotency strategy | At-least-once broker redelivery me duplicate user messages prevent hote hain |
| Ack/nack decision rules | Broker message lifecycle predictable hota hai |
| Consumer unit/integration test blueprint | Event behaviour confidently verify ho sakta hai |
| Minimal queue metrics/logging guidance | Operations ko consumer health dikhti hai |

### Explicitly Not Included

| Deferred / Owned Elsewhere | Reason |
|---|---|
| Order, Payment ya User Service me events publish karne ka code | Respective producer services ki responsibility hai |
| Kafka consumer implementation | RabbitMQ local default select hua hai; Kafka alternate only documented hai |
| OTP request processing | Notification Service - Task 4 synchronous Auth-driven flow hai |
| Automatic exponential retry queues, retry schedule aur DLQ wiring | Notification Service - Task 6 |
| Marketing opt-in/opt-out ya channel preference enforcement | Notification Service - Task 7 |
| Delivered/opened/provider webhook analytics | Notification Service - Task 8 |
| Price-drop/wishlist event consumption | Requirement is task me sirf order/payment/user events bolti hai |
| Actual backend source, migration, compose ya dependency changes | Requested output sirf `task5.md` documentation artifact hai |

> 🔴 **Boundary rule:** Task 5 consumer message receive karke a single intended notification trigger tak responsibility rakhta hai. Provider failure ko sophisticated retry/DLQ pipeline me move karna Task 6 me implement hoga.

---

## 🔗 Existing Contracts Jo Reuse Honge

| Previous Task | Task 5 Me Reuse |
|---|---|
| Task 1: channel/provider abstraction | Rule selected `email`, `sms` ya `push` ko configured provider tak bhejne ke liye |
| Task 2: MongoDB persistence | Triggered delivery aur source event id ko trace/deduplicate karne ke liye |
| Task 3: template engine | Event payload variables ko approved template se final content me render karne ke liye |
| Task 4: safe delivery patterns | Sensitive payload/error logging aur provider acceptance semantics ke patterns ke liye |

### Consumer Kya Nahi Karega

```text
Consumer -> provider SDK directly                 # ❌ use send pipeline instead
Consumer -> producer database query               # ❌ service ownership break hota hai
Consumer -> OTP generation or verification        # ❌ Auth/Task 4 concern
Consumer -> marketing preference implementation   # ❌ Task 7 concern
Consumer -> retry queue/DLQ declaration           # ❌ Task 6 concern
```

---

## 🧭 High-Level Event Flow

```mermaid
sequenceDiagram
    autonumber
    participant O as Order / Payment / User Service
    participant MQ as RabbitMQ Event Exchange
    participant C as Notification Consumer
    participant I as Idempotency Store
    participant S as SendNotification Use Case
    participant T as Template Renderer
    participant P as Channel Provider
    participant D as notification_deliveries

    O->>MQ: Publish versioned domain event
    MQ->>C: Deliver message
    C->>C: Decode envelope + validate event
    C->>I: Claim event/rule/channel key
    alt already processed
        I-->>C: Duplicate
        C->>MQ: Ack without sending again
    else first processing
        I-->>C: Claimed
        C->>S: Trigger(template, channel, variables)
        S->>T: Render existing active template
        T-->>S: Subject/body
        S->>P: Send notification
        P-->>S: Accepted / failure
        S->>D: Store delivery with source_event_id
        S-->>C: Trigger result
        C->>MQ: Ack success or Nack transient failure
    end
```

### Flow Ka Meaning

| Stage | Hinglish Explanation |
|---|---|
| Publish | Producer apna domain decision publish karta hai; provider ka kaam producer me nahi hota |
| Consume | Notification consumer sirf subscribed business events uthata hai |
| Validate | Invalid version/payload ko send pipeline tak jaane se pehle roka jata hai |
| Claim | Same `event_id` redeliver ho to user ko duplicate notification nahi bhejte |
| Map | Event type se template key, default channel aur render variables decide hote hain |
| Trigger | Existing template/provider/delivery flow actual message dispatch karta hai |
| Ack | Successful ya safely ignored event broker se remove hota hai |

---

## 🏗️ Task 5 Architecture

```mermaid
flowchart LR
    Order[Order Service] -->|Order events| OX[order.events exchange]
    Payment[Payment Service] -->|Payment events| PX[payment.events exchange]
    User[User Service] -->|User events| UX[user.events exchange]

    OX --> OQ[notification.order.events.v1 queue]
    PX --> PQ[notification.payment.events.v1 queue]
    UX --> UQ[notification.user.events.v1 queue]

    OQ --> Consume[RabbitMQ Consumers<br/>Task 5]
    PQ --> Consume
    UQ --> Consume

    Consume --> Validate[Envelope Validation<br/>+ Rule Router]
    Validate --> Dedupe[(MongoDB<br/>idempotency key)]
    Validate --> Trigger[SendNotification Trigger]
    Trigger --> Render[Task 3 Renderer]
    Trigger --> Provider[Task 1 Provider Registry]
    Trigger --> Delivery[(Task 2<br/>notification_deliveries)]

    classDef task5 fill:#e8f5e9,stroke:#2e7d32,color:#1b5e20;
    classDef dependency fill:#e3f2fd,stroke:#1565c0,color:#0d47a1;
    classDef producer fill:#fff8e1,stroke:#f9a825,color:#6d4c41;
    class Consume,Validate,Dedupe task5;
    class Trigger,Render,Provider,Delivery dependency;
    class Order,Payment,User,OX,PX,UX,OQ,PQ,UQ producer;
```

---

## 📁 Clean Implementation Folder Structure

### Actual Documentation Output

```text
TaskImplementation/
└── Notification Service/
    ├── task1.md
    ├── task2.md
    ├── task3.md
    ├── task4.md
    └── task5.md
```

### Reference Source Layout For Implementing Only Task 5

Ye project ke documented Go service layout me Task 5 ko fit karta hai. Neeche listed backend files is documentation output me physically create nahi kiye ja rahe.

```text
backend/
└── services/
    └── notification-service/
        ├── cmd/
        │   └── server/
        │       └── main.go                              # Future wiring: consumer startup/shutdown
        └── internal/
            ├── domain/
            │   ├── channel.go                           # Task 1 dependency
            │   ├── delivery.go                          # Task 2 dependency
            │   └── event_trigger.go                     # Task 5: trigger request + idempotency key
            ├── repository/
            │   ├── notification_repository.go           # Task 2/5: delivery/idempotency contracts
            │   └── mongo_notification_repository.go     # Task 2/5: source event claim/storage
            ├── usecase/
            │   ├── render_template.go                   # Task 3 dependency
            │   └── send_notification.go                 # Task 5 invokes generic send operation
            └── events/
                ├── envelope.go                          # Task 5: common event schema validation
                ├── payloads.go                          # Task 5: order/payment/user event payloads
                ├── rules.go                             # Task 5: event -> notification mapping
                ├── handler.go                           # Task 5: idempotent trigger orchestration
                ├── rabbitmq_consumer.go                 # Task 5: delivery + ack/nack adapter
                └── handler_test.go                      # Task 5: consumer behaviour tests
```

### Intentionally Absent From Task 5

```text
internal/retry/                                        # Task 6
internal/events/retry_consumer.go                      # Task 6
internal/domain/preference.go                          # Task 7
internal/analytics/                                    # Task 8
```

---

## 🪜 Step-by-Step Implementation

## Step 1: Broker Choice Aur Event Boundary Freeze Karo

Project docs Kafka aur RabbitMQ dono allow karti hain. Platform Foundation local setup RabbitMQ ko default choose karta hai, isliye Task 5 blueprint **RabbitMQ topic exchanges aur durable consumer queues** use karega.

| Decision | Reason |
|---|---|
| RabbitMQ local baseline | Management UI ke through beginner exchange, queue aur binding inspect kar sakta hai |
| Producer-specific exchanges | `order.events`, `payment.events`, `user.events` existing documented event groups match karte hain |
| Notification-owned queues | Har consumer apna backlog own karta hai; Analytics ya other consumers se conflict nahi |
| Durable queues + persistent published messages | Service restart par accepted queued work loss ka chance reduce hota hai |
| Kafka alternate only | Kafka ka alag adapter later ho sakta hai; is task me do brokers implement nahi karenge |

### Recommended RabbitMQ Topology

| Exchange | Exchange Type | Notification Queue | Bound Event Types |
|---|---|---|---|
| `order.events` | `topic`, durable | `notification.order.events.v1` | `OrderCreated`, `OrderPaid`, `OrderCancelled`, `OrderDelivered` |
| `payment.events` | `topic`, durable | `notification.payment.events.v1` | `PaymentSucceeded`, `PaymentFailed` |
| `user.events` | `topic`, durable | `notification.user.events.v1` | `UserCreated`, `SellerApproved`, `AddressUpdated` |

> 🟡 `payment.events` event names is guide ka recommended version-1 consumer contract hain because docs outcome define karti hain, exact names nahi. Producers start hone se pehle contract test/protobuf or JSON schema se naming freeze karo.

### `notification.commands` Kyon Consume Nahi Kar Rahe?

`notification.commands` project infrastructure me possible direct notification command queue hai. Task 5 ka exact wording **order/payment/user events** bolta hai, isliye is guide ka consumer domain event exchanges tak limited hai. OTP ka internal request Task 4 flow me hi rahega.

---

## Step 2: Versioned Event Envelope Define Karo

Har producer payload alag ho sakta hai, lekin metadata same hona chahiye. `docs/11-devops-external-services.md` ka event envelope reuse karke consumer ko stable parsing aur tracing milti hai.

### Event Envelope Example

```json
{
  "event_id": "evt_order_paid_123",
  "event_type": "OrderPaid",
  "version": 1,
  "occurred_at": "2026-05-27T14:00:00Z",
  "producer": "order-service",
  "trace_id": "trace_checkout_123",
  "payload": {
    "order_id": "order_123",
    "user_id": "user_123",
    "email": "riya@example.com",
    "name": "Riya",
    "status": "Paid"
  }
}
```

### Reference Code: `internal/events/envelope.go`

```go
package events

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var ErrInvalidEvent = errors.New("invalid notification source event")

type Envelope struct {
	EventID    string          `json:"event_id"`
	EventType  string          `json:"event_type"`
	Version    int             `json:"version"`
	OccurredAt time.Time       `json:"occurred_at"`
	Producer   string          `json:"producer"`
	TraceID    string          `json:"trace_id"`
	Payload    json.RawMessage `json:"payload"`
}

func DecodeEnvelope(body []byte) (Envelope, error) {
	var envelope Envelope
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&envelope); err != nil {
		return Envelope{}, fmt.Errorf("%w: decode envelope: %v", ErrInvalidEvent, err)
	}
	if envelope.EventID == "" || envelope.EventType == "" ||
		envelope.Producer == "" || envelope.OccurredAt.IsZero() {
		return Envelope{}, fmt.Errorf("%w: required metadata missing", ErrInvalidEvent)
	}
	if envelope.Version != 1 {
		return Envelope{}, fmt.Errorf("%w: unsupported version %d", ErrInvalidEvent, envelope.Version)
	}
	if len(envelope.Payload) == 0 {
		return Envelope{}, fmt.Errorf("%w: payload missing", ErrInvalidEvent)
	}
	return envelope, nil
}
```

### Har Field Ka Purpose

| Field | Consumer Me Use |
|---|---|
| `event_id` | Idempotency key ka source; duplicate delivery identify karta hai |
| `event_type` | Notification rule/template select karta hai |
| `version` | Breaking payload change ko safely reject ya route karne deta hai |
| `occurred_at` | Audit aur delayed-message monitoring me useful hai |
| `producer` | Wrong exchange/producer input reject aur logs correlate karne me help |
| `trace_id` | Checkout/payment se notification dispatch tak distributed trace link |
| `payload` | Template variables aur intended recipient ke liye event-specific data |

> 🔐 Payload me minimum required contact/display fields publish karo. Payment tokens, full address, OTP, passwords ya provider secrets event me kabhi mat bhejo.

---

## Step 3: Event Payload Contracts Banao

Consumer producer database read nahi karega. Isliye event me notification trigger ke liye required minimal snapshot hona chahiye.

### Reference Code: `internal/events/payloads.go`

```go
package events

type OrderPayload struct {
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Status  string `json:"status"`
}

type PaymentPayload struct {
	OrderID       string `json:"order_id"`
	UserID        string `json:"user_id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	PaymentStatus string `json:"payment_status"`
	Amount        string `json:"amount"`
}

type UserPayload struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}
```

### Payload Design Rules

| Rule | Hinglish Reason |
|---|---|
| `user_id` include karo | Delivery history aur preference integration later user se correlate ho sake |
| Recipient snapshot include karo | Consumer ko User DB directly query nahi karna padta |
| Order/payment display values include karo | Template ko business result ka stable snapshot milta hai |
| Sensitive payment/credential values omit karo | Queue aur logs me unnecessary secret exposure avoid hoti hai |
| Payload version envelope se bind karo | Future schema evolution controlled rahegi |

---

## Step 4: Event-To-Notification Rules Define Karo

Consumer ko arbitrary content construct nahi karna chahiye. Ek allowlisted rule catalog decide karega kaunsa event kaunsa existing/configured template trigger kare.

### Version 1 Rule Catalogue

| Source Event | Producer | Template Key | Default Channel | Template Variables | User Experience |
|---|---|---|---|---|---|
| `OrderCreated` | Order | `order_status_update` | `email` | `name`, `order_id`, `status=Created` | Order received confirmation |
| `OrderPaid` | Order | `order_status_update` | `email` | `name`, `order_id`, `status=Paid` | Paid order state update |
| `OrderCancelled` | Order | `order_status_update` | `email` | `name`, `order_id`, `status=Cancelled` | Cancellation information |
| `OrderDelivered` | Order | `order_status_update` | `email` | `name`, `order_id`, `status=Delivered` | Delivery completion notice |
| `PaymentSucceeded` | Payment | `payment_status_update` | `email` | `name`, `order_id`, `payment_status=Successful`, `amount` | Payment confirmation |
| `PaymentFailed` | Payment | `payment_status_update` | `email` | `name`, `order_id`, `payment_status=Failed`, `amount` | Payment failure information |
| `UserCreated` | User | `welcome_user` | `email` | `name` | Welcome/account creation message |
| `SellerApproved` | User | `seller_approved` | `email` | `name` | Seller approval update |
| `AddressUpdated` | User | `address_updated_security_notice` | `email` | `name` | Account-change security notice |

### Template Dependency Note

`order_status_update` aur `payment_status_update` Task 3 template families reuse karte hain. User-event keys ke active email templates ko template catalog me provision karna hoga before corresponding consumer bindings production me enable kiye jayen. Task 5 consumer un templates ko **manage** nahi karta; sirf selected key se trigger karta hai.

### Task 7 Se Boundary

Task 5 me upar ke examples transactional/account notifications hain aur predictable default email channel use karte hain. Marketing opt-in, user-selected channel preference, promotional suppression ya multi-channel fan-out ko yahan implement nahi karna; wo Task 7 ka kaam hai.

---

## Step 5: Notification Trigger Contract Rakho

Event layer ko template renderer, MongoDB aur provider ke internals nahi pata hone chahiye. Consumer ek small trigger request bana kar generic send operation ko call karega.

### Reference Code: `internal/domain/event_trigger.go`

```go
package domain

type EventNotificationTrigger struct {
	IdempotencyKey string
	SourceEventID  string
	SourceType     string
	TraceID        string
	UserID         string
	Channel        Channel
	Recipient      string
	TemplateKey    string
	Variables      map[string]string
}
```

### Reference Interface: `internal/events/handler.go`

```go
package events

import (
	"context"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
)

type TriggerSender interface {
	TriggerFromEvent(ctx context.Context, req domain.EventNotificationTrigger) error
}

type EventClaimStore interface {
	Claim(ctx context.Context, idempotencyKey string) (claimed bool, err error)
	Complete(ctx context.Context, idempotencyKey string) error
	Release(ctx context.Context, idempotencyKey string) error
}
```

### Separation Ka Fayda

| Layer | Responsibility |
|---|---|
| RabbitMQ adapter | Broker deliveries read aur ack/nack kare |
| Event handler | Envelope validate, rule choose, deduplicate, trigger call kare |
| Trigger sender | Task 1-3 reuse karke render/provider/delivery operation chalaye |
| Repository | Claim/delivery state safely store kare |
| Producer service | Business state decide karke event publish kare |

---

## Step 6: Idempotency Implement Karo

RabbitMQ delivery at-least-once ho sakti hai: consumer send karne ke baad crash ho jaye, acknowledgement miss ho sakta hai aur same message phir receive hoga. `event_id` ko ignore karna duplicate customer emails create kar sakta hai.

### Deterministic Key

```text
source_event_id : template_key : channel

evt_order_paid_123:order_status_update:email
```

Ek event future me multiple valid notifications/channels trigger kare to har intended send ki key separate rahegi, lekin same intended send duplicate nahi hoga.

### MongoDB Storage Guidance

Task 2 ki `notification_deliveries` document me source fields add karke unique index use kiya ja sakta hai:

```javascript
{
  _id: "delivery_evt_order_paid_123_email",
  source_event_id: "evt_order_paid_123",
  source_event_type: "OrderPaid",
  idempotency_key: "evt_order_paid_123:order_status_update:email",
  trace_id: "trace_checkout_123",
  user_id: "user_123",
  channel: "email",
  template_key: "order_status_update",
  status: "processing",
  created_at: ISODate("2026-05-27T14:00:01Z"),
  updated_at: ISODate("2026-05-27T14:00:01Z")
}
```

```javascript
db.notification_deliveries.createIndex(
  { idempotency_key: 1 },
  { unique: true, name: "uniq_event_notification_key" }
)
```

### Status Behaviour

| Situation | Action |
|---|---|
| New unique key | Insert `processing`, then trigger notification |
| Existing completed/accepted key | Treat as already handled; `Ack` without second send |
| Rendering/config validation failure | Mark failed; reject/alert because same input will not self-heal |
| Temporary provider/broker dependency failure | Keep recoverable state and `Nack`/requeue baseline; full policy Task 6 |
| Provider accepts send but completion save fails | Use deterministic provider idempotency key where supported; alert because exactly-once external delivery cannot be assumed |

> 🟠 **Reality check:** Database uniqueness prevents the consumer from intentionally starting a duplicate send. External provider acceptance plus a process crash is still a distributed failure case. Provider-side idempotency key or a Task 6 delivery-outbox/reconciliation policy is needed for stronger guarantees.

---

## Step 7: Rule Router Se Trigger Request Banao

Rule router valid event ko typed payload me decode karta hai aur template variables create karta hai. Unsupported event types ko safe no-op treat kiya ja sakta hai, kyunki ek exchange me producer ke extra events ho sakte hain.

### Reference Code: `internal/events/rules.go`

```go
package events

import (
	"encoding/json"
	"fmt"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
)

func BuildTrigger(envelope Envelope) (domain.EventNotificationTrigger, bool, error) {
	switch envelope.EventType {
	case "OrderCreated", "OrderPaid", "OrderCancelled", "OrderDelivered":
		var payload OrderPayload
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return domain.EventNotificationTrigger{}, false, fmt.Errorf("%w: order payload", ErrInvalidEvent)
		}
		status := map[string]string{
			"OrderCreated":   "Created",
			"OrderPaid":      "Paid",
			"OrderCancelled": "Cancelled",
			"OrderDelivered": "Delivered",
		}[envelope.EventType]
		return newEmailTrigger(envelope, payload.UserID, payload.Email,
			"order_status_update", map[string]string{
				"name": payload.Name, "order_id": payload.OrderID, "status": status,
			}), true, nil

	case "PaymentSucceeded", "PaymentFailed":
		var payload PaymentPayload
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return domain.EventNotificationTrigger{}, false, fmt.Errorf("%w: payment payload", ErrInvalidEvent)
		}
		status := "Failed"
		if envelope.EventType == "PaymentSucceeded" {
			status = "Successful"
		}
		return newEmailTrigger(envelope, payload.UserID, payload.Email,
			"payment_status_update", map[string]string{
				"name": payload.Name, "order_id": payload.OrderID,
				"payment_status": status, "amount": payload.Amount,
			}), true, nil

	case "UserCreated", "SellerApproved", "AddressUpdated":
		var payload UserPayload
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return domain.EventNotificationTrigger{}, false, fmt.Errorf("%w: user payload", ErrInvalidEvent)
		}
		templateKey := map[string]string{
			"UserCreated":   "welcome_user",
			"SellerApproved": "seller_approved",
			"AddressUpdated": "address_updated_security_notice",
		}[envelope.EventType]
		return newEmailTrigger(envelope, payload.UserID, payload.Email,
			templateKey, map[string]string{"name": payload.Name}), true, nil
	default:
		return domain.EventNotificationTrigger{}, false, nil
	}
}

func newEmailTrigger(
	envelope Envelope,
	userID string,
	email string,
	templateKey string,
	variables map[string]string,
) domain.EventNotificationTrigger {
	channel := domain.ChannelEmail
	return domain.EventNotificationTrigger{
		IdempotencyKey: envelope.EventID + ":" + templateKey + ":" + string(channel),
		SourceEventID:  envelope.EventID,
		SourceType:     envelope.EventType,
		TraceID:        envelope.TraceID,
		UserID:         userID,
		Channel:        channel,
		Recipient:      email,
		TemplateKey:    templateKey,
		Variables:      variables,
	}
}
```

### Router Validation Add Karo

Code example mapping intent show karta hai. Production implementation me trigger return karne se pehle ye validation mandatory hogi:

| Validate | Failure Result |
|---|---|
| Expected `producer` event group se match kare | Invalid event; send mat karo |
| `user_id`, email aur required IDs empty na hon | Invalid event; send mat karo |
| Email Task 1 recipient validator pass kare | Invalid recipient; send mat karo |
| Template key active/allowed ho | Trigger failure surface karo |
| Variables me secrets/raw payment data absent ho | Contract/test failure |

---

## Step 8: Handler Me Claim, Trigger Aur Completion Orchestrate Karo

Handler broker-specific nahi hoga. Isliye unit tests RabbitMQ start kiye bina business consumer behaviour test kar sakte hain.

### Reference Code: `internal/events/handler.go`

```go
package events

import (
	"context"
	"errors"
	"fmt"
)

var ErrTemporaryProcessing = errors.New("temporary notification event processing failure")

type Outcome string

const (
	OutcomeProcessed   Outcome = "processed"
	OutcomeDuplicate   Outcome = "duplicate"
	OutcomeUnsupported Outcome = "unsupported"
)

type Handler struct {
	claims EventClaimStore
	sender TriggerSender
}

func (h *Handler) Handle(ctx context.Context, body []byte) (Outcome, error) {
	envelope, err := DecodeEnvelope(body)
	if err != nil {
		return "", err
	}

	trigger, supported, err := BuildTrigger(envelope)
	if err != nil {
		return "", err
	}
	if !supported {
		return OutcomeUnsupported, nil
	}

	claimed, err := h.claims.Claim(ctx, trigger.IdempotencyKey)
	if err != nil {
		return "", fmt.Errorf("%w: claim event: %v", ErrTemporaryProcessing, err)
	}
	if !claimed {
		return OutcomeDuplicate, nil
	}

	if err := h.sender.TriggerFromEvent(ctx, trigger); err != nil {
		_ = h.claims.Release(ctx, trigger.IdempotencyKey)
		return "", fmt.Errorf("%w: trigger send: %v", ErrTemporaryProcessing, err)
	}
	if err := h.claims.Complete(ctx, trigger.IdempotencyKey); err != nil {
		return "", fmt.Errorf("%w: complete event: %v", ErrTemporaryProcessing, err)
	}

	return OutcomeProcessed, nil
}
```

### Handler Build Explanation

| Code Part | Hinglish Explanation |
|---|---|
| `DecodeEnvelope` | Malformed ya unsupported schema event early stop hota hai |
| `BuildTrigger` | Sirf allowlisted event notification request me badalta hai |
| `Claim` | Same intended message second time send hone se rokta hai |
| `TriggerFromEvent` | Existing rendering/provider/storage operation use hota hai |
| `Complete` | Consumer audit karta hai ki event handling successful complete hui |
| Unsupported -> success outcome | Non-notification event binding accidentally broad ho tab endless redelivery avoid hoti hai |

> 🟡 `Release` reference behaviour transient failure retry allow karta hai. Exact retry attempt counts, delays aur final dead-letter flow Task 6 me design/implement karne hain.

---

## Step 9: RabbitMQ Consumer Adapter Add Karo

RabbitMQ adapter queue se delivery uthata hai aur handler outcome ke base par acknowledgement decide karta hai. Business mapping adapter ke andar nahi honi chahiye.

### Reference Code: `internal/events/rabbitmq_consumer.go`

```go
package events

import (
	"context"
	"errors"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitConsumer struct {
	channel *amqp.Channel
	handler *Handler
	logger  *slog.Logger
}

func (c *RabbitConsumer) Consume(ctx context.Context, queue string) error {
	deliveries, err := c.channel.ConsumeWithContext(
		ctx, queue, "", false, false, false, false, nil,
	)
	if err != nil {
		return err
	}

	for delivery := range deliveries {
		outcome, handleErr := c.handler.Handle(ctx, delivery.Body)
		switch {
		case handleErr == nil:
			c.logger.Info("notification event handled", "outcome", outcome, "queue", queue)
			_ = delivery.Ack(false)
		case errors.Is(handleErr, ErrInvalidEvent):
			c.logger.Error("invalid notification event rejected", "queue", queue)
			_ = delivery.Reject(false)
		default:
			c.logger.Warn("notification event processing unavailable", "queue", queue)
			_ = delivery.Nack(false, true)
		}
	}
	return ctx.Err()
}
```

### Ack/Nack Rules

| Result | Broker Action In Task 5 Baseline | Reason |
|---|---|---|
| Notification processed | `Ack` | Work successfully done |
| Duplicate idempotency key | `Ack` | User ko already intended message mil chuka/started hai |
| Unsupported but valid event | `Ack` | Event is consumer rule ka part nahi, retry se result change nahi hoga |
| Malformed/version-invalid event | `Reject(requeue=false)` + error metric | Bad message repeat process nahi karna |
| Temporary repository/provider problem | `Nack(requeue=true)` | Short interruption ke baad processing possible ho sakti hai |

> 🔴 **Production note:** `Reject(false)` ko dead-letter exchange se capture karna aur unbounded immediate requeue ko backoff retry me convert karna required reliability work hai, lekin wo explicitly Task 6 ka implementation scope hai. Task 5 sirf decision points expose karta hai.

### Logging Rule

Logs me `event_id`, `event_type`, `trace_id`, queue, outcome aur sanitized error category record kar sakte hain. Raw payload, full recipient, rendered body, payment details ya provider secret log nahi karne hain.

---

## Step 10: Queue Declaration Aur Bindings Wire Karo

Local environment me Notification Service apni queues declare/bind karega. Producer exchanges ko infrastructure ownership policy ke according bootstrap ya producer declare kar sakta hai; names contract-compatible rehne chahiye.

### Reference Topology Function

```go
func DeclareNotificationEventQueues(ch *amqp.Channel) error {
	bindings := map[string][]string{
		"order.events":   {"OrderCreated", "OrderPaid", "OrderCancelled", "OrderDelivered"},
		"payment.events": {"PaymentSucceeded", "PaymentFailed"},
		"user.events":    {"UserCreated", "SellerApproved", "AddressUpdated"},
	}
	queues := map[string]string{
		"order.events":   "notification.order.events.v1",
		"payment.events": "notification.payment.events.v1",
		"user.events":    "notification.user.events.v1",
	}

	for exchange, eventTypes := range bindings {
		if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
			return err
		}
		queue, err := ch.QueueDeclare(queues[exchange], true, false, false, false, nil)
		if err != nil {
			return err
		}
		for _, eventType := range eventTypes {
			if err := ch.QueueBind(queue.Name, eventType, exchange, false, nil); err != nil {
				return err
			}
		}
	}
	return nil
}
```

### Configuration Shape

```dotenv
# Task 5 message consumption
NOTIFICATION_RABBITMQ_URL=amqp://ecommerce:ecommerce_password@rabbitmq:5672/ecommerce
NOTIFICATION_ORDER_EVENTS_QUEUE=notification.order.events.v1
NOTIFICATION_PAYMENT_EVENTS_QUEUE=notification.payment.events.v1
NOTIFICATION_USER_EVENTS_QUEUE=notification.user.events.v1
NOTIFICATION_CONSUMER_PREFETCH=10

# Existing Notification dependencies
NOTIFICATION_MONGO_URI=mongodb://mongo:27017/notification_db
NOTIFICATION_MONGO_DATABASE=notification_db
```

### Runtime Wiring Order

| Order | Startup Work | Failure Behaviour |
|---:|---|---|
| 1 | Config validate karo | Missing broker URI/queue value par service start fail |
| 2 | Mongo repository connect karo | Idempotency absent ho to consumer start mat karo |
| 3 | Renderer/provider/send pipeline wire karo | Trigger target unavailable ho to start fail |
| 4 | RabbitMQ connection/channel create karo | Broker unavailable par health not-ready |
| 5 | Exchange/queues/bindings declare karo | Wrong topology silently consume na kare |
| 6 | Prefetch set karke three consumers start karo | Graceful shutdown par in-flight context cancel |

---

## Step 11: Graceful Shutdown Aur Observability Add Karo

Event consumer long-running component hai. Deployment stop ke time in-flight message ko half-process karke lost acknowledgement nahi chhodna chahiye.

### Shutdown Behaviour

| Situation | Behaviour |
|---|---|
| Service gets termination signal | New consume loop context cancel karo |
| In-flight handler already executing | Short shutdown deadline tak finish allow karo |
| Handler finish nahi kar pata | Broker unacked delivery redeliver kar sakega |
| RabbitMQ connection closes | Health not-ready mark; reconnect policy deployment/runtime layer me handle karo |

### Minimum Metrics

| Metric | Labels | Kyon Useful |
|---|---|---|
| `notification_events_consumed_total` | `event_type`, `outcome` | Processed/duplicate/invalid distribution |
| `notification_event_processing_seconds` | `event_type` | Slow rendering/provider path spot karna |
| `notification_event_trigger_failures_total` | `event_type`, `category` | Dependencies failure detect karna |
| `notification_event_redeliveries_total` | `queue` | Retry/DLQ requirement quantify karna |
| RabbitMQ queue depth | `queue` | Consumer backlog/incident signal |

### Safe Structured Log Example

```json
{
  "level": "INFO",
  "message": "notification event handled",
  "event_id": "evt_order_paid_123",
  "event_type": "OrderPaid",
  "trace_id": "trace_checkout_123",
  "outcome": "processed",
  "channel": "email",
  "template_key": "order_status_update"
}
```

---

## Step 12: Tests Se Consumer Behaviour Lock Karo

### Unit Test Matrix

| Test | Expected Result |
|---|---|
| Valid `OrderPaid` v1 event | `order_status_update` email trigger exactly once |
| Valid `PaymentFailed` event | `payment_status_update` with failed variables |
| Valid `UserCreated` event | `welcome_user` trigger |
| Duplicate same `event_id`/template/channel | Sender second time call nahi hota; outcome duplicate |
| Unknown valid event type | Sender call nahi hota; outcome unsupported and broker ack |
| Invalid JSON / missing event id | Invalid event; no claim/send |
| Unsupported event version | Invalid event; no send |
| Missing required order/payment payload value | Invalid event; no provider invocation |
| Claim repository unavailable | Temporary error; consumer requeue decision |
| Sender temporary failure | Temporary error; acknowledgement nahi |
| Logs/records inspection | Raw recipient/full payload/payment secret absent |

### Reference Unit Test: Duplicate Event

```go
func TestHandlerDoesNotSendDuplicateEvent(t *testing.T) {
	claims := &fakeClaims{claimed: false}
	sender := &fakeSender{}
	handler := &Handler{claims: claims, sender: sender}

	body := []byte(`{
		"event_id":"evt_order_paid_123",
		"event_type":"OrderPaid",
		"version":1,
		"occurred_at":"2026-05-27T14:00:00Z",
		"producer":"order-service",
		"trace_id":"trace_123",
		"payload":{"order_id":"order_123","user_id":"user_123","email":"riya@example.com","name":"Riya"}
	}`)

	outcome, err := handler.Handle(context.Background(), body)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if outcome != OutcomeDuplicate {
		t.Fatalf("outcome = %q, want duplicate", outcome)
	}
	if sender.calls != 0 {
		t.Fatalf("sender calls = %d, want 0", sender.calls)
	}
}
```

### RabbitMQ Integration Test Plan

| Step | Verification |
|---:|---|
| 1 | Local RabbitMQ aur MongoDB start karo |
| 2 | Consumer with test provider/template repository start karo |
| 3 | `order.events` exchange par valid `OrderPaid` persistent message publish karo |
| 4 | `notification_deliveries` me source `event_id` aur selected template record check karo |
| 5 | Same message dobara publish karo |
| 6 | Sirf one intended delivery exists aur duplicate outcome metric increment hua verify karo |
| 7 | Invalid version publish karo |
| 8 | Provider call nahi hua aur invalid event metric/log generated verify karo |

---

## 🧰 External Libraries And Tools

### Runtime And Development Dependencies

| Library / Tool | External? | Why Used In Task 5 | Install / Use |
|---|---:|---|---|
| Go standard library (`encoding/json`, `context`, `errors`, `log/slog`) | No | Envelope decode, cancellation, categorized errors aur safe structured logs | Go ke saath built-in; additional install nahi |
| RabbitMQ | Yes | `order.events`, `payment.events`, `user.events` ko asynchronously deliver karne ke liye local default broker | Platform local Docker Compose stack me RabbitMQ service run karo; management UI se exchanges/queues inspect karo |
| `github.com/rabbitmq/amqp091-go` | Yes, when source is implemented | Go service ko RabbitMQ AMQP connection, queue binding aur ack/nack API dene ke liye | `cd backend/services/notification-service` then `go get github.com/rabbitmq/amqp091-go` |
| MongoDB | Yes, Task 2 dependency | Delivery records aur event idempotency key persist karne ke liye | Local stack ka MongoDB use karo; database `notification_db` |
| MongoDB Go Driver | Yes, already relevant to Task 2 storage | `idempotency_key` unique insert/find aur delivery update ke liye | Service implementation me `go.mongodb.org/mongo-driver/v2` use hoga |
| Kafka | Optional alternate, not implemented | High-throughput/replay future architecture option | Is Task 5 reference code me install/use nahi; broker migration par separate adapter banao |
| Mermaid | Documentation only | Architecture aur sequence flow visually explain karne ke liye | Markdown viewer/GitHub-style renderer support; runtime dependency nahi |

### RabbitMQ Go Client Basic Use

```bash
cd backend/services/notification-service
go get github.com/rabbitmq/amqp091-go
```

```go
connection, err := amqp.Dial(config.RabbitMQURL)
if err != nil {
	return err
}
channel, err := connection.Channel()
if err != nil {
	return err
}
if err := DeclareNotificationEventQueues(channel); err != nil {
	return err
}
```

> 🟢 RabbitMQ client sirf actual backend implementation phase me add kiya jayega. Current requested artifact documentation file hai, isliye dependency install command run nahi ki gayi.

### Tools Select Karne Ka Reason

| Choice | Reasoning |
|---|---|
| RabbitMQ adapter | Existing local platform decision match hota hai aur routing/inspection simple hai |
| Interface-backed handler | Unit tests broker ke bina fast run ho sakte hain; Kafka alternate future me business logic reuse kar sakta hai |
| Mongo unique idempotency key | At-least-once duplicate redelivery ke against durable guard milta hai |
| Existing renderer/provider pipeline | Event consumer me content/vendor logic duplicate nahi hota |
| Mermaid diagrams | Beginner ko async boundaries visually samajh aati hain |

---

## 🛡️ Security And Reliability Checklist

| Check | Task 5 Implementation Rule |
|---|---|
| Cross-service DB access | Consumer Order/Payment/User DB read nahi kare; event snapshot use kare |
| Payload minimization | Recipient aur required display variables ke alawa sensitive information publish nahi |
| Secret logging | Raw payload, email/phone, rendered body, payment token aur keys logs me nahi |
| Event validation | Unknown version, missing metadata aur invalid recipient provider tak nahi pahunchte |
| Idempotency | `event_id:template_key:channel` unique key before send claim karo |
| Traceability | `event_id`, `trace_id`, source type aur delivery id safely correlate karo |
| Ack discipline | Processing successfully complete hone se pehle `Ack` nahi |
| Bad message | Invalid event ko endlessly requeue nahi; metric/alert emit karo |
| Transient dependency issue | Recoverable failure surface karo; retry/DLQ policy Task 6 me finalize karo |
| Provider acceptance | `accepted` ko delivered/opened label nahi karo |
| Template control | Consumer allowlisted template key hi invoke kare |
| Marketing/privacy | Preference behaviour Task 7 ke bina promotional sends start nahi karo |

---

## 🔍 Beginner-Friendly Example Walkthrough

### Example: Order Paid

1. Order Service apna payment-confirmed business state store karta hai aur `OrderPaid` event `order.events` par publish karta hai.
2. RabbitMQ event ko `notification.order.events.v1` queue me route karta hai.
3. Notification consumer envelope ke `event_id`, type aur version validate karta hai.
4. Router `OrderPaid` ko `order_status_update` email template aur `status=Paid` variable se map karta hai.
5. Handler `evt_order_paid_123:order_status_update:email` claim karta hai.
6. Existing send pipeline template render karke configured email provider ko message deta hai.
7. Delivery record source event identity ke saath store hota hai.
8. Handler successful hua to consumer message `Ack` karta hai.
9. Broker same message redeliver kare to unique claim duplicate detect karta hai aur second email send nahi hota.

```mermaid
flowchart TD
    A[OrderPaid published] --> B[RabbitMQ routes to notification queue]
    B --> C{Envelope valid?}
    C -- No --> X[Reject + metric]
    C -- Yes --> D[Map to order_status_update email]
    D --> E{Idempotency key new?}
    E -- No --> Y[Ack duplicate]
    E -- Yes --> F[Render + send + store delivery]
    F --> G{Trigger success?}
    G -- Yes --> H[Ack processed]
    G -- Temporary failure --> I[Nack/requeue baseline<br/>Task 6 later improves]
```

---

## ✅ Implementation Checklist

| Requirement / Check | Status In This Guide |
|---|---|
| Notification Service Task 5 scope identified from project docs | ✅ |
| Only required `task5.md` documentation output added | ✅ |
| Order/payment/user event consumer flow documented | ✅ |
| RabbitMQ local broker choice explained | ✅ |
| Versioned envelope and payload contracts provided | ✅ |
| Event-to-template/channel rule mapping provided | ✅ |
| Idempotency and acknowledgement handling explained | ✅ |
| Clean reference folder structure provided | ✅ |
| External libraries/tools, installation and usage documented | ✅ |
| Mermaid architecture/flow diagrams included | ✅ |
| Tests, security and observability guidance included | ✅ |
| Retry/DLQ, preferences and analytics left to later tasks | ✅ |

---

## 🏁 Final Result

Notification Service - Task 5 ke liye ek focused event-consumer blueprint ready hai:

- `order.events`, `payment.events` aur `user.events` RabbitMQ queues se consume honge.
- Common versioned envelope aur minimal payload contract event handling predictable banayenge.
- Allowlisted rules events ko template-based notification triggers me map karenge.
- Durable idempotency key at-least-once redelivery me duplicate intended messages ko rokegi.
- Existing channels, templates aur delivery storage reuse honge; provider/retry/preference logic consumer me duplicate nahi hogi.

> ✅ **Task 5 complete:** Documentation-level step-by-step implementation guide ready hai. Actual service code, broker wiring aur later Notification tasks intentionally implement nahi kiye gaye, kyunki requested output sirf required folder structure aur `task5.md` content hai.
