#!/bin/bash

# Тестовые init_data для двух игроков
PLAYER1_DATA="auth_date=1768213726&hash=ecc6bdb6e2c4234c6c7f60e11ec379b3b7a1c568c144cdc42394680076b11fee&query_id=test_query_id&user=%7B%22id%22%3A123456789%2C%22first_name%22%3A%22Test%22%2C%22last_name%22%3A%22User%22%2C%22username%22%3A%22testuser%22%2C%22language_code%22%3A%22en%22%7D"

# Генерируем второго игрока
go run utils/generate_test_init_data.go > /tmp/player2_data.txt
PLAYER2_DATA=$(grep "auth_date=" /tmp/player2_data.txt)

echo "=== Тестирование игры в покер ==="
echo

echo "1. Игрок 1 - профиль:"
curl -s -H "x-init-data: $PLAYER1_DATA" http://localhost:3000/api/v1/profile | jq .
echo

echo "2. Игрок 2 - профиль:"
curl -s -H "x-init-data: $PLAYER2_DATA" http://localhost:3000/api/v1/profile | jq .
echo

echo "3. Игрок 1 присоединяется к столу 1:"
curl -s -X POST -H "x-init-data: $PLAYER1_DATA" http://localhost:3000/api/v1/tables/1/join | jq .
echo

echo "4. Игрок 2 присоединяется к столу 1:"
curl -s -X POST -H "x-init-data: $PLAYER2_DATA" http://localhost:3000/api/v1/tables/1/join | jq .
echo

echo "5. Игроки за столом:"
curl -s http://localhost:3000/api/v1/public/tables/1/players | jq .
echo

echo "6. Игрок 1 начинает игру:"
GAME_RESPONSE=$(curl -s -X POST -H "x-init-data: $PLAYER1_DATA" http://localhost:3000/api/v1/tables/1/start-game)
echo $GAME_RESPONSE | jq .
GAME_ID=$(echo $GAME_RESPONSE | jq -r '.game.id')
echo "Game ID: $GAME_ID"
echo

if [ "$GAME_ID" != "null" ] && [ "$GAME_ID" != "" ]; then
    echo "7. Состояние игры:"
    curl -s -H "x-init-data: $PLAYER1_DATA" http://localhost:3000/api/v1/games/$GAME_ID | jq .
    echo

    echo "8. Игрок 1 делает call:"
    curl -s -X POST -H "x-init-data: $PLAYER1_DATA" \
         -H "Content-Type: application/json" \
         -d '{"action": "call"}' \
         http://localhost:3000/api/v1/games/$GAME_ID/action | jq .
    echo

    echo "9. Игрок 2 делает raise:"
    curl -s -X POST -H "x-init-data: $PLAYER2_DATA" \
         -H "Content-Type: application/json" \
         -d '{"action": "raise", "amount": 10}' \
         http://localhost:3000/api/v1/games/$GAME_ID/action | jq .
    echo

    echo "10. История игры:"
    curl -s -H "x-init-data: $PLAYER1_DATA" http://localhost:3000/api/v1/games/$GAME_ID/history | jq .
    echo

    echo "11. Активные игры игрока 1:"
    curl -s -H "x-init-data: $PLAYER1_DATA" http://localhost:3000/api/v1/my-games | jq .
    echo
fi

echo "=== Тестирование завершено ==="