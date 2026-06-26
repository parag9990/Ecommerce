#!/usr/bin/env bash
set -euo pipefail

find backend -name go.mod -not -path '*/vendor/*' -exec dirname {} \; | sort

