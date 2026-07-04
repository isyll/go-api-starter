package codes

var (
	InternalError = register("INTERNAL_ERROR")

	InvalidPayload = register("INVALID_PAYLOAD")

	InvalidQuery = register("INVALID_QUERY")

	InvalidID = register("INVALID_ID")

	InvalidParam = register("INVALID_PARAM")

	ValidationError = register("VALIDATION_ERROR")

	Forbidden = register("FORBIDDEN")

	NotFound = register("NOT_FOUND")

	NoFieldsToUpdate = register("NO_FIELDS_TO_UPDATE")

	NoChanges = register("NO_CHANGES")

	InvalidDateRange = register("INVALID_DATE_RANGE")

	DateInPast = register("DATE_IN_PAST")

	MethodNotAllowed = register("METHOD_NOT_ALLOWED")

	RouteNotFound = register("ROUTE_NOT_FOUND")

	StorageUnavailable = register("STORAGE_UNAVAILABLE")

	UploadFailed = register("UPLOAD_FAILED")
)
