.PHONY: run build test lint docker-up docker-down clean

GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
GOMOD=$(GOCMD) mod
BINARY_NAME=api
BINARY_DIR=bin

# Запустить приложение локально
run:
	$(GOCMD) run ./cmd/api

# Собрать бинарник
build:
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/api

# Запустить тесты с race detector и покрытием
test:
	$(GOTEST) -race -cover ./...

# Проверить код линтером
lint:
	$(GOVET) ./...

# Запустить сервисы через Docker Compose
docker-up:
	docker compose up --build -d

# Остановить и удалить контейнеры
docker-down:
	docker compose down

# Просмотр логов контейнеров
docker-logs:
	docker compose logs -f

# Удалить бинарники и артефакты тестов
clean:
	@rm -rf $(BINARY_DIR)
	@rm -f coverage.out coverage.html

# Показать справку
help:
	@echo "Доступные команды:"
	@grep -E '^# ' Makefile | sed 's/# //' | awk 'BEGIN {FS = ": "}; {printf "  %-20s %s\n", $$1, $$2}'