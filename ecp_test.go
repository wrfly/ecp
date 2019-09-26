package ecp

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

type subConfig struct {
	Bool      bool   `default:"true"`
	BoolSlice []bool `default:"true false true"`

	Book  string   `default:"go-101"`
	Books []string `default:"golang docker rust"`

	Int        int     `default:"1" env:"int"`
	Int8       int8    `default:"8"`
	Int16      int16   `default:"-16"`
	Int32      int32   `default:"32"`
	Int64      int64   `default:"-64"`
	IntSlice   []int   `default:"1 1 -1"`
	Int8Slice  []int8  `default:"8 8 -8"`
	Int16Slice []int16 `default:"-16 16 -16"`
	Int32Slice []int32 `default:"32 32 -32"`
	Int64Slice []int64 `default:"-64 64 -64"`

	Uint        uint     `default:"1"`
	Uint8       uint8    `default:"8"`
	Uint16      uint16   `default:"16"`
	Uint32      uint32   `default:"32"`
	Uint64      uint64   `default:"64"`
	UintSlice   []uint   `default:"1 1 1 1"`
	Uint8Slice  []uint8  `default:"8 8 8 8"`
	Uint16Slice []uint16 `default:"16 16 16 16"`
	Uint32Slice []uint32 `default:"32 32 32 32"`
	Uint64Slice []uint64 `default:"64 64 64 64"`

	F32          float32   `default:"3.14"`
	F64          float64   `default:"3.15"`
	FloatSlice   []float32 `default:"1.1 2.2 3.3"`
	Float64Slice []float64 `default:"1.1 2.2 3.3"`

	Duration    time.Duration `default:"1m"`
	DurationDay time.Duration `default:"6d"`

	IgnoreMeToo string `yaml:"-"`
}

type configType struct {
	// basic types
	LogLevel string   `yaml:"log-level" default:"info"`
	Port     int      `yaml:"port" default:"8080"`
	SliceStr []string `env:"STRING_SLICE" default:"aa bb cc"`
	SliceInt []int    `env:"INT_SLICE" default:"-1 -2 -3"`

	// sub type
	Sub subConfig `yaml:"sub"`

	// sub struct type
	SubStruct struct {
		Str       string `default:"skrskr"`
		Int       int64  `default:"111"`
		SubStruct struct {
			Bool bool `default:"true"`
		}
	}

	// // pointer struct
	// PinterStruct *struct {
	// 	Value string `default:"hey"`
	// } `yaml:"p"`

	// the default value will not work since the ignore `-`
	IgnoreMe string `yaml:"-" default:"ignore me"`

	// should not parse this field when it's a nil pointer
	Nil     *string `default:"nil"`
	NilBool *bool   `default:"true"`

	NilInt   *int   `default:"1"`
	NilInt8  *int8  `default:"8"`
	NilInt16 *int16 `default:"16"`
	NilInt32 *int32 `default:"32"`
	NilInt64 *int64 `default:"64"`

	NilUInt   *uint   `default:"1"`
	NilUInt8  *uint8  `default:"88"`
	NilUInt16 *uint16 `default:"168"`
	NilUInt32 *uint32 `default:"32"`
	NilUInt64 *uint64 `default:"64"`

	PointerFloat   *float32 `default:"1.2"`
	PointerFloat64 *float64 `default:"4.3"`

	Empty string `json:"empty"`

	// ignore unexposed fields
	x struct {
		X int
	}
}

func TestList(t *testing.T) {
	config := configType{
		LogLevel: "debug",
	}
	t.Run("list with empty prefix", func(t *testing.T) {
		list := List(config)
		for _, key := range list[:9] {
			fmt.Printf("%s\n", key)
		}
	})

	t.Run("list with user defined prefix", func(t *testing.T) {
		list := List(config, "PPPPREFIX")
		for _, key := range list[9:18] {
			fmt.Printf("%s\n", key)
		}
	})

	t.Run("with get key", func(t *testing.T) {
		GetKey = func(parentName, structName string, tag reflect.StructTag) (key string) {
			return strings.ToLower(parentName) + "." + strings.ToLower(structName)
		}
		list := List(config)
		for _, key := range list[18:27] {
			fmt.Printf("%s\n", key)
		}

		// reset get key function
		GetKey = getKeyFromEnv
	})
}

