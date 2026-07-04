-- Reverses 00013_create_notification_logs_table.up.sql. Drops the RLS
-- policies, validation trigger and its function, the partitioned table
-- (CASCADE removes the default and any child partitions), and the enum type.

DROP POLICY IF EXISTS notification_logs_owner_select ON notifications.notification_logs;
DROP POLICY IF EXISTS notification_logs_admin_all ON notifications.notification_logs;
DROP POLICY IF EXISTS notification_logs_ro_select ON notifications.notification_logs;
ALTER TABLE notifications.notification_logs DISABLE ROW LEVEL SECURITY;

DROP TABLE IF EXISTS notifications.notification_logs CASCADE;

DROP FUNCTION IF EXISTS validate_notification_log();

DROP TYPE IF EXISTS notifications.notification_status;
