# Standard library, server-rendered HTML, no JavaScript

The app uses `net/http` with Go 1.22 method patterns, `html/template` and `embed`, and plain HTML forms that post back to the same page, with no web framework (an earlier unfinished attempt used Fiber) and no JavaScript (it used htmx). The roadmap.sh spec asks for exactly this submit-and-re-render flow, the standard library covers it, and a framework built on a non-standard HTTP stack would teach less idiomatic Go; htmx's partial page updates are a different model from the one the spec describes.
