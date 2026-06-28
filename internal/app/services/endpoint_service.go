package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	domain "github.com/adesubomi/pigeon-server/internal/domain/endpoint"
	"github.com/adesubomi/pigeon-server/pkg/apperr"
	"github.com/adesubomi/pigeon-server/pkg/token"
)

type EndpointService struct {
	repo EndpointRepository
}

type EndpointRepository interface {
	CreateEndpoint(context.Context, *domain.Endpoint) error
	ListEndpoints(context.Context, string) ([]domain.Endpoint, error)
	FindUserEndpoint(context.Context, string, string) (*domain.Endpoint, error)
	SaveEndpoint(context.Context, *domain.Endpoint) error
	DeleteEndpoint(context.Context, *domain.Endpoint) error
	CreatePairingCode(context.Context, *domain.PairingCode) error
	ListEndpointDevices(context.Context, string) ([]domain.DeviceSummary, error)
	ListEndpointEvents(context.Context, string) ([]domain.EventSummary, error)
}

func NewEndpointSvc(repo EndpointRepository) *EndpointService {
	return &EndpointService{repo: repo}
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
	if err := s.repo.CreateEndpoint(ctx, &endpoint); err != nil {
		return nil, err
	}

	return endpointToResponse(&endpoint), nil
}

func (s *EndpointService) ListEndpoints(ctx context.Context, userID string) ([]domain.EndpointResponse, error) {
	endpoints, err := s.repo.ListEndpoints(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]domain.EndpointResponse, 0, len(endpoints))
	for i := range endpoints {
		responses = append(responses, *endpointToResponse(&endpoints[i]))
	}
	return responses, nil
}

func (s *EndpointService) GetEndpoint(ctx context.Context, userID, id string) (*domain.EndpointResponse, error) {
	endpoint, err := s.repo.FindUserEndpoint(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return endpointToResponse(endpoint), nil
}

func (s *EndpointService) UpdateEndpoint(ctx context.Context, input domain.UpdateEndpointInput) (*domain.EndpointResponse, error) {
	endpoint, err := s.repo.FindUserEndpoint(ctx, input.UserID, input.ID)
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
	if err := s.repo.SaveEndpoint(ctx, endpoint); err != nil {
		return nil, err
	}
	return endpointToResponse(endpoint), nil
}

func (s *EndpointService) DeleteEndpoint(ctx context.Context, userID, id string) error {
	endpoint, err := s.repo.FindUserEndpoint(ctx, userID, id)
	if err != nil {
		return err
	}
	return s.repo.DeleteEndpoint(ctx, endpoint)
}

func (s *EndpointService) GeneratePairingCode(ctx context.Context, userID, endpointID string) (*domain.PairingCodeResponse, error) {
	if _, err := s.repo.FindUserEndpoint(ctx, userID, endpointID); err != nil {
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
	if err := s.repo.CreatePairingCode(ctx, &pairingCode); err != nil {
		return nil, err
	}

	return &domain.PairingCodeResponse{Code: code, ExpiresAt: pairingCode.ExpiresAt}, nil
}

func (s *EndpointService) ListEndpointDevices(ctx context.Context, userID, endpointID string) ([]domain.DeviceSummary, error) {
	if _, err := s.repo.FindUserEndpoint(ctx, userID, endpointID); err != nil {
		return nil, err
	}

	return s.repo.ListEndpointDevices(ctx, endpointID)
}

func (s *EndpointService) ListEndpointEvents(ctx context.Context, userID, endpointID string) ([]domain.EventSummary, error) {
	if _, err := s.repo.FindUserEndpoint(ctx, userID, endpointID); err != nil {
		return nil, err
	}

	return s.repo.ListEndpointEvents(ctx, endpointID)
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
