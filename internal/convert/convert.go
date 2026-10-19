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
	ErrBelowZeroK  = errors.New("value is below absolute zero")
	ErrNotFinite   = errors.New("value must be a finite number")
)

// Unit is one named measure within a Category.
type Unit struct {
	ID     string  // form value, e.g. "meter"
	Name   string  // shown in dropdowns, e.g. "Meter"
	Symbol string  // shown in results, e.g. "m"
	Factor float64 // how many base units (e.g. meters) one of this Unit is
	Offset float64 // added after Factor on the way to the base unit (e.g. 273.15 for Celsius)
	Min    float64 // lowest allowed value in this Unit: 0 for length and weight, -273.15 for Celsius
}

// Category is a family of Units that can be converted into each other.
type Category struct {
	ID    string // URL path segment, e.g. "length"
	Name  string // page title, e.g. "Length"
	Units []Unit
	// BelowZero is the error for a value that is below zero in the base
	// unit: a negative length or weight, a temperature below absolute zero.
	BelowZero error
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
	value += 0 // turns -0 into 0, so "-0" is shown as "0"
	fromUnit, ok := c.Unit(from)
	if !ok {
		return 0, fmt.Errorf("%w %q", ErrUnknownUnit, from)
	}
	toUnit, ok := c.Unit(to)
	if !ok {
		return 0, fmt.Errorf("%w %q", ErrUnknownUnit, to)
	}
	// Each Unit knows its own lower bound (0 m, -273.15 °C, ...), so the check
	// is an exact comparison in the Unit the user typed, with no float rounding.
	if value < fromUnit.Min {
		return 0, c.BelowZero
	}
	if from == to {
		return value, nil // exact; value*f/f may differ in the last bit
	}
	// The same as (value*from.Factor + from.Offset - to.Offset) / to.Factor,
	// arranged so the Factor ratio comes first: a huge value whose result
	// fits in a float64 can't overflow in an intermediate step.
	out := value*(fromUnit.Factor/toUnit.Factor) + (fromUnit.Offset-toUnit.Offset)/toUnit.Factor
	if math.IsInf(out, 0) {
		return 0, ErrNotFinite
	}
	// Offsets leave float noise where the result should be exactly 0
	// (e.g. 32 °F -> 5.7e-14 °C). Only offset Conversions are rounded, so
	// genuinely tiny lengths and weights stay as they are.
	if (fromUnit.Offset != 0 || toUnit.Offset != 0) && math.Abs(out) < 1e-9 {
		out = 0
	}
	return out, nil
}

// Length is the length Category; its base unit is the meter.
var Length = Category{
	ID:        "length",
	Name:      "Length",
	BelowZero: ErrNegative,
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
	ID:        "weight",
	Name:      "Weight",
	BelowZero: ErrNegative,
	Units: []Unit{
		{ID: "milligram", Name: "Milligram", Symbol: "mg", Factor: 0.001},
		{ID: "gram", Name: "Gram", Symbol: "g", Factor: 1},
		{ID: "kilogram", Name: "Kilogram", Symbol: "kg", Factor: 1000},
		{ID: "ounce", Name: "Ounce", Symbol: "oz", Factor: 28.349523125},
		{ID: "pound", Name: "Pound", Symbol: "lb", Factor: 453.59237},
	},
}

// Temperature is the temperature Category; its base unit is the kelvin.
// Unlike length and weight it needs an Offset: 0 °C is 273.15 K and
// 0 °F is 459.67 × 5/9 K.
var Temperature = Category{
	ID:        "temperature",
	Name:      "Temperature",
	BelowZero: ErrBelowZeroK,
	Units: []Unit{
		{ID: "celsius", Name: "Celsius", Symbol: "°C", Factor: 1, Offset: 273.15, Min: -273.15},
		{ID: "fahrenheit", Name: "Fahrenheit", Symbol: "°F", Factor: 5.0 / 9, Offset: 459.67 * 5.0 / 9, Min: -459.67},
		{ID: "kelvin", Name: "Kelvin", Symbol: "K", Factor: 1},
	},
}
