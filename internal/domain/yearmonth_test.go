package domain

import "testing"

func TestParseYearMonth(t *testing.T) {
	for _, s := range []string{"07-2025", "2025-07"} {
		if _, err := ParseYearMonth(s); err != nil {
			t.Fatalf("parse %s: %v", s, err)
		}
	}
}
