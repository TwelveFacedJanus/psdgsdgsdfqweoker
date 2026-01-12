package handlers

import (
	"poker/database"
	"poker/models"

	"github.com/gofiber/fiber/v3"
)

// GetProfile возвращает профиль текущего пользователя
// @Summary Получить профиль пользователя
// @Description Возвращает профиль авторизованного пользователя
// @Tags users
// @Accept json
// @Produce json
// @Security TelegramAuth
// @Success 200 {object} map[string]models.User
// @Failure 401 {object} map[string]string
// @Router /profile [get]
func GetProfile(c fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	
	return c.JSON(fiber.Map{
		"user": user,
	})
}

// UpdateProfile обновляет профиль пользователя
// @Summary Обновить профиль пользователя
// @Description Обновляет данные профиля авторизованного пользователя
// @Tags users
// @Accept json
// @Produce json
// @Security TelegramAuth
// @Param profile body map[string]string true "Данные для обновления"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /profile [put]
func UpdateProfile(c fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	var updateData struct {
		Username string `json:"username"`
	}

	if err := c.Bind().JSON(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if updateData.Username != "" {
		user.Username = updateData.Username
	}

	// Сохраняем изменения в базе данных
	if err := database.DB.Save(user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update profile",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Profile updated successfully",
		"user":    user,
	})
}