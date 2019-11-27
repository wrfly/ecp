package ecp

import "testing"

func TestParseScientific(t *testing.T) {
	testCases := map[string]string{
		"1e3":       "1000",
		"1E3":       "1000",
		"1e11":      "100000000000",
		"1,000,000": "1000000",
	}
	for k, v := range testCases {
		r, err := parseScientific(k)
		if err != nil {
			t.Error(err)
		} else if r != v {
			t.Errorf("parse %s error, result=%s", k, r)
		}
	}

	badCases := map[string]string{
		"1e":     "1",
		"1E":     "1",
		"1e1e1e": "1",
	}
	for k := range badCases {
		_, err := parseScientific(k)
		if err == nil {
			t.Errorf("??? %s", k)
		}
	}
}
