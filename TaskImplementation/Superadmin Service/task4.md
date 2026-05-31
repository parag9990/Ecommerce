# 🛡️ Superadmin Service - Task 4: User/Seller Controls

![Task](https://img.shields.io/badge/Task-Superadmin%20Service%20%234-blue)
![Priority](https://img.shields.io/badge/Priority-P1-orange)
![Dependency](https://img.shields.io/badge/Dependency-User%20Service-yellow)
![Status](https://img.shields.io/badge/Status-Documented-success)
![Security](https://img.shields.io/badge/Security-RBAC%20%2B%20Audit-critical)
![Scope](https://img.shields.io/badge/Scope-User%2FSeller%20Controls-purple)
![Guide](https://img.shields.io/badge/Guide-Hinglish-ff69b4)

## 📌 Task Summary

| Field | Detail |
|---|---|
| Task Name | User/seller controls |
| Source | `docs/01-micro-tasks.md` -> `Superadmin Service` -> Task 4 |
| Goal | Block user, approve seller, suspend seller, aur verify KYC flows banana |
| Dependency | User Service |
| Priority | P1 |
| Output Type | Documentation-only implementation guide |
| Not Included | Order/payment controls, refund review, session visibility, platform settings, final audit-log service implementation, frontend panel |

> **Simple Hinglish goal:** Is task ka kaam Superadmin Service ke through users aur sellers ko safely control karna hai. Admin user ko block/unblock kar sakega, seller ko approve/reject/suspend kar sakega, aur seller KYC review flow ko User Service ke saath coordinate karega. Actual source code files is task me create nahi kiye gaye; ye beginner-friendly implementation guide hai.

---

## ✅ Final Output Created

```text
TaskImplementation/
└── Superadmin Service/
    ├── task1.md
    ├── task2.md
    ├── task3.md
    └── task4.md
```

### Why this structure?

| Folder/File | Purpose |
|---|---|
| `TaskImplementation/` | Saare task-wise implementation guides ka central location |
| `Superadmin Service/` | Superadmin Service ke tasks ko logically group karta hai |
| `task1.md` | Admin domain boundaries define karta hai |
| `task2.md` | Superadmin Service ke liye MySQL decision explain karta hai |
| `task3.md` | Admin RBAC roles and permissions define karta hai |
| `task4.md` | Sirf **Superadmin Service - Task 4** ka user/seller controls guide |

> 🟢 **Important:** `Superadmin Service` folder already present tha, isliye usko keep kiya gaya. Is task me backend source code, proto files, DB migrations, ya frontend pages create nahi kiye gaye.

---

## 📚 Documents Studied

| Document | Kya use kiya gaya |
|---|---|
| `docs/01-micro-tasks.md` | Task 4 ka exact scope: block user, approve/suspend seller, verify KYC flows |
| `TaskImplementation/Superadmin Service/task1.md` | Admin domain, controlled resources, high-risk action rules |
| `TaskImplementation/Superadmin Service/task2.md` | MySQL ownership, review task and audit-log direction |
| `TaskImplementation/Superadmin Service/task3.md` | Admin RBAC, permissions, admin context, high-risk guard |
| `docs/04-microservice-design.md` | Superadmin Service APIs and User Service ownership |
| `docs/05-database-design.md` | User/Seller and Superadmin DB table direction |
| `docs/06-auth-security.md` | RBAC, admin mutation rate limit, audit requirements |
| `docs/09-cms-superadmin.md` | Superadmin modules, seller approval workflow, high-risk controls |
| `api/master-api.json` | Admin REST routes, gRPC mapping, request/response schema names |
| `database/draw.sql` | Reference statuses for `users`, `seller_profiles`, `seller_kyc_documents`, `admin_review_tasks` |

---

## 🧭 Implementation Approach

Task 4 ko **workflow implementation guide** treat kiya gaya. Isme Superadmin Service khud user/seller data ka owner nahi banega. User profile, seller profile, aur KYC metadata ka source of truth **User Service** rahega.

### Final decision

```text
Admin REST route: API Gateway
Workflow owner: Superadmin Service
User/seller source of truth: User Service
Security: Task 3 RBAC permissions
Audit direction: admin mutation audit record
Storage helper: admin_review_tasks for approval/review workflow
```

### One-line reason

User/seller controls high-risk hain, isliye direct DB update ke bajay Superadmin Service permission check karegi, reason require karegi, audit context attach karegi, aur User Service ko gRPC se status update request bhejegi.

---

## 🪜 Step-by-Step Implementation

## Step 1: Task Boundary Clear Kiya

`docs/01-micro-tasks.md` me Superadmin Service Task 4 ye hai:

| S.No | Task Name | Detail | Dependency | Priority |
|---:|---|---|---|---|
| 4 | User/seller controls | Block user, approve seller, suspend seller, verify KYC flows banao. | User Service | P1 |

### Is task me allowed work

- Admin ke liye user list/search flow define karna
- User block/unblock flow define karna
- Seller list/search/KYC review flow define karna
- Seller approve/reject/suspend flow define karna
- User Service ke saath gRPC coordination explain karna
- RBAC permissions apply karna
- Reason, validation, status transition, audit context explain karna
- Code examples dena for handlers/usecases/client contracts

### Is task me not allowed work

| Area | Reason |
|---|---|
| Refund approve/reject flow | Ye **Task 5: Order/payment controls** ka scope hai |
| Session analytics visibility | Ye **Task 6: Session visibility** ka scope hai |
| Platform settings | Ye **Task 7: Platform settings** ka scope hai |
| Full immutable audit-log production implementation | Ye **Task 8: Admin audit logs** ka scope hai |
| Superadmin frontend pages | Ye frontend/superadmin panel ka separate work hai |
| Direct User Service DB writes | Microservice ownership rule ke against hai |

> 🔴 **Boundary rule:** Superadmin Service User Service ke database me direct update nahi karegi. Status change hamesha User Service gRPC API ke through hoga.

---

## Step 2: Existing API Contract Identify Kiya

`api/master-api.json` aur microservice docs ke according Task 4 ke main routes ye hain:

| REST API | gRPC Method | Auth | Purpose |
|---|---|---|---|
| `GET /api/v1/admin/users` | `SuperadminService.ListUsersForAdmin` | `admin` | Admin user search/list |
| `PATCH /api/v1/admin/users/{user_id}/status` | `SuperadminService.UpdateUserStatus` | `admin` | User block/unblock |
| `GET /api/v1/admin/sellers` | `SuperadminService.ListSellersForAdmin` | `admin` | Admin seller search/list |
| `PATCH /api/v1/admin/sellers/{seller_id}/status` | `SuperadminService.UpdateSellerStatus` | `admin` | Seller approve/reject/suspend |

### Request/response schemas

```json
{
  "AdminUserListRequest": {
    "q": "search text",
    "status": "active|blocked|deleted",
    "page": 1,
    "page_size": 20
  },
  "AdminSellerListRequest": {
    "q": "search text",
    "status": "draft|pending_review|active|suspended|rejected",
    "page": 1,
    "page_size": 20
  },
  "StatusUpdateRequest": {
    "status": "new_status",
    "reason": "mandatory admin reason"
  }
}
```

### Explanation

- `GET` APIs admin panel ko searchable lists denge.
- `PATCH status` APIs mutation APIs hain, isliye reason mandatory hai.
- Gateway route-level `admin` auth check karega.
- Superadmin Service exact permission check karegi.

---

## Step 3: Permissions Map Kiye

Task 3 RBAC ke permission keys reuse honge.

| Action | Required Permission | Allowed Roles |
|---|---|---|
| List/search users | `users:read` | `superadmin`, `operations_admin`, `readonly_admin` |
| Block/unblock user | `users:status:update` | `superadmin`, `operations_admin` |
| List/search sellers | `sellers:read` | `superadmin`, `operations_admin`, `catalog_admin`, `readonly_admin` |
| Approve/reject seller | `sellers:status:update` | `superadmin`, `operations_admin` |
| Suspend seller | `sellers:status:update` | `superadmin`, `operations_admin` |
| View KYC metadata | `sellers:read` | `superadmin`, `operations_admin`, `catalog_admin`, `readonly_admin` |
| Review KYC decision | `sellers:status:update` | `superadmin`, `operations_admin` |

### Beginner note

`admin` auth sirf entry gate hai. Real permission check service ke andar hota hai. Example: `readonly_admin` admin route access kar sakta hai, but user block nahi kar sakta.

---

## Step 4: Status Models Final Kiye

Reference `database/draw.sql` ke status values:

### User statuses

| Status | Meaning | Admin action |
|---|---|---|
| `active` | User normal account use kar sakta hai | Unblock/restore active |
| `blocked` | User restricted hai | Block suspicious/policy-violating account |
| `deleted` | Deleted/closed account | Admin status API se direct restore nahi karna |

### Seller statuses

| Status | Meaning | Admin action |
|---|---|---|
| `draft` | Seller profile incomplete hai | Usually no admin action |
| `pending_review` | Seller approval/KYC review pending hai | Approve or reject |
| `active` | Seller sell kar sakta hai | Approve/un-suspend |
| `suspended` | Seller temporarily restricted hai | Suspend for policy/risk |
| `rejected` | Seller application reject ho gayi | Reject KYC/application |

### KYC document statuses

| Status | Meaning |
|---|---|
| `pending` | Document review pending |
| `approved` | Document valid mark hua |
| `rejected` | Document invalid/insufficient mark hua |

---

## Step 5: Valid Status Transitions Define Kiye

Status transition rules invalid admin actions ko block karte hain.

### User transition table

| Current | Allowed New Status | Notes |
|---|---|---|
| `active` | `blocked` | Reason mandatory |
| `blocked` | `active` | Unblock reason mandatory |
| `deleted` | none | Deleted user ko normal status API se mutate nahi karna |

### Seller transition table

| Current | Allowed New Status | Notes |
|---|---|---|
| `pending_review` | `active` | KYC approve / seller approve |
| `pending_review` | `rejected` | Rejection reason mandatory |
| `active` | `suspended` | Suspension reason mandatory |
| `suspended` | `active` | Reinstatement reason mandatory |
| `draft` | none | Seller ne submit nahi kiya |
| `rejected` | `pending_review` | Re-submit flow User Service/CMS trigger karega |

### Status validator example

```go
package domain

func CanUpdateUserStatus(current string, next string) bool {
    allowed := map[string][]string{
        "active":  {"blocked"},
        "blocked": {"active"},
    }
    return contains(allowed[current], next)
}

func CanUpdateSellerStatus(current string, next string) bool {
    allowed := map[string][]string{
        "pending_review": {"active", "rejected"},
        "active":         {"suspended"},
        "suspended":      {"active"},
    }
    return contains(allowed[current], next)
}
```

**Explanation:**  
Validator simple guard hai. Agar admin invalid transition kare, jaise `deleted -> active`, service `400 INVALID_STATUS_TRANSITION` return karegi.

---

## Step 6: User List Flow Banaya

User list API admin panel ko users search karne degi.

```mermaid
sequenceDiagram
    participant Admin as Admin User
    participant Panel as Superadmin Panel
    participant GW as API Gateway
    participant SA as Superadmin Service
    participant UserSvc as User Service

    Admin->>Panel: Search users
    Panel->>GW: GET /api/v1/admin/users?q=rahul&status=active
    GW->>GW: Validate JWT + admin route
    GW->>SA: ListUsersForAdmin(request + admin context)
    SA->>SA: RequirePermission(users:read)
    SA->>UserSvc: ListUsersForAdmin(q, status, pagination)
    UserSvc-->>SA: User list
    SA-->>GW: AdminUserListResponse
    GW-->>Panel: 200 OK
```

### Usecase example

```go
func (s *SuperadminService) ListUsersForAdmin(
    ctx context.Context,
    req *ListUsersForAdminRequest,
) (*AdminUserListResponse, error) {
    actor := ActorFromContext(ctx)

    if err := s.authorizer.RequirePermission(ctx, actor, PermissionUsersRead); err != nil {
        return nil, err
    }

    cleanReq := NormalizeUserListRequest(req)
    return s.userClient.ListUsersForAdmin(ctx, cleanReq)
}
```

### Explanation

- Superadmin Service list data khud DB se nahi uthata.
- User Service user data owner hai, isliye list query User Service execute karega.
- Superadmin Service permission gate ki tarah kaam karega.

---

## Step 7: User Block/Unblock Flow Banaya

User block/unblock high-risk mutation hai.

### Required checks

| Check | Why |
|---|---|
| Admin actor present | Anonymous mutation block karna |
| `users:status:update` permission | Sirf allowed admins status change kar sake |
| Reason non-empty | Support/compliance trace ke liye |
| Status transition valid | Invalid business state avoid karna |
| Request ID present | Debugging and audit traceability |
| Admin context downstream attach | User Service ko pata ho action kisne kiya |

### Flow diagram

```mermaid
sequenceDiagram
    participant Admin
    participant GW as API Gateway
    participant SA as Superadmin Service
    participant UserSvc as User Service
    participant DB as User DB
    participant Audit as Admin Audit Direction

    Admin->>GW: PATCH /api/v1/admin/users/{user_id}/status
    GW->>SA: UpdateUserStatus(status, reason, actor)
    SA->>SA: RequirePermission(users:status:update)
    SA->>SA: Validate reason + next status
    SA->>UserSvc: GetUser(user_id)
    UserSvc-->>SA: Current user status
    SA->>SA: Validate transition
    SA->>UserSvc: UpdateUserStatus(user_id, status, admin metadata)
    UserSvc->>DB: Update users.status
    UserSvc-->>SA: Success
    SA-->>Audit: Write/emit audit event direction
    SA-->>GW: SuccessResponse
```

### Request example

```http
PATCH /api/v1/admin/users/user_123/status
Content-Type: application/json
Authorization: Bearer <admin_access_token>

{
  "status": "blocked",
  "reason": "Multiple payment abuse complaints verified by support"
}
```

### Usecase example

```go
func (s *SuperadminService) UpdateUserStatus(
    ctx context.Context,
    userID string,
    req *StatusUpdateRequest,
) (*SuccessResponse, error) {
    actor := ActorFromContext(ctx)

    if err := s.authorizer.RequirePermission(ctx, actor, PermissionUsersStatusUpdate); err != nil {
        return nil, err
    }
    if err := ValidateMutationReason(req.Reason); err != nil {
        return nil, err
    }

    user, err := s.userClient.GetUser(ctx, userID)
    if err != nil {
        return nil, err
    }
    if !CanUpdateUserStatus(user.Status, req.Status) {
        return nil, ErrInvalidStatusTransition
    }

    downstreamCtx := AttachAdminContext(ctx, actor, req.Reason)
    if err := s.userClient.UpdateUserStatus(downstreamCtx, userID, req.Status, req.Reason); err != nil {
        return nil, err
    }

    s.audit.RecordAdminMutation(ctx, AuditRecord{
        ActorAdminID: actor.AdminID,
        Action:       "user.status.update",
        ResourceType: "user",
        ResourceID:   userID,
        Reason:       req.Reason,
        Before:       map[string]string{"status": user.Status},
        After:        map[string]string{"status": req.Status},
    })

    return &SuccessResponse{Success: true}, nil
}
```

### Explanation

- Pehle permission check hota hai.
- Phir reason validate hota hai.
- Current status User Service se fetch hota hai.
- Valid transition ke baad User Service status update karta hai.
- Audit direction record hota hai. Full immutable audit implementation Task 8 me complete hogi.

---

## Step 8: Seller List/KYC View Flow Banaya

Seller list API admin ko sellers aur KYC review candidates dikhayegi.

```mermaid
flowchart TD
    A[Admin opens sellers page] --> B[Gateway validates admin JWT]
    B --> C[Superadmin Service checks sellers:read]
    C --> D[User Service lists seller profiles]
    D --> E{Status filter?}
    E -->|pending_review| F[Show KYC review queue]
    E -->|active/suspended/rejected| G[Show seller operations list]
    F --> H[Admin opens seller review detail]
    G --> H
```

### Query examples

```http
GET /api/v1/admin/sellers?status=pending_review&page=1&page_size=20
GET /api/v1/admin/sellers?q=store-name&status=active
```

### Usecase example

```go
func (s *SuperadminService) ListSellersForAdmin(
    ctx context.Context,
    req *AdminSellerListRequest,
) (*AdminSellerListResponse, error) {
    actor := ActorFromContext(ctx)

    if err := s.authorizer.RequirePermission(ctx, actor, PermissionSellersRead); err != nil {
        return nil, err
    }

    cleanReq := NormalizeSellerListRequest(req)
    return s.userClient.ListSellersForAdmin(ctx, cleanReq)
}
```

### Explanation

- `pending_review` filter KYC review queue ki tarah use ho sakta hai.
- Seller detail me KYC metadata User Service se aayega.
- Storage URL raw expose karte waqt signed/temporary URL use karna chahiye, public permanent URL nahi.

---

## Step 9: Seller Approve/Reject/Suspend Flow Banaya

Seller status update bhi high-risk mutation hai.

### Status action mapping

| Admin action | Current status | New status | Meaning |
|---|---|---|---|
| Approve seller | `pending_review` | `active` | Seller platform par sell kar sakta hai |
| Reject seller | `pending_review` | `rejected` | KYC/profile unacceptable |
| Suspend seller | `active` | `suspended` | Seller temporarily restricted |
| Reinstate seller | `suspended` | `active` | Seller allowed again |

### Flow diagram

```mermaid
sequenceDiagram
    participant Admin
    participant GW as API Gateway
    participant SA as Superadmin Service
    participant UserSvc as User Service
    participant ReviewDB as Superadmin MySQL admin_review_tasks
    participant Notif as Notification Service Future

    Admin->>GW: PATCH /api/v1/admin/sellers/{seller_id}/status
    GW->>SA: UpdateSellerStatus(status, reason, actor)
    SA->>SA: RequirePermission(sellers:status:update)
    SA->>UserSvc: GetSellerProfile(seller_id)
    UserSvc-->>SA: Current seller + KYC status
    SA->>SA: Validate transition and reason
    SA->>UserSvc: UpdateSellerStatus(admin metadata)
    UserSvc-->>SA: Success
    SA->>ReviewDB: Close matching review task if exists
    SA-->>Notif: Future SellerStatusChanged event
    SA-->>GW: SuccessResponse
```

### Request examples

```http
PATCH /api/v1/admin/sellers/seller_456/status
Content-Type: application/json
Authorization: Bearer <admin_access_token>

{
  "status": "active",
  "reason": "GST and bank verification passed"
}
```

```http
PATCH /api/v1/admin/sellers/seller_456/status
Content-Type: application/json
Authorization: Bearer <admin_access_token>

{
  "status": "suspended",
  "reason": "Repeated counterfeit product complaints under review"
}
```

### Usecase example

```go
func (s *SuperadminService) UpdateSellerStatus(
    ctx context.Context,
    sellerID string,
    req *StatusUpdateRequest,
) (*SuccessResponse, error) {
    actor := ActorFromContext(ctx)

    if err := s.authorizer.RequirePermission(ctx, actor, PermissionSellersStatusUpdate); err != nil {
        return nil, err
    }
    if err := ValidateMutationReason(req.Reason); err != nil {
        return nil, err
    }

    seller, err := s.userClient.GetSellerProfile(ctx, sellerID)
    if err != nil {
        return nil, err
    }
    if !CanUpdateSellerStatus(seller.Status, req.Status) {
        return nil, ErrInvalidStatusTransition
    }

    downstreamCtx := AttachAdminContext(ctx, actor, req.Reason)
    if err := s.userClient.UpdateSellerStatus(downstreamCtx, sellerID, req.Status, req.Reason); err != nil {
        return nil, err
    }

    _ = s.reviewTasks.CloseForResource(ctx, "seller_kyc", "seller", sellerID, req.Status, actor.AdminID, req.Reason)
    _ = s.audit.RecordAdminMutation(ctx, AuditRecord{
        ActorAdminID: actor.AdminID,
        Action:       "seller.status.update",
        ResourceType: "seller",
        ResourceID:   sellerID,
        Reason:       req.Reason,
        Before:       map[string]string{"status": seller.Status},
        After:        map[string]string{"status": req.Status},
    })

    return &SuccessResponse{Success: true}, nil
}
```

### Explanation

- Seller status User Service me update hota hai.
- Superadmin Service admin review task close kar sakti hai because review queue Superadmin domain ka part hai.
- Notification Service ko future event se seller ko inform kiya jayega.

---

## Step 10: KYC Review Flow Define Kiya

KYC documents User Service ke `seller_kyc_documents` table me metadata ke form me rahenge. Actual files object storage/CDN me rahenge.

### KYC review states

```mermaid
stateDiagram-v2
    [*] --> pending
    pending --> approved
    pending --> rejected
    rejected --> pending: Seller resubmits document
    approved --> [*]
```

### Seller approval workflow

```mermaid
stateDiagram-v2
    [*] --> draft
    draft --> pending_review
    pending_review --> active: Admin approves
    pending_review --> rejected: Admin rejects
    active --> suspended: Admin suspends
    suspended --> active: Admin reinstates
```

### Review task record

```sql
INSERT INTO admin_review_tasks (
  task_id,
  task_type,
  resource_type,
  resource_id,
  status,
  created_by,
  metadata
) VALUES (
  'task_kyc_123',
  'seller_kyc',
  'seller',
  'seller_456',
  'open',
  'system',
  JSON_OBJECT('source', 'seller_profile_submitted')
);
```

### Explanation

- Seller profile submit hone ke baad review task open ho sakta hai.
- Admin seller detail/KYC metadata dekhkar approve/reject karega.
- Approve means seller profile `active`.
- Reject means seller profile `rejected`, aur KYC documents rejected reason ke saath mark honge.

> 🟡 **Note:** Exact KYC document-level update method User Service contract me add/extend karna future implementation detail hai. Task 4 ka main admin route seller status update contract already defined karta hai.

---

## Step 11: Admin Context Downstream Pass Kiya

Superadmin Service jab User Service ko call karegi, admin metadata attach karegi.

```go
func AttachAdminContext(ctx context.Context, actor AdminActor, reason string) context.Context {
    return metadata.AppendToOutgoingContext(
        ctx,
        "x-admin-id", actor.AdminID,
        "x-admin-roles", strings.Join(RolesToStrings(actor.Roles), ","),
        "x-request-id", actor.RequestID,
        "x-session-id", actor.SessionID,
        "x-admin-reason", reason,
    )
}
```

### Why needed?

| Metadata | Purpose |
|---|---|
| `x-admin-id` | User Service ko actor identity mile |
| `x-admin-roles` | Downstream service optional authorization/logging kar sake |
| `x-request-id` | Logs trace ho sake |
| `x-session-id` | Admin session traceability |
| `x-admin-reason` | Status history/audit me reason capture ho |

---

## Step 12: Error Handling Define Kiya

| Error Code | HTTP | Meaning |
|---|---:|---|
| `ADMIN_CONTEXT_MISSING` | 401 | Admin actor context missing |
| `PERMISSION_DENIED` | 403 | Admin ke paas required permission nahi hai |
| `REASON_REQUIRED` | 400 | Mutation reason missing/too short |
| `INVALID_STATUS` | 400 | Requested status unknown hai |
| `INVALID_STATUS_TRANSITION` | 400 | Current -> next status allowed nahi |
| `USER_NOT_FOUND` | 404 | User Service me user nahi mila |
| `SELLER_NOT_FOUND` | 404 | User Service me seller nahi mila |
| `DOWNSTREAM_UNAVAILABLE` | 503 | User Service temporarily unavailable |

### Error response example

```json
{
  "error": {
    "code": "INVALID_STATUS_TRANSITION",
    "message": "seller status cannot change from draft to active through admin status API",
    "request_id": "req_abc123"
  }
}
```

---

## Step 13: Audit Direction Add Kiya

Task 8 me full immutable audit logs implement honge, but Task 4 ke workflows audit-ready hone chahiye.

### Audit records required

| Action | Resource Type | Action Key |
|---|---|---|
| User blocked | `user` | `user.status.blocked` |
| User unblocked | `user` | `user.status.active` |
| Seller approved | `seller` | `seller.status.active` |
| Seller rejected | `seller` | `seller.status.rejected` |
| Seller suspended | `seller` | `seller.status.suspended` |
| Seller reinstated | `seller` | `seller.status.active` |

### Audit payload example

```json
{
  "actor_admin_id": "admin_789",
  "action": "seller.status.update",
  "resource_type": "seller",
  "resource_id": "seller_456",
  "request_id": "req_abc123",
  "ip_hash": "sha256...",
  "reason": "GST and bank verification passed",
  "before_json": {"status": "pending_review"},
  "after_json": {"status": "active"}
}
```

### Explanation

- Audit payload me actor, action, resource, reason, before/after status hona chahiye.
- Admin mutations me reason compulsory hai.
- Production immutable audit writer Task 8 me finalize hoga, but Task 4 usecases audit record create/emit karne ke point define karte hain.

---

## Step 14: Folder Structure Plan

Task 4 ke production implementation ke liye expected future folder structure:

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
        │   │   ├── admin_permission.go
        │   │   ├── status_transition.go
        │   │   └── admin_audit.go
        │   ├── usecase/
        │   │   ├── list_users.go
        │   │   ├── update_user_status.go
        │   │   ├── list_sellers.go
        │   │   ├── update_seller_status.go
        │   │   └── review_seller_kyc.go
        │   ├── repository/
        │   │   ├── mysql_permission_repository.go
        │   │   └── mysql_review_task_repository.go
        │   ├── clients/
        │   │   └── user_service_client.go
        │   ├── transport/
        │   │   └── grpc/
        │   │       ├── server.go
        │   │       └── superadmin_handler.go
        │   └── middleware/
        │       └── admin_authorization.go
        ├── migrations/
        │   ├── 001_create_superadmin_tables.up.sql
        │   ├── 001_create_superadmin_tables.down.sql
        │   ├── 002_seed_admin_permissions.up.sql
        │   └── 002_seed_admin_permissions.down.sql
        └── deploy/
            ├── Dockerfile
            └── k8s.yaml

proto/
└── ecommerce/
    └── superadmin/
        └── v1/
            └── superadmin.proto

TaskImplementation/
└── Superadmin Service/
    ├── task1.md
    ├── task2.md
    ├── task3.md
    └── task4.md
```

### Folder explanation

| Folder/File | Purpose |
|---|---|
| `domain/status_transition.go` | User/seller valid status transition rules |
| `usecase/update_user_status.go` | User block/unblock workflow |
| `usecase/update_seller_status.go` | Seller approve/reject/suspend workflow |
| `clients/user_service_client.go` | User Service gRPC client wrapper |
| `repository/mysql_review_task_repository.go` | Seller KYC review task close/update |
| `transport/grpc/superadmin_handler.go` | gRPC methods exposed by Superadmin Service |
| `proto/.../superadmin.proto` | Contract for admin list/status APIs |

> 🟢 **Current task output:** Sirf `TaskImplementation/Superadmin Service/task4.md` create kiya gaya. Upar wala backend folder structure future implementation guide hai.

---

## Step 15: External Libraries/Tools

Task 4 guide me koi new runtime dependency install nahi ki gayi. Future production implementation me ye tools useful honge:

| Tool/Library | What it is | Why used | Install/Use |
|---|---|---|---|
| `google.golang.org/grpc` | Go gRPC framework | Superadmin Service -> User Service internal calls | `go get google.golang.org/grpc` |
| `google.golang.org/grpc/metadata` | gRPC metadata helper | Admin context downstream pass karne ke liye | gRPC package ke saath use hota hai |
| `github.com/go-sql-driver/mysql` | MySQL driver for Go | RBAC/review task/audit tables query karne ke liye | `go get github.com/go-sql-driver/mysql` |
| `golang-migrate/migrate` | SQL migration CLI | Superadmin MySQL migrations run karne ke liye | `go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest` |
| Redis rate limiter | External cache/rate-limit store | Admin mutations per admin limit enforce karne ke liye | Platform Foundation Docker Compose me Redis |

### Example install commands

```bash
go get google.golang.org/grpc
go get github.com/go-sql-driver/mysql
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Example migrate command

```bash
migrate \
  -path backend/services/superadmin-service/migrations \
  -database "mysql://root:localroot@tcp(localhost:3306)/superadmin_db" \
  up
```

> 🟡 **Beginner note:** Is task me commands run nahi kiye gaye. Ye future implementation ke liye installation guide hai.

---

## Step 16: Security Rules Apply Kiye

| Rule | Implementation direction |
|---|---|
| Admin JWT required | Gateway validates `admin` auth |
| Service-level permission | Superadmin Service `RequirePermission` call karegi |
| Reason mandatory | Every status update request me `reason` required |
| MFA for high-risk | Seller suspension/user block ke liye admin MFA recommended |
| Rate limit | Admin mutations: approx `30 per admin per min` |
| No raw PII in logs | Logs me token, full IP, raw document URL avoid |
| Audit required | Every status mutation audit-ready |
| Downstream context | User Service call me admin metadata attach |

### High-risk guard example

```go
func ValidateMutationReason(reason string) error {
    reason = strings.TrimSpace(reason)
    if len(reason) < 10 {
        return ErrReasonRequired
    }
    if len(reason) > 512 {
        return ErrReasonTooLong
    }
    return nil
}
```

---

## Step 17: API Examples

### List users

```http
GET /api/v1/admin/users?q=blocked-user@example.com&status=blocked&page=1&page_size=20
Authorization: Bearer <admin_access_token>
```

### Block user

```http
PATCH /api/v1/admin/users/user_123/status
Authorization: Bearer <admin_access_token>
Content-Type: application/json

{
  "status": "blocked",
  "reason": "Confirmed account takeover risk from support investigation"
}
```

### List sellers pending review

```http
GET /api/v1/admin/sellers?status=pending_review&page=1&page_size=20
Authorization: Bearer <admin_access_token>
```

### Approve seller

```http
PATCH /api/v1/admin/sellers/seller_456/status
Authorization: Bearer <admin_access_token>
Content-Type: application/json

{
  "status": "active",
  "reason": "KYC documents, GST number, and support email verified"
}
```

### Suspend seller

```http
PATCH /api/v1/admin/sellers/seller_456/status
Authorization: Bearer <admin_access_token>
Content-Type: application/json

{
  "status": "suspended",
  "reason": "Policy violation confirmed after counterfeit item complaints"
}
```

---

## Step 18: Testing Plan

### Unit tests

| Test Case | Expected Result |
|---|---|
| Admin with `users:read` lists users | Pass |
| `readonly_admin` blocks user | `403 PERMISSION_DENIED` |
| Empty reason on user block | `400 REASON_REQUIRED` |
| `active -> blocked` user transition | Pass |
| `deleted -> active` user transition | `400 INVALID_STATUS_TRANSITION` |
| `pending_review -> active` seller transition | Pass |
| `draft -> active` seller transition | `400 INVALID_STATUS_TRANSITION` |
| Seller suspend without permission | `403 PERMISSION_DENIED` |

### Integration tests

| Flow | Verify |
|---|---|
| User block through Superadmin Service | User Service receives admin metadata and updates status |
| Seller approval | Seller status becomes `active`, review task closes |
| Seller rejection | Seller status becomes `rejected`, reason stored/audit-ready |
| User Service unavailable | Superadmin returns mapped `503 DOWNSTREAM_UNAVAILABLE` |

### Security tests

| Scenario | Expected |
|---|---|
| Missing admin token | `401` |
| Buyer token on admin route | `403` |
| Admin route allowed but missing exact permission | `403` |
| Oversized reason | `400` |
| Suspicious rapid mutations | Rate limit triggers |

---

## 🧱 Architecture Diagram

```mermaid
flowchart LR
    Panel[Superadmin Panel] --> GW[API Gateway]
    GW --> Auth[JWT + Route Auth]
    Auth --> SA[Superadmin Service]
    SA --> RBAC[Permission Check]
    RBAC --> Review[(MySQL: admin_review_tasks)]
    RBAC --> UserSvc[User Service]
    UserSvc --> UserDB[(MySQL: users + seller_profiles + KYC)]
    SA -. audit-ready .-> Audit[(Admin Audit Logs)]
    SA -. future event .-> Notif[Notification Service]

    classDef service fill:#e8f1ff,stroke:#4479A1,color:#111;
    classDef db fill:#fff4db,stroke:#d18b00,color:#111;
    classDef security fill:#ffe8e8,stroke:#cc3333,color:#111;

    class Panel,GW,SA,UserSvc,Notif service;
    class Review,UserDB,Audit db;
    class Auth,RBAC security;
```

### Explanation

1. Admin panel request Gateway ko bhejta hai.
2. Gateway JWT and route auth validate karta hai.
3. Superadmin Service exact permission check karti hai.
4. User/seller status ka source of truth User Service update karta hai.
5. Superadmin Service review task close/update kar sakti hai.
6. Audit record Task 8 ke immutable audit system ke liye ready hota hai.

---

## 🔄 End-to-End Flow Summary

```mermaid
flowchart TD
    A[Admin action request] --> B{Read or Mutation?}
    B -->|Read| C[Check read permission]
    C --> D[Call User Service list/detail API]
    D --> E[Return admin response]

    B -->|Mutation| F[Check status update permission]
    F --> G[Validate reason]
    G --> H[Fetch current resource status]
    H --> I{Valid transition?}
    I -->|No| J[Return 400]
    I -->|Yes| K[Attach admin context]
    K --> L[Call User Service status update]
    L --> M[Close review task if applicable]
    M --> N[Record/emit audit direction]
    N --> O[Return success]
```

---

## ✅ Completion Checklist

| Requirement | Status | Notes |
|---|---|---|
| `TaskImplementation/` folder exists | ✅ | Already present |
| `TaskImplementation/Superadmin Service/` folder exists | ✅ | Existing folder kept |
| `task4.md` created | ✅ | This file |
| Step-by-step implementation in Hinglish | ✅ | Steps 1-18 |
| Clear explanation of each part | ✅ | Each flow has explanation |
| External libraries/tools mentioned | ✅ | gRPC, MySQL driver, migrate, Redis |
| Install/use examples included | ✅ | `go get`, `go install`, migrate command |
| Clean folder structure included | ✅ | Current output + future backend structure |
| Code examples included | ✅ | Go, SQL, HTTP, JSON |
| Mermaid diagrams included | ✅ | Sequence, flowchart, state diagrams |
| Scope limited to Task 4 | ✅ | No Task 5-8 implementation included |

---

## 🧪 Verification Notes

```text
PASS: Superadmin Service Task 4 documented as user/seller controls guide.
PASS: Existing Superadmin Service folder preserved.
PASS: No backend source code, proto, migration, or frontend implementation created.
PASS: Task 4 scope limited to user/seller controls.
```

---

## 🏁 Final Summary

Superadmin Service Task 4 ka user/seller control model ab clearly documented hai:

- Admin users ko list/search kar sakte hain.
- Authorized admins user block/unblock kar sakte hain.
- Sellers ko list/search/KYC review queue me dekh sakte hain.
- Authorized admins seller approve, reject, suspend, ya reinstate kar sakte hain.
- Status transitions controlled hain.
- User Service source of truth rahega.
- Superadmin Service RBAC, reason validation, admin context, review task closure, aur audit-ready mutation flow handle karegi.

> 🟢 **Task 4 complete:** Ab next Superadmin task me order/payment controls build kiye ja sakte hain, without user/seller workflows ko mix kiye.
