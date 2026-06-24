package services

import (
	"context"

	domain "github.com/adesubomi/pigeon-server/internal/domain/push"
	"github.com/adesubomi/pigeon-server/pkg/apperr"
	"gorm.io/gorm"
)

type PushService struct {
	db  *gorm.DB
	hub *domain.Hub
}

type PushEventInput struct {
	EventID    string
	EndpointID string
	Payload    any
}

func NewPushSvc(db *gorm.DB, hub *domain.Hub) *PushService {
	return &PushService{db: db, hub: hub}
}

func (s *PushService) Hub() *domain.Hub {
	return s.hub
}

func (s *PushService) PushEvent(ctx context.Context, input PushEventInput) error {
	attempts := s.hub.SendToEndpoint(input.EndpointID, domain.Message{
		Event: "webhook.event",
		Data:  input.Payload,
	})

	for _, attempt := range attempts {
		log := domain.DeliveryLog{
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
