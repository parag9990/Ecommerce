# Merge Superadmin Panel Work Into Dev

Use this runbook to bring the Superadmin Panel work from
`feature/superadmin_panel` into `dev` without merging the whole feature branch
and without deleting existing files from `dev`.

The source branch contains many unrelated deletions. Only added and modified
Superadmin Panel files are restored. Shared files are reconciled additively so
existing `dev` content remains intact.

## Rules

- Restore added and modified files only from `frontend/superadmin-panel` and
  `TaskImplementation/Superadmin Panel`.
- Treat Git rename detection as advisory. The source branch reports seven
  Superadmin Panel destination files as renames from unrelated frontend apps;
  `--no-renames` makes those destination files safe additions while their
  source-side deletions stay excluded.
- Do not apply any deletion diff.
- Do not restore shared files wholesale from the source branch.
- Do not modify, restore, stage, or commit `docs/01-micro-tasks.md`.
- Keep `api/master-api.json`, `backend/go.work`, and `backend/go.work.sum` from
  `dev`; the source versions remove existing `dev` content and are not needed
  to add this frontend module.
- Reconcile `frontend/package.json`, `frontend/pnpm-workspace.yaml`, and
  `frontend/pnpm-lock.yaml` additively.
- Commit with this exact message:

```text
Merge Superadmin Panel work into dev
```

## 1. Precheck

Run from the repository root:

```bash
set -euo pipefail

git status --short --branch
git rev-parse --verify dev
git rev-parse --verify feature/superadmin_panel
test -z "$(git status --porcelain)" || {
  echo "Working tree is dirty. Commit or stash existing work first."
  exit 1
}

if git show-ref --verify --quiet refs/heads/chore/Superadmin-Panel-selective; then
  echo "Integration branch already exists. Inspect it before continuing."
  exit 1
fi
```

What this does:

- Confirms both required branches exist.
- Requires a clean worktree and index before any branch change.
- Avoids accidentally reusing an old integration branch.

## 2. Create A Safe Integration Branch From Dev

```bash
git switch dev
test -z "$(git status --porcelain)" || { echo "dev is dirty. Stop."; exit 1; }

git switch -c chore/Superadmin-Panel-selective
test "$(git rev-parse HEAD)" = "$(git rev-parse dev)"
```

What this does:

- Starts the selective work at the current `dev` commit.
- Keeps `dev` unchanged until the final fast-forward.

## 3. Review The Full Branch Difference

```bash
git diff --name-status dev..feature/superadmin_panel
git diff --no-renames --name-status --diff-filter=D \
  dev..feature/superadmin_panel
git diff --no-renames --name-status --diff-filter=AM \
  dev..feature/superadmin_panel
```

Also review the dedicated module paths and all changed shared files:

```bash
git diff --no-renames --name-status --diff-filter=AM \
  dev..feature/superadmin_panel -- \
  frontend/superadmin-panel \
  'TaskImplementation/Superadmin Panel'

git diff --stat dev..feature/superadmin_panel -- \
  api/master-api.json \
  backend/go.work \
  backend/go.work.sum \
  docs/01-micro-tasks.md \
  frontend/package.json \
  frontend/pnpm-lock.yaml \
  frontend/pnpm-workspace.yaml
```

In the repository state used to write this runbook, the full branch difference
contains 2,151 paths. The relevant dedicated candidates are 200 files under
`frontend/superadmin-panel` and 27 files under
`TaskImplementation/Superadmin Panel`. The broad unrelated deletions must not
be applied.

## 4. Build The Auto-Restore File List

Build a NUL-delimited list so the path containing a space is handled safely:

```bash
git diff --no-renames --name-only -z --diff-filter=AM \
  dev..feature/superadmin_panel -- \
  frontend/superadmin-panel \
  'TaskImplementation/Superadmin Panel' \
  > /tmp/superadmin-panel-auto-restore.pathspec

while IFS= read -r -d '' path; do
  case "$path" in
    frontend/superadmin-panel/*|'TaskImplementation/Superadmin Panel'/*)
      printf '%s\n' "$path"
      ;;
    *)
      printf 'Unexpected path in restore list: %s\n' "$path" >&2
      exit 1
      ;;
  esac
done < /tmp/superadmin-panel-auto-restore.pathspec

test -s /tmp/superadmin-panel-auto-restore.pathspec || {
  echo "No Superadmin Panel files were found. Stop and recheck the branches."
  exit 1
}
```

