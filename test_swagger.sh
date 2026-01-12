#!/bin/bash

echo "=== Тестирование Swagger документации ==="
echo

echo "1. Проверка доступности Swagger UI:"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/swagger/)
if [ "$HTTP_CODE" = "200" ]; then
    echo "✅ Swagger UI доступен: http://localhost:3000/swagger/"
else
    echo "❌ Swagger UI недоступен (HTTP $HTTP_CODE)"
fi
echo

echo "2. Проверка swagger.json:"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/swagger/swagger.json)
if [ "$HTTP_CODE" = "200" ]; then
    echo "✅ swagger.json доступен"
    echo "Информация об API:"
    curl -s http://localhost:3000/swagger/swagger.json | jq '.info'
else
    echo "❌ swagger.json недоступен (HTTP $HTTP_CODE)"
fi
echo

echo "3. Проверка swagger.yaml:"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/swagger/swagger.yaml)
if [ "$HTTP_CODE" = "200" ]; then
    echo "✅ swagger.yaml доступен"
else
    echo "❌ swagger.yaml недоступен (HTTP $HTTP_CODE)"
fi
echo

echo "4. Проверка основных endpoints в документации:"
ENDPOINTS=$(curl -s http://localhost:3000/swagger/swagger.json | jq -r '.paths | keys[]')
echo "Документированные endpoints:"
echo "$ENDPOINTS" | while read endpoint; do
    echo "  📍 $endpoint"
done
echo

echo "5. Проверка моделей данных:"
MODELS=$(curl -s http://localhost:3000/swagger/swagger.json | jq -r '.definitions | keys[]')
echo "Определенные модели:"
echo "$MODELS" | while read model; do
    echo "  📋 $model"
done
echo

echo "=== Swagger документация готова к использованию! ==="
echo "🌐 Откройте http://localhost:3000/swagger/ в браузере"
echo "🔑 Используйте кнопку 'Set Telegram Auth' для авторизации"
echo "🧪 Тестируйте API endpoints прямо в интерфейсе"