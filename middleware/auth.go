package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"poker/database"
	"poker/models"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// TelegramInitData структура для парсинга init_data
type TelegramInitData struct {
	QueryID      string `json:"query_id"`
	User         TelegramUser `json:"user"`
	AuthDate     int64  `json:"auth_date"`
	Hash         string `json:"hash"`
}

// TelegramUser структура пользователя Telegram
type TelegramUser struct {
	ID           int64  `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
	IsPremium    bool   `json:"is_premium,omitempty"`
}

// AuthMiddleware проверяет авторизацию по x-init-data
func AuthMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		initData := c.Get("x-init-data")
		if initData == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Missing x-init-data header",
			})
		}

		// Парсим init_data без проверки хеша (для тестирования)
		telegramData, err := parseInitData(initData)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid init_data: " + err.Error(),
			})
		}

		// Проверяем срок действия (не старше 24 часов)
		if time.Now().Unix()-telegramData.AuthDate > 86400 {
			return c.Status(401).JSON(fiber.Map{
				"error": "Init data expired",
			})
		}

		// Получаем или создаем пользователя
		user, err := getOrCreateUser(telegramData.User)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to get user: " + err.Error(),
			})
		}

		// Сохраняем пользователя в контексте
		c.Locals("user", user)
		c.Locals("telegram_data", telegramData)

		return c.Next()
	}
}

// parseInitData парсит init_data без проверки хеша (для тестирования)
func parseInitData(initData string) (*TelegramInitData, error) {
	// Парсим URL-encoded данные
	values, err := url.ParseQuery(initData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse init_data: %v", err)
	}

	// Парсим данные пользователя
	var telegramData TelegramInitData
	
	if userStr := values.Get("user"); userStr != "" {
		if err := json.Unmarshal([]byte(userStr), &telegramData.User); err != nil {
			return nil, fmt.Errorf("failed to parse user data: %v", err)
		}
	}

	telegramData.QueryID = values.Get("query_id")
	telegramData.Hash = values.Get("hash")

	if authDateStr := values.Get("auth_date"); authDateStr != "" {
		if authDate, err := strconv.ParseInt(authDateStr, 10, 64); err == nil {
			telegramData.AuthDate = authDate
		}
	}

	return &telegramData, nil
}

// validateInitData валидирует init_data от Telegram
func validateInitData(initData string) (*TelegramInitData, error) {
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		return nil, fmt.Errorf("BOT_TOKEN not set")
	}

	// Парсим URL-encoded данные
	values, err := url.ParseQuery(initData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse init_data: %v", err)
	}

	// Извлекаем hash
	hash := values.Get("hash")
	if hash == "" {
		return nil, fmt.Errorf("missing hash")
	}

	// Удаляем hash и signature из параметров для проверки
	values.Del("hash")
	values.Del("signature")

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

	if hash != expectedHash {
		return nil, fmt.Errorf("invalid hash")
	}

	// Парсим данные пользователя
	var telegramData TelegramInitData
	
	if userStr := values.Get("user"); userStr != "" {
		if err := json.Unmarshal([]byte(userStr), &telegramData.User); err != nil {
			return nil, fmt.Errorf("failed to parse user data: %v", err)
		}
	}

	telegramData.QueryID = values.Get("query_id")
	telegramData.Hash = hash

	if authDateStr := values.Get("auth_date"); authDateStr != "" {
		if authDate, err := strconv.ParseInt(authDateStr, 10, 64); err == nil {
			telegramData.AuthDate = authDate
		}
	}

	return &telegramData, nil
}

// getOrCreateUser получает существующего пользователя или создает нового
func getOrCreateUser(telegramUser TelegramUser) (*models.User, error) {
	var user models.User

	// Ищем пользователя по telegram_id
	err := database.DB.Where("telegram_id = ?", telegramUser.ID).First(&user).Error
	if err == nil {
		// Пользователь найден, обновляем данные если нужно
		updated := false
		
		username := telegramUser.Username
		if username == "" {
			username = telegramUser.FirstName
			if telegramUser.LastName != "" {
				username += " " + telegramUser.LastName
			}
		}

		if user.Username != username {
			user.Username = username
			updated = true
		}

		if updated {
			database.DB.Save(&user)
		}

		return &user, nil
	}

	// Пользователь не найден, создаем нового
	username := telegramUser.Username
	if username == "" {
		username = telegramUser.FirstName
		if telegramUser.LastName != "" {
			username += " " + telegramUser.LastName
		}
	}

	user = models.User{
		UUID:       uuid.New().String(),
		Username:   username,
		TelegramID: telegramUser.ID,
		Balance:    1000, // Начальный баланс
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	return &user, nil
}

// OptionalAuthMiddleware - опциональная авторизация (не блокирует запрос если нет токена)
func OptionalAuthMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		initData := c.Get("x-init-data")
		if initData == "" {
			return c.Next()
		}

		telegramData, err := parseInitData(initData)
		if err != nil {
			return c.Next()
		}

		if time.Now().Unix()-telegramData.AuthDate > 86400 {
			return c.Next()
		}

		user, err := getOrCreateUser(telegramData.User)
		if err != nil {
			return c.Next()
		}

		c.Locals("user", user)
		c.Locals("telegram_data", telegramData)

		return c.Next()
	}
}