# unit-converter

[![CI](https://github.com/stonalampa/go-mastery-unit-converter/actions/workflows/ci.yml/badge.svg)](https://github.com/stonalampa/go-mastery-unit-converter/actions/workflows/ci.yml)

A small Go web app that converts length, weight and temperature between units. It is my solution to the [roadmap.sh Unit Converter](https://roadmap.sh/projects/unit-converter) project and part of my `go-mastery` series. It uses only the Go standard library, renders HTML on the server, and has no JavaScript.

## Run

```sh
go run ./cmd/unit-converter                 # http://localhost:8080
go run ./cmd/unit-converter -addr :3000     # another address
make build && ./bin/unit-converter          # a single binary with the HTML embedded
```

There is one page per category: `/length`, `/weight` and `/temperature`. `/` redirects to `/length`. Each page has a form with a value, a "from" unit and a "to" unit. The form posts back to the same page, which shows the result, or an error with HTTP 400. The value and selected units stay in the form.

```text
POST /length       value=1    from=mile      to=kilometer    → 1 mi = 1.60934 km
POST /weight       value=2.5  from=kilogram  to=pound        → 2.5 kg = 5.51156 lb
POST /temperature  value=100  from=celsius   to=fahrenheit   → 100 °C = 212 °F
POST /length       value=-5   from=meter     to=foot         → 400: value must not be negative
POST /temperature  value=-300 from=celsius   to=kelvin       → 400: value is below absolute zero
```

Units:

| Category | Units | Base unit |
|---|---|---|
| Length | millimeter, centimeter, meter, kilometer, inch, foot, yard, mile | meter |
| Weight | milligram, gram, kilogram, ounce, pound | gram |
| Temperature | Celsius, Fahrenheit, Kelvin | kelvin |

## How conversion works

Every unit has a `Factor` and an `Offset` that map it onto its category's base unit:

```text
base = value × Factor + Offset          e.g. °C → K:  Factor 1, Offset 273.15
```

For length and weight, `Offset` is 0, so the conversion is plain scaling (1 mi = 1609.344 m). Temperature needs the offset because 0 °C is not 0 K. A conversion goes from the source unit to the base unit and then to the target unit, computed as a single expression with the factor ratio first:

```text
out = value × (from.Factor ÷ to.Factor) + (from.Offset − to.Offset) ÷ to.Factor
```

Computing the ratio first means a huge value (e.g. 1e308 km) cannot overflow in an intermediate step when the result itself fits in a `float64`.

Converting through one base unit needs one number pair per unit, instead of one formula for every pair of units.

Validation:

- A value must be a finite number. Only `.` is accepted as the decimal separator.
- Each unit has a minimum: 0 for length, weight and Kelvin, −273.15 for Celsius, −459.67 for Fahrenheit. A value below it is rejected with its category's error: "value must not be negative" or "value is below absolute zero".
- An unknown unit is rejected. This can only happen with a request sent outside the form.

Results are rounded to 6 significant digits, and very small or very large values use exponent notation. Results that are off zero only by temperature float noise are shown as 0.

Factors use the exact definitions: 1 in = 2.54 cm, 1 ft = 0.3048 m, 1 yd = 0.9144 m, 1 mi = 1609.344 m, 1 lb = 453.59237 g, 1 oz = 1/16 lb.

## Go concepts this project practises

- **`net/http` server.** It uses Go 1.22 method routing (`GET /length`, `POST /length`, `GET /{$}` for the exact root). Routes are registered in a loop over the categories, and each handler is a closure over its category.
- **`html/template`.** One layout and one data-driven page template serve all three categories. Output is escaped automatically, which protects against XSS. Each page is rendered into a buffer first, so a template error becomes a clean 500 instead of half a page.
- **`embed`.** The templates are compiled into the binary.
- **A production-shaped server.** It sets read-header, read, write and idle timeouts. `signal.NotifyContext` plus `Server.Shutdown` give a graceful shutdown on Ctrl-C or SIGTERM: requests that are in flight get up to 10 seconds to finish before the process exits.
- **Testing.** Handler tests go through `httptest` with real form posts and are table-driven with `t.Run`. The server start, shutdown and address-in-use paths are tested through a testable `run(ctx, args, stderr)`.
- **Fuzzing.** `FuzzConvertRoundTrip` checks that A → B → A returns the original value for every unit pair. Its seed corpus runs in every `go test`.

## Project layout

```text
cmd/unit-converter/   flags, server, graceful shutdown
internal/convert/     Category, Unit, Convert (no HTTP); fuzz test
internal/web/         routes, handlers, templates (embedded), handler tests
docs/adr/             architecture decision records
```

## Development

```sh
make test     # go test -race ./...   (includes the fuzz seed corpus)
make vet
make lint     # golangci-lint v2
make vuln     # govulncheck
go test -fuzz FuzzConvertRoundTrip -fuzztime 1m ./internal/convert   # explore beyond the seeds
```

`go.mod` requires Go 1.27.2, because earlier 1.27 releases have `net/http` vulnerabilities that this server would reach.

CI runs the shared workflow from [`stonalampa/.github`](https://github.com/stonalampa/.github), pinned to `v1` (see [ADR-0001](docs/adr/0001-shared-ci-workflow-pinned-to-v1.md)). `main` accepts changes only through pull requests with a green `ci / check`.

The work was planned as a spec ([#1](https://github.com/stonalampa/go-mastery-unit-converter/issues/1)) and split into small tickets. Each ticket became one PR, built test-first and reviewed by an independent AI reviewer (OpenAI Codex). The temperature design was mine.
