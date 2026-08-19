package ecp

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// regression tests for the bugs fixed in this round

type (
	level    int
	name     string
	duration time.Duration
)

// named element types used to panic with
// "reflect.Set: value of type []int is not assignable to type []ecp.level"
func TestNamedTypeSlice(t *testing.T) {
	c := &struct {
		Levels []level `default:"1 2 3"`
		Names  []name  `default:"a b"`
	}{}
	if err := Parse(c); err != nil {
		t.Fatal(err)
	}
	if len(c.Levels) != 3 || c.Levels[2] != 3 {
		t.Errorf("parse []level failed: %v", c.Levels)
	}
	if len(c.Names) != 2 || c.Names[1] != "b" {
		t.Errorf("parse []name failed: %v", c.Names)
	}
}

// pointers to named types used to panic with
// "reflect.Set: value of type *int64 is not assignable to type *time.Duration"
func TestNamedTypePointer(t *testing.T) {
	c := &struct {
		Dur *time.Duration `env:"REG_DUR"`
		Lvl *level         `default:"7"`
	}{}
	withEnv(t, "REG_DUR", "10s")
	if err := Parse(c); err != nil {
		t.Fatal(err)
	}
	// *time.Duration accepts the same syntax as time.Duration
	if c.Dur == nil || *c.Dur != 10*time.Second {
		t.Errorf("parse *time.Duration failed: %v", c.Dur)
	}
	if c.Lvl == nil || *c.Lvl != 7 {
		t.Errorf("parse *level failed: %v", c.Lvl)
	}
}

