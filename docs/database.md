# Database and migrations

PostgreSQL accessed through pgx + sqlc (see `internal/store`), organized
into schemas: `auth`, `notifications`, `audit`, `events`. Queries are
written in `internal/store/query/*.sql` and compiled to type-safe Go in
`gen/db` with `just sqlc`.

## Roles

| Role | Used by | Access |
| ---- | ------- | ------ |
| `app_owner` | migrations | owns schemas and tables (DDL) |
| `app_api` | the gRPC server | read/write app tables |
| `app_worker` | workers, outbox drain | read/write, incl. audit |
| `app_readonly` | dashboards | select only |

Roles, schemas, and extensions are provisioned by
`infra/postgres/init/00-init.sql` on first container start.

## Row-level security

Tables enable `FORCE ROW LEVEL SECURITY`. The app role has full access
(the API always scopes queries to the caller); the read-only role is
select-only. Every store transaction sets `app.current_user_id`,
`app.current_role`, and `app.change_reason` GUCs (via `SET LOCAL`), which
the schema's audit triggers and RLS policies read, so you can tighten a
policy to owner-only on any table you add.

## Migrations

golang-migrate with paired `NNNNNN_name.up.sql` / `.down.sql` files.

```sh
just migrate            # up
just migrate-down 1     # roll back one
just migrate-create x   # scaffold a new pair
just migrate-status     # current version
```

Every run is recorded in `public.migration_history`.
