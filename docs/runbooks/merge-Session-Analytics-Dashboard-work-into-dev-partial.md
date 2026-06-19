# Merge Session Analytics Dashboard Work Into Dev

Use this runbook to bring Session Analytics Dashboard work from
`feature/session-analytics-dashboard-service` into `dev` without merging the
whole feature branch and without deleting existing files from `dev`.

Do not run:

```bash
git merge feature/session-analytics-dashboard-service
```

The feature branch has diverged from `dev` and contains many unrelated
deletions. Only added and modified Session Analytics Dashboard files are
selected. Shared files are updated additively so current `dev` content is not
removed.

## Rules

- Restore only added and modified files under these owned paths:
  `backend/services/session-service`, `frontend/session-analytics-dashboard`,
  and `TaskImplementation/Session Analytics Dashboard Service`.
- Keep every file that exists on `dev` but is deleted or missing on the feature
  branch.
- Do not apply deletion diffs.
- Handle `api/master-api.json`, `backend/go.work`, and `backend/go.work.sum`
  manually. Do not replace them with their feature-branch versions.
- Do not modify, restore, stage, or commit `docs/01-micro-tasks.md`.
- Commit with this exact message:

```text
Merge Session Analytics Dashboard work into dev
```

## 1. Precheck

```bash
set -euo pipefail

git status --short --branch
git rev-parse --verify dev
git rev-parse --verify feature/session-analytics-dashboard-service
test -z "$(git status --porcelain)" || { echo "Working tree is dirty. Commit/stash first."; exit 1; }

if git show-ref --verify --quiet refs/heads/chore/Session-Analytics-Dashboard-selective; then
  echo "Integration branch already exists. Stop and inspect it."
  exit 1
fi
```

What this does:

- Confirms the source and target branches exist.
- Requires a clean working tree before any branch operation.
- Prevents accidentally reusing an existing integration branch.

## 2. Create A Safe Integration Branch From `dev`

```bash
git switch dev
test -z "$(git status --porcelain)" || { echo "dev is dirty. Stop."; exit 1; }

git switch -c chore/Session-Analytics-Dashboard-selective
test "$(git rev-parse HEAD)" = "$(git rev-parse dev)"
```

What this does:

- Starts the selective work at the current `dev` commit.
- Keeps `dev` unchanged until the final fast-forward.

## 3. Review The Full Branch Difference

```bash
git diff --name-status --find-renames dev..feature/session-analytics-dashboard-service
git diff --name-status --no-renames --diff-filter=D dev..feature/session-analytics-dashboard-service
git diff --name-status --no-renames --diff-filter=AM dev..feature/session-analytics-dashboard-service
```

Review the shared files separately:

```bash
git diff dev..feature/session-analytics-dashboard-service -- api/master-api.json
git diff dev..feature/session-analytics-dashboard-service -- backend/go.work
git diff dev..feature/session-analytics-dashboard-service -- backend/go.work.sum
git diff dev..feature/session-analytics-dashboard-service -- docs/01-micro-tasks.md
```

Important observations from the current repository state:

- The feature branch contains Session Analytics Dashboard work in the backend,
  frontend, and task-implementation paths listed in the rules.
- It is missing many files that now exist on `dev`, including files within
  `backend/services/session-service`. Those deletions must remain unapplied.
- Its `backend/go.work` removes existing `dev` workspace modules. The
  `./services/session-service` entry is already present on the current `dev`.
- Its `backend/go.work.sum` removes existing checksums while adding four new
  `golang.org/x/sync` and `golang.org/x/text` checksum lines.
- Its `api/master-api.json` combines relevant analytics additions with
  unrelated removals and schema edits, so the entire file must not be restored.
- Its `docs/01-micro-tasks.md` change is intentionally excluded from this
  runbook.

## 4. Build The Auto-Restore File List

```bash
git diff --name-only --no-renames --diff-filter=AM \
  dev..feature/session-analytics-dashboard-service -- \
  backend/services/session-service \
  frontend/session-analytics-dashboard \
  'TaskImplementation/Session Analytics Dashboard Service' \
  ':(exclude)docs/01-micro-tasks.md' \
  > /tmp/session-analytics-dashboard-auto-restore.txt

sed -n '1,260p' /tmp/session-analytics-dashboard-auto-restore.txt

if grep -Fxq 'docs/01-micro-tasks.md' /tmp/session-analytics-dashboard-auto-restore.txt; then
  echo "Protected micro-task file entered the restore list. Stop."
  exit 1
fi
```

