package handlers

import (
	"fmt"
	"log"
	"strings"

	"poker/database"
	"poker/models"
	"poker/services"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

// GetTables возвращает столы по категории
// @Summary Получить список столов
// @Description Возвращает список столов по указанной категории
// @Tags tables
// @Accept json
// @Produce json
// @Param category query string false "Категория столов" Enums(ALL,LOW,MID,VIP) default(ALL)
// @Success 200 {object} models.TableResponse
// @Failure 500 {object} map[string]string
// @Router /tables [get]
// @Router /public/tables [get]
func GetTables(c fiber.Ctx) error {
	category := c.Query("category", "ALL")
	category = strings.ToUpper(category)

	var tables []models.Table

	if category == "ALL" {
		// Получаем все столы
		if err := database.DB.Find(&tables).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Ошибка при получении столов",
			})
		}
	} else {
		// Фильтруем по категории
		if err := database.DB.Where("category = ?", category).Find(&tables).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Ошибка при получении столов",
			})
		}
	}

	response := models.TableResponse{
		Tables: tables,
	}

	return c.JSON(response)
}

// GetTableByID возвращает конкретный стол по ID
// @Summary Получить стол по ID
// @Description Возвращает информацию о конкретном столе
// @Tags tables
// @Accept json
// @Produce json
// @Param id path int true "ID стола"
// @Success 200 {object} models.Table
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /tables/{id} [get]
// @Router /public/tables/{id} [get]
func GetTableByID(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid table ID",
		})
	}

	// Конвертируем строку в int
	var tableID int
	if _, err := fmt.Sscanf(id, "%d", &tableID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid table ID format",
		})
	}

	var table models.Table
	if err := database.DB.First(&table, tableID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Table not found",
		})
	}

	return c.JSON(table)
}

// JoinTable позволяет игроку присоединиться к столу
// @Summary Присоединиться к столу
// @Description Позволяет авторизованному игроку присоединиться к указанному столу
// @Tags tables
// @Accept json
// @Produce json
// @Security TelegramAuth
// @Param id path int true "ID стола"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tables/{id}/join [post]
func JoinTable(c fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid table ID",
		})
	}

	// Конвертируем строку в int
	var tableID int
	if _, err := fmt.Sscanf(id, "%d", &tableID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid table ID format",
		})
	}

	// Начинаем транзакцию
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Получаем стол с блокировкой для обновления
	var table models.Table
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&table, tableID).Error; err != nil {
		tx.Rollback()
		return c.Status(404).JSON(fiber.Map{
			"error": "Table not found",
		})
	}

	// Проверяем, достаточно ли средств для buy-in
	if user.Balance < table.BuyIn {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{
			"error": "Insufficient balance for buy-in",
		})
	}

	// Проверяем, не сидит ли пользователь уже за этим столом
	var existingPlayer models.TablePlayer
	if err := tx.Where("table_id = ? AND user_uuid = ?", tableID, user.UUID).First(&existingPlayer).Error; err == nil {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{
			"error": "Already sitting at this table",
		})
	}

	// Проверяем, есть ли свободные места
	if table.Players >= table.MaxSeats {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{
			"error": "Table is full",
		})
	}

	// Проверяем, нужно ли создать новый стол той же категории
	var shouldCreateNewTable bool
	if table.Players >= 2 {
		shouldCreateNewTable = true
	}

	// Находим свободное место
	var occupiedSeats []int
	tx.Model(&models.TablePlayer{}).Where("table_id = ?", tableID).Pluck("seat_number", &occupiedSeats)
	
	seatNumber := 1
	for seatNumber <= table.MaxSeats {
		occupied := false
		for _, seat := range occupiedSeats {
			if seat == seatNumber {
				occupied = true
				break
			}
		}
		if !occupied {
			break
		}
		seatNumber++
	}

	// Создаем запись игрока за столом
	tablePlayer := models.TablePlayer{
		TableID:    tableID,
		UserUUID:   user.UUID,
		SeatNumber: seatNumber,
		Chips:      table.BuyIn,
	}

	if err := tx.Create(&tablePlayer).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to join table",
		})
	}

	// Списываем buy-in с баланса пользователя
	user.Balance -= table.BuyIn
	if err := tx.Save(user).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update balance",
		})
	}

	// Увеличиваем количество игроков
	table.Players++
	if err := tx.Save(&table).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update table",
		})
	}

	// Создаем новый стол той же категории, если нужно
	var newTable *models.Table
	if shouldCreateNewTable {
		newTable = &models.Table{
			Category: table.Category,
			Blinds:   table.Blinds,
			BuyIn:    table.BuyIn,
			Players:  0,
			MaxSeats: table.MaxSeats,
		}

		if err := tx.Create(newTable).Error; err != nil {
			// Не критичная ошибка, продолжаем
			log.Printf("Предупреждение: не удалось создать новый стол: %v", err)
		}
	}

	// Подтверждаем транзакцию
	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save changes",
		})
	}

	// Отправляем события в Kafka
	if services.Kafka != nil {
		// Событие присоединения к столу
		joinEvent := map[string]interface{}{
			"user":        user,
			"table":       table,
			"seat_number": seatNumber,
		}
		services.Kafka.PublishTableEvent(tableID, "player_joined", joinEvent)

		// Событие создания нового стола
		if newTable != nil {
			services.Kafka.PublishTableEvent(newTable.ID, "table_created", newTable)
		}
	}

	response := fiber.Map{
		"message": "Successfully joined table",
		"table":   table,
		"seat":    seatNumber,
		"chips":   table.BuyIn,
		"balance": user.Balance,
	}

	// Добавляем информацию о новом столе, если он был создан
	if newTable != nil {
		response["new_table_created"] = true
		response["new_table"] = newTable
	}

	return c.JSON(response)
}

