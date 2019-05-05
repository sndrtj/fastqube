package main

import "testing"

func intSliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for idx, val := range a {
		if b[idx] != val {
			return false
		}
	}
	return true
}

func TestQualityDecoding(t *testing.T) {
	cases := []struct {
		s    string
		want []int
	}{
		{"JJJJJ", []int{41, 41, 41, 41, 41}},
	}

	for _, c := range cases {
		got := decodeQualitryString(c.s)
		if !intSliceEqual(got, c.want) {
			t.Errorf("Wanted %q, got %q", c.want, got)
		}
	}
}
