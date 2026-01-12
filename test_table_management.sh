#!/bin/bash

# Тестовые init_data для игроков
PLAYER1_DATA="auth_date=1768213726&hash=ecc6bdb6e2c4234c6c7f60e11ec379b3b7a1c568c144cdc42394680076b11fee&query_id=test_query_id&user=%7B%22id%22%3A123456789%2C%22first_name%22%3A%22Player1%22%2C%22last_name%22%3A%22Test%22%2C%22username%22%3A%22player1%22%2C%22language_code%22%3A%22en%22%7D"

# Генерируем данные для дополнительных игроков
generate_player_data() {
    local player_id=$1
    local timestamp=$(date +%s)
    echo "auth_date=${timestamp}&hash=test_hash_${player_id}&query_id=test_query_${player_id}&user=%7B%22id%22%3A${player_id}%2C%22first_name%22%3A%22Player${player_id}%22%2C%22username%22%3A%22player${player_id}%22%7D"
}

PLAYER2_DATA=$(generate_player_data 987654321)
PLAYER3_DATA=$(generate_player_data 555666777)

echo "=== Тестирование управления столами ==="
echo

echo "1. Статистика столов до тестирования:"
curl -s -H "x-init-data: $PLAYER1_DATA" http://localhost:3000/api/v1/table-statistics | jq .
echo

echo "2. Доступные столы:"
curl -s -H "x-init-data: $PLAYER1_DATA" http://localhost:3000/api/v1/available-tables | jq .
echo

echo "3. Игрок 1 присоединяется к доступному LOW столу:"
curl -s -X POST -H "x-init-data: $PLAYER1_DATA" \
     -H "Content-Type: application/json" \
     -d '{"category": "LOW"}' \
     http://localhost:3000/api/v1/join-available-table | jq .
echo

echo "4. Игрок 2 присоединяется к доступному LOW столу:"
curl -s -X POST -H "x-init-data: $PLAYER2_DATA" \
     -H "Content-Type: application/json" \
     -d '{"category": "LOW"}' \
     http://localhost:3000/api/v1/join-available-table | jq .
echo

echo "5. Игрок 3 присоединяется к доступному LOW столу (должен создаться новый стол):"
curl -s -X POST -H "x-init-data: $PLAYER3_DATA" \
     -H "Content-Type: application/json" \
     -d '{"category": "LOW"}' \
     http://localhost:3000/api/v1/join-available-table | jq .
echo

echo "6. Статистика столов после присоединения игроков:"
curl -s -H "x-init-data: $PLAYER1_DATA" http://localhost:3000/api/v1/table-statistics | jq .
echo

echo "7. Доступные столы после присоединения:"
curl -s -H "x-init-data: $PLAYER1_DATA" http://localhost:3000/api/v1/available-tables | jq .
echo

echo "8. Все столы LOW категории:"
curl -s http://localhost:3000/api/v1/public/tables?category=LOW | jq .
echo

echo "9. Тестирование MID категории:"
curl -s -X POST -H "x-init-data: $PLAYER1_DATA" \
     -H "Content-Type: application/json" \
     -d '{"category": "MID"}' \
     http://localhost:3000/api/v1/join-available-table | jq .
echo

echo "10. Финальная статистика столов:"
curl -s -H "x-init-data: $PLAYER1_DATA" http://localhost:3000/api/v1/table-statistics | jq .
echo

echo "11. Очистка пустых столов:"
curl -s -X POST -H "x-init-data: $PLAYER1_DATA" http://localhost:3000/api/v1/cleanup-empty-tables | jq .
echo

echo "=== Тестирование завершено ==="