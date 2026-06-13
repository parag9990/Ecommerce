# Merge CMS Service Work Into Dev

Use this runbook to bring CMS Service work from `feature/CMS-service` into
`dev` without merging the whole feature branch and without deleting existing
files from `dev`.

Never merge `feature/CMS-service` directly into `dev`.

This branch has many unrelated deletions and shared-file regressions. Only
added and modified CMS Service related files are applied. Shared files are
handled manually so existing `dev` content remains intact.

## Rules

- Add new CMS Service related files from `feature/CMS-service`.
- If a CMS-owned file already exists on `dev`, update it only when it is
  modified in `feature/CMS-service`.
- If a file exists on `dev` but is deleted or missing on
  `feature/CMS-service`, keep the `dev` file.
- Do not apply deletion diffs.
- Do not modify, restore, stage, or commit `docs/01-micro-tasks.md`.
- Handle shared files manually so existing `dev` content is not removed.
- Commit with this exact message:

```text
Merge CMS Service work into dev
```

Current repository analysis:

- The automatic restore set contains 94 added or modified CMS-owned files
  under `backend/services/cms-service`, `TaskImplementation/CMS Service`, and
  `proto/ecommerce/cms/v1/cms.proto`.
- `backend/go.work` must be updated manually because the feature version
  removes every existing `dev` workspace entry.
- `database/draw.sql` must be updated manually because it is shared.
- Root `buf.gen.yaml` and `buf.yaml` are CMS-supporting additions, but Git can
  detect `buf.yaml` as a rename of `proto/buf.yaml`. Restore only the new root
  files and keep both existing `proto/buf.*` files.
- Do not import `api/master-api.json` or `backend/go.work.sum`. Their endpoint
  differences contain no CMS Service additions and would remove newer `dev`
  content.
- `docs/01-micro-tasks.md` differs on the feature branch but is protected and
  must remain byte-for-byte unchanged.
- With rename detection disabled, the current endpoint comparison contains
  826 deletions. None of them belong in this selective integration.

## 1. Precheck

```bash
set -euo pipefail

git status --short --branch
git rev-parse --verify dev
git rev-parse --verify feature/CMS-service
test -z "$(git status --porcelain)" || { echo "Working tree is dirty. Commit/stash first."; exit 1; }

if git show-ref --verify --quiet refs/heads/chore/CMS-Service-selective; then
  echo "chore/CMS-Service-selective already exists. Stop and inspect it."
  exit 1
fi

git rev-parse dev > /tmp/cms-service-dev-before.txt
git rev-parse dev:docs/01-micro-tasks.md > /tmp/cms-service-protected-doc-before.txt
```

What this does:

- Confirms both branches exist.
- Confirms the current working tree is clean before changing branches.
- Prevents accidental reuse of an existing integration branch.
- Records the original `dev` commit and protected document blob for final
  verification.

## 2. Create A Safe Integration Branch From Dev

```bash
git switch dev
test -z "$(git status --porcelain)" || { echo "dev is dirty. Stop."; exit 1; }

git switch -c chore/CMS-Service-selective
```

What this does:

- Starts from `dev`.
- Creates a temporary branch for the selective integration work.
- Keeps `dev` safe until the final fast-forward step.

## 3. Review The Full Branch Difference

```bash
git diff --name-status --no-renames dev..feature/CMS-service
git diff --name-status --no-renames --diff-filter=D dev..feature/CMS-service
git diff --name-status --no-renames --diff-filter=AM dev..feature/CMS-service

git diff --name-status --no-renames dev...feature/CMS-service
git diff --name-status --no-renames --diff-filter=D dev...feature/CMS-service
```

What this does:

- Shows the complete endpoint difference from `dev` to
  `feature/CMS-service`.
- Uses `--no-renames` so a rename-like result is reviewed as an addition plus
  a deletion. The deletion side must not be applied.
- Shows source-branch changes since the merge base with the three-dot
  comparison.
- Shows deletions separately. Do not apply any of those files.
- Shows added and modified files. Only reviewed CMS-related candidates may be
  selected.

