package services

import (
	"context"

	domain "github.com/adesubomi/pigeon-server/internal/domain/push"
)

type PushService struct {
	repo PushRepository
	hub  *domain.Hub
}

type PushEventInput struct {
	EventID    string
	EndpointID string
	Payload    any
}

type PushRepository interface {
	CreateDeliveryLog(context.Context, *domain.DeliveryLog) error
}

func NewPushSvc(repo PushRepository, hub *domain.Hub) *PushService {
	return &PushService{repo: repo, hub: hub}
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
		if err := s.repo.CreateDeliveryLog(ctx, &log); err != nil {
			return err
		}
	}
	return nil
}
