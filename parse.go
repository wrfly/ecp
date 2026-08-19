package ecp

import (
	"fmt"
	"reflect"
	"strings"
)

// split cuts a value into slice elements. With the default whitespace
// separator, repeated separators are collapsed, so "a  b" yields two
// elements instead of three (one of them empty and unparsable).
func (e *ECP) split(v string) []string {
	if strings.TrimSpace(e.Advance.SplitChar) == "" {
		return strings.Fields(v)
	}
	return strings.Split(v, e.Advance.SplitChar)
}

// parseSlice supports slices of string, bool, int, int8, int16, int32,
// int64, uint, uint8, uint16, uint32, uint64, float32, float64 and
// time.Duration, including named types built on top of them.
func (e *ECP) parseSlice(v string, field reflect.Value) error {
	if v == "" {
		return nil
	}

	if !field.CanAddr() {
		return fmt.Errorf("field is not addressable")
	}
	if field.Kind() != reflect.Slice {
		return fmt.Errorf("field is not slice")
	}

	// either space nor commas is perfect, but I think space is better
	// since it's more natural: fmt.Println([]int{1, 2, 3}) = [1 2 3]
	parts := e.split(v)

	// build the slice through the field's own type so that a named
	// element type ([]Level) stays assignable
	slice := reflect.MakeSlice(field.Type(), len(parts), len(parts))
	for i, s := range parts {
		if err := setValue(slice.Index(i), s); err != nil {
			return err
		}
	}
	field.Set(slice)

	return nil
}

// setPointer fills a pointer field, allocating the pointed-to value.
//
// The value is built with reflect.New from the field's own element type,
// which keeps named pointer types (*time.Duration, *Level, ...)
// assignable and lets *time.Duration accept the same "10s" syntax as
// time.Duration.
func (e *ECP) setPointer(field reflect.Value, v string) error {
	elemType := field.Type().Elem()
	pointer := reflect.New(elemType)

	if elemType.Kind() == reflect.Slice {
		// parseSlice needs the slice value itself, not the pointer to it
		if err := e.parseSlice(v, pointer.Elem()); err != nil {
			return err
		}
	} else {
		v, err := expandNumber(elemType, v)
		if err != nil {
			return err
		}
		if err := setValue(pointer.Elem(), v); err != nil {
			return err
		}
	}

	field.Set(pointer)
	return nil
}
