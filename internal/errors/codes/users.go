package codes

// Users-domain codes. USER_ prefix is kept on the not-found
// code (peers with TripNotFound / BookingNotFound); other codes
// are unique in the registry without a prefix.

var (
	// UserNotFound — 404. FindByID / FindByAccountID / FindByPhone miss.
	UserNotFound = register("USER_NOT_FOUND")

	// InvalidUserID — 400. Obfuscated user_id path parameter failed to decode.
	InvalidUserID = register("INVALID_USER_ID")

	// AccountSuspended — 403. Suspended principal attempting blocked action.
	AccountSuspended = register("ACCOUNT_SUSPENDED")

	// CannotReactivateSuspended — 400. Self-reactivation forbidden; admin only.
	CannotReactivateSuspended = register("CANNOT_REACTIVATE_SUSPENDED")

	// AccountNotSuspended — 404. Suspension lookup on a non-suspended account.
	AccountNotSuspended = register("ACCOUNT_NOT_SUSPENDED")

	// InvalidRoleTransition — 400. Role switch not allowed by the role-state matrix.
	InvalidRoleTransition = register("INVALID_ROLE_TRANSITION")

	// CannotDowngradeRole — 400. role=both -> single requires admin.
	CannotDowngradeRole = register("CANNOT_DOWNGRADE_ROLE")

	// CannotSwitchSingleRole — 400. Single -> other single requires intermediate 'both'.
	CannotSwitchSingleRole = register("CANNOT_SWITCH_SINGLE_ROLE")

	// InvalidCountryCode — 400. Country code not in the ISO 3166-1 catalog.
	InvalidCountryCode = register("INVALID_COUNTRY_CODE")

	// NonAvailableCountryCode — 400. Country exists but not operationally supported.
	NonAvailableCountryCode = register("NON_AVAILABLE_COUNTRY_CODE")

	// InvalidTimezone — 400. IANA timezone string not recognized.
	InvalidTimezone = register("INVALID_TIMEZONE")

	// InvalidTimezoneForCountry — 400. Timezone valid but inconsistent with country.
	InvalidTimezoneForCountry = register("INVALID_TIMEZONE_FOR_COUNTRY")
)
