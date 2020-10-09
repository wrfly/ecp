package ecp

import (
	"reflect"
	"testing"
	"time"
)

func TestParseSlice(t *testing.T) {
	type slices struct {
		String []string
		Bool   []bool

		Int   []int
		Int8  []int8
		Int16 []int16
		Int32 []int32
		Int64 []int64

		UInt   []uint
		UInt8  []uint8
		UInt16 []uint16
		UInt32 []uint32
		UInt64 []uint64

		Float32 []float32
		Float64 []float64

		Times []time.Duration

		Unsupported []*string
		Pointer     *string
		NotSlice    string
	}

	var parseSlice = globalEcp.parseSlice

	s := &slices{}

	t.Run("test duration", func(t *testing.T) {
		v := toValue(s).FieldByName("Times")
		if err := parseSlice("1h 2m 3d", v); err != nil {
			t.Errorf("parse time slice failed: %v", err)
		} else if len(s.Times) != 3 || s.Times[0] != time.Hour {
			t.Errorf("parse time slice failed: %v", s.Times)
		}
	})

	t.Run("test string", func(t *testing.T) {
		v := toValue(s).FieldByName("String")
		if err := parseSlice("a b c", v); err != nil {
			t.Errorf("parse string slice failed: %s", err)
		} else if len(s.String) != 3 || s.String[1] != "b" {
			t.Errorf("parse string slice failed: %v", s.String)
		}

		s = &slices{}
		if parseSlice("", v) != nil || len(s.String) != 0 {
			t.Errorf("test empty string failed: %v", s.String)
		}

	})

	t.Run("test bool", func(t *testing.T) {
		v := toValue(s).FieldByName("Bool")
		if err := parseSlice("trUe tRue True", v); err != nil {
			t.Errorf("parse bool slice failed: %s", err)
		} else if len(s.Bool) != 3 && !s.Bool[1] {
			t.Errorf("parse bool slice failed: %v", s.String)
		}

		if parseSlice("233", v) == nil {
			t.Error("should not")
		}
	})

	t.Run("test int", func(t *testing.T) {
		v := reflect.Value{}
		x := []string{"Int", "Int8", "Int16", "Int32", "Int64"}
		for _, name := range x {
			v = toValue(s).FieldByName(name)
			if err := parseSlice("1 2 3", v); err != nil {
				t.Errorf("parse int slice failed: %s", err)
			}
		}

		switch {
		case len(s.Int) != 3:
		case len(s.Int8) != 3:
		case len(s.Int16) != 3:
		case len(s.Int32) != 3:
		case len(s.Int64) != 3:
		default:
			for _, name := range x {
				v = toValue(s).FieldByName(name)
				if parseSlice("1.1", v) == nil {
					t.Error("should not")
				}
			}
			return
		}

		t.Errorf("parse int slice error")
	})

	t.Run("test uint", func(t *testing.T) {
		v := reflect.Value{}
		x := []string{"UInt", "UInt8", "UInt16", "UInt32", "UInt64"}
		for _, name := range x {
			v = toValue(s).FieldByName(name)
			if err := parseSlice("1 2 3", v); err != nil {
				t.Errorf("parse int slice failed: %s", err)
			}
		}

		switch {
		case len(s.UInt) != 3:
		case len(s.UInt8) != 3:
		case len(s.UInt16) != 3:
		case len(s.UInt32) != 3:
		case len(s.UInt64) != 3:
		default:
			for _, name := range x {
				v = toValue(s).FieldByName(name)
				if parseSlice("-1", v) == nil {
					t.Error("should not")
				}
			}
			return
		}

		t.Errorf("parse int slice error")
	})

	t.Run("test float", func(t *testing.T) {
		v := reflect.Value{}
		for _, name := range []string{"Float32", "Float64"} {
			v = toValue(s).FieldByName(name)
			if err := parseSlice("1.1 2.2 3.3", v); err != nil {
				t.Errorf("parse int slice failed: %s", err)
			}
		}

		switch {
		case len(s.Float32) != 3:
		case len(s.Float64) != 3:
		default:
			for _, name := range []string{"Float32", "Float64"} {
				v = toValue(s).FieldByName(name)
				if parseSlice("1,", v) == nil {
					t.Error("should not")
				}
			}
			return
		}

		t.Errorf("parse float slice error")
	})

	t.Run("test unsupported type", func(t *testing.T) {
		v := toValue(s).FieldByName("Unsupported")
		if err := parseSlice("a b c", v); err == nil {
			t.Errorf("???")
		}
	})

	t.Run("test error", func(t *testing.T) {
		pointer := toValue(s).FieldByName("Pointer")
		if parseSlice("a b c", pointer) == nil {
			t.Errorf("???")
		}
		ns := toValue(s).FieldByName("NotSlice")
		if parseSlice("a b c", ns) == nil {
			t.Errorf("???")
		}

		if parseSlice("a b c", reflect.ValueOf(s.String)) == nil {
			t.Errorf("???")
		}
	})
}
