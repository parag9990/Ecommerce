# Post-Run Checklist

- [ ] `docker compose ps` shows expected services running or healthy.
- [ ] API Gateway liveness passes: `curl http://localhost:8080/health/live`.
- [ ] API Gateway readiness passes or known degraded dependencies are understood: `curl http://localhost:8080/health/ready`.
- [ ] Mailpit opens at `http://localhost:8025`.
- [ ] User app opens at `http://localhost:3000`.
- [ ] Seller dashboard opens at `http://localhost:3001`.
- [ ] Session analytics dashboard opens at `http://localhost:3002`.
- [ ] Superadmin panel opens at `http://localhost:3003`.
- [ ] Frontend Docker health endpoints pass on `/healthz` for ports `3000`, `3001`, `3002`, and `3003`.
- [ ] No service is crash-looping in `docker compose logs`.
- [ ] Known blockers are recorded before functional testing: missing seed users, auth signup mismatch, disabled payment providers, blank internal CMS/payment/event config.
