package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

func validateInitDataWithToken(initData, botToken string) bool {
	// Парсим URL-encoded данные
	values, err := url.ParseQuery(initData)
	if err != nil {
		return false
	}

	// Извлекаем hash
	hash := values.Get("hash")
	if hash == "" {
		return false
	}

	// Удаляем hash из параметров для проверки
	values.Del("hash")

	// Сортируем параметры и создаем строку для проверки
	var keys []string
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var dataCheckString []string
	for _, key := range keys {
		dataCheckString = append(dataCheckString, fmt.Sprintf("%s=%s", key, values.Get(key)))
	}
	dataCheckStr := strings.Join(dataCheckString, "\n")

	// Создаем секретный ключ
	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(botToken))
	secret := secretKey.Sum(nil)

	// Проверяем подпись
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(dataCheckStr))
	expectedHash := hex.EncodeToString(h.Sum(nil))

	return hash == expectedHash
}

func main() {
	initData := "query_id=AAHaO4xbAwAAANo7jFsTvCas&user=%7B%22id%22%3A7978367962%2C%22first_name%22%3A%22TWELVEFACED%F0%9F%98%AD%F0%9F%98%8F%F0%9F%8E%83%F0%9F%98%8A%F0%9F%98%AD%2B%5E.%22%2C%22last_name%22%3A%22%22%2C%22username%22%3A%22twelvefacedjanu%22%2C%22language_code%22%3A%22ru%22%2C%22allows_write_to_pm%22%3Atrue%2C%22photo_url%22%3A%22https%3A%5C%2F%5C%2Ft.me%5C%2Fi%5C%2Fuserpic%5C%2F320%5C%2FBW72zL9ncSqhPEWrFWXonpX7_QHbUBb3abm0yUTUOvJxw2rsEOcy5gouHJQob1Jj.svg%22%7D&auth_date=1768216174&signature=4xuN8ZpW_GX6fy0LsJzOpU6OZbQU7TeaE3Y6qHPzdE-vX3QqBGuM3jQDav2cfGliVgkqt1T9ZjkYD12IwmNaBQ&hash=3df1da2a9cc7516e6aa0808c20c900a8713556455564afa1d7a6e0910cf2c1e7"

	fmt.Println("Проверяем ваш init_data...")
	fmt.Println()

	// Текущий токен из .env
	currentToken := "8573303851:AAGPLBbgFD0BwFHIp4V0s5Ia3NYOxRb7uzs"
	
	fmt.Printf("Проверяем текущий BOT_TOKEN: %s...\n", currentToken[:10])
	if validateInitDataWithToken(initData, currentToken) {
		fmt.Println("✅ Ваш init_data ВАЛИДЕН с текущим BOT_TOKEN!")
		return
	}
	fmt.Println("❌ Ваш init_data НЕ валиден с текущим BOT_TOKEN")
	
	fmt.Println()
	fmt.Println("Ваш init_data был сгенерирован другим ботом.")
	fmt.Println("Вам нужно:")
	fmt.Println("1. Узнать правильный BOT_TOKEN от вашего бота")
	fmt.Println("2. Обновить .env файл с правильным токеном")
	fmt.Println("3. Или получить новый init_data от бота с токеном", currentToken[:10]+"...")
	
	// Парсим данные пользователя для отладки
	values, _ := url.ParseQuery(initData)
	fmt.Println()
	fmt.Println("Данные из вашего init_data:")
	fmt.Printf("- User ID: %s\n", extractUserID(values.Get("user")))
	fmt.Printf("- Auth Date: %s\n", values.Get("auth_date"))
	fmt.Printf("- Query ID: %s\n", values.Get("query_id"))
}

func extractUserID(userStr string) string {
	// Простое извлечение ID из JSON строки
	if strings.Contains(userStr, `"id":`) {
		start := strings.Index(userStr, `"id":`) + 5
		end := strings.Index(userStr[start:], ",")
		if end == -1 {
			end = strings.Index(userStr[start:], "}")
		}
		if end != -1 {
			return userStr[start : start+end]
		}
	}
	return "unknown"
}