func TestParse(t *testing.T) {
	ENV := map[string]string{
		"INT_SLICE":        "1 2 3",
		"ECP_SUB_INT":      "-2333",
		"ECP_SUB_BOOL":     "true",
		"STRING_SLICE":     "a b c",
		"ECP_SUB_UINT":     "123456789",
		"ECP_SUB_INT64":    "6666",
		"ECP_LOG-LEVEL":    "info",
		"ECP_SUB_DURATION": "10s",
		"ECP_NILINT":       "2",
		"ECP_NILINT8":      "9",
		"ECP_NILINT16":     "17",
		// "ECP_P_VALUE":      "yoo",
	}

	for k, v := range ENV {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	config := configType{
		LogLevel: "debug",
		Port:     999,
	}

	if err := Parse(&config); err != nil {
		t.Error(err)
	}
	// check
	if config.Sub.Duration != time.Second*10 {
		t.Error("parse time duration failed")
	}
	if config.Sub.Uint != 123456789 {
		t.Error("parse uint failed")
	}
	if config.NilInt8 == nil {
		t.Error("parse pointer failed")
	}

	if *config.NilInt != 2 {
		t.Error("???", *config.NilInt)
	}

	if *config.NilInt8 != 9 {
		t.Error("???", *config.NilInt8)
	}

	if *config.NilInt16 != 17 {
		t.Error("???", *config.NilInt16)
	}

	if *config.PointerFloat64 != 4.3 {
		t.Error("???")
	}

	// if config.PinterStruct == nil {
	// 	t.Error("???")
	// } else if config.PinterStruct.Value != "yoo" {
	// 	t.Error("???")
	// }
}

func TestDefault(t *testing.T) {
	empty := ""
	_int8 := int8(8)
	_bool := true

	config := configType{
		LogLevel: "debug",
		Port:     999,
		Nil:      &empty,
		NilInt8:  &_int8,
		NilBool:  &_bool,
	}
	if err := Default(&config); err != nil {
		t.Errorf("set default error: %s", err)
	}

	var passed bool
	switch {
	case config.LogLevel != "debug":
	case config.SliceStr[0] != "aa":
	case config.Sub.F32 != 3.14:
	case config.SubStruct.Int != 111:
	case *config.Nil != "":
	case *config.NilInt64 != 64:
	case *config.NilInt8 != 8:
	default:
		passed = true
	}
	if !passed {
		t.Errorf("err config: %+v", config)
	}

	// test pointers
	config.Nil = nil
	config.NilBool = nil
	if err := Default(&config); err != nil {
		t.Errorf("set default error: %s", err)
	}
	if config.Nil == nil {
		t.Errorf("config.Nil is nil pointer")
	} else {
		if *config.Nil != "nil" {
			t.Errorf("config.Nil != `nil`")
		}
	}
	if config.NilBool == nil {
		t.Errorf("config.NilBool is nil pointer")
	} else {
		if !*config.NilBool {
			t.Errorf("config.NilBool != true ")
		}
	}
}

func TestGetKeyLookupValue(t *testing.T) {
	config := configType{}

	GetKey = func(parentName, structName string, tag reflect.StructTag) (key string) {
		return parentName + "." + structName
	}
	defer func() {
		GetKey = getKeyFromEnv
	}()

	LookupValue = func(field reflect.Value, key string) (value string, exist bool) {
		switch field.Kind() {
		case reflect.String:
			return "string", true
		case reflect.Int:
			return "-100", true
		case reflect.Uint32:
			return "32", true
		case reflect.Float64:
			return "-3.1415", true
		case reflect.Bool:
			return "True", true
		}
		return "", false
	}

	if err := Parse(&config); err != nil {
		t.Error(err)
	}
	switch {
	case config.Port != -100:
	case config.LogLevel != "string":
	case config.Sub.Book != "string":
	case !config.Sub.Bool:
	default:
		return
	}
	t.Errorf("parse failed, config: %+v", config)
}

func TestIgnoreFunc(t *testing.T) {
	config1 := configType{}
	if err := Default(&config1); err != nil {
		t.Error(err)
	}

	config2 := configType{}
	IgnoreKey = func(field reflect.Value, key string) bool {
		switch field.Kind() {
		case reflect.Int64, reflect.Int, reflect.Uint:
		case reflect.String:
		default:
			return false
		}
		return true
	}
	if err := Default(&config2); err != nil {
		t.Fatal(err)
	}

	if config1.Port == config2.Port {
		t.Error("not going to happen")
	}

	if config1.LogLevel == config2.LogLevel {
		t.Error("not going to happen")
	}
}

func ExampleParse() {
	type config struct {
		Age  int
		Name string
	}
	c := &config{}
	os.Setenv("ECP_AGE", "10")
	if err := Parse(&c); err != nil {
		panic(err)
	}

	// c.Age=10
	if c.Age != 10 {
		panic("???")
	}
}

func ExampleList() {
	type config struct {
		Age  int
		Name string
	}
	for _, key := range List(config{}) {
		fmt.Printf("env %s", key)
	}

	// env ECP_AGE=
	// env ECP_NAME=
}

func ExampleDefault() {
	type config struct {
		Age      int           `default:"10"`
		Name     string        `default:"wrfly"`
		Duration time.Duration `default:"10d"`
	}
	c := &config{}
	if err := Default(&c); err != nil {
		panic(err)
	}

	// now you'll get a config with
	// `Age=10` and `Name=wrfly`
	if c.Age != 10 || c.Name != "wrfly" || c.Duration != time.Hour*24*10 {
		panic("???")
	}
}
