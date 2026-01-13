#!/bin/bash

DOMAIN="цветынинасалават.рф"

echo "🚀 Настройка Poker API с SSL для домена: $DOMAIN"
echo ""

# Проверяем, что домен указывает на этот сервер
echo "🔍 Проверяем DNS настройки..."
CURRENT_IP=$(curl -s ifconfig.me)
DOMAIN_IP=$(dig +short $DOMAIN | tail -n1)

echo "   Текущий IP сервера: $CURRENT_IP"
echo "   IP домена $DOMAIN: $DOMAIN_IP"

if [ "$CURRENT_IP" != "$DOMAIN_IP" ]; then
    echo "⚠️  ВНИМАНИЕ: Домен не указывает на этот сервер!"
    echo "   Убедитесь, что A-запись домена указывает на $CURRENT_IP"
    read -p "Продолжить? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Создаем необходимые директории
echo "📁 Создаем директории..."
mkdir -p certbot/conf
mkdir -p certbot/www

# Останавливаем существующие контейнеры
echo "🛑 Останавливаем существующие контейнеры..."
docker-compose down

# Запускаем базовые сервисы (без nginx)
echo "🚀 Запускаем базовые сервисы..."
docker-compose up -d postgres redis kafka zookeeper poker-app

# Ждем запуска сервисов
echo "⏳ Ждем запуска сервисов..."
sleep 30

# Получаем SSL сертификат
echo "🔐 Получаем SSL сертификат..."
./nginx/init-letsencrypt.sh

if [ $? -eq 0 ]; then
    echo "✅ SSL сертификат получен успешно!"
    
    # Запускаем nginx
    echo "🚀 Запускаем nginx с SSL..."
    docker-compose up -d nginx certbot
    
    # Проверяем статус
    echo "📊 Статус сервисов:"
    docker-compose ps
    
    echo ""
    echo "🎉 Настройка завершена!"
    echo ""
    echo "🌐 Ваш API доступен по адресам:"
    echo "   HTTPS: https://$DOMAIN"
    echo "   Swagger: https://$DOMAIN/swagger/"
    echo "   API: https://$DOMAIN/api/v1/"
    echo ""
    echo "🔒 SSL сертификат будет автоматически обновляться каждые 12 часов"
    
else
    echo "❌ Ошибка получения SSL сертификата"
    echo "🔧 Запускаем без SSL (только HTTP)..."
    
    # Создаем временную конфигурацию nginx без SSL
    cat > nginx/nginx-http.conf << 'EOF'
events {
    worker_connections 1024;
}

http {
    upstream poker_app {
        server poker-app:3000;
    }

    server {
        listen 80;
        server_name цветынинасалават.рф www.цветынинасалават.рф;

        location / {
            proxy_pass http://poker_app;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
EOF
    
    # Запускаем nginx с HTTP конфигурацией
    docker run -d \
        --name poker_nginx_http \
        --network poker_poker-network \
        -p 80:80 \
        -v $(pwd)/nginx/nginx-http.conf:/etc/nginx/nginx.conf:ro \
        nginx:alpine
    
    echo "🌐 API доступен по HTTP: http://$DOMAIN"
fi