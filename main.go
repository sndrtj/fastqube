package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"flag"
	"log"
	"os"
)

type fastqRead struct {
	readID    string
	seq       []int
	qualities []int
}

func (read fastqRead) compressedSeq() []byte {
	return compressIntSlice(read.seq, 3)
}

func (read fastqRead) compressedQual() []byte {
	return compressIntSlice(read.qualities, 6)
}

func (read fastqRead) byteID(capacity int) ([]byte, error) {
	paddingLength := capacity - len(read.readID)
	if paddingLength < 0 {
		return nil, errors.New("Read ID is too large")
	}
	byteID := []byte(read.readID)
	if paddingLength > 0 {
		padding := make([]byte, paddingLength) // defaults = 0
		byteID = append(byteID, padding...)
	}
	return byteID, nil
}

func compressIntSlice(s []int, bitsPerItem int) []byte {
	var boolSlice []bool
	for _, item := range s {
		slice, err := uint8ToBoolSlice(uint8(item), bitsPerItem)
		if err != nil {
			log.Fatalln("Could not compress seq")
		}
		boolSlice = append(boolSlice, slice...)
	}
	amountToPad := len(boolSlice) % 8
	padding := make([]bool, amountToPad) // default value is false
	boolSlice = append(boolSlice, padding...)
	nBytes := len(boolSlice) / 8
	compressed := make([]byte, nBytes)
	for i := 0; i < nBytes; i++ {
		start := i * 8
		end := (i + 1) * 8
		oneByteSlice := boolSlice[start:end]
		oneByte, _ := boolSliceToByte(oneByteSlice)
		compressed[i] = oneByte
	}
	return compressed
}

func reverseSliceB(s []bool) []bool {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// returns a byte from a slice of booleans
func boolSliceToByte(s []bool) (byte, error) {
	if len(s) != 8 {
		return byte(0), errors.New("Slice must have length 8")
	}

	var result uint8
	reversed := reverseSliceB(s)

	for idx, el := range reversed {
		if el {
			result += uint8(1) << uint8(idx)
		}
	}
	return result, nil
}

// convert a uint8 to to a slice of booleans
// capacity controls how large the slice will become
// capacity is maximally 8.
func uint8ToBoolSlice(b uint8, capacity int) ([]bool, error) {
	if capacity > 8 || capacity < 1 {
		return nil, errors.New("Capacity must be between 1 and 8")
	}
	result := make([]bool, 0, capacity)

	for i := capacity - 1; i >= 0; i-- {
		bit := b >> uint8(i) & 1
		if bit == 1 {
			result = append(result, true)
		} else {
			result = append(result, false)
		}
	}
	return result, nil
}

// returns a slice of integers representing the decoded quality
// string of a fastq read
func decodeQualitryString(s string) []int {
	result := make([]int, len(s))
	for idx, item := range s {
		quality := int(item) - 33 // qualities are offset by 33
		result[idx] = quality
	}
	return result
}

// return DNA sequence as slice of ints.
// errors when encountering an unknown base. Only A,C,T,G,N are allowed.
func seqStringToInts(s string) ([]int, error) {
	baseMap := map[string]int{
		"A": 0,
		"T": 1,
		"C": 2,
		"G": 3,
		"N": 4}

	result := make([]int, len(s))

	for idx, runeValue := range s {
		intValue, present := baseMap[string(runeValue)]
		if !present {
			return nil, errors.New("Unknown base " + string(runeValue))
		}
		result[idx] = intValue
	}
	return result, nil
}

func fastqReadFromBucket(bucket []string) (fastqRead, error) {
	var read fastqRead
	if len(bucket) != 3 {
		return read, errors.New("Read must consist of 3 strings")
	}
	readID := bucket[0]
	seq, err := seqStringToInts(bucket[1])
	if err != nil {
		return read, err
	}
	qualities := decodeQualitryString(bucket[2])
	read = fastqRead{readID: readID, seq: seq, qualities: qualities}
	return read, nil
}

func compressFastqBucket(bucket []string) []byte {
	read, err := fastqReadFromBucket(bucket)
	if err != nil {
		log.Fatalln(err)
	}
	var readAsBytes []byte
	readID, _ := read.byteID(64)
	readAsBytes = append(readAsBytes, readID...)
	readAsBytes = append(readAsBytes, read.compressedSeq()...)
	readAsBytes = append(readAsBytes, read.compressedQual()...)
	return readAsBytes
}

func compressPath(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	bucket := make([]string, 0, 3) // hold bucket of strings, representing a read

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if len(bucket) == 3 {
			compressed := compressFastqBucket(bucket)
			binary.Write(os.Stdout, binary.BigEndian, compressed)
			bucket = make([]string, 0, 3)
		}
		currentLine := scanner.Text()
		if currentLine != "+" {
			bucket = append(bucket, currentLine)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalln(err)
	}

	// process final bucket
	if len(bucket) == 3 {
		compressFastqBucket(bucket)
	}
}

func main() {
	var (
		compress   bool
		decompress bool
		filePath   string
	)
	flag.BoolVar(&compress, "c", false, "Compress")
	flag.BoolVar(&decompress, "d", false, "Decompress")

	flag.Parse()

	tail := flag.Args()
	if len(tail) > 0 {
		// parse the file
		filePath = tail[0]
	} else {
		log.Fatalln("No input file specified")
	}

	if compress && decompress {
		log.Fatalln("Cannot set both -c and -d")
	}

	if !compress && !decompress {
		// setting compress to true as default for the group
		compress = true
	}

	if compress {
		compressPath(filePath)
	}
}
