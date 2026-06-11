# 💳 Payment Service - Task 2: Create MySQL Schema

![Task](https://img.shields.io/badge/Task-Payment%20Service%20%232-blue)
![Priority](https://img.shields.io/badge/Priority-P0-red)
![Dependency](https://img.shields.io/badge/Dependency-State%20Machine-yellow)
![Status](https://img.shields.io/badge/Status-Documented-success)
![Guide](https://img.shields.io/badge/Guide-Hinglish-orange)
![Scope](https://img.shields.io/badge/Scope-MySQL%20Schema-purple)

## 📌 Task Summary

| Field | Detail |
|---|---|
| Task Name | Create MySQL schema |
| Source | `docs/01-micro-tasks.md` → `Payment Service` → Task 2 |
| Priority | `P0` foundation/blocker |
| Dependency | Payment Service Task 1: State machine |
| Main Goal | Payment financial records ke liye auditable, consistent MySQL schema define karna |
| Output Type | Documentation-only implementation guide |
| Not Included | Provider SDK integration, payment intent API, webhook handler code, refund business flow, retry policy, reconciliation job implementation |

> **Simple Hinglish goal:** Payment Service money-related data handle karega, isliye schema strongly consistent, auditable, idempotent aur reconciliation-friendly hona chahiye. Task 2 me hum `payment_db` ke tables, columns, indexes, constraints aur migration strategy define karte hain.

---

## ✅ Final Output Created

```text
TaskImplementation/
└── Payment Service/
    ├── task1.md
    └── task2.md
```

### Why this structure?

| Path | Purpose |
|---|---|
| `TaskImplementation/` | Saare task-wise implementation guides ka central folder |
| `TaskImplementation/Payment Service/` | Payment Service related task guides ka group |
| `task2.md` | Sirf **Payment Service - Task 2** ka complete guide |

> 🟢 **Boundary:** Is task me actual provider integration, runtime API code, webhook handler, ya new migration files create nahi kiye gaye. User request ka output only `task2.md` hai, isliye schema implementation guide yahin documented hai.

---

## 🧭 Source Documents Studied

| Document | Kya use kiya gaya |
|---|---|
| `docs/01-micro-tasks.md` | Payment Service Task 2 ka exact scope: `Create MySQL schema` |
| `docs/07-payment-system.md` | Payment states, idempotency, webhook, refund, reconciliation rules |
| `docs/04-microservice-design.md` | Payment Service tables, APIs, internal logic boundaries |
| `docs/05-database-design.md` | Payment DB strategy, relationships, indexes, scalability notes |
| `database/draw.sql` | Existing master SQL DDL for `payment_db` |
| `docs/03-folder-structure.md` | Future service migration folder conventions |
| `docs/06-auth-security.md` | Card data aur secrets store na karne ka rule |
| `docs/13-developer-guide.md` | Migration guide and DB change rules |

---

## 🪜 Step-by-Step Implementation

## Step 1: Task Boundary Clear Kiya

Task 2 ka focus **Payment Service ke MySQL schema** par hai. Schema ka kaam hai financial data ko safe, queryable aur audit-friendly format me store karna.

### Included

- `payment_db` database define kiya
- Core payment tables define kiye
- Payment attempt tracking define kiya
- Refund record schema define kiya
- Webhook event idempotency schema define kiya
- Reconciliation result schema define kiya
- Primary keys, unique keys, indexes aur foreign keys explain kiye
- Migration strategy and verification commands document kiye

### Not Included

- `CreatePaymentIntent` API implementation
- Stripe/Razorpay SDK integration
- Webhook signature verification code
- Refund approval business workflow
- Retry limits and retry policy implementation
- Daily reconciliation worker implementation
- Superadmin payment UI/API

> 🟡 **Reason:** Ye sab Payment Service ke later tasks me aayega. Task 2 sirf database foundation banata hai.

---

## Step 2: MySQL Choose Kiya

Payment data financial hota hai, isliye MySQL choose kiya gaya.

| Reason | Explanation |
|---|---|
| Strong consistency | Payment status, captured amount, refund amount mismatch nahi hona chahiye |
| Transactions | Payment + attempt + webhook updates atomic transaction me ho sakte hain |
| Indexes | Order lookup, provider lookup, refund lookup fast hoga |
| Auditability | Created/updated timestamps, immutable event rows, reconciliation rows maintain ho sakte hain |
| Reporting | Admin payment operations aur finance reconciliation queries easy hoti hain |

> 🔴 **Important:** Card number, CVV, raw secret keys, OTP ya sensitive provider secrets schema me store nahi honge. Provider hosted flow/client secret approach use hoga.

---

## Step 3: Database Ownership Define Kiya

Payment Service apna separate database own karega:

```sql
CREATE DATABASE IF NOT EXISTS payment_db
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

USE payment_db;
```

### Ownership rules

| Rule | Meaning |
|---|---|
| Service-owned DB | Payment tables sirf Payment Service write karega |
| No cross-service FK | `order_id`, `user_id` references store honge but Order/User DB FK nahi banega |
| Provider IDs stored | Provider reconciliation ke liye provider payment/refund/event IDs store honge |
| Amounts as integer | Paisa minor unit me store hoga, jaise paise/cents |
| Webhook idempotency | Duplicate provider webhook repeated update nahi karega |

---

## Step 4: Tables Identify Kiye

Task 2 ke liye 5 core tables define kiye gaye:

| Table | Purpose |
|---|---|
| `payments` | Main payment aggregate: order, user, provider, amount, status |
| `payment_attempts` | Har provider attempt ka audit trail |
| `refunds` | Full/partial refund requests and outcomes |
| `payment_webhook_events` | Provider webhook payload + idempotency tracking |
| `payment_reconciliations` | Settlement report compare results |

### Relationship overview

```mermaid
erDiagram
    payments ||--o{ payment_attempts : has
    payments ||--o{ refunds : has
    payment_webhook_events {
        bigint id PK
        varchar webhook_event_id UK
        varchar provider
        varchar provider_event_id UK
        varchar event_type
        boolean processed
        json payload
        timestamp received_at
        timestamp processed_at
    }
    payment_reconciliations {
        bigint id PK
        varchar reconciliation_id UK
        varchar provider
        varchar settlement_id
        enum status
        varchar payment_id
        varchar provider_payment_id
        json details
        timestamp created_at
    }
    payments {
        bigint id PK
        varchar payment_id UK
        varchar order_id
        varchar user_id
        varchar provider
        varchar provider_payment_id
        varchar provider_intent_id
        enum status
        char currency
        bigint amount
        bigint captured_amount
        bigint refunded_amount
        varchar idempotency_key
        timestamp created_at
        timestamp updated_at
    }
    payment_attempts {
        bigint id PK
        varchar attempt_id UK
        varchar payment_id FK
        varchar provider_attempt_id
        enum status
        json raw_provider_response
        timestamp created_at
    }
    refunds {
        bigint id PK
        varchar refund_id UK
        varchar payment_id FK
        varchar provider_refund_id
        enum status
        char currency
        bigint amount
        varchar idempotency_key
        timestamp created_at
        timestamp updated_at
    }
```

---

## Step 5: `payments` Table Build Kiya

`payments` table Payment Service ka main financial record hai.

### Design decisions

| Column group | Why needed |
|---|---|
| `payment_id` | Public/internal stable ID, APIs/events me use hoga |
| `order_id`, `user_id` | Order aur buyer ke saath payment map karne ke liye |
| `provider` | Stripe/Razorpay-like provider identify karne ke liye |
| `provider_payment_id`, `provider_intent_id` | Provider side record se local record match karne ke liye |
| `status` | Task 1 state machine ka persisted payment status |
| `amount`, `captured_amount`, `refunded_amount` | Financial totals minor unit me |
| `idempotency_key` | Duplicate payment intent prevent karne ke liye |
| `failure_code`, `failure_message` | Failure troubleshooting ke liye safe provider details |

### Persisted statuses

| Status | Meaning |
|---|---|
| `initiated` | Payment create hua, provider intent/attempt start ho raha hai |
| `requires_action` | User action needed: OTP/3DS/UPI approval/redirect |
| `authorized` | Provider ne amount authorize kar diya |
| `captured` | Money successfully captured |
| `failed` | Payment attempt failed/expired/cancelled |
| `refunded` | Full refund complete |
| `partially_refunded` | Partial refund complete |

> 🟣 **Note:** Task 1 me `retry_allowed` helper state mention hai. Schema level par retry ko new `payment_attempts` row aur future retry policy se handle kiya jayega. Isliye current persisted `payments.status` enum me `retry_allowed` store nahi kiya gaya.

### SQL

```sql
CREATE TABLE IF NOT EXISTS payments (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  payment_id VARCHAR(64) NOT NULL,
  order_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  provider VARCHAR(64) NOT NULL,
  provider_payment_id VARCHAR(128) NULL,
  provider_intent_id VARCHAR(128) NULL,
  status ENUM(
    'initiated',
    'requires_action',
    'authorized',
    'captured',
    'failed',
    'refunded',
    'partially_refunded'
  ) NOT NULL DEFAULT 'initiated',
  currency CHAR(3) NOT NULL,
  amount BIGINT NOT NULL,
  captured_amount BIGINT NOT NULL DEFAULT 0,
  refunded_amount BIGINT NOT NULL DEFAULT 0,
  idempotency_key VARCHAR(128) NOT NULL,
  failure_code VARCHAR(128) NULL,
  failure_message VARCHAR(512) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_payments_payment_id (payment_id),
  UNIQUE KEY uk_payments_idempotency (provider, idempotency_key),
  KEY idx_payments_order (order_id),
  KEY idx_payments_provider_payment (provider, provider_payment_id),
  KEY idx_payments_status_created (status, created_at)
) ENGINE=InnoDB;
```

### Important indexes

| Index | Why used |
|---|---|
| `uk_payments_payment_id` | Payment lookup by stable ID |
| `uk_payments_idempotency` | Same provider + same idempotency key duplicate payment block karta hai |
| `idx_payments_order` | Order detail page/payment list ke liye |
| `idx_payments_provider_payment` | Provider webhook/reconciliation lookup ke liye |
| `idx_payments_status_created` | Admin filters and pending/failed jobs ke liye |

---

## Step 6: `payment_attempts` Table Build Kiya

Ek order/payment ke liye multiple attempts ho sakte hain. Example: first UPI attempt failed, second card attempt success.

### Why separate table?

| Need | Explanation |
|---|---|
| Audit | Failed attempts bhi preserve honge |
| Retry support | Same payment/order ke andar new attempt track ho sakta hai |
| Provider debugging | Provider response safely inspect ki ja sakti hai |
| No overwrite | Previous attempt data lost nahi hota |

### SQL

```sql
CREATE TABLE IF NOT EXISTS payment_attempts (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  attempt_id VARCHAR(64) NOT NULL,
  payment_id VARCHAR(64) NOT NULL,
  provider_attempt_id VARCHAR(128) NULL,
  status ENUM('initiated', 'succeeded', 'failed') NOT NULL DEFAULT 'initiated',
  failure_code VARCHAR(128) NULL,
  failure_message VARCHAR(512) NULL,
  raw_provider_response JSON NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_payment_attempts_attempt_id (attempt_id),
  KEY idx_payment_attempts_payment (payment_id),
  CONSTRAINT fk_payment_attempts_payment
    FOREIGN KEY (payment_id) REFERENCES payments(payment_id)
) ENGINE=InnoDB;
```

### Field explanation

| Field | Use |
|---|---|
| `attempt_id` | Internal attempt ID |
| `payment_id` | Parent payment aggregate |
| `provider_attempt_id` | Provider side attempt/session ID |
| `status` | Attempt-level result |
| `raw_provider_response` | Sanitized provider response for audit/debug |

> 🔴 **Security rule:** `raw_provider_response` me card number, CVV, raw token, API secret, webhook secret ya PII-heavy payload store nahi karna.

---

## Step 7: `refunds` Table Build Kiya

Refunds immutable/auditable records me track honge. Full aur partial dono refunds support honge.

### Refund states

| Status | Meaning |
|---|---|
| `requested` | Refund request create hui |
| `approved` | Finance/admin/system ne approve kiya |
| `rejected` | Refund reject hui |
| `processing` | Provider refund API call/process chal raha hai |
| `succeeded` | Provider ne refund success confirm ki |
| `failed` | Provider refund fail hui |

### SQL

```sql
CREATE TABLE IF NOT EXISTS refunds (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  refund_id VARCHAR(64) NOT NULL,
  payment_id VARCHAR(64) NOT NULL,
  provider_refund_id VARCHAR(128) NULL,
  status ENUM(
    'requested',
    'approved',
    'rejected',
    'processing',
    'succeeded',
    'failed'
  ) NOT NULL DEFAULT 'requested',
  currency CHAR(3) NOT NULL,
  amount BIGINT NOT NULL,
  reason VARCHAR(512) NOT NULL,
  requested_by VARCHAR(64) NOT NULL,
  reviewed_by VARCHAR(64) NULL,
  reviewed_at TIMESTAMP NULL,
  idempotency_key VARCHAR(128) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_refunds_refund_id (refund_id),
  UNIQUE KEY uk_refunds_idempotency (payment_id, idempotency_key),
  KEY idx_refunds_payment_status (payment_id, status),
  CONSTRAINT fk_refunds_payment
    FOREIGN KEY (payment_id) REFERENCES payments(payment_id)
) ENGINE=InnoDB;
```

### Refund safety rules

| Rule | Why |
|---|---|
| Refund only captured payment par | Failed/initiated payment refund nahi ho sakta |
| `amount <= captured_amount - refunded_amount` | Over-refund prevent karna |
| Same `payment_id + idempotency_key` unique | Duplicate refund request block karna |
| `requested_by`, `reviewed_by` store | Audit and admin accountability |

---

## Step 8: `payment_webhook_events` Table Build Kiya

Provider webhook source of truth hoga. Duplicate webhook common hote hain, isliye idempotency mandatory hai.

### Why this table is critical?

| Problem | Table ka solution |
|---|---|
| Provider duplicate event bhej sakta hai | `(provider, provider_event_id)` unique key duplicate block karta hai |
| Webhook processing fail ho sakti hai | `processed=false` events retry job process kar sakta hai |
| Audit chahiye | Raw sanitized payload store hota hai |
| Debugging chahiye | `event_type`, `received_at`, `processed_at` trace dete hain |

### SQL

```sql
CREATE TABLE IF NOT EXISTS payment_webhook_events (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  webhook_event_id VARCHAR(64) NOT NULL,
  provider VARCHAR(64) NOT NULL,
  provider_event_id VARCHAR(128) NOT NULL,
  event_type VARCHAR(128) NOT NULL,
  processed BOOLEAN NOT NULL DEFAULT FALSE,
  payload JSON NOT NULL,
  received_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  processed_at TIMESTAMP NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_webhook_provider_event (provider, provider_event_id),
  UNIQUE KEY uk_webhook_event_id (webhook_event_id),
  KEY idx_webhook_processed_received (processed, received_at)
) ENGINE=InnoDB;
```

### Webhook idempotency flow

```mermaid
sequenceDiagram
    participant Provider as Payment Provider
    participant Payment as Payment Service
    participant DB as payment_db

    Provider->>Payment: POST webhook event
    Payment->>Payment: Verify signature
    Payment->>DB: Insert webhook_event(provider, provider_event_id)
    alt Duplicate provider_event_id
        DB-->>Payment: Unique key conflict
        Payment-->>Provider: 200 OK no-op
    else New event
        DB-->>Payment: Stored
        Payment->>DB: Update payment/refund in transaction
        Payment->>DB: Mark webhook processed
        Payment-->>Provider: 200 OK
    end
```

> 🟢 **Rule:** Duplicate webhook error nahi hota. Agar event already processed hai to Payment Service provider ko success response dekar no-op karega.

---

## Step 9: `payment_reconciliations` Table Build Kiya

Reconciliation ka kaam provider settlement report ko local payment records ke saath compare karna hai.

### Reconciliation statuses

| Status | Meaning |
|---|---|
| `matched` | Provider aur local payment data same hai |
| `mismatch` | Local/provider amount, currency, status, fee, settlement data mismatch hai |
| `missing_local` | Provider me payment hai but local DB me nahi mila |
| `missing_provider` | Local DB me payment hai but provider settlement me nahi mila |

### SQL

```sql
CREATE TABLE IF NOT EXISTS payment_reconciliations (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  reconciliation_id VARCHAR(64) NOT NULL,
  provider VARCHAR(64) NOT NULL,
  settlement_id VARCHAR(128) NULL,
  status ENUM(
    'matched',
    'mismatch',
    'missing_local',
    'missing_provider'
  ) NOT NULL,
  payment_id VARCHAR(64) NULL,
  provider_payment_id VARCHAR(128) NULL,
  details JSON NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_reconciliation_id (reconciliation_id),
  KEY idx_reconciliation_provider_status (provider, status)
) ENGINE=InnoDB;
```

### Reconciliation batch flow

```mermaid
flowchart TD
    A[Download provider settlement report] --> B[Read provider payment rows]
    B --> C[Find local payment by provider + provider_payment_id]
    C --> D{Matched?}
    D -->|Yes| E[Insert status matched]
    D -->|No local row| F[Insert status missing_local]
    D -->|Amount/status mismatch| G[Insert status mismatch]
    D -->|Local not in provider report| H[Insert status missing_provider]
    F --> I[Finance alert]
    G --> I
    H --> I
```

> 🟡 **Task boundary:** Reconciliation job actual code Task 8 me implement hoga. Task 2 me us job ke result store karne ke liye schema ready kiya gaya.

---

## Step 10: Complete Payment Schema

Yeh Task 2 ka complete schema hai, aligned with `database/draw.sql`.

```sql
CREATE DATABASE IF NOT EXISTS payment_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE payment_db;

CREATE TABLE IF NOT EXISTS payments (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  payment_id VARCHAR(64) NOT NULL,
  order_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  provider VARCHAR(64) NOT NULL,
  provider_payment_id VARCHAR(128) NULL,
  provider_intent_id VARCHAR(128) NULL,
  status ENUM('initiated', 'requires_action', 'authorized', 'captured', 'failed', 'refunded', 'partially_refunded') NOT NULL DEFAULT 'initiated',
  currency CHAR(3) NOT NULL,
  amount BIGINT NOT NULL,
  captured_amount BIGINT NOT NULL DEFAULT 0,
  refunded_amount BIGINT NOT NULL DEFAULT 0,
  idempotency_key VARCHAR(128) NOT NULL,
  failure_code VARCHAR(128) NULL,
  failure_message VARCHAR(512) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_payments_payment_id (payment_id),
  UNIQUE KEY uk_payments_idempotency (provider, idempotency_key),
  KEY idx_payments_order (order_id),
  KEY idx_payments_provider_payment (provider, provider_payment_id),
  KEY idx_payments_status_created (status, created_at)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS payment_attempts (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  attempt_id VARCHAR(64) NOT NULL,
  payment_id VARCHAR(64) NOT NULL,
  provider_attempt_id VARCHAR(128) NULL,
  status ENUM('initiated', 'succeeded', 'failed') NOT NULL DEFAULT 'initiated',
  failure_code VARCHAR(128) NULL,
  failure_message VARCHAR(512) NULL,
  raw_provider_response JSON NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_payment_attempts_attempt_id (attempt_id),
  KEY idx_payment_attempts_payment (payment_id),
  CONSTRAINT fk_payment_attempts_payment FOREIGN KEY (payment_id) REFERENCES payments(payment_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS refunds (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  refund_id VARCHAR(64) NOT NULL,
  payment_id VARCHAR(64) NOT NULL,
  provider_refund_id VARCHAR(128) NULL,
  status ENUM('requested', 'approved', 'rejected', 'processing', 'succeeded', 'failed') NOT NULL DEFAULT 'requested',
  currency CHAR(3) NOT NULL,
  amount BIGINT NOT NULL,
  reason VARCHAR(512) NOT NULL,
  requested_by VARCHAR(64) NOT NULL,
  reviewed_by VARCHAR(64) NULL,
  reviewed_at TIMESTAMP NULL,
  idempotency_key VARCHAR(128) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_refunds_refund_id (refund_id),
  UNIQUE KEY uk_refunds_idempotency (payment_id, idempotency_key),
  KEY idx_refunds_payment_status (payment_id, status),
  CONSTRAINT fk_refunds_payment FOREIGN KEY (payment_id) REFERENCES payments(payment_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS payment_webhook_events (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  webhook_event_id VARCHAR(64) NOT NULL,
  provider VARCHAR(64) NOT NULL,
  provider_event_id VARCHAR(128) NOT NULL,
  event_type VARCHAR(128) NOT NULL,
  processed BOOLEAN NOT NULL DEFAULT FALSE,
  payload JSON NOT NULL,
  received_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  processed_at TIMESTAMP NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_webhook_provider_event (provider, provider_event_id),
  UNIQUE KEY uk_webhook_event_id (webhook_event_id),
  KEY idx_webhook_processed_received (processed, received_at)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS payment_reconciliations (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  reconciliation_id VARCHAR(64) NOT NULL,
  provider VARCHAR(64) NOT NULL,
  settlement_id VARCHAR(128) NULL,
  status ENUM('matched', 'mismatch', 'missing_local', 'missing_provider') NOT NULL,
  payment_id VARCHAR(64) NULL,
  provider_payment_id VARCHAR(128) NULL,
  details JSON NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_reconciliation_id (reconciliation_id),
  KEY idx_reconciliation_provider_status (provider, status)
) ENGINE=InnoDB;
```

---

## Step 11: Migration Structure Plan

Actual user output only `task2.md` hai. Future implementation me Payment Service migrations ka structure ye rahega:

```text
backend/
└── services/
    └── payment-service/
        ├── internal/
        │   ├── domain/
        │   │   ├── payment.go
        │   │   ├── refund.go
        │   │   └── webhook_event.go
        │   └── repository/
        │       └── mysql_payment_repository.go
        └── migrations/
            ├── 001_create_payment_tables.up.sql
            └── 001_create_payment_tables.down.sql
```

### Up migration

`001_create_payment_tables.up.sql` me Step 10 wala complete DDL rahega.

### Down migration

Down migration reverse dependency order me tables drop karega:

```sql
DROP TABLE IF EXISTS payment_reconciliations;
DROP TABLE IF EXISTS payment_webhook_events;
DROP TABLE IF EXISTS refunds;
DROP TABLE IF EXISTS payment_attempts;
DROP TABLE IF EXISTS payments;
```

> 🟡 **Why reverse order?** `refunds` aur `payment_attempts` tables `payments(payment_id)` par foreign key rakhti hain. Parent table drop karne se pehle child tables drop karna safe hota hai.

---

## Step 12: External Libraries / Tools

Task 2 me koi new Go package install nahi kiya gaya. Schema ke liye recommended tools ye hain:

| Tool | What | Why used | Install/use |
|---|---|---|---|
| MySQL 8+ | Relational database | Financial consistency, transactions, indexes, audit records | Docker ya local package se run karo |
| MySQL CLI | DB command-line client | Schema apply/verify karne ke liye | `mysql -h 127.0.0.1 -u root -p < file.sql` |
| Docker Compose | Local dependency runner | MySQL local stack fast start karne ke liye | `docker compose up -d mysql` |
| golang-migrate | Migration runner | Versioned up/down migrations safely run karne ke liye | `go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest` |

### MySQL via Docker example

```bash
docker run --name ecommerce-mysql \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=payment_db \
  -p 3306:3306 \
  -d mysql:8
```

### Apply schema with MySQL CLI

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p < backend/services/payment-service/migrations/001_create_payment_tables.up.sql
```

### Apply schema with golang-migrate

```bash
migrate \
  -path backend/services/payment-service/migrations \
  -database "mysql://root:root@tcp(127.0.0.1:3306)/payment_db?multiStatements=true" \
  up
```

> 🟣 **Note:** Commands future migration execution ke liye hain. Is documentation task me tools install/run nahi kiye gaye.

---

## Step 13: Query Examples

### Create a payment

```sql
INSERT INTO payments (
  payment_id,
  order_id,
  user_id,
  provider,
  provider_intent_id,
  status,
  currency,
  amount,
  idempotency_key
) VALUES (
  'pay_1001',
  'ord_1001',
  'usr_1001',
  'stripe_like',
  'pi_provider_1001',
  'initiated',
  'INR',
  129900,
  'payment_intent:ord_1001:attempt_1'
);
```

### Add payment attempt

```sql
INSERT INTO payment_attempts (
  attempt_id,
  payment_id,
  provider_attempt_id,
  status,
  raw_provider_response
) VALUES (
  'pat_1001',
  'pay_1001',
  'provider_attempt_1001',
  'initiated',
  JSON_OBJECT('provider_status', 'requires_action')
);
```

### Mark payment captured

```sql
UPDATE payments
SET
  status = 'captured',
  provider_payment_id = 'provider_payment_1001',
  captured_amount = amount,
  failure_code = NULL,
  failure_message = NULL
WHERE payment_id = 'pay_1001'
  AND status IN ('initiated', 'requires_action', 'authorized');
```

### Store webhook event idempotently

```sql
INSERT INTO payment_webhook_events (
  webhook_event_id,
  provider,
  provider_event_id,
  event_type,
  payload
) VALUES (
  'wh_1001',
  'stripe_like',
  'evt_provider_1001',
  'payment.captured',
  JSON_OBJECT('payment_id', 'provider_payment_1001')
);
```

### Create refund request

```sql
INSERT INTO refunds (
  refund_id,
  payment_id,
  status,
  currency,
  amount,
  reason,
  requested_by,
  idempotency_key
) VALUES (
  'ref_1001',
  'pay_1001',
  'requested',
  'INR',
  50000,
  'Customer requested partial refund',
  'admin_1001',
  'refund:pay_1001:item_1001'
);
```

---

## Step 14: Go Repository Example

Future Payment Service repository me insert operation transaction + idempotency aware hona chahiye.

```go
type Payment struct {
    PaymentID        string
    OrderID          string
    UserID           string
    Provider         string
    ProviderIntentID string
    Status           string
    Currency         string
    Amount           int64
    IdempotencyKey   string
}

func (r *MySQLPaymentRepository) CreatePayment(ctx context.Context, p Payment) error {
    const query = `
        INSERT INTO payments (
            payment_id,
            order_id,
            user_id,
            provider,
            provider_intent_id,
            status,
            currency,
            amount,
            idempotency_key
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `

    _, err := r.db.ExecContext(
        ctx,
        query,
        p.PaymentID,
        p.OrderID,
        p.UserID,
        p.Provider,
        p.ProviderIntentID,
        p.Status,
        p.Currency,
        p.Amount,
        p.IdempotencyKey,
    )
    return err
}
```

### Repository rules

| Rule | Reason |
|---|---|
| `context.Context` first arg | Request timeout/cancel support |
| Parameterized SQL | SQL injection prevent |
| No card data | Compliance and security |
| Unique key errors mapped | Idempotency response clean rahe |
| Transactions for state changes | Payment/refund/webhook consistency |

---

## Step 15: Data Integrity Rules

| Rule | Where enforced |
|---|---|
| `payment_id` globally unique | `uk_payments_payment_id` |
| Same provider idempotency key duplicate nahi | `uk_payments_idempotency` |
| Same provider webhook event duplicate nahi | `uk_webhook_provider_event` |
| Same payment refund idempotency duplicate nahi | `uk_refunds_idempotency` |
| Refund belongs to valid payment | `fk_refunds_payment` |
| Attempt belongs to valid payment | `fk_payment_attempts_payment` |

### App-level checks

Kuch checks DB se zyada application layer me safe hote hain:

| Check | Why app-level |
|---|---|
| Amount positive hai | Better validation error return kar sakte hain |
| Currency allowed list me hai | Business config driven |
| Refund amount available balance se kam/equal hai | Needs current captured/refunded totals |
| State transition allowed hai | Task 1 state machine logic use hoga |
| Webhook signature valid hai | Provider-specific cryptographic check |

---

## Step 16: Payment Write Flow with DB Tables

```mermaid
flowchart LR
    A[Order Service: CreatePaymentIntent] --> B[Payment Service]
    B --> C[(payments)]
    B --> D[(payment_attempts)]
    B --> E[Provider Intent]
    E --> F[Frontend Provider UI]
    F --> G[Provider Webhook]
    G --> H[(payment_webhook_events)]
    H --> I[(payments status update)]
    I --> J[Publish PaymentCaptured or PaymentFailed]
```

### Explanation

1. Order Service payment intent request bhejta hai.
2. Payment Service `payments` me main record create karta hai.
3. Payment Service `payment_attempts` me attempt record create karta hai.
4. Provider intent create hota hai.
5. Provider webhook final status bhejta hai.
6. Webhook event pehle `payment_webhook_events` me idempotently store hota hai.
7. Same transaction me payment status update hota hai.
8. Future event publisher Order/Notification services ko update karega.

---

## Step 17: Verification Checklist

| Check | Expected |
|---|---|
| `payment_db` database exists | ✅ |
| `payments` table exists | ✅ |
| `payment_attempts` table exists | ✅ |
| `refunds` table exists | ✅ |
| `payment_webhook_events` table exists | ✅ |
| `payment_reconciliations` table exists | ✅ |
| Payment idempotency unique key exists | ✅ |
| Webhook provider event unique key exists | ✅ |
| Refund idempotency unique key exists | ✅ |
| Child tables reference `payments(payment_id)` | ✅ |

### Useful verification SQL

```sql
SHOW DATABASES LIKE 'payment_db';
USE payment_db;
SHOW TABLES;
SHOW INDEX FROM payments;
SHOW INDEX FROM payment_webhook_events;
SHOW INDEX FROM refunds;
```

### Idempotency test idea

```sql
-- First insert should pass.
INSERT INTO payments (
  payment_id,
  order_id,
  user_id,
  provider,
  status,
  currency,
  amount,
  idempotency_key
) VALUES (
  'pay_test_1',
  'ord_test_1',
  'usr_test_1',
  'stripe_like',
  'initiated',
  'INR',
  10000,
  'payment_intent:ord_test_1:attempt_1'
);

-- Second insert with same provider + idempotency_key should fail.
INSERT INTO payments (
  payment_id,
  order_id,
  user_id,
  provider,
  status,
  currency,
  amount,
  idempotency_key
) VALUES (
  'pay_test_2',
  'ord_test_1',
  'usr_test_1',
  'stripe_like',
  'initiated',
  'INR',
  10000,
  'payment_intent:ord_test_1:attempt_1'
);
```

Expected result: duplicate key error on `uk_payments_idempotency`.

---

## Step 18: Performance and Scalability Notes

| Area | Decision |
|---|---|
| Hot order lookup | `idx_payments_order` |
| Webhook lookup | `idx_payments_provider_payment` and `uk_webhook_provider_event` |
| Admin payment filters | `idx_payments_status_created` |
| Pending webhook retry | `idx_webhook_processed_received` |
| Refund list | `idx_refunds_payment_status` |
| Finance mismatch view | `idx_reconciliation_provider_status` |

### Future scaling options

- Read replicas for admin payment search.
- Batch reconciliation inserts.
- Partition large webhook/reconciliation tables by month if volume high ho.
- Archive old provider payload JSON to object storage if DB size grow kare.
- Outbox pattern add karna for reliable `payment.events` publishing.

---

## Step 19: Security and Compliance Notes

| Security rule | Schema impact |
|---|---|
| Card data platform par store nahi hoga | No card number/CVV columns |
| Provider secrets DB me store nahi honge | No API key/webhook secret columns |
| Logs and payloads sanitize honge | `payload` and `raw_provider_response` me sensitive data redact hoga |
| Amount server-side trusted hoga | `amount` Order Service validated total se aayega |
| Refund actor tracked hoga | `requested_by`, `reviewed_by`, `reviewed_at` fields |
| Webhook replay safe hoga | Unique `(provider, provider_event_id)` |

> 🔴 **Compliance reminder:** Agar future me provider payload store karna zaruri ho, to pehle redaction layer add karo. Raw card/cardholder-sensitive data kabhi persist nahi karna.

---

## Step 20: What This Unlocks

Task 2 complete hone ke baad future Payment Service tasks ke liye foundation ready hai:

| Future Task | Schema support |
|---|---|
| Task 3: Gateway abstraction | `provider`, provider IDs, provider response fields |
| Task 4: Create payment intent | `payments`, `payment_attempts`, idempotency key |
| Task 5: Webhook handler | `payment_webhook_events`, provider event unique key |
| Task 6: Refund flow | `refunds`, refund status, refund idempotency |
| Task 7: Retry handling | `payment_attempts`, payment idempotency |
| Task 8: Reconciliation job | `payment_reconciliations` |

---

## ✅ Completion Checklist

| Requirement | Status |
|---|---:|
| `TaskImplementation/` folder created/kept | ✅ |
| `TaskImplementation/Payment Service/` folder created/kept | ✅ |
| `task2.md` created | ✅ |
| Step-by-step Hinglish guide added | ✅ |
| MySQL schema explained | ✅ |
| External tools/libraries mentioned | ✅ |
| Folder structure included | ✅ |
| Code examples included | ✅ |
| Mermaid diagrams included | ✅ |
| Scope limited to Payment Service Task 2 | ✅ |

---

## 🏁 Final Summary

Payment Service Task 2 ke liye MySQL schema design complete hai:

- Main payment records ke liye `payments`
- Retry/attempt audit ke liye `payment_attempts`
- Full/partial refunds ke liye `refunds`
- Webhook idempotency ke liye `payment_webhook_events`
- Finance reconciliation ke liye `payment_reconciliations`

Is schema se Payment Service ke future provider integration, intent creation, webhook handling, refund flow, retry handling aur reconciliation tasks safely build ho sakte hain.
