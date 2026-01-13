# SSL Setup для Poker API

Автоматическая настройка HTTPS с Let's Encrypt для домена `цветынинасалават.рф`.

## Быстрый старт

### 1. Подготовка DNS
Убедитесь, что A-запись вашего домена указывает на IP сервера:
```bash
# Проверить текущий IP сервера
curl ifconfig.me

# Проверить IP домена
dig +short цветынинасалават.рф
```

### 2. Автоматическая настройка
```bash
# Полная настройка с SSL
./setup-ssl.sh
```

### 3. Ручная настройка (если нужно)

#### Получение SSL сертификата:
```bash
# Создать директории
mkdir -p certbot/conf certbot/www

# Получить сертификат
./nginx/init-letsencrypt.sh

# Запустить все сервисы
docker-compose up -d
```

## Управление

### Перезапуск приложения:
```bash
./rebuild.sh
```

### Проверка SSL:
```bash
./check-ssl.sh
```

### Просмотр логов:
```bash
# Логи nginx
docker-compose logs nginx

# Логи certbot
docker-compose logs certbot

# Логи приложения
docker-compose logs poker-app
```

### Ручное обновление сертификата:
```bash
docker-compose exec certbot certbot renew
docker-compose restart nginx
```

## Доступные адреса

- **API**: https://цветынинасалават.рф
- **Swagger**: https://цветынинасалават.рф/swagger/
- **API Endpoints**: https://цветынинасалават.рф/api/v1/
- **Health Check**: https://цветынинасалават.рф/healthcheck

## Архитектура

```
Internet → nginx (80/443) → poker-app (3000)
                ↓
            Let's Encrypt (SSL)
```

### Компоненты:
- **nginx**: Reverse proxy, SSL termination, статические файлы
- **certbot**: Автоматическое получение и обновление SSL сертификатов
- **poker-app**: Go приложение (внутренний порт 3000)

## Безопасность

### SSL настройки:
- TLS 1.2 и 1.3
- HSTS заголовки
- Безопасные cipher suites
- Автоматическое обновление сертификатов каждые 12 часов

### CORS настройки:
- Разрешены все origins для API
- Поддержка preflight запросов
- Безопасные заголовки

## Troubleshooting

### Проблемы с получением сертификата:
1. Проверьте DNS настройки
2. Убедитесь, что порт 80 открыт
3. Проверьте логи: `docker-compose logs certbot`

### Проблемы с HTTPS:
1. Проверьте статус nginx: `docker-compose ps nginx`
2. Проверьте конфигурацию: `docker-compose exec nginx nginx -t`
3. Перезапустите nginx: `docker-compose restart nginx`

### Проблемы с API:
1. Проверьте статус приложения: `docker-compose ps poker-app`
2. Проверьте логи: `docker-compose logs poker-app`
3. Проверьте подключение к БД: `docker-compose logs postgres`

## Мониторинг

### Проверка статуса сертификата:
```bash
# Информация о сертификате
openssl s_client -servername цветынинасалават.рф -connect цветынинасалават.рф:443 < /dev/null 2>/dev/null | openssl x509 -noout -dates

# SSL Labs тест
# https://www.ssllabs.com/ssltest/analyze.html?d=цветынинасалават.рф
```

### Автоматическое обновление:
Сертификаты обновляются автоматически каждые 12 часов через контейнер certbot.

## Backup

### Важные файлы для резервного копирования:
- `certbot/conf/` - SSL сертификаты и ключи
- `nginx/nginx.conf` - Конфигурация nginx
- `docker-compose.yml` - Конфигурация сервисов