package persistence

import (
	"fmt"

	"gorm.io/gorm"
)

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
