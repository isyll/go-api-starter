-- Runs once on first container start (as the superuser) to provision
-- extensions, schemas, roles, and grants. Dev passwords match
-- .env.example; override them in any real environment.

CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_owner') THEN
        CREATE ROLE app_owner LOGIN PASSWORD 'app_owner_password';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_api') THEN
        CREATE ROLE app_api LOGIN PASSWORD 'app_api_password';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_worker') THEN
        CREATE ROLE app_worker LOGIN PASSWORD 'app_worker_password';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_readonly') THEN
        CREATE ROLE app_readonly LOGIN PASSWORD 'app_readonly_password';
    END IF;
END $$;

CREATE SCHEMA IF NOT EXISTS auth AUTHORIZATION app_owner;
CREATE SCHEMA IF NOT EXISTS notifications AUTHORIZATION app_owner;
CREATE SCHEMA IF NOT EXISTS audit AUTHORIZATION app_owner;
CREATE SCHEMA IF NOT EXISTS events AUTHORIZATION app_owner;

-- Migrations run as app_owner and create shared trigger functions and the
-- migration-history table in the public schema, which PostgreSQL 15+ no longer
-- grants to non-owners by default.
GRANT CREATE ON SCHEMA public TO app_owner;

DO $$
DECLARE s TEXT;
BEGIN
    FOREACH s IN ARRAY ARRAY['public','auth','notifications','audit','events'] LOOP
        EXECUTE format('GRANT USAGE ON SCHEMA %I TO app_api, app_worker, app_readonly', s);
        EXECUTE format('ALTER DEFAULT PRIVILEGES FOR ROLE app_owner IN SCHEMA %I GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_api, app_worker', s);
        EXECUTE format('ALTER DEFAULT PRIVILEGES FOR ROLE app_owner IN SCHEMA %I GRANT SELECT ON TABLES TO app_readonly', s);
        EXECUTE format('ALTER DEFAULT PRIVILEGES FOR ROLE app_owner IN SCHEMA %I GRANT USAGE, SELECT ON SEQUENCES TO app_api, app_worker', s);
    END LOOP;
END $$;
