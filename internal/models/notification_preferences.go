package models

import "time"

type NotificationPreferences struct {
	UserID            int64     `gorm:"primaryKey"             json:"user_id"             msgpack:"user_id"`
	Push              bool      `gorm:"default:true;not null"  json:"push"                msgpack:"push"`
	Email             bool      `gorm:"default:true;not null"  json:"email"               msgpack:"email"`
	Marketing         bool      `gorm:"default:false;not null" json:"marketing"           msgpack:"marketing"`
	QuietHoursEnabled bool      `gorm:"default:false;not null" json:"quiet_hours_enabled" msgpack:"quiet_hours_enabled"`
	QuietHoursStart   *string   `gorm:"type:time" json:"quiet_hours_start,omitempty" msgpack:"quiet_hours_start,omitempty"`
	QuietHoursEnd     *string   `gorm:"type:time" json:"quiet_hours_end,omitempty"   msgpack:"quiet_hours_end,omitempty"`
	Timezone          string    `gorm:"type:varchar(50);default:'UTC';not null" json:"timezone"   msgpack:"timezone"`
	CreatedAt         time.Time `gorm:"type:timestamptz;not null;default:now()" json:"created_at" msgpack:"created_at"`
	UpdatedAt         time.Time `gorm:"type:timestamptz;not null;default:now()" json:"updated_at" msgpack:"updated_at"`

	User *User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty" msgpack:"user,omitempty"`
}

func (NotificationPreferences) TableName() string {
	return "notifications.notification_preferences"
}
