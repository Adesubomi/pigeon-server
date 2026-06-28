package repo

import (
	"context"

	"github.com/adesubomi/pigeon-server/internal/domain/auth"
	"github.com/adesubomi/pigeon-server/pkg/apperr"
	"gorm.io/gorm"
)

type AuthRepo struct {
	db *gorm.DB
}

func NewAuthRepo(db *gorm.DB) *AuthRepo {
	return &AuthRepo{db: db}
}

func (r *AuthRepo) WithTx(fn func(tx *gorm.DB) error) error {
	return fn(r.db)
}

func (r *AuthRepo) FindUserByID(ctx context.Context, id string) (*auth.User, error) {
	var user auth.User
	if err := r.db.WithContext(ctx).
		Where(&auth.User{ID: id}).
		First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepo) UpsertGitHubUser(ctx context.Context, user auth.User) (*auth.User, error) {
	if err := r.db.WithContext(ctx).
		Where(&auth.User{GithubID: user.GithubID}).
		Assign(auth.User{
			Email:     user.Email,
			Name:      user.Name,
			AvatarURL: user.AvatarURL,
		}).
		FirstOrCreate(&user).Error; err != nil {
		return nil, apperr.Internal(err)
	}
	return &user, nil
}
