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

type gaOption struct {
	rType  reflect.Type
	rValue reflect.Value
	index  int    // field index
	pName  string // parent field name (struct name)
}

type getAllResult struct {
	rValue reflect.Value
	rTag   reflect.StructTag
	sName  string // struct name
	kName  string // key name
	value  string // default value
}

func (e *ecp) getAll(opts gaOption) getAllResult {
	structField := opts.rType.Field(opts.index)

	r := getAllResult{
		rTag:   structField.Tag,
		rValue: opts.rValue.Field(opts.index),
		sName:  structField.Name,
		value:  structField.Tag.Get("default"),
	}

	// support yaml or json
	if v, exist := r.rTag.Lookup("yaml"); exist {
		r.sName = strings.Split(v, ",")[0]
	}
	if v, exist := r.rTag.Lookup("json"); exist {
		r.sName = strings.Split(v, ",")[0]
	}

	r.kName = e.GetKey(opts.pName, r.sName, r.rTag)

	return r
}

// range over option
type roOption struct {
	target interface{}
	setDef bool   // set default value
	prefix string // prefix, usually the parent struct name
	lookup string // lookup some key
}

func (e *ecp) rangeOver(opts roOption) (reflect.Value, error) {

	rValue := toValue(opts.target)
	rType := rValue.Type()

	fieldNum := rValue.NumField()
	for index := 0; index < fieldNum; index++ {
		all := e.getAll(gaOption{rType, rValue, index, opts.prefix})
		field := all.rValue
		structName := all.sName
		keyName := all.kName
		defaultV := all.value

		if opts.lookup != "" {
			keyName = e.LookupKey(keyName, opts.prefix, structName)
			if !strings.HasPrefix(opts.lookup, keyName) {
				continue
			}
			if opts.lookup == keyName {
				return field, nil
			} else if index == fieldNum {
				return field, fmt.Errorf("key %s not found", opts.lookup)
			}
		}

		// ignore this key
		if e.IgnoreKey(field, structName) || e.IgnoreKey(field, keyName) {
			continue
		}

		v, exist := e.LookupValue(field, keyName)
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
			if v == "" {
				continue
			}

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
			prefix := e.GetKey(opts.prefix, structName, all.rTag)
			if opts.lookup != "" {
				prefix = structName
				if opts.prefix != "" {
					prefix = opts.prefix + "." + structName
				}
			}
			v, err := e.rangeOver(roOption{field, opts.setDef, prefix, opts.lookup})
			if err != nil {
				return field, err
			} else if opts.lookup != "" {
				return v, nil
			}
			field = v

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
