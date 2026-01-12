package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Загружаем переменные окружения
	if err := godotenv.Load(); err != nil {
		fmt.Println("Файл .env не найден, используем переменные окружения системы")
	}

	// Получаем BOT_TOKEN из переменных окружения
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		fmt.Println("Ошибка: BOT_TOKEN не установлен в .env файле")
		return
	}
	userID := 123456789
	firstName := "Test"
	lastName := "User"
	username := "testuser"
	authDate := time.Now().Unix()

	// Формируем данные пользователя
	userData := fmt.Sprintf(`{"id":%d,"first_name":"%s","last_name":"%s","username":"%s","language_code":"en"}`,
		userID, firstName, lastName, username)

	// Создаем параметры
	params := map[string]string{
		"query_id":  "test_query_id",
		"user":      userData,
		"auth_date": strconv.FormatInt(authDate, 10),
	}

	// Сортируем параметры
	var keys []string
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Создаем строку для подписи
	var dataCheckString []string
	for _, key := range keys {
		dataCheckString = append(dataCheckString, fmt.Sprintf("%s=%s", key, params[key]))
	}
	dataCheckStr := strings.Join(dataCheckString, "\n")

	// Создаем секретный ключ
	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(botToken))
	secret := secretKey.Sum(nil)

	// Создаем подпись
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(dataCheckStr))
	hash := hex.EncodeToString(h.Sum(nil))

	// Добавляем hash к параметрам
	params["hash"] = hash

	// Формируем URL-encoded строку
	values := url.Values{}
	for key, value := range params {
		values.Set(key, value)
	}

	initData := values.Encode()

	fmt.Println("Тестовый init_data:")
	fmt.Println(initData)
	fmt.Println()
	fmt.Println("Пример использования:")
	fmt.Printf("curl -H \"x-init-data: %s\" http://localhost:3000/api/v1/profile\n", initData)
	fmt.Println()
	fmt.Printf("Используется BOT_TOKEN: %s\n", botToken[:10]+"...")
}