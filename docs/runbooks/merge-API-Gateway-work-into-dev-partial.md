# Merge API Gateway Work Into Dev

Use this runbook to bring API Gateway work from
`feature/api-gateway-service` into `dev` without merging the whole feature
branch, without deleting existing files from `dev`, and without modifying
`docs/01-micro-tasks.md`.

Do not run:

```bash
git merge feature/api-gateway-service
```

The source branch has unrelated deletions and also deletes older API Gateway
files that must remain on `dev` under this no-deletions workflow. Only added
and modified API Gateway files are restored. Shared files are handled
individually so newer `dev` content is not removed.

## Rules

- Add or update only API Gateway related files from
  `feature/api-gateway-service`.
- If a file exists on `dev` but is deleted or missing on the source branch,
  keep the `dev` file.
- Do not apply or stage any deletion diff.
- Do not modify, restore, stage, or commit `docs/01-micro-tasks.md`.
- Do not apply the source-branch changes to `api/master-api.json`; the current
  diff removes unrelated Product and Payment API contract content.
- Do not restore or stage `backend/services/api-gateway/server`; it is a
  tracked 29.9 MB compiled Go executable, not source code.
- Handle `backend/go.work`, `backend/go.work.sum`, and
  `infra/envoy/grpc-web/envoy.yaml` as shared files.
- Commit with this exact message:

```text
Merge API Gateway work into dev
```

## 1. Precheck

```bash
set -euo pipefail

git status --short --branch
git rev-parse --verify dev
git rev-parse --verify feature/api-gateway-service
test -z "$(git status --porcelain)" || { echo "Working tree is dirty. Commit/stash first."; exit 1; }

if git show-ref --verify --quiet refs/heads/chore/API-Gateway-selective; then
  echo "Integration branch already exists. Stop and inspect it."
  exit 1
fi
```

What this does:

- Confirms the target and source branches exist.
- Confirms the current working tree is clean.
- Prevents accidentally reusing an existing integration branch.

## 2. Create A Safe Integration Branch From Dev

```bash
git switch dev
test -z "$(git status --porcelain)" || { echo "dev is dirty. Stop."; exit 1; }

git switch -c chore/API-Gateway-selective
```

What this does:

- Starts from `dev`.
- Creates a temporary branch for the selective integration.
- Keeps `dev` unchanged until the final fast-forward step.

## 3. Review The Full Branch Difference

```bash
git diff --name-status dev..feature/api-gateway-service
git diff --name-status --diff-filter=D dev..feature/api-gateway-service
git diff --name-status --diff-filter=AM dev..feature/api-gateway-service
```

What this does:

- Shows the complete source state compared with `dev`.
- Shows deletions separately. Do not apply any of them.
- Shows added and modified files, which are the only restore candidates.

Review the current shared and protected file differences:

```bash
git diff dev..feature/api-gateway-service -- backend/go.work
git diff dev..feature/api-gateway-service -- backend/go.work.sum
git diff dev..feature/api-gateway-service -- infra/envoy/grpc-web/envoy.yaml
git diff dev..feature/api-gateway-service -- api/master-api.json
git diff dev..feature/api-gateway-service -- docs/01-micro-tasks.md
```

In the current repository state:

- The source `backend/go.work` keeps only `./services/api-gateway`, which would
  remove all other `dev` workspace entries. Never restore that whole file.
- The source `backend/go.work.sum` adds API Gateway sums but removes many sums
  already present on `dev`. Merge it by taking the union of both files.
- `infra/envoy/grpc-web/envoy.yaml` is a new API Gateway related shared
  infrastructure file. Review it explicitly before restoring it.
- `api/master-api.json` removes unrelated Product and Payment contract
  content. Leave it unchanged.
- `docs/01-micro-tasks.md` differs, but it is protected by this runbook and
  must remain unchanged.
- The source branch deletes 16 older files inside
  `backend/services/api-gateway`. Those files must remain on `dev`.

## 4. Build The Auto-Restore File List

```bash
git diff --name-only --diff-filter=AM dev..feature/api-gateway-service -- \
  backend/services/api-gateway \
  'TaskImplementation/API Gateway Service' \
  ':(exclude)backend/services/api-gateway/server' \
  > /tmp/api-gateway-auto-restore.txt

sed -n '1,240p' /tmp/api-gateway-auto-restore.txt

grep -Fxq 'docs/01-micro-tasks.md' /tmp/api-gateway-auto-restore.txt && {
  echo "Protected micro-task file entered the restore list. Stop."
  exit 1
} || true

grep -Fxq 'backend/services/api-gateway/server' /tmp/api-gateway-auto-restore.txt && {
  echo "Compiled executable entered the restore list. Stop."
  exit 1
} || true
```

