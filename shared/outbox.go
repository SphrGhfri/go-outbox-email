package shared

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"gorm.io/gorm"
)

type OutBoxMessage struct {
	ID          string    `gorm:"id" json:"id"`
	EventName   string    `gorm:"event_name" json:"event_name"`
	UserID      uuid.UUID `gorm:"user_id" json:"user_id"`
	Name        string    `gorm:"name" json:"name"`
	Email       string    `gorm:"email" json:"email"`
	Message     string    `gorm:"message" json:"message"`
	IsProcessed bool      `gorm:"is_processed" json:"is_processed"`
}

type OutboxProcesser struct {
	DB        *gorm.DB
	JSContext nats.JetStreamContext
	Subject   string
}

func (p *OutboxProcesser) HandleOutboxMessage() {
	messages := make([]OutBoxMessage, 0)
	err := p.DB.
		Where("is_processed = ?", false).
		Find(&messages).Error
	if err != nil {
		log.Println("query outbox messages error: ", err)
		return
	}

	// no waiting message
	if len(messages) == 0 {
		return
	}

	// Publish each message.
	// If success, add to processed slice
	processedID := make([]string, 0)
	for _, m := range messages {
		b, err := json.Marshal(m)
		if err != nil {
			continue
		}

		// publish message to a queue
		if err := p.publishMessage(b); err != nil {
			log.Println("publish outbox message error: ", err)
			continue
		}

		processedID = append(processedID, m.ID)
	}

	// Update processed messages in database
	err = p.DB.Model(&OutBoxMessage{}).
		Where("id IN ?", processedID).
		UpdateColumn("is_processed", true).Error
	if err != nil {
		log.Println("update outbox error: ", err)
		return
	}

	log.Println("Published messages:", processedID)
}

func (p *OutboxProcesser) publishMessage(body []byte) error {
	_, err := p.JSContext.Publish(p.Subject, body)
	return err
}
