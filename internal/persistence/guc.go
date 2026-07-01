package persistence

import (
	"fmt"

	"gorm.io/gorm"
)

// SetChangeReason stashes reason in the transaction-local
// app.change_reason GUC that the log_*_change history triggers read
// into the *_status_history.reason column.
//
// History tables (auth.user_status_history, rides.trip_status_history,
// rides.booking_status_history, …) are populated EXCLUSIVELY by SQL
// AFTER UPDATE triggers — never from Go. The triggers recover the
// actor from app.current_user_id (set by the RLS callback) and the
// reason from app.change_reason. Call SetChangeReason on the SAME tx
// handle, immediately before the status-changing UPDATE; a nil or
// empty reason clears the GUC so the audit row's reason is NULL.
//
// It panics on failure: a failed set_config means the transaction's
// connection is unusable, which the recovery middleware turns into a
// 500 with the full stack — the same contract repositories use for
// unexpected infrastructure failures.
func SetChangeReason(tx *gorm.DB, reason *string) {
	reasonVal := ""
	if reason != nil {
		reasonVal = *reason
	}
	if err := tx.Exec(
		"SELECT set_config('app.change_reason', ?, TRUE)", reasonVal,
	).Error; err != nil {
		panic(fmt.Errorf("set app.change_reason GUC: %w", err))
	}
}
