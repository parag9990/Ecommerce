# Merge Superadmin Service Work Into Dev

Use this runbook to bring Superadmin Service work from
`feature/superadmin-service` into `dev` without merging the whole feature branch
and without deleting existing files from `dev`.

Never merge the source branch directly. This runbook uses a selective restore
onto a safe integration branch instead.

This branch has many unrelated deletions. We only apply added and modified
Superadmin Service related files, and we handle shared files carefully.

## Rules

- Add new Superadmin Service related files from `feature/superadmin-service`.
- If a file already exists on `dev`, update it only when it is modified in
  `feature/superadmin-service`.
- If a file exists on `dev` but is deleted or missing on
  `feature/superadmin-service`, keep the `dev` file.
- Do not apply deletion diffs.
- Do not restore, edit, stage, or commit `docs/01-micro-tasks.md`.
- Handle shared files manually so existing `dev` content is not removed.
- Commit with this exact message:

```text
Merge Superadmin Service work into dev
```

## 1. Precheck

```bash
set -euo pipefail

git status --short --branch
git rev-parse --verify dev
git rev-parse --verify feature/superadmin-service
test -z "$(git status --porcelain)" || { echo "Working tree is dirty. Commit/stash first."; exit 1; }
```

What this does:

- Confirms both branches exist.
- Confirms the current working tree is clean before changing branches.
- Prevents local work from being mixed into the selective integration commit.

## 2. Create A Safe Integration Branch From Dev

```bash
git switch dev
test -z "$(git status --porcelain)" || { echo "dev is dirty. Stop."; exit 1; }

git switch -c chore/Superadmin-Service-selective
```

What this does:

- Starts from `dev`.
- Creates a temporary branch for the selective merge work.
- Keeps `dev` safe until the final fast-forward step.

## 3. Review The Full Branch Difference

```bash
git diff --name-status dev..feature/superadmin-service
git diff --name-status --diff-filter=D dev..feature/superadmin-service
git diff --name-status --diff-filter=AM dev..feature/superadmin-service
```

What this does:

- Shows the complete difference from `dev` to
  `feature/superadmin-service`.
- Shows deletions separately. Do not apply those files.
- Shows only added and modified files. These are the only candidates for this
  runbook.

Current important shared files to check:

```bash
git diff dev..feature/superadmin-service -- api/master-api.json
git diff dev..feature/superadmin-service -- backend/go.work
git diff dev..feature/superadmin-service -- backend/go.work.sum
git diff dev..feature/superadmin-service -- docs/01-micro-tasks.md
```

In the current repo state, the Superadmin Service branch updates
`api/master-api.json`, `backend/go.work`, `backend/go.work.sum`, and
`docs/01-micro-tasks.md`.

Important notes:

- `backend/go.work` on the feature branch contains only
  `./services/superadmin-service`; do not restore the whole file.
- `backend/go.work.sum` on the feature branch removes many existing `dev`
  checksum lines; do not restore the whole file.
- `api/master-api.json` includes Superadmin additions mixed with unrelated API
  changes; apply only Superadmin related hunks.
- `docs/01-micro-tasks.md` must remain unchanged in this runbook.

## 4. Build The Auto-Restore File List

```bash
git diff --name-only --diff-filter=AM dev..feature/superadmin-service -- \
  backend/services/superadmin-service \
  'TaskImplementation/Superadmin Service' \
  > /tmp/superadmin-service-auto-restore.txt

sed -n '1,240p' /tmp/superadmin-service-auto-restore.txt
```

What this does:

- Lists Superadmin Service related files that are added or modified.
- Excludes deleted files.
- Includes the Superadmin Service backend module and task implementation docs.
- Excludes protected shared files that need manual handling:
  `api/master-api.json`, `backend/go.work`, `backend/go.work.sum`, and
  `docs/01-micro-tasks.md`.

In the current repo state, this list should include the new
`backend/services/superadmin-service` module and
`TaskImplementation/Superadmin Service` files.

If this file is empty, stop and re-check the branch names and path filters.

## 5. Restore Added And Modified Superadmin Service Files