What this does:

- Includes only added and modified files in the two dedicated module paths.
- Uses `--no-renames` so all Superadmin Panel destination files are included as
  additions without selecting the unrelated source deletions.
- Cannot include `docs/01-micro-tasks.md` or any shared file.

## 5. Restore Added And Modified Superadmin Panel Files

```bash
git restore --overlay \
  --source=feature/superadmin_panel \
  --worktree \
  --pathspec-from-file=/tmp/superadmin-panel-auto-restore.pathspec \
  --pathspec-file-nul
```

What this does:

- Copies only the allowlisted Superadmin Panel source and task files.
- Uses `--overlay`, so files already present on `dev` are not removed merely
  because the source branch lacks them.
- Does not restore any shared file.

## 6. Handle Shared Files Manually

### Keep Protected And Unrelated Shared Files From Dev

Review these source differences, but do not apply them:

```bash
git diff dev..feature/superadmin_panel -- api/master-api.json
git diff dev..feature/superadmin_panel -- backend/go.work
git diff --stat dev..feature/superadmin_panel -- backend/go.work.sum
git diff dev..feature/superadmin_panel -- docs/01-micro-tasks.md
```

The source changes to the API master file and Go workspace files remove or
reduce content already on `dev`. The micro-task file also changes many unrelated
task statuses. Leave all four files exactly as they are on `dev`:

```bash
git diff --exit-code -- \
  api/master-api.json \
  backend/go.work \
  backend/go.work.sum \
  docs/01-micro-tasks.md
```

This command must print nothing.

### Add Superadmin Scripts Without Replacing `frontend/package.json`

First review the source diff:

```bash
git diff dev..feature/superadmin_panel -- frontend/package.json
```

Then add only the four Superadmin scripts with a JSON-aware update:

```bash
node <<'NODE'
const fs = require("node:fs");

const path = "frontend/package.json";
const packageJson = JSON.parse(fs.readFileSync(path, "utf8"));
const additions = {
  "superadmin:dev": "pnpm --filter superadmin-panel dev",
  "superadmin:test": "pnpm --filter superadmin-panel test",
  "superadmin:typecheck": "pnpm --filter superadmin-panel typecheck",
  "superadmin:build": "pnpm --filter superadmin-panel build",
};

packageJson.scripts ||= {};
for (const [name, command] of Object.entries(additions)) {
  const current = packageJson.scripts[name];
  if (current !== undefined && current !== command) {
    throw new Error(`Refusing to replace existing script ${name}: ${current}`);
  }
  packageJson.scripts[name] = command;
}

fs.writeFileSync(path, `${JSON.stringify(packageJson, null, 2)}\n`);
NODE

node -e 'JSON.parse(require("node:fs").readFileSync("frontend/package.json", "utf8"))'
git diff -- frontend/package.json
```

Verify that the existing package manager, engine requirements, and all existing
user and seller scripts remain present. Do not copy the source file wholesale,
because its version removes those `dev` fields.

### Add The Workspace Entry Without Replacing The Workspace File

Review the source diff:

```bash
git diff dev..feature/superadmin_panel -- frontend/pnpm-workspace.yaml
```

Manually add only this entry under `packages:` in
`frontend/pnpm-workspace.yaml`:

```yaml
  - "superadmin-panel"
```

Keep the existing `user-app` and `packages/*` entries. Then verify:

```bash
git diff -- frontend/pnpm-workspace.yaml
grep -Fx '  - "user-app"' frontend/pnpm-workspace.yaml
grep -Fx '  - "packages/*"' frontend/pnpm-workspace.yaml
grep -Fx '  - "superadmin-panel"' frontend/pnpm-workspace.yaml
test "$(grep -Fxc '  - "superadmin-panel"' frontend/pnpm-workspace.yaml)" -eq 1
```

### Regenerate The Lockfile From The Combined Dev Workspace

Do not restore `frontend/pnpm-lock.yaml` from the source branch; that version
removes the existing `dev` importers and dependency resolutions. Regenerate it
after the panel package and workspace entry are present:

