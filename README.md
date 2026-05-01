# guru

A two-service Go playground built around a product catalog and asynchronous
notifications. The **products** service exposes a REST API (Fiber) and emits
domain events to Kafka; the **notification** service consumes those events and
logs them.

## Stack

| Layer          | Tech                                    |
| -------------- | --------------------------------------- |
| Language       | Go                                      |
| HTTP           | Fiber v2                                |
| DI / lifecycle | uber/fx                                 |
| Persistence    | PostgreSQL via GORM                     |
| Messaging      | Kafka via IBM/sarama                    |
| Migrations     | pressly/goose                           |
| Tracing        | OpenTelemetry → Tempo (OTLP/HTTP)       |
| Metrics        | Prometheus                              |
| Logging        | uber-go/zap                             |
| Config         | spf13/viper + go-playground/validator   |
| API docs       | swaggo/swag                             |

## Architecture

```
                +-------------+
                | PostgreSQL  |
                +------+------+
                       ^
                       | GORM
                       |
   HTTP/REST    +------+-------+   produce    +---------+   consume   +------------------+
  ===========> |   products   | ===========> |  Kafka  | ==========> |  notification    |
               +------+-------+               +---------+              +--------+---------+
                      |                                                        |
                      |  metrics / traces                                      |  logs events
                      v                                                        v
                +-----+------+        +---------+        +---------+
                | Prometheus |        |  Tempo  |        | stdout  |
                +------------+        +---------+        +---------+
```

## Quick start

1. Copy the example configs:
   ```sh
   cp backend/products/config.example.yaml backend/products/config.yaml
   cp backend/notification/config.example.yaml backend/notification/config.yaml
   ```
2. Set the secrets via env vars (see below) — `config.yaml` is gitignored.
3. Apply migrations and run the services:
   ```sh
   make migrate-up
   make run-products       # in one terminal
   make run-notification   # in another
   ```

## Environment variables

Both services bind environment variables under the `CFG_` prefix. Dots in the
config key map to underscores: `database.host` → `CFG_DATABASE_HOST`.

| Var                        | Description                              |
| -------------------------- | ---------------------------------------- |
| `CFG_DATABASE_HOST`        | Postgres host (products only)            |
| `CFG_DATABASE_PORT`        | Postgres port                            |
| `CFG_DATABASE_USER`        | Postgres user                            |
| `CFG_DATABASE_PASS`        | Postgres password                        |
| `CFG_DATABASE_NAME`        | Postgres database name                   |
| `CFG_KAFKA_BROKERS`        | Kafka bootstrap servers                  |
| `CFG_KAFKA_TOPIC`          | Kafka topic                              |
| `CFG_SERVER_PORT`          | HTTP listen port                         |
| `CFG_LOGGER_LEVEL`         | Log level (debug/info/warn/error)        |
| `CFG_TRACER_DISABLED`      | Disable OTLP exporter                    |
| `CFG_TRACER_ENDPOINT`      | OTLP/HTTP endpoint (e.g. `tempo:4318`)   |

## Development commands

| Command               | Purpose                                   |
| --------------------- | ----------------------------------------- |
| `make build`          | Build both service binaries to `bin/`     |
| `make test`           | `go test -race -timeout 60s ./...`        |
| `make lint`           | Run `golangci-lint`                       |
| `make tidy`           | `go mod tidy`                             |
| `make run-products`   | Run the products service                  |
| `make run-notification` | Run the notification service            |
| `make migrate-up`     | Apply DB migrations                       |
| `make migrate-down`   | Roll back the last migration              |
| `make swag`           | Regenerate Swagger docs for products      |
| `make proto`          | Regenerate protobuf code                  |
| `make clean`          | Remove `bin/`                             |

## Production hardening — deliberately deferred

This is a test-assignment project. The following items are well understood but
left out to keep the scope tight; each would be a real concern in production:

- **Kafka SASL/TLS** — the brokers are reached over plaintext.
- **Transactional outbox** — the products write-then-publish path is not
  atomic; a Postgres commit followed by a broker outage drops the event.
- **Dead-letter queue** for poison pills on the notification consumer.
- **Multi-event Kafka dispatcher** — only one event type is wired.
- **`otelsarama`** — Kafka producer/consumer spans are not propagated.
- **`protovalidate` / `buf-lint`** in CI for `proto/`.
- **Generated mocks** (`mockery` / `gomock`) — the test suite stubs by hand.
