// Package version holds the gherkinator version string. The value is
// injected at build time via -ldflags; unbuilt or development builds
// report "dev".
package version

// Version is the gherkinator version. It is set at build time via
// -ldflags "-X github.com/canonical/gherkinator/internal/version.Version=<tag>".
// Unbuilt or development builds report "dev".
var Version = "dev"
