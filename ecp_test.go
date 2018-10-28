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
	Bool        bool          `default:"true"`
	Int64       int64         `default:"666664"`
	Int         int           `default:"-1"`
	Uint        uint          `default:"1"`
	F           float32       `default:"3.14"`
	FloatSlice  []float32     `default:"1.1 2.2 3.3"`
	F64         float64       `default:"3.15"`
	Duration    time.Duration `default:"1m"`
	DurationDay time.Duration `default:"6d"`
	IgnoreMeToo string        `yaml:"-"`
}

type configType struct {
	LogLevel string    `yaml:"log-level" default:"info"`
	Port     int       `yaml:"port" default:"8080"`
	SliceStr []string  `env:"STRING_SLICE" default:"aa bb cc"`
	SliceInt []int     `env:"INT_SLICE" default:"-1 -2 -3"`
	Sub      subConfig `yaml:"sub"`
	// the default value will not work since the ignote `-`
	IgnoreMe string `yaml:"-" default:"ignore me"`
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

	t.Log("set environment")
	for k, v := range envs {
		os.Setenv(k, v)
		if strings.Contains(v, " ") {
			v = fmt.Sprintf("\"%s\"", v)
		}
		t.Logf("export %s=%s", k, v)
	}

	config := configType{
		LogLevel: "debug",
		Port:     999,
	}
	if err := Parse(&config); err != nil {
		t.Error(err)
	}
	if config.Sub.Duration != time.Second*10 {
		t.Error("parse time duration failed")
	}
	t.Logf("%+v", config)
}

func TestDefault(t *testing.T) {
	config := configType{
		LogLevel: "debug",
		Port:     999,
	}
	if err := Default(&config); err != nil {
		t.Error(err)
	}
	t.Logf("%+v", config)
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
	if err := Parse(&config1); err != nil {
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
	if err := Parse(&config2); err != nil {
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
