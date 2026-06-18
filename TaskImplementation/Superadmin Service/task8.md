# 🛡️ Superadmin Service - Task 8: Admin Audit Logs

![Task](https://img.shields.io/badge/Task-Superadmin%20Service%20%238-blue)
![Priority](https://img.shields.io/badge/Priority-P0-red)
![Dependency](https://img.shields.io/badge/Dependency-Schema-yellow)
![Status](https://img.shields.io/badge/Status-Documented-success)
![Security](https://img.shields.io/badge/Security-Immutable%20Audit-critical)
![Database](https://img.shields.io/badge/Database-MySQL%208-orange)
![Guide](https://img.shields.io/badge/Guide-Hinglish-ff69b4)

## 📌 Task Summary

| Field | Detail |
|---|---|
| Task Name | Admin audit logs |
| Source | `docs/01-micro-tasks.md` -> `Superadmin Service` -> Task 8 |
| Goal | Har admin action ko immutable audit log me store karna |
| Dependency | Schema |
| Priority | P0 |
| Output Type | Documentation-only implementation guide |
| Not Included | Superadmin Panel audit viewer UI, object-storage archival job, OpenSearch read model, SIEM integration, unrelated service changes |

> **Simple Hinglish goal:** Is task ka kaam Superadmin Service ke andar ek reliable audit-log system define karna hai. Jab bhi admin koi sensitive action kare, jaise user block, seller suspend, refund review, platform setting update, ya audit export, us action ka immutable record MySQL me store hona chahiye. Record me actor, action, resource, reason, request id, IP hash, before/after summary, aur timestamp hona must hai.

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
    ├── task7.md
    └── task8.md
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
| `task7.md` | Platform settings guide |
| `task8.md` | Sirf **Superadmin Service - Task 8** ka immutable admin audit-log guide |

> 🟢 **Important:** `Superadmin Service` folder already present tha, isliye usko keep kiya gaya. Is task me backend source code, migration, proto, ya frontend file create nahi ki gayi. Ye beginner-friendly implementation guide hai.

---

## 📚 Documents Studied

| Document/File | Kya use kiya gaya |
|---|---|
| `docs/01-micro-tasks.md` | Task 8 ka exact scope: "Har admin action immutable audit log me store karo" |
| `docs/04-microservice-design.md` | Superadmin Service responsibilities, `ListAuditLogs`, REST route, and transaction expectation |
| `docs/05-database-design.md` | Superadmin DB tables and audit indexes |
| `docs/06-auth-security.md` | Audit fields, high-risk operations, and sensitive-data logging rules |
| `docs/09-cms-superadmin.md` | Admin audit log fields and high-risk controls |
| `docs/03-folder-structure.md` | Future `superadmin-service` folder layout and audit feature location |
| `api/master-api.json` | `GET /api/v1/admin/audit-logs` and `ListAuditLogs` contract |
| `TaskImplementation/Superadmin Service/task1.md` | Domain rule: every mutation requires audit intent |
| `TaskImplementation/Superadmin Service/task2.md` | MySQL schema direction for `admin_audit_logs` |
| `TaskImplementation/Superadmin Service/task3.md` | Permissions: `audit:logs:read`, `audit:logs:export` |
| `TaskImplementation/Superadmin Service/task4.md` | User/seller mutations already prepare audit payloads |
| `TaskImplementation/Superadmin Service/task5.md` | Order/payment controls need audit for refund/manual review |
| `TaskImplementation/Superadmin Service/task6.md` | Session visibility read access should be logged/auditable |
| `TaskImplementation/Superadmin Service/task7.md` | Platform setting mutations need final immutable audit writer |
| `backend/services/superadmin-service/internal/domain/admin_audit.go` | Existing `AuditRecord` shape for mutation audit payloads |
| `backend/services/superadmin-service/internal/usecase/audit_recorder.go` | Current temporary logger-based audit recorder |
| `backend/services/superadmin-service/internal/domain/permission.go` | Audit permissions already exist |
| `backend/services/superadmin-service/internal/rbac/matrix.go` | `superadmin` and `readonly_admin` audit read access direction |

---

## 🧭 Implementation Approach

Task 8 ko **immutable audit layer** treat kiya gaya. Iska matlab normal application log enough nahi hai. Application logs debugging ke liye hote hain; audit logs compliance, security review, dispute investigation, aur accountability ke liye hote hain.

### Final decision

```text
Source of truth: Superadmin MySQL admin_audit_logs
Write style: append-only insert
Update/delete: application level par disallowed, DB permission se restricted
Reader API: GET /api/v1/admin/audit-logs
Read permission: audit:logs:read
Export permission: audit:logs:export, future export endpoint/UI ke liye
Existing temporary recorder: LoggingAuditRecorder
Task 8 target recorder: MySQLAuditRecorder / AuditLogService
```

### One-line reason

Admin action ke baad ye answer hamesha milna chahiye:

```text
Kis admin ne kya action kiya, kis resource par kiya, kyun kiya,
kis request/session/IP context me kiya, aur before/after state kya thi?
```

---

## 🪜 Step-by-Step Implementation

## Step 1: Task Boundary Clear Kiya

`docs/01-micro-tasks.md` me Superadmin Service Task 8 ye hai:

| S.No | Task Name | Detail | Dependency | Priority |
|---:|---|---|---|---|
| 8 | Admin audit logs | Har admin action immutable audit log me store karo. Compliance and security ke liye must hai. | Schema | P0 |

### Is task me allowed work

- `admin_audit_logs` schema define karna
- Audit domain model define karna
- Existing `AuditRecorder` ko MySQL-backed recorder se replace karne ka guide
- Admin mutation flows me audit write point clear karna
- `GET /api/v1/admin/audit-logs` list/filter API define karna
- RBAC permission `audit:logs:read` enforce karna
- High-risk export permission `audit:logs:export` future-safe rakhna
- PII minimization and immutable rules define karna
- Tests and verification checklist define karna

### Is task me not allowed work

| Area | Reason |
|---|---|
| Superadmin Panel audit table UI | Ye **Superadmin Panel - Task 8** ka scope hai |
| Audit export CSV/download endpoint | API contract me current route sirf list hai; export future route ho sakta hai |
| OpenSearch/SIEM pipeline | Ye scaling/monitoring future phase hai |
| Object storage archival job | Retention/archive future operational task hai |
| User/seller/order/payment business logic change | Wo Tasks 4 and 5 me covered hain |
| Platform settings behavior change | Wo Task 7 ka scope hai |

> 🔴 **Boundary rule:** Task 8 sirf admin audit-log storage, query, permission, and integration points define karta hai. Ye unrelated service data ko direct mutate nahi karega.

---

## Step 2: Audit Events Identify Kiye

Har admin action audit-worthy nahi hota, but Superadmin domain me risky actions definitely audit hone chahiye.

| Category | Example Action | Audit Required? | Reason |
|---|---|:---:|---|
| User control | Block/unblock user | ✅ | Account access affect hota hai |
| Seller control | Approve/reject/suspend seller | ✅ | Seller revenue and marketplace trust affect hota hai |
| Order operations | Manual order review create | ✅ | Support/compliance trace chahiye |
| Payment operations | Refund approve/reject | ✅ | Financial decision hai |
| Session visibility | Risky session view/export attempt | ✅ | Privacy-sensitive data access hai |
| Platform settings | Maintenance mode, commission, feature flag update | ✅ | Platform-wide impact hai |
| Search/catalog | Search synonym update, moderation decision | ✅ | Discovery/catalog behavior affect hota hai |
| Admin access | Audit log export | ✅ | Audit data itself sensitive hai |
| Read-only list | Normal user/order list | Optional | Request logs enough ho sakte hain, but high-risk reads log kar sakte hain |

### Recommended action keys

| Action Key | Resource Type | Source Task |
|---|---|---|
| `user.status.blocked` | `user` | Task 4 |
| `user.status.active` | `user` | Task 4 |
| `seller.status.active` | `seller` | Task 4 |
| `seller.status.rejected` | `seller` | Task 4 |
| `seller.status.suspended` | `seller` | Task 4 |
| `refund.review.approved` | `refund` | Task 5 |
| `refund.review.rejected` | `refund` | Task 5 |
| `order.manual_review.create` | `order` | Task 5 |
| `session.analytics.access` | `session` or `session_analytics` | Task 6 |
| `platform.setting.updated` | `platform_setting` | Task 7 |
| `search.synonym.updated` | `search_synonym` | Task 7 |
| `audit.logs.exported` | `admin_audit_log` | Task 8 future export |

---

## Step 3: Audit Record Contract Banaya

Docs me required fields already defined hain:

- actor admin id
- action
- resource type
- resource id
- request id
- IP hash
- before summary
- after summary
- reason
- created at

### Final audit record example

```json
{
  "audit_id": "aud_01JZ4ADMIN8V4T4X9P01",
  "actor_admin_id": "admin_789",
  "action": "seller.status.suspended",
  "resource_type": "seller",
  "resource_id": "seller_456",
  "request_id": "req_abc123",
  "session_id": "sess_admin_123",
  "ip_hash": "sha256:1f3d...",
  "reason": "Repeated counterfeit complaints verified by operations team",
  "before_json": {
    "status": "active"
  },
  "after_json": {
    "status": "suspended"
  },
  "created_at": "2026-05-31T12:30:00Z"
}
```

### Field explanation

| Field | Required? | Explanation |
|---|:---:|---|
| `audit_id` | ✅ | Public-safe unique audit id |
| `actor_admin_id` | ✅ | Kis admin ne action kiya |
| `action` | ✅ | Machine-readable action key |
| `resource_type` | ✅ | Resource category: `user`, `seller`, `refund`, etc. |
| `resource_id` | ✅ | Target resource id |
| `request_id` | ✅ for mutations | API request trace connect karta hai |
| `session_id` | Recommended | Admin session trace ke liye useful |
| `ip_hash` | Recommended | Raw IP store nahi karna; hash safer hai |
| `reason` | ✅ for high-risk | Admin ne action kyun liya |
| `before_json` | Recommended | Change se pehle ka safe summary |
| `after_json` | Recommended | Change ke baad ka safe summary |
| `created_at` | ✅ | Audit event timestamp |

> 🟡 **Beginner note:** `before_json` and `after_json` me full user profile, card details, token, password, OTP, ya raw secret nahi aana chahiye. Sirf safe summary store karo.

---

## Step 4: MySQL Table Schema Design Kiya

Task 2 me base schema already planned tha. Task 8 me usko production-ready append-only direction ke saath finalize karte hain.

### Migration file

```text
backend/services/superadmin-service/migrations/
├── 005_create_admin_audit_logs.up.sql
└── 005_create_admin_audit_logs.down.sql
```

### `005_create_admin_audit_logs.up.sql`

```sql
CREATE TABLE IF NOT EXISTS admin_audit_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  audit_id VARCHAR(64) NOT NULL,
  actor_admin_id VARCHAR(64) NOT NULL,
  action VARCHAR(128) NOT NULL,
  resource_type VARCHAR(64) NOT NULL,
  resource_id VARCHAR(128) NOT NULL,
  request_id VARCHAR(64) NULL,
  session_id VARCHAR(64) NULL,
  ip_hash CHAR(64) NULL,
  reason VARCHAR(512) NULL,
  before_json JSON NULL,
  after_json JSON NULL,
  metadata_json JSON NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_audit_logs_audit_id (audit_id),
  KEY idx_admin_audit_actor_created (actor_admin_id, created_at),
  KEY idx_admin_audit_resource (resource_type, resource_id),
  KEY idx_admin_audit_action_created (action, created_at),
  KEY idx_admin_audit_request (request_id),
  KEY idx_admin_audit_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### `005_create_admin_audit_logs.down.sql`

```sql
DROP TABLE IF EXISTS admin_audit_logs;
```

### Why these indexes?

| Index | Query use |
|---|---|
| `uk_admin_audit_logs_audit_id` | Duplicate audit id avoid karta hai |
| `idx_admin_audit_actor_created` | Actor-wise audit history fast hoti hai |
| `idx_admin_audit_resource` | Resource timeline, jaise seller/refund/order investigation |
| `idx_admin_audit_action_created` | Action-wise security review |
| `idx_admin_audit_request` | Request trace se audit lookup |
| `idx_admin_audit_created` | Time-range pagination and retention jobs |

---

## Step 5: Immutability Rules Define Kiye

Immutable ka simple meaning: audit row create hone ke baad application usko update/delete nahi karegi.

### Application rules

| Rule | Implementation |
|---|---|
| Insert only | Repository me sirf `Insert` and `List` methods hon |
| No update method | `UpdateAuditLog` type method create hi na ho |
| No delete method | `DeleteAuditLog` application code me na ho |
| Append correction | Galti fix karni ho to new audit row create karo |
| Sensitive data block | Secrets/raw PII before insert sanitize karo |

### Database rules

Production me DB user permissions alag rakho:

```text
superadmin_app_user:
  allowed: SELECT, INSERT on admin_audit_logs
  denied: UPDATE, DELETE on admin_audit_logs
```

> 🔴 **Important:** MySQL table technically mutable hoti hai agar privileged user update/delete kare. Immutability enforce karne ke liye app code + DB permissions + backups + monitoring ka combination use hota hai.

---

## Step 6: Domain Model Define Kiya

Existing code me `domain.AuditRecord` already hai:

```go
type AuditRecord struct {
    ActorAdminID string            `json:"actor_admin_id"`
    Action       string            `json:"action"`
    ResourceType string            `json:"resource_type"`
    ResourceID   string            `json:"resource_id"`
    RequestID    string            `json:"request_id,omitempty"`
    IPHash       string            `json:"ip_hash,omitempty"`
    Reason       string            `json:"reason,omitempty"`
    Before       map[string]string `json:"before_json,omitempty"`
    After        map[string]string `json:"after_json,omitempty"`
}
```

Task 8 me is contract ko list/query ke liye extend karna useful hoga.

### Suggested domain file

```text
backend/services/superadmin-service/internal/domain/admin_audit.go
```

### Code example

```go
package domain

import (
    "strings"
    "time"
)

type AuditLog struct {
    AuditID      string            `json:"audit_id"`
    ActorAdminID string            `json:"actor_admin_id"`
    Action       string            `json:"action"`
    ResourceType string            `json:"resource_type"`
    ResourceID   string            `json:"resource_id"`
    RequestID    string            `json:"request_id,omitempty"`
    SessionID    string            `json:"session_id,omitempty"`
    IPHash       string            `json:"ip_hash,omitempty"`
    Reason       string            `json:"reason,omitempty"`
    Before       map[string]string `json:"before_json,omitempty"`
    After        map[string]string `json:"after_json,omitempty"`
    CreatedAt    time.Time         `json:"created_at"`
}

type AuditLogListRequest struct {
    ActorAdminID string
    Action       string
    ResourceType string
    ResourceID   string
    From         time.Time
    To           time.Time
    Pagination   Pagination
}

type AuditLogListResponse struct {
    Logs       []AuditLog `json:"logs"`
    NextCursor string     `json:"next_cursor,omitempty"`
}

func (r AuditRecord) Validate() error {
    if strings.TrimSpace(r.ActorAdminID) == "" {
        return NewValidationError("actor_admin_id is required")
    }
    if strings.TrimSpace(r.Action) == "" {
        return NewValidationError("action is required")
    }
    if strings.TrimSpace(r.ResourceType) == "" {
        return NewValidationError("resource_type is required")
    }
    if strings.TrimSpace(r.ResourceID) == "" {
        return NewValidationError("resource_id is required")
    }
    return nil
}
```

### Explanation

- `AuditRecord` write input hai.
- `AuditLog` persisted/read output hai.
- `AuditLogListRequest` filters carry karta hai.
- `Validate()` empty actor/action/resource ko block karta hai.

---

## Step 7: PII-Safe Snapshot Rule Banaya

Audit log useful hona chahiye, lekin dangerous data store nahi karna chahiye.

### Safe vs unsafe examples

| Field | Safe? | Reason |
|---|:---:|---|
| `status: active -> suspended` | ✅ | Investigation ke liye enough |
| `refund_amount: 2500` | ✅ | Financial decision summary |
| `email: masked p***@example.com` | ✅ | Masked value acceptable |
| Full email/phone/address | ⚠️ | Avoid unless strictly required |
| Password/OTP/token | ❌ | Kabhi store nahi karna |
| Raw card data | ❌ | Platform servers ko touch nahi karna chahiye |
| Raw IP address | ❌ | Hash store karo |

### Sanitizer code example

```go
func SafeAuditSnapshot(values map[string]string) map[string]string {
    out := make(map[string]string, len(values))
    blocked := map[string]struct{}{
        "password": {},
        "otp": {},
        "token": {},
        "access_token": {},
        "refresh_token": {},
        "card_number": {},
        "cvv": {},
    }

    for key, value := range values {
        normalized := strings.ToLower(strings.TrimSpace(key))
        if _, unsafe := blocked[normalized]; unsafe {
            continue
        }
        out[key] = value
    }
    return out
}
```

> 🟢 **Simple rule:** Audit me enough context rakho, full sensitive data nahi.

---

## Step 8: Repository Layer Banaya

Repository DB details hide karta hai. Usecase ko SQL query pata nahi honi chahiye.

### Suggested file

```text
backend/services/superadmin-service/internal/repository/mysql_audit_log_repository.go
```

### Repository interface

```go
type AuditLogRepository interface {
    InsertAuditLog(ctx context.Context, record domain.AuditRecord) error
    ListAuditLogs(ctx context.Context, filter domain.AuditLogListRequest) (domain.AuditLogListResponse, error)
}
```

### Insert code example

```go
func (r *MySQLAuditLogRepository) InsertAuditLog(ctx context.Context, record domain.AuditRecord) error {
    if err := record.Validate(); err != nil {
        return err
    }

    beforeJSON, err := json.Marshal(domain.SafeAuditSnapshot(record.Before))
    if err != nil {
        return err
    }
    afterJSON, err := json.Marshal(domain.SafeAuditSnapshot(record.After))
    if err != nil {
        return err
    }

    _, err = r.db.ExecContext(ctx, `
        INSERT INTO admin_audit_logs (
            audit_id,
            actor_admin_id,
            action,
            resource_type,
            resource_id,
            request_id,
            session_id,
            ip_hash,
            reason,
            before_json,
            after_json
        )
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CAST(? AS JSON), CAST(? AS JSON))
    `,
        newAuditID(),
        record.ActorAdminID,
        record.Action,
        record.ResourceType,
        record.ResourceID,
        record.RequestID,
        record.SessionID,
        record.IPHash,
        record.Reason,
        string(beforeJSON),
        string(afterJSON),
    )
    return err
}
```

### Beginner explanation

- `InsertAuditLog` sirf insert karta hai.
- `newAuditID()` public audit id generate karta hai.
- `json.Marshal` maps ko JSON me convert karta hai.
- `CAST(? AS JSON)` MySQL ko batata hai ki value JSON column ke liye hai.

> 🟡 **Note:** Existing `AuditRecord` me `SessionID` field nahi hai. Actual code implementation me ya to field add karo, ya session id `metadata_json` me store karo.

---

## Step 9: Audit ID Generate Kiya

Audit id predictable nahi hona chahiye. External library avoid karni ho to Go standard library se secure random id generate kar sakte hain.

```go
func newAuditID() string {
    var b [16]byte
    if _, err := rand.Read(b[:]); err != nil {
        panic(err)
    }
    return "aud_" + hex.EncodeToString(b[:])
}
```

### Imports

```go
import (
    "crypto/rand"
    "encoding/hex"
)
```

### Why standard library?

| Choice | Decision |
|---|---|
| `crypto/rand` | Secure random id ke liye built-in |
| UUID external library | Required nahi, dependency avoid hoti hai |
| Auto-increment id expose karna | Avoid, internal DB sequence leak hota hai |

---

## Step 10: List Audit Logs Query Banayi

Audit logs security team/admin ko filterable milne chahiye.

### API contract

```text
GET /api/v1/admin/audit-logs
```

### Supported query params

| Query Param | Example | Purpose |
|---|---|---|
| `actor_id` | `admin_123` | Specific admin ke actions |
| `action` | `refund.review.approved` | Specific action type |
| `resource_type` | `seller` | Resource category |
| `resource_id` | `seller_456` | Resource timeline |
| `from` | `2026-05-01T00:00:00Z` | Start time |
| `to` | `2026-05-31T23:59:59Z` | End time |
| `page_size` | `50` | Pagination size |
| `cursor` | `100928` | Next page cursor |

### SQL example

```sql
SELECT
  audit_id,
  actor_admin_id,
  action,
  resource_type,
  resource_id,
  request_id,
  session_id,
  ip_hash,
  reason,
  before_json,
  after_json,
  created_at
FROM admin_audit_logs
WHERE (? = '' OR actor_admin_id = ?)
  AND (? = '' OR action = ?)
  AND (? = '' OR resource_type = ?)
  AND (? = '' OR resource_id = ?)
  AND (? IS NULL OR created_at >= ?)
  AND (? IS NULL OR created_at <= ?)
ORDER BY id DESC
LIMIT ?;
```

### Pagination rule

```text
Default page_size: 50
Max page_size: 100
Sort: newest first
Cursor: internal numeric id or encoded cursor
```

> 🟢 **Beginner note:** Audit logs grow fast, isliye unbounded query avoid karo. Always pagination use karo.

---

## Step 11: Usecase Layer Banaya

Usecase layer permission enforce karegi aur repository call karegi.

### Suggested file

```text
backend/services/superadmin-service/internal/usecase/audit_log.go
```

### Code example

```go
type AuditLogService struct {
    authorizer PermissionAuthorizer
    repo       AuditLogRepository
}

func NewAuditLogService(authorizer PermissionAuthorizer, repo AuditLogRepository) (*AuditLogService, error) {
    if authorizer == nil {
        return nil, domain.NewValidationError("authorizer is required")
    }
    if repo == nil {
        return nil, domain.NewValidationError("audit log repository is required")
    }
    return &AuditLogService{authorizer: authorizer, repo: repo}, nil
}

func (s *AuditLogService) ListAuditLogs(
    ctx context.Context,
    req domain.AuditLogListRequest,
) (domain.AuditLogListResponse, error) {
    actor, ok := domain.ActorFromContext(ctx)
    if !ok {
        return domain.AuditLogListResponse{}, domain.NewAdminContextMissing("admin actor is missing")
    }

    if err := s.authorizer.RequirePermission(ctx, actor, domain.PermissionAuditLogsRead); err != nil {
        return domain.AuditLogListResponse{}, err
    }

    normalized, err := req.Normalize()
    if err != nil {
        return domain.AuditLogListResponse{}, err
    }
    return s.repo.ListAuditLogs(ctx, normalized)
}
```

### Explanation

- Actor context middleware se aata hai.
- `audit:logs:read` permission mandatory hai.
- Request normalize karke page size and filters validate hote hain.
- Repository actual MySQL query run karti hai.

---

## Step 12: MySQL Audit Recorder Implement Kiya

Existing code me `LoggingAuditRecorder` temporary hai. Task 8 ka main implementation MySQL-backed recorder hoga.

### Existing flow

```text
ControlService -> AuditRecorder interface -> LoggingAuditRecorder -> application log
```

### Task 8 target flow

```text
ControlService -> AuditRecorder interface -> MySQLAuditRecorder -> admin_audit_logs
```

### Code example

```go
type MySQLAuditRecorder struct {
    repo AuditLogRepository
}

func NewMySQLAuditRecorder(repo AuditLogRepository) *MySQLAuditRecorder {
    return &MySQLAuditRecorder{repo: repo}
}

func (r *MySQLAuditRecorder) RecordAdminMutation(ctx context.Context, record domain.AuditRecord) error {
    return r.repo.InsertAuditLog(ctx, record)
}
```

### Wiring example

```go
auditRepo := repository.NewMySQLAuditLogRepository(db)
auditRecorder := usecase.NewMySQLAuditRecorder(auditRepo)

controls, err := usecase.NewControlService(
    authz,
    userClient,
    reviewTasks,
    auditRecorder,
    logger,
)
```

> 🟢 **Good news:** Existing user/seller and order/payment usecases already depend on `AuditRecorder` interface, so Task 8 can replace recorder implementation without rewriting business logic.

---

## Step 13: Transaction Strategy Define Ki

`docs/04-microservice-design.md` clearly bolta hai:

> Every admin mutation writes audit log in same transaction where possible.

### Case A: Superadmin DB mutation

Example: platform setting update, review task update.

```text
BEGIN
  read old row
  update/upsert business row
  insert admin_audit_logs row
COMMIT
```

### Case B: Downstream service mutation

Example: User Service ko seller suspend request.

```text
1. Superadmin validates actor + permission + reason.
2. Superadmin calls downstream service with admin context.
3. Downstream succeeds.
4. Superadmin inserts audit log immediately.
5. Superadmin returns success only after audit insert succeeds.
```

### Why not same transaction for downstream?

User Service, Payment Service, aur Superadmin Service ke DB alag hote hain. Cross-service DB transaction avoid karna chahiye. Isliye audit log request id, resource id, aur downstream context ke saath durable insert hota hai.

### Failure policy

| Situation | Recommended behavior |
|---|---|
| Local DB mutation + audit insert fails | Rollback transaction |
| Downstream call fails | Audit `failed`/application log, return mapped error |
| Downstream succeeds but audit insert fails | Return error and alert; retry/repair queue future hardening |
| Audit list query fails | Return `500` with request id |

> 🔴 **High-risk rule:** Refund approval, seller suspension, user block, and platform setting update should not return success unless audit write is successful.

---

## Step 14: HTTP Handler Add Kiya

### Suggested file

```text
backend/services/superadmin-service/internal/transport/http/audit_log_handler.go
```

### Route

```go
func (h *AuditLogHandler) Register(mux *http.ServeMux) {
    mux.Handle("/api/v1/admin/audit-logs", ActorMiddleware(http.HandlerFunc(h.listAuditLogs)))
}
```

### Handler example

```go
func (h *AuditLogHandler) listAuditLogs(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/api/v1/admin/audit-logs" {
        writeError(w, r, domain.NewValidationError("expected /api/v1/admin/audit-logs"))
        return
    }
    if r.Method != http.MethodGet {
        w.Header().Set("Allow", http.MethodGet)
        writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
            Code: "METHOD_NOT_ALLOWED",
            Message: "method not allowed",
        })
        return
    }

    req, err := auditLogListRequestFromQuery(r)
    if err != nil {
        writeError(w, r, err)
        return
    }

    response, err := h.auditLogs.ListAuditLogs(r.Context(), req)
    if err != nil {
        writeError(w, r, err)
        return
    }
    writeJSON(w, http.StatusOK, response)
}
```

### Query parsing example

```go
func auditLogListRequestFromQuery(r *http.Request) (domain.AuditLogListRequest, error) {
    q := r.URL.Query()
    pageSize, err := positiveIntQuery(q.Get("page_size"), "page_size")
    if err != nil {
        return domain.AuditLogListRequest{}, err
    }

    req := domain.AuditLogListRequest{
        ActorAdminID: q.Get("actor_id"),
        Action:       q.Get("action"),
        ResourceType: q.Get("resource_type"),
        ResourceID:   q.Get("resource_id"),
        Pagination: domain.Pagination{
            PageSize: pageSize,
            Cursor:   q.Get("cursor"),
        },
    }
    return req.Normalize()
}
```

---

## Step 15: RBAC Rules Attach Kiye

Task 3 me permissions already defined hain.

| Permission | Risk | Who gets it |
|---|---|---|
| `audit:logs:read` | High | `superadmin`, `readonly_admin` |
| `audit:logs:export` | Critical | `superadmin` only |

### Why `readonly_admin` can read audit logs?

Readonly/audit/support team ko investigation ke liye logs dekhne hote hain, but mutation permission nahi milni chahiye.

### Why export only `superadmin`?

Export me bulk sensitive operational data nikal sakta hai. Isliye reason, request id, and strict permission required honi chahiye.

### High-risk export policy

Existing `rbac.PolicyForPermission` me export already strict hai:

```go
case domain.PermissionAuditLogsExport:
    return HighRiskPolicy{
        Permission:       permission,
        RequireReason:    true,
        RequireRequestID: true,
    }
```

> 🟡 **Note:** Current Task 8 me list API define hai. Export API future me aaye to same audit system us export action ko bhi log karega.

---

## Step 16: Integration Points Update Kiye

Previous Superadmin tasks ne audit-ready payloads already prepare kiye hain. Task 8 me unhe persistent writer milta hai.

| Source Flow | Current Direction | Task 8 Final Write |
|---|---|---|
| User block/unblock | `AuditRecorder.RecordAdminMutation` | `admin_audit_logs` insert |
| Seller approve/reject/suspend | `AuditRecorder.RecordAdminMutation` | `admin_audit_logs` insert |
| Refund approve/reject | `AuditRecorder.RecordAdminMutation` | `admin_audit_logs` insert |
| Manual order review | `AuditRecorder.RecordAdminMutation` | `admin_audit_logs` insert |
| Session risk access | Access log now | Audit row for risky access/export |
| Platform setting update | Audit-ready in Task 7 | Same transaction where possible |
| Audit export | Permission exists | Future export action audit row |

### Example mutation audit call

```go
err := auditRecorder.RecordAdminMutation(ctx, domain.AuditRecord{
    ActorAdminID: actor.AdminID,
    Action:       "refund.review.approved",
    ResourceType: "refund",
    ResourceID:   refundID,
    RequestID:    actor.RequestID,
    IPHash:       actor.IPHash,
    Reason:       req.Reason,
    Before: map[string]string{
        "status": "pending_review",
    },
    After: map[string]string{
        "status": "approved",
    },
})
if err != nil {
    return domain.RefundSnapshot{}, err
}
```

---

## Step 17: Architecture Diagram

```mermaid
flowchart TD
    Panel[Superadmin Panel] --> Gateway[API Gateway]
    Gateway -->|Admin JWT + headers| SA[Superadmin Service]

    SA --> AuthZ[Service-level RBAC]
    AuthZ --> Usecase[Admin Usecase]
    Usecase --> Downstream[User / Order / Payment / Session / Search / CMS]
    Usecase --> Recorder[AuditRecorder Interface]
    Recorder --> Repo[MySQL Audit Repository]
    Repo --> AuditDB[(admin_audit_logs)]

    SA --> ListAPI[ListAuditLogs Usecase]
    ListAPI --> Repo

    classDef app fill:#e8f1ff,stroke:#1d4ed8,color:#0f172a;
    classDef db fill:#fff7ed,stroke:#ea580c,color:#0f172a;
    classDef guard fill:#fef2f2,stroke:#dc2626,color:#0f172a;

    class Panel,Gateway,SA,Usecase,Downstream,Recorder,ListAPI app;
    class Repo,AuditDB db;
    class AuthZ guard;
```

---

## Step 18: Mutation Flow Diagram

```mermaid
sequenceDiagram
    participant Admin as Admin User
    participant GW as API Gateway
    participant SA as Superadmin Service
    participant AuthZ as RBAC
    participant DS as Downstream Service
    participant Audit as admin_audit_logs

    Admin->>GW: PATCH /api/v1/admin/sellers/{id}/status
    GW->>SA: Forward admin headers + request id
    SA->>AuthZ: Require sellers:status:update
    AuthZ-->>SA: Allowed
    SA->>SA: Validate reason + status transition
    SA->>DS: Update seller status with admin context
    DS-->>SA: Updated seller snapshot
    SA->>Audit: INSERT immutable audit row
    Audit-->>SA: Insert OK
    SA-->>GW: 200 OK
    GW-->>Admin: Success
```

---

## Step 19: Audit Read Flow Diagram

```mermaid
sequenceDiagram
    participant Admin as Readonly/Superadmin
    participant GW as API Gateway
    participant SA as Superadmin Service
    participant AuthZ as RBAC
    participant DB as MySQL admin_audit_logs

    Admin->>GW: GET /api/v1/admin/audit-logs?resource_type=seller
    GW->>SA: Forward admin context
    SA->>AuthZ: Require audit:logs:read
    AuthZ-->>SA: Allowed
    SA->>DB: SELECT filtered audit logs
    DB-->>SA: Logs + next cursor
    SA-->>GW: AuditLogListResponse
    GW-->>Admin: 200 OK
```

---

## Step 20: ER Diagram

```mermaid
erDiagram
    ADMIN_USERS ||--o{ ADMIN_AUDIT_LOGS : performs

    ADMIN_USERS {
        bigint id PK
        string admin_id UK
        string user_id UK
        string role
        string status
    }

    ADMIN_AUDIT_LOGS {
        bigint id PK
        string audit_id UK
        string actor_admin_id
        string action
        string resource_type
        string resource_id
        string request_id
        string session_id
        string ip_hash
        string reason
        json before_json
        json after_json
        json metadata_json
        timestamp created_at
    }
```

---

## Step 21: Folder Structure of Real Implementation

Task 8 ka actual backend implementation future me kuch is tarah organized hoga:

```text
backend/
└── services/
    └── superadmin-service/
        ├── cmd/
        │   └── server/
        │       └── main.go
        ├── internal/
        │   ├── domain/
        │   │   ├── admin_actor.go
        │   │   ├── admin_audit.go
        │   │   ├── errors.go
        │   │   └── permission.go
        │   ├── repository/
        │   │   ├── mysql_audit_log_repository.go
        │   │   ├── mysql_permission_repository.go
        │   │   └── mysql_review_task_repository.go
        │   ├── usecase/
        │   │   ├── audit_log.go
        │   │   ├── audit_recorder.go
        │   │   ├── order_payment_controls.go
        │   │   └── user_seller_controls.go
        │   ├── rbac/
        │   │   ├── guards.go
        │   │   └── matrix.go
        │   └── transport/
        │       └── http/
        │           ├── audit_log_handler.go
        │           ├── middleware.go
        │           ├── respond.go
        │           └── routes.go
        └── migrations/
            ├── 005_create_admin_audit_logs.up.sql
            └── 005_create_admin_audit_logs.down.sql
```

### File purpose

| File | Purpose |
|---|---|
| `domain/admin_audit.go` | Audit write/read models and validation |
| `repository/mysql_audit_log_repository.go` | Insert/list audit logs in MySQL |
| `usecase/audit_recorder.go` | MySQL-backed recorder implements existing interface |
| `usecase/audit_log.go` | List audit logs with RBAC |
| `transport/http/audit_log_handler.go` | REST handler for `GET /api/v1/admin/audit-logs` |
| `migrations/005_create_admin_audit_logs.up.sql` | Creates immutable audit table |
| `cmd/server/main.go` | Wires repository, recorder, usecase, and handler |

> 🟢 **Current task output:** Sirf `TaskImplementation/Superadmin Service/task8.md` create kiya gaya. Upar wala backend structure implementation blueprint hai.

---

## Step 22: External Libraries and Tools

Task 8 documentation file create karne ke liye koi external library install nahi ki gayi. Actual backend implementation me ye tools useful honge:

| Library/Tool | Type | Why used | Install/Use |
|---|---|---|---|
| MySQL 8.x | Database | Durable structured audit logs, indexes, transactions | Docker Compose/local MySQL |
| `github.com/go-sql-driver/mysql` | Go library | Go service ko MySQL se connect karne ke liye | Already present in `backend/services/superadmin-service/go.mod`; otherwise `go get github.com/go-sql-driver/mysql` |
| `golang-migrate/migrate` | CLI tool | SQL migrations apply/rollback karne ke liye | `go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest` |
| Go standard library `database/sql` | Built-in | SQL query execution | No install |
| Go standard library `encoding/json` | Built-in | Before/after snapshots JSON me convert karne ke liye | No install |
| Go standard library `crypto/rand` | Built-in | Secure audit id generate karne ke liye | No install |

### Migration run example

```bash
migrate \
  -path backend/services/superadmin-service/migrations \
  -database "mysql://root:localroot@tcp(localhost:3306)/superadmin_db" \
  up
```

### Go test example

```bash
cd backend/services/superadmin-service
go test ./...
```

---

## Step 23: Security Rules

| Rule | Why |
|---|---|
| Audit writes are append-only | Tampering risk kam hota hai |
| High-risk actions require reason | Accountability clear hoti hai |
| Request id required for mutations | Distributed trace connect hota hai |
| IP hash only, raw IP avoid | Privacy better hoti hai |
| No secrets in snapshots | Compliance and breach impact control |
| Audit read requires permission | Audit data sensitive hai |
| Export requires stricter permission | Bulk data exfiltration risk hota hai |
| Page size capped | Large scans and accidental leakage avoid hota hai |

### Sensitive data blocklist

```text
password
otp
access_token
refresh_token
authorization
cookie
card_number
cvv
secret
api_key
```

---

## Step 24: Testing Strategy

### Unit tests

| Test | Expected result |
|---|---|
| `AuditRecord.Validate` missing actor | Validation error |
| Missing action | Validation error |
| Missing resource | Validation error |
| Snapshot sanitizer removes secrets | Secret keys absent |
| Audit id generated | Starts with `aud_` and unique enough |
| List request default page size | Defaults to 50 |
| List request page size over max | Validation error or capped to 100 |

### Repository tests

| Test | Expected result |
|---|---|
| Insert valid audit record | Row created |
| Insert duplicate `audit_id` | Unique constraint error |
| List by actor | Only actor logs returned |
| List by resource | Only target resource logs returned |
| List by action | Only matching action logs returned |
| Newest first ordering | Latest rows first |

### Usecase tests

| Test | Expected result |
|---|---|
| `superadmin` lists audit logs | Allowed |
| `readonly_admin` lists audit logs | Allowed |
| `operations_admin` lists audit logs | Forbidden |
| Missing admin context | Unauthorized/admin context error |
| Repository failure | Error returned with request id |

### Mutation integration tests

| Flow | Expected audit row |
|---|---|
| User block | `user.status.blocked` |
| Seller suspend | `seller.status.suspended` |
| Refund approve | `refund.review.approved` |
| Manual order review | `order.manual_review.create` |
| Platform setting update | `platform.setting.updated` |

---

## Step 25: Manual Verification Checklist

```text
PASS: TaskImplementation/Superadmin Service/task8.md exists.
PASS: Existing Superadmin Service folder preserved.
PASS: No unrelated backend/frontend code generated in this task.
PASS: Task 8 scope is admin audit logs only.
PASS: Audit table schema documented.
PASS: Append-only immutability rules documented.
PASS: Audit write flow documented.
PASS: Audit list API documented.
PASS: RBAC permissions documented.
PASS: PII minimization documented.
PASS: External tools and install/use commands documented.
PASS: Mermaid architecture and sequence diagrams included.
```

---

## ✅ Completion Criteria

| Requirement | Status | Notes |
|---|:---:|---|
| Create `TaskImplementation/` | ✅ | Already existed |
| Keep/create `TaskImplementation/Superadmin Service/` | ✅ | Existing folder kept |
| Create `task8.md` | ✅ | This file |
| Step-by-step implementation in Hinglish | ✅ | Steps 1-25 |
| Explain how each part was built | ✅ | Schema, domain, repo, usecase, handler, RBAC |
| Mention external tools/libraries | ✅ | MySQL, driver, migrate, stdlib |
| Clean folder structure | ✅ | Actual output + future backend structure |
| Code examples | ✅ | SQL, Go domain, repository, usecase, handler |
| Diagrams | ✅ | Architecture, mutation flow, read flow, ERD |
| Beginner-friendly and visually readable | ✅ | Badges, tables, notes, diagrams |
| Do not implement beyond Task 8 | ✅ | Documentation-only guide |

---

## 🧾 Final Summary

Superadmin Service Task 8 ka immutable audit-log model ab clearly documented hai:

- `admin_audit_logs` MySQL table append-only source of truth hogi.
- Existing `AuditRecorder` interface ko MySQL-backed recorder se replace kiya jayega.
- Har high-risk admin mutation audit insert ke bina successful treat nahi hogi.
- `GET /api/v1/admin/audit-logs` RBAC-protected read API hoga.
- `audit:logs:read` permission `superadmin` and `readonly_admin` ko milegi.
- `audit:logs:export` future export ke liye `superadmin` only rahega.
- PII/secrets audit snapshots me store nahi honge.
- Folder structure, code examples, diagrams, tools, and test plan define ho gaye.

> 🟢 **Task 8 complete:** Superadmin Service ke admin actions ke liye immutable audit storage, query, security, and verification approach ready hai.
