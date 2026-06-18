# 🛡️ Superadmin Service - Task 7: Platform Settings

![Task](https://img.shields.io/badge/Task-Superadmin%20Service%20%237-blue)
![Priority](https://img.shields.io/badge/Priority-P2-lightgrey)
![Dependency](https://img.shields.io/badge/Dependency-CMS%20%2B%20Search-yellow)
![Status](https://img.shields.io/badge/Status-Documented-success)
![Security](https://img.shields.io/badge/Security-High%20Risk%20Settings-critical)
![Scope](https://img.shields.io/badge/Scope-Platform%20Control%20Plane-purple)
![Guide](https://img.shields.io/badge/Guide-Hinglish-ff69b4)

## 📌 Task Summary

| Field | Detail |
|---|---|
| Task Name | Platform settings |
| Source | `docs/01-micro-tasks.md` -> `Superadmin Service` -> Task 7 |
| Goal | Search synonyms, commission, feature flags, maintenance mode settings banana |
| Dependency | CMS Service + Search Service |
| Priority | P2 |
| Output Type | Documentation-only implementation guide |
| Not Included | Frontend platform settings UI, final immutable audit-log service, full search reindex job, CMS coupon engine, deployment config |

> **Simple Hinglish goal:** Is task ka kaam Superadmin Service ke andar platform-wide settings ka safe control model define karna hai. Superadmin commission rules, feature flags, maintenance mode, aur search synonyms ko controlled way me manage karega. Sensitive settings change karte time permission check, reason, validation, versioning, cache invalidation, downstream sync, aur audit-ready record mandatory hoga.

---

## ✅ Final Output Created

```text
TaskImplementation/
└── Superadmin Service/
    ├── task1.md
    ├── task2.md
    ├── task3.md
    ├── task4.md
    ├── task5.md
    ├── task6.md
    └── task7.md
```

### Why this structure?

| Folder/File | Purpose |
|---|---|
| `TaskImplementation/` | Saare task-wise implementation guides ka central location |
| `Superadmin Service/` | Superadmin Service ke tasks ko logically group karta hai |
| `task1.md` | Admin domain boundaries define karta hai |
| `task2.md` | Superadmin Service ke liye MySQL decision explain karta hai |
| `task3.md` | Admin RBAC roles and permissions define karta hai |
| `task4.md` | User/seller controls guide |
| `task5.md` | Order/payment controls guide |
| `task6.md` | Session visibility access-control guide |
| `task7.md` | Sirf **Superadmin Service - Task 7** ka platform settings guide |

> 🟢 **Important:** `Superadmin Service` folder already present tha, isliye usko keep kiya gaya. Is task me backend source code, DB migration, proto file, ya frontend page create nahi kiye gaye. Ye beginner-friendly implementation guide hai.

---

## 📚 Documents Studied

| Document | Kya use kiya gaya |
|---|---|
| `docs/01-micro-tasks.md` | Task 7 ka exact scope: search synonyms, commission, feature flags, maintenance mode |
| `docs/04-microservice-design.md` | Superadmin, CMS, Search, API Gateway responsibilities |
| `docs/05-database-design.md` | `platform_settings(setting_key)` unique index and settings cache direction |
| `docs/06-auth-security.md` | `platform:settings:write` sirf superadmin ke liye high-risk permission |
| `docs/09-cms-superadmin.md` | Superadmin modules: Search, Platform Settings, high-risk controls |
| `docs/10-frontend-implementation.md` | Future Superadmin Panel settings page location |
| `docs/12-logging-monitoring-scalability.md` | Platform settings Redis cache TTL 5 min and invalidation rule |
| `api/master-api.json` | `GET /api/v1/admin/settings`, `PATCH /api/v1/admin/settings/{key}`, schema names |
| `TaskImplementation/Superadmin Service/task1.md` | Domain rule: platform changes critical and audit mandatory |
| `TaskImplementation/Superadmin Service/task2.md` | MySQL chosen for platform settings and audit-sensitive data |
| `TaskImplementation/Superadmin Service/task3.md` | Permissions: `platform:settings:*`, `search:synonyms:*` |
| `TaskImplementation/Superadmin Service/task6.md` | Existing style for documenting Superadmin-side access-control boundaries |
| `backend/services/superadmin-service/internal/domain/permission.go` | Current permission constants already include settings and search synonym permissions |
| `backend/services/superadmin-service/internal/rbac/matrix.go` | Current role-permission matrix already grants settings only to `superadmin` |
| `backend/services/superadmin-service/internal/usecase/audit_recorder.go` | Existing audit-recorder style can log mutation context until Task 8 finalizes immutable logs |

---

## 🧭 Implementation Approach

Task 7 ko **platform control-plane settings layer** treat kiya gaya. Iska matlab Superadmin Service authoritative admin decision aur setting persistence own karegi, lekin downstream services apne domain ka actual behavior apply karenge.

### Final decision

```text
General platform setting source of truth: Superadmin MySQL platform_settings
Search synonym apply owner: Search Service / Typesense
Commission consumer: CMS / Order / Payment workflows, depending on calculation stage
Maintenance mode consumer: API Gateway + frontend apps
Feature flag consumer: API Gateway, frontend apps, and selected backend services
Exact permission owner: Superadmin Service RBAC
Cache strategy: Redis/local cache with 5 min TTL + invalidation event
Audit direction: audit-ready mutation record now, immutable audit-log finalization in Task 8
```

### One-line reason

Platform settings chhoti config lag sakti hain, but impact bada hota hai. Commission galat hua to revenue affect hoga, maintenance mode galat hua to site down feel hogi, feature flag galat hua to broken feature public ho sakta hai, aur search synonyms galat hue to product discovery hurt hogi.

---

## 🪜 Step-by-Step Implementation

## Step 1: Task Boundary Clear Kiya

`docs/01-micro-tasks.md` me Superadmin Service Task 7 ye hai:

| S.No | Task Name | Detail | Dependency | Priority |
|---:|---|---|---|---|
| 7 | Platform settings | Search synonyms, commission, feature flags, maintenance mode settings banao. | CMS/Search | P2 |

### Is task me allowed work

- Platform setting categories define karna
- Setting JSON contracts define karna
- `platform_settings` table ka implementation blueprint dena
- Settings read/update APIs ka flow explain karna
- RBAC permission mapping define karna
- Validation rules define karna
- Search/CMS/Gateway sync flow define karna
- Cache invalidation strategy define karna
- Code examples dena for domain, repository, usecase, handler, validation, and tests

### Is task me not allowed work

| Area | Reason |
|---|---|
| Superadmin frontend settings page banana | Ye **Superadmin Panel - Task 7** ka scope hai |
| Full Typesense reindex job banana | Ye **Search Service - Task 8** ka scope hai |
| Coupon/commission settlement engine banana | Ye CMS/Order/Payment implementation ka scope hai |
| Final immutable audit-log query/export system banana | Ye **Superadmin Service - Task 8** ka scope hai |
| User/seller/order/payment controls change karna | Wo Tasks 4 and 5 me covered hain |
| Session privacy controls UI banana | Wo Session Analytics Dashboard/Superadmin Panel tasks ka scope hai |

> 🔴 **Boundary rule:** Task 7 me Superadmin Service settings ko validate, persist, authorize, aur publish karegi. Wo Search Service ke Typesense index ko direct mutate nahi karegi, aur CMS/Order/Payment ke internal DBs me direct update nahi karegi.

---

## Step 2: Platform Settings Categories Define Ki

Task 7 ke settings ko four main categories me split kiya gaya.

| Setting Key | Purpose | Risk | Main Permission | Consumer |
|---|---|---|---|---|
| `maintenance_mode` | Site ko maintenance mode me le jana | Critical | `platform:settings:write` | API Gateway, frontend apps |
| `commission_rules` | Platform commission rates control karna | Critical | `platform:settings:write` | CMS, Order, Payment/settlement workflows |
| `feature_flags` | Features ko enable/disable/rollout karna | High/Critical | `platform:settings:write` | Gateway, frontend apps, backend services |
| `search_synonyms` | Search synonyms config manage karna | Medium | `search:synonyms:write` | Search Service / Typesense |

### Why split by category?

| Reason | Explanation |
|---|---|
| Validation alag hai | Commission numeric hai, synonyms text list hai, feature flags rollout config hai |
| Permissions alag ho sakti hain | Catalog admin search synonyms update kar sakta hai, but commission nahi |
| Downstream consumer alag hai | Maintenance Gateway consume karega, synonyms Search Service consume karega |
| Audit risk alag hai | Commission and maintenance critical, synonyms medium risk |

---

## Step 3: Ownership Rules Final Kiye

### Service ownership table

| Data / Behavior | Source of Truth | Superadmin Service ka role |
|---|---|---|
| Current platform settings | Superadmin MySQL | Persist and version settings |
| Admin permission decision | Superadmin RBAC | Exact permission enforce karega |
| Search synonyms applied to index | Search Service / Typesense | Validated admin change ko downstream apply karwayega |
| Commission usage in seller/coupon flows | CMS / Order / Payment workflows | Commission config provide karega |
| Maintenance mode enforcement | API Gateway | Gateway ko current setting/cache invalidation dega |
| Feature flag behavior | Consumer services/frontend | Flag config expose/sync karega |
| Immutable audit history | Superadmin Service Task 8 | Task 7 audit-ready record prepare karega |

### Correct flow

```text
Superadmin Service -> MySQL platform_settings -> event/cache invalidation -> consumer services
```

### Wrong flow

```text
Superadmin Service -> Search Typesense direct write
Superadmin Service -> CMS DB direct update
Superadmin Service -> Gateway config file direct edit
```

> 🟡 **Beginner note:** Source of truth ka matlab hai final authoritative data kahan stored hai. Behavior apply karna alag baat hai. Example: `maintenance_mode` setting Superadmin DB me stored hai, lekin request block karna API Gateway karega.

---

## Step 4: Permission Policy Define Kiya

Task 3 me ye permissions already define hain:

| Permission | Risk | Meaning |
|---|---|---|
| `platform:settings:read` | Medium | Platform settings view karna |
| `platform:settings:write` | Critical | Platform settings update karna |
| `search:synonyms:read` | Low | Search synonyms view karna |
| `search:synonyms:write` | Medium | Search synonyms update karna |

### Role-permission mapping

| Role | Settings Read | Settings Write | Synonyms Read | Synonyms Write | Reason |
|---|---:|---:|---:|---:|---|
| `superadmin` | ✅ | ✅ | ✅ | ✅ | Full platform owner |
| `catalog_admin` | ❌ | ❌ | ✅ | ✅ | Search/catalog improvements ka scope |
| `readonly_admin` | ❌ | ❌ | ✅ | ❌ | Read-only search visibility |
| `operations_admin` | ❌ | ❌ | ❌ | ❌ | Ops ka scope users/sellers/orders/sessions hai |
| `finance_admin` | ❌ | ❌ | ❌ | ❌ | Finance ka scope payments/refunds hai |

### Route permission map

| REST API | Required Permission | Allowed roles |
|---|---|---|
| `GET /api/v1/admin/settings` | `platform:settings:read` | `superadmin` |
| `PATCH /api/v1/admin/settings/{key}` | `platform:settings:write` | `superadmin` |
| `GET /api/v1/admin/search/synonyms` | `search:synonyms:read` | `superadmin`, `catalog_admin`, `readonly_admin` |
| `POST /api/v1/admin/search/synonyms` | `search:synonyms:write` | `superadmin`, `catalog_admin` |

> 🔴 **Security note:** API Gateway ka `auth: superadmin` marker broad route protection ke liye useful hai, but Superadmin Service ko final permission check khud karna chahiye.

---

## Step 5: Setting Contracts Banaye

Settings ko random JSON blob ki tarah accept nahi karna chahiye. Har supported setting key ka schema fixed hona chahiye.

### 5.1 `maintenance_mode`

```json
{
  "enabled": true,
  "message": "Platform maintenance is active",
  "starts_at": "2026-06-01T01:00:00Z",
  "ends_at": "2026-06-01T02:00:00Z",
  "allow_admins": true,
  "allow_health_checks": true
}
```

| Field | Rule |
|---|---|
| `enabled` | Required boolean |
| `message` | Required when enabled, max 200 chars |
| `starts_at` | Optional ISO timestamp |
| `ends_at` | Optional ISO timestamp, must be after `starts_at` |
| `allow_admins` | Default true, admin panel should remain accessible |
| `allow_health_checks` | Default true, `/healthz` should not break |

### 5.2 `commission_rules`

```json
{
  "default_rate_bps": 1200,
  "currency": "INR",
  "category_overrides": [
    {
      "category_id": "cat_electronics",
      "rate_bps": 900
    }
  ],
  "seller_overrides": [
    {
      "seller_id": "seller_123",
      "rate_bps": 700,
      "expires_at": "2026-12-31T23:59:59Z"
    }
  ],
  "effective_from": "2026-06-15T00:00:00Z"
}
```

| Field | Rule |
|---|---|
| `default_rate_bps` | Required, `0` to `5000` basis points |
| `currency` | Required, currently `INR` |
| `category_overrides` | Optional list, each rate `0` to `5000` |
| `seller_overrides` | Optional list, each seller id valid format |
| `effective_from` | Required for critical rate changes |

> 🟡 **Beginner note:** `bps` ka matlab basis points hota hai. `1200 bps` = `12%`. Is format se floating-point rounding bugs avoid hote hain.

### 5.3 `feature_flags`

```json
{
  "flags": {
    "new_checkout": {
      "enabled": false,
      "rollout_percent": 0,
      "allowed_roles": ["buyer"],
      "description": "New checkout experience"
    },
    "seller_bulk_upload": {
      "enabled": true,
      "rollout_percent": 25,
      "allowed_roles": ["seller"],
      "description": "Bulk product upload beta"
    }
  }
}
```

| Field | Rule |
|---|---|
| `flags` | Required object |
| `enabled` | Required boolean |
| `rollout_percent` | Integer `0` to `100` |
| `allowed_roles` | Optional known app roles only |
| `description` | Optional, max 300 chars |

### 5.4 `search_synonyms`

```json
{
  "synonyms": [
    {
      "root": "sneakers",
      "synonyms": ["shoes", "trainers", "sports shoes"]
    },
    {
      "root": "mobile",
      "synonyms": ["phone", "smartphone"]
    }
  ]
}
```

| Field | Rule |
|---|---|
| `root` | Required, normalized lowercase, max 80 chars |
| `synonyms` | 1 to 20 values |
| Each synonym | Trimmed, lowercase, duplicate-free |
| Total list | Keep bounded; large imports should use Search Service batch tooling later |

---

## Step 6: Database Design Blueprint

Docs already recommend `platform_settings(setting_key)` unique. Task 7 ke liye table ko versioning and audit context ke saath design karna better hai.

### Recommended migration

```sql
CREATE TABLE IF NOT EXISTS platform_settings (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  setting_key VARCHAR(128) NOT NULL,
  setting_type ENUM('maintenance', 'commission', 'feature_flags', 'search') NOT NULL,
  value_json JSON NOT NULL,
  risk_level ENUM('low', 'medium', 'high', 'critical') NOT NULL DEFAULT 'medium',
  version BIGINT UNSIGNED NOT NULL DEFAULT 1,
  updated_by_admin_id VARCHAR(64) NOT NULL,
  update_reason VARCHAR(512) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_platform_settings_key (setting_key),
  KEY idx_platform_settings_type (setting_type),
  KEY idx_platform_settings_updated (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### Why these columns?

| Column | Reason |
|---|---|
| `setting_key` | Stable lookup key like `maintenance_mode` |
| `setting_type` | Validation and grouping ke liye |
| `value_json` | Setting value flexible but controlled JSON |
| `risk_level` | Audit/approval policy select karne ke liye |
| `version` | Concurrent update conflict avoid karne ke liye |
| `updated_by_admin_id` | Last actor trace |
| `update_reason` | High-risk change reason mandatory |
| `updated_at` | Cache invalidation and UI display |

### Seed defaults

```sql
INSERT INTO platform_settings
  (setting_key, setting_type, value_json, risk_level, version, updated_by_admin_id, update_reason)
VALUES
  ('maintenance_mode', 'maintenance', JSON_OBJECT(
    'enabled', false,
    'message', '',
    'allow_admins', true,
    'allow_health_checks', true
  ), 'critical', 1, 'system', 'initial default'),
  ('commission_rules', 'commission', JSON_OBJECT(
    'default_rate_bps', 1000,
    'currency', 'INR',
    'category_overrides', JSON_ARRAY(),
    'seller_overrides', JSON_ARRAY()
  ), 'critical', 1, 'system', 'initial default'),
  ('feature_flags', 'feature_flags', JSON_OBJECT(
    'flags', JSON_OBJECT()
  ), 'high', 1, 'system', 'initial default'),
  ('search_synonyms', 'search', JSON_OBJECT(
    'synonyms', JSON_ARRAY()
  ), 'medium', 1, 'system', 'initial default')
ON DUPLICATE KEY UPDATE setting_key = setting_key;
```

> 🟢 **Implementation note:** Current backend migration folder already has RBAC and review-task migrations. Task 7 future implementation should add a new settings migration instead of editing old migrations.

---

## Step 7: API Contract Define Kiya

`api/master-api.json` me settings APIs already listed hain.

### 7.1 List platform settings

```http
GET /api/v1/admin/settings
Authorization: Bearer <admin_token>
X-Request-ID: req_123
```

Required permission:

```text
platform:settings:read
```

Example response:

```json
{
  "settings": [
    {
      "key": "maintenance_mode",
      "value": {
        "enabled": false,
        "message": "",
        "allow_admins": true,
        "allow_health_checks": true
      },
      "updated_at": "2026-05-31T10:00:00Z"
    }
  ]
}
```

### 7.2 Update platform setting

```http
PATCH /api/v1/admin/settings/maintenance_mode
Authorization: Bearer <admin_token>
Content-Type: application/json
X-Request-ID: req_456
```

Required permission:

```text
platform:settings:write
```

Request:

```json
{
  "value": {
    "enabled": true,
    "message": "Scheduled maintenance is active",
    "starts_at": "2026-06-01T01:00:00Z",
    "ends_at": "2026-06-01T02:00:00Z",
    "allow_admins": true,
    "allow_health_checks": true
  },
  "reason": "Database maintenance window"
}
```

Response:

```json
{
  "key": "maintenance_mode",
  "value": {
    "enabled": true,
    "message": "Scheduled maintenance is active",
    "starts_at": "2026-06-01T01:00:00Z",
    "ends_at": "2026-06-01T02:00:00Z",
    "allow_admins": true,
    "allow_health_checks": true
  },
  "updated_at": "2026-05-31T10:05:00Z"
}
```

### 7.3 Search synonyms route

Search synonyms ko generic settings route se manage kiya ja sakta hai, but better UX/API ke liye dedicated Search admin routes bhi useful hain.

```http
POST /api/v1/admin/search/synonyms
Authorization: Bearer <admin_token>
Content-Type: application/json
```

Required permission:

```text
search:synonyms:write
```

Request:

```json
{
  "root": "mobile",
  "synonyms": ["phone", "smartphone"]
}
```

> 🟡 **Design choice:** Generic `PATCH /admin/settings/search_synonyms` bulk config ke liye useful hai. Dedicated `/admin/search/synonyms` route day-to-day catalog admin workflow ke liye cleaner hai.

---

## Step 8: Domain Model Banaya

### Go domain structs example

```go
package domain

import "time"

type PlatformSettingKey string
type SettingType string

const (
    SettingMaintenanceMode PlatformSettingKey = "maintenance_mode"
    SettingCommissionRules PlatformSettingKey = "commission_rules"
    SettingFeatureFlags    PlatformSettingKey = "feature_flags"
    SettingSearchSynonyms  PlatformSettingKey = "search_synonyms"
)

const (
    SettingTypeMaintenance SettingType = "maintenance"
    SettingTypeCommission  SettingType = "commission"
    SettingTypeFeatureFlag SettingType = "feature_flags"
    SettingTypeSearch      SettingType = "search"
)

type PlatformSetting struct {
    Key              PlatformSettingKey
    Type             SettingType
    Value            map[string]any
    Risk             RiskLevel
    Version          uint64
    UpdatedByAdminID string
    UpdateReason     string
    UpdatedAt        time.Time
}
```

### Explanation

- `PlatformSettingKey` known keys ko restrict karta hai.
- `SettingType` validation route decide karta hai.
- `Value` JSON object store karta hai, but validation key-specific rahegi.
- `Version` update conflicts avoid karta hai.
- `UpdatedByAdminID` and `UpdateReason` audit trail ke liye important hain.

---

## Step 9: Repository Layer Design Kiya

Repository DB details ko usecase se separate rakhega.

```go
package usecase

import (
    "context"

    "ecommerce/superadmin-service/internal/domain"
)

type PlatformSettingsRepository interface {
    ListSettings(ctx context.Context) ([]domain.PlatformSetting, error)
    GetSetting(ctx context.Context, key domain.PlatformSettingKey) (domain.PlatformSetting, error)
    UpdateSetting(ctx context.Context, setting domain.PlatformSetting) (domain.PlatformSetting, error)
}
```

### MySQL update query example

```sql
UPDATE platform_settings
SET
  value_json = ?,
  version = version + 1,
  updated_by_admin_id = ?,
  update_reason = ?,
  updated_at = CURRENT_TIMESTAMP
WHERE setting_key = ?
  AND version = ?;
```

### Why version check?

Version check optimistic locking ke liye hai.

Example:

1. Admin A setting version 4 open karta hai.
2. Admin B version 4 update karke version 5 bana deta hai.
3. Admin A purani version 4 ke basis par update kare to reject hoga.

Isse accidental overwrite avoid hota hai.

---

## Step 10: Usecase Flow Build Kiya

### Update flow

```mermaid
flowchart TD
    A[Admin PATCH setting] --> B[Gateway verifies admin token]
    B --> C[Superadmin extracts AdminActor]
    C --> D[Check required permission]
    D --> E[Validate reason and MFA context]
    E --> F[Load current setting from MySQL]
    F --> G[Validate new JSON by setting key]
    G --> H[Update setting with version check]
    H --> I[Record audit-ready mutation event]
    I --> J[Publish PlatformSettingUpdated event]
    J --> K[Invalidate Gateway/Admin cache]
    K --> L[Return updated setting]
```

### Usecase code example

```go
package usecase

import (
    "context"

    "ecommerce/superadmin-service/internal/domain"
)

type SettingsEventPublisher interface {
    PublishPlatformSettingUpdated(ctx context.Context, setting domain.PlatformSetting) error
}

type PlatformSettingsService struct {
    repo      PlatformSettingsRepository
    authz     *AuthorizationService
    audit     AdminMutationRecorder
    publisher SettingsEventPublisher
}

func (s *PlatformSettingsService) UpdateSetting(
    ctx context.Context,
    actor domain.AdminActor,
    key domain.PlatformSettingKey,
    value map[string]any,
    reason string,
    expectedVersion uint64,
) (domain.PlatformSetting, error) {
    permission := permissionForSettingUpdate(key)
    if err := s.authz.RequireHighRiskPermission(ctx, actor, permission, reason); err != nil {
        return domain.PlatformSetting{}, err
    }

    current, err := s.repo.GetSetting(ctx, key)
    if err != nil {
        return domain.PlatformSetting{}, err
    }

    if err := ValidatePlatformSetting(key, value); err != nil {
        return domain.PlatformSetting{}, err
    }

    next := current
    next.Value = value
    next.Version = expectedVersion
    next.UpdatedByAdminID = actor.AdminID
    next.UpdateReason = reason

    updated, err := s.repo.UpdateSetting(ctx, next)
    if err != nil {
        return domain.PlatformSetting{}, err
    }

    _ = s.audit.RecordAdminMutation(ctx, domain.AuditRecord{
        ActorAdminID: actor.AdminID,
        Action:       "platform_setting.update",
        ResourceType: "platform_setting",
        ResourceID:   string(key),
        RequestID:    actor.RequestID,
        IPHash:       actor.IPHash,
        Reason:       reason,
        Before:       summarizeSetting(current),
        After:        summarizeSetting(updated),
    })

    _ = s.publisher.PublishPlatformSettingUpdated(ctx, updated)

    return updated, nil
}
```

### Permission helper

```go
func permissionForSettingUpdate(key domain.PlatformSettingKey) domain.Permission {
    switch key {
    case domain.SettingSearchSynonyms:
        return domain.PermissionSearchSynonymsWrite
    default:
        return domain.PermissionSettingsWrite
    }
}
```

> 🟢 **Why this helper?** Search synonyms catalog team ka domain hai, isliye `catalog_admin` ko allow kiya ja sakta hai. Maintenance, commission, aur feature flags sirf `superadmin` ke liye rahenge.

---

## Step 11: Validation Layer Define Kiya

Settings validation strict honi chahiye. Invalid config production behavior break kar sakti hai.

### Validator function

```go
package usecase

import "ecommerce/superadmin-service/internal/domain"

func ValidatePlatformSetting(key domain.PlatformSettingKey, value map[string]any) error {
    switch key {
    case domain.SettingMaintenanceMode:
        return validateMaintenanceMode(value)
    case domain.SettingCommissionRules:
        return validateCommissionRules(value)
    case domain.SettingFeatureFlags:
        return validateFeatureFlags(value)
    case domain.SettingSearchSynonyms:
        return validateSearchSynonyms(value)
    default:
        return domain.NewValidationError("unsupported platform setting key")
    }
}
```

### Validation checklist

| Setting | Required validations |
|---|---|
| `maintenance_mode` | Boolean enabled, message length, time window, admin/health access stays allowed |
| `commission_rules` | BPS range, currency, effective date, duplicate category/seller override check |
| `feature_flags` | Known flag names, rollout `0..100`, allowed roles known, no empty flag config |
| `search_synonyms` | Normalize text, remove duplicates, bounded list size, no empty root/synonym |

### Commission validation example

```go
func validateRateBPS(rate int) error {
    if rate < 0 || rate > 5000 {
        return domain.NewValidationError("commission rate must be between 0 and 5000 bps")
    }
    return nil
}
```

### Synonym normalization example

```go
func normalizeTerm(term string) string {
    term = strings.TrimSpace(strings.ToLower(term))
    term = strings.Join(strings.Fields(term), " ")
    return term
}
```

---

## Step 12: Downstream Sync Design Kiya

Settings update hone ke baad consumers ko update milna chahiye.

### Event payload

```json
{
  "event_id": "evt_setting_123",
  "event_type": "PlatformSettingUpdated",
  "setting_key": "maintenance_mode",
  "version": 5,
  "updated_by_admin_id": "admin_123",
  "updated_at": "2026-05-31T10:05:00Z"
}
```

### Consumer behavior

| Consumer | Event receive hone par kya karega |
|---|---|
| API Gateway | `maintenance_mode` and `feature_flags` cache invalidate karega |
| Search Service | `search_synonyms` update ko Typesense synonym config me apply karega |
| CMS Service | `commission_rules` cache invalidate karega |
| Superadmin Panel | Next GET par latest settings show karega |

### Mermaid sequence

```mermaid
sequenceDiagram
    participant Admin as Superadmin Panel
    participant GW as API Gateway
    participant SA as Superadmin Service
    participant DB as MySQL platform_settings
    participant Bus as Event Bus
    participant Search as Search Service
    participant CMS as CMS Service

    Admin->>GW: PATCH /api/v1/admin/settings/{key}
    GW->>SA: UpdatePlatformSetting(actor, key, value, reason)
    SA->>SA: Permission + validation
    SA->>DB: Update setting version
    SA->>SA: Record audit-ready mutation
    SA->>Bus: PlatformSettingUpdated
    Bus-->>Search: If search_synonyms changed
    Bus-->>CMS: If commission_rules changed
    Bus-->>GW: Invalidate maintenance/flag cache
    SA-->>GW: Updated setting
    GW-->>Admin: 200 OK
```

---

## Step 13: Maintenance Mode Enforcement

Maintenance mode enforcement Superadmin Service me nahi, Gateway me hoga.

### Gateway behavior

```text
Request arrives
-> Gateway loads cached maintenance_mode
-> If disabled: continue normally
-> If enabled:
   -> Allow health checks
   -> Allow admin routes if allow_admins = true
   -> Block buyer/seller public routes with 503
```

### Example response during maintenance

```json
{
  "code": "MAINTENANCE_MODE",
  "message": "Scheduled maintenance is active",
  "request_id": "req_789"
}
```

### Important rules

| Rule | Why |
|---|---|
| Health checks should stay open | Kubernetes/load balancer health should not fail unnecessarily |
| Admin routes should stay open | Superadmin must be able to disable maintenance mode |
| Message should be safe | No internal DB/deployment details expose nahi karne |
| Cache invalidation should be quick | Maintenance on/off delay user-facing hota hai |

---

## Step 14: Commission Rules Integration

Commission rules financial impact wali setting hai.

### Recommended usage

| Stage | Service | Usage |
|---|---|---|
| Seller dashboard estimate | CMS Service | Seller ko expected platform fee show karna |
| Order settlement calculation | Order/Payment/Settlement flow | Final commission snapshot calculate karna |
| Admin config | Superadmin Service | Commission rules validate and persist karna |

### Commission snapshot rule

Order/payment settlement me current commission setting ko directly mutable reference ki tarah use nahi karna chahiye. Us waqt applied commission snapshot store karo.

```json
{
  "order_id": "ord_123",
  "seller_id": "seller_123",
  "commission_rate_bps": 1200,
  "commission_amount": {
    "amount": 24000,
    "currency": "INR"
  },
  "setting_version": 8
}
```

> 🟡 **Why snapshot?** Agar kal commission 12% se 10% ho jaaye, purane orders ka settlement change nahi hona chahiye.

---

## Step 15: Search Synonyms Integration

Search synonyms ka operational owner Search Service hai.

### Synonym update flow

```mermaid
flowchart LR
    A[Catalog Admin] --> B[Superadmin RBAC]
    B --> C[Validate synonym text]
    C --> D[Store setting/change record]
    D --> E[Publish event]
    E --> F[Search Service]
    F --> G[Typesense synonym config]
```

### Search Service contract idea

```go
type SearchAdminClient interface {
    CreateSynonym(ctx context.Context, root string, synonyms []string) error
    ListSynonyms(ctx context.Context) ([]SearchSynonym, error)
}
```

### Why not direct Typesense call from Superadmin?

| Reason | Explanation |
|---|---|
| Ownership | Search Service owns Typesense details |
| Encapsulation | Superadmin ko search index schema nahi pata hona chahiye |
| Safer changes | Search Service validation + retry + provider-specific handling kar sakta hai |
| Future flexibility | Kal Typesense replace hua to Superadmin code nahi badlega |

---

## Step 16: Feature Flags Integration

Feature flags controlled rollout ke liye useful hain.

### Recommended flag evaluation

```text
Feature enabled?
-> role allowed?
-> rollout percent allows user/session?
-> service-specific safety check passes?
```

### Deterministic rollout example

```go
func InRollout(stableID string, rolloutPercent int) bool {
    if rolloutPercent <= 0 {
        return false
    }
    if rolloutPercent >= 100 {
        return true
    }
    bucket := crc32.ChecksumIEEE([]byte(stableID)) % 100
    return int(bucket) < rolloutPercent
}
```

### Why deterministic?

Same user/session ko baar-baar same result milna chahiye. Agar random use karenge to user kabhi new checkout dekhega, kabhi old checkout. UX confusing ho jayega.

---

## Step 17: Caching Strategy

`docs/12-logging-monitoring-scalability.md` ke according platform settings cache TTL:

```text
Platform settings: Redis, 5 min TTL
```

### Cache keys

```text
platform_settings:all
platform_settings:key:maintenance_mode
platform_settings:key:feature_flags
platform_settings:key:commission_rules
platform_settings:key:search_synonyms
```

### Cache rules

| Rule | Detail |
|---|---|
| Read-through cache | Missing cache par DB read, phir cache set |
| TTL | 5 minutes default |
| Update invalidation | Update ke baad key-specific cache delete |
| Event invalidation | Gateway/CMS/Search local caches event se clear |
| Critical setting | Maintenance mode cache short or event-first invalidation |

### Redis usage example

```bash
docker run --name ecommerce-redis -p 6379:6379 redis:7.2-alpine
```

> 🟢 **MVP option:** Agar Redis integration ready nahi hai, local in-memory cache + short TTL use kiya ja sakta hai. Production me multi-replica consistency ke liye Redis/event invalidation better rahega.

---

## Step 18: Audit and Approval Rules

Task 8 final immutable audit logs ka scope hai, but Task 7 mutation audit-ready honi chahiye.

### Every setting update should capture

| Field | Example |
|---|---|
| `actor_admin_id` | `admin_123` |
| `action` | `platform_setting.update` |
| `resource_type` | `platform_setting` |
| `resource_id` | `maintenance_mode` |
| `request_id` | `req_456` |
| `ip_hash` | Hashed admin IP |
| `before` | Old value summary |
| `after` | New value summary |
| `reason` | "Database maintenance window" |

### Maker-checker recommendation

| Setting | Direct update? | Maker-checker recommended? |
|---|---:|---:|
| `maintenance_mode.enabled=true` | ✅ | Optional for scheduled maintenance |
| `commission_rules` | ⚠️ | ✅ Yes |
| `feature_flags` low-risk beta | ✅ | Optional |
| `feature_flags` checkout/payment flag | ⚠️ | ✅ Yes |
| `search_synonyms` | ✅ | Usually no |

> 🔴 **Important:** Maker-checker final approval queue can reuse `admin_review_tasks`, but full approval workflow should not be expanded beyond Task 7 unless explicitly required.

---

## Step 19: HTTP Handler Shape

### DTOs

```go
type platformSettingDTO struct {
    Key       string         `json:"key"`
    Value     map[string]any `json:"value"`
    UpdatedAt string         `json:"updated_at"`
}

type platformSettingsResponse struct {
    Settings []platformSettingDTO `json:"settings"`
}

type platformSettingInput struct {
    Value   map[string]any `json:"value"`
    Reason  string         `json:"reason"`
    Version uint64         `json:"version,omitempty"`
}
```

### Handler flow

```go
func (h *SettingsHandler) updateSetting(w http.ResponseWriter, r *http.Request) {
    actor, err := adminActorFromRequest(r)
    if err != nil {
        writeError(w, r, err)
        return
    }

    key := domain.PlatformSettingKey(r.PathValue("key"))

    var input platformSettingInput
    if err := decodeJSON(r, &input); err != nil {
        writeError(w, r, err)
        return
    }

    setting, err := h.service.UpdateSetting(
        r.Context(),
        actor,
        key,
        input.Value,
        input.Reason,
        input.Version,
    )
    if err != nil {
        writeError(w, r, err)
        return
    }

    writeJSON(w, http.StatusOK, toSettingDTO(setting))
}
```

### Beginner explanation

- Handler sirf request/response handle karta hai.
- Permission, validation, DB update usecase layer me hota hai.
- Repository SQL details handle karta hai.
- Is separation se tests easy aur code maintainable hota hai.

---

## Step 20: Error Handling

### Expected errors

| Scenario | HTTP | Code | Message idea |
|---|---:|---|---|
| Missing token | 401 | `UNAUTHORIZED` | Admin login required |
| Permission missing | 403 | `FORBIDDEN` | Required permission missing |
| Unknown setting key | 400 | `VALIDATION_ERROR` | Unsupported platform setting key |
| Invalid JSON contract | 400 | `VALIDATION_ERROR` | Invalid setting value |
| Version conflict | 409 | `SETTING_VERSION_CONFLICT` | Setting was updated by another admin |
| Downstream sync failed | 202/500 | Depends | Setting saved but sync pending, or update rejected |
| DB unavailable | 503 | `DATABASE_UNAVAILABLE` | Retry later |

### Error response example

```json
{
  "code": "FORBIDDEN",
  "message": "required permission missing",
  "required_permission": "platform:settings:write",
  "request_id": "req_456"
}
```

---

## Step 21: Testing Plan

### Unit tests

| Test | Expected |
|---|---|
| `superadmin` can read settings | Pass |
| `superadmin` can update maintenance mode | Pass |
| `catalog_admin` cannot update commission | Forbidden |
| `catalog_admin` can update search synonyms | Pass |
| `readonly_admin` can read synonyms but cannot write | Pass |
| Invalid commission rate rejected | Validation error |
| Maintenance `ends_at` before `starts_at` rejected | Validation error |
| Feature flag rollout over 100 rejected | Validation error |
| Duplicate synonyms normalized/rejected | Validation error |
| Version conflict returns conflict | Pass |

### Example Go test

```go
func TestCatalogAdminCanOnlyUpdateSearchSynonyms(t *testing.T) {
    actor := domain.AdminActor{
        AdminID: "admin_catalog",
        Roles:   []domain.AdminRole{domain.RoleCatalog},
    }

    err := service.UpdateSetting(
        context.Background(),
        actor,
        domain.SettingCommissionRules,
        map[string]any{"default_rate_bps": 1200},
        "trying commission update",
        1,
    )

    require.Error(t, err)
    require.True(t, domain.IsForbidden(err))
}
```

### Integration tests

| Flow | Verify |
|---|---|
| PATCH maintenance mode | DB updated, event published, cache invalidated |
| PATCH commission rules | Validation + version update + audit-ready record |
| PATCH search synonyms | Permission accepts `catalog_admin`, event sent to Search Service |
| GET settings | Only `platform:settings:read` actor sees general settings |

---

## 📁 Clean Implementation Folder Structure

Recommended future backend structure for Task 7:

```text
backend/
└── services/
    └── superadmin-service/
        ├── migrations/
        │   ├── 005_create_platform_settings.up.sql
        │   └── 005_create_platform_settings.down.sql
        └── internal/
            ├── domain/
            │   └── platform_setting.go
            ├── repository/
            │   └── mysql_platform_settings_repository.go
            ├── usecase/
            │   ├── platform_settings.go
            │   ├── platform_settings_validation.go
            │   └── platform_settings_test.go
            ├── transport/
            │   └── http/
            │       ├── settings_handler.go
            │       └── settings_handler_test.go
            └── clients/
                ├── search_admin_client.go
                └── settings_event_publisher.go

TaskImplementation/
└── Superadmin Service/
    └── task7.md
```

> 🟢 **Current task output:** Sirf `TaskImplementation/Superadmin Service/task7.md` create kiya gaya. Upar wala backend folder structure future implementation guide hai.

---

## 🧩 External Libraries / Tools

Task 7 documentation create karne ke liye koi new external library install nahi ki gayi. Future backend implementation me ye tools useful honge:

| Tool / Library | What it is | Why used | Install | Usage |
|---|---|---|---|---|
| MySQL 8.x | Relational database | `platform_settings` structured + auditable storage ke liye | Docker/local package | `platform_settings` table store karega |
| `github.com/go-sql-driver/mysql` | Go MySQL driver | `database/sql` ko MySQL se connect karne ke liye | `go get github.com/go-sql-driver/mysql` | Blank import: `_ "github.com/go-sql-driver/mysql"` |
| `golang-migrate/migrate` | SQL migration CLI | Versioned settings migrations run/rollback karne ke liye | `go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest` | `migrate up` |
| gRPC / Protobuf | Internal service contract tooling | Superadmin -> Search/CMS/Gateway typed integration ke liye | `go get google.golang.org/grpc google.golang.org/protobuf` | Generated clients use karo |
| Redis | Cache/invalidation support | Platform settings 5 min TTL cache ke liye | `docker run --name ecommerce-redis -p 6379:6379 redis:7.2-alpine` | Cache `platform_settings:key:*` |
| Event bus | Async invalidation/sync | `PlatformSettingUpdated` consumers notify karne ke liye | Project infra ke according Kafka/RabbitMQ | Consumers cache clear/apply karenge |

### Example migrate command

```bash
migrate \
  -path backend/services/superadmin-service/migrations \
  -database "mysql://root:localroot@tcp(localhost:3306)/superadmin_db" \
  up
```

### Example Go install commands

```bash
go get github.com/go-sql-driver/mysql
go get google.golang.org/grpc google.golang.org/protobuf
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

---

## 🧱 Architecture Diagram

```mermaid
flowchart TB
    Panel[Superadmin Panel] --> GW[API Gateway]
    GW --> SA[Superadmin Service]
    SA --> RBAC[Admin RBAC]
    SA --> DB[(MySQL: platform_settings)]
    SA --> Audit[Audit-ready mutation record]
    SA --> Bus[PlatformSettingUpdated Event]

    Bus --> GatewayCache[Gateway settings cache]
    Bus --> Search[Search Service]
    Bus --> CMS[CMS Service]

    Search --> Typesense[(Typesense)]
    CMS --> SellerFlows[Seller/commission consumers]
    GatewayCache --> PublicTraffic[Public buyer/seller traffic]
```

### Diagram explanation

1. Admin panel request Gateway ke through aata hai.
2. Superadmin Service exact permission check karta hai.
3. Valid setting MySQL me save hoti hai.
4. Audit-ready mutation record create hota hai.
5. Event publish hota hai.
6. Gateway/Search/CMS apna cache ya behavior update karte hain.

---

## 🔐 Security Checklist

| Check | Status | Detail |
|---|---:|---|
| General settings read protected | ✅ | `platform:settings:read` |
| General settings write protected | ✅ | `platform:settings:write` |
| Search synonyms separate permission | ✅ | `search:synonyms:write` for catalog admins |
| Reason mandatory | ✅ | High-risk update ke liye |
| MFA recommended | ✅ | Critical updates ke liye |
| Admin actor captured | ✅ | Audit-ready mutation |
| Version conflict handled | ✅ | Optimistic locking |
| Unknown keys rejected | ✅ | No arbitrary JSON settings |
| Direct downstream DB writes avoided | ✅ | Service boundary maintained |
| Sensitive internals hidden | ✅ | Maintenance message safe |

---

## ✅ Completion Checklist

| Requirement | Status | Notes |
|---|---:|---|
| `TaskImplementation/` folder exists | ✅ | Existing folder preserved |
| `TaskImplementation/Superadmin Service/` folder exists | ✅ | Existing folder preserved |
| `task7.md` created | ✅ | This file |
| Hinglish step-by-step guide | ✅ | Beginner-friendly explanation included |
| External libraries/tools mentioned | ✅ | MySQL, driver, migrate, gRPC, Redis, event bus |
| Install/use examples included | ✅ | `go get`, `go install`, `docker run`, `migrate` |
| Clean folder structure included | ✅ | Current output + future backend structure |
| Code examples included | ✅ | SQL, Go domain/usecase/handler/validation/test snippets |
| Diagrams included | ✅ | Mermaid flowchart and sequence diagrams |
| Scope limited to Superadmin Task 7 | ✅ | No source code implementation beyond documentation |

---

## 🚫 Out of Scope for Task 7

| Not Done | Reason |
|---|---|
| Real backend migration added | User requested required folder structure and `task7.md` content only |
| Actual Settings API handler implemented | This task output is documentation-only |
| Superadmin Panel settings UI built | Frontend task has separate scope |
| Full immutable audit log system built | Superadmin Service Task 8 |
| Search full reindex implemented | Search Service Task 8 |
| CMS commission settlement implemented | CMS/Order/Payment future workflows |

---

## ✅ Final Task 7 Standard

Superadmin Service Task 7 ka platform settings model ab clearly documented hai:

- `maintenance_mode`, `commission_rules`, `feature_flags`, aur `search_synonyms` ke contracts defined hain.
- General platform settings sirf `superadmin` update karega.
- Search synonyms `catalog_admin` bhi update kar sakta hai through dedicated permission.
- Settings MySQL me versioned JSON ke form me store hongi.
- Updates permission, reason, validation, audit-ready record, event publish, aur cache invalidation ke saath honge.
- Superadmin Service downstream services ke DB/index ko direct mutate nahi karegi.

> 🟢 **Task 7 complete:** Ab platform-level settings safely implement karne ke liye clear Superadmin-side blueprint ready hai, without Search/CMS/Gateway ownership boundaries ko mix kiye.
