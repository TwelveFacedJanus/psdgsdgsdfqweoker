#!/bin/bash

DOMAIN="цветынинасалават.рф"

echo "🔍 Проверка SSL конфигурации для $DOMAIN"
echo ""

# Проверяем HTTP редирект
echo "1. Проверяем HTTP -> HTTPS редирект:"
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -L "http://$DOMAIN/healthcheck")
echo "   HTTP статус: $HTTP_STATUS"

# Проверяем HTTPS
echo ""
echo "2. Проверяем HTTPS соединение:"
HTTPS_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "https://$DOMAIN/healthcheck")
echo "   HTTPS статус: $HTTPS_STATUS"

# Проверяем сертификат
echo ""
echo "3. Информация о сертификате:"
echo | openssl s_client -servername $DOMAIN -connect $DOMAIN:443 2>/dev/null | openssl x509 -noout -dates

# Проверяем Swagger
echo ""
echo "4. Проверяем Swagger UI:"
SWAGGER_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "https://$DOMAIN/swagger/")
echo "   Swagger статус: $SWAGGER_STATUS"

# Проверяем API
echo ""
echo "5. Проверяем API:"
API_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "https://$DOMAIN/api/v1/public/tables")
echo "   API статус: $API_STATUS"

# SSL Labs тест (ссылка)
echo ""
echo "🔗 Полная проверка SSL:"
echo "   https://www.ssllabs.com/ssltest/analyze.html?d=$DOMAIN"

echo ""
echo "✅ Проверка завершена!"