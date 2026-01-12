package handlers

import (
	"strconv"
	"time"

	"poker/database"
	"poker/game"
	"poker/models"
	"poker/services"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// StartGame начинает новую игру за столом
// @Summary Начать игру за столом
// @Description Начинает новую игру в покер за указанным столом
// @Tags game
// @Accept json
// @Produce json
// @Security TelegramAuth
// @Param id path int true "ID стола"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tables/{id}/start-game [post]
func StartGame(c fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	
	tableIDStr := c.Params("id")
	tableID, err := strconv.Atoi(tableIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid table ID",
		})
	}

	// Проверяем, что пользователь сидит за столом
	var tablePlayer models.TablePlayer
	if err := database.DB.Where("table_id = ? AND user_uuid = ?", tableID, user.UUID).First(&tablePlayer).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "You are not sitting at this table",
		})
	}

	// Проверяем, нет ли уже активной игры
	var existingGame models.Game
	if err := database.DB.Where("table_id = ? AND state NOT IN (?)", tableID, []string{"finished"}).First(&existingGame).Error; err == nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Game already in progress",
		})
	}

	// Получаем всех игроков за столом
	var players []models.TablePlayer
	if err := database.DB.Preload("User").Where("table_id = ?", tableID).Find(&players).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get players",
		})
	}

	if len(players) < 2 {
		return c.Status(400).JSON(fiber.Map{
			"error": "Need at least 2 players to start game",
		})
	}

	// Получаем информацию о столе
	var table models.Table
	if err := database.DB.First(&table, tableID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Table not found",
		})
	}

	// Создаем новую игру
	newGame := models.Game{
		ID:             uuid.New().String(),
		TableID:        tableID,
		State:          models.GameStateWaiting,
		Deck:           game.CreateDeck(),
		CommunityCards: []models.Card{},
		Pot:            0,
		CurrentBet:     0,
		DealerPosition: 0,
		CurrentPlayer:  0,
		SmallBlind:     table.BuyIn / 100, // 1% от buy-in
		BigBlind:       table.BuyIn / 50,  // 2% от buy-in
	}

	// Создаем игроков в игре
	var gamePlayers []models.GamePlayer
	for i, player := range players {
		gamePlayer := models.GamePlayer{
			GameID:     newGame.ID,
			UserUUID:   player.UserUUID,
			Position:   i,
			Cards:      []models.Card{},
			Chips:      player.Chips,
			Bet:        0,
			IsFolded:   false,
			IsAllIn:    false,
			LastAction: "",
		}
		gamePlayers = append(gamePlayers, gamePlayer)
	}

	newGame.Players = gamePlayers

	// Сохраняем игру в базу данных
	if err := database.DB.Create(&newGame).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create game",
		})
	}

	// Сохраняем состояние игры в Redis
	if services.Redis != nil {
		services.Redis.SetGameState(newGame.ID, &newGame)
	}

	// Отправляем событие в Kafka
	if services.Kafka != nil {
		event := models.GameEvent{
			Type:      "game_started",
			GameID:    newGame.ID,
			TableID:   tableID,
			Data:      newGame,
			Timestamp: time.Now(),
		}
		services.Kafka.PublishGameEvent(event)
	}

	return c.JSON(fiber.Map{
		"message": "Game started successfully",
		"game":    newGame,
	})
}

// GetGameState возвращает текущее состояние игры
// @Summary Получить состояние игры
// @Description Возвращает текущее состояние игры по ID
// @Tags game
// @Accept json
// @Produce json
// @Security TelegramAuth
// @Param gameId path string true "ID игры"
// @Success 200 {object} models.Game
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /games/{gameId} [get]
func GetGameState(c fiber.Ctx) error {
	gameID := c.Params("gameId")
	if gameID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid game ID",
		})
	}

	// Сначала пробуем получить из Redis
	if services.Redis != nil {
		if gameState, err := services.Redis.GetGameState(gameID); err == nil {
			return c.JSON(gameState)
		}
	}

	// Если не найдено в Redis, получаем из базы данных
	var game models.Game
	if err := database.DB.Preload("Players.User").Preload("Table").First(&game, "id = ?", gameID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Game not found",
		})
	}

	return c.JSON(game)
}

