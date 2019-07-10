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

func boolSliceEqual(a, b []bool) bool {
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

func byteSliceEqual(a, b []byte) bool {
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
		got := decodeQualitryString(c.s, false)
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
		got, err := seqStringToInts(c.s, 3)
		if err != nil {
			t.Errorf(err.Error())
		}
		if !intSliceEqual(got, c.want) {
			t.Errorf("Wanted %q, got %q", c.want, got)
		}
	}
}

func TestFastqReadFromBucket(t *testing.T) {
	bucket := make([]string, 3)
	bucket[0] = "some fancy id"
	bucket[1] = "AAAAA"
	bucket[2] = "FFFFF"

	read, err := fastqReadFromBucket(bucket, compressOptions{3, 6, 64})
	if err != nil {
		t.Errorf(err.Error())
	}

	wantedQual := []int{37, 37, 37, 37, 37}
	wantedSeq := []int{0, 0, 0, 0, 0}
	if !intSliceEqual(wantedSeq, read.seq) {
		t.Errorf("Wanted %q, got %q", wantedSeq, read.seq)
	}
	if !intSliceEqual(wantedQual, read.qualities) {
		t.Errorf("Wanted %q, got %q", wantedQual, read.qualities)
	}
}

func TestReversed(t *testing.T) {
	a := []bool{true, true, false}
	b := []bool{false, true, true}
	got := reverseSliceB(a)
	if !boolSliceEqual(b, got) {
		t.Errorf("not equal")
	}
}

func TestBoolSliceToByte(t *testing.T) {
	cases := []struct {
		s    []bool
		want byte
	}{
		{[]bool{false, false, false, false, false, false, false, false}, byte(0)},
		{[]bool{false, false, false, false, false, false, false, true}, byte(1)},
		{[]bool{true, false, false, false, false, false, false, false}, byte(128)},
		{[]bool{true, true, true, true, true, true, true, true}, byte(255)},
	}

	for _, c := range cases {
		got, _ := boolSliceToByte(c.s)
		if got != c.want {
			t.Errorf("Wanted %d, got %d", int(c.want), int(got))
		}
	}
}

func TestUint8ToBoolSlice(t *testing.T) {
	cases := []struct {
		u    uint8
		c    int
		want []bool
	}{
		{1, 8, []bool{false, false, false, false, false, false, false, true}},
		{6, 3, []bool{true, true, false}},
	}
	for _, c := range cases {
		got, _ := uint8ToBoolSlice(c.u, c.c)
		if !boolSliceEqual(got, c.want) {
			t.Errorf("Not equal")
		}
	}
}

func TestCompressIntSlice(t *testing.T) {
	cases := []struct {
		s    []int
		c    int
		want []byte
	}{
		{[]int{1, 1, 1, 1}, 3, []byte{36, 144}},
	}
	for _, c := range cases {
		got := compressIntSlice(c.s, c.c)
		if !byteSliceEqual(c.want, got) {
			t.Errorf("Not equal")
		}
	}
}

func TestTwoBitDna(t *testing.T) {
	got := twoBitDNA('N')
	if got != 3 {
		t.Errorf("Not equal")
	}
}

func TestBlockQual(t *testing.T) {
	cases := []struct {
		q    int
		want int
	}{
		{1, 0},
		{20, 1},
		{30, 2},
		{40, 3},
		{500, 4},
	}

	for _, c := range cases {
		got := blockQual(c.q)
		if got != c.want {
			t.Errorf("Not equal")
		}
	}
}

func TestCompressedSeq(t *testing.T) {
	cases := []struct {
		seq         []int
		quals       []int
		bitsPerBase int
		want        []byte
	}{
		{[]int{0, 1, 2, 3}, []int{0, 0, 0, 0}, 3, []byte{5, 48}}, // this is ACTG
		{[]int{0, 1, 2, 3}, []int{0, 0, 0, 0}, 2, []byte{27}},    // this is also ACTG
	}

	for _, c := range cases {
		read := fastqRead{"la", c.seq, c.quals}
		got := read.compressedSeq(c.bitsPerBase)
		if !byteSliceEqual(c.want, got) {
			t.Errorf("Not equals")
		}
	}
}

func TestCompressedQual(t *testing.T) {
	cases := []struct {
		seq         []int
		quals       []int
		bitsPerBase int
		want        []byte
	}{
		{[]int{0, 1, 2, 3}, []int{10, 20, 30, 40}, 6, []byte{41, 71, 168}}, // this is ACTG
		{[]int{0, 1, 2, 3}, []int{0, 1, 2, 3}, 3, []byte{5, 48}},           // this is also ACTG
	}

	for _, c := range cases {
		read := fastqRead{"la", c.seq, c.quals}
		got := read.compressedQual(c.bitsPerBase)
		if !byteSliceEqual(c.want, got) {
			t.Errorf("Not equals")
		}
	}
}

func TestCompressedID(t *testing.T) {
	cases := []struct {
		readID   string
		capacity int
		want     []byte
	}{
		{"la", 2, []byte{108, 97}},
		{"la", 3, []byte{108, 97, 0}},
	}

	for _, c := range cases {
		read := fastqRead{c.readID, []int{1, 1, 1, 1}, []int{1, 1, 1, 1}}
		got, _ := read.byteID(c.capacity)
		if !byteSliceEqual(c.want, got) {
			t.Errorf("Not equals")
		}
	}
}
