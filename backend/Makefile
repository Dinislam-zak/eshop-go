.PHONY: run build test lint docker-up docker-down clean help

GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
BINARY_NAME=api
BINARY_DIR=bin

## run: Запустить приложение локально
run:
	$(GOCMD) run ./cmd/api

## build: Собрать бинарник
build:
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/api

## test: Запустить тесты с race detector и покрытием
test:
	$(GOTEST) -race -cover ./...

## lint: Проверить код линтером
lint:
	$(GOVET) ./...

## docker-up: Запустить сервисы через Docker Compose
docker-up:
	docker compose up --build -d

## docker-down: Остановить и удалить контейнеры
docker-down:
	docker compose down

## docker-logs: Просмотр логов контейнеров
docker-logs:
	docker compose logs -f

## clean: Удалить бинарники и артефакты тестов
clean:
	@rm -rf $(BINARY_DIR)
	@rm -f coverage.out coverage.html

## help: Показать справку
help:
	@echo "Доступные команды:"
	@grep -E '^## ' Makefile | sed 's/## //' | awk -F': ' '{printf "  %-20s %s\n", $$1, $$2}'