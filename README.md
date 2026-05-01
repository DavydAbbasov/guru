# guru

Two Go services around a product catalog with async notifications.
**products** exposes a REST API (Fiber) and emits domain events to Kafka.
**notification** consumes those events and logs them. Full local stack
(infra + apps + observability) ships as a single `docker compose`.

## Stack

| Layer          | Tech                                  |
| -------------- | ------------------------------------- |
| Language       | Go                                    |
| HTTP           | Fiber v2                              |
| DI / lifecycle | uber/fx                               |
| Persistence    | PostgreSQL via GORM                   |
| Messaging      | Kafka via IBM/sarama                  |
| Migrations     | pressly/goose                         |
| Tracing        | OpenTelemetry → Tempo (OTLP/HTTP)     |
| Metrics        | Prometheus                            |
| Logs           | zap → Promtail → Loki                 |
| Config         | spf13/viper + go-playground/validator |
| API docs       | swaggo/swag                           |

## Architecture

```
                +-------------+
                | PostgreSQL  |
                +------+------+
                       ^ GORM
   HTTP/REST    +------+-------+   produce    +---------+   consume   +---------------+
  ===========> |   products   | ===========> |  Kafka  | ==========> |  notification |
               +------+-------+               +---------+              +-------+-------+
                      | metrics / traces / logs                                |
                      v                                                        v
                +-----+------+        +---------+        +---------+      +-------+
                | Prometheus |        |  Tempo  |        |  Loki   | <--- | logs  |
                +-----+------+        +----+----+        +----+----+      +-------+
                      \________________ Grafana _____________/
```

## Run (one command)

```sh
make up
```

Brings up Postgres, Kafka, products, notification, Tempo, Prometheus, Loki,
Promtail, Grafana. Migrations run automatically via a one-shot `migrator`
container; products is gated on `service_completed_successfully` so it only
starts on a migrated DB.

| Endpoint                                      | What                |
| --------------------------------------------- | ------------------- |
| `http://localhost:8080/api/v1/products`       | products REST API   |
| `http://localhost:8080/docs/swagger/index.html` | Swagger UI        |
| `http://localhost:8080/metrics`               | Prometheus scrape   |
| `http://localhost:8080/health/live` `/ready`  | health probes       |
| `http://localhost:3000`                       | Grafana (anon admin) |
| `http://localhost:9090`                       | Prometheus UI       |
| `localhost:9094`                              | Kafka (host listener) |
| `localhost:55432`                             | Postgres            |

`make down` stops everything; `make logs` tails all containers.

## Configuration

Each service reads `config.yaml` from its working dir. In compose, that file
is mounted from `devops/local/etc/{products,notification}.yaml`. For local
runs, `backend/<svc>/config.yaml` is gitignored — copy from
`config.example.yaml`.

Any key can be overridden by env with `CFG_` prefix and dots replaced by
underscores: `database.host` → `CFG_DATABASE_HOST`. Viper does **not**
expand `${...}` — secrets must come from env, not from the YAML.

Compose port overrides (host side only):
`POSTGRES_HOST_PORT`, `KAFKA_HOST_PORT`, `PRODUCTS_HOST_PORT`,
`PROMETHEUS_HOST_PORT`, `GRAFANA_HOST_PORT`.

## Local dev (without Docker)

For iterating on a single service against already-running infra:

```sh
cp backend/products/config.example.yaml      backend/products/config.yaml
cp backend/notification/config.example.yaml  backend/notification/config.yaml
make migrate-up
make run-products       # terminal 1
make run-notification   # terminal 2
```

Requires Postgres and Kafka reachable at the addresses in `config.yaml`
(simplest: `make up` for infra only, then `docker compose stop products
notification` and run the Go binaries on the host).

## Make targets

| Command                 | Purpose                                |
| ----------------------- | -------------------------------------- |
| `make up`               | Start full stack via docker compose    |
| `make down`             | Stop the stack                         |
| `make logs`             | Tail compose logs                      |
| `make build`            | Build both binaries to `bin/`          |
| `make test`             | `go test -race -timeout 60s ./...`     |
| `make lint`             | `golangci-lint run ./...`              |
| `make tidy`             | `go mod tidy`                          |
| `make run-products`     | Run products on host                   |
| `make run-notification` | Run notification on host               |
| `make migrate-up/down`  | Apply / roll back DB migrations        |
| `make swag`             | Regenerate Swagger for products        |
| `make proto`            | Regenerate protobuf code               |
| `make clean`            | `rm -rf bin/`                          |

## Design notes

- **fx-wired DI.** Every package exposes an `fx.Module`; `internal/container`
  composes them. Lifecycle hooks (`OnStart`/`OnStop`) own goroutines and
  graceful shutdown — server failure calls `fx.Shutdowner` so the process
  exits cleanly instead of leaking goroutines.
- **Event publishing is non-atomic.** Products commits to Postgres, then
  publishes to Kafka. A broker outage between commit and publish drops the
  event. A transactional outbox is the right fix; left out on purpose.
- **Single event type.** The Kafka dispatcher in notification is wired for
  one topic / one event shape. Adding more types means a small dispatcher,
  not a new pipeline.
- **Auto topic creation** is enabled in the Kafka image — fine for dev,
  not for prod.
- **Migrator as a compose service** runs once at stack start, then exits.
  Products `depends_on` it with `service_completed_successfully`, so
  start-up order is guaranteed without sleep loops.
- **Observability is fully wired:** Fiber middleware emits Prometheus
  metrics with request labels; OTLP/HTTP traces go to Tempo; container
  stdout is scraped by Promtail into Loki; Grafana datasources are
  auto-provisioned.
- **Config validation** runs at boot via `go-playground/validator` — the
  service refuses to start on a malformed YAML rather than failing later.

## Deliberately deferred

Out of scope for this exercise; understood and tracked:

- Kafka SASL/TLS (currently plaintext).
- Transactional outbox for products write-then-publish.
- Dead-letter queue on the notification consumer.
- Multi-event Kafka dispatcher.
- `otelsarama` for Kafka producer/consumer span propagation.
- `protovalidate` / `buf-lint` in CI for `proto/`.
- Generated mocks (`mockery` / `gomock`) — tests use hand-written stubs.
