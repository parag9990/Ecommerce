-- Harden RBAC role assignment persistence.
-- MySQL unique indexes treat NULL values as distinct, so generated key columns
-- are used to enforce one active assignment per account/role/scope.

USE auth_db;

ALTER TABLE role_assignments
  ADD COLUMN reason VARCHAR(512) NULL AFTER assigned_by,
  ADD COLUMN active_role_key TINYINT
    GENERATED ALWAYS AS (CASE WHEN revoked_at IS NULL THEN 1 ELSE NULL END) STORED,
  ADD COLUMN scope_type_key VARCHAR(64)
    GENERATED ALWAYS AS (COALESCE(scope_type, '')) STORED,
  ADD COLUMN scope_id_key VARCHAR(64)
    GENERATED ALWAYS AS (COALESCE(scope_id, '')) STORED,
  ADD UNIQUE KEY uk_role_assignments_active (
    account_id,
    role,
    scope_type_key,
    scope_id_key,
    active_role_key
  ),
  ADD KEY idx_role_assignments_scope (scope_type, scope_id, role);
