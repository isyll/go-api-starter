package codes

var (
	UserNotFound = register("USER_NOT_FOUND")

	InvalidUserID = register("INVALID_USER_ID")

	AccountSuspended = register("ACCOUNT_SUSPENDED")

	CannotReactivateSuspended = register("CANNOT_REACTIVATE_SUSPENDED")

	AccountNotSuspended = register("ACCOUNT_NOT_SUSPENDED")

	InvalidRoleTransition = register("INVALID_ROLE_TRANSITION")

	CannotDowngradeRole = register("CANNOT_DOWNGRADE_ROLE")

	CannotSwitchSingleRole = register("CANNOT_SWITCH_SINGLE_ROLE")

	InvalidCountryCode = register("INVALID_COUNTRY_CODE")

	NonAvailableCountryCode = register("NON_AVAILABLE_COUNTRY_CODE")

	InvalidTimezone = register("INVALID_TIMEZONE")

	InvalidTimezoneForCountry = register("INVALID_TIMEZONE_FOR_COUNTRY")
)
