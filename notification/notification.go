package notification

import (
	"encoding/json"
	"outbox/shared"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Notification struct {
	ID        string `json:"id" gorm:"id,primarykey"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Handler struct {
	DB *gorm.DB
}

func (h *Handler) Add(c *fiber.Ctx) error {

	var notification Notification
	if err := c.BodyParser(&notification); err != nil {
		return err
	}
	notification.ID = uuid.NewString()
	notification.CreatedAt = time.Now()

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		b, err := json.Marshal(notification)
		if err != nil {
			return err
		}

		notificationCreatedEvent := shared.OutBoxMessage{
			ID:          uuid.NewString(),
			EventName:   "NotificationCreated",
			Payload:     datatypes.JSON(b),
			IsProcessed: false,
		}

		if err := tx.FirstOrCreate(&notification).Error; err != nil {
			return err
		}

		if err := tx.Create(&notificationCreatedEvent).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
