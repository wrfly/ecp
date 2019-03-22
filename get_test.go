package ecp

import (
	"testing"
)

func TestGet(t *testing.T) {
	config := &configType{
		LogLevel: "debug",
		Port:     999,
		Sub: subConfig{
			Int:   66,
			Int8:  66,
			Int16: 66,
			Int32: 66,
			Int64: 66,
			F32:   -3.14,
			F64:   2.987,
			Bool:  true,
			Book:  "1982",
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
	if p, err := GetInt64(config, "port"); err != nil {
		t.Error(err)
	} else if p != 999 {
		t.Error("!=999")
	}
	if p, err := GetInt64(config, "sub.Int"); err != nil {
		t.Error(err)
	} else if p != 66 {
		t.Error("!=66")
	}
	if p, err := GetInt64(config, "sub.Int8"); err != nil {
		t.Error(err)
	} else if p != 66 {
		t.Error("!=66")
	}
	if p, err := GetInt64(config, "sub.Int16"); err != nil {
		t.Error(err)
	} else if p != 66 {
		t.Error("!=66")
	}
	if p, err := GetInt64(config, "sub.Int32"); err != nil {
		t.Error(err)
	} else if p != 66 {
		t.Error("!=66")
	}
	if p, err := GetInt64(config, "sub.Int64"); err != nil {
		t.Error(err)
	} else if p != 66 {
		t.Error("!=66")
	}

	// string
	s, err := GetString(config, "sub.Book")
	if err != nil {
		t.Error(err)
	}
	if s != "1982" {
		t.Error("!=1982")
	}

	// bool
	b, err := GetBool(config, "sub.Bool")
	if err != nil {
		t.Error(err)
	}
	if !b {
		t.Error("not true")
	}

	// float
	if f, err := GetFloat64(config, "sub.F64"); err != nil {
		t.Error(err)
	} else if f != 2.987 {
		t.Error("not true")
	}
	if f, err := GetFloat64(config, "sub.F32"); err != nil {
		t.Error(err)
	} else if int(f+3.14) != 0 {
		t.Errorf("not true (%d)", int(f+3.14))
	}

	// sub.sub
	subBool, err := GetBool(config, "SubStruct.SubStruct.Bool")
	if err != nil {
		t.Error(err)
	}
	if !subBool {
		t.Error("not true")
	}
	subStr, err := GetString(config, "SubStruct.Str")
	if err != nil {
		t.Error(err)
	}
	if subStr != "skrskr" {
		t.Error("not true")
	}

}
