package models

import "time"

// NotificationTemplate is the GORM model for
// notifications.notification_templates. One row per event type;
// titles and body copy live in Translations keyed by language code.
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
} // @name NotificationTemplate

// TableName returns the schema-qualified table name for GORM.
func (NotificationTemplate) TableName() string {
	return "notifications.notification_templates"
}

// NotificationTemplateTranslation holds the locale-specific title
// and body copy for one language of a NotificationTemplate.
type NotificationTemplateTranslation struct {
	ID         int       `gorm:"primaryKey;autoIncrement" json:"id"          msgpack:"id"`
	TemplateID int       `                                json:"template_id" msgpack:"template_id"`
	Language   string    `                                json:"language"    msgpack:"language"`
	Title      string    `                                json:"title"       msgpack:"title"`
	Body       string    `                                json:"body"        msgpack:"body"`
	CreatedAt  time.Time `                                json:"created_at"  msgpack:"created_at"`
	UpdatedAt  time.Time `                                json:"updated_at"  msgpack:"updated_at"`

	Template *NotificationTemplate `gorm:"foreignKey:TemplateID;references:ID" json:"template,omitempty" msgpack:"template,omitempty"`
} // @name NotificationTemplateTranslation

// TableName returns the schema-qualified table name for GORM.
func (NotificationTemplateTranslation) TableName() string {
	return "notifications.notification_template_translations"
}
