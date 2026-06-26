# Commands Reference

| Goal | Command |
|---|---|
| Validate Compose | `docker compose config` |
| Start everything | `docker compose up -d --build` |
| Start infra | `make infra-up` |
| Start backends | `make backend-up` |
| Start frontends | `make frontend-up` |
| Status | `docker compose ps` |
| Follow all logs | `docker compose logs -f` |
| One service logs | `docker compose logs -f api-gateway` |
| Rebuild one service | `docker compose up -d --build search-service` |
| Stop | `docker compose down` |
| Reset data | `docker compose down -v` |
| Go tests | `make test-go` |
| Frontend tests | `make test-frontend` |
| K8s manifest check | `kubectl apply --dry-run=client -k infra/k8s/overlays/dev` |
| Tilt Compose UI | `tilt up` |

On Windows, run Make commands from a shell where GNU Make works; the underlying Docker commands work directly in PowerShell.
