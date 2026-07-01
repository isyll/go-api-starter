package codes

// Settings-domain codes. /my/settings PATCH and the per-key
// validators that gate it.

var (
	// SettingsNotFound — 404. No settings row for the user (row is normally seeded).
	SettingsNotFound = register("SETTINGS_NOT_FOUND")

	// InvalidPhoneVisibility — 400. phone_visibility not in the allowed enum.
	InvalidPhoneVisibility = register("INVALID_PHONE_VISIBILITY")

	// InvalidArchivePolicy — 400. trip_auto_archive_policy not in the allowed enum.
	InvalidArchivePolicy = register("INVALID_ARCHIVE_POLICY")

	// InvalidDataSharingPolicy — 400. data_sharing_with_third_parties out of enum.
	InvalidDataSharingPolicy = register("INVALID_DATA_SHARING_POLICY")

	// InvalidSettingKey — 400. PATCH references an undefined settings key.
	InvalidSettingKey = register("INVALID_SETTING_KEY")

	// InvalidSettingValue — 400. PATCH value fails the per-key validator.
	InvalidSettingValue = register("INVALID_SETTING_VALUE")

	// AutoAcceptNotAllowed — 403. Auto-accept toggled without the required verification.
	AutoAcceptNotAllowed = register("AUTO_ACCEPT_NOT_ALLOWED")
)
