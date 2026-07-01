// Package authjwt defines the claim shape of the short-lived
// Service-JWT (HS256, 15-minute TTL) issued by the Next.js admin
// app and validated by the Go API on every /system/admin/**
// request. Signing uses INTERNAL_JWT_SECRET.
//
// This package is admin-only. User-facing authentication uses
// Opaque Access Tokens (pkg/token); importing pkg/jwt from any
// user-facing path is forbidden.
package authjwt

import "github.com/golang-jwt/jwt/v5"

// ServiceJWTClaims are the claims embedded in the short-lived
// service JWT that the Next.js admin app signs and the Go
// backend validates on every request to /system/admin/**.
// The admin ID is carried in the standard Subject (sub) field
// of RegisteredClaims.
type ServiceJWTClaims struct {
	// AdminEmail is the admin's login email. Used for audit
	// logging only — authorization decisions read Permissions.
	AdminEmail string `json:"email"`
	// Permissions is the flat list of "resource:action" strings
	// granted by the admin's role bundles. Drives the
	// RequirePermission middleware.
	Permissions []string `json:"permissions"`
	// CountryCodes scopes the admin's reach to specific
	// ISO 3166-1 alpha-2 codes. Empty means global access.
	CountryCodes []string `json:"country_codes"`
	// ServiceSource identifies the calling service (e.g.
	// "admin-next"). Logged for traceability.
	ServiceSource string `json:"service_source"`
	jwt.RegisteredClaims
}
