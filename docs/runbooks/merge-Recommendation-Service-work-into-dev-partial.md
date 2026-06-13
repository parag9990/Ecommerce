# Merge Recommendation Service Work Into Dev (Partial)

Use this runbook to bring Recommendation Service work from
`feature/recommendation-service` into `dev` without merging the whole feature
branch, without deleting existing files from `dev`, and without modifying
`docs/01-micro-tasks.md`.

Never merge `feature/recommendation-service` directly. This branch has many
unrelated deletions and destructive shared-file changes. Only restore reviewed
added and modified Recommendation Service files.

## Rules

- Add new Recommendation Service files and directly required support files from
  `feature/recommendation-service`.
- Do not apply deletion diffs.
- Use `--no-renames` while building the restore list. Git otherwise detects
  some required new files as renames from unrelated deleted service files.
- Use `git restore --overlay` so files already on `dev` are not removed.
- Handle existing shared files manually so their `dev` content is preserved.
- Do not modify, restore, stage, or commit `docs/01-micro-tasks.md`.
- Commit with this exact message:

```text
Merge Recommendation Service work into dev
```

## 1. Precheck

```bash
set -euo pipefail

git status --short --branch
git rev-parse --verify dev
git rev-parse --verify feature/recommendation-service
test -z "$(git status --porcelain)" || { echo "Working tree is dirty. Commit/stash first."; exit 1; }
```

What this does:

- Confirms both branches exist.
- Confirms the current working tree is clean before changing branches.
- Protects unrelated local work, including any local change to
  `docs/01-micro-tasks.md`.

## 2. Create A Safe Integration Branch From Dev

```bash
git switch dev
test -z "$(git status --porcelain)" || { echo "dev is dirty. Stop."; exit 1; }

git switch -c chore/Recommendation-Service-selective
```

What this does:

- Starts from `dev`.
- Creates a temporary branch for the selective integration work.
- Keeps `dev` unchanged until the final fast-forward step.

## 3. Review The Full Branch Difference

```bash
git diff --name-status --no-renames dev..feature/recommendation-service
git diff --name-status --no-renames --diff-filter=D dev..feature/recommendation-service
git diff --name-status --no-renames --diff-filter=AM dev..feature/recommendation-service
```

What this does:

- Shows the complete difference from `dev` to
  `feature/recommendation-service`.
- Shows deletions separately. Do not apply any of those files.
- Shows added and modified files as the only possible candidates.
- Disables rename detection so required Recommendation Service destination
  files are not hidden as renames from unrelated deleted files.

Review the existing shared files separately:

```bash
git diff dev..feature/recommendation-service -- \
  api/master-api.json \
  backend/go.work \
  backend/go.work.sum \
  docs/01-micro-tasks.md \
  proto/buf.gen.yaml \
  proto/buf.yaml
```

In the current repo state:

- The branch contains many unrelated deletions, including existing services,
  shared modules, task documents, protos, reports, and a runbook.
- `backend/go.work` removes existing `dev` workspace modules while adding
  `./services/recommendation-service`. Add only that new entry manually.
- `backend/go.work.sum` only removes existing `dev` checksum lines. Keep the
  `dev` version unchanged.
- `api/master-api.json` only contains unrelated Product and Payment API
  removals/changes. Keep the `dev` version unchanged.
- `proto/buf.yaml` removes existing `dev` configuration, and
  `proto/buf.gen.yaml` replaces the existing shared generation strategy. Keep
  both `dev` versions unchanged.
- `docs/01-micro-tasks.md` contains broad unrelated status changes and is
  protected by this runbook. Keep it unchanged.

## 4. Build The Auto-Restore File List

```bash
git diff --no-renames --name-only --diff-filter=AM \
  dev..feature/recommendation-service -- \
  backend/services/recommendation-service \
  'TaskImplementation/Recommendation Service' \
  proto/ecommerce/recommendation/v1 \
  backend/proto-gen/go/ecommerce/recommendation/v1 \
  backend/proto-gen/go/go.mod \
  backend/proto-gen/go/go.sum \
  > /tmp/recommendation-service-auto-restore.txt

if grep -Fxq 'docs/01-micro-tasks.md' /tmp/recommendation-service-auto-restore.txt; then
  echo "Protected docs/01-micro-tasks.md entered the restore list. Stop."
  exit 1
fi

wc -l /tmp/recommendation-service-auto-restore.txt
sed -n '1,240p' /tmp/recommendation-service-auto-restore.txt
```

What this does:

- Lists only added and modified Recommendation Service files and required
  Recommendation proto/generated-module support files.
- Excludes all deleted files and all modified shared files.
- Explicitly stops if the protected micro-task file enters the allowlist.
- Uses `--no-renames` so these required destination files are included:
  `backend/services/recommendation-service/internal/transport/http/routes.go`
  and `backend/proto-gen/go/go.sum`.

The current reviewed branch difference produces 103 auto-restore candidates.
If the count or paths change, review the new list before continuing.

## 5. Restore Added And Modified Recommendation Service Files

```bash
if [ -s /tmp/recommendation-service-auto-restore.txt ]; then
  git restore --overlay \
    --source=feature/recommendation-service \
    --worktree \
    --pathspec-from-file=/tmp/recommendation-service-auto-restore.txt
else
  echo "No Recommendation Service files to auto-restore."
fi
```

What this does:

- Copies only files from the validated Recommendation Service allowlist.
- Uses `--overlay` so Git does not remove files from `dev`.
- Does not restore any modified shared file or
  `docs/01-micro-tasks.md`.

