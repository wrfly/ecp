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
	Bool         bool          `default:"true"`
	Int64        int64         `default:"666664"`
	Int          int           `default:"-1"`
	Uint         uint          `default:"1"`
	F            float32       `default:"3.14"`
	FloatSlice   []float32     `default:"1.1 2.2 3.3"`
	Float64Slice []float64     `default:"1.1 2.2 3.3"`
	F64          float64       `default:"3.15"`
	Duration     time.Duration `default:"1m"`
	DurationDay  time.Duration `default:"6d"`
	IgnoreMeToo  string        `yaml:"-"`
	Book         string
}

type configType struct {
	LogLevel string    `yaml:"log-level" default:"info"`
	Port     int       `yaml:"port" default:"8080"`
	SliceStr []string  `env:"STRING_SLICE" default:"aa bb cc"`
	SliceInt []int     `env:"INT_SLICE" default:"-1 -2 -3"`
	Sub      subConfig `yaml:"sub"`
	// the default value will not work since the ignote `-`
	SubStruct struct {
		Str       string `default:"skrskr"`
		Int       int64  `default:"111"`
		SubStruct struct {
			Bool bool `default:"true"`
		}
	}
	IgnoreMe string `yaml:"-" default:"ignore me"`

	// should not parse this field when it's a nil pointer
	Nil      *string `default:"nil"`
	NilInt   *int    `default:"1"`
	NilInt8  *int8   `default:"8"`
	NilInt16 *int16  `default:"16"`
	NilInt32 *int32  `default:"32"`
	NilInt64 *int64  `default:"64"`
	NilBool  *bool   `default:"true"`

	// ignore unexported fields
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
		for _, key := range list {
			t.Logf("%s\n", key)
		}
	})

	t.Run("list with user defined prefix", func(t *testing.T) {
		list := List(config, "PPPPREFIX")
		for _, key := range list {
			t.Logf("%s\n", key)
		}
	})

	t.Run("with get key", func(t *testing.T) {
		GetKey = func(parentName, structName string, tag reflect.StructTag) (key string) {
			return strings.ToLower(parentName) + "." + strings.ToLower(structName)
		}
		list := List(config)
		for _, key := range list {
			t.Logf("%s\n", key)
		}

		GetKey = EnvGetKey
	})
}

func TestParse(t *testing.T) {
	envs := map[string]string{
		"INT_SLICE":        "1 2 3",
		"ECP_SUB_INT":      "-2333",
		"ECP_SUB_BOOL":     "true",
		"STRING_SLICE":     "a b c",
		"ECP_SUB_UINT":     "123456789",
		"ECP_SUB_INT64":    "6666",
		"ECP_LOG-LEVEL":    "info",
		"ECP_SUB_DURATION": "10s",
	}

	for k, v := range envs {
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
	case config.Sub.F != 3.14:
	case config.SubStruct.Int != 111:
	case *config.Nil != "":
	default:
		passed = true
	}
	if !passed {
		t.Errorf("%+v", config)
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

	LookupValue = func(field reflect.Value, key string) (value string, exist bool) {
		switch field.Kind() {
		case reflect.String:
			return "string", true
		case reflect.Int:
			return "1", true
		case reflect.Int64:
			return "164", true
		case reflect.Float64:
			return "2.333", true
		case reflect.Bool:
			return "tRuE", true
		}
		return "", false
	}

	if err := Parse(&config); err != nil {
		t.Error(err)
	}
	t.Logf("%+v", config)
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

func TestGet(t *testing.T) {
	config := &configType{
		LogLevel: "debug",
		Port:     999,
		Sub: subConfig{
			Bool: true,
			Book: "1982",
			F64:  2.987,
		},
		SubStruct: struct {
			Str       string `default:"skrskr"`
			Int       int64  `default:"111"`
			SubStruct struct {
				Bool bool `default:"true"`
			}
		}{
			Str: "skrskr",
			SubStruct: struct {
				Bool bool `default:"true"`
			}{
				Bool: true,
			},
		},
	}

	// int
	p, err := GetInt64(config, "port")
	if err != nil {
		t.Fatal(err)
	}
	if p != 999 {
		t.Fatal("!=999")
	}

	// string
	s, err := GetString(config, "sub.Book")
	if err != nil {
		t.Fatal(err)
	}
	if s != "1982" {
		t.Fatal("!=1982")
	}

	// bool
	b, err := GetBool(config, "sub.Bool")
	if err != nil {
		t.Fatal(err)
	}
	if !b {
		t.Fatal("not true")
	}

	// float
	f, err := GetFloat64(config, "sub.F64")
	if err != nil {
		t.Fatal(err)
	}
	if f != 2.987 {
		t.Fatal("not true")
	}

	// sub.sub
	subBool, err := GetBool(config, "SubStruct.SubStruct.Bool")
	if err != nil {
		t.Fatal(err)
	}
	if !subBool {
		t.Fatal("not true")
	}
	subStr, err := GetString(config, "SubStruct.Str")
	if err != nil {
		t.Fatal(err)
	}
	if subStr != "skrskr" {
		t.Fatal("not true")
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
