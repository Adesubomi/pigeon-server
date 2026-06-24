package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	domain "github.com/adesubomi/pigeon-server/internal/domain/endpoint"
	"github.com/adesubomi/pigeon-server/pkg/apperr"
	"github.com/adesubomi/pigeon-server/pkg/token"
	"gorm.io/gorm"
)

type EndpointService struct {
	db *gorm.DB
}

func NewEndpointSvc(db *gorm.DB) *EndpointService {
	return &EndpointService{db: db}
}

func (s *EndpointService) CreateEndpoint(ctx context.Context, input domain.CreateEndpointInput) (*domain.EndpointResponse, error) {
	if strings.TrimSpace(input.Name) == "" {
		return nil, apperr.Validation(map[string]string{"name": "Name is required"})
	}

	suffix, err := token.GenerateURLSafe(4)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	endpoint := domain.Endpoint{
		UserID:   input.UserID,
		Name:     strings.TrimSpace(input.Name),
		Slug:     slugify(input.Name) + "-" + strings.ToLower(suffix),
		IsActive: true,
	}
	if err := s.db.WithContext(ctx).Create(&endpoint).Error; err != nil {
		return nil, apperr.Internal(err)
	}

	return endpointToResponse(&endpoint), nil
}

func (s *EndpointService) ListEndpoints(ctx context.Context, userID string) ([]domain.EndpointResponse, error) {
	var endpoints []domain.Endpoint
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Find(&endpoints).Error; err != nil {
		return nil, apperr.Internal(err)
	}

	responses := make([]domain.EndpointResponse, 0, len(endpoints))
	for i := range endpoints {
		responses = append(responses, *endpointToResponse(&endpoints[i]))
	}
	return responses, nil
}

func (s *EndpointService) GetEndpoint(ctx context.Context, userID, id string) (*domain.EndpointResponse, error) {
	endpoint, err := s.findUserEndpoint(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return endpointToResponse(endpoint), nil
}

func (s *EndpointService) UpdateEndpoint(ctx context.Context, input domain.UpdateEndpointInput) (*domain.EndpointResponse, error) {
	endpoint, err := s.findUserEndpoint(ctx, input.UserID, input.ID)
	if err != nil {
		return nil, err
	}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, apperr.Validation(map[string]string{"name": "Name is required"})
		}
		endpoint.Name = name
	}
	if input.IsActive != nil {
		endpoint.IsActive = *input.IsActive
	}
	if err := s.db.WithContext(ctx).Save(endpoint).Error; err != nil {
		return nil, apperr.Internal(err)
	}
	return endpointToResponse(endpoint), nil
}

func (s *EndpointService) DeleteEndpoint(ctx context.Context, userID, id string) error {
	endpoint, err := s.findUserEndpoint(ctx, userID, id)
	if err != nil {
		return err
	}
	if err := s.db.WithContext(ctx).Delete(endpoint).Error; err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (s *EndpointService) GeneratePairingCode(ctx context.Context, userID, endpointID string) (*domain.PairingCodeResponse, error) {
	if _, err := s.findUserEndpoint(ctx, userID, endpointID); err != nil {
		return nil, err
	}

	code, err := token.GeneratePairingCode()
	if err != nil {
		return nil, apperr.Internal(err)
	}

	pairingCode := domain.PairingCode{
		EndpointID: endpointID,
		CodeHash:   token.Hash(code),
		ExpiresAt:  time.Now().Add(10 * time.Minute),
	}
	if err := s.db.WithContext(ctx).Create(&pairingCode).Error; err != nil {
		return nil, apperr.Internal(err)
	}

	return &domain.PairingCodeResponse{Code: code, ExpiresAt: pairingCode.ExpiresAt}, nil
}

func (s *EndpointService) ListEndpointDevices(ctx context.Context, userID, endpointID string) ([]domain.DeviceSummary, error) {
	if _, err := s.findUserEndpoint(ctx, userID, endpointID); err != nil {
		return nil, err
	}

	var devices []domain.DeviceSummary
	if err := s.db.WithContext(ctx).
		Table("devices").
		Select("id, device_id, device_name, is_active, last_seen_at, created_at").
		Where("endpoint_id = ? AND deleted_at IS NULL", endpointID).
		Order("created_at desc").
		Scan(&devices).Error; err != nil {
		return nil, apperr.Internal(err)
	}
	return devices, nil
}

func (s *EndpointService) ListEndpointEvents(ctx context.Context, userID, endpointID string) ([]domain.EventSummary, error) {
	if _, err := s.findUserEndpoint(ctx, userID, endpointID); err != nil {
		return nil, err
	}

	var events []domain.EventSummary
	if err := s.db.WithContext(ctx).
		Table("events").
		Select("id, method, path, content_type, received_at, created_at").
		Where("endpoint_id = ?", endpointID).
		Order("received_at desc").
		Scan(&events).Error; err != nil {
		return nil, apperr.Internal(err)
	}
	return events, nil
}

func (s *EndpointService) findUserEndpoint(ctx context.Context, userID, id string) (*domain.Endpoint, error) {
	var endpoint domain.Endpoint
	err := s.db.WithContext(ctx).First(&endpoint, "id = ? AND user_id = ?", id, userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.NotFound("endpoint.not_found", "Endpoint not found")
	}
	if err != nil {
		return nil, apperr.Internal(err)
	}
	return &endpoint, nil
}

func endpointToResponse(in *domain.Endpoint) *domain.EndpointResponse {
	return &domain.EndpointResponse{
		ID:        in.ID,
		Name:      in.Name,
		Slug:      in.Slug,
		IsActive:  in.IsActive,
		CreatedAt: in.CreatedAt,
		UpdatedAt: in.UpdatedAt,
	}
}

var slugPattern = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(value string) string {
	slug := strings.Trim(slugPattern.ReplaceAllString(strings.ToLower(value), "-"), "-")
	if slug == "" {
		return fmt.Sprintf("endpoint-%d", time.Now().Unix())
	}
	return slug
}
