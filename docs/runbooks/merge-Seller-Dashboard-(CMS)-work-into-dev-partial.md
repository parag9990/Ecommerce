# Merge Seller Dashboard (CMS) Work Into Dev

Use this runbook to bring Seller Dashboard (CMS) work from
`feature/seller-dashboard-cms` into `dev` without merging the whole feature
branch and without deleting existing files from `dev`.

Never merge the source branch directly. This branch has many unrelated
deletions. We only apply added and modified Seller Dashboard (CMS) related
files, and we handle shared files carefully.

## Rules

- Add new Seller Dashboard (CMS) files from `feature/seller-dashboard-cms`.
- If a Seller Dashboard (CMS) file already exists on `dev`, update it only when
  it is modified in `feature/seller-dashboard-cms`.
- If a file exists on `dev` but is deleted or missing on
  `feature/seller-dashboard-cms`, keep the `dev` file.
- Do not apply deletion diffs.
- Do not merge `feature/seller-dashboard-cms` directly.
- Do not modify `docs/01-micro-tasks.md`.
- Exclude `docs/01-micro-tasks.md` from all restore, staging, and commit
  commands.
- Commit with this exact message:

```text
Merge Seller Dashboard (CMS) work into dev
```

## 1. Precheck

```bash
set -euo pipefail

git status --short --branch
git rev-parse --verify dev
git rev-parse --verify feature/seller-dashboard-cms
git rev-parse --verify 'chore/Seller-Dashboard-(CMS)-selective' >/dev/null 2>&1 && {
  echo "Integration branch already exists. Stop or choose a new branch name."
  exit 1
} || true
test -z "$(git status --porcelain)" || { echo "Working tree is dirty. Commit/stash first."; exit 1; }
```

What this does:

- Confirms both branches exist.
- Confirms the integration branch does not already exist.
- Confirms the current working tree is clean before changing branches.

## 2. Create A Safe Integration Branch From `dev`

```bash
git switch dev
test -z "$(git status --porcelain)" || { echo "dev is dirty. Stop."; exit 1; }

git switch -c 'chore/Seller-Dashboard-(CMS)-selective'
```

What this does:

- Starts from `dev`.
- Creates a temporary branch for the selective merge work.
- Keeps `dev` safe until the final fast-forward step.

## 3. Review The Full Branch Difference

```bash
git diff --name-status dev..feature/seller-dashboard-cms
git diff --name-status --diff-filter=D dev..feature/seller-dashboard-cms
git diff --name-status --diff-filter=AM dev..feature/seller-dashboard-cms
```

What this does:

- Shows the complete difference from `dev` to `feature/seller-dashboard-cms`.
- Shows deletions separately. Do not apply those files.
- Shows only added and modified files. These are the only candidates for this
  runbook.

Current important shared files to check:

```bash
git diff dev..feature/seller-dashboard-cms -- \
  api/master-api.json \
  backend/go.work \
  backend/go.work.sum \
  frontend/package.json \
  frontend/pnpm-lock.yaml \
  frontend/pnpm-workspace.yaml \
  docs/01-micro-tasks.md
```

In the current repo state, the Seller Dashboard (CMS) branch:

- Adds `frontend/seller-dashboard`.
- Adds `TaskImplementation/Seller Dashboard (CMS)`.
- Modifies shared frontend workspace files.
- Modifies `api/master-api.json`, `backend/go.work`, `backend/go.work.sum`, and
  `docs/01-micro-tasks.md`.
- Contains many unrelated deletions from other services and frontend modules.

Do not restore shared files wholesale. Do not touch `docs/01-micro-tasks.md`.

## 4. Build The Auto-Restore File List

```bash
git diff --name-only --diff-filter=AM dev..feature/seller-dashboard-cms -- \
  frontend/seller-dashboard \
  'TaskImplementation/Seller Dashboard (CMS)' \
  > /tmp/seller-dashboard-cms-auto-restore.txt

sed -n '1,260p' /tmp/seller-dashboard-cms-auto-restore.txt

grep -Fx 'docs/01-micro-tasks.md' /tmp/seller-dashboard-cms-auto-restore.txt && {
  echo "Protected docs/01-micro-tasks.md was included. Stop."
  exit 1
} || true
```

What this does:

- Lists Seller Dashboard (CMS) files that are added or modified.
- Excludes deleted files.
- Excludes protected shared files such as `frontend/package.json`,
  `frontend/pnpm-lock.yaml`, `frontend/pnpm-workspace.yaml`,
  `api/master-api.json`, `backend/go.work`, `backend/go.work.sum`, and
  `docs/01-micro-tasks.md`.

