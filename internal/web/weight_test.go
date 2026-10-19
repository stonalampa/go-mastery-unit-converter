package web_test

import (
	"net/http"
	"strings"
	"testing"
)

func TestWeightPageShowsForm(t *testing.T) {
	resp := get(t, "/weight")
	html := body(t, resp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if !strings.Contains(html, `<form method="post" action="/weight">`) {
		t.Error("form does not post to /weight")
	}
	for _, unit := range []string{"Milligram", "Gram", "Kilogram", "Ounce", "Pound"} {
		if !strings.Contains(html, ">"+unit+"</option>") {
			t.Errorf("dropdowns do not list %q", unit)
		}
	}
	if got := selected(t, html, "from"); got != "milligram" {
		t.Errorf("from selected = %q, want milligram", got)
	}
}

func TestWeightConversions(t *testing.T) {
	// Expected values come from the international definitions
	// (1 lb = 453.59237 g, 1 oz = 1/16 lb), rounded to 6 significant digits.
	tests := []struct {
		value, from, to string
		want            string
	}{
		{"1", "pound", "gram", "1 lb = 453.592 g"},
		{"1", "kilogram", "ounce", "1 kg = 35.274 oz"},
		{"16", "ounce", "pound", "16 oz = 1 lb"},
		{"1", "gram", "milligram", "1 g = 1000 mg"},
		{"2.5", "kilogram", "pound", "2.5 kg = 5.51156 lb"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			resp := post(t, "/weight", tt.value, tt.from, tt.to)
			text := shown(t, resp)

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status = %d, want %d; page:\n%s", resp.StatusCode, http.StatusOK, text)
			}
			if !strings.Contains(text, tt.want) {
				t.Errorf("page does not show %q", tt.want)
			}
		})
	}
}

func TestWeightRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name, value, from, to string
		wantError             string
	}{
		{"negative weight", "-2", "kilogram", "pound", "value must not be negative"},
		{"not a number", "heavy", "kilogram", "pound", `"heavy" is not a number`},
		{"length unit on weight page", "1", "meter", "pound", `unknown unit "meter"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := post(t, "/weight", tt.value, tt.from, tt.to)
			text := shown(t, resp)

			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
			}
			if !strings.Contains(text, tt.wantError) {
				t.Errorf("page does not show error %q", tt.wantError)
			}
		})
	}
}
