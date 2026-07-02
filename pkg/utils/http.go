package utils

import "net/http"

const (
	authRequiredWriteMsgKey    = "auth.required.write"
	authRequiredReadOnlyMsgKey = "auth.required.read_only"
)

func IsMutationMethod(method string) bool {
	return method == http.MethodPost ||
		method == http.MethodPut ||
		method == http.MethodPatch ||
		method == http.MethodDelete
}

func GetAuthMessageKey(method string) string {
	if IsMutationMethod(method) {
		return authRequiredWriteMsgKey
	}
	return authRequiredReadOnlyMsgKey
}
