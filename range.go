package ecp

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// the expansion of parseScientific is only ever fed to ParseInt/ParseUint,
// so anything beyond 20 digits is out of range for every Go integer type
const maxExponent = 20

func toValue(config interface{}) reflect.Value {
	value, ok := config.(reflect.Value)
	if !ok {
		value = reflect.ValueOf(config)
	}
	// dereference all pointer levels, e.g. Parse(&c) where c is a pointer
	for value.Kind() == reflect.Ptr && !value.IsNil() {
		value = value.Elem()
	}
	return value
}

// isSection reports whether a field is a nested config section, that is a
// struct or a pointer to a struct. Sections are walked into even when no
// value of their own is available.
func isSection(field reflect.Value) bool {
	switch field.Kind() {
	case reflect.Struct:
		return true
	case reflect.Ptr:
		return field.Type().Elem().Kind() == reflect.Struct
	}
	return false
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
	key    string // key name, empty means "ignore this field"
	defVal string // default value
}

func (e *ECP) getAll(opts getAllOpt) getAllResult {
	field := opts.typ.Field(opts.index)

	r := getAllResult{
		tag:    field.Tag,
		value:  opts.value.Field(opts.index),
		parent: field.Name,
		defVal: field.Tag.Get("default"),
	}

	for _, tag := range []string{"yaml", "json"} {
		v, exist := r.tag.Lookup(tag)
		if !exist {
			continue
		}
		if key := strings.Split(v, ",")[0]; key != "" {
			r.parent = key
			break
		}
	}

	// a "-" tag means "ignore this field", the same way encoding/json
	// reads it; leaving the key empty makes both Parse and List skip it
	if r.parent == "-" || field.Tag.Get("env") == "-" {
		return r
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
	// struct types currently being walked, so that a self referencing
	// type (type Node struct{ Next *Node }) stops instead of recursing
	// until the stack blows up
	visiting map[reflect.Type]bool
}

func (e *ECP) rangeOver(opts roOption) (reflect.Value, error) {

	rValue := toValue(opts.target)
	if !rValue.IsValid() || rValue.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("config must be a struct or a non-nil pointer to a struct, got %v", opts.target)
	}
	rType := rValue.Type()

	if opts.visiting == nil {
		opts.visiting = make(map[reflect.Type]bool, 1)
	}
	opts.visiting[rType] = true
	defer delete(opts.visiting, rType)

	for index := 0; index < rValue.NumField(); index++ {
		if !rType.Field(index).IsExported() {
			continue
		}

		info := e.getAll(getAllOpt{rType, rValue, index, opts.prefix})
		field := info.value
		structName := info.parent
		keyName := info.key
		defaultV := info.defVal

		// ignore this key
		if keyName == "" {
			continue
		}

		section := isSection(field)

		if opts.find != "" {
			if opts.find == keyName {
				return field, nil
			}
			// skip this field
			if !section {
				continue
			}
		}

		v, exist := e.LookupValue(keyName)
		if opts.setDef && !exist {
			v = defaultV
		}

		if !field.CanAddr() || !field.CanSet() {
			// a read-only config can still be searched, just not filled
			if opts.find == "" || !section {
				continue
			}
		}

		kind := field.Kind()
		if v == "" && !section {
			continue
		}

		// set value via self-defined function
		if opts.find == "" && e.Advance.SetValue != nil &&
			e.Advance.SetValue(info.tag, field, v) {
			continue
		}

		switch kind {
		case reflect.Struct:
			prefix := e.BuildKey(opts.prefix, structName, info.tag)
			found, err := e.rangeOver(roOption{
				target:   field,
				setDef:   opts.setDef,
				prefix:   prefix,
				find:     opts.find,
				visiting: opts.visiting,
			})
			if err != nil {
				return reflect.Value{}, err
			}
			if opts.find != "" && found.IsValid() {
				return found, nil
			}

		case reflect.Ptr:
			if section {
				found, err := e.rangeOverPointer(field, structName, info.tag, opts)
				if err != nil {
					return reflect.Value{}, err
				}
				if opts.find != "" && found.IsValid() {
					return found, nil
				}
				continue
			}
			// only set the default value to a nil pointer, but still
			// allow an existing environment value to overwrite a set one
			if !field.IsNil() && !exist {
				continue
			}
			if err := e.setPointer(field, v); err != nil {
				return field, fmt.Errorf("convert %s error: %w", keyName, err)
			}

		case reflect.Slice:
			if !field.IsNil() && !exist {
				continue
			}
			if err := e.parseSlice(v, field); err != nil {
				return field, fmt.Errorf("convert %s error: %w", keyName, err)
			}

		default:
			// a value already set by the caller wins over the default,
			// but never over an environment value
			if !exist && !field.IsZero() {
				continue
			}
			value, err := expandNumber(field.Type(), v)
			if err != nil {
				return field, fmt.Errorf("convert %s error: %w", keyName, err)
			}
			if err := setValue(field, value); err != nil {
				return field, fmt.Errorf("convert %s error: %w", keyName, err)
			}
		}

	}
	return reflect.Value{}, nil
}

