package main

import (
	"log"
	"poker/database"
	"poker/handlers"
	"poker/middleware"
	"poker/services"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/joho/godotenv"

	_ "poker/docs" // Импорт сгенерированных Swagger документов
)

// @title Poker REST API
// @version 1.0
// @description Полнофункциональный REST API для игры в покер с категориями столов LOW, MID, VIP
// @description Использует PostgreSQL, Kafka, Redis и авторизацию через Telegram Web App
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host 208.123.185.204:3000
// @BasePath /api/v1

// @securityDefinitions.apikey TelegramAuth
// @in header
// @name x-init-data
// @description Telegram Web App init_data для авторизации

func main() {
	// Загружаем переменные окружения из .env файла
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используем переменные окружения системы")
	}

	// Подключаемся к базе данных
	database.ConnectDB()

	// Инициализируем Redis
	if err := services.InitRedis(); err != nil {
		log.Printf("Предупреждение: не удалось подключиться к Redis: %v", err)
	}

	// Инициализируем Kafka
	if err := services.InitKafka(); err != nil {
		log.Printf("Предупреждение: не удалось подключиться к Kafka: %v", err)
	}

	// Инициализируем менеджер столов
	services.InitTableManager()

	app := fiber.New()

	// CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,x-init-data",
		AllowCredentials: false,
	}))

	// Обработка preflight OPTIONS запросов
	app.Options("/*", func(c fiber.Ctx) error {
		return c.SendStatus(204)
	})

	// Swagger документация
	app.Get("/swagger/", func(c fiber.Ctx) error {
		return c.SendFile("./docs/index.html")
	})
	app.Get("/swagger/swagger.json", func(c fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.SendFile("./docs/swagger.json")
	})
	app.Get("/swagger/swagger.yaml", func(c fiber.Ctx) error {
		c.Set("Content-Type", "application/x-yaml")
		return c.SendFile("./docs/swagger.yaml")
	})

	// Базовые маршруты
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Poker API Server - Swagger docs: /swagger/")
	})
	app.Get("/healthcheck", healthcheck)

	// API маршруты
	api := app.Group("/api/v1")
	
	// Публичные маршруты (без авторизации)
	public := api.Group("/public")
	public.Get("/tables", handlers.GetTables)
	public.Get("/tables/:id", handlers.GetTableByID)
	public.Get("/tables/:id/players", handlers.GetTablePlayers)
	
	// Защищенные маршруты (требуют авторизации)
	protected := api.Group("/", middleware.AuthMiddleware())
	
	// Пользователи
	protected.Get("/profile", handlers.GetProfile)
	protected.Put("/profile", handlers.UpdateProfile)
	protected.Get("/my-tables", handlers.GetMyTables)
	
	// Столы (действия требуют авторизации)
	protected.Post("/tables/:id/join", handlers.JoinTable)
	protected.Post("/tables/:id/leave", handlers.LeaveTable)
	protected.Post("/join-available-table", handlers.JoinAvailableTable)
	protected.Get("/available-tables", handlers.GetAvailableTables)
	protected.Get("/table-statistics", handlers.GetTableStatistics)
	protected.Post("/cleanup-empty-tables", handlers.CleanupEmptyTables)

	// Игровые маршруты
	protected.Post("/tables/:id/start-game", handlers.StartGame)
	protected.Get("/games/:gameId", handlers.GetGameState)
	protected.Post("/games/:gameId/action", handlers.PlayerAction)
	protected.Get("/games/:gameId/history", handlers.GetGameHistory)
	protected.Get("/my-games", handlers.GetActiveGames)

	// Маршруты с опциональной авторизацией
	optional := api.Group("/", middleware.OptionalAuthMiddleware())
	optional.Get("/tables", handlers.GetTables)
	optional.Get("/tables/:id", handlers.GetTableByID)

	log.Println("Сервер запускается на порту 3000...")
	log.Println("Swagger документация доступна по адресу: http://localhost:3000/swagger/")
	app.Listen(":3000")
}

// healthcheck проверяет состояние сервера
// @Summary Проверка состояния сервера
// @Description Возвращает статус работы сервера
// @Tags system
// @Accept json
// @Produce plain
// @Success 200 {string} string "OK"
// @Router /healthcheck [get]
func healthcheck(c fiber.Ctx) error {
	return c.SendString("OK")
}
