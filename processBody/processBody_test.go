package processBody

import (
	"fmt"
	"testing"
	"time"
)

func TestParseHumanDate_Valid(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"2025-06-19", time.Date(2025, 6, 19, 0, 0, 0, 0, time.UTC).Unix()},
		{"2025-06-19 13:47:35", time.Date(2025, 6, 19, 13, 47, 35, 0, time.UTC).Unix()},
		{"2025-06-19 13:47", time.Date(2025, 6, 19, 13, 47, 0, 0, time.UTC).Unix()},
	}
	for _, test := range tests {
		ts, err := parseHumanDate(test.input)
		if err != nil {
			t.Errorf("parseHumanDate(%q) returned error: %v", test.input, err)
		}
		if ts != test.expected {
			t.Errorf("parseHumanDate(%q) = %d, want %d", test.input, ts, test.expected)
		}
	}
}

func TestParseHumanDate_Invalid(t *testing.T) {
	invalids := []string{
		"2025/06/19", // Slashes are not handled by parseHumanDate
		"2025/06/19 13:47:35",
		"19-06-2025",
		"2025-13-19",
		"2025-06-32",
		"notadate",
	}
	for _, input := range invalids {
		_, err := parseHumanDate(input)
		if err == nil {
			t.Errorf("parseHumanDate(%q) should have failed, but did not", input)
		}
	}
}

func TestProcessFilter(t *testing.T) {
	base := "date > \"2025-06-19 13:47:35\""
	ts := time.Date(2025, 6, 19, 13, 47, 35, 0, time.UTC).Unix()
	expected := fmt.Sprintf("date > %d", ts)
	result, err := processFilter(base)
	if err != nil {
		t.Fatalf("processFilter error: %v", err)
	}
	if result != expected {
		t.Errorf("processFilter = %q, want %q", result, expected)
	}

	// Slashed format should now be converted by processFilter
	slashed := "date > \"2025/06/19 13:47:35\""
	result, err = processFilter(slashed)
	if err != nil {
		t.Errorf("processFilter should not error for slashed date format, got: %v", err)
	}
	if result != expected {
		t.Errorf("processFilter for slashed date = %q, want %q", result, expected)
	}
}
