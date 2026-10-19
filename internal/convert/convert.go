// Package convert holds the Categories and Units of the unit converter.
package convert

// Unit is one named measure within a Category.
type Unit struct {
	ID     string // form value, e.g. "meter"
	Name   string // shown in dropdowns, e.g. "Meter"
	Symbol string // shown in results, e.g. "m"
}

// Category is a family of Units that can be converted into each other.
type Category struct {
	ID    string // URL path segment, e.g. "length"
	Name  string // page title, e.g. "Length"
	Units []Unit
}

// Length is the length Category.
var Length = Category{
	ID:   "length",
	Name: "Length",
	Units: []Unit{
		{ID: "millimeter", Name: "Millimeter", Symbol: "mm"},
		{ID: "centimeter", Name: "Centimeter", Symbol: "cm"},
		{ID: "meter", Name: "Meter", Symbol: "m"},
		{ID: "kilometer", Name: "Kilometer", Symbol: "km"},
		{ID: "inch", Name: "Inch", Symbol: "in"},
		{ID: "foot", Name: "Foot", Symbol: "ft"},
		{ID: "yard", Name: "Yard", Symbol: "yd"},
		{ID: "mile", Name: "Mile", Symbol: "mi"},
	},
}
