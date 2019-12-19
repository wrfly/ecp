package ecp

import (
	"os"
	"reflect"
	"strings"
)

type (
	// GetKeyFunc how to get the key name
	GetKeyFunc func(parentName, structName string, tag reflect.StructTag) (key string)
	// LookupValueFunc returns the value and whether exist
	LookupValueFunc func(field reflect.Value, key string) (value string, exist bool)
	// IgnoreKeyFunc ignore this key
	IgnoreKeyFunc func(field reflect.Value, key string) bool
	// LookupKeyFunc returns the key name
	LookupKeyFunc func(original, prefix, structName string) string
)

// default functions
var (
	getKeyFromEnv = func(pName, sName string, tag reflect.StructTag) string {
		if e := tag.Get("env"); e != "" {
			return strings.Split(e, ",")[0]
		}
		if pName == "" {
			return strings.ToUpper(sName)
		}
		return strings.ToUpper(pName + "_" + sName)
	}
	lookupValueFromEnv = func(field reflect.Value, key string) (string, bool) {
		return os.LookupEnv(key)
	}
	ignoreEnvKey = func(field reflect.Value, key string) bool {
		return key == "-"
	}
	lookupKey = func(original, prefix, structName string) string {
		if prefix == "" {
			return structName
		}
		return prefix + "." + structName
	}
)
