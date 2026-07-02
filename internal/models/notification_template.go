package models

import "time"

type NotificationTemplate struct {
	ID               int       `gorm:"primaryKey;autoIncrement" json:"id"                           msgpack:"id"`
	EventType        string    `                                json:"event_type"                   msgpack:"event_type"`
	Icon             *string   `                                json:"icon,omitempty"               msgpack:"icon,omitempty"`
	Sound            string    `                                json:"sound"                        msgpack:"sound"`
	Priority         string    `                                json:"priority"                     msgpack:"priority"`
	AndroidChannelID *string   `                                json:"android_channel_id,omitempty" msgpack:"android_channel_id,omitempty"`
	Action           *string   `                                json:"action,omitempty"             msgpack:"action,omitempty"`
	CreatedAt        time.Time `                                json:"created_at"                   msgpack:"created_at"`
	UpdatedAt        time.Time `                                json:"updated_at"                   msgpack:"updated_at"`

	Translations []*NotificationTemplateTranslation `gorm:"foreignKey:TemplateID" json:"translations,omitempty" msgpack:"translations,omitempty"`
}

func (NotificationTemplate) TableName() string {
	return "notifications.notification_templates"
}

type NotificationTemplateTranslation struct {
	ID         int       `gorm:"primaryKey;autoIncrement" json:"id"          msgpack:"id"`
	TemplateID int       `                                json:"template_id" msgpack:"template_id"`
	Language   string    `                                json:"language"    msgpack:"language"`
	Title      string    `                                json:"title"       msgpack:"title"`
	Body       string    `                                json:"body"        msgpack:"body"`
	CreatedAt  time.Time `                                json:"created_at"  msgpack:"created_at"`
	UpdatedAt  time.Time `                                json:"updated_at"  msgpack:"updated_at"`

	Template *NotificationTemplate `gorm:"foreignKey:TemplateID;references:ID" json:"template,omitempty" msgpack:"template,omitempty"`
}

func (NotificationTemplateTranslation) TableName() string {
	return "notifications.notification_template_translations"
}
