// Package models contains the GORM models for every database-backed
// entity in the App API. Each file in this package corresponds to
// one or more closely related tables. Models are shared across all
// domain packages; they must never import from internal/domain.
//
// Monetary amounts are always stored as int64 minor units (e.g.
// CFA centimes). All PostGIS geometry fields use SRID 4326. Soft-
// deleted rows carry a non-nil DeletedAt; repositories must filter
// deleted_at IS NULL explicitly or via a GORM scope.
package models
