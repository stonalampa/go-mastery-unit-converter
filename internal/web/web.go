// Package web serves the unit converter's pages.
package web

import (
	"bytes"
	"embed"
	"html/template"
	"net/http"

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

// pageData is everything the page template renders.
type pageData struct {
	Nav      []navLink
	Category convert.Category
	From, To string // selected Unit IDs
}

// NewHandler returns the app's HTTP handler with all routes registered.
func NewHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/length", http.StatusSeeOther)
	})
	mux.HandleFunc("GET /length", func(w http.ResponseWriter, _ *http.Request) {
		render(w, http.StatusOK, newPage(convert.Length))
	})
	return mux
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
