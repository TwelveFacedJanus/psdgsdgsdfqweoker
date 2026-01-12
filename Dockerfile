# Используем официальный образ Go
FROM golang:1.25-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go mod и sum файлы
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Генерируем Swagger документацию
RUN go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/main.go -o docs

# Финальный образ
FROM alpine:latest

# Устанавливаем ca-certificates для HTTPS запросов
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарный файл из builder образа
COPY --from=builder /app/main .

# Копируем необходимые файлы
COPY --from=builder /app/docs ./docs/
COPY --from=builder /app/.env .env

# Открываем порт
EXPOSE 3000

# Запускаем приложение
CMD ["./main"]