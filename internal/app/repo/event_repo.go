package repo

import (
	"context"
	"errors"

	endpointDomain "github.com/adesubomi/pigeon-server/internal/domain/endpoint"
	eventDomain "github.com/adesubomi/pigeon-server/internal/domain/event"
	"github.com/adesubomi/pigeon-server/pkg/apperr"
	"gorm.io/gorm"
)

type EventRepo struct {
	db *gorm.DB
}

func NewEventRepo(db *gorm.DB) *EventRepo {
	return &EventRepo{db: db}
}

func (r *EventRepo) FindActiveEndpointBySlug(ctx context.Context, slug string) (*endpointDomain.Endpoint, error) {
	var endpoint endpointDomain.Endpoint
	err := r.db.WithContext(ctx).
		Where(&endpointDomain.Endpoint{Slug: slug, IsActive: true}).
		First(&endpoint).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.NotFound("endpoint.not_found", "Endpoint not found")
	}
	if err != nil {
		return nil, apperr.Internal(err)
	}
	return &endpoint, nil
}

func (r *EventRepo) CreateEvent(ctx context.Context, event *eventDomain.Event) error {
	if err := r.db.WithContext(ctx).Create(event).Error; err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (r *EventRepo) FindUserEvent(ctx context.Context, userID, id string) (*eventDomain.Event, error) {
	var event eventDomain.Event
	err := r.db.WithContext(ctx).
		Model(&eventDomain.Event{}).
		Joins("JOIN endpoints ON endpoints.id = events.endpoint_id").
		Where(&eventDomain.Event{ID: id}).
		Where("endpoints.user_id = ?", userID).
		Where("endpoints.deleted_at IS NULL").
		Select("events.*").
		First(&event).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.NotFound("event.not_found", "Event not found")
	}
	if err != nil {
		return nil, apperr.Internal(err)
	}
	return &event, nil
}
