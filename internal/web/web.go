// Package web serves the unit converter's pages.
package web

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/stonalampa/go-mastery-unit-converter/internal/convert"
)

//go:embed templates/*.html
var templateFS embed.FS

var templates = template.Must(template.ParseFS(templateFS, "templates/*.html"))

// navLink is one entry of the navigation bar.
type navLink struct {
	Path, Name string
}

var nav = []navLink{
	{"/length", "Length"},
	{"/weight", "Weight"},
	{"/temperature", "Temperature"},
}

// categories are the Categories served, one page each.
var categories = []convert.Category{convert.Length, convert.Weight, convert.Temperature}

// pageData is everything the page template renders.
type pageData struct {
	Nav      []navLink
	Category convert.Category
	From, To string // selected Unit IDs
	Value    string // the value as typed, kept in the form
	Result   string // e.g. "1 mi = 1609.34 m"
	Error    string // shown instead of a Result when the input is invalid
}

// NewHandler returns the app's HTTP handler with all routes registered.
func NewHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/length", http.StatusSeeOther)
	})
	for _, c := range categories {
		// Each iteration has its own c (Go 1.22+), so every closure below
		// keeps the Category it was registered for.
		mux.HandleFunc("GET /"+c.ID, func(w http.ResponseWriter, _ *http.Request) {
			render(w, http.StatusOK, newPage(c))
		})
		mux.HandleFunc("POST /"+c.ID, convertHandler(c))
	}
	return mux
}

// convertHandler handles a submitted form: it converts the value and renders
// the form again with the Result, or with an Error and status 400.
func convertHandler(c convert.Category) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := newPage(c)
		data.Value = r.FormValue("value")
		data.From = r.FormValue("from")
		data.To = r.FormValue("to")

		result, err := convertForm(c, data.Value, data.From, data.To)
		if err != nil {
			data.Error = err.Error()
			render(w, http.StatusBadRequest, data)
			return
		}
		data.Result = result
		render(w, http.StatusOK, data)
	}
}

// convertForm parses the submitted value, converts it and formats the Result.
func convertForm(c convert.Category, value, from, to string) (string, error) {
	v, err := strconv.ParseFloat(value, 64)
	if errors.Is(err, strconv.ErrRange) {
		return "", convert.ErrNotFinite // a number, but too large for float64
	}
	if err != nil {
		return "", fmt.Errorf("%q is not a number", value)
	}
	out, err := c.Convert(v, from, to)
	if err != nil {
		return "", err
	}
	// Convert succeeded, so both Units exist.
	fromUnit, _ := c.Unit(from)
	toUnit, _ := c.Unit(to)
	return fmt.Sprintf("%s %s = %s %s", formatNumber(v+0), fromUnit.Symbol, formatNumber(out), toUnit.Symbol), nil
}

// formatNumber rounds to 6 significant digits; very small or large numbers
// use exponent notation.
func formatNumber(v float64) string {
	return strconv.FormatFloat(v, 'g', 6, 64)
}

// newPage returns the data for a Category's empty form: "from" defaults to
// the first Unit and "to" to the second.
func newPage(c convert.Category) pageData {
	return pageData{Nav: nav, Category: c, From: c.Units[0].ID, To: c.Units[1].ID}
}

// render executes the page into a buffer first, so a template error becomes
// a clean 500 instead of a half-written page with status 200.
func render(w http.ResponseWriter, status int, data pageData) {
	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, "layout", data); err != nil {
		http.Error(w, "internal error rendering page", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}
