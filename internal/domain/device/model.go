package device

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Device struct {
	ID         string         `gorm:"type:uuid;primaryKey" json:"id"`
	EndpointID string         `gorm:"type:uuid;index;not null" json:"endpoint_id"`
	DeviceID   string         `gorm:"index;not null" json:"device_id"`
	DeviceName string         `gorm:"not null" json:"device_name"`
	TokenHash  string         `gorm:"uniqueIndex;not null" json:"-"`
	IsActive   bool           `gorm:"not null;default:true" json:"is_active"`
	LastSeenAt *time.Time     `json:"last_seen_at"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (d *Device) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.NewString()
	}
	return nil
}