// LeaveTable позволяет игроку покинуть стол
func LeaveTable(c fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid table ID",
		})
	}

	// Конвертируем строку в int
	var tableID int
	if _, err := fmt.Sscanf(id, "%d", &tableID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid table ID format",
		})
	}

	// Начинаем транзакцию
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Ищем игрока за столом
	var tablePlayer models.TablePlayer
	if err := tx.Where("table_id = ? AND user_uuid = ?", tableID, user.UUID).First(&tablePlayer).Error; err != nil {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{
			"error": "You are not sitting at this table",
		})
	}

	// Получаем стол с блокировкой для обновления
	var table models.Table
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&table, tableID).Error; err != nil {
		tx.Rollback()
		return c.Status(404).JSON(fiber.Map{
			"error": "Table not found",
		})
	}

	// Возвращаем фишки на баланс пользователя
	user.Balance += tablePlayer.Chips
	if err := tx.Save(user).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update balance",
		})
	}

	// Удаляем игрока из-за стола
	if err := tx.Delete(&tablePlayer).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to leave table",
		})
	}

	// Уменьшаем количество игроков
	table.Players--
	if err := tx.Save(&table).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update table",
		})
	}

	// Подтверждаем транзакцию
	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save changes",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Successfully left table",
		"table":   table,
		"chips_returned": tablePlayer.Chips,
		"balance": user.Balance,
	})
}

// GetTablePlayers возвращает список игроков за столом
func GetTablePlayers(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid table ID",
		})
	}

	var tableID int
	if _, err := fmt.Sscanf(id, "%d", &tableID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid table ID format",
		})
	}

	var players []models.TablePlayer
	if err := database.DB.Preload("User").Where("table_id = ?", tableID).Find(&players).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get players",
		})
	}

	return c.JSON(fiber.Map{
		"players": players,
	})
}

