-- Reverses 000047_create_notification_templates_table.up.sql by
-- dropping lifecycle triggers, indexes, the translations table, and
-- the templates table.

DROP TRIGGER IF EXISTS prevent_notification_template_translations_created_at_update ON notifications.notification_template_translations;

DROP TRIGGER IF EXISTS update_notification_template_translations_updated_at ON notifications.notification_template_translations;

DROP TRIGGER IF EXISTS prevent_notification_templates_created_at_update ON notifications.notification_templates;

DROP TRIGGER IF EXISTS update_notification_templates_updated_at ON notifications.notification_templates;

DROP INDEX IF EXISTS idx_notification_template_translations_language;

DROP INDEX IF EXISTS idx_notification_template_translations_template_id;

DROP INDEX IF EXISTS idx_notification_templates_priority;

DROP INDEX IF EXISTS idx_notification_templates_event_type;

-- Drop translations before templates (FK dependency).
DROP TABLE IF EXISTS notifications.notification_template_translations;

DROP TABLE IF EXISTS notifications.notification_templates;
