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
		{"!J!J!J!", []int{0, 41, 0, 41, 0, 41, 0}},
	}

	for _, c := range cases {
		got := decodeQualitryString(c.s)
		if !intSliceEqual(got, c.want) {
			t.Errorf("Wanted %q, got %q", c.want, got)
		}
	}
}

func TestSeqStringAsIntValid(t *testing.T) {
	cases := []struct {
		s    string
		want []int
	}{
		{"AAAAA", []int{0, 0, 0, 0, 0}},
		{"ACTGN", []int{0, 2, 1, 3, 4}},
	}

	for _, c := range cases {
		got, err := seqStringToInts(c.s)
		if err != nil {
			t.Errorf(err.Error())
		}
		if !intSliceEqual(got, c.want) {
			t.Errorf("Wanted %q, got %q", c.want, got)
		}
	}
}

func TestSeqStringAsIntInvalid(t *testing.T) {
	cases := []string{
		"LALALA", "NONONONONO", "YES!",
	}
	for _, c := range cases {
		_, err := seqStringToInts(c)
		if err == nil {
			t.Errorf("Sequence %q did not raise an error", c)
		}
	}
}
