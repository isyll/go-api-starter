// Package authjwt defines the claims for short-lived admin tokens.
package authjwt

import "github.com/golang-jwt/jwt/v5"

type ServiceJWTClaims struct {
	AdminEmail    string   `json:"email"`
	Permissions   []string `json:"permissions"`
	CountryCodes  []string `json:"country_codes"`
	ServiceSource string   `json:"service_source"`
	jwt.RegisteredClaims
}
