package services

import (
	"fmt"
	"log"
	"time"

	"poker/database"
	"poker/models"
)

type TableManager struct {
	ticker *time.Ticker
	done   chan bool
}

var Manager *TableManager

// InitTableManager инициализирует менеджер столов
func InitTableManager() {
	Manager = &TableManager{
		ticker: time.NewTicker(30 * time.Second), // Проверяем каждые 30 секунд
		done:   make(chan bool),
	}

	go Manager.run()
	log.Println("Менеджер столов запущен")
}

// run основной цикл менеджера столов
func (tm *TableManager) run() {
	for {
		select {
		case <-tm.done:
			return
		case <-tm.ticker.C:
			tm.checkAndCreateTables()
		}
	}
}

// checkAndCreateTables проверяет и создает новые столы при необходимости
func (tm *TableManager) checkAndCreateTables() {
	categories := []string{"LOW", "MID", "VIP"}

	for _, category := range categories {
		// Проверяем, есть ли доступные столы в категории
		var availableCount int64
		database.DB.Model(&models.Table{}).
			Where("category = ? AND players < max_seats", category).
			Count(&availableCount)

		// Если нет доступных столов, создаем новый
		if availableCount == 0 {
			if err := tm.createTableForCategory(category); err != nil {
				log.Printf("Ошибка создания стола для категории %s: %v", category, err)
			} else {
				log.Printf("Создан новый стол для категории %s", category)
			}
		}

		// Проверяем, нужны ли дополнительные столы
		// Если все доступные столы заполнены более чем на 50%, создаем еще один
		var tables []models.Table
		database.DB.Where("category = ? AND players < max_seats", category).Find(&tables)

		needNewTable := true
		for _, table := range tables {
			occupancyRate := float64(table.Players) / float64(table.MaxSeats)
			if occupancyRate < 0.5 { // Если есть стол заполненный менее чем на 50%
				needNewTable = false
				break
			}
		}

		if needNewTable && len(tables) > 0 {
			if err := tm.createTableForCategory(category); err != nil {
				log.Printf("Ошибка создания дополнительного стола для категории %s: %v", category, err)
			} else {
				log.Printf("Создан дополнительный стол для категории %s", category)
			}
		}
	}
}

// createTableForCategory создает новый стол для указанной категории
func (tm *TableManager) createTableForCategory(category string) error {
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
		return fmt.Errorf("invalid category: %s", category)
	}

	newTable := models.Table{
		Category: category,
		Blinds:   blinds,
		BuyIn:    buyIn,
		Players:  0,
		MaxSeats: maxSeats,
	}

	if err := database.DB.Create(&newTable).Error; err != nil {
		return err
	}

	// Отправляем событие в Kafka
	if Kafka != nil {
		Kafka.PublishTableEvent(newTable.ID, "table_auto_created", newTable)
	}

	return nil
}

// CleanupEmptyTables удаляет пустые столы (кроме одного в каждой категории)
func (tm *TableManager) CleanupEmptyTables() {
	categories := []string{"LOW", "MID", "VIP"}

	for _, category := range categories {
		var emptyTables []models.Table
		database.DB.Where("category = ? AND players = 0", category).
			Order("created_at DESC").
			Find(&emptyTables)

		// Оставляем один пустой стол в каждой категории
		if len(emptyTables) > 1 {
			tablesToDelete := emptyTables[1:] // Удаляем все кроме первого (самого нового)
			
			for _, table := range tablesToDelete {
				// Проверяем, что за столом действительно никого нет
				var playerCount int64
				database.DB.Model(&models.TablePlayer{}).
					Where("table_id = ?", table.ID).
					Count(&playerCount)

				if playerCount == 0 {
					database.DB.Delete(&table)
					log.Printf("Удален пустой стол ID: %d, категория: %s", table.ID, table.Category)
					
					// Отправляем событие в Kafka
					if Kafka != nil {
						Kafka.PublishTableEvent(table.ID, "table_removed", table)
					}
				}
			}
		}
	}
}

// GetTableStatistics возвращает статистику по столам
func (tm *TableManager) GetTableStatistics() map[string]interface{} {
	stats := make(map[string]interface{})
	categories := []string{"LOW", "MID", "VIP"}

	for _, category := range categories {
		var totalTables int64
		var availableTables int64
		var totalPlayers int64

		database.DB.Model(&models.Table{}).
			Where("category = ?", category).
			Count(&totalTables)

		database.DB.Model(&models.Table{}).
			Where("category = ? AND players < max_seats", category).
			Count(&availableTables)

		database.DB.Model(&models.Table{}).
			Where("category = ?", category).
			Select("COALESCE(SUM(players), 0)").
			Scan(&totalPlayers)

		stats[category] = map[string]interface{}{
			"total_tables":     totalTables,
			"available_tables": availableTables,
			"total_players":    totalPlayers,
		}
	}

	return stats
}

// Stop останавливает менеджер столов
func (tm *TableManager) Stop() {
	tm.ticker.Stop()
	tm.done <- true
	log.Println("Менеджер столов остановлен")
}