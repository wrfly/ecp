package ecp

import (
	"os"
	"reflect"
	"strings"
)

var (
	GetKey      func(parentName, structName string, tag reflect.StructTag) (key string)
	LookupValue func(field reflect.Value, key string) (value string, exist bool)
	IgnoreKey   func(field reflect.Value, key string) bool

	// env get functions
	getKeyFromEnv = func(pName, sName string, tag reflect.StructTag) string {
		if e := tag.Get("env"); e != "" {
			return strings.Split(e, ",")[0]
		}
		return strings.ToUpper(pName + "_" + sName)
	}
	lookupValueFromEnv = func(field reflect.Value, key string) (string, bool) {
		return os.LookupEnv(key)
	}
	ignoreEnvKey = func(field reflect.Value, key string) bool {
		return key == "-"
	}
)

func init() {
	GetKey = getKeyFromEnv
	IgnoreKey = ignoreEnvKey
	LookupValue = lookupValueFromEnv
}
