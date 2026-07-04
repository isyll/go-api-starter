package grpc

import (
	"context"

	"github.com/isyll/go-grpc-starter/internal/domain/notifications"
	"github.com/isyll/go-grpc-starter/internal/domain/settings"
	"github.com/isyll/go-grpc-starter/internal/domain/users"
	apiv1 "github.com/isyll/go-grpc-starter/internal/gen/api/v1"
	"github.com/isyll/go-grpc-starter/pkg/idenc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserServer struct {
	apiv1.UnimplementedUserServiceServer
	users    *users.Service
	settings *settings.Service
	notifs   *notifications.Service
	enc      idenc.IDEncoder
}

func NewUserServer(
	u *users.Service,
	s *settings.Service,
	n *notifications.Service,
	enc idenc.IDEncoder,
) *UserServer {
	return &UserServer{users: u, settings: s, notifs: n, enc: enc}
}

func (s *UserServer) GetMe(ctx context.Context, _ *emptypb.Empty) (*apiv1.User, error) {
	u, err := s.users.Get(ctx, currentUserID(ctx))
	if err != nil {
		return nil, err
	}
	return toProtoUser(u, s.enc), nil
}

func (s *UserServer) UpdateMe(ctx context.Context, req *apiv1.UpdateMeRequest) (*apiv1.User, error) {
	u, err := s.users.UpdateProfile(ctx, currentUserID(ctx), users.ProfileUpdate{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Bio:       req.Bio,
		Avatar:    req.Avatar,
	})
	if err != nil {
		return nil, err
	}
	return toProtoUser(u, s.enc), nil
}

func (s *UserServer) DeleteMe(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := s.users.DeleteAccount(ctx, currentUserID(ctx)); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *UserServer) GetUser(ctx context.Context, req *apiv1.GetUserRequest) (*apiv1.PublicUser, error) {
	id, err := s.enc.Decode(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "user.invalid_id")
	}
	u, err := s.users.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return toProtoPublicUser(u, s.enc), nil
}

func (s *UserServer) GetSettings(ctx context.Context, _ *emptypb.Empty) (*apiv1.Settings, error) {
	set, err := s.settings.Get(ctx, currentUserID(ctx))
	if err != nil {
		return nil, err
	}
	return toProtoSettings(set), nil
}

func (s *UserServer) UpdateSettings(ctx context.Context, req *apiv1.Settings) (*apiv1.Settings, error) {
	if err := s.settings.Update(ctx, currentUserID(ctx), fromProtoSettings(req)); err != nil {
		return nil, err
	}
	return req, nil
}

func (s *UserServer) RegisterPushToken(ctx context.Context, req *apiv1.RegisterPushTokenRequest) (*emptypb.Empty, error) {
	err := s.notifs.RegisterToken(ctx, currentUserID(ctx), notifications.RegisterTokenInput{
		DeviceID:   req.GetDeviceId(),
		Token:      req.GetToken(),
		Platform:   notifications.NotificationPlatform(req.GetPlatform()),
		AppVersion: req.GetAppVersion(),
	})
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *UserServer) GetNotificationPreferences(ctx context.Context, _ *emptypb.Empty) (*apiv1.NotificationPreferences, error) {
	prefs, err := s.notifs.GetPreferences(ctx, currentUserID(ctx))
	if err != nil {
		return nil, err
	}
	return toProtoNotifPrefs(prefs), nil
}

func (s *UserServer) UpdateNotificationPreferences(ctx context.Context, req *apiv1.NotificationPreferences) (*apiv1.NotificationPreferences, error) {
	push, email, marketing, qenabled := req.GetPush(), req.GetEmail(), req.GetMarketing(), req.GetQuietHoursEnabled()
	start, end, tz := req.GetQuietHoursStart(), req.GetQuietHoursEnd(), req.GetTimezone()
	prefs, err := s.notifs.UpdatePreferences(ctx, currentUserID(ctx), notifications.PreferencesUpdate{
		Push:              &push,
		Email:             &email,
		Marketing:         &marketing,
		QuietHoursEnabled: &qenabled,
		QuietHoursStart:   &start,
		QuietHoursEnd:     &end,
		Timezone:          &tz,
	})
	if err != nil {
		return nil, err
	}
	return toProtoNotifPrefs(prefs), nil
}