What this does:

- Lists only added and modified API Gateway source and task documentation
  files.
- Excludes every deletion.
- Excludes shared files for manual handling.
- Excludes `docs/01-micro-tasks.md` and the compiled API Gateway executable.

In the current repository state, this list contains 129 files. If the count
changes, review the list before continuing:

```bash
wc -l /tmp/api-gateway-auto-restore.txt
```

## 5. Restore Added And Modified API Gateway Files Using `git restore --overlay`

```bash
if [ -s /tmp/api-gateway-auto-restore.txt ]; then
  git restore --overlay \
    --source=feature/api-gateway-service \
    --worktree \
    --pathspec-from-file=/tmp/api-gateway-auto-restore.txt
else
  echo "No API Gateway source or task documentation files to auto-restore."
fi
```

What this does:

- Copies only the selected added and modified files from the source branch.
- Uses `--overlay` so Git does not remove files that exist on `dev`.
- Does not touch shared files, protected files, or source-side deletions.

## 6. Handle Shared Files Manually

### Keep `backend/go.work` Entries

```bash
(
  cd backend
  go work use ./services/api-gateway
)

git diff -- backend/go.work
```

What this does:

- Ensures the API Gateway module is in the workspace.
- Keeps every existing `dev` workspace entry.
- Avoids replacing `backend/go.work` with the source version.

The API Gateway entry already exists on `dev` in the current repository state,
so this command should normally produce no `backend/go.work` diff.

### Merge `backend/go.work.sum` Without Removing Existing Sums

```bash
git show dev:backend/go.work.sum > /tmp/dev-go.work.sum
git show feature/api-gateway-service:backend/go.work.sum > /tmp/feature-api-gateway-go.work.sum
LC_ALL=C sort -u \
  /tmp/dev-go.work.sum \
  /tmp/feature-api-gateway-go.work.sum \
  > backend/go.work.sum

git diff -- backend/go.work.sum
```

What this does:

- Keeps every checksum line already present on `dev`.
- Adds checksum lines present on the API Gateway source branch.
- Prevents the source version from removing existing workspace sums.

### Review And Restore The New Envoy Configuration

```bash
if git cat-file -e dev:infra/envoy/grpc-web/envoy.yaml 2>/dev/null; then
  echo "Envoy config now exists on dev. Stop and reconcile it manually."
  exit 1
fi

git show feature/api-gateway-service:infra/envoy/grpc-web/envoy.yaml | sed -n '1,240p'

git restore --overlay \
  --source=feature/api-gateway-service \
  --worktree \
  -- infra/envoy/grpc-web/envoy.yaml

git diff -- infra/envoy/grpc-web/envoy.yaml
```

What this does:

- Stops if `dev` has gained its own Envoy configuration since this runbook was
  written.
- Shows the source file for review before restoring it.
- Restores the file only because it is currently new and API Gateway related.

### Confirm Protected And Unrelated Shared Files Are Untouched

```bash
git diff --quiet -- docs/01-micro-tasks.md || {
  echo "docs/01-micro-tasks.md changed. Stop; do not continue."
  exit 1
}

git diff --quiet -- api/master-api.json || {
  echo "api/master-api.json changed. Stop; do not continue."
  exit 1
}

test -z "$(git status --porcelain -- backend/services/api-gateway/server)" || {
  echo "Compiled API Gateway executable changed or was restored. Stop."
  exit 1
}
```

## 7. Confirm No Deletions Were Applied

```bash
git diff --name-status --diff-filter=D --exit-code
```

What this does:

- Confirms the working tree has no deletions.
- This command must print nothing and exit successfully.

If it prints any `D` entry, stop before staging.

## 8. Show Status And Diff Stat Before Staging

Run the API Gateway tests and code-focused whitespace check first:

```bash
(
  cd backend/services/api-gateway
  go test ./...
)

git diff --check -- \
  backend/services/api-gateway \
  backend/go.work \
  backend/go.work.sum \
  infra/envoy/grpc-web/envoy.yaml
```

Mark new selected files as intent-to-add so the unstaged diff summary includes
them:

```bash
if [ -s /tmp/api-gateway-auto-restore.txt ]; then
  git add -N --pathspec-from-file=/tmp/api-gateway-auto-restore.txt
fi

git add -N -- infra/envoy/grpc-web/envoy.yaml

git status
git diff --stat
git diff --name-status --diff-filter=D --exit-code
git diff --quiet -- docs/01-micro-tasks.md
git diff --quiet -- api/master-api.json
```

