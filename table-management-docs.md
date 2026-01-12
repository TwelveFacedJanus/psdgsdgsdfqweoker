# Управление столами - Документация

## Автоматическое создание столов

Система автоматически создает новые столы в следующих случаях:

### 1. При присоединении к столу
Когда игрок присоединяется к столу, где уже есть минимум 2 игрока, система автоматически создает новый стол той же категории.

### 2. Менеджер столов (каждые 30 секунд)
- Проверяет наличие доступных столов в каждой категории
- Создает новый стол, если нет доступных
- Создает дополнительные столы, если все доступные заполнены более чем на 50%

## API Endpoints

### Присоединиться к доступному столу

```
POST /api/v1/join-available-table
```

**Требует авторизации**: Да  
**Описание**: Автоматически находит доступный стол указанной категории или создает новый

**Тело запроса**:
```json
{
  "category": "LOW|MID|VIP"
}
```

**Пример запроса**:
```bash
curl -X POST \
  -H "x-init-data: YOUR_INIT_DATA" \
  -H "Content-Type: application/json" \
  -d '{"category": "LOW"}' \
  http://localhost:3000/api/v1/join-available-table
```

**Пример ответа**:
```json
{
  "message": "Successfully joined table",
  "table": {
    "id": 4,
    "category": "LOW",
    "blinds": "1/2",
    "buy_in": 50,
    "players": 1,
    "max_seats": 6
  },
  "seat": 1,
  "chips": 50,
  "balance": 950,
  "additional_table_created": true,
  "additional_table": {
    "id": 5,
    "category": "LOW",
    "blinds": "1/2",
    "buy_in": 50,
    "players": 0,
    "max_seats": 6
  }
}
```

### Получить доступные столы

```
GET /api/v1/available-tables?category={ALL|LOW|MID|VIP}
```

**Требует авторизации**: Да  
**Описание**: Возвращает все доступные столы, сгруппированные по категориям

**Пример ответа**:
```json
{
  "available_tables": {
    "LOW": [
      {
        "id": 1,
        "category": "LOW",
        "blinds": "1/2",
        "buy_in": 50,
        "players": 3,
        "max_seats": 6
      },
      {
        "id": 4,
        "category": "LOW",
        "blinds": "1/2",
        "buy_in": 50,
        "players": 0,
        "max_seats": 6
      }
    ],
    "MID": [...],
    "VIP": [...]
  },
  "total_tables": 5
}
```

### Статистика столов

```
GET /api/v1/table-statistics
```

**Требует авторизации**: Да  
**Описание**: Возвращает статистику по всем категориям столов

**Пример ответа**:
```json
{
  "statistics": {
    "LOW": {
      "total_tables": 2,
      "available_tables": 2,
      "total_players": 3
    },
    "MID": {
      "total_tables": 2,
      "available_tables": 2,
      "total_players": 5
    },
    "VIP": {
      "total_tables": 1,
      "available_tables": 1,
      "total_players": 2
    }
  }
}
```

### Очистка пустых столов

```
POST /api/v1/cleanup-empty-tables
```

**Требует авторизации**: Да  
**Описание**: Принудительно удаляет пустые столы (оставляет по одному в каждой категории)

## Логика создания столов

### Параметры столов по категориям

| Категория | Blinds | Buy-in | Max Seats |
|-----------|--------|--------|-----------|
| LOW       | 1/2    | 50     | 6         |
| MID       | 5/10   | 200    | 9         |
| VIP       | 25/50  | 1000   | 6         |

### Алгоритм выбора стола

1. **Поиск доступного стола**: Ищет столы с `players < max_seats`
2. **Приоритет заполненности**: Предпочитает столы с большим количеством игроков
3. **Создание нового стола**: Если нет доступных столов, создает новый
4. **Автоматическое расширение**: При достижении 2+ игроков создает дополнительный стол

### События Kafka

#### table_created
```json
{
  "type": "table_created",
  "table_id": 4,
  "data": {
    "id": 4,
    "category": "LOW",
    "blinds": "1/2",
    "buy_in": 50,
    "players": 0,
    "max_seats": 6
  }
}
```

#### table_auto_created
```json
{
  "type": "table_auto_created",
  "table_id": 5,
  "data": {
    "id": 5,
    "category": "MID",
    "blinds": "5/10",
    "buy_in": 200,
    "players": 0,
    "max_seats": 9
  }
}
```

#### player_joined
```json
{
  "type": "player_joined",
  "table_id": 1,
  "data": {
    "user": {
      "uuid": "user-uuid",
      "username": "player1"
    },
    "table": {
      "id": 1,
      "category": "LOW"
    },
    "seat_number": 3
  }
}
```

## Мониторинг и управление

### Автоматические процессы

- **Проверка каждые 30 секунд**: Менеджер столов автоматически создает недостающие столы
- **Очистка пустых столов**: Удаляет лишние пустые столы (оставляет по одному)
- **Балансировка нагрузки**: Распределяет игроков по столам оптимально

### Ручное управление

- Принудительная очистка через API
- Мониторинг статистики в реальном времени
- События в Kafka для внешних систем

## Примеры использования

### Быстрое присоединение к игре

```javascript
// Присоединиться к любому доступному LOW столу
const response = await fetch('/api/v1/join-available-table', {
  method: 'POST',
  headers: {
    'x-init-data': initData,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({ category: 'LOW' })
});

const result = await response.json();
console.log('Joined table:', result.table.id);

if (result.additional_table_created) {
  console.log('New table created:', result.additional_table.id);
}
```

### Мониторинг столов

```javascript
// Получить статистику
const stats = await fetch('/api/v1/table-statistics', {
  headers: { 'x-init-data': initData }
}).then(r => r.json());

console.log('Total LOW tables:', stats.statistics.LOW.total_tables);
console.log('Available LOW tables:', stats.statistics.LOW.available_tables);
```

## Преимущества системы

1. **Автоматическое масштабирование**: Столы создаются по мере необходимости
2. **Оптимальное заполнение**: Игроки направляются в наиболее заполненные столы
3. **Минимальное ожидание**: Всегда есть доступные столы для новых игроков
4. **Эффективное использование ресурсов**: Пустые столы автоматически удаляются
5. **Реальное время**: События через Kafka для мгновенных обновлений UI