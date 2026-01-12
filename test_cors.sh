#!/bin/bash

SERVER_URL="http://208.123.185.204:3000"

echo "🧪 Тестирование CORS для $SERVER_URL"
echo ""

# Тест OPTIONS запроса
echo "1. Тестируем OPTIONS запрос:"
curl -X OPTIONS \
  -H "Origin: http://example.com" \
  -H "Access-Control-Request-Method: GET" \
  -H "Access-Control-Request-Headers: Content-Type" \
  -v "$SERVER_URL/api/v1/profile" 2>&1 | grep -E "(< HTTP|< Access-Control|< Allow)"

echo ""
echo "2. Тестируем GET запрос с CORS заголовками:"
curl -X GET \
  -H "Origin: http://example.com" \
  -v "$SERVER_URL/healthcheck" 2>&1 | grep -E "(< HTTP|< Access-Control)"

echo ""
echo "3. Проверяем доступность Swagger:"
curl -s -o /dev/null -w "HTTP Status: %{http_code}\n" "$SERVER_URL/swagger/"

echo ""
echo "✅ Тест завершен!"