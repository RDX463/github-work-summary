package version

const (
	// Version is the current version of the github-work-summary CLI.
	Version = "v1.5.0"

	// Repo is the source repository for version checks.
	Repo = "RDX463/github-work-summary"
)

// Current returns the current project version.
func Current() string {
	return Version
}
