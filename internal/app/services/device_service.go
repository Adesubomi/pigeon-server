package services

import (
	"context"
	"net/http"
	"strings"
	"time"

	domain "github.com/adesubomi/pigeon-server/internal/domain/device"
	"github.com/adesubomi/pigeon-server/internal/domain/endpoint"
	"github.com/adesubomi/pigeon-server/pkg/apperr"
	"github.com/adesubomi/pigeon-server/pkg/clock"
	"github.com/adesubomi/pigeon-server/pkg/respond"
	"github.com/adesubomi/pigeon-server/pkg/token"
)

type contextKey string

const deviceContextKey contextKey = "device.device"

type DeviceService struct {
	repo  DeviceRepository
	clock clock.Clock
}

type DeviceRepository interface {
	FindUsablePairingCode(context.Context, string, time.Time) (*endpoint.PairingCode, error)
	CreateDeviceAndUsePairingCode(context.Context, *domain.Device, *endpoint.PairingCode) error
	UpdateLastSeen(context.Context, *domain.Device, time.Time) error
	UpdateDevice(context.Context, *domain.Device, map[string]any) error
	FindDeviceByID(context.Context, string) (*domain.Device, error)
	DeleteDevice(context.Context, *domain.Device) error
	FindActiveDeviceByTokenHash(context.Context, string) (*domain.Device, error)
}

func NewDevice(repo DeviceRepository) *DeviceService {
	return &DeviceService{
		repo:  repo,
		clock: clock.RealClock{},
	}
}

func (s *DeviceService) PairDevice(ctx context.Context, input domain.PairDeviceInput) (*domain.PairDeviceResponse, error) {
	if strings.TrimSpace(input.Code) == "" {
		return nil, apperr.Validation(map[string]string{"code": "Pairing code is required"})
	}
	if strings.TrimSpace(input.DeviceName) == "" {
		return nil, apperr.Validation(map[string]string{"device_name": "Device name is required"})
	}

	now := s.clock.Now()
	pairingCode, err := s.repo.FindUsablePairingCode(ctx, token.Hash(strings.ToUpper(strings.TrimSpace(input.Code))), now)
	if err != nil {
		return nil, err
	}

	rawToken, err := token.GenerateURLSafe(32)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	deviceID := strings.TrimSpace(input.DeviceID)
	if deviceID == "" {
		deviceID, err = token.GenerateURLSafe(12)
		if err != nil {
			return nil, apperr.Internal(err)
		}
	}

	newDevice := domain.Device{
		EndpointID: pairingCode.EndpointID,
		DeviceID:   deviceID,
		DeviceName: strings.TrimSpace(input.DeviceName),
		TokenHash:  token.Hash(rawToken),
		IsActive:   true,
		LastSeenAt: &now,
	}

	pairingCode.UsedAt = &now
	if err := s.repo.CreateDeviceAndUsePairingCode(ctx, &newDevice, pairingCode); err != nil {
		return nil, err
	}

	return &domain.PairDeviceResponse{DeviceID: newDevice.ID, Token: rawToken}, nil
}

func (s *DeviceService) Heartbeat(ctx context.Context, device *domain.Device) (*domain.HeartbeatResponse, error) {
	now := s.clock.Now()
	if err := s.repo.UpdateLastSeen(ctx, device, now); err != nil {
		return nil, err
	}
	return &domain.HeartbeatResponse{LastSeenAt: now}, nil
}

func (s *DeviceService) UpdateDevice(ctx context.Context, input domain.UpdateDeviceInput) (*domain.DeviceResponse, error) {
	if input.Device == nil || input.Device.ID != input.ID {
		return nil, apperr.Forbidden("device.forbidden", "Device token cannot update this device")
	}

	updates := map[string]any{}
	if input.DeviceName != nil {
		name := strings.TrimSpace(*input.DeviceName)
		if name == "" {
			return nil, apperr.Validation(map[string]string{"device_name": "Device name is required"})
		}
		updates["device_name"] = name
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}
	if len(updates) > 0 {
		if err := s.repo.UpdateDevice(ctx, input.Device, updates); err != nil {
			return nil, err
		}
	}

	updated, err := s.repo.FindDeviceByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	return deviceToResponse(updated), nil
}

func (s *DeviceService) DeleteDevice(ctx context.Context, currentDevice *domain.Device, id string) error {
	if currentDevice == nil || currentDevice.ID != id {
		return apperr.Forbidden("device.forbidden", "Device token cannot delete this device")
	}
	return s.repo.DeleteDevice(ctx, currentDevice)
}

func (s *DeviceService) RequireDevice(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawToken := bearerToken(r)
		if rawToken == "" {
			rawToken = r.Header.Get("X-Pigeon-Device-Token")
		}
		if rawToken == "" {
			respond.Error(w, apperr.Unauthorized("device.unauthorized", "Device authentication required"))
			return
		}

		device, err := s.repo.FindActiveDeviceByTokenHash(r.Context(), token.Hash(rawToken))
		if err != nil {
			respond.Error(w, err)
			return
		}

		next.ServeHTTP(w, r.WithContext(ContextWithDevice(r.Context(), device)))
	})
}

func ContextWithDevice(ctx context.Context, device *domain.Device) context.Context {
	return context.WithValue(ctx, deviceContextKey, device)
}

func DeviceFromContext(ctx context.Context) (*domain.Device, bool) {
	device, ok := ctx.Value(deviceContextKey).(*domain.Device)
	return device, ok
}

func deviceToResponse(device *domain.Device) *domain.DeviceResponse {
	return &domain.DeviceResponse{
		ID:         device.ID,
		EndpointID: device.EndpointID,
		DeviceID:   device.DeviceID,
		DeviceName: device.DeviceName,
		IsActive:   device.IsActive,
		LastSeenAt: device.LastSeenAt,
		CreatedAt:  device.CreatedAt,
		UpdatedAt:  device.UpdatedAt,
	}
}
