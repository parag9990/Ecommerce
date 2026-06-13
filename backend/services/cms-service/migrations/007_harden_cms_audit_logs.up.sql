ALTER TABLE cms_audit_logs
  ADD COLUMN actor_roles_json JSON NULL AFTER actor_user_id,
  ADD COLUMN request_id VARCHAR(128) NULL AFTER resource_id,
  ADD COLUMN trace_id VARCHAR(128) NULL AFTER request_id,
  ADD COLUMN decision ENUM('allowed', 'denied') NOT NULL DEFAULT 'allowed' AFTER trace_id,
  ADD COLUMN reason VARCHAR(512) NULL AFTER decision,
  ADD COLUMN ip_hash CHAR(64) NULL AFTER reason,
  ADD COLUMN user_agent_hash CHAR(64) NULL AFTER ip_hash,
  ADD KEY idx_cms_audit_request (request_id),
  ADD KEY idx_cms_audit_action_created (action, created_at);