What this does:

- Verifies the API Gateway module before staging.
- Shows the required status and diff stat.
- Confirms again that there are no deletions.
- Confirms protected and unrelated shared files remain unchanged.

## 9. Stage Only The Selected Changes

```bash
if [ -s /tmp/api-gateway-auto-restore.txt ]; then
  git add --pathspec-from-file=/tmp/api-gateway-auto-restore.txt
fi

git add -- backend/go.work backend/go.work.sum infra/envoy/grpc-web/envoy.yaml

cp /tmp/api-gateway-auto-restore.txt /tmp/api-gateway-allowed.txt
printf '%s\n' \
  backend/go.work \
  backend/go.work.sum \
  infra/envoy/grpc-web/envoy.yaml \
  >> /tmp/api-gateway-allowed.txt
LC_ALL=C sort -u -o /tmp/api-gateway-allowed.txt /tmp/api-gateway-allowed.txt

git diff --cached --name-only | LC_ALL=C sort -u > /tmp/api-gateway-staged.txt
comm -23 /tmp/api-gateway-staged.txt /tmp/api-gateway-allowed.txt \
  > /tmp/api-gateway-unexpected-staged.txt

test ! -s /tmp/api-gateway-unexpected-staged.txt || {
  echo "Unexpected staged files:"
  sed -n '1,240p' /tmp/api-gateway-unexpected-staged.txt
  exit 1
}

git status
git diff --cached --stat
git diff --cached --name-status
git diff --cached --name-status --diff-filter=D --exit-code
test -z "$(git diff --cached --name-only -- docs/01-micro-tasks.md)"
test -z "$(git diff --cached --name-only -- api/master-api.json)"
test -z "$(git diff --cached --name-only -- backend/services/api-gateway/server)"
```

What this does:

- Stages only the reviewed API Gateway files and shared-file updates.
- Verifies the staged paths are all in the explicit allowlist.
- Confirms no deletion, protected file, unrelated API contract, or compiled
  executable is staged.

The staged deletion check and the three protected-path checks must print
nothing and exit successfully.

## 10. Commit With The Exact Message

Verify the staged selection one final time, then commit:

```bash
git diff --cached --quiet && {
  echo "Nothing is staged. Stop."
  exit 1
}

git diff --cached --name-status --diff-filter=D --exit-code
test -z "$(git diff --cached --name-only -- docs/01-micro-tasks.md)"
test -z "$(git diff --cached --name-only -- api/master-api.json)"
test -z "$(git diff --cached --name-only -- backend/services/api-gateway/server)"

git commit -m "Merge API Gateway work into dev"
```

Verify immediately after the commit:

```bash
git status --short --branch
test "$(git log -1 --format=%s)" = "Merge API Gateway work into dev"
git show --stat --oneline HEAD
git show --format= --name-status --diff-filter=D --exit-code HEAD
test -z "$(git show --format= --name-only HEAD -- docs/01-micro-tasks.md)"
test -z "$(git show --format= --name-only HEAD -- api/master-api.json)"
test -z "$(git show --format= --name-only HEAD -- backend/services/api-gateway/server)"
```

What this does:

- Creates the required selective integration commit.
- Confirms the exact commit message.
- Confirms the commit contains no deletions or protected/unwanted paths.

## 11. Fast-Forward Dev

```bash
git switch dev
git merge --ff-only chore/API-Gateway-selective
```

What this does:

- Moves `dev` to the selective integration commit.
- Does not merge `feature/api-gateway-service`.
- Fails safely if `dev` changed and cannot fast-forward.

## 12. Final Verification

```bash
test "$(git rev-parse dev)" = "$(git rev-parse chore/API-Gateway-selective)"
test "$(git log -1 --format=%s)" = "Merge API Gateway work into dev"

git status --short --branch
git log --oneline -1
git show --stat --oneline HEAD
git show --format= --name-status --diff-filter=D --exit-code HEAD
test -z "$(git show --format= --name-only HEAD -- docs/01-micro-tasks.md)"
test -z "$(git show --format= --name-only HEAD -- api/master-api.json)"
test -z "$(git show --format= --name-only HEAD -- backend/services/api-gateway/server)"

(
  cd backend/services/api-gateway
  go test ./...
)
```

What this does:

- Confirms `dev` exactly matches the integration branch.
- Confirms the latest commit has the required message.
- Confirms the commit contains no deletions and did not touch protected or
  excluded files.
- Runs the API Gateway tests again after fast-forwarding `dev`.