// "1e1000000" used to spend minutes building a one megabyte string
func TestScientificOutOfRange(t *testing.T) {
	done := make(chan error, 1)
	go func() {
		c := &struct {
			N int64 `env:"REG_HUGE"`
		}{}
		done <- Parse(c)
	}()

	os.Setenv("REG_HUGE", "1e1000000")
	defer os.Unsetenv("REG_HUGE")

	select {
	case err := <-done:
		if err == nil {
			t.Error("expected an out of range error")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("parsing a huge exponent took too long")
	}
}

// "1.5e3" is 1500, it used to expand to the unparsable "1.5000"
func TestScientificDecimalMantissa(t *testing.T) {
	r, err := parseScientific("1.5e3")
	if err != nil {
		t.Fatal(err)
	}
	if r != "1500" {
		t.Errorf("parse 1.5e3 error, result=%s", r)
	}
	if _, err := parseScientific("1.5e0"); err == nil {
		t.Error("1.5e0 is not an integer, it should be an error")
	}
}

// a "-" tag means "ignore this field", Parse used to honour it in List
// only and still filled the field in
func TestIgnoreTag(t *testing.T) {
	c := &struct {
		Yaml string `yaml:"-" default:"nope"`
		JSON string `json:"-" default:"nope"`
		Env  string `env:"-" default:"nope"`
	}{}
	// the bogus keys the ignored fields used to be bound to
	withEnv(t, "-", "leaked")

	if err := Parse(c); err != nil {
		t.Fatal(err)
	}
	if c.Yaml != "" || c.JSON != "" || c.Env != "" {
		t.Errorf("ignored fields were parsed: %+v", c)
	}
	if list := List(*c); len(list) != 0 {
		t.Errorf("ignored fields were listed: %v", list)
	}
}

// Parse(config) instead of Parse(&config) used to report success after
// filling in exactly nothing
func TestParseRequiresPointer(t *testing.T) {
	type conf struct {
		S string `default:"x"`
	}
	if err := Parse(conf{}); err == nil {
		t.Error("expected an error when config is not a pointer")
	}
	c := conf{}
	if err := Parse(&c); err != nil || c.S != "x" {
		t.Errorf("Parse(&c) failed: %v %+v", err, c)
	}
}

// a *struct is a config section, its fields used to be skipped entirely
func TestPointerSection(t *testing.T) {
	type sub struct {
		Value string `default:"hey"`
		Port  int    `default:"80"`
	}
	type conf struct {
		P     *sub `yaml:"p"`
		Empty *sub `yaml:"empty"`
	}

	t.Run("nil section gets filled", func(t *testing.T) {
		withEnv(t, "P_VALUE", "yoo")
		c := &conf{}
		if err := Parse(c); err != nil {
			t.Fatal(err)
		}
		if c.P == nil {
			t.Fatal("nil section was not filled")
		}
		if c.P.Value != "yoo" || c.P.Port != 80 {
			t.Errorf("section not parsed: %+v", *c.P)
		}
	})

	t.Run("existing section is walked into", func(t *testing.T) {
		withEnv(t, "P_VALUE", "yoo")
		c := &conf{P: &sub{Port: 8080}}
		if err := Parse(c); err != nil {
			t.Fatal(err)
		}
		if c.P.Value != "yoo" || c.P.Port != 8080 {
			t.Errorf("section not parsed: %+v", *c.P)
		}
	})

	t.Run("listed with the section prefix", func(t *testing.T) {
		list := strings.Join(List(conf{}), " ")
		for _, want := range []string{"P_VALUE=hey", "P_PORT=80", "EMPTY_VALUE=hey"} {
			if !strings.Contains(list, want) {
				t.Errorf("missing %s in %s", want, list)
			}
		}
	})
}

// a self referencing type used to be impossible to describe, walking into
// pointer sections must not recurse until the stack blows up
func TestCyclicType(t *testing.T) {
	type node struct {
		Name string `default:"n"`
		Next *node
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		c := &node{}
		if err := Parse(c); err != nil {
			t.Error(err)
		}
		if c.Name != "n" {
			t.Errorf("name not parsed: %+v", c)
		}
		if c.Next != nil {
			t.Errorf("cyclic section should stay nil, got %+v", c.Next)
		}
		_ = List(node{})
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("cyclic type did not terminate")
	}
}

// every other kind keeps a value set by the caller, bool used to be reset
// to false by a `default:"false"` tag
func TestBoolDefaultKeepsCallerValue(t *testing.T) {
	c := &struct {
		Feature bool `env:"REG_FEATURE" default:"false"`
	}{Feature: true}
	if err := Parse(c); err != nil {
		t.Fatal(err)
	}
	if !c.Feature {
		t.Error("default:\"false\" overwrote the value set by the caller")
	}

	// an environment value still wins
	withEnv(t, "REG_FEATURE", "false")
	if err := Parse(c); err != nil {
		t.Fatal(err)
	}
	if c.Feature {
		t.Error("the environment value should have won")
	}
}

// kinds that cannot be filled from a string used to be ignored without a
// word, now they are reported and never listed
func TestUnsupportedKinds(t *testing.T) {
	c := &struct {
		M map[string]string `env:"REG_MAP"`
	}{}
	withEnv(t, "REG_MAP", "a=1")
	if err := Parse(c); err == nil {
		t.Error("expected an error for a map field with a value")
	}

	arr := &struct {
		A [3]int `env:"REG_ARR"`
	}{}
	withEnv(t, "REG_ARR", "1 2 3")
	if err := Parse(arr); err == nil {
		t.Error("expected an error for an array field with a value")
	}

	if list := List(*c); len(list) != 0 {
		t.Errorf("unsupported kinds should not be listed: %v", list)
	}
}

// "a  b" used to be split into three elements, the empty one making every
// numeric slice fail to parse
func TestSliceRepeatedSeparators(t *testing.T) {
	c := &struct {
		S []string `env:"REG_S"`
		I []int    `env:"REG_I"`
	}{}
	withEnv(t, "REG_S", " a  b ")
	withEnv(t, "REG_I", "1  2")
	if err := Parse(c); err != nil {
		t.Fatal(err)
	}
	if len(c.S) != 2 || c.S[1] != "b" {
		t.Errorf("parse []string failed: %q", c.S)
	}
	if len(c.I) != 2 || c.I[1] != 2 {
		t.Errorf("parse []int failed: %v", c.I)
	}
}

// a default value with a quote in it used to produce a broken listing
func TestListQuoting(t *testing.T) {
	list := List(struct {
		A string `default:"a \"b\" c"`
		B string `default:"plain"`
	}{})
	want := []string{`A="a \"b\" c"`, "B=plain"}
	for i, w := range want {
		if list[i] != w {
			t.Errorf("got %s, want %s", list[i], w)
		}
	}
}

// Get had no way to reach a config parsed with a prefix
func TestGetWithPrefix(t *testing.T) {
	type conf struct {
		Port int `default:"8080"`
		Sub  struct {
			Name string `default:"x"`
		}
	}
	c := &conf{}
	if err := Parse(c, "APP"); err != nil {
		t.Fatal(err)
	}

	if v, err := GetInt64(c, "APP_PORT", "APP"); err != nil || v != 8080 {
		t.Errorf("Get with prefix failed: %v %v", v, err)
	}
	if v, err := GetString(c, "APP_SUB_NAME", "APP"); err != nil || v != "x" {
		t.Errorf("Get with prefix failed: %q %v", v, err)
	}
	// and a read-only config can be searched too
	if v, err := GetInt64(*c, "PORT"); err != nil || v != 8080 {
		t.Errorf("Get on a value config failed: %v %v", v, err)
	}
}

// named types used to fall through the type assertions of the Get helpers
func TestGetNamedType(t *testing.T) {
	c := &struct {
		Lvl  level `default:"5"`
		Name name  `default:"hey"`
	}{}
	if err := Parse(c); err != nil {
		t.Fatal(err)
	}
	if v, err := GetInt64(c, "LVL"); err != nil || v != 5 {
		t.Errorf("GetInt64 on a named int failed: %v %v", v, err)
	}
	if v, err := GetString(c, "NAME"); err != nil || v != "hey" {
		t.Errorf("GetString on a named string failed: %q %v", v, err)
	}
}

// parse errors are wrapped, so the caller can inspect them
func TestErrorUnwrap(t *testing.T) {
	c := &struct {
		N int `env:"REG_N"`
	}{}
	withEnv(t, "REG_N", "not-a-number")
	err := Parse(c)
	if err == nil {
		t.Fatal("expected an error")
	}
	var numErr *strconv.NumError
	if !errors.As(err, &numErr) {
		t.Errorf("error is not wrapped: %v", err)
	}
}

// the Get helpers report a clear error when the field has another type
func TestGetTypeMismatch(t *testing.T) {
	c := &struct {
		S string  `default:"x"`
		N int     `default:"1"`
		F float64 `default:"1.5"`
		B bool    `default:"true"`
	}{}
	if err := Parse(c); err != nil {
		t.Fatal(err)
	}
	if _, err := GetBool(c, "S"); err == nil {
		t.Error("expected an error from GetBool on a string")
	}
	if _, err := GetInt64(c, "S"); err == nil {
		t.Error("expected an error from GetInt64 on a string")
	}
	if _, err := GetString(c, "N"); err == nil {
		t.Error("expected an error from GetString on an int")
	}
	if _, err := GetFloat64(c, "B"); err == nil {
		t.Error("expected an error from GetFloat64 on a bool")
	}
	if v, err := GetFloat64(c, "F"); err != nil || v != 1.5 {
		t.Errorf("GetFloat64 failed: %v %v", v, err)
	}
	if v, err := GetBool(c, "B"); err != nil || !v {
		t.Errorf("GetBool failed: %v %v", v, err)
	}
	for _, get := range []func() error{
		func() error { _, err := Get(c, "NOPE"); return err },
		func() error { _, err := GetBool(c, "NOPE"); return err },
		func() error { _, err := GetInt64(c, "NOPE"); return err },
		func() error { _, err := GetString(c, "NOPE"); return err },
		func() error { _, err := GetFloat64(c, "NOPE"); return err },
	} {
		if err := get(); err == nil {
			t.Error("expected a not found error")
		}
	}
}

// a custom separator keeps its empty elements, only the default
// whitespace one collapses them
func TestCustomSplitChar(t *testing.T) {
	e := New()
	e.Advance.SplitChar = ","
	c := &struct {
		S []string `default:"a,b,,c"`
	}{}
	if err := e.Parse(c); err != nil {
		t.Fatal(err)
	}
	if len(c.S) != 4 || c.S[3] != "c" {
		t.Errorf("custom split char failed: %q", c.S)
	}
}

// a pointer to a slice is filled through the same path as a slice
func TestPointerToSliceErrors(t *testing.T) {
	c := &struct {
		P *[]int `env:"REG_PSL"`
	}{}
	withEnv(t, "REG_PSL", "1 nope")
	if err := Parse(c); err == nil {
		t.Error("expected an error for a bad element in *[]int")
	}
}

// the cycle guard must only stop a type from containing itself, two
// sections of the same type next to each other are perfectly fine
func TestSameTypeSections(t *testing.T) {
	type common struct {
		Host string `default:"localhost"`
	}
	type conf struct {
		Value common  `yaml:"value"`
		A     *common `yaml:"a"`
		B     *common `yaml:"b"`
	}

	withEnv(t, "A_HOST", "a.local")
	withEnv(t, "B_HOST", "b.local")

	c := &conf{}
	if err := Parse(c); err != nil {
		t.Fatal(err)
	}
	if c.Value.Host != "localhost" {
		t.Errorf("value section: %+v", c.Value)
	}
	if c.A == nil || c.A.Host != "a.local" {
		t.Errorf("section A: %+v", c.A)
	}
	if c.B == nil || c.B.Host != "b.local" {
		t.Errorf("section B: %+v", c.B)
	}

	list := strings.Join(List(conf{}), " ")
	for _, want := range []string{"VALUE_HOST=localhost", "A_HOST=localhost", "B_HOST=localhost"} {
		if !strings.Contains(list, want) {
			t.Errorf("missing %s in %s", want, list)
		}
	}
}

// a separator the caller chose is taken literally; only the default one
// collapses repeats. A "is the separator whitespace" test used to route
// a deliberate tab through strings.Fields and split on spaces too.
func TestCustomWhitespaceSplitChar(t *testing.T) {
	for _, tc := range []struct {
		sep  string
		in   string
		want []string
	}{
		{"\t", "a b\tc d", []string{"a b", "c d"}},
		{"\n", "a b\nc d", []string{"a b", "c d"}},
		{",", "a b,c d", []string{"a b", "c d"}},
		{" ", " a  b ", []string{"a", "b"}}, // the default still collapses
		{"", "a  b", []string{"a", "b"}},    // unset falls back to the default
	} {
		e := New()
		e.Advance.SplitChar = tc.sep
		c := &struct {
			S []string `env:"SPLIT_S"`
		}{}
		os.Setenv("SPLIT_S", tc.in)
		err := e.Parse(c)
		os.Unsetenv("SPLIT_S")
		if err != nil {
			t.Errorf("sep %q: %v", tc.sep, err)
			continue
		}
		if strings.Join(c.S, "|") != strings.Join(tc.want, "|") {
			t.Errorf("sep %q: got %q, want %q", tc.sep, c.S, tc.want)
		}
	}
}

// a section is allocated when one of its fields was assigned, even if
// every value in it is zero. Testing the result for zero instead used to
// leave the section nil for an explicit PORT=0.
func TestPointerSectionExplicitZero(t *testing.T) {
	type sub struct {
		Port int
		Name string
	}
	type conf struct {
		P     *sub `yaml:"p"`
		Empty *sub `yaml:"empty"`
	}

	withEnv(t, "P_PORT", "0")
	c := &conf{}
	if err := Parse(c); err != nil {
		t.Fatal(err)
	}
	if c.P == nil {
		t.Error("a section with an explicit value must be allocated")
	} else if c.P.Port != 0 {
		t.Errorf("port: %d", c.P.Port)
	}
	// and a section nothing was said about stays nil
	if c.Empty != nil {
		t.Errorf("untouched section should stay nil, got %+v", c.Empty)
	}

	// every key List advertises has to be readable back after Parse
	for _, item := range List(conf{}) {
		key := strings.Split(item, "=")[0]
		if _, err := Get(c, key); err != nil && strings.HasPrefix(key, "P_") {
			t.Errorf("listed key %s is not gettable: %v", key, err)
		}
	}
}

// a value that merely contains an "e" is reported as a whole, it used to
// be blamed on the fragment after the e
func TestNonNumericValueErrorMessage(t *testing.T) {
	for _, in := range []string{"not-a-number", "hello"} {
		c := &struct {
			N int `env:"MSG_N"`
		}{}
		os.Setenv("MSG_N", in)
		err := Parse(c)
		os.Unsetenv("MSG_N")
		if err == nil {
			t.Errorf("%q should not parse", in)
			continue
		}
		if !strings.Contains(err.Error(), in) {
			t.Errorf("error for %q does not mention it: %v", in, err)
		}
	}
}