Review the important shared and protected files:

```bash
git diff dev..feature/CMS-service -- backend/go.work
git diff dev..feature/CMS-service -- backend/go.work.sum
git diff dev..feature/CMS-service -- database/draw.sql
git diff dev..feature/CMS-service -- buf.gen.yaml buf.yaml proto/buf.gen.yaml proto/buf.yaml
git diff dev..feature/CMS-service -- api/master-api.json
git diff dev..feature/CMS-service -- docs/01-micro-tasks.md
```

In the current repository state:

- The feature version of `backend/go.work` removes existing service and shared
  module entries. Do not restore it wholesale.
- The feature branch did not change `backend/go.work.sum` from its merge-base
  version, and its endpoint version has no checksum lines missing from `dev`.
  Leave the `dev` file unchanged.
- The `api/master-api.json` endpoint diff only removes or regresses unrelated
  Product and Payment API content. Leave it unchanged.
- The root Buf files are additions, while the source branch also removes
  `proto/buf.gen.yaml` and `proto/buf.yaml`. Add the root files only and keep
  the existing `proto/buf.*` files.
- `docs/01-micro-tasks.md` must not be changed even though the branch updates
  it.

## 4. Build The Auto-Restore File List

```bash
git diff --name-only --no-renames --diff-filter=AM dev..feature/CMS-service -- \
  backend/services/cms-service \
  'TaskImplementation/CMS Service' \
  proto/ecommerce/cms/v1/cms.proto \
  ':(exclude)docs/01-micro-tasks.md' \
  > /tmp/cms-service-auto-restore.txt

test ! -s /tmp/cms-service-auto-restore.txt || \
  ! grep -Fxq 'docs/01-micro-tasks.md' /tmp/cms-service-auto-restore.txt

sed -n '1,240p' /tmp/cms-service-auto-restore.txt
wc -l /tmp/cms-service-auto-restore.txt
```

What this does:

- Lists only added and modified CMS-owned service, task documentation, and
  proto files.
- Disables rename detection so deletion sides cannot enter the restore list.
- Explicitly excludes `docs/01-micro-tasks.md`.
- Excludes shared files that need manual handling.

In the current repository state, this list contains 94 files. If that count
changes, review every listed path before continuing.

## 5. Restore Added And Modified CMS Service Files Using `git restore --overlay`

```bash
if [ -s /tmp/cms-service-auto-restore.txt ]; then
  git restore --overlay \
    --source=feature/CMS-service \
    --worktree \
    --pathspec-from-file=/tmp/cms-service-auto-restore.txt
else
  echo "No CMS Service owned files to auto-restore."
fi
```

What this does:

- Copies selected new CMS Service files from `feature/CMS-service`.
- Updates selected existing CMS Service files only when they differ from
  `dev`.
- Uses `--overlay` so Git does not remove `dev` files.
- Cannot restore `docs/01-micro-tasks.md` because the protected path is not in
  the reviewed restore list.

## 6. Handle Shared Files Manually

### Update `backend/go.work` Without Removing Existing Modules

```bash
(
  cd backend
  go work use ./services/cms-service
)

git diff -- backend/go.work
```

The diff must only add `./services/cms-service`. Existing `dev` entries such as
other services and shared modules must remain.

### Keep `backend/go.work.sum` And `api/master-api.json` Unchanged

```bash
comm -13 \
  <(git show dev:backend/go.work.sum | LC_ALL=C sort -u) \
  <(git show feature/CMS-service:backend/go.work.sum | LC_ALL=C sort -u)

git diff --exit-code -- backend/go.work.sum
git diff --exit-code -- api/master-api.json
```

The `comm` command and both diffs must print nothing. Do not restore either
file from the feature branch.

### Update `database/draw.sql` In Patch Mode

```bash
git diff dev..feature/CMS-service -- database/draw.sql
git restore -p --source=feature/CMS-service --worktree -- database/draw.sql
git diff -- database/draw.sql
```

For the current diff:

- Accept the CMS audit-log schema hunk that adds actor roles, request and trace
  IDs, decision metadata, hashes, and indexes.
