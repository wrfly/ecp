package ecp

import (
	"fmt"
	"reflect"
)

func (e *ecp) getValue(config interface{}, keyName string) (reflect.Value, error) {
	v, err := e.rangeOver(roOption{config, false, "", keyName})
	if err != nil {
		return reflect.Value{}, err
	}

	if !v.IsValid() {
		return reflect.Value{}, fmt.Errorf("key %s not found", keyName)
	}
	return v, nil
}

func (e *ecp) Get(config interface{}, keyName string) (interface{}, error) {
	v, err := e.getValue(config, keyName)
	if err != nil {
		return nil, err
	}

	return v.Interface(), nil
}

func (e *ecp) GetBool(config interface{}, keyName string) (bool, error) {
	v, err := e.getValue(config, keyName)
	if err != nil {
		return false, err
	}

	if vv, ok := v.Interface().(bool); ok {
		return vv, nil
	}
	return false, fmt.Errorf("value is not bool, it's %s", v.Kind())
}

func (e *ecp) GetInt64(config interface{}, keyName string) (int64, error) {
	v, err := e.getValue(config, keyName)
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

func (e *ecp) GetString(config interface{}, keyName string) (string, error) {
	v, err := e.getValue(config, keyName)
	if err != nil {
		return "", err
	}

	if vv, ok := v.Interface().(string); ok {
		return vv, nil
	}
	return "", fmt.Errorf("value is not string, it's %s", v.Kind())
}

func (e *ecp) GetFloat64(config interface{}, keyName string) (float64, error) {
	v, err := e.getValue(config, keyName)
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

// Get the value of the keyName in that struct
func Get(config interface{}, keyName string) (interface{}, error) {
	return globalEcp.Get(config, keyName)
}

// GetBool returns bool
func GetBool(config interface{}, keyName string) (bool, error) {
	return globalEcp.GetBool(config, keyName)
}

// GetInt64 returns int64
func GetInt64(config interface{}, keyName string) (int64, error) {
	return globalEcp.GetInt64(config, keyName)
}

// GetString returns string
func GetString(config interface{}, keyName string) (string, error) {
	return globalEcp.GetString(config, keyName)
}

// GetFloat64 returns float64
func GetFloat64(config interface{}, keyName string) (float64, error) {
	return globalEcp.GetFloat64(config, keyName)
}
