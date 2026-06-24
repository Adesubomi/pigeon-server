package services

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	endpointDomain "github.com/adesubomi/pigeon-server/internal/domain/endpoint"
	eventDomain "github.com/adesubomi/pigeon-server/internal/domain/event"
	"github.com/adesubomi/pigeon-server/pkg/apperr"
	"gorm.io/gorm"
)

type EventService struct {
	db      *gorm.DB
	pushSvc *PushService
}

func NewEvent(db *gorm.DB, pushSvc *PushService) *EventService {
	return &EventService{db: db, pushSvc: pushSvc}
}

func (s *EventService) ReceiveWebhook(ctx context.Context, input eventDomain.ReceiveWebhookInput) (*eventDomain.WebhookReceivedResponse, error) {
	var hookEndpoint endpointDomain.Endpoint
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

	webhookEvent := eventDomain.Event{
		EndpointID:  hookEndpoint.ID,
		Method:      input.Method,
		Path:        input.Path,
		HeadersJSON: headersJSON,
		QueryJSON:   queryJSON,
		Body:        input.Body,
		ContentType: input.ContentType,
		ReceivedAt:  time.Now(),
	}
	if err := s.db.WithContext(ctx).Create(&webhookEvent).Error; err != nil {
		return nil, apperr.Internal(err)
	}

	if err := s.pushSvc.PushEvent(ctx, PushEventInput{
		EventID:    webhookEvent.ID,
		EndpointID: hookEndpoint.ID,
		Payload:    eventToPayload(&webhookEvent),
	}); err != nil {
		return nil, err
	}

	return &eventDomain.WebhookReceivedResponse{EventID: webhookEvent.ID}, nil
}

func (s *EventService) GetEvent(ctx context.Context, userID, id string) (*eventDomain.EventResponse, error) {
	webhookEvent, err := s.findUserEvent(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return eventToResponse(webhookEvent), nil
}

func (s *EventService) ReplayEvent(ctx context.Context, input eventDomain.ReplayEventInput) (*eventDomain.ReplayEventResponse, error) {
	webhookEvent, err := s.findUserEvent(ctx, input.UserID, input.EventID)
	if err != nil {
		return nil, err
	}
	if err := s.pushSvc.PushEvent(ctx, PushEventInput{
		EventID:    webhookEvent.ID,
		EndpointID: webhookEvent.EndpointID,
		Payload:    eventToPayload(webhookEvent),
	}); err != nil {
		return nil, err
	}
	return &eventDomain.ReplayEventResponse{EventID: webhookEvent.ID, Status: "queued"}, nil
}

func (s *EventService) findUserEvent(ctx context.Context, userID, id string) (*eventDomain.Event, error) {
	var webhookEvent eventDomain.Event
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

func eventToResponse(webhookEvent *eventDomain.Event) *eventDomain.EventResponse {
	var headers any
	var query any
	_ = json.Unmarshal(webhookEvent.HeadersJSON, &headers)
	_ = json.Unmarshal(webhookEvent.QueryJSON, &query)

	return &eventDomain.EventResponse{
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

func eventToPayload(webhookEvent *eventDomain.Event) eventDomain.PushPayload {
	response := eventToResponse(webhookEvent)
	return eventDomain.PushPayload{
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