- Reject the trailing blank-line-only hunk.
- Confirm no unrelated schema content was removed.

### Add The New Root Buf Files Without Removing `proto/buf.*`

```bash
for file in buf.gen.yaml buf.yaml; do
  if git cat-file -e "dev:$file" 2>/dev/null; then
    echo "$file already exists on dev. Stop and merge it manually."
    exit 1
  fi
done

git restore --overlay --source=feature/CMS-service --worktree -- \
  buf.gen.yaml \
  buf.yaml \
  ':(exclude)docs/01-micro-tasks.md'

git status --short -- buf.gen.yaml buf.yaml
cmp <(git show feature/CMS-service:buf.gen.yaml) buf.gen.yaml
cmp <(git show feature/CMS-service:buf.yaml) buf.yaml
test -f proto/buf.gen.yaml
test -f proto/buf.yaml
git diff --exit-code -- proto/buf.gen.yaml proto/buf.yaml
```

What this does:

- Adds the CMS-supporting root Buf configuration only because those root paths
  do not exist on `dev`.
- Confirms the new root files exactly match their reviewed feature-branch
  versions.
- Keeps `proto/buf.gen.yaml` and `proto/buf.yaml`, even though the feature
  branch deletes them.
- Explicitly excludes the protected micro-tasks document from the restore.

### Confirm Protected And Excluded Files Are Unchanged

```bash
git diff --exit-code -- docs/01-micro-tasks.md
git diff --exit-code -- backend/go.work.sum
git diff --exit-code -- api/master-api.json
git diff --exit-code -- proto/buf.gen.yaml proto/buf.yaml
```

All four commands must print nothing.

## 7. Confirm No Deletions Were Applied

```bash
git diff --name-status --diff-filter=D --exit-code
git diff --exit-code -- docs/01-micro-tasks.md
```

What this does:

- Confirms the working tree has no deletions.
- Confirms the protected micro-tasks document is unchanged.

Both commands must print nothing. If either prints a diff, stop before
staging.

## 8. Build The Selected Change List

```bash
{
  cat /tmp/cms-service-auto-restore.txt
  printf '%s\n' \
    backend/go.work \
    database/draw.sql \
    buf.gen.yaml \
    buf.yaml
} | LC_ALL=C sort -u > /tmp/cms-service-selected.txt

test ! -s /tmp/cms-service-selected.txt || \
  ! grep -Fxq 'docs/01-micro-tasks.md' /tmp/cms-service-selected.txt

sed -n '1,240p' /tmp/cms-service-selected.txt
```

What this does:

- Combines the reviewed auto-restore files with the manually handled shared
  files.
- Keeps `docs/01-micro-tasks.md`, `backend/go.work.sum`,
  `api/master-api.json`, and `proto/buf.*` out of the selected list.
- Creates the exact allowlist used for intent-to-add, staging, and commit.

## 9. Show Status And Diff Stat Before Staging

```bash
git add -N --pathspec-from-file=/tmp/cms-service-selected.txt

git status
git diff --stat
git diff --name-status
git diff --name-status --diff-filter=D --exit-code
git diff --exit-code -- docs/01-micro-tasks.md

git diff --name-only | LC_ALL=C sort -u > /tmp/cms-service-worktree-files.txt
comm -23 \
  /tmp/cms-service-worktree-files.txt \
  /tmp/cms-service-selected.txt \
  > /tmp/cms-service-unexpected-worktree-files.txt
test ! -s /tmp/cms-service-unexpected-worktree-files.txt || {
  echo "Unexpected working-tree files:"
  cat /tmp/cms-service-unexpected-worktree-files.txt
  exit 1
}
```

What this does:

- Marks new selected files as intent-to-add so `git diff --stat` shows them.
- Shows status, diff stat, and file status before full staging.
- Confirms no deletion or protected-document change is present.
- Stops if any changed tracked or intent-to-add file is outside the selected
  allowlist.

## 10. Stage Only The Selected Changes

