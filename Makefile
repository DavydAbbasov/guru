.PHONY: proto build test lint tidy run-products run-notification migrate-up migrate-down swag clean help up down logs

proto:
	protoc --go_out=./apis --go_opt=paths=source_relative \
		-I=./proto \
		./proto/products/v1/events/events.proto

build:
	go build -o bin/products ./backend/products/cmd
	go build -o bin/notification ./backend/notification/cmd

test:
	go test -race -timeout 60s ./...

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

run-products:
	cd backend/products && go run ./cmd

run-notification:
	cd backend/notification && go run ./cmd

migrate-up:
	cd backend/products && go run ../../cmd/migrator up

migrate-down:
	cd backend/products && go run ../../cmd/migrator down

swag:
	swag init -g backend/products/cmd/main.go -o backend/products/docs

clean:
	rm -rf bin/

up:
	cd devops/local && docker compose up -d

down:
	cd devops/local && docker compose down

logs:
	cd devops/local && docker compose logs -f

help:
	@echo "Targets: build, test, lint, tidy, run-products, run-notification, migrate-up, migrate-down, swag, proto, clean, up, down, logs"
