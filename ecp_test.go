package ecp

import (
	"os"
	"testing"
	"time"
)

type subConfig struct {
	Bool     bool
	Int64    int64
	Int      int
	Uint     uint
	Duration time.Duration
}

type config struct {
	LogLevel string `yaml:"log-level"`
	Slice    []string
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

	conf := config{}
	if err := Parse(&conf, "PREFIX"); err != nil {
		t.Error(err)
	}
	t.Logf("%+v", conf)
}
