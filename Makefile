.PHONY: help build run dev services-up services-down db-up db-down db-reset test clean

# Показать справку
help:
	@echo "Доступные команды:"
	@echo "  build        - Собрать приложение"
	@echo "  run          - Запустить приложение"
	@echo "  dev          - Запустить в режиме разработки"
	@echo "  services-up  - Запустить все сервисы (PostgreSQL, Kafka, Redis)"
	@echo "  services-down- Остановить все сервисы"
	@echo "  db-up        - Запустить только PostgreSQL"
	@echo "  db-down      - Остановить PostgreSQL"
	@echo "  db-reset     - Пересоздать базу данных"
	@echo "  test         - Запустить тесты"
	@echo "  clean        - Очистить сборку"

# Собрать приложение
build:
	go build -o bin/poker cmd/main.go

# Запустить приложение
run: build
	./bin/poker

# Запустить в режиме разработки
dev:
	go run cmd/main.go

# Запустить все сервисы
services-up:
	docker-compose up -d
	@echo "Ожидание запуска сервисов..."
	@sleep 10
	@echo "Все сервисы запущены:"
	@echo "  - PostgreSQL: localhost:5432"
	@echo "  - Kafka: localhost:9092"
	@echo "  - Redis: localhost:6379"

# Остановить все сервисы
services-down:
	docker-compose down

# Запустить только PostgreSQL
db-up:
	docker-compose up -d postgres
	@echo "Ожидание запуска PostgreSQL..."
	@sleep 5
	@echo "PostgreSQL запущен на порту 5432"

# Остановить PostgreSQL
db-down:
	docker-compose stop postgres

# Пересоздать базу данных
db-reset: services-down
	docker-compose down -v
	docker-compose up -d postgres
	@echo "База данных пересоздана"

# Запустить тесты
test:
	go test ./...

# Очистить сборку
clean:
	rm -rf bin/
	go clean