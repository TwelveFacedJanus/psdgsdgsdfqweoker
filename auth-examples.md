# Примеры использования API с авторизацией

## Получение init_data от Telegram Web App

В Telegram Web App вы получаете `initData` через:

```javascript
// В Telegram Web App
const initData = window.Telegram.WebApp.initData;
```

## Примеры запросов с авторизацией

### 1. Получить профиль пользователя

```bash
curl -H "x-init-data: query_id=AAHdF6IQAAAAAN0XohDhrOrc&user=%7B%22id%22%3A279058397%2C%22first_name%22%3A%22Vladislav%22%2C%22last_name%22%3A%22Kibenko%22%2C%22username%22%3A%22vdkfrost%22%2C%22language_code%22%3A%22ru%22%7D&auth_date=1662771648&hash=c501b71e775f74ce10e377dea85a7ea24ecd640b223ea86dfe453e0eaed2e2b2" \
     http://localhost:3000/api/v1/profile
```

### 2. Присоединиться к столу

```bash
curl -X POST \
     -H "x-init-data: query_id=AAHdF6IQAAAAAN0XohDhrOrc&user=%7B%22id%22%3A279058397%2C%22first_name%22%3A%22Vladislav%22%2C%22last_name%22%3A%22Kibenko%22%2C%22username%22%3A%22vdkfrost%22%2C%22language_code%22%3A%22ru%22%7D&auth_date=1662771648&hash=c501b71e775f74ce10e377dea85a7ea24ecd640b223ea86dfe453e0eaed2e2b2" \
     http://localhost:3000/api/v1/tables/1/join
```

### 3. Получить мои столы

```bash
curl -H "x-init-data: query_id=AAHdF6IQAAAAAN0XohDhrOrc&user=%7B%22id%22%3A279058397%2C%22first_name%22%3A%22Vladislav%22%2C%22last_name%22%3A%22Kibenko%22%2C%22username%22%3A%22vdkfrost%22%2C%22language_code%22%3A%22ru%22%7D&auth_date=1662771648&hash=c501b71e775f74ce10e377dea85a7ea24ecd640b223ea86dfe453e0eaed2e2b2" \
     http://localhost:3000/api/v1/my-tables
```

## JavaScript пример для фронтенда

```javascript
// Получаем init_data от Telegram
const initData = window.Telegram.WebApp.initData;

// Функция для выполнения авторизованных запросов
async function apiRequest(url, options = {}) {
    const headers = {
        'Content-Type': 'application/json',
        'x-init-data': initData,
        ...options.headers
    };

    const response = await fetch(url, {
        ...options,
        headers
    });

    if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
    }

    return response.json();
}

// Примеры использования
async function examples() {
    try {
        // Получить профиль
        const profile = await apiRequest('/api/v1/profile');
        console.log('Profile:', profile);

        // Присоединиться к столу
        const joinResult = await apiRequest('/api/v1/tables/1/join', {
            method: 'POST'
        });
        console.log('Join result:', joinResult);

        // Получить мои столы
        const myTables = await apiRequest('/api/v1/my-tables');
        console.log('My tables:', myTables);

        // Покинуть стол
        const leaveResult = await apiRequest('/api/v1/tables/1/leave', {
            method: 'POST'
        });
        console.log('Leave result:', leaveResult);

    } catch (error) {
        console.error('API Error:', error);
    }
}
```

## Публичные маршруты (без авторизации)

```bash
# Получить все столы
curl http://localhost:3000/api/v1/public/tables

# Получить конкретный стол
curl http://localhost:3000/api/v1/public/tables/1

# Получить игроков за столом
curl http://localhost:3000/api/v1/public/tables/1/players
```

## Структура ответов

### Профиль пользователя
```json
{
  "user": {
    "uuid": "123e4567-e89b-12d3-a456-426614174000",
    "username": "Vladislav Kibenko",
    "telegram_id": 279058397,
    "balance": 1000,
    "created_at": "2026-01-12T10:00:00Z",
    "updated_at": "2026-01-12T10:00:00Z"
  }
}
```

### Присоединение к столу
```json
{
  "message": "Successfully joined table",
  "table": {
    "id": 1,
    "category": "LOW",
    "blinds": "1/2",
    "buy_in": 50,
    "players": 4,
    "max_seats": 6
  },
  "seat": 3,
  "chips": 50,
  "balance": 950
}
```

### Мои столы
```json
{
  "tables": [
    {
      "id": 1,
      "table_id": 1,
      "user_uuid": "123e4567-e89b-12d3-a456-426614174000",
      "seat_number": 3,
      "chips": 50,
      "joined_at": "2026-01-12T10:30:00Z",
      "table": {
        "id": 1,
        "category": "LOW",
        "blinds": "1/2",
        "buy_in": 50,
        "players": 4,
        "max_seats": 6
      }
    }
  ]
}
```

## Ошибки авторизации

### Отсутствует заголовок
```json
{
  "error": "Missing x-init-data header"
}
```

### Неверная подпись
```json
{
  "error": "Invalid init_data: invalid hash"
}
```

### Истек срок действия
```json
{
  "error": "Init data expired"
}
```

### Недостаточно средств
```json
{
  "error": "Insufficient balance for buy-in"
}
```