// GetMyTables возвращает столы, за которыми сидит текущий пользователь
func GetMyTables(c fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	var tablePlayers []models.TablePlayer
	if err := database.DB.Preload("Table").Where("user_uuid = ?", user.UUID).Find(&tablePlayers).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get your tables",
		})
	}

	return c.JSON(fiber.Map{
		"tables": tablePlayers,
	})
}
// JoinAvailableTable присоединяет игрока к доступному столу или создает новый
// @Summary Присоединиться к доступному столу
// @Description Автоматически находит доступный стол указанной категории или создает новый
// @Tags tables
// @Accept json
// @Produce json
// @Security TelegramAuth
// @Param request body map[string]string true "Категория стола"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /join-available-table [post]
func JoinAvailableTable(c fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	
	var requestData struct {
		Category string `json:"category"` // LOW, MID, VIP
	}

	if err := c.Bind().JSON(&requestData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	category := strings.ToUpper(requestData.Category)
	if category == "" {
		category = "LOW" // По умолчанию LOW
	}

	// Проверяем валидность категории
	if category != "LOW" && category != "MID" && category != "VIP" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid category. Must be LOW, MID, or VIP",
		})
	}

	// Начинаем транзакцию
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Ищем доступный стол в указанной категории
	var availableTable models.Table
	err := tx.Where("category = ? AND players < max_seats", category).
		Order("players DESC, id ASC"). // Предпочитаем столы с большим количеством игроков
		First(&availableTable).Error

	if err != nil {
		// Нет доступных столов, создаем новый
		newTable, createErr := createNewTable(tx, category)
		if createErr != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to create new table: " + createErr.Error(),
			})
		}
		availableTable = *newTable
	}

	// Проверяем, достаточно ли средств для buy-in
	if user.Balance < availableTable.BuyIn {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{
			"error": "Insufficient balance for buy-in",
		})
	}

	// Проверяем, не сидит ли пользователь уже за этим столом
	var existingPlayer models.TablePlayer
	if err := tx.Where("table_id = ? AND user_uuid = ?", availableTable.ID, user.UUID).First(&existingPlayer).Error; err == nil {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{
			"error": "Already sitting at this table",
		})
	}

	// Находим свободное место
	var occupiedSeats []int
	tx.Model(&models.TablePlayer{}).Where("table_id = ?", availableTable.ID).Pluck("seat_number", &occupiedSeats)
	
	seatNumber := 1
	for seatNumber <= availableTable.MaxSeats {
		occupied := false
		for _, seat := range occupiedSeats {
			if seat == seatNumber {
				occupied = true
				break
			}
		}
		if !occupied {
			break
		}
		seatNumber++
	}

	// Создаем запись игрока за столом
	tablePlayer := models.TablePlayer{
		TableID:    availableTable.ID,
		UserUUID:   user.UUID,
		SeatNumber: seatNumber,
		Chips:      availableTable.BuyIn,
	}

	if err := tx.Create(&tablePlayer).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to join table",
		})
	}

	// Списываем buy-in с баланса пользователя
	user.Balance -= availableTable.BuyIn
	if err := tx.Save(user).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update balance",
		})
	}

	// Увеличиваем количество игроков
	availableTable.Players++
	if err := tx.Save(&availableTable).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update table",
		})
	}

	// Проверяем, нужно ли создать еще один стол той же категории
	var newTable *models.Table
	if availableTable.Players >= 2 {
		newTable, err = createNewTable(tx, category)
		if err != nil {
			log.Printf("Предупреждение: не удалось создать дополнительный стол: %v", err)
		}
	}

	// Подтверждаем транзакцию
	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save changes",
		})
	}

	// Отправляем события в Kafka
	if services.Kafka != nil {
		joinEvent := map[string]interface{}{
			"user":        user,
			"table":       availableTable,
			"seat_number": seatNumber,
		}
		services.Kafka.PublishTableEvent(availableTable.ID, "player_joined", joinEvent)

		if newTable != nil {
			services.Kafka.PublishTableEvent(newTable.ID, "table_created", newTable)
		}
	}

	response := fiber.Map{
		"message": "Successfully joined table",
		"table":   availableTable,
		"seat":    seatNumber,
		"chips":   availableTable.BuyIn,
		"balance": user.Balance,
	}

	if newTable != nil {
		response["additional_table_created"] = true
		response["additional_table"] = newTable
	}

	return c.JSON(response)
}

// createNewTable создает новый стол указанной категории
func createNewTable(tx *gorm.DB, category string) (*models.Table, error) {
	var blinds string
	var buyIn int
	var maxSeats int

	switch category {
	case "LOW":
		blinds = "1/2"
		buyIn = 50
		maxSeats = 6
	case "MID":
		blinds = "5/10"
		buyIn = 200
		maxSeats = 9
	case "VIP":
		blinds = "25/50"
		buyIn = 1000
		maxSeats = 6
	default:
		return nil, fmt.Errorf("invalid category: %s", category)
	}

	newTable := &models.Table{
		Category: category,
		Blinds:   blinds,
		BuyIn:    buyIn,
		Players:  0,
		MaxSeats: maxSeats,
	}

	if err := tx.Create(newTable).Error; err != nil {
		return nil, err
	}

	return newTable, nil
}

// GetAvailableTables возвращает доступные столы по категориям
func GetAvailableTables(c fiber.Ctx) error {
	category := c.Query("category", "ALL")
	category = strings.ToUpper(category)

	var tables []models.Table
	query := database.DB.Where("players < max_seats")

	if category != "ALL" {
		query = query.Where("category = ?", category)
	}

	if err := query.Order("category ASC, players DESC").Find(&tables).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get available tables",
		})
	}

	// Группируем столы по категориям
	tablesByCategory := make(map[string][]models.Table)
	for _, table := range tables {
		tablesByCategory[table.Category] = append(tablesByCategory[table.Category], table)
	}

	return c.JSON(fiber.Map{
		"available_tables": tablesByCategory,
		"total_tables":     len(tables),
	})
}
// GetTableStatistics возвращает статистику по столам
func GetTableStatistics(c fiber.Ctx) error {
	if services.Manager == nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Table manager not initialized",
		})
	}

	stats := services.Manager.GetTableStatistics()
	
	return c.JSON(fiber.Map{
		"statistics": stats,
	})
}

// CleanupEmptyTables принудительно очищает пустые столы
func CleanupEmptyTables(c fiber.Ctx) error {
	if services.Manager == nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Table manager not initialized",
		})
	}

	services.Manager.CleanupEmptyTables()
	
	return c.JSON(fiber.Map{
		"message": "Empty tables cleanup completed",
	})
}