## 6. Handle Shared Files Manually

Add only the Recommendation Service workspace entry:

```bash
(
  cd backend
  go work use ./services/recommendation-service
)

git diff -- backend/go.work

if git diff --unified=0 -- backend/go.work | grep -q '^-[^-]'; then
  echo "backend/go.work contains removed dev lines. Stop."
  exit 1
fi
```

The `backend/go.work` diff must add
`./services/recommendation-service` without removing any existing `dev`
workspace entry.

Keep all other existing shared files unchanged:

```bash
git diff --quiet -- \
  api/master-api.json \
  backend/go.work.sum \
  docs/01-micro-tasks.md \
  proto/buf.gen.yaml \
  proto/buf.yaml \
  || { echo "A protected/shared file changed. Stop and review."; exit 1; }
```

Why these files stay unchanged:

- The feature version of `backend/go.work.sum` has no feature-only lines to
  add; it only removes existing checksums.
- The feature changes to `api/master-api.json` are unrelated to Recommendation
  Service.
- The feature changes to the Buf configs remove or replace existing `dev`
  configuration rather than adding a safe Recommendation-only hunk.
- `docs/01-micro-tasks.md` must not be modified.

The added `backend/proto-gen/go/go.mod`, `backend/proto-gen/go/go.sum`, and
Recommendation generated files are already included in the validated
auto-restore list because the Recommendation Service imports that module.

## 7. Confirm No Deletions Were Applied

```bash
git diff --name-status --diff-filter=D --exit-code
git diff --check
```

What this does:

- Confirms the working tree has no file deletions.
- Checks the selected changes for whitespace errors.

The deletion check must print nothing. If it prints any `D` entry, stop before
staging.

## 8. Show Status And Diff Stat Before Staging

```bash
if [ -s /tmp/recommendation-service-auto-restore.txt ]; then
  git add -N --pathspec-from-file=/tmp/recommendation-service-auto-restore.txt
fi

git status
git diff --stat
git diff --name-status --diff-filter=D --exit-code
git diff --quiet -- docs/01-micro-tasks.md \
  || { echo "docs/01-micro-tasks.md changed. Stop."; exit 1; }
```

What this does:

- Marks only allowlisted new files as intent-to-add so `git diff --stat` shows
  them before full staging.
- Shows the required status and diff stat.
- Confirms again that no deletion or protected micro-task change is present.

## 9. Stage Only The Selected Changes

```bash
if [ -s /tmp/recommendation-service-auto-restore.txt ]; then
  git add --pathspec-from-file=/tmp/recommendation-service-auto-restore.txt
fi

git add backend/go.work

git status
git diff --cached --stat
git diff --cached --name-status
git diff --cached --name-status --diff-filter=D --exit-code
git diff --cached --quiet -- docs/01-micro-tasks.md \
  || { echo "docs/01-micro-tasks.md is staged. Stop."; exit 1; }
```

What this does:

- Stages only the validated Recommendation Service allowlist and the manually
  reviewed additive `backend/go.work` change.
- Does not stage any other shared file.
- Confirms no deletion and no protected micro-task change is staged.

Review `git diff --cached --name-status` carefully. It must contain only
Recommendation Service/support additions and the `backend/go.work`
modification.

## 10. Commit

Run the safety checks again immediately before committing:

```bash
git diff --cached --name-status --diff-filter=D --exit-code
git diff --cached --quiet -- docs/01-micro-tasks.md \
  || { echo "docs/01-micro-tasks.md is staged. Stop."; exit 1; }

git commit -m "Merge Recommendation Service work into dev"
```

What this does:

- Stops if a deletion or protected micro-task change is staged.
- Creates the required selective integration commit with the exact message.

## 11. Verify After Commit

```bash
git status --short --branch
test "$(git log -1 --format=%s)" = "Merge Recommendation Service work into dev"
git show --stat --oneline HEAD
git diff --name-status --diff-filter=D --exit-code HEAD^ HEAD
git diff --quiet HEAD^ HEAD -- docs/01-micro-tasks.md \
  || { echo "The commit changed docs/01-micro-tasks.md. Stop."; exit 1; }
```

What this does:

- Confirms the integration branch is clean.
- Confirms the exact commit message.
- Shows the committed file summary.
- Confirms the commit contains no deletions and no protected micro-task change.

## 12. Fast-Forward Dev

```bash
git switch dev
git merge --ff-only chore/Recommendation-Service-selective
```

What this does:

- Moves `dev` to the reviewed selective integration commit.
- Does not merge `feature/recommendation-service`.
- Fails safely if `dev` changed and cannot fast-forward.

## 13. Final Verification

```bash
git status --short --branch
test "$(git rev-parse dev)" = "$(git rev-parse chore/Recommendation-Service-selective)"
test "$(git log -1 --format=%s)" = "Merge Recommendation Service work into dev"
git show --stat --oneline HEAD
git diff --name-status --diff-filter=D --exit-code HEAD^ HEAD
git diff --quiet HEAD^ HEAD -- docs/01-micro-tasks.md \
  || { echo "Final commit changed docs/01-micro-tasks.md. Stop."; exit 1; }
```

What this does:

- Confirms `dev` exactly matches the reviewed integration branch.
- Confirms the latest commit message and final committed file summary.
- Confirms the fast-forwarded commit has no deletions.
- Confirms `docs/01-micro-tasks.md` was not changed.
