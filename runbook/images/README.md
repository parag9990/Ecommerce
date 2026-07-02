# Runbook Screenshot Placeholders

Use this folder for screenshots referenced by [../05_FULL_LOCAL_RUNBOOK.md](../05_FULL_LOCAL_RUNBOOK.md).

No real screenshots were captured during this documentation pass. Capture these manually after the local stack runs. Prefer FHD `1920x1080` PNG files; HD `1366x768` is acceptable when screen size is limited.

| File | What to capture |
| --- | --- |
| `step-01-prerequisites-fhd.png` | Terminal showing Docker, Compose, Go, Node, pnpm, and Buf version checks |
| `step-02-folder-structure-fhd.png` | Repository root in the IDE or terminal with `backend`, `frontend`, `runbook`, `infra`, and `docker-compose.yml` visible |
| `step-03-environment-setup-fhd.png` | Service `.env.example` files or compose env wiring visible in the IDE |
| `step-04-service-start-order-fhd.png` | The service dependency/start order section or `docker compose config` output |
| `step-05-full-local-run-fhd.png` | Terminal after `docker compose up -d --build` completes |
| `step-06-health-checks-fhd.png` | Terminal showing gateway health checks and `docker compose ps` |
| `step-07-user-seller-superadmin-flow-fhd.png` | Browser tabs for user app, seller dashboard, session analytics dashboard, and superadmin panel |
| `step-08-testing-checklist-fhd.png` | Terminal showing test command start or completed summary |
| `step-09-troubleshooting-fhd.png` | Terminal showing `docker compose logs -f api-gateway` or a focused service log |
| `step-10-stop-reset-fhd.png` | Terminal showing `docker compose down` or reset command confirmation |

Keep screenshots free of real secrets, tokens, customer data, or production credentials.
