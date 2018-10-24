// Package ecp can help you convert environments into configurations
// it's an environment config parser
package ecp

import (
	"fmt"
	"os"
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

func getEnvName(configType reflect.Type, configValue reflect.Value, i int,
	prefix string) (reflect.Value, string, string, string) {
	field := configValue.Field(i)
	structName := configType.Field(i).Name
	tag := configType.Field(i).Tag

	if y := tag.Get("yaml"); y != "" {
		structName = strings.Split(y, ",")[0]
	}

	envName := strings.ToUpper(prefix + "_" + structName)
	if e := tag.Get("env"); e != "" {
		envName = strings.Split(e, ",")[0]
	}

	return field, structName, envName, tag.Get("default")

}

func parseSlice(v string, field reflect.Value) error {
	if v == "" {
		return nil
	}
	// either space nor commas is perfect, but I think space is better
	// since it's more natural: fmt.Println([]int{1, 2, 3}) = [1 2 3]
	stringSlice := strings.Split(v, " ") // split by space
	field.Set(reflect.MakeSlice(field.Type(), len(stringSlice), cap(stringSlice)))

	switch field.Type().String() {
	case "[]string":
		field.Set(reflect.ValueOf(stringSlice))
	case "[]int":
		intSlice := []int{}
		for _, s := range stringSlice {
			i, err := strconv.Atoi(s)
			if err != nil {
				return err
			}
			intSlice = append(intSlice, i)
		}
		field.Set(reflect.ValueOf(intSlice))
	case "[]float32":
		floatSlice := []float32{}
		for _, s := range stringSlice {
			i, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return err
			}
			floatSlice = append(floatSlice, float32(i))
		}
		field.Set(reflect.ValueOf(floatSlice))

	}
	return nil
}

func rangeOver(config interface{}, parseDefault bool, prefix string) error {
	configValue := toValue(config)
	configType := configValue.Type()
	for i := 0; i < configValue.NumField(); i++ {
		field, structName, envName, d := getEnvName(configType, configValue, i, prefix)

		// skip this filed if its name is `-`
		if structName == "-" || envName == "-" {
			continue
		}

		v, exist := os.LookupEnv(envName)
		if parseDefault || !exist {
			v = d
		}

		if !field.CanAddr() {
			continue
		}

		kind := field.Kind()
		if v == "" && kind != reflect.Struct {
			continue
		}

		switch kind {
		case reflect.String:
			if field.String() != "" && !exist {
				break
			}
			field.SetString(v)

		case reflect.Float32, reflect.Float64:
			if field.Float() != 0 && !exist {
				break
			}
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return fmt.Errorf("convert %s error: %s", envName, err)
			}
			field.SetFloat(f)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if field.Int() != 0 && !exist {
				break
			}
			// since duration is int64 too, parse it first
			// if the duration contains `d` (day), we should support it
			// fix #6
			if strings.Contains(v, "d") {
				day := v[:len(v)-1]
				dayN, err := strconv.Atoi(day)
				if err != nil {
					return fmt.Errorf("convert %s error: %s", envName, err)
				}
				v = fmt.Sprintf("%dh", dayN*24)
			}
			d, err := time.ParseDuration(v)
			if err == nil {
				field.SetInt(int64(d))
				break
			}
			vint, err := strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("convert %s error: %s", envName, err)
			}
			field.SetInt(int64(vint))

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if field.Uint() != 0 && !exist {
				break
			}
			vint, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return fmt.Errorf("convert %s error: %s", envName, err)
			}
			field.SetUint(vint)

		case reflect.Bool:
			b, err := strconv.ParseBool(v)
			if err != nil {
				return err
			}
			if !exist && field.Bool() != b {
				break
			}
			field.SetBool(b)

		case reflect.Slice:
			if !field.IsNil() && !exist {
				break
			}
			if err := parseSlice(v, field); err != nil {
				return err
			}

		case reflect.Struct:
			pref := strings.ToUpper(prefix + "_" + structName)
			if err := rangeOver(field, parseDefault, pref); err != nil {
				return err
			}

		}
	}
	return nil
}

// note Parse function will overwrite the existing value if there is a
// envitonment configration matched with the struct name or the "env" tag
// name.

// Parse the configuration through environments starting with the prefix
// or you can ignore the prefix and the default prefix key will be `ECP`
// ecp.Parse(&config) or ecp.Parse(&config, "PREFIX")
func Parse(config interface{}, prefix ...string) error {
	if prefix == nil {
		prefix = []string{"ECP"}
	}
	return rangeOver(config, false, prefix[0])
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
// config key has already been set nomatter whether it has a default tag.
// And the default value will be nil (nil of the type) if the "default" tag is
// empty.

// Default set config with its default value
func Default(config interface{}) error {
	return rangeOver(config, true, "")
}

// List function will also fill up the value of the envitonment key
// it the "default" tag has value

// List all the config environments
func List(config interface{}, prefix ...string) []string {
	list := []string{}

	if prefix == nil {
		prefix = []string{"ECP"}
	}

	configValue := toValue(config)
	configType := configValue.Type()
	for i := 0; i < configValue.NumField(); i++ {
		field, structName, envName, d := getEnvName(configType, configValue, i, prefix[0])
		if structName == "-" || envName == "" {
			continue
		}
		switch field.Kind() {
		case reflect.Struct:
			list = append(list,
				List(field, strings.Join([]string{prefix[0], structName}, "_"))...)
		default:
			if strings.Contains(d, " ") {
				d = fmt.Sprintf("\"%s\"", d)
			}
			list = append(list, envName+"="+d)
		}
	}

	return list
}
