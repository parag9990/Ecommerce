# Pre-Run Checklist

- [ ] Docker is running.
- [ ] Root `docker-compose.yml` is the local runtime source of truth for this run.
- [ ] `docker compose config` succeeds.
- [ ] Ports in [09 Ports And Endpoints](../09_PORTS_AND_ENDPOINTS.md) are free.
- [ ] Go version is compatible with `backend/go.work`.
- [ ] Node and pnpm match `frontend/package.json`.
- [ ] No production secrets are stored in local `.env` files.
- [ ] Payment provider variables are either intentionally disabled or configured for sandbox.
- [ ] Seed/demo users are confirmed with the team.
- [ ] Auth signup route mismatch is understood before relying on signup for test users.
- [ ] Blank internal CMS/payment/event token and endpoint defaults are either acceptable for the test or configured locally.
- [ ] You know whether to run Docker Compose or manual services.
