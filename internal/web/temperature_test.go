package web_test

import (
	"net/http"
	"strings"
	"testing"
)

func TestTemperaturePageShowsForm(t *testing.T) {
	resp := get(t, "/temperature")
	html := body(t, resp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	for _, unit := range []string{"Celsius", "Fahrenheit", "Kelvin"} {
		if !strings.Contains(html, ">"+unit+"</option>") {
			t.Errorf("dropdowns do not list %q", unit)
		}
	}
}

func TestTemperatureConversions(t *testing.T) {
	// Expected values from the scale definitions: water freezes at 0 °C =
	// 32 °F and boils at 100 °C = 212 °F; absolute zero is 0 K = -273.15 °C
	// = -459.67 °F; -40 is the same in Celsius and Fahrenheit.
	tests := []struct {
		value, from, to string
		want            string
	}{
		{"100", "celsius", "fahrenheit", "100 °C = 212 °F"},
		{"32", "fahrenheit", "celsius", "32 °F = 0 °C"},
		{"-40", "celsius", "fahrenheit", "-40 °C = -40 °F"},
		{"-10", "celsius", "fahrenheit", "-10 °C = 14 °F"},
		{"0", "kelvin", "celsius", "0 K = -273.15 °C"},
		{"0", "kelvin", "fahrenheit", "0 K = -459.67 °F"},
		{"-273.15", "celsius", "kelvin", "-273.15 °C = 0 K"},
		{"-459.67", "fahrenheit", "kelvin", "-459.67 °F = 0 K"},
		{"300", "kelvin", "celsius", "300 K = 26.85 °C"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			resp := post(t, "/temperature", tt.value, tt.from, tt.to)
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

func TestTemperatureRejectsBelowAbsoluteZero(t *testing.T) {
	tests := []struct{ value, from string }{
		{"-300", "celsius"},
		{"-273.16", "celsius"},
		{"-459.68", "fahrenheit"},
		{"-1", "kelvin"},
		{"-1e-10", "kelvin"},
	}
	for _, tt := range tests {
		t.Run(tt.value+" "+tt.from, func(t *testing.T) {
			resp := post(t, "/temperature", tt.value, tt.from, "kelvin")
			text := shown(t, resp)

			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
			}
			if !strings.Contains(text, "below absolute zero") {
				t.Error("page does not say the value is below absolute zero")
			}
		})
	}
}
