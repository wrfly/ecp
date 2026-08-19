// Package ecp can help you convert environments into configurations
// it's an environment config parser
package ecp

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// ECP is an environment config parser. Create one with New when the
// default behaviour has to be changed, or use the package level Parse,
// List and Get functions to work with the default one.
type ECP struct {
	// BuildKey builds the environment key of a field
	BuildKey BuildKeyFunc
	// LookupValue returns the value of a key and whether it exists
	LookupValue LookupValueFunc

	Advance AdvanceConfig
}

// AdvanceConfig holds the optional knobs of an ECP
type AdvanceConfig struct {
	SplitChar string // split slice
	SetValue  SetValueFunc
}

var globalEcp = New()

// New ecp object
func New() *ECP {
	return &ECP{
		BuildKey:    buildKeyFromEnv,
		LookupValue: lookupValueFromEnv,
		Advance: AdvanceConfig{
			SplitChar: space,
		},
	}
}

// Parse the configuration through environments starting with the
// prefix (or not), see the package level Parse for the details
func (e *ECP) Parse(config interface{}, prefix ...string) error {
	if len(prefix) == 0 {
		prefix = []string{""}
	}

	// catch the classic Parse(config) instead of Parse(&config): without
	// a pointer every field is read-only, so Parse would report success
	// after filling in exactly nothing
	if value := toValue(config); value.IsValid() &&
		value.Kind() == reflect.Struct && !value.CanSet() {
		return fmt.Errorf("config must be a pointer to a struct, got %s", value.Type())
	}

	_, err := e.rangeOver(roOption{target: config, setDef: true, prefix: prefix[0]})
	return err
}

// List all the config environments, see the package level List for
// the details
func (e *ECP) List(config interface{}, prefix ...string) []string {
	if len(prefix) == 0 {
		prefix = []string{""}
	}
	return e.list(config, prefix[0], make(map[reflect.Type]bool, 1))
}

func (e *ECP) list(config interface{}, parentName string,
	visiting map[reflect.Type]bool) []string {

	list := []string{}

	configValue := toValue(config)
	if !configValue.IsValid() || configValue.Kind() != reflect.Struct {
		return list
	}
	configType := configValue.Type()

	// stop a self referencing type from recursing forever
	if visiting[configType] {
		return list
	}
	visiting[configType] = true
	defer delete(visiting, configType)

	for index := 0; index < configValue.NumField(); index++ {
		if !configType.Field(index).IsExported() {
			continue
		}
		all := e.getAll(getAllOpt{configType, configValue, index, parentName})
		if all.key == "" {
			continue
		}
		switch {
		case all.value.Kind() == reflect.Struct:
			prefix := e.BuildKey(parentName, all.parent, all.tag)
			list = append(list, e.list(all.value, prefix, visiting)...)

		case isSection(all.value):
			// an optional section: list the keys of the pointed-to
			// struct, a nil pointer still has all of them
			prefix := e.BuildKey(parentName, all.parent, all.tag)
			section := all.value
			if section.IsNil() {
				section = reflect.New(all.value.Type().Elem())
			}
			list = append(list, e.list(section.Elem(), prefix, visiting)...)

		case !canSetKind(all.value.Kind()):
			// maps, arrays, channels... cannot be filled from a string,
			// so listing a key for them would be misleading
			continue

		default:
			list = append(list, fmt.Sprintf("%s=%s", all.key, quoteValue(all.defVal)))
		}
	}

	return list
}

// quoteValue quotes a default value that would not survive a round trip
// through a shell or an env file unquoted
func quoteValue(v string) string {
	if strings.ContainsAny(v, " \t\r\n\"'\\`$") {
		return strconv.Quote(v)
	}
	return v
}

// List all the config environments.
//
// The value of each key is the one from the "default" tag, empty if the
// field has no default. Fields tagged with `env:"-"`, `yaml:"-"` or
// `json:"-"` are skipped.
func List(config interface{}, prefix ...string) []string {
	return globalEcp.List(config, prefix...)
}

// Parse the configuration through environments starting with
// the prefix (or not)
// ecp.Parse(&config) or ecp.Parse(&config, "PREFIX")
//
// Parse will overwrite the existing value if there is an environment
// configuration matched with the struct name or the "env" tag
// name.
//
// Also, Parse will set the default value to the config, if it's not set
// values. For basic types, if the value is zero value, then it will be
// set to the default value. You can change the basic type to a pointer
// type, thus Parse will only set the default value when the field is
// nil, not the zero value.
// for example:
//
//	type config struct {
//	    One   string   `default:"1"`
//	    Two   int      `default:"2"`
//	    Three []string `default:"1 2 3"`
//	}
//	c := &config{}
//
// Slice values are separated by Advance.SplitChar, a space by default.
//
// config must be a pointer to a struct, otherwise Parse returns an error
// instead of silently doing nothing.
func Parse(config interface{}, prefix ...string) error {
	return globalEcp.Parse(config, prefix...)
}
