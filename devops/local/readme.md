# Local stack

```bash
cd devops/local
docker compose up -d
```

## Endpoints
- products HTTP — http://localhost:8080
- Grafana (anon admin) — http://localhost:3000
- Postgres — localhost:55432 (postgres/postgres)
- Kafka external — localhost:9094

## Override host ports
```bash
PRODUCTS_HOST_PORT=8090 GRAFANA_HOST_PORT=3030 docker compose up -d
```

## Layout
- `compose.yaml` — service graph
- `etc/<service>.yaml` — application config mounted as `/app/config.yaml`
- `etc/infra/<infra>.yaml` — infra component config (tempo/prometheus/loki/promtail/grafana)
- `apps/builder/Dockerfile` — shared Go multi-stage build (parametrized via `SERVICE_PATH` arg)
