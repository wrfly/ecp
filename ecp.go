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

func convertV(conf interface{}) reflect.Value {
	confV, ok := conf.(reflect.Value)
	if !ok {
		confV = reflect.Indirect(reflect.ValueOf(conf))
	}
	return confV
}

func getEnvName(confT reflect.Type, confV reflect.Value, i int,
	prefix string) (reflect.Value, string, string, string) {
	field := confV.Field(i)
	sName := confT.Field(i).Name
	tag := confT.Field(i).Tag

	if y := tag.Get("yaml"); y != "" {
		sName = y
	}

	envName := strings.ToUpper(prefix + "_" + sName)
	if e := tag.Get("env"); e != "" {
		envName = e
	}

	return field, sName, envName, tag.Get("default")

}

func parseSlice(v string, field reflect.Value) error {
	if v == "" {
		return nil
	}
	stringSlice := strings.Split(v, ",") // split by commas
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

func rangeOver(conf interface{}, parseDefault bool, prefix string) error {
	confV := convertV(conf)
	confT := confV.Type()
	for i := 0; i < confV.NumField(); i++ {
		field, sName, envName, d := getEnvName(confT, confV, i, prefix)

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
				return fmt.Errorf("convert %s error: %s\n", envName, err)
			}
			field.SetFloat(f)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if field.Int() != 0 && !exist {
				break
			}
			// since duration is int64 too, parse it first
			d, err := time.ParseDuration(v)
			if err == nil {
				field.SetInt(int64(d))
				break
			}
			vint, err := strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("convert %s error: %s\n", envName, err)
			}
			field.SetInt(int64(vint))

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if field.Uint() != 0 && !exist {
				break
			}
			vint, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return fmt.Errorf("convert %s error: %s\n", envName, err)
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
			pref := strings.ToUpper(prefix + "_" + sName)
			if err := rangeOver(field, parseDefault, pref); err != nil {
				return err
			}

		}
	}
	return nil
}

// Parse the conf through environments with the prefix key
func Parse(conf interface{}, prefix string) error {
	return rangeOver(conf, false, prefix)
}

// Default set the conf with its default value
func Default(conf interface{}) error {
	return rangeOver(conf, true, "")
}

// List all the config environment keys
func List(conf interface{}, prefix string) []string {
	list := []string{}

	confV := convertV(conf)
	confT := confV.Type()
	for i := 0; i < confV.NumField(); i++ {
		field, sName, envName, d := getEnvName(confT, confV, i, prefix)
		switch field.Kind() {
		case reflect.Struct:
			list = append(list,
				List(field, strings.Join([]string{prefix, sName}, "_"))...)
		default:
			list = append(list, envName+"="+d)
		}
	}

	return list
}
