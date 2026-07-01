package models

import "time"

// UserStatusHistory is the GORM model for auth.user_status_history.
// Immutable audit log written by the DB trigger on every status
// change to auth.users.
type UserStatusHistory struct {
	ID          int64      `gorm:"primaryKey"                       json:"id"                        msgpack:"id"`
	UserID      int64      `gorm:"not null;index"                   json:"user_id"                   msgpack:"user_id"`
	User        *User      `gorm:"foreignKey:UserID"                json:"user,omitempty"            msgpack:"user,omitempty"`
	Status      UserStatus `gorm:"column:new_status;not null;index" json:"status"                    msgpack:"status"`
	OldStatus   UserStatus `                                        json:"old_status"                msgpack:"old_status"`
	ChangedByID *int64     `gorm:"column:changed_by"                json:"changed_by_id"             msgpack:"changed_by_id"`
	ChangedBy   *User      `gorm:"foreignKey:ChangedByID"           json:"changed_by_user,omitempty" msgpack:"changed_by_user,omitempty"`
	Reason      *string    `                                        json:"reason,omitempty"          msgpack:"reason,omitempty"`
	CreatedAt   time.Time  `gorm:"not null;default:now()"           json:"created_at"                msgpack:"created_at"`
}

// TableName returns the schema-qualified table name for GORM.
func (ush *UserStatusHistory) TableName() string {
	return "auth.user_status_history"
}
