# 🛡️ Superadmin Service - Task 2: Choose MySQL

![Task](https://img.shields.io/badge/Task-Superadmin%20Service%20%232-blue)
![Priority](https://img.shields.io/badge/Priority-P1-orange)
![Dependency](https://img.shields.io/badge/Dependency-Admin%20Domain-yellow)
![Status](https://img.shields.io/badge/Status-Documented-success)
![Database](https://img.shields.io/badge/Database-MySQL%208.x-4479A1)
![Scope](https://img.shields.io/badge/Scope-DB%20Choice%20Only-purple)
![Guide](https://img.shields.io/badge/Guide-Hinglish-ff69b4)

## 📌 Task Summary

| Field | Detail |
|---|---|
| Task Name | Choose MySQL |
| Source | `docs/01-micro-tasks.md` → `Superadmin Service` → Task 2 |
| Goal | Admin actions aur permissions highly structured + auditable hain, isliye Superadmin Service ke liye MySQL choose karna |
| Dependency | Superadmin Service Task 1: Define admin domain |
| Priority | P1 |
| Output Type | Documentation-only database selection and implementation guide |
| Not Included | Actual migration files, repository implementation, Admin RBAC code, gRPC handlers, REST routes, frontend panel |

> **Simple Hinglish goal:** Is task ka kaam Superadmin Service ke liye correct database choose karna hai. Admin permissions, platform settings, review tasks, aur audit logs structured data hain, aur inme strong consistency + audit trail important hai. Isliye MySQL select kiya gaya.

---

## ✅ Final Output Created

```text
TaskImplementation/
└── Superadmin Service/
    ├── task1.md
    └── task2.md
```

### Why this structure?

| Folder/File | Purpose |
|---|---|
| `TaskImplementation/` | Saare task-wise implementation guides ka central location |
| `Superadmin Service/` | Superadmin Service ke tasks ko logically group karta hai |
| `task1.md` | Admin domain definition guide, jo Task 2 ki dependency hai |
| `task2.md` | Sirf **Superadmin Service - Task 2** ka MySQL decision guide |

> 🟢 **Important:** `Superadmin Service` folder already present tha, isliye usko keep kiya gaya. Is task me project ke backend code, migrations, ya service files create nahi kiye gaye.

---

## 📚 Documents Studied

| Document | Kya use kiya gaya |
|---|---|
| `docs/01-micro-tasks.md` | Task 2 ka exact name, dependency, priority, aur scope |
| `TaskImplementation/Superadmin Service/task1.md` | Admin domain boundaries aur controlled resources |
| `docs/04-microservice-design.md` | Superadmin Service purpose, responsibilities, database choice, tables, APIs |
| `docs/05-database-design.md` | Superadmin DB tables, relationships, indexes, scalability notes |
| `database/draw.sql` | Existing reference DDL for `superadmin_db` tables |
| `docs/06-auth-security.md` | RBAC, audit, admin mutation security, rate-limit expectations |
| `docs/09-cms-superadmin.md` | Superadmin modules, high-risk controls, workflows |
| `docs/03-folder-structure.md` | Future `superadmin-service` folder and migrations layout |

---

## 🧭 Implementation Approach

Task 2 ko **database decision task** treat kiya gaya. Pehle Task 1 me admin domain define ho chuka hai; ab us domain ke data shape ko dekhkar database choose kiya gaya.

### Final decision

```text
Superadmin Service primary database: MySQL
Recommended engine: InnoDB
Recommended charset: utf8mb4
Recommended collation: utf8mb4_unicode_ci
```

### One-line reason

Superadmin data mostly structured, relational, permission-based, compliance-sensitive, aur audit-heavy hai; MySQL transactions, indexes, unique constraints, and relational modeling ke liye strong fit hai.

---

## 🪜 Step-by-Step Implementation

## Step 1: Task Boundary Clear Kiya

`docs/01-micro-tasks.md` me Superadmin Service Task 2 ye hai:

| S.No | Task Name | Detail | Dependency | Priority |
|---:|---|---|---|---|
| 2 | Choose MySQL | Admin actions and permissions highly structured and auditable hain, MySQL suitable hai. | Admin domain | P1 |

### Is task me allowed work

- Superadmin Service ke liye DB choice define karna
- MySQL choose karne ka reason explain karna
- Tables, relationships, indexes ka target direction document karna
- External tools/libraries ka explanation dena
- Future folder structure aur migration approach explain karna
- SQL/Go examples dena for understanding

### Is task me not allowed work

| Area | Reason |
|---|---|
| Actual `.sql` migration create karna | Ye schema implementation step me hoga, Task 2 ka output decision guide hai |
| Admin RBAC enforce karna | Ye **Task 3: Admin RBAC** ka scope hai |
| User/seller control APIs banana | Ye **Task 4: User/seller controls** ka scope hai |
| Refund review code banana | Ye **Task 5: Order/payment controls** ka scope hai |
| Session visibility implement karna | Ye **Task 6: Session visibility** ka scope hai |
| Platform settings feature implement karna | Ye **Task 7: Platform settings** ka scope hai |
| Immutable audit log production code banana | Ye **Task 8: Admin audit logs** ka scope hai |

> 🔴 **Boundary rule:** Is task me MySQL choose aur document kiya gaya hai. Production migration, repository, API, ya handler code implement nahi kiya gaya.

---

## Step 2: Admin Domain Data Shape Analyze Kiya

Task 1 ke domain ke hisaab se Superadmin Service in resources ko control karegi:

| Domain Area | Data Type | DB Requirement |
|---|---|---|
| Admin users | Structured identity + role + status | Unique constraints, status filter, fast lookup |
| Permissions | Role-permission mapping | Relational joins / unique mapping |
| Platform settings | Key-value but controlled | Unique key, JSON value, audit trail |
| Audit logs | Append-only operational records | Indexed by actor, resource, action, time |
| Review tasks | Approval queue | Status transitions, assignment, history |
| Cross-service references | IDs from User/Order/Payment/etc. | Store external IDs as references, no direct FK to other service DBs |

### Conclusion

Ye data random flexible documents jaisa nahi hai. Isme:

- fixed entities hain,
- predictable fields hain,
- duplicate permission mapping avoid karni hai,
- audit records query karne hain,
- status transitions reliable chahiye,
- high-risk mutations me transaction useful hoga.

Isliye MySQL fit hai.

---

## Step 3: MySQL Select Kiya

### Why MySQL?

| Requirement | MySQL ka benefit |
|---|---|
| Admin user uniqueness | `UNIQUE KEY` se duplicate admin mapping block hoti hai |
| Role-permission mapping | Relational table se clean mapping maintain hoti hai |
| Audit trail | Indexed append-only table fast filtering support karta hai |
| Platform settings | `setting_key` unique + `JSON` value flexible config support karta hai |
| Review workflow | Status enum, timestamps, assigned/reviewed fields structured rahte hain |
| Consistency | InnoDB transaction se setting update + audit insert ek saath ho sakta hai |
| Reporting | SQL filters, joins, date range queries admin dashboard ke liye useful hain |
| Compliance | Immutable logs, indexed lookup, archival strategy easier hoti hai |

### Why not MongoDB as primary DB?

MongoDB flexible documents ke liye great hai, but Superadmin data mostly relational and permission-oriented hai. Permissions aur audit filters ko predictable schema chahiye.

### Why not Redis as primary DB?

Redis cache/rate-limit ke liye useful hai, but durable audit logs aur admin permissions ke source of truth ke liye primary DB nahi hona chahiye.

### Why not Typesense/OpenSearch as primary DB?

Search engines admin search ke liye later read model ban sakte hain, but authoritative settings, permissions, and audit data ka source of truth nahi banenge.

---

## Step 4: Database Ownership Rule Define Kiya

Project ka golden rule:

> Har service apni database own karegi. Dusri service ke DB ko directly read/write nahi karna.

### Superadmin Service DB ownership

```text
superadmin-service owns: superadmin_db
```

### Allowed

- Superadmin Service apne MySQL DB me admin users, permissions, settings, review tasks, audit logs store karegi.
- Downstream service data chahiye to gRPC call use hoga.
- Cross-service IDs ko string reference ke form me store kiya jayega.

### Not allowed

- Superadmin Service directly User Service DB query nahi karegi.
- Superadmin Service directly Payment Service DB update nahi karegi.
- Superadmin Service direct foreign key kisi doosri service ke database table par nahi banayegi.

### Example

```text
Correct:
Superadmin Service -> User Service gRPC -> UpdateSellerStatus

Wrong:
Superadmin Service -> user_db.seller_profiles table direct update
```

---

## Step 5: Target MySQL Database Name Decide Kiya

Reference DDL me database name:

```sql
CREATE DATABASE IF NOT EXISTS superadmin_db
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;
```

### Explanation

| Part | Meaning |
|---|---|
| `superadmin_db` | Superadmin Service ka private database |
| `utf8mb4` | Full Unicode support, future-safe text storage |
| `utf8mb4_unicode_ci` | Case-insensitive Unicode collation |
| `InnoDB` | Transaction, row-level locking, and crash recovery support |

> 🟡 **Beginner note:** `utf8mb4` use karna better hai kyunki ye normal text ke saath emoji/special Unicode bhi support karta hai. Admin notes/reasons me future me rich text ya symbols aa sakte hain.

---

## Step 6: Planned Tables Identify Kiye

Docs ke according Superadmin DB me ye tables planned hain:

| Table | Purpose |
|---|---|
| `admin_users` | Platform admin identity, role, status, MFA requirement |
| `admin_permissions` | Permission registry, jaise `platform:settings:write` |
| `admin_role_permissions` | Role se permission mapping |
| `platform_settings` | Commission, feature flags, maintenance mode jaise settings |
| `admin_audit_logs` | Har admin mutation ka immutable audit record |
| `admin_review_tasks` | Seller approval, refund review, setting approval jaise workflow tasks |

### Scope reminder

Ye table list **target schema direction** hai. Is task me actual migration file create nahi ki gayi.

---

## Step 7: Table-by-Table Design Explain Kiya

## 7.1 `admin_users`

Admin user table platform admin identity store karegi.

```sql
CREATE TABLE IF NOT EXISTS admin_users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  admin_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  role VARCHAR(64) NOT NULL,
  status ENUM('active', 'disabled') NOT NULL DEFAULT 'active',
  mfa_required BOOLEAN NOT NULL DEFAULT TRUE,
  created_by VARCHAR(64) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_users_admin_id (admin_id),
  UNIQUE KEY uk_admin_users_user_id (user_id),
  KEY idx_admin_users_role_status (role, status)
) ENGINE=InnoDB;
```

### Explanation

| Column | Why needed |
|---|---|
| `admin_id` | Public/internal admin identifier |
| `user_id` | Auth/User Service user reference |
| `role` | MVP me primary admin role |
| `status` | Admin ko active/disabled mark karne ke liye |
| `mfa_required` | Admin security strict rakhne ke liye |
| `created_by` | Kis admin ne admin account create kiya |

> 🟡 **Note:** Future me multi-role support chahiye to `admin_role_assignments` table add ho sakti hai. Current reference schema MVP primary-role approach follow karta hai.

---

## 7.2 `admin_permissions`

Permission registry me allowed permission keys store honge.

```sql
CREATE TABLE IF NOT EXISTS admin_permissions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  permission_id VARCHAR(64) NOT NULL,
  permission_key VARCHAR(128) NOT NULL,
  description VARCHAR(512) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_permissions_id (permission_id),
  UNIQUE KEY uk_admin_permissions_key (permission_key)
) ENGINE=InnoDB;
```

### Example permissions

| Permission Key | Meaning |
|---|---|
| `admin:users:read` | Users ko admin panel me view karna |
| `admin:users:status:update` | User block/unblock karna |
| `admin:sellers:review` | Seller KYC approve/reject karna |
| `payment:refund:review` | Refund approve/reject karna |
| `platform:settings:write` | Platform settings change karna |
| `audit:logs:read` | Audit logs view karna |

---

## 7.3 `admin_role_permissions`

Ye table roles ko permissions se map karegi.

```sql
CREATE TABLE IF NOT EXISTS admin_role_permissions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  role VARCHAR(64) NOT NULL,
  permission_key VARCHAR(128) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_role_permission (role, permission_key)
) ENGINE=InnoDB;
```

### Why relational mapping?

Same permission multiple roles ko mil sakti hai, aur same role ke paas multiple permissions ho sakti hain. Relational mapping duplicate aur inconsistent data reduce karti hai.

### Example mapping

```sql
INSERT INTO admin_role_permissions (role, permission_key)
VALUES
  ('superadmin', 'platform:settings:write'),
  ('superadmin', 'audit:logs:read'),
  ('finance_admin', 'payment:refund:review'),
  ('operations_admin', 'admin:users:status:update'),
  ('catalog_admin', 'admin:sellers:review');
```

> 🔵 **Task boundary:** Actual RBAC enforcement and seed data **Task 3: Admin RBAC** me implement hoga.

---

## 7.4 `platform_settings`

Platform settings table controlled key-value settings store karegi.

```sql
CREATE TABLE IF NOT EXISTS platform_settings (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  setting_key VARCHAR(128) NOT NULL,
  setting_value JSON NOT NULL,
  updated_by VARCHAR(64) NOT NULL,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_platform_settings_key (setting_key)
) ENGINE=InnoDB;
```

### Why JSON value?

Settings ka key fixed ho sakta hai, but value shape setting ke type ke hisaab se vary karega.

### Example setting values

```json
{
  "enabled": true,
  "message": "Platform maintenance is active",
  "starts_at": "2026-06-01T01:00:00Z",
  "ends_at": "2026-06-01T02:00:00Z"
}
```

```json
{
  "default_rate_bps": 1200,
  "category_overrides": {
    "electronics": 900,
    "fashion": 1500
  }
}
```

### Example keys

| Setting Key | Value Type | Risk |
|---|---|---|
| `maintenance_mode` | JSON object | Critical |
| `commission_rules` | JSON object | High |
| `feature_flags` | JSON object | High |
| `search_synonyms_config` | JSON object | Medium |

---

## 7.5 `admin_audit_logs`

Audit log table sensitive admin actions ka immutable record store karegi.

```sql
CREATE TABLE IF NOT EXISTS admin_audit_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  audit_id VARCHAR(64) NOT NULL,
  actor_admin_id VARCHAR(64) NOT NULL,
  action VARCHAR(128) NOT NULL,
  resource_type VARCHAR(64) NOT NULL,
  resource_id VARCHAR(64) NOT NULL,
  request_id VARCHAR(64) NULL,
  ip_hash CHAR(64) NULL,
  reason VARCHAR(512) NULL,
  before_json JSON NULL,
  after_json JSON NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_audit_logs_audit_id (audit_id),
  KEY idx_admin_audit_actor_created (actor_admin_id, created_at),
  KEY idx_admin_audit_resource (resource_type, resource_id),
  KEY idx_admin_audit_action_created (action, created_at)
) ENGINE=InnoDB;
```

### Why audit table is critical?

Superadmin actions high-risk hote hain. Example:

- User block
- Seller suspend
- Refund approve
- Platform setting update
- Admin data export

Har mutation ke liye ye answer milna chahiye:

```text
Kis admin ne action kiya?
Kis resource par action hua?
Kab hua?
Reason kya tha?
Before/after state kya thi?
Request/session/IP context kya tha?
```

> 🔴 **Important:** Production me audit logs append-only hone chahiye. Update/delete operations strictly restricted honi chahiye.

---

## 7.6 `admin_review_tasks`

Review tasks table approval workflows ke liye queue store karegi.

```sql
CREATE TABLE IF NOT EXISTS admin_review_tasks (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  task_id VARCHAR(64) NOT NULL,
  task_type VARCHAR(64) NOT NULL,
  resource_type VARCHAR(64) NOT NULL,
  resource_id VARCHAR(64) NOT NULL,
  status ENUM('open', 'approved', 'rejected', 'cancelled') NOT NULL DEFAULT 'open',
  assigned_to VARCHAR(64) NULL,
  created_by VARCHAR(64) NULL,
  reviewed_by VARCHAR(64) NULL,
  reviewed_at TIMESTAMP NULL,
  reason VARCHAR(512) NULL,
  metadata JSON NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_review_tasks_task_id (task_id),
  KEY idx_review_tasks_status_type (status, task_type),
  KEY idx_review_tasks_resource (resource_type, resource_id)
) ENGINE=InnoDB;
```

### Example task types

| Task Type | Resource Type | Example |
|---|---|---|
| `seller_kyc_review` | `seller` | Seller KYC approve/reject |
| `refund_review` | `refund` | High-value refund approval |
| `setting_change_review` | `platform_setting` | Critical setting approval |
| `seller_suspension_review` | `seller` | High-risk seller suspension |

---

## Step 8: Logical ER Diagram Banaya

> Ye logical diagram hai. Cross-service IDs string references hain; direct foreign keys doosri services ke DB par nahi banenge.

```mermaid
erDiagram
    ADMIN_USERS ||--o{ ADMIN_AUDIT_LOGS : performs
    ADMIN_USERS ||--o{ ADMIN_REVIEW_TASKS : reviews
    ADMIN_PERMISSIONS ||--o{ ADMIN_ROLE_PERMISSIONS : maps
    PLATFORM_SETTINGS ||--o{ ADMIN_AUDIT_LOGS : audited_as_resource
    ADMIN_REVIEW_TASKS ||--o{ ADMIN_AUDIT_LOGS : audited_as_resource

    ADMIN_USERS {
        bigint id PK
        string admin_id UK
        string user_id UK
        string role
        enum status
        boolean mfa_required
        string created_by
        timestamp created_at
        timestamp updated_at
    }

    ADMIN_PERMISSIONS {
        bigint id PK
        string permission_id UK
        string permission_key UK
        string description
        timestamp created_at
    }

    ADMIN_ROLE_PERMISSIONS {
        bigint id PK
        string role
        string permission_key
        timestamp created_at
    }

    PLATFORM_SETTINGS {
        bigint id PK
        string setting_key UK
        json setting_value
        string updated_by
        timestamp updated_at
    }

    ADMIN_AUDIT_LOGS {
        bigint id PK
        string audit_id UK
        string actor_admin_id
        string action
        string resource_type
        string resource_id
        string request_id
        string ip_hash
        string reason
        json before_json
        json after_json
        timestamp created_at
    }

    ADMIN_REVIEW_TASKS {
        bigint id PK
        string task_id UK
        string task_type
        string resource_type
        string resource_id
        enum status
        string assigned_to
        string reviewed_by
        timestamp reviewed_at
        json metadata
        timestamp created_at
        timestamp updated_at
    }
```

---

## Step 9: Architecture Flow Define Kiya

```mermaid
flowchart LR
    Panel[Superadmin Panel] --> GW[API Gateway]
    GW --> SA[Superadmin Service]

    SA --> DB[(MySQL: superadmin_db)]

    SA --> User[User Service]
    SA --> Order[Order Service]
    SA --> Payment[Payment Service]
    SA --> Session[Session Service]
    SA --> Search[Search Service]
    SA --> CMS[CMS Service]

    DB --> Audit[Admin Audit Logs]
    DB --> Settings[Platform Settings]
    DB --> Reviews[Review Tasks]

    classDef primary fill:#e8f1ff,stroke:#4479A1,color:#102a43
    classDef db fill:#fff4e5,stroke:#cc7a00,color:#3d2500
    classDef service fill:#ecfdf3,stroke:#12805c,color:#063b2f

    class Panel,GW,SA primary
    class DB,Audit,Settings,Reviews db
    class User,Order,Payment,Session,Search,CMS service
```

### Explanation

- Frontend direct MySQL se connect nahi karega.
- API Gateway admin JWT/RBAC check karega.
- Superadmin Service apne DB me admin-specific records maintain karegi.
- User/Payment/Order jaise resources ke original owners unki services hi rahengi.
- Mutations me audit record store hoga.

---

## Step 10: Transaction Strategy Define Kiya

High-risk admin mutations me MySQL transaction useful hoga.

### Example: platform setting update

```mermaid
sequenceDiagram
    participant Admin
    participant GW as API Gateway
    participant SA as Superadmin Service
    participant DB as MySQL

    Admin->>GW: PATCH /api/v1/admin/settings/{key}
    GW->>SA: UpdatePlatformSetting(admin context)
    SA->>DB: BEGIN
    SA->>DB: Read old setting
    SA->>DB: Upsert new setting
    SA->>DB: Insert admin_audit_logs row
    SA->>DB: COMMIT
    SA-->>GW: updated setting
    GW-->>Admin: 200 OK
```

### Why transaction?

```text
Setting update successful ho aur audit missing ho jaye = bad.
Audit insert successful ho aur setting update fail ho jaye = confusing.

Transaction dono ko ek atomic unit banata hai.
```

### Conceptual Go example

```go
func (r *Repository) UpdateSettingWithAudit(ctx context.Context, input UpdateSettingInput) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // 1. Existing setting read karo for before_json.
    // 2. New setting upsert karo.
    // 3. Audit log insert karo.
    // 4. Commit karo.

    if err := tx.Commit(); err != nil {
        return err
    }
    return nil
}
```

> 🟡 **Note:** Ye sirf code example hai. Actual repository implementation is task me create nahi ki gayi.

---

## Step 11: Indexing Strategy Explain Kiya

Admin panel me common queries predictable hongi:

| Query | Useful index |
|---|---|
| Admin user by auth user ID | `admin_users(user_id)` unique |
| Active admins by role | `admin_users(role, status)` |
| Setting by key | `platform_settings(setting_key)` unique |
| Audit logs by actor/date | `admin_audit_logs(actor_admin_id, created_at)` |
| Audit logs by resource | `admin_audit_logs(resource_type, resource_id)` |
| Audit logs by action/date | `admin_audit_logs(action, created_at)` |
| Open review tasks by type | `admin_review_tasks(status, task_type)` |
| Review tasks by resource | `admin_review_tasks(resource_type, resource_id)` |

### Beginner explanation

Index ek book ke index jaisa hota hai. Agar admin dashboard me baar-baar "is admin ne kya actions kiye?" query chalti hai, to `actor_admin_id + created_at` index query ko fast banata hai.

---

## Step 12: Data Consistency Rules Define Kiye

### Rule 1: Internal DB changes transaction me

`platform_settings` update + `admin_audit_logs` insert same transaction me ho sakta hai.

### Rule 2: Cross-service changes gRPC se

Seller status update User Service own karegi. Superadmin Service sirf admin decision aur audit context coordinate karegi.

### Rule 3: No cross-service foreign keys

`resource_id` me `user_123`, `seller_456`, `payment_789` jaise IDs store honge, but MySQL FK doosri service ke DB par nahi banega.

### Rule 4: Use idempotency for risky workflows later

Refund review jaise actions me duplicate approval avoid karna important hoga. Iska full implementation Task 5/Task 8 side me handle hoga.

---

## Step 13: External Libraries/Tools Mention Kiye

## 13.1 What was used in this task?

| Tool/Library | Used now? | Reason |
|---|---:|---|
| External package install | ❌ No | Ye documentation-only task tha |
| MySQL server run | ❌ No | Actual DB runtime start nahi kiya gaya |
| Migration tool run | ❌ No | Migration files create/apply nahi kiye gaye |

> 🟢 **Current task output:** Sirf `task2.md` guide generate kiya gaya.

## 13.2 Recommended tools for actual MySQL implementation

Ye tools future implementation me useful honge jab Superadmin Service ka DB code/migrations banenge.

| Tool/Library | What it is | Why used | Install |
|---|---|---|---|
| MySQL 8.x | Relational database | Admin permissions, settings, audit logs ka source of truth | Docker ya package manager |
| InnoDB | MySQL storage engine | Transactions, row-level locking, durability | MySQL ke saath built-in |
| `database/sql` | Go standard DB API | Go service se SQL DB connect karne ke liye | Built-in Go package |
| `github.com/go-sql-driver/mysql` | Go MySQL driver | `database/sql` ko MySQL protocol support deta hai | `go get github.com/go-sql-driver/mysql` |
| `golang-migrate/migrate` | Migration CLI/library | Versioned SQL migrations apply/rollback karne ke liye | `go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest` |
| Docker | Container runtime | Local MySQL quickly run karne ke liye | Docker Desktop / Docker Engine |

### MySQL install/use example with Docker

```bash
docker run --name ecommerce-superadmin-mysql \
  -e MYSQL_ROOT_PASSWORD=localroot \
  -e MYSQL_DATABASE=superadmin_db \
  -p 3306:3306 \
  -d mysql:8.4
```

### MySQL CLI example

```bash
docker exec -it ecommerce-superadmin-mysql mysql -uroot -plocalroot superadmin_db
```

### Go driver install example

```bash
go get github.com/go-sql-driver/mysql
```

### Go connection example

```go
package repository

import (
    "database/sql"

    _ "github.com/go-sql-driver/mysql"
)

func OpenSuperadminDB(dsn string) (*sql.DB, error) {
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, err
    }
    if err := db.Ping(); err != nil {
        return nil, err
    }
    return db, nil
}
```

### Migration command example

```bash
migrate \
  -path backend/services/superadmin-service/migrations \
  -database "mysql://root:localroot@tcp(localhost:3306)/superadmin_db" \
  up
```

> 🟡 **Task boundary:** Commands above are installation/use examples for future implementation. Is task me ye commands run nahi kiye gaye.

---

## Step 14: Future Folder Structure Define Kiya

Task 2 me backend files create nahi karne, but final service implementation ka expected structure docs ke according ye hoga:

```text
backend/
└── services/
    └── superadmin-service/
        ├── cmd/
        │   └── server/
        │       └── main.go
        ├── internal/
        │   ├── domain/
        │   │   ├── admin_user.go
        │   │   ├── audit_log.go
        │   │   └── platform_setting.go
        │   ├── usecase/
        │   │   ├── manage_user.go
        │   │   ├── manage_seller.go
        │   │   ├── manage_payment.go
        │   │   └── audit_log.go
        │   ├── repository/
        │   │   └── mysql_admin_repository.go
        │   └── transport/
        │       └── grpc/
        ├── migrations/
        │   ├── 001_create_superadmin_tables.up.sql
        │   └── 001_create_superadmin_tables.down.sql
        └── deploy/
```

### Current task created structure

```text
TaskImplementation/
└── Superadmin Service/
    ├── task1.md
    └── task2.md
```

> 🔵 **Clear separation:** `TaskImplementation/...` documentation guide hai. `backend/services/...` actual implementation folder future service work me aayega.

---

## Step 15: Environment Configuration Direction

Future Superadmin Service ko MySQL config env vars se milega.

### Example env vars

```env
SUPERADMIN_DB_HOST=localhost
SUPERADMIN_DB_PORT=3306
SUPERADMIN_DB_NAME=superadmin_db
SUPERADMIN_DB_USER=superadmin_app
SUPERADMIN_DB_PASSWORD=change_me
SUPERADMIN_DB_MAX_OPEN_CONNS=20
SUPERADMIN_DB_MAX_IDLE_CONNS=10
SUPERADMIN_DB_CONN_MAX_LIFETIME_SECONDS=300
```

### Why env vars?

- Local, staging, production alag config use kar sakte hain.
- Secrets code me hardcode nahi hote.
- Kubernetes Secrets/External Secrets se inject karna easy hota hai.

> 🔴 **Security note:** DB password, root password, JWT key, provider secret kabhi git me commit nahi karne.

---

## Step 16: Data Lifecycle and Retention Plan

Superadmin DB me audit data long-term grow karega, isliye retention strategy important hai.

| Data | Retention idea | Reason |
|---|---|---|
| `admin_users` | Active + disabled historical records | Admin accountability |
| `admin_permissions` | Long-lived | Stable permission registry |
| `admin_role_permissions` | Version changes audited | Security review |
| `platform_settings` | Current value in table, history via audit logs | Operational config |
| `admin_audit_logs` | Long retention, archive old data | Compliance/security |
| `admin_review_tasks` | Keep closed tasks, archive old if needed | Dispute/debugging |

### Archival idea

```mermaid
flowchart LR
    Hot[(MySQL hot audit logs)] --> Job[Archive job]
    Job --> Obj[(Object storage)]
    Job --> Keep[Keep recent rows in MySQL]
```

> 🟡 **Note:** Archive job Task 2 me implement nahi hua. Ye scalability planning note hai.

---

## Step 17: Read Scaling Direction

Admin audit logs aur search queries heavy ho sakti hain.

### MVP

```text
MySQL primary DB + proper indexes
```

### Later scale options

| Need | Option |
|---|---|
| Heavy audit reads | MySQL read replica |
| Complex audit search | OpenSearch read model |
| Fast settings access | Redis/local cache with invalidation |
| Large audit storage | Partition/archive by date |

### Why not start with everything?

MVP me complexity kam rakhni chahiye. MySQL indexes enough honge jab tak traffic high nahi hota. Read replicas/search projections later add karna safer hai.

---

## Step 18: Security Rules for MySQL Choice

Superadmin DB high-risk hai. Isliye DB design ke saath security rules bhi clear hone chahiye.

| Rule | Explanation |
|---|---|
| Least privilege DB user | App user ko only required permissions milni chahiye |
| No root in app | Service root DB user se connect nahi karegi |
| TLS in production | DB traffic encrypted hona chahiye |
| Secrets manager | DB password env/secret manager se inject hoga |
| Audit append-only | Audit table update/delete restrict honge |
| PII minimization | Audit me full sensitive data store nahi karna |
| IP hash | Raw IP ke bajay hash store karna better hai |
| Backup enabled | Compliance and recovery ke liye backups required |

### App DB user example

```sql
CREATE USER 'superadmin_app'@'%' IDENTIFIED BY 'change_me';

GRANT SELECT, INSERT, UPDATE
ON superadmin_db.*
TO 'superadmin_app'@'%';

-- Audit delete/update permissions production me restrict karne ka plan rakho.
```

> 🔴 **Production note:** Real password secret manager se generate hoga. `change_me` sirf example hai.

---

## Step 19: Query Examples Add Kiye

### Find admin by auth user

```sql
SELECT admin_id, user_id, role, status, mfa_required
FROM admin_users
WHERE user_id = ?;
```

### Check role permission

```sql
SELECT 1
FROM admin_role_permissions
WHERE role = ?
  AND permission_key = ?
LIMIT 1;
```

### Get platform setting

```sql
SELECT setting_key, setting_value, updated_by, updated_at
FROM platform_settings
WHERE setting_key = ?;
```

### List audit logs by resource

```sql
SELECT audit_id, actor_admin_id, action, reason, created_at
FROM admin_audit_logs
WHERE resource_type = ?
  AND resource_id = ?
ORDER BY created_at DESC
LIMIT 50;
```

### List open review tasks

```sql
SELECT task_id, task_type, resource_type, resource_id, assigned_to, created_at
FROM admin_review_tasks
WHERE status = 'open'
  AND task_type = ?
ORDER BY created_at ASC
LIMIT 100;
```

---

## Step 20: MySQL Decision Flow Diagram

```mermaid
flowchart TD
    A[Superadmin data requirement] --> B{Is data structured?}
    B -->|Yes| C{Needs transactions and audit?}
    B -->|No| X[Consider MongoDB]
    C -->|Yes| D{Needs relational permission mapping?}
    C -->|No| Y[Consider simpler store/cache]
    D -->|Yes| E[Choose MySQL]
    D -->|No| Z[Evaluate per feature]

    E --> F[Use InnoDB]
    F --> G[Use unique keys and indexes]
    G --> H[Use JSON columns only for controlled flexible values]
    H --> I[Keep audit logs indexed and append-only]
```

---

## Step 21: Implementation Checklist

| Check | Status | Notes |
|---|---:|---|
| `TaskImplementation/` folder available | ✅ | Existing folder used |
| `TaskImplementation/Superadmin Service/` folder available | ✅ | Existing folder kept |
| `task2.md` created | ✅ | This guide |
| Task scope identified from docs | ✅ | `Choose MySQL` |
| Dependency checked | ✅ | Depends on admin domain |
| MySQL selected | ✅ | Structured/auditable admin data |
| External tools documented | ✅ | MySQL, driver, migrate, Docker |
| Folder structure documented | ✅ | Current and future structure |
| Code examples added | ✅ | SQL + Go examples |
| Diagrams added | ✅ | Architecture, ER, transaction, decision flow |
| Backend code changes made | ❌ | Intentionally not done |
| Migration files created | ❌ | Intentionally not done |

---

## 🧪 Validation Notes

Since this task is documentation-only, validation means:

- Requirement source matches `docs/01-micro-tasks.md`
- MySQL decision matches `docs/04-microservice-design.md`
- Table list matches `docs/05-database-design.md`
- DDL direction matches `database/draw.sql`
- Scope does not enter Task 3-8 implementation
- Folder output matches requested structure

### Validation result

```text
PASS: Superadmin Service Task 2 documented as MySQL choice.
PASS: No backend service implementation added.
PASS: No migration file generated.
PASS: Required task2.md file created in correct folder.
```

---

## 🧾 Final Task 2 Conclusion

Superadmin Service ke liye **MySQL** choose kiya gaya because admin users, permissions, role mappings, platform settings, review tasks, and audit logs structured and compliance-sensitive data hain.

### Final database choice

```text
Database: MySQL 8.x
Storage Engine: InnoDB
DB Name: superadmin_db
Use Case: Admin permissions, platform settings, review workflows, audit trails
```

### Final scope statement

> Ye task sirf DB choice aur implementation guide complete karta hai. Actual schema migration, repository code, RBAC enforcement, audit log writing, and admin APIs future Superadmin Service tasks me implement honge.

