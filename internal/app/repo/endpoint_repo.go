package repo

import (
	"context"
	"errors"

	deviceDomain "github.com/adesubomi/pigeon-server/internal/domain/device"
	domain "github.com/adesubomi/pigeon-server/internal/domain/endpoint"
	eventDomain "github.com/adesubomi/pigeon-server/internal/domain/event"
	"github.com/adesubomi/pigeon-server/pkg/apperr"
	"gorm.io/gorm"
)

type EndpointRepo struct {
	db *gorm.DB
}

func NewEndpointRepo(db *gorm.DB) *EndpointRepo {
	return &EndpointRepo{db: db}
}

func (r *EndpointRepo) CreateEndpoint(ctx context.Context, endpoint *domain.Endpoint) error {
	if err := r.db.WithContext(ctx).Create(endpoint).Error; err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (r *EndpointRepo) ListEndpoints(ctx context.Context, userID string) ([]domain.Endpoint, error) {
	var endpoints []domain.Endpoint
	if err := r.db.WithContext(ctx).
		Where(&domain.Endpoint{UserID: userID}).
		Order("created_at desc").
		Find(&endpoints).Error; err != nil {
		return nil, apperr.Internal(err)
	}
	return endpoints, nil
}

func (r *EndpointRepo) FindUserEndpoint(ctx context.Context, userID, id string) (*domain.Endpoint, error) {
	var endpoint domain.Endpoint
	err := r.db.WithContext(ctx).
		Where(&domain.Endpoint{ID: id, UserID: userID}).
		First(&endpoint).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.NotFound("endpoint.not_found", "Endpoint not found")
	}
	if err != nil {
		return nil, apperr.Internal(err)
	}
	return &endpoint, nil
}

func (r *EndpointRepo) SaveEndpoint(ctx context.Context, endpoint *domain.Endpoint) error {
	if err := r.db.WithContext(ctx).Save(endpoint).Error; err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (r *EndpointRepo) DeleteEndpoint(ctx context.Context, endpoint *domain.Endpoint) error {
	if err := r.db.WithContext(ctx).Delete(endpoint).Error; err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (r *EndpointRepo) CreatePairingCode(ctx context.Context, pairingCode *domain.PairingCode) error {
	if err := r.db.WithContext(ctx).Create(pairingCode).Error; err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (r *EndpointRepo) ListEndpointDevices(ctx context.Context, endpointID string) ([]domain.DeviceSummary, error) {
	var devices []domain.DeviceSummary
	if err := r.db.WithContext(ctx).
		Model(&deviceDomain.Device{}).
		Select("id, device_id, device_name, is_active, last_seen_at, created_at").
		Where(&deviceDomain.Device{EndpointID: endpointID}).
		Where("deleted_at IS NULL").
		Order("created_at desc").
		Scan(&devices).Error; err != nil {
		return nil, apperr.Internal(err)
	}
	return devices, nil
}

func (r *EndpointRepo) ListEndpointEvents(ctx context.Context, endpointID string) ([]domain.EventSummary, error) {
	var events []domain.EventSummary
	if err := r.db.WithContext(ctx).
		Model(&eventDomain.Event{}).
		Select("id, method, path, content_type, received_at, created_at").
		Where(&eventDomain.Event{EndpointID: endpointID}).
		Order("received_at desc").
		Scan(&events).Error; err != nil {
		return nil, apperr.Internal(err)
	}
	return events, nil
}
