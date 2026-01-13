#!/bin/bash

# Скрипт для первоначальной настройки Let's Encrypt SSL сертификатов

DOMAIN="цветынинасалават.рф"
EMAIL="admin@цветынинасалават.рф" # Замените на ваш email
STAGING=0 # Установите в 1 для тестирования

echo "🔐 Инициализация Let's Encrypt для домена: $DOMAIN"

# Проверяем, существует ли уже сертификат
if [ -d "/etc/letsencrypt/live/$DOMAIN" ]; then
    echo "⚠️  Сертификат для $DOMAIN уже существует. Удалите его для пересоздания."
    exit 1
fi

# Создаем временный nginx конфиг для получения сертификата
echo "📝 Создаем временную конфигурацию nginx..."
cat > /tmp/nginx-temp.conf << EOF
events {
    worker_connections 1024;
}

http {
    server {
        listen 80;
        server_name $DOMAIN www.$DOMAIN;

        location /.well-known/acme-challenge/ {
            root /var/www/certbot;
        }

        location / {
            return 200 'OK';
            add_header Content-Type text/plain;
        }
    }
}
EOF

# Останавливаем nginx если запущен
echo "🛑 Останавливаем nginx..."
docker-compose stop nginx 2>/dev/null || true

# Запускаем временный nginx
echo "🚀 Запускаем временный nginx..."
docker run --rm -d \
    --name nginx-temp \
    -p 80:80 \
    -v /tmp/nginx-temp.conf:/etc/nginx/nginx.conf:ro \
    -v ./certbot/www:/var/www/certbot:ro \
    nginx:alpine

sleep 5

# Получаем сертификат
echo "📜 Получаем SSL сертификат..."

if [ $STAGING != "0" ]; then
    STAGING_ARG="--staging"
    echo "⚠️  Используем staging окружение Let's Encrypt"
else
    STAGING_ARG=""
fi

docker run --rm \
    -v ./certbot/conf:/etc/letsencrypt \
    -v ./certbot/www:/var/www/certbot \
    certbot/certbot \
    certonly --webroot \
    --webroot-path=/var/www/certbot \
    --email $EMAIL \
    --agree-tos \
    --no-eff-email \
    $STAGING_ARG \
    -d $DOMAIN \
    -d www.$DOMAIN

# Останавливаем временный nginx
echo "🛑 Останавливаем временный nginx..."
docker stop nginx-temp

# Проверяем, получен ли сертификат
if [ -d "./certbot/conf/live/$DOMAIN" ]; then
    echo "✅ Сертификат успешно получен!"
    echo "🚀 Теперь можно запустить полную конфигурацию:"
    echo "   docker-compose up -d"
else
    echo "❌ Ошибка получения сертификата"
    exit 1
fi

echo ""
echo "📋 Полезная информация:"
echo "   - Сертификат действителен 90 дней"
echo "   - Автообновление настроено в docker-compose.yml"
echo "   - Логи certbot: docker-compose logs certbot"
echo "   - Проверка сертификата: https://www.ssllabs.com/ssltest/"