```bash
(
  cd frontend
  pnpm install --lockfile-only
)

git diff -- frontend/pnpm-lock.yaml
grep -E '^  (packages/proto-client|user-app|superadmin-panel):$' \
  frontend/pnpm-lock.yaml
```

The regenerated lockfile must retain the existing `packages/proto-client` and
`user-app` importers and add `superadmin-panel`. Review its diff before
continuing; stop if existing `dev` workspace importers are removed.

## 7. Confirm No Deletions Were Applied

```bash
git diff --name-status --diff-filter=D --exit-code
git diff --exit-code -- \
  api/master-api.json \
  backend/go.work \
  backend/go.work.sum \
  docs/01-micro-tasks.md
```

Both commands must print nothing. If either command prints a path or diff, stop
before staging.

## 8. Show Status And Diff Stat Before Staging

```bash
git status
git diff --stat

git diff --no-renames --stat dev..feature/superadmin_panel -- \
  frontend/superadmin-panel \
  'TaskImplementation/Superadmin Panel'

git diff --check
git diff --name-status --diff-filter=D --exit-code
```

What this does:

- Shows tracked shared-file edits and all untracked panel paths before staging.
- Shows the expected source-branch stat for the dedicated allowlisted paths.
- Checks whitespace errors and confirms there are no deleted files.

## 9. Stage Only The Selected Changes

```bash
git add \
  --pathspec-from-file=/tmp/superadmin-panel-auto-restore.pathspec \
  --pathspec-file-nul

git add \
  frontend/package.json \
  frontend/pnpm-workspace.yaml \
  frontend/pnpm-lock.yaml
```

Verify the index immediately after staging:

```bash
git status
git diff --cached --stat
git diff --cached --check
git diff --cached --name-status --diff-filter=D --exit-code

test -z "$(git diff --cached --name-only -- \
  api/master-api.json \
  backend/go.work \
  backend/go.work.sum \
  docs/01-micro-tasks.md)" || {
  echo "A protected or unrelated shared file is staged. Stop."
  exit 1
}
```

The staged deletion check and protected-file check must print nothing. The index
should contain only the two dedicated Superadmin Panel paths and the three
manually reconciled frontend shared files.

## 10. Commit

Run one final guard, then commit with the exact required message:

```bash
git diff --cached --quiet && { echo "Nothing is staged. Stop."; exit 1; }
test -z "$(git diff --cached --name-only -- docs/01-micro-tasks.md)"
git diff --cached --name-status --diff-filter=D --exit-code

git commit -m "Merge Superadmin Panel work into dev"
```

Verify immediately after the commit:

```bash
test "$(git log -1 --pretty=%s)" = "Merge Superadmin Panel work into dev"
test -z "$(git diff-tree --no-commit-id --name-only --diff-filter=D -r HEAD)"
git diff --exit-code HEAD^ HEAD -- docs/01-micro-tasks.md
git status --short --branch
git show --stat --oneline HEAD
```

The commit must contain no deleted files and no change to
`docs/01-micro-tasks.md`.

## 11. Fast-Forward Dev

Only after all commit checks pass:

```bash
git switch dev
git merge --ff-only chore/Superadmin-Panel-selective
```

What this does:

- Moves `dev` to the already reviewed selective integration commit.
- Fails safely if `dev` moved and can no longer fast-forward.

## 12. Final Verification

```bash
test "$(git branch --show-current)" = "dev"
test "$(git rev-parse dev)" = \
  "$(git rev-parse chore/Superadmin-Panel-selective)"
test "$(git log -1 --pretty=%s)" = "Merge Superadmin Panel work into dev"
test -z "$(git diff-tree --no-commit-id --name-only --diff-filter=D -r HEAD)"
git diff --exit-code HEAD^ HEAD -- docs/01-micro-tasks.md

git status --short --branch
git log --oneline -1
git show --stat --oneline HEAD
```

Final expected results:

- The current branch is `dev`.
- `dev` and `chore/Superadmin-Panel-selective` point to the same commit.
- The latest commit message is exactly
  `Merge Superadmin Panel work into dev`.
- The selective commit contains no file deletions.
- `docs/01-micro-tasks.md` is unchanged by the selective commit.
- The working tree is clean.
