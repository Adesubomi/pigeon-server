package push

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeliveryLog struct {
	ID          string     `gorm:"type:uuid;primaryKey" json:"id"`
	EventID     string     `gorm:"type:uuid;index;not null" json:"event_id"`
	DeviceID    string     `gorm:"type:uuid;index;not null" json:"device_id"`
	Status      string     `gorm:"not null" json:"status"`
	Error       string     `json:"error,omitempty"`
	DeliveredAt *time.Time `json:"delivered_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

func (d *DeliveryLog) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.NewString()
	}
	return nil
}
