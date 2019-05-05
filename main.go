package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
)

type fastqRead struct {
	readID    string
	seq       []int
	qualities []int
}

// returns a slice of integers representing the decoded quality
// string of a fastq read
func decodeQualitryString(s string) []int {
	result := make([]int, len(s))
	for idx, item := range s {
		quality := int(item) - 32 // qualities are offset by 32
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

func compressFastqBucket(bucket []string) {
	_, err := fastqReadFromBucket(bucket)
	if err != nil {
		log.Fatalln(err)
	}
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
			compressFastqBucket(bucket)
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

	fmt.Println(compress)
	fmt.Println(decompress)
}
