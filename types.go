package ecp

import (
	"os"
	"reflect"
	"strings"
)

type (
	// BuildKeyFunc how to get the key name
	BuildKeyFunc func(structure, field string, tag reflect.StructTag) string
	// LookupValueFunc returns the value and whether exist
	LookupValueFunc func(key string) (value string, exist bool)
	// SetValueFunc set the field value and returns whether this filed is set by this function
	SetValueFunc func(tag reflect.StructTag, field reflect.Value, val string) bool
)

const space = " "

// default functions
var (
	buildKeyFromEnv = func(structure, field string, tag reflect.StructTag) (key string) {
		if e := tag.Get("env"); e != "" {
			return e
		}
		if structure == "" {
			key = field
		} else {
			key = structure + "_" + field
		}
		return strings.ToUpper(key)
	}

	lookupValueFromEnv = func(key string) (string, bool) { return os.LookupEnv(key) }
)
