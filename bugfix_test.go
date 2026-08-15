package ecp

import (
	"os"
	"strings"
	"testing"
	"time"
)

// regression tests for the bugs fixed in this round

type bugFixConfig struct {
	Port            int     `env:"FIX_PORT"`
	Int8F           int8    `env:"FIX_INT8"`
	Uint8F          uint8   `env:"FIX_UINT8"`
	Int64SL         []int64 `env:"FIX_INT64SL"`
	DurSL           []time.Duration
	PtrSL           *[]string `env:"FIX_PTRSL"`
	PtrInt          *int      `env:"FIX_PTRINT"`
	unexportedField string
}

func withEnv(t *testing.T, key, value string) {
	os.Setenv(key, value)
	t.Cleanup(func() { os.Unsetenv(key) })
}

// a plain int field must not be parsed as a duration: "10s" used to
// silently set the field to 10e9 (nanoseconds)
func TestParseIntFieldRejectsDurationValue(t *testing.T) {
	c := &bugFixConfig{}
	withEnv(t, "FIX_PORT", "10s")
	if err := Parse(c); err == nil {
		t.Errorf("expected error for duration value on int field, got port=%d", c.Port)
	}
}

// out-of-range values must error out instead of being silently truncated
func TestParseIntFieldOverflow(t *testing.T) {
	c := &bugFixConfig{}
	withEnv(t, "FIX_INT8", "1000")
	if err := Parse(c); err == nil {
		t.Errorf("expected error for int8 overflow, got %d", c.Int8F)
	}

	c = &bugFixConfig{}
	withEnv(t, "FIX_UINT8", "300")
	if err := Parse(c); err == nil {
		t.Errorf("expected error for uint8 overflow, got %d", c.Uint8F)
	}
}

// []int64 and duration-like values used to panic with
// "reflect.Set: value of type []time.Duration is not assignable to type []int64"
func TestParseInt64SliceRejectsDurationValues(t *testing.T) {
	c := &bugFixConfig{}
	withEnv(t, "FIX_INT64SL", "1h 2h")
	if err := Parse(c); err == nil {
		t.Errorf("expected error for duration values in []int64, got %v", c.Int64SL)
	}
}

// []time.Duration and plain numbers used to panic the other way around
func TestParseDurationSliceRejectsPlainNumbers(t *testing.T) {
	c := &bugFixConfig{}
	withEnv(t, "DURSL", "1 2")
	if err := Parse(c); err == nil {
		t.Errorf("expected error for plain numbers in []time.Duration, got %v", c.DurSL)
	}
}

// pointer-to-slice fields failed with "field is not addressable"
func TestParsePointerToSlice(t *testing.T) {
	c := &bugFixConfig{}
	withEnv(t, "FIX_PTRSL", "a b c")
	if err := Parse(c); err != nil {
		t.Fatal(err)
	}
	if c.PtrSL == nil || len(*c.PtrSL) != 3 || (*c.PtrSL)[1] != "b" {
		t.Errorf("parse *[]string failed: %v", c.PtrSL)
	}
}

// an existing environment value must overwrite a non-nil pointer field
func TestParseEnvOverridesNonNilPointer(t *testing.T) {
	c := &bugFixConfig{}
	old := 42
	c.PtrInt = &old
	withEnv(t, "FIX_PTRINT", "7")
	if err := Parse(c); err != nil {
		t.Fatal(err)
	}
	if *c.PtrInt != 7 {
		t.Errorf("env should override non-nil pointer, got %d", *c.PtrInt)
	}
}

func TestParseScientificNegativeExponent(t *testing.T) {
	if _, err := parseScientific("1e-3"); err == nil {
		t.Error("negative exponent should be an error, not silently become 1")
	}
}

func TestParseScientificCommaAndExponent(t *testing.T) {
	r, err := parseScientific("1,000e3")
	if err != nil {
		t.Fatal(err)
	}
	if r != "1000000" {
		t.Errorf("parse 1,000e3 error, result=%s", r)
	}
}

func TestGetMissingKeyError(t *testing.T) {
	_, err := Get(&bugFixConfig{}, "NO-SUCH-KEY")
	if err == nil || !strings.Contains(err.Error(), "NO-SUCH-KEY") {
		t.Errorf("expected 'key NO-SUCH-KEY not found', got %v", err)
	}
}

// Get on an unexported field used to panic inside reflect.Value.Interface
func TestGetUnexportedField(t *testing.T) {
	if _, err := Get(&bugFixConfig{}, "UNEXPORTEDFIELD"); err == nil {
		t.Error("expected error for unexported field")
	}
}

// List used to include unexported fields, unlike Parse which ignores them
func TestListSkipsUnexportedFields(t *testing.T) {
	list := List(bugFixConfig{})
	for _, item := range list {
		if strings.HasPrefix(item, "UNEXPORTEDFIELD=") {
			t.Errorf("unexported field listed: %s", item)
		}
	}
}

// Parse(&c) where c is already a pointer used to panic with
// "reflect: call of reflect.Value.NumField on ptr Value"
func TestParseDoublePointer(t *testing.T) {
	c := &bugFixConfig{}
	if err := Parse(&c); err != nil {
		t.Errorf("Parse(&c) with c being a pointer should work, got %v", err)
	}
	if c == nil {
		t.Error("config became nil")
	}
}

// time.Duration fields keep accepting duration syntax after the fixes
func TestParseDurationFieldStillWorks(t *testing.T) {
	c := &bugFixConfig{}
	withEnv(t, "DURSL", "1h 2m 3d")
	if err := Parse(c); err != nil {
		t.Fatal(err)
	}
	if len(c.DurSL) != 3 || c.DurSL[2] != 72*time.Hour {
		t.Errorf("parse duration slice failed: %v", c.DurSL)
	}
}