If this file is empty, stop and review the path filters before continuing.

## 5. Restore Added And Modified Service Files Using `git restore --overlay`

```bash
git restore --overlay --source=feature/seller-dashboard-cms --worktree --pathspec-from-file=/tmp/seller-dashboard-cms-auto-restore.txt
```

What this does:

- Copies selected new Seller Dashboard (CMS) files from
  `feature/seller-dashboard-cms`.
- Updates selected existing Seller Dashboard (CMS) files only when they differ
  from `dev`.
- Uses `--overlay` so Git does not remove existing `dev` files.

## 6. Handle Shared Files Manually

First review the shared-file diffs again:

```bash
git diff dev..feature/seller-dashboard-cms -- \
  frontend/package.json \
  frontend/pnpm-workspace.yaml \
  frontend/pnpm-lock.yaml \
  api/master-api.json \
  backend/go.work \
  backend/go.work.sum \
  docs/01-micro-tasks.md
```

### `frontend/pnpm-workspace.yaml`

Add only the Seller Dashboard workspace entry. Keep the existing `user-app` and
`packages/*` entries.

Expected result:

```yaml
packages:
  - "user-app"
  - "seller-dashboard"
  - "packages/*"
```

Verify:

```bash
git diff -- frontend/pnpm-workspace.yaml
```

### `frontend/package.json`

Do not restore the feature branch version of this file. The feature version
removes existing `dev` metadata and user-app scripts. Add only the Seller
Dashboard scripts while keeping the current `packageManager`, `engines`, and
existing user-app scripts.

Use a targeted edit:

```bash
node <<'NODE'
const fs = require('fs');
const file = 'frontend/package.json';
const pkg = JSON.parse(fs.readFileSync(file, 'utf8'));
pkg.scripts = pkg.scripts || {};
pkg.scripts['build:seller'] = 'pnpm --filter seller-dashboard build';
pkg.scripts['test:seller'] = 'pnpm --filter seller-dashboard test';
pkg.scripts['typecheck:seller'] = 'pnpm --filter seller-dashboard typecheck';
fs.writeFileSync(file, JSON.stringify(pkg, null, 2) + '\n');
NODE

git diff -- frontend/package.json
```

What this does:

- Adds the seller scripts introduced by the feature branch.
- Keeps existing `dev`, `dev:user`, `build:user`, `lint:user`,
  `typecheck:user`, and `preview:user` scripts.
- Avoids replacing root frontend metadata with the seller-only feature version.

### `frontend/pnpm-lock.yaml`

Do not restore the feature branch lockfile wholesale. The feature lockfile
removes existing `user-app` and `packages/proto-client` importer entries.

After `frontend/seller-dashboard`, `frontend/package.json`, and
`frontend/pnpm-workspace.yaml` are in place, regenerate the lockfile from the
merged workspace:

```bash
(
  cd frontend
  pnpm install --lockfile-only
)

git diff -- frontend/pnpm-lock.yaml
grep -n '^  user-app:' frontend/pnpm-lock.yaml
grep -n '^  packages/proto-client:' frontend/pnpm-lock.yaml
grep -n '^  seller-dashboard:' frontend/pnpm-lock.yaml
```

What this does:

- Adds the Seller Dashboard importer and dependency resolution.
- Keeps existing `dev` lockfile content for `user-app` and
  `packages/proto-client`.
- Avoids applying the feature branch lockfile deletions.

### `api/master-api.json`

In the current branch diff, this file does not add Seller Dashboard (CMS)
frontend files or routes. The diff mostly removes unrelated admin/session/search
API entries and weakens existing schema details.

Do not restore or stage this file by default.

Verify it is unchanged:

```bash
git diff -- api/master-api.json --exit-code
```

If a future version of this branch adds a clearly Seller Dashboard specific API
entry, add only that JSON entry manually and re-run the diff. Do not remove
existing `dev` API entries.

### `backend/go.work` And `backend/go.work.sum`

Seller Dashboard (CMS) is a frontend module in this branch. The feature version
of `backend/go.work` removes existing `dev` workspace entries, and
`backend/go.work.sum` is deletion-only in the current diff.

Do not restore or stage these files.

Verify they are unchanged:

```bash
git diff -- backend/go.work backend/go.work.sum --exit-code
```

### `docs/01-micro-tasks.md`

