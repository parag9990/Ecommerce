# Merge Session Management Service Work Into Dev (Partial)

Use this runbook to bring Session Management Service work from
`feature/session-management-service` into `dev` without merging the whole
feature branch, without applying deletion diffs, and without modifying
`docs/01-micro-tasks.md`.

The Session Management Service implementation is located in:

- `backend/services/session-service`
- `TaskImplementation/Session Management Service`

The source branch is behind `dev` and its full snapshot difference contains
many unrelated deletions. Never merge the source branch into `dev` or the
integration branch.

## Rules

- Restore only added and modified Session Management Service files from
  `feature/session-management-service`.
- Keep every file that exists on `dev` but is deleted or missing on the source
  branch.
- Do not apply deletion or rename diffs.
- Use `--no-renames` when building the restore list. In the current repository
  state, Git otherwise detects some Session Management Service files as
  renames from unrelated deleted service files.
- Handle `backend/go.work` and `backend/go.work.sum` manually so existing
  `dev` entries are preserved.
- Keep `api/master-api.json` from `dev`; it differs in the full snapshot
  comparison but is not a feature-side Session Management Service change.
- Do not modify, restore, stage, or commit `docs/01-micro-tasks.md`.
- Commit with this exact message:

```text
Merge Session Management Service work into dev
```

## 1. Precheck

```bash
set -euo pipefail

git status --short --branch
git rev-parse --verify dev
git rev-parse --verify feature/session-management-service
test -z "$(git status --porcelain)" || { echo "Working tree is dirty. Commit/stash first."; exit 1; }
```

What this does:

- Confirms both branches exist.
- Confirms the current working tree is clean before changing branches.
- Protects existing work from being mixed into the selective integration.

## 2. Create A Safe Integration Branch From Dev

```bash
git switch dev
test -z "$(git status --porcelain)" || { echo "dev is dirty. Stop."; exit 1; }

git switch -c chore/Session-Management-Service-selective
```

What this does:

- Starts from `dev`.
- Creates a temporary branch for the selective integration work.
- Keeps `dev` unchanged until the final fast-forward step.

## 3. Review The Full Branch Difference

First review the full current snapshot difference:

```bash
git diff --no-renames --name-status dev..feature/session-management-service
git diff --no-renames --name-status --diff-filter=D dev..feature/session-management-service
git diff --no-renames --name-status --diff-filter=AM dev..feature/session-management-service
```

Then review feature-side changes made since the branches' common ancestor:

```bash
git diff --no-renames --name-status dev...feature/session-management-service
git diff --no-renames --name-status --diff-filter=D dev...feature/session-management-service
git diff --no-renames --name-status --diff-filter=AM dev...feature/session-management-service
```

What this does:

- The two-dot comparison shows the complete difference between the current
  branch snapshots. It contains many unrelated deletions and must not be used
  as a blanket restore list.
- The three-dot comparison identifies changes introduced on the feature side
  since the common ancestor.
- `--no-renames` prevents Session Management Service additions from being
  classified as renames from unrelated deleted files.
- Deletions shown by either comparison are review-only. Do not apply them.

In the current repository state, the feature-side comparison contains 87 added
files under the two Session Management Service paths. It also contains
unrelated deletions that must remain unapplied.

Review the important shared and protected files:

```bash
git diff dev...feature/session-management-service -- backend/go.work
git diff dev...feature/session-management-service -- backend/go.work.sum
git diff dev...feature/session-management-service -- docs/01-micro-tasks.md

git diff --name-status dev..feature/session-management-service -- api/master-api.json
git diff --name-status dev...feature/session-management-service -- api/master-api.json
```

Current repository-specific findings:

- The source version of `backend/go.work` removes existing `dev` workspace
  modules, so it must not replace the `dev` file.
- The source version of `backend/go.work.sum` removes existing `dev` checksum
  lines, so the two versions must be combined without removing lines.
- The source branch changes `docs/01-micro-tasks.md`, but this protected file
  must remain exactly as it is on `dev`.
- `api/master-api.json` appears in the two-dot comparison but not the
  feature-side three-dot comparison. Keep the `dev` version unchanged.

## 4. Build The Auto-Restore File List

