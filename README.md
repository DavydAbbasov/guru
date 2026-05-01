# guru

Two Go services:

- **products** — REST API on Fiber, stores products in Postgres, emits create/delete events to Kafka.
- **notification** — Kafka consumer, logs incoming events.

Full stack (services + Postgres + Kafka + Prometheus/Grafana/Tempo/Loki) starts with one command.

## Run

```sh
make up
```

Builds the two service images, starts dependencies, runs migrations in a one-shot container, then starts products and notification.

| URL                                             | What                |
| ----------------------------------------------- | ------------------- |
| `http://localhost:8080/api/v1/products`         | products REST API   |
| `http://localhost:8080/docs/swagger/index.html` | Swagger UI          |
| `http://localhost:8080/metrics`                 | Prometheus metrics  |
| `http://localhost:8080/health/{live,ready}`     | health probes       |
| `http://localhost:3000`                         | Grafana             |
| `http://localhost:9090`                         | Prometheus UI       |

`make down` stops everything; `make logs` tails containers.

## API

```sh
# create
curl -X POST http://localhost:8080/api/v1/products/ \
  -H 'Content-Type: application/json' -d '{"name":"yoga mat"}'

# list (paginated)
curl 'http://localhost:8080/api/v1/products/?page=1&limit=20'

# delete
curl -X DELETE http://localhost:8080/api/v1/products/<id>
```

Or open Swagger UI and use *Try it out*.

## Tests

```sh
make test    # go test -race -timeout 60s ./...
```

Service-layer unit tests cover create / list / delete (validation, error mapping, publisher contract, metric counters) and the notification consumer.

## Metrics

Exposed at `/metrics`, scraped by Prometheus, visualised in Grafana.

| Metric                                | Type        | Labels         |
| ------------------------------------- | ----------- | -------------- |
| `products_products_created_total`     | counter     | —              |
| `products_products_deleted_total`     | counter     | —              |
| `products_http_requests_total`        | counter     | method, status, path |
| `products_http_request_duration_seconds` | histogram | method, path  |
| `notification_events_consumed_total`  | counter     | type           |
| `notification_events_processed_total` | counter     | type           |
| `notification_events_failed_total`    | counter     | type, reason   |
| `notification_events_processing_duration_seconds` | histogram | type |

## Stack

| Layer       | Library                               |
| ----------- | ------------------------------------- |
| HTTP        | `gofiber/fiber/v2`                    |
| DI          | `uber-go/fx`                          |
| Persistence | `gorm.io/gorm` (PostgreSQL)           |
| Messaging   | `IBM/sarama` (Kafka)                  |
| Migrations  | `pressly/goose`                       |
| Tracing     | OpenTelemetry → Tempo (OTLP/HTTP)     |
| Metrics     | `prometheus/client_golang`            |
| Logs        | `uber-go/zap` → Promtail → Loki       |
| Config      | `spf13/viper` + `go-playground/validator` |
| API docs    | `swaggo/swag`                         |

## Configuration

Each service reads `config.yaml` from its working dir. In compose it's mounted from `devops/local/etc/{products,notification}.yaml`. For host-side runs, copy `config.example.yaml` → `config.yaml`.

Any key can be overridden via env: `CFG_<UPPER_DOTTED_KEY>`, e.g. `database.host` → `CFG_DATABASE_HOST`.

## Make targets

| Command                     | Purpose                          |
| --------------------------- | -------------------------------- |
| `make up` / `down` / `logs` | Stack lifecycle                  |
| `make build`                | Build both binaries              |
| `make test`                 | Run tests with race detector     |
| `make lint`                 | `golangci-lint run ./...`        |
| `make migrate-up` / `-down` | Apply / roll back migrations     |
| `make run-products` / `-notification` | Run a service on the host |
| `make swag` / `make proto`  | Regenerate Swagger / protobuf    |