// PlayerAction обрабатывает действие игрока
// @Summary Сделать ход в игре
// @Description Обрабатывает игровое действие игрока (fold, call, raise, check, bet)
// @Tags game
// @Accept json
// @Produce json
// @Security TelegramAuth
// @Param gameId path string true "ID игры"
// @Param action body map[string]interface{} true "Действие игрока"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /games/{gameId}/action [post]
func PlayerAction(c fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	gameID := c.Params("gameId")

	var actionData struct {
		Action string `json:"action"`
		Amount int    `json:"amount"`
	}

	if err := c.Bind().JSON(&actionData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Получаем состояние игры
	var gameState models.Game
	if services.Redis != nil {
		if gs, err := services.Redis.GetGameState(gameID); err == nil {
			gameState = *gs
		}
	}

	if gameState.ID == "" {
		if err := database.DB.Preload("Players.User").First(&gameState, "id = ?", gameID).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "Game not found",
			})
		}
	}

	// Создаем движок игры
	engine := game.NewPokerEngine(&gameState)

	// Проверяем, может ли игрок действовать
	if !engine.CanPlayerAct(user.UUID) {
		return c.Status(400).JSON(fiber.Map{
			"error": "It's not your turn or you cannot act",
		})
	}

	// Обрабатываем действие
	action := models.PlayerAction(actionData.Action)
	if err := engine.ProcessAction(user.UUID, action, actionData.Amount); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Сохраняем действие в базу данных
	gameAction := models.GameAction{
		GameID:   gameID,
		UserUUID: user.UUID,
		Action:   action,
		Amount:   actionData.Amount,
	}
	database.DB.Create(&gameAction)

	// Проверяем, завершен ли раунд
	if engine.IsRoundComplete() {
		engine.AdvanceGameState()
	}

	// Обновляем состояние игры
	if err := database.DB.Save(&gameState).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save game state",
		})
	}

	// Обновляем Redis
	if services.Redis != nil {
		services.Redis.SetGameState(gameID, &gameState)
	}

	// Отправляем событие в Kafka
	if services.Kafka != nil {
		playerActionEvent := models.PlayerActionEvent{
			Action: action,
			Amount: actionData.Amount,
		}

		// Находим игрока
		for _, player := range gameState.Players {
			if player.UserUUID == user.UUID {
				playerActionEvent.Player = player
				break
			}
		}

		event := models.GameEvent{
			Type:      "player_action",
			GameID:    gameID,
			TableID:   gameState.TableID,
			UserUUID:  user.UUID,
			Data:      playerActionEvent,
			Timestamp: time.Now(),
		}
		services.Kafka.PublishGameEvent(event)

		// Если состояние игры изменилось
		if engine.IsRoundComplete() {
			stateEvent := models.GameStateEvent{
				State:          gameState.State,
				CommunityCards: gameState.CommunityCards,
				Pot:            gameState.Pot,
				CurrentBet:     gameState.CurrentBet,
				CurrentPlayer:  gameState.CurrentPlayer,
			}

			stateEventMsg := models.GameEvent{
				Type:      "game_state_changed",
				GameID:    gameID,
				TableID:   gameState.TableID,
				Data:      stateEvent,
				Timestamp: time.Now(),
			}
			services.Kafka.PublishGameEvent(stateEventMsg)
		}
	}

	return c.JSON(fiber.Map{
		"message": "Action processed successfully",
		"game":    gameState,
	})
}

// GetGameHistory возвращает историю действий игры
func GetGameHistory(c fiber.Ctx) error {
	gameID := c.Params("gameId")
	if gameID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid game ID",
		})
	}

	var actions []models.GameAction
	if err := database.DB.Preload("User").Where("game_id = ?", gameID).Order("created_at ASC").Find(&actions).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get game history",
		})
	}

	return c.JSON(fiber.Map{
		"actions": actions,
	})
}

// GetActiveGames возвращает активные игры пользователя
func GetActiveGames(c fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	var games []models.Game
	if err := database.DB.
		Joins("JOIN game_players ON games.id = game_players.game_id").
		Where("game_players.user_uuid = ? AND games.state NOT IN (?)", user.UUID, []string{"finished"}).
		Preload("Players.User").
		Preload("Table").
		Find(&games).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get active games",
		})
	}

	return c.JSON(fiber.Map{
		"games": games,
	})
}