// rangeOverPointer walks into a pointer to a struct, that is an optional
// config section. A nil section is filled through a temporary value and
// only allocated when something was actually set, so that an untouched
// optional section stays nil.
func (e *ECP) rangeOverPointer(field reflect.Value, structName string,
	tag reflect.StructTag, opts roOption) (reflect.Value, error) {

	elemType := field.Type().Elem()
	if opts.visiting[elemType] {
		// cyclic type, stop here
		return reflect.Value{}, nil
	}

	target := field
	if field.IsNil() {
		if opts.find != "" || !field.CanSet() {
			// nothing to read from, and nothing to fill
			return reflect.Value{}, nil
		}
		target = reflect.New(elemType)
	}

	found, err := e.rangeOver(roOption{
		target:   target.Elem(),
		setDef:   opts.setDef,
		prefix:   e.BuildKey(opts.prefix, structName, tag),
		find:     opts.find,
		visiting: opts.visiting,
	})
	if err != nil {
		return reflect.Value{}, err
	}

	if field.IsNil() && !target.Elem().IsZero() {
		field.Set(target)
	}

	return found, nil
}

// parseScientific rewrites "1e3" and "1,000" into a plain integer literal
func parseScientific(v string) (string, error) {
	v = strings.ReplaceAll(v, ",", "")

	index := strings.IndexAny(v, "eE")
	if index == -1 {
		return v, nil
	}
	if strings.Count(v, "e")+strings.Count(v, "E") != 1 {
		return "", fmt.Errorf("bad number %s", v)
	}
	if index+1 == len(v) {
		return "", fmt.Errorf("bad number %s", v)
	}
	n, err := strconv.Atoi(v[index+1:])
	if err != nil {
		return "", err
	}
	// a negative exponent would be silently ignored by the expansion
	// below (e.g. "1e-3" -> "1"), which is worse than an error
	if n < 0 {
		return "", fmt.Errorf("bad number %s", v)
	}
	// without an upper bound, "1e1000000" would spend minutes building a
	// one megabyte string that cannot fit in any integer type anyway
	if n > maxExponent {
		return "", fmt.Errorf("number %s out of range", v)
	}

	mantissa := v[:index]
	// shift the decimal point instead of blindly appending zeros, so that
	// "1.5e3" becomes 1500 instead of the unparsable "1.5000"
	if dot := strings.IndexByte(mantissa, '.'); dot != -1 {
		decimals := len(mantissa) - dot - 1
		if decimals > n {
			return "", fmt.Errorf("number %s is not an integer", v)
		}
		mantissa = mantissa[:dot] + mantissa[dot+1:]
		n -= decimals
	}

	return mantissa + strings.Repeat("0", n), nil
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
		hours := int64(dayN) * 24
		if dayN != 0 && hours/int64(dayN) != 24 {
			return 0, fmt.Errorf("duration %s out of range", v)
		}
		v = fmt.Sprintf("%dh", hours)
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, err
	}

	return d, nil
}
