# 🧩 CMS Service - Task 5: Offer Campaigns

![Task](https://img.shields.io/badge/Task-CMS%20Service%20%235-blue)
![Priority](https://img.shields.io/badge/Priority-P2-yellow)
![Dependency](https://img.shields.io/badge/Dependency-Coupon%20Engine-purple)
![Feature](https://img.shields.io/badge/Feature-Offer%20Campaigns-2ea44f)
![Database](https://img.shields.io/badge/Database-MySQL%208%2B-4479A1)
![Status](https://img.shields.io/badge/Status-Documented-success)
![Guide](https://img.shields.io/badge/Guide-Hinglish-orange)

## 📌 Task Summary

| Field | Detail |
|---|---|
| Task Name | Offer campaigns |
| Source | `docs/01-micro-tasks.md` -> `CMS Service` -> Task 5 |
| Goal | Start/end time, budget, usage limits, seller ownership implement karna |
| Dependency | `Coupon engine` |
| Priority | `P2` |
| Scope | Campaign domain design, campaign lifecycle, seller ownership, schedule window, budget/usage guards, coupon link strategy, API/gRPC references, DB query examples, audit expectations |
| Not Included | Actual Go backend code, SQL migration files, API spec edits, seller dashboard UI, campaign analytics dashboards, payment/order implementation |

> **Simple Hinglish goal:** CMS Service me seller offer campaigns ka design banana hai jisme seller apni campaign create/list/manage kar sake. Campaign ke andar start time, end time, budget cap, usage cap, status, seller ownership, aur coupon engine ke saath relation clearly define hona chahiye.

---

## ✅ Final Output Created

```text
TaskImplementation/
└── CMS Service/
    ├── task1.md
    ├── task2.md
    ├── task3.md
    ├── task4.md
    └── task5.md
```

### Why this structure?

- `TaskImplementation/` project ka existing task documentation folder hai.
- `CMS Service/` folder already present tha, isliye usko preserve kiya gaya.
- `task5.md` sirf **CMS Service - Task 5** ka offer campaign implementation guide document karta hai.
- Existing backend code, migrations, API contracts, ya frontend files modify nahi kiye gaye.

---

## 🧭 Requirement Source Mapping

| Document | Kya use kiya gaya |
|---|---|
| `docs/01-micro-tasks.md` | CMS Service Task 5 ka exact scope: `Offer campaigns` |
| `docs/03-folder-structure.md` | Future `backend/services/cms-service/internal/domain/campaign.go` and `usecase/manage_campaign.go` placement |
| `docs/04-microservice-design.md` | CMS responsibility: offer and campaign management |
| `docs/05-database-design.md` | CMS DB table list and `campaigns(seller_id, starts_at, ends_at)` index |
| `docs/09-cms-superadmin.md` | Seller CMS module: offers/coupons with campaign scheduling and usage limits |
| `database/draw.sql` | Existing `campaigns` table reference |
| `api/master-api.json` | `CMSService.ListCampaigns` and `CMSService.CreateCampaign` route references |
| `TaskImplementation/CMS Service/task1.md` | Campaign permissions: `cms:campaigns:read`, `cms:campaigns:create`, `cms:campaigns:update`, `cms:campaigns:disable` |
| `TaskImplementation/CMS Service/task2.md` | MySQL campaign schema and index strategy |
| `TaskImplementation/CMS Service/task4.md` | Coupon engine dependency and redemption usage limit pattern |

---

## 🚦 Scope Boundary

### ✅ In Scope

- Campaign lifecycle design.
- Seller ownership rule.
- Start and end time validation.
- Budget cap design in minor currency units.
- Usage limit design.
- Campaign status model.
- Coupon engine ke saath campaign relation.
- MySQL query examples.
- Future Go domain/usecase snippets.
- REST and gRPC contract reference.
- Audit log expectations.
- Mermaid architecture and flow diagrams.
- External tools/libraries explanation.

### 🚫 Out of Scope

- Actual `backend/services/cms-service/` code create karna.
- Actual migration file create karna.
- `api/master-api.json` update karna.
- Seller Dashboard campaign calendar UI banana.
- Campaign analytics aggregation implement karna.
- Order/Payment Service me discount settlement change karna.
- Task 6 seller analytics APIs implement karna.
- Task 7 full CMS gRPC implementation karna.

> 🔴 **Reason:** Task 5 ka requirement campaign behavior design aur implementation guide hai. Is file me only documentation create ki gayi hai, production code nahi.

---

## 🪜 Step-by-Step Implementation

## Step 1: Task requirement identify kiya

`docs/01-micro-tasks.md` me CMS Service ka Task 5 ye hai:

| S.No | Task Name | Detail | Dependency |
|---:|---|---|---|
| 5 | Offer campaigns | Start/end time, budget, usage limits, seller ownership implement karo | Coupon engine |

**Implementation decision:**

- Campaign ka source of truth **CMS Service** hoga.
- Seller sirf apne `seller_id` ke campaigns manage karega.
- Campaign start/end window UTC timestamps me store honge.
- Budget amount money minor units me store hoga, jaise INR paise.
- Usage limit coupon redemption/campaign usage records se enforce hoga.
- Campaign coupon engine ke upar build hoga, coupon engine ko replace nahi karega.

> 🟢 **Why:** Coupon engine discount calculate karta hai. Campaign us discount ko business window, budget, seller ownership, aur usage governance deta hai.

---

## Step 2: Campaign aur coupon ka relation clear kiya

Task 4 me coupon engine MVP design ho chuka hai. Task 5 me campaign layer add hoti hai.

### Coupon vs Campaign

| Concept | Responsibility | Example |
|---|---|---|
| Coupon | Discount rule calculate karna | `SAVE10`, 10% off, min cart INR 999 |
| Campaign | Offer window and budget control | Diwali Sale, Oct 20 to Oct 31, INR 50,000 budget |

### Relation rule

```text
Campaign controls when and how much promotion can run.
Coupon controls discount amount and eligibility.
```

### MVP relation options

| Option | Description | MVP Decision |
|---|---|---|
| Campaign owns coupon ids in `metadata` | Existing `campaigns.metadata` JSON me coupon ids store | ✅ Good for current schema |
| Add `campaign_id` to `coupons` table | Coupon directly campaign se linked | Future migration |
| Create `campaign_coupons` join table | Many campaigns to many coupons | Future advanced model |

**MVP decision:** Existing schema ke saath compatible rehne ke liye `campaigns.metadata` me coupon references store kar sakte hain.

```json
{
  "coupon_ids": ["coupon_123", "coupon_456"],
  "channels": ["seller_dashboard", "homepage_slot"],
  "description": "Festive seller offer"
}
```

> 🔵 **Note:** Future me agar campaign-coupon relation complex ho jaye, to `campaign_coupons` join table add karna better hoga. Task 5 me existing `campaigns` table enough hai.

---

## Step 3: Campaign ownership define kiya

Seller ownership Task 5 ka core requirement hai.

### Ownership rules

| Campaign Type | `seller_id` | Who can manage? |
|---|---|---|
| Seller campaign | Not null | Same seller only |
| Platform campaign | Null | Future superadmin only |

### Seller campaign rule

```text
auth.seller_id must equal campaign.seller_id
```

### Example

| Actor | Campaign seller | Result |
|---|---|---|
| `seller_10` | `seller_10` | ✅ Allowed |
| `seller_10` | `seller_20` | ❌ Denied |
| `seller_10` | `NULL` platform campaign | ❌ Denied for seller APIs |

### Permission mapping

| Action | Required permission |
|---|---|
| List campaigns | `cms:campaigns:read` |
| Create campaign | `cms:campaigns:create` |
| Update campaign | `cms:campaigns:update` |
| Pause/disable campaign | `cms:campaigns:disable` |

**Explanation:**  
Auth Service/Gateway identity verify karega, but CMS Service final seller ownership check karega. Isse cross-seller data leak avoid hota hai.

---

## Step 4: Campaign lifecycle design kiya

Existing DB reference me campaign statuses:

```sql
ENUM('draft', 'active', 'paused', 'completed')
```

### Status meaning

| Status | Meaning | Seller Action |
|---|---|---|
| `draft` | Campaign created but not running | Edit, activate |
| `active` | Time window me campaign live hai | Pause, view |
| `paused` | Temporarily stopped | Resume, edit limited fields |
| `completed` | Ended or budget/usage exhausted | Read-only |

### Status transition diagram

```mermaid
stateDiagram-v2
    [*] --> draft
    draft --> active: start window reached or activate
    active --> paused: seller pauses
    paused --> active: seller resumes
    active --> completed: end time reached
    active --> completed: budget exhausted
    active --> completed: usage limit reached
    paused --> completed: end time reached
    completed --> [*]
```

### Why built this way?

- `draft` seller ko campaign prepare karne ka safe stage deta hai.
- `active` campaign validation path me eligible hota hai.
- `paused` campaign ko delete kiye bina stop karne deta hai.
- `completed` immutable historical record rakhta hai.

---

## Step 5: Time window validation define kiya

Campaign me `starts_at` and `ends_at` required hain.

### Validation rules

| Field | Rule |
|---|---|
| `starts_at` | Valid UTC timestamp |
| `ends_at` | Valid UTC timestamp |
| `ends_at` | Must be greater than `starts_at` |
| Duration | Optional product rule, for example max 90 days |
| Timezone | Store in UTC |

### Active window rule

```text
campaign is active when:
status = active
starts_at <= now_utc
ends_at >= now_utc
```

### SQL example

```sql
SELECT campaign_id, seller_id, name, status, budget_amount,
       currency, starts_at, ends_at, metadata
FROM campaigns
WHERE seller_id = ?
  AND status = 'active'
  AND starts_at <= UTC_TIMESTAMP()
  AND ends_at >= UTC_TIMESTAMP()
ORDER BY starts_at ASC;
```

**Explanation:**  
Seller campaign list aur active offer validation dono me ye predictable query useful hai. `idx_campaigns_seller_window` index seller/time window lookup fast banata hai.

---

## Step 6: Budget model define kiya

Campaign budget ka matlab hai maximum total discount spend jo campaign ke through allow hoga.

### Money representation

Money floating point me store nahi karna.

```text
INR 500.00 => 50000 paise
INR 1,499.99 => 149999 paise
```

### Existing DB fields

```sql
budget_amount BIGINT NULL,
currency CHAR(3) NOT NULL DEFAULT 'INR'
```

### Budget rules

| Rule | Explanation |
|---|---|
| `budget_amount` null | Unlimited budget for MVP, but use carefully |
| `budget_amount > 0` | Campaign spend cap |
| Currency required | Coupon/cart/order currency mismatch avoid karne ke liye |
| Spend counted after redemption | Preview spend reserve nahi karega |
| Final write transactional | Budget overrun race condition avoid karne ke liye |

### Budget check formula

```text
remaining_budget = campaign.budget_amount - redeemed_discount_sum

if discount_amount > remaining_budget:
    campaign invalid = budget_exhausted
```

### Example

| Budget | Already Redeemed | New Discount | Result |
|---:|---:|---:|---|
| INR 50,000 | INR 49,000 | INR 500 | ✅ Allowed |
| INR 50,000 | INR 49,800 | INR 500 | ❌ Budget exhausted |

> 🟡 **Important:** Cart preview me budget estimate dikh sakta hai, but final budget consume only Order Service payment success ke baad redemption record karte time hoga.

---

## Step 7: Usage limits define kiye

Campaign usage limit ka matlab hai campaign ke under total successful redemptions ka cap.

### MVP usage limit storage

Current `campaigns` table me direct `usage_limit` column nahi hai. Task 5 ke requirement ke liye two compatible approaches:

| Approach | Description | Decision |
|---|---|---|
| Store in `metadata.usage_limit` | Existing schema ke saath compatible | ✅ MVP |
| Add `usage_limit INT NULL` column | Cleaner query, needs migration | Future improvement |

### MVP metadata example

```json
{
  "coupon_ids": ["coupon_123"],
  "usage_limit": 10000,
  "per_user_limit": 1,
  "budget_alert_threshold_percent": 80
}
```

### Usage limit rules

| Limit Type | Scope | Source |
|---|---|---|
| Global campaign usage | Total successful campaign redemptions | Campaign usage/redemption aggregate |
| Per-user campaign usage | Same user per campaign | Redemption records |
| Coupon usage | Coupon-specific limit | Coupon engine Task 4 |

### Validation order

```text
1. Campaign exists
2. Seller ownership valid
3. Status active
4. Current time inside start/end window
5. Usage limit available
6. Budget available
7. Coupon engine validates coupon
```

**Explanation:**  
Cheap checks pehle karna better hai. Coupon engine calculation tabhi chale jab campaign basic eligibility pass ho.

---

## Step 8: Database design align kiya

Existing `database/draw.sql` me campaign table reference:

```sql
CREATE TABLE IF NOT EXISTS campaigns (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  campaign_id VARCHAR(64) NOT NULL,
  seller_id VARCHAR(64) NULL,
  name VARCHAR(255) NOT NULL,
  status ENUM('draft', 'active', 'paused', 'completed') NOT NULL DEFAULT 'draft',
  budget_amount BIGINT NULL,
  currency CHAR(3) NOT NULL DEFAULT 'INR',
  starts_at TIMESTAMP NOT NULL,
  ends_at TIMESTAMP NOT NULL,
  metadata JSON NULL,
  created_by VARCHAR(64) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_campaigns_campaign_id (campaign_id),
  KEY idx_campaigns_seller_window (seller_id, starts_at, ends_at),
  KEY idx_campaigns_status (status)
) ENGINE=InnoDB;
```

### Field explanation

| Field | Purpose |
|---|---|
| `campaign_id` | Public stable campaign id |
| `seller_id` | Seller ownership, nullable for future platform campaigns |
| `name` | Seller-facing campaign name |
| `status` | Lifecycle state |
| `budget_amount` | Max discount spend in minor units |
| `currency` | Budget currency |
| `starts_at` | Campaign start UTC |
| `ends_at` | Campaign end UTC |
| `metadata` | MVP flexible config like coupon ids, usage limit |
| `created_by` | Actor user id |

### Why MySQL?

- Campaigns structured data hain.
- Seller ownership and status filtering common hai.
- Budget and usage checks transactional safety need karte hain.
- Audit logs and campaign changes relationally traceable hain.

---

## Step 9: Campaign create flow design kiya

Seller dashboard `POST /api/v1/seller/campaigns` call karega.

### API reference

| Route | gRPC | Auth | Purpose |
|---|---|---|---|
| `POST /api/v1/seller/campaigns` | `CMSService.CreateCampaign` | Seller | New campaign create |
| `GET /api/v1/seller/campaigns` | `CMSService.ListCampaigns` | Seller | Seller campaigns list |

### Request example

```json
{
  "name": "Festive Weekend Sale",
  "starts_at": "2026-10-20T00:00:00Z",
  "ends_at": "2026-10-31T23:59:59Z",
  "budget": {
    "amount": 5000000,
    "currency": "INR"
  },
  "metadata": {
    "coupon_ids": ["coupon_123"],
    "usage_limit": 10000
  }
}
```

### Response example

```json
{
  "campaign_id": "camp_01JABCDEF123",
  "seller_id": "seller_456",
  "name": "Festive Weekend Sale",
  "status": "draft",
  "budget": {
    "amount": 5000000,
    "currency": "INR"
  },
  "starts_at": "2026-10-20T00:00:00Z",
  "ends_at": "2026-10-31T23:59:59Z"
}
```

### Create flow

```mermaid
sequenceDiagram
    participant Seller as Seller Dashboard
    participant GW as API Gateway
    participant CMS as CMS Service
    participant DB as MySQL cms_db

    Seller->>GW: POST /api/v1/seller/campaigns
    GW->>GW: Validate JWT + seller role
    GW->>CMS: CreateCampaign(auth_context + input)
    CMS->>CMS: Check cms:campaigns:create
    CMS->>CMS: Validate seller ownership + time + budget
    CMS->>DB: INSERT campaigns
    CMS->>DB: INSERT cms_audit_logs
    DB-->>CMS: Created
    CMS-->>GW: Campaign response
    GW-->>Seller: 201 Created
```

---

## Step 10: Validation rules define kiye

### Create campaign validations

| Field | Rule | Error Code |
|---|---|---|
| `name` | Required, trimmed, max 255 chars | `invalid_name` |
| `seller_id` | From auth context, not client body | `invalid_seller` |
| `starts_at` | Required valid UTC time | `invalid_start_time` |
| `ends_at` | Required and greater than `starts_at` | `invalid_end_time` |
| `budget.amount` | Optional, positive if provided | `invalid_budget` |
| `budget.currency` | Required when budget present | `invalid_currency` |
| `metadata.coupon_ids` | Optional array of coupon ids owned by seller | `invalid_coupon_scope` |
| `metadata.usage_limit` | Optional positive int | `invalid_usage_limit` |

### Campaign active validation

| Check | Invalid Reason |
|---|---|
| Campaign not found | `campaign_not_found` |
| Seller mismatch | `seller_scope_mismatch` |
| Status not active | `campaign_not_active` |
| Starts in future | `campaign_not_started` |
| Ends in past | `campaign_expired` |
| Usage cap reached | `campaign_usage_limit_reached` |
| Budget spent | `campaign_budget_exhausted` |

**Explanation:**  
Machine-readable error codes frontend and logs ke liye stable hote hain. Human-readable messages frontend translate kar sakta hai.

---

## Step 11: Campaign budget and usage check flow banaya

Campaign budget and usage final redemption ke time enforce honge.

```mermaid
flowchart TD
    A[Order payment success] --> B[Order Service calls CMS redemption]
    B --> C[Load campaign and coupon]
    C --> D{Campaign active window valid?}
    D -- No --> X[Reject redemption]
    D -- Yes --> E{Usage limit available?}
    E -- No --> Y[Reject: usage limit reached]
    E -- Yes --> F{Budget available?}
    F -- No --> Z[Reject: budget exhausted]
    F -- Yes --> G[Record coupon redemption]
    G --> H[Update campaign spend/read model if present]
    H --> I[Audit and return success]
```

### Race condition handling

High traffic me multiple orders same campaign budget consume kar sakte hain. Final implementation me transaction or locking required hoga.

```sql
START TRANSACTION;

SELECT campaign_id, budget_amount, metadata
FROM campaigns
WHERE campaign_id = ?
  AND seller_id = ?
FOR UPDATE;

-- Then calculate redeemed amount and usage count.
-- If still available, insert redemption record.

COMMIT;
```

> 🔵 **Why `FOR UPDATE`:** Same campaign row par concurrent redemptions ko serialize karke budget overspend risk reduce hota hai.

---

## Step 12: Campaign list flow design kiya

Seller dashboard ko campaign calendar/list view chahiye hoga.

### List request

```http
GET /api/v1/seller/campaigns?page=1&page_size=20
```

### Optional filters for future

| Filter | Purpose |
|---|---|
| `status` | Draft/active/paused/completed campaigns |
| `starts_after` | Calendar date range |
| `ends_before` | Calendar date range |
| `q` | Name search |

### SQL example

```sql
SELECT campaign_id, seller_id, name, status, budget_amount,
       currency, starts_at, ends_at, metadata, created_at, updated_at
FROM campaigns
WHERE seller_id = ?
ORDER BY starts_at DESC
LIMIT ? OFFSET ?;
```

### Seller ownership safety

```text
WHERE seller_id = auth.seller_id
```

**Explanation:**  
Client se `seller_id` trust nahi karna. Seller id JWT/auth context se aayega.

---

## Step 13: Campaign update and pause behavior define kiya

Task 5 me API reference only create/list tak documented hai, but permissions me update/disable already defined hain. Future implementation ke liye behavior clear hona chahiye.

### Editable fields by status

| Field | Draft | Active | Paused | Completed |
|---|---|---|---|---|
| `name` | ✅ | ✅ | ✅ | ❌ |
| `starts_at` | ✅ | ⚠️ limited | ✅ | ❌ |
| `ends_at` | ✅ | ✅ extend only | ✅ | ❌ |
| `budget_amount` | ✅ | ✅ increase only | ✅ | ❌ |
| `metadata.coupon_ids` | ✅ | ⚠️ limited | ✅ | ❌ |
| `status` | ✅ | ✅ | ✅ | ❌ |

### Why active campaign edits limited?

Active campaign ke rules suddenly reduce karna buyer experience break kar sakta hai. Budget increase safe hai, but budget decrease current in-flight checkout ko affect kar sakta hai.

### Pause rule

```text
paused campaign coupon preview and redemption dono me invalid hoga.
```

---

## Step 14: Audit logs design kiye

Campaign changes business-sensitive hain, isliye audit logs required hain.

### Audit actions

| Action | Resource Type |
|---|---|
| `cms:campaigns:create` | `campaign` |
| `cms:campaigns:update` | `campaign` |
| `cms:campaigns:disable` | `campaign` |
| `cms:campaigns:resume` | `campaign` |
| `cms:campaigns:complete` | `campaign` |

### Audit log example

```json
{
  "audit_id": "audit_01JXYZ",
  "seller_id": "seller_456",
  "actor_user_id": "user_123",
  "action": "cms:campaigns:create",
  "resource_type": "campaign",
  "resource_id": "camp_01JABCDEF123",
  "before_json": null,
  "after_json": {
    "name": "Festive Weekend Sale",
    "status": "draft",
    "budget_amount": 5000000
  },
  "request_id": "req_789",
  "created_at": "2026-10-01T10:00:00Z"
}
```

**Explanation:**  
Audit logs seller dispute, finance review, and debugging ke liye helpful honge.

---

## Step 15: Architecture diagram banaya

```mermaid
flowchart LR
    Seller[Seller Dashboard] --> GW[API Gateway]
    GW --> CMS[CMS Service]
    CMS --> DB[(MySQL cms_db)]
    CMS --> Coupon[Coupon Engine]
    Cart[Cart Service] -->|Preview coupon| CMS
    Order[Order Service] -->|Payment success redemption| CMS

    subgraph CMSDB[CMS DB Tables]
      C1[campaigns]
      C2[coupons]
      C3[coupon_rules]
      C4[coupon_redemptions]
      C5[cms_audit_logs]
    end

    DB --> CMSDB
```

### Flow summary

| Flow | Service Path |
|---|---|
| Campaign create/list | Seller Dashboard -> Gateway -> CMS -> MySQL |
| Coupon preview | User App -> Gateway -> Cart -> CMS |
| Final redemption | Payment success -> Order -> CMS -> MySQL |
| Audit | CMS -> `cms_audit_logs` |

---

## Step 16: Future folder structure define kiya

Task 5 me actual files create nahi kiye gaye, but implementation ka clean placement ye hoga:

```text
backend/
└── services/
    └── cms-service/
        ├── cmd/
        │   └── server/
        │       └── main.go
        ├── internal/
        │   ├── domain/
        │   │   ├── coupon.go
        │   │   ├── campaign.go
        │   │   └── seller_setting.go
        │   ├── usecase/
        │   │   ├── validate_coupon.go
        │   │   ├── manage_campaign.go
        │   │   └── record_redemption.go
        │   ├── repository/
        │   │   └── mysql_cms_repository.go
        │   └── transport/
        │       └── grpc/
        │           └── campaign_handler.go
        ├── migrations/
        │   ├── 001_create_cms_tables.up.sql
        │   └── 001_create_cms_tables.down.sql
        └── deploy/
```

### Responsibility table

| File | Responsibility |
|---|---|
| `domain/campaign.go` | Campaign entity, status constants, validation helpers |
| `usecase/manage_campaign.go` | Create/list/update/pause campaign business flow |
| `usecase/record_redemption.go` | Final budget and usage guard before redemption |
| `repository/mysql_cms_repository.go` | Campaign queries and transactions |
| `transport/grpc/campaign_handler.go` | gRPC request/response mapping |
| `migrations/` | Campaign table and future usage/spend columns |

---

## Step 17: Domain model code example

> Ye code actual repo me add nahi kiya gaya. Ye future implementation reference hai.

```go
package domain

import "time"

type CampaignStatus string

const (
	CampaignStatusDraft     CampaignStatus = "draft"
	CampaignStatusActive    CampaignStatus = "active"
	CampaignStatusPaused    CampaignStatus = "paused"
	CampaignStatusCompleted CampaignStatus = "completed"
)

type Campaign struct {
	CampaignID  string
	SellerID    *string
	Name        string
	Status      CampaignStatus
	BudgetMinor *int64
	Currency    string
	StartsAt    time.Time
	EndsAt      time.Time
	Metadata    CampaignMetadata
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CampaignMetadata struct {
	CouponIDs     []string `json:"coupon_ids,omitempty"`
	UsageLimit   *int64   `json:"usage_limit,omitempty"`
	PerUserLimit *int64   `json:"per_user_limit,omitempty"`
}

func (c Campaign) IsInsideWindow(now time.Time) bool {
	return !now.Before(c.StartsAt) && !now.After(c.EndsAt)
}

func (c Campaign) IsSellerOwnedBy(sellerID string) bool {
	return c.SellerID != nil && *c.SellerID == sellerID
}
```

### Explanation

- `CampaignStatus` invalid string status avoid karta hai.
- `BudgetMinor` pointer hai kyunki budget optional ho sakta hai.
- `SellerID` pointer hai kyunki future platform campaign me null allowed ho sakta hai.
- `IsInsideWindow` start/end date logic centralize karta hai.
- `IsSellerOwnedBy` cross-seller update/read ko block karne me help karta hai.

---

## Step 18: Usecase code example

> Ye pseudo-real Go example hai. Is task me file create nahi hui.

```go
package usecase

import (
	"context"
	"errors"
	"time"
)

type CreateCampaignInput struct {
	ActorUserID string
	SellerID    string
	Name        string
	BudgetMinor *int64
	Currency    string
	StartsAt    time.Time
	EndsAt      time.Time
	CouponIDs   []string
	UsageLimit  *int64
	RequestID   string
}

func (u *CampaignUsecase) CreateCampaign(ctx context.Context, in CreateCampaignInput) (*Campaign, error) {
	if in.SellerID == "" {
		return nil, errors.New("invalid_seller")
	}
	if in.Name == "" {
		return nil, errors.New("invalid_name")
	}
	if !in.EndsAt.After(in.StartsAt) {
		return nil, errors.New("invalid_time_window")
	}
	if in.BudgetMinor != nil && *in.BudgetMinor <= 0 {
		return nil, errors.New("invalid_budget")
	}

	campaign := NewDraftCampaign(in)

	if err := u.repo.CreateCampaign(ctx, campaign); err != nil {
		return nil, err
	}

	_ = u.audit.Log(ctx, AuditEntry{
		SellerID:     in.SellerID,
		ActorUserID:  in.ActorUserID,
		Action:       "cms:campaigns:create",
		ResourceType: "campaign",
		ResourceID:   campaign.CampaignID,
		RequestID:    in.RequestID,
	})

	return campaign, nil
}
```

### Explanation

- Auth context se `SellerID` aata hai, request body se nahi.
- Time window and budget validation usecase layer me hoti hai.
- Repository sirf persistence details handle karta hai.
- Audit best-effort nahi hona chahiye production me. Transactional outbox future me better rahega.

---

## Step 19: Repository query examples

### Insert campaign

```sql
INSERT INTO campaigns (
  campaign_id,
  seller_id,
  name,
  status,
  budget_amount,
  currency,
  starts_at,
  ends_at,
  metadata,
  created_by
) VALUES (?, ?, ?, 'draft', ?, ?, ?, ?, CAST(? AS JSON), ?);
```

### List seller campaigns

```sql
SELECT campaign_id, seller_id, name, status, budget_amount,
       currency, starts_at, ends_at, metadata, created_at, updated_at
FROM campaigns
WHERE seller_id = ?
ORDER BY starts_at DESC
LIMIT ? OFFSET ?;
```

### Active campaign lookup

```sql
SELECT campaign_id, seller_id, name, status, budget_amount,
       currency, starts_at, ends_at, metadata
FROM campaigns
WHERE seller_id = ?
  AND status = 'active'
  AND starts_at <= UTC_TIMESTAMP()
  AND ends_at >= UTC_TIMESTAMP();
```

### Pause campaign

```sql
UPDATE campaigns
SET status = 'paused'
WHERE campaign_id = ?
  AND seller_id = ?
  AND status IN ('draft', 'active');
```

---

## Step 20: Campaign validation response design kiya

Internal services ko stable response chahiye hoga.

```json
{
  "valid": true,
  "campaign_id": "camp_01JABCDEF123",
  "coupon_id": "coupon_123",
  "remaining_budget": {
    "amount": 2500000,
    "currency": "INR"
  },
  "reason": null
}
```

### Invalid response

```json
{
  "valid": false,
  "campaign_id": "camp_01JABCDEF123",
  "coupon_id": "coupon_123",
  "remaining_budget": {
    "amount": 0,
    "currency": "INR"
  },
  "reason": "campaign_budget_exhausted"
}
```

### Reason codes

| Reason | Meaning |
|---|---|
| `campaign_not_found` | Campaign id invalid |
| `seller_scope_mismatch` | Campaign seller and request seller mismatch |
| `campaign_not_active` | Status active nahi hai |
| `campaign_not_started` | Start time future me hai |
| `campaign_expired` | End time past me hai |
| `campaign_usage_limit_reached` | Usage limit complete ho chuka |
| `campaign_budget_exhausted` | Budget remaining nahi hai |
| `coupon_not_in_campaign` | Coupon id campaign metadata me allowed nahi |

---

## Step 21: Campaign and coupon validation combined flow

```mermaid
flowchart TD
    A[Cart or Order requests coupon validation] --> B[CMS loads coupon]
    B --> C[Coupon basic checks]
    C --> D{Coupon linked to campaign?}
    D -- No --> H[Run normal coupon engine]
    D -- Yes --> E[Load campaign]
    E --> F{Campaign eligible?}
    F -- No --> X[Return invalid campaign reason]
    F -- Yes --> G{Budget and usage available?}
    G -- No --> Y[Return budget or usage reason]
    G -- Yes --> H[Run coupon discount calculation]
    H --> I[Return valid discount preview]
```

### Why campaign check before coupon calculation?

- Campaign status/time/budget checks cheap hain.
- Discount calculation cart items, scope, and rules inspect karta hai.
- Invalid campaign pe compute waste avoid hota hai.

---

## Step 22: External libraries/tools mention kiye

Is documentation-only task me koi external package install nahi kiya gaya. Real CMS Service Task 5 implement karte time ye tools useful honge:

| Tool/Library | What it is | Why used | Install |
|---|---|---|---|
| MySQL 8+ | Relational database | Campaigns, coupons, audit logs structured and transactional hain | Docker ya managed MySQL |
| `github.com/go-sql-driver/mysql` | Go MySQL driver | Go service ko MySQL se connect karne ke liye | `go get github.com/go-sql-driver/mysql` |
| `github.com/golang-migrate/migrate/v4` | DB migration tool | Versioned schema changes run karne ke liye | `go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest` |
| `google.golang.org/grpc` | gRPC framework | Gateway/Internal services se CMS calls typed banane ke liye | `go get google.golang.org/grpc` |
| `google.golang.org/protobuf` | Proto runtime | Generated proto messages support ke liye | `go get google.golang.org/protobuf` |
| Mermaid | Markdown diagram syntax | Architecture and flow diagrams readable banane ke liye | GitHub/Markdown viewer me built-in support often available |
| Shields.io badges | Markdown badge images | Document ko visually scannable banane ke liye | No install required |

### Local MySQL example

```bash
docker run --name ecommerce-cms-mysql \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=cms_db \
  -e MYSQL_USER=cms_user \
  -e MYSQL_PASSWORD=cms_password \
  -p 3306:3306 \
  -d mysql:8
```

### Migration example

```bash
migrate -path backend/services/cms-service/migrations \
  -database "mysql://cms_user:cms_password@tcp(localhost:3306)/cms_db" \
  up
```

> 🟡 **Note:** Commands reference ke liye hain. Current task me koi package install ya service run nahi kiya gaya.

---

## Step 23: Testing strategy define ki

### Unit tests

| Test Case | Expected Result |
|---|---|
| End time before start time | Validation error |
| Empty campaign name | Validation error |
| Negative budget | Validation error |
| Seller mismatch | Access denied |
| Active campaign inside window | Eligible |
| Active campaign before start | Not started |
| Active campaign after end | Expired |
| Budget remaining enough | Allow |
| Budget exhausted | Reject |
| Usage limit reached | Reject |

### Repository tests

| Test Case | Expected Result |
|---|---|
| Insert campaign | Row created |
| List campaigns by seller | Only own seller rows returned |
| Active campaign lookup | Uses status and time window |
| Pause campaign with seller id | Only own campaign updated |

### Integration tests

| Flow | Expected Result |
|---|---|
| Create campaign then list | Campaign visible in seller list |
| Cross-seller list | Other seller campaign hidden |
| Coupon linked to inactive campaign | Validation returns invalid |
| Redemption when budget exhausted | Redemption rejected |

---

## Step 24: Observability define ki

### Logs

Campaign logs me ye fields hone chahiye:

| Field | Purpose |
|---|---|
| `request_id` | Trace one request |
| `seller_id` | Seller scope |
| `actor_user_id` | Who performed action |
| `campaign_id` | Resource id |
| `action` | create/update/pause/list |
| `status` | Campaign status |
| `reason` | Failure reason |

### Metrics

| Metric | Type | Purpose |
|---|---|---|
| `cms_campaign_create_total` | Counter | Campaign creation volume |
| `cms_campaign_active_total` | Gauge | Active campaigns count |
| `cms_campaign_validation_total` | Counter | Validation checks |
| `cms_campaign_validation_failed_total` | Counter | Invalid reason trend |
| `cms_campaign_budget_exhausted_total` | Counter | Budget cap hit count |
| `cms_campaign_usage_exhausted_total` | Counter | Usage cap hit count |

### Tracing

```text
Gateway request span
  -> CMS CreateCampaign span
    -> MySQL insert campaign span
    -> MySQL insert audit log span
```

---

## Step 25: Security and edge cases cover kiye

| Risk | Protection |
|---|---|
| Seller tries another seller campaign id | Always filter by `seller_id = auth.seller_id` |
| Client sends fake seller id | Ignore body seller id, use auth context |
| Campaign active outside window | Check UTC `starts_at` and `ends_at` |
| Budget overspend under concurrency | Use transaction and row lock on final redemption |
| Usage limit race | Count/aggregate check inside transaction |
| Coupon from other seller linked | Validate coupon ownership before link |
| Platform campaign misuse | Seller APIs reject `seller_id IS NULL` campaigns |
| Completed campaign edited | Make completed read-only |
| Currency mismatch | Compare campaign/coupon/cart currency |

---

## ✅ Acceptance Checklist

| Requirement | Status |
|---|---|
| `TaskImplementation/` folder exists | ✅ |
| `TaskImplementation/CMS Service/` folder kept | ✅ |
| `task5.md` created | ✅ |
| Step-by-step implementation in Hinglish | ✅ |
| Clear explanation of each part | ✅ |
| External tools/libraries listed with why/install/use | ✅ |
| Clean folder structure included | ✅ |
| Code examples included | ✅ |
| Mermaid diagrams included | ✅ |
| Scope limited to CMS Service Task 5 | ✅ |
| No backend code or migrations implemented | ✅ |

---

## 🧾 Final Notes

CMS Service Task 5 ke liye Offer Campaigns ka final design:

| Area | Decision |
|---|---|
| Source of truth | CMS Service |
| Database | MySQL `cms_db` |
| Main table | `campaigns` |
| Ownership | `campaigns.seller_id` + auth seller scope |
| Time control | `starts_at`, `ends_at`, UTC |
| Budget control | `budget_amount` in minor units |
| Usage control | MVP via `metadata.usage_limit`, future direct column |
| Coupon dependency | Campaign wraps coupon engine eligibility |
| Audit | `cms_audit_logs` for campaign mutations |
| Public seller routes | `GET /api/v1/seller/campaigns`, `POST /api/v1/seller/campaigns` |

> 🟢 **Task 5 complete:** Offer campaigns ke liye start/end time, budget, usage limits, seller ownership, lifecycle, coupon engine integration, database strategy, API references, diagrams, tools, tests, and security considerations documented hain. No implementation beyond CMS Service Task 5 kiya gaya.
