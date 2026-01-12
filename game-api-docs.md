# Poker Game API Documentation

## Игровые маршруты

### Начать игру за столом

```
POST /api/v1/tables/:id/start-game
```

**Требует авторизации**: Да  
**Описание**: Начинает новую игру за указанным столом

**Пример запроса**:
```bash
curl -X POST \
  -H "x-init-data: YOUR_INIT_DATA" \
  http://localhost:3000/api/v1/tables/1/start-game
```

**Пример ответа**:
```json
{
  "message": "Game started successfully",
  "game": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "table_id": 1,
    "state": "waiting",
    "community_cards": [],
    "pot": 0,
    "current_bet": 0,
    "dealer_position": 0,
    "current_player": 0,
    "small_blind": 1,
    "big_blind": 2,
    "players": [
      {
        "id": 1,
        "user_uuid": "user-uuid-1",
        "position": 0,
        "chips": 50,
        "bet": 0,
        "is_folded": false,
        "is_all_in": false
      }
    ]
  }
}
```

### Получить состояние игры

```
GET /api/v1/games/:gameId
```

**Требует авторизации**: Да  
**Описание**: Возвращает текущее состояние игры

**Пример запроса**:
```bash
curl -H "x-init-data: YOUR_INIT_DATA" \
  http://localhost:3000/api/v1/games/123e4567-e89b-12d3-a456-426614174000
```

### Сделать ход в игре

```
POST /api/v1/games/:gameId/action
```

**Требует авторизации**: Да  
**Описание**: Выполняет действие игрока в игре

**Тело запроса**:
```json
{
  "action": "call|raise|fold|check|bet",
  "amount": 100
}
```

**Доступные действия**:
- `fold` - сброс карт
- `call` - уравнять ставку
- `raise` - повысить ставку (требует amount)
- `check` - пас (только если нет ставки)
- `bet` - поставить (требует amount, только если нет ставки)

**Пример запроса**:
```bash
curl -X POST \
  -H "x-init-data: YOUR_INIT_DATA" \
  -H "Content-Type: application/json" \
  -d '{"action": "call"}' \
  http://localhost:3000/api/v1/games/123e4567-e89b-12d3-a456-426614174000/action
```

### Получить историю игры

```
GET /api/v1/games/:gameId/history
```

**Требует авторизации**: Да  
**Описание**: Возвращает историю всех действий в игре

### Получить активные игры пользователя

```
GET /api/v1/my-games
```

**Требует авторизации**: Да  
**Описание**: Возвращает все активные игры пользователя

## Состояния игры

1. **waiting** - Ожидание начала игры
2. **preflop** - Префлоп (карты розданы, торговля до флопа)
3. **flop** - Флоп (3 общие карты открыты)
4. **turn** - Терн (4-я общая карта открыта)
5. **river** - Ривер (5-я общая карта открыта)
6. **showdown** - Вскрытие карт
7. **finished** - Игра завершена

## События Kafka

### Топики

- `poker-game-events` - События игры
- `poker-table-events` - События столов

### Типы событий

#### game_started
```json
{
  "type": "game_started",
  "game_id": "123e4567-e89b-12d3-a456-426614174000",
  "table_id": 1,
  "data": {
    "game": "полное состояние игры"
  },
  "timestamp": "2026-01-12T15:30:00Z"
}
```

#### player_action
```json
{
  "type": "player_action",
  "game_id": "123e4567-e89b-12d3-a456-426614174000",
  "table_id": 1,
  "user_uuid": "user-uuid",
  "data": {
    "action": "call",
    "amount": 50,
    "player": {
      "position": 0,
      "chips": 100,
      "bet": 50
    }
  },
  "timestamp": "2026-01-12T15:30:00Z"
}
```

#### game_state_changed
```json
{
  "type": "game_state_changed",
  "game_id": "123e4567-e89b-12d3-a456-426614174000",
  "table_id": 1,
  "data": {
    "state": "flop",
    "community_cards": [
      {"suit": "hearts", "rank": "A", "value": 14},
      {"suit": "diamonds", "rank": "K", "value": 13},
      {"suit": "clubs", "rank": "Q", "value": 12}
    ],
    "pot": 150,
    "current_bet": 50,
    "current_player": 1
  },
  "timestamp": "2026-01-12T15:30:00Z"
}
```

## Кэширование в Redis

### Ключи Redis

- `game:{gameId}` - Состояние игры
- `player_session:{userUUID}` - Сессия игрока
- `table_players:{tableId}` - Игроки за столом
- `table_lock:{tableId}` - Блокировка стола

## Комбинации в покере

1. **Старшая карта** (High Card) - Rank: 1
2. **Пара** (Pair) - Rank: 2
3. **Две пары** (Two Pair) - Rank: 3
4. **Тройка** (Three of a Kind) - Rank: 4
5. **Стрит** (Straight) - Rank: 5
6. **Флеш** (Flush) - Rank: 6
7. **Фулл хаус** (Full House) - Rank: 7
8. **Каре** (Four of a Kind) - Rank: 8
9. **Стрит флеш** (Straight Flush) - Rank: 9
10. **Роял флеш** (Royal Flush) - Rank: 10

## Пример полного игрового процесса

### 1. Присоединиться к столу
```bash
curl -X POST -H "x-init-data: YOUR_INIT_DATA" \
  http://localhost:3000/api/v1/tables/1/join
```

### 2. Начать игру
```bash
curl -X POST -H "x-init-data: YOUR_INIT_DATA" \
  http://localhost:3000/api/v1/tables/1/start-game
```

### 3. Сделать ход
```bash
curl -X POST -H "x-init-data: YOUR_INIT_DATA" \
  -H "Content-Type: application/json" \
  -d '{"action": "call"}' \
  http://localhost:3000/api/v1/games/GAME_ID/action
```

### 4. Проверить состояние игры
```bash
curl -H "x-init-data: YOUR_INIT_DATA" \
  http://localhost:3000/api/v1/games/GAME_ID
```

## Ошибки

### Игровые ошибки

- `Game already in progress` - Игра уже идет за столом
- `Need at least 2 players to start game` - Нужно минимум 2 игрока
- `It's not your turn` - Не ваш ход
- `Invalid action` - Недопустимое действие
- `Insufficient chips` - Недостаточно фишек

### Примеры ошибок

```json
{
  "error": "It's not your turn or you cannot act"
}
```

```json
{
  "error": "Need at least 2 players to start game"
}
```