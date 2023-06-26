package ecp

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func toValue(config interface{}) reflect.Value {
	value, ok := config.(reflect.Value)
	if !ok {
		value = reflect.Indirect(reflect.ValueOf(config))
	}
	return value
}

type getAllOpt struct {
	typ    reflect.Type
	value  reflect.Value
	index  int    // field index
	parent string // parent field name (struct name)
}

type getAllResult struct {
	value  reflect.Value
	tag    reflect.StructTag
	parent string // struct name
	key    string // key name
	defVal string // default value
}

func (e *ecp) getAll(opts getAllOpt) getAllResult {
	field := opts.typ.Field(opts.index)

	r := getAllResult{
		tag:    field.Tag,
		value:  opts.value.Field(opts.index),
		parent: field.Name,
		defVal: field.Tag.Get("default"),
	}

	// support yaml or json
	if v, exist := r.tag.Lookup("yaml"); exist {
		r.parent = strings.Split(v, ",")[0]
	} else if v, exist := r.tag.Lookup("json"); exist {
		r.parent = strings.Split(v, ",")[0]
	}

	r.key = e.BuildKey(opts.parent, r.parent, r.tag)

	return r
}

// range over option
type roOption struct {
	target interface{}
	setDef bool   // set default value
	prefix string // prefix, usually the parent struct name
	find   string // lookup some key
}

func (e *ecp) rangeOver(opts roOption) (reflect.Value, error) {

	rValue := toValue(opts.target)
	rType := rValue.Type()

	for index := 0; index < rValue.NumField(); index++ {
		info := e.getAll(getAllOpt{rType, rValue, index, opts.prefix})
		field := info.value
		structName := info.parent
		keyName := info.key
		defaultV := info.defVal

		// ignore this key
		if keyName == "" {
			continue
		}

		if opts.find != "" {
			if opts.find == keyName {
				return field, nil
			}
			// skip this field
			if field.Kind() != reflect.Struct {
				continue
			}
		}

		v, exist := e.LookupValue(keyName)
		if opts.setDef && !exist {
			v = defaultV
		}

		if !field.CanAddr() || !field.CanSet() {
			continue
		}

		kind := field.Kind()
		if v == "" && kind != reflect.Struct {
			continue
		}

		switch kind {
		case reflect.String:
			if field.String() != "" && !exist {
				continue
			}
			field.SetString(v)

		case reflect.Float32, reflect.Float64:
			if field.Float() != 0 && !exist {
				continue
			}
			parsed, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return field, fmt.Errorf("convert %s error: %s", keyName, err)
			}
			field.SetFloat(parsed)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if field.Int() != 0 && !exist {
				continue
			}
			// since duration is int64 too, parse it first
			// if the duration contains `d` (day), we should support it
			// fix #6
			d, err := parseDuration(v)
			if err == nil {
				field.SetInt(int64(d))
				continue
			}
			v, err = parseScientific(v)
			if err != nil {
				return field, fmt.Errorf("convert %s error: %s", keyName, err)
			}
			parsed, err := strconv.Atoi(v)
			if err != nil {
				return field, fmt.Errorf("convert %s error: %s", keyName, err)
			}
			field.SetInt(int64(parsed))

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if field.Uint() != 0 && !exist {
				continue
			}
			v, err := parseScientific(v)
			if err != nil {
				return field, fmt.Errorf("convert %s error: %s", keyName, err)
			}
			parsed, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return field, fmt.Errorf("convert %s error: %s", keyName, err)
			}
			field.SetUint(parsed)

		case reflect.Bool:
			parsed, err := strconv.ParseBool(strings.ToLower(v))
			if err != nil {
				return field, fmt.Errorf("convert %s error: %s", keyName, err)
			}
			field.SetBool(parsed)

		case reflect.Slice:
			if !field.IsNil() && !exist {
				continue
			}
			if err := e.parseSlice(v, field); err != nil {
				return field, fmt.Errorf("convert %s error: %s", keyName, err)
			}

		case reflect.Struct:
			prefix := e.BuildKey(opts.prefix, structName, info.tag)
			v, err := e.rangeOver(roOption{field, opts.setDef, prefix, opts.find})
			if err != nil {
				return reflect.Value{}, err
			}
			if opts.find != "" && v.IsValid() {
				return v, nil
			}

		case reflect.Ptr:
			// only set default value to nil pointer
			if !field.IsNil() {
				continue
			}
			// get pointer real kind
			value, err := e.parsePointer(field.Type().Elem(), v)
			if err != nil {
				return field, fmt.Errorf("convert %s error: %s", keyName, err)
			}
			if value != nil {
				field.Set(reflect.ValueOf(value))
			}
		}

	}
	return reflect.Value{}, nil
}

func parseScientific(v string) (string, error) {
	switch {
	case strings.Contains(v, ","):
		v = strings.ReplaceAll(v, ",", "")
	case strings.Contains(v, "e"):
		v = strings.ReplaceAll(v, "e", "E")
		fallthrough
	case strings.Contains(v, "E"):
		if strings.Count(v, "E") != 1 {
			return "", fmt.Errorf("bad number %s", v)
		}
		index := strings.Index(v, "E")
		if index+1 == len(v) {
			return "", fmt.Errorf("bad number %s", v)
		}
		result := v[:index]
		n, err := strconv.Atoi(v[index+1:])
		if err != nil {
			return "", err
		}
		for i := 0; i < n; i++ {
			result += "0"
		}
		v = result
	}
	return v, nil
}

// parseDuration wrapper func of time.ParseDuration to support `Xd` = `X*24h`
func parseDuration(v string) (time.Duration, error) {
	last := len(v) - 1
	if last > 0 && v[last] == 'd' {
		day := v[:last]
		dayN, err := strconv.Atoi(day)
		if err != nil {
			return 0, err
		}
		v = fmt.Sprintf("%dh", dayN*24)
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, err
	}

	return d, nil
}
