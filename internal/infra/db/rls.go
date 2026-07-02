package db

import (
	"context"
	"strconv"

	"gorm.io/gorm"

	"github.com/isyll/go-api-starter/internal/authz"
)

type rlsBypassKey struct{}

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
		"db:rls_set_session_vars_query", apply,
	)
	_ = cb.Create().Before("gorm:create").Register(
		"db:rls_set_session_vars_create", apply,
	)
	_ = cb.Update().Before("gorm:update").Register(
		"db:rls_set_session_vars_update", apply,
	)
	_ = cb.Delete().Before("gorm:delete").Register(
		"db:rls_set_session_vars_delete", apply,
	)
	_ = cb.Raw().Before("gorm:raw").Register(
		"db:rls_set_session_vars_raw", apply,
	)
}
