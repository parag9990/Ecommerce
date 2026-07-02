# Recommendation Service Troubleshooting

## Common Error

- `/healthz` fails.
- Recommendations are empty.
- Kafka consumer logs retry or DLQ events.

## Possible Cause

- MongoDB, Redis, or Kafka is unavailable.
- Interaction/product feature data has not been ingested.
- Event topic names differ from producers.

## Fix

- Start MongoDB, Redis, Kafka, and run Mongo migrations.
- Generate product/user interaction events.
- Confirm `RECOMMENDATION_EVENTS_TOPIC` and related topic names.

## Verification Command

```powershell
curl http://localhost:8089/healthz
docker compose logs recommendation-service
```
