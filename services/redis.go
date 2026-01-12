package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"poker/models"

	"github.com/redis/go-redis/v9"
)

type RedisService struct {
	client *redis.Client
	ctx    context.Context
}

var Redis *RedisService

// InitRedis инициализирует подключение к Redis
func InitRedis() error {
	addr := getEnv("REDIS_ADDR", "redis:6379")
	password := getEnv("REDIS_PASSWORD", "")
	db := 0

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()
	
	// Проверяем соединение
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("не удалось подключиться к Redis: %v", err)
	}

	Redis = &RedisService{
		client: client,
		ctx:    ctx,
	}

	log.Println("Успешно подключились к Redis")
	return nil
}

// SetGameState сохраняет состояние игры в Redis
func (r *RedisService) SetGameState(gameID string, game *models.Game) error {
	data, err := json.Marshal(game)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("game:%s", gameID)
	return r.client.Set(r.ctx, key, data, time.Hour).Err()
}

// GetGameState получает состояние игры из Redis
func (r *RedisService) GetGameState(gameID string) (*models.Game, error) {
	key := fmt.Sprintf("game:%s", gameID)
	data, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("игра не найдена")
		}
		return nil, err
	}

	var game models.Game
	if err := json.Unmarshal([]byte(data), &game); err != nil {
		return nil, err
	}

	return &game, nil
}

// SetPlayerSession сохраняет сессию игрока
func (r *RedisService) SetPlayerSession(userUUID string, tableID int) error {
	key := fmt.Sprintf("player_session:%s", userUUID)
	return r.client.Set(r.ctx, key, tableID, time.Hour*24).Err()
}

// GetPlayerSession получает сессию игрока
func (r *RedisService) GetPlayerSession(userUUID string) (int, error) {
	key := fmt.Sprintf("player_session:%s", userUUID)
	result, err := r.client.Get(r.ctx, key).Int()
	if err != nil {
		if err == redis.Nil {
			return 0, fmt.Errorf("сессия не найдена")
		}
		return 0, err
	}
	return result, nil
}

// AddPlayerToTable добавляет игрока к столу в Redis
func (r *RedisService) AddPlayerToTable(tableID int, userUUID string) error {
	key := fmt.Sprintf("table_players:%d", tableID)
	return r.client.SAdd(r.ctx, key, userUUID).Err()
}

// RemovePlayerFromTable удаляет игрока из стола в Redis
func (r *RedisService) RemovePlayerFromTable(tableID int, userUUID string) error {
	key := fmt.Sprintf("table_players:%d", tableID)
	return r.client.SRem(r.ctx, key, userUUID).Err()
}

// GetTablePlayers получает список игроков за столом
func (r *RedisService) GetTablePlayers(tableID int) ([]string, error) {
	key := fmt.Sprintf("table_players:%d", tableID)
	return r.client.SMembers(r.ctx, key).Result()
}

// SetTableLock устанавливает блокировку стола
func (r *RedisService) SetTableLock(tableID int, duration time.Duration) error {
	key := fmt.Sprintf("table_lock:%d", tableID)
	return r.client.Set(r.ctx, key, "locked", duration).Err()
}

// IsTableLocked проверяет, заблокирован ли стол
func (r *RedisService) IsTableLocked(tableID int) bool {
	key := fmt.Sprintf("table_lock:%d", tableID)
	_, err := r.client.Get(r.ctx, key).Result()
	return err != redis.Nil
}

// PublishToChannel публикует сообщение в канал Redis
func (r *RedisService) PublishToChannel(channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	
	return r.client.Publish(r.ctx, channel, data).Err()
}

// SubscribeToChannel подписывается на канал Redis
func (r *RedisService) SubscribeToChannel(channel string, handler func(string)) error {
	pubsub := r.client.Subscribe(r.ctx, channel)
	defer pubsub.Close()

	ch := pubsub.Channel()
	
	go func() {
		for msg := range ch {
			handler(msg.Payload)
		}
	}()

	return nil
}

// Close закрывает соединение с Redis
func (r *RedisService) Close() error {
	return r.client.Close()
}