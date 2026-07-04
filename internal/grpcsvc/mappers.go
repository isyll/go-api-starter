package grpcsvc

import (
	commonv1 "github.com/isyll/go-grpc-starter/gen/common/v1"
	userv1 "github.com/isyll/go-grpc-starter/gen/user/v1"
	"github.com/isyll/go-grpc-starter/internal/auth"
	"github.com/isyll/go-grpc-starter/internal/notifications"
	"github.com/isyll/go-grpc-starter/internal/settings"
	"github.com/isyll/go-grpc-starter/internal/users"
	"github.com/isyll/go-grpc-starter/pkg/idenc"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func toProtoUser(u *users.User, enc idenc.IDEncoder) *commonv1.User {
	return &commonv1.User{
		Id:            enc.Encode(u.ID),
		Email:         u.Email,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		Avatar:        u.Avatar,
		Bio:           u.Bio,
		Status:        string(u.Status),
		Role:          string(u.Role),
		EmailVerified: u.IsEmailVerified(),
		CreatedAt:     timestamppb.New(u.CreatedAt),
		UpdatedAt:     timestamppb.New(u.UpdatedAt),
	}
}

func toProtoPublicUser(u *users.User, enc idenc.IDEncoder) *commonv1.PublicUser {
	return &commonv1.PublicUser{
		Id:          enc.Encode(u.ID),
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		Avatar:      u.Avatar,
		Bio:         u.Bio,
		MemberSince: timestamppb.New(u.CreatedAt),
	}
}

func toProtoTokenPair(tp *auth.TokenPair, enc idenc.IDEncoder) *commonv1.TokenPair {
	return &commonv1.TokenPair{
		AccessToken:  tp.AccessToken,
		RefreshToken: tp.RefreshToken,
		ExpiresIn:    tp.ExpiresIn,
		User:         toProtoUser(tp.User, enc),
	}
}

func toProtoSettings(s *settings.Settings) *commonv1.Settings {
	return &commonv1.Settings{
		Locale:             s.Locale,
		Timezone:           s.Timezone,
		Theme:              string(s.Theme),
		EmailNotifications: s.EmailNotifications,
		PushNotifications:  s.PushNotifications,
		MarketingEmails:    s.MarketingEmails,
	}
}

func fromProtoSettings(s *commonv1.Settings) settings.Settings {
	return settings.Settings{
		Locale:             s.GetLocale(),
		Timezone:           s.GetTimezone(),
		Theme:              settings.Theme(s.GetTheme()),
		EmailNotifications: s.GetEmailNotifications(),
		PushNotifications:  s.GetPushNotifications(),
		MarketingEmails:    s.GetMarketingEmails(),
	}
}

func toProtoDevices(devices []auth.DeviceSessionInfo) []*commonv1.Device {
	out := make([]*commonv1.Device, len(devices))
	for i, d := range devices {
		out[i] = &commonv1.Device{
			DeviceId:     d.DeviceID,
			Name:         d.DeviceName,
			Platform:     d.Platform,
			Manufacturer: d.Manufacturer,
			Model:        d.Model,
			LastUsedAt:   timestamppb.New(d.LastUsedAt),
			Current:      d.Current,
		}
	}
	return out
}

func toProtoNotifPrefs(p *notifications.NotificationPreferences) *userv1.NotificationPreferences {
	out := &userv1.NotificationPreferences{
		Push:              p.Push,
		Email:             p.Email,
		Marketing:         p.Marketing,
		QuietHoursEnabled: p.QuietHoursEnabled,
		Timezone:          p.Timezone,
	}
	if p.QuietHoursStart != nil {
		out.QuietHoursStart = *p.QuietHoursStart
	}
	if p.QuietHoursEnd != nil {
		out.QuietHoursEnd = *p.QuietHoursEnd
	}
	return out
}
