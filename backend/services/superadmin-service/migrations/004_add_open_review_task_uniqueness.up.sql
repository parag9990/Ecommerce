ALTER TABLE admin_review_tasks
  ADD COLUMN open_resource_key VARCHAR(256)
    GENERATED ALWAYS AS (
      CASE
        WHEN status = 'open' THEN CONCAT(task_type, ':', resource_type, ':', resource_id)
        ELSE NULL
      END
    ) STORED,
  ADD UNIQUE KEY uk_admin_review_tasks_open_resource (open_resource_key);
