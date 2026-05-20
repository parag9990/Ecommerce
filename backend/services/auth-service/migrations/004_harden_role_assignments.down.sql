-- Roll back RBAC role-assignment hardening.

USE auth_db;

ALTER TABLE role_assignments
  DROP KEY idx_role_assignments_scope,
  DROP KEY uk_role_assignments_active,
  DROP COLUMN scope_id_key,
  DROP COLUMN scope_type_key,
  DROP COLUMN active_role_key,
  DROP COLUMN reason;
