package push

import (
	"context"

	"github.com/adesubomi/pigeon-server/pkg/apperr"
	"gorm.io/gorm"
)

type Service struct {
	db  *gorm.DB
	hub *Hub
}

type PushEventInput struct {
	EventID    string
	EndpointID string
	Payload    any
}

func NewService(db *gorm.DB, hub *Hub) *Service {
	return &Service{db: db, hub: hub}
}

func (s *Service) Hub() *Hub {
	return s.hub
}

func (s *Service) PushEvent(ctx context.Context, input PushEventInput) error {
	attempts := s.hub.SendToEndpoint(input.EndpointID, Message{
		Event: "webhook.event",
		Data:  input.Payload,
	})

	for _, attempt := range attempts {
		log := DeliveryLog{
			EventID:     input.EventID,
			DeviceID:    attempt.DeviceID,
			Status:      attempt.Status,
			Error:       attempt.Error,
			DeliveredAt: attempt.DeliveredAt,
		}
		if err := s.db.WithContext(ctx).Create(&log).Error; err != nil {
			return apperr.Internal(err)
		}
	}
	return nil
}
