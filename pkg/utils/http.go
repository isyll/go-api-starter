package utils

import "net/http"

const (
	authRequiredWriteMsgKey    = "auth.required.write"
	authRequiredReadOnlyMsgKey = "auth.required.read_only"
)

// IsMutationMethod reports whether method is a state-changing HTTP verb
// (POST, PUT, PATCH, or DELETE).
func IsMutationMethod(method string) bool {
	return method == http.MethodPost ||
		method == http.MethodPut ||
		method == http.MethodPatch ||
		method == http.MethodDelete
}

// GetAuthMessageKey returns the i18n message key for authentication-required
// errors, distinguishing read from write requests.
func GetAuthMessageKey(method string) string {
	if IsMutationMethod(method) {
		return authRequiredWriteMsgKey
	}
	return authRequiredReadOnlyMsgKey
}
