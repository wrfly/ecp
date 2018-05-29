package ecp

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	debug    = false
	duration = reflect.TypeOf(time.Second * 1).Kind()
)

func convertV(conf interface{}) reflect.Value {
	confV, ok := conf.(reflect.Value)
	if !ok {
		confV = reflect.Indirect(reflect.ValueOf(conf))
	}
	return confV
}

func getEnvName(confT reflect.Type, confV reflect.Value, i int, prefix ...string) (reflect.Value, string, string) {
	field := confV.Field(i)
	sName := confT.Field(i).Name
	if y := confT.Field(i).Tag.Get("yaml"); y != "" {
		sName = y
	}
	return field, sName, strings.ToUpper(strings.Join(append(prefix, sName), "_"))

}

func Parse(conf interface{}, prefix ...string) error {
	confV := convertV(conf)
	confT := confV.Type()
	for i := 0; i < confV.NumField(); i++ {
		field, sName, envName := getEnvName(confT, confV, i, prefix...)
		if debug {
			fmt.Printf("got env config %s\n", envName)
		}
		v := os.Getenv(envName)
		if v == "" && field.Kind() != reflect.Struct {
			continue
		}
		if debug {
			fmt.Printf("set %s to %s\n", envName, v)
		}
		switch field.Kind() {
		case reflect.String:
			field.SetString(v)
		case reflect.Float32, reflect.Float64:
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return fmt.Errorf("convert %s error: %s\n", envName, err)
			}
			field.SetFloat(f)
		case duration:
			// since duration is int64 too, parse it first
			d, err := time.ParseDuration(v)
			if err == nil {
				field.SetInt(int64(d))
				continue
			}
			fallthrough
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			vint, err := strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("convert %s error: %s\n", envName, err)
			}
			field.SetInt(int64(vint))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			vint, err := strconv.ParseUint(v, 10, 0)
			if err != nil {
				return fmt.Errorf("convert %s error: %s\n", envName, err)
			}
			field.SetUint(vint)
		case reflect.Bool:
			b, err := strconv.ParseBool(v)
			if err != nil {
				return err
			}
			field.SetBool(b)
		case reflect.Slice:
			fmt.Printf("%s is a slice\n", envName)
			// field.Sli
		case reflect.Struct:
			if err := Parse(field, strings.Join(append(prefix, sName), "_")); err != nil {
				return err
			}
		}
	}
	return nil
}

func Default(conf interface{}) error {
	confV := convertV(conf)
	confT := confV.Type()
	for i := 0; i < confV.NumField(); i++ {
		var (
			fieldStruct = confT.Field(i)
			field       = confV.Field(i)
			d           = fieldStruct.Tag.Get("default")
		)
		switch field.Kind() {
		case reflect.String:
			field.SetString(d)
		case reflect.Int:
			vint, err := strconv.Atoi(d)
			if err != nil {
				return fmt.Errorf("convert %s error: %s\n", fieldStruct.Name, err)

			}
			field.SetInt(int64(vint))
		case reflect.Bool:
			if strings.ToLower(d) == "true" {
				field.SetBool(true)
			} else {
				field.SetBool(false)
			}
		case duration:
			d, err := time.ParseDuration(d)
			if err != nil {
				return fmt.Errorf("parse duration %s error: %s\n", fieldStruct.Name, err)

			}
			field.SetInt(int64(d))
		case reflect.Struct:
			if err := Default(field); err != nil {
				return err
			}
		}
	}
	return nil
}

func List(conf interface{}, prefix string) []string {
	list := []string{}

	confV := convertV(conf)
	confT := confV.Type()
	for i := 0; i < confV.NumField(); i++ {
		field, sName, envName := getEnvName(confT, confV, i, prefix)
		switch field.Kind() {
		case reflect.Struct:
			list = append(list,
				List(field, strings.Join([]string{prefix, sName}, "_"))...)
		default:
			list = append(list, envName+"=")
		}
	}

	return list
}
