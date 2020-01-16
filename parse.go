package ecp

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// parseSlice support slice of string, int, int8, int16, int32, int64
// float32, float64, uint, uint8, uint16, uint32, uint64, bool
func (e *ecp) parseSlice(v string, field reflect.Value) error {
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
	stringSlice := strings.Split(v, e.SplitChar) // split by space

	field.Set(reflect.MakeSlice(field.Type(), len(stringSlice), cap(stringSlice)))

	switch field.Type().Elem().Kind() {
	case reflect.String:
		field.Set(reflect.ValueOf(stringSlice))

	case reflect.Int:
		slice := make([]int, 0, len(stringSlice))
		for _, s := range stringSlice {
			i, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return err
			}
			slice = append(slice, int(i))
		}
		field.Set(reflect.ValueOf(slice))

	case reflect.Int8:
		slice := make([]int8, 0, len(stringSlice))
		for _, s := range stringSlice {
			i, err := strconv.ParseInt(s, 10, 8)
			if err != nil {
				return err
			}
			slice = append(slice, int8(i))
		}
		field.Set(reflect.ValueOf(slice))

	case reflect.Int16:
		slice := make([]int16, 0, len(stringSlice))
		for _, s := range stringSlice {
			i, err := strconv.ParseInt(s, 10, 16)
			if err != nil {
				return err
			}
			slice = append(slice, int16(i))
		}
		field.Set(reflect.ValueOf(slice))

	case reflect.Int32:
		slice := make([]int32, 0, len(stringSlice))
		for _, s := range stringSlice {
			i, err := strconv.ParseInt(s, 10, 32)
			if err != nil {
				return err
			}
			slice = append(slice, int32(i))
		}
		field.Set(reflect.ValueOf(slice))

	case reflect.Int64:
		slice := make([]int64, 0, len(stringSlice))
		for _, s := range stringSlice {
			i, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return err
			}
			slice = append(slice, i)
		}
		field.Set(reflect.ValueOf(slice))

	case reflect.Float32:
		slice := make([]float32, 0, len(stringSlice))
		for _, s := range stringSlice {
			i, err := strconv.ParseFloat(s, 32)
			if err != nil {
				return err
			}
			slice = append(slice, float32(i))
		}
		field.Set(reflect.ValueOf(slice))

	case reflect.Float64:
		slice := make([]float64, 0, len(stringSlice))
		for _, s := range stringSlice {
			i, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return err
			}
			slice = append(slice, float64(i))
		}
		field.Set(reflect.ValueOf(slice))

	case reflect.Uint:
		slice := make([]uint, 0, len(stringSlice))
		for _, s := range stringSlice {
			i, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				return err
			}
			slice = append(slice, uint(i))
		}
		field.Set(reflect.ValueOf(slice))

	case reflect.Uint8:
		slice := make([]uint8, 0, len(stringSlice))
		for _, s := range stringSlice {
			i, err := strconv.ParseUint(s, 10, 8)
			if err != nil {
				return err
			}
			slice = append(slice, uint8(i))
		}
		field.Set(reflect.ValueOf(slice))

	case reflect.Uint16:
		slice := make([]uint16, 0, len(stringSlice))
		for _, s := range stringSlice {
			i, err := strconv.ParseUint(s, 10, 16)
			if err != nil {
				return err
			}
			slice = append(slice, uint16(i))
		}
		field.Set(reflect.ValueOf(slice))

	case reflect.Uint32:
		slice := make([]uint32, 0, len(stringSlice))
		for _, s := range stringSlice {
			i, err := strconv.ParseUint(s, 10, 32)
			if err != nil {
				return err
			}
			slice = append(slice, uint32(i))
		}
		field.Set(reflect.ValueOf(slice))

	case reflect.Uint64:
		slice := make([]uint64, 0, len(stringSlice))
		for _, s := range stringSlice {
			i, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				return err
			}
			slice = append(slice, i)
		}
		field.Set(reflect.ValueOf(slice))

	case reflect.Bool:
		slice := make([]bool, 0, len(stringSlice))
		for _, s := range stringSlice {
			i, err := strconv.ParseBool(s)
			if err != nil {
				return err
			}
			slice = append(slice, i)
		}
		field.Set(reflect.ValueOf(slice))

	default:
		return fmt.Errorf("unsupported slice kind %s", field.Type().Elem().Kind())
	}

	return nil
}

func (e *ecp) parsePointer(typ reflect.Type, value string) (interface{}, error) {
	var rValue interface{}
	switch typ.Kind() {
	case reflect.String:
		rValue = &value

	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		vInt, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, err
		}
		switch typ.Kind() {
		case reflect.Int:
			parsed := int(vInt)
			rValue = &parsed
		case reflect.Int8:
			parsed := int8(vInt)
			rValue = &parsed
		case reflect.Int16:
			parsed := int16(vInt)
			rValue = &parsed
		case reflect.Int32:
			parsed := int32(vInt)
			rValue = &parsed
		case reflect.Int64:
			rValue = &vInt
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return nil, err
		}
		switch typ.Kind() {
		case reflect.Uint:
			parsed := uint(v)
			rValue = &parsed
		case reflect.Uint8:
			parsed := uint8(v)
			rValue = &parsed
		case reflect.Uint16:
			parsed := uint16(v)
			rValue = &parsed
		case reflect.Uint32:
			parsed := uint32(v)
			rValue = &parsed
		case reflect.Uint64:
			rValue = &v
		}

	case reflect.Bool:
		if b, err := strconv.ParseBool(strings.ToLower(value)); err == nil {
			rValue = &b
		} else {
			return nil, err
		}

	case reflect.Float32:
		v, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return nil, err
		}
		x := float32(v)
		rValue = &x

	case reflect.Float64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}
		rValue = &v

	case reflect.Slice:
		newValue := reflect.New(typ)
		if err := e.parseSlice(value, newValue); err != nil {
			return rValue, err
		}
		rValue = newValue

	default:
		return rValue, fmt.Errorf("unsupported pointer kind %s", typ.Kind())
	}

	return rValue, nil
}
