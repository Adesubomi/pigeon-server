package repo

import (
	"context"

	domain "github.com/adesubomi/pigeon-server/internal/domain/push"
	"github.com/adesubomi/pigeon-server/pkg/apperr"
	"gorm.io/gorm"
)

type PushRepo struct {
	db *gorm.DB
}

func NewPushRepo(db *gorm.DB) *PushRepo {
	return &PushRepo{db: db}
}

func (r *PushRepo) CreateDeliveryLog(ctx context.Context, log *domain.DeliveryLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return apperr.Internal(err)
	}
	return nil
}
