// Package users owns user profiles and account lifecycle.
package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/isyll/go-grpc-starter/gen/db"
	"github.com/isyll/go-grpc-starter/internal/errs"
	"github.com/isyll/go-grpc-starter/internal/store"
)

type Repository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id int64) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	ExistsByEmail(ctx context.Context, email string) bool
	UpdateLastLogin(ctx context.Context, id int64) error
	UpdatePasswordHash(ctx context.Context, id int64, hash string) error
	MarkEmailVerified(ctx context.Context, id int64) error
	UpdateProfile(ctx context.Context, id int64, upd ProfileUpdate) (*User, error)
	UpdateStatus(ctx context.Context, id int64, status UserStatus) error
	UpdateRole(ctx context.Context, id int64, role UserRole) error
	SoftDeleteByID(ctx context.Context, id int64) error
	List(ctx context.Context, offset, limit int) ([]User, int64, error)
}

type repository struct {
	store *store.Store
}

func NewRepository(s *store.Store) Repository {
	return &repository{store: s}
}

func toUser(r db.AuthUser) *User {
	return &User{
		ID:              r.ID,
		Email:           r.Email,
		PasswordHash:    r.PasswordHash,
		FirstName:       r.FirstName,
		LastName:        r.LastName,
		Avatar:          r.Avatar,
		Bio:             r.Bio,
		Status:          UserStatus(r.Status),
		Role:            UserRole(r.Role),
		EmailVerifiedAt: store.TimePtr(r.EmailVerifiedAt),
		LastLoginAt:     store.TimePtr(r.LastLoginAt),
		CreatedAt:       store.Time(r.CreatedAt),
		UpdatedAt:       store.Time(r.UpdatedAt),
		DeletedAt:       store.TimePtr(r.DeletedAt),
	}
}

func (r *repository) Create(ctx context.Context, user *User) error {
	return r.store.Run(ctx, func(ctx context.Context, q *db.Queries) error {
		row, err := q.CreateUser(ctx, db.CreateUserParams{
			Email:        user.Email,
			PasswordHash: user.PasswordHash,
			FirstName:    user.FirstName,
			LastName:     user.LastName,
		})
		if err != nil {
			return fmt.Errorf("create user: %w", err)
		}
		*user = *toUser(row)
		return nil
	})
}

func (r *repository) FindByID(ctx context.Context, id int64) (*User, error) {
	var out *User
	err := r.store.Run(ctx, func(ctx context.Context, q *db.Queries) error {
		row, err := q.GetUserByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errs.ErrUserNotFound
			}
			return fmt.Errorf("find user %d: %w", id, err)
		}
		out = toUser(row)
		return nil
	})
	return out, err
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var out *User
	err := r.store.Run(ctx, func(ctx context.Context, q *db.Queries) error {
		row, err := q.GetUserByEmail(ctx, email)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errs.ErrUserNotFound
			}
			return fmt.Errorf("find user by email: %w", err)
		}
		out = toUser(row)
		return nil
	})
	return out, err
}

func (r *repository) ExistsByEmail(ctx context.Context, email string) bool {
	var exists bool
	_ = r.store.Run(ctx, func(ctx context.Context, q *db.Queries) error {
		var err error
		exists, err = q.ExistsUserByEmail(ctx, email)
		return err
	})
	return exists
}

func (r *repository) UpdateLastLogin(ctx context.Context, id int64) error {
	return r.store.Run(ctx, func(ctx context.Context, q *db.Queries) error {
		return q.UpdateUserLastLogin(ctx, id)
	})
}

func (r *repository) UpdatePasswordHash(ctx context.Context, id int64, hash string) error {
	return r.store.Run(ctx, func(ctx context.Context, q *db.Queries) error {
		return q.UpdateUserPasswordHash(ctx, db.UpdateUserPasswordHashParams{ID: id, PasswordHash: hash})
	})
}

func (r *repository) MarkEmailVerified(ctx context.Context, id int64) error {
	return r.store.Run(ctx, func(ctx context.Context, q *db.Queries) error {
		return q.MarkUserEmailVerified(ctx, id)
	})
}

func (r *repository) UpdateProfile(
	ctx context.Context, id int64, upd ProfileUpdate,
) (*User, error) {
	var out *User
	err := r.store.Run(ctx, func(ctx context.Context, q *db.Queries) error {
		row, err := q.UpdateUserProfile(ctx, db.UpdateUserProfileParams{
			FirstName: upd.FirstName,
			LastName:  upd.LastName,
			Bio:       upd.Bio,
			Avatar:    upd.Avatar,
			ID:        id,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errs.ErrUserNotFound
			}
			return fmt.Errorf("update profile: %w", err)
		}
		out = toUser(row)
		return nil
	})
	return out, err
}

func (r *repository) UpdateStatus(ctx context.Context, id int64, status UserStatus) error {
	return r.store.Run(ctx, func(ctx context.Context, q *db.Queries) error {
		return q.UpdateUserStatus(ctx, db.UpdateUserStatusParams{ID: id, Status: db.AuthUserStatus(status)})
	})
}

func (r *repository) UpdateRole(ctx context.Context, id int64, role UserRole) error {
	return r.store.Run(ctx, func(ctx context.Context, q *db.Queries) error {
		return q.UpdateUserRole(ctx, db.UpdateUserRoleParams{ID: id, Role: db.AuthUserRole(role)})
	})
}

func (r *repository) SoftDeleteByID(ctx context.Context, id int64) error {
	return r.store.Run(ctx, func(ctx context.Context, q *db.Queries) error {
		return q.SoftDeleteUser(ctx, id)
	})
}

func (r *repository) List(ctx context.Context, offset, limit int) ([]User, int64, error) {
	var users []User
	var total int64
	err := r.store.Run(ctx, func(ctx context.Context, q *db.Queries) error {
		var err error
		total, err = q.CountUsers(ctx)
		if err != nil {
			return fmt.Errorf("count users: %w", err)
		}
		rows, err := q.ListUsers(ctx, db.ListUsersParams{Limit: int32(limit), Offset: int32(offset)})
		if err != nil {
			return fmt.Errorf("list users: %w", err)
		}
		users = make([]User, len(rows))
		for i, row := range rows {
			users[i] = *toUser(row)
		}
		return nil
	})
	return users, total, err
}
