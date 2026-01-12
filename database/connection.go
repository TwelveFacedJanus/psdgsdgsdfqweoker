package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// ConnectDB подключается к PostgreSQL базе данных
func ConnectDB() {
	// Получаем параметры подключения из переменных окружения или используем значения по умолчанию
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "poker_user")
	password := getEnv("DB_PASSWORD", "poker_password")
	dbname := getEnv("DB_NAME", "poker_db")
	sslmode := getEnv("DB_SSLMODE", "disable")

	// Формируем строку подключения
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	// Подключаемся к базе данных
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Не удалось подключиться к базе данных:", err)
	}

	log.Println("Успешно подключились к PostgreSQL базе данных")
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}