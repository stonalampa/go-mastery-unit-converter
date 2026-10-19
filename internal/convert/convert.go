// Package convert holds the Categories and Units of the unit converter and
// performs Conversions between Units of the same Category.
package convert

import (
	"errors"
	"fmt"
	"math"
)

// Errors returned by Convert; they are wrapped with details, so check them
// with errors.Is.
var (
	ErrUnknownUnit = errors.New("unknown unit")
	ErrNegative    = errors.New("value must not be negative")
	ErrNotFinite   = errors.New("value must be a finite number")
)

// Unit is one named measure within a Category.
type Unit struct {
	ID     string  // form value, e.g. "meter"
	Name   string  // shown in dropdowns, e.g. "Meter"
	Symbol string  // shown in results, e.g. "m"
	Factor float64 // how many base units (e.g. meters) one of this Unit is
}

// Category is a family of Units that can be converted into each other.
type Category struct {
	ID    string // URL path segment, e.g. "length"
	Name  string // page title, e.g. "Length"
	Units []Unit
}

// Unit returns the Unit with the given ID.
func (c Category) Unit(id string) (Unit, bool) {
	for _, u := range c.Units {
		if u.ID == id {
			return u, true
		}
	}
	return Unit{}, false
}

// Convert converts value from the Unit with ID from to the Unit with ID to.
func (c Category) Convert(value float64, from, to string) (float64, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, ErrNotFinite
	}
	if value < 0 {
		return 0, ErrNegative
	}
	value += 0 // turns -0 into 0, so "-0" is shown as "0"
	fromUnit, ok := c.Unit(from)
	if !ok {
		return 0, fmt.Errorf("%w %q", ErrUnknownUnit, from)
	}
	toUnit, ok := c.Unit(to)
	if !ok {
		return 0, fmt.Errorf("%w %q", ErrUnknownUnit, to)
	}
	if from == to {
		return value, nil // exact; value*f/f may differ in the last bit
	}
	// Ratio first: it is a modest number, so a large value can't overflow in
	// an intermediate step when the result itself fits in a float64.
	out := value * (fromUnit.Factor / toUnit.Factor)
	if math.IsInf(out, 0) {
		return 0, ErrNotFinite
	}
	return out, nil
}

// Length is the length Category; its base unit is the meter.
var Length = Category{
	ID:   "length",
	Name: "Length",
	Units: []Unit{
		{ID: "millimeter", Name: "Millimeter", Symbol: "mm", Factor: 0.001},
		{ID: "centimeter", Name: "Centimeter", Symbol: "cm", Factor: 0.01},
		{ID: "meter", Name: "Meter", Symbol: "m", Factor: 1},
		{ID: "kilometer", Name: "Kilometer", Symbol: "km", Factor: 1000},
		{ID: "inch", Name: "Inch", Symbol: "in", Factor: 0.0254},
		{ID: "foot", Name: "Foot", Symbol: "ft", Factor: 0.3048},
		{ID: "yard", Name: "Yard", Symbol: "yd", Factor: 0.9144},
		{ID: "mile", Name: "Mile", Symbol: "mi", Factor: 1609.344},
	},
}

// Weight is the weight Category; its base unit is the gram. Pound and ounce
// use the international definitions (1 lb = 453.59237 g, 1 oz = 1/16 lb).
var Weight = Category{
	ID:   "weight",
	Name: "Weight",
	Units: []Unit{
		{ID: "milligram", Name: "Milligram", Symbol: "mg", Factor: 0.001},
		{ID: "gram", Name: "Gram", Symbol: "g", Factor: 1},
		{ID: "kilogram", Name: "Kilogram", Symbol: "kg", Factor: 1000},
		{ID: "ounce", Name: "Ounce", Symbol: "oz", Factor: 28.349523125},
		{ID: "pound", Name: "Pound", Symbol: "lb", Factor: 453.59237},
	},
}
