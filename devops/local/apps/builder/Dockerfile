ARG GO_VERSION=1.25

FROM golang:${GO_VERSION}-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ARG SERVICE_PATH
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/app ./${SERVICE_PATH}

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 app
USER app
WORKDIR /app
COPY --from=builder /out/app /app/app
ENTRYPOINT ["/app/app"]
