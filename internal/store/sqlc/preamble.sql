-- sqlc-only catalog preamble. Not a migration and never executed against
-- a real database. It declares the objects that infra/postgres/init
-- provisions out-of-band (schemas, citext) plus a stub for pg_partman's
-- create_parent, so sqlc can build a complete catalog from migrations/.

CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS notifications;
CREATE SCHEMA IF NOT EXISTS audit;
CREATE SCHEMA IF NOT EXISTS events;

CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Enums that sqlc's DO-block parser does not pick up from the migrations.
-- Kept identical to the CREATE TYPE statements in migrations/.
CREATE TYPE auth.user_status AS ENUM ('active', 'inactive', 'suspended');
CREATE TYPE auth.user_role AS ENUM ('user', 'admin');
CREATE TYPE auth.change_actor AS ENUM ('admin', 'system');
CREATE TYPE auth.suspension_reason AS ENUM (
    'terms_violation', 'fraudulent_activity', 'harassment', 'spam',
    'security_breach', 'legal_request', 'other'
);
CREATE TYPE auth.notification_platform AS ENUM ('android', 'ios', 'web');
CREATE TYPE notifications.notification_status AS ENUM (
    'sent', 'failed', 'clicked', 'dismissed'
);

CREATE FUNCTION public.create_parent(
    p_parent_table text,
    p_control text,
    p_interval text,
    p_type text DEFAULT 'range',
    p_premake integer DEFAULT 4
) RETURNS boolean AS $$ SELECT true $$ LANGUAGE sql;
