package endpoint

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Endpoint struct {
	ID        string         `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    string         `gorm:"type:uuid;index;not null" json:"user_id"`
	Name      string         `gorm:"not null" json:"name"`
	Slug      string         `gorm:"uniqueIndex;not null" json:"slug"`
	IsActive  bool           `gorm:"not null;default:true" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (e *Endpoint) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.NewString()
	}
	return nil
}

type PairingCode struct {
	ID         string     `json:"id" gorm:"type:uuid;primaryKey"`
	EndpointID string     `json:"endpoint_id" gorm:"type:uuid;index;not null"`
	CodeHash   string     `json:"-" gorm:"uniqueIndex;not null"`
	ExpiresAt  time.Time  `json:"expires_at" gorm:"index;not null"`
	UsedAt     *time.Time `json:"used_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (p *PairingCode) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	return nil
}
