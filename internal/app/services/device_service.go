package services

import (
	"context"
	"errors"
	"net/http"
	"strings"

	domain "github.com/adesubomi/pigeon-server/internal/domain/device"
	"github.com/adesubomi/pigeon-server/internal/domain/endpoint"
	"github.com/adesubomi/pigeon-server/pkg/apperr"
	"github.com/adesubomi/pigeon-server/pkg/clock"
	"github.com/adesubomi/pigeon-server/pkg/respond"
	"github.com/adesubomi/pigeon-server/pkg/token"
	"gorm.io/gorm"
)

type contextKey string

const deviceContextKey contextKey = "device.device"

type DeviceService struct {
	db    *gorm.DB
	clock clock.Clock
}

func NewDevice(db *gorm.DB) *DeviceService {
	return &DeviceService{
		db: db,
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
	var pairingCode endpoint.PairingCode
	err := s.db.WithContext(ctx).
		Where("code_hash = ? AND used_at IS NULL AND expires_at > ?", token.Hash(strings.ToUpper(strings.TrimSpace(input.Code))), now).
		First(&pairingCode).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.BadRequest("pairing_code.invalid", "Pairing code is invalid or expired")
	}
	if err != nil {
		return nil, apperr.Internal(err)
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

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&newDevice).Error; err != nil {
			return err
		}
		pairingCode.UsedAt = &now
		return tx.Save(&pairingCode).Error
	})
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return &domain.PairDeviceResponse{DeviceID: newDevice.ID, Token: rawToken}, nil
}

func (s *DeviceService) Heartbeat(ctx context.Context, device *domain.Device) (*domain.HeartbeatResponse, error) {
	now := s.clock.Now()
	if err := s.db.WithContext(ctx).Model(device).Update("last_seen_at", now).Error; err != nil {
		return nil, apperr.Internal(err)
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
		if err := s.db.WithContext(ctx).Model(input.Device).Updates(updates).Error; err != nil {
			return nil, apperr.Internal(err)
		}
	}

	var updated domain.Device
	if err := s.db.WithContext(ctx).First(&updated, "id = ?", input.ID).Error; err != nil {
		return nil, apperr.Internal(err)
	}
	return deviceToResponse(&updated), nil
}

func (s *DeviceService) DeleteDevice(ctx context.Context, currentDevice *domain.Device, id string) error {
	if currentDevice == nil || currentDevice.ID != id {
		return apperr.Forbidden("device.forbidden", "Device token cannot delete this device")
	}
	if err := s.db.WithContext(ctx).Delete(currentDevice).Error; err != nil {
		return apperr.Internal(err)
	}
	return nil
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

		var device domain.Device
		err := s.db.WithContext(r.Context()).
			First(&device, "token_hash = ? AND is_active = true", token.Hash(rawToken)).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respond.Error(w, apperr.Unauthorized("device.unauthorized", "Device authentication required"))
			return
		}
		if err != nil {
			respond.Error(w, apperr.Internal(err))
			return
		}

		next.ServeHTTP(w, r.WithContext(ContextWithDevice(r.Context(), &device)))
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
