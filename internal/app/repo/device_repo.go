package repo

import (
	"context"
	"errors"
	"time"

	deviceDomain "github.com/adesubomi/pigeon-server/internal/domain/device"
	endpointDomain "github.com/adesubomi/pigeon-server/internal/domain/endpoint"
	"github.com/adesubomi/pigeon-server/pkg/apperr"
	"gorm.io/gorm"
)

type DeviceRepo struct {
	db *gorm.DB
}

func NewDeviceRepo(db *gorm.DB) *DeviceRepo {
	return &DeviceRepo{db: db}
}

func (r *DeviceRepo) FindUsablePairingCode(ctx context.Context, codeHash string, now time.Time) (*endpointDomain.PairingCode, error) {
	var pairingCode endpointDomain.PairingCode
	err := r.db.WithContext(ctx).
		Where(&endpointDomain.PairingCode{CodeHash: codeHash}).
		Where("used_at IS NULL").
		Where("expires_at > ?", now).
		First(&pairingCode).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.BadRequest("pairing_code.invalid", "Pairing code is invalid or expired")
	}
	if err != nil {
		return nil, apperr.Internal(err)
	}
	return &pairingCode, nil
}

func (r *DeviceRepo) CreateDeviceAndUsePairingCode(ctx context.Context, device *deviceDomain.Device, pairingCode *endpointDomain.PairingCode) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(device).Error; err != nil {
			return err
		}
		return tx.Save(pairingCode).Error
	}); err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (r *DeviceRepo) UpdateLastSeen(ctx context.Context, device *deviceDomain.Device, at time.Time) error {
	if err := r.db.WithContext(ctx).Model(device).Updates(map[string]any{"last_seen_at": at}).Error; err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (r *DeviceRepo) UpdateDevice(ctx context.Context, device *deviceDomain.Device, updates map[string]any) error {
	if err := r.db.WithContext(ctx).Model(device).Updates(updates).Error; err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (r *DeviceRepo) FindDeviceByID(ctx context.Context, id string) (*deviceDomain.Device, error) {
	var device deviceDomain.Device
	if err := r.db.WithContext(ctx).
		Where(&deviceDomain.Device{ID: id}).
		First(&device).Error; err != nil {
		return nil, apperr.Internal(err)
	}
	return &device, nil
}

func (r *DeviceRepo) DeleteDevice(ctx context.Context, device *deviceDomain.Device) error {
	if err := r.db.WithContext(ctx).Delete(device).Error; err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (r *DeviceRepo) FindActiveDeviceByTokenHash(ctx context.Context, tokenHash string) (*deviceDomain.Device, error) {
	var device deviceDomain.Device
	err := r.db.WithContext(ctx).
		Where(&deviceDomain.Device{TokenHash: tokenHash, IsActive: true}).
		First(&device).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.Unauthorized("device.unauthorized", "Device authentication required")
	}
	if err != nil {
		return nil, apperr.Internal(err)
	}
	return &device, nil
}