```bash
git add --pathspec-from-file=/tmp/cms-service-selected.txt

git status
git diff --cached --stat
git diff --cached --name-status
git diff --cached --name-status --diff-filter=D --exit-code
git diff --cached --check

test -z "$(git diff --cached --name-only -- docs/01-micro-tasks.md)" || {
  echo "Protected docs/01-micro-tasks.md is staged. Stop."
  exit 1
}

git diff --cached --name-only | LC_ALL=C sort -u > /tmp/cms-service-staged-files.txt
comm -3 \
  /tmp/cms-service-selected.txt \
  /tmp/cms-service-staged-files.txt \
  > /tmp/cms-service-staged-mismatch.txt
test ! -s /tmp/cms-service-staged-mismatch.txt || {
  echo "Staged files do not exactly match the selected allowlist:"
  cat /tmp/cms-service-staged-mismatch.txt
  exit 1
}
```

What this does:

- Stages only the reviewed selected paths.
- Confirms no deletion is staged.
- Confirms `docs/01-micro-tasks.md` is not staged.
- Confirms the staged file set exactly matches the selected allowlist.

The staged deletion and protected-document checks must print nothing.

## 11. Commit With The Exact Message

```bash
git commit --dry-run --only \
  --pathspec-from-file=/tmp/cms-service-selected.txt \
  -m "Merge CMS Service work into dev"

git commit --only \
  --pathspec-from-file=/tmp/cms-service-selected.txt \
  -m "Merge CMS Service work into dev"
```

What this does:

- Previews and creates a commit containing only paths from the selected
  allowlist.
- Excludes `docs/01-micro-tasks.md` from the commit because it is not in that
  allowlist.
- Creates the selective integration commit with the exact required message.

## 12. Verify After Commit

```bash
git status --short --branch
git log --oneline -1
git show --stat --oneline HEAD
git show --name-status --format= HEAD

test "$(git log -1 --pretty=%s)" = "Merge CMS Service work into dev"
test -z "$(git diff-tree --no-commit-id --name-status --diff-filter=D -r HEAD)"
test -z "$(git diff-tree --no-commit-id --name-only -r HEAD -- docs/01-micro-tasks.md)"
test "$(git rev-parse HEAD:docs/01-micro-tasks.md)" = \
  "$(cat /tmp/cms-service-protected-doc-before.txt)"
test "$(git rev-parse HEAD^)" = "$(cat /tmp/cms-service-dev-before.txt)"
```

What this does:

- Confirms the commit message is exact.
- Confirms the commit contains no deletions.
- Confirms the protected document is absent from the commit and its blob is
  unchanged.
- Confirms the selective commit has the original `dev` tip as its parent.

## 13. Fast-Forward Dev

```bash
git switch dev
git merge --ff-only chore/CMS-Service-selective
```

What this does:

- Moves `dev` to the selective integration commit.
- Does not merge `feature/CMS-service`.
- Fails safely if `dev` changed and cannot fast-forward.

## 14. Final Verification

```bash
git status --short --branch
git log --oneline -1
git show --stat --oneline HEAD
git show --name-status --format= HEAD

test "$(git branch --show-current)" = "dev"
test "$(git log -1 --pretty=%s)" = "Merge CMS Service work into dev"
test -z "$(git diff-tree --no-commit-id --name-status --diff-filter=D -r HEAD)"
test -z "$(git diff-tree --no-commit-id --name-only -r HEAD -- docs/01-micro-tasks.md)"
test "$(git rev-parse HEAD:docs/01-micro-tasks.md)" = \
  "$(cat /tmp/cms-service-protected-doc-before.txt)"
test "$(git rev-parse HEAD^)" = "$(cat /tmp/cms-service-dev-before.txt)"
git merge-base --is-ancestor chore/CMS-Service-selective dev
```

What this does:

- Confirms `dev` is on the new selective integration commit.
- Confirms the latest commit message is exact.
- Confirms no deletion was committed.
- Confirms `docs/01-micro-tasks.md` is unchanged and absent from the commit.
- Confirms `dev` fast-forwarded from its original tip to the integration
  branch commit.
