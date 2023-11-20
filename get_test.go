package ecp

import (
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	config := &configType{
		LogLevel: "debug",
		Port:     999,
		Sub: subConfig{
			Int:   6464,
			Int8:  8,
			Int16: 16,
			Int32: 32,
			Int64: 64,
			F32:   -3.14,
			F64:   2.987,
			Bool:  true,
			Book:  "1984",
		},
		SubStruct: struct {
			Str       string "default:\"skrskr\""
			Int       int64  "default:\"111\""
			SubStruct struct {
				Bool bool "default:\"true\""
				Int  int  "default:\"123\""
			}
		}{
			Str: "skrskr",
			SubStruct: struct {
				Bool bool `default:"true"`
				Int  int  "default:\"123\""
			}{
				Bool: true,
			},
		},
	}

	toEnvName := func(s string) string { return strings.ToUpper(strings.ReplaceAll(s, ".", "_")) }

	// int
	if p, err := GetInt64(config, toEnvName("port")); err != nil {
		t.Fatal(err)
	} else if p != int64(config.Port) {
		t.Fatal("wrong port")
	}
	if p, err := GetInt64(config, "int"); err != nil {
		t.Fatal(err)
	} else if p != int64(config.Sub.Int) {
		t.Fatal("wrong sub.Int")
	}
	if p, err := GetInt64(config, toEnvName("sub.Int8")); err != nil {
		t.Fatal(err)
	} else if p != int64(config.Sub.Int8) {
		t.Fatal("wrong sub.Int8")
	}
	if p, err := GetInt64(config, toEnvName("sub.Int16")); err != nil {
		t.Fatal(err)
	} else if p != int64(config.Sub.Int16) {
		t.Fatal("wrong sub.Int16")
	}
	if p, err := GetInt64(config, toEnvName("sub.Int32")); err != nil {
		t.Fatal(err)
	} else if p != int64(config.Sub.Int32) {
		t.Fatal("wrong sub.Int32")
	}
	if p, err := GetInt64(config, toEnvName("sub.Int64")); err != nil {
		t.Fatal(err)
	} else if p != int64(config.Sub.Int64) {
		t.Fatal("wrong sub.Int64")
	}

	// string
	s, err := GetString(config, toEnvName("sub.Book"))
	if err != nil {
		t.Fatal(err)
	}
	if s != config.Sub.Book {
		t.Fatal("wrong config.Sub.Book")
	}

	// bool
	b, err := GetBool(config, toEnvName("sub.Bool"))
	if err != nil {
		t.Fatal(err)
	}
	if !b {
		t.Fatal("not true")
	}

	// float
	if f, err := GetFloat64(config, toEnvName("sub.F64")); err != nil {
		t.Fatal(err)
	} else if f != config.Sub.F64 {
		t.Fatal("wrong config.Sub.F64")
	}
	if f, err := GetFloat64(config, toEnvName("sub.F32")); err != nil {
		t.Fatal(err)
	} else if float32(f) != config.Sub.F32 {
		t.Fatalf("wrong config.Sub.F32 %v", f)
	}

	// sub.sub
	subBool, err := GetBool(config, toEnvName("SubStruct.SubStruct.Bool"))
	if err != nil {
		t.Fatal(err)
	}
	if !subBool {
		t.Fatal("not true")
	}
	subStr, err := GetString(config, toEnvName("SubStruct.Str"))
	if err != nil {
		t.Fatal(err)
	}
	if subStr != config.SubStruct.Str {
		t.Fatal("wrong config.SubStruct.Str")
	}

}
