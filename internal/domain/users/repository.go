// Package users owns user profiles and account lifecycle.
package users

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	apperrors "github.com/isyll/go-api-starter/internal/errors"
	"github.com/isyll/go-api-starter/internal/models"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id int64) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	ExistsByEmail(ctx context.Context, email string) bool
	UpdateLastLogin(ctx context.Context, id int64) error
	UpdatePasswordHash(ctx context.Context, id int64, hash string) error
	MarkEmailVerified(ctx context.Context, id int64) error
	UpdateProfile(ctx context.Context, id int64, fields map[string]any) (*models.User, error)
	UpdateStatus(ctx context.Context, id int64, status models.UserStatus) error
	UpdateRole(ctx context.Context, id int64, role models.UserRole) error
	SoftDeleteByID(ctx context.Context, id int64) error
	List(ctx context.Context, offset, limit int) ([]models.User, int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *repository) FindByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Preload("UserSettings").First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		panic(fmt.Errorf("find user %d: %w", id, err))
	}
	return &user, nil
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Where("email = ?", strings.ToLower(email)).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		panic(fmt.Errorf("find user by email: %w", err))
	}
	return &user, nil
}

func (r *repository) ExistsByEmail(ctx context.Context, email string) bool {
	var count int64
	r.db.WithContext(ctx).Model(&models.User{}).
		Where("email = ?", strings.ToLower(email)).
		Count(&count)
	return count > 0
}

func (r *repository) UpdateLastLogin(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", id).
		Update("last_login_at", time.Now().UTC()).Error
}

func (r *repository) UpdatePasswordHash(ctx context.Context, id int64, hash string) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", id).
		Update("password_hash", hash).Error
}

func (r *repository) MarkEmailVerified(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", id).
		Update("email_verified_at", time.Now().UTC()).Error
}

func (r *repository) UpdateProfile(
	ctx context.Context, id int64, fields map[string]any,
) (*models.User, error) {
	if len(fields) == 0 {
		return r.FindByID(ctx, id)
	}
	if err := r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", id).Updates(fields).Error; err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}
	return r.FindByID(ctx, id)
}

func (r *repository) UpdateStatus(ctx context.Context, id int64, status models.UserStatus) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", id).Update("status", status).Error
}

func (r *repository) UpdateRole(ctx context.Context, id int64, role models.UserRole) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", id).Update("role", role).Error
}

func (r *repository) SoftDeleteByID(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, id).Error
}

func (r *repository) List(ctx context.Context, offset, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").Offset(offset).Limit(limit).
		Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	return users, total, nil
}
