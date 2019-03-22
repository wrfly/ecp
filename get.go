package ecp

import (
	"fmt"
	"reflect"
)

func getValue(config interface{}, keyName string) (reflect.Value, error) {
	v, err := rangeOver(roOption{config, false, "", keyName})
	if err != nil {
		return reflect.Value{}, err
	}

	if !v.IsValid() {
		return reflect.Value{}, fmt.Errorf("invalid type %v", v)
	}
	return v, err
}

// Get the value of the keyName in that struct
func Get(config interface{}, keyName string) (interface{}, error) {
	v, err := getValue(config, keyName)
	if err != nil {
		return nil, err
	}

	return v.Interface(), nil
}

// GetBool returns bool
func GetBool(config interface{}, keyName string) (bool, error) {
	v, err := getValue(config, keyName)
	if err != nil {
		return false, err
	}

	if vv, ok := v.Interface().(bool); ok {
		return vv, nil
	}
	return false, fmt.Errorf("value is not bool, it's %s", v.Kind())
}

// GetInt64 returns int64
func GetInt64(config interface{}, keyName string) (int64, error) {
	v, err := getValue(config, keyName)
	if err != nil {
		return -1, err
	}
	i := v.Interface()

	if vv, ok := i.(int); ok {
		return int64(vv), nil
	} else if vv, ok := i.(int8); ok {
		return int64(vv), nil
	} else if vv, ok := i.(int16); ok {
		return int64(vv), nil
	} else if vv, ok := i.(int32); ok {
		return int64(vv), nil
	} else if vv, ok := i.(int64); ok {
		return vv, nil
	}

	return -1, fmt.Errorf("value is %s", v.Kind())
}

// GetString returns string
func GetString(config interface{}, keyName string) (string, error) {
	v, err := getValue(config, keyName)
	if err != nil {
		return "", err
	}

	if vv, ok := v.Interface().(string); ok {
		return vv, nil
	}
	return "", fmt.Errorf("value is not string, it's %s", v.Kind())
}

// GetFloat64 returns float64
func GetFloat64(config interface{}, keyName string) (float64, error) {
	v, err := getValue(config, keyName)
	if err != nil {
		return -1, err
	}
	i := v.Interface()
	if vv, ok := i.(float32); ok {
		return float64(vv), nil
	}
	if vv, ok := i.(float64); ok {
		return vv, nil
	}
	return -1, fmt.Errorf("value is %s", v.Kind())
}