Also display owned-path deletions, but do not add them to the restore list:

```bash
git diff --name-status --no-renames --diff-filter=D \
  dev..feature/session-analytics-dashboard-service -- \
  backend/services/session-service \
  frontend/session-analytics-dashboard \
  'TaskImplementation/Session Analytics Dashboard Service'
```

What this does:

- Selects only added and modified files in the three service-owned paths.
- Uses `--no-renames` so
  `backend/services/session-service/internal/transport/http/routes.go` is
  treated as a modified session-service file instead of being paired with an
  unrelated Auth Service deletion.
- Keeps all deletions out of the restore list.

If the auto-restore list is empty, continue with the shared-file steps.

## 5. Restore Added And Modified Service Files Using `git restore --overlay`

```bash
if [ -s /tmp/session-analytics-dashboard-auto-restore.txt ]; then
  git restore --overlay \
    --source=feature/session-analytics-dashboard-service \
    --worktree \
    --pathspec-from-file=/tmp/session-analytics-dashboard-auto-restore.txt
else
  echo "No Session Analytics Dashboard owned files to auto-restore."
fi
```

What this does:

- Copies only the selected added and modified files from the source branch.
- Uses overlay mode so files present on `dev` but absent on the source branch
  are not removed.

## 6. Handle Shared Files Manually

### 6.1 Update `api/master-api.json` Additively

Review the file again, then use patch mode:

```bash
git diff dev..feature/session-analytics-dashboard-service -- api/master-api.json
git restore -p \
  --source=feature/session-analytics-dashboard-service \
  --worktree -- api/master-api.json
```

Accept only additions for these endpoint IDs:

```text
analytics.privacy_deletion_create
analytics.privacy_deletion_list
analytics.privacy_deletion_preview
analytics.privacy_retention_get
analytics.privacy_retention_update
analytics.privacy_settings_get
analytics.privacy_settings_update
analytics.report_export
analytics.report_schedule_create
analytics.report_schedule_delete
analytics.report_schedule_update
analytics.report_schedules_list
analytics.retention
```

Accept only additions for these schemas:

```text
AnalyticsReportExportRequest
CSVFileResponse
CreateDeletionRequest
DeletionPreviewRequest
DeletionPreviewResponse
DeletionRequest
DeletionRequestListResponse
PrivacyMaskingSettings
PrivacyPermissions
PrivacySettingsResponse
PrivacySettingsUpdateRequest
ReportFilterSnapshot
ReportSchedule
ReportScheduleInput
ReportScheduleListResponse
ReportScheduleStatusUpdateRequest
RetentionSettings
RetentionSettingsUpdateRequest
```

Press `n` for all endpoint or schema removals and for unrelated changes. In
particular, do not accept feature-branch edits to `PlatformSetting`,
`PlatformSettingInput`, or `SearchSynonymListResponse`. Use `s` to split mixed
hunks and `e` to edit a hunk when an allowed addition shares a hunk with an
unrelated change.

Validate the result:

```bash
python3 -m json.tool api/master-api.json > /dev/null
git diff --check -- api/master-api.json
git diff --unified=0 -- api/master-api.json | sed -n '/^---/d; /^-/p'
git diff -- api/master-api.json
```

The removed-line command must print nothing. If the result contains any
removal or unrelated change, discard only this working-tree edit and retry:

```bash
git restore --source=HEAD --worktree -- api/master-api.json
```

### 6.2 Keep `backend/go.work` Additive

```bash
(
  cd backend
  go work use ./services/session-service
)

git diff --unified=0 -- backend/go.work | sed -n '/^---/d; /^-/p'
git diff -- backend/go.work
```

The current `dev` already contains `./services/session-service`, so this should
normally produce no diff. It must not remove any other workspace module.

### 6.3 Union `backend/go.work.sum`

```bash
git show dev:backend/go.work.sum > /tmp/dev-go.work.sum
git show feature/session-analytics-dashboard-service:backend/go.work.sum \
  > /tmp/feature-session-analytics-go.work.sum

LC_ALL=C sort -u \
  /tmp/dev-go.work.sum \
  /tmp/feature-session-analytics-go.work.sum \
  > backend/go.work.sum

git diff --unified=0 -- backend/go.work.sum | sed -n '/^---/d; /^-/p'
git diff -- backend/go.work.sum
```

