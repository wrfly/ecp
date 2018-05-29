package ecp

import (
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
	FloatSlice []float32     `default:"1.1 2.2 3.3"`
	F64        float64       `default:"3.15"`
	Duration   time.Duration `default:"1m"`
}

type config struct {
	LogLevel string    `yaml:"log-level"`
	SliceStr []string  `env:"STRING_SLICE" default:"aa bb cc"`
	SliceInt []int     `env:"INT_SLICE" default:"-1 -2 -3"`
	Sub      subConfig `yaml:"sub"`
}

func TestList(t *testing.T) {
	debug = true

	conf := config{}
	list := List(conf, "PREFIX")
	for _, key := range list {
		t.Logf("%s", key)
	}
}

func TestParse(t *testing.T) {
	debug = true

	os.Setenv("PREFIX_LOG-LEVEL", "info")
	os.Setenv("PREFIX_SLICE", "1 2 3 4")
	os.Setenv("PREFIX_SUB_BOOL", "true")
	os.Setenv("PREFIX_SUB_INT64", "6666")
	os.Setenv("PREFIX_SUB_INT", "-2333")
	os.Setenv("PREFIX_SUB_UINT", "123456789")
	os.Setenv("PREFIX_SUB_DURATION", "10s")
	os.Setenv("STRING_SLICE", "a b c")
	os.Setenv("INT_SLICE", "1 2 3")

	conf := config{}
	if err := Parse(&conf, "PREFIX"); err != nil {
		t.Error(err)
	}
	t.Logf("%+v", conf)
}

func TestDefault(t *testing.T) {
	debug = true

	conf := config{}
	if err := Default(&conf); err != nil {
		t.Error(err)
	}
	t.Logf("%+v", conf)
}
