package db

import (
	"context"
	"strconv"

	"gorm.io/gorm"

	"github.com/isyll/go-api-starter/internal/authz"
)

// rlsBypassKey, when present in a Statement's context, signals that
// the callback chain is currently emitting the set_config statement
// itself and should not recurse.
type rlsBypassKey struct{}

// registerRLSCallback installs the GORM Before-callbacks that
// surface authz.Subject to PostgreSQL via two transaction-local
// session variables:
//
//   - app.current_user_id : the BIGINT user id used in the
//     "<owner_col> = current_setting(...)" predicates of every
//     RLS policy.
//   - app.current_role    : the textual role (driver / passenger /
//     both / admin / anonymous) - reserved for future role-aware
//     policies; today it is informational and read by tests.
//
// Behavior by role:
//
//   - RoleMigration : no callback installed. The migration runner
//     owns the tables and must run DDL without RLS context.
//   - RoleAdmin     : callback runs, but every RLS policy contains
//     a FOR ALL TO app_worker USING (true) clause so the variable
//     does not gate access.
//   - RoleApp       : callback runs every operation. A request
//     without authz.Subject lands with current_user_id = 0, which
//     RLS interprets as "no rows match" - safe default.
//
// set_config(..., true) is transaction-local. GORM wraps every
// non-transactional statement in an implicit transaction, and the
// callback fires on the same connection as the original statement,
// so the SET always lands in the right place.
//
// Recursion is bounded: the set_config call carries a sentinel
// context value (rlsBypassKey) so its own pass through the Raw
// callback is a no-op.
func registerRLSCallback(db *gorm.DB, role Role) {
	if role == RoleMigration {
		return
	}

	apply := func(tx *gorm.DB) {
		if tx.Statement == nil || tx.Statement.Context == nil {
			return
		}
		if tx.Statement.Context.Value(rlsBypassKey{}) != nil {
			return
		}
		s := authz.From(tx.Statement.Context)
		userID := "0"
		roleStr := string(authz.RoleAnonymous)
		if s.IsAdmin {
			roleStr = string(authz.RoleAdmin)
		} else if s.UserID > 0 {
			userID = strconv.FormatInt(s.UserID, 10)
			if s.Role != "" {
				roleStr = string(s.Role)
			} else {
				roleStr = string(authz.RoleAnonymous)
			}
		}
		bypassCtx := context.WithValue(
			tx.Statement.Context, rlsBypassKey{}, true,
		)
		_ = tx.Session(&gorm.Session{NewDB: true}).
			WithContext(bypassCtx).
			Exec(
				`SELECT set_config('app.current_user_id', ?, true),
				        set_config('app.current_role',    ?, true)`,
				userID, roleStr,
			).Error
	}

	cb := db.Callback()
	_ = cb.Query().Before("gorm:query").Register(
		"app_owner:rls_set_session_vars_query", apply,
	)
	_ = cb.Create().Before("gorm:create").Register(
		"app_owner:rls_set_session_vars_create", apply,
	)
	_ = cb.Update().Before("gorm:update").Register(
		"app_owner:rls_set_session_vars_update", apply,
	)
	_ = cb.Delete().Before("gorm:delete").Register(
		"app_owner:rls_set_session_vars_delete", apply,
	)
	_ = cb.Raw().Before("gorm:raw").Register(
		"app_owner:rls_set_session_vars_raw", apply,
	)
}
