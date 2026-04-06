package schedule

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected HourMinuteDay
		wantErr  bool
	}{
		{"Valid HH:MM", "09:00", HourMinuteDay{Hour: 9, Minute: 0, Day: -1}, false},
		{"Valid Weekday HH:MM", "Monday 10:30", HourMinuteDay{Hour: 10, Minute: 30, Day: 1}, false},
		{"Invalid format", "morning", HourMinuteDay{}, true},
		{"Invalid hour", "25:00", HourMinuteDay{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Hour != tt.expected.Hour || got.Minute != tt.expected.Minute || int(got.Day) != int(tt.expected.Day) {
					t.Errorf("Parse() got = %v, want %v", got, tt.expected)
				}
			}
		})
	}
}

type HourMinuteDay struct {
	Hour   int
	Minute int
	Day    int
}

func TestGenerateLaunchAgent(t *testing.T) {
	config := LaunchAgentConfig{
		Label:          "com.user.gws",
		ExecutablePath: "/usr/local/bin/gws",
		Hour:           10,
		Minute:         30,
		Day:            1, // Monday
	}

	plist, err := RenderLaunchAgent(config)
	if err != nil {
		t.Fatalf("RenderLaunchAgent() error = %v", err)
	}

	// Verify key elements in the XML
	expectedStrings := []string{
		"<key>Label</key>",
		"<string>com.user.gws</string>",
		"<key>Hour</key>",
		"<integer>10</integer>",
		"<key>Minute</key>",
		"<integer>30</integer>",
		"<key>Day</key>",
		"<integer>1</integer>",
	}

	for _, s := range expectedStrings {
		if !strings.Contains(plist, s) {
			t.Errorf("Plist missing expected string: %s", s)
		}
	}
}
