package version

import (
	"runtime/debug"
	"strings"
)

const Repo = "RDX463/github-work-summary"

var (
	// Version can be injected at build time:
	//   -ldflags "-X github.com/RDX463/github-work-summary/internal/version.Version=v0.1.1"
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// Current returns the best-available version string for the running binary.
func Current() string {
	v := strings.TrimSpace(Version)
	if v != "" && v != "dev" {
		return v
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	if info.Main.Version == "" || info.Main.Version == "(devel)" {
		return "dev"
	}
	return info.Main.Version
}