This file is protected for this runbook.

Do not restore, patch, edit, stage, or commit it.

Verify it is unchanged:

```bash
git diff -- docs/01-micro-tasks.md --exit-code
```

## 7. Confirm No Deletions Were Applied

```bash
git diff --name-status --diff-filter=D --exit-code
git diff -- docs/01-micro-tasks.md --exit-code
```

What this does:

- Confirms the working tree has no deletions.
- Confirms `docs/01-micro-tasks.md` was not changed.
- Both commands must print nothing.

If either command prints output, stop before staging.

## 8. Show Status And Diff Stat Before Staging

```bash
git add -N --pathspec-from-file=/tmp/seller-dashboard-cms-auto-restore.txt
git add -N frontend/package.json frontend/pnpm-workspace.yaml frontend/pnpm-lock.yaml

git status
git diff --stat
git diff --name-status
git diff --name-status --diff-filter=D --exit-code
git diff -- docs/01-micro-tasks.md --exit-code
```

What this does:

- Marks new auto-restored files as intent-to-add so `git diff --stat` shows
  them.
- Shows the required `git status`.
- Shows the required `git diff --stat`.
- Confirms again that no deletion is present.
- Confirms again that `docs/01-micro-tasks.md` is untouched.

## 9. Stage Only The Selected Changes

```bash
git add --pathspec-from-file=/tmp/seller-dashboard-cms-auto-restore.txt
git add frontend/package.json frontend/pnpm-workspace.yaml frontend/pnpm-lock.yaml

git status
git diff --cached --stat
git diff --cached --name-status
git diff --cached --name-status --diff-filter=D --exit-code
git diff --cached -- docs/01-micro-tasks.md --exit-code

git diff --cached --name-only | grep -Fx 'docs/01-micro-tasks.md' && {
  echo "Protected docs/01-micro-tasks.md is staged. Stop."
  exit 1
} || true
```

What this does:

- Stages selected Seller Dashboard (CMS) added and modified files.
- Stages only the manually reviewed frontend workspace/package/lockfile changes.
- Does not stage `api/master-api.json`, `backend/go.work`, `backend/go.work.sum`,
  or `docs/01-micro-tasks.md`.
- Confirms no deletions are staged.

The staged deletion check must print nothing.

## 10. Commit With The Exact Commit Message

```bash
git commit -m "Merge Seller Dashboard (CMS) work into dev"

git status --short --branch
git log --oneline -1
git show --stat --oneline HEAD
git show --name-status --format='' HEAD | grep '^D' && {
  echo "Commit contains deletions. Stop."
  exit 1
} || true
git show --name-only --format='' HEAD | grep -Fx 'docs/01-micro-tasks.md' && {
  echo "Commit contains protected docs/01-micro-tasks.md. Stop."
  exit 1
} || true
```

What this does:

- Creates the required selective integration commit.
- Verifies the latest commit summary.
- Verifies the commit contains no deletions.
- Verifies the commit does not include `docs/01-micro-tasks.md`.

The latest commit message must be:

```text
Merge Seller Dashboard (CMS) work into dev
```

## 11. Fast-Forward `dev`

```bash
git switch dev
git merge --ff-only 'chore/Seller-Dashboard-(CMS)-selective'
```

What this does:

- Moves `dev` to the selective integration commit.
- Does not merge `feature/seller-dashboard-cms`.
- Fails safely if `dev` changed and cannot fast-forward.

## 12. Final Verification

```bash
git status --short --branch
git log --oneline -1
git show --stat --oneline HEAD
git show --name-status --format='' HEAD | grep '^D' && {
  echo "Final dev commit contains deletions. Stop."
  exit 1
} || true
git show --name-only --format='' HEAD | grep -Fx 'docs/01-micro-tasks.md' && {
  echo "Final dev commit contains protected docs/01-micro-tasks.md. Stop."
  exit 1
} || true
git diff HEAD^..HEAD -- docs/01-micro-tasks.md --exit-code
```

What this does:

- Confirms `dev` is on the new selective integration commit.
- Shows the final committed file summary.
- Confirms the final commit has no deletions.
- Confirms `docs/01-micro-tasks.md` was not committed.
- Confirms the latest commit message is:

```text
Merge Seller Dashboard (CMS) work into dev
```

Recommended frontend checks after the fast-forward:

```bash
(
  cd frontend
  pnpm --filter seller-dashboard typecheck
  pnpm --filter seller-dashboard test
)
```
