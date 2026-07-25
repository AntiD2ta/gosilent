package cli

import "runtime/debug"

// ResolveVersion determines the version string to report. It prefers the
// ldflags-injected value; when that is the default "dev" placeholder (e.g. a
// plain `go install`/`go build` with no version stamped), it falls back to the
// module version recorded in the binary's build info.
func ResolveVersion(injected string) string {
	return resolveVersion(injected, debug.ReadBuildInfo)
}

// resolveVersion determines the version string to report, preferring the
// ldflags-injected value and falling back to module build info.
func resolveVersion(injected string, readBuildInfo func() (*debug.BuildInfo, bool)) string {
	if injected != "dev" {
		return injected
	}
	if info, ok := readBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return injected
}
