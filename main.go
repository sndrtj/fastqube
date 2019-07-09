package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

const programVersion = "0.0.1"

type compressOptions struct {
	bitsPerBase int
	bitsPerQual int
	bytesPerID  int
}

type fastqRead struct {
	readID    string
	seq       []int
	qualities []int
}

func (read fastqRead) compressedSeq(bitsPerBase int) []byte {
	return compressIntSlice(read.seq, bitsPerBase)
}

func (read fastqRead) compressedQual(bitsPerQual int) []byte {
	return compressIntSlice(read.qualities, bitsPerQual)
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

// Uses three-bit encoding for a DNA base
// returns 4 for everything that is not [ACTG]
func threeBitDNA(base rune) int {
	var result int

	switch base {
	case 'A':
		result = 0
	case 'T':
		result = 1
	case 'C':
		result = 2
	case 'G':
		result = 3
	default:
		result = 4
	}

	return result
}

// Uses two-bit encoding for a DNA base
// effectively squashes all non-[ACTG] bases to a G.
func twoBitDNA(base rune) int {
	var result int

	switch base {
	case 'A':
		result = 0
	case 'T':
		result = 1
	case 'C':
		result = 2
	case 'G':
		result = 3
	default:
		result = 3
	}

	return result
}

func blockQual(qual int) int {
	var result int

	switch {
	case qual < 2:
		result = 0
	case qual < 26:
		result = 1 // decodes to 2
	case qual < 31:
		result = 2 // decodes to 26
	case qual < 41:
		result = 3 // decodes to 31
	case qual >= 41:
		result = 4 // decodes to 41
	}

	return result
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
func decodeQualitryString(s string, blockQuals bool) []int {
	result := make([]int, len(s))
	for idx, item := range s {
		quality := int(item) - 33 // qualities are offset by 33
		if blockQuals {
			result[idx] = blockQual(quality)
		} else {
			result[idx] = quality
		}
	}
	return result
}

// return DNA sequence as slice of ints.
// errors when using unknown number of bits per base
func seqStringToInts(s string, bitsPerBase int) ([]int, error) {
	if bitsPerBase != 2 && bitsPerBase != 3 {
		return nil, errors.New("Must use 2 or 3 bits per base")
	}

	result := make([]int, len(s))

	for idx, runeValue := range s {
		var intValue int
		if bitsPerBase == 2 {
			intValue = twoBitDNA(runeValue)
		} else if bitsPerBase == 3 {
			intValue = threeBitDNA(runeValue)
		}
		result[idx] = intValue
	}
	return result, nil
}

func fastqReadFromBucket(bucket []string, opts compressOptions) (fastqRead, error) {
	var read fastqRead
	if len(bucket) != 3 {
		return read, errors.New("Read must consist of 3 strings")
	}
	readID := bucket[0]
	seq, err := seqStringToInts(bucket[1], opts.bitsPerBase)
	if err != nil {
		return read, err
	}
	var block bool
	if opts.bitsPerQual == 6 {
		block = false
	} else {
		block = true
	}
	qualities := decodeQualitryString(bucket[2], block)
	read = fastqRead{readID: readID, seq: seq, qualities: qualities}
	return read, nil
}

func compressFastqBucket(bucket []string, opts compressOptions) []byte {
	read, err := fastqReadFromBucket(bucket, opts)
	if err != nil {
		log.Fatalln(err)
	}
	var readAsBytes []byte

	// IDs do not get stored at all when bytesPerID is zero
	if opts.bytesPerID > 0 {
		readID, _ := read.byteID(opts.bytesPerID)
		readAsBytes = append(readAsBytes, readID...)
	}

	readAsBytes = append(readAsBytes, read.compressedSeq(opts.bitsPerBase)...)
	readAsBytes = append(readAsBytes, read.compressedQual(opts.bitsPerQual)...)
	return readAsBytes
}

func utcTime() string {
	loc, _ := time.LoadLocation("UTC")
	return time.Now().In(loc).String()
}
func createHeader(capacity int, bitsPerBase int) ([]byte, error) {
	var header string
	programLines := fmt.Sprintf("Program: fastqube\nVersion: %s\n", programVersion)
	modeLine := "Mode: LOSSLESS\n"
	encodingLine := fmt.Sprintf("Encoding:\n\tSequence: %d bit\n\tQualities: 6 bit\n", bitsPerBase)
	capacityLine := "Capacities:\n\tHeader: 4096 bytes\n\tRead IDs: 64 bytes\n"
	dateLine := fmt.Sprintf("Date: %s\n", utcTime())
	header = programLines + modeLine + encodingLine + capacityLine + dateLine
	byteHeader := []byte(header)
	paddingLength := capacity - len(byteHeader)
	if paddingLength < 0 {
		return nil, errors.New("Header too long")
	} else if paddingLength > 0 {
		padding := make([]byte, paddingLength)
		byteHeader = append(byteHeader, padding...)
	}
	return byteHeader, nil
}

func compressPath(path string, opts compressOptions) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalln(err)
	}
	header, _ := createHeader(4096, opts.bitsPerBase)
	binary.Write(os.Stdout, binary.BigEndian, header)
	defer file.Close()

	bucket := make([]string, 0, 3) // hold bucket of strings, representing a read

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if len(bucket) == 3 {
			compressed := compressFastqBucket(bucket, opts)
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
		compressed := compressFastqBucket(bucket, opts)
		binary.Write(os.Stdout, binary.BigEndian, compressed)
	}
}

func main() {
	var (
		compress       bool
		decompress     bool
		filePath       string
		twoBitEncoding bool
		bitsPerBase    int
		blockQualities bool
		bitsPerQual    int
		bytesPerID     int
	)
	flag.BoolVar(&compress, "c", false, "Compress")
	flag.BoolVar(&decompress, "d", false, "Decompress")
	flag.BoolVar(&twoBitEncoding, "2", false, "2Bit-encoding")
	flag.BoolVar(&blockQualities, "b", false, "Block Qualities")
	flag.IntVar(&bytesPerID, "B", 64, "Bytes per ID")

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

	if twoBitEncoding {
		bitsPerBase = 2
	} else {
		bitsPerBase = 3
	}

	if blockQualities {
		bitsPerQual = 3
	} else {
		bitsPerQual = 6
	}

	opts := compressOptions{bitsPerBase, bitsPerQual, bytesPerID}

	if compress {
		compressPath(filePath, opts)
	}
}