The removed-line command must print nothing. In the current repository state,
the union keeps all `dev` checksums and adds four source-branch checksum lines.

### 6.4 Protect `docs/01-micro-tasks.md`

Do not run any restore, add, or commit command specifically for this file.
Confirm that it remains unchanged and unstaged:

```bash
git diff --quiet -- docs/01-micro-tasks.md || { echo "Protected file changed. Stop."; exit 1; }
git diff --cached --quiet -- docs/01-micro-tasks.md || { echo "Protected file staged. Stop."; exit 1; }
```

## 7. Confirm No Deletions Were Applied

```bash
git diff --name-status --diff-filter=D --exit-code
git diff --quiet -- docs/01-micro-tasks.md || { echo "Protected file changed. Stop."; exit 1; }
```

Both commands must succeed, and the deletion check must print nothing. If a
deletion appears, stop before staging and inspect it.

## 8. Show Status And Diff Stat Before Staging

```bash
git status --short

if [ -s /tmp/session-analytics-dashboard-auto-restore.txt ]; then
  git add -N --pathspec-from-file=/tmp/session-analytics-dashboard-auto-restore.txt
fi

git status
git diff --stat
git diff --name-status
git diff --name-status --diff-filter=D --exit-code
git diff --quiet -- docs/01-micro-tasks.md || { echo "Protected file changed. Stop."; exit 1; }
```

`git add -N` records intent-to-add for new owned files so they appear in the
unstaged diff stat. It does not stage their contents. The restore list cannot
contain `docs/01-micro-tasks.md`.

## 9. Stage Only The Selected Changes

```bash
if [ -s /tmp/session-analytics-dashboard-auto-restore.txt ]; then
  git add --pathspec-from-file=/tmp/session-analytics-dashboard-auto-restore.txt
fi

git add -p api/master-api.json
git add backend/go.work backend/go.work.sum

git status
git diff --cached --stat
git diff --cached --name-status
git diff --cached --name-status --diff-filter=D --exit-code
git diff --cached --quiet -- docs/01-micro-tasks.md || { echo "Protected file staged. Stop."; exit 1; }
```

In `git add -p`, stage only the already-reviewed Session Analytics Dashboard
API additions. The staged deletion check must print nothing, and the protected
micro-task check must succeed.

## 10. Commit With The Exact Message

Verify the staged content once more, then commit while explicitly excluding the
protected micro-task file:

```bash
git diff --cached --check
git diff --cached --quiet -- docs/01-micro-tasks.md || { echo "Protected file staged. Stop."; exit 1; }

git commit -m "Merge Session Analytics Dashboard work into dev" -- \
  . \
  ':(exclude)docs/01-micro-tasks.md'
```

## 11. Verify The Integration Commit

```bash
test "$(git log -1 --format=%s)" = "Merge Session Analytics Dashboard work into dev"
git status --short --branch
git log --oneline -1
git show --stat --oneline --summary HEAD
git show --name-status --format= HEAD | sed -n '/^D/p'
git show --name-only --format= HEAD -- docs/01-micro-tasks.md
```

The last two commands must print nothing: the commit must contain neither
deletions nor `docs/01-micro-tasks.md`.

## 12. Fast-Forward `dev`

```bash
git switch dev
git merge --ff-only chore/Session-Analytics-Dashboard-selective
```

This fast-forwards `dev` to the reviewed integration commit. It does not merge
`feature/session-analytics-dashboard-service`, and it fails safely if `dev`
changed and can no longer fast-forward.

## 13. Final Verification

```bash
test "$(git branch --show-current)" = "dev"
test "$(git rev-parse dev)" = "$(git rev-parse chore/Session-Analytics-Dashboard-selective)"
test "$(git log -1 --format=%s)" = "Merge Session Analytics Dashboard work into dev"

git status --short --branch
git log --oneline -1
git show --stat --oneline --summary HEAD
git show --name-status --format= HEAD | sed -n '/^D/p'
git show --name-only --format= HEAD -- docs/01-micro-tasks.md
```

The working tree should be clean. The final two commands must print nothing,
confirming that the selective commit applied no deletions and did not modify
`docs/01-micro-tasks.md`.
