// Package ecp can help you convert environments into configurations
// it's an environment config parser
package ecp

import (
	"fmt"
	"reflect"
	"strings"
)

// note Parse function will overwrite the existing value if there is a
// environment configration matched with the struct name or the "env" tag
// name.

// Parse the configuration through environments starting with the prefix
// or you can ignore the prefix and the default prefix key will be `ECP`
// ecp.Parse(&config) or ecp.Parse(&config, "PREFIX")
func Parse(config interface{}, prefix ...string) error {
	if prefix == nil {
		prefix = []string{"ECP"}
	}
	_, err := rangeOver(roOption{config, false, prefix[0], ""})
	return err
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
func Default(config interface{}) error {
	_, err := rangeOver(roOption{config, true, "", ""})
	return err
}

// List function will also fill up the value of the environment key
// it the "default" tag has value

// List all the config environments
func List(config interface{}, prefix ...string) []string {
	list := []string{}

	if prefix == nil {
		prefix = []string{"ECP"}
	}
	parentName := prefix[0]

	configValue := toValue(config)
	configType := configValue.Type()
	for index := 0; index < configValue.NumField(); index++ {
		all := getAll(gaOption{configType, configValue, index, parentName})
		if all.sName == "-" || all.kName == "" {
			continue
		}
		switch all.rValue.Kind() {
		case reflect.Struct:
			prefix := GetKey(parentName, all.sName, all.rTag)
			list = append(list, List(all.rValue, prefix)...)
		default:
			if strings.Contains(all.value, " ") {
				all.value = fmt.Sprintf("\"%s\"", all.value)
			}
			list = append(list, fmt.Sprintf("%s=%s", all.kName, all.value))
		}
	}

	return list
}