```bash
if [ -s /tmp/superadmin-service-auto-restore.txt ]; then
  git restore --overlay --source=feature/superadmin-service --worktree --pathspec-from-file=/tmp/superadmin-service-auto-restore.txt
else
  echo "No Superadmin Service source/doc files to auto-restore."
fi
```

What this does:

- Copies selected new files from `feature/superadmin-service`.
- Updates selected existing files only when they differ from `dev`.
- Uses `--overlay` so Git does not remove dev files.
- Does not touch `docs/01-micro-tasks.md`.

## 6. Handle Shared Files Manually

Shared files must not be restored wholesale from `feature/superadmin-service`.
Apply only the Superadmin Service parts and preserve existing `dev` content.

### Update `backend/go.work` Safely

```bash
(
  cd backend
  go work use ./services/superadmin-service
)

git diff -- backend/go.work
```

What this does:

- Adds the Superadmin Service workspace module.
- Keeps existing dev workspace entries such as `./services/api-gateway`,
  `./services/auth-service`, `./services/user-service`, `./shared/gen/go`, and
  `./shared/validation`.
- Avoids replacing `backend/go.work` with the feature branch version, because
  the feature version removes existing dev entries.

### Update `backend/go.work.sum` Without Removing Existing Sums

```bash
git show dev:backend/go.work.sum > /tmp/dev-go.work.sum
git show feature/superadmin-service:backend/go.work.sum > /tmp/feature-superadmin-go.work.sum
LC_ALL=C sort -u /tmp/dev-go.work.sum /tmp/feature-superadmin-go.work.sum > backend/go.work.sum

git diff -- backend/go.work.sum
```

What this does:

- Keeps checksum lines that already exist on `dev`.
- Adds checksum lines introduced by `feature/superadmin-service`, if any.
- Prevents the feature branch from deleting existing `dev` checksum lines.
- In the current repo state, the feature branch mostly removes checksum lines,
  so this step may leave `backend/go.work.sum` unchanged.

### Update `api/master-api.json` Safely

First review the complete shared API diff:

```bash
git diff dev..feature/superadmin-service -- api/master-api.json
```

Now apply only Superadmin Service related API hunks using patch mode:

```bash
git restore -p --source=feature/superadmin-service --worktree -- api/master-api.json
```

When Git asks this question:

```text
Apply this hunk to worktree [y,n,q,a,d,s,e,p,?]?
```

Use these options carefully:

- Press `y` only for hunks that add or update Superadmin Service routes,
  Superadmin Service methods, or Superadmin-owned schemas.
- Press `n` for unrelated product, payment, seller, or non-Superadmin API
  changes.
- Press `n` for any deletion-only hunk.
- Press `s` to split a large hunk into smaller hunks.
- Press `e` to manually edit a hunk if Superadmin additions and unrelated
  changes are mixed together.
- Press `q` to stop patch mode if the diff looks unsafe or confusing.

Superadmin related additions in the current repo state include:

- `admin.order_manual_review`
- `admin.order_dispute_view`
- `admin.session_analytics_access`
- `admin.session_analytics_authorize`
- `admin.search_synonyms`
- `admin.search_synonym_upsert`
- `SuperadminService.GetSessionAnalyticsAccess`
- `SuperadminService.AuthorizeSessionAnalytics`
- `SuperadminService.ListSearchSynonyms`
- `SuperadminService.UpsertSearchSynonym`
- `ManualOrderReviewRequest`
- `AdminReviewTask`
- `PathOrderIDRequest`
- `OrderDisputeViewResponse`
- `AuthorizeSessionAnalyticsRequest`
- `AuthorizeSessionAnalyticsResponse`
- `SessionDashboardAccessResponse`
- Superadmin-specific platform setting and search synonym schema fields

After patch mode finishes, verify the file is valid JSON and that only
Superadmin Service related changes remain:

```bash
python3 -m json.tool api/master-api.json > /tmp/master-api.json.checked
git diff -- api/master-api.json
```

Do not use this command for the shared API file:

```bash
git restore --source=feature/superadmin-service --worktree -- api/master-api.json
```

That command can replace the whole shared API file and apply unrelated removals.
Use patch mode or manual editing only.

### Keep `docs/01-micro-tasks.md` Unchanged

The feature branch changes Superadmin task statuses in
`docs/01-micro-tasks.md`, but this runbook must not modify that file.

