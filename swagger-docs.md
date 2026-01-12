# Swagger API Documentation

## Доступ к документации

Swagger UI доступен по адресу: **http://localhost:3000/swagger/**

## Форматы документации

- **Swagger UI**: http://localhost:3000/swagger/ - Интерактивная документация
- **JSON**: http://localhost:3000/swagger/swagger.json - OpenAPI спецификация в JSON
- **YAML**: http://localhost:3000/swagger/swagger.yaml - OpenAPI спецификация в YAML

## Авторизация в Swagger UI

### Для тестирования защищенных endpoints:

1. Откройте Swagger UI: http://localhost:3000/swagger/
2. Нажмите кнопку **"Set Telegram Auth"** в верхней части страницы
3. Введите ваш `init_data` от Telegram Web App
4. Теперь вы можете тестировать защищенные endpoints

### Получение тестового init_data:

```bash
# Сгенерировать тестовый init_data
go run utils/generate_test_init_data.go
```

Скопируйте полученную строку и используйте её в Swagger UI.

## Категории API

### 🔧 System
- `GET /healthcheck` - Проверка состояния сервера

### 👤 Users (требует авторизации)
- `GET /profile` - Получить профиль пользователя
- `PUT /profile` - Обновить профиль пользователя

### 🎲 Tables
**Публичные:**
- `GET /tables` - Получить список столов
- `GET /tables/{id}` - Получить стол по ID
- `GET /public/tables` - Публичный доступ к столам
- `GET /public/tables/{id}` - Публичный доступ к столу
- `GET /public/tables/{id}/players` - Игроки за столом

**Защищенные (требуют авторизации):**
- `POST /tables/{id}/join` - Присоединиться к конкретному столу
- `POST /tables/{id}/leave` - Покинуть стол
- `POST /join-available-table` - Присоединиться к доступному столу
- `GET /available-tables` - Получить доступные столы
- `GET /table-statistics` - Статистика столов
- `POST /cleanup-empty-tables` - Очистить пустые столы

### 🎮 Game (требует авторизации)
- `POST /tables/{id}/start-game` - Начать игру за столом
- `GET /games/{gameId}` - Получить состояние игры
- `POST /games/{gameId}/action` - Сделать ход в игре
- `GET /games/{gameId}/history` - История игры
- `GET /my-games` - Мои активные игры

## Модели данных

### User
```json
{
  "uuid": "string",
  "username": "string", 
  "telegram_id": "integer",
  "balance": "integer",
  "created_at": "string (date-time)",
  "updated_at": "string (date-time)"
}
```

### Table
```json
{
  "id": "integer",
  "category": "string (LOW|MID|VIP)",
  "blinds": "string",
  "buy_in": "integer", 
  "players": "integer",
  "max_seats": "integer",
  "created_at": "string (date-time)",
  "updated_at": "string (date-time)"
}
```

### Game
```json
{
  "id": "string",
  "table_id": "integer",
  "state": "string (waiting|preflop|flop|turn|river|showdown|finished)",
  "community_cards": "array of Card",
  "pot": "integer",
  "current_bet": "integer",
  "dealer_position": "integer",
  "current_player": "integer",
  "small_blind": "integer",
  "big_blind": "integer",
  "players": "array of GamePlayer",
  "created_at": "string (date-time)",
  "updated_at": "string (date-time)"
}
```

### Card
```json
{
  "suit": "string (hearts|diamonds|clubs|spades)",
  "rank": "string (2-10|J|Q|K|A)",
  "value": "integer (2-14)"
}
```

## Примеры использования

### 1. Получить все столы
```bash
curl -X GET "http://localhost:3000/api/v1/tables?category=ALL"
```

### 2. Присоединиться к доступному столу
```bash
curl -X POST "http://localhost:3000/api/v1/join-available-table" \
  -H "x-init-data: YOUR_INIT_DATA" \
  -H "Content-Type: application/json" \
  -d '{"category": "LOW"}'
```

### 3. Начать игру
```bash
curl -X POST "http://localhost:3000/api/v1/tables/1/start-game" \
  -H "x-init-data: YOUR_INIT_DATA"
```

### 4. Сделать ход в игре
```bash
curl -X POST "http://localhost:3000/api/v1/games/{gameId}/action" \
  -H "x-init-data: YOUR_INIT_DATA" \
  -H "Content-Type: application/json" \
  -d '{"action": "call"}'
```

## Коды ответов

- **200** - Успешный запрос
- **400** - Неверный запрос (Bad Request)
- **401** - Не авторизован (Unauthorized)
- **404** - Не найдено (Not Found)
- **500** - Внутренняя ошибка сервера (Internal Server Error)

## Авторизация

API использует кастомную авторизацию через заголовок `x-init-data`, который содержит данные от Telegram Web App.

### Формат заголовка:
```
x-init-data: auth_date=1234567890&hash=abc123...&query_id=xyz&user=%7B%22id%22%3A123...
```

### Безопасность:
- Подпись проверяется через HMAC-SHA256
- Срок действия: не более 24 часов
- Автоматическое создание пользователей при первом входе

## Особенности Swagger UI

1. **Интерактивное тестирование** - Можно выполнять запросы прямо из интерфейса
2. **Автоматическая авторизация** - После установки init_data все защищенные запросы будут авторизованы
3. **Валидация данных** - Swagger проверяет корректность входных данных
4. **Примеры ответов** - Показывает структуру ожидаемых ответов
5. **Группировка по тегам** - API endpoints сгруппированы по функциональности

## Обновление документации

При изменении API необходимо обновить файлы:
- `docs/swagger.json`
- `docs/swagger.yaml`
- `docs/docs.go`

Или использовать автоматическую генерацию:
```bash
swag init -g cmd/main.go -o docs
```