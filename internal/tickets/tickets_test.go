package tickets

import (
	"testing"
)

func TestExtractTicketIDs(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected []string
	}{
		{"Single Jira ID", "FIX-123: resolve login bug", []string{"FIX-123"}},
		{"Multiple Jira IDs", "FIX-123 and PROJ-456: fix all bugs", []string{"FIX-123", "PROJ-456"}},
		{"No IDs", "resolve login bug", nil},
		{"Linear-style ID", "ENG-123: add feature", []string{"ENG-123"}},
		{"Duplicate IDs", "FIX-123 FIX-123", []string{"FIX-123"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractTicketIDs(tt.message)
			if len(got) != len(tt.expected) {
				t.Errorf("ExtractTicketIDs() got %v, want %v", got, tt.expected)
				return
			}
			for i, v := range got {
				if v != tt.expected[i] {
					t.Errorf("ExtractTicketIDs() got %v, want %v", got, tt.expected)
				}
			}
		})
	}
}
