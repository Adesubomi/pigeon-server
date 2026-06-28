package services

import (
	"context"
	"encoding/json"
	"time"

	endpointDomain "github.com/adesubomi/pigeon-server/internal/domain/endpoint"
	eventDomain "github.com/adesubomi/pigeon-server/internal/domain/event"
	"github.com/adesubomi/pigeon-server/pkg/apperr"
)

type EventService struct {
	repo    EventRepository
	pushSvc *PushService
}

type EventRepository interface {
	FindActiveEndpointBySlug(context.Context, string) (*endpointDomain.Endpoint, error)
	CreateEvent(context.Context, *eventDomain.Event) error
	FindUserEvent(context.Context, string, string) (*eventDomain.Event, error)
}

func NewEvent(repo EventRepository, pushSvc *PushService) *EventService {
	return &EventService{repo: repo, pushSvc: pushSvc}
}

func (s *EventService) ReceiveWebhook(ctx context.Context, input eventDomain.ReceiveWebhookInput) (*eventDomain.WebhookReceivedResponse, error) {
	hookEndpoint, err := s.repo.FindActiveEndpointBySlug(ctx, input.Slug)
	if err != nil {
		return nil, err
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
	if err := s.repo.CreateEvent(ctx, &webhookEvent); err != nil {
		return nil, err
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
	webhookEvent, err := s.repo.FindUserEvent(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return eventToResponse(webhookEvent), nil
}

func (s *EventService) ReplayEvent(ctx context.Context, input eventDomain.ReplayEventInput) (*eventDomain.ReplayEventResponse, error) {
	webhookEvent, err := s.repo.FindUserEvent(ctx, input.UserID, input.EventID)
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
