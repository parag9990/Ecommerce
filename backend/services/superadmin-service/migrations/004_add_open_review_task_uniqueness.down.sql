ALTER TABLE admin_review_tasks
  DROP KEY uk_admin_review_tasks_open_resource,
  DROP COLUMN open_resource_key;
