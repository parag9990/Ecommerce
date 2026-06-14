# 🔔 Notification Service - Task 6: Retry And DLQ

![Task](https://img.shields.io/badge/Task-Notification%20Service%20%236-blue)
![Priority](https://img.shields.io/badge/Priority-P1-orange)
![Dependency](https://img.shields.io/badge/Dependency-Queue-blueviolet)
![Broker](https://img.shields.io/badge/Local%20Default-RabbitMQ-FF6600)
![Pattern](https://img.shields.io/badge/Pattern-Retry%20%2B%20DLQ-informational)
![Reliability](https://img.shields.io/badge/Reliability-Exponential%20Backoff-success)
![Guide](https://img.shields.io/badge/Language-Hinglish-yellow)

## 📌 Task Summary

| Field | Detail |
|---|---|
| Task Name | Retry and DLQ |
| Requirement Source | `docs/01-micro-tasks.md` -> `Notification Service` -> Task 6 |
| Original Goal | Provider failure pe retry karo; final failure pe dead-letter queue me bhejo |
| Dependency | Message queue |
| Priority | `P1` |
| Local Broker Decision | RabbitMQ, kyunki existing local platform guide isko beginner-friendly default choose karti hai |
| Retry Strategy | Bounded exponential backoff: initial attempt + delayed retry tiers |
| Terminal Failure Strategy | Durable dead-letter queue (DLQ) + failed delivery state + alerting |
| Reused Building Blocks | Providers (Task 1), Mongo deliveries (Task 2), templates (Task 3), send/event trigger paths (Tasks 4-5) |
| Deliverable | Retry/DLQ implementation blueprint aur beginner-friendly step-by-step guide |
| Implementation Type | Documentation + reference code/design only |

> **Simple Hinglish goal:** Kabhi email/SMS/push provider temporarily down hota hai ya timeout deta hai. Aise transient failure par notification turant lose nahi honi chahiye. Notification Service bounded delay ke saath dobara attempt karega. Agar retries exhaust ho jayen, ya request permanently invalid ho, to message `DLQ` me safely park hoga jahan operator investigate/replay kar sakta hai.

---

## ✅ Output Created

Request ke according repository me sirf ye documentation artifact add kiya gaya:

```text
TaskImplementation/
└── Notification Service/
    ├── task1.md                         # Existing: channels/provider abstraction
    ├── task2.md                         # Existing: MongoDB templates/delivery storage
    ├── task3.md                         # Existing: template engine
    ├── task4.md                         # Existing: OTP send flow
    ├── task5.md                         # Existing: event consumers
    └── task6.md                         # Current: retry and dead-letter queue guide
```

> 🟢 **Important:** Neeche ka Go code, RabbitMQ topology, MongoDB field extensions, configuration aur test plan ek implementable reference blueprint hai. Is output me `backend/services/notification-service/`, queue declarations, dependencies, migrations, providers ya running integration physically create/modify nahi kiye gaye.

---

## 📚 Project Requirements Se Kya Samjha

| Source | Task 6 ke liye decision |
|---|---|
| `docs/01-micro-tasks.md` | Exact Task 6 scope: provider failure par retry, final failure par dead-letter queue; dependency `Queue`, priority `P1` |
| `docs/04-microservice-design.md` | Notification Service delivery retry aur DLQ manage karega; retry exponential backoff ke saath hona chahiye |
| `docs/11-devops-external-services.md` | Consumers ke liye idempotency, DLQ aur retry with backoff required operational rules hain |
| `docs/11-devops-external-services.md` | Message broker Kafka ya RabbitMQ ho sakta hai; simple routing/command queues ke liye RabbitMQ acceptable hai |
| `docs/03-folder-structure.md` | Notification source home `backend/services/notification-service/internal/` hai; existing event/send use cases ke saath retry layer fit hogi |
| `database/mongodb-schema-design.md` | Existing `notification_deliveries` collection me `status`, `attempts`, provider fields aur timestamps already fit hain |
| `TaskImplementation/Platform Foundation/task4.md` | Local environment ke liye RabbitMQ default aur Management UI documented hai |
| `TaskImplementation/Notification Service/task1.md` | Provider abstraction channel-specific send aur error surface ka base hai |
| `TaskImplementation/Notification Service/task2.md` | MongoDB delivery record attempts/status store karne ka service-owned home hai |
| `TaskImplementation/Notification Service/task3.md` | Retry ko same template rendering contract reuse karna hai, new arbitrary content nahi banana |
| `TaskImplementation/Notification Service/task4.md` | OTP plaintext persist/log nahi ho sakta; durable OTP retry payload forbidden rahega |
| `TaskImplementation/Notification Service/task5.md` | Event consumer transient provider error ko Task 6 retry/DLQ pipeline tak hand off karega |

---

## 🚧 Scope Boundary

Task 6 ka kaam **already intended notification delivery ko provider failure ke baad reliably retry ya dead-letter karna** hai. Ye naya campaign engine, preference feature ya analytics product nahi hai.

### Included In Task 6

| Included | Kyon |
|---|---|
| Retryable aur terminal provider failures ki classification | Har error ko blind retry karna wasteful aur unsafe ho sakta hai |
| Bounded exponential backoff policy | Temporary provider incident me recovery chance milta hai, provider flood nahi hota |
| RabbitMQ delayed retry queues with dead-letter routing | Default RabbitMQ features se plugin-free scheduled redelivery milti hai |
| Final-failure DLQ | Exhausted/invalid delivery investigation ke liye preserve hoti hai |
| Mongo delivery retry metadata/status update design | Attempts aur final reason traceable rahte hain |
| Idempotent retry job contract | Broker redelivery se duplicate sends reduce hote hain |
| Task 5 event-send handoff rules | Async notifications ka provider failure correctly retry layer me enter hota hai |
| OTP safety exception | Security boundary accidentally break nahi hoti |
| Metrics, logging, tests aur operational replay guidance | Reliability feature observable aur supportable banta hai |

### Explicitly Not Included

| Deferred / Owned Elsewhere | Reason |
|---|---|
| Actual email/SMS/push provider SDK integration | Task 1/provider adapters ka concern hai |
| OTP generation, verification, resend policy ya rate limits | Auth Service aur Task 4 boundary hai |
| Event consumer business mapping | Notification Service - Task 5 |
| User opt-in/opt-out and preferred channels | Notification Service - Task 7 |
| Delivered/opened/campaign analytics pipeline | Notification Service - Task 8 |
| Admin DLQ web UI or public replay API | Is task me queue reliability design chahiye, management product nahi |
| Kafka retry-topic implementation | Local baseline RabbitMQ chosen hai; Kafka future adapter ho sakta hai |
| Backend files, migrations, Docker compose ya installed modules | Requested output sirf `task6.md` documentation artifact hai |

> 🔴 **Boundary rule:** Retry job me rendered body, provider secret, password, token ya OTP plaintext store/publish nahi karna. OTP send fail ho to durable retry queue me OTP payload bhejne ke bajay Auth flow generic failure return karega aur user-controlled fresh OTP resend policy follow hogi.

---

## 🔗 Existing Tasks Ke Saath Integration

| Previous Task | Task 6 Me Reuse |
|---|---|
| Task 1: Channels/providers | Provider error ko transient ya terminal category me map karna; retry par same configured channel provider invoke karna |
| Task 2: MongoDB deliveries | `attempts`, status, provider aur timestamps ke saath retry audit metadata maintain karna |
| Task 3: Template engine | Retryable non-OTP delivery me approved template operation reuse karna; arbitrary content inject nahi karna |
| Task 4: Send OTP | No durable plaintext OTP retries; security rule preserve karna |
| Task 5: Event consumers | Order/payment/user triggered notification failure ko retry command me hand off karna |

### Task 6 Kya Replace Nahi Karta

```text
Task 5 event validation / mapping       # Retry business event ko reinterpret nahi karta
Task 3 template authorization           # Retry unapproved template create nahi karta
Task 1 provider implementation          # Retry vendor SDK ka alternate nahi hai
Task 4 OTP secrecy                      # Retry ke naam par OTP persist nahi hota
Task 7 preferences                      # Retry channel preference engine nahi hai
```

---

## 🧠 Reliability Concepts: Beginner Quick Guide

| Term | Simple Hinglish Meaning |
|---|---|
| Attempt | Provider ko actual send call karne ka ek try |
| Transient failure | Temporary problem, jaise timeout, provider unavailable ya HTTP `5xx`; baad me successful ho sakta hai |
| Terminal failure | Same request dobara bhejne se fix nahi hogi, jaise invalid recipient/template ya provider rejection |
| Backoff | Har retry se pehle increasing wait, taaki failing provider par pressure kam rahe |
| Retry queue | Message ko configured delay tak hold karke phir work queue me return karne wali queue |
| DLQ | Dead-letter queue; jahan automated processing se resolve na hone wali job investigation ke liye rakhi jati hai |
| Idempotency | Same job redeliver ho tab duplicate customer message avoid karne ka guard |
| Publisher confirm | RabbitMQ acknowledgement ki retry/DLQ message broker ne accept kiya, uske baad current job `Ack` hoti hai |

---

## 🏗️ Architecture

### Retry And DLQ Components

```mermaid
flowchart LR
    Trigger[Task 5 Event Trigger<br/>or Send Pipeline] --> AttemptQ[notification.delivery.attempt.v1]
    AttemptQ --> Worker[Delivery Attempt Worker<br/>Task 6]
    Worker --> Provider[Task 1 Provider]
    Worker --> Mongo[(notification_deliveries)]

    Provider -->|Accepted| Done[Mark accepted + Ack]
    Provider -->|Transient failure| Policy{Retries remain?}
    Provider -->|Terminal failure| DeadPub[Publish final failure]

    Policy -->|Yes| RetryX[notification.delivery.retry exchange]
    RetryX --> R1[retry.after.30s]
    RetryX --> R2[retry.after.2m]
    RetryX --> R3[retry.after.8m]
    R1 -->|TTL expires| AttemptQ
    R2 -->|TTL expires| AttemptQ
    R3 -->|TTL expires| AttemptQ
    Policy -->|No| DeadPub

    DeadPub --> DLX[notification.delivery.dlx]
    DLX --> DLQ[notification.delivery.dlq.v1]
    DLQ --> Ops[Operator investigation<br/>controlled replay]

    classDef task fill:#e8f5e9,stroke:#2e7d32,color:#1b5e20;
    classDef retry fill:#fff8e1,stroke:#f9a825,color:#6d4c41;
    classDef failed fill:#ffebee,stroke:#c62828,color:#b71c1c;
    classDef dep fill:#e3f2fd,stroke:#1565c0,color:#0d47a1;
    class Worker,Policy task;
    class RetryX,R1,R2,R3 retry;
    class DeadPub,DLX,DLQ failed;
    class Trigger,AttemptQ,Provider,Mongo,Done,Ops dep;
```

### End-To-End Attempt Flow

```mermaid
sequenceDiagram
    autonumber
    participant E as Task 5 / Send Pipeline
    participant WQ as Attempt Queue
    participant W as Retry Worker
    participant DB as notification_deliveries
    participant P as Provider
    participant RQ as Delayed Retry Queue
    participant DQ as DLQ

    E->>WQ: Retry-safe job (delivery_id, attempt=1)
    WQ->>W: Consume attempt
    W->>DB: Claim attempt idempotently
    W->>P: Send via configured provider
    alt Accepted
        P-->>W: provider_message_id
        W->>DB: status=accepted, attempts=attempt
        W->>WQ: Ack
    else Temporary provider failure and retries remain
        P-->>W: unavailable / timeout
        W->>RQ: Publish attempt+1 with delay (confirmed)
        W->>DB: status=retry_scheduled, next_retry_at
        W->>WQ: Ack current attempt
        RQ-->>WQ: TTL expires, route job back
    else Permanent or retries exhausted
        P-->>W: rejected / final unavailable
        W->>DQ: Publish sanitized final job (confirmed)
        W->>DB: status=dead_lettered, failure_code
        W->>WQ: Ack current attempt
    end
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
    ├── task5.md
    └── task6.md
```

### Reference Source Layout For Implementing Only Task 6

Neeche ka backend layout Task 6 ko documented Go notification service me fit karta hai. Ye files is documentation-only output me physically create nahi ki ja rahi hain.

```text
backend/
└── services/
    └── notification-service/
        ├── .env.example                               # Task 6: retry/DLQ env values
        ├── cmd/
        │   └── server/
        │       └── main.go                            # Future wiring: attempt worker startup
        └── internal/
            ├── domain/
            │   ├── delivery.go                        # Existing: delivery + attempts
            │   └── retry_job.go                       # Task 6: metadata-only retry contract
            ├── provider/
            │   └── provider.go                        # Existing errors classified by retry policy
            ├── repository/
            │   ├── notification_repository.go         # Task 6: retry transition contract
            │   └── mongo_notification_repository.go   # Task 6: atomic state transitions
            ├── retry/
            │   ├── policy.go                          # Task 6: max attempts/backoff/classifier
            │   ├── rabbitmq_topology.go               # Task 6: work/retry/DLQ declarations
            │   ├── publisher.go                       # Task 6: confirmed retry/DLQ publish
            │   ├── worker.go                          # Task 6: process one attempt
            │   └── worker_test.go                     # Task 6: reliability behaviour tests
            └── events/
                └── handler.go                         # Task 5 integration: enqueue attempt
```

### Intentionally Absent From Task 6

```text
internal/preference/                                  # Task 7
internal/analytics/                                   # Task 8
transport/http/admin_dlq_replay.go                    # Separate admin/product decision
provider/vendor_specific_new_integration.go           # Task 1/provider implementation concern
```

---

## 🪜 Step-by-Step Implementation

## Step 1: Failure Types Decide Karo

Retry ka first rule hai: **sirf woh failure retry karo jo time ke saath theek ho sakta hai**. Invalid data ko baar-baar provider ko bhejna user experience aur queue health dono kharab karta hai.

### Failure Decision Table

| Failure Example | Category | Retry? | Final Behaviour |
|---|---|---:|---|
| Provider timeout / connection reset | `transient` | Yes | Attempts finish hone par DLQ |
| Provider service temporarily unavailable / HTTP `5xx` | `transient` | Yes | Attempts finish hone par DLQ |
| Broker/Mongo short outage before send | `transient` | Yes | Message unacked/retry-safe rahe |
| Provider rejects invalid email/phone | `terminal` | No | Direct DLQ + failed status |
| Disabled channel / provider not configured | `terminal_config` | No | Direct DLQ + configuration alert |
| Missing/inactive template or invalid variables | `terminal_content` | No | Direct DLQ + engineering alert |
| Unsupported/corrupt retry job | `terminal_message` | No | DLQ, provider invoke nahi |
| OTP durable retry request containing code | `terminal_security` | No | Reject safely, security alert; OTP store/send nahi |

### Reference Code: `internal/retry/policy.go`

```go
package retry

import (
	"context"
	"errors"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/provider"
)

type FailureKind string

const (
	FailureNone      FailureKind = ""
	FailureTransient FailureKind = "transient"
	FailureTerminal  FailureKind = "terminal"
)

type Policy struct {
	MaxAttempts int
	Delays      []time.Duration
}

func DefaultPolicy() Policy {
	return Policy{
		MaxAttempts: 4, // attempt 1 immediately, then three retries
		Delays: []time.Duration{
			30 * time.Second,
			2 * time.Minute,
			8 * time.Minute,
		},
	}
}

// NextDelay receives the provider attempt which just failed.
func (p Policy) NextDelay(failedAttempt int) (time.Duration, bool) {
	if failedAttempt < 1 || failedAttempt >= p.MaxAttempts {
		return 0, false
	}
	index := failedAttempt - 1
	if index >= len(p.Delays) {
		return 0, false
	}
	return p.Delays[index], true
}

func ClassifyProviderError(err error) FailureKind {
	switch {
	case err == nil:
		return FailureNone
	case errors.Is(err, provider.ErrProviderUnavailable),
		errors.Is(err, context.DeadlineExceeded):
		return FailureTransient
	case errors.Is(err, provider.ErrProviderRejected),
		errors.Is(err, provider.ErrChannelDisabled),
		errors.Is(err, provider.ErrProviderNotConfigured):
		return FailureTerminal
	default:
		// Unknown provider failures are sanitized by provider adapters and can
		// be retried within the bounded maximum rather than silently dropped.
		return FailureTransient
	}
}
```

### Build Explanation

| Code Part | Hinglish Explanation |
|---|---|
| `MaxAttempts: 4` | Infinite retry loop nahi; total provider calls bounded rahengi |
| `30s -> 2m -> 8m` | Increasing/exponential-style delay provider outage me pressure reduce karta hai |
| `ClassifyProviderError` | Existing Task 1 provider error contract ko reliability decision me convert karta hai |
| Unknown error bounded retry | Accidental loss se better limited recovery attempt; exhausted hone par visibility DLQ me milti hai |

> 🟡 Production me provider ke documented HTTP/error codes ko adapter ke andar sanitized `ErrProviderUnavailable` ya `ErrProviderRejected` me consistently map karna zaruri hai.

---

## Step 2: Metadata-Only Retry Job Contract Banao

RabbitMQ retry message me entire rendered email/SMS body daalne ki zarurat nahi. Minimal job metadata use karne se PII aur secret leakage ka blast radius reduce hota hai.

### Reference Code: `internal/domain/retry_job.go`

```go
package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInvalidRetryJob = errors.New("invalid notification retry job")

type RetryJob struct {
	DeliveryID      string    `json:"delivery_id"`
	IdempotencyKey  string    `json:"idempotency_key"`
	Attempt         int       `json:"attempt"`
	MaxAttempts     int       `json:"max_attempts"`
	LastFailureCode string    `json:"last_failure_code,omitempty"`
	TraceID         string    `json:"trace_id,omitempty"`
	QueuedAt        time.Time `json:"queued_at"`
}

func (j RetryJob) Validate() error {
	if strings.TrimSpace(j.DeliveryID) == "" || strings.TrimSpace(j.IdempotencyKey) == "" {
		return fmt.Errorf("%w: identity is required", ErrInvalidRetryJob)
	}
	if j.Attempt < 1 || j.MaxAttempts < 1 || j.Attempt > j.MaxAttempts {
		return fmt.Errorf("%w: invalid attempt bounds", ErrInvalidRetryJob)
	}
	if j.QueuedAt.IsZero() {
		return fmt.Errorf("%w: queued_at is required", ErrInvalidRetryJob)
	}
	return nil
}
```

### Retry Job Me Kya Hoga Aur Kya Nahi

| Included | Reason |
|---|---|
| `delivery_id` | Mongo delivery operation ko safely reload/transition karne ke liye |
| `idempotency_key` | Redelivery par same intended send ko identify karne ke liye |
| `attempt`, `max_attempts` | Policy enforce aur observability ke liye |
| Sanitized `last_failure_code` | Debugging ke liye, raw vendor response ke bina |
| `trace_id`, `queued_at` | Trace/queue latency measure karne ke liye |

| Not Included | Reason |
|---|---|
| Rendered email/SMS body | User content queue copies me expose nahi karna |
| Recipient OTP/password/token | Sensitive payload retry job me forbidden hai |
| Provider API key/authorization header | Secret leakage prevent karna |
| Raw provider error payload | PII/secrets ho sakte hain; sanitized category enough hai |

### Send Data Kahan Se Aayega?

Non-OTP event notification ka initial send operation `notification_deliveries` me required retry-safe reference/snapshot store karega, jaise template key aur allowlisted non-sensitive variables. Recipient agar retry ke liye persist karna required ho to encrypted/minimized form me service-owned storage me rakho, queue payload me nahi. Retry worker `delivery_id` se wahi approved instruction reload karke same send use case invoke karega.

> 🔐 **OTP exception:** OTP plaintext ko reload karne layak persist karna allowed nahi. Isliye OTP delivery ko metadata-only durable retry worker ke through automatically replay nahi karna; fresh OTP lifecycle Auth Service own karega.

---

## Step 3: Delivery Record Me Retry Lifecycle Track Karo

Task 2 se existing `notification_deliveries` collection reuse hogi. Task 6 me new collection mandatory nahi; document me retry state fields extend karke operational truth ek jagah rakhi ja sakti hai.

### Recommended Delivery Status Lifecycle

```mermaid
stateDiagram-v2
    [*] --> pending
    pending --> processing: worker claims attempt
    processing --> accepted: provider accepts
    processing --> retry_scheduled: transient failure + attempt remains
    retry_scheduled --> processing: retry TTL expires
    processing --> dead_lettered: terminal failure
    processing --> dead_lettered: retries exhausted
    dead_lettered --> pending: controlled manual replay
    accepted --> [*]
```

### Example MongoDB Document After A Scheduled Retry

```json
{
  "_id": "delivery_evt_order_paid_123_email",
  "idempotency_key": "evt_order_paid_123:order_status_update:email",
  "source_event_id": "evt_order_paid_123",
  "user_id": "user_123",
  "channel": "email",
  "template_key": "order_status_update",
  "status": "retry_scheduled",
  "provider": "mailpit",
  "attempts": 1,
  "max_attempts": 4,
  "last_failure_code": "provider_unavailable",
  "next_retry_at": "2026-05-27T14:00:30Z",
  "dead_lettered_at": null,
  "created_at": "2026-05-27T14:00:00Z",
  "updated_at": "2026-05-27T14:00:00Z",
  "payload": {
    "order_id": "order_123"
  }
}
```

### Suggested Task 6 Indexes

```javascript
db.notification_deliveries.createIndex(
  { idempotency_key: 1 },
  { unique: true, name: "uniq_notification_idempotency_key" }
)

db.notification_deliveries.createIndex(
  { status: 1, next_retry_at: 1 },
  { name: "delivery_retry_schedule_lookup" }
)

db.notification_deliveries.createIndex(
  { status: 1, updated_at: -1 },
  { name: "delivery_failed_operations_view" }
)
```

### Atomic Transition Contract

Retry worker ko arbitrary overwrite nahi karna chahiye. Repository conditional state transitions de:

```go
type AttemptClaim string

const (
	AttemptAcquired  AttemptClaim = "acquired"
	AttemptCompleted AttemptClaim = "completed"
	AttemptBusy      AttemptClaim = "busy"
)

type RetryDeliveryRepository interface {
	BeginAttempt(ctx context.Context, deliveryID string, attempt int) (AttemptClaim, error)
	MarkAccepted(ctx context.Context, deliveryID string, attempt int, providerMessageID string) error
	MarkRetryScheduled(ctx context.Context, deliveryID string, attempt int, nextRetryAt time.Time, code string) error
	MarkDeadLettered(ctx context.Context, deliveryID string, attempt int, code string) error
}
```

| Method | Purpose |
|---|---|
| `BeginAttempt` | Same `delivery_id + attempt` redelivery ko second provider call banne se rokta hai; active processing lease ko `busy`, finished attempt ko `completed` batata hai |
| `MarkAccepted` | Provider accepted message id aur terminal success persist karta hai |
| `MarkRetryScheduled` | Current failed attempt, safe failure category aur due time record karta hai |
| `MarkDeadLettered` | Automatic processing stop aur investigation state mark karta hai |

`processing` claim me bounded lease/expiry bhi rakho. Consumer crash ke baad expired busy attempt recover ho sake; live worker ke active claim ko duplicate message mistakenly complete samajh kar `Ack` na kiya jaye.

> 🟠 Provider ne accept kar liya aur process DB update se pehle crash ho gaya to pure exactly-once delivery guarantee automatically nahi milti. Provider-side idempotency key support ho to `idempotency_key` forward karo; warna duplicate-risk incident ko metric/alert se visible rakho.

---

## Step 4: RabbitMQ Retry Topology Declare Karo

RabbitMQ default me arbitrary scheduled messages ka built-in calendar nahi deta, lekin queue TTL aur dead-letter exchange milkar simple, reliable delay tiers bana sakte hain. External delayed-message plugin install karna is baseline me required nahi hai.

### Queue And Exchange Topology

| Name | Type | Purpose |
|---|---|---|
| `notification.delivery.attempt` | Direct exchange | Fresh aur delayed-complete jobs ko active worker queue tak route kare |
| `notification.delivery.attempt.v1` | Durable work queue | Worker yahi queue consume karke provider call kare |
| `notification.delivery.retry` | Direct exchange | Retry delay tier select kare |
| `notification.delivery.retry.30s.v1` | Durable TTL queue | First transient failure ke baad 30-second wait |
| `notification.delivery.retry.2m.v1` | Durable TTL queue | Second transient failure ke baad 2-minute wait |
| `notification.delivery.retry.8m.v1` | Durable TTL queue | Third transient failure ke baad 8-minute wait |
| `notification.delivery.dlx` | Direct dead-letter exchange | Final failures ka routing point |
| `notification.delivery.dlq.v1` | Durable queue | Investigation aur controlled replay ke liye final jobs |

### Routing Flow

```mermaid
flowchart TD
    A[Attempt exchange: routing key send] --> B[Attempt queue]
    B --> C{Worker result}
    C -->|Accepted| D[Ack and finish]
    C -->|Attempt 1 transient fail| E[Retry exchange: after.30s]
    C -->|Attempt 2 transient fail| F[Retry exchange: after.2m]
    C -->|Attempt 3 transient fail| G[Retry exchange: after.8m]
    E -->|TTL + DLX routing send| A
    F -->|TTL + DLX routing send| A
    G -->|TTL + DLX routing send| A
    C -->|Terminal or attempt 4 fails| H[DLX: failed]
    H --> I[DLQ]

    classDef live fill:#e8f5e9,stroke:#2e7d32,color:#1b5e20;
    classDef waiting fill:#fff8e1,stroke:#f9a825,color:#6d4c41;
    classDef dead fill:#ffebee,stroke:#c62828,color:#b71c1c;
    class A,B,C,D live;
    class E,F,G waiting;
    class H,I dead;
```

### Reference Code: `internal/retry/rabbitmq_topology.go`

```go
package retry

import (
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	AttemptExchange = "notification.delivery.attempt"
	AttemptQueue    = "notification.delivery.attempt.v1"
	RetryExchange   = "notification.delivery.retry"
	DeadExchange    = "notification.delivery.dlx"
	DeadQueue       = "notification.delivery.dlq.v1"
)

type retryTier struct {
	Route string
	Queue string
	Delay time.Duration
}

var tiers = []retryTier{
	{Route: "after.30s", Queue: "notification.delivery.retry.30s.v1", Delay: 30 * time.Second},
	{Route: "after.2m", Queue: "notification.delivery.retry.2m.v1", Delay: 2 * time.Minute},
	{Route: "after.8m", Queue: "notification.delivery.retry.8m.v1", Delay: 8 * time.Minute},
}

func DeclareTopology(ch *amqp.Channel) error {
	for _, exchange := range []string{AttemptExchange, RetryExchange, DeadExchange} {
		if err := ch.ExchangeDeclare(exchange, "direct", true, false, false, false, nil); err != nil {
			return err
		}
	}

	attempt, err := ch.QueueDeclare(
		AttemptQueue, true, false, false, false,
		amqp.Table{
			"x-dead-letter-exchange":    DeadExchange,
			"x-dead-letter-routing-key": "rejected",
		},
	)
	if err != nil {
		return err
	}
	if err := ch.QueueBind(attempt.Name, "send", AttemptExchange, false, nil); err != nil {
		return err
	}

	for _, tier := range tiers {
		queue, err := ch.QueueDeclare(
			tier.Queue, true, false, false, false,
			amqp.Table{
				"x-message-ttl":             tier.Delay.Milliseconds(),
				"x-dead-letter-exchange":    AttemptExchange,
				"x-dead-letter-routing-key": "send",
			},
		)
		if err != nil {
			return err
		}
		if err := ch.QueueBind(queue.Name, tier.Route, RetryExchange, false, nil); err != nil {
			return err
		}
	}

	dead, err := ch.QueueDeclare(DeadQueue, true, false, false, false, nil)
	if err != nil {
		return err
	}
	for _, route := range []string{"failed", "invalid", "rejected"} {
		if err := ch.QueueBind(dead.Name, route, DeadExchange, false, nil); err != nil {
			return err
		}
	}
	return nil
}
```

### How This Was Built

| Part | Hinglish Explanation |
|---|---|
| Durable exchanges/queues | Broker restart ke baad topology aur persistent messages survive kar saken |
| Attempt queue | Ek hi worker path har actual provider attempt execute karta hai |
| Retry tier queue | Message consume nahi hoti; TTL tak park rehti hai |
| Retry queue DLX -> attempt exchange | Delay end hone par job automatically active processing me laut ti hai |
| DLQ binding | Terminal/exhausted/rejected failures inspectable final backlog bante hain |

---

## Step 5: Retry Policy Ko Queue Tier Se Map Karo

Application failed attempt ke baad next delay calculate karegi aur corresponding routing key choose karegi.

### Policy Table

| Provider Call | Agar Transient Failure Ho | Next Routing Key | Next Attempt |
|---:|---|---|---:|
| Attempt 1 | 30 seconds wait | `after.30s` | 2 |
| Attempt 2 | 2 minutes wait | `after.2m` | 3 |
| Attempt 3 | 8 minutes wait | `after.8m` | 4 |
| Attempt 4 | No retry remains | `failed` to DLQ | Stop |

### Reference Mapping Code

```go
func routeForDelay(delay time.Duration) (string, bool) {
	switch delay {
	case 30 * time.Second:
		return "after.30s", true
	case 2 * time.Minute:
		return "after.2m", true
	case 8 * time.Minute:
		return "after.8m", true
	default:
		return "", false
	}
}
```

### Environment Configuration

```dotenv
# Existing broker dependency selected by Task 5 / local foundation
NOTIFICATION_RABBITMQ_URL=amqp://ecommerce:ecommerce_password@rabbitmq:5672/ecommerce

# Task 6 retry behaviour
NOTIFICATION_RETRY_MAX_ATTEMPTS=4
NOTIFICATION_RETRY_DELAY_1=30s
NOTIFICATION_RETRY_DELAY_2=2m
NOTIFICATION_RETRY_DELAY_3=8m
NOTIFICATION_RETRY_PREFETCH=10
NOTIFICATION_RETRY_PUBLISH_CONFIRM_TIMEOUT=5s

# Existing delivery persistence
NOTIFICATION_MONGO_URI=mongodb://mongo:27017/notification_db
NOTIFICATION_MONGO_DATABASE=notification_db
NOTIFICATION_MONGO_DELIVERIES_COLLECTION=notification_deliveries
```

> 🟡 RabbitMQ queue TTL declaration immutable arguments ka part hoti hai. Environment delay value change karne par existing queue same naam se incompatible ho sakti hai. Production me versioned queue names ya broker policy migration plan use karo.

---

## Step 6: Confirmed Retry Aur DLQ Publishing Add Karo

Current work message ko tabhi `Ack` karo jab next retry message ya final DLQ message RabbitMQ ne accept kar liya ho. Warna outage ke beech failed send silently lose ho sakta hai.

### Reference Publisher Contract

```go
type Publisher interface {
	PublishAttempt(ctx context.Context, job domain.RetryJob) error
	PublishRetry(ctx context.Context, route string, job domain.RetryJob) error
	PublishDeadLetter(ctx context.Context, route string, job domain.RetryJob) error
}
```

### Publish Message Shape

```json
{
  "delivery_id": "delivery_evt_order_paid_123_email",
  "idempotency_key": "evt_order_paid_123:order_status_update:email",
  "attempt": 2,
  "max_attempts": 4,
  "last_failure_code": "provider_unavailable",
  "trace_id": "trace_checkout_123",
  "queued_at": "2026-05-27T14:00:00Z"
}
```

### RabbitMQ Publish Guidance

```go
func publishPersistent(
	ctx context.Context,
	ch *amqp.Channel,
	exchange string,
	route string,
	body []byte,
) error {
	return ch.PublishWithContext(
		ctx,
		exchange,
		route,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now().UTC(),
			Body:         body,
		},
	)
}
```

Actual implementation me channel ko confirm mode me enable karke broker confirmation wait karo:

```go
if err := ch.Confirm(false); err != nil {
	return err
}
confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))
// PublishPersistent(...) ke baad positive confirmation receive hone par hi
// currently consumed RabbitMQ delivery ko Ack karo.
```

### Why Confirmation Matters

| Without Confirm | With Confirm |
|---|---|
| Worker current message `Ack` kar sakta hai before retry broker me durable aaye | Worker ko pata chalta hai next-hop message broker ne receive kiya |
| Retry handoff gap me notification disappear ho sakti hai | Failure par current job unacked/re-deliverable rakhi ja sakti hai |
| Incident silently customer-facing loss ban sakta hai | Incident retry/redelivery aur monitoring me visible rahta hai |

---

## Step 7: Attempt Worker Implement Karo

Worker ek job consume karta hai, job validate karta hai, current attempt claim karta hai, existing send operation invoke karta hai, aur result ke basis par success/retry/DLQ transition choose karta hai.

### Reference Orchestration: `internal/retry/worker.go`

```go
package retry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
)

type AttemptSender interface {
	SendDeliveryAttempt(ctx context.Context, deliveryID string) (providerMessageID string, err error)
}

type AttemptClaim string

const (
	AttemptAcquired  AttemptClaim = "acquired"
	AttemptCompleted AttemptClaim = "completed"
	AttemptBusy      AttemptClaim = "busy"
)

var ErrAttemptBusy = errors.New("notification delivery attempt is already processing")

type DeliveryStore interface {
	BeginAttempt(ctx context.Context, deliveryID string, attempt int) (AttemptClaim, error)
	MarkAccepted(ctx context.Context, deliveryID string, attempt int, providerMessageID string) error
	MarkRetryScheduled(ctx context.Context, deliveryID string, attempt int, nextRetryAt time.Time, code string) error
	MarkDeadLettered(ctx context.Context, deliveryID string, attempt int, code string) error
}

type JobPublisher interface {
	PublishRetry(ctx context.Context, route string, job domain.RetryJob) error
	PublishDeadLetter(ctx context.Context, route string, job domain.RetryJob) error
}

type Worker struct {
	policy    Policy
	store     DeliveryStore
	sender    AttemptSender
	publisher JobPublisher
	now       func() time.Time
}

func (w *Worker) Process(ctx context.Context, job domain.RetryJob) error {
	if err := job.Validate(); err != nil {
		return w.publisher.PublishDeadLetter(ctx, "invalid", job)
	}

	claim, err := w.store.BeginAttempt(ctx, job.DeliveryID, job.Attempt)
	if err != nil {
		return fmt.Errorf("begin delivery attempt: %w", err)
	}
	switch claim {
	case AttemptCompleted:
		return nil // attempt is already durably terminal/safely handed off
	case AttemptBusy:
		return ErrAttemptBusy // adapter must not ack unfinished live/leased work
	case AttemptAcquired:
		// This worker owns the current provider call.
	default:
		return fmt.Errorf("unknown delivery attempt claim %q", claim)
	}

	messageID, sendErr := w.sender.SendDeliveryAttempt(ctx, job.DeliveryID)
	if sendErr == nil {
		return w.store.MarkAccepted(ctx, job.DeliveryID, job.Attempt, messageID)
	}

	code := safeFailureCode(sendErr)
	if ClassifyProviderError(sendErr) == FailureTransient {
		if delay, ok := w.policy.NextDelay(job.Attempt); ok {
			nextAt := w.now().Add(delay)
			job.Attempt++
			job.LastFailureCode = code
			job.QueuedAt = w.now()
			route, _ := routeForDelay(delay)
			// Confirm the next-hop job first. If persistence fails afterward,
			// duplicate queue jobs are safer than a silently lost retry.
			if err := w.publisher.PublishRetry(ctx, route, job); err != nil {
				return err
			}
			return w.store.MarkRetryScheduled(ctx, job.DeliveryID, job.Attempt-1, nextAt, code)
		}
	}

	job.LastFailureCode = code
	if err := w.publisher.PublishDeadLetter(ctx, "failed", job); err != nil {
		return err
	}
	return w.store.MarkDeadLettered(ctx, job.DeliveryID, job.Attempt, code)
}
```

`safeFailureCode` raw provider error ko persist nahi karta; wo allowlisted category return karta hai:

```go
func safeFailureCode(err error) string {
	switch {
	case errors.Is(err, provider.ErrProviderUnavailable),
		errors.Is(err, context.DeadlineExceeded):
		return "provider_unavailable"
	case errors.Is(err, provider.ErrProviderRejected):
		return "provider_rejected"
	case errors.Is(err, provider.ErrChannelDisabled):
		return "channel_disabled"
	case errors.Is(err, provider.ErrProviderNotConfigured):
		return "provider_not_configured"
	default:
		return "provider_unknown_failure"
	}
}
```

### Ack Rule At RabbitMQ Adapter

`Worker.Process(...)` ke return ke baad broker adapter rule:

| Worker Result | RabbitMQ Action | Reason |
|---|---|---|
| `nil` after accepted mark | `Ack` | Delivery success durable hai |
| `nil` after confirmed retry publish | `Ack` | Next attempt safely queue me hai |
| `nil` after confirmed DLQ publish | `Ack` | Final failed job safely parked hai |
| Invalid job DLQ publish succeeds | `Ack` | Poison message active queue ko block nahi karegi |
| `ErrAttemptBusy` | `Nack`/delayed requeue after lease policy | Active unfinished claim ko completed samajh kar lose nahi karna |
| Store unavailable / retry publish unconfirmed | Leave unacked or `Nack(requeue=true)` | Work lose nahi honi chahiye |

### Important Ordering Note

Mongo status update aur RabbitMQ publish ek single distributed transaction nahi hain. Is guide ka safe baseline:

1. Every attempt key ko idempotent claim karo.
2. Next retry ya DLQ publish ko publisher confirm ke saath **pehle** durable banao, phir scheduled/dead-letter state mark karo.
3. Confirmed publish ke baad state update fail ho to current delivery acknowledge mat karo; duplicate queued metadata jobs ko next-attempt idempotency suppress karegi.
4. Accepted-provider-result ke baad DB update fail hone wale case ke liye provider idempotency key ya reconciliation alert rakho.
5. Provider idempotency key supported ho to same notification idempotency key pass karo.
6. Duplicate-risk aur stuck-state metrics/alerts mandatory rakho.

Very high reliability requirement me transactional outbox/inbox pattern future hardening ho sakta hai, lekin is Task 6 guide ka required outcome bounded retry plus DLQ hai.

---

## Step 8: Task 5 Event Consumers Se Handoff Wire Karo

Task 5 order/payment/user event ko notification intent me map karta hai. Task 6 ke saath do compatible integration styles hain:

| Integration Style | Flow | Recommended Use |
|---|---|---|
| Queue-first attempts | Task 5 delivery intent create karke `attempt=1` work job publish kare; Task 6 worker first provider call bhi kare | Async event notifications ke liye recommended, because all attempts same reliability path use karte hain |
| Direct first attempt then retry | Existing send pipeline attempt 1 directly kare; transient failure par `attempt=2` retry queue me publish kare | Existing synchronous-ish call path ko minimally adapt karne ke liye |

### Recommended Queue-First Event Flow

```mermaid
sequenceDiagram
    participant C as Task 5 Consumer
    participant DB as Mongo Delivery
    participant A as Attempt Queue
    participant W as Task 6 Worker
    participant P as Provider

    C->>DB: Create pending delivery/idempotency key
    C->>A: Publish attempt=1 (confirmed)
    C->>C: Ack source event
    A->>W: Consume first attempt
    W->>P: Send
    alt Transient provider error
        W->>DB: retry_scheduled
        W->>A: Later retry through TTL tiers
    else Accepted
        W->>DB: accepted
    end
```

### Source Event Ack Safety

| Situation | Task 5 Source Event Action |
|---|---|
| Pending delivery exists and initial attempt publish confirmed | Source event `Ack` ho sakta hai |
| Mongo delivery create fails | Source event `Nack/requeue`; retry job create assume mat karo |
| Initial attempt publish fails/unconfirmed | Source event `Nack/requeue`; existing `pending` delivery ko dekh kar missing initial job re-publish karo, sirf duplicate samajh kar `Ack` nahi |
| Duplicate source event already has same pending/accepted delivery | Second user notification publish mat karo; safe acknowledgement decide by saved state |

---

## Step 9: OTP Ke Liye Security-Safe Behaviour Rakho

Task 4 ke OTP flow ka plaintext code send karne ke waqt memory me available hota hai, lekin DB/log/queue me persist nahi ho sakta. Generic Task 6 worker ko OTP body dobara reconstruct karne ke liye OTP store karna padta, jo security boundary break karega.

### OTP Decision Table

| OTP Situation | Correct Behaviour |
|---|---|
| Provider current request ke andar briefly timeout kare and in-memory bounded immediate adapter retry allowed ho | Same request deadline ke andar optional short retry; OTP persist nahi |
| Provider unavailable and request fails | Auth ko generic send failure return karo; user fresh resend flow use kare |
| Queue retry job me OTP/body/code detected | Job reject/DLQ-security route; code log/store/send nahi |
| Delivery audit needed | Only safe metadata store karo: channel, accepted/failed, provider category, timestamps; no OTP/body |

```mermaid
flowchart TD
    OTP[Auth -> Send OTP] --> Send[Transient in-memory send]
    Send -->|Accepted| Meta[Store safe metadata only]
    Send -->|Unavailable| Fail[Return generic failure to Auth]
    Fail --> Fresh[Auth-controlled fresh resend request]
    Send -. Never persist OTP .-> Block[No durable retry payload]

    classDef safe fill:#e8f5e9,stroke:#2e7d32,color:#1b5e20;
    classDef stop fill:#ffebee,stroke:#c62828,color:#b71c1c;
    class OTP,Send,Meta,Fail,Fresh safe;
    class Block stop;
```

---

## Step 10: DLQ Processing Aur Controlled Replay Define Karo

DLQ dustbin nahi hai. Ye actionable failure backlog hai. Har DLQ record sanitized reason ke saath investigate hona chahiye.

### DLQ Reason Codes

| Code | Meaning | Typical Action |
|---|---|---|
| `provider_unavailable_exhausted` | Maximum attempts tak provider unavailable raha | Provider health check; recovery ke baad selected replay |
| `provider_rejected_recipient` | Recipient invalid/rejected | User/contact correction; automatic replay nahi |
| `channel_disabled` | Config me channel disabled tha | Configuration fix; affected safe messages review/replay |
| `provider_not_configured` | Enabled send ke liye adapter config missing | Deploy/config fix; replay |
| `template_invalid` | Template/variables render nahi huye | Template fix; affected transactional deliveries review |
| `invalid_retry_job` | Queue job corrupt/unsupported tha | Engineering investigation; blind replay nahi |
| `security_payload_blocked` | Forbidden sensitive retry payload detected | Security review; replay nahi |

### Replay Safety Checklist

| Before Replay | Check |
|---|---|
| Failure root cause fixed? | Same failing message endless cycle me return na ho |
| Delivery already accepted later? | Duplicate message avoid karo |
| Notification still useful/current? | Old order/payment update user ko confuse na kare |
| OTP or expired security content involved? | Never replay |
| Idempotency key handling defined? | Manual replay deliberate new attempt ho, accidental duplicate nahi |
| Operator identity/audit present? | Sensitive operation traceable ho |

> 🟡 Task 6 DLQ aur replay rules define karta hai. Dedicated admin HTTP endpoint/dashboard banana scope ke bahar hai; initial operations RabbitMQ UI ya controlled internal tooling se ho sakti hain.

---

## Step 11: Runtime Wiring Aur Configuration Validate Karo

### Startup Order

| Order | Startup Work | Failure Behaviour |
|---:|---|---|
| 1 | Env/config parse aur retry values validate karo | Invalid max attempts/delays par service fail-fast |
| 2 | Mongo delivery repository connect karo | Durable state absent ho to worker start nahi |
| 3 | Provider registry/send use case ready karo | Retry worker incomplete dependency ke saath consume nahi kare |
| 4 | RabbitMQ connection/channel open karo | Readiness false rakho |
| 5 | Retry/DLQ exchanges aur queues declare karo | Topology mismatch par startup error |
| 6 | Publisher confirm channel enable karo | Reliable handoff unavailable ho to consume start nahi |
| 7 | Prefetch set karke attempt consumer start karo | Controlled concurrent provider calls |

### Config Validation Rules

| Config | Validation |
|---|---|
| `NOTIFICATION_RETRY_MAX_ATTEMPTS` | `>= 1` and bounded operational maximum, e.g. `<= 10` |
| Retry delays | Valid positive Go durations aur strictly increasing |
| Delay tiers count | `max_attempts - 1` ke equal |
| RabbitMQ URL | Required and secret-safe logging; password log nahi |
| MongoDB config | Existing Task 2 validation reuse |
| Prefetch | Positive limit; provider quota se compatible |
| Publish confirm timeout | Positive duration; overly short flaky setting avoid |

---

## Step 12: Observability Aur Alerts Add Karo

Retries visible nahi hain to service outwardly “working” dikh sakti hai jab customer notifications silently delayed ho rahi hon.

### Minimum Metrics

| Metric | Labels | Reason |
|---|---|---|
| `notification_delivery_attempts_total` | `channel`, `provider`, `result` | Provider acceptance/failure rate |
| `notification_delivery_retries_scheduled_total` | `channel`, `delay_tier`, `failure_code` | Retry pressure aur incident diagnosis |
| `notification_delivery_dead_lettered_total` | `channel`, `failure_code` | Customer-impacting final failures alert |
| `notification_delivery_attempt_latency_seconds` | `channel`, `provider` | Slow provider identify karna |
| `notification_retry_queue_depth` | `queue` | Backlog/stuck retry detection |
| `notification_dlq_depth` | `queue` | Manual action required signal |
| `notification_duplicate_attempt_skipped_total` | `channel` | Redelivery/idempotency behaviour visibility |

### Suggested Alerts

| Alert | Example Condition | Action |
|---|---|---|
| Provider failure spike | Transient failure rate > threshold for 5 min | Provider/config/network investigate |
| DLQ non-empty | `notification_dlq_depth > 0` for transactional notification | Failure code inspect karo |
| Retry backlog old | Oldest retry/attempt job expected delay se bahut old | Consumer/broker health check |
| Publish confirmation failures | Confirm timeouts > 0 sustained | RabbitMQ health/connectivity investigate |
| Security-blocked payload | Any `security_payload_blocked` | Immediate security review |

### Safe Log Example

```json
{
  "level": "WARN",
  "message": "notification retry scheduled",
  "delivery_id": "delivery_evt_order_paid_123_email",
  "trace_id": "trace_checkout_123",
  "channel": "email",
  "attempt": 1,
  "next_attempt": 2,
  "delay": "30s",
  "failure_code": "provider_unavailable"
}
```

### Logs Me Kabhi Nahi

```text
OTP code
rendered email or SMS body
full phone/email unless separately masked and necessary
provider API key/token
raw vendor payload containing recipient/content
```

---

## Step 13: Tests Se Reliability Behaviour Lock Karo

### Unit Test Matrix

| Test | Expected Result |
|---|---|
| Attempt 1 provider accepted | Delivery `accepted`, no retry/DLQ publish |
| Attempt 1 transient unavailable | Status `retry_scheduled`, `after.30s` publish with `attempt=2` |
| Attempt 2 transient unavailable | `after.2m` publish with `attempt=3` |
| Attempt 3 transient unavailable | `after.8m` publish with `attempt=4` |
| Attempt 4 transient unavailable | Status `dead_lettered`, `failed` DLQ publish |
| Provider rejected recipient | No retry, direct DLQ |
| Disabled/not configured provider | No retry, direct DLQ and config failure reason |
| Invalid retry job JSON/bounds | Provider not called, `invalid` DLQ route |
| Duplicate same delivery attempt claim | Provider not called again |
| Retry publish fails/unconfirmed | Current message not acknowledged |
| DB transition fails | Current message not silently completed |
| Job contains OTP/body/secret field | Block/DLQ-security test; sensitive value not logged |

### Reference Policy Unit Test

```go
func TestDefaultPolicyStopsAfterFourthAttempt(t *testing.T) {
	policy := DefaultPolicy()

	cases := []struct {
		attempt int
		want    time.Duration
		retry   bool
	}{
		{attempt: 1, want: 30 * time.Second, retry: true},
		{attempt: 2, want: 2 * time.Minute, retry: true},
		{attempt: 3, want: 8 * time.Minute, retry: true},
		{attempt: 4, want: 0, retry: false},
	}

	for _, tc := range cases {
		delay, ok := policy.NextDelay(tc.attempt)
		if ok != tc.retry || delay != tc.want {
			t.Fatalf("NextDelay(%d) = (%v, %v), want (%v, %v)",
				tc.attempt, delay, ok, tc.want, tc.retry)
		}
	}
}
```

### RabbitMQ Integration Test Plan

| Step | Verification |
|---:|---|
| 1 | Local RabbitMQ aur MongoDB start karo |
| 2 | Retry topology declare karke test worker run karo |
| 3 | Fake provider ko first call `ErrProviderUnavailable`, next call accepted configure karo |
| 4 | `attempt=1` job attempt exchange me publish karo |
| 5 | Mongo me `retry_scheduled`, RabbitMQ me first delay tier movement observe karo |
| 6 | Test-friendly short TTL override/versioned queue ke baad second attempt invoke hota verify karo |
| 7 | Delivery accepted aur DLQ empty verify karo |
| 8 | Provider ko always reject/unavailable configure karke final DLQ case publish karo |
| 9 | Attempts bounded aur DLQ reason sanitized verify karo |
| 10 | Duplicate job publish karke provider call count duplicate nahi hota verify karo |

### Manual Local Verification Commands

Commands actual implementation aur local compose available hone ke baad useful honge:

```bash
# RabbitMQ health
docker exec ecommerce-rabbitmq rabbitmq-diagnostics -q ping

# Notification service tests after Task 6 source is implemented
cd backend/services/notification-service
go test ./...
go test -race ./...
```

> 🟢 Current requested output documentation-only hai, isliye Task 6 backend tests ya RabbitMQ topology commands is step me execute nahi kiye gaye.

---

## 🧰 External Libraries And Tools

### Runtime And Development Dependencies

| Library / Tool | External? | Why Used In Task 6 | Install / Use |
|---|---:|---|---|
| Go standard library (`context`, `errors`, `encoding/json`, `time`) | No | Retry policy, cancellation, JSON job encoding aur delays represent karne ke liye | Go ke saath built-in; extra install nahi |
| RabbitMQ | Yes | Durable attempt queue, TTL delay tiers, dead-letter exchange aur DLQ provide karne ke liye | Existing local Docker Compose RabbitMQ service run karo; UI `http://localhost:15672` se queues inspect karo |
| `github.com/rabbitmq/amqp091-go` | Yes, when source is implemented | Go code se exchanges/queues declare, consume, ack/nack aur publish confirm use karne ke liye | `cd backend/services/notification-service` then `go get github.com/rabbitmq/amqp091-go` |
| MongoDB | Yes, existing Task 2 dependency | Delivery attempts, retry state, final reason aur idempotency persist karne ke liye | Local stack ka `notification_db.notification_deliveries` use karo |
| MongoDB Go Driver | Yes, existing service dependency | Conditional delivery transitions aur indexes ke implementation ke liye | Existing service module me Mongo driver reuse karo |
| RabbitMQ Management UI | Yes, RabbitMQ image ke saath | Beginner ko retry queues, bindings aur DLQ depth visually inspect karne ke liye | Management-enabled RabbitMQ image use karo; browser me UI login |
| Prometheus/Grafana | Optional operational tools | Retry/DLQ metrics aur alert dashboards ke liye | Platform monitoring setup ke saath integrate karo; Task 6 source scope me dashboards mandatory nahi |
| Mermaid | Documentation only | Architecture, flow aur lifecycle diagrams readable banane ke liye | Markdown renderer/GitHub support; runtime dependency nahi |

### RabbitMQ Go Client Install And Use

Actual backend implementation phase me:

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
if err := DeclareTopology(channel); err != nil {
	return err
}
```

### Tool Choice Reasoning

| Choice | Reason |
|---|---|
| RabbitMQ TTL + DLX | Default broker features se delayed retry milta hai; local setup me additional plugin nahi |
| Bounded exponential backoff | Temporary incident recover kar sakta hai without infinite or aggressive retries |
| Mongo delivery lifecycle | Existing service-owned record ke saath investigation aur idempotency align hoti hai |
| Metadata-only job | Queue payload me customer content/secrets ka exposure minimize hota hai |
| Publisher confirms | Attempt-to-retry/DLQ handoff me silent message loss reduce hota hai |
| Mermaid | Async retry concept beginner ko visually follow karna easy hota hai |

---

## 🔍 Beginner-Friendly Example Walkthrough

### Example: Order Paid Email Ke Time Provider Temporary Down Hai

1. Task 5 `OrderPaid` event ko `order_status_update` email notification intent me map karta hai.
2. Delivery record `pending` state aur idempotency key ke saath MongoDB me create hota hai.
3. `attempt=1` metadata-only job `notification.delivery.attempt.v1` me publish hota hai.
4. Task 6 worker job claim karke configured email provider ko call karta hai.
5. Provider temporarily unavailable return karta hai. Policy isko `transient` classify karti hai.
6. Confirmed publish ke through `attempt=2` job `notification.delivery.retry.30s.v1` me safely jata hai.
7. Worker Mongo record ko `retry_scheduled`, `attempts=1`, `next_retry_at=+30s` update karta hai.
8. Queue TTL expire hone par RabbitMQ job ko attempt queue me wapas route karta hai.
9. Second provider call accepted hoti hai; worker `accepted` aur provider message id store karta hai.
10. Agar provider fourth attempt tak fail hota, delivery `dead_lettered` banti aur job DLQ me investigation ke liye park hoti.

```mermaid
flowchart TD
    A[OrderPaid notification intent] --> B[Pending delivery + attempt 1]
    B --> C[Email provider call]
    C -->|Unavailable| D[Schedule attempt 2 after 30s]
    D --> E[Email provider call]
    E -->|Accepted| F[Accepted delivery: done]
    E -->|Unavailable| G[Further bounded tiers]
    G -->|Still failing after attempt 4| H[DLQ + operator alert]

    classDef start fill:#e3f2fd,stroke:#1565c0,color:#0d47a1;
    classDef waiting fill:#fff8e1,stroke:#f9a825,color:#6d4c41;
    classDef success fill:#e8f5e9,stroke:#2e7d32,color:#1b5e20;
    classDef fail fill:#ffebee,stroke:#c62828,color:#b71c1c;
    class A,B,C,E start;
    class D,G waiting;
    class F success;
    class H fail;
```

---

## 🛡️ Security And Reliability Checklist

| Check | Task 6 Rule |
|---|---|
| Infinite retry prevention | Maximum attempts configurable and bounded rakho |
| Retry only when useful | Transient vs terminal errors explicitly classify karo |
| Duplicate prevention | Delivery/attempt idempotent claim before provider send |
| Message loss prevention | Retry/DLQ handoff publish confirm hone ke baad current message ack |
| Queue durability | Durable queues/exchanges + persistent publications |
| Payload minimization | Retry job me identifiers/sanitized metadata only |
| OTP secrecy | OTP plaintext ko retry DB/queue/log me kabhi nahi |
| Provider error hygiene | Raw vendor payload log/store na karo; sanitized code store karo |
| Recipient privacy | Full recipient unnecessary logs me nahi; persistence encrypt/minimize karo |
| DLQ actionability | Failure reason, alert aur controlled replay checklist maintain karo |
| Stale replay prevention | Old transaction notification manually replay karne se pehle usefulness verify karo |
| State traceability | Attempt count, next retry, timestamps, trace id aur final status persist karo |
| Operations | DLQ depth aur retry backlog alert karo |

---

## ✅ Implementation Checklist

| Requirement / Check | Status In This Guide |
|---|---|
| Notification Service - Task 6 requirement identified | ✅ |
| Queue dependency and `P1` priority reflected | ✅ |
| RabbitMQ local broker decision aligned with existing project guide | ✅ |
| Retryable vs terminal provider errors explained | ✅ |
| Exponential backoff attempt policy provided | ✅ |
| Durable work/retry/DLQ topology designed | ✅ |
| Mongo delivery lifecycle and indexes shown | ✅ |
| Idempotency and publisher-confirm safety covered | ✅ |
| Task 5 event-consumer integration documented | ✅ |
| OTP non-persistence boundary explicitly protected | ✅ |
| External tools/libraries with install/use guidance documented | ✅ |
| Reference code examples included | ✅ |
| Mermaid architecture/flow/state diagrams included | ✅ |
| Unit/integration/operational verification plan included | ✅ |
| Task 7/8 and unrelated implementations excluded | ✅ |
| Only requested documentation artifact created | ✅ |

---

## 🏁 Final Result

Notification Service - Task 6 ke liye structured implementation guide ready hai:

- Provider failure ko `transient` ya `terminal` classify karke only useful retries schedule hongi.
- RabbitMQ TTL retry tiers `30s -> 2m -> 8m` delay ke saath maximum four provider attempts enforce karenge.
- Terminal ya exhausted jobs durable DLQ me jayengi, Mongo delivery state aur sanitized reason ke saath.
- Idempotent claims, persistent messages aur publisher confirms reliability gaps reduce karenge.
- OTP plaintext durable retry pipeline se explicitly excluded rahega, Task 4 security boundary preserve hogi.
- User preferences, delivery analytics aur unrelated service implementation is Task 6 scope me add nahi ki gayi.
