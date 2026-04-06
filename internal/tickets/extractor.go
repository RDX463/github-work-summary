package tickets

import (
	"regexp"
	"strings"
)

var (
	// Standard ticket pattern: PROJECT-123
	idRegex = regexp.MustCompile(`([A-Z][A-Z0-9]+-[0-9]+)`)
)

// ExtractTicketIDs finds all unique ticket IDs in a block of text.
func ExtractTicketIDs(text string) []string {
	matches := idRegex.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[string]struct{})
	var ids []string
	for _, match := range matches {
		id := strings.ToUpper(match[1])
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}
	return ids
}
