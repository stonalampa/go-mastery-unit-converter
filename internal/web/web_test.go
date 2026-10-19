package web_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/stonalampa/go-mastery-unit-converter/internal/web"
)

// get sends a GET request for path to the app's handler and returns the response.
func get(t *testing.T, path string) *http.Response {
	t.Helper()
	rec := httptest.NewRecorder()
	web.NewHandler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec.Result()
}

func body(t *testing.T, resp *http.Response) string {
	t.Helper()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestRootRedirectsToLength(t *testing.T) {
	resp := get(t, "/")

	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusSeeOther)
	}
	if got := resp.Header.Get("Location"); got != "/length" {
		t.Errorf("Location = %q, want %q", got, "/length")
	}
}

func TestUnknownPathIsNotFound(t *testing.T) {
	if got := get(t, "/parsecs").StatusCode; got != http.StatusNotFound {
		t.Errorf("status = %d, want %d", got, http.StatusNotFound)
	}
}

func TestLengthPageShowsForm(t *testing.T) {
	resp := get(t, "/length")
	html := body(t, resp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
	for _, want := range []string{
		`<form method="post" action="/length">`,
		`<input type="number" step="any" name="value"`,
		`<select name="from"`,
		`<select name="to"`,
		`href="/length"`, `href="/weight"`, `href="/temperature"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("page does not contain %q", want)
		}
	}
	for _, unit := range []string{"Millimeter", "Centimeter", "Meter", "Kilometer", "Inch", "Foot", "Yard", "Mile"} {
		if !strings.Contains(html, ">"+unit+"</option>") {
			t.Errorf("dropdowns do not list %q", unit)
		}
	}
}

// selectedOption matches an <option> marked selected and captures its value.
var selectedOption = regexp.MustCompile(`<option value="([^"]*)" selected>`)

// selected returns the value of the selected option in the <select> named name.
func selected(t *testing.T, html, name string) string {
	t.Helper()
	start := strings.Index(html, `<select name="`+name+`"`)
	if start < 0 {
		t.Fatalf("no <select name=%q>", name)
	}
	sel, _, ok := strings.Cut(html[start:], "</select>")
	if !ok {
		t.Fatalf("<select name=%q> is not closed", name)
	}
	m := selectedOption.FindStringSubmatch(sel)
	if m == nil {
		return ""
	}
	return m[1]
}

func TestLengthFormDefaultsToFirstAndSecondUnit(t *testing.T) {
	html := body(t, get(t, "/length"))

	if got := selected(t, html, "from"); got != "millimeter" {
		t.Errorf("from selected = %q, want millimeter", got)
	}
	if got := selected(t, html, "to"); got != "centimeter" {
		t.Errorf("to selected = %q, want centimeter", got)
	}
}
