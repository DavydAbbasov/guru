ARG GO_VERSION=1.25
ARG SWAG_VERSION=v1.16.4

FROM golang:${GO_VERSION}-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ARG SERVICE_PATH
ARG SWAG_VERSION
RUN if [ "$SERVICE_PATH" = "backend/products/cmd" ]; then \
      go install github.com/swaggo/swag/cmd/swag@${SWAG_VERSION} && \
      swag init -g backend/products/cmd/main.go --parseDependency --output backend/products/docs; \
    fi

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/app ./${SERVICE_PATH}

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 app
USER app
WORKDIR /app
COPY --from=builder /out/app /app/app
ENTRYPOINT ["/app/app"]
