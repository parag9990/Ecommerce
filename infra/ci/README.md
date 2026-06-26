# CI foundation

Reusable CI helpers live in `infra/ci/scripts` so local verification and GitHub Actions can run the same checks.

```bash
bash infra/ci/scripts/list-go-modules.sh
bash infra/ci/scripts/check-generated.sh
bash infra/ci/scripts/docker-build-all.sh
```

Policy files under `infra/ci/policies` document the scanner and Dockerfile lint defaults used by the platform CI gate.