Use this verification command only:

```bash
git diff --exit-code -- docs/01-micro-tasks.md
```

What this does:

- Confirms `docs/01-micro-tasks.md` is unchanged in the integration branch.
- Keeps the file excluded from restore, staging, and commit commands.

If this command prints a diff, stop and inspect the working tree before
continuing. Do not restore, stage, or commit `docs/01-micro-tasks.md` as part of
this runbook.

## 7. Confirm No Deletions Were Applied

```bash
git diff --name-status --diff-filter=D --exit-code
```

What this does:

- Confirms the working tree has no deletions.
- This command must print nothing.

If it prints any `D` entry, stop before staging.

## 8. Show Status And Diff Stat Before Staging

```bash
if [ -s /tmp/superadmin-service-auto-restore.txt ]; then
  git add -N --pathspec-from-file=/tmp/superadmin-service-auto-restore.txt
fi

git status
git diff --stat
git diff --name-status --diff-filter=D --exit-code
git diff --exit-code -- docs/01-micro-tasks.md
```

What this does:

- Marks new auto-restored files as intent-to-add so `git diff --stat` shows
  them.
- Shows the required `git status`.
- Shows the required `git diff --stat`.
- Confirms again that no deletion is present.
- Confirms again that `docs/01-micro-tasks.md` is unchanged.

## 9. Stage Only The Selected Changes

```bash
if [ -s /tmp/superadmin-service-auto-restore.txt ]; then
  git add --pathspec-from-file=/tmp/superadmin-service-auto-restore.txt
fi

git add api/master-api.json backend/go.work backend/go.work.sum

git status
git diff --cached --stat
git diff --cached --name-status --diff-filter=D --exit-code
git diff --cached --name-only | grep -Fx 'docs/01-micro-tasks.md' && { echo "docs/01-micro-tasks.md is staged. Stop."; exit 1; } || true
```

What this does:

- Stages selected Superadmin Service added and modified files.
- Stages the reviewed shared files only:
  `api/master-api.json`, `backend/go.work`, and `backend/go.work.sum`.
- Does not stage `docs/01-micro-tasks.md`.
- Confirms no deletions are staged.

The staged deletion check must print nothing.

## 10. Commit

```bash
git commit -m "Merge Superadmin Service work into dev"
```

What this does:

- Creates the required selective integration commit.
- Uses the exact required commit message.

## 11. Verify After Commit

```bash
git status --short --branch
git log --oneline -1
git show --stat --oneline HEAD
git show --name-status --oneline HEAD | grep '^D' && { echo "Commit contains deletions. Stop."; exit 1; } || true
git show --name-only --oneline HEAD | grep -Fx 'docs/01-micro-tasks.md' && { echo "Commit contains docs/01-micro-tasks.md. Stop."; exit 1; } || true
```

What this does:

- Confirms the integration branch is clean after the commit.
- Confirms the latest commit message is:

```text
Merge Superadmin Service work into dev
```

- Confirms the commit contains no deletions.
- Confirms the commit does not contain `docs/01-micro-tasks.md`.

## 12. Fast-Forward Dev

```bash
git switch dev
git merge --ff-only chore/Superadmin-Service-selective
```

What this does:

- Moves `dev` to the selective integration commit.
- Does not merge `feature/superadmin-service`.
- Fails safely if `dev` changed and cannot fast-forward.

## 13. Final Verification

```bash
git status --short --branch
git log --oneline -1
git show --stat --oneline HEAD
git show --name-status --oneline HEAD | grep '^D' && { echo "Commit contains deletions. Stop."; exit 1; } || true
git show --name-only --oneline HEAD | grep -Fx 'docs/01-micro-tasks.md' && { echo "Commit contains docs/01-micro-tasks.md. Stop."; exit 1; } || true
git diff --exit-code HEAD^ HEAD -- docs/01-micro-tasks.md
```

What this does:

- Confirms `dev` is on the new commit.
- Shows the final committed file summary.
- Confirms the final commit contains no deletions.
- Confirms the final commit does not modify `docs/01-micro-tasks.md`.
- Confirms the latest commit message is:

```text
Merge Superadmin Service work into dev
```
