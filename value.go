package ecp

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var durationType = reflect.TypeOf(time.Duration(0))

// setValue sets a single scalar value from its string form.
//
// The value is always assigned through the field's own reflect.Value, so
// named types (type Level int, type Name string, time.Duration, ...) are
// handled like their underlying kind instead of panicking on an
// unassignable concrete type such as []int -> []Level.
func setValue(field reflect.Value, v string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(v)

	case reflect.Bool:
		b, err := strconv.ParseBool(strings.ToLower(v))
		if err != nil {
			return err
		}
		field.SetBool(b)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// only time.Duration (an int64 based type) accepts duration syntax
		// like "10s" or "1d"; parsing a plain int field that way would
		// silently turn "10s" into 1e10
		if field.Type() == durationType {
			d, err := parseDuration(v)
			if err != nil {
				return err
			}
			field.SetInt(int64(d))
			return nil
		}
		// parse with the field's bit size so an out-of-range value errors
		// out instead of being silently truncated
		n, err := strconv.ParseInt(v, 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(v, 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetUint(n)

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(v, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetFloat(f)

	default:
		return fmt.Errorf("unsupported kind %s", field.Kind())
	}

	return nil
}

// expandNumber rewrites the scientific and thousand-separator notation
// ("1e3", "1,000") into a plain integer literal. Duration fields keep
// their own syntax, and non integer kinds are returned untouched.
//
// It is deliberately not applied to slice elements: there, "1,2" is much
// more likely to be a wrong separator than the number 12.
func expandNumber(typ reflect.Type, v string) (string, error) {
	switch typ.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if typ == durationType {
			return v, nil
		}
		return parseScientific(v)
	}
	return v, nil
}

// canSetKind reports whether a field of this kind can be filled from a
// string value at all. Unsupported kinds (map, array, chan, ...) are
// reported so that they can be skipped when listing and rejected with a
// clear error when a value is actually provided for them.
func canSetKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Slice, reflect.Ptr, reflect.Struct:
		return true
	}
	return false
}
