// Package version holds the version of the build. The Docker build sets it:
// go build -ldflags "-X .../platform/version.Version=<git describe>".
package version

// Version is "dev" in a build without the flag, for example "go run".
var Version = "dev"