```bash
git diff --no-renames --name-only --diff-filter=AM \
  dev...feature/session-management-service -- \
  backend/services/session-service \
  'TaskImplementation/Session Management Service' \
  > /tmp/session-management-service-auto-restore.txt

sed -n '1,240p' /tmp/session-management-service-auto-restore.txt
wc -l /tmp/session-management-service-auto-restore.txt

if grep -Fxq 'docs/01-micro-tasks.md' /tmp/session-management-service-auto-restore.txt; then
  echo "Protected docs/01-micro-tasks.md entered the restore list. Stop."
  exit 1
fi

while IFS= read -r path; do
  case "$path" in
    backend/services/session-service/*|"TaskImplementation/Session Management Service/"*) ;;
    *) echo "Unexpected auto-restore path: $path"; exit 1 ;;
  esac
done < /tmp/session-management-service-auto-restore.txt
```

What this does:

- Lists only added and modified feature-side files in the two approved Session
  Management Service paths.
- Excludes deletions and disables rename detection.
- Excludes shared files that require manual handling.
- Defensively rejects the protected `docs/01-micro-tasks.md` file and any path
  outside the approved service roots.

In the current repository state, `wc -l` should report 87 files. If the count
changes, stop and review the updated source branch difference before
continuing.

## 5. Restore Added And Modified Session Management Service Files

```bash
if [ -s /tmp/session-management-service-auto-restore.txt ]; then
  git restore --overlay \
    --source=feature/session-management-service \
    --worktree \
    --pathspec-from-file=/tmp/session-management-service-auto-restore.txt
else
  echo "No Session Management Service files to auto-restore."
fi
```

What this does:

- Copies only the selected Session Management Service files from the source
  branch.
- Uses `--overlay` so Git does not remove files from `dev`.
- Leaves the index unstaged for review.

## 6. Update `backend/go.work` Safely

```bash
(
  cd backend
  go work use ./services/session-service
)

git diff -- backend/go.work
```

What this does:

- Adds `./services/session-service` to the existing `dev` workspace.
- Preserves all existing `dev` workspace modules.
- Avoids replacing `backend/go.work` with the source version, which removes
  existing `dev` entries.

The diff must add the Session Management Service entry without removing any
existing workspace entry.

## 7. Update `backend/go.work.sum` Without Removing Existing Sums

```bash
git show dev:backend/go.work.sum > /tmp/dev-go.work.sum
git show feature/session-management-service:backend/go.work.sum \
  > /tmp/feature-session-management-go.work.sum

LC_ALL=C sort -u \
  /tmp/dev-go.work.sum \
  /tmp/feature-session-management-go.work.sum \
  > backend/go.work.sum

git diff -- backend/go.work.sum
```

What this does:

- Keeps every checksum line already present on `dev`.
- Adds checksum lines available on the Session Management Service branch.
- Prevents the source branch from deleting existing `dev` checksum lines.

## 8. Preserve Protected And Unselected Shared Files

```bash
git diff --name-status dev..feature/session-management-service -- \
  api/master-api.json \
  docs/01-micro-tasks.md

git diff --name-status dev...feature/session-management-service -- \
  api/master-api.json \
  docs/01-micro-tasks.md

test -z "$(git status --porcelain -- api/master-api.json docs/01-micro-tasks.md)" || {
  echo "A protected or unselected shared file changed. Stop."
  exit 1
}
```

What this does:

- Reviews why the shared files differ without restoring either file.
- Confirms `api/master-api.json` remains at the `dev` version.
- Confirms `docs/01-micro-tasks.md` remains completely unchanged.

Do not add either file to a restore list or staging command.

## 9. Confirm No Deletions Were Applied

```bash
git diff --name-status --diff-filter=D --exit-code

test -z "$(git status --porcelain -- api/master-api.json docs/01-micro-tasks.md)" || {
  echo "A protected or unselected shared file changed. Stop."
  exit 1
}
```

What this does:

- Confirms the working tree contains no deletion diffs.
- Confirms the protected and unselected shared files are still unchanged.

The deletion check must print nothing. If it prints any `D` entry, stop before
staging.

## 10. Show Status And Diff Stat Before Staging

```bash
if [ -s /tmp/session-management-service-auto-restore.txt ]; then
  git add -N --pathspec-from-file=/tmp/session-management-service-auto-restore.txt
fi

git status
git diff --stat
git diff --name-status
git diff --check
git diff --name-status --diff-filter=D --exit-code

test -z "$(git status --porcelain -- api/master-api.json docs/01-micro-tasks.md)" || {
  echo "A protected or unselected shared file changed. Stop."
  exit 1
}
```

What this does:

- Marks new auto-restored files as intent-to-add so `git diff --stat` includes
  them without staging their contents.
- Shows the required status, diff stat, and file-level change list.
- Checks whitespace errors and confirms no deletion is present.
- Confirms protected and unselected shared files remain unchanged.

