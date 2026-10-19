package web_test

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stonalampa/go-mastery-unit-converter/internal/web"
)

// post submits the form at path with the given value and units.
func post(t *testing.T, path, value, from, to string) *http.Response {
	t.Helper()
	form := url.Values{"value": {value}, "from": {from}, "to": {to}}
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	web.NewHandler().ServeHTTP(rec, req)
	return rec.Result()
}

func TestLengthConversions(t *testing.T) {
	// Expected values come from the unit definitions (1 in = 2.54 cm,
	// 1 ft = 12 in, 1 yd = 3 ft, 1 mi = 1609.344 m), rounded to 6 significant digits.
	tests := []struct {
		value, from, to string
		want            string
	}{
		{"1", "mile", "meter", "1 mi = 1609.34 m"},
		{"1", "kilometer", "mile", "1 km = 0.621371 mi"},
		{"12", "inch", "foot", "12 in = 1 ft"},
		{"1", "yard", "inch", "1 yd = 36 in"},
		{"2.5", "kilometer", "meter", "2.5 km = 2500 m"},
		{"1", "millimeter", "mile", "1 mm = 6.21371e-07 mi"},
		{"5", "meter", "meter", "5 m = 5 m"},
		{"0", "foot", "centimeter", "0 ft = 0 cm"},
		{"-0", "foot", "centimeter", "0 ft = 0 cm"},
		{"0.1", "inch", "inch", "0.1 in = 0.1 in"},
		{"1e308", "kilometer", "mile", "1e+308 km = 6.21371e+307 mi"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			resp := post(t, "/length", tt.value, tt.from, tt.to)
			html := body(t, resp)

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
			}
			if !strings.Contains(html, tt.want) {
				t.Errorf("page does not show %q", tt.want)
			}
		})
	}
}

func TestLengthFormKeepsInputAfterSubmit(t *testing.T) {
	html := body(t, post(t, "/length", "3.5", "yard", "foot"))

	if !strings.Contains(html, `name="value" value="3.5"`) {
		t.Error("submitted value 3.5 is not kept in the form")
	}
	if got := selected(t, html, "from"); got != "yard" {
		t.Errorf("from selected = %q, want yard", got)
	}
	if got := selected(t, html, "to"); got != "foot" {
		t.Errorf("to selected = %q, want foot", got)
	}
}

func TestLengthRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name, value, from, to string
		wantError             string
	}{
		{"empty value", "", "meter", "foot", `"" is not a number`},
		{"not a number", "abc", "meter", "foot", `"abc" is not a number`},
		{"comma decimal", "1,5", "meter", "foot", `"1,5" is not a number`},
		{"negative length", "-1", "meter", "foot", "value must not be negative"},
		{"not finite", "Inf", "meter", "foot", "value must be a finite number"},
		{"too large to represent", "1e999", "meter", "foot", "value must be a finite number"},
		{"result overflows", "1e308", "mile", "millimeter", "value must be a finite number"},
		{"unknown from unit", "1", "parsec", "foot", `unknown unit "parsec"`},
		{"unknown to unit", "1", "meter", "lightyear", `unknown unit "lightyear"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := post(t, "/length", tt.value, tt.from, tt.to)
			html := body(t, resp)

			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
			}
			if !strings.Contains(html, htmlEscape(tt.wantError)) {
				t.Errorf("page does not show error %q", tt.wantError)
			}
			if strings.Contains(html, `class="result"`) {
				t.Error("page shows a result for invalid input")
			}
			if !strings.Contains(html, `name="value" value="`+htmlEscape(tt.value)+`"`) {
				t.Errorf("value %q is not kept in the form", tt.value)
			}
		})
	}
}

// htmlEscape mirrors html/template's escaping of quotes in text and attributes.
func htmlEscape(s string) string {
	return template.HTMLEscapeString(s)
}
