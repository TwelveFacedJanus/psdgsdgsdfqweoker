package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"poker/models"

	"github.com/IBM/sarama"
)

type KafkaService struct {
	producer sarama.SyncProducer
	consumer sarama.Consumer
}

var Kafka *KafkaService

// InitKafka инициализирует подключение к Kafka
func InitKafka() error {
	brokers := strings.Split(getEnv("KAFKA_BROKERS", "kafka:9092"), ",")
	
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3
	config.Consumer.Return.Errors = true

	// Создаем продюсера
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return err
	}

	// Создаем консьюмера
	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		producer.Close()
		return err
	}

	Kafka = &KafkaService{
		producer: producer,
		consumer: consumer,
	}

	log.Println("Успешно подключились к Kafka")
	return nil
}

// PublishGameEvent публикует игровое событие
func (k *KafkaService) PublishGameEvent(event models.GameEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	topic := "poker-game-events"
	message := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(event.GameID),
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = k.producer.SendMessage(message)
	if err != nil {
		log.Printf("Ошибка отправки сообщения в Kafka: %v", err)
		return err
	}

	log.Printf("Отправлено событие: %s для игры %s", event.Type, event.GameID)
	return nil
}

// PublishTableEvent публикует событие стола
func (k *KafkaService) PublishTableEvent(tableID int, eventType string, data interface{}) error {
	event := models.GameEvent{
		Type:      eventType,
		TableID:   tableID,
		Data:      data,
		Timestamp: time.Now(),
	}

	eventData, err := json.Marshal(event)
	if err != nil {
		return err
	}

	topic := "poker-table-events"
	message := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(fmt.Sprintf("table-%d", tableID)),
		Value: sarama.ByteEncoder(eventData),
	}

	_, _, err = k.producer.SendMessage(message)
	if err != nil {
		log.Printf("Ошибка отправки события стола в Kafka: %v", err)
		return err
	}

	log.Printf("Отправлено событие стола: %s для стола %d", eventType, tableID)
	return nil
}

// ConsumeGameEvents потребляет игровые события
func (k *KafkaService) ConsumeGameEvents(handler func(models.GameEvent)) error {
	topic := "poker-game-events"
	partitionConsumer, err := k.consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		return err
	}

	go func() {
		defer partitionConsumer.Close()
		
		for {
			select {
			case message := <-partitionConsumer.Messages():
				var event models.GameEvent
				if err := json.Unmarshal(message.Value, &event); err != nil {
					log.Printf("Ошибка парсинга события: %v", err)
					continue
				}
				
				handler(event)
				
			case err := <-partitionConsumer.Errors():
				log.Printf("Ошибка консьюмера: %v", err)
			}
		}
	}()

	return nil
}

// Close закрывает соединения с Kafka
func (k *KafkaService) Close() error {
	if k.producer != nil {
		k.producer.Close()
	}
	if k.consumer != nil {
		k.consumer.Close()
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}