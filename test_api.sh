#!/bin/bash

# Тестовый init_data
INIT_DATA="auth_date=1768213726&hash=ecc6bdb6e2c4234c6c7f60e11ec379b3b7a1c568c144cdc42394680076b11fee&query_id=test_query_id&user=%7B%22id%22%3A123456789%2C%22first_name%22%3A%22Test%22%2C%22last_name%22%3A%22User%22%2C%22username%22%3A%22testuser%22%2C%22language_code%22%3A%22en%22%7D"

echo "=== Тестирование Poker API ==="
echo

echo "1. Публичные столы (без авторизации):"
curl -s http://localhost:3000/api/v1/public/tables | jq .
echo

echo "2. Профиль пользователя (с авторизацией):"
curl -s -H "x-init-data: $INIT_DATA" http://localhost:3000/api/v1/profile | jq .
echo

echo "3. Присоединение к столу LOW (ID: 1):"
curl -s -X POST -H "x-init-data: $INIT_DATA" http://localhost:3000/api/v1/tables/1/join | jq .
echo

echo "4. Мои столы:"
curl -s -H "x-init-data: $INIT_DATA" http://localhost:3000/api/v1/my-tables | jq .
echo

echo "5. Игроки за столом 1:"
curl -s http://localhost:3000/api/v1/public/tables/1/players | jq .
echo

echo "6. Покидание стола:"
curl -s -X POST -H "x-init-data: $INIT_DATA" http://localhost:3000/api/v1/tables/1/leave | jq .
echo

echo "7. Проверка баланса после покидания:"
curl -s -H "x-init-data: $INIT_DATA" http://localhost:3000/api/v1/profile | jq .
echo

echo "=== Тестирование завершено ==="