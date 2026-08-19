package ecp

import (
	"fmt"
	"reflect"
)

// getValue looks up a field by the environment key it is bound to. The
// optional prefix must match the one Parse was called with, since that is
// what the key names are built from.
func (e *ECP) getValue(config interface{}, keyName string, prefix ...string) (reflect.Value, error) {
	if len(prefix) == 0 {
		prefix = []string{""}
	}

	v, err := e.rangeOver(roOption{target: config, find: keyName, prefix: prefix[0]})
	if err != nil {
		return reflect.Value{}, err
	}

	if !v.IsValid() {
		return reflect.Value{}, fmt.Errorf("key %s not found", keyName)
	}

	if !v.CanInterface() {
		return reflect.Value{}, fmt.Errorf("bad structure field %s", keyName)
	}
	return v, nil
}

// Get the value of the keyName in that struct
func (e *ECP) Get(config interface{}, keyName string, prefix ...string) (interface{}, error) {
	v, err := e.getValue(config, keyName, prefix...)
	if err != nil {
		return nil, err
	}

	return v.Interface(), nil
}

// GetBool returns bool
func (e *ECP) GetBool(config interface{}, keyName string, prefix ...string) (bool, error) {
	v, err := e.getValue(config, keyName, prefix...)
	if err != nil {
		return false, err
	}

	if v.Kind() == reflect.Bool {
		return v.Bool(), nil
	}
	return false, fmt.Errorf("value is not bool, it's %s", v.Kind())
}

// GetInt64 returns int64
func (e *ECP) GetInt64(config interface{}, keyName string, prefix ...string) (int64, error) {
	v, err := e.getValue(config, keyName, prefix...)
	if err != nil {
		return -1, err
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), nil
	}

	return -1, fmt.Errorf("value is %s", v.Kind())
}

// GetString returns string
func (e *ECP) GetString(config interface{}, keyName string, prefix ...string) (string, error) {
	v, err := e.getValue(config, keyName, prefix...)
	if err != nil {
		return "", err
	}

	if v.Kind() == reflect.String {
		return v.String(), nil
	}
	return "", fmt.Errorf("value is not string, it's %s", v.Kind())
}

// GetFloat64 returns float64
func (e *ECP) GetFloat64(config interface{}, keyName string, prefix ...string) (float64, error) {
	v, err := e.getValue(config, keyName, prefix...)
	if err != nil {
		return -1, err
	}

	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		return v.Float(), nil
	}
	return -1, fmt.Errorf("value is %s", v.Kind())
}

// Get the value of the keyName in that struct
func Get(config interface{}, keyName string, prefix ...string) (interface{}, error) {
	return globalEcp.Get(config, keyName, prefix...)
}

// GetBool returns bool
func GetBool(config interface{}, keyName string, prefix ...string) (bool, error) {
	return globalEcp.GetBool(config, keyName, prefix...)
}

// GetInt64 returns int64
func GetInt64(config interface{}, keyName string, prefix ...string) (int64, error) {
	return globalEcp.GetInt64(config, keyName, prefix...)
}

// GetString returns string
func GetString(config interface{}, keyName string, prefix ...string) (string, error) {
	return globalEcp.GetString(config, keyName, prefix...)
}

// GetFloat64 returns float64
func GetFloat64(config interface{}, keyName string, prefix ...string) (float64, error) {
	return globalEcp.GetFloat64(config, keyName, prefix...)
}
