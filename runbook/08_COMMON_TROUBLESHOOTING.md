# Common Troubleshooting

| Issue | Possible cause | Fix | Verification |
| --- | --- | --- | --- |
| Port already in use | Another local process owns the port | Stop the process or change compose host port | `docker compose ps` |
| Missing environment variables | Manual host run did not load `.env` | Copy `.env.example` and update hostnames | Service startup logs |
| Database connection failed | MySQL/MongoDB not healthy or wrong DNS | Start infra first; use `localhost` for host runs | `docker compose ps mysql mongodb` |
| Redis/cache connection failed | Redis not healthy or password mismatch | Use `local_redis_password` locally | `docker compose exec redis redis-cli -a local_redis_password ping` |
| Kafka issue | Kafka still starting or wrong broker address | Use `kafka:29092` in Docker, `localhost:9092` on host | Kafka topics list command |
| RabbitMQ issue | RabbitMQ vhost/user mismatch | Use `ecommerce/local_rabbitmq_password/ecommerce` | `http://localhost:15672` |
| Docker not running | Docker daemon unavailable | Start Docker Desktop/Engine | `docker compose version` |
| Docker compose failure | Invalid config, build failure, unhealthy dependency | Run `docker compose config`, then inspect failing logs | `docker compose logs <service>` |
| Migration failure | DB not ready or migration already partially applied | Wait for DB health, rerun migration job | `docker compose up migrate-auth` |
| npm/pnpm/yarn install failure | Wrong Node/pnpm version or stale lockfile | Use Node `>=22.13.0` and pnpm `11.5.0` | `pnpm --version` |
| Go module issue | Workspace/module mismatch or stale cache | Run tests from each module folder or use `make test-go`; do not run `go test ./...` from `backend/` | `go env GOWORK` |
| Python dependency issue | No Python service found | Not found in codebase - please confirm | N/A |
| Frontend API URL issue | Wrong `VITE_API_BASE_URL` | Use `http://localhost:8080` | Browser network tab |
| CORS issue | Frontend origin missing in gateway env | Update `CORS_ALLOWED_ORIGINS` | Gateway logs |
| API gateway routing issue | Downstream URL/gRPC address mismatch | Check `api-gateway/.env.example` mappings | `/health/ready` |
| Authentication/JWT/session issue | JWKS URL, issuer, audience, or key mismatch | Confirm auth and gateway JWT env values | `curl http://localhost:8081/.well-known/jwks.json` |
| Signup fails or route is missing | Auth signup is referenced by catalog/frontend, but auth route registration was not confirmed | Use existing confirmed credentials or resolve the auth route mismatch before signup testing | Auth service and gateway logs |
| Payment webhook/local callback issue | Providers disabled or webhook secrets missing | Configure sandbox provider vars and callback URL | Payment service logs |
| Payment event updates do not reach orders | `PAYMENT_EVENTS_ENDPOINT` or matching internal token is blank/mismatched | Configure the local endpoint and token pair before payment flow testing | Payment and order service logs |
| CMS/internal calls fail | Internal CMS token defaults are blank in examples | Set matching local internal tokens for the services under test | CMS and caller logs |
| Search index issue | Typesense empty/unhealthy or reindex not run | Start Typesense and run search reindex command | `curl http://localhost:8108/health` |
| Notification/email/SMS config issue | Mailpit/SMTP disabled or SMS provider missing | Use Mailpit locally; configure SMS only when needed | `http://localhost:8025` |

## Useful Log Commands

```powershell
docker compose logs -f
docker compose logs -f api-gateway
docker compose logs -f auth-service
```
