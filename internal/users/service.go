package users

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/isyll/go-grpc-starter/internal/errs"
	"github.com/isyll/go-grpc-starter/internal/errs/codes"
	"github.com/isyll/go-grpc-starter/internal/events"
	"github.com/isyll/go-grpc-starter/internal/platform/storage"
	"github.com/isyll/go-grpc-starter/pkg/id"
	"github.com/isyll/go-grpc-starter/pkg/logger"
)

// MaxAvatarBytes bounds an uploaded avatar; the gRPC handler enforces it while
// streaming so no oversized payload is ever buffered.
const MaxAvatarBytes = 5 << 20

var avatarContentTypes = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/webp": "webp",
}

type ProfileUpdate struct {
	FirstName *string
	LastName  *string
	Bio       *string
	Avatar    *string
}

type Service struct {
	repo     Repository
	sessions SessionRevoker
	bus      *events.Bus
	storage  storage.Storage
	logger   *logger.Logger
}

func NewService(
	repo Repository,
	sessions SessionRevoker,
	bus *events.Bus,
	store storage.Storage,
	logx *logger.Logger,
) *Service {
	return &Service{repo: repo, sessions: sessions, bus: bus, storage: store, logger: logx}
}

// UploadAvatar validates and stores an avatar image, then persists its URL on
// the user. The bytes are already bounded by MaxAvatarBytes by the caller.
func (s *Service) UploadAvatar(
	ctx context.Context, userID int64, contentType string, data []byte,
) (string, error) {
	if s.storage == nil {
		return "", errs.Internal(codes.StorageUnavailable, "storage.unavailable")
	}
	ext, ok := avatarContentTypes[contentType]
	if !ok {
		return "", errs.BadRequest(codes.AvatarInvalidType, "user.avatar_invalid_type")
	}
	if len(data) == 0 || len(data) > MaxAvatarBytes {
		return "", errs.BadRequest(codes.AvatarTooLarge, "user.avatar_too_large")
	}

	key := fmt.Sprintf("avatars/%d/%s.%s", userID, id.NewUUIDNoDash(), ext)
	if err := s.storage.Put(ctx, key, bytes.NewReader(data), int64(len(data)), contentType); err != nil {
		s.logger.Error("avatar upload failed", "error", err, "user_id", userID)
		return "", errs.Internal(codes.UploadFailed, "storage.upload_failed")
	}

	url := s.storage.PublicURL(key)
	if _, err := s.repo.UpdateProfile(ctx, userID, ProfileUpdate{Avatar: &url}); err != nil {
		return "", err
	}
	return url, nil
}

func (s *Service) Get(ctx context.Context, id int64) (*User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) UpdateProfile(
	ctx context.Context, id int64, upd ProfileUpdate,
) (*User, error) {
	return s.repo.UpdateProfile(ctx, id, upd)
}

func (s *Service) DeleteAccount(ctx context.Context, id int64) error {
	if err := s.repo.SoftDeleteByID(ctx, id); err != nil {
		return err
	}
	if err := s.sessions.RevokeAllSessions(ctx, id, "account_deleted"); err != nil {
		s.logger.Warn("revoke sessions on delete failed", "error", err, "user_id", id)
	}
	if err := s.bus.Publish(ctx, &events.UserAccountDeleted{
		UserID:     id,
		OccurredAt: time.Now().UTC(),
	}); err != nil {
		s.logger.Warn("publish account deleted failed", "error", err, "user_id", id)
	}
	return nil
}

func (s *Service) List(ctx context.Context, offset, limit int) ([]User, int64, error) {
	return s.repo.List(ctx, offset, limit)
}

func (s *Service) SetStatus(ctx context.Context, id int64, status UserStatus) error {
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *Service) SetRole(ctx context.Context, id int64, role UserRole) error {
	return s.repo.UpdateRole(ctx, id, role)
}
