// Package version exposes build metadata stamped at link time:
//
//	go build -ldflags "-X github.com/isyll/go-grpc-starter/pkg/version.version=v1.2.3 ..."
//
// The justfile build recipe and the Dockerfile stamp these automatically.
package version

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Version is the semantic version or git describe output.
func Version() string { return version }

// Commit is the short git commit hash.
func Commit() string { return commit }

// BuildDate is the UTC build timestamp.
func BuildDate() string { return date }
