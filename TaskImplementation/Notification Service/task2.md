# 🔔 Notification Service - Task 2: Choose MongoDB

![Task](https://img.shields.io/badge/Task-Notification%20Service%20%232-blue)
![Priority](https://img.shields.io/badge/Priority-P1-orange)
![Dependency](https://img.shields.io/badge/Dependency-Channels-blueviolet)
![Database](https://img.shields.io/badge/Database-MongoDB-47A248)
![Status](https://img.shields.io/badge/Status-Documented-success)
![Guide](https://img.shields.io/badge/Language-Hinglish-yellow)
![Scope](https://img.shields.io/badge/Scope-Templates%20%26%20Delivery%20Logs-informational)

## 📌 Task Summary

| Field | Detail |
|---|---|
| Task Name | Choose MongoDB |
| Requirement Source | `docs/01-micro-tasks.md` → `Notification Service` → Task 2 |
| Original Goal | Notification templates aur delivery logs flexible hain, isliye MongoDB suitable database choose karo |
| Dependency | Notification Service - Task 1: Channels |
| Priority | `P1` |
| Database Decision | MongoDB, service-owned database: `notification_db` |
| Included Collections | `notification_templates`, `notification_deliveries` |
| Deliverable | MongoDB persistence decision aur implementation blueprint |
| Implementation Type | Documentation + reference schema/code design only |

> **Simple Hinglish goal:** Task 1 me `email`, `sms`, `push`, aur `whatsapp_like` channel abstraction define hui. Ab Task 2 me un channels ke templates aur delivery logs ke liye aisa storage choose karna hai jahan har channel ka changing document shape safely store ho sake. Project requirements ke according ye storage **MongoDB** hoga.

---

## ✅ Output Created

Request ke according repository me sirf ye documentation artifact create kiya gaya:

```text
TaskImplementation/
└── Notification Service/
    ├── task1.md
    └── task2.md
```

> 🟢 **Important:** Neeche diya gaya database structure, commands aur Go code ek implementable reference blueprint hai. Is output me `backend/services/notification-service/`, MongoDB collections, Docker configuration, dependencies, ya running database physically create nahi kiye gaye.

---

## 📚 Project Docs Se Kya Samjha

| Document | Task 2 ke liye relevant decision |
|---|---|
| `docs/01-micro-tasks.md` | Task 2 ka exact scope: templates aur delivery logs ke liye MongoDB select karna |
| `docs/02-system-architecture.md` | Notification Service apna Mongo Notification DB own karta hai; direct cross-service DB access allowed nahi |
| `docs/03-folder-structure.md` | Future repository location: `internal/repository/mongo_notification_repository.go` |
| `docs/04-microservice-design.md` | Notification Service ka DB MongoDB hai because template/provider/delivery metadata flexible hota hai |
| `docs/05-database-design.md` | Notification ke high-value query fields: `user_id`, `status`, `template_key` |
| `database/mongodb-schema-design.md` | Database `notification_db`, templates/deliveries document examples aur required indexes already defined hain |
| `docs/11-devops-external-services.md` | Local stack me MongoDB expected dependency hai |
| `docs/13-developer-guide.md` | Clean architecture rule: repository DB details handle kare, domain DB implementation se independent rahe |
| `TaskImplementation/Notification Service/task1.md` | Persisted `channel` field ko Task 1 constants: `email`, `sms`, `push`, `whatsapp_like` follow karne hain |

---

## 🚧 Scope Boundary

Task 2 ka kaam **MongoDB persistence choice aur templates/delivery log storage blueprint** banana hai. Notification bhejna ya full feature complete karna is task ka part nahi hai.

### Included In Task 2

| Included | Kyon |
|---|---|
| MongoDB selection and reason | Flexible documents ke liye correct storage boundary decide hoti hai |
| Service-owned `notification_db` | Microservice database ownership clear hoti hai |
| `notification_templates` collection design | Channel-wise template document safely persist ho sakta hai |
| `notification_deliveries` collection design | Notification attempts/status metadata log ho sakta hai |
| Document fields, index plan and validation guidance | Query aur operational behaviour predictable banta hai |
| Mongo connection/repository reference shape | Future Go implementation ko clean layer boundary milti hai |
| Security and testing checklist | Sensitive notification data ko careless store/log hone se bachaya ja sakta hai |

### Not Included In Task 2

| Deferred Item | Later Task / Reason |
|---|---|
| Template rendering and variable interpolation | Notification Service - Task 3 |
| OTP email/SMS send method | Notification Service - Task 4 |
| Event consumers and queue wiring | Notification Service - Task 5 |
| Retry scheduler and dead-letter queue | Notification Service - Task 6 |
| `notification_preferences` behaviour or schema implementation | Notification Service - Task 7 |
| Open/delivered analytics and provider event processing | Notification Service - Task 8 / delivery lifecycle work |
| Real MongoDB container, Go module, service source files | Requested output only `task2.md` documentation hai |

> 🔴 **Boundary rule:** `notification_deliveries` collection ka storage shape define hoga, lekin provider ko call karke delivery create/update karne wala workflow is guide me implement nahi hoga.

---

## 🔗 Task 1 Dependency: Channel Values

Task 2 documents me `channel` ek free-form random string nahi hoga. Wo Task 1 ke canonical values me se hoga:

| Persisted Value | Meaning |
|---|---|
| `email` | Email notification template/delivery |
| `sms` | SMS template/delivery |
| `push` | Mobile/web push template/delivery |
| `whatsapp_like` | Vendor-neutral messaging template/delivery |

Example: `order_paid` event ke liye `email` aur `sms` templates alag documents ho sakte hain because subject/body requirements different hain.

---

## 🧠 Why MongoDB?

### Data Flexible Kyon Hai?

| Data Area | Flexible Example |
|---|---|
| Template content | Email ko `subject` chahiye; SMS ko short `body`; push me future title/data fields aa sakte hain |
| Template versions | Same `template_key` aur channel ka version history preserve ho sakta hai |
| Delivery payload | Order update me `order_id`, payment update me `payment_id`, OTP flow me alag metadata ho sakta hai |
| Provider metadata | Provider message id aur future provider response fields channel/vendor ke hisab se vary kar sakte hain |

### Decision Matrix

| Need | MongoDB Fit | Task 2 Decision |
|---|---|---|
| JSON-like variable document shape | Native BSON documents me natural storage | ✅ Use MongoDB |
| New optional template/payload fields | Schema migration ke bina additive changes possible | ✅ Useful |
| Index by template/channel/version | Compound indexes supported | ✅ Define index |
| Search deliveries by user/status/provider id | Secondary indexes supported | ✅ Define indexes |
| Financial-grade relational transactions | Notification logs ka primary need nahi | MySQL choose nahi kiya |

> 🟡 **Clarification:** MongoDB flexible hai, iska matlab validation skip karna nahi hai. Required identity/status/time fields stable rahenge; sirf channel/provider-specific metadata ko controlled flexibility milegi.

---

## 🏗️ Architecture Decision

```mermaid
flowchart LR
    C[Task 1 Channel Contract] --> N[Future Notification Usecase]
    N --> TR[(notification_templates)]
    N --> DL[(notification_deliveries)]
    TR --> DB[(MongoDB: notification_db)]
    DL --> DB

    R[Template Rendering<br/>Task 3] -. reads template .-> TR
    S[Send OTP / Provider Send<br/>Task 4+] -. records delivery .-> DL
    Q[Retry / Analytics<br/>Task 6/8] -. later query .-> DL

    classDef task2 fill:#e8f5e9,stroke:#2e7d32,color:#1b5e20;
    classDef later fill:#fff8e1,stroke:#f9a825,color:#6d4c41,stroke-dasharray: 5 5;
    class C,TR,DL,DB task2;
    class N,R,S,Q later;
```

### Database Ownership Rule

```mermaid
flowchart TB
    NS[Notification Service] --> ND[(notification_db)]
    AS[Auth Service] -. request/message later .-> NS
    OS[Order Service] -. event later .-> NS
    PS[Payment Service] -. event later .-> NS

    AS -. no direct DB access .-> ND
    OS -. no direct DB access .-> ND
    PS -. no direct DB access .-> ND
```

**Hinglish explanation:** Auth, Order ya Payment Service ko `notification_db` directly query/update nahi karna chahiye. Notification Service hi apne templates aur delivery records ka owner hoga. Isse coupling aur accidental data corruption kam hota hai.

---

## 📁 Clean Implementation Folder Structure

### Actual Task Output

```text
TaskImplementation/
└── Notification Service/
    ├── task1.md                         # Existing: channel abstraction guide
    └── task2.md                         # Current: MongoDB persistence guide
```

### Future Source Structure For Task 2 Blueprint

Project ke documented Go clean-architecture convention me Task 2 persistence later is jagah fit hogi:

```text
backend/
└── services/
    └── notification-service/
        ├── cmd/
        │   └── server/
        │       └── main.go                          # Future: config/client wiring
        └── internal/
            ├── config/
            │   └── config.go                        # Task 2 blueprint: Mongo settings
            ├── domain/
            │   ├── channel.go                       # Task 1 dependency
            │   ├── template.go                      # Task 2 blueprint: persisted template model
            │   └── delivery.go                      # Task 2 blueprint: delivery log model
            └── repository/
                ├── notification_repository.go       # Task 2 blueprint: interfaces
                └── mongo_notification_repository.go # Task 2 blueprint: Mongo adapter
```

### Intentionally Absent In Task 2

```text
internal/usecase/render_template.go       # Task 3
internal/usecase/send_otp.go              # Task 4
internal/events/consumer.go               # Task 5
internal/retry/                           # Task 6
internal/domain/preference.go             # Task 7
internal/analytics/                       # Task 8
```

---

## 🪜 Step-by-Step Implementation

## Step 1: Database Boundary Freeze Karo

Sabse pehle service-owned database aur collection names stable define karo. Ye naming project ke existing MongoDB schema document se directly align karti hai.

| Item | Chosen Value | Reason |
|---|---|---|
| Database | `notification_db` | Notification Service ka isolated ownership boundary |
| Templates Collection | `notification_templates` | Reusable/versioned channel templates |
| Deliveries Collection | `notification_deliveries` | Per-notification delivery record/log |

### Environment Configuration Shape

```dotenv
# Local development example; real credential secret manager/.env se aayega.
NOTIFICATION_MONGO_URI=mongodb://ecommerce_root:ecommerce_password@localhost:27017/notification_db?authSource=admin
NOTIFICATION_MONGO_DATABASE=notification_db
NOTIFICATION_MONGO_TEMPLATES_COLLECTION=notification_templates
NOTIFICATION_MONGO_DELIVERIES_COLLECTION=notification_deliveries
```

### Kaise Built Hua?

| Part | Explanation |
|---|---|
| Separate `notification_db` | Har microservice apni persistence own karegi |
| Config-driven URI | Local, staging aur production endpoint code change ke bina swap ho sakta hai |
| Explicit collection config | Deployment me naming accidental mismatch avoid karna easy hota hai |
| URI secrets | Source code ya log me commit/print nahi karne hain |

---

## Step 2: `notification_templates` Collection Design Karo

Template document message ka reusable content definition rakhta hai. Task 2 me hum document persist karne ka shape define kar rahe hain; `{{name}}` jaise variables actually replace/render karna Task 3 me hoga.

### Document Example

```json
{
  "_id": "tpl_order_paid_email",
  "template_key": "order_paid",
  "channel": "email",
  "subject": "Your order is confirmed",
  "body": "Hi {{name}}, your order {{order_id}} is confirmed.",
  "status": "active",
  "version": 1,
  "created_at": "2026-05-18T00:00:00Z",
  "updated_at": "2026-05-18T00:00:00Z"
}
```

### Field Explanation

| Field | Required | Why Store It |
|---|---:|---|
| `_id` | ✅ | Template document ka stable identifier |
| `template_key` | ✅ | Business meaning, e.g. `order_paid`; future usecase isi naam se template locate karega |
| `channel` | ✅ | Task 1 channel identity; same event ka email/SMS content separate rahega |
| `subject` | Conditional | Email ke liye subject; channels jahan applicable nahi wahan empty/omitted ho sakta hai |
| `body` | ✅ | Stored message content; actual rendering later task hai |
| `status` | ✅ | Active/inactive template selection ko represent karne ke liye |
| `version` | ✅ | Content update hone par revisions distinguish karne ke liye |
| `created_at`, `updated_at` | ✅ | Audit aur debugging ke liye timestamps |

### Reference Go Domain Shape: `internal/domain/template.go`

```go
package domain

import "time"

type Template struct {
	ID          string    `bson:"_id" json:"id"`
	TemplateKey string    `bson:"template_key" json:"template_key"`
	Channel     Channel   `bson:"channel" json:"channel"`
	Subject     string    `bson:"subject,omitempty" json:"subject,omitempty"`
	Body        string    `bson:"body" json:"body"`
	Status      string    `bson:"status" json:"status"`
	Version     int       `bson:"version" json:"version"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`
}
```

### Design Notes

| Decision | Hinglish Reason |
|---|---|
| `channel` Task 1 `Channel` type reuse karega | Invalid channel vocabulary spread nahi hogi |
| `version` store karenge | Template edit se old deliveries ka context lose nahi hoga |
| `subject` optional rakha | SMS/WhatsApp-like channel me email subject force nahi karenge |
| Variables body me text form me rahenge | Rendering logic abhi add nahi ho raha; Task 3 handle karega |

---

## Step 3: `notification_deliveries` Collection Design Karo

Delivery document batata hai ki kis user ke liye kis template/channel ki delivery record bani aur uska current recorded status/provider reference kya hai. Is task me record ka shape define hota hai, actual notification send nahi hota.

### Document Example

```json
{
  "_id": "delivery_123",
  "user_id": "user_123",
  "channel": "email",
  "template_key": "order_paid",
  "status": "sent",
  "provider": "ses",
  "provider_message_id": "msg_123",
  "attempts": 1,
  "payload": {
    "order_id": "order_123"
  },
  "created_at": "2026-05-18T00:00:00Z",
  "updated_at": "2026-05-18T00:00:00Z"
}
```

### Field Explanation

| Field | Required | Why Store It |
|---|---:|---|
| `_id` | ✅ | Delivery record locate/update karne ka stable id |
| `user_id` | ✅ | User delivery history/query support |
| `channel` | ✅ | Delivery kis Task 1 channel par targeted thi |
| `template_key` | ✅ | Delivery kis business template se associated thi |
| `status` | ✅ | Stored lifecycle status ke basis par queries/support investigation |
| `provider` | As available | Kaunsa external adapter/provider selected tha, later sending stage par fill hoga |
| `provider_message_id` | As available | Provider acknowledgement/reference lookup me useful |
| `attempts` | ✅ | Recorded attempts count; retry policy Task 6 me implement hogi |
| `payload` | Optional | Event-specific non-sensitive identifiers/metadata ka flexible BSON object |
| `created_at`, `updated_at` | ✅ | Timeline and status query support |

### Reference Go Domain Shape: `internal/domain/delivery.go`

```go
package domain

import "time"

type Delivery struct {
	ID                string    `bson:"_id" json:"id"`
	UserID            string    `bson:"user_id" json:"user_id"`
	Channel           Channel   `bson:"channel" json:"channel"`
	TemplateKey       string    `bson:"template_key" json:"template_key"`
	Status            string    `bson:"status" json:"status"`
	Provider          string    `bson:"provider,omitempty" json:"provider,omitempty"`
	ProviderMessageID string    `bson:"provider_message_id,omitempty" json:"provider_message_id,omitempty"`
	Attempts          int       `bson:"attempts" json:"attempts"`
	Payload           map[string]any `bson:"payload,omitempty" json:"payload,omitempty"`
	CreatedAt         time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time `bson:"updated_at" json:"updated_at"`
}
```

`Payload` ke liye native Go map use hua hai, Mongo-specific `bson.M` nahi. Isse `domain` database driver import nahi karta aur project ka clean-layer rule preserve hota hai.

### Payload Safety Rule

```json
{
  "order_id": "order_123",
  "trace_id": "trace_123"
}
```

| Store In `payload` | Do Not Store In `payload` |
|---|---|
| Resource IDs such as `order_id` | Plain OTP code |
| Correlation/trace ID | Password/token/API key |
| Non-sensitive template inputs when necessary | Full provider secrets |
| Minimal debugging metadata | Unmasked sensitive personal content by default |

> 🔐 Notification records me PII ya security-sensitive content aa sakta hai. Principle hai: **minimum required data store karo**, secrets log mat karo, aur access service boundary ke andar restrict rakho.

---

## Step 4: MongoDB Indexes Define Karo

Indexes existing `database/mongodb-schema-design.md` se aligned hain. Inka goal templates quickly select karna aur delivery records operationally query karna hai.

### `mongosh` Index Commands

```javascript
use notification_db

db.notification_templates.createIndex(
  { template_key: 1, channel: 1, version: -1 }
)

db.notification_deliveries.createIndex(
  { user_id: 1, created_at: -1 }
)

db.notification_deliveries.createIndex(
  { status: 1, created_at: 1 }
)

db.notification_deliveries.createIndex(
  { provider: 1, provider_message_id: 1 }
)
```

### Index Usage Matrix

| Index | Future Query It Supports | Benefit |
|---|---|---|
| `{ template_key, channel, version: -1 }` | Latest version of `order_paid` email template find karna | Full scan avoid hota hai |
| `{ user_id, created_at: -1 }` | User ki recent notification history/status support query | Recent-first read efficient |
| `{ status, created_at: 1 }` | Pending/failed records timeline query | Operations aur later retry workflow ko base milta hai |
| `{ provider, provider_message_id }` | Provider reference se delivery lookup | Provider callback/support traceability |

> 🟣 **Scope note:** Index Task 2 me define hai. Retry processing, callbacks ya analytics query code later tasks me aayega.

---

## Step 5: Collection Validation Baseline Rakho

MongoDB ka document model flexible hai, phir bhi core fields missing ya wrong type hone nahi dene chahiye. Project database notes production collections par schema validation recommend karte hain.

### Reference Validation Command: Templates

```javascript
db.createCollection("notification_templates", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["_id", "template_key", "channel", "body", "status", "version", "created_at", "updated_at"],
      properties: {
        _id: { bsonType: "string" },
        template_key: { bsonType: "string" },
        channel: { enum: ["email", "sms", "push", "whatsapp_like"] },
        subject: { bsonType: "string" },
        body: { bsonType: "string" },
        status: { bsonType: "string" },
        version: { bsonType: "int", minimum: 1 },
        created_at: { bsonType: "date" },
        updated_at: { bsonType: "date" }
      }
    }
  }
})
```

### Reference Validation Command: Deliveries

```javascript
db.createCollection("notification_deliveries", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["_id", "user_id", "channel", "template_key", "status", "attempts", "created_at", "updated_at"],
      properties: {
        _id: { bsonType: "string" },
        user_id: { bsonType: "string" },
        channel: { enum: ["email", "sms", "push", "whatsapp_like"] },
        template_key: { bsonType: "string" },
        status: { bsonType: "string" },
        provider: { bsonType: "string" },
        provider_message_id: { bsonType: "string" },
        attempts: { bsonType: "int", minimum: 0 },
        payload: { bsonType: "object" },
        created_at: { bsonType: "date" },
        updated_at: { bsonType: "date" }
      }
    }
  }
})
```

### Important Timestamp Note

Documentation JSON me dates readable ISO strings dikhte hain, lekin application insert ke time Go `time.Time` MongoDB me BSON `date` ke form me store karega. Isse date ordering/index queries correct rehti hain.

---

## Step 6: Repository Boundary Define Karo

Clean architecture ke according usecase ko direct MongoDB collection methods call nahi karne chahiye. Domain/usecase interface depend karega; Mongo adapter us interface ko implement karega.

### Reference Interface: `internal/repository/notification_repository.go`

```go
package repository

import (
	"context"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
)

type TemplateRepository interface {
	FindLatestActive(ctx context.Context, templateKey string, channel domain.Channel) (domain.Template, error)
	Insert(ctx context.Context, template domain.Template) error
}

type DeliveryRepository interface {
	Insert(ctx context.Context, delivery domain.Delivery) error
	FindByID(ctx context.Context, id string) (domain.Delivery, error)
}
```

### Interface Me Kya Intentionally Nahi Hai?

| Not Added | Reason |
|---|---|
| `RenderTemplate(...)` | Rendering Task 3 ka usecase hai, repository ka kaam nahi |
| `SendOTP(...)` | Delivery execution Task 4 ka workflow hai |
| `Retry(...)` | Retry orchestration Task 6 me aayegi |
| `UpdatePreference(...)` | Preferences Task 7 ka concern hai |

---

## Step 7: Mongo Connection Aur Adapter Wiring Shape Define Karo

Actual service source is task me create nahi ho raha, lekin future implementation ka Mongo dependency boundary clear rehna chahiye.

### Reference Connection Code

```go
package repository

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoNotificationRepository struct {
	templates  *mongo.Collection
	deliveries *mongo.Collection
}

func NewMongoNotificationRepository(ctx context.Context, uri, database string) (*MongoNotificationRepository, *mongo.Client, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, nil, fmt.Errorf("connect mongo: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, nil, fmt.Errorf("ping mongo: %w", err)
	}

	db := client.Database(database)
	repo := &MongoNotificationRepository{
		templates:  db.Collection("notification_templates"),
		deliveries: db.Collection("notification_deliveries"),
	}
	return repo, client, nil
}
```

### Reference Latest Template Read

```go
func (r *MongoNotificationRepository) FindLatestActive(
	ctx context.Context,
	templateKey string,
	channel domain.Channel,
) (domain.Template, error) {
	filter := bson.D{
		{Key: "template_key", Value: templateKey},
		{Key: "channel", Value: channel},
		{Key: "status", Value: "active"},
	}
	opts := options.FindOne().SetSort(bson.D{{Key: "version", Value: -1}})

	var result domain.Template
	if err := r.templates.FindOne(ctx, filter, opts).Decode(&result); err != nil {
		return domain.Template{}, err
	}
	return result, nil
}
```

### Is Layering Ka Benefit

| Layer | Responsibility |
|---|---|
| `domain` | Template/delivery meaning and channel type |
| `repository` interface | Business-facing persistence operations |
| Mongo repository adapter | BSON, Mongo client, query/index-compatible filters |
| Future usecase | Template read ya delivery create kab karna hai decide karega |

> 🟡 `FindLatestActive` yahan persistence read pattern demonstrate karta hai. Returned body ko variables ke saath render karna Task 3 me hi implement hoga.

---

## Step 8: Insert And Lookup Examples Samjho

Ye commands developer ko document shape verify karne me help karenge jab actual Mongo environment later setup hoga.

### Template Insert Example

```javascript
use notification_db

db.notification_templates.insertOne({
  _id: "tpl_order_paid_email",
  template_key: "order_paid",
  channel: "email",
  subject: "Your order is confirmed",
  body: "Hi {{name}}, your order {{order_id}} is confirmed.",
  status: "active",
  version: NumberInt(1),
  created_at: new Date("2026-05-18T00:00:00Z"),
  updated_at: new Date("2026-05-18T00:00:00Z")
})
```

### Latest Template Lookup Example

```javascript
db.notification_templates.find({
  template_key: "order_paid",
  channel: "email",
  status: "active"
}).sort({ version: -1 }).limit(1)
```

### Delivery Insert Example

```javascript
db.notification_deliveries.insertOne({
  _id: "delivery_123",
  user_id: "user_123",
  channel: "email",
  template_key: "order_paid",
  status: "sent",
  provider: "ses",
  provider_message_id: "msg_123",
  attempts: NumberInt(1),
  payload: { order_id: "order_123" },
  created_at: new Date("2026-05-18T00:00:00Z"),
  updated_at: new Date("2026-05-18T00:00:00Z")
})
```

### User Delivery Lookup Example

```javascript
db.notification_deliveries.find({
  user_id: "user_123"
}).sort({ created_at: -1 }).limit(20)
```

> 🟢 Commands schema/index behaviour demonstrate karte hain. Ye guide execution ka evidence nahi hai, kyunki Task 2 deliverable documentation-only hai.

---

## 🔄 Storage Flow Diagram

```mermaid
sequenceDiagram
    participant Future as Future Usecase
    participant Repo as Repository Interface
    participant Mongo as MongoDB notification_db

    Future->>Repo: FindLatestActive(template_key, channel)
    Repo->>Mongo: Query notification_templates + version sort
    Mongo-->>Repo: Template document
    Repo-->>Future: Domain Template

    Note over Future: Rendering/sending is later task scope

    Future->>Repo: Insert(Delivery record)
    Repo->>Mongo: Insert notification_deliveries document
    Mongo-->>Repo: Acknowledged
    Repo-->>Future: Stored
```

### Collection Relationship

```mermaid
erDiagram
    NOTIFICATION_TEMPLATES {
        string _id
        string template_key
        string channel
        string body
        string status
        int version
        date created_at
    }
    NOTIFICATION_DELIVERIES {
        string _id
        string user_id
        string channel
        string template_key
        string status
        string provider_message_id
        object payload
        date created_at
    }
    NOTIFICATION_TEMPLATES ||--o{ NOTIFICATION_DELIVERIES : "identified by template_key + channel"
```

**Hinglish explanation:** MongoDB me relational foreign key enforce nahi hota. Delivery me `template_key` aur `channel` traceability ke liye stored rahenge; template selection correctness service layer enforce karegi.

---

## 🔐 Security And Data Handling

| Rule | Implementation Guidance |
|---|---|
| Credentials secure rakho | Mongo URI ko environment secret/secret manager se inject karo; source ya logs me print mat karo |
| Least privilege | Notification Service ke DB user ko sirf `notification_db` ke required permissions do |
| PII minimize karo | Payload me unnecessary email/phone/body copies avoid karo; required ho to mask/encrypt policy follow karo |
| OTP secret avoid karo | Plain OTP ko delivery payload ya log me persist nahi karna |
| Provider response control | Raw provider data future me store ho to secrets/token redact karke store karo |
| Access boundary | Other services Notification DB directly access na karein |
| Backups and reliability | Production me Mongo replica set, backups aur recovery policy use karo |
| Document growth control | Large/unbounded arrays avoid karo; each delivery separate document rakho |

### Do And Avoid

| ✅ Do | ❌ Avoid |
|---|---|
| `payload: { "order_id": "order_123" }` | `payload: { "otp": "123456" }` |
| Queryable typed timestamps store karo | Date ko inconsistent plain display strings ke form me write karna |
| Task 1 canonical `channel` validate karo | Arbitrary vendor-specific channel strings insert karna |
| Index-supported query patterns use karo | Delivery history ke liye full collection scan |

---

## 🛠️ External Libraries And Tools

### Is Deliverable Me Actual Usage

| Library / Tool | Used In This Deliverable? | Why |
|---|---:|---|
| Markdown | ✅ Yes | Structured implementation guide likhne ke liye |
| Mermaid | ✅ Yes | Architecture, flow aur collection relationship diagrams ke liye |
| Shields.io badges | ✅ Yes | Status, scope aur database choice visually readable banane ke liye |
| MongoDB server | ❌ Reference only | Actual database run karna source/runtime implementation phase ka kaam hai |
| MongoDB Go Driver v2 | ❌ Reference snippets only | Future Go repository ko MongoDB se connect/query karne ke liye |
| `mongosh` | ❌ Reference commands only | Future local schema/index inspection ke liye |

### MongoDB Kya Hai Aur Kyon Use Hoga?

**MongoDB** document database hai jo BSON documents store karta hai. Notification template aur delivery payload channel/provider ke basis par vary kar sakte hain, isliye project docs ne Notification Service ke liye MongoDB select kiya hai.

#### Future Local Use With Existing Project Convention

Platform Foundation local-stack guide MongoDB ko Docker Compose dependency ke form me define karti hai. Jab actual stack file implement ho:

```bash
docker compose -f infra/compose/docker-compose.local.yml up -d mongo
docker exec -it ecommerce-mongo mongosh
```

Use:

```javascript
use notification_db
show collections
db.notification_templates.getIndexes()
db.notification_deliveries.getIndexes()
```

> Is Task 2 output me compose file ya Mongo container create/run nahi kiya gaya.

### Official MongoDB Go Driver v2

**What:** Go application ke liye official MongoDB client library.  
**Why:** Future `mongo_notification_repository.go` ko connect, insert, query aur index-compatible operations execute karne ke liye.  
**Install later, notification service module ke andar:**

```bash
go get go.mongodb.org/mongo-driver/v2/mongo
```

**Use in code:**

```go
import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)
```

> 🟣 Driver command official MongoDB Go Driver v2 documentation ke according diya gaya hai. Dependency abhi install nahi ki gayi because requested output documentation artifact tak limited hai.

### `mongosh`

**What:** MongoDB shell CLI.  
**Why:** Collections create karna, index verify karna, test documents insert/read karna aur local troubleshooting.  
**How to use:** Local Mongo container me installed shell ho to `docker exec -it ecommerce-mongo mongosh` run karke above JavaScript commands execute karo. Host install optional hai.

### Markdown Visual Tools

| Tool | Install Needed? | Use |
|---|---:|---|
| Mermaid | No backend install | Supported Markdown viewer fenced `mermaid` blocks render karta hai |
| Shields.io | No install | Badge image Markdown URL se display hoti hai; offline mode me image unavailable ho sakti hai |

### Official References

- [MongoDB Go Driver v2 - Getting Started](https://www.mongodb.com/docs/drivers/go/v2.0/quick-start/)
- [MongoDB Community With Docker](https://www.mongodb.com/docs/v7.0/tutorial/install-mongodb-community-with-docker/)

---

## 🧪 Validation And Testing Strategy

Actual Go/Mongo implementation add hone ke baad ye focused checks run hone chahiye:

### Schema/Index Checks

| Check | Expected Result |
|---|---|
| Valid email template insert | Insert succeeds |
| Unknown channel template insert | Collection validation rejects it |
| Missing `template_key` | Collection validation rejects it |
| Valid delivery with small payload | Insert succeeds |
| Delivery with unknown channel | Collection validation rejects it |
| Template index list | `{ template_key: 1, channel: 1, version: -1 }` present |
| Delivery index list | Three documented query indexes present |

### Repository Tests To Add When Source Exists

| Test | Purpose |
|---|---|
| `FindLatestActive` sorts version descending | Newest active channel template return hota hai |
| Insert/read delivery preserves `payload` | Flexible BSON metadata safely round-trip hoti hai |
| Context deadline/cancellation test | DB request indefinitely block nahi karta |
| Mongo unavailable error test | Repository useful wrapped error return karta hai |
| Channel validation test | Task 1 values ke outside data reject hota hai |

### Future Verification Commands

```bash
# Only after notification-service Go source/module exists:
go test ./backend/services/notification-service/...

# Only after the local Mongo stack is implemented and running:
docker exec ecommerce-mongo mongosh --quiet --eval "db.adminCommand({ ping: 1 })"
```

> 🟡 Current deliverable documentation-only hai, isliye koi Go tests, DB connection, collection creation ya index command actually run nahi kiya gaya.

---

## ⚙️ Operational Notes

| Concern | Task 2 Guidance |
|---|---|
| Database availability | Production me replica set/managed Mongo deployment aur backups configure karo |
| Index changes | Collection launch se pehle documented indexes create karo; heavy production index builds plan karke run karo |
| Payload size | Provider/raw payload blindly store mat karo; Mongo document size limit aur privacy dono respect karo |
| Status queries | `status + created_at` index later operational reads/retry selection ko support karega |
| Template history | Version field overwrite ke badle historical revisions preserve karne ka path deta hai |
| Observability | Connection errors/latency log/metric me expose karo, lekin URI/payload secrets redact karo |

---

## ✅ Definition Of Done Checklist

| Check | Status |
|---|---:|
| Notification Service Task 2 requirement identify kiya | ✅ |
| Task 1 channel dependency respect ki | ✅ |
| MongoDB database choice explain ki | ✅ |
| Service ownership boundary clear ki | ✅ |
| `notification_db` name documented kiya | ✅ |
| `notification_templates` schema documented ki | ✅ |
| `notification_deliveries` schema documented ki | ✅ |
| Existing project-defined indexes included kiye | ✅ |
| Collection validation baseline explain ki | ✅ |
| Clean future folder structure provided ki | ✅ |
| Go repository/reference snippets provided kiye | ✅ |
| Mermaid diagrams and visual badges included kiye | ✅ |
| MongoDB/Go Driver/`mongosh` usage and installation explained ki | ✅ |
| PII/secrets storage safety rules documented kiye | ✅ |
| Template engine, sending, events, retries, preferences, analytics excluded rakhe | ✅ |
| Actual output sirf `task2.md` artifact tak limited rakha | ✅ |

---

## 🎯 Final Result

Notification Service - Task 2 ke liye MongoDB persistence blueprint ready hai:

- Service-owned database `notification_db` select hua.
- `notification_templates` channel-specific versioned template storage ke liye defined hai.
- `notification_deliveries` flexible delivery log metadata ke liye defined hai.
- Project schema ke exact query indexes aur validation guidance documented hain.
- Future Go repository ko MongoDB Go Driver v2 ke through cleanly wire karne ka reference shape available hai.
- OTP sending, rendering, event handling, retries, preferences aur analytics ko intentionally later tasks ke liye untouched rakha gaya hai.

Is decision ke baad Task 3 template rendering behaviour ko isi persisted template foundation par safely build kar sakta hai.
