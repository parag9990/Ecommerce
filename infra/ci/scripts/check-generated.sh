#!/usr/bin/env bash
set -euo pipefail

buf generate

if ! git diff --exit-code -- backend/shared/gen/go frontend/packages/proto-client/src/gen; then
  echo "Generated proto clients are stale. Run 'buf generate' and commit the output." >&2
  exit 1
fi

