// Package idenc obfuscates internal int64 database IDs as opaque
// Sqids strings at the API boundary. Carved out of pkg/utils so
// the package name names exactly what it owns. Consumers that
// only need to encode/decode IDs no longer pull in the legacy
// utils grab-bag.
//
// The package exports:
//   - IDEncoder (the encode/decode contract)
//   - NewSqidsEncoder (Sqids-backed implementation)
//   - SqID (int64-shaped type that automatically decodes from URI /
//     query parameters via encoding.TextUnmarshaler)
//   - SetGlobalEncoder (singleton hook used by SqID's
//     UnmarshalText; called once at boot)
package idenc
