#!/usr/bin/env bash
set -euo pipefail

while IFS= read -r dockerfile; do
  service="$(basename "$(dirname "$dockerfile")")"
  echo "Building ${service} from ${dockerfile}"
  docker build -f "$dockerfile" -t "ecommerce/${service}:ci-local" .
done < <(find backend/services -maxdepth 2 -name Dockerfile | sort)

