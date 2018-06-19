package ecp

import (
	"fmt"
	"os"
	"testing"
	"time"
)

type subConfig struct {
	Bool       bool          `default:"true"`
	Int64      int64         `default:"666664"`
	Int        int           `default:"-1"`
	Uint       uint          `default:"1"`
	F          float32       `default:"3.14"`
	FloatSlice []float32     `default:"1.1,2.2,3.3"`
	F64        float64       `default:"3.15"`
	Duration   time.Duration `default:"1m"`
}

type configType struct {
	LogLevel string    `yaml:"log-level"`
	Port     int       `yaml:"port" default:"8080"`
	SliceStr []string  `env:"STRING_SLICE" default:"aa,bb,cc"`
	SliceInt []int     `env:"INT_SLICE" default:"-1,-2,-3"`
	Sub      subConfig `yaml:"sub"`
}

func TestList(t *testing.T) {
	config := configType{
		LogLevel: "debug",
	}
	t.Run("list with empty prefix", func(t *testing.T) {
		list := List(config)
		for _, key := range list {
			t.Logf("%s", key)
		}
	})

	t.Run("list with user defined prefix", func(t *testing.T) {
		list := List(config, "PPPPREFIX")
		for _, key := range list {
			t.Logf("%s", key)
		}
	})
}

func TestParse(t *testing.T) {
	envs := map[string]string{
		"ECP_SLICE":        "1 2 3 4",
		"INT_SLICE":        "1,2,3",
		"ECP_SUB_INT":      "-2333",
		"ECP_SUB_BOOL":     "true",
		"STRING_SLICE":     "a,b,c",
		"ECP_SUB_UINT":     "123456789",
		"ECP_SUB_INT64":    "6666",
		"ECP_LOG-LEVEL":    "info",
		"ECP_SUB_DURATION": "10s",
	}
	for k, v := range envs {
		fmt.Printf("export %s=%s\n", k, v)
		os.Setenv(k, v)
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
