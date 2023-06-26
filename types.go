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
)

const space = " "

// default functions
var (
	buildKeyFromEnv = func(structure, field string, tag reflect.StructTag) (key string) {
		for _, key := range []string{"env", "yaml", "json"} {
			if e := tag.Get(key); e != "" {
				return strings.Split(e, ",")[0]
			}
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
