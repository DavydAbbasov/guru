# guru

Two Go services:

- **products** — REST API on Fiber, stores products in Postgres, emits create/delete events to Kafka via a transactional outbox.
- **notification** — Kafka consumer with idempotent delivery (consumer-side dedup), logs incoming events.

Producer writes the event row to an `outbox` table inside the same transaction as the product change; a background dispatcher drains the table to Kafka with retries, exponential backoff, and parking after `MaxAttempts`. A periodic cleanup tick removes old sent rows. Each event carries an `outbox-event-id` header; the consumer records it in `processed_events` (`INSERT ... ON CONFLICT DO NOTHING`) to skip duplicates. DB writes use a `TransactionManager` with commit-timeout protection, and pool sizing is selected per service via `Default` / `HighLoad` / `Light` profiles.

Full stack (services + Postgres + Kafka + Prometheus/Grafana/Tempo/Loki) starts with one command.

> **Scope.** This project deliberately exercises production-shape patterns end-to-end — transactional outbox, consumer-side idempotency, OTEL across the broker, structured config + validation, full observability stack. Treated as a sandbox for the techniques, not a search for the simplest implementation.

## Run

```sh
make up
```

Builds the two service images, starts dependencies, runs each service's migrations in a one-shot container (`migrator-products`, `migrator-notification`), then starts products and notification.

| URL                                             | What                                        |
| ----------------------------------------------- | ------------------------------------------- |
| `http://localhost:8080/api/v1/products`         | products REST API                           |
| `http://localhost:8080/docs/swagger/index.html` | Swagger UI                                  |
| `http://localhost:8080/metrics`                 | products Prometheus metrics                 |
| `http://localhost:8080/debug/pprof/`            | products pprof                              |
| `http://localhost:8080/health/{live,ready}`     | products health probes                      |
| `http://localhost:8082/metrics`                 | notification Prometheus metrics (admin)     |
| `http://localhost:8082/debug/pprof/`            | notification pprof (admin)                  |
| `http://localhost:8082/health/{live,ready}`     | notification health probes (admin)          |
| `http://localhost:3000`                         | Grafana                                     |
| `http://localhost:9090`                         | Prometheus UI                               |

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

Service-layer unit tests cover create / list / delete (validation, error mapping, outbox writes, metric counters) and the notification consumer (including idempotent dedup).

## Metrics

Exposed at `/metrics`, scraped by Prometheus, visualised in Grafana.

Per-service metric names use the service namespace (`products_*` / `notification_*`); shared metrics below are written as `<ns>_*`.

| Metric                                | Type        | Labels                |
| ------------------------------------- | ----------- | --------------------- |
| `products_products_created_total`     | counter     | —                     |
| `products_products_deleted_total`     | counter     | —                     |
| `products_products_active`            | gauge       | —                     |
| `<ns>_http_requests_total`            | counter     | method, path, status  |
| `<ns>_http_request_duration_seconds`  | histogram   | method, path          |
| `<ns>_http_requests_in_flight`        | gauge       | —                     |
| `<ns>_errors_total`                   | counter     | type                  |
| `<ns>_db_query_duration_seconds`      | histogram   | operation, table      |
| `<ns>_outbox_pending`                 | gauge       | —                     |
| `<ns>_outbox_dispatched_total`        | counter     | event_type            |
| `<ns>_outbox_dispatch_failures_total` | counter     | event_type            |
| `<ns>_outbox_abandoned_total`         | counter     | event_type            |
| `<ns>_outbox_dispatch_duration_seconds` | histogram | —                     |
| `<ns>_outbox_cleanup_deleted_total`   | counter     | —                     |
| `<ns>_idempotency_processed_total`    | counter     | event_type            |
| `<ns>_idempotency_duplicate_total`    | counter     | event_type            |
| `<ns>_idempotency_cleanup_deleted_total` | counter  | —                     |
| `notification_events_consumed_total`  | counter     | type                  |
| `notification_events_processed_total` | counter     | type                  |
| `notification_events_failed_total`    | counter     | type, reason          |
| `notification_events_parse_errors_total` | counter  | —                     |
| `notification_events_processing_duration_seconds` | histogram | type      |

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

## Design notes

- **Kafka, not SQS/NATS.** Chosen for partition-keyed ordering (events for one product land on one partition), consumer groups, and replay; outbox + idempotency would map cleanly to SQS or NATS too.
- **Transactional outbox.** DB write and event row land in the same transaction; a background dispatcher drains the outbox to Kafka with retries, exponential backoff, and parking after `MaxAttempts`. No "committed but lost" window.
- **Consumer-side idempotency.** `processed_events` with `INSERT ... ON CONFLICT DO NOTHING` makes redeliveries safe; duplicates are counted in `*_idempotency_duplicate_total`.
- **OTEL through Kafka.** Trace context is injected as Kafka headers and re-extracted by the consumer — same `trace_id` shows up in both services' logs and stitches into one trace in Tempo.
- **fx-wired lifecycle.** Each package is an `fx.Module`; server failure calls `fx.Shutdowner` to exit cleanly. The Kafka consumer drains in-flight messages on shutdown.

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
