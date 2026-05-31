package event

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/adesubomi/pigeon-server/internal/domain/endpoint"
	"github.com/adesubomi/pigeon-server/internal/domain/push"
	"github.com/adesubomi/pigeon-server/pkg/apperr"
	"github.com/adesubomi/pigeon-server/pkg/clock"
	"gorm.io/gorm"
)

type Service struct {
	db      *gorm.DB
	clock   clock.Clock
	pushSvc *push.Service
}

func NewService(db *gorm.DB, clock clock.Clock, pushSvc *push.Service) *Service {
	return &Service{db: db, clock: clock, pushSvc: pushSvc}
}

func (s *Service) ReceiveWebhook(ctx context.Context, input ReceiveWebhookInput) (*WebhookReceivedResponse, error) {
	var hookEndpoint endpoint.Endpoint
	err := s.db.WithContext(ctx).First(&hookEndpoint, "slug = ? AND is_active = true", input.Slug).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.NotFound("endpoint.not_found", "Endpoint not found")
	}
	if err != nil {
		return nil, apperr.Internal(err)
	}

	headersJSON, err := json.Marshal(input.Headers)
	if err != nil {
		return nil, apperr.BadRequest("request.invalid_headers", "Invalid request headers")
	}
	queryJSON, err := json.Marshal(input.Query)
	if err != nil {
		return nil, apperr.BadRequest("request.invalid_query", "Invalid query string")
	}

	webhookEvent := Event{
		EndpointID:  hookEndpoint.ID,
		Method:      input.Method,
		Path:        input.Path,
		HeadersJSON: headersJSON,
		QueryJSON:   queryJSON,
		Body:        input.Body,
		ContentType: input.ContentType,
		ReceivedAt:  s.clock.Now(),
	}
	if err := s.db.WithContext(ctx).Create(&webhookEvent).Error; err != nil {
		return nil, apperr.Internal(err)
	}

	if err := s.pushSvc.PushEvent(ctx, push.PushEventInput{
		EventID:    webhookEvent.ID,
		EndpointID: hookEndpoint.ID,
		Payload:    eventToPayload(&webhookEvent),
	}); err != nil {
		return nil, err
	}

	return &WebhookReceivedResponse{EventID: webhookEvent.ID}, nil
}

func (s *Service) GetEvent(ctx context.Context, userID, id string) (*EventResponse, error) {
	webhookEvent, err := s.findUserEvent(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return eventToResponse(webhookEvent), nil
}

func (s *Service) ReplayEvent(ctx context.Context, input ReplayEventInput) (*ReplayEventResponse, error) {
	webhookEvent, err := s.findUserEvent(ctx, input.UserID, input.EventID)
	if err != nil {
		return nil, err
	}
	if err := s.pushSvc.PushEvent(ctx, push.PushEventInput{
		EventID:    webhookEvent.ID,
		EndpointID: webhookEvent.EndpointID,
		Payload:    eventToPayload(webhookEvent),
	}); err != nil {
		return nil, err
	}
	return &ReplayEventResponse{EventID: webhookEvent.ID, Status: "queued"}, nil
}

func (s *Service) findUserEvent(ctx context.Context, userID, id string) (*Event, error) {
	var webhookEvent Event
	err := s.db.WithContext(ctx).
		Table("events").
		Joins("JOIN endpoints ON endpoints.id = events.endpoint_id").
		Where("events.id = ? AND endpoints.user_id = ? AND endpoints.deleted_at IS NULL", id, userID).
		Select("events.*").
		First(&webhookEvent).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.NotFound("event.not_found", "Event not found")
	}
	if err != nil {
		return nil, apperr.Internal(err)
	}
	return &webhookEvent, nil
}

func eventToResponse(webhookEvent *Event) *EventResponse {
	var headers any
	var query any
	_ = json.Unmarshal(webhookEvent.HeadersJSON, &headers)
	_ = json.Unmarshal(webhookEvent.QueryJSON, &query)

	return &EventResponse{
		ID:          webhookEvent.ID,
		EndpointID:  webhookEvent.EndpointID,
		Method:      webhookEvent.Method,
		Path:        webhookEvent.Path,
		Headers:     headers,
		Query:       query,
		Body:        string(webhookEvent.Body),
		ContentType: webhookEvent.ContentType,
		ReceivedAt:  webhookEvent.ReceivedAt,
		CreatedAt:   webhookEvent.CreatedAt,
	}
}

func eventToPayload(webhookEvent *Event) PushPayload {
	response := eventToResponse(webhookEvent)
	return PushPayload{
		EventID:     response.ID,
		EndpointID:  response.EndpointID,
		Method:      response.Method,
		Path:        response.Path,
		Headers:     response.Headers,
		Query:       response.Query,
		Body:        response.Body,
		ContentType: response.ContentType,
		ReceivedAt:  response.ReceivedAt,
	}
}
