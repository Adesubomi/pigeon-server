package event

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Event struct {
	ID          string    `gorm:"type:uuid;primaryKey" json:"id"`
	EndpointID  string    `gorm:"type:uuid;index;not null" json:"endpoint_id"`
	Method      string    `gorm:"not null" json:"method"`
	Path        string    `json:"path"`
	HeadersJSON []byte    `gorm:"type:jsonb;not null;default:'{}'" json:"headers"`
	QueryJSON   []byte    `gorm:"type:jsonb;not null;default:'{}'" json:"query"`
	Body        []byte    `json:"body"`
	ContentType string    `json:"content_type"`
	ReceivedAt  time.Time `gorm:"index;not null" json:"received_at"`
	CreatedAt   time.Time `json:"created_at"`
}

func (e *Event) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.NewString()
	}
	return nil
}
