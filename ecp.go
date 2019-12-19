// Package ecp can help you convert environments into configurations
// it's an environment config parser
package ecp

import (
	"fmt"
	"reflect"
	"strings"
)

type ecp struct {
	GetKey      GetKeyFunc
	LookupValue LookupValueFunc
	IgnoreKey   IgnoreKeyFunc
	LookupKey   LookupKeyFunc
}

var globalEcp = &ecp{
	GetKey:      getKeyFromEnv,
	IgnoreKey:   ignoreEnvKey,
	LookupValue: lookupValueFromEnv,
	LookupKey:   lookupKey,
}

// New ecp object
func New() *ecp {
	return &ecp{
		GetKey:      getKeyFromEnv,
		IgnoreKey:   ignoreEnvKey,
		LookupValue: lookupValueFromEnv,
		LookupKey:   lookupKey,
	}
}

func (e *ecp) Parse(config interface{}, prefix ...string) error {
	if prefix == nil {
		prefix = []string{"ECP"}
	}
	_, err := e.rangeOver(roOption{config, true, prefix[0], ""})
	return err
}

func (e *ecp) Default(config interface{}) error {
	_, err := e.rangeOver(roOption{config, true, "", ""})
	return err
}

func (e *ecp) List(config interface{}, prefix ...string) []string {
	list := []string{}

	if prefix == nil {
		prefix = []string{"ECP"}
	}
	parentName := prefix[0]

	configValue := toValue(config)
	configType := configValue.Type()
	for index := 0; index < configValue.NumField(); index++ {
		all := e.getAll(gaOption{configType, configValue, index, parentName})
		if all.sName == "-" || all.kName == "" {
			continue
		}
		switch all.rValue.Kind() {
		case reflect.Struct:
			prefix := e.GetKey(parentName, all.sName, all.rTag)
			list = append(list, e.List(all.rValue, prefix)...)
		default:
			if strings.Contains(all.value, " ") {
				all.value = fmt.Sprintf("\"%s\"", all.value)
			}
			list = append(list, fmt.Sprintf("%s=%s", all.kName, all.value))
		}
	}

	return list
}

// List function will also fill up the value of the environment key
// it the "default" tag has value

// List all the config environments
func List(config interface{}, prefix ...string) []string {
	return globalEcp.List(config, prefix...)
}

// Parse the configuration through environments starting with the prefix
// or you can ignore the prefix and the default prefix key will be `ECP`
// ecp.Parse(&config) or ecp.Parse(&config, "PREFIX")
//
// Parse will overwrite the existing value if there is an environment
// configration matched with the struct name or the "env" tag
// name.
//
// Also, Parse will set the default value to the config, if it's not set
// values. For basic types, if the value is zero value, then it will be
// set to the default value. You can change the basic type to a pointer
// type, thus Parse will only set the default value when the field is
// nil, not the zero value.
func Parse(config interface{}, prefix ...string) error {
	return globalEcp.Parse(config, prefix...)
}

// the default value of the config is set by a tag named "default"
// for example, you can define a struct like:
//
//    type config struct {
//        One   string   `default:"1"`
//        Two   int      `default:"2"`
//        Three []string `default:"1,2,3"`
//    }
//    c := &config{}
//
// then you can use ecp.Default(&c) to parse the default value to the struct.
// note, the Default function will not overwrite the existing value, if the
// config key has already been set no matter whether it has a default tag.
// And the default value will be nil (nil of the type) if the "default" tag is
// empty.

// Default set config with its default value
// DEPRECATED: just use `Parse`
func Default(config interface{}) error {
	return globalEcp.Default(config)
}
