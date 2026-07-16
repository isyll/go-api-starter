// Package version exposes build metadata stamped at link time.
package version

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func Version() string { return version }

func Commit() string { return commit }

func BuildDate() string { return date }
