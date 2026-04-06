package summary

import (
	"sort"
	"strings"
)

// ShortSubject extracts the first line of a commit message and trims it.
func ShortSubject(message string) string {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return "(no commit message)"
	}
	line := trimmed
	if idx := strings.IndexByte(trimmed, '\n'); idx >= 0 {
		line = strings.TrimSpace(trimmed[:idx])
	}
	if len(line) > 90 {
		return line[:87] + "..."
	}
	return line
}

// SanitizeAndSortBranches deduplicates and sorts branch names.
func SanitizeAndSortBranches(branches []string) []string {
	if len(branches) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(branches))
	out := make([]string, 0, len(branches))
	for _, branch := range branches {
		name := strings.TrimSpace(branch)
		if name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
