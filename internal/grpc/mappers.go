package grpc

import (
	"github.com/isyll/go-grpc-starter/internal/domain/auth"
	"github.com/isyll/go-grpc-starter/internal/domain/notifications"
	"github.com/isyll/go-grpc-starter/internal/domain/settings"
	"github.com/isyll/go-grpc-starter/internal/domain/users"
	apiv1 "github.com/isyll/go-grpc-starter/internal/gen/api/v1"
	"github.com/isyll/go-grpc-starter/pkg/idenc"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func toProtoUser(u *users.User, enc idenc.IDEncoder) *apiv1.User {
	return &apiv1.User{
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

func toProtoPublicUser(u *users.User, enc idenc.IDEncoder) *apiv1.PublicUser {
	return &apiv1.PublicUser{
		Id:          enc.Encode(u.ID),
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		Avatar:      u.Avatar,
		Bio:         u.Bio,
		MemberSince: timestamppb.New(u.CreatedAt),
	}
}

func toProtoTokenPair(tp *auth.TokenPair, enc idenc.IDEncoder) *apiv1.TokenPair {
	return &apiv1.TokenPair{
		AccessToken:  tp.AccessToken,
		RefreshToken: tp.RefreshToken,
		ExpiresIn:    tp.ExpiresIn,
		User:         toProtoUser(tp.User, enc),
	}
}

func toProtoSettings(s *settings.Settings) *apiv1.Settings {
	return &apiv1.Settings{
		Locale:             s.Locale,
		Timezone:           s.Timezone,
		Theme:              string(s.Theme),
		EmailNotifications: s.EmailNotifications,
		PushNotifications:  s.PushNotifications,
		MarketingEmails:    s.MarketingEmails,
	}
}

func fromProtoSettings(s *apiv1.Settings) settings.Settings {
	return settings.Settings{
		Locale:             s.GetLocale(),
		Timezone:           s.GetTimezone(),
		Theme:              settings.Theme(s.GetTheme()),
		EmailNotifications: s.GetEmailNotifications(),
		PushNotifications:  s.GetPushNotifications(),
		MarketingEmails:    s.GetMarketingEmails(),
	}
}

func toProtoDevices(devices []auth.DeviceSessionInfo) []*apiv1.Device {
	out := make([]*apiv1.Device, len(devices))
	for i, d := range devices {
		out[i] = &apiv1.Device{
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

func toProtoNotifPrefs(p *notifications.NotificationPreferences) *apiv1.NotificationPreferences {
	out := &apiv1.NotificationPreferences{
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