## 11. Stage Only The Selected Changes

```bash
if [ -s /tmp/session-management-service-auto-restore.txt ]; then
  git add --pathspec-from-file=/tmp/session-management-service-auto-restore.txt
fi

git add backend/go.work backend/go.work.sum

git status
git diff --cached --stat
git diff --cached --name-status
git diff --cached --check
git diff --cached --name-status --diff-filter=D --exit-code

unexpected_staged="$(
  git diff --cached --name-only -- . \
    ':(exclude)backend/services/session-service' \
    ':(exclude)TaskImplementation/Session Management Service' \
    ':(exclude)backend/go.work' \
    ':(exclude)backend/go.work.sum'
)"

test -z "$unexpected_staged" || {
  printf 'Unexpected staged paths:\n%s\n' "$unexpected_staged"
  exit 1
}

test -z "$(git diff --cached --name-only -- api/master-api.json docs/01-micro-tasks.md)" || {
  echo "A protected or unselected shared file is staged. Stop."
  exit 1
}
```

What this does:

- Stages only the approved Session Management Service files and the two
  manually reviewed Go workspace files.
- Confirms no deletion is staged.
- Rejects any staged file outside the approved paths.
- Confirms `api/master-api.json` and `docs/01-micro-tasks.md` are not staged.

The staged deletion and unexpected-path checks must print nothing.

## 12. Commit

```bash
git commit -m "Merge Session Management Service work into dev"
```

What this does:

- Creates the required selective integration commit from the reviewed staged
  changes only.

## 13. Verify The Integration Commit

```bash
git status --short --branch
git log -1 --format='%s'
git show --stat --oneline HEAD

test -z "$(git status --porcelain)" || { echo "Integration branch is dirty. Stop."; exit 1; }
test "$(git log -1 --format='%s')" = "Merge Session Management Service work into dev" || {
  echo "Unexpected commit message. Stop."
  exit 1
}

git diff-tree --no-renames --no-commit-id --name-status -r \
  --diff-filter=D --exit-code HEAD

test -z "$(git diff-tree --no-renames --no-commit-id --name-only -r HEAD -- api/master-api.json docs/01-micro-tasks.md)" || {
  echo "The commit contains a protected or unselected shared file. Stop."
  exit 1
}

git diff --exit-code HEAD^ HEAD -- api/master-api.json docs/01-micro-tasks.md
```

What this does:

- Confirms the integration branch is clean after the commit.
- Confirms the exact required commit message.
- Confirms the commit contains no deletion.
- Confirms the commit does not modify `api/master-api.json` or
  `docs/01-micro-tasks.md`.

Do not fast-forward `dev` if any check fails.

## 14. Fast-Forward Dev

```bash
git switch dev
test -z "$(git status --porcelain)" || { echo "dev is dirty. Stop."; exit 1; }

git merge --ff-only chore/Session-Management-Service-selective

test "$(git rev-parse dev)" = "$(git rev-parse chore/Session-Management-Service-selective)" || {
  echo "dev did not fast-forward to the integration commit. Stop."
  exit 1
}
```

What this does:

- Moves `dev` to the reviewed selective integration commit only.
- Does not merge the source feature branch.
- Fails safely if `dev` changed and cannot fast-forward.
- Confirms `dev` and the integration branch point to the same commit.

## 15. Final Verification

```bash
git status --short --branch
git log --oneline -1
git show --stat --oneline HEAD

test -z "$(git status --porcelain)" || { echo "dev is dirty after fast-forward. Stop."; exit 1; }
test "$(git log -1 --format='%s')" = "Merge Session Management Service work into dev" || {
  echo "Unexpected final commit message. Stop."
  exit 1
}

git diff-tree --no-renames --no-commit-id --name-status -r \
  --diff-filter=D --exit-code HEAD
git diff --exit-code HEAD^ HEAD -- api/master-api.json docs/01-micro-tasks.md

git diff --exit-code HEAD feature/session-management-service -- \
  backend/services/session-service \
  'TaskImplementation/Session Management Service'

grep -Fq './services/session-service' backend/go.work
test -d backend/services/session-service
test -d 'TaskImplementation/Session Management Service'
```

What this does:

- Confirms `dev` is clean and points to the selective integration commit.
- Confirms the latest commit has the exact required message and no deletions.
- Confirms `api/master-api.json` and `docs/01-micro-tasks.md` were not changed
  by the selective commit.
- Confirms the selected Session Management Service files match the source
  branch.
- Confirms the Session Management Service workspace entry and directories
  exist.

The final deletion and diff checks must print